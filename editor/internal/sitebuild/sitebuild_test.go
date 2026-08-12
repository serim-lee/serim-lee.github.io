package sitebuild

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func repositoryRoot(t *testing.T) string {
	t.Helper()
	root, e := filepath.Abs(filepath.Join("..", "..", ".."))
	if e != nil {
		t.Fatal(e)
	}
	if _, e = os.Stat(filepath.Join(root, "data", "settings.json")); e != nil {
		t.Fatal(e)
	}
	return root
}
func repositoryData(t *testing.T) (string, record) {
	t.Helper()
	root := repositoryRoot(t)
	data, e := readData(root)
	if e != nil {
		t.Fatal(e)
	}
	return root, data
}
func first(data record, key string) record { return obj(array(data[key])[0]) }
func fixtureRoot(t *testing.T) string {
	t.Helper()
	source := repositoryRoot(t)
	target := t.TempDir()
	for _, directory := range []string{"data", "data/document"} {
		entries, e := os.ReadDir(filepath.Join(source, directory))
		if e != nil {
			t.Fatal(e)
		}
		if e = os.MkdirAll(filepath.Join(target, directory), 0755); e != nil {
			t.Fatal(e)
		}
		for _, entry := range entries {
			if !entry.Type().IsRegular() {
				continue
			}
			if directory == "data" && !strings.HasSuffix(entry.Name(), ".json") {
				continue
			}
			content, e := os.ReadFile(filepath.Join(source, directory, entry.Name()))
			if e != nil {
				t.Fatal(e)
			}
			if e = os.WriteFile(filepath.Join(target, directory, entry.Name()), content, 0644); e != nil {
				t.Fatal(e)
			}
		}
	}
	return target
}

func TestRepositoryValidationAndExactExports(t *testing.T) {
	root := fixtureRoot(t)
	output, e := Build(root, true)
	if e != nil {
		t.Fatal(e)
	}
	if !strings.Contains(output, "Validated ") {
		t.Fatalf("missing summary: %q", output)
	}
	if _, e = os.Stat(filepath.Join(root, "exports")); !os.IsNotExist(e) {
		t.Fatalf("validation wrote output directory: %v", e)
	}
	generated := filepath.Join(root, "exports")
	if e = os.MkdirAll(generated, 0755); e != nil {
		t.Fatal(e)
	}
	for _, name := range []string{"publications.bib", "publications.ris", "publications.csl.json", "build-report.json"} {
		if e = os.WriteFile(filepath.Join(generated, name), []byte("existing output must remain untouched"), 0644); e != nil {
			t.Fatal(e)
		}
	}
	if _, e = Build(root, true); e != nil {
		t.Fatal(e)
	}
	for _, name := range []string{"publications.bib", "publications.ris", "publications.csl.json", "build-report.json"} {
		got, e := os.ReadFile(filepath.Join(generated, name))
		if e != nil {
			t.Fatal(e)
		}
		if string(got) != "existing output must remain untouched" {
			t.Fatalf("validation changed %s", name)
		}
	}
	if _, e = Build(root, false); e != nil {
		t.Fatal(e)
	}
	firstBuild := map[string][]byte{}
	for _, name := range []string{"publications.bib", "publications.ris", "publications.csl.json", "build-report.json"} {
		got, e := os.ReadFile(filepath.Join(root, "exports", name))
		if e != nil {
			t.Fatal(e)
		}
		firstBuild[name] = got
		if baseline := os.Getenv("SITEBUILD_EXPORT_BASELINE"); baseline != "" {
			want, e := os.ReadFile(filepath.Join(baseline, name))
			if e != nil {
				t.Fatal(e)
			}
			if !bytes.Equal(got, want) {
				t.Errorf("%s differs from independently established export bytes", name)
			}
		}
		if strings.HasSuffix(name, ".json") && !json.Valid(got) {
			t.Errorf("%s is not valid JSON", name)
		}
	}
	if _, e = Build(root, false); e != nil {
		t.Fatal(e)
	}
	for name, want := range firstBuild {
		got, e := os.ReadFile(filepath.Join(root, "exports", name))
		if e != nil {
			t.Fatal(e)
		}
		if !bytes.Equal(got, want) {
			t.Errorf("%s changes on repeat build", name)
		}
	}
	var csl []record
	if e = json.Unmarshal(firstBuild["publications.csl.json"], &csl); e != nil {
		t.Fatal(e)
	}
	_, data := repositoryData(t)
	if len(csl) != len(array(data["publications"])) {
		t.Fatal("export lost publications")
	}
	if bytes.Count(firstBuild["publications.ris"], []byte("TY  - ")) != len(csl) {
		t.Fatal("RIS export lost publications")
	}
	for _, entry := range csl {
		if len(array(entry["author"])) == 0 || text(entry["id"]) == "" || text(entry["title"]) == "" {
			t.Fatal("CSL missing author/id/title")
		}
	}
}

func TestRejectMalformedData(t *testing.T) {
	type mutation func(record)
	tests := map[string]mutation{
		"publication unknown field":      func(d record) { first(d, "publications")["obsolete"] = true },
		"publication missing field":      func(d record) { delete(first(d, "publications"), "abstract_en") },
		"author unknown":                 func(d record) { first(d, "publications")["author_ids"] = []any{"missing-person"} },
		"author numeric":                 func(d record) { first(d, "publications")["author_ids"] = []any{json.Number("2")} },
		"author duplicate":               func(d record) { first(d, "publications")["author_ids"] = []any{"jo-hg", "jo-hg"} },
		"author empty":                   func(d record) { first(d, "publications")["author_ids"] = []any{} },
		"author noncanonical":            func(d record) { first(d, "publications")["author_ids"] = []any{" jo-hg"} },
		"publication award reference":    func(d record) { first(d, "publications")["award_id"] = "missing-award" },
		"publication topic reference":    func(d record) { first(d, "publications")["topic"] = "missing-topic" },
		"publication catalog reference":  func(d record) { first(d, "publications")["journal_id"] = "missing-journal" },
		"publication catalog mismatch":   func(d record) { first(d, "publications")["conference_id"] = "ASHRAE 2026W" },
		"publication catalog ID type":    func(d record) { first(d, "publications")["journal_id"] = true },
		"publication manual venue":       func(d record) { first(d, "publications")["venue_en"] = "obsolete" },
		"publication scalar type":        func(d record) { first(d, "publications")["title_en"] = 42 },
		"publication empty title":        func(d record) { r := first(d, "publications"); r["title_en"], r["title_ko"] = "", "" },
		"publication type":               func(d record) { first(d, "publications")["publication_type"] = "journal" },
		"publication status type":        func(d record) { first(d, "publications")["in_press"] = "false" },
		"publication conflicting status": func(d record) { r := first(d, "publications"); r["in_press"], r["under_review"] = true, true },
		"publication old date":           func(d record) { first(d, "publications")["date"] = "2026-02-30" },
		"publication online date":        func(d record) { first(d, "publications")["online_date"] = "2026-13" },
		"publication zero year":          func(d record) { first(d, "publications")["issue_date"] = "0000" },
		"publication precision":          func(d record) { first(d, "publications")["issue_date"] = "2026-1" },
		"publication control":            func(d record) { first(d, "publications")["pages"] = "1\x00-2" },
		"publication long numbering":     func(d record) { first(d, "publications")["volume"] = strings.Repeat("가", 501) },
		"publication keywords uppercase": func(d record) { first(d, "publications")["keywords_en"] = []any{"Energy"} },
		"publication keywords blank":     func(d record) { first(d, "publications")["keywords_ko"] = []any{" "} },
		"visibility type":                func(d record) { first(d, "publications")["visible_in_CV"] = "true" },
		"awards unknown field":           func(d record) { first(d, "awards")["obsolete"] = true },
		"award invalid date":             func(d record) { first(d, "awards")["date"] = "2026-2-3" },
		"award localization":             func(d record) { r := first(d, "awards"); r["organization_en"], r["organization_ko"] = "", "" },
		"people duplicate":               func(d record) { array(d["people"])[1].(map[string]any)["id"] = first(d, "people")["id"] },
		"people self": func(d record) {
			for _, x := range array(d["people"]) {
				obj(x)["is_self"] = false
			}
		},
		"people name":               func(d record) { r := first(d, "people"); r["name_en"], r["name_ko"] = "", "" },
		"people notes":              func(d record) { first(d, "people")["notes_en"] = []any{""} },
		"software stage":            func(d record) { first(d, "software")["stage"] = "complete" },
		"software obsolete summary": func(d record) { first(d, "software")["summary"] = "obsolete" },
		"software blank unmatched note": func(d record) {
			first(d, "software")["notes_en"] = []any{"one", "two", " "}
			first(d, "software")["notes_kr"] = []any{"one"}
		},
		"project id":             func(d record) { first(d, "projects")["id"] = "unused" },
		"project period":         func(d record) { first(d, "projects")["period"] = "unused" },
		"project notes":          func(d record) { first(d, "projects")["notes_kr"] = []any{} },
		"project theme":          func(d record) { first(d, "projects")["theme"] = "missing-theme" },
		"project dates":          func(d record) { first(d, "projects")["start_date"] = "2026-02-30" },
		"project reversed dates": func(d record) { r := first(d, "projects"); r["start_date"], r["end_date"] = "2026-02-03", "2026-01-01" },
		"project caption pair": func(d record) {
			first(d, "projects")["media"] = []any{record{"caption_en": "Caption", "caption_ko": ""}}
		},
		"project obsolete caption":   func(d record) { first(d, "projects")["media"] = []any{record{"caption_kr": "old"}} },
		"board old date":             func(d record) { first(d, "board")["date"] = "2026-01-01" },
		"board invalid start":        func(d record) { first(d, "board")["start_date"] = "2026-13-01" },
		"board reversed dates":       func(d record) { r := first(d, "board"); r["start_date"], r["end_date"] = "2026-02-03", "2026-01-01" },
		"board content":              func(d record) { first(d, "board")["content_ko"] = false },
		"settings theme":             func(d record) { obj(d["settings"])["color_theme"] = "unknown" },
		"settings sections":          func(d record) { obj(d["settings"])["cv_sections"] = []any{"education", "education"} },
		"settings section nonstring": func(d record) { obj(d["settings"])["main_page_sections"] = []any{record{"x": true}} },
		"topic duplicate folded label": func(d record) {
			topics := array(obj(d["settings"])["topics"])
			obj(topics[0])["label_en"] = "Straße"
			obj(topics[1])["label_en"] = "STRASSE"
		},
		"topic color":       func(d record) { obj(array(obj(d["settings"])["topics"])[0])["color"] = "#abc" },
		"topic blank label": func(d record) { obj(array(obj(d["settings"])["topics"])[0])["label_ko"] = " " },
		"topic fallback":    func(d record) { obj(d["settings"])["topic_fallback"] = nil },
		"catalog duplicate": func(d record) {
			r := obj(d["settings"])
			rows := array(r["journals"])
			obj(rows[1])["id"] = obj(rows[0])["id"]
		},
		"catalog uppercase journal ID": func(d record) { obj(array(obj(d["settings"])["journals"])[0])["id"] = "Journal" },
		"catalog ISSN":                 func(d record) { obj(array(obj(d["settings"])["journals"])[0])["issn"] = "1234567X" },
		"catalog URL":                  func(d record) { obj(array(obj(d["settings"])["journals"])[0])["url"] = "javascript:alert(1)" },
		"catalog text control":         func(d record) { obj(array(obj(d["settings"])["journals"])[0])["publisher_en"] = "bad\x01control" },
		"catalog text type":            func(d record) { obj(array(obj(d["settings"])["journals"])[0])["name_en"] = false },
		"catalog text length": func(d record) {
			obj(array(obj(d["settings"])["journals"])[0])["publisher_en"] = strings.Repeat("가", 501)
		},
		"catalog indexing":         func(d record) { obj(array(obj(d["settings"])["journals"])[0])["indexing"] = []any{"SCIE", "SCOPUS"} },
		"catalog conference date":  func(d record) { obj(array(obj(d["settings"])["conferences"])[0])["start_date"] = "2026-01" },
		"catalog proceedings date": func(d record) { obj(array(obj(d["settings"])["conferences"])[0])["proceedings_date"] = "2026-02-30" },
		"catalog reversed dates": func(d record) {
			r := obj(array(obj(d["settings"])["conferences"])[0])
			r["start_date"], r["end_date"] = "2026-02-03", "2026-01-01"
		},
		"catalog too many": func(d record) {
			rows := []any{}
			for i := 0; i < 1001; i++ {
				rows = append(rows, record{})
			}
			obj(d["settings"])["journals"] = rows
		},
		"profile missing section":      func(d record) { delete(obj(d["profile"]), "education") },
		"profile credentials":          func(d record) { obj(obj(d["profile"])["profile_card"])["credentials"] = false },
		"scholarship obsolete details": func(d record) { obj(array(obj(d["profile"])["scholarships"])[0])["details"] = "obsolete" },
		"scholarship amount bool":      func(d record) { obj(obj(array(obj(d["profile"])["scholarships"])[0])["amount"])["value"] = true },
		"scholarship amount negative":  func(d record) { obj(obj(array(obj(d["profile"])["scholarships"])[0])["amount"])["value"] = -1 },
		"scholarship amount infinity":  func(d record) { obj(obj(array(obj(d["profile"])["scholarships"])[0])["amount"])["value"] = math.Inf(1) },
		"scholarship currency":         func(d record) { obj(obj(array(obj(d["profile"])["scholarships"])[0])["amount"])["currency"] = "krw" },
		"education GPA bool": func(d record) {
			obj(array(obj(d["profile"])["education"])[0])["gpa"] = record{"value": true, "max": 4.3}
		},
		"education GPA range": func(d record) {
			obj(array(obj(d["profile"])["education"])[0])["gpa"] = record{"value": 5.0, "max": 4.3}
		},
		"education GPA large integer precision": func(d record) {
			obj(array(obj(d["profile"])["education"])[0])["gpa"] = record{"value": json.Number("9007199254740993"), "max": json.Number("9007199254740992")}
		},
		"education thesis URL":      func(d record) { obj(array(obj(d["profile"])["education"])[0])["thesis_url"] = "file:///private" },
		"activity schema bool":      func(d record) { obj(d["academic_activities"])["schema_version"] = true },
		"activity schema float":     func(d record) { obj(d["academic_activities"])["schema_version"] = json.Number("1.0") },
		"activity missing category": func(d record) { delete(obj(d["academic_activities"]), "reviews") },
		"activity review dates": func(d record) {
			obj(d["academic_activities"])["reviews"] = []any{record{"journal_en": "Journal", "journal_ko": "", "completed_dates": []any{"2026-01-01", "2026-01-01"}}}
		},
		"activity empty review dates": func(d record) {
			obj(d["academic_activities"])["reviews"] = []any{record{"journal_en": "Journal", "journal_ko": "", "completed_dates": []any{}}}
		},
		"activity invalid review date": func(d record) {
			obj(d["academic_activities"])["reviews"] = []any{record{"journal_en": "Journal", "journal_ko": "", "completed_dates": []any{"2026-02-30"}}}
		},
		"activity missing localization": func(d record) {
			obj(d["academic_activities"])["invited_talks"] = []any{record{"event_en": "", "event_ko": "", "topic_en": "", "topic_ko": "", "date": "2026-01-01"}}
		},
		"activity reversed dates": func(d record) {
			obj(d["academic_activities"])["conference_service"] = []any{record{"conference_en": "Conference", "conference_ko": "", "role_en": "Member", "role_ko": "", "start_date": "2026-02-01", "end_date": "2026-01-01"}}
		},
		"activity empty required start": func(d record) {
			obj(d["academic_activities"])["conference_service"] = []any{record{"conference_en": "Conference", "conference_ko": "", "role_en": "Member", "role_ko": "", "start_date": "", "end_date": ""}}
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			root, data := repositoryData(t)
			mutate(data)
			if errors := validateData(root, data); len(errors) == 0 {
				t.Fatal("malformed fixture was accepted")
			}
		})
	}
}

func TestLegacyMigrationsAndAllowedUnknownFields(t *testing.T) {
	root, data := repositoryData(t)
	settings := obj(data["settings"])
	settings["schema_version"] = json.Number("5")
	for _, field := range []string{"main_page_sections", "hidden_main_page_sections"} {
		next := []any{}
		for _, value := range array(settings[field]) {
			if value != "academic_activities" && value != "patents" {
				next = append(next, value)
			}
		}
		settings[field] = next
	}
	for _, field := range []string{"cv_sections", "hidden_cv_sections"} {
		next := []any{}
		for _, value := range array(settings[field]) {
			if value != "patents" {
				next = append(next, value)
			}
		}
		settings[field] = next
	}
	topics := settings["topics"]
	delete(settings, "topics")
	settings["project_themes"], settings["publication_topics"] = topics, topics
	settings["project_theme_fallback"], settings["publication_topic_fallback"] = settings["topic_fallback"], settings["topic_fallback"]
	delete(settings, "topic_fallback")
	first(data, "people")["custom"] = record{"preserve": true}
	first(data, "projects")["custom"] = true
	first(data, "software")["custom"] = true
	obj(array(settings["journals"])[0])["custom"] = true
	if errors := validateData(root, data); len(errors) > 0 {
		t.Fatal(strings.Join(errors, "\n"))
	}
	if !sectionSet(settings["main_page_sections"], settings["hidden_main_page_sections"], mainSections) {
		t.Fatal("legacy sections not migrated")
	}
	if has(settings, "project_themes") || !has(settings, "topics") {
		t.Fatal("legacy topics not migrated")
	}
	settings["schema_version"] = json.Number("4")
	delete(settings, "cv_sections")
	delete(settings, "hidden_cv_sections")
	if errors := validateData(root, data); len(errors) > 0 {
		t.Fatal(strings.Join(errors, "\n"))
	}
}

func TestIndependentContentNotes(t *testing.T) {
	for _, collection := range []string{"projects", "software"} {
		for _, counts := range [][2]int{{1, 3}, {3, 1}} {
			t.Run(fmt.Sprintf("%s_en%d_kr%d", collection, counts[0], counts[1]), func(t *testing.T) {
				root, data := repositoryData(t)
				item := first(data, collection)
				for language, count := range map[string]int{"en": counts[0], "kr": counts[1]} {
					notes := make([]any, count)
					for index := range notes {
						notes[index] = fmt.Sprintf("%s description %d", language, index+1)
					}
					item["notes_"+language] = notes
				}
				before, _ := json.Marshal([]any{item["notes_en"], item["notes_kr"]})
				if problems := validateData(root, data); len(problems) > 0 {
					t.Fatal(strings.Join(problems, "\n"))
				}
				after, _ := json.Marshal([]any{item["notes_en"], item["notes_kr"]})
				if !bytes.Equal(before, after) {
					t.Fatal("validation padded, truncated or reordered language lists")
				}
				longer := "notes_en"
				if counts[1] > counts[0] {
					longer = "notes_kr"
				}
				last := array(item[longer])
				last[len(last)-1] = " "
				if problems := validateData(root, data); len(problems) == 0 {
					t.Fatal("invalid unmatched trailing note was accepted")
				}
			})
		}
	}
}

func TestDocumentReferences(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, "data", "document")
	if e := os.MkdirAll(directory, 0755); e != nil {
		t.Fatal(e)
	}
	if e := os.WriteFile(filepath.Join(directory, "valid.PDF"), []byte("%PDF-test"), 0644); e != nil {
		t.Fatal(e)
	}
	if e := os.WriteFile(filepath.Join(directory, "fake.pdf"), []byte("hello"), 0644); e != nil {
		t.Fatal(e)
	}
	for _, name := range []any{true, "../escape.pdf", "name.txt", " bad.pdf", "missing.pdf", "fake.pdf", strings.Repeat("가", 81) + ".pdf"} {
		v := &validator{root: root}
		v.document(record{"pdf": name}, "fixture")
		if len(v.errors) == 0 {
			t.Errorf("accepted invalid PDF reference %v", name)
		}
	}
	for _, name := range []string{"", "valid.PDF"} {
		v := &validator{root: root}
		v.document(record{"pdf": name}, "fixture")
		if len(v.errors) > 0 {
			t.Fatal(v.errors)
		}
	}
	if e := os.Symlink(filepath.Join(directory, "valid.PDF"), filepath.Join(directory, "link.pdf")); e == nil {
		v := &validator{root: root}
		v.document(record{"pdf": "link.pdf"}, "fixture")
		if len(v.errors) == 0 {
			t.Fatal("accepted PDF symlink")
		}
	}
}

func TestResolvePublicationAndExportUnicode(t *testing.T) {
	settings := record{"journals": []any{record{"id": "journal", "name_en": "English Journal", "name_ko": "국문 저널", "indexing": []any{"SCIE"}}}, "conferences": []any{record{"id": "Conf. 2026", "name_en": "Conference", "name_ko": "학회", "start_date": "2026-01-30", "end_date": "2026-02-02", "proceedings_date": "2026-03", "proceedings_title_en": "Proceedings", "proceedings_title_ko": "논문집", "proceedings_series_en": "Series", "location_en": "Seoul", "location_ko": "서울", "indexing": []any{"SCOPUS"}}}}
	journal := record{"publication_type": "international-journal", "journal_id": "journal", "online_date": "2025", "issue_date": "2026-04", "volume": "2", "issue": "1", "pages": "1-9", "custom": true}
	resolved := ResolvePublication(journal, settings)
	if resolved["date"] != "2026-04" || resolved["venue_en"] != "English Journal, Vol. 2, No. 1, pp. 1-9." || resolved["venue_en"] != resolved["venue_kr"] || resolved["custom"] != true {
		t.Fatal(resolved)
	}
	if has(journal, "venue_en") {
		t.Fatal("input mutated")
	}
	journal["publication_type"] = "domestic-journal"
	if ResolvePublication(journal, settings)["reference_ko"] != "국문 저널, 제2권 제1호, pp. 1-9." {
		t.Fatal("Korean numbering lost")
	}
	journal["pages"] = ""
	journal["elocator"] = "e123"
	if !strings.HasSuffix(str(ResolvePublication(journal, settings)["reference_en"]), ", e123.") {
		t.Fatal("elocator fallback lost")
	}
	conference := record{"publication_type": "domestic-conference", "conference_id": "Conf. 2026", "pages": "10-20"}
	r := ResolvePublication(conference, settings)
	if r["date"] != "2026-03" || r["reference_en"] != "Proceedings, Series, pp. 10-20." || r["conference_info_en"] != "Conference, January 30-February 2, 2026, Seoul." || r["conference_info_ko"] != "학회, 2026년 1월 30일-2월 2일, 서울." {
		t.Fatal(r)
	}
	unknown := ResolvePublication(record{"publication_type": "international-journal", "journal_id": "unknown", "online_date": "2026"}, settings)
	if unknown["date"] != "2026" || unknown["venue_en"] != "" {
		t.Fatal(unknown)
	}
	got := prettyJSON(orderedObject{{"unicode", "한글 <&>\u2028\u2029\n\""}, {"array", []any{1, "x"}}}, 0) + "\n"
	want := "{\n  \"unicode\": \"한글 <&>\u2028\u2029\\n\\\"\",\n  \"array\": [\n    1,\n    \"x\"\n  ]\n}\n"
	if got != want {
		t.Fatalf("Unicode JSON differs\ngot %q\nwant%q", got, want)
	}
	if !reflect.DeepEqual(dateParts(record{"date": "2026-03"}), []int{2026, 3}) {
		t.Fatal("partial dates expanded")
	}
}

func TestReadDataRejectsMalformedEncodingAndTrailingJSON(t *testing.T) {
	root := fixtureRoot(t)
	path := filepath.Join(root, "data", "people.json")
	for _, content := range [][]byte{[]byte(`[] []`), []byte(`{"broken":`), append([]byte(`[{"id":"`), 0xff, '"', '}', ']')} {
		if e := os.WriteFile(path, content, 0644); e != nil {
			t.Fatal(e)
		}
		if _, e := readData(root); e == nil {
			t.Fatalf("accepted invalid JSON/UTF8 %q", content)
		}
	}
}

func TestLegacyAcceptedEdgesAndExactCitationKeys(t *testing.T) {
	root, data := repositoryData(t)
	first(data, "board")["start_date"] = "2026-1-2"
	first(data, "board")["end_date"] = "2026-11-2"
	catalog := obj(array(obj(data["settings"])["journals"])[0])
	catalog["url"] = "https://example.org:bad/%"
	catalog["publisher_en"] = strings.Repeat("😀", 500)
	obj(obj(array(obj(data["profile"])["scholarships"])[0])["amount"])["value"] = json.Number(strings.Repeat("9", 400))
	if errors := validateData(root, data); len(errors) > 0 {
		t.Fatal(strings.Join(errors, "\n"))
	}
	for _, test := range []struct {
		item record
		want string
	}{
		{record{"doi": " 10.1234/example ", "date": "2026"}, "publication-2026-8t2fr"},
		{record{"doi": "  ", "url": "url", "date": "2026"}, "publication-2026-1dbptvx"},
		{record{"doi": "", "url": " https://example.org/ ", "date": "2026"}, "publication-2026-m4gxe1"},
	} {
		if got := citationKey(test.item); got != test.want {
			t.Errorf("citation hash %v: got %s want %s", test.item, got, test.want)
		}
	}
	for _, test := range []struct {
		url   string
		valid bool
	}{{"https://example.org:bad/%", true}, {"https://example.org/50%", true}, {"https://[::1]/", true}, {"https://[not-ip]/", false}, {"https://", false}, {"javascript:example", false}} {
		if got := httpURL(test.url); got != test.valid {
			t.Errorf("URL %q: got %v", test.url, got)
		}
	}
}
