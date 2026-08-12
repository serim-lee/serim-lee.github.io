package sitebuild

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func patentFixture() record {
	return record{"title_ko": "예시 발명", "type": "domestic", "application_number": "KR-EXAMPLE-001", "inventor_ids": []any{"patent-inventor"}}
}

func TestPatentsValidation(t *testing.T) {
	for _, patentType := range patentTypes {
		item := patentFixture()
		item["type"] = patentType
		item["future_field"] = record{"keep": true}
		v := &validator{}
		v.patent(item, "patent")
		if len(v.errors) != 0 {
			t.Fatalf("valid %s patent rejected: %v", patentType, v.errors)
		}
	}
	tests := map[string]func(record){
		"title":                           func(p record) { delete(p, "title_ko") },
		"type":                            func(p record) { p["type"] = "foreign" },
		"missing type":                    func(p record) { delete(p, "type") },
		"number":                          func(p record) { delete(p, "application_number") },
		"null inventors":                  func(p record) { p["inventor_ids"] = nil },
		"scalar inventors":                func(p record) { p["inventor_ids"] = "patent-inventor" },
		"null inventor":                   func(p record) { p["inventor_ids"] = []any{nil} },
		"numeric inventor":                func(p record) { p["inventor_ids"] = []any{123} },
		"noncanonical inventor":           func(p record) { p["inventor_ids"] = []any{"Patent Inventor"} },
		"whitespace inventor":             func(p record) { p["inventor_ids"] = []any{" patent-inventor"} },
		"duplicate inventors":             func(p record) { p["inventor_ids"] = []any{"patent-inventor", "patent-inventor"} },
		"date":                            func(p record) { p["application_date"] = "2025-02-29" },
		"partial date":                    func(p record) { p["publication_date"] = "2025-03" },
		"zero year":                       func(p record) { p["registration_date"] = "0000-01-01" },
		"publication before application":  func(p record) { p["application_date"], p["publication_date"] = "2025-01-02", "2025-01-01" },
		"registration before application": func(p record) { p["application_date"], p["registration_date"] = "2025-01-02", "2025-01-01" },
		"unsafe URL":                      func(p record) { p["url"] = "javascript:alert(1)" },
		"relative URL":                    func(p record) { p["url"] = "/patents/001" },
		"control in URL":                  func(p record) { p["url"] = "https://example.com/\n001" },
	}
	for _, field := range patentStringFields {
		field := field
		tests["null "+field] = func(p record) { p[field] = nil }
		tests["numeric "+field] = func(p record) { p[field] = 123 }
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			item := patentFixture()
			mutate(item)
			v := &validator{}
			v.patent(item, "patent")
			if len(v.errors) == 0 {
				t.Fatal("invalid patent accepted")
			}
		})
	}
}

func TestPatentsURLUnicodeHostAndPortBoundaries(t *testing.T) {
	prefix := "https://example.com/"
	tests := []struct {
		name, value string
		valid       bool
	}{
		{"blank optional URL", "", true},
		{"4096 ASCII characters", prefix + strings.Repeat("a", 4096-len(prefix)), true},
		{"4097 ASCII characters", prefix + strings.Repeat("a", 4097-len(prefix)), false},
		{"4096 Korean characters", prefix + strings.Repeat("가", 4096-len(prefix)), true},
		{"4097 Korean characters", prefix + strings.Repeat("가", 4097-len(prefix)), false},
		{"4096 non BMP characters", prefix + strings.Repeat("🧪", 4096-len(prefix)), true},
		{"4097 non BMP characters", prefix + strings.Repeat("🧪", 4097-len(prefix)), false},
		{"uppercase scheme", "HTTPs://example.com/patent", true},
		{"international domain", "https://특허.한국/발명", true},
		{"IPv6 host", "https://[2001:db8::1]:65535/patent", true},
		{"port zero", "https://example.com:0/patent", true},
		{"port 65535", "https://example.com:65535/patent", true},
		{"port 65536", "https://example.com:65536/patent", false},
		{"overflowing port", "https://example.com:999999999999999999999/patent", false},
		{"non numeric port", "https://example.com:port/patent", false},
		{"missing hostname", "https://:80/patent", false},
		{"missing hostname with user", "https://user@:443/patent", false},
		{"empty IPv6 hostname", "https://[]:443/patent", false},
		{"C1 next line", prefix + "\u0085patent", false},
		{"C1 control", prefix + "\u009fpatent", false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			item := patentFixture()
			item["url"] = test.value
			v := &validator{}
			v.patent(item, "patent")
			if (len(v.errors) == 0) != test.valid {
				t.Fatalf("URL accepted = %t, want %t: %v", len(v.errors) == 0, test.valid, v.errors)
			}
		})
	}
}

func TestPatentsCollectionDefaultsAndRejectsNull(t *testing.T) {
	root, data := repositoryData(t)
	profile := obj(data["profile"])
	delete(profile, "patents")
	if errors := validateData(root, data); len(errors) != 0 {
		t.Fatal(strings.Join(errors, "\n"))
	}
	if values, ok := profile["patents"].([]any); !ok || len(values) != 0 {
		t.Fatal("missing patents did not default to an empty array")
	}
	for name, value := range map[string]any{"null": nil, "object": record{}, "scalar": "patent", "null row": []any{nil}} {
		t.Run(name, func(t *testing.T) {
			profile["patents"] = value
			if errors := validateData(root, data); len(errors) == 0 {
				t.Fatal("invalid patent collection accepted")
			}
		})
	}
	patent := patentFixture()
	patent["inventor_ids"] = []any{str(first(data, "people")["id"])}
	profile["patents"] = []any{patent}
	if errors := validateData(root, data); len(errors) != 0 {
		t.Fatal(strings.Join(errors, "\n"))
	}
	obj(array(profile["patents"])[0])["visible_in_CV"] = "false"
	if errors := validateData(root, data); len(errors) == 0 {
		t.Fatal("invalid patent CV visibility accepted")
	}
	delete(patent, "visible_in_CV")
	patent["inventor_ids"] = []any{"unknown-inventor"}
	if errors := validateData(root, data); !strings.Contains(strings.Join(errors, "\n"), "unknown inventor_ids") {
		t.Fatalf("unknown inventor reference accepted: %v", errors)
	}
}

func TestPatentsLegacySectionMigrationPreservesOrder(t *testing.T) {
	for _, version := range []int{5, 8} {
		_, data := repositoryData(t)
		settings := obj(data["settings"])
		settings["schema_version"] = version
		previous := map[string][]any{}
		for _, field := range []string{"main_page_sections", "hidden_main_page_sections", "cv_sections", "hidden_cv_sections"} {
			next := []any{}
			for _, value := range array(settings[field]) {
				if value == "patents" || (version == 5 && strings.Contains(field, "main_page") && value == "academic_activities") {
					continue
				}
				next = append(next, value)
			}
			settings[field], previous[field] = next, append([]any{}, next...)
		}
		normalizeMainSections(settings)
		for _, field := range []string{"main_page_sections", "cv_sections"} {
			if !reflect.DeepEqual(array(settings[field]), previous[field]) {
				t.Fatalf("schema %d migration reordered %s", version, field)
			}
		}
		for _, field := range []string{"hidden_main_page_sections", "hidden_cv_sections"} {
			want := previous[field]
			if version == 5 && field == "hidden_main_page_sections" {
				want = append(want, "academic_activities")
			}
			want = append(want, "patents")
			if !reflect.DeepEqual(array(settings[field]), want) {
				t.Fatalf("schema %d hidden migration = %v, want %v", version, settings[field], want)
			}
		}
		if !sectionSet(settings["main_page_sections"], settings["hidden_main_page_sections"], mainSections) || !sectionSet(settings["cv_sections"], settings["hidden_cv_sections"], cvSections) {
			t.Fatal("migrated section partition is incomplete")
		}
	}
}

func TestPatentsBuildReportIncludesProfilePatentCount(t *testing.T) {
	root := fixtureRoot(t)
	data, err := readData(root)
	if err != nil {
		t.Fatal(err)
	}
	profile := obj(data["profile"])
	patent := patentFixture()
	patent["inventor_ids"] = []any{str(first(data, "people")["id"])}
	profile["patents"] = []any{patent}
	raw, err := json.Marshal(profile)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "data", "profile.json"), raw, 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := Build(root, false); err != nil {
		t.Fatal(err)
	}
	raw, err = os.ReadFile(filepath.Join(root, "exports", "build-report.json"))
	if err != nil {
		t.Fatal(err)
	}
	var report record
	if err := json.Unmarshal(raw, &report); err != nil {
		t.Fatal(err)
	}
	if got := obj(report["counts"])["patents"]; got != float64(1) {
		t.Fatalf("patent report count = %v, want 1", got)
	}
}
