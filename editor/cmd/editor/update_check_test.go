package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Gonie-Gonie/Gonie-Gonie.github.io/editor/internal/updates"
)

func updateMetadataFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "app"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "app", "release.json"), []byte(`{"version":"1.9.0","upstream":{"repository":"Gonie-Gonie/Gonie-Gonie.github.io","ref":"main"}}`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "data"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "data", "version.json"), []byte(`{"version":"2.0.0"}`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "data", "profile.json"), []byte("invalid personal JSON"), 0644); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestUpdateMetadataBypassesPersonalDataValidationAndDoesNotChangeDrafts(t *testing.T) {
	root := updateMetadataFixture(t)
	before, _ := os.ReadFile(filepath.Join(root, "data", "profile.json"))
	app := &App{repoRoot: root, dirty: true, startupErr: errors.New("personal data error")}
	info, err := app.GetUpdateInfo()
	if err != nil || info.CurrentVersion != "1.9.0" {
		t.Fatal(info, err)
	}
	if _, err := app.LoadEditorData(); err == nil {
		t.Fatal("invalid personal data unexpectedly loaded")
	}
	called := false
	response, err := app.checkUpdates(func(ctx context.Context, options updates.Options) (updates.Response, error) {
		called = true
		if options.Root != root || options.Client != nil || options.BaseURL != "" {
			t.Fatal(options)
		}
		deadline, ok := ctx.Deadline()
		if !ok || time.Until(deadline) <= 0 || time.Until(deadline) > updates.CheckTimeout {
			t.Fatal("check has no bounded context")
		}
		return updates.Response{Info: info, LatestVersion: "1.10.0", UpdateAvailable: true, Releases: []updates.Release{{Version: "1.10.0", Date: "2026-10-28", Notes: []string{"Change"}}}}, nil
	})
	if err != nil || !called || !response.UpdateAvailable || !app.dirty || app.operationBusy.Load() {
		t.Fatal(response, err)
	}
	after, _ := os.ReadFile(filepath.Join(root, "data", "profile.json"))
	if !bytes.Equal(before, after) {
		t.Fatal("check changed personal data")
	}
	raw, _ := json.Marshal(response)
	var fields map[string]any
	json.Unmarshal(raw, &fields)
	if fields["current_version"] != "1.9.0" || fields["source_repository"] != updates.DefaultRepository || fields["Info"] != nil {
		t.Fatal(string(raw))
	}
}

func TestUpdateFailureLeavesCurrentMetadataReadable(t *testing.T) {
	app := &App{repoRoot: updateMetadataFixture(t)}
	info, err := app.GetUpdateInfo()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := app.checkUpdates(func(context.Context, updates.Options) (updates.Response, error) {
		return updates.Response{}, errors.New("HTTP 404")
	}); err == nil {
		t.Fatal("failure was hidden")
	}
	after, err := app.GetUpdateInfo()
	if err != nil || after != info {
		t.Fatal(after, err)
	}
}
