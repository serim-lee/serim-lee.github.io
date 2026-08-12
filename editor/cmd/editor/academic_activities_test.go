package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestAcademicActivitiesLoadOrderAndNoPersistentIDs(t *testing.T) {
	snapshot, err := readAcademicActivitiesBytes(fixtureAcademicActivitiesJSON())
	if err != nil {
		t.Fatalf("readAcademicActivitiesBytes() unexpected error: %v", err)
	}
	wantOrder := []string{"editorial_service", "professional_service", "conference_service", "invited_talks", "reviews"}
	gotOrder := make([]string, 0, len(snapshot.Rows))
	for _, row := range snapshot.Rows {
		gotOrder = append(gotOrder, row.Category)
	}
	if !reflect.DeepEqual(gotOrder, wantOrder) {
		t.Fatalf("category order = %#v, want %#v", gotOrder, wantOrder)
	}
	encoded, err := json.Marshal(academicActivityItemSummaries(snapshot))
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(encoded, []byte(`"id"`)) {
		t.Fatalf("editor summaries unexpectedly persisted IDs: %s", encoded)
	}
	if bytes.Contains(encoded, []byte(`"review_count"`)) || !bytes.Contains(encoded, []byte(`"completed_dates"`)) {
		t.Fatalf("editor summaries do not use completed_dates exclusively: %s", encoded)
	}
	for _, item := range academicActivityItemSummaries(snapshot) {
		if !item.VisibleInCV {
			t.Fatalf("missing visible_in_CV did not default true for %s", item.Category)
		}
	}
	invalidVisibility := bytes.Replace(
		fixtureAcademicActivitiesJSON(),
		[]byte(`{"journal_en":"Review Journal"`),
		[]byte(`{"visible_in_CV":"yes","journal_en":"Review Journal"`),
		1,
	)
	if _, err := readAcademicActivitiesBytes(invalidVisibility); err == nil || !strings.Contains(err.Error(), "visible_in_CV") {
		t.Fatalf("non-Boolean academic visible_in_CV was accepted: %v", err)
	}
	withID := bytes.Replace(
		fixtureAcademicActivitiesJSON(),
		[]byte(`{"journal_en":"Review Journal"`),
		[]byte(`{"id":"legacy-review","journal_en":"Review Journal"`),
		1,
	)
	if _, err := readAcademicActivitiesBytes(withID); err == nil || !strings.Contains(err.Error(), "id 필드") {
		t.Fatalf("persistent academic activity ID was accepted: %v", err)
	}
	legacyCount := bytes.Replace(
		fixtureAcademicActivitiesJSON(),
		[]byte(`"completed_dates":["2026-01-15","2026-05-20","2026-08-10"]`),
		[]byte(`"review_count":3`),
		1,
	)
	if _, err := readAcademicActivitiesBytes(legacyCount); err == nil || !strings.Contains(err.Error(), "completed_dates") {
		t.Fatalf("legacy review_count was accepted: %v", err)
	}
}

func TestAcademicActivitiesCRUDPreservesUnknownFieldsAndNeverAddsIDs(t *testing.T) {
	current, err := readAcademicActivitiesBytes(fixtureAcademicActivitiesJSON())
	if err != nil {
		t.Fatal(err)
	}
	byCategory := make(map[string]academicActivityRow)
	for _, row := range current.Rows {
		byCategory[row.Category] = row
	}
	editorial := byCategory["editorial_service"].Item.AcademicActivityInput
	editorial.RoleKO = "편집위원장"
	newReview := AcademicActivityInput{
		Category: "reviews", JournalEN: "New Review Journal", JournalKO: "신규 리뷰 저널",
		CompletedDates: []string{"2026-02-01", "2026-04-12"},
	}
	request := []AcademicActivitySaveItem{
		{EditorKey: byCategory["editorial_service"].Key, Activity: &editorial},
		// Omitting professional_service deletes it.
		{EditorKey: byCategory["conference_service"].Key},
		{EditorKey: byCategory["invited_talks"].Key},
		{Activity: &newReview},
	}
	encoded, err := buildAcademicActivitiesSave(current, request)
	if err != nil {
		t.Fatalf("buildAcademicActivitiesSave() unexpected error: %v", err)
	}
	if bytes.Contains(encoded, []byte(`"id"`)) {
		t.Fatalf("saved academic activities unexpectedly contain IDs: %s", encoded)
	}
	var root map[string]any
	if err := json.Unmarshal(encoded, &root); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(root["future_top_level"], map[string]any{"preserve": true}) {
		t.Fatalf("unknown top-level field was lost: %#v", root["future_top_level"])
	}
	editorialRows := root["editorial_service"].([]any)
	if len(editorialRows) != 1 || editorialRows[0].(map[string]any)["future_item"] != "retain" {
		t.Fatalf("unknown item field was lost: %#v", editorialRows)
	}
	if got := editorialRows[0].(map[string]any)["role_ko"]; got != "편집위원장" {
		t.Fatalf("edited role_ko = %#v", got)
	}
	if got := root["professional_service"].([]any); len(got) != 0 {
		t.Fatalf("omitted professional service was not deleted: %#v", got)
	}
	reviews := root["reviews"].([]any)
	if len(reviews) != 1 || reviews[0].(map[string]any)["journal_en"] != "New Review Journal" {
		t.Fatalf("new review was not saved: %#v", reviews)
	}
	if dates := reviews[0].(map[string]any)["completed_dates"].([]any); len(dates) != 2 {
		t.Fatalf("new review completed dates = %#v", dates)
	}
	if _, exists := reviews[0].(map[string]any)["review_count"]; exists {
		t.Fatalf("review_count survived completed_dates save: %#v", reviews[0])
	}
	if visible, ok := reviews[0].(map[string]any)["visible_in_CV"].(bool); !ok || !visible {
		t.Fatalf("new academic activity visible_in_CV = %#v, want true", reviews[0].(map[string]any)["visible_in_CV"])
	}
	if _, err := readAcademicActivitiesBytes(encoded); err != nil {
		t.Fatalf("saved document did not round-trip: %v", err)
	}
}

func TestAcademicActivitiesSavePreservesManualOrderWithinCategory(t *testing.T) {
	current, err := readAcademicActivitiesBytes(manualOrderAcademicActivitiesJSON())
	if err != nil {
		t.Fatal(err)
	}
	request := []AcademicActivitySaveItem{
		{EditorKey: current.Rows[2].Key},
		{EditorKey: current.Rows[0].Key},
		{EditorKey: current.Rows[1].Key},
	}
	encoded, err := buildAcademicActivitiesSave(current, request)
	if err != nil {
		t.Fatalf("buildAcademicActivitiesSave() unexpected error: %v", err)
	}
	var root map[string]any
	if err := json.Unmarshal(encoded, &root); err != nil {
		t.Fatal(err)
	}
	reviews := root["reviews"].([]any)
	got := make([]string, 0, len(reviews))
	for _, value := range reviews {
		got = append(got, value.(map[string]any)["journal_en"].(string))
	}
	want := []string{"Third Journal", "First Journal", "Second Journal"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("saved review order = %#v, want %#v", got, want)
	}
	if got := reviews[0].(map[string]any)["future_item"]; got != "retain-on-reorder" {
		t.Fatalf("reordered item lost its unknown field: %#v", reviews[0])
	}
	if _, err := readAcademicActivitiesBytes(encoded); err != nil {
		t.Fatalf("reordered document did not round-trip: %v", err)
	}
}

func TestAcademicActivitiesRejectInvalidDatesCountsAndRequiredLabels(t *testing.T) {
	empty, err := readAcademicActivitiesBytes(emptyAcademicActivitiesJSON())
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name  string
		input AcademicActivityInput
	}{
		{name: "review dates empty", input: AcademicActivityInput{Category: "reviews", JournalEN: "Journal", CompletedDates: []string{}}},
		{name: "review invalid date", input: AcademicActivityInput{Category: "reviews", JournalEN: "Journal", CompletedDates: []string{"2026-02-30"}}},
		{name: "review duplicate date", input: AcademicActivityInput{Category: "reviews", JournalEN: "Journal", CompletedDates: []string{"2026-02-01", "2026-02-01"}}},
		{name: "invited talk date", input: AcademicActivityInput{Category: "invited_talks", EventEN: "Event", TopicEN: "Topic", Date: "2026-02-30"}},
		{name: "conference reversed period", input: AcademicActivityInput{Category: "conference_service", ConferenceEN: "Conference", RoleEN: "Chair", StartDate: "2026-08-20", EndDate: "2026-08-19"}},
		{name: "professional missing organization", input: AcademicActivityInput{Category: "professional_service", RoleEN: "Member", StartDate: "2026-01-01"}},
		{name: "editorial invalid start", input: AcademicActivityInput{Category: "editorial_service", JournalEN: "Journal", RoleEN: "Editor", StartDate: "2026-13-01"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := buildAcademicActivitiesSave(empty, []AcademicActivitySaveItem{{Activity: &test.input}})
			if err == nil {
				t.Fatal("buildAcademicActivitiesSave() error = nil, want validation rejection")
			}
		})
	}
}

func TestSaveEditorDataAcademicRevisionConflictAndLastDeleteAutoHide(t *testing.T) {
	t.Run("revision conflict", func(t *testing.T) {
		root := newTempRepoFixture(t, fixtureSettingsJSON())
		path := filepath.Join(root, "data", "academic_activities.json")
		writeFixtureFile(t, path, fixtureAcademicActivitiesJSON())
		app := &App{repoRoot: root}
		loaded, err := app.LoadEditorData()
		if err != nil {
			t.Fatal(err)
		}
		external := bytes.Replace(fixtureAcademicActivitiesJSON(), []byte(`"2026-08-10"`), []byte(`"2026-08-11"`), 1)
		if bytes.Equal(external, fixtureAcademicActivitiesJSON()) {
			t.Fatal("external edit fixture did not change")
		}
		writeFixtureFile(t, path, external)
		request := saveRequestRevisions(loaded)
		request.AcademicActivities = []AcademicActivitySaveItem{}
		request.SaveAcademicActivities = true
		if _, err := app.SaveEditorData(request); err == nil || !strings.Contains(err.Error(), "academic_activities.json") {
			t.Fatalf("SaveEditorData() conflict error = %v", err)
		}
		if got, readErr := os.ReadFile(path); readErr != nil || !bytes.Equal(got, external) {
			t.Fatalf("external academic edit was overwritten: %v\n%s", readErr, got)
		}
	})

	t.Run("reorder persists without settings mutation", func(t *testing.T) {
		root := newTempRepoFixture(t, fixtureSettingsJSON())
		path := filepath.Join(root, "data", "academic_activities.json")
		settingsPath := filepath.Join(root, "data", "settings.json")
		writeFixtureFile(t, path, manualOrderAcademicActivitiesJSON())
		app := &App{repoRoot: root}
		loaded, err := app.LoadEditorData()
		if err != nil {
			t.Fatal(err)
		}
		if len(loaded.AcademicActivities) != 3 {
			t.Fatalf("loaded academic activities = %d, want 3", len(loaded.AcademicActivities))
		}
		normaliseRequest := saveRequestRevisions(loaded)
		normaliseRequest.Settings = loaded.Settings
		normaliseRequest.SaveSettings = true
		loaded, err = app.SaveEditorData(normaliseRequest)
		if err != nil {
			t.Fatalf("normalise fixture settings: %v", err)
		}
		settingsBefore, err := os.ReadFile(settingsPath)
		if err != nil {
			t.Fatal(err)
		}
		request := saveRequestRevisions(loaded)
		request.AcademicActivities = []AcademicActivitySaveItem{
			{EditorKey: loaded.AcademicActivities[2].EditorKey},
			{EditorKey: loaded.AcademicActivities[0].EditorKey},
			{EditorKey: loaded.AcademicActivities[1].EditorKey},
		}
		request.SaveAcademicActivities = true
		saved, err := app.SaveEditorData(request)
		if err != nil {
			t.Fatal(err)
		}
		if saved.AcademicActivitiesRevision == loaded.AcademicActivitiesRevision {
			t.Fatal("academic activities revision did not change after reordering")
		}
		got := make([]string, 0, len(saved.AcademicActivities))
		for _, item := range saved.AcademicActivities {
			got = append(got, item.JournalEN)
		}
		want := []string{"Third Journal", "First Journal", "Second Journal"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("reloaded review order = %#v, want %#v", got, want)
		}
		settingsAfter, err := os.ReadFile(settingsPath)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(settingsBefore, settingsAfter) {
			t.Fatal("reordering academic activities unexpectedly changed settings.json")
		}
	})

	t.Run("CV visibility auto hides academic section", func(t *testing.T) {
		root := newTempRepoFixture(t, fixtureSettingsJSON())
		writeFixtureFile(t, filepath.Join(root, "data", "academic_activities.json"), fixtureAcademicActivitiesJSON())
		app := &App{repoRoot: root}
		loaded, err := app.LoadEditorData()
		if err != nil {
			t.Fatal(err)
		}
		if !containsString(loaded.Settings.CVSections, "academic_activities") {
			t.Fatalf("academic CV section was not available: %#v", loaded.Settings.CVSections)
		}
		hidden := false
		request := saveRequestRevisions(loaded)
		request.AcademicActivities = make([]AcademicActivitySaveItem, 0, len(loaded.AcademicActivities))
		for _, item := range loaded.AcademicActivities {
			request.AcademicActivities = append(request.AcademicActivities, AcademicActivitySaveItem{
				EditorKey: item.EditorKey, VisibleInCV: &hidden,
			})
		}
		request.SaveAcademicActivities = true
		saved, err := app.SaveEditorData(request)
		if err != nil {
			t.Fatal(err)
		}
		for _, item := range saved.AcademicActivities {
			if item.VisibleInCV {
				t.Fatalf("academic activity remained visible in CV: %#v", item)
			}
		}
		if containsString(saved.Settings.CVSections, "academic_activities") ||
			!containsString(saved.Settings.HiddenCVSections, "academic_activities") {
			t.Fatalf("academic CV section with no selected items was not auto-hidden: visible %#v hidden %#v", saved.Settings.CVSections, saved.Settings.HiddenCVSections)
		}
	})

	t.Run("last delete auto hides", func(t *testing.T) {
		root := newTempRepoFixture(t, fixtureSettingsJSON())
		path := filepath.Join(root, "data", "academic_activities.json")
		writeFixtureFile(t, path, []byte(`{
  "schema_version": 1,
  "reviews": [{"journal_en":"Journal","journal_ko":"저널","completed_dates":["2026-01-01"]}],
  "invited_talks": [],
  "conference_service": [],
  "professional_service": [],
  "editorial_service": []
}
`))
		app := &App{repoRoot: root}
		loaded, err := app.LoadEditorData()
		if err != nil {
			t.Fatal(err)
		}
		settings := loaded.Settings
		settings.HiddenMainPageSections = removeString(settings.HiddenMainPageSections, "academic_activities")
		settings.MainPageSections = append(settings.MainPageSections, "academic_activities")
		request := saveRequestRevisions(loaded)
		request.Settings = settings
		request.SaveSettings = true
		request.AcademicActivities = []AcademicActivitySaveItem{}
		request.SaveAcademicActivities = true
		saved, err := app.SaveEditorData(request)
		if err != nil {
			t.Fatal(err)
		}
		if containsString(saved.Settings.MainPageSections, "academic_activities") ||
			!containsString(saved.Settings.HiddenMainPageSections, "academic_activities") {
			t.Fatalf("empty academic activities were not auto-hidden: visible %#v hidden %#v", saved.Settings.MainPageSections, saved.Settings.HiddenMainPageSections)
		}
		stored, err := readAcademicActivities(path)
		if err != nil || len(stored.Rows) != 0 {
			t.Fatalf("stored activities after last delete = %#v, %v", stored.Rows, err)
		}
	})
}

func removeString(items []string, target string) []string {
	result := make([]string, 0, len(items))
	for _, item := range items {
		if item != target {
			result = append(result, item)
		}
	}
	return result
}

func emptyAcademicActivitiesJSON() []byte {
	return []byte(`{
  "schema_version": 1,
  "reviews": [],
  "invited_talks": [],
  "conference_service": [],
  "professional_service": [],
  "editorial_service": []
}
`)
}

func manualOrderAcademicActivitiesJSON() []byte {
	return []byte(`{
  "schema_version": 1,
  "editorial_service": [],
  "professional_service": [],
  "conference_service": [],
  "invited_talks": [],
  "reviews": [
    {"journal_en":"First Journal","journal_ko":"첫 번째 저널","completed_dates":["2026-01-01"]},
    {"journal_en":"Second Journal","journal_ko":"두 번째 저널","completed_dates":["2026-02-01"]},
    {"journal_en":"Third Journal","journal_ko":"세 번째 저널","completed_dates":["2026-03-01"],"future_item":"retain-on-reorder"}
  ]
}
`)
}

func fixtureAcademicActivitiesJSON() []byte {
	return []byte(`{
  "schema_version": 1,
  "reviews": [
    {"journal_en":"Review Journal","journal_ko":"리뷰 저널","completed_dates":["2026-01-15","2026-05-20","2026-08-10"]}
  ],
  "invited_talks": [
    {"event_en":"Seminar","event_ko":"세미나","topic_en":"Simulation","topic_ko":"시뮬레이션","date":"2026-08-15"}
  ],
  "conference_service": [
    {"conference_en":"Conference","conference_ko":"학술대회","role_en":"Chair","role_ko":"좌장","start_date":"2025-08-27","end_date":"2025-08-29"}
  ],
  "professional_service": [
    {"organization_en":"Institute","organization_ko":"기관","role_en":"Member","role_ko":"위원","start_date":"2025-01-01","end_date":""}
  ],
  "editorial_service": [
    {"journal_en":"Editorial Journal","journal_ko":"편집 저널","role_en":"Editor","role_ko":"편집위원","start_date":"2026-01-01","end_date":"2026-12-31","future_item":"retain"}
  ],
  "future_top_level": {"preserve": true}
}
`)
}
