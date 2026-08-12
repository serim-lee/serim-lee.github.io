package main

import (
	"strings"
	"testing"
)

func TestEditorMediaPreviewAssetsAreWired(t *testing.T) {
	for filename, snippets := range map[string][]string{
		"index.html": {
			`<dialog id="media-preview-dialog"`,
			`id="media-preview-close"`,
			`aria-label="확대 보기 닫기"`,
			`id="media-preview-content"`,
			`id="media-preview-status"`,
		},
		"app.js": {
			`function makeMediaPreviewButton(`,
			`query.set("stage_token", media.stage_token)`,
			`query.set("src", media.src)`,
			"return `/__editor_media?${query}`",
			`if (!dialog.open) dialog.showModal()`,
			`addEventListener("cancel",`,
			`element.pause()`,
			`element.removeAttribute("src")`,
			`trigger.focus({ preventScroll: true })`,
		},
		"styles.css": {
			`.media-preview-dialog::backdrop`,
			`.media-preview-dialog [hidden]`,
			`.media-preview-trigger`,
			`.media-preview-full`,
		},
	} {
		content, err := assets.ReadFile("frontend/dist/" + filename)
		if err != nil {
			t.Fatal(err)
		}
		for _, snippet := range snippets {
			if !strings.Contains(string(content), snippet) {
				t.Errorf("%s is missing media preview wiring %q", filename, snippet)
			}
		}
	}
}
