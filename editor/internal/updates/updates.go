// Package updates checks the original source and stages recoverable code updates.
// Update checks are read-only; installation uses protected transaction directories.
package updates

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Gonie-Gonie/Gonie-Gonie.github.io/editor/internal/dataversion"
	"github.com/Gonie-Gonie/Gonie-Gonie.github.io/editor/internal/release"
)

const (
	DefaultRepository = "Gonie-Gonie/Gonie-Gonie.github.io"
	DefaultRef        = "main"
	DefaultBaseURL    = "https://raw.githubusercontent.com"
	CheckTimeout      = 30 * time.Second
	RequestTimeout    = 15 * time.Second
	MaxManifestBytes  = 64 << 10
	MaxIndexBytes     = 256 << 10
	MaxNoteBytes      = 128 << 10
	maxReleases       = 200
)

type Source struct {
	Repository string `json:"repository"`
	Ref        string `json:"ref"`
}

type Info struct {
	CurrentVersion   string `json:"current_version"`
	SourceRepository string `json:"source_repository"`
	SourceRef        string `json:"source_ref"`
	SourceURL        string `json:"source_url"`
	RecoveryStatus   string `json:"recovery_status,omitempty"`
	RecoveryMessage  string `json:"recovery_message,omitempty"`
}

type Release struct {
	Version string   `json:"version"`
	Date    string   `json:"date"`
	Notes   []string `json:"notes"`
}

type Response struct {
	Info
	LatestVersion   string    `json:"latest_version"`
	UpdateAvailable bool      `json:"update_available"`
	Releases        []Release `json:"releases"`
}

// Client and BaseURL are injectable for isolated HTTPS tests. The application
// supplies only Root and therefore always uses raw.githubusercontent.com.
type Options struct {
	Root    string
	Client  *http.Client
	BaseURL string
}

var repositoryPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9-]*/[A-Za-z0-9_.-]+$`)
var refPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._/-]*$`)

func validateSource(source Source) error {
	if len(source.Repository) > 200 || !repositoryPattern.MatchString(source.Repository) {
		return fmt.Errorf("원본 저장소 형식이 올바르지 않습니다")
	}
	for _, part := range strings.Split(source.Repository, "/") {
		if part == "." || part == ".." {
			return fmt.Errorf("원본 저장소 형식이 올바르지 않습니다")
		}
	}
	if len(source.Ref) > 200 || !refPattern.MatchString(source.Ref) || strings.Contains(source.Ref, "..") || strings.Contains(source.Ref, "//") || strings.HasSuffix(source.Ref, "/") || strings.HasSuffix(source.Ref, ".") || strings.HasSuffix(source.Ref, ".lock") {
		return fmt.Errorf("원본 브랜치 형식이 올바르지 않습니다")
	}
	for _, part := range strings.Split(source.Ref, "/") {
		if part == "." || part == ".." || strings.HasPrefix(part, ".") || strings.HasSuffix(part, ".lock") {
			return fmt.Errorf("원본 브랜치 형식이 올바르지 않습니다")
		}
	}
	return nil
}

func boundedRead(reader io.Reader, maximum int64) ([]byte, error) {
	raw, err := io.ReadAll(io.LimitReader(reader, maximum+1))
	if err != nil {
		return nil, err
	}
	if int64(len(raw)) > maximum {
		return nil, fmt.Errorf("업데이트 정보가 허용 크기를 초과했습니다")
	}
	if !utf8.Valid(raw) {
		return nil, fmt.Errorf("업데이트 정보가 올바른 UTF-8이 아닙니다")
	}
	return raw, nil
}

// GetInfo requires only app/release.json, independently from personal data.
func GetInfo(root string) (Info, error) {
	var info Info
	file, err := os.Open(filepath.Join(root, "app", "release.json"))
	if err != nil {
		return info, fmt.Errorf("현재 버전을 읽지 못했습니다: %w", err)
	}
	defer file.Close()
	raw, err := boundedRead(file, MaxManifestBytes)
	if err != nil {
		return info, err
	}
	var local struct {
		Version  string `json:"version"`
		Upstream Source `json:"upstream"`
	}
	if err := dataversion.Decode(raw, &local); err != nil {
		return info, fmt.Errorf("현재 버전 정보가 올바르지 않습니다: %w", err)
	}
	if _, err := dataversion.Parse(local.Version); err != nil {
		return info, err
	}
	if local.Upstream.Repository == "" {
		local.Upstream.Repository = DefaultRepository
	}
	if local.Upstream.Ref == "" {
		local.Upstream.Ref = DefaultRef
	}
	if err := validateSource(local.Upstream); err != nil {
		return info, err
	}
	info = Info{CurrentVersion: local.Version, SourceRepository: local.Upstream.Repository, SourceRef: local.Upstream.Ref,
		SourceURL: "https://github.com/" + local.Upstream.Repository + "/tree/" + url.PathEscape(local.Upstream.Ref)}
	return info, nil
}

type remote struct {
	client *http.Client
	base   *url.URL
	prefix string
}

func newRemote(options Options, info Info) (remote, error) {
	baseURL := options.BaseURL
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	base, err := url.Parse(baseURL)
	if err != nil || base.Scheme != "https" || base.Host == "" || base.User != nil || base.RawQuery != "" || base.Fragment != "" {
		return remote{}, fmt.Errorf("업데이트 주소는 HTTPS여야 합니다")
	}
	prefix := strings.TrimSuffix(base.Path, "/") + "/" + info.SourceRepository + "/" + info.SourceRef + "/"
	client := http.Client{Timeout: RequestTimeout}
	if options.Client != nil {
		client = *options.Client
	}
	if client.Timeout <= 0 || client.Timeout > RequestTimeout {
		client.Timeout = RequestTimeout
	}
	client.Jar = nil
	previousRedirect := client.CheckRedirect
	client.CheckRedirect = func(request *http.Request, via []*http.Request) error {
		if len(via) >= 5 || request.URL.Scheme != "https" || !strings.EqualFold(request.URL.Host, base.Host) || !strings.HasPrefix(request.URL.Path, prefix) || path.Clean(request.URL.Path) != request.URL.Path || strings.Contains(request.URL.Path, "\\") || request.URL.RawQuery != "" {
			return fmt.Errorf("원본 저장소 밖으로 이동하는 업데이트 요청을 차단했습니다")
		}
		if previousRedirect != nil {
			return previousRedirect(request, via)
		}
		return nil
	}
	return remote{client: &client, base: base, prefix: prefix}, nil
}

func (remote remote) get(ctx context.Context, name string, maximum int64) ([]byte, error) {
	address := *remote.base
	address.Path = remote.prefix + name
	address.RawPath = ""
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, address.String(), nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "application/json, text/plain")
	request.Header.Set("User-Agent", "Profile-Editor update check")
	response, err := remote.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("업데이트 정보를 가져오지 못했습니다 (%s): %w", name, err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("업데이트 정보를 가져오지 못했습니다: %s (HTTP %d)", name, response.StatusCode)
	}
	if response.ContentLength > maximum {
		return nil, fmt.Errorf("업데이트 정보가 허용 크기를 초과했습니다: %s", name)
	}
	raw, err := boundedRead(response.Body, maximum)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	return raw, nil
}

func validateEntry(entry release.Entry) error {
	if _, err := dataversion.Parse(entry.Version); err != nil {
		return err
	}
	date, err := time.Parse("2006-01-02", entry.Date)
	if err != nil || date.Year() < 1 || date.Format("2006-01-02") != entry.Date {
		return fmt.Errorf("릴리스 날짜가 올바르지 않습니다: %s", entry.Version)
	}
	if entry.Notes != "app/releases/"+entry.Version+".md" {
		return fmt.Errorf("릴리스 노트 경로가 올바르지 않습니다: %s", entry.Version)
	}
	return nil
}

// Check enables availability only after the complete applicable history loads.
func Check(ctx context.Context, options Options) (Response, error) {
	info, err := GetInfo(options.Root)
	result := Response{Info: info, Releases: []Release{}}
	if err != nil {
		return result, err
	}
	ctx, cancel := context.WithTimeout(ctx, CheckTimeout)
	defer cancel()
	remote, err := newRemote(options, info)
	if err != nil {
		return result, err
	}
	raw, err := remote.get(ctx, "app/release.json", MaxManifestBytes)
	if err != nil {
		return result, err
	}
	var latest release.Manifest
	if err := dataversion.Decode(raw, &latest); err != nil {
		return result, err
	}
	if err := validateEntry(latest.Entry); err != nil {
		return result, err
	}
	comparison, _ := dataversion.Compare(info.CurrentVersion, latest.Version)
	if comparison >= 0 {
		result.LatestVersion = latest.Version
		return result, nil
	}
	raw, err = remote.get(ctx, "app/releases/index.json", MaxIndexBytes)
	if err != nil {
		return result, err
	}
	var index release.Index
	if err := dataversion.Decode(raw, &index); err != nil {
		return result, err
	}
	if err := dataversion.Validate(index.Version); err != nil {
		return result, err
	}
	if len(index.Releases) > maxReleases {
		return result, fmt.Errorf("릴리스 이력이 허용 개수를 초과했습니다")
	}
	selected := []release.Entry{}
	seen := map[string]bool{}
	hasLatest := false
	for _, entry := range index.Releases {
		if _, err := dataversion.Parse(entry.Version); err != nil {
			return result, err
		}
		if seen[entry.Version] {
			return result, fmt.Errorf("중복된 릴리스 버전입니다: %s", entry.Version)
		}
		seen[entry.Version] = true
		newer, _ := dataversion.Compare(entry.Version, info.CurrentVersion)
		within, _ := dataversion.Compare(entry.Version, latest.Version)
		if newer <= 0 || within > 0 {
			continue
		}
		if err := validateEntry(entry); err != nil {
			return result, err
		}
		if entry.Version == latest.Version {
			if entry != latest.Entry {
				return result, fmt.Errorf("최신 릴리스 정보와 변경 이력이 일치하지 않습니다")
			}
			hasLatest = true
		}
		selected = append(selected, entry)
	}
	if !hasLatest {
		return result, fmt.Errorf("최신 버전의 릴리스 노트가 없습니다")
	}
	sort.Slice(selected, func(i, j int) bool {
		order, _ := dataversion.Compare(selected[i].Version, selected[j].Version)
		return order > 0
	})
	history := make([]Release, 0, len(selected))
	for _, entry := range selected {
		raw, err := remote.get(ctx, entry.Notes, MaxNoteBytes)
		if err != nil {
			return result, err
		}
		notes, err := parseNotes(raw, entry)
		if err != nil {
			return result, err
		}
		history = append(history, Release{Version: entry.Version, Date: entry.Date, Notes: notes})
	}
	result.LatestVersion = latest.Version
	result.UpdateAvailable = true
	result.Releases = history
	return result, nil
}

func parseNotes(raw []byte, entry release.Entry) ([]string, error) {
	if !utf8.Valid(raw) {
		return nil, fmt.Errorf("릴리스 노트가 올바른 UTF-8이 아닙니다")
	}
	note := strings.TrimPrefix(strings.ReplaceAll(string(raw), "\r\n", "\n"), "\ufeff")
	if !strings.HasPrefix(note, "# "+entry.Version+"\n") {
		return nil, fmt.Errorf("릴리스 노트 버전이 일치하지 않습니다: %s", entry.Version)
	}
	items := []string{}
	fence := ""
	hasDate := false
	continuation := false
	for _, line := range strings.Split(note, "\n") {
		text := strings.TrimSpace(line)
		if fence != "" {
			if strings.HasPrefix(text, fence) && strings.Trim(text, string(fence[0])) == "" {
				fence = ""
			}
			continuation = false
			continue
		}
		if strings.HasPrefix(text, "```") || strings.HasPrefix(text, "~~~") {
			length := 0
			for length < len(text) && text[length] == text[0] {
				length++
			}
			fence = text[:length]
			continuation = false
			continue
		}
		if text == entry.Date {
			hasDate = true
			continuation = false
			continue
		}
		if strings.HasPrefix(text, "#") {
			continuation = false
			continue
		}
		if len(text) >= 2 && (strings.HasPrefix(text, "- ") || strings.HasPrefix(text, "* ") || strings.HasPrefix(text, "+ ")) {
			item := strings.TrimSpace(text[2:])
			if item != "" {
				items = append(items, item)
				continuation = true
			}
			continue
		}
		if continuation && text != "" && len(line) > len(strings.TrimLeft(line, " \t")) {
			items[len(items)-1] += " " + text
		} else if text != "" {
			continuation = false
		}
	}
	if !hasDate || len(items) == 0 || len(items) > 200 {
		return nil, fmt.Errorf("릴리스 노트에 날짜와 변경 목록이 필요합니다: %s", entry.Version)
	}
	return items, nil
}
