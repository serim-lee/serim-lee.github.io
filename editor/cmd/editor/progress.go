package main

import (
	"io"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const operationProgressEvent = "editor:operation-progress"

type OperationProgress struct {
	OperationID string `json:"operation_id"`
	Message     string `json:"message"`
	Percent     int    `json:"percent"` // -1 means the amount of work is not yet known.
	Output      string `json:"output,omitempty"`
}

func (a *App) reportProgress(id, message string, percent int) {
	a.reportProgressOutput(id, message, percent, "")
}

func (a *App) reportProgressOutput(id, message string, percent int, output string) {
	if id == "" {
		return
	}
	event := OperationProgress{OperationID: id, Message: message, Percent: percent, Output: output}
	if a.progressSink != nil {
		a.progressSink(event)
	}
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, operationProgressEvent, event)
	}
}

type progressWriter struct {
	writer    io.Writer
	completed int64
	total     int64
	last      time.Time
	notify    func(int64, int64)
}

func (w *progressWriter) Write(data []byte) (int, error) {
	n, err := w.writer.Write(data)
	w.completed += int64(n)
	if w.notify != nil && (time.Since(w.last) >= 50*time.Millisecond || w.completed == w.total) {
		w.last = time.Now()
		w.notify(w.completed, w.total)
	}
	return n, err
}
