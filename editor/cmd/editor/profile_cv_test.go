package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestLegacySettingsDefaultCVPartitionAndRejectInvalidCVPartition(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	if err := os.WriteFile(path, fixtureSettingsJSON(), 0o644); err != nil {
		t.Fatal(err)
	}
	document, _, err := readSettings(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := normaliseLoadedSettings(&document); err != nil {
		t.Fatal(err)
	}
	legacyCVSections := []string{}
	for _, section := range supportedCVSections {
		if section != "patents" {
			legacyCVSections = append(legacyCVSections, section)
		}
	}
	if document.SchemaVersion != settingsSchemaVersion || !reflect.DeepEqual(document.CVSections, legacyCVSections) || !reflect.DeepEqual(document.HiddenCVSections, []string{"patents"}) {
		t.Fatalf("legacy CV defaults = schema %d, visible %#v, hidden %#v", document.SchemaVersion, document.CVSections, document.HiddenCVSections)
	}
	document.CVSections = append(document.CVSections, "experience")
	if err := validateNamedSectionPartition(document.CVSections, document.HiddenCVSections, supportedCVSections, "CV"); err == nil {
		t.Fatal("duplicate CV section partition was accepted")
	}
}

func TestProfileRoundTripPreservesUnknownFieldsAndRejectsStaleRevision(t *testing.T) {
	root := newTempRepoFixture(t, fixtureSettingsJSON())
	profilePath := filepath.Join(root, "data", "profile.json")
	var source map[string]any
	if err := json.Unmarshal(readFixtureFile(t, profilePath), &source); err != nil {
		t.Fatal(err)
	}
	source["future_root"] = map[string]any{"keep": true}
	source["experience"].([]any)[0].(map[string]any)["future_item"] = []any{"keep", float64(7)}
	writeJSONFixture(t, profilePath, source)

	app := &App{repoRoot: root}
	loaded, err := app.LoadEditorData()
	if err != nil {
		t.Fatal(err)
	}
	var edited map[string]any
	if err := json.Unmarshal(loaded.Profile, &edited); err != nil {
		t.Fatal(err)
	}
	edited["identity"].(map[string]any)["affiliation_en"] = "Edited affiliation"
	profileInput, _ := json.Marshal(edited)
	request := saveRequestRevisions(loaded)
	request.Profile = profileInput
	request.SaveProfile = true
	saved, err := app.SaveEditorData(request)
	if err != nil {
		t.Fatal(err)
	}
	if saved.ProfileRevision == loaded.ProfileRevision {
		t.Fatal("profile revision did not change")
	}
	var stored map[string]any
	if err := json.Unmarshal(readFixtureFile(t, profilePath), &stored); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(stored["future_root"], source["future_root"]) ||
		!reflect.DeepEqual(stored["experience"].([]any)[0].(map[string]any)["future_item"], source["experience"].([]any)[0].(map[string]any)["future_item"]) {
		t.Fatal("profile unknown fields were not preserved")
	}

	stale := saved
	stored["identity"].(map[string]any)["affiliation_en"] = "External edit"
	writeJSONFixture(t, profilePath, stored)
	staleRequest := saveRequestRevisions(stale)
	staleRequest.Board = []BoardSaveItem{}
	staleRequest.SaveBoard = true
	if _, err := app.SaveEditorData(staleRequest); err == nil || !strings.Contains(err.Error(), "profile.json") {
		t.Fatalf("unrelated save with stale profile revision error = %v", err)
	}
}

func TestProfileRetiredIdentityRolesExcludedFromEditorAndSave(t *testing.T) {
	root := newTempRepoFixture(t, fixtureSettingsJSON())
	profilePath := filepath.Join(root, "data", "profile.json")
	fixture := readFixtureFile(t, profilePath)
	tests := []struct {
		name  string
		roles map[string]any
	}{
		{name: "omitted"},
		{name: "legacy strings", roles: map[string]any{"role_en": "Former role", "role_ko": "이전 역할"}},
		{name: "retired values do not require validation", roles: map[string]any{"role_en": map[string]any{"old": true}, "role_ko": nil}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var source map[string]any
			if err := json.Unmarshal(fixture, &source); err != nil {
				t.Fatal(err)
			}
			identity := source["identity"].(map[string]any)
			delete(identity, "role_en")
			delete(identity, "role_ko")
			for key, value := range test.roles {
				identity[key] = value
			}
			identity["future_identity"] = map[string]any{"role_en": "Preserve nested identity metadata"}
			source["future_root"] = map[string]any{"role_en": "Preserve other role", "role_ko": "다른 역할 보존"}
			input, err := json.Marshal(source)
			if err != nil {
				t.Fatal(err)
			}
			snapshot, err := readProfileBytes(input)
			if err != nil {
				t.Fatalf("profile without current identity roles could not load: %v", err)
			}
			editorRaw, err := profileForEditor(snapshot)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(snapshot.Raw, input) {
				t.Fatal("editor preparation changed the original revision bytes")
			}
			for key := range test.roles {
				if _, exists := snapshot.Data["identity"].(map[string]any)[key]; !exists {
					t.Fatalf("editor preparation removed %s from the source snapshot", key)
				}
			}
			assertCleanProfile := func(raw []byte) {
				t.Helper()
				var profile map[string]any
				if err := json.Unmarshal(raw, &profile); err != nil {
					t.Fatal(err)
				}
				gotIdentity := profile["identity"].(map[string]any)
				for _, key := range []string{"role_en", "role_ko"} {
					if _, exists := gotIdentity[key]; exists {
						t.Fatalf("retired identity %s reappeared", key)
					}
				}
				if !reflect.DeepEqual(gotIdentity["future_identity"], identity["future_identity"]) ||
					!reflect.DeepEqual(profile["future_root"], source["future_root"]) {
					t.Fatal("removing identity roles changed unrelated role metadata")
				}
				if !reflect.DeepEqual(profile["intro"], source["intro"]) {
					t.Fatal("removing identity roles changed the biography text")
				}
			}
			assertCleanProfile(editorRaw)
			// Direct input from an older client must also lose the retired fields.
			encoded, stored, err := buildProfileSave(input)
			if err != nil {
				t.Fatal(err)
			}
			assertCleanProfile(encoded)
			reopened, err := profileForEditor(stored)
			if err != nil {
				t.Fatal(err)
			}
			assertCleanProfile(reopened)
			delete(identity, "affiliation_en")
			invalid, err := json.Marshal(source)
			if err != nil {
				t.Fatal(err)
			}
			if _, _, err := buildProfileSave(invalid); err == nil || !strings.Contains(err.Error(), "affiliation_en") {
				t.Fatalf("unrelated identity validation was lost: %v", err)
			}
		})
	}
	if !bytes.Equal(readFixtureFile(t, profilePath), fixture) {
		t.Fatal("reading or normalizing a profile changed its source file")
	}
}

func TestProfileLinkedInDefaultsRoundTripsAndValidates(t *testing.T) {
	root := newTempRepoFixture(t, fixtureSettingsJSON())
	fixture := readFixtureFile(t, filepath.Join(root, "data", "profile.json"))
	snapshot, err := readProfileBytes(fixture)
	if err != nil {
		t.Fatal(err)
	}
	if _, exists := snapshot.Data["contact"].(map[string]any)["linkedin"]; exists {
		t.Fatal("legacy fixture unexpectedly contains linkedin")
	}

	editorRaw, err := profileForEditor(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	var editorProfile map[string]any
	if err := json.Unmarshal(editorRaw, &editorProfile); err != nil {
		t.Fatal(err)
	}
	editorContact := editorProfile["contact"].(map[string]any)
	if linkedin, ok := editorContact["linkedin"].(string); !ok || linkedin != "" {
		t.Fatalf("legacy linkedin default = %#v, want empty string", editorContact["linkedin"])
	}
	if _, exists := snapshot.Data["contact"].(map[string]any)["linkedin"]; exists {
		t.Fatal("pre-save editor materialization mutated the loaded profile snapshot")
	}

	const linkedinURL = "https://www.linkedin.com/in/hyeong-gon-jo-23a920409/"
	editorContact["linkedin"] = linkedinURL
	input, err := json.Marshal(editorProfile)
	if err != nil {
		t.Fatal(err)
	}
	encoded, stored, err := buildProfileSave(input)
	if err != nil {
		t.Fatal(err)
	}
	if got := stored.Data["contact"].(map[string]any)["linkedin"]; got != linkedinURL {
		t.Fatalf("stored linkedin = %#v, want %q", got, linkedinURL)
	}
	var encodedProfile map[string]any
	if err := json.Unmarshal(encoded, &encodedProfile); err != nil {
		t.Fatal(err)
	}
	if got := encodedProfile["contact"].(map[string]any)["linkedin"]; got != linkedinURL {
		t.Fatalf("encoded linkedin = %#v, want %q", got, linkedinURL)
	}

	editorContact["linkedin"] = map[string]any{"url": linkedinURL}
	invalid, err := json.Marshal(editorProfile)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := readProfileBytes(invalid); err == nil || !strings.Contains(err.Error(), "linkedin") {
		t.Fatalf("load error = %v, want linkedin string validation", err)
	}
	if _, _, err := buildProfileSave(invalid); err == nil || !strings.Contains(err.Error(), "linkedin") {
		t.Fatalf("save error = %v, want linkedin string validation", err)
	}
}

func TestOptionalEducationGPAValidationAndRoundTrip(t *testing.T) {
	root := newTempRepoFixture(t, fixtureSettingsJSON())
	fixture := readFixtureFile(t, filepath.Join(root, "data", "profile.json"))
	cases := []struct {
		name    string
		gpa     any
		omit    bool
		invalid bool
	}{
		{name: "omitted", omit: true},
		{name: "whole numbers", gpa: map[string]any{"value": 3, "max": 4}},
		{name: "decimal scale", gpa: map[string]any{"value": 3.96, "max": 4.3, "future_field": "keep"}},
		{name: "zero score", gpa: map[string]any{"value": 0, "max": 4}},
		{name: "perfect score", gpa: map[string]any{"value": 4.5, "max": 4.5}},
		{name: "null", gpa: nil, invalid: true},
		{name: "nonobject", gpa: "3.0 / 4.0", invalid: true},
		{name: "missing score", gpa: map[string]any{"max": 4}, invalid: true},
		{name: "missing maximum", gpa: map[string]any{"value": 3}, invalid: true},
		{name: "text score", gpa: map[string]any{"value": "3.0", "max": 4}, invalid: true},
		{name: "boolean maximum", gpa: map[string]any{"value": 3, "max": true}, invalid: true},
		{name: "negative score", gpa: map[string]any{"value": -0.1, "max": 4}, invalid: true},
		{name: "zero maximum", gpa: map[string]any{"value": 0, "max": 0}, invalid: true},
		{name: "above maximum", gpa: map[string]any{"value": 4.01, "max": 4}, invalid: true},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			var profile map[string]any
			if err := json.Unmarshal(fixture, &profile); err != nil {
				t.Fatal(err)
			}
			education := map[string]any{
				"degree_en": "Ph.D.", "degree_ko": "박사", "institution_en": "University", "institution_ko": "대학교",
				"period": "2022–2026", "advisor_en": "", "advisor_ko": "", "thesis_en": "Thesis", "thesis_ko": "졸업논문",
				"future_education_field": "keep",
			}
			if !test.omit {
				education["gpa"] = test.gpa
			}
			profile["education"] = []any{education}
			input, err := json.Marshal(profile)
			if err != nil {
				t.Fatal(err)
			}
			_, loadErr := readProfileBytes(input)
			encoded, _, saveErr := buildProfileSave(input)
			if test.invalid {
				if loadErr == nil || saveErr == nil || !strings.Contains(loadErr.Error(), "gpa") || !strings.Contains(saveErr.Error(), "gpa") {
					t.Fatalf("invalid GPA accepted: load=%v save=%v", loadErr, saveErr)
				}
				return
			}
			if loadErr != nil || saveErr != nil {
				t.Fatalf("valid optional GPA rejected: load=%v save=%v", loadErr, saveErr)
			}
			var result map[string]any
			if err := json.Unmarshal(encoded, &result); err != nil {
				t.Fatal(err)
			}
			stored := result["education"].([]any)[0].(map[string]any)
			if stored["future_education_field"] != "keep" {
				t.Fatal("saving GPA lost existing education metadata")
			}
			if test.omit {
				if _, exists := stored["gpa"]; exists {
					t.Fatal("optional GPA generated without input")
				}
			} else {
				var original map[string]any
				if err := json.Unmarshal(input, &original); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(stored["gpa"], original["education"].([]any)[0].(map[string]any)["gpa"]) {
					t.Fatal("GPA changed on save")
				}
				delete(stored, "gpa")
				withoutGPA, _ := json.Marshal(result)
				if _, _, err := buildProfileSave(withoutGPA); err != nil {
					t.Fatalf("removing optional GPA: %v", err)
				}
			}
		})
	}
}

func TestOptionalEducationThesisURLValidationAndRoundTrip(t *testing.T) {
	root := newTempRepoFixture(t, fixtureSettingsJSON())
	fixture := readFixtureFile(t, filepath.Join(root, "data", "profile.json"))
	cases := []struct {
		name    string
		link    any
		omit    bool
		invalid bool
	}{
		{name: "omitted", omit: true},
		{name: "cleared", link: ""},
		{name: "https", link: "https://example.org/thesis?id=123&lang=ko"},
		{name: "http", link: "http://example.org/thesis.pdf"},
		{name: "unicode title", link: "https://example.org/학위논문"},
		{name: "null", link: nil, invalid: true},
		{name: "number", link: 123, invalid: true},
		{name: "relative", link: "/thesis", invalid: true},
		{name: "javascript", link: "javascript:alert(1)", invalid: true},
		{name: "mailto", link: "mailto:test@example.org", invalid: true},
		{name: "missing host", link: "https://", invalid: true},
		{name: "control character", link: "https://example.org/\n", invalid: true},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			var profile map[string]any
			if err := json.Unmarshal(fixture, &profile); err != nil {
				t.Fatal(err)
			}
			education := map[string]any{
				"degree_en": "Ph.D.", "degree_ko": "박사", "institution_en": "University", "institution_ko": "대학교",
				"period": "2022–2026", "advisor_en": "", "advisor_ko": "", "thesis_en": "Thesis", "thesis_ko": "학위논문",
				"gpa": map[string]any{"value": 3, "max": 4}, "future_education_field": "keep", "visible_in_CV": true,
			}
			if !test.omit {
				education["thesis_url"] = test.link
			}
			profile["education"] = []any{education}
			input, err := json.Marshal(profile)
			if err != nil {
				t.Fatal(err)
			}
			_, loadErr := readProfileBytes(input)
			encoded, _, saveErr := buildProfileSave(input)
			if test.invalid {
				if loadErr == nil || saveErr == nil || !strings.Contains(loadErr.Error(), "thesis_url") || !strings.Contains(saveErr.Error(), "thesis_url") {
					t.Fatalf("invalid thesis URL accepted: load=%v save=%v", loadErr, saveErr)
				}
				return
			}
			if loadErr != nil || saveErr != nil {
				t.Fatalf("valid optional thesis URL rejected: load=%v save=%v", loadErr, saveErr)
			}
			var result map[string]any
			if err := json.Unmarshal(encoded, &result); err != nil {
				t.Fatal(err)
			}
			var original map[string]any
			if err := json.Unmarshal(input, &original); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(result["education"], original["education"]) {
				t.Fatal("saving optional thesis URL changed education metadata")
			}
		})
	}
}

func TestScholarshipAmountRoundTripsUnknownFieldsAndDropsLegacyDetails(t *testing.T) {
	root := newTempRepoFixture(t, fixtureSettingsJSON())
	fixture := readFixtureFile(t, filepath.Join(root, "data", "profile.json"))
	var profile map[string]any
	if err := json.Unmarshal(fixture, &profile); err != nil {
		t.Fatal(err)
	}
	scholarship := map[string]any{
		"name_en": "Research Scholarship", "name_ko": "연구 장학금",
		"period_en": "2025", "period_ko": "2025년",
		"summary_en": "Support", "summary_ko": "지원",
		"amount": map[string]any{
			"value": 13911000.5, "currency": "KRW", "future_amount_field": "keep",
		},
		"details":     []any{map[string]any{"label_en": "Legacy", "value_en": "drop"}},
		"future_item": map[string]any{"keep": true},
	}
	profile["scholarships"] = []any{scholarship}
	input, err := json.Marshal(profile)
	if err != nil {
		t.Fatal(err)
	}

	snapshot, err := readProfileBytes(input)
	if err != nil {
		t.Fatal(err)
	}
	loadedScholarship := snapshot.Data["scholarships"].([]any)[0].(map[string]any)
	if _, exists := loadedScholarship["details"]; !exists {
		t.Fatal("readProfileBytes unexpectedly mutated the exact loaded snapshot")
	}
	editorRaw, err := profileForEditor(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	var editorProfile map[string]any
	if err := json.Unmarshal(editorRaw, &editorProfile); err != nil {
		t.Fatal(err)
	}
	editorScholarship := editorProfile["scholarships"].([]any)[0].(map[string]any)
	if _, exists := editorScholarship["details"]; exists {
		t.Fatal("detached editor profile still exposes legacy scholarship details")
	}

	encoded, stored, err := buildProfileSave(input)
	if err != nil {
		t.Fatal(err)
	}
	storedScholarship := stored.Data["scholarships"].([]any)[0].(map[string]any)
	if _, exists := storedScholarship["details"]; exists {
		t.Fatal("newly saved scholarship regenerated legacy details")
	}
	if !reflect.DeepEqual(storedScholarship["future_item"], loadedScholarship["future_item"]) {
		t.Fatal("unknown scholarship item field was not preserved")
	}
	storedAmount := storedScholarship["amount"].(map[string]any)
	if storedAmount["currency"] != "KRW" || storedAmount["future_amount_field"] != "keep" {
		t.Fatalf("stored amount = %#v, want currency and unknown field preserved", storedAmount)
	}
	var encodedProfile map[string]any
	if err := json.Unmarshal(encoded, &encodedProfile); err != nil {
		t.Fatal(err)
	}
	encodedAmount := encodedProfile["scholarships"].([]any)[0].(map[string]any)["amount"].(map[string]any)
	if encodedAmount["value"] != 13911000.5 || encodedAmount["currency"] != "KRW" {
		t.Fatalf("encoded scholarship amount = %#v", encodedAmount)
	}
}

func TestScholarshipAmountValidation(t *testing.T) {
	root := newTempRepoFixture(t, fixtureSettingsJSON())
	fixture := readFixtureFile(t, filepath.Join(root, "data", "profile.json"))
	tests := []struct {
		name       string
		mutate     func(map[string]any)
		errorField string
	}{
		{name: "missing object", mutate: func(item map[string]any) { delete(item, "amount") }, errorField: "amount"},
		{name: "string value", mutate: func(item map[string]any) { item["amount"].(map[string]any)["value"] = "1000" }, errorField: "amount.value"},
		{name: "zero value", mutate: func(item map[string]any) { item["amount"].(map[string]any)["value"] = 0 }, errorField: "amount.value"},
		{name: "negative value", mutate: func(item map[string]any) { item["amount"].(map[string]any)["value"] = -0.5 }, errorField: "amount.value"},
		{name: "lowercase currency", mutate: func(item map[string]any) { item["amount"].(map[string]any)["currency"] = "krw" }, errorField: "amount.currency"},
		{name: "long currency", mutate: func(item map[string]any) { item["amount"].(map[string]any)["currency"] = "KRWX" }, errorField: "amount.currency"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var profile map[string]any
			if err := json.Unmarshal(fixture, &profile); err != nil {
				t.Fatal(err)
			}
			item := map[string]any{
				"name_en": "Scholarship", "name_ko": "장학금",
				"period_en": "2025", "period_ko": "2025년",
				"summary_en": "Support", "summary_ko": "지원",
				"amount": map[string]any{"value": 1000.25, "currency": "USD"},
			}
			test.mutate(item)
			profile["scholarships"] = []any{item}
			input, err := json.Marshal(profile)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := readProfileBytes(input); err == nil || !strings.Contains(err.Error(), test.errorField) {
				t.Fatalf("load error = %v, want %s validation", err, test.errorField)
			}
			if _, _, err := buildProfileSave(input); err == nil || !strings.Contains(err.Error(), test.errorField) {
				t.Fatalf("save error = %v, want %s validation", err, test.errorField)
			}
		})
	}
}

func TestProfileImageStagesPublishesAndRemovesOnlyOnSave(t *testing.T) {
	root := newTempRepoFixture(t, fixtureSettingsJSON())
	mediaDirectory := filepath.Join(root, "data", "media")
	originalImagePath := filepath.Join(mediaDirectory, "profile.jpg")
	writeProjectTestFile(t, originalImagePath, projectTestJPEGBytes())

	app := &App{repoRoot: root}
	loaded, err := app.LoadEditorData()
	if err != nil {
		t.Fatal(err)
	}
	if loaded.ProfileMedia.Src != "profile.jpg" || loaded.ProfileMedia.EditorKey == "" || loaded.ProfileMedia.PreviewURL == "" {
		t.Fatalf("loaded profile media summary = %#v, want existing preview and editor key", loaded.ProfileMedia)
	}

	replacementBytes := projectTestPNGBytes()
	replacementPath := filepath.Join(t.TempDir(), "replacement.png")
	writeProjectTestFile(t, replacementPath, replacementBytes)
	staged, err := app.StageBoardMedia([]string{replacementPath})
	if err != nil || len(staged.Items) != 1 {
		t.Fatalf("StageBoardMedia() = %#v, %v; want one profile image", staged, err)
	}
	if entries, err := os.ReadDir(mediaDirectory); err != nil || len(entries) != 1 {
		t.Fatalf("staging changed data/media before save: entries=%d err=%v", len(entries), err)
	}

	request := saveRequestRevisions(loaded)
	request.Profile = loaded.Profile
	request.SaveProfile = true
	request.ProfileMediaStageToken = staged.Items[0].StageToken
	saved, err := app.SaveEditorData(request)
	if err != nil {
		t.Fatal(err)
	}
	if saved.ProfileMedia.Src == "" || saved.ProfileMedia.Src == "profile.jpg" ||
		!strings.HasPrefix(saved.ProfileMedia.Src, "profile-") || !strings.HasSuffix(saved.ProfileMedia.Src, ".png") ||
		saved.ProfileMedia.PreviewURL == "" {
		t.Fatalf("saved profile media summary = %#v, want published replacement", saved.ProfileMedia)
	}
	publishedPath := filepath.Join(mediaDirectory, saved.ProfileMedia.Src)
	if published, err := os.ReadFile(publishedPath); err != nil || !bytes.Equal(published, replacementBytes) {
		t.Fatalf("published profile image differs from staged bytes: err=%v", err)
	}
	if _, err := os.Stat(originalImagePath); err != nil {
		t.Fatalf("replacing profile image deleted the existing user file: %v", err)
	}
	var stored map[string]any
	if err := json.Unmarshal(readFixtureFile(t, filepath.Join(root, "data", "profile.json")), &stored); err != nil {
		t.Fatal(err)
	}
	storedMedia := stored["profile_card"].(map[string]any)["media"].(map[string]any)
	if storedMedia["src"] != saved.ProfileMedia.Src || storedMedia["type"] != "image" {
		t.Fatalf("stored profile media = %#v, want replacement source and image type", storedMedia)
	}
	if _, exists := storedMedia["poster"]; exists {
		t.Fatal("image replacement retained an obsolete video poster")
	}

	removeRequest := saveRequestRevisions(saved)
	removeRequest.Profile = saved.Profile
	removeRequest.SaveProfile = true
	removeRequest.RemoveProfileMedia = true
	removed, err := app.SaveEditorData(removeRequest)
	if err != nil {
		t.Fatal(err)
	}
	if removed.ProfileMedia.Src != "" || removed.ProfileMedia.PreviewURL != "" {
		t.Fatalf("removed profile media summary = %#v, want empty reference", removed.ProfileMedia)
	}
	if _, err := os.Stat(publishedPath); err != nil {
		t.Fatalf("removing the profile reference deleted the user media file: %v", err)
	}
}

func TestProfileMediaActionRequiresProfileSave(t *testing.T) {
	root := newTempRepoFixture(t, fixtureSettingsJSON())
	app := &App{repoRoot: root}
	loaded, err := app.LoadEditorData()
	if err != nil {
		t.Fatal(err)
	}
	request := saveRequestRevisions(loaded)
	request.Board = []BoardSaveItem{}
	request.SaveBoard = true
	request.RemoveProfileMedia = true
	if _, err := app.SaveEditorData(request); err == nil || !strings.Contains(err.Error(), "profile.json") {
		t.Fatalf("profile media action without profile save error = %v", err)
	}
}

func TestProfileRowsRequireMeaningfulPrimaryFieldsOnLoadAndSave(t *testing.T) {
	root := newTempRepoFixture(t, fixtureSettingsJSON())
	fixture := readFixtureFile(t, filepath.Join(root, "data", "profile.json"))
	tests := []struct {
		section      string
		primaryField string
		validField   string
		item         map[string]any
	}{
		{
			section:      "experience",
			primaryField: "title_en",
			validField:   "title_ko",
			item: map[string]any{
				"title_en": " \t", "title_ko": "\n", "institution_en": "Lab", "institution_ko": "연구실", "period_en": "2025", "period_ko": "2025년",
			},
		},
		{
			section:      "education",
			primaryField: "degree_en",
			validField:   "degree_ko",
			item: map[string]any{
				"degree_en": " ", "degree_ko": "\r\n", "institution_en": "University", "institution_ko": "대학교", "period": "2025", "advisor_en": "Advisor", "advisor_ko": "지도교수", "thesis_en": "Thesis", "thesis_ko": "논문",
			},
		},
		{
			section:      "teaching",
			primaryField: "title_en",
			validField:   "title_ko",
			item: map[string]any{
				"title_en": "", "title_ko": "  ", "detail_en": "Secondary detail does not qualify", "detail_ko": "상세 내용", "period": "2025",
			},
		},
		{
			section:      "scholarships",
			primaryField: "name_en",
			validField:   "name_ko",
			item: map[string]any{
				"name_en": "\t", "name_ko": " ", "period_en": "2025", "period_ko": "2025년", "summary_en": "Summary", "summary_ko": "요약",
				"amount": map[string]any{"value": 1000, "currency": "KRW"},
			},
		},
		{
			section:      "certifications",
			primaryField: "name_en",
			validField:   "name_ko",
			item: map[string]any{
				"name_en": "\n", "name_ko": " ", "issuer_en": "Issuer", "issuer_ko": "발급 기관", "date": "2025-01-01",
			},
		},
		{
			section:      "skills",
			primaryField: "name",
			validField:   "name",
			item: map[string]any{
				"name": " \t\r\n", "detail_en": "A populated detail does not qualify", "detail_ko": "상세 내용",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.section, func(t *testing.T) {
			var profile map[string]any
			if err := json.Unmarshal(fixture, &profile); err != nil {
				t.Fatal(err)
			}
			profile[test.section] = []any{test.item}
			input, err := json.Marshal(profile)
			if err != nil {
				t.Fatal(err)
			}

			if _, err := readProfileBytes(input); err == nil || !strings.Contains(err.Error(), test.primaryField) {
				t.Fatalf("load error = %v, want meaningful %s validation", err, test.primaryField)
			}
			if _, _, err := buildProfileSave(input); err == nil || !strings.Contains(err.Error(), test.primaryField) {
				t.Fatalf("save error = %v, want meaningful %s validation", err, test.primaryField)
			}

			test.item[test.validField] = " Meaningful value "
			validInput, err := json.Marshal(profile)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := readProfileBytes(validInput); err != nil {
				t.Fatalf("one meaningful primary field was rejected: %v", err)
			}
		})
	}
}

func TestProfileCredentialRowsRequireMeaningfulNames(t *testing.T) {
	root := newTempRepoFixture(t, fixtureSettingsJSON())
	fixture := readFixtureFile(t, filepath.Join(root, "data", "profile.json"))

	t.Run("credential name", func(t *testing.T) {
		var profile map[string]any
		if err := json.Unmarshal(fixture, &profile); err != nil {
			t.Fatal(err)
		}
		card := profile["profile_card"].(map[string]any)
		credential := map[string]any{"name_en": " \t", "name_ko": "\n"}
		card["credentials"] = []any{credential}
		input, err := json.Marshal(profile)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := readProfileBytes(input); err == nil || !strings.Contains(err.Error(), "name_en") {
			t.Fatalf("load error = %v, want meaningful credential name validation", err)
		}
		if _, _, err := buildProfileSave(input); err == nil || !strings.Contains(err.Error(), "name_en") {
			t.Fatalf("save error = %v, want meaningful credential name validation", err)
		}

		credential["name_ko"] = " 자격 "
		validInput, err := json.Marshal(profile)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := readProfileBytes(validInput); err != nil {
			t.Fatalf("localized credential name was rejected: %v", err)
		}
	})

}

func TestVisibleInCVDefaultsValidationAndMultiFileSave(t *testing.T) {
	missing := []byte(`[{"id":"award","date":"2025-01-01","title_en":"Award","title_ko":"수상","organization_en":"Org","organization_ko":"기관"}]`)
	snapshot, err := readAwardsBytes(missing)
	if err != nil {
		t.Fatal(err)
	}
	if !awardItemSummaries(snapshot)[0].VisibleInCV {
		t.Fatal("missing visible_in_CV did not default true")
	}
	invalid := bytes.Replace(missing, []byte(`"organization_ko":"기관"`), []byte(`"organization_ko":"기관","visible_in_CV":"yes"`), 1)
	if _, err := readAwardsBytes(invalid); err == nil || !strings.Contains(err.Error(), "visible_in_CV") {
		t.Fatalf("non-Boolean visible_in_CV error = %v", err)
	}

	root := newTempRepoFixture(t, fixtureSettingsJSON())
	if err := os.MkdirAll(filepath.Join(root, "data", "media"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFixtureFile(t, filepath.Join(root, "data", "awards.json"), append(missing, '\n'))
	writeFixtureFile(t, filepath.Join(root, "data", "software.json"), []byte(`[{"id":"tool","name":"Tool","stage":"release","links":[],"notes_en":["Note"],"notes_kr":["설명"],"media":[],"technologies":["Go"]}]
`))
	app := &App{repoRoot: root}
	loaded, err := app.LoadEditorData()
	if err != nil {
		t.Fatal(err)
	}
	visible := false
	request := saveRequestRevisions(loaded)
	request.Projects = []ProjectSaveItem{{EditorKey: loaded.Projects[0].EditorKey, VisibleInCV: &visible}}
	request.SaveProjects = true
	request.Software = []SoftwareSaveItem{{EditorKey: loaded.Software[0].EditorKey, VisibleInCV: &visible}}
	request.SaveSoftware = true
	request.Awards = []AwardSaveItem{{EditorKey: loaded.Awards[0].EditorKey, VisibleInCV: &visible}}
	request.SaveAwards = true
	request.Publications = []PublicationSaveItem{{EditorKey: loaded.Publications[0].EditorKey, VisibleInCV: &visible}}
	request.SavePublications = true
	saved, err := app.SaveEditorData(request)
	if err != nil {
		t.Fatal(err)
	}
	if saved.Projects[0].VisibleInCV || saved.Software[0].VisibleInCV || saved.Awards[0].VisibleInCV || saved.Publications[0].VisibleInCV {
		t.Fatal("multi-file visible_in_CV false did not round-trip")
	}
	for _, section := range []string{"projects", "software", "awards", "publications"} {
		if containsString(saved.Settings.CVSections, section) || !containsString(saved.Settings.HiddenCVSections, section) {
			t.Fatalf("CV section %s with no selected items was not auto-hidden: visible %#v hidden %#v", section, saved.Settings.CVSections, saved.Settings.HiddenCVSections)
		}
	}
}

func TestSavingCannotExposeEmptyCVSection(t *testing.T) {
	root := newTempRepoFixture(t, fixtureSettingsJSON())
	app := &App{repoRoot: root}
	loaded, err := app.LoadEditorData()
	if err != nil {
		t.Fatal(err)
	}
	if !containsString(loaded.Settings.CVSections, "teaching") {
		t.Fatalf("legacy fixture teaching section should initially be visible: %#v", loaded.Settings.CVSections)
	}

	request := saveRequestRevisions(loaded)
	request.Settings = loaded.Settings
	request.SaveSettings = true
	saved, err := app.SaveEditorData(request)
	if err != nil {
		t.Fatal(err)
	}
	if containsString(saved.Settings.CVSections, "teaching") || !containsString(saved.Settings.HiddenCVSections, "teaching") {
		t.Fatalf("empty teaching section was exposed after save: visible %#v hidden %#v", saved.Settings.CVSections, saved.Settings.HiddenCVSections)
	}
	if !containsString(saved.Settings.CVSections, "experience") {
		t.Fatalf("non-empty experience section was unexpectedly hidden: visible %#v hidden %#v", saved.Settings.CVSections, saved.Settings.HiddenCVSections)
	}
}

func TestSavingLastProfileAndAwardItemsAutoHidesMainSections(t *testing.T) {
	root := newTempRepoFixture(t, fixtureSettingsJSON())
	writeFixtureFile(t, filepath.Join(root, "data", "awards.json"), []byte(`[{"id":"award","date":"2025-01-01","title_en":"Award","title_ko":"수상","organization_en":"Org","organization_ko":"기관"}]
`))
	app := &App{repoRoot: root}
	loaded, err := app.LoadEditorData()
	if err != nil {
		t.Fatal(err)
	}
	var profile map[string]any
	if err := json.Unmarshal(loaded.Profile, &profile); err != nil {
		t.Fatal(err)
	}
	profile["experience"] = []any{}
	profileInput, _ := json.Marshal(profile)
	request := saveRequestRevisions(loaded)
	request.Profile = profileInput
	request.SaveProfile = true
	request.Awards = []AwardSaveItem{}
	request.SaveAwards = true
	saved, err := app.SaveEditorData(request)
	if err != nil {
		t.Fatal(err)
	}
	for _, section := range []string{"experience", "awards"} {
		if containsString(saved.Settings.MainPageSections, section) || !containsString(saved.Settings.HiddenMainPageSections, section) {
			t.Fatalf("empty %s section was not auto-hidden: visible %#v hidden %#v", section, saved.Settings.MainPageSections, saved.Settings.HiddenMainPageSections)
		}
	}
}

func saveRequestRevisions(loaded EditorDataResponse) SaveEditorDataRequest {
	return SaveEditorDataRequest{
		Settings: loaded.Settings, SettingsRevision: loaded.SettingsRevision,
		ProfileRevision: loaded.ProfileRevision, PeopleRevision: loaded.PeopleRevision,
		AwardsRevision: loaded.AwardsRevision, AcademicActivitiesRevision: loaded.AcademicActivitiesRevision,
		PublicationsRevision: loaded.PublicationsRevision,
		ProjectsRevision:     loaded.ProjectsRevision, SoftwareRevision: loaded.SoftwareRevision,
		BoardRevision: loaded.BoardRevision,
	}
}

func writeJSONFixture(t *testing.T, path string, value any) {
	t.Helper()
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	writeFixtureFile(t, path, append(raw, '\n'))
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
