package main

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Gonie-Gonie/Gonie-Gonie.github.io/editor/internal/citations"
)

func TestCitationUpdateStreamsNativeResultsAndKeepsFailureOutput(t *testing.T) {
	for _, status := range []string{"changed", "unchanged", "partial failure", "file error"} {
		t.Run(status, func(t *testing.T) {
			root := newTempRepoFixture(t, fixtureSettingsJSON())
			app := &App{repoRoot: root}
			var events []OperationProgress
			app.progressSink = func(event OperationProgress) { events = append(events, event) }
			output := "[OpenAlex] 한글 논문: cite 1 -> 2\n"
			if status == "unchanged" {
				output = "No citation count changes.\n"
			}
			response, err := app.updateCitations(CitationUpdateRequest{OperationID: "native"}, func(ctx context.Context, options citations.Options) (citations.Result, error) {
				if options.Root != root || options.Client != nil || options.KCIOnly || len(options.KCICounts) != 0 {
					t.Fatalf("incorrect native options: %#v", options)
				}
				deadline, exists := ctx.Deadline()
				if !exists || time.Until(deadline) <= 0 || time.Until(deadline) > 10*time.Minute {
					t.Fatal("missing bounded context")
				}
				options.Log(output)
				if status == "file error" {
					return citations.Result{Output: output}, errors.New("cannot write snapshot")
				}
				return citations.Result{Success: status != "partial failure", Output: output}, nil
			})
			if err != nil {
				t.Fatal(err)
			}
			if response.Success != (status == "changed" || status == "unchanged") || !strings.Contains(response.Output, output) {
				t.Fatalf("incorrect result: %#v", response)
			}
			var streamed strings.Builder
			for i, event := range events {
				if event.OperationID != "native" {
					t.Fatal(event)
				}
				streamed.WriteString(event.Output)
				if event.Percent == 100 && (!response.Success || i != len(events)-1) {
					t.Fatal("premature completion")
				}
			}
			if streamed.String() != response.Output {
				t.Fatal("stream and final console output differ")
			}
			if status == "file error" && !strings.Contains(response.Output, "cannot write snapshot") {
				t.Fatal("lost file error")
			}
			if app.dirty || app.operationBusy.Load() {
				t.Fatal("update changed drafts or retained lock")
			}
		})
	}
}

func TestCitationUpdateRejectsUnsavedAndConcurrentWork(t *testing.T) {
	root := newTempRepoFixture(t, fixtureSettingsJSON())
	app := &App{repoRoot: root, dirty: true}
	called := false
	update := func(context.Context, citations.Options) (citations.Result, error) {
		called = true
		return citations.Result{Success: true}, nil
	}
	if _, err := app.updateCitations(CitationUpdateRequest{}, update); err == nil || !strings.Contains(err.Error(), "먼저 저장") {
		t.Fatal(err)
	}
	if called || !app.dirty || app.operationBusy.Load() {
		t.Fatal("unsaved update was executed or draft state changed")
	}
	app.dirty = false
	app.boardMedia = map[string]stagedBoardMedia{"pending": {}}
	if _, err := app.updateCitations(CitationUpdateRequest{}, update); err == nil {
		t.Fatal("staged media accepted")
	}
	app.boardMedia = nil
	_, err := app.updateCitations(CitationUpdateRequest{}, func(context.Context, citations.Options) (citations.Result, error) {
		for name, operation := range map[string]func() error{
			"update": func() error { _, err := app.UpdateCitations(CitationUpdateRequest{}); return err },
			"save":   func() error { _, err := app.SaveEditorData(SaveEditorDataRequest{}); return err },
			"push":   func() error { _, err := app.PushGit(GitPushRequest{}); return err },
		} {
			if err := operation(); err == nil || !strings.Contains(err.Error(), "진행 중") {
				t.Fatalf("concurrent %s accepted: %v", name, err)
			}
		}
		app.SetDirty(true)
		return citations.Result{Success: true}, nil
	})
	if err != nil || !app.dirty || app.operationBusy.Load() {
		t.Fatal("native update lost concurrent dirty state or lock was retained", err)
	}
}

func TestCitationProgressOutputIsOptionalInJSON(t *testing.T) {
	plain, err := json.Marshal(OperationProgress{OperationID: "normal", Message: "진행 중", Percent: -1})
	if err != nil || strings.Contains(string(plain), `"output"`) {
		t.Fatal("normal progress JSON changed", err)
	}
	logged, err := json.Marshal(OperationProgress{OperationID: "console", Output: "한글\n", Percent: -1})
	if err != nil || !strings.Contains(string(logged), `"output"`) {
		t.Fatal("console progress omitted output", err)
	}
}
