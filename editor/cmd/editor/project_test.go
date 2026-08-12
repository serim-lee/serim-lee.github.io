package main

import (
	"bytes"
	"context"
	"encoding/json"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestSaveEditorDataIntegratesProjectEditAddAndStagedMedia(t *testing.T) {
	root := newBoardRepoFixture(t)
	app := &App{repoRoot: root}
	t.Cleanup(func() { app.shutdown(context.Background()) })
	loaded, err := app.LoadEditorData()
	if err != nil {
		t.Fatalf("LoadEditorData() unexpected error: %v", err)
	}
	if len(loaded.Projects) != 1 || loaded.Projects[0].TitleEN != "Fixture project" {
		t.Fatalf("loaded projects = %#v, want the fixture project", loaded.Projects)
	}
	projectsBeforeStage := readProjectTestFile(t, projectFixturePath(root))
	boardBeforeSave := readProjectTestFile(t, filepath.Join(root, "data", "board.json"))

	inputPath := filepath.Join(t.TempDir(), "integrated-project-image.png")
	imageBytes := projectTestPNGBytes()
	writeProjectTestFile(t, inputPath, imageBytes)
	staged, err := app.StageBoardMedia([]string{inputPath})
	if err != nil || len(staged.Items) != 1 || len(staged.Rejected) != 0 {
		t.Fatalf("StageBoardMedia() = %#v, %v; want one staged image", staged, err)
	}
	if got := readProjectTestFile(t, projectFixturePath(root)); !bytes.Equal(got, projectsBeforeStage) {
		t.Fatal("staging project media changed projects.json before explicit save")
	}

	existing := loaded.Projects[0]
	request := SaveEditorDataRequest{
		Settings:                   loaded.Settings,
		SettingsRevision:           loaded.SettingsRevision,
		ProfileRevision:            loaded.ProfileRevision,
		PeopleRevision:             loaded.PeopleRevision,
		AwardsRevision:             loaded.AwardsRevision,
		AcademicActivitiesRevision: loaded.AcademicActivitiesRevision,
		PublicationsRevision:       loaded.PublicationsRevision,
		ProjectsRevision:           loaded.ProjectsRevision,
		SaveProjects:               true,
		SoftwareRevision:           loaded.SoftwareRevision,
		BoardRevision:              loaded.BoardRevision,
		Projects: []ProjectSaveItem{
			{Project: &ProjectInput{
				TitleEN:   "Older project deliberately first",
				TitleKO:   "요청 순서 첫 프로젝트",
				StartDate: "2010-01-01",
				EndDate:   "2010-12-31",
				Theme:     "retrofit",
				FunderEN:  "New funder",
				FunderKO:  "신규 발주처",
				NotesEN:   []string{"New project note"},
				NotesKR:   []string{"신규 프로젝트 노트"},
				Media:     []ProjectMediaInput{},
			}},
			{
				EditorKey: existing.EditorKey,
				Project: &ProjectInput{
					TitleEN:   "Edited fixture project",
					TitleKO:   "수정된 기존 프로젝트",
					StartDate: "2028-01-01",
					EndDate:   "2028-12-31",
					Theme:     existing.Theme,
					FunderEN:  existing.FunderEN,
					FunderKO:  existing.FunderKO,
					NotesEN:   []string{"Edited fixture note"},
					NotesKR:   []string{"수정된 기존 노트"},
					Media: []ProjectMediaInput{{
						StageToken: staged.Items[0].StageToken,
						CaptionEN:  "Integrated English caption",
						CaptionKO:  "통합 국문 캡션",
					}},
				},
			},
		},
	}
	saved, err := app.SaveEditorData(request)
	if err != nil {
		t.Fatalf("SaveEditorData() project integration unexpected error: %v", err)
	}
	if len(saved.Projects) != 2 ||
		saved.Projects[0].TitleEN != "Older project deliberately first" ||
		saved.Projects[1].TitleEN != "Edited fixture project" {
		t.Fatalf("saved project response changed request order or fields: %#v", saved.Projects)
	}
	if saved.ProjectsRevision == loaded.ProjectsRevision {
		t.Fatal("project revision did not change after save")
	}
	writtenProjects := readProjectTestFile(t, projectFixturePath(root))
	if saved.ProjectsRevision != revisionOf(writtenProjects) {
		t.Errorf("project response revision = %q, want written file revision %q", saved.ProjectsRevision, revisionOf(writtenProjects))
	}
	if len(saved.Projects[1].Media) != 1 || saved.Projects[1].Media[0].Src == "" || saved.Projects[1].Media[0].PreviewURL == "" {
		t.Fatalf("saved staged project media descriptor = %#v", saved.Projects[1].Media)
	}
	publishedPath := filepath.Join(root, "data", "media", filepath.FromSlash(saved.Projects[1].Media[0].Src))
	if got := readProjectTestFile(t, publishedPath); !bytes.Equal(got, imageBytes) {
		t.Fatal("published project media differs from the staged image")
	}
	if got := readProjectTestFile(t, filepath.Join(root, "data", "board.json")); !bytes.Equal(got, boardBeforeSave) {
		t.Fatal("project-only SaveEditorData changed board.json")
	}
	if app.stagingDir != "" || len(app.boardMedia) != 0 {
		t.Errorf("successful project save left staging state: dir=%q media=%d", app.stagingDir, len(app.boardMedia))
	}
}

func TestReadProjectsReturnsEditableRowsInArrayOrder(t *testing.T) {
	root := newProjectRepoFixture(t)
	snapshot, err := readProjects(projectFixturePath(root), root, projectFixtureThemeIDs())
	if err != nil {
		t.Fatalf("readProjects() unexpected error: %v", err)
	}
	items := projectItemsResponse(snapshot)
	if len(items) != 2 {
		t.Fatalf("project response length = %d, want 2", len(items))
	}
	if items[0].TitleEN != "First in file" || items[1].TitleEN != "Second in file" {
		t.Fatalf("project response changed array order: %#v", items)
	}
	for index, item := range items {
		if item.EditorKey == "" {
			t.Errorf("project %d has an empty editor key", index)
		}
		if item.NotesEN == nil || item.NotesKR == nil || item.Media == nil {
			t.Errorf("project %d has nil editable arrays: %#v", index, item)
		}
	}
	media := items[1].Media
	if len(media) != 1 || media[0].EditorKey == "" {
		t.Fatalf("existing project media lacks an editor key: %#v", media)
	}
	if media[0].Src != "project-existing.jpg" || media[0].Type != "image" || media[0].Poster != "project-poster.jpg" {
		t.Errorf("existing project media identity = %#v", media[0])
	}
	if media[0].OriginalName != "project-existing.jpg" || media[0].Size <= 0 || media[0].PreviewURL == "" {
		t.Errorf("existing project media descriptor incomplete: %#v", media[0])
	}
}

func TestBuildProjectSavePreservesRequestedOrderAndUnknownFields(t *testing.T) {
	root := newProjectRepoFixture(t)
	snapshot := mustReadProjectSnapshot(t, root)
	items := projectItemsResponse(snapshot)
	request := []ProjectSaveItem{
		{EditorKey: items[1].EditorKey},
		{Project: &ProjectInput{
			TitleEN:   "New final row",
			TitleKO:   "새 마지막 항목",
			StartDate: "2027-02-01",
			EndDate:   "2027-12-31",
			Theme:     "hvac-modeling-control",
			FunderEN:  "New funder",
			FunderKO:  "새 발주처",
			NotesEN:   []string{"New English note"},
			NotesKR:   []string{"새 국문 노트"},
			Media:     []ProjectMediaInput{},
		}},
	}

	encoded, pending, tokens, err := (&App{}).buildProjectSaveLocked(
		root,
		snapshot,
		request,
		projectFixtureThemeIDs(),
		nil,
	)
	if err != nil {
		t.Fatalf("buildProjectSaveLocked() unexpected error: %v", err)
	}
	if len(pending) != 0 || len(tokens) != 0 {
		t.Fatalf("project save unexpectedly prepared media: pending=%#v tokens=%#v", pending, tokens)
	}
	var rows []map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &rows); err != nil {
		t.Fatalf("decode prepared projects: %v", err)
	}
	if len(rows) != 2 || projectRawString(t, rows[0], "title_en") != "Second in file" || projectRawString(t, rows[1], "title_en") != "New final row" {
		t.Fatalf("prepared projects did not preserve request order: %#v", rows)
	}
	assertProjectRawField(t, rows[0], "future_row_field", map[string]any{
		"rank": float64(9),
		"tags": []any{"keep", "project"},
	})
	var media []map[string]json.RawMessage
	if err := json.Unmarshal(rows[0]["media"], &media); err != nil || len(media) != 1 {
		t.Fatalf("decode preserved media: %v; media=%#v", err, media)
	}
	assertProjectRawField(t, media[0], "future_media_field", map[string]any{
		"credit": "fixture",
		"flags":  []any{true, float64(3)},
	})
	if _, exists := rows[1]["id"]; exists {
		t.Fatal("new project unexpectedly received an id")
	}
}

func TestBuildProjectSaveEditsExistingMediaAndStagesNewMediaLosslessly(t *testing.T) {
	root := newProjectRepoFixture(t)
	snapshot := mustReadProjectSnapshot(t, root)
	items := projectItemsResponse(snapshot)
	existing := items[1]
	inputPath := filepath.Join(t.TempDir(), "new-project-image.dat")
	pngBytes := projectTestPNGBytes()
	writeProjectTestFile(t, inputPath, pngBytes)
	app := &App{}
	staged, err := app.StageBoardMedia([]string{inputPath})
	if err != nil || len(staged.Items) != 1 || len(staged.Rejected) != 0 {
		t.Fatalf("StageBoardMedia() = %#v, %v; want one staged project image", staged, err)
	}
	t.Cleanup(func() { _ = app.DiscardBoardMedia(nil) })

	request := []ProjectSaveItem{{
		EditorKey: existing.EditorKey,
		Project: &ProjectInput{
			TitleEN:   "Edited project",
			TitleKO:   "수정 프로젝트",
			StartDate: existing.StartDate,
			EndDate:   existing.EndDate,
			Theme:     existing.Theme,
			FunderEN:  existing.FunderEN,
			FunderKO:  existing.FunderKO,
			NotesEN:   []string{"Edited English note"},
			NotesKR:   []string{"수정 국문 노트"},
			Media: []ProjectMediaInput{
				{
					EditorKey: existing.Media[0].EditorKey,
					CaptionEN: "Edited English caption",
					CaptionKO: "수정 국문 캡션",
				},
				{
					StageToken: staged.Items[0].StageToken,
					CaptionEN:  "New English caption",
					CaptionKO:  "새 국문 캡션",
				},
			},
		},
	}}

	encoded, pending, tokens, err := app.buildProjectSaveLocked(
		root,
		snapshot,
		request,
		projectFixtureThemeIDs(),
		nil,
	)
	if err != nil {
		t.Fatalf("buildProjectSaveLocked() unexpected error: %v", err)
	}
	if len(pending) != 1 || len(tokens) != 1 || tokens[0] != staged.Items[0].StageToken {
		t.Fatalf("prepared project media = %#v, tokens=%#v", pending, tokens)
	}
	if _, err := os.Stat(pending[0].DestinationPath); !os.IsNotExist(err) {
		t.Fatalf("buildProjectSaveLocked() published before explicit save: %v", err)
	}
	if !strings.HasSuffix(filepath.ToSlash(pending[0].DestinationPath), "project-20250101-edited-project-02.png") {
		t.Errorf("new project media destination = %q", pending[0].DestinationPath)
	}

	var rows []map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &rows); err != nil || len(rows) != 1 {
		t.Fatalf("decode edited project: %v; rows=%#v", err, rows)
	}
	assertProjectRawField(t, rows[0], "future_row_field", map[string]any{
		"rank": float64(9),
		"tags": []any{"keep", "project"},
	})
	var media []map[string]json.RawMessage
	if err := json.Unmarshal(rows[0]["media"], &media); err != nil || len(media) != 2 {
		t.Fatalf("decode edited project media: %v; media=%#v", err, media)
	}
	if projectRawString(t, media[0], "src") != "project-existing.jpg" ||
		projectRawString(t, media[0], "type") != "image" ||
		projectRawString(t, media[0], "poster") != "project-poster.jpg" {
		t.Errorf("retained project media identity changed: %#v", media[0])
	}
	if projectRawString(t, media[0], "caption_en") != "Edited English caption" || projectRawString(t, media[0], "caption_ko") != "수정 국문 캡션" {
		t.Errorf("retained project media captions were not edited: %#v", media[0])
	}
	assertProjectRawField(t, media[0], "future_media_field", map[string]any{
		"credit": "fixture",
		"flags":  []any{true, float64(3)},
	})
	if projectRawString(t, media[1], "caption_en") != "New English caption" || projectRawString(t, media[1], "caption_ko") != "새 국문 캡션" {
		t.Errorf("new project media captions = %#v", media[1])
	}

	created, err := publishBoardMedia(pending)
	if err != nil {
		t.Fatalf("publish prepared project media: %v", err)
	}
	t.Cleanup(func() { removeFiles(created) })
	got, err := os.ReadFile(pending[0].DestinationPath)
	if err != nil || !bytes.Equal(got, pngBytes) {
		t.Fatalf("published project media = %x, %v; want staged bytes", got, err)
	}
	if got := readProjectTestFile(t, filepath.Join(root, "data", "media", "project-existing.jpg")); !bytes.Equal(got, projectTestJPEGBytes()) {
		t.Fatal("editing project media changed its existing source file")
	}
}

func TestBuildProjectSaveFullUnchangedPayloadPreservesMissingOptionalFields(t *testing.T) {
	root := newProjectRepoFixture(t)
	snapshot := mustReadProjectSnapshot(t, root)
	items := projectItemsResponse(snapshot)
	request := make([]ProjectSaveItem, 0, len(items))
	for _, item := range items {
		media := make([]ProjectMediaInput, 0, len(item.Media))
		for _, currentMedia := range item.Media {
			media = append(media, ProjectMediaInput{
				EditorKey: currentMedia.EditorKey,
				CaptionEN: currentMedia.CaptionEN,
				CaptionKO: currentMedia.CaptionKO,
			})
		}
		request = append(request, ProjectSaveItem{
			EditorKey: item.EditorKey,
			Project: &ProjectInput{
				TitleEN:   item.TitleEN,
				TitleKO:   item.TitleKO,
				StartDate: item.StartDate,
				EndDate:   item.EndDate,
				Theme:     item.Theme,
				FunderEN:  item.FunderEN,
				FunderKO:  item.FunderKO,
				NotesEN:   item.NotesEN,
				NotesKR:   item.NotesKR,
				Media:     media,
			},
		})
	}
	encoded, _, _, err := (&App{}).buildProjectSaveLocked(root, snapshot, request, projectFixtureThemeIDs(), nil)
	if err != nil {
		t.Fatalf("buildProjectSaveLocked() unchanged payload unexpected error: %v", err)
	}
	var rows []map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &rows); err != nil {
		t.Fatalf("decode unchanged projects: %v", err)
	}
	if _, exists := rows[0]["funder_ko"]; exists {
		t.Fatal("unchanged full payload introduced an originally omitted funder_ko")
	}
	assertProjectRawField(t, rows[0], "future_optional", map[string]any{
		"nested": map[string]any{"value": "preserve"},
	})
}

func TestProjectValidationRejectsUnknownThemeAndMismatchedCaptions(t *testing.T) {
	root := newProjectRepoFixture(t)
	snapshot := mustReadProjectSnapshot(t, root)
	base := ProjectInput{
		TitleEN:   "Invalid project",
		TitleKO:   "잘못된 프로젝트",
		StartDate: "2026-01-01",
		EndDate:   "2026-12-31",
		Theme:     "unknown-theme",
		NotesEN:   []string{"English note"},
		NotesKR:   []string{"국문 노트"},
		Media:     []ProjectMediaInput{},
	}
	if _, _, _, err := (&App{}).buildProjectSaveLocked(
		root,
		snapshot,
		[]ProjectSaveItem{{Project: &base}},
		projectFixtureThemeIDs(),
		nil,
	); err == nil {
		t.Fatal("unknown project theme was accepted")
	}

	base.Theme = "retrofit"
	base.Media = []ProjectMediaInput{{StageToken: "missing-token", CaptionEN: "only English"}}
	if _, _, _, err := (&App{}).buildProjectSaveLocked(
		root,
		snapshot,
		[]ProjectSaveItem{{Project: &base}},
		projectFixtureThemeIDs(),
		nil,
	); err == nil {
		t.Fatal("mismatched project media captions were accepted")
	}
}

func newProjectRepoFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	mediaDir := filepath.Join(root, "data", "media")
	if err := os.MkdirAll(mediaDir, 0o755); err != nil {
		t.Fatalf("create project fixture media directory: %v", err)
	}
	writeProjectTestFile(t, filepath.Join(mediaDir, "project-existing.jpg"), projectTestJPEGBytes())
	writeProjectTestFile(t, filepath.Join(mediaDir, "project-poster.jpg"), projectTestJPEGBytes())
	projects := []byte(`[
  {
    "title_en": "First in file",
    "title_ko": "첫 번째 프로젝트",
    "start_date": "2020-01-01",
    "end_date": "2020-12-31",
    "theme": "retrofit",
    "funder_en": "First funder",
    "notes_en": ["First English note"],
    "notes_kr": ["첫 번째 국문 노트"],
    "media": [],
    "future_optional": {"nested": {"value": "preserve"}}
  },
  {
    "title_en": "Second in file",
    "title_ko": "두 번째 프로젝트",
    "start_date": "2025-01-01",
    "end_date": "2025-12-31",
    "theme": "retrofit",
    "funder_en": "Second funder",
    "funder_ko": "두 번째 발주처",
    "notes_en": ["Second English note"],
    "notes_kr": ["두 번째 국문 노트"],
    "media": [
      {
        "src": "project-existing.jpg",
        "type": "image",
        "poster": "project-poster.jpg",
        "caption_en": "Existing English caption",
        "caption_ko": "기존 국문 캡션",
        "future_media_field": {"credit": "fixture", "flags": [true, 3]}
      }
    ],
    "future_row_field": {"rank": 9, "tags": ["keep", "project"]}
  }
]
`)
	writeProjectTestFile(t, projectFixturePath(root), projects)
	return root
}

func projectFixturePath(root string) string {
	return filepath.Join(root, "data", "projects.json")
}

func projectFixtureThemeIDs() map[string]bool {
	return map[string]bool{
		"retrofit":              true,
		"hvac-modeling-control": true,
	}
}

func mustReadProjectSnapshot(t *testing.T, root string) projectSnapshot {
	t.Helper()
	snapshot, err := readProjects(projectFixturePath(root), root, projectFixtureThemeIDs())
	if err != nil {
		t.Fatalf("read project fixture: %v", err)
	}
	return snapshot
}

func projectRawString(t *testing.T, fields map[string]json.RawMessage, name string) string {
	t.Helper()
	var value string
	if err := json.Unmarshal(fields[name], &value); err != nil {
		t.Fatalf("decode field %q: %v", name, err)
	}
	return value
}

func assertProjectRawField(t *testing.T, fields map[string]json.RawMessage, name string, want any) {
	t.Helper()
	raw, exists := fields[name]
	if !exists {
		t.Fatalf("unknown project JSON field %q was dropped", name)
	}
	var got any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("decode project JSON field %q: %v", name, err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("project JSON field %q = %#v, want %#v", name, got, want)
	}
}

func writeProjectTestFile(t *testing.T, path string, contents []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create parent directory for %q: %v", path, err)
	}
	if err := os.WriteFile(path, contents, 0o644); err != nil {
		t.Fatalf("write project test file %q: %v", path, err)
	}
}

func readProjectTestFile(t *testing.T, path string) []byte {
	t.Helper()
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read project test file %q: %v", path, err)
	}
	return contents
}

func projectTestJPEGBytes() []byte {
	var encoded bytes.Buffer
	picture := image.NewRGBA(image.Rect(0, 0, 2, 2))
	picture.Set(0, 0, color.RGBA{R: 210, G: 100, B: 30, A: 255})
	_ = jpeg.Encode(&encoded, picture, nil)
	return encoded.Bytes()
}

func projectTestPNGBytes() []byte {
	var encoded bytes.Buffer
	picture := image.NewRGBA(image.Rect(0, 0, 2, 2))
	picture.Set(0, 0, color.RGBA{R: 30, G: 100, B: 210, A: 255})
	_ = png.Encode(&encoded, picture)
	return encoded.Bytes()
}
