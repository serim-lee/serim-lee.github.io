"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const vm = require("node:vm");
const root = path.resolve(__dirname, "../..");
const versions = require("../assets/js/data-version.js");
const source = fs.readFileSync(path.join(root, "app/assets/js/data.js"), "utf8");
const profile = versions.decode(JSON.parse(fs.readFileSync(path.join(root, "data/profile.json"), "utf8")), "profile.json").data;
const settings = versions.decode(JSON.parse(fs.readFileSync(path.join(root, "data/settings.json"), "utf8")), "settings.json").data;
const actualPeople = versions.decode(JSON.parse(fs.readFileSync(path.join(root, "data/people.json"), "utf8")), "people.json", { collection: true }).data;
const usedPeopleIDs = new Set(actualPeople.map((person) => person.id));
const fixturePeople = ["one", "two"].map((suffix) => {
  let id = `patent-test-inventor-${suffix}`;
  while (usedPeopleIDs.has(id)) id += "-test";
  usedPeopleIDs.add(id);
  return { id, name_en: `Test Inventor ${suffix}`, name_ko: `시험 발명자 ${suffix}`, is_self: false, notes_en: [], notes_ko: [] };
}).concat(actualPeople);
const plain = (value) => JSON.parse(JSON.stringify(value));

function environment(overrides = {}) {
  const window = { SiteDataVersion: versions, PublicationBibliography: require("../assets/js/publication-bibliography.js") };
  const fetch = async (name) => name === "data/version.json" && (overrides[name] === null || (!Object.hasOwn(overrides, name) && !fs.existsSync(path.join(root, name))))
    ? { ok: false, status: 404 } : ({
    ok: true,
    json: async () => structuredClone(Object.hasOwn(overrides, name)
      ? overrides[name]
      : name === "data/people.json" ? fixturePeople : JSON.parse(fs.readFileSync(path.join(root, name), "utf8"))),
  });
  vm.runInNewContext(source, { window, fetch, Intl, Date, URL });
  return window.SiteData;
}

function loadPatents(patents, customSettings = settings) {
  return environment({
    ...(!Object.hasOwn(customSettings, "version") && Object.hasOwn(customSettings, "schema_version") ? { "data/version.json": null } : {}),
    "data/profile.json": { ...structuredClone(profile), patents },
    "data/settings.json": customSettings,
  }).loadAll();
}

async function main() {
  const D = environment();
  const people = (await D.loadAll()).people;
  const minimal = { title_ko: "시험 특허", type: "domestic", inventor_ids: [people[0].id], application_number: "10-2024-0000001" };
  const full = {
    ...minimal,
    title_en: "Test patent",
    type: "international",
    inventor_ids: [people[1].id, people[0].id],
    applicant_en: "Example University",
    applicant_ko: "예시대학교",
    application_date: "2024-02-29",
    publication_number: "10-2025-0000001",
    publication_date: "2025-08-29",
    registration_number: "10-0000001",
    registration_date: "2026-01-15",
    url: "https://patents.example.org/patent/1?lang=ko&view=full",
  };
  const records = [{ ...full, visible_in_CV: false }, minimal];
  const loaded = await loadPatents(records);
  assert.deepEqual(plain(loaded.profile.patents.map(({ inventors, ...item }) => item)), records, "Patent records keep the editor's order and CV selection.");
  assert.deepEqual(plain(loaded.profile.patents[0].inventors), [people[1], people[0]], "Inventors follow the selected order rather than people.json order.");
  assert.equal(D.visibleInCV(loaded.profile.patents[0]), false);
  assert.equal(D.visibleInCV(loaded.profile.patents[1]), true);
  assert.deepEqual(plain((await loadPatents([])).profile.patents), []);
  const withoutPatents = structuredClone(profile);
  delete withoutPatents.patents;
  assert.deepEqual(plain((await environment({ "data/profile.json": withoutPatents }).loadAll()).profile.patents), []);
  for (const invalid of [null, {}, "", [null], [42]]) {
    await assert.rejects(loadPatents(invalid), /patents/);
  }
  for (const [numbers, korean, english] of [
    [{ application_number: "A" }, "출원", "Filed"],
    [{ publication_number: "P" }, "공개", "Published"],
    [{ registration_number: "R" }, "등록", "Granted"],
    [{ application_number: "A", publication_number: "P" }, "공개", "Published"],
    [{ application_number: "A", publication_number: "P", registration_number: "R" }, "등록", "Granted"],
    [{ application_number: " A ", publication_number: " ", registration_number: "\t" }, "출원", "Filed"],
  ]) {
    assert.equal(D.patentStatus({ ...numbers, status: "rejected" }, "ko"), korean);
    assert.equal(D.patentStatus(numbers, "en"), english);
  }
  assert.equal(D.patentStatus({ status: "granted" }, "ko"), "");
  assert.equal(D.patentType(minimal, "ko"), "국내특허");
  assert.equal(D.patentType(full, "ko"), "국제특허");
  assert.equal(D.patentType(minimal, "en"), "Domestic Patents");
  assert.equal(D.patentType(full, "en"), "International Patents");
  const groupedItems = [
    { ...minimal, type: "international", title_en: "International filed" },
    { ...full, type: "domestic", title_en: "Domestic granted" },
    { ...minimal, title_en: "Domestic filed first" },
    { ...minimal, publication_number: "P", title_en: "Domestic published" },
    { ...minimal, title_en: "Domestic filed second" },
    { ...full, title_en: "International granted" },
  ];
  assert.deepEqual(plain(D.patentGroups(groupedItems, "ko").map((group) => [group.title, group.items.map((item) => item.title_en)])), [
    ["국내특허 [출원]", ["Domestic filed first", "Domestic filed second"]],
    ["국내특허 [공개]", ["Domestic published"]],
    ["국내특허 [등록]", ["Domestic granted"]],
    ["국제특허 [출원]", ["International filed"]],
    ["국제특허 [등록]", ["International granted"]],
  ]);
  assert.deepEqual(plain(D.patentGroups([], "ko")), []);
  assert.deepEqual(plain(D.patentGroups(loaded.profile.patents.filter(D.visibleInCV), "en").map((group) => group.title)), ["Domestic Patents [Filed]"]);
  assert.equal(D.patentGroups([{ ...full, registration_number: "" }], "ko")[0].title, "국제특허 [공개]");
  assert.equal(D.patentTitle({ ...minimal, title_en: " " }, "en"), minimal.title_ko);
  assert.equal(D.patentPeopleHTML(loaded.profile.patents[0], "en"), `${D.displayAuthors([people[1], people[0]], "en", false)} (filed by Example University)`);
  assert.equal(D.patentDate(full), "2026-01-15");
  assert.equal(D.patentDate(minimal), "");
  assert.equal(D.patentURL(full), full.url);
  assert.equal(D.patentURL({ url: "javascript:alert(1)" }), "");
  assert.equal(D.patentURL({ url: "mailto:inventor@example.org" }), "");
  assert.equal(D.patentPeopleHTML(loaded.profile.patents[0], "ko"), `${D.displayAuthors([people[1], people[0]], "ko", false)} (예시대학교 출원)`);
  const renamedPeople = structuredClone(people);
  renamedPeople[0].name_ko = "변경된 발명자";
  renamedPeople[1].name_en = "";
  const renamed = await environment({ "data/people.json": renamedPeople, "data/profile.json": { ...profile, patents: [full] } }).loadAll();
  assert.ok(D.patentPeopleHTML(renamed.profile.patents[0], "ko").includes("변경된 발명자"));
  assert.ok(D.patentPeopleHTML(renamed.profile.patents[0], "en").includes(D.personName(renamedPeople[1], "ko")));
  assert.ok(!D.patentPeopleHTML(loaded.profile.patents[0], "ko").includes("tooltip"));
  const withoutInventors = { ...full };
  delete withoutInventors.inventor_ids;
  assert.deepEqual(plain((await loadPatents([withoutInventors])).profile.patents[0].inventor_ids), []);
  const noInventors = (await loadPatents([{ ...full, inventor_ids: [] }])).profile.patents[0];
  assert.deepEqual(plain(noInventors.inventors), []);
  assert.equal(D.patentPeopleHTML(noInventors, "ko"), "(예시대학교 출원)");
  assert.equal(D.patentPeopleHTML({ ...noInventors, applicant_en: "", applicant_ko: "" }, "ko"), "");
  assert.ok(D.patentPeopleHTML({ ...noInventors, applicant_ko: "<b>산학협력단</b>" }, "ko").includes("&lt;b&gt;산학협력단&lt;/b&gt;"));
  assert.ok(!D.patentDetails(full, "en").some((line) => line.includes("Applicant:")));
  assert.ok(!D.patentDetails(full, "ko").some((line) => line.includes("출원일")));
  assert.ok(!D.patentDetails(full, "en").some((line) => line.includes(full.application_date)));
  assert.ok(D.patentDetails(full, "en").some((line) => line.includes("10-0000001") && line.includes("2026-01-15")));
  assert.ok(D.patentDetails(minimal, "en").some((line) => line.includes(minimal.application_number)));
  assert.ok(!D.patentDetails(minimal, "en").some((line) => /undefined|null/.test(line)));
  for (const url of [
    `https://example.com/${"가".repeat(4076)}`,
    `https://example.com/${"😀".repeat(4076)}`,
    "HTTPS://example.com/patent",
    "https://특허.example/문헌",
    "https://[2001:db8::1]:65535/patent",
  ]) {
    assert.equal((await loadPatents([{ ...full, url }])).profile.patents[0].url, url);
  }
  const invalidChanges = [
    [{ title_ko: "", title_en: " " }, /title/],
    [{ type: "" }, /type/],
    [{ type: "overseas" }, /type/],
    [{ type: null }, /type/],
    [{ inventor_ids: null }, /inventor_ids/],
    [{ inventor_ids: people[0].id }, /inventor_ids/],
    [{ inventor_ids: [people[0].id, people[0].id] }, /duplicate/],
    [{ inventor_ids: ["unknown-patent-inventor"] }, /inventor_ids/],
    [{ inventor_ids: [1] }, /inventor_ids/],
    [{ application_number: "", publication_number: "", registration_number: "" }, /number/],
    [{ application_date: "2023-02-29" }, /application_date/],
    [{ application_date: "0000-01-01" }, /application_date/],
    [{ application_date: "2024-13-01" }, /application_date/],
    [{ application_date: " 2024-02-29" }, /application_date/],
    [{ publication_date: "2024-02-28" }, /publication_date/],
    [{ registration_date: "2024-02-28" }, /registration_date/],
    [{ url: "javascript:alert(1)" }, /url/],
    [{ url: "mailto:inventor@example.org" }, /url/],
    [{ url: "https://patents.example.org/\nunsafe" }, /url/],
    [{ url: "https://example.com/\u0085unsafe" }, /url/],
    [{ url: "https://:443/patent" }, /url/],
    [{ url: "https://example.com:65536/patent" }, /url/],
    [{ url: `https://example.com/${"😀".repeat(4077)}` }, /url/],
    [{ url: `https://patents.example.org/${"a".repeat(4096)}` }, /url/],
    [{ visible_in_CV: "false" }, /visible_in_CV/],
    [{ registration_number: 12345 }, /registration_number/],
  ];
  for (const [change, message] of invalidChanges) {
    await assert.rejects(loadPatents([{ ...full, ...change }]), message);
  }
  for (const field of ["type"]) {
    const missing = { ...full };
    delete missing[field];
    await assert.rejects(loadPatents([missing]), new RegExp(field));
  }

  const oldSettings = structuredClone(settings);
  delete oldSettings.version;
  oldSettings.schema_version = 8;
  for (const key of ["main_page_sections", "hidden_main_page_sections", "cv_sections", "hidden_cv_sections"]) {
    oldSettings[key] = oldSettings[key].filter((section) => section !== "patents");
  }
  oldSettings.main_page_sections.reverse();
  oldSettings.cv_sections.reverse();
  const migrated = await loadPatents([], oldSettings);
  for (const prefix of ["main_page", "cv"]) {
    const visible = `${prefix}_sections`;
    const hidden = `hidden_${prefix}_sections`;
    assert.deepEqual(plain(migrated.settings[visible]), oldSettings[visible]);
    assert.deepEqual(plain(migrated.settings[hidden]), [...oldSettings[hidden], "patents"]);
  }
  const legacySettings = structuredClone(oldSettings);
  legacySettings.schema_version = 5;
  legacySettings.main_page_sections = legacySettings.main_page_sections.filter((section) => section !== "academic_activities");
  legacySettings.hidden_main_page_sections = legacySettings.hidden_main_page_sections.filter((section) => section !== "academic_activities");
  const legacy = await loadPatents([], legacySettings);
  assert.deepEqual(plain(legacy.settings.hidden_main_page_sections), [...legacySettings.hidden_main_page_sections, "academic_activities", "patents"]);
  legacySettings.schema_version = 4;
  delete legacySettings.cv_sections;
  delete legacySettings.hidden_cv_sections;
  const preCV = await loadPatents([], legacySettings);
  assert.ok(!preCV.settings.cv_sections.includes("patents"));
  assert.deepEqual(plain(preCV.settings.hidden_cv_sections), ["patents"]);
  const malformed = structuredClone(oldSettings);
  malformed.cv_sections.push(malformed.cv_sections[0]);
  await assert.rejects(loadPatents([], malformed), /cv_sections/);
  const currentMissingSection = structuredClone(oldSettings);
  currentMissingSection.schema_version = 9;
  await assert.rejects(loadPatents([], currentMissingSection), /cv_sections/);
  const futureSettings = structuredClone(settings);
  futureSettings.version = "1.5.7";
  const future = await environment({ "data/settings.json": futureSettings, "data/profile.json": { ...profile, patents: [] }, "data/version.json": { version: "1.5.7" } }).loadAll();
  assert.equal(future.version, "1.5.7");
  assert.equal(future.settings.version, "1.5.7");
  const currentWithoutCV = structuredClone(futureSettings);
  delete currentWithoutCV.cv_sections;
  delete currentWithoutCV.hidden_cv_sections;
  currentWithoutCV.schema_version = 4;
  await assert.rejects(loadPatents([], currentWithoutCV), /cv_sections/);
  await assert.rejects(environment({ "data/settings.json": oldSettings, "data/version.json": { version: "1.0.0" } }).loadAll(), /cv_sections/,
    "A bundle marker prevents old numeric settings migrations even when obsolete schema metadata remains.");
  const withoutMetadata = structuredClone(currentWithoutCV);
  delete withoutMetadata.version;
  delete withoutMetadata.schema_version;
  await assert.rejects(environment({ "data/settings.json": withoutMetadata, "data/version.json": null }).loadAll(), /cv_sections/,
    "A missing bundle marker alone must not activate schema migrations without a numeric legacy schema.");
  await assert.rejects(environment({ "data/version.json": { version: "2.0.0" } }).loadAll(), /unsupported version/);
  console.log("Patents: inventor references, number-derived status groups, localized metadata, CV selection, manual order and legacy section migrations checked.");
}

main().catch((error) => { console.error(error); process.exitCode = 1; });
