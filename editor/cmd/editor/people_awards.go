package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Gonie-Gonie/Gonie-Gonie.github.io/editor/internal/dataversion"
	"os"
	"strings"
)

const (
	maxPeopleItems = 5_000
	maxPersonNotes = 100
	maxAwardItems  = 1_000
)

// PeopleItemSummary is the flat person shape returned to Wails. EditorKey is
// an opaque snapshot handle; ID is a stable data relationship key and is not
// intended to be edited directly in the UI.
type PeopleItemSummary struct {
	EditorKey string   `json:"editor_key"`
	ID        string   `json:"id"`
	NameEN    string   `json:"name_en"`
	NameKO    string   `json:"name_ko"`
	IsSelf    bool     `json:"is_self"`
	NotesEN   []string `json:"notes_en"`
	NotesKO   []string `json:"notes_ko"`
}

type PersonInput struct {
	// ID is normally hidden from users. It lets a newly-created person be
	// referenced by a publication or patent in the same transaction. Existing IDs are
	// immutable and may only be omitted or repeated exactly.
	ID      string   `json:"id,omitempty"`
	NameEN  string   `json:"name_en"`
	NameKO  string   `json:"name_ko"`
	IsSelf  bool     `json:"is_self"`
	NotesEN []string `json:"notes_en"`
	NotesKO []string `json:"notes_ko"`
}

type PersonSaveItem struct {
	EditorKey string       `json:"editor_key,omitempty"`
	Person    *PersonInput `json:"person,omitempty"`
}

type personDocumentItem struct {
	ID      string
	NameEN  string
	NameKO  string
	IsSelf  bool
	NotesEN []string
	NotesKO []string
}

type personRow struct {
	Key  string
	Raw  json.RawMessage
	Item personDocumentItem
}

type peopleSnapshot struct {
	Raw  []byte
	Rows []personRow
}

// AwardItemSummary is the flat award shape returned to Wails. As with people,
// the stable ID is exposed for relationship selectors but not direct editing.
type AwardItemSummary struct {
	PDF            string `json:"pdf,omitempty"`
	EditorKey      string `json:"editor_key"`
	VisibleInCV    bool   `json:"visible_in_CV"`
	ID             string `json:"id"`
	Date           string `json:"date"`
	TitleEN        string `json:"title_en"`
	TitleKO        string `json:"title_ko"`
	OrganizationEN string `json:"organization_en"`
	OrganizationKO string `json:"organization_ko"`
}

type AwardInput struct {
	PDF            *string `json:"pdf,omitempty"`
	ID             string  `json:"id,omitempty"`
	Date           string  `json:"date"`
	TitleEN        string  `json:"title_en"`
	TitleKO        string  `json:"title_ko"`
	OrganizationEN string  `json:"organization_en"`
	OrganizationKO string  `json:"organization_ko"`
}

type AwardSaveItem struct {
	EditorKey   string      `json:"editor_key,omitempty"`
	Award       *AwardInput `json:"award,omitempty"`
	VisibleInCV *bool       `json:"visible_in_CV,omitempty"`
}

type awardDocumentItem struct {
	PDF            string
	ID             string
	Date           string
	TitleEN        string
	TitleKO        string
	OrganizationEN string
	OrganizationKO string
}

type awardRow struct {
	Key  string
	Raw  json.RawMessage
	Item awardDocumentItem
}

type awardSnapshot struct {
	Raw  []byte
	Rows []awardRow
}

func readPeople(path string) (peopleSnapshot, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return peopleSnapshot{}, fmt.Errorf("people.json 읽기 실패: %w", err)
	}
	return readPeopleBytes(raw)
}

func readPeopleBytes(raw []byte) (peopleSnapshot, error) {
	rows, err := decodePeopleAwardsArray(raw, "people.json")
	if err != nil {
		return peopleSnapshot{}, err
	}
	if len(rows) > maxPeopleItems {
		return peopleSnapshot{}, fmt.Errorf("사람은 최대 %d명까지 관리할 수 있습니다", maxPeopleItems)
	}

	snapshot := peopleSnapshot{
		Raw:  append([]byte(nil), raw...),
		Rows: make([]personRow, 0, len(rows)),
	}
	usedIDs := make(map[string]bool, len(rows))
	selfCount := 0
	for index, rowRaw := range rows {
		contextLabel := fmt.Sprintf("people.json %d번째 항목", index+1)
		item, err := parsePersonItem(rowRaw, contextLabel)
		if err != nil {
			return peopleSnapshot{}, err
		}
		if usedIDs[item.ID] {
			return peopleSnapshot{}, fmt.Errorf("people.json에 중복 ID '%s'이(가) 있습니다", item.ID)
		}
		usedIDs[item.ID] = true
		if item.IsSelf {
			selfCount++
		}
		hash := revisionOf(rowRaw)
		snapshot.Rows = append(snapshot.Rows, personRow{
			Key:  fmt.Sprintf("%d:%s", index, hash[:20]),
			Raw:  cloneRawMessage(rowRaw),
			Item: item,
		})
	}
	if selfCount != 1 {
		return peopleSnapshot{}, errors.New("people.json에는 is_self가 true인 사람이 정확히 한 명 있어야 합니다")
	}
	return snapshot, nil
}

func parsePersonItem(raw json.RawMessage, contextLabel string) (personDocumentItem, error) {
	fields, err := decodeObject(raw)
	if err != nil {
		return personDocumentItem{}, fmt.Errorf("%s은 JSON 객체여야 합니다", contextLabel)
	}
	for _, field := range []string{"id", "name_en", "name_ko", "is_self", "notes_en", "notes_ko"} {
		if _, exists := fields[field]; !exists {
			return personDocumentItem{}, fmt.Errorf("%s에 %s 필드가 없습니다", contextLabel, field)
		}
	}

	id, err := decodeRequiredString(fields["id"], contextLabel+" ID")
	if err != nil {
		return personDocumentItem{}, err
	}
	if !taxonomyIDPattern.MatchString(id) {
		return personDocumentItem{}, fmt.Errorf("%s의 ID '%s' 형식이 올바르지 않습니다", contextLabel, id)
	}
	nameEN, err := decodeRequiredString(fields["name_en"], contextLabel+" 영문 이름")
	if err != nil {
		return personDocumentItem{}, err
	}
	nameKO, err := decodeRequiredString(fields["name_ko"], contextLabel+" 국문 이름")
	if err != nil {
		return personDocumentItem{}, err
	}
	isSelf, err := decodeRequiredBool(fields["is_self"], contextLabel+" is_self")
	if err != nil {
		return personDocumentItem{}, err
	}
	notesEN, err := decodeRequiredStringArray(fields["notes_en"], contextLabel+" notes_en")
	if err != nil {
		return personDocumentItem{}, err
	}
	notesKO, err := decodeRequiredStringArray(fields["notes_ko"], contextLabel+" notes_ko")
	if err != nil {
		return personDocumentItem{}, err
	}

	item := personDocumentItem{
		ID: id, NameEN: nameEN, NameKO: nameKO, IsSelf: isSelf,
		NotesEN: notesEN, NotesKO: notesKO,
	}
	if err := validatePersonDocument(item, contextLabel, false); err != nil {
		return personDocumentItem{}, err
	}
	return item, nil
}

func peopleItemSummaries(snapshot peopleSnapshot) []PeopleItemSummary {
	items := make([]PeopleItemSummary, 0, len(snapshot.Rows))
	for _, row := range snapshot.Rows {
		items = append(items, PeopleItemSummary{
			EditorKey: row.Key,
			ID:        row.Item.ID,
			NameEN:    row.Item.NameEN,
			NameKO:    row.Item.NameKO,
			IsSelf:    row.Item.IsSelf,
			NotesEN:   append([]string{}, row.Item.NotesEN...),
			NotesKO:   append([]string{}, row.Item.NotesKO...),
		})
	}
	return items
}

func peopleIDSet(snapshot peopleSnapshot) map[string]bool {
	ids := make(map[string]bool, len(snapshot.Rows))
	for _, row := range snapshot.Rows {
		ids[row.Item.ID] = true
	}
	return ids
}

func readAwards(path string) (awardSnapshot, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return awardSnapshot{}, fmt.Errorf("awards.json 읽기 실패: %w", err)
	}
	return readAwardsBytes(raw)
}

func readAwardsBytes(raw []byte) (awardSnapshot, error) {
	rows, err := decodePeopleAwardsArray(raw, "awards.json")
	if err != nil {
		return awardSnapshot{}, err
	}
	if len(rows) > maxAwardItems {
		return awardSnapshot{}, fmt.Errorf("수상은 최대 %d개까지 관리할 수 있습니다", maxAwardItems)
	}

	snapshot := awardSnapshot{
		Raw:  append([]byte(nil), raw...),
		Rows: make([]awardRow, 0, len(rows)),
	}
	usedIDs := make(map[string]bool, len(rows))
	for index, rowRaw := range rows {
		contextLabel := fmt.Sprintf("awards.json %d번째 항목", index+1)
		item, err := parseAwardItem(rowRaw, contextLabel)
		if err != nil {
			return awardSnapshot{}, err
		}
		if _, err := visibleInCV(rowRaw, contextLabel); err != nil {
			return awardSnapshot{}, err
		}
		if usedIDs[item.ID] {
			return awardSnapshot{}, fmt.Errorf("awards.json에 중복 ID '%s'이(가) 있습니다", item.ID)
		}
		usedIDs[item.ID] = true
		hash := revisionOf(rowRaw)
		snapshot.Rows = append(snapshot.Rows, awardRow{
			Key:  fmt.Sprintf("%d:%s", index, hash[:20]),
			Raw:  cloneRawMessage(rowRaw),
			Item: item,
		})
	}
	return snapshot, nil
}

func parseAwardItem(raw json.RawMessage, contextLabel string) (awardDocumentItem, error) {
	fields, err := decodeObject(raw)
	if err != nil {
		return awardDocumentItem{}, fmt.Errorf("%s은 JSON 객체여야 합니다", contextLabel)
	}
	for _, field := range []string{"id", "date", "title_en", "title_ko", "organization_en", "organization_ko"} {
		if _, exists := fields[field]; !exists {
			return awardDocumentItem{}, fmt.Errorf("%s에 %s 필드가 없습니다", contextLabel, field)
		}
	}

	values := make(map[string]string, 6)
	for _, field := range []string{"id", "date", "title_en", "title_ko", "organization_en", "organization_ko"} {
		value, err := decodeRequiredString(fields[field], contextLabel+" "+field)
		if err != nil {
			return awardDocumentItem{}, err
		}
		values[field] = value
	}
	if !taxonomyIDPattern.MatchString(values["id"]) {
		return awardDocumentItem{}, fmt.Errorf("%s의 ID '%s' 형식이 올바르지 않습니다", contextLabel, values["id"])
	}
	item := awardDocumentItem{
		ID:             values["id"],
		Date:           values["date"],
		TitleEN:        values["title_en"],
		TitleKO:        values["title_ko"],
		OrganizationEN: values["organization_en"],
		OrganizationKO: values["organization_ko"],
	}
	if rawPDF, exists := fields["pdf"]; exists {
		item.PDF, err = decodeRequiredString(rawPDF, contextLabel+" pdf")
		if err != nil {
			return awardDocumentItem{}, err
		}
	}
	if err := validateAwardDocument(item, contextLabel, false); err != nil {
		return awardDocumentItem{}, err
	}
	return item, nil
}

func awardItemSummaries(snapshot awardSnapshot) []AwardItemSummary {
	items := make([]AwardItemSummary, 0, len(snapshot.Rows))
	for _, row := range snapshot.Rows {
		items = append(items, AwardItemSummary{
			PDF:            row.Item.PDF,
			EditorKey:      row.Key,
			VisibleInCV:    mustVisibleInCV(row.Raw),
			ID:             row.Item.ID,
			Date:           row.Item.Date,
			TitleEN:        row.Item.TitleEN,
			TitleKO:        row.Item.TitleKO,
			OrganizationEN: row.Item.OrganizationEN,
			OrganizationKO: row.Item.OrganizationKO,
		})
	}
	return items
}

func awardIDSet(snapshot awardSnapshot) map[string]bool {
	ids := make(map[string]bool, len(snapshot.Rows))
	for _, row := range snapshot.Rows {
		ids[row.Item.ID] = true
	}
	return ids
}

// buildPeopleSaveLocked prepares people.json but does not write it. Request
// order is the display order; omission deletes a row and key-only preserves it.
func (a *App) buildPeopleSaveLocked(current peopleSnapshot, request []PersonSaveItem) ([]byte, error) {
	if len(request) > maxPeopleItems {
		return nil, fmt.Errorf("사람은 최대 %d명까지 저장할 수 있습니다", maxPeopleItems)
	}
	currentByKey := make(map[string]personRow, len(current.Rows))
	usedIDs := make(map[string]bool, len(current.Rows)+len(request))
	for _, row := range current.Rows {
		currentByKey[row.Key] = row
		usedIDs[row.Item.ID] = true
	}

	rows := make([]json.RawMessage, 0, len(request))
	usedExisting := make(map[string]bool)
	selfCount := 0
	documentUnchanged := len(request) == len(current.Rows)
	for index, requested := range request {
		contextLabel := fmt.Sprintf("사람 %d", index+1)
		var currentRow *personRow
		if requested.EditorKey != "" {
			row, exists := currentByKey[requested.EditorKey]
			if !exists {
				return nil, fmt.Errorf("%s이(가) 현재 people.json에 없습니다. 다시 불러와 주세요", contextLabel)
			}
			if usedExisting[requested.EditorKey] {
				return nil, fmt.Errorf("%s이(가) 중복되었습니다", contextLabel)
			}
			usedExisting[requested.EditorKey] = true
			currentRow = &row
			if requested.Person == nil {
				rows = append(rows, cloneRawMessage(row.Raw))
				if row.Item.IsSelf {
					selfCount++
				}
				if documentUnchanged && requested.EditorKey != current.Rows[index].Key {
					documentUnchanged = false
				}
				continue
			}
		} else if requested.Person == nil {
			return nil, fmt.Errorf("%s의 person 입력이 없습니다", contextLabel)
		}

		input := *requested.Person
		if currentRow != nil && input.ID != "" && input.ID != currentRow.Item.ID {
			return nil, fmt.Errorf("%s의 기존 ID '%s'은(는) 변경할 수 없습니다", contextLabel, currentRow.Item.ID)
		}
		if currentRow != nil && personInputExactlyMatches(input, currentRow.Item) {
			rows = append(rows, cloneRawMessage(currentRow.Raw))
			if currentRow.Item.IsSelf {
				selfCount++
			}
			if documentUnchanged && requested.EditorKey != current.Rows[index].Key {
				documentUnchanged = false
			}
			continue
		}

		normalized, err := normalisePersonInput(input, contextLabel)
		if err != nil {
			return nil, err
		}
		id := ""
		if currentRow != nil {
			id = currentRow.Item.ID
			preserveUnchangedPersonValues(&normalized, input, currentRow.Item)
		} else if normalized.ID != "" {
			id = normalized.ID
			if !taxonomyIDPattern.MatchString(id) {
				return nil, fmt.Errorf("%s의 제안 ID '%s' 형식이 올바르지 않습니다", contextLabel, id)
			}
			if usedIDs[id] {
				return nil, fmt.Errorf("%s의 제안 ID '%s'이(가) 이미 사용 중입니다", contextLabel, id)
			}
		} else {
			id, err = newTaxonomyID(normalized.NameEN, usedIDs)
			if err != nil {
				return nil, fmt.Errorf("%s ID 생성 실패: %w", contextLabel, err)
			}
		}
		normalized.ID = id
		usedIDs[id] = true
		document := personDocumentItem{
			ID: id, NameEN: normalized.NameEN, NameKO: normalized.NameKO,
			IsSelf: normalized.IsSelf, NotesEN: normalized.NotesEN, NotesKO: normalized.NotesKO,
		}
		if currentRow != nil && personDocumentMatches(document, currentRow.Item) {
			rows = append(rows, cloneRawMessage(currentRow.Raw))
			if document.IsSelf {
				selfCount++
			}
			if documentUnchanged && requested.EditorKey != current.Rows[index].Key {
				documentUnchanged = false
			}
			continue
		}
		rowRaw, err := marshalPersonItem(currentRow, document)
		if err != nil {
			return nil, fmt.Errorf("%s 인코딩 실패: %w", contextLabel, err)
		}
		rows = append(rows, rowRaw)
		if document.IsSelf {
			selfCount++
		}
		documentUnchanged = false
	}
	if selfCount != 1 {
		return nil, errors.New("사람 목록에는 '본인'이 정확히 한 명 있어야 합니다")
	}
	if documentUnchanged && bytes.HasPrefix(bytes.TrimSpace(current.Raw), []byte("[")) {
		return append([]byte(nil), current.Raw...), nil
	}
	return marshalPeopleAwardsRows(current.Raw, rows, "people.json")
}

// buildAwardsSaveLocked prepares awards.json without touching the repository.
func (a *App) buildAwardsSaveLocked(current awardSnapshot, request []AwardSaveItem) ([]byte, error) {
	if len(request) > maxAwardItems {
		return nil, fmt.Errorf("수상은 최대 %d개까지 저장할 수 있습니다", maxAwardItems)
	}
	currentByKey := make(map[string]awardRow, len(current.Rows))
	usedIDs := make(map[string]bool, len(current.Rows)+len(request))
	for _, row := range current.Rows {
		currentByKey[row.Key] = row
		usedIDs[row.Item.ID] = true
	}

	rows := make([]json.RawMessage, 0, len(request))
	usedExisting := make(map[string]bool)
	documentUnchanged := len(request) == len(current.Rows)
	appendVisible := func(raw json.RawMessage, requested *bool) error {
		updated, err := withVisibleInCV(raw, requested)
		if err != nil {
			return err
		}
		if !bytes.Equal(updated, raw) {
			documentUnchanged = false
		}
		rows = append(rows, updated)
		return nil
	}
	for index, requested := range request {
		contextLabel := fmt.Sprintf("수상 %d", index+1)
		var currentRow *awardRow
		if requested.EditorKey != "" {
			row, exists := currentByKey[requested.EditorKey]
			if !exists {
				return nil, fmt.Errorf("%s이(가) 현재 awards.json에 없습니다. 다시 불러와 주세요", contextLabel)
			}
			if usedExisting[requested.EditorKey] {
				return nil, fmt.Errorf("%s이(가) 중복되었습니다", contextLabel)
			}
			usedExisting[requested.EditorKey] = true
			currentRow = &row
			if requested.Award == nil {
				if err := appendVisible(row.Raw, requested.VisibleInCV); err != nil {
					return nil, fmt.Errorf("%s visible_in_CV 인코딩 실패: %w", contextLabel, err)
				}
				if documentUnchanged && requested.EditorKey != current.Rows[index].Key {
					documentUnchanged = false
				}
				continue
			}
		} else if requested.Award == nil {
			return nil, fmt.Errorf("%s의 award 입력이 없습니다", contextLabel)
		}

		input := *requested.Award
		if input.PDF == nil && currentRow != nil {
			filename := currentRow.Item.PDF
			input.PDF = &filename
		}
		if currentRow != nil && input.ID != "" && input.ID != currentRow.Item.ID {
			return nil, fmt.Errorf("%s의 기존 ID '%s'은(는) 변경할 수 없습니다", contextLabel, currentRow.Item.ID)
		}
		if currentRow != nil && awardInputExactlyMatches(input, currentRow.Item) {
			if err := appendVisible(currentRow.Raw, requested.VisibleInCV); err != nil {
				return nil, fmt.Errorf("%s visible_in_CV 인코딩 실패: %w", contextLabel, err)
			}
			if documentUnchanged && requested.EditorKey != current.Rows[index].Key {
				documentUnchanged = false
			}
			continue
		}

		normalized, err := normaliseAwardInput(input, contextLabel)
		if err != nil {
			return nil, err
		}
		id := ""
		if currentRow != nil {
			id = currentRow.Item.ID
			preserveUnchangedAwardValues(&normalized, input, currentRow.Item)
		} else if normalized.ID != "" {
			id = normalized.ID
			if !taxonomyIDPattern.MatchString(id) {
				return nil, fmt.Errorf("%s의 제안 ID '%s' 형식이 올바르지 않습니다", contextLabel, id)
			}
			if usedIDs[id] {
				return nil, fmt.Errorf("%s의 제안 ID '%s'이(가) 이미 사용 중입니다", contextLabel, id)
			}
		} else {
			id, err = newTaxonomyID(normalized.TitleEN, usedIDs)
			if err != nil {
				return nil, fmt.Errorf("%s ID 생성 실패: %w", contextLabel, err)
			}
		}
		normalized.ID = id
		usedIDs[id] = true
		document := awardDocumentItem{
			ID: id, Date: normalized.Date, TitleEN: normalized.TitleEN, TitleKO: normalized.TitleKO,
			OrganizationEN: normalized.OrganizationEN, OrganizationKO: normalized.OrganizationKO,
			PDF: awardPDF(normalized),
		}
		if currentRow != nil && awardDocumentMatches(document, currentRow.Item) {
			if err := appendVisible(currentRow.Raw, requested.VisibleInCV); err != nil {
				return nil, fmt.Errorf("%s visible_in_CV 인코딩 실패: %w", contextLabel, err)
			}
			if documentUnchanged && requested.EditorKey != current.Rows[index].Key {
				documentUnchanged = false
			}
			continue
		}
		rowRaw, err := marshalAwardItem(currentRow, document)
		if err != nil {
			return nil, fmt.Errorf("%s 인코딩 실패: %w", contextLabel, err)
		}
		if err := appendVisible(rowRaw, visibleInCVForSave(requested.VisibleInCV, currentRow == nil)); err != nil {
			return nil, fmt.Errorf("%s visible_in_CV 인코딩 실패: %w", contextLabel, err)
		}
		documentUnchanged = false
	}
	if documentUnchanged && bytes.HasPrefix(bytes.TrimSpace(current.Raw), []byte("[")) {
		return append([]byte(nil), current.Raw...), nil
	}
	return marshalPeopleAwardsRows(current.Raw, rows, "awards.json")
}

func normalisePersonInput(input PersonInput, contextLabel string) (PersonInput, error) {
	if input.NotesEN == nil || input.NotesKO == nil {
		return PersonInput{}, fmt.Errorf("%s의 notes_en과 notes_ko는 JSON 배열이어야 합니다", contextLabel)
	}
	result := PersonInput{
		ID:      input.ID,
		NameEN:  normalisePeopleAwardsText(input.NameEN),
		NameKO:  normalisePeopleAwardsText(input.NameKO),
		IsSelf:  input.IsSelf,
		NotesEN: normalisePeopleAwardsStrings(input.NotesEN),
		NotesKO: normalisePeopleAwardsStrings(input.NotesKO),
	}
	if result.ID != "" && !taxonomyIDPattern.MatchString(result.ID) {
		return PersonInput{}, fmt.Errorf("%s의 제안 ID '%s' 형식이 올바르지 않습니다", contextLabel, result.ID)
	}
	document := personDocumentItem{
		ID: result.ID, NameEN: result.NameEN, NameKO: result.NameKO, IsSelf: result.IsSelf,
		NotesEN: result.NotesEN, NotesKO: result.NotesKO,
	}
	if err := validatePersonDocument(document, contextLabel, true); err != nil {
		return PersonInput{}, err
	}
	return result, nil
}

func normaliseAwardInput(input AwardInput, contextLabel string) (AwardInput, error) {
	result := AwardInput{
		PDF:            input.PDF,
		ID:             input.ID,
		Date:           strings.TrimSpace(input.Date),
		TitleEN:        normalisePeopleAwardsText(input.TitleEN),
		TitleKO:        normalisePeopleAwardsText(input.TitleKO),
		OrganizationEN: normalisePeopleAwardsText(input.OrganizationEN),
		OrganizationKO: normalisePeopleAwardsText(input.OrganizationKO),
	}
	if result.ID != "" && !taxonomyIDPattern.MatchString(result.ID) {
		return AwardInput{}, fmt.Errorf("%s의 제안 ID '%s' 형식이 올바르지 않습니다", contextLabel, result.ID)
	}
	document := awardDocumentItem{
		ID: result.ID, Date: result.Date, TitleEN: result.TitleEN, TitleKO: result.TitleKO,
		OrganizationEN: result.OrganizationEN, OrganizationKO: result.OrganizationKO,
		PDF: awardPDF(result),
	}
	if err := validateAwardDocument(document, contextLabel, true); err != nil {
		return AwardInput{}, err
	}
	return result, nil
}

func validatePersonDocument(item personDocumentItem, contextLabel string, normalized bool) error {
	if err := validateSoftwareText(item.NameEN, contextLabel+" 영문 이름", 300, false, normalized); err != nil {
		return err
	}
	if err := validateSoftwareText(item.NameKO, contextLabel+" 국문 이름", 300, false, normalized); err != nil {
		return err
	}
	if strings.TrimSpace(item.NameEN) == "" && strings.TrimSpace(item.NameKO) == "" {
		return fmt.Errorf("%s에는 영문 또는 국문 이름이 필요합니다", contextLabel)
	}
	if len(item.NotesEN) > maxPersonNotes || len(item.NotesKO) > maxPersonNotes {
		return fmt.Errorf("%s의 언어별 소속/메모는 최대 %d개까지 입력할 수 있습니다", contextLabel, maxPersonNotes)
	}
	// The two note arrays are intentionally independent. A hand-edited file may
	// have different counts for English and Korean, including either side empty.
	for index, note := range item.NotesEN {
		if err := validateSoftwareText(note, fmt.Sprintf("%s 영문 소속/메모 %d", contextLabel, index+1), 5_000, true, normalized); err != nil {
			return err
		}
	}
	for index, note := range item.NotesKO {
		if err := validateSoftwareText(note, fmt.Sprintf("%s 국문 소속/메모 %d", contextLabel, index+1), 5_000, true, normalized); err != nil {
			return err
		}
	}
	return nil
}

func validateAwardDocument(item awardDocumentItem, contextLabel string, normalized bool) error {
	if err := validateDocumentFilename(item.PDF); err != nil {
		return fmt.Errorf("%s pdf: %w", contextLabel, err)
	}
	if _, err := parseCanonicalDate(item.Date); err != nil {
		return fmt.Errorf("%s의 날짜는 YYYY-MM-DD 형식의 실제 날짜여야 합니다", contextLabel)
	}
	for label, value := range map[string]string{
		"영문 수상명":  item.TitleEN,
		"국문 수상명":  item.TitleKO,
		"영문 수여기관": item.OrganizationEN,
		"국문 수여기관": item.OrganizationKO,
	} {
		if err := validateSoftwareText(value, contextLabel+" "+label, 1_000, false, normalized); err != nil {
			return err
		}
	}
	if strings.TrimSpace(item.TitleEN) == "" && strings.TrimSpace(item.TitleKO) == "" {
		return fmt.Errorf("%s에는 영문 또는 국문 수상명이 필요합니다", contextLabel)
	}
	if strings.TrimSpace(item.OrganizationEN) == "" && strings.TrimSpace(item.OrganizationKO) == "" {
		return fmt.Errorf("%s에는 영문 또는 국문 수여기관이 필요합니다", contextLabel)
	}
	return nil
}

func personInputExactlyMatches(input PersonInput, current personDocumentItem) bool {
	return (input.ID == "" || input.ID == current.ID) &&
		input.NameEN == current.NameEN && input.NameKO == current.NameKO && input.IsSelf == current.IsSelf &&
		equalSoftwareStrings(input.NotesEN, current.NotesEN) && equalSoftwareStrings(input.NotesKO, current.NotesKO)
}

func awardInputExactlyMatches(input AwardInput, current awardDocumentItem) bool {
	return (input.ID == "" || input.ID == current.ID) && input.Date == current.Date &&
		awardPDF(input) == current.PDF &&
		input.TitleEN == current.TitleEN && input.TitleKO == current.TitleKO &&
		input.OrganizationEN == current.OrganizationEN && input.OrganizationKO == current.OrganizationKO
}

func personDocumentMatches(left, right personDocumentItem) bool {
	return left.ID == right.ID && left.NameEN == right.NameEN && left.NameKO == right.NameKO &&
		left.IsSelf == right.IsSelf && equalSoftwareStrings(left.NotesEN, right.NotesEN) &&
		equalSoftwareStrings(left.NotesKO, right.NotesKO)
}

func awardDocumentMatches(left, right awardDocumentItem) bool {
	return left.ID == right.ID && left.Date == right.Date && left.TitleEN == right.TitleEN &&
		left.PDF == right.PDF &&
		left.TitleKO == right.TitleKO && left.OrganizationEN == right.OrganizationEN &&
		left.OrganizationKO == right.OrganizationKO
}

func preserveUnchangedPersonValues(result *PersonInput, original PersonInput, current personDocumentItem) {
	if original.NameEN == current.NameEN {
		result.NameEN = current.NameEN
	}
	if original.NameKO == current.NameKO {
		result.NameKO = current.NameKO
	}
	if equalSoftwareStrings(original.NotesEN, current.NotesEN) {
		result.NotesEN = append([]string{}, current.NotesEN...)
	}
	if equalSoftwareStrings(original.NotesKO, current.NotesKO) {
		result.NotesKO = append([]string{}, current.NotesKO...)
	}
}

func preserveUnchangedAwardValues(result *AwardInput, original AwardInput, current awardDocumentItem) {
	if original.Date == current.Date {
		result.Date = current.Date
	}
	if original.TitleEN == current.TitleEN {
		result.TitleEN = current.TitleEN
	}
	if original.TitleKO == current.TitleKO {
		result.TitleKO = current.TitleKO
	}
	if original.OrganizationEN == current.OrganizationEN {
		result.OrganizationEN = current.OrganizationEN
	}
	if original.OrganizationKO == current.OrganizationKO {
		result.OrganizationKO = current.OrganizationKO
	}
}

func marshalPersonItem(current *personRow, item personDocumentItem) (json.RawMessage, error) {
	fields := map[string]json.RawMessage{}
	if current != nil {
		var err error
		fields, err = decodeObject(current.Raw)
		if err != nil {
			return nil, err
		}
	}
	merged, err := mergeObjectFields(fields, map[string]any{
		"id": item.ID, "name_en": item.NameEN, "name_ko": item.NameKO,
		"is_self": item.IsSelf, "notes_en": item.NotesEN, "notes_ko": item.NotesKO,
	})
	if err != nil {
		return nil, err
	}
	return json.Marshal(merged)
}

func marshalAwardItem(current *awardRow, item awardDocumentItem) (json.RawMessage, error) {
	fields := map[string]json.RawMessage{}
	if current != nil {
		var err error
		fields, err = decodeObject(current.Raw)
		if err != nil {
			return nil, err
		}
	}
	merged, err := mergeObjectFields(fields, map[string]any{
		"id": item.ID, "date": item.Date, "title_en": item.TitleEN, "title_ko": item.TitleKO,
		"organization_en": item.OrganizationEN, "organization_ko": item.OrganizationKO,
	})
	if err != nil {
		return nil, err
	}
	if item.PDF == "" {
		delete(merged, "pdf")
	} else {
		merged["pdf"], err = json.Marshal(item.PDF)
		if err != nil {
			return nil, err
		}
	}
	return json.Marshal(merged)
}

func awardPDF(input AwardInput) string {
	if input.PDF == nil {
		return ""
	}
	return *input.PDF
}

func decodePeopleAwardsArray(raw []byte, filename string) ([]json.RawMessage, error) {
	return dataversion.Array(raw, filename)
}

func decodeRequiredString(raw json.RawMessage, contextLabel string) (string, error) {
	var value *string
	if err := json.Unmarshal(raw, &value); err != nil || value == nil {
		return "", fmt.Errorf("%s은 문자열이어야 합니다", contextLabel)
	}
	return *value, nil
}

func decodeRequiredBool(raw json.RawMessage, contextLabel string) (bool, error) {
	var value *bool
	if err := json.Unmarshal(raw, &value); err != nil || value == nil {
		return false, fmt.Errorf("%s은 true 또는 false여야 합니다", contextLabel)
	}
	return *value, nil
}

func decodeRequiredStringArray(raw json.RawMessage, contextLabel string) ([]string, error) {
	var values []string
	if err := json.Unmarshal(raw, &values); err != nil || values == nil {
		return nil, fmt.Errorf("%s은 문자열 JSON 배열이어야 합니다", contextLabel)
	}
	return values, nil
}

func normalisePeopleAwardsText(value string) string {
	return strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(value, "\r\n", "\n"), "\r", "\n"))
}

func normalisePeopleAwardsStrings(values []string) []string {
	result := make([]string, len(values))
	for index, value := range values {
		result[index] = normalisePeopleAwardsText(value)
	}
	return result
}

func marshalPeopleAwardsRows(previous []byte, rows []json.RawMessage, filename string) ([]byte, error) {
	encoded, err := dataversion.MarshalArray(previous, rows, filename)
	if err != nil {
		return nil, fmt.Errorf("%s 인코딩 실패: %w", filename, err)
	}
	return encoded, nil
}
