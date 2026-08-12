package main

import (
	"encoding/json"
	"fmt"
	"github.com/Gonie-Gonie/Gonie-Gonie.github.io/editor/internal/dataversion"
	"os"
	pathpkg "path"
	"path/filepath"
	"strings"
)

const (
	maxProjects          = 500
	maxProjectNotes      = 100
	maxProjectNoteLength = 5_000
)

// ProjectItemSummary is the editable, flat project shape returned to Wails.
// EditorKey and the media EditorKeys are opaque snapshot handles, not data IDs.
type ProjectItemSummary struct {
	EditorKey   string                `json:"editor_key"`
	VisibleInCV bool                  `json:"visible_in_CV"`
	TitleEN     string                `json:"title_en"`
	TitleKO     string                `json:"title_ko"`
	StartDate   string                `json:"start_date"`
	EndDate     string                `json:"end_date"`
	Theme       string                `json:"theme"`
	FunderEN    string                `json:"funder_en"`
	FunderKO    string                `json:"funder_ko"`
	NotesEN     []string              `json:"notes_en"`
	NotesKR     []string              `json:"notes_kr"`
	Media       []ProjectMediaSummary `json:"media"`
	MediaCount  int                   `json:"media_count"`
}

type ProjectMediaSummary struct {
	EditorKey    string `json:"editor_key"`
	Src          string `json:"src"`
	Type         string `json:"type,omitempty"`
	Poster       string `json:"poster,omitempty"`
	CaptionEN    string `json:"caption_en"`
	CaptionKO    string `json:"caption_ko"`
	OriginalName string `json:"original_name"`
	Size         int64  `json:"size"`
	PreviewURL   string `json:"preview_url,omitempty"`
}

type ProjectMediaInput struct {
	EditorKey  string `json:"editor_key,omitempty"`
	StageToken string `json:"stage_token,omitempty"`
	CaptionEN  string `json:"caption_en"`
	CaptionKO  string `json:"caption_ko"`
}

type ProjectInput struct {
	TitleEN   string              `json:"title_en"`
	TitleKO   string              `json:"title_ko"`
	StartDate string              `json:"start_date"`
	EndDate   string              `json:"end_date"`
	Theme     string              `json:"theme"`
	FunderEN  string              `json:"funder_en"`
	FunderKO  string              `json:"funder_ko"`
	NotesEN   []string            `json:"notes_en"`
	NotesKR   []string            `json:"notes_kr"`
	Media     []ProjectMediaInput `json:"media"`
}

// Existing projects use EditorKey plus Project. New projects leave EditorKey
// blank. Omitting an existing row from the request deletes only its JSON row;
// referenced files in data/media are deliberately never deleted.
type ProjectSaveItem struct {
	EditorKey   string        `json:"editor_key,omitempty"`
	Project     *ProjectInput `json:"project,omitempty"`
	VisibleInCV *bool         `json:"visible_in_CV,omitempty"`
}

type projectDocumentMedia struct {
	Src       string `json:"src"`
	Type      string `json:"type,omitempty"`
	Poster    string `json:"poster,omitempty"`
	CaptionEN string `json:"caption_en,omitempty"`
	CaptionKO string `json:"caption_ko,omitempty"`

	editorKey    string
	raw          json.RawMessage
	originalName string
	size         int64
	previewURL   string
}

type projectDocumentItem struct {
	TitleEN   string                 `json:"title_en"`
	TitleKO   string                 `json:"title_ko"`
	StartDate string                 `json:"start_date"`
	EndDate   string                 `json:"end_date"`
	Theme     string                 `json:"theme"`
	FunderEN  string                 `json:"funder_en"`
	FunderKO  string                 `json:"funder_ko"`
	NotesEN   []string               `json:"notes_en"`
	NotesKR   []string               `json:"notes_kr"`
	Media     []projectDocumentMedia `json:"media"`
}

type projectRow struct {
	Key  string
	Raw  json.RawMessage
	Item projectDocumentItem
}

type projectSnapshot struct {
	Raw  []byte
	Rows []projectRow
}

func projectThemeIDs(settings SettingsDocument) map[string]bool {
	ids := make(map[string]bool, len(settings.Topics))
	for _, theme := range settings.Topics {
		ids[theme.ID] = true
	}
	return ids
}

func readProjects(path, root string, themeIDs map[string]bool) (projectSnapshot, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return projectSnapshot{}, fmt.Errorf("projects.json 읽기 실패: %w", err)
	}
	return readProjectsBytes(raw, root, themeIDs)
}

func readProjectsBytes(raw []byte, root string, themeIDs map[string]bool) (projectSnapshot, error) {
	rows, err := dataversion.Array(raw, "projects.json")
	if err != nil {
		return projectSnapshot{}, err
	}

	snapshot := projectSnapshot{
		Raw:  append([]byte(nil), raw...),
		Rows: make([]projectRow, 0, len(rows)),
	}
	for index, rowRaw := range rows {
		item, err := parseProjectItem(
			rowRaw,
			root,
			themeIDs,
			fmt.Sprintf("projects.json %d번째 프로젝트", index+1),
		)
		if err != nil {
			return projectSnapshot{}, err
		}
		if _, err := visibleInCV(rowRaw, fmt.Sprintf("projects.json %d번째 프로젝트", index+1)); err != nil {
			return projectSnapshot{}, err
		}
		hash := revisionOf(rowRaw)
		snapshot.Rows = append(snapshot.Rows, projectRow{
			Key:  fmt.Sprintf("%d:%s", index, hash[:20]),
			Raw:  cloneRawMessage(rowRaw),
			Item: item,
		})
	}
	return snapshot, nil
}

func parseProjectItem(raw json.RawMessage, root string, themeIDs map[string]bool, contextLabel string) (projectDocumentItem, error) {
	fields, err := decodeObject(raw)
	if err != nil {
		return projectDocumentItem{}, fmt.Errorf("%s는 JSON 객체여야 합니다", contextLabel)
	}
	for _, forbidden := range []string{"id", "period", "summary", "summary_en", "summary_ko", "photos"} {
		if _, exists := fields[forbidden]; exists {
			return projectDocumentItem{}, fmt.Errorf("%s에는 사용하지 않는 %s 필드가 있습니다", contextLabel, forbidden)
		}
	}
	for _, required := range []string{"start_date", "end_date", "theme", "notes_en", "notes_kr", "media"} {
		if _, exists := fields[required]; !exists {
			return projectDocumentItem{}, fmt.Errorf("%s의 %s 필드가 없습니다", contextLabel, required)
		}
	}

	var item projectDocumentItem
	if err := json.Unmarshal(raw, &item); err != nil {
		return projectDocumentItem{}, fmt.Errorf("%s 필드 형식이 올바르지 않습니다: %w", contextLabel, err)
	}
	if item.NotesEN == nil || item.NotesKR == nil {
		return projectDocumentItem{}, fmt.Errorf("%s의 notes_en과 notes_kr는 JSON 배열이어야 합니다", contextLabel)
	}
	if err := validateLoadedProject(item, themeIDs, contextLabel); err != nil {
		return projectDocumentItem{}, err
	}

	var mediaRaw []json.RawMessage
	if err := json.Unmarshal(fields["media"], &mediaRaw); err != nil || len(mediaRaw) != len(item.Media) {
		return projectDocumentItem{}, fmt.Errorf("%s의 media 항목을 분석할 수 없습니다", contextLabel)
	}
	for index, media := range item.Media {
		mediaContext := fmt.Sprintf("%s 미디어 %d", contextLabel, index+1)
		mediaFields, err := decodeObject(mediaRaw[index])
		if err != nil {
			return projectDocumentItem{}, fmt.Errorf("%s는 JSON 객체여야 합니다", mediaContext)
		}
		if _, exists := mediaFields["caption_kr"]; exists {
			return projectDocumentItem{}, fmt.Errorf("%s는 caption_kr 대신 caption_ko를 사용해야 합니다", mediaContext)
		}
		if _, exists := mediaFields["caption-en"]; exists {
			return projectDocumentItem{}, fmt.Errorf("%s는 caption-en 대신 caption_en을 사용해야 합니다", mediaContext)
		}
		if media.Src == "" {
			return projectDocumentItem{}, fmt.Errorf("%s의 src가 비어 있습니다", mediaContext)
		}
		mediaType := strings.ToLower(strings.TrimSpace(media.Type))
		if mediaType != "" && mediaType != "image" && mediaType != "video" {
			return projectDocumentItem{}, fmt.Errorf("%s의 type은 image 또는 video여야 합니다", mediaContext)
		}
		if err := validateProjectCaptionPair(media.CaptionEN, media.CaptionKO, mediaContext, true); err != nil {
			return projectDocumentItem{}, err
		}
		resolvedSource, sourceInfo, err := resolveBoardMediaPath(root, media.Src, mediaContext)
		if err != nil {
			return projectDocumentItem{}, err
		}
		previewURL := boardThumbnailDataURL(resolvedSource)
		if media.Poster != "" {
			resolvedPoster, _, err := resolveBoardMediaPath(root, media.Poster, mediaContext+" poster")
			if err != nil {
				return projectDocumentItem{}, err
			}
			if previewURL == "" {
				previewURL = boardThumbnailDataURL(resolvedPoster)
			}
		}
		media.editorKey = fmt.Sprintf("%d:%s", index, revisionOf(mediaRaw[index])[:20])
		media.raw = cloneRawMessage(mediaRaw[index])
		media.originalName = pathpkg.Base(media.Src)
		media.size = sourceInfo.Size()
		media.previewURL = previewURL
		item.Media[index] = media
	}
	return item, nil
}

func validateLoadedProject(item projectDocumentItem, themeIDs map[string]bool, contextLabel string) error {
	if err := validateProjectDates(item.StartDate, item.EndDate, contextLabel); err != nil {
		return err
	}
	if err := validateProjectTheme(item.Theme, themeIDs, contextLabel); err != nil {
		return err
	}
	if err := validateProjectNotes(item.NotesEN, item.NotesKR, contextLabel); err != nil {
		return err
	}
	if item.Media == nil {
		return fmt.Errorf("%s의 media는 JSON 배열이어야 합니다", contextLabel)
	}
	return nil
}

func validateProjectNotes(notesEN, notesKR []string, contextLabel string) error {
	for _, language := range []struct {
		field string
		notes []string
	}{
		{"notes_en", notesEN},
		{"notes_kr", notesKR},
	} {
		if len(language.notes) == 0 {
			return fmt.Errorf("%s의 %s에는 하나 이상의 노트가 필요합니다", contextLabel, language.field)
		}
		if len(language.notes) > maxProjectNotes {
			return fmt.Errorf("%s의 %s는 최대 %d개까지 저장할 수 있습니다", contextLabel, language.field, maxProjectNotes)
		}
		for index, note := range language.notes {
			if err := validateBoardText(strings.TrimSpace(note), fmt.Sprintf("%s %s %d", contextLabel, language.field, index+1), maxProjectNoteLength, true); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateProjectDates(startDate, endDate, contextLabel string) error {
	start, err := parseCanonicalDate(startDate)
	if err != nil {
		return fmt.Errorf("%s의 시작일은 YYYY-MM-DD 형식의 실제 날짜여야 합니다", contextLabel)
	}
	end, err := parseCanonicalDate(endDate)
	if err != nil {
		return fmt.Errorf("%s의 종료일은 YYYY-MM-DD 형식의 실제 날짜여야 합니다", contextLabel)
	}
	if start.After(end) {
		return fmt.Errorf("%s의 시작일은 종료일보다 늦을 수 없습니다", contextLabel)
	}
	return nil
}

func validateProjectTheme(theme string, themeIDs map[string]bool, contextLabel string) error {
	if !taxonomyIDPattern.MatchString(theme) {
		return fmt.Errorf("%s의 theme은 영문 소문자, 숫자, 하이픈으로 된 canonical ID여야 합니다", contextLabel)
	}
	if themeIDs != nil && !themeIDs[theme] {
		return fmt.Errorf("%s가 존재하지 않는 프로젝트 테마 '%s'을(를) 참조합니다", contextLabel, theme)
	}
	return nil
}

func validateProjectCaptionPair(captionEN, captionKO, contextLabel string, trim bool) error {
	if trim {
		captionEN = strings.TrimSpace(captionEN)
		captionKO = strings.TrimSpace(captionKO)
	}
	if (captionEN == "") != (captionKO == "") {
		return fmt.Errorf("%s의 caption_en과 caption_ko는 모두 채우거나 모두 비워야 합니다", contextLabel)
	}
	return nil
}

func projectItemsResponse(snapshot projectSnapshot) []ProjectItemSummary {
	items := make([]ProjectItemSummary, 0, len(snapshot.Rows))
	for _, row := range snapshot.Rows {
		media := make([]ProjectMediaSummary, 0, len(row.Item.Media))
		for _, current := range row.Item.Media {
			media = append(media, ProjectMediaSummary{
				EditorKey:    current.editorKey,
				Src:          current.Src,
				Type:         current.Type,
				Poster:       current.Poster,
				CaptionEN:    current.CaptionEN,
				CaptionKO:    current.CaptionKO,
				OriginalName: current.originalName,
				Size:         current.size,
				PreviewURL:   current.previewURL,
			})
		}
		items = append(items, ProjectItemSummary{
			EditorKey:   row.Key,
			VisibleInCV: mustVisibleInCV(row.Raw),
			TitleEN:     row.Item.TitleEN,
			TitleKO:     row.Item.TitleKO,
			StartDate:   row.Item.StartDate,
			EndDate:     row.Item.EndDate,
			Theme:       row.Item.Theme,
			FunderEN:    row.Item.FunderEN,
			FunderKO:    row.Item.FunderKO,
			NotesEN:     append([]string(nil), row.Item.NotesEN...),
			NotesKR:     append([]string(nil), row.Item.NotesKR...),
			Media:       media,
			MediaCount:  len(media),
		})
	}
	return items
}

// buildProjectSaveLocked prepares projects.json and repository media copies.
// It does not publish or write either: SaveEditorData owns the shared revision
// checks and transaction. unavailableStageTokens should contain tokens already
// claimed by another collection in the same save.
func (a *App) buildProjectSaveLocked(
	root string,
	current projectSnapshot,
	request []ProjectSaveItem,
	themeIDs map[string]bool,
	unavailableStageTokens map[string]bool,
) ([]byte, []pendingBoardMedia, []string, error) {
	if len(request) > maxProjects {
		return nil, nil, nil, fmt.Errorf("프로젝트는 최대 %d개까지 저장할 수 있습니다", maxProjects)
	}
	currentByKey := make(map[string]projectRow, len(current.Rows))
	for _, row := range current.Rows {
		currentByKey[row.Key] = row
	}
	reservedNames, err := boardMediaNames(filepath.Join(root, "data", "media"))
	if err != nil {
		return nil, nil, nil, err
	}
	usedExisting := make(map[string]bool)
	usedTokens := make(map[string]bool, len(unavailableStageTokens))
	for token, unavailable := range unavailableStageTokens {
		if unavailable {
			usedTokens[token] = true
		}
	}
	type preparedProjectRow struct {
		raw  json.RawMessage
		item projectDocumentItem
	}
	prepared := make([]preparedProjectRow, 0, len(request))
	pending := make([]pendingBoardMedia, 0)
	tokens := make([]string, 0)

	for index, requested := range request {
		contextLabel := fmt.Sprintf("프로젝트 %d", index+1)
		var currentRow *projectRow
		if requested.EditorKey != "" {
			row, exists := currentByKey[requested.EditorKey]
			if !exists {
				return nil, nil, nil, fmt.Errorf("%s가 현재 projects.json에 없습니다. 다시 불러와 주세요", contextLabel)
			}
			if usedExisting[requested.EditorKey] {
				return nil, nil, nil, fmt.Errorf("%s가 중복되었습니다", contextLabel)
			}
			usedExisting[requested.EditorKey] = true
			currentRow = &row
			if requested.Project == nil {
				rowRaw, err := withVisibleInCV(row.Raw, requested.VisibleInCV)
				if err != nil {
					return nil, nil, nil, fmt.Errorf("%s visible_in_CV 인코딩 실패: %w", contextLabel, err)
				}
				prepared = append(prepared, preparedProjectRow{raw: rowRaw, item: row.Item})
				continue
			}
		} else if requested.Project == nil {
			return nil, nil, nil, fmt.Errorf("새 %s의 project 내용이 없습니다", contextLabel)
		}

		document, raw, rowPending, rowTokens, err := a.prepareProjectLocked(
			root,
			*requested.Project,
			currentRow,
			contextLabel,
			themeIDs,
			reservedNames,
			usedTokens,
		)
		if err != nil {
			return nil, nil, nil, err
		}
		raw, err = withVisibleInCV(raw, visibleInCVForSave(requested.VisibleInCV, currentRow == nil))
		if err != nil {
			return nil, nil, nil, fmt.Errorf("%s visible_in_CV 인코딩 실패: %w", contextLabel, err)
		}
		prepared = append(prepared, preparedProjectRow{raw: raw, item: document})
		pending = append(pending, rowPending...)
		tokens = append(tokens, rowTokens...)
	}

	// Array position is the website and CV display order. Preserve the request
	// order exactly; do not apply board-style date sorting here.
	rows := make([]json.RawMessage, 0, len(prepared))
	for _, row := range prepared {
		rows = append(rows, row.raw)
	}
	encoded, err := dataversion.MarshalArray(current.Raw, rows, "projects.json")
	if err != nil {
		return nil, nil, nil, fmt.Errorf("projects.json 인코딩 실패: %w", err)
	}
	return encoded, pending, tokens, nil
}

func (a *App) prepareProjectLocked(
	root string,
	input ProjectInput,
	current *projectRow,
	contextLabel string,
	themeIDs map[string]bool,
	reservedNames map[string]bool,
	usedTokens map[string]bool,
) (projectDocumentItem, json.RawMessage, []pendingBoardMedia, []string, error) {
	if input.NotesEN == nil || input.NotesKR == nil || input.Media == nil {
		return projectDocumentItem{}, nil, nil, nil, fmt.Errorf("%s의 notes_en, notes_kr, media는 JSON 배열이어야 합니다", contextLabel)
	}
	if current != nil && projectInputMatchesCurrent(input, current.Item) {
		return current.Item, cloneRawMessage(current.Raw), nil, nil, nil
	}
	project, err := normaliseProjectInput(input, themeIDs, contextLabel)
	if err != nil {
		return projectDocumentItem{}, nil, nil, nil, err
	}
	if current != nil {
		preserveUnchangedProjectValues(&project, input, current.Item)
	}

	currentMedia := make(map[string]projectDocumentMedia)
	if current != nil {
		for _, media := range current.Item.Media {
			currentMedia[media.editorKey] = media
		}
	}
	usedExistingMedia := make(map[string]bool)
	documentMedia := make([]projectDocumentMedia, 0, len(project.Media))
	mediaRaw := make([]json.RawMessage, 0, len(project.Media))
	pending := make([]pendingBoardMedia, 0)
	tokens := make([]string, 0)
	mediaChanged := current == nil || len(project.Media) != len(current.Item.Media)

	for mediaIndex, mediaInput := range project.Media {
		if mediaInput.EditorKey != "" {
			existing, exists := currentMedia[mediaInput.EditorKey]
			if !exists {
				return projectDocumentItem{}, nil, nil, nil, fmt.Errorf("%s 미디어 %d가 현재 projects.json에 없습니다. 다시 불러와 주세요", contextLabel, mediaIndex+1)
			}
			if usedExistingMedia[mediaInput.EditorKey] {
				return projectDocumentItem{}, nil, nil, nil, fmt.Errorf("%s 미디어 %d가 중복되었습니다", contextLabel, mediaIndex+1)
			}
			usedExistingMedia[mediaInput.EditorKey] = true
			captionEN := mediaInput.CaptionEN
			captionKO := mediaInput.CaptionKO
			if input.Media[mediaIndex].CaptionEN == existing.CaptionEN {
				captionEN = existing.CaptionEN
			}
			if input.Media[mediaIndex].CaptionKO == existing.CaptionKO {
				captionKO = existing.CaptionKO
			}
			updated := existing
			updated.CaptionEN = captionEN
			updated.CaptionKO = captionKO
			raw, err := marshalEditedProjectMedia(existing, captionEN, captionKO)
			if err != nil {
				return projectDocumentItem{}, nil, nil, nil, fmt.Errorf("%s 미디어 %d 인코딩 실패: %w", contextLabel, mediaIndex+1, err)
			}
			updated.raw = raw
			documentMedia = append(documentMedia, updated)
			mediaRaw = append(mediaRaw, raw)
			if current == nil || mediaIndex >= len(current.Item.Media) ||
				mediaInput.EditorKey != current.Item.Media[mediaIndex].editorKey ||
				captionEN != current.Item.Media[mediaIndex].CaptionEN ||
				captionKO != current.Item.Media[mediaIndex].CaptionKO {
				mediaChanged = true
			}
			continue
		}

		staged, exists := a.boardMedia[mediaInput.StageToken]
		if !exists {
			return projectDocumentItem{}, nil, nil, nil, fmt.Errorf("%s 미디어 %d 임시 파일이 없습니다. 다시 드롭해 주세요", contextLabel, mediaIndex+1)
		}
		if usedTokens[mediaInput.StageToken] {
			return projectDocumentItem{}, nil, nil, nil, fmt.Errorf("미디어 '%s'가 중복 연결되었습니다", staged.OriginalName)
		}
		usedTokens[mediaInput.StageToken] = true
		filename := nextProjectMediaName(project.StartDate, project.TitleEN, mediaIndex, staged.Extension, reservedNames)
		reservedNames[strings.ToLower(filename)] = true
		pending = append(pending, pendingBoardMedia{
			Token:           mediaInput.StageToken,
			StagedPath:      staged.Path,
			DestinationPath: filepath.Join(root, "data", "media", filename),
		})
		tokens = append(tokens, mediaInput.StageToken)
		media := projectDocumentMedia{
			Src:          filename,
			CaptionEN:    mediaInput.CaptionEN,
			CaptionKO:    mediaInput.CaptionKO,
			originalName: staged.OriginalName,
			size:         staged.Size,
			previewURL:   boardThumbnailDataURL(staged.Path),
		}
		raw, err := json.Marshal(media)
		if err != nil {
			return projectDocumentItem{}, nil, nil, nil, fmt.Errorf("%s 미디어 %d 인코딩 실패: %w", contextLabel, mediaIndex+1, err)
		}
		media.raw = raw
		documentMedia = append(documentMedia, media)
		mediaRaw = append(mediaRaw, raw)
		mediaChanged = true
	}

	document := projectDocumentItem{
		TitleEN:   project.TitleEN,
		TitleKO:   project.TitleKO,
		StartDate: project.StartDate,
		EndDate:   project.EndDate,
		Theme:     project.Theme,
		FunderEN:  project.FunderEN,
		FunderKO:  project.FunderKO,
		NotesEN:   append([]string(nil), project.NotesEN...),
		NotesKR:   append([]string(nil), project.NotesKR...),
		Media:     documentMedia,
	}
	known := projectChangedFields(document, current, mediaRaw, mediaChanged)
	if current != nil && len(known) == 0 {
		return document, cloneRawMessage(current.Raw), pending, tokens, nil
	}
	fields := make(map[string]json.RawMessage)
	if current != nil {
		fields, err = decodeObject(current.Raw)
		if err != nil {
			return projectDocumentItem{}, nil, nil, nil, err
		}
	}
	merged, err := mergeObjectFields(fields, known)
	if err != nil {
		return projectDocumentItem{}, nil, nil, nil, fmt.Errorf("%s 인코딩 실패: %w", contextLabel, err)
	}
	raw, err := json.Marshal(merged)
	if err != nil {
		return projectDocumentItem{}, nil, nil, nil, fmt.Errorf("%s 인코딩 실패: %w", contextLabel, err)
	}
	return document, raw, pending, tokens, nil
}

func projectInputMatchesCurrent(input ProjectInput, current projectDocumentItem) bool {
	if input.TitleEN != current.TitleEN || input.TitleKO != current.TitleKO ||
		input.StartDate != current.StartDate || input.EndDate != current.EndDate ||
		input.Theme != current.Theme || input.FunderEN != current.FunderEN ||
		input.FunderKO != current.FunderKO ||
		!equalStrings(input.NotesEN, current.NotesEN) ||
		!equalStrings(input.NotesKR, current.NotesKR) ||
		len(input.Media) != len(current.Media) {
		return false
	}
	for index, requested := range input.Media {
		existing := current.Media[index]
		if requested.EditorKey != existing.editorKey || requested.StageToken != "" ||
			requested.CaptionEN != existing.CaptionEN || requested.CaptionKO != existing.CaptionKO {
			return false
		}
	}
	return true
}

func normaliseProjectInput(input ProjectInput, themeIDs map[string]bool, contextLabel string) (ProjectInput, error) {
	project := ProjectInput{
		TitleEN:   strings.TrimSpace(input.TitleEN),
		TitleKO:   strings.TrimSpace(input.TitleKO),
		StartDate: strings.TrimSpace(input.StartDate),
		EndDate:   strings.TrimSpace(input.EndDate),
		Theme:     strings.TrimSpace(input.Theme),
		FunderEN:  strings.TrimSpace(input.FunderEN),
		FunderKO:  strings.TrimSpace(input.FunderKO),
		NotesEN:   make([]string, 0, len(input.NotesEN)),
		NotesKR:   make([]string, 0, len(input.NotesKR)),
		Media:     make([]ProjectMediaInput, 0, len(input.Media)),
	}
	if err := validateProjectDates(project.StartDate, project.EndDate, contextLabel); err != nil {
		return ProjectInput{}, err
	}
	if err := validateProjectTheme(project.Theme, themeIDs, contextLabel); err != nil {
		return ProjectInput{}, err
	}
	for label, value := range map[string]string{
		"영문 제목":  project.TitleEN,
		"국문 제목":  project.TitleKO,
		"영문 발주처": project.FunderEN,
		"국문 발주처": project.FunderKO,
	} {
		if err := validateBoardText(value, contextLabel+" "+label, 500, false); err != nil {
			return ProjectInput{}, err
		}
	}
	for _, note := range input.NotesEN {
		project.NotesEN = append(project.NotesEN, strings.TrimSpace(note))
	}
	for _, note := range input.NotesKR {
		project.NotesKR = append(project.NotesKR, strings.TrimSpace(note))
	}
	if err := validateProjectNotes(project.NotesEN, project.NotesKR, contextLabel); err != nil {
		return ProjectInput{}, err
	}
	if len(input.Media) > maxBoardImagesTotal {
		return ProjectInput{}, fmt.Errorf("%s에는 미디어를 최대 %d개까지 넣을 수 있습니다", contextLabel, maxBoardImagesTotal)
	}
	for index, media := range input.Media {
		captionEN := strings.TrimSpace(media.CaptionEN)
		captionKO := strings.TrimSpace(media.CaptionKO)
		if (media.EditorKey == "") == (media.StageToken == "") {
			return ProjectInput{}, fmt.Errorf("%s 미디어 %d의 참조 형식이 올바르지 않습니다", contextLabel, index+1)
		}
		if err := validateProjectCaptionPair(captionEN, captionKO, fmt.Sprintf("%s 미디어 %d", contextLabel, index+1), false); err != nil {
			return ProjectInput{}, err
		}
		if err := validateBoardText(captionEN, fmt.Sprintf("%s 미디어 %d 영문 설명", contextLabel, index+1), 500, false); err != nil {
			return ProjectInput{}, err
		}
		if err := validateBoardText(captionKO, fmt.Sprintf("%s 미디어 %d 국문 설명", contextLabel, index+1), 500, false); err != nil {
			return ProjectInput{}, err
		}
		project.Media = append(project.Media, ProjectMediaInput{
			EditorKey:  media.EditorKey,
			StageToken: media.StageToken,
			CaptionEN:  captionEN,
			CaptionKO:  captionKO,
		})
	}
	return project, nil
}

func preserveUnchangedProjectValues(project *ProjectInput, input ProjectInput, current projectDocumentItem) {
	if input.TitleEN == current.TitleEN {
		project.TitleEN = current.TitleEN
	}
	if input.TitleKO == current.TitleKO {
		project.TitleKO = current.TitleKO
	}
	if input.StartDate == current.StartDate {
		project.StartDate = current.StartDate
	}
	if input.EndDate == current.EndDate {
		project.EndDate = current.EndDate
	}
	if input.Theme == current.Theme {
		project.Theme = current.Theme
	}
	if input.FunderEN == current.FunderEN {
		project.FunderEN = current.FunderEN
	}
	if input.FunderKO == current.FunderKO {
		project.FunderKO = current.FunderKO
	}
	if equalStrings(input.NotesEN, current.NotesEN) {
		project.NotesEN = append([]string(nil), current.NotesEN...)
	}
	if equalStrings(input.NotesKR, current.NotesKR) {
		project.NotesKR = append([]string(nil), current.NotesKR...)
	}
}

func projectChangedFields(document projectDocumentItem, current *projectRow, media []json.RawMessage, mediaChanged bool) map[string]any {
	known := make(map[string]any)
	if current == nil || document.TitleEN != current.Item.TitleEN {
		known["title_en"] = document.TitleEN
	}
	if current == nil || document.TitleKO != current.Item.TitleKO {
		known["title_ko"] = document.TitleKO
	}
	if current == nil || document.StartDate != current.Item.StartDate {
		known["start_date"] = document.StartDate
	}
	if current == nil || document.EndDate != current.Item.EndDate {
		known["end_date"] = document.EndDate
	}
	if current == nil || document.Theme != current.Item.Theme {
		known["theme"] = document.Theme
	}
	if current == nil || document.FunderEN != current.Item.FunderEN {
		known["funder_en"] = document.FunderEN
	}
	if current == nil || document.FunderKO != current.Item.FunderKO {
		known["funder_ko"] = document.FunderKO
	}
	if current == nil || !equalStrings(document.NotesEN, current.Item.NotesEN) {
		known["notes_en"] = document.NotesEN
	}
	if current == nil || !equalStrings(document.NotesKR, current.Item.NotesKR) {
		known["notes_kr"] = document.NotesKR
	}
	if current == nil || mediaChanged {
		known["media"] = media
	}
	return known
}

func marshalEditedProjectMedia(current projectDocumentMedia, captionEN, captionKO string) (json.RawMessage, error) {
	known := make(map[string]any, 2)
	if captionEN != current.CaptionEN {
		known["caption_en"] = captionEN
	}
	if captionKO != current.CaptionKO {
		known["caption_ko"] = captionKO
	}
	if len(known) == 0 {
		return cloneRawMessage(current.raw), nil
	}
	fields, err := decodeObject(current.raw)
	if err != nil {
		return nil, err
	}
	merged, err := mergeObjectFields(fields, known)
	if err != nil {
		return nil, err
	}
	return json.Marshal(merged)
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func nextProjectMediaName(startDate, titleEN string, mediaIndex int, extension string, used map[string]bool) string {
	datePart := strings.ReplaceAll(startDate, "-", "")
	titlePart := boardTitleSlug(titleEN)
	base := fmt.Sprintf("project-%s-%s-%02d", datePart, titlePart, mediaIndex+1)
	for suffix := 1; ; suffix++ {
		candidate := base + extension
		if suffix > 1 {
			candidate = fmt.Sprintf("%s-%d%s", base, suffix, extension)
		}
		if !used[strings.ToLower(candidate)] {
			return candidate
		}
	}
}
