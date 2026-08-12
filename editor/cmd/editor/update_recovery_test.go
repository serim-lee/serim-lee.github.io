package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/Gonie-Gonie/Gonie-Gonie.github.io/editor/internal/selfupdate"
	"github.com/Gonie-Gonie/Gonie-Gonie.github.io/editor/internal/updates"
)

func recoveryFixture(t *testing.T) (string, string, updates.Journal, recoveryServices, *[]string) {
	t.Helper()
	root := t.TempDir()
	executable := filepath.Join(root, updates.ExecutablePath)
	if err := os.WriteFile(executable, []byte("new editor"), 0600); err != nil {
		t.Fatal(err)
	}
	newHash := sha256.Sum256([]byte("new editor"))
	oldHash := sha256.Sum256([]byte("old editor"))
	journal := updates.Journal{ID: "fixture", State: updates.StatePendingExecutable, CurrentVersion: "1.0.0", TargetVersion: "1.1.0", Files: []updates.InstallFile{{Path: updates.ExecutablePath, Deferred: true, SHA256: hex.EncodeToString(newHash[:]), Existed: true, BackupReady: true, BackupSHA256: hex.EncodeToString(oldHash[:])}}}
	calls := []string{}
	services := recoveryServices{
		pending: func(string) ([]string, error) {
			return []string{filepath.Join(root, "editor", "runtime", "updates", "fixture", "journal.json")}, nil
		},
		load:   func(string, string) (updates.Journal, error) { return journal, nil },
		verify: func(string, string) (updates.Journal, error) { calls = append(calls, "verify"); return journal, nil },
		rollback: func(string, string) (updates.Journal, error) {
			calls = append(calls, "rollback")
			j := journal
			j.State = updates.StateRolledBack
			return j, nil
		},
		launch: func(selfupdate.Request) (selfupdate.Pending, error) {
			calls = append(calls, "launch")
			return selfupdate.Pending{}, nil
		},
		recover: func(string, string, string, string) (selfupdate.Recovery, error) {
			calls = append(calls, "recover")
			return selfupdate.Recovery{Action: selfupdate.ActionNone}, nil
		},
		finalize: func(_, _, status, _ string) error { calls = append(calls, "finalize:"+status); return nil },
	}
	return root, executable, journal, services, &calls
}

func TestStartupVerifiesRunningVersionAndExecutableBeforeCompletion(t *testing.T) {
	root, exe, _, services, calls := recoveryFixture(t)
	report := recoverProgramUpdate(root, exe, "1.1.0", services)
	if report.Exit || report.Status != "complete" || !reflect.DeepEqual(*calls, []string{"recover", "verify", "finalize:complete"}) {
		t.Fatal(report, *calls)
	}
	*calls = nil
	report = recoverProgramUpdate(root, exe, "1.0.0", services)
	if !report.Exit || !reflect.DeepEqual(*calls, []string{"recover", "launch"}) {
		t.Fatal("old loaded image reported complete", report, *calls)
	}
	*calls = nil
	if err := os.WriteFile(exe, []byte("changed executable"), 0600); err != nil {
		t.Fatal(err)
	}
	report = recoverProgramUpdate(root, exe, "1.1.0", services)
	if !report.Exit || !reflect.DeepEqual(*calls, []string{"recover", "launch"}) {
		t.Fatal("wrong payload reported complete", report, *calls)
	}
}

func TestStartupRollsBackInterruptedApply(t *testing.T) {
	root, exe, journal, services, calls := recoveryFixture(t)
	journal.State = updates.StateApplying
	services.load = func(string, string) (updates.Journal, error) { return journal, nil }
	report := recoverProgramUpdate(root, exe, "1.0.0", services)
	if report.Exit || report.Status != "rolled_back" || !reflect.DeepEqual(*calls, []string{"rollback", "finalize:rolled_back"}) {
		t.Fatal(report, *calls)
	}
}

func TestStartupFailedHandoffOrVerificationRollsBack(t *testing.T) {
	for _, kind := range []string{"handoff", "verification", "apply"} {
		t.Run(kind, func(t *testing.T) {
			root, exe, journal, services, calls := recoveryFixture(t)
			switch kind {
			case "handoff":
				services.recover = func(string, string, string, string) (selfupdate.Recovery, error) {
					return selfupdate.Recovery{Action: selfupdate.ActionFailed, Message: "helper failed"}, nil
				}
			case "verification":
				services.verify = func(string, string) (updates.Journal, error) { return journal, errors.New("missing program file") }
			case "apply":
				journal.State = updates.StateApplying
				services.load = func(string, string) (updates.Journal, error) { return journal, nil }
			}
			report := recoverProgramUpdate(root, exe, "1.1.0", services)
			if report.Exit || report.Status != "rolled_back" {
				t.Fatal(report, *calls)
			}
			if len(*calls) < 2 || (*calls)[len(*calls)-2] != "rollback" || (*calls)[len(*calls)-1] != "finalize:rolled_back" {
				t.Fatal(*calls)
			}
		})
	}
}

func TestStartupRestoresOldExecutableBeforeCompletingRollback(t *testing.T) {
	root, exe, journal, services, calls := recoveryFixture(t)
	journal.State = updates.StateRollbackPendingExecutable
	services.load = func(string, string) (updates.Journal, error) { return journal, nil }
	services.launch = func(request selfupdate.Request) (selfupdate.Pending, error) {
		*calls = append(*calls, "launch-old")
		if request.TargetVersion != "1.0.0" || request.ExpectedSHA256 != journal.Files[0].BackupSHA256 || filepath.Base(filepath.Dir(request.StagedExecutable)) != "backups" {
			t.Fatal(request)
		}
		return selfupdate.Pending{}, nil
	}
	report := recoverProgramUpdate(root, exe, "1.1.0", services)
	if !report.Exit || !reflect.DeepEqual(*calls, []string{"launch-old"}) {
		t.Fatal(report, *calls)
	}
	*calls = nil
	if err := os.WriteFile(exe, []byte("old editor"), 0600); err != nil {
		t.Fatal(err)
	}
	report = recoverProgramUpdate(root, exe, "1.0.0", services)
	if report.Exit || report.Status != "rolled_back" || !reflect.DeepEqual(*calls, []string{"verify", "finalize:rolled_back"}) {
		t.Fatal(report, *calls)
	}
}

func TestStartupLeavesFailedRecoveryVisible(t *testing.T) {
	root, exe, _, services, _ := recoveryFixture(t)
	services.pending = func(string) ([]string, error) { return nil, errors.New("bad transaction") }
	report := recoverProgramUpdate(root, exe, "1.1.0", services)
	if report.Status != "failed" || report.Exit || report.Message == "" {
		t.Fatal(report)
	}
	services.pending = func(string) ([]string, error) { return nil, nil }
	if report = recoverProgramUpdate(root, exe, "1.1.0", services); report != (updateRecovery{}) {
		t.Fatal(report)
	}
}
