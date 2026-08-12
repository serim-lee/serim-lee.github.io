package updatelock

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestExclusiveLifetimeAndExistingFilePreservation(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "editor", "runtime", "updates", "process.lock")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("existing lock metadata"), 0600); err != nil {
		t.Fatal(err)
	}
	lock, err := Acquire(root)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { lock.Close() })
	if other, err := Acquire(root); !errors.Is(err, ErrBusy) {
		if other != nil {
			other.Close()
		}
		t.Fatalf("second acquisition: %v", err)
	}
	if err := lock.Close(); err != nil {
		t.Fatal(err)
	}
	if err := lock.Close(); err != nil {
		t.Fatal(err)
	}
	other, err := Acquire(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := other.Close(); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil || string(raw) != "existing lock metadata" {
		t.Fatal(string(raw), err)
	}
}

func TestLockRejectsNestedLinkWithoutWritingItsTarget(t *testing.T) {
	root, outside := t.TempDir(), t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "editor"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "editor", "runtime")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if lock, err := Acquire(root); err == nil {
		lock.Close()
		t.Fatal("nested link accepted")
	}
	entries, err := os.ReadDir(outside)
	if err != nil || len(entries) != 0 {
		t.Fatal(entries, err)
	}
}
