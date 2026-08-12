package sitebuild

import (
	"fmt"
	"github.com/Gonie-Gonie/Gonie-Gonie.github.io/editor/internal/dataversion"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// Build validates repository data and optionally regenerates its four exports.
// All migrations are applied only in memory; source JSON is never rewritten.
func Build(root string, validateOnly bool) (string, error) {
	absolute, e := filepath.Abs(root)
	if e != nil {
		return "", e
	}
	data, e := readData(absolute)
	if e != nil {
		return "", fmt.Errorf("ERROR: could not read data: %w", e)
	}
	if errors := validateData(absolute, data); len(errors) > 0 {
		return "", fmt.Errorf("Validation failed:\n  - %s", strings.Join(errors, "\n  - "))
	}
	output := ""
	if !validateOnly {
		if e = generate(absolute, data); e != nil {
			return "", e
		}
		output = "Generated exports in: " + filepath.Join(absolute, "exports") + "\n"
	}
	output += fmt.Sprintf("Validated %d publications, %d projects, and %d software records.\n", len(array(data["publications"])), len(array(data["projects"])), len(array(data["software"])))
	return output, nil
}

// Validate checks all site data and file references without writing exports.
func Validate(root string) error { _, e := Build(root, true); return e }

func latexEscape(s string) string {
	replacements := map[rune]string{'\\': `\textbackslash{}`, '&': `\&`, '%': `\%`, '$': `\$`, '#': `\#`, '_': `\_`, '{': `\{`, '}': `\}`, '~': `\textasciitilde{}`, '^': `\textasciicircum{}`}
	var out strings.Builder
	for _, r := range s {
		if replacement, ok := replacements[r]; ok {
			out.WriteString(replacement)
		} else {
			out.WriteRune(r)
		}
	}
	return out.String()
}
func citationKey(item record) string {
	source := ""
	if pyBool(item["doi"]) {
		source = pyString(item["doi"])
	} else if pyBool(item["url"]) {
		source = pyString(item["url"])
	}
	if source == "" {
		authors := []string{}
		for _, x := range array(item["authors"]) {
			authors = append(authors, personName(obj(x), "en"))
		}
		source = strings.Join([]string{localized(item, "title", "en"), str(item["date"]), strings.Join(authors, "|")}, "|")
	}
	hash := uint32(2166136261)
	for _, r := range source {
		hash ^= uint32(r)
		hash *= 16777619
	}
	date := year(item)
	if date == "" {
		date = "nd"
	}
	return "publication-" + date + "-" + strconv.FormatUint(uint64(hash), 36)
}
func dateLabel(item record) string {
	if s := year(item); s != "" {
		return s
	}
	if s := publicationStatus(item); s != "" {
		return s
	}
	return "n.d."
}
func bibtexEntry(item record) string {
	entryType, container := "inproceedings", "booktitle"
	if isJournal(item) {
		entryType, container = "article", "journal"
	}
	authors := []string{}
	for _, x := range array(item["authors"]) {
		authors = append(authors, latexEscape(personName(obj(x), "en")))
	}
	fields := []string{"  title = {" + latexEscape(localized(item, "title", "en")) + "}", "  author = {" + strings.Join(authors, " and ") + "}", "  year = {" + dateLabel(item) + "}", "  " + container + " = {" + latexEscape(str(item["venue_en"])) + "}"}
	for _, field := range []string{"doi", "url"} {
		if pyBool(item[field]) {
			fields = append(fields, "  "+field+" = {"+pyString(item[field])+"}")
		}
	}
	if s := localized(item, "abstract", "en"); s != "" {
		fields = append(fields, "  abstract = {"+latexEscape(s)+"}")
	}
	if keywords := localizedList(item, "keywords", "en"); len(keywords) > 0 {
		fields = append(fields, "  keywords = {"+latexEscape(strings.Join(keywords, "; "))+"}")
	}
	return "@" + entryType + "{" + citationKey(item) + ",\n" + strings.Join(fields, ",\n") + "\n}"
}
func risEntry(item record) string {
	kind, container := "CPAPER", "T2"
	if isJournal(item) {
		kind, container = "JOUR", "JO"
	}
	lines := []string{"TY  - " + kind, "TI  - " + localized(item, "title", "en")}
	for _, x := range array(item["authors"]) {
		lines = append(lines, "AU  - "+personName(obj(x), "en"))
	}
	lines = append(lines, "PY  - "+dateLabel(item))
	if s := text(item["venue_en"]); s != "" {
		lines = append(lines, container+"  - "+s)
	}
	for _, field := range [][2]string{{"doi", "DO"}, {"url", "UR"}} {
		if pyBool(item[field[0]]) {
			lines = append(lines, field[1]+"  - "+pyString(item[field[0]]))
		}
	}
	if s := strings.TrimSpace(localized(item, "abstract", "en")); s != "" {
		lines = append(lines, "AB  - "+s)
	}
	for _, s := range localizedList(item, "keywords", "en") {
		lines = append(lines, "KW  - "+s)
	}
	lines = append(lines, "ER  -")
	return strings.Join(lines, "\n")
}

type jsonField struct {
	key   string
	value any
}
type orderedObject []jsonField

func cslEntry(item record) orderedObject {
	kind := "paper-conference"
	if isJournal(item) {
		kind = "article-journal"
	}
	authors := []any{}
	for _, x := range array(item["authors"]) {
		name := personName(obj(x), "en")
		if family, given, ok := strings.Cut(name, ","); ok {
			authors = append(authors, orderedObject{{"family", strings.TrimSpace(family)}, {"given", strings.TrimSpace(given)}})
		} else {
			authors = append(authors, orderedObject{{"literal", strings.TrimSpace(name)}})
		}
	}
	entry := orderedObject{{"id", citationKey(item)}, {"type", kind}, {"title", localized(item, "title", "en")}, {"author", authors}, {"container-title", str(item["venue_en"])}}
	if parts := dateParts(item); len(parts) > 0 {
		values := []any{}
		for _, n := range parts {
			values = append(values, n)
		}
		entry = append(entry, jsonField{"issued", orderedObject{{"date-parts", []any{values}}}})
	}
	if s := publicationStatus(item); s != "" {
		entry = append(entry, jsonField{"status", s})
	}
	for _, field := range [][2]string{{"doi", "DOI"}, {"url", "URL"}} {
		if pyBool(item[field[0]]) {
			entry = append(entry, jsonField{field[1], item[field[0]]})
		}
	}
	if s := localized(item, "abstract", "en"); s != "" {
		entry = append(entry, jsonField{"abstract", s})
	}
	if keywords := localizedList(item, "keywords", "en"); len(keywords) > 0 {
		entry = append(entry, jsonField{"keyword", strings.Join(keywords, "; ")})
	}
	return entry
}
func jsonQuote(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\b':
			b.WriteString(`\b`)
		case '\f':
			b.WriteString(`\f`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		default:
			if r < 32 {
				fmt.Fprintf(&b, `\u%04x`, r)
			} else {
				b.WriteRune(r)
			}
		}
	}
	b.WriteByte('"')
	return b.String()
}

// Python's ensure_ascii=False JSON formatting is retained byte-for-byte, including
// insertion order, Unicode, HTML characters and two-space indentation.
func prettyJSON(value any, level int) string {
	indent := strings.Repeat("  ", level)
	childIndent := indent + "  "
	switch x := value.(type) {
	case orderedObject:
		if len(x) == 0 {
			return "{}"
		}
		parts := []string{}
		for _, field := range x {
			parts = append(parts, childIndent+jsonQuote(field.key)+": "+prettyJSON(field.value, level+1))
		}
		return "{\n" + strings.Join(parts, ",\n") + "\n" + indent + "}"
	case []any:
		if len(x) == 0 {
			return "[]"
		}
		parts := []string{}
		for _, item := range x {
			parts = append(parts, childIndent+prettyJSON(item, level+1))
		}
		return "[\n" + strings.Join(parts, ",\n") + "\n" + indent + "]"
	case string:
		return jsonQuote(x)
	case int:
		return strconv.Itoa(x)
	case nil:
		return "null"
	default:
		return fmt.Sprint(value)
	}
}
func generate(root string, data record) error {
	bundleVersion, err := dataversion.BundleVersion(root)
	if err != nil {
		return err
	}
	people := map[string]record{}
	for _, x := range array(data["people"]) {
		r := obj(x)
		people[str(r["id"])] = r
	}
	publications := []record{}
	for _, x := range array(data["publications"]) {
		r := ResolvePublication(obj(x), obj(data["settings"]))
		authors := []any{}
		for _, id := range splitList(r["author_ids"]) {
			authors = append(authors, people[id])
		}
		r["authors"] = authors
		publications = append(publications, r)
	}
	sort.SliceStable(publications, func(i, j int) bool {
		a, b := publications[i], publications[j]
		if statusRank(a) != statusRank(b) {
			return statusRank(a) > statusRank(b)
		}
		ap, bp := append(dateParts(a), 0, 0, 0), append(dateParts(b), 0, 0, 0)
		for n := 0; n < 3; n++ {
			if ap[n] != bp[n] {
				return ap[n] > bp[n]
			}
		}
		return lower(localized(a, "title", "en")) < lower(localized(b, "title", "en"))
	})
	bib, ris, csl := []string{}, []string{}, []any{}
	for _, r := range publications {
		bib = append(bib, bibtexEntry(r))
		ris = append(ris, risEntry(r))
		csl = append(csl, cslEntry(r))
	}
	counts := orderedObject{}
	for _, name := range []string{"projects", "software", "publications", "people", "awards", "board"} {
		counts = append(counts, jsonField{name, len(array(data[name]))})
	}
	activities := 0
	for _, category := range activityOrder {
		activities += len(array(obj(data["academic_activities"])[category]))
	}
	counts = append(counts, jsonField{"academic_activities", activities})
	counts = append(counts, jsonField{"patents", len(array(obj(data["profile"])["patents"]))})
	typeCounts := map[string]int{}
	for _, x := range array(data["publications"]) {
		typeCounts[str(obj(x)["publication_type"])]++
	}
	keys := []string{}
	for key := range typeCounts {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	types := orderedObject{}
	for _, key := range keys {
		types = append(types, jsonField{key, typeCounts[key]})
	}
	report := orderedObject{{"data_version", bundleVersion}, {"counts", counts}, {"publication_types", types}}
	files := []struct{ name, content string }{{"publications.bib", strings.Join(bib, "\n\n") + "\n"}, {"publications.ris", strings.Join(ris, "\n\n") + "\n"}, {"publications.csl.json", prettyJSON(csl, 0) + "\n"}, {"build-report.json", prettyJSON(report, 0) + "\n"}}
	directory := filepath.Join(root, "exports")
	if e := os.MkdirAll(directory, 0755); e != nil {
		return e
	}
	for _, file := range files {
		if e := os.WriteFile(filepath.Join(directory, file.name), []byte(file.content), 0644); e != nil {
			return e
		}
	}
	return nil
}
