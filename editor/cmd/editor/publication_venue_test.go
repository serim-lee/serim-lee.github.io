package main

import (
	"encoding/json"
	"testing"
)

func TestPublicationRejectsRetiredManualVenueFields(t *testing.T) {
	for _, field := range []string{"venue", "venue_en", "venue_kr"} {
		row := publicationFixtureItem("Paper", "2026", false, false)
		row[field] = "Old manual venue"
		raw, _ := json.Marshal([]any{row})
		if _, err := readPublicationsBytes(raw, publicationPersonIDs(), publicationAwardIDs(), publicationTopicIDs()); err == nil {
			t.Errorf("accepted retired %s field", field)
		}
	}
}

func TestPublicationDateUsesCatalogAndBibliographyPriority(t *testing.T) {
	settings := SettingsDocument{Conferences: []PublicationVenue{
		{ID: "ASIM 2024", ProceedingsDate: "2025-04", EndDate: "2024-12-10"},
		{ID: "WSC 2022", EndDate: "2022-12-14"},
	}}
	for _, test := range []struct {
		item publicationDocumentItem
		want string
	}{
		{publicationDocumentItem{PublicationType: "international-journal", Date: "1999", IssueDate: "2025-10-21", OnlineDate: "2025-08-13"}, "2025-10-21"},
		{publicationDocumentItem{PublicationType: "domestic-journal", OnlineDate: "2024-02"}, "2024-02"},
		{publicationDocumentItem{PublicationType: "international-conference", ConferenceID: "ASIM 2024", Date: "1999"}, "2025-04"},
		{publicationDocumentItem{PublicationType: "domestic-conference", ConferenceID: "WSC 2022"}, "2022-12-14"},
		{publicationDocumentItem{PublicationType: "international-journal", Date: "1999"}, ""},
		{publicationDocumentItem{PublicationType: "international-conference", Date: "1999", ConferenceID: "Unknown"}, ""},
	} {
		if got := effectivePublicationDate(test.item, settings); got != test.want {
			t.Errorf("date = %q, want %q for %#v", got, test.want, test.item)
		}
	}
}

func TestPublicationSaveOrdersAndCachesDerivedDates(t *testing.T) {
	older := publicationFixtureItem("Old issue, new online", "2099", false, false)
	older["issue_date"], older["online_date"] = "2024-10", "2026-03"
	newer := publicationFixtureItem("New issue", "1900", false, false)
	newer["issue_date"] = "2025-01"
	raw, _ := json.Marshal([]any{older, newer})
	snapshot, err := readPublicationsBytes(raw, publicationPersonIDs(), publicationAwardIDs(), publicationTopicIDs())
	if err != nil {
		t.Fatal(err)
	}
	items := publicationItemSummaries(snapshot)
	encoded, err := buildPublicationSaveLocked(snapshot, []PublicationSaveItem{
		{EditorKey: items[0].EditorKey}, {EditorKey: items[1].EditorKey},
	}, publicationPersonIDs(), publicationAwardIDs(), publicationTopicIDs(), SettingsDocument{})
	if err != nil {
		t.Fatal(err)
	}
	saved, err := readPublicationsBytes(encoded, publicationPersonIDs(), publicationAwardIDs(), publicationTopicIDs())
	if err != nil {
		t.Fatal(err)
	}
	if saved.Rows[0].Item.TitleEN != "New issue" || saved.Rows[0].Item.Date != "2025-01" || saved.Rows[1].Item.Date != "2024-10" {
		t.Fatalf("saved rows did not use issue dates: %#v", saved.Rows)
	}
}
