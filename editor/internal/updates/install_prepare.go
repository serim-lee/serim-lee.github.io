package updates

import (
	"context"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Gonie-Gonie/Gonie-Gonie.github.io/editor/internal/dataversion"
)

type installFetcher struct {
	client *http.Client
	base   *url.URL
	prefix string
}

func newInstallFetcher(baseURL, prefix string, provided *http.Client) (installFetcher, error) {
	base, err := url.Parse(baseURL)
	if err != nil || base.Scheme != "https" || base.Host == "" || base.User != nil || base.RawQuery != "" || base.Fragment != "" {
		return installFetcher{}, fmt.Errorf("업데이트 서버는 HTTPS여야 합니다")
	}
	base.Path = strings.TrimSuffix(base.Path, "/")
	prefix = base.Path + prefix
	client := http.Client{Timeout: 90 * time.Second}
	if provided != nil {
		client = *provided
	}
	if client.Timeout <= 0 || client.Timeout > 90*time.Second {
		client.Timeout = 90 * time.Second
	}
	client.Jar = nil
	client.CheckRedirect = func(request *http.Request, via []*http.Request) error {
		if len(via) >= 5 || request.URL.Scheme != "https" || !strings.EqualFold(request.URL.Host, base.Host) || !strings.HasPrefix(request.URL.Path, prefix) || path.Clean(request.URL.Path) != request.URL.Path || strings.Contains(request.URL.Path, "\\") {
			return fmt.Errorf("업데이트 요청이 원본 경계를 벗어났습니다")
		}
		return nil
	}
	return installFetcher{client: &client, base: base, prefix: prefix}, nil
}

func (fetch installFetcher) get(ctx context.Context, name string, query url.Values, maximum int64) ([]byte, error) {
	address := *fetch.base
	address.Path = fetch.prefix + name
	address.RawPath = ""
	address.RawQuery = query.Encode()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, address.String(), nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("User-Agent", "Profile-Editor updater")
	request.Header.Set("Accept", "application/vnd.github+json, application/octet-stream")
	response, err := fetch.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("업데이트 다운로드 실패: %s (HTTP %d)", name, response.StatusCode)
	}
	if response.ContentLength > maximum {
		return nil, fmt.Errorf("업데이트 파일이 크기 제한을 초과했습니다: %s", name)
	}
	raw, err := io.ReadAll(io.LimitReader(response.Body, maximum+1))
	if err != nil {
		return nil, err
	}
	if int64(len(raw)) > maximum {
		return nil, fmt.Errorf("업데이트 파일이 크기 제한을 초과했습니다: %s", name)
	}
	return raw, nil
}

func gitBlobSHA(raw []byte) string {
	hash := sha1.New()
	fmt.Fprintf(hash, "blob %d\x00", len(raw))
	hash.Write(raw)
	return hex.EncodeToString(hash.Sum(nil))
}
func payloadSHA(raw []byte) string { sum := sha256.Sum256(raw); return hex.EncodeToString(sum[:]) }

func validExecutable(raw []byte) bool {
	if len(raw) < 68 || raw[0] != 'M' || raw[1] != 'Z' {
		return false
	}
	offset := uint64(raw[60]) | uint64(raw[61])<<8 | uint64(raw[62])<<16 | uint64(raw[63])<<24
	return offset >= 64 && offset+4 <= uint64(len(raw)) && string(raw[offset:offset+4]) == "PE\x00\x00"
}

func installProgress(options InstallOptions, phase, message, name string, completed, total int, bytes, totalBytes int64) {
	if options.Progress == nil {
		return
	}
	percent := -1
	if total > 0 {
		percent = 5 + int(65*float64(completed)/float64(total))
	}
	options.Progress(InstallProgress{Phase: phase, Message: message, Path: name, Completed: completed, Total: total, Bytes: bytes, TotalBytes: totalBytes, Percent: percent})
}

func readInstalled(root string) installedManifest {
	file, err := safePath(root, "editor/runtime/updates/installed.json")
	if err != nil {
		return installedManifest{}
	}
	raw, err := os.ReadFile(file)
	if err != nil || len(raw) > 8<<20 {
		return installedManifest{}
	}
	var installed installedManifest
	if dataversion.Decode(raw, &installed) != nil || len(installed.Files) > 5000 {
		return installedManifest{}
	}
	return installed
}

// PrepareInstall pins one original commit and downloads only policy-approved
// program blobs. Personal data, exports and repository history are not fetched.
func PrepareInstall(ctx context.Context, options InstallOptions) (prepared PreparedInstall, err error) {
	root, err := canonicalRoot(options.Root)
	if err != nil {
		return prepared, err
	}
	info, err := GetInfo(root)
	if err != nil {
		return prepared, err
	}
	pending, err := PendingJournals(root)
	if err != nil {
		return prepared, err
	}
	if len(pending) > 0 {
		return prepared, fmt.Errorf("이전 업데이트의 복구를 먼저 완료해 주세요")
	}
	ctx, cancel := context.WithTimeout(ctx, InstallTimeout)
	defer cancel()
	apiBase := options.APIBaseURL
	if apiBase == "" {
		apiBase = "https://api.github.com"
	}
	api, err := newInstallFetcher(apiBase, "/repos/"+info.SourceRepository+"/", options.Client)
	if err != nil {
		return prepared, err
	}
	installProgress(options, "resolve", "원본 버전 확인 중…", "", 0, 0, 0, 0)
	raw, err := api.get(ctx, "commits/"+info.SourceRef, nil, 1<<20)
	if err != nil {
		return prepared, err
	}
	var commit struct {
		SHA    string `json:"sha"`
		Commit struct {
			Tree struct {
				SHA string `json:"sha"`
			} `json:"tree"`
		} `json:"commit"`
	}
	if err := dataversion.Decode(raw, &commit); err != nil {
		return prepared, err
	}
	if !gitSHAPattern.MatchString(commit.SHA) || !gitSHAPattern.MatchString(commit.Commit.Tree.SHA) {
		return prepared, fmt.Errorf("원본 커밋 정보가 올바르지 않습니다")
	}
	rawBase := options.RawBaseURL
	if rawBase == "" {
		rawBase = DefaultBaseURL
	}
	source, err := newInstallFetcher(rawBase, "/"+info.SourceRepository+"/"+commit.SHA+"/", options.Client)
	if err != nil {
		return prepared, err
	}
	manifestRaw, err := source.get(ctx, "app/release.json", nil, MaxManifestBytes)
	if err != nil {
		return prepared, err
	}
	var target struct {
		Version string `json:"version"`
	}
	if err := dataversion.Decode(manifestRaw, &target); err != nil {
		return prepared, err
	}
	if _, err := dataversion.Parse(target.Version); err != nil {
		return prepared, err
	}
	if options.TargetVersion != "" && target.Version != options.TargetVersion {
		return prepared, fmt.Errorf("확인한 버전과 원본 버전이 달라졌습니다. 업데이트를 다시 확인해 주세요")
	}
	if comparison, _ := dataversion.Compare(info.CurrentVersion, target.Version); comparison >= 0 {
		return prepared, fmt.Errorf("현재 버전보다 새로운 업데이트가 없습니다")
	}
	raw, err = api.get(ctx, "git/trees/"+commit.Commit.Tree.SHA, url.Values{"recursive": {"1"}}, 8<<20)
	if err != nil {
		return prepared, err
	}
	var tree struct {
		SHA       string `json:"sha"`
		Truncated bool   `json:"truncated"`
		Tree      []struct {
			Path string `json:"path"`
			Mode string `json:"mode"`
			Type string `json:"type"`
			SHA  string `json:"sha"`
			Size int64  `json:"size"`
		} `json:"tree"`
	}
	if err := dataversion.Decode(raw, &tree); err != nil {
		return prepared, err
	}
	if tree.Truncated || tree.SHA != commit.Commit.Tree.SHA {
		return prepared, fmt.Errorf("원본 파일 목록을 완전히 확인할 수 없습니다")
	}
	policy := installPolicy(root)
	files := []InstallFile{}
	seen := map[string]bool{}
	var totalBytes int64
	for _, item := range tree.Tree {
		if item.Type == "tree" || !replaceAllowed(policy, item.Path) {
			continue
		}
		if (strings.EqualFold(item.Path, ExecutablePath) && item.Path != ExecutablePath) || (strings.EqualFold(item.Path, "app/release.json") && item.Path != "app/release.json") {
			return prepared, fmt.Errorf("update critical path has inconsistent casing: %s", item.Path)
		}
		if item.Type != "blob" || (item.Mode != "100644" && item.Mode != "100755") || !gitSHAPattern.MatchString(item.SHA) || item.Size < 0 || item.Size > MaxInstallFileBytes {
			return prepared, fmt.Errorf("지원하지 않는 원본 파일입니다: %s", item.Path)
		}
		if _, err := safePath(root, item.Path); err != nil {
			return prepared, err
		}
		key := strings.ToLower(item.Path)
		if seen[key] {
			return prepared, fmt.Errorf("대소문자가 충돌하는 원본 파일입니다: %s", item.Path)
		}
		seen[key] = true
		totalBytes += item.Size
		if totalBytes > MaxInstallBytes || len(files) >= 5000 {
			return prepared, fmt.Errorf("업데이트 파일의 총 크기를 초과했습니다")
		}
		mode := uint32(0644)
		if item.Mode == "100755" || item.Path == ExecutablePath {
			mode = 0755
		}
		files = append(files, InstallFile{Path: item.Path, Action: "replace", GitSHA: item.SHA, Size: item.Size, Deferred: item.Path == ExecutablePath, Mode: mode})
	}
	if !seen[strings.ToLower(ExecutablePath)] || !seen["app/release.json"] {
		return prepared, fmt.Errorf("원본 업데이트에 실행 파일 또는 버전 파일이 없습니다")
	}
	for _, old := range readInstalled(root).Files {
		key := strings.ToLower(old.Path)
		if !seen[key] && replaceAllowed(policy, old.Path) {
			if _, err := safePath(root, old.Path); err != nil {
				return prepared, err
			}
			files = append(files, InstallFile{Path: old.Path, Action: "delete", Mode: 0644})
			seen[key] = true
		}
	}
	sort.Slice(files, func(i, j int) bool {
		if files[i].Action != files[j].Action {
			return files[i].Action == "delete"
		}
		return files[i].Path < files[j].Path
	})
	var random [6]byte
	if _, err := rand.Read(random[:]); err != nil {
		return prepared, err
	}
	id := time.Now().UTC().Format("20060102T150405Z") + "-" + hex.EncodeToString(random[:])
	stageRelative := "editor/runtime/updates/" + id
	stage, err := safePath(root, stageRelative)
	if err != nil {
		return prepared, err
	}
	if err := os.MkdirAll(stage, 0700); err != nil {
		return prepared, err
	}
	journal := Journal{FormatVersion: 1, ID: id, Root: root, State: StateDownloading, CurrentVersion: info.CurrentVersion, TargetVersion: target.Version, Repository: info.SourceRepository, Ref: info.SourceRef, Commit: commit.SHA, CreatedAt: time.Now().UTC().Format(time.RFC3339), ProcessID: os.Getpid(), Files: files}
	journalPath := filepath.Join(stage, "journal.json")
	prepared = PreparedInstall{Root: root, JournalPath: journalPath, Journal: journal}
	if err := SaveJournal(root, journalPath, journal); err != nil {
		return prepared, err
	}
	defer func() {
		if err != nil {
			journal.State = StateDownloadFailed
			journal.Error = err.Error()
			_ = SaveJournal(root, journalPath, journal)
			prepared.Journal = journal
		}
	}()
	var completedBytes int64
	for index := range journal.Files {
		file := &journal.Files[index]
		if file.Action == "delete" {
			continue
		}
		installProgress(options, "download", "업데이트 파일 다운로드 중…", file.Path, index, len(journal.Files), completedBytes, totalBytes)
		var content []byte
		if file.Path == "app/release.json" {
			content = manifestRaw
		} else {
			content, err = source.get(ctx, file.Path, nil, MaxInstallFileBytes)
			if err != nil {
				return prepared, err
			}
		}
		if int64(len(content)) != file.Size || gitBlobSHA(content) != file.GitSHA {
			return prepared, fmt.Errorf("다운로드한 파일의 무결성을 확인하지 못했습니다: %s", file.Path)
		}
		if file.Deferred && !validExecutable(content) {
			return prepared, fmt.Errorf("업데이트 실행 파일의 PE 형식이 올바르지 않습니다")
		}
		file.SHA256 = payloadSHA(content)
		payload, err := safePath(root, stageRelative+"/payload/"+file.Path)
		if err != nil {
			return prepared, err
		}
		if err := atomicWrite(payload, content, 0600); err != nil {
			return prepared, err
		}
		completedBytes += file.Size
	}
	journal.State = StatePrepared
	if err := SaveJournal(root, journalPath, journal); err != nil {
		return prepared, err
	}
	prepared.Journal = journal
	installProgress(options, "prepared", "업데이트 파일 준비 완료", "", len(journal.Files), len(journal.Files), completedBytes, totalBytes)
	return prepared, nil
}
