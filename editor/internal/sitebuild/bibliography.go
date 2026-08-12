package sitebuild

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

var months = []string{"January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"}

func catalogFor(item, settings record) record {
	kind, field := "journals", "journal_id"
	if strings.HasSuffix(text(item["publication_type"]), "-conference") {
		kind, field = "conferences", "conference_id"
	}
	for _, v := range array(settings[kind]) {
		r := obj(v)
		if r != nil && str(r["id"]) == str(item[field]) {
			return r
		}
	}
	return record{}
}
func bibliographyLocalized(r record, key, lang string) string {
	if v := text(r[key+"_"+lang]); v != "" {
		return v
	}
	other := "ko"
	if lang == "ko" {
		other = "en"
	}
	return text(r[key+"_"+other])
}
func publicationDate(item, settings record) string {
	if strings.HasSuffix(text(item["publication_type"]), "-conference") {
		r := catalogFor(item, settings)
		if v := text(r["proceedings_date"]); v != "" {
			return v
		}
		return text(r["end_date"])
	}
	if v := text(item["issue_date"]); v != "" {
		return v
	}
	return text(item["online_date"])
}
func conferencePeriod(r record, lang string) string {
	start, end := text(r["start_date"]), text(r["end_date"])
	parse := func(s string) (int, int, int) {
		p := strings.Split(s, "-")
		if len(p) != 3 {
			return 0, 0, 0
		}
		y, _ := strconv.Atoi(p[0])
		m, _ := strconv.Atoi(p[1])
		d, _ := strconv.Atoi(p[2])
		return y, m, d
	}
	format := func(s string) string {
		if !fullDate.MatchString(s) {
			return s
		}
		y, m, d := parse(s)
		if m < 1 || m > 12 {
			return s
		}
		if lang == "ko" {
			return fmt.Sprintf("%d년 %d월 %d일", y, m, d)
		}
		return fmt.Sprintf("%s %d, %d", months[m-1], d, y)
	}
	if start == "" || end == "" || start == end {
		if start != "" {
			return format(start)
		}
		return format(end)
	}
	sy, sm, sd := parse(start)
	ey, em, ed := parse(end)
	if sy != ey {
		return format(start) + " – " + format(end)
	}
	if sm < 1 || sm > 12 || em < 1 || em > 12 {
		return format(start) + " – " + format(end)
	}
	if lang == "ko" {
		if sm == em {
			return fmt.Sprintf("%d년 %d월 %d-%d일", sy, sm, sd, ed)
		}
		return fmt.Sprintf("%d년 %d월 %d일-%d월 %d일", sy, sm, sd, em, ed)
	}
	if sm == em {
		return fmt.Sprintf("%s %d-%d, %d", months[sm-1], sd, ed, sy)
	}
	return fmt.Sprintf("%s %d-%s %d, %d", months[sm-1], sd, months[em-1], ed, sy)
}
func nonempty(v []string) []string {
	out := []string{}
	for _, s := range v {
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

func numberLabel(value, kind, lang string) string {
	if lang == "ko" {
		suffix := "권"
		if kind == "issue" {
			suffix = "호"
		}
		return "제" + value + suffix
	}
	prefix := "Vol. "
	if kind == "issue" {
		prefix = "No. "
	}
	return prefix + value
}

// Detect only the explicitly supplied value; never infer metadata from titles.
func containsNumber(value, number, kind string) bool {
	if number == "" {
		return false
	}
	normalize := func(s string) string {
		return strings.Map(func(r rune) rune {
			if unicode.IsSpace(r) || r == '\ufeff' {
				return ' '
			}
			return r
		}, s)
	}
	value, number = normalize(value), normalize(number)
	escaped := regexp.QuoteMeta(number)
	english, korean := `vol(?:ume)?\.?`, "권"
	if kind == "issue" {
		english, korean = `(?:no\.?|issue|part)`, "호"
	}
	boundary := `($|[\s,;:.()\[\]])`
	return regexp.MustCompile(`(?i)(^|[^a-z0-9])`+english+`\s*`+escaped+boundary).MatchString(value) ||
		regexp.MustCompile(`(?i)(^|[^a-z0-9])(?:제\s*)?`+escaped+`\s*`+korean+`($|[\s,;:.()\[\]제])`).MatchString(value)
}

func seriesLabel(series, volume, lang string) string {
	if volume == "" || containsNumber(series, volume, "volume") {
		return series
	}
	label := numberLabel(volume, "volume", lang)
	if series != "" {
		return series + ", " + label
	}
	if lang == "ko" {
		return "시리즈 " + label
	}
	return "Series " + label
}

func referenceParts(item, settings record, language string) (string, string) {
	r := catalogFor(item, settings)
	if len(r) == 0 {
		return "", ""
	}
	lang := "en"
	if language == "ko" && strings.HasPrefix(text(item["publication_type"]), "domestic-") {
		lang = "ko"
	}
	name, pages := bibliographyLocalized(r, "name", lang), text(item["pages"])
	if isJournal(item) {
		parts := []string{name}
		if lang == "ko" {
			numbering := []string{}
			if v := text(item["volume"]); v != "" {
				numbering = append(numbering, "제"+v+"권")
			}
			if v := text(item["issue"]); v != "" {
				numbering = append(numbering, "제"+v+"호")
			}
			if len(numbering) > 0 {
				parts = append(parts, strings.Join(numbering, " "))
			}
		} else {
			if v := text(item["volume"]); v != "" {
				parts = append(parts, "Vol. "+v)
			}
			if v := text(item["issue"]); v != "" {
				parts = append(parts, "No. "+v)
			}
		}
		if pages != "" {
			parts = append(parts, "pp. "+pages)
		} else if v := text(item["elocator"]); v != "" {
			parts = append(parts, v)
		}
		citation := strings.Join(nonempty(parts), ", ")
		if len(parts) > 1 {
			citation += "."
		}
		return citation, ""
	}
	title, series := bibliographyLocalized(r, "proceedings_title", lang), bibliographyLocalized(r, "proceedings_series", lang)
	volume, issue, seriesVolume := text(r["proceedings_volume"]), text(r["proceedings_issue"]), text(r["proceedings_series_volume"])
	parts := []string{title}
	numbering := []string{}
	if volume != "" && !containsNumber(title, volume, "volume") {
		numbering = append(numbering, numberLabel(volume, "volume", lang))
	}
	if issue != "" && !containsNumber(title, issue, "issue") {
		numbering = append(numbering, numberLabel(issue, "issue", lang))
	}
	if len(numbering) > 0 {
		separator := ", "
		if lang == "ko" {
			separator = " "
		}
		parts = append(parts, strings.Join(numbering, separator))
	}
	seriesIndex := -1
	if series != "" {
		seriesIndex = strings.Index(lower(title), lower(series))
	}
	if seriesIndex < 0 {
		if label := seriesLabel(series, seriesVolume, lang); label != "" {
			parts = append(parts, label)
		}
	} else if seriesVolume != "" {
		// Use the original suffix rather than editing a title that includes the series.
		location := regexp.MustCompile(`(?i)` + regexp.QuoteMeta(series)).FindStringIndex(title)
		if location == nil || !containsNumber(title[location[0]:], seriesVolume, "volume") {
			parts = append(parts, seriesLabel("", seriesVolume, lang))
		}
	}
	if pages != "" {
		parts = append(parts, "pp. "+pages)
	}
	citation := strings.TrimSuffix(strings.Join(nonempty(parts), ", "), ".")
	details := strings.Join(nonempty([]string{name, conferencePeriod(r, lang), bibliographyLocalized(r, "location", lang)}), ", ")
	if citation != "" {
		citation += "."
	}
	if details != "" {
		details = strings.TrimSuffix(details, ".") + "."
	}
	return citation, details
}

// ResolvePublication derives dates and bilingual references without mutating item.
func ResolvePublication(item map[string]any, settings map[string]any) map[string]any {
	out := copyRecord(item)
	en, enConf := referenceParts(item, settings, "en")
	ko, koConf := referenceParts(item, settings, "ko")
	out["date"] = publicationDate(item, settings)
	out["venue_en"] = strings.Join(nonempty([]string{en, enConf}), " ")
	out["venue_kr"] = strings.Join(nonempty([]string{ko, koConf}), " ")
	out["reference_en"], out["reference_ko"], out["conference_info_en"], out["conference_info_ko"] = en, ko, enConf, koConf
	out["indexing"] = ""
	for _, v := range array(catalogFor(item, settings)["indexing"]) {
		if contains([]string{"SCIE", "SCOPUS", "KCI"}, str(v)) {
			out["indexing"] = v
			break
		}
	}
	if strings.HasSuffix(text(item["publication_type"]), "-conference") {
		conference := catalogFor(item, settings)
		if text(conference["proceedings_volume"]) != "" || text(conference["proceedings_issue"]) != "" || text(conference["proceedings_series_volume"]) != "" {
			out["volume"], out["issue"] = text(conference["proceedings_volume"]), text(conference["proceedings_issue"])
			out["proceedings_series_volume"] = text(conference["proceedings_series_volume"])
		}
	}
	return out
}
