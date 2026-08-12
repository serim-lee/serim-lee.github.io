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

func TestHiddenWebPagesDefaultsAndValidation(t *testing.T) {
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
	if document.HiddenWebPages == nil || len(document.HiddenWebPages) != 0 {
		t.Fatalf("missing setting should default to every Web page visible: %#v", document.HiddenWebPages)
	}

	for _, test := range []struct {
		name  string
		value any
		valid bool
	}{
		{"empty", []string{}, true},
		{"all pages", supportedWebPages, true},
		{"one page", []string{"software"}, true},
		{"null", nil, false},
		{"string", "software", false},
		{"object", map[string]any{"software": false}, false},
		{"unknown page", []string{"home"}, false},
		{"duplicate", []string{"software", "software"}, false},
		{"wrong item type", []any{true}, false},
		{"null item", []any{nil}, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			var fields map[string]any
			if err := json.Unmarshal(fixtureSettingsJSON(), &fields); err != nil {
				t.Fatal(err)
			}
			fields["hidden_web_pages"] = test.value
			writeJSONFixture(t, path, fields)
			loaded, _, err := readSettings(path)
			if err == nil {
				err = normaliseLoadedSettings(&loaded)
			}
			if (err == nil) != test.valid {
				t.Fatalf("load error = %v, valid %v", err, test.valid)
			}

			// Wails request decoding and save validation must agree with loading.
			raw, _ := json.Marshal(fields)
			var input SettingsDocument
			err = json.Unmarshal(raw, &input)
			if err == nil {
				input.MainPageSections = document.MainPageSections
				input.HiddenMainPageSections = document.HiddenMainPageSections
				input.CVSections = document.CVSections
				input.HiddenCVSections = document.HiddenCVSections
				input.Topics = document.Topics
				input.TopicFallback = document.TopicFallback
				_, err = normaliseForSave(input, document, SettingsUsage{})
			}
			if (err == nil) != test.valid {
				t.Fatalf("save error = %v, valid %v", err, test.valid)
			}
		})
	}
}

func TestHiddenWebPagesSavePreservesContentAndCV(t *testing.T) {
	var settings map[string]any
	if err := json.Unmarshal(fixtureSettingsJSON(), &settings); err != nil {
		t.Fatal(err)
	}
	settings["cv_sections"] = []string{"experience"}
	hiddenCV := []string{}
	for _, section := range supportedCVSections {
		if section != "experience" {
			hiddenCV = append(hiddenCV, section)
		}
	}
	settings["hidden_cv_sections"] = hiddenCV
	settings["future_web_metadata"] = map[string]any{"preserved": true}
	raw, _ := json.Marshal(settings)
	root := newTempRepoFixture(t, raw)
	app := &App{repoRoot: root}
	loaded, err := app.LoadEditorData()
	if err != nil {
		t.Fatal(err)
	}
	before := map[string][]byte{}
	for _, name := range []string{"profile.json", "people.json", "publications.json", "projects.json", "software.json", "board.json", "awards.json", "academic_activities.json", "version.json"} {
		path := filepath.Join(root, "data", name)
		content, err := os.ReadFile(path)
		if os.IsNotExist(err) && name == "version.json" {
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		before[name] = content
	}

	for _, hidden := range [][]string{supportedWebPages, {}} {
		request := saveRequestRevisions(loaded)
		request.SaveSettings = true
		request.Settings.HiddenWebPages = append([]string{}, hidden...)
		saved, err := app.SaveEditorData(request)
		if err != nil {
			t.Fatal(err)
		}
		loaded, err = app.LoadEditorData()
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(loaded.Settings.HiddenWebPages, hidden) {
			t.Fatalf("hidden Web pages after reload = %#v, want %#v", loaded.Settings.HiddenWebPages, hidden)
		}
		if !reflect.DeepEqual(saved.Settings.CVSections, []string{"experience"}) || !reflect.DeepEqual(saved.Settings.HiddenCVSections, hiddenCV) {
			t.Fatal("Web visibility changed CV settings")
		}
		for name, content := range before {
			if !bytes.Equal(readFixtureFile(t, filepath.Join(root, "data", name)), content) {
				t.Fatalf("Web visibility save modified data/%s", name)
			}
		}
	}
	stored := readFixtureFile(t, filepath.Join(root, "data", "settings.json"))
	if !strings.Contains(string(stored), `"future_web_metadata"`) || !strings.Contains(string(stored), `"hidden_web_pages": []`) {
		t.Fatalf("unknown setting or empty Web setting lost: %s", stored)
	}
}

func TestOlderSettingsClientsPreserveHiddenWebPages(t *testing.T) {
	root := newTempRepoFixture(t, fixtureSettingsJSON())
	app := &App{repoRoot: root}
	loaded, err := app.LoadSettings()
	if err != nil {
		t.Fatal(err)
	}
	draft := loaded.Settings
	draft.HiddenWebPages = []string{"publications", "board"}
	saved, err := app.SaveSettings(SaveSettingsRequest{Settings: draft, Revision: loaded.Revision})
	if err != nil {
		t.Fatal(err)
	}
	// An older client omits the optional field, which decodes as nil.
	olderDraft := saved.Settings
	olderDraft.HiddenWebPages = nil
	resaved, err := app.SaveSettings(SaveSettingsRequest{Settings: olderDraft, Revision: saved.Revision})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(resaved.Settings.HiddenWebPages, draft.HiddenWebPages) {
		t.Fatalf("older client restored hidden Web pages: %#v", resaved.Settings.HiddenWebPages)
	}
}
