// Package updatelock prevents concurrent editor sessions from recovering an
// update that another process is still applying. Locks expire with the process.
package updatelock

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

var ErrBusy = errors.New("another editor is running for this repository")

type Lock struct {
	file *os.File
	once sync.Once
	err  error
}

func (lock *Lock) Close() error {
	if lock == nil {
		return nil
	}
	lock.once.Do(func() { lock.err = lock.file.Close() })
	return lock.err
}

func Acquire(root string) (*Lock, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(root)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("invalid editor repository")
	}
	directory := root
	for _, part := range []string{"editor", "runtime", "updates"} {
		directory = filepath.Join(directory, part)
		info, err = os.Lstat(directory)
		if os.IsNotExist(err) {
			if err = os.Mkdir(directory, 0755); err != nil && !os.IsExist(err) {
				return nil, err
			}
			info, err = os.Lstat(directory)
		}
		if err != nil {
			return nil, err
		}
		if !info.IsDir() || isReparse(info) {
			return nil, fmt.Errorf("editor lock path contains a link or is not a directory")
		}
	}
	path := filepath.Join(directory, "process.lock")
	if info, err = os.Lstat(path); err == nil {
		if !info.Mode().IsRegular() || isReparse(info) {
			return nil, fmt.Errorf("invalid editor lock file")
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	file, err := openExclusive(path)
	if err != nil {
		return nil, err
	}
	// Verify parents again after acquiring the handle.
	resolved, err := filepath.EvalSymlinks(directory)
	if err != nil || resolved != directory {
		file.Close()
		return nil, fmt.Errorf("editor lock directory changed")
	}
	return &Lock{file: file}, nil
}
