package main

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func patentPeopleRepoFixture(t *testing.T) string {
	t.Helper()
	root := newTempRepoFixture(t, fixtureSettingsJSON())
	peoplePath := filepath.Join(root, "data", "people.json")
	var people []map[string]any
	if err := json.Unmarshal(readFixtureFile(t, peoplePath), &people); err != nil {
		t.Fatal(err)
	}
	people = append(people, map[string]any{"id": "patent-only-inventor", "name_en": "Patent Inventor", "name_ko": "특허 발명자", "is_self": false, "notes_en": []string{}, "notes_ko": []string{}})
	writeJSONFixture(t, peoplePath, people)
	profilePath := filepath.Join(root, "data", "profile.json")
	var profile map[string]any
	if err := json.Unmarshal(readFixtureFile(t, profilePath), &profile); err != nil {
		t.Fatal(err)
	}
	patent := patentFixture()
	patent["inventor_ids"] = []any{"patent-only-inventor", "fixture-self"}
	patent["future_field"] = map[string]any{"keep": true}
	profile["patents"] = []any{patent}
	writeJSONFixture(t, profilePath, profile)
	return root
}

func patentPeopleProfileInput(t *testing.T, loaded EditorDataResponse, ids []any) json.RawMessage {
	t.Helper()
	var profile map[string]any
	if err := json.Unmarshal(loaded.Profile, &profile); err != nil {
		t.Fatal(err)
	}
	items := profile["patents"].([]any)
	if len(items) == 0 {
		items = []any{patentFixture()}
	}
	items[0].(map[string]any)["inventor_ids"] = ids
	profile["patents"] = items
	raw, err := json.Marshal(profile)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func patentPeopleSnapshot(t *testing.T, root string) map[string][]byte {
	t.Helper()
	before := relationshipSnapshotEditorFiles(t, root)
	before["profile.json"] = readFixtureFile(t, filepath.Join(root, "data", "profile.json"))
	before["academic_activities.json"] = readFixtureFile(t, filepath.Join(root, "data", "academic_activities.json"))
	return before
}

func assertPatentPeopleSnapshotUnchanged(t *testing.T, root string, before map[string][]byte) {
	t.Helper()
	for filename, want := range before {
		if got := readFixtureFile(t, filepath.Join(root, "data", filename)); !bytes.Equal(got, want) {
			t.Errorf("rejected inventor transaction changed %s", filename)
		}
	}
}

func TestPatentsPeopleReferencesRejectUnknownOnLoadAndSave(t *testing.T) {
	t.Run("load", func(t *testing.T) {
		root := patentPeopleRepoFixture(t)
		app := &App{repoRoot: root}
		loaded, err := app.LoadEditorData()
		if err != nil {
			t.Fatal(err)
		}
		writeFixtureFile(t, filepath.Join(root, "data", "profile.json"), patentPeopleProfileInput(t, loaded, []any{"unknown-inventor"}))
		if _, err := app.LoadEditorData(); err == nil || !strings.Contains(err.Error(), "inventor_ids") {
			t.Fatalf("unknown inventor load error = %v", err)
		}
	})
	t.Run("save", func(t *testing.T) {
		root := patentPeopleRepoFixture(t)
		app := &App{repoRoot: root}
		loaded, err := app.LoadEditorData()
		if err != nil {
			t.Fatal(err)
		}
		request := saveRequestRevisions(loaded)
		request.SaveProfile = true
		request.Profile = patentPeopleProfileInput(t, loaded, []any{"unknown-inventor"})
		before := patentPeopleSnapshot(t, root)
		if _, err := app.SaveEditorData(request); err == nil || !strings.Contains(err.Error(), "inventor_ids") {
			t.Fatalf("unknown inventor save error = %v", err)
		}
		assertPatentPeopleSnapshotUnchanged(t, root, before)
	})
}

func TestPatentsPeopleDeletionRequiresSameTransactionReferenceChange(t *testing.T) {
	root := patentPeopleRepoFixture(t)
	app := &App{repoRoot: root}
	loaded, err := app.LoadEditorData()
	if err != nil {
		t.Fatal(err)
	}
	request := saveRequestRevisions(loaded)
	request.SavePeople = true
	for _, person := range loaded.People {
		if person.ID != "patent-only-inventor" {
			request.People = append(request.People, PersonSaveItem{EditorKey: person.EditorKey})
		}
	}
	before := patentPeopleSnapshot(t, root)
	if _, err := app.SaveEditorData(request); err == nil || !strings.Contains(err.Error(), "inventor_ids") {
		t.Fatalf("deleting a patent-only inventor without changing profile was accepted: %v", err)
	}
	assertPatentPeopleSnapshotUnchanged(t, root, before)
	request.SaveProfile = true
	request.Profile = patentPeopleProfileInput(t, loaded, []any{"fixture-self"})
	if _, err := app.SaveEditorData(request); err != nil {
		t.Fatalf("same transaction inventor reassignment and deletion failed: %v", err)
	}
	reloaded, err := app.LoadEditorData()
	if err != nil {
		t.Fatal(err)
	}
	if relationshipHasPersonID(reloaded.People, "patent-only-inventor") {
		t.Fatal("unreferenced inventor was not deleted")
	}
	stored, err := readProfile(filepath.Join(root, "data", "profile.json"))
	if err != nil {
		t.Fatal(err)
	}
	patent := stored.Data["patents"].([]any)[0].(map[string]any)
	if !reflect.DeepEqual(patent["inventor_ids"], []any{"fixture-self"}) || !reflect.DeepEqual(patent["future_field"], map[string]any{"keep": true}) {
		t.Fatal("inventor reassignment lost references or future patent metadata")
	}
}

func TestPatentsPeopleCreateAndRenamePreserveOrderedReferences(t *testing.T) {
	root := newTempRepoFixture(t, fixtureSettingsJSON())
	app := &App{repoRoot: root}
	loaded, err := app.LoadEditorData()
	if err != nil {
		t.Fatal(err)
	}
	request := saveRequestRevisions(loaded)
	request.SavePeople, request.SaveProfile = true, true
	request.People = append(relationshipKeepPeople(loaded.People), PersonSaveItem{Person: &PersonInput{ID: "new-inventor", NameEN: "New Inventor", NameKO: "새 발명자", NotesEN: []string{}, NotesKO: []string{}}})
	request.Profile = patentPeopleProfileInput(t, loaded, []any{"new-inventor", "fixture-self"})
	saved, err := app.SaveEditorData(request)
	if err != nil {
		t.Fatalf("creating inventor and patent together failed: %v", err)
	}
	profilePath := filepath.Join(root, "data", "profile.json")
	profileBefore := readFixtureFile(t, profilePath)
	inventor := relationshipPersonByID(t, saved.People, "new-inventor")
	input := personInputFromSummary(inventor)
	input.NameEN, input.NameKO = "Renamed Inventor", "바뀐 발명자"
	rename := saveRequestRevisions(saved)
	rename.SavePeople = true
	for _, person := range saved.People {
		row := PersonSaveItem{EditorKey: person.EditorKey}
		if person.ID == inventor.ID {
			row.Person = input
		}
		rename.People = append(rename.People, row)
	}
	if _, err := app.SaveEditorData(rename); err != nil {
		t.Fatal(err)
	}
	reloaded, err := app.LoadEditorData()
	if err != nil {
		t.Fatal(err)
	}
	if relationshipPersonByID(t, reloaded.People, "new-inventor").NameEN != "Renamed Inventor" || !bytes.Equal(profileBefore, readFixtureFile(t, profilePath)) {
		t.Fatal("renaming an inventor changed the stable patent relationship or failed to save the name")
	}
	profile, err := readProfile(profilePath)
	if err != nil {
		t.Fatal(err)
	}
	if got := profile.Data["patents"].([]any)[0].(map[string]any)["inventor_ids"]; !reflect.DeepEqual(got, []any{"new-inventor", "fixture-self"}) {
		t.Fatalf("inventor order = %v", got)
	}
	mutated := saveRequestRevisions(reloaded)
	mutated.SavePeople = true
	for _, person := range reloaded.People {
		row := PersonSaveItem{EditorKey: person.EditorKey}
		if person.ID == inventor.ID {
			row.Person = personInputFromSummary(person)
			row.Person.ID = "changed-id"
		}
		mutated.People = append(mutated.People, row)
	}
	before := patentPeopleSnapshot(t, root)
	if _, err := app.SaveEditorData(mutated); err == nil {
		t.Fatal("changing a referenced person's immutable ID was accepted")
	}
	assertPatentPeopleSnapshotUnchanged(t, root, before)
}

func TestPatentsOptionalInventorsAndRetiredFields(t *testing.T) {
	for _, numberField := range []string{"application_number", "publication_number", "registration_number"} {
		item := patentFixture()
		delete(item, "application_number")
		delete(item, "inventor_ids")
		item[numberField] = "EXAMPLE-001"
		item["status"], item["jurisdiction"] = "stale-status", "old-free-text"
		item["inventors_en"], item["inventors_ko"] = "Old Inventor", "옛 발명자"
		item["future_field"] = map[string]any{"keep": true}
		snapshot, err := readProfileBytes(patentProfileFixture(t, []any{item}))
		if err != nil {
			t.Fatal(err)
		}
		if got := snapshot.Data["patents"].([]any)[0].(map[string]any)["inventor_ids"]; !reflect.DeepEqual(got, []any{}) {
			t.Fatalf("missing inventor IDs did not default to an empty array: %v", got)
		}
		if err := validateProfilePeopleReferences(snapshot, map[string]bool{}); err != nil {
			t.Fatal(err)
		}
		editor, err := profileForEditor(snapshot)
		if err != nil {
			t.Fatal(err)
		}
		encoded, saved, err := buildProfileSave(snapshot.Raw)
		if err != nil {
			t.Fatal(err)
		}
		for _, field := range []string{"status", "jurisdiction", "inventors_en", "inventors_ko"} {
			if bytes.Contains(editor, []byte(`"`+field+`"`)) || bytes.Contains(encoded, []byte(`"`+field+`"`)) {
				t.Fatalf("retired patent field %s survived the editor or save payload", field)
			}
		}
		patent := saved.Data["patents"].([]any)[0].(map[string]any)
		if patent[numberField] != "EXAMPLE-001" || !reflect.DeepEqual(patent["future_field"], item["future_field"]) || !reflect.DeepEqual(patent["inventor_ids"], []any{}) {
			t.Fatal("retiring derived fields lost patent data or empty inventor array")
		}
	}
}
