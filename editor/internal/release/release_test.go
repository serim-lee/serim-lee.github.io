package release

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func fixturePolicy() Policy {
	return Policy{Version: "1.0.0", Default: "preserve",
		Replace:    []string{"app/", "editor/", ".github/workflows/validate.yml", "index.html", "cv.html", "fork-guide.html", "Profile-Editor.exe", "README.md"},
		Preserve:   []string{"data/", "exports/", "CNAME", ".git/", ".vscode/", ".idea/", "editor/runtime/", "editor/cmd/editor/build/bin/", "editor/cmd/editor/frontend/wailsjs/go/"},
		Regenerate: []string{"exports/publications.bib", "exports/publications.ris", "exports/publications.csl.json", "exports/build-report.json"},
	}
}

func writeFixture(t *testing.T, root, name string, value any) {
	t.Helper()
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	writeFixtureBytes(t, root, name, append(raw, '\n'))
}

func writeFixtureBytes(t *testing.T, root, name string, raw []byte) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, raw, 0644); err != nil {
		t.Fatal(err)
	}
}

func releaseFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	entry := Entry{Version: "1.0.0", Date: "2026-10-28", Notes: "app/releases/1.0.0.md"}
	writeFixture(t, root, "app/release.json", Manifest{Entry: entry, DataVersion: "1.0.0"})
	writeFixture(t, root, "app/releases/index.json", Index{Version: "1.0.0", Releases: []Entry{entry}})
	writeFixture(t, root, "app/update-policy.json", fixturePolicy())
	writeFixture(t, root, "data/version.json", map[string]any{"version": "1.0.0"})
	for _, name := range []string{"awards", "board", "people", "projects", "publications", "software"} {
		writeFixture(t, root, "data/"+name+".json", []any{})
	}
	for _, name := range []string{"academic_activities", "profile", "settings"} {
		writeFixture(t, root, "data/"+name+".json", map[string]any{})
	}
	writeFixture(t, root, "editor/cmd/editor/wails.json", map[string]any{"$schema": "https://wails.io/schemas/config.v2.json", "info": map[string]any{"productVersion": "1.0.0"}})
	writeFixtureBytes(t, root, entry.Notes, []byte("# 1.0.0\n\n2026-10-28\n\n- Initial release.\n"))
	return root
}

func TestCheckIsReadOnlyAndDoesNotRequirePersonalRecords(t *testing.T) {
	root := releaseFixture(t)
	before, _ := os.ReadFile(filepath.Join(root, "data", "version.json"))
	manifest, err := Check(root)
	if err != nil {
		t.Fatal(err)
	}
	if manifest.Version != "1.0.0" || manifest.DataVersion != "1.0.0" {
		t.Fatal(manifest)
	}
	after, _ := os.ReadFile(filepath.Join(root, "data", "version.json"))
	if !bytes.Equal(before, after) {
		t.Fatal("release check changed data marker")
	}
	if _, err := os.Stat(filepath.Join(root, "exports")); !os.IsNotExist(err) {
		t.Fatal("release check generated exports")
	}
}

func TestPreservedDataCanKeepOlderOrUnrecognizedFormat(t *testing.T) {
	for _, document := range []string{`{"version":"0.1.0"}`, `{"version":"99.0.0"}`, `{"schema_version":1}`, `invalid personal JSON`} {
		t.Run(document, func(t *testing.T) {
			root := releaseFixture(t)
			writeFixtureBytes(t, root, "data/version.json", []byte(document))
			writeFixtureBytes(t, root, "data/profile.json", []byte("unfinished personal edit"))
			if _, err := Check(root); err == nil {
				t.Fatal("strict check accepted incompatible data")
			}
			if _, err := CheckWithOptions(root, CheckOptions{PreserveData: true}); err != nil {
				t.Fatal(err)
			}
			raw, _ := os.ReadFile(filepath.Join(root, "data/version.json"))
			if string(raw) != document {
				t.Fatal("preserved data was rewritten")
			}
		})
	}
	root := releaseFixture(t)
	if err := os.RemoveAll(filepath.Join(root, "data")); err != nil {
		t.Fatal(err)
	}
	if _, err := CheckWithOptions(root, CheckOptions{PreserveData: true}); err != nil {
		t.Fatal(err)
	}
	writeFixtureBytes(t, root, "app/releases/1.0.0.md", []byte("wrong release heading"))
	if _, err := CheckWithOptions(root, CheckOptions{PreserveData: true}); err == nil {
		t.Fatal("data bypass also bypassed program release checks")
	}
}

func TestPolicyClassificationProtectsPersonalAndLocalPaths(t *testing.T) {
	policy := fixturePolicy()
	if err := policy.Validate(); err != nil {
		t.Fatal(err)
	}
	cases := map[string]string{
		"app/assets/js/app.js": "replace", "APP/ASSETS/styles.css": "replace", `editor\cmd\editor\main.go`: "replace", "Profile-Editor.exe": "replace",
		"data/profile.json": "preserve", `DATA\MEDIA\photo.jpg`: "preserve", "data/version.json": "preserve", "CNAME": "preserve", "cname": "preserve",
		"exports/publication-citations.json": "preserve", "exports/unknown-custom.json": "preserve", ".GiT/config": "preserve", ".vscode/settings.json": "preserve",
		"EDITOR/RUNTIME/go/bin/go.exe": "preserve", "editor/cmd/editor/build/bin/Profile-Editor.exe": "preserve", "editor/cmd/editor/frontend/wailsjs/go/main/App.js": "preserve",
		"exports/publications.bib": "regenerate", `EXPORTS\PUBLICATIONS.CSL.JSON`: "regenerate", "exports/build-report.json": "regenerate",
		"custom/index.html": "preserve", "appetite/config.json": "preserve", ".github/workflows/personal.yml": "preserve", ".gitignore": "preserve",
	}
	for name, want := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := policy.Classify(name)
			if err != nil || got != want {
				t.Fatalf("%q = %q, %v; want %q", name, got, err, want)
			}
		})
	}
}

func TestPolicyRejectsTraversalAndAmbiguousWindowsPaths(t *testing.T) {
	policy := fixturePolicy()
	for _, name := range []string{"", "/data/profile.json", `\data\profile.json`, `\\server\share`, "../data/profile.json", "app/../data/profile.json", "app/./x", "app//x", "app/file.", "app /file", "app/file ", "C:/data/profile.json", "app/file:stream", "app/file\x00name", "app/file\rname", "app/file\tname", "app/file?", "app/file*", "app/<name>", `app/"name"`, "app/CON", "app/nul.txt", "app/COM1.log", "app/LPT9", "app/CONOUT$"} {
		t.Run(name, func(t *testing.T) {
			if _, err := policy.Classify(name); err == nil {
				t.Fatalf("accepted unsafe path %q", name)
			}
		})
	}
}

func TestPolicyRequiresExplicitRegenerationAndPreserveByDefault(t *testing.T) {
	mutations := map[string]func(*Policy){
		"default replace":           func(p *Policy) { p.Default = "replace" },
		"unsupported policy format": func(p *Policy) { p.Version = "2.0.0" },
		"duplicate case":            func(p *Policy) { p.Replace = append(p.Replace, "APP/") },
		"backslash rule":            func(p *Policy) { p.Replace = append(p.Replace, `app\extra`) },
		"traversal":                 func(p *Policy) { p.Replace = append(p.Replace, "app/../data/") },
		"missing CNAME":             func(p *Policy) { p.Preserve = append(p.Preserve[:2], p.Preserve[3:]...) },
		"missing generated export":  func(p *Policy) { p.Regenerate = p.Regenerate[:3] },
		"broad regeneration": func(p *Policy) {
			p.Regenerate = []string{"exports/", "exports/publications.ris", "exports/publications.csl.json", "exports/build-report.json"}
		},
		"regenerate citation snapshot": func(p *Policy) { p.Regenerate[3] = "exports/publication-citations.json" },
		"local runtime overwrite":      func(p *Policy) { p.Preserve = append(p.Preserve[:6], p.Preserve[7:]...) },
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			policy := fixturePolicy()
			mutate(&policy)
			if policy.Validate() == nil {
				t.Fatal("unsafe policy accepted")
			}
		})
	}
}

func TestBundleMinorAndPatchCompatibility(t *testing.T) {
	for _, version := range []string{"1.0.0", "1.0.9", "1.8.0", "1.8.23"} {
		t.Run(version, func(t *testing.T) {
			root := releaseFixture(t)
			writeFixture(t, root, "data/version.json", map[string]any{"version": version})
			if _, err := Check(root); err != nil {
				t.Fatal(err)
			}
		})
	}
	for _, version := range []string{"0.9.0", "2.0.0", "1.0", "01.0.0"} {
		t.Run("reject "+version, func(t *testing.T) {
			root := releaseFixture(t)
			writeFixture(t, root, "data/version.json", map[string]any{"version": version})
			if _, err := Check(root); err == nil {
				t.Fatal("incompatible bundle accepted")
			}
		})
	}
	root := releaseFixture(t)
	var manifest Manifest
	readJSON(root, "app/release.json", &manifest)
	manifest.DataVersion = "1.2.0"
	writeFixture(t, root, "app/release.json", manifest)
	if _, err := Check(root); err == nil {
		t.Fatal("bundle older than required format accepted")
	}
	manifest.DataVersion = "2.0.0"
	writeFixture(t, root, "app/release.json", manifest)
	writeFixture(t, root, "data/version.json", map[string]any{"version": "2.0.0"})
	if _, err := Check(root); err == nil {
		t.Fatal("manifest declared an unsupported data major")
	}
}

func TestCanonicalDataSharesOnlyOneVersionMarker(t *testing.T) {
	for _, name := range []string{"academic_activities", "profile", "settings"} {
		for _, key := range []string{"version", "schema_version"} {
			t.Run(name+" "+key, func(t *testing.T) {
				root := releaseFixture(t)
				writeFixture(t, root, "data/"+name+".json", map[string]any{key: "1.0.0"})
				if _, err := Check(root); err == nil {
					t.Fatal("per-file metadata accepted")
				}
			})
		}
	}
	root := releaseFixture(t)
	writeFixture(t, root, "data/people.json", map[string]any{"version": "1.0.0", "items": []any{}})
	if _, err := Check(root); err == nil {
		t.Fatal("array envelope accepted as canonical data")
	}
}

func TestReleaseNotesRequireMatchingVersionRealDateAndNonemptyBullet(t *testing.T) {
	for name, note := range map[string]string{
		"wrong heading":     "# 0.9.0\n2026-10-28\n- Change.\n",
		"missing date":      "# 1.0.0\n- Change.\n",
		"date inside prose": "# 1.0.0\nReleased on 2026-10-28.\n- Change.\n",
		"empty bullet":      "# 1.0.0\n2026-10-28\n- \n",
		"no bullet":         "# 1.0.0\n2026-10-28\nChange.\n",
		"invalid UTF8":      "# 1.0.0\n2026-10-28\n- Change.\xff\n",
	} {
		t.Run(name, func(t *testing.T) {
			root := releaseFixture(t)
			writeFixtureBytes(t, root, "app/releases/1.0.0.md", []byte(note))
			if _, err := Check(root); err == nil {
				t.Fatal("invalid release note accepted")
			}
		})
	}
	root := releaseFixture(t)
	if err := checkNote(root, Entry{Version: "1.0.0", Date: "2026-02-29", Notes: "app/releases/1.0.0.md"}); err == nil {
		t.Fatal("impossible date accepted")
	}
	if err := checkNote(root, Entry{Version: "1.0.0", Date: "0000-01-01", Notes: "app/releases/1.0.0.md"}); err == nil {
		t.Fatal("year zero accepted")
	}
	if err := checkNote(root, Entry{Version: "1.0.0", Date: "2026-10-28", Notes: "app/releases/../secret.md"}); err == nil {
		t.Fatal("note path traversal accepted")
	}
}

func TestIndexAndEditorVersionsMustAgreeWithManifest(t *testing.T) {
	root := releaseFixture(t)
	older := Entry{Version: "0.9.0", Date: "2026-10-27", Notes: "app/releases/0.9.0.md"}
	writeFixtureBytes(t, root, older.Notes, []byte("# 0.9.0\r\n\r\n2026-10-27\r\n\r\n- Previous release.\r\n"))
	current := Entry{Version: "1.0.0", Date: "2026-10-28", Notes: "app/releases/1.0.0.md"}
	writeFixture(t, root, "app/releases/index.json", Index{Version: "1.0.0", Releases: []Entry{current, older}})
	if _, err := Check(root); err != nil {
		t.Fatal(err)
	}
	for name, entries := range map[string][]Entry{"wrong latest": {older, current}, "duplicate": {current, current}, "out of order": {current, older, current}, "empty": {}} {
		t.Run(name, func(t *testing.T) {
			writeFixture(t, root, "app/releases/index.json", Index{Version: "1.0.0", Releases: entries})
			if _, err := Check(root); err == nil {
				t.Fatal("invalid index accepted")
			}
		})
	}
	writeFixture(t, root, "app/releases/index.json", Index{Version: "1.0.0", Releases: []Entry{current}})
	writeFixture(t, root, "editor/cmd/editor/wails.json", map[string]any{"info": map[string]any{"productVersion": "1.0.1"}})
	if _, err := Check(root); err == nil || !strings.Contains(err.Error(), "productVersion") {
		t.Fatal(err)
	}
}

func TestMetadataJSONRejectsInvalidUTF8AndTrailingDocuments(t *testing.T) {
	for _, raw := range []string{`{"version":"1.0.0"} {}`, "{\"version\":\"1.0.0\",\"extra\":\"\xff\"}"} {
		root := releaseFixture(t)
		writeFixtureBytes(t, root, "data/version.json", []byte(raw))
		if _, err := Check(root); err == nil {
			t.Fatal("invalid JSON accepted")
		}
	}
}
