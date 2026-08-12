package updates

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

const fixtureCommit = "1111111111111111111111111111111111111111"
const fixtureTree = "2222222222222222222222222222222222222222"

func fixturePE(label string) []byte {
	raw := make([]byte, 128)
	copy(raw, "MZ")
	raw[60] = 64
	copy(raw[64:], "PE\x00\x00")
	copy(raw[72:], label)
	return raw
}

func putInstallFile(t *testing.T, root, name string, raw []byte) {
	t.Helper()
	file := filepath.Join(root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(file), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, raw, 0644); err != nil {
		t.Fatal(err)
	}
}

func getInstallFile(t *testing.T, root, name string) []byte {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(name)))
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func installFixture(t *testing.T) (string, map[string][]byte) {
	t.Helper()
	root := metadataFixture(t, "1.9.0")
	local := map[string][]byte{
		ExecutablePath:                       fixturePE("old executable"),
		"index.html":                         []byte("locally edited homepage"),
		"app/code.js":                        []byte("locally edited code"),
		"app/custom.js":                      []byte("unmanaged personal code"),
		"data/version.json":                  []byte("intentionally invalid personal data"),
		"data/profile.json":                  []byte("private profile"),
		"exports/publication-citations.json": []byte("saved citations"),
		"app/generated/legacy.txt":           []byte("legacy generated content"),
		"CNAME":                              []byte("personal.example"),
		".git/config":                        []byte("fork origin is never read"),
		"editor/runtime/cache.txt":           []byte("local runtime"),
	}
	for name, raw := range local {
		putInstallFile(t, root, name, raw)
	}
	local["app/release.json"] = getInstallFile(t, root, "app/release.json")
	return root, local
}

type installRemote struct {
	files     map[string][]byte
	tree      []map[string]any
	log       *requestLog
	badBody   string
	missing   string
	truncated bool
	server    *httptest.Server
}

func remoteInstaller(t *testing.T, version string) *installRemote {
	t.Helper()
	remote := &installRemote{files: map[string][]byte{
		"app/release.json":                   []byte(encoded(map[string]any{"version": version, "data_version": "99.0.0", "upstream": Source{Repository: DefaultRepository, Ref: DefaultRef}})),
		ExecutablePath:                       fixturePE("new executable " + version),
		"app/code.js":                        []byte("updated application code"),
		"index.html":                         []byte("updated homepage"),
		"app/new.js":                         []byte("new framework file"),
		"data/profile.json":                  []byte("original author's personal data"),
		"exports/publication-citations.json": []byte("original author's citations"),
		"CNAME":                              []byte("original.example"),
		".git/config":                        []byte("original repository config"),
		"app/generated/legacy.txt":           []byte("original generated content"),
		"editor/runtime/cache.txt":           []byte("original runtime"),
		"personal-photo.jpg":                 []byte("large personal media must not be fetched"),
	}, log: &requestLog{}}
	for name, raw := range remote.files {
		remote.tree = append(remote.tree, map[string]any{"path": name, "type": "blob", "mode": "100644", "sha": gitBlobSHA(raw), "size": len(raw)})
	}
	sort.Slice(remote.tree, func(i, j int) bool { return remote.tree[i]["path"].(string) < remote.tree[j]["path"].(string) })
	remote.server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		remote.log.add(r.URL.Path)
		if r.Method != http.MethodGet {
			t.Errorf("unexpected update method %s", r.Method)
			w.WriteHeader(405)
			return
		}
		switch r.URL.Path {
		case "/repos/" + DefaultRepository + "/commits/" + DefaultRef:
			fmt.Fprint(w, encoded(map[string]any{"sha": fixtureCommit, "commit": map[string]any{"tree": map[string]any{"sha": fixtureTree}}}))
			return
		case "/repos/" + DefaultRepository + "/git/trees/" + fixtureTree:
			if r.URL.Query().Get("recursive") != "1" {
				t.Error("missing recursive tree request")
			}
			fmt.Fprint(w, encoded(map[string]any{"sha": fixtureTree, "tree": remote.tree, "truncated": remote.truncated}))
			return
		}
		prefix := "/" + DefaultRepository + "/" + fixtureCommit + "/"
		if strings.HasPrefix(r.URL.Path, prefix) {
			name := strings.TrimPrefix(r.URL.Path, prefix)
			if name == remote.missing {
				w.WriteHeader(404)
				return
			}
			if raw, ok := remote.files[name]; ok {
				if name == remote.badBody {
					_, _ = w.Write(bytes.Repeat([]byte("x"), len(raw)))
				} else {
					_, _ = w.Write(raw)
				}
				return
			}
		}
		w.WriteHeader(404)
	}))
	t.Cleanup(remote.server.Close)
	return remote
}

func (remote *installRemote) options(root string) InstallOptions {
	return InstallOptions{Root: root, TargetVersion: "1.10.0", Client: remote.server.Client(), APIBaseURL: remote.server.URL, RawBaseURL: remote.server.URL}
}

func assertLocalFiles(t *testing.T, root string, want map[string][]byte) {
	t.Helper()
	for name, raw := range want {
		if got := getInstallFile(t, root, name); !bytes.Equal(got, raw) {
			t.Errorf("local file changed: %s", name)
		}
	}
}

func TestInstallSelectivePrepareApplyAndFinalize(t *testing.T) {
	root, local := installFixture(t)
	// Installation reads no data schema, Wails version, Git state or release-note
	// completeness. Even a broad local policy cannot override hard protection.
	putInstallFile(t, root, "app/update-policy.json", []byte(`{"version":"future","default":"replace","replace":["app/","editor/","data/","exports/",".git/","CNAME","Profile-Editor.exe","index.html"],"regenerate":[]}`))
	remote := remoteInstaller(t, "1.10.0")
	progress := []InstallProgress{}
	options := remote.options(root)
	options.Progress = func(item InstallProgress) { progress = append(progress, item) }
	prepared, err := PrepareInstall(context.Background(), options)
	if err != nil {
		t.Fatal(err)
	}
	if prepared.Journal.State != StatePrepared || prepared.Journal.Commit != fixtureCommit {
		t.Fatalf("unexpected journal: %+v", prepared.Journal)
	}
	assertLocalFiles(t, root, local)
	if len(progress) == 0 || progress[len(progress)-1].Phase != "prepared" {
		t.Fatal("missing preparation progress")
	}
	for _, request := range remote.log.all() {
		if strings.Contains(request, "/data/") || strings.Contains(request, "/exports/") || strings.Contains(request, "/runtime/") || strings.Contains(request, "/app/generated/") || strings.HasSuffix(request, "/CNAME") || strings.Contains(request, "/.git/") || strings.HasSuffix(request, ".jpg") {
			t.Errorf("downloaded personal or generated file: %s", request)
		}
		if strings.Contains(request, "/raw/") || strings.Contains(request, "/main/app/") {
			t.Errorf("download was not commit-pinned: %s", request)
		}
	}
	journal, err := ApplyPrepared(root, prepared.JournalPath)
	if err != nil {
		t.Fatal(err)
	}
	if journal.State != StatePendingExecutable {
		t.Fatalf("state=%s", journal.State)
	}
	for _, file := range journal.Files {
		if !file.BackupReady {
			t.Errorf("backup not ready: %s", file.Path)
		}
		if file.Path == ExecutablePath && (!file.Existed || file.BackupSHA256 != payloadSHA(local[ExecutablePath]) || file.Applied) {
			t.Errorf("deferred executable was not backed up: %+v", file)
		}
	}
	for _, name := range []string{"app/code.js", "index.html", "app/new.js", "app/release.json"} {
		if !bytes.Equal(getInstallFile(t, root, name), remote.files[name]) {
			t.Errorf("not applied: %s", name)
		}
	}
	for _, name := range []string{ExecutablePath, "data/version.json", "data/profile.json", "exports/publication-citations.json", "CNAME", ".git/config", "editor/runtime/cache.txt", "app/generated/legacy.txt", "app/custom.js"} {
		if !bytes.Equal(getInstallFile(t, root, name), local[name]) {
			t.Errorf("preserved file changed: %s", name)
		}
	}
	if _, err := VerifyAndFinalize(root, prepared.JournalPath); err == nil {
		t.Fatal("finalized before executable handoff")
	}
	if pending, err := PendingJournals(root); err != nil || !reflect.DeepEqual(pending, []string{prepared.JournalPath}) {
		t.Fatalf("pending=%v err=%v", pending, err)
	}
	// Simulate the external helper only in this temporary fixture.
	putInstallFile(t, root, ExecutablePath, remote.files[ExecutablePath])
	putInstallFile(t, root, "app/code.js", []byte("permitted edit after application"))
	journal, err = VerifyAndFinalize(root, prepared.JournalPath)
	if err != nil || journal.State != StateCompleted {
		t.Fatalf("finalize state=%s err=%v", journal.State, err)
	}
	installed := readInstalled(root)
	if installed.Version != "1.10.0" || installed.Commit != fixtureCommit || len(installed.Files) != 5 {
		t.Fatalf("bad managed inventory: %+v", installed)
	}
	if pending, err := PendingJournals(root); err != nil || len(pending) != 0 {
		t.Fatalf("completed transaction remained pending: %v %v", pending, err)
	}
}

func TestInstallRollbackRestoresActualLocalChangesAndManagedDeletes(t *testing.T) {
	root, local := installFixture(t)
	putInstallFile(t, root, "app/obsolete.js", []byte("locally edited retired code"))
	local["app/obsolete.js"] = getInstallFile(t, root, "app/obsolete.js")
	installed := installedManifest{Version: "1.9.0", Commit: fixtureCommit, Files: []managedFile{{Path: "app/obsolete.js"}, {Path: "data/profile.json"}, {Path: "app/generated/legacy.txt"}}}
	raw, _ := json.Marshal(installed)
	putInstallFile(t, root, "editor/runtime/updates/installed.json", raw)
	remote := remoteInstaller(t, "1.10.0")
	prepared, err := PrepareInstall(context.Background(), remote.options(root))
	if err != nil {
		t.Fatal(err)
	}
	journal, err := ApplyPrepared(root, prepared.JournalPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "app", "obsolete.js")); !os.IsNotExist(err) {
		t.Fatalf("managed obsolete file remained: %v", err)
	}
	putInstallFile(t, root, ExecutablePath, remote.files[ExecutablePath])
	putInstallFile(t, root, "app/code.js", []byte("edit made after failed update"))
	journal, err = Rollback(root, prepared.JournalPath)
	if err != nil || journal.State != StateRolledBack {
		t.Fatalf("rollback state=%s err=%v", journal.State, err)
	}
	assertLocalFiles(t, root, local)
	if _, err := os.Stat(filepath.Join(root, "app", "new.js")); !os.IsNotExist(err) {
		t.Fatalf("new file survived rollback: %v", err)
	}
	if got := getInstallFile(t, filepath.Dir(prepared.JournalPath), "rollback-current/app/code.js"); string(got) != "edit made after failed update" {
		t.Fatalf("post-update edit was not retained: %q", got)
	}
	if old := readInstalled(root); old.Version != "1.9.0" {
		t.Fatal("rollback altered previous inventory")
	}
}

func TestInstallFailuresDoNotModifyCode(t *testing.T) {
	cases := []struct {
		name   string
		change func(*installRemote, *InstallOptions)
	}{
		{"changed version", func(r *installRemote, o *InstallOptions) { o.TargetVersion = "1.9.9" }},
		{"404", func(r *installRemote, o *InstallOptions) { r.missing = "app/code.js" }},
		{"wrong blob", func(r *installRemote, o *InstallOptions) { r.badBody = "app/code.js" }},
		{"truncated tree", func(r *installRemote, o *InstallOptions) { r.truncated = true }},
		{"symlink entry", func(r *installRemote, o *InstallOptions) {
			r.tree = append(r.tree, map[string]any{"path": "app/link.js", "type": "blob", "mode": "120000", "sha": fixtureCommit, "size": 4})
		}},
		{"oversized file", func(r *installRemote, o *InstallOptions) {
			r.tree = append(r.tree, map[string]any{"path": "app/large.bin", "type": "blob", "mode": "100644", "sha": fixtureCommit, "size": MaxInstallFileBytes + 1})
		}},
		{"case collision", func(r *installRemote, o *InstallOptions) {
			r.tree = append(r.tree, map[string]any{"path": "app/CODE.js", "type": "blob", "mode": "100644", "sha": fixtureCommit, "size": 4})
		}},
		{"invalid executable", func(r *installRemote, o *InstallOptions) {
			r.files[ExecutablePath] = []byte("not PE")
			for _, item := range r.tree {
				if item["path"] == ExecutablePath {
					item["sha"] = gitBlobSHA(r.files[ExecutablePath])
					item["size"] = len(r.files[ExecutablePath])
				}
			}
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root, local := installFixture(t)
			remote := remoteInstaller(t, "1.10.0")
			options := remote.options(root)
			tc.change(remote, &options)
			if _, err := PrepareInstall(context.Background(), options); err == nil {
				t.Fatal("unsafe/incomplete update was prepared")
			}
			assertLocalFiles(t, root, local)
			if pending, err := PendingJournals(root); err != nil || len(pending) != 0 {
				t.Fatalf("download failure became recovery transaction: %v %v", pending, err)
			}
		})
	}
}

func TestInstallRejectsTamperedPayloadBeforeAnyWrite(t *testing.T) {
	root, local := installFixture(t)
	remote := remoteInstaller(t, "1.10.0")
	prepared, err := PrepareInstall(context.Background(), remote.options(root))
	if err != nil {
		t.Fatal(err)
	}
	putInstallFile(t, filepath.Dir(prepared.JournalPath), "payload/app/code.js", []byte("tampered stage"))
	if _, err := ApplyPrepared(root, prepared.JournalPath); err == nil {
		t.Fatal("applied modified payload")
	}
	assertLocalFiles(t, root, local)
}

func TestInstallRollbackChecksAllBackupsBeforeRestoring(t *testing.T) {
	root, local := installFixture(t)
	remote := remoteInstaller(t, "1.10.0")
	prepared, err := PrepareInstall(context.Background(), remote.options(root))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ApplyPrepared(root, prepared.JournalPath); err != nil {
		t.Fatal(err)
	}
	putInstallFile(t, filepath.Dir(prepared.JournalPath), "backups/app/code.js", []byte("damaged backup"))
	journal, err := Rollback(root, prepared.JournalPath)
	if err == nil || journal.State != StateRollbackFailed {
		t.Fatalf("damaged backup accepted: %s %v", journal.State, err)
	}
	if got := getInstallFile(t, root, "index.html"); !bytes.Equal(got, remote.files["index.html"]) {
		t.Fatal("rollback partially restored before validating every backup")
	}
	if pending, err := PendingJournals(root); err != nil || len(pending) != 1 {
		t.Fatalf("failed rollback lost recovery marker: %v %v", pending, err)
	}
	putInstallFile(t, filepath.Dir(prepared.JournalPath), "backups/app/code.js", local["app/code.js"])
	journal, err = Rollback(root, prepared.JournalPath)
	if err != nil || journal.State != StateRolledBack {
		t.Fatalf("recoverable rollback could not resume: %s %v", journal.State, err)
	}
	assertLocalFiles(t, root, local)
}

func TestInstallInterruptedBackupRefreshRetainsJournalChecksum(t *testing.T) {
	root, local := installFixture(t)
	remote := remoteInstaller(t, "1.10.0")
	prepared, err := PrepareInstall(context.Background(), remote.options(root))
	if err != nil {
		t.Fatal(err)
	}
	journal, err := ApplyPrepared(root, prepared.JournalPath)
	if err != nil {
		t.Fatal(err)
	}
	putInstallFile(t, root, "app/code.js", []byte("concurrent local edit"))
	for index := range journal.Files {
		if journal.Files[index].Path == "app/code.js" {
			if err := backupFile(root, prepared.JournalPath, &journal.Files[index]); err != nil {
				t.Fatal(err)
			}
		}
	}
	// Simulate a stop after refreshing a backup and before saving its new hash.
	journal, err = Rollback(root, prepared.JournalPath)
	if err != nil || journal.State != StateRolledBack {
		t.Fatalf("interrupted backup refresh lost recovery: %s %v", journal.State, err)
	}
	assertLocalFiles(t, root, local)
	if got := getInstallFile(t, filepath.Dir(prepared.JournalPath), "backups/app/code.js"); string(got) != "concurrent local edit" {
		t.Fatal("concurrent local edit was not retained")
	}
}

func TestInstallVersionComparisonAndCancelledRequest(t *testing.T) {
	for _, version := range []string{"1.9.0", "1.8.9", "2.0.0"} {
		t.Run(version, func(t *testing.T) {
			root, local := installFixture(t)
			remote := remoteInstaller(t, version)
			options := remote.options(root)
			options.TargetVersion = version
			_, err := PrepareInstall(context.Background(), options)
			if (version == "2.0.0") != (err == nil) {
				t.Fatalf("current=1.9.0 target=%s err=%v", version, err)
			}
			assertLocalFiles(t, root, local)
		})
	}
	root, local := installFixture(t)
	remote := remoteInstaller(t, "1.10.0")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := PrepareInstall(ctx, remote.options(root)); err == nil {
		t.Fatal("cancelled request prepared an update")
	}
	assertLocalFiles(t, root, local)
}

func TestInstallIgnoresOrphanPreparationButRetainsRecoveryEvidence(t *testing.T) {
	root, _ := installFixture(t)
	id := "20261028T010203Z-123456abcdef"
	putInstallFile(t, root, "editor/runtime/updates/"+id+"/payload/app/code.js", []byte("downloaded before journal creation"))
	if pending, err := PendingJournals(root); err != nil || len(pending) != 0 {
		t.Fatalf("orphan preparation blocked updates: %v %v", pending, err)
	}
	remote := remoteInstaller(t, "1.10.0")
	if _, err := PrepareInstall(context.Background(), remote.options(root)); err != nil {
		t.Fatalf("could not prepare after orphan staging: %v", err)
	}
	putInstallFile(t, root, "editor/runtime/updates/"+id+"/backups/app/code.js", []byte("evidence of application"))
	if _, err := PendingJournals(root); err == nil {
		t.Fatal("missing journal with recovery evidence was ignored")
	}
}

func TestInstallBackupFailureAutomaticallyRollsBack(t *testing.T) {
	root, local := installFixture(t)
	remote := remoteInstaller(t, "1.10.0")
	prepared, err := PrepareInstall(context.Background(), remote.options(root))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "app", "new.js"), 0755); err != nil {
		t.Fatal(err)
	}
	journal, err := ApplyPrepared(root, prepared.JournalPath)
	if err == nil || journal.State != StateRolledBack {
		t.Fatalf("backup failure was not rolled back: %s %v", journal.State, err)
	}
	assertLocalFiles(t, root, local)
}

func TestInstallRejectsLinkEscapesAndJournalTraversal(t *testing.T) {
	root, _ := installFixture(t)
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, "app", "linked")); err != nil {
		t.Logf("symlink unavailable: %v", err)
	} else if _, err := safePath(root, "app/linked/code.js"); err == nil {
		t.Fatal("followed symlink outside repository")
	}
	for _, name := range []string{"../outside", "app/../data/profile.json", "C:/Windows/file", "app/code.js:stream", "app\\code.js", "app/CON.txt"} {
		if _, err := safePath(root, name); err == nil {
			t.Errorf("unsafe update path accepted: %s", name)
		}
	}
	if _, err := LoadJournal(root, filepath.Join(outside, "journal.json")); err == nil {
		t.Fatal("loaded journal outside protected transaction directory")
	}
	policy := defaultInstallPolicy()
	for _, name := range []string{"data/profile.json", "DATA/profile.json", "exports/result", "CNAME", "cname", ".git/config", "app/sub/.git/config", "app/generated/result", "editor/runtime/cache"} {
		if replaceAllowed(policy, name) {
			t.Errorf("protected update target accepted: %s", name)
		}
	}
}
