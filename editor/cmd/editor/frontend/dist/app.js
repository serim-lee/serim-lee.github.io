const SITE_COLOR_THEMES = window.SiteThemes.presets;
const DEFAULT_EDITOR_SHORTCUTS = Object.freeze({
  save: "Ctrl+S",
  citation_update: "Ctrl+U",
  push: "Ctrl+P",
  revert: "Ctrl+Shift+R",
  settings: "Ctrl+Comma",
});
const EDITOR_SHORTCUT_LABELS = Object.freeze({
  save: "저장",
  citation_update: "인용수 갱신",
  push: "Git push",
  revert: "변경 취소",
  settings: "설정 열기",
});
const RESERVED_EDITOR_SHORTCUTS = new Set([
  "Ctrl+A", "Ctrl+C", "Ctrl+V", "Ctrl+X", "Ctrl+Y", "Ctrl+Z", "Ctrl+Shift+Z",
  "Ctrl+F4", "Ctrl+L", "Ctrl+N", "Ctrl+O", "Ctrl+R", "Ctrl+T", "Ctrl+W",
  "Ctrl+Shift+N", "Ctrl+Shift+T", "Ctrl+Shift+W", "Alt+F4", "Alt+Tab",
]);

function normaliseSiteColorTheme(value) {
  return window.SiteThemes.normaliseThemeID(value);
}

const SECTION_LABELS = {
  experience: "경력",
  education: "학력",
  scholarships: "장학금",
  certifications: "자격",
  awards: "수상",
  patents: "특허",
  teaching: "교육",
  skills: "기술",
  academic_activities: "학술활동",
};

const ACADEMIC_ACTIVITY_CATEGORIES = Object.freeze([
  { value: "editorial_service", label: "Editorial Service · 편집 활동" },
  { value: "professional_service", label: "Professional Service · 전문 활동" },
  { value: "conference_service", label: "Conference Service · 학술대회 활동" },
  { value: "invited_talks", label: "Invited Talk · 초청 발표" },
  { value: "reviews", label: "Review · 리뷰" },
]);

const PATENT_TYPES = Object.freeze([
  { value: "domestic", label: "국내특허" },
  { value: "international", label: "국제특허" },
]);

function patentStatusLabel(item) {
  if (String(item.registration_number || "").trim()) return "등록";
  if (String(item.publication_number || "").trim()) return "공개";
  return String(item.application_number || "").trim() ? "출원" : "번호 미입력";
}

const CV_SECTION_LABELS = {
  experience: "경력",
  education: "학력",
  projects: "프로젝트",
  publications: "논문",
  software: "소프트웨어",
  awards: "수상",
  academic_activities: "학술활동",
  patents: "특허",
  teaching: "교육",
  scholarships: "장학",
  certifications: "자격",
  skills: "기술",
};

const WEB_PAGE_LABELS = Object.freeze({
  publications: "논문",
  projects: "프로젝트",
  software: "소프트웨어",
  board: "게시판",
});

const SOFTWARE_TECHNOLOGY_GROUPS = Object.freeze([
  {
    label: "프로그래밍 언어",
    items: [
      { key: "python", value: "Python", label: "Python", icon: "python.svg" },
      { key: "go", value: "Go", label: "Go", icon: "go.svg" },
      { key: "rust", value: "Rust", label: "Rust", icon: "rust.svg" },
      { key: "javascript", value: "JavaScript", label: "JavaScript", icon: "javascript.svg" },
      { key: "typescript", value: "TypeScript", label: "TypeScript", icon: "typescript.svg" },
      { key: "java", value: "Java", label: "Java", icon: "java.svg" },
      { key: "c", value: "C", label: "C", icon: "c.svg" },
      { key: "c++", value: "C++", label: "C++", icon: "cplusplus.svg" },
      { key: "c#", value: "C#", label: "C#", icon: "csharp.svg" },
      { key: "r", value: "R", label: "R", icon: "r.svg" },
      { key: "matlab", value: "MATLAB", label: "MATLAB", icon: "matlab.svg" },
    ],
  },
  {
    label: "웹",
    items: [
      { key: "html", value: "HTML", label: "HTML", icon: "html.svg" },
      { key: "css", value: "CSS", label: "CSS", icon: "css.svg" },
      { key: "react", value: "React", label: "React", icon: "react.svg" },
      { key: "vue", value: "Vue", label: "Vue", icon: "vue.svg" },
      { key: "svelte", value: "Svelte", label: "Svelte", icon: "svelte.svg" },
    ],
  },
  {
    label: "설계 · 시뮬레이션",
    items: [
      { key: "rhino", value: "Rhino", label: "Rhino", icon: "rhino.svg" },
      { key: "grasshopper", value: "Grasshopper", label: "Grasshopper", icon: "grasshopper.svg" },
      { key: "energyplus", value: "EnergyPlus", label: "EnergyPlus", icon: "energyplus.svg" },
    ],
  },
  {
    label: "앱 프레임워크",
    items: [
      { key: "wails", value: "Wails", label: "Wails", icon: "wails.svg" },
      { key: "egui", value: "egui", label: "egui", icon: "egui.svg" },
    ],
  },
  {
    label: "문서",
    items: [
      { key: "excel", value: "Excel", label: "Excel", icon: "excel.svg" },
      { key: "word", value: "Word", label: "Microsoft Word", icon: "word.svg" },
      { key: "pdf", value: "PDF", label: "PDF", icon: "pdf.svg" },
      { key: "latex", value: "LaTeX", label: "LaTeX", icon: "latex.svg" },
    ],
  },
]);

const SOFTWARE_TECHNOLOGIES = Object.freeze(SOFTWARE_TECHNOLOGY_GROUPS.flatMap((group) => group.items));
const SOFTWARE_TECHNOLOGY_BY_KEY = new Map(SOFTWARE_TECHNOLOGIES.map((technology) => [technology.key, technology]));
const SOFTWARE_TECHNOLOGY_ALIASES = Object.freeze({
  js: "javascript",
  ts: "typescript",
  golang: "go",
  csharp: "c#",
  "c-sharp": "c#",
  rhinoceros: "rhino",
  rhino3d: "rhino",
  "rhino 3d": "rhino",
  grasshopper3d: "grasshopper",
  "grasshopper 3d": "grasshopper",
  eplus: "energyplus",
  tex: "latex",
  "microsoft excel": "excel",
  "microsoft word": "word",
  docx: "word",
  acrobat: "pdf",
  "adobe acrobat": "pdf",
});

const state = {
  bridge: null,
  baseline: null,
  draft: null,
  revision: "",
  profileBaseline: null,
  profileDraft: null,
  profileMediaBaseline: null,
  profileMediaDraft: null,
  profileRevision: "",
  boardBaseline: null,
  boardDraft: null,
  boardRevision: "",
  selectedBoardKey: "",
  projectsBaseline: null,
  projectsDraft: null,
  projectsRevision: "",
  selectedProjectKey: "",
  softwareBaseline: null,
  softwareDraft: null,
  softwareRevision: "",
  selectedSoftwareKey: "",
  peopleBaseline: null,
  peopleDraft: null,
  peopleRevision: "",
  selectedPersonKey: "",
  awardsBaseline: null,
  awardsDraft: null,
  awardsRevision: "",
  selectedAwardKey: "",
  academicActivitiesBaseline: null,
  academicActivitiesDraft: null,
  academicActivitiesRevision: "",
  selectedAcademicActivityKey: "",
  publicationsBaseline: null,
  publicationsDraft: null,
  publicationsRevision: "",
  selectedPublicationKey: "",
  usage: { project_themes: {}, publication_topics: {} },
  loading: true,
  saving: false,
  dirty: false,
  validationErrors: new Map(),
  nextClientKey: 1,
  nextBoardKey: 1,
  nextBoardMediaKey: 1,
  nextProjectKey: 1,
  nextSoftwareKey: 1,
  nextPersonKey: 1,
  nextAwardKey: 1,
  nextAcademicActivityKey: 1,
  nextPublicationKey: 1,
  nextContentRowKey: 1,
  drag: null,
  statusText: "데이터를 불러오는 중…",
  statusTone: "",
  dirtySync: Promise.resolve(),
  stagingOps: Promise.resolve(),
  boardDropBusy: false,
  boardDropCount: 0,
  boardDropMessage: "",
  profileDropMessage: "",
  documentUploads: {},
  profileExpandedRows: new Set(),
  contentDropMessages: {},
  listFilters: {
    peopleSearch: "",
    publicationsSearch: "",
    publicationsStatus: "all",
  },
  cvExpandedSections: new Set(),
  cvPublicationSearch: "",
  fileDropBound: false,
  discarding: false,
  saveCompletionTimer: null,
  saveCompletionVisible: false,
  operation: null,
  operationSequence: 0,
  operationCloseTimer: null,
  operationEventsBound: false,
  shortcutCaptureAction: "",
  editorUpdate: {
    info: null,
    infoLoading: false,
    checking: false,
    installing: false,
    phase: "initial",
    latestVersion: "",
    available: false,
    releases: [],
    error: "",
  },
};

const taxonomyConfig = {
  topics: {
    listKey: "topics",
    fallbackKey: "topic_fallback",
    usageKey: "topics",
    listElement: "topic-list",
    fallbackElement: "topic-fallback",
    itemName: "주제",
    colors: ["#315f49", "#3d6f86", "#80643b", "#775778", "#4f6e31", "#8a5148"],
  },
};

const TOPIC_COLOR_PALETTES = [
  { id: "balanced", label: "균형", colors: ["#0b6fc9","#ad5505","#1d7f34","#7b3c96","#cc283e","#087a74","#736c04","#a74f6f","#56341c","#425772"] },
  { id: "vivid", label: "선명", colors: ["#056ada","#b05201","#047d30","#8c36a1","#da0439","#037c78","#777104","#a53461","#58341b","#3c5b7d"] },
  { id: "pastel", label: "파스텔", colors: ["#89b7e2","#f2b280","#a9cc95","#aba1de","#ed9c9a","#8dcec3","#e9d580","#daa7cb","#c4ab93","#a3a5ac"] },
  { id: "deep", label: "깊은", colors: ["#0655a7","#833d01","#06611b","#5a186f","#ab0d2d","#08645f","#605b01","#671336","#382107","#17334e"] },
  { id: "nordic", label: "북유럽", colors: ["#7098b8","#cd816f","#8aa48d","#c4aa63","#4f6280","#bd9ba8","#cedadd","#3e7056","#8e7a9f","#d2bea3"] },
  { id: "earth", label: "대지", colors: ["#b66b4b","#b98b3c","#7f8848","#4f6081","#a56e78","#c3937a","#6e4e3f","#a5b69a","#e2c58f","#788181"] },
  { id: "retro", label: "레트로", colors: ["#287d7b","#d2aa3b","#c74b41","#8b4969","#f0ddaa","#899b4a","#577d9e","#d78943","#6d4f45","#d49a9b"] },
  { id: "jewel", label: "보석", colors: ["#0467c6","#a34c03","#057923","#6d2c92","#c10936","#04726a","#6b6501","#b91976","#4f2d1a","#476889"] },
];
let taxonomyColorTarget = null;

function closeTaxonomyColorPopover(restoreFocus = false) {
  const target = taxonomyColorTarget;
  taxonomyColorTarget = null;
  document.querySelector("#taxonomy-color-popover").hidden = true;
  target?.trigger.setAttribute("aria-expanded", "false");
  if (restoreFocus && target?.trigger.isConnected) target.trigger.focus({ preventScroll: true });
}

function syncTaxonomyColorSelection() {
  if (!taxonomyColorTarget) return;
  const { kind, clientKey } = taxonomyColorTarget;
  const found = taxonomyItem(kind, clientKey);
  if (found.index < 0) return closeTaxonomyColorPopover();
  const item = found.items[found.index];
  const color = String(item.color || "").trim().toLowerCase();
  const popup = document.querySelector("#taxonomy-color-popover");
  popup.querySelector(".taxonomy-color-selection").textContent = `${item.label_ko || item.label_en || "이름 없는 주제"} · ${color || "자동"}`;
  popup.querySelectorAll("[data-palette-color]").forEach((button) => {
    button.setAttribute("aria-pressed", String(button.dataset.paletteColor === color));
  });
  popup.querySelector(".taxonomy-color-auto").setAttribute("aria-pressed", String(!color));
  popup.querySelector('[type="color"]').value = /^#[0-9a-f]{6}$/.test(color) ? color : taxonomyDefaultColor(kind, found.index);
}

function chooseTaxonomyColor(color, close = true) {
  if (!taxonomyColorTarget || state.loading || state.saving || state.discarding || state.boardDropBusy) return;
  const input = taxonomyColorTarget.trigger.closest(".taxonomy-color-control")?.querySelector('[type="text"]');
  if (!input?.isConnected) return closeTaxonomyColorPopover();
  input.value = color;
  input.dispatchEvent(new Event("input", { bubbles: true }));
  if (close) closeTaxonomyColorPopover(true);
  else syncTaxonomyColorSelection();
}

function initializeTaxonomyColorPopover() {
  const popup = document.querySelector("#taxonomy-color-popover");
  const palettes = popup.querySelector(".taxonomy-color-palettes");
  if (palettes.childElementCount) return popup;
  TOPIC_COLOR_PALETTES.forEach((palette) => {
    const group = document.createElement("section");
    group.className = "taxonomy-color-palette";
    const title = document.createElement("h4");
    title.textContent = palette.label;
    const colors = document.createElement("div");
    colors.className = "taxonomy-palette-swatches";
    colors.setAttribute("role", "group");
    colors.setAttribute("aria-label", `${palette.label} 팔레트`);
    palette.colors.forEach((color, index) => {
      const button = document.createElement("button");
      button.type = "button";
      button.className = "taxonomy-palette-swatch";
      button.dataset.paletteColor = color;
      button.style.backgroundColor = color;
      button.tabIndex = -1;
      button.title = `${palette.label} ${index + 1} · ${color.toUpperCase()}`;
      button.setAttribute("aria-label", button.title);
      button.addEventListener("click", () => chooseTaxonomyColor(color));
      colors.append(button);
    });
    group.append(title, colors);
    palettes.append(group);
  });
  popup.querySelector(".taxonomy-color-close").addEventListener("click", () => closeTaxonomyColorPopover(true));
  popup.querySelector(".taxonomy-color-auto").addEventListener("click", () => chooseTaxonomyColor(""));
  const custom = popup.querySelector('[type="color"]');
  custom.addEventListener("input", () => chooseTaxonomyColor(custom.value, false));
  custom.addEventListener("change", () => closeTaxonomyColorPopover(true));
  popup.addEventListener("keydown", (event) => {
    const button = event.target.closest("[data-palette-color]");
    if (!button) return;
    const buttons = [...popup.querySelectorAll("[data-palette-color]")];
    const index = buttons.indexOf(button);
    const offset = { ArrowLeft: -1, ArrowRight: 1, ArrowUp: -10, ArrowDown: 10 }[event.key];
    let next = offset === undefined ? undefined : index + offset;
    if (event.key === "Home") next = event.ctrlKey ? 0 : Math.floor(index / 10) * 10;
    if (event.key === "End") next = event.ctrlKey ? buttons.length - 1 : Math.floor(index / 10) * 10 + 9;
    if (next === undefined) return;
    event.preventDefault();
    if (next < 0 || next >= buttons.length) return;
    button.tabIndex = -1;
    buttons[next].tabIndex = 0;
    buttons[next].focus({ preventScroll: true });
    buttons[next].scrollIntoView({ block: "nearest", inline: "nearest" });
  });
  document.addEventListener("keydown", (event) => {
    if (!taxonomyColorTarget || event.key !== "Escape") return;
    event.preventDefault();
    closeTaxonomyColorPopover(true);
  });
  document.addEventListener("pointerdown", (event) => {
    if (taxonomyColorTarget && !popup.contains(event.target) && !taxonomyColorTarget.trigger.contains(event.target)) closeTaxonomyColorPopover();
  }, true);
  document.addEventListener("focusin", (event) => {
    if (taxonomyColorTarget && !popup.contains(event.target) && !taxonomyColorTarget.trigger.contains(event.target)) closeTaxonomyColorPopover();
  });
  document.addEventListener("scroll", (event) => {
    if (taxonomyColorTarget && event.target !== popup && !popup.contains(event.target)) closeTaxonomyColorPopover();
  }, true);
  window.addEventListener("resize", () => closeTaxonomyColorPopover());
  return popup;
}

function openTaxonomyColorPopover(trigger) {
  if (taxonomyColorTarget?.trigger === trigger) return closeTaxonomyColorPopover(true);
  closeTaxonomyColorPopover();
  const popup = initializeTaxonomyColorPopover();
  taxonomyColorTarget = { trigger, kind: trigger.dataset.kind, clientKey: trigger.dataset.clientKey };
  trigger.setAttribute("aria-expanded", "true");
  syncTaxonomyColorSelection();
  popup.hidden = false;
  const rect = trigger.getBoundingClientRect();
  const size = popup.getBoundingClientRect();
  popup.style.left = `${Math.max(12, Math.min(rect.right - size.width, window.innerWidth - size.width - 12))}px`;
  const below = rect.bottom + 6;
  const above = rect.top - size.height - 6;
  popup.style.top = `${below + size.height <= window.innerHeight - 12 ? below : above >= 12 ? above : Math.max(12, window.innerHeight - size.height - 12)}px`;
  const buttons = [...popup.querySelectorAll("[data-palette-color]")];
  const selected = buttons.find((button) => button.getAttribute("aria-pressed") === "true") || buttons[0];
  buttons.forEach((button) => { button.tabIndex = button === selected ? 0 : -1; });
  selected.focus({ preventScroll: true });
  selected.scrollIntoView({ block: "nearest", inline: "nearest" });
}

const catalogConfig = {
  journals: {
    label: "저널",
    fields: [["name", "이름", "bilingual"], ["publisher", "출판사", "bilingual"],
      ["issn", "ISSN"], ["url", "URL", "url"]],
  },
  conferences: {
    label: "학회",
    fields: [["name", "학술대회명", "bilingual"],
      ["start_date", "기간 · 시작일", "date"], ["end_date", "기간 · 종료일", "date"],
      ["location", "장소", "bilingual"], ["url", "공식 URL", "url"],
      ["proceedings_title", "이름", "bilingual"], ["proceedings_publisher", "발행기관", "bilingual"],
      ["proceedings_volume", "권"], ["proceedings_issue", "호"],
      ["proceedings_series", "시리즈명", "bilingual"], ["proceedings_series_volume", "시리즈 권"],
      ["proceedings_date", "발행일", "publication-date"], ["proceedings_issn", "ISSN"]],
  },
};
const catalogSelection = { journals: "", conferences: "" };
const catalogIndexingOptions = [["SCIE", "SCIE"], ["SCOPUS", "SCOPUS"], ["KCI", "KCI"]];
const publicationBibliographyFields = [
  ["volume", "권"], ["issue", "호"], ["pages", "쪽수"], ["elocator", "논문번호 (eLocator)"],
  ["online_date", "온라인 게재일"], ["issue_date", "Issue 발행일"],
];

const entityConfig = {
  people: {
    draftKey: "peopleDraft",
    selectedKey: "selectedPersonKey",
    domPrefix: "people",
    singular: "사람",
    manualOrder: false,
    sorter: sortedPeopleForDisplay,
  },
  awards: {
    draftKey: "awardsDraft",
    selectedKey: "selectedAwardKey",
    domPrefix: "awards",
    singular: "수상",
    manualOrder: false,
    sorter: sortedAwards,
  },
  publications: {
    draftKey: "publicationsDraft",
    selectedKey: "selectedPublicationKey",
    domPrefix: "publications",
    singular: "논문",
    manualOrder: false,
    sorter: sortedPublications,
  },
};

const tabs = Array.from(document.querySelectorAll('[role="tab"]'));
const panels = Array.from(document.querySelectorAll('[role="tabpanel"]'));
const saveButton = document.querySelector("#save-button");
const discardButton = document.querySelector("#discard-button");
const citationUpdateButton = document.querySelector("#citation-update-button");
const pushButton = document.querySelector("#push-button");
const operationPopover = document.querySelector("#operation-popover");
const operationHomeParent = operationPopover.parentElement;
const editorUpdateDialog = document.querySelector("#editor-update-dialog");
const operationProgress = document.querySelector("#operation-progress");
const operationMessage = document.querySelector("#operation-message");
const operationOutput = document.querySelector("#operation-output");
const operationError = document.querySelector("#operation-error");
const operationPercent = document.querySelector("#operation-percent");
const operationClose = document.querySelector("#operation-close");
const saveStatus = document.querySelector("#save-status");
const loadError = document.querySelector("#load-error");
const loadErrorMessage = document.querySelector("#load-error-message");
const saveError = document.querySelector("#save-error");
const saveErrorMessage = document.querySelector("#save-error-message");
const retryButton = document.querySelector("#retry-button");
const moveAnnouncer = document.querySelector("#move-announcer");

function clone(value) {
  return JSON.parse(JSON.stringify(value));
}

function errorMessage(error) {
  return String(error?.message || error || "알 수 없는 오류").replace(/^Error:\s*/, "");
}

function confirmDeleteItem(kind, title) {
  return window.confirm(`삭제할까요?\n\n${kind}: ${title}`);
}

function setStatus(text, tone = "") {
  state.statusText = text;
  state.statusTone = tone;
  saveStatus.textContent = text;
  saveStatus.dataset.tone = tone;
}

function cancelSaveCompletionStatus() {
  if (state.saveCompletionTimer !== null) {
    window.clearTimeout(state.saveCompletionTimer);
    state.saveCompletionTimer = null;
  }
  state.saveCompletionVisible = false;
}

function showSaveCompletionStatus() {
  cancelSaveCompletionStatus();
  state.saveCompletionVisible = true;
  setStatus("저장 완료", "success");
  state.saveCompletionTimer = window.setTimeout(() => {
    state.saveCompletionTimer = null;
    state.saveCompletionVisible = false;
    updateToolbar();
  }, 1500);
}

function announce(message) {
  moveAnnouncer.textContent = "";
  requestAnimationFrame(() => {
    moveAnnouncer.textContent = message;
  });
}

function beginOperation(kind) {
  if (state.operationCloseTimer !== null) window.clearTimeout(state.operationCloseTimer);
  state.operationCloseTimer = null;
  state.operation = { id: `${Date.now()}-${++state.operationSequence}`, kind, done: false };
  document.querySelector("#operation-title").textContent = kind === "update" ? "업데이트" : kind === "citations" ? "인용수 갱신" : kind === "push" ? "Git push" : "저장";
  operationOutput.setAttribute("aria-label", kind === "update" ? "업데이트 진행 기록" : "인용수 갱신 결과");
  operationPopover.dataset.kind = kind;
  operationPopover.dataset.tone = "";
  operationPopover.hidden = false;
  operationPopover.setAttribute("aria-busy", "true");
  operationError.hidden = true;
  operationError.textContent = "";
  operationOutput.textContent = "";
  operationOutput.hidden = true;
  operationClose.disabled = true;
  saveButton.setAttribute("aria-expanded", String(kind === "save"));
  citationUpdateButton.setAttribute("aria-expanded", String(kind === "citations"));
  pushButton.setAttribute("aria-expanded", String(kind === "push"));
  document.querySelector("#editor-update-install")?.setAttribute("aria-expanded", String(kind === "update"));
  if (kind === "update" && editorUpdateDialog) {
    editorUpdateDialog.append(operationPopover);
    if (!editorUpdateDialog.open) editorUpdateDialog.showModal();
    editorUpdateDialog.focus({ preventScroll: true });
  }
  updateOperationProgress("처리 중…", -1);
}

function updateOperationProgress(message, percent) {
  operationMessage.textContent = message;
  if (Number.isFinite(percent) && percent >= 0) {
    operationProgress.value = Math.min(100, percent);
    operationPercent.textContent = `${Math.min(100, percent)}%`;
  } else {
    operationProgress.removeAttribute("value");
    operationPercent.textContent = "";
  }
}

function updateOperationOutput(output, append = false) {
  const followOutput = operationOutput.hidden || operationOutput.scrollTop + operationOutput.clientHeight >= operationOutput.scrollHeight - 24;
  const value = String(output || "");
  operationOutput.textContent = append ? operationOutput.textContent + value : value;
  operationOutput.hidden = !operationOutput.textContent;
  if (followOutput) operationOutput.scrollTop = operationOutput.scrollHeight;
}

function closeOperation() {
  if (state.saving || state.editorUpdate.installing) return;
  operationPopover.hidden = true;
  if (editorUpdateDialog?.open) editorUpdateDialog.close();
  if (operationPopover.parentElement === editorUpdateDialog) operationHomeParent.append(operationPopover);
  saveButton.setAttribute("aria-expanded", "false");
  citationUpdateButton.setAttribute("aria-expanded", "false");
  pushButton.setAttribute("aria-expanded", "false");
  document.querySelector("#editor-update-install")?.setAttribute("aria-expanded", "false");
}

function finishOperation(success, message, detail = "") {
  updateOperationProgress(message, success ? 100 : (operationProgress.hasAttribute("value") ? operationProgress.value : 0));
  state.operation.done = true;
  state.operation.success = success;
  operationPopover.dataset.tone = success ? "success" : "error";
  operationPopover.setAttribute("aria-busy", "false");
  operationClose.disabled = false;
  if (!success) {
    operationPercent.textContent = "";
    operationError.textContent = detail;
    operationError.hidden = !detail;
  } else if (state.operation.kind !== "citations" && state.operation.kind !== "update") {
    state.operationCloseTimer = window.setTimeout(closeOperation, 1800);
  }
}

function installOperationProgress() {
  if (state.operationEventsBound || !window.runtime?.EventsOnMultiple) return;
  window.runtime.EventsOnMultiple("editor:operation-progress", (progress) => {
    if (!state.operation || state.operation.done || progress.operation_id !== state.operation.id) return;
    if (typeof progress.output === "string") updateOperationOutput(progress.output, true);
    updateOperationProgress(progress.message, progress.percent);
  }, -1);
  state.operationEventsBound = true;
}

async function waitForBridge() {
  for (let attempt = 0; attempt < 100; attempt += 1) {
    const bridge = window.go?.main?.App;
    if (
      bridge?.LoadEditorData
      && bridge?.SaveEditorData
      && bridge?.PushGit
      && bridge?.StageBoardMedia
      && bridge?.DiscardBoardMedia
      && bridge?.StageDocument
      && bridge?.DiscardDocuments
      && bridge?.SetDirty
    ) return bridge;
    await new Promise((resolve) => window.setTimeout(resolve, 50));
  }
  throw new Error("Profile-Editor 백엔드에 연결할 수 없습니다.");
}

function addClientKeys(settings) {
  const result = clone(settings);
  result.hidden_web_pages = [...(settings.hidden_web_pages || [])];
  result.color_theme = normaliseSiteColorTheme(result.color_theme);
  result.custom_themes = window.SiteThemes.normaliseCustomThemes(result.custom_themes);
  result.editor_shortcuts = { ...DEFAULT_EDITOR_SHORTCUTS, ...(result.editor_shortcuts || {}) };
  Object.values(taxonomyConfig).forEach((config) => {
    result[config.listKey] = (result[config.listKey] || []).map((item) => ({
      ...item,
      _clientKey: item.id ? `id:${item.id}` : `new:${state.nextClientKey++}`,
    }));
  });
  Object.keys(catalogConfig).forEach((kind) => {
    result[kind] = (result[kind] || []).map((item) => ({
      ...normaliseCatalogItem(item), indexing: [...(item.indexing || [])], _clientKey: `catalog:${kind}:${state.nextClientKey++}`,
    }));
  });
  return result;
}

function toSettingsPayload(settings) {
  return {
    version: window.SiteDataVersion.validateVersion(settings.version || window.SiteDataVersion.CURRENT_VERSION, "settings.json"),
    journals: catalogPayload(settings.journals),
    conferences: catalogPayload(settings.conferences),
    color_theme: normaliseSiteColorTheme(settings.color_theme),
    custom_themes: clone(settings.custom_themes),
    editor_shortcuts: { ...DEFAULT_EDITOR_SHORTCUTS, ...(settings.editor_shortcuts || {}) },
    main_page_sections: [...settings.main_page_sections],
    hidden_main_page_sections: [...settings.hidden_main_page_sections],
    cv_sections: [...settings.cv_sections],
    hidden_cv_sections: [...settings.hidden_cv_sections],
    hidden_web_pages: [...(settings.hidden_web_pages || [])],
    topics: settings.topics.map(({ id = "", label_en, label_ko, color }) => ({
      id, label_en, label_ko, ...(color !== undefined ? { color } : {}),
    })),
    topic_fallback: {
      label_en: settings.topic_fallback.label_en,
      label_ko: settings.topic_fallback.label_ko,
    },
  };
}

function comparable(settings) {
  return JSON.stringify(toSettingsPayload(settings));
}

function hydrateProfile(profile, people = []) {
  const hydrated = clone(profile || {});
  if (!Object.hasOwn(hydrated, "patents")) hydrated.patents = [];
  const addKeys = (value, prefix) => {
    if (!Array.isArray(value)) return;
    value.forEach((item, index) => {
      if (!item || typeof item !== "object" || Array.isArray(item)) return;
      item._clientKey = `${prefix}:${index}:${state.nextContentRowKey++}`;
      if (["experience", "education", "teaching", "scholarships", "certifications", "patents", "skills"].includes(prefix)) {
        item.visible_in_CV = item.visible_in_CV !== false;
      }
    });
  };
  ["experience", "education", "teaching", "scholarships", "certifications", "patents", "skills"]
    .forEach((section) => addKeys(hydrated[section], section));
  const peopleByID = new Map(people.map((person, index) => [person.id,
    person._clientKey || (person.editor_key ? `existing-person:${person.editor_key}` : `existing-person:${index}`)]));
  (hydrated.patents || []).forEach((item) => {
    item._unresolvedInventorIDs = {};
    item.inventor_keys = (item.inventor_ids || []).map((id, index) => {
      const known = peopleByID.get(id);
      if (known) return known;
      const key = `unresolved-inventor:${item._clientKey}:${index}`;
      item._unresolvedInventorIDs[key] = id;
      return key;
    });
  });
  (hydrated.scholarships || []).forEach((item) => delete item.details);
  addKeys(hydrated.profile_card?.credentials, "credential");
  return hydrated;
}

function hydrateProfileMedia(mediaSummary = {}) {
  return {
    editor_key: mediaSummary.editor_key || "",
    original_name: mediaSummary.original_name || mediaSummary.src || "",
    size: Number(mediaSummary.size || 0),
    preview_url: mediaSummary.preview_url || "",
    stage_token: "",
    remove: false,
  };
}

function toProfilePayload(profile) {
  const result = clone(profile || {});
  (result.patents || []).forEach((item) => {
    if (Array.isArray(item.inventor_keys)) {
      item.inventor_ids = item.inventor_keys.map((key) =>
        (state.peopleDraft || []).find((person) => person._clientKey === key)?.id || item._unresolvedInventorIDs?.[key] || "");
    }
    ["inventor_keys", "_unresolvedInventorIDs", "_inventorSearch", "inventors_en", "inventors_ko", "status", "jurisdiction"]
      .forEach((key) => delete item[key]);
  });
  const clean = (value) => {
    if (Array.isArray(value)) {
      value.forEach(clean);
      return;
    }
    if (!value || typeof value !== "object") return;
    delete value._clientKey;
    delete value._editorKey;
    Object.values(value).forEach(clean);
  };
  clean(result);
  (result.scholarships || []).forEach((item) => delete item.details);
  return result;
}

function profileComparable(profile, mediaDraft) {
  return JSON.stringify({
    profile: toProfilePayload(profile),
    media_stage_token: mediaDraft?.stage_token || "",
    remove_media: mediaDraft?.remove === true,
  });
}

function hydrateBoard(items) {
  return (items || []).map((item) => {
    const hydrated = clone(item);
    return {
      ...hydrated,
      content_en: hydrated.content_en || "",
      content_ko: hydrated.content_ko || "",
      media: (hydrated.media || []).map((media, index) => ({
        ...media,
        _clientKey: media.editor_key
          ? `existing-media:${media.editor_key}`
          : `existing-media:${item.editor_key}:${index}`,
        _isNew: false,
        caption_en: media.caption_en || "",
        caption_ko: media.caption_ko || "",
      })),
      _clientKey: `existing:${item.editor_key}`,
      _isNew: false,
    };
  });
}

function compareBoardItems(left, right) {
  return String(right.start_date || "").localeCompare(String(left.start_date || ""))
    || String(right.end_date || "").localeCompare(String(left.end_date || ""))
    || Number(left._sortOrder || 0) - Number(right._sortOrder || 0);
}

function sortedBoardItems(items) {
  return items
    .map((item, index) => ({ item, index }))
    .sort((left, right) => compareBoardItems(
      { ...left.item, _sortOrder: left.index },
      { ...right.item, _sortOrder: right.index },
    ))
    .map(({ item }) => item);
}

function toBoardMediaPayload(media) {
  const reference = media.stage_token
    ? { stage_token: media.stage_token }
    : { editor_key: media.editor_key };
  return {
    ...reference,
    caption_en: media.caption_en || "",
    caption_ko: media.caption_ko || "",
  };
}

function toBoardPostPayload(item) {
  return {
    start_date: item.start_date || "",
    end_date: item.end_date || "",
    title_en: item.title_en || "",
    title_ko: item.title_ko || "",
    content_en: item.content_en || "",
    content_ko: item.content_ko || "",
    media: (item.media || []).map(toBoardMediaPayload),
  };
}

function toBoardPayload(items) {
  return sortedBoardItems(items).map((item) => {
    const post = toBoardPostPayload(item);
    return item._isNew
      ? { new_post: post }
      : { editor_key: item.editor_key, post };
  });
}

function toBoardSavePayload(items, baselineItems) {
  const baselinePosts = new Map(
    (baselineItems || [])
      .filter((item) => !item._isNew && item.editor_key)
      .map((item) => [item.editor_key, JSON.stringify(toBoardPostPayload(item))]),
  );
  return sortedBoardItems(items).map((item) => {
    const post = toBoardPostPayload(item);
    if (item._isNew) return { new_post: post };
    if (baselinePosts.get(item.editor_key) === JSON.stringify(post)) {
      return { editor_key: item.editor_key };
    }
    return { editor_key: item.editor_key, post };
  });
}

function boardComparable(items) {
  return JSON.stringify(toBoardPayload(items));
}

function hydrateContentMedia(media, ownerKey, prefix) {
  return (media || []).map((item, index) => ({
    ...clone(item),
    _clientKey: item.editor_key
      ? `existing-${prefix}-media:${item.editor_key}`
      : `existing-${prefix}-media:${ownerKey}:${index}`,
    _isNew: false,
    caption_en: item.caption_en || "",
    caption_ko: item.caption_ko || "",
  }));
}

function hydrateProjects(items) {
  return (items || []).map((item, index) => {
    const hydrated = clone(item);
    const ownerKey = item.editor_key || `project:${index}`;
    return {
      ...hydrated,
      visible_in_CV: hydrated.visible_in_CV !== false,
      _clientKey: item.editor_key ? `existing-project:${item.editor_key}` : `existing-project:${index}`,
      _isNew: false,
      start_date: hydrated.start_date || "",
      end_date: hydrated.end_date || "",
      title_en: hydrated.title_en || "",
      title_ko: hydrated.title_ko || "",
      theme: hydrated.theme || "",
      funder_en: hydrated.funder_en || "",
      funder_ko: hydrated.funder_ko || "",
      notes_en_rows: hydrateStringRows(hydrated.notes_en, "project-note-en", ownerKey),
      notes_kr_rows: hydrateStringRows(hydrated.notes_kr, "project-note-kr", ownerKey),
      media: hydrateContentMedia(hydrated.media, ownerKey, "project"),
    };
  });
}

function hydrateSoftware(items) {
  return (items || []).map((item, index) => {
    const hydrated = clone(item);
    const ownerKey = item.editor_key || item.id || `software:${index}`;
    return {
      ...hydrated,
      visible_in_CV: hydrated.visible_in_CV !== false,
      _clientKey: item.editor_key ? `existing-software:${item.editor_key}` : `existing-software:${index}`,
      _isNew: false,
      id: hydrated.id || "",
      name: hydrated.name || "",
      stage: hydrated.stage || "development",
      notes_en_rows: hydrateStringRows(hydrated.notes_en, "software-note-en", ownerKey),
      notes_kr_rows: hydrateStringRows(hydrated.notes_kr, "software-note-kr", ownerKey),
      links: (hydrated.links || []).map((link, linkIndex) => {
        const isString = typeof link === "string";
        return {
          ...(isString ? {} : clone(link)),
          _clientKey: `software-link:${ownerKey}:${linkIndex}:${state.nextContentRowKey++}`,
          url: isString ? link : link?.url || "",
          label: isString ? "" : link?.label || "",
          label_en: isString ? "" : link?.label_en || "",
          label_ko: isString ? "" : link?.label_ko || "",
        };
      }),
      technologies: (hydrated.technologies || []).map((technology, technologyIndex) => ({
        _clientKey: `software-technology:${ownerKey}:${technologyIndex}:${state.nextContentRowKey++}`,
        value: technology || "",
      })),
      media: hydrateContentMedia(hydrated.media, ownerKey, "software"),
    };
  });
}

function hydrateStringRows(values, prefix, ownerKey) {
  return (Array.isArray(values) ? values : []).map((value, index) => ({
    _clientKey: `${prefix}:${ownerKey}:${index}:${state.nextContentRowKey++}`,
    value: String(value || ""),
  }));
}

function hydratePeople(items) {
  return (items || []).map((item, index) => {
    const hydrated = clone(item);
    const ownerKey = item.editor_key || item.id || `person:${index}`;
    return {
      ...hydrated,
      _clientKey: item.editor_key ? `existing-person:${item.editor_key}` : `existing-person:${index}`,
      _isNew: false,
      id: hydrated.id || "",
      name_en: hydrated.name_en || "",
      name_ko: hydrated.name_ko || "",
      is_self: hydrated.is_self === true,
      notes_en_rows: hydrateStringRows(hydrated.notes_en, "person-note-en", ownerKey),
      notes_ko_rows: hydrateStringRows(hydrated.notes_ko, "person-note-ko", ownerKey),
    };
  });
}

function hydrateAwards(items) {
  return (items || []).map((item, index) => {
    const hydrated = clone(item);
    return {
      ...hydrated,
      visible_in_CV: hydrated.visible_in_CV !== false,
      _clientKey: item.editor_key ? `existing-award:${item.editor_key}` : `existing-award:${index}`,
      _isNew: false,
      id: hydrated.id || "",
      date: hydrated.date || "",
      title_en: hydrated.title_en || "",
      title_ko: hydrated.title_ko || "",
      organization_en: hydrated.organization_en || "",
      organization_ko: hydrated.organization_ko || "",
    };
  });
}

function hydrateAcademicActivities(items) {
  return (items || []).map((item, index) => {
    const ownerKey = item.editor_key || `academic-activity:${index}`;
    return {
      ...clone(item),
      visible_in_CV: item.visible_in_CV !== false,
      _clientKey: item.editor_key
        ? `existing-academic-activity:${item.editor_key}`
        : `existing-academic-activity:${index}`,
      _isNew: false,
      category: item.category || "editorial_service",
      journal_en: item.journal_en || "",
      journal_ko: item.journal_ko || "",
      completed_date_rows: hydrateStringRows(item.completed_dates, "academic-review-date", ownerKey),
      event_en: item.event_en || "",
      event_ko: item.event_ko || "",
      topic_en: item.topic_en || "",
      topic_ko: item.topic_ko || "",
      conference_en: item.conference_en || "",
      conference_ko: item.conference_ko || "",
      organization_en: item.organization_en || "",
      organization_ko: item.organization_ko || "",
      role_en: item.role_en || "",
      role_ko: item.role_ko || "",
      date: item.date || "",
      start_date: item.start_date || "",
      end_date: item.end_date || "",
    };
  });
}

function hydratePublications(items, people, awards) {
  const peopleByID = new Map((people || []).filter((item) => item.id).map((item) => [item.id, item._clientKey]));
  const awardsByID = new Map((awards || []).filter((item) => item.id).map((item) => [item.id, item._clientKey]));
  return (items || []).map((item, index) => {
    const hydrated = clone(item);
    const ownerKey = item.editor_key || `publication:${index}`;
    const unresolvedAuthorIDs = {};
    const authorKeys = (hydrated.author_ids || []).map((id, authorIndex) => {
      const known = peopleByID.get(id);
      if (known) return known;
      const key = `unresolved-author:${ownerKey}:${authorIndex}`;
      unresolvedAuthorIDs[key] = id;
      return key;
    });
    const knownAwardKey = hydrated.award_id ? awardsByID.get(hydrated.award_id) : "";
    return {
      ...hydrated,
      visible_in_CV: hydrated.visible_in_CV !== false,
      _clientKey: item.editor_key ? `existing-publication:${item.editor_key}` : `existing-publication:${index}`,
      _isNew: false,
      title_en: hydrated.title_en || "",
      title_ko: hydrated.title_ko || "",
      abstract_en: hydrated.abstract_en || "",
      abstract_ko: hydrated.abstract_ko || "",
      keywords_en_rows: hydrateStringRows(hydrated.keywords_en, "publication-keyword-en", ownerKey),
      keywords_ko_rows: hydrateStringRows(hydrated.keywords_ko, "publication-keyword-ko", ownerKey),
      author_keys: authorKeys,
      _unresolvedAuthorIDs: unresolvedAuthorIDs,
      date: hydrated.date || "",
      journal_id: hydrated.journal_id || "",
      conference_id: hydrated.conference_id || "",
      ...Object.fromEntries(publicationBibliographyFields.map(([field]) => [field, hydrated[field] || ""])),
      _status: hydrated.under_review ? "under_review" : hydrated.in_press ? "in_press" : "published",
      publication_type: hydrated.publication_type || "international-journal",
      topic: hydrated.topic || "",
      award_key: knownAwardKey || (hydrated.award_id ? `unresolved-award:${ownerKey}` : ""),
      _unresolvedAwardID: knownAwardKey ? "" : hydrated.award_id || "",
      doi: hydrated.doi || "",
      url: hydrated.url || "",
      note: hydrated.note || "",
    };
  });
}

function personRelationshipID(publication, key) {
  return (state.peopleDraft || []).find((person) => person._clientKey === key)?.id
    || publication._unresolvedAuthorIDs?.[key]
    || "";
}

function awardRelationshipID(publication) {
  if (!publication.award_key) return "";
  return (state.awardsDraft || []).find((award) => award._clientKey === publication.award_key)?.id
    || publication._unresolvedAwardID
    || "";
}

function toPersonPayload(item) {
  return {
    ...(item._isNew && item.id ? { id: item.id } : {}),
    name_en: item.name_en || "",
    name_ko: item.name_ko || "",
    is_self: item.is_self === true,
    notes_en: (item.notes_en_rows || []).map((row) => row.value || ""),
    notes_ko: (item.notes_ko_rows || []).map((row) => row.value || ""),
  };
}

function toAwardPayload(item) {
  return {
    ...(item._isNew && item.id ? { id: item.id } : {}),
    date: item.date || "",
    title_en: item.title_en || "",
    title_ko: item.title_ko || "",
    organization_en: item.organization_en || "",
    organization_ko: item.organization_ko || "",
    pdf: item.pdf || "",
  };
}

function toPublicationPayload(item) {
  return {
    title_en: item.title_en || "",
    title_ko: item.title_ko || "",
    abstract_en: item.abstract_en || "",
    abstract_ko: item.abstract_ko || "",
    keywords_en: (item.keywords_en_rows || []).map((row) => row.value || ""),
    keywords_ko: (item.keywords_ko_rows || []).map((row) => row.value || ""),
    author_ids: (item.author_keys || []).map((key) => personRelationshipID(item, key)).filter(Boolean),
    date: publicationMetadata(item).date,
    ...(item.journal_id ? { journal_id: item.journal_id } : {}),
    ...(item.conference_id ? { conference_id: item.conference_id } : {}),
    ...Object.fromEntries(publicationBibliographyFields.filter(([field]) => item[field]).map(([field]) => [field, item[field]])),
    under_review: item._status === "under_review",
    in_press: item._status === "in_press",
    publication_type: item.publication_type || "international-journal",
    topic: item.topic || "",
    award_id: awardRelationshipID(item),
    doi: item.doi || "",
    url: item.url || "",
    note: item.note || "",
  };
}

function publicationSortValue(item) {
  return item._status === "under_review" ? 2 : item._status === "in_press" ? 1 : 0;
}

function sortedAwards(items) {
  return (items || [])
    .map((item, index) => ({ item, index }))
    .sort((left, right) => String(right.item.date || "").localeCompare(String(left.item.date || ""))
      || left.index - right.index)
    .map(({ item }) => item);
}

function sortedPublications(items) {
  return (items || [])
    .map((item, index) => ({ item, index }))
    .sort((left, right) => publicationSortValue(right.item) - publicationSortValue(left.item)
      || String(publicationMetadata(right.item).date).localeCompare(String(publicationMetadata(left.item).date))
      || (() => {
        const leftTitle = String(left.item.title_en || left.item.title_ko || "").toLocaleLowerCase();
        const rightTitle = String(right.item.title_en || right.item.title_ko || "").toLocaleLowerCase();
        return leftTitle < rightTitle ? -1 : leftTitle > rightTitle ? 1 : 0;
      })()
      || left.index - right.index)
    .map(({ item }) => item);
}

function relationshipComparablePublication(item) {
  return {
    ...toPublicationPayload(item),
    author_ids: [...(item.author_keys || [])],
    award_id: item.award_key || "",
  };
}

function toEntitySavePayload(items, baselineItems, wrapper, payload, sorter = (value) => value || []) {
  const baseline = new Map(
    (baselineItems || [])
      .filter((item) => !item._isNew && item.editor_key)
      .map((item) => [item.editor_key, JSON.stringify(payload(item))]),
  );
  return sorter(items).map((item) => {
    const value = payload(item);
    const visibility = wrapper === "award" ? { visible_in_CV: item.visible_in_CV !== false } : {};
    if (item._isNew) return { [wrapper]: value, ...visibility };
    if (baseline.get(item.editor_key) === JSON.stringify(value)) return { editor_key: item.editor_key, ...visibility };
    return { editor_key: item.editor_key, [wrapper]: value, ...visibility };
  });
}

function toPublicationSavePayload(items, baselineItems) {
  const baseline = new Map(
    (baselineItems || [])
      .filter((item) => !item._isNew && item.editor_key)
      .map((item) => [item.editor_key, JSON.stringify(relationshipComparablePublication(item))]),
  );
  return sortedPublications(items).map((item) => {
    if (item._isNew) return { publication: toPublicationPayload(item), visible_in_CV: item.visible_in_CV !== false };
    if (baseline.get(item.editor_key) === JSON.stringify(relationshipComparablePublication(item))) {
      return { editor_key: item.editor_key, visible_in_CV: item.visible_in_CV !== false };
    }
    return { editor_key: item.editor_key, publication: toPublicationPayload(item), visible_in_CV: item.visible_in_CV !== false };
  });
}

function peopleComparable(items) {
  return JSON.stringify((items || []).map((item) => ({
    identity: item._isNew ? `new:${item._clientKey}` : `existing:${item.editor_key}`,
    person: toPersonPayload(item),
  })));
}

function awardsComparable(items) {
  return JSON.stringify(sortedAwards(items).map((item) => ({
    identity: item._isNew ? `new:${item._clientKey}` : `existing:${item.editor_key}`,
    award: toAwardPayload(item), visible_in_CV: item.visible_in_CV !== false,
  })));
}

function toAcademicActivityPayload(item) {
  return {
    category: item.category || "editorial_service",
    journal_en: item.journal_en || "",
    journal_ko: item.journal_ko || "",
    completed_dates: (item.completed_date_rows || []).map((row) => row.value || ""),
    event_en: item.event_en || "",
    event_ko: item.event_ko || "",
    topic_en: item.topic_en || "",
    topic_ko: item.topic_ko || "",
    conference_en: item.conference_en || "",
    conference_ko: item.conference_ko || "",
    organization_en: item.organization_en || "",
    organization_ko: item.organization_ko || "",
    role_en: item.role_en || "",
    role_ko: item.role_ko || "",
    date: item.date || "",
    start_date: item.start_date || "",
    end_date: item.end_date || "",
  };
}

function academicActivitiesSourceComparable(items) {
  return JSON.stringify((items || []).map((item) => ({
    identity: item._isNew ? `new:${item._clientKey}` : `existing:${item.editor_key}`,
    activity: toAcademicActivityPayload(item),
  })));
}

function academicActivitiesComparable(items) {
  return JSON.stringify((items || []).map((item) => ({
    identity: item._isNew ? `new:${item._clientKey}` : `existing:${item.editor_key}`,
    activity: toAcademicActivityPayload(item),
    visible_in_CV: item.visible_in_CV !== false,
  })));
}

function toAcademicActivitiesSavePayload(items, baselineItems) {
  const baseline = new Map(
    (baselineItems || [])
      .filter((item) => !item._isNew && item.editor_key)
      .map((item) => [item.editor_key, JSON.stringify(toAcademicActivityPayload(item))]),
  );
  return (items || []).map((item) => {
    const activity = toAcademicActivityPayload(item);
    const visibility = { visible_in_CV: item.visible_in_CV !== false };
    if (item._isNew) return { activity, ...visibility };
    if (baseline.get(item.editor_key) === JSON.stringify(activity)) return { editor_key: item.editor_key, ...visibility };
    return { editor_key: item.editor_key, activity, ...visibility };
  });
}

function publicationsComparable(items) {
  return JSON.stringify(sortedPublications(items).map((item) => ({
    identity: item._isNew ? `new:${item._clientKey}` : `existing:${item.editor_key}`,
    publication: relationshipComparablePublication(item), visible_in_CV: item.visible_in_CV !== false,
  })));
}

function randomRelationshipID(prefix) {
  const uuid = globalThis.crypto?.randomUUID?.()
    || `${Date.now().toString(36)}-${Math.random().toString(36).slice(2)}-${Math.random().toString(36).slice(2)}`;
  return `${prefix}-${uuid.toLocaleLowerCase()}`.replace(/[^a-z0-9-]+/g, "-").replace(/-+/g, "-");
}

function canonicalSlug(value) {
  return String(value || "")
    .normalize("NFKD")
    .replace(/[\u0300-\u036f]/g, "")
    .toLocaleLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "")
    .slice(0, 80)
    .replace(/-+$/g, "");
}

function assignPendingIDs(items, prefix, label) {
  const used = new Set((items || []).map((item) => item.id).filter(Boolean));
  (items || []).filter((item) => item._isNew && !item.id).forEach((item) => {
    const base = canonicalSlug(label(item)) || randomRelationshipID(prefix);
    let candidate = base;
    for (let suffix = 2; used.has(candidate); suffix += 1) candidate = `${base}-${suffix}`;
    item.id = candidate;
    used.add(candidate);
  });
}

function assignPendingRelationshipIDs() {
  assignPendingIDs(state.peopleDraft, "person", (item) => item.name_en);
  assignPendingIDs(state.awardsDraft, "award", (item) => item.title_en);
}

function toContentMediaPayload(media) {
  const reference = media.stage_token
    ? { stage_token: media.stage_token }
    : { editor_key: media.editor_key };
  return {
    ...reference,
    caption_en: media.caption_en || "",
    caption_ko: media.caption_ko || "",
  };
}

function toProjectPayload(item) {
  return {
    start_date: item.start_date || "",
    end_date: item.end_date || "",
    title_en: item.title_en || "",
    title_ko: item.title_ko || "",
    theme: item.theme || "",
    funder_en: item.funder_en || "",
    funder_ko: item.funder_ko || "",
    notes_en: (item.notes_en_rows || []).map((note) => note.value || ""),
    notes_kr: (item.notes_kr_rows || []).map((note) => note.value || ""),
    media: (item.media || []).map(toContentMediaPayload),
  };
}

function toSoftwareLinkPayload(link) {
  return {
    ...(link.editor_key ? { editor_key: link.editor_key } : {}),
    url: link.url || "",
    label: link.label || "",
    label_en: link.label_en || "",
    label_ko: link.label_ko || "",
  };
}

function toSoftwarePayload(item) {
  return {
    name: item.name || "",
    stage: item.stage || "development",
    links: (item.links || []).map(toSoftwareLinkPayload),
    notes_en: (item.notes_en_rows || []).map((note) => note.value || ""),
    notes_kr: (item.notes_kr_rows || []).map((note) => note.value || ""),
    media: (item.media || []).map(toContentMediaPayload),
    technologies: (item.technologies || []).map((technology) => technology.value || ""),
  };
}

function toProjectsSavePayload(items, baselineItems) {
  const baseline = new Map(
    (baselineItems || [])
      .filter((item) => !item._isNew && item.editor_key)
      .map((item) => [item.editor_key, JSON.stringify(toProjectPayload(item))]),
  );
  return (items || []).map((item) => {
    const project = toProjectPayload(item);
    if (item._isNew) return { project, visible_in_CV: item.visible_in_CV !== false };
    if (baseline.get(item.editor_key) === JSON.stringify(project)) return { editor_key: item.editor_key, visible_in_CV: item.visible_in_CV !== false };
    return { editor_key: item.editor_key, project, visible_in_CV: item.visible_in_CV !== false };
  });
}

function toSoftwareSavePayload(items, baselineItems) {
  const baseline = new Map(
    (baselineItems || [])
      .filter((item) => !item._isNew && item.editor_key)
      .map((item) => [item.editor_key, JSON.stringify(toSoftwarePayload(item))]),
  );
  return (items || []).map((item) => {
    const software = toSoftwarePayload(item);
    if (item._isNew) return { software, visible_in_CV: item.visible_in_CV !== false };
    if (baseline.get(item.editor_key) === JSON.stringify(software)) return { editor_key: item.editor_key, visible_in_CV: item.visible_in_CV !== false };
    return { editor_key: item.editor_key, software, visible_in_CV: item.visible_in_CV !== false };
  });
}

function projectsComparable(items) {
  return JSON.stringify((items || []).map((item) => ({
    identity: item._isNew ? `new:${item._clientKey}` : `existing:${item.editor_key}`,
    project: toProjectPayload(item), visible_in_CV: item.visible_in_CV !== false,
  })));
}

function softwareComparable(items) {
  return JSON.stringify((items || []).map((item) => ({
    identity: item._isNew ? `new:${item._clientKey}` : `existing:${item.editor_key}`,
    software: toSoftwarePayload(item), visible_in_CV: item.visible_in_CV !== false,
  })));
}

function syncNativeDirty(dirty) {
  if (!state.bridge) return;
  state.dirtySync = state.dirtySync
    .catch(() => {})
    .then(() => state.bridge.SetDirty(dirty))
    .catch(() => {});
}

function updateDirtyState() {
  cancelSaveCompletionStatus();
  if (
    !state.draft
    || !state.baseline
    || !state.profileDraft
    || !state.profileBaseline
    || !state.boardDraft
    || !state.boardBaseline
    || !state.projectsDraft
    || !state.projectsBaseline
    || !state.softwareDraft
    || !state.softwareBaseline
    || !state.peopleDraft
    || !state.peopleBaseline
    || !state.awardsDraft
    || !state.awardsBaseline
    || !state.academicActivitiesDraft
    || !state.academicActivitiesBaseline
    || !state.publicationsDraft
    || !state.publicationsBaseline
  ) return;
  const nextDirty = comparable(state.draft) !== comparable(state.baseline)
    || profileComparable(state.profileDraft, state.profileMediaDraft)
      !== profileComparable(state.profileBaseline, state.profileMediaBaseline)
    || boardComparable(state.boardDraft) !== boardComparable(state.boardBaseline)
    || projectsComparable(state.projectsDraft) !== projectsComparable(state.projectsBaseline)
    || softwareComparable(state.softwareDraft) !== softwareComparable(state.softwareBaseline)
    || peopleComparable(state.peopleDraft) !== peopleComparable(state.peopleBaseline)
    || awardsComparable(state.awardsDraft) !== awardsComparable(state.awardsBaseline)
    || academicActivitiesComparable(state.academicActivitiesDraft)
      !== academicActivitiesComparable(state.academicActivitiesBaseline)
    || publicationsComparable(state.publicationsDraft) !== publicationsComparable(state.publicationsBaseline);
  if (state.dirty !== nextDirty) {
    state.dirty = nextDirty;
    syncNativeDirty(nextDirty);
  }
  saveError.hidden = true;
  validateDraft();
  updateToolbar();
}

function settingsSubsetComparable(settings, keys) {
  if (!settings) return "";
  return JSON.stringify(Object.fromEntries(keys.map((key) => [key, settings[key]])));
}

function profileSourceComparable(profile, mediaDraft) {
  const payload = toProfilePayload(profile);
  const removeCVVisibility = (value) => {
    if (Array.isArray(value)) {
      value.forEach(removeCVVisibility);
      return;
    }
    if (!value || typeof value !== "object") return;
    delete value.visible_in_CV;
    Object.values(value).forEach(removeCVVisibility);
  };
  removeCVVisibility(payload);
  return JSON.stringify({
    profile: payload,
    media_stage_token: mediaDraft?.stage_token || "",
    remove_media: mediaDraft?.remove === true,
  });
}

function identifiedCollectionComparable(items, payloadFor, sorter = (values) => values || []) {
  return JSON.stringify(sorter(items || []).map((item) => ({
    identity: item._isNew ? `new:${item._clientKey}` : `existing:${item.editor_key}`,
    value: payloadFor(item),
  })));
}

function cvVisibilityComparable(settings, profile, projects, software, awards, academicActivities, publications) {
  const entries = [];
  ["experience", "education", "teaching", "scholarships", "certifications", "patents", "skills"]
    .forEach((section) => (profile?.[section] || []).forEach((item, index) => {
      entries.push([`profile:${section}:${item._clientKey || index}`, item.visible_in_CV !== false]);
    }));
  [
    ["projects", projects],
    ["software", software],
    ["awards", awards],
    ["academic_activities", academicActivities],
    ["publications", publications],
  ].forEach(([collection, items]) => (items || []).forEach((item, index) => {
    entries.push([`${collection}:${item._clientKey || item.editor_key || index}`, item.visible_in_CV !== false]);
  }));
  return JSON.stringify({
    sections: settingsSubsetComparable(settings, ["cv_sections", "hidden_cv_sections", "hidden_web_pages"]),
    items: entries,
  });
}

function menuDirtyMap() {
  const empty = Object.fromEntries(tabs.map((tab) => [tab.dataset.panel, false]));
  if (!state.draft || !state.baseline) return empty;
  return {
    ...empty,
    settings: settingsSubsetComparable(state.draft, ["color_theme", "custom_themes", "editor_shortcuts"])
      !== settingsSubsetComparable(state.baseline, ["color_theme", "custom_themes", "editor_shortcuts"]),
    journals: JSON.stringify(catalogPayload(state.draft.journals)) !== JSON.stringify(catalogPayload(state.baseline.journals)),
    conferences: JSON.stringify(catalogPayload(state.draft.conferences)) !== JSON.stringify(catalogPayload(state.baseline.conferences)),
    profile: profileSourceComparable(state.profileDraft, state.profileMediaDraft)
        !== profileSourceComparable(state.profileBaseline, state.profileMediaBaseline)
      || settingsSubsetComparable(state.draft, ["main_page_sections", "hidden_main_page_sections"])
        !== settingsSubsetComparable(state.baseline, ["main_page_sections", "hidden_main_page_sections"]),
    projects: identifiedCollectionComparable(state.projectsDraft, toProjectPayload)
      !== identifiedCollectionComparable(state.projectsBaseline, toProjectPayload),
    topics: settingsSubsetComparable(state.draft, [
      "topics", "topic_fallback",
    ]) !== settingsSubsetComparable(state.baseline, [
      "topics", "topic_fallback",
    ]),
    software: identifiedCollectionComparable(state.softwareDraft, toSoftwarePayload)
      !== identifiedCollectionComparable(state.softwareBaseline, toSoftwarePayload),
    people: peopleComparable(state.peopleDraft) !== peopleComparable(state.peopleBaseline),
    awards: identifiedCollectionComparable(state.awardsDraft, toAwardPayload, sortedAwards)
      !== identifiedCollectionComparable(state.awardsBaseline, toAwardPayload, sortedAwards),
    "academic-activities": academicActivitiesSourceComparable(state.academicActivitiesDraft)
      !== academicActivitiesSourceComparable(state.academicActivitiesBaseline),
    publications: identifiedCollectionComparable(
      state.publicationsDraft,
      relationshipComparablePublication,
      sortedPublications,
    ) !== identifiedCollectionComparable(
      state.publicationsBaseline,
      relationshipComparablePublication,
      sortedPublications,
    ),
    board: boardComparable(state.boardDraft) !== boardComparable(state.boardBaseline),
    cv: cvVisibilityComparable(
      state.draft,
      state.profileDraft,
      state.projectsDraft,
      state.softwareDraft,
      state.awardsDraft,
      state.academicActivitiesDraft,
      state.publicationsDraft,
    ) !== cvVisibilityComparable(
      state.baseline,
      state.profileBaseline,
      state.projectsBaseline,
      state.softwareBaseline,
      state.awardsBaseline,
      state.academicActivitiesBaseline,
      state.publicationsBaseline,
    ),
  };
}

function menuErrorCounts() {
  const result = Object.fromEntries(tabs.map((tab) => [tab.dataset.panel, 0]));
  const panelForPrefix = {
    settings: "settings",
    journals: "journals",
    conferences: "conferences",
    profile: "profile",
    topics: "topics",
    projects: "projects",
    software: "software",
    people: "people",
    awards: "awards",
    academicActivities: "academic-activities",
    publications: "publications",
    board: "board",
  };
  state.validationErrors.forEach((_message, key) => {
    const panel = panelForPrefix[String(key).split("|", 1)[0]];
    if (panel) result[panel] += 1;
  });
  return result;
}

function updateNavigationStatus() {
  const dirtyByPanel = menuDirtyMap();
  const errorsByPanel = menuErrorCounts();
  tabs.forEach((tab) => {
    const panel = tab.dataset.panel;
    const dirty = Boolean(dirtyByPanel[panel]);
    const errorCount = errorsByPanel[panel] || 0;
    const dirtyDot = tab.querySelector(".nav-dirty-dot");
    const errorBadge = tab.querySelector(".nav-error-badge");
    if (dirtyDot) dirtyDot.hidden = !dirty;
    if (errorBadge) {
      errorBadge.hidden = errorCount === 0;
      errorBadge.textContent = errorCount ? String(errorCount) : "";
    }
    const label = tab.querySelector(".nav-item-label")?.textContent?.trim() || tab.textContent.trim();
    const states = [];
    if (dirty) states.push("저장되지 않은 변경");
    if (errorCount) states.push(`입력 오류 ${errorCount}개`);
    tab.setAttribute("aria-label", [label, ...states].join(", "));
  });
}

function compareUpdateVersions(left, right) {
  const pattern = /^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$/;
  const first = typeof left === "string" ? left.match(pattern) : null;
  const second = typeof right === "string" ? right.match(pattern) : null;
  if (!first || !second || first[0] !== left || second[0] !== right) return null;
  for (let index = 1; index <= 3; index += 1) {
    if (first[index] === second[index]) continue;
    return first[index].length !== second[index].length
      ? (first[index].length > second[index].length ? 1 : -1)
      : (first[index] > second[index] ? 1 : -1);
  }
  return 0;
}

function editorUpdateBusy() {
  return state.loading || state.saving || state.discarding || state.boardDropBusy
    || state.editorUpdate.infoLoading || state.editorUpdate.checking || state.editorUpdate.installing;
}

function updateEditorUpdateControls() {
  const check = document.querySelector("#editor-update-check");
  const install = document.querySelector("#editor-update-install");
  if (check) check.disabled = editorUpdateBusy() || typeof state.bridge?.CheckUpdates !== "function";
  if (install) install.disabled = editorUpdateBusy() || !state.editorUpdate.available || typeof state.bridge?.StartUpdate !== "function";
}

function renderEditorUpdate() {
  const update = state.editorUpdate;
  const info = update.info || {};
  const status = document.querySelector("#editor-update-status");
  if (!status) return;
  const labels = {
    initial: "확인 전", checking: "확인 중", available: "업데이트 가능", current: "업데이트 없음", failed: "확인 실패",
    installing: "업데이트 중", closing: "편집기 종료", restart: "재실행 필요", install_failed: "업데이트 실패",
  };
  status.textContent = labels[update.phase] || labels.initial;
  status.dataset.kind = ["failed", "install_failed"].includes(update.phase) ? "error" : "";
  status.closest(".editor-update-card")?.setAttribute("aria-busy", String(update.checking || update.infoLoading || update.installing));
  document.querySelector("#editor-update-current").textContent = info.current_version || "—";
  document.querySelector("#editor-update-latest").textContent = update.latestVersion || "—";
  const source = document.querySelector("#editor-update-source");
  const link = document.querySelector("#editor-update-source-link");
  const sourceText = [info.source_repository, info.source_ref].filter(Boolean).join(" · ") || "—";
  source.textContent = sourceText;
  link.textContent = sourceText;
  const safeURL = typeof info.source_url === "string" && isHTTPURL(info.source_url) ? info.source_url : "";
  source.hidden = Boolean(safeURL);
  link.hidden = !safeURL;
  if (safeURL) link.href = safeURL;
  else link.removeAttribute("href");
  const error = document.querySelector("#editor-update-error");
  error.textContent = update.error;
  error.hidden = !update.error;
  const recovery = document.querySelector("#editor-update-recovery");
  const recoveryStatus = ["complete", "rolled_back", "failed"].includes(info.recovery_status) ? info.recovery_status : "";
  recovery.textContent = recoveryStatus ? info.recovery_message || "" : "";
  recovery.dataset.kind = recoveryStatus;
  recovery.hidden = !recovery.textContent;
  const history = document.querySelector("#editor-update-history");
  const list = document.querySelector("#editor-update-notes");
  list.replaceChildren();
  update.releases.forEach((release) => {
    const item = document.createElement("li");
    item.className = "editor-update-release";
    const heading = document.createElement("div");
    heading.className = "editor-update-release-heading";
    const version = document.createElement("h5");
    version.textContent = release.version;
    heading.append(version);
    if (release.date) {
      const date = document.createElement("time");
      date.textContent = release.date;
      date.dateTime = release.date;
      heading.append(date);
    }
    const notes = document.createElement("ul");
    (Array.isArray(release.notes) ? release.notes : []).forEach((note) => {
      const row = document.createElement("li");
      row.textContent = String(note);
      notes.append(row);
    });
    item.append(heading, notes);
    list.append(item);
  });
  history.hidden = update.releases.length === 0;
  updateEditorUpdateControls();
}

async function loadEditorUpdateInfo() {
  const update = state.editorUpdate;
  if (update.info || update.infoLoading || typeof state.bridge?.GetUpdateInfo !== "function") return;
  update.infoLoading = true;
  renderEditorUpdate();
  try {
    const info = await state.bridge.GetUpdateInfo();
    if (!info || compareUpdateVersions(info.current_version, info.current_version) === null) {
      throw new Error("현재 버전 정보를 확인할 수 없습니다.");
    }
    update.info = info;
    update.error = "";
    update.phase = "initial";
  } catch (error) {
    update.phase = "failed";
    update.error = errorMessage(error);
  } finally {
    update.infoLoading = false;
    renderEditorUpdate();
  }
}

async function checkEditorUpdates() {
  const update = state.editorUpdate;
  if (editorUpdateBusy() || typeof state.bridge?.CheckUpdates !== "function") return false;
  update.checking = true;
  update.phase = "checking";
  update.error = "";
  update.latestVersion = "";
  update.available = false;
  update.releases = [];
  renderEditorUpdate();
  try {
    const result = await state.bridge.CheckUpdates();
    const comparison = compareUpdateVersions(result?.current_version, result?.latest_version);
    if (comparison === null) throw new Error("현재 버전과 원본 버전 정보를 확인할 수 없습니다.");
    update.info = { ...update.info, ...result };
    update.latestVersion = result.latest_version;
    update.available = result.update_available === true && comparison < 0;
    if (update.available) {
      const releases = new Map();
      (Array.isArray(result.releases) ? result.releases : []).forEach((release) => {
        const afterCurrent = compareUpdateVersions(release?.version, result.current_version);
        const beforeLatest = compareUpdateVersions(release?.version, result.latest_version);
        if (afterCurrent !== null && beforeLatest !== null && afterCurrent > 0 && beforeLatest <= 0) releases.set(release.version, release);
      });
      update.releases = [...releases.values()].sort((first, second) => compareUpdateVersions(second.version, first.version));
    }
    update.phase = update.available ? "available" : "current";
    return true;
  } catch (error) {
    update.latestVersion = "";
    update.available = false;
    update.releases = [];
    update.phase = "failed";
    update.error = errorMessage(error);
    return false;
  } finally {
    update.checking = false;
    renderEditorUpdate();
  }
}

async function startEditorUpdate() {
  const update = state.editorUpdate;
  if (editorUpdateBusy() || !update.available || typeof state.bridge?.StartUpdate !== "function") return false;
  const expectedVersion = update.latestVersion;
  if (!expectedVersion) return false;
  cancelSaveCompletionStatus();
  update.installing = true;
  update.phase = "installing";
  update.available = false;
  update.error = "";
  state.saving = true;
  beginOperation("update");
  renderAll();
  renderEditorUpdate();
  updateToolbar();
  let closing = false;
  let failureTitle = "업데이트 실패";
  try {
    if (state.dirty) {
      failureTitle = "저장 실패";
      updateOperationProgress("변경 내용을 저장합니다", -1);
      if (!await saveChanges({ withinOperation: true })) {
        throw new Error(state.validationErrors.size ? "입력 오류를 확인해 주세요." : saveErrorMessage.textContent || "변경 내용을 저장하지 못했습니다.");
      }
      failureTitle = "업데이트 실패";
    }
    await state.stagingOps;
    await state.dirtySync;
    updateOperationProgress("업데이트를 요청합니다", -1);
    const response = await state.bridge.StartUpdate({
      operation_id: state.operation.id,
      expected_version: expectedVersion,
      expected_commit: "",
    });
    if (response?.target_version !== expectedVersion) throw new Error("업데이트 결과 버전이 확인한 버전과 다릅니다.");
    if (response.restart_required === true) {
      closing = true;
      update.phase = "closing";
      updateOperationProgress("편집기를 종료합니다", -1);
    } else {
      update.phase = "restart";
      finishOperation(true, response.message || "편집기를 다시 실행해 적용 결과를 확인하세요.");
    }
    return true;
  } catch (error) {
    update.available = true;
    update.phase = "install_failed";
    update.error = errorMessage(error);
    finishOperation(false, failureTitle, update.error);
    return false;
  } finally {
    if (!closing) {
      update.installing = false;
      state.saving = false;
      renderAll();
    }
    renderEditorUpdate();
    updateToolbar();
  }
}

function updateToolbar() {
  const invalid = state.validationErrors.size > 0;
  const locked = state.loading || state.saving || state.discarding || state.boardDropBusy;
  if (locked) closeTaxonomyColorPopover();
  saveButton.disabled = locked || !state.dirty;
  discardButton.disabled = locked || !state.dirty;
  citationUpdateButton.disabled = locked || !state.draft || !state.bridge?.UpdateCitations;
  pushButton.disabled = locked || !state.draft || !state.bridge?.PushGit;
  document.querySelectorAll("[data-editor-control]").forEach((control) => {
    control.disabled = locked || (!state.draft && control.dataset.siteColorTheme !== undefined);
  });
  const addProjectButton = document.querySelector("#add-project-button");
  const addSoftwareButton = document.querySelector("#add-software-button");
  const addPersonButton = document.querySelector("#add-person-button");
  const addAwardButton = document.querySelector("#add-award-button");
  const addPublicationButton = document.querySelector("#add-publication-button");
  if (addProjectButton) addProjectButton.disabled = locked || (state.projectsDraft?.length || 0) >= 500;
  if (addSoftwareButton) addSoftwareButton.disabled = locked || (state.softwareDraft?.length || 0) >= 500;
  if (addPersonButton) addPersonButton.disabled = locked || (state.peopleDraft?.length || 0) >= 500;
  if (addAwardButton) addAwardButton.disabled = locked || (state.awardsDraft?.length || 0) >= 500;
  document.querySelectorAll("[data-academic-activity-add]").forEach((button) => {
    button.disabled = locked || (state.academicActivitiesDraft?.length || 0) >= 5000;
  });
  if (addPublicationButton) addPublicationButton.disabled = locked || (state.publicationsDraft?.length || 0) >= 1000;
  updateEditorUpdateControls();
  updateNavigationStatus();

  if (state.loading) {
    setStatus("데이터를 불러오는 중…");
  } else if (state.saving) {
    setStatus(state.operation?.kind === "update" ? "업데이트 중…" : state.operation?.kind === "citations" ? "인용수 갱신 중…" : state.operation?.kind === "push" ? "Git push 중…" : "저장 중…");
  } else if (state.discarding) {
    setStatus("변경을 취소하는 중…");
  } else if (state.boardDropBusy) {
    setStatus("미디어 처리 중…");
  } else if (!state.draft) {
    // The load error shown in the workspace already contains the details.
  } else if (invalid) {
    setStatus(`입력 오류 ${state.validationErrors.size}개`, "error");
  } else if (state.dirty) {
    setStatus("저장되지 않은 변경", "dirty");
  } else if (state.operation?.kind === "push" && state.operation.done) {
    setStatus(state.operation.success ? "Git push 완료" : "Git push 실패", state.operation.success ? "success" : "error");
  } else if (state.operation?.kind === "citations" && state.operation.done) {
    setStatus(state.operation.success ? "인용수 갱신 완료" : "인용수 갱신 실패", state.operation.success ? "success" : "error");
  } else if (state.saveCompletionVisible) {
    setStatus("저장 완료", "success");
  } else {
    setStatus("저장됨", "success");
  }
}

function validationKey(kind, clientKey, field) {
  return `${kind}|${clientKey}|${field}`;
}

function fallbackValidationKey(kind, field) {
  return `${kind}|fallback|${field}`;
}

function validateLabelValue(value, key, context) {
  const trimmed = value.trim();
  if (!trimmed) {
    state.validationErrors.set(key, `${context}을(를) 입력해 주세요.`);
  } else if ([...trimmed].length > 100) {
    state.validationErrors.set(key, `${context}은(는) 100자 이하여야 합니다.`);
  }
}

function validateTaxonomy(kind) {
  const config = taxonomyConfig[kind];
  const items = state.draft[config.listKey];
  const fields = [
    ["label_en", "영문 이름"],
    ["label_ko", "국문 이름"],
  ];

  fields.forEach(([field, fieldLabel]) => {
    const values = new Map();
    items.forEach((item) => {
      const key = validationKey(kind, item._clientKey, field);
      validateLabelValue(item[field] || "", key, fieldLabel);
      const normalised = (item[field] || "").trim().toLocaleLowerCase();
      if (!normalised) return;
      const matches = values.get(normalised) || [];
      matches.push(item);
      values.set(normalised, matches);
    });
    values.forEach((matches) => {
      if (matches.length < 2) return;
      matches.forEach((item) => {
        state.validationErrors.set(
          validationKey(kind, item._clientKey, field),
          `${fieldLabel}은(는) 중복될 수 없습니다.`,
        );
      });
    });
  });

  items.forEach((item) => {
    if (item.color && !/^#[0-9a-f]{6}$/i.test(item.color.trim())) {
      state.validationErrors.set(
        validationKey(kind, item._clientKey, "color"),
        "색상은 #RRGGBB 형식으로 입력해 주세요.",
      );
    }
  });

  const fallback = state.draft[config.fallbackKey];
  validateLabelValue(fallback.label_en || "", fallbackValidationKey(kind, "label_en"), "기타 분류 영문 이름");
  validateLabelValue(fallback.label_ko || "", fallbackValidationKey(kind, "label_ko"), "기타 분류 국문 이름");
}

function boardValidationKey(clientKey, field) {
  return `board|${clientKey}|${field}`;
}

function isCanonicalDate(value) {
  if (!/^\d{4}-\d{2}-\d{2}$/.test(value)) return false;
  const [year, month, day] = value.split("-").map(Number);
  const date = new Date(Date.UTC(year, month - 1, day));
  return date.getUTCFullYear() === year
    && date.getUTCMonth() === month - 1
    && date.getUTCDate() === day;
}

function validateBoardText(value, key, label, maximum, required = false) {
  const trimmed = String(value || "").trim();
  if (required && !trimmed) {
    state.validationErrors.set(key, `${label}을(를) 입력해 주세요.`);
  } else if ([...trimmed].length > maximum) {
    state.validationErrors.set(key, `${label}은(는) ${maximum}자 이하여야 합니다.`);
  }
}

function validateBoard() {
  state.boardDraft.forEach((item) => {
    const startKey = boardValidationKey(item._clientKey, "start_date");
    const endKey = boardValidationKey(item._clientKey, "end_date");
    const startDate = String(item.start_date || "").trim();
    const endDate = String(item.end_date || "").trim();
    if (!isCanonicalDate(startDate)) {
      state.validationErrors.set(startKey, "시작일을 입력해 주세요.");
    }
    if (endDate && !isCanonicalDate(endDate)) {
      state.validationErrors.set(endKey, "종료일 형식을 확인해 주세요.");
    } else if (isCanonicalDate(startDate) && endDate && startDate > endDate) {
      state.validationErrors.set(endKey, "종료일은 시작일보다 빠를 수 없습니다.");
    }
    const titleENKey = boardValidationKey(item._clientKey, "title_en");
    const titleKOKey = boardValidationKey(item._clientKey, "title_ko");
    validateBoardText(item.title_en, titleENKey, "영문 제목", 300);
    validateBoardText(item.title_ko, titleKOKey, "국문 제목", 300);
    if (!String(item.title_en || "").trim() && !String(item.title_ko || "").trim()) {
      const message = "영문 또는 국문 제목을 입력해 주세요.";
      state.validationErrors.set(titleENKey, message);
      state.validationErrors.set(titleKOKey, message);
    }
    validateBoardText(item.content_en, boardValidationKey(item._clientKey, "content_en"), "영문 본문", 20000);
    validateBoardText(item.content_ko, boardValidationKey(item._clientKey, "content_ko"), "국문 본문", 20000);
    item.media.forEach((media) => {
      validateBoardText(
        media.caption_en,
        boardValidationKey(item._clientKey, `media:${media._clientKey}:caption_en`),
        "영문 사진 설명",
        500,
      );
      validateBoardText(
        media.caption_ko,
        boardValidationKey(item._clientKey, `media:${media._clientKey}:caption_ko`),
        "국문 사진 설명",
        500,
      );
    });
  });
}

function contentValidationKey(collection, clientKey, field) {
  return `${collection}|${clientKey}|${field}`;
}

function validateContentText(collection, item, field, label, maximum, required = false) {
  validateBoardText(
    item[field],
    contentValidationKey(collection, item._clientKey, field),
    label,
    maximum,
    required,
  );
}

function validateContentNotes(collection, item) {
  [["en", "영문"], ["kr", "국문"]].forEach(([language, label]) => {
    const rows = item[`notes_${language}_rows`] || [];
    if (!rows.length || rows.length > 100) {
      state.validationErrors.set(
        contentValidationKey(collection, item._clientKey, `notes_${language}`),
        `${label} 설명은 1개 이상 100개 이하로 입력해 주세요.`,
      );
    }
    rows.forEach((note) => validateBoardText(
      note.value,
      contentValidationKey(collection, item._clientKey, `note:${note._clientKey}:${language}`),
      `${label} 설명`,
      5000,
      true,
    ));
  });
}

function validateProjects() {
  const knownThemes = new Set(
    (state.draft?.topics || []).map((item) => item.id).filter(Boolean),
  );
  (state.projectsDraft || []).forEach((item) => {
    const startKey = contentValidationKey("projects", item._clientKey, "start_date");
    const endKey = contentValidationKey("projects", item._clientKey, "end_date");
    const startDate = String(item.start_date || "").trim();
    const endDate = String(item.end_date || "").trim();
    if (!isCanonicalDate(startDate)) state.validationErrors.set(startKey, "시작일을 입력해 주세요.");
    if (!isCanonicalDate(endDate)) {
      state.validationErrors.set(endKey, "종료일을 입력해 주세요.");
    } else if (isCanonicalDate(startDate) && startDate > endDate) {
      state.validationErrors.set(endKey, "종료일은 시작일보다 빠를 수 없습니다.");
    }

    validateContentText("projects", item, "title_en", "영문 제목", 500);
    validateContentText("projects", item, "title_ko", "국문 제목", 500);
    validateContentText("projects", item, "funder_en", "영문 지원기관", 500);
    validateContentText("projects", item, "funder_ko", "국문 지원기관", 500);
    if (!item.theme || !knownThemes.has(item.theme)) {
      state.validationErrors.set(
        contentValidationKey("projects", item._clientKey, "theme"),
        "저장된 프로젝트 테마를 선택해 주세요.",
      );
    }
    validateContentNotes("projects", item);
    (item.media || []).forEach((media) => {
      const englishKey = contentValidationKey("projects", item._clientKey, `media:${media._clientKey}:caption_en`);
      const koreanKey = contentValidationKey("projects", item._clientKey, `media:${media._clientKey}:caption_ko`);
      validateBoardText(media.caption_en, englishKey, "영문 사진 설명", 500);
      validateBoardText(media.caption_ko, koreanKey, "국문 사진 설명", 500);
      if (Boolean(String(media.caption_en || "").trim()) !== Boolean(String(media.caption_ko || "").trim())) {
        const message = "영문과 국문 사진 설명을 모두 입력하거나 모두 비워 주세요.";
        state.validationErrors.set(englishKey, message);
        state.validationErrors.set(koreanKey, message);
      }
    });
  });
}

function isHTTPURL(value) {
  try {
    const parsed = new URL(String(value || "").trim());
    return (parsed.protocol === "http:" || parsed.protocol === "https:") && Boolean(parsed.host);
  } catch (_) {
    return false;
  }
}

function validateSoftware() {
  const stages = new Set(["release", "preview", "development"]);
  (state.softwareDraft || []).forEach((item) => {
    validateContentText("software", item, "name", "이름", 300, true);
    if (!stages.has(item.stage)) {
      state.validationErrors.set(
        contentValidationKey("software", item._clientKey, "stage"),
        "단계를 선택해 주세요.",
      );
    }
    validateContentNotes("software", item);
    (item.links || []).forEach((link) => {
      const key = contentValidationKey("software", item._clientKey, `link:${link._clientKey}:url`);
      if (!isHTTPURL(link.url)) state.validationErrors.set(key, "http 또는 https 주소를 입력해 주세요.");
      [["label", "링크 이름"], ["label_en", "영문 링크 이름"], ["label_ko", "국문 링크 이름"]]
        .forEach(([field, label]) => validateBoardText(
          link[field],
          contentValidationKey("software", item._clientKey, `link:${link._clientKey}:${field}`),
          label,
          300,
        ));
    });
    if (!(item.technologies || []).length) {
      state.validationErrors.set(
        contentValidationKey("software", item._clientKey, "technologies"),
        "기술을 하나 이상 선택해 주세요.",
      );
    }
    (item.technologies || []).forEach((technology) => {
      const key = contentValidationKey("software", item._clientKey, `technology:${technology._clientKey}`);
      if (!String(technology.value || "").trim()) {
        state.validationErrors.set(key, "기술을 선택해 주세요.");
      } else {
        validateBoardText(technology.value, key, "기술", 200);
      }
    });
    (item.media || []).forEach((media) => {
      validateBoardText(
        media.caption_en,
        contentValidationKey("software", item._clientKey, `media:${media._clientKey}:caption_en`),
        "영문 사진 설명",
        500,
        true,
      );
      validateBoardText(
        media.caption_ko,
        contentValidationKey("software", item._clientKey, `media:${media._clientKey}:caption_ko`),
        "국문 사진 설명",
        500,
        true,
      );
    });
  });
}

function validateEntityText(collection, item, field, label, maximum, required = false) {
  validateBoardText(
    item[field],
    contentValidationKey(collection, item._clientKey, field),
    label,
    maximum,
    required,
  );
}

function validateStringRows(collection, item, field, label, maximum) {
  (item[field] || []).forEach((row) => validateBoardText(
    row.value,
    contentValidationKey(collection, item._clientKey, `${field}:${row._clientKey}`),
    label,
    maximum,
    true,
  ));
}

function validatePeople() {
  const selfPeople = (state.peopleDraft || []).filter((item) => item.is_self);
  (state.peopleDraft || []).forEach((item) => {
    const englishKey = contentValidationKey("people", item._clientKey, "name_en");
    const koreanKey = contentValidationKey("people", item._clientKey, "name_ko");
    validateEntityText("people", item, "name_en", "영문 이름", 300);
    validateEntityText("people", item, "name_ko", "국문 이름", 300);
    if (!String(item.name_en || "").trim() && !String(item.name_ko || "").trim()) {
      const message = "영문 또는 국문 이름을 입력해 주세요.";
      state.validationErrors.set(englishKey, message);
      state.validationErrors.set(koreanKey, message);
    }
    validateStringRows("people", item, "notes_en_rows", "영문 소속 / 메모", 5000);
    validateStringRows("people", item, "notes_ko_rows", "국문 소속/메모", 5000);
    if (selfPeople.length !== 1) {
      state.validationErrors.set(
        contentValidationKey("people", item._clientKey, "is_self"),
        "본인을 한 명 선택해 주세요.",
      );
    }
  });
}

function validateAwards() {
  (state.awardsDraft || []).forEach((item) => {
    if (!validDocumentFilename(item.pdf)) state.validationErrors.set(contentValidationKey("awards", item._clientKey, "pdf"), "data/document 안의 PDF 파일명만 입력해 주세요.");
    const date = String(item.date || "").trim();
    if (!isCanonicalDate(date)) {
      state.validationErrors.set(contentValidationKey("awards", item._clientKey, "date"), "날짜를 입력해 주세요.");
    }
    [
      ["title_en", "영문 제목"],
      ["title_ko", "국문 제목"],
      ["organization_en", "영문 기관"],
      ["organization_ko", "국문 기관"],
    ].forEach(([field, label]) => validateEntityText("awards", item, field, label, 1000));
    [["title_en", "title_ko", "영문 또는 국문 제목을 입력해 주세요."],
      ["organization_en", "organization_ko", "영문 또는 국문 기관을 입력해 주세요."]]
      .forEach(([english, korean, message]) => {
        if (String(item[english] || "").trim() || String(item[korean] || "").trim()) return;
        state.validationErrors.set(contentValidationKey("awards", item._clientKey, english), message);
        state.validationErrors.set(contentValidationKey("awards", item._clientKey, korean), message);
      });
  });
}

function isPublicationDate(value) {
  if (!/^\d{4}(?:-\d{2}(?:-\d{2})?)?$/.test(value)) return false;
  if (value.length === 4) return Number(value) >= 1;
  const [year, month, day] = value.split("-").map(Number);
  if (month < 1 || month > 12) return false;
  if (value.length === 7) return true;
  const date = new Date(Date.UTC(year, month - 1, day));
  return date.getUTCFullYear() === year
    && date.getUTCMonth() === month - 1
    && date.getUTCDate() === day;
}

function validatePublications() {
  const types = new Set([
    "international-journal",
    "domestic-journal",
    "international-conference",
    "domestic-conference",
  ]);
  const topics = new Set((state.draft?.topics || []).map((item) => item.id).filter(Boolean));
  const peopleKeys = new Set((state.peopleDraft || []).map((item) => item._clientKey));
  const awardKeys = new Set((state.awardsDraft || []).map((item) => item._clientKey));
  (state.publicationsDraft || []).forEach((item) => {
    const englishKey = contentValidationKey("publications", item._clientKey, "title_en");
    const koreanKey = contentValidationKey("publications", item._clientKey, "title_ko");
    validateEntityText("publications", item, "title_en", "영문 제목", Number.POSITIVE_INFINITY);
    validateEntityText("publications", item, "title_ko", "국문 제목", Number.POSITIVE_INFINITY);
    if (!String(item.title_en || "").trim() && !String(item.title_ko || "").trim()) {
      const message = "영문 또는 국문 제목을 입력해 주세요.";
      state.validationErrors.set(englishKey, message);
      state.validationErrors.set(koreanKey, message);
    }
    validateEntityText("publications", item, "abstract_en", "영문 초록", Number.POSITIVE_INFINITY);
    validateEntityText("publications", item, "abstract_ko", "국문 초록", Number.POSITIVE_INFINITY);
    validateStringRows("publications", item, "keywords_en_rows", "영문 키워드", Number.POSITIVE_INFINITY);
    validateStringRows("publications", item, "keywords_ko_rows", "국문 키워드", Number.POSITIVE_INFINITY);
    const authorKeys = item.author_keys || [];
    if (!authorKeys.length) {
      state.validationErrors.set(
        contentValidationKey("publications", item._clientKey, "author_keys"),
        "저자를 한 명 이상 선택해 주세요.",
      );
    } else if (authorKeys.some((key) => !peopleKeys.has(key) && !item._unresolvedAuthorIDs?.[key])) {
      state.validationErrors.set(
        contentValidationKey("publications", item._clientKey, "author_keys"),
        "저자 연결을 확인해 주세요.",
      );
    }
    if (!types.has(item.publication_type)) {
      state.validationErrors.set(
        contentValidationKey("publications", item._clientKey, "publication_type"),
        "논문 유형을 선택해 주세요.",
      );
    }
    if (item.topic && !topics.has(item.topic)) {
      state.validationErrors.set(
        contentValidationKey("publications", item._clientKey, "topic"),
        "논문 주제를 선택해 주세요.",
      );
    }
    if (item.award_key && !awardKeys.has(item.award_key) && !item._unresolvedAwardID) {
      state.validationErrors.set(
        contentValidationKey("publications", item._clientKey, "award_key"),
        "수상 연결을 확인해 주세요.",
      );
    }
    if (!new Set(["published", "under_review", "in_press"]).has(item._status)) {
      state.validationErrors.set(
        contentValidationKey("publications", item._clientKey, "_status"),
        "논문 상태를 선택해 주세요.",
      );
    }
    publicationBibliographyFields.forEach(([field, label]) => {
      validateEntityText("publications", item, field, label, 500);
      const value = String(item[field] || "").trim();
      if (/[\u0000-\u001f\u007f-\u009f]/.test(value)) {
        state.validationErrors.set(contentValidationKey("publications", item._clientKey, field), `${label}에 제어문자를 사용할 수 없습니다.`);
      }
      if (value && (field === "online_date" || field === "issue_date") && !isPublicationDate(value)) {
        state.validationErrors.set(contentValidationKey("publications", item._clientKey, field), "실제 날짜를 YYYY / YYYY-MM / YYYY-MM-DD 형식으로 입력해 주세요.");
      }
    });
    [["journal_id", "journals", "-journal"], ["conference_id", "conferences", "-conference"]].forEach(([field, kind, suffix]) => {
      const id = item[field] || "";
      if (!id) return;
      if (!String(item.publication_type).endsWith(suffix) || !state.draft[kind].some((row) => row.id === id)) {
        state.validationErrors.set(contentValidationKey("publications", item._clientKey, field), "논문 유형에 맞는 목록에서 선택해 주세요.");
      }
    });
    validateEntityText("publications", item, "doi", "DOI", Number.POSITIVE_INFINITY);
    validateEntityText("publications", item, "url", "URL", Number.POSITIVE_INFINITY);
    validateEntityText("publications", item, "note", "메모", Number.POSITIVE_INFINITY);
  });
}

function academicActivityValidationKey(item, field) {
  return contentValidationKey("academicActivities", item._clientKey, field);
}

function validateAcademicActivityPair(item, englishField, koreanField, label) {
  validateBoardText(item[englishField], academicActivityValidationKey(item, englishField), `영문 ${label}`, 2000);
  validateBoardText(item[koreanField], academicActivityValidationKey(item, koreanField), `국문 ${label}`, 2000);
  if (String(item[englishField] || "").trim() || String(item[koreanField] || "").trim()) return;
  const message = `영문 또는 국문 ${label}을(를) 입력해 주세요.`;
  state.validationErrors.set(academicActivityValidationKey(item, englishField), message);
  state.validationErrors.set(academicActivityValidationKey(item, koreanField), message);
}

function validateAcademicActivityPeriod(item) {
  const startDate = String(item.start_date || "").trim();
  const endDate = String(item.end_date || "").trim();
  if (!isCanonicalDate(startDate)) {
    state.validationErrors.set(academicActivityValidationKey(item, "start_date"), "시작일을 입력해 주세요.");
  }
  if (endDate && !isCanonicalDate(endDate)) {
    state.validationErrors.set(academicActivityValidationKey(item, "end_date"), "종료일 형식을 확인해 주세요.");
  } else if (isCanonicalDate(startDate) && endDate && startDate > endDate) {
    state.validationErrors.set(academicActivityValidationKey(item, "end_date"), "종료일은 시작일보다 빠를 수 없습니다.");
  }
}

function validateAcademicActivities() {
  const categories = new Set(ACADEMIC_ACTIVITY_CATEGORIES.map(({ value }) => value));
  (state.academicActivitiesDraft || []).forEach((item) => {
    if (!categories.has(item.category)) {
      state.validationErrors.set(academicActivityValidationKey(item, "category"), "분류를 선택해 주세요.");
      return;
    }
    if (item.category === "reviews") {
      validateAcademicActivityPair(item, "journal_en", "journal_ko", "저널명");
      const rows = item.completed_date_rows || [];
      if (!rows.length) {
        state.validationErrors.set(academicActivityValidationKey(item, "completed_dates"), "완료일을 하나 이상 추가해 주세요.");
      }
      const seenDates = new Set();
      rows.forEach((row) => {
        const value = String(row.value || "").trim();
        const key = academicActivityValidationKey(item, `completed_dates:${row._clientKey}`);
        if (!isCanonicalDate(value)) {
          state.validationErrors.set(key, "실제 완료일을 입력해 주세요.");
        } else if (seenDates.has(value)) {
          state.validationErrors.set(key, "같은 완료일을 중복해서 입력할 수 없습니다.");
        }
        seenDates.add(value);
      });
      if (rows.length > 1000) {
        state.validationErrors.set(academicActivityValidationKey(item, "completed_dates"), "완료일은 최대 1000개까지 입력할 수 있습니다.");
      }
      return;
    }
    if (item.category === "invited_talks") {
      validateAcademicActivityPair(item, "event_en", "event_ko", "행사명");
      validateAcademicActivityPair(item, "topic_en", "topic_ko", "주제");
      if (!isCanonicalDate(String(item.date || "").trim())) {
        state.validationErrors.set(academicActivityValidationKey(item, "date"), "날짜를 입력해 주세요.");
      }
      return;
    }
    if (item.category === "conference_service") {
      validateAcademicActivityPair(item, "conference_en", "conference_ko", "학술대회명");
      validateAcademicActivityPair(item, "role_en", "role_ko", "역할");
      validateAcademicActivityPeriod(item);
      return;
    }
    if (item.category === "professional_service") {
      validateAcademicActivityPair(item, "organization_en", "organization_ko", "기관명");
      validateAcademicActivityPair(item, "role_en", "role_ko", "역할");
      validateAcademicActivityPeriod(item);
      return;
    }
    validateAcademicActivityPair(item, "journal_en", "journal_ko", "저널명");
    validateAcademicActivityPair(item, "role_en", "role_ko", "역할");
    validateAcademicActivityPeriod(item);
  });
}

function validateProfile() {
  (state.profileDraft?.patents || []).forEach((item, index) => {
    const error = (field, message) => state.validationErrors.set(profileValidationKey(["patents", index, field]), message);
    ["title_en", "title_ko", "applicant_en", "applicant_ko", "url",
      "application_number", "application_date", "publication_number", "publication_date", "registration_number", "registration_date"]
      .forEach((field) => {
        if (Object.hasOwn(item, field) && typeof item[field] !== "string") error(field, "문자열로 입력해 주세요.");
      });
    if (!PATENT_TYPES.some(({ value }) => value === item.type)) error("type", "특허 종류를 선택해 주세요.");
    const inventorKeys = item.inventor_keys || [];
    if (!Array.isArray(inventorKeys)) error("inventor_ids", "발명자 목록을 확인해 주세요.");
    else if (new Set(inventorKeys).size !== inventorKeys.length) error("inventor_ids", "같은 발명자를 중복해서 선택할 수 없습니다.");
    else if (inventorKeys.some((key) => !(state.peopleDraft || []).some((person) => person._clientKey === key))) {
      error("inventor_ids", "사람 목록에 등록된 발명자를 선택해 주세요.");
    }
    const numberFields = ["application_number", "publication_number", "registration_number"];
    if (!numberFields.some((field) => typeof item[field] === "string" && item[field].trim())) {
      numberFields.forEach((field) => error(field, "출원번호·공개번호·등록번호 중 하나 이상을 입력해 주세요."));
    }
    ["application_date", "publication_date", "registration_date"].forEach((field) => {
      if (item[field]) {
        const parsed = new Date(`${item[field]}T00:00:00Z`);
        if (typeof item[field] !== "string" || !/^\d{4}-\d{2}-\d{2}$/.test(item[field])
          || item[field].startsWith("0000-") || Number.isNaN(parsed.getTime()) || parsed.toISOString().slice(0, 10) !== item[field]) {
          error(field, "실제 날짜를 YYYY-MM-DD 형식으로 입력해 주세요.");
        }
      }
    });
    ["publication_date", "registration_date"].forEach((field) => {
      if (item.application_date && item[field] && item[field] < item.application_date) {
        error(field, "출원일과 같거나 이후인 날짜를 입력해 주세요.");
      }
    });
    if (item.url && (typeof item.url !== "string" || [...item.url].length > 4096
      || !/^https?:\/\//i.test(item.url) || /[\u0000-\u001f\u007f-\u009f]/.test(item.url) || !isHTTPURL(item.url))) {
      error("url", "4096자 이내의 http 또는 https 주소를 입력하거나 비워 주세요.");
    }
    if (Object.hasOwn(item, "visible_in_CV") && typeof item.visible_in_CV !== "boolean") {
      error("title_en", "CV 포함 여부를 다시 선택해 주세요.");
      error("title_ko", "CV 포함 여부를 다시 선택해 주세요.");
    }
  });
  (state.profileDraft?.education || []).forEach((item, index) => {
    if (item.thesis_url !== undefined && (typeof item.thesis_url !== "string" ||
      (item.thesis_url && (!/^https?:\/\//.test(item.thesis_url) || /[\u0000-\u001f\u007f]/.test(item.thesis_url) || !isHTTPURL(item.thesis_url))))) {
      state.validationErrors.set(profileValidationKey(["education", index, "thesis_url"]), "http 또는 https 학위논문 주소를 입력하거나 비워 주세요.");
    }
    if (item.gpa === undefined) return;
    const { value, max } = item.gpa || {};
    if (!Number.isFinite(max) || max <= 0) {
      state.validationErrors.set(profileValidationKey(["education", index, "gpa", "max"]), "GPA 만점을 0보다 큰 숫자로 입력해 주세요.");
    }
    if (!Number.isFinite(value) || value < 0 || (Number.isFinite(max) && value > max)) {
      state.validationErrors.set(profileValidationKey(["education", index, "gpa", "value"]), "GPA 획득점수를 0 이상, 만점 이하로 입력해 주세요.");
    }
  });
  (state.profileDraft?.certifications || []).forEach((item, index) => {
    if (!validDocumentFilename(item.pdf)) state.validationErrors.set(profileValidationKey(["certifications", index, "pdf"]), "data/document 안의 PDF 파일명만 입력해 주세요.");
  });
  Object.entries(PROFILE_PRIMARY_FIELDS).forEach(([section, fields]) => {
    (state.profileDraft?.[section] || []).forEach((item, index) => {
      if (profileItemIsMeaningful(section, item)) return;
      const message = `${PROFILE_REPEAT_LABELS[section]} 항목의 ${PROFILE_PRIMARY_PROMPTS[section]} 입력해 주세요.`;
      fields.forEach((field) => {
        state.validationErrors.set(profileValidationKey([section, index, field]), message);
      });
    });
  });
  (state.profileDraft?.profile_card?.credentials || []).forEach((item, index) => {
    const fields = ["name_en", "name_ko"];
    if (fields.some((field) => String(item?.[field] || "").trim())) return;
    fields.forEach((field) => {
      state.validationErrors.set(
        profileValidationKey(["profile_card", "credentials", index, field]),
        "자격 표기 항목의 이름을 입력해 주세요.",
      );
    });
  });
  (state.profileDraft?.scholarships || []).forEach((scholarship, scholarshipIndex) => {
    const amount = scholarship.amount || {};
    const rawValue = amount.value;
    const numericValue = typeof rawValue === "number" ? rawValue : Number(rawValue);
    if (rawValue === "" || rawValue === null || rawValue === undefined || !Number.isFinite(numericValue) || numericValue <= 0) {
      state.validationErrors.set(
        profileValidationKey(["scholarships", scholarshipIndex, "amount", "value"]),
        "금액은 0보다 큰 숫자로 입력해 주세요.",
      );
    }
    if (!/^[A-Z]{3}$/.test(String(amount.currency || ""))) {
      state.validationErrors.set(
        profileValidationKey(["scholarships", scholarshipIndex, "amount", "currency"]),
        "통화 코드는 대문자 영문 3자리로 입력해 주세요.",
      );
    }
  });
}

function validateDraft() {
  state.validationErrors = new Map();
  if (!state.draft) return;
  validateColorThemes();
  validateProfile();
  validateTaxonomy("topics");
  Object.keys(catalogConfig).forEach(validateCatalog);
  validateBoard();
  validateProjects();
  validateSoftware();
  validatePeople();
  validateAwards();
  validateAcademicActivities();
  validatePublications();
  applyValidationErrors();
}

function applyValidationErrors() {
  document.querySelectorAll("[data-validation-key]").forEach((input) => {
    const message = state.validationErrors.get(input.dataset.validationKey) || "";
    input.setAttribute("aria-invalid", String(Boolean(message)));
    input.setCustomValidity(message);
    if (message) input.closest("details")?.setAttribute("open", "");
  });
  document.querySelectorAll("[data-error-scope]").forEach((element) => {
    const prefix = element.dataset.errorScope;
    const messages = [];
    state.validationErrors.forEach((message, key) => {
      if (key.startsWith(prefix) && !messages.includes(message)) messages.push(message);
    });
    element.textContent = messages.join(" ");
    if (messages.length) element.closest("details")?.setAttribute("open", "");
  });
  document.querySelectorAll(".board-item-row[data-board-key]").forEach((row) => {
    const prefix = `board|${row.dataset.boardKey}|`;
    const invalid = Array.from(state.validationErrors.keys()).some((key) => key.startsWith(prefix));
    row.setAttribute("aria-invalid", String(invalid));
    const badge = row.querySelector("[data-board-error-badge]");
    if (badge) badge.hidden = !invalid;
  });
  document.querySelectorAll(".board-item-row[data-content-collection][data-content-key]").forEach((row) => {
    const prefix = `${row.dataset.contentCollection}|${row.dataset.contentKey}|`;
    const invalid = Array.from(state.validationErrors.keys()).some((key) => key.startsWith(prefix));
    row.setAttribute("aria-invalid", String(invalid));
    const badge = row.querySelector("[data-content-error-badge]");
    if (badge) badge.hidden = !invalid;
  });
  document.querySelectorAll(".board-item-row[data-entity-collection][data-entity-key]").forEach((row) => {
    const prefix = `${row.dataset.entityCollection}|${row.dataset.entityKey}|`;
    const invalid = Array.from(state.validationErrors.keys()).some((key) => key.startsWith(prefix));
    row.setAttribute("aria-invalid", String(invalid));
    const badge = row.querySelector("[data-entity-error-badge]");
    if (badge) badge.hidden = !invalid;
  });
  document.querySelectorAll(".board-item-row[data-academic-activity-key]").forEach((row) => {
    const prefix = `academicActivities|${row.dataset.academicActivityKey}|`;
    const invalid = Array.from(state.validationErrors.keys()).some((key) => key.startsWith(prefix));
    row.setAttribute("aria-invalid", String(invalid));
    const badge = row.querySelector("[data-academic-activity-error-badge]");
    if (badge) badge.hidden = !invalid;
  });
  document.querySelectorAll(".board-item-row[data-catalog-kind]").forEach((row) => {
    const prefix = `${row.dataset.catalogKind}|${row.dataset.catalogKey}|`;
    const invalid = [...state.validationErrors.keys()].some((key) => key.startsWith(prefix));
    row.setAttribute("aria-invalid", String(invalid));
    row.querySelector("[data-catalog-error]").hidden = !invalid;
  });
  document.querySelectorAll(".profile-repeat-row[data-profile-section][data-profile-index]").forEach((row) => {
    const prefix = row.dataset.profileSection === "credentials"
      ? `profile|profile_card|credentials|${row.dataset.profileIndex}|`
      : `profile|${row.dataset.profileSection}|${row.dataset.profileIndex}|`;
    const invalid = Array.from(state.validationErrors.keys()).some((key) => key.startsWith(prefix));
    row.setAttribute("aria-invalid", String(invalid));
    if (invalid) {
      state.profileExpandedRows.add(row.dataset.profileKey);
      setProfileRepeatRowExpanded(row, true);
    }
  });
}

function activateTab(nextTab, focus = true) {
  closeTaxonomyColorPopover();
  tabs.forEach((tab) => {
    const active = tab === nextTab;
    tab.setAttribute("aria-selected", String(active));
    tab.tabIndex = active ? 0 : -1;
  });
  panels.forEach((panel) => {
    panel.hidden = panel.dataset.panelName !== nextTab.dataset.panel;
  });
  if (state.draft && nextTab.dataset.panel === "cv") {
    const autoHidden = syncEmptyCVSections();
    renderCVHierarchy();
    if (autoHidden) updateDirtyState();
  }
  if (state.draft && nextTab.dataset.panel === "profile") renderProfileContentEditor();
  if (state.draft && catalogConfig[nextTab.dataset.panel]) {
    renderCatalog(nextTab.dataset.panel);
    applyValidationErrors();
  }
  if (state.draft && nextTab.dataset.panel === "topics") {
    renderTaxonomy("topics");
    applyValidationErrors();
  }
  if (state.draft && nextTab.dataset.panel === "publications") {
    renderPublicationEditor();
    applyValidationErrors();
  }
  if (focus) nextTab.focus();
}

const validationNavigation = Object.freeze({
  settings: { panel: "settings" },
  journals: { panel: "journals", catalogKind: "journals" },
  conferences: { panel: "conferences", catalogKind: "conferences" },
  profile: { panel: "profile" },
  topics: { panel: "topics", taxonomyKind: "topics" },
  projects: { panel: "projects", collection: "projects" },
  software: { panel: "software", collection: "software" },
  people: { panel: "people", collection: "people" },
  awards: { panel: "awards", collection: "awards" },
  academicActivities: { panel: "academic-activities", collection: "academicActivities" },
  publications: { panel: "publications", collection: "publications" },
  board: { panel: "board", collection: "board" },
});

function validationTargetForKey(key) {
  const [prefix, clientKey = ""] = String(key || "").split("|");
  const config = validationNavigation[prefix];
  return config ? { ...config, prefix, clientKey } : null;
}

function renderValidationTarget(target) {
  if (!target) return;
  if (target.catalogKind) {
    catalogSelection[target.catalogKind] = target.clientKey;
    renderCatalog(target.catalogKind);
    return;
  }
  if (target.taxonomyKind) {
    renderTaxonomy(target.taxonomyKind);
    return;
  }
  if (target.collection === "board") {
    if (state.boardDraft?.some((item) => item._clientKey === target.clientKey)) {
      state.selectedBoardKey = target.clientKey;
    }
    renderBoardList();
    renderBoardEditor();
    return;
  }
  if (target.collection === "projects" || target.collection === "software") {
    if (contentDraft(target.collection)?.some((item) => item._clientKey === target.clientKey)) {
      setSelectedContentKey(target.collection, target.clientKey);
    }
    renderContentList(target.collection);
    if (target.collection === "projects") renderProjectEditor();
    else renderSoftwareEditor();
    return;
  }
  if (["people", "awards", "publications"].includes(target.collection)) {
    if (entityDraft(target.collection)?.some((item) => item._clientKey === target.clientKey)) {
      setSelectedEntityKey(target.collection, target.clientKey);
    }
    renderEntityCollection(target.collection);
    return;
  }
  if (target.collection === "academicActivities") {
    if (state.academicActivitiesDraft?.some((item) => item._clientKey === target.clientKey)) {
      state.selectedAcademicActivityKey = target.clientKey;
    }
    renderAcademicActivities();
    return;
  }
  if (target.panel === "profile") renderProfileContentEditor();
}

function validationControlForKey(key) {
  return [...document.querySelectorAll("[data-validation-key]")]
    .find((control) => control.dataset.validationKey === key) || null;
}

function narrowestValidationScope(key) {
  return [...document.querySelectorAll("[data-error-scope]")]
    .filter((scope) => key.startsWith(scope.dataset.errorScope || ""))
    .sort((left, right) => (
      (right.dataset.errorScope || "").length - (left.dataset.errorScope || "").length
    ))[0] || null;
}

function focusValidationTarget(key, target) {
  requestAnimationFrame(() => {
    const exact = validationControlForKey(key);
    const scope = narrowestValidationScope(key);
    const panel = document.querySelector(`[data-panel-name="${target?.panel || ""}"]`);
    const editorSelectors = {
      journals: "#journals-form",
      conferences: "#conferences-form",
      profile: "#profile-content-editor",
      board: "#board-form",
      projects: "#project-form",
      software: "#software-form",
      people: "#people-form",
      awards: "#awards-form",
      "academic-activities": "#academic-activities-form",
      publications: "#publications-form",
    };
    const taxonomyPanel = target?.taxonomyKind
      ? document.querySelector(`[data-taxonomy-panel="${target.taxonomyKind}"]`)
      : null;
    const editorSelector = editorSelectors[target?.panel];
    const editor = taxonomyPanel
      || (editorSelector ? document.querySelector(editorSelector) : null)
      || panel;
    const invalidControl = editor?.querySelector?.('[aria-invalid="true"]:not([disabled])');
    const fallbackControl = editor?.querySelector?.(
      'input:not([disabled]), textarea:not([disabled]), select:not([disabled]), button:not([disabled]), [tabindex]:not([tabindex="-1"])',
    );
    const focusTarget = exact || invalidControl || scope || fallbackControl || editor || panel;
    const scrollTarget = exact || scope || editor || panel;
    scrollTarget?.scrollIntoView({ block: "center" });
    if (focusTarget && typeof focusTarget.focus === "function") {
      if (!focusTarget.matches?.('input, textarea, select, button, [tabindex]')) focusTarget.tabIndex = -1;
      focusTarget.focus({ preventScroll: true });
    }
    if (exact && typeof exact.reportValidity === "function") exact.reportValidity();
  });
}

function revealFirstValidationError() {
  const firstKey = state.validationErrors.keys().next().value;
  if (!firstKey) return false;
  const target = validationTargetForKey(firstKey);
  const tab = tabs.find((candidate) => candidate.dataset.panel === target?.panel);
  if (tab) activateTab(tab, false);
  renderValidationTarget(target);
  applyValidationErrors();
  updateToolbar();
  focusValidationTarget(firstKey, target);
  return true;
}

tabs.forEach((tab, index) => {
  tab.addEventListener("click", () => activateTab(tab));
  tab.addEventListener("keydown", (event) => {
    const keyDirections = {
      ArrowDown: 1,
      ArrowRight: 1,
      ArrowUp: -1,
      ArrowLeft: -1,
    };
    if (event.key === "Home" || event.key === "End") {
      event.preventDefault();
      activateTab(event.key === "Home" ? tabs[0] : tabs[tabs.length - 1]);
      return;
    }
    if (!keyDirections[event.key]) return;
    event.preventDefault();
    const direction = keyDirections[event.key];
    activateTab(tabs[(index + direction + tabs.length) % tabs.length]);
  });
});

function setActionButtonIcon(button, kind, accessibleLabel) {
  button.classList.add("action-icon-button", `${kind}-action-button`);
  button.dataset.actionIcon = kind;
  button.replaceChildren();
  if (kind === "add") {
    const icon = document.createElement("span");
    icon.className = "action-icon-glyph";
    icon.setAttribute("aria-hidden", "true");
    icon.textContent = "+";
    button.append(icon);
  } else {
    const namespace = "http://www.w3.org/2000/svg";
    const icon = document.createElementNS(namespace, "svg");
    icon.classList.add("action-icon-svg");
    icon.setAttribute("viewBox", "0 0 24 24");
    icon.setAttribute("aria-hidden", "true");
    icon.setAttribute("focusable", "false");
    [
      "M4 7h16",
      "M9 7V4h6v3",
      "M7 7l1 13h8l1-13",
      "M10 11v5",
      "M14 11v5",
    ].forEach((data) => {
      const path = document.createElementNS(namespace, "path");
      path.setAttribute("d", data);
      icon.append(path);
    });
    button.append(icon);
  }
  button.title = accessibleLabel;
  button.setAttribute("aria-label", accessibleLabel);
}

function makeButton(label, className, action, title = label) {
  const button = document.createElement("button");
  button.type = "button";
  button.className = className;
  button.textContent = label;
  button.title = title;
  button.dataset.action = action;
  if (button.classList.contains("delete-button")) {
    setActionButtonIcon(button, "delete", title);
  } else if (/(?:^|-)add$/.test(action)) {
    setActionButtonIcon(button, "add", title);
  } else if (title !== label) {
    button.setAttribute("aria-label", title);
  }
  button.disabled = state.loading || state.saving || state.discarding || state.boardDropBusy;
  return button;
}

function applyListRowSelection(row, select, selected) {
  row.classList.toggle("is-selected", selected);
  if (selected) select.setAttribute("aria-current", "true");
  else select.removeAttribute("aria-current");
}

function makeListStateBadges(item, errorDataset) {
  const badges = document.createElement("span");
  badges.className = "board-item-badges";
  if (item._isNew) {
    const newBadge = document.createElement("span");
    newBadge.className = "board-badge board-badge-new";
    newBadge.textContent = "\uc2e0\uaddc";
    badges.append(newBadge);
  }
  const errorBadge = document.createElement("span");
  errorBadge.className = "board-badge board-badge-error";
  errorBadge.dataset[errorDataset] = "";
  errorBadge.textContent = "\uc785\ub825 \uc624\ub958";
  errorBadge.hidden = true;
  return { badges, errorBadge };
}

function academicActivityIsMeaningful(item) {
  if (!item) return false;
  if (item.category === "reviews" || item.category === "editorial_service") {
    return Boolean(String(item.journal_en || "").trim() || String(item.journal_ko || "").trim());
  }
  if (item.category === "invited_talks") {
    return Boolean(String(item.event_en || "").trim() || String(item.event_ko || "").trim());
  }
  if (item.category === "conference_service") {
    return Boolean(String(item.conference_en || "").trim() || String(item.conference_ko || "").trim());
  }
  return Boolean(String(item.organization_en || "").trim() || String(item.organization_ko || "").trim());
}

function mainSectionItemCount(section) {
  if (section === "awards") {
    return (state.awardsDraft || []).filter((item) => (
      String(item.title_en || "").trim() || String(item.title_ko || "").trim()
    )).length;
  }
  if (section === "academic_activities") {
    return (state.academicActivitiesDraft || []).filter(academicActivityIsMeaningful).length;
  }
  return Array.isArray(state.profileDraft?.[section])
    ? state.profileDraft[section].filter((item) => profileItemIsMeaningful(section, item)).length
    : 0;
}

function setProfileSectionFeedback(message = "") {
  const feedback = document.querySelector("#profile-section-feedback");
  if (feedback) feedback.textContent = message;
}

function syncEmptyMainSections() {
  if (!state.draft) return false;
  let changed = false;
  [...state.draft.main_page_sections].forEach((section) => {
    if (section === "patents") return;
    if (mainSectionItemCount(section) > 0) return;
    const index = state.draft.main_page_sections.indexOf(section);
    if (index >= 0) state.draft.main_page_sections.splice(index, 1);
    if (!state.draft.hidden_main_page_sections.includes(section)) {
      state.draft.hidden_main_page_sections.push(section);
    }
    changed = true;
  });
  return changed;
}

function cvSectionItemIsMeaningful(section, item) {
  if (!item || item.visible_in_CV === false) return false;
  if (["experience", "education", "teaching", "scholarships", "certifications", "patents", "skills"].includes(section)) {
    return profileItemIsMeaningful(section, item);
  }
  if (section === "projects") {
    return Boolean(String(item.title_en || "").trim() || String(item.title_ko || "").trim());
  }
  if (section === "publications") {
    return Boolean(String(item.title_en || "").trim() || String(item.title_ko || "").trim());
  }
  if (section === "software") return Boolean(String(item.name || "").trim());
  if (section === "awards") {
    return Boolean(String(item.title_en || "").trim() || String(item.title_ko || "").trim());
  }
  if (section === "academic_activities") return academicActivityIsMeaningful(item);
  return false;
}

function cvSectionItemCount(section) {
  return cvSectionItems(section).filter((item) => cvSectionItemIsMeaningful(section, item)).length;
}

function emptyCVSectionFeedback(section) {
  return `${CV_SECTION_LABELS[section] || section}: CV에 노출할 항목이 없습니다. 먼저 항목을 추가하거나 CV 항목을 선택해 주세요.`;
}

function setCVSectionFeedback(message = "") {
  const feedback = document.querySelector("#cv-section-feedback");
  if (feedback) feedback.textContent = message;
}

function syncEmptyCVSections() {
  if (!state.draft) return false;
  let changed = false;
  [...state.draft.cv_sections].forEach((section) => {
    if (section === "patents") return;
    if (cvSectionItemCount(section) > 0) return;
    const index = state.draft.cv_sections.indexOf(section);
    if (index >= 0) state.draft.cv_sections.splice(index, 1);
    if (!state.draft.hidden_cv_sections.includes(section)) {
      state.draft.hidden_cv_sections.push(section);
    }
    changed = true;
  });
  return changed;
}

function syncCVSectionsAfterItemsChange() {
  const previouslyVisible = [...(state.draft?.cv_sections || [])];
  if (!syncEmptyCVSections()) return false;
  const autoHidden = previouslyVisible.find((section) => !state.draft.cv_sections.includes(section));
  if (autoHidden) {
    setCVSectionFeedback(
      `${CV_SECTION_LABELS[autoHidden] || autoHidden}: CV에 노출할 항목이 없어 자동으로 숨겼습니다.`,
    );
  }
  return true;
}

function profileValue(path) {
  return path.reduce((value, key) => value?.[key], state.profileDraft);
}

function setProfileValue(path, value) {
  let target = state.profileDraft;
  path.slice(0, -1).forEach((key) => {
    if (!target[key] || typeof target[key] !== "object") target[key] = {};
    target = target[key];
  });
  target[path[path.length - 1]] = value;
}

function profileField(label, path, options = {}) {
  const field = document.createElement("label");
  field.className = `profile-field${options.full ? " is-full" : ""}`;
  const heading = document.createElement("span");
  heading.className = "profile-field-label";
  heading.append(document.createTextNode(label));
  if (options.language) {
    const marker = document.createElement("span");
    marker.className = "language-marker";
    marker.textContent = options.language;
    heading.append(marker);
  }
  const input = document.createElement(options.choices ? "select" : options.textarea ? "textarea" : "input");
  if (options.choices) {
    options.choices.forEach(({ value, label: choiceLabel }) => {
      const option = document.createElement("option");
      option.value = value;
      option.textContent = choiceLabel;
      input.append(option);
    });
  } else if (!options.textarea) {
    input.type = options.type || "text";
    if (options.min !== undefined) input.min = String(options.min);
    if (options.step !== undefined) input.step = String(options.step);
    if (options.maxLength !== undefined) input.maxLength = Number(options.maxLength);
    if (options.list) input.setAttribute("list", options.list);
    if (options.inputMode) input.inputMode = options.inputMode;
  }
  input.value = String(profileValue(path) ?? "");
  input.placeholder = options.placeholder || "";
  input.disabled = state.loading || state.saving || state.discarding;
  input.dataset.profilePath = JSON.stringify(path);
  field.append(heading, input);
  if (isProfilePrimaryPath(path)) {
    const error = document.createElement("span");
    error.className = "profile-field-error";
    error.dataset.errorScope = profileValidationKey(path);
    error.id = `profile-error-${path.map((part) => String(part).replace(/[^a-z0-9-]/gi, "-")).join("-")}`;
    input.dataset.validationKey = profileValidationKey(path);
    input.setAttribute("aria-describedby", error.id);
    field.append(error);
  }
  return field;
}

function makeLanguagePairFields(className, createField) {
  const fields = document.createElement("div");
  fields.className = className;
  [["\uc601\ubb38", "english"], ["\uad6d\ubb38", "korean"]].forEach(([language, key], index) => {
    fields.append(createField({ language, key, index }));
  });
  return fields;
}

function appendProfilePair(container, label, basePath, englishKey, koreanKey, options = {}) {
  const group = document.createElement("fieldset");
  group.className = "profile-language-pair";
  const legend = document.createElement("legend");
  legend.className = "profile-pair-label";
  legend.textContent = label;
  const pairOptions = { ...options, full: false };
  const fields = makeLanguagePairFields("profile-language-pair-fields", ({ language, key }) => (
    profileField("", [...basePath, key === "english" ? englishKey : koreanKey], { ...pairOptions, language })
  ));
  group.append(legend, fields);
  container.append(group);
}

function profileStaticSection(title, fields, options = {}) {
  const section = document.createElement("section");
  section.className = `profile-editor-section${options.className ? ` ${options.className}` : ""}`;
  if (options.id) section.id = options.id;
  const heading = document.createElement("div");
  heading.className = "profile-editor-section-heading";
  const titleElement = document.createElement("h4");
  titleElement.textContent = title;
  heading.append(titleElement);
  const grid = document.createElement("div");
  grid.className = "profile-fields";
  fields(grid);
  section.append(heading, grid);
  return section;
}

const PROFILE_REPEAT_LABELS = {
  credentials: "자격 표기",
  experience: "경력",
  education: "학력",
  scholarships: "장학",
  certifications: "자격",
  teaching: "교육",
  patents: "특허",
  skills: "기술",
};

const PROFILE_PRIMARY_FIELDS = {
  experience: ["title_en", "title_ko"],
  education: ["degree_en", "degree_ko"],
  scholarships: ["name_en", "name_ko"],
  certifications: ["name_en", "name_ko"],
  teaching: ["title_en", "title_ko"],
  patents: ["title_en", "title_ko"],
  skills: ["name"],
};

const PROFILE_PRIMARY_PROMPTS = {
  experience: "직함을",
  education: "학위를",
  scholarships: "이름을",
  certifications: "이름을",
  teaching: "제목을",
  patents: "제목을",
  skills: "이름을",
  awards: "제목을",
};

function profileValidationKey(path) {
  return `profile|${path.map(String).join("|")}`;
}

function isProfilePrimaryPath(path) {
  if (path[0] === "patents" && Number.isInteger(path[1]) && path.length === 3) return true;
  if (path[0] === "education" && path[2] === "thesis_url") return true;
  if (path[0] === "education" && path[2] === "gpa" && ["value", "max"].includes(path[3])) return true;
  if (Number.isInteger(path[1]) && (PROFILE_PRIMARY_FIELDS[path[0]] || []).includes(path[2])) return true;
  if (
    path[0] === "profile_card"
    && path[1] === "credentials"
    && Number.isInteger(path[2])
    && ["name_en", "name_ko"].includes(path[3])
  ) return true;
  if (path[0] === "scholarships"
    && Number.isInteger(path[1])
    && path[2] === "amount"
    && ["value", "currency"].includes(path[3])) return true;
  return false;
}

function profileItemIsMeaningful(section, item) {
  return (PROFILE_PRIMARY_FIELDS[section] || []).some((field) => String(item?.[field] || "").trim());
}

function defaultProfileItem(section) {
  const common = { _clientKey: `profile-${section}:new:${state.nextContentRowKey++}` };
  if (section === "credentials") return { ...common, name_en: "", name_ko: "" };
  if (section === "experience") return { ...common, visible_in_CV: true, title_en: "", title_ko: "", institution_en: "", institution_ko: "", period_en: "", period_ko: "" };
  if (section === "education") return { ...common, visible_in_CV: true, degree_en: "", degree_ko: "", institution_en: "", institution_ko: "", period: "", advisor_en: "", advisor_ko: "", thesis_en: "", thesis_ko: "" };
  if (section === "scholarships") return { ...common, visible_in_CV: true, name_en: "", name_ko: "", period_en: "", period_ko: "", summary_en: "", summary_ko: "", amount: { value: "", currency: "KRW" } };
  if (section === "certifications") return { ...common, visible_in_CV: true, name_en: "", name_ko: "", issuer_en: "", issuer_ko: "", date: localToday() };
  if (section === "teaching") return { ...common, visible_in_CV: true, title_en: "", title_ko: "", detail_en: "", detail_ko: "", period: "" };
  if (section === "patents") return {
    ...common, visible_in_CV: true, title_en: "", title_ko: "", type: "domestic",
    application_number: "", application_date: "", publication_number: "", publication_date: "",
    registration_number: "", registration_date: "", inventor_ids: [], inventor_keys: [], _unresolvedInventorIDs: {},
    applicant_en: "", applicant_ko: "", url: "",
  };
  return { ...common, visible_in_CV: true, name: "", detail_en: "", detail_ko: "" };
}

function profileRepeatItems(section) {
  if (section === "credentials") return state.profileDraft.profile_card.credentials;
  return state.profileDraft[section];
}

function profileSummaryValue(...values) {
  return [...new Set(values.map((value) => String(value || "").trim()).filter(Boolean))].join(" · ");
}

function scholarshipAmountLabel(amount) {
  const rawValue = amount?.value;
  const currency = String(amount?.currency || "").trim();
  if (rawValue === "" || rawValue === null || rawValue === undefined || !currency) return "";
  const numericValue = Number(rawValue);
  const value = Number.isFinite(numericValue)
    ? numericValue.toLocaleString("en-US", { maximumFractionDigits: 20 })
    : String(rawValue);
  return `${value} ${currency}`;
}

function profileRepeatSummary(section, item) {
  if (section === "patents") {
    return {
      title: String(item.title_ko || "").trim() || String(item.title_en || "").trim() || "새 특허",
      detail: profileSummaryValue(
        PATENT_TYPES.find(({ value }) => value === item.type)?.label,
        patentStatusLabel(item),
        item.registration_number || item.publication_number || item.application_number,
        patentInventorNames(item),
      ),
    };
  }
  if (section === "credentials") {
    return {
      title: item.name_ko || item.name_en || "새 자격 표기",
      detail: profileSummaryValue(item.name_ko ? item.name_en : ""),
    };
  }
  if (section === "experience") {
    return {
      title: item.title_ko || item.title_en || "새 경력",
      detail: profileSummaryValue(
        item.institution_ko || item.institution_en,
        item.period_ko || item.period_en,
      ),
    };
  }
  if (section === "education") {
    return {
      title: item.degree_ko || item.degree_en || "새 학력",
      detail: profileSummaryValue(item.institution_ko || item.institution_en, item.period),
    };
  }
  if (section === "scholarships") {
    return {
      title: item.name_ko || item.name_en || "새 장학",
      detail: profileSummaryValue(item.period_ko || item.period_en, scholarshipAmountLabel(item.amount)),
    };
  }
  if (section === "certifications") {
    return {
      title: item.name_ko || item.name_en || "새 자격",
      detail: profileSummaryValue(item.issuer_ko || item.issuer_en, item.date),
    };
  }
  if (section === "teaching") {
    return {
      title: item.title_ko || item.title_en || "새 교육",
      detail: profileSummaryValue(item.period),
    };
  }
  return {
    title: item.name || "새 기술",
    detail: profileSummaryValue(item.detail_ko || item.detail_en),
  };
}

function setProfileRepeatRowExpanded(row, expanded) {
  const disclosure = row.querySelector('[data-action="profile-item-toggle"]');
  const body = row.querySelector(".profile-repeat-body");
  if (!disclosure || !body) return;
  row.classList.toggle("is-expanded", expanded);
  disclosure.setAttribute("aria-expanded", String(expanded));
  body.hidden = !expanded;
  const chevron = disclosure.querySelector(".profile-repeat-chevron");
  if (chevron) chevron.textContent = expanded ? "접기" : "펼치기";
  const summary = disclosure.querySelector(".profile-repeat-summary-text strong")?.textContent || "항목";
  const label = `${summary} 상세 ${expanded ? "접기" : "펼치기"}`;
  disclosure.title = label;
  disclosure.setAttribute("aria-label", label);
}

function patentInventorNames(item) {
  return (item.inventor_keys || []).map((key) => {
    const person = (state.peopleDraft || []).find((candidate) => candidate._clientKey === key);
    return person ? personNameLabel(person) : item._unresolvedInventorIDs?.[key] || "연결된 사람을 찾을 수 없음";
  }).join("; ");
}

function patentByKey(key) {
  return (state.profileDraft?.patents || []).find((item) => item._clientKey === key);
}

function updatePatentSummary(item) {
  const row = [...document.querySelectorAll('.profile-repeat-row[data-profile-section="patents"]')]
    .find((element) => element.dataset.profileKey === item._clientKey);
  if (!row) return;
  const summary = profileRepeatSummary("patents", item);
  const text = row.querySelector(".profile-repeat-summary-text");
  const title = text?.querySelector("strong");
  if (title) title.textContent = summary.title;
  let detail = text?.querySelector("span");
  if (!detail && text) { detail = document.createElement("span"); text.append(detail); }
  if (detail) detail.textContent = summary.detail;
  const status = row.querySelector("[data-patent-status-key]");
  if (status) status.textContent = patentStatusLabel(item);
}

function updatePatentInventorOptions(section, item) {
  const select = section.querySelector("[data-patent-inventor-select]");
  if (!select) return;
  const query = String(item._inventorSearch || "").trim().toLocaleLowerCase("ko");
  const candidates = sortedPeopleForDisplay((state.peopleDraft || []).filter((person) =>
    !(item.inventor_keys || []).includes(person._clientKey)
      && (!query || personRelationshipLabel(person).toLocaleLowerCase("ko").includes(query))));
  const placeholder = document.createElement("option");
  placeholder.value = "";
  placeholder.textContent = candidates.length ? "발명자 선택" : "선택할 사람이 없습니다";
  select.replaceChildren(placeholder);
  candidates.forEach((person) => {
    const option = document.createElement("option");
    option.value = person._clientKey;
    option.textContent = personRelationshipLabel(person);
    select.append(option);
  });
  select.disabled = state.loading || state.saving || state.discarding || !candidates.length;
}

function renderPatentInventors(item, index) {
  const keys = item.inventor_keys || [];
  const section = document.createElement("section");
  section.className = "content-repeat-section patent-inventors";
  section.dataset.patentInventorsKey = item._clientKey;
  section.append(contentSectionHeading("발명자", keys.length));
  const addRow = document.createElement("div");
  addRow.className = "relationship-add-row patent-inventor-add-row";
  const search = document.createElement("input");
  search.type = "search";
  search.className = "board-input";
  search.placeholder = "이름·소속 검색";
  search.setAttribute("aria-label", "발명자로 추가할 사람 검색");
  search.dataset.patentInventorSearch = item._clientKey;
  search.value = item._inventorSearch || "";
  search.disabled = state.loading || state.saving || state.discarding;
  const select = document.createElement("select");
  select.className = "board-input";
  select.dataset.patentInventorSelect = item._clientKey;
  select.setAttribute("aria-label", "추가할 발명자");
  select.dataset.validationKey = profileValidationKey(["patents", index, "inventor_ids"]);
  const addPerson = makeButton("사람 추가", "button button-secondary", "patent-person-add", "새 사람을 추가하고 발명자로 선택");
  addPerson.disabled = state.loading || state.saving || state.discarding || (state.peopleDraft?.length || 0) >= 500;
  addRow.append(search, select, addPerson);
  const list = document.createElement("ol");
  list.className = "relationship-list";
  keys.forEach((key, position) => {
    const person = (state.peopleDraft || []).find((candidate) => candidate._clientKey === key);
    const row = document.createElement("li");
    row.className = "relationship-row";
    row.dataset.patentInventorKey = key;
    const label = document.createElement("span");
    label.className = "relationship-person";
    const name = document.createElement("strong");
    name.textContent = person ? personNameLabel(person) : "연결된 사람을 찾을 수 없음";
    const affiliation = document.createElement("span");
    affiliation.textContent = person ? personAffiliationLabel(person) || "소속/메모 없음" : item._unresolvedInventorIDs?.[key] || "";
    label.append(name, affiliation);
    const edit = makeButton("사람 편집", "text-button", "patent-inventor-edit", "선택한 사람 편집");
    edit.disabled = !person || state.saving || state.discarding;
    row.append(label, edit, repeatRowActions("patent-inventor", position, keys.length, "발명자 제거"));
    list.append(row);
  });
  const error = document.createElement("span");
  error.className = "board-field-error";
  error.dataset.errorScope = select.dataset.validationKey;
  error.id = `patent-inventor-error-${index}`;
  select.setAttribute("aria-describedby", error.id);
  section.append(addRow, list, error);
  updatePatentInventorOptions(section, item);
  return section;
}

function refreshPatentInventorEditors() {
  document.querySelectorAll("[data-patent-inventors-key]").forEach((section) => {
    const item = patentByKey(section.dataset.patentInventorsKey);
    if (!item) return;
    const index = state.profileDraft.patents.indexOf(item);
    section.replaceWith(renderPatentInventors(item, index));
    updatePatentSummary(item);
  });
}

function refreshPatentInventorChange(item) {
  refreshPatentInventorEditors();
  updatePatentSummary(item);
  renderEntityList("people");
  renderCVHierarchy();
  updateDirtyState();
}

function renderProfileRepeatFields(grid, section, index) {
  const base = section === "credentials" ? ["profile_card", "credentials", index] : [section, index];
  if (section === "patents") {
    appendProfilePair(grid, "제목", base, "title_en", "title_ko");
    const typeField = profileField("종류", [...base, "type"], { choices: PATENT_TYPES });
    typeField.classList.add("patent-type-field");
    grid.append(typeField);
    const statusField = document.createElement("div");
    statusField.className = "profile-field patent-status-field";
    const statusLabel = document.createElement("span");
    statusLabel.className = "profile-field-label";
    statusLabel.textContent = "상태";
    const status = document.createElement("output");
    status.dataset.patentStatusKey = profileRepeatItems(section)[index]._clientKey;
    status.setAttribute("aria-live", "polite");
    status.textContent = patentStatusLabel(profileRepeatItems(section)[index]);
    const statusSpacing = document.createElement("span");
    statusSpacing.className = "profile-field-error";
    statusSpacing.setAttribute("aria-hidden", "true");
    statusField.append(statusLabel, status, statusSpacing);
    grid.append(statusField);
    [["application", "출원"], ["publication", "공개"], ["registration", "등록"]].forEach(([prefix, label]) => {
      grid.append(
        profileField(`${label}번호`, [...base, `${prefix}_number`]),
        profileField(`${label}일`, [...base, `${prefix}_date`], { type: "date" }),
      );
    });
    grid.append(renderPatentInventors(profileRepeatItems(section)[index], index));
    appendProfilePair(grid, "출원인", base, "applicant_en", "applicant_ko");
    grid.append(profileField("URL", [...base, "url"], { full: true, type: "url", placeholder: "https://..." }));
  }
  if (section === "credentials") appendProfilePair(grid, "이름", base, "name_en", "name_ko");
  if (section === "experience") {
    appendProfilePair(grid, "직함", base, "title_en", "title_ko");
    appendProfilePair(grid, "소속", base, "institution_en", "institution_ko");
    appendProfilePair(grid, "기간", base, "period_en", "period_ko");
  }
  if (section === "education") {
    appendProfilePair(grid, "학위", base, "degree_en", "degree_ko");
    appendProfilePair(grid, "기관", base, "institution_en", "institution_ko");
    grid.append(profileField("기간", [...base, "period"], { full: true }));
    appendProfilePair(grid, "지도교수", base, "advisor_en", "advisor_ko");
    appendProfilePair(grid, "학위논문", base, "thesis_en", "thesis_ko", { textarea: true });
    grid.append(profileField("학위논문 링크 (선택)", [...base, "thesis_url"], { full: true, type: "url", placeholder: "https://..." }));
    const gpa = document.createElement("div");
    gpa.className = "profile-gpa";
    gpa.append(
      profileField("GPA 획득점수 (선택)", [...base, "gpa", "value"], { type: "number", min: 0, step: "any", inputMode: "decimal", placeholder: "3.0" }),
      profileField("GPA 만점 (선택)", [...base, "gpa", "max"], { type: "number", min: 0, step: "any", inputMode: "decimal", placeholder: "4.0" }),
    );
    grid.append(gpa);
  }
  if (section === "scholarships") {
    appendProfilePair(grid, "이름", base, "name_en", "name_ko");
    appendProfilePair(grid, "기간", base, "period_en", "period_ko");
    appendProfilePair(grid, "설명", base, "summary_en", "summary_ko", { textarea: true });
    const amount = document.createElement("div");
    amount.className = "profile-scholarship-amount";
    amount.append(
      profileField("금액", [...base, "amount", "value"], {
        type: "number", min: 0, step: "any", inputMode: "decimal",
      }),
      profileField("통화", [...base, "amount", "currency"], {
        list: "scholarship-currency-codes", maxLength: 3, placeholder: "KRW",
      }),
    );
    grid.append(amount);
  }
  if (section === "certifications") {
    appendProfilePair(grid, "이름", base, "name_en", "name_ko");
    appendProfilePair(grid, "발급기관", base, "issuer_en", "issuer_ko");
    grid.append(profileField("날짜", [...base, "date"], { full: true, type: "date" }));
    grid.append(profileField("PDF 파일명 (선택)", [...base, "pdf"], { full: true, placeholder: "data/document 안의 파일명.pdf" }));
    grid.append(makeDocumentSection("certifications", profileRepeatItems(section)[index]));
  }
  if (section === "teaching") {
    appendProfilePair(grid, "제목", base, "title_en", "title_ko");
    appendProfilePair(grid, "내용", base, "detail_en", "detail_ko", { textarea: true });
    grid.append(profileField("기간", [...base, "period"], { full: true }));
  }
  if (section === "skills") {
    grid.append(profileField("이름", [...base, "name"], { full: true }));
    appendProfilePair(grid, "설명", base, "detail_en", "detail_ko", { textarea: true });
  }
}

function profileRepeatSection(section, options = {}) {
  const items = profileRepeatItems(section);
  const container = document.createElement("section");
  container.className = `profile-editor-section${options.nested ? " profile-card-subsection" : ""}`;
  if (options.id) container.id = options.id;
  const heading = document.createElement("div");
  heading.className = "profile-editor-section-heading";
  const title = document.createElement(options.nested ? "h5" : "h4");
  title.textContent = PROFILE_REPEAT_LABELS[section];
  const headingActions = document.createElement("div");
  headingActions.className = "list-heading-actions";
  const count = document.createElement("span");
  count.className = "count-chip";
  count.textContent = String(items.length);
  const add = makeButton("+", "button button-primary", "profile-item-add", `${PROFILE_REPEAT_LABELS[section]} 항목 추가`);
  add.dataset.profileSection = section;
  headingActions.append(count, add);
  heading.append(title, headingActions);
  const list = document.createElement("div");
  list.className = "profile-repeat-list";
  items.forEach((item, index) => {
    const row = document.createElement("article");
    row.className = "profile-repeat-row";
    row.dataset.profileSection = section;
    row.dataset.profileIndex = String(index);
    row.dataset.profileKey = item._clientKey;
    const expanded = state.profileExpandedRows.has(item._clientKey)
      || item._isNew === true
      || String(item._clientKey || "").includes(":new:");
    if (expanded) state.profileExpandedRows.add(item._clientKey);
    const summary = profileRepeatSummary(section, item);
    const summaryRow = document.createElement("div");
    summaryRow.className = "profile-repeat-summary";
    const disclosure = makeButton("", "profile-repeat-disclosure", "profile-item-toggle", `${summary.title} 상세 ${expanded ? "접기" : "펼치기"}`);
    disclosure.dataset.profileSection = section;
    disclosure.dataset.profileIndex = String(index);
    disclosure.dataset.profileKey = item._clientKey;
    disclosure.setAttribute("aria-expanded", String(expanded));
    const bodyID = `profile-repeat-body-${section}-${index}`;
    disclosure.setAttribute("aria-controls", bodyID);
    const summaryText = document.createElement("span");
    summaryText.className = "profile-repeat-summary-text";
    const summaryTitle = document.createElement("strong");
    summaryTitle.textContent = summary.title;
    summaryText.append(summaryTitle);
    if (summary.detail) {
      const summaryDetail = document.createElement("span");
      summaryDetail.textContent = summary.detail;
      summaryText.append(summaryDetail);
    }
    const chevron = document.createElement("span");
    chevron.className = "profile-repeat-chevron";
    chevron.textContent = expanded ? "접기" : "펼치기";
    disclosure.append(summaryText, chevron);
    const grid = document.createElement("div");
    grid.className = "profile-fields";
    renderProfileRepeatFields(grid, section, index);
    const actions = document.createElement("div");
    actions.className = "profile-repeat-summary-actions";
    const up = makeButton("↑", "icon-button", "profile-item-up", "위로");
    const down = makeButton("↓", "icon-button", "profile-item-down", "아래로");
    const remove = makeButton("삭제", "button button-secondary delete-button", "profile-item-delete", `${PROFILE_REPEAT_LABELS[section]} 항목 삭제`);
    [up, down, remove].forEach((button) => {
      button.dataset.profileSection = section;
      button.dataset.profileIndex = String(index);
    });
    up.disabled = state.saving || index === 0;
    down.disabled = state.saving || index === items.length - 1;
    actions.append(up, down, remove);
    summaryRow.append(disclosure, actions);
    const body = document.createElement("div");
    body.className = "profile-repeat-body";
    body.id = bodyID;
    body.hidden = !expanded;
    body.append(grid);
    row.append(summaryRow, body);
    list.append(row);
  });
  container.append(heading, list);
  return container;
}

function renderProfileMediaSection() {
  const media = state.profileDraft?.profile_card?.media || {};
  const editorMedia = state.profileMediaDraft || {};
  const hasMedia = !editorMedia.remove && Boolean(editorMedia.stage_token || media.src);
  const section = document.createElement("section");
  section.className = "board-media-section profile-media-section profile-card-subsection";
  const heading = document.createElement("div");
  heading.className = "board-media-heading";
  const title = document.createElement("h5");
  title.textContent = "이미지";
  const count = document.createElement("span");
  count.className = "count-chip";
  count.textContent = hasMedia ? "1" : "0";
  heading.append(title, count);

  const dropzone = document.createElement("div");
  dropzone.className = `board-dropzone profile-media-dropzone${state.boardDropBusy ? " is-receiving" : ""}`;
  dropzone.dataset.profileMediaDropzone = "";
  dropzone.setAttribute("aria-label", "Explorer에서 프로필 이미지를 드롭하는 영역");
  dropzone.setAttribute("aria-busy", String(state.boardDropBusy));
  const dropTitle = document.createElement("strong");
  dropTitle.textContent = state.boardDropBusy ? "이미지를 준비하는 중…" : "Explorer에서 이미지 한 장을 이곳으로 드래그";
  dropzone.append(dropTitle);
  if (state.profileDropMessage) {
    const feedback = document.createElement("p");
    feedback.className = "board-drop-feedback";
    feedback.setAttribute("role", "status");
    feedback.textContent = state.profileDropMessage;
    dropzone.append(feedback);
  }
  section.append(heading, dropzone);

  if (!hasMedia) return section;
  const list = document.createElement("ol");
  list.className = "board-media-list profile-media-list";
  const row = document.createElement("li");
  row.className = "board-media-row profile-media-row";
  const preview = makeMediaPreviewButton(
    { ...media, ...editorMedia, type: editorMedia.stage_token ? "image" : media.type },
    "프로필 이미지",
    "profile-media-preview",
  );
  const details = document.createElement("div");
  details.className = "board-media-fields";
  const name = document.createElement("span");
  name.className = "board-media-name";
  const sizeLabel = formatBytes(editorMedia.size || 0);
  name.textContent = `${editorMedia.original_name || media.src || "프로필 이미지"}${sizeLabel ? ` · ${sizeLabel}` : ""}`;
  const hint = document.createElement("span");
  hint.className = "profile-media-hint";
  hint.textContent = editorMedia.stage_token ? "저장하면 새 이미지로 교체됩니다." : "새 이미지를 드롭하면 교체됩니다.";
  details.append(name, hint);
  const remove = makeButton("제거", "text-button delete-button", "profile-media-delete", "프로필 이미지 제거");
  row.append(preview, details, remove);
  list.append(row);
  section.append(list);
  return section;
}

function renderProfileContentEditor() {
  const container = document.querySelector("#profile-content-editor");
  if (!container || !state.profileDraft) return;
  container.replaceChildren();
  const outline = document.createElement("nav");
  outline.className = "profile-outline";
  outline.setAttribute("aria-label", "프로필 정보 목차");
  [
    ["기본 정보", "profile-section-basic"],
    ["연락처", "profile-section-contact"],
    ["경력", "profile-section-experience"],
    ["학력", "profile-section-education"],
    ["장학", "profile-section-scholarships"],
    ["자격", "profile-section-certifications"],
    ["수상", "awards"],
    ["교육", "profile-section-teaching"],
    ["특허", "profile-section-patents"],
    ["기술", "profile-section-skills"],
    ["페이지 구성", "profile-page-composition"],
  ].forEach(([label, target]) => {
    const button = makeButton(label, "profile-outline-button", "profile-section-nav", `${label}(으)로 이동`);
    button.dataset.profileNavTarget = target;
    if (target === "awards") button.classList.add("is-external");
    outline.append(button);
  });
  const sections = document.createElement("div");
  sections.className = "profile-editor-sections";
  const basic = profileStaticSection("기본 정보", (grid) => {
    appendProfilePair(grid, "이름", ["identity"], "name_en", "name_ko");
    appendProfilePair(grid, "약력", ["intro"], "short_en", "short_ko", { textarea: true });
    appendProfilePair(grid, "소속", ["identity"], "affiliation_en", "affiliation_ko");
    appendProfilePair(grid, "직함", ["profile_card"], "title_en", "title_ko");
    appendProfilePair(grid, "위치", ["identity"], "location_en", "location_ko", { textarea: true });
  }, { id: "profile-section-basic", className: "profile-primary-section" });
  basic.append(
    profileRepeatSection("credentials", { nested: true }),
    renderProfileMediaSection(),
  );
  const contact = profileStaticSection("연락처", (grid) => {
    ["email", "github", "orcid", "scholar", "linkedin"].forEach((field) => (
      grid.append(profileField(field === "linkedin" ? "LinkedIn" : field, ["contact", field]))
    ));
  }, { id: "profile-section-contact" });
  sections.append(basic, contact);
  ["experience", "education", "scholarships", "certifications", "teaching", "patents", "skills"]
    .forEach((section) => sections.append(profileRepeatSection(section, { id: `profile-section-${section}` })));
  container.append(outline, sections);
}

function cvSectionZone(section) {
  return state.draft?.cv_sections.includes(section) ? "visible" : "hidden";
}

function cvHierarchySections() {
  return [...(state.draft?.cv_sections || []), ...(state.draft?.hidden_cv_sections || [])];
}

function cvSectionItems(section) {
  if (["experience", "education", "teaching", "scholarships", "certifications", "patents", "skills"].includes(section)) {
    return state.profileDraft?.[section] || [];
  }
  if (section === "projects") return state.projectsDraft || [];
  if (section === "publications") return sortedPublications(state.publicationsDraft || []);
  if (section === "software") return state.softwareDraft || [];
  if (section === "awards") return sortedAwards(state.awardsDraft || []);
  if (section === "academic_activities") return state.academicActivitiesDraft || [];
  return [];
}

function cvDisplayValues(...values) {
  return [...new Set(values.map((value) => String(value || "").trim()).filter(Boolean))].join(" / ");
}

function cvDetailLine(label, ...values) {
  const value = cvDisplayValues(...values);
  return value ? `${label}: ${value}` : "";
}

function cvYearOrDate(value, fallback) {
  const normalized = String(value || "").trim();
  return normalized.match(/^\d{4}/)?.[0] || normalized || fallback;
}

function cvPublicationAuthors(item) {
  const keyedAuthors = (item.author_keys || []).map((key) => {
    const person = (state.peopleDraft || []).find((candidate) => candidate._clientKey === key);
    return person ? personNameLabel(person) : item._unresolvedAuthorIDs?.[key] || "";
  });
  const idAuthors = (item.author_ids || []).map((id) => {
    const person = (state.peopleDraft || []).find((candidate) => candidate.id === id);
    return person ? personNameLabel(person) : id;
  });
  return cvDisplayValues(...(keyedAuthors.length ? keyedAuthors : idAuthors));
}

function cvItemPresentation(section, item) {
  const summary = (...values) => values.map((value) => String(value || "").trim()).find(Boolean) || "이름 미입력";
  const period = (...values) => summary(...values, "기간 미입력");
  const details = [];
  let category = CV_SECTION_LABELS[section] || section;
  let label = "";

  if (section === "experience") {
    category = period(item.period_ko, item.period_en);
    label = cvDisplayValues(item.title_ko || item.title_en, item.institution_ko || item.institution_en);
    details.push(
      cvDetailLine("역할", item.title_ko, item.title_en),
      cvDetailLine("소속", item.institution_ko, item.institution_en),
      cvDetailLine("기간", item.period_ko, item.period_en),
    );
  } else if (section === "education") {
    category = period(item.period);
    label = cvDisplayValues(item.degree_ko || item.degree_en, item.institution_ko || item.institution_en);
    details.push(
      cvDetailLine("학위", item.degree_ko, item.degree_en),
      cvDetailLine("기관", item.institution_ko, item.institution_en),
      cvDetailLine("기간", item.period),
      cvDetailLine("지도교수", item.advisor_ko, item.advisor_en),
      cvDetailLine("논문", item.thesis_ko, item.thesis_en),
      cvDetailLine("학위논문 링크", item.thesis_url),
    );
    if (Number.isFinite(item.gpa?.value) && Number.isFinite(item.gpa?.max) && item.gpa.value >= 0 && item.gpa.max > 0 && item.gpa.value <= item.gpa.max) {
      const options = { minimumFractionDigits: 1, maximumFractionDigits: 20, useGrouping: false };
      details.push(`GPA: ${item.gpa.value.toLocaleString("en-US", options)} / ${item.gpa.max.toLocaleString("en-US", options)}`);
    }
  } else if (section === "projects") {
    category = projectThemeLabel(item.theme);
    label = summary(item.title_ko, item.title_en);
    details.push(
      cvDetailLine("제목", item.title_ko, item.title_en),
      cvDetailLine("테마", projectThemeLabel(item.theme)),
      cvDetailLine("기간", [item.start_date, item.end_date].filter(Boolean).join(" – ")),
      cvDetailLine("지원기관", item.funder_ko, item.funder_en),
    );
  } else if (section === "publications") {
    category = publicationTypeLabel(item.publication_type);
    label = summary(item.title_ko, item.title_en);
    details.push(
      cvDetailLine("제목", item.title_ko, item.title_en),
      cvDetailLine("저자", cvPublicationAuthors(item)),
      cvDetailLine("게재처", publicationMetadata(item).venue_kr, publicationMetadata(item).venue_en),
      cvDetailLine("상태", publicationStatusLabel(item)),
      cvDetailLine("주제", topicLabel(item.topic)),
      cvDetailLine("DOI", item.doi),
    );
  } else if (section === "software") {
    category = softwareStageLabel(item.stage);
    label = summary(item.name);
    details.push(
      cvDetailLine("상태", softwareStageLabel(item.stage)),
      cvDetailLine("기술", ...(item.technologies || []).map((technology) => technology.value)),
      cvDetailLine("링크", `${(item.links || []).length}개`),
    );
  } else if (section === "awards") {
    category = cvYearOrDate(item.date, "날짜 미입력");
    label = summary(item.title_ko, item.title_en);
    details.push(
      cvDetailLine("수상", item.title_ko, item.title_en),
      cvDetailLine("기관", item.organization_ko, item.organization_en),
      cvDetailLine("날짜", item.date),
    );
  } else if (section === "academic_activities") {
    const dateRange = academicActivityPeriod(item);
    label = academicActivityTitle(item);
    if (item.category === "reviews") {
      const dates = (item.completed_date_rows || []).map((row) => row.value).filter(Boolean);
      category = `${dates.length}회`;
      details.push(
        cvDetailLine("저널", item.journal_ko, item.journal_en),
        cvDetailLine("리뷰", `${dates.length}회`),
        cvDetailLine("완료일", ...dates),
      );
    } else if (item.category === "invited_talks") {
      category = dateRange;
      details.push(
        cvDetailLine("행사", item.event_ko, item.event_en),
        cvDetailLine("주제", item.topic_ko, item.topic_en),
        cvDetailLine("날짜", item.date),
      );
    } else if (item.category === "conference_service") {
      category = dateRange;
      details.push(
        cvDetailLine("학술대회", item.conference_ko, item.conference_en),
        cvDetailLine("역할", item.role_ko, item.role_en),
        cvDetailLine("기간", dateRange),
      );
    } else if (item.category === "professional_service") {
      category = dateRange;
      details.push(
        cvDetailLine("기관", item.organization_ko, item.organization_en),
        cvDetailLine("역할", item.role_ko, item.role_en),
        cvDetailLine("기간", dateRange),
      );
    } else {
      category = dateRange;
      details.push(
        cvDetailLine("저널", item.journal_ko, item.journal_en),
        cvDetailLine("역할", item.role_ko, item.role_en),
        cvDetailLine("기간", dateRange),
      );
    }
  } else if (section === "patents") {
    category = `${PATENT_TYPES.find(({ value }) => value === item.type)?.label || "종류 미입력"} · ${patentStatusLabel(item)}`;
    label = summary(item.title_ko, item.title_en);
    details.push(
      cvDetailLine("제목", item.title_ko, item.title_en),
      cvDetailLine("발명자", patentInventorNames(item)),
      cvDetailLine("출원인", item.applicant_ko, item.applicant_en),
      cvDetailLine("출원번호", item.application_number),
      cvDetailLine("출원일", item.application_date),
      cvDetailLine("공개번호", item.publication_number),
      cvDetailLine("공개일", item.publication_date),
      cvDetailLine("등록번호", item.registration_number),
      cvDetailLine("등록일", item.registration_date),
      cvDetailLine("URL", item.url),
    );
  } else if (section === "teaching") {
    category = period(item.period);
    label = summary(item.title_ko, item.title_en);
    details.push(
      cvDetailLine("교육", item.title_ko, item.title_en),
      cvDetailLine("기간", item.period),
      cvDetailLine("상세", item.detail_ko, item.detail_en),
    );
  } else if (section === "scholarships") {
    category = period(item.period_ko, item.period_en);
    label = summary(item.name_ko, item.name_en);
    details.push(
      cvDetailLine("장학", item.name_ko, item.name_en),
      cvDetailLine("기간", item.period_ko, item.period_en),
      cvDetailLine("금액", scholarshipAmountLabel(item.amount)),
      cvDetailLine("요약", item.summary_ko, item.summary_en),
    );
  } else if (section === "certifications") {
    category = cvYearOrDate(item.date, "날짜 미입력");
    label = summary(item.name_ko, item.name_en);
    details.push(
      cvDetailLine("자격", item.name_ko, item.name_en),
      cvDetailLine("발급기관", item.issuer_ko, item.issuer_en),
      cvDetailLine("날짜", item.date),
    );
  } else {
    category = "기술";
    label = summary(item.name);
    details.push(cvDetailLine("기술", item.name), cvDetailLine("상세", item.detail_ko, item.detail_en));
  }

  return {
    category,
    label: label || "이름 미입력",
    detail: details.filter(Boolean).join("\n") || "상세 정보가 없습니다.",
  };
}

function cvPublicationSearchText(item, presentation) {
  return [
    presentation.category,
    presentation.label,
    presentation.detail,
    publicationMetadata(item).venue_kr,
    publicationMetadata(item).venue_en,
    item.doi,
  ].map((value) => String(value || "").toLocaleLowerCase("ko")).join(" ");
}

function updateCVPublicationFilter() {
  const list = document.querySelector("#cv-hierarchy-list");
  if (!list) return;
  const query = state.cvPublicationSearch.trim().toLocaleLowerCase("ko");
  let visibleCount = 0;
  const rows = list.querySelectorAll("[data-cv-publication-search-text]");
  rows.forEach((row) => {
    const matches = !query || row.dataset.cvPublicationSearchText.includes(query);
    row.hidden = !matches;
    if (matches) visibleCount += 1;
  });
  const result = list.querySelector("[data-cv-publication-result-count]");
  if (result) result.textContent = query ? `${visibleCount}/${rows.length}` : `${rows.length}`;
  const empty = list.querySelector("[data-cv-publication-empty]");
  if (empty) empty.hidden = !query || visibleCount > 0;
}

function makeCVItemRow(section, item, index) {
  const row = document.createElement("li");
  const toggle = document.createElement("button");
  toggle.type = "button";
  toggle.className = "cv-item-toggle";
  toggle.dataset.cvItemSection = section;
  toggle.dataset.cvItemKey = item._clientKey;
  toggle.setAttribute("aria-pressed", String(item.visible_in_CV !== false));
  const presentation = cvItemPresentation(section, item);
  if (section === "publications") {
    row.dataset.cvPublicationSearchText = cvPublicationSearchText(item, presentation);
  }
  const category = document.createElement("span");
  category.className = "cv-item-category";
  category.textContent = presentation.category;
  const labelText = document.createElement("span");
  labelText.className = "cv-item-label";
  labelText.textContent = presentation.label;
  const detail = document.createElement("span");
  detail.className = "cv-item-detail";
  detail.id = `cv-item-detail-${section}-${index}`;
  detail.setAttribute("role", "tooltip");
  detail.textContent = presentation.detail;
  toggle.setAttribute("aria-describedby", detail.id);
  toggle.setAttribute(
    "aria-label",
    `${presentation.category}, ${presentation.label}, CV ${item.visible_in_CV !== false ? "포함" : "제외"}`,
  );
  toggle.title = `${presentation.category}\n${presentation.label}\n${presentation.detail}`;
  toggle.append(category, labelText, detail);
  toggle.disabled = state.loading || state.saving || state.discarding;
  row.append(toggle);
  return row;
}

function makeCVPatentGroups(items) {
  const groups = document.createElement("ul");
  groups.className = "cv-academic-activity-groups";
  PATENT_TYPES.forEach((type) => [
    { value: "filed", label: "출원" },
    { value: "published", label: "공개" },
    { value: "granted", label: "등록" },
  ].forEach((status) => {
    const typeItems = items.filter((item) => item.type === type.value && patentStatusLabel(item) === status.label);
    if (!typeItems.length) return;
    const group = document.createElement("li");
    group.className = "cv-academic-activity-group";
    group.dataset.cvPatentType = type.value;
    group.dataset.cvPatentStatus = status.value;
    const heading = document.createElement("div");
    heading.className = "cv-academic-activity-group-heading";
    const label = document.createElement("h4");
    label.textContent = `${type.label} [${status.label}]`;
    const count = document.createElement("span");
    count.className = "count-chip";
    count.textContent = String(typeItems.length);
    const list = document.createElement("ul");
    list.className = "cv-item-list";
    typeItems.forEach((item) => list.append(makeCVItemRow("patents", item, items.indexOf(item))));
    heading.append(label, count);
    group.append(heading, list);
    groups.append(group);
  }));
  return groups;
}

function makeCVAcademicActivityGroups(items) {
  const groups = document.createElement("ul");
  groups.className = "cv-academic-activity-groups";
  ACADEMIC_ACTIVITY_CATEGORIES.forEach((category) => {
    const categoryItems = items.filter((item) => item.category === category.value);
    const group = document.createElement("li");
    group.className = "cv-academic-activity-group";
    group.dataset.cvAcademicActivityCategory = category.value;
    const heading = document.createElement("div");
    heading.className = "cv-academic-activity-group-heading";
    const label = document.createElement("h4");
    label.textContent = category.label;
    const count = document.createElement("span");
    count.className = "count-chip";
    count.textContent = String(categoryItems.length);
    heading.append(label, count);
    const itemList = document.createElement("ul");
    itemList.className = "cv-item-list cv-academic-activity-item-list";
    if (!categoryItems.length) {
      const empty = document.createElement("li");
      empty.className = "cv-item-empty";
      empty.textContent = "항목 없음";
      itemList.append(empty);
    } else {
      categoryItems.forEach((item) => itemList.append(
        makeCVItemRow("academic_activities", item, items.indexOf(item)),
      ));
    }
    group.append(heading, itemList);
    groups.append(group);
  });
  return groups;
}

function makeCVPublicationTools(items) {
  const tools = document.createElement("div");
  tools.className = "cv-publication-tools";
  const searchLabel = document.createElement("label");
  searchLabel.className = "cv-publication-search";
  const searchText = document.createElement("span");
  searchText.className = "sr-only";
  searchText.textContent = "CV 논문 검색";
  const search = document.createElement("input");
  search.type = "search";
  search.className = "list-filter-control";
  search.placeholder = "제목, 저자 또는 게재지 검색";
  search.autocomplete = "off";
  search.value = state.cvPublicationSearch;
  search.dataset.cvPublicationSearch = "";
  search.disabled = state.loading || state.saving || state.discarding;
  searchLabel.append(searchText, search);
  const resultCount = document.createElement("span");
  resultCount.className = "count-chip cv-publication-result-count";
  resultCount.dataset.cvPublicationResultCount = "";
  resultCount.textContent = String(items.length);
  resultCount.setAttribute("aria-live", "polite");
  const actions = document.createElement("span");
  actions.className = "cv-publication-bulk-actions";
  actions.append(
    makeButton("전체 선택", "text-button cv-bulk-button", "cv-items-select-all", "모든 논문을 CV에 포함"),
    makeButton("전체 해제", "text-button cv-bulk-button", "cv-items-clear-all", "모든 논문을 CV에서 제외"),
  );
  tools.append(searchLabel, resultCount, actions);
  return tools;
}

function renderWebPages() {
  const list = document.querySelector("#web-pages-list");
  if (!list || !state.draft) return;
  list.replaceChildren();
  const hidden = new Set(state.draft.hidden_web_pages);
  Object.entries(WEB_PAGE_LABELS).forEach(([page, label]) => {
    const visible = !hidden.has(page);
    const row = document.createElement("li");
    row.className = `web-page-row${visible ? "" : " is-page-hidden"}`;
    row.dataset.webPage = page;
    const title = document.createElement("span");
    title.className = "web-page-label";
    title.textContent = label;
    const toggle = makeButton(visible ? "노출" : "숨김", "visibility-switch", "web-page-toggle", `${label} ${visible ? "숨기기" : "노출하기"}`);
    toggle.dataset.editorControl = "";
    toggle.setAttribute("role", "switch");
    toggle.setAttribute("aria-checked", String(visible));
    row.append(title, toggle);
    list.append(row);
  });
}

function toggleWebPage(page) {
  if (!Object.hasOwn(WEB_PAGE_LABELS, page)) return;
  const hidden = new Set(state.draft.hidden_web_pages);
  if (hidden.has(page)) hidden.delete(page);
  else hidden.add(page);
  state.draft.hidden_web_pages = Object.keys(WEB_PAGE_LABELS).filter((value) => hidden.has(value));
  renderWebPages();
  updateDirtyState();
  document.querySelector(`[data-web-page="${page}"] [data-action="web-page-toggle"]`)?.focus({ preventScroll: true });
}

function renderCVHierarchy() {
  renderWebPages();
  const list = document.querySelector("#cv-hierarchy-list");
  if (!list || !state.profileDraft || !state.draft) return;
  list.replaceChildren();
  const sections = cvHierarchySections();
  const sectionCount = document.querySelector("#cv-section-count");
  if (sectionCount) sectionCount.textContent = String(sections.length);

  sections.forEach((section) => {
    const zone = cvSectionZone(section);
    const isVisible = zone === "visible";
    const expanded = state.cvExpandedSections.has(section);
    const items = cvSectionItems(section);
    const zoneItems = state.draft[isVisible ? "cv_sections" : "hidden_cv_sections"];
    const zoneIndex = zoneItems.indexOf(section);
    const sectionLabel = CV_SECTION_LABELS[section] || section;

    const sectionRow = document.createElement("li");
    sectionRow.className = `cv-hierarchy-section${isVisible ? "" : " is-section-hidden"}`;
    sectionRow.dataset.cvSection = section;
    sectionRow.dataset.cvZone = zone;

    const heading = document.createElement("div");
    heading.className = "cv-hierarchy-section-heading";
    const handle = makeButton("", "drag-handle", "drag-cv-section", `${sectionLabel} 순서 이동`);
    handle.draggable = !(state.loading || state.saving || state.discarding);
    handle.setAttribute("aria-label", `${sectionLabel} 드래그하여 순서 이동`);

    const disclosure = makeButton("", "cv-section-disclosure", "cv-section-expand", `${sectionLabel} 항목 ${expanded ? "접기" : "펼치기"}`);
    disclosure.setAttribute("aria-expanded", String(expanded));
    disclosure.setAttribute("aria-controls", `cv-section-items-${section}`);
    const chevron = document.createElement("span");
    chevron.className = "cv-section-chevron";
    chevron.setAttribute("aria-hidden", "true");
    const label = document.createElement("span");
    label.className = "cv-section-label";
    label.textContent = sectionLabel;
    const count = document.createElement("span");
    count.className = "count-chip cv-item-count";
    count.textContent = String(items.length);
    count.setAttribute("aria-label", `${items.length}개 항목`);
    disclosure.append(chevron, label, count);

    if (items.length && cvSectionItemCount(section) === 0) {
      const emptySelection = document.createElement("span");
      emptySelection.className = "cv-section-empty-label";
      emptySelection.textContent = "선택 없음";
      disclosure.append(emptySelection);
    } else if (!items.length) {
      const emptyItems = document.createElement("span");
      emptyItems.className = "cv-section-empty-label";
      emptyItems.textContent = "항목 없음";
      disclosure.append(emptyItems);
    }

    const actions = document.createElement("span");
    actions.className = "cv-section-actions";
    const up = makeButton("↑", "icon-button cv-order-button", "cv-section-up", `${sectionLabel} 위로`);
    const down = makeButton("↓", "icon-button cv-order-button", "cv-section-down", `${sectionLabel} 아래로`);
    up.disabled = state.loading || state.saving || state.discarding || zoneIndex <= 0;
    down.disabled = state.loading || state.saving || state.discarding || zoneIndex >= zoneItems.length - 1;
    const visibility = makeButton(isVisible ? "노출" : "숨김", "visibility-switch cv-section-visibility", "cv-section-toggle", `${sectionLabel} ${isVisible ? "숨기기" : "노출하기"}`);
    visibility.setAttribute("role", "switch");
    visibility.setAttribute("aria-checked", String(isVisible));
    actions.append(up, down, visibility);
    heading.append(handle, disclosure, actions);

    const body = document.createElement("div");
    body.className = "cv-hierarchy-section-body";
    body.id = `cv-section-items-${section}`;
    body.hidden = !expanded;
    if (section === "publications" && items.length) body.append(makeCVPublicationTools(items));
    if (section === "academic_activities") {
      body.append(makeCVAcademicActivityGroups(items));
    } else if (section === "patents") {
      body.append(makeCVPatentGroups(items));
    } else {
      const itemList = document.createElement("ul");
      itemList.className = "cv-item-list";
      if (!items.length) {
        const empty = document.createElement("li");
        empty.className = "cv-item-empty";
        empty.textContent = "항목이 없습니다.";
        itemList.append(empty);
      } else {
        items.forEach((item, index) => itemList.append(makeCVItemRow(section, item, index)));
        if (section === "publications") {
          const noResults = document.createElement("li");
          noResults.className = "cv-item-empty";
          noResults.dataset.cvPublicationEmpty = "";
          noResults.textContent = "검색 결과가 없습니다.";
          noResults.hidden = true;
          itemList.append(noResults);
        }
      }
      body.append(itemList);
    }
    sectionRow.append(heading, body);
    list.append(sectionRow);
  });
  updateCVPublicationFilter();
}

function mutateProfileItem(action, section, index) {
  const items = profileRepeatItems(section);
  if (!items) return;
  let addedKey = "";
  if (action === "add") {
    const item = defaultProfileItem(section);
    items.push(item);
    addedKey = item._clientKey;
    state.profileExpandedRows.add(item._clientKey);
  }
  if (action === "delete" && index >= 0) {
    state.profileExpandedRows.delete(items[index]?._clientKey);
    items.splice(index, 1);
  }
  if (action === "up" && index > 0) [items[index - 1], items[index]] = [items[index], items[index - 1]];
  if (action === "down" && index >= 0 && index < items.length - 1) [items[index + 1], items[index]] = [items[index], items[index + 1]];
  const autoHidden = section !== "credentials" && syncEmptyMainSections();
  if (autoHidden) {
    setProfileSectionFeedback(`${SECTION_LABELS[section]}: 노출할 완성된 항목이 없어 자동으로 숨겼습니다.`);
  }
  renderSectionList("visible");
  renderSectionList("hidden");
  renderProfileContentEditor();
  syncCVSectionsAfterItemsChange();
  renderCVHierarchy();
  updateDirtyState();
  if (addedKey) {
    requestAnimationFrame(() => {
      document.querySelector(`.profile-repeat-row[data-profile-key="${addedKey}"] input, .profile-repeat-row[data-profile-key="${addedKey}"] textarea`)?.focus();
    });
  }
}

function findCVItem(section, clientKey) {
  return cvSectionItems(section).find((item) => item._clientKey === clientKey);
}

function setCVItemsVisibility(section, visible) {
  const items = cvSectionItems(section);
  if (!items.length) return;
  items.forEach((item) => {
    item.visible_in_CV = visible;
  });
  setCVSectionFeedback("");
  syncCVSectionsAfterItemsChange();
  renderCVHierarchy();
  updateDirtyState();
  announce(`${CV_SECTION_LABELS[section] || section} 항목을 모두 ${visible ? "선택" : "해제"}했습니다.`);
}

function renderSectionList(zone) {
  const key = zone === "visible" ? "main_page_sections" : "hidden_main_page_sections";
  const list = document.querySelector(`#${zone}-section-list`);
  if (!list) return;
  list.replaceChildren();

  state.draft[key].forEach((section, index, sections) => {
    const itemCount = mainSectionItemCount(section);
    const row = document.createElement("li");
    row.className = `section-row profile-section-order-row${zone === "hidden" ? " is-hidden-section" : ""}${itemCount === 0 ? " is-empty" : ""}`;
    row.dataset.section = section;
    row.dataset.zone = zone;

    const handle = makeButton("", "drag-handle", "drag-section", `${SECTION_LABELS[section]} 순서 이동`);
    handle.draggable = !state.saving;
    handle.setAttribute("aria-label", `${SECTION_LABELS[section]} 드래그하여 이동`);

    const copy = document.createElement("span");
    copy.className = "section-row-copy";
    const label = document.createElement("span");
    label.className = "row-label";
    label.textContent = SECTION_LABELS[section] || section;
    copy.append(label);
    if (itemCount === 0) {
      const empty = document.createElement("span");
      empty.className = "section-empty-label";
      empty.textContent = "항목 없음";
      copy.append(empty);
    }

    const actions = document.createElement("span");
    actions.className = "row-actions";
    const up = makeButton("↑", "icon-button", "section-up", "위로");
    const down = makeButton("↓", "icon-button", "section-down", "아래로");
    const toggle = makeButton(zone === "visible" ? "노출" : "숨김", "visibility-switch", "section-toggle", `${SECTION_LABELS[section]} ${zone === "visible" ? "숨기기" : "노출하기"}`);
    toggle.setAttribute("role", "switch");
    toggle.setAttribute("aria-checked", String(zone === "visible"));
    up.disabled = state.saving || index === 0;
    down.disabled = state.saving || index === sections.length - 1;
    if (zone === "hidden" && itemCount === 0) {
      toggle.disabled = true;
      toggle.title = `${SECTION_LABELS[section]}에는 노출할 항목이 없습니다.`;
    }
    actions.append(up, down, toggle);
    row.append(handle, copy, actions);
    list.append(row);
  });

  const count = document.querySelector(`#${zone}-section-count`);
  if (count) count.textContent = String(state.draft[key].length);
}

function createLabelInput(value, key, placeholder) {
  const input = document.createElement("input");
  input.className = "label-input";
  input.type = "text";
  input.maxLength = 100;
  input.value = value || "";
  input.placeholder = placeholder;
  input.autocomplete = "off";
  input.disabled = state.saving || state.discarding;
  input.dataset.validationKey = key;
  return input;
}

function taxonomyUsage(_kind, item) {
  if (!item.id) return 0;
  if (state.projectsDraft && state.publicationsDraft) {
    return state.projectsDraft.filter((project) => project.theme === item.id).length
      + state.publicationsDraft.filter((publication) => publication.topic === item.id).length;
  }
  return (state.usage["project_themes"]?.[item.id] || 0) + (state.usage["publication_topics"]?.[item.id] || 0);
}

function taxonomyDefaultColor(kind, index) {
  const config = taxonomyConfig[kind];
  if (config.colors[index]) return config.colors[index];
  const context = document.createElement("canvas").getContext("2d");
  context.fillStyle = `hsl(${(index * 137.508).toFixed(2)} 44% 38%)`;
  return context.fillStyle;
}

function makeTaxonomyColorControl(kind, item, index) {
  const config = taxonomyConfig[kind];
  const control = document.createElement("div");
  control.className = "taxonomy-color-control";
  const trigger = makeButton("", "taxonomy-color-trigger", "taxonomy-color-open", `${config.itemName} 색상 팔레트 열기`);
  trigger.dataset.editorControl = "";
  trigger.dataset.kind = kind;
  trigger.dataset.clientKey = item._clientKey;
  trigger.setAttribute("aria-haspopup", "dialog");
  trigger.setAttribute("aria-controls", "taxonomy-color-popover");
  trigger.setAttribute("aria-expanded", "false");
  trigger.style.setProperty("--taxonomy-selected-color", /^#[0-9a-f]{6}$/i.test((item.color || "").trim())
    ? item.color.trim() : taxonomyDefaultColor(kind, index));
  const swatch = document.createElement("span");
  swatch.className = "taxonomy-color-preview";
  swatch.setAttribute("aria-hidden", "true");
  const arrow = document.createElement("span");
  arrow.className = "taxonomy-color-arrow";
  arrow.textContent = "▾";
  arrow.setAttribute("aria-hidden", "true");
  trigger.append(swatch, arrow);
  const code = createLabelInput(item.color, validationKey(kind, item._clientKey, "color"), "자동");
  code.maxLength = 7;
  code.setAttribute("aria-label", `${config.itemName} 색상 코드 (비워 두면 자동)`);
  code.spellcheck = false;
  code.dataset.kind = kind;
  code.dataset.clientKey = item._clientKey;
  code.dataset.field = "color";
  control.append(trigger, code);
  return control;
}

function renderTaxonomy(kind) {
  closeTaxonomyColorPopover();
  const config = taxonomyConfig[kind];
  const list = document.querySelector(`#${config.listElement}`);
  const items = state.draft[config.listKey];
  list.replaceChildren();

  items.forEach((item, index) => {
    const row = document.createElement("li");
    row.className = "taxonomy-row";
    row.dataset.kind = kind;
    row.dataset.clientKey = item._clientKey;

    const handle = makeButton("", "drag-handle", "drag-taxonomy", `${config.itemName} 순서 이동`);
    handle.draggable = !state.saving;
    handle.setAttribute("aria-label", `${config.itemName} 드래그하여 순서 이동`);

    const englishKey = validationKey(kind, item._clientKey, "label_en");
    const koreanKey = validationKey(kind, item._clientKey, "label_ko");
    const english = createLabelInput(item.label_en, englishKey, "영문");
    const korean = createLabelInput(item.label_ko, koreanKey, "국문");
    english.setAttribute("aria-label", `${config.itemName} 이름, 영문`);
    korean.setAttribute("aria-label", `${config.itemName} 이름, 국문`);
    english.dataset.kind = kind;
    english.dataset.clientKey = item._clientKey;
    english.dataset.field = "label_en";
    korean.dataset.kind = kind;
    korean.dataset.clientKey = item._clientKey;
    korean.dataset.field = "label_ko";

    const count = taxonomyUsage(kind, item);
    const usage = document.createElement("span");
    usage.className = "usage-chip";
    usage.textContent = count ? `${count}개 사용` : "미사용";

    const actions = document.createElement("span");
    actions.className = "taxonomy-actions";
    const up = makeButton("↑", "icon-button", "taxonomy-up", "위로");
    const down = makeButton("↓", "icon-button", "taxonomy-down", "아래로");
    const remove = makeButton("삭제", "text-button delete-button", "taxonomy-delete", `${config.itemName} 삭제`);
    up.disabled = state.saving || index === 0;
    down.disabled = state.saving || index === items.length - 1;
    remove.disabled = state.saving || count > 0;
    if (count > 0) remove.title = `${count}개 데이터에서 사용 중입니다. 먼저 해당 데이터를 재분류해 주세요.`;
    actions.append(up, down, remove);

    const error = document.createElement("div");
    error.className = "row-error";
    error.dataset.errorScope = `${kind}|${item._clientKey}|`;

    row.append(handle, english, korean, makeTaxonomyColorControl(kind, item, index), usage, actions, error);
    list.append(row);
  });

  renderFallback(kind);
}

function renderFallback(kind) {
  const config = taxonomyConfig[kind];
  const fallback = state.draft[config.fallbackKey];
  const container = document.querySelector(`#${config.fallbackElement}`);
  container.replaceChildren();

  const label = document.createElement("span");
  label.className = "fallback-label";
  label.textContent = "기타";
  const englishKey = fallbackValidationKey(kind, "label_en");
  const koreanKey = fallbackValidationKey(kind, "label_ko");
  const english = createLabelInput(fallback.label_en, englishKey, "영문");
  const korean = createLabelInput(fallback.label_ko, koreanKey, "국문");
  english.setAttribute("aria-label", `${config.itemName} 기타 이름, 영문`);
  korean.setAttribute("aria-label", `${config.itemName} 기타 이름, 국문`);
  english.dataset.fallbackKind = kind;
  english.dataset.field = "label_en";
  korean.dataset.fallbackKind = kind;
  korean.dataset.field = "label_ko";

  const error = document.createElement("div");
  error.className = "row-error";
  error.dataset.errorScope = `${kind}|fallback|`;
  container.append(label, english, korean, error);
}

function boardItem(clientKey) {
  const index = state.boardDraft.findIndex((item) => item._clientKey === clientKey);
  return { index, item: index >= 0 ? state.boardDraft[index] : null };
}

function boardDateLabel(item) {
  if (!item.start_date) return "날짜 미입력";
  return item.end_date ? `${item.start_date} – ${item.end_date}` : item.start_date;
}

function renderBoardList() {
  const list = document.querySelector("#board-item-list");
  const listColumn = list.closest(".master-detail-list");
  const listScrollTop = listColumn?.scrollTop || 0;
  const items = sortedBoardItems(state.boardDraft);
  if (state.selectedBoardKey && !state.boardDraft.some((item) => item._clientKey === state.selectedBoardKey)) {
    state.selectedBoardKey = "";
  }
  list.replaceChildren();
  items.forEach((item) => {
    const row = document.createElement("li");
    row.className = "board-item-row";
    row.dataset.boardKey = item._clientKey;

    const select = document.createElement("button");
    select.type = "button";
    select.className = "board-item-select";
    select.dataset.action = "board-select";
    select.disabled = state.saving;
    select.setAttribute("aria-label", `${item.title_ko || item.title_en || "새 게시글"} 선택`);
    applyListRowSelection(row, select, item._clientKey === state.selectedBoardKey);

    const date = document.createElement("span");
    date.className = "board-item-date";
    date.textContent = boardDateLabel(item);
    const title = document.createElement("span");
    title.className = "board-item-title";
    title.textContent = item.title_ko || item.title_en || "(제목 미입력)";
    title.title = title.textContent;
    const { badges, errorBadge } = makeListStateBadges(item, "boardErrorBadge");
    if ((item.media || []).length) {
      const mediaBadge = document.createElement("span");
      mediaBadge.className = "board-badge";
      mediaBadge.textContent = `사진 ${item.media.length}개`;
      badges.append(mediaBadge);
    }
    badges.append(errorBadge);
    select.append(date, title, badges);

    const remove = makeButton("삭제", "text-button delete-button board-item-delete", "board-delete", "게시글 삭제");
    remove.setAttribute("aria-label", `${item.title_ko || item.title_en || "새 게시글"} 삭제`);
    row.append(select, remove);
    list.append(row);
  });
  document.querySelector("#board-item-count").textContent = String(items.length);
  if (listColumn) listColumn.scrollTop = listScrollTop;
}

function makeBoardField(item, labelText, field, options = {}) {
  const label = document.createElement("label");
  label.className = `board-field${options.full ? " board-field-full" : ""}`;
  const title = document.createElement("span");
  title.textContent = labelText;
  const control = options.textarea ? document.createElement("textarea") : document.createElement("input");
  control.className = options.textarea ? "board-textarea" : "board-input";
  if (!options.textarea) control.type = options.type || "text";
  control.value = item[field] || "";
  control.maxLength = options.maxLength || (options.textarea ? 20000 : 300);
  control.disabled = state.saving || state.discarding;
  control.autocomplete = "off";
  control.dataset.boardKey = item._clientKey;
  control.dataset.boardField = field;
  control.dataset.validationKey = boardValidationKey(item._clientKey, field);
  const error = document.createElement("span");
  error.className = "board-field-error";
  error.dataset.errorScope = boardValidationKey(item._clientKey, field);
  error.id = `board-error-${item._clientKey.replace(/[^a-z0-9-]/gi, "-")}-${field}`;
  control.setAttribute("aria-describedby", error.id);
  label.append(title, control, error);
  return label;
}

function makeBilingualFieldGroup(titleText, englishField, koreanField, options = {}) {
  const group = document.createElement("div");
  group.className = `bilingual-field-group${options.stacked ? " is-stacked" : ""}`;
  group.setAttribute("role", "group");
  group.setAttribute("aria-label", titleText);
  const title = document.createElement("span");
  title.className = "bilingual-field-title";
  title.textContent = titleText;
  if (options.hideTitle) title.classList.add("sr-only");
  const pair = { english: englishField, korean: koreanField };
  const fields = makeLanguagePairFields("bilingual-field-grid", ({ language, key }) => {
    const field = pair[key];
    field.querySelector(":scope > span:first-child")?.classList.add("language-marker");
    field.querySelector("input, textarea, select")?.setAttribute("aria-label", `${titleText}, ${language}`);
    return field;
  });
  group.append(title, fields);
  return group;
}

function formatBytes(size) {
  if (!Number.isFinite(Number(size)) || Number(size) < 0) return "";
  size = Number(size);
  if (size < 1024) return `${size} B`;
  if (size < 1024 * 1024) return `${Math.round(size / 1024)} KB`;
  return `${(size / (1024 * 1024)).toFixed(1)} MB`;
}

function boardMediaName(media, index) {
  if (media.original_name) return media.original_name;
  const source = String(media.src || "").split(/[\\/]/).pop();
  if (source) return source;
  return media.type === "video" ? `동영상 ${index + 1}` : `사진 ${index + 1}`;
}

let mediaPreviewReturnFocus = null;

function mediaPreviewURL(media) {
  const query = new URLSearchParams();
  if (media.stage_token) query.set("stage_token", media.stage_token);
  else if (media.src) query.set("src", media.src);
  else return "";
  return `/__editor_media?${query}`;
}

function makeMediaPreviewButton(media, label, className = "") {
  const preview = document.createElement("button");
  preview.type = "button";
  preview.className = `board-media-preview media-preview-trigger ${className}`.trim();
  preview.setAttribute("aria-label", `${label} 확대 보기`);
  preview.setAttribute("aria-haspopup", "dialog");
  preview.disabled = state.saving || state.discarding || !mediaPreviewURL(media);
  if (media.preview_url) {
    const image = document.createElement("img");
    image.src = media.preview_url;
    image.alt = "";
    image.draggable = false;
    preview.append(image);
  } else {
    preview.textContent = media.type === "video" ? "동영상" : "사진";
  }
  preview.addEventListener("click", (event) => {
    event.stopPropagation();
    openMediaPreview(media, label, preview);
  });
  return preview;
}

function openMediaPreview(media, label, trigger) {
  const dialog = document.querySelector("#media-preview-dialog");
  const source = mediaPreviewURL(media);
  if (!dialog || !source || state.saving || state.discarding) return;
  const content = dialog.querySelector("#media-preview-content");
  const status = dialog.querySelector("#media-preview-status");
  const caption = dialog.querySelector("#media-preview-caption");
  mediaPreviewReturnFocus = trigger || document.activeElement;
  dialog.setAttribute("aria-label", `${label} 확대 보기`);
  content.replaceChildren();
  content.setAttribute("aria-busy", "true");
  status.textContent = "불러오는 중…";
  status.hidden = false;
  caption.textContent = media.caption_ko || media.caption_en || "";
  caption.hidden = !caption.textContent;
  const isVideo = media.type === "video" && !media.stage_token;
  const element = document.createElement(isVideo ? "video" : "img");
  element.className = "media-preview-full";
  if (isVideo) {
    element.controls = true;
    element.preload = "metadata";
    element.playsInline = true;
    element.setAttribute("aria-label", label);
  } else {
    element.alt = media.caption_ko || media.caption_en || label;
  }
  const loaded = () => {
    if (!content.contains(element)) return;
    content.setAttribute("aria-busy", "false");
    status.textContent = "";
    status.hidden = true;
  };
  element.addEventListener(isVideo ? "loadedmetadata" : "load", loaded, { once: true });
  element.addEventListener("error", () => {
    if (!content.contains(element)) return;
    element.hidden = true;
    content.setAttribute("aria-busy", "false");
    status.textContent = "미디어를 불러올 수 없습니다.";
    status.hidden = false;
  }, { once: true });
  content.append(element);
  if (!dialog.open) dialog.showModal();
  dialog.querySelector("#media-preview-close").focus({ preventScroll: true });
  element.src = source;
}

function closeMediaPreview() {
  const dialog = document.querySelector("#media-preview-dialog");
  if (dialog?.open) dialog.close();
}

function makeBilingualCaptionGroup(titleText, english, korean, error) {
  const group = document.createElement("div");
  group.className = "bilingual-caption-group";
  group.setAttribute("role", "group");
  group.setAttribute("aria-label", titleText);
  const title = document.createElement("span");
  title.className = "bilingual-caption-title";
  title.textContent = titleText;
  const fields = document.createElement("div");
  fields.className = "bilingual-caption-grid";
  [["영문", english], ["국문", korean]].forEach(([labelText, input]) => {
    const label = document.createElement("label");
    label.className = "bilingual-caption-field";
    const language = document.createElement("span");
    language.className = "language-marker";
    language.textContent = labelText;
    label.append(language, input);
    fields.append(label);
  });
  group.append(title, fields, error);
  return group;
}

function renderBoardEditor() {
  const placeholder = document.querySelector("#board-editor-placeholder");
  const form = document.querySelector("#board-form");
  const found = boardItem(state.selectedBoardKey);
  form.replaceChildren();

  if (!found.item) {
    placeholder.hidden = false;
    form.hidden = true;
    return;
  }
  placeholder.hidden = true;
  form.hidden = false;

  const heading = document.createElement("div");
  heading.className = "board-form-heading";
  const headingTitle = document.createElement("h3");
  headingTitle.textContent = found.item._isNew
    ? "새 게시글"
    : found.item.title_ko || found.item.title_en || "(제목 미입력)";
  heading.append(headingTitle);
  form.append(heading);

  const grid = document.createElement("div");
  grid.className = "board-form-grid";
  grid.append(
    makeBoardField(found.item, "시작일", "start_date", { type: "date" }),
    makeBoardField(found.item, "종료일(선택)", "end_date", { type: "date" }),
    makeBilingualFieldGroup(
      "제목",
      makeBoardField(found.item, "영문", "title_en"),
      makeBoardField(found.item, "국문", "title_ko"),
    ),
    makeBilingualFieldGroup(
      "본문 (선택)",
      makeBoardField(found.item, "영문", "content_en", { textarea: true }),
      makeBoardField(found.item, "국문", "content_ko", { textarea: true }),
      { stacked: true },
    ),
  );

  const mediaSection = document.createElement("section");
  mediaSection.className = "board-media-section";
  const mediaHeading = document.createElement("div");
  mediaHeading.className = "board-media-heading";
  const mediaTitle = document.createElement("h4");
  mediaTitle.textContent = "사진";
  const mediaCount = document.createElement("span");
  mediaCount.className = "count-chip";
  mediaCount.textContent = String(found.item.media.length);
  mediaHeading.append(mediaTitle, mediaCount);

  const dropzone = document.createElement("div");
  dropzone.className = `board-dropzone${state.boardDropBusy ? " is-receiving" : ""}`;
  dropzone.dataset.boardMediaDropzone = "";
  dropzone.setAttribute("aria-label", "Explorer에서 사진을 드롭하는 영역");
  dropzone.setAttribute("aria-busy", String(state.boardDropBusy));
  const dropTitle = document.createElement("strong");
  dropTitle.textContent = state.boardDropBusy ? "사진을 준비하는 중…" : "사진을 이곳에 드래그하세요";
  dropzone.append(dropTitle);
  if (state.boardDropMessage) {
    const feedback = document.createElement("p");
    feedback.className = "board-drop-feedback";
    feedback.setAttribute("role", "status");
    feedback.textContent = state.boardDropMessage;
    dropzone.append(feedback);
  }

  const mediaList = document.createElement("ol");
  mediaList.className = "board-media-list";
  found.item.media.forEach((media, index) => {
    const row = document.createElement("li");
    row.className = "board-media-row";
    row.dataset.mediaKey = media._clientKey;
    const preview = makeMediaPreviewButton(media, boardMediaName(media, index));
    const fields = document.createElement("div");
    fields.className = "board-media-fields";
    const name = document.createElement("span");
    name.className = "board-media-name";
    const sizeLabel = formatBytes(media.size);
    name.textContent = `${boardMediaName(media, index)}${sizeLabel ? ` · ${sizeLabel}` : ""}`;
    const english = document.createElement("input");
    english.type = "text";
    english.className = "board-caption-input";
    english.maxLength = 500;
    english.value = media.caption_en || "";
    english.disabled = state.saving || state.discarding;
    english.dataset.boardKey = found.item._clientKey;
    english.dataset.mediaKey = media._clientKey;
    english.dataset.boardMediaField = "caption_en";
    english.dataset.validationKey = boardValidationKey(found.item._clientKey, `media:${media._clientKey}:caption_en`);
    english.setAttribute("aria-label", `${boardMediaName(media, index)} 영문 사진 설명`);
    const korean = document.createElement("input");
    korean.type = "text";
    korean.className = "board-caption-input";
    korean.maxLength = 500;
    korean.value = media.caption_ko || "";
    korean.disabled = state.saving || state.discarding;
    korean.dataset.boardKey = found.item._clientKey;
    korean.dataset.mediaKey = media._clientKey;
    korean.dataset.boardMediaField = "caption_ko";
    korean.dataset.validationKey = boardValidationKey(found.item._clientKey, `media:${media._clientKey}:caption_ko`);
    korean.setAttribute("aria-label", `${boardMediaName(media, index)} 국문 사진 설명`);
    const error = document.createElement("span");
    error.className = "board-caption-error";
    error.dataset.errorScope = boardValidationKey(found.item._clientKey, `media:${media._clientKey}:`);
    error.id = `board-media-error-${found.item._clientKey.replace(/[^a-z0-9-]/gi, "-")}-${media._clientKey.replace(/[^a-z0-9-]/gi, "-")}`;
    english.setAttribute("aria-describedby", error.id);
    korean.setAttribute("aria-describedby", error.id);
    fields.append(name, makeBilingualCaptionGroup("사진 설명 (선택)", english, korean, error));
    const remove = makeButton("제거", "text-button delete-button", "board-media-delete", "사진 제거");
    row.append(preview, fields, remove);
    mediaList.append(row);
  });
  if (!found.item.media.length) {
    const empty = document.createElement("p");
    empty.className = "board-media-empty";
    empty.textContent = "추가된 사진이 없습니다.";
    mediaSection.append(mediaHeading, dropzone, empty);
  } else {
    mediaSection.append(mediaHeading, dropzone, mediaList);
  }
  form.append(grid, mediaSection);
}

function contentDraft(collection) {
  return collection === "projects" ? state.projectsDraft : state.softwareDraft;
}

function selectedContentKey(collection) {
  return collection === "projects" ? state.selectedProjectKey : state.selectedSoftwareKey;
}

function setSelectedContentKey(collection, clientKey) {
  if (collection === "projects") state.selectedProjectKey = clientKey;
  else state.selectedSoftwareKey = clientKey;
}

function contentItem(collection, clientKey) {
  const items = contentDraft(collection) || [];
  const index = items.findIndex((item) => item._clientKey === clientKey);
  return { items, index, item: index >= 0 ? items[index] : null };
}

function contentSingular(collection) {
  return collection === "projects" ? "프로젝트" : "소프트웨어";
}

function moveManualItemToPosition(items, clientKey, destinationPosition, isMovable = () => true) {
  const movableSlots = [];
  const movableItems = [];
  items.forEach((item, index) => {
    if (!isMovable(item)) return;
    movableSlots.push(index);
    movableItems.push(item);
  });
  const sourcePosition = movableItems.findIndex((item) => item._clientKey === clientKey);
  if (sourcePosition < 0 || movableItems.length < 2) return false;
  const destination = Math.max(0, Math.min(Math.trunc(destinationPosition), movableItems.length - 1));
  if (sourcePosition === destination) return false;
  const [moved] = movableItems.splice(sourcePosition, 1);
  movableItems.splice(destination, 0, moved);
  movableSlots.forEach((slot, index) => {
    items[slot] = movableItems[index];
  });
  return true;
}

function projectThemeLabel(themeID) {
  const theme = (state.draft?.topics || []).find((item) => item.id === themeID);
  return theme?.label_ko || theme?.label_en || themeID || "테마 미선택";
}

function softwareStageLabel(stage) {
  return ({ release: "출시", preview: "미리보기", development: "개발 중" })[stage] || stage || "단계 미선택";
}

function renderContentList(collection) {
  const items = contentDraft(collection) || [];
  let selectedKey = selectedContentKey(collection);
  if (selectedKey && !items.some((item) => item._clientKey === selectedKey)) {
    selectedKey = "";
    setSelectedContentKey(collection, "");
  }
  const list = document.querySelector(`#${collection === "projects" ? "project" : "software"}-item-list`);
  const listColumn = list.closest(".master-detail-list");
  const listScrollTop = listColumn?.scrollTop || 0;
  list.dataset.manualOrderList = "content";
  list.dataset.manualOrderCollection = collection;
  list.replaceChildren();
  items.forEach((item, index) => {
    const singular = contentSingular(collection);
    const titleText = collection === "projects"
      ? item.title_ko || item.title_en || "(제목 미입력)"
      : item.name || "(이름 미입력)";
    const row = document.createElement("li");
    row.className = "board-item-row content-item-row manual-order-row";
    row.dataset.contentCollection = collection;
    row.dataset.contentKey = item._clientKey;

    const handle = makeButton("", "drag-handle list-drag-handle", "drag-content-item", `${titleText} 순서 끌어 이동`);
    handle.draggable = !handle.disabled;

    const select = document.createElement("button");
    select.type = "button";
    select.className = "board-item-select";
    select.dataset.action = "content-select";
    select.disabled = state.saving || state.discarding;
    select.setAttribute("aria-label", `${titleText} 선택`);
    applyListRowSelection(row, select, item._clientKey === selectedKey);

    const order = document.createElement("span");
    order.className = "board-item-date";
    order.textContent = collection === "projects"
      ? `${item.start_date || "날짜 미입력"} – ${item.end_date || "날짜 미입력"}`
      : `${index + 1}. ${softwareStageLabel(item.stage)}`;
    const title = document.createElement("span");
    title.className = "board-item-title";
    title.textContent = titleText;
    title.title = title.textContent;
    const subtitle = document.createElement("span");
    subtitle.className = "board-item-subtitle";
    subtitle.textContent = collection === "projects"
      ? item.funder_ko || item.funder_en || "발주처 미입력"
      : (item.technologies || []).map((technology) => technology.value).filter(Boolean).join(" · ") || "기술 미선택";
    subtitle.title = subtitle.textContent;
    const { badges, errorBadge } = makeListStateBadges(item, "contentErrorBadge");
    if ((item.media || []).length) {
      const mediaBadge = document.createElement("span");
      mediaBadge.className = "board-badge";
      mediaBadge.textContent = `미디어 ${item.media.length}개`;
      badges.append(mediaBadge);
    }
    badges.append(errorBadge);
    select.append(order, title, subtitle, badges);

    const actions = document.createElement("span");
    actions.className = "content-item-actions manual-order-actions";
    const up = makeButton("↑", "icon-button manual-order-step", "content-up", `${titleText} 위로`);
    const down = makeButton("↓", "icon-button manual-order-step", "content-down", `${titleText} 아래로`);
    const remove = makeButton("삭제", "text-button delete-button", "content-delete", `${singular} 삭제`);
    up.disabled = state.saving || state.discarding || index === 0;
    down.disabled = state.saving || state.discarding || index === items.length - 1;
    remove.setAttribute("aria-label", `${titleText} 삭제`);
    actions.append(up, down, remove);
    row.append(handle, select, actions);
    list.append(row);
  });
  document.querySelector(`#${collection === "projects" ? "project" : "software"}-item-count`).textContent = String(items.length);
  if (listColumn) listColumn.scrollTop = listScrollTop;
}

function makeContentField(collection, item, labelText, field, options = {}) {
  const label = document.createElement("label");
  label.className = `board-field${options.full ? " board-field-full" : ""}`;
  const title = document.createElement("span");
  title.textContent = labelText;
  let control;
  if (options.options) {
    control = document.createElement("select");
    options.options.forEach(({ value, label: optionLabel, disabled = false }) => {
      const option = document.createElement("option");
      option.value = value;
      option.textContent = optionLabel;
      option.disabled = disabled;
      control.append(option);
    });
  } else {
    control = options.textarea ? document.createElement("textarea") : document.createElement("input");
    if (!options.textarea) control.type = options.type || "text";
    control.maxLength = options.maxLength || (options.textarea ? 20000 : 300);
    control.autocomplete = "off";
  }
  control.className = options.textarea ? "board-textarea" : "board-input";
  control.value = item[field] || "";
  control.disabled = state.saving || state.discarding;
  control.dataset.contentCollection = collection;
  control.dataset.contentKey = item._clientKey;
  control.dataset.contentField = field;
  control.dataset.validationKey = contentValidationKey(collection, item._clientKey, field);
  const error = document.createElement("span");
  error.className = "board-field-error";
  error.dataset.errorScope = contentValidationKey(collection, item._clientKey, field);
  error.id = `content-error-${collection}-${item._clientKey.replace(/[^a-z0-9-]/gi, "-")}-${field}`;
  control.setAttribute("aria-describedby", error.id);
  label.append(title, control, error);
  return label;
}

function contentSectionHeading(titleText, count, action, buttonLabel, maximum = 0) {
  const heading = document.createElement("div");
  heading.className = "content-repeat-heading";
  const title = document.createElement("h4");
  title.textContent = titleText;
  const actions = document.createElement("span");
  actions.className = "content-heading-actions";
  const countChip = document.createElement("span");
  countChip.className = "count-chip";
  countChip.textContent = String(count);
  actions.append(countChip);
  if (action) {
    const add = makeButton(buttonLabel, "button button-secondary content-add-button", action);
    if (maximum && count >= maximum) add.disabled = true;
    actions.append(add);
  }
  heading.append(title, actions);
  return heading;
}

function repeatRowActions(actionPrefix, index, total, deleteLabel, options = {}) {
  const actions = document.createElement("span");
  actions.className = `content-row-actions${options.compact ? " content-row-actions-compact" : ""}`;
  const up = makeButton("↑", "icon-button", `${actionPrefix}-up`, "위로");
  const down = makeButton("↓", "icon-button", `${actionPrefix}-down`, "아래로");
  const remove = makeButton("삭제", "text-button delete-button", `${actionPrefix}-delete`, deleteLabel);
  up.disabled = state.saving || state.discarding || index === 0;
  down.disabled = state.saving || state.discarding || index === total - 1;
  actions.append(up, down, remove);
  return actions;
}

function makeRepeatField(collection, item, row, field, labelText, validationField, options = {}) {
  const label = document.createElement("label");
  label.className = "board-field";
  const title = document.createElement("span");
  title.textContent = labelText;
  if (options.languageLabel) title.classList.add("language-marker");
  const input = options.textarea ? document.createElement("textarea") : document.createElement("input");
  input.className = options.textarea ? "board-textarea content-note-input" : "board-input";
  if (!options.textarea) input.type = options.type || "text";
  input.value = row[field] || "";
  input.maxLength = options.maxLength || (options.textarea ? 20000 : 300);
  input.disabled = state.saving || state.discarding;
  input.autocomplete = "off";
  input.dataset.contentCollection = collection;
  input.dataset.contentKey = item._clientKey;
  input.dataset.contentRowKey = row._clientKey;
  input.dataset[options.datasetField || "repeatField"] = field;
  if (options.ariaLabel) input.setAttribute("aria-label", options.ariaLabel);
  const key = contentValidationKey(collection, item._clientKey, validationField);
  input.dataset.validationKey = key;
  const error = document.createElement("span");
  error.className = "board-field-error";
  error.dataset.errorScope = key;
  error.id = `content-repeat-error-${collection}-${item._clientKey}-${row._clientKey}-${field}`
    .replace(/[^a-z0-9-_]/gi, "-");
  input.setAttribute("aria-describedby", error.id);
  label.append(title, input, error);
  return label;
}

function technologyCatalogKey(value) {
  const normalized = String(value || "").trim().toLocaleLowerCase();
  return SOFTWARE_TECHNOLOGY_ALIASES[normalized] || normalized;
}

function technologyCatalogEntry(value) {
  return SOFTWARE_TECHNOLOGY_BY_KEY.get(technologyCatalogKey(value)) || null;
}

function selectedTechnologyKeys(item, excludedRowKey = "") {
  return new Set((item.technologies || [])
    .filter((technology) => technology._clientKey !== excludedRowKey)
    .map((technology) => technologyCatalogEntry(technology.value)?.key)
    .filter(Boolean));
}

function canAddTechnology(item) {
  if ((item.technologies || []).some((technology) => !String(technology.value || "").trim())) return false;
  const selected = selectedTechnologyKeys(item);
  return SOFTWARE_TECHNOLOGIES.some((technology) => !selected.has(technology.key));
}

function makeTechnologyField(item, row, index) {
  const rawValue = String(row.value || "");
  const currentTechnology = technologyCatalogEntry(rawValue);
  const selectedByOtherRows = selectedTechnologyKeys(item, row._clientKey);
  const label = document.createElement("label");
  label.className = "board-field content-technology-field";

  const title = document.createElement("span");
  title.textContent = "기술";

  const control = document.createElement("span");
  control.className = `content-technology-control${rawValue.trim() ? "" : " is-empty"}`;
  const icon = document.createElement("span");
  icon.className = `content-technology-icon${currentTechnology ? "" : " is-unsupported"}`;
  icon.setAttribute("aria-hidden", "true");
  if (currentTechnology) {
    icon.dataset.technologyKey = currentTechnology.key;
    const image = document.createElement("img");
    image.src = `assets/technologies/${currentTechnology.icon}`;
    image.alt = "";
    icon.append(image);
  } else {
    icon.textContent = "?";
  }

  const select = document.createElement("select");
  select.className = "board-input content-technology-select";
  select.disabled = state.saving || state.discarding;
  select.dataset.contentCollection = "software";
  select.dataset.contentKey = item._clientKey;
  select.dataset.contentRowKey = row._clientKey;
  select.dataset.technologyField = "value";
  select.setAttribute("aria-label", `${rawValue.trim() || `기술 ${index + 1}`} 선택`);

  const placeholder = document.createElement("option");
  placeholder.value = "";
  placeholder.textContent = "선택";
  placeholder.disabled = true;
  select.append(placeholder);

  if (rawValue.trim() && !currentTechnology) {
    const legacyGroup = document.createElement("optgroup");
    legacyGroup.label = "JSON에서 직접 입력된 값";
    const legacyOption = document.createElement("option");
    legacyOption.value = rawValue;
    legacyOption.textContent = `${rawValue.trim()} (아이콘 미지원)`;
    legacyGroup.append(legacyOption);
    select.append(legacyGroup);
  }

  SOFTWARE_TECHNOLOGY_GROUPS.forEach((group) => {
    const optionGroup = document.createElement("optgroup");
    optionGroup.label = group.label;
    group.items.forEach((technology) => {
      const option = document.createElement("option");
      const isCurrent = currentTechnology?.key === technology.key;
      option.value = isCurrent ? rawValue : technology.value;
      option.textContent = technology.label;
      option.disabled = !isCurrent && selectedByOtherRows.has(technology.key);
      optionGroup.append(option);
    });
    select.append(optionGroup);
  });
  select.value = rawValue;

  const key = contentValidationKey("software", item._clientKey, `technology:${row._clientKey}`);
  select.dataset.validationKey = key;
  const error = document.createElement("span");
  error.className = "sr-only";
  error.dataset.errorScope = key;
  error.id = `content-repeat-error-software-${item._clientKey}-${row._clientKey}-value`
    .replace(/[^a-z0-9-_]/gi, "-");
  select.setAttribute("aria-describedby", error.id);
  if (rawValue.trim()) control.append(icon);
  control.append(select);
  label.append(title, control, error);
  return label;
}

function makeEditorFormSection(titleText, ...children) {
  const section = document.createElement("section");
  section.className = "editor-form-section";
  const heading = document.createElement("h4");
  heading.textContent = titleText;
  section.append(heading, ...children);
  return section;
}

function noteSummary(note, index) {
  const text = String(note.value || "").replace(/\s+/g, " ").trim();
  if (!text) return `설명 ${index + 1} · 내용 입력 필요`;
  return `설명 ${index + 1} · ${text.length > 80 ? `${text.slice(0, 80)}…` : text}`;
}

function renderNotesColumn(collection, item, language, label, options) {
  const notes = item[`notes_${language}_rows`] || [];
  const kind = `note-${language}`;
  const section = document.createElement("section");
  section.className = "bilingual-repeat-column";
  section.dataset.contentNoteLanguage = language;
  const heading = contentSectionHeading(label, notes.length, `content-${kind}-add`, `${label} 설명 추가`, 100);
  heading.querySelector("h4").classList.add("language-marker");
  section.append(heading);
  const list = document.createElement("div");
  list.className = `content-repeat-list${options.collapsible ? " content-collapsible-list" : ""}`;
  notes.forEach((note, index) => {
    const row = document.createElement("div");
    row.className = "entity-string-row";
    row.dataset.contentRowKey = note._clientKey;
    const field = makeRepeatField(collection, item, note, "value", `${label} 설명`, `note:${note._clientKey}:${language}`, {
      textarea: true, datasetField: "noteField", maxLength: 5000, ariaLabel: `${label} 설명`,
    });
    field.querySelector("span").classList.add("sr-only");
    field.querySelector("textarea").dataset.contentNoteLanguage = language;
    row.append(
      field,
      repeatRowActions(`content-${kind}`, index, notes.length, `${label} 설명 삭제`),
    );
    if (!options.collapsible) {
      list.append(row);
      return;
    }
    const details = document.createElement("details");
    details.className = "content-collapsible-row";
    details.dataset.contentRowKey = note._clientKey;
    details.open = !String(note.value || "").trim();
    const summary = document.createElement("summary");
    summary.textContent = noteSummary(note, index);
    row.removeAttribute("data-content-row-key");
    details.append(summary, row);
    list.append(details);
  });
  const error = document.createElement("span");
  error.className = "board-field-error";
  error.dataset.errorScope = contentValidationKey(collection, item._clientKey, `notes_${language}`);
  section.append(list, error);
  return section;
}

function renderNotesSection(collection, item, options = {}) {
  const section = document.createElement("section");
  section.className = "content-repeat-section bilingual-repeat-section";
  const heading = document.createElement("div");
  heading.className = "content-repeat-heading";
  const title = document.createElement("h4");
  title.textContent = "설명";
  heading.append(title);
  const columns = document.createElement("div");
  columns.className = "bilingual-repeat-grid";
  columns.append(
    renderNotesColumn(collection, item, "en", "영문", options),
    renderNotesColumn(collection, item, "kr", "국문", options),
  );
  section.append(heading, columns);
  return section;
}

function renderTechnologiesSection(item) {
  const collection = "software";
  const section = document.createElement("section");
  section.className = "content-repeat-section software-technologies-section";
  const heading = contentSectionHeading("기술", item.technologies.length, "content-technology-add", "기술 추가", 100);
  const addButton = heading.querySelector('[data-action="content-technology-add"]');
  if (addButton) addButton.disabled = state.saving || state.discarding || !canAddTechnology(item);
  section.append(heading);
  const list = document.createElement("div");
  list.className = "content-technology-list";
  list.dataset.technologySortList = "";
  list.dataset.contentCollection = collection;
  list.dataset.contentKey = item._clientKey;
  item.technologies.forEach((technology, index) => {
    const row = document.createElement("div");
    row.className = "content-technology-item";
    row.dataset.contentRowKey = technology._clientKey;
    const chip = document.createElement("div");
    chip.className = "content-technology-chip";
    const technologyPreset = technologyCatalogEntry(technology.value);
    const technologyLabel = technologyPreset?.label || String(technology.value || "").trim() || `기술 ${index + 1}`;
    if (String(technology.value || "").trim() && !technologyPreset) chip.classList.add("is-unsupported");
    const handle = makeButton(
      "",
      "drag-handle content-technology-drag-handle",
      "drag-technology",
      `${technologyLabel} 순서 끌어 이동. Alt+방향키로도 이동할 수 있습니다.`,
    );
    handle.draggable = !handle.disabled;
    handle.setAttribute("aria-keyshortcuts", "Alt+ArrowUp Alt+ArrowDown Alt+ArrowLeft Alt+ArrowRight");
    const field = makeTechnologyField(item, technology, index);
    const actions = repeatRowActions("content-technology", index, item.technologies.length, "기술 삭제", { compact: true });
    actions.classList.add("content-technology-actions");
    chip.append(handle, field, actions);
    row.append(chip);
    list.append(row);
  });
  if (!item.technologies.length) {
    const error = document.createElement("span");
    error.className = "board-field-error content-technology-list-error";
    error.dataset.errorScope = contentValidationKey(collection, item._clientKey, "technologies");
    list.append(error);
  }
  section.append(list);
  return section;
}

function renderLinksSection(item) {
  const collection = "software";
  const section = document.createElement("section");
  section.className = "content-repeat-section software-links-section";
  section.append(contentSectionHeading("링크", item.links.length, "content-link-add", "링크 추가", 100));
  const list = document.createElement("div");
  list.className = "content-repeat-list";
  item.links.forEach((link, index) => {
    const row = document.createElement("div");
    row.className = "content-link-row";
    row.dataset.contentRowKey = link._clientKey;
    const fields = document.createElement("div");
    fields.className = "content-link-fields";
    const primary = document.createElement("div");
    primary.className = "content-link-primary";
    primary.append(
      makeRepeatField(collection, item, link, "url", "URL", `link:${link._clientKey}:url`, { datasetField: "linkField", maxLength: 4096, type: "url" }),
      makeRepeatField(collection, item, link, "label", "대표 이름", `link:${link._clientKey}:label`, { datasetField: "linkField" }),
    );
    const languageDetails = document.createElement("details");
    languageDetails.className = "content-link-language-details";
    const languageSummary = document.createElement("summary");
    languageSummary.textContent = "언어별 이름";
    const languageFields = document.createElement("div");
    languageFields.className = "content-link-language-fields";
    languageFields.append(
      makeRepeatField(collection, item, link, "label_en", "영문", `link:${link._clientKey}:label_en`, { datasetField: "linkField", languageLabel: true, ariaLabel: "링크 영문 이름" }),
      makeRepeatField(collection, item, link, "label_ko", "국문", `link:${link._clientKey}:label_ko`, { datasetField: "linkField", languageLabel: true, ariaLabel: "링크 국문 이름" }),
    );
    languageDetails.append(languageSummary, languageFields);
    fields.append(primary, languageDetails);
    const actions = repeatRowActions("content-link", index, item.links.length, "링크 삭제", { compact: true });
    actions.classList.add("content-link-actions");
    row.append(fields, actions);
    list.append(row);
  });
  section.append(list);
  return section;
}

function dropMessageKey(collection, clientKey) {
  return `${collection}:${clientKey}`;
}

function renderContentMediaSection(collection, item) {
  const section = document.createElement("section");
  section.className = "board-media-section";
  section.append(contentSectionHeading("미디어", item.media.length));
  const dropzone = document.createElement("div");
  dropzone.className = `board-dropzone${state.boardDropBusy ? " is-receiving" : ""}`;
  dropzone.dataset.contentMediaDropzone = "";
  dropzone.dataset.contentCollection = collection;
  dropzone.dataset.contentKey = item._clientKey;
  dropzone.setAttribute("aria-label", "Explorer에서 사진을 드롭하는 영역");
  dropzone.setAttribute("aria-busy", String(state.boardDropBusy));
  const dropTitle = document.createElement("strong");
  dropTitle.textContent = state.boardDropBusy ? "사진을 준비하는 중…" : "Explorer에서 사진을 이곳으로 드래그";
  dropzone.append(dropTitle);
  const message = state.contentDropMessages[dropMessageKey(collection, item._clientKey)] || "";
  if (message) {
    const feedback = document.createElement("p");
    feedback.className = "board-drop-feedback";
    feedback.setAttribute("role", "status");
    feedback.textContent = message;
    dropzone.append(feedback);
  }
  section.append(dropzone);

  if (!item.media.length) return section;
  const list = document.createElement("ol");
  list.className = "board-media-list";
  item.media.forEach((media, index) => {
    const row = document.createElement("li");
    row.className = "board-media-row content-media-row";
    row.dataset.contentMediaKey = media._clientKey;
    const preview = makeMediaPreviewButton(media, boardMediaName(media, index));
    const fields = document.createElement("div");
    fields.className = "board-media-fields";
    const name = document.createElement("span");
    name.className = "board-media-name";
    const sizeLabel = formatBytes(media.size);
    name.textContent = `${boardMediaName(media, index)}${sizeLabel ? ` · ${sizeLabel}` : ""}`;
    const english = document.createElement("input");
    const korean = document.createElement("input");
    [[english, "caption_en", "영문 사진 설명"], [korean, "caption_ko", "국문 사진 설명"]]
      .forEach(([input, field, ariaLabel]) => {
        input.type = "text";
        input.className = "board-caption-input";
        input.maxLength = 500;
        input.value = media[field] || "";
        input.disabled = state.saving || state.discarding;
        input.dataset.contentCollection = collection;
        input.dataset.contentKey = item._clientKey;
        input.dataset.contentMediaKey = media._clientKey;
        input.dataset.contentMediaField = field;
        input.dataset.validationKey = contentValidationKey(collection, item._clientKey, `media:${media._clientKey}:${field}`);
        input.setAttribute("aria-label", `${boardMediaName(media, index)} ${ariaLabel}`);
      });
    const error = document.createElement("span");
    error.className = "board-caption-error";
    error.dataset.errorScope = contentValidationKey(collection, item._clientKey, `media:${media._clientKey}:`);
    error.id = `content-media-error-${collection}-${item._clientKey}-${media._clientKey}`
      .replace(/[^a-z0-9-_]/gi, "-");
    english.setAttribute("aria-describedby", error.id);
    korean.setAttribute("aria-describedby", error.id);
    fields.append(name, makeBilingualCaptionGroup("사진 설명", english, korean, error));
    const actions = repeatRowActions("content-media", index, item.media.length, "미디어 제거");
    row.append(preview, fields, actions);
    list.append(row);
  });
  section.append(list);
  return section;
}

function renderProjectEditor() {
  const placeholder = document.querySelector("#project-editor-placeholder");
  const form = document.querySelector("#project-form");
  const found = contentItem("projects", state.selectedProjectKey);
  form.replaceChildren();
  if (!found.item) {
    placeholder.hidden = false;
    form.hidden = true;
    return;
  }
  placeholder.hidden = true;
  form.hidden = false;
  form.dataset.contentCollection = "projects";
  form.dataset.contentKey = found.item._clientKey;
  const heading = document.createElement("div");
  heading.className = "board-form-heading";
  const title = document.createElement("h3");
  title.dataset.editorHeading = "projects";
  title.textContent = found.item.title_ko || found.item.title_en || "(제목 미입력)";
  heading.append(title);
  form.append(heading);
  const themeOptions = [{ value: "", label: "주제 선택" }].concat(sharedTopicOptions());
  const basics = document.createElement("div");
  basics.className = "project-basics-grid";
  basics.append(
    makeContentField("projects", found.item, "시작일", "start_date", { type: "date" }),
    makeContentField("projects", found.item, "종료일", "end_date", { type: "date" }),
    makeContentField("projects", found.item, "주제", "theme", { options: themeOptions }),
  );
  const grid = document.createElement("div");
  grid.className = "board-form-grid project-language-grid";
  grid.append(
    makeBilingualFieldGroup(
      "제목",
      makeContentField("projects", found.item, "영문", "title_en", { maxLength: 500 }),
      makeContentField("projects", found.item, "국문", "title_ko", { maxLength: 500 }),
    ),
    makeBilingualFieldGroup(
      "지원기관",
      makeContentField("projects", found.item, "영문", "funder_en", { maxLength: 500 }),
      makeContentField("projects", found.item, "국문", "funder_ko", { maxLength: 500 }),
    ),
  );
  form.append(
    basics,
    grid,
    renderNotesSection("projects", found.item, { collapsible: true }),
    renderContentMediaSection("projects", found.item),
  );
}

function renderSoftwareEditor() {
  const placeholder = document.querySelector("#software-editor-placeholder");
  const form = document.querySelector("#software-form");
  const found = contentItem("software", state.selectedSoftwareKey);
  form.replaceChildren();
  if (!found.item) {
    placeholder.hidden = false;
    form.hidden = true;
    return;
  }
  placeholder.hidden = true;
  form.hidden = false;
  form.dataset.contentCollection = "software";
  form.dataset.contentKey = found.item._clientKey;
  const heading = document.createElement("div");
  heading.className = "board-form-heading";
  const title = document.createElement("h3");
  title.textContent = found.item._isNew ? "새 소프트웨어" : found.item.name || "(이름 미입력)";
  heading.append(title);
  form.append(heading);
  const grid = document.createElement("div");
  grid.className = "board-form-grid software-basic-grid";
  const nameField = makeContentField("software", found.item, "이름", "name");
  nameField.classList.add("software-name-field");
  const stageField = makeContentField("software", found.item, "개발 단계", "stage", {
    options: [
      { value: "release", label: "출시" },
      { value: "preview", label: "미리보기" },
      { value: "development", label: "개발 중" },
    ],
  });
  stageField.classList.add("software-stage-field");
  grid.append(nameField, stageField);
  const descriptionSection = renderNotesSection("software", found.item);
  descriptionSection.classList.add("software-description-section");
  const mediaSection = renderContentMediaSection("software", found.item);
  mediaSection.classList.add("software-media-section");
  form.append(
    grid,
    descriptionSection,
    renderTechnologiesSection(found.item),
    renderLinksSection(found.item),
    mediaSection,
  );
}

function entityDraft(collection) {
  return state[entityConfig[collection].draftKey] || [];
}

function selectedEntityKey(collection) {
  return state[entityConfig[collection].selectedKey] || "";
}

function setSelectedEntityKey(collection, key) {
  state[entityConfig[collection].selectedKey] = key;
}

function entityItem(collection, clientKey) {
  const items = entityDraft(collection);
  const index = items.findIndex((item) => item._clientKey === clientKey);
  return { items, index, item: index >= 0 ? items[index] : null };
}

function normalizedListSearch(value) {
  return String(value || "").normalize("NFKC").toLocaleLowerCase().trim();
}

function matchesListSearch(query, ...values) {
  const normalizedQuery = normalizedListSearch(query);
  if (!normalizedQuery) return true;
  return normalizedListSearch(values.flat(Infinity).join(" ")).includes(normalizedQuery);
}

function resetEditorScroll(panelName) {
  const editor = document.querySelector(`#panel-${panelName} .master-detail-editor`);
  if (editor) editor.scrollTop = 0;
}

function personNameLabel(person) {
  const names = [person?.name_ko, person?.name_en].map((value) => String(value || "").trim()).filter(Boolean);
  return [...new Set(names)].join(" · ") || "이름 미입력";
}

function personAffiliationLabel(person) {
  const latestKorean = [...(person?.notes_ko_rows || [])]
    .reverse().find((row) => String(row.value || "").trim())?.value?.trim();
  const latestEnglish = [...(person?.notes_en_rows || [])]
    .reverse().find((row) => String(row.value || "").trim())?.value?.trim();
  return [...new Set([latestKorean, latestEnglish].filter(Boolean))].join(" / ");
}

function personRelationshipLabel(person) {
  const affiliation = personAffiliationLabel(person);
  return `${personNameLabel(person)}${affiliation ? ` — ${affiliation}` : ""}`;
}

const HANGUL_NAME_PATTERN = /[\uac00-\ud7a3]/;
const KOREAN_NAME_COLLATOR = new Intl.Collator("ko-KR", { sensitivity: "base", numeric: true });
const ENGLISH_NAME_COLLATOR = new Intl.Collator("en-US", { sensitivity: "base", numeric: true });

function comparePeopleForDisplay(left, right) {
  const leftKoreanName = String(left?.name_ko || "").normalize("NFKC").trim();
  const rightKoreanName = String(right?.name_ko || "").normalize("NFKC").trim();
  const leftHasKoreanName = HANGUL_NAME_PATTERN.test(leftKoreanName);
  const rightHasKoreanName = HANGUL_NAME_PATTERN.test(rightKoreanName);
  if (leftHasKoreanName !== rightHasKoreanName) return leftHasKoreanName ? -1 : 1;

  if (leftHasKoreanName) {
    const koreanOrder = KOREAN_NAME_COLLATOR.compare(leftKoreanName, rightKoreanName);
    if (koreanOrder) return koreanOrder;
  }

  const leftEnglishName = String(left?.name_en || leftKoreanName).normalize("NFKC").trim();
  const rightEnglishName = String(right?.name_en || rightKoreanName).normalize("NFKC").trim();
  const englishOrder = ENGLISH_NAME_COLLATOR.compare(leftEnglishName, rightEnglishName);
  if (englishOrder) return englishOrder;
  return KOREAN_NAME_COLLATOR.compare(leftKoreanName, rightKoreanName);
}

function sortedPeopleForDisplay(items) {
  return (items || [])
    .map((item, index) => ({ item, index }))
    .sort((left, right) => comparePeopleForDisplay(left.item, right.item) || left.index - right.index)
    .map(({ item }) => item);
}

function personMatchesListSearch(person) {
  return matchesListSearch(
    state.listFilters.peopleSearch,
    person?.name_ko,
    person?.name_en,
    personAffiliationLabel(person),
  );
}

function publicationTypeLabel(type) {
  return ({
    "international-journal": "국제학술지",
    "domestic-journal": "국내학술지",
    "international-conference": "국제학술대회",
    "domestic-conference": "국내학술대회",
  })[type] || "유형 미선택";
}

function publicationStatusLabel(item) {
  if (item._status === "under_review") return "심사 중";
  if (item._status === "in_press") return "게재 예정";
  return "출판됨";
}

function publicationYearLabel(item) {
  const match = String(publicationMetadata(item).date).match(/^\d{4}/);
  return match ? match[0] : "연도 미입력";
}

function publicationTypeShortLabel(type) {
  return ({
    "international-journal": "국제 저널",
    "domestic-journal": "국내 저널",
    "international-conference": "국제 학회",
    "domestic-conference": "국내 학회",
  })[type] || "유형 미선택";
}

function topicLabel(topicID) {
  const topic = (state.draft?.topics || []).find((item) => item.id === topicID);
  if (!topic) return topicID ? "연결된 주제를 찾을 수 없음" : "주제 없음";
  return [...new Set([topic.label_ko, topic.label_en].filter(Boolean))].join(" · ");
}

function personUsage(clientKey) {
  const counts = personUsageCounts(clientKey);
  return counts.publications + counts.patents;
}

function personUsageCounts(clientKey) {
  return {
    publications: (state.publicationsDraft || []).filter((publication) => publication.author_keys?.includes(clientKey)).length,
    patents: (state.profileDraft?.patents || []).filter((patent) => patent.inventor_keys?.includes(clientKey)).length,
  };
}

function personUsageLabel(clientKey) {
  const counts = personUsageCounts(clientKey);
  return [counts.publications ? `논문 ${counts.publications}개` : "", counts.patents ? `특허 ${counts.patents}개` : ""].filter(Boolean).join(" · ");
}

function defaultPersonSelection(items) {
  return ((items || []).find((item) => item.is_self) || (items || [])[0])?._clientKey || "";
}

function awardUsage(clientKey) {
  return (state.publicationsDraft || []).filter((publication) => publication.award_key === clientKey).length;
}

function makeAwardPublicationLink(publications) {
  const link = document.createElement("span");
  link.className = "award-publication-link";
  link.title = publications.map((publication) => publication.title_ko || publication.title_en || "(제목 미입력)").join("\n");
  link.setAttribute("role", "img");
  link.setAttribute("aria-label", `연결된 논문: ${link.title}`);
  link.tabIndex = 0;
  const namespace = "http://www.w3.org/2000/svg";
  const icon = document.createElementNS(namespace, "svg");
  icon.classList.add("action-icon-svg");
  icon.setAttribute("viewBox", "0 0 24 24");
  icon.setAttribute("aria-hidden", "true");
  icon.setAttribute("focusable", "false");
  [
    "M10 13a5 5 0 0 0 7.07 0l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71",
    "M14 11a5 5 0 0 0-7.07 0l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71",
  ].forEach((data) => {
    const path = document.createElementNS(namespace, "path");
    path.setAttribute("d", data);
    icon.append(path);
  });
  link.append(icon);
  return link;
}

function publicationMatchesListFilters(item) {
  const status = state.listFilters.publicationsStatus;
  if (status !== "all" && item._status !== status) return false;
  const authors = (item.author_keys || []).map((key) => {
    const person = (state.peopleDraft || []).find((candidate) => candidate._clientKey === key);
    return person ? personRelationshipLabel(person) : item._unresolvedAuthorIDs?.[key] || "";
  });
  return matchesListSearch(
    state.listFilters.publicationsSearch,
    item.title_ko,
    item.title_en,
    authors,
    publicationMetadata(item).venue_kr,
    publicationMetadata(item).venue_en,
  );
}

function renderEntityListRow(collection, item, selectedKey, positionIndex, positionLength) {
  const config = entityConfig[collection];
  const isOwner = collection === "people" && item.is_self;
  const manualOrderSearchLocked = collection === "people"
    && !isOwner
    && Boolean(state.listFilters.peopleSearch.trim());
  const manualOrderSearchMessage = "검색을 해제한 후 순서를 변경할 수 있습니다.";
  const titleText = collection === "people"
    ? personNameLabel(item)
    : collection === "awards"
      ? item.title_ko || item.title_en || "(제목 미입력)"
      : item.title_ko || item.title_en || "(제목 미입력)";
  const row = document.createElement("li");
  row.className = "board-item-row entity-item-row";
  row.dataset.entityCollection = collection;
  row.dataset.entityKey = item._clientKey;
  if (collection === "people") row.dataset.personGroup = isOwner ? "owner" : "coauthor";

  if (!isOwner && config.manualOrder) {
    row.classList.add("manual-order-row");
    const handle = makeButton(
      "",
      "drag-handle list-drag-handle",
      "drag-entity-item",
      manualOrderSearchLocked ? manualOrderSearchMessage : `${titleText} 순서 끌어 이동`,
    );
    handle.disabled ||= manualOrderSearchLocked;
    handle.draggable = !handle.disabled;
    row.append(handle);
  }

  const select = document.createElement("button");
  select.type = "button";
  select.className = "board-item-select";
  select.dataset.action = "entity-select";
  select.disabled = state.saving || state.discarding;
  const selectionContext = collection === "people" ? `${isOwner ? "본인" : "공저자"} ` : "";
  select.setAttribute("aria-label", `${titleText} ${selectionContext}선택`);
  applyListRowSelection(row, select, item._clientKey === selectedKey);

  const order = document.createElement("span");
  order.className = "board-item-date";
  if (collection === "people") order.textContent = `${positionIndex + 1}.`;
  if (collection === "awards") order.textContent = item.date || "날짜 미입력";
  if (collection === "publications") {
    order.textContent = [
      publicationStatusLabel(item),
      publicationYearLabel(item),
      publicationTypeShortLabel(item.publication_type),
    ].join(" · ");
  }
  const title = document.createElement("span");
  title.className = "board-item-title";
  title.textContent = titleText;
  title.title = title.textContent;
  const subtitle = document.createElement("span");
  subtitle.className = "board-item-subtitle";
  if (collection === "people") subtitle.textContent = personAffiliationLabel(item) || "소속/메모 없음";
  if (collection === "awards") {
    subtitle.textContent = item.organization_ko || item.organization_en || "기관 미입력";
  }
  if (collection === "publications") {
    subtitle.textContent = publicationMetadata(item).venue_kr || publicationMetadata(item).venue_en || "게재지 미입력";
  }
  const { badges, errorBadge } = makeListStateBadges(item, "entityErrorBadge");
  const linkedPublications = collection === "awards"
    ? (state.publicationsDraft || []).filter((publication) => publication.award_key === item._clientKey)
    : [];
  const usage = collection === "people" ? personUsage(item._clientKey) : linkedPublications.length;
  if (collection === "people" && usage) {
    const badge = document.createElement("span");
    badge.className = "board-badge entity-usage-badge";
    badge.textContent = personUsageLabel(item._clientKey);
    badges.append(badge);
  }
  if (collection === "publications" && (item.author_keys || []).length) {
    const badge = document.createElement("span");
    badge.className = "board-badge";
    badge.textContent = `저자 ${(item.author_keys || []).length}명`;
    badges.append(badge);
  }
  badges.append(errorBadge);
  if (!isOwner) select.append(order);
  select.append(title, subtitle, badges);

  row.append(select);
  if (!isOwner) {
    const actions = document.createElement("span");
    actions.className = "content-item-actions entity-item-actions";
    if (config.manualOrder) {
      actions.classList.add("manual-order-actions");
      const up = makeButton("↑", "icon-button manual-order-step", "entity-up", `${titleText} 위로`);
      const down = makeButton("↓", "icon-button manual-order-step", "entity-down", `${titleText} 아래로`);
      up.disabled = state.saving || state.discarding || manualOrderSearchLocked || positionIndex === 0;
      down.disabled = state.saving || state.discarding || manualOrderSearchLocked || positionIndex === positionLength - 1;
      if (manualOrderSearchLocked) {
        [up, down].forEach((button) => {
          button.title = manualOrderSearchMessage;
          button.setAttribute("aria-label", manualOrderSearchMessage);
        });
      }
      actions.append(up, down);
    }
    const remove = makeButton("삭제", "text-button delete-button", "entity-delete", `${config.singular} 삭제`);
    remove.disabled = state.saving || state.discarding || usage > 0;
    if (usage) remove.title = collection === "awards" ? "연결된 논문에서 사용 중입니다." : `${personUsageLabel(item._clientKey)}에서 사용 중입니다.`;
    if (linkedPublications.length) {
      actions.classList.add("award-linked-actions");
      actions.append(makeAwardPublicationLink(linkedPublications));
    }
    actions.append(remove);
    row.append(actions);
  }
  return row;
}

function renderEntityList(collection) {
  const config = entityConfig[collection];
  const rawItems = entityDraft(collection);
  const listColumn = document.querySelector(`#panel-${collection} .master-detail-list`);
  const listScrollTop = listColumn?.scrollTop || 0;
  let selectedKey = selectedEntityKey(collection);
  if (selectedKey && !rawItems.some((item) => item._clientKey === selectedKey)) {
    selectedKey = "";
    setSelectedEntityKey(collection, "");
  }

  if (collection === "people") {
    const owners = sortedPeopleForDisplay(rawItems.filter((item) => item.is_self));
    const coauthors = sortedPeopleForDisplay(rawItems.filter((item) => !item.is_self));
    const visibleCoauthors = coauthors.filter(personMatchesListSearch);
    const ownerList = document.querySelector("#people-owner-list");
    const coauthorList = document.querySelector("#people-coauthor-list");
    delete coauthorList.dataset.manualOrderList;
    delete coauthorList.dataset.manualOrderCollection;
    ownerList.replaceChildren(...owners.map((item, index) => (
      renderEntityListRow(collection, item, selectedKey, index, owners.length)
    )));
    coauthorList.replaceChildren(...visibleCoauthors.map((item) => (
      renderEntityListRow(collection, item, selectedKey, coauthors.indexOf(item), coauthors.length)
    )));
    coauthorList.dataset.emptyLabel = state.listFilters.peopleSearch ? "검색 결과가 없습니다" : "공저자가 없습니다";
    document.querySelector("#people-owner-count").textContent = String(owners.length);
    document.querySelector("#people-coauthor-count").textContent = String(coauthors.length);
  } else {
    const items = config.sorter(rawItems).filter((item) => (
      collection !== "publications" || publicationMatchesListFilters(item)
    ));
    const list = document.querySelector(`#${config.domPrefix}-item-list`);
    if (collection === "publications") {
      list.dataset.emptyLabel = state.listFilters.publicationsSearch || state.listFilters.publicationsStatus !== "all"
        ? "검색 또는 필터 결과가 없습니다"
        : "논문이 없습니다";
    }
    list.replaceChildren(...items.map((item, index) => (
      renderEntityListRow(collection, item, selectedKey, index, items.length)
    )));
  }
  document.querySelector(`#${config.domPrefix}-item-count`).textContent = String(rawItems.length);
  if (listColumn) listColumn.scrollTop = listScrollTop;
}

function makeEntityField(collection, item, labelText, field, options = {}) {
  const label = document.createElement("label");
  label.className = `board-field${options.full ? " board-field-full" : ""}`;
  const title = document.createElement("span");
  title.textContent = labelText;
  let control;
  if (options.options) {
    control = document.createElement("select");
    options.options.forEach(({ value, label: optionLabel, disabled = false }) => {
      const option = document.createElement("option");
      option.value = value;
      option.textContent = optionLabel;
      option.disabled = disabled;
      control.append(option);
    });
  } else {
    control = options.textarea ? document.createElement("textarea") : document.createElement("input");
    if (!options.textarea) control.type = options.type || "text";
    if (options.maxLength !== null) {
      control.maxLength = options.maxLength || (options.textarea ? 50000 : 1000);
    }
    control.autocomplete = "off";
    if (options.placeholder) control.placeholder = options.placeholder;
  }
  control.className = options.textarea
    ? `board-textarea${options.controlClass ? ` ${options.controlClass}` : ""}`
    : "board-input";
  control.value = item[field] || "";
  control.disabled = state.saving || state.discarding || options.disabled;
  control.dataset.entityCollection = collection;
  control.dataset.entityKey = item._clientKey;
  control.dataset.entityField = field;
  control.dataset.validationKey = contentValidationKey(collection, item._clientKey, field);
  const error = document.createElement("span");
  error.className = "board-field-error";
  error.dataset.errorScope = contentValidationKey(collection, item._clientKey, field);
  error.id = `entity-error-${collection}-${item._clientKey}-${field}`.replace(/[^a-z0-9-_]/gi, "-");
  control.setAttribute("aria-describedby", error.id);
  label.append(title, control, error);
  return label;
}

function renderEntityStringSection(collection, item, field, titleText, addLabel, inputLabel, options = {}) {
  const rows = item[field] || [];
  const section = document.createElement("section");
  section.className = `${options.nested ? "bilingual-repeat-column" : "content-repeat-section"}${options.compact ? " entity-string-section-compact" : ""}`;
  const heading = contentSectionHeading(titleText, rows.length, "entity-string-add", addLabel, options.maximum || 100);
  if (options.languageHeading) heading.querySelector("h4")?.classList.add("language-marker");
  heading.querySelector("[data-action='entity-string-add']").dataset.stringField = field;
  section.append(heading);
  const list = document.createElement("div");
  list.className = `content-repeat-list${options.compact ? " entity-string-list-compact" : ""}`;
  rows.forEach((rowItem, index) => {
    const row = document.createElement("div");
    row.className = `entity-string-row${options.compact ? " entity-string-row-compact" : ""}`;
    row.dataset.entityStringKey = rowItem._clientKey;
    row.dataset.entityStringField = field;
    const label = document.createElement("label");
    label.className = "board-field";
    const caption = document.createElement("span");
    caption.textContent = inputLabel;
    const input = options.textarea ? document.createElement("textarea") : document.createElement("input");
    input.className = options.textarea ? "board-textarea content-note-input" : "board-input";
    if (!options.textarea) input.type = "text";
    input.value = rowItem.value || "";
    if (options.maxLength !== null) input.maxLength = options.maxLength || 5000;
    input.disabled = state.saving || state.discarding;
    input.autocomplete = "off";
    input.dataset.entityCollection = collection;
    input.dataset.entityKey = item._clientKey;
    input.dataset.entityStringField = field;
    input.dataset.entityStringKey = rowItem._clientKey;
    input.dataset.validationKey = contentValidationKey(collection, item._clientKey, `${field}:${rowItem._clientKey}`);
    const error = document.createElement("span");
    error.className = "board-field-error";
    error.dataset.errorScope = input.dataset.validationKey;
    if (options.hideInputLabel) {
      input.setAttribute("aria-label", inputLabel);
      label.append(input, error);
    } else {
      label.append(caption, input, error);
    }
    row.append(
      label,
      repeatRowActions("entity-string", index, rows.length, `${inputLabel} 삭제`, { compact: options.compact }),
    );
    list.append(row);
  });
  section.append(list);
  return section;
}

function renderBilingualStringSection(collection, item, titleText, fields) {
  const section = document.createElement("section");
  section.className = `content-repeat-section bilingual-repeat-section${fields.options?.compact ? " bilingual-repeat-section-compact" : ""}`;
  const heading = document.createElement("div");
  heading.className = "content-repeat-heading";
  const title = document.createElement("h4");
  title.textContent = titleText;
  heading.append(title);
  const columns = document.createElement("div");
  columns.className = `bilingual-repeat-grid${fields.options?.compact ? " bilingual-repeat-grid-compact" : ""}`;
  columns.append(
    renderEntityStringSection(
      collection,
      item,
      fields.english.field,
      "영문",
      `${fields.english.ariaLabel} 추가`,
      fields.english.ariaLabel,
      { ...fields.options, nested: true, hideInputLabel: true, languageHeading: true },
    ),
    renderEntityStringSection(
      collection,
      item,
      fields.korean.field,
      "국문",
      `${fields.korean.ariaLabel} 추가`,
      fields.korean.ariaLabel,
      { ...fields.options, nested: true, hideInputLabel: true, languageHeading: true },
    ),
  );
  section.append(heading, columns);
  return section;
}

function renderAffiliationSection(item) {
  return renderBilingualStringSection("people", item, "소속 / 메모", {
    english: { field: "notes_en_rows", ariaLabel: "영문 소속 / 메모" },
    korean: { field: "notes_ko_rows", ariaLabel: "국문 소속 / 메모" },
    options: { textarea: true },
  });
}

function renderPersonEditor() {
  const placeholder = document.querySelector("#people-editor-placeholder");
  const form = document.querySelector("#people-form");
  const found = entityItem("people", state.selectedPersonKey);
  form.replaceChildren();
  if (!found.item) {
    placeholder.hidden = false;
    form.hidden = true;
    return;
  }
  placeholder.hidden = true;
  form.hidden = false;
  form.dataset.entityCollection = "people";
  form.dataset.entityKey = found.item._clientKey;
  const heading = document.createElement("div");
  heading.className = "board-form-heading";
  const title = document.createElement("h3");
  title.textContent = found.item._isNew ? "새 사람" : personNameLabel(found.item);
  heading.append(title);
  const grid = document.createElement("div");
  grid.className = "board-form-grid";
  grid.append(
    makeBilingualFieldGroup(
      "이름",
      makeEntityField("people", found.item, "영문", "name_en"),
      makeEntityField("people", found.item, "국문", "name_ko"),
    ),
  );
  form.append(
    heading,
    grid,
    renderAffiliationSection(found.item),
  );
}

function renderAwardEditor() {
  const placeholder = document.querySelector("#awards-editor-placeholder");
  const form = document.querySelector("#awards-form");
  const found = entityItem("awards", state.selectedAwardKey);
  form.replaceChildren();
  if (!found.item) {
    placeholder.hidden = false;
    form.hidden = true;
    return;
  }
  placeholder.hidden = true;
  form.hidden = false;
  form.dataset.entityCollection = "awards";
  form.dataset.entityKey = found.item._clientKey;
  const heading = document.createElement("div");
  heading.className = "board-form-heading";
  const title = document.createElement("h3");
  title.textContent = found.item._isNew ? "새 수상" : found.item.title_ko || found.item.title_en || "(제목 미입력)";
  heading.append(title);
  const grid = document.createElement("div");
  grid.className = "board-form-grid";
  grid.append(
    makeEntityField("awards", found.item, "날짜", "date", { type: "date", full: true }),
    makeBilingualFieldGroup(
      "제목",
      makeEntityField("awards", found.item, "영문", "title_en"),
      makeEntityField("awards", found.item, "국문", "title_ko"),
    ),
    makeBilingualFieldGroup(
      "기관",
      makeEntityField("awards", found.item, "영문", "organization_en"),
      makeEntityField("awards", found.item, "국문", "organization_ko"),
    ),
  );
  grid.append(makeEntityField("awards", found.item, "PDF 파일명 (선택)", "pdf", { full: true, placeholder: "data/document 안의 파일명.pdf" }));
  form.append(heading, grid, makeDocumentSection("awards", found.item));
}

function documentItem(collection, clientKey) {
  const items = collection === "awards" ? state.awardsDraft : state.profileDraft?.certifications;
  return (items || []).find((item) => item._clientKey === clientKey);
}

function validDocumentFilename(filename) {
  if (filename === undefined || filename === "") return true;
  return typeof filename === "string" && filename.trim() === filename && new TextEncoder().encode(filename).length <= 240
    && !/[<>:"/\\|?*\u0000-\u001f\u007f]/.test(filename) && /\.pdf$/i.test(filename)
    && !/^(?:CON|PRN|AUX|NUL|COM[1-9]|LPT[1-9])\./i.test(filename);
}

function makeDocumentSection(collection, item) {
  const section = document.createElement("section");
  section.className = "document-upload-section";
  section.dataset.documentCollection = collection;
  section.dataset.documentKey = item._clientKey;
  const dropzone = document.createElement("div");
  dropzone.className = `board-dropzone${state.boardDropBusy ? " is-receiving" : ""}`;
  dropzone.dataset.documentDropzone = "";
  dropzone.setAttribute("aria-label", "Explorer에서 PDF 한 개를 드롭하는 영역");
  dropzone.setAttribute("aria-busy", String(state.boardDropBusy));
  const title = document.createElement("strong");
  title.textContent = state.boardDropBusy ? "파일을 준비하는 중…" : "PDF 한 개를 이곳으로 드래그";
  const hint = document.createElement("p");
  hint.textContent = item.pdf ? `첨부: ${item.pdf}` : "PDF를 추가하면 홈페이지에 pdf 버튼이 표시됩니다. (최대 20MB)";
  dropzone.append(title, hint);
  const message = state.contentDropMessages[dropMessageKey(collection, item._clientKey)];
  if (message) {
    const feedback = document.createElement("p");
    feedback.className = "board-drop-feedback";
    feedback.setAttribute("role", "status");
    feedback.textContent = message;
    dropzone.append(feedback);
  }
  section.append(dropzone);
  if (item.pdf) section.append(makeButton("PDF 연결 제거", "text-button delete-button", "document-remove", "PDF 연결 제거"));
  return section;
}

function documentStageTokens() {
  const filenames = new Set([...(state.awardsDraft || []), ...(state.profileDraft?.certifications || [])].map((item) => item.pdf).filter(Boolean));
  return Object.entries(state.documentUploads).filter(([filename]) => filenames.has(filename)).map(([, item]) => item.stage_token);
}

function refreshDocuments(collection) {
  if (collection === "awards") renderEntityCollection("awards");
  else renderProfileContentEditor();
  updateDirtyState();
}

function queueDocumentPath(paths, collection, clientKey) {
  state.boardDropCount += 1;
  state.boardDropBusy = true;
  refreshDocuments(collection);
  state.stagingOps = state.stagingOps.catch(() => {}).then(async () => {
    const messageKey = dropMessageKey(collection, clientKey);
    try {
      const response = await state.bridge.StageDocument(Array.from(paths));
      const item = documentItem(collection, clientKey);
      const staged = response.items?.[0];
      if (!item) {
        await state.bridge.DiscardDocuments((response.items || []).map((document) => document.stage_token));
      } else if (staged) {
        item.pdf = staged.filename;
        state.documentUploads[staged.filename] = staged;
        state.contentDropMessages[messageKey] = "PDF를 준비했습니다. 모두 저장을 누르면 반영됩니다.";
      } else {
        state.contentDropMessages[messageKey] = (response.rejected || []).map((document) => `${document.original_name}: ${document.reason}`).join(" · ");
      }
    } catch (error) {
      state.contentDropMessages[messageKey] = errorMessage(error);
    } finally {
      state.boardDropCount = Math.max(0, state.boardDropCount - 1);
      state.boardDropBusy = state.boardDropCount > 0;
      refreshDocuments(collection);
    }
  });
}

function academicActivityItem(clientKey) {
  const items = state.academicActivitiesDraft || [];
  const index = items.findIndex((item) => item._clientKey === clientKey);
  return { items, index, item: index >= 0 ? items[index] : null };
}

function academicActivityTitle(item) {
  if (item.category === "reviews" || item.category === "editorial_service") {
    return item.journal_ko || item.journal_en || "(저널명 미입력)";
  }
  if (item.category === "invited_talks") return item.event_ko || item.event_en || "(행사명 미입력)";
  if (item.category === "conference_service") {
    return item.conference_ko || item.conference_en || "(학술대회명 미입력)";
  }
  return item.organization_ko || item.organization_en || "(기관명 미입력)";
}

function academicActivitySubtitle(item) {
  if (item.category === "reviews") return `리뷰 ${(item.completed_date_rows || []).length}회`;
  if (item.category === "invited_talks") return item.topic_ko || item.topic_en || "주제 미입력";
  return item.role_ko || item.role_en || "역할 미입력";
}

function academicActivityPeriod(item) {
  if (item.category === "reviews") return "";
  if (item.category === "invited_talks") return item.date || "날짜 미입력";
  if (!item.start_date) return "기간 미입력";
  return item.end_date ? `${item.start_date} – ${item.end_date}` : item.start_date;
}

function orderedAcademicActivities(items) {
  const order = new Map(ACADEMIC_ACTIVITY_CATEGORIES.map(({ value }, index) => [value, index]));
  return (items || [])
    .map((item, index) => ({ item, index }))
    .sort((left, right) => (order.get(left.item.category) ?? 999) - (order.get(right.item.category) ?? 999)
      || left.index - right.index)
    .map(({ item }) => item);
}

function renderAcademicActivityListRow(item, position, categoryLength) {
  const row = document.createElement("li");
  row.className = "board-item-row entity-item-row academic-activity-order-row";
  row.dataset.academicActivityKey = item._clientKey;
  const select = document.createElement("button");
  select.type = "button";
  select.className = "board-item-select";
  select.dataset.action = "academic-activity-select";
  select.disabled = state.saving || state.discarding;
  select.setAttribute("aria-label", `${academicActivityTitle(item)} 선택`);
  applyListRowSelection(row, select, item._clientKey === state.selectedAcademicActivityKey);
  const periodValue = academicActivityPeriod(item);
  if (periodValue) {
    const period = document.createElement("span");
    period.className = "board-item-date";
    period.textContent = periodValue;
    select.append(period);
  }
  const title = document.createElement("span");
  title.className = "board-item-title";
  title.textContent = academicActivityTitle(item);
  const subtitle = document.createElement("span");
  subtitle.className = "board-item-subtitle";
  subtitle.textContent = academicActivitySubtitle(item);
  const { badges, errorBadge } = makeListStateBadges(item, "academicActivityErrorBadge");
  badges.append(errorBadge);
  select.append(title, subtitle, badges);
  const actions = document.createElement("span");
  actions.className = "content-item-actions entity-item-actions academic-activity-order-actions";
  const up = makeButton("↑", "icon-button", "academic-activity-up", `${academicActivityTitle(item)} 위로`);
  const down = makeButton("↓", "icon-button", "academic-activity-down", `${academicActivityTitle(item)} 아래로`);
  const remove = makeButton("삭제", "text-button delete-button", "academic-activity-delete", "학술활동 삭제");
  up.disabled = up.disabled || position === 0;
  down.disabled = down.disabled || position === categoryLength - 1;
  actions.append(up, down, remove);
  row.append(select, actions);
  return row;
}

function renderAcademicActivityGroup(category) {
  const items = (state.academicActivitiesDraft || []).filter((item) => item.category === category.value);
  const group = document.createElement("li");
  group.className = "academic-activity-list-group";
  group.dataset.academicActivityCategory = category.value;
  const heading = document.createElement("div");
  heading.className = "academic-activity-group-heading";
  const title = document.createElement("h4");
  title.textContent = category.label;
  const actions = document.createElement("span");
  actions.className = "list-heading-actions";
  const count = document.createElement("span");
  count.className = "count-chip";
  count.textContent = String(items.length);
  const add = makeButton("+", "button button-primary", "academic-activity-add", `${category.label} 추가`);
  add.dataset.academicActivityAdd = category.value;
  add.disabled = add.disabled || (state.academicActivitiesDraft?.length || 0) >= 5000;
  actions.append(count, add);
  heading.append(title, actions);
  const categoryList = document.createElement("ol");
  categoryList.className = "academic-activity-category-list";
  categoryList.dataset.emptyLabel = "항목 없음";
  categoryList.replaceChildren(...items.map((item, position) => (
    renderAcademicActivityListRow(item, position, items.length)
  )));
  group.append(heading, categoryList);
  return group;
}

function renderAcademicActivityList() {
  const list = document.querySelector("#academic-activities-item-list");
  const count = document.querySelector("#academic-activities-item-count");
  const listColumn = document.querySelector("#panel-academic-activities .master-detail-list");
  if (!list || !count) return;
  const scrollTop = listColumn?.scrollTop || 0;
  const items = orderedAcademicActivities(state.academicActivitiesDraft);
  if (state.selectedAcademicActivityKey
    && !items.some((item) => item._clientKey === state.selectedAcademicActivityKey)) {
    state.selectedAcademicActivityKey = "";
  }
  list.replaceChildren(...ACADEMIC_ACTIVITY_CATEGORIES.map(renderAcademicActivityGroup));
  count.textContent = String((state.academicActivitiesDraft || []).length);
  if (listColumn) listColumn.scrollTop = scrollTop;
}

function makeAcademicActivityField(item, labelText, field, options = {}) {
  const label = document.createElement("label");
  label.className = `board-field${options.full ? " board-field-full" : ""}`;
  const title = document.createElement("span");
  title.textContent = labelText;
  let control;
  if (options.options) {
    control = document.createElement("select");
    options.options.forEach(({ value, label: optionLabel }) => {
      const option = document.createElement("option");
      option.value = value;
      option.textContent = optionLabel;
      control.append(option);
    });
  } else {
    control = document.createElement("input");
    control.type = options.type || "text";
    if (options.min !== undefined) control.min = String(options.min);
    if (options.step !== undefined) control.step = String(options.step);
    if (control.type === "text") control.maxLength = options.maxLength || 2000;
    control.autocomplete = "off";
  }
  control.className = "board-input";
  control.value = item[field] ?? "";
  control.disabled = state.saving || state.discarding;
  control.dataset.academicActivityKey = item._clientKey;
  control.dataset.academicActivityField = field;
  control.dataset.validationKey = academicActivityValidationKey(item, field);
  const error = document.createElement("span");
  error.className = "board-field-error";
  error.dataset.errorScope = control.dataset.validationKey;
  error.id = `academic-activity-error-${item._clientKey}-${field}`.replace(/[^a-z0-9-_]/gi, "-");
  control.setAttribute("aria-describedby", error.id);
  label.append(title, control, error);
  return label;
}

function academicActivityBilingualFields(item, title, englishField, koreanField) {
  return makeBilingualFieldGroup(
    title,
    makeAcademicActivityField(item, "영문", englishField),
    makeAcademicActivityField(item, "국문", koreanField),
  );
}

function renderReviewCompletedDates(item) {
  const rows = item.completed_date_rows || [];
  const section = document.createElement("section");
  section.className = "content-repeat-section entity-string-section-compact";
  const heading = contentSectionHeading("완료일", rows.length, "academic-review-date-add", "완료일 추가", 1000);
  section.append(heading);
  const list = document.createElement("div");
  list.className = "content-repeat-list entity-string-list-compact";
  rows.forEach((rowItem, index) => {
    const row = document.createElement("div");
    row.className = "entity-string-row entity-string-row-compact";
    row.dataset.academicReviewDateKey = rowItem._clientKey;
    const label = document.createElement("label");
    label.className = "board-field";
    const caption = document.createElement("span");
    caption.className = "sr-only";
    caption.textContent = `완료일 ${index + 1}`;
    const input = document.createElement("input");
    input.className = "board-input";
    input.type = "date";
    input.value = rowItem.value || "";
    input.disabled = state.saving || state.discarding;
    input.dataset.academicActivityKey = item._clientKey;
    input.dataset.academicReviewDateKey = rowItem._clientKey;
    input.dataset.validationKey = academicActivityValidationKey(
      item,
      `completed_dates:${rowItem._clientKey}`,
    );
    const error = document.createElement("span");
    error.className = "board-field-error";
    error.dataset.errorScope = input.dataset.validationKey;
    label.append(caption, input, error);
    row.append(
      label,
      makeButton("삭제", "text-button delete-button", "academic-review-date-delete", `완료일 ${index + 1} 삭제`),
    );
    list.append(row);
  });
  section.append(list);
  const sectionError = document.createElement("span");
  sectionError.className = "board-field-error";
  sectionError.dataset.errorScope = academicActivityValidationKey(item, "completed_dates");
  section.append(sectionError);
  return section;
}

function renderAcademicActivityEditor() {
  const placeholder = document.querySelector("#academic-activities-editor-placeholder");
  const form = document.querySelector("#academic-activities-form");
  const found = academicActivityItem(state.selectedAcademicActivityKey);
  form.replaceChildren();
  if (!found.item) {
    placeholder.hidden = false;
    form.hidden = true;
    return;
  }
  const item = found.item;
  placeholder.hidden = true;
  form.hidden = false;
  form.dataset.academicActivityKey = item._clientKey;
  const heading = document.createElement("div");
  heading.className = "board-form-heading";
  const title = document.createElement("h3");
  title.textContent = item._isNew ? "새 학술활동" : academicActivityTitle(item);
  heading.append(title);
  const grid = document.createElement("div");
  grid.className = "board-form-grid";
  grid.append(makeAcademicActivityField(item, "분류", "category", {
    options: ACADEMIC_ACTIVITY_CATEGORIES,
    full: true,
  }));
  if (item.category === "reviews") {
    grid.append(
      academicActivityBilingualFields(item, "저널명", "journal_en", "journal_ko"),
    );
  } else if (item.category === "invited_talks") {
    grid.append(
      academicActivityBilingualFields(item, "행사명", "event_en", "event_ko"),
      academicActivityBilingualFields(item, "주제", "topic_en", "topic_ko"),
      makeAcademicActivityField(item, "날짜", "date", { type: "date", full: true }),
    );
  } else if (item.category === "conference_service") {
    grid.append(
      academicActivityBilingualFields(item, "학술대회명", "conference_en", "conference_ko"),
      academicActivityBilingualFields(item, "역할", "role_en", "role_ko"),
      makeAcademicActivityField(item, "시작일", "start_date", { type: "date" }),
      makeAcademicActivityField(item, "종료일 (하루짜리는 비움)", "end_date", { type: "date" }),
    );
  } else if (item.category === "professional_service") {
    grid.append(
      academicActivityBilingualFields(item, "기관명", "organization_en", "organization_ko"),
      academicActivityBilingualFields(item, "역할", "role_en", "role_ko"),
      makeAcademicActivityField(item, "시작일", "start_date", { type: "date" }),
      makeAcademicActivityField(item, "종료일 (현재 진행 중이면 비움)", "end_date", { type: "date" }),
    );
  } else {
    grid.append(
      academicActivityBilingualFields(item, "저널명", "journal_en", "journal_ko"),
      academicActivityBilingualFields(item, "역할", "role_en", "role_ko"),
      makeAcademicActivityField(item, "시작일", "start_date", { type: "date" }),
      makeAcademicActivityField(item, "종료일 (현재 진행 중이면 비움)", "end_date", { type: "date" }),
    );
  }
  form.append(heading, makeEditorFormSection("기본 정보", grid));
  if (item.category === "reviews") form.append(renderReviewCompletedDates(item));
}

function renderAcademicActivities() {
  renderAcademicActivityList();
  renderAcademicActivityEditor();
}

function renderPublicationAuthors(item) {
  const section = document.createElement("section");
  section.className = "content-repeat-section";
  section.append(contentSectionHeading("저자", item.author_keys.length));
  const candidates = sortedPeopleForDisplay(
    (state.peopleDraft || []).filter((person) => !item.author_keys.includes(person._clientKey)),
  );
  const addRow = document.createElement("div");
  addRow.className = "relationship-add-row";
  const select = document.createElement("select");
  select.className = "board-input";
  select.dataset.publicationAuthorSelect = "";
  select.setAttribute("aria-label", "추가할 저자");
  const placeholder = document.createElement("option");
  placeholder.value = "";
  placeholder.textContent = "저자 선택";
  select.append(placeholder);
  candidates.forEach((person) => {
    const option = document.createElement("option");
    option.value = person._clientKey;
    option.textContent = personRelationshipLabel(person);
    select.append(option);
  });
  select.disabled = state.saving || state.discarding || !candidates.length;
  addRow.append(select);
  section.append(addRow);
  const list = document.createElement("ol");
  list.className = "relationship-list";
  item.author_keys.forEach((key, index) => {
    const person = (state.peopleDraft || []).find((candidate) => candidate._clientKey === key);
    const row = document.createElement("li");
    row.className = "relationship-row";
    row.dataset.publicationAuthorKey = key;
    const label = document.createElement("span");
    label.className = "relationship-person";
    const name = document.createElement("strong");
    name.textContent = person ? personNameLabel(person) : "연결된 사람을 찾을 수 없음";
    const affiliation = document.createElement("span");
    affiliation.textContent = person ? personAffiliationLabel(person) || "소속/메모 없음" : "";
    label.append(name, affiliation);
    row.append(label, repeatRowActions("publication-author", index, item.author_keys.length, "저자 제거"));
    list.append(row);
  });
  if (!item.author_keys.length) {
    const empty = document.createElement("p");
    empty.className = "relationship-empty";
    empty.textContent = "선택한 저자가 없습니다.";
    empty.dataset.errorScope = contentValidationKey("publications", item._clientKey, "author_keys");
    section.append(empty);
  } else {
    const error = document.createElement("span");
    error.className = "board-field-error";
    error.dataset.errorScope = contentValidationKey("publications", item._clientKey, "author_keys");
    section.append(list, error);
  }
  return section;
}

function renderPublicationAwardField(item, awardOptions) {
  const wrapper = document.createElement("div");
  wrapper.className = "publication-award-field";
  wrapper.append(makeEntityField("publications", item, "수상", "award_key", { options: awardOptions }));
  if (!item.award_key) return wrapper;

  const award = (state.awardsDraft || []).find((candidate) => candidate._clientKey === item.award_key);
  const detail = document.createElement("div");
  detail.className = "relationship-selection-detail";
  if (!award) {
    detail.textContent = "연결된 수상 정보를 찾을 수 없습니다.";
  } else {
    const title = [...new Set([award.title_ko, award.title_en].filter(Boolean))].join(" · ");
    const organization = [...new Set([award.organization_ko, award.organization_en].filter(Boolean))].join(" · ");
    detail.textContent = [award.date, title, organization].filter(Boolean).join(" — ");
  }
  wrapper.append(detail);
  return wrapper;
}

function publicationReviewRegion(element, item) {
  if (item._status !== "under_review") return element;
  element.classList.add("publication-review-disabled");
  element.inert = true;
  element.setAttribute("aria-disabled", "true");
  element.querySelectorAll("input, select, textarea, button").forEach((control) => {
    control.disabled = true;
  });
  return element;
}

function renderPublicationEditor() {
  const placeholder = document.querySelector("#publications-editor-placeholder");
  const form = document.querySelector("#publications-form");
  const found = entityItem("publications", state.selectedPublicationKey);
  form.replaceChildren();
  if (!found.item) {
    placeholder.hidden = false;
    form.hidden = true;
    return;
  }
  const item = found.item;
  placeholder.hidden = true;
  form.hidden = false;
  form.dataset.entityCollection = "publications";
  form.dataset.entityKey = item._clientKey;
  const heading = document.createElement("div");
  heading.className = "board-form-heading";
  const title = document.createElement("h3");
  title.dataset.editorHeading = "publications";
  title.textContent = item.title_ko || item.title_en || "(제목 미입력)";
  heading.append(title);
  const typeOptions = [
    { value: "international-journal", label: "국제학술지" },
    { value: "domestic-journal", label: "국내학술지" },
    { value: "international-conference", label: "국제학술대회" },
    { value: "domestic-conference", label: "국내학술대회" },
  ];
  const topicOptions = [{ value: "", label: "주제 없음" }].concat(sharedTopicOptions());
  const awardOptions = [{ value: "", label: "수상 없음" }].concat(
    sortedAwards(state.awardsDraft).map((award) => ({
      value: award._clientKey,
      label: award.title_ko || award.title_en || "(제목 미입력)",
    })),
  );
  if (item.award_key && !awardOptions.some((option) => option.value === item.award_key)) {
    awardOptions.push({ value: item.award_key, label: "연결된 수상을 찾을 수 없음" });
  }
  const basicGrid = document.createElement("div");
  basicGrid.className = "board-form-grid publication-basic-grid";
  basicGrid.append(
    makeBilingualFieldGroup(
      "제목",
      makeEntityField("publications", item, "영문", "title_en", { maxLength: null }),
      makeEntityField("publications", item, "국문", "title_ko", { maxLength: null }),
    ),
    makeEntityField("publications", item, "상태", "_status", {
      options: [
        { value: "published", label: "출판됨" },
        { value: "under_review", label: "심사 중" },
        { value: "in_press", label: "게재 예정" },
      ],
    }),
    makeEntityField("publications", item, "논문 유형", "publication_type", { options: typeOptions }),
    makeEntityField("publications", item, "주제", "topic", { options: topicOptions }),
    makeEntityField("publications", item, "DOI", "doi", { maxLength: null }),
    publicationReviewRegion(makeEntityField("publications", item, "URL", "url", { maxLength: null }), item),
    publicationReviewRegion(renderPublicationAwardField(item, awardOptions), item),
  );

  const abstractGroup = makeBilingualFieldGroup(
    "초록",
    makeEntityField("publications", item, "영문", "abstract_en", { textarea: true, controlClass: "publication-abstract", maxLength: null }),
    makeEntityField("publications", item, "국문", "abstract_ko", { textarea: true, controlClass: "publication-abstract", maxLength: null }),
    { stacked: true, hideTitle: true },
  );
  const noteGrid = document.createElement("div");
  noteGrid.className = "board-form-grid";
  noteGrid.append(
    makeEntityField("publications", item, "내용", "note", { full: true, textarea: true, controlClass: "publication-note", maxLength: null }),
  );
  form.append(
    heading,
    makeEditorFormSection("기본 정보", basicGrid),
    publicationReviewRegion(renderPublicationBibliography(item), item),
    renderPublicationAuthors(item),
    renderBilingualStringSection("publications", item, "키워드", {
      english: { field: "keywords_en_rows", ariaLabel: "영문 키워드" },
      korean: { field: "keywords_ko_rows", ariaLabel: "국문 키워드" },
      options: { maxLength: null, compact: true },
    }),
    makeEditorFormSection("초록", abstractGroup),
    makeEditorFormSection("비고", noteGrid),
  );
}

function renderEntityCollection(collection) {
  renderEntityList(collection);
  if (collection === "people") renderPersonEditor();
  if (collection === "awards") renderAwardEditor();
  if (collection === "publications") {
    renderPublicationEditor();
    renderTaxonomy("topics");
  }
}

function validateColorThemes() {
  window.SiteThemes.customIDs.forEach((id) => {
    const theme = state.draft.custom_themes[id];
    const name = String(theme.name || "").trim();
    if (!name || [...name].length > 80 || /[\u0000-\u001f\u007f-\u009f]/.test(name)) {
      state.validationErrors.set(`settings|${id}|name`, "테마 이름을 제어문자 없이 1~80자로 입력해 주세요.");
    }
    if (!SITE_COLOR_THEMES.some((preset) => preset.id === theme.base_theme)) {
      state.validationErrors.set(`settings|${id}|base_theme`, "기본 테마를 선택해 주세요.");
    }
    Object.entries(theme.colors).forEach(([key, value]) => {
      if (!window.SiteThemes.fields.some((field) => field.key === key) || typeof value !== "string" || !/^#[0-9a-f]{6}$/i.test(value)) {
        state.validationErrors.set(`settings|${id}|colors|${key}`, "#RRGGBB 형식의 색상 코드를 입력해 주세요.");
      }
    });
  });
}

function renderThemeColorFields(theme) {
  const fields = document.querySelector("#theme-color-fields");
  fields.replaceChildren();
  const groups = new Map();
  window.SiteThemes.fields.forEach(({ key, label, group }) => {
    if (!groups.has(group)) {
      const section = document.createElement("section");
      section.className = "theme-color-group";
      const heading = document.createElement("h5");
      heading.textContent = group;
      const grid = document.createElement("div");
      grid.className = "theme-color-grid";
      section.append(heading, grid);
      groups.set(group, grid);
      fields.append(section);
    }
    const row = document.createElement("div");
    row.className = "theme-color-field";
    const caption = document.createElement("span");
    caption.id = `theme-label-${theme.id}-${key}`;
    caption.textContent = label;
    const controls = document.createElement("div");
    controls.className = "theme-color-inputs";
    const swatch = document.createElement(theme.custom ? "input" : "span");
    swatch.className = "theme-color-swatch";
    if (theme.custom) {
      swatch.type = "color";
      swatch.value = theme.colors[key];
      swatch.dataset.themeId = theme.id;
      swatch.dataset.themeColorKey = key;
      swatch.dataset.editorControl = "";
      swatch.setAttribute("aria-label", `${label} 색상 선택`);
    } else {
      swatch.style.backgroundColor = theme.colors[key];
      swatch.setAttribute("aria-hidden", "true");
    }
    const code = document.createElement("input");
    code.type = "text";
    code.className = "board-input theme-color-code";
    code.maxLength = 7;
    code.spellcheck = false;
    code.autocomplete = "off";
    code.readOnly = !theme.custom;
    code.value = theme.custom ? state.draft.custom_themes[theme.id].colors[key] ?? theme.colors[key] : theme.colors[key];
    code.setAttribute("aria-labelledby", caption.id);
    code.dataset.editorControl = "";
    if (theme.custom) {
      code.dataset.themeId = theme.id;
      code.dataset.themeColorKey = key;
      code.dataset.validationKey = `settings|${theme.id}|colors|${key}`;
    }
    const error = document.createElement("span");
    error.className = "board-field-error";
    if (theme.custom) {
      error.dataset.errorScope = `settings|${theme.id}|colors|${key}`;
      error.id = `theme-error-${theme.id}-${key}`;
      code.setAttribute("aria-describedby", error.id);
    }
    controls.append(swatch, code);
    row.append(caption, controls, error);
    groups.get(group).append(row);
  });
}

function renderShortcutSettings() {
  const list = document.querySelector("#shortcut-list");
  if (!list || !state.draft) return;
  const settings = state.draft.editor_shortcuts || (state.draft.editor_shortcuts = { ...DEFAULT_EDITOR_SHORTCUTS });
  list.replaceChildren();
  Object.entries(EDITOR_SHORTCUT_LABELS).forEach(([action, label]) => {
    const row = document.createElement("div");
    row.className = "shortcut-row";
    const name = document.createElement("span");
    name.className = "shortcut-row-label";
    name.textContent = label;
    const value = document.createElement("kbd");
    value.dataset.empty = String(!settings[action]);
    value.textContent = settings[action] || "사용 안 함";
    const record = document.createElement("button");
    record.type = "button";
    record.className = "button button-secondary";
    record.dataset.shortcutRecord = action;
    record.dataset.editorControl = "";
    record.setAttribute("aria-pressed", String(state.shortcutCaptureAction === action));
    record.textContent = state.shortcutCaptureAction === action ? "키 입력 대기 중…" : "단축키 변경";
    record.addEventListener("click", () => {
      state.shortcutCaptureAction = action;
      document.querySelector("#shortcut-setting-status").textContent = `${label}: 원하는 키 조합을 누르세요. Esc는 취소합니다.`;
      renderShortcutSettings();
      list.querySelector(`[data-shortcut-record="${action}"]`)?.focus();
    });
    const clear = document.createElement("button");
    clear.type = "button";
    clear.className = "button button-secondary";
    clear.dataset.shortcutClear = action;
    clear.dataset.editorControl = "";
    clear.textContent = "끄기";
    clear.disabled = !settings[action] || state.loading || state.saving || state.discarding || state.boardDropBusy;
    clear.addEventListener("click", () => setEditorShortcut(action, ""));
    const reset = document.createElement("button");
    reset.type = "button";
    reset.className = "button button-secondary";
    reset.dataset.shortcutReset = action;
    reset.dataset.editorControl = "";
    reset.textContent = "기본값";
    reset.disabled = settings[action] === DEFAULT_EDITOR_SHORTCUTS[action] || state.loading || state.saving || state.discarding || state.boardDropBusy;
    reset.addEventListener("click", () => setEditorShortcut(action, DEFAULT_EDITOR_SHORTCUTS[action]));
    row.append(name, value, record, clear, reset);
    list.append(row);
  });
  const resetAll = document.querySelector("#shortcut-reset-all");
  if (resetAll) {
    resetAll.disabled = Object.entries(DEFAULT_EDITOR_SHORTCUTS).every(([action, shortcut]) => settings[action] === shortcut)
      || state.loading || state.saving || state.discarding || state.boardDropBusy;
    resetAll.onclick = () => {
      state.shortcutCaptureAction = "";
      state.draft.editor_shortcuts = { ...DEFAULT_EDITOR_SHORTCUTS };
      document.querySelector("#shortcut-setting-status").textContent = "기본 단축키를 복원했습니다. 저장하면 적용됩니다.";
      renderShortcutSettings();
      updateDirtyState();
    };
  }
  const descriptions = {
    save: "모두 저장",
    citation_update: "인용수 갱신 (Citation update)",
    push: "저장 후 Git push",
    revert: "변경 취소",
    settings: "설정 열기",
  };
  Object.entries({ saveButton, citationUpdateButton, pushButton, discardButton }).forEach(([name, button]) => {
    const action = { saveButton: "save", citationUpdateButton: "citation_update", pushButton: "push", discardButton: "revert" }[name];
    const chord = settings[action];
    const description = descriptions[action];
    button.title = chord ? `${description} (${chord})` : `${description} (단축키 사용 안 함)`;
    button.setAttribute("aria-label", button.title);
    if (chord) button.setAttribute("aria-keyshortcuts", chord.replace("Ctrl", "Control").replace("Comma", ","));
    else button.removeAttribute("aria-keyshortcuts");
  });
}

function setEditorShortcut(action, shortcut) {
  if (!state.draft || !Object.hasOwn(DEFAULT_EDITOR_SHORTCUTS, action)) return;
  if (shortcut) {
    const conflictingAction = Object.entries(state.draft.editor_shortcuts || {}).find(([otherAction, value]) => otherAction !== action && value === shortcut)?.[0];
    if (conflictingAction) {
      const status = document.querySelector("#shortcut-setting-status");
      status.textContent = `${shortcut}는 이미 ${EDITOR_SHORTCUT_LABELS[conflictingAction]}에 지정되어 있습니다.`;
      status.dataset.kind = "error";
      return;
    }
  }
  state.draft.editor_shortcuts[action] = shortcut;
  state.shortcutCaptureAction = "";
  const status = document.querySelector("#shortcut-setting-status");
  status.textContent = `${EDITOR_SHORTCUT_LABELS[action]} 단축키를 ${shortcut || "사용 안 함"}(으)로 변경했습니다. 저장하면 적용됩니다.`;
  status.dataset.kind = "";
  renderShortcutSettings();
  updateDirtyState();
}

function editorShortcutFromEvent(event) {
  const code = event.code || "";
  let key = "";
  if (/^Key[A-Z]$/.test(code)) key = code.slice(3);
  else if (/^Digit[0-9]$/.test(code)) key = code.slice(5);
  else if (/^F(?:[1-9]|1[0-2])$/.test(code)) key = code;
  else if (code === "Comma") key = "Comma";
  if (!key) return "";
  const modifiers = [];
  if (event.ctrlKey || event.metaKey) modifiers.push("Ctrl");
  if (event.altKey) modifiers.push("Alt");
  if (event.shiftKey) modifiers.push("Shift");
  if (!modifiers.length) return "";
  return [...modifiers, key].join("+");
}

function handleEditorShortcutKeydown(event) {
  if (state.shortcutCaptureAction) {
    if (event.isComposing) return;
    event.preventDefault();
    event.stopImmediatePropagation();
    if (event.repeat) return;
    if (event.key === "Escape") {
      state.shortcutCaptureAction = "";
      document.querySelector("#shortcut-setting-status").textContent = "단축키 입력을 취소했습니다.";
      renderShortcutSettings();
      return;
    }
    const chord = editorShortcutFromEvent(event);
    if (!chord) {
      document.querySelector("#shortcut-setting-status").textContent = "Ctrl, Alt 또는 Shift를 포함한 지원 키 조합을 입력하세요.";
      return;
    }
    if (RESERVED_EDITOR_SHORTCUTS.has(chord)) {
      document.querySelector("#shortcut-setting-status").textContent = "기본 텍스트 편집 또는 창 제어 단축키는 사용할 수 없습니다.";
      return;
    }
    if (!state.draft.editor_shortcuts) state.draft.editor_shortcuts = { ...DEFAULT_EDITOR_SHORTCUTS };
    const conflict = Object.entries(state.draft.editor_shortcuts).find(([action, value]) => action !== state.shortcutCaptureAction && value === chord)?.[0];
    if (conflict) {
      document.querySelector("#shortcut-setting-status").textContent = `${chord}는 이미 ${EDITOR_SHORTCUT_LABELS[conflict]}에 지정되어 있습니다.`;
      return;
    }
    setEditorShortcut(state.shortcutCaptureAction, chord);
    return;
  }
  if (event.defaultPrevented || event.isComposing) return;
  const chord = editorShortcutFromEvent(event);
  if (!chord || RESERVED_EDITOR_SHORTCUTS.has(chord)) return;
  const action = Object.entries(state.draft?.editor_shortcuts || DEFAULT_EDITOR_SHORTCUTS).find(([, value]) => value === chord)?.[0];
  if (!action) return;
  event.preventDefault();
  event.stopImmediatePropagation();
  if (event.repeat) return;
  if (state.loading || state.saving || state.discarding || state.boardDropBusy || document.querySelector("dialog[open]")) return;
  if (action === "save") saveChanges();
  else if (action === "citation_update") updateCitations();
  else if (action === "push") pushChanges();
  else if (action === "revert") discardChanges();
  else if (action === "settings") {
    const tab = tabs.find((item) => item.dataset.panel === "settings");
    if (tab) activateTab(tab);
  }
}

function renderAppearanceSettings(renderDetails = true) {
  const options = document.querySelector("#color-theme-options");
  if (!options.childElementCount) {
    [...SITE_COLOR_THEMES, ...window.SiteThemes.customIDs.map((id) => ({ id }))].forEach((theme) => {
      const option = document.createElement("label");
      option.className = "theme-option";
      option.dataset.colorTheme = theme.id;
      const input = document.createElement("input");
      input.type = "radio";
      input.name = "color-theme";
      input.value = theme.id;
      input.dataset.siteColorTheme = "";
      input.dataset.editorControl = "";
      input.setAttribute("aria-labelledby", `color-theme-title-${theme.id}`);
      input.setAttribute("aria-describedby", `color-theme-description-${theme.id}`);

      const preview = document.createElement("span");
      preview.className = "theme-preview";
      preview.setAttribute("aria-hidden", "true");
      const sidebar = document.createElement("span");
      sidebar.className = "theme-preview-sidebar";
      const main = document.createElement("span");
      main.className = "theme-preview-main";
      for (let index = 0; index < 3; index += 1) {
        const line = document.createElement("span");
        line.className = "theme-preview-line";
        main.append(line);
      }
      preview.append(sidebar, main);

      const title = document.createElement("span");
      title.id = `color-theme-title-${theme.id}`;
      title.className = "theme-option-title";
      const description = document.createElement("span");
      description.id = `color-theme-description-${theme.id}`;
      description.className = "theme-option-description";
      option.append(input, preview, title, description);
      if (window.SiteThemes.customIDs.includes(theme.id)) {
        const error = document.createElement("span");
        error.className = "theme-option-error";
        error.dataset.errorScope = `settings|${theme.id}|`;
        option.append(error);
      }
      options.append(option);
    });
  }
  const themeID = normaliseSiteColorTheme(state.draft?.color_theme);
  const settings = state.draft || { color_theme: themeID };
  const selected = window.SiteThemes.applyEditor(settings);
  options.querySelectorAll(".theme-option").forEach((option) => {
    const id = option.querySelector("input").value;
    const theme = window.SiteThemes.resolve(settings, id);
    window.SiteThemes.applyEditor({ ...settings, color_theme: id }, option);
    option.querySelector(".theme-option-title").textContent = theme.label;
    option.querySelector(".theme-option-description").textContent = theme.custom ? "이름과 색상을 직접 편집" : theme.description;
  });
  options.querySelectorAll("input").forEach((input) => {
    input.checked = input.value === themeID;
    input.disabled = !state.draft || state.loading || state.saving || state.discarding || state.boardDropBusy;
  });
  document.querySelector("#theme-details-title").textContent = selected.label;
  const customControls = document.querySelector("#theme-custom-controls");
  customControls.hidden = !selected.custom;
  if (!selected.custom) {
    const name = document.querySelector("#theme-custom-name");
    delete name.dataset.validationKey;
    name.setCustomValidity("");
    delete document.querySelector("#theme-name-error").dataset.errorScope;
  }
  if (renderDetails) {
    if (selected.custom) {
      const name = document.querySelector("#theme-custom-name");
      name.value = state.draft.custom_themes[themeID].name;
      name.dataset.themeId = themeID;
      name.dataset.validationKey = `settings|${themeID}|name`;
      name.setAttribute("aria-describedby", "theme-name-error");
      document.querySelector("#theme-name-error").dataset.errorScope = name.dataset.validationKey;
      const base = document.querySelector("#theme-base-preset");
      if (!base.options.length) SITE_COLOR_THEMES.forEach((preset) => {
        const option = document.createElement("option");
        option.value = preset.id;
        option.textContent = preset.label;
        base.append(option);
      });
      base.value = selected.baseTheme;
    }
    renderThemeColorFields(selected);
  }
  const status = document.querySelector("#theme-setting-status");
  const changed = state.draft && state.baseline && settingsSubsetComparable(state.draft, ["color_theme", "custom_themes"])
    !== settingsSubsetComparable(state.baseline, ["color_theme", "custom_themes"]);
  status.hidden = Boolean(state.draft) && !changed;
  if (!state.draft) {
    status.textContent = state.loading
      ? "설정을 불러오는 중…"
      : "데이터를 불러온 뒤 테마를 변경할 수 있습니다.";
  } else if (changed) {
    status.textContent = "에디터에 미리 적용했습니다. 상단의 모두 저장 버튼을 누르면 홈페이지와 CV에도 적용됩니다.";
  } else {
    status.textContent = "";
  }
}

function normaliseCatalogItem(item) {
  const result = { ...item };
  ["name", "publisher", "location", "proceedings_publisher", "proceedings_title", "proceedings_series"].forEach((base) => {
    if (Object.hasOwn(result, base) && !Object.hasOwn(result, `${base}_en`) && !Object.hasOwn(result, `${base}_ko`)) {
      result[`${base}_${/\p{Script=Hangul}/u.test(result[base]) ? "ko" : "en"}`] = result[base];
    }
    delete result[base];
  });
  return result;
}

function catalogPayload(items = []) {
  return items.map(({ _clientKey, ...item }) => ({ ...normaliseCatalogItem(item), indexing: [...(item.indexing || [])] }));
}

function catalogName(item) {
  return String(item.name_ko || "").trim() || String(item.name_en || "").trim() || "(이름 미입력)";
}

function publicationCatalogKind(item) {
  return String(item.publication_type).endsWith("-conference") ? "conferences" : "journals";
}

function publicationMetadata(item) {
  return window.PublicationBibliography.resolvePublication(item, state.draft);
}

function updatePublicationReferencePreview(item) {
  const preview = document.querySelector("#publication-reference-preview");
  if (!preview) return;
  const resolved = publicationMetadata(item);
  preview.textContent = [
    [resolved.date || "날짜 미지정", resolved.reference_ko || resolved.reference_en || (resolved.conference_info_en || resolved.conference_info_ko ? "" : "저널·학술대회를 선택해 주세요.")].filter(Boolean).join(" · "),
    resolved.conference_info_ko || resolved.conference_info_en,
  ].filter(Boolean).join("\n");
}

function renderPublicationBibliography(item) {
  const grid = document.createElement("div");
  grid.className = "board-form-grid";
  grid.append(renderPublicationCatalogField(item));
  publicationBibliographyFields.forEach(([field, label]) => {
    if (publicationCatalogKind(item) === "conferences" && field !== "pages") return;
    grid.append(makeEntityField("publications", item, label, field, {
      maxLength: 500,
      placeholder: field.endsWith("_date") ? "YYYY / YYYY-MM / YYYY-MM-DD" : field === "pages" ? "101–110" : "",
    }));
  });
  const preview = document.createElement("p");
  preview.id = "publication-reference-preview";
  preview.className = "field-hint";
  preview.style.gridColumn = "1 / -1";
  preview.style.whiteSpace = "pre-line";
  const resolved = publicationMetadata(item);
  preview.textContent = [
    [resolved.date || "날짜 미지정", resolved.reference_ko || resolved.reference_en || (resolved.conference_info_en || resolved.conference_info_ko ? "" : "저널·학술대회를 선택해 주세요.")].filter(Boolean).join(" · "),
    resolved.conference_info_ko || resolved.conference_info_en,
  ].filter(Boolean).join("\n");
  grid.append(preview);
  return makeEditorFormSection("서지 정보", grid);
}

function catalogPublicationUsage(kind, id) {
  if (!id) return 0;
  const field = kind === "journals" ? "journal_id" : "conference_id";
  return (state.publicationsDraft || []).filter((item) => item[field] === id).length;
}

function renderPublicationCatalogField(item) {
  const kind = publicationCatalogKind(item);
  const label = kind === "journals" ? "저널" : "학술대회";
  const field = kind === "journals" ? "journal_id" : "conference_id";
  const options = [{ value: "", label: "선택 안 함" }].concat(state.draft[kind]
    .filter((row) => isPublicationCatalogID(kind, row.id))
    .map((row) => ({ value: row.id, label: [...new Set([row.name_ko, row.name_en].filter(Boolean))].join(" · ") || catalogName(row) })));
  if (item[field] && !options.some((option) => option.value === item[field])) {
    options.push({ value: item[field], label: `연결된 ${label}을 찾을 수 없음` });
  }
  return makeEntityField("publications", item, label, field, { options, full: true });
}

function isPublicationCatalogID(kind, value) {
  return typeof value === "string" && (kind === "conferences"
    ? /^[A-Za-z0-9]+(?:[ -][A-Za-z0-9]+\.?)*$/.test(value)
    : /^[a-z0-9]+(?:-[a-z0-9]+)*$/.test(value));
}

function validateCatalog(kind) {
  const items = state.draft[kind];
  const ids = new Map();
  items.forEach((item) => {
    const error = (field, message) => state.validationErrors.set(validationKey(kind, item._clientKey, field), message);
    const id = String(item.id || "").trim();
    if (!isPublicationCatalogID(kind, id)) error("id", kind === "conferences"
      ? "AIK Conf. 2024F처럼 영문·숫자·공백·하이픈·약어 마침표로 입력해 주세요."
      : "ID는 소문자 영문, 숫자와 단일 하이픈으로 입력해 주세요.");
    const matches = ids.get(id) || [];
    matches.push(item);
    ids.set(id, matches);
    if (!String(item.name_en || "").trim() && !String(item.name_ko || "").trim()) {
      ["name_en", "name_ko"].forEach((field) => error(field, "영문 또는 국문 이름을 입력해 주세요."));
    }
    catalogConfig[kind].fields.forEach(([field, label, type]) => {
      if (type === "bilingual") {
        ["en", "ko"].forEach((language) => {
          const value = String(item[`${field}_${language}`] || "").trim();
          if ([...value].length > 500 || /[\u0000-\u001f\u007f-\u009f]/.test(value)) error(`${field}_${language}`, `${label}은(는) 제어문자 없이 500자 이하로 입력해 주세요.`);
        });
        return;
      }
      const value = String(item[field] || "").trim();
      if ([...value].length > (type === "url" ? 4096 : 500) || /[\u0000-\u001f\u007f-\u009f]/.test(value)) error(field, `${label}은(는) 제어문자 없이 ${type === "url" ? 4096 : 500}자 이하로 입력해 주세요.`);
      if (type === "date" && value && !isCanonicalDate(value)) error(field, "YYYY-MM-DD 형식의 실제 날짜를 입력해 주세요.");
      if (type === "publication-date" && value && !isPublicationDate(value)) error(field, "실제 날짜를 YYYY / YYYY-MM / YYYY-MM-DD 형식으로 입력해 주세요.");
      if (type === "url" && value) {
        try { const url = new URL(value); if (!["http:", "https:"].includes(url.protocol) || !url.hostname) throw new Error(); }
        catch (_) { error(field, "http 또는 https URL을 입력해 주세요."); }
      }
      if (field.endsWith("issn") && value && !/^[0-9]{4}-[0-9]{3}[0-9x]$/i.test(value)) error(field, "ISSN은 1234-567X 형식으로 입력해 주세요.");
    });
    if (item.end_date && (!item.start_date || item.start_date > item.end_date)) error("end_date", "시작일을 입력하고 종료일보다 늦지 않게 지정해 주세요.");
    if (!Array.isArray(item.indexing) || item.indexing.length > 1 || item.indexing.some((value) => !catalogIndexingOptions.some(([key]) => key === value))) error("indexing", "등재 구분은 제공된 항목 중 하나만 선택해 주세요.");
  });
  ids.forEach((matches, id) => {
    if (id && matches.length > 1) matches.forEach((item) => state.validationErrors.set(validationKey(kind, item._clientKey, "id"), "같은 목록에서 ID가 중복될 수 없습니다."));
  });
}

function renderCatalogList(kind) {
  const list = document.querySelector(`#${kind}-item-list`);
  const column = list.closest(".master-detail-list");
  const scrollTop = column.scrollTop;
  const items = state.draft[kind];
  list.replaceChildren();
  items.forEach((item) => {
    const row = document.createElement("li");
    row.className = "board-item-row";
    row.dataset.catalogKind = kind;
    row.dataset.catalogKey = item._clientKey;
    const select = makeButton("", "board-item-select", "catalog-select", `${catalogName(item)} 선택`);
    applyListRowSelection(row, select, catalogSelection[kind] === item._clientKey);
    const title = document.createElement("span");
    title.className = "board-item-title";
    title.textContent = catalogName(item);
    const subtitle = document.createElement("span");
    subtitle.className = "board-item-date";
    const usage = catalogPublicationUsage(kind, item.id);
    subtitle.textContent = [(item.indexing || []).join(" · "), usage ? `논문 ${usage}건` : ""].filter(Boolean).join(" | ");
    const error = document.createElement("span");
    error.className = "board-badge board-error-badge";
    error.dataset.catalogError = "";
    error.textContent = "오류";
    error.hidden = true;
    select.append(title);
    if (subtitle.textContent) select.append(subtitle);
    select.append(error);
    const remove = makeButton("삭제", "text-button delete-button board-item-delete", "catalog-delete", `${catalogConfig[kind].label} 삭제`);
    if (usage) {
      remove.disabled = true;
      remove.title = "연결된 논문에서 선택을 해제한 뒤 삭제할 수 있습니다.";
    }
    row.append(select, remove);
    list.append(row);
  });
  document.querySelector(`#${kind}-item-count`).textContent = String(items.length);
  column.scrollTop = scrollTop;
}

function makeCatalogField(kind, item, field, labelText, type = "text") {
  const label = document.createElement("label");
  label.className = `board-field${["url", "proceedings_series_volume"].includes(field) ? " board-field-full" : ""}`;
  const title = document.createElement("span");
  title.textContent = labelText;
  const input = document.createElement("input");
  input.className = "board-input";
  input.type = type === "publication-date" ? "text" : type;
  input.value = item[field] || "";
  input.maxLength = type === "url" ? 4096 : 500;
  input.autocomplete = "off";
  input.disabled = state.loading || state.saving || state.discarding || state.boardDropBusy;
  input.dataset.catalogKind = kind;
  input.dataset.catalogKey = item._clientKey;
  input.dataset.catalogField = field;
  input.dataset.validationKey = validationKey(kind, item._clientKey, field);
  if (field.endsWith("issn")) input.placeholder = "1234-567X";
  if (type === "publication-date") input.placeholder = "YYYY / YYYY-MM / YYYY-MM-DD";
  const error = document.createElement("span");
  error.className = "board-field-error";
  error.dataset.errorScope = input.dataset.validationKey;
  error.id = `catalog-error-${kind}-${item._clientKey.replace(/[^a-z0-9-]/gi, "-")}-${field}`;
  input.setAttribute("aria-describedby", error.id);
  label.append(title, input, error);
  return label;
}

function renderCatalog(kind) {
  const items = state.draft[kind];
  if (!items.some((item) => item._clientKey === catalogSelection[kind])) catalogSelection[kind] = items[0]?._clientKey || "";
  renderCatalogList(kind);
  const item = items.find((row) => row._clientKey === catalogSelection[kind]);
  const form = document.querySelector(`#${kind}-form`);
  form.replaceChildren();
  form.hidden = !item;
  document.querySelector(`#${kind}-editor-placeholder`).hidden = Boolean(item);
  if (!item) return;
  const makeSection = (titleText) => {
    const section = document.createElement("section");
    section.className = "catalog-form-section";
    const heading = document.createElement("div");
    heading.className = "board-form-heading";
    const title = document.createElement("h3");
    title.textContent = titleText;
    heading.append(title);
    const grid = document.createElement("div");
    grid.className = "board-form-grid";
    section.append(heading, grid);
    form.append(section);
    return grid;
  };
  const grid = makeSection(`${catalogConfig[kind].label} 정보`);
  let proceedingsGrid;
  catalogConfig[kind].fields.forEach(([field, label, type]) => {
    const fieldGrid = field.startsWith("proceedings_")
      ? (proceedingsGrid ||= makeSection("Proceedings 정보"))
      : grid;
    fieldGrid.append(type === "bilingual"
      ? makeBilingualFieldGroup(label,
        makeCatalogField(kind, item, `${field}_en`, "영문"),
        makeCatalogField(kind, item, `${field}_ko`, "국문"))
      : makeCatalogField(kind, item, field, label, type));
  });
  const indexing = document.createElement("fieldset");
  indexing.className = "catalog-indexing board-field-full";
  const legend = document.createElement("legend");
  legend.textContent = "등재 구분";
  indexing.append(legend);
  [["", "없음"], ...catalogIndexingOptions].forEach(([value, labelText]) => {
    const label = document.createElement("label");
    const radio = document.createElement("input");
    radio.type = "radio";
    radio.name = `catalog-indexing-${kind}-${item._clientKey}`;
    radio.value = value;
    radio.checked = value ? item.indexing.includes(value) : item.indexing.length === 0;
    radio.disabled = state.loading || state.saving || state.discarding || state.boardDropBusy;
    radio.dataset.catalogKind = kind;
    radio.dataset.catalogKey = item._clientKey;
    radio.dataset.catalogField = "indexing";
    label.append(radio, document.createTextNode(labelText));
    indexing.append(label);
  });
  const error = document.createElement("span");
  error.className = "board-field-error";
  error.dataset.errorScope = validationKey(kind, item._clientKey, "indexing");
  indexing.append(error);
  (proceedingsGrid || grid).append(indexing);
}

function renderAll() {
  if (!state.draft) return;
  renderAppearanceSettings();
  renderShortcutSettings();
  syncEmptyMainSections();
  syncEmptyCVSections();
  renderSectionList("visible");
  renderSectionList("hidden");
  renderProfileContentEditor();
  renderCVHierarchy();
  renderTaxonomy("topics");
  Object.keys(catalogConfig).forEach(renderCatalog);
  renderBoardList();
  renderBoardEditor();
  renderContentList("projects");
  renderProjectEditor();
  renderContentList("software");
  renderSoftwareEditor();
  renderEntityList("people");
  renderPersonEditor();
  renderEntityList("awards");
  renderAwardEditor();
  renderAcademicActivities();
  renderEntityList("publications");
  renderPublicationEditor();
  validateDraft();
  updateToolbar();
}

function findSection(section, zone) {
  const key = zone === "visible" ? "main_page_sections" : "hidden_main_page_sections";
  return { key, items: state.draft[key], index: state.draft[key].indexOf(section) };
}

function placeMainSection(section, sourceZone, destinationZone, options = {}) {
  if (sourceZone === "hidden" && destinationZone === "visible" && mainSectionItemCount(section) === 0) {
    setProfileSectionFeedback("");
    announce(`${SECTION_LABELS[section]} 섹션에는 노출할 항목이 없습니다.`);
    return false;
  }
  const source = findSection(section, sourceZone);
  if (source.index < 0) return false;
  const destinationIndex = Number.isInteger(options.delta)
    ? source.index + options.delta
    : null;
  if (destinationIndex !== null && (
    sourceZone !== destinationZone || destinationIndex < 0 || destinationIndex >= source.items.length
  )) return false;

  source.items.splice(source.index, 1);
  const destination = findSection(section, destinationZone);
  let index;
  if (destinationIndex !== null) {
    index = destinationIndex;
  } else if (options.beforeSection) {
    index = destination.items.indexOf(options.beforeSection);
  } else {
    index = destination.items.length;
  }
  if (index < 0) index = destination.items.length;
  destination.items.splice(Math.min(index, destination.items.length), 0, section);
  setProfileSectionFeedback("");
  renderSectionList("visible");
  renderSectionList("hidden");
  updateDirtyState();
  return true;
}

function moveSectionWithin(section, zone, delta) {
  if (placeMainSection(section, zone, zone, { delta })) {
    announce(`${SECTION_LABELS[section]} 섹션을 ${delta < 0 ? "위" : "아래"}로 이동했습니다.`);
  }
}

function toggleSection(section, zone) {
  const destinationZone = zone === "visible" ? "hidden" : "visible";
  if (placeMainSection(section, zone, destinationZone)) {
    announce(`${SECTION_LABELS[section]} 섹션을 ${destinationZone === "visible" ? "노출" : "숨김"}으로 변경했습니다.`);
  }
}

function moveSectionByDrop(section, sourceZone, destinationZone, beforeSection) {
  if (placeMainSection(section, sourceZone, destinationZone, { beforeSection })) {
    announce(`${SECTION_LABELS[section]} 섹션을 이동했습니다.`);
  }
}

function findCVSection(section, zone) {
  const key = zone === "visible" ? "cv_sections" : "hidden_cv_sections";
  return { key, items: state.draft[key], index: state.draft[key].indexOf(section) };
}

function placeCVSection(section, sourceZone, destinationZone, beforeSection = null) {
  if (sourceZone === "hidden" && destinationZone === "visible" && cvSectionItemCount(section) === 0) {
    setCVSectionFeedback(emptyCVSectionFeedback(section));
    announce(`${CV_SECTION_LABELS[section] || section} 섹션에는 CV에 노출할 항목이 없습니다.`);
    return false;
  }
  const source = findCVSection(section, sourceZone);
  if (source.index < 0) return false;
  source.items.splice(source.index, 1);
  const destination = findCVSection(section, destinationZone);
  let index = beforeSection ? destination.items.indexOf(beforeSection) : destination.items.length;
  if (index < 0) index = destination.items.length;
  destination.items.splice(index, 0, section);
  setCVSectionFeedback("");
  renderCVHierarchy();
  updateDirtyState();
  return true;
}

function moveCVSectionWithin(section, zone, delta) {
  const found = findCVSection(section, zone);
  const destination = found.index + delta;
  if (found.index < 0 || destination < 0 || destination >= found.items.length) return;
  const beforeSection = delta < 0 ? found.items[destination] : found.items[destination + 1] || null;
  if (placeCVSection(section, zone, zone, beforeSection)) {
    announce(`${CV_SECTION_LABELS[section] || section} 섹션을 ${delta < 0 ? "위" : "아래"}로 이동했습니다.`);
  }
}

function toggleCVSection(section, zone) {
  const destinationZone = zone === "visible" ? "hidden" : "visible";
  if (placeCVSection(section, zone, destinationZone)) {
    announce(`${CV_SECTION_LABELS[section] || section} 섹션을 ${destinationZone === "visible" ? "노출" : "숨김"}으로 변경했습니다.`);
  }
}

function moveCVSectionByDrop(section, sourceZone, destinationZone, beforeSection) {
  if (placeCVSection(section, sourceZone, destinationZone, beforeSection)) {
    announce(`${CV_SECTION_LABELS[section] || section} 섹션을 이동했습니다.`);
  }
}

function taxonomyItem(kind, clientKey) {
  const config = taxonomyConfig[kind];
  const items = state.draft[config.listKey];
  return { config, items, index: items.findIndex((item) => item._clientKey === clientKey) };
}

function moveTaxonomy(kind, clientKey, delta) {
  const found = taxonomyItem(kind, clientKey);
  const destination = found.index + delta;
  if (found.index < 0 || destination < 0 || destination >= found.items.length) return;
  [found.items[found.index], found.items[destination]] = [found.items[destination], found.items[found.index]];
  renderTaxonomy(kind);
  refreshTopicEditors();
  updateDirtyState();
  announce(`${found.config.itemName} 순서를 이동했습니다.`);
}

function moveTaxonomyByDrop(kind, clientKey, beforeClientKey) {
  const found = taxonomyItem(kind, clientKey);
  if (found.index < 0) return;
  const [item] = found.items.splice(found.index, 1);
  let destination = beforeClientKey
    ? found.items.findIndex((candidate) => candidate._clientKey === beforeClientKey)
    : found.items.length;
  if (destination < 0) destination = found.items.length;
  found.items.splice(destination, 0, item);
  renderTaxonomy(kind);
  refreshTopicEditors();
  updateDirtyState();
  announce(`${found.config.itemName} 순서를 이동했습니다.`);
}

function addTaxonomy(kind) {
  const config = taxonomyConfig[kind];
  const clientKey = `new:${state.nextClientKey++}`;
  state.draft[config.listKey].push({ id: "", label_en: "", label_ko: "", _clientKey: clientKey });
  renderTaxonomy(kind);
  refreshTopicEditors();
  updateDirtyState();
  requestAnimationFrame(() => {
    const row = Array.from(document.querySelectorAll(`[data-kind="${kind}"]`))
      .find((element) => element.dataset.clientKey === clientKey);
    row?.querySelector('[data-field="label_en"]')?.focus();
  });
}

function sharedTopicOptions() {
  return (state.draft.topics || []).filter((topic) => topic.id).map((topic) => ({
    value: topic.id,
    label: [...new Set([topic.label_ko, topic.label_en].filter(Boolean))].join(" · "),
  }));
}

function refreshTopicEditors() {
  renderContentList("projects");
  renderProjectEditor();
  renderEntityList("publications");
  renderPublicationEditor();
}

function deleteTaxonomy(kind, clientKey) {
  const found = taxonomyItem(kind, clientKey);
  if (found.index < 0) return;
  const item = found.items[found.index];
  if (taxonomyUsage(kind, item) > 0) return;
  found.items.splice(found.index, 1);
  renderTaxonomy(kind);
  refreshTopicEditors();
  updateDirtyState();
  announce(`${found.config.itemName}을(를) 삭제했습니다.`);
}

function localToday() {
  const now = new Date();
  const year = String(now.getFullYear()).padStart(4, "0");
  const month = String(now.getMonth() + 1).padStart(2, "0");
  const day = String(now.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

function addBoardPost() {
  const item = {
    _clientKey: `new-board:${state.nextBoardKey++}`,
    _isNew: true,
    start_date: localToday(),
    end_date: "",
    title_en: "",
    title_ko: "",
    content_en: "",
    content_ko: "",
    media: [],
  };
  state.boardDraft.push(item);
  state.selectedBoardKey = item._clientKey;
  state.boardDropMessage = "";
  renderBoardList();
  renderBoardEditor();
  resetEditorScroll("board");
  updateDirtyState();
  requestAnimationFrame(() => {
    document.querySelector(`[data-board-key="${item._clientKey}"][data-board-field="title_en"]`)?.focus();
  });
}

function queueDiscardBoardMedia(tokens) {
  if (!tokens.length || !state.bridge) return;
  state.stagingOps = state.stagingOps
    .catch(() => {})
    .then(() => state.bridge.DiscardBoardMedia(tokens))
    .catch(() => {});
}

function removeProfileMedia() {
  const media = state.profileMediaDraft;
  if (!media) return;
  if (media.stage_token) queueDiscardBoardMedia([media.stage_token]);
  media.stage_token = "";
  media.remove = true;
  media.original_name = "";
  media.size = 0;
  media.preview_url = "";
  state.profileDropMessage = "저장하면 프로필 이미지 연결이 제거됩니다.";
  renderProfileContentEditor();
  updateDirtyState();
}

function deleteBoardPost(clientKey) {
  const found = boardItem(clientKey);
  if (!found.item) return;
  const deletingSelected = state.selectedBoardKey === clientKey;
  if (
    !found.item._isNew
    && !confirmDeleteItem("게시글", found.item.title_ko || found.item.title_en || "제목 미입력")
  ) return;
  queueDiscardBoardMedia(found.item.media.map((media) => media.stage_token).filter(Boolean));
  state.boardDraft.splice(found.index, 1);
  const sorted = sortedBoardItems(state.boardDraft);
  if (state.selectedBoardKey === clientKey) {
    state.selectedBoardKey = sorted[0]?._clientKey || "";
  }
  state.boardDropMessage = "";
  renderBoardList();
  renderBoardEditor();
  if (deletingSelected) resetEditorScroll("board");
  updateDirtyState();
  announce("게시글을 삭제했습니다.");
}

function deleteBoardMedia(boardKey, mediaKey) {
  const found = boardItem(boardKey);
  if (!found.item) return;
  const mediaIndex = found.item.media.findIndex((media) => media._clientKey === mediaKey);
  if (mediaIndex < 0) return;
  const [removed] = found.item.media.splice(mediaIndex, 1);
  if (removed.stage_token) queueDiscardBoardMedia([removed.stage_token]);
  renderBoardList();
  renderBoardEditor();
  updateDirtyState();
}

function renderContentCollection(collection) {
  renderContentList(collection);
  if (collection === "projects") {
    renderProjectEditor();
    renderTaxonomy("topics");
  } else {
    renderSoftwareEditor();
  }
  syncCVSectionsAfterItemsChange();
  renderCVHierarchy();
  updateDirtyState();
}

function addContentItem(collection) {
  if ((contentDraft(collection)?.length || 0) >= 500) return;
  const today = localToday();
  let item;
  if (collection === "projects") {
    item = {
      _clientKey: `new-project:${state.nextProjectKey++}`,
      _isNew: true,
      visible_in_CV: true,
      start_date: today,
      end_date: today,
      title_en: "",
      title_ko: "",
      theme: (state.draft.topics || []).find((theme) => theme.id)?.id || "",
      funder_en: "",
      funder_ko: "",
      notes_en_rows: [{ _clientKey: `project-note-en:new:${state.nextContentRowKey++}`, value: "" }],
      notes_kr_rows: [{ _clientKey: `project-note-kr:new:${state.nextContentRowKey++}`, value: "" }],
      media: [],
    };
  } else {
    item = {
      _clientKey: `new-software:${state.nextSoftwareKey++}`,
      _isNew: true,
      visible_in_CV: true,
      id: "",
      name: "",
      stage: "development",
      links: [],
      notes_en_rows: [{ _clientKey: `software-note-en:new:${state.nextContentRowKey++}`, value: "" }],
      notes_kr_rows: [{ _clientKey: `software-note-kr:new:${state.nextContentRowKey++}`, value: "" }],
      technologies: [{ _clientKey: `software-technology:new:${state.nextContentRowKey++}`, value: "" }],
      media: [],
    };
  }
  contentDraft(collection).unshift(item);
  setSelectedContentKey(collection, item._clientKey);
  state.contentDropMessages[dropMessageKey(collection, item._clientKey)] = "";
  renderContentCollection(collection);
  resetEditorScroll(collection);
  requestAnimationFrame(() => {
    const field = collection === "projects" ? "title_en" : "name";
    document.querySelector(
      `[data-content-collection="${collection}"][data-content-key="${item._clientKey}"][data-content-field="${field}"]`,
    )?.focus();
  });
}

function moveContentItemToPosition(collection, clientKey, destinationPosition) {
  const found = contentItem(collection, clientKey);
  if (!found.item || !moveManualItemToPosition(found.items, clientKey, destinationPosition)) return;
  renderContentCollection(collection);
  announce(`${contentSingular(collection)} 순서를 이동했습니다.`);
}

function moveContentItem(collection, clientKey, delta) {
  const found = contentItem(collection, clientKey);
  if (!found.item) return;
  moveContentItemToPosition(collection, clientKey, found.index + delta);
}

function deleteContentItem(collection, clientKey) {
  const found = contentItem(collection, clientKey);
  if (!found.item) return;
  const deletingSelected = selectedContentKey(collection) === clientKey;
  const title = collection === "projects"
    ? found.item.title_ko || found.item.title_en || "제목 미입력"
    : found.item.name || "이름 미입력";
  if (!found.item._isNew && !confirmDeleteItem(contentSingular(collection), title)) return;
  queueDiscardBoardMedia(found.item.media.map((media) => media.stage_token).filter(Boolean));
  found.items.splice(found.index, 1);
  delete state.contentDropMessages[dropMessageKey(collection, clientKey)];
  if (selectedContentKey(collection) === clientKey) {
    setSelectedContentKey(collection, found.items[Math.min(found.index, found.items.length - 1)]?._clientKey || "");
  }
  renderContentCollection(collection);
  if (deletingSelected) resetEditorScroll(collection);
  announce(`${contentSingular(collection)}를 삭제했습니다.`);
}

function repeatCollection(item, kind) {
  if (kind === "note-en") return item.notes_en_rows;
  if (kind === "note-kr") return item.notes_kr_rows;
  if (kind === "link") return item.links;
  if (kind === "technology") return item.technologies;
  if (kind === "media") return item.media;
  return null;
}

function addContentRow(collection, clientKey, kind) {
  const found = contentItem(collection, clientKey);
  const rows = found.item && repeatCollection(found.item, kind);
  if (!rows) return;
  if (rows.length >= 100) return;
  let focusTechnologyKey = "";
  if (kind === "note-en" || kind === "note-kr") {
    rows.push({ _clientKey: `${collection}-${kind}:new:${state.nextContentRowKey++}`, value: "" });
  } else if (kind === "link") {
    rows.push({
      _clientKey: `software-link:new:${state.nextContentRowKey++}`,
      url: "",
      label: "",
      label_en: "",
      label_ko: "",
    });
  } else if (kind === "technology") {
    const blank = rows.find((technology) => !String(technology.value || "").trim());
    if (blank) {
      focusTechnologySelect(blank._clientKey);
      return;
    }
    if (!canAddTechnology(found.item)) return;
    focusTechnologyKey = `software-technology:new:${state.nextContentRowKey++}`;
    rows.push({ _clientKey: focusTechnologyKey, value: "" });
  }
  renderContentCollection(collection);
  if (focusTechnologyKey) focusTechnologySelect(focusTechnologyKey);
}

function moveContentRow(collection, clientKey, kind, rowKey, delta) {
  const found = contentItem(collection, clientKey);
  const rows = found.item && repeatCollection(found.item, kind);
  const index = rows?.findIndex((row) => row._clientKey === rowKey) ?? -1;
  const destination = index + delta;
  if (!rows || index < 0 || destination < 0 || destination >= rows.length) return;
  [rows[index], rows[destination]] = [rows[destination], rows[index]];
  renderContentCollection(collection);
}

function moveContentRowToPosition(collection, clientKey, kind, rowKey, destinationPosition) {
  const found = contentItem(collection, clientKey);
  const rows = found.item && repeatCollection(found.item, kind);
  const index = rows?.findIndex((row) => row._clientKey === rowKey) ?? -1;
  if (!rows || index < 0 || rows.length < 2) return false;
  const destination = Math.max(0, Math.min(Math.trunc(destinationPosition), rows.length - 1));
  if (index === destination) return false;
  const [moved] = rows.splice(index, 1);
  rows.splice(destination, 0, moved);
  renderContentCollection(collection);
  return true;
}

function focusTechnologyDragHandle(rowKey) {
  requestAnimationFrame(() => {
    const row = [...document.querySelectorAll("#software-form .content-technology-item[data-content-row-key]")]
      .find((candidate) => candidate.dataset.contentRowKey === rowKey);
    row?.querySelector('[data-action="drag-technology"]')?.focus({ preventScroll: true });
  });
}

function focusTechnologySelect(rowKey) {
  requestAnimationFrame(() => {
    const select = [...document.querySelectorAll("#software-form .content-technology-select[data-content-row-key]")]
      .find((candidate) => candidate.dataset.contentRowKey === rowKey);
    select?.focus({ preventScroll: true });
  });
}

function rerenderTechnologiesSection(item, focusRowKey = "") {
  const current = document.querySelector("#software-form .software-technologies-section");
  if (!current) return;
  current.replaceWith(renderTechnologiesSection(item));
  if (focusRowKey) focusTechnologySelect(focusRowKey);
}

function deleteContentRow(collection, clientKey, kind, rowKey) {
  const found = contentItem(collection, clientKey);
  const rows = found.item && repeatCollection(found.item, kind);
  const index = rows?.findIndex((row) => row._clientKey === rowKey) ?? -1;
  if (!rows || index < 0) return;
  const [removed] = rows.splice(index, 1);
  if (kind === "media" && removed.stage_token) queueDiscardBoardMedia([removed.stage_token]);
  renderContentCollection(collection);
}

function addAcademicActivity(categoryValue) {
  if ((state.academicActivitiesDraft?.length || 0) >= 5000) return;
  const category = ACADEMIC_ACTIVITY_CATEGORIES.some(({ value }) => value === categoryValue)
    ? categoryValue
    : "editorial_service";
  const reviewDateRows = category === "reviews"
    ? [{ _clientKey: `academic-review-date:new:${state.nextContentRowKey++}`, value: localToday() }]
    : [];
  const item = {
    _clientKey: `new-academic-activity:${state.nextAcademicActivityKey++}`,
    _isNew: true,
    visible_in_CV: true,
    category,
    journal_en: "",
    journal_ko: "",
    completed_date_rows: reviewDateRows,
    event_en: "",
    event_ko: "",
    topic_en: "",
    topic_ko: "",
    conference_en: "",
    conference_ko: "",
    organization_en: "",
    organization_ko: "",
    role_en: "",
    role_ko: "",
    date: localToday(),
    start_date: localToday(),
    end_date: "",
  };
  const firstCategoryIndex = state.academicActivitiesDraft.findIndex((candidate) => candidate.category === category);
  state.academicActivitiesDraft.splice(firstCategoryIndex < 0 ? state.academicActivitiesDraft.length : firstCategoryIndex, 0, item);
  state.selectedAcademicActivityKey = item._clientKey;
  renderAcademicActivities();
  resetEditorScroll("academic-activities");
  updateDirtyState();
  const firstField = {
    editorial_service: "journal_en",
    professional_service: "organization_en",
    conference_service: "conference_en",
    invited_talks: "event_en",
    reviews: "journal_en",
  }[category];
  requestAnimationFrame(() => {
    document.querySelector(
      `[data-academic-activity-key="${item._clientKey}"][data-academic-activity-field="${firstField}"]`,
    )?.focus();
  });
}

function moveAcademicActivity(clientKey, direction) {
  const found = academicActivityItem(clientKey);
  if (!found.item || ![-1, 1].includes(direction)) return;
  const categoryItems = found.items.filter((item) => item.category === found.item.category);
  const position = categoryItems.indexOf(found.item);
  const target = categoryItems[position + direction];
  if (!target) return;
  const targetIndex = found.items.indexOf(target);
  [found.items[found.index], found.items[targetIndex]] = [found.items[targetIndex], found.items[found.index]];
  renderAcademicActivityList();
  updateDirtyState();
  announce(`${academicActivityTitle(found.item)} 순서를 이동했습니다.`);
  requestAnimationFrame(() => {
    document.querySelector(
      `[data-academic-activity-key="${clientKey}"] [data-action="academic-activity-${direction < 0 ? "up" : "down"}"]`,
    )?.focus();
  });
}

function moveAcademicActivityToCategoryTop(found, category) {
  if (!found.item || found.item.category === category) return;
  found.items.splice(found.index, 1);
  found.item.category = category;
  const targetIndex = found.items.findIndex((item) => item.category === category);
  found.items.splice(targetIndex < 0 ? found.items.length : targetIndex, 0, found.item);
}

function addAcademicReviewDate(clientKey) {
  const found = academicActivityItem(clientKey);
  if (!found.item) return;
  found.item.completed_date_rows ||= [];
  if (found.item.completed_date_rows.length >= 1000) return;
  const today = localToday();
  const value = found.item.completed_date_rows.some((row) => row.value === today) ? "" : today;
  const row = {
    _clientKey: `academic-review-date:new:${state.nextContentRowKey++}`,
    value,
  };
  found.item.completed_date_rows.push(row);
  renderAcademicActivityEditor();
  updateDirtyState();
  requestAnimationFrame(() => {
    document.querySelector(`[data-academic-review-date-key="${row._clientKey}"] input`)?.focus();
  });
}

function deleteAcademicReviewDate(clientKey, rowKey) {
  const found = academicActivityItem(clientKey);
  const rows = found.item?.completed_date_rows;
  const index = rows?.findIndex((row) => row._clientKey === rowKey) ?? -1;
  if (!rows || index < 0) return;
  rows.splice(index, 1);
  renderAcademicActivityList();
  renderAcademicActivityEditor();
  updateDirtyState();
}

function deleteAcademicActivity(clientKey) {
  const found = academicActivityItem(clientKey);
  if (!found.item) return;
  if (!found.item._isNew && !confirmDeleteItem("학술활동", academicActivityTitle(found.item))) return;
  const deletingSelected = state.selectedAcademicActivityKey === clientKey;
  found.items.splice(found.index, 1);
  if (deletingSelected) {
    state.selectedAcademicActivityKey = orderedAcademicActivities(found.items)[0]?._clientKey || "";
  }
  if (syncEmptyMainSections()) {
    setProfileSectionFeedback("학술활동: 항목이 없어 자동으로 숨겼습니다.");
    renderSectionList("visible");
    renderSectionList("hidden");
  }
  syncCVSectionsAfterItemsChange();
  renderAcademicActivities();
  renderCVHierarchy();
  if (deletingSelected) resetEditorScroll("academic-activities");
  updateDirtyState();
  announce("학술활동을 삭제했습니다.");
}

function refreshEntityAfterChange(collection, renderEditor = true) {
  if (collection === "people") refreshPatentInventorEditors();
  if (collection === "awards" && syncEmptyMainSections()) {
    setProfileSectionFeedback("수상: 항목이 없어 자동으로 숨겼습니다.");
    renderSectionList("visible");
    renderSectionList("hidden");
  }
  renderEntityList(collection);
  if (renderEditor) {
    if (collection === "people") renderPersonEditor();
    if (collection === "awards") renderAwardEditor();
    if (collection === "publications") renderPublicationEditor();
  }
  if (collection === "people") renderPublicationEditor();
  if (collection === "awards") renderPublicationEditor();
  if (collection === "publications") {
    renderEntityList("people");
    renderEntityList("awards");
    renderTaxonomy("topics");
  }
  syncCVSectionsAfterItemsChange();
  renderCVHierarchy();
  updateDirtyState();
}

function addEntityItem(collection) {
  const config = entityConfig[collection];
  const items = entityDraft(collection);
  if (items.length >= (collection === "publications" ? 1000 : 500)) return;
  let item;
  if (collection === "people") {
    item = {
      _clientKey: `new-person:${state.nextPersonKey++}`,
      _isNew: true,
      id: "",
      name_en: "",
      name_ko: "",
      is_self: false,
      notes_en_rows: [],
      notes_ko_rows: [],
    };
    items.push(item);
  } else if (collection === "awards") {
    item = {
      _clientKey: `new-award:${state.nextAwardKey++}`,
      _isNew: true,
      visible_in_CV: true,
      id: "",
      date: localToday(),
      title_en: "",
      title_ko: "",
      organization_en: "",
      organization_ko: "",
    };
    items.push(item);
  } else {
    const defaultAuthor = (state.peopleDraft || []).find((person) => person.is_self)
      || (state.peopleDraft || [])[0];
    item = {
      _clientKey: `new-publication:${state.nextPublicationKey++}`,
      _isNew: true,
      visible_in_CV: true,
      title_en: "",
      title_ko: "",
      abstract_en: "",
      abstract_ko: "",
      keywords_en_rows: [],
      keywords_ko_rows: [],
      author_keys: defaultAuthor ? [defaultAuthor._clientKey] : [],
      _unresolvedAuthorIDs: {},
      date: "",
      _status: "published",
      publication_type: "international-journal",
      topic: "",
      award_key: "",
      _unresolvedAwardID: "",
      doi: "",
      url: "",
      note: "",
    };
    items.push(item);
  }
  setSelectedEntityKey(collection, item._clientKey);
  refreshEntityAfterChange(collection);
  resetEditorScroll(collection);
  requestAnimationFrame(() => {
    const field = collection === "people" ? "name_en" : "title_en";
    document.querySelector(
      `[data-entity-collection="${collection}"][data-entity-key="${item._clientKey}"][data-entity-field="${field}"]`,
    )?.focus();
  });
  announce(`새 ${config.singular} 항목을 추가했습니다.`);
}

function moveEntityItemToPosition(collection, clientKey, destinationPosition) {
  const config = entityConfig[collection];
  if (!config.manualOrder) return;
  const found = entityItem(collection, clientKey);
  if (!found.item) return;
  if (collection === "people") {
    if (found.item.is_self) return;
    if (!moveManualItemToPosition(found.items, clientKey, destinationPosition, (item) => !item.is_self)) return;
  } else {
    if (!moveManualItemToPosition(found.items, clientKey, destinationPosition)) return;
  }
  refreshEntityAfterChange(collection);
  announce(`${config.singular} 순서를 이동했습니다.`);
}

function moveEntityItem(collection, clientKey, delta) {
  const config = entityConfig[collection];
  if (!config?.manualOrder) return;
  const found = entityItem(collection, clientKey);
  if (!found.item || (collection === "people" && found.item.is_self)) return;
  const movableItems = collection === "people"
    ? found.items.filter((item) => !item.is_self)
    : found.items;
  const sourcePosition = movableItems.findIndex((item) => item._clientKey === clientKey);
  moveEntityItemToPosition(collection, clientKey, sourcePosition + delta);
}

function deleteEntityItem(collection, clientKey) {
  const config = entityConfig[collection];
  const found = entityItem(collection, clientKey);
  if (!found.item) return;
  if (collection === "people" && (found.item.is_self || personUsage(clientKey))) return;
  if (collection === "awards" && awardUsage(clientKey)) return;
  const title = collection === "people"
    ? personNameLabel(found.item)
    : found.item.title_ko || found.item.title_en || "제목 미입력";
  if (!found.item._isNew && !confirmDeleteItem(config.singular, title)) return;
  const deletingSelected = selectedEntityKey(collection) === clientKey;
  const deletedCoauthorIndex = collection === "people"
    ? found.items.filter((item) => !item.is_self).indexOf(found.item)
    : -1;
  found.items.splice(found.index, 1);
  if (selectedEntityKey(collection) === clientKey) {
    if (collection === "people") {
      const coauthors = found.items.filter((item) => !item.is_self);
      const nextPerson = coauthors[Math.min(deletedCoauthorIndex, coauthors.length - 1)]
        || found.items.find((item) => item.is_self)
        || found.items[0];
      setSelectedEntityKey(collection, nextPerson?._clientKey || "");
    } else {
      const ordered = config.sorter(found.items);
      setSelectedEntityKey(collection, ordered[Math.min(found.index, ordered.length - 1)]?._clientKey || ordered[0]?._clientKey || "");
    }
  }
  refreshEntityAfterChange(collection);
  if (deletingSelected) resetEditorScroll(collection);
  announce(`${config.singular} 항목을 삭제했습니다.`);
}

function entityStringRows(collection, clientKey, field) {
  const found = entityItem(collection, clientKey);
  return { found, rows: found.item?.[field] || null };
}

function addEntityStringRow(collection, clientKey, field) {
  const { rows } = entityStringRows(collection, clientKey, field);
  if (!rows || rows.length >= 100) return;
  const row = { _clientKey: `${collection}-${field}:new:${state.nextContentRowKey++}`, value: "" };
  rows.push(row);
  refreshEntityAfterChange(collection);
  const input = document.querySelector(
    `input[data-entity-string-key="${row._clientKey}"], textarea[data-entity-string-key="${row._clientKey}"]`,
  );
  input?.focus({ preventScroll: true });
  input?.scrollIntoView({ block: "nearest", inline: "nearest" });
}

function moveEntityStringRow(collection, clientKey, field, rowKey, delta) {
  const { rows } = entityStringRows(collection, clientKey, field);
  const index = rows?.findIndex((row) => row._clientKey === rowKey) ?? -1;
  const destination = index + delta;
  if (!rows || index < 0 || destination < 0 || destination >= rows.length) return;
  [rows[index], rows[destination]] = [rows[destination], rows[index]];
  refreshEntityAfterChange(collection);
}

function deleteEntityStringRow(collection, clientKey, field, rowKey) {
  const { rows } = entityStringRows(collection, clientKey, field);
  const index = rows?.findIndex((row) => row._clientKey === rowKey) ?? -1;
  if (!rows || index < 0) return;
  rows.splice(index, 1);
  refreshEntityAfterChange(collection);
}

function addPublicationAuthor(publicationKey, authorKey) {
  const found = entityItem("publications", publicationKey);
  if (!found.item || !authorKey || found.item.author_keys.includes(authorKey)) return;
  if (!entityItem("people", authorKey).item) return;
  found.item.author_keys.push(authorKey);
  refreshEntityAfterChange("publications");
}

function movePublicationAuthor(publicationKey, authorKey, delta) {
  const found = entityItem("publications", publicationKey);
  const index = found.item?.author_keys?.indexOf(authorKey) ?? -1;
  const destination = index + delta;
  if (!found.item || index < 0 || destination < 0 || destination >= found.item.author_keys.length) return;
  [found.item.author_keys[index], found.item.author_keys[destination]] = [
    found.item.author_keys[destination],
    found.item.author_keys[index],
  ];
  refreshEntityAfterChange("publications");
}

function deletePublicationAuthor(publicationKey, authorKey) {
  const found = entityItem("publications", publicationKey);
  const index = found.item?.author_keys?.indexOf(authorKey) ?? -1;
  if (!found.item || index < 0) return;
  found.item.author_keys.splice(index, 1);
  delete found.item._unresolvedAuthorIDs?.[authorKey];
  refreshEntityAfterChange("publications");
}

async function stageContentPaths(paths, collection, clientKey) {
  if (!paths?.length) return;
  const messageKey = dropMessageKey(collection, clientKey);
  const targetBeforeStage = contentItem(collection, clientKey);
  if (!targetBeforeStage.item) return;
  if (targetBeforeStage.item.media.length + paths.length > 60) {
    state.contentDropMessages[messageKey] = "미디어는 최대 60개까지 추가할 수 있습니다.";
    if (collection === "projects") renderProjectEditor();
    else renderSoftwareEditor();
    return;
  }
  state.contentDropMessages[messageKey] = "";
  if (collection === "projects") renderProjectEditor();
  else renderSoftwareEditor();
  updateToolbar();
  try {
    const response = await state.bridge.StageBoardMedia(Array.from(paths));
    const found = contentItem(collection, clientKey);
    const acceptedTokens = (response.items || []).map((item) => item.stage_token);
    if (!found.item) {
      queueDiscardBoardMedia(acceptedTokens);
      state.contentDropMessages[messageKey] = `${contentSingular(collection)}를 찾지 못해 사진을 추가하지 않았습니다.`;
    } else {
      (response.items || []).forEach((media) => {
        found.item.media.push({
          ...media,
          _clientKey: `new-content-media:${state.nextBoardMediaKey++}`,
          _isNew: true,
          caption_en: "",
          caption_ko: "",
        });
      });
      const rejected = response.rejected || [];
      state.contentDropMessages[messageKey] = rejected.length
        ? rejected.map((item) => `${item.original_name}: ${item.reason}`).join(" · ")
        : `${response.items.length}개 사진을 추가했습니다.`;
    }
  } catch (error) {
    state.contentDropMessages[messageKey] = errorMessage(error);
  } finally {
    renderContentList(collection);
    if (collection === "projects") renderProjectEditor();
    else renderSoftwareEditor();
    updateDirtyState();
  }
}

async function stageProfilePath(paths) {
  if (!paths?.length || !state.profileDraft) return;
  state.profileDropMessage = "";
  renderProfileContentEditor();
  updateToolbar();
  try {
    const response = await state.bridge.StageBoardMedia([paths[0]]);
    const staged = response.items?.[0];
    const media = state.profileMediaDraft;
    if (!media) {
      queueDiscardBoardMedia((response.items || []).map((item) => item.stage_token));
      state.profileDropMessage = "프로필 이미지 정보를 찾지 못했습니다.";
    } else if (staged) {
      if (media.stage_token) queueDiscardBoardMedia([media.stage_token]);
      media.stage_token = staged.stage_token;
      media.remove = false;
      media.original_name = staged.original_name || "프로필 이미지";
      media.size = Number(staged.size || 0);
      media.preview_url = staged.preview_url || "";
      state.profileDropMessage = paths.length > 1
        ? "프로필 이미지는 한 장만 사용하므로 첫 번째 이미지만 준비했습니다."
        : "새 프로필 이미지를 준비했습니다.";
    } else {
      state.profileDropMessage = (response.rejected || [])
        .map((item) => `${item.original_name}: ${item.reason}`)
        .join(" · ") || "이미지를 추가하지 못했습니다.";
    }
  } catch (error) {
    state.profileDropMessage = errorMessage(error);
  } finally {
    renderProfileContentEditor();
    updateDirtyState();
  }
}

function queueProfilePath(paths) {
  state.boardDropCount += 1;
  state.boardDropBusy = true;
  renderProfileContentEditor();
  updateToolbar();
  const operation = state.stagingOps
    .catch(() => {})
    .then(() => stageProfilePath(paths))
    .finally(() => {
      state.boardDropCount = Math.max(0, state.boardDropCount - 1);
      state.boardDropBusy = state.boardDropCount > 0;
      renderProfileContentEditor();
      updateToolbar();
    });
  state.stagingOps = operation;
}

function queueContentPaths(paths, collection, clientKey) {
  state.boardDropCount += 1;
  state.boardDropBusy = true;
  if (collection === "projects") renderProjectEditor();
  else renderSoftwareEditor();
  updateToolbar();
  const operation = state.stagingOps
    .catch(() => {})
    .then(() => stageContentPaths(paths, collection, clientKey))
    .finally(() => {
      state.boardDropCount = Math.max(0, state.boardDropCount - 1);
      state.boardDropBusy = state.boardDropCount > 0;
      if (collection === "projects") renderProjectEditor();
      else renderSoftwareEditor();
      updateToolbar();
    });
  state.stagingOps = operation;
}

async function stageBoardPaths(paths, boardKey) {
  if (!paths?.length) return;
  state.boardDropMessage = "";
  renderBoardEditor();
  updateToolbar();
  try {
    const response = await state.bridge.StageBoardMedia(Array.from(paths));
    const found = boardItem(boardKey);
    const acceptedTokens = (response.items || []).map((item) => item.stage_token);
    if (!found.item) {
      queueDiscardBoardMedia(acceptedTokens);
      state.boardDropMessage = "선택한 게시글을 찾지 못해 사진을 추가하지 않았습니다.";
    } else {
      (response.items || []).forEach((media) => {
        found.item.media.push({
          ...media,
          _clientKey: `new-media:${state.nextBoardMediaKey++}`,
          _isNew: true,
          caption_en: "",
          caption_ko: "",
        });
      });
      const rejected = response.rejected || [];
      state.boardDropMessage = rejected.length
        ? rejected.map((item) => `${item.original_name}: ${item.reason}`).join(" · ")
        : `${response.items.length}개 사진을 추가했습니다.`;
    }
  } catch (error) {
    state.boardDropMessage = errorMessage(error);
  } finally {
    renderBoardList();
    renderBoardEditor();
    updateDirtyState();
  }
}

function queueBoardPaths(paths, boardKey) {
  state.boardDropCount += 1;
  state.boardDropBusy = true;
  renderBoardEditor();
  updateToolbar();
  const operation = state.stagingOps
    .catch(() => {})
    .then(() => stageBoardPaths(paths, boardKey))
    .finally(() => {
      state.boardDropCount = Math.max(0, state.boardDropCount - 1);
      state.boardDropBusy = state.boardDropCount > 0;
      renderBoardEditor();
      updateToolbar();
    });
  state.stagingOps = operation;
}

function installFileDrop() {
  if (state.fileDropBound || !window.runtime?.OnFileDrop) return;
  window.runtime.OnFileDrop((x, y, paths) => {
    const target = document.elementFromPoint(x, y)?.closest?.(
      "[data-profile-media-dropzone], [data-board-media-dropzone], [data-content-media-dropzone], [data-document-dropzone]",
    );
    if (!target || state.saving || state.discarding || state.boardDropBusy) return;
    if (target.dataset.documentDropzone !== undefined) {
      const section = target.closest("[data-document-collection][data-document-key]");
      const { documentCollection, documentKey } = section.dataset;
      if (documentItem(documentCollection, documentKey)) queueDocumentPath(paths, documentCollection, documentKey);
      return;
    }
    if (target.dataset.profileMediaDropzone !== undefined) {
      queueProfilePath(paths);
      return;
    }
    if (target.dataset.contentMediaDropzone !== undefined) {
      const { contentCollection, contentKey } = target.dataset;
      if (!contentItem(contentCollection, contentKey).item) return;
      queueContentPaths(paths, contentCollection, contentKey);
      return;
    }
    const found = boardItem(state.selectedBoardKey);
    if (!found.item) return;
    queueBoardPaths(paths, found.item._clientKey);
  }, true);
  state.fileDropBound = true;
}

document.addEventListener("input", (event) => {
  const input = event.target;
  if (!(
    input instanceof HTMLInputElement
    || input instanceof HTMLTextAreaElement
    || input instanceof HTMLSelectElement
  )) return;
  if (input.dataset.catalogField) {
    if (!state.draft || state.loading || state.saving || state.discarding || state.boardDropBusy) return;
    const kind = input.dataset.catalogKind;
    const item = state.draft[kind]?.find((row) => row._clientKey === input.dataset.catalogKey);
    if (!item) return;
    if (input.dataset.catalogField === "indexing") {
      if (!input.checked) return;
      item.indexing = input.value ? [input.value] : [];
    } else item[input.dataset.catalogField] = input.value;
    renderCatalogList(kind);
    renderEntityList("publications");
    const publication = entityItem("publications", state.selectedPublicationKey).item;
    if (publication) updatePublicationReferencePreview(publication);
    renderCVHierarchy();
    updateDirtyState();
    return;
  }
  if (input.dataset.cvPublicationSearch !== undefined) {
    state.cvPublicationSearch = input.value;
    updateCVPublicationFilter();
    return;
  }
  if (input.dataset.listFilter) {
    if (!state.draft) return;
    const filterKeys = {
      "people-search": "peopleSearch",
      "publications-search": "publicationsSearch",
      "publications-status": "publicationsStatus",
    };
    const filterKey = filterKeys[input.dataset.listFilter];
    if (!filterKey) return;
    state.listFilters[filterKey] = input.value;
    renderEntityList(input.dataset.listFilter.startsWith("people-") ? "people" : "publications");
    return;
  }
  if (!state.draft || state.saving || state.discarding) return;
  if (input.dataset.siteColorTheme !== undefined) {
    if (!input.checked) return;
    state.draft.color_theme = normaliseSiteColorTheme(input.value);
    renderAppearanceSettings();
    updateDirtyState();
    return;
  }
  if (input.dataset.themeCustomName !== undefined) {
    const theme = state.draft.custom_themes[input.dataset.themeId];
    if (!theme) return;
    theme.name = input.value;
    renderAppearanceSettings(false);
    updateDirtyState();
    return;
  }
  if (input.dataset.themeColorKey) {
    const theme = state.draft.custom_themes[input.dataset.themeId];
    if (!theme) return;
    theme.colors[input.dataset.themeColorKey] = input.value;
    const row = input.closest(".theme-color-field");
    if (input.type === "color") row.querySelector('[type="text"]').value = input.value;
    else if (/^#[0-9a-f]{6}$/i.test(input.value)) row.querySelector('[type="color"]').value = input.value;
    renderAppearanceSettings(false);
    updateDirtyState();
    return;
  }
  if (input.dataset.publicationAuthorSelect !== undefined) {
    const publicationForm = input.closest("#publications-form[data-entity-key]");
    if (input.value) addPublicationAuthor(publicationForm?.dataset.entityKey || "", input.value);
    return;
  }
  if (input.dataset.patentInventorSearch !== undefined) {
    const item = patentByKey(input.dataset.patentInventorSearch);
    if (!item) return;
    item._inventorSearch = input.value;
    updatePatentInventorOptions(input.closest("[data-patent-inventors-key]"), item);
    return;
  }
  if (input.dataset.patentInventorSelect !== undefined) {
    const item = patentByKey(input.dataset.patentInventorSelect);
    if (!item || !input.value || item.inventor_keys.includes(input.value) || !entityItem("people", input.value).item) return;
    item.inventor_keys.push(input.value);
    refreshPatentInventorChange(item);
    return;
  }
  if (input.dataset.profilePath) {
    const path = JSON.parse(input.dataset.profilePath);
    const visibleBefore = [...state.draft.main_page_sections];
    let nextValue = input.value;
    if ((path[0] === "scholarships" && path[2] === "amount" && path[3] === "value")
      || (path[0] === "education" && path[2] === "gpa" && ["value", "max"].includes(path[3]))) {
      nextValue = input.value === "" ? "" : input.valueAsNumber;
    }
    if (path[0] === "scholarships" && path[2] === "amount" && path[3] === "currency") {
      nextValue = input.value.toUpperCase();
      input.value = nextValue;
    }
    setProfileValue(path, nextValue);
    if (path[0] === "patents") updatePatentSummary(state.profileDraft.patents[path[1]]);
    if (path[0] === "education" && path[2] === "gpa") {
      const item = state.profileDraft.education[path[1]];
      if (["value", "max"].every((key) => item.gpa?.[key] === "" || item.gpa?.[key] === undefined)) delete item.gpa;
    }
    if (syncEmptyMainSections()) {
      const autoHidden = visibleBefore.find((section) => !state.draft.main_page_sections.includes(section));
      if (autoHidden) {
        setProfileSectionFeedback(
          `${SECTION_LABELS[autoHidden]}: 노출할 완성된 항목이 없어 자동으로 숨겼습니다.`,
        );
      }
      renderSectionList("visible");
      renderSectionList("hidden");
    }
    syncCVSectionsAfterItemsChange();
    renderCVHierarchy();
  } else if (input.dataset.academicReviewDateKey) {
    const found = academicActivityItem(input.dataset.academicActivityKey);
    const row = found.item?.completed_date_rows?.find(
      (candidate) => candidate._clientKey === input.dataset.academicReviewDateKey,
    );
    if (!row) return;
    row.value = input.value;
    renderAcademicActivityList();
  } else if (input.dataset.academicActivityField) {
    const found = academicActivityItem(input.dataset.academicActivityKey);
    if (!found.item) return;
    const field = input.dataset.academicActivityField;
    if (field === "category") moveAcademicActivityToCategoryTop(found, input.value);
    else found.item[field] = input.value;
    if (field === "category" && input.value === "reviews" && !found.item.completed_date_rows?.length) {
      found.item.completed_date_rows = [{
        _clientKey: `academic-review-date:new:${state.nextContentRowKey++}`,
        value: localToday(),
      }];
    }
    if (syncEmptyMainSections()) {
      setProfileSectionFeedback("학술활동: 노출할 완성된 항목이 없어 자동으로 숨겼습니다.");
    }
    renderSectionList("visible");
    renderSectionList("hidden");
    renderAcademicActivityList();
    syncCVSectionsAfterItemsChange();
    renderCVHierarchy();
    if (field === "category") renderAcademicActivityEditor();
  } else if (input.dataset.entityField) {
    const collection = input.dataset.entityCollection;
    const found = entityItem(collection, input.dataset.entityKey);
    if (!found.item) return;
    const field = input.dataset.entityField;
    found.item[field] = input.value;
    if (collection === "publications" && (field === "title_en" || field === "title_ko")) {
      const heading = document.querySelector('#publications-form [data-editor-heading="publications"]');
      if (heading) heading.textContent = found.item.title_ko || found.item.title_en || "(제목 미입력)";
    }
    if (collection === "publications" && field === "_status") {
      if (input.value !== "published") found.item.date = "";
      renderPublicationEditor();
    }
    if (collection === "publications" && field === "publication_type") {
      if (publicationCatalogKind(found.item) === "journals") found.item.conference_id = "";
      else {
        found.item.journal_id = "";
        publicationBibliographyFields.forEach(([name]) => { if (name !== "pages") found.item[name] = ""; });
      }
      renderPublicationEditor();
    }
    if (collection === "publications" && (field === "journal_id" || field === "conference_id")) {
      Object.keys(catalogConfig).forEach(renderCatalog);
    }
    if (collection === "publications") updatePublicationReferencePreview(found.item);
    if (collection === "publications" && field === "award_key") {
      found.item._unresolvedAwardID = "";
      renderEntityList("awards");
      renderPublicationEditor();
    }
    renderEntityList(collection);
    if (collection === "publications" && field === "topic") renderTaxonomy("topics");
    if (collection === "people" || collection === "awards") renderPublicationEditor();
    if (collection === "people") {
      refreshPatentInventorEditors();
      renderCVHierarchy();
    }
  } else if (input.dataset.entityStringField) {
    const collection = input.dataset.entityCollection;
    const { rows } = entityStringRows(collection, input.dataset.entityKey, input.dataset.entityStringField);
    const row = rows?.find((candidate) => candidate._clientKey === input.dataset.entityStringKey);
    if (row) row.value = input.value;
    renderEntityList(collection);
    if (collection === "people") renderPublicationEditor();
    if (collection === "people") {
      refreshPatentInventorEditors();
      renderCVHierarchy();
    }
  } else if (input.dataset.contentField) {
    const found = contentItem(input.dataset.contentCollection, input.dataset.contentKey);
    if (found.item) found.item[input.dataset.contentField] = input.value;
    if (input.dataset.contentCollection === "projects") {
      if (input.dataset.contentField === "title_en" || input.dataset.contentField === "title_ko") {
        const heading = document.querySelector('#project-form [data-editor-heading="projects"]');
        if (heading) heading.textContent = found.item.title_ko || found.item.title_en || "(제목 미입력)";
      }
      renderContentList("projects");
      if (input.dataset.contentField === "theme") renderTaxonomy("topics");
    } else if (input.dataset.contentField === "name" || input.dataset.contentField === "stage") {
      renderContentList("software");
    }
  } else if (input.dataset.noteField) {
    const found = contentItem(input.dataset.contentCollection, input.dataset.contentKey);
    const notes = found.item?.[`notes_${input.dataset.contentNoteLanguage}_rows`];
    const row = notes?.find((candidate) => candidate._clientKey === input.dataset.contentRowKey);
    if (row) row[input.dataset.noteField] = input.value;
    const details = input.closest(".content-collapsible-row");
    if (details && found.item && row) {
      const index = notes.indexOf(row);
      const summary = details.querySelector(":scope > summary");
      if (summary) summary.textContent = noteSummary(row, index);
    }
  } else if (input.dataset.linkField) {
    const found = contentItem("software", input.dataset.contentKey);
    const row = found.item?.links?.find((candidate) => candidate._clientKey === input.dataset.contentRowKey);
    if (row) row[input.dataset.linkField] = input.value;
  } else if (input.dataset.technologyField) {
    const found = contentItem("software", input.dataset.contentKey);
    const row = found.item?.technologies?.find((candidate) => candidate._clientKey === input.dataset.contentRowKey);
    if (row) {
      row[input.dataset.technologyField] = input.value;
      rerenderTechnologiesSection(found.item, row._clientKey);
    }
    renderContentList("software");
  } else if (input.dataset.contentMediaField) {
    const found = contentItem(input.dataset.contentCollection, input.dataset.contentKey);
    const media = found.item?.media?.find((candidate) => candidate._clientKey === input.dataset.contentMediaKey);
    if (media) media[input.dataset.contentMediaField] = input.value;
  } else if (input.dataset.boardField) {
    const found = boardItem(input.dataset.boardKey);
    if (found.item) found.item[input.dataset.boardField] = input.value;
    if (input.dataset.boardField.startsWith("title_") || input.dataset.boardField.endsWith("_date")) {
      renderBoardList();
    }
  } else if (input.dataset.boardMediaField) {
    const found = boardItem(input.dataset.boardKey);
    const media = found.item?.media?.find((candidate) => candidate._clientKey === input.dataset.mediaKey);
    if (media) media[input.dataset.boardMediaField] = input.value;
  } else if (input.dataset.kind) {
    const found = taxonomyItem(input.dataset.kind, input.dataset.clientKey);
    if (found.index >= 0) found.items[found.index][input.dataset.field] = input.value;
    if (input.dataset.field === "color") {
      const control = input.closest(".taxonomy-color-control");
      if (/^#[0-9a-f]{6}$/i.test(input.value.trim()) || !input.value) {
        control.querySelector(".taxonomy-color-trigger").style.setProperty("--taxonomy-selected-color", input.value.trim()
          || taxonomyDefaultColor(input.dataset.kind, found.index));
      }
    }
    refreshTopicEditors();
  } else if (input.dataset.fallbackKind) {
    const config = taxonomyConfig[input.dataset.fallbackKind];
    state.draft[config.fallbackKey][input.dataset.field] = input.value;
  } else {
    return;
  }
  updateDirtyState();
});

document.addEventListener("click", (event) => {
  const button = event.target.closest("button");
  if (!button || button.disabled || !state.draft || state.saving || state.discarding || state.boardDropBusy) return;
  const inventorSection = button.closest("[data-patent-inventors-key]");
  if (inventorSection && button.dataset.action?.startsWith("patent-")) {
    const item = patentByKey(inventorSection.dataset.patentInventorsKey);
    if (!item) return;
    const action = button.dataset.action;
    if (action === "patent-person-add") {
      if ((state.peopleDraft?.length || 0) >= 500) return;
      activateTab(document.querySelector("#tab-people"));
      addEntityItem("people");
      item.inventor_keys.push(state.selectedPersonKey);
      refreshPatentInventorChange(item);
      return;
    }
    const key = button.closest("[data-patent-inventor-key]")?.dataset.patentInventorKey;
    const index = item.inventor_keys.indexOf(key);
    if (index < 0) return;
    if (action === "patent-inventor-edit") {
      state.selectedPersonKey = key;
      activateTab(document.querySelector("#tab-people"));
      renderEntityCollection("people");
      resetEditorScroll("people");
      document.querySelector('#people-form [data-entity-field="name_en"]')?.focus();
      return;
    }
    if (action === "patent-inventor-delete") item.inventor_keys.splice(index, 1);
    if (action === "patent-inventor-up" && index > 0) {
      [item.inventor_keys[index - 1], item.inventor_keys[index]] = [item.inventor_keys[index], item.inventor_keys[index - 1]];
    }
    if (action === "patent-inventor-down" && index < item.inventor_keys.length - 1) {
      [item.inventor_keys[index + 1], item.inventor_keys[index]] = [item.inventor_keys[index], item.inventor_keys[index + 1]];
    }
    refreshPatentInventorChange(item);
    return;
  }
  if (button.id === "theme-base-start") {
    const theme = state.draft.custom_themes[state.draft.color_theme];
    const base = document.querySelector("#theme-base-preset").value;
    if (!theme || !SITE_COLOR_THEMES.some((preset) => preset.id === base)) return;
    theme.base_theme = base;
    theme.colors = {};
    renderAppearanceSettings();
    updateDirtyState();
    return;
  }
  if (button.dataset.action === "taxonomy-color-open") {
    openTaxonomyColorPopover(button);
    return;
  }
  if (button.dataset.catalogAdd) {
    const kind = button.dataset.catalogAdd;
    if (state.draft[kind].length >= 1000) return;
    const item = Object.fromEntries(catalogConfig[kind].fields.flatMap(([field, , type]) => (
      type === "bilingual" ? [[`${field}_en`, ""], [`${field}_ko`, ""]] : [[field, ""]]
    )));
    const usedIDs = new Set(state.draft[kind].map((row) => row.id));
    do {
      item.id = randomRelationshipID(kind === "journals" ? "journal" : "conference");
    } while (usedIDs.has(item.id));
    item.indexing = [];
    item._clientKey = `catalog:${kind}:${state.nextClientKey++}`;
    state.draft[kind].push(item);
    catalogSelection[kind] = item._clientKey;
    renderCatalog(kind);
    updateDirtyState();
    document.querySelector(`#${kind}-form [data-catalog-field=name_en]`)?.focus();
    return;
  }
  const catalogRow = button.closest("[data-catalog-kind][data-catalog-key]");
  if (catalogRow && button.dataset.action?.startsWith("catalog-")) {
    const { catalogKind: kind, catalogKey: key } = catalogRow.dataset;
    const items = state.draft[kind];
    if (button.dataset.action === "catalog-select") catalogSelection[kind] = key;
    if (button.dataset.action === "catalog-delete") {
      const index = items.findIndex((item) => item._clientKey === key);
      if (index < 0) return;
      if (catalogPublicationUsage(kind, items[index].id)) return;
      const existing = state.baseline[kind].some((item) => item._clientKey === key);
      if (existing && !confirmDeleteItem(catalogConfig[kind].label, catalogName(items[index]))) return;
      items.splice(index, 1);
    }
    renderCatalog(kind);
    if (button.dataset.action === "catalog-delete") updateDirtyState();
    else applyValidationErrors();
    return;
  }

  if (button.dataset.action === "document-remove") {
    const section = button.closest("[data-document-collection][data-document-key]");
    const { documentCollection, documentKey } = section.dataset;
    const item = documentItem(documentCollection, documentKey);
    if (item) delete item.pdf;
    delete state.contentDropMessages[dropMessageKey(documentCollection, documentKey)];
    refreshDocuments(documentCollection);
    return;
  }

  if (button.id === "add-board-button") {
    addBoardPost();
    return;
  }
  if (button.id === "add-project-button") {
    addContentItem("projects");
    return;
  }
  if (button.id === "add-software-button") {
    addContentItem("software");
    return;
  }
  if (button.id === "add-person-button") {
    addEntityItem("people");
    return;
  }
  if (button.id === "add-award-button") {
    addEntityItem("awards");
    return;
  }
  if (button.id === "add-publication-button") {
    addEntityItem("publications");
    return;
  }
  if (button.dataset.academicActivityAdd) {
    addAcademicActivity(button.dataset.academicActivityAdd);
    return;
  }

  if (button.dataset.action === "profile-section-nav") {
    const targetName = button.dataset.profileNavTarget;
    if (targetName === "awards") {
      const awardsTab = document.querySelector("#tab-awards");
      if (awardsTab) activateTab(awardsTab);
      return;
    }
    const target = document.getElementById(targetName);
    if (target) target.scrollIntoView({ behavior: "smooth", block: "start" });
    return;
  }

  if (button.dataset.action === "profile-item-toggle") {
    const row = button.closest(".profile-repeat-row[data-profile-key]");
    if (!row) return;
    const expanded = button.getAttribute("aria-expanded") !== "true";
    if (expanded) state.profileExpandedRows.add(row.dataset.profileKey);
    else state.profileExpandedRows.delete(row.dataset.profileKey);
    setProfileRepeatRowExpanded(row, expanded);
    return;
  }

  if (button.dataset.action === "profile-media-delete") {
    removeProfileMedia();
    return;
  }

  if (button.dataset.cvItemSection) {
    const item = findCVItem(button.dataset.cvItemSection, button.dataset.cvItemKey);
    if (!item) return;
    item.visible_in_CV = item.visible_in_CV === false;
    syncCVSectionsAfterItemsChange();
    renderCVHierarchy();
    updateDirtyState();
    return;
  }

  if (button.dataset.action?.startsWith("profile-item-")) {
    const action = button.dataset.action.replace("profile-item-", "");
    const section = button.dataset.profileSection;
    mutateProfileItem(action, section, Number(button.dataset.profileIndex ?? -1));
    return;
  }
  const academicActivityForm = button.closest("#academic-activities-form[data-academic-activity-key]");
  if (academicActivityForm && button.dataset.action === "academic-review-date-add") {
    addAcademicReviewDate(academicActivityForm.dataset.academicActivityKey);
    return;
  }
  if (academicActivityForm && button.dataset.action === "academic-review-date-delete") {
    const dateRow = button.closest("[data-academic-review-date-key]");
    if (dateRow) {
      deleteAcademicReviewDate(
        academicActivityForm.dataset.academicActivityKey,
        dateRow.dataset.academicReviewDateKey,
      );
    }
    return;
  }

  const academicActivityRow = button.closest("[data-academic-activity-key]");
  if (academicActivityRow && button.dataset.action === "academic-activity-select") {
    state.selectedAcademicActivityKey = academicActivityRow.dataset.academicActivityKey;
    renderAcademicActivities();
    resetEditorScroll("academic-activities");
    validateDraft();
    return;
  }
  if (academicActivityRow && button.dataset.action === "academic-activity-delete") {
    deleteAcademicActivity(academicActivityRow.dataset.academicActivityKey);
    return;
  }
  if (academicActivityRow && button.dataset.action === "academic-activity-up") {
    moveAcademicActivity(academicActivityRow.dataset.academicActivityKey, -1);
    return;
  }
  if (academicActivityRow && button.dataset.action === "academic-activity-down") {
    moveAcademicActivity(academicActivityRow.dataset.academicActivityKey, 1);
    return;
  }

  const entityListRow = button.closest(".entity-item-row[data-entity-collection][data-entity-key]");
  if (entityListRow) {
    const collection = entityListRow.dataset.entityCollection;
    const clientKey = entityListRow.dataset.entityKey;
    if (button.dataset.action === "entity-select") {
      setSelectedEntityKey(collection, clientKey);
      renderEntityCollection(collection);
      resetEditorScroll(collection);
      validateDraft();
    } else if (button.dataset.action === "entity-up") {
      moveEntityItem(collection, clientKey, -1);
    } else if (button.dataset.action === "entity-down") {
      moveEntityItem(collection, clientKey, 1);
    } else if (button.dataset.action === "entity-delete") {
      deleteEntityItem(collection, clientKey);
    }
    return;
  }

  const entityForm = button.closest(
    "#people-form[data-entity-collection], #awards-form[data-entity-collection], #publications-form[data-entity-collection]",
  );
  if (entityForm) {
    const collection = entityForm.dataset.entityCollection;
    const clientKey = entityForm.dataset.entityKey;
    const action = button.dataset.action || "";
    if (action === "entity-string-add") {
      addEntityStringRow(collection, clientKey, button.dataset.stringField);
    } else {
      const stringRow = button.closest("[data-entity-string-key][data-entity-string-field]");
      const authorRow = button.closest("[data-publication-author-key]");
      if (stringRow && action.startsWith("entity-string-")) {
        const operation = action.replace("entity-string-", "");
        if (operation === "up") moveEntityStringRow(collection, clientKey, stringRow.dataset.entityStringField, stringRow.dataset.entityStringKey, -1);
        if (operation === "down") moveEntityStringRow(collection, clientKey, stringRow.dataset.entityStringField, stringRow.dataset.entityStringKey, 1);
        if (operation === "delete") deleteEntityStringRow(collection, clientKey, stringRow.dataset.entityStringField, stringRow.dataset.entityStringKey);
      }
      if (authorRow && action.startsWith("publication-author-")) {
        const operation = action.replace("publication-author-", "");
        if (operation === "up") movePublicationAuthor(clientKey, authorRow.dataset.publicationAuthorKey, -1);
        if (operation === "down") movePublicationAuthor(clientKey, authorRow.dataset.publicationAuthorKey, 1);
        if (operation === "delete") deletePublicationAuthor(clientKey, authorRow.dataset.publicationAuthorKey);
      }
    }
    return;
  }

  const contentListRow = button.closest(".content-item-row[data-content-collection][data-content-key]");
  if (contentListRow) {
    const collection = contentListRow.dataset.contentCollection;
    const clientKey = contentListRow.dataset.contentKey;
    if (button.dataset.action === "content-select") {
      setSelectedContentKey(collection, clientKey);
      state.contentDropMessages[dropMessageKey(collection, clientKey)] = "";
      renderContentList(collection);
      if (collection === "projects") renderProjectEditor();
      else renderSoftwareEditor();
      resetEditorScroll(collection);
      validateDraft();
    } else if (button.dataset.action === "content-up") {
      moveContentItem(collection, clientKey, -1);
    } else if (button.dataset.action === "content-down") {
      moveContentItem(collection, clientKey, 1);
    } else if (button.dataset.action === "content-delete") {
      deleteContentItem(collection, clientKey);
    }
    return;
  }

  const contentForm = button.closest("#project-form[data-content-collection], #software-form[data-content-collection]");
  if (contentForm) {
    const collection = contentForm.dataset.contentCollection;
    const clientKey = contentForm.dataset.contentKey;
    const action = button.dataset.action || "";
    if (action === "content-note-en-add") addContentRow(collection, clientKey, "note-en");
    else if (action === "content-note-kr-add") addContentRow(collection, clientKey, "note-kr");
    else if (action === "content-link-add") addContentRow(collection, clientKey, "link");
    else if (action === "content-technology-add") addContentRow(collection, clientKey, "technology");
    else {
      const mediaRow = button.closest("[data-content-media-key]");
      const repeatRow = button.closest("[data-content-row-key]");
      const rowKey = mediaRow?.dataset.contentMediaKey || repeatRow?.dataset.contentRowKey;
      const match = action.match(/^content-(note-en|note-kr|link|technology|media)-(up|down|delete)$/);
      if (match && rowKey) {
        const [, kind, operation] = match;
        if (operation === "up") moveContentRow(collection, clientKey, kind, rowKey, -1);
        if (operation === "down") moveContentRow(collection, clientKey, kind, rowKey, 1);
        if (operation === "delete") deleteContentRow(collection, clientKey, kind, rowKey);
      }
    }
    return;
  }

  const boardRow = button.closest("[data-board-key]");
  if (boardRow && button.dataset.action === "board-select") {
    const selectedKey = boardRow.dataset.boardKey;
    state.selectedBoardKey = selectedKey;
    state.boardDropMessage = "";
    renderBoardList();
    renderBoardEditor();
    resetEditorScroll("board");
    validateDraft();
    requestAnimationFrame(() => {
      document.querySelector(`[data-board-key="${selectedKey}"] [data-action="board-select"]`)?.focus({ preventScroll: true });
    });
    return;
  }
  if (boardRow && button.dataset.action === "board-delete") {
    deleteBoardPost(boardRow.dataset.boardKey);
    return;
  }
  const mediaRow = button.closest("[data-media-key]");
  if (mediaRow && button.dataset.action === "board-media-delete") {
    deleteBoardMedia(state.selectedBoardKey, mediaRow.dataset.mediaKey);
    return;
  }

  if (button.dataset.addTaxonomy) {
    addTaxonomy(button.dataset.addTaxonomy);
    return;
  }

  const webPageRow = button.closest("[data-web-page]");
  if (webPageRow && button.dataset.action === "web-page-toggle") {
    toggleWebPage(webPageRow.dataset.webPage);
    return;
  }

  const sectionRow = button.closest("[data-section]");
  if (sectionRow) {
    if (button.dataset.action === "section-up") moveSectionWithin(sectionRow.dataset.section, sectionRow.dataset.zone, -1);
    if (button.dataset.action === "section-down") moveSectionWithin(sectionRow.dataset.section, sectionRow.dataset.zone, 1);
    if (button.dataset.action === "section-toggle") toggleSection(sectionRow.dataset.section, sectionRow.dataset.zone);
    return;
  }

  const cvSectionRow = button.closest("[data-cv-section]");
  if (cvSectionRow) {
    const { cvSection: section, cvZone: zone } = cvSectionRow.dataset;
    if (button.dataset.action === "cv-section-expand") {
      if (state.cvExpandedSections.has(section)) state.cvExpandedSections.delete(section);
      else state.cvExpandedSections.add(section);
      renderCVHierarchy();
    }
    if (button.dataset.action === "cv-section-up") moveCVSectionWithin(section, zone, -1);
    if (button.dataset.action === "cv-section-down") moveCVSectionWithin(section, zone, 1);
    if (button.dataset.action === "cv-section-toggle") toggleCVSection(section, zone);
    if (button.dataset.action === "cv-items-select-all") setCVItemsVisibility(section, true);
    if (button.dataset.action === "cv-items-clear-all") setCVItemsVisibility(section, false);
    return;
  }

  const taxonomyRow = button.closest("[data-kind][data-client-key]");
  if (taxonomyRow) {
    const kind = taxonomyRow.dataset.kind;
    const clientKey = taxonomyRow.dataset.clientKey;
    if (button.dataset.action === "taxonomy-up") moveTaxonomy(kind, clientKey, -1);
    if (button.dataset.action === "taxonomy-down") moveTaxonomy(kind, clientKey, 1);
    if (button.dataset.action === "taxonomy-delete") deleteTaxonomy(kind, clientKey);
  }
});

document.addEventListener("dragstart", (event) => {
  const handle = event.target.closest(".drag-handle");
  if (!handle || state.saving || state.discarding) return;
  const technologyRow = handle.closest(".content-technology-item[data-content-row-key]");
  const technologyList = technologyRow?.closest("[data-technology-sort-list][data-content-key]");
  if (technologyRow && technologyList && handle.dataset.action === "drag-technology") {
    state.drag = {
      type: "content-technology",
      collection: "software",
      clientKey: technologyList.dataset.contentKey,
      rowKey: technologyRow.dataset.contentRowKey,
    };
    technologyRow.classList.add("is-dragging");
    event.dataTransfer.effectAllowed = "move";
    event.dataTransfer.setData("text/plain", technologyRow.dataset.contentRowKey);
    return;
  }
  const contentRow = handle.closest(".manual-order-row[data-content-collection][data-content-key]");
  if (contentRow && handle.dataset.action === "drag-content-item") {
    state.drag = {
      type: "content-item",
      collection: contentRow.dataset.contentCollection,
      clientKey: contentRow.dataset.contentKey,
    };
    contentRow.classList.add("is-dragging");
    event.dataTransfer.effectAllowed = "move";
    event.dataTransfer.setData("text/plain", contentRow.dataset.contentKey);
    return;
  }
  const entityRow = handle.closest(".manual-order-row[data-entity-collection][data-entity-key]");
  if (entityRow && handle.dataset.action === "drag-entity-item") {
    state.drag = {
      type: "entity-item",
      collection: entityRow.dataset.entityCollection,
      clientKey: entityRow.dataset.entityKey,
    };
    entityRow.classList.add("is-dragging");
    event.dataTransfer.effectAllowed = "move";
    event.dataTransfer.setData("text/plain", entityRow.dataset.entityKey);
    return;
  }
  const sectionRow = handle.closest("[data-section]");
  if (sectionRow) {
    state.drag = { type: "section", section: sectionRow.dataset.section, zone: sectionRow.dataset.zone };
    sectionRow.classList.add("is-dragging");
    event.dataTransfer.effectAllowed = "move";
    event.dataTransfer.setData("text/plain", sectionRow.dataset.section);
    return;
  }
  const cvSectionRow = handle.closest("[data-cv-section]");
  if (cvSectionRow) {
    state.drag = { type: "cv-section", section: cvSectionRow.dataset.cvSection, zone: cvSectionRow.dataset.cvZone };
    cvSectionRow.classList.add("is-dragging");
    event.dataTransfer.effectAllowed = "move";
    event.dataTransfer.setData("text/plain", cvSectionRow.dataset.cvSection);
    return;
  }
  const taxonomyRow = handle.closest("[data-kind][data-client-key]");
  if (taxonomyRow) {
    state.drag = { type: "taxonomy", kind: taxonomyRow.dataset.kind, clientKey: taxonomyRow.dataset.clientKey };
    taxonomyRow.classList.add("is-dragging");
    event.dataTransfer.effectAllowed = "move";
    event.dataTransfer.setData("text/plain", taxonomyRow.dataset.clientKey);
  }
});

document.addEventListener("dragend", () => {
  document.querySelectorAll(".is-dragging").forEach((element) => element.classList.remove("is-dragging"));
  state.drag = null;
});

function manualOrderDropDestination(list, row, clientKey, event) {
  const rows = [...list.querySelectorAll(":scope > .manual-order-row")];
  const sourcePosition = rows.findIndex((candidate) => (
    candidate.dataset.contentKey === clientKey || candidate.dataset.entityKey === clientKey
  ));
  if (sourcePosition < 0) return -1;
  if (!row) return rows.length - 1;
  const targetPosition = rows.indexOf(row);
  if (targetPosition < 0) return -1;
  const bounds = row.getBoundingClientRect();
  const insertAfter = event.clientY > bounds.top + (bounds.height / 2);
  let destination = targetPosition + (insertAfter ? 1 : 0);
  if (sourcePosition < destination) destination -= 1;
  return Math.max(0, Math.min(destination, rows.length - 1));
}

function technologyDropDestination(list, row, rowKey, event) {
  const rows = [...list.querySelectorAll(":scope > .content-technology-item")];
  const sourcePosition = rows.findIndex((candidate) => candidate.dataset.contentRowKey === rowKey);
  if (sourcePosition < 0) return -1;
  if (!row) return rows.length - 1;
  const targetPosition = rows.indexOf(row);
  if (targetPosition < 0) return -1;
  const bounds = row.getBoundingClientRect();
  const verticalDistance = Math.abs(event.clientY - (bounds.top + (bounds.height / 2)));
  const insertAfter = verticalDistance < bounds.height / 2
    ? event.clientX > bounds.left + (bounds.width / 2)
    : event.clientY > bounds.top + (bounds.height / 2);
  let destination = targetPosition + (insertAfter ? 1 : 0);
  if (sourcePosition < destination) destination -= 1;
  return Math.max(0, Math.min(destination, rows.length - 1));
}

document.addEventListener("dragover", (event) => {
  const technologyList = event.target.closest("[data-technology-sort-list][data-content-key]");
  if (technologyList) {
    if (state.drag?.type !== "content-technology" || state.drag.clientKey !== technologyList.dataset.contentKey) return;
    event.preventDefault();
    event.dataTransfer.dropEffect = "move";
    return;
  }
  const list = event.target.closest("[data-manual-order-list][data-manual-order-collection]");
  if (!list) return;
  const expectedType = `${list.dataset.manualOrderList}-item`;
  if (state.drag?.type !== expectedType || state.drag.collection !== list.dataset.manualOrderCollection) return;
  event.preventDefault();
  event.dataTransfer.dropEffect = "move";
});

document.addEventListener("drop", (event) => {
  const technologyList = event.target.closest("[data-technology-sort-list][data-content-key]");
  if (technologyList) {
    if (state.drag?.type !== "content-technology" || state.drag.clientKey !== technologyList.dataset.contentKey) return;
    event.preventDefault();
    const rowCandidate = event.target.closest(".content-technology-item");
    const row = rowCandidate?.parentElement === technologyList ? rowCandidate : null;
    if (row?.dataset.contentRowKey === state.drag.rowKey) return;
    const destination = technologyDropDestination(technologyList, row, state.drag.rowKey, event);
    if (destination >= 0) {
      moveContentRowToPosition("software", state.drag.clientKey, "technology", state.drag.rowKey, destination);
    }
    return;
  }
  const list = event.target.closest("[data-manual-order-list][data-manual-order-collection]");
  if (!list) return;
  const expectedType = `${list.dataset.manualOrderList}-item`;
  if (state.drag?.type !== expectedType || state.drag.collection !== list.dataset.manualOrderCollection) return;
  event.preventDefault();
  const rowCandidate = event.target.closest(".manual-order-row");
  const row = rowCandidate?.parentElement === list ? rowCandidate : null;
  if (row?.dataset.contentKey === state.drag.clientKey || row?.dataset.entityKey === state.drag.clientKey) return;
  const destination = manualOrderDropDestination(list, row, state.drag.clientKey, event);
  if (destination < 0) return;
  if (state.drag.type === "content-item") {
    moveContentItemToPosition(state.drag.collection, state.drag.clientKey, destination);
  } else {
    moveEntityItemToPosition(state.drag.collection, state.drag.clientKey, destination);
  }
});

document.querySelectorAll("[data-section-zone]").forEach((zoneElement) => {
  zoneElement.addEventListener("dragover", (event) => {
    if (state.drag?.type !== "section") return;
    event.preventDefault();
    event.dataTransfer.dropEffect = "move";
  });
  zoneElement.addEventListener("drop", (event) => {
    if (state.drag?.type !== "section") return;
    event.preventDefault();
    const row = event.target.closest("[data-section]");
    if (row?.dataset.section === state.drag.section) return;
    const beforeSection = row?.dataset.section || null;
    moveSectionByDrop(state.drag.section, state.drag.zone, zoneElement.dataset.sectionZone, beforeSection);
  });
});

const cvHierarchyList = document.querySelector("#cv-hierarchy-list");
cvHierarchyList?.addEventListener("dragover", (event) => {
  if (state.drag?.type !== "cv-section") return;
  event.preventDefault();
  event.dataTransfer.dropEffect = "move";
});
cvHierarchyList?.addEventListener("drop", (event) => {
  if (state.drag?.type !== "cv-section") return;
  event.preventDefault();
  const row = event.target.closest("[data-cv-section]");
  if (row?.dataset.cvSection === state.drag.section) return;
  const destinationZone = row?.dataset.cvZone || state.drag.zone;
  let beforeSection = row?.dataset.cvSection || null;
  if (row) {
    const bounds = row.getBoundingClientRect();
    if (event.clientY > bounds.top + (bounds.height / 2)) {
      const next = [...cvHierarchyList.querySelectorAll(":scope > [data-cv-section]")]
        .slice([...cvHierarchyList.children].indexOf(row) + 1)
        .find((candidate) => candidate.dataset.cvZone === destinationZone);
      beforeSection = next?.dataset.cvSection || null;
    }
  }
  moveCVSectionByDrop(state.drag.section, state.drag.zone, destinationZone, beforeSection);
});

document.querySelectorAll("[data-taxonomy-zone]").forEach((list) => {
  list.addEventListener("dragover", (event) => {
    if (state.drag?.type !== "taxonomy" || state.drag.kind !== list.dataset.taxonomyZone) return;
    event.preventDefault();
    event.dataTransfer.dropEffect = "move";
  });
  list.addEventListener("drop", (event) => {
    if (state.drag?.type !== "taxonomy" || state.drag.kind !== list.dataset.taxonomyZone) return;
    event.preventDefault();
    const row = event.target.closest("[data-kind][data-client-key]");
    if (row?.dataset.clientKey === state.drag.clientKey) return;
    const beforeClientKey = row?.dataset.clientKey || null;
    moveTaxonomyByDrop(state.drag.kind, state.drag.clientKey, beforeClientKey);
  });
});

async function loadEditorData() {
  cancelSaveCompletionStatus();
  state.loading = true;
  state.saving = false;
  loadError.hidden = true;
  updateToolbar();
  try {
    state.bridge ||= await waitForBridge();
    loadEditorUpdateInfo();
    installOperationProgress();
    const response = await state.bridge.LoadEditorData();
    state.baseline = addClientKeys(response.settings);
    state.draft = clone(state.baseline);
    state.revision = response.settings_revision;
    state.usage = response.usage;
    state.profileBaseline = hydrateProfile(response.profile, response.people);
    state.profileDraft = clone(state.profileBaseline);
    state.profileMediaBaseline = hydrateProfileMedia(response.profile_media);
    state.profileMediaDraft = clone(state.profileMediaBaseline);
    state.profileRevision = response.profile_revision || "";
    state.boardBaseline = hydrateBoard(response.board);
    state.boardDraft = clone(state.boardBaseline);
    state.boardRevision = response.board_revision;
    state.selectedBoardKey = sortedBoardItems(state.boardDraft)[0]?._clientKey || "";
    state.projectsBaseline = hydrateProjects(response.projects);
    state.projectsDraft = clone(state.projectsBaseline);
    state.projectsRevision = response.projects_revision || "";
    state.selectedProjectKey = state.projectsDraft[0]?._clientKey || "";
    state.softwareBaseline = hydrateSoftware(response.software);
    state.softwareDraft = clone(state.softwareBaseline);
    state.softwareRevision = response.software_revision || "";
    state.selectedSoftwareKey = state.softwareDraft[0]?._clientKey || "";
    state.peopleBaseline = hydratePeople(response.people);
    state.peopleDraft = clone(state.peopleBaseline);
    state.peopleRevision = response.people_revision || "";
    state.selectedPersonKey = defaultPersonSelection(state.peopleDraft);
    state.awardsBaseline = hydrateAwards(response.awards);
    state.awardsDraft = clone(state.awardsBaseline);
    state.awardsRevision = response.awards_revision || "";
    state.selectedAwardKey = sortedAwards(state.awardsDraft)[0]?._clientKey || "";
    state.academicActivitiesBaseline = hydrateAcademicActivities(response.academic_activities);
    state.academicActivitiesDraft = clone(state.academicActivitiesBaseline);
    state.academicActivitiesRevision = response.academic_activities_revision || "";
    state.selectedAcademicActivityKey = orderedAcademicActivities(state.academicActivitiesDraft)[0]?._clientKey || "";
    state.publicationsBaseline = hydratePublications(response.publications, state.peopleBaseline, state.awardsBaseline);
    state.publicationsDraft = clone(state.publicationsBaseline);
    state.publicationsRevision = response.publications_revision || "";
    state.selectedPublicationKey = sortedPublications(state.publicationsDraft)[0]?._clientKey || "";
    state.boardDropMessage = "";
    state.profileDropMessage = "";
    state.contentDropMessages = {};
    state.dirty = false;
    state.loading = false;
    installFileDrop();
    setStatus("불러옴", "");
    renderAll();
    // Loading can discover a main-page section with no items. Keep that
    // automatic move to hidden as an explicit unsaved edit, so the user can
    // review it and apply it with the normal Save button.
    updateDirtyState();
  } catch (error) {
    state.loading = false;
    state.draft = null;
    state.profileDraft = null;
    state.profileMediaBaseline = null;
    state.profileMediaDraft = null;
    state.boardDraft = null;
    state.projectsDraft = null;
    state.softwareDraft = null;
    state.peopleDraft = null;
    state.awardsDraft = null;
    state.academicActivitiesDraft = null;
    state.publicationsDraft = null;
    loadError.hidden = false;
    loadErrorMessage.textContent = errorMessage(error);
    setStatus("불러오기 실패", "error");
    renderAppearanceSettings();
    updateToolbar();
  }
}

async function saveChanges(options = {}) {
  const withinOperation = options?.forPush === true || options?.withinOperation === true;
  if (
    !state.dirty
    || state.loading
    || (state.saving && !withinOperation)
    || state.discarding
    || state.boardDropBusy
  ) return false;
  cancelSaveCompletionStatus();
  saveError.hidden = true;
  validateDraft();
  if (state.validationErrors.size) {
    revealFirstValidationError();
    return false;
  }
  if (!withinOperation) beginOperation("save");
  state.saving = true;
  renderAll();
  updateToolbar();
  try {
    await state.stagingOps;
    await state.dirtySync;
    validateDraft();
    if (state.validationErrors.size) {
      state.saving = withinOperation;
      revealFirstValidationError();
      if (!withinOperation) finishOperation(false, "입력 오류를 확인해 주세요");
      return false;
    }
    assignPendingRelationshipIDs();
    const settingsChanged = comparable(state.draft) !== comparable(state.baseline);
    const profileChanged = profileComparable(state.profileDraft, state.profileMediaDraft)
      !== profileComparable(state.profileBaseline, state.profileMediaBaseline);
    const boardChanged = boardComparable(state.boardDraft) !== boardComparable(state.boardBaseline);
    const projectsChanged = projectsComparable(state.projectsDraft) !== projectsComparable(state.projectsBaseline);
    const softwareChanged = softwareComparable(state.softwareDraft) !== softwareComparable(state.softwareBaseline);
    const peopleChanged = peopleComparable(state.peopleDraft) !== peopleComparable(state.peopleBaseline);
    const awardsChanged = awardsComparable(state.awardsDraft) !== awardsComparable(state.awardsBaseline);
    const academicActivitiesChanged = academicActivitiesComparable(state.academicActivitiesDraft)
      !== academicActivitiesComparable(state.academicActivitiesBaseline);
    const publicationsChanged = publicationsComparable(state.publicationsDraft) !== publicationsComparable(state.publicationsBaseline);
    const request = {
      operation_id: state.operation.id,
      settings: toSettingsPayload(state.draft),
      settings_revision: state.revision,
      save_settings: settingsChanged,
      profile: toProfilePayload(state.profileDraft),
      profile_revision: state.profileRevision,
      save_profile: profileChanged,
      profile_media_stage_token: state.profileMediaDraft?.stage_token || "",
      remove_profile_media: state.profileMediaDraft?.remove === true,
      document_stage_tokens: documentStageTokens(),
      board: toBoardSavePayload(state.boardDraft, state.boardBaseline),
      board_revision: state.boardRevision,
      save_board: boardChanged,
      projects: toProjectsSavePayload(state.projectsDraft, state.projectsBaseline),
      projects_revision: state.projectsRevision,
      save_projects: projectsChanged,
      software: toSoftwareSavePayload(state.softwareDraft, state.softwareBaseline),
      software_revision: state.softwareRevision,
      save_software: softwareChanged,
      people: toEntitySavePayload(state.peopleDraft, state.peopleBaseline, "person", toPersonPayload),
      people_revision: state.peopleRevision,
      save_people: peopleChanged,
      awards: toEntitySavePayload(state.awardsDraft, state.awardsBaseline, "award", toAwardPayload, sortedAwards),
      awards_revision: state.awardsRevision,
      save_awards: awardsChanged,
      academic_activities: toAcademicActivitiesSavePayload(
        state.academicActivitiesDraft,
        state.academicActivitiesBaseline,
      ),
      academic_activities_revision: state.academicActivitiesRevision,
      save_academic_activities: academicActivitiesChanged,
      publications: toPublicationSavePayload(state.publicationsDraft, state.publicationsBaseline),
      publications_revision: state.publicationsRevision,
      save_publications: publicationsChanged,
    };
    const response = await state.bridge.SaveEditorData(request);
    state.documentUploads = {};
    state.baseline = addClientKeys(response.settings);
    state.draft = clone(state.baseline);
    state.revision = response.settings_revision;
    state.usage = response.usage;
    state.profileBaseline = hydrateProfile(response.profile, response.people);
    state.profileDraft = clone(state.profileBaseline);
    state.profileMediaBaseline = hydrateProfileMedia(response.profile_media);
    state.profileMediaDraft = clone(state.profileMediaBaseline);
    state.profileRevision = response.profile_revision || "";
    state.boardBaseline = hydrateBoard(response.board);
    state.boardDraft = clone(state.boardBaseline);
    state.boardRevision = response.board_revision;
    state.selectedBoardKey = sortedBoardItems(state.boardDraft)[0]?._clientKey || "";
    state.projectsBaseline = hydrateProjects(response.projects);
    state.projectsDraft = clone(state.projectsBaseline);
    state.projectsRevision = response.projects_revision || "";
    state.selectedProjectKey = state.projectsDraft[0]?._clientKey || "";
    state.softwareBaseline = hydrateSoftware(response.software);
    state.softwareDraft = clone(state.softwareBaseline);
    state.softwareRevision = response.software_revision || "";
    state.selectedSoftwareKey = state.softwareDraft[0]?._clientKey || "";
    state.peopleBaseline = hydratePeople(response.people);
    state.peopleDraft = clone(state.peopleBaseline);
    state.peopleRevision = response.people_revision || "";
    state.selectedPersonKey = defaultPersonSelection(state.peopleDraft);
    state.awardsBaseline = hydrateAwards(response.awards);
    state.awardsDraft = clone(state.awardsBaseline);
    state.awardsRevision = response.awards_revision || "";
    state.selectedAwardKey = sortedAwards(state.awardsDraft)[0]?._clientKey || "";
    state.academicActivitiesBaseline = hydrateAcademicActivities(response.academic_activities);
    state.academicActivitiesDraft = clone(state.academicActivitiesBaseline);
    state.academicActivitiesRevision = response.academic_activities_revision || "";
    state.selectedAcademicActivityKey = orderedAcademicActivities(state.academicActivitiesDraft)[0]?._clientKey || "";
    state.publicationsBaseline = hydratePublications(response.publications, state.peopleBaseline, state.awardsBaseline);
    state.publicationsDraft = clone(state.publicationsBaseline);
    state.publicationsRevision = response.publications_revision || "";
    state.selectedPublicationKey = sortedPublications(state.publicationsDraft)[0]?._clientKey || "";
    state.boardDropMessage = "";
    state.profileDropMessage = "";
    state.contentDropMessages = {};
    state.dirty = false;
    state.saving = withinOperation;
    renderAll();
    if (!withinOperation) {
      showSaveCompletionStatus();
      finishOperation(true, "저장 완료");
    }
    return true;
  } catch (error) {
    state.saving = withinOperation;
    const message = errorMessage(error);
    setStatus("저장 실패", "error");
    saveErrorMessage.textContent = message;
    saveError.hidden = false;
    renderAll();
    syncNativeDirty(true);
    if (!withinOperation) finishOperation(false, "저장 실패", message);
    return false;
  }
}

async function updateCitations() {
  if (!state.draft || !state.bridge?.UpdateCitations || state.loading || state.saving || state.discarding || state.boardDropBusy) return;
  cancelSaveCompletionStatus();
  beginOperation("citations");
  state.saving = true;
  renderAll();
  try {
    if (state.dirty && !await saveChanges({ withinOperation: true })) {
      finishOperation(false, "저장 실패", state.validationErrors.size ? "입력 오류를 확인해 주세요" : saveErrorMessage.textContent);
      return;
    }
    await state.stagingOps;
    await state.dirtySync;
    updateOperationProgress("인용수 조회 중…", -1);
    const response = await state.bridge.UpdateCitations({ operation_id: state.operation.id });
    if (typeof response.output === "string") updateOperationOutput(response.output);
    const success = response.success === true;
    finishOperation(success, response.message || (success ? "인용수 갱신 완료" : "인용수 갱신 실패"));
  } catch (error) {
    finishOperation(false, "인용수 갱신 실패", errorMessage(error));
  } finally {
    state.saving = false;
    renderAll();
  }
}

async function pushChanges() {
  if (!state.draft || state.loading || state.saving || state.discarding || state.boardDropBusy) return;
  cancelSaveCompletionStatus();
  beginOperation("push");
  state.saving = true;
  renderAll();
  updateToolbar();
  try {
    if (state.dirty && !await saveChanges({ forPush: true })) {
      finishOperation(false, "저장 실패", state.validationErrors.size ? "입력 오류를 확인해 주세요" : saveErrorMessage.textContent);
      return;
    }
    await state.stagingOps;
    await state.dirtySync;
    updateOperationProgress("저장소 확인 중…", -1);
    const response = await state.bridge.PushGit({ operation_id: state.operation.id });
    finishOperation(true, response.message || "Git push 완료");
    setStatus("Git push 완료", "success");
  } catch (error) {
    finishOperation(false, "Git push 실패", errorMessage(error));
    setStatus("Git push 실패", "error");
  } finally {
    state.saving = false;
    renderAll();
  }
}

async function discardChanges() {
  if (!state.dirty || state.saving || state.discarding || state.boardDropBusy) return;
  if (!window.confirm("저장하지 않은 변경을 모두 취소할까요?")) return;
  cancelSaveCompletionStatus();
  state.discarding = true;
  renderAll();
  updateToolbar();
  try {
    await state.stagingOps.catch(() => {});
    const documentTokens = Object.values(state.documentUploads).map((document) => document.stage_token);
    if (documentTokens.length) await state.bridge.DiscardDocuments(documentTokens);
    state.documentUploads = {};
    const stageTokens = [state.boardDraft, state.projectsDraft, state.softwareDraft]
      .flat()
      .flatMap((item) => item.media.map((media) => media.stage_token))
      .filter(Boolean);
    const profileStageToken = state.profileMediaDraft?.stage_token;
    if (profileStageToken) stageTokens.push(profileStageToken);
    if (stageTokens.length) {
      await state.bridge.DiscardBoardMedia(stageTokens);
    }
    state.draft = clone(state.baseline);
    state.profileDraft = clone(state.profileBaseline);
    state.profileMediaDraft = clone(state.profileMediaBaseline);
    state.boardDraft = clone(state.boardBaseline);
    state.selectedBoardKey = sortedBoardItems(state.boardDraft)[0]?._clientKey || "";
    state.projectsDraft = clone(state.projectsBaseline);
    state.selectedProjectKey = state.projectsDraft[0]?._clientKey || "";
    state.softwareDraft = clone(state.softwareBaseline);
    state.selectedSoftwareKey = state.softwareDraft[0]?._clientKey || "";
    state.peopleDraft = clone(state.peopleBaseline);
    state.selectedPersonKey = defaultPersonSelection(state.peopleDraft);
    state.awardsDraft = clone(state.awardsBaseline);
    state.selectedAwardKey = sortedAwards(state.awardsDraft)[0]?._clientKey || "";
    state.academicActivitiesDraft = clone(state.academicActivitiesBaseline);
    state.selectedAcademicActivityKey = orderedAcademicActivities(state.academicActivitiesDraft)[0]?._clientKey || "";
    state.publicationsDraft = clone(state.publicationsBaseline);
    state.selectedPublicationKey = sortedPublications(state.publicationsDraft)[0]?._clientKey || "";
    state.boardDropMessage = "";
    state.profileDropMessage = "";
    state.contentDropMessages = {};
    state.dirty = false;
    saveError.hidden = true;
    setStatus("변경 취소됨", "");
    syncNativeDirty(false);
  } catch (error) {
    const message = errorMessage(error);
    setStatus("변경 취소 실패", "error");
    saveErrorMessage.textContent = message;
    saveError.hidden = false;
  } finally {
    state.discarding = false;
    renderAll();
  }
}

saveButton.addEventListener("click", saveChanges);
document.querySelector("#editor-update-check")?.addEventListener("click", checkEditorUpdates);
document.querySelector("#editor-update-install")?.addEventListener("click", startEditorUpdate);
citationUpdateButton.addEventListener("click", updateCitations);
pushButton.addEventListener("click", pushChanges);
operationClose.addEventListener("click", closeOperation);
editorUpdateDialog?.addEventListener("cancel", (event) => {
  event.preventDefault();
  closeOperation();
});
document.addEventListener("keydown", handleEditorShortcutKeydown, true);
document.addEventListener("keydown", (event) => {
  if (event.key === "Escape" && !operationPopover.hidden && !state.saving) closeOperation();
});
discardButton.addEventListener("click", discardChanges);
retryButton.addEventListener("click", loadEditorData);
const mediaPreviewDialog = document.querySelector("#media-preview-dialog");
mediaPreviewDialog?.querySelector("#media-preview-close").addEventListener("click", closeMediaPreview);
mediaPreviewDialog?.addEventListener("cancel", (event) => {
  event.preventDefault();
  closeMediaPreview();
});
mediaPreviewDialog?.addEventListener("close", () => {
  if (mediaPreviewDialog.open) return;
  const content = mediaPreviewDialog.querySelector("#media-preview-content");
  const element = content.firstElementChild;
  content.replaceChildren();
  if (element) {
    if (element.tagName === "VIDEO") element.pause();
    element.removeAttribute("src");
    if (element.tagName === "VIDEO") element.load();
  }
  content.setAttribute("aria-busy", "false");
  const status = mediaPreviewDialog.querySelector("#media-preview-status");
  status.textContent = "";
  status.hidden = true;
  const trigger = mediaPreviewReturnFocus;
  mediaPreviewReturnFocus = null;
  if (trigger?.isConnected) trigger.focus({ preventScroll: true });
});
document.addEventListener("keydown", (event) => {
  if (mediaPreviewDialog?.open) {
    if ((event.ctrlKey || event.metaKey) && event.key.toLocaleLowerCase() === "s") event.preventDefault();
    return;
  }
  const technologyHandle = event.target.closest?.('[data-action="drag-technology"]');
  const technologyDirection = {
    ArrowUp: -1,
    ArrowLeft: -1,
    ArrowDown: 1,
    ArrowRight: 1,
  }[event.key];
  if (technologyHandle && event.altKey && !event.ctrlKey && !event.metaKey && technologyDirection) {
    const row = technologyHandle.closest(".content-technology-item[data-content-row-key]");
    const list = row?.closest("[data-technology-sort-list][data-content-key]");
    const rows = list ? [...list.querySelectorAll(":scope > .content-technology-item")] : [];
    const index = rows.indexOf(row);
    if (row && list && index >= 0) {
      event.preventDefault();
      const moved = moveContentRowToPosition(
        "software",
        list.dataset.contentKey,
        "technology",
        row.dataset.contentRowKey,
        index + technologyDirection,
      );
      if (moved) {
        focusTechnologyDragHandle(row.dataset.contentRowKey);
        announce(`기술 순서를 ${technologyDirection < 0 ? "앞으로" : "뒤로"} 이동했습니다.`);
      }
    }
    return;
  }
});

activateTab(tabs[0], false);
renderAppearanceSettings();
renderEditorUpdate();
loadEditorData();
