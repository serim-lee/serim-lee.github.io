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

func TestReadPeopleAllowsIndependentEmptyAndUnequalNoteLists(t *testing.T) {
	snapshot, err := readPeopleBytes(fixturePeopleJSON())
	if err != nil {
		t.Fatalf("readPeopleBytes() unexpected error: %v", err)
	}
	items := peopleItemSummaries(snapshot)
	if len(items) != 3 {
		t.Fatalf("people summaries length = %d, want 3", len(items))
	}
	if !items[0].IsSelf || items[0].ID != "jo-hg" || items[0].NotesEN == nil || items[0].NotesKO == nil {
		t.Fatalf("self summary = %#v", items[0])
	}
	if len(items[0].NotesEN) != 0 || len(items[0].NotesKO) != 0 {
		t.Fatalf("empty note arrays changed: %#v", items[0])
	}
	if !reflect.DeepEqual(items[1].NotesEN, []string{"Lab", "Kyung Hee University"}) ||
		!reflect.DeepEqual(items[1].NotesKO, []string{"경희대학교"}) {
		t.Fatalf("independent note arrays changed: %#v", items[1])
	}
	items[1].NotesEN[0] = "mutated response"
	if snapshot.Rows[1].Item.NotesEN[0] != "Lab" {
		t.Fatal("people summary leaked the snapshot's note slice")
	}
	if snapshot.Rows[1].Key == "" || !strings.HasPrefix(snapshot.Rows[1].Key, "1:") {
		t.Fatalf("opaque person editor key = %q", snapshot.Rows[1].Key)
	}
}

func TestReadAwardsValidatesAndReturnsFlatSummaries(t *testing.T) {
	snapshot, err := readAwardsBytes(fixtureAwardsJSON())
	if err != nil {
		t.Fatalf("readAwardsBytes() unexpected error: %v", err)
	}
	items := awardItemSummaries(snapshot)
	if len(items) != 2 {
		t.Fatalf("award summaries length = %d, want 2", len(items))
	}
	if got := items[0]; got.ID != "best-paper" || got.Date != "2025-12-02" ||
		got.TitleKO != "우수논문상" || got.OrganizationKO != "대한건축학회" || got.EditorKey == "" {
		t.Fatalf("first award summary = %#v", got)
	}
}

func TestReadPeopleRejectsInvalidSchemaIDsAndSelfCount(t *testing.T) {
	valid := fixturePeopleJSON()
	tests := []struct {
		name string
		raw  []byte
	}{
		{name: "top-level null", raw: []byte(`null`)},
		{name: "invalid UTF-8", raw: append([]byte(`[{"id":"`), 0xff)},
		{name: "trailing JSON", raw: append(append([]byte(nil), valid...), []byte(` {}`)...)},
		{name: "missing field", raw: bytes.Replace(valid, []byte(`"name_ko": "조형곤",`), nil, 1)},
		{name: "null string", raw: bytes.Replace(valid, []byte(`"name_ko": "조형곤"`), []byte(`"name_ko": null`), 1)},
		{name: "string bool", raw: bytes.Replace(valid, []byte(`"is_self": true`), []byte(`"is_self": "true"`), 1)},
		{name: "null bool", raw: bytes.Replace(valid, []byte(`"is_self": true`), []byte(`"is_self": null`), 1)},
		{name: "null notes", raw: bytes.Replace(valid, []byte(`"notes_en": []`), []byte(`"notes_en": null`), 1)},
		{name: "empty note", raw: bytes.Replace(valid, []byte(`"notes_ko": ["경희대학교"]`), []byte(`"notes_ko": [""]`), 1)},
		{name: "noncanonical ID", raw: bytes.Replace(valid, []byte(`"id": "cho-sk"`), []byte(`"id": "Cho_SK"`), 1)},
		{name: "duplicate ID", raw: bytes.Replace(valid, []byte(`"id": "choi-sh"`), []byte(`"id": "cho-sk"`), 1)},
		{name: "no self", raw: bytes.Replace(valid, []byte(`"is_self": true`), []byte(`"is_self": false`), 1)},
		{name: "multiple self", raw: bytes.Replace(valid, []byte(`"is_self": false`), []byte(`"is_self": true`), 1)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := readPeopleBytes(test.raw); err == nil {
				t.Fatal("readPeopleBytes() error = nil, want strict validation rejection")
			}
		})
	}
}

func TestReadAwardsRejectsInvalidSchemaIDsDatesAndText(t *testing.T) {
	valid := fixtureAwardsJSON()
	tests := []struct {
		name string
		raw  []byte
	}{
		{name: "top-level null", raw: []byte(`null`)},
		{name: "invalid UTF-8", raw: append([]byte(`[{"id":"`), 0xff)},
		{name: "missing field", raw: bytes.Replace(valid, []byte(`"organization_ko": "대한건축학회"`), nil, 1)},
		{name: "null title", raw: bytes.Replace(valid, []byte(`"title_en": "Best Paper Award"`), []byte(`"title_en": null`), 1)},
		{name: "empty organization", raw: bytes.Replace(
			bytes.Replace(valid, []byte(`"organization_en": "Architectural Institute of Korea"`), []byte(`"organization_en": ""`), 1),
			[]byte(`"organization_ko": "대한건축학회"`), []byte(`"organization_ko": ""`), 1,
		)},
		{name: "impossible date", raw: bytes.Replace(valid, []byte(`"date": "2025-12-02"`), []byte(`"date": "2025-02-29"`), 1)},
		{name: "noncanonical date", raw: bytes.Replace(valid, []byte(`"date": "2025-12-02"`), []byte(`"date": "2025-12-2"`), 1)},
		{name: "noncanonical ID", raw: bytes.Replace(valid, []byte(`"id": "best-paper"`), []byte(`"id": "best_paper"`), 1)},
		{name: "duplicate ID", raw: bytes.Replace(valid, []byte(`"id": "young-scholar"`), []byte(`"id": "best-paper"`), 1)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := readAwardsBytes(test.raw); err == nil {
				t.Fatal("readAwardsBytes() error = nil, want strict validation rejection")
			}
		})
	}
}

func TestBuildPeopleSaveUpdatesAddsDeletesReordersAndPreservesUnknownFields(t *testing.T) {
	current, err := readPeopleBytes(fixturePeopleJSON())
	if err != nil {
		t.Fatalf("readPeopleBytes() unexpected error: %v", err)
	}
	app := &App{}
	cho := peopleSummaryByID(t, peopleItemSummaries(current), "cho-sk")
	self := peopleSummaryByID(t, peopleItemSummaries(current), "jo-hg")
	edited := personInputFromSummary(cho)
	edited.ID = cho.ID
	edited.NameKO = "조성권 수정"
	edited.NotesEN = []string{"Kyung Hee University"}
	edited.NotesKO = []string{"경희대학교", "주거환경학과"}
	newPerson := PersonInput{
		NameEN: "Cho SK", NameKO: "동명이인", IsSelf: false,
		NotesEN: []string{}, NotesKO: []string{"다른 소속"},
	}

	result, err := app.buildPeopleSaveLocked(current, []PersonSaveItem{
		{EditorKey: cho.EditorKey, Person: edited},
		{Person: &newPerson},
		{EditorKey: self.EditorKey},
	})
	if err != nil {
		t.Fatalf("buildPeopleSaveLocked() unexpected error: %v", err)
	}
	stored, err := readPeopleBytes(result)
	if err != nil {
		t.Fatalf("read built people: %v\n%s", err, result)
	}
	gotIDs := []string{stored.Rows[0].Item.ID, stored.Rows[1].Item.ID, stored.Rows[2].Item.ID}
	if want := []string{"cho-sk", "cho-sk-2", "jo-hg"}; !reflect.DeepEqual(gotIDs, want) {
		t.Fatalf("people IDs/order = %#v, want %#v", gotIDs, want)
	}
	if stored.Rows[0].Item.NameKO != "조성권 수정" ||
		len(stored.Rows[0].Item.NotesEN) != 1 || len(stored.Rows[0].Item.NotesKO) != 2 {
		t.Fatalf("edited person = %#v", stored.Rows[0].Item)
	}
	if strings.Contains(string(result), `"id": "choi-sh"`) {
		t.Fatal("omitted person was not deleted")
	}
	rawCho := peopleRawItemByID(t, result, "cho-sk")
	assertRawJSONField(t, rawCho, "future_person_field", map[string]any{
		"keep": true,
		"rank": float64(7),
	})
	if !bytes.Contains(result, []byte(`"future_self_field"`)) {
		t.Fatal("key-only person's unknown field was lost")
	}
}

func TestBuildPeopleSaveSupportsProposedNewIDAndRejectsIDMutationOrCollision(t *testing.T) {
	current, err := readPeopleBytes(fixturePeopleJSON())
	if err != nil {
		t.Fatal(err)
	}
	app := &App{}
	all := peopleItemSummaries(current)
	request := make([]PersonSaveItem, 0, len(all)+1)
	for _, item := range all {
		request = append(request, PersonSaveItem{EditorKey: item.EditorKey})
	}
	proposed := PersonInput{
		ID: "new-researcher", NameEN: "New Researcher", NameKO: "새 연구자",
		NotesEN: []string{}, NotesKO: []string{},
	}
	result, err := app.buildPeopleSaveLocked(current, append(request, PersonSaveItem{Person: &proposed}))
	if err != nil {
		t.Fatalf("proposed canonical ID rejected: %v", err)
	}
	if !bytes.Contains(result, []byte(`"id": "new-researcher"`)) {
		t.Fatalf("proposed person ID not retained:\n%s", result)
	}

	cho := peopleSummaryByID(t, all, "cho-sk")
	mutated := personInputFromSummary(cho)
	mutated.ID = "renamed-person"
	if _, err := app.buildPeopleSaveLocked(current, []PersonSaveItem{
		{EditorKey: all[0].EditorKey},
		{EditorKey: cho.EditorKey, Person: mutated},
	}); err == nil {
		t.Fatal("existing person ID mutation was accepted")
	}

	for _, id := range []string{"cho-sk", "Bad_ID"} {
		candidate := proposed
		candidate.ID = id
		if _, err := app.buildPeopleSaveLocked(current, append(request, PersonSaveItem{Person: &candidate})); err == nil {
			t.Fatalf("new proposed person ID %q was accepted", id)
		}
	}
}

func TestBuildPeopleSavePreservesWholeFileBytesAndUnequalListsWhenUnchanged(t *testing.T) {
	raw := fixturePeopleJSON()
	current, err := readPeopleBytes(raw)
	if err != nil {
		t.Fatal(err)
	}
	request := make([]PersonSaveItem, 0, len(current.Rows))
	for _, item := range peopleItemSummaries(current) {
		request = append(request, PersonSaveItem{EditorKey: item.EditorKey, Person: personInputFromSummary(item)})
	}
	result, err := (&App{}).buildPeopleSaveLocked(current, request)
	if err != nil {
		t.Fatalf("unchanged people build: %v", err)
	}
	if !bytes.Equal(result, raw) {
		t.Fatal("unchanged full people payload did not preserve exact source bytes")
	}
}

func TestBuildPeopleSaveRejectsInvalidRequestsAndSelfCounts(t *testing.T) {
	current, err := readPeopleBytes(fixturePeopleJSON())
	if err != nil {
		t.Fatal(err)
	}
	app := &App{}
	items := peopleItemSummaries(current)
	self := peopleSummaryByID(t, items, "jo-hg")
	other := peopleSummaryByID(t, items, "cho-sk")
	otherInput := personInputFromSummary(other)
	otherInput.IsSelf = true

	tests := []struct {
		name    string
		request []PersonSaveItem
	}{
		{name: "unknown key", request: []PersonSaveItem{{EditorKey: "missing"}}},
		{name: "duplicate key", request: []PersonSaveItem{{EditorKey: self.EditorKey}, {EditorKey: self.EditorKey}}},
		{name: "new without input", request: []PersonSaveItem{{EditorKey: self.EditorKey}, {}}},
		{name: "no self", request: []PersonSaveItem{{EditorKey: other.EditorKey}}},
		{name: "multiple self", request: []PersonSaveItem{{EditorKey: self.EditorKey}, {EditorKey: other.EditorKey, Person: otherInput}}},
	}
	badNilNotes := personInputFromSummary(other)
	badNilNotes.NotesEN = nil
	tests = append(tests, struct {
		name    string
		request []PersonSaveItem
	}{name: "nil note array", request: []PersonSaveItem{{EditorKey: self.EditorKey}, {EditorKey: other.EditorKey, Person: badNilNotes}}})
	badNote := personInputFromSummary(other)
	badNote.NotesKO = []string{"   "}
	tests = append(tests, struct {
		name    string
		request []PersonSaveItem
	}{name: "blank note", request: []PersonSaveItem{{EditorKey: self.EditorKey}, {EditorKey: other.EditorKey, Person: badNote}}})

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := app.buildPeopleSaveLocked(current, test.request); err == nil {
				t.Fatal("buildPeopleSaveLocked() error = nil, want rejection")
			}
		})
	}
}

func TestBuildAwardsSaveUpdatesAddsDeletesReordersAndPreservesUnknownFields(t *testing.T) {
	current, err := readAwardsBytes(fixtureAwardsJSON())
	if err != nil {
		t.Fatal(err)
	}
	items := awardItemSummaries(current)
	best := awardSummaryByID(t, items, "best-paper")
	edited := awardInputFromSummary(best)
	edited.ID = best.ID
	edited.TitleKO = "최우수논문상"
	newAward := AwardInput{
		Date: "2026-08-24", TitleEN: "Best Paper", TitleKO: "우수논문상",
		OrganizationEN: "New Society", OrganizationKO: "새 학회",
	}
	result, err := (&App{}).buildAwardsSaveLocked(current, []AwardSaveItem{
		{Award: &newAward},
		{EditorKey: best.EditorKey, Award: edited},
	})
	if err != nil {
		t.Fatalf("buildAwardsSaveLocked() unexpected error: %v", err)
	}
	stored, err := readAwardsBytes(result)
	if err != nil {
		t.Fatalf("read built awards: %v\n%s", err, result)
	}
	if got, want := []string{stored.Rows[0].Item.ID, stored.Rows[1].Item.ID}, []string{"best-paper-2", "best-paper"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("award IDs/order = %#v, want %#v", got, want)
	}
	if stored.Rows[1].Item.TitleKO != "최우수논문상" {
		t.Fatalf("edited award = %#v", stored.Rows[1].Item)
	}
	if strings.Contains(string(result), `"id": "young-scholar"`) {
		t.Fatal("omitted award was not deleted")
	}
	rawBest := awardRawItemByID(t, result, "best-paper")
	assertRawJSONField(t, rawBest, "future_award_field", map[string]any{"keep": "yes"})
}

func TestBuildAwardsSaveSupportsProposedIDAndPreservesUnchangedBytes(t *testing.T) {
	raw := fixtureAwardsJSON()
	current, err := readAwardsBytes(raw)
	if err != nil {
		t.Fatal(err)
	}
	items := awardItemSummaries(current)
	unchanged := make([]AwardSaveItem, 0, len(items))
	for _, item := range items {
		unchanged = append(unchanged, AwardSaveItem{EditorKey: item.EditorKey, Award: awardInputFromSummary(item)})
	}
	result, err := (&App{}).buildAwardsSaveLocked(current, unchanged)
	if err != nil {
		t.Fatalf("unchanged awards build: %v", err)
	}
	if !bytes.Equal(result, raw) {
		t.Fatal("unchanged full awards payload did not preserve exact source bytes")
	}

	proposed := AwardInput{
		ID: "explicit-award", Date: "2026-08-24", TitleEN: "Explicit Award", TitleKO: "명시적 수상",
		OrganizationEN: "Society", OrganizationKO: "학회",
	}
	result, err = (&App{}).buildAwardsSaveLocked(current, append(unchanged, AwardSaveItem{Award: &proposed}))
	if err != nil {
		t.Fatalf("proposed canonical award ID rejected: %v", err)
	}
	if !bytes.Contains(result, []byte(`"id": "explicit-award"`)) {
		t.Fatalf("proposed award ID not retained:\n%s", result)
	}
}

func TestBuildAwardsSaveRejectsMutatedCollidingAndInvalidInputs(t *testing.T) {
	current, err := readAwardsBytes(fixtureAwardsJSON())
	if err != nil {
		t.Fatal(err)
	}
	items := awardItemSummaries(current)
	best := awardSummaryByID(t, items, "best-paper")
	edited := awardInputFromSummary(best)

	tests := []struct {
		name    string
		request []AwardSaveItem
	}{
		{name: "unknown key", request: []AwardSaveItem{{EditorKey: "missing"}}},
		{name: "duplicate key", request: []AwardSaveItem{{EditorKey: best.EditorKey}, {EditorKey: best.EditorKey}}},
		{name: "new without input", request: []AwardSaveItem{{}}},
	}
	mutatedID := *edited
	mutatedID.ID = "renamed-award"
	tests = append(tests, struct {
		name    string
		request []AwardSaveItem
	}{name: "mutated existing ID", request: []AwardSaveItem{{EditorKey: best.EditorKey, Award: &mutatedID}}})
	badDate := *edited
	badDate.Date = "2025-02-29"
	tests = append(tests, struct {
		name    string
		request []AwardSaveItem
	}{name: "invalid date", request: []AwardSaveItem{{EditorKey: best.EditorKey, Award: &badDate}}})
	emptyTitle := *edited
	emptyTitle.TitleEN = " "
	emptyTitle.TitleKO = " "
	tests = append(tests, struct {
		name    string
		request []AwardSaveItem
	}{name: "blank text", request: []AwardSaveItem{{EditorKey: best.EditorKey, Award: &emptyTitle}}})
	for _, id := range []string{"best-paper", "Bad_ID"} {
		candidate := AwardInput{
			ID: id, Date: "2026-08-24", TitleEN: "Candidate", TitleKO: "후보",
			OrganizationEN: "Society", OrganizationKO: "학회",
		}
		tests = append(tests, struct {
			name    string
			request []AwardSaveItem
		}{name: "invalid proposed ID " + id, request: []AwardSaveItem{{Award: &candidate}}})
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := (&App{}).buildAwardsSaveLocked(current, test.request); err == nil {
				t.Fatal("buildAwardsSaveLocked() error = nil, want rejection")
			}
		})
	}
}

func TestReadPeopleAndAwardsFromFiles(t *testing.T) {
	root := t.TempDir()
	peoplePath := filepath.Join(root, "people.json")
	awardsPath := filepath.Join(root, "awards.json")
	if err := os.WriteFile(peoplePath, fixturePeopleJSON(), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(awardsPath, fixtureAwardsJSON(), 0o644); err != nil {
		t.Fatal(err)
	}
	if snapshot, err := readPeople(peoplePath); err != nil || len(snapshot.Rows) != 3 {
		t.Fatalf("readPeople() = %d rows, %v", len(snapshot.Rows), err)
	}
	if snapshot, err := readAwards(awardsPath); err != nil || len(snapshot.Rows) != 2 {
		t.Fatalf("readAwards() = %d rows, %v", len(snapshot.Rows), err)
	}
}

func fixturePeopleJSON() []byte {
	return []byte(`[
  {
    "id": "jo-hg",
    "name_en": "Jo, H. G.",
    "name_ko": "조형곤",
    "is_self": true,
    "notes_en": [],
    "notes_ko": [],
    "future_self_field": {"raw": [1, 2, 3]}
  },
  {
    "id": "cho-sk",
    "name_en": "Cho, S. K.",
    "name_ko": "조성권",
    "is_self": false,
    "notes_en": ["Lab", "Kyung Hee University"],
    "notes_ko": ["경희대학교"],
    "future_person_field": {"keep": true, "rank": 7}
  },
  {
    "id": "choi-sh",
    "name_en": "Choi, S. H.",
    "name_ko": "최서희",
    "is_self": false,
    "notes_en": ["LG Electronics"],
    "notes_ko": ["LG전자"]
  }
]
`)
}

func fixtureAwardsJSON() []byte {
	return []byte(`[
  {
    "id": "best-paper",
    "date": "2025-12-02",
    "title_en": "Best Paper Award",
    "title_ko": "우수논문상",
    "organization_en": "Architectural Institute of Korea",
    "organization_ko": "대한건축학회",
    "future_award_field": {"keep": "yes"}
  },
  {
    "id": "young-scholar",
    "date": "2024-10-25",
    "title_en": "Young Scholar Award",
    "title_ko": "신진학자상",
    "organization_en": "Example Society",
    "organization_ko": "예시 학회"
  }
]
`)
}

func peopleSummaryByID(t *testing.T, items []PeopleItemSummary, id string) PeopleItemSummary {
	t.Helper()
	for _, item := range items {
		if item.ID == id {
			return item
		}
	}
	t.Fatalf("person summary %q not found", id)
	return PeopleItemSummary{}
}

func awardSummaryByID(t *testing.T, items []AwardItemSummary, id string) AwardItemSummary {
	t.Helper()
	for _, item := range items {
		if item.ID == id {
			return item
		}
	}
	t.Fatalf("award summary %q not found", id)
	return AwardItemSummary{}
}

func personInputFromSummary(item PeopleItemSummary) *PersonInput {
	return &PersonInput{
		NameEN: item.NameEN, NameKO: item.NameKO, IsSelf: item.IsSelf,
		NotesEN: append([]string{}, item.NotesEN...), NotesKO: append([]string{}, item.NotesKO...),
	}
}

func awardInputFromSummary(item AwardItemSummary) *AwardInput {
	return &AwardInput{
		Date: item.Date, TitleEN: item.TitleEN, TitleKO: item.TitleKO,
		OrganizationEN: item.OrganizationEN, OrganizationKO: item.OrganizationKO,
	}
}

func peopleRawItemByID(t *testing.T, raw []byte, id string) map[string]json.RawMessage {
	t.Helper()
	return peopleAwardsRawItemByID(t, raw, id, "person")
}

func awardRawItemByID(t *testing.T, raw []byte, id string) map[string]json.RawMessage {
	t.Helper()
	return peopleAwardsRawItemByID(t, raw, id, "award")
}

func peopleAwardsRawItemByID(t *testing.T, raw []byte, id, label string) map[string]json.RawMessage {
	t.Helper()
	var rows []map[string]json.RawMessage
	if err := json.Unmarshal(raw, &rows); err != nil {
		t.Fatalf("decode raw %s rows: %v", label, err)
	}
	for _, row := range rows {
		var rowID string
		if err := json.Unmarshal(row["id"], &rowID); err == nil && rowID == id {
			return row
		}
	}
	t.Fatalf("raw %s %q not found", label, id)
	return nil
}
