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

func TestLoadEditorDataAcceptsRepositoryDataWithoutWriting(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatalf("resolve repository root: %v", err)
	}
	if !isRepoRoot(root) {
		t.Fatalf("%s is not the repository root", root)
	}

	filenames := []string{
		"settings.json", "people.json", "awards.json", "publications.json",
		"projects.json", "software.json", "board.json", "academic_activities.json",
	}
	before := make(map[string][]byte, len(filenames))
	for _, filename := range filenames {
		before[filename], err = os.ReadFile(filepath.Join(root, "data", filename))
		if err != nil {
			t.Fatalf("read %s before load: %v", filename, err)
		}
	}

	response, err := (&App{repoRoot: root}).LoadEditorData()
	if err != nil {
		t.Fatalf("LoadEditorData() rejected repository data: %v", err)
	}
	if len(response.People) == 0 || len(response.Publications) == 0 {
		t.Fatalf("LoadEditorData() returned people=%d publications=%d", len(response.People), len(response.Publications))
	}
	for _, filename := range filenames {
		after, readErr := os.ReadFile(filepath.Join(root, "data", filename))
		if readErr != nil {
			t.Fatalf("read %s after load: %v", filename, readErr)
		}
		if !bytes.Equal(after, before[filename]) {
			t.Errorf("LoadEditorData() changed data/%s", filename)
		}
	}
}

func TestValidateSectionPartition(t *testing.T) {
	tests := []struct {
		name    string
		visible []string
		hidden  []string
		wantErr bool
	}{
		{
			name:    "valid partition",
			visible: []string{"experience", "education", "skills", "academic_activities"},
			hidden:  []string{"scholarships", "certifications", "awards", "patents", "teaching"},
		},
		{
			name:    "missing section",
			visible: []string{"experience", "academic_activities"},
			hidden:  []string{"education", "scholarships", "certifications", "awards", "patents", "teaching"},
			wantErr: true,
		},
		{
			name:    "duplicate section",
			visible: []string{"experience", "education", "skills", "academic_activities"},
			hidden:  []string{"experience", "scholarships", "certifications", "awards", "patents", "teaching"},
			wantErr: true,
		},
		{
			name:    "unknown section",
			visible: []string{"experience", "education", "skills", "academic_activities"},
			hidden:  []string{"scholarships", "certifications", "awards", "patents", "unknown"},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateSectionPartition(test.visible, test.hidden)
			if test.wantErr && err == nil {
				t.Fatal("validateSectionPartition() error = nil, want error")
			}
			if !test.wantErr && err != nil {
				t.Fatalf("validateSectionPartition() unexpected error: %v", err)
			}
		})
	}
}

func TestNormaliseTaxonomyPreservesIDsAndAssignsCollisionSafeSlug(t *testing.T) {
	current := []TaxonomyItem{
		{
			ID: "hvac-controls",
			BilingualLabel: BilingualLabel{
				LabelEN: "HVAC",
				LabelKO: "HVAC Korean",
			},
		},
	}
	input := []TaxonomyItem{
		{
			ID: "hvac-controls",
			BilingualLabel: BilingualLabel{
				LabelEN: " Heating and Cooling ",
				LabelKO: " Heating Korean ",
			},
		},
		{
			BilingualLabel: BilingualLabel{
				LabelEN: "HVAC Controls",
				LabelKO: "Controls Korean",
			},
		},
	}

	got, err := normaliseTaxonomy(input, current, map[string]int{}, "taxonomy")
	if err != nil {
		t.Fatalf("normaliseTaxonomy() unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("normaliseTaxonomy() returned %d items, want 2", len(got))
	}
	if got[0].ID != "hvac-controls" {
		t.Errorf("existing ID = %q, want %q", got[0].ID, "hvac-controls")
	}
	if got[0].LabelEN != "Heating and Cooling" || got[0].LabelKO != "Heating Korean" {
		t.Errorf("existing labels were not trimmed: %#v", got[0].BilingualLabel)
	}
	if got[1].ID != "hvac-controls-2" {
		t.Errorf("new collision-safe ID = %q, want %q", got[1].ID, "hvac-controls-2")
	}
}

func TestNormaliseTaxonomyBlocksReferencedDeletion(t *testing.T) {
	current := []TaxonomyItem{
		{
			ID: "retrofit",
			BilingualLabel: BilingualLabel{
				LabelEN: "Retrofit",
				LabelKO: "Retrofit Korean",
			},
		},
	}

	_, err := normaliseTaxonomy(nil, current, map[string]int{"retrofit": 3}, "project theme")
	if err == nil {
		t.Fatal("normaliseTaxonomy() error = nil, want referenced deletion error")
	}
}

func TestNormaliseTaxonomyDistinguishesOmittedAndClearedColors(t *testing.T) {
	current := []TaxonomyItem{{
		ID:             "hvac-controls",
		Color:          "#aB123F",
		BilingualLabel: BilingualLabel{LabelEN: "HVAC", LabelKO: "HVAC Korean"},
	}}
	tests := []struct {
		name      string
		colorJSON string
		want      string
	}{
		{name: "omitted preserves existing", want: "#aB123F"},
		{name: "explicit empty restores default", colorJSON: `,"color":""`, want: ""},
		{name: "edited color is trimmed", colorJSON: `,"color":" #123abc "`, want: "#123abc"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var input []TaxonomyItem
			if err := json.Unmarshal([]byte(`[{"id":"hvac-controls","label_en":"HVAC","label_ko":"HVAC Korean"`+test.colorJSON+`}]`), &input); err != nil {
				t.Fatalf("decode API input: %v", err)
			}
			got, err := normaliseTaxonomy(input, current, map[string]int{}, "taxonomy")
			if err != nil {
				t.Fatalf("normaliseTaxonomy() unexpected error: %v", err)
			}
			if got[0].Color != test.want {
				t.Errorf("color = %q, want %q", got[0].Color, test.want)
			}
			if current[0].Color != "#aB123F" {
				t.Fatal("normaliseTaxonomy() changed the current settings")
			}
		})
	}
}

func TestTaxonomyRejectsInvalidColorsOnLoadAndSave(t *testing.T) {
	for _, color := range []string{"#abc", "#12345g", "rgb(1,2,3)", "red", "#11223344", "123456"} {
		t.Run(color, func(t *testing.T) {
			for _, taxonomy := range []string{"project", "publication"} {
				document := validSettingsDocument()
				if err := normaliseLoadedSettings(&document); err != nil {
					t.Fatalf("prepare valid settings: %v", err)
				}
				if taxonomy == "project" {
					document.Topics[0].Color = color
				} else {
					document.Topics[1].Color = color
				}
				if err := normaliseLoadedSettings(&document); err == nil {
					t.Errorf("normaliseLoadedSettings() accepted invalid %s color %q", taxonomy, color)
				}
				if _, err := normaliseForSave(document, validSettingsDocument(), SettingsUsage{}); err == nil || !strings.Contains(err.Error(), "#RRGGBB") {
					t.Errorf("normaliseForSave() error = %v, want invalid %s color %q error", err, taxonomy, color)
				}
			}
		})
	}
}

func TestAppSaveSettingsPersistsAndClearsTaxonomyColors(t *testing.T) {
	repoRoot := newTempRepoFixture(t, fixtureSettingsJSON())
	app := &App{repoRoot: repoRoot}
	loaded, err := app.LoadSettings()
	if err != nil {
		t.Fatalf("LoadSettings(): %v", err)
	}
	draft := loaded.Settings
	draft.Topics[0].Color = "#123aBc"
	draft.Topics[1].Color = "#aBc456"
	draft.Topics = append(draft.Topics, TaxonomyItem{
		Color:          "#987654",
		BilingualLabel: BilingualLabel{LabelEN: "Indoor Air Quality", LabelKO: "IAQ Korean"},
	})
	saved, err := app.SaveSettings(SaveSettingsRequest{Settings: draft, Revision: loaded.Revision})
	if err != nil {
		t.Fatalf("SaveSettings() with new colors: %v", err)
	}
	reopened, err := (&App{repoRoot: repoRoot}).LoadSettings()
	if err != nil {
		t.Fatalf("reopen settings with colors: %v", err)
	}
	if reopened.Settings.Topics[0].Color != "#123aBc" || reopened.Settings.Topics[1].Color != "#aBc456" {
		t.Fatalf("saved colors were not reloaded: project %q, publication %q", reopened.Settings.Topics[0].Color, reopened.Settings.Topics[1].Color)
	}
	if len(reopened.Settings.Topics) != 3 || reopened.Settings.Topics[2].ID != "indoor-air-quality" || reopened.Settings.Topics[2].Color != "#987654" {
		t.Fatalf("new project theme did not retain its generated ID and color: %#v", reopened.Settings.Topics)
	}

	// Decode the next request as Wails does: omission preserves a color, explicit empty clears it.
	encoded, err := json.Marshal(SaveSettingsRequest{Settings: saved.Settings, Revision: saved.Revision})
	if err != nil {
		t.Fatalf("encode next API request: %v", err)
	}
	var requestShape map[string]any
	if err := json.Unmarshal(encoded, &requestShape); err != nil {
		t.Fatalf("decode next API request shape: %v", err)
	}
	settingsShape := requestShape["settings"].(map[string]any)
	delete(settingsShape["topics"].([]any)[0].(map[string]any), "color")
	settingsShape["topics"].([]any)[1].(map[string]any)["color"] = ""
	encoded, err = json.Marshal(requestShape)
	if err != nil {
		t.Fatalf("encode request with omitted and cleared colors: %v", err)
	}
	var request SaveSettingsRequest
	if err := json.Unmarshal(encoded, &request); err != nil {
		t.Fatalf("decode request with omitted and cleared colors: %v", err)
	}
	if _, err := app.SaveSettings(request); err != nil {
		t.Fatalf("SaveSettings() with cleared publication color: %v", err)
	}
	written, raw, err := readSettings(filepath.Join(repoRoot, "data", "settings.json"))
	if err != nil {
		t.Fatalf("read settings after color reset: %v", err)
	}
	if written.Topics[0].Color != "#123aBc" || written.Topics[1].Color != "" {
		t.Fatalf("color preservation/reset failed: project %q, publication %q", written.Topics[0].Color, written.Topics[1].Color)
	}
	if written.Topics[0].ID != "retrofit" || written.Topics[1].ID != "hvac" {
		t.Fatal("editing colors changed existing taxonomy IDs")
	}
	var diskShape map[string]any
	if err := json.Unmarshal(raw, &diskShape); err != nil {
		t.Fatalf("decode written settings shape: %v", err)
	}
	if _, exists := diskShape["topics"].([]any)[1].(map[string]any)["color"]; exists {
		t.Fatal("cleared color should be absent on disk so the automatic palette applies")
	}
}

func TestNormaliseLoadedSettingsRejectsInvalidExistingIDsWithoutGeneratingOne(t *testing.T) {
	tests := []struct {
		name      string
		projectID string
		topicID   string
	}{
		{name: "blank project ID", projectID: "", topicID: "hvac"},
		{name: "whitespace-only project ID", projectID: "   ", topicID: "hvac"},
		{name: "whitespace project ID", projectID: " retrofit", topicID: "hvac"},
		{name: "noncanonical project ID", projectID: "Retrofit", topicID: "hvac"},
		{name: "double hyphen project ID", projectID: "retrofit--deep", topicID: "hvac"},
		{name: "blank publication ID", projectID: "retrofit", topicID: ""},
		{name: "all topics sentinel", projectID: "retrofit", topicID: "__all_topics__"},
		{name: "fallback topic sentinel", projectID: "retrofit", topicID: "__fallback_topic__"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			document := validSettingsDocument()
			document.Topics[0].ID = test.projectID
			document.Topics[1].ID = test.topicID

			err := normaliseLoadedSettings(&document)
			if err == nil {
				t.Fatal("normaliseLoadedSettings() error = nil, want invalid ID error")
			}
			if document.Topics[0].ID != test.projectID || document.Topics[1].ID != test.topicID {
				t.Fatalf("normaliseLoadedSettings() changed IDs after rejection: project %q, topic %q", document.Topics[0].ID, document.Topics[1].ID)
			}
		})
	}
}

func TestNormaliseLoadedSettingsRejectsLabelSurroundingWhitespace(t *testing.T) {
	document := validSettingsDocument()
	document.Topics[0].LabelEN = " Retrofit "

	err := normaliseLoadedSettings(&document)
	if err == nil {
		t.Fatal("normaliseLoadedSettings() error = nil, want surrounding whitespace error")
	}
	if document.Topics[0].LabelEN != " Retrofit " {
		t.Fatalf("normaliseLoadedSettings() silently changed label to %q", document.Topics[0].LabelEN)
	}
}

func TestValidateReferences(t *testing.T) {
	document := SettingsDocument{
		Topics: []TaxonomyItem{{ID: "retrofit"}, {ID: "daylight"}},
	}

	t.Run("valid references", func(t *testing.T) {
		err := validateReferences(document, SettingsUsage{
			ProjectThemes:     map[string]int{"retrofit": 2},
			PublicationTopics: map[string]int{"daylight": 4, "": 1},
		})
		if err != nil {
			t.Fatalf("validateReferences() unexpected error: %v", err)
		}
	})

	t.Run("unknown project theme", func(t *testing.T) {
		err := validateReferences(document, SettingsUsage{
			ProjectThemes:     map[string]int{"missing-theme": 1},
			PublicationTopics: map[string]int{},
		})
		if err == nil {
			t.Fatal("validateReferences() error = nil, want unknown project theme error")
		}
		if !strings.Contains(err.Error(), "missing-theme") {
			t.Fatalf("validateReferences() error %q does not identify missing theme", err)
		}
	})

	t.Run("unknown publication topic", func(t *testing.T) {
		err := validateReferences(document, SettingsUsage{
			ProjectThemes:     map[string]int{},
			PublicationTopics: map[string]int{"missing-topic": 1},
		})
		if err == nil {
			t.Fatal("validateReferences() error = nil, want unknown publication topic error")
		}
		if !strings.Contains(err.Error(), "missing-topic") {
			t.Fatalf("validateReferences() error %q does not identify missing topic", err)
		}
	})
}

func TestWriteFileAtomicallyReplacesContentsAndCleansTemporaryFile(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "settings.json")
	if err := os.WriteFile(path, []byte("old\n"), 0o640); err != nil {
		t.Fatalf("seed settings file: %v", err)
	}

	want := []byte("{\n  \"schema_version\": 4\n}\n")
	if err := writeFileAtomically(path, want); err != nil {
		t.Fatalf("writeFileAtomically() unexpected error: %v", err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read replaced settings file: %v", err)
	}
	if string(got) != string(want) {
		t.Fatalf("replaced contents = %q, want %q", got, want)
	}
	temporaryFiles, err := filepath.Glob(filepath.Join(directory, ".settings-*.tmp"))
	if err != nil {
		t.Fatalf("find temporary files: %v", err)
	}
	if len(temporaryFiles) != 0 {
		t.Fatalf("temporary files were not cleaned up: %v", temporaryFiles)
	}
}

func TestAppLoadSettingsDoesNotModifySettingsFile(t *testing.T) {
	settingsBytes := fixtureSettingsJSON()
	repoRoot := newTempRepoFixture(t, settingsBytes)
	settingsPath := filepath.Join(repoRoot, "data", "settings.json")
	app := &App{repoRoot: repoRoot, dirty: true}

	response, err := app.LoadSettings()
	if err != nil {
		t.Fatalf("LoadSettings() unexpected error: %v", err)
	}
	after, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("read settings after LoadSettings(): %v", err)
	}
	if string(after) != string(settingsBytes) {
		t.Fatalf("LoadSettings() changed settings.json\ngot:  %q\nwant: %q", after, settingsBytes)
	}
	if response.Revision != revisionOf(settingsBytes) {
		t.Errorf("LoadSettings() revision = %q, want %q", response.Revision, revisionOf(settingsBytes))
	}
	if response.Settings.SchemaVersion != settingsSchemaVersion {
		t.Errorf("LoadSettings() response schema = %d, want normalized schema %d", response.Settings.SchemaVersion, settingsSchemaVersion)
	}
	if !reflect.DeepEqual(response.Settings.HiddenMainPageSections, []string{"academic_activities", "patents"}) {
		t.Errorf("LoadSettings() hidden sections = %#v, want legacy academic activities and patents hidden", response.Settings.HiddenMainPageSections)
	}
	if app.dirty {
		t.Error("LoadSettings() left app dirty, want clean")
	}
}

func TestAppSaveSettingsRequiresRevisionAndWritesNormalizedSettings(t *testing.T) {
	settingsBytes := fixtureSettingsJSON()
	repoRoot := newTempRepoFixture(t, settingsBytes)
	settingsPath := filepath.Join(repoRoot, "data", "settings.json")
	app := &App{repoRoot: repoRoot}

	loaded, err := app.LoadSettings()
	if err != nil {
		t.Fatalf("LoadSettings() unexpected error: %v", err)
	}
	draft := loaded.Settings
	draft.MainPageSections = []string{
		"education",
		"experience",
		"scholarships",
		"certifications",
		"awards",
		"teaching",
		"academic_activities",
	}
	draft.HiddenMainPageSections = []string{"skills", "patents"}
	draft.Topics[0].LabelEN = " Retrofit Updated "
	draft.Topics[0].LabelKO = " Retrofit Korean Updated "
	draft.Topics = append(draft.Topics, TaxonomyItem{
		BilingualLabel: BilingualLabel{
			LabelEN: " Indoor Air Quality ",
			LabelKO: " IAQ Korean ",
		},
	})

	_, err = app.SaveSettings(SaveSettingsRequest{
		Settings: draft,
		Revision: "stale-revision",
	})
	if err == nil {
		t.Fatal("SaveSettings() with stale revision error = nil, want conflict error")
	}
	afterRejectedSave, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("read settings after rejected SaveSettings(): %v", err)
	}
	if string(afterRejectedSave) != string(settingsBytes) {
		t.Fatalf("rejected SaveSettings() changed settings.json\ngot:  %q\nwant: %q", afterRejectedSave, settingsBytes)
	}

	app.SetDirty(true)
	saved, err := app.SaveSettings(SaveSettingsRequest{
		Settings: draft,
		Revision: loaded.Revision,
	})
	if err != nil {
		t.Fatalf("SaveSettings() with current revision unexpected error: %v", err)
	}
	written, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("read settings after successful SaveSettings(): %v", err)
	}
	if string(written) == string(settingsBytes) {
		t.Fatal("successful SaveSettings() did not update settings.json")
	}
	if saved.Revision != revisionOf(written) {
		t.Errorf("SaveSettings() revision = %q, want hash of written bytes %q", saved.Revision, revisionOf(written))
	}
	if app.dirty {
		t.Error("SaveSettings() left app dirty, want clean")
	}

	writtenDocument, raw, err := readSettings(settingsPath)
	if err != nil {
		t.Fatalf("readSettings() after successful save: %v", err)
	}
	if string(raw) != string(written) {
		t.Fatal("readSettings() raw bytes differ from settings.json bytes")
	}
	if writtenDocument.SchemaVersion != settingsSchemaVersion {
		t.Errorf("written schema = %d, want %d", writtenDocument.SchemaVersion, settingsSchemaVersion)
	}
	if !reflect.DeepEqual(writtenDocument.HiddenMainPageSections, []string{"skills", "patents"}) {
		t.Errorf("written hidden sections = %#v, want [skills patents]", writtenDocument.HiddenMainPageSections)
	}
	if writtenDocument.Topics[0].ID != "retrofit" {
		t.Errorf("existing project theme ID = %q, want preserved ID retrofit", writtenDocument.Topics[0].ID)
	}
	if writtenDocument.Topics[0].LabelEN != "Retrofit Updated" || writtenDocument.Topics[0].LabelKO != "Retrofit Korean Updated" {
		t.Errorf("existing project theme labels were not normalized: %#v", writtenDocument.Topics[0])
	}
	if len(writtenDocument.Topics) != 3 || writtenDocument.Topics[2].ID != "indoor-air-quality" {
		t.Errorf("new project theme = %#v, want normalized slug indoor-air-quality", writtenDocument.Topics)
	}
	if writtenDocument.Topics[2].LabelEN != "Indoor Air Quality" || writtenDocument.Topics[2].LabelKO != "IAQ Korean" {
		t.Errorf("new project theme labels were not normalized: %#v", writtenDocument.Topics[1])
	}
}

func TestAppSaveSettingsPreservesUnknownSettingsFieldsSemantically(t *testing.T) {
	settingsBytes := []byte(`{
  "schema_version": 4,
  "main_page_sections": ["experience", "education", "scholarships", "certifications", "awards", "teaching", "skills"],
  "hidden_main_page_sections": [],
  "project_themes": [
    {
      "id": "retrofit",
      "label_en": "Retrofit",
      "label_ko": "Retrofit Korean",
      "color": "#12aBcD",
      "future_row_setting": {"enabled": true, "levels": [1, 2, 3]}
    }
  ],
  "project_theme_fallback": {
    "label_en": "Other",
    "label_ko": "Other Korean",
    "future_fallback_setting": ["alpha", {"beta": 2}]
  },
  "publication_topics": [
    {
      "id": "hvac",
      "label_en": "HVAC",
      "label_ko": "HVAC Korean",
      "color": "#cba321",
      "future_topic_setting": null
    }
  ],
  "publication_topic_fallback": {
    "label_en": "Other",
    "label_ko": "Other Korean",
    "future_topic_fallback_setting": false
  },
  "future_top_level_setting": {
    "mode": "manual",
    "threshold": 1.25
  }
}
`)
	repoRoot := newTempRepoFixture(t, settingsBytes)
	settingsPath := filepath.Join(repoRoot, "data", "settings.json")
	app := &App{repoRoot: repoRoot}

	loaded, err := app.LoadSettings()
	if err != nil {
		t.Fatalf("LoadSettings() unexpected error: %v", err)
	}
	apiEncoded, err := json.Marshal(loaded.Settings)
	if err != nil {
		t.Fatalf("marshal Wails settings response shape: %v", err)
	}
	var apiShape map[string]any
	if err := json.Unmarshal(apiEncoded, &apiShape); err != nil {
		t.Fatalf("decode Wails settings response shape: %v", err)
	}
	if _, exists := apiShape["future_top_level_setting"]; exists {
		t.Fatal("unknown file-only top-level field leaked into the Wails API")
	}
	apiProjectRow := apiShape["topics"].([]any)[0].(map[string]any)
	if apiProjectRow["color"] != "#12aBcD" {
		t.Fatalf("color was not exposed in the Wails API: %#v", apiProjectRow["color"])
	}
	if _, exists := apiProjectRow["future_row_setting"]; exists {
		t.Fatal("unknown file-only taxonomy field leaked into the Wails API")
	}

	draft := loaded.Settings
	draft.Topics[0].LabelEN = "Retrofit Updated"
	if _, err := app.SaveSettings(SaveSettingsRequest{Settings: draft, Revision: loaded.Revision}); err != nil {
		t.Fatalf("SaveSettings() unexpected error: %v", err)
	}

	written, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("read settings after save: %v", err)
	}
	var got map[string]any
	var want map[string]any
	if err := json.Unmarshal(written, &got); err != nil {
		t.Fatalf("decode written settings: %v", err)
	}
	if err := json.Unmarshal(settingsBytes, &want); err != nil {
		t.Fatalf("decode source settings: %v", err)
	}

	assertSemanticValueEqual(t, got["future_top_level_setting"], want["future_top_level_setting"], "top-level unknown field")
	gotProjectRow := got["topics"].([]any)[0].(map[string]any)
	wantProjectRow := want["project_themes"].([]any)[0].(map[string]any)
	assertSemanticValueEqual(t, gotProjectRow["future_row_setting"], wantProjectRow["future_row_setting"], "project taxonomy unknown field")
	assertSemanticValueEqual(t, gotProjectRow["color"], wantProjectRow["color"], "project color")
	gotProjectFallback := got["topic_fallback"].(map[string]any)
	wantProjectFallback := want["project_theme_fallback"].(map[string]any)
	assertSemanticValueEqual(t, gotProjectFallback["future_fallback_setting"], wantProjectFallback["future_fallback_setting"], "project fallback unknown field")
	gotTopicRow := got["topics"].([]any)[1].(map[string]any)
	wantTopicRow := want["publication_topics"].([]any)[0].(map[string]any)
	if _, exists := gotTopicRow["future_topic_setting"]; !exists {
		t.Fatal("publication taxonomy null-valued unknown field was dropped")
	}
	assertSemanticValueEqual(t, gotTopicRow["future_topic_setting"], wantTopicRow["future_topic_setting"], "publication taxonomy unknown field")
	assertSemanticValueEqual(t, gotTopicRow["color"], wantTopicRow["color"], "publication color")
	gotTopicFallback := got["topic_fallback"].(map[string]any)
	wantTopicFallback := want["publication_topic_fallback"].(map[string]any)
	assertSemanticValueEqual(t, gotTopicFallback["future_topic_fallback_setting"], wantTopicFallback["future_topic_fallback_setting"], "publication fallback unknown field")
	if gotProjectRow["label_en"] != "Retrofit Updated" {
		t.Fatalf("known edited label = %#v, want Retrofit Updated", gotProjectRow["label_en"])
	}
}

func TestAppReopenLoadsManualJSONEditsAndPreservesManualTaxonomyIDsOnSave(t *testing.T) {
	initialSettings := fixtureSettingsJSON()
	repoRoot := newTempRepoFixture(t, initialSettings)
	settingsPath := filepath.Join(repoRoot, "data", "settings.json")
	projectsPath := filepath.Join(repoRoot, "data", "projects.json")
	publicationsPath := filepath.Join(repoRoot, "data", "publications.json")

	// Load once with the original values, then discard this App instance to
	// model closing the editor before editing the JSON files by hand.
	closedApp := &App{repoRoot: repoRoot}
	if _, err := closedApp.LoadSettings(); err != nil {
		t.Fatalf("initial LoadSettings() unexpected error: %v", err)
	}

	manualSettings := []byte(`{
  "schema_version": 4,
  "main_page_sections": [
    "skills",
    "experience",
    "education",
    "scholarships",
    "certifications",
    "awards"
  ],
  "hidden_main_page_sections": [
    "teaching"
  ],
  "project_themes": [
    {
      "id": "project-theme-edited-by-hand",
      "label_en": "Climate Responsive Design",
      "label_ko": "수동 프로젝트 테마"
    }
  ],
  "project_theme_fallback": {
    "label_en": "Manual Project Fallback",
    "label_ko": "수동 프로젝트 기타"
  },
  "publication_topics": [
    {
      "id": "publication-topic-edited-by-hand",
      "label_en": "Building Performance",
      "label_ko": "수동 논문 주제"
    }
  ],
  "publication_topic_fallback": {
    "label_en": "Manual Publication Fallback",
    "label_ko": "수동 논문 기타"
  }
}
`)
	manualProjects := []byte(`[
  {
    "theme": "project-theme-edited-by-hand",
    "title_en": "Project edited while Profile-Editor was closed"
  }
]
`)
	manualPublications := []byte(`[
  {
    "topic": "publication-topic-edited-by-hand",
    "title_en": "Publication edited while Profile-Editor was closed"
  }
]
`)
	manualFiles := map[string][]byte{
		settingsPath:     manualSettings,
		projectsPath:     manualProjects,
		publicationsPath: manualPublications,
	}
	for path, contents := range manualFiles {
		if err := os.WriteFile(path, contents, 0o644); err != nil {
			t.Fatalf("write manual JSON edit %s: %v", filepath.Base(path), err)
		}
	}

	// A newly opened App must read the current files, not values retained from
	// the prior process or generated bindings.
	reopenedApp := &App{repoRoot: repoRoot}
	loaded, err := reopenedApp.LoadSettings()
	if err != nil {
		t.Fatalf("LoadSettings() after manual JSON edits unexpected error: %v", err)
	}
	if loaded.Revision != revisionOf(manualSettings) {
		t.Errorf("revision after reopen = %q, want manual settings revision %q", loaded.Revision, revisionOf(manualSettings))
	}
	if len(loaded.Settings.Topics) != 2 {
		t.Fatalf("project themes after reopen = %#v, want one manually edited theme", loaded.Settings.Topics)
	}
	projectTheme := loaded.Settings.Topics[0]
	if projectTheme.ID != "project-theme-edited-by-hand" || projectTheme.LabelEN != "Climate Responsive Design" || projectTheme.LabelKO != "수동 프로젝트 테마" {
		t.Errorf("project theme after reopen = %#v, want manually edited values", projectTheme)
	}
	if len(loaded.Settings.Topics) != 2 {
		t.Fatalf("publication topics after reopen = %#v, want one manually edited topic", loaded.Settings.Topics)
	}
	publicationTopic := loaded.Settings.Topics[1]
	if publicationTopic.ID != "publication-topic-edited-by-hand" || publicationTopic.LabelEN != "Building Performance" || publicationTopic.LabelKO != "수동 논문 주제" {
		t.Errorf("publication topic after reopen = %#v, want manually edited values", publicationTopic)
	}
	if loaded.Usage.ProjectThemes["project-theme-edited-by-hand"] != 1 {
		t.Errorf("manual project theme usage = %d, want 1", loaded.Usage.ProjectThemes["project-theme-edited-by-hand"])
	}
	if loaded.Usage.PublicationTopics["publication-topic-edited-by-hand"] != 1 {
		t.Errorf("manual publication topic usage = %d, want 1", loaded.Usage.PublicationTopics["publication-topic-edited-by-hand"])
	}
	if loaded.Settings.MainPageSections[0] != "skills" || !reflect.DeepEqual(loaded.Settings.HiddenMainPageSections, []string{"teaching", "academic_activities", "patents"}) {
		t.Errorf("section settings after reopen = visible %#v, hidden %#v; want manual ordering and visibility", loaded.Settings.MainPageSections, loaded.Settings.HiddenMainPageSections)
	}

	// Saving an unrelated UI change after reopening must treat the manual IDs
	// as established IDs, retain their references, and leave record JSON alone.
	draft := loaded.Settings
	draft.TopicFallback.LabelEN = "Fallback updated in editor"
	saved, err := reopenedApp.SaveSettings(SaveSettingsRequest{
		Settings: draft,
		Revision: loaded.Revision,
	})
	if err != nil {
		t.Fatalf("SaveSettings() after manual JSON edits unexpected error: %v", err)
	}
	if saved.Settings.Topics[0].ID != "project-theme-edited-by-hand" {
		t.Errorf("saved project theme ID = %q, want manually assigned ID", saved.Settings.Topics[0].ID)
	}
	if saved.Settings.Topics[1].ID != "publication-topic-edited-by-hand" {
		t.Errorf("saved publication topic ID = %q, want manually assigned ID", saved.Settings.Topics[1].ID)
	}

	writtenSettings, _, err := readSettings(settingsPath)
	if err != nil {
		t.Fatalf("read settings after save: %v", err)
	}
	if writtenSettings.Topics[0].ID != "project-theme-edited-by-hand" || writtenSettings.Topics[1].ID != "publication-topic-edited-by-hand" {
		t.Errorf("written taxonomy IDs changed: project %q, publication %q", writtenSettings.Topics[0].ID, writtenSettings.Topics[1].ID)
	}
	projectsAfterSave, err := os.ReadFile(projectsPath)
	if err != nil {
		t.Fatalf("read projects after save: %v", err)
	}
	if string(projectsAfterSave) != string(manualProjects) {
		t.Fatalf("SaveSettings() changed manually edited projects.json\ngot:  %q\nwant: %q", projectsAfterSave, manualProjects)
	}
	publicationsAfterSave, err := os.ReadFile(publicationsPath)
	if err != nil {
		t.Fatalf("read publications after save: %v", err)
	}
	if string(publicationsAfterSave) != string(manualPublications) {
		t.Fatalf("SaveSettings() changed manually edited publications.json\ngot:  %q\nwant: %q", publicationsAfterSave, manualPublications)
	}

	loadedAgain, err := (&App{repoRoot: repoRoot}).LoadSettings()
	if err != nil {
		t.Fatalf("second reopen LoadSettings() unexpected error: %v", err)
	}
	if loadedAgain.Settings.Topics[0].ID != "project-theme-edited-by-hand" || loadedAgain.Settings.Topics[1].ID != "publication-topic-edited-by-hand" {
		t.Errorf("taxonomy IDs after save and second reopen changed: project %q, publication %q", loadedAgain.Settings.Topics[0].ID, loadedAgain.Settings.Topics[1].ID)
	}
}

func validSettingsDocument() SettingsDocument {
	return SettingsDocument{
		SchemaVersion: settingsSchemaVersion,
		MainPageSections: []string{
			"experience",
			"education",
			"scholarships",
			"certifications",
			"awards",
			"teaching",
			"skills",
			"academic_activities",
			"patents",
		},
		HiddenMainPageSections: []string{},
		Topics: []TaxonomyItem{
			{ID: "retrofit", BilingualLabel: BilingualLabel{LabelEN: "Retrofit", LabelKO: "Retrofit Korean"}},
			{ID: "hvac", BilingualLabel: BilingualLabel{LabelEN: "HVAC", LabelKO: "HVAC Korean"}},
		},
		TopicFallback: BilingualLabel{LabelEN: "Other", LabelKO: "Other Korean"},
	}
}

func assertSemanticValueEqual(t *testing.T, got, want any, context string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%s changed semantically\ngot:  %#v\nwant: %#v", context, got, want)
	}
}

func newTempRepoFixture(t *testing.T, settings []byte) string {
	t.Helper()

	root := t.TempDir()
	dataDirectory := filepath.Join(root, "data")
	if err := os.MkdirAll(dataDirectory, 0o755); err != nil {
		t.Fatalf("create fixture data directory: %v", err)
	}
	files := map[string][]byte{
		filepath.Join("data", "settings.json"): settings,
		filepath.Join("data", "profile.json"): []byte(`{
  "schema_version": 1,
  "identity": {"name_en":"Fixture","name_ko":"픽스처","role_en":"Role","role_ko":"역할","affiliation_en":"Lab","affiliation_ko":"연구실","location_en":"Seoul","location_ko":"서울"},
  "profile_card": {"media":{"src":"profile.jpg","type":"image"},"title_en":"Title","title_ko":"직함","credentials":[]},
  "intro": {"short_en":"Intro","short_ko":"소개"},
  "contact": {"email":"fixture@example.com","github":"https://example.com/github","orcid":"https://example.com/orcid","scholar":"https://example.com/scholar"},
  "experience": [{"title_en":"Researcher","title_ko":"연구원","institution_en":"Lab","institution_ko":"연구실","period_en":"2025-present","period_ko":"2025-현재"}],
  "education": [],
  "scholarships": [],
  "certifications": [],
  "teaching": [],
  "skills": []
}
`),
		filepath.Join("data", "projects.json"): []byte(`[
  {
    "title_en": "Fixture project",
    "title_ko": "Fixture project Korean",
    "start_date": "2025-01-01",
    "end_date": "2025-12-31",
    "theme": "retrofit",
    "funder_en": "Fixture funder",
    "funder_ko": "Fixture funder Korean",
    "notes_en": ["Fixture note"],
    "notes_kr": ["Fixture note Korean"],
    "media": []
  }
]
`),
		filepath.Join("data", "software.json"): []byte("[]\n"),
		filepath.Join("data", "people.json"): []byte(`[
  {
    "id": "fixture-self",
    "name_en": "Fixture Self",
    "name_ko": "픽스처 본인",
    "is_self": true,
    "notes_en": [],
    "notes_ko": []
  }
]
`),
		filepath.Join("data", "awards.json"): []byte("[]\n"),
		filepath.Join("data", "academic_activities.json"): []byte(`{
  "schema_version": 1,
  "reviews": [],
  "invited_talks": [],
  "conference_service": [],
  "professional_service": [],
  "editorial_service": []
}
`),
		filepath.Join("data", "board.json"): []byte("[]\n"),
		filepath.Join("data", "publications.json"): []byte(`[
  {
    "title_en": "Fixture publication",
    "title_ko": "",
    "abstract_en": "",
    "abstract_ko": "",
    "keywords_en": [],
    "keywords_ko": [],
    "author_ids": ["fixture-self"],
    "date": "2025",

    "under_review": false,
    "in_press": false,
    "publication_type": "international-journal",
    "topic": "hvac",
    "award_id": "",
    "doi": "",
    "url": "",
    "note": ""
  }
]
`),
		"index.html": []byte("<!doctype html><title>fixture</title>\n"),
	}
	for relativePath, contents := range files {
		path := filepath.Join(root, relativePath)
		if err := os.WriteFile(path, contents, 0o644); err != nil {
			t.Fatalf("write fixture file %s: %v", relativePath, err)
		}
	}
	if !isRepoRoot(root) {
		t.Fatal("temporary repository fixture does not satisfy repo root markers")
	}
	return root
}

func fixtureSettingsJSON() []byte {
	return []byte(`{
    "schema_version": 3,
    "main_page_sections": [
        "experience",
        "education",
        "scholarships",
        "certifications",
        "awards",
        "teaching",
        "skills"
    ],
    "project_themes": [
        {
            "id": "retrofit",
            "label_en": "Retrofit",
            "label_ko": "Retrofit Korean"
        }
    ],
    "project_theme_fallback": {
        "label_en": "Other",
        "label_ko": "Other Korean"
    },
    "publication_topics": [
        {
            "id": "hvac",
            "label_en": "HVAC",
            "label_ko": "HVAC Korean"
        }
    ],
    "publication_topic_fallback": {
        "label_en": "Other",
        "label_ko": "Other Korean"
    }
}
`)
}
