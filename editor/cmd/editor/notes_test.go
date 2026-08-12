package main

import (
	"context"
	"encoding/json"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

type independentNotesDomain struct {
	name      string
	fixture   func(*testing.T) string
	read      func([]byte, string) error
	configure func(*SaveEditorDataRequest, EditorDataResponse, []string, []string)
	notes     func(EditorDataResponse) ([]string, []string)
}

func independentNotesDomains() []independentNotesDomain {
	return []independentNotesDomain{
		{
			name:    "projects",
			fixture: newBoardRepoFixture,
			read: func(raw []byte, root string) error {
				_, err := readProjectsBytes(raw, root, projectFixtureThemeIDs())
				return err
			},
			configure: func(request *SaveEditorDataRequest, loaded EditorDataResponse, notesEN, notesKR []string) {
				item := loaded.Projects[0]
				request.SaveProjects = true
				request.Projects = []ProjectSaveItem{{
					EditorKey: item.EditorKey,
					Project: &ProjectInput{
						TitleEN: item.TitleEN, TitleKO: item.TitleKO,
						StartDate: item.StartDate, EndDate: item.EndDate, Theme: item.Theme,
						FunderEN: item.FunderEN, FunderKO: item.FunderKO,
						NotesEN: notesEN, NotesKR: notesKR, Media: []ProjectMediaInput{},
					},
				}}
			},
			notes: func(response EditorDataResponse) ([]string, []string) {
				return response.Projects[0].NotesEN, response.Projects[0].NotesKR
			},
		},
		{
			name:    "software",
			fixture: newSoftwareRepoFixture,
			read: func(raw []byte, root string) error {
				_, err := readSoftwareBytes(raw, root)
				return err
			},
			configure: func(request *SaveEditorDataRequest, loaded EditorDataResponse, notesEN, notesKR []string) {
				item := loaded.Software[0]
				input := softwareInputFromSummary(item)
				input.NotesEN, input.NotesKR = notesEN, notesKR
				request.SaveSoftware = true
				request.Software = []SoftwareSaveItem{{EditorKey: item.EditorKey, Software: input}}
			},
			notes: func(response EditorDataResponse) ([]string, []string) {
				return response.Software[0].NotesEN, response.Software[0].NotesKR
			},
		},
	}
}

func TestIndependentNotesLoadSaveRoundTrip(t *testing.T) {
	for _, domain := range independentNotesDomains() {
		t.Run(domain.name, func(t *testing.T) {
			for _, test := range []struct {
				name             string
				notesEN, notesKR []string
			}{
				{"English1Korean3", []string{"English one"}, []string{"국문 하나", "국문 둘", "국문 셋"}},
				{"English3Korean1", []string{"English one", "English two", "English three"}, []string{"국문 하나"}},
				{"English1Korean100", []string{"English one"}, repeatedNativeNotes(100)},
				{"English100Korean1", repeatedNativeNotes(100), []string{"국문 하나"}},
			} {
				t.Run(test.name, func(t *testing.T) {
					root := domain.fixture(t)
					path := filepath.Join(root, "data", domain.name+".json")
					raw := nativeNotesJSON(t, readFixtureFile(t, path), test.notesEN, test.notesKR)
					writeFixtureFile(t, path, raw)
					app := &App{repoRoot: root}
					t.Cleanup(func() { app.shutdown(context.Background()) })
					loaded, err := app.LoadEditorData()
					if err != nil {
						t.Fatalf("LoadEditorData() with independent notes: %v", err)
					}
					assertNativeNotes(t, domain, loaded, test.notesEN, test.notesKR)

					wantEN, wantKR := updatedNativeNotes(test.notesEN), updatedNativeNotes(test.notesKR)
					request := nativeNotesSaveRequest(loaded)
					domain.configure(&request, loaded, paddedNativeNotes(wantEN), paddedNativeNotes(wantKR))
					saved, err := app.SaveEditorData(request)
					if err != nil {
						t.Fatalf("SaveEditorData() with independent notes: %v", err)
					}
					assertNativeNotes(t, domain, saved, wantEN, wantKR)
					reloaded, err := app.LoadEditorData()
					if err != nil {
						t.Fatalf("LoadEditorData() after save: %v", err)
					}
					assertNativeNotes(t, domain, reloaded, wantEN, wantKR)
					var rows []map[string]json.RawMessage
					if err := json.Unmarshal(readFixtureFile(t, path), &rows); err != nil {
						t.Fatal(err)
					}
					if _, exists := rows[0]["notes_ko"]; exists {
						t.Fatal("save introduced notes_ko instead of canonical notes_kr")
					}
				})
			}
		})
	}
}

func TestIndependentNotesRejectInvalidEntriesOnLoadAndSave(t *testing.T) {
	for _, domain := range independentNotesDomains() {
		t.Run(domain.name, func(t *testing.T) {
			for _, language := range []string{"notes_en", "notes_kr"} {
				t.Run(language, func(t *testing.T) {
					for _, test := range []struct {
						name  string
						notes []string
					}{
						{"null", nil},
						{"empty", []string{}},
						{"tooMany", repeatedNativeNotes(101)},
						{"lastUnmatchedBlank", []string{"one", "two", " \n\t "}},
						{"lastUnmatchedTooLong", []string{"one", "two", strings.Repeat("a", 5001)}},
						{"lastUnmatchedControl", []string{"one", "two", "invalid\x01note"}},
					} {
						t.Run(test.name, func(t *testing.T) {
							root := domain.fixture(t)
							path := filepath.Join(root, "data", domain.name+".json")
							notesEN, notesKR := []string{"English"}, []string{"국문"}
							if language == "notes_en" {
								notesEN = test.notes
							} else {
								notesKR = test.notes
							}
							raw := nativeNotesJSON(t, readFixtureFile(t, path), notesEN, notesKR)
							if err := domain.read(raw, root); err == nil {
								t.Fatal("load accepted invalid independent notes")
							}
							app := &App{repoRoot: root}
							t.Cleanup(func() { app.shutdown(context.Background()) })
							loaded, err := app.LoadEditorData()
							if err != nil {
								t.Fatalf("load valid fixture: %v", err)
							}
							request := nativeNotesSaveRequest(loaded)
							domain.configure(&request, loaded, notesEN, notesKR)
							before := snapshotBoardRepo(t, root)
							if _, err := app.SaveEditorData(request); err == nil {
								t.Fatal("save accepted invalid independent notes")
							}
							assertBoardRepoSnapshot(t, root, before, "invalid independent notes")
						})
					}
				})
			}
		})
	}
}

func nativeNotesSaveRequest(loaded EditorDataResponse) SaveEditorDataRequest {
	return SaveEditorDataRequest{
		Settings: loaded.Settings, SettingsRevision: loaded.SettingsRevision,
		ProfileRevision: loaded.ProfileRevision, PeopleRevision: loaded.PeopleRevision,
		AwardsRevision: loaded.AwardsRevision, AcademicActivitiesRevision: loaded.AcademicActivitiesRevision,
		PublicationsRevision: loaded.PublicationsRevision, ProjectsRevision: loaded.ProjectsRevision,
		SoftwareRevision: loaded.SoftwareRevision, BoardRevision: loaded.BoardRevision,
	}
}

func nativeNotesJSON(t *testing.T, raw []byte, notesEN, notesKR []string) []byte {
	t.Helper()
	var rows []map[string]any
	if err := json.Unmarshal(raw, &rows); err != nil {
		t.Fatal(err)
	}
	rows[0]["notes_en"], rows[0]["notes_kr"] = notesEN, notesKR
	encoded, err := json.Marshal(rows)
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}

func assertNativeNotes(t *testing.T, domain independentNotesDomain, response EditorDataResponse, wantEN, wantKR []string) {
	t.Helper()
	gotEN, gotKR := domain.notes(response)
	if !reflect.DeepEqual(gotEN, wantEN) || !reflect.DeepEqual(gotKR, wantKR) {
		t.Fatalf("independent notes = %#v / %#v, want %#v / %#v", gotEN, gotKR, wantEN, wantKR)
	}
}

func repeatedNativeNotes(count int) []string {
	notes := make([]string, count)
	for index := range notes {
		notes[index] = "Valid note"
	}
	return notes
}

func updatedNativeNotes(notes []string) []string {
	result := make([]string, len(notes))
	for index, note := range notes {
		result[index] = note + " updated"
	}
	return result
}

func paddedNativeNotes(notes []string) []string {
	result := make([]string, len(notes))
	for index, note := range notes {
		result[index] = " " + note + " "
	}
	return result
}
