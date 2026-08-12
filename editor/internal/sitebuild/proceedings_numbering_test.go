package sitebuild

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func proceedingsFixture() (record, record, record) {
	conference := record{
		"id": "Conf. 2026", "name_en": "Test Conference", "name_ko": "시험 학술대회",
		"proceedings_title_en": "Proceedings of Test Conference", "proceedings_title_ko": "시험 학술대회 논문집",
		"proceedings_series_en": "Research Series", "proceedings_series_ko": "연구 시리즈",
	}
	settings := record{"conferences": []any{conference}}
	item := record{"publication_type": "international-conference", "conference_id": "Conf. 2026", "pages": "10-20", "volume": "stale", "issue": "stale"}
	return item, settings, conference
}

func TestProceedingsNumberingReferencesAndResolution(t *testing.T) {
	item, settings, conference := proceedingsFixture()
	conference["proceedings_volume"], conference["proceedings_issue"], conference["proceedings_series_volume"] = "40", "2", "7"
	beforeItem, _ := json.Marshal(item)
	beforeSettings, _ := json.Marshal(settings)
	english := "Proceedings of Test Conference, Vol. 40, No. 2, Research Series, Vol. 7, pp. 10-20."
	resolved := ResolvePublication(item, settings)
	if resolved["reference_en"] != english || resolved["reference_ko"] != english || resolved["volume"] != "40" || resolved["issue"] != "2" || resolved["proceedings_series_volume"] != "7" {
		t.Fatalf("incorrect international proceedings resolution: %#v", resolved)
	}
	afterItem, _ := json.Marshal(item)
	afterSettings, _ := json.Marshal(settings)
	if string(beforeItem) != string(afterItem) || string(beforeSettings) != string(afterSettings) {
		t.Fatal("reference resolution mutated source publication or catalog")
	}
	item["publication_type"] = "domestic-conference"
	resolved = ResolvePublication(item, settings)
	if resolved["reference_ko"] != "시험 학술대회 논문집, 제40권 제2호, 연구 시리즈, 제7권, pp. 10-20." || resolved["reference_en"] != english {
		t.Fatalf("incorrect bilingual proceedings: %#v", resolved)
	}
}

func TestProceedingsSeriesVolumeAndDuplicateNumbering(t *testing.T) {
	for _, test := range []struct {
		name, title, series, volume, issue, seriesVolume, want string
	}{
		{"series number without name", "Proceedings", "", "", "", "7", "Proceedings, Series Vol. 7, pp. 10-20."},
		{"existing title numbers", "Proceedings Vol. 40, No. 2", "", "40", "2", "", "Proceedings Vol. 40, No. 2, pp. 10-20."},
		{"existing long labels", "Proceedings Volume 40, Issue 2", "", "40", "2", "", "Proceedings Volume 40, Issue 2, pp. 10-20."},
		{"existing part label", "Proceedings Vol. 132, Part 1", "", "132", "1", "", "Proceedings Vol. 132, Part 1, pp. 10-20."},
		{"Unicode number spacing", "Proceedings Vol.\u00a02", "", "2", "", "", "Proceedings Vol.\u00a02, pp. 10-20."},
		{"existing series number", "Proceedings", "Research Series, Vol. 7", "", "", "7", "Proceedings, Research Series, Vol. 7, pp. 10-20."},
		{"series included in title", "Proceedings in Research Series", "Research Series", "", "", "7", "Proceedings in Research Series, Series Vol. 7, pp. 10-20."},
		{"series is title prefix", "Proceedings of X", "Proceedings", "", "", "591", "Proceedings of X, Series Vol. 591, pp. 10-20."},
		{"series and number in title", "Proceedings in Research Series, Vol. 7", "Research Series", "", "", "7", "Proceedings in Research Series, Vol. 7, pp. 10-20."},
		{"distinct numeric boundary", "Proceedings Vol. 40", "", "4", "", "", "Proceedings Vol. 40, Vol. 4, pp. 10-20."},
		{"literal numbering punctuation", "Proceedings Vol. 40(1)", "", "40(1)", "", "", "Proceedings Vol. 40(1), pp. 10-20."},
	} {
		t.Run(test.name, func(t *testing.T) {
			item, settings, conference := proceedingsFixture()
			conference["proceedings_title_en"], conference["proceedings_series_en"] = test.title, test.series
			conference["proceedings_series_ko"] = ""
			conference["proceedings_volume"], conference["proceedings_issue"], conference["proceedings_series_volume"] = test.volume, test.issue, test.seriesVolume
			if got := ResolvePublication(item, settings)["reference_en"]; got != test.want {
				t.Fatalf("reference = %q, want %q", got, test.want)
			}
		})
	}
	item, settings, conference := proceedingsFixture()
	item["publication_type"] = "domestic-conference"
	conference["proceedings_series_en"], conference["proceedings_series_ko"] = "", ""
	conference["proceedings_series_volume"] = "7"
	if got := ResolvePublication(item, settings)["reference_ko"]; got != "시험 학술대회 논문집, 시리즈 제7권, pp. 10-20." {
		t.Fatalf("unnamed Korean series = %q", got)
	}
	conference["proceedings_title_ko"] = "시험 학술대회 논문집 제40권 제2호"
	conference["proceedings_volume"], conference["proceedings_issue"] = "40", "2"
	if got := str(ResolvePublication(item, settings)["reference_ko"]); strings.Count(got, "제40권") != 1 || strings.Count(got, "제2호") != 1 {
		t.Fatalf("duplicate Korean numbering: %s", got)
	}
	conference["proceedings_title_ko"] = "논문집 제40권제2호"
	if got := str(ResolvePublication(item, settings)["reference_ko"]); strings.Count(got, "제40권") != 1 || strings.Count(got, "제2호") != 1 {
		t.Fatalf("duplicate unspaced Korean numbering: %s", got)
	}
}

func TestBlankProceedingsNumberingPreservesReferencesAndExports(t *testing.T) {
	item, settings, conference := proceedingsFixture()
	baseline := ResolvePublication(item, settings)
	for _, field := range []string{"proceedings_volume", "proceedings_issue", "proceedings_series_volume"} {
		conference[field] = "  "
	}
	blank := ResolvePublication(item, settings)
	if !reflect.DeepEqual(blank, baseline) || bibtexEntry(blank) != bibtexEntry(baseline) || risEntry(blank) != risEntry(baseline) || prettyJSON(cslEntry(blank), 0) != prettyJSON(cslEntry(baseline), 0) {
		t.Fatal("blank optional proceedings fields changed existing references/exports")
	}
	conference["proceedings_volume"], conference["proceedings_issue"], conference["proceedings_series_volume"] = "40", "2", "7"
	numbered := ResolvePublication(item, settings)
	for name, output := range map[string]string{"BibTeX": bibtexEntry(numbered), "RIS": risEntry(numbered), "CSL": prettyJSON(cslEntry(numbered), 0)} {
		for _, required := range []string{"Vol. 40", "No. 2", "Research Series, Vol. 7"} {
			if !strings.Contains(output, required) {
				t.Fatalf("%s lost proceedings metadata %q: %s", name, required, output)
			}
		}
	}
}

func TestProceedingsFieldsDoNotAffectJournalNumbering(t *testing.T) {
	journal := record{"id": "journal", "name_en": "Test Journal", "name_ko": "시험 저널", "proceedings_volume": "40", "proceedings_issue": "2", "proceedings_series_volume": "7"}
	item := record{"publication_type": "domestic-journal", "journal_id": "journal", "volume": "11", "issue": "3", "pages": "10-20"}
	resolved := ResolvePublication(item, record{"journals": []any{journal}})
	if resolved["reference_en"] != "Test Journal, Vol. 11, No. 3, pp. 10-20." || resolved["reference_ko"] != "시험 저널, 제11권 제3호, pp. 10-20." || resolved["volume"] != "11" || resolved["issue"] != "3" {
		t.Fatalf("journal numbering changed: %#v", resolved)
	}
}
