package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type CustomTheme struct {
	Name      string            `json:"name"`
	BaseTheme string            `json:"base_theme"`
	Colors    map[string]string `json:"colors,omitempty"`

	extraFields map[string]json.RawMessage
}

var customThemeColorKeys = map[string]bool{
	"bg": true, "surface": true, "surface-muted": true, "surface-strong": true,
	"ink": true, "ink-soft": true, "ink-faint": true, "line": true, "line-strong": true,
	"accent": true, "accent-strong": true, "accent-soft": true, "focus": true, "danger": true,
	"on-accent": true, "on-danger": true, "on-ink": true, "warning": true, "warning-soft": true,
	"cv-bg": true, "danger-strong": true, "danger-soft": true, "danger-line": true,
	"disabled": true, "info": true, "info-soft": true,
}

func (document *SettingsDocument) UnmarshalJSON(raw []byte) error {
	type settingsJSON SettingsDocument
	var decoded settingsJSON
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return err
	}
	fields, err := decodeObject(raw)
	if err != nil {
		return err
	}
	if value, exists := fields["custom_themes"]; exists && bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
		return fmt.Errorf("custom_themes는 JSON 객체여야 합니다")
	}
	if value, exists := fields["hidden_web_pages"]; exists && bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
		return fmt.Errorf("hidden_web_pages는 JSON 배열이어야 합니다")
	}
	*document = SettingsDocument(decoded)
	return nil
}

func (theme *CustomTheme) UnmarshalJSON(raw []byte) error {
	fields, err := decodeObject(raw)
	if err != nil {
		return fmt.Errorf("사용자 테마는 JSON 객체여야 합니다: %w", err)
	}
	for _, field := range []string{"name", "base_theme"} {
		if value, exists := fields[field]; exists {
			if _, err := decodeRequiredString(value, "사용자 테마 "+field); err != nil {
				return err
			}
		}
	}
	if value, exists := fields["colors"]; exists {
		colors, err := decodeObject(value)
		if err != nil {
			return fmt.Errorf("사용자 테마 colors는 JSON 객체여야 합니다: %w", err)
		}
		for key, value := range colors {
			if _, err := decodeRequiredString(value, "사용자 테마 colors."+key); err != nil {
				return err
			}
		}
	}
	type themeJSON CustomTheme
	var decoded themeJSON
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return err
	}
	*theme = CustomTheme(decoded)
	theme.extraFields = withoutKnownFields(fields, "name", "base_theme", "colors")
	return nil
}

func (theme CustomTheme) MarshalJSON() ([]byte, error) {
	known := map[string]any{"name": theme.Name, "base_theme": theme.BaseTheme}
	if len(theme.Colors) > 0 {
		known["colors"] = theme.Colors
	}
	fields, err := mergeObjectFields(theme.extraFields, known)
	if err != nil {
		return nil, err
	}
	return json.Marshal(fields)
}

func isCustomThemeSlot(value string) bool {
	return value == "custom-1" || value == "custom-2" || value == "custom-3"
}

func isPresetColorTheme(value string) bool {
	switch value {
	case "sage", "ocean", "violet", "rose", "sand", "slate", "mint", "teal", "sky", "indigo", "amber", "graphite":
		return true
	}
	return false
}

func normaliseCustomThemes(input, current map[string]CustomTheme) (map[string]CustomTheme, error) {
	// Older clients omit the optional map; an explicit {} resets all slots.
	if input == nil {
		input = current
	}
	if input == nil {
		return nil, nil
	}
	result := make(map[string]CustomTheme, len(input))
	for slot, theme := range input {
		if !isCustomThemeSlot(slot) {
			return nil, fmt.Errorf("custom_themes 슬롯이 올바르지 않습니다: %s", slot)
		}
		if !utf8.ValidString(theme.Name) || strings.IndexFunc(theme.Name, unicode.IsControl) >= 0 {
			return nil, fmt.Errorf("%s 이름에는 제어문자를 사용할 수 없습니다", slot)
		}
		theme.Name = strings.TrimSpace(theme.Name)
		if theme.Name == "" || utf8.RuneCountInString(theme.Name) > 80 {
			return nil, fmt.Errorf("%s 이름은 비어 있지 않은 80자 이하 문자열이어야 합니다", slot)
		}
		if !isPresetColorTheme(theme.BaseTheme) {
			return nil, fmt.Errorf("%s base_theme에는 기본 색상 테마를 선택해 주세요", slot)
		}
		if theme.Colors != nil {
			colors := make(map[string]string, len(theme.Colors))
			for key, value := range theme.Colors {
				if !customThemeColorKeys[key] {
					return nil, fmt.Errorf("%s colors에 지원하지 않는 색상 속성이 있습니다: %s", slot, key)
				}
				if !taxonomyColorPattern.MatchString(value) {
					return nil, fmt.Errorf("%s colors.%s는 #RRGGBB 형식이어야 합니다", slot, key)
				}
				colors[key] = value
			}
			theme.Colors = colors
		}
		extras := cloneExtraFields(current[slot].extraFields)
		if extras == nil {
			extras = make(map[string]json.RawMessage)
		}
		for key, value := range theme.extraFields {
			extras[key] = cloneRawMessage(value)
		}
		theme.extraFields = extras
		result[slot] = theme
	}
	return result, nil
}
