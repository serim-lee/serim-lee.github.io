package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var fixturePDF = []byte("%PDF-1.4\n1 0 obj\n<< /Type /Catalog >>\nendobj\n%%EOF\n")

func certificationWithPDF(filename string) map[string]any {
	return map[string]any{"name_en": "Qualification", "name_ko": "자격", "issuer_en": "Issuer", "issuer_ko": "기관", "date": "2020-11-20", "pdf": filename}
}

func profileWithDocument(t *testing.T, loaded EditorDataResponse, filename string) json.RawMessage {
	t.Helper()
	var profile map[string]any
	if err := json.Unmarshal(loaded.Profile, &profile); err != nil {
		t.Fatal(err)
	}
	profile["certifications"] = []any{certificationWithPDF(filename)}
	raw, err := json.Marshal(profile)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestDocumentsSaveAwardsAndCertificationsTogether(t *testing.T) {
	root := newTempRepoFixture(t, fixtureSettingsJSON())
	app := &App{repoRoot: root}
	t.Cleanup(func() { app.shutdown(context.Background()) })
	loaded, err := app.LoadEditorData()
	if err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(t.TempDir(), "자격 증명.pdf")
	writeFixtureFile(t, source, fixturePDF)
	first, err := app.StageDocument([]string{source})
	if err != nil || len(first.Items) != 1 {
		t.Fatalf("stage: %#v %v", first, err)
	}
	second, err := app.StageDocument([]string{source})
	if err != nil || len(second.Items) != 1 {
		t.Fatalf("stage second: %#v %v", second, err)
	}
	if first.Items[0].Filename != "자격 증명.pdf" || second.Items[0].Filename == first.Items[0].Filename {
		t.Fatal("duplicate names not separated")
	}
	if _, err := os.Stat(filepath.Join(root, "data", "document")); !os.IsNotExist(err) {
		t.Fatal("drop published files before save")
	}
	request := saveRequestRevisions(loaded)
	request.SaveProfile = true
	request.Profile = profileWithDocument(t, loaded, first.Items[0].Filename)
	request.SaveAwards = true
	filename := second.Items[0].Filename
	request.Awards = []AwardSaveItem{{Award: &AwardInput{Date: "2020-11-20", TitleEN: "Example award", TitleKO: "수상", OrganizationEN: "Society", OrganizationKO: "학회", PDF: &filename}}}
	request.DocumentStageTokens = []string{first.Items[0].StageToken, second.Items[0].StageToken}
	saved, err := app.SaveEditorData(request)
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range []StagedDocumentItem{first.Items[0], second.Items[0]} {
		got, err := os.ReadFile(filepath.Join(root, "data", "document", item.Filename))
		if err != nil || !bytes.Equal(got, fixturePDF) {
			t.Fatalf("PDF bytes changed: %v", err)
		}
	}
	if len(app.documents) != 0 || app.documentStagingDir != "" {
		t.Fatal("staging not cleaned after save")
	}
	if len(saved.Awards) != 1 || saved.Awards[0].PDF != filename {
		t.Fatal("award PDF did not round trip")
	}
	var profile map[string]any
	if err := json.Unmarshal(saved.Profile, &profile); err != nil {
		t.Fatal(err)
	}
	certification := profile["certifications"].([]any)[0].(map[string]any)
	if certification["pdf"] != first.Items[0].Filename {
		t.Fatal("certification PDF did not round trip")
	}
	delete(certification, "pdf")
	remove := saveRequestRevisions(saved)
	remove.SaveProfile = true
	remove.Profile, _ = json.Marshal(profile)
	remove.SaveAwards = true
	input := awardInputFromSummary(saved.Awards[0])
	empty := ""
	input.PDF = &empty
	remove.Awards = []AwardSaveItem{{EditorKey: saved.Awards[0].EditorKey, Award: input}}
	removed, err := app.SaveEditorData(remove)
	if err != nil {
		t.Fatal(err)
	}
	if removed.Awards[0].PDF != "" || bytes.Contains(removed.Profile, []byte(`"pdf"`)) {
		t.Fatal("PDF references not removed")
	}
	for _, name := range []string{first.Items[0].Filename, filename} {
		if _, err := os.Stat(filepath.Join(root, "data", "document", name)); err != nil {
			t.Fatal("unlink deleted an archived PDF")
		}
	}
}

func TestDocumentFailedSavePreservesDataAndStaging(t *testing.T) {
	root := newTempRepoFixture(t, fixtureSettingsJSON())
	beforeProfile := readFixtureFile(t, filepath.Join(root, "data", "profile.json"))
	app := &App{repoRoot: root}
	t.Cleanup(func() { app.shutdown(context.Background()) })
	loaded, err := app.LoadEditorData()
	if err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(t.TempDir(), "proof.pdf")
	writeFixtureFile(t, source, fixturePDF)
	staged, err := app.StageDocument([]string{source})
	if err != nil || len(staged.Items) != 1 {
		t.Fatal(err)
	}
	item := staged.Items[0]
	request := saveRequestRevisions(loaded)
	request.SaveProfile = true
	request.Profile = profileWithDocument(t, loaded, item.Filename)
	request.DocumentStageTokens = []string{item.StageToken}
	stale := request
	stale.ProfileRevision = "stale"
	if _, err := app.SaveEditorData(stale); err == nil {
		t.Fatal("stale save accepted")
	}
	missing := request
	missing.DocumentStageTokens = nil
	if _, err := app.SaveEditorData(missing); err == nil {
		t.Fatal("missing document accepted")
	}
	unlinked := request
	unlinked.Profile = loaded.Profile
	if _, err := app.SaveEditorData(unlinked); err == nil {
		t.Fatal("unlinked upload accepted")
	}
	destination := filepath.Join(root, "data", "document", item.Filename)
	writeProjectTestFile(t, destination, fixturePDF)
	if _, err := app.SaveEditorData(request); err == nil {
		t.Fatal("save overwrote an external file")
	}
	if !bytes.Equal(readFixtureFile(t, filepath.Join(root, "data", "profile.json")), beforeProfile) {
		t.Fatal("failed save changed profile")
	}
	if len(app.documents) != 1 {
		t.Fatal("failed save lost pending upload")
	}
	if !bytes.Equal(readFixtureFile(t, destination), fixturePDF) {
		t.Fatal("failed save changed existing PDF")
	}
	if err := os.Remove(destination); err != nil {
		t.Fatal(err)
	}
	if _, err := app.SaveEditorData(request); err != nil {
		t.Fatalf("retry: %v", err)
	}
}

func TestDocumentValidationAndDiscard(t *testing.T) {
	for _, filename := range []string{"../proof.pdf", `folder\proof.pdf`, "https://example.com/proof.pdf", "proof.jpg", " CON.pdf", "NUL.pdf"} {
		if validateDocumentFilename(filename) == nil {
			t.Errorf("unsafe filename accepted: %s", filename)
		}
	}
	root := newTempRepoFixture(t, fixtureSettingsJSON())
	app := &App{repoRoot: root}
	t.Cleanup(func() { app.shutdown(context.Background()) })
	input := filepath.Join(t.TempDir(), "invalid.pdf")
	for _, contents := range [][]byte{[]byte("not a pdf"), []byte("%PDF-1.4\ntruncated")} {
		writeFixtureFile(t, input, contents)
		response, err := app.StageDocument([]string{input})
		if err != nil || len(response.Items) != 0 || len(response.Rejected) != 1 {
			t.Fatal("invalid PDF accepted")
		}
	}
	writeFixtureFile(t, input, fixturePDF)
	response, err := app.StageDocument([]string{input})
	if err != nil || len(response.Items) != 1 {
		t.Fatal(err)
	}
	stagedPath := app.documents[response.Items[0].StageToken].Path
	if err := app.DiscardDocuments([]string{response.Items[0].StageToken}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(stagedPath); !os.IsNotExist(err) {
		t.Fatal("discard kept staged PDF")
	}
	if _, err := os.Stat(input); err != nil {
		t.Fatal("discard removed source PDF")
	}
}

func TestAwardOptionalPDFPreservesOldClientsAndRemovesExplicitly(t *testing.T) {
	raw := bytes.Replace(fixtureAwardsJSON(), []byte(`"id": "best-paper",`), []byte(`"id": "best-paper", "pdf": "proof.pdf",`), 1)
	current, err := readAwardsBytes(raw)
	if err != nil {
		t.Fatal(err)
	}
	input := awardInputFromSummary(awardItemSummaries(current)[0])
	input.TitleKO = "수정된 제목"
	items := []AwardSaveItem{{EditorKey: current.Rows[0].Key, Award: input}, {EditorKey: current.Rows[1].Key}}
	updated, err := (&App{}).buildAwardsSaveLocked(current, items)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(updated), `"pdf": "proof.pdf"`) {
		t.Fatal("older client dropped optional PDF")
	}
	stored, err := readAwardsBytes(updated)
	if err != nil {
		t.Fatal(err)
	}
	empty := ""
	input.PDF = &empty
	items[0].EditorKey = stored.Rows[0].Key
	items[1].EditorKey = stored.Rows[1].Key
	removed, err := (&App{}).buildAwardsSaveLocked(stored, items)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(removed, []byte(`"pdf"`)) {
		t.Fatal("explicit removal kept PDF")
	}
}
