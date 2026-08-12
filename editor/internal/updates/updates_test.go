package updates

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Gonie-Gonie/Gonie-Gonie.github.io/editor/internal/release"
)

func metadataFixture(t *testing.T, version string) string {
	t.Helper()
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "app"), 0755); err != nil {
		t.Fatal(err)
	}
	writeMetadata(t, root, version, Source{Repository: DefaultRepository, Ref: DefaultRef})
	return root
}

func writeMetadata(t *testing.T, root, version string, source Source) {
	t.Helper()
	raw, err := json.Marshal(map[string]any{"version": version, "upstream": source})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "app", "release.json"), raw, 0644); err != nil {
		t.Fatal(err)
	}
}

func entry(version string) release.Entry {
	return release.Entry{Version: version, Date: "2026-10-28", Notes: "app/releases/" + version + ".md"}
}
func encoded(value any) string         { raw, _ := json.Marshal(value); return string(raw) }
func note(version, text string) string { return "# " + version + "\n\n2026-10-28\n\n- " + text + "\n" }

type requestLog struct {
	mu    sync.Mutex
	paths []string
}

func (log *requestLog) add(path string) {
	log.mu.Lock()
	defer log.mu.Unlock()
	log.paths = append(log.paths, path)
}
func (log *requestLog) all() []string {
	log.mu.Lock()
	defer log.mu.Unlock()
	return append([]string{}, log.paths...)
}

func sourceServer(t *testing.T, files map[string]string) (*httptest.Server, *requestLog) {
	t.Helper()
	log := &requestLog{}
	prefix := "/" + DefaultRepository + "/" + DefaultRef + "/"
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.add(r.URL.Path)
		if r.Method != http.MethodGet {
			t.Errorf("unexpected method %s", r.Method)
			w.WriteHeader(405)
			return
		}
		if !strings.HasPrefix(r.URL.Path, prefix) {
			t.Errorf("request outside original source: %s", r.URL.Path)
			w.WriteHeader(404)
			return
		}
		if value, ok := files[strings.TrimPrefix(r.URL.Path, prefix)]; ok {
			fmt.Fprint(w, value)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(server.Close)
	return server, log
}

func fixtureFiles(latest string, versions ...string) map[string]string {
	entries := []release.Entry{}
	files := map[string]string{"app/release.json": encoded(release.Manifest{Entry: entry(latest), DataVersion: "1.0.0"})}
	for _, version := range versions {
		entries = append(entries, entry(version))
		files[entry(version).Notes] = note(version, "Changes in "+version)
	}
	files["app/releases/index.json"] = encoded(release.Index{Version: "1.0.0", Releases: entries})
	return files
}

func TestGetInfoReadsOnlyLocalMetadataAndKeepsConfiguredOriginalSource(t *testing.T) {
	root := metadataFixture(t, "1.9.0")
	if err := os.Mkdir(filepath.Join(root, ".git"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".git", "config"), []byte("[remote \"origin\"]\nurl=https://github.com/fork-owner/personal-site.git\n"), 0644); err != nil {
		t.Fatal(err)
	}
	before, _ := os.ReadFile(filepath.Join(root, "app", "release.json"))
	info, err := GetInfo(root)
	if err != nil {
		t.Fatal(err)
	}
	if info.CurrentVersion != "1.9.0" || info.SourceRepository != DefaultRepository || info.SourceRef != "main" || info.SourceURL != "https://github.com/"+DefaultRepository+"/tree/main" {
		t.Fatal(info)
	}
	after, _ := os.ReadFile(filepath.Join(root, "app", "release.json"))
	if !bytes.Equal(before, after) {
		t.Fatal("metadata changed")
	}
	if _, err := os.Stat(filepath.Join(root, "data")); !os.IsNotExist(err) {
		t.Fatal("local info required or created user data")
	}
	if err := os.WriteFile(filepath.Join(root, "app", "release.json"), []byte(`{"version":"1.0.0"}`), 0644); err != nil {
		t.Fatal(err)
	}
	if fallback, err := GetInfo(root); err != nil || fallback.SourceRepository != DefaultRepository {
		t.Fatal(fallback, err)
	}
}

func TestNumericVersionComparisonAndApplicableHistoryFiltering(t *testing.T) {
	root := metadataFixture(t, "1.9.0")
	files := fixtureFiles("2.0.0", "1.9.0", "3.0.0", "1.9.1", "2.0.0", "1.10.0", "0.9.0")
	files[entry("1.10.0").Notes] = "# 1.10.0\n\n2026-10-28\n\n- First change\n  with a continued description.\n\n```md\n- Hidden code example\n```\n* Second change <script>literal</script>\n"
	server, log := sourceServer(t, files)
	before, _ := os.ReadFile(filepath.Join(root, "app", "release.json"))
	result, err := Check(context.Background(), Options{Root: root, BaseURL: server.URL, Client: server.Client()})
	if err != nil {
		t.Fatal(err)
	}
	if !result.UpdateAvailable || result.CurrentVersion != "1.9.0" || result.LatestVersion != "2.0.0" {
		t.Fatal(result)
	}
	versions := []string{}
	for _, release := range result.Releases {
		versions = append(versions, release.Version)
	}
	if !reflect.DeepEqual(versions, []string{"2.0.0", "1.10.0", "1.9.1"}) {
		t.Fatal(versions)
	}
	if !reflect.DeepEqual(result.Releases[1].Notes, []string{"First change with a continued description.", "Second change <script>literal</script>"}) {
		t.Fatal(result.Releases[1].Notes)
	}
	requests := strings.Join(log.all(), "\n")
	for _, excluded := range []string{"/1.9.0.md", "/3.0.0.md", "/0.9.0.md"} {
		if strings.Contains(requests, excluded) {
			t.Fatalf("fetched excluded notes: %s", excluded)
		}
	}
	after, _ := os.ReadFile(filepath.Join(root, "app", "release.json"))
	if !bytes.Equal(before, after) {
		t.Fatal("update check wrote metadata")
	}
	if len(log.all()) != 5 {
		t.Fatal(log.all())
	}
}

func TestEqualOrLocallyAheadReturnsNoHistory(t *testing.T) {
	for _, current := range []string{"1.10.0", "2.0.0"} {
		t.Run(current, func(t *testing.T) {
			root := metadataFixture(t, current)
			server, log := sourceServer(t, fixtureFiles("1.10.0"))
			result, err := Check(context.Background(), Options{Root: root, BaseURL: server.URL, Client: server.Client()})
			if err != nil || result.UpdateAvailable || result.LatestVersion != "1.10.0" || len(result.Releases) != 0 || len(log.all()) != 1 {
				t.Fatal(result, err, log.all())
			}
		})
	}
}

func TestAnyFailedRequiredFilePreventsAvailabilityAndPartialHistory(t *testing.T) {
	mutations := map[string]func(map[string]string){
		"manifest 404":          func(files map[string]string) { delete(files, "app/release.json") },
		"index 404":             func(files map[string]string) { delete(files, "app/releases/index.json") },
		"note 404":              func(files map[string]string) { delete(files, entry("1.1.0").Notes) },
		"invalid manifest JSON": func(files map[string]string) { files["app/release.json"] = `{"version":` },
		"invalid version": func(files map[string]string) {
			files["app/release.json"] = encoded(release.Manifest{Entry: entry("bad")})
		},
		"invalid date": func(files map[string]string) {
			e := entry("1.2.0")
			e.Date = "2026-02-29"
			files["app/release.json"] = encoded(release.Manifest{Entry: e})
		},
		"unsafe note path": func(files map[string]string) {
			e := entry("1.2.0")
			e.Notes = "../data/profile.json"
			files["app/release.json"] = encoded(release.Manifest{Entry: e})
		},
		"missing latest index entry": func(files map[string]string) {
			files["app/releases/index.json"] = encoded(release.Index{Version: "1.0.0", Releases: []release.Entry{entry("1.1.0")}})
		},
		"duplicate release": func(files map[string]string) {
			files["app/releases/index.json"] = encoded(release.Index{Version: "1.0.0", Releases: []release.Entry{entry("1.2.0"), entry("1.2.0")}})
		},
		"inconsistent latest": func(files map[string]string) {
			e := entry("1.2.0")
			e.Date = "2026-10-27"
			files["app/releases/index.json"] = encoded(release.Index{Version: "1.0.0", Releases: []release.Entry{e}})
		},
		"unsupported index format": func(files map[string]string) {
			files["app/releases/index.json"] = encoded(release.Index{Version: "2.0.0", Releases: []release.Entry{entry("1.2.0")}})
		},
		"empty change list":      func(files map[string]string) { files[entry("1.2.0").Notes] = "# 1.2.0\n\n2026-10-28\n\nNo bullets." },
		"incorrect note version": func(files map[string]string) { files[entry("1.2.0").Notes] = note("1.0.0", "Wrong version") },
		"invalid note UTF8":      func(files map[string]string) { files[entry("1.2.0").Notes] = note("1.2.0", "Bad") + "\xff" },
		"large manifest":         func(files map[string]string) { files["app/release.json"] = strings.Repeat("x", MaxManifestBytes+1) },
		"large index":            func(files map[string]string) { files["app/releases/index.json"] = strings.Repeat("x", MaxIndexBytes+1) },
		"large note":             func(files map[string]string) { files[entry("1.2.0").Notes] = strings.Repeat("x", MaxNoteBytes+1) },
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			root := metadataFixture(t, "1.0.0")
			files := fixtureFiles("1.2.0", "1.2.0", "1.1.0")
			mutate(files)
			server, _ := sourceServer(t, files)
			result, err := Check(context.Background(), Options{Root: root, BaseURL: server.URL, Client: server.Client()})
			if err == nil || result.UpdateAvailable || len(result.Releases) != 0 || result.CurrentVersion != "1.0.0" {
				t.Fatal(result, err)
			}
			if name == "manifest 404" && !strings.Contains(err.Error(), "HTTP 404") {
				t.Fatal(err)
			}
		})
	}
}

func TestSourceValidationAndHTTPSRedirectBoundaries(t *testing.T) {
	root := metadataFixture(t, "1.0.0")
	for _, source := range []Source{{Repository: "https://github.com/owner/repo", Ref: "main"}, {Repository: "owner/../repo", Ref: "main"}, {Repository: "owner/repo", Ref: "../main"}, {Repository: "owner/repo", Ref: "main?token=x"}, {Repository: "owner/repo", Ref: `main\branch`}, {Repository: "owner/repo", Ref: "main//branch"}} {
		writeMetadata(t, root, "1.0.0", source)
		if _, err := GetInfo(root); err == nil {
			t.Fatal("unsafe source accepted", source)
		}
	}
	writeMetadata(t, root, "1.0.0", Source{Repository: DefaultRepository, Ref: DefaultRef})
	if _, err := Check(context.Background(), Options{Root: root, BaseURL: "http://example.invalid"}); err == nil {
		t.Fatal("insecure base URL accepted")
	}
	requests := 0
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		http.Redirect(w, r, "http://example.invalid/stolen", http.StatusFound)
	}))
	defer server.Close()
	if result, err := Check(context.Background(), Options{Root: root, BaseURL: server.URL, Client: server.Client()}); err == nil || result.UpdateAvailable || requests != 1 {
		t.Fatal(result, err, requests)
	}
}

func TestCanceledContextStopsRequests(t *testing.T) {
	root := metadataFixture(t, "1.0.0")
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { <-r.Context().Done() }))
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	started := time.Now()
	result, err := Check(ctx, Options{Root: root, BaseURL: server.URL, Client: server.Client()})
	if err == nil || result.UpdateAvailable || time.Since(started) > 2*time.Second {
		t.Fatal(result, err)
	}
}

func TestParserIgnoresFencedExamplesAndRequiresRealChanges(t *testing.T) {
	e := entry("1.2.0")
	for _, raw := range []string{"# 1.2.0\n2026-10-28\n```\n- example\n```\n", "# 1.2.0\n- Missing date\n"} {
		if _, err := parseNotes([]byte(raw), e); err == nil {
			t.Fatal("invalid note accepted")
		}
	}
	items, err := parseNotes([]byte("# 1.2.0\r\n\r\n2026-10-28\r\n\r\n+ One\r\n~~~\r\n- ignored\r\n~~~\r\n- Two\r\n"), e)
	if err != nil || !reflect.DeepEqual(items, []string{"One", "Two"}) {
		t.Fatal(items, err)
	}
}
