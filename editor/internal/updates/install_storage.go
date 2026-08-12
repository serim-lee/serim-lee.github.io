package updates

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"

	"github.com/Gonie-Gonie/Gonie-Gonie.github.io/editor/internal/dataversion"
	"github.com/Gonie-Gonie/Gonie-Gonie.github.io/editor/internal/release"
)

var transactionIDPattern = regexp.MustCompile(`^[0-9]{8}T[0-9]{6}Z-[a-f0-9]{12}$`)
var gitSHAPattern = regexp.MustCompile(`^[a-f0-9]{40}$`)
var sha256Pattern = regexp.MustCompile(`^[a-f0-9]{64}$`)

func samePath(a, b string) bool {
	if runtime.GOOS == "windows" {
		return strings.EqualFold(filepath.Clean(a), filepath.Clean(b))
	}
	return filepath.Clean(a) == filepath.Clean(b)
}

func canonicalRoot(root string) (string, error) {
	absolute, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(absolute)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(resolved)
	if err != nil || !info.IsDir() {
		return "", fmt.Errorf("업데이트 저장소 경로가 올바르지 않습니다")
	}
	return filepath.Clean(resolved), nil
}

func safePath(root, relative string) (string, error) {
	if relative == "" || strings.Contains(relative, "\\") {
		return "", fmt.Errorf("invalid update path %q", relative)
	}
	if _, err := (release.Policy{}).Classify(relative); err != nil {
		return "", err
	}
	target := filepath.Join(root, filepath.FromSlash(relative))
	rel, err := filepath.Rel(root, target)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return "", fmt.Errorf("update path escapes repository")
	}
	current := root
	for index, part := range strings.Split(filepath.FromSlash(relative), string(filepath.Separator)) {
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if os.IsNotExist(err) {
			break
		}
		if err != nil {
			return "", err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("symbolic links are not updated: %s", relative)
		}
		resolved, err := filepath.EvalSymlinks(current)
		if err != nil || !samePath(resolved, current) {
			return "", fmt.Errorf("reparse paths are not updated: %s", relative)
		}
		if index < len(strings.Split(filepath.FromSlash(relative), string(filepath.Separator)))-1 && !info.IsDir() {
			return "", fmt.Errorf("update parent is not a directory: %s", relative)
		}
	}
	return target, nil
}

func protectedPath(name string) bool {
	name = strings.ToLower(name)
	for _, prefix := range []string{"data", "exports", ".git", ".vscode", ".idea", "editor/runtime", "editor/cmd/editor/build/bin", "editor/cmd/editor/frontend/wailsjs/go", "app/generated"} {
		if name == prefix || strings.HasPrefix(name, prefix+"/") {
			return true
		}
	}
	if name == "cname" {
		return true
	}
	for _, part := range strings.Split(name, "/") {
		if part == ".git" {
			return true
		}
	}
	return false
}

func defaultInstallPolicy() release.Policy {
	return release.Policy{Default: "preserve", Replace: []string{"app/", "editor/", "index.html", "cv.html", "fork-guide.html", "Profile-Editor.exe", "README.md", ".github/workflows/validate.yml"}}
}

// Installer policy intentionally ignores schema counters and regeneration
// configuration. Hard protection always wins over locally declared rules.
func installPolicy(root string) release.Policy {
	policy := defaultInstallPolicy()
	raw, err := os.ReadFile(filepath.Join(root, "app", "update-policy.json"))
	if err != nil || len(raw) > 64<<10 {
		return policy
	}
	var configured release.Policy
	if dataversion.Decode(raw, &configured) != nil {
		return policy
	}
	validRules := func(rules []string) []string {
		out := []string{}
		for _, rule := range rules {
			trimmed := strings.TrimSuffix(rule, "/")
			if trimmed == "" || strings.Contains(rule, "\\") {
				continue
			}
			if _, err := (release.Policy{}).Classify(trimmed); err == nil {
				out = append(out, rule)
			}
		}
		return out
	}
	if configured.Replace != nil {
		policy.Replace = validRules(configured.Replace)
	}
	policy.Preserve = validRules(configured.Preserve)
	return policy
}

func replaceAllowed(policy release.Policy, name string) bool {
	if protectedPath(name) {
		return false
	}
	action, err := policy.Classify(name)
	return err == nil && action == "replace"
}

func hashFile(path string) (string, int64, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", 0, err
	}
	defer file.Close()
	hash := sha256.New()
	size, err := io.Copy(hash, io.LimitReader(file, MaxInstallFileBytes+1))
	if err != nil {
		return "", 0, err
	}
	if size > MaxInstallFileBytes {
		return "", 0, fmt.Errorf("update file exceeds maximum size")
	}
	return hex.EncodeToString(hash.Sum(nil)), size, nil
}

func atomicWrite(path string, raw []byte, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	temp, err := os.CreateTemp(filepath.Dir(path), ".profile-update-*.tmp")
	if err != nil {
		return err
	}
	name := temp.Name()
	defer os.Remove(name)
	if _, err = temp.Write(raw); err == nil {
		err = temp.Sync()
	}
	if closeErr := temp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	if err = os.Chmod(name, mode.Perm()); err != nil {
		return err
	}
	return os.Rename(name, path)
}

func atomicCopy(source, destination string, mode os.FileMode) error {
	file, err := os.Open(source)
	if err != nil {
		return err
	}
	defer file.Close()
	raw, err := io.ReadAll(io.LimitReader(file, MaxInstallFileBytes+1))
	if err != nil {
		return err
	}
	if int64(len(raw)) > MaxInstallFileBytes {
		return fmt.Errorf("update file exceeds maximum size")
	}
	return atomicWrite(destination, raw, mode)
}

func journalLocation(root, name string) (string, string, error) {
	root, err := canonicalRoot(root)
	if err != nil {
		return "", "", err
	}
	absolute := name
	if !filepath.IsAbs(absolute) {
		absolute = filepath.Join(root, name)
	}
	absolute = filepath.Clean(absolute)
	id := filepath.Base(filepath.Dir(absolute))
	if !transactionIDPattern.MatchString(id) || filepath.Base(absolute) != "journal.json" {
		return "", "", fmt.Errorf("invalid update journal location")
	}
	expected := filepath.Join(root, "editor", "runtime", "updates", id, "journal.json")
	if !samePath(expected, absolute) {
		return "", "", fmt.Errorf("update journal must stay inside editor/runtime/updates")
	}
	rel, _ := filepath.Rel(root, expected)
	if _, err := safePath(root, filepath.ToSlash(rel)); err != nil {
		return "", "", err
	}
	return root, expected, nil
}

func validateJournal(root, path string, journal Journal) error {
	if journal.FormatVersion != 1 || journal.ID != filepath.Base(filepath.Dir(path)) || !samePath(journal.Root, root) || !gitSHAPattern.MatchString(journal.Commit) {
		return fmt.Errorf("invalid update journal metadata")
	}
	if _, err := dataversion.Parse(journal.CurrentVersion); err != nil {
		return err
	}
	if _, err := dataversion.Parse(journal.TargetVersion); err != nil {
		return err
	}
	if err := validateSource(Source{Repository: journal.Repository, Ref: journal.Ref}); err != nil {
		return err
	}
	if len(journal.Files) > 5000 {
		return fmt.Errorf("update journal has too many files")
	}
	seen := map[string]bool{}
	for _, file := range journal.Files {
		if protectedPath(file.Path) {
			return fmt.Errorf("protected path in update journal: %s", file.Path)
		}
		if _, err := safePath(root, file.Path); err != nil {
			return err
		}
		key := strings.ToLower(file.Path)
		if seen[key] {
			return fmt.Errorf("duplicate update path: %s", file.Path)
		}
		seen[key] = true
		if file.Action != "replace" && file.Action != "delete" {
			return fmt.Errorf("invalid update action")
		}
		if file.Deferred != (file.Path == ExecutablePath) {
			return fmt.Errorf("invalid deferred executable path")
		}
		if file.Size < 0 || file.Size > MaxInstallFileBytes {
			return fmt.Errorf("invalid update file size")
		}
		if file.SHA256 != "" && !sha256Pattern.MatchString(file.SHA256) {
			return fmt.Errorf("invalid payload checksum")
		}
		if file.BackupReady && file.Existed && !sha256Pattern.MatchString(file.BackupSHA256) {
			return fmt.Errorf("invalid backup checksum")
		}
	}
	return nil
}

func LoadJournal(root, path string) (Journal, error) {
	root, path, err := journalLocation(root, path)
	if err != nil {
		return Journal{}, err
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return Journal{}, err
	}
	if len(raw) > 8<<20 {
		return Journal{}, fmt.Errorf("update journal exceeds maximum size")
	}
	var journal Journal
	if err := dataversion.Decode(raw, &journal); err != nil {
		return journal, err
	}
	return journal, validateJournal(root, path, journal)
}

func SaveJournal(root, path string, journal Journal) error {
	root, path, err := journalLocation(root, path)
	if err != nil {
		return err
	}
	if err := validateJournal(root, path, journal); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(journal, "", "  ")
	if err != nil {
		return err
	}
	return atomicWrite(path, append(raw, '\n'), 0600)
}

func PendingJournals(root string) ([]string, error) {
	root, err := canonicalRoot(root)
	if err != nil {
		return nil, err
	}
	base, err := safePath(root, "editor/runtime/updates")
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(base)
	if os.IsNotExist(err) {
		return []string{}, nil
	}
	if err != nil {
		return nil, err
	}
	paths := []string{}
	for _, entry := range entries {
		if !entry.IsDir() || !transactionIDPattern.MatchString(entry.Name()) {
			continue
		}
		path := filepath.Join(base, entry.Name(), "journal.json")
		journal, err := LoadJournal(root, path)
		if err != nil {
			if os.IsNotExist(err) {
				// A process can stop after creating its staging directory but
				// before the first atomic journal rename. Payload alone cannot
				// indicate an applied update; backups or helper state can.
				hasRecoveryEvidence := false
				for _, name := range []string{"backups", "backup-history", "rollback-current", "self-update"} {
					_, evidenceErr := os.Lstat(filepath.Join(base, entry.Name(), name))
					if evidenceErr == nil {
						hasRecoveryEvidence = true
					} else if !os.IsNotExist(evidenceErr) {
						return nil, evidenceErr
					}
				}
				if !hasRecoveryEvidence {
					continue
				}
			}
			return nil, err
		}
		switch journal.State {
		case StateApplying, StatePendingExecutable, StateAwaitingVerification, StateRollbackFailed, StateRollbackPendingExecutable:
			paths = append(paths, path)
		}
	}
	sort.Strings(paths)
	return paths, nil
}
