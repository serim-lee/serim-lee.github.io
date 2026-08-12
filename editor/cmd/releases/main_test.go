package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Gonie-Gonie/Gonie-Gonie.github.io/editor/internal/release"
)

func cliFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	write := func(name string, value any) {
		raw, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, raw, 0644); err != nil {
			t.Fatal(err)
		}
	}
	entry := release.Entry{Version: "1.0.0", Date: "2026-10-28", Notes: "app/releases/1.0.0.md"}
	write("app/release.json", release.Manifest{Entry: entry, DataVersion: "1.0.0"})
	write("app/releases/index.json", release.Index{Version: "1.0.0", Releases: []release.Entry{entry}})
	write("app/update-policy.json", release.Policy{Version: "1.0.0", Default: "preserve",
		Replace:    []string{"app/", "editor/", "index.html"},
		Preserve:   []string{"data/", "exports/", "CNAME", ".git/", "editor/runtime/", "editor/cmd/editor/build/bin/", "editor/cmd/editor/frontend/wailsjs/go/"},
		Regenerate: []string{"exports/publications.bib", "exports/publications.ris", "exports/publications.csl.json", "exports/build-report.json"},
	})
	write("data/version.json", map[string]any{"version": "1.0.0"})
	for _, name := range []string{"awards", "board", "people", "projects", "publications", "software"} {
		write("data/"+name+".json", []any{})
	}
	for _, name := range []string{"academic_activities", "profile", "settings"} {
		write("data/"+name+".json", map[string]any{})
	}
	write("editor/cmd/editor/wails.json", map[string]any{"info": map[string]any{"productVersion": "1.0.0"}})
	if err := os.WriteFile(filepath.Join(root, "app", "releases", "1.0.0.md"), []byte("# 1.0.0\n\n2026-10-28\n\n- Initial release.\n"), 0644); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestCheckAndClassifyCommandsUseRequestedRepository(t *testing.T) {
	root := cliFixture(t)
	markerPath := filepath.Join(root, "data", "version.json")
	before, _ := os.ReadFile(markerPath)
	var output, diagnostics bytes.Buffer
	if err := runWithWriters([]string{"check", "--root", root}, &output, &diagnostics); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "Release 1.0.0:") || diagnostics.Len() != 0 {
		t.Fatal(output.String(), diagnostics.String())
	}
	output.Reset()
	if err := runWithWriters([]string{"classify", "--root", root, "data/profile.json", "app/assets/js/app.js", "exports/build-report.json", "exports/publication-citations.json", "unknown/file"}, &output, &diagnostics); err != nil {
		t.Fatal(err)
	}
	want := "preserve\tdata/profile.json\nreplace\tapp/assets/js/app.js\nregenerate\texports/build-report.json\npreserve\texports/publication-citations.json\npreserve\tunknown/file\n"
	if output.String() != want || diagnostics.Len() != 0 {
		t.Fatalf("output=%q diagnostics=%q", output.String(), diagnostics.String())
	}
	after, _ := os.ReadFile(markerPath)
	if !bytes.Equal(before, after) {
		t.Fatal("commands modified source data")
	}
	if _, err := os.Stat(filepath.Join(root, "exports")); !os.IsNotExist(err) {
		t.Fatal("commands created output files")
	}
}

func TestCommandErrorsDoNotWriteSuccessOutput(t *testing.T) {
	root := cliFixture(t)
	for name, args := range map[string][]string{
		"unknown command":       {"wrong", "--root", root},
		"unknown option":        {"check", "--root", root, "--unknown"},
		"root value missing":    {"check", "--root"},
		"extra check argument":  {"check", "--root", root, "extra"},
		"classify needs paths":  {"classify", "--root", root},
		"unsafe classification": {"classify", "--root", root, "../data/profile.json"},
		"invalid root":          {"check", "--root", t.TempDir()},
	} {
		t.Run(name, func(t *testing.T) {
			var output, diagnostics bytes.Buffer
			if err := runWithWriters(args, &output, &diagnostics); err == nil || output.Len() != 0 {
				t.Fatal(err, output.String())
			}
		})
	}
	if err := os.WriteFile(filepath.Join(root, "data", "version.json"), []byte(`{"version":"2.0.0"}`), 0644); err != nil {
		t.Fatal(err)
	}
	var output, diagnostics bytes.Buffer
	if err := runWithWriters([]string{"check", "--root", root}, &output, &diagnostics); err == nil || output.Len() != 0 {
		t.Fatal("incompatible bundle reported success")
	}
}

func TestPreserveDataFlagSkipsPersonalDataAndReportsIt(t *testing.T) {
	root := cliFixture(t)
	path := filepath.Join(root, "data", "version.json")
	if err := os.WriteFile(path, []byte("unfinished personal data"), 0644); err != nil {
		t.Fatal(err)
	}
	var output, diagnostics bytes.Buffer
	if err := runWithWriters([]string{"check", "--root", root, "--preserve-data"}, &output, &diagnostics); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "personal data checks skipped") || strings.Contains(output.String(), "data versions,") {
		t.Fatal(output.String())
	}
	raw, _ := os.ReadFile(path)
	if string(raw) != "unfinished personal data" {
		t.Fatal("check modified data")
	}
}
