//go:build windows

package updatelock

import (
	"errors"
	"os"
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

func openExclusive(path string) (*os.File, error) {
	name, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}
	handle, err := windows.CreateFile(name, windows.GENERIC_READ|windows.GENERIC_WRITE, 0, nil, windows.OPEN_ALWAYS, windows.FILE_ATTRIBUTE_NORMAL|windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
	if errors.Is(err, windows.ERROR_SHARING_VIOLATION) || errors.Is(err, windows.ERROR_LOCK_VIOLATION) {
		return nil, ErrBusy
	}
	if err != nil {
		return nil, err
	}
	var info windows.ByHandleFileInformation
	if err = windows.GetFileInformationByHandle(handle, &info); err != nil || info.FileAttributes&(windows.FILE_ATTRIBUTE_REPARSE_POINT|windows.FILE_ATTRIBUTE_DIRECTORY) != 0 {
		windows.CloseHandle(handle)
		if err == nil {
			err = windows.ERROR_INVALID_DATA
		}
		return nil, err
	}
	return os.NewFile(uintptr(handle), path), nil
}
