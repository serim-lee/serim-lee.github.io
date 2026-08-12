package main

import (
	"bytes"
	"context"
	"image"
	"image/png"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestMediaPreviewServesOriginalSavedImageWithoutChanges(t *testing.T) {
	root := newBoardRepoFixture(t)
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, image.NewRGBA(image.Rect(0, 0, 800, 600))); err != nil {
		t.Fatal(err)
	}
	source := "nested/미리 보기.png"
	if err := os.Mkdir(filepath.Join(root, "data", "media", "nested"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFixtureFile(t, filepath.Join(root, "data", "media", filepath.FromSlash(source)), encoded.Bytes())
	before := snapshotBoardRepo(t, root)
	app := &App{repoRoot: root, dirty: true}
	response := requestMediaPreview(app, http.MethodGet, "src="+url.QueryEscape(source), "")
	if response.Code != http.StatusOK {
		t.Fatalf("preview status = %d: %s", response.Code, response.Body.String())
	}
	if !bytes.Equal(response.Body.Bytes(), encoded.Bytes()) {
		t.Fatal("preview did not preserve original image bytes")
	}
	configuration, _, err := image.DecodeConfig(bytes.NewReader(response.Body.Bytes()))
	if err != nil || configuration.Width != 800 || configuration.Height != 600 {
		t.Fatalf("preview dimensions = %+v, error = %v; want original 800x600", configuration, err)
	}
	assertMediaPreviewHeaders(t, response, "image/png")
	if !app.dirty {
		t.Error("preview cleared the dirty state")
	}
	assertBoardRepoSnapshot(t, root, before, "media preview")
}

func TestMediaPreviewServesStagedImageAndRejectsRevokedToken(t *testing.T) {
	root := newBoardRepoFixture(t)
	app := &App{repoRoot: root}
	t.Cleanup(func() { app.shutdown(context.Background()) })
	original := fixturePNGBytes()
	input := filepath.Join(t.TempDir(), "new.png")
	writeFixtureFile(t, input, original)
	before := snapshotBoardRepo(t, root)
	staged, err := app.StageBoardMedia([]string{input})
	if err != nil || len(staged.Items) != 1 {
		t.Fatalf("staging failed: %+v, %v", staged, err)
	}
	token := staged.Items[0].StageToken
	response := requestMediaPreview(app, http.MethodGet, "stage_token="+token, "")
	if response.Code != http.StatusOK || !bytes.Equal(response.Body.Bytes(), original) {
		t.Fatalf("staged preview status = %d, original bytes preserved = %v", response.Code, bytes.Equal(response.Body.Bytes(), original))
	}
	assertMediaPreviewHeaders(t, response, "image/png")
	if len(app.boardMedia) != 1 || app.dirty {
		t.Error("preview changed staging or dirty state")
	}
	assertBoardRepoSnapshot(t, root, before, "staged media preview")
	if err := app.DiscardBoardMedia([]string{token}); err != nil {
		t.Fatal(err)
	}
	response = requestMediaPreview(app, http.MethodGet, "stage_token="+token, "")
	if response.Code != http.StatusNotFound {
		t.Fatalf("revoked stage token status = %d, want 404", response.Code)
	}
	assertBoardRepoSnapshot(t, root, before, "revoked media preview")
}

func TestMediaPreviewVideoSupportsRangeAndHead(t *testing.T) {
	root := newBoardRepoFixture(t)
	// ISO BMFF ftyp signature recognized by net/http as video/mp4. Payload
	// bytes exercise streaming/ranges without requiring an installed encoder.
	original := append([]byte("\x00\x00\x00\x18ftypmp42\x00\x00\x00\x00mp42isom"), bytes.Repeat([]byte("video-data"), 100)...)
	writeFixtureFile(t, filepath.Join(root, "data", "media", "clip.mp4"), original)
	app := &App{repoRoot: root}
	response := requestMediaPreview(app, http.MethodGet, "src=clip.mp4", "bytes=10-29")
	if response.Code != http.StatusPartialContent || !bytes.Equal(response.Body.Bytes(), original[10:30]) {
		t.Fatalf("video range status = %d, body = %q", response.Code, response.Body.Bytes())
	}
	assertMediaPreviewHeaders(t, response, "video/mp4")
	if response.Header().Get("Content-Range") != "bytes 10-29/"+strconv.Itoa(len(original)) {
		t.Errorf("unexpected content range %q", response.Header().Get("Content-Range"))
	}
	response = requestMediaPreview(app, http.MethodHead, "src=clip.mp4", "")
	if response.Code != http.StatusOK || response.Body.Len() != 0 {
		t.Fatalf("HEAD status = %d, body length = %d", response.Code, response.Body.Len())
	}
	if response.Header().Get("Content-Length") != strconv.Itoa(len(original)) {
		t.Error("HEAD did not return the original media size")
	}
	assertMediaPreviewHeaders(t, response, "video/mp4")
}

func TestMediaPreviewRejectsInvalidRequestsWithoutChanges(t *testing.T) {
	root := newBoardRepoFixture(t)
	writeFixtureFile(t, filepath.Join(root, "data", "media", "fake.png"), []byte("<html>not an image</html>"))
	writeFixtureFile(t, filepath.Join(root, "data", "media", "script.svg"), []byte(`<svg xmlns="http://www.w3.org/2000/svg"><script>alert(1)</script></svg>`))
	writeFixtureFile(t, filepath.Join(root, "data", "outside.png"), fixturePNGBytes())
	before := snapshotBoardRepo(t, root)
	app := &App{repoRoot: root}
	tests := []struct {
		name, method, query string
		status              int
	}{
		{"no source", "GET", "", 400},
		{"empty source", "GET", "src=", 400},
		{"empty token", "GET", "stage_token=", 400},
		{"duplicate source", "GET", "src=a&src=b", 400},
		{"duplicate token", "GET", "stage_token=a&stage_token=b", 400},
		{"both source kinds", "GET", "src=a&stage_token=b", 400},
		{"unknown key", "GET", "path=a", 400},
		{"extra key", "GET", "src=a&extra=b", 400},
		{"bad encoding", "GET", "src=%zz", 400},
		{"bad separator", "GET", "src=a;b", 400},
		{"missing file", "GET", "src=missing.png", 404},
		{"missing token", "GET", "stage_token=missing", 404},
		{"traversal", "GET", "src=..%2Foutside.png", 404},
		{"noncanonical path", "GET", "src=nested%2F..%2Ffake.png", 404},
		{"absolute path", "GET", "src=%2Foutside.png", 404},
		{"windows absolute path", "GET", "src=C%3A%5Coutside.png", 404},
		{"backslash traversal", "GET", "src=..%5Coutside.png", 404},
		{"not an image", "GET", "src=fake.png", 415},
		{"active SVG rejected", "GET", "src=script.svg", 415},
		{"write method", "POST", "src=fake.png", 405},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := requestMediaPreview(app, test.method, test.query, "")
			if response.Code != test.status {
				t.Errorf("status = %d, want %d: %s", response.Code, test.status, response.Body.String())
			}
			if strings.Contains(response.Body.String(), root) {
				t.Error("error response leaked the filesystem root")
			}
			if response.Header().Get("Cache-Control") != "no-store" || response.Header().Get("X-Content-Type-Options") != "nosniff" {
				t.Error("error response is missing safe cache/type headers")
			}
		})
	}
	response := httptest.NewRecorder()
	app.mediaPreviewHandler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/unknown?src=fake.png", nil))
	if response.Code != http.StatusNotFound {
		t.Errorf("unrelated path status = %d, want 404", response.Code)
	}
	if app.dirty || len(app.boardMedia) != 0 || app.stagingDir != "" {
		t.Error("invalid preview requests changed editor state")
	}
	assertBoardRepoSnapshot(t, root, before, "rejected media previews")
}

func TestMediaPreviewRejectsSymlinkOutsideMediaRoot(t *testing.T) {
	root := newBoardRepoFixture(t)
	outside := filepath.Join(t.TempDir(), "outside.png")
	writeFixtureFile(t, outside, fixturePNGBytes())
	if err := os.Symlink(outside, filepath.Join(root, "data", "media", "escape.png")); err != nil {
		t.Skipf("symlink creation unavailable: %v", err)
	}
	response := requestMediaPreview(&App{repoRoot: root}, http.MethodGet, "src=escape.png", "")
	if response.Code != http.StatusNotFound {
		t.Fatalf("outside-media symlink status = %d, want 404", response.Code)
	}
}

func requestMediaPreview(app *App, method, query, rangeHeader string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, editorMediaPreviewPath+"?"+query, nil)
	if rangeHeader != "" {
		request.Header.Set("Range", rangeHeader)
	}
	response := httptest.NewRecorder()
	app.mediaPreviewHandler().ServeHTTP(response, request)
	return response
}

func assertMediaPreviewHeaders(t *testing.T, response *httptest.ResponseRecorder, mimeType string) {
	t.Helper()
	for name, want := range map[string]string{
		"Content-Type": mimeType, "Cache-Control": "no-store", "X-Content-Type-Options": "nosniff", "Accept-Ranges": "bytes",
	} {
		if got := response.Header().Get(name); got != want {
			t.Errorf("%s = %q, want %q", name, got, want)
		}
	}
}
