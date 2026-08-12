package dataversion

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestStableVersionsAndCompatibility(t *testing.T) {
	for _, value := range []string{"1.0.0", "1.2.30", "1.999.1"} {
		if err := Validate(value); err != nil {
			t.Fatal(err)
		}
	}
	for _, value := range []string{"1", "1.0", "01.0.0", "1.01.0", "1.0.00", "1.0.0-beta", "2.0.0", "0.9.0", " 1.0.0"} {
		if Validate(value) == nil {
			t.Fatalf("accepted %q", value)
		}
	}
	if n, err := Compare("1.10.0", "1.9.9"); err != nil || n != 1 {
		t.Fatal(n, err)
	}
}

func TestLegacyArrayEnvelopesBecomeRawArraysWithoutLosingRows(t *testing.T) {
	previous := []byte(`{"version":"1.2.3","items":[{"id":"second","future_row":42},{"id":"first"}]}`)
	rows, err := Array(previous, "people.json")
	if err != nil {
		t.Fatal(err)
	}
	rows[0], rows[1] = rows[1], rows[0]
	encoded, err := MarshalArray(previous, rows, "people.json")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(encoded, []byte("[")) {
		t.Fatal("array envelope was written")
	}
	decoded, _ := Array(encoded, "people.json")
	if !bytes.Contains(decoded[0], []byte("first")) || !bytes.Contains(decoded[1], []byte("future_row")) {
		t.Fatal(string(encoded))
	}
	unsafe := []byte(`{"version":"1.2.3","future":{"retain":true},"items":[]}`)
	if _, err := MarshalArray(unsafe, []json.RawMessage{}, "people.json"); err == nil {
		t.Fatal("unknown envelope metadata silently discarded")
	}
	for _, bad := range []string{`null`, `{"version":"1.0.0","items":null}`, `{"version":"2.0.0","items":[]}`, `{"version":1,"items":[]}`, `{"version":"1.0.0"}`, `[] []`} {
		if _, err := Array([]byte(bad), "people.json"); err == nil {
			t.Fatal("accepted", bad)
		}
	}
}

func TestLegacyObjectMetadataIsRemovedButContentSurvives(t *testing.T) {
	for name, schema := range map[string]int{"settings.json": 9, "profile.json": 1, "academic_activities.json": 1} {
		raw, _ := json.Marshal(map[string]any{"schema_version": schema, "future": "keep"})
		root, info, err := Object(raw, name)
		if err != nil {
			t.Fatal(err)
		}
		if info.LegacySchema != schema || root["version"] != nil || root["schema_version"] != nil || root["future"] == nil {
			t.Fatal(root, info)
		}
	}
	for _, raw := range []string{`{"records":{}}`, `{"version":1,"records":{}}`, `{"version":"1.4.2","records":{}}`} {
		root, _, err := Object([]byte(raw), "publication-citations.json")
		if err != nil || root["version"] != nil || root["records"] == nil {
			t.Fatal(root, err)
		}
	}
}

func TestBundleMarkerCompatibilityAndValidation(t *testing.T) {
	root := t.TempDir()
	if version, err := BundleVersion(root); err != nil || version != Current {
		t.Fatal(version, err)
	}
	if err := os.Mkdir(filepath.Join(root, "data"), 0755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "data", "version.json")
	for _, version := range []string{"1.0.0", "1.8.12"} {
		before := []byte(`{"version":"` + version + `","future":true}`)
		if err := os.WriteFile(path, before, 0644); err != nil {
			t.Fatal(err)
		}
		if actual, err := BundleVersion(root); err != nil || actual != version {
			t.Fatal(actual, err)
		}
		after, _ := os.ReadFile(path)
		if !bytes.Equal(before, after) {
			t.Fatal("validation changed marker")
		}
	}
	for _, invalid := range []string{`null`, `{}`, `{"version":1}`, `{"version":"2.0.0"}`, `{"version":"1.0.0"} {}`} {
		os.WriteFile(path, []byte(invalid), 0644)
		if ValidateBundle(root) == nil {
			t.Fatal("accepted invalid bundle", invalid)
		}
	}
}
