//go:build windows

package selfupdate

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"golang.org/x/sys/windows"
)

func sleepingProcess(t *testing.T, seconds string) *exec.Cmd {
	t.Helper()
	command := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-WindowStyle", "Hidden", "-Command", "Start-Sleep -Seconds "+seconds)
	command.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if err := command.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = command.Wait() })
	return command
}
func waitHelper(t *testing.T, pending Pending, status string) Result {
	t.Helper()
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		result, err := readResult(pending.TransactionDir)
		if err == nil && result.ID == pending.ID {
			if result.Status == status {
				return result
			}
			if result.Status == "failed" {
				t.Fatalf("helper failed: %s", result.Message)
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	log, _ := os.ReadFile(filepath.Join(pending.TransactionDir, "self-update", "helper.log"))
	result, _ := readResult(pending.TransactionDir)
	t.Fatalf("helper did not reach %s; result=%+v; log=%s", status, result, log)
	return Result{}
}

func TestHelperWaitsForProcessThenReplaces(t *testing.T) {
	request := fixture(t)
	process := sleepingProcess(t, "4")
	request.WaitPID = process.Process.Pid
	request.WaitStarted = ""
	pending, err := Launch(request)
	if err != nil {
		t.Fatal(err)
	}
	waitHelper(t, pending, "waiting")
	assertBytes(t, filepath.Join(request.Root, "Profile-Editor.exe"), "old editor")
	waitHelper(t, pending, "applied")
	assertBytes(t, filepath.Join(request.Root, "Profile-Editor.exe"), "new editor")
	assertBytes(t, filepath.Join(request.TransactionDir, "self-update", "previous.exe"), "old editor")
	assertBytes(t, filepath.Join(request.Root, "data", "profile.json"), "personal data")
}

func TestHelperSkipsReusedProcessID(t *testing.T) {
	request := fixture(t)
	process := sleepingProcess(t, "4")
	request.WaitPID = process.Process.Pid
	request.WaitStarted = "0"
	pending, err := Launch(request)
	if err != nil {
		t.Fatal(err)
	}
	waitHelper(t, pending, "applied")
	assertBytes(t, filepath.Join(request.Root, "Profile-Editor.exe"), "new editor")
}

func TestNewReverseHandoffCancelsWaitingHelper(t *testing.T) {
	request := fixture(t)
	process := sleepingProcess(t, "5")
	request.WaitPID = process.Process.Pid
	request.WaitStarted = ""
	first, err := Launch(request)
	if err != nil {
		t.Fatal(err)
	}
	waitHelper(t, first, "waiting")
	reverse := request
	reverse.StagedExecutable = filepath.Join(request.TransactionDir, "backups", "Profile-Editor.exe")
	put(t, reverse.StagedExecutable, []byte("old editor"))
	reverse.ExpectedSHA256, _ = hashFile(reverse.StagedExecutable)
	reverse.TargetVersion = "1.0.0"
	pending, err := Launch(reverse)
	if err != nil {
		t.Fatal(err)
	}
	waitHelper(t, pending, "applied")
	assertBytes(t, filepath.Join(request.Root, "Profile-Editor.exe"), "old editor")
	assertBytes(t, filepath.Join(request.Root, "data", "profile.json"), "personal data")
	if _, err := os.Stat(filepath.Join(request.TransactionDir, "self-update", "previous.exe")); !os.IsNotExist(err) {
		t.Fatalf("cancelled helper modified executable backup: %v", err)
	}
}

func TestFinalizedRollbackCancelsWaitingHelper(t *testing.T) {
	request := fixture(t)
	process := sleepingProcess(t, "3")
	request.WaitPID = process.Process.Pid
	request.WaitStarted = ""
	pending, err := Launch(request)
	if err != nil {
		t.Fatal(err)
	}
	waitHelper(t, pending, "waiting")
	if err := Finalize(request.Root, request.TransactionDir, "rolled_back", ""); err != nil {
		t.Fatal(err)
	}
	_ = process.Wait()
	time.Sleep(700 * time.Millisecond)
	assertBytes(t, filepath.Join(request.Root, "Profile-Editor.exe"), "old editor")
	result, err := readResult(request.TransactionDir)
	if err != nil || result.Status != "rolled_back" || !result.Finalized {
		t.Fatalf("rollback overwritten: %+v, %v", result, err)
	}
}

func TestHelperRetriesLockedExecutable(t *testing.T) {
	request := fixture(t)
	process := sleepingProcess(t, "1")
	request.WaitPID, request.WaitStarted = process.Process.Pid, "0"
	target := filepath.Join(request.Root, "Profile-Editor.exe")
	name, err := windows.UTF16PtrFromString(target)
	if err != nil {
		t.Fatal(err)
	}
	handle, err := windows.CreateFile(name, windows.GENERIC_READ, 0, nil, windows.OPEN_EXISTING, windows.FILE_ATTRIBUTE_NORMAL, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if handle != windows.InvalidHandle {
			_ = windows.CloseHandle(handle)
		}
	}()
	pending, err := Launch(request)
	if err != nil {
		t.Fatal(err)
	}
	waitHelper(t, pending, "waiting")
	time.Sleep(300 * time.Millisecond)
	if err := windows.CloseHandle(handle); err != nil {
		t.Fatal(err)
	}
	handle = windows.InvalidHandle
	waitHelper(t, pending, "applied")
	assertBytes(t, target, "new editor")
}

func TestHelperRejectsPayloadChangedAfterPreparation(t *testing.T) {
	request := fixture(t)
	pending, err := Prepare(request)
	if err != nil {
		t.Fatal(err)
	}
	put(t, request.StagedExecutable, []byte("changed payload"))
	dir := filepath.Join(request.TransactionDir, "self-update")
	if err := startScript(filepath.Join(dir, "apply-editor-update.ps1"), filepath.Join(dir, "handoff.json"), filepath.Join(dir, "helper.log"), pending.ID); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		result, err := readResult(request.TransactionDir)
		if err == nil && result.ID == pending.ID && result.Status == "failed" {
			assertBytes(t, filepath.Join(request.Root, "Profile-Editor.exe"), "old editor")
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	log, _ := os.ReadFile(filepath.Join(dir, "helper.log"))
	t.Fatalf("modified payload was not rejected: %s", log)
}
