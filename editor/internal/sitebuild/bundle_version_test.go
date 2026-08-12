package sitebuild

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIncompatibleBundleIsRejectedBeforeReadingDataOrGeneratingExports(t *testing.T) {
	root := t.TempDir()
	for _, directory := range []string{"data", "exports"} {
		if err := os.Mkdir(filepath.Join(root, directory), 0755); err != nil {
			t.Fatal(err)
		}
	}
	marker := []byte(`{"version":"2.0.0"}`)
	markerPath := filepath.Join(root, "data", "version.json")
	os.WriteFile(markerPath, marker, 0644)
	outputPath := filepath.Join(root, "exports", "build-report.json")
	before := []byte("existing report")
	os.WriteFile(outputPath, before, 0644)
	if _, err := Build(root, false); err == nil || !strings.Contains(err.Error(), "unsupported data version") {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(outputPath)
	markerAfter, _ := os.ReadFile(markerPath)
	if !bytes.Equal(before, after) || !bytes.Equal(marker, markerAfter) {
		t.Fatal("rejected build changed files")
	}
}
