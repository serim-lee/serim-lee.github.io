// Package citations refreshes the repository's citation snapshot from OpenAlex.
package citations

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Gonie-Gonie/Gonie-Gonie.github.io/editor/internal/dataversion"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const maxSafeInteger int64 = 9007199254740991

type KCICount struct {
	DOI   string
	Count int64
}

type Options struct {
	Root      string
	Client    *http.Client
	Endpoint  string
	Now       func() time.Time
	Log       func(string)
	KCIOnly   bool
	KCICounts []KCICount
}

type Result struct {
	Success bool
	Output  string
}

type console struct {
	output strings.Builder
	notify func(string)
}

func (c *console) line(format string, args ...any) {
	line := strings.ToValidUTF8(fmt.Sprintf(format, args...), "\uFFFD") + "\n"
	c.output.WriteString(line)
	if c.notify != nil {
		c.notify(line)
	}
}

var (
	doiPrefix = regexp.MustCompile(`(?i)^doi:\s*`)
	doiURL    = regexp.MustCompile(`(?i)^(?:https?://)?(?:dx\.)?doi\.org/`)
	doiFormat = regexp.MustCompile(`^10\.[0-9]{4,9}/[^\s]+$`)
	workID    = regexp.MustCompile(`^W[0-9]+$`)
)

func normalizeDOI(value any) string {
	doi := strings.TrimSpace(text(value))
	doi = strings.TrimSpace(doiPrefix.ReplaceAllString(doi, ""))
	doi = strings.ToLower(doiURL.ReplaceAllString(doi, ""))
	if !doiFormat.MatchString(doi) || strings.IndexFunc(doi, unicode.IsSpace) >= 0 {
		return ""
	}
	return doi
}

// Update preserves manual KCI records and unknown metadata. Individual OpenAlex
// request failures keep affected records, while later batches can still succeed.
func Update(ctx context.Context, options Options) (result Result, err error) {
	log := &console{notify: options.Log}
	defer func() { result.Output = log.output.String() }()
	root := options.Root
	if root == "" {
		root = "."
	}
	if err := dataversion.ValidateBundle(root); err != nil {
		return result, err
	}
	snapshotPath := filepath.Join(root, "exports", "publication-citations.json")
	snapshot, exists, err := readSnapshot(snapshotPath)
	if err == nil && !exists {
		legacy, legacyExists, legacyErr := readSnapshot(filepath.Join(root, "data", "publication-citations.json"))
		if legacyErr != nil {
			return result, legacyErr
		}
		if legacyExists {
			snapshot = legacy
		}
	}
	if err != nil {
		return result, err
	}
	before, err := json.Marshal(snapshot)
	if err != nil {
		return result, err
	}
	records := snapshot["records"].(map[string]any)
	previous := make(map[string]any, len(records))
	for doi, value := range records {
		record, ok := value.(map[string]any)
		if !ok {
			return result, fmt.Errorf("invalid citation snapshot record for %s: must be an object", doi)
		}
		previous[doi] = record["count"]
	}
	publicationsContent, err := os.ReadFile(filepath.Join(root, "data", "publications.json"))
	if err != nil {
		return result, err
	}
	publicationsRaw, err := dataversion.Unwrap(publicationsContent, "publications.json")
	if err != nil {
		return result, err
	}
	publications := publicationsRaw.([]any)

	published := make(map[string]map[string]any)
	var publicationOrder []string
	for index, value := range publications {
		item, ok := value.(map[string]any)
		if !ok {
			return result, fmt.Errorf("publications.json item %d must be an object", index+1)
		}
		doi := normalizeDOI(item["doi"])
		if doi == "" || truthy(item["under_review"]) || truthy(item["in_press"]) {
			continue
		}
		if _, exists := published[doi]; !exists {
			publicationOrder = append(publicationOrder, doi)
		}
		published[doi] = item
	}
	updates := make([]KCICount, 0, len(options.KCICounts))
	for _, update := range options.KCICounts {
		key := normalizeDOI(update.DOI)
		record, _ := records[key].(map[string]any)
		if _, eligible := published[key]; !eligible || record == nil || record["provider"] != "kci" {
			return result, fmt.Errorf("No published, mapped KCI article for %s", update.DOI)
		}
		if update.Count < 0 || update.Count > maxSafeInteger {
			return result, errors.New("KCI count must be a nonnegative safe integer")
		}
		updates = append(updates, KCICount{DOI: key, Count: update.Count})
	}
	endpoint := options.Endpoint
	if endpoint == "" {
		endpoint = "https://api.openalex.org/works"
	}
	baseURL, err := url.Parse(endpoint)
	if err != nil || baseURL.Host == "" || baseURL.Scheme != "http" && baseURL.Scheme != "https" {
		return result, errors.New("OpenAlex endpoint must be an http(s) URL")
	}
	client := options.Client
	if client == nil {
		client = &http.Client{Timeout: 12 * time.Second}
	}
	now := options.Now
	if now == nil {
		now = time.Now
	}
	var pending []string
	for doi := range published {
		record, _ := records[doi].(map[string]any)
		if record["provider"] != "kci" {
			pending = append(pending, doi)
		}
	}
	sort.Strings(pending)
	openalexVerified := make(map[string]bool)
	kciVerified := make(map[string]bool)
	failed := false
	if !options.KCIOnly {
		for offset := 0; offset < len(pending); offset += 100 {
			end := min(offset+100, len(pending))
			if err := refreshBatch(ctx, client, baseURL, pending[offset:end], records, openalexVerified, now); err != nil {
				log.line("OpenAlex refresh failed; existing counts and dates retained: %v", err)
				failed = true
			}
		}
	}
	for _, update := range updates {
		record := records[update.DOI].(map[string]any)
		record["count"] = update.Count
		record["checked_at"] = timestamp(now())
		kciVerified[update.DOI] = true
	}
	changed := make(map[string]bool)
	newlyRecorded := make(map[string]bool)
	for _, verified := range []map[string]bool{openalexVerified, kciVerified} {
		for doi := range verified {
			record := records[doi].(map[string]any)
			if previous[doi] == nil {
				newlyRecorded[doi] = true
			} else if !equalCount(previous[doi], record["count"].(int64)) {
				changed[doi] = true
			}
		}
	}
	for doi := range records {
		if _, eligible := published[doi]; !eligible {
			delete(records, doi)
		}
	}
	after, err := json.Marshal(snapshot)
	if err != nil {
		return result, err
	}
	if !exists || !bytes.Equal(before, after) {
		if err := dataversion.ValidateBundle(root); err != nil {
			return result, err
		}
		delete(snapshot, "version")
		delete(snapshot, "schema_version")
		content, err := json.MarshalIndent(snapshot, "", "  ")
		if err != nil {
			return result, err
		}
		if err := writeSnapshot(snapshotPath, append(content, '\n')); err != nil {
			return result, err
		}
	}
	for _, doi := range publicationOrder {
		if !changed[doi] && !newlyRecorded[doi] {
			continue
		}
		item := published[doi]
		title := strings.TrimSpace(text(item["title_ko"]))
		if title == "" {
			title = strings.TrimSpace(text(item["title_en"]))
		}
		if title == "" {
			title = doi
		}
		title = strings.Join(strings.Fields(title), " ")
		record := records[doi].(map[string]any)
		provider := "OpenAlex"
		if record["provider"] == "kci" {
			provider = "KCI"
		}
		old := text(previous[doi])
		if newlyRecorded[doi] {
			old = "not previously recorded"
		}
		log.line("[%s] %s: cite %s -> %d", provider, title, old, record["count"])
	}
	if len(changed) == 0 && len(newlyRecorded) == 0 {
		log.line("No citation count changes.")
	}
	log.line("OpenAlex: %d checked; KCI: %d manually verified; citation counts changed: %d; newly recorded: %d; total: %d.", len(openalexVerified), len(kciVerified), len(changed), len(newlyRecorded), len(records))
	result.Success = !failed
	return result, nil
}

func refreshBatch(ctx context.Context, client *http.Client, endpoint *url.URL, batch []string, records map[string]any, verified map[string]bool, now func() time.Time) error {
	requestURL := *endpoint
	query := requestURL.Query()
	query.Set("filter", "doi:"+strings.Join(batch, "|"))
	query.Set("select", "id,doi,cited_by_count")
	query.Set("per_page", "100")
	requestURL.RawQuery = query.Encode()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL.String(), nil)
	if err != nil {
		return err
	}
	request.Header.Set("Accept", "application/json")
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("HTTP %s", response.Status)
	}
	payload, err := decodeJSON(response.Body)
	if err != nil {
		return err
	}
	object, ok := payload.(map[string]any)
	if !ok {
		return errors.New("Invalid OpenAlex response")
	}
	results, ok := object["results"].([]any)
	if !ok {
		return errors.New("Invalid OpenAlex response")
	}
	allowed := make(map[string]bool, len(batch))
	for _, doi := range batch {
		allowed[doi] = true
	}
	checkedAt := timestamp(now())
	for _, value := range results {
		work, ok := value.(map[string]any)
		if !ok {
			return errors.New("Invalid OpenAlex work: must be an object")
		}
		doi := normalizeDOI(work["doi"])
		id := strings.TrimPrefix(text(work["id"]), "https://openalex.org/")
		count, validCount := safeCount(work["cited_by_count"])
		if !allowed[doi] || !workID.MatchString(id) || !validCount {
			continue
		}
		record, _ := records[doi].(map[string]any)
		if record == nil {
			record = make(map[string]any)
			records[doi] = record
		}
		record["provider"], record["work_id"], record["count"], record["checked_at"] = "openalex", id, count, checkedAt
		verified[doi] = true
	}
	return nil
}

func readSnapshot(path string) (map[string]any, bool, error) {
	value, err := readJSON(path)
	if errors.Is(err, os.ErrNotExist) {
		return map[string]any{"records": map[string]any{}}, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	snapshot, ok := value.(map[string]any)
	if !ok {
		return nil, false, errors.New("Invalid citation snapshot")
	}
	if _, present := snapshot["version"]; present {
		if legacy, ok := snapshot["version"].(json.Number); ok {
			if legacy.String() != "1" {
				return nil, false, errors.New("Invalid citation snapshot version")
			}
		} else if version, ok := snapshot["version"].(string); !ok || dataversion.Validate(version) != nil {
			return nil, false, errors.New("Invalid citation snapshot version")
		}
	}

	if _, ok := snapshot["records"].(map[string]any); !ok {
		return nil, false, errors.New("Invalid citation snapshot records")
	}
	return snapshot, true, nil
}

func readJSON(path string) (any, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if !utf8.Valid(content) {
		return nil, fmt.Errorf("%s is not valid UTF-8", path)
	}
	value, err := decodeJSON(bytes.NewReader(content))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return value, nil
}

func decodeJSON(reader io.Reader) (any, error) {
	decoder := json.NewDecoder(reader)
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return nil, errors.New("unexpected trailing JSON value")
		}
		return nil, err
	}
	return value, nil
}

func writeSnapshot(path string, content []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	mode := os.FileMode(0644)
	if info, err := os.Stat(path); err == nil {
		mode = info.Mode().Perm()
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".publication-citations-*.tmp")
	if err != nil {
		return err
	}
	temporary := file.Name()
	defer os.Remove(temporary)
	if err := file.Chmod(mode); err != nil {
		file.Close()
		return err
	}
	if _, err := file.Write(content); err != nil {
		file.Close()
		return err
	}
	if err := file.Sync(); err != nil {
		file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(temporary, path)
}

func safeCount(value any) (int64, bool) {
	number, ok := value.(json.Number)
	if !ok {
		return 0, false
	}
	count, err := strconv.ParseInt(string(number), 10, 64)
	return count, err == nil && count >= 0 && count <= maxSafeInteger
}

func equalCount(value any, count int64) bool {
	number, ok := value.(json.Number)
	if !ok {
		return false
	}
	previous, err := number.Float64()
	return err == nil && previous == float64(count)
}

func timestamp(value time.Time) string {
	return value.UTC().Format("2006-01-02T15:04:05+00:00")
}

func text(value any) string {
	if value == nil {
		return ""
	}
	if text, ok := value.(string); ok {
		return text
	}
	return fmt.Sprint(value)
}

func truthy(value any) bool {
	switch value := value.(type) {
	case nil:
		return false
	case bool:
		return value
	case string:
		return value != ""
	case json.Number:
		number, err := value.Float64()
		return err != nil || number != 0
	case []any:
		return len(value) > 0
	case map[string]any:
		return len(value) > 0
	default:
		return true
	}
}
