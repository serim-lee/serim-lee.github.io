package selfupdate

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func put(t *testing.T, name string, data []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(name), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(name, data, 0600); err != nil {
		t.Fatal(err)
	}
}
func fixture(t *testing.T) Request {
	t.Helper()
	root := t.TempDir()
	tx := filepath.Join(root, "editor", "runtime", "updates", "fixture")
	staged := filepath.Join(tx, "payload", "Profile-Editor.exe")
	put(t, staged, []byte("new editor"))
	put(t, filepath.Join(root, "Profile-Editor.exe"), []byte("old editor"))
	put(t, filepath.Join(root, "data", "profile.json"), []byte("personal data"))
	sum := sha256.Sum256([]byte("new editor"))
	return Request{Root: root, TransactionDir: tx, StagedExecutable: staged, ExpectedSHA256: hex.EncodeToString(sum[:]), TargetVersion: "1.1.0", WaitPID: 12345, WaitStarted: "1"}
}
func assertBytes(t *testing.T, name, want string) {
	t.Helper()
	raw, err := os.ReadFile(name)
	if err != nil || string(raw) != want {
		t.Fatalf("%s: got %q, %v; want %q", name, raw, err, want)
	}
}

func TestPreparePreservesFilesAndResetsReverseHandoff(t *testing.T) {
	request := fixture(t)
	first, err := Prepare(request)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Prepare(request)
	if err != nil {
		t.Fatal(err)
	}
	if first.ID == second.ID || second.Launches != 2 {
		t.Fatalf("retry did not get separate generation: %+v", second)
	}
	assertBytes(t, filepath.Join(request.Root, "Profile-Editor.exe"), "old editor")
	assertBytes(t, filepath.Join(request.Root, "data", "profile.json"), "personal data")
	request.StagedExecutable = filepath.Join(request.TransactionDir, "backups", "Profile-Editor.exe")
	put(t, request.StagedExecutable, []byte("old editor"))
	request.ExpectedSHA256, _ = hashFile(request.StagedExecutable)
	request.TargetVersion = "1.0.0"
	reverse, err := Prepare(request)
	if err != nil {
		t.Fatal(err)
	}
	if reverse.Launches != 1 || reverse.ID == second.ID {
		t.Fatalf("reverse handoff was not reset: %+v", reverse)
	}
}

func TestPrepareRejectsUnsafeOrUnverifiedPaths(t *testing.T) {
	for _, kind := range []string{"outside", "nested", "digest", "directory", "symlink", "helper-symlink"} {
		t.Run(kind, func(t *testing.T) {
			request := fixture(t)
			switch kind {
			case "outside":
				request.StagedExecutable = filepath.Join(t.TempDir(), "Profile-Editor.exe")
			case "nested":
				request.TransactionDir = filepath.Join(request.TransactionDir, "nested")
			case "digest":
				request.ExpectedSHA256 = string(make([]byte, 64))
			case "directory":
				if err := os.Remove(request.StagedExecutable); err != nil {
					t.Fatal(err)
				}
				if err := os.Mkdir(request.StagedExecutable, 0700); err != nil {
					t.Fatal(err)
				}
			case "symlink":
				if err := os.Remove(request.StagedExecutable); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(filepath.Join(request.Root, "Profile-Editor.exe"), request.StagedExecutable); err != nil {
					t.Skipf("symlinks unavailable: %v", err)
				}
			case "helper-symlink":
				dir := filepath.Join(request.TransactionDir, "self-update")
				if err := os.MkdirAll(dir, 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(filepath.Join(request.Root, "Profile-Editor.exe"), filepath.Join(dir, "apply-editor-update.ps1")); err != nil {
					t.Skipf("symlinks unavailable: %v", err)
				}
			}
			if _, err := Prepare(request); err == nil {
				t.Fatal("unsafe request accepted")
			}
			assertBytes(t, filepath.Join(request.Root, "Profile-Editor.exe"), "old editor")
		})
	}
}

func TestRecoveryVerifiesCompiledVersionAndStopsRetries(t *testing.T) {
	request := fixture(t)
	pending, err := Prepare(request)
	if err != nil {
		t.Fatal(err)
	}
	pending.Launches = 3
	if err := writeJSON(filepath.Join(request.TransactionDir, "self-update", "handoff.json"), pending); err != nil {
		t.Fatal(err)
	}
	put(t, filepath.Join(request.Root, "Profile-Editor.exe"), []byte("new editor"))
	got, err := RecoverTransaction(request.Root, request.TransactionDir, filepath.Join(request.Root, "Profile-Editor.exe"), "1.0.0")
	if err != nil || got.Action != ActionFailed {
		t.Fatalf("wrong compiled version: %+v, %v", got, err)
	}
	got, err = RecoverTransaction(request.Root, request.TransactionDir, filepath.Join(request.Root, "Profile-Editor.exe"), "1.1.0")
	if err != nil || got.Action != ActionComplete {
		t.Fatalf("updated executable: %+v, %v", got, err)
	}
	if err := Finalize(request.Root, request.TransactionDir, "complete", ""); err != nil {
		t.Fatal(err)
	}
	got, err = RecoverTransaction(request.Root, request.TransactionDir, filepath.Join(request.Root, "Profile-Editor.exe"), "1.1.0")
	if err != nil || got.Action != ActionNone {
		t.Fatalf("finalized executable: %+v, %v", got, err)
	}
}

func TestRecoveryIgnoresPreviousGenerationResult(t *testing.T) {
	request := fixture(t)
	first, err := Prepare(request)
	if err != nil {
		t.Fatal(err)
	}
	if err := writeJSON(filepath.Join(request.TransactionDir, "self-update", "result.json"), Result{ID: first.ID, Status: "failed", Finalized: true}); err != nil {
		t.Fatal(err)
	}
	pending, err := Prepare(request)
	if err != nil {
		t.Fatal(err)
	}
	pending.Launches = 3
	if err := writeJSON(filepath.Join(request.TransactionDir, "self-update", "handoff.json"), pending); err != nil {
		t.Fatal(err)
	}
	got, err := RecoverTransaction(request.Root, request.TransactionDir, filepath.Join(request.Root, "Profile-Editor.exe"), "1.0.0")
	if err != nil || got.Action != ActionFailed {
		t.Fatalf("stale terminal result hid active handoff: %+v, %v", got, err)
	}
}

func TestPrepareLimitsSameArtifactLaunches(t *testing.T) {
	request := fixture(t)
	for i := 0; i < 3; i++ {
		if _, err := Prepare(request); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := Prepare(request); err == nil {
		t.Fatal("unbounded helper restarts accepted")
	}
	request.StagedExecutable = filepath.Join(request.TransactionDir, "backups", "Profile-Editor.exe")
	put(t, request.StagedExecutable, []byte("old editor"))
	request.ExpectedSHA256, _ = hashFile(request.StagedExecutable)
	request.TargetVersion = "1.0.0"
	if pending, err := Prepare(request); err != nil || pending.Launches != 1 {
		t.Fatalf("rollback must have its own retry count: %+v, %v", pending, err)
	}
}
