package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Gonie-Gonie/Gonie-Gonie.github.io/editor/internal/dataversion"
	"image"
	"image/color"
	stddraw "image/draw"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	pathpkg "path"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	_ "golang.org/x/image/bmp"
	xdraw "golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
)

const (
	maxBoardImagesPerDrop = 30
	maxBoardImageSize     = 20 << 20
	maxBoardImagesTotal   = 60
	maxBoardPreviewPixels = 25_000_000
)

var supportedBoardImageTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/gif":  ".gif",
	"image/webp": ".webp",
	"image/bmp":  ".bmp",
}

type stagedBoardMedia struct {
	Path         string
	OriginalName string
	MIMEType     string
	Extension    string
	Size         int64
}

type BoardItemSummary struct {
	EditorKey  string              `json:"editor_key"`
	StartDate  string              `json:"start_date"`
	EndDate    string              `json:"end_date"`
	TitleEN    string              `json:"title_en"`
	TitleKO    string              `json:"title_ko"`
	ContentEN  string              `json:"content_en"`
	ContentKO  string              `json:"content_ko"`
	Media      []BoardMediaSummary `json:"media"`
	MediaCount int                 `json:"media_count"`
}

type BoardMediaSummary struct {
	EditorKey    string `json:"editor_key"`
	Src          string `json:"src"`
	Type         string `json:"type,omitempty"`
	Poster       string `json:"poster,omitempty"`
	CaptionEN    string `json:"caption_en"`
	CaptionKO    string `json:"caption_ko"`
	OriginalName string `json:"original_name"`
	Size         int64  `json:"size"`
	PreviewURL   string `json:"preview_url,omitempty"`
}

type EditorDataResponse struct {
	Settings                   SettingsDocument              `json:"settings"`
	SettingsRevision           string                        `json:"settings_revision"`
	Usage                      SettingsUsage                 `json:"usage"`
	SettingsLocation           string                        `json:"settings_location"`
	Profile                    json.RawMessage               `json:"profile"`
	ProfileMedia               ProfileMediaSummary           `json:"profile_media"`
	ProfileRevision            string                        `json:"profile_revision"`
	ProfileLocation            string                        `json:"profile_location"`
	People                     []PeopleItemSummary           `json:"people"`
	PeopleRevision             string                        `json:"people_revision"`
	PeopleLocation             string                        `json:"people_location"`
	Awards                     []AwardItemSummary            `json:"awards"`
	AwardsRevision             string                        `json:"awards_revision"`
	AwardsLocation             string                        `json:"awards_location"`
	AcademicActivities         []AcademicActivityItemSummary `json:"academic_activities"`
	AcademicActivitiesRevision string                        `json:"academic_activities_revision"`
	AcademicActivitiesLocation string                        `json:"academic_activities_location"`
	Publications               []PublicationItemSummary      `json:"publications"`
	PublicationsRevision       string                        `json:"publications_revision"`
	PublicationsLocation       string                        `json:"publications_location"`
	Projects                   []ProjectItemSummary          `json:"projects"`
	ProjectsRevision           string                        `json:"projects_revision"`
	ProjectsLocation           string                        `json:"projects_location"`
	Software                   []SoftwareItemSummary         `json:"software"`
	SoftwareRevision           string                        `json:"software_revision"`
	SoftwareLocation           string                        `json:"software_location"`
	Board                      []BoardItemSummary            `json:"board"`
	BoardRevision              string                        `json:"board_revision"`
	BoardLocation              string                        `json:"board_location"`
}

type BoardMediaInput struct {
	EditorKey  string `json:"editor_key,omitempty"`
	StageToken string `json:"stage_token"`
	CaptionEN  string `json:"caption_en"`
	CaptionKO  string `json:"caption_ko"`
}

type BoardPostInput struct {
	StartDate string            `json:"start_date"`
	EndDate   string            `json:"end_date"`
	TitleEN   string            `json:"title_en"`
	TitleKO   string            `json:"title_ko"`
	ContentEN string            `json:"content_en"`
	ContentKO string            `json:"content_ko"`
	Media     []BoardMediaInput `json:"media"`
}

// These aliases keep the original new-post API source-compatible while the
// unified Post payload is adopted by the frontend.
type NewBoardMedia = BoardMediaInput
type NewBoardPost = BoardPostInput

type BoardSaveItem struct {
	EditorKey string          `json:"editor_key,omitempty"`
	Post      *BoardPostInput `json:"post,omitempty"`
	NewPost   *NewBoardPost   `json:"new_post,omitempty"`
}

type SaveEditorDataRequest struct {
	OperationID                string                     `json:"operation_id,omitempty"`
	Settings                   SettingsDocument           `json:"settings"`
	SettingsRevision           string                     `json:"settings_revision"`
	SaveSettings               bool                       `json:"save_settings"`
	Profile                    json.RawMessage            `json:"profile"`
	ProfileRevision            string                     `json:"profile_revision"`
	SaveProfile                bool                       `json:"save_profile"`
	ProfileMediaStageToken     string                     `json:"profile_media_stage_token,omitempty"`
	RemoveProfileMedia         bool                       `json:"remove_profile_media,omitempty"`
	DocumentStageTokens        []string                   `json:"document_stage_tokens,omitempty"`
	People                     []PersonSaveItem           `json:"people"`
	PeopleRevision             string                     `json:"people_revision"`
	SavePeople                 bool                       `json:"save_people"`
	Awards                     []AwardSaveItem            `json:"awards"`
	AwardsRevision             string                     `json:"awards_revision"`
	SaveAwards                 bool                       `json:"save_awards"`
	AcademicActivities         []AcademicActivitySaveItem `json:"academic_activities"`
	AcademicActivitiesRevision string                     `json:"academic_activities_revision"`
	SaveAcademicActivities     bool                       `json:"save_academic_activities"`
	Publications               []PublicationSaveItem      `json:"publications"`
	PublicationsRevision       string                     `json:"publications_revision"`
	SavePublications           bool                       `json:"save_publications"`
	Projects                   []ProjectSaveItem          `json:"projects"`
	ProjectsRevision           string                     `json:"projects_revision"`
	SaveProjects               bool                       `json:"save_projects"`
	Software                   []SoftwareSaveItem         `json:"software"`
	SoftwareRevision           string                     `json:"software_revision"`
	SaveSoftware               bool                       `json:"save_software"`
	Board                      []BoardSaveItem            `json:"board"`
	BoardRevision              string                     `json:"board_revision"`
	SaveBoard                  bool                       `json:"save_board"`
}

type StagedBoardMediaItem struct {
	StageToken   string `json:"stage_token"`
	OriginalName string `json:"original_name"`
	MIMEType     string `json:"mime_type"`
	Size         int64  `json:"size"`
	PreviewURL   string `json:"preview_url,omitempty"`
}

type RejectedBoardMediaItem struct {
	OriginalName string `json:"original_name"`
	Reason       string `json:"reason"`
}

type StageBoardMediaResponse struct {
	Items    []StagedBoardMediaItem   `json:"items"`
	Rejected []RejectedBoardMediaItem `json:"rejected"`
}

type boardDocumentMedia struct {
	Src       string `json:"src"`
	Type      string `json:"type,omitempty"`
	Poster    string `json:"poster,omitempty"`
	CaptionEN string `json:"caption_en,omitempty"`
	CaptionKO string `json:"caption_ko,omitempty"`

	editorKey    string
	raw          json.RawMessage
	originalName string
	size         int64
	previewURL   string
}

type boardDocumentItem struct {
	StartDate string               `json:"start_date"`
	EndDate   string               `json:"end_date"`
	TitleEN   string               `json:"title_en"`
	TitleKO   string               `json:"title_ko"`
	ContentEN string               `json:"content_en"`
	ContentKO string               `json:"content_ko"`
	Media     []boardDocumentMedia `json:"media"`
}

type boardRow struct {
	Key  string
	Raw  json.RawMessage
	Item boardDocumentItem
}

type boardSnapshot struct {
	Raw  []byte
	Rows []boardRow
}

type pendingBoardMedia struct {
	Token           string
	StagedPath      string
	DestinationPath string
}

// LoadEditorData reads every editable data source afresh. Nothing is written.
func (a *App) LoadEditorData() (EditorDataResponse, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	root, err := a.rootLocked()
	if err != nil {
		return EditorDataResponse{}, err
	}
	settingsPath := filepath.Join(root, "data", "settings.json")
	settings, settingsRaw, err := readSettings(settingsPath)
	if err != nil {
		return EditorDataResponse{}, err
	}
	if err := normaliseLoadedSettings(&settings); err != nil {
		return EditorDataResponse{}, err
	}
	profile, err := readProfile(filepath.Join(root, "data", "profile.json"))
	if err != nil {
		return EditorDataResponse{}, err
	}
	people, err := readPeople(filepath.Join(root, "data", "people.json"))
	if err != nil {
		return EditorDataResponse{}, err
	}
	if err := validateProfilePeopleReferences(profile, peopleIDSet(people)); err != nil {
		return EditorDataResponse{}, err
	}
	awards, err := readAwards(filepath.Join(root, "data", "awards.json"))
	if err != nil {
		return EditorDataResponse{}, err
	}
	academicActivities, err := readAcademicActivities(filepath.Join(root, "data", "academic_activities.json"))
	if err != nil {
		return EditorDataResponse{}, err
	}
	projects, err := readProjects(filepath.Join(root, "data", "projects.json"), root, projectThemeIDs(settings))
	if err != nil {
		return EditorDataResponse{}, err
	}
	software, err := readSoftware(filepath.Join(root, "data", "software.json"), root)
	if err != nil {
		return EditorDataResponse{}, err
	}
	publications, err := readPublications(
		filepath.Join(root, "data", "publications.json"),
		peopleIDSet(people),
		awardIDSet(awards),
		settingsPublicationTopicIDs(settings),
	)
	if err != nil {
		return EditorDataResponse{}, err
	}
	board, err := readBoard(filepath.Join(root, "data", "board.json"), root)
	if err != nil {
		return EditorDataResponse{}, err
	}

	if err := a.cleanupBoardStagingLocked(nil); err != nil {
		return EditorDataResponse{}, fmt.Errorf("이전 사진 임시 파일 정리 실패: %w", err)
	}
	usage := usageFromContent(projects, publications)
	if err := validatePublicationCatalogReferences(settings, usage); err != nil {
		return EditorDataResponse{}, err
	}
	a.dirty = false
	return editorDataResponse(root, settings, settingsRaw, usage, profile, people, awards, academicActivities, publications, projects, software, board)
}

// SaveEditorData validates every loaded revision before committing the editor's
// JSON files. Newly dropped photos are published before their JSON references,
// so the site can never observe a new reference to a missing file.
func (a *App) SaveEditorData(request SaveEditorDataRequest) (EditorDataResponse, error) {
	if !a.operationBusy.CompareAndSwap(false, true) {
		return EditorDataResponse{}, errors.New("다른 작업이 진행 중입니다")
	}
	defer a.operationBusy.Store(false)
	a.mu.Lock()
	defer a.mu.Unlock()
	a.reportProgress(request.OperationID, "변경 사항 확인 중…", -1)

	if !request.SaveSettings && !request.SaveProfile && !request.SavePeople && !request.SaveAwards &&
		!request.SaveAcademicActivities && !request.SavePublications && !request.SaveProjects && !request.SaveSoftware && !request.SaveBoard {
		return EditorDataResponse{}, errors.New("저장할 변경 내용이 없습니다")
	}
	if (strings.TrimSpace(request.ProfileMediaStageToken) != "" || request.RemoveProfileMedia) && !request.SaveProfile {
		return EditorDataResponse{}, errors.New("프로필 이미지 변경은 profile.json 저장과 함께 요청해야 합니다")
	}
	root, err := a.rootLocked()
	if err != nil {
		return EditorDataResponse{}, err
	}
	settingsPath := filepath.Join(root, "data", "settings.json")
	profilePath := filepath.Join(root, "data", "profile.json")
	peoplePath := filepath.Join(root, "data", "people.json")
	awardsPath := filepath.Join(root, "data", "awards.json")
	academicActivitiesPath := filepath.Join(root, "data", "academic_activities.json")
	publicationsPath := filepath.Join(root, "data", "publications.json")
	projectsPath := filepath.Join(root, "data", "projects.json")
	softwarePath := filepath.Join(root, "data", "software.json")
	boardPath := filepath.Join(root, "data", "board.json")

	currentSettings, settingsRaw, err := readSettings(settingsPath)
	if err != nil {
		return EditorDataResponse{}, err
	}
	if request.SettingsRevision == "" || request.SettingsRevision != revisionOf(settingsRaw) {
		return EditorDataResponse{}, errors.New("settings.json이 편집기 밖에서 변경되었습니다. 다시 불러온 뒤 수정해 주세요")
	}
	if err := normaliseLoadedSettings(&currentSettings); err != nil {
		return EditorDataResponse{}, err
	}
	currentProfile, err := readProfile(profilePath)
	if err != nil {
		return EditorDataResponse{}, err
	}
	if request.ProfileRevision == "" || request.ProfileRevision != revisionOf(currentProfile.Raw) {
		return EditorDataResponse{}, errors.New("profile.json이 편집기 밖에서 변경되었습니다. 다시 불러온 뒤 수정해 주세요")
	}
	currentPeople, err := readPeople(peoplePath)
	if err != nil {
		return EditorDataResponse{}, err
	}
	if request.PeopleRevision == "" || request.PeopleRevision != revisionOf(currentPeople.Raw) {
		return EditorDataResponse{}, errors.New("people.json이 편집기 밖에서 변경되었습니다. 다시 불러온 뒤 수정해 주세요")
	}
	if err := validateProfilePeopleReferences(currentProfile, peopleIDSet(currentPeople)); err != nil {
		return EditorDataResponse{}, err
	}
	currentAwards, err := readAwards(awardsPath)
	if err != nil {
		return EditorDataResponse{}, err
	}
	if request.AwardsRevision == "" || request.AwardsRevision != revisionOf(currentAwards.Raw) {
		return EditorDataResponse{}, errors.New("awards.json이 편집기 밖에서 변경되었습니다. 다시 불러온 뒤 수정해 주세요")
	}
	currentAcademicActivities, err := readAcademicActivities(academicActivitiesPath)
	if err != nil {
		return EditorDataResponse{}, err
	}
	if request.AcademicActivitiesRevision == "" || request.AcademicActivitiesRevision != revisionOf(currentAcademicActivities.Raw) {
		return EditorDataResponse{}, errors.New("academic_activities.json이 편집기 밖에서 변경되었습니다. 다시 불러온 뒤 수정해 주세요")
	}
	currentProjects, err := readProjects(projectsPath, root, projectThemeIDs(currentSettings))
	if err != nil {
		return EditorDataResponse{}, err
	}
	if request.ProjectsRevision == "" || request.ProjectsRevision != revisionOf(currentProjects.Raw) {
		return EditorDataResponse{}, errors.New("projects.json이 편집기 밖에서 변경되었습니다. 다시 불러온 뒤 수정해 주세요")
	}
	currentSoftware, err := readSoftware(softwarePath, root)
	if err != nil {
		return EditorDataResponse{}, err
	}
	if request.SoftwareRevision == "" || request.SoftwareRevision != revisionOf(currentSoftware.Raw) {
		return EditorDataResponse{}, errors.New("software.json이 편집기 밖에서 변경되었습니다. 다시 불러온 뒤 수정해 주세요")
	}
	currentPublications, err := readPublications(
		publicationsPath,
		peopleIDSet(currentPeople),
		awardIDSet(currentAwards),
		settingsPublicationTopicIDs(currentSettings),
	)
	if err != nil {
		return EditorDataResponse{}, err
	}
	if request.PublicationsRevision == "" || request.PublicationsRevision != revisionOf(currentPublications.Raw) {
		return EditorDataResponse{}, errors.New("publications.json이 편집기 밖에서 변경되었습니다. 다시 불러온 뒤 수정해 주세요")
	}
	currentBoard, err := readBoard(boardPath, root)
	if err != nil {
		return EditorDataResponse{}, err
	}
	if request.BoardRevision == "" || request.BoardRevision != revisionOf(currentBoard.Raw) {
		return EditorDataResponse{}, errors.New("board.json이 편집기 밖에서 변경되었습니다. 다시 불러온 뒤 수정해 주세요")
	}
	usage := usageFromContent(currentProjects, currentPublications)
	currentMainSections := append([]string(nil), currentSettings.MainPageSections...)
	currentHiddenMainSections := append([]string(nil), currentSettings.HiddenMainPageSections...)
	currentCVSections := append([]string(nil), currentSettings.CVSections...)
	currentHiddenCVSections := append([]string(nil), currentSettings.HiddenCVSections...)

	nextSettings := currentSettings
	nextSettingsRaw := settingsRaw
	if request.SaveSettings {
		settingsUsage := usageForEditorRequest(
			usage,
			currentProjects,
			request.Projects,
			request.SaveProjects,
			currentPublications,
			request.Publications,
			request.SavePublications,
		)
		nextSettings, err = normaliseForSave(request.Settings, currentSettings, settingsUsage)
		if err != nil {
			return EditorDataResponse{}, err
		}
	}

	nextProfile := currentProfile
	nextProfileRaw := currentProfile.Raw
	if request.SaveProfile {
		nextProfileRaw, nextProfile, err = buildProfileSave(request.Profile)
		if err != nil {
			return EditorDataResponse{}, err
		}
	}

	nextPeopleRaw := currentPeople.Raw
	nextAwardsRaw := currentAwards.Raw
	nextAcademicActivitiesRaw := currentAcademicActivities.Raw
	nextPublicationsRaw := currentPublications.Raw
	nextPeople := currentPeople
	nextAwards := currentAwards
	if request.SavePeople {
		nextPeopleRaw, err = a.buildPeopleSaveLocked(currentPeople, request.People)
		if err != nil {
			return EditorDataResponse{}, err
		}
		nextPeople, err = readPeopleBytes(nextPeopleRaw)
		if err != nil {
			return EditorDataResponse{}, fmt.Errorf("저장할 people.json 확인 실패: %w", err)
		}
	}
	// Patent inventors share the people collection with publication authors.
	// Check the next snapshots together so references can change in the same
	// transaction, while people-only saves cannot leave a dangling inventor.
	if err := validateProfilePeopleReferences(nextProfile, peopleIDSet(nextPeople)); err != nil {
		return EditorDataResponse{}, err
	}
	if request.SaveAwards {
		nextAwardsRaw, err = a.buildAwardsSaveLocked(currentAwards, request.Awards)
		if err != nil {
			return EditorDataResponse{}, err
		}
		nextAwards, err = readAwardsBytes(nextAwardsRaw)
		if err != nil {
			return EditorDataResponse{}, fmt.Errorf("저장할 awards.json 확인 실패: %w", err)
		}
	}
	if request.SaveAcademicActivities {
		nextAcademicActivitiesRaw, err = buildAcademicActivitiesSave(currentAcademicActivities, request.AcademicActivities)
		if err != nil {
			return EditorDataResponse{}, err
		}
	}
	if request.SaveSettings || request.SaveProfile || request.SaveAwards || request.SaveAcademicActivities {
		counts := profileSectionCounts(nextProfile, len(nextAwards.Rows))
		nextAcademicActivities, readErr := readAcademicActivitiesBytes(nextAcademicActivitiesRaw)
		if readErr != nil {
			return EditorDataResponse{}, fmt.Errorf("저장할 academic_activities.json 확인 실패: %w", readErr)
		}
		counts["academic_activities"] = academicActivityCount(nextAcademicActivities)
		autoHideEmptyMainSections(&nextSettings, counts)
	}
	mainPartitionChanged := !equalStrings(currentMainSections, nextSettings.MainPageSections) ||
		!equalStrings(currentHiddenMainSections, nextSettings.HiddenMainPageSections)
	writeSettings := request.SaveSettings || mainPartitionChanged
	if request.SavePublications {
		nextPublicationsRaw, err = buildPublicationSaveLocked(
			currentPublications,
			request.Publications,
			peopleIDSet(nextPeople),
			awardIDSet(nextAwards),
			settingsPublicationTopicIDs(nextSettings),
			nextSettings,
		)
		if err != nil {
			return EditorDataResponse{}, err
		}
	}
	// Always revalidate the dependent publication snapshot against the next
	// people, awards, and topic collections. This blocks deleting a referenced
	// entity unless the publication relationship is changed in the same save.
	nextPublications, err := readPublicationsBytes(
		nextPublicationsRaw,
		peopleIDSet(nextPeople),
		awardIDSet(nextAwards),
		settingsPublicationTopicIDs(nextSettings),
	)
	if err != nil {
		return EditorDataResponse{}, err
	}
	if err := validatePublicationCatalogReferences(nextSettings, usageFromContent(projectSnapshot{}, nextPublications)); err != nil {
		return EditorDataResponse{}, err
	}

	nextProjectsRaw := currentProjects.Raw
	nextSoftwareRaw := currentSoftware.Raw
	nextBoardRaw := currentBoard.Raw
	var pendingMedia []pendingBoardMedia
	var usedStageTokens []string
	claimedStageTokens := make(map[string]bool)
	if request.SaveBoard {
		var boardPending []pendingBoardMedia
		var boardTokens []string
		nextBoardRaw, boardPending, boardTokens, err = a.buildBoardSaveLocked(root, currentBoard, request.Board)
		if err != nil {
			return EditorDataResponse{}, err
		}
		pendingMedia = append(pendingMedia, boardPending...)
		usedStageTokens = append(usedStageTokens, boardTokens...)
		claimStageTokens(claimedStageTokens, boardTokens)
	}
	if request.SaveProfile && (strings.TrimSpace(request.ProfileMediaStageToken) != "" || request.RemoveProfileMedia) {
		reservedNames, reserveErr := boardMediaNames(filepath.Join(root, "data", "media"))
		if reserveErr != nil {
			return EditorDataResponse{}, reserveErr
		}
		for _, pending := range pendingMedia {
			reservedNames[strings.ToLower(filepath.Base(pending.DestinationPath))] = true
		}
		var profilePending []pendingBoardMedia
		var profileTokens []string
		nextProfileRaw, nextProfile, profilePending, profileTokens, err = a.applyProfileMediaActionLocked(
			root,
			nextProfile,
			request.ProfileMediaStageToken,
			request.RemoveProfileMedia,
			reservedNames,
			claimedStageTokens,
		)
		if err != nil {
			return EditorDataResponse{}, err
		}
		pendingMedia = append(pendingMedia, profilePending...)
		usedStageTokens = append(usedStageTokens, profileTokens...)
		claimStageTokens(claimedStageTokens, profileTokens)
	}
	if request.SaveProjects {
		var projectPending []pendingBoardMedia
		var projectTokens []string
		nextProjectsRaw, projectPending, projectTokens, err = a.buildProjectSaveLocked(
			root,
			currentProjects,
			request.Projects,
			projectThemeIDs(nextSettings),
			claimedStageTokens,
		)
		if err != nil {
			return EditorDataResponse{}, err
		}
		pendingMedia = append(pendingMedia, projectPending...)
		usedStageTokens = append(usedStageTokens, projectTokens...)
		claimStageTokens(claimedStageTokens, projectTokens)
	}
	if request.SaveSoftware {
		var softwarePending []pendingBoardMedia
		var softwareTokens []string
		nextSoftwareRaw, softwarePending, softwareTokens, err = a.buildSoftwareSaveLocked(
			root,
			currentSoftware,
			request.Software,
			claimedStageTokens,
		)
		if err != nil {
			return EditorDataResponse{}, err
		}
		pendingMedia = append(pendingMedia, softwarePending...)
		usedStageTokens = append(usedStageTokens, softwareTokens...)
		claimStageTokens(claimedStageTokens, softwareTokens)
	}
	documentPending, err := a.prepareDocumentsLocked(root, nextProfile, nextAwards, request)
	if err != nil {
		return EditorDataResponse{}, err
	}
	pendingMedia = append(pendingMedia, documentPending...)

	counts, err := cvSectionCounts(
		nextProfile,
		nextProjectsRaw,
		nextPublicationsRaw,
		nextSoftwareRaw,
		nextAwardsRaw,
		nextAcademicActivitiesRaw,
	)
	if err != nil {
		return EditorDataResponse{}, fmt.Errorf("CV 섹션 항목 확인 실패: %w", err)
	}
	affectedCVSections := make(map[string]bool, len(supportedCVSections))
	if request.SaveSettings {
		for _, section := range supportedCVSections {
			affectedCVSections[section] = true
		}
	}
	if request.SaveProfile {
		for _, section := range profileCVSections {
			affectedCVSections[section] = true
		}
	}
	if request.SaveProjects {
		affectedCVSections["projects"] = true
	}
	if request.SavePublications {
		affectedCVSections["publications"] = true
	}
	if request.SaveSoftware {
		affectedCVSections["software"] = true
	}
	if request.SaveAwards {
		affectedCVSections["awards"] = true
	}
	if request.SaveAcademicActivities {
		affectedCVSections["academic_activities"] = true
	}
	autoHideEmptyCVSections(&nextSettings, counts, affectedCVSections)
	cvPartitionChanged := !equalStrings(currentCVSections, nextSettings.CVSections) ||
		!equalStrings(currentHiddenCVSections, nextSettings.HiddenCVSections)
	writeSettings = writeSettings || cvPartitionChanged
	if writeSettings {
		nextSettingsRaw, err = marshalSettings(nextSettings)
		if err != nil {
			return EditorDataResponse{}, fmt.Errorf("settings.json 인코딩 실패: %w", err)
		}
	}

	// Recheck after all potentially expensive validation and encoding.
	if err := requireEditorRevisions(
		editorRevisionCheck{settingsPath, request.SettingsRevision, "settings.json"},
		editorRevisionCheck{profilePath, request.ProfileRevision, "profile.json"},
		editorRevisionCheck{peoplePath, request.PeopleRevision, "people.json"},
		editorRevisionCheck{awardsPath, request.AwardsRevision, "awards.json"},
		editorRevisionCheck{academicActivitiesPath, request.AcademicActivitiesRevision, "academic_activities.json"},
		editorRevisionCheck{publicationsPath, request.PublicationsRevision, "publications.json"},
		editorRevisionCheck{projectsPath, request.ProjectsRevision, "projects.json"},
		editorRevisionCheck{softwarePath, request.SoftwareRevision, "software.json"},
		editorRevisionCheck{boardPath, request.BoardRevision, "board.json"},
	); err != nil {
		return EditorDataResponse{}, err
	}

	writes := []editorFileWrite{
		{Name: "settings.json", Path: settingsPath, Revision: request.SettingsRevision, Previous: settingsRaw, Next: nextSettingsRaw, Enabled: writeSettings},
		{Name: "profile.json", Path: profilePath, Revision: request.ProfileRevision, Previous: currentProfile.Raw, Next: nextProfileRaw, Enabled: request.SaveProfile},
		{Name: "people.json", Path: peoplePath, Revision: request.PeopleRevision, Previous: currentPeople.Raw, Next: nextPeopleRaw, Enabled: request.SavePeople},
		{Name: "awards.json", Path: awardsPath, Revision: request.AwardsRevision, Previous: currentAwards.Raw, Next: nextAwardsRaw, Enabled: request.SaveAwards},
		{Name: "academic_activities.json", Path: academicActivitiesPath, Revision: request.AcademicActivitiesRevision, Previous: currentAcademicActivities.Raw, Next: nextAcademicActivitiesRaw, Enabled: request.SaveAcademicActivities},
		{Name: "publications.json", Path: publicationsPath, Revision: request.PublicationsRevision, Previous: currentPublications.Raw, Next: nextPublicationsRaw, Enabled: request.SavePublications},
		{Name: "projects.json", Path: projectsPath, Revision: request.ProjectsRevision, Previous: currentProjects.Raw, Next: nextProjectsRaw, Enabled: request.SaveProjects},
		{Name: "software.json", Path: softwarePath, Revision: request.SoftwareRevision, Previous: currentSoftware.Raw, Next: nextSoftwareRaw, Enabled: request.SaveSoftware},
		{Name: "board.json", Path: boardPath, Revision: request.BoardRevision, Previous: currentBoard.Raw, Next: nextBoardRaw, Enabled: request.SaveBoard},
	}
	fileCount := 0
	if err := dataversion.ValidateBundle(root); err != nil {
		return EditorDataResponse{}, err
	}
	for _, write := range writes {
		if write.Enabled && !bytes.Equal(write.Previous, write.Next) {
			fileCount++
		}
	}
	// Count real file writes, plus the final verification. Copy progress uses
	// the actual bytes written for the current photo, never elapsed time.
	total := len(pendingMedia) + fileCount + 1
	completed := 0
	percent := func(fraction float64) int { return int(100 * (float64(completed) + fraction) / float64(total)) }
	if len(documentPending) > 0 {
		if err := os.MkdirAll(filepath.Join(root, "data", "document"), 0o755); err != nil {
			return EditorDataResponse{}, fmt.Errorf("PDF 폴더 생성 실패: %w", err)
		}
	}
	a.reportProgress(request.OperationID, fmt.Sprintf("데이터 %d개 · 첨부 파일 %d개 저장", fileCount, len(pendingMedia)), 0)
	createdMedia, err := publishBoardMediaWithProgress(pendingMedia, func(index int, copied, size int64, done bool) {
		fraction := 0.0
		if size > 0 {
			fraction = float64(copied) / float64(size)
		}
		a.reportProgress(request.OperationID, fmt.Sprintf("첨부 파일 %d/%d개 저장 중…", index+1, len(pendingMedia)), percent(fraction))
		if done {
			completed++
		}
	})
	if err != nil {
		removeFiles(createdMedia)
		return EditorDataResponse{}, fmt.Errorf("첨부 파일 저장 실패: %w", err)
	}
	rollbackMedia := true
	defer func() {
		if rollbackMedia {
			removeFiles(createdMedia)
		}
	}()
	// Publishing a large batch can take long enough for a manual JSON edit to
	// happen. Recheck after the copies, not only before them.
	if err := requireEditorRevisions(
		editorRevisionCheck{settingsPath, request.SettingsRevision, "settings.json"},
		editorRevisionCheck{profilePath, request.ProfileRevision, "profile.json"},
		editorRevisionCheck{peoplePath, request.PeopleRevision, "people.json"},
		editorRevisionCheck{awardsPath, request.AwardsRevision, "awards.json"},
		editorRevisionCheck{academicActivitiesPath, request.AcademicActivitiesRevision, "academic_activities.json"},
		editorRevisionCheck{publicationsPath, request.PublicationsRevision, "publications.json"},
		editorRevisionCheck{projectsPath, request.ProjectsRevision, "projects.json"},
		editorRevisionCheck{softwarePath, request.SoftwareRevision, "software.json"},
		editorRevisionCheck{boardPath, request.BoardRevision, "board.json"},
	); err != nil {
		return EditorDataResponse{}, err
	}

	filesWritten := 0
	for index := range writes {
		writes[index].OnWritten = func() {
			filesWritten++
			completed++
			a.reportProgress(request.OperationID, fmt.Sprintf("데이터 %d/%d개 저장 중…", filesWritten, fileCount), percent(0))
		}
	}
	rollbackComplete, err := commitEditorFiles(writes)
	if err != nil {
		if !rollbackComplete {
			// At least one JSON file may still reference a just-published image.
			// Keep possible orphans: deleting them here could create a dangling
			// reference and lose the only durable copy of the staged media.
			rollbackMedia = false
		}
		return EditorDataResponse{}, err
	}
	rollbackMedia = false
	if err := a.cleanupDocumentsLocked(nil); err != nil {
		fmt.Fprintf(os.Stderr, "Profile-Editor temporary PDF cleanup failed: %v\n", err)
	}
	a.reportProgress(request.OperationID, "저장 결과 확인 중…", percent(0))

	if strings.TrimSpace(request.ProfileMediaStageToken) != "" || request.SaveBoard || request.SaveProjects || request.SaveSoftware {
		// Removed previews may still be staged client-side; after a successful
		// content save none of them belongs to a remaining unsaved draft.
		cleanupErr := a.cleanupBoardStagingLocked(usedStageTokens)
		cleanupErr = errors.Join(cleanupErr, a.cleanupBoardStagingLocked(nil))
		if cleanupErr != nil {
			fmt.Fprintf(os.Stderr, "Profile-Editor temporary image cleanup failed: %v\n", cleanupErr)
		}
	}
	storedProjects, err := readProjectsBytes(nextProjectsRaw, root, projectThemeIDs(nextSettings))
	if err != nil {
		return EditorDataResponse{}, fmt.Errorf("저장된 projects.json 확인 실패: %w", err)
	}
	storedSoftware, err := readSoftwareBytes(nextSoftwareRaw, root)
	if err != nil {
		return EditorDataResponse{}, fmt.Errorf("저장된 software.json 확인 실패: %w", err)
	}
	storedBoard, err := readBoardBytes(nextBoardRaw, root)
	if err != nil {
		return EditorDataResponse{}, fmt.Errorf("저장된 board.json 확인 실패: %w", err)
	}
	storedAcademicActivities, err := readAcademicActivitiesBytes(nextAcademicActivitiesRaw)
	if err != nil {
		return EditorDataResponse{}, fmt.Errorf("저장된 academic_activities.json 확인 실패: %w", err)
	}
	nextUsage := usageFromContent(storedProjects, nextPublications)
	a.dirty = false
	response, err := editorDataResponse(
		root,
		nextSettings,
		nextSettingsRaw,
		nextUsage,
		nextProfile,
		nextPeople,
		nextAwards,
		storedAcademicActivities,
		nextPublications,
		storedProjects,
		storedSoftware,
		storedBoard,
	)
	if err == nil {
		a.reportProgress(request.OperationID, "저장 완료", 100)
	}
	return response, err
}

type editorRevisionCheck struct {
	Path     string
	Revision string
	Name     string
}

type editorFileWrite struct {
	Name      string
	Path      string
	Revision  string
	Previous  []byte
	Next      []byte
	Enabled   bool
	OnWritten func()
}

func claimStageTokens(claimed map[string]bool, tokens []string) {
	for _, token := range tokens {
		claimed[token] = true
	}
}

func requireEditorRevisions(checks ...editorRevisionCheck) error {
	for _, check := range checks {
		if err := requireUnchangedRevision(check.Path, check.Revision, check.Name); err != nil {
			return err
		}
	}
	return nil
}

var commitEditorFiles = commitEditorFilesImpl

func commitEditorFilesImpl(writes []editorFileWrite) (bool, error) {
	written := make([]editorFileWrite, 0, len(writes))
	rollback := func(cause error) (bool, error) {
		result := []error{cause}
		complete := true
		for index := len(written) - 1; index >= 0; index-- {
			current := written[index]
			if err := rollbackFileIfCurrent(current.Path, current.Next, current.Previous); err != nil {
				complete = false
				result = append(result, fmt.Errorf("%s 복구 실패: %w", current.Name, err))
			}
		}
		return complete, errors.Join(result...)
	}

	for _, write := range writes {
		if !write.Enabled || bytes.Equal(write.Previous, write.Next) {
			continue
		}
		if err := requireUnchangedRevision(write.Path, write.Revision, write.Name); err != nil {
			return rollback(err)
		}
		if err := writeFileAtomically(write.Path, write.Next); err != nil {
			return rollback(fmt.Errorf("%s 저장 실패: %w", write.Name, err))
		}
		written = append(written, write)
		if write.OnWritten != nil {
			write.OnWritten()
		}
	}
	return true, nil
}

func cloneSettingsUsage(usage SettingsUsage) SettingsUsage {
	result := SettingsUsage{
		ProjectThemes:     make(map[string]int, len(usage.ProjectThemes)),
		PublicationTopics: make(map[string]int, len(usage.PublicationTopics)),
		Journals:          make(map[string]int, len(usage.Journals)),
		Conferences:       make(map[string]int, len(usage.Conferences)),
	}
	for id, count := range usage.ProjectThemes {
		result.ProjectThemes[id] = count
	}
	for id, count := range usage.PublicationTopics {
		result.PublicationTopics[id] = count
	}
	for id, count := range usage.Journals {
		result.Journals[id] = count
	}
	for id, count := range usage.Conferences {
		result.Conferences[id] = count
	}
	return result
}

func usageForEditorRequest(
	usage SettingsUsage,
	currentProjects projectSnapshot,
	projectRequest []ProjectSaveItem,
	saveProjects bool,
	currentPublications publicationSnapshot,
	publicationRequest []PublicationSaveItem,
	savePublications bool,
) SettingsUsage {
	result := cloneSettingsUsage(usage)
	if saveProjects {
		result.ProjectThemes = make(map[string]int)
		currentByKey := make(map[string]projectRow, len(currentProjects.Rows))
		for _, row := range currentProjects.Rows {
			currentByKey[row.Key] = row
		}
		for _, requested := range projectRequest {
			theme := ""
			if requested.Project != nil {
				theme = strings.TrimSpace(requested.Project.Theme)
			} else if row, exists := currentByKey[requested.EditorKey]; exists {
				theme = row.Item.Theme
			}
			if theme != "" {
				result.ProjectThemes[theme]++
			}
		}
	}
	if savePublications {
		result.PublicationTopics = make(map[string]int)
		result.Journals = make(map[string]int)
		result.Conferences = make(map[string]int)
		currentByKey := make(map[string]publicationRow, len(currentPublications.Rows))
		for _, row := range currentPublications.Rows {
			currentByKey[row.Key] = row
		}
		for _, requested := range publicationRequest {
			topic, journal, conference := "", "", ""
			if requested.Publication != nil {
				topic = strings.TrimSpace(requested.Publication.Topic)
				journal = strings.TrimSpace(requested.Publication.JournalID)
				conference = strings.TrimSpace(requested.Publication.ConferenceID)
			} else if row, exists := currentByKey[requested.EditorKey]; exists {
				topic = row.Item.Topic
				journal = row.Item.JournalID
				conference = row.Item.ConferenceID
			}
			if topic != "" {
				result.PublicationTopics[topic]++
			}
			if journal != "" {
				result.Journals[journal]++
			}
			if conference != "" {
				result.Conferences[conference]++
			}
		}
	}
	return result
}

func usageFromContent(projects projectSnapshot, publications publicationSnapshot) SettingsUsage {
	result := SettingsUsage{
		ProjectThemes:     make(map[string]int),
		PublicationTopics: make(map[string]int),
		Journals:          make(map[string]int),
		Conferences:       make(map[string]int),
	}
	for _, row := range projects.Rows {
		result.ProjectThemes[row.Item.Theme]++
	}
	for _, row := range publications.Rows {
		if row.Item.Topic != "" {
			result.PublicationTopics[row.Item.Topic]++
		}
		if row.Item.JournalID != "" {
			result.Journals[row.Item.JournalID]++
		}
		if row.Item.ConferenceID != "" {
			result.Conferences[row.Item.ConferenceID]++
		}
	}
	return result
}

func editorDataResponse(
	root string,
	settings SettingsDocument,
	settingsRaw []byte,
	usage SettingsUsage,
	profile profileSnapshot,
	people peopleSnapshot,
	awards awardSnapshot,
	academicActivities academicActivitiesSnapshot,
	publications publicationSnapshot,
	projects projectSnapshot,
	software softwareSnapshot,
	board boardSnapshot,
) (EditorDataResponse, error) {
	profileJSON, err := profileForEditor(profile)
	if err != nil {
		return EditorDataResponse{}, fmt.Errorf("profile.json 편집 데이터 준비 실패: %w", err)
	}
	items := make([]BoardItemSummary, 0, len(board.Rows))
	rows := append([]boardRow(nil), board.Rows...)
	sort.SliceStable(rows, func(left, right int) bool {
		if rows[left].Item.StartDate != rows[right].Item.StartDate {
			return rows[left].Item.StartDate > rows[right].Item.StartDate
		}
		return rows[left].Item.EndDate > rows[right].Item.EndDate
	})
	for _, row := range rows {
		media := make([]BoardMediaSummary, 0, len(row.Item.Media))
		for _, item := range row.Item.Media {
			media = append(media, BoardMediaSummary{
				EditorKey:    item.editorKey,
				Src:          item.Src,
				Type:         item.Type,
				Poster:       item.Poster,
				CaptionEN:    item.CaptionEN,
				CaptionKO:    item.CaptionKO,
				OriginalName: item.originalName,
				Size:         item.size,
				PreviewURL:   item.previewURL,
			})
		}
		items = append(items, BoardItemSummary{
			EditorKey:  row.Key,
			StartDate:  row.Item.StartDate,
			EndDate:    row.Item.EndDate,
			TitleEN:    row.Item.TitleEN,
			TitleKO:    row.Item.TitleKO,
			ContentEN:  row.Item.ContentEN,
			ContentKO:  row.Item.ContentKO,
			Media:      media,
			MediaCount: len(row.Item.Media),
		})
	}
	return EditorDataResponse{
		Settings:                   settings,
		SettingsRevision:           revisionOf(settingsRaw),
		Usage:                      usage,
		SettingsLocation:           filepath.ToSlash(filepath.Join("data", "settings.json")),
		Profile:                    profileJSON,
		ProfileMedia:               profileMediaForEditor(profile, root),
		ProfileRevision:            revisionOf(profile.Raw),
		ProfileLocation:            filepath.ToSlash(filepath.Join("data", "profile.json")),
		People:                     peopleItemSummaries(people),
		PeopleRevision:             revisionOf(people.Raw),
		PeopleLocation:             filepath.ToSlash(filepath.Join("data", "people.json")),
		Awards:                     awardItemSummaries(awards),
		AwardsRevision:             revisionOf(awards.Raw),
		AwardsLocation:             filepath.ToSlash(filepath.Join("data", "awards.json")),
		AcademicActivities:         academicActivityItemSummaries(academicActivities),
		AcademicActivitiesRevision: revisionOf(academicActivities.Raw),
		AcademicActivitiesLocation: filepath.ToSlash(filepath.Join("data", "academic_activities.json")),
		Publications:               publicationItemSummaries(publications),
		PublicationsRevision:       revisionOf(publications.Raw),
		PublicationsLocation:       filepath.ToSlash(filepath.Join("data", "publications.json")),
		Projects:                   projectItemsResponse(projects),
		ProjectsRevision:           revisionOf(projects.Raw),
		ProjectsLocation:           filepath.ToSlash(filepath.Join("data", "projects.json")),
		Software:                   softwareItemSummaries(software),
		SoftwareRevision:           revisionOf(software.Raw),
		SoftwareLocation:           filepath.ToSlash(filepath.Join("data", "software.json")),
		Board:                      items,
		BoardRevision:              revisionOf(board.Raw),
		BoardLocation:              filepath.ToSlash(filepath.Join("data", "board.json")),
	}, nil
}

func readBoard(path, root string) (boardSnapshot, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return boardSnapshot{}, fmt.Errorf("board.json 읽기 실패: %w", err)
	}
	return readBoardBytes(raw, root)
}

func readBoardBytes(raw []byte, root string) (boardSnapshot, error) {
	rows, err := dataversion.Array(raw, "board.json")
	if err != nil {
		return boardSnapshot{}, err
	}

	snapshot := boardSnapshot{Raw: append([]byte(nil), raw...), Rows: make([]boardRow, 0, len(rows))}
	for index, rowRaw := range rows {
		item, err := parseBoardItem(rowRaw, root, fmt.Sprintf("board.json %d번째 게시글", index+1))
		if err != nil {
			return boardSnapshot{}, err
		}
		hash := revisionOf(rowRaw)
		snapshot.Rows = append(snapshot.Rows, boardRow{
			Key:  fmt.Sprintf("%d:%s", index, hash[:20]),
			Raw:  append(json.RawMessage(nil), rowRaw...),
			Item: item,
		})
	}
	return snapshot, nil
}

func parseBoardItem(raw json.RawMessage, root, contextLabel string) (boardDocumentItem, error) {
	fields, err := decodeObject(raw)
	if err != nil {
		return boardDocumentItem{}, fmt.Errorf("%s은 JSON 객체여야 합니다", contextLabel)
	}
	if _, exists := fields["date"]; exists {
		return boardDocumentItem{}, fmt.Errorf("%s은 date 대신 start_date와 end_date를 사용해야 합니다", contextLabel)
	}
	if _, exists := fields["photos"]; exists {
		return boardDocumentItem{}, fmt.Errorf("%s은 photos 대신 media를 사용해야 합니다", contextLabel)
	}
	for _, field := range []string{"start_date", "end_date", "title_en", "title_ko", "content_en", "content_ko", "media"} {
		if _, exists := fields[field]; !exists {
			return boardDocumentItem{}, fmt.Errorf("%s에 %s 필드가 없습니다", contextLabel, field)
		}
	}
	var item boardDocumentItem
	if err := json.Unmarshal(raw, &item); err != nil {
		return boardDocumentItem{}, fmt.Errorf("%s 필드 형식이 올바르지 않습니다: %w", contextLabel, err)
	}
	if item.Media == nil {
		return boardDocumentItem{}, fmt.Errorf("%s의 media는 JSON 배열이어야 합니다", contextLabel)
	}
	var mediaRaw []json.RawMessage
	if err := json.Unmarshal(fields["media"], &mediaRaw); err != nil || len(mediaRaw) != len(item.Media) {
		return boardDocumentItem{}, fmt.Errorf("%s의 media 항목을 분석할 수 없습니다", contextLabel)
	}
	if err := validateBoardDates(item.StartDate, item.EndDate, contextLabel, false); err != nil {
		return boardDocumentItem{}, err
	}
	if strings.TrimSpace(item.TitleEN) == "" && strings.TrimSpace(item.TitleKO) == "" {
		return boardDocumentItem{}, fmt.Errorf("%s의 영문 또는 국문 제목을 입력해 주세요", contextLabel)
	}
	for mediaIndex, media := range item.Media {
		mediaContext := fmt.Sprintf("%s 사진 %d", contextLabel, mediaIndex+1)
		if media.Src == "" {
			return boardDocumentItem{}, fmt.Errorf("%s의 src가 비어 있습니다", mediaContext)
		}
		if media.Type != "" && media.Type != "image" && media.Type != "video" {
			return boardDocumentItem{}, fmt.Errorf("%s의 type은 image 또는 video여야 합니다", mediaContext)
		}
		resolvedSource, sourceInfo, err := resolveBoardMediaPath(root, media.Src, mediaContext)
		if err != nil {
			return boardDocumentItem{}, err
		}
		previewURL := boardThumbnailDataURL(resolvedSource)
		if media.Poster != "" {
			resolvedPoster, _, err := resolveBoardMediaPath(root, media.Poster, mediaContext+" poster")
			if err != nil {
				return boardDocumentItem{}, err
			}
			if previewURL == "" {
				previewURL = boardThumbnailDataURL(resolvedPoster)
			}
		}
		media.editorKey = fmt.Sprintf("%d:%s", mediaIndex, revisionOf(mediaRaw[mediaIndex])[:20])
		media.raw = append(json.RawMessage(nil), mediaRaw[mediaIndex]...)
		media.originalName = pathpkg.Base(media.Src)
		media.size = sourceInfo.Size()
		media.previewURL = previewURL
		item.Media[mediaIndex] = media
	}
	return item, nil
}

func validateBoardMediaPath(root, source, contextLabel string) error {
	_, _, err := resolveBoardMediaPath(root, source, contextLabel)
	return err
}

func resolveBoardMediaPath(root, source, contextLabel string) (string, os.FileInfo, error) {
	if source != strings.TrimSpace(source) || strings.Contains(source, "\\") || strings.HasPrefix(source, "/") {
		return "", nil, fmt.Errorf("%s 경로는 data/media 내부의 정규 상대경로여야 합니다", contextLabel)
	}
	cleaned := pathpkg.Clean(source)
	if cleaned != source || cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return "", nil, fmt.Errorf("%s 경로는 data/media 내부의 정규 상대경로여야 합니다", contextLabel)
	}
	mediaRoot := filepath.Join(root, "data", "media")
	target := filepath.Join(mediaRoot, filepath.FromSlash(source))
	resolvedRoot, err := filepath.EvalSymlinks(mediaRoot)
	if err != nil {
		return "", nil, fmt.Errorf("data/media 실제 경로를 확인할 수 없습니다: %w", err)
	}
	resolvedTarget, err := filepath.EvalSymlinks(target)
	if err != nil {
		return "", nil, fmt.Errorf("%s 파일을 찾을 수 없습니다: %s", contextLabel, source)
	}
	relative, err := filepath.Rel(resolvedRoot, resolvedTarget)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", nil, fmt.Errorf("%s 경로가 data/media 밖을 가리킵니다", contextLabel)
	}
	info, err := os.Stat(resolvedTarget)
	if err != nil {
		return "", nil, fmt.Errorf("%s 파일을 찾을 수 없습니다: %s", contextLabel, source)
	}
	if !info.Mode().IsRegular() {
		return "", nil, fmt.Errorf("%s 경로는 일반 파일이어야 합니다: %s", contextLabel, source)
	}
	return resolvedTarget, info, nil
}

func validateBoardDates(startDate, endDate, contextLabel string, trim bool) error {
	if trim {
		startDate = strings.TrimSpace(startDate)
		endDate = strings.TrimSpace(endDate)
	}
	start, err := parseCanonicalDate(startDate)
	if err != nil {
		return fmt.Errorf("%s의 시작일은 YYYY-MM-DD 형식의 실제 날짜여야 합니다", contextLabel)
	}
	if endDate == "" {
		return nil
	}
	end, err := parseCanonicalDate(endDate)
	if err != nil {
		return fmt.Errorf("%s의 종료일은 비워 두거나 YYYY-MM-DD 형식의 실제 날짜여야 합니다", contextLabel)
	}
	if start.After(end) {
		return fmt.Errorf("%s의 시작일은 종료일보다 늦을 수 없습니다", contextLabel)
	}
	return nil
}

func parseCanonicalDate(value string) (time.Time, error) {
	if len(value) != len("2006-01-02") {
		return time.Time{}, errors.New("날짜 길이가 올바르지 않습니다")
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil || parsed.Format("2006-01-02") != value {
		return time.Time{}, errors.New("날짜 형식이 올바르지 않습니다")
	}
	return parsed, nil
}

// StageBoardMedia copies Explorer-dropped images into a private OS temporary
// directory. The repository remains untouched until SaveEditorData succeeds.
func (a *App) StageBoardMedia(paths []string) (StageBoardMediaResponse, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	response := StageBoardMediaResponse{
		Items:    []StagedBoardMediaItem{},
		Rejected: []RejectedBoardMediaItem{},
	}
	if len(paths) == 0 {
		return response, nil
	}
	if len(paths) > maxBoardImagesPerDrop {
		return response, fmt.Errorf("한 번에 사진을 %d개까지만 추가할 수 있습니다", maxBoardImagesPerDrop)
	}
	if len(a.boardMedia)+len(paths) > maxBoardImagesTotal {
		return response, fmt.Errorf("저장 전 사진은 최대 %d개까지 보관할 수 있습니다", maxBoardImagesTotal)
	}
	if a.stagingDir == "" {
		directory, err := os.MkdirTemp("", "profile-editor-board-")
		if err != nil {
			return response, fmt.Errorf("사진 임시 폴더 생성 실패: %w", err)
		}
		a.stagingDir = directory
	}
	if a.boardMedia == nil {
		a.boardMedia = make(map[string]stagedBoardMedia)
	}

	for _, sourcePath := range paths {
		originalName := filepath.Base(sourcePath)
		item, staged, err := a.stageBoardMediaLocked(sourcePath, originalName)
		if err != nil {
			response.Rejected = append(response.Rejected, RejectedBoardMediaItem{
				OriginalName: originalName,
				Reason:       err.Error(),
			})
			continue
		}
		a.boardMedia[item.StageToken] = staged
		response.Items = append(response.Items, item)
	}
	return response, nil
}

func (a *App) stageBoardMediaLocked(sourcePath, originalName string) (StagedBoardMediaItem, stagedBoardMedia, error) {
	if !filepath.IsAbs(sourcePath) {
		return StagedBoardMediaItem{}, stagedBoardMedia{}, errors.New("절대경로 파일만 추가할 수 있습니다")
	}
	info, err := os.Lstat(sourcePath)
	if err != nil {
		return StagedBoardMediaItem{}, stagedBoardMedia{}, errors.New("파일을 읽을 수 없습니다")
	}
	if !info.Mode().IsRegular() {
		return StagedBoardMediaItem{}, stagedBoardMedia{}, errors.New("일반 이미지 파일만 추가할 수 있습니다")
	}
	if info.Size() <= 0 {
		return StagedBoardMediaItem{}, stagedBoardMedia{}, errors.New("빈 파일은 추가할 수 없습니다")
	}
	if info.Size() > maxBoardImageSize {
		return StagedBoardMediaItem{}, stagedBoardMedia{}, fmt.Errorf("파일 크기는 %dMB 이하여야 합니다", maxBoardImageSize>>20)
	}

	token, err := randomBoardToken()
	if err != nil {
		return StagedBoardMediaItem{}, stagedBoardMedia{}, fmt.Errorf("사진 토큰 생성 실패: %w", err)
	}
	stagedPath := filepath.Join(a.stagingDir, token+".stage")
	source, err := os.Open(sourcePath)
	if err != nil {
		return StagedBoardMediaItem{}, stagedBoardMedia{}, errors.New("파일을 열 수 없습니다")
	}
	defer source.Close()
	target, err := os.OpenFile(stagedPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return StagedBoardMediaItem{}, stagedBoardMedia{}, fmt.Errorf("임시 파일 생성 실패: %w", err)
	}
	written, copyErr := io.Copy(target, io.LimitReader(source, maxBoardImageSize+1))
	if copyErr == nil && written > maxBoardImageSize {
		copyErr = errors.New("복사 중 파일 크기가 제한을 초과했습니다")
	}
	if copyErr == nil {
		copyErr = target.Sync()
	}
	closeErr := target.Close()
	if copyErr == nil {
		copyErr = closeErr
	}
	if copyErr != nil {
		_ = os.Remove(stagedPath)
		return StagedBoardMediaItem{}, stagedBoardMedia{}, copyErr
	}

	mimeType, extension, err := detectBoardImage(stagedPath)
	if err != nil {
		_ = os.Remove(stagedPath)
		return StagedBoardMediaItem{}, stagedBoardMedia{}, err
	}
	previewURL := boardThumbnailDataURL(stagedPath)
	item := StagedBoardMediaItem{
		StageToken:   token,
		OriginalName: originalName,
		MIMEType:     mimeType,
		Size:         written,
		PreviewURL:   previewURL,
	}
	staged := stagedBoardMedia{
		Path:         stagedPath,
		OriginalName: originalName,
		MIMEType:     mimeType,
		Extension:    extension,
		Size:         written,
	}
	return item, staged, nil
}

func detectBoardImage(path string) (string, string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", "", err
	}
	defer file.Close()
	prefix := make([]byte, 512)
	count, err := file.Read(prefix)
	if err != nil && err != io.EOF {
		return "", "", err
	}
	prefix = prefix[:count]
	mimeType := http.DetectContentType(prefix)
	if len(prefix) >= 12 && string(prefix[:4]) == "RIFF" && string(prefix[8:12]) == "WEBP" {
		mimeType = "image/webp"
	}
	extension, supported := supportedBoardImageTypes[mimeType]
	if !supported {
		return "", "", errors.New("JPEG, PNG, GIF, WebP, BMP 사진만 추가할 수 있습니다")
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", "", err
	}
	configuration, _, err := image.DecodeConfig(file)
	if err != nil || configuration.Width <= 0 || configuration.Height <= 0 {
		return "", "", errors.New("손상되었거나 읽을 수 없는 이미지입니다")
	}
	if int64(configuration.Width)*int64(configuration.Height) > maxBoardPreviewPixels {
		return "", "", errors.New("이미지 해상도는 2,500만 픽셀 이하여야 합니다")
	}
	return mimeType, extension, nil
}

func boardThumbnailDataURL(path string) string {
	file, err := os.Open(path)
	if err != nil {
		return ""
	}
	configuration, _, err := image.DecodeConfig(file)
	_ = file.Close()
	if err != nil || configuration.Width <= 0 || configuration.Height <= 0 {
		return ""
	}
	pixels := int64(configuration.Width) * int64(configuration.Height)
	if pixels > maxBoardPreviewPixels {
		return ""
	}

	file, err = os.Open(path)
	if err != nil {
		return ""
	}
	source, _, err := image.Decode(file)
	_ = file.Close()
	if err != nil {
		return ""
	}
	const maxWidth, maxHeight = 320, 220
	width, height := source.Bounds().Dx(), source.Bounds().Dy()
	scale := min(float64(maxWidth)/float64(width), float64(maxHeight)/float64(height), 1)
	thumbnailWidth := max(1, int(float64(width)*scale))
	thumbnailHeight := max(1, int(float64(height)*scale))
	thumbnail := image.NewRGBA(image.Rect(0, 0, thumbnailWidth, thumbnailHeight))
	stddraw.Draw(thumbnail, thumbnail.Bounds(), &image.Uniform{C: color.White}, image.Point{}, stddraw.Src)
	xdraw.CatmullRom.Scale(thumbnail, thumbnail.Bounds(), source, source.Bounds(), xdraw.Over, nil)

	var encoded bytes.Buffer
	if err := jpeg.Encode(&encoded, thumbnail, &jpeg.Options{Quality: 78}); err != nil {
		return ""
	}
	return "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(encoded.Bytes())
}

func randomBoardToken() (string, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(value[:]), nil
}

// DiscardBoardMedia removes only OS-temporary staging files.
func (a *App) DiscardBoardMedia(tokens []string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.cleanupBoardStagingLocked(tokens)
}

func (a *App) shutdown(_ context.Context) {
	a.mu.Lock()
	if err := a.cleanupDocumentsLocked(nil); err != nil {
		fmt.Fprintf(os.Stderr, "Profile-Editor temporary PDF cleanup failed: %v\n", err)
	}
	if err := a.cleanupBoardStagingLocked(nil); err != nil {
		fmt.Fprintf(os.Stderr, "Profile-Editor temporary image cleanup failed: %v\n", err)
	}
	a.mu.Unlock()
}

func (a *App) cleanupBoardStagingLocked(tokens []string) error {
	if tokens == nil {
		tokens = make([]string, 0, len(a.boardMedia))
		for token := range a.boardMedia {
			tokens = append(tokens, token)
		}
	}
	var cleanupErrors []error
	for _, token := range tokens {
		staged, exists := a.boardMedia[token]
		if !exists {
			continue
		}
		if err := os.Remove(staged.Path); err != nil && !os.IsNotExist(err) {
			cleanupErrors = append(cleanupErrors, fmt.Errorf("%s: %w", staged.OriginalName, err))
			continue
		}
		delete(a.boardMedia, token)
	}
	if len(a.boardMedia) == 0 && a.stagingDir != "" {
		if err := os.Remove(a.stagingDir); err != nil && !os.IsNotExist(err) {
			cleanupErrors = append(cleanupErrors, fmt.Errorf("임시 폴더: %w", err))
		} else {
			a.stagingDir = ""
			a.boardMedia = nil
		}
	}
	return errors.Join(cleanupErrors...)
}

func (a *App) buildBoardSaveLocked(root string, current boardSnapshot, request []BoardSaveItem) ([]byte, []pendingBoardMedia, []string, error) {
	if len(request) > 500 {
		return nil, nil, nil, errors.New("게시글은 최대 500개까지 저장할 수 있습니다")
	}
	currentByKey := make(map[string]boardRow, len(current.Rows))
	for _, row := range current.Rows {
		currentByKey[row.Key] = row
	}

	type preparedRow struct {
		raw   json.RawMessage
		item  boardDocumentItem
		order int
	}
	prepared := make([]preparedRow, 0, len(request))
	usedExisting := make(map[string]bool)
	usedTokens := make(map[string]bool)
	reservedNames, err := boardMediaNames(filepath.Join(root, "data", "media"))
	if err != nil {
		return nil, nil, nil, err
	}
	pending := make([]pendingBoardMedia, 0)
	tokens := make([]string, 0)

	for index, requested := range request {
		hasExisting := requested.EditorKey != ""
		if hasExisting {
			if requested.NewPost != nil {
				return nil, nil, nil, fmt.Errorf("게시글 %d의 저장 형식이 올바르지 않습니다", index+1)
			}
			row, exists := currentByKey[requested.EditorKey]
			if !exists {
				return nil, nil, nil, fmt.Errorf("게시글 %d가 현재 board.json에 없습니다. 다시 불러와 주세요", index+1)
			}
			if usedExisting[requested.EditorKey] {
				return nil, nil, nil, fmt.Errorf("게시글 %d가 중복되었습니다", index+1)
			}
			usedExisting[requested.EditorKey] = true
			if requested.Post == nil {
				prepared = append(prepared, preparedRow{raw: row.Raw, item: row.Item, order: index})
				continue
			}
			document, raw, rowPending, rowTokens, err := a.prepareBoardPostLocked(
				root,
				*requested.Post,
				&row,
				fmt.Sprintf("게시글 %d", index+1),
				reservedNames,
				usedTokens,
			)
			if err != nil {
				return nil, nil, nil, err
			}
			pending = append(pending, rowPending...)
			tokens = append(tokens, rowTokens...)
			prepared = append(prepared, preparedRow{raw: raw, item: document, order: index})
			continue
		}

		if (requested.Post == nil) == (requested.NewPost == nil) {
			return nil, nil, nil, fmt.Errorf("게시글 %d의 저장 형식이 올바르지 않습니다", index+1)
		}
		post := requested.Post
		if post == nil {
			post = requested.NewPost
		}
		document, raw, rowPending, rowTokens, err := a.prepareBoardPostLocked(
			root,
			*post,
			nil,
			fmt.Sprintf("새 게시글 %d", index+1),
			reservedNames,
			usedTokens,
		)
		if err != nil {
			return nil, nil, nil, err
		}
		pending = append(pending, rowPending...)
		tokens = append(tokens, rowTokens...)
		prepared = append(prepared, preparedRow{raw: raw, item: document, order: index})
	}

	// The editor mirrors the website: newest start date first, then newest end
	// date, with stable order for exact ties. There is no manual reordering.
	sort.SliceStable(prepared, func(left, right int) bool {
		if prepared[left].item.StartDate != prepared[right].item.StartDate {
			return prepared[left].item.StartDate > prepared[right].item.StartDate
		}
		if prepared[left].item.EndDate != prepared[right].item.EndDate {
			return prepared[left].item.EndDate > prepared[right].item.EndDate
		}
		return prepared[left].order < prepared[right].order
	})
	rows := make([]json.RawMessage, 0, len(prepared))
	for _, row := range prepared {
		rows = append(rows, row.raw)
	}
	encoded, err := dataversion.MarshalArray(current.Raw, rows, "board.json")
	if err != nil {
		return nil, nil, nil, fmt.Errorf("board.json 인코딩 실패: %w", err)
	}
	return encoded, pending, tokens, nil
}

func (a *App) prepareBoardPostLocked(
	root string,
	input BoardPostInput,
	current *boardRow,
	contextLabel string,
	reservedNames map[string]bool,
	usedTokens map[string]bool,
) (boardDocumentItem, json.RawMessage, []pendingBoardMedia, []string, error) {
	// The frontend submits a full post snapshot for every existing row. Reuse an
	// identical row verbatim so an unrelated edit cannot rewrite manually added
	// fields or turn an omitted optional field into an explicit empty value.
	if input.Media == nil {
		return boardDocumentItem{}, nil, nil, nil, fmt.Errorf("%s의 media는 JSON 배열이어야 합니다", contextLabel)
	}
	if current != nil && boardPostMatchesCurrent(input, current.Item) {
		return current.Item, cloneRawMessage(current.Raw), nil, nil, nil
	}

	post, err := normaliseBoardPost(input, contextLabel)
	if err != nil {
		return boardDocumentItem{}, nil, nil, nil, err
	}
	if current != nil {
		preserveUnchangedBoardText(&post, input, current.Item)
	}

	currentMedia := make(map[string]boardDocumentMedia)
	if current != nil {
		for _, media := range current.Item.Media {
			currentMedia[media.editorKey] = media
		}
	}
	usedExistingMedia := make(map[string]bool)
	documentMedia := make([]boardDocumentMedia, 0, len(post.Media))
	mediaRaw := make([]json.RawMessage, 0, len(post.Media))
	pending := make([]pendingBoardMedia, 0)
	tokens := make([]string, 0)
	for mediaIndex, mediaInput := range post.Media {
		if mediaInput.EditorKey != "" {
			existing, exists := currentMedia[mediaInput.EditorKey]
			if !exists {
				return boardDocumentItem{}, nil, nil, nil, fmt.Errorf("%s 사진 %d가 현재 board.json에 없습니다. 다시 불러와 주세요", contextLabel, mediaIndex+1)
			}
			if usedExistingMedia[mediaInput.EditorKey] {
				return boardDocumentItem{}, nil, nil, nil, fmt.Errorf("%s 사진 %d가 중복되었습니다", contextLabel, mediaIndex+1)
			}
			usedExistingMedia[mediaInput.EditorKey] = true
			captionEN := mediaInput.CaptionEN
			captionKO := mediaInput.CaptionKO
			if input.Media[mediaIndex].CaptionEN == existing.CaptionEN {
				captionEN = existing.CaptionEN
			}
			if input.Media[mediaIndex].CaptionKO == existing.CaptionKO {
				captionKO = existing.CaptionKO
			}
			updated := existing
			updated.CaptionEN = captionEN
			updated.CaptionKO = captionKO
			raw, err := marshalEditedBoardMedia(existing, captionEN, captionKO)
			if err != nil {
				return boardDocumentItem{}, nil, nil, nil, fmt.Errorf("%s 사진 %d 인코딩 실패: %w", contextLabel, mediaIndex+1, err)
			}
			updated.raw = raw
			documentMedia = append(documentMedia, updated)
			mediaRaw = append(mediaRaw, raw)
			continue
		}

		staged, exists := a.boardMedia[mediaInput.StageToken]
		if !exists {
			return boardDocumentItem{}, nil, nil, nil, fmt.Errorf("%s 사진 %d 임시 파일이 없습니다. 다시 드롭해 주세요", contextLabel, mediaIndex+1)
		}
		if usedTokens[mediaInput.StageToken] {
			return boardDocumentItem{}, nil, nil, nil, fmt.Errorf("사진 '%s'이 중복으로 연결되었습니다", staged.OriginalName)
		}
		usedTokens[mediaInput.StageToken] = true
		filename := nextBoardMediaName(post.StartDate, post.TitleEN, mediaIndex, staged.Extension, reservedNames)
		reservedNames[strings.ToLower(filename)] = true
		pending = append(pending, pendingBoardMedia{
			Token:           mediaInput.StageToken,
			StagedPath:      staged.Path,
			DestinationPath: filepath.Join(root, "data", "media", filename),
		})
		tokens = append(tokens, mediaInput.StageToken)
		media := boardDocumentMedia{
			Src:          filename,
			CaptionEN:    mediaInput.CaptionEN,
			CaptionKO:    mediaInput.CaptionKO,
			originalName: filename,
			size:         staged.Size,
		}
		raw, err := json.Marshal(media)
		if err != nil {
			return boardDocumentItem{}, nil, nil, nil, fmt.Errorf("%s 사진 %d 인코딩 실패: %w", contextLabel, mediaIndex+1, err)
		}
		media.raw = raw
		documentMedia = append(documentMedia, media)
		mediaRaw = append(mediaRaw, raw)
	}

	document := boardDocumentItem{
		StartDate: post.StartDate,
		EndDate:   post.EndDate,
		TitleEN:   post.TitleEN,
		TitleKO:   post.TitleKO,
		ContentEN: post.ContentEN,
		ContentKO: post.ContentKO,
		Media:     documentMedia,
	}
	rowFields := make(map[string]json.RawMessage)
	if current != nil {
		rowFields, err = decodeObject(current.Raw)
		if err != nil {
			return boardDocumentItem{}, nil, nil, nil, err
		}
	}
	merged, err := mergeObjectFields(rowFields, map[string]any{
		"start_date": document.StartDate,
		"end_date":   document.EndDate,
		"title_en":   document.TitleEN,
		"title_ko":   document.TitleKO,
		"content_en": document.ContentEN,
		"content_ko": document.ContentKO,
		"media":      mediaRaw,
	})
	if err != nil {
		return boardDocumentItem{}, nil, nil, nil, fmt.Errorf("%s 인코딩 실패: %w", contextLabel, err)
	}
	raw, err := json.Marshal(merged)
	if err != nil {
		return boardDocumentItem{}, nil, nil, nil, fmt.Errorf("%s 인코딩 실패: %w", contextLabel, err)
	}
	return document, raw, pending, tokens, nil
}

func boardPostMatchesCurrent(input BoardPostInput, current boardDocumentItem) bool {
	if input.StartDate != current.StartDate ||
		input.EndDate != current.EndDate ||
		input.TitleEN != current.TitleEN ||
		input.TitleKO != current.TitleKO ||
		input.ContentEN != current.ContentEN ||
		input.ContentKO != current.ContentKO ||
		len(input.Media) != len(current.Media) {
		return false
	}
	for index, requested := range input.Media {
		existing := current.Media[index]
		if requested.EditorKey != existing.editorKey ||
			requested.StageToken != "" ||
			requested.CaptionEN != existing.CaptionEN ||
			requested.CaptionKO != existing.CaptionKO {
			return false
		}
	}
	return true
}

func marshalEditedBoardMedia(current boardDocumentMedia, captionEN, captionKO string) (json.RawMessage, error) {
	known := make(map[string]any, 2)
	if captionEN != current.CaptionEN {
		known["caption_en"] = captionEN
	}
	if captionKO != current.CaptionKO {
		known["caption_ko"] = captionKO
	}
	if len(known) == 0 {
		return cloneRawMessage(current.raw), nil
	}
	fields, err := decodeObject(current.raw)
	if err != nil {
		return nil, err
	}
	merged, err := mergeObjectFields(fields, known)
	if err != nil {
		return nil, err
	}
	return json.Marshal(merged)
}

func preserveUnchangedBoardText(post *BoardPostInput, input BoardPostInput, current boardDocumentItem) {
	if input.StartDate == current.StartDate {
		post.StartDate = current.StartDate
	}
	if input.EndDate == current.EndDate {
		post.EndDate = current.EndDate
	}
	if input.TitleEN == current.TitleEN {
		post.TitleEN = current.TitleEN
	}
	if input.TitleKO == current.TitleKO {
		post.TitleKO = current.TitleKO
	}
	if input.ContentEN == current.ContentEN {
		post.ContentEN = current.ContentEN
	}
	if input.ContentKO == current.ContentKO {
		post.ContentKO = current.ContentKO
	}
}

func normaliseBoardPost(input BoardPostInput, contextLabel string) (BoardPostInput, error) {
	if input.Media == nil {
		return BoardPostInput{}, fmt.Errorf("%s의 media는 JSON 배열이어야 합니다", contextLabel)
	}
	post := BoardPostInput{
		StartDate: strings.TrimSpace(input.StartDate),
		EndDate:   strings.TrimSpace(input.EndDate),
		TitleEN:   strings.TrimSpace(input.TitleEN),
		TitleKO:   strings.TrimSpace(input.TitleKO),
		ContentEN: normaliseBoardContent(input.ContentEN),
		ContentKO: normaliseBoardContent(input.ContentKO),
		Media:     make([]BoardMediaInput, 0, len(input.Media)),
	}
	if err := validateBoardDates(post.StartDate, post.EndDate, contextLabel, false); err != nil {
		return BoardPostInput{}, err
	}
	if err := validateBoardText(post.TitleEN, contextLabel+" 영문 제목", 300, false); err != nil {
		return BoardPostInput{}, err
	}
	if err := validateBoardText(post.TitleKO, contextLabel+" 국문 제목", 300, false); err != nil {
		return BoardPostInput{}, err
	}
	if post.TitleEN == "" && post.TitleKO == "" {
		return BoardPostInput{}, fmt.Errorf("%s의 영문 또는 국문 제목을 입력해 주세요", contextLabel)
	}
	if err := validateBoardText(post.ContentEN, contextLabel+" 영문 본문", 20000, false); err != nil {
		return BoardPostInput{}, err
	}
	if err := validateBoardText(post.ContentKO, contextLabel+" 국문 본문", 20000, false); err != nil {
		return BoardPostInput{}, err
	}
	if len(input.Media) > maxBoardImagesTotal {
		return BoardPostInput{}, fmt.Errorf("%s에는 사진을 최대 %d개까지 넣을 수 있습니다", contextLabel, maxBoardImagesTotal)
	}
	for index, media := range input.Media {
		captionEN := strings.TrimSpace(media.CaptionEN)
		captionKO := strings.TrimSpace(media.CaptionKO)
		if (media.EditorKey == "") == (media.StageToken == "") {
			return BoardPostInput{}, fmt.Errorf("%s 사진 %d의 저장 형식이 올바르지 않습니다", contextLabel, index+1)
		}
		if err := validateBoardText(captionEN, fmt.Sprintf("%s 사진 %d 영문 설명", contextLabel, index+1), 500, false); err != nil {
			return BoardPostInput{}, err
		}
		if err := validateBoardText(captionKO, fmt.Sprintf("%s 사진 %d 국문 설명", contextLabel, index+1), 500, false); err != nil {
			return BoardPostInput{}, err
		}
		post.Media = append(post.Media, BoardMediaInput{
			EditorKey:  media.EditorKey,
			StageToken: media.StageToken,
			CaptionEN:  captionEN,
			CaptionKO:  captionKO,
		})
	}
	return post, nil
}

func normaliseNewBoardPost(input NewBoardPost, contextLabel string) (NewBoardPost, error) {
	return normaliseBoardPost(input, contextLabel)
}

func normaliseBoardContent(value string) string {
	value = strings.ReplaceAll(value, "\r\n", "\n")
	return strings.TrimSpace(strings.ReplaceAll(value, "\r", "\n"))
}

func validateBoardText(value, contextLabel string, maximum int, required bool) error {
	if required && value == "" {
		return fmt.Errorf("%s을(를) 입력해 주세요", contextLabel)
	}
	if utf8.RuneCountInString(value) > maximum {
		return fmt.Errorf("%s은(는) %d자 이하여야 합니다", contextLabel, maximum)
	}
	for _, character := range value {
		if unicode.IsControl(character) && character != '\n' && character != '\t' {
			return fmt.Errorf("%s에는 제어 문자를 사용할 수 없습니다", contextLabel)
		}
	}
	return nil
}

func boardMediaNames(mediaDirectory string) (map[string]bool, error) {
	entries, err := os.ReadDir(mediaDirectory)
	if err != nil {
		return nil, fmt.Errorf("data/media 읽기 실패: %w", err)
	}
	used := make(map[string]bool, len(entries))
	for _, entry := range entries {
		used[strings.ToLower(entry.Name())] = true
	}
	return used, nil
}

func nextBoardMediaName(startDate, titleEN string, mediaIndex int, extension string, used map[string]bool) string {
	datePart := strings.ReplaceAll(startDate, "-", "")
	titlePart := boardTitleSlug(titleEN)
	base := fmt.Sprintf("board-%s-%s-%02d", datePart, titlePart, mediaIndex+1)
	for suffix := 1; ; suffix++ {
		candidate := base + extension
		if suffix > 1 {
			candidate = fmt.Sprintf("%s-%d%s", base, suffix, extension)
		}
		if !used[strings.ToLower(candidate)] {
			return candidate
		}
	}
}

func boardTitleSlug(value string) string {
	var builder strings.Builder
	lastHyphen := false
	for _, character := range strings.ToLower(value) {
		switch {
		case character >= 'a' && character <= 'z', character >= '0' && character <= '9':
			if builder.Len() < 48 {
				builder.WriteRune(character)
			}
			lastHyphen = false
		case builder.Len() > 0 && !lastHyphen:
			builder.WriteByte('-')
			lastHyphen = true
		}
	}
	result := strings.Trim(builder.String(), "-")
	if result == "" {
		return "post"
	}
	return result
}

func publishBoardMedia(pending []pendingBoardMedia) ([]string, error) {
	return publishBoardMediaWithProgress(pending, nil)
}

func publishBoardMediaWithProgress(pending []pendingBoardMedia, progress func(int, int64, int64, bool)) ([]string, error) {
	created := make([]string, 0, len(pending))
	for index, media := range pending {
		source, err := os.Open(media.StagedPath)
		if err != nil {
			return created, err
		}
		info, err := source.Stat()
		if err != nil {
			_ = source.Close()
			return created, err
		}
		destination, err := os.OpenFile(media.DestinationPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if err != nil {
			_ = source.Close()
			return created, err
		}
		if progress != nil {
			progress(index, 0, info.Size(), false)
		}
		counter := &progressWriter{writer: destination, total: info.Size(), notify: func(copied, total int64) {
			if progress != nil {
				progress(index, copied, total, false)
			}
		}}
		_, copyErr := io.Copy(counter, source)
		if copyErr == nil {
			copyErr = destination.Sync()
		}
		closeDestinationErr := destination.Close()
		closeSourceErr := source.Close()
		if copyErr == nil {
			copyErr = closeDestinationErr
		}
		if copyErr == nil {
			copyErr = closeSourceErr
		}
		if copyErr != nil {
			_ = os.Remove(media.DestinationPath)
			return created, copyErr
		}
		created = append(created, media.DestinationPath)
		if progress != nil {
			progress(index, info.Size(), info.Size(), true)
		}
	}
	return created, nil
}

func removeFiles(paths []string) {
	for _, path := range paths {
		_ = os.Remove(path)
	}
}

func requireUnchangedRevision(path, expected, filename string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("%s 재확인 실패: %w", filename, err)
	}
	if revisionOf(raw) != expected {
		return fmt.Errorf("%s이 편집기 밖에서 변경되었습니다. 다시 불러온 뒤 수정해 주세요", filename)
	}
	return nil
}

func rollbackFileIfCurrent(path string, expectedCurrent, previous []byte) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("복구 전 파일 확인 실패: %w", err)
	}
	if revisionOf(raw) != revisionOf(expectedCurrent) {
		return errors.New("저장 직후 외부 변경이 감지되어 해당 변경을 보호하기 위해 자동 복구하지 않았습니다")
	}
	return writeFileAtomically(path, previous)
}
