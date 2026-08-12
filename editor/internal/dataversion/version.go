// Package dataversion owns the on-disk versions of editable JSON documents.
package dataversion

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

const Current = "1.0.0"

var semverPattern = regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`)

type SemVer struct{ Major, Minor, Patch uint64 }

// Parse accepts the stable major.minor.patch form used by data and releases.
func Parse(value string) (SemVer, error) {
	parts := semverPattern.FindStringSubmatch(value)
	if parts == nil {
		return SemVer{}, fmt.Errorf("invalid version %q: use major.minor.patch", value)
	}
	var result SemVer
	for i, target := range []*uint64{&result.Major, &result.Minor, &result.Patch} {
		n, err := strconv.ParseUint(parts[i+1], 10, 64)
		if err != nil {
			return SemVer{}, fmt.Errorf("invalid version %q", value)
		}
		*target = n
	}
	return result, nil
}

// Validate rejects incompatible data majors while retaining later 1.x formats.
func Validate(value string) error {
	v, err := Parse(value)
	if err != nil {
		return err
	}
	if v.Major != 1 {
		return fmt.Errorf("unsupported data version %q", value)
	}
	return nil
}

func Compare(left, right string) (int, error) {
	a, err := Parse(left)
	if err != nil {
		return 0, err
	}
	b, err := Parse(right)
	if err != nil {
		return 0, err
	}
	for _, pair := range [][2]uint64{{a.Major, b.Major}, {a.Minor, b.Minor}, {a.Patch, b.Patch}} {
		if pair[0] < pair[1] {
			return -1, nil
		}
		if pair[0] > pair[1] {
			return 1, nil
		}
	}
	return 0, nil
}

type Info struct {
	Version      string
	LegacySchema int
}

func Decode(raw []byte, destination any) error {
	if !utf8.Valid(raw) {
		return fmt.Errorf("invalid UTF-8 JSON")
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	if err := d.Decode(destination); err != nil {
		return err
	}
	var extra any
	if err := d.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("multiple JSON values")
		}
		return err
	}
	return nil
}

// Object normalizes only format metadata; content and unknown fields survive.
func Object(raw []byte, name string) (map[string]json.RawMessage, Info, error) {
	var root map[string]json.RawMessage
	if err := Decode(raw, &root); err != nil || root == nil {
		if err == nil {
			err = fmt.Errorf("must be a JSON object")
		}
		return nil, Info{}, fmt.Errorf("%s: %w", name, err)
	}
	info := Info{Version: Current}
	if value, exists := root["version"]; exists {
		if _, mixed := root["schema_version"]; mixed {
			return nil, info, fmt.Errorf("%s: use version without schema_version", name)
		}
		var version string
		if err := json.Unmarshal(value, &version); err != nil {
			if name != "publication-citations.json" || string(value) != "1" {
				return nil, info, fmt.Errorf("%s version must be a string", name)
			}
		} else {
			if err := Validate(version); err != nil {
				return nil, info, fmt.Errorf("%s: %w", name, err)
			}
			info.Version = version
		}
	} else if value, exists := root["schema_version"]; exists {
		var schema int
		if err := json.Unmarshal(value, &schema); err != nil {
			return nil, info, fmt.Errorf("%s schema_version must be an integer", name)
		}
		valid := schema == 1 && (name == "profile.json" || name == "academic_activities.json")
		if name == "settings.json" {
			valid = schema >= 3 && schema <= 9
		}
		if !valid {
			return nil, info, fmt.Errorf("%s: unsupported schema_version %d", name, schema)
		}
		info.LegacySchema = schema
	}
	delete(root, "schema_version")
	delete(root, "version")
	return root, info, nil
}

// Array accepts legacy arrays and the versioned {version,items} envelope.
func Array(raw []byte, name string) ([]json.RawMessage, error) {
	trimmed := bytes.TrimSpace(raw)
	var rows []json.RawMessage
	if len(trimmed) > 0 && trimmed[0] == '[' {
		if err := Decode(raw, &rows); err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
	} else {
		root, _, err := Object(raw, name)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(root["items"], &rows); err != nil {
			return nil, fmt.Errorf("%s items must be an array", name)
		}
	}
	if rows == nil {
		return nil, fmt.Errorf("%s must contain a JSON array", name)
	}
	return rows, nil
}

// MarshalArray emits the canonical raw array. Unknown legacy envelope metadata
// blocks the write instead of silently deleting information.
func MarshalArray(previous []byte, rows []json.RawMessage, name string) ([]byte, error) {
	if len(bytes.TrimSpace(previous)) > 0 && bytes.TrimSpace(previous)[0] == '{' {
		root, _, err := Object(previous, name)
		if err != nil {
			return nil, err
		}
		for key := range root {
			if key != "items" {
				return nil, fmt.Errorf("%s: legacy envelope field %q must be retained before converting to an array", name, key)
			}
		}
	}
	if rows == nil {
		rows = []json.RawMessage{}
	}
	encoded, err := json.MarshalIndent(rows, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(encoded, '\n'), nil
}

// BundleVersion is the only current data-format version. The absence of the
// marker remains supported for repositories created before bundle versioning.
func BundleVersion(root string) (string, error) {
	path := filepath.Join(root, "data", "version.json")
	raw, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return Current, nil
	}
	if err != nil {
		return "", err
	}
	var marker map[string]json.RawMessage
	if err := Decode(raw, &marker); err != nil || marker == nil {
		if err == nil {
			err = fmt.Errorf("must be an object")
		}
		return "", fmt.Errorf("data/version.json: %w", err)
	}
	var version string
	if err := json.Unmarshal(marker["version"], &version); err != nil {
		return "", fmt.Errorf("data/version.json: version must be a string")
	}
	if err := Validate(version); err != nil {
		return "", fmt.Errorf("data/version.json: %w", err)
	}
	return version, nil
}

func ValidateBundle(root string) error { _, err := BundleVersion(root); return err }

func IsArray(name string) bool {
	switch strings.TrimSuffix(name, ".json") {
	case "people", "projects", "software", "publications", "awards", "board":
		return true
	}
	return false
}

// Unwrap keeps existing runtime collection APIs independent of file envelopes.
func Unwrap(raw []byte, name string) (any, error) {
	if IsArray(name) {
		rows, err := Array(raw, name)
		if err != nil {
			return nil, err
		}
		encoded, _ := json.Marshal(rows)
		var items []any
		err = Decode(encoded, &items)
		return items, err
	}
	var value any
	if err := Decode(raw, &value); err != nil {
		return nil, err
	}
	if _, _, err := Object(raw, name); err != nil {
		return nil, err
	}
	return value, nil
}
