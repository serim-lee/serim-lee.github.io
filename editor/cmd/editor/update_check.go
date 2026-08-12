package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Gonie-Gonie/Gonie-Gonie.github.io/editor/internal/updates"
)

type UpdateInfo = updates.Info
type UpdateCheckResponse = updates.Response

// Update metadata is readable even when personal data needs repair or a newer
// data reader. This path deliberately bypasses rootLocked's data validation.
func (a *App) updateMetadataRootLocked() (string, error) {
	if a.repoRoot != "" {
		return a.repoRoot, nil
	}
	starts := []string{}
	if executable, err := os.Executable(); err == nil {
		starts = append(starts, filepath.Dir(executable))
	}
	if directory, err := os.Getwd(); err == nil {
		starts = append(starts, directory)
	}
	seen := map[string]bool{}
	for _, start := range starts {
		root, err := filepath.Abs(start)
		if err != nil {
			continue
		}
		for {
			key := strings.ToLower(filepath.Clean(root))
			if !seen[key] {
				seen[key] = true
				if info, err := os.Stat(filepath.Join(root, "app", "release.json")); err == nil && info.Mode().IsRegular() {
					a.repoRoot = root
					return root, nil
				}
			}
			parent := filepath.Dir(root)
			if parent == root {
				break
			}
			root = parent
		}
	}
	return "", fmt.Errorf("현재 버전 파일을 찾을 수 없습니다. 저장소의 Profile-Editor.exe를 실행해 주세요")
}

func (a *App) GetUpdateInfo() (UpdateInfo, error) {
	a.mu.Lock()
	root, err := a.updateMetadataRootLocked()
	status, message := a.updateRecoveryStatus, a.updateRecoveryMessage
	a.mu.Unlock()
	if err != nil {
		return UpdateInfo{}, err
	}
	info, err := updates.GetInfo(root)
	info.RecoveryStatus, info.RecoveryMessage = status, message
	return info, err
}

func (a *App) CheckUpdates() (UpdateCheckResponse, error) { return a.checkUpdates(updates.Check) }

func (a *App) checkUpdates(check func(context.Context, updates.Options) (updates.Response, error)) (UpdateCheckResponse, error) {
	a.mu.Lock()
	root, err := a.updateMetadataRootLocked()
	ctx := a.ctx
	a.mu.Unlock()
	if err != nil {
		return UpdateCheckResponse{}, err
	}
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, updates.CheckTimeout)
	defer cancel()
	return check(ctx, updates.Options{Root: root})
}
