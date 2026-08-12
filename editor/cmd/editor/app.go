package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Gonie-Gonie/Gonie-Gonie.github.io/editor/internal/dataversion"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"unicode"
	"unicode/utf8"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Frozen compatibility counter for pre-bundle settings; new formats use data/version.json.
const settingsSchemaVersion = 9
const defaultColorTheme = "sage"

var supportedMainPageSections = []string{
	"experience",
	"education",
	"scholarships",
	"certifications",
	"awards",
	"academic_activities",
	"patents",
	"teaching",
	"skills",
}

var supportedCVSections = []string{
	"experience",
	"education",
	"projects",
	"publications",
	"software",
	"awards",
	"academic_activities",
	"patents",
	"teaching",
	"scholarships",
	"certifications",
	"skills",
}

var supportedWebPages = []string{"publications", "projects", "software", "board"}

type BilingualLabel struct {
	LabelEN string `json:"label_en"`
	LabelKO string `json:"label_ko"`

	extraFields map[string]json.RawMessage
}

type TaxonomyItem struct {
	ID    string `json:"id,omitempty"`
	Color string `json:"color,omitempty"`
	BilingualLabel

	extraFields map[string]json.RawMessage
	// Older clients omit color, while an explicit empty value resets to the default palette.
	colorProvided bool
}

func (item *TaxonomyItem) UnmarshalJSON(raw []byte) error {
	type taxonomyItemJSON TaxonomyItem
	var decoded taxonomyItemJSON
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return err
	}
	*item = TaxonomyItem(decoded)
	_, item.colorProvided = fields["color"]
	return nil
}

type SettingsDocument struct {
	SchemaVersion          int                    `json:"schema_version,omitempty"`
	Version                string                 `json:"version,omitempty"`
	ColorTheme             string                 `json:"color_theme"`
	CustomThemes           map[string]CustomTheme `json:"custom_themes,omitempty"`
	EditorShortcuts        map[string]string      `json:"editor_shortcuts,omitempty"`
	MainPageSections       []string               `json:"main_page_sections"`
	HiddenMainPageSections []string               `json:"hidden_main_page_sections"`
	CVSections             []string               `json:"cv_sections"`
	HiddenCVSections       []string               `json:"hidden_cv_sections"`
	HiddenWebPages         []string               `json:"hidden_web_pages"`
	Topics                 []TaxonomyItem         `json:"topics"`
	TopicFallback          BilingualLabel         `json:"topic_fallback"`
	Journals               []PublicationVenue     `json:"journals"`
	Conferences            []PublicationVenue     `json:"conferences"`

	extraFields map[string]json.RawMessage
}

var taxonomyIDPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
var taxonomyColorPattern = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

var reservedPublicationTopicIDs = map[string]bool{
	"__all_topics__":     true,
	"__fallback_topic__": true,
}

type SettingsUsage struct {
	ProjectThemes     map[string]int `json:"project_themes"`
	PublicationTopics map[string]int `json:"publication_topics"`
	Journals          map[string]int `json:"journals"`
	Conferences       map[string]int `json:"conferences"`
}

type SettingsResponse struct {
	Settings SettingsDocument `json:"settings"`
	Revision string           `json:"revision"`
	Usage    SettingsUsage    `json:"usage"`
	Location string           `json:"location"`
}

type SaveSettingsRequest struct {
	Settings SettingsDocument `json:"settings"`
	Revision string           `json:"revision"`
}

type App struct {
	mu                    sync.Mutex
	operationBusy         atomic.Bool
	updateClosing         atomic.Bool
	updateLockError       error
	updateRecoveryStatus  string
	updateRecoveryMessage string
	progressSink          func(OperationProgress)
	siteBuilder           func(context.Context, string) error
	ctx                   context.Context
	repoRoot              string
	startupErr            error
	dirty                 bool
	stagingDir            string
	boardMedia            map[string]stagedBoardMedia
	documentStagingDir    string
	documents             map[string]stagedDocument
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	root, err := discoverRepoRoot()
	a.mu.Lock()
	a.ctx = ctx
	if a.repoRoot == "" {
		a.repoRoot = root
	}
	a.startupErr = err
	a.mu.Unlock()
}

// SetDirty mirrors the frontend draft state so the native window-close hook can
// warn before discarding edits. It never writes a file.
func (a *App) SetDirty(dirty bool) {
	a.mu.Lock()
	a.dirty = dirty
	a.mu.Unlock()
}

func (a *App) beforeClose(ctx context.Context) bool {
	if a.updateClosing.Load() {
		return false
	}
	if a.operationBusy.Load() {
		return true
	}
	a.mu.Lock()
	dirty := a.dirty || len(a.boardMedia) > 0
	a.mu.Unlock()
	if !dirty {
		return false
	}

	choice, err := runtime.MessageDialog(ctx, runtime.MessageDialogOptions{
		Type:          runtime.QuestionDialog,
		Title:         "저장하지 않은 변경",
		Message:       "저장하지 않은 변경 내용이 있습니다. 저장하지 않고 닫을까요?",
		DefaultButton: "No",
		CancelButton:  "No",
	})
	if err != nil {
		return true
	}
	return choice != "Yes"
}

func (a *App) LoadSettings() (SettingsResponse, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	root, err := a.rootLocked()
	if err != nil {
		return SettingsResponse{}, err
	}
	settingsPath := filepath.Join(root, "data", "settings.json")
	document, raw, err := readSettings(settingsPath)
	if err != nil {
		return SettingsResponse{}, err
	}
	if err := normaliseLoadedSettings(&document); err != nil {
		return SettingsResponse{}, err
	}
	usage, err := loadUsage(root)
	if err != nil {
		return SettingsResponse{}, err
	}
	if err := validateReferences(document, usage); err != nil {
		return SettingsResponse{}, err
	}

	a.dirty = false
	return SettingsResponse{
		Settings: document,
		Revision: revisionOf(raw),
		Usage:    usage,
		Location: filepath.ToSlash(filepath.Join("data", "settings.json")),
	}, nil
}

func (a *App) SaveSettings(request SaveSettingsRequest) (SettingsResponse, error) {
	if !a.operationBusy.CompareAndSwap(false, true) {
		return SettingsResponse{}, errors.New("다른 작업이 진행 중입니다")
	}
	defer a.operationBusy.Store(false)
	a.mu.Lock()
	defer a.mu.Unlock()

	root, err := a.rootLocked()
	if err != nil {
		return SettingsResponse{}, err
	}
	settingsPath := filepath.Join(root, "data", "settings.json")
	current, raw, err := readSettings(settingsPath)
	if err != nil {
		return SettingsResponse{}, err
	}
	if request.Revision == "" || request.Revision != revisionOf(raw) {
		return SettingsResponse{}, errors.New("settings.json이 편집기 밖에서 변경되었습니다. 다시 불러온 뒤 수정해 주세요")
	}
	if err := normaliseLoadedSettings(&current); err != nil {
		return SettingsResponse{}, err
	}
	usage, err := loadUsage(root)
	if err != nil {
		return SettingsResponse{}, err
	}

	normalised, err := normaliseForSave(request.Settings, current, usage)
	if err != nil {
		return SettingsResponse{}, err
	}
	encoded, err := marshalSettings(normalised)
	if err != nil {
		return SettingsResponse{}, fmt.Errorf("settings.json 인코딩 실패: %w", err)
	}
	if err := requireUnchangedRevision(settingsPath, request.Revision, "settings.json"); err != nil {
		return SettingsResponse{}, err
	}
	if err := dataversion.ValidateBundle(root); err != nil {
		return SettingsResponse{}, err
	}
	if err := writeFileAtomically(settingsPath, encoded); err != nil {
		return SettingsResponse{}, fmt.Errorf("settings.json 저장 실패: %w", err)
	}

	a.dirty = false
	return SettingsResponse{
		Settings: normalised,
		Revision: revisionOf(encoded),
		Usage:    usage,
		Location: filepath.ToSlash(filepath.Join("data", "settings.json")),
	}, nil
}

func (a *App) rootLocked() (string, error) {
	if a.repoRoot != "" {
		if err := dataversion.ValidateBundle(a.repoRoot); err != nil {
			return "", err
		}
		return a.repoRoot, nil
	}
	if a.startupErr != nil {
		return "", a.startupErr
	}
	root, err := discoverRepoRoot()
	if err != nil {
		return "", err
	}
	a.repoRoot = root
	if err := dataversion.ValidateBundle(root); err != nil {
		return "", err
	}
	return root, nil
}

func discoverRepoRoot() (string, error) {
	var starts []string
	if executable, err := os.Executable(); err == nil {
		starts = append(starts, filepath.Dir(executable))
	}
	if cwd, err := os.Getwd(); err == nil {
		starts = append(starts, cwd)
	}

	seen := make(map[string]bool)
	for _, start := range starts {
		current, err := filepath.Abs(start)
		if err != nil {
			continue
		}
		for {
			key := strings.ToLower(filepath.Clean(current))
			if !seen[key] {
				seen[key] = true
				if isRepoRoot(current) {
					return current, nil
				}
			}
			parent := filepath.Dir(current)
			if parent == current {
				break
			}
			current = parent
		}
	}
	return "", errors.New("저장소 루트를 찾지 못했습니다. Profile-Editor.exe를 저장소 최상위에서 실행해 주세요")
}

func isRepoRoot(root string) bool {
	markers := []string{
		filepath.Join("data", "settings.json"),
		filepath.Join("data", "profile.json"),
		filepath.Join("data", "people.json"),
		filepath.Join("data", "projects.json"),
		filepath.Join("data", "software.json"),
		filepath.Join("data", "publications.json"),
		filepath.Join("data", "awards.json"),
		filepath.Join("data", "board.json"),
		filepath.Join("data", "academic_activities.json"),
		"index.html",
	}
	for _, marker := range markers {
		info, err := os.Stat(filepath.Join(root, marker))
		if err != nil || info.IsDir() {
			return false
		}
	}
	return true
}

func readSettings(path string) (SettingsDocument, []byte, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return SettingsDocument{}, nil, fmt.Errorf("settings.json 읽기 실패: %w", err)
	}
	fields, info, err := dataversion.Object(raw, "settings.json")
	if err != nil {
		return SettingsDocument{}, nil, err
	}
	normalized, err := json.Marshal(fields)
	if err != nil {
		return SettingsDocument{}, nil, err
	}
	var document SettingsDocument
	if err := json.Unmarshal(normalized, &document); err != nil {
		return SettingsDocument{}, nil, err
	}
	document.Version = info.Version
	document.SchemaVersion = info.LegacySchema
	if document.SchemaVersion == 0 {
		document.SchemaVersion = settingsSchemaVersion
	}

	if err := captureSettingsExtraFields(&document, raw); err != nil {
		return SettingsDocument{}, nil, fmt.Errorf("settings.json 추가 필드 분석 실패: %w", err)
	}
	return document, raw, nil
}

func captureSettingsExtraFields(document *SettingsDocument, raw []byte) error {
	root, err := decodeObject(raw)
	if err != nil {
		return err
	}
	document.extraFields = withoutKnownFields(root,
		"schema_version",
		"version",
		"color_theme",
		"custom_themes",
		"editor_shortcuts",
		"main_page_sections",
		"hidden_main_page_sections",
		"cv_sections",
		"hidden_cv_sections",
		"hidden_web_pages",
		"topics",
		"topic_fallback",
		"project_themes",
		"project_theme_fallback",
		"publication_topics",
		"publication_topic_fallback",
		"journals",
		"conferences",
	)
	for _, key := range []string{"journals", "conferences"} {
		if value, exists := root[key]; exists && bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return fmt.Errorf("%s는 JSON 배열이어야 합니다", key)
		}
	}

	if err := captureUnifiedTopics(document, root); err != nil {
		return err
	}
	return nil
}

func captureTaxonomyExtraFields(items []TaxonomyItem, raw json.RawMessage) error {
	var rows []json.RawMessage
	if err := json.Unmarshal(raw, &rows); err != nil {
		return err
	}
	if len(rows) != len(items) {
		return errors.New("항목 수를 확인할 수 없습니다")
	}
	for index, row := range rows {
		fields, err := decodeObject(row)
		if err != nil {
			return fmt.Errorf("%d번째 항목: %w", index+1, err)
		}
		items[index].extraFields = withoutKnownFields(fields, "id", "label_en", "label_ko", "color")
	}
	return nil
}

func decodeObject(raw []byte) (map[string]json.RawMessage, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return nil, err
	}
	if fields == nil {
		return nil, errors.New("JSON 객체여야 합니다")
	}
	return fields, nil
}

func withoutKnownFields(fields map[string]json.RawMessage, known ...string) map[string]json.RawMessage {
	knownSet := make(map[string]bool, len(known))
	for _, name := range known {
		knownSet[name] = true
	}
	extras := make(map[string]json.RawMessage)
	for name, value := range fields {
		if !knownSet[name] {
			extras[name] = cloneRawMessage(value)
		}
	}
	return extras
}

func cloneRawMessage(value json.RawMessage) json.RawMessage {
	return append(json.RawMessage(nil), value...)
}

func cloneExtraFields(fields map[string]json.RawMessage) map[string]json.RawMessage {
	if len(fields) == 0 {
		return nil
	}
	cloned := make(map[string]json.RawMessage, len(fields))
	for name, value := range fields {
		cloned[name] = cloneRawMessage(value)
	}
	return cloned
}

func normaliseLoadedSettings(document *SettingsDocument) error {
	loadedSchemaVersion := document.SchemaVersion
	if document.Version != "" {
		if err := dataversion.Validate(document.Version); err != nil {
			return err
		}
		if loadedSchemaVersion == 0 {
			loadedSchemaVersion = settingsSchemaVersion
		}
	}
	if loadedSchemaVersion != 3 && loadedSchemaVersion != 4 && loadedSchemaVersion != 5 && loadedSchemaVersion != 6 && loadedSchemaVersion != 7 && loadedSchemaVersion != 8 && loadedSchemaVersion != settingsSchemaVersion {
		return fmt.Errorf("지원하지 않는 settings.json schema_version입니다: %d", document.SchemaVersion)
	}
	colorTheme, err := normaliseColorTheme(document.ColorTheme)
	if err != nil {
		return err
	}
	document.ColorTheme = colorTheme
	if document.CustomThemes, err = normaliseCustomThemes(document.CustomThemes, nil); err != nil {
		return err
	}
	if document.EditorShortcuts, err = normaliseEditorShortcuts(document.EditorShortcuts, nil); err != nil {
		return err
	}
	if document.Journals, err = normalisePublicationCatalog(document.Journals, nil, "journals"); err != nil {
		return err
	}
	if document.Conferences, err = normalisePublicationCatalog(document.Conferences, nil, "conferences"); err != nil {
		return err
	}
	if document.HiddenWebPages, err = normaliseHiddenWebPages(document.HiddenWebPages); err != nil {
		return err
	}
	if document.HiddenMainPageSections == nil {
		visible := make(map[string]bool, len(document.MainPageSections))
		for _, section := range document.MainPageSections {
			visible[section] = true
		}
		document.HiddenMainPageSections = make([]string, 0)
		for _, section := range supportedMainPageSections {
			if !visible[section] {
				document.HiddenMainPageSections = append(document.HiddenMainPageSections, section)
			}
		}
	}
	if document.MainPageSections == nil {
		document.MainPageSections = []string{}
	}
	// schema v6 adds academic activities to the main-page partition. Older
	// settings remain editable and start with this new section hidden until a
	// save with at least one activity explicitly exposes it.
	if loadedSchemaVersion <= 5 &&
		!stringSliceContains(document.MainPageSections, "academic_activities") &&
		!stringSliceContains(document.HiddenMainPageSections, "academic_activities") {
		document.HiddenMainPageSections = append(document.HiddenMainPageSections, "academic_activities")
	}
	if loadedSchemaVersion <= 8 &&
		!stringSliceContains(document.MainPageSections, "patents") &&
		!stringSliceContains(document.HiddenMainPageSections, "patents") {
		document.HiddenMainPageSections = append(document.HiddenMainPageSections, "patents")
	}
	// schema v4 predates CV configuration. Treat every supported CV section as
	// visible, while preserving an explicitly configured v5+ partition.
	if document.CVSections == nil && document.HiddenCVSections == nil {
		document.CVSections = []string{}
		for _, section := range supportedCVSections {
			if loadedSchemaVersion > 8 || section != "patents" {
				document.CVSections = append(document.CVSections, section)
			}
		}
		document.HiddenCVSections = []string{}
	} else {
		if document.CVSections == nil {
			document.CVSections = []string{}
		}
		if document.HiddenCVSections == nil {
			document.HiddenCVSections = []string{}
		}
	}
	if loadedSchemaVersion <= 8 &&
		!stringSliceContains(document.CVSections, "patents") &&
		!stringSliceContains(document.HiddenCVSections, "patents") {
		document.HiddenCVSections = append(document.HiddenCVSections, "patents")
	}
	if document.Topics == nil {
		return errors.New("topics는 JSON 배열이어야 합니다")
	}
	document.SchemaVersion = settingsSchemaVersion
	if document.Version == "" {
		document.Version = dataversion.Current
	}
	if err := validateSectionPartition(document.MainPageSections, document.HiddenMainPageSections); err != nil {
		return err
	}
	if err := validateNamedSectionPartition(document.CVSections, document.HiddenCVSections, supportedCVSections, "CV"); err != nil {
		return err
	}
	if err := validateLoadedTaxonomy(document.Topics, "주제", reservedPublicationTopicIDs); err != nil {
		return err
	}
	if err := validateLoadedLabel(document.TopicFallback, "기타 분류"); err != nil {
		return err
	}
	return nil
}

func stringSliceContains(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func validateLoadedTaxonomy(items []TaxonomyItem, label string, reservedIDs map[string]bool) error {
	seenIDs := make(map[string]bool, len(items))
	seenEN := make(map[string]bool, len(items))
	seenKO := make(map[string]bool, len(items))
	for index, item := range items {
		context := fmt.Sprintf("%s %d", label, index+1)
		if item.ID == "" {
			return fmt.Errorf("%s의 ID가 비어 있습니다", context)
		}
		if reservedIDs[item.ID] {
			return fmt.Errorf("%s의 ID '%s'는 화면 필터용 예약 ID라 사용할 수 없습니다", context, item.ID)
		}
		if !taxonomyIDPattern.MatchString(item.ID) {
			return fmt.Errorf("%s의 ID '%s'는 소문자 영문, 숫자, 단일 하이픈만 사용할 수 있습니다", context, item.ID)
		}
		if seenIDs[item.ID] {
			return fmt.Errorf("%s ID가 중복되었습니다: %s", label, item.ID)
		}
		if err := validateLoadedLabel(item.BilingualLabel, context); err != nil {
			return err
		}
		if err := validateTaxonomyColor(item.Color, context); err != nil {
			return err
		}
		enKey := strings.ToLower(item.LabelEN)
		koKey := strings.ToLower(item.LabelKO)
		if seenEN[enKey] {
			return fmt.Errorf("%s 영문 이름이 중복되었습니다: %s", label, item.LabelEN)
		}
		if seenKO[koKey] {
			return fmt.Errorf("%s 국문 이름이 중복되었습니다: %s", label, item.LabelKO)
		}
		seenIDs[item.ID] = true
		seenEN[enKey] = true
		seenKO[koKey] = true
	}
	return nil
}

func validateLoadedLabel(input BilingualLabel, context string) error {
	validated, err := normaliseLabel(input, context)
	if err != nil {
		return err
	}
	if validated.LabelEN != input.LabelEN || validated.LabelKO != input.LabelKO {
		return fmt.Errorf("%s 이름의 앞뒤 공백을 제거해 주세요", context)
	}
	return nil
}

func normaliseForSave(input, current SettingsDocument, usage SettingsUsage) (SettingsDocument, error) {
	if input.Version != "" {
		if err := dataversion.Validate(input.Version); err != nil {
			return SettingsDocument{}, err
		}
	}
	colorThemeInput := input.ColorTheme
	// Older clients omit this optional setting; keep the current selection.
	if colorThemeInput == "" {
		colorThemeInput = current.ColorTheme
	}
	colorTheme, err := normaliseColorTheme(colorThemeInput)
	if err != nil {
		return SettingsDocument{}, err
	}
	customThemes, err := normaliseCustomThemes(input.CustomThemes, current.CustomThemes)
	if err != nil {
		return SettingsDocument{}, err
	}
	shortcutsInput := input.EditorShortcuts
	if shortcutsInput == nil {
		shortcutsInput = current.EditorShortcuts
	}
	shortcuts, err := normaliseEditorShortcuts(shortcutsInput, current.EditorShortcuts)
	if err != nil {
		return SettingsDocument{}, err
	}
	hiddenWebPagesInput := input.HiddenWebPages
	if hiddenWebPagesInput == nil {
		hiddenWebPagesInput = current.HiddenWebPages
	}
	hiddenWebPages, err := normaliseHiddenWebPages(hiddenWebPagesInput)
	if err != nil {
		return SettingsDocument{}, err
	}
	if err := validateSectionPartition(input.MainPageSections, input.HiddenMainPageSections); err != nil {
		return SettingsDocument{}, err
	}
	if err := validateNamedSectionPartition(input.CVSections, input.HiddenCVSections, supportedCVSections, "CV"); err != nil {
		return SettingsDocument{}, err
	}
	topics, err := normaliseTaxonomyWithReserved(input.Topics, current.Topics, combinedTopicUsage(usage), "주제", reservedPublicationTopicIDs)
	if err != nil {
		return SettingsDocument{}, err
	}
	fallbackInput := input.TopicFallback
	fallbackInput.extraFields = cloneExtraFields(current.TopicFallback.extraFields)
	fallback, err := normaliseLabel(fallbackInput, "기타 분류")
	if err != nil {
		return SettingsDocument{}, err
	}
	journals, err := normalisePublicationCatalog(input.Journals, current.Journals, "journals")
	if err != nil {
		return SettingsDocument{}, err
	}
	conferences, err := normalisePublicationCatalog(input.Conferences, current.Conferences, "conferences")
	if err != nil {
		return SettingsDocument{}, err
	}
	document := SettingsDocument{
		Version:                settingsVersion(current),
		SchemaVersion:          settingsSchemaVersion,
		ColorTheme:             colorTheme,
		CustomThemes:           customThemes,
		EditorShortcuts:        shortcuts,
		MainPageSections:       append([]string{}, input.MainPageSections...),
		HiddenMainPageSections: append([]string{}, input.HiddenMainPageSections...),
		CVSections:             append([]string{}, input.CVSections...),
		HiddenCVSections:       append([]string{}, input.HiddenCVSections...),
		HiddenWebPages:         hiddenWebPages,
		Topics:                 topics,
		TopicFallback:          fallback,
		Journals:               journals,
		Conferences:            conferences,
		extraFields:            cloneExtraFields(current.extraFields),
	}
	if err := validateReferences(document, usage); err != nil {
		return SettingsDocument{}, err
	}
	return document, nil
}

func normaliseColorTheme(value string) (string, error) {
	switch value {
	case "":
		return defaultColorTheme, nil
	case "sage", "ocean", "violet", "rose", "sand", "slate",
		"mint", "teal", "sky", "indigo", "amber", "graphite", "custom-1", "custom-2", "custom-3":
		return value, nil
	default:
		return "", fmt.Errorf("color_theme 색상 테마가 올바르지 않습니다: %s", value)
	}
}

func normaliseHiddenWebPages(pages []string) ([]string, error) {
	seen := make(map[string]bool, len(pages))
	for _, page := range pages {
		if !stringSliceContains(supportedWebPages, page) {
			return nil, fmt.Errorf("hidden_web_pages의 Web 페이지가 올바르지 않습니다: %s", page)
		}
		if seen[page] {
			return nil, fmt.Errorf("hidden_web_pages에 중복된 Web 페이지가 있습니다: %s", page)
		}
		seen[page] = true
	}
	return append([]string{}, pages...), nil
}

func validateSectionPartition(visible, hidden []string) error {
	return validateNamedSectionPartition(visible, hidden, supportedMainPageSections, "프로필")
}

func validateNamedSectionPartition(visible, hidden, supported []string, label string) error {
	configured := append(append([]string{}, visible...), hidden...)
	if len(configured) != len(supported) {
		return fmt.Errorf("%s 노출/숨김 목록에는 지원 섹션이 각각 한 번씩 있어야 합니다", label)
	}
	allowed := make(map[string]bool, len(supported))
	for _, section := range supported {
		allowed[section] = true
	}
	seen := make(map[string]bool, len(configured))
	for _, section := range configured {
		if !allowed[section] || seen[section] {
			return fmt.Errorf("잘못되거나 중복된 %s 섹션입니다: %s", label, section)
		}
		seen[section] = true
	}
	return nil
}

func normaliseTaxonomy(input, current []TaxonomyItem, usage map[string]int, label string) ([]TaxonomyItem, error) {
	return normaliseTaxonomyWithReserved(input, current, usage, label, nil)
}

func normaliseTaxonomyWithReserved(input, current []TaxonomyItem, usage map[string]int, label string, reservedIDs map[string]bool) ([]TaxonomyItem, error) {
	currentByID := make(map[string]TaxonomyItem, len(current))
	usedIDs := make(map[string]bool, len(current)+len(usage))
	for _, item := range current {
		currentByID[item.ID] = item
		usedIDs[item.ID] = true
	}
	for id := range usage {
		if id != "" {
			usedIDs[id] = true
		}
	}

	result := make([]TaxonomyItem, 0, len(input))
	seenIDs := make(map[string]bool, len(input))
	seenEN := make(map[string]bool, len(input))
	seenKO := make(map[string]bool, len(input))
	for index, item := range input {
		labels, err := normaliseLabel(item.BilingualLabel, fmt.Sprintf("%s %d", label, index+1))
		if err != nil {
			return nil, err
		}
		id := strings.TrimSpace(item.ID)
		if id == "" {
			id, err = newTaxonomyID(labels.LabelEN, usedIDs)
			if err != nil {
				return nil, err
			}
		} else if _, exists := currentByID[id]; !exists {
			return nil, fmt.Errorf("%s ID는 직접 만들거나 변경할 수 없습니다", label)
		}
		if reservedIDs[id] {
			return nil, fmt.Errorf("%s ID '%s'는 화면 필터용 예약 ID라 사용할 수 없습니다", label, id)
		}
		if !taxonomyIDPattern.MatchString(id) {
			return nil, fmt.Errorf("%s ID '%s'는 소문자 영문, 숫자, 단일 하이픈만 사용할 수 있습니다", label, id)
		}
		if seenIDs[id] {
			return nil, fmt.Errorf("%s ID가 중복되었습니다: %s", label, id)
		}
		enKey := strings.ToLower(labels.LabelEN)
		koKey := strings.ToLower(labels.LabelKO)
		if seenEN[enKey] {
			return nil, fmt.Errorf("%s 영문 이름이 중복되었습니다: %s", label, labels.LabelEN)
		}
		if seenKO[koKey] {
			return nil, fmt.Errorf("%s 국문 이름이 중복되었습니다: %s", label, labels.LabelKO)
		}
		seenIDs[id] = true
		seenEN[enKey] = true
		seenKO[koKey] = true
		usedIDs[id] = true
		extras := map[string]json.RawMessage(nil)
		color := strings.TrimSpace(item.Color)
		if existing, exists := currentByID[id]; exists {
			extras = cloneExtraFields(existing.extraFields)
			if item.Color == "" && !item.colorProvided {
				color = existing.Color
			}
		}
		if err := validateTaxonomyColor(color, fmt.Sprintf("%s %d", label, index+1)); err != nil {
			return nil, err
		}
		result = append(result, TaxonomyItem{
			ID: id, Color: color, BilingualLabel: labels, extraFields: extras,
			colorProvided: item.colorProvided || color != "",
		})
	}

	for _, item := range current {
		if !seenIDs[item.ID] && usage[item.ID] > 0 {
			return nil, fmt.Errorf("%s '%s'은(는) %d개 데이터에서 사용 중이라 삭제할 수 없습니다", label, item.LabelKO, usage[item.ID])
		}
	}
	return result, nil
}

func validateTaxonomyColor(color, context string) error {
	if color != "" && !taxonomyColorPattern.MatchString(color) {
		return fmt.Errorf("%s 색상은 #RRGGBB 형식의 6자리 HEX 코드여야 합니다", context)
	}
	return nil
}

func normaliseLabel(input BilingualLabel, context string) (BilingualLabel, error) {
	label := BilingualLabel{
		LabelEN:     strings.TrimSpace(input.LabelEN),
		LabelKO:     strings.TrimSpace(input.LabelKO),
		extraFields: cloneExtraFields(input.extraFields),
	}
	if label.LabelEN == "" || label.LabelKO == "" {
		return BilingualLabel{}, fmt.Errorf("%s의 영문/국문 이름을 모두 입력해 주세요", context)
	}
	for _, value := range []string{label.LabelEN, label.LabelKO} {
		if utf8.RuneCountInString(value) > 100 {
			return BilingualLabel{}, fmt.Errorf("%s 이름은 100자 이하여야 합니다", context)
		}
		if strings.IndexFunc(value, unicode.IsControl) >= 0 {
			return BilingualLabel{}, fmt.Errorf("%s 이름에는 제어 문자를 사용할 수 없습니다", context)
		}
	}
	return label, nil
}

func newTaxonomyID(labelEN string, used map[string]bool) (string, error) {
	var builder strings.Builder
	lastHyphen := false
	for _, char := range strings.ToLower(labelEN) {
		switch {
		case char >= 'a' && char <= 'z', char >= '0' && char <= '9':
			builder.WriteRune(char)
			lastHyphen = false
		case builder.Len() > 0 && !lastHyphen:
			builder.WriteByte('-')
			lastHyphen = true
		}
	}
	base := strings.Trim(builder.String(), "-")
	if base == "" {
		var random [8]byte
		if _, err := rand.Read(random[:]); err != nil {
			return "", fmt.Errorf("새 분류 ID 생성 실패: %w", err)
		}
		base = "item-" + hex.EncodeToString(random[:])
	}
	candidate := base
	for suffix := 2; used[candidate]; suffix++ {
		candidate = fmt.Sprintf("%s-%d", base, suffix)
	}
	return candidate, nil
}

func loadUsage(root string) (SettingsUsage, error) {
	usage := SettingsUsage{
		ProjectThemes:     make(map[string]int),
		PublicationTopics: make(map[string]int),
		Journals:          make(map[string]int),
		Conferences:       make(map[string]int),
	}
	var projects []struct {
		Theme string `json:"theme"`
	}
	if err := readJSONArray(filepath.Join(root, "data", "projects.json"), &projects); err != nil {
		return SettingsUsage{}, err
	}
	for _, project := range projects {
		if project.Theme != "" {
			usage.ProjectThemes[project.Theme]++
		}
	}

	var publications []struct {
		Topic        string `json:"topic"`
		JournalID    string `json:"journal_id"`
		ConferenceID string `json:"conference_id"`
	}
	if err := readJSONArray(filepath.Join(root, "data", "publications.json"), &publications); err != nil {
		return SettingsUsage{}, err
	}
	for _, publication := range publications {
		if publication.Topic != "" {
			usage.PublicationTopics[publication.Topic]++
		}
		if publication.JournalID != "" {
			usage.Journals[publication.JournalID]++
		}
		if publication.ConferenceID != "" {
			usage.Conferences[publication.ConferenceID]++
		}
	}
	return usage, nil
}

func readJSONArray(path string, destination any) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	rows, err := dataversion.Array(raw, filepath.Base(path))
	if err != nil {
		return err
	}
	encoded, err := json.Marshal(rows)
	if err != nil {
		return err
	}
	return json.Unmarshal(encoded, destination)
}

func validateReferences(document SettingsDocument, usage SettingsUsage) error {
	projectIDs := make(map[string]bool, len(document.Topics))
	for _, item := range document.Topics {
		projectIDs[item.ID] = true
	}
	for id, count := range usage.ProjectThemes {
		if id != "" && count > 0 && !projectIDs[id] {
			return fmt.Errorf("projects.json의 %d개 항목이 존재하지 않는 프로젝트 테마 '%s'을(를) 참조합니다", count, id)
		}
	}
	publicationIDs := make(map[string]bool, len(document.Topics))
	for _, item := range document.Topics {
		publicationIDs[item.ID] = true
	}
	for id, count := range usage.PublicationTopics {
		if id != "" && count > 0 && !publicationIDs[id] {
			return fmt.Errorf("publications.json의 %d개 항목이 존재하지 않는 논문 주제 '%s'을(를) 참조합니다", count, id)
		}
	}
	return validatePublicationCatalogReferences(document, usage)
}

func marshalSettings(document SettingsDocument) ([]byte, error) {
	if document.Version == "" {
		document.Version = dataversion.Current
	}
	if err := dataversion.Validate(document.Version); err != nil {
		return nil, err
	}
	topics, err := marshalTaxonomy(document.Topics)
	if err != nil {
		return nil, fmt.Errorf("topics: %w", err)
	}
	fallback, err := mergeObjectFields(document.TopicFallback.extraFields, map[string]any{
		"label_en": document.TopicFallback.LabelEN,
		"label_ko": document.TopicFallback.LabelKO,
	})
	if err != nil {
		return nil, fmt.Errorf("topic_fallback: %w", err)
	}
	root, err := mergeObjectFields(document.extraFields, map[string]any{
		"color_theme":               document.ColorTheme,
		"main_page_sections":        document.MainPageSections,
		"hidden_main_page_sections": document.HiddenMainPageSections,
		"cv_sections":               document.CVSections,
		"hidden_cv_sections":        document.HiddenCVSections,
		"hidden_web_pages":          append([]string{}, document.HiddenWebPages...),
		"topics":                    topics,
		"topic_fallback":            fallback,
		"journals":                  document.Journals,
		"conferences":               document.Conferences,
	})
	if err != nil {
		return nil, err
	}
	if document.CustomThemes != nil {
		root, err = mergeObjectFields(root, map[string]any{"custom_themes": document.CustomThemes})
		if err != nil {
			return nil, err
		}
	}
	if document.EditorShortcuts != nil {
		root, err = mergeObjectFields(root, map[string]any{"editor_shortcuts": document.EditorShortcuts})
		if err != nil {
			return nil, err
		}
	}
	raw, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(raw, '\n'), nil
}

func marshalTaxonomy(items []TaxonomyItem) ([]json.RawMessage, error) {
	rows := make([]json.RawMessage, 0, len(items))
	for _, item := range items {
		known := map[string]any{
			"id":       item.ID,
			"label_en": item.LabelEN,
			"label_ko": item.LabelKO,
		}
		if item.Color != "" {
			known["color"] = item.Color
		}
		row, err := mergeObjectFields(item.extraFields, known)
		if err != nil {
			return nil, err
		}
		encoded, err := json.Marshal(row)
		if err != nil {
			return nil, err
		}
		rows = append(rows, encoded)
	}
	return rows, nil
}

func mergeObjectFields(extras map[string]json.RawMessage, known map[string]any) (map[string]json.RawMessage, error) {
	fields := cloneExtraFields(extras)
	if fields == nil {
		fields = make(map[string]json.RawMessage, len(known))
	}
	for name, value := range known {
		encoded, err := json.Marshal(value)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		fields[name] = encoded
	}
	return fields, nil
}

func revisionOf(raw []byte) string {
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func writeFileAtomically(path string, contents []byte) (err error) {
	directory := filepath.Dir(path)
	temporary, err := os.CreateTemp(directory, "."+filepath.Base(path)+"-*.tmp")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer func() {
		_ = temporary.Close()
		if err != nil {
			_ = os.Remove(temporaryPath)
		}
	}()
	if info, statErr := os.Stat(path); statErr == nil {
		if chmodErr := temporary.Chmod(info.Mode().Perm()); chmodErr != nil {
			return chmodErr
		}
	}
	if _, err = temporary.Write(contents); err != nil {
		return err
	}
	if err = temporary.Sync(); err != nil {
		return err
	}
	if err = temporary.Close(); err != nil {
		return err
	}
	if err = replaceFile(temporaryPath, path); err != nil {
		return err
	}
	return nil
}
