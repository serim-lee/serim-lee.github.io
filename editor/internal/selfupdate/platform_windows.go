//go:build windows

package selfupdate

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"syscall"

	"golang.org/x/sys/windows"
)

func isReparse(info os.FileInfo) bool {
	if info.Mode()&os.ModeSymlink != 0 {
		return true
	}
	if data, ok := info.Sys().(*syscall.Win32FileAttributeData); ok {
		return data.FileAttributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0
	}
	return false
}
func processStarted(pid int) (string, error) {
	handle, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
	if err != nil {
		return "", fmt.Errorf("read editor process: %w", err)
	}
	defer windows.CloseHandle(handle)
	var creation, exit, kernel, user windows.Filetime
	if err := windows.GetProcessTimes(handle, &creation, &exit, &kernel, &user); err != nil {
		return "", err
	}
	ticks := uint64(creation.HighDateTime)<<32 | uint64(creation.LowDateTime)
	return strconv.FormatUint(ticks, 10), nil
}
func startScript(script, handoff, logName, generation string) error {
	log, err := os.OpenFile(logName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer log.Close()
	command := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-WindowStyle", "Hidden", "-File", script, "-Handoff", handoff, "-Generation", generation)
	command.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: windows.CREATE_NEW_PROCESS_GROUP}
	command.Stdout = log
	command.Stderr = log
	if err := command.Start(); err != nil {
		return err
	}
	// Once started, the helper owns the handoff. Reporting a later error would
	// let the caller roll back while a live helper can still replace the editor.
	_ = command.Process.Release()
	return nil
}
