"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const vm = require("node:vm");
const { spawnSync } = require("node:child_process");
const root = path.resolve(__dirname, "../..");
const bibliography = require("../assets/js/publication-bibliography.js");
const versions = require("../assets/js/data-version.js");
const read = (file) => JSON.parse(fs.readFileSync(path.join(root, file), "utf8"));
const settings = versions.decode(read("data/settings.json"), "settings.json").data;
const publications = versions.decode(read("data/publications.json"), "publications.json", { collection: true }).data;
const fixtureSettings = {
  journals: [{ id: "journal", name_en: "Test Journal", name_ko: "시험 저널" }],
  conferences: [
    { id: "Conf. 2024", name_en: "Test Conference", proceedings_title_en: "Proceedings of Test Conference", start_date: "2024-12-30", end_date: "2025-01-02", proceedings_date: "2025-04", location_en: "Seoul, Korea" },
    { id: "Conf. 2026", name_en: "Test Conference", start_date: "2026-10-01", end_date: "2026-10-03" }
  ]
};
const fixtures = [
  { publication_type: "domestic-journal", journal_id: "journal", volume: "40", issue: "1", pages: "169-176", issue_date: "2025-10-21", online_date: "2025-08-13", date: "1900", venue_en: "Ignored legacy text" },
  { publication_type: "international-journal", journal_id: "journal", volume: "372", elocator: "118260", online_date: "2026-09-23" },
  { publication_type: "international-conference", conference_id: "Conf. 2024", pages: "10-20", date: "1900" },
  { publication_type: "domestic-conference", conference_id: "Conf. 2026" },
  { publication_type: "international-journal", date: "1900" }
];
const expectedDates = ["2025-10-21", "2026-09-23", "2025-04", "2026-10-03", ""];
assert.deepEqual(fixtures.map((item) => bibliography.publicationDate(item, fixtureSettings)), expectedDates);
assert.equal(bibliography.publicationVenue(fixtures[0], fixtureSettings, "ko"), "시험 저널, 제40권 제1호, pp. 169-176.");
assert.equal(bibliography.publicationVenue(fixtures[1], fixtureSettings, "ko"), "Test Journal, Vol. 372, 118260.");
assert.equal(bibliography.publicationVenue(fixtures[2], fixtureSettings), "Proceedings of Test Conference, pp. 10-20. Test Conference, December 30, 2024 – January 2, 2025, Seoul, Korea.");
assert.equal(bibliography.publicationVenue(fixtures[4], fixtureSettings), "");
assert.deepEqual(bibliography.publicationReferenceParts(fixtures[2], fixtureSettings), {
  citation: "Proceedings of Test Conference, pp. 10-20.",
  conference: "Test Conference, December 30, 2024 – January 2, 2025, Seoul, Korea."
});
assert.deepEqual(bibliography.publicationReferenceParts(fixtures[0], fixtureSettings, "ko"), {
  citation: "시험 저널, 제40권 제1호, pp. 169-176.", conference: ""
});
assert.deepEqual(bibliography.publicationReferenceParts(fixtures[3], fixtureSettings), {
  citation: "", conference: "Test Conference, October 1-3, 2026."
});
for (const value of ["SCIE", "SCOPUS", "KCI"]) {
  assert.equal(bibliography.publicationIndexing(fixtures[0], { journals: [{ id: "journal", indexing: [value] }] }), value);
}
assert.equal(bibliography.publicationIndexing(fixtures[4], fixtureSettings), "");
assert.equal(bibliography.publicationIndexing(fixtures[0], { journals: [{ id: "journal", indexing: ["unknown"] }] }), "");
assert.deepEqual(bibliography.resolvePublication(fixtures[0], fixtureSettings).date, "2025-10-21");
const context = { window: { SiteDataVersion: versions } };
vm.runInNewContext(fs.readFileSync(path.join(root, "app/assets/js/data.js"), "utf8"), context);
const siteData = context.window.SiteData;
const citation = {
  authors: [{ name_en: "First Author", name_ko: "첫째 저자" }, { name_en: "Second Author", name_ko: "둘째 저자" }],
  title_en: "English Paper Title", title_ko: "국문 논문 제목", date: "2026-10-19",
  venue_en: "English Venue", venue_kr: "국문 게재지", doi: "10.1234/example"
};
const englishCitation = "First Author, Second Author (2026). English Paper Title. English Venue https://doi.org/10.1234/example";
const koreanCitation = "첫째 저자, 둘째 저자 (2026). 국문 논문 제목. 국문 게재지 https://doi.org/10.1234/example";
for (const publication_type of ["international-journal", "international-conference"]) {
  const item = { ...citation, publication_type };
  assert.equal(siteData.publicationCitation(item, "ko"), englishCitation);
  assert.equal(siteData.publicationCitation(item, "en"), englishCitation);
  assert.equal(siteData.publicationCitation(item), englishCitation);
}
for (const publication_type of ["domestic-journal", "domestic-conference"]) {
  const item = { ...citation, publication_type };
  assert.equal(siteData.publicationCitation(item, "ko"), koreanCitation);
  assert.equal(siteData.publicationCitation(item, "en"), englishCitation);
}
const numberedCases = [];
for (const fixture of [
  { title: "Proceedings", series: "Research Series", volume: "40", issue: "2", seriesVolume: "7", expected: "Proceedings, Vol. 40, No. 2, Research Series, Vol. 7, pp. 10-20." },
  { title: "Proceedings", series: "", seriesVolume: "7", expected: "Proceedings, Series Vol. 7, pp. 10-20." },
  { title: "Proceedings Vol. 40, No. 2", volume: "40", issue: "2", expected: "Proceedings Vol. 40, No. 2, pp. 10-20." },
  { title: "Proceedings Volume 40, Issue 2", volume: "40", issue: "2", expected: "Proceedings Volume 40, Issue 2, pp. 10-20." },
  { title: "Proceedings Vol. 132, Part 1", volume: "132", issue: "1", expected: "Proceedings Vol. 132, Part 1, pp. 10-20." },
  { title: "Proceedings Vol.\u00a02", volume: "2", expected: "Proceedings Vol.\u00a02, pp. 10-20." },
  { title: "Proceedings", series: "Research Series, Vol. 7", seriesVolume: "7", expected: "Proceedings, Research Series, Vol. 7, pp. 10-20." },
  { title: "Proceedings in Research Series", series: "Research Series", seriesVolume: "7", expected: "Proceedings in Research Series, Series Vol. 7, pp. 10-20." },
  { title: "Proceedings of X", series: "Proceedings", seriesVolume: "591", expected: "Proceedings of X, Series Vol. 591, pp. 10-20." },
  { title: "Proceedings in Research Series, Vol. 7", series: "Research Series", seriesVolume: "7", expected: "Proceedings in Research Series, Vol. 7, pp. 10-20." },
  { title: "Proceedings Vol. 40", volume: "4", expected: "Proceedings Vol. 40, Vol. 4, pp. 10-20." },
  { title: "Proceedings Vol. 40(1)", volume: "40(1)", expected: "Proceedings Vol. 40(1), pp. 10-20." }
]) {
  const venue = { id: "Numbered Conf.", name_en: "Test Conference", name_ko: "시험 학술대회",
    proceedings_title_en: fixture.title, proceedings_title_ko: "시험 학술대회 논문집",
    proceedings_series_en: fixture.series || "", proceedings_series_ko: fixture.series ? "연구 시리즈" : "",
    proceedings_volume: fixture.volume || "", proceedings_issue: fixture.issue || "", proceedings_series_volume: fixture.seriesVolume || "" };
  const numberedSettings = { conferences: [venue] };
  const international = { publication_type: "international-conference", conference_id: venue.id, pages: "10-20", volume: "stale", issue: "stale" };
  const domestic = { ...international, publication_type: "domestic-conference" };
  const before = JSON.stringify([numberedSettings, international, domestic]);
  const result = bibliography.resolvePublication(international, numberedSettings);
  assert.equal(result.reference_en, fixture.expected);
  assert.equal(result.reference_ko, fixture.expected);
  assert.equal(result.volume, fixture.volume || "");
  assert.equal(result.issue, fixture.issue || "");
  assert.equal(JSON.stringify([numberedSettings, international, domestic]), before);
  numberedCases.push({ settings: numberedSettings, items: [international, domestic] });
}
const domesticNumbering = bibliography.resolvePublication(numberedCases[0].items[1], numberedCases[0].settings);
assert.equal(domesticNumbering.reference_ko, "시험 학술대회 논문집, 제40권 제2호, 연구 시리즈, 제7권, pp. 10-20.");
assert.equal(bibliography.resolvePublication(numberedCases[1].items[1], numberedCases[1].settings).reference_ko, "시험 학술대회 논문집, 시리즈 제7권, pp. 10-20.");
const unspacedKorean = { conferences: [{ id: "Korean Conf.", proceedings_title_ko: "논문집 제2권제1호", proceedings_volume: "2", proceedings_issue: "1" }] };
const unspacedItem = { publication_type: "domestic-conference", conference_id: "Korean Conf." };
assert.equal(bibliography.resolvePublication(unspacedItem, unspacedKorean).reference_ko, "논문집 제2권제1호.");
numberedCases.push({ settings: unspacedKorean, items: [unspacedItem] });
const blankNumbering = { conferences: [{ ...fixtureSettings.conferences[0], proceedings_volume: " ", proceedings_issue: "", proceedings_series_volume: "" }] };
assert.deepEqual(bibliography.resolvePublication(fixtures[2], blankNumbering), bibliography.resolvePublication(fixtures[2], fixtureSettings));
const cases = [{ settings: fixtureSettings, items: fixtures }, { settings, items: publications }, ...numberedCases];
const runtime = path.join(root, "editor/runtime");
const portableGo = path.join(runtime, "go/bin", process.platform === "win32" ? "go.exe" : "go");
const goEnv = { ...process.env };
if (fs.existsSync(portableGo)) {
  Object.assign(goEnv, { GOROOT: path.join(runtime, "go"), GOPATH: path.join(runtime, "gopath"),
    GOMODCACHE: path.join(runtime, "gomodcache"), GOCACHE: path.join(runtime, "gocache") });
}
const native = spawnSync(fs.existsSync(portableGo) ? portableGo : "go", ["run", "./cmd/site", "bibliography"],
  { cwd: path.join(root, "editor"), env: goEnv, input: JSON.stringify(cases), encoding: "utf8" });
assert.equal(native.status, 0, native.stderr || String(native.error || ""));
assert.deepEqual(JSON.parse(native.stdout), cases.map((c) => c.items.map((item) => {
  const result = bibliography.resolvePublication(item, c.settings);
  return Object.fromEntries(["date", "venue_en", "venue_kr", "reference_en", "reference_ko", "conference_info_en", "conference_info_ko", "indexing"].map(key => [key, result[key]]));
})));
assert.equal(fs.readFileSync(path.join(root, "editor/cmd/editor/frontend/dist/publication-bibliography.js"), "utf8"), fs.readFileSync(path.join(root, "app/assets/js/publication-bibliography.js"), "utf8"));
assert.equal(publications.some((p) => "venue_en" in p || "venue_kr" in p || "venue" in p), false);
assert.ok(fixtures.length > 0, "bibliography fixtures must not be empty");
console.log(`Publication bibliography: date precedence, precision, bilingual references, English international citations, missing metadata, and browser/editor/Go parity for all ${publications.length} papers passed.`);
