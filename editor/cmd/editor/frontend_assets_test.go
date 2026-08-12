package main

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestAcademicActivitiesEditorAssetsAreWired(t *testing.T) {
	htmlBytes, err := assets.ReadFile("frontend/dist/index.html")
	if err != nil {
		t.Fatal(err)
	}
	jsBytes, err := assets.ReadFile("frontend/dist/app.js")
	if err != nil {
		t.Fatal(err)
	}
	cssBytes, err := assets.ReadFile("frontend/dist/styles.css")
	if err != nil {
		t.Fatal(err)
	}
	html := string(htmlBytes)
	js := string(jsBytes)
	css := string(cssBytes)
	for _, snippet := range []string{
		`id="tab-academic-activities"`,
		`id="panel-academic-activities"`,
		`id="academic-activities-item-list"`,
		`id="academic-activities-form"`,
	} {
		if !strings.Contains(html, snippet) {
			t.Errorf("editor HTML is missing %s", snippet)
		}
	}
	for _, snippet := range []string{
		`academic_activities_revision: state.academicActivitiesRevision`,
		`save_academic_activities: academicActivitiesChanged`,
		`completed_dates: (item.completed_date_rows || []).map`,
		`academic-review-date-add`,
		`academic-review-date-delete`,
		`리뷰 ${(item.completed_date_rows || []).length}회`,
		`type: "date"`,
		`Invited Talk · 초청 발표`,
		`ACADEMIC_ACTIVITY_CATEGORIES.map(renderAcademicActivityGroup)`,
		`add.dataset.academicActivityAdd = category.value`,
		`"academic-activity-up"`,
		`"academic-activity-down"`,
		`function moveAcademicActivity(clientKey, direction)`,
		`found.items.filter((item) => item.category === found.item.category)`,
		`[found.items[found.index], found.items[targetIndex]] = [found.items[targetIndex], found.items[found.index]]`,
		`up.disabled = up.disabled || position === 0`,
		`down.disabled = down.disabled || position === categoryLength - 1`,
		`function moveAcademicActivityToCategoryTop(found, category)`,
		`found.items.splice(targetIndex < 0 ? found.items.length : targetIndex, 0, found.item)`,
		`const reviewDateRows = category === "reviews"`,
		`function academicActivitiesSourceComparable(items)`,
		`const visibility = { visible_in_CV: item.visible_in_CV !== false }`,
		`["academic_activities", academicActivities]`,
		`if (section === "academic_activities") return state.academicActivitiesDraft || []`,
		`function makeCVAcademicActivityGroups(items)`,
		`makeCVItemRow("academic_activities", item, items.indexOf(item))`,
		`body.append(makeCVAcademicActivityGroups(items))`,
	} {
		if !strings.Contains(js, snippet) {
			t.Errorf("editor JavaScript is missing %q", snippet)
		}
	}
	if strings.Contains(js, "초청 강연") || strings.Contains(js, `type: "datetime-local"`) {
		t.Error("editor still contains obsolete invited-talk wording or time input")
	}
	if strings.Contains(js, "review_count") {
		t.Error("editor still contains the obsolete numeric review_count field")
	}
	if strings.Contains(html, `id="add-academic-activity-button"`) || strings.Contains(js, `#add-academic-activity-button`) {
		t.Error("obsolete global academic activity add button is still wired")
	}
	for _, snippet := range []string{
		`.academic-activity-list-group`,
		`.academic-activity-group-heading`,
		`.academic-activity-category-list`,
		`.cv-academic-activity-groups`,
		`.cv-academic-activity-group-heading`,
		`.cv-academic-activity-item-list`,
	} {
		if !strings.Contains(css, snippet) {
			t.Errorf("editor CSS is missing %q", snippet)
		}
	}
	if strings.Contains(js, `makeAcademicActivityField(item, "CV`) {
		t.Error("academic activity edit form unexpectedly exposes a CV visibility field")
	}
	cvLabelsStart := strings.Index(js, "const CV_SECTION_LABELS = {")
	if cvLabelsStart < 0 {
		t.Fatal("CV section labels were not found")
	}
	cvLabelsEnd := strings.Index(js[cvLabelsStart:], "};")
	if cvLabelsEnd < 0 {
		t.Fatal("CV section labels were not terminated")
	}
	cvLabels := js[cvLabelsStart : cvLabelsStart+cvLabelsEnd]
	if awardsIndex, academicIndex := strings.Index(cvLabels, `awards:`), strings.Index(cvLabels, `academic_activities:`); awardsIndex < 0 || academicIndex <= awardsIndex {
		t.Error("academic_activities is not configured after awards in CV section labels")
	}
	previous := -1
	for _, category := range []string{"editorial_service", "professional_service", "conference_service", "invited_talks", "reviews"} {
		index := strings.Index(js, `{ value: "`+category+`"`)
		if index < 0 || index <= previous {
			t.Fatalf("academic activity category order is incorrect at %q", category)
		}
		previous = index
	}
}

func TestPeopleListsUseKoreanThenEnglishDisplayOrder(t *testing.T) {
	jsBytes, err := assets.ReadFile("frontend/dist/app.js")
	if err != nil {
		t.Fatal(err)
	}
	js := string(jsBytes)
	for _, snippet := range []string{
		`const HANGUL_NAME_PATTERN = /[\uac00-\ud7a3]/;`,
		`new Intl.Collator("ko-KR"`,
		`new Intl.Collator("en-US"`,
		`function sortedPeopleForDisplay(items)`,
		`const owners = sortedPeopleForDisplay(rawItems.filter((item) => item.is_self))`,
		`const coauthors = sortedPeopleForDisplay(rawItems.filter((item) => !item.is_self))`,
		`const candidates = sortedPeopleForDisplay(`,
		`if (input.dataset.publicationAuthorSelect !== undefined) {`,
		`addPublicationAuthor(publicationForm?.dataset.entityKey || "", input.value)`,
		`addRow.append(select)`,
		`people: toEntitySavePayload(state.peopleDraft, state.peopleBaseline, "person", toPersonPayload)`,
	} {
		if !strings.Contains(js, snippet) {
			t.Errorf("editor JavaScript is missing %q", snippet)
		}
	}
	peopleConfigStart := strings.Index(js, "  people: {")
	if peopleConfigStart < 0 {
		t.Fatal("people entity configuration was not found")
	}
	peopleConfigEnd := strings.Index(js[peopleConfigStart:], "  awards: {")
	if peopleConfigEnd < 0 {
		t.Fatal("people entity configuration was not terminated")
	}
	peopleConfig := js[peopleConfigStart : peopleConfigStart+peopleConfigEnd]
	if !strings.Contains(peopleConfig, "manualOrder: false") || !strings.Contains(peopleConfig, "sorter: sortedPeopleForDisplay") {
		t.Error("people entity list still permits manual ordering or lacks automatic display sorting")
	}
	if strings.Contains(js, `coauthorList.dataset.manualOrderList = "entity"`) ||
		strings.Contains(js, `coauthorList.dataset.manualOrderCollection = "people"`) {
		t.Error("people coauthor list still enables drag-based manual ordering")
	}
	if strings.Contains(js, `publication-author-add`) {
		t.Error("publication author picker still requires a separate add button")
	}
}

func TestScholarshipAmountEditorAssetsAreWired(t *testing.T) {
	htmlBytes, err := assets.ReadFile("frontend/dist/index.html")
	if err != nil {
		t.Fatal(err)
	}
	jsBytes, err := assets.ReadFile("frontend/dist/app.js")
	if err != nil {
		t.Fatal(err)
	}
	cssBytes, err := assets.ReadFile("frontend/dist/styles.css")
	if err != nil {
		t.Fatal(err)
	}
	html := string(htmlBytes)
	js := string(jsBytes)
	css := string(cssBytes)
	for _, snippet := range []string{
		`id="scholarship-currency-codes"`,
		`<option value="KRW">`,
		`<option value="USD">`,
	} {
		if !strings.Contains(html, snippet) {
			t.Errorf("editor HTML is missing %s", snippet)
		}
	}
	for _, snippet := range []string{
		`amount: { value: "", currency: "KRW" }`,
		`profileField("금액", [...base, "amount", "value"]`,
		`profileField("통화", [...base, "amount", "currency"]`,
		`list: "scholarship-currency-codes"`,
		`input.value.toUpperCase()`,
		`cvDetailLine("금액", scholarshipAmountLabel(item.amount))`,
		`(result.scholarships || []).forEach((item) => delete item.details)`,
	} {
		if !strings.Contains(js, snippet) {
			t.Errorf("editor JavaScript is missing %q", snippet)
		}
	}
	for _, obsolete := range []string{"profile-detail-", "renderScholarshipDetails", "mutateScholarshipDetail", "scholarship.details"} {
		if strings.Contains(js, obsolete) {
			t.Errorf("editor JavaScript still contains obsolete scholarship details UI %q", obsolete)
		}
	}
	if !strings.Contains(css, ".profile-scholarship-amount") {
		t.Error("editor CSS is missing the compact scholarship amount layout")
	}
}

var (
	editorTechnologyPattern = regexp.MustCompile(`\{ key: "([^"]+)", value: "([^"]+)", label: "([^"]+)", icon: "([^"]+)" \}`)
	siteTechnologyPattern   = regexp.MustCompile(`(?m)^\s+(?:"([^"]+)"|([a-z][a-z0-9+]*)):\s+\{ label: "([^"]+)", icon: "app/assets/icons/technologies/([^"]+)"`)
)

func TestEditorTechnologyPickerUsesSelectionOnly(t *testing.T) {
	js, err := assets.ReadFile("frontend/dist/app.js")
	if err != nil {
		t.Fatal(err)
	}
	content := string(js)
	start := strings.Index(content, "function makeTechnologyField(")
	end := strings.Index(content, "function makeEditorFormSection(")
	if start < 0 || end <= start {
		t.Fatal("technology picker function not found")
	}
	picker := content[start:end]
	for _, snippet := range []string{
		`document.createElement("select")`,
		`placeholder.textContent = "선택"`,
		`if (rawValue.trim()) control.append(icon)`,
		`error.className = "sr-only"`,
		`select.setAttribute("aria-describedby", error.id)`,
	} {
		if !strings.Contains(picker, snippet) {
			t.Errorf("technology picker is missing %q", snippet)
		}
	}
	if strings.Contains(picker, `document.createElement("input")`) {
		t.Error("technology picker must not offer free-text input")
	}
	if !strings.Contains(content, `state.validationErrors.set(key, "기술을 선택해 주세요.")`) {
		t.Error("empty technology selection must retain selection-specific validation")
	}
}

func TestEditorTechnologyCatalogMatchesSiteIcons(t *testing.T) {
	editorJS, err := assets.ReadFile("frontend/dist/app.js")
	if err != nil {
		t.Fatalf("read embedded editor app.js: %v", err)
	}

	type technology struct {
		value string
		label string
		icon  string
	}
	editorCatalog := make(map[string]technology)
	for _, match := range editorTechnologyPattern.FindAllSubmatch(editorJS, -1) {
		key := string(match[1])
		if _, exists := editorCatalog[key]; exists {
			t.Fatalf("editor technology key %q is duplicated", key)
		}
		entry := technology{value: string(match[2]), label: string(match[3]), icon: string(match[4])}
		editorCatalog[key] = entry
		if filepath.Ext(entry.icon) != ".svg" {
			t.Errorf("editor technology %q must use an SVG icon, got %q", key, entry.icon)
		}
		if _, readErr := assets.ReadFile("frontend/dist/assets/technologies/" + entry.icon); readErr != nil {
			t.Errorf("editor technology %q icon %q is not embedded: %v", key, entry.icon, readErr)
		}
	}

	if len(editorCatalog) != 25 {
		t.Fatalf("editor technology catalog has %d entries, want 25", len(editorCatalog))
	}
	for key, wantValue := range map[string]string{"c#": "C#", "rhino": "Rhino", "grasshopper": "Grasshopper"} {
		if got := editorCatalog[key].value; got != wantValue {
			t.Errorf("editor technology %q value = %q, want %q", key, got, wantValue)
		}
	}

	repoRoot, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatalf("resolve repository root: %v", err)
	}
	siteJS, err := os.ReadFile(filepath.Join(repoRoot, "app", "assets", "js", "app.js"))
	if err != nil {
		t.Fatalf("read site app.js: %v", err)
	}
	siteCatalog := make(map[string]technology)
	for _, match := range siteTechnologyPattern.FindAllSubmatch(siteJS, -1) {
		key := string(match[1])
		if key == "" {
			key = string(match[2])
		}
		siteCatalog[key] = technology{label: string(match[3]), icon: string(match[4])}
	}

	for key, editorEntry := range editorCatalog {
		siteEntry, exists := siteCatalog[key]
		if !exists {
			t.Errorf("editor technology %q is missing from the site catalog", key)
			continue
		}
		if siteEntry.label != editorEntry.label || siteEntry.icon != editorEntry.icon {
			t.Errorf(
				"technology %q differs: editor label/icon %q/%q, site %q/%q",
				key, editorEntry.label, editorEntry.icon, siteEntry.label, siteEntry.icon,
			)
		}
		editorIcon, readErr := assets.ReadFile("frontend/dist/assets/technologies/" + editorEntry.icon)
		if readErr != nil {
			t.Errorf("read embedded editor technology %q icon: %v", key, readErr)
			continue
		}
		siteIcon, readErr := os.ReadFile(filepath.Join(repoRoot, "app", "assets", "icons", "technologies", siteEntry.icon))
		if readErr != nil {
			t.Errorf("read site technology %q icon: %v", key, readErr)
			continue
		}
		if !bytes.Equal(editorIcon, siteIcon) {
			t.Errorf("technology %q icon differs between the editor and site assets", key)
		}
	}
	if len(siteCatalog) != len(editorCatalog) {
		t.Errorf("site technology catalog has %d entries, editor has %d", len(siteCatalog), len(editorCatalog))
	}
}
