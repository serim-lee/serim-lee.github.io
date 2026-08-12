package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func publicationCatalogLinkRequest(loaded EditorDataResponse, first *PublicationInput) SaveEditorDataRequest {
	request := saveRequestRevisions(loaded)
	request.SavePublications = true
	for index, item := range loaded.Publications {
		row := PublicationSaveItem{EditorKey: item.EditorKey}
		if index == 0 {
			row.Publication = first
		}
		request.Publications = append(request.Publications, row)
	}
	return request
}

func TestPublicationCatalogLinksSaveReopenAndUnlinkAtomically(t *testing.T) {
	for _, publicationType := range []string{"international-journal", "domestic-journal", "international-conference", "domestic-conference"} {
		t.Run(publicationType, func(t *testing.T) {
			root := newTempRepoFixture(t, fixtureSettingsJSON())
			app := &App{repoRoot: root}
			loaded, err := app.LoadEditorData()
			if err != nil {
				t.Fatal(err)
			}
			input := publicationInputFromSummary(loaded.Publications[0])
			input.PublicationType = publicationType
			conferenceID := "SASBE 2024"
			if publicationType == "domestic-conference" {
				conferenceID = "AIK Conf. 2024F"
			}
			journal := publicationType == "international-journal" || publicationType == "domestic-journal"
			if journal {
				input.JournalID = "test-journal"
			} else {
				input.ConferenceID = conferenceID
			}
			request := publicationCatalogLinkRequest(loaded, input)
			request.SaveSettings = true
			request.Settings = loaded.Settings
			request.Settings.Journals = []PublicationVenue{{ID: "test-journal", NameEN: "Test Journal"}}
			request.Settings.Conferences = []PublicationVenue{{ID: conferenceID, NameKO: "시험 학술대회"}}
			if _, err := app.SaveEditorData(request); err != nil {
				t.Fatal(err)
			}
			reopened, err := app.LoadEditorData()
			if err != nil {
				t.Fatal(err)
			}
			var linked PublicationItemSummary
			for _, item := range reopened.Publications {
				if item.TitleEN == input.TitleEN {
					linked = item
				}
			}
			if linked.JournalID != input.JournalID || linked.ConferenceID != input.ConferenceID {
				t.Fatal("catalog selection changed after reopening")
			}
			if journal && reopened.Usage.Journals[input.JournalID] != 1 {
				t.Fatal("journal usage not counted")
			}
			if !journal && reopened.Usage.Conferences[input.ConferenceID] != 1 {
				t.Fatal("conference usage not counted")
			}

			settingsPath := filepath.Join(root, "data", "settings.json")
			publicationsPath := filepath.Join(root, "data", "publications.json")
			beforeSettings, _ := os.ReadFile(settingsPath)
			beforePublications, _ := os.ReadFile(publicationsPath)
			remove := saveRequestRevisions(reopened)
			remove.SaveSettings = true
			remove.Settings = reopened.Settings
			remove.Settings.Journals = []PublicationVenue{}
			remove.Settings.Conferences = []PublicationVenue{}
			if _, err := app.SaveEditorData(remove); err == nil {
				t.Fatal("referenced catalog deletion was accepted")
			}
			if _, err := app.SaveSettings(SaveSettingsRequest{Settings: remove.Settings, Revision: reopened.SettingsRevision}); err == nil {
				t.Fatal("legacy settings API allowed referenced catalog deletion")
			}
			afterSettings, _ := os.ReadFile(settingsPath)
			afterPublications, _ := os.ReadFile(publicationsPath)
			if !bytes.Equal(beforeSettings, afterSettings) || !bytes.Equal(beforePublications, afterPublications) {
				t.Fatal("rejected deletion wrote data")
			}

			remove.SavePublications = true
			for _, item := range reopened.Publications {
				row := PublicationSaveItem{EditorKey: item.EditorKey}
				if item.EditorKey == linked.EditorKey {
					row.Publication = publicationInputFromSummary(item)
					row.Publication.JournalID = ""
					row.Publication.ConferenceID = ""
				}
				remove.Publications = append(remove.Publications, row)
			}
			if _, err := app.SaveEditorData(remove); err != nil {
				t.Fatalf("unlink and catalog deletion in one save: %v", err)
			}
			if _, err := app.LoadEditorData(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestPublicationCatalogLinksRejectMissingOrWrongKind(t *testing.T) {
	for _, mutate := range []func(*PublicationInput){
		func(item *PublicationInput) {
			item.PublicationType = "international-journal"
			item.JournalID = "missing"
		},
		func(item *PublicationInput) {
			item.PublicationType = "international-conference"
			item.JournalID = "test-journal"
		},
		func(item *PublicationInput) {
			item.PublicationType = "domestic-journal"
			item.ConferenceID = "test-conference"
		},
		func(item *PublicationInput) { item.PublicationType = "domestic-journal"; item.JournalID = "invalid id" },
	} {
		root := newTempRepoFixture(t, fixtureSettingsJSON())
		app := &App{repoRoot: root}
		loaded, err := app.LoadEditorData()
		if err != nil {
			t.Fatal(err)
		}
		input := publicationInputFromSummary(loaded.Publications[0])
		mutate(input)
		request := publicationCatalogLinkRequest(loaded, input)
		request.SaveSettings = true
		request.Settings = loaded.Settings
		request.Settings.Journals = []PublicationVenue{{ID: "test-journal", NameEN: "Journal"}}
		request.Settings.Conferences = []PublicationVenue{{ID: "test-conference", NameEN: "Conference"}}
		if _, err := app.SaveEditorData(request); err == nil {
			t.Fatalf("invalid catalog selection was accepted: %#v", input)
		}
	}
}
