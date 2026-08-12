
(function () {
  "use strict";

  const D = window.SiteData;
  const citations = window.PublicationCitations.createClient();
  let citationOrderPending = false;
  const PAGE_VIEWS = new Set(["publications", "projects", "software", "board"]);
  const PAGE_SEQUENCE = ["home", "publications", "projects", "software", "board"];
  const SWIPE_MIN_DISTANCE = 64;
  const SWIPE_DIRECTION_RATIO = 1.35;
  const PROJECT_TIMELINE_YEAR_WIDTH = 140;
  const PROJECT_TIMELINE_LANE_HEIGHT = 30;
  const BOARD_PAGE_SIZE = 5;
  const PROJECT_TIMELINE_COLORS = ["#315f49", "#3d6f86", "#80643b", "#775778", "#4f6e31", "#8a5148"];
  const PUBLICATION_TYPES = ["international-journal", "domestic-journal", "international-conference", "domestic-conference"];
  const SOFTWARE_STAGES = ["release", "preview", "development"];
  const SOFTWARE_STAGE_LABELS = Object.freeze({
    release: { en: "Release", ko: "공개 배포 단계" },
    preview: { en: "Preview", ko: "시험 공개 단계" },
    development: { en: "Development", ko: "개발 단계" }
  });
  const MAIN_PAGE_SECTIONS = ["experience", "education", "scholarships", "certifications", "awards", "academic_activities", "patents", "teaching", "skills"];
  const ALL_TOPIC_FILTER = "__all_topics__";
  const FALLBACK_TOPIC_FILTER = "__fallback_topic__";
  const LINK_ICONS = Object.freeze({
    github: { icon: "app/assets/icons/links/github.svg", monochrome: true },
    food4rhino: { icon: "app/assets/icons/links/food4rhino.svg" },
    repository: { icon: "app/assets/icons/links/repository.svg", monochrome: true },
    package: { icon: "app/assets/icons/links/package.svg", monochrome: true },
    container: { icon: "app/assets/icons/links/container.svg", monochrome: true },
    archive: { icon: "app/assets/icons/links/archive.svg", monochrome: true },
    webpage: { icon: "app/assets/icons/links/webpage.svg", monochrome: true },
    email: { icon: "app/assets/icons/links/email.svg", monochrome: true },
    orcid: { icon: "app/assets/icons/links/orcid.svg" },
    scholar: { icon: "app/assets/icons/links/scholar.svg" },
    linkedin: { icon: "app/assets/icons/links/linkedin.svg" },
    address: { icon: "app/assets/icons/links/address.svg", monochrome: true }
  });
  const TECHNOLOGY_ICONS = Object.freeze({
    python: { label: "Python", icon: "app/assets/icons/technologies/python.svg" },
    go: { label: "Go", icon: "app/assets/icons/technologies/go.svg" },
    rust: { label: "Rust", icon: "app/assets/icons/technologies/rust.svg" },
    javascript: { label: "JavaScript", icon: "app/assets/icons/technologies/javascript.svg" },
    typescript: { label: "TypeScript", icon: "app/assets/icons/technologies/typescript.svg" },
    java: { label: "Java", icon: "app/assets/icons/technologies/java.svg" },
    c: { label: "C", icon: "app/assets/icons/technologies/c.svg" },
    "c++": { label: "C++", icon: "app/assets/icons/technologies/cplusplus.svg" },
    "c#": { label: "C#", icon: "app/assets/icons/technologies/csharp.svg" },
    r: { label: "R", icon: "app/assets/icons/technologies/r.svg" },
    matlab: { label: "MATLAB", icon: "app/assets/icons/technologies/matlab.svg" },
    html: { label: "HTML", icon: "app/assets/icons/technologies/html.svg" },
    css: { label: "CSS", icon: "app/assets/icons/technologies/css.svg" },
    react: { label: "React", icon: "app/assets/icons/technologies/react.svg" },
    vue: { label: "Vue", icon: "app/assets/icons/technologies/vue.svg" },
    svelte: { label: "Svelte", icon: "app/assets/icons/technologies/svelte.svg" },
    rhino: { label: "Rhino", icon: "app/assets/icons/technologies/rhino.svg" },
    grasshopper: { label: "Grasshopper", icon: "app/assets/icons/technologies/grasshopper.svg" },
    wails: { label: "Wails", icon: "app/assets/icons/technologies/wails.svg", className: "tech-icon-wails" },
    egui: { label: "egui", icon: "app/assets/icons/technologies/egui.svg", className: "tech-icon-egui" },
    energyplus: { label: "EnergyPlus", icon: "app/assets/icons/technologies/energyplus.svg" },
    excel: { label: "Excel", icon: "app/assets/icons/technologies/excel.svg", className: "tech-icon-excel" },
    word: { label: "Microsoft Word", icon: "app/assets/icons/technologies/word.svg" },
    pdf: { label: "PDF", icon: "app/assets/icons/technologies/pdf.svg" },
    latex: { label: "LaTeX", icon: "app/assets/icons/technologies/latex.svg" }
  });
  const TECHNOLOGY_ALIASES = Object.freeze({
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
    energyplus: "energyplus",
    eplus: "energyplus",
    tex: "latex",
    "microsoft excel": "excel",
    "microsoft word": "word",
    docx: "word",
    acrobat: "pdf",
    "adobe acrobat": "pdf"
  });
  const state = {
    lang: "ko",
    data: null,
    publication: {
      query: "",
      publicationType: "all",
      indexing: "all",
      topic: ALL_TOPIC_FILTER,
      collaboratorId: "",
      firstAuthorOnly: false,
      sort: "newest",
      expanded: new Set()
    },
    award: {
      expanded: new Set()
    },
    scholarshipRenderVersion: 0,
    board: {
      page: 0,
      activeItemIndex: null
    },
    pageTransitionDirection: 0,
    lightbox: {
      collection: "board",
      itemIndex: 0,
      index: 0
    }
  };

  const I18N = {
    en: {
      skip: "Skip to content",
      home: "Home",
      navProjects: "Projects",
      navSoftware: "Software",
      navPublications: "Publications",
      navBoard: "Board",
      siteReuseNote: "This GitHub page was built for anyone to copy, modify, and use.",
      siteReuseGuide: "View instructions",
      boardList: "Board post list",
      boardNavigation: "Board quick navigation",
      boardDate: "Date",
      boardTitle: "Title",
      boardPagination: "Board list pages",
      previousPage: "Previous",
      nextPage: "Next",
      projectTimeline: "Project timeline",
      publicationTimeline: "Publication timeline",
      publicationDatePending: "under review",
      publicationVenuePending: "Venue pending",
      onlinePublicationDate: "Published online",
      issuePublicationDate: "Issue date",
      publicationPublisher: "Publisher",
      proceedingsPublisher: "Publisher",
      publicationConference: "Conference",
      coAuthor: "Co-author",
      searchLabel: "Search",
      searchPlaceholder: "Search: title, author, journal, conference…",
      typeLabel: "Publication type",
      allTypes: "All types",
      indexLabel: "Index",
      allIndexes: "All indexes",
      otherIndexes: "Other",
      topicLabel: "Publication topic",
      allTopics: "All topics",
      resetFilters: "Reset filters",
      collaborators: "Collaborators",
      sortNewest: "Newest first",
      sortCited: "Most cited",
      sortType: "By type",
      sortTopic: "By topic",
      publicationCount: "papers",
      firstAuthorOnly: "First author",
      results: "results",
      noPublications: "No publications match the current filters.",
      experienceTitle: "Experience",
      educationTitle: "Education",
      scholarshipsTitle: "Scholarships",
      scholarshipAmount: "Amount",
      certificationsTitle: "Certifications",
      awardsTitle: "Awards",
      patentsTitle: "Patents",
      academicActivitiesTitle: "Academic Activities",
      academicReviewTitle: "Review",
      academicInvitedTalkTitle: "Invited Talk",
      academicConferenceServiceTitle: "Conference Service",
      academicProfessionalServiceTitle: "Professional Service",
      academicEditorialServiceTitle: "Editorial Service",
      present: "Present",
      teachingTitle: "Teaching",
      skillsTitle: "Technical Skills",
      award: "Awarded",
      source: "Source",
      copy: "Copy",
      bib: "Bib",
      repository: "Repository ↗",
      technologies: "Technologies",
      copied: "Citation copied.",
      copyFailed: "Could not copy automatically.",
      loadError: "The page could not be loaded. Please refresh and try again.",
      loadErrorReason: "Reason",
      mediaGallery: "Media gallery",
      previousMedia: "Previous media",
      nextMedia: "Next media",
      expandMedia: "Open full-screen media",
      closeMedia: "Close media viewer",
      email: "Email",
      address: "Address"
    },
    ko: {
      skip: "본문으로 건너뛰기",
      home: "홈으로",
      navProjects: "프로젝트",
      navSoftware: "소프트웨어",
      navPublications: "논문",
      navBoard: "게시판",
      siteReuseNote: "이 GitHub page는 누구나 복사/수정하여 사용 가능하게 제작되었습니다.",
      siteReuseGuide: "사용법 보기",
      boardList: "게시물 목록",
      boardNavigation: "게시물 빠른 이동",
      boardDate: "날짜",
      boardTitle: "제목",
      boardPagination: "게시물 목록 페이지",
      previousPage: "이전",
      nextPage: "다음",
      projectTimeline: "프로젝트 기간",
      publicationTimeline: "논문 출판 시점",
      publicationDatePending: "심사중",
      publicationVenuePending: "게재지 미정",
      onlinePublicationDate: "온라인 게재",
      issuePublicationDate: "권호 발행",
      publicationPublisher: "출판사",
      proceedingsPublisher: "출판사",
      publicationConference: "학술대회",
      coAuthor: "공저자",
      searchLabel: "검색",
      searchPlaceholder: "검색: 제목, 저자, 학술지, 학술대회…",
      typeLabel: "유형",
      allTypes: "전체 유형",
      indexLabel: "색인",
      allIndexes: "전체 색인",
      otherIndexes: "기타",
      topicLabel: "주제",
      allTopics: "전체 주제",
      resetFilters: "필터 초기화",
      collaborators: "공저자",
      sortNewest: "최신순",
      sortCited: "인용순",
      sortType: "유형별",
      sortTopic: "주제별",
      publicationCount: "건",
      firstAuthorOnly: "주저자",
      results: "건",
      noPublications: "현재 필터와 일치하는 논문이 없습니다.",
      experienceTitle: "경력",
      educationTitle: "학력",
      scholarshipsTitle: "장학",
      scholarshipAmount: "금액",
      certificationsTitle: "자격사항",
      awardsTitle: "수상",
      patentsTitle: "특허",
      academicActivitiesTitle: "학술활동",
      academicReviewTitle: "리뷰",
      academicInvitedTalkTitle: "초청 발표",
      academicConferenceServiceTitle: "학술대회 활동",
      academicProfessionalServiceTitle: "전문 활동",
      academicEditorialServiceTitle: "편집 활동",
      present: "현재",
      teachingTitle: "교육",
      skillsTitle: "기술 역량",
      award: "수상",
      source: "원문",
      copy: "인용 복사",
      bib: "Bib",
      repository: "저장소 ↗",
      technologies: "기술",
      copied: "인용정보를 복사했습니다.",
      copyFailed: "자동 복사에 실패했습니다.",
      loadError: "페이지를 불러오지 못했습니다. 새로고침 후 다시 시도해 주세요.",
      loadErrorReason: "원인",
      mediaGallery: "미디어 모음",
      previousMedia: "이전 미디어",
      nextMedia: "다음 미디어",
      expandMedia: "미디어 크게 보기",
      closeMedia: "미디어 보기 닫기",
      email: "이메일",
      address: "주소"
    }
  };

  const $ = (selector, root = document) => root.querySelector(selector);
  const $$ = (selector, root = document) => Array.from(root.querySelectorAll(selector));
  const t = (key) => I18N[state.lang]?.[key] || I18N.en[key] || key;
  const p = (record, key) => D.pick(record, key, state.lang);

  function initialLanguage() {
    const query = new URLSearchParams(window.location.search).get("lang");
    if (query === "en" || query === "ko") return query;
    const saved = window.localStorage.getItem("siteLanguage");
    if (saved === "en" || saved === "ko") return saved;
    return "ko";
  }

  function alignMobileBrand() {
    const header = $(".site-header .header-inner");
    const firstLink = $(".primary-nav a:not([hidden])");
    if (!header) return;
    if (!firstLink || !window.matchMedia("(max-width: 620px)").matches) {
      header.style.removeProperty("--mobile-brand-offset");
      return;
    }
    const offset = firstLink.getBoundingClientRect().left - header.getBoundingClientRect().left;
    header.style.setProperty("--mobile-brand-offset", `${Math.max(0, offset)}px`);
  }

  function setLanguage(language, options = {}) {
    state.lang = language === "ko" ? "ko" : "en";
    document.documentElement.lang = state.lang;
    document.documentElement.dataset.lang = state.lang;
    window.localStorage.setItem("siteLanguage", state.lang);

    if (!options.skipURL) {
      const url = new URL(window.location.href);
      url.searchParams.set("lang", state.lang);
      window.history.replaceState({}, "", url);
    }

    $$("[data-lang-switch]").forEach((button) => {
      const active = button.dataset.langSwitch === state.lang;
      button.setAttribute("aria-pressed", String(active));
    });

    $$("[data-i18n]").forEach((element) => {
      element.textContent = t(element.dataset.i18n);
    });

    $$("[data-i18n-placeholder]").forEach((element) => {
      element.setAttribute("placeholder", t(element.dataset.i18nPlaceholder));
    });

    $$("[data-i18n-aria]").forEach((element) => {
      element.setAttribute("aria-label", t(element.dataset.i18nAria));
    });

    $$("[data-i18n-title]").forEach((element) => {
      element.setAttribute("title", t(element.dataset.i18nTitle));
    });

    const guideLink = $(".site-reuse-note a");
    if (guideLink) guideLink.href = `fork-guide.html?lang=${state.lang}`;

    if (state.data) {
      const identity = state.data.profile.identity;
      document.title = D.personName(identity, state.lang);
      renderAll();
    }
    alignMobileBrand();
  }

  function updateProfileBindings() {
    const profile = state.data.profile;
    const identity = profile.identity;
    const intro = profile.intro;
    const profileCard = profile.profile_card || {};
    const values = {
      name: D.personName(identity, state.lang),
      affiliation: identity[`affiliation_${state.lang}`] || identity.affiliation_en,
      "intro-short": intro[`short_${state.lang}`] || intro.short_en
    };

    $$("[data-profile]").forEach((element) => {
      element.textContent = values[element.dataset.profile] || "";
    });
    const profileMedia = profileCard.media;
    const profileMediaType = D.contentMediaType(profileMedia);
    const profileMediaSource = D.contentMediaPath(profileMedia);
    const profilePhoto = $("#profile-photo");
    const profileVideo = $("#profile-video");
    if (!profileMediaSource) {
      profilePhoto.hidden = true;
      profilePhoto.removeAttribute("src");
      profileVideo.pause();
      profileVideo.hidden = true;
      profileVideo.removeAttribute("src");
      profileVideo.removeAttribute("poster");
    } else if (profileMediaType === "video") {
      profilePhoto.hidden = true;
      profilePhoto.removeAttribute("src");
      profileVideo.hidden = false;
      profileVideo.src = profileMediaSource;
      const poster = typeof profileMedia === "object" ? D.contentMediaPath(profileMedia.poster) : "";
      if (poster) profileVideo.poster = poster;
      else profileVideo.removeAttribute("poster");
    } else {
      profileVideo.pause();
      profileVideo.hidden = true;
      profileVideo.removeAttribute("src");
      profileVideo.removeAttribute("poster");
      profilePhoto.hidden = false;
      profilePhoto.src = profileMediaSource;
    }
    $("#profile-card-title").textContent = profileCard[`title_${state.lang}`] || profileCard.title_en || "";
    $("#profile-card-credentials").innerHTML = (profileCard.credentials || []).map((item) => `
      <li>${D.escapeHTML(item[`name_${state.lang}`] || item.name_en || "")}</li>
    `).join("");
  }

  function renderProjects() {
    if (!pageIsVisible("projects")) return;
    const items = state.data.projects;

    renderProjectTimeline(items);

    $("#projects-grid").innerHTML = items.map((item, itemIndex) => {
      const mediaItems = itemMedia(item);
      const notes = D.projectNotes(item, state.lang);
      return `
        <article class="project-card" data-project-card-index="${itemIndex}" tabindex="-1">
          <div class="project-card-meta">
            <div class="project-topline">${D.escapeHTML(D.projectPeriod(item))} ${D.escapeHTML(p(item, "funder"))}</div>
          </div>
          <h3>${D.escapeHTML(p(item, "title"))}</h3>
          <ul class="project-notes">
            ${notes.map((note) => `<li>${D.escapeHTML(note)}</li>`).join("")}
          </ul>
          ${mediaItems.length ? renderMediaGallery(item, mediaItems, "projects", "project-gallery", itemIndex) : ""}
        </article>
      `;
    }).join("");
  }

  function projectDateValue(value) {
    if (!/^\d{4}-\d{2}-\d{2}$/.test(String(value || ""))) return NaN;
    return Date.parse(`${value}T00:00:00Z`);
  }

  function projectTimelineGroups(records) {
    const configuredThemes = state.data.settings.topics || [];
    const configuredIds = new Set(configuredThemes.map((theme) => theme.id));
    const groups = configuredThemes.map((theme, index) => ({
      id: theme.id,
      label: D.pick(theme, "label", state.lang),
      color: D.topicColor(theme, index),
      items: records.filter((record) => record.item.theme === theme.id)
    })).filter((group) => group.items.length);
    const fallbackItems = records.filter((record) => !configuredIds.has(record.item.theme));
    if (fallbackItems.length) {
      groups.push({
        id: "__fallback__",
        label: D.pick(state.data.settings.topic_fallback, "label", state.lang),
        items: fallbackItems
      });
    }
    return groups;
  }

  function assignProjectTimelineLanes(records) {
    const locale = state.lang === "ko" ? "ko-KR" : "en";
    const sorted = [...records].sort((a, b) => a.start - b.start
      || p(a.item, "funder").localeCompare(p(b.item, "funder"), locale, { sensitivity: "base" })
      || a.index - b.index);
    const lanes = [];
    sorted.forEach((record) => {
      const funder = p(record.item, "funder");
      const available = lanes
        .map((lane, index) => ({ lane, index }))
        .filter(({ lane }) => lane.end < record.start)
        .sort((a, b) => Number(b.lane.funder === funder) - Number(a.lane.funder === funder)
          || a.index - b.index);
      const laneIndex = available.length ? available[0].index : lanes.length;
      lanes[laneIndex] = { end: record.end, funder };
      record.lane = laneIndex;
    });
    return { records: sorted, count: Math.max(1, lanes.length) };
  }

  function projectTimelineThemeWidth(groups) {
    const context = document.createElement("canvas").getContext("2d");
    if (!context) return 190;
    context.font = `800 11px ${window.getComputedStyle(document.body).fontFamily}`;
    const textWidth = Math.max(...groups.map((group) => context.measureText(group.label).width), 0);
    return Math.ceil(textWidth + 48);
  }

  function scrollProjectTimelineToLatest() {
    const scroller = $("[data-project-timeline-scroll]");
    if (!scroller || scroller.dataset.initialPositionSet === "true" || !scroller.clientWidth) return;
    const latestPosition = Math.max(0, scroller.scrollWidth - scroller.clientWidth);
    scroller.scrollLeft = latestPosition;
    if (latestPosition > 0) scroller.dataset.initialPositionSet = "true";
  }

  function renderProjectTimeline(items) {
    hideProjectTimelineTooltip();
    const records = items.map((item, index) => ({
      item,
      index,
      start: projectDateValue(item.start_date),
      end: projectDateValue(item.end_date)
    })).filter((record) => Number.isFinite(record.start) && Number.isFinite(record.end));
    const root = $("#projects-timeline");
    if (!records.length) {
      root.innerHTML = "";
      root.hidden = true;
      return;
    }

    root.hidden = false;
    const firstYear = Math.min(...records.map((record) => new Date(record.start).getUTCFullYear()));
    const lastYear = Math.max(...records.map((record) => new Date(record.end).getUTCFullYear()));
    const axisStart = Date.UTC(firstYear, 0, 1);
    const axisEnd = Date.UTC(lastYear + 1, 0, 1);
    const axisSpan = axisEnd - axisStart;
    const yearCount = lastYear - firstYear + 1;
    const timelineWidth = yearCount * PROJECT_TIMELINE_YEAR_WIDTH;
    const position = (time) => ((time - axisStart) / axisSpan) * timelineWidth;
    const yearTicks = Array.from({ length: yearCount }, (_, offset) => {
      const year = firstYear + offset;
      return `<span class="project-roadmap-year" style="left:${position(Date.UTC(year, 0, 1)).toFixed(3)}px">${year}</span>`;
    }).join("");

    const groups = projectTimelineGroups(records);
    const themeLabelWidth = projectTimelineThemeWidth(groups);
    const rows = groups.map((group, groupIndex) => {
      const layout = assignProjectTimelineLanes(group.items);
      const trackHeight = layout.count * PROJECT_TIMELINE_LANE_HEIGHT + 8;
      const color = group.color || PROJECT_TIMELINE_COLORS[groupIndex % PROJECT_TIMELINE_COLORS.length];
      const bars = layout.records.map((record) => {
        const start = Math.min(record.start, record.end);
        const inclusiveEnd = Math.max(record.start, record.end) + 86400000;
        const left = position(start);
        const width = Math.max(5, position(inclusiveEnd) - left);
        const funder = p(record.item, "funder");
        const title = p(record.item, "title");
        const period = D.projectPeriod(record.item);
        const accessibleLabel = [period, funder, title].join(". ");
        return `
          <button type="button" class="project-roadmap-bar" data-project-timeline-target="${record.index}"
            style="left:${left.toFixed(3)}px;top:${record.lane * PROJECT_TIMELINE_LANE_HEIGHT + 4}px;width:${width.toFixed(3)}px"
            aria-label="${D.escapeHTML(accessibleLabel)}" aria-describedby="project-timeline-tooltip">
            <span class="project-roadmap-bar-funder">${D.escapeHTML(funder)}</span>
          </button>
        `;
      }).join("");
      return `
        <div class="project-roadmap-theme" style="--project-theme-color:${color};--project-theme-ink:${D.topicForegroundColor(color)}">
          <div class="project-roadmap-theme-label"><span></span>${D.escapeHTML(group.label)}</div>
          <div class="project-roadmap-track" style="height:${trackHeight}px">${bars}</div>
        </div>
      `;
    }).join("");

    root.innerHTML = `
      <div class="project-roadmap-scroll" data-project-timeline-scroll tabindex="0" aria-label="${D.escapeHTML(t("projectTimeline"))}">
        <div class="project-roadmap-canvas" style="--project-timeline-width:${timelineWidth}px;--project-theme-label-natural-width:${themeLabelWidth}px">
          <div class="project-roadmap-header">
            <div class="project-roadmap-corner" aria-hidden="true"></div>
            <div class="project-roadmap-axis">${yearTicks}</div>
          </div>
          ${rows}
        </div>
      </div>
    `;
    window.requestAnimationFrame(scrollProjectTimelineToLatest);
  }

  function scrollToProjectCard(index) {
    const card = $(`[data-project-card-index="${index}"]`);
    if (!card) return;
    hideProjectTimelineTooltip();
    $$("[data-project-card-index]").forEach((item) => item.classList.remove("is-timeline-target"));
    card.classList.add("is-timeline-target");
    card.scrollIntoView({ behavior: "smooth", block: "start", inline: "nearest" });
    card.focus({ preventScroll: true });
    window.clearTimeout(scrollToProjectCard.timer);
    scrollToProjectCard.timer = window.setTimeout(() => {
      card.classList.remove("is-timeline-target");
    }, 1800);
  }

  function renderSoftware() {
    if (!pageIsVisible("software")) return;
    const records = state.data.software.map((item, itemIndex) => ({ item, itemIndex }));

    $("#software-grid").innerHTML = SOFTWARE_STAGES.map((stage) => {
      const stageRecords = records.filter(({ item }) => item.stage === stage);
      if (!stageRecords.length) return "";
      const stageLabel = SOFTWARE_STAGE_LABELS[stage]?.[state.lang] || stage;

      return `
        <section class="software-stage-group" data-software-stage="${stage}">
          <header class="software-stage-header">
            <h2>${D.escapeHTML(stageLabel)}</h2>
            <span aria-label="${D.escapeHTML(`${stageLabel}: ${stageRecords.length}`)}">${stageRecords.length}</span>
          </header>
          <div class="card-grid software-grid">
            ${stageRecords.map(({ item, itemIndex }) => {
              const mediaItems = itemMedia(item);
              const links = D.softwareLinks(item, state.lang);
              const notes = D.softwareNotes(item, state.lang);
              return `
                <article class="software-card">
                  <h3>${D.escapeHTML(item.name)}</h3>
                  <div class="software-stack" aria-label="${t("technologies")}">${softwareTechnologiesHTML(item.technologies)}</div>
                  <ul class="software-notes">
                    ${notes.map((note) => `<li>${D.escapeHTML(note)}</li>`).join("")}
                  </ul>
                  ${mediaItems.length ? renderMediaGallery(item, mediaItems, "software", "software-gallery", itemIndex) : ""}
                  ${links.length ? `<div class="software-actions">${softwareLinksHTML(links)}</div>` : ""}
                </article>
              `;
            }).join("")}
          </div>
        </section>
      `;
    }).join("");
  }

  function softwareTechnologiesHTML(value) {
    return D.splitList(value).map((name) => {
      const normalized = name.trim().toLowerCase();
      const key = TECHNOLOGY_ALIASES[normalized] || normalized;
      const preset = TECHNOLOGY_ICONS[key];
      const fallbackLabel = name.trim();
      const label = D.escapeHTML(preset?.label || fallbackLabel);
      const className = preset?.className ? ` ${preset.className}` : "";
      const content = preset
        ? `<span class="software-tech-image"><img src="${D.escapeHTML(preset.icon)}" alt="" aria-hidden="true"></span>`
        : `<span aria-hidden="true">${D.escapeHTML(fallbackLabel.slice(0, 2))}</span>`;
      return `
        <span class="software-tech-icon${className}${preset ? "" : " software-tech-fallback"}" role="img" tabindex="0" aria-label="${label}" data-tech-label="${label}" data-tech-key="${D.escapeHTML(key)}">
          ${content}
        </span>
      `;
    }).join("");
  }

  function softwareLinksHTML(links) {
    return links.map((link) => `
      <a class="software-link" data-link-platform="${D.escapeHTML(link.platform)}"
        href="${D.escapeHTML(link.url)}" target="_blank" rel="noreferrer"
        aria-label="${D.escapeHTML(`${link.label}: ${link.url}`)}"
        data-link-url="${D.escapeHTML(link.url)}">
        ${linkIconHTML(link.icon, "software-link-icon")}
        <span>${D.escapeHTML(link.label)}</span>
      </a>
    `).join("");
  }

  function linkIconHTML(icon, className) {
    const preset = Object.hasOwn(LINK_ICONS, icon) ? LINK_ICONS[icon] : LINK_ICONS.webpage;
    if (preset.monochrome) {
      // CSS resolves mask URLs relative to the stylesheet, so use the document's absolute asset URL.
      const url = D.escapeHTML(new URL(preset.icon, document.baseURI).href);
      return `<span class="${className} link-icon-mask" style="--link-icon-url: url('${url}')" aria-hidden="true"></span>`;
    }
    return `<img class="${className}" src="${preset.icon}" alt="" aria-hidden="true">`;
  }

  function renderPublicationControls() {
    const types = ["international-journal", "domestic-journal", "international-conference", "domestic-conference"];
    $("#pub-type").innerHTML = [
      `<option value="all">${t("allTypes")}</option>`,
      ...types.map((type) => `<option value="${type}">${D.escapeHTML(D.publicationTypeLabel({ publication_type: type }, state.lang))}</option>`)
    ].join("");
    $("#pub-type").value = state.publication.publicationType;

    $("#pub-topic").innerHTML = [
      `<option value="${ALL_TOPIC_FILTER}">${t("allTopics")}</option>`,
      ...publicationTopics().map((topic) => (
        `<option value="${D.escapeHTML(topic.id)}">${D.escapeHTML(D.publicationTopicLabel({ topic: topic.id }, state.data.settings, state.lang))}</option>`
      )),
      `<option value="${FALLBACK_TOPIC_FILTER}">${D.escapeHTML(D.publicationTopicLabel({ topic: FALLBACK_TOPIC_FILTER }, state.data.settings, state.lang))}</option>`
    ].join("");
    $("#pub-topic").value = state.publication.topic;

    $("#pub-index").innerHTML = [
      `<option value="all">${t("allIndexes")}</option>`,
      '<option value="SCIE">SCIE</option>',
      '<option value="SCOPUS">SCOPUS</option>',
      '<option value="KCI">KCI</option>',
      `<option value="other">${t("otherIndexes")}</option>`
    ].join("");
    $("#pub-index").value = state.publication.indexing;

    $$("[data-pub-sort]").forEach((button) => {
      const isActive = button.dataset.pubSort === state.publication.sort;
      button.classList.toggle("is-active", isActive);
      button.setAttribute("aria-pressed", String(isActive));
    });

    $("#pub-first-author").checked = state.publication.firstAuthorOnly;
    $("#pub-search").value = state.publication.query;
  }

  function publicationTopics() {
    return state.data.settings.topics;
  }

  function renderPublicationCollaborators() {
    const counts = new Map();
    state.data.publications.forEach((publication) => {
      const collaborators = new Set(
        publication.authors.filter((person) => !person.is_self).map((person) => person.id)
      );
      collaborators.forEach((id) => counts.set(id, (counts.get(id) || 0) + 1));
    });

    const locale = state.lang === "ko" ? "ko-KR" : "en";
    const collaborators = state.data.people
      .filter((person) => !person.is_self && counts.has(person.id))
      .sort((a, b) => D.personName(a, state.lang).localeCompare(
        D.personName(b, state.lang),
        locale,
        { sensitivity: "base" }
      ));

    $("#publication-collaborators").innerHTML = `
      <span class="publication-collaborators-label">Thanks to…</span>
      <div class="publication-collaborator-list">
        ${collaborators.map((person) => {
          const isActive = state.publication.collaboratorId === person.id;
          return `
            <button type="button" class="publication-collaborator${isActive ? " is-active" : ""}"
              data-collaborator-id="${D.escapeHTML(person.id)}" aria-pressed="${isActive}">
              ${D.escapeHTML(D.personName(person, state.lang))}<span class="publication-collaborator-count">(${counts.get(person.id)})</span>
              ${D.personNotesHTML(person, state.lang)}
            </button>
          `;
        }).join("")}
      </div>
    `;
  }

  function comparePublicationRecency(a, b) {
    return D.publicationStatusRank(b) - D.publicationStatusRank(a)
      || String(b.date || "").localeCompare(String(a.date || ""))
      || D.publicationTitle(a, state.lang).localeCompare(D.publicationTitle(b, state.lang));
  }

  function filteredPublications() {
    const query = D.normalizeText(state.publication.query);
    const topicOrder = new Map(publicationTopics().map((topic, index) => [topic.id, index]));
    const configuredTopicIds = new Set(topicOrder.keys());
    const fallbackTopicOrder = publicationTopics().length;
    return state.data.publications
      .filter((item) => state.publication.publicationType === "all"
        || item.publication_type === state.publication.publicationType)
      .filter((item) => state.publication.indexing === "all"
        || (state.publication.indexing === "other" ? !item.indexing : item.indexing === state.publication.indexing))
      .filter((item) => {
        if (state.publication.topic === ALL_TOPIC_FILTER) return true;
        if (state.publication.topic === FALLBACK_TOPIC_FILTER) return !configuredTopicIds.has(item.topic);
        return item.topic === state.publication.topic;
      })
      .filter((item) => !state.publication.collaboratorId
        || item.authors.some((person) => person.id === state.publication.collaboratorId))
      .filter((item) => !state.publication.firstAuthorOnly || D.isFirstAuthor(item))
      .filter((item) => {
        if (!query) return true;
        return D.normalizeText([
          item.title_en,
          item.title_ko,
          item.abstract_en,
          item.abstract_ko,
          D.splitList(item.keywords_en).join(" "),
          D.splitList(item.keywords_ko).join(" "),
          item.authors.flatMap((person) => [person.name_en, person.name_ko]).join(" "),
          item.venue_en,
          item.venue_kr,
          item.date
        ].join(" ")).includes(query);
      })
      .sort((a, b) => {
        if (state.publication.sort === "cited") {
          return citations.compare(a, b, comparePublicationRecency);
        }
        if (state.publication.sort === "type") {
          return PUBLICATION_TYPES.indexOf(a.publication_type) - PUBLICATION_TYPES.indexOf(b.publication_type)
            || comparePublicationRecency(a, b);
        }
        if (state.publication.sort === "topic") {
          return (topicOrder.get(a.topic) ?? fallbackTopicOrder) - (topicOrder.get(b.topic) ?? fallbackTopicOrder)
            || comparePublicationRecency(a, b);
        }
        return comparePublicationRecency(a, b);
      });
  }

  function publicationTimelineTopics() {
    return [
      ...publicationTopics().map((topic, index) => ({
        id: topic.id,
        label: D.pick(topic, "label", state.lang),
        color: D.topicColor(topic, index)
      })),
      {
        id: FALLBACK_TOPIC_FILTER,
        label: D.pick(state.data.settings.topic_fallback, "label", state.lang),
        color: "#6b6b6b"
      }
    ];
  }

  function publicationTimelineLabelWidth() {
    const probe = document.createElement("div");
    probe.className = "publication-timeline-type";
    probe.style.cssText = "position:absolute;top:0;left:0;width:max-content;visibility:hidden;pointer-events:none";
    document.body.append(probe);
    const width = Math.max(...PUBLICATION_TYPES.map((type) => {
      probe.textContent = D.publicationTypeLabel({ publication_type: type }, state.lang);
      return probe.getBoundingClientRect().width;
    }));
    probe.remove();
    return width;
  }

  function scrollPublicationTimelineToLatest() {
    const scroller = $("[data-publication-timeline-scroll]");
    if (!scroller || !scroller.clientWidth || scroller.dataset.initialPositionSet === "true") return;
    const target = Number(scroller.dataset.latestPosition || 0);
    const labelWidth = scroller.querySelector(".publication-timeline-type")?.offsetWidth || 0;
    scroller.scrollLeft = Math.max(0, labelWidth + target + 40 - scroller.clientWidth);
    scroller.dataset.initialPositionSet = "true";
  }

  function renderPublicationTimeline(items) {
    hidePublicationTimelineTooltip();
    const root = $("#publications-timeline");
    if (!items.length) {
      root.hidden = true;
      root.innerHTML = "";
      return;
    }
    const previous = root.querySelector("[data-publication-timeline-scroll]");
    const filterKey = JSON.stringify([
      state.publication.query, state.publication.publicationType, state.publication.topic, state.publication.indexing,
      state.publication.collaboratorId, state.publication.firstAuthorOnly
    ]);
    const preservePosition = previous?.dataset.initialPositionSet === "true" && previous.dataset.filterKey === filterKey;
    const previousScroll = previous?.scrollLeft || 0;
    const layout = window.PublicationTimeline.layout(items.map((item) => ({
      id: item.publicationIndex, date: item.date, type: item.publication_type,
      firstAuthor: D.isFirstAuthor(item)
    })), {
      types: PUBLICATION_TYPES,
      yearWidth: PROJECT_TIMELINE_YEAR_WIDTH,
      axisDates: state.data.publications.map((item) => item.date)
    });
    const topics = publicationTimelineTopics();
    const colors = new Map(topics.map((topic) => [topic.id, topic.color]));
    const byId = new Map(items.map((item) => [item.publicationIndex, item]));
    const years = layout.yearTicks.filter((tick) => tick.year <= layout.lastYear).map((tick) => (
      `<span class="publication-timeline-year" style="left:${tick.x.toFixed(3)}px">${tick.year}</span>`
    )).join("");
    const yearLines = layout.yearTicks.map((tick) => (
      `<span class="publication-timeline-year-line" style="left:${tick.x.toFixed(3)}px" aria-hidden="true"></span>`
    )).join("");
    const pendingArea = layout.undatedWidth
      ? `<span class="publication-timeline-undated-area" aria-hidden="true"></span>` : "";
    const rows = layout.rows.map((row) => {
      const marks = row.marks.map((mark) => {
        const item = byId.get(mark.id);
        const color = colors.get(item.topic) || colors.get(FALLBACK_TOPIC_FILTER);
        const role = D.isFirstAuthor(item) ? t("firstAuthorOnly") : t("coAuthor");
        return `<button type="button" class="publication-timeline-mark${D.isFirstAuthor(item) ? " is-first-author" : ""}"
          style="left:${mark.x.toFixed(3)}px;top:${mark.y.toFixed(3)}px;--publication-topic-color:${color}"
          data-publication-timeline-target="${mark.id}" data-publication-timeline-date="${D.escapeHTML(item.date || "")}"
          aria-label="${D.escapeHTML(`${D.publicationTitle(item, state.lang)} · ${role}`)}"
          aria-describedby="publication-timeline-tooltip"></button>`;
      }).join("");
      return `<div class="publication-timeline-row" data-publication-timeline-type="${D.escapeHTML(row.type)}">
        <div class="publication-timeline-type">${D.escapeHTML(D.publicationTypeLabel({ publication_type: row.type }, state.lang))}</div>
        <div class="publication-timeline-track" style="height:${row.height}px">${yearLines}${pendingArea}${marks}</div>
      </div>`;
    }).join("");
    const legend = topics.map((topic) => (
      `<span class="publication-timeline-legend-item"><span class="publication-timeline-swatch" style="--publication-topic-color:${topic.color}" aria-hidden="true"></span>${D.escapeHTML(topic.label)}</span>`
    )).join("");
    root.hidden = false;
    root.setAttribute("aria-label", t("publicationTimeline"));
    root.innerHTML = `
      <div class="publication-timeline-scroll" data-publication-timeline-scroll tabindex="0" aria-label="${D.escapeHTML(t("publicationTimeline"))}">
        <div class="publication-timeline-canvas" style="--publication-type-width:${publicationTimelineLabelWidth()}px;--publication-axis-width:${layout.width}px;--publication-undated-width:${layout.undatedWidth}px;--publication-year-width:${layout.yearWidth}px">
          <div class="publication-timeline-header">
            <div class="publication-timeline-corner" aria-hidden="true"></div>
            <div class="publication-timeline-axis">${years}${layout.undatedWidth ? `<span class="publication-timeline-undated-heading">${D.escapeHTML(t("publicationDatePending"))}</span>` : ""}</div>
          </div>
          ${rows}
        </div>
      </div>
      <div class="publication-timeline-legend">
        ${legend}
        <span class="publication-timeline-legend-roles">
          <span class="publication-timeline-legend-separator" aria-hidden="true"></span>
          <span class="publication-timeline-legend-item is-author-role"><span class="publication-timeline-swatch is-author-role" aria-hidden="true"></span>${D.escapeHTML(t("firstAuthorOnly"))}</span>
          <span class="publication-timeline-legend-item is-author-role"><span class="publication-timeline-swatch is-author-role is-outline" aria-hidden="true"></span>${D.escapeHTML(t("coAuthor"))}</span>
        </span>
      </div>`;
    const scroller = root.querySelector("[data-publication-timeline-scroll]");
    scroller.dataset.filterKey = filterKey;
    scroller.dataset.latestPosition = String(Math.max(...layout.rows.flatMap((row) => row.marks.map((mark) => mark.x)), 0));
    if (preservePosition) {
      scroller.dataset.initialPositionSet = "true";
      scroller.scrollLeft = previousScroll;
    }
    window.requestAnimationFrame(scrollPublicationTimelineToLatest);
  }

  function hidePublicationTimelineTooltip() {
    const tooltip = $("#publication-timeline-tooltip");
    if (tooltip) tooltip.hidden = true;
  }

  function hideProjectTimelineTooltip() {
    const tooltip = $("#project-timeline-tooltip");
    if (tooltip) tooltip.hidden = true;
  }

  let publicationActionTooltipButton = null;

  function hidePublicationActionTooltip() {
    window.clearTimeout(hidePublicationActionTooltip.timer);
    publicationActionTooltipButton?.removeAttribute("aria-describedby");
    publicationActionTooltipButton = null;
    const tooltip = $("#publication-action-tooltip");
    if (tooltip) tooltip.hidden = true;
  }

  function scheduleHidePublicationActionTooltip() {
    window.clearTimeout(hidePublicationActionTooltip.timer);
    hidePublicationActionTooltip.timer = window.setTimeout(() => {
      const tooltip = $("#publication-action-tooltip");
      if (tooltip?.matches(":hover") || publicationActionTooltipButton?.matches(":hover, :focus-visible")) return;
      hidePublicationActionTooltip();
    }, 150);
  }

  function showPublicationActionTooltip(button) {
    const citation = button.hasAttribute("data-copy-citation");
    const index = Number(citation ? button.dataset.copyCitation : button.dataset.exportOne);
    const item = state.data.publications[index];
    const tooltip = $("#publication-action-tooltip");
    if (!item || !tooltip) return;
    hidePublicationActionTooltip();
    publicationActionTooltipButton = button;
    tooltip.dataset.kind = citation ? "citation" : "bib";
    tooltip.querySelector(".publication-action-tooltip-label").textContent = citation ? t("copy") : `${D.publicationCitationKey(item)}.bib`;
    tooltip.querySelector(".publication-action-tooltip-content").textContent = citation ? D.publicationCitation(item, state.lang) : `${D.bibtexEntry(item)}\n`;
    button.setAttribute("aria-describedby", tooltip.id);
    const viewportWidth = document.documentElement.clientWidth;
    const viewportHeight = document.documentElement.clientHeight;
    const minTop = Math.max(12, ($(".site-header")?.getBoundingClientRect().bottom || 0) + 12);
    tooltip.style.width = `${Math.min(560, viewportWidth - 24)}px`;
    tooltip.style.maxHeight = `${Math.min(440, Math.max(0, viewportHeight - minTop - 12))}px`;
    tooltip.hidden = false;
    tooltip.scrollTop = 0;
    const rect = button.getBoundingClientRect();
    const tip = tooltip.getBoundingClientRect();
    const below = rect.bottom + 8;
    const above = rect.top - tip.height - 8;
    const top = below + tip.height <= viewportHeight - 12 ? below : above >= minTop ? above : viewportHeight - tip.height - 12;
    tooltip.style.left = `${Math.max(12, Math.min(rect.right - tip.width, viewportWidth - tip.width - 12))}px`;
    tooltip.style.top = `${Math.max(minTop, top)}px`;
  }

  function showPublicationTimelineTooltip(mark) {
    const item = state.data.publications[Number(mark.dataset.publicationTimelineTarget)];
    if (!item) return;
    const authors = (item.authors || []).map((person) => D.personName(person, state.lang)).join(", ");
    const date = item.date || t("publicationDatePending");
    showTimelineTooltip(mark, $("#publication-timeline-tooltip"),
      D.publicationTitle(item, state.lang), `${authors} · ${date}`,
      [D.publicationReference(item, state.lang), D.publicationConferenceInfo(item, state.lang)].filter(Boolean).join("\n") || t("publicationVenuePending"));
  }

  function showProjectTimelineTooltip(bar) {
    const item = state.data.projects[Number(bar.dataset.projectTimelineTarget)];
    if (!item) return;
    showTimelineTooltip(bar, $("#project-timeline-tooltip"),
      p(item, "title"), D.projectPeriod(item), p(item, "funder"));
  }

  function showTimelineTooltip(target, tooltip, title, metadata, detail) {
    if (!tooltip) return;
    tooltip.innerHTML = `
      <p class="timeline-tooltip-title">${D.escapeHTML(title)}</p>
      <p class="timeline-tooltip-meta">${D.escapeHTML(metadata)}</p>
      <p class="timeline-tooltip-detail">${D.escapeHTML(detail)}</p>`;
    const viewportWidth = document.documentElement.clientWidth;
    tooltip.style.maxWidth = `${Math.min(440, viewportWidth - 24)}px`;
    tooltip.hidden = false;
    const rect = target.getBoundingClientRect();
    const tip = tooltip.getBoundingClientRect();
    const viewportHeight = document.documentElement.clientHeight;
    const headerBottom = $(".site-header")?.getBoundingClientRect().bottom || 0;
    const minTop = Math.max(12, headerBottom + 12);
    const left = Math.max(12, Math.min(rect.left + rect.width / 2 - tip.width / 2, viewportWidth - tip.width - 12));
    const above = rect.top - tip.height - 10;
    const top = above >= minTop ? above : Math.min(rect.bottom + 10, viewportHeight - tip.height - 12);
    tooltip.style.left = `${left}px`;
    tooltip.style.top = `${Math.max(minTop, top)}px`;
  }

  function scrollToPublication(index) {
    const article = $(`.publication-item[data-publication-index="${index}"]`);
    if (!article) return;
    hidePublicationTimelineTooltip();
    $$(".publication-item.is-timeline-target").forEach((item) => item.classList.remove("is-timeline-target"));
    article.classList.add("is-timeline-target");
    article.tabIndex = -1;
    article.scrollIntoView({ behavior: "smooth", block: "start", inline: "nearest" });
    article.focus({ preventScroll: true });
    window.clearTimeout(scrollToPublication.timer);
    scrollToPublication.timer = window.setTimeout(() => article.classList.remove("is-timeline-target"), 1800);
  }

  function publicationCitationHTML(item) {
    const record = citations.get(item);
    if (!record) return "";
    return `<a class="publication-citation-count" data-citation-source="${item.publicationIndex}" href="${D.escapeHTML(window.PublicationCitations.citingURL(record))}" target="_blank" rel="noreferrer">Cited by ${record.count.toLocaleString(state.lang === "ko" ? "ko-KR" : "en-US")}</a>`;
  }

  function hidePublicationCitationTooltip() {
    const tooltip = $("#publication-citation-tooltip");
    if (tooltip) tooltip.hidden = true;
    $$("[data-citation-source][aria-describedby]").forEach((link) => link.removeAttribute("aria-describedby"));
  }

  function showPublicationCitationTooltip(link) {
    const item = state.data?.publications[Number(link.dataset.citationSource)];
    const record = item && citations.get(item);
    if (!record) return;
    const source = record.provider === "kci" ? "KCI · kci.go.kr" : "OpenAlex · openalex.org";
    const checkedAt = new Date(record.checked_at).toLocaleDateString("en-CA");
    showTimelineTooltip(link, $("#publication-citation-tooltip"), source,
      `${state.lang === "ko" ? "조회일" : "Retrieved"}: ${checkedAt}`,
      record.list_pending ? (state.lang === "ko" ? "KCI 인용논문 목록 공개 대기" : "Citing-paper list pending release by KCI") : "");
    link.setAttribute("aria-describedby", "publication-citation-tooltip");
  }

  function updatePublicationCitationOrder() {
    if (!citationOrderPending || state.publication.sort !== "cited" || !state.data) return;
    const list = $("#publication-list");
    const selection = window.getSelection();
    if (list.querySelector(".publication-summary:active")) return;
    if (selection && !selection.isCollapsed
      && (list.contains(selection.anchorNode) || list.contains(selection.focusNode))) return;
    const groups = new Map([...list.children].map((group) => [Number(group.dataset.citationRankIndex), group]));
    const focused = document.activeElement;
    let cursor = list.firstElementChild;
    for (const item of filteredPublications()) {
      const group = groups.get(item.publicationIndex);
      if (!group) continue;
      if (group !== cursor) {
        if (list.moveBefore) list.moveBefore(group, cursor);
        else list.insertBefore(group, cursor);
      }
      cursor = group.nextElementSibling;
    }
    if (focused?.isConnected && document.activeElement !== focused) focused.focus({ preventScroll: true });
    citationOrderPending = false;
  }

  function updatePublicationCitations() {
    $$("[data-publication-citations]").forEach((slot) => {
      const item = state.data?.publications[Number(slot.dataset.publicationCitations)];
      if (!item) return;
      const html = publicationCitationHTML(item);
      const link = slot.querySelector("[data-citation-source]");
      if (html && link) {
        const record = citations.get(item);
        link.textContent = `Cited by ${record.count.toLocaleString(state.lang === "ko" ? "ko-KR" : "en-US")}`;
        link.href = window.PublicationCitations.citingURL(record);
        if (!$("#publication-citation-tooltip").hidden && (link === document.activeElement || link.matches(":hover"))) {
          showPublicationCitationTooltip(link);
        }
      } else if (slot.innerHTML !== html) {
        slot.innerHTML = html;
      }
      slot.hidden = !html;
    });
    citationOrderPending = state.publication.sort === "cited";
    updatePublicationCitationOrder();
  }

  function publicationItemHTML(item) {
    const doi = window.PublicationCitations.normalizeDOI(item.doi);
    const sourceURL = doi ? `https://doi.org/${doi}` : item.url;
    const citationActions = D.truthy(item.under_review) || D.truthy(item.in_press)
      ? ""
      : `
          <button type="button" class="publication-action" data-copy-citation="${item.publicationIndex}">${t("copy")}</button>
          <button type="button" class="publication-action" data-export-one="${item.publicationIndex}">${t("bib")}</button>
        `;
    const publicationActions = `
      ${sourceURL ? `<a class="publication-action" href="${D.escapeHTML(sourceURL)}" target="_blank" rel="noreferrer">${t("source")}</a>` : ""}
      ${citationActions}
    `.trim();
    const title = D.publicationTitle(item, state.lang);
    const venue = D.publicationReference(item, state.lang);
    const venueHTML = `<span class="publication-reference">${D.escapeHTML(venue)}</span>`;
    const indexingLabels = { SCIE: "Science Citation Index Expanded (SCIE)", SCOPUS: "Scopus", KCI: "KCI 등재지 / Korea Citation Index" };
    const indexing = Object.hasOwn(indexingLabels, item.indexing) ? item.indexing : "";
    const indexingTag = indexing
      ? `<span class="badge badge-indexing" data-indexing="${indexing}" title="${D.escapeHTML(indexingLabels[indexing])}">${indexing}</span>` : "";
    const publicationStatus = D.publicationStatus(item);
    const abstract = p(item, "abstract");
    const keywords = D.splitList(p(item, "keywords"));
    const isJournal = item.publication_type.endsWith("-journal");
    const catalog = window.PublicationBibliography.catalogFor(item, state.data.settings);
    const dateSource = isJournal ? item : catalog;
    const dateFields = isJournal
      ? [["online_date", "onlinePublicationDate"], ["issue_date", "issuePublicationDate"]]
      : [["proceedings_date", "issuePublicationDate"]];
    const publicationDetails = dateFields
      .map(([field, label]) => [t(label), String(dateSource?.[field] || "").trim()])
      .filter(([, value]) => value);
    const publisher = String(p(catalog, isJournal ? "publisher" : "proceedings_publisher")).trim();
    if (publisher) publicationDetails.push([t(isJournal ? "publicationPublisher" : "proceedingsPublisher"), publisher]);
    const conferenceInfo = isJournal ? "" : D.publicationConferenceInfo(item, state.lang).trim();
    const detailRows = [];
    if (conferenceInfo) detailRows.push([[t("publicationConference"), conferenceInfo]]);
    if (publicationDetails.length) detailRows.push(publicationDetails);
    const hasDetails = Boolean(abstract || keywords.length || detailRows.length);
    const itemIndex = item.publicationIndex;
    const isExpanded = hasDetails && state.publication.expanded.has(itemIndex);
    const detailsId = `publication-details-${itemIndex}`;
    const awardLabel = item.award_id?.trim()
      ? `<span class="badge badge-award">${t("award")}</span>`
      : "";
    const summaryAttributes = hasDetails
      ? `class="publication-summary" type="button" data-publication-toggle="${itemIndex}" aria-expanded="${isExpanded}" aria-controls="${D.escapeHTML(detailsId)}"`
      : `class="publication-summary publication-summary-static"`;

    return `
      <article class="publication-item${isExpanded ? " is-expanded" : ""}" data-publication-index="${itemIndex}">
        <${hasDetails ? "button" : "div"} ${summaryAttributes}>
          <div class="publication-main">
            <div class="publication-type">
              <span class="badge">${D.escapeHTML(D.publicationTypeLabel(item, state.lang))}</span>
              ${indexingTag}
              ${awardLabel}
            </div>
            <h3>${D.escapeHTML(title)}</h3>
            <p class="publication-authors">${D.displayAuthors(item.authors, state.lang)}</p>
            <p class="publication-venue">${publicationStatus
              ? `<em>(${D.escapeHTML(publicationStatus)})</em>${publicationStatus === "in press" && venue ? ` ${venueHTML}` : ""}`
              : venueHTML}</p>
            ${item.note?.trim() ? `<p class="publication-note">${D.escapeHTML(item.note.trim())}</p>` : ""}
          </div>
        </${hasDetails ? "button" : "div"}>
        <div class="publication-links">
          ${publicationActions ? `<div class="publication-buttons">${publicationActions}</div>` : ""}
          <div class="publication-citations" data-publication-citations="${itemIndex}"${citations.get(item) ? "" : " hidden"}>${publicationCitationHTML(item)}</div>
        </div>
        ${hasDetails ? `
          <div class="publication-details" id="${D.escapeHTML(detailsId)}" aria-hidden="${!isExpanded}">
            <div class="publication-details-inner">
              ${abstract ? `
                <section class="publication-abstract">
                  <h4>${state.lang === "ko" ? "초록" : "Abstract"}</h4>
                  <p>${D.escapeHTML(abstract)}</p>
                </section>
              ` : ""}
              ${keywords.length ? `
                <section class="publication-keywords">
                  <h4>${state.lang === "ko" ? "주제어" : "Keywords"}</h4>
                  <ul>${keywords.map((keyword) => `<li>${D.escapeHTML(keyword)}</li>`).join("")}</ul>
                </section>
              ` : ""}
              ${detailRows.map((row) => `
                <p class="publication-dates">${row.map(([label, value]) => (
                  `<span class="publication-date"><span class="publication-date-label">${D.escapeHTML(label)}</span> ${D.escapeHTML(value)}</span>`
                )).join(" · ")}</p>
              `).join("")}
            </div>
          </div>
        ` : ""}
      </article>
    `;
  }

  function publicationYearGroupsHTML(items) {
    const groups = [];
    items.forEach((item) => {
      const year = D.publicationYear(item);
      const current = groups.at(-1);
      if (!current || current.year !== year) {
        groups.push({ year, items: [item] });
      } else {
        current.items.push(item);
      }
    });

    return groups.map((group) => `
      <section class="publication-year-group" aria-label="${D.escapeHTML(group.year)}">
        <div class="publication-year">${D.escapeHTML(group.year)}</div>
        <div class="publication-year-items">
          ${group.items.map(publicationItemHTML).join("")}
        </div>
      </section>
    `).join("");
  }

  function renderPublications() {
    if (!pageIsVisible("publications")) return;
    hidePublicationActionTooltip();
    hidePublicationCitationTooltip();
    citationOrderPending = false;
    renderPublicationControls();
    renderPublicationCollaborators();
    const items = filteredPublications();
    renderPublicationTimeline(items);
    $("#pub-result-count").textContent = String(items.length);
    $("#publications-empty").hidden = items.length > 0;
    $("#publication-list").className = `publication-list${["type", "topic"].includes(state.publication.sort) ? " is-grouped" : ""}`;

    if (!items.length) {
      $("#publication-list").innerHTML = "";
      return;
    }

    if (state.publication.sort === "cited") {
      $("#publication-list").innerHTML = items.map((item) => `
        <section class="publication-year-group" data-citation-rank-index="${item.publicationIndex}" aria-label="${D.escapeHTML(D.publicationYear(item))}">
          <div class="publication-year">${D.escapeHTML(D.publicationYear(item))}</div>
          <div class="publication-year-items">${publicationItemHTML(item)}</div>
        </section>`).join("");
      return;
    }

    if (state.publication.sort === "type") {
      $("#publication-list").innerHTML = PUBLICATION_TYPES.map((type) => {
        const typeItems = items.filter((item) => item.publication_type === type);
        if (!typeItems.length) return "";
        return `
          <section class="publication-type-group">
            <div class="publication-type-heading" role="heading" aria-level="3">
              <span>${D.escapeHTML(D.publicationTypeLabel({ publication_type: type }, state.lang))}</span>
              <span class="publication-type-count">${typeItems.length}${state.lang === "ko" ? "" : " "}${t("publicationCount")}</span>
            </div>
            ${publicationYearGroupsHTML(typeItems)}
          </section>
        `;
      }).join("");
      return;
    }

    if (state.publication.sort === "topic") {
      const configuredTopicIds = new Set(publicationTopics().map((topic) => topic.id));
      const groups = [
        ...publicationTopics().map((topic) => ({
          id: topic.id,
          items: items.filter((item) => item.topic === topic.id)
        })),
        {
          id: "",
          items: items.filter((item) => !configuredTopicIds.has(item.topic))
        }
      ];

      $("#publication-list").innerHTML = groups.map((group) => {
        if (!group.items.length) return "";
        return `
          <section class="publication-type-group">
            <div class="publication-type-heading" role="heading" aria-level="3">
              <span>${D.escapeHTML(D.publicationTopicLabel({ topic: group.id }, state.data.settings, state.lang))}</span>
              <span class="publication-type-count">${group.items.length}${state.lang === "ko" ? "" : " "}${t("publicationCount")}</span>
            </div>
            ${publicationYearGroupsHTML(group.items)}
          </section>
        `;
      }).join("");
      return;
    }

    $("#publication-list").innerHTML = publicationYearGroupsHTML(items);
  }

  function itemMedia(item) {
    if (Array.isArray(item.media)) {
      return item.media.map((mediaItem) => ({
        src: D.contentMediaPath(mediaItem),
        type: D.contentMediaType(mediaItem),
        poster: mediaItem.poster ? D.contentMediaPath(mediaItem.poster) : "",
        caption: (state.lang === "ko" ? mediaItem.caption_ko : mediaItem.caption_en)
          || mediaItem.caption_en
          || mediaItem.caption_ko
          || p(item, "title")
          || item.name
          || ""
      })).filter((mediaItem) => mediaItem.src);
    }
    return [];
  }

  function noteContent(item) {
    return item[`content_${state.lang}`] || item.content_en || item.content_ko || "";
  }

  function renderMediaGallery(item, mediaItems, collection, modifier = "", itemIndex = 0) {
    const isMultiple = mediaItems.length > 1;
    const galleryTitle = p(item, "title") || item.name || "";
    const slides = mediaItems.map((mediaItem, index) => {
      const mediaElement = mediaItem.type === "video"
        ? `<video src="${D.escapeHTML(mediaItem.src)}"${mediaItem.poster ? ` poster="${D.escapeHTML(mediaItem.poster)}"` : ""} preload="metadata" muted playsinline aria-hidden="true"></video><span class="media-play-indicator" aria-hidden="true">▶</span>`
        : `<img src="${D.escapeHTML(mediaItem.src)}" alt="${D.escapeHTML(mediaItem.caption)}" width="1200" height="760" loading="lazy">`;
      return `
      <figure class="note-gallery-slide" data-gallery-slide aria-hidden="${index === 0 ? "false" : "true"}">
        <button type="button" class="note-media-button" data-open-lightbox data-lightbox-collection="${D.escapeHTML(collection)}" data-lightbox-item-index="${itemIndex}" data-media-index="${index}" aria-label="${D.escapeHTML(`${t("expandMedia")} ${index + 1}`)}" tabindex="${index === 0 ? "0" : "-1"}">
          ${mediaElement}
        </button>
        <figcaption>${D.escapeHTML(mediaItem.caption)}</figcaption>
      </figure>
    `;
    }).join("");

    return `
      <div class="note-gallery${modifier ? ` ${modifier}` : ""}" data-media-gallery data-gallery-index="0">
        <div class="note-gallery-viewport" data-gallery-viewport ${isMultiple ? 'tabindex="0"' : ""} aria-label="${D.escapeHTML(`${galleryTitle} · ${t("mediaGallery")}`)}">
          <div class="note-gallery-track">${slides}</div>
        </div>
        ${isMultiple ? `
          <button type="button" class="note-gallery-arrow note-gallery-prev" data-gallery-step="-1" aria-label="${D.escapeHTML(t("previousMedia"))}" disabled>←</button>
          <button type="button" class="note-gallery-arrow note-gallery-next" data-gallery-step="1" aria-label="${D.escapeHTML(t("nextMedia"))}">→</button>
          <div class="note-gallery-footer">
            <div class="note-gallery-dots" aria-hidden="true">
              ${mediaItems.map((_, index) => `<span class="note-gallery-dot${index === 0 ? " is-active" : ""}" data-gallery-dot="${index}"></span>`).join("")}
            </div>
            <span class="note-gallery-count" data-gallery-count>1 / ${mediaItems.length}</span>
          </div>
        ` : ""}
      </div>
    `;
  }

  function updateGalleryControls(gallery, index) {
    const slides = $$('[data-gallery-slide]', gallery);
    if (!slides.length) return;
    const nextIndex = Math.max(0, Math.min(index, slides.length - 1));
    gallery.dataset.galleryIndex = String(nextIndex);
    slides.forEach((slide, slideIndex) => {
      slide.setAttribute("aria-hidden", String(slideIndex !== nextIndex));
      const mediaButton = $("[data-open-lightbox]", slide);
      if (mediaButton) mediaButton.tabIndex = slideIndex === nextIndex ? 0 : -1;
    });
    $$('[data-gallery-dot]', gallery).forEach((dot, dotIndex) => {
      dot.classList.toggle("is-active", dotIndex === nextIndex);
    });
    const previous = $('[data-gallery-step="-1"]', gallery);
    const next = $('[data-gallery-step="1"]', gallery);
    if (previous) previous.disabled = nextIndex === 0;
    if (next) next.disabled = nextIndex === slides.length - 1;
    const count = $("[data-gallery-count]", gallery);
    if (count) count.textContent = `${nextIndex + 1} / ${slides.length}`;
  }

  function setGalleryIndex(gallery, index, options = {}) {
    const viewport = $("[data-gallery-viewport]", gallery);
    const slides = $$('[data-gallery-slide]', gallery);
    if (!viewport || !slides.length) return;
    const nextIndex = Math.max(0, Math.min(index, slides.length - 1));
    updateGalleryControls(gallery, nextIndex);
    viewport.scrollTo({
      left: nextIndex * viewport.clientWidth,
      behavior: options.instant || window.matchMedia("(prefers-reduced-motion: reduce)").matches ? "auto" : "smooth"
    });
  }

  function bindMediaGalleryScroll() {
    $$('[data-gallery-viewport]').forEach((viewport) => {
      let frame = 0;
      viewport.addEventListener("scroll", () => {
        window.cancelAnimationFrame(frame);
        frame = window.requestAnimationFrame(() => {
          const gallery = viewport.closest("[data-media-gallery]");
          if (!gallery || !viewport.clientWidth) return;
          updateGalleryControls(gallery, Math.round(viewport.scrollLeft / viewport.clientWidth));
        });
      }, { passive: true });
    });
  }

  function boardRecords() {
    return state.data.board
      .map((item, itemIndex) => ({ item, itemIndex }))
      .sort((a, b) => b.item.start_date.localeCompare(a.item.start_date)
        || String(b.item.end_date || "").localeCompare(String(a.item.end_date || ""))
        || a.itemIndex - b.itemIndex);
  }

  function boardDateLabel(item) {
    return item.end_date ? `${item.start_date} – ${item.end_date}` : item.start_date;
  }

  function boardDateHTML(item) {
    const startDate = D.escapeHTML(item.start_date);
    if (!item.end_date) return `<time datetime="${startDate}">${startDate}</time>`;
    const endDate = D.escapeHTML(item.end_date);
    return `<time datetime="${startDate}">${startDate}</time><span aria-hidden="true">–</span><time datetime="${endDate}">${endDate}</time>`;
  }

  function renderBoardList(records) {
    const pageCount = Math.max(1, Math.ceil(records.length / BOARD_PAGE_SIZE));
    state.board.page = Math.max(0, Math.min(state.board.page, pageCount - 1));
    const startIndex = state.board.page * BOARD_PAGE_SIZE;
    const pageRecords = records.slice(startIndex, startIndex + BOARD_PAGE_SIZE);

    $("#board-list").innerHTML = `
      <div class="board-list-table-wrap">
        <table class="board-list-table" aria-label="${D.escapeHTML(t("boardList"))}">
          <thead>
            <tr>
              <th scope="col">${t("boardDate")}</th>
              <th scope="col">${t("boardTitle")}</th>
            </tr>
          </thead>
          <tbody>
            ${pageRecords.map(({ item, itemIndex }) => `
              <tr>
                <td>${D.escapeHTML(boardDateLabel(item))}</td>
                <td><button type="button" class="board-list-title" data-board-target="${itemIndex}">${D.escapeHTML(p(item, "title"))}</button></td>
              </tr>
            `).join("")}
          </tbody>
        </table>
      </div>
      <nav class="board-pagination" aria-label="${D.escapeHTML(t("boardPagination"))}">
        <button type="button" data-board-page="-1"${state.board.page === 0 ? " disabled" : ""}>← ${t("previousPage")}</button>
        <span aria-live="polite">${state.board.page + 1} / ${pageCount}</span>
        <button type="button" data-board-page="1"${state.board.page === pageCount - 1 ? " disabled" : ""}>${t("nextPage")} →</button>
      </nav>
    `;
  }

  function setActiveBoardEntry(itemIndex) {
    state.board.activeItemIndex = itemIndex;
    $$('[data-board-dot]').forEach((button) => {
      const active = Number(button.dataset.boardDot) === itemIndex;
      button.classList.toggle("is-active", active);
      if (active) button.setAttribute("aria-current", "location");
      else button.removeAttribute("aria-current");
    });
  }

  function renderBoardRail(records) {
    if (!records.some((record) => record.itemIndex === state.board.activeItemIndex)) {
      state.board.activeItemIndex = records[0]?.itemIndex ?? null;
    }
    const groups = [];
    records.forEach((record) => {
      const year = record.item.start_date.slice(0, 4);
      const current = groups.at(-1);
      if (current?.year === year) current.records.push(record);
      else groups.push({ year, records: [record] });
    });

    $("#board-rail").innerHTML = `
      <div class="board-rail-inner">
        ${groups.map((group) => `
          <section class="board-rail-year-group" aria-label="${D.escapeHTML(group.year)}">
            <div class="board-rail-year">${D.escapeHTML(group.year)}</div>
            <div class="board-rail-dots">
              ${group.records.map(({ item, itemIndex }) => {
                const title = p(item, "title");
                const active = itemIndex === state.board.activeItemIndex;
                return `
                  <button type="button" class="board-rail-dot${active ? " is-active" : ""}"
                    data-board-target="${itemIndex}" data-board-dot="${itemIndex}"
                    aria-label="${D.escapeHTML(`${boardDateLabel(item)}. ${title}`)}"${active ? ' aria-current="location"' : ""}>
                    <span class="board-rail-tooltip" role="tooltip">${D.escapeHTML(title)}</span>
                  </button>
                `;
              }).join("")}
            </div>
          </section>
        `).join("")}
      </div>
    `;
  }

  function updateActiveBoardEntry() {
    if (currentPageView() !== "board") return;
    const entries = $$('[data-board-entry-index]');
    if (!entries.length) return;
    const headerHeight = Number.parseFloat(window.getComputedStyle(document.documentElement).getPropertyValue("--header-height")) || 0;
    const activationLine = headerHeight + Math.min(window.innerHeight * 0.24, 160);
    let activeEntry = entries[0];
    entries.forEach((entry) => {
      if (entry.getBoundingClientRect().top <= activationLine) activeEntry = entry;
    });
    setActiveBoardEntry(Number(activeEntry.dataset.boardEntryIndex));
  }

  function scheduleBoardActiveUpdate() {
    if (scheduleBoardActiveUpdate.frame) return;
    scheduleBoardActiveUpdate.frame = window.requestAnimationFrame(() => {
      scheduleBoardActiveUpdate.frame = 0;
      updateActiveBoardEntry();
    });
  }

  function scrollToBoardEntry(itemIndex) {
    const entry = $(`[data-board-entry-index="${itemIndex}"]`);
    if (!entry) return;
    $$('[data-board-entry-index]').forEach((item) => item.classList.remove("is-board-target"));
    entry.classList.add("is-board-target");
    setActiveBoardEntry(itemIndex);
    entry.scrollIntoView({ behavior: "smooth", block: "start", inline: "nearest" });
    entry.focus({ preventScroll: true });
    window.clearTimeout(scrollToBoardEntry.timer);
    scrollToBoardEntry.timer = window.setTimeout(() => {
      entry.classList.remove("is-board-target");
    }, 1800);
  }

  function renderBoard() {
    if (!pageIsVisible("board")) return;
    const records = boardRecords();
    renderBoardRail(records);
    renderBoardList(records);

    $("#board-grid").innerHTML = records.map(({ item, itemIndex }) => {
      const mediaItems = itemMedia(item);
      const content = noteContent(item);
      return `
        <article id="board-entry-${itemIndex}" class="note-entry" data-board-entry-index="${itemIndex}" tabindex="-1">
          <header class="note-header">
            <div class="note-meta">${boardDateHTML(item)}</div>
            <h2>${D.escapeHTML(p(item, "title"))}</h2>
          </header>
          <div class="note-content">
            ${content ? `<p>${D.escapeHTML(content)}</p>` : ""}
          </div>
          ${mediaItems.length ? renderMediaGallery(item, mediaItems, "board", "", itemIndex) : ""}
        </article>
      `;
    }).join("");
    bindMediaGalleryScroll();
    window.requestAnimationFrame(updateActiveBoardEntry);
    if ($("#media-lightbox").open) updateLightbox();
  }

  function updateLightbox() {
    const items = state.data?.[state.lightbox.collection] || [];
    const item = items[state.lightbox.itemIndex];
    const mediaItems = item ? itemMedia(item) : [];
    if (!mediaItems.length) return;
    state.lightbox.index = (state.lightbox.index + mediaItems.length) % mediaItems.length;
    const mediaItem = mediaItems[state.lightbox.index];
    const image = $("#lightbox-image");
    const video = $("#lightbox-video");
    image.classList.remove("is-entering");
    video.classList.remove("is-entering");
    let activeMedia = image;
    if (mediaItem.type === "video") {
      image.hidden = true;
      image.removeAttribute("src");
      video.pause();
      video.hidden = false;
      video.src = mediaItem.src;
      if (mediaItem.poster) video.poster = mediaItem.poster;
      else video.removeAttribute("poster");
      video.load();
      activeMedia = video;
    } else {
      video.pause();
      video.hidden = true;
      video.removeAttribute("src");
      video.removeAttribute("poster");
      video.load();
      image.hidden = false;
      image.src = mediaItem.src;
      image.alt = mediaItem.caption || item.name || p(item, "title");
    }
    const caption = $("#lightbox-caption");
    caption.textContent = mediaItem.caption;
    caption.hidden = !mediaItem.caption;
    $("#lightbox-count").textContent = `${state.lightbox.index + 1} / ${mediaItems.length}`;
    $("[data-lightbox-prev]").hidden = mediaItems.length < 2;
    $("[data-lightbox-next]").hidden = mediaItems.length < 2;
    $("#lightbox-count").hidden = mediaItems.length < 2;
    window.requestAnimationFrame(() => activeMedia.classList.add("is-entering"));
  }

  function openLightbox(collection, itemIndex, index) {
    const dialog = $("#media-lightbox");
    state.lightbox.collection = collection;
    state.lightbox.itemIndex = itemIndex;
    state.lightbox.index = index;
    updateLightbox();
    if (!dialog.open) dialog.showModal();
  }

  function stepLightbox(delta) {
    state.lightbox.index += delta;
    updateLightbox();
  }

  function renderContact() {
    const identity = state.data.profile.identity;
    const contact = state.data.profile.contact;
    const githubName = contact.github.replace(/\/+$/, "").split("/").pop();
    const orcidId = contact.orcid.replace(/\/+$/, "").split("/").pop();
    const scholarName = D.personName(identity, state.lang);
    const items = [
      { url: contact.github, icon: "github", label: "GitHub", value: githubName },
      { url: contact.orcid, icon: "orcid", label: "ORCID", value: orcidId },
      { url: contact.scholar, icon: "scholar", label: "Scholar", value: scholarName },
      ...(contact.linkedin ? [{ url: contact.linkedin, icon: "linkedin", label: "LinkedIn", value: D.personName(identity, state.lang) }] : []),
      { url: `mailto:${contact.email}`, icon: "email", label: t("email"), value: contact.email },
      { url: "", icon: "address", label: t("address"), value: identity[`location_${state.lang}`] || identity.location_en }
    ];
    $("#contact-links").innerHTML = items.map((item) => {
      const content = `
        ${linkIconHTML(item.icon, "contact-icon")}
        <span class="contact-link-label">${D.escapeHTML(item.label)}</span>
        <span class="contact-link-separator" aria-hidden="true">·</span>
        <span class="contact-link-value">${D.escapeHTML(item.value)}</span>
      `;
      if (!item.url) return `<div class="contact-link contact-link-static">${content}</div>`;
      return `<a class="contact-link" href="${D.escapeHTML(item.url)}" ${item.url.startsWith("http") ? 'target="_blank" rel="noreferrer"' : ""}>${content}</a>`;
    }).join("");
  }

  function renderEducation() {
    $("#education-list").innerHTML = state.data.profile.education.map((item) => {
      const degree = item[`degree_${state.lang}`] || item.degree_en;
      const institution = item[`institution_${state.lang}`] || item.institution_en;
      const advisor = item[`advisor_${state.lang}`] || item.advisor_en;
      const gpa = D.formatGPA(item.gpa);
      return `
        <article class="timeline-item">
          <div class="timeline-period">${D.periodHTML(item.period)}</div>
          <div>
            <h4>${D.escapeHTML(institution)} ${D.escapeHTML(degree)}</h4>
            ${advisor ? `<p>${D.escapeHTML(advisor)}</p>` : ""}
            ${D.educationThesisHTML(item, state.lang)}
            ${gpa ? `<p class="education-gpa">${D.escapeHTML(gpa)}</p>` : ""}
          </div>
        </article>
      `;
    }).join("");
  }

  function renderExperience() {
    $("#experience-list").innerHTML = state.data.profile.experience.map((item) => `
      <article class="timeline-item">
        <div class="timeline-period">${D.periodHTML(p(item, "period"))}</div>
        <div>
          <h4>${D.escapeHTML(p(item, "institution"))}</h4>
          <p>${D.escapeHTML(p(item, "title"))}</p>
        </div>
      </article>
    `).join("");
  }

  function documentButton(item, title) {
    const url = D.documentPath(item.pdf);
    return url ? `<a class="publication-action document-action" href="${D.escapeHTML(url)}" target="_blank" rel="noreferrer" aria-label="${D.escapeHTML(title)} pdf">pdf</a>` : "";
  }

  function renderAwards() {
    const items = [...state.data.awards].sort((a, b) => b.date.localeCompare(a.date));
    $("#awards-list").innerHTML = items.map((item) => {
      const publications = state.data.publications.filter((publication) => publication.award_id === item.id);
      const hasPublications = publications.length > 0;
      const isExpanded = hasPublications && state.award.expanded.has(item.id);
      const detailsId = `award-details-${item.id}`;
      const publicationDetails = publications.map((publication) => `
        <div class="award-publication">
          <p class="award-publication-title">${D.escapeHTML(D.publicationTitle(publication, state.lang))}</p>
          <p class="award-publication-authors">
            ${D.displayAuthors(publication.authors, state.lang, false)}
            (${D.escapeHTML(D.publicationYear(publication) || D.publicationStatus(publication))})
          </p>
          <p class="award-publication-venue">${D.publicationVenueHTML(publication, state.lang)}</p>
        </div>
      `).join("");
      return `
        <article class="timeline-item${item.pdf ? " has-document" : ""}${hasPublications ? ` award-item${isExpanded ? " is-expanded" : ""}` : ""}"${hasPublications ? ` data-award-id="${D.escapeHTML(item.id)}"` : ""}>
          ${documentButton(item, p(item, "title"))}
          ${hasPublications ? `
            <button type="button" class="award-summary" data-award-toggle="${D.escapeHTML(item.id)}"
              aria-expanded="${isExpanded}" aria-controls="${D.escapeHTML(detailsId)}">
              <span class="timeline-period">${D.escapeHTML(item.date)}</span>
              <span class="award-summary-main">
                <span class="award-summary-title">${D.escapeHTML(p(item, "title"))}</span>
                <span class="award-summary-organization">${D.escapeHTML(p(item, "organization"))}</span>
              </span>
            </button>
            <div class="award-details" id="${D.escapeHTML(detailsId)}" aria-hidden="${!isExpanded}">
              <div class="award-details-inner">${publicationDetails}</div>
            </div>
          ` : `
            <div class="timeline-period">${D.escapeHTML(item.date)}</div>
            <div>
              <h4>${D.escapeHTML(p(item, "title"))}</h4>
              <p>${D.escapeHTML(p(item, "organization"))}</p>
            </div>
          `}
        </article>
      `;
    }).join("");
  }

  function renderPatents() {
    $("#patents-list").innerHTML = D.patentGroups(state.data.profile.patents, state.lang).map((group) => {
      return `
        <section class="patent-group" data-patent-type="${group.type}" data-patent-status="${group.status}">
          <h4>${D.escapeHTML(group.title)}</h4>
          <div class="timeline compact-timeline">
            ${group.items.map((item) => {
              const title = D.escapeHTML(D.patentTitle(item, state.lang));
              const url = D.patentURL(item);
              const people = D.patentPeopleHTML(item, state.lang);
              return `
                <article class="timeline-item patent-item">
                  <div class="timeline-period">${D.escapeHTML(D.patentDate(item))}</div>
                  <div>
                    <h5>${url ? `<a href="${D.escapeHTML(url)}" target="_blank" rel="noreferrer">${title}</a>` : title}</h5>
                    ${people ? `<p class="patent-people">${people}</p>` : ""}
                    ${D.patentDetails(item, state.lang).map((detail) => `<p class="patent-detail">${D.escapeHTML(detail)}</p>`).join("")}
                  </div>
                </article>
              `;
            }).join("")}
          </div>
        </section>
      `;
    }).join("");
  }

  async function renderScholarships() {
    const items = state.data.profile.scholarships;
    const language = state.lang;
    const renderVersion = ++state.scholarshipRenderVersion;
    const note = $("#scholarships-conversion-note");
    note.textContent = "";
    note.hidden = true;
    $("#scholarships-list").innerHTML = items.map((item, index) => `
      <article class="timeline-item scholarship-item">
        <div class="timeline-period">${D.periodHTML(p(item, "period"))}</div>
        <div>
          <h4>${D.escapeHTML(p(item, "name"))}</h4>
          ${p(item, "summary") ? `<p>${D.escapeHTML(p(item, "summary"))}</p>` : ""}
          <p class="scholarship-amount">
            <span class="scholarship-amount-label">${D.escapeHTML(t("scholarshipAmount"))}</span>
            <span data-scholarship-amount="${index}" aria-live="polite">${D.escapeHTML(D.formatScholarshipAmount(item.amount, language))}</span>
          </p>
        </div>
      </article>
    `).join("");

    const amounts = await Promise.all(items.map((item) => D.resolveScholarshipAmountDetails(item.amount, language)));
    if (state.scholarshipRenderVersion !== renderVersion || state.lang !== language) return;
    amounts.forEach((amount, index) => {
      const target = $("#scholarships-list")?.querySelector(`[data-scholarship-amount="${index}"]`);
      if (target) target.textContent = amount.text;
    });
    if (language === "en") {
      note.textContent = D.formatScholarshipConversionNote(amounts.map((amount) => amount.conversion));
      note.hidden = !note.textContent;
    }
  }

  function renderCertifications() {
    const items = [...state.data.profile.certifications].sort((a, b) => b.date.localeCompare(a.date));
    $("#certifications-list").innerHTML = items.map((item) => `
      <article class="timeline-item${item.pdf ? " has-document" : ""}">
        ${documentButton(item, p(item, "name"))}
        <div class="timeline-period">${D.escapeHTML(item.date)}</div>
        <div>
          <h4>${D.escapeHTML(p(item, "name"))}</h4>
          <p>${D.escapeHTML(p(item, "issuer"))}</p>
        </div>
      </article>
    `).join("");
  }

  function renderTeaching() {
    $("#teaching-list").innerHTML = state.data.profile.teaching.map((item) => `
      <article class="timeline-item">
        <div class="timeline-period">${D.periodHTML(item.period)}</div>
        <div>
          <h4>${D.escapeHTML(item[`title_${state.lang}`] || item.title_en)}</h4>
          <p>${D.escapeHTML(item[`detail_${state.lang}`] || item.detail_en)}</p>
        </div>
      </article>
    `).join("");
  }

  function academicDate(value) {
    const match = String(value || "").match(/^(\d{4})-(\d{2})-(\d{2})$/);
    return match ? `${match[1]}.${match[2]}.${match[3]}.` : String(value || "");
  }

  function academicPeriod(item) {
    return `${academicDate(item.start_date)} – ${item.end_date ? academicDate(item.end_date) : t("present")}`;
  }

  function academicDateRange(item) {
    return item.end_date
      ? `${academicDate(item.start_date)} – ${academicDate(item.end_date)}`
      : academicDate(item.start_date);
  }

  function academicActivityItem(period, title, detail = "") {
    return `
      <article class="academic-activity-item">
        <div class="academic-activity-period">${D.escapeHTML(period)}</div>
        <div class="academic-activity-copy">
          <h5>${D.escapeHTML(title)}</h5>
          ${detail ? `<p>${D.escapeHTML(detail)}</p>` : ""}
        </div>
      </article>
    `;
  }

  function renderAcademicActivities() {
    const data = state.data.academicActivities;
    const groups = [
      {
        key: "editorial_service",
        title: t("academicEditorialServiceTitle"),
        render: (item) => academicActivityItem(academicPeriod(item), p(item, "journal"), p(item, "role"))
      },
      {
        key: "professional_service",
        title: t("academicProfessionalServiceTitle"),
        render: (item) => academicActivityItem(academicPeriod(item), p(item, "organization"), p(item, "role"))
      },
      {
        key: "conference_service",
        title: t("academicConferenceServiceTitle"),
        render: (item) => academicActivityItem(academicDateRange(item), p(item, "conference"), p(item, "role"))
      },
      {
        key: "invited_talks",
        title: t("academicInvitedTalkTitle"),
        render: (item) => academicActivityItem(academicDate(item.date), p(item, "event"), p(item, "topic"))
      },
      {
        key: "reviews",
        title: t("academicReviewTitle"),
        render: (item) => {
          const count = item.completed_dates.length;
          return academicActivityItem(
            state.lang === "ko" ? `${count}편` : `${count} paper${count === 1 ? "" : "s"}`,
            p(item, "journal")
          );
        }
      }
    ];

    $("#academic_activities-list").innerHTML = groups
      .filter((group) => data[group.key].length > 0)
      .map((group) => `
        <section class="academic-activity-group">
          <h4>${D.escapeHTML(group.title)}</h4>
          <div class="academic-activity-list">
            ${data[group.key].map(group.render).join("")}
          </div>
        </section>
      `).join("");
  }

  function renderSkills() {
    $("#skills-list").innerHTML = state.data.profile.skills.map((item) => `
      <article class="skill-item">
        <strong>${D.escapeHTML(item.name)}</strong>
        <p>${D.escapeHTML(item[`detail_${state.lang}`] || item.detail_en)}</p>
      </article>
    `).join("");
  }

  function applyMainPageSectionOrder() {
    const visibleOrder = state.data.settings.main_page_sections;
    const hiddenOrder = state.data.settings.hidden_main_page_sections || [];
    const order = Array.isArray(visibleOrder) && Array.isArray(hiddenOrder)
      ? [...visibleOrder, ...hiddenOrder]
      : [];
    const valid = Array.isArray(visibleOrder)
      && Array.isArray(hiddenOrder)
      && order.length === MAIN_PAGE_SECTIONS.length
      && order.every((section) => MAIN_PAGE_SECTIONS.includes(section))
      && new Set(order).size === MAIN_PAGE_SECTIONS.length;
    if (!valid) {
      const expected = MAIN_PAGE_SECTIONS.join(", ");
      throw new Error(state.lang === "ko"
        ? `settings.json의 main_page_sections와 hidden_main_page_sections에는 다음 항목이 각각 정확히 한 번씩 있어야 합니다: ${expected}`
        : `settings.json main_page_sections and hidden_main_page_sections must together contain each supported section exactly once: ${expected}`);
    }

    const sectionNodes = new Map(
      [...document.querySelectorAll("[data-main-section]")]
        .map((section) => [section.dataset.mainSection, section])
    );
    if (
      sectionNodes.size !== MAIN_PAGE_SECTIONS.length
      || MAIN_PAGE_SECTIONS.some((section) => !sectionNodes.has(section))
    ) {
      throw new Error("Main-page section markup does not match settings.json.");
    }
    const container = sectionNodes.values().next().value.parentElement;
    order.forEach((section) => container.append(sectionNodes.get(section)));
  }

  function updateMainPageSectionVisibility() {
    const hiddenSections = new Set(state.data.settings.hidden_main_page_sections || []);
    MAIN_PAGE_SECTIONS.forEach((sectionName) => {
      const section = document.querySelector(`[data-main-section="${sectionName}"]`);
      const list = section?.querySelector(`#${sectionName}-list`);
      if (!list) throw new Error(`Main-page section '${sectionName}' is missing its list.`);
      section.hidden = hiddenSections.has(sectionName) || list.children.length === 0;
    });
  }

  function renderAbout() {
    applyMainPageSectionOrder();
    renderContact();
    renderExperience();
    renderEducation();
    renderScholarships();
    renderCertifications();
    renderAwards();
    renderTeaching();
    renderSkills();
    renderAcademicActivities();
    renderPatents();
    updateMainPageSectionVisibility();
  }

  function renderAll() {
    updateProfileBindings();
    renderProjects();
    renderSoftware();
    renderPublications();
    renderBoard();
    renderAbout();
  }

  function pageIsVisible(page) {
    return page === "home" || Boolean(state.data && D.webPageVisible(state.data.settings, page));
  }

  function currentPageView() {
    const hash = window.location.hash.slice(1);
    return PAGE_VIEWS.has(hash) && pageIsVisible(hash) ? hash : "home";
  }

  function navigatePageBySwipe(direction) {
    const pages = PAGE_SEQUENCE.filter(pageIsVisible);
    const currentIndex = pages.indexOf(currentPageView());
    const nextIndex = Math.max(0, Math.min(currentIndex + direction, pages.length - 1));
    if (nextIndex === currentIndex) return;

    const nextView = pages[nextIndex];
    state.pageTransitionDirection = direction;
    if (nextView === "home") {
      window.location.hash = "";
    } else {
      window.location.hash = nextView;
    }
  }

  function animatePageEntry(section) {
    const direction = state.pageTransitionDirection;
    state.pageTransitionDirection = 0;
    if (!section || !direction || window.matchMedia("(prefers-reduced-motion: reduce)").matches) return;

    const className = direction > 0 ? "page-swipe-enter-next" : "page-swipe-enter-previous";
    section.classList.add(className);
    section.addEventListener("animationend", () => section.classList.remove(className), { once: true });
  }

  function renderPageView(options = {}) {
    const view = currentPageView();
    const requestedPage = window.location.hash.slice(1);
    if (state.data && PAGE_VIEWS.has(requestedPage) && !pageIsVisible(requestedPage)) {
      const url = new URL(window.location.href);
      url.hash = "";
      window.history.replaceState(window.history.state, "", url);
    }
    let activeSection = null;
    $$('[data-page-section]').forEach((section) => {
      section.classList.remove("page-swipe-enter-next", "page-swipe-enter-previous");
      section.hidden = section.dataset.pageSection !== view || !pageIsVisible(section.dataset.pageSection);
      if (!section.hidden) activeSection = section;
    });
    $$('[data-page-link]').forEach((link) => {
      link.hidden = !pageIsVisible(link.dataset.pageLink);
      if (link.dataset.pageLink === view) {
        link.setAttribute("aria-current", "page");
      } else {
        link.removeAttribute("aria-current");
      }
    });
    $(".primary-nav").hidden = Boolean(state.data && !PAGE_SEQUENCE.slice(1).some(pageIsVisible));
    alignMobileBrand();

    if (options.scroll !== false) {
      window.requestAnimationFrame(() => window.scrollTo({ top: 0, behavior: "auto" }));
    }
    if (view === "projects") {
      window.requestAnimationFrame(scrollProjectTimelineToLatest);
    } else {
      hideProjectTimelineTooltip();
    }
    if (view === "publications") {
      window.requestAnimationFrame(scrollPublicationTimelineToLatest);
    } else {
      hidePublicationTimelineTooltip();
      hidePublicationActionTooltip();
      hidePublicationCitationTooltip();
    }
    if (view === "board") {
      window.requestAnimationFrame(updateActiveBoardEntry);
    }
    animatePageEntry(activeSection);
  }

  function bindSwipeNavigation() {
    let gesture = null;

    document.addEventListener("touchstart", (event) => {
      if (event.touches.length !== 1) {
        gesture = null;
        return;
      }

      const touch = event.touches[0];
      const target = event.target;
      const gallery = target.closest?.("[data-media-gallery]");
      const projectTimeline = target.closest?.("[data-project-timeline-scroll], [data-publication-timeline-scroll], .publication-timeline-legend");
      let mode = "page";
      if (projectTimeline) {
        mode = "ignore";
      } else if (gallery) {
        mode = "gallery";
      } else if (target.closest?.("#lightbox-video")) {
        mode = "ignore";
      } else if ($("#media-lightbox")?.open || target.closest?.("#media-lightbox")) {
        mode = "lightbox";
      } else if (target.closest?.("input, textarea, select, [contenteditable='true']")) {
        mode = "ignore";
      }

      gesture = {
        mode,
        gallery,
        galleryIndex: gallery ? Number(gallery.dataset.galleryIndex || 0) : 0,
        startX: touch.clientX,
        startY: touch.clientY
      };
    }, { passive: true });

    document.addEventListener("touchend", (event) => {
      if (!gesture || event.changedTouches.length !== 1) {
        gesture = null;
        return;
      }

      const touch = event.changedTouches[0];
      const deltaX = touch.clientX - gesture.startX;
      const deltaY = touch.clientY - gesture.startY;
      const isHorizontalSwipe = Math.abs(deltaX) >= SWIPE_MIN_DISTANCE
        && Math.abs(deltaX) > Math.abs(deltaY) * SWIPE_DIRECTION_RATIO;
      const mode = gesture.mode;
      const gallery = gesture.gallery;
      const galleryIndex = gesture.galleryIndex;
      gesture = null;

      if (!isHorizontalSwipe || mode === "ignore") return;
      const direction = deltaX < 0 ? 1 : -1;

      if (mode === "gallery") {
        if (event.cancelable) event.preventDefault();
        setGalleryIndex(gallery, galleryIndex + direction);
        return;
      }

      if (mode === "lightbox") {
        if (event.cancelable) event.preventDefault();
        stepLightbox(direction);
        return;
      }

      if (window.matchMedia("(max-width: 900px)").matches) {
        if (event.cancelable) event.preventDefault();
        navigatePageBySwipe(direction);
      }
    }, { passive: false });

    document.addEventListener("touchmove", (event) => {
      if (!gesture || gesture.mode === "gallery" || gesture.mode === "ignore" || event.touches.length !== 1) return;
      if (gesture.mode === "page" && !window.matchMedia("(max-width: 900px)").matches) return;
      const touch = event.touches[0];
      const deltaX = touch.clientX - gesture.startX;
      const deltaY = touch.clientY - gesture.startY;
      if (Math.abs(deltaX) > 12 && Math.abs(deltaX) > Math.abs(deltaY) * SWIPE_DIRECTION_RATIO) {
        event.preventDefault();
      }
    }, { passive: false });

    document.addEventListener("touchcancel", () => {
      gesture = null;
    }, { passive: true });
  }

  function showToast(message) {
    const toast = $("#toast");
    toast.textContent = message;
    toast.hidden = false;
    window.clearTimeout(showToast.timer);
    showToast.timer = window.setTimeout(() => {
      toast.hidden = true;
    }, 2300);
  }

  async function copyText(text) {
    try {
      await navigator.clipboard.writeText(text);
      showToast(t("copied"));
    } catch {
      const textarea = document.createElement("textarea");
      textarea.value = text;
      textarea.style.position = "fixed";
      textarea.style.opacity = "0";
      document.body.appendChild(textarea);
      textarea.select();
      const success = document.execCommand("copy");
      textarea.remove();
      showToast(success ? t("copied") : t("copyFailed"));
    }
  }

  function bindEvents() {
    let publicationPointer = null;
    document.addEventListener("pointerdown", (event) => {
      const summary = event.target.closest?.("[data-publication-toggle]");
      publicationPointer = event.isPrimary && event.button === 0 && summary ? {
        pointerId: event.pointerId,
        startX: event.clientX,
        startY: event.clientY,
        selectionAtStart: window.getSelection()?.toString() || "",
        dragged: false
      } : null;
    });
    const trackPublicationDrag = (event) => {
      if (!publicationPointer || event.pointerId !== publicationPointer.pointerId) return;
      if (Math.hypot(event.clientX - publicationPointer.startX, event.clientY - publicationPointer.startY) > 5) {
        publicationPointer.dragged = true;
      }
    };
    document.addEventListener("pointermove", trackPublicationDrag);
    document.addEventListener("pointerup", trackPublicationDrag);
    document.addEventListener("pointercancel", () => { publicationPointer = null; });

    document.addEventListener("pointerover", (event) => {
      const mark = event.target.closest?.(".publication-timeline-mark");
      if (mark) showPublicationTimelineTooltip(mark);
      const bar = event.target.closest?.(".project-roadmap-bar");
      if (bar) showProjectTimelineTooltip(bar);
      const action = event.target.closest?.("[data-copy-citation], [data-export-one]");
      if (action && event.pointerType !== "touch") showPublicationActionTooltip(action);
      const count = event.target.closest?.("[data-citation-source]");
      if (count && event.pointerType !== "touch") showPublicationCitationTooltip(count);
      if (event.target.closest?.("#publication-action-tooltip")) window.clearTimeout(hidePublicationActionTooltip.timer);
    });
    document.addEventListener("pointerout", (event) => {
      if (event.target.closest?.(".publication-timeline-mark") && !event.relatedTarget?.closest?.(".publication-timeline-mark")) hidePublicationTimelineTooltip();
      if (event.target.closest?.(".project-roadmap-bar") && !event.relatedTarget?.closest?.(".project-roadmap-bar")) hideProjectTimelineTooltip();
      if (event.target.closest?.("[data-copy-citation], [data-export-one], #publication-action-tooltip")) scheduleHidePublicationActionTooltip();
      if (event.target.closest?.("[data-citation-source]") && !event.relatedTarget?.closest?.("[data-citation-source]")) hidePublicationCitationTooltip();
    });
    document.addEventListener("focusin", (event) => {
      const mark = event.target.closest?.(".publication-timeline-mark");
      if (mark) showPublicationTimelineTooltip(mark);
      const bar = event.target.closest?.(".project-roadmap-bar");
      if (bar) showProjectTimelineTooltip(bar);
      const action = event.target.closest?.("[data-copy-citation], [data-export-one]");
      if (action) showPublicationActionTooltip(action);
      const count = event.target.closest?.("[data-citation-source]");
      if (count) showPublicationCitationTooltip(count);
    });
    document.addEventListener("focusout", (event) => {
      if (event.target.closest?.(".publication-timeline-mark")) hidePublicationTimelineTooltip();
      if (event.target.closest?.(".project-roadmap-bar")) hideProjectTimelineTooltip();
      if (event.target.closest?.("[data-copy-citation], [data-export-one]")) scheduleHidePublicationActionTooltip();
      if (event.target.closest?.("[data-citation-source]")) hidePublicationCitationTooltip();
    });
    document.addEventListener("scroll", (event) => {
      hidePublicationCitationTooltip();
      hidePublicationTimelineTooltip();
      hideProjectTimelineTooltip();
      if (!$("#publication-action-tooltip")?.contains(event.target)) hidePublicationActionTooltip();
    }, true);
    window.addEventListener("resize", () => {
      hidePublicationCitationTooltip();
      hidePublicationTimelineTooltip();
      hideProjectTimelineTooltip();
      hidePublicationActionTooltip();
      const canvas = $(".publication-timeline-canvas");
      if (canvas) canvas.style.setProperty("--publication-type-width", `${publicationTimelineLabelWidth()}px`);
    });
    $$("[data-lang-switch]").forEach((button) => {
      button.addEventListener("click", () => setLanguage(button.dataset.langSwitch));
    });
    window.addEventListener("storage", (event) => {
      if (event.key === "siteLanguage" && ["en", "ko"].includes(event.newValue) && event.newValue !== state.lang) setLanguage(event.newValue);
    });

    document.addEventListener("click", (event) => {
      const publicationTarget = event.target.closest("[data-publication-timeline-target]");
      if (publicationTarget) {
        scrollToPublication(Number(publicationTarget.dataset.publicationTimelineTarget));
        return;
      }
      const boardTarget = event.target.closest("[data-board-target]");
      if (boardTarget) {
        scrollToBoardEntry(Number(boardTarget.dataset.boardTarget));
        return;
      }

      const boardPage = event.target.closest("[data-board-page]");
      if (boardPage) {
        state.board.page += Number(boardPage.dataset.boardPage);
        renderBoardList(boardRecords());
        return;
      }

      const projectTarget = event.target.closest("[data-project-timeline-target]");
      if (projectTarget) {
        scrollToProjectCard(Number(projectTarget.dataset.projectTimelineTarget));
        return;
      }

      const galleryStep = event.target.closest("[data-gallery-step]");
      if (galleryStep) {
        const gallery = galleryStep.closest("[data-media-gallery]");
        const index = Number(gallery?.dataset.galleryIndex || 0) + Number(galleryStep.dataset.galleryStep);
        if (gallery) setGalleryIndex(gallery, index);
        return;
      }

      const mediaButton = event.target.closest("[data-open-lightbox]");
      if (mediaButton) {
        openLightbox(mediaButton.dataset.lightboxCollection, Number(mediaButton.dataset.lightboxItemIndex || 0), Number(mediaButton.dataset.mediaIndex || 0));
        return;
      }

      if (event.target.closest("[data-lightbox-close]")) {
        $("#media-lightbox").close();
        return;
      }

      if (event.target.closest("[data-lightbox-prev]")) {
        stepLightbox(-1);
        return;
      }

      if (event.target.closest("[data-lightbox-next]")) {
        stepLightbox(1);
        return;
      }

      const sortButton = event.target.closest("[data-pub-sort]");
      if (sortButton) {
        state.publication.sort = sortButton.dataset.pubSort;
        renderPublications();
        return;
      }

      const collaboratorButton = event.target.closest("[data-collaborator-id]");
      if (collaboratorButton) {
        const id = collaboratorButton.dataset.collaboratorId;
        state.publication.collaboratorId = state.publication.collaboratorId === id ? "" : id;
        renderPublications();
        return;
      }

      const publicationToggle = event.target.closest("[data-publication-toggle]");
      if (publicationToggle) {
        const pointer = publicationPointer;
        publicationPointer = null;
        // Keyboard activation still toggles even when text is selected.
        if (event.detail > 0) {
          const selection = window.getSelection();
          const selectingText = pointer && selection && !selection.isCollapsed
            && selection.toString() !== pointer.selectionAtStart
            && Array.from({ length: selection.rangeCount }, (_, index) => selection.getRangeAt(index))
              .some((range) => range.intersectsNode(publicationToggle));
          if (pointer?.dragged || selectingText) return;
        }
        const itemIndex = Number(publicationToggle.dataset.publicationToggle);
        const article = publicationToggle.closest("[data-publication-index]");
        const details = article?.querySelector(".publication-details");
        const expanded = !state.publication.expanded.has(itemIndex);
        if (expanded) state.publication.expanded.add(itemIndex);
        else state.publication.expanded.delete(itemIndex);
        article?.classList.toggle("is-expanded", expanded);
        publicationToggle.setAttribute("aria-expanded", String(expanded));
        details?.setAttribute("aria-hidden", String(!expanded));
        return;
      }

      const awardToggle = event.target.closest("[data-award-toggle]");
      if (awardToggle) {
        const awardId = awardToggle.dataset.awardToggle;
        const article = awardToggle.closest("[data-award-id]");
        const details = article?.querySelector(".award-details");
        const expanded = !state.award.expanded.has(awardId);
        if (expanded) state.award.expanded.add(awardId);
        else state.award.expanded.delete(awardId);
        article?.classList.toggle("is-expanded", expanded);
        awardToggle.setAttribute("aria-expanded", String(expanded));
        details?.setAttribute("aria-hidden", String(!expanded));
        return;
      }

      const copyButton = event.target.closest("[data-copy-citation]");
      if (copyButton) {
        const item = state.data.publications[Number(copyButton.dataset.copyCitation)];
        if (item) copyText(D.publicationCitation(item, state.lang));
        return;
      }

      const oneExport = event.target.closest("[data-export-one]");
      if (oneExport) {
        const item = state.data.publications[Number(oneExport.dataset.exportOne)];
        if (item) D.downloadPublicationBib(item);
      }
    });

    $("#pub-search").addEventListener("input", (event) => {
      state.publication.query = event.target.value;
      renderPublications();
      event.target.focus();
      event.target.setSelectionRange(event.target.value.length, event.target.value.length);
    });
    $("#pub-type").addEventListener("change", (event) => {
      state.publication.publicationType = event.target.value;
      renderPublications();
    });
    $("#pub-topic").addEventListener("change", (event) => {
      state.publication.topic = event.target.value;
      renderPublications();
    });
    $("#pub-index").addEventListener("change", (event) => {
      state.publication.indexing = event.target.value;
      renderPublications();
    });
    $("#pub-reset").addEventListener("click", () => {
      state.publication.query = "";
      state.publication.publicationType = "all";
      state.publication.indexing = "all";
      state.publication.topic = ALL_TOPIC_FILTER;
      state.publication.collaboratorId = "";
      state.publication.firstAuthorOnly = false;
      renderPublications();
    });
    $("#pub-first-author").addEventListener("change", (event) => {
      state.publication.firstAuthorOnly = event.target.checked;
      renderPublications();
    });

    document.addEventListener("keydown", (event) => {
      if (event.key === "Escape") {
        hidePublicationCitationTooltip();
        hidePublicationTimelineTooltip();
        hideProjectTimelineTooltip();
        hidePublicationActionTooltip();
      }
      const dialog = $("#media-lightbox");
      if (dialog.open) {
        if (event.key === "ArrowLeft" || event.key === "ArrowRight") {
          event.preventDefault();
          stepLightbox(event.key === "ArrowLeft" ? -1 : 1);
        }
        return;
      }

      const viewport = event.target.closest?.("[data-gallery-viewport]");
      if (viewport && (event.key === "ArrowLeft" || event.key === "ArrowRight")) {
        event.preventDefault();
        const gallery = viewport.closest("[data-media-gallery]");
        const index = Number(gallery.dataset.galleryIndex || 0) + (event.key === "ArrowLeft" ? -1 : 1);
        setGalleryIndex(gallery, index);
      }
    });

    $("#media-lightbox").addEventListener("click", (event) => {
      if (!event.target.closest(".lightbox-figure, .lightbox-nav, .lightbox-close")) {
        $("#media-lightbox").close();
      }
    });

    $("#media-lightbox").addEventListener("close", () => {
      $("#lightbox-image").removeAttribute("src");
      const video = $("#lightbox-video");
      video.pause();
      video.removeAttribute("src");
      video.removeAttribute("poster");
      video.load();
    });

    window.addEventListener("resize", () => {
      alignMobileBrand();
      $$('[data-media-gallery]').forEach((gallery) => {
        setGalleryIndex(gallery, Number(gallery.dataset.galleryIndex || 0), { instant: true });
      });
      scrollProjectTimelineToLatest();
      scheduleBoardActiveUpdate();
    });
    window.addEventListener("scroll", scheduleBoardActiveUpdate, { passive: true });
  }

  function renderLoadError(error) {
    console.error(error);
    const main = $("#main");
    const reason = error instanceof Error ? error.message : String(error || "Unknown error");
    main.innerHTML = `
      <section class="section-shell">
        <div class="container">
          <div class="empty-state load-error" role="alert">
            <p class="load-error-message">${D.escapeHTML(t("loadError"))}</p>
            <p class="load-error-detail">
              <strong>${D.escapeHTML(t("loadErrorReason"))}:</strong>
              <code>${D.escapeHTML(reason)}</code>
            </p>
          </div>
        </div>
      </section>
    `;
  }

  async function init() {
    state.lang = initialLanguage();
    setLanguage(state.lang, { skipURL: true });
    bindEvents();
    document.fonts?.ready.then(alignMobileBrand);
    if (window.ResizeObserver) {
      new ResizeObserver(alignMobileBrand).observe($(".primary-nav"));
    }
    bindSwipeNavigation();
    document.addEventListener("selectionchange", updatePublicationCitationOrder);
    document.addEventListener("pointerup", () => window.setTimeout(updatePublicationCitationOrder, 0));
    document.addEventListener("pointercancel", updatePublicationCitationOrder);
    window.addEventListener("hashchange", () => renderPageView());
    renderPageView({ scroll: false });

    try {
      state.data = await D.loadAll();
      D.applyColorTheme(state.data.settings);
      setLanguage(state.lang, { skipURL: true });
      renderPageView({ scroll: false });
      // Run after a paint; citation services are never part of the initial data load.
      if (pageIsVisible("publications")) {
        window.requestAnimationFrame(() => window.setTimeout(() => {
          void citations.start(state.data.publications, updatePublicationCitations);
        }, 0));
      }
    } catch (error) {
      renderLoadError(error);
    }
  }

  init();
}());
