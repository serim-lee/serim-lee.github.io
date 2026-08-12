
(function () {
  "use strict";

  const FILES = {
    version: "data/version.json",
    settings: "data/settings.json",
    profile: "data/profile.json",
    people: "data/people.json",
    projects: "data/projects.json",
    software: "data/software.json",
    publications: "data/publications.json",
    awards: "data/awards.json",
    board: "data/board.json",
    academicActivities: "data/academic_activities.json"
  };
  const SOFTWARE_STAGES = new Set(["release", "preview", "development"]);
  const WEB_PAGES = Object.freeze(["publications", "projects", "software", "board"]);
  const PATENT_TYPES = Object.freeze(["domestic", "international"]);
  const PATENT_STAGES = Object.freeze(["filed", "published", "granted"]);
  const CANONICAL_ID_PATTERN = /^[a-z0-9]+(?:-[a-z0-9]+)*$/;
  const TOPIC_COLORS = ["#315f49", "#3d6f86", "#80643b", "#775778", "#4f6e31", "#8a5148"];
  const CURRENCY_CODE_PATTERN = /^[A-Z]{3}$/;
  const EXCHANGE_RATE_API = "https://api.frankfurter.dev/v2/rate";
  const exchangeRateRequests = new Map();
  const MAIN_PAGE_SECTIONS = [
    "experience",
    "education",
    "scholarships",
    "certifications",
    "awards",
    "academic_activities",
    "patents",
    "teaching",
    "skills"
  ];
  const PRE_PATENT_MAIN_PAGE_SECTIONS = MAIN_PAGE_SECTIONS.filter((section) => section !== "patents");
  const LEGACY_MAIN_PAGE_SECTIONS = PRE_PATENT_MAIN_PAGE_SECTIONS.filter((section) => section !== "academic_activities");
  const ACADEMIC_ACTIVITY_FIELDS = Object.freeze({
    reviews: ["journal_en", "journal_ko", "completed_dates"],
    invited_talks: ["event_en", "event_ko", "topic_en", "topic_ko", "date"],
    conference_service: ["conference_en", "conference_ko", "role_en", "role_ko", "start_date", "end_date"],
    professional_service: ["organization_en", "organization_ko", "role_en", "role_ko", "start_date", "end_date"],
    editorial_service: ["journal_en", "journal_ko", "role_en", "role_ko", "start_date", "end_date"]
  });
  const ACADEMIC_ACTIVITY_CATEGORIES = Object.freeze([
    "editorial_service",
    "professional_service",
    "conference_service",
    "invited_talks",
    "reviews"
  ]);
  const CV_SECTIONS = [
    "experience",
    "education",
    "projects",
    "publications",
    "software",
    "awards",
    "academic_activities",
    "patents",
    "teaching",
    "scholarships",
    "certifications",
    "skills"
  ];
  const PROFILE_CV_SECTIONS = [
    "experience",
    "education",
    "scholarships",
    "teaching",
    "certifications",
    "skills",
    "patents"
  ];

  function applyColorTheme(settings) {
    const theme = window.SiteThemes.applyWebsite(settings);
    const meta = document.querySelector('meta[name="theme-color"]');
    if (meta) {
      meta.content = theme.colors.bg;
    }
    return theme.id;
  }

  async function fetchJSON(url) {
    const response = await fetch(url, { cache: "no-cache" });
    if (!response.ok) {
      throw new Error(`Failed to load ${url}: ${response.status}`);
    }
    return response.json();
  }

  async function loadVersion() {
    const response = await fetch(FILES.version, { cache: "no-cache" });
    if (response.status === 404) return null;
    if (!response.ok) throw new Error(`Failed to load ${FILES.version}: ${response.status}`);
    const document = await response.json();
    return window.SiteDataVersion.validateVersion(document?.version, FILES.version);
  }

  function isValidISODate(value) {
    const text = String(value || "").trim();
    if (!/^\d{4}-\d{2}-\d{2}$/.test(text)) return false;
    const parsed = new Date(`${text}T00:00:00Z`);
    return !Number.isNaN(parsed.getTime()) && parsed.toISOString().slice(0, 10) === text;
  }

  function requireArray(value, context) {
    if (!Array.isArray(value)) {
      throw new Error(`${context}: must be an array.`);
    }
    return value;
  }

  function requireRecord(value, context) {
    if (!value || typeof value !== "object" || Array.isArray(value)) {
      throw new Error(`${context}: must be an object.`);
    }
    return value;
  }

  function validateScholarships(value) {
    requireArray(value, "profile.json scholarships").forEach((row, index) => {
      const context = `profile.json scholarships item ${index + 1}`;
      const item = requireRecord(row, context);
      if (Object.prototype.hasOwnProperty.call(item, "details")) {
        throw new Error(`${context}: details is obsolete; use amount instead.`);
      }
      const amount = requireRecord(item.amount, `${context} amount`);
      if (typeof amount.value !== "number" || !Number.isFinite(amount.value) || amount.value <= 0) {
        throw new Error(`${context} amount: value must be a positive finite number.`);
      }
      if (typeof amount.currency !== "string" || !CURRENCY_CODE_PATTERN.test(amount.currency)) {
        throw new Error(`${context} amount: currency must be a three-letter uppercase code.`);
      }
    });
  }

  function scholarshipTargetCurrency(language) {
    return language === "ko" ? "KRW" : "USD";
  }

  function roundedCurrencyValue(value) {
    const rounded = Math.round(Number(value) / 10) * 10;
    return Object.is(rounded, -0) ? 0 : rounded;
  }

  function formatCurrency(value, currency, language) {
    return new Intl.NumberFormat(language === "ko" ? "ko-KR" : "en-US", {
      style: "currency",
      currency,
      currencyDisplay: "code",
      minimumFractionDigits: 0,
      maximumFractionDigits: 0
    }).format(roundedCurrencyValue(value));
  }

  function formatScholarshipAmount(amount, language = "en") {
    return formatCurrency(amount.value, amount.currency, language);
  }

  function exchangeRate(base, quote) {
    const key = `${base}/${quote}`;
    if (!exchangeRateRequests.has(key)) {
      const request = fetchJSON(`${EXCHANGE_RATE_API}/${encodeURIComponent(base)}/${encodeURIComponent(quote)}`)
        .then((response) => {
          const rate = Number(response?.rate);
          if (!Number.isFinite(rate) || rate <= 0 || !isValidISODate(response?.date)) {
            throw new Error(`Invalid exchange rate for ${key}.`);
          }
          return { base, quote, rate, date: response.date };
        });
      exchangeRateRequests.set(key, request);
    }
    return exchangeRateRequests.get(key);
  }

  async function resolveScholarshipAmountDetails(amount, language = "en") {
    const fallback = { text: formatScholarshipAmount(amount, language), conversion: null };
    const targetCurrency = scholarshipTargetCurrency(language);
    if (amount.currency === targetCurrency) return fallback;
    try {
      const conversion = await exchangeRate(amount.currency, targetCurrency);
      return {
        text: formatCurrency(amount.value * conversion.rate, targetCurrency, language),
        conversion
      };
    } catch (_error) {
      return fallback;
    }
  }

  async function resolveScholarshipAmount(amount, language = "en") {
    return (await resolveScholarshipAmountDetails(amount, language)).text;
  }

  function formatScholarshipConversionNote(conversions) {
    const unique = new Map();
    conversions.filter(Boolean).forEach((conversion) => {
      unique.set(`${conversion.base}/${conversion.quote}/${conversion.date}`, conversion);
    });
    const number = new Intl.NumberFormat("en-US", { maximumFractionDigits: 2 });
    return Array.from(unique.values(), ({ base, quote, rate, date }) =>
      `Converted to ${quote} · ${date} · 1 ${quote} = ${number.format(1 / rate)} ${base}`
    ).join("; ");
  }

  function requireOptionalCVVisibility(value, context) {
    if (Object.prototype.hasOwnProperty.call(value, "visible_in_CV") && typeof value.visible_in_CV !== "boolean") {
      throw new Error(`${context}: visible_in_CV must be a boolean when present.`);
    }
  }

  function visibleInCV(value) {
    return value?.visible_in_CV !== false;
  }

  function normalizeWebPageSettings(settings) {
    if (!Object.hasOwn(settings, "hidden_web_pages")) settings.hidden_web_pages = [];
    const hidden = settings.hidden_web_pages;
    if (!Array.isArray(hidden) || hidden.some((page) => !WEB_PAGES.includes(page)) || new Set(hidden).size !== hidden.length) {
      throw new Error("settings.json: hidden_web_pages must be an array of unique page names: " + WEB_PAGES.join(", "));
    }
  }

  function webPageVisible(settings, page) {
    return WEB_PAGES.includes(page) && !settings?.hidden_web_pages?.includes(page);
  }

  function normalizeCVSectionSettings(settings, legacySchemaVersion) {
    const hasVisible = Object.prototype.hasOwnProperty.call(settings, "cv_sections");
    const hasHidden = Object.prototype.hasOwnProperty.call(settings, "hidden_cv_sections");
    if (!hasVisible && !hasHidden && legacySchemaVersion !== null && legacySchemaVersion <= 4) {
      settings.cv_sections = CV_SECTIONS.filter((section) => section !== "patents");
      settings.hidden_cv_sections = ["patents"];
      return;
    }

    normalizePatentSectionSettings(settings, "cv_sections", "hidden_cv_sections", CV_SECTIONS, legacySchemaVersion);
    const visible = settings.cv_sections;
    const hidden = settings.hidden_cv_sections;
    const configured = Array.isArray(visible) && Array.isArray(hidden) ? [...visible, ...hidden] : [];
    const valid = Array.isArray(visible)
      && Array.isArray(hidden)
      && configured.length === CV_SECTIONS.length
      && configured.every((section) => CV_SECTIONS.includes(section))
      && new Set(configured).size === CV_SECTIONS.length;
    if (!valid) {
      throw new Error(
        "settings.json: cv_sections and hidden_cv_sections must together contain each supported section exactly once: "
        + CV_SECTIONS.join(", ")
      );
    }
  }

  function normalizeMainPageSectionSettings(settings, legacySchemaVersion) {
    const visible = settings.main_page_sections;
    const hidden = settings.hidden_main_page_sections;
    const configured = Array.isArray(visible) && Array.isArray(hidden) ? [...visible, ...hidden] : [];
    const isLegacyConfiguration = legacySchemaVersion !== null && legacySchemaVersion <= 5
      && configured.length === LEGACY_MAIN_PAGE_SECTIONS.length
      && configured.every((section) => LEGACY_MAIN_PAGE_SECTIONS.includes(section))
      && new Set(configured).size === LEGACY_MAIN_PAGE_SECTIONS.length;
    if (isLegacyConfiguration) {
      settings.hidden_main_page_sections = [...hidden, "academic_activities"];
    }
    normalizePatentSectionSettings(settings, "main_page_sections", "hidden_main_page_sections", MAIN_PAGE_SECTIONS, legacySchemaVersion);
  }

  function normalizePatentSectionSettings(settings, visibleKey, hiddenKey, supported, legacySchemaVersion) {
    const visible = settings[visibleKey];
    const hidden = settings[hiddenKey];
    if (!Array.isArray(visible) || !Array.isArray(hidden) || legacySchemaVersion === null || legacySchemaVersion > 8) return;
    const previous = supported.filter((section) => section !== "patents");
    const configured = [...visible, ...hidden];
    if (configured.length === previous.length
      && configured.every((section) => previous.includes(section))
      && new Set(configured).size === previous.length) {
      settings[hiddenKey] = [...hidden, "patents"];
    }
  }

  function validatePatents(value, peopleIds) {
    const textFields = [
      "title_en", "title_ko", "applicant_en", "applicant_ko", "url",
      "application_number", "application_date", "publication_number", "publication_date", "registration_number", "registration_date"
    ];
    requireArray(value, "profile.json patents").forEach((row, index) => {
      const context = `profile.json patents item ${index + 1}`;
      const item = requireRecord(row, context);
      requireOptionalCVVisibility(item, context);
      textFields.forEach((field) => {
        if (Object.hasOwn(item, field) && typeof item[field] !== "string") {
          throw new Error(`${context}: ${field} must be a string when present.`);
        }
      });
      if (![item.title_en, item.title_ko].some((title) => typeof title === "string" && title.trim())) {
        throw new Error(`${context}: at least one localized title is required.`);
      }
      if (!PATENT_TYPES.includes(item.type)) {
        throw new Error(`${context}: type must be ${PATENT_TYPES.join(", ")}.`);
      }
      if (!Object.hasOwn(item, "inventor_ids")) item.inventor_ids = [];
      if (!Array.isArray(item.inventor_ids)) {
        throw new Error(`${context} inventor_ids: must be an array of canonical people IDs.`);
      }
      const seenInventors = new Set();
      item.inventor_ids.forEach((id, inventorIndex) => {
        requireKnownId(id, peopleIds, `${context} inventor_ids[${inventorIndex}]`);
        if (seenInventors.has(id)) {
          throw new Error(`${context} inventor_ids[${inventorIndex}]: duplicate inventor id '${id}'.`);
        }
        seenInventors.add(id);
      });
      if (![item.application_number, item.publication_number, item.registration_number].some((number) => typeof number === "string" && number.trim())) {
        throw new Error(`${context}: at least one application, publication, or registration number is required.`);
      }
      ["application_date", "publication_date", "registration_date"].forEach((field) => {
        if (item[field] && (!/^\d{4}-\d{2}-\d{2}$/.test(item[field]) || item[field].startsWith("0000-") || !isValidISODate(item[field]))) {
          throw new Error(`${context}: ${field} must be a valid YYYY-MM-DD date or an empty string.`);
        }
      });
      ["publication_date", "registration_date"].forEach((field) => {
        if (item.application_date && item[field] && item[field] < item.application_date) {
          throw new Error(`${context}: ${field} must be on or after application_date.`);
        }
      });
      if (item.url && !patentURL(item)) {
        throw new Error(`${context}: url must be an http(s) URL or an empty string.`);
      }
    });
  }

  function validateAcademicActivities(value) {
    const document = requireRecord(value, "academic_activities.json");
    window.SiteDataVersion.validateVersion(document.version, "academic_activities.json");

    Object.entries(ACADEMIC_ACTIVITY_FIELDS).forEach(([category, fields]) => {
      requireArray(document[category], `academic_activities.json ${category}`).forEach((row, index) => {
        const context = `academic_activities.json ${category} item ${index + 1}`;
        const item = requireRecord(row, context);
        requireOptionalCVVisibility(item, context);
        const missingFields = fields.filter((field) => !Object.prototype.hasOwnProperty.call(item, field));
        if (missingFields.length) throw new Error(`${context}: missing field(s): ${missingFields.join(", ")}.`);

        fields.filter((field) => field !== "completed_dates").forEach((field) => {
          if (typeof item[field] !== "string") throw new Error(`${context}: ${field} must be a string.`);
        });
        if (category === "reviews") {
          const completedDates = requireArray(item.completed_dates, `${context} completed_dates`);
          if (!completedDates.length) {
            throw new Error(`${context}: completed_dates must contain at least one date.`);
          }
          completedDates.forEach((date, dateIndex) => {
            if (typeof date !== "string" || date !== date.trim() || !isValidISODate(date)) {
              throw new Error(`${context}: completed_dates[${dateIndex}] must be a valid YYYY-MM-DD date.`);
            }
          });
          if (new Set(completedDates).size !== completedDates.length) {
            throw new Error(`${context}: completed_dates must not contain duplicate dates.`);
          }
          if (![item.journal_en, item.journal_ko].some((text) => text.trim())) {
            throw new Error(`${context}: at least one localized journal name is required.`);
          }
          return;
        }

        const localizedPairs = category === "invited_talks"
          ? [["event_en", "event_ko"], ["topic_en", "topic_ko"]]
          : category === "conference_service"
            ? [["conference_en", "conference_ko"], ["role_en", "role_ko"]]
            : category === "professional_service"
              ? [["organization_en", "organization_ko"], ["role_en", "role_ko"]]
              : [["journal_en", "journal_ko"], ["role_en", "role_ko"]];
        localizedPairs.forEach((pair) => {
          if (!pair.some((field) => item[field].trim())) {
            throw new Error(`${context}: at least one localized value is required for ${pair.join("/")}.`);
          }
        });
        if (category === "invited_talks") {
          if (!isValidISODate(item.date)) throw new Error(`${context}: date must be a valid YYYY-MM-DD date.`);
          return;
        }
        if (!isValidISODate(item.start_date)) {
          throw new Error(`${context}: start_date must be a valid YYYY-MM-DD date.`);
        }
        if (item.end_date && !isValidISODate(item.end_date)) {
          throw new Error(`${context}: end_date must be blank or a valid YYYY-MM-DD date.`);
        }
        if (item.end_date && item.start_date > item.end_date) {
          throw new Error(`${context}: start_date must not be later than end_date.`);
        }
      });
    });
    return document;
  }

  function requireCanonicalId(value, context) {
    if (typeof value !== "string" || !CANONICAL_ID_PATTERN.test(value)) {
      throw new Error(
        `${context}: id must be a lowercase kebab-case string matching `
        + "/^[a-z0-9]+(?:-[a-z0-9]+)*$/. Reserved sentinel values are not valid data IDs."
      );
    }
    return value;
  }

  function collectUniqueIds(rows, context) {
    const ids = new Set();
    requireArray(rows, context).forEach((row, index) => {
      const itemContext = `${context} item ${index + 1}`;
      requireRecord(row, itemContext);
      const id = requireCanonicalId(row.id, `${itemContext} id`);
      if (ids.has(id)) {
        throw new Error(`${itemContext}: duplicate id '${id}' in ${context}; every id must be unique.`);
      }
      ids.add(id);
    });
    return ids;
  }

  function requireKnownId(value, knownIds, context) {
    const id = requireCanonicalId(value, context);
    if (!knownIds.has(id)) {
      throw new Error(`${context}: unknown id '${id}'; add it to the referenced JSON collection or correct the reference.`);
    }
    return id;
  }

  function validateTopicColors(rows, context) {
    rows.forEach((row, index) => {
      if (Object.prototype.hasOwnProperty.call(row, "color")
        && (typeof row.color !== "string" || !/^#[0-9a-fA-F]{6}$/.test(row.color))) {
        throw new Error(`${context} item ${index + 1}: color must use #RRGGBB when present.`);
      }
    });
  }

  async function loadAll() {
    const [bundleVersion, ...sources] = await Promise.all([
      loadVersion(),
      fetchJSON(FILES.settings),
      fetchJSON(FILES.profile),
      fetchJSON(FILES.people),
      fetchJSON(FILES.projects),
      fetchJSON(FILES.software),
      fetchJSON(FILES.publications),
      fetchJSON(FILES.awards),
      fetchJSON(FILES.board),
      fetchJSON(FILES.academicActivities)
    ]);
    const contexts = ["settings", "profile", "people", "projects", "software", "publications", "awards", "board", "academic_activities"];
    const documents = sources.map((source, index) => window.SiteDataVersion.decode(source, `${contexts[index]}.json`, {
      collection: index >= 2 && index <= 7, bundleVersion
    }));
    const [settings, profile, people, projects, software, publications, awards, board, academicActivities] = documents.map((document) => document.data);
    const legacySettingsSchema = !bundleVersion && documents[0].legacy && Number.isInteger(sources[0].schema_version)
      ? documents[0].legacySchemaVersion : null;
    if (!bundleVersion && documents[8].legacy && Object.hasOwn(sources[8], "schema_version") && documents[8].legacySchemaVersion !== 1) {
      throw new Error("academic_activities.json: legacy schema_version must be 1.");
    }

    requireRecord(settings, "settings.json");
    requireRecord(profile, "profile.json");
    if (!Object.hasOwn(profile, "patents")) profile.patents = [];
    requireArray(projects, "projects.json");
    requireArray(publications, "publications.json");
    requireArray(board, "board.json");
    validateAcademicActivities(academicActivities);
    if (legacySettingsSchema !== null) normalizeTopicSettings(settings);
    else {
      requireArray(settings.topics, "settings.json topics");
      requireRecord(settings.topic_fallback, "settings.json topic_fallback");
    }
    normalizeMainPageSectionSettings(settings, legacySettingsSchema);
    normalizeCVSectionSettings(settings, legacySettingsSchema);
    normalizeWebPageSettings(settings);
    PROFILE_CV_SECTIONS.forEach((sectionName) => {
      requireArray(profile[sectionName], `profile.json ${sectionName}`).forEach((item, index) => {
        const context = `profile.json ${sectionName} item ${index + 1}`;
        requireRecord(item, context);
        requireOptionalCVVisibility(item, context);
        if (sectionName === "certifications") validateDocumentReference(item, context);
        if (sectionName === "education") {
          validateEducationGPA(item, context);
          if (Object.hasOwn(item, "thesis_url") &&
            (typeof item.thesis_url !== "string" || (item.thesis_url && !educationThesisURL(item.thesis_url)))) {
            throw new Error(`${context}: thesis_url must be an http(s) URL or an empty string.`);
          }
        }
      });
    });
    validateScholarships(profile.scholarships);
    const peopleIds = collectUniqueIds(people, "people.json");
    validatePatents(profile.patents, peopleIds);
    const awardIds = collectUniqueIds(awards, "awards.json");
    collectUniqueIds(software, "software.json");
    const projectThemeIds = collectUniqueIds(settings.topics, "settings.json topics");
    const publicationTopicIds = projectThemeIds;
    validateTopicColors(settings.topics, "settings.json topics");

    const peopleById = new Map(people.map((person) => [person.id, person]));
    const resolvedPatents = profile.patents.map((item) => ({
      ...item,
      inventors: item.inventor_ids.map((id) => peopleById.get(id))
    }));
    awards.forEach((award, index) => {
      const context = `awards.json item ${index + 1}`;
      requireRecord(award, context);
      requireOptionalCVVisibility(award, context);
      validateDocumentReference(award, context);
      ["date", "title_en", "title_ko", "organization_en", "organization_ko"].forEach((field) => {
        if (typeof award[field] !== "string") {
          throw new Error(`${context}: ${field} must be a string.`);
        }
      });
      if (!/^\d{4}-\d{2}-\d{2}$/.test(award.date) || !isValidISODate(award.date)) {
        throw new Error(`${context}: date must be a valid YYYY-MM-DD date exactly.`);
      }
      if (![award.title_en, award.title_ko].some((value) => value.trim())) {
        throw new Error(`${context}: at least one localized title is required.`);
      }
      if (![award.organization_en, award.organization_ko].some((value) => value.trim())) {
        throw new Error(`${context}: at least one localized organization is required.`);
      }
    });
    projects.forEach((project, index) => {
      const context = `projects.json item ${index + 1}`;
      requireRecord(project, context);
      requireOptionalCVVisibility(project, context);
      if (typeof project.theme !== "string" || !project.theme) {
        throw new Error(`${context} theme: must be a non-empty canonical project theme id.`);
      }
      requireKnownId(project.theme, projectThemeIds, `${context} theme`);
      if ("summary" in project || "summary_en" in project || "summary_ko" in project) {
        throw new Error(`projects.json item ${index + 1}: use notes_en and notes_kr instead of summary.`);
      }
      ["notes_en", "notes_kr"].forEach((field) => {
        if (
          !Array.isArray(project[field])
          || !project[field].length
          || project[field].some((note) => typeof note !== "string" || !note.trim())
        ) {
          throw new Error(`projects.json item ${index + 1}: ${field} must be a non-empty array of non-empty strings.`);
        }
      });
      (project.media || []).forEach((mediaItem, mediaIndex) => {
        if ("caption_kr" in mediaItem || "caption-en" in mediaItem) {
          throw new Error(`projects.json item ${index + 1} media ${mediaIndex + 1}: use caption_en and caption_ko.`);
        }
        const captionEn = String(mediaItem.caption_en || "").trim();
        const captionKo = String(mediaItem.caption_ko || "").trim();
        if (Boolean(captionEn) !== Boolean(captionKo)) {
          throw new Error(`projects.json item ${index + 1} media ${mediaIndex + 1}: caption_en and caption_ko must both be filled or both be blank.`);
        }
      });
    });
    software.forEach((item, index) => {
      const context = `software.json item ${index + 1}`;
      requireRecord(item, context);
      requireOptionalCVVisibility(item, context);
      if (!SOFTWARE_STAGES.has(item.stage)) {
        throw new Error(`software.json item ${index + 1}: stage must be release, preview, or development.`);
      }
      if ("summary" in item || "summary_en" in item || "summary_ko" in item) {
        throw new Error(`software.json item ${index + 1}: use notes_en and notes_kr instead of summary.`);
      }
      ["notes_en", "notes_kr"].forEach((field) => {
        if (
          !Array.isArray(item[field])
          || !item[field].length
          || item[field].some((note) => typeof note !== "string" || !note.trim())
        ) {
          throw new Error(`software.json item ${index + 1}: ${field} must be a non-empty array of non-empty strings.`);
        }
      });
    });
    board.forEach((item, index) => {
      const context = `board.json item ${index + 1}`;
      requireRecord(item, context);
      if ("date" in item) {
        throw new Error(`${context}: use start_date and end_date instead of date.`);
      }
      const startDate = String(item.start_date || "").trim();
      const endDate = String(item.end_date || "").trim();
      if (!isValidISODate(startDate)) {
        throw new Error(`${context}: start_date must be a valid YYYY-MM-DD date.`);
      }
      if (endDate && !isValidISODate(endDate)) {
        throw new Error(`${context}: end_date must be blank or a valid YYYY-MM-DD date.`);
      }
      if (endDate && startDate > endDate) {
        throw new Error(`${context}: start_date must not be later than end_date.`);
      }
      ["content_en", "content_ko"].forEach((field) => {
        if (typeof item[field] !== "string") {
          throw new Error(`${context}: ${field} must be a string.`);
        }
      });
    });
    const resolvedPublications = publications.map((publication, publicationIndex) => {
      const context = `publications.json item ${publicationIndex + 1}`;
      requireRecord(publication, context);
      requireOptionalCVVisibility(publication, context);
      if (["venue", "venue_en", "venue_kr"].some((field) => Object.hasOwn(publication, field))) {
        throw new Error(`${context}: use journal_id / conference_id and bibliography fields instead of a manual venue.`);
      }
      ["journal_id", "conference_id"].forEach((field) => {
        const kind = field === "journal_id" ? "journals" : "conferences";
        const suffix = kind === "journals" ? "-journal" : "-conference";
        if (publication[field] && (!publication.publication_type.endsWith(suffix)
          || !(settings[kind] || []).some((record) => record.id === publication[field]))) {
          throw new Error(`${context}: unknown or mismatched ${field}.`);
        }
      });

      if (!Array.isArray(publication.author_ids) || !publication.author_ids.length) {
        throw new Error(`${context} author_ids: must be a non-empty array of canonical people IDs.`);
      }
      const authorIds = publication.author_ids.map((authorId, authorIndex) => (
        requireKnownId(authorId, peopleIds, `${context} author_ids[${authorIndex}]`)
      ));
      const seenAuthorIds = new Set();
      authorIds.forEach((authorId, authorIndex) => {
        if (seenAuthorIds.has(authorId)) {
          throw new Error(`${context} author_ids[${authorIndex}]: duplicate author id '${authorId}'.`);
        }
        seenAuthorIds.add(authorId);
      });

      if (typeof publication.award_id !== "string") {
        throw new Error(`${context} award_id: must be an empty string or a canonical awards.json id.`);
      }
      if (publication.award_id) {
        requireKnownId(publication.award_id, awardIds, `${context} award_id`);
      }

      if (typeof publication.topic !== "string") {
        throw new Error(`${context} topic: must be an empty string or a canonical settings.json publication topic id.`);
      }
      if (publication.topic) {
        requireKnownId(publication.topic, publicationTopicIds, `${context} topic`);
      }

      return {
        ...window.PublicationBibliography.resolvePublication(publication, settings),
        publicationIndex,
        authors: authorIds.map((id) => peopleById.get(id))
      };
    });

    return {
      version: bundleVersion || window.SiteDataVersion.CURRENT_VERSION,
      settings,
      profile: { ...profile, patents: resolvedPatents },
      people,
      projects,
      software,
      publications: resolvedPublications,
      awards,
      board,
      academicActivities
    };
  }

  function splitList(value) {
    if (Array.isArray(value)) {
      return value.map((part) => String(part).trim()).filter(Boolean);
    }
    return String(value || "")
      .split(";")
      .map((part) => part.trim())
      .filter(Boolean);
  }

  function truthy(value) {
    return ["true", "1", "yes", "y"].includes(String(value || "").toLowerCase());
  }

  function escapeHTML(value) {
    return String(value ?? "")
      .replaceAll("&", "&amp;")
      .replaceAll("<", "&lt;")
      .replaceAll(">", "&gt;")
      .replaceAll('"', "&quot;")
      .replaceAll("'", "&#039;");
  }

  function periodHTML(value) {
    const text = String(value ?? "");
    const annotationStart = text.search(/[（(]/u);
    const parts = annotationStart > 0
      ? [text.slice(0, annotationStart), text.slice(annotationStart)]
      : [text];
    return parts.map((part) => {
      const [, before, content, after] = part.match(/^(\s*)([\s\S]*?)(\s*)$/u);
      return `${escapeHTML(before)}<span class="period-part">${escapeHTML(content)}</span>${escapeHTML(after)}`;
    }).join("");
  }

  function pick(record, key, language) {
    if (!record) return "";
    return [record[`${key}_${language}`], record[key], record[`${key}_en`], record[`${key}_ko`]]
      .find((value) => (Array.isArray(value) ? value.length > 0 : Boolean(value))) || "";
  }

  function publicationTitle(item, language = "en") {
    return pick(item, "title", language);
  }

  function patentStage(item) {
    const prefix = ["registration", "publication", "application"]
      .find((prefix) => String(item[`${prefix}_number`] || "").trim());
    return { registration: "granted", publication: "published", application: "filed" }[prefix] || "";
  }

  function patentStatus(item, language = "en") {
    const labels = {
      filed: { en: "Filed", ko: "출원" },
      published: { en: "Published", ko: "공개" },
      granted: { en: "Granted", ko: "등록" }
    };
    return labels[patentStage(item)]?.[language === "ko" ? "ko" : "en"] || "";
  }

  function patentType(item, language = "en") {
    const labels = {
      domestic: { en: "Domestic Patents", ko: "국내특허" },
      international: { en: "International Patents", ko: "국제특허" }
    };
    return labels[item.type]?.[language === "ko" ? "ko" : "en"] || "";
  }

  function patentPeopleHTML(item, language = "en") {
    const inventors = displayAuthors(item.inventors, language, false);
    const applicant = escapeHTML(patentText(item, "applicant", language));
    const filing = applicant ? (language === "ko" ? `(${applicant} 출원)` : `(filed by ${applicant})`) : "";
    return [inventors, filing].filter(Boolean).join(" ");
  }

  function patentGroups(items, language = "en") {
    return PATENT_TYPES.flatMap((type) => PATENT_STAGES.map((status) => {
      const members = items.filter((item) => item.type === type && patentStage(item) === status);
      return {
        type,
        status,
        title: members.length ? `${patentType({ type }, language)} [${patentStatus(members[0], language)}]` : "",
        items: members
      };
    })).filter((group) => group.items.length);
  }

  function patentDate(item) {
    return item.registration_date || item.publication_date || item.application_date || "";
  }

  function patentText(item, key, language) {
    return [item[`${key}_${language}`], item[`${key}_en`], item[`${key}_ko`]]
      .map((value) => String(value || "").trim()).find(Boolean) || "";
  }

  function patentTitle(item, language = "en") {
    return patentText(item, "title", language);
  }

  function patentDetails(item, language = "en") {
    const korean = language === "ko";
    const details = [];
    [
      ["application", korean ? "출원번호" : "Application No.", korean ? "출원일" : "Filed"],
      ["publication", korean ? "공개번호" : "Publication No.", korean ? "공개일" : "Published"],
      ["registration", korean ? "등록번호" : "Patent No.", korean ? "등록일" : "Granted"]
    ].forEach(([prefix, numberLabel, dateLabel]) => {
      const values = [];
      const number = String(item[`${prefix}_number`] || "").trim();
      if (number) values.push(`${numberLabel} ${number}`);
      if (prefix !== "application" && item[`${prefix}_date`]) values.push(`${dateLabel} ${item[`${prefix}_date`]}`);
      if (values.length) details.push(values.join(" · "));
    });
    return details;
  }

  function patentURL(item) {
    if (typeof item.url !== "string" || [...item.url].length > 4096
      || !/^https?:\/\//i.test(item.url) || /[\u0000-\u001f\u007f-\u009f]/.test(item.url)) return "";
    try {
      const url = new URL(item.url);
      return ["http:", "https:"].includes(url.protocol) && url.hostname ? url.href : "";
    } catch {
      return "";
    }
  }

  function publicationVenue(item, language = "en") {
    const koreanVenue = String(item.venue_kr || "").trim();
    return language === "ko" && koreanVenue ? koreanVenue : item.venue_en || "";
  }

  function publicationReference(item, language = "en") {
    return String(item[`reference_${language}`] || item.reference_en || item.reference_ko || "");
  }

  function publicationConferenceInfo(item, language = "en") {
    return String(item[`conference_info_${language}`] || item.conference_info_en || item.conference_info_ko || "");
  }

  function publicationVenueHTML(item, language = "en") {
    const conference = publicationConferenceInfo(item, language);
    return `<span class="publication-reference">${escapeHTML(publicationReference(item, language))}</span>`
      + (conference ? `<span class="publication-conference-info">${escapeHTML(conference)}</span>` : "");
  }

  function personName(person, language = "en") {
    return person?.[`name_${language}`] || person?.name_en || person?.name_ko || person?.id || "";
  }

  function personNotesHTML(person, language = "en") {
    const cleanNotes = (value) => Array.isArray(value)
      ? value.map((note) => String(note).trim()).filter(Boolean)
      : [];
    const preferred = cleanNotes(person?.[`notes_${language}`]);
    const fallbackLanguage = language === "ko" ? "en" : "ko";
    const fallback = cleanNotes(person?.[`notes_${fallbackLanguage}`]);
    const legacy = cleanNotes(person?.notes);
    const notes = preferred.length ? preferred : fallback.length ? fallback : legacy;
    if (!notes.length) return "";
    return `
      <span class="person-notes-tooltip" role="tooltip">
        ${notes.map((note) => `<span class="person-note">${escapeHTML(note)}</span>`).join("")}
      </span>
    `;
  }

  function displayAuthors(people, language = "en", showNotes = true) {
    return (people || []).map((person) => {
      const name = escapeHTML(personName(person, language));
      const notes = showNotes ? personNotesHTML(person, language) : "";
      const label = person.is_self ? `<strong>${name}</strong>` : name;
      return `<span class="publication-person"${notes ? ' tabindex="0"' : ""}>${label}${notes}</span>`;
    }).join(", ");
  }

  function isFirstAuthor(item) {
    return Boolean(item.authors?.[0]?.is_self);
  }

  function publicationTypeLabel(item, language) {
    const map = {
      en: {
        "international-journal": "International journal",
        "domestic-journal": "Domestic journal",
        "international-conference": "International conference",
        "domestic-conference": "Domestic conference"
      },
      ko: {
        "international-journal": "국제학술지",
        "domestic-journal": "국내학술지",
        "international-conference": "국제학술대회",
        "domestic-conference": "국내학술대회"
      }
    };
    return map[language]?.[item.publication_type] || item.publication_type;
  }

  function normalizeTopicSettings(settings) {
    if (!Object.hasOwn(settings, "topics")) {
      const topics = new Map();
      ["project_themes", "publication_topics"].forEach((key) => {
        collectUniqueIds(settings[key], `settings.json ${key}`);
        validateTopicColors(settings[key], `settings.json ${key}`);
        settings[key].forEach((item) => {
          const current = topics.get(item.id);
          topics.set(item.id, current ? { ...item, ...current, ...(current.color ? {} : item.color ? { color: item.color } : {}) } : { ...item });
        });
      });
      settings.topics = [...topics.values()];
      settings.topic_fallback = { ...settings.publication_topic_fallback, ...settings.project_theme_fallback };
    }
    requireArray(settings.topics, "settings.json topics");
    requireRecord(settings.topic_fallback, "settings.json topic_fallback");
    ["project_themes", "publication_topics", "project_theme_fallback", "publication_topic_fallback"].forEach((key) => delete settings[key]);
    return settings;
  }

  function topicColor(topic, index) {
    return topic.color || TOPIC_COLORS[index] || `hsl(${(index * 137.508).toFixed(2)} 44% 38%)`;
  }

  function topicForegroundColor(color) {
    if (typeof color !== "string" || !/^#[0-9a-f]{6}$/i.test(color)) return "#ffffff";
    const channels = color.slice(1).match(/../g).map((channel) => {
      const value = parseInt(channel, 16) / 255;
      return value <= 0.04045 ? value / 12.92 : ((value + 0.055) / 1.055) ** 2.4;
    });
    const luminance = channels[0] * 0.2126 + channels[1] * 0.7152 + channels[2] * 0.0722;
    const blackContrast = (luminance + 0.05) / 0.05;
    const whiteContrast = 1.05 / (luminance + 0.05);
    return blackContrast >= whiteContrast ? "#000000" : "#ffffff";
  }

  function publicationTopicLabel(item, settings, language) {
    const topic = (settings.topics || []).find((candidate) => candidate.id === item.topic);
    return pick(topic || settings.topic_fallback, "label", language);
  }

  function projectThemeLabel(item, settings, language) {
    const theme = (settings.topics || []).find((candidate) => candidate.id === item.theme);
    return pick(theme || settings.topic_fallback, "label", language);
  }

  function formatProjectDate(value) {
    const match = String(value || "").match(/^(\d{4})-(\d{2})-(\d{2})$/);
    return match ? `${match[1]}.${match[2]}.${match[3]}.` : String(value || "");
  }

  function projectPeriod(item) {
    return [formatProjectDate(item.start_date), formatProjectDate(item.end_date)]
      .filter(Boolean)
      .join(" – ");
  }

  function projectNotes(item, language = "en") {
    const primary = language === "ko" ? item.notes_kr : item.notes_en;
    const fallback = language === "ko" ? item.notes_en : item.notes_kr;
    const notes = Array.isArray(primary) && primary.length ? primary : fallback;
    return (Array.isArray(notes) ? notes : [])
      .map((note) => String(note).trim())
      .filter(Boolean);
  }

  function softwareNotes(item, language = "en") {
    return projectNotes(item, language);
  }

  function publicationStatus(item) {
    if (truthy(item.under_review)) return "under review";
    if (truthy(item.in_press)) return "in press";
    return "";
  }

  function publicationStatusRank(item) {
    if (truthy(item.under_review)) return 2;
    if (truthy(item.in_press)) return 1;
    return 0;
  }

  function publicationYear(item) {
    return String(item.date || "").slice(0, 4);
  }

  function normalizeText(value) {
    return String(value || "").toLocaleLowerCase().normalize("NFKD");
  }

  function publicationCitation(item, language = "en") {
    const citationLanguage = item.publication_type?.startsWith("international-") ? "en" : language;
    const authorText = (item.authors || []).map((person) => personName(person, citationLanguage)).join(", ");
    const doiText = item.doi ? ` https://doi.org/${item.doi}` : "";
    const dateLabel = publicationYear(item) || publicationStatus(item) || "n.d.";
    return `${authorText} (${dateLabel}). ${publicationTitle(item, citationLanguage)}. ${publicationVenue(item, citationLanguage)}${doiText}`.replace(/\s+/g, " ").trim();
  }

  function bibtexEscape(value) {
    return String(value || "")
      .replaceAll("\\", "\\textbackslash{}")
      .replaceAll("&", "\\&")
      .replaceAll("%", "\\%")
      .replaceAll("#", "\\#")
      .replaceAll("_", "\\_")
      .replaceAll("{", "\\{")
      .replaceAll("}", "\\}");
  }

  function publicationCitationKey(item) {
    const source = item.doi || item.url || [
      publicationTitle(item, "en"),
      item.date,
      (item.authors || []).map((person) => personName(person, "en")).join("|")
    ].join("|");
    let hash = 2166136261;
    for (let index = 0; index < source.length; index += 1) {
      hash ^= source.charCodeAt(index);
      hash = Math.imul(hash, 16777619);
    }
    const year = publicationYear(item) || "nd";
    return `publication-${year}-${(hash >>> 0).toString(36)}`;
  }

  function bibtexEntry(item) {
    const entryType = item.publication_type.endsWith("-journal") ? "article" : "inproceedings";
    const dateLabel = publicationYear(item) || publicationStatus(item) || "n.d.";
    const fields = [
      `  title = {${bibtexEscape(publicationTitle(item, "en"))}}`,
      `  author = {${(item.authors || []).map((person) => bibtexEscape(personName(person, "en"))).join(" and ")}}`,
      `  year = {${dateLabel}}`
    ];
    if (entryType === "article") {
      fields.push(`  journal = {${bibtexEscape(publicationVenue(item, "en"))}}`);
    } else {
      fields.push(`  booktitle = {${bibtexEscape(publicationVenue(item, "en"))}}`);
    }
    if (item.doi) fields.push(`  doi = {${item.doi}}`);
    if (item.url) fields.push(`  url = {${item.url}}`);
    return `@${entryType}{${publicationCitationKey(item)},\n${fields.join(",\n")}\n}`;
  }

  function download(filename, content, mimeType) {
    const blob = new Blob([content], { type: mimeType || "text/plain;charset=utf-8" });
    const url = URL.createObjectURL(blob);
    const anchor = document.createElement("a");
    anchor.href = url;
    anchor.download = filename;
    document.body.appendChild(anchor);
    anchor.click();
    anchor.remove();
    window.setTimeout(() => URL.revokeObjectURL(url), 1000);
  }

  function downloadPublicationBib(item) {
    download(`${publicationCitationKey(item)}.bib`, `${bibtexEntry(item)}\n`, "application/x-bibtex;charset=utf-8");
  }

  function safeExternalURL(value) {
    try {
      const url = new URL(value);
      return ["http:", "https:", "mailto:"].includes(url.protocol) ? url.href : "";
    } catch {
      return "";
    }
  }

  function softwareLinkInfo(value, language = "en") {
    const source = typeof value === "object" && value ? value.url : value;
    let url;
    try {
      url = new URL(String(source || "").trim());
    } catch {
      return null;
    }
    if (!['http:', 'https:'].includes(url.protocol)) return null;

    const host = url.hostname.toLowerCase().replace(/^www\./, "");
    const path = url.pathname.toLowerCase();
    const matches = (domain) => host === domain || host.endsWith(`.${domain}`);
    let platform = "webpage";
    let labelEn = "Webpage";
    let labelKo = "웹페이지";
    let icon = "webpage";

    if (matches("github.com")) {
      [platform, labelEn, labelKo, icon] = ["github", "GitHub", "GitHub", "github"];
    } else if (matches("gitlab.com")) {
      [platform, labelEn, labelKo, icon] = ["gitlab", "GitLab", "GitLab", "repository"];
    } else if (matches("bitbucket.org")) {
      [platform, labelEn, labelKo, icon] = ["bitbucket", "Bitbucket", "Bitbucket", "repository"];
    } else if (matches("codeberg.org")) {
      [platform, labelEn, labelKo, icon] = ["codeberg", "Codeberg", "Codeberg", "repository"];
    } else if (matches("sourceforge.net")) {
      [platform, labelEn, labelKo, icon] = ["sourceforge", "SourceForge", "SourceForge", "repository"];
    } else if (matches("food4rhino.com")) {
      [platform, labelEn, labelKo, icon] = ["food4rhino", "food4Rhino", "food4Rhino", "food4rhino"];
    } else if (matches("pypi.org")) {
      [platform, labelEn, labelKo, icon] = ["pypi", "PyPI", "PyPI", "package"];
    } else if (matches("npmjs.com")) {
      [platform, labelEn, labelKo, icon] = ["npm", "npm", "npm", "package"];
    } else if (matches("jsr.io")) {
      [platform, labelEn, labelKo, icon] = ["jsr", "JSR", "JSR", "package"];
    } else if (matches("crates.io")) {
      [platform, labelEn, labelKo, icon] = ["crates", "crates.io", "crates.io", "package"];
    } else if (matches("nuget.org")) {
      [platform, labelEn, labelKo, icon] = ["nuget", "NuGet", "NuGet", "package"];
    } else if (["central.sonatype.com", "search.maven.org", "repo1.maven.org", "mvnrepository.com"].some(matches)) {
      [platform, labelEn, labelKo, icon] = ["maven", "Maven Central", "Maven Central", "package"];
    } else if (matches("hub.docker.com")) {
      [platform, labelEn, labelKo, icon] = ["docker", "Docker Hub", "Docker Hub", "container"];
    } else if (matches("anaconda.org")) {
      [platform, labelEn, labelKo, icon] = ["anaconda", "Anaconda", "Anaconda", "package"];
    } else if (matches("rubygems.org")) {
      [platform, labelEn, labelKo, icon] = ["rubygems", "RubyGems", "RubyGems", "package"];
    } else if (matches("packagist.org")) {
      [platform, labelEn, labelKo, icon] = ["packagist", "Packagist", "Packagist", "package"];
    } else if (matches("pub.dev")) {
      [platform, labelEn, labelKo, icon] = ["pub", "pub.dev", "pub.dev", "package"];
    } else if (matches("hex.pm")) {
      [platform, labelEn, labelKo, icon] = ["hex", "Hex", "Hex", "package"];
    } else if (matches("metacpan.org") || matches("cpan.org")) {
      [platform, labelEn, labelKo, icon] = ["cpan", "CPAN", "CPAN", "package"];
    } else if (matches("cran.r-project.org")) {
      [platform, labelEn, labelKo, icon] = ["cran", "CRAN", "CRAN", "package"];
    } else if (matches("bioconductor.org")) {
      [platform, labelEn, labelKo, icon] = ["bioconductor", "Bioconductor", "Bioconductor", "package"];
    } else if (matches("huggingface.co")) {
      [platform, labelEn, labelKo, icon] = ["huggingface", "Hugging Face", "Hugging Face", "package"];
    } else if (matches("pkg.go.dev")) {
      [platform, labelEn, labelKo, icon] = ["go-package", "Go package", "Go 패키지", "package"];
    } else if (matches("quay.io")) {
      [platform, labelEn, labelKo, icon] = ["quay", "Quay", "Quay", "container"];
    } else if (matches("flathub.org")) {
      [platform, labelEn, labelKo, icon] = ["flathub", "Flathub", "Flathub", "package"];
    } else if (matches("snapcraft.io")) {
      [platform, labelEn, labelKo, icon] = ["snapcraft", "Snap Store", "Snap Store", "package"];
    } else if (matches("formulae.brew.sh")) {
      [platform, labelEn, labelKo, icon] = ["homebrew", "Homebrew", "Homebrew", "package"];
    } else if (matches("zenodo.org") || (matches("doi.org") && path.includes("zenodo."))) {
      [platform, labelEn, labelKo, icon] = ["zenodo", "Zenodo", "Zenodo", "archive"];
    } else if (matches("figshare.com")) {
      [platform, labelEn, labelKo, icon] = ["figshare", "Figshare", "Figshare", "archive"];
    } else if (matches("osf.io")) {
      [platform, labelEn, labelKo, icon] = ["osf", "OSF", "OSF", "archive"];
    } else if (matches("softwareheritage.org")) {
      [platform, labelEn, labelKo, icon] = ["software-heritage", "Software Heritage", "Software Heritage", "archive"];
    } else if (matches("doi.org")) {
      [platform, labelEn, labelKo, icon] = ["doi", "DOI", "DOI", "archive"];
    }

    const customLabel = typeof value === "object" && value
      ? value[`label_${language}`] || value.label || ""
      : "";
    return {
      url: url.href,
      platform,
      icon,
      label: customLabel || (language === "ko" ? labelKo : labelEn)
    };
  }

  function softwareLinks(item, language = "en") {
    return (Array.isArray(item?.links) ? item.links : [])
      .map((value) => softwareLinkInfo(value, language))
      .filter(Boolean);
  }

  function contentMediaPath(value) {
    const source = typeof value === "object" && value ? value.src : value;
    const relative = String(source || "").trim().replaceAll("\\", "/").replace(/^\/+/, "");
    if (!relative || relative.split("/").some((part) => !part || part === "." || part === "..")) return "";
    return `data/media/${relative}`;
  }

  function contentMediaType(value) {
    if (typeof value === "object" && value?.type === "video") return "video";
    if (typeof value === "object" && value?.type === "image") return "image";
    const source = typeof value === "object" && value ? value.src : value;
    const extension = String(source || "").split(/[?#]/, 1)[0].split(".").pop().toLowerCase();
    return ["mp4", "webm", "ogv", "ogg", "mov", "m4v"].includes(extension) ? "video" : "image";
  }

  function documentPath(filename) {
    if (typeof filename !== "string" || !filename || filename.trim() !== filename ||
      filename.length > 240 || /[<>:"/\\|?*\u0000-\u001f\u007f]/.test(filename) || !/\.pdf$/i.test(filename)) return "";
    return `data/document/${encodeURIComponent(filename)}`;
  }

  function validGPA(gpa) {
    return gpa && typeof gpa === "object" && !Array.isArray(gpa)
      && Number.isFinite(gpa.value) && Number.isFinite(gpa.max)
      && gpa.value >= 0 && gpa.max > 0 && gpa.value <= gpa.max;
  }

  function educationThesisURL(value) {
    if (typeof value !== "string" || !/^https?:\/\//.test(value) || /[\u0000-\u001f\u007f]/.test(value)) return "";
    return safeExternalURL(value);
  }

  function educationThesisHTML(item, language = "en") {
    const title = item[`thesis_${language}`] || item.thesis_en;
    if (!title) return "";
    const url = educationThesisURL(item.thesis_url);
    const text = escapeHTML(title);
    const linkedTitle = url ? `<a class="education-thesis-link" href="${escapeHTML(url)}" target="_blank" rel="noreferrer">${text}</a>` : text;
    return `<p>${language === "ko" ? "학위논문:" : "Thesis:"} <em>${linkedTitle}</em></p>`;
  }

  function validateEducationGPA(item, context) {
    if (Object.hasOwn(item, "gpa") && !validGPA(item.gpa)) {
      throw new Error(`${context}: gpa must contain numeric value and max, with 0 <= value <= max and max > 0.`);
    }
  }

  function formatGPA(gpa) {
    if (!validGPA(gpa)) return "";
    const options = { minimumFractionDigits: 1, maximumFractionDigits: 20, useGrouping: false };
    return `GPA: ${gpa.value.toLocaleString("en-US", options)} / ${gpa.max.toLocaleString("en-US", options)}`;
  }

  function validateDocumentReference(item, context) {
    if (!Object.hasOwn(item, "pdf")) return;
    if (typeof item.pdf !== "string" || (item.pdf && !documentPath(item.pdf))) {
      throw new Error(`${context}: pdf must be a PDF filename in data/document or an empty string.`);
    }
  }

  window.SiteData = Object.freeze({
    FILES,
    WEB_PAGES,
    ACADEMIC_ACTIVITY_CATEGORIES,
    PATENT_TYPES,
    applyColorTheme,
    loadAll,
    loadVersion,
    visibleInCV,
    webPageVisible,
    splitList,
    truthy,
    escapeHTML,
    periodHTML,
    pick,
    patentStatus,
    patentType,
    patentPeopleHTML,
    patentGroups,
    patentDate,
    patentTitle,
    patentDetails,
    patentURL,
    publicationTitle,
    publicationVenue,
    publicationReference,
    publicationConferenceInfo,
    publicationVenueHTML,
    personName,
    personNotesHTML,
    displayAuthors,
    isFirstAuthor,
    publicationTypeLabel,
    publicationTopicLabel,
    projectThemeLabel,
    projectPeriod,
    projectNotes,
    softwareNotes,
    publicationStatus,
    publicationStatusRank,
    publicationYear,
    normalizeText,
    normalizeTopicSettings,
    topicColor,
    topicForegroundColor,
    publicationCitation,
    publicationCitationKey,
    bibtexEntry,
    download,
    downloadPublicationBib,
    safeExternalURL,
    softwareLinkInfo,
    softwareLinks,
    contentMediaPath,
    contentMediaType,
    documentPath,
    educationThesisHTML,
    formatGPA,
    scholarshipTargetCurrency,
    formatScholarshipAmount,
    resolveScholarshipAmount,
    resolveScholarshipAmountDetails,
    formatScholarshipConversionNote
  });
}());
