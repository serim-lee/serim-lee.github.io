package sitecheck

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func write(t *testing.T, root, name, value string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(value), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestHTMLAndCSSReferences(t *testing.T) {
	root := t.TempDir()
	write(t, root, "index.html", "<!doctype html>\n<div id='same'></div>\n<img id='same' src='missing.png'><a href='https://example.org/missing'>external</a><a href='#same'>local</a><img src='assets/a%20b.png?x=1#image'>")
	write(t, root, "assets/a b.png", "image")
	c := &checker{root: root}
	c.checkHTML("index.html")
	joined := strings.Join(c.errors, "\n")
	if len(c.errors) != 2 || !strings.Contains(joined, "index.html:3: duplicate id") || !strings.Contains(joined, "missing.png") {
		t.Fatalf("unexpected errors: %v", c.errors)
	}
	write(t, root, "assets/style.css", "a{background:url('a%20b.png')}\nb{background:url(\"absent.svg\")}\nc{background:url(data:image/png;base64,a)}")
	c.errors = nil
	c.checkCSS("assets/style.css")
	if len(c.errors) != 1 || !strings.Contains(c.errors[0], "assets/style.css:2") || !strings.Contains(c.errors[0], "absent.svg") {
		t.Fatal(c.errors)
	}
}

func TestMediaConfinementAndCaptions(t *testing.T) {
	root := t.TempDir()
	write(t, root, "data/media/image.png", "image")
	write(t, root, "outside.png", "outside")
	for _, src := range []string{"../outside.png", "../../outside.png", filepath.Join(root, "outside.png"), "missing.png", ""} {
		c := &checker{root: root}
		c.media(src, "fixture")
		if len(c.errors) == 0 {
			t.Fatalf("unsafe/missing media accepted: %s", src)
		}
	}
	c := &checker{root: root}
	c.checkMedia(map[string]any{"src": "image.png", "type": "image", "caption_en": "English", "caption_ko": "국문"}, "valid", true, true)
	if len(c.errors) > 0 {
		t.Fatal(c.errors)
	}
	c.checkMedia(map[string]any{"src": "image.png", "type": "invalid", "caption_en": "English", "caption_kr": "legacy"}, "invalid", true, true)
	if len(c.errors) < 3 {
		t.Fatal("invalid kind/caption pair/legacy field accepted", c.errors)
	}
}

func TestIconSVGValidation(t *testing.T) {
	for name, svg := range map[string]string{
		"valid":     `<svg xmlns="http://www.w3.org/2000/svg"><path d="M0 0"/></svg>`,
		"raster":    `<svg xmlns="http://www.w3.org/2000/svg"><image href="a.png"/></svg>`,
		"namespace": `<svg><path/></svg>`,
		"invalid":   `<svg xmlns="http://www.w3.org/2000/svg"><path></svg>`,
		"multiple":  `<svg xmlns="http://www.w3.org/2000/svg"/><svg/>`,
		"trailing":  `<svg xmlns="http://www.w3.org/2000/svg"/>invalid text`,
	} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			write(t, root, "app/assets/js/app.js", `const icon = {icon: "app/assets/icons/technologies/test.svg"};`)
			write(t, root, "app/assets/icons/technologies/test.svg", svg)
			c := &checker{root: root}
			c.checkIcons()
			if (len(c.errors) == 0) != (name == "valid") {
				t.Fatal(c.errors)
			}
		})
	}
}

func TestPresentationDataChecks(t *testing.T) {
	c := &checker{root: t.TempDir()}
	c.software(map[string]any{"stage": "unknown", "links": []any{"javascript:alert(1)"}, "technologies": []any{""}}, "fixture")
	c.dates(map[string]any{"start_date": "2025-02-29", "end_date": "2025-01-01"}, "fixture", true)
	if len(c.errors) < 4 {
		t.Fatal(c.errors)
	}
	valid := &checker{root: t.TempDir()}
	valid.software(map[string]any{"stage": "release", "links": []any{"https://example.org"}, "technologies": []any{"Go"}}, "fixture")
	valid.dates(map[string]any{"start_date": "2024-02-29", "end_date": "2025-01-01"}, "fixture", true)
	if len(valid.errors) > 0 {
		t.Fatal(valid.errors)
	}
}

func TestIndependentContentNotes(t *testing.T) {
	for _, collection := range []string{"projects", "software"} {
		for _, counts := range [][2]int{{1, 3}, {3, 1}} {
			t.Run(fmt.Sprintf("%s_en%d_kr%d", collection, counts[0], counts[1]), func(t *testing.T) {
				root := t.TempDir()
				item := map[string]any{
					"start_date": "2025-01-01", "end_date": "2026-01-01", "theme": "research",
					"media": []any{}, "stage": "release", "links": []any{}, "technologies": []any{"Go"},
				}
				for language, count := range map[string]int{"en": counts[0], "kr": counts[1]} {
					notes := make([]any, count)
					for index := range notes {
						notes[index] = fmt.Sprintf("%s note %d", language, index+1)
					}
					item["notes_"+language] = notes
				}
				file := "data/" + collection + ".json"
				raw, err := json.Marshal([]any{item})
				if err != nil {
					t.Fatal(err)
				}
				write(t, root, file, string(raw))
				c := &checker{root: root}
				c.checkCollection(collection)
				if len(c.errors) > 0 {
					t.Fatal(c.errors)
				}
				after, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(file)))
				if err != nil || string(after) != string(raw) {
					t.Fatal("check changed source data", err)
				}
				longer := "notes_en"
				if counts[1] > counts[0] {
					longer = "notes_kr"
				}
				notes := item[longer].([]any)
				notes[len(notes)-1] = false
				raw, err = json.Marshal([]any{item})
				if err != nil {
					t.Fatal(err)
				}
				write(t, root, file, string(raw))
				c = &checker{root: root}
				c.checkCollection(collection)
				if !strings.Contains(strings.Join(c.errors, "\n"), longer) {
					t.Fatal("invalid unmatched note accepted", c.errors)
				}
			})
		}
	}
}
