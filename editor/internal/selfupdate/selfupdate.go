// Package selfupdate hands off executable replacement and verifies it at startup.
package selfupdate

import (
	"crypto/rand"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/Gonie-Gonie/Gonie-Gonie.github.io/editor/internal/dataversion"
)

//go:embed apply-editor-update.ps1
var scriptFiles embed.FS

const (
	ActionNone           = "none"
	ActionComplete       = "complete"
	ActionExitForReplace = "exit_for_replace"
	ActionFailed         = "failed"
)

type Request struct {
	Root             string `json:"root"`
	TransactionDir   string `json:"transaction_dir"`
	StagedExecutable string `json:"staged_executable"`
	ExpectedSHA256   string `json:"expected_sha256"`
	TargetVersion    string `json:"target_version"`
	WaitPID          int    `json:"wait_pid"`
	WaitStarted      string `json:"wait_started"`
	Relaunch         bool   `json:"relaunch"`
}
type Pending struct {
	Request
	FormatVersion int    `json:"format_version"`
	ID            string `json:"id"`
	CreatedAt     string `json:"created_at"`
	Launches      int    `json:"launches"`
}
type Result struct {
	ID        string `json:"id"`
	Status    string `json:"status"`
	Message   string `json:"message"`
	UpdatedAt string `json:"updated_at"`
	Finalized bool   `json:"finalized"`
}
type Recovery struct {
	Action         string `json:"action"`
	TransactionDir string `json:"transaction_dir,omitempty"`
	TargetVersion  string `json:"target_version,omitempty"`
	Message        string `json:"message,omitempty"`
}

var digestPattern = regexp.MustCompile(`^[a-f0-9]{64}$`)
var idPattern = regexp.MustCompile(`^[a-f0-9]{24}$`)

func normalizedRoot(root string) (string, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	info, err := os.Lstat(root)
	if err != nil {
		return "", err
	}
	if !info.IsDir() || isReparse(info) {
		return "", fmt.Errorf("invalid update root")
	}
	return filepath.Clean(root), nil
}
func within(root, name string) error {
	name, err := filepath.Abs(name)
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(root, name)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) || filepath.IsAbs(rel) {
		return fmt.Errorf("update path leaves repository")
	}
	current := root
	for _, part := range strings.Split(rel, string(os.PathSeparator)) {
		if part == "." {
			continue
		}
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return err
		}
		if isReparse(info) {
			return fmt.Errorf("update paths cannot use links or reparse points")
		}
	}
	return nil
}
func normalizeRequest(request Request) (Request, error) {
	root, err := normalizedRoot(request.Root)
	if err != nil {
		return request, err
	}
	request.Root = root
	request.TransactionDir, err = filepath.Abs(request.TransactionDir)
	if err != nil {
		return request, err
	}
	updatesRoot := filepath.Join(root, "editor", "runtime", "updates")
	rel, err := filepath.Rel(updatesRoot, request.TransactionDir)
	if err != nil || rel == "." || rel == ".." || strings.ContainsAny(rel, "/\\") || filepath.IsAbs(rel) {
		return request, fmt.Errorf("invalid update transaction directory")
	}
	if err := within(root, request.TransactionDir); err != nil {
		return request, err
	}
	request.StagedExecutable, err = filepath.Abs(request.StagedExecutable)
	if err != nil {
		return request, err
	}
	if !strings.EqualFold(request.StagedExecutable, filepath.Join(request.TransactionDir, "payload", "Profile-Editor.exe")) && !strings.EqualFold(request.StagedExecutable, filepath.Join(request.TransactionDir, "backups", "Profile-Editor.exe")) {
		return request, fmt.Errorf("invalid staged editor path")
	}
	for _, name := range []string{request.StagedExecutable, filepath.Join(root, "Profile-Editor.exe"), filepath.Join(request.TransactionDir, "self-update")} {
		if err := within(root, name); err != nil {
			return request, err
		}
	}
	request.ExpectedSHA256 = strings.ToLower(request.ExpectedSHA256)
	if !digestPattern.MatchString(request.ExpectedSHA256) {
		return request, fmt.Errorf("invalid executable digest")
	}
	if _, err := dataversion.Parse(request.TargetVersion); err != nil {
		return request, err
	}
	if request.WaitPID < 0 {
		return request, fmt.Errorf("invalid editor process ID")
	}
	return request, nil
}
func hashFile(name string) (string, error) {
	file, err := os.Open(name)
	if err != nil {
		return "", err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("executable is not a regular file")
	}
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
func writeJSON(name string, value any) error {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	raw = append(raw, '\n')
	file, err := os.CreateTemp(filepath.Dir(name), ".self-update-*.tmp")
	if err != nil {
		return err
	}
	temporary := file.Name()
	defer os.Remove(temporary)
	if _, err = file.Write(raw); err == nil {
		err = file.Sync()
	}
	closeErr := file.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	return os.Rename(temporary, name)
}
func readPending(root, transactionDir string) (Pending, error) {
	var pending Pending
	raw, err := os.ReadFile(filepath.Join(transactionDir, "self-update", "handoff.json"))
	if err != nil {
		return pending, err
	}
	if len(raw) > 64<<10 {
		return pending, fmt.Errorf("self-update handoff is too large")
	}
	if err := dataversion.Decode(raw, &pending); err != nil {
		return pending, err
	}
	if pending.FormatVersion != 1 || !idPattern.MatchString(pending.ID) || pending.Launches < 0 {
		return pending, fmt.Errorf("invalid self-update handoff")
	}
	request, err := normalizeRequest(pending.Request)
	if err != nil {
		return pending, err
	}
	if !strings.EqualFold(request.Root, root) || !strings.EqualFold(request.TransactionDir, transactionDir) {
		return pending, fmt.Errorf("self-update handoff belongs to another repository")
	}
	pending.Request = request
	return pending, nil
}
func readResult(transactionDir string) (Result, error) {
	var result Result
	raw, err := os.ReadFile(filepath.Join(transactionDir, "self-update", "result.json"))
	if os.IsNotExist(err) {
		return result, nil
	}
	if err != nil {
		return result, err
	}
	if len(raw) > 64<<10 {
		return result, fmt.Errorf("self-update result is too large")
	}
	if err := dataversion.Decode(raw, &result); err != nil {
		return result, err
	}
	return result, nil
}

// Prepare writes only helper files in the staged transaction.
func Prepare(request Request) (Pending, error) {
	request, err := normalizeRequest(request)
	if err != nil {
		return Pending{}, err
	}
	digest, err := hashFile(request.StagedExecutable)
	if err != nil {
		return Pending{}, err
	}
	if digest != request.ExpectedSHA256 {
		return Pending{}, fmt.Errorf("staged executable digest mismatch")
	}
	if request.WaitPID == 0 {
		request.WaitPID = os.Getpid()
	}
	if request.WaitStarted == "" {
		request.WaitStarted, err = processStarted(request.WaitPID)
		if err != nil {
			return Pending{}, err
		}
	}
	directory := filepath.Join(request.TransactionDir, "self-update")
	if err := os.MkdirAll(directory, 0700); err != nil {
		return Pending{}, err
	}
	if err := within(request.Root, directory); err != nil {
		return Pending{}, err
	}
	for _, leaf := range []string{"apply-editor-update.ps1", "handoff.json", "result.json", "helper.log", "helper.lock", "previous.exe"} {
		name := filepath.Join(directory, leaf)
		if err := within(request.Root, name); err != nil {
			return Pending{}, err
		}
		if info, err := os.Stat(name); err == nil && !info.Mode().IsRegular() {
			return Pending{}, fmt.Errorf("helper path is not a regular file")
		} else if err != nil && !os.IsNotExist(err) {
			return Pending{}, err
		}
	}
	pending, readErr := readPending(request.Root, request.TransactionDir)
	if readErr != nil && !os.IsNotExist(readErr) {
		return Pending{}, readErr
	}
	if readErr != nil || pending.ExpectedSHA256 != request.ExpectedSHA256 || pending.TargetVersion != request.TargetVersion {
		pending = Pending{FormatVersion: 1, CreatedAt: time.Now().UTC().Format(time.RFC3339)}
	}
	if pending.Launches >= 3 {
		return Pending{}, fmt.Errorf("editor replacement retry limit reached")
	}
	identifier := make([]byte, 12)
	if _, err := rand.Read(identifier); err != nil {
		return Pending{}, err
	}
	pending.ID = hex.EncodeToString(identifier)
	pending.Request = request
	pending.Launches++
	script, err := scriptFiles.ReadFile("apply-editor-update.ps1")
	if err != nil {
		return Pending{}, err
	}
	if err := os.WriteFile(filepath.Join(directory, "apply-editor-update.ps1"), script, 0600); err != nil {
		return Pending{}, err
	}
	if err := writeJSON(filepath.Join(directory, "handoff.json"), pending); err != nil {
		return Pending{}, err
	}
	return pending, nil
}

// Launch starts a hidden helper and never terminates the editor process.
func Launch(request Request) (Pending, error) {
	pending, err := Prepare(request)
	if err != nil {
		return pending, err
	}
	directory := filepath.Join(pending.TransactionDir, "self-update")
	if err := startScript(filepath.Join(directory, "apply-editor-update.ps1"), filepath.Join(directory, "handoff.json"), filepath.Join(directory, "helper.log"), pending.ID); err != nil {
		return pending, err
	}
	return pending, nil
}

// Finalize follows the core journal's successful completion or rollback.
func Finalize(root, transactionDir, status, message string) error {
	root, err := normalizedRoot(root)
	if err != nil {
		return err
	}
	pending, err := readPending(root, transactionDir)
	if err != nil {
		return err
	}
	if status != "complete" && status != "rolled_back" {
		return fmt.Errorf("invalid final self-update status")
	}
	return writeJSON(filepath.Join(transactionDir, "self-update", "result.json"), Result{ID: pending.ID, Status: status, Message: message, UpdatedAt: time.Now().UTC().Format(time.RFC3339), Finalized: true})
}

func RecoverTransaction(root, transactionDir, runningExecutable, compiledVersion string) (Recovery, error) {
	root, err := normalizedRoot(root)
	if err != nil {
		return Recovery{}, err
	}
	transactionDir, err = filepath.Abs(transactionDir)
	if err != nil {
		return Recovery{}, err
	}
	pending, err := readPending(root, transactionDir)
	if os.IsNotExist(err) {
		return Recovery{Action: ActionNone}, nil
	}
	if err != nil {
		return Recovery{Action: ActionFailed, TransactionDir: transactionDir, Message: err.Error()}, err
	}
	result, err := readResult(transactionDir)
	if err != nil {
		return Recovery{Action: ActionFailed, TransactionDir: transactionDir, Message: err.Error()}, err
	}
	if result.ID == pending.ID && result.Finalized {
		return Recovery{Action: ActionNone}, nil
	}
	recovery := Recovery{TransactionDir: transactionDir, TargetVersion: pending.TargetVersion}
	target := filepath.Join(root, "Profile-Editor.exe")
	digest, digestErr := hashFile(target)
	runningExecutable, _ = filepath.Abs(runningExecutable)
	if digestErr == nil && digest == pending.ExpectedSHA256 && compiledVersion == pending.TargetVersion && strings.EqualFold(runningExecutable, target) {
		recovery.Action = ActionComplete
		return recovery, nil
	}
	if (digestErr != nil || digest != pending.ExpectedSHA256) && result.ID == pending.ID && result.Status == "failed" {
		recovery.Action = ActionFailed
		recovery.Message = result.Message
		return recovery, nil
	}
	if pending.Launches >= 3 {
		recovery.Action = ActionFailed
		recovery.Message = "편집기 교체를 확인하지 못했습니다. 이전 파일로 복원한 뒤 다시 업데이트하세요."
		return recovery, nil
	}
	request := pending.Request
	request.WaitPID = os.Getpid()
	request.WaitStarted = ""
	request.Relaunch = true
	if _, err := Launch(request); err != nil {
		recovery.Action = ActionFailed
		recovery.Message = err.Error()
		return recovery, err
	}
	recovery.Action = ActionExitForReplace
	return recovery, nil
}

// Recover finds the newest unfinished handoff. The core journal still owns
// file rollback/finalization; a caller with that journal may use RecoverTransaction.
func Recover(root, runningExecutable, compiledVersion string) (Recovery, error) {
	root, err := normalizedRoot(root)
	if err != nil {
		return Recovery{}, err
	}
	directory := filepath.Join(root, "editor", "runtime", "updates")
	entries, err := os.ReadDir(directory)
	if os.IsNotExist(err) {
		return Recovery{Action: ActionNone}, nil
	}
	if err != nil {
		return Recovery{}, err
	}
	type candidate struct {
		dir      string
		modified time.Time
	}
	candidates := []candidate{}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		info, err := os.Stat(filepath.Join(directory, entry.Name(), "self-update", "handoff.json"))
		if err != nil {
			continue
		}
		result, err := readResult(filepath.Join(directory, entry.Name()))
		pending, pendingErr := readPending(root, filepath.Join(directory, entry.Name()))
		if err == nil && pendingErr == nil && result.ID == pending.ID && result.Finalized {
			continue
		}
		candidates = append(candidates, candidate{filepath.Join(directory, entry.Name()), info.ModTime()})
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].modified.After(candidates[j].modified) })
	if len(candidates) == 0 {
		return Recovery{Action: ActionNone}, nil
	}
	return RecoverTransaction(root, candidates[0].dir, runningExecutable, compiledVersion)
}
