package main

import (
	"bytes"
	"context"
	"encoding/json"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestLoadEditorDataReturnsNewestBoardSummariesWithoutWriting(t *testing.T) {
	root := newBoardRepoFixture(t)
	before := snapshotBoardRepo(t, root)
	boardRaw := readBoardFixtureFile(t, root)
	app := &App{repoRoot: root, dirty: true}

	response, err := app.LoadEditorData()
	if err != nil {
		t.Fatalf("LoadEditorData() unexpected error: %v", err)
	}

	wantDates := []string{"2026-06-10", "2025-04-03", "2024-01-02"}
	wantTitles := []string{"Newest existing post", "Post to delete", "Oldest existing post"}
	if len(response.Board) != len(wantDates) {
		t.Fatalf("LoadEditorData() returned %d board summaries, want %d", len(response.Board), len(wantDates))
	}
	for index, item := range response.Board {
		if item.StartDate != wantDates[index] || item.TitleEN != wantTitles[index] {
			t.Errorf(
				"LoadEditorData() board[%d] = (%q, %q), want (%q, %q)",
				index,
				item.StartDate,
				item.TitleEN,
				wantDates[index],
				wantTitles[index],
			)
		}
		if item.EditorKey == "" {
			t.Errorf("LoadEditorData() board[%d] has an empty editor key", index)
		}
	}
	if response.Board[0].MediaCount != 1 {
		t.Errorf("newest board media count = %d, want 1", response.Board[0].MediaCount)
	}
	if response.BoardRevision != revisionOf(boardRaw) {
		t.Errorf("board revision = %q, want %q", response.BoardRevision, revisionOf(boardRaw))
	}
	if response.BoardLocation != filepath.ToSlash(filepath.Join("data", "board.json")) {
		t.Errorf("board location = %q, want data/board.json", response.BoardLocation)
	}
	if app.dirty {
		t.Error("LoadEditorData() left app dirty, want clean")
	}
	assertBoardRepoSnapshot(t, root, before, "LoadEditorData")
}

func TestSaveEditorDataDeletesAndAddsBoardPostsInDateOrder(t *testing.T) {
	root := newBoardRepoFixture(t)
	app := &App{repoRoot: root}
	loaded, err := app.LoadEditorData()
	if err != nil {
		t.Fatalf("LoadEditorData() unexpected error: %v", err)
	}

	keys := boardKeysByTitle(loaded.Board)
	requestRows := []BoardSaveItem{
		{EditorKey: keys["Oldest existing post"]},
		{NewPost: &NewBoardPost{
			StartDate: "2027-02-01",
			EndDate:   "",
			TitleEN:   "Added post",
			TitleKO:   "추가 게시글",
			ContentEN: "New English content",
			ContentKO: "새 국문 본문",
			Media:     []NewBoardMedia{},
		}},
		{EditorKey: keys["Newest existing post"]},
	}

	saved, err := app.SaveEditorData(boardOnlySaveRequest(loaded, requestRows))
	if err != nil {
		t.Fatalf("SaveEditorData() unexpected error: %v", err)
	}

	items := readBoardFixtureItems(t, root)
	wantTitles := []string{"Added post", "Newest existing post", "Oldest existing post"}
	wantDates := []string{"2027-02-01", "2026-06-10", "2024-01-02"}
	if len(items) != len(wantTitles) {
		t.Fatalf("saved board has %d items, want %d", len(items), len(wantTitles))
	}
	for index, item := range items {
		if item.TitleEN != wantTitles[index] || item.StartDate != wantDates[index] {
			t.Errorf(
				"saved board[%d] = (%q, %q), want (%q, %q)",
				index,
				item.StartDate,
				item.TitleEN,
				wantDates[index],
				wantTitles[index],
			)
		}
		if item.TitleEN == "Post to delete" {
			t.Error("deleted post remains in board.json")
		}
	}
	if len(saved.Board) != 3 || saved.Board[0].TitleEN != "Added post" {
		t.Errorf("SaveEditorData() board summaries are not in stored date order: %#v", saved.Board)
	}
	written := readBoardFixtureFile(t, root)
	if saved.BoardRevision != revisionOf(written) {
		t.Errorf("saved board revision = %q, want hash of written board %q", saved.BoardRevision, revisionOf(written))
	}
	var rawItems []map[string]json.RawMessage
	if err := json.Unmarshal(written, &rawItems); err != nil {
		t.Fatalf("decode saved board raw fields: %v", err)
	}
	var preserved bool
	for _, item := range rawItems {
		var title string
		_ = json.Unmarshal(item["title_en"], &title)
		if title == "Oldest existing post" {
			var future map[string]bool
			if err := json.Unmarshal(item["future_field"], &future); err == nil {
				preserved = future["preserve"]
			}
		}
	}
	if !preserved {
		t.Error("existing board future_field was not preserved semantically")
	}
}

func TestSaveEditorDataRejectsStaleBoardRevisionWithoutWriting(t *testing.T) {
	root := newBoardRepoFixture(t)
	boardPath := filepath.Join(root, "data", "board.json")
	app := &App{repoRoot: root}
	loaded, err := app.LoadEditorData()
	if err != nil {
		t.Fatalf("LoadEditorData() unexpected error: %v", err)
	}

	externalBoard := []byte(`[
  {
    "start_date": "2030-01-01",
    "end_date": "",
    "title_en": "External edit",
    "title_ko": "외부 수정",
    "content_en": "Changed outside the editor",
    "content_ko": "편집기 밖에서 변경",
    "media": []
  }
]
`)
	if err := os.WriteFile(boardPath, externalBoard, 0o644); err != nil {
		t.Fatalf("write external board edit: %v", err)
	}
	beforeRejectedSave := snapshotBoardRepo(t, root)

	_, err = app.SaveEditorData(boardOnlySaveRequest(loaded, []BoardSaveItem{{
		EditorKey: loaded.Board[0].EditorKey,
	}}))
	if err == nil {
		t.Fatal("SaveEditorData() with stale board revision error = nil, want conflict")
	}
	if !strings.Contains(err.Error(), "board.json") {
		t.Fatalf("stale revision error = %q, want board.json conflict", err)
	}
	assertBoardRepoSnapshot(t, root, beforeRejectedSave, "stale SaveEditorData")
	if got := readBoardFixtureFile(t, root); !bytes.Equal(got, externalBoard) {
		t.Fatalf("stale save replaced external board edit\ngot:  %s\nwant: %s", got, externalBoard)
	}
}

func TestSaveEditorDataCommitsSettingsAndBoardTogether(t *testing.T) {
	root := newBoardRepoFixture(t)
	settingsPath := filepath.Join(root, "data", "settings.json")
	boardPath := filepath.Join(root, "data", "board.json")
	app := &App{repoRoot: root}
	loaded, err := app.LoadEditorData()
	if err != nil {
		t.Fatalf("LoadEditorData() unexpected error: %v", err)
	}
	draft := loaded.Settings
	draft.MainPageSections[0], draft.MainPageSections[1] = draft.MainPageSections[1], draft.MainPageSections[0]
	rows := []BoardSaveItem{{EditorKey: boardKeysByTitle(loaded.Board)["Newest existing post"]}}

	saved, err := app.SaveEditorData(SaveEditorDataRequest{
		Settings:                   draft,
		SettingsRevision:           loaded.SettingsRevision,
		ProfileRevision:            loaded.ProfileRevision,
		SaveSettings:               true,
		PeopleRevision:             loaded.PeopleRevision,
		AwardsRevision:             loaded.AwardsRevision,
		AcademicActivitiesRevision: loaded.AcademicActivitiesRevision,
		PublicationsRevision:       loaded.PublicationsRevision,
		ProjectsRevision:           loaded.ProjectsRevision,
		SoftwareRevision:           loaded.SoftwareRevision,
		Board:                      rows,
		BoardRevision:              loaded.BoardRevision,
		SaveBoard:                  true,
	})
	if err != nil {
		t.Fatalf("SaveEditorData() combined save unexpected error: %v", err)
	}
	settingsRaw, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("read saved settings: %v", err)
	}
	boardRaw, err := os.ReadFile(boardPath)
	if err != nil {
		t.Fatalf("read saved board: %v", err)
	}
	if saved.SettingsRevision != revisionOf(settingsRaw) || saved.BoardRevision != revisionOf(boardRaw) {
		t.Error("combined save revisions do not match both written files")
	}
	if len(saved.Board) != 1 || saved.Board[0].TitleEN != "Newest existing post" {
		t.Fatalf("combined save board response = %#v", saved.Board)
	}
}

func TestStageBoardMediaUsesImageMagicAndDoesNotTouchRepository(t *testing.T) {
	root := newBoardRepoFixture(t)
	app := &App{repoRoot: root}
	t.Cleanup(func() { app.shutdown(context.Background()) })
	inputDirectory := t.TempDir()
	pngPath := filepath.Join(inputDirectory, "png-magic-with-wrong-extension.txt")
	jpegPath := filepath.Join(inputDirectory, "jpeg-magic-without-extension")
	fakePath := filepath.Join(inputDirectory, "fake-image.png")
	writeFixtureFile(t, pngPath, fixturePNGBytes())
	writeFixtureFile(t, jpegPath, fixtureJPEGBytes())
	writeFixtureFile(t, fakePath, []byte("not an image"))
	before := snapshotBoardRepo(t, root)

	response, err := app.StageBoardMedia([]string{pngPath, jpegPath, fakePath})
	if err != nil {
		t.Fatalf("StageBoardMedia() unexpected batch error: %v", err)
	}
	if len(response.Items) != 2 {
		t.Fatalf("StageBoardMedia() accepted %d files, want PNG and JPEG only; rejected: %#v", len(response.Items), response.Rejected)
	}
	gotMIMEs := map[string]bool{}
	tokens := make([]string, 0, len(response.Items))
	for _, item := range response.Items {
		gotMIMEs[item.MIMEType] = true
		if item.StageToken == "" {
			t.Error("accepted staged image has an empty token")
		}
		if !strings.HasPrefix(item.PreviewURL, "data:image/jpeg;base64,") {
			t.Errorf("preview URL for %q is not a JPEG thumbnail", item.OriginalName)
		}
		tokens = append(tokens, item.StageToken)
	}
	if !gotMIMEs["image/png"] || !gotMIMEs["image/jpeg"] {
		t.Errorf("detected MIME types = %#v, want image/png and image/jpeg", gotMIMEs)
	}
	if len(response.Rejected) != 1 || response.Rejected[0].OriginalName != filepath.Base(fakePath) {
		t.Errorf("rejected media = %#v, want only fake-image.png", response.Rejected)
	}
	assertBoardRepoSnapshot(t, root, before, "StageBoardMedia")

	app.DiscardBoardMedia(tokens)
	assertBoardRepoSnapshot(t, root, before, "DiscardBoardMedia")
	if app.stagingDir != "" || len(app.boardMedia) != 0 {
		t.Errorf("DiscardBoardMedia() left staging state: dir=%q media=%d", app.stagingDir, len(app.boardMedia))
	}
}

func TestSaveEditorDataPublishesUniqueStagedImageAndBoardReference(t *testing.T) {
	root := newBoardRepoFixture(t)
	mediaDirectory := filepath.Join(root, "data", "media")
	collisionPath := filepath.Join(mediaDirectory, "board-20260814-new-post-01.png")
	collisionContents := []byte("existing file must not be overwritten")
	writeFixtureFile(t, collisionPath, collisionContents)
	app := &App{repoRoot: root}
	t.Cleanup(func() { app.shutdown(context.Background()) })
	loaded, err := app.LoadEditorData()
	if err != nil {
		t.Fatalf("LoadEditorData() unexpected error: %v", err)
	}

	inputPath := filepath.Join(t.TempDir(), "photo.bin")
	pngContents := fixturePNGBytes()
	writeFixtureFile(t, inputPath, pngContents)
	beforeStage := snapshotBoardRepo(t, root)
	staged, err := app.StageBoardMedia([]string{inputPath})
	if err != nil {
		t.Fatalf("StageBoardMedia() unexpected error: %v", err)
	}
	if len(staged.Items) != 1 || len(staged.Rejected) != 0 {
		t.Fatalf("StageBoardMedia() = %#v, want one accepted image", staged)
	}
	assertBoardRepoSnapshot(t, root, beforeStage, "StageBoardMedia before save")

	rows := make([]BoardSaveItem, 0, len(loaded.Board)+1)
	for _, item := range loaded.Board {
		rows = append(rows, BoardSaveItem{EditorKey: item.EditorKey})
	}
	rows = append(rows, BoardSaveItem{NewPost: &NewBoardPost{
		StartDate: "2026-08-14",
		EndDate:   "",
		TitleEN:   "New Post",
		TitleKO:   "새 게시글",
		ContentEN: "New post with a staged image",
		ContentKO: "새 이미지가 있는 게시글",
		Media: []NewBoardMedia{{
			StageToken: staged.Items[0].StageToken,
			CaptionEN:  "English caption",
			CaptionKO:  "국문 캡션",
		}},
	}})

	if _, err := app.SaveEditorData(boardOnlySaveRequest(loaded, rows)); err != nil {
		t.Fatalf("SaveEditorData() unexpected error: %v", err)
	}

	wantName := "board-20260814-new-post-01-2.png"
	writtenMedia := filepath.Join(mediaDirectory, wantName)
	gotContents, err := os.ReadFile(writtenMedia)
	if err != nil {
		t.Fatalf("read uniquely published media %q: %v", wantName, err)
	}
	if !bytes.Equal(gotContents, pngContents) {
		t.Fatalf("published media bytes differ\ngot:  %x\nwant: %x", gotContents, pngContents)
	}
	if got, err := os.ReadFile(collisionPath); err != nil || !bytes.Equal(got, collisionContents) {
		t.Fatalf("existing collision target was changed: bytes=%q err=%v", got, err)
	}

	items := readBoardFixtureItems(t, root)
	if items[0].TitleEN != "New Post" || len(items[0].Media) != 1 {
		t.Fatalf("saved newest board item = %#v, want new post with one image", items[0])
	}
	if items[0].Media[0].Src != wantName {
		t.Errorf("saved media src = %q, want %q", items[0].Media[0].Src, wantName)
	}
	if items[0].Media[0].CaptionEN != "English caption" || items[0].Media[0].CaptionKO != "국문 캡션" {
		t.Errorf("saved media captions = %#v", items[0].Media[0])
	}
}

func TestSaveEditorDataEditsExistingPostAndMediaLosslessly(t *testing.T) {
	root := newBoardRepoFixture(t)
	app := &App{repoRoot: root}
	t.Cleanup(func() { app.shutdown(context.Background()) })
	loaded, err := app.LoadEditorData()
	if err != nil {
		t.Fatalf("LoadEditorData() unexpected error: %v", err)
	}
	apiItems := boardAPIItems(t, loaded)
	existing := boardAPIItemByTitle(t, apiItems, "Newest existing post")
	if existing.ContentEN != "Newer content" || existing.ContentKO != "최신 본문" {
		t.Fatalf("existing post content was not exposed for editing: %#v", existing)
	}
	if len(existing.Media) != 1 || existing.Media[0].EditorKey == "" {
		t.Fatalf("existing media was not exposed with an opaque editor key: %#v", existing.Media)
	}
	if existing.Media[0].Src != "existing-board-photo.jpg" || existing.Media[0].PreviewURL == "" {
		t.Fatalf("existing media response lacks source or preview: %#v", existing.Media[0])
	}
	if existing.Media[0].OriginalName != "existing-board-photo.jpg" || existing.Media[0].Size <= 0 {
		t.Fatalf("existing media response lacks original name or file size: %#v", existing.Media[0])
	}

	inputPath := filepath.Join(t.TempDir(), "new-existing-post-photo.dat")
	pngContents := fixturePNGBytes()
	writeFixtureFile(t, inputPath, pngContents)
	staged, err := app.StageBoardMedia([]string{inputPath})
	if err != nil {
		t.Fatalf("StageBoardMedia() unexpected error: %v", err)
	}
	if len(staged.Items) != 1 || len(staged.Rejected) != 0 {
		t.Fatalf("StageBoardMedia() = %#v, want one accepted image", staged)
	}

	rows := make([]BoardSaveItem, 0, len(apiItems))
	for _, item := range apiItems {
		if item.EditorKey != existing.EditorKey {
			rows = append(rows, BoardSaveItem{EditorKey: item.EditorKey})
			continue
		}
		rows = append(rows, boardSaveItemFromJSON(t, map[string]any{
			"editor_key": item.EditorKey,
			"post": map[string]any{
				"start_date": "2028-01-15",
				"end_date":   "2028-01-17",
				"title_en":   "Edited existing post",
				"title_ko":   "수정된 기존 게시글",
				"content_en": "Edited English content",
				"content_ko": "수정된 국문 본문",
				"media": []map[string]any{
					{
						"editor_key": existing.Media[0].EditorKey,
						"caption_en": "Updated existing caption",
						"caption_ko": "",
					},
					{
						"stage_token": staged.Items[0].StageToken,
						"caption_en":  "New staged caption",
						"caption_ko":  "새 사진 캡션",
					},
				},
			},
		}))
	}

	if _, err := app.SaveEditorData(boardOnlySaveRequest(loaded, rows)); err != nil {
		t.Fatalf("SaveEditorData() editing existing post unexpected error: %v", err)
	}

	items := readBoardFixtureItems(t, root)
	if items[0].TitleEN != "Edited existing post" {
		t.Fatalf("edited post was not re-sorted to newest position: %#v", items)
	}
	edited := items[0]
	if edited.StartDate != "2028-01-15" || edited.EndDate != "2028-01-17" {
		t.Errorf("edited dates = %q to %q", edited.StartDate, edited.EndDate)
	}
	if edited.TitleKO != "수정된 기존 게시글" || edited.ContentEN != "Edited English content" || edited.ContentKO != "수정된 국문 본문" {
		t.Errorf("edited text fields were not stored: %#v", edited)
	}
	if len(edited.Media) != 2 {
		t.Fatalf("edited media count = %d, want retained and staged media", len(edited.Media))
	}
	retained := edited.Media[0]
	if retained.Src != "existing-board-photo.jpg" || retained.Type != "image" || retained.Poster != "existing-board-poster.jpg" {
		t.Errorf("retained media identity fields changed: %#v", retained)
	}
	if retained.CaptionEN != "Updated existing caption" || retained.CaptionKO != "" {
		t.Errorf("retained media caption edit/removal was not stored: %#v", retained)
	}
	added := edited.Media[1]
	if added.Src == "" || added.Src == retained.Src {
		t.Fatalf("new staged media reference = %q, want a new repository path", added.Src)
	}
	if added.CaptionEN != "New staged caption" || added.CaptionKO != "새 사진 캡션" {
		t.Errorf("new staged media captions = %#v", added)
	}
	addedBytes, err := os.ReadFile(filepath.Join(root, "data", "media", filepath.FromSlash(added.Src)))
	if err != nil {
		t.Fatalf("read newly published media %q: %v", added.Src, err)
	}
	if !bytes.Equal(addedBytes, pngContents) {
		t.Fatalf("newly published media bytes differ\ngot:  %x\nwant: %x", addedBytes, pngContents)
	}

	rawEdited := boardRawItemByTitle(t, root, "Edited existing post")
	assertRawJSONField(t, rawEdited, "future_row_field", map[string]any{
		"category": "fixture",
		"rank":     float64(7),
	})
	var rawMedia []map[string]json.RawMessage
	if err := json.Unmarshal(rawEdited["media"], &rawMedia); err != nil {
		t.Fatalf("decode edited raw media: %v", err)
	}
	assertRawJSONField(t, rawMedia[0], "future_media_field", map[string]any{
		"credit": "fixture photographer",
		"rights": []any{"retain", "verbatim"},
	})
}

func TestSaveEditorDataFullUnchangedPayloadPreservesOmittedMediaFields(t *testing.T) {
	root := newBoardRepoFixture(t)
	boardPath := filepath.Join(root, "data", "board.json")
	var rows []map[string]json.RawMessage
	if err := json.Unmarshal(readBoardFixtureFile(t, root), &rows); err != nil {
		t.Fatalf("decode board fixture: %v", err)
	}
	found := false
	for _, row := range rows {
		var title string
		if err := json.Unmarshal(row["title_en"], &title); err != nil {
			t.Fatalf("decode board title: %v", err)
		}
		if title != "Newest existing post" {
			continue
		}
		var media []map[string]json.RawMessage
		if err := json.Unmarshal(row["media"], &media); err != nil || len(media) != 1 {
			t.Fatalf("decode existing media: %v; media=%#v", err, media)
		}
		delete(media[0], "caption_ko")
		encodedMedia, err := json.Marshal(media)
		if err != nil {
			t.Fatalf("encode existing media: %v", err)
		}
		row["media"] = encodedMedia
		found = true
	}
	if !found {
		t.Fatal("newest fixture row not found")
	}
	encoded, err := json.MarshalIndent(rows, "", "  ")
	if err != nil {
		t.Fatalf("encode board fixture: %v", err)
	}
	writeFixtureFile(t, boardPath, append(encoded, '\n'))

	app := &App{repoRoot: root}
	loaded, err := app.LoadEditorData()
	if err != nil {
		t.Fatalf("LoadEditorData() unexpected error: %v", err)
	}
	apiItems := boardAPIItems(t, loaded)
	requestRows := make([]BoardSaveItem, 0, len(apiItems))
	for _, item := range apiItems {
		media := make([]BoardMediaInput, 0, len(item.Media))
		for _, currentMedia := range item.Media {
			media = append(media, BoardMediaInput{
				EditorKey: currentMedia.EditorKey,
				CaptionEN: currentMedia.CaptionEN,
				CaptionKO: currentMedia.CaptionKO,
			})
		}
		requestRows = append(requestRows, BoardSaveItem{
			EditorKey: item.EditorKey,
			Post: &BoardPostInput{
				StartDate: item.StartDate,
				EndDate:   item.EndDate,
				TitleEN:   item.TitleEN,
				TitleKO:   item.TitleKO,
				ContentEN: item.ContentEN,
				ContentKO: item.ContentKO,
				Media:     media,
			},
		})
	}
	if _, err := app.SaveEditorData(boardOnlySaveRequest(loaded, requestRows)); err != nil {
		t.Fatalf("SaveEditorData() unchanged full payload unexpected error: %v", err)
	}

	rawSaved := boardRawItemByTitle(t, root, "Newest existing post")
	var savedMedia []map[string]json.RawMessage
	if err := json.Unmarshal(rawSaved["media"], &savedMedia); err != nil || len(savedMedia) != 1 {
		t.Fatalf("decode saved media: %v; media=%#v", err, savedMedia)
	}
	if _, exists := savedMedia[0]["caption_ko"]; exists {
		t.Fatal("unchanged full post save introduced an omitted caption_ko field")
	}
	assertRawJSONField(t, rawSaved, "future_row_field", map[string]any{
		"category": "fixture",
		"rank":     float64(7),
	})
	assertRawJSONField(t, savedMedia[0], "future_media_field", map[string]any{
		"credit": "fixture photographer",
		"rights": []any{"retain", "verbatim"},
	})
}

func TestSaveEditorDataCanRemoveExistingMediaWithoutDeletingItsFiles(t *testing.T) {
	root := newBoardRepoFixture(t)
	photoPath := filepath.Join(root, "data", "media", "existing-board-photo.jpg")
	posterPath := filepath.Join(root, "data", "media", "existing-board-poster.jpg")
	wantPhoto := readFixtureFile(t, photoPath)
	wantPoster := readFixtureFile(t, posterPath)
	app := &App{repoRoot: root}
	loaded, err := app.LoadEditorData()
	if err != nil {
		t.Fatalf("LoadEditorData() unexpected error: %v", err)
	}
	apiItems := boardAPIItems(t, loaded)
	existing := boardAPIItemByTitle(t, apiItems, "Newest existing post")

	rows := make([]BoardSaveItem, 0, len(apiItems))
	for _, item := range apiItems {
		if item.EditorKey != existing.EditorKey {
			rows = append(rows, BoardSaveItem{EditorKey: item.EditorKey})
			continue
		}
		rows = append(rows, boardSaveItemFromJSON(t, map[string]any{
			"editor_key": item.EditorKey,
			"post": map[string]any{
				"start_date": item.StartDate,
				"end_date":   item.EndDate,
				"title_en":   item.TitleEN,
				"title_ko":   item.TitleKO,
				"content_en": item.ContentEN,
				"content_ko": item.ContentKO,
				"media":      []any{},
			},
		}))
	}

	if _, err := app.SaveEditorData(boardOnlySaveRequest(loaded, rows)); err != nil {
		t.Fatalf("SaveEditorData() removing existing media unexpected error: %v", err)
	}
	items := readBoardFixtureItems(t, root)
	var saved *boardDocumentItem
	for index := range items {
		if items[index].TitleEN == "Newest existing post" {
			saved = &items[index]
			break
		}
	}
	if saved == nil {
		t.Fatal("edited existing post is missing after save")
	}
	if saved.Media == nil || len(saved.Media) != 0 {
		t.Fatalf("saved existing media = %#v, want an explicit empty array", saved.Media)
	}
	if got := readFixtureFile(t, photoPath); !bytes.Equal(got, wantPhoto) {
		t.Fatal("removing a media reference changed or deleted its shared source file")
	}
	if got := readFixtureFile(t, posterPath); !bytes.Equal(got, wantPoster) {
		t.Fatal("removing a media reference changed or deleted its poster file")
	}
	rawSaved := boardRawItemByTitle(t, root, "Newest existing post")
	assertRawJSONField(t, rawSaved, "future_row_field", map[string]any{
		"category": "fixture",
		"rank":     float64(7),
	})
}

func TestUnsavedExistingBoardDraftDiscardAndCloseLeaveRepositoryUntouched(t *testing.T) {
	for _, operation := range []string{"discard", "close"} {
		t.Run(operation, func(t *testing.T) {
			root := newBoardRepoFixture(t)
			app := &App{repoRoot: root}
			loaded, err := app.LoadEditorData()
			if err != nil {
				t.Fatalf("LoadEditorData() unexpected error: %v", err)
			}
			existing := boardAPIItemByTitle(t, boardAPIItems(t, loaded), "Newest existing post")
			inputPath := filepath.Join(t.TempDir(), "unsaved-photo.png")
			writeFixtureFile(t, inputPath, fixturePNGBytes())
			before := snapshotBoardRepo(t, root)
			staged, err := app.StageBoardMedia([]string{inputPath})
			if err != nil || len(staged.Items) != 1 {
				t.Fatalf("StageBoardMedia() = %#v, %v; want one staged image", staged, err)
			}

			// Model a fully edited in-memory draft without invoking SaveEditorData.
			// The backend must not publish either its fields or staged media merely
			// because dirty state is set, discarded, or the application shuts down.
			_ = boardSaveItemFromJSON(t, map[string]any{
				"editor_key": existing.EditorKey,
				"post": map[string]any{
					"start_date": "2031-01-01",
					"end_date":   "",
					"title_en":   "Unsaved edit",
					"title_ko":   "저장되지 않은 수정",
					"content_en": "Unsaved content",
					"content_ko": "저장되지 않은 본문",
					"media": []map[string]any{{
						"stage_token": staged.Items[0].StageToken,
						"caption_en":  "Unsaved caption",
						"caption_ko":  "저장되지 않은 캡션",
					}},
				},
			})
			app.SetDirty(true)

			if operation == "discard" {
				if err := app.DiscardBoardMedia([]string{staged.Items[0].StageToken}); err != nil {
					t.Fatalf("DiscardBoardMedia() unexpected error: %v", err)
				}
				app.SetDirty(false)
			} else {
				app.shutdown(context.Background())
			}
			assertBoardRepoSnapshot(t, root, before, "unsaved board "+operation)
			if app.stagingDir != "" || len(app.boardMedia) != 0 {
				t.Errorf("%s left staging state: dir=%q media=%d", operation, app.stagingDir, len(app.boardMedia))
			}
		})
	}
}

func TestSaveEditorDataRejectsInvalidBoardDatesWithoutWriting(t *testing.T) {
	tests := []struct {
		name      string
		startDate string
		endDate   string
	}{
		{name: "noncanonical start", startDate: "2026-2-03", endDate: ""},
		{name: "impossible start", startDate: "2026-02-30", endDate: ""},
		{name: "noncanonical end", startDate: "2026-02-03", endDate: "2026-2-04"},
		{name: "end before start", startDate: "2026-02-03", endDate: "2026-02-02"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := newBoardRepoFixture(t)
			app := &App{repoRoot: root}
			loaded, err := app.LoadEditorData()
			if err != nil {
				t.Fatalf("LoadEditorData() unexpected error: %v", err)
			}
			before := snapshotBoardRepo(t, root)
			request := boardOnlySaveRequest(loaded, []BoardSaveItem{{NewPost: &NewBoardPost{
				StartDate: test.startDate,
				EndDate:   test.endDate,
				TitleEN:   "Invalid date post",
				TitleKO:   "잘못된 날짜",
				ContentEN: "",
				ContentKO: "",
				Media:     []NewBoardMedia{},
			}}})

			if _, err := app.SaveEditorData(request); err == nil {
				t.Fatal("SaveEditorData() invalid date error = nil, want rejection")
			}
			assertBoardRepoSnapshot(t, root, before, "invalid-date SaveEditorData")
		})
	}
}

func TestSaveEditorDataRejectsUnknownStageTokenWithoutWriting(t *testing.T) {
	root := newBoardRepoFixture(t)
	app := &App{repoRoot: root}
	loaded, err := app.LoadEditorData()
	if err != nil {
		t.Fatalf("LoadEditorData() unexpected error: %v", err)
	}
	before := snapshotBoardRepo(t, root)
	request := boardOnlySaveRequest(loaded, []BoardSaveItem{{NewPost: &NewBoardPost{
		StartDate: "2026-08-14",
		EndDate:   "",
		TitleEN:   "Unknown token post",
		TitleKO:   "잘못된 토큰",
		ContentEN: "",
		ContentKO: "",
		Media: []NewBoardMedia{{
			StageToken: "not-a-staged-token",
		}},
	}}})

	if _, err := app.SaveEditorData(request); err == nil {
		t.Fatal("SaveEditorData() unknown stage token error = nil, want rejection")
	}
	assertBoardRepoSnapshot(t, root, before, "unknown-token SaveEditorData")
}

func newBoardRepoFixture(t *testing.T) string {
	t.Helper()
	root := newTempRepoFixture(t, fixtureSettingsJSON())
	mediaDirectory := filepath.Join(root, "data", "media")
	if err := os.MkdirAll(mediaDirectory, 0o755); err != nil {
		t.Fatalf("create board media fixture directory: %v", err)
	}
	writeFixtureFile(t, filepath.Join(mediaDirectory, "existing-board-photo.jpg"), fixtureJPEGBytes())
	writeFixtureFile(t, filepath.Join(mediaDirectory, "existing-board-poster.jpg"), fixtureJPEGBytes())
	writeFixtureFile(t, filepath.Join(root, "data", "board.json"), fixtureBoardJSON())
	return root
}

func fixtureBoardJSON() []byte {
	// Deliberately not date-sorted: LoadEditorData and SaveEditorData must expose
	// the same newest-first ordering as the website, independent of file order.
	return []byte(`[
  {
    "start_date": "2024-01-02",
    "end_date": "",
    "title_en": "Oldest existing post",
    "title_ko": "가장 오래된 게시글",
    "content_en": "Old content",
    "content_ko": "오래된 본문",
    "media": [],
    "future_field": {"preserve": true}
  },
  {
    "start_date": "2026-06-10",
    "end_date": "2026-06-11",
    "title_en": "Newest existing post",
    "title_ko": "가장 최신 게시글",
    "content_en": "Newer content",
    "content_ko": "최신 본문",
    "media": [
      {
        "src": "existing-board-photo.jpg",
        "type": "image",
        "poster": "existing-board-poster.jpg",
        "caption_en": "Existing photo",
        "caption_ko": "기존 사진",
        "future_media_field": {
          "credit": "fixture photographer",
          "rights": ["retain", "verbatim"]
        }
      }
    ],
    "future_row_field": {
      "category": "fixture",
      "rank": 7
    }
  },
  {
    "start_date": "2025-04-03",
    "end_date": "2025-04-05",
    "title_en": "Post to delete",
    "title_ko": "삭제할 게시글",
    "content_en": "Delete this content",
    "content_ko": "삭제할 본문",
    "media": []
  }
]
`)
}

func fixturePNGBytes() []byte {
	var encoded bytes.Buffer
	picture := image.NewRGBA(image.Rect(0, 0, 2, 2))
	picture.Set(0, 0, color.RGBA{R: 32, G: 96, B: 192, A: 255})
	_ = png.Encode(&encoded, picture)
	return encoded.Bytes()
}

func fixtureJPEGBytes() []byte {
	var encoded bytes.Buffer
	picture := image.NewRGBA(image.Rect(0, 0, 2, 2))
	picture.Set(0, 0, color.RGBA{R: 192, G: 96, B: 32, A: 255})
	_ = jpeg.Encode(&encoded, picture, nil)
	return encoded.Bytes()
}

type boardAPIMediaFixture struct {
	EditorKey    string `json:"editor_key"`
	Src          string `json:"src"`
	Type         string `json:"type"`
	Poster       string `json:"poster"`
	CaptionEN    string `json:"caption_en"`
	CaptionKO    string `json:"caption_ko"`
	OriginalName string `json:"original_name"`
	Size         int64  `json:"size"`
	PreviewURL   string `json:"preview_url"`
}

type boardAPIItemFixture struct {
	EditorKey string                 `json:"editor_key"`
	StartDate string                 `json:"start_date"`
	EndDate   string                 `json:"end_date"`
	TitleEN   string                 `json:"title_en"`
	TitleKO   string                 `json:"title_ko"`
	ContentEN string                 `json:"content_en"`
	ContentKO string                 `json:"content_ko"`
	Media     []boardAPIMediaFixture `json:"media"`
}

func boardAPIItems(t *testing.T, response EditorDataResponse) []boardAPIItemFixture {
	t.Helper()
	raw, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("encode EditorDataResponse: %v", err)
	}
	var envelope struct {
		Board []boardAPIItemFixture `json:"board"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		t.Fatalf("decode board API response: %v", err)
	}
	return envelope.Board
}

func boardAPIItemByTitle(t *testing.T, items []boardAPIItemFixture, title string) boardAPIItemFixture {
	t.Helper()
	for _, item := range items {
		if item.TitleEN == title {
			return item
		}
	}
	t.Fatalf("board API item %q not found in %#v", title, items)
	return boardAPIItemFixture{}
}

func boardSaveItemFromJSON(t *testing.T, value any) BoardSaveItem {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("encode board save item fixture: %v", err)
	}
	var item BoardSaveItem
	if err := json.Unmarshal(raw, &item); err != nil {
		t.Fatalf("decode board save item fixture: %v", err)
	}
	return item
}

func boardOnlySaveRequest(loaded EditorDataResponse, rows []BoardSaveItem) SaveEditorDataRequest {
	return SaveEditorDataRequest{
		Settings:                   loaded.Settings,
		SettingsRevision:           loaded.SettingsRevision,
		ProfileRevision:            loaded.ProfileRevision,
		SaveSettings:               false,
		PeopleRevision:             loaded.PeopleRevision,
		AwardsRevision:             loaded.AwardsRevision,
		AcademicActivitiesRevision: loaded.AcademicActivitiesRevision,
		PublicationsRevision:       loaded.PublicationsRevision,
		ProjectsRevision:           loaded.ProjectsRevision,
		SoftwareRevision:           loaded.SoftwareRevision,
		Board:                      rows,
		BoardRevision:              loaded.BoardRevision,
		SaveBoard:                  true,
	}
}

func boardKeysByTitle(items []BoardItemSummary) map[string]string {
	result := make(map[string]string, len(items))
	for _, item := range items {
		result[item.TitleEN] = item.EditorKey
	}
	return result
}

func readBoardFixtureItems(t *testing.T, root string) []boardDocumentItem {
	t.Helper()
	raw := readBoardFixtureFile(t, root)
	var items []boardDocumentItem
	if err := json.Unmarshal(raw, &items); err != nil {
		t.Fatalf("decode board fixture: %v", err)
	}
	return items
}

func readBoardFixtureFile(t *testing.T, root string) []byte {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(root, "data", "board.json"))
	if err != nil {
		t.Fatalf("read board fixture: %v", err)
	}
	return raw
}

func readFixtureFile(t *testing.T, path string) []byte {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture file %s: %v", path, err)
	}
	return raw
}

func boardRawItemByTitle(t *testing.T, root, title string) map[string]json.RawMessage {
	t.Helper()
	var items []map[string]json.RawMessage
	if err := json.Unmarshal(readBoardFixtureFile(t, root), &items); err != nil {
		t.Fatalf("decode raw board items: %v", err)
	}
	for _, item := range items {
		var itemTitle string
		if err := json.Unmarshal(item["title_en"], &itemTitle); err != nil {
			t.Fatalf("decode raw board title: %v", err)
		}
		if itemTitle == title {
			return item
		}
	}
	t.Fatalf("raw board item %q not found", title)
	return nil
}

func assertRawJSONField(t *testing.T, fields map[string]json.RawMessage, name string, want any) {
	t.Helper()
	raw, exists := fields[name]
	if !exists {
		t.Fatalf("unknown JSON field %q was dropped", name)
	}
	var got any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("decode JSON field %q: %v", name, err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("JSON field %q changed\ngot:  %#v\nwant: %#v", name, got, want)
	}
}

func writeFixtureFile(t *testing.T, path string, contents []byte) {
	t.Helper()
	if err := os.WriteFile(path, contents, 0o644); err != nil {
		t.Fatalf("write fixture file %s: %v", path, err)
	}
}

func snapshotBoardRepo(t *testing.T, root string) map[string][]byte {
	t.Helper()
	snapshot := make(map[string][]byte)
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		contents, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		snapshot[filepath.ToSlash(relative)] = contents
		return nil
	})
	if err != nil {
		t.Fatalf("snapshot board fixture repository: %v", err)
	}
	return snapshot
}

func assertBoardRepoSnapshot(t *testing.T, root string, want map[string][]byte, operation string) {
	t.Helper()
	got := snapshotBoardRepo(t, root)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%s changed repository files\ngot:  %#v\nwant: %#v", operation, got, want)
	}
}
