package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSaveProgressCountsActualDataAndMediaWrites(t *testing.T) {
	root := newBoardRepoFixture(t)
	app := &App{repoRoot: root}
	t.Cleanup(func() { app.shutdown(context.Background()) })
	loaded, err := app.LoadEditorData()
	if err != nil {
		t.Fatal(err)
	}
	input := filepath.Join(t.TempDir(), "photo.png")
	writeFixtureFile(t, input, fixturePNGBytes())
	staged, err := app.StageBoardMedia([]string{input})
	if err != nil || len(staged.Items) != 1 {
		t.Fatalf("stage: %#v, %v", staged, err)
	}
	rows := make([]BoardSaveItem, 0, len(loaded.Board)+1)
	for _, item := range loaded.Board {
		rows = append(rows, BoardSaveItem{EditorKey: item.EditorKey})
	}
	rows = append(rows, BoardSaveItem{NewPost: &NewBoardPost{
		StartDate: "2026-10-02", TitleEN: "Progress test", TitleKO: "진행률 확인",
		Media: []NewBoardMedia{{StageToken: staged.Items[0].StageToken}},
	}})
	request := boardOnlySaveRequest(loaded, rows)
	request.OperationID = "save-progress"
	var progress []OperationProgress
	app.progressSink = func(p OperationProgress) { progress = append(progress, p) }
	if _, err := app.SaveEditorData(request); err != nil {
		t.Fatal(err)
	}
	last := -1
	seenCopy, seenWrite := false, false
	for _, p := range progress {
		if p.OperationID != request.OperationID {
			t.Fatal(p)
		}
		if p.Percent >= 0 && p.Percent < last {
			t.Fatalf("progress reversed: %#v", progress)
		}
		last = p.Percent
		seenCopy = seenCopy || strings.Contains(p.Message, "첨부 파일 1/1개")
		seenWrite = seenWrite || strings.Contains(p.Message, "데이터 1/1개")
	}
	if !seenCopy || !seenWrite || last != 100 || progress[len(progress)-1].Message != "저장 완료" {
		t.Fatalf("incorrect save progress: %#v", progress)
	}
	// A stale revision must stop before writes and never report success.
	progress = nil
	if _, err := app.SaveEditorData(request); err == nil {
		t.Fatal("stale save succeeded")
	}
	for _, p := range progress {
		if p.Percent == 100 {
			t.Fatal("failed save reported completion")
		}
	}
}

func TestPhotoProgressReportsCopiedBytesAfterSuccessfulWrites(t *testing.T) {
	root := t.TempDir()
	input, output := filepath.Join(root, "in.jpg"), filepath.Join(root, "out.jpg")
	contents := make([]byte, 2<<20)
	if err := os.WriteFile(input, contents, 0600); err != nil {
		t.Fatal(err)
	}
	previous := int64(0)
	complete := false
	_, err := publishBoardMediaWithProgress([]pendingBoardMedia{{StagedPath: input, DestinationPath: output}}, func(index int, copied, total int64, done bool) {
		if index != 0 || total != int64(len(contents)) || copied < previous || copied > total {
			t.Fatalf("invalid copy progress: %d %d %d", index, copied, total)
		}
		previous = copied
		complete = done
	})
	if err != nil {
		t.Fatal(err)
	}
	if !complete || previous != int64(len(contents)) {
		t.Fatal("copy did not report completion")
	}
}
