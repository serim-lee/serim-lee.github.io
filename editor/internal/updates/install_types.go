package updates

import (
	"net/http"
	"time"
)

const (
	StateDownloading                     = "downloading"
	StateDownloadFailed                  = "download-failed"
	StatePrepared                        = "prepared"
	StateApplying                        = "applying"
	StatePendingExecutable               = "pending-executable"
	StateAwaitingVerification            = "awaiting-verification"
	StateCompleted                       = "completed"
	StateRolledBack                      = "rolled-back"
	StateRollbackFailed                  = "rollback-failed"
	StateRollbackPendingExecutable       = "rollback-pending-executable"
	ExecutablePath                       = "Profile-Editor.exe"
	MaxInstallFileBytes            int64 = 128 << 20
	MaxInstallBytes                int64 = 256 << 20
	InstallTimeout                       = 10 * time.Minute
)

type InstallProgress struct {
	Message    string
	Percent    int
	Phase      string
	Path       string
	Completed  int
	Total      int
	Bytes      int64
	TotalBytes int64
}

type InstallOptions struct {
	Root          string
	Client        *http.Client
	APIBaseURL    string
	RawBaseURL    string
	TargetVersion string
	Progress      func(InstallProgress)
}

type InstallFile struct {
	Path         string `json:"path"`
	Action       string `json:"action"`
	SHA256       string `json:"sha256,omitempty"`
	GitSHA       string `json:"git_sha,omitempty"`
	Size         int64  `json:"size"`
	Deferred     bool   `json:"deferred"`
	Existed      bool   `json:"existed"`
	BackupSHA256 string `json:"backup_sha256,omitempty"`
	Mode         uint32 `json:"mode"`
	BackupReady  bool   `json:"backup_ready"`
	Applied      bool   `json:"applied"`
}

type Journal struct {
	FormatVersion  int           `json:"format_version"`
	ID             string        `json:"id"`
	Root           string        `json:"root"`
	State          string        `json:"state"`
	CurrentVersion string        `json:"current_version"`
	TargetVersion  string        `json:"target_version"`
	Repository     string        `json:"repository"`
	Ref            string        `json:"ref"`
	Commit         string        `json:"commit"`
	CreatedAt      string        `json:"created_at"`
	ProcessID      int           `json:"process_id"`
	Files          []InstallFile `json:"files"`
	Error          string        `json:"error,omitempty"`
}

type PreparedInstall struct {
	Root        string  `json:"root"`
	JournalPath string  `json:"journal_path"`
	Journal     Journal `json:"journal"`
}

type managedFile struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size"`
}
type installedManifest struct {
	Version string        `json:"version"`
	Commit  string        `json:"commit"`
	Files   []managedFile `json:"files"`
}
