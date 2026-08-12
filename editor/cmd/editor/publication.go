package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Gonie-Gonie/Gonie-Gonie.github.io/editor/internal/dataversion"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

var publicationTypes = map[string]bool{
	"international-journal":    true,
	"domestic-journal":         true,
	"international-conference": true,
	"domestic-conference":      true,
}

var publicationFields = []string{
	"title_en",
	"title_ko",
	"abstract_en",
	"abstract_ko",
	"keywords_en",
	"keywords_ko",
	"author_ids",
	"date",
	"under_review",
	"in_press",
	"publication_type",
	"topic",
	"award_id",
	"doi",
	"url",
	"note",
}

func settingsPublicationTopicIDs(settings SettingsDocument) map[string]bool {
	ids := make(map[string]bool, len(settings.Topics))
	for _, topic := range settings.Topics {
		ids[topic.ID] = true
	}
	return ids
}

// PublicationItemSummary is the complete editable publication shape exposed to
// Wails. EditorKey is an opaque handle for one row in the loaded snapshot;
// publications deliberately do not have persistent IDs of their own.
type PublicationItemSummary struct {
	EditorKey       string   `json:"editor_key"`
	VisibleInCV     bool     `json:"visible_in_CV"`
	TitleEN         string   `json:"title_en"`
	TitleKO         string   `json:"title_ko"`
	AbstractEN      string   `json:"abstract_en"`
	AbstractKO      string   `json:"abstract_ko"`
	KeywordsEN      []string `json:"keywords_en"`
	KeywordsKO      []string `json:"keywords_ko"`
	AuthorIDs       []string `json:"author_ids"`
	Date            string   `json:"date"`
	JournalID       string   `json:"journal_id"`
	ConferenceID    string   `json:"conference_id"`
	Volume          string   `json:"volume"`
	Issue           string   `json:"issue"`
	Pages           string   `json:"pages"`
	ELocator        string   `json:"elocator"`
	OnlineDate      string   `json:"online_date"`
	IssueDate       string   `json:"issue_date"`
	UnderReview     bool     `json:"under_review"`
	InPress         bool     `json:"in_press"`
	PublicationType string   `json:"publication_type"`
	Topic           string   `json:"topic"`
	AwardID         string   `json:"award_id"`
	DOI             string   `json:"doi"`
	URL             string   `json:"url"`
	Note            string   `json:"note"`
}

type PublicationInput struct {
	TitleEN         string   `json:"title_en"`
	TitleKO         string   `json:"title_ko"`
	AbstractEN      string   `json:"abstract_en"`
	AbstractKO      string   `json:"abstract_ko"`
	KeywordsEN      []string `json:"keywords_en"`
	KeywordsKO      []string `json:"keywords_ko"`
	AuthorIDs       []string `json:"author_ids"`
	Date            string   `json:"date"`
	JournalID       string   `json:"journal_id"`
	ConferenceID    string   `json:"conference_id"`
	Volume          string   `json:"volume"`
	Issue           string   `json:"issue"`
	Pages           string   `json:"pages"`
	ELocator        string   `json:"elocator"`
	OnlineDate      string   `json:"online_date"`
	IssueDate       string   `json:"issue_date"`
	UnderReview     bool     `json:"under_review"`
	InPress         bool     `json:"in_press"`
	PublicationType string   `json:"publication_type"`
	Topic           string   `json:"topic"`
	AwardID         string   `json:"award_id"`
	DOI             string   `json:"doi"`
	URL             string   `json:"url"`
	Note            string   `json:"note"`
}

// Existing publications use EditorKey plus Publication. A key-only item keeps
// its row byte-for-byte. New publications omit EditorKey, and omitting an
// existing row from the request deletes it.
type PublicationSaveItem struct {
	EditorKey   string            `json:"editor_key,omitempty"`
	Publication *PublicationInput `json:"publication,omitempty"`
	VisibleInCV *bool             `json:"visible_in_CV,omitempty"`
}

type publicationDocumentItem struct {
	TitleEN         string
	TitleKO         string
	AbstractEN      string
	AbstractKO      string
	KeywordsEN      []string
	KeywordsKO      []string
	AuthorIDs       []string
	Date            string
	JournalID       string
	ConferenceID    string
	Volume          string
	Issue           string
	Pages           string
	ELocator        string
	OnlineDate      string
	IssueDate       string
	UnderReview     bool
	InPress         bool
	PublicationType string
	Topic           string
	AwardID         string
	DOI             string
	URL             string
	Note            string
}

type publicationRow struct {
	Key  string
	Raw  json.RawMessage
	Item publicationDocumentItem
}

type publicationSnapshot struct {
	Raw  []byte
	Rows []publicationRow
}

func readPublications(path string, personIDs, awardIDs, topicIDs map[string]bool) (publicationSnapshot, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return publicationSnapshot{}, fmt.Errorf("publications.json 읽기 실패: %w", err)
	}
	return readPublicationsBytes(raw, personIDs, awardIDs, topicIDs)
}

func readPublicationsBytes(raw []byte, personIDs, awardIDs, topicIDs map[string]bool) (publicationSnapshot, error) {
	rows, err := dataversion.Array(raw, "publications.json")
	if err != nil {
		return publicationSnapshot{}, err
	}

	snapshot := publicationSnapshot{
		Raw:  append([]byte(nil), raw...),
		Rows: make([]publicationRow, 0, len(rows)),
	}
	for index, rowRaw := range rows {
		contextLabel := fmt.Sprintf("publications.json %d번째 논문", index+1)
		item, err := parsePublicationItem(rowRaw, personIDs, awardIDs, topicIDs, contextLabel)
		if err != nil {
			return publicationSnapshot{}, err
		}
		if _, err := visibleInCV(rowRaw, contextLabel); err != nil {
			return publicationSnapshot{}, err
		}
		hash := revisionOf(rowRaw)
		snapshot.Rows = append(snapshot.Rows, publicationRow{
			Key:  fmt.Sprintf("%d:%s", index, hash[:20]),
			Raw:  cloneRawMessage(rowRaw),
			Item: item,
		})
	}
	return snapshot, nil
}

func parsePublicationItem(
	raw json.RawMessage,
	personIDs, awardIDs, topicIDs map[string]bool,
	contextLabel string,
) (publicationDocumentItem, error) {
	fields, err := decodeObject(raw)
	if err != nil {
		return publicationDocumentItem{}, fmt.Errorf("%s은 JSON 객체여야 합니다", contextLabel)
	}
	if _, exists := fields["id"]; exists {
		return publicationDocumentItem{}, fmt.Errorf("%s에는 id 필드를 사용할 수 없습니다", contextLabel)
	}
	for _, field := range []string{"venue", "venue_en", "venue_kr"} {
		if _, exists := fields[field]; exists {
			return publicationDocumentItem{}, fmt.Errorf("%s의 %s 대신 저널·학술대회 연결과 서지 정보를 사용해 주세요", contextLabel, field)
		}
	}
	for _, name := range publicationFields {
		if _, exists := fields[name]; !exists {
			return publicationDocumentItem{}, fmt.Errorf("%s에 %s 필드가 없습니다", contextLabel, name)
		}
	}

	item := publicationDocumentItem{}
	if item.TitleEN, err = publicationStringField(fields, "title_en", contextLabel); err != nil {
		return publicationDocumentItem{}, err
	}
	if item.TitleKO, err = publicationStringField(fields, "title_ko", contextLabel); err != nil {
		return publicationDocumentItem{}, err
	}
	if item.AbstractEN, err = publicationStringField(fields, "abstract_en", contextLabel); err != nil {
		return publicationDocumentItem{}, err
	}
	if item.AbstractKO, err = publicationStringField(fields, "abstract_ko", contextLabel); err != nil {
		return publicationDocumentItem{}, err
	}
	if item.KeywordsEN, err = publicationStringArrayField(fields, "keywords_en", contextLabel); err != nil {
		return publicationDocumentItem{}, err
	}
	if item.KeywordsKO, err = publicationStringArrayField(fields, "keywords_ko", contextLabel); err != nil {
		return publicationDocumentItem{}, err
	}
	if item.AuthorIDs, err = publicationStringArrayField(fields, "author_ids", contextLabel); err != nil {
		return publicationDocumentItem{}, err
	}
	if item.Date, err = publicationStringField(fields, "date", contextLabel); err != nil {
		return publicationDocumentItem{}, err
	}
	for _, field := range []struct {
		name   string
		target *string
	}{
		{"journal_id", &item.JournalID}, {"conference_id", &item.ConferenceID},
		{"volume", &item.Volume}, {"issue", &item.Issue}, {"pages", &item.Pages},
		{"elocator", &item.ELocator}, {"online_date", &item.OnlineDate}, {"issue_date", &item.IssueDate},
	} {
		if _, exists := fields[field.name]; exists {
			if *field.target, err = publicationStringField(fields, field.name, contextLabel); err != nil {
				return publicationDocumentItem{}, err
			}
		}
	}
	if item.UnderReview, err = publicationBoolField(fields, "under_review", contextLabel); err != nil {
		return publicationDocumentItem{}, err
	}
	if item.InPress, err = publicationBoolField(fields, "in_press", contextLabel); err != nil {
		return publicationDocumentItem{}, err
	}
	if item.PublicationType, err = publicationStringField(fields, "publication_type", contextLabel); err != nil {
		return publicationDocumentItem{}, err
	}
	if item.Topic, err = publicationStringField(fields, "topic", contextLabel); err != nil {
		return publicationDocumentItem{}, err
	}
	if item.AwardID, err = publicationStringField(fields, "award_id", contextLabel); err != nil {
		return publicationDocumentItem{}, err
	}
	if item.DOI, err = publicationStringField(fields, "doi", contextLabel); err != nil {
		return publicationDocumentItem{}, err
	}
	if item.URL, err = publicationStringField(fields, "url", contextLabel); err != nil {
		return publicationDocumentItem{}, err
	}
	if item.Note, err = publicationStringField(fields, "note", contextLabel); err != nil {
		return publicationDocumentItem{}, err
	}
	if err := validatePublication(item, personIDs, awardIDs, topicIDs, contextLabel); err != nil {
		return publicationDocumentItem{}, err
	}
	return item, nil
}

func publicationStringField(fields map[string]json.RawMessage, name, contextLabel string) (string, error) {
	var value *string
	if err := json.Unmarshal(fields[name], &value); err != nil || value == nil {
		return "", fmt.Errorf("%s의 %s 필드는 문자열이어야 합니다", contextLabel, name)
	}
	return *value, nil
}

func publicationStringArrayField(fields map[string]json.RawMessage, name, contextLabel string) ([]string, error) {
	var value *[]string
	if err := json.Unmarshal(fields[name], &value); err != nil || value == nil {
		return nil, fmt.Errorf("%s의 %s 필드는 문자열 배열이어야 합니다", contextLabel, name)
	}
	return clonePublicationStrings(*value), nil
}

func publicationBoolField(fields map[string]json.RawMessage, name, contextLabel string) (bool, error) {
	var value *bool
	if err := json.Unmarshal(fields[name], &value); err != nil || value == nil {
		return false, fmt.Errorf("%s의 %s 필드는 Boolean이어야 합니다", contextLabel, name)
	}
	return *value, nil
}

func publicationItemSummaries(snapshot publicationSnapshot) []PublicationItemSummary {
	items := make([]PublicationItemSummary, 0, len(snapshot.Rows))
	for _, row := range snapshot.Rows {
		items = append(items, PublicationItemSummary{
			EditorKey:       row.Key,
			VisibleInCV:     mustVisibleInCV(row.Raw),
			TitleEN:         row.Item.TitleEN,
			TitleKO:         row.Item.TitleKO,
			AbstractEN:      row.Item.AbstractEN,
			AbstractKO:      row.Item.AbstractKO,
			KeywordsEN:      clonePublicationStrings(row.Item.KeywordsEN),
			KeywordsKO:      clonePublicationStrings(row.Item.KeywordsKO),
			AuthorIDs:       clonePublicationStrings(row.Item.AuthorIDs),
			Date:            row.Item.Date,
			JournalID:       row.Item.JournalID,
			ConferenceID:    row.Item.ConferenceID,
			Volume:          row.Item.Volume,
			Issue:           row.Item.Issue,
			Pages:           row.Item.Pages,
			ELocator:        row.Item.ELocator,
			OnlineDate:      row.Item.OnlineDate,
			IssueDate:       row.Item.IssueDate,
			UnderReview:     row.Item.UnderReview,
			InPress:         row.Item.InPress,
			PublicationType: row.Item.PublicationType,
			Topic:           row.Item.Topic,
			AwardID:         row.Item.AwardID,
			DOI:             row.Item.DOI,
			URL:             row.Item.URL,
			Note:            row.Item.Note,
		})
	}
	return items
}

// buildPublicationSaveLocked prepares publications.json without writing it.
// Rows are sorted by the same contract as the website/build exports: review
// status, press status, date descending, then localized title ascending. The
// sort is stable, so complete ties retain request order.
func buildPublicationSaveLocked(
	current publicationSnapshot,
	request []PublicationSaveItem,
	personIDs, awardIDs, topicIDs map[string]bool,
	catalogs ...SettingsDocument,
) ([]byte, error) {
	currentByKey := make(map[string]publicationRow, len(current.Rows))
	for _, row := range current.Rows {
		currentByKey[row.Key] = row
	}
	usedExisting := make(map[string]bool, len(current.Rows))

	type preparedPublicationRow struct {
		raw  json.RawMessage
		item publicationDocumentItem
	}
	prepared := make([]preparedPublicationRow, 0, len(request))
	for index, requested := range request {
		contextLabel := fmt.Sprintf("논문 %d", index+1)
		var currentRow *publicationRow
		if requested.EditorKey != "" {
			row, exists := currentByKey[requested.EditorKey]
			if !exists {
				return nil, fmt.Errorf("%s이 현재 publications.json에 없습니다. 다시 불러와 주세요", contextLabel)
			}
			if usedExisting[requested.EditorKey] {
				return nil, fmt.Errorf("%s이 중복되었습니다", contextLabel)
			}
			usedExisting[requested.EditorKey] = true
			currentRow = &row
			if requested.Publication == nil {
				rowRaw, err := withVisibleInCV(row.Raw, requested.VisibleInCV)
				if err != nil {
					return nil, fmt.Errorf("%s visible_in_CV 인코딩 실패: %w", contextLabel, err)
				}
				prepared = append(prepared, preparedPublicationRow{raw: rowRaw, item: row.Item})
				continue
			}
		} else if requested.Publication == nil {
			return nil, fmt.Errorf("%s의 publication 입력이 없습니다", contextLabel)
		}

		if currentRow != nil && publicationInputExactlyMatches(*requested.Publication, currentRow.Item) {
			rowRaw, err := withVisibleInCV(currentRow.Raw, requested.VisibleInCV)
			if err != nil {
				return nil, fmt.Errorf("%s visible_in_CV 인코딩 실패: %w", contextLabel, err)
			}
			prepared = append(prepared, preparedPublicationRow{
				raw: rowRaw, item: currentRow.Item,
			})
			continue
		}
		item, err := normalisePublicationInput(*requested.Publication, personIDs, awardIDs, topicIDs, contextLabel)
		if err != nil {
			return nil, err
		}
		if currentRow != nil {
			preserveUnchangedPublicationValues(&item, *requested.Publication, currentRow.Item)
			if err := validatePublication(item, personIDs, awardIDs, topicIDs, contextLabel); err != nil {
				return nil, err
			}
		}
		if currentRow != nil && publicationDocumentEqual(item, currentRow.Item) {
			rowRaw, err := withVisibleInCV(currentRow.Raw, requested.VisibleInCV)
			if err != nil {
				return nil, fmt.Errorf("%s visible_in_CV 인코딩 실패: %w", contextLabel, err)
			}
			prepared = append(prepared, preparedPublicationRow{
				raw: rowRaw, item: currentRow.Item,
			})
			continue
		}
		rowRaw, err := marshalPublicationItem(currentRow, item)
		if err != nil {
			return nil, fmt.Errorf("%s 인코딩 실패: %w", contextLabel, err)
		}
		rowRaw, err = withVisibleInCV(rowRaw, visibleInCVForSave(requested.VisibleInCV, currentRow == nil))
		if err != nil {
			return nil, fmt.Errorf("%s visible_in_CV 인코딩 실패: %w", contextLabel, err)
		}
		prepared = append(prepared, preparedPublicationRow{raw: rowRaw, item: item})
	}

	if len(catalogs) > 0 {
		for index := range prepared {
			item := &prepared[index]
			date := effectivePublicationDate(item.item, catalogs[0])
			if date == item.item.Date {
				continue
			}
			item.item.Date = date
			fields, err := decodeObject(item.raw)
			if err != nil {
				return nil, err
			}
			fields, err = mergeObjectFields(fields, map[string]any{"date": date})
			if err != nil {
				return nil, err
			}
			item.raw, err = json.Marshal(fields)
			if err != nil {
				return nil, err
			}
		}
	}

	sort.SliceStable(prepared, func(left, right int) bool {
		return publicationSortLess(prepared[left].item, prepared[right].item)
	})
	rows := make([]json.RawMessage, 0, len(prepared))
	for _, row := range prepared {
		rows = append(rows, row.raw)
	}
	encoded, err := dataversion.MarshalArray(current.Raw, rows, "publications.json")
	if err != nil {
		return nil, fmt.Errorf("publications.json 인코딩 실패: %w", err)
	}
	return encoded, nil
}

func normalisePublicationInput(
	input PublicationInput,
	personIDs, awardIDs, topicIDs map[string]bool,
	contextLabel string,
) (publicationDocumentItem, error) {
	if input.KeywordsEN == nil || input.KeywordsKO == nil || input.AuthorIDs == nil {
		return publicationDocumentItem{}, fmt.Errorf("%s의 keywords_en, keywords_ko, author_ids는 JSON 배열이어야 합니다", contextLabel)
	}
	item := publicationDocumentItem{
		TitleEN:         normalisePublicationText(input.TitleEN),
		TitleKO:         normalisePublicationText(input.TitleKO),
		AbstractEN:      normalisePublicationText(input.AbstractEN),
		AbstractKO:      normalisePublicationText(input.AbstractKO),
		KeywordsEN:      normalisePublicationStrings(input.KeywordsEN),
		KeywordsKO:      normalisePublicationStrings(input.KeywordsKO),
		AuthorIDs:       normalisePublicationStrings(input.AuthorIDs),
		Date:            strings.TrimSpace(input.Date),
		JournalID:       strings.TrimSpace(input.JournalID),
		ConferenceID:    strings.TrimSpace(input.ConferenceID),
		Volume:          strings.TrimSpace(input.Volume),
		Issue:           strings.TrimSpace(input.Issue),
		Pages:           strings.TrimSpace(input.Pages),
		ELocator:        strings.TrimSpace(input.ELocator),
		OnlineDate:      strings.TrimSpace(input.OnlineDate),
		IssueDate:       strings.TrimSpace(input.IssueDate),
		UnderReview:     input.UnderReview,
		InPress:         input.InPress,
		PublicationType: strings.TrimSpace(input.PublicationType),
		Topic:           strings.TrimSpace(input.Topic),
		AwardID:         strings.TrimSpace(input.AwardID),
		DOI:             strings.TrimSpace(input.DOI),
		URL:             strings.TrimSpace(input.URL),
		Note:            normalisePublicationText(input.Note),
	}
	if err := validatePublication(item, personIDs, awardIDs, topicIDs, contextLabel); err != nil {
		return publicationDocumentItem{}, err
	}
	return item, nil
}

func normalisePublicationText(value string) string {
	value = strings.ReplaceAll(strings.ReplaceAll(value, "\r\n", "\n"), "\r", "\n")
	return strings.TrimSpace(value)
}

func normalisePublicationStrings(values []string) []string {
	result := make([]string, len(values))
	for index, value := range values {
		result[index] = normalisePublicationText(value)
	}
	return result
}

func clonePublicationStrings(values []string) []string {
	if values == nil {
		return nil
	}
	result := make([]string, len(values))
	copy(result, values)
	return result
}

func preserveUnchangedPublicationValues(
	result *publicationDocumentItem,
	original PublicationInput,
	current publicationDocumentItem,
) {
	stringFields := []struct {
		original string
		current  string
		target   *string
	}{
		{original.TitleEN, current.TitleEN, &result.TitleEN},
		{original.TitleKO, current.TitleKO, &result.TitleKO},
		{original.AbstractEN, current.AbstractEN, &result.AbstractEN},
		{original.AbstractKO, current.AbstractKO, &result.AbstractKO},
		{original.Date, current.Date, &result.Date},
		{original.JournalID, current.JournalID, &result.JournalID},
		{original.ConferenceID, current.ConferenceID, &result.ConferenceID},
		{original.Volume, current.Volume, &result.Volume},
		{original.Issue, current.Issue, &result.Issue},
		{original.Pages, current.Pages, &result.Pages},
		{original.ELocator, current.ELocator, &result.ELocator},
		{original.OnlineDate, current.OnlineDate, &result.OnlineDate},
		{original.IssueDate, current.IssueDate, &result.IssueDate},
		{original.PublicationType, current.PublicationType, &result.PublicationType},
		{original.Topic, current.Topic, &result.Topic},
		{original.AwardID, current.AwardID, &result.AwardID},
		{original.DOI, current.DOI, &result.DOI},
		{original.URL, current.URL, &result.URL},
		{original.Note, current.Note, &result.Note},
	}
	for _, field := range stringFields {
		if field.original == field.current {
			*field.target = field.current
		}
	}
	if equalPublicationStrings(original.KeywordsEN, current.KeywordsEN) {
		result.KeywordsEN = clonePublicationStrings(current.KeywordsEN)
	}
	if equalPublicationStrings(original.KeywordsKO, current.KeywordsKO) {
		result.KeywordsKO = clonePublicationStrings(current.KeywordsKO)
	}
	if equalPublicationStrings(original.AuthorIDs, current.AuthorIDs) {
		result.AuthorIDs = clonePublicationStrings(current.AuthorIDs)
	}
}

func validatePublication(
	item publicationDocumentItem,
	personIDs, awardIDs, topicIDs map[string]bool,
	contextLabel string,
) error {
	textFields := []struct {
		name  string
		value string
	}{
		{"title_en", item.TitleEN}, {"title_ko", item.TitleKO},
		{"abstract_en", item.AbstractEN}, {"abstract_ko", item.AbstractKO},
		{"date", item.Date},
		{"publication_type", item.PublicationType}, {"topic", item.Topic},
		{"award_id", item.AwardID}, {"doi", item.DOI}, {"url", item.URL},
		{"note", item.Note},
	}
	for _, field := range textFields {
		if !utf8.ValidString(field.value) {
			return fmt.Errorf("%s의 %s 문자열이 유효한 UTF-8이 아닙니다", contextLabel, field.name)
		}
	}
	for field, value := range publicationBibliographyValues(item) {
		if !utf8.ValidString(value) || utf8.RuneCountInString(value) > 500 || strings.IndexFunc(value, unicode.IsControl) >= 0 {
			return fmt.Errorf("%s의 %s는 제어문자 없이 500자 이하의 문자열로 입력해 주세요", contextLabel, field)
		}
		if value != "" && (field == "online_date" || field == "issue_date") {
			if err := validatePublicationDate(value); err != nil {
				return fmt.Errorf("%s의 %s: %w", contextLabel, field, err)
			}
		}
	}
	if strings.TrimSpace(item.TitleEN) == "" && strings.TrimSpace(item.TitleKO) == "" {
		return fmt.Errorf("%s에는 영문 또는 국문 제목이 필요합니다", contextLabel)
	}
	if err := validatePublicationKeywords(item.KeywordsEN, "keywords_en", contextLabel); err != nil {
		return err
	}
	if err := validatePublicationKeywords(item.KeywordsKO, "keywords_ko", contextLabel); err != nil {
		return err
	}
	if len(item.AuthorIDs) == 0 {
		return fmt.Errorf("%s의 author_ids에는 저자가 한 명 이상 필요합니다", contextLabel)
	}
	usedAuthors := make(map[string]bool, len(item.AuthorIDs))
	for index, id := range item.AuthorIDs {
		if !taxonomyIDPattern.MatchString(id) {
			return fmt.Errorf("%s의 author_ids %d번째 값은 canonical ID여야 합니다", contextLabel, index+1)
		}
		if !personIDs[id] {
			return fmt.Errorf("%s이 존재하지 않는 사람 '%s'을 참조합니다", contextLabel, id)
		}
		if usedAuthors[id] {
			return fmt.Errorf("%s의 저자 '%s'이 중복되었습니다", contextLabel, id)
		}
		usedAuthors[id] = true
	}
	if !publicationTypes[item.PublicationType] {
		return fmt.Errorf("%s의 publication_type이 지원되지 않습니다: %s", contextLabel, item.PublicationType)
	}
	for _, field := range []struct{ name, value, suffix, kind string }{
		{"journal_id", item.JournalID, "-journal", "journals"}, {"conference_id", item.ConferenceID, "-conference", "conferences"},
	} {
		if field.value == "" {
			continue
		}
		if !validPublicationCatalogID(field.kind, field.value) {
			return fmt.Errorf("%s의 %s는 해당 목록의 ID 형식이거나 빈 문자열이어야 합니다", contextLabel, field.name)
		}
		if !strings.HasSuffix(item.PublicationType, field.suffix) {
			return fmt.Errorf("%s의 %s는 논문 유형에 맞는 목록에서 선택해 주세요", contextLabel, field.name)
		}
	}
	if item.UnderReview && item.InPress {
		return fmt.Errorf("%s은 under_review와 in_press를 동시에 사용할 수 없습니다", contextLabel)
	}
	date := strings.TrimSpace(item.Date)
	if date != "" {
		if err := validatePublicationDate(date); err != nil {
			return fmt.Errorf("%s의 date: %w", contextLabel, err)
		}
	}
	if err := validateOptionalPublicationReference(item.Topic, topicIDs, "topic", contextLabel); err != nil {
		return err
	}
	if err := validateOptionalPublicationReference(item.AwardID, awardIDs, "award_id", contextLabel); err != nil {
		return err
	}
	return nil
}

func validatePublicationKeywords(values []string, field, contextLabel string) error {
	if values == nil {
		return fmt.Errorf("%s의 %s는 JSON 배열이어야 합니다", contextLabel, field)
	}
	for index, value := range values {
		if !utf8.ValidString(value) {
			return fmt.Errorf("%s의 %s %d번째 문자열이 유효한 UTF-8이 아닙니다", contextLabel, field, index+1)
		}
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s의 %s %d번째 값은 비어 있지 않은 문자열이어야 합니다", contextLabel, field, index+1)
		}
	}
	return nil
}

func validateOptionalPublicationReference(value string, knownIDs map[string]bool, field, contextLabel string) error {
	if value == "" {
		return nil
	}
	if !taxonomyIDPattern.MatchString(value) {
		return fmt.Errorf("%s의 %s는 canonical ID이거나 빈 문자열이어야 합니다", contextLabel, field)
	}
	if !knownIDs[value] {
		return fmt.Errorf("%s이 존재하지 않는 %s '%s'을 참조합니다", contextLabel, field, value)
	}
	return nil
}

func validatePublicationDate(value string) error {
	var layout string
	switch len(value) {
	case 4:
		layout = "2006"
	case 7:
		layout = "2006-01"
	case 10:
		layout = "2006-01-02"
	default:
		return errors.New("YYYY, YYYY-MM 또는 YYYY-MM-DD 형식이어야 합니다")
	}
	parsed, err := time.Parse(layout, value)
	if err != nil || parsed.Format(layout) != value {
		return errors.New("실제 존재하는 YYYY, YYYY-MM 또는 YYYY-MM-DD 날짜여야 합니다")
	}
	return nil
}

func publicationInputExactlyMatches(input PublicationInput, current publicationDocumentItem) bool {
	if input.KeywordsEN == nil || input.KeywordsKO == nil || input.AuthorIDs == nil {
		return false
	}
	return input.TitleEN == current.TitleEN &&
		input.TitleKO == current.TitleKO &&
		input.AbstractEN == current.AbstractEN &&
		input.AbstractKO == current.AbstractKO &&
		equalPublicationStrings(input.KeywordsEN, current.KeywordsEN) &&
		equalPublicationStrings(input.KeywordsKO, current.KeywordsKO) &&
		equalPublicationStrings(input.AuthorIDs, current.AuthorIDs) &&
		input.Date == current.Date &&
		input.JournalID == current.JournalID &&
		input.ConferenceID == current.ConferenceID &&
		input.Volume == current.Volume &&
		input.Issue == current.Issue &&
		input.Pages == current.Pages &&
		input.ELocator == current.ELocator &&
		input.OnlineDate == current.OnlineDate &&
		input.IssueDate == current.IssueDate &&
		input.UnderReview == current.UnderReview &&
		input.InPress == current.InPress &&
		input.PublicationType == current.PublicationType &&
		input.Topic == current.Topic &&
		input.AwardID == current.AwardID &&
		input.DOI == current.DOI &&
		input.URL == current.URL &&
		input.Note == current.Note
}

func publicationDocumentEqual(left, right publicationDocumentItem) bool {
	return left.TitleEN == right.TitleEN &&
		left.TitleKO == right.TitleKO &&
		left.AbstractEN == right.AbstractEN &&
		left.AbstractKO == right.AbstractKO &&
		equalPublicationStrings(left.KeywordsEN, right.KeywordsEN) &&
		equalPublicationStrings(left.KeywordsKO, right.KeywordsKO) &&
		equalPublicationStrings(left.AuthorIDs, right.AuthorIDs) &&
		left.Date == right.Date &&
		left.JournalID == right.JournalID &&
		left.ConferenceID == right.ConferenceID &&
		left.Volume == right.Volume &&
		left.Issue == right.Issue &&
		left.Pages == right.Pages &&
		left.ELocator == right.ELocator &&
		left.OnlineDate == right.OnlineDate &&
		left.IssueDate == right.IssueDate &&
		left.UnderReview == right.UnderReview &&
		left.InPress == right.InPress &&
		left.PublicationType == right.PublicationType &&
		left.Topic == right.Topic &&
		left.AwardID == right.AwardID &&
		left.DOI == right.DOI &&
		left.URL == right.URL &&
		left.Note == right.Note
}

func equalPublicationStrings(left, right []string) bool {
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

func marshalPublicationItem(current *publicationRow, item publicationDocumentItem) (json.RawMessage, error) {
	fields := map[string]json.RawMessage{}
	if current != nil {
		var err error
		fields, err = decodeObject(current.Raw)
		if err != nil {
			return nil, err
		}
	}
	merged, err := mergeObjectFields(fields, map[string]any{
		"title_en":         item.TitleEN,
		"title_ko":         item.TitleKO,
		"abstract_en":      item.AbstractEN,
		"abstract_ko":      item.AbstractKO,
		"keywords_en":      item.KeywordsEN,
		"keywords_ko":      item.KeywordsKO,
		"author_ids":       item.AuthorIDs,
		"date":             item.Date,
		"under_review":     item.UnderReview,
		"in_press":         item.InPress,
		"publication_type": item.PublicationType,
		"topic":            item.Topic,
		"award_id":         item.AwardID,
		"doi":              item.DOI,
		"url":              item.URL,
		"note":             item.Note,
	})
	if err != nil {
		return nil, err
	}
	optional := publicationBibliographyValues(item)
	optional["journal_id"] = item.JournalID
	optional["conference_id"] = item.ConferenceID
	for name, value := range optional {
		if _, exists := fields[name]; exists || value != "" {
			merged, err = mergeObjectFields(merged, map[string]any{name: value})
			if err != nil {
				return nil, err
			}
		}
	}
	return json.Marshal(merged)
}

func publicationBibliographyValues(item publicationDocumentItem) map[string]string {
	return map[string]string{
		"volume": item.Volume, "issue": item.Issue, "pages": item.Pages,
		"elocator": item.ELocator, "online_date": item.OnlineDate, "issue_date": item.IssueDate,
	}
}

func publicationSortLess(left, right publicationDocumentItem) bool {
	leftRank := publicationStatusRank(left)
	rightRank := publicationStatusRank(right)
	if leftRank != rightRank {
		return leftRank > rightRank
	}
	leftDate := publicationDateParts(left.Date)
	rightDate := publicationDateParts(right.Date)
	for index := range leftDate {
		if leftDate[index] != rightDate[index] {
			return leftDate[index] > rightDate[index]
		}
	}
	leftTitle := strings.ToLower(publicationLocalizedTitle(left))
	rightTitle := strings.ToLower(publicationLocalizedTitle(right))
	return leftTitle < rightTitle
}

func publicationStatusRank(item publicationDocumentItem) int {
	if item.UnderReview {
		return 2
	}
	if item.InPress {
		return 1
	}
	return 0
}

func publicationDateParts(value string) [3]int {
	var result [3]int
	for index, part := range strings.Split(strings.TrimSpace(value), "-") {
		if index >= len(result) || part == "" {
			break
		}
		result[index], _ = strconv.Atoi(part)
	}
	return result
}

func publicationLocalizedTitle(item publicationDocumentItem) string {
	if item.TitleEN != "" {
		return item.TitleEN
	}
	return item.TitleKO
}

// date is a derived cache for older consumers; catalog/bibliography fields are
// authoritative, including their original year/month/day precision.
func effectivePublicationDate(item publicationDocumentItem, settings SettingsDocument) string {
	if strings.HasSuffix(item.PublicationType, "-conference") {
		for _, conference := range settings.Conferences {
			if conference.ID != item.ConferenceID {
				continue
			}
			if date := strings.TrimSpace(conference.ProceedingsDate); date != "" {
				return date
			}
			return strings.TrimSpace(conference.EndDate)
		}
		return ""
	}
	if date := strings.TrimSpace(item.IssueDate); date != "" {
		return date
	}
	return strings.TrimSpace(item.OnlineDate)
}
