package main

import (
	"fmt"
	"strconv"
	"strings"
)

var defaultEditorShortcuts = map[string]string{
	"save":            "Ctrl+S",
	"citation_update": "Ctrl+U",
	"push":            "Ctrl+P",
	"revert":          "Ctrl+Shift+R",
	"settings":        "Ctrl+Comma",
}

func normaliseEditorShortcuts(input, current map[string]string) (map[string]string, error) {
	if input == nil {
		input = current
	}
	result := make(map[string]string, len(defaultEditorShortcuts))
	for action, shortcut := range defaultEditorShortcuts {
		result[action] = shortcut
	}
	for action, shortcut := range input {
		if _, ok := defaultEditorShortcuts[action]; !ok {
			return nil, fmt.Errorf("editor_shortcuts에 알 수 없는 작업이 있습니다: %s", action)
		}
		if strings.TrimSpace(shortcut) == "" {
			result[action] = ""
			continue
		}
		normalised, err := normaliseEditorShortcut(shortcut)
		if err != nil {
			return nil, fmt.Errorf("editor_shortcuts.%s: %w", action, err)
		}
		result[action] = normalised
	}
	owners := make(map[string]string, len(result))
	for action, shortcut := range result {
		if shortcut == "" {
			continue
		}
		if previous, exists := owners[shortcut]; exists {
			return nil, fmt.Errorf("editor_shortcuts 단축키가 중복되었습니다: %s (%s, %s)", shortcut, previous, action)
		}
		owners[shortcut] = action
	}
	return result, nil
}

func normaliseEditorShortcut(value string) (string, error) {
	parts := strings.Split(strings.TrimSpace(value), "+")
	if len(parts) < 2 {
		return "", fmt.Errorf("Ctrl, Alt 또는 Shift와 함께 키를 지정해 주세요")
	}
	modifiers := map[string]bool{}
	key := ""
	for _, raw := range parts {
		part := strings.ToLower(strings.TrimSpace(raw))
		switch part {
		case "ctrl", "control", "meta", "cmd", "command":
			modifiers["Ctrl"] = true
		case "alt", "option":
			modifiers["Alt"] = true
		case "shift":
			modifiers["Shift"] = true
		case ",", "comma":
			if key != "" {
				return "", fmt.Errorf("한 개의 키만 지정할 수 있습니다")
			}
			key = "Comma"
		default:
			if key != "" {
				return "", fmt.Errorf("한 개의 키만 지정할 수 있습니다")
			}
			upper := strings.ToUpper(part)
			if len(upper) == 1 && ((upper[0] >= 'A' && upper[0] <= 'Z') || (upper[0] >= '0' && upper[0] <= '9')) {
				key = upper
				continue
			}
			if strings.HasPrefix(upper, "F") {
				if number, err := strconv.Atoi(upper[1:]); err == nil && number >= 1 && number <= 12 {
					key = "F" + strconv.Itoa(number)
					continue
				}
			}
			return "", fmt.Errorf("지원하지 않는 키입니다: %s", raw)
		}
	}
	if key == "" || len(modifiers) == 0 {
		return "", fmt.Errorf("Ctrl, Alt 또는 Shift와 함께 키를 지정해 주세요")
	}
	if isReservedEditorShortcut(modifiers, key) {
		return "", fmt.Errorf("기본 텍스트 편집 또는 창 제어 단축키는 사용할 수 없습니다")
	}
	result := make([]string, 0, 4)
	for _, modifier := range []string{"Ctrl", "Alt", "Shift"} {
		if modifiers[modifier] {
			result = append(result, modifier)
		}
	}
	result = append(result, key)
	return strings.Join(result, "+"), nil
}

func isReservedEditorShortcut(modifiers map[string]bool, key string) bool {
	if modifiers["Alt"] && (key == "F4" || key == "Tab") {
		return true
	}
	if modifiers["Ctrl"] && (key == "F4" || key == "L" || key == "N" || key == "O" || (key == "R" && !modifiers["Shift"]) || key == "T" || key == "W") {
		return true
	}
	if !modifiers["Ctrl"] || modifiers["Alt"] {
		return false
	}
	if modifiers["Shift"] {
		return key == "Z" || key == "N" || key == "T" || key == "W"
	}
	switch key {
	case "A", "C", "V", "X", "Y", "Z":
		return true
	default:
		return false
	}
}
