//go:build !windows

package selfupdate

import (
	"fmt"
	"os"
)

func isReparse(info os.FileInfo) bool { return info.Mode()&os.ModeSymlink != 0 }
func processStarted(pid int) (string, error) {
	return "", fmt.Errorf("editor self-update requires Windows")
}
func startScript(script, handoff, logName, generation string) error {
	return fmt.Errorf("editor self-update requires Windows")
}
