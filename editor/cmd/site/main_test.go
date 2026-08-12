package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func fixtureRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "data"), 0755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"publications.json", "settings.json"} {
		if err := os.WriteFile(filepath.Join(root, "data", name), []byte("[]\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func TestRunDispatchesBuildValidationAndCheck(t *testing.T) {
	root := fixtureRoot(t)
	for _, test := range []struct {
		name         string
		args         []string
		wantBuild    bool
		validateOnly bool
	}{
		{"build", []string{"build", "--root", root}, true, false},
		{"build validation", []string{"build", "--validate-only", "--root=" + root}, true, true},
		{"validate", []string{"validate", "--root", root}, false, false},
		{"check", []string{"check", "--root", root}, false, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			calls := 0
			actions := commands{
				build: func(gotRoot string, validateOnly bool) (string, error) {
					calls++
					if !test.wantBuild || gotRoot != root || validateOnly != test.validateOnly {
						t.Fatalf("unexpected build(%q, %v)", gotRoot, validateOnly)
					}
					return "Build complete", nil
				},
				check: func(gotRoot string) (string, error) {
					calls++
					if test.wantBuild || gotRoot != root {
						t.Fatalf("unexpected check(%q)", gotRoot)
					}
					return "Checks passed\n", nil
				},
			}
			exit := run(context.Background(), test.args, nil, &stdout, &stderr, actions)
			wantOutput := "Checks passed\n"
			if test.wantBuild {
				wantOutput = "Build complete\n"
			}
			if exit != 0 || calls != 1 || stderr.Len() != 0 || stdout.String() != wantOutput {
				t.Fatalf("exit=%d calls=%d stdout=%q stderr=%q", exit, calls, stdout.String(), stderr.String())
			}
		})
	}
}

func TestRunReportsSetupAndCommandErrors(t *testing.T) {
	root := fixtureRoot(t)
	for _, name := range []string{"build", "validate", "check", "serve"} {
		var stdout, stderr bytes.Buffer
		failure := errors.New("fixture operation failed")
		actions := commands{
			build: func(string, bool) (string, error) { return "partial output\n", failure },
			check: func(string) (string, error) { return "partial output\n", failure },
			serve: func(context.Context, string, int, io.Writer) error { return failure },
		}
		exit := run(context.Background(), []string{name, "--root", root}, nil, &stdout, &stderr, actions)
		if exit != 1 || !strings.Contains(stderr.String(), name+": fixture operation failed") {
			t.Fatalf("%s: exit=%d stdout=%q stderr=%q", name, exit, stdout.String(), stderr.String())
		}
		if name != "serve" && stdout.String() != "partial output\n" {
			t.Fatalf("%s lost partial output: %q", name, stdout.String())
		}
	}
	var stdout, stderr bytes.Buffer
	if exit := run(context.Background(), []string{"build", "--root", t.TempDir()}, nil, &stdout, &stderr, commands{}); exit != 1 || !strings.Contains(stderr.String(), "invalid repository root") {
		t.Fatalf("invalid root: exit=%d stderr=%q", exit, stderr.String())
	}
}

func TestRunRejectsInvalidCommandArguments(t *testing.T) {
	cases := [][]string{
		{}, {"unknown"}, {"build", "--unknown"}, {"build", "--root"},
		{"build", "--root="}, {"build", "--root", "--validate-only"},
		{"validate", "--validate-only"}, {"check", "extra"},
		{"bibliography", "--root", "somewhere"}, {"serve", "--port"},
		{"serve", "--port", "0"}, {"serve", "--port", "-1"},
		{"serve", "--port", "65536"}, {"serve", "--port=not-a-number"},
		{"serve", "--port=1.5"}, {"build", "--port", "8000"},
	}
	for _, args := range cases {
		var stdout, stderr bytes.Buffer
		exit := run(context.Background(), args, nil, &stdout, &stderr, commands{})
		if exit != 2 || stdout.Len() != 0 || !strings.Contains(stderr.String(), "Usage:") {
			t.Fatalf("args=%v exit=%d stdout=%q stderr=%q", args, exit, stdout.String(), stderr.String())
		}
	}
	for _, args := range [][]string{{"--help"}, {"-h"}, {"serve", "--help"}} {
		var stdout, stderr bytes.Buffer
		exit := run(context.Background(), args, nil, &stdout, &stderr, commands{})
		if exit != 0 || stdout.String() != usage || stderr.Len() != 0 {
			t.Fatalf("help args=%v exit=%d stdout=%q stderr=%q", args, exit, stdout.String(), stderr.String())
		}
	}
}

func TestServeDispatchUsesDefaultAndRequestedPortAndContext(t *testing.T) {
	root := fixtureRoot(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	for _, port := range []int{8000, 9123, 65535} {
		args := []string{"serve", "--root", root}
		if port != 8000 {
			args = append(args, "--port="+strconv.Itoa(port))
		}
		var stdout, stderr bytes.Buffer
		called := false
		actions := commands{serve: func(gotContext context.Context, gotRoot string, gotPort int, output io.Writer) error {
			called = true
			if gotContext != ctx || gotContext.Err() != context.Canceled || gotRoot != root || gotPort != port || output != &stdout {
				t.Fatalf("incorrect serve context/root/port: %v %q %d", gotContext, gotRoot, gotPort)
			}
			return nil
		}}
		if exit := run(ctx, args, nil, &stdout, &stderr, actions); exit != 0 || !called || stderr.Len() != 0 {
			t.Fatalf("port=%d exit=%d called=%v stderr=%q", port, exit, called, stderr.String())
		}
	}
}

func TestBibliographyPreservesCaseShapeAndFiltersFields(t *testing.T) {
	input := `[{"settings":{"label":"저널"},"items":[{"title":"첫째"},{"title":"둘째"}]},{"settings":{},"items":[]}]`
	var stdout, stderr bytes.Buffer
	resolved := 0
	actions := commands{resolve: func(item, settings map[string]any) map[string]any {
		resolved++
		if settings["label"] != "저널" {
			t.Fatalf("settings were not passed to resolver: %#v", settings)
		}
		return map[string]any{"date": "2026-10-19", "venue_en": item["title"], "indexing": "SCIE", "ignored": "internal metadata"}
	}}
	exit := run(context.Background(), []string{"bibliography"}, strings.NewReader(input), &stdout, &stderr, actions)
	var got [][]map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("output is not JSON: %v; %q", err, stdout.String())
	}
	want := [][]map[string]any{{
		{"date": "2026-10-19", "venue_en": "첫째", "indexing": "SCIE"},
		{"date": "2026-10-19", "venue_en": "둘째", "indexing": "SCIE"},
	}, {}}
	if exit != 0 || stderr.Len() != 0 || resolved != 2 || !reflect.DeepEqual(got, want) {
		t.Fatalf("exit=%d resolved=%d got=%#v stderr=%q", exit, resolved, got, stderr.String())
	}
}

func TestBibliographyRejectsInvalidJSONBeforeResolving(t *testing.T) {
	for _, input := range []string{
		"", "null", "{}", "[", "[] []", "[] trailing", "[null]",
		`[{"items":[]}]`, `[{"settings":{}}]`, `[{"settings":{},"items":null}]`,
		`[{"settings":[],"items":[]}]`, `[{"settings":{},"items":[null]}]`,
		`[{"settings":{},"items":[1]}]`,
	} {
		var stdout, stderr bytes.Buffer
		exit := run(context.Background(), []string{"bibliography"}, strings.NewReader(input), &stdout, &stderr, commands{resolve: func(map[string]any, map[string]any) map[string]any {
			t.Fatalf("invalid input reached resolver: %q", input)
			return nil
		}})
		if exit != 1 || stdout.Len() != 0 || !strings.Contains(stderr.String(), "site bibliography:") {
			t.Fatalf("input=%q exit=%d stdout=%q stderr=%q", input, exit, stdout.String(), stderr.String())
		}
	}
	var output bytes.Buffer
	if err := bibliography(strings.NewReader("[]"), &output, nil); err != nil || output.String() != "[]\n" {
		t.Fatalf("empty cases: output=%q err=%v", output.String(), err)
	}
}

func TestStaticHandlerServesOnlyFilesInsideRoot(t *testing.T) {
	parent := t.TempDir()
	root := filepath.Join(parent, "site")
	if err := os.MkdirAll(filepath.Join(root, "assets"), 0755); err != nil {
		t.Fatal(err)
	}
	for path, content := range map[string]string{
		filepath.Join(root, "index.html"):      "home page",
		filepath.Join(root, "assets", "a.txt"): "asset content",
		filepath.Join(parent, "outside.txt"):   "outside secret",
	} {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	handler, closeFiles, err := staticHandler(root)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = closeFiles() })
	for path, want := range map[string]string{"/": "home page", "/assets/a.txt": "asset content"} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusOK || response.Body.String() != want {
			t.Fatalf("%s: status=%d body=%q", path, response.Code, response.Body.String())
		}
	}
	for _, path := range []string{"/../outside.txt", "/%2e%2e/outside.txt", "/..%5coutside.txt", "/missing"} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code == http.StatusOK || strings.Contains(response.Body.String(), "outside secret") {
			t.Fatalf("unsafe request %s: status=%d body=%q", path, response.Code, response.Body.String())
		}
	}
	t.Run("symlink escape", func(t *testing.T) {
		if err := os.Symlink(filepath.Join(parent, "outside.txt"), filepath.Join(root, "escape.txt")); err != nil {
			t.Skipf("symlinks unavailable: %v", err)
		}
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/escape.txt", nil))
		if response.Code == http.StatusOK || strings.Contains(response.Body.String(), "outside secret") {
			t.Fatalf("symlink escaped root: status=%d body=%q", response.Code, response.Body.String())
		}
	})
}
