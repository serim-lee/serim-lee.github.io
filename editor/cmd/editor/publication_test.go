package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestReadPublicationsReturnsFlatEditableRowsInSourceOrder(t *testing.T) {
	path, raw := newPublicationFixture(t)
	snapshot, err := readPublications(path, publicationPersonIDs(), publicationAwardIDs(), publicationTopicIDs())
	if err != nil {
		t.Fatalf("readPublications() unexpected error: %v", err)
	}
	if revisionOf(snapshot.Raw) != revisionOf(raw) {
		t.Error("publication snapshot does not retain the source revision")
	}
	items := publicationItemSummaries(snapshot)
	if len(items) != 4 {
		t.Fatalf("publication summaries length = %d, want 4", len(items))
	}
	wantTitles := []string{"Older paper", "Review paper", "Press paper", "Newest paper"}
	for index, wantTitle := range wantTitles {
		if items[index].TitleEN != wantTitle {
			t.Errorf("summary[%d] title = %q, want source-order title %q", index, items[index].TitleEN, wantTitle)
		}
		if items[index].EditorKey == "" {
			t.Errorf("summary[%d] has no opaque editor key", index)
		}
		if items[index].KeywordsEN == nil || items[index].KeywordsKO == nil || items[index].AuthorIDs == nil {
			t.Errorf("summary[%d] has nil editable arrays: %#v", index, items[index])
		}
	}
	if !reflect.DeepEqual(items[0].AuthorIDs, []string{"jo-sg", "park-cs"}) ||
		!reflect.DeepEqual(items[0].KeywordsEN, []string{"building energy", "control"}) {
		t.Errorf("flat publication arrays = authors %#v, keywords %#v", items[0].AuthorIDs, items[0].KeywordsEN)
	}
	if items[1].Date != "" || !items[1].UnderReview || items[1].InPress {
		t.Errorf("review publication status fields = %#v", items[1])
	}
	if items[0].Topic != "retrofit" || items[0].AwardID != "best-paper" ||
		items[0].DOI != "10.1234/example.1" || items[0].URL != "https://example.org/paper" {
		t.Errorf("flat publication relationship/link fields = %#v", items[0])
	}
}

func TestBuildPublicationSaveEditsAddsDeletesPreservesUnknownAndSortsLikeWebsite(t *testing.T) {
	path, _ := newPublicationFixture(t)
	snapshot := mustReadPublicationSnapshot(t, path)
	items := publicationItemSummaries(snapshot)
	older := publicationSummaryByTitle(t, items, "Older paper")
	review := publicationSummaryByTitle(t, items, "Review paper")
	press := publicationSummaryByTitle(t, items, "Press paper")
	newest := publicationSummaryByTitle(t, items, "Newest paper")

	edited := publicationInputFromSummary(older)
	edited.TitleEN = "Edited ordinary paper"
	edited.Date = "2026-03"
	edited.AuthorIDs = []string{"park-cs", "jo-sg"}
	edited.Topic = "hvac"
	edited.AwardID = ""
	edited.KeywordsEN = []string{"model predictive control"}
	edited.KeywordsKO = []string{"모델 예측 제어"}
	newItem := PublicationInput{
		TitleEN:    "Alpha on same day",
		TitleKO:    "같은 날 알파",
		AbstractEN: "",
		AbstractKO: "",
		KeywordsEN: []string{},
		KeywordsKO: []string{},
		AuthorIDs:  []string{"jo-sg"},
		Date:       "2026-03",

		UnderReview:     false,
		InPress:         false,
		PublicationType: "international-journal",
		Topic:           "",
		AwardID:         "",
		DOI:             "",
		URL:             "",
		Note:            "",
	}

	// The request is intentionally scrambled. Newest is omitted (deleted).
	encoded, err := buildPublicationSaveLocked(snapshot, []PublicationSaveItem{
		{EditorKey: press.EditorKey},
		{EditorKey: older.EditorKey, Publication: edited},
		{Publication: &newItem},
		{EditorKey: review.EditorKey},
	}, publicationPersonIDs(), publicationAwardIDs(), publicationTopicIDs())
	if err != nil {
		t.Fatalf("buildPublicationSaveLocked() unexpected error: %v", err)
	}
	stored, err := readPublicationsBytes(encoded, publicationPersonIDs(), publicationAwardIDs(), publicationTopicIDs())
	if err != nil {
		t.Fatalf("read built publications: %v", err)
	}
	gotTitles := make([]string, 0, len(stored.Rows))
	for _, row := range stored.Rows {
		gotTitles = append(gotTitles, row.Item.TitleEN)
	}
	wantTitles := []string{"Review paper", "Press paper", "Alpha on same day", "Edited ordinary paper"}
	if !reflect.DeepEqual(gotTitles, wantTitles) {
		t.Errorf("built publication order = %#v, want website order %#v", gotTitles, wantTitles)
	}
	if strings.Contains(string(encoded), newest.TitleEN) {
		t.Error("omitted publication was not deleted")
	}

	editedFields := publicationRawItemByTitle(t, encoded, "Edited ordinary paper")
	assertPublicationRawField(t, editedFields, "future_row_field", map[string]any{
		"flags": []any{true, float64(7)},
		"mode":  "preserve",
	})
	if got := publicationRawStrings(t, editedFields, "author_ids"); !reflect.DeepEqual(got, edited.AuthorIDs) {
		t.Errorf("edited author order = %#v, want %#v", got, edited.AuthorIDs)
	}
	if got := publicationRawString(t, editedFields, "topic"); got != "hvac" {
		t.Errorf("edited topic = %q, want hvac", got)
	}

	newFields := publicationRawItemByTitle(t, encoded, "Alpha on same day")
	if len(newFields) != len(publicationFields)+1 {
		t.Errorf("new publication field count = %d, want exactly %d; fields=%v", len(newFields), len(publicationFields)+1, publicationRawFieldNames(newFields))
	}
	assertPublicationRawField(t, newFields, "visible_in_CV", true)
	if _, exists := newFields["venue_en"]; exists {
		t.Fatal("manual venue was reintroduced")
	}
	if _, exists := newFields["id"]; exists {
		t.Error("new publication unexpectedly received an id")
	}
}

func TestBuildPublicationSaveFullUnchangedPayloadKeepsRawFutureFields(t *testing.T) {
	path, _ := newPublicationFixture(t)
	snapshot := mustReadPublicationSnapshot(t, path)
	items := publicationItemSummaries(snapshot)
	request := make([]PublicationSaveItem, 0, len(items))
	for _, item := range items {
		request = append(request, PublicationSaveItem{
			EditorKey:   item.EditorKey,
			Publication: publicationInputFromSummary(item),
		})
	}
	encoded, err := buildPublicationSaveLocked(snapshot, request, publicationPersonIDs(), publicationAwardIDs(), publicationTopicIDs())
	if err != nil {
		t.Fatalf("unchanged build unexpected error: %v", err)
	}
	fields := publicationRawItemByTitle(t, encoded, "Older paper")
	assertPublicationRawField(t, fields, "future_row_field", map[string]any{
		"flags": []any{true, float64(7)},
		"mode":  "preserve",
	})
	if got := publicationRawString(t, fields, "abstract_en"); got != "  deliberately padded abstract  " {
		t.Errorf("full unchanged payload normalized source text: %q", got)
	}
}

func TestBuildPublicationSavePreservesUntouchedSourceTextWhenEditingRelationship(t *testing.T) {
	path, _ := newPublicationFixture(t)
	snapshot := mustReadPublicationSnapshot(t, path)
	items := publicationItemSummaries(snapshot)
	older := publicationSummaryByTitle(t, items, "Older paper")
	edited := publicationInputFromSummary(older)
	edited.AwardID = ""

	request := make([]PublicationSaveItem, 0, len(items))
	for _, item := range items {
		if item.EditorKey == older.EditorKey {
			request = append(request, PublicationSaveItem{EditorKey: item.EditorKey, Publication: edited})
		} else {
			request = append(request, PublicationSaveItem{EditorKey: item.EditorKey})
		}
	}
	encoded, err := buildPublicationSaveLocked(snapshot, request, publicationPersonIDs(), publicationAwardIDs(), publicationTopicIDs())
	if err != nil {
		t.Fatalf("relationship-only build unexpected error: %v", err)
	}
	fields := publicationRawItemByTitle(t, encoded, "Older paper")
	if got := publicationRawString(t, fields, "abstract_en"); got != "  deliberately padded abstract  " {
		t.Errorf("editing award_id changed untouched abstract_en: %q", got)
	}
	if got := publicationRawString(t, fields, "award_id"); got != "" {
		t.Errorf("edited award_id = %q, want empty", got)
	}
	assertPublicationRawField(t, fields, "future_row_field", map[string]any{
		"flags": []any{true, float64(7)},
		"mode":  "preserve",
	})
}

func TestBuildPublicationSaveUsesStableOrderForCompleteSortTies(t *testing.T) {
	path, _ := newPublicationFixture(t)
	snapshot := mustReadPublicationSnapshot(t, path)
	first := publicationInputFromSummary(publicationSummaryByTitle(t, publicationItemSummaries(snapshot), "Older paper"))
	first.TitleEN = "Same title"
	first.TitleKO = "첫 번째"
	first.Date = "2025-02-01"
	first.AwardID = ""
	second := *first
	second.TitleKO = "두 번째"

	encoded, err := buildPublicationSaveLocked(snapshot, []PublicationSaveItem{
		{Publication: &second},
		{Publication: first},
	}, publicationPersonIDs(), publicationAwardIDs(), publicationTopicIDs())
	if err != nil {
		t.Fatalf("stable-tie build unexpected error: %v", err)
	}
	var rows []struct {
		TitleKO string `json:"title_ko"`
	}
	if err := json.Unmarshal(encoded, &rows); err != nil {
		t.Fatalf("decode stable-tie result: %v", err)
	}
	if got := []string{rows[0].TitleKO, rows[1].TitleKO}; !reflect.DeepEqual(got, []string{"두 번째", "첫 번째"}) {
		t.Errorf("complete sort ties did not retain request order: %#v", got)
	}
}

func TestPublicationValidationRejectsInvalidSchemaAndReferences(t *testing.T) {
	valid := publicationFixtureItem("Valid paper", "2025-01-01", false, false)
	tests := []struct {
		name   string
		mutate func(map[string]any)
	}{
		{name: "persistent id", mutate: func(row map[string]any) { row["id"] = "not-allowed" }},
		{name: "missing field", mutate: func(row map[string]any) { delete(row, "title_en") }},
		{name: "deprecated venue", mutate: func(row map[string]any) { row["venue"] = row["venue_en"] }},
		{name: "null Korean venue", mutate: func(row map[string]any) { row["venue_kr"] = nil }},
		{name: "numeric Korean venue", mutate: func(row map[string]any) { row["venue_kr"] = 7 }},
		{name: "null string", mutate: func(row map[string]any) { row["doi"] = nil }},
		{name: "wrong boolean type", mutate: func(row map[string]any) { row["under_review"] = "false" }},
		{name: "null list", mutate: func(row map[string]any) { row["keywords_en"] = nil }},
		{name: "no localized title", mutate: func(row map[string]any) { row["title_en"], row["title_ko"] = "", "" }},
		{name: "empty keyword", mutate: func(row map[string]any) { row["keywords_en"] = []any{""} }},
		{name: "no authors", mutate: func(row map[string]any) { row["author_ids"] = []any{} }},
		{name: "duplicate author", mutate: func(row map[string]any) { row["author_ids"] = []any{"jo-sg", "jo-sg"} }},
		{name: "unknown author", mutate: func(row map[string]any) { row["author_ids"] = []any{"unknown-person"} }},
		{name: "noncanonical author", mutate: func(row map[string]any) { row["author_ids"] = []any{"Jo SG"} }},
		{name: "unsupported type", mutate: func(row map[string]any) { row["publication_type"] = "book" }},
		{name: "both statuses", mutate: func(row map[string]any) { row["date"], row["under_review"], row["in_press"] = "", true, true }},
		{name: "invalid day", mutate: func(row map[string]any) { row["date"] = "2025-02-29" }},
		{name: "invalid date format", mutate: func(row map[string]any) { row["date"] = "2025-1" }},
		{name: "unknown topic", mutate: func(row map[string]any) { row["topic"] = "missing-topic" }},
		{name: "noncanonical topic", mutate: func(row map[string]any) { row["topic"] = "HVAC" }},
		{name: "unknown award", mutate: func(row map[string]any) { row["award_id"] = "missing-award" }},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			row := clonePublicationFixtureMap(t, valid)
			test.mutate(row)
			raw, err := json.Marshal([]any{row})
			if err != nil {
				t.Fatalf("marshal invalid fixture: %v", err)
			}
			if _, err := readPublicationsBytes(raw, publicationPersonIDs(), publicationAwardIDs(), publicationTopicIDs()); err == nil {
				t.Fatal("readPublicationsBytes() error = nil, want validation rejection")
			}
		})
	}
}

func TestPublicationKeywordsAllowAndPreserveUppercase(t *testing.T) {
	path, _ := newPublicationFixture(t)
	snapshot := mustReadPublicationSnapshot(t, path)
	older := publicationSummaryByTitle(t, publicationItemSummaries(snapshot), "Older paper")
	edited := publicationInputFromSummary(older)
	edited.KeywordsEN = []string{"EnergyPlus", "HVAC Control"}

	encoded, err := buildPublicationSaveLocked(snapshot, []PublicationSaveItem{{EditorKey: older.EditorKey, Publication: edited}}, publicationPersonIDs(), publicationAwardIDs(), publicationTopicIDs())
	if err != nil {
		t.Fatalf("uppercase keywords were rejected while saving: %v", err)
	}
	stored, err := readPublicationsBytes(encoded, publicationPersonIDs(), publicationAwardIDs(), publicationTopicIDs())
	if err != nil {
		t.Fatalf("uppercase keywords were rejected while reading saved data: %v", err)
	}
	got := publicationSummaryByTitle(t, publicationItemSummaries(stored), "Older paper").KeywordsEN
	if !reflect.DeepEqual(got, []string{"EnergyPlus", "HVAC Control"}) {
		t.Fatalf("uppercase keyword casing changed: %#v", got)
	}
}

func TestBuildPublicationSaveRejectsInvalidInputAndOpaqueKeys(t *testing.T) {
	path, _ := newPublicationFixture(t)
	snapshot := mustReadPublicationSnapshot(t, path)
	base := publicationInputFromSummary(publicationSummaryByTitle(t, publicationItemSummaries(snapshot), "Older paper"))
	tests := []struct {
		name   string
		item   PublicationSaveItem
		mutate func(*PublicationInput)
	}{
		{name: "nil new publication", item: PublicationSaveItem{}},
		{name: "unknown editor key", item: PublicationSaveItem{EditorKey: "missing:key"}},
		{name: "nil author ids", mutate: func(input *PublicationInput) { input.AuthorIDs = nil }},
		{name: "nil keywords", mutate: func(input *PublicationInput) { input.KeywordsKO = nil }},
		{name: "duplicate authors", mutate: func(input *PublicationInput) { input.AuthorIDs = []string{"jo-sg", "jo-sg"} }},
		{name: "unknown topic", mutate: func(input *PublicationInput) { input.Topic = "unknown-topic" }},
		{name: "bad date", mutate: func(input *PublicationInput) { input.Date = "2024-13" }},
		{name: "invalid UTF-8 Korean venue", mutate: func(input *PublicationInput) { input.Note = string([]byte{0xff}) }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			item := test.item
			if test.mutate != nil {
				input := *base
				input.KeywordsEN = append([]string(nil), base.KeywordsEN...)
				input.KeywordsKO = append([]string(nil), base.KeywordsKO...)
				input.AuthorIDs = append([]string(nil), base.AuthorIDs...)
				test.mutate(&input)
				item = PublicationSaveItem{Publication: &input}
			}
			if _, err := buildPublicationSaveLocked(snapshot, []PublicationSaveItem{item}, publicationPersonIDs(), publicationAwardIDs(), publicationTopicIDs()); err == nil {
				t.Fatal("buildPublicationSaveLocked() error = nil, want validation rejection")
			}
		})
	}

	existing := publicationSummaryByTitle(t, publicationItemSummaries(snapshot), "Older paper")
	if _, err := buildPublicationSaveLocked(snapshot, []PublicationSaveItem{
		{EditorKey: existing.EditorKey},
		{EditorKey: existing.EditorKey},
	}, publicationPersonIDs(), publicationAwardIDs(), publicationTopicIDs()); err == nil {
		t.Fatal("duplicate existing editor key was accepted")
	}
}

func TestPublicationDateAllowsAllCanonicalPrecisions(t *testing.T) {
	for _, date := range []string{"2025", "2025-02", "2024-02-29"} {
		if err := validatePublicationDate(date); err != nil {
			t.Errorf("validatePublicationDate(%q) unexpected error: %v", date, err)
		}
	}
}

func newPublicationFixture(t *testing.T) (string, []byte) {
	t.Helper()
	rows := []map[string]any{
		publicationFixtureItem("Older paper", "2024-05-06", false, false),
		publicationFixtureItem("Review paper", "", true, false),
		publicationFixtureItem("Press paper", "", false, true),
		publicationFixtureItem("Newest paper", "2027", false, false),
	}
	rows[0]["title_ko"] = "오래된 논문"
	rows[0]["abstract_en"] = "  deliberately padded abstract  "
	rows[0]["keywords_en"] = []string{"building energy", "control"}
	rows[0]["keywords_ko"] = []string{"건물 에너지"}
	rows[0]["author_ids"] = []string{"jo-sg", "park-cs"}
	rows[0]["topic"] = "retrofit"
	rows[0]["award_id"] = "best-paper"
	rows[0]["doi"] = "10.1234/example.1"
	rows[0]["url"] = "https://example.org/paper"
	rows[0]["future_row_field"] = map[string]any{"mode": "preserve", "flags": []any{true, 7}}
	rows[1]["title_ko"] = "심사 중 논문"
	rows[2]["title_ko"] = "출판 예정 논문"
	rows[3]["title_ko"] = "최신 논문"
	raw, err := json.MarshalIndent(rows, "", "  ")
	if err != nil {
		t.Fatalf("marshal publication fixture: %v", err)
	}
	raw = append(raw, '\n')
	path := filepath.Join(t.TempDir(), "data", "publications.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create publication fixture directory: %v", err)
	}
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatalf("write publication fixture: %v", err)
	}
	return path, raw
}

func publicationFixtureItem(title, date string, underReview, inPress bool) map[string]any {
	return map[string]any{
		"title_en":    title,
		"title_ko":    "",
		"abstract_en": "",
		"abstract_ko": "",
		"keywords_en": []string{},
		"keywords_ko": []string{},
		"author_ids":  []string{"jo-sg"},
		"date":        date,

		"under_review":     underReview,
		"in_press":         inPress,
		"publication_type": "international-journal",
		"topic":            "",
		"award_id":         "",
		"doi":              "",
		"url":              "",
		"note":             "",
	}
}

func publicationPersonIDs() map[string]bool {
	return map[string]bool{"jo-sg": true, "park-cs": true}
}

func publicationAwardIDs() map[string]bool {
	return map[string]bool{"best-paper": true}
}

func publicationTopicIDs() map[string]bool {
	return map[string]bool{"retrofit": true, "hvac": true}
}

func mustReadPublicationSnapshot(t *testing.T, path string) publicationSnapshot {
	t.Helper()
	snapshot, err := readPublications(path, publicationPersonIDs(), publicationAwardIDs(), publicationTopicIDs())
	if err != nil {
		t.Fatalf("read publication fixture: %v", err)
	}
	return snapshot
}

func publicationSummaryByTitle(t *testing.T, items []PublicationItemSummary, title string) PublicationItemSummary {
	t.Helper()
	for _, item := range items {
		if item.TitleEN == title {
			return item
		}
	}
	t.Fatalf("publication summary %q not found", title)
	return PublicationItemSummary{}
}

func publicationInputFromSummary(item PublicationItemSummary) *PublicationInput {
	return &PublicationInput{
		TitleEN:    item.TitleEN,
		TitleKO:    item.TitleKO,
		AbstractEN: item.AbstractEN,
		AbstractKO: item.AbstractKO,
		KeywordsEN: clonePublicationStrings(item.KeywordsEN),
		KeywordsKO: clonePublicationStrings(item.KeywordsKO),
		AuthorIDs:  clonePublicationStrings(item.AuthorIDs),
		Date:       item.Date,

		JournalID:       item.JournalID,
		ConferenceID:    item.ConferenceID,
		Volume:          item.Volume,
		Issue:           item.Issue,
		Pages:           item.Pages,
		ELocator:        item.ELocator,
		OnlineDate:      item.OnlineDate,
		IssueDate:       item.IssueDate,
		UnderReview:     item.UnderReview,
		InPress:         item.InPress,
		PublicationType: item.PublicationType,
		Topic:           item.Topic,
		AwardID:         item.AwardID,
		DOI:             item.DOI,
		URL:             item.URL,
		Note:            item.Note,
	}
}

func publicationRawItemByTitle(t *testing.T, raw []byte, title string) map[string]json.RawMessage {
	t.Helper()
	var rows []map[string]json.RawMessage
	if err := json.Unmarshal(raw, &rows); err != nil {
		t.Fatalf("decode raw publications: %v", err)
	}
	for _, row := range rows {
		if publicationRawString(t, row, "title_en") == title {
			return row
		}
	}
	t.Fatalf("raw publication %q not found", title)
	return nil
}

func publicationRawString(t *testing.T, fields map[string]json.RawMessage, name string) string {
	t.Helper()
	var value string
	if err := json.Unmarshal(fields[name], &value); err != nil {
		t.Fatalf("decode publication field %q: %v", name, err)
	}
	return value
}

func publicationRawStrings(t *testing.T, fields map[string]json.RawMessage, name string) []string {
	t.Helper()
	var value []string
	if err := json.Unmarshal(fields[name], &value); err != nil {
		t.Fatalf("decode publication field %q: %v", name, err)
	}
	return value
}

func assertPublicationRawField(t *testing.T, fields map[string]json.RawMessage, name string, want any) {
	t.Helper()
	raw, exists := fields[name]
	if !exists {
		t.Fatalf("unknown publication JSON field %q was dropped", name)
	}
	var got any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("decode publication JSON field %q: %v", name, err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("publication JSON field %q = %#v, want %#v", name, got, want)
	}
}

func publicationRawFieldNames(fields map[string]json.RawMessage) []string {
	names := make([]string, 0, len(fields))
	for name := range fields {
		names = append(names, name)
	}
	return names
}

func clonePublicationFixtureMap(t *testing.T, source map[string]any) map[string]any {
	t.Helper()
	raw, err := json.Marshal(source)
	if err != nil {
		t.Fatalf("marshal fixture clone: %v", err)
	}
	var clone map[string]any
	if err := json.Unmarshal(raw, &clone); err != nil {
		t.Fatalf("unmarshal fixture clone: %v", err)
	}
	return clone
}
