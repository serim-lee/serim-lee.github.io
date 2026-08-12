package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Gonie-Gonie/Gonie-Gonie.github.io/editor/internal/dataversion"
	"os"
	"strings"
)

const (
	maxAcademicActivityItems = 5_000
	maxAcademicReviewDates   = 1_000
)

var academicActivityCategories = []string{
	"editorial_service",
	"professional_service",
	"conference_service",
	"invited_talks",
	"reviews",
}

var academicActivityCategorySet = map[string]bool{
	"reviews":              true,
	"invited_talks":        true,
	"conference_service":   true,
	"professional_service": true,
	"editorial_service":    true,
}

// AcademicActivityItemSummary is deliberately flat for the Wails frontend.
// EditorKey is a transient snapshot handle; academic activity records do not
// persist IDs because no other data source relates to individual activities.
type AcademicActivityItemSummary struct {
	EditorKey      string   `json:"editor_key"`
	VisibleInCV    bool     `json:"visible_in_CV"`
	Category       string   `json:"category"`
	JournalEN      string   `json:"journal_en"`
	JournalKO      string   `json:"journal_ko"`
	CompletedDates []string `json:"completed_dates"`
	EventEN        string   `json:"event_en"`
	EventKO        string   `json:"event_ko"`
	TopicEN        string   `json:"topic_en"`
	TopicKO        string   `json:"topic_ko"`
	ConferenceEN   string   `json:"conference_en"`
	ConferenceKO   string   `json:"conference_ko"`
	OrganizationEN string   `json:"organization_en"`
	OrganizationKO string   `json:"organization_ko"`
	RoleEN         string   `json:"role_en"`
	RoleKO         string   `json:"role_ko"`
	Date           string   `json:"date"`
	StartDate      string   `json:"start_date"`
	EndDate        string   `json:"end_date"`
}

type AcademicActivityInput struct {
	Category       string   `json:"category"`
	JournalEN      string   `json:"journal_en"`
	JournalKO      string   `json:"journal_ko"`
	CompletedDates []string `json:"completed_dates"`
	EventEN        string   `json:"event_en"`
	EventKO        string   `json:"event_ko"`
	TopicEN        string   `json:"topic_en"`
	TopicKO        string   `json:"topic_ko"`
	ConferenceEN   string   `json:"conference_en"`
	ConferenceKO   string   `json:"conference_ko"`
	OrganizationEN string   `json:"organization_en"`
	OrganizationKO string   `json:"organization_ko"`
	RoleEN         string   `json:"role_en"`
	RoleKO         string   `json:"role_ko"`
	Date           string   `json:"date"`
	StartDate      string   `json:"start_date"`
	EndDate        string   `json:"end_date"`
}

type AcademicActivitySaveItem struct {
	EditorKey   string                 `json:"editor_key,omitempty"`
	Activity    *AcademicActivityInput `json:"activity,omitempty"`
	VisibleInCV *bool                  `json:"visible_in_CV,omitempty"`
}

type academicActivityDocumentItem struct {
	AcademicActivityInput
}

type academicActivityRow struct {
	Key      string
	Raw      json.RawMessage
	Category string
	Item     academicActivityDocumentItem
}

type academicActivitiesSnapshot struct {
	Raw        []byte
	Rows       []academicActivityRow
	RootFields map[string]json.RawMessage
}

func readAcademicActivities(path string) (academicActivitiesSnapshot, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return academicActivitiesSnapshot{}, fmt.Errorf("academic_activities.json 읽기 실패: %w", err)
	}
	return readAcademicActivitiesBytes(raw)
}

func readAcademicActivitiesBytes(raw []byte) (academicActivitiesSnapshot, error) {
	root, _, err := dataversion.Object(raw, "academic_activities.json")
	if err != nil {
		return academicActivitiesSnapshot{}, err
	}

	snapshot := academicActivitiesSnapshot{
		Raw:        append([]byte(nil), raw...),
		Rows:       make([]academicActivityRow, 0),
		RootFields: make(map[string]json.RawMessage, len(root)),
	}
	for name, value := range root {
		snapshot.RootFields[name] = cloneRawMessage(value)
	}
	for _, category := range academicActivityCategories {
		categoryRaw, exists := root[category]
		if !exists {
			return academicActivitiesSnapshot{}, fmt.Errorf("academic_activities.json에 %s 배열이 없습니다", category)
		}
		var rows []json.RawMessage
		if err := json.Unmarshal(categoryRaw, &rows); err != nil || rows == nil {
			return academicActivitiesSnapshot{}, fmt.Errorf("academic_activities.json의 %s는 JSON 배열이어야 합니다", category)
		}
		if len(snapshot.Rows)+len(rows) > maxAcademicActivityItems {
			return academicActivitiesSnapshot{}, fmt.Errorf("학술활동은 최대 %d개까지 관리할 수 있습니다", maxAcademicActivityItems)
		}
		for index, rowRaw := range rows {
			contextLabel := fmt.Sprintf("academic_activities.json %s %d번째 항목", category, index+1)
			item, err := parseAcademicActivityItem(rowRaw, category, contextLabel)
			if err != nil {
				return academicActivitiesSnapshot{}, err
			}
			hash := revisionOf(rowRaw)
			snapshot.Rows = append(snapshot.Rows, academicActivityRow{
				Key:      fmt.Sprintf("%s:%d:%s", category, index, hash[:20]),
				Raw:      cloneRawMessage(rowRaw),
				Category: category,
				Item:     item,
			})
		}
	}
	return snapshot, nil
}

func parseAcademicActivityItem(raw json.RawMessage, category, contextLabel string) (academicActivityDocumentItem, error) {
	fields, err := decodeObject(raw)
	if err != nil {
		return academicActivityDocumentItem{}, fmt.Errorf("%s은 JSON 객체여야 합니다", contextLabel)
	}
	if _, err := visibleInCV(raw, contextLabel); err != nil {
		return academicActivityDocumentItem{}, err
	}
	if _, exists := fields["id"]; exists {
		return academicActivityDocumentItem{}, fmt.Errorf("%s은 id 필드를 사용하지 않습니다", contextLabel)
	}
	if category == "reviews" {
		if _, exists := fields["review_count"]; exists {
			return academicActivityDocumentItem{}, fmt.Errorf("%s은 review_count 대신 completed_dates를 사용해야 합니다", contextLabel)
		}
	}
	input := AcademicActivityInput{Category: category}
	for _, field := range academicActivityStringFields(category) {
		value, fieldErr := academicRequiredString(fields, field, contextLabel)
		if fieldErr != nil {
			return academicActivityDocumentItem{}, fieldErr
		}
		setAcademicActivityString(&input, field, value)
	}
	if category == "reviews" {
		value, exists := fields["completed_dates"]
		if !exists {
			return academicActivityDocumentItem{}, fmt.Errorf("%s에 completed_dates 필드가 없습니다", contextLabel)
		}
		input.CompletedDates, err = decodeRequiredStringArray(value, contextLabel+" completed_dates")
		if err != nil {
			return academicActivityDocumentItem{}, err
		}
	}
	if err := validateAcademicActivity(input, contextLabel, false); err != nil {
		return academicActivityDocumentItem{}, err
	}
	return academicActivityDocumentItem{AcademicActivityInput: input}, nil
}

func academicRequiredString(fields map[string]json.RawMessage, field, contextLabel string) (string, error) {
	raw, exists := fields[field]
	if !exists {
		return "", fmt.Errorf("%s에 %s 필드가 없습니다", contextLabel, field)
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", fmt.Errorf("%s의 %s는 문자열이어야 합니다", contextLabel, field)
	}
	return value, nil
}

func academicActivityStringFields(category string) []string {
	switch category {
	case "reviews":
		return []string{"journal_en", "journal_ko"}
	case "invited_talks":
		return []string{"event_en", "event_ko", "topic_en", "topic_ko", "date"}
	case "conference_service":
		return []string{"conference_en", "conference_ko", "role_en", "role_ko", "start_date", "end_date"}
	case "professional_service":
		return []string{"organization_en", "organization_ko", "role_en", "role_ko", "start_date", "end_date"}
	case "editorial_service":
		return []string{"journal_en", "journal_ko", "role_en", "role_ko", "start_date", "end_date"}
	default:
		return nil
	}
}

func setAcademicActivityString(input *AcademicActivityInput, field, value string) {
	switch field {
	case "journal_en":
		input.JournalEN = value
	case "journal_ko":
		input.JournalKO = value
	case "event_en":
		input.EventEN = value
	case "event_ko":
		input.EventKO = value
	case "topic_en":
		input.TopicEN = value
	case "topic_ko":
		input.TopicKO = value
	case "conference_en":
		input.ConferenceEN = value
	case "conference_ko":
		input.ConferenceKO = value
	case "organization_en":
		input.OrganizationEN = value
	case "organization_ko":
		input.OrganizationKO = value
	case "role_en":
		input.RoleEN = value
	case "role_ko":
		input.RoleKO = value
	case "date":
		input.Date = value
	case "start_date":
		input.StartDate = value
	case "end_date":
		input.EndDate = value
	}
}

func academicActivityItemSummaries(snapshot academicActivitiesSnapshot) []AcademicActivityItemSummary {
	items := make([]AcademicActivityItemSummary, 0, len(snapshot.Rows))
	for _, row := range snapshot.Rows {
		item := row.Item.AcademicActivityInput
		items = append(items, AcademicActivityItemSummary{
			EditorKey: row.Key, VisibleInCV: mustVisibleInCV(row.Raw), Category: row.Category,
			JournalEN: item.JournalEN, JournalKO: item.JournalKO,
			CompletedDates: append([]string{}, item.CompletedDates...),
			EventEN:        item.EventEN, EventKO: item.EventKO, TopicEN: item.TopicEN, TopicKO: item.TopicKO,
			ConferenceEN: item.ConferenceEN, ConferenceKO: item.ConferenceKO,
			OrganizationEN: item.OrganizationEN, OrganizationKO: item.OrganizationKO,
			RoleEN: item.RoleEN, RoleKO: item.RoleKO, Date: item.Date,
			StartDate: item.StartDate, EndDate: item.EndDate,
		})
	}
	return items
}

func academicActivityCount(snapshot academicActivitiesSnapshot) int { return len(snapshot.Rows) }

func academicActivityCVCount(snapshot academicActivitiesSnapshot) int {
	count := 0
	for _, row := range snapshot.Rows {
		if mustVisibleInCV(row.Raw) {
			count++
		}
	}
	return count
}

func buildAcademicActivitiesSave(current academicActivitiesSnapshot, request []AcademicActivitySaveItem) ([]byte, error) {
	if len(request) > maxAcademicActivityItems {
		return nil, fmt.Errorf("학술활동은 최대 %d개까지 저장할 수 있습니다", maxAcademicActivityItems)
	}
	currentByKey := make(map[string]academicActivityRow, len(current.Rows))
	for _, row := range current.Rows {
		currentByKey[row.Key] = row
	}
	rows := make([]academicActivityRow, 0, len(request))
	usedExisting := make(map[string]bool)
	documentUnchanged := len(request) == len(current.Rows)
	appendRow := func(row academicActivityRow, requested *bool, isNew bool, contextLabel string) error {
		updated, err := withVisibleInCV(row.Raw, visibleInCVForSave(requested, isNew))
		if err != nil {
			return fmt.Errorf("%s visible_in_CV 인코딩 실패: %w", contextLabel, err)
		}
		if !bytes.Equal(updated, row.Raw) {
			documentUnchanged = false
		}
		row.Raw = updated
		rows = append(rows, row)
		return nil
	}
	for index, requested := range request {
		contextLabel := fmt.Sprintf("학술활동 %d", index+1)
		var currentRow *academicActivityRow
		if requested.EditorKey != "" {
			row, exists := currentByKey[requested.EditorKey]
			if !exists {
				return nil, fmt.Errorf("%s이(가) 현재 academic_activities.json에 없습니다. 다시 불러와 주세요", contextLabel)
			}
			if usedExisting[requested.EditorKey] {
				return nil, fmt.Errorf("%s이(가) 중복되었습니다", contextLabel)
			}
			usedExisting[requested.EditorKey] = true
			currentRow = &row
			if requested.Activity == nil {
				if err := appendRow(row, requested.VisibleInCV, false, contextLabel); err != nil {
					return nil, err
				}
				if documentUnchanged && requested.EditorKey != current.Rows[index].Key {
					documentUnchanged = false
				}
				continue
			}
		} else if requested.Activity == nil {
			return nil, fmt.Errorf("%s의 activity 입력이 없습니다", contextLabel)
		}

		input := *requested.Activity
		if currentRow != nil && academicActivityInputExactlyMatches(input, currentRow.Item.AcademicActivityInput) {
			if err := appendRow(*currentRow, requested.VisibleInCV, false, contextLabel); err != nil {
				return nil, err
			}
			if documentUnchanged && requested.EditorKey != current.Rows[index].Key {
				documentUnchanged = false
			}
			continue
		}
		normalized, err := normaliseAcademicActivityInput(input, contextLabel)
		if err != nil {
			return nil, err
		}
		document := academicActivityDocumentItem{AcademicActivityInput: normalized}
		rowRaw, err := marshalAcademicActivityItem(currentRow, document)
		if err != nil {
			return nil, fmt.Errorf("%s 인코딩 실패: %w", contextLabel, err)
		}
		row := academicActivityRow{Raw: rowRaw, Category: normalized.Category, Item: document}
		if err := appendRow(row, requested.VisibleInCV, currentRow == nil, contextLabel); err != nil {
			return nil, err
		}
		documentUnchanged = false
	}
	if documentUnchanged {
		return append([]byte(nil), current.Raw...), nil
	}
	return marshalAcademicActivities(current, rows)
}

func normaliseAcademicActivityInput(input AcademicActivityInput, contextLabel string) (AcademicActivityInput, error) {
	result := input
	result.Category = strings.TrimSpace(input.Category)
	if input.CompletedDates != nil {
		result.CompletedDates = make([]string, len(input.CompletedDates))
		for index, completedDate := range input.CompletedDates {
			result.CompletedDates[index] = strings.TrimSpace(completedDate)
		}
	}
	for _, field := range []string{"journal_en", "journal_ko", "event_en", "event_ko", "topic_en", "topic_ko", "conference_en", "conference_ko", "organization_en", "organization_ko", "role_en", "role_ko", "date", "start_date", "end_date"} {
		setAcademicActivityString(&result, field, strings.TrimSpace(academicActivityString(input, field)))
	}
	if err := validateAcademicActivity(result, contextLabel, true); err != nil {
		return AcademicActivityInput{}, err
	}
	return result, nil
}

func academicActivityString(input AcademicActivityInput, field string) string {
	switch field {
	case "journal_en":
		return input.JournalEN
	case "journal_ko":
		return input.JournalKO
	case "event_en":
		return input.EventEN
	case "event_ko":
		return input.EventKO
	case "topic_en":
		return input.TopicEN
	case "topic_ko":
		return input.TopicKO
	case "conference_en":
		return input.ConferenceEN
	case "conference_ko":
		return input.ConferenceKO
	case "organization_en":
		return input.OrganizationEN
	case "organization_ko":
		return input.OrganizationKO
	case "role_en":
		return input.RoleEN
	case "role_ko":
		return input.RoleKO
	case "date":
		return input.Date
	case "start_date":
		return input.StartDate
	case "end_date":
		return input.EndDate
	default:
		return ""
	}
}

func validateAcademicActivity(input AcademicActivityInput, contextLabel string, normalized bool) error {
	if !academicActivityCategorySet[input.Category] {
		return fmt.Errorf("%s의 분류가 올바르지 않습니다", contextLabel)
	}
	for _, field := range academicActivityStringFields(input.Category) {
		if err := validateSoftwareText(academicActivityString(input, field), contextLabel+" "+field, 2_000, false, normalized); err != nil {
			return err
		}
	}
	requireBilingual := func(en, ko, label string) error {
		if strings.TrimSpace(en) == "" && strings.TrimSpace(ko) == "" {
			return fmt.Errorf("%s에는 영문 또는 국문 %s이(가) 필요합니다", contextLabel, label)
		}
		return nil
	}
	date := func(value, label string, optional bool) error {
		if optional && value == "" {
			return nil
		}
		if _, err := parseCanonicalDate(value); err != nil {
			return fmt.Errorf("%s의 %s는 YYYY-MM-DD 형식의 실제 날짜여야 합니다", contextLabel, label)
		}
		return nil
	}
	switch input.Category {
	case "reviews":
		if err := requireBilingual(input.JournalEN, input.JournalKO, "저널명"); err != nil {
			return err
		}
		if len(input.CompletedDates) == 0 {
			return fmt.Errorf("%s에는 완료일이 하나 이상 필요합니다", contextLabel)
		}
		if len(input.CompletedDates) > maxAcademicReviewDates {
			return fmt.Errorf("%s의 완료일은 최대 %d개까지 입력할 수 있습니다", contextLabel, maxAcademicReviewDates)
		}
		seenDates := make(map[string]bool, len(input.CompletedDates))
		for index, completedDate := range input.CompletedDates {
			if err := date(completedDate, fmt.Sprintf("완료일 %d", index+1), false); err != nil {
				return err
			}
			if seenDates[completedDate] {
				return fmt.Errorf("%s의 완료일 '%s'이(가) 중복되었습니다", contextLabel, completedDate)
			}
			seenDates[completedDate] = true
		}
	case "invited_talks":
		if err := requireBilingual(input.EventEN, input.EventKO, "행사명"); err != nil {
			return err
		}
		if err := requireBilingual(input.TopicEN, input.TopicKO, "주제"); err != nil {
			return err
		}
		return date(input.Date, "날짜", false)
	case "conference_service":
		if err := requireBilingual(input.ConferenceEN, input.ConferenceKO, "학술대회명"); err != nil {
			return err
		}
		if err := requireBilingual(input.RoleEN, input.RoleKO, "역할"); err != nil {
			return err
		}
		if err := date(input.StartDate, "시작일", false); err != nil {
			return err
		}
		if err := date(input.EndDate, "종료일", true); err != nil {
			return err
		}
	case "professional_service":
		if err := requireBilingual(input.OrganizationEN, input.OrganizationKO, "기관명"); err != nil {
			return err
		}
		if err := requireBilingual(input.RoleEN, input.RoleKO, "역할"); err != nil {
			return err
		}
		if err := date(input.StartDate, "시작일", false); err != nil {
			return err
		}
		if err := date(input.EndDate, "종료일", true); err != nil {
			return err
		}
	case "editorial_service":
		if err := requireBilingual(input.JournalEN, input.JournalKO, "저널명"); err != nil {
			return err
		}
		if err := requireBilingual(input.RoleEN, input.RoleKO, "역할"); err != nil {
			return err
		}
		if err := date(input.StartDate, "시작일", false); err != nil {
			return err
		}
		if err := date(input.EndDate, "종료일", true); err != nil {
			return err
		}
	}
	if input.EndDate != "" && input.StartDate != "" && input.EndDate < input.StartDate {
		return fmt.Errorf("%s의 종료일은 시작일보다 빠를 수 없습니다", contextLabel)
	}
	return nil
}

func academicActivityInputExactlyMatches(left, right AcademicActivityInput) bool {
	return left.Category == right.Category &&
		left.JournalEN == right.JournalEN && left.JournalKO == right.JournalKO && equalSoftwareStrings(left.CompletedDates, right.CompletedDates) &&
		left.EventEN == right.EventEN && left.EventKO == right.EventKO && left.TopicEN == right.TopicEN && left.TopicKO == right.TopicKO &&
		left.ConferenceEN == right.ConferenceEN && left.ConferenceKO == right.ConferenceKO &&
		left.OrganizationEN == right.OrganizationEN && left.OrganizationKO == right.OrganizationKO &&
		left.RoleEN == right.RoleEN && left.RoleKO == right.RoleKO && left.Date == right.Date &&
		left.StartDate == right.StartDate && left.EndDate == right.EndDate
}

func marshalAcademicActivityItem(current *academicActivityRow, item academicActivityDocumentItem) (json.RawMessage, error) {
	fields := map[string]json.RawMessage{}
	if current != nil {
		var err error
		fields, err = decodeObject(current.Raw)
		if err != nil {
			return nil, err
		}
	}
	for _, field := range []string{"id", "journal_en", "journal_ko", "review_count", "completed_dates", "event_en", "event_ko", "topic_en", "topic_ko", "conference_en", "conference_ko", "organization_en", "organization_ko", "role_en", "role_ko", "date", "start_date", "end_date"} {
		delete(fields, field)
	}
	values := map[string]any{}
	for _, field := range academicActivityStringFields(item.Category) {
		values[field] = academicActivityString(item.AcademicActivityInput, field)
	}
	if item.Category == "reviews" {
		values["completed_dates"] = item.CompletedDates
	}
	merged, err := mergeObjectFields(fields, values)
	if err != nil {
		return nil, err
	}
	return json.Marshal(merged)
}

func marshalAcademicActivities(current academicActivitiesSnapshot, rows []academicActivityRow) ([]byte, error) {
	root := make(map[string]json.RawMessage, len(current.RootFields))
	for name, value := range current.RootFields {
		root[name] = cloneRawMessage(value)
	}
	delete(root, "schema_version")
	delete(root, "version")
	for _, category := range academicActivityCategories {
		items := make([]json.RawMessage, 0)
		for _, row := range rows {
			if row.Category == category {
				items = append(items, cloneRawMessage(row.Raw))
			}
		}
		encoded, err := json.Marshal(items)
		if err != nil {
			return nil, err
		}
		root[category] = encoded
	}
	encoded, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("academic_activities.json 인코딩 실패: %w", err)
	}
	return append(encoded, '\n'), nil
}
