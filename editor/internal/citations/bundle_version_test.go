package citations

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLegacySnapshotIsReadButUpdatesOnlyWriteExports(t *testing.T) {
	root := newFixture(t, []any{testPaper(1)}, nil)
	legacyPath := filepath.Join(root, "data", "publication-citations.json")
	legacy := writeFixtureJSON(t, legacyPath, testSnapshot(map[string]any{testDOI(1): testRecord(4)}))
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"results": []any{resultWork(testDOI(1), 9)}})
	}))
	defer server.Close()
	result, err := Update(context.Background(), Options{Root: root, Endpoint: server.URL, Now: func() time.Time { return fixedNow }})
	if err != nil || !result.Success {
		t.Fatal(result, err)
	}
	stored, _ := os.ReadFile(filepath.Join(root, "exports", "publication-citations.json"))
	var snapshot map[string]any
	if err := json.Unmarshal(stored, &snapshot); err != nil {
		t.Fatal(err)
	}
	if snapshot["version"] != nil || snapshot["schema_version"] != nil || snapshot["records"] == nil {
		t.Fatal(string(stored))
	}
	after, _ := os.ReadFile(legacyPath)
	if !bytes.Equal(legacy, after) {
		t.Fatal("legacy source snapshot was overwritten")
	}
}

func TestIncompatibleBundleStopsCitationUpdateBeforeNetworkOrWrites(t *testing.T) {
	root := newFixture(t, []any{testPaper(1)}, testSnapshot(map[string]any{testDOI(1): testRecord(4)}))
	path := filepath.Join(root, "exports", "publication-citations.json")
	before, _ := os.ReadFile(path)
	if err := os.WriteFile(filepath.Join(root, "data", "version.json"), []byte(`{"version":"2.0.0"}`), 0644); err != nil {
		t.Fatal(err)
	}
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls++; w.WriteHeader(http.StatusInternalServerError) }))
	defer server.Close()
	if _, err := Update(context.Background(), Options{Root: root, Endpoint: server.URL}); err == nil {
		t.Fatal("incompatible bundle accepted")
	}
	after, _ := os.ReadFile(path)
	if calls != 0 || !bytes.Equal(before, after) {
		t.Fatal("rejected update had side effects")
	}
}
