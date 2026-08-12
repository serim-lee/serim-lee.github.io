package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSharedTopicsMigrateLegacyIDsColorsAndMetadata(t *testing.T) {
	rootSettings, err := decodeObject(fixtureSettingsJSON())
	if err != nil {
		t.Fatal(err)
	}
	rootSettings["project_themes"] = json.RawMessage(`[{"id":"retrofit","label_en":"Retrofit","label_ko":"Retrofit Korean","project_extra":true}]`)
	rootSettings["publication_topics"] = json.RawMessage(`[{"id":"retrofit","label_en":"Paper retrofit","label_ko":"Paper retrofit Korean","color":"#12aBcD","paper_extra":true},{"id":"hvac","label_en":"HVAC","label_ko":"HVAC Korean"}]`)
	rootSettings["project_theme_fallback"] = json.RawMessage(`{"label_en":"Other","label_ko":"Other Korean","project_extra":true}`)
	rootSettings["publication_topic_fallback"] = json.RawMessage(`{"label_en":"Other","label_ko":"Other Korean","paper_extra":true}`)
	legacy, err := json.Marshal(rootSettings)
	if err != nil {
		t.Fatal(err)
	}
	root := newTempRepoFixture(t, legacy)
	app := &App{repoRoot: root}
	loaded, err := app.LoadSettings()
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Settings.Topics) != 2 || loaded.Settings.Topics[0].ID != "retrofit" || loaded.Settings.Topics[1].ID != "hvac" {
		t.Fatalf("legacy topics were lost, duplicated or reordered: %#v", loaded.Settings.Topics)
	}
	shared := loaded.Settings.Topics[0]
	if shared.LabelEN != "Retrofit" || shared.Color != "#12aBcD" || len(shared.extraFields) != 2 {
		t.Fatalf("shared topic did not preserve its name, color and metadata: %#v", shared)
	}
	if len(loaded.Settings.TopicFallback.extraFields) != 2 {
		t.Fatal("legacy fallback metadata was lost")
	}
	settingsPath := filepath.Join(root, "data", "settings.json")
	before, _ := os.ReadFile(settingsPath)
	if !bytes.Equal(before, legacy) {
		t.Fatal("loading legacy settings wrote to disk")
	}
	projectPath := filepath.Join(root, "data", "projects.json")
	publicationPath := filepath.Join(root, "data", "publications.json")
	projectsBefore, _ := os.ReadFile(projectPath)
	publicationsBefore, _ := os.ReadFile(publicationPath)
	loaded.Settings.Topics[0].LabelEN = "Shared retrofit"
	saved, err := app.SaveSettings(SaveSettingsRequest{Settings: loaded.Settings, Revision: loaded.Revision})
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(settingsPath)
	written, err := decodeObject(raw)
	if err != nil {
		t.Fatal(err)
	}
	for _, oldKey := range []string{"project_themes", "publication_topics", "project_theme_fallback", "publication_topic_fallback"} {
		if _, exists := written[oldKey]; exists {
			t.Fatalf("saved settings still contain the separate catalog %s", oldKey)
		}
	}
	if written["topics"] == nil || written["topic_fallback"] == nil || saved.Settings.SchemaVersion != settingsSchemaVersion {
		t.Fatal("shared settings were not saved in the current schema")
	}
	projectsAfter, _ := os.ReadFile(projectPath)
	publicationsAfter, _ := os.ReadFile(publicationPath)
	if !bytes.Equal(projectsBefore, projectsAfter) || !bytes.Equal(publicationsBefore, publicationsAfter) {
		t.Fatal("migrating topics changed content references")
	}
}

func TestSharedTopicsDeletionProtectsBothKindsOfUsage(t *testing.T) {
	current := validSettingsDocument()
	if err := normaliseLoadedSettings(&current); err != nil {
		t.Fatal(err)
	}
	input := current
	input.Topics = append([]TaxonomyItem{}, current.Topics[1:]...)
	usage := SettingsUsage{ProjectThemes: map[string]int{"retrofit": 2}, PublicationTopics: map[string]int{"retrofit": 3}}
	if _, err := normaliseForSave(input, current, usage); err == nil || !strings.Contains(err.Error(), "5개") {
		t.Fatalf("shared usage was not aggregated when protecting deletion: %v", err)
	}
	usage.ProjectThemes = map[string]int{}
	if _, err := normaliseForSave(input, current, usage); err == nil {
		t.Fatal("a topic used only by papers could be deleted")
	}
	usage.PublicationTopics = map[string]int{}
	if _, err := normaliseForSave(input, current, usage); err != nil {
		t.Fatalf("an unused shared topic could not be deleted: %v", err)
	}
}
