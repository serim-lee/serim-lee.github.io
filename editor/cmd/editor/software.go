package main

import (
	"encoding/json"
	"fmt"
	"github.com/Gonie-Gonie/Gonie-Gonie.github.io/editor/internal/dataversion"
	"net/url"
	"os"
	pathpkg "path"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	maxSoftwareItems        = 500
	maxSoftwareLinks        = 100
	maxSoftwareTechnologies = 100
	maxSoftwareNotes        = 100
)

var softwareStages = map[string]bool{
	"release":     true,
	"preview":     true,
	"development": true,
}

type SoftwareLinkSummary struct {
	EditorKey string `json:"editor_key"`
	URL       string `json:"url"`
	Label     string `json:"label"`
	LabelEN   string `json:"label_en"`
	LabelKO   string `json:"label_ko"`
}

type SoftwareMediaSummary struct {
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

type SoftwareItemSummary struct {
	EditorKey    string                 `json:"editor_key"`
	VisibleInCV  bool                   `json:"visible_in_CV"`
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Stage        string                 `json:"stage"`
	Links        []SoftwareLinkSummary  `json:"links"`
	NotesEN      []string               `json:"notes_en"`
	NotesKR      []string               `json:"notes_kr"`
	Media        []SoftwareMediaSummary `json:"media"`
	Technologies []string               `json:"technologies"`
}

type SoftwareLinkInput struct {
	EditorKey string `json:"editor_key,omitempty"`
	URL       string `json:"url"`
	Label     string `json:"label"`
	LabelEN   string `json:"label_en"`
	LabelKO   string `json:"label_ko"`
}

type SoftwareMediaInput struct {
	EditorKey  string `json:"editor_key,omitempty"`
	StageToken string `json:"stage_token,omitempty"`
	CaptionEN  string `json:"caption_en"`
	CaptionKO  string `json:"caption_ko"`
}

type SoftwareInput struct {
	Name         string               `json:"name"`
	Stage        string               `json:"stage"`
	Links        []SoftwareLinkInput  `json:"links"`
	NotesEN      []string             `json:"notes_en"`
	NotesKR      []string             `json:"notes_kr"`
	Media        []SoftwareMediaInput `json:"media"`
	Technologies []string             `json:"technologies"`
}

type SoftwareSaveItem struct {
	EditorKey   string         `json:"editor_key,omitempty"`
	Software    *SoftwareInput `json:"software,omitempty"`
	VisibleInCV *bool          `json:"visible_in_CV,omitempty"`
}

type softwareDocumentLink struct {
	URL     string
	Label   string
	LabelEN string
	LabelKO string

	editorKey string
	raw       json.RawMessage
	isObject  bool
}

type softwareDocumentMedia struct {
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

type softwareDocumentItem struct {
	ID           string
	Name         string
	Stage        string
	Links        []softwareDocumentLink
	NotesEN      []string
	NotesKR      []string
	Media        []softwareDocumentMedia
	Technologies []string
}

type softwareRow struct {
	Key  string
	Raw  json.RawMessage
	Item softwareDocumentItem
}

type softwareSnapshot struct {
	Raw  []byte
	Rows []softwareRow
}

func readSoftware(path, root string) (softwareSnapshot, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return softwareSnapshot{}, fmt.Errorf("software.json 읽기 실패: %w", err)
	}
	return readSoftwareBytes(raw, root)
}

func readSoftwareBytes(raw []byte, root string) (softwareSnapshot, error) {
	rows, err := dataversion.Array(raw, "software.json")
	if err != nil {
		return softwareSnapshot{}, err
	}
	if len(rows) > maxSoftwareItems {
		return softwareSnapshot{}, fmt.Errorf("소프트웨어는 최대 %d개까지 관리할 수 있습니다", maxSoftwareItems)
	}

	snapshot := softwareSnapshot{
		Raw:  append([]byte(nil), raw...),
		Rows: make([]softwareRow, 0, len(rows)),
	}
	usedIDs := make(map[string]bool, len(rows))
	for index, rowRaw := range rows {
		contextLabel := fmt.Sprintf("software.json %d번째 항목", index+1)
		item, err := parseSoftwareItem(rowRaw, root, contextLabel)
		if err != nil {
			return softwareSnapshot{}, err
		}
		if _, err := visibleInCV(rowRaw, contextLabel); err != nil {
			return softwareSnapshot{}, err
		}
		if usedIDs[item.ID] {
			return softwareSnapshot{}, fmt.Errorf("software.json에 중복 ID '%s'이(가) 있습니다", item.ID)
		}
		usedIDs[item.ID] = true
		hash := revisionOf(rowRaw)
		snapshot.Rows = append(snapshot.Rows, softwareRow{
			Key:  fmt.Sprintf("%d:%s", index, hash[:20]),
			Raw:  append(json.RawMessage(nil), rowRaw...),
			Item: item,
		})
	}
	return snapshot, nil
}

func parseSoftwareItem(raw json.RawMessage, root, contextLabel string) (softwareDocumentItem, error) {
	fields, err := decodeObject(raw)
	if err != nil {
		return softwareDocumentItem{}, fmt.Errorf("%s은 JSON 객체여야 합니다", contextLabel)
	}
	for _, legacy := range []string{"summary", "summary_en", "summary_ko"} {
		if _, exists := fields[legacy]; exists {
			return softwareDocumentItem{}, fmt.Errorf("%s은 summary 대신 notes_en과 notes_kr를 사용해야 합니다", contextLabel)
		}
	}
	for _, legacy := range []string{"repo_url", "website_url"} {
		if _, exists := fields[legacy]; exists {
			return softwareDocumentItem{}, fmt.Errorf("%s은 %s 대신 links 배열을 사용해야 합니다", contextLabel, legacy)
		}
	}
	if _, exists := fields["photos"]; exists {
		return softwareDocumentItem{}, fmt.Errorf("%s은 photos 대신 media를 사용해야 합니다", contextLabel)
	}
	if _, exists := fields["order"]; exists {
		return softwareDocumentItem{}, fmt.Errorf("%s은 order를 사용하지 않고 배열 순서로 표시합니다", contextLabel)
	}
	for _, field := range []string{"id", "name", "stage", "links", "notes_en", "notes_kr", "media", "technologies"} {
		if _, exists := fields[field]; !exists {
			return softwareDocumentItem{}, fmt.Errorf("%s에 %s 필드가 없습니다", contextLabel, field)
		}
	}

	var wire struct {
		ID           string            `json:"id"`
		Name         string            `json:"name"`
		Stage        string            `json:"stage"`
		Links        []json.RawMessage `json:"links"`
		NotesEN      []string          `json:"notes_en"`
		NotesKR      []string          `json:"notes_kr"`
		Media        []json.RawMessage `json:"media"`
		Technologies []string          `json:"technologies"`
	}
	if err := json.Unmarshal(raw, &wire); err != nil {
		return softwareDocumentItem{}, fmt.Errorf("%s 필드 형식이 올바르지 않습니다: %w", contextLabel, err)
	}
	if wire.Links == nil || wire.Media == nil || wire.NotesEN == nil || wire.NotesKR == nil || wire.Technologies == nil {
		return softwareDocumentItem{}, fmt.Errorf("%s의 links, notes_en, notes_kr, media, technologies는 JSON 배열이어야 합니다", contextLabel)
	}
	if !taxonomyIDPattern.MatchString(wire.ID) {
		return softwareDocumentItem{}, fmt.Errorf("%s의 ID '%s' 형식이 올바르지 않습니다", contextLabel, wire.ID)
	}
	if err := validateSoftwareText(wire.Name, contextLabel+" 이름", 300, true, false); err != nil {
		return softwareDocumentItem{}, err
	}
	if !softwareStages[wire.Stage] {
		return softwareDocumentItem{}, fmt.Errorf("%s의 stage는 release, preview, development 중 하나여야 합니다", contextLabel)
	}
	if err := validateSoftwareNotes(wire.NotesEN, wire.NotesKR, contextLabel, false); err != nil {
		return softwareDocumentItem{}, err
	}
	if err := validateSoftwareTechnologies(wire.Technologies, contextLabel, false); err != nil {
		return softwareDocumentItem{}, err
	}

	item := softwareDocumentItem{
		ID:           wire.ID,
		Name:         wire.Name,
		Stage:        wire.Stage,
		Links:        make([]softwareDocumentLink, 0, len(wire.Links)),
		NotesEN:      append([]string(nil), wire.NotesEN...),
		NotesKR:      append([]string(nil), wire.NotesKR...),
		Media:        make([]softwareDocumentMedia, 0, len(wire.Media)),
		Technologies: append([]string(nil), wire.Technologies...),
	}
	if len(wire.Links) > maxSoftwareLinks {
		return softwareDocumentItem{}, fmt.Errorf("%s의 링크는 최대 %d개까지 입력할 수 있습니다", contextLabel, maxSoftwareLinks)
	}
	for index, linkRaw := range wire.Links {
		link, err := parseSoftwareLink(linkRaw, fmt.Sprintf("%s 링크 %d", contextLabel, index+1))
		if err != nil {
			return softwareDocumentItem{}, err
		}
		link.editorKey = fmt.Sprintf("%d:%s", index, revisionOf(linkRaw)[:20])
		link.raw = append(json.RawMessage(nil), linkRaw...)
		item.Links = append(item.Links, link)
	}
	if len(wire.Media) > maxBoardImagesTotal {
		return softwareDocumentItem{}, fmt.Errorf("%s의 미디어는 최대 %d개까지 입력할 수 있습니다", contextLabel, maxBoardImagesTotal)
	}
	for index, mediaRaw := range wire.Media {
		media, err := parseSoftwareMedia(mediaRaw, root, fmt.Sprintf("%s 미디어 %d", contextLabel, index+1))
		if err != nil {
			return softwareDocumentItem{}, err
		}
		media.editorKey = fmt.Sprintf("%d:%s", index, revisionOf(mediaRaw)[:20])
		media.raw = append(json.RawMessage(nil), mediaRaw...)
		item.Media = append(item.Media, media)
	}
	return item, nil
}

func parseSoftwareLink(raw json.RawMessage, contextLabel string) (softwareDocumentLink, error) {
	var stringValue string
	if err := json.Unmarshal(raw, &stringValue); err == nil {
		link := softwareDocumentLink{URL: strings.TrimSpace(stringValue)}
		if err := validateSoftwareURL(link.URL, contextLabel); err != nil {
			return softwareDocumentLink{}, err
		}
		return link, nil
	}

	fields, err := decodeObject(raw)
	if err != nil {
		return softwareDocumentLink{}, fmt.Errorf("%s는 URL 문자열 또는 URL 객체여야 합니다", contextLabel)
	}
	if _, exists := fields["url"]; !exists {
		return softwareDocumentLink{}, fmt.Errorf("%s에 url 필드가 없습니다", contextLabel)
	}
	var object struct {
		URL     string `json:"url"`
		Label   string `json:"label"`
		LabelEN string `json:"label_en"`
		LabelKO string `json:"label_ko"`
	}
	if err := json.Unmarshal(raw, &object); err != nil {
		return softwareDocumentLink{}, fmt.Errorf("%s 필드 형식이 올바르지 않습니다: %w", contextLabel, err)
	}
	link := softwareDocumentLink{
		URL:      strings.TrimSpace(object.URL),
		Label:    object.Label,
		LabelEN:  object.LabelEN,
		LabelKO:  object.LabelKO,
		isObject: true,
	}
	if err := validateSoftwareURL(link.URL, contextLabel); err != nil {
		return softwareDocumentLink{}, err
	}
	for field, value := range map[string]string{
		"label": link.Label, "label_en": link.LabelEN, "label_ko": link.LabelKO,
	} {
		if err := validateSoftwareText(value, contextLabel+" "+field, 300, false, false); err != nil {
			return softwareDocumentLink{}, err
		}
	}
	return link, nil
}

func parseSoftwareMedia(raw json.RawMessage, root, contextLabel string) (softwareDocumentMedia, error) {
	fields, err := decodeObject(raw)
	if err != nil {
		return softwareDocumentMedia{}, fmt.Errorf("%s는 JSON 객체여야 합니다", contextLabel)
	}
	if _, exists := fields["src"]; !exists {
		return softwareDocumentMedia{}, fmt.Errorf("%s에 src 필드가 없습니다", contextLabel)
	}
	var media softwareDocumentMedia
	if err := json.Unmarshal(raw, &media); err != nil {
		return softwareDocumentMedia{}, fmt.Errorf("%s 필드 형식이 올바르지 않습니다: %w", contextLabel, err)
	}
	if media.Type != "" && media.Type != "image" && media.Type != "video" {
		return softwareDocumentMedia{}, fmt.Errorf("%s의 type은 image 또는 video여야 합니다", contextLabel)
	}
	if err := validateSoftwareText(media.CaptionEN, contextLabel+" 영문 설명", 500, true, false); err != nil {
		return softwareDocumentMedia{}, err
	}
	if err := validateSoftwareText(media.CaptionKO, contextLabel+" 국문 설명", 500, true, false); err != nil {
		return softwareDocumentMedia{}, err
	}
	resolvedSource, sourceInfo, err := resolveBoardMediaPath(root, media.Src, contextLabel)
	if err != nil {
		return softwareDocumentMedia{}, err
	}
	previewURL := boardThumbnailDataURL(resolvedSource)
	if media.Poster != "" {
		resolvedPoster, _, err := resolveBoardMediaPath(root, media.Poster, contextLabel+" poster")
		if err != nil {
			return softwareDocumentMedia{}, err
		}
		if previewURL == "" {
			previewURL = boardThumbnailDataURL(resolvedPoster)
		}
	}
	media.originalName = pathpkg.Base(media.Src)
	media.size = sourceInfo.Size()
	media.previewURL = previewURL
	return media, nil
}

func softwareItemSummaries(snapshot softwareSnapshot) []SoftwareItemSummary {
	result := make([]SoftwareItemSummary, 0, len(snapshot.Rows))
	for _, row := range snapshot.Rows {
		links := make([]SoftwareLinkSummary, 0, len(row.Item.Links))
		for _, link := range row.Item.Links {
			links = append(links, SoftwareLinkSummary{
				EditorKey: link.editorKey,
				URL:       link.URL,
				Label:     link.Label,
				LabelEN:   link.LabelEN,
				LabelKO:   link.LabelKO,
			})
		}
		media := make([]SoftwareMediaSummary, 0, len(row.Item.Media))
		for _, item := range row.Item.Media {
			media = append(media, SoftwareMediaSummary{
				EditorKey:    item.editorKey,
				Src:          item.Src,
				Type:         item.Type,
				Poster:       item.Poster,
				CaptionEN:    item.CaptionEN,
				CaptionKO:    item.CaptionKO,
				OriginalName: item.originalName,
				Size:         item.size,
				PreviewURL:   item.previewURL,
			})
		}
		result = append(result, SoftwareItemSummary{
			EditorKey:    row.Key,
			VisibleInCV:  mustVisibleInCV(row.Raw),
			ID:           row.Item.ID,
			Name:         row.Item.Name,
			Stage:        row.Item.Stage,
			Links:        links,
			NotesEN:      append([]string(nil), row.Item.NotesEN...),
			NotesKR:      append([]string(nil), row.Item.NotesKR...),
			Media:        media,
			Technologies: append([]string(nil), row.Item.Technologies...),
		})
	}
	return result
}

func (a *App) buildSoftwareSaveLocked(
	root string,
	current softwareSnapshot,
	request []SoftwareSaveItem,
	unavailableStageTokens map[string]bool,
) ([]byte, []pendingBoardMedia, []string, error) {
	if len(request) > maxSoftwareItems {
		return nil, nil, nil, fmt.Errorf("소프트웨어는 최대 %d개까지 저장할 수 있습니다", maxSoftwareItems)
	}
	currentByKey := make(map[string]softwareRow, len(current.Rows))
	usedIDs := make(map[string]bool, len(current.Rows)+len(request))
	for _, row := range current.Rows {
		currentByKey[row.Key] = row
		usedIDs[row.Item.ID] = true
	}
	reservedNames, err := boardMediaNames(filepath.Join(root, "data", "media"))
	if err != nil {
		return nil, nil, nil, err
	}

	type preparedRow struct {
		raw  json.RawMessage
		item softwareDocumentItem
	}
	prepared := make([]preparedRow, 0, len(request))
	usedExisting := make(map[string]bool)
	usedTokens := make(map[string]bool, len(unavailableStageTokens))
	for token, unavailable := range unavailableStageTokens {
		if unavailable {
			usedTokens[token] = true
		}
	}
	pending := make([]pendingBoardMedia, 0)
	tokens := make([]string, 0)

	for index, requested := range request {
		contextLabel := fmt.Sprintf("소프트웨어 %d", index+1)
		var currentRow *softwareRow
		id := ""
		if requested.EditorKey != "" {
			row, exists := currentByKey[requested.EditorKey]
			if !exists {
				return nil, nil, nil, fmt.Errorf("%s가 현재 software.json에 없습니다. 다시 불러와 주세요", contextLabel)
			}
			if usedExisting[requested.EditorKey] {
				return nil, nil, nil, fmt.Errorf("%s가 중복되었습니다", contextLabel)
			}
			usedExisting[requested.EditorKey] = true
			currentRow = &row
			id = row.Item.ID
			if requested.Software == nil {
				rowRaw, err := withVisibleInCV(row.Raw, requested.VisibleInCV)
				if err != nil {
					return nil, nil, nil, fmt.Errorf("%s visible_in_CV 인코딩 실패: %w", contextLabel, err)
				}
				prepared = append(prepared, preparedRow{raw: rowRaw, item: row.Item})
				continue
			}
		} else if requested.Software == nil {
			return nil, nil, nil, fmt.Errorf("%s의 software 입력이 없습니다", contextLabel)
		}

		// A full, unchanged editor payload is equivalent to the key-only form.
		// Check it before normalization so valid hand-edited whitespace and raw
		// link/media representations are not rewritten incidentally.
		if currentRow != nil && softwareInputExactlyMatches(*requested.Software, currentRow.Item) {
			rowRaw, err := withVisibleInCV(currentRow.Raw, requested.VisibleInCV)
			if err != nil {
				return nil, nil, nil, fmt.Errorf("%s visible_in_CV 인코딩 실패: %w", contextLabel, err)
			}
			prepared = append(prepared, preparedRow{raw: rowRaw, item: currentRow.Item})
			continue
		}
		input, err := normaliseSoftwareInput(*requested.Software, contextLabel)
		if err != nil {
			return nil, nil, nil, err
		}
		if currentRow != nil && softwareInputExactlyMatches(input, currentRow.Item) {
			rowRaw, err := withVisibleInCV(currentRow.Raw, requested.VisibleInCV)
			if err != nil {
				return nil, nil, nil, fmt.Errorf("%s visible_in_CV 인코딩 실패: %w", contextLabel, err)
			}
			prepared = append(prepared, preparedRow{raw: rowRaw, item: currentRow.Item})
			continue
		}
		if currentRow == nil {
			id, err = newTaxonomyID(input.Name, usedIDs)
			if err != nil {
				return nil, nil, nil, err
			}
			usedIDs[id] = true
		}

		links, linkRaw, err := buildSoftwareLinks(input.Links, currentRow, contextLabel)
		if err != nil {
			return nil, nil, nil, err
		}
		media, mediaRaw, mediaPending, mediaTokens, err := a.buildSoftwareMediaLocked(
			root, id, input.Media, currentRow, contextLabel, reservedNames, usedTokens,
		)
		if err != nil {
			return nil, nil, nil, err
		}
		pending = append(pending, mediaPending...)
		tokens = append(tokens, mediaTokens...)
		document := softwareDocumentItem{
			ID:           id,
			Name:         input.Name,
			Stage:        input.Stage,
			Links:        links,
			NotesEN:      input.NotesEN,
			NotesKR:      input.NotesKR,
			Media:        media,
			Technologies: input.Technologies,
		}
		rowRaw, err := marshalSoftwareItem(currentRow, document, linkRaw, mediaRaw)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("%s 인코딩 실패: %w", contextLabel, err)
		}
		rowRaw, err = withVisibleInCV(rowRaw, visibleInCVForSave(requested.VisibleInCV, currentRow == nil))
		if err != nil {
			return nil, nil, nil, fmt.Errorf("%s visible_in_CV 인코딩 실패: %w", contextLabel, err)
		}
		prepared = append(prepared, preparedRow{raw: rowRaw, item: document})
	}

	rows := make([]json.RawMessage, 0, len(prepared))
	for _, row := range prepared {
		rows = append(rows, row.raw)
	}
	encoded, err := dataversion.MarshalArray(current.Raw, rows, "software.json")
	if err != nil {
		return nil, nil, nil, fmt.Errorf("software.json 인코딩 실패: %w", err)
	}
	return encoded, pending, tokens, nil
}

func normaliseSoftwareInput(input SoftwareInput, contextLabel string) (SoftwareInput, error) {
	result := SoftwareInput{
		Name:         strings.TrimSpace(input.Name),
		Stage:        strings.TrimSpace(input.Stage),
		Links:        append([]SoftwareLinkInput(nil), input.Links...),
		NotesEN:      normaliseSoftwareStringList(input.NotesEN),
		NotesKR:      normaliseSoftwareStringList(input.NotesKR),
		Media:        append([]SoftwareMediaInput(nil), input.Media...),
		Technologies: normaliseSoftwareStringList(input.Technologies),
	}
	if input.Links == nil || input.NotesEN == nil || input.NotesKR == nil || input.Media == nil || input.Technologies == nil {
		return SoftwareInput{}, fmt.Errorf("%s의 links, notes_en, notes_kr, media, technologies는 배열이어야 합니다", contextLabel)
	}
	if err := validateSoftwareText(result.Name, contextLabel+" 이름", 300, true, true); err != nil {
		return SoftwareInput{}, err
	}
	if !softwareStages[result.Stage] {
		return SoftwareInput{}, fmt.Errorf("%s의 stage는 release, preview, development 중 하나여야 합니다", contextLabel)
	}
	if err := validateSoftwareNotes(result.NotesEN, result.NotesKR, contextLabel, true); err != nil {
		return SoftwareInput{}, err
	}
	if err := validateSoftwareTechnologies(result.Technologies, contextLabel, true); err != nil {
		return SoftwareInput{}, err
	}
	if len(result.Links) > maxSoftwareLinks {
		return SoftwareInput{}, fmt.Errorf("%s의 링크는 최대 %d개까지 입력할 수 있습니다", contextLabel, maxSoftwareLinks)
	}
	if len(result.Media) > maxBoardImagesTotal {
		return SoftwareInput{}, fmt.Errorf("%s의 미디어는 최대 %d개까지 입력할 수 있습니다", contextLabel, maxBoardImagesTotal)
	}
	return result, nil
}

func softwareInputExactlyMatches(input SoftwareInput, current softwareDocumentItem) bool {
	if input.Links == nil || input.NotesEN == nil || input.NotesKR == nil ||
		input.Media == nil || input.Technologies == nil {
		return false
	}
	if input.Name != current.Name || input.Stage != current.Stage ||
		!equalSoftwareStrings(input.NotesEN, current.NotesEN) ||
		!equalSoftwareStrings(input.NotesKR, current.NotesKR) ||
		!equalSoftwareStrings(input.Technologies, current.Technologies) ||
		len(input.Links) != len(current.Links) || len(input.Media) != len(current.Media) {
		return false
	}
	for index, link := range input.Links {
		if link.EditorKey != current.Links[index].editorKey ||
			link.URL != current.Links[index].URL ||
			link.Label != current.Links[index].Label ||
			link.LabelEN != current.Links[index].LabelEN ||
			link.LabelKO != current.Links[index].LabelKO {
			return false
		}
	}
	for index, media := range input.Media {
		if media.EditorKey != current.Media[index].editorKey || media.StageToken != "" ||
			media.CaptionEN != current.Media[index].CaptionEN ||
			media.CaptionKO != current.Media[index].CaptionKO {
			return false
		}
	}
	return true
}

func softwareLinkMatches(left, right softwareDocumentLink) bool {
	return left.URL == right.URL && left.Label == right.Label &&
		left.LabelEN == right.LabelEN && left.LabelKO == right.LabelKO
}

func equalSoftwareStrings(left, right []string) bool {
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

func buildSoftwareLinks(input []SoftwareLinkInput, current *softwareRow, contextLabel string) ([]softwareDocumentLink, []json.RawMessage, error) {
	currentByKey := make(map[string]softwareDocumentLink)
	if current != nil {
		for _, link := range current.Item.Links {
			currentByKey[link.editorKey] = link
		}
	}
	usedExisting := make(map[string]bool)
	links := make([]softwareDocumentLink, 0, len(input))
	rawRows := make([]json.RawMessage, 0, len(input))
	for index, value := range input {
		linkContext := fmt.Sprintf("%s 링크 %d", contextLabel, index+1)
		value.URL = strings.TrimSpace(value.URL)
		value.Label = strings.TrimSpace(value.Label)
		value.LabelEN = strings.TrimSpace(value.LabelEN)
		value.LabelKO = strings.TrimSpace(value.LabelKO)
		if err := validateSoftwareURL(value.URL, linkContext); err != nil {
			return nil, nil, err
		}
		for field, text := range map[string]string{"label": value.Label, "label_en": value.LabelEN, "label_ko": value.LabelKO} {
			if err := validateSoftwareText(text, linkContext+" "+field, 300, false, true); err != nil {
				return nil, nil, err
			}
		}

		link := softwareDocumentLink{URL: value.URL, Label: value.Label, LabelEN: value.LabelEN, LabelKO: value.LabelKO}
		var existing *softwareDocumentLink
		if value.EditorKey != "" {
			found, exists := currentByKey[value.EditorKey]
			if !exists {
				return nil, nil, fmt.Errorf("%s가 현재 소프트웨어에 없습니다", linkContext)
			}
			if usedExisting[value.EditorKey] {
				return nil, nil, fmt.Errorf("%s가 중복되었습니다", linkContext)
			}
			usedExisting[value.EditorKey] = true
			existing = &found
			link.editorKey = found.editorKey
			link.isObject = found.isObject
		}
		var (
			rowRaw json.RawMessage
			err    error
		)
		if existing != nil && softwareLinkMatches(link, *existing) {
			rowRaw = append(json.RawMessage(nil), existing.raw...)
		} else {
			rowRaw, err = marshalSoftwareLink(existing, link)
		}
		if err != nil {
			return nil, nil, fmt.Errorf("%s 인코딩 실패: %w", linkContext, err)
		}
		link.raw = rowRaw
		links = append(links, link)
		rawRows = append(rawRows, rowRaw)
	}
	return links, rawRows, nil
}

func (a *App) buildSoftwareMediaLocked(
	root, softwareID string,
	input []SoftwareMediaInput,
	current *softwareRow,
	contextLabel string,
	reservedNames map[string]bool,
	usedTokens map[string]bool,
) ([]softwareDocumentMedia, []json.RawMessage, []pendingBoardMedia, []string, error) {
	currentByKey := make(map[string]softwareDocumentMedia)
	if current != nil {
		for _, media := range current.Item.Media {
			currentByKey[media.editorKey] = media
		}
	}
	usedExisting := make(map[string]bool)
	mediaItems := make([]softwareDocumentMedia, 0, len(input))
	rawRows := make([]json.RawMessage, 0, len(input))
	pending := make([]pendingBoardMedia, 0)
	tokens := make([]string, 0)
	for index, value := range input {
		mediaContext := fmt.Sprintf("%s 미디어 %d", contextLabel, index+1)
		value.CaptionEN = strings.TrimSpace(value.CaptionEN)
		value.CaptionKO = strings.TrimSpace(value.CaptionKO)
		if err := validateSoftwareText(value.CaptionEN, mediaContext+" 영문 설명", 500, true, true); err != nil {
			return nil, nil, nil, nil, err
		}
		if err := validateSoftwareText(value.CaptionKO, mediaContext+" 국문 설명", 500, true, true); err != nil {
			return nil, nil, nil, nil, err
		}
		if (value.EditorKey == "") == (value.StageToken == "") {
			return nil, nil, nil, nil, fmt.Errorf("%s는 기존 editor_key 또는 신규 stage_token 중 하나만 가져야 합니다", mediaContext)
		}

		if value.EditorKey != "" {
			existing, exists := currentByKey[value.EditorKey]
			if !exists {
				return nil, nil, nil, nil, fmt.Errorf("%s가 현재 소프트웨어에 없습니다", mediaContext)
			}
			if usedExisting[value.EditorKey] {
				return nil, nil, nil, nil, fmt.Errorf("%s가 중복되었습니다", mediaContext)
			}
			usedExisting[value.EditorKey] = true
			existing.CaptionEN = value.CaptionEN
			existing.CaptionKO = value.CaptionKO
			rowRaw := append(json.RawMessage(nil), existing.raw...)
			if existing.CaptionEN != currentByKey[value.EditorKey].CaptionEN || existing.CaptionKO != currentByKey[value.EditorKey].CaptionKO {
				var err error
				rowRaw, err = marshalSoftwareMedia(existing)
				if err != nil {
					return nil, nil, nil, nil, fmt.Errorf("%s 인코딩 실패: %w", mediaContext, err)
				}
			}
			existing.raw = rowRaw
			mediaItems = append(mediaItems, existing)
			rawRows = append(rawRows, rowRaw)
			continue
		}

		staged, exists := a.boardMedia[value.StageToken]
		if !exists {
			return nil, nil, nil, nil, fmt.Errorf("%s 임시 파일이 없습니다. 다시 드롭해 주세요", mediaContext)
		}
		if usedTokens[value.StageToken] {
			return nil, nil, nil, nil, fmt.Errorf("사진 '%s'이 중복으로 연결되었습니다", staged.OriginalName)
		}
		usedTokens[value.StageToken] = true
		filename := nextSoftwareMediaName(softwareID, index, staged.Extension, reservedNames)
		reservedNames[strings.ToLower(filename)] = true
		pending = append(pending, pendingBoardMedia{
			Token:           value.StageToken,
			StagedPath:      staged.Path,
			DestinationPath: filepath.Join(root, "data", "media", filename),
		})
		tokens = append(tokens, value.StageToken)
		item := softwareDocumentMedia{Src: filename, CaptionEN: value.CaptionEN, CaptionKO: value.CaptionKO}
		rowRaw, err := marshalSoftwareMedia(item)
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("%s 인코딩 실패: %w", mediaContext, err)
		}
		item.raw = rowRaw
		mediaItems = append(mediaItems, item)
		rawRows = append(rawRows, rowRaw)
	}
	return mediaItems, rawRows, pending, tokens, nil
}

func marshalSoftwareItem(current *softwareRow, item softwareDocumentItem, links, media []json.RawMessage) (json.RawMessage, error) {
	fields := map[string]json.RawMessage{}
	if current != nil {
		var err error
		fields, err = decodeObject(current.Raw)
		if err != nil {
			return nil, err
		}
	}
	known := map[string]any{
		"id":           item.ID,
		"name":         item.Name,
		"stage":        item.Stage,
		"links":        links,
		"notes_en":     item.NotesEN,
		"notes_kr":     item.NotesKR,
		"media":        media,
		"technologies": item.Technologies,
	}
	merged, err := mergeObjectFields(fields, known)
	if err != nil {
		return nil, err
	}
	return json.Marshal(merged)
}

func marshalSoftwareLink(current *softwareDocumentLink, link softwareDocumentLink) (json.RawMessage, error) {
	hasLabels := link.Label != "" || link.LabelEN != "" || link.LabelKO != ""
	if current == nil || !current.isObject {
		if !hasLabels {
			return json.Marshal(link.URL)
		}
		return json.Marshal(map[string]string{
			"url": link.URL, "label": link.Label, "label_en": link.LabelEN, "label_ko": link.LabelKO,
		})
	}
	fields, err := decodeObject(current.raw)
	if err != nil {
		return nil, err
	}
	merged, err := mergeObjectFields(fields, map[string]any{"url": link.URL})
	if err != nil {
		return nil, err
	}
	for name, value := range map[string]string{"label": link.Label, "label_en": link.LabelEN, "label_ko": link.LabelKO} {
		if value == "" {
			delete(merged, name)
			continue
		}
		encoded, err := json.Marshal(value)
		if err != nil {
			return nil, err
		}
		merged[name] = encoded
	}
	return json.Marshal(merged)
}

func marshalSoftwareMedia(media softwareDocumentMedia) (json.RawMessage, error) {
	fields := map[string]json.RawMessage{}
	if media.raw != nil {
		var err error
		fields, err = decodeObject(media.raw)
		if err != nil {
			return nil, err
		}
	}
	merged, err := mergeObjectFields(fields, map[string]any{
		"src": media.Src, "caption_en": media.CaptionEN, "caption_ko": media.CaptionKO,
	})
	if err != nil {
		return nil, err
	}
	if media.Type == "" {
		delete(merged, "type")
	} else {
		encoded, _ := json.Marshal(media.Type)
		merged["type"] = encoded
	}
	if media.Poster == "" {
		delete(merged, "poster")
	} else {
		encoded, _ := json.Marshal(media.Poster)
		merged["poster"] = encoded
	}
	return json.Marshal(merged)
}

func validateSoftwareNotes(notesEN, notesKR []string, contextLabel string, normalized bool) error {
	if len(notesEN) == 0 || len(notesKR) == 0 {
		return fmt.Errorf("%s의 notes_en과 notes_kr에는 각각 하나 이상의 설명이 필요합니다", contextLabel)
	}
	if len(notesEN) > maxSoftwareNotes || len(notesKR) > maxSoftwareNotes {
		return fmt.Errorf("%s의 언어별 설명은 최대 %d개까지 입력할 수 있습니다", contextLabel, maxSoftwareNotes)
	}
	for index, note := range notesEN {
		if err := validateSoftwareText(note, fmt.Sprintf("%s 영문 설명 %d", contextLabel, index+1), 5000, true, normalized); err != nil {
			return err
		}
	}
	for index, note := range notesKR {
		if err := validateSoftwareText(note, fmt.Sprintf("%s 국문 설명 %d", contextLabel, index+1), 5000, true, normalized); err != nil {
			return err
		}
	}
	return nil
}

func validateSoftwareTechnologies(values []string, contextLabel string, normalized bool) error {
	if len(values) == 0 {
		return fmt.Errorf("%s에는 하나 이상의 기술이 필요합니다", contextLabel)
	}
	if len(values) > maxSoftwareTechnologies {
		return fmt.Errorf("%s의 기술은 최대 %d개까지 입력할 수 있습니다", contextLabel, maxSoftwareTechnologies)
	}
	for index, value := range values {
		if err := validateSoftwareText(value, fmt.Sprintf("%s 기술 %d", contextLabel, index+1), 200, true, normalized); err != nil {
			return err
		}
	}
	return nil
}

func validateSoftwareText(value, contextLabel string, maximum int, required, normalized bool) error {
	if !utf8.ValidString(value) {
		return fmt.Errorf("%s의 문자열 인코딩이 올바르지 않습니다", contextLabel)
	}
	trimmed := strings.TrimSpace(value)
	if required && trimmed == "" {
		return fmt.Errorf("%s을(를) 입력해 주세요", contextLabel)
	}
	if normalized && value != trimmed {
		return fmt.Errorf("%s의 앞뒤 공백을 제거해 주세요", contextLabel)
	}
	if utf8.RuneCountInString(value) > maximum {
		return fmt.Errorf("%s은(는) %d자 이하여야 합니다", contextLabel, maximum)
	}
	for _, character := range value {
		if unicode.IsControl(character) && character != '\n' && character != '\t' {
			return fmt.Errorf("%s에는 제어 문자를 사용할 수 없습니다", contextLabel)
		}
	}
	return nil
}

func validateSoftwareURL(value, contextLabel string) error {
	if value == "" || len(value) > 4096 || strings.IndexFunc(value, unicode.IsControl) >= 0 {
		return fmt.Errorf("%s에는 유효한 http(s) URL이 필요합니다", contextLabel)
	}
	parsed, err := url.Parse(value)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return fmt.Errorf("%s에는 유효한 http(s) URL이 필요합니다", contextLabel)
	}
	return nil
}

func normaliseSoftwareStringList(values []string) []string {
	result := make([]string, len(values))
	for index, value := range values {
		result[index] = strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(value, "\r\n", "\n"), "\r", "\n"))
	}
	return result
}

func nextSoftwareMediaName(softwareID string, mediaIndex int, extension string, used map[string]bool) string {
	base := fmt.Sprintf("software-%s-%02d", softwareID, mediaIndex+1)
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
