//go:build !(windows || linux || darwin || freebsd || netbsd || openbsd || dragonfly)

package updatelock

import (
	"fmt"
	"os"
)

func isReparse(info os.FileInfo) bool { return info.Mode()&os.ModeSymlink != 0 }
func openExclusive(string) (*os.File, error) {
	return nil, fmt.Errorf("editor update locking is unsupported on this platform")
}
