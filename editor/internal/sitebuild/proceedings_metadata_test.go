package sitebuild

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestConferenceProceedingsNumberingValidation(t *testing.T) {
	for _, field := range []string{"proceedings_volume", "proceedings_issue", "proceedings_series_volume"} {
		t.Run(field, func(t *testing.T) {
			for _, value := range []string{"", "12A", " Suppl. 2 ", "I–IV", strings.Repeat("가", 500)} {
				conference := record{"id": "Test Conf. 2026", "name_en": "Conference", field: value, "future_metadata": record{"keep": true}}
				settings := record{"conferences": []any{conference}}
				before, err := json.Marshal(settings)
				if err != nil {
					t.Fatal(err)
				}
				validator := &validator{}
				validator.catalogs(settings)
				if len(validator.errors) != 0 {
					t.Fatalf("valid numbering rejected: %v", validator.errors)
				}
				after, err := json.Marshal(settings)
				if err != nil || !reflect.DeepEqual(before, after) {
					t.Fatal("validation changed numbering, IDs, or unknown metadata")
				}
			}
			for _, invalid := range []any{nil, json.Number("12"), true, []any{"12"}, record{"value": "12"}, strings.Repeat("가", 501), "12\x00A", "Suppl.\x7f2"} {
				validator := &validator{}
				validator.catalogs(record{"conferences": []any{record{"id": "Test Conf. 2026", "name_en": "Conference", field: invalid}}})
				if len(validator.errors) == 0 || !strings.Contains(strings.Join(validator.errors, "\n"), field) {
					t.Fatalf("invalid %s was not rejected: %#v, %v", field, invalid, validator.errors)
				}
			}
		})
	}
	validator := &validator{}
	validator.catalogs(record{"conferences": []any{record{"id": "Test Conf. 2026", "name_en": "Conference"}}})
	if len(validator.errors) != 0 {
		t.Fatalf("omitted optional proceedings fields were rejected: %v", validator.errors)
	}
}
