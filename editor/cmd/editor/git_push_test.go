package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func gitTestCommand(t *testing.T, root string, args ...string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = root
	cmd.Env = commandEnvironment()
	hideCommandWindow(cmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, output)
	}
	return strings.TrimSpace(string(output))
}

func writeGitFixture(t *testing.T, root, name, text string) {
	t.Helper()
	location := filepath.Join(root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(location), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(location, []byte(text), 0755); err != nil {
		t.Fatal(err)
	}
}

func newGitPushFixture(t *testing.T) (root, remote string, app *App) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("Git is not installed")
	}
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	root = newBoardRepoFixture(t)
	remote = filepath.Join(t.TempDir(), "remote.git")
	gitTestCommand(t, root, "init", "--initial-branch=main")
	gitTestCommand(t, root, "config", "user.name", "Editor Test")
	gitTestCommand(t, root, "config", "user.email", "editor-test@example.invalid")
	gitTestCommand(t, root, "add", "-A")
	gitTestCommand(t, root, "commit", "-m", "unique fixture commit")
	gitTestCommand(t, root, "init", "--bare", "--initial-branch=main", remote)
	gitTestCommand(t, root, "remote", "add", "origin", remote)
	gitTestCommand(t, root, "push", "-u", "origin", "main")
	return root, remote, &App{repoRoot: root, siteBuilder: func(ctx context.Context, root string) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		path := filepath.Join(root, "exports", "export.txt")
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}
		return os.WriteFile(path, []byte("generated export\n"), 0644)
	}}
}

func TestGitPushAmendsSameMessageAndUploadsGeneratedFiles(t *testing.T) {
	root, remote, app := newGitPushFixture(t)
	original := gitTestCommand(t, root, "rev-parse", "HEAD")
	writeGitFixture(t, root, "data/new-item.json", "[]\n")
	var progress []OperationProgress
	app.progressSink = func(p OperationProgress) { progress = append(progress, p) }
	response, err := app.PushGit(GitPushRequest{OperationID: "publish-test"})
	if err != nil {
		t.Fatal(err)
	}
	if response.Commit == original || response.ChangedFiles != 2 {
		t.Fatalf("unexpected response: %#v", response)
	}
	if head := gitTestCommand(t, remote, "rev-parse", "main"); head != response.Commit {
		t.Fatalf("remote %s != saved %s", head, response.Commit)
	}
	if count := gitTestCommand(t, root, "rev-list", "--count", "HEAD"); count != "1" {
		t.Fatalf("amend added history: %s", count)
	}
	if message := gitTestCommand(t, root, "log", "-1", "--format=%s"); message != "unique fixture commit" {
		t.Fatalf("message changed: %s", message)
	}
	if export := gitTestCommand(t, remote, "show", "main:exports/export.txt"); export != "generated export" {
		t.Fatalf("export missing: %q", export)
	}
	if status := gitTestCommand(t, root, "status", "--porcelain"); status != "" {
		t.Fatalf("left changes: %s", status)
	}
	if len(progress) == 0 || progress[len(progress)-1].Percent != 100 || progress[len(progress)-1].Message != "Git push 완료" {
		t.Fatalf("missing completion: %#v", progress)
	}
	for _, p := range progress {
		if p.OperationID != "publish-test" {
			t.Fatal(p)
		}
	}
	second, err := app.PushGit(GitPushRequest{OperationID: "clean-test"})
	if err != nil {
		t.Fatal(err)
	}
	if second.Commit != response.Commit || second.Message != "이미 최신 상태입니다" {
		t.Fatalf("clean push changed history: %#v", second)
	}
}

func TestGitPushFailureKeepsCommitAndRetryUsesOriginalLease(t *testing.T) {
	root, remote, app := newGitPushFixture(t)
	base := gitTestCommand(t, remote, "rev-parse", "main")
	writeGitFixture(t, remote, "hooks/pre-receive", "#!/bin/sh\necho 'fixture rejection' >&2\nexit 1\n")
	writeGitFixture(t, root, "data/new-item.json", "[]\n")
	if _, err := app.PushGit(GitPushRequest{OperationID: "reject"}); err == nil {
		t.Fatal("rejected push unexpectedly succeeded")
	}
	failedCommit := gitTestCommand(t, root, "rev-parse", "HEAD")
	if failedCommit == base {
		t.Fatal("failed push discarded local commit")
	}
	if head := gitTestCommand(t, remote, "rev-parse", "main"); head != base {
		t.Fatal("rejected push changed remote")
	}
	if recorded := gitTestCommand(t, root, "config", "profile-editor.push-base"); recorded != base {
		t.Fatalf("wrong retry lease: %s", recorded)
	}
	if err := os.Remove(filepath.Join(remote, "hooks", "pre-receive")); err != nil {
		t.Fatal(err)
	}
	response, err := app.PushGit(GitPushRequest{OperationID: "retry"})
	if err != nil {
		t.Fatal(err)
	}
	if response.Commit != failedCommit || gitTestCommand(t, remote, "rev-parse", "main") != failedCommit {
		t.Fatal("retry did not upload the saved commit")
	}
}

func advanceGitRemote(t *testing.T, remote string) string {
	t.Helper()
	other := filepath.Join(t.TempDir(), "other")
	gitTestCommand(t, filepath.Dir(other), "clone", remote, other)
	gitTestCommand(t, other, "config", "user.name", "Other Test")
	gitTestCommand(t, other, "config", "user.email", "other@example.invalid")
	writeGitFixture(t, other, "other-change.txt", "remote change\n")
	gitTestCommand(t, other, "add", "-A")
	gitTestCommand(t, other, "commit", "-m", "other change")
	gitTestCommand(t, other, "push")
	return gitTestCommand(t, other, "rev-parse", "HEAD")
}

func TestGitPushRejectsRemoteChangesBeforeCommitting(t *testing.T) {
	root, remote, app := newGitPushFixture(t)
	local := gitTestCommand(t, root, "rev-parse", "HEAD")
	other := advanceGitRemote(t, remote)
	writeGitFixture(t, root, "data/new-item.json", "[]\n")
	if _, err := app.PushGit(GitPushRequest{OperationID: "diverged"}); err == nil || !strings.Contains(err.Error(), "원격에 다른 변경") {
		t.Fatalf("expected remote conflict, got %v", err)
	}
	if gitTestCommand(t, root, "rev-parse", "HEAD") != local || gitTestCommand(t, remote, "rev-parse", "main") != other {
		t.Fatal("conflict altered history")
	}
}

func TestGitPushLeaseRejectsConcurrentRemoteUpdateAndRetry(t *testing.T) {
	root, remote, app := newGitPushFixture(t)
	writeGitFixture(t, root, "data/new-item.json", "[]\n")
	other := ""
	app.progressSink = func(p OperationProgress) {
		if p.Message == "전송 준비 · 인증 확인 중…" {
			other = advanceGitRemote(t, remote)
		}
	}
	if _, err := app.PushGit(GitPushRequest{OperationID: "race"}); err == nil {
		t.Fatal("lease accepted concurrent remote update")
	}
	if other == "" || gitTestCommand(t, remote, "rev-parse", "main") != other {
		t.Fatal("remote change overwritten")
	}
	app.progressSink = nil
	local := gitTestCommand(t, root, "rev-parse", "HEAD")
	if _, err := app.PushGit(GitPushRequest{OperationID: "retry"}); err == nil || !strings.Contains(err.Error(), "원격에 다른 변경") {
		t.Fatalf("unsafe retry: %v", err)
	}
	if gitTestCommand(t, root, "rev-parse", "HEAD") != local {
		t.Fatal("unsafe retry amended another commit")
	}
}

func TestGitPushValidationFailureLeavesIndexAndHistoryUntouched(t *testing.T) {
	root, remote, app := newGitPushFixture(t)
	base := gitTestCommand(t, root, "rev-parse", "HEAD")
	app.siteBuilder = func(context.Context, string) error { return errors.New("invalid data fixture") }
	if _, err := app.PushGit(GitPushRequest{OperationID: "invalid"}); err == nil || !strings.Contains(err.Error(), "invalid data fixture") {
		t.Fatalf("expected validation failure, got %v", err)
	}
	if gitTestCommand(t, root, "rev-parse", "HEAD") != base || gitTestCommand(t, remote, "rev-parse", "main") != base {
		t.Fatal("invalid data was committed or uploaded")
	}
	if staged := gitTestCommand(t, root, "diff", "--cached", "--name-only"); staged != "" {
		t.Fatalf("validation staged files: %s", staged)
	}
}

func TestGitPushRequiresSavedDataAndRejectsConcurrentOperations(t *testing.T) {
	app := &App{repoRoot: newBoardRepoFixture(t), dirty: true}
	if _, err := app.PushGit(GitPushRequest{}); err == nil || !strings.Contains(err.Error(), "먼저 저장") {
		t.Fatalf("unsaved data accepted: %v", err)
	}
	app.operationBusy.Store(true)
	if _, err := app.PushGit(GitPushRequest{}); err == nil || !strings.Contains(err.Error(), "진행 중") {
		t.Fatalf("concurrent push accepted: %v", err)
	}
	if _, err := app.SaveEditorData(SaveEditorDataRequest{}); err == nil || !strings.Contains(err.Error(), "진행 중") {
		t.Fatalf("save during push accepted: %v", err)
	}
}

func TestGitProgressParsesCarriageReturnsAndWaitsForServer(t *testing.T) {
	var messages []OperationProgress
	git := gitRunner{report: func(message string, percent int) {
		messages = append(messages, OperationProgress{Message: message, Percent: percent})
	}}
	writer := &commandMessages{progress: git.transferProgress}
	_, _ = writer.Write([]byte("Compressing objects: 42%\rWriting obj"))
	_, _ = writer.Write([]byte("ects: 100%\r\nResolving deltas: 50%"))
	writer.flush()
	if len(messages) != 3 || messages[0].Percent != 42 || messages[1].Percent != -1 || messages[2].Percent != 50 {
		t.Fatalf("incorrect progress: %#v", messages)
	}
	if !strings.Contains(messages[1].Message, "서버 응답 대기") {
		t.Fatal("reported success before server acceptance")
	}
	if count := gitStatusCount(" M first\x00R  new\x00old\x00?? third\x00"); count != 3 {
		t.Fatalf("rename counted twice: %d", count)
	}
}
