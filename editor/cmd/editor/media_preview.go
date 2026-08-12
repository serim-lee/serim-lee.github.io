package main

import (
	"io"
	"net/http"
	"net/url"
	"os"
)

const editorMediaPreviewPath = "/__editor_media"

// mediaPreviewHandler serves original media only when the preview is opened.
// It never publishes staged files or changes the editor's saved/draft state.
func (a *App) mediaPreviewHandler() http.Handler {
	return http.HandlerFunc(a.serveMediaPreview)
}

func (a *App) serveMediaPreview(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	if r.URL.Path != editorMediaPreviewPath {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, "허용되지 않는 요청입니다.", http.StatusMethodNotAllowed)
		return
	}
	query, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil || len(query) != 1 {
		http.Error(w, "미디어를 하나만 선택해 주세요.", http.StatusBadRequest)
		return
	}
	source, sourceOK := singleMediaPreviewValue(query, "src")
	token, tokenOK := singleMediaPreviewValue(query, "stage_token")
	if sourceOK == tokenOK {
		http.Error(w, "미디어를 하나만 선택해 주세요.", http.StatusBadRequest)
		return
	}

	// Open while holding the staging lock so a concurrent discard cannot revoke
	// the token between its lookup and opening. Do not hold it while streaming.
	file, err := a.openMediaPreview(source, token)
	if err != nil {
		http.Error(w, "미디어 파일을 찾을 수 없습니다.", http.StatusNotFound)
		return
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		http.Error(w, "미디어 파일을 읽을 수 없습니다.", http.StatusNotFound)
		return
	}
	prefix := make([]byte, 512)
	count, err := file.Read(prefix)
	if err != nil && err != io.EOF {
		http.Error(w, "미디어 파일을 읽을 수 없습니다.", http.StatusInternalServerError)
		return
	}
	mimeType := http.DetectContentType(prefix[:count])
	if !supportedMediaPreviewType(mimeType) {
		http.Error(w, "미리 볼 수 없는 파일 형식입니다.", http.StatusUnsupportedMediaType)
		return
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		http.Error(w, "미디어 파일을 읽을 수 없습니다.", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", mimeType)
	// ServeContent preserves the original bytes and supports HEAD and video
	// Range requests without loading the complete file into memory.
	http.ServeContent(w, r, info.Name(), info.ModTime(), file)
}

func singleMediaPreviewValue(query url.Values, key string) (string, bool) {
	values, exists := query[key]
	if !exists || len(values) != 1 || values[0] == "" {
		return "", false
	}
	return values[0], true
}

func (a *App) openMediaPreview(source, token string) (*os.File, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if token != "" {
		media, exists := a.boardMedia[token]
		if !exists {
			return nil, os.ErrNotExist
		}
		return os.Open(media.Path)
	}
	root, err := a.rootLocked()
	if err != nil {
		return nil, err
	}
	resolved, _, err := resolveBoardMediaPath(root, source, "미디어 미리보기")
	if err != nil {
		return nil, err
	}
	return os.Open(resolved)
}

func supportedMediaPreviewType(mimeType string) bool {
	if _, exists := supportedBoardImageTypes[mimeType]; exists {
		return true
	}
	switch mimeType {
	case "image/x-icon", "video/mp4", "video/webm", "video/avi":
		return true
	default:
		return false
	}
}
