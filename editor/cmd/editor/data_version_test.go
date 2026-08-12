package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDataBundleVersionSettingsAndProfileSavesKeepSingleMarker(t *testing.T) {
	root := newTempRepoFixture(t, fixtureSettingsJSON())
	marker := []byte("{\"version\":\"1.4.2\",\"future\":true}\n")
	versionPath := filepath.Join(root, "data", "version.json")
	if err := os.WriteFile(versionPath, marker, 0644); err != nil {
		t.Fatal(err)
	}
	app := &App{repoRoot: root}
	loaded, err := app.LoadSettings()
	if err != nil {
		t.Fatal(err)
	}
	if _, err = app.SaveSettings(SaveSettingsRequest{Settings: loaded.Settings, Revision: loaded.Revision}); err != nil {
		t.Fatal(err)
	}
	settingsRaw, _ := os.ReadFile(filepath.Join(root, "data", "settings.json"))
	var settings map[string]any
	if err := json.Unmarshal(settingsRaw, &settings); err != nil {
		t.Fatal(err)
	}
	if settings["version"] != nil || settings["schema_version"] != nil {
		t.Fatal("independent settings version was written")
	}
	profile, err := readProfile(filepath.Join(root, "data", "profile.json"))
	if err != nil {
		t.Fatal(err)
	}
	profile.Data["version"] = "1.4.2" // An earlier per-file envelope remains readable.
	profile.Data["future_root"] = map[string]any{"keep": true}
	input, _ := json.Marshal(profile.Data)
	stored, next, err := buildProfileSave(input)
	if err != nil {
		t.Fatal(err)
	}
	if next.Data["future_root"] == nil || next.Data["version"] != nil || bytes.Contains(stored, []byte("schema_version")) {
		t.Fatal(string(stored))
	}
	after, _ := os.ReadFile(versionPath)
	if !bytes.Equal(after, marker) {
		t.Fatal("content save changed bundle format version")
	}
}

func TestIncompatibleDataBundleRejectsLoadAndSaveWithoutWriting(t *testing.T) {
	root := newTempRepoFixture(t, fixtureSettingsJSON())
	app := &App{repoRoot: root}
	loaded, err := app.LoadSettings()
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "data", "settings.json")
	before, _ := os.ReadFile(path)
	if err := os.WriteFile(filepath.Join(root, "data", "version.json"), []byte(`{"version":"2.0.0"}`), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := app.LoadEditorData(); err == nil {
		t.Fatal("incompatible bundle loaded")
	}
	if _, err := app.SaveSettings(SaveSettingsRequest{Settings: loaded.Settings, Revision: loaded.Revision}); err == nil {
		t.Fatal("incompatible bundle saved")
	}
	after, _ := os.ReadFile(path)
	if !bytes.Equal(before, after) {
		t.Fatal("rejected save modified data")
	}
}
