"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const vm = require("node:vm");
const versions = require("../assets/js/data-version.js");
const root = path.resolve(__dirname, "../..");
const source = fs.readFileSync(path.join(root, "app/assets/js/publication-citations.js"), "utf8");
const doi = (number) => `10.1234/paper${number}`;
const paper = (number, date = "2024-01-01", extra = {}) => ({ doi: doi(number), date, ...extra });
const recency = (a, b) => b.date.localeCompare(a.date);
const oaRecord = (count = 5) => ({ provider: "openalex", work_id: "W101", count, checked_at: "2024-01-01" });
const kciRecord = { provider: "kci", article_id: "ART123", count: 1, checked_at: "2025-01-01", list_pending: true };

function environment({ records = {}, snapshotVersion = null, legacySnapshot = false, storage = new Map(), respond, storageDisabled = false, quickTimeout = false } = {}) {
  const calls = [];
  const window = {
    SiteDataVersion: versions,
    SiteData: { truthy: (value) => value === true || value === "true" },
    setTimeout: (callback, delay) => setTimeout(callback, quickTimeout ? 5 : delay), clearTimeout,
    localStorage: {
      getItem(key) { if (storageDisabled) throw new Error("Storage disabled"); return storage.get(key) || null; },
      setItem(key, value) { if (storageDisabled) throw new Error("Storage disabled"); storage.set(key, value); }
    }
  };
  const fetch = async (url, options) => {
    if (url === "exports/publication-citations.json" && legacySnapshot) return { ok: false, status: 404 };
    if (url === "exports/publication-citations.json" || url === "data/publication-citations.json") {
      return { ok: true, json: async () => snapshotVersion === null ? { records } : { version: snapshotVersion, records } };
    }
    calls.push(url);
    return respond ? respond(url, options) : { ok: true, json: async () => ({ results: [] }) };
  };
  vm.runInNewContext(source, { window, fetch, URLSearchParams, AbortController });
  return { api: window.PublicationCitations, calls, storage };
}

async function main() {
  const { api } = environment();
  for (const snapshotVersion of [null, 1, "1.0.0", "1.8.2"]) {
    const snapshot = environment({ records: { [doi(1)]: oaRecord(17) }, snapshotVersion });
    const client = snapshot.api.createClient();
    await client.start([paper(1)], () => {});
    assert.equal(client.get(paper(1)).count, 17, "Legacy and compatible versioned snapshots remain readable.");
    assert.equal(JSON.parse(snapshot.storage.get("portfolio-publication-citations-v1")).version, 1, "Browser cache version stays independent of file version.");
  }
  const legacy = environment({ records: { [doi(1)]: oaRecord(17) }, snapshotVersion: 1, legacySnapshot: true });
  const legacyClient = legacy.api.createClient();
  await legacyClient.start([paper(1)], () => {});
  assert.equal(legacyClient.get(paper(1)).count, 17, "Old data snapshot remains readable when the export snapshot is absent.");
  for (const value of ["10.1234/Paper1", "doi: 10.1234/Paper1", "doi.org/10.1234/Paper1", "https://dx.doi.org/10.1234/Paper1 "]) {
    assert.equal(api.normalizeDOI(value), doi(1));
  }
  for (const value of ["", "garbage", "https://evil.example/10.1234/Paper1", "10.1234/two words"]) assert.equal(api.normalizeDOI(value), "");
  assert.equal(api.citingURL(oaRecord()), "https://openalex.org/works?filter=cites:W101");
  assert.ok(api.citingURL(kciRecord).includes("artiId=ART123#:~:text="));

  let release;
  const publications = [paper(1, "2023-01-01"), paper(2, "2025-01-01"), paper(3), paper(4), paper(5, "2026-01-01"),
    paper(6, "2027-01-01", { under_review: true }), paper(7, "2027-01-01", { in_press: true }), { date: "2022-01-01" }];
  const records = { [doi(1)]: oaRecord(), [doi(4)]: kciRecord, [doi(6)]: oaRecord(100), [doi(7)]: oaRecord(200) };
  const storage = new Map();
  const env = environment({ records, storage, respond: () => new Promise((resolve) => { release = resolve; }) });
  const client = env.api.createClient();
  let updates = 0;
  let complete = false;
  const task = client.start(publications, () => { updates += 1; }).then(() => { complete = true; });
  await new Promise(setImmediate);
  assert.equal(complete, false, "Slow APIs stay asynchronous");
  assert.equal(client.get(publications[0]).count, 5, "Snapshot available before the external API responds");
  assert.equal(client.get(publications[3]).count, 1);
  assert.equal(client.get(publications[5]), null, "Under-review counts are suppressed");
  assert.equal(client.get(publications[6]), null, "In-press counts are suppressed");
  const filter = new URL(env.calls[0]).searchParams.get("filter");
  assert.ok(!filter.includes(doi(4)) && !filter.includes(doi(6)) && !filter.includes(doi(7)));
  assert.ok(!env.calls[0].includes("api_key"), "Public requests do not contain a key");
  release({ ok: true, json: async () => ({ results: [
    { id: "https://openalex.org/W101", doi: "https://doi.org/10.1234/PAPER1", cited_by_count: 9 },
    { id: "https://openalex.org/W102", doi: doi(2), cited_by_count: 9 },
    { id: "https://openalex.org/W103", doi: doi(3), cited_by_count: 0 },
    { id: "https://openalex.org/W105", doi: doi(5), cited_by_count: null },
    { id: "https://openalex.org/W999", doi: doi(999), cited_by_count: 999 }
  ] }) });
  await task;
  assert.equal(updates, 2);
  assert.equal(client.get(publications[0]).count, 9);
  assert.equal(client.get(publications[2]).count, 0, "A matched zero remains a real count");
  assert.equal(client.get(publications[4]), null, "Unknown/malformed counts are not zero");
  const sorted = [...publications].sort((a, b) => client.compare(a, b, recency));
  assert.deepEqual(sorted.slice(0, 4).map((item) => item.doi), [doi(2), doi(1), doi(4), doi(3)], "Count descending, then recency; known zero precedes unknown");
  assert.ok(sorted.indexOf(publications[5]) < sorted.indexOf(publications[4]), "Unknown records retain recency ordering");

  const second = environment({ records, storage });
  const secondClient = second.api.createClient();
  await secondClient.start(publications, () => {});
  assert.equal(second.calls.length, 0, "Daily cache includes both successful and unmatched lookups");
  assert.equal(secondClient.get(publications[0]).count, 9, "Older snapshot cannot overwrite newer browser results");

  const outage = environment({ records, respond: async () => ({ ok: false, status: 429 }) });
  const outageClient = outage.api.createClient();
  await outageClient.start(publications, () => {});
  assert.equal(outageClient.get(publications[0]).checked_at, "2024-01-01", "Outages retain the actual retrieval date");
  assert.equal(outageClient.get(publications[4]), null);
  const outageAgain = environment({ records, storage: outage.storage });
  await outageAgain.api.createClient().start(publications, () => {});
  assert.equal(outageAgain.calls.length, 0, "Failures have a retry cooldown");

  const timeout = environment({ records, quickTimeout: true, storageDisabled: true,
    respond: (url, options) => new Promise((resolve, reject) => options.signal.addEventListener("abort", () => reject(new Error("Timed out")))) });
  const timeoutClient = timeout.api.createClient();
  await timeoutClient.start(publications, () => {});
  assert.equal(timeoutClient.get(publications[0]).count, 5, "Timeouts and unavailable storage are tolerated");

  const corruptedStorage = new Map([["portfolio-publication-citations-v1", "{corrupt"]]);
  const corrupt = environment({ records, storage: corruptedStorage });
  await corrupt.api.createClient().start(publications, () => {});
  assert.equal(corrupt.calls.length, 1, "Corrupt cache does not prevent requests");
  const oldKCI = new Map([["portfolio-publication-citations-v1", JSON.stringify({ version: 1,
    records: { [doi(4)]: { ...kciRecord, count: 900, checked_at: "2099-01-01" } } })]]);
  const corrected = environment({ records, storage: oldKCI });
  const correctedClient = corrected.api.createClient();
  await correctedClient.start(publications, () => {});
  assert.equal(correctedClient.get(publications[3]).count, 1, "KCI corrections on disk override browser storage");

  const many = environment();
  await many.api.createClient().start(Array.from({ length: 101 }, (_, index) => paper(index + 1000)), () => {});
  assert.equal(many.calls.length, 2, "More than 100 unique DOIs use bounded batches");
  for (const url of many.calls) assert.ok(new URL(url).searchParams.get("filter").split("|").length <= 100);

  const checkedIn = versions.decode(JSON.parse(fs.readFileSync(path.join(root, "exports/publication-citations.json"), "utf8")), "publication-citations.json").data;
  const allPapers = versions.decode(JSON.parse(fs.readFileSync(path.join(root, "data/publications.json"), "utf8")), "publications.json", { collection: true }).data;
  const published = new Set(allPapers.filter((item) => !item.under_review && !item.in_press).map((item) => api.normalizeDOI(item.doi)));
  assert.equal(versions.validateVersion(checkedIn.version), checkedIn.version);
  for (const [key, record] of Object.entries(checkedIn.records)) {
    assert.ok(key && key === api.normalizeDOI(key) && published.has(key), "Snapshot maps to a published paper");
    assert.ok(Number.isSafeInteger(record.count) && record.count >= 0);
    assert.ok(Number.isFinite(Date.parse(record.checked_at)));
    assert.ok(record.provider === "openalex" ? /^W\d+$/.test(record.work_id) : record.provider === "kci" && /^ART\d+$/.test(record.article_id));
  }
  console.log("Publication citations: normalization, batching, async delivery, sorting, cache, failures, timeouts and snapshot checks passed.");
}

main().catch((error) => { console.error(error); process.exitCode = 1; });
