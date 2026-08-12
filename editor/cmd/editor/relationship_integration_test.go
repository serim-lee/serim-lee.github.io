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

func TestRelationshipSaveCreatesHiddenIDsAndUsesThemFromNewAndEditedPublications(t *testing.T) {
	root := newRelationshipIntegrationRepo(t)
	app := &App{repoRoot: root}

	loaded, err := app.LoadEditorData()
	if err != nil {
		t.Fatalf("LoadEditorData() unexpected error: %v", err)
	}

	existingPerson := relationshipPersonByID(t, loaded.People, "existing-collaborator")
	editedPerson := relationshipPersonInput(existingPerson)
	editedPerson.NameKO = "기존 공동저자 수정"
	existingAward := relationshipAwardByID(t, loaded.Awards, "existing-award")
	editedAward := relationshipAwardInput(existingAward)
	editedAward.OrganizationKO = "기존 학회 수정"

	people := relationshipKeepPeople(loaded.People)
	for index := range people {
		if people[index].EditorKey == existingPerson.EditorKey {
			people[index].Person = editedPerson
		}
	}
	people = append(people, PersonSaveItem{Person: &PersonInput{
		ID:      "new-researcher",
		NameEN:  "New Researcher",
		NameKO:  "신규 연구자",
		IsSelf:  false,
		NotesEN: []string{"Kyung Hee University"},
		NotesKO: []string{"경희대학교"},
	}})

	awards := relationshipKeepAwards(loaded.Awards)
	for index := range awards {
		if awards[index].EditorKey == existingAward.EditorKey {
			awards[index].Award = editedAward
		}
	}
	awards = append(awards, AwardSaveItem{Award: &AwardInput{
		ID:             "new-award",
		Date:           "2026-12-15",
		TitleEN:        "New Paper Award",
		TitleKO:        "신규 논문상",
		OrganizationEN: "New Society",
		OrganizationKO: "신규 학회",
	}})

	modifiedSummary := relationshipPublicationByTitle(t, loaded.Publications, "Relationship paper")
	modified := relationshipPublicationInput(modifiedSummary)
	modified.TitleEN = "Modified relationship paper"
	modified.Date = "2027-03"
	modified.AuthorIDs = []string{"profile-owner", "new-researcher"}
	modified.AwardID = "new-award"
	newPublication := &PublicationInput{
		TitleEN:    "New relationship paper",
		TitleKO:    "신규 관계 논문",
		AbstractEN: "Created in the same transaction.",
		AbstractKO: "같은 저장에서 생성되었습니다.",
		KeywordsEN: []string{"relationship"},
		KeywordsKO: []string{"관계"},
		AuthorIDs:  []string{"new-researcher"},
		Date:       "2026-08-24",

		PublicationType: "international-journal",
		Topic:           "hvac",
		AwardID:         "new-award",
	}
	retained := relationshipPublicationByTitle(t, loaded.Publications, "Retained newer paper")

	request := relationshipSaveRequest(loaded)
	request.SavePeople = true
	request.People = people
	request.SaveAwards = true
	request.Awards = awards
	request.SavePublications = true
	// Intentionally scramble the request to prove that persisted and returned
	// publication rows use the website's automatic sort contract.
	request.Publications = []PublicationSaveItem{
		{Publication: newPublication},
		{EditorKey: retained.EditorKey},
		{EditorKey: modifiedSummary.EditorKey, Publication: modified},
	}

	saved, err := app.SaveEditorData(request)
	if err != nil {
		t.Fatalf("SaveEditorData() unexpected error: %v", err)
	}

	peopleRaw := relationshipReadFile(t, root, "people.json")
	awardsRaw := relationshipReadFile(t, root, "awards.json")
	publicationsRaw := relationshipReadFile(t, root, "publications.json")
	for name, check := range map[string]struct {
		got string
		old string
		raw []byte
	}{
		"people":       {saved.PeopleRevision, loaded.PeopleRevision, peopleRaw},
		"awards":       {saved.AwardsRevision, loaded.AwardsRevision, awardsRaw},
		"publications": {saved.PublicationsRevision, loaded.PublicationsRevision, publicationsRaw},
	} {
		if check.got == check.old {
			t.Errorf("%s revision did not change", name)
		}
		if want := revisionOf(check.raw); check.got != want {
			t.Errorf("%s response revision = %q, want disk revision %q", name, check.got, want)
		}
	}

	if got := relationshipPersonByID(t, saved.People, "new-researcher"); got.NameKO != "신규 연구자" {
		t.Errorf("new person response = %#v", got)
	}
	if got := relationshipAwardByID(t, saved.Awards, "new-award"); got.TitleKO != "신규 논문상" {
		t.Errorf("new award response = %#v", got)
	}
	if got := relationshipRawString(t, relationshipRawRowByID(t, peopleRaw, "new-researcher"), "name_ko"); got != "신규 연구자" {
		t.Errorf("stored new person name_ko = %q", got)
	}
	if got := relationshipRawString(t, relationshipRawRowByID(t, awardsRaw, "new-award"), "title_ko"); got != "신규 논문상" {
		t.Errorf("stored new award title_ko = %q", got)
	}
	if got := relationshipPersonByID(t, saved.People, "existing-collaborator"); got.ID != existingPerson.ID || got.NameKO != editedPerson.NameKO {
		t.Errorf("existing person ID/value was not retained: %#v", got)
	}
	if got := relationshipAwardByID(t, saved.Awards, "existing-award"); got.ID != existingAward.ID || got.OrganizationKO != editedAward.OrganizationKO {
		t.Errorf("existing award ID/value was not retained: %#v", got)
	}

	wantTitles := []string{"Modified relationship paper", "New relationship paper", "Retained newer paper"}
	if got := relationshipPublicationTitles(saved.Publications); !reflect.DeepEqual(got, wantTitles) {
		t.Errorf("response publication order = %#v, want %#v", got, wantTitles)
	}
	var diskPublications []PublicationInput
	if err := json.Unmarshal(publicationsRaw, &diskPublications); err != nil {
		t.Fatalf("decode stored publications: %v", err)
	}
	gotDiskTitles := make([]string, 0, len(diskPublications))
	for _, publication := range diskPublications {
		gotDiskTitles = append(gotDiskTitles, publication.TitleEN)
	}
	if !reflect.DeepEqual(gotDiskTitles, wantTitles) {
		t.Errorf("disk publication order = %#v, want %#v", gotDiskTitles, wantTitles)
	}
	for _, title := range wantTitles[:2] {
		publication := relationshipPublicationByTitle(t, saved.Publications, title)
		wantAuthors := map[string][]string{
			"Modified relationship paper": {"profile-owner", "new-researcher"},
			"New relationship paper":      {"new-researcher"},
		}[title]
		if !reflect.DeepEqual(publication.AuthorIDs, wantAuthors) || publication.AwardID != "new-award" {
			t.Errorf("%s response relationships = authors %#v, award %q", title, publication.AuthorIDs, publication.AwardID)
		}
		stored := relationshipRawPublicationByTitle(t, publicationsRaw, title)
		if got := relationshipRawStrings(t, stored, "author_ids"); !reflect.DeepEqual(got, wantAuthors) {
			t.Errorf("%s stored author_ids = %#v, want %#v", title, got, wantAuthors)
		}
		if got := relationshipRawString(t, stored, "award_id"); got != "new-award" {
			t.Errorf("%s stored award_id = %q, want new-award", title, got)
		}
	}

	relationshipAssertRawField(t, relationshipRawRowByID(t, peopleRaw, "existing-collaborator"), "future_person_field", map[string]any{
		"flags":  []any{true, float64(7)},
		"source": "legacy",
	})
	relationshipAssertRawField(t, relationshipRawRowByID(t, awardsRaw, "existing-award"), "future_award_field", map[string]any{
		"keep": true,
		"rank": float64(3),
	})
	relationshipAssertRawField(t, relationshipRawPublicationByTitle(t, publicationsRaw, "Modified relationship paper"), "future_publication_field", map[string]any{
		"nested": map[string]any{"keep": "verbatim"},
	})

	reloaded, err := app.LoadEditorData()
	if err != nil {
		t.Fatalf("LoadEditorData() after save unexpected error: %v", err)
	}
	if reloaded.PeopleRevision != saved.PeopleRevision || reloaded.AwardsRevision != saved.AwardsRevision ||
		reloaded.PublicationsRevision != saved.PublicationsRevision {
		t.Errorf("reload revisions differ from save response: people %q/%q, awards %q/%q, publications %q/%q",
			reloaded.PeopleRevision, saved.PeopleRevision,
			reloaded.AwardsRevision, saved.AwardsRevision,
			reloaded.PublicationsRevision, saved.PublicationsRevision)
	}
	if got := relationshipPublicationTitles(reloaded.Publications); !reflect.DeepEqual(got, wantTitles) {
		t.Errorf("reloaded publication order = %#v, want %#v", got, wantTitles)
	}
}

func TestRelationshipSaveRejectsDeletingReferencedPersonOrAwardWithoutPublicationChange(t *testing.T) {
	tests := []struct {
		name    string
		prepare func(EditorDataResponse, *SaveEditorDataRequest)
	}{
		{
			name: "person",
			prepare: func(loaded EditorDataResponse, request *SaveEditorDataRequest) {
				request.SavePeople = true
				for _, person := range loaded.People {
					if person.ID != "existing-collaborator" {
						request.People = append(request.People, PersonSaveItem{EditorKey: person.EditorKey})
					}
				}
			},
		},
		{
			name: "award",
			prepare: func(loaded EditorDataResponse, request *SaveEditorDataRequest) {
				request.SaveAwards = true
				for _, award := range loaded.Awards {
					if award.ID != "existing-award" {
						request.Awards = append(request.Awards, AwardSaveItem{EditorKey: award.EditorKey})
					}
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := newRelationshipIntegrationRepo(t)
			app := &App{repoRoot: root}
			loaded, err := app.LoadEditorData()
			if err != nil {
				t.Fatalf("LoadEditorData() unexpected error: %v", err)
			}
			request := relationshipSaveRequest(loaded)
			test.prepare(loaded, &request)
			before := relationshipSnapshotEditorFiles(t, root)

			if _, err := app.SaveEditorData(request); err == nil {
				t.Fatalf("SaveEditorData() deleting referenced %s error = nil, want rejection", test.name)
			}
			relationshipAssertEditorFilesEqual(t, root, before)
		})
	}
}

func TestRelationshipSaveAllowsReassignmentAndUnsetWithEntityDeletionInSameTransaction(t *testing.T) {
	root := newRelationshipIntegrationRepo(t)
	app := &App{repoRoot: root}
	loaded, err := app.LoadEditorData()
	if err != nil {
		t.Fatalf("LoadEditorData() unexpected error: %v", err)
	}

	request := relationshipSaveRequest(loaded)
	request.SavePeople = true
	for _, person := range loaded.People {
		if person.ID != "existing-collaborator" {
			request.People = append(request.People, PersonSaveItem{EditorKey: person.EditorKey})
		}
	}
	request.SaveAwards = true
	for _, award := range loaded.Awards {
		if award.ID != "existing-award" {
			request.Awards = append(request.Awards, AwardSaveItem{EditorKey: award.EditorKey})
		}
	}
	request.SavePublications = true
	for _, publication := range loaded.Publications {
		if publication.TitleEN == "Relationship paper" {
			edited := relationshipPublicationInput(publication)
			edited.AuthorIDs = []string{"profile-owner", "replacement-collaborator"}
			edited.AwardID = ""
			request.Publications = append(request.Publications, PublicationSaveItem{
				EditorKey: publication.EditorKey, Publication: edited,
			})
		} else {
			request.Publications = append(request.Publications, PublicationSaveItem{EditorKey: publication.EditorKey})
		}
	}

	saved, err := app.SaveEditorData(request)
	if err != nil {
		t.Fatalf("SaveEditorData() reassignment/deletion unexpected error: %v", err)
	}
	if relationshipHasPersonID(saved.People, "existing-collaborator") {
		t.Error("deleted person remains in response")
	}
	if relationshipHasAwardID(saved.Awards, "existing-award") {
		t.Error("deleted award remains in response")
	}
	publication := relationshipPublicationByTitle(t, saved.Publications, "Relationship paper")
	if !reflect.DeepEqual(publication.AuthorIDs, []string{"profile-owner", "replacement-collaborator"}) || publication.AwardID != "" {
		t.Errorf("updated publication relationships = authors %#v, award %q", publication.AuthorIDs, publication.AwardID)
	}

	peopleRaw := relationshipReadFile(t, root, "people.json")
	awardsRaw := relationshipReadFile(t, root, "awards.json")
	publicationsRaw := relationshipReadFile(t, root, "publications.json")
	if bytes.Contains(peopleRaw, []byte(`"id": "existing-collaborator"`)) {
		t.Error("deleted person remains on disk")
	}
	if bytes.Contains(awardsRaw, []byte(`"id": "existing-award"`)) {
		t.Error("deleted award remains on disk")
	}
	storedPublication := relationshipRawPublicationByTitle(t, publicationsRaw, "Relationship paper")
	if got := relationshipRawStrings(t, storedPublication, "author_ids"); !reflect.DeepEqual(got, publication.AuthorIDs) {
		t.Errorf("stored author_ids = %#v, want %#v", got, publication.AuthorIDs)
	}
	if got := relationshipRawString(t, storedPublication, "award_id"); got != "" {
		t.Errorf("stored award_id = %q, want empty", got)
	}
	relationshipAssertRawField(t, storedPublication, "future_publication_field", map[string]any{
		"nested": map[string]any{"keep": "verbatim"},
	})
}

func TestRelationshipSaveRejectsStalePeopleAwardsAndPublicationsRevisions(t *testing.T) {
	for _, name := range []string{"people.json", "awards.json", "publications.json"} {
		t.Run(name, func(t *testing.T) {
			root := newRelationshipIntegrationRepo(t)
			app := &App{repoRoot: root}
			loaded, err := app.LoadEditorData()
			if err != nil {
				t.Fatalf("LoadEditorData() unexpected error: %v", err)
			}

			path := filepath.Join(root, "data", name)
			external := append(relationshipReadFile(t, root, name), []byte(" \n")...)
			if err := os.WriteFile(path, external, 0o644); err != nil {
				t.Fatalf("simulate external edit of %s: %v", name, err)
			}
			before := relationshipSnapshotEditorFiles(t, root)
			request := relationshipSaveRequest(loaded)
			request.SavePeople = true
			request.People = relationshipKeepPeople(loaded.People)

			_, err = app.SaveEditorData(request)
			if err == nil {
				t.Fatalf("SaveEditorData() with stale %s revision error = nil, want conflict", name)
			}
			if !strings.Contains(err.Error(), name) {
				t.Errorf("stale conflict error %q does not identify %s", err, name)
			}
			relationshipAssertEditorFilesEqual(t, root, before)
		})
	}
}

func newRelationshipIntegrationRepo(t *testing.T) string {
	t.Helper()
	root := newTempRepoFixture(t, fixtureSettingsJSON())
	relationshipWriteFile(t, root, "people.json", []byte(`[
  {
    "id": "profile-owner",
    "name_en": "Profile Owner",
    "name_ko": "프로필 소유자",
    "is_self": true,
    "notes_en": ["Professor", "Kyung Hee University"],
    "notes_ko": ["교수", "경희대학교"],
    "future_self_field": {"keep": true}
  },
  {
    "id": "existing-collaborator",
    "name_en": "Existing Collaborator",
    "name_ko": "기존 공동저자",
    "is_self": false,
    "notes_en": ["Kyung Hee University"],
    "notes_ko": ["경희대학교"],
    "future_person_field": {"source": "legacy", "flags": [true, 7]}
  },
  {
    "id": "replacement-collaborator",
    "name_en": "Replacement Collaborator",
    "name_ko": "대체 공동저자",
    "is_self": false,
    "notes_en": ["Another University"],
    "notes_ko": ["다른 대학교"]
  }
]
`))
	relationshipWriteFile(t, root, "awards.json", []byte(`[
  {
    "id": "existing-award",
    "date": "2024-11-05",
    "title_en": "Existing Award",
    "title_ko": "기존 수상",
    "organization_en": "Existing Society",
    "organization_ko": "기존 학회",
    "future_award_field": {"keep": true, "rank": 3}
  },
  {
    "id": "retained-award",
    "date": "2023-02-01",
    "title_en": "Retained Award",
    "title_ko": "유지 수상",
    "organization_en": "Retained Society",
    "organization_ko": "유지 학회"
  }
]
`))
	relationshipWriteFile(t, root, "publications.json", []byte(`[
  {
    "title_en": "Relationship paper",
    "title_ko": "관계 논문",
    "abstract_en": "Existing abstract",
    "abstract_ko": "기존 초록",
    "keywords_en": ["relationship"],
    "keywords_ko": ["관계"],
    "author_ids": ["profile-owner", "existing-collaborator"],
    "date": "2024-06-01",

    "under_review": false,
    "in_press": false,
    "publication_type": "international-journal",
    "topic": "hvac",
    "award_id": "existing-award",
    "doi": "10.1234/relationship",
    "url": "https://example.org/relationship",
    "note": "",
    "future_publication_field": {"nested": {"keep": "verbatim"}}
  },
  {
    "title_en": "Retained newer paper",
    "title_ko": "유지 최신 논문",
    "abstract_en": "",
    "abstract_ko": "",
    "keywords_en": [],
    "keywords_ko": [],
    "author_ids": ["profile-owner"],
    "date": "2025",

    "under_review": false,
    "in_press": false,
    "publication_type": "domestic-journal",
    "topic": "hvac",
    "award_id": "",
    "doi": "",
    "url": "",
    "note": ""
  }
]
`))
	return root
}

func relationshipSaveRequest(loaded EditorDataResponse) SaveEditorDataRequest {
	return SaveEditorDataRequest{
		Settings:                   loaded.Settings,
		SettingsRevision:           loaded.SettingsRevision,
		ProfileRevision:            loaded.ProfileRevision,
		PeopleRevision:             loaded.PeopleRevision,
		AwardsRevision:             loaded.AwardsRevision,
		AcademicActivitiesRevision: loaded.AcademicActivitiesRevision,
		PublicationsRevision:       loaded.PublicationsRevision,
		ProjectsRevision:           loaded.ProjectsRevision,
		SoftwareRevision:           loaded.SoftwareRevision,
		BoardRevision:              loaded.BoardRevision,
	}
}

func relationshipKeepPeople(items []PeopleItemSummary) []PersonSaveItem {
	result := make([]PersonSaveItem, 0, len(items))
	for _, item := range items {
		result = append(result, PersonSaveItem{EditorKey: item.EditorKey})
	}
	return result
}

func relationshipKeepAwards(items []AwardItemSummary) []AwardSaveItem {
	result := make([]AwardSaveItem, 0, len(items))
	for _, item := range items {
		result = append(result, AwardSaveItem{EditorKey: item.EditorKey})
	}
	return result
}

func relationshipPersonInput(item PeopleItemSummary) *PersonInput {
	return &PersonInput{
		NameEN: item.NameEN, NameKO: item.NameKO, IsSelf: item.IsSelf,
		NotesEN: append([]string{}, item.NotesEN...), NotesKO: append([]string{}, item.NotesKO...),
	}
}

func relationshipAwardInput(item AwardItemSummary) *AwardInput {
	return &AwardInput{
		Date: item.Date, TitleEN: item.TitleEN, TitleKO: item.TitleKO,
		OrganizationEN: item.OrganizationEN, OrganizationKO: item.OrganizationKO,
	}
}

func relationshipPublicationInput(item PublicationItemSummary) *PublicationInput {
	return &PublicationInput{
		TitleEN: item.TitleEN, TitleKO: item.TitleKO,
		AbstractEN: item.AbstractEN, AbstractKO: item.AbstractKO,
		KeywordsEN: append([]string{}, item.KeywordsEN...), KeywordsKO: append([]string{}, item.KeywordsKO...),
		AuthorIDs: append([]string{}, item.AuthorIDs...), Date: item.Date,
		UnderReview: item.UnderReview, InPress: item.InPress, PublicationType: item.PublicationType,
		Topic: item.Topic, AwardID: item.AwardID, DOI: item.DOI, URL: item.URL, Note: item.Note,
	}
}

func relationshipPersonByID(t *testing.T, items []PeopleItemSummary, id string) PeopleItemSummary {
	t.Helper()
	for _, item := range items {
		if item.ID == id {
			return item
		}
	}
	t.Fatalf("person %q not found", id)
	return PeopleItemSummary{}
}

func relationshipAwardByID(t *testing.T, items []AwardItemSummary, id string) AwardItemSummary {
	t.Helper()
	for _, item := range items {
		if item.ID == id {
			return item
		}
	}
	t.Fatalf("award %q not found", id)
	return AwardItemSummary{}
}

func relationshipPublicationByTitle(t *testing.T, items []PublicationItemSummary, title string) PublicationItemSummary {
	t.Helper()
	for _, item := range items {
		if item.TitleEN == title {
			return item
		}
	}
	t.Fatalf("publication %q not found", title)
	return PublicationItemSummary{}
}

func relationshipPublicationTitles(items []PublicationItemSummary) []string {
	result := make([]string, 0, len(items))
	for _, item := range items {
		result = append(result, item.TitleEN)
	}
	return result
}

func relationshipHasPersonID(items []PeopleItemSummary, id string) bool {
	for _, item := range items {
		if item.ID == id {
			return true
		}
	}
	return false
}

func relationshipHasAwardID(items []AwardItemSummary, id string) bool {
	for _, item := range items {
		if item.ID == id {
			return true
		}
	}
	return false
}

var relationshipEditorFiles = []string{
	"settings.json",
	"people.json",
	"awards.json",
	"publications.json",
	"projects.json",
	"software.json",
	"board.json",
}

func relationshipSnapshotEditorFiles(t *testing.T, root string) map[string][]byte {
	t.Helper()
	result := make(map[string][]byte, len(relationshipEditorFiles))
	for _, name := range relationshipEditorFiles {
		result[name] = relationshipReadFile(t, root, name)
	}
	return result
}

func relationshipAssertEditorFilesEqual(t *testing.T, root string, want map[string][]byte) {
	t.Helper()
	for _, name := range relationshipEditorFiles {
		got := relationshipReadFile(t, root, name)
		if !bytes.Equal(got, want[name]) {
			t.Errorf("%s changed after rejected save\ngot:\n%s\nwant:\n%s", name, got, want[name])
		}
	}
}

func relationshipReadFile(t *testing.T, root, name string) []byte {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(root, "data", name))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return raw
}

func relationshipWriteFile(t *testing.T, root, name string, raw []byte) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(root, "data", name), raw, 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

func relationshipRawRowByID(t *testing.T, raw []byte, id string) map[string]json.RawMessage {
	t.Helper()
	var rows []map[string]json.RawMessage
	if err := json.Unmarshal(raw, &rows); err != nil {
		t.Fatalf("decode rows: %v", err)
	}
	for _, row := range rows {
		if relationshipRawString(t, row, "id") == id {
			return row
		}
	}
	t.Fatalf("raw row with id %q not found", id)
	return nil
}

func relationshipRawPublicationByTitle(t *testing.T, raw []byte, title string) map[string]json.RawMessage {
	t.Helper()
	var rows []map[string]json.RawMessage
	if err := json.Unmarshal(raw, &rows); err != nil {
		t.Fatalf("decode publications: %v", err)
	}
	for _, row := range rows {
		if relationshipRawString(t, row, "title_en") == title {
			return row
		}
	}
	t.Fatalf("raw publication %q not found", title)
	return nil
}

func relationshipRawString(t *testing.T, row map[string]json.RawMessage, name string) string {
	t.Helper()
	var value string
	if err := json.Unmarshal(row[name], &value); err != nil {
		t.Fatalf("decode string field %q: %v", name, err)
	}
	return value
}

func relationshipRawStrings(t *testing.T, row map[string]json.RawMessage, name string) []string {
	t.Helper()
	var value []string
	if err := json.Unmarshal(row[name], &value); err != nil {
		t.Fatalf("decode string array field %q: %v", name, err)
	}
	return value
}

func relationshipAssertRawField(t *testing.T, row map[string]json.RawMessage, name string, want any) {
	t.Helper()
	raw, ok := row[name]
	if !ok {
		t.Fatalf("unknown field %q was dropped", name)
	}
	var got any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("decode unknown field %q: %v", name, err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("unknown field %q = %#v, want %#v", name, got, want)
	}
}
