//go:build linux || darwin || freebsd || netbsd || openbsd || dragonfly

package updatelock

import (
	"errors"
	"os"
	"syscall"
)

func isReparse(info os.FileInfo) bool { return info.Mode()&os.ModeSymlink != 0 }
func openExclusive(path string) (*os.File, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	if err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		file.Close()
		if errors.Is(err, syscall.EWOULDBLOCK) || errors.Is(err, syscall.EAGAIN) {
			return nil, ErrBusy
		}
		return nil, err
	}
	return file, nil
}
