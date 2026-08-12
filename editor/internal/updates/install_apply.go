package updates

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func transactionPath(root, journalPath, kind, name string) (string, error) {
	rel, err := filepath.Rel(root, filepath.Join(filepath.Dir(journalPath), kind, filepath.FromSlash(name)))
	if err != nil {
		return "", err
	}
	return safePath(root, filepath.ToSlash(rel))
}

func validatePayloads(root, journalPath string, journal Journal) error {
	hasExecutable, hasManifest := false, false
	for _, file := range journal.Files {
		if file.Action != "replace" {
			continue
		}
		if !sha256Pattern.MatchString(file.SHA256) || !gitSHAPattern.MatchString(file.GitSHA) {
			return fmt.Errorf("업데이트 검증 정보가 없습니다: %s", file.Path)
		}
		payload, err := transactionPath(root, journalPath, "payload", file.Path)
		if err != nil {
			return err
		}
		checksum, size, err := hashFile(payload)
		if err != nil || checksum != file.SHA256 || size != file.Size {
			return fmt.Errorf("업데이트 파일이 손상되었습니다: %s", file.Path)
		}
		if file.Deferred {
			raw, err := os.ReadFile(payload)
			if err != nil || !validExecutable(raw) {
				return fmt.Errorf("실행 파일 검증에 실패했습니다")
			}
			hasExecutable = true
		}
		if file.Path == "app/release.json" {
			hasManifest = true
		}
	}
	if !hasExecutable || !hasManifest {
		return fmt.Errorf("필수 업데이트 파일이 없습니다")
	}
	return nil
}

func backupFile(root, journalPath string, file *InstallFile) error {
	target, err := safePath(root, file.Path)
	if err != nil {
		return err
	}
	info, err := os.Lstat(target)
	if os.IsNotExist(err) {
		file.Existed = false
		file.BackupSHA256 = ""
		file.BackupReady = true
		return nil
	}
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("업데이트 대상은 일반 파일이어야 합니다: %s", file.Path)
	}
	if info.Size() > MaxInstallFileBytes {
		return fmt.Errorf("백업할 파일이 크기 제한을 초과했습니다: %s", file.Path)
	}
	backup, err := transactionPath(root, journalPath, "backups", file.Path)
	if err != nil {
		return err
	}
	if file.BackupReady && file.Existed {
		// A local edit can arrive between the first backup and replacement.
		// Retain the checksum-addressed previous backup before refreshing it,
		// so a stop before the next journal rename remains recoverable.
		previous, err := verifiedBackup(root, journalPath, *file)
		if err != nil {
			return err
		}
		history, err := transactionPath(root, journalPath, "backup-history", file.BackupSHA256+"/"+file.Path)
		if err != nil {
			return err
		}
		if err := atomicCopy(previous, history, 0600); err != nil {
			return err
		}
	}
	if err := atomicCopy(target, backup, 0600); err != nil {
		return err
	}
	checksum, _, err := hashFile(backup)
	if err != nil {
		return err
	}
	file.Existed = true
	file.BackupSHA256 = checksum
	file.Mode = uint32(info.Mode().Perm())
	file.BackupReady = true
	return nil
}

// ApplyPrepared backs up the actual local code, including uncommitted changes.
// The executable is backed up too but left for the external handoff after exit.
func ApplyPrepared(root, journalPath string) (Journal, error) {
	journal, err := LoadJournal(root, journalPath)
	if err != nil {
		return journal, err
	}
	root = journal.Root
	if journal.State != StatePrepared {
		return journal, fmt.Errorf("준비가 완료된 업데이트가 아닙니다: %s", journal.State)
	}
	if err := validatePayloads(root, journalPath, journal); err != nil {
		return journal, err
	}
	info, err := GetInfo(root)
	if err != nil {
		return journal, err
	}
	if info.CurrentVersion != journal.CurrentVersion || info.SourceRepository != journal.Repository || info.SourceRef != journal.Ref {
		return journal, fmt.Errorf("현재 버전 또는 원본 설정이 달라졌습니다. 다시 확인해 주세요")
	}
	policy := installPolicy(root)
	for _, file := range journal.Files {
		if !replaceAllowed(policy, file.Path) {
			return journal, fmt.Errorf("보존 대상은 업데이트할 수 없습니다: %s", file.Path)
		}
	}
	journal.State = StateApplying
	journal.Error = ""
	if err := SaveJournal(root, journalPath, journal); err != nil {
		return journal, err
	}
	fail := func(problem error) (Journal, error) {
		journal.Error = problem.Error()
		_ = SaveJournal(root, journalPath, journal)
		restored, recoveryErr := Rollback(root, journalPath)
		if recoveryErr != nil {
			return restored, fmt.Errorf("업데이트 실패: %w; 복구 실패: %v", problem, recoveryErr)
		}
		return restored, problem
	}
	for index := range journal.Files {
		if err := backupFile(root, journalPath, &journal.Files[index]); err != nil {
			return fail(err)
		}
		if err := SaveJournal(root, journalPath, journal); err != nil {
			return fail(err)
		}
	}
	for index := range journal.Files {
		file := &journal.Files[index]
		if file.Deferred {
			continue
		}
		target, err := safePath(root, file.Path)
		if err != nil {
			return fail(err)
		}
		currentHash, _, currentErr := hashFile(target)
		if (file.Existed && (currentErr != nil || currentHash != file.BackupSHA256)) || (!file.Existed && !os.IsNotExist(currentErr)) {
			if err := backupFile(root, journalPath, file); err != nil {
				return fail(err)
			}
			if err := SaveJournal(root, journalPath, journal); err != nil {
				return fail(err)
			}
		}
		if file.Action == "delete" {
			if err := os.Remove(target); err != nil && !os.IsNotExist(err) {
				return fail(err)
			}
		} else {
			payload, err := transactionPath(root, journalPath, "payload", file.Path)
			if err != nil {
				return fail(err)
			}
			mode := os.FileMode(file.Mode)
			if mode == 0 {
				mode = 0644
			}
			if err := atomicCopy(payload, target, mode); err != nil {
				return fail(err)
			}
			checksum, _, err := hashFile(target)
			if err != nil || checksum != file.SHA256 {
				return fail(fmt.Errorf("적용한 파일을 확인할 수 없습니다: %s", file.Path))
			}
		}
		file.Applied = true
		if err := SaveJournal(root, journalPath, journal); err != nil {
			return fail(err)
		}
	}
	journal.State = StatePendingExecutable
	if err := SaveJournal(root, journalPath, journal); err != nil {
		return fail(err)
	}
	return journal, nil
}

func verifiedBackup(root, journalPath string, file InstallFile) (string, error) {
	backup, err := transactionPath(root, journalPath, "backups", file.Path)
	if err != nil {
		return "", err
	}
	checksum, _, err := hashFile(backup)
	if err != nil || checksum != file.BackupSHA256 {
		history, historyErr := transactionPath(root, journalPath, "backup-history", file.BackupSHA256+"/"+file.Path)
		if historyErr == nil {
			if historicalSHA, _, e := hashFile(history); e == nil && historicalSHA == file.BackupSHA256 {
				return history, nil
			}
		}
		return "", fmt.Errorf("백업 파일이 손상되었습니다: %s", file.Path)
	}
	return backup, nil
}

func executingTarget(target string) bool {
	current, err := os.Executable()
	if err != nil {
		return false
	}
	if resolved, e := filepath.EvalSymlinks(current); e == nil {
		current = resolved
	}
	return samePath(current, target)
}

// Rollback defers the old executable when the new image is currently running.
// The caller launches the external helper with backups/Profile-Editor.exe.
func Rollback(root, journalPath string) (journal Journal, err error) {
	journal, err = LoadJournal(root, journalPath)
	if err != nil {
		return journal, err
	}
	root = journal.Root
	defer func() {
		if err != nil {
			journal.State = StateRollbackFailed
			journal.Error = err.Error()
			_ = SaveJournal(root, journalPath, journal)
		}
	}()
	for _, file := range journal.Files {
		if file.BackupReady && file.Existed {
			if _, err := verifiedBackup(root, journalPath, file); err != nil {
				return journal, err
			}
		}
	}
	pendingExecutable := false
	for index := len(journal.Files) - 1; index >= 0; index-- {
		file := &journal.Files[index]
		if !file.BackupReady {
			continue
		}
		target, err := safePath(root, file.Path)
		if err != nil {
			return journal, err
		}
		checksum, _, hashErr := hashFile(target)
		if file.Existed && hashErr == nil && checksum == file.BackupSHA256 {
			file.Applied = false
			continue
		}
		if file.Deferred && executingTarget(target) {
			pendingExecutable = true
			continue
		}
		if !os.IsNotExist(hashErr) {
			if hashErr != nil {
				return journal, hashErr
			}
			copyPath, err := transactionPath(root, journalPath, "rollback-current", file.Path)
			if err != nil {
				return journal, err
			}
			if err := atomicCopy(target, copyPath, 0600); err != nil {
				return journal, err
			}
		}
		if file.Existed {
			backup, err := verifiedBackup(root, journalPath, *file)
			if err != nil {
				return journal, err
			}
			if err := atomicCopy(backup, target, os.FileMode(file.Mode)); err != nil {
				if file.Deferred {
					pendingExecutable = true
					continue
				}
				return journal, err
			}
		} else if err := os.Remove(target); err != nil && !os.IsNotExist(err) {
			if file.Deferred {
				pendingExecutable = true
				continue
			}
			return journal, err
		}
		file.Applied = false
	}
	if pendingExecutable {
		journal.State = StateRollbackPendingExecutable
	} else {
		journal.State = StateRolledBack
	}
	if err := SaveJournal(root, journalPath, journal); err != nil {
		return journal, err
	}
	return journal, nil
}

// Startup verification uses the executable's checksum and recorded file
// presence. Later edits of noncritical code are allowed and do not gate data.
func VerifyAndFinalize(root, journalPath string) (Journal, error) {
	journal, err := LoadJournal(root, journalPath)
	if err != nil {
		return journal, err
	}
	root = journal.Root
	if journal.State == StateCompleted || journal.State == StateRolledBack {
		return journal, nil
	}
	rollback := journal.State == StateRollbackPendingExecutable
	if !rollback && journal.State != StatePendingExecutable && journal.State != StateAwaitingVerification {
		return journal, fmt.Errorf("확인할 수 없는 업데이트 상태입니다: %s", journal.State)
	}
	hasExecutable := false
	for _, file := range journal.Files {
		target, err := safePath(root, file.Path)
		if err != nil {
			return journal, err
		}
		shouldExist := file.Action == "replace"
		if rollback {
			shouldExist = file.Existed && file.BackupReady
		}
		info, statErr := os.Lstat(target)
		if !shouldExist {
			if statErr == nil {
				return journal, fmt.Errorf("삭제할 파일이 남아 있습니다: %s", file.Path)
			}
			if !os.IsNotExist(statErr) {
				return journal, statErr
			}
			continue
		}
		if statErr != nil || !info.Mode().IsRegular() {
			return journal, fmt.Errorf("업데이트 파일이 없습니다: %s", file.Path)
		}
		if file.Deferred {
			hasExecutable = true
			expected := file.SHA256
			if rollback {
				expected = file.BackupSHA256
			}
			checksum, _, err := hashFile(target)
			if err != nil || checksum != expected {
				return journal, fmt.Errorf("실행 파일 교체가 완료되지 않았습니다")
			}
		}
	}
	if !hasExecutable {
		return journal, fmt.Errorf("실행 파일 확인 기록이 없습니다")
	}
	if rollback {
		journal.State = StateRolledBack
	} else {
		installed := installedManifest{Version: journal.TargetVersion, Commit: journal.Commit, Files: []managedFile{}}
		for _, file := range journal.Files {
			if file.Action == "replace" {
				installed.Files = append(installed.Files, managedFile{Path: file.Path, SHA256: file.SHA256, Size: file.Size})
			}
		}
		raw, err := json.MarshalIndent(installed, "", "  ")
		if err != nil {
			return journal, err
		}
		ledger, err := safePath(root, "editor/runtime/updates/installed.json")
		if err != nil {
			return journal, err
		}
		if err := atomicWrite(ledger, append(raw, '\n'), 0600); err != nil {
			return journal, err
		}
		journal.State = StateCompleted
		for index := range journal.Files {
			journal.Files[index].Applied = true
		}
	}
	journal.Error = ""
	if err := SaveJournal(root, journalPath, journal); err != nil {
		return journal, err
	}
	return journal, nil
}
