package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestReadSoftwareReturnsFullItemsInArrayOrderWithoutWriting(t *testing.T) {
	root := newSoftwareRepoFixture(t)
	softwarePath := filepath.Join(root, "data", "software.json")
	before := snapshotBoardRepo(t, root)
	raw := readFixtureFile(t, softwarePath)

	snapshot, err := readSoftware(softwarePath, root)
	if err != nil {
		t.Fatalf("readSoftware() unexpected error: %v", err)
	}
	if snapshot.Raw == nil || revisionOf(snapshot.Raw) != revisionOf(raw) {
		t.Error("readSoftware() snapshot revision does not match source bytes")
	}
	items := softwareItemSummaries(snapshot)
	wantIDs := []string{"development-tool", "stable-tool", "preview-tool"}
	if len(items) != len(wantIDs) {
		t.Fatalf("softwareItemSummaries() returned %d items, want %d", len(items), len(wantIDs))
	}
	for index, wantID := range wantIDs {
		if items[index].ID != wantID {
			t.Errorf("software summary[%d] ID = %q, want source array order ID %q", index, items[index].ID, wantID)
		}
		if items[index].EditorKey == "" {
			t.Errorf("software summary[%d] has an empty editor key", index)
		}
	}
	stable := softwareSummaryByID(t, items, "stable-tool")
	if stable.Name != "Stable Tool" || stable.Stage != "release" {
		t.Errorf("stable software summary = %#v", stable)
	}
	if !reflect.DeepEqual(stable.NotesEN, []string{"Stable English note"}) ||
		!reflect.DeepEqual(stable.NotesKR, []string{"안정판 국문 설명"}) ||
		!reflect.DeepEqual(stable.Technologies, []string{"Go", "Wails"}) {
		t.Errorf("stable software detail arrays = notes %#v/%#v, technologies %#v", stable.NotesEN, stable.NotesKR, stable.Technologies)
	}
	if len(stable.Links) != 2 || stable.Links[0].EditorKey == "" || stable.Links[1].EditorKey == "" {
		t.Fatalf("software links were not exposed with opaque keys: %#v", stable.Links)
	}
	if stable.Links[0].URL != "https://github.com/example/stable-tool" || stable.Links[1].LabelEN != "Documentation" {
		t.Errorf("software link summaries = %#v", stable.Links)
	}
	if len(stable.Media) != 1 || stable.Media[0].EditorKey == "" || stable.Media[0].PreviewURL == "" {
		t.Fatalf("software media was not exposed with key and preview: %#v", stable.Media)
	}
	assertBoardRepoSnapshot(t, root, before, "readSoftware")
}

func TestBuildSoftwareSavePreservesRequestOrderAndUnchangedRawRows(t *testing.T) {
	root := newSoftwareRepoFixture(t)
	snapshot, err := readSoftware(filepath.Join(root, "data", "software.json"), root)
	if err != nil {
		t.Fatalf("readSoftware() unexpected error: %v", err)
	}
	items := softwareItemSummaries(snapshot)
	development := softwareSummaryByID(t, items, "development-tool")
	stable := softwareSummaryByID(t, items, "stable-tool")
	preview := softwareSummaryByID(t, items, "preview-tool")

	request := []SoftwareSaveItem{
		{EditorKey: preview.EditorKey},
		{EditorKey: stable.EditorKey, Software: softwareInputFromSummary(stable)},
		{EditorKey: development.EditorKey},
	}
	app := &App{repoRoot: root}
	encoded, pending, tokens, err := app.buildSoftwareSaveLocked(root, snapshot, request, nil)
	if err != nil {
		t.Fatalf("buildSoftwareSaveLocked() unexpected error: %v", err)
	}
	if len(pending) != 0 || len(tokens) != 0 {
		t.Fatalf("unchanged software build planned media writes: pending=%#v tokens=%#v", pending, tokens)
	}
	stored, err := readSoftwareBytes(encoded, root)
	if err != nil {
		t.Fatalf("readSoftwareBytes() built result: %v", err)
	}
	gotIDs := []string{stored.Rows[0].Item.ID, stored.Rows[1].Item.ID, stored.Rows[2].Item.ID}
	wantIDs := []string{"preview-tool", "stable-tool", "development-tool"}
	if !reflect.DeepEqual(gotIDs, wantIDs) {
		t.Errorf("built software order = %#v, want request order %#v", gotIDs, wantIDs)
	}
	rawStable := softwareRawItemByID(t, encoded, "stable-tool")
	assertRawJSONField(t, rawStable, "future_row_field", map[string]any{
		"mode": "lossless",
		"rank": float64(4),
	})
	var rawLinks []json.RawMessage
	if err := json.Unmarshal(rawStable["links"], &rawLinks); err != nil {
		t.Fatalf("decode unchanged links: %v", err)
	}
	if len(rawLinks) != 2 || len(rawLinks[0]) == 0 || rawLinks[0][0] != '"' {
		t.Fatalf("unchanged URL-string link changed representation: %s", rawStable["links"])
	}
}

func TestBuildSoftwareSaveEditsAddsDeletesAndPreservesUnknownFields(t *testing.T) {
	root := newSoftwareRepoFixture(t)
	mediaDirectory := filepath.Join(root, "data", "media")
	collisionPath := filepath.Join(mediaDirectory, "software-stable-tool-02.png")
	collisionContents := []byte("do not overwrite this existing path")
	writeFixtureFile(t, collisionPath, collisionContents)
	softwarePath := filepath.Join(root, "data", "software.json")
	snapshot, err := readSoftware(softwarePath, root)
	if err != nil {
		t.Fatalf("readSoftware() unexpected error: %v", err)
	}
	items := softwareItemSummaries(snapshot)
	stable := softwareSummaryByID(t, items, "stable-tool")
	preview := softwareSummaryByID(t, items, "preview-tool")

	app := &App{repoRoot: root}
	t.Cleanup(func() { app.shutdown(context.Background()) })
	inputPath := filepath.Join(t.TempDir(), "new-software-media.bin")
	pngContents := fixturePNGBytes()
	writeFixtureFile(t, inputPath, pngContents)
	staged, err := app.StageBoardMedia([]string{inputPath})
	if err != nil || len(staged.Items) != 1 {
		t.Fatalf("StageBoardMedia() = %#v, %v; want one staged image", staged, err)
	}
	beforeBuild := snapshotBoardRepo(t, root)

	edited := SoftwareInput{
		Name:  "Stable Tool Updated",
		Stage: "development",
		Links: []SoftwareLinkInput{
			softwareLinkInputFromSummary(stable.Links[0]),
			{
				EditorKey: stable.Links[1].EditorKey,
				URL:       "https://example.org/new-docs",
				LabelEN:   "Updated Documentation",
				LabelKO:   "수정된 문서",
			},
			{URL: "https://pypi.org/project/stable-tool"},
		},
		NotesEN: []string{"Updated English note", "Second English note"},
		NotesKR: []string{"수정된 국문 설명", "두 번째 국문 설명"},
		Media: []SoftwareMediaInput{
			{
				EditorKey: stable.Media[0].EditorKey,
				CaptionEN: "Updated existing screenshot",
				CaptionKO: "수정된 기존 화면",
			},
			{
				StageToken: staged.Items[0].StageToken,
				CaptionEN:  "New screenshot",
				CaptionKO:  "새 화면",
			},
		},
		Technologies: []string{"Go", "Wails", "SQLite"},
	}
	newItem := SoftwareInput{
		Name:         "Stable Tool",
		Stage:        "preview",
		Links:        []SoftwareLinkInput{},
		NotesEN:      []string{"New collision-safe software"},
		NotesKR:      []string{"ID 충돌을 피하는 신규 소프트웨어"},
		Media:        []SoftwareMediaInput{},
		Technologies: []string{"Rust"},
	}
	request := []SoftwareSaveItem{
		{EditorKey: preview.EditorKey},
		{EditorKey: stable.EditorKey, Software: &edited},
		{Software: &newItem},
	}

	encoded, pending, tokens, err := app.buildSoftwareSaveLocked(root, snapshot, request, nil)
	if err != nil {
		t.Fatalf("buildSoftwareSaveLocked() unexpected error: %v", err)
	}
	assertBoardRepoSnapshot(t, root, beforeBuild, "buildSoftwareSaveLocked before commit")
	if len(pending) != 1 || len(tokens) != 1 || tokens[0] != staged.Items[0].StageToken {
		t.Fatalf("software pending media = %#v tokens=%#v, want one staged file", pending, tokens)
	}
	wantMediaName := "software-stable-tool-02-2.png"
	if filepath.Base(pending[0].DestinationPath) != wantMediaName {
		t.Errorf("pending media destination = %q, want %q", filepath.Base(pending[0].DestinationPath), wantMediaName)
	}

	created, err := publishBoardMedia(pending)
	if err != nil {
		t.Fatalf("publish software media: %v", err)
	}
	if err := writeFileAtomically(softwarePath, encoded); err != nil {
		removeFiles(created)
		t.Fatalf("write built software fixture: %v", err)
	}
	if err := app.DiscardBoardMedia(tokens); err != nil {
		t.Fatalf("discard committed staging files: %v", err)
	}

	stored, err := readSoftware(softwarePath, root)
	if err != nil {
		t.Fatalf("readSoftware() after commit: %v", err)
	}
	gotIDs := []string{stored.Rows[0].Item.ID, stored.Rows[1].Item.ID, stored.Rows[2].Item.ID}
	wantIDs := []string{"preview-tool", "stable-tool", "stable-tool-2"}
	if !reflect.DeepEqual(gotIDs, wantIDs) {
		t.Errorf("stored software IDs/order = %#v, want %#v", gotIDs, wantIDs)
	}
	if strings.Contains(string(encoded), "development-tool") {
		t.Error("omitted software item was not deleted")
	}
	storedStable := stored.Rows[1].Item
	if storedStable.ID != "stable-tool" || storedStable.Name != "Stable Tool Updated" || storedStable.Stage != "development" {
		t.Errorf("existing ID or edited scalar fields changed incorrectly: %#v", storedStable)
	}
	if !reflect.DeepEqual(storedStable.NotesEN, edited.NotesEN) || !reflect.DeepEqual(storedStable.Technologies, edited.Technologies) {
		t.Errorf("edited lists were not saved: %#v", storedStable)
	}
	if len(storedStable.Links) != 3 || storedStable.Links[0].isObject ||
		storedStable.Links[0].URL != "https://github.com/example/stable-tool" {
		t.Errorf("edited links = %#v", storedStable.Links)
	}
	if len(storedStable.Media) != 2 || storedStable.Media[1].Src != wantMediaName {
		t.Fatalf("edited media = %#v, want retained plus staged", storedStable.Media)
	}
	if got := readFixtureFile(t, filepath.Join(mediaDirectory, wantMediaName)); !bytes.Equal(got, pngContents) {
		t.Fatal("published software media bytes differ from staged image")
	}
	if got := readFixtureFile(t, collisionPath); !bytes.Equal(got, collisionContents) {
		t.Fatal("software media collision target was overwritten")
	}

	rawStable := softwareRawItemByID(t, readFixtureFile(t, softwarePath), "stable-tool")
	assertRawJSONField(t, rawStable, "future_row_field", map[string]any{
		"mode": "lossless",
		"rank": float64(4),
	})
	var rawLinks []map[string]json.RawMessage
	var heterogeneousLinks []json.RawMessage
	if err := json.Unmarshal(rawStable["links"], &heterogeneousLinks); err != nil {
		t.Fatalf("decode heterogeneous links: %v", err)
	}
	if len(heterogeneousLinks) != 3 || heterogeneousLinks[0][0] != '"' {
		t.Fatalf("retained string link representation changed: %s", rawStable["links"])
	}
	if err := json.Unmarshal([]byte("["+string(heterogeneousLinks[1])+"]"), &rawLinks); err != nil {
		t.Fatalf("decode edited object link: %v", err)
	}
	assertRawJSONField(t, rawLinks[0], "future_link_field", map[string]any{"keep": true})
	var rawMedia []map[string]json.RawMessage
	if err := json.Unmarshal(rawStable["media"], &rawMedia); err != nil {
		t.Fatalf("decode raw software media: %v", err)
	}
	assertRawJSONField(t, rawMedia[0], "future_media_field", map[string]any{
		"credit": "fixture",
		"flags":  []any{"keep"},
	})
}

func TestSaveEditorDataSoftwareRoundTripUpdatesAddsDeletesAndReturnsRevision(t *testing.T) {
	root := newSoftwareRepoFixture(t)
	app := &App{repoRoot: root}
	t.Cleanup(func() { app.shutdown(context.Background()) })

	loaded, err := app.LoadEditorData()
	if err != nil {
		t.Fatalf("LoadEditorData() unexpected error: %v", err)
	}
	if loaded.SoftwareRevision == "" || loaded.SoftwareLocation != "data/software.json" {
		t.Fatalf("software load metadata = revision %q, location %q", loaded.SoftwareRevision, loaded.SoftwareLocation)
	}
	stable := softwareSummaryByID(t, loaded.Software, "stable-tool")
	preview := softwareSummaryByID(t, loaded.Software, "preview-tool")
	edited := softwareInputFromSummary(stable)
	edited.Name = "Stable Tool Revised"
	edited.NotesEN = []string{"Revised English note"}
	edited.NotesKR = []string{"수정한 국문 설명"}
	newItem := SoftwareInput{
		Name:         "Fresh Utility",
		Stage:        "development",
		Links:        []SoftwareLinkInput{{URL: "https://example.org/fresh-utility"}},
		NotesEN:      []string{"A newly added utility"},
		NotesKR:      []string{"새로 추가한 유틸리티"},
		Media:        []SoftwareMediaInput{},
		Technologies: []string{"Go"},
	}

	saved, err := app.SaveEditorData(SaveEditorDataRequest{
		Settings:                   loaded.Settings,
		SettingsRevision:           loaded.SettingsRevision,
		ProfileRevision:            loaded.ProfileRevision,
		PeopleRevision:             loaded.PeopleRevision,
		AwardsRevision:             loaded.AwardsRevision,
		AcademicActivitiesRevision: loaded.AcademicActivitiesRevision,
		PublicationsRevision:       loaded.PublicationsRevision,
		ProjectsRevision:           loaded.ProjectsRevision,
		Software: []SoftwareSaveItem{
			{EditorKey: stable.EditorKey, Software: edited},
			{Software: &newItem},
			{EditorKey: preview.EditorKey},
		},
		SoftwareRevision: loaded.SoftwareRevision,
		SaveSoftware:     true,
		BoardRevision:    loaded.BoardRevision,
	})
	if err != nil {
		t.Fatalf("SaveEditorData() software save unexpected error: %v", err)
	}
	wantIDs := []string{"stable-tool", "fresh-utility", "preview-tool"}
	gotIDs := make([]string, 0, len(saved.Software))
	for _, item := range saved.Software {
		gotIDs = append(gotIDs, item.ID)
	}
	if !reflect.DeepEqual(gotIDs, wantIDs) {
		t.Errorf("saved software IDs/order = %#v, want %#v", gotIDs, wantIDs)
	}
	if strings.Contains(string(readFixtureFile(t, filepath.Join(root, "data", "software.json"))), "development-tool") {
		t.Error("software omitted from the public save request was not deleted")
	}
	storedStable := softwareSummaryByID(t, saved.Software, "stable-tool")
	if storedStable.Name != edited.Name || !reflect.DeepEqual(storedStable.NotesEN, edited.NotesEN) {
		t.Errorf("public software save did not return edited detail: %#v", storedStable)
	}
	storedRaw := readFixtureFile(t, filepath.Join(root, "data", "software.json"))
	if saved.SoftwareRevision == loaded.SoftwareRevision || saved.SoftwareRevision != revisionOf(storedRaw) {
		t.Errorf("software response revision = %q, old %q, stored %q", saved.SoftwareRevision, loaded.SoftwareRevision, revisionOf(storedRaw))
	}
	if saved.SoftwareLocation != loaded.SoftwareLocation {
		t.Errorf("software response location = %q, want %q", saved.SoftwareLocation, loaded.SoftwareLocation)
	}
}

func TestBuildSoftwareSaveRejectsInvalidFieldsAndTokensWithoutWriting(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*SoftwareInput)
	}{
		{name: "invalid stage", mutate: func(input *SoftwareInput) { input.Stage = "stable" }},
		{name: "empty Korean note", mutate: func(input *SoftwareInput) { input.NotesKR = append(input.NotesKR, " ") }},
		{name: "empty technologies", mutate: func(input *SoftwareInput) { input.Technologies = []string{} }},
		{name: "null links array", mutate: func(input *SoftwareInput) { input.Links = nil }},
		{name: "invalid link scheme", mutate: func(input *SoftwareInput) { input.Links = []SoftwareLinkInput{{URL: "ftp://example.org/file"}} }},
		{name: "missing English caption", mutate: func(input *SoftwareInput) { input.Media[0].CaptionEN = "" }},
		{name: "unknown media token", mutate: func(input *SoftwareInput) {
			input.Media = []SoftwareMediaInput{{StageToken: "unknown-token", CaptionEN: "English", CaptionKO: "국문"}}
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := newSoftwareRepoFixture(t)
			softwarePath := filepath.Join(root, "data", "software.json")
			snapshot, err := readSoftware(softwarePath, root)
			if err != nil {
				t.Fatalf("readSoftware() unexpected error: %v", err)
			}
			stable := softwareSummaryByID(t, softwareItemSummaries(snapshot), "stable-tool")
			input := softwareInputFromSummary(stable)
			test.mutate(input)
			before := snapshotBoardRepo(t, root)
			app := &App{repoRoot: root}
			_, _, _, err = app.buildSoftwareSaveLocked(root, snapshot, []SoftwareSaveItem{{
				EditorKey: stable.EditorKey,
				Software:  input,
			}}, nil)
			if err == nil {
				t.Fatal("buildSoftwareSaveLocked() error = nil, want validation rejection")
			}
			assertBoardRepoSnapshot(t, root, before, "invalid software build")
		})
	}
}

func TestBuildSoftwareSaveRejectsUnavailableCrossDomainStageToken(t *testing.T) {
	root := newSoftwareRepoFixture(t)
	snapshot, err := readSoftware(filepath.Join(root, "data", "software.json"), root)
	if err != nil {
		t.Fatalf("readSoftware() unexpected error: %v", err)
	}
	app := &App{repoRoot: root}
	t.Cleanup(func() { app.shutdown(context.Background()) })
	inputPath := filepath.Join(t.TempDir(), "shared-token.png")
	writeFixtureFile(t, inputPath, fixturePNGBytes())
	staged, err := app.StageBoardMedia([]string{inputPath})
	if err != nil || len(staged.Items) != 1 {
		t.Fatalf("StageBoardMedia() = %#v, %v", staged, err)
	}
	newItem := SoftwareInput{
		Name:    "Cross Domain Token",
		Stage:   "development",
		Links:   []SoftwareLinkInput{},
		NotesEN: []string{"English"},
		NotesKR: []string{"국문"},
		Media: []SoftwareMediaInput{{
			StageToken: staged.Items[0].StageToken,
			CaptionEN:  "English caption",
			CaptionKO:  "국문 캡션",
		}},
		Technologies: []string{"Go"},
	}
	_, _, _, err = app.buildSoftwareSaveLocked(root, snapshot, []SoftwareSaveItem{{Software: &newItem}}, map[string]bool{
		staged.Items[0].StageToken: true,
	})
	if err == nil {
		t.Fatal("buildSoftwareSaveLocked() accepted a stage token already used by another domain")
	}
}

func TestSaveEditorDataRejectsCrossDomainStageTokenWithoutWriting(t *testing.T) {
	root := newSoftwareRepoFixture(t)
	app := &App{repoRoot: root}
	t.Cleanup(func() { app.shutdown(context.Background()) })
	loaded, err := app.LoadEditorData()
	if err != nil {
		t.Fatalf("LoadEditorData() unexpected error: %v", err)
	}
	inputPath := filepath.Join(t.TempDir(), "shared-public-token.png")
	writeFixtureFile(t, inputPath, fixturePNGBytes())
	staged, err := app.StageBoardMedia([]string{inputPath})
	if err != nil || len(staged.Items) != 1 {
		t.Fatalf("StageBoardMedia() = %#v, %v; want one staged image", staged, err)
	}
	token := staged.Items[0].StageToken
	before := snapshotBoardRepo(t, root)
	boardPost := BoardPostInput{
		StartDate: "2026-08-15",
		TitleEN:   "Shared token board post",
		TitleKO:   "공유 토큰 게시글",
		ContentEN: "Board content",
		ContentKO: "게시글 본문",
		Media: []BoardMediaInput{{
			StageToken: token,
			CaptionEN:  "Board caption",
			CaptionKO:  "게시글 설명",
		}},
	}
	software := SoftwareInput{
		Name:    "Shared Token Utility",
		Stage:   "development",
		Links:   []SoftwareLinkInput{},
		NotesEN: []string{"English"},
		NotesKR: []string{"국문"},
		Media: []SoftwareMediaInput{{
			StageToken: token,
			CaptionEN:  "Software caption",
			CaptionKO:  "소프트웨어 설명",
		}},
		Technologies: []string{"Go"},
	}
	_, err = app.SaveEditorData(SaveEditorDataRequest{
		Settings:                   loaded.Settings,
		SettingsRevision:           loaded.SettingsRevision,
		ProfileRevision:            loaded.ProfileRevision,
		PeopleRevision:             loaded.PeopleRevision,
		AwardsRevision:             loaded.AwardsRevision,
		AcademicActivitiesRevision: loaded.AcademicActivitiesRevision,
		PublicationsRevision:       loaded.PublicationsRevision,
		ProjectsRevision:           loaded.ProjectsRevision,
		Software:                   []SoftwareSaveItem{{Software: &software}},
		SoftwareRevision:           loaded.SoftwareRevision,
		SaveSoftware:               true,
		Board:                      []BoardSaveItem{{Post: &boardPost}},
		BoardRevision:              loaded.BoardRevision,
		SaveBoard:                  true,
	})
	if err == nil || !strings.Contains(err.Error(), "중복") {
		t.Fatalf("SaveEditorData() duplicate cross-domain token error = %v, want duplicate rejection", err)
	}
	assertBoardRepoSnapshot(t, root, before, "cross-domain duplicate save")
	app.mu.Lock()
	retained, exists := app.boardMedia[token]
	app.mu.Unlock()
	if !exists {
		t.Fatal("rejected cross-domain save discarded the staged image")
	}
	if _, err := os.Stat(retained.Path); err != nil {
		t.Fatalf("rejected cross-domain save removed staged bytes: %v", err)
	}
}

func TestSaveEditorDataRejectsExternalSoftwareJSONEditWithoutWriting(t *testing.T) {
	root := newSoftwareRepoFixture(t)
	app := &App{repoRoot: root}
	loaded, err := app.LoadEditorData()
	if err != nil {
		t.Fatalf("LoadEditorData() unexpected error: %v", err)
	}
	softwarePath := filepath.Join(root, "data", "software.json")
	external := bytes.Replace(
		readFixtureFile(t, softwarePath),
		[]byte(`"name": "Development Tool"`),
		[]byte(`"name": "Manually Edited Tool"`),
		1,
	)
	if bytes.Equal(external, readFixtureFile(t, softwarePath)) {
		t.Fatal("external software fixture replacement did not change source bytes")
	}
	writeFixtureFile(t, softwarePath, external)
	before := snapshotBoardRepo(t, root)
	stable := softwareSummaryByID(t, loaded.Software, "stable-tool")
	edited := softwareInputFromSummary(stable)
	edited.Name = "Editor Attempt"

	_, err = app.SaveEditorData(SaveEditorDataRequest{
		Settings:                   loaded.Settings,
		SettingsRevision:           loaded.SettingsRevision,
		ProfileRevision:            loaded.ProfileRevision,
		PeopleRevision:             loaded.PeopleRevision,
		AwardsRevision:             loaded.AwardsRevision,
		AcademicActivitiesRevision: loaded.AcademicActivitiesRevision,
		PublicationsRevision:       loaded.PublicationsRevision,
		ProjectsRevision:           loaded.ProjectsRevision,
		Software: []SoftwareSaveItem{{
			EditorKey: stable.EditorKey,
			Software:  edited,
		}},
		SoftwareRevision: loaded.SoftwareRevision,
		SaveSoftware:     true,
		BoardRevision:    loaded.BoardRevision,
	})
	if err == nil || !strings.Contains(err.Error(), "software.json") {
		t.Fatalf("SaveEditorData() external software conflict error = %v, want software.json rejection", err)
	}
	assertBoardRepoSnapshot(t, root, before, "external software conflict")
	if got := readFixtureFile(t, softwarePath); !bytes.Equal(got, external) {
		t.Fatal("external software edit was replaced by stale editor data")
	}
}

func TestSaveEditorDataKeepsPublishedMediaWhenCommitMayBePartial(t *testing.T) {
	root := newSoftwareRepoFixture(t)
	app := &App{repoRoot: root}
	t.Cleanup(func() { app.shutdown(context.Background()) })
	loaded, err := app.LoadEditorData()
	if err != nil {
		t.Fatalf("LoadEditorData() unexpected error: %v", err)
	}
	inputPath := filepath.Join(t.TempDir(), "partial-commit.png")
	pngContents := fixturePNGBytes()
	writeFixtureFile(t, inputPath, pngContents)
	staged, err := app.StageBoardMedia([]string{inputPath})
	if err != nil || len(staged.Items) != 1 {
		t.Fatalf("StageBoardMedia() = %#v, %v; want one staged image", staged, err)
	}
	newItem := SoftwareInput{
		Name:    "Partial Commit Tool",
		Stage:   "development",
		Links:   []SoftwareLinkInput{},
		NotesEN: []string{"English"},
		NotesKR: []string{"국문"},
		Media: []SoftwareMediaInput{{
			StageToken: staged.Items[0].StageToken,
			CaptionEN:  "Published before commit",
			CaptionKO:  "커밋 전에 게시한 이미지",
		}},
		Technologies: []string{"Go"},
	}

	originalCommit := commitEditorFiles
	commitEditorFiles = func(writes []editorFileWrite) (bool, error) {
		for _, write := range writes {
			if !write.Enabled || bytes.Equal(write.Previous, write.Next) {
				continue
			}
			if err := writeFileAtomically(write.Path, write.Next); err != nil {
				return false, err
			}
			// Model a later failure whose rollback could not safely restore this
			// already-written JSON file.
			return false, errors.New("injected incomplete rollback")
		}
		return false, errors.New("injected commit found no enabled write")
	}
	t.Cleanup(func() { commitEditorFiles = originalCommit })

	_, err = app.SaveEditorData(SaveEditorDataRequest{
		Settings:                   loaded.Settings,
		SettingsRevision:           loaded.SettingsRevision,
		ProfileRevision:            loaded.ProfileRevision,
		PeopleRevision:             loaded.PeopleRevision,
		AwardsRevision:             loaded.AwardsRevision,
		AcademicActivitiesRevision: loaded.AcademicActivitiesRevision,
		PublicationsRevision:       loaded.PublicationsRevision,
		ProjectsRevision:           loaded.ProjectsRevision,
		Software:                   []SoftwareSaveItem{{Software: &newItem}},
		SoftwareRevision:           loaded.SoftwareRevision,
		SaveSoftware:               true,
		BoardRevision:              loaded.BoardRevision,
	})
	if err == nil || !strings.Contains(err.Error(), "incomplete rollback") {
		t.Fatalf("SaveEditorData() partial commit error = %v", err)
	}
	stored, err := readSoftware(filepath.Join(root, "data", "software.json"), root)
	if err != nil {
		t.Fatalf("partially committed software must retain its published media: %v", err)
	}
	if len(stored.Rows) != 1 || stored.Rows[0].Item.ID != "partial-commit-tool" || len(stored.Rows[0].Item.Media) != 1 {
		t.Fatalf("partial software commit = %#v", stored.Rows)
	}
	mediaPath := filepath.Join(root, "data", "media", stored.Rows[0].Item.Media[0].Src)
	if got := readFixtureFile(t, mediaPath); !bytes.Equal(got, pngContents) {
		t.Fatal("partial commit removed or changed media still referenced by software.json")
	}
}

func TestReadSoftwareRejectsInvalidOrDuplicateIDs(t *testing.T) {
	root := newSoftwareRepoFixture(t)
	valid := fixtureSoftwareJSON()
	tests := []struct {
		name string
		raw  []byte
	}{
		{name: "numeric ID", raw: bytes.Replace(valid, []byte(`"id": "development-tool"`), []byte(`"id": 123`), 1)},
		{name: "whitespace ID", raw: bytes.Replace(valid, []byte(`"id": "development-tool"`), []byte(`"id": " development-tool"`), 1)},
		{name: "duplicate ID", raw: bytes.Replace(valid, []byte(`"id": "preview-tool"`), []byte(`"id": "stable-tool"`), 1)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := readSoftwareBytes(test.raw, root); err == nil {
				t.Fatal("readSoftwareBytes() error = nil, want invalid ID rejection")
			}
		})
	}
}

func newSoftwareRepoFixture(t *testing.T) string {
	t.Helper()
	root := newBoardRepoFixture(t)
	mediaDirectory := filepath.Join(root, "data", "media")
	writeFixtureFile(t, filepath.Join(mediaDirectory, "software-stable-tool.jpg"), fixtureJPEGBytes())
	writeFixtureFile(t, filepath.Join(mediaDirectory, "software-stable-poster.jpg"), fixtureJPEGBytes())
	writeFixtureFile(t, filepath.Join(root, "data", "software.json"), fixtureSoftwareJSON())
	return root
}

func fixtureSoftwareJSON() []byte {
	return []byte(`[
  {
    "id": "development-tool",
    "name": "Development Tool",
    "stage": "development",
    "links": [],
    "notes_en": ["Development English note"],
    "notes_kr": ["개발 국문 설명"],
    "media": [],
    "technologies": ["Rust"]
  },
  {
    "id": "stable-tool",
    "name": "Stable Tool",
    "stage": "release",
    "links": [
      "https://github.com/example/stable-tool",
      {
        "url": "https://example.org/docs",
        "label_en": "Documentation",
        "label_ko": "문서",
        "future_link_field": {"keep": true}
      }
    ],
    "notes_en": ["Stable English note"],
    "notes_kr": ["안정판 국문 설명"],
    "media": [
      {
        "src": "software-stable-tool.jpg",
        "type": "image",
        "poster": "software-stable-poster.jpg",
        "caption_en": "Stable screenshot",
        "caption_ko": "안정판 화면",
        "future_media_field": {"credit": "fixture", "flags": ["keep"]}
      }
    ],
    "technologies": ["Go", "Wails"],
    "future_row_field": {"mode": "lossless", "rank": 4}
  },
  {
    "id": "preview-tool",
    "name": "Preview Tool",
    "stage": "preview",
    "links": [{"url": "https://example.org/preview", "label": "Preview"}],
    "notes_en": ["Preview English note"],
    "notes_kr": ["프리뷰 국문 설명"],
    "media": [],
    "technologies": ["Python"]
  }
]
`)
}

func softwareSummaryByID(t *testing.T, items []SoftwareItemSummary, id string) SoftwareItemSummary {
	t.Helper()
	for _, item := range items {
		if item.ID == id {
			return item
		}
	}
	t.Fatalf("software summary %q not found in %#v", id, items)
	return SoftwareItemSummary{}
}

func softwareInputFromSummary(item SoftwareItemSummary) *SoftwareInput {
	links := make([]SoftwareLinkInput, 0, len(item.Links))
	for _, link := range item.Links {
		links = append(links, softwareLinkInputFromSummary(link))
	}
	media := make([]SoftwareMediaInput, 0, len(item.Media))
	for _, value := range item.Media {
		media = append(media, SoftwareMediaInput{
			EditorKey: value.EditorKey,
			CaptionEN: value.CaptionEN,
			CaptionKO: value.CaptionKO,
		})
	}
	return &SoftwareInput{
		Name:         item.Name,
		Stage:        item.Stage,
		Links:        links,
		NotesEN:      append([]string(nil), item.NotesEN...),
		NotesKR:      append([]string(nil), item.NotesKR...),
		Media:        media,
		Technologies: append([]string(nil), item.Technologies...),
	}
}

func softwareLinkInputFromSummary(link SoftwareLinkSummary) SoftwareLinkInput {
	return SoftwareLinkInput{
		EditorKey: link.EditorKey,
		URL:       link.URL,
		Label:     link.Label,
		LabelEN:   link.LabelEN,
		LabelKO:   link.LabelKO,
	}
}

func softwareRawItemByID(t *testing.T, raw []byte, id string) map[string]json.RawMessage {
	t.Helper()
	var items []map[string]json.RawMessage
	if err := json.Unmarshal(raw, &items); err != nil {
		t.Fatalf("decode raw software items: %v", err)
	}
	for _, item := range items {
		var itemID string
		if err := json.Unmarshal(item["id"], &itemID); err != nil {
			continue
		}
		if itemID == id {
			return item
		}
	}
	t.Fatalf("raw software item %q not found", id)
	return nil
}
