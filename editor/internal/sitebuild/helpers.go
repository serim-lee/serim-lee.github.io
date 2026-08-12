// Package sitebuild validates the site's JSON data and creates bibliography exports.
package sitebuild

import (
	"encoding/json"
	"fmt"
	"github.com/Gonie-Gonie/Gonie-Gonie.github.io/editor/internal/dataversion"
	"math"
	"math/big"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type record = map[string]any

var canonicalID = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
var conferenceID = regexp.MustCompile(`^[A-Za-z0-9]+(?:[ -][A-Za-z0-9]+\.?)*$`)
var fullDate = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}$`)
var partialDate = regexp.MustCompile(`^[0-9]{4}(?:-[0-9]{2}(?:-[0-9]{2})?)?$`)
var controls = regexp.MustCompile(`[\x00-\x1f\x7f-\x{009f}]`)
var mainSections = []string{"experience", "education", "scholarships", "certifications", "awards", "academic_activities", "patents", "teaching", "skills"}
var cvSections = []string{"experience", "education", "projects", "publications", "software", "awards", "academic_activities", "patents", "teaching", "scholarships", "certifications", "skills"}
var profileSections = []string{"experience", "education", "scholarships", "teaching", "certifications", "patents", "skills"}
var patentStringFields = []string{"title_en", "title_ko", "type", "application_number", "application_date", "publication_number", "publication_date", "registration_number", "registration_date", "applicant_en", "applicant_ko", "url"}
var patentTypes = []string{"domestic", "international"}
var publicationFields = []string{"title_en", "title_ko", "abstract_en", "abstract_ko", "keywords_en", "keywords_ko", "author_ids", "date", "under_review", "in_press", "publication_type", "topic", "award_id", "doi", "url", "note"}
var awardFields = []string{"id", "date", "title_en", "title_ko", "organization_en", "organization_ko"}
var activityOrder = []string{"reviews", "invited_talks", "conference_service", "professional_service", "editorial_service"}
var activityFields = map[string][]string{
	"reviews":              {"journal_en", "journal_ko", "completed_dates"},
	"invited_talks":        {"event_en", "event_ko", "topic_en", "topic_ko", "date"},
	"conference_service":   {"conference_en", "conference_ko", "role_en", "role_ko", "start_date", "end_date"},
	"professional_service": {"organization_en", "organization_ko", "role_en", "role_ko", "start_date", "end_date"},
	"editorial_service":    {"journal_en", "journal_ko", "role_en", "role_ko", "start_date", "end_date"},
}

func obj(v any) record    { r, _ := v.(map[string]any); return r }
func array(v any) []any   { r, _ := v.([]any); return r }
func isString(v any) bool { _, ok := v.(string); return ok }
func str(v any) string    { s, _ := v.(string); return s }
func pyString(v any) string {
	switch x := v.(type) {
	case nil:
		return "None"
	case bool:
		if x {
			return "True"
		}
		return "False"
	case string:
		return x
	case json.Number:
		return string(x)
	default:
		return fmt.Sprint(v)
	}
}
func pyBool(v any) bool {
	switch x := v.(type) {
	case nil:
		return false
	case bool:
		return x
	case string:
		return x != ""
	case []any:
		return len(x) > 0
	case map[string]any:
		return len(x) > 0
	case json.Number:
		n, _ := x.Float64()
		return n != 0
	case float64:
		return x != 0
	case int:
		return x != 0
	}
	return true
}
func text(v any) string {
	if !pyBool(v) {
		return ""
	}
	return strings.TrimSpace(pyString(v))
}
func lower(v string) string { return cases.Lower(language.Und).String(v) }
func truthy(v any) bool {
	if b, ok := v.(bool); ok {
		return b
	}
	return contains([]string{"true", "1", "yes", "y"}, lower(text(v)))
}
func contains(values []string, v string) bool {
	for _, s := range values {
		if s == v {
			return true
		}
	}
	return false
}
func has(m record, key string) bool { _, ok := m[key]; return ok }
func copyRecord(m record) record {
	out := record{}
	for k, v := range m {
		out[k] = v
	}
	return out
}
func localized(m record, key, lang string) string {
	for _, k := range []string{key + "_" + lang, key, key + "_en", key + "_ko"} {
		if pyBool(m[k]) {
			return pyString(m[k])
		}
	}
	return ""
}
func splitList(v any) []string {
	out := []string{}
	if values, ok := v.([]any); ok {
		for _, x := range values {
			s := strings.TrimSpace(pyString(x))
			if s != "" {
				out = append(out, s)
			}
		}
		return out
	}
	for _, part := range strings.Split(text(v), ";") {
		if s := strings.TrimSpace(part); s != "" {
			out = append(out, s)
		}
	}
	return out
}
func localizedList(m record, key, lang string) []string {
	for _, k := range []string{key + "_" + lang, key, key + "_en", key + "_ko"} {
		if v := splitList(m[k]); len(v) > 0 {
			return v
		}
	}
	return []string{}
}
func personName(m record, lang string) string {
	for _, k := range []string{"name_" + lang, "name_en", "name_ko", "id"} {
		if pyBool(m[k]) {
			return pyString(m[k])
		}
	}
	return ""
}
func isJournal(m record) bool { return strings.HasSuffix(str(m["publication_type"]), "-journal") }
func publicationStatus(m record) string {
	if truthy(m["under_review"]) {
		return "under review"
	}
	if truthy(m["in_press"]) {
		return "in press"
	}
	return ""
}
func statusRank(m record) int {
	if truthy(m["under_review"]) {
		return 2
	}
	if truthy(m["in_press"]) {
		return 1
	}
	return 0
}
func dateParts(m record) []int {
	out := []int{}
	for _, part := range strings.Split(str(m["date"]), "-") {
		if part != "" {
			n, _ := strconv.Atoi(part)
			out = append(out, n)
		}
	}
	return out
}
func year(m record) string {
	p := dateParts(m)
	if len(p) > 0 {
		return strconv.Itoa(p[0])
	}
	return ""
}
func validID(v any) bool { s, ok := v.(string); return ok && canonicalID.MatchString(s) }
func validCatalogID(kind string, v any) bool {
	if kind == "conferences" {
		s, ok := v.(string)
		return ok && conferenceID.MatchString(s)
	}
	return validID(v)
}
func number(v any) (*big.Rat, bool) {
	var n float64
	switch x := v.(type) {
	case json.Number:
		if !strings.ContainsAny(string(x), ".eE") {
			integer, ok := new(big.Int).SetString(string(x), 10)
			if !ok {
				return nil, false
			}
			return new(big.Rat).SetInt(integer), true
		}
		var e error
		n, e = x.Float64()
		if e != nil {
			return nil, false
		}
	case float64:
		n = x
	case int:
		return new(big.Rat).SetInt64(int64(x)), true
	case int64:
		return new(big.Rat).SetInt64(x), true
	default:
		return nil, false
	}
	if math.IsNaN(n) || math.IsInf(n, 0) {
		return nil, false
	}
	return new(big.Rat).SetFloat64(n), true
}
func integer(v any) (int, bool) {
	switch x := v.(type) {
	case json.Number:
		if strings.ContainsAny(string(x), ".eE") {
			return 0, false
		}
		n, e := strconv.Atoi(string(x))
		return n, e == nil
	case int:
		return x, true
	case int64:
		return int(x), true
	}
	return 0, false
}
func validDate(s string, partial bool) bool {
	if (!partial && !fullDate.MatchString(s)) || (partial && !partialDate.MatchString(s)) {
		return false
	}
	layout := "2006-01-02"
	if len(s) == 4 {
		layout = "2006"
	} else if len(s) == 7 {
		layout = "2006-01"
	}
	t, e := time.Parse(layout, s)
	return e == nil && t.Year() >= 1 && t.Format(layout) == s
}
func readData(root string) (record, error) {
	if err := dataversion.ValidateBundle(root); err != nil {
		return nil, err
	}
	out := record{}
	for _, name := range []string{"settings", "profile", "people", "projects", "software", "publications", "awards", "board", "academic_activities"} {
		path := filepath.Join(root, "data", name+".json")
		content, e := os.ReadFile(path)
		if e != nil {
			return nil, e
		}
		v, e := dataversion.Unwrap(content, name+".json")
		if e != nil {
			return nil, fmt.Errorf("%s: %w", path, e)
		}

		out[name] = v
	}
	return out, nil
}
