(function (root, factory) {
  "use strict";
  const api = factory();
  if (typeof module === "object" && module.exports) module.exports = api;
  else root.SiteDataVersion = api;
}(typeof window !== "undefined" ? window : globalThis, function () {
  "use strict";

  const CURRENT_VERSION = "1.0.0";
  const SEMVER = /^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$/;
  const UINT64_MAX = "18446744073709551615";

  function validateVersion(value, context = "JSON document") {
    const match = typeof value === "string" ? value.match(SEMVER) : null;
    if (!match || match[0] !== value || match.slice(1).some((part) => part.length > UINT64_MAX.length
      || (part.length === UINT64_MAX.length && part > UINT64_MAX))) {
      throw new Error(`${context}: version must use stable major.minor.patch numbers within the uint64 range, such as ${CURRENT_VERSION}.`);
    }
    if (match[1] !== "1") throw new Error(`${context}: unsupported version '${value}'; supported major version is 1.`);
    return value;
  }

  function decode(value, context = "JSON document", { collection = false, bundleVersion = null } = {}) {
    const sharedVersion = bundleVersion === null ? null : validateVersion(bundleVersion, "data/version.json");
    if (collection && Array.isArray(value)) {
      return { version: sharedVersion || CURRENT_VERSION, data: value, legacy: !sharedVersion, legacySchemaVersion: null };
    }
    if (!value || typeof value !== "object" || Array.isArray(value)) {
      throw new Error(`${context}: must be a JSON object${collection ? " with an items array" : ""}.`);
    }
    const hasVersion = Object.hasOwn(value, "version");
    const legacyDocument = !hasVersion || value.version === 1;
    const legacy = !sharedVersion && legacyDocument;
    if (collection && !hasVersion) throw new Error(`${context}: version is required for collection documents.`);
    const documentVersion = legacyDocument ? CURRENT_VERSION : validateVersion(value.version, context);
    const version = sharedVersion || documentVersion;
    let legacySchemaVersion = null;
    if (legacy && Object.hasOwn(value, "schema_version")) {
      if (!Number.isInteger(value.schema_version) || value.schema_version < 0) {
        throw new Error(`${context}: legacy schema_version must be a non-negative integer.`);
      }
      legacySchemaVersion = value.schema_version;
    } else if (legacy) {
      legacySchemaVersion = 0;
    }
    if (collection) {
      if (!Array.isArray(value.items)) throw new Error(`${context}: items must be an array.`);
      return { version, data: value.items, legacy, legacySchemaVersion };
    }
    const data = { ...value, version };
    delete data.schema_version;
    return { version, data, legacy, legacySchemaVersion };
  }

  return Object.freeze({ CURRENT_VERSION, validateVersion, decode });
}));
