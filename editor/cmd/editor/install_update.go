package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Gonie-Gonie/Gonie-Gonie.github.io/editor/internal/selfupdate"
	"github.com/Gonie-Gonie/Gonie-Gonie.github.io/editor/internal/updates"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type StartUpdateRequest struct {
	OperationID     string `json:"operation_id"`
	ExpectedVersion string `json:"expected_version"`
	ExpectedCommit  string `json:"expected_commit,omitempty"`
}

type StartUpdateResponse struct {
	TargetVersion   string `json:"target_version"`
	TransactionID   string `json:"transaction_id"`
	RestartRequired bool   `json:"restart_required"`
	Message         string `json:"message"`
}

type installServices struct {
	prepare  func(context.Context, updates.InstallOptions) (updates.PreparedInstall, error)
	apply    func(string, string) (updates.Journal, error)
	rollback func(string, string) (updates.Journal, error)
	pending  func(string) ([]string, error)
	launch   func(selfupdate.Request) (selfupdate.Pending, error)
	quit     func(context.Context)
}

func defaultInstallServices() installServices {
	return installServices{
		prepare: updates.PrepareInstall, apply: updates.ApplyPrepared,
		rollback: updates.Rollback, pending: updates.PendingJournals,
		launch: selfupdate.Launch,
		quit: func(ctx context.Context) {
			// Let Wails deliver the method response before closing its webview.
			go func() { time.Sleep(250 * time.Millisecond); runtime.Quit(ctx) }()
		},
	}
}

func (a *App) StartUpdate(request StartUpdateRequest) (StartUpdateResponse, error) {
	return a.startUpdate(request, defaultInstallServices())
}

func (a *App) startUpdate(request StartUpdateRequest, services installServices) (StartUpdateResponse, error) {
	if a.updateClosing.Load() || !a.operationBusy.CompareAndSwap(false, true) {
		return StartUpdateResponse{}, errors.New("다른 작업이 진행 중입니다")
	}
	defer a.operationBusy.Store(false)
	a.mu.Lock()
	root, err := a.updateMetadataRootLocked()
	lockError := a.updateLockError
	dirty := a.dirty || len(a.boardMedia) > 0
	appContext := a.ctx
	a.mu.Unlock()
	if err != nil {
		return StartUpdateResponse{}, err
	}
	if lockError != nil {
		return StartUpdateResponse{}, fmt.Errorf("업데이트 잠금: %w", lockError)
	}
	if dirty {
		return StartUpdateResponse{}, errors.New("변경 사항을 먼저 저장해 주세요")
	}
	if request.ExpectedVersion == "" {
		return StartUpdateResponse{}, errors.New("업데이트를 먼저 확인해 주세요")
	}
	pending, err := services.pending(root)
	if err != nil {
		return StartUpdateResponse{}, err
	}
	if len(pending) > 0 {
		return StartUpdateResponse{}, errors.New("이전 업데이트를 확인하려면 편집기를 다시 실행해 주세요")
	}
	if appContext == nil {
		appContext = context.Background()
	}
	ctx, cancel := context.WithTimeout(appContext, updates.InstallTimeout)
	defer cancel()
	a.reportProgress(request.OperationID, "업데이트 다운로드 중…", 0)
	prepared, err := services.prepare(ctx, updates.InstallOptions{
		Root: root, TargetVersion: request.ExpectedVersion,
		Progress: func(progress updates.InstallProgress) {
			a.reportProgress(request.OperationID, progress.Message, progress.Percent)
		},
	})
	if err != nil {
		return StartUpdateResponse{}, fmt.Errorf("업데이트 다운로드: %w", err)
	}
	if request.ExpectedCommit != "" && request.ExpectedCommit != prepared.Journal.Commit {
		return StartUpdateResponse{}, errors.New("원본이 변경되었습니다. 업데이트를 다시 확인해 주세요")
	}
	a.mu.Lock()
	dirty = a.dirty || len(a.boardMedia) > 0
	a.mu.Unlock()
	if dirty {
		return StartUpdateResponse{}, errors.New("변경 사항을 먼저 저장해 주세요")
	}
	a.reportProgress(request.OperationID, "프로그램 파일 적용 중…", 85)
	journal, err := services.apply(root, prepared.JournalPath)
	if err != nil {
		return StartUpdateResponse{}, fmt.Errorf("업데이트 적용: %w", err)
	}
	handoff, err := executableRequest(root, prepared.JournalPath, journal, false)
	if err == nil {
		_, err = services.launch(handoff)
	}
	if err != nil {
		_, rollbackErr := services.rollback(root, prepared.JournalPath)
		return StartUpdateResponse{}, errors.Join(fmt.Errorf("편집기 교체: %w", err), rollbackErr)
	}
	a.updateClosing.Store(true)
	a.reportProgress(request.OperationID, "편집기를 종료합니다", 98)
	services.quit(appContext)
	return StartUpdateResponse{TargetVersion: journal.TargetVersion, TransactionID: journal.ID, RestartRequired: true, Message: "편집기를 종료합니다"}, nil
}

func executableRequest(root, journalPath string, journal updates.Journal, rollback bool) (selfupdate.Request, error) {
	for _, file := range journal.Files {
		if !file.Deferred {
			continue
		}
		directory, version, digest := "payload", journal.TargetVersion, file.SHA256
		if rollback {
			if !file.Existed || !file.BackupReady {
				return selfupdate.Request{}, errors.New("이전 편집기 백업이 없습니다")
			}
			directory, version, digest = "backups", journal.CurrentVersion, file.BackupSHA256
		}
		transaction := filepath.Dir(journalPath)
		return selfupdate.Request{Root: root, TransactionDir: transaction, StagedExecutable: filepath.Join(transaction, directory, updates.ExecutablePath), ExpectedSHA256: digest, TargetVersion: version, WaitPID: os.Getpid(), Relaunch: true}, nil
	}
	return selfupdate.Request{}, errors.New("업데이트에 편집기 실행 파일이 없습니다")
}
