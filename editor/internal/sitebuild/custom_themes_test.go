package sitebuild

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCustomThemesValidationAndOptionalSlots(t *testing.T) {
	for _, selected := range append(append([]string{}, presetColorThemes...), customColorSlots...) {
		v := &validator{}
		v.themeSettings(record{"color_theme": selected})
		if len(v.errors) != 0 {
			t.Fatalf("theme selection without custom definitions rejected: %s, %v", selected, v.errors)
		}
	}
	for _, base := range presetColorThemes {
		colors := record{}
		for _, key := range customColorKeys {
			colors[key] = "#AbC123"
		}
		settings := record{"color_theme": "custom-2", "custom_themes": record{"custom-1": record{"name": " 사용자 테마 ", "base_theme": base, "colors": colors, "future_metadata": record{"keep": true}}, "custom-2": record{"name": strings.Repeat("가", 80), "base_theme": base}}}
		before, _ := json.Marshal(settings)
		v := &validator{}
		v.themeSettings(settings)
		after, _ := json.Marshal(settings)
		if len(v.errors) != 0 || string(before) != string(after) {
			t.Fatalf("valid custom themes rejected or mutated: %v", v.errors)
		}
	}
}

func TestCustomThemesRejectInvalidDefinitions(t *testing.T) {
	for _, invalid := range []any{nil, []any{}, "theme"} {
		v := &validator{}
		v.themeSettings(record{"custom_themes": invalid})
		if len(v.errors) == 0 {
			t.Fatalf("nonobject custom themes accepted: %#v", invalid)
		}
	}
	for name, mutate := range map[string]func(record){
		"unknown slot":     func(themes record) { themes["custom-4"] = themes["custom-1"] },
		"null theme":       func(themes record) { themes["custom-1"] = nil },
		"empty name":       func(themes record) { obj(themes["custom-1"])["name"] = " " },
		"missing name":     func(themes record) { delete(obj(themes["custom-1"]), "name") },
		"numeric name":     func(themes record) { obj(themes["custom-1"])["name"] = json.Number("1") },
		"long name":        func(themes record) { obj(themes["custom-1"])["name"] = strings.Repeat("가", 81) },
		"control name":     func(themes record) { obj(themes["custom-1"])["name"] = "\nTheme" },
		"missing base":     func(themes record) { delete(obj(themes["custom-1"]), "base_theme") },
		"custom base":      func(themes record) { obj(themes["custom-1"])["base_theme"] = "custom-2" },
		"null colors":      func(themes record) { obj(themes["custom-1"])["colors"] = nil },
		"nonobject colors": func(themes record) { obj(themes["custom-1"])["colors"] = []any{} },
		"unknown token":    func(themes record) { obj(obj(themes["custom-1"])["colors"])["other"] = "#123456" },
		"numeric color":    func(themes record) { obj(obj(themes["custom-1"])["colors"])["bg"] = json.Number("123456") },
		"null color":       func(themes record) { obj(obj(themes["custom-1"])["colors"])["bg"] = nil },
		"short hex":        func(themes record) { obj(obj(themes["custom-1"])["colors"])["bg"] = "#fff" },
		"alpha hex":        func(themes record) { obj(obj(themes["custom-1"])["colors"])["bg"] = "#12345678" },
		"CSS value":        func(themes record) { obj(obj(themes["custom-1"])["colors"])["bg"] = "red" },
		"spaced hex":       func(themes record) { obj(obj(themes["custom-1"])["colors"])["bg"] = " #123456 " },
	} {
		t.Run(name, func(t *testing.T) {
			themes := record{"custom-1": record{"name": "Theme", "base_theme": "sage", "colors": record{"bg": "#123456"}}}
			mutate(themes)
			v := &validator{}
			v.themeSettings(record{"custom_themes": themes})
			if len(v.errors) == 0 {
				t.Fatal("invalid custom themes accepted")
			}
		})
	}
}

func TestRepositoryAcceptsCustomSelectionAndValidatesDefinitions(t *testing.T) {
	root, data := repositoryData(t)
	settings := obj(data["settings"])
	settings["color_theme"] = "custom-3"
	settings["custom_themes"] = record{"custom-3": record{"name": "사용자 테마 3", "base_theme": "sage", "colors": record{"accent": "#123456"}}}
	if failures := validateData(root, data); len(failures) != 0 {
		t.Fatal(strings.Join(failures, "\n"))
	}
	obj(obj(settings["custom_themes"])["custom-3"])["base_theme"] = "custom-1"
	if failures := validateData(root, data); len(failures) == 0 {
		t.Fatal("repository validation skipped invalid custom theme metadata")
	}
}
