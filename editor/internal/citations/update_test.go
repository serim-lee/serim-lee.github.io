package citations

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
	"sort"
	"strings"
	"testing"
	"time"
)

var fixedNow = time.Date(2026, 10, 19, 15, 4, 5, 800, time.FixedZone("KST", 9*60*60))

const checkedAt = "2026-10-19T06:04:05+00:00"

func testDOI(number int) string { return fmt.Sprintf("10.1234/paper%03d", number) }

func testPaper(number int) map[string]any {
	return map[string]any{"doi": testDOI(number), "title_en": fmt.Sprintf("Paper %d", number)}
}

func testRecord(count int64) map[string]any {
	return map[string]any{"provider": "openalex", "work_id": "W100", "count": count, "checked_at": "2025-01-01T00:00:00+00:00"}
}

func testSnapshot(records map[string]any) map[string]any {
	return map[string]any{"version": 1, "records": records}
}

func writeFixtureJSON(t *testing.T, path string, value any) []byte {
	t.Helper()
	content, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	content = append(content, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}
	return content
}

func newFixture(t *testing.T, publications any, snapshot any) string {
	t.Helper()
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "data"), 0755); err != nil {
		t.Fatal(err)
	}
	writeFixtureJSON(t, filepath.Join(root, "data", "publications.json"), publications)
	if snapshot != nil {
		writeFixtureJSON(t, filepath.Join(root, "exports", "publication-citations.json"), snapshot)
	}
	return root
}

func readFixtureSnapshot(t *testing.T, root string) map[string]any {
	t.Helper()
	snapshot, _, err := readSnapshot(filepath.Join(root, "exports", "publication-citations.json"))
	if err != nil {
		t.Fatal(err)
	}
	return snapshot
}

func fixtureCount(t *testing.T, records map[string]any, doi string) int64 {
	t.Helper()
	count, ok := safeCount(records[doi].(map[string]any)["count"])
	if !ok {
		t.Fatalf("invalid stored count for %s", doi)
	}
	return count
}

func resultWork(doi string, count any) map[string]any {
	return map[string]any{"id": "https://openalex.org/W123", "doi": doi, "cited_by_count": count}
}

func TestUpdateCountsPreservesMetadataAndPrunesUnpublishedRecords(t *testing.T) {
	first, second, third, fourth, unmatched, kci := testPaper(1), testPaper(2), testPaper(3), testPaper(4), testPaper(5), testPaper(6)
	first["doi"], first["title_ko"] = " DOI: https://dx.doi.org/10.1234/PAPER001 ", " 첫째\n 논문 "
	delete(third, "title_en")
	review, inPress := testPaper(7), testPaper(8)
	review["under_review"], inPress["in_press"] = true, true
	oldFirst := testRecord(1)
	oldFirst["future"] = map[string]any{"retained": true}
	kciRecord := map[string]any{"provider": "kci", "article_id": "ART123", "count": 9, "checked_at": "2024-01-01", "list_pending": true}
	records := map[string]any{
		testDOI(1): oldFirst, testDOI(4): testRecord(2), testDOI(5): testRecord(7), testDOI(6): kciRecord,
		testDOI(7): testRecord(20), testDOI(8): testRecord(30), testDOI(99): testRecord(99), "10.1234/PAPER001": testRecord(90),
	}
	snapshot := testSnapshot(records)
	snapshot["future_root"] = map[string]any{"value": "한글 보존"}
	root := newFixture(t, []any{first, second, third, fourth, unmatched, kci, review, inPress, map[string]any{"doi": "invalid"}}, snapshot)
	publicationPath := filepath.Join(root, "data", "publications.json")
	publicationBefore, _ := os.ReadFile(publicationPath)
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		query := r.URL.Query()
		if r.Method != http.MethodGet || r.Header.Get("Accept") != "application/json" || query.Get("select") != "id,doi,cited_by_count" || query.Get("per_page") != "100" || query.Has("api_key") || r.Header.Get("Authorization") != "" {
			t.Errorf("unexpected keyless request: %s %#v", r.URL, r.Header)
		}
		if filter := query.Get("filter"); filter != "doi:"+strings.Join([]string{testDOI(1), testDOI(2), testDOI(3), testDOI(4), testDOI(5)}, "|") {
			t.Errorf("wrong eligible batch: %s", filter)
		}
		json.NewEncoder(w).Encode(map[string]any{"results": []any{
			resultWork("https://doi.org/10.1234/PAPER001", 3), resultWork(testDOI(2), 0), resultWork(testDOI(3), 4), resultWork(testDOI(4), 0),
			resultWork(testDOI(99), 200), resultWork(testDOI(6), 1000),
		}})
	}))
	defer server.Close()
	var streamed strings.Builder
	result, err := Update(context.Background(), Options{Root: root, Endpoint: server.URL, Now: func() time.Time { return fixedNow }, Log: func(line string) {
		if !strings.HasSuffix(line, "\n") {
			t.Error("console callback did not receive a full line")
		}
		streamed.WriteString(line)
	}})
	if err != nil || !result.Success || calls != 1 || streamed.String() != result.Output {
		t.Fatalf("update failed: %#v, %v, calls=%d", result, err, calls)
	}
	for _, expected := range []string{
		"[OpenAlex] 첫째 논문: cite 1 -> 3\n", "[OpenAlex] Paper 2: cite not previously recorded -> 0\n",
		"[OpenAlex] " + testDOI(3) + ": cite not previously recorded -> 4\n", "[OpenAlex] Paper 4: cite 2 -> 0\n",
		"OpenAlex: 4 checked; KCI: 0 manually verified; citation counts changed: 2; newly recorded: 2; total: 6.\n",
	} {
		if !strings.Contains(result.Output, expected) {
			t.Errorf("missing output %q in %s", expected, result.Output)
		}
	}
	stored := readFixtureSnapshot(t, root)
	storedRecords := stored["records"].(map[string]any)
	if len(storedRecords) != 6 || fixtureCount(t, storedRecords, testDOI(2)) != 0 || fixtureCount(t, storedRecords, testDOI(4)) != 0 || fixtureCount(t, storedRecords, testDOI(5)) != 7 {
		t.Fatalf("wrong snapshot records: %#v", storedRecords)
	}
	if storedRecords[testDOI(1)].(map[string]any)["checked_at"] != checkedAt || storedRecords[testDOI(1)].(map[string]any)["future"] == nil || stored["future_root"] == nil {
		t.Fatal("UTC check date or unknown metadata was lost")
	}
	for _, doi := range []string{testDOI(5), testDOI(6)} {
		before, _ := json.Marshal(records[doi])
		after, _ := json.Marshal(storedRecords[doi])
		if !bytes.Equal(before, after) {
			t.Fatalf("untouched record %s changed: %s != %s", doi, before, after)
		}
	}
	publicationAfter, _ := os.ReadFile(publicationPath)
	if !bytes.Equal(publicationBefore, publicationAfter) {
		t.Fatal("updater modified publications.json")
	}
}

func TestUpdateBatchesSortedDOIsAndRejectsWrongBatchResults(t *testing.T) {
	var papers []any
	for number := 205; number > 0; number-- {
		papers = append(papers, testPaper(number))
	}
	root := newFixture(t, papers, nil)
	var sizes []int
	var observed []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		batch := strings.Split(strings.TrimPrefix(r.URL.Query().Get("filter"), "doi:"), "|")
		sizes = append(sizes, len(batch))
		observed = append(observed, batch...)
		var results []any
		for _, doi := range batch {
			results = append(results, resultWork(doi, 0))
		}
		if len(sizes) == 1 {
			results = append(results, resultWork(testDOI(205), 999))
		}
		json.NewEncoder(w).Encode(map[string]any{"results": results})
	}))
	defer server.Close()
	result, err := Update(context.Background(), Options{Root: root, Endpoint: server.URL, Now: func() time.Time { return fixedNow }})
	if err != nil || !result.Success || !reflect.DeepEqual(sizes, []int{100, 100, 5}) || !sort.StringsAreSorted(observed) {
		t.Fatalf("incorrect batching: %v, %#v, %v", sizes, result, err)
	}
	records := readFixtureSnapshot(t, root)["records"].(map[string]any)
	if len(records) != 205 || fixtureCount(t, records, testDOI(205)) != 0 || strings.Contains(result.Output, "999") {
		t.Fatal("foreign batch response was applied")
	}
}

func TestUpdateIgnoresMalformedWorkFieldsAndAcceptsSafeIntegerZero(t *testing.T) {
	cases := []struct {
		name  string
		value any
		valid bool
	}{
		{"zero", json.Number("0"), true}, {"safe maximum", json.Number("9007199254740991"), true},
		{"negative", json.Number("-1"), false}, {"too large", json.Number("9007199254740992"), false},
		{"fraction", json.Number("1.2"), false}, {"decimal integer", json.Number("1.0"), false}, {"exponent", json.Number("1e0"), false},
		{"boolean", true, false}, {"null", nil, false}, {"string", "1", false},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			root := newFixture(t, []any{testPaper(1)}, testSnapshot(map[string]any{}))
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				badID := resultWork(testDOI(1), 77)
				badID["id"] = "https://evil.example/W123"
				json.NewEncoder(w).Encode(map[string]any{"results": []any{badID, resultWork(testDOI(99), 77), resultWork(testDOI(1), test.value)}})
			}))
			defer server.Close()
			result, err := Update(context.Background(), Options{Root: root, Endpoint: server.URL, Now: func() time.Time { return fixedNow }})
			if err != nil || !result.Success {
				t.Fatalf("invalid work fields caused batch failure: %#v, %v", result, err)
			}
			records := readFixtureSnapshot(t, root)["records"].(map[string]any)
			if (len(records) == 1) != test.valid {
				t.Fatalf("wrong field validation: %#v", records)
			}
		})
	}
}

func TestUpdateUnchangedCountsRefreshDateWithoutSpuriousChangeLog(t *testing.T) {
	record := testRecord(2)
	root := newFixture(t, []any{testPaper(1)}, testSnapshot(map[string]any{testDOI(1): record}))
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"results": []any{resultWork(testDOI(1), 2)}})
	}))
	defer server.Close()
	result, err := Update(context.Background(), Options{Root: root, Endpoint: server.URL, Now: func() time.Time { return fixedNow }})
	if err != nil || !result.Success || !strings.HasPrefix(result.Output, "No citation count changes.\n") || strings.Contains(result.Output, "[OpenAlex]") {
		t.Fatalf("same count reported as changed: %#v, %v", result, err)
	}
	path := filepath.Join(root, "exports", "publication-citations.json")
	oldTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	if err := os.Chtimes(path, oldTime, oldTime); err != nil {
		t.Fatal(err)
	}
	before, _ := os.ReadFile(path)
	result, err = Update(context.Background(), Options{Root: root, Endpoint: server.URL, Now: func() time.Time { return fixedNow }})
	info, statErr := os.Stat(path)
	after, _ := os.ReadFile(path)
	if err != nil || !result.Success || statErr != nil || !info.ModTime().Equal(oldTime) || !bytes.Equal(before, after) {
		t.Fatal("identical snapshot content was rewritten")
	}
}

func TestUpdateKCIOnlyValidatesAllManualUpdatesBeforeWrites(t *testing.T) {
	record := map[string]any{"provider": "kci", "article_id": "ART123", "count": 2, "checked_at": "2024-01-01", "future": "retain"}
	root := newFixture(t, []any{testPaper(1), testPaper(2)}, testSnapshot(map[string]any{testDOI(1): record, testDOI(2): testRecord(5)}))
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls++; t.Error("KCI-only update made a request") }))
	defer server.Close()
	path := filepath.Join(root, "exports", "publication-citations.json")
	before, _ := os.ReadFile(path)
	for _, invalid := range []KCICount{{DOI: testDOI(2), Count: 3}, {DOI: testDOI(99), Count: 3}, {DOI: "bad", Count: 3}, {DOI: testDOI(1), Count: -1}, {DOI: testDOI(1), Count: maxSafeInteger + 1}} {
		result, err := Update(context.Background(), Options{Root: root, Endpoint: server.URL, KCIOnly: true, KCICounts: []KCICount{{DOI: testDOI(1), Count: 6}, invalid}})
		after, _ := os.ReadFile(path)
		if err == nil || result.Success || result.Output != "" || !bytes.Equal(before, after) || calls != 0 {
			t.Fatalf("manual validation changed state: %#v, %v", result, err)
		}
	}
	result, err := Update(context.Background(), Options{Root: root, Endpoint: server.URL, KCIOnly: true, Now: func() time.Time { return fixedNow }, KCICounts: []KCICount{{DOI: "doi: " + testDOI(1), Count: 3}, {DOI: testDOI(1), Count: 0}}})
	if err != nil || !result.Success || calls != 0 || !strings.Contains(result.Output, "[KCI] Paper 1: cite 2 -> 0") || !strings.Contains(result.Output, "KCI: 1 manually verified") {
		t.Fatalf("manual KCI update failed: %#v, %v", result, err)
	}
	stored := readFixtureSnapshot(t, root)["records"].(map[string]any)[testDOI(1)].(map[string]any)
	if stored["article_id"] != "ART123" || stored["future"] != "retain" || stored["checked_at"] != checkedAt {
		t.Fatal("manual KCI update lost metadata")
	}
}

func TestUpdatePartialHTTPFailureContinuesBatchesAndRetainsFailedRecords(t *testing.T) {
	var papers []any
	records := make(map[string]any)
	for number := 1; number <= 101; number++ {
		papers = append(papers, testPaper(number))
		records[testDOI(number)] = testRecord(1)
	}
	root := newFixture(t, papers, testSnapshot(records))
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			http.Error(w, "temporary failure", http.StatusServiceUnavailable)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"results": []any{resultWork(testDOI(101), 2)}})
	}))
	defer server.Close()
	var log strings.Builder
	result, err := Update(context.Background(), Options{Root: root, Endpoint: server.URL, Now: func() time.Time { return fixedNow }, Log: func(line string) { log.WriteString(line) }})
	if err != nil || result.Success || calls != 2 || result.Output != log.String() || !strings.Contains(result.Output, "HTTP 503") || !strings.Contains(result.Output, "citation counts changed: 1") {
		t.Fatalf("partial failure was not retained: %#v, %v", result, err)
	}
	stored := readFixtureSnapshot(t, root)["records"].(map[string]any)
	if fixtureCount(t, stored, testDOI(1)) != 1 || stored[testDOI(1)].(map[string]any)["checked_at"] != "2025-01-01T00:00:00+00:00" || fixtureCount(t, stored, testDOI(101)) != 2 {
		t.Fatal("failed batch changed, or successful later batch was discarded")
	}
}

func TestUpdateCanceledRequestRetainsExistingSnapshot(t *testing.T) {
	root := newFixture(t, []any{testPaper(1)}, testSnapshot(map[string]any{testDOI(1): testRecord(4)}))
	path := filepath.Join(root, "exports", "publication-citations.json")
	before, _ := os.ReadFile(path)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cancel()
		<-r.Context().Done()
	}))
	defer server.Close()
	result, err := Update(ctx, Options{Root: root, Endpoint: server.URL})
	after, _ := os.ReadFile(path)
	if err != nil || result.Success || !strings.Contains(result.Output, "context canceled") || !bytes.Equal(before, after) {
		t.Fatalf("cancellation changed snapshot or lost diagnostics: %#v, %v", result, err)
	}
}

func TestUpdateMalformedResponseReportsFailureAndKeepsEarlierValidWorks(t *testing.T) {
	for _, body := range []string{`null`, `{}`, `{"results":null}`, `{"results":[]`, `{"results":[]} {}`, `{"results":[{"id":"W123","doi":"10.1234/paper001","cited_by_count":2},null]}`} {
		t.Run(body, func(t *testing.T) {
			root := newFixture(t, []any{testPaper(1)}, testSnapshot(map[string]any{testDOI(1): testRecord(1)}))
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, body) }))
			defer server.Close()
			result, err := Update(context.Background(), Options{Root: root, Endpoint: server.URL, Now: func() time.Time { return fixedNow }})
			if err != nil || result.Success || !strings.Contains(result.Output, "OpenAlex refresh failed") {
				t.Fatalf("malformed response accepted: %#v, %v", result, err)
			}
			expected := int64(1)
			if strings.Contains(body, `cited_by_count`) {
				expected = 2
			}
			if fixtureCount(t, readFixtureSnapshot(t, root)["records"].(map[string]any), testDOI(1)) != expected {
				t.Fatal("partially parsed valid works were handled incorrectly")
			}
		})
	}
}

func TestUpdateRejectsInvalidSnapshotAndPublicationMetadataBeforeNetwork(t *testing.T) {
	for _, test := range []struct {
		name         string
		snapshot     any
		publications any
	}{
		{"snapshot array", []any{}, []any{}}, {"wrong version", map[string]any{"version": 2, "records": map[string]any{}}, []any{}},
		{"null records", map[string]any{"version": 1, "records": nil}, []any{}},
		{"nonobject record", testSnapshot(map[string]any{testDOI(1): nil}), []any{testPaper(1)}},
		{"publications object", testSnapshot(map[string]any{}), map[string]any{}},
		{"nonobject publication", testSnapshot(map[string]any{}), []any{nil}},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := newFixture(t, test.publications, test.snapshot)
			path := filepath.Join(root, "exports", "publication-citations.json")
			before, _ := os.ReadFile(path)
			result, err := Update(context.Background(), Options{Root: root, Endpoint: "http://127.0.0.1:1"})
			after, _ := os.ReadFile(path)
			if err == nil || result.Success || result.Output != "" || !bytes.Equal(before, after) {
				t.Fatalf("invalid metadata accepted or changed files: %#v, %v", result, err)
			}
		})
	}
}

func TestAtomicSnapshotFailureCleansTemporaryFileAndPreservesDestination(t *testing.T) {
	root := t.TempDir()
	destination := filepath.Join(root, "snapshot.json")
	if err := os.Mkdir(destination, 0755); err != nil {
		t.Fatal(err)
	}
	sentinel := filepath.Join(destination, "preserved")
	if err := os.WriteFile(sentinel, []byte("previous snapshot"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := writeSnapshot(destination, []byte("replacement")); err == nil {
		t.Fatal("atomic replacement of a nonempty directory unexpectedly succeeded")
	}
	contents, err := os.ReadFile(sentinel)
	entries, listErr := os.ReadDir(root)
	if err != nil || string(contents) != "previous snapshot" || listErr != nil || len(entries) != 1 {
		t.Fatal("failed atomic write damaged the previous destination or leaked a temporary file")
	}
}

func TestUpdateFileFailureReturnsPriorConsoleOutput(t *testing.T) {
	var papers []any
	for number := 1; number <= 101; number++ {
		papers = append(papers, testPaper(number))
	}
	root := newFixture(t, papers, testSnapshot(map[string]any{}))
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			http.Error(w, "unavailable", http.StatusServiceUnavailable)
			return
		}
		if err := os.Rename(filepath.Join(root, "exports"), filepath.Join(root, "previous-exports")); err != nil {
			t.Error(err)
		}
		if err := os.WriteFile(filepath.Join(root, "exports"), []byte("blocking path"), 0644); err != nil {
			t.Error(err)
		}
		json.NewEncoder(w).Encode(map[string]any{"results": []any{resultWork(testDOI(101), 2)}})
	}))
	defer server.Close()
	result, err := Update(context.Background(), Options{Root: root, Endpoint: server.URL, Now: func() time.Time { return fixedNow }})
	if err == nil || result.Success || !strings.Contains(result.Output, "HTTP 503") || strings.Contains(result.Output, "total:") {
		t.Fatalf("file failure lost prior console output or reported completion: %#v, %v", result, err)
	}
	previous, err := readJSON(filepath.Join(root, "previous-exports", "publication-citations.json"))
	if err != nil || len(previous.(map[string]any)["records"].(map[string]any)) != 0 {
		t.Fatal("failed write replaced the previous snapshot")
	}
}

func TestDOINormalization(t *testing.T) {
	for _, value := range []string{"10.1234/PAPER001", "doi: 10.1234/PAPER001", "doi:\u00a010.1234/PAPER001", "https://doi.org/10.1234/PAPER001", "https://dx.doi.org/10.1234/PAPER001", "doi.org/10.1234/PAPER001 "} {
		if got := normalizeDOI(value); got != testDOI(1) {
			t.Errorf("normalizeDOI(%q) = %q", value, got)
		}
	}
	for _, value := range []string{"", "bad", "10.123/short", "https://evil.example/10.1234/paper001", "10.1234/two words", "10.1234/two\u00a0words"} {
		if got := normalizeDOI(value); got != "" {
			t.Errorf("normalizeDOI(%q) accepted %q", value, got)
		}
	}
}
