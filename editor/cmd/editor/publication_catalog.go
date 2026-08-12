package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// PublicationVenue is shared catalog metadata referenced by publications.
type PublicationVenue struct {
	ID                      string   `json:"id"`
	NameEN                  string   `json:"name_en"`
	NameKO                  string   `json:"name_ko"`
	PublisherEN             string   `json:"publisher_en,omitempty"`
	PublisherKO             string   `json:"publisher_ko,omitempty"`
	ISSN                    string   `json:"issn,omitempty"`
	URL                     string   `json:"url,omitempty"`
	Indexing                []string `json:"indexing"`
	StartDate               string   `json:"start_date,omitempty"`
	EndDate                 string   `json:"end_date,omitempty"`
	LocationEN              string   `json:"location_en,omitempty"`
	LocationKO              string   `json:"location_ko,omitempty"`
	ProceedingsPublisherEN  string   `json:"proceedings_publisher_en,omitempty"`
	ProceedingsPublisherKO  string   `json:"proceedings_publisher_ko,omitempty"`
	ProceedingsTitleEN      string   `json:"proceedings_title_en,omitempty"`
	ProceedingsTitleKO      string   `json:"proceedings_title_ko,omitempty"`
	ProceedingsVolume       string   `json:"proceedings_volume,omitempty"`
	ProceedingsIssue        string   `json:"proceedings_issue,omitempty"`
	ProceedingsSeriesEN     string   `json:"proceedings_series_en,omitempty"`
	ProceedingsSeriesKO     string   `json:"proceedings_series_ko,omitempty"`
	ProceedingsSeriesVolume string   `json:"proceedings_series_volume,omitempty"`
	ProceedingsDate         string   `json:"proceedings_date,omitempty"`
	ProceedingsISSN         string   `json:"proceedings_issn,omitempty"`
	extraFields             map[string]json.RawMessage
}

func (item *PublicationVenue) UnmarshalJSON(raw []byte) error {
	fields, err := decodeObject(raw)
	if err != nil {
		return err
	}
	for _, field := range []string{"proceedings_volume", "proceedings_issue", "proceedings_series_volume"} {
		if value, exists := fields[field]; exists {
			if _, err := decodeRequiredString(value, field); err != nil {
				return err
			}
		}
	}
	// Preserve older single-language entries, without reviving a deliberately cleared pair.
	for _, base := range []string{"name", "publisher", "location", "proceedings_publisher", "proceedings_title", "proceedings_series"} {
		legacy, exists := fields[base]
		_, hasEN := fields[base+"_en"]
		_, hasKO := fields[base+"_ko"]
		if exists && !hasEN && !hasKO {
			var value string
			if err := json.Unmarshal(legacy, &value); err != nil {
				return fmt.Errorf("%s는 문자열이어야 합니다", base)
			}
			language := "_en"
			if strings.IndexFunc(value, func(r rune) bool { return unicode.Is(unicode.Hangul, r) }) >= 0 {
				language = "_ko"
			}
			fields[base+language] = legacy
		}
		delete(fields, base)
	}
	raw, err = json.Marshal(fields)
	if err != nil {
		return err
	}
	type venueJSON PublicationVenue
	var decoded venueJSON
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return err
	}
	*item = PublicationVenue(decoded)
	item.extraFields = withoutKnownFields(fields, "id", "name_en", "name_ko", "publisher_en", "publisher_ko", "issn", "url", "indexing",
		"start_date", "end_date", "location_en", "location_ko", "proceedings_publisher_en", "proceedings_publisher_ko",
		"proceedings_title_en", "proceedings_title_ko", "proceedings_volume", "proceedings_issue", "proceedings_series_en", "proceedings_series_ko", "proceedings_series_volume", "proceedings_date", "proceedings_issn")
	return nil
}

func (item PublicationVenue) MarshalJSON() ([]byte, error) {
	type venueJSON PublicationVenue
	raw, err := json.Marshal(venueJSON(item))
	if err != nil {
		return nil, err
	}
	known, err := decodeObject(raw)
	if err != nil {
		return nil, err
	}
	fields := cloneExtraFields(item.extraFields)
	if fields == nil {
		fields = make(map[string]json.RawMessage)
	}
	for key, value := range known {
		fields[key] = value
	}
	return json.Marshal(fields)
}

var catalogISSNPattern = regexp.MustCompile(`^[0-9]{4}-[0-9]{3}[0-9X]$`)
var conferenceIDPattern = regexp.MustCompile(`^[A-Za-z0-9]+(?:[ -][A-Za-z0-9]+\.?)*$`)

func validPublicationCatalogID(kind, value string) bool {
	if kind == "conferences" {
		return conferenceIDPattern.MatchString(value)
	}
	return taxonomyIDPattern.MatchString(value)
}

func validatePublicationCatalogReferences(settings SettingsDocument, usage SettingsUsage) error {
	for _, catalog := range []struct {
		label string
		items []PublicationVenue
		usage map[string]int
	}{{"저널", settings.Journals, usage.Journals}, {"학술대회", settings.Conferences, usage.Conferences}} {
		ids := make(map[string]bool, len(catalog.items))
		for _, item := range catalog.items {
			ids[item.ID] = true
		}
		for id, count := range catalog.usage {
			if id != "" && count > 0 && !ids[id] {
				return fmt.Errorf("publications.json의 %d개 항목이 존재하지 않는 %s '%s'을(를) 참조합니다", count, catalog.label, id)
			}
		}
	}
	return nil
}

func normalisePublicationCatalog(input, current []PublicationVenue, kind string) ([]PublicationVenue, error) {
	label := "저널"
	if kind == "conferences" {
		label = "학회"
	}
	// Omitted arrays from older clients preserve catalogs; an explicit [] clears them.
	if input == nil {
		input = current
	}
	if len(input) > 1000 {
		return nil, fmt.Errorf("%s은(는) 1000개 이하로 관리해 주세요", label)
	}
	result := make([]PublicationVenue, 0, len(input))
	seenIDs := make(map[string]bool)
	for index, item := range input {
		context := fmt.Sprintf("%s %d", label, index+1)
		item.extraFields = cloneExtraFields(item.extraFields)
		for _, field := range []struct {
			name  string
			value *string
		}{
			{"ID", &item.ID}, {"영문 이름", &item.NameEN}, {"국문 이름", &item.NameKO},
			{"영문 출판사", &item.PublisherEN}, {"국문 출판사", &item.PublisherKO}, {"ISSN", &item.ISSN},
			{"URL", &item.URL}, {"시작일", &item.StartDate}, {"종료일", &item.EndDate},
			{"영문 장소", &item.LocationEN}, {"국문 장소", &item.LocationKO},
			{"영문 proceedings 발행기관", &item.ProceedingsPublisherEN}, {"국문 proceedings 발행기관", &item.ProceedingsPublisherKO},
			{"영문 proceedings 이름", &item.ProceedingsTitleEN}, {"국문 proceedings 이름", &item.ProceedingsTitleKO},
			{"proceedings 권", &item.ProceedingsVolume}, {"proceedings 호", &item.ProceedingsIssue},
			{"영문 시리즈명", &item.ProceedingsSeriesEN}, {"국문 시리즈명", &item.ProceedingsSeriesKO},
			{"시리즈 권", &item.ProceedingsSeriesVolume},
			{"proceedings 발행일", &item.ProceedingsDate}, {"proceedings ISSN", &item.ProceedingsISSN},
		} {
			*field.value = strings.TrimSpace(*field.value)
			maximum := 500
			if field.name == "URL" {
				maximum = 4096
			}
			if utf8.RuneCountInString(*field.value) > maximum || strings.IndexFunc(*field.value, unicode.IsControl) >= 0 {
				return nil, fmt.Errorf("%s의 %s은(는) 제어문자 없이 %d자 이하로 입력해 주세요", context, field.name, maximum)
			}
		}
		if !validPublicationCatalogID(kind, item.ID) {
			if kind == "conferences" {
				return nil, fmt.Errorf("%s ID는 AIK Conf. 2024F처럼 영문·숫자·공백·하이픈·약어 마침표로 입력해 주세요", context)
			}
			return nil, fmt.Errorf("%s ID는 소문자 영문, 숫자, 단일 하이픈으로 입력해 주세요", context)
		}
		if seenIDs[item.ID] {
			return nil, fmt.Errorf("%s ID가 중복되었습니다: %s", label, item.ID)
		}
		seenIDs[item.ID] = true
		if item.NameEN == "" && item.NameKO == "" {
			return nil, fmt.Errorf("%s의 영문 또는 국문 이름을 입력해 주세요", context)
		}
		if item.URL != "" {
			if err := validateSoftwareURL(item.URL, context+" URL"); err != nil {
				return nil, err
			}
		}
		item.ISSN = strings.ToUpper(item.ISSN)
		item.ProceedingsISSN = strings.ToUpper(item.ProceedingsISSN)
		for _, issn := range []string{item.ISSN, item.ProceedingsISSN} {
			if issn != "" && !catalogISSNPattern.MatchString(issn) {
				return nil, fmt.Errorf("%s ISSN은 1234-567X 형식으로 입력해 주세요", context)
			}
		}
		for _, date := range []string{item.StartDate, item.EndDate} {
			if date != "" {
				if _, err := parseCanonicalDate(date); err != nil {
					return nil, fmt.Errorf("%s 날짜는 YYYY-MM-DD 형식의 실제 날짜여야 합니다", context)
				}
			}
		}
		if item.ProceedingsDate != "" {
			if err := validatePublicationDate(item.ProceedingsDate); err != nil {
				return nil, fmt.Errorf("%s Proceedings 발행일: %w", context, err)
			}
		}
		if item.EndDate != "" && (item.StartDate == "" || item.StartDate > item.EndDate) {
			return nil, fmt.Errorf("%s의 시작일을 입력하고 종료일보다 늦지 않게 지정해 주세요", context)
		}
		if len(item.Indexing) > 1 {
			return nil, fmt.Errorf("%s 등재 구분은 SCIE, SCOPUS, KCI 등재지 중 하나만 선택해 주세요", context)
		}
		indexing := make([]string, 0, len(item.Indexing))
		for _, value := range item.Indexing {
			value = strings.ToUpper(strings.TrimSpace(value))
			if value != "SCIE" && value != "SCOPUS" && value != "KCI" {
				return nil, fmt.Errorf("%s 등재 구분은 SCIE, SCOPUS, KCI 중 선택해 주세요", context)
			}
			indexing = append(indexing, value)
		}
		item.Indexing = indexing
		result = append(result, item)
	}
	return result, nil
}
