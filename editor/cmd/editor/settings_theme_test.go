package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestSiteColorThemeLegacyDefaultDoesNotWrite(t *testing.T) {
	settings := fixtureSettingsJSON()
	root := newTempRepoFixture(t, settings)
	loaded, err := (&App{repoRoot: root}).LoadEditorData()
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Settings.ColorTheme != "sage" {
		t.Fatalf("legacy color theme = %q, want sage", loaded.Settings.ColorTheme)
	}
	if raw := readFixtureFile(t, filepath.Join(root, "data", "settings.json")); !bytes.Equal(raw, settings) {
		t.Fatal("loading the legacy color theme changed settings.json")
	}
}

func TestSiteColorThemeRoundTripPreservesUnknownSettings(t *testing.T) {
	root := newTempRepoFixture(t, fixtureSettingsJSON())
	path := filepath.Join(root, "data", "settings.json")
	var original map[string]any
	if err := json.Unmarshal(readFixtureFile(t, path), &original); err != nil {
		t.Fatal(err)
	}
	original["future_settings"] = map[string]any{"keep": []any{"value", float64(7)}}
	writeJSONFixture(t, path, original)
	app := &App{repoRoot: root}
	for _, theme := range []string{
		"sage", "ocean", "violet", "rose", "sand", "slate",
		"mint", "teal", "sky", "indigo", "amber", "graphite",
		"custom-1", "custom-2", "custom-3",
	} {
		t.Run(theme, func(t *testing.T) {
			loaded, err := app.LoadEditorData()
			if err != nil {
				t.Fatal(err)
			}
			request := saveRequestRevisions(loaded)
			request.Settings = loaded.Settings
			request.Settings.ColorTheme = theme
			request.SaveSettings = true
			saved, err := app.SaveEditorData(request)
			if err != nil {
				t.Fatal(err)
			}
			if saved.Settings.ColorTheme != theme {
				t.Fatalf("saved color theme = %q, want %q", saved.Settings.ColorTheme, theme)
			}
			var stored map[string]any
			if err := json.Unmarshal(readFixtureFile(t, path), &stored); err != nil {
				t.Fatal(err)
			}
			if stored["color_theme"] != theme {
				t.Fatalf("stored color theme = %#v, want %q", stored["color_theme"], theme)
			}
			if !reflect.DeepEqual(stored["future_settings"], original["future_settings"]) {
				t.Fatal("saving the color theme changed unknown settings")
			}
			reopened, err := app.LoadEditorData()
			if err != nil {
				t.Fatal(err)
			}
			if reopened.Settings.ColorTheme != theme {
				t.Fatalf("reopened color theme = %q, want %q", reopened.Settings.ColorTheme, theme)
			}
		})
	}
}

func TestSiteColorThemeRejectsInvalidSaveWithoutWriting(t *testing.T) {
	root := newTempRepoFixture(t, fixtureSettingsJSON())
	app := &App{repoRoot: root}
	loaded, err := app.LoadEditorData()
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "data", "settings.json")
	before := readFixtureFile(t, path)
	for _, invalid := range []string{"unknown", "SAGE", " ocean "} {
		request := saveRequestRevisions(loaded)
		request.Settings = loaded.Settings
		request.Settings.ColorTheme = invalid
		request.SaveSettings = true
		if _, err := app.SaveEditorData(request); err == nil || !strings.Contains(err.Error(), "color_theme") {
			t.Fatalf("saving invalid theme %q error = %v", invalid, err)
		}
		if raw := readFixtureFile(t, path); !bytes.Equal(raw, before) {
			t.Fatalf("rejected color theme %q changed settings.json", invalid)
		}
	}
}

func TestSiteColorThemeOlderSaveRequestPreservesSelection(t *testing.T) {
	root := newTempRepoFixture(t, fixtureSettingsJSON())
	path := filepath.Join(root, "data", "settings.json")
	var settings map[string]any
	if err := json.Unmarshal(readFixtureFile(t, path), &settings); err != nil {
		t.Fatal(err)
	}
	settings["color_theme"] = "violet"
	writeJSONFixture(t, path, settings)
	app := &App{repoRoot: root}
	loaded, err := app.LoadSettings()
	if err != nil {
		t.Fatal(err)
	}
	input := loaded.Settings
	input.ColorTheme = ""
	input.TopicFallback.LabelEN = "Uncategorized"
	saved, err := app.SaveSettings(SaveSettingsRequest{Settings: input, Revision: loaded.Revision})
	if err != nil {
		t.Fatal(err)
	}
	if saved.Settings.ColorTheme != "violet" {
		t.Fatalf("older client reset color theme to %q", saved.Settings.ColorTheme)
	}
}

func TestSiteColorThemeRejectsInvalidStoredSelection(t *testing.T) {
	root := newTempRepoFixture(t, fixtureSettingsJSON())
	path := filepath.Join(root, "data", "settings.json")
	var settings map[string]any
	if err := json.Unmarshal(readFixtureFile(t, path), &settings); err != nil {
		t.Fatal(err)
	}
	settings["color_theme"] = "unknown"
	writeJSONFixture(t, path, settings)
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := (&App{repoRoot: root}).LoadSettings(); err == nil || !strings.Contains(err.Error(), "color_theme") {
		t.Fatalf("loading invalid stored theme error = %v", err)
	}
	if raw := readFixtureFile(t, path); !bytes.Equal(raw, before) {
		t.Fatal("loading invalid color theme changed settings.json")
	}
}
