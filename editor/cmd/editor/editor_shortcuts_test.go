package main

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

func TestEditorShortcutsDefaultsAndNormalisation(t *testing.T) {
	got, err := normaliseEditorShortcuts(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	for action, want := range defaultEditorShortcuts {
		if got[action] != want {
			t.Errorf("default shortcut for %s = %q, want %q", action, got[action], want)
		}
	}
	got, err = normaliseEditorShortcuts(map[string]string{
		"save": " command + shift + f2 ", "citation_update": "alt+comma",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got["save"] != "Ctrl+Shift+F2" || got["citation_update"] != "Alt+Comma" {
		t.Fatalf("normalised shortcuts = %#v", got)
	}
	if got["push"] != "Ctrl+P" {
		t.Fatal("a partial shortcut map did not retain other defaults")
	}
}

func TestEditorShortcutsAllowDisableAndRejectConflicts(t *testing.T) {
	got, err := normaliseEditorShortcuts(map[string]string{"save": "", "settings": "Ctrl+Comma"}, nil)
	if err != nil || got["save"] != "" {
		t.Fatalf("disabled shortcut = %#v, %v", got, err)
	}
	for _, input := range []map[string]string{
		{"save": "Ctrl+F", "push": "Ctrl+F"},
		{"save": "Ctrl+Z"},
		{"save": "S"},
		{"unknown": "Ctrl+Q"},
	} {
		if _, err := normaliseEditorShortcuts(input, nil); err == nil {
			t.Errorf("invalid shortcuts accepted: %#v", input)
		}
	}
}

func TestEditorShortcutsSettingsRoundTripAndOlderClientPreservation(t *testing.T) {
	root := newTempRepoFixture(t, fixtureSettingsJSON())
	app := &App{repoRoot: root}
	loaded, err := app.LoadSettings()
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Settings.EditorShortcuts["citation_update"] != "Ctrl+U" {
		t.Fatalf("legacy default shortcuts = %#v", loaded.Settings.EditorShortcuts)
	}
	input := loaded.Settings
	input.EditorShortcuts = map[string]string{"save": "Ctrl+Shift+S", "citation_update": "Ctrl+U", "push": "Ctrl+P", "revert": "Ctrl+Shift+R", "settings": "Ctrl+Comma"}
	saved, err := app.SaveSettings(SaveSettingsRequest{Settings: input, Revision: loaded.Revision})
	if err != nil {
		t.Fatal(err)
	}
	if saved.Settings.EditorShortcuts["save"] != "Ctrl+Shift+S" {
		t.Fatalf("saved shortcuts = %#v", saved.Settings.EditorShortcuts)
	}
	path := filepath.Join(root, "data", "settings.json")
	var raw map[string]any
	if err := json.Unmarshal(readFixtureFile(t, path), &raw); err != nil {
		t.Fatal(err)
	}
	if _, ok := raw["editor_shortcuts"].(map[string]any); !ok {
		t.Fatal("editor_shortcuts was not persisted")
	}
	loadedAgain, err := app.LoadSettings()
	if err != nil || loadedAgain.Settings.EditorShortcuts["save"] != "Ctrl+Shift+S" {
		t.Fatalf("reloaded shortcuts = %#v, %v", loadedAgain.Settings.EditorShortcuts, err)
	}
	legacyClient := loadedAgain.Settings
	legacyClient.EditorShortcuts = nil
	legacyClient.TopicFallback.LabelEN = "Uncategorized"
	saved, err = app.SaveSettings(SaveSettingsRequest{Settings: legacyClient, Revision: loadedAgain.Revision})
	if err != nil || saved.Settings.EditorShortcuts["save"] != "Ctrl+Shift+S" {
		t.Fatalf("older client did not preserve shortcuts: %#v, %v", saved.Settings.EditorShortcuts, err)
	}
}

func TestEditorShortcutsRejectInvalidSaveWithoutWriting(t *testing.T) {
	root := newTempRepoFixture(t, fixtureSettingsJSON())
	app := &App{repoRoot: root}
	loaded, err := app.LoadSettings()
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "data", "settings.json")
	before := readFixtureFile(t, path)
	input := loaded.Settings
	input.EditorShortcuts = map[string]string{"save": "Ctrl+Z"}
	if _, err := app.SaveSettings(SaveSettingsRequest{Settings: input, Revision: loaded.Revision}); err == nil || !strings.Contains(err.Error(), "editor_shortcuts") {
		t.Fatalf("invalid shortcut save error = %v", err)
	}
	if got := readFixtureFile(t, path); string(got) != string(before) {
		t.Fatal("invalid shortcuts changed settings.json")
	}
}
