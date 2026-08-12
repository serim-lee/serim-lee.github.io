package main

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func patentProfileFixture(t *testing.T, patents any) []byte {
	t.Helper()
	root := newTempRepoFixture(t, fixtureSettingsJSON())
	var profile map[string]any
	if err := json.Unmarshal(readFixtureFile(t, filepath.Join(root, "data", "profile.json")), &profile); err != nil {
		t.Fatal(err)
	}
	profile["patents"] = patents
	raw, err := json.Marshal(profile)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func patentFixture() map[string]any {
	return map[string]any{"title_en": "Example invention", "type": "domestic", "application_number": "KR-EXAMPLE-001", "inventor_ids": []any{"fixture-self"}}
}

func TestPatentsDefaultMissingArrayAndRejectMalformedEntries(t *testing.T) {
	root := newTempRepoFixture(t, fixtureSettingsJSON())
	raw := readFixtureFile(t, filepath.Join(root, "data", "profile.json"))
	snapshot, err := readProfileBytes(raw)
	if err != nil {
		t.Fatal(err)
	}
	if values, ok := snapshot.Data["patents"].([]any); !ok || len(values) != 0 || !bytes.Equal(snapshot.Raw, raw) {
		t.Fatal("missing patents did not default to an empty array without changing source bytes")
	}
	for name, value := range map[string]any{"null": nil, "object": map[string]any{}, "scalar": "patent", "null row": []any{nil}, "scalar row": []any{"patent"}} {
		t.Run(name, func(t *testing.T) {
			if _, err := readProfileBytes(patentProfileFixture(t, value)); err == nil {
				t.Fatal("malformed patents were accepted")
			}
		})
	}
}

func TestPatentsValidateTypesNumbersInventorsDatesAndLinks(t *testing.T) {
	for _, patentType := range patentTypes {
		item := patentFixture()
		item["type"] = patentType
		if _, err := readProfileBytes(patentProfileFixture(t, []any{item})); err != nil {
			t.Fatalf("valid %s patent rejected: %v", patentType, err)
		}
	}
	valid := patentFixture()
	delete(valid, "application_number")
	valid["publication_number"] = "WO-EXAMPLE-A1"
	valid["application_date"], valid["publication_date"], valid["registration_date"] = "2024-02-29", "2025-01-01", "2026-01-01"
	valid["url"] = "https://example.com/patent?number=001#details"
	if _, err := readProfileBytes(patentProfileFixture(t, []any{valid})); err != nil {
		t.Fatal(err)
	}
	tests := map[string]func(map[string]any){
		"no title":                   func(p map[string]any) { delete(p, "title_en") },
		"blank title":                func(p map[string]any) { p["title_en"] = "  " },
		"no type":                    func(p map[string]any) { delete(p, "type") },
		"unknown type":               func(p map[string]any) { p["type"] = "foreign" },
		"no number":                  func(p map[string]any) { delete(p, "application_number") },
		"blank number":               func(p map[string]any) { p["application_number"] = " \t" },
		"null inventors":             func(p map[string]any) { p["inventor_ids"] = nil },
		"scalar inventors":           func(p map[string]any) { p["inventor_ids"] = "fixture-self" },
		"null inventor":              func(p map[string]any) { p["inventor_ids"] = []any{nil} },
		"numeric inventor":           func(p map[string]any) { p["inventor_ids"] = []any{123} },
		"noncanonical inventor":      func(p map[string]any) { p["inventor_ids"] = []any{"Fixture Self"} },
		"whitespace inventor":        func(p map[string]any) { p["inventor_ids"] = []any{" fixture-self"} },
		"duplicate inventors":        func(p map[string]any) { p["inventor_ids"] = []any{"fixture-self", "fixture-self"} },
		"invalid date":               func(p map[string]any) { p["application_date"] = "2025-02-29" },
		"partial date":               func(p map[string]any) { p["application_date"] = "2025" },
		"zero year":                  func(p map[string]any) { p["application_date"] = "0000-01-01" },
		"publication before filing":  func(p map[string]any) { p["application_date"], p["publication_date"] = "2025-01-02", "2025-01-01" },
		"registration before filing": func(p map[string]any) { p["application_date"], p["registration_date"] = "2025-01-02", "2025-01-01" },
		"unsafe link":                func(p map[string]any) { p["url"] = "javascript:alert(1)" },
		"relative link":              func(p map[string]any) { p["url"] = "/patent/1" },
		"control in link":            func(p map[string]any) { p["url"] = "https://example.com/\npatent" },
		"visibility type":            func(p map[string]any) { p["visible_in_CV"] = "false" },
	}
	for _, field := range patentStringFields {
		field := field
		tests["null "+field] = func(p map[string]any) { p[field] = nil }
		tests["numeric "+field] = func(p map[string]any) { p[field] = 123 }
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			item := patentFixture()
			mutate(item)
			if _, err := readProfileBytes(patentProfileFixture(t, []any{item})); err == nil {
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
			err := validatePatent(item, "patent")
			if (err == nil) != test.valid {
				t.Fatalf("URL accepted = %t, want %t: %v", err == nil, test.valid, err)
			}
		})
	}
}

func TestPatentsSaveRoundTripVisibilityCountsAndUnknownFields(t *testing.T) {
	item := patentFixture()
	item["future_field"], item["_clientKey"] = map[string]any{"keep": true}, "editor-only"
	raw, saved, err := buildProfileSave(patentProfileFixture(t, []any{item}))
	if err != nil {
		t.Fatal(err)
	}
	stored := saved.Data["patents"].([]any)[0].(map[string]any)
	if !reflect.DeepEqual(stored["future_field"], item["future_field"]) || stored["visible_in_CV"] != true || bytes.Contains(raw, []byte("_clientKey")) {
		t.Fatal("patent save lost unknown fields, visibility default or editor metadata cleanup")
	}
	if counts := profileSectionCounts(saved, 0); counts["patents"] != 1 {
		t.Fatalf("main patent count = %d", counts["patents"])
	}
	emptyActivities := []byte(`{"schema_version":1,"reviews":[],"invited_talks":[],"conference_service":[],"professional_service":[],"editorial_service":[]}`)
	counts, err := cvSectionCounts(saved, []byte("[]"), []byte("[]"), []byte("[]"), []byte("[]"), emptyActivities)
	if err != nil || counts["patents"] != 1 {
		t.Fatalf("CV patent count = %d, error %v", counts["patents"], err)
	}
	stored["visible_in_CV"] = false
	counts, err = cvSectionCounts(saved, []byte("[]"), []byte("[]"), []byte("[]"), []byte("[]"), emptyActivities)
	if err != nil || counts["patents"] != 0 {
		t.Fatalf("hidden CV patent count = %d, error %v", counts["patents"], err)
	}
	settings := SettingsDocument{MainPageSections: []string{"academic_activities", "patents", "skills"}, CVSections: []string{"academic_activities", "patents", "skills"}}
	autoHideEmptyMainSections(&settings, map[string]int{"academic_activities": 1, "skills": 1})
	autoHideEmptyCVSections(&settings, map[string]int{"academic_activities": 1, "skills": 1}, map[string]bool{"patents": true})
	if !reflect.DeepEqual(settings.MainPageSections, []string{"academic_activities", "patents", "skills"}) || !reflect.DeepEqual(settings.CVSections, settings.MainPageSections) {
		t.Fatal("empty patents changed configured section position")
	}
}

func TestPatentsEditorSaveAddsEditsAndRemovesProfileRows(t *testing.T) {
	root := newTempRepoFixture(t, fixtureSettingsJSON())
	app := &App{repoRoot: root}
	loaded, err := app.LoadEditorData()
	if err != nil {
		t.Fatal(err)
	}
	for _, items := range [][]any{{patentFixture()}, {map[string]any{"title_ko": "수정 발명", "type": "international", "inventor_ids": []any{"fixture-self"}, "registration_number": "KR-EXAMPLE-REG", "visible_in_CV": false, "future_field": "keep"}}, {}} {
		var profile map[string]any
		if err := json.Unmarshal(loaded.Profile, &profile); err != nil {
			t.Fatal(err)
		}
		profile["patents"] = items
		request := saveRequestRevisions(loaded)
		request.Profile, err = json.Marshal(profile)
		if err != nil {
			t.Fatal(err)
		}
		request.SaveProfile = true
		if _, err := app.SaveEditorData(request); err != nil {
			t.Fatal(err)
		}
		loaded, err = app.LoadEditorData()
		if err != nil {
			t.Fatal(err)
		}
		stored, err := readProfile(filepath.Join(root, "data", "profile.json"))
		if err != nil {
			t.Fatal(err)
		}
		patents := stored.Data["patents"].([]any)
		if len(patents) != len(items) {
			t.Fatalf("saved patents = %d, want %d", len(patents), len(items))
		}
		if len(items) != 0 {
			for field, want := range items[0].(map[string]any) {
				if !reflect.DeepEqual(patents[0].(map[string]any)[field], want) {
					t.Fatalf("saved patent %s = %v, want %v", field, patents[0].(map[string]any)[field], want)
				}
			}
		}
	}
}

func TestPatentsSettingsMigrationPreservesExistingOrder(t *testing.T) {
	for _, version := range []int{5, 6, 7, 8} {
		t.Run(string(rune('0'+version)), func(t *testing.T) {
			root := newTempRepoFixture(t, fixtureSettingsJSON())
			document, _, err := readSettings(filepath.Join(root, "data", "settings.json"))
			if err != nil {
				t.Fatal(err)
			}
			if err := normaliseLoadedSettings(&document); err != nil {
				t.Fatal(err)
			}
			without := func(items []string, excluded ...string) []string {
				result := []string{}
				for _, value := range items {
					if !stringSliceContains(excluded, value) {
						result = append(result, value)
					}
				}
				return result
			}
			document.SchemaVersion = version
			document.MainPageSections = without(document.MainPageSections, "patents")
			document.HiddenMainPageSections = without(document.HiddenMainPageSections, "patents")
			if version == 5 {
				document.HiddenMainPageSections = without(document.HiddenMainPageSections, "academic_activities")
			}
			document.CVSections = without(document.CVSections, "patents")
			document.HiddenCVSections = without(document.HiddenCVSections, "patents")
			mainBefore, cvBefore := append([]string{}, document.MainPageSections...), append([]string{}, document.CVSections...)
			if err := normaliseLoadedSettings(&document); err != nil {
				t.Fatal(err)
			}
			if document.SchemaVersion != 9 || !reflect.DeepEqual(document.MainPageSections, mainBefore) || !reflect.DeepEqual(document.CVSections, cvBefore) || !stringSliceContains(document.HiddenMainPageSections, "patents") || !stringSliceContains(document.HiddenCVSections, "patents") {
				t.Fatal("legacy settings order or new patent visibility changed")
			}
			document.HiddenMainPageSections = without(document.HiddenMainPageSections, "patents")
			if err := normaliseLoadedSettings(&document); err == nil {
				t.Fatalf("schema 9 missing section was accepted: %v", err)
			}
		})
	}
}
