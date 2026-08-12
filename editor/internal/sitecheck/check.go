// Package sitecheck validates local website references and presentation data.
package sitecheck

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"github.com/Gonie-Gonie/Gonie-Gonie.github.io/editor/internal/dataversion"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/Gonie-Gonie/Gonie-Gonie.github.io/editor/internal/sitebuild"
	"golang.org/x/net/html"
)

type checker struct {
	root   string
	errors []string
}

// Check validates the source data, HTML, CSS, media, icons, and CV wiring.
func Check(root string) (string, error) {
	if err := sitebuild.Validate(root); err != nil {
		return "", err
	}
	c := &checker{root: root}
	for _, name := range []string{"index.html", "cv.html", "fork-guide.html"} {
		c.checkHTML(name)
	}
	c.checkCSS("app/assets/css/styles.css")
	c.checkCSS("app/assets/css/fork-guide.css")
	for _, name := range []string{"board", "projects", "software"} {
		c.checkCollection(name)
	}
	c.checkProfile()
	c.checkIcons()
	c.checkCV()
	if len(c.errors) > 0 {
		return "", fmt.Errorf("Static site check failed:\n  - %s", strings.Join(c.errors, "\n  - "))
	}
	return "Static site references and HTML IDs are valid.\n", nil
}

func (c *checker) fail(format string, args ...any) {
	c.errors = append(c.errors, fmt.Sprintf(format, args...))
}
func (c *checker) read(name string) []byte {
	b, err := os.ReadFile(filepath.Join(c.root, filepath.FromSlash(name)))
	if err != nil {
		c.fail("%s: %v", name, err)
	}
	return b
}
func (c *checker) data(name string) any {
	var v any
	b := c.read("data/" + name + ".json")
	if b != nil {
		var err error
		v, err = dataversion.Unwrap(b, name+".json")
		if err != nil {
			c.fail("data/%s.json: %v", name, err)
		}
	}
	return v
}
func object(v any) map[string]any { m, _ := v.(map[string]any); return m }
func array(v any) []any           { a, _ := v.([]any); return a }
func text(v any) string           { s, _ := v.(string); return strings.TrimSpace(s) }
func exists(path string) bool     { _, err := os.Stat(path); return err == nil }

func (c *checker) target(source, value string) (string, bool) {
	value = strings.TrimSpace(value)
	if value == "" || strings.HasPrefix(value, "#") {
		return "", false
	}
	u, err := url.Parse(value)
	if err != nil {
		c.fail("%s: malformed reference '%s': %v", source, value, err)
		return "", false
	}
	if u.Scheme != "" || u.Host != "" || u.Path == "" {
		return "", false
	}
	if strings.HasPrefix(u.Path, "/") {
		return filepath.Join(c.root, filepath.FromSlash(strings.TrimLeft(u.Path, "/"))), true
	}
	return filepath.Join(c.root, filepath.Dir(filepath.FromSlash(source)), filepath.FromSlash(u.Path)), true
}

func (c *checker) checkHTML(name string) {
	b := c.read(name)
	if b == nil {
		return
	}
	z := html.NewTokenizer(bytes.NewReader(b))
	seen := map[string]int{}
	line := 1
	for {
		typeOf := z.Next()
		if typeOf == html.ErrorToken {
			if err := z.Err(); err != io.EOF {
				c.fail("%s: %v", name, err)
			}
			return
		}
		startLine := line
		line += bytes.Count(z.Raw(), []byte{'\n'})
		if typeOf != html.StartTagToken && typeOf != html.SelfClosingTagToken {
			continue
		}
		token := z.Token()
		for _, attr := range token.Attr {
			if attr.Key == "id" && attr.Val != "" {
				if first, ok := seen[attr.Val]; ok {
					c.fail("%s:%d: duplicate id '%s' (first seen at line %d).", name, startLine, attr.Val, first)
				} else {
					seen[attr.Val] = startLine
				}
			}
			if attr.Key == "src" || attr.Key == "href" || attr.Key == "poster" {
				if target, local := c.target(name, attr.Val); local && !exists(target) {
					c.fail("%s:%d: missing %s target '%s'.", name, startLine, attr.Key, attr.Val)
				}
			}
		}
	}
}

var cssURL = regexp.MustCompile(`(?is)url\(\s*(?:"([^"]*)"|'([^']*)'|([^)]*))\s*\)`)

func (c *checker) checkCSS(name string) {
	b := c.read(name)
	for _, match := range cssURL.FindAllSubmatchIndex(b, -1) {
		value := ""
		for i := 2; i < len(match); i += 2 {
			if match[i] >= 0 {
				value = string(b[match[i]:match[i+1]])
				break
			}
		}
		if target, local := c.target(name, value); local && !exists(target) {
			c.fail("%s:%d: missing CSS asset '%s'.", name, bytes.Count(b[:match[0]], []byte{'\n'})+1, value)
		}
	}
}

func inside(base, target string) bool {
	rel, err := filepath.Rel(base, target)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)
}

func (c *checker) media(value any, context string) {
	src := text(value)
	if src == "" {
		c.fail("%s: src is required.", context)
		return
	}
	base, _ := filepath.Abs(filepath.Join(c.root, "data", "media"))
	path := filepath.FromSlash(src)
	if filepath.IsAbs(path) || filepath.VolumeName(path) != "" {
		c.fail("%s: media must be relative to data/media.", context)
		return
	}
	target := filepath.Join(base, path)
	if resolved, err := filepath.EvalSymlinks(target); err == nil {
		target = resolved
	}
	if !inside(base, target) {
		c.fail("%s: media must be relative to data/media.", context)
		return
	}
	info, err := os.Stat(target)
	if err != nil || !info.Mode().IsRegular() {
		c.fail("%s: missing media '%s'.", context, src)
	}
}

func (c *checker) checkMedia(v any, context string, requiredCaptions bool, pairedCaptions bool) {
	m := object(v)
	if m == nil {
		c.fail("%s: must be an object.", context)
		return
	}
	kind := strings.ToLower(text(m["type"]))
	if value := m["type"]; value != nil {
		if _, ok := value.(string); !ok {
			c.fail("%s: type must be 'image' or 'video'.", context)
		}
	}
	if kind != "" && kind != "image" && kind != "video" {
		c.fail("%s: type must be 'image' or 'video'.", context)
	}
	c.media(m["src"], context)
	if text(m["poster"]) != "" {
		c.media(m["poster"], context+" poster")
	}
	if requiredCaptions {
		for _, lang := range []string{"en", "ko"} {
			if text(m["caption_"+lang]) == "" {
				c.fail("%s: caption_%s is required.", context, lang)
			}
		}
	}
	if pairedCaptions {
		if _, old := m["caption_kr"]; old {
			c.fail("%s: use caption_en and caption_ko.", context)
		}
		if _, old := m["caption-en"]; old {
			c.fail("%s: use caption_en and caption_ko.", context)
		}
		if (text(m["caption_en"]) == "") != (text(m["caption_ko"]) == "") {
			c.fail("%s: caption_en and caption_ko must both be filled or both be blank.", context)
		}
	}
}

func (c *checker) checkCollection(name string) {
	for i, value := range array(c.data(name)) {
		m := object(value)
		if m == nil {
			c.fail("data/%s.json item %d: must be an object.", name, i+1)
			continue
		}
		context := fmt.Sprintf("data/%s.json item %d", name, i+1)
		if _, old := m["photos"]; old {
			c.fail("%s: use media instead of the legacy photos field.", context)
		}
		media, ok := m["media"].([]any)
		if !ok {
			c.fail("%s: media must be a list.", context)
		}
		for j, item := range media {
			c.checkMedia(item, fmt.Sprintf("%s media %d", context, j+1), name == "software", name == "projects")
		}
		if name == "projects" || name == "software" {
			if _, old := m["order"]; old {
				c.fail("%s: remove order; array position controls display order.", context)
			}
			for _, old := range []string{"summary", "summary_en", "summary_ko"} {
				if _, found := m[old]; found {
					c.fail("%s: use notes_en and notes_kr instead of summary.", context)
				}
			}
			for _, field := range []string{"notes_en", "notes_kr"} {
				notes, ok := m[field].([]any)
				valid := ok && len(notes) > 0
				for _, note := range notes {
					if text(note) == "" {
						valid = false
					}
				}
				if !valid {
					c.fail("%s: %s must be a non-empty array of non-empty strings.", context, field)
				}
			}
		}
		if name == "projects" {
			if _, old := m["id"]; old {
				c.fail("%s: remove id; projects use their array position.", context)
			}
			if _, old := m["period"]; old {
				c.fail("%s: use start_date and end_date instead of period.", context)
			}
			c.dates(m, context, true)
			if text(m["theme"]) == "" {
				c.fail("%s: theme is required.", context)
			}
		}
		if name == "board" {
			if _, old := m["date"]; old {
				c.fail("%s: use start_date and end_date instead of date.", context)
			}
			c.dates(m, context, false)
		}
		if name == "software" {
			c.software(m, context)
		}
	}
}

func (c *checker) dates(m map[string]any, context string, endRequired bool) {
	start, end := text(m["start_date"]), text(m["end_date"])
	valid := func(value string) bool {
		parsed, err := time.Parse("2006-01-02", value)
		return err == nil && parsed.Format("2006-01-02") == value && parsed.Year() > 0
	}
	if !valid(start) {
		c.fail("%s: start_date must be a valid YYYY-MM-DD date.", context)
	}
	if (end != "" || endRequired) && !valid(end) {
		c.fail("%s: end_date must be a valid YYYY-MM-DD date.", context)
	}
	if valid(start) && valid(end) && start > end {
		c.fail("%s: start_date must not be later than end_date.", context)
	}
}

func (c *checker) software(m map[string]any, context string) {
	stage := text(m["stage"])
	if stage != "release" && stage != "preview" && stage != "development" {
		c.fail("%s: stage must be release, preview, or development.", context)
	}
	for _, old := range []string{"repo_url", "website_url"} {
		if _, found := m[old]; found {
			c.fail("%s: use the links array instead of repo_url or website_url.", context)
		}
	}
	links, ok := m["links"].([]any)
	if !ok {
		c.fail("%s: links must be a list.", context)
	}
	for i, value := range links {
		label := fmt.Sprintf("%s link %d", context, i+1)
		address, isString := value.(string)
		if !isString {
			m := object(value)
			if m == nil {
				c.fail("%s: must be a URL string or an object with a url field.", label)
				continue
			}
			address = text(m["url"])
			for _, field := range []string{"label", "label_en", "label_ko"} {
				if v, found := m[field]; found {
					if _, ok := v.(string); !ok {
						c.fail("%s: %s must be a string.", label, field)
					}
				}
			}
		}
		u, err := url.Parse(strings.TrimSpace(address))
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			c.fail("%s: must contain a valid http(s) URL.", label)
		}
	}
	tech, ok := m["technologies"].([]any)
	if !ok || len(tech) == 0 {
		c.fail("%s: technologies must be a non-empty list.", context)
	}
	for _, name := range tech {
		if text(name) == "" {
			c.fail("%s: every technology must be a non-empty string.", context)
		}
	}
}

func (c *checker) checkProfile() {
	card := object(object(c.data("profile"))["profile_card"])
	if _, old := card["image"]; old {
		c.fail("data/profile.json profile_card: use media instead of the legacy image field.")
	}
	media := object(card["media"])
	if media == nil {
		c.fail("data/profile.json profile_card.media must be an object.")
		return
	}
	kind := strings.ToLower(text(media["type"]))
	if value := media["type"]; value != nil {
		if _, ok := value.(string); !ok {
			c.fail("data/profile.json profile_card.media: type must be 'image' or 'video'.")
		}
	}
	if value := media["src"]; value != nil {
		if _, ok := value.(string); !ok {
			c.fail("data/profile.json profile_card.media: src must be a string.")
		}
	}
	if kind != "" && kind != "image" && kind != "video" {
		c.fail("data/profile.json profile_card.media: type must be 'image' or 'video'.")
	}
	if text(media["src"]) != "" {
		c.media(media["src"], "data/profile.json profile_card.media")
	}
	if text(media["poster"]) != "" {
		c.media(media["poster"], "data/profile.json profile_card.media poster")
	}
}

var iconPattern = regexp.MustCompile(`icon:\s*"(app/assets/icons/[^"]+)"`)

func (c *checker) checkIcons() {
	for _, match := range iconPattern.FindAllSubmatch(c.read("app/assets/js/app.js"), -1) {
		name := string(match[1])
		path := filepath.Join(c.root, filepath.FromSlash(name))
		if info, err := os.Stat(path); err != nil || !info.Mode().IsRegular() {
			c.fail("app/assets/js/app.js: missing icon asset '%s'.", name)
			continue
		}
		if strings.Contains(name, "/technologies/") && !strings.EqualFold(filepath.Ext(name), ".svg") {
			c.fail("app/assets/js/app.js: technology icon must use SVG: '%s'.", name)
		}
		if !strings.EqualFold(filepath.Ext(name), ".svg") {
			continue
		}
		file, err := os.Open(path)
		if err != nil {
			c.fail("%s: %v", name, err)
			continue
		}
		decoder := xml.NewDecoder(file)
		rootSeen := false
		depth := 0
		for {
			token, err := decoder.Token()
			if err != nil {
				if err != io.EOF {
					c.fail("%s: invalid SVG: %v.", name, err)
				}
				break
			}
			if _, ok := token.(xml.EndElement); ok {
				depth--
				continue
			}
			if chars, ok := token.(xml.CharData); ok && depth == 0 && strings.TrimSpace(string(chars)) != "" {
				c.fail("%s: invalid SVG: text outside the root element.", name)
			}
			start, ok := token.(xml.StartElement)
			if !ok {
				continue
			}
			if depth == 0 && rootSeen {
				c.fail("%s: invalid SVG: multiple root elements.", name)
			}
			depth++
			if !rootSeen {
				rootSeen = true
				if start.Name.Local != "svg" || start.Name.Space != "http://www.w3.org/2000/svg" {
					c.fail("%s: missing SVG root element or namespace.", name)
				}
			}
			if start.Name.Local == "image" && start.Name.Space == "http://www.w3.org/2000/svg" {
				c.fail("%s: vector icons must not embed raster images.", name)
			}
		}
		file.Close()
		if !rootSeen {
			c.fail("%s: missing SVG root element or namespace.", name)
		}
	}
}

var categoriesPattern = regexp.MustCompile(`(?s)const ACADEMIC_ACTIVITY_CATEGORIES\s*=\s*Object\.freeze\(\[(.*?)\]\);`)
var quotedCategory = regexp.MustCompile(`"([a-z_]+)"`)

func (c *checker) checkCV() {
	data := c.read("app/assets/js/data.js")
	cv := string(c.read("app/assets/js/cv.js"))
	want := []string{"editorial_service", "professional_service", "conference_service", "invited_talks", "reviews"}
	var actual []string
	if match := categoriesPattern.FindSubmatch(data); match != nil {
		for _, name := range quotedCategory.FindAllSubmatch(match[1], -1) {
			actual = append(actual, string(name[1]))
		}
	}
	if strings.Join(actual, ",") != strings.Join(want, ",") {
		c.fail("app/assets/js/data.js: academic activity CV category order must be %s.", strings.Join(want, ", "))
	}
	for _, item := range [][2]string{
		{"academic_activities: renderAcademicActivities", "academic activities are not registered as a CV section renderer"},
		{"D.ACADEMIC_ACTIVITY_CATEGORIES.map((category)", "academic activity CV groups do not use the canonical category order"},
		{"(data[category] || []).filter(D.visibleInCV)", "academic activity CV items do not apply visible_in_CV"},
		{`if (!items.length) return "";`, "empty academic activity CV groups are not suppressed"},
		{`return content.trim() ? section(t(sectionName), content) : "";`, "empty CV sections are not suppressed"},
	} {
		if !strings.Contains(cv, item[0]) {
			c.fail("app/assets/js/cv.js: %s.", item[1])
		}
	}
}
