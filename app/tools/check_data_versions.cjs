"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const V = require("../assets/js/data-version.js");
const root = path.resolve(__dirname, "../..");
const collectionNames = ["people", "projects", "software", "publications", "awards", "board"];

assert.equal(V.CURRENT_VERSION, "1.0.0");
for (const version of ["1.0.0", "1.12.34", "1.999.999", "1.9007199254740993.9007199254740993", "1.18446744073709551615.0", "1.0.18446744073709551615"]) {
  const source = { version, future_field: { keep: true } };
  const decoded = V.decode(source, "future object");
  assert.equal(decoded.version, version);
  assert.equal(decoded.data.version, version);
  assert.equal(decoded.legacy, false);
  assert.deepEqual(JSON.parse(JSON.stringify(decoded.data)), source, "Compatible future versions and unknown fields survive round trips.");
  const collection = V.decode({ version, items: [{ id: "one" }] }, "future collection", { collection: true });
  assert.equal(collection.version, version);
  assert.deepEqual(collection.data, [{ id: "one" }]);
}
for (const version of ["2.0.0", "0.9.0", "10.0.0"]) {
  assert.throws(() => V.decode({ version }, "object"), /unsupported version/);
  assert.throws(() => V.decode({ version, items: [] }, "collection", { collection: true }), /unsupported version/);
}
for (const version of [null, false, 2, "", "1", "1.0", "01.0.0", "1.01.0", "1.0.01", "1.0.0-01", "1.2.3-beta.2", "1.2.3+build.007", " 1.0.0", "1.0.0 ", "1.0.0-", "1.0.0+", "1.0.0+bad field", "18446744073709551616.0.0", "1.18446744073709551616.0", "1.0.18446744073709551616"]) {
  assert.throws(() => V.decode({ version }, "bad version"), /version/);
}
for (const suffix of ["\n", "\r", "\r\n", "\u2028", "\u2029"]) {
  assert.throws(() => V.validateVersion(`1.0.0${suffix}`), /version/, "Trailing line terminators must match Go's strict end-of-text validation.");
}
for (const schema_version of [1, 4, 5, 8, 9]) {
  const source = { schema_version, value: "kept" };
  const decoded = V.decode(source, "legacy");
  assert.equal(decoded.legacy, true);
  assert.equal(decoded.legacySchemaVersion, schema_version);
  assert.deepEqual(decoded.data, { version: "1.0.0", value: "kept" });
  assert.equal(source.schema_version, schema_version, "Reading does not mutate source metadata.");
}
assert.deepEqual(V.decode({ version: 1, records: {} }, "legacy citations").data, { version: "1.0.0", records: {} });
const explicit = V.decode({ version: "1.7.8", schema_version: 4, extra: true }, "current settings");
assert.equal(explicit.legacy, false);
assert.equal(explicit.legacySchemaVersion, null, "A current version must not trigger numeric schema migrations.");
assert.deepEqual(explicit.data, { version: "1.7.8", extra: true });
assert.deepEqual(V.decode([{ id: "old" }], "legacy collection", { collection: true }).data, [{ id: "old" }]);
assert.deepEqual(V.decode([], "empty legacy collection", { collection: true }).data, []);
assert.deepEqual(V.decode({ version: "1.0.0", items: [] }, "empty collection", { collection: true }).data, []);
for (const source of [{ value: "kept" }, { schema_version: 4, value: "kept" }]) {
  const decoded = V.decode(source, "current bundle object", { bundleVersion: "1.7.8" });
  assert.equal(decoded.legacy, false);
  assert.equal(decoded.legacySchemaVersion, null);
  assert.equal(decoded.version, "1.7.8");
  assert.deepEqual(decoded.data, { value: "kept", version: "1.7.8" });
}
const bundleCollection = V.decode([], "current bundle collection", { collection: true, bundleVersion: "1.7.8" });
assert.equal(bundleCollection.version, "1.7.8");
assert.equal(bundleCollection.legacy, false);
assert.throws(() => V.decode({}, "current bundle", { bundleVersion: "2.0.0" }), /unsupported version/);
for (const source of [null, false, 42, "[]", {}, { items: [] }, { version: "1.0.0", items: null }, { version: "1.0.0", items: {} }]) {
  assert.throws(() => V.decode(source, "invalid collection", { collection: true }), /object|version|items/);
}
assert.throws(() => V.decode([], "invalid object"), /object/);
assert.throws(() => V.decode({ schema_version: "9" }, "invalid legacy schema"), /schema_version/);
const bundleVersion = V.validateVersion(JSON.parse(fs.readFileSync(path.join(root, "data/version.json"), "utf8")).version);
for (const name of ["settings", "profile", "academic_activities", ...collectionNames]) {
  const source = JSON.parse(fs.readFileSync(path.join(root, `data/${name}.json`), "utf8"));
  const collection = collectionNames.includes(name);
  assert.equal(Array.isArray(source), collection, "Current collections use raw arrays.");
  if (!collection) {
    assert.ok(!Object.hasOwn(source, "version") && !Object.hasOwn(source, "schema_version"), "Current payloads use the central version marker.");
  }
  const decoded = V.decode(source, `${name}.json`, { collection, bundleVersion });
  assert.equal(decoded.version, bundleVersion);
  assert.equal(decoded.legacy, false);
}
const source = fs.readFileSync(path.join(root, "app/assets/js/data-version.js"), "utf8");
assert.equal(fs.readFileSync(path.join(root, "editor/cmd/editor/frontend/dist/data-version.js"), "utf8"), source);
for (const name of ["index.html", "cv.html", "fork-guide.html"]) {
  const html = fs.readFileSync(path.join(root, name), "utf8");
  assert.ok(html.indexOf('src="app/assets/js/data-version.js"') >= 0);
  assert.ok(html.indexOf('src="app/assets/js/data-version.js"') < html.indexOf('src="app/assets/js/data.js"'));
}
const editorHTML = fs.readFileSync(path.join(root, "editor/cmd/editor/frontend/dist/index.html"), "utf8");
assert.ok(editorHTML.indexOf('src="data-version.js"') < editorHTML.indexOf('src="app.js"'));
console.log("Data versions: Semantic Version compatibility, future version preservation, legacy readers, collection envelopes and shared browser/editor codec checked.");
