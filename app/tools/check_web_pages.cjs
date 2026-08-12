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
const pages = ["publications", "projects", "software", "board"];

function environment(settings) {
  const window = {
    SiteDataVersion: versions,
    PublicationBibliography: require("../assets/js/publication-bibliography.js"),
  };
  const fetch = async (name) => ({
    ok: true,
    json: async () => structuredClone(name === "data/settings.json" ? settings : readJSON(name)),
  });
  vm.runInNewContext(source, { window, fetch, Intl, Date, URL });
  return window.SiteData;
}

async function main() {
  const original = readJSON("data/settings.json");
  const legacy = structuredClone(original);
  delete legacy.hidden_web_pages;
  const baseline = await environment(legacy).loadAll();
  assert.deepEqual(plain(baseline.settings.hidden_web_pages), [], "Existing websites keep all tabs visible");
  assert.equal(Object.hasOwn(legacy, "hidden_web_pages"), false, "Loading does not rewrite source settings");

  for (let mask = 0; mask < 16; mask += 1) {
    const hidden = pages.filter((_, index) => mask & (1 << index));
    const settings = { ...structuredClone(original), hidden_web_pages: hidden };
    const before = JSON.stringify(settings);
    const D = environment(settings);
    const loaded = await D.loadAll();
    assert.deepEqual(plain(loaded.settings.hidden_web_pages), hidden);
    for (const page of pages) {
      assert.equal(D.webPageVisible(loaded.settings, page), !hidden.includes(page));
      assert.deepEqual(plain(loaded[page]), plain(baseline[page]), `${page}: web visibility preserves every record and CV inclusion`);
    }
    assert.deepEqual(plain(loaded.settings.cv_sections), plain(baseline.settings.cv_sections));
    assert.deepEqual(plain(loaded.settings.hidden_cv_sections), plain(baseline.settings.hidden_cv_sections));
    assert.deepEqual(plain(loaded.settings.main_page_sections), plain(baseline.settings.main_page_sections));
    assert.equal(JSON.stringify(settings), before);
  }

  for (const invalid of [null, "software", {}, 1, ["home"], ["cv"], ["Software"], ["software", "software"], [false]]) {
    await assert.rejects(environment({ ...structuredClone(original), hidden_web_pages: invalid }).loadAll(), /hidden_web_pages/);
  }
  console.log("Web pages: all visibility combinations preserve content and CV settings; legacy defaults and invalid settings checked.");
}

main().catch((error) => { console.error(error); process.exitCode = 1; });
