package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPublicationCatalogProceedingsNumberingSaveReopenAndClear(t *testing.T) {
	root := newTempRepoFixture(t, fixtureSettingsJSON())
	app := &App{repoRoot: root}
	loaded, err := app.LoadEditorData()
	if err != nil {
		t.Fatal(err)
	}
	publicationPath := filepath.Join(root, "data", "publications.json")
	before, err := os.ReadFile(publicationPath)
	if err != nil {
		t.Fatal(err)
	}
	var conference PublicationVenue
	if err := json.Unmarshal([]byte(`{"id":"AIK Conf. 2024F","name_en":"Conference","proceedings_title_en":"Proceedings","proceedings_publisher_en":"Society","proceedings_volume":" 12A ","proceedings_issue":" Suppl. 2 ","proceedings_series_en":"Research Series","proceedings_series_volume":" I–IV ","future_numbering":{"keep":true}}`), &conference); err != nil {
		t.Fatal(err)
	}
	request := saveRequestRevisions(loaded)
	request.SaveSettings = true
	request.Settings = loaded.Settings
	request.Settings.Conferences = []PublicationVenue{conference}
	if _, err := app.SaveEditorData(request); err != nil {
		t.Fatal(err)
	}
	reopened, err := app.LoadEditorData()
	if err != nil {
		t.Fatal(err)
	}
	stored := reopened.Settings.Conferences[0]
	if stored.ID != "AIK Conf. 2024F" || stored.ProceedingsVolume != "12A" || stored.ProceedingsIssue != "Suppl. 2" || stored.ProceedingsSeriesVolume != "I–IV" ||
		stored.ProceedingsTitleEN != "Proceedings" || stored.ProceedingsPublisherEN != "Society" || stored.ProceedingsSeriesEN != "Research Series" {
		t.Fatalf("proceedings numbering or adjacent metadata changed: %#v", stored)
	}
	for _, field := range []string{"proceedings_volume", "proceedings_issue", "proceedings_series_volume"} {
		if _, exists := stored.extraFields[field]; exists {
			t.Fatalf("known field %s was also retained as unknown metadata", field)
		}
	}
	// Reproduce the Wails JSON round trip before clearing all optional fields.
	raw, err := json.Marshal(reopened.Settings)
	if err != nil {
		t.Fatal(err)
	}
	var draft SettingsDocument
	if err := json.Unmarshal(raw, &draft); err != nil {
		t.Fatal(err)
	}
	draft.Conferences[0].ProceedingsVolume = " "
	draft.Conferences[0].ProceedingsIssue = " "
	draft.Conferences[0].ProceedingsSeriesVolume = " "
	request = saveRequestRevisions(reopened)
	request.SaveSettings = true
	request.Settings = draft
	if _, err := app.SaveEditorData(request); err != nil {
		t.Fatal(err)
	}
	reopened, err = app.LoadEditorData()
	if err != nil {
		t.Fatal(err)
	}
	stored = reopened.Settings.Conferences[0]
	if stored.ProceedingsVolume != "" || stored.ProceedingsIssue != "" || stored.ProceedingsSeriesVolume != "" || stored.ID != "AIK Conf. 2024F" {
		t.Fatal("cleared optional numbering reappeared or changed the venue ID")
	}
	var future map[string]bool
	if err := json.Unmarshal(stored.extraFields["future_numbering"], &future); err != nil || !future["keep"] {
		t.Fatal("unknown proceedings metadata was lost")
	}
	raw, err = json.Marshal(stored)
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"proceedings_volume", "proceedings_issue", "proceedings_series_volume"} {
		if bytes.Contains(raw, []byte(`"`+field+`"`)) {
			t.Fatalf("empty optional %s was not omitted from JSON", field)
		}
	}
	after, err := os.ReadFile(publicationPath)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("catalog update modified publications.json")
	}
}

func TestPublicationCatalogProceedingsNumberingValidation(t *testing.T) {
	for _, field := range []struct {
		name string
		set  func(*PublicationVenue, string)
	}{
		{"proceedings_volume", func(item *PublicationVenue, value string) { item.ProceedingsVolume = value }},
		{"proceedings_issue", func(item *PublicationVenue, value string) { item.ProceedingsIssue = value }},
		{"proceedings_series_volume", func(item *PublicationVenue, value string) { item.ProceedingsSeriesVolume = value }},
	} {
		t.Run(field.name, func(t *testing.T) {
			for _, value := range []string{"", "12A", "Suppl. 2", " I–IV ", strings.Repeat("가", 500)} {
				item := PublicationVenue{ID: "conference", NameEN: "Conference"}
				field.set(&item, value)
				if _, err := normalisePublicationCatalog([]PublicationVenue{item}, nil, "conferences"); err != nil {
					t.Fatalf("valid free text %q rejected: %v", value, err)
				}
			}
			for _, value := range []string{strings.Repeat("가", 501), "12\x00A", "Suppl.\x7f2"} {
				item := PublicationVenue{ID: "conference", NameEN: "Conference"}
				field.set(&item, value)
				if _, err := normalisePublicationCatalog([]PublicationVenue{item}, nil, "conferences"); err == nil {
					t.Fatal("invalid proceedings numbering text was accepted")
				}
			}
			for _, invalid := range []any{nil, 12, true, []string{"12"}, map[string]string{"value": "12"}} {
				raw, err := json.Marshal(map[string]any{"id": "conference", "name_en": "Conference", field.name: invalid})
				if err != nil {
					t.Fatal(err)
				}
				var item PublicationVenue
				if err := json.Unmarshal(raw, &item); err == nil {
					t.Fatalf("non-string %s accepted: %s", field.name, raw)
				}
			}
		})
	}
}

func TestPublicationCatalogsSaveAndReopenWithoutChangingPublications(t *testing.T) {
	root := newTempRepoFixture(t, fixtureSettingsJSON())
	app := &App{repoRoot: root}
	loaded, err := app.LoadEditorData()
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Settings.Journals == nil || loaded.Settings.Conferences == nil {
		t.Fatal("legacy settings should expose empty catalog arrays")
	}
	publicationPath := filepath.Join(root, "data", "publications.json")
	before, err := os.ReadFile(publicationPath)
	if err != nil {
		t.Fatal(err)
	}
	var journal PublicationVenue
	if err := json.Unmarshal([]byte(`{"id":"test-journal","name_en":" Test Journal ","name_ko":" 시험 저널 ","publisher_en":"Publisher","publisher_ko":"발행기관","issn":"1234-567x","url":"https://example.org/journal","indexing":["SCIE"],"future_field":{"keep":true}}`), &journal); err != nil {
		t.Fatal(err)
	}
	request := saveRequestRevisions(loaded)
	request.SaveSettings = true
	request.Settings = loaded.Settings
	request.Settings.Journals = []PublicationVenue{journal}
	request.Settings.Conferences = []PublicationVenue{{
		ID: "test-conference-2026", NameEN: "Test Conference", NameKO: "시험 학술대회", StartDate: "2026-10-18", EndDate: "2026-10-20",
		LocationEN: "Seoul, Korea", LocationKO: "대한민국 서울", URL: "https://example.org/conference",
		ProceedingsPublisherEN: "Proceedings Publisher", ProceedingsPublisherKO: "논문집 발행기관",
		ProceedingsTitleEN: "Conference Proceedings", ProceedingsTitleKO: "학술대회 논문집",
		ProceedingsSeriesEN: "Research Series", ProceedingsSeriesKO: "연구 시리즈", ProceedingsDate: "2026-11-01",
		ProceedingsISSN: "2345-678X", Indexing: []string{"KCI"},
	}}
	saved, err := app.SaveEditorData(request)
	if err != nil {
		t.Fatal(err)
	}
	reopened, err := app.LoadEditorData()
	if err != nil {
		t.Fatal(err)
	}
	savedCatalogs, _ := json.Marshal([][]PublicationVenue{saved.Settings.Journals, saved.Settings.Conferences})
	reopenedCatalogs, _ := json.Marshal([][]PublicationVenue{reopened.Settings.Journals, reopened.Settings.Conferences})
	if !bytes.Equal(savedCatalogs, reopenedCatalogs) {
		t.Fatalf("catalog metadata changed after reopening\nsaved: %s\nreopened: %s", savedCatalogs, reopenedCatalogs)
	}
	if reopened.Settings.Journals[0].NameEN != "Test Journal" || reopened.Settings.Journals[0].NameKO != "시험 저널" || reopened.Settings.Journals[0].ISSN != "1234-567X" {
		t.Fatal("catalog text and ISSN were not normalized")
	}
	assertExtraKept := func(item PublicationVenue) {
		t.Helper()
		var extra map[string]bool
		if err := json.Unmarshal(item.extraFields["future_field"], &extra); err != nil || !extra["keep"] {
			t.Fatal("unknown catalog metadata was lost")
		}
	}
	assertExtraKept(reopened.Settings.Journals[0])
	// Reproduce a Wails JSON round trip, then rename the editable reference ID.
	raw, err := json.Marshal(reopened.Settings)
	if err != nil {
		t.Fatal(err)
	}
	var draft SettingsDocument
	if err := json.Unmarshal(raw, &draft); err != nil {
		t.Fatal(err)
	}
	draft.Journals[0].ID = "renamed-journal"
	request = saveRequestRevisions(reopened)
	request.SaveSettings = true
	request.Settings = draft
	renamed, err := app.SaveEditorData(request)
	if err != nil {
		t.Fatal(err)
	}
	if renamed.Settings.Journals[0].ID != "renamed-journal" {
		t.Fatal("renaming a reference ID lost the entry")
	}
	assertExtraKept(renamed.Settings.Journals[0])
	after, err := os.ReadFile(publicationPath)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("catalog-only saving changed publications.json")
	}
	// Old clients omit the new arrays. Explicit empty arrays are a deliberate deletion.
	draft = renamed.Settings
	draft.Journals = nil
	draft.Conferences = nil
	preserved, err := normaliseForSave(draft, renamed.Settings, renamed.Usage)
	if err != nil || len(preserved.Journals) != 1 || len(preserved.Conferences) != 1 {
		t.Fatalf("omitted catalog arrays did not preserve saved data: %v", err)
	}
	request = saveRequestRevisions(renamed)
	request.SaveSettings = true
	request.Settings = renamed.Settings
	request.Settings.Journals = []PublicationVenue{}
	request.Settings.Conferences = []PublicationVenue{}
	cleared, err := app.SaveEditorData(request)
	if err != nil || cleared.Settings.Journals == nil || len(cleared.Settings.Journals) != 0 || len(cleared.Settings.Conferences) != 0 {
		t.Fatalf("explicit empty arrays did not clear the catalogs: %v", err)
	}
}

func TestPublicationCatalogValidationAndAtomicRejection(t *testing.T) {
	valid := PublicationVenue{ID: "conference-2026", NameEN: "Conference", Indexing: []string{"SCOPUS"}}
	for label, mutate := range map[string]func(*PublicationVenue){
		"missing ID":               func(item *PublicationVenue) { item.ID = "" },
		"invalid ID":               func(item *PublicationVenue) { item.ID = "Invalid/ID" },
		"missing name":             func(item *PublicationVenue) { item.NameEN = " " },
		"invalid Korean text":      func(item *PublicationVenue) { item.ProceedingsTitleKO = "제어\x00문자" },
		"unsafe URL":               func(item *PublicationVenue) { item.URL = "javascript:alert(1)" },
		"invalid ISSN":             func(item *PublicationVenue) { item.ISSN = "12345678" },
		"invalid proceedings ISSN": func(item *PublicationVenue) { item.ProceedingsISSN = "invalid" },
		"invalid date":             func(item *PublicationVenue) { item.StartDate = "2026-02-30" },
		"invalid proceedings date": func(item *PublicationVenue) { item.ProceedingsDate = "2026-13-01" },
		"missing start date":       func(item *PublicationVenue) { item.EndDate = "2026-10-18" },
		"reversed period":          func(item *PublicationVenue) { item.StartDate = "2026-10-20"; item.EndDate = "2026-10-18" },
		"unknown index":            func(item *PublicationVenue) { item.Indexing = []string{"UNKNOWN"} },
		"multiple indexes":         func(item *PublicationVenue) { item.Indexing = []string{"SCIE", "SCOPUS"} },
	} {
		t.Run(label, func(t *testing.T) {
			item := valid
			mutate(&item)
			if _, err := normalisePublicationCatalog([]PublicationVenue{item}, nil, "conferences"); err == nil {
				t.Fatal("invalid catalog entry was accepted")
			}
		})
	}
	root := newTempRepoFixture(t, fixtureSettingsJSON())
	app := &App{repoRoot: root}
	loaded, err := app.LoadEditorData()
	if err != nil {
		t.Fatal(err)
	}
	settingsPath := filepath.Join(root, "data", "settings.json")
	before, _ := os.ReadFile(settingsPath)
	request := saveRequestRevisions(loaded)
	request.SaveSettings = true
	request.Settings = loaded.Settings
	request.Settings.Journals = []PublicationVenue{valid, valid}
	if _, err := app.SaveEditorData(request); err == nil {
		t.Fatal("duplicate IDs were accepted")
	}
	after, _ := os.ReadFile(settingsPath)
	if !bytes.Equal(before, after) {
		t.Fatal("invalid catalog save changed settings.json")
	}
}

func TestPublicationCatalogProceedingsDatePrecision(t *testing.T) {
	for _, date := range []string{"2026", "2026-10", "2024-02-29", ""} {
		item := PublicationVenue{ID: "conference", NameEN: "Conference", ProceedingsDate: date}
		items, err := normalisePublicationCatalog([]PublicationVenue{item}, nil, "conferences")
		if err != nil || items[0].ProceedingsDate != date {
			t.Fatalf("date precision %q was not preserved: %v", date, err)
		}
	}
	for _, date := range []string{"2025-02-29", "2026-13", "2026/10", "2026-1"} {
		item := PublicationVenue{ID: "conference", NameEN: "Conference", ProceedingsDate: date}
		if _, err := normalisePublicationCatalog([]PublicationVenue{item}, nil, "conferences"); err == nil {
			t.Fatalf("invalid proceedings date %q was accepted", date)
		}
	}
	item := PublicationVenue{ID: "conference", NameEN: "Conference", StartDate: "2026-10"}
	if _, err := normalisePublicationCatalog([]PublicationVenue{item}, nil, "conferences"); err == nil {
		t.Fatal("event start dates must retain full date precision")
	}
}

func TestPublicationCatalogBilingualLegacyMigration(t *testing.T) {
	for _, value := range []string{"Legacy entry", "기존 항목"} {
		fields := map[string]any{"id": "legacy-entry", "indexing": []string{}}
		bases := []string{"name", "publisher", "location", "proceedings_publisher", "proceedings_title", "proceedings_series"}
		for _, base := range bases {
			fields[base] = value
		}
		raw, _ := json.Marshal(fields)
		var item PublicationVenue
		if err := json.Unmarshal(raw, &item); err != nil {
			t.Fatal(err)
		}
		normalized, err := normalisePublicationCatalog([]PublicationVenue{item}, nil, "conferences")
		if err != nil {
			t.Fatal(err)
		}
		raw, _ = json.Marshal(normalized[0])
		var saved map[string]any
		if err := json.Unmarshal(raw, &saved); err != nil {
			t.Fatal(err)
		}
		language := "_en"
		if value == "기존 항목" {
			language = "_ko"
		}
		for _, base := range bases {
			if _, exists := saved[base]; exists || saved[base+language] != value {
				t.Fatalf("legacy %s was not preserved in %s: %s", base, language, raw)
			}
		}
	}
	var cleared PublicationVenue
	if err := json.Unmarshal([]byte(`{"id":"cleared","name":"Stale name","name_en":"","name_ko":""}`), &cleared); err != nil {
		t.Fatal(err)
	}
	if cleared.NameEN != "" || cleared.NameKO != "" {
		t.Fatal("legacy name revived a deliberately cleared bilingual pair")
	}
}

func TestConferenceCatalogIDsAcceptAbbreviationsAndLegacyIDs(t *testing.T) {
	for _, id := range []string{"ASIM 2024", "ACEEE 2024", "SASBE 2024", "AIK Conf. 2024F", "AIK Conf. 2024S", "KIAEBS 2025", "KORMS-KIIE-KSS 2024", "legacy-conference-2024"} {
		item := PublicationVenue{ID: id, NameEN: "Conference"}
		items, err := normalisePublicationCatalog([]PublicationVenue{item}, nil, "conferences")
		if err != nil || items[0].ID != id {
			t.Fatalf("conference ID %q was not preserved: %v", id, err)
		}
	}
	for _, id := range []string{"AIK/2024", "AIK  2024", "AIK_2024", "AIK... 2024", ".AIK 2024"} {
		item := PublicationVenue{ID: id, NameEN: "Conference"}
		if _, err := normalisePublicationCatalog([]PublicationVenue{item}, nil, "conferences"); err == nil {
			t.Fatalf("invalid conference ID %q was accepted", id)
		}
	}
	if validPublicationCatalogID("journals", "AIK Conf. 2024F") {
		t.Fatal("journal IDs must retain their canonical format")
	}
}
