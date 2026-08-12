// Package release checks metadata and the boundary of a future update.
// It never copies or updates files.
package release

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/Gonie-Gonie/Gonie-Gonie.github.io/editor/internal/dataversion"
)

type Entry struct {
	Version string `json:"version"`
	Date    string `json:"date"`
	Notes   string `json:"notes"`
}
type Manifest struct {
	Entry
	DataVersion string `json:"data_version"`
}
type Index struct {
	Version  string  `json:"version"`
	Releases []Entry `json:"releases"`
}
type Policy struct {
	Version    string   `json:"version"`
	Default    string   `json:"default"`
	Replace    []string `json:"replace"`
	Preserve   []string `json:"preserve"`
	Regenerate []string `json:"regenerate"`
}

var dataFiles = []string{"academic_activities.json", "awards.json", "board.json", "people.json", "profile.json", "projects.json", "publications.json", "settings.json", "software.json"}
var regeneratedFiles = []string{"exports/publications.bib", "exports/publications.ris", "exports/publications.csl.json", "exports/build-report.json"}

func readJSON(root, name string, target any) error {
	raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(name)))
	if err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	if err := dataversion.Decode(raw, target); err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	return nil
}

func cleanPath(name string) (string, error) {
	name = strings.ReplaceAll(name, "\\", "/")
	if name == "" || strings.HasPrefix(name, "/") || strings.ContainsAny(name, ":*?\"<>|") || strings.IndexFunc(name, unicode.IsControl) >= 0 {
		return "", fmt.Errorf("invalid repository path %q", name)
	}
	for _, part := range strings.Split(strings.TrimSuffix(name, "/"), "/") {
		if part == "" || part == "." || part == ".." || strings.TrimSpace(part) != part || strings.HasSuffix(part, ".") {
			return "", fmt.Errorf("invalid repository path %q", name)
		}
		base := strings.ToUpper(strings.SplitN(part, ".", 2)[0])
		reserved := base == "CON" || base == "PRN" || base == "AUX" || base == "NUL" || base == "CONIN$" || base == "CONOUT$"
		if len(base) == 4 && (strings.HasPrefix(base, "COM") || strings.HasPrefix(base, "LPT")) && base[3] >= '1' && base[3] <= '9' {
			reserved = true
		}
		if reserved {
			return "", fmt.Errorf("invalid Windows repository path %q", name)
		}
	}
	return name, nil
}

func matches(name, rule string) bool {
	name, rule = strings.ToLower(name), strings.ToLower(rule)
	if strings.HasSuffix(rule, "/") {
		return name == strings.TrimSuffix(rule, "/") || strings.HasPrefix(name, rule)
	}
	return name == rule
}
func anyMatch(name string, rules []string) bool {
	for _, rule := range rules {
		if matches(name, rule) {
			return true
		}
	}
	return false
}

// Regenerated exports come from the fork's data, never from upstream.
func (p Policy) Classify(name string) (string, error) {
	name, err := cleanPath(name)
	if err != nil {
		return "", err
	}
	if anyMatch(name, p.Regenerate) {
		return "regenerate", nil
	}
	if anyMatch(name, p.Preserve) {
		return "preserve", nil
	}
	if anyMatch(name, p.Replace) {
		return "replace", nil
	}
	return "preserve", nil
}

func (p Policy) Validate() error {
	if err := dataversion.Validate(p.Version); err != nil {
		return fmt.Errorf("update policy version: %w", err)
	}
	if p.Default != "preserve" {
		return fmt.Errorf("unknown paths must default to preserve")
	}
	for _, rules := range [][]string{p.Replace, p.Preserve, p.Regenerate} {
		seen := map[string]bool{}
		for _, rule := range rules {
			cleaned, err := cleanPath(rule)
			if err != nil || cleaned != rule || seen[strings.ToLower(rule)] {
				return fmt.Errorf("invalid or duplicate update rule %q", rule)
			}
			seen[strings.ToLower(rule)] = true
		}
	}
	for _, required := range []string{"data/", "exports/", "CNAME", ".git/", "editor/runtime/"} {
		found := false
		for _, rule := range p.Preserve {
			if strings.EqualFold(rule, required) {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("update policy must preserve %q", required)
		}
	}
	if len(p.Regenerate) != len(regeneratedFiles) {
		return fmt.Errorf("only the four local build exports may be regenerated")
	}
	for _, name := range regeneratedFiles {
		if !anyMatch(name, p.Regenerate) {
			return fmt.Errorf("missing regenerated export %q", name)
		}
	}
	for _, rule := range p.Regenerate {
		if strings.HasSuffix(rule, "/") {
			return fmt.Errorf("regenerated exports must use explicit file paths")
		}
	}
	for _, name := range []string{"data/profile.json", "data/version.json", "data/media/photo.jpg", "exports/publication-citations.json", "CNAME", ".git/config", ".vscode/settings.json", "editor/runtime/go/bin/go.exe", "editor/cmd/editor/build/bin/Profile-Editor.exe", "editor/cmd/editor/frontend/wailsjs/go/main/App.js"} {
		if action, _ := p.Classify(name); action != "preserve" {
			return fmt.Errorf("personal or local path %q must be preserved", name)
		}
	}
	if action, _ := p.Classify("exports/publications.bib"); action != "regenerate" {
		return fmt.Errorf("exports/ must be regenerated from local data")
	}
	return nil
}

func LoadPolicy(root string) (Policy, error) {
	var policy Policy
	if err := readJSON(root, "app/update-policy.json", &policy); err != nil {
		return policy, err
	}
	return policy, policy.Validate()
}

func checkNote(root string, entry Entry) error {
	if _, err := dataversion.Parse(entry.Version); err != nil {
		return err
	}
	if date, err := time.Parse("2006-01-02", entry.Date); err != nil || date.Year() < 1 || date.Format("2006-01-02") != entry.Date {
		return fmt.Errorf("release %s has an invalid date", entry.Version)
	}
	if entry.Notes != "app/releases/"+entry.Version+".md" {
		return fmt.Errorf("release %s must use app/releases/%s.md", entry.Version, entry.Version)
	}
	raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(entry.Notes)))
	if err != nil {
		return err
	}
	note := strings.ReplaceAll(string(raw), "\r\n", "\n")
	hasDate, hasBullet := false, false
	for _, line := range strings.Split(note, "\n") {
		if strings.TrimSpace(line) == entry.Date {
			hasDate = true
		}
		if strings.HasPrefix(line, "- ") && strings.TrimSpace(line[2:]) != "" {
			hasBullet = true
		}
	}
	if !utf8.Valid(raw) || !strings.HasPrefix(note, "# "+entry.Version+"\n") || !hasDate || !hasBullet {
		return fmt.Errorf("%s requires its version, date and bullet list", entry.Notes)
	}
	return nil
}

func Check(root string) (Manifest, error) {
	return CheckWithOptions(root, CheckOptions{})
}

type CheckOptions struct{ PreserveData bool }

// PreserveData checks the distributed program without requiring a fork's
// untouched data files to match the publisher's current data format.
func CheckWithOptions(root string, options CheckOptions) (Manifest, error) {
	var manifest Manifest
	if err := readJSON(root, "app/release.json", &manifest); err != nil {
		return manifest, err
	}
	if err := checkNote(root, manifest.Entry); err != nil {
		return manifest, err
	}
	if _, err := LoadPolicy(root); err != nil {
		return manifest, err
	}
	if !options.PreserveData {
		if err := dataversion.Validate(manifest.DataVersion); err != nil {
			return manifest, fmt.Errorf("expected data bundle version: %w", err)
		}
		expected, err := dataversion.Parse(manifest.DataVersion)
		if err != nil {
			return manifest, fmt.Errorf("expected data bundle version: %w", err)
		}
		var bundle struct {
			Version string `json:"version"`
		}
		if err := readJSON(root, "data/version.json", &bundle); err != nil {
			return manifest, err
		}
		actual, err := dataversion.Parse(bundle.Version)
		if err != nil || actual.Major != expected.Major || actual.Minor < expected.Minor || (actual.Minor == expected.Minor && actual.Patch < expected.Patch) {
			return manifest, fmt.Errorf("data bundle version %q is incompatible with %q", bundle.Version, manifest.DataVersion)
		}
		for _, name := range dataFiles {
			var document any
			if err := readJSON(root, "data/"+name, &document); err != nil {
				return manifest, err
			}
			if dataversion.IsArray(name) {
				if _, ok := document.([]any); !ok {
					return manifest, fmt.Errorf("data/%s must be a plain array", name)
				}
			} else {
				object, ok := document.(map[string]any)
				if !ok {
					return manifest, fmt.Errorf("data/%s must be an object", name)
				}
				if _, exists := object["version"]; exists {
					return manifest, fmt.Errorf("data/%s must use the shared data/version.json", name)
				}
				if _, exists := object["schema_version"]; exists {
					return manifest, fmt.Errorf("data/%s still has a legacy schema_version", name)
				}
			}
		}
	} else if _, err := dataversion.Parse(manifest.DataVersion); err != nil {
		return manifest, fmt.Errorf("expected data bundle version: %w", err)
	}
	var index Index
	if err := readJSON(root, "app/releases/index.json", &index); err != nil {
		return manifest, err
	}
	if err := dataversion.Validate(index.Version); err != nil {
		return manifest, err
	}
	if len(index.Releases) == 0 || index.Releases[0] != manifest.Entry {
		return manifest, fmt.Errorf("latest release index entry must match app/release.json")
	}
	for i, entry := range index.Releases {
		if err := checkNote(root, entry); err != nil {
			return manifest, err
		}
		if i > 0 {
			if order, err := dataversion.Compare(index.Releases[i-1].Version, entry.Version); err != nil || order <= 0 {
				return manifest, fmt.Errorf("release index must list distinct versions newest first")
			}
		}
	}
	var wails struct {
		Info struct {
			ProductVersion string `json:"productVersion"`
		} `json:"info"`
	}
	if err := readJSON(root, "editor/cmd/editor/wails.json", &wails); err != nil {
		return manifest, err
	}
	if wails.Info.ProductVersion != manifest.Version {
		return manifest, fmt.Errorf("editor productVersion must match release version %s", manifest.Version)
	}
	return manifest, nil
}
