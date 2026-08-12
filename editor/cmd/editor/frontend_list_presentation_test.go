package main

import (
	"strings"
	"testing"
)

func listPresentationAsset(t *testing.T, filename string) string {
	t.Helper()
	content, err := assets.ReadFile("frontend/dist/" + filename)
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}

func listPresentationFunction(t *testing.T, source, name string) string {
	t.Helper()
	start := strings.Index(source, "function "+name+"(")
	if start < 0 {
		t.Fatalf("missing function %s", name)
	}
	end := strings.Index(source[start+1:], "\nfunction ")
	if end < 0 {
		return source[start:]
	}
	return source[start : start+1+end]
}

func TestEditorSidebarGroupsUseSeparatorsWithoutHeadings(t *testing.T) {
	html := listPresentationAsset(t, "index.html")
	start := strings.Index(html, `<nav class="sidebar-nav"`)
	if start < 0 {
		t.Fatal("missing sidebar navigation")
	}
	end := strings.Index(html[start:], "</nav>")
	if end < 0 {
		t.Fatal("missing sidebar navigation end")
	}
	nav := html[start : start+end]
	if strings.Count(nav, `class="nav-group"`) != 3 {
		t.Error("sidebar must retain its three navigation groups")
	}
	for _, obsolete := range []string{"<h2", "콘텐츠", "참조 데이터", "출력 설정"} {
		if strings.Contains(nav, obsolete) {
			t.Errorf("sidebar still contains obsolete group label %q", obsolete)
		}
	}
	css := listPresentationAsset(t, "styles.css")
	start = strings.Index(css, ".nav-group + .nav-group {")
	if start < 0 {
		t.Fatal("missing adjacent navigation group separator")
	}
	end = strings.Index(css[start:], "}")
	if end < 0 || !strings.Contains(css[start:start+end], "border-top: 1px solid var(--color-line)") {
		t.Error("navigation groups must have a thin separator")
	}
}

func TestEditorProjectListShowsFunderInsteadOfTheme(t *testing.T) {
	js := listPresentationAsset(t, "app.js")
	render := listPresentationFunction(t, js, "renderContentList")
	if !strings.Contains(render, `item.funder_ko || item.funder_en || "발주처 미입력"`) {
		t.Error("project list must prefer Korean funder with English and empty fallbacks")
	}
	if strings.Contains(render, "projectThemeLabel(") {
		t.Error("project list still displays its theme instead of its funder")
	}
	if !strings.Contains(render, "subtitle.title = subtitle.textContent") {
		t.Error("project list funder must remain readable when truncated")
	}
}

func TestEditorAwardListUsesLinkedPublicationIcon(t *testing.T) {
	js := listPresentationAsset(t, "app.js")
	render := listPresentationFunction(t, js, "renderEntityListRow")
	for _, snippet := range []string{
		`publication.award_key === item._clientKey`,
		`if (collection === "people" && usage)`,
		`badge.className = "board-badge entity-usage-badge"`,
		`remove.disabled = state.saving || state.discarding || usage > 0`,
		`if (linkedPublications.length)`,
		`actions.classList.add("award-linked-actions")`,
		`actions.append(makeAwardPublicationLink(linkedPublications))`,
	} {
		if !strings.Contains(render, snippet) {
			t.Errorf("award list is missing %q", snippet)
		}
	}
	linkPosition := strings.Index(render, "actions.append(makeAwardPublicationLink(linkedPublications))")
	deletePosition := strings.Index(render, "actions.append(remove)")
	if linkPosition < 0 || deletePosition <= linkPosition {
		t.Error("publication icon must precede the trash button")
	}
	link := listPresentationFunction(t, js, "makeAwardPublicationLink")
	for _, snippet := range []string{
		`link.className = "award-publication-link"`,
		`link.title = publications.map((publication) => publication.title_ko || publication.title_en || "(제목 미입력)").join("\n")`,
		`link.setAttribute("aria-label",`,
		`link.tabIndex = 0`,
		`document.createElementNS(namespace, "svg")`,
	} {
		if !strings.Contains(link, snippet) {
			t.Errorf("publication link indicator is missing %q", snippet)
		}
	}
	css := listPresentationAsset(t, "styles.css")
	start := strings.Index(css, ".award-linked-actions {")
	if start < 0 {
		t.Fatal("missing linked award action layout")
	}
	end := strings.Index(css[start:], "}")
	if end < 0 || !strings.Contains(css[start:start+end], "display: flex") {
		t.Error("publication icon and trash must share a horizontal action row")
	}
}

func TestEditorBoardListOmitsDuplicateEnglishTitle(t *testing.T) {
	js := listPresentationAsset(t, "app.js")
	render := listPresentationFunction(t, js, "renderBoardList")
	if strings.Contains(render, "board-item-subtitle") || strings.Contains(render, "subtitle") {
		t.Error("board list still contains the duplicate title subtitle")
	}
	if !strings.Contains(render, `title.textContent = item.title_ko || item.title_en || "(제목 미입력)"`) {
		t.Error("board title must retain its English fallback when Korean is empty")
	}
	if !strings.Contains(render, "select.append(date, title, badges)") {
		t.Error("board list must retain its date, primary title, and media/state badges")
	}
	editor := listPresentationFunction(t, js, "renderBoardEditor")
	if !strings.Contains(editor, `"title_en"`) || !strings.Contains(editor, `"title_ko"`) {
		t.Error("board form must still support both language title fields")
	}
}

func TestEditorProfileBiographyWithoutIdentityRoleInBasicInfo(t *testing.T) {
	js := listPresentationAsset(t, "app.js")
	render := listPresentationFunction(t, js, "renderProfileContentEditor")
	start := strings.Index(render, `const basic = profileStaticSection("기본 정보",`)
	end := strings.Index(render, `}, { id: "profile-section-basic"`)
	if start < 0 || end <= start {
		t.Fatal("missing basic profile section")
	}
	basic := render[start:end]
	previous := -1
	for _, field := range []string{
		`appendProfilePair(grid, "이름", ["identity"], "name_en", "name_ko");`,
		`appendProfilePair(grid, "약력", ["intro"], "short_en", "short_ko", { textarea: true });`,
		`appendProfilePair(grid, "소속", ["identity"], "affiliation_en", "affiliation_ko");`,
		`appendProfilePair(grid, "직함", ["profile_card"], "title_en", "title_ko");`,
		`appendProfilePair(grid, "위치", ["identity"], "location_en", "location_ko", { textarea: true });`,
	} {
		position := strings.Index(basic, field)
		if position < 0 || position <= previous {
			t.Fatalf("basic profile info must follow name, biography, affiliation, title, location order with unchanged data paths; misplaced field %s", field)
		}
		previous = position
	}
	if strings.Count(basic, "appendProfilePair(") != 5 {
		t.Error("basic profile info must contain exactly its five bilingual field pairs")
	}
	if strings.Contains(basic, `"role_en"`) || strings.Contains(basic, `"role_ko"`) {
		t.Error("basic profile info must not expose the removed identity role fields")
	}
	if strings.Count(render, `appendProfilePair(grid, "약력"`) != 1 || strings.Contains(render, `appendProfilePair(grid, "소개"`) {
		t.Error("profile biography must have exactly one pair of input fields with its updated label")
	}
	for _, obsolete := range []string{`profile-section-intro`, `profileStaticSection("소개"`} {
		if strings.Contains(js, obsolete) {
			t.Errorf("obsolete standalone intro section or navigation remains: %s", obsolete)
		}
	}
	if !strings.Contains(render, "sections.append(basic, contact)") {
		t.Error("contact section must follow basic info without a separate intro section")
	}
}
