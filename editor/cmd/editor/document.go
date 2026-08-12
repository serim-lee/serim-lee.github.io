package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"
)

const maxDocumentSize = 20 << 20

type stagedDocument struct {
	Path     string
	Filename string
}

type StagedDocumentItem struct {
	StageToken   string `json:"stage_token"`
	Filename     string `json:"filename"`
	OriginalName string `json:"original_name"`
	Size         int64  `json:"size"`
}

type StageDocumentResponse struct {
	Items    []StagedDocumentItem     `json:"items"`
	Rejected []RejectedBoardMediaItem `json:"rejected"`
}

func validateDocumentFilename(filename string) error {
	if filename == "" {
		return nil
	}
	if !utf8.ValidString(filename) || len(filename) > 240 || strings.TrimSpace(filename) != filename ||
		strings.ContainsAny(filename, `<>:"/\|?*`) || strings.HasSuffix(filename, ".") ||
		strings.IndexFunc(filename, unicode.IsControl) >= 0 || !strings.EqualFold(filepath.Ext(filename), ".pdf") {
		return errors.New("data/document 안의 PDF 파일명만 입력해 주세요")
	}
	stem := strings.ToUpper(strings.SplitN(filename, ".", 2)[0])
	if stem == "CON" || stem == "PRN" || stem == "AUX" || stem == "NUL" ||
		(len(stem) == 4 && (strings.HasPrefix(stem, "COM") || strings.HasPrefix(stem, "LPT")) && stem[3] >= '1' && stem[3] <= '9') {
		return errors.New("이 PDF 파일명은 사용할 수 없습니다")
	}
	return nil
}

func checkDocumentFile(path string) error {
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() {
		return errors.New("PDF 파일을 찾을 수 없습니다")
	}
	if info.Size() < 8 || info.Size() > maxDocumentSize {
		return errors.New("PDF 파일은 20MB 이하여야 합니다")
	}
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	var prefix [5]byte
	if _, err := io.ReadFull(file, prefix[:]); err != nil || string(prefix[:]) != "%PDF-" {
		return errors.New("PDF 형식의 파일만 추가할 수 있습니다")
	}
	if _, err := file.Seek(max(int64(0), info.Size()-1024), io.SeekStart); err != nil {
		return err
	}
	tail, err := io.ReadAll(file)
	if err != nil {
		return err
	}
	if !bytes.Contains(tail, []byte("%%EOF")) {
		return errors.New("완전한 PDF 파일을 추가해 주세요")
	}
	return nil
}

// Explorer drops stay private until the JSON and document files are saved together.
func (a *App) StageDocument(paths []string) (StageDocumentResponse, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	response := StageDocumentResponse{Items: []StagedDocumentItem{}, Rejected: []RejectedBoardMediaItem{}}
	if len(paths) != 1 {
		return response, errors.New("각 항목에는 PDF 한 개를 드롭해 주세요")
	}
	if len(a.documents) >= 100 {
		return response, errors.New("저장 전 PDF는 최대 100개까지 보관할 수 있습니다")
	}
	root, err := a.rootLocked()
	if err != nil {
		return response, err
	}
	source := paths[0]
	name := filepath.Base(source)
	reject := func(err error) (StageDocumentResponse, error) {
		response.Rejected = append(response.Rejected, RejectedBoardMediaItem{OriginalName: name, Reason: err.Error()})
		return response, nil
	}
	if !filepath.IsAbs(source) {
		return reject(errors.New("절대경로 파일만 추가할 수 있습니다"))
	}
	if err := validateDocumentFilename(name); err != nil {
		return reject(err)
	}
	if err := checkDocumentFile(source); err != nil {
		return reject(err)
	}
	if a.documentStagingDir == "" {
		a.documentStagingDir, err = os.MkdirTemp("", "profile-editor-document-")
		if err != nil {
			return response, err
		}
	}
	token, err := randomBoardToken()
	if err != nil {
		return response, err
	}
	stagedPath := filepath.Join(a.documentStagingDir, token+".pdf")
	input, err := os.Open(source)
	if err != nil {
		return reject(err)
	}
	defer input.Close()
	output, err := os.OpenFile(stagedPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return response, err
	}
	written, copyErr := io.Copy(output, io.LimitReader(input, maxDocumentSize+1))
	if copyErr == nil {
		copyErr = output.Sync()
	}
	closeErr := output.Close()
	if copyErr == nil {
		copyErr = closeErr
	}
	if copyErr == nil {
		copyErr = checkDocumentFile(stagedPath)
	}
	if copyErr != nil {
		_ = os.Remove(stagedPath)
		return reject(copyErr)
	}
	used := map[string]bool{}
	entries, err := os.ReadDir(filepath.Join(root, "data", "document"))
	if err != nil && !os.IsNotExist(err) {
		_ = os.Remove(stagedPath)
		return response, err
	}
	for _, entry := range entries {
		used[strings.ToLower(entry.Name())] = true
	}
	for _, document := range a.documents {
		used[strings.ToLower(document.Filename)] = true
	}
	filename := name
	if used[strings.ToLower(filename)] {
		filename = strings.TrimSuffix(name, filepath.Ext(name)) + "-" + token[:8] + ".pdf"
	}
	if err := validateDocumentFilename(filename); err != nil {
		_ = os.Remove(stagedPath)
		return reject(err)
	}
	if a.documents == nil {
		a.documents = map[string]stagedDocument{}
	}
	a.documents[token] = stagedDocument{Path: stagedPath, Filename: filename}
	response.Items = append(response.Items, StagedDocumentItem{StageToken: token, Filename: filename, OriginalName: name, Size: written})
	return response, nil
}

func documentReferences(profile profileSnapshot, awards awardSnapshot) map[string]bool {
	references := map[string]bool{}
	items, _ := profile.Data["certifications"].([]any)
	for _, value := range items {
		item, _ := value.(map[string]any)
		if filename, _ := item["pdf"].(string); filename != "" {
			references[filename] = true
		}
	}
	for _, row := range awards.Rows {
		if row.Item.PDF != "" {
			references[row.Item.PDF] = true
		}
	}
	return references
}

func (a *App) prepareDocumentsLocked(root string, profile profileSnapshot, awards awardSnapshot, request SaveEditorDataRequest) ([]pendingBoardMedia, error) {
	pending := []pendingBoardMedia{}
	references := documentReferences(profile, awards)
	changed := map[string]bool{}
	if request.SaveProfile {
		for name := range documentReferences(profile, awardSnapshot{}) {
			changed[name] = true
		}
	}
	if request.SaveAwards {
		for name := range documentReferences(profileSnapshot{Data: map[string]any{}}, awards) {
			changed[name] = true
		}
	}
	used := map[string]bool{}
	for _, token := range request.DocumentStageTokens {
		document, exists := a.documents[token]
		if !exists || used[document.Filename] || !changed[document.Filename] {
			return nil, errors.New("저장할 항목에 연결된 PDF만 저장할 수 있습니다")
		}
		if err := checkDocumentFile(document.Path); err != nil {
			return nil, err
		}
		used[document.Filename] = true
		pending = append(pending, pendingBoardMedia{Token: token, StagedPath: document.Path, DestinationPath: filepath.Join(root, "data", "document", document.Filename)})
	}
	for filename := range references {
		if err := validateDocumentFilename(filename); err != nil {
			return nil, err
		}
		if !used[filename] {
			if err := checkDocumentFile(filepath.Join(root, "data", "document", filename)); err != nil {
				return nil, fmt.Errorf("PDF %s: %w", filename, err)
			}
		}
	}
	return pending, nil
}

func (a *App) DiscardDocuments(tokens []string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.cleanupDocumentsLocked(tokens)
}

func (a *App) cleanupDocumentsLocked(tokens []string) error {
	if tokens == nil {
		for token := range a.documents {
			tokens = append(tokens, token)
		}
	}
	var failures []error
	for _, token := range tokens {
		document, exists := a.documents[token]
		if !exists {
			continue
		}
		if err := os.Remove(document.Path); err != nil && !os.IsNotExist(err) {
			failures = append(failures, err)
			continue
		}
		delete(a.documents, token)
	}
	if len(a.documents) == 0 && a.documentStagingDir != "" {
		if err := os.Remove(a.documentStagingDir); err != nil && !os.IsNotExist(err) {
			failures = append(failures, err)
		} else {
			a.documentStagingDir = ""
		}
	}
	return errors.Join(failures...)
}
