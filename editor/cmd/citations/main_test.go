package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/Gonie-Gonie/Gonie-Gonie.github.io/editor/internal/citations"
	"github.com/Gonie-Gonie/Gonie-Gonie.github.io/editor/internal/toolcli"
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

func TestRunPassesOptionsAndStreamsWithoutDuplicatingOutput(t *testing.T) {
	root := fixtureRoot(t)
	ctx := context.WithValue(context.Background(), struct{}{}, "test context")
	var stdout, stderr bytes.Buffer
	called := false
	update := func(gotContext context.Context, options citations.Options) (citations.Result, error) {
		called = true
		if gotContext != ctx || options.Root != root || !options.KCIOnly {
			t.Fatalf("incorrect update context/options: %v, %#v", gotContext, options)
		}
		want := []citations.KCICount{{DOI: "10.1234/one", Count: 0}, {DOI: "https://doi.org/10.1234/two", Count: 9007199254740991}}
		if !reflect.DeepEqual(options.KCICounts, want) {
			t.Fatalf("KCI updates = %#v, want %#v", options.KCICounts, want)
		}
		options.Log("[KCI] 한글 논문: cite 1 -> 2\n")
		options.Log("complete\n")
		return citations.Result{Success: true, Output: "[KCI] 한글 논문: cite 1 -> 2\ncomplete\n"}, nil
	}
	exit := run(ctx, []string{"--root", root, "--kci-count", "10.1234/one", "0", "--kci-only", "--kci-count", "https://doi.org/10.1234/two", "9007199254740991"}, &stdout, &stderr, update)
	if exit != 0 || !called || stderr.Len() != 0 || stdout.String() != "[KCI] 한글 논문: cite 1 -> 2\ncomplete\n" {
		t.Fatalf("exit=%d called=%v stdout=%q stderr=%q", exit, called, stdout.String(), stderr.String())
	}
}

func TestRunRejectsInvalidArgumentsBeforeCallingUpdater(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want string
	}{
		{"unknown flag", []string{"--unknown"}, "unrecognized argument"},
		{"positional argument", []string{"paper.json"}, "unrecognized argument"},
		{"positional after separator", []string{"--", "paper.json"}, "unrecognized argument"},
		{"missing root", []string{"--root"}, "requires a repository path"},
		{"empty root", []string{"--root="}, "requires a repository path"},
		{"root followed by flag", []string{"--root", "--kci-only"}, "requires a repository path"},
		{"missing DOI and count", []string{"--kci-count"}, "requires DOI and COUNT"},
		{"missing count", []string{"--kci-count", "10.1234/one"}, "requires DOI and COUNT"},
		{"missing DOI", []string{"--kci-count", "", "1"}, "requires DOI and COUNT"},
		{"flag as DOI", []string{"--kci-count", "--kci-only", "1"}, "requires DOI and COUNT"},
		{"negative count", []string{"--kci-count", "10.1234/one", "-1"}, "nonnegative integer"},
		{"signed count", []string{"--kci-count", "10.1234/one", "+1"}, "nonnegative integer"},
		{"fractional count", []string{"--kci-count", "10.1234/one", "1.5"}, "nonnegative integer"},
		{"unsafe count", []string{"--kci-count", "10.1234/one", "9007199254740992"}, "nonnegative integer"},
		{"integer overflow", []string{"--kci-count", "10.1234/one", "99999999999999999999"}, "nonnegative integer"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			called := false
			exit := run(context.Background(), test.args, &stdout, &stderr, func(context.Context, citations.Options) (citations.Result, error) {
				called = true
				return citations.Result{Success: true}, nil
			})
			if exit != 2 || called || stdout.Len() != 0 || !strings.Contains(stderr.String(), test.want) {
				t.Fatalf("exit=%d called=%v stdout=%q stderr=%q", exit, called, stdout.String(), stderr.String())
			}
		})
	}
}

func TestHelpDoesNotRequireRepositoryOrUpdater(t *testing.T) {
	for _, arg := range []string{"-h", "--help"} {
		var stdout, stderr bytes.Buffer
		exit := run(context.Background(), []string{arg}, &stdout, &stderr, nil)
		if exit != 0 || stderr.Len() != 0 || !strings.Contains(stdout.String(), usage) {
			t.Fatalf("%s: exit=%d stdout=%q stderr=%q", arg, exit, stdout.String(), stderr.String())
		}
	}
}

func TestRunReturnsFailureAndPreservesDiagnostics(t *testing.T) {
	root := fixtureRoot(t)
	for _, updateErr := range []error{nil, errors.New("cannot save citation snapshot"), context.Canceled} {
		var stdout, stderr bytes.Buffer
		exit := run(context.Background(), []string{"--root=" + root}, &stdout, &stderr, func(ctx context.Context, options citations.Options) (citations.Result, error) {
			options.Log("OpenAlex refresh failed; existing counts and dates retained\n")
			return citations.Result{Output: "must not duplicate streamed output"}, updateErr
		})
		if exit != 1 || stdout.String() != "OpenAlex refresh failed; existing counts and dates retained\n" {
			t.Fatalf("error=%v exit=%d stdout=%q stderr=%q", updateErr, exit, stdout.String(), stderr.String())
		}
		if updateErr == nil && stderr.Len() != 0 || updateErr != nil && !strings.Contains(stderr.String(), updateErr.Error()) {
			t.Fatalf("error=%v stderr=%q", updateErr, stderr.String())
		}
	}
}

func TestRunRejectsInvalidRootWithoutCallingUpdater(t *testing.T) {
	var stdout, stderr bytes.Buffer
	exit := run(context.Background(), []string{"--root", t.TempDir()}, &stdout, &stderr, func(context.Context, citations.Options) (citations.Result, error) {
		t.Fatal("invalid repository reached updater")
		return citations.Result{}, nil
	})
	if exit != 1 || stdout.Len() != 0 || !strings.Contains(stderr.String(), "invalid repository root") {
		t.Fatalf("exit=%d stdout=%q stderr=%q", exit, stdout.String(), stderr.String())
	}
}

func TestFindRootFromNestedDirectoryAndExplicitPaths(t *testing.T) {
	root := fixtureRoot(t)
	nested := filepath.Join(root, "editor", "cmd", "citations")
	if err := os.MkdirAll(nested, 0755); err != nil {
		t.Fatal(err)
	}
	for _, requested := range []string{"", root, filepath.Join("..", "..", "..")} {
		got, err := toolcli.FindRoot(nested, requested)
		if err != nil || got != root {
			t.Fatalf("requested=%q got=%q err=%v", requested, got, err)
		}
	}
	if _, err := toolcli.FindRoot(nested, nested); err == nil {
		t.Fatal("explicit root unexpectedly searched its parents")
	}
}

func TestFindRootRequiresBothRegularDataFiles(t *testing.T) {
	root := t.TempDir()
	data := filepath.Join(root, "data")
	if err := os.MkdirAll(data, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(data, "publications.json"), []byte("[]"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := toolcli.FindRoot(root, root); err == nil {
		t.Fatal("missing settings.json accepted")
	}
	if err := os.Mkdir(filepath.Join(data, "settings.json"), 0755); err != nil {
		t.Fatal(err)
	}
	if _, err := toolcli.FindRoot(root, root); err == nil || !strings.Contains(err.Error(), "must be a file") {
		t.Fatalf("directory accepted as settings.json: %v", err)
	}
	if _, err := toolcli.FindRoot(t.TempDir(), ""); err == nil || !strings.Contains(err.Error(), "use --root PATH") {
		t.Fatalf("missing-root error = %v", err)
	}
}
