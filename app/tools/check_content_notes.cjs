"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const vm = require("node:vm");
const root = path.resolve(__dirname, "../..");
const versions = require("../assets/js/data-version.js");
const source = fs.readFileSync(path.join(root, "app/assets/js/data.js"), "utf8");
const readJSON = (name) => JSON.parse(fs.readFileSync(path.join(root, name), "utf8"));
const plain = (value) => JSON.parse(JSON.stringify(value));

function environment(overrides) {
  const window = {
    SiteDataVersion: versions,
    PublicationBibliography: require("../assets/js/publication-bibliography.js"),
  };
  const fetch = async (name) => ({
    ok: true,
    json: async () => structuredClone(Object.hasOwn(overrides, name) ? overrides[name] : readJSON(name)),
  });
  vm.runInNewContext(source, { window, fetch, Intl, Date, URL });
  return window.SiteData;
}

async function main() {
  for (const collection of ["projects", "software"]) {
    const file = `data/${collection}.json`;
    const render = collection === "projects" ? "projectNotes" : "softwareNotes";
    for (const [englishCount, koreanCount] of [[1, 3], [3, 1]]) {
      const items = versions.decode(readJSON(file), file, { collection: true }).data;
      const english = Array.from({ length: englishCount }, (_, index) => `English description ${index + 1}`);
      const korean = Array.from({ length: koreanCount }, (_, index) => `Korean description ${index + 1}`);
      Object.assign(items[0], { notes_en: english, notes_kr: korean });
      const before = JSON.stringify(items);
      const D = environment({ [file]: items });
      const loaded = (await D.loadAll())[collection][0];
      assert.deepEqual(plain(loaded.notes_en), english);
      assert.deepEqual(plain(loaded.notes_kr), korean);
      assert.deepEqual(plain(D[render](loaded, "en")), english, `${collection}: all English notes retain their order`);
      assert.deepEqual(plain(D[render](loaded, "ko")), korean, `${collection}: all Korean notes retain their order`);
      assert.equal(JSON.stringify(items), before, "Loading never pads or truncates source arrays");

      const longer = englishCount > koreanCount ? "notes_en" : "notes_kr";
      for (const invalid of ["", " ", null, 42]) {
        const invalidItems = structuredClone(items);
        invalidItems[0][longer][invalidItems[0][longer].length - 1] = invalid;
        await assert.rejects(environment({ [file]: invalidItems }).loadAll(), /notes_(en|kr)/);
      }
      for (const language of ["notes_en", "notes_kr"]) {
        const invalidItems = structuredClone(items);
        invalidItems[0][language] = [];
        await assert.rejects(environment({ [file]: invalidItems }).loadAll(), /non-empty array/);
      }
    }
  }
  console.log("Content notes: independent English/Korean counts, language selection, order, source preservation and unmatched-entry validation passed.");
}

main().catch((error) => { console.error(error); process.exitCode = 1; });
