package main

import (
	"embed"
	"errors"
	"os"

	"github.com/Gonie-Gonie/Gonie-Gonie.github.io/editor/internal/updatelock"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

// Set by editor/build.bat; recovery compares the running image to its target.
var compiledVersion = "development"

func main() {
	app := NewApp()
	if root, err := app.updateMetadataRootLocked(); err == nil {
		lock, lockErr := updatelock.Acquire(root)
		if errors.Is(lockErr, updatelock.ErrBusy) {
			return
		}
		if lockErr != nil {
			app.updateLockError = lockErr
			app.updateRecoveryStatus, app.updateRecoveryMessage = "failed", "업데이트 잠금: "+lockErr.Error()
		} else {
			// Keep the OS handle through recovery and the complete editor session.
			// A crash releases it, while a second window cannot roll back live work.
			defer lock.Close()
		}
		if executable, err := os.Executable(); err == nil && lockErr == nil {
			recovery := recoverProgramUpdate(root, executable, compiledVersion, defaultRecoveryServices())
			if recovery.Exit {
				return
			}
			app.updateRecoveryStatus, app.updateRecoveryMessage = recovery.Status, recovery.Message
		}
	}
	err := wails.Run(&options.App{
		Title:     "Profile-Editor",
		Width:     1440,
		Height:    900,
		MinWidth:  1280,
		MinHeight: 720,
		AssetServer: &assetserver.Options{
			Assets:  assets,
			Handler: app.mediaPreviewHandler(),
		},
		OnStartup:     app.startup,
		OnShutdown:    app.shutdown,
		OnBeforeClose: app.beforeClose,
		DragAndDrop: &options.DragAndDrop{
			EnableFileDrop: true,
		},
		Bind: []interface{}{
			app,
		},
	})
	if err != nil {
		println("Error:", err.Error())
	}
}
