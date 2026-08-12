package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Gonie-Gonie/Gonie-Gonie.github.io/editor/internal/selfupdate"
	"github.com/Gonie-Gonie/Gonie-Gonie.github.io/editor/internal/updates"
)

type updateRecovery struct {
	Status, Message string
	Exit            bool
}

type recoveryServices struct {
	pending  func(string) ([]string, error)
	load     func(string, string) (updates.Journal, error)
	verify   func(string, string) (updates.Journal, error)
	rollback func(string, string) (updates.Journal, error)
	launch   func(selfupdate.Request) (selfupdate.Pending, error)
	recover  func(string, string, string, string) (selfupdate.Recovery, error)
	finalize func(string, string, string, string) error
}

func defaultRecoveryServices() recoveryServices {
	return recoveryServices{pending: updates.PendingJournals, load: updates.LoadJournal, verify: updates.VerifyAndFinalize, rollback: updates.Rollback, launch: selfupdate.Launch, recover: selfupdate.RecoverTransaction, finalize: selfupdate.Finalize}
}

// This runs before opening the webview. It reads only the transaction and
// program files; personal documents and their format versions are untouched.
func recoverProgramUpdate(root, executable, version string, services recoveryServices) updateRecovery {
	paths, err := services.pending(root)
	if err != nil {
		return recoveryFailure(err)
	}
	report := updateRecovery{}
	for _, path := range paths {
		journal, err := services.load(root, path)
		if err != nil {
			return recoveryFailure(err)
		}
		if journal.State == updates.StateApplying {
			return recoverRollback(root, path, executable, version, errors.New("프로그램 파일 적용이 중단되었습니다"), services)
		}
		if journal.State == updates.StateRollbackFailed {
			return recoverRollback(root, path, executable, version, errors.New(journal.Error), services)
		}
		if journal.State == updates.StateRollbackPendingExecutable {
			return finishRollback(root, path, executable, version, journal, services)
		}
		outcome, err := services.recover(root, filepath.Dir(path), executable, version)
		if err != nil || outcome.Action == selfupdate.ActionFailed {
			if err == nil {
				err = errors.New(outcome.Message)
			}
			return recoverRollback(root, path, executable, version, err, services)
		}
		if outcome.Action == selfupdate.ActionExitForReplace {
			return updateRecovery{Exit: true}
		}
		request, err := executableRequest(root, path, journal, false)
		if err != nil {
			return recoverRollback(root, path, executable, version, err, services)
		}
		if runningEditorMatches(executable, version, request) {
			if _, err = services.verify(root, path); err != nil {
				return recoverRollback(root, path, executable, version, err, services)
			}
			// A crash before creating the handoff has no sidecar to finalize.
			if err = services.finalize(root, filepath.Dir(path), "complete", ""); err != nil && !os.IsNotExist(err) {
				return recoveryFailure(err)
			}
			report = updateRecovery{Status: "complete", Message: journal.TargetVersion + " 업데이트가 완료되었습니다."}
			continue
		}
		// Includes a crash after applying code but before starting the helper.
		if _, err = services.launch(request); err != nil {
			return recoverRollback(root, path, executable, version, err, services)
		}
		return updateRecovery{Exit: true}
	}
	return report
}

func recoveryFailure(err error) updateRecovery {
	return updateRecovery{Status: "failed", Message: "업데이트 복구: " + err.Error()}
}

func recoverRollback(root, path, executable, version string, cause error, services recoveryServices) updateRecovery {
	journal, err := services.rollback(root, path)
	if err != nil {
		return recoveryFailure(errors.Join(cause, err))
	}
	if journal.State == updates.StateRollbackPendingExecutable {
		return finishRollback(root, path, executable, version, journal, services)
	}
	if err := services.finalize(root, filepath.Dir(path), "rolled_back", cause.Error()); err != nil && !os.IsNotExist(err) {
		return recoveryFailure(errors.Join(cause, err))
	}
	return updateRecovery{Status: "rolled_back", Message: "업데이트를 적용하지 못해 이전 프로그램 파일을 복원했습니다. " + cause.Error()}
}

func finishRollback(root, path, executable, version string, journal updates.Journal, services recoveryServices) updateRecovery {
	request, err := executableRequest(root, path, journal, true)
	if err != nil {
		return recoveryFailure(err)
	}
	if runningEditorMatches(executable, version, request) {
		if _, err = services.verify(root, path); err != nil {
			return recoveryFailure(err)
		}
		if err = services.finalize(root, filepath.Dir(path), "rolled_back", journal.Error); err != nil && !os.IsNotExist(err) {
			return recoveryFailure(err)
		}
		return updateRecovery{Status: "rolled_back", Message: "이전 프로그램 파일을 복원했습니다."}
	}
	if _, err = services.launch(request); err != nil {
		return recoveryFailure(fmt.Errorf("이전 편집기 복원: %w", err))
	}
	return updateRecovery{Exit: true}
}

func runningEditorMatches(executable, version string, request selfupdate.Request) bool {
	absolute, err := filepath.Abs(executable)
	if err != nil || !strings.EqualFold(absolute, filepath.Join(request.Root, updates.ExecutablePath)) || version != request.TargetVersion {
		return false
	}
	file, err := os.Open(absolute)
	if err != nil {
		return false
	}
	defer file.Close()
	digest := sha256.New()
	if _, err = io.Copy(digest, file); err != nil {
		return false
	}
	return hex.EncodeToString(digest.Sum(nil)) == request.ExpectedSHA256
}
