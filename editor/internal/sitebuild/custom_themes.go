package sitebuild

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

var presetColorThemes = []string{"sage", "ocean", "violet", "rose", "sand", "slate", "mint", "teal", "sky", "indigo", "amber", "graphite"}
var customColorSlots = []string{"custom-1", "custom-2", "custom-3"}
var customColorKeys = []string{
	"bg", "surface", "surface-muted", "surface-strong", "ink", "ink-soft", "ink-faint", "line", "line-strong",
	"accent", "accent-strong", "accent-soft", "focus", "danger", "on-accent", "on-danger", "on-ink",
	"warning", "warning-soft", "cv-bg", "danger-strong", "danger-soft", "danger-line", "disabled", "info", "info-soft",
}
var customHexColor = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

func (v *validator) themeSettings(settings record) {
	selected := "sage"
	if has(settings, "color_theme") {
		selected = str(settings["color_theme"])
	}
	if !contains(presetColorThemes, selected) && !contains(customColorSlots, selected) {
		v.fail("settings.json", "color_theme must be a supported website color theme.")
	}
	if !has(settings, "custom_themes") {
		return
	}
	themes := obj(settings["custom_themes"])
	if themes == nil {
		v.fail("settings.json", "custom_themes must be an object.")
		return
	}
	for slot, value := range themes {
		context := "settings.json custom_themes " + slot
		if !contains(customColorSlots, slot) {
			v.fail(context, "only custom-1, custom-2 and custom-3 are supported.")
			continue
		}
		theme := obj(value)
		if theme == nil {
			v.fail(context, "must be an object.")
			continue
		}
		name, ok := theme["name"].(string)
		if !ok || strings.TrimSpace(name) == "" || utf8.RuneCountInString(strings.TrimSpace(name)) > 80 || controls.MatchString(name) {
			v.fail(context, "name must be a nonempty string of at most 80 characters without control characters.")
		}
		if !isString(theme["base_theme"]) || !contains(presetColorThemes, str(theme["base_theme"])) {
			v.fail(context, "base_theme must be one of the twelve preset themes.")
		}
		if !has(theme, "colors") {
			continue
		}
		colors := obj(theme["colors"])
		if colors == nil {
			v.fail(context, "colors must be an object.")
			continue
		}
		for key, value := range colors {
			if !contains(customColorKeys, key) {
				v.fail(context, "unsupported colors key '%s'.", key)
			} else if !isString(value) || !customHexColor.MatchString(str(value)) {
				v.fail(context, "colors.%s must use the #RRGGBB format.", key)
			}
		}
	}
}
