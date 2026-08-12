package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Gonie-Gonie/Gonie-Gonie.github.io/editor/internal/sitebuild"
	"github.com/Gonie-Gonie/Gonie-Gonie.github.io/editor/internal/sitecheck"
)

type GitPushRequest struct {
	OperationID string `json:"operation_id"`
}

type GitPushResponse struct {
	Message      string `json:"message"`
	Branch       string `json:"branch"`
	Commit       string `json:"commit"`
	ChangedFiles int    `json:"changed_files"`
}

// PushGit preserves the last commit's message and parents. A failed upload
// keeps the local commit and records its original remote revision for retry.
func (a *App) PushGit(request GitPushRequest) (GitPushResponse, error) {
	if !a.operationBusy.CompareAndSwap(false, true) {
		return GitPushResponse{}, errors.New("다른 작업이 진행 중입니다")
	}
	defer a.operationBusy.Store(false)
	a.mu.Lock()
	root, err := a.rootLocked()
	dirty := a.dirty || len(a.boardMedia) > 0
	appContext := a.ctx
	a.mu.Unlock()
	if err != nil {
		return GitPushResponse{}, err
	}
	if dirty {
		return GitPushResponse{}, errors.New("변경 사항을 먼저 저장해 주세요")
	}
	if appContext == nil {
		appContext = context.Background()
	}
	ctx, cancel := context.WithTimeout(appContext, 10*time.Minute)
	defer cancel()
	report := func(message string, percent int) { a.reportProgress(request.OperationID, message, percent) }
	report("저장소 확인 중…", -1)
	gitPath, err := exec.LookPath("git")
	if err != nil {
		return GitPushResponse{}, errors.New("Git을 찾을 수 없습니다. Git for Windows를 설치해 주세요")
	}
	git := gitRunner{ctx: ctx, root: root, executable: gitPath, report: report}
	top, err := git.run("저장소 확인", nil, "rev-parse", "--show-toplevel")
	if err != nil {
		return GitPushResponse{}, err
	}
	if !strings.EqualFold(filepath.Clean(strings.TrimSpace(top)), filepath.Clean(root)) {
		return GitPushResponse{}, errors.New("편집기 폴더가 Git 저장소 루트와 다릅니다")
	}
	branch, err := git.run("브랜치 확인", nil, "symbolic-ref", "--quiet", "--short", "HEAD")
	if err != nil {
		return GitPushResponse{}, errors.New("현재 브랜치를 확인할 수 없습니다. 브랜치에 체크아웃해 주세요")
	}
	branch = strings.TrimSpace(branch)
	remote, err := git.run("원격 저장소 확인", nil, "config", "--get", "branch."+branch+".remote")
	if err != nil {
		return GitPushResponse{}, errors.New("현재 브랜치의 원격 저장소가 설정되어 있지 않습니다")
	}
	remote = strings.TrimSpace(remote)
	ref, err := git.run("원격 브랜치 확인", nil, "config", "--get", "branch."+branch+".merge")
	if err != nil || remote == "." || !strings.HasPrefix(strings.TrimSpace(ref), "refs/heads/") {
		return GitPushResponse{}, errors.New("현재 브랜치의 push 대상이 설정되어 있지 않습니다")
	}
	ref = strings.TrimSpace(ref)
	if err := git.checkReady(); err != nil {
		return GitPushResponse{}, err
	}
	localHead, err := git.run("커밋 확인", nil, "rev-parse", "HEAD")
	if err != nil {
		return GitPushResponse{}, err
	}
	localHead = strings.TrimSpace(localHead)
	report("원격 변경 확인 중…", -1)
	if _, err := git.run("원격 변경 확인", git.transferProgress, "fetch", "--no-tags", "--progress", "--", remote, ref); err != nil {
		return GitPushResponse{}, err
	}
	remoteHead, err := git.run("원격 커밋 확인", nil, "rev-parse", "FETCH_HEAD^{commit}")
	if err != nil {
		return GitPushResponse{}, err
	}
	remoteHead = strings.TrimSpace(remoteHead)
	if remoteHead != localHead {
		_, ancestorErr := git.run("커밋 비교", nil, "merge-base", "--is-ancestor", remoteHead, localHead)
		if ancestorErr != nil {
			if !isGitExitCode(ancestorErr, 1) {
				return GitPushResponse{}, ancestorErr
			}
			if git.config("profile-editor.push-base") != remoteHead ||
				git.config("profile-editor.push-head") != localHead ||
				git.config("profile-editor.push-branch") != branch {
				return GitPushResponse{}, errors.New("원격에 다른 변경이 있습니다. Git에서 변경을 확인한 뒤 다시 시도해 주세요")
			}
		}
	}

	report("데이터 검증 · 내보내기 생성 중…", -1)
	build := a.siteBuilder
	if build == nil {
		build = buildForGitPush
	}
	if err := build(ctx, root); err != nil {
		return GitPushResponse{}, err
	}
	if err := git.checkReady(); err != nil {
		return GitPushResponse{}, err
	}
	currentHead, err := git.run("커밋 재확인", nil, "rev-parse", "HEAD")
	if err != nil {
		return GitPushResponse{}, err
	}
	currentBranch, err := git.run("브랜치 재확인", nil, "symbolic-ref", "--quiet", "--short", "HEAD")
	if err != nil {
		return GitPushResponse{}, err
	}
	if strings.TrimSpace(currentHead) != localHead || strings.TrimSpace(currentBranch) != branch {
		return GitPushResponse{}, errors.New("작업 중 Git 브랜치 또는 커밋이 변경되었습니다. 다시 시도해 주세요")
	}
	status, err := git.run("변경 파일 확인", nil, "status", "--porcelain=v1", "-z", "--untracked-files=all")
	if err != nil {
		return GitPushResponse{}, err
	}
	changedFiles := gitStatusCount(status)
	if changedFiles > 0 {
		report(fmt.Sprintf("변경 파일 %d개 반영 중…", changedFiles), -1)
		if _, err := git.run("변경 파일 반영", nil, "add", "-A", "--", "."); err != nil {
			return GitPushResponse{}, err
		}
		_, diffErr := git.run("커밋할 변경 확인", nil, "diff", "--cached", "--quiet")
		if diffErr != nil {
			if !isGitExitCode(diffErr, 1) {
				return GitPushResponse{}, diffErr
			}
			report("커밋 갱신 중…", -1)
			if _, err := git.run("커밋 갱신", nil, "commit", "--amend", "--no-edit"); err != nil {
				return GitPushResponse{}, err
			}
		}
	}
	commit, err := git.run("저장된 커밋 확인", nil, "rev-parse", "HEAD")
	if err != nil {
		return GitPushResponse{}, err
	}
	commit = strings.TrimSpace(commit)
	if commit == remoteHead {
		report("이미 최신 상태입니다", 100)
		return GitPushResponse{Message: "이미 최신 상태입니다", Branch: branch, Commit: commit, ChangedFiles: changedFiles}, nil
	}
	// Persist the exact lease before uploading; later background fetches cannot
	// silently change the revision we are willing to replace.
	for _, entry := range [][2]string{
		{"profile-editor.push-base", remoteHead},
		{"profile-editor.push-head", commit},
		{"profile-editor.push-branch", branch},
	} {
		if _, err := git.run("재시도 정보 저장", nil, "config", "--local", entry[0], entry[1]); err != nil {
			return GitPushResponse{}, err
		}
	}
	status, err = git.run("변경 재확인", nil, "status", "--porcelain=v1", "-z", "--untracked-files=all")
	if err != nil {
		return GitPushResponse{}, err
	}
	if gitStatusCount(status) > 0 {
		return GitPushResponse{}, errors.New("작업 중 파일이 변경되었습니다. 변경을 확인한 뒤 다시 시도해 주세요")
	}
	report("전송 준비 · 인증 확인 중…", -1)
	_, err = git.run("Git push", git.transferProgress,
		"push", "--progress", "--porcelain", "--force-with-lease="+ref+":"+remoteHead, "--", remote, commit+":"+ref)
	if err != nil {
		return GitPushResponse{}, err
	}
	report("Git push 완료", 100)
	return GitPushResponse{Message: "Git push 완료", Branch: branch, Commit: commit, ChangedFiles: changedFiles}, nil
}

type gitRunner struct {
	ctx        context.Context
	root       string
	executable string
	report     func(string, int)
}

func commandEnvironment() []string {
	env := os.Environ()
	for _, item := range []string{"GIT_TERMINAL_PROMPT=0", "LC_ALL=C", "LANG=C"} {
		key := strings.SplitN(item, "=", 2)[0] + "="
		filtered := env[:0]
		for _, old := range env {
			if !strings.HasPrefix(strings.ToUpper(old), key) {
				filtered = append(filtered, old)
			}
		}
		env = append(filtered, item)
	}
	return env
}

func (g gitRunner) run(label string, progress func(string), args ...string) (string, error) {
	cmd := exec.CommandContext(g.ctx, g.executable, args...)
	cmd.Dir = g.root
	cmd.Env = commandEnvironment()
	hideCommandWindow(cmd)
	var stdout bytes.Buffer
	stderr := &commandMessages{progress: progress}
	cmd.Stdout = &stdout
	cmd.Stderr = stderr
	err := cmd.Run()
	stderr.flush()
	if err != nil {
		if g.ctx.Err() != nil {
			return "", fmt.Errorf("%s 시간이 초과되거나 작업이 종료되었습니다: %w", label, g.ctx.Err())
		}
		detail := strings.TrimSpace(stderr.tail)
		if detail == "" {
			detail = strings.TrimSpace(stdout.String())
		}
		if detail == "" {
			detail = err.Error()
		}
		if len(detail) > 1600 {
			detail = detail[len(detail)-1600:]
		}
		return "", fmt.Errorf("%s 실패: %s: %w", label, detail, err)
	}
	return stdout.String(), nil
}

func isGitExitCode(err error, code int) bool {
	var exitError *exec.ExitError
	return errors.As(err, &exitError) && exitError.ExitCode() == code
}

func (g gitRunner) config(key string) string {
	value, err := g.run("Git 설정 확인", nil, "config", "--local", "--get", key)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(value)
}

func (g gitRunner) checkReady() error {
	unmerged, err := g.run("충돌 확인", nil, "ls-files", "--unmerged")
	if err != nil {
		return err
	}
	if strings.TrimSpace(unmerged) != "" {
		return errors.New("Git 충돌을 해결한 뒤 다시 시도해 주세요")
	}
	for _, name := range []string{"MERGE_HEAD", "CHERRY_PICK_HEAD", "REVERT_HEAD", "rebase-merge", "rebase-apply"} {
		location, err := g.run("Git 작업 상태 확인", nil, "rev-parse", "--git-path", name)
		if err != nil {
			return err
		}
		location = strings.TrimSpace(location)
		if !filepath.IsAbs(location) {
			location = filepath.Join(g.root, location)
		}
		if _, err := os.Stat(location); err == nil {
			return errors.New("진행 중인 Git 작업을 마친 뒤 다시 시도해 주세요")
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	return nil
}

func gitStatusCount(status string) int {
	entries := strings.Split(status, "\x00")
	count := 0
	for i := 0; i < len(entries); i++ {
		if len(entries[i]) < 3 {
			continue
		}
		count++
		if strings.ContainsAny(entries[i][:2], "RC") {
			i++
		}
	}
	return count
}

var gitPercentPattern = regexp.MustCompile(`(Counting|Compressing|Writing|Receiving|Resolving) (?:objects|deltas):\s+(\d+)%`)

func (g gitRunner) transferProgress(line string) {
	match := gitPercentPattern.FindStringSubmatch(line)
	if len(match) == 0 {
		return
	}
	percent, _ := strconv.Atoi(match[2])
	label := map[string]string{"Counting": "전송 목록 확인", "Compressing": "압축", "Writing": "전송", "Receiving": "원격 변경 수신", "Resolving": "원격 변경 확인"}[match[1]]
	if g.report != nil {
		if match[1] == "Writing" && percent == 100 {
			g.report("전송 완료 · 서버 응답 대기 중…", -1)
		} else {
			g.report(fmt.Sprintf("%s 중… %d%%", label, percent), percent)
		}
	}
}

// Git progress uses carriage returns as well as newlines. Keep only a bounded
// error tail so a large transfer cannot grow memory without limit.
type commandMessages struct {
	mu       sync.Mutex
	pending  string
	tail     string
	progress func(string)
}

func (w *commandMessages) Write(data []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.pending += string(data)
	for {
		index := strings.IndexAny(w.pending, "\r\n")
		if index < 0 {
			break
		}
		line := w.pending[:index]
		w.pending = w.pending[index+1:]
		w.line(line)
	}
	if len(w.pending) > 8192 {
		w.pending = w.pending[len(w.pending)-8192:]
	}
	return len(data), nil
}

func (w *commandMessages) line(line string) {
	if line == "" {
		return
	}
	w.tail += line + "\n"
	if len(w.tail) > 8192 {
		w.tail = w.tail[len(w.tail)-8192:]
	}
	if w.progress != nil {
		w.progress(line)
	}
}

func (w *commandMessages) flush() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.line(w.pending)
	w.pending = ""
}

func buildForGitPush(ctx context.Context, root string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if _, err := sitebuild.Build(root, false); err != nil {
		return fmt.Errorf("데이터 검증 실패: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if _, err := sitecheck.Check(root); err != nil {
		return fmt.Errorf("사이트 검사 실패: %w", err)
	}
	return ctx.Err()
}
