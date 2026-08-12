package main

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Gonie-Gonie/Gonie-Gonie.github.io/editor/internal/citations"
)

type CitationUpdateRequest struct {
	OperationID string `json:"operation_id"`
}
type CitationUpdateResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Output  string `json:"output"`
}

// UpdateCitations refreshes counts directly without reloading editor drafts.
func (a *App) UpdateCitations(request CitationUpdateRequest) (CitationUpdateResponse, error) {
	return a.updateCitations(request, citations.Update)
}

func (a *App) updateCitations(request CitationUpdateRequest, update func(context.Context, citations.Options) (citations.Result, error)) (CitationUpdateResponse, error) {
	if !a.operationBusy.CompareAndSwap(false, true) {
		return CitationUpdateResponse{}, errors.New("다른 작업이 진행 중입니다")
	}
	defer a.operationBusy.Store(false)
	a.mu.Lock()
	root, err := a.rootLocked()
	dirty := a.dirty || len(a.boardMedia) > 0
	appContext := a.ctx
	a.mu.Unlock()
	if err != nil {
		return CitationUpdateResponse{}, err
	}
	if dirty {
		return CitationUpdateResponse{}, errors.New("변경 사항을 먼저 저장해 주세요")
	}
	if appContext == nil {
		appContext = context.Background()
	}
	ctx, cancel := context.WithTimeout(appContext, 10*time.Minute)
	defer cancel()
	a.reportProgress(request.OperationID, "인용수 갱신 중…", -1)
	result, err := update(ctx, citations.Options{Root: root, Log: func(chunk string) {
		a.reportProgressOutput(request.OperationID, "인용수 갱신 중…", -1, chunk)
	}})
	if err != nil {
		detail := "실행 오류: " + err.Error() + "\n"
		if ctx.Err() != nil {
			detail = "인용수 갱신 시간이 초과되거나 작업이 종료되었습니다: " + ctx.Err().Error() + "\n"
		}
		if result.Output != "" && !strings.HasSuffix(result.Output, "\n") {
			detail = "\n" + detail
		}
		result.Output += detail
		a.reportProgressOutput(request.OperationID, "인용수 갱신 실패", -1, detail)
		result.Success = false
	}
	if !result.Success {
		a.reportProgress(request.OperationID, "인용수 갱신 실패", -1)
		return CitationUpdateResponse{Message: "인용수 갱신 실패", Output: result.Output}, nil
	}
	a.reportProgress(request.OperationID, "인용수 갱신 완료", 100)
	return CitationUpdateResponse{Success: true, Message: "인용수 갱신 완료", Output: result.Output}, nil
}
