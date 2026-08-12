package main

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestCustomThemesRoundTripPreservesMetadataAndOlderClients(t *testing.T) {
	root := newTempRepoFixture(t, fixtureSettingsJSON())
	path := filepath.Join(root, "data", "settings.json")
	var settings map[string]any
	if err := json.Unmarshal(readFixtureFile(t, path), &settings); err != nil {
		t.Fatal(err)
	}
	colors := make(map[string]string)
	for key := range customThemeColorKeys {
		colors[key] = "#12aBcD"
	}
	settings["color_theme"] = "custom-1"
	settings["custom_themes"] = map[string]any{
		"custom-1": map[string]any{"name": "  사용자 테마 1  ", "base_theme": "violet", "colors": colors, "future_theme": map[string]any{"keep": true}},
		"custom-2": map[string]any{"name": "사용자 테마 2", "base_theme": "ocean"},
	}
	settings["future_settings"] = map[string]any{"keep": true}
	writeJSONFixture(t, path, settings)
	before := readFixtureFile(t, path)
	app := &App{repoRoot: root}
	loaded, err := app.LoadEditorData()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, readFixtureFile(t, path)) {
		t.Fatal("loading custom themes wrote settings.json")
	}
	if loaded.Settings.ColorTheme != "custom-1" || loaded.Settings.CustomThemes["custom-1"].Name != "사용자 테마 1" || !reflect.DeepEqual(loaded.Settings.CustomThemes["custom-1"].Colors, colors) {
		t.Fatal("custom theme selection or values were lost")
	}
	if _, exists := loaded.Settings.extraFields["custom_themes"]; exists {
		t.Fatal("custom themes were retained as stale unknown root metadata")
	}
	raw, err := json.Marshal(loaded.Settings)
	if err != nil {
		t.Fatal(err)
	}
	var input SettingsDocument
	if err := json.Unmarshal(raw, &input); err != nil {
		t.Fatal(err)
	}
	// A client sending only known fields must retain existing future metadata.
	input.CustomThemes["custom-1"] = CustomTheme{Name: "  수정 테마  ", BaseTheme: "mint", Colors: map[string]string{"accent": "#ABCDEF"}}
	request := saveRequestRevisions(loaded)
	request.SaveSettings = true
	request.Settings = input
	saved, err := app.SaveEditorData(request)
	if err != nil {
		t.Fatal(err)
	}
	if saved.Settings.SchemaVersion != settingsSchemaVersion || saved.Settings.CustomThemes["custom-1"].Name != "수정 테마" {
		t.Fatal("custom theme normalization changed schema or lost edited name")
	}
	var stored map[string]any
	if err := json.Unmarshal(readFixtureFile(t, path), &stored); err != nil {
		t.Fatal(err)
	}
	custom := stored["custom_themes"].(map[string]any)["custom-1"].(map[string]any)
	if custom["future_theme"].(map[string]any)["keep"] != true || stored["future_settings"].(map[string]any)["keep"] != true {
		t.Fatal("saving custom themes lost unknown nested or root settings")
	}
	reopened, err := app.LoadSettings()
	if err != nil {
		t.Fatal(err)
	}
	if reopened.Settings.ColorTheme != "custom-1" || reopened.Settings.CustomThemes["custom-1"].Colors["accent"] != "#ABCDEF" || reopened.Settings.CustomThemes["custom-2"].BaseTheme != "ocean" {
		t.Fatal("custom themes changed after reopening")
	}
	input = reopened.Settings
	resetTheme := input.CustomThemes["custom-1"]
	resetTheme.Colors = map[string]string{}
	input.CustomThemes["custom-1"] = resetTheme
	if _, err := app.SaveSettings(SaveSettingsRequest{Settings: input, Revision: reopened.Revision}); err != nil {
		t.Fatal(err)
	}
	reopened, err = app.LoadSettings()
	if err != nil {
		t.Fatal(err)
	}
	if len(reopened.Settings.CustomThemes["custom-1"].Colors) != 0 || reopened.Settings.CustomThemes["custom-1"].BaseTheme != "mint" || len(reopened.Settings.CustomThemes["custom-1"].extraFields) == 0 {
		t.Fatal("clearing overrides revived saved colors or lost the base theme/future metadata")
	}
	input = reopened.Settings
	input.CustomThemes = nil
	input.ColorTheme = ""
	input.TopicFallback.LabelEN = "Updated fallback"
	older, err := app.SaveSettings(SaveSettingsRequest{Settings: input, Revision: reopened.Revision})
	if err != nil {
		t.Fatal(err)
	}
	if older.Settings.ColorTheme != "custom-1" || !reflect.DeepEqual(older.Settings.CustomThemes, reopened.Settings.CustomThemes) {
		t.Fatal("older client omission reset selected or saved custom themes")
	}
	input = older.Settings
	input.CustomThemes = map[string]CustomTheme{}
	cleared, err := app.SaveSettings(SaveSettingsRequest{Settings: input, Revision: older.Revision})
	if err != nil {
		t.Fatal(err)
	}
	if cleared.Settings.CustomThemes == nil || len(cleared.Settings.CustomThemes) != 0 || cleared.Settings.ColorTheme != "custom-1" {
		t.Fatal("explicit empty map did not reset slots while allowing a missing selected slot")
	}
	if err := json.Unmarshal(readFixtureFile(t, path), &stored); err != nil {
		t.Fatal(err)
	}
	if themes, ok := stored["custom_themes"].(map[string]any); !ok || len(themes) != 0 {
		t.Fatal("cleared custom themes were not stored as an empty object")
	}
}

func TestCustomThemesValidationRejectsInvalidSaveWithoutWriting(t *testing.T) {
	root := newTempRepoFixture(t, fixtureSettingsJSON())
	path := filepath.Join(root, "data", "settings.json")
	app := &App{repoRoot: root}
	loaded, err := app.LoadSettings()
	if err != nil {
		t.Fatal(err)
	}
	before := readFixtureFile(t, path)
	for name, mutate := range map[string]func(map[string]CustomTheme){
		"unknown slot": func(themes map[string]CustomTheme) { themes["custom-4"] = themes["custom-1"] },
		"empty name": func(themes map[string]CustomTheme) {
			item := themes["custom-1"]
			item.Name = " "
			themes["custom-1"] = item
		},
		"long name": func(themes map[string]CustomTheme) {
			item := themes["custom-1"]
			item.Name = strings.Repeat("가", 81)
			themes["custom-1"] = item
		},
		"control name": func(themes map[string]CustomTheme) {
			item := themes["custom-1"]
			item.Name = "테마\x00이름"
			themes["custom-1"] = item
		},
		"leading control name": func(themes map[string]CustomTheme) {
			item := themes["custom-1"]
			item.Name = "\n테마"
			themes["custom-1"] = item
		},
		"custom base": func(themes map[string]CustomTheme) {
			item := themes["custom-1"]
			item.BaseTheme = "custom-2"
			themes["custom-1"] = item
		},
		"unknown base": func(themes map[string]CustomTheme) {
			item := themes["custom-1"]
			item.BaseTheme = "unknown"
			themes["custom-1"] = item
		},
		"unknown token": func(themes map[string]CustomTheme) { themes["custom-1"].Colors["future-token"] = "#123456" },
		"CSS value":     func(themes map[string]CustomTheme) { themes["custom-1"].Colors["bg"] = "var(--ink)" },
		"short hex":     func(themes map[string]CustomTheme) { themes["custom-1"].Colors["bg"] = "#fff" },
		"alpha hex":     func(themes map[string]CustomTheme) { themes["custom-1"].Colors["bg"] = "#12345678" },
		"spaced hex":    func(themes map[string]CustomTheme) { themes["custom-1"].Colors["bg"] = " #123456 " },
	} {
		t.Run(name, func(t *testing.T) {
			input := loaded.Settings
			input.ColorTheme = "custom-1"
			input.CustomThemes = map[string]CustomTheme{"custom-1": {Name: "테마", BaseTheme: "sage", Colors: map[string]string{"bg": "#123456"}}}
			mutate(input.CustomThemes)
			if _, err := app.SaveSettings(SaveSettingsRequest{Settings: input, Revision: loaded.Revision}); err == nil {
				t.Fatal("invalid custom theme save succeeded")
			}
			if !bytes.Equal(before, readFixtureFile(t, path)) {
				t.Fatal("invalid custom theme save changed settings.json")
			}
		})
	}
}

func TestCustomThemesRejectsInvalidJSONTypes(t *testing.T) {
	for _, raw := range []string{
		`{"custom_themes":null}`, `{"custom_themes":[]}`, `{"custom_themes":{"custom-1":null}}`,
		`{"custom_themes":{"custom-1":{"name":null,"base_theme":"sage"}}}`,
		`{"custom_themes":{"custom-1":{"name":1,"base_theme":"sage"}}}`,
		`{"custom_themes":{"custom-1":{"name":"Theme","base_theme":null}}}`,
		`{"custom_themes":{"custom-1":{"name":"Theme","base_theme":"sage","colors":null}}}`,
		`{"custom_themes":{"custom-1":{"name":"Theme","base_theme":"sage","colors":[]}}}`,
		`{"custom_themes":{"custom-1":{"name":"Theme","base_theme":"sage","colors":{"bg":null}}}}`,
		`{"custom_themes":{"custom-1":{"name":"Theme","base_theme":"sage","colors":{"bg":12}}}}`,
	} {
		var settings SettingsDocument
		if err := json.Unmarshal([]byte(raw), &settings); err == nil {
			t.Fatalf("invalid custom theme JSON accepted: %s", raw)
		}
	}
}

func TestCustomThemeOptionalColorsAndNameLimit(t *testing.T) {
	for _, base := range []string{"sage", "ocean", "violet", "rose", "sand", "slate", "mint", "teal", "sky", "indigo", "amber", "graphite"} {
		input := map[string]CustomTheme{"custom-3": {Name: strings.Repeat("가", 80), BaseTheme: base}}
		got, err := normaliseCustomThemes(input, nil)
		if err != nil || got["custom-3"].BaseTheme != base || got["custom-3"].Colors != nil {
			t.Fatalf("valid base/name or omitted colors rejected: %#v, %v", got, err)
		}
	}
}

func TestCustomThemesLegacySaveKeepsMetadataOptional(t *testing.T) {
	root := newTempRepoFixture(t, fixtureSettingsJSON())
	path := filepath.Join(root, "data", "settings.json")
	app := &App{repoRoot: root}
	loaded, err := app.LoadSettings()
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Settings.CustomThemes != nil {
		t.Fatal("loading legacy settings invented saved custom themes")
	}
	input := loaded.Settings
	input.ColorTheme = "custom-3"
	if _, err := app.SaveSettings(SaveSettingsRequest{Settings: input, Revision: loaded.Revision}); err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(readFixtureFile(t, path), []byte(`"custom_themes"`)) {
		t.Fatal("an undefined custom selection inserted optional theme metadata")
	}
	if _, err := app.LoadSettings(); err != nil {
		t.Fatal(err)
	}
}
