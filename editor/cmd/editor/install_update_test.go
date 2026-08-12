package main

import (
	"context"
	"errors"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/Gonie-Gonie/Gonie-Gonie.github.io/editor/internal/selfupdate"
	"github.com/Gonie-Gonie/Gonie-Gonie.github.io/editor/internal/updates"
)

func installFixture(t *testing.T) (*App, installServices, *[]string) {
	t.Helper()
	root := updateMetadataFixture(t)
	app := &App{repoRoot: root, startupErr: errors.New("data is incompatible")}
	path := filepath.Join(root, "editor", "runtime", "updates", "fixture", "journal.json")
	journal := updates.Journal{ID: "fixture", State: updates.StatePendingExecutable, CurrentVersion: "1.9.0", TargetVersion: "1.10.0", Commit: "pinned", Files: []updates.InstallFile{{Path: updates.ExecutablePath, Deferred: true, SHA256: "new-sha", Existed: true, BackupReady: true, BackupSHA256: "old-sha"}}}
	calls := []string{}
	services := installServices{
		pending: func(string) ([]string, error) { return nil, nil },
		prepare: func(ctx context.Context, options updates.InstallOptions) (updates.PreparedInstall, error) {
			calls = append(calls, "prepare")
			if options.Root != root || options.TargetVersion != "1.10.0" {
				t.Fatal(options)
			}
			if _, ok := ctx.Deadline(); !ok {
				t.Fatal("unbounded installation")
			}
			options.Progress(updates.InstallProgress{Message: "download", Percent: 30})
			return updates.PreparedInstall{Root: root, JournalPath: path, Journal: journal}, nil
		},
		apply: func(gotRoot, gotPath string) (updates.Journal, error) {
			calls = append(calls, "apply")
			if gotRoot != root || gotPath != path {
				t.Fatal(gotRoot, gotPath)
			}
			return journal, nil
		},
		rollback: func(string, string) (updates.Journal, error) { calls = append(calls, "rollback"); return journal, nil },
		launch: func(request selfupdate.Request) (selfupdate.Pending, error) {
			calls = append(calls, "launch")
			if !request.Relaunch || request.WaitPID == 0 || request.ExpectedSHA256 != "new-sha" || request.StagedExecutable != filepath.Join(filepath.Dir(path), "payload", updates.ExecutablePath) {
				t.Fatal(request)
			}
			return selfupdate.Pending{}, nil
		},
		quit: func(ctx context.Context) {
			calls = append(calls, "quit")
			if ctx.Err() != nil || !app.updateClosing.Load() || app.beforeClose(ctx) {
				t.Fatal("update cannot close safely")
			}
		},
	}
	return app, services, &calls
}

func TestInstallBypassesInvalidDataAndClosesOnlyAfterHandoff(t *testing.T) {
	app, services, calls := installFixture(t)
	progress := []OperationProgress{}
	app.progressSink = func(value OperationProgress) { progress = append(progress, value) }
	result, err := app.startUpdate(StartUpdateRequest{OperationID: "install-1", ExpectedVersion: "1.10.0"}, services)
	if err != nil || !result.RestartRequired || result.TargetVersion != "1.10.0" || result.TransactionID != "fixture" {
		t.Fatal(result, err)
	}
	if !reflect.DeepEqual(*calls, []string{"prepare", "apply", "launch", "quit"}) {
		t.Fatal(*calls)
	}
	if len(progress) < 3 || progress[1].OperationID != "install-1" || progress[1].Message != "download" {
		t.Fatal(progress)
	}
	if app.operationBusy.Load() {
		t.Fatal("operation lock leaked")
	}
	if _, err = app.startUpdate(StartUpdateRequest{ExpectedVersion: "1.10.0"}, services); err == nil {
		t.Fatal("duplicate closing update accepted")
	}
}

func TestInstallFailureKeepsEditorOpenAndRollsBackHandoffFailure(t *testing.T) {
	app, services, calls := installFixture(t)
	services.launch = func(selfupdate.Request) (selfupdate.Pending, error) {
		*calls = append(*calls, "launch")
		return selfupdate.Pending{}, errors.New("script refused")
	}
	_, err := app.startUpdate(StartUpdateRequest{ExpectedVersion: "1.10.0"}, services)
	if err == nil || app.updateClosing.Load() || app.operationBusy.Load() {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(*calls, []string{"prepare", "apply", "launch", "rollback"}) {
		t.Fatal(*calls)
	}
}

func TestInstallGuardsDirtyBusyPendingAndChangedSource(t *testing.T) {
	for _, kind := range []string{"dirty", "busy", "lock-error", "pending", "changed-source", "dirty-during-download"} {
		t.Run(kind, func(t *testing.T) {
			app, services, calls := installFixture(t)
			req := StartUpdateRequest{ExpectedVersion: "1.10.0"}
			switch kind {
			case "dirty":
				app.dirty = true
			case "busy":
				app.operationBusy.Store(true)
			case "lock-error":
				app.updateLockError = errors.New("lock unavailable")
			case "pending":
				services.pending = func(string) ([]string, error) { return []string{"unfinished"}, nil }
			case "changed-source":
				req.ExpectedCommit = "other"
			case "dirty-during-download":
				original := services.prepare
				services.prepare = func(ctx context.Context, o updates.InstallOptions) (updates.PreparedInstall, error) {
					result, err := original(ctx, o)
					app.dirty = true
					return result, err
				}
			}
			if _, err := app.startUpdate(req, services); err == nil || app.updateClosing.Load() {
				t.Fatal("guard did not block")
			}
			for _, call := range *calls {
				if call != "prepare" {
					t.Fatal(*calls)
				}
			}
		})
	}
}
