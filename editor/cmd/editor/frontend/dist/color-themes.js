(function (root) {
  "use strict";

  const presetData = [
    {
      "id": "sage",
      "label": "세이지",
      "description": "차분한 녹색 · 기본 테마",
      "colors": {
        "bg": "#f7f8f5",
        "surface": "#ffffff",
        "surface-muted": "#eef2ec",
        "surface-strong": "#e2e9df",
        "ink": "#16231d",
        "ink-soft": "#52675c",
        "ink-faint": "#74837b",
        "line": "#cbd5cd",
        "line-strong": "#9eada4",
        "accent": "#305744",
        "accent-strong": "#183c2d",
        "accent-soft": "#d8e3d7",
        "focus": "#0c68c7",
        "danger": "#8c2f26",
        "on-accent": "#ffffff",
        "on-danger": "#ffffff",
        "on-ink": "#ffffff",
        "warning": "#653d0c",
        "warning-soft": "#f3e5c5",
        "cv-bg": "#edf0ec",
        "danger-strong": "#6f211b",
        "danger-soft": "#fbeae7",
        "danger-line": "#d9a59d",
        "disabled": "#7c8d83",
        "info": "#0369a1",
        "info-soft": "#e0f2fe"
      },
      "dark": false
    },
    {
      "id": "ocean",
      "label": "오션",
      "description": "시원하고 선명한 블루",
      "colors": {
        "bg": "#f4f8fc",
        "surface": "#ffffff",
        "surface-muted": "#ebf2f8",
        "surface-strong": "#dae7f2",
        "ink": "#142b40",
        "ink-soft": "#506b80",
        "ink-faint": "#506b80",
        "line": "#c8d8e7",
        "line-strong": "#94afc6",
        "accent": "#245d87",
        "accent-strong": "#164465",
        "accent-soft": "#dcecf8",
        "focus": "#0c68c7",
        "danger": "#8c2f26",
        "on-accent": "#ffffff",
        "on-danger": "#ffffff",
        "on-ink": "#ffffff",
        "warning": "#653d0c",
        "warning-soft": "#f3e5c5",
        "cv-bg": "#ebf2f8",
        "danger-strong": "#6f211b",
        "danger-soft": "#fbeae7",
        "danger-line": "#d9a59d",
        "disabled": "#70889a",
        "info": "#0369a1",
        "info-soft": "#e0f2fe"
      },
      "dark": false
    },
    {
      "id": "violet",
      "label": "바이올렛",
      "description": "부드러운 보라색",
      "colors": {
        "bg": "#f8f6fc",
        "surface": "#ffffff",
        "surface-muted": "#f0ebf7",
        "surface-strong": "#e5ddf0",
        "ink": "#2b2140",
        "ink-soft": "#6b5a7d",
        "ink-faint": "#6b5a7d",
        "line": "#d8cfe5",
        "line-strong": "#b3a1c8",
        "accent": "#6c478e",
        "accent-strong": "#4f2e70",
        "accent-soft": "#ebdff5",
        "focus": "#6f42c1",
        "danger": "#8c2f26",
        "on-accent": "#ffffff",
        "on-danger": "#ffffff",
        "on-ink": "#ffffff",
        "warning": "#653d0c",
        "warning-soft": "#f3e5c5",
        "cv-bg": "#f0ebf7",
        "danger-strong": "#6f211b",
        "danger-soft": "#fbeae7",
        "danger-line": "#d9a59d",
        "disabled": "#8b7b9b",
        "info": "#0369a1",
        "info-soft": "#e0f2fe"
      },
      "dark": false
    },
    {
      "id": "rose",
      "label": "로즈",
      "description": "은은하고 따뜻한 장미색",
      "colors": {
        "bg": "#fcf6f7",
        "surface": "#ffffff",
        "surface-muted": "#f7edef",
        "surface-strong": "#f0dfe4",
        "ink": "#3a222c",
        "ink-soft": "#7b5966",
        "ink-faint": "#7b5966",
        "line": "#e4ccd4",
        "line-strong": "#c39ca9",
        "accent": "#93445e",
        "accent-strong": "#712c42",
        "accent-soft": "#f5dee6",
        "focus": "#a52f60",
        "danger": "#8c2f26",
        "on-accent": "#ffffff",
        "on-danger": "#ffffff",
        "on-ink": "#ffffff",
        "warning": "#653d0c",
        "warning-soft": "#f3e5c5",
        "cv-bg": "#f7edef",
        "danger-strong": "#6f211b",
        "danger-soft": "#fbeae7",
        "danger-line": "#d9a59d",
        "disabled": "#9d7c88",
        "info": "#0369a1",
        "info-soft": "#e0f2fe"
      },
      "dark": false
    },
    {
      "id": "sand",
      "label": "샌드",
      "description": "편안한 베이지와 브라운",
      "colors": {
        "bg": "#faf8f3",
        "surface": "#ffffff",
        "surface-muted": "#f3eee3",
        "surface-strong": "#e9dfcd",
        "ink": "#332b1c",
        "ink-soft": "#75654b",
        "ink-faint": "#75654b",
        "line": "#ddcfb6",
        "line-strong": "#bba883",
        "accent": "#82612c",
        "accent-strong": "#60451c",
        "accent-soft": "#eee1c5",
        "focus": "#9c6a19",
        "danger": "#8c2f26",
        "on-accent": "#ffffff",
        "on-danger": "#ffffff",
        "on-ink": "#ffffff",
        "warning": "#653d0c",
        "warning-soft": "#f3e5c5",
        "cv-bg": "#f3eee3",
        "danger-strong": "#6f211b",
        "danger-soft": "#fbeae7",
        "danger-line": "#d9a59d",
        "disabled": "#9a8b70",
        "info": "#0369a1",
        "info-soft": "#e0f2fe"
      },
      "dark": false
    },
    {
      "id": "slate",
      "label": "슬레이트",
      "description": "짙은 배경의 다크 테마",
      "colors": {
        "bg": "#111827",
        "surface": "#1b2535",
        "surface-muted": "#243144",
        "surface-strong": "#303e53",
        "ink": "#edf2fa",
        "ink-soft": "#afbdd0",
        "ink-faint": "#afbdd0",
        "line": "#3c4b61",
        "line-strong": "#60738e",
        "accent": "#9bb9f0",
        "accent-strong": "#bed2f7",
        "accent-soft": "#2a3e61",
        "focus": "#91bcff",
        "danger": "#ffaaa3",
        "on-accent": "#16243c",
        "on-danger": "#3d1717",
        "on-ink": "#111827",
        "warning": "#f0ca7b",
        "warning-soft": "#3d3320",
        "cv-bg": "#243144",
        "danger-strong": "#ffc6c0",
        "danger-soft": "#45282c",
        "danger-line": "#865158",
        "disabled": "#9aa9be",
        "info": "#9bd4f7",
        "info-soft": "#1d3b50"
      },
      "dark": true
    },
    {
      "id": "mint",
      "label": "민트",
      "description": "산뜻하고 부드러운 민트색",
      "colors": {
        "bg": "#f3faf6",
        "surface": "#ffffff",
        "surface-muted": "#e7f5ec",
        "surface-strong": "#d1eadd",
        "ink": "#173d2a",
        "ink-soft": "#4c6f5b",
        "ink-faint": "#4c6f5b",
        "line": "#bfd8c9",
        "line-strong": "#85b49a",
        "accent": "#276747",
        "accent-strong": "#194e32",
        "accent-soft": "#d4efdf",
        "focus": "#176f42",
        "danger": "#8c2f26",
        "on-accent": "#ffffff",
        "on-danger": "#ffffff",
        "on-ink": "#ffffff",
        "warning": "#653d0c",
        "warning-soft": "#f3e5c5",
        "cv-bg": "#e7f5ec",
        "danger-strong": "#6f211b",
        "danger-soft": "#fbeae7",
        "danger-line": "#d9a59d",
        "disabled": "#739584",
        "info": "#0369a1",
        "info-soft": "#e0f2fe"
      },
      "dark": false
    },
    {
      "id": "teal",
      "label": "틸",
      "description": "깊이 있는 청록색",
      "colors": {
        "bg": "#f2faf9",
        "surface": "#ffffff",
        "surface-muted": "#e3f3ef",
        "surface-strong": "#cce7e0",
        "ink": "#173b37",
        "ink-soft": "#4d706b",
        "ink-faint": "#4d706b",
        "line": "#b9d8d0",
        "line-strong": "#7faea3",
        "accent": "#176c62",
        "accent-strong": "#0e514a",
        "accent-soft": "#ceeae4",
        "focus": "#007d72",
        "danger": "#8c2f26",
        "on-accent": "#ffffff",
        "on-danger": "#ffffff",
        "on-ink": "#ffffff",
        "warning": "#653d0c",
        "warning-soft": "#f3e5c5",
        "cv-bg": "#e3f3ef",
        "danger-strong": "#6f211b",
        "danger-soft": "#fbeae7",
        "danger-line": "#d9a59d",
        "disabled": "#71938b",
        "info": "#0369a1",
        "info-soft": "#e0f2fe"
      },
      "dark": false
    },
    {
      "id": "sky",
      "label": "스카이",
      "description": "맑고 가벼운 하늘색",
      "colors": {
        "bg": "#f2faff",
        "surface": "#ffffff",
        "surface-muted": "#e4f3fc",
        "surface-strong": "#cce7f7",
        "ink": "#17354b",
        "ink-soft": "#4b6d84",
        "ink-faint": "#4b6d84",
        "line": "#bdd8eb",
        "line-strong": "#83aecb",
        "accent": "#176b9b",
        "accent-strong": "#0d5178",
        "accent-soft": "#d1eafa",
        "focus": "#007cb5",
        "danger": "#8c2f26",
        "on-accent": "#ffffff",
        "on-danger": "#ffffff",
        "on-ink": "#ffffff",
        "warning": "#653d0c",
        "warning-soft": "#f3e5c5",
        "cv-bg": "#e4f3fc",
        "danger-strong": "#6f211b",
        "danger-soft": "#fbeae7",
        "danger-line": "#d9a59d",
        "disabled": "#7495ab",
        "info": "#0369a1",
        "info-soft": "#e0f2fe"
      },
      "dark": false
    },
    {
      "id": "indigo",
      "label": "인디고",
      "description": "또렷한 남보라색",
      "colors": {
        "bg": "#f6f7ff",
        "surface": "#ffffff",
        "surface-muted": "#ebedfc",
        "surface-strong": "#daddf7",
        "ink": "#272c52",
        "ink-soft": "#5e648c",
        "ink-faint": "#5e648c",
        "line": "#cdd1ed",
        "line-strong": "#9da5d1",
        "accent": "#4d55a8",
        "accent-strong": "#363e89",
        "accent-soft": "#e0e3fc",
        "focus": "#4b50bd",
        "danger": "#8c2f26",
        "on-accent": "#ffffff",
        "on-danger": "#ffffff",
        "on-ink": "#ffffff",
        "warning": "#653d0c",
        "warning-soft": "#f3e5c5",
        "cv-bg": "#ebedfc",
        "danger-strong": "#6f211b",
        "danger-soft": "#fbeae7",
        "danger-line": "#d9a59d",
        "disabled": "#828ab0",
        "info": "#0369a1",
        "info-soft": "#e0f2fe"
      },
      "dark": false
    },
    {
      "id": "amber",
      "label": "앰버",
      "description": "따뜻한 호박색과 오렌지",
      "colors": {
        "bg": "#fffbf1",
        "surface": "#ffffff",
        "surface-muted": "#fff2d4",
        "surface-strong": "#f5e3b4",
        "ink": "#493514",
        "ink-soft": "#7f6842",
        "ink-faint": "#7f6842",
        "line": "#dfcc9f",
        "line-strong": "#c4a165",
        "accent": "#925d13",
        "accent-strong": "#714609",
        "accent-soft": "#f9e7b8",
        "focus": "#9c6209",
        "danger": "#8c2f26",
        "on-accent": "#ffffff",
        "on-danger": "#ffffff",
        "on-ink": "#ffffff",
        "warning": "#653d0c",
        "warning-soft": "#f3e5c5",
        "cv-bg": "#fff2d4",
        "danger-strong": "#6f211b",
        "danger-soft": "#fbeae7",
        "danger-line": "#d9a59d",
        "disabled": "#a18e67",
        "info": "#0369a1",
        "info-soft": "#e0f2fe"
      },
      "dark": false
    },
    {
      "id": "graphite",
      "label": "그래파이트",
      "description": "무채색 중심의 다크 테마",
      "colors": {
        "bg": "#171717",
        "surface": "#232323",
        "surface-muted": "#2d2d2d",
        "surface-strong": "#393939",
        "ink": "#f3f3f3",
        "ink-soft": "#bcbcbc",
        "ink-faint": "#bcbcbc",
        "line": "#474747",
        "line-strong": "#6c6c6c",
        "accent": "#d2d2d2",
        "accent-strong": "#e9e9e9",
        "accent-soft": "#3d3d3d",
        "focus": "#a9c7ff",
        "danger": "#ffaaa3",
        "on-accent": "#242424",
        "on-danger": "#3d1717",
        "on-ink": "#171717",
        "warning": "#f0ca7b",
        "warning-soft": "#41351f",
        "cv-bg": "#2d2d2d",
        "danger-strong": "#ffc6c0",
        "danger-soft": "#482a29",
        "danger-line": "#8a5350",
        "disabled": "#a8a8a8",
        "info": "#a9d8f4",
        "info-soft": "#233c4d"
      },
      "dark": true
    }
  ];

  const websitePresetDerived = {
    "sage": {
      "surface-soft": "var(--surface-muted)",
      "header-tint": "color-mix(in srgb, var(--bg) 91%, transparent)",
      "cv-header-tint": "color-mix(in srgb, var(--bg) 94%, transparent)",
      "header-line": "color-mix(in srgb, var(--line-strong) 55%, transparent)",
      "section-line": "color-mix(in srgb, var(--line-strong) 42%, transparent)",
      "row-line": "color-mix(in srgb, var(--line) 62%, transparent)",
      "timeline-grid": "color-mix(in srgb, var(--line-strong) 34%, transparent)",
      "page-grid": "color-mix(in srgb, var(--ink) 1.8%, transparent)",
      "empty-bg": "color-mix(in srgb, var(--surface) 42%, transparent)",
      "shadow-rgb": "25 45 35",
      "ink-shadow-rgb": "20 36 29",
      "cv-shadow-rgb": "23 40 32",
      "media-shadow-rgb": "10 20 15",
      "lightbox-bg": "rgb(7 12 10 / 96%)",
      "lightbox-backdrop": "rgb(7 12 10 / 82%)",
      "lightbox-ink": "#ffffff",
      "lightbox-ink-soft": "rgb(255 255 255 / 82%)",
      "lightbox-ink-faint": "rgb(255 255 255 / 68%)",
      "lightbox-line": "rgb(255 255 255 / 32%)",
      "lightbox-control-bg": "rgb(255 255 255 / 8%)",
      "lightbox-control-hover": "rgb(255 255 255 / 17%)",
      "media-control-bg": "rgb(16 30 23 / 72%)",
      "media-control-hover": "rgb(16 30 23 / 90%)",
      "video-bg": "#08100c",
      "shadow": "0 18px 50px rgb(var(--shadow-rgb) / 8%)"
    },
    "ocean": {
      "surface-soft": "var(--surface-muted)",
      "header-tint": "color-mix(in srgb, var(--bg) 91%, transparent)",
      "cv-header-tint": "color-mix(in srgb, var(--bg) 94%, transparent)",
      "header-line": "color-mix(in srgb, var(--line-strong) 55%, transparent)",
      "section-line": "color-mix(in srgb, var(--line-strong) 42%, transparent)",
      "row-line": "color-mix(in srgb, var(--line) 62%, transparent)",
      "timeline-grid": "color-mix(in srgb, var(--line-strong) 34%, transparent)",
      "page-grid": "color-mix(in srgb, var(--ink) 1.8%, transparent)",
      "empty-bg": "color-mix(in srgb, var(--surface) 42%, transparent)",
      "shadow-rgb": "20 43 64",
      "ink-shadow-rgb": "20 43 64",
      "cv-shadow-rgb": "20 43 64",
      "media-shadow-rgb": "20 43 64",
      "lightbox-bg": "rgb(20 43 64 / 96%)",
      "lightbox-backdrop": "rgb(20 43 64 / 82%)",
      "lightbox-ink": "#ffffff",
      "lightbox-ink-soft": "rgb(255 255 255 / 82%)",
      "lightbox-ink-faint": "rgb(255 255 255 / 68%)",
      "lightbox-line": "rgb(255 255 255 / 32%)",
      "lightbox-control-bg": "rgb(255 255 255 / 8%)",
      "lightbox-control-hover": "rgb(255 255 255 / 17%)",
      "media-control-bg": "rgb(20 43 64 / 72%)",
      "media-control-hover": "rgb(20 43 64 / 90%)",
      "video-bg": "#164465",
      "shadow": "0 18px 50px rgb(var(--shadow-rgb) / 8%)"
    },
    "violet": {
      "surface-soft": "var(--surface-muted)",
      "header-tint": "color-mix(in srgb, var(--bg) 91%, transparent)",
      "cv-header-tint": "color-mix(in srgb, var(--bg) 94%, transparent)",
      "header-line": "color-mix(in srgb, var(--line-strong) 55%, transparent)",
      "section-line": "color-mix(in srgb, var(--line-strong) 42%, transparent)",
      "row-line": "color-mix(in srgb, var(--line) 62%, transparent)",
      "timeline-grid": "color-mix(in srgb, var(--line-strong) 34%, transparent)",
      "page-grid": "color-mix(in srgb, var(--ink) 1.8%, transparent)",
      "empty-bg": "color-mix(in srgb, var(--surface) 42%, transparent)",
      "shadow-rgb": "43 33 64",
      "ink-shadow-rgb": "43 33 64",
      "cv-shadow-rgb": "43 33 64",
      "media-shadow-rgb": "43 33 64",
      "lightbox-bg": "rgb(43 33 64 / 96%)",
      "lightbox-backdrop": "rgb(43 33 64 / 82%)",
      "lightbox-ink": "#ffffff",
      "lightbox-ink-soft": "rgb(255 255 255 / 82%)",
      "lightbox-ink-faint": "rgb(255 255 255 / 68%)",
      "lightbox-line": "rgb(255 255 255 / 32%)",
      "lightbox-control-bg": "rgb(255 255 255 / 8%)",
      "lightbox-control-hover": "rgb(255 255 255 / 17%)",
      "media-control-bg": "rgb(43 33 64 / 72%)",
      "media-control-hover": "rgb(43 33 64 / 90%)",
      "video-bg": "#4f2e70",
      "shadow": "0 18px 50px rgb(var(--shadow-rgb) / 8%)"
    },
    "rose": {
      "surface-soft": "var(--surface-muted)",
      "header-tint": "color-mix(in srgb, var(--bg) 91%, transparent)",
      "cv-header-tint": "color-mix(in srgb, var(--bg) 94%, transparent)",
      "header-line": "color-mix(in srgb, var(--line-strong) 55%, transparent)",
      "section-line": "color-mix(in srgb, var(--line-strong) 42%, transparent)",
      "row-line": "color-mix(in srgb, var(--line) 62%, transparent)",
      "timeline-grid": "color-mix(in srgb, var(--line-strong) 34%, transparent)",
      "page-grid": "color-mix(in srgb, var(--ink) 1.8%, transparent)",
      "empty-bg": "color-mix(in srgb, var(--surface) 42%, transparent)",
      "shadow-rgb": "58 34 44",
      "ink-shadow-rgb": "58 34 44",
      "cv-shadow-rgb": "58 34 44",
      "media-shadow-rgb": "58 34 44",
      "lightbox-bg": "rgb(58 34 44 / 96%)",
      "lightbox-backdrop": "rgb(58 34 44 / 82%)",
      "lightbox-ink": "#ffffff",
      "lightbox-ink-soft": "rgb(255 255 255 / 82%)",
      "lightbox-ink-faint": "rgb(255 255 255 / 68%)",
      "lightbox-line": "rgb(255 255 255 / 32%)",
      "lightbox-control-bg": "rgb(255 255 255 / 8%)",
      "lightbox-control-hover": "rgb(255 255 255 / 17%)",
      "media-control-bg": "rgb(58 34 44 / 72%)",
      "media-control-hover": "rgb(58 34 44 / 90%)",
      "video-bg": "#712c42",
      "shadow": "0 18px 50px rgb(var(--shadow-rgb) / 8%)"
    },
    "sand": {
      "surface-soft": "var(--surface-muted)",
      "header-tint": "color-mix(in srgb, var(--bg) 91%, transparent)",
      "cv-header-tint": "color-mix(in srgb, var(--bg) 94%, transparent)",
      "header-line": "color-mix(in srgb, var(--line-strong) 55%, transparent)",
      "section-line": "color-mix(in srgb, var(--line-strong) 42%, transparent)",
      "row-line": "color-mix(in srgb, var(--line) 62%, transparent)",
      "timeline-grid": "color-mix(in srgb, var(--line-strong) 34%, transparent)",
      "page-grid": "color-mix(in srgb, var(--ink) 1.8%, transparent)",
      "empty-bg": "color-mix(in srgb, var(--surface) 42%, transparent)",
      "shadow-rgb": "51 43 28",
      "ink-shadow-rgb": "51 43 28",
      "cv-shadow-rgb": "51 43 28",
      "media-shadow-rgb": "51 43 28",
      "lightbox-bg": "rgb(51 43 28 / 96%)",
      "lightbox-backdrop": "rgb(51 43 28 / 82%)",
      "lightbox-ink": "#ffffff",
      "lightbox-ink-soft": "rgb(255 255 255 / 82%)",
      "lightbox-ink-faint": "rgb(255 255 255 / 68%)",
      "lightbox-line": "rgb(255 255 255 / 32%)",
      "lightbox-control-bg": "rgb(255 255 255 / 8%)",
      "lightbox-control-hover": "rgb(255 255 255 / 17%)",
      "media-control-bg": "rgb(51 43 28 / 72%)",
      "media-control-hover": "rgb(51 43 28 / 90%)",
      "video-bg": "#60451c",
      "shadow": "0 18px 50px rgb(var(--shadow-rgb) / 8%)"
    },
    "slate": {
      "surface-soft": "var(--surface-muted)",
      "header-tint": "color-mix(in srgb, var(--bg) 91%, transparent)",
      "cv-header-tint": "color-mix(in srgb, var(--bg) 94%, transparent)",
      "header-line": "color-mix(in srgb, var(--line-strong) 55%, transparent)",
      "section-line": "color-mix(in srgb, var(--line-strong) 42%, transparent)",
      "row-line": "color-mix(in srgb, var(--line) 62%, transparent)",
      "timeline-grid": "color-mix(in srgb, var(--line-strong) 34%, transparent)",
      "page-grid": "color-mix(in srgb, var(--ink) 1.8%, transparent)",
      "empty-bg": "color-mix(in srgb, var(--surface) 42%, transparent)",
      "shadow-rgb": "0 0 0",
      "ink-shadow-rgb": "0 0 0",
      "cv-shadow-rgb": "0 0 0",
      "media-shadow-rgb": "0 0 0",
      "lightbox-bg": "rgb(17 24 39 / 96%)",
      "lightbox-backdrop": "rgb(17 24 39 / 82%)",
      "lightbox-ink": "#ffffff",
      "lightbox-ink-soft": "rgb(255 255 255 / 82%)",
      "lightbox-ink-faint": "rgb(255 255 255 / 68%)",
      "lightbox-line": "rgb(255 255 255 / 32%)",
      "lightbox-control-bg": "rgb(255 255 255 / 8%)",
      "lightbox-control-hover": "rgb(255 255 255 / 17%)",
      "media-control-bg": "rgb(36 49 68 / 72%)",
      "media-control-hover": "rgb(36 49 68 / 90%)",
      "video-bg": "#111827",
      "shadow": "0 18px 50px rgb(var(--shadow-rgb) / 32%)"
    },
    "mint": {
      "surface-soft": "var(--surface-muted)",
      "header-tint": "color-mix(in srgb, var(--bg) 91%, transparent)",
      "cv-header-tint": "color-mix(in srgb, var(--bg) 94%, transparent)",
      "header-line": "color-mix(in srgb, var(--line-strong) 55%, transparent)",
      "section-line": "color-mix(in srgb, var(--line-strong) 42%, transparent)",
      "row-line": "color-mix(in srgb, var(--line) 62%, transparent)",
      "timeline-grid": "color-mix(in srgb, var(--line-strong) 34%, transparent)",
      "page-grid": "color-mix(in srgb, var(--ink) 1.8%, transparent)",
      "empty-bg": "color-mix(in srgb, var(--surface) 42%, transparent)",
      "shadow-rgb": "23 61 42",
      "ink-shadow-rgb": "23 61 42",
      "cv-shadow-rgb": "23 61 42",
      "media-shadow-rgb": "23 61 42",
      "lightbox-bg": "rgb(23 61 42 / 96%)",
      "lightbox-backdrop": "rgb(23 61 42 / 82%)",
      "lightbox-ink": "#ffffff",
      "lightbox-ink-soft": "rgb(255 255 255 / 82%)",
      "lightbox-ink-faint": "rgb(255 255 255 / 68%)",
      "lightbox-line": "rgb(255 255 255 / 32%)",
      "lightbox-control-bg": "rgb(255 255 255 / 8%)",
      "lightbox-control-hover": "rgb(255 255 255 / 17%)",
      "media-control-bg": "rgb(23 61 42 / 72%)",
      "media-control-hover": "rgb(23 61 42 / 90%)",
      "video-bg": "#194e32",
      "shadow": "0 18px 50px rgb(var(--shadow-rgb) / 8%)"
    },
    "teal": {
      "surface-soft": "var(--surface-muted)",
      "header-tint": "color-mix(in srgb, var(--bg) 91%, transparent)",
      "cv-header-tint": "color-mix(in srgb, var(--bg) 94%, transparent)",
      "header-line": "color-mix(in srgb, var(--line-strong) 55%, transparent)",
      "section-line": "color-mix(in srgb, var(--line-strong) 42%, transparent)",
      "row-line": "color-mix(in srgb, var(--line) 62%, transparent)",
      "timeline-grid": "color-mix(in srgb, var(--line-strong) 34%, transparent)",
      "page-grid": "color-mix(in srgb, var(--ink) 1.8%, transparent)",
      "empty-bg": "color-mix(in srgb, var(--surface) 42%, transparent)",
      "shadow-rgb": "23 59 55",
      "ink-shadow-rgb": "23 59 55",
      "cv-shadow-rgb": "23 59 55",
      "media-shadow-rgb": "23 59 55",
      "lightbox-bg": "rgb(23 59 55 / 96%)",
      "lightbox-backdrop": "rgb(23 59 55 / 82%)",
      "lightbox-ink": "#ffffff",
      "lightbox-ink-soft": "rgb(255 255 255 / 82%)",
      "lightbox-ink-faint": "rgb(255 255 255 / 68%)",
      "lightbox-line": "rgb(255 255 255 / 32%)",
      "lightbox-control-bg": "rgb(255 255 255 / 8%)",
      "lightbox-control-hover": "rgb(255 255 255 / 17%)",
      "media-control-bg": "rgb(23 59 55 / 72%)",
      "media-control-hover": "rgb(23 59 55 / 90%)",
      "video-bg": "#0e514a",
      "shadow": "0 18px 50px rgb(var(--shadow-rgb) / 8%)"
    },
    "sky": {
      "surface-soft": "var(--surface-muted)",
      "header-tint": "color-mix(in srgb, var(--bg) 91%, transparent)",
      "cv-header-tint": "color-mix(in srgb, var(--bg) 94%, transparent)",
      "header-line": "color-mix(in srgb, var(--line-strong) 55%, transparent)",
      "section-line": "color-mix(in srgb, var(--line-strong) 42%, transparent)",
      "row-line": "color-mix(in srgb, var(--line) 62%, transparent)",
      "timeline-grid": "color-mix(in srgb, var(--line-strong) 34%, transparent)",
      "page-grid": "color-mix(in srgb, var(--ink) 1.8%, transparent)",
      "empty-bg": "color-mix(in srgb, var(--surface) 42%, transparent)",
      "shadow-rgb": "23 53 75",
      "ink-shadow-rgb": "23 53 75",
      "cv-shadow-rgb": "23 53 75",
      "media-shadow-rgb": "23 53 75",
      "lightbox-bg": "rgb(23 53 75 / 96%)",
      "lightbox-backdrop": "rgb(23 53 75 / 82%)",
      "lightbox-ink": "#ffffff",
      "lightbox-ink-soft": "rgb(255 255 255 / 82%)",
      "lightbox-ink-faint": "rgb(255 255 255 / 68%)",
      "lightbox-line": "rgb(255 255 255 / 32%)",
      "lightbox-control-bg": "rgb(255 255 255 / 8%)",
      "lightbox-control-hover": "rgb(255 255 255 / 17%)",
      "media-control-bg": "rgb(23 53 75 / 72%)",
      "media-control-hover": "rgb(23 53 75 / 90%)",
      "video-bg": "#0d5178",
      "shadow": "0 18px 50px rgb(var(--shadow-rgb) / 8%)"
    },
    "indigo": {
      "surface-soft": "var(--surface-muted)",
      "header-tint": "color-mix(in srgb, var(--bg) 91%, transparent)",
      "cv-header-tint": "color-mix(in srgb, var(--bg) 94%, transparent)",
      "header-line": "color-mix(in srgb, var(--line-strong) 55%, transparent)",
      "section-line": "color-mix(in srgb, var(--line-strong) 42%, transparent)",
      "row-line": "color-mix(in srgb, var(--line) 62%, transparent)",
      "timeline-grid": "color-mix(in srgb, var(--line-strong) 34%, transparent)",
      "page-grid": "color-mix(in srgb, var(--ink) 1.8%, transparent)",
      "empty-bg": "color-mix(in srgb, var(--surface) 42%, transparent)",
      "shadow-rgb": "39 44 82",
      "ink-shadow-rgb": "39 44 82",
      "cv-shadow-rgb": "39 44 82",
      "media-shadow-rgb": "39 44 82",
      "lightbox-bg": "rgb(39 44 82 / 96%)",
      "lightbox-backdrop": "rgb(39 44 82 / 82%)",
      "lightbox-ink": "#ffffff",
      "lightbox-ink-soft": "rgb(255 255 255 / 82%)",
      "lightbox-ink-faint": "rgb(255 255 255 / 68%)",
      "lightbox-line": "rgb(255 255 255 / 32%)",
      "lightbox-control-bg": "rgb(255 255 255 / 8%)",
      "lightbox-control-hover": "rgb(255 255 255 / 17%)",
      "media-control-bg": "rgb(39 44 82 / 72%)",
      "media-control-hover": "rgb(39 44 82 / 90%)",
      "video-bg": "#363e89",
      "shadow": "0 18px 50px rgb(var(--shadow-rgb) / 8%)"
    },
    "amber": {
      "surface-soft": "var(--surface-muted)",
      "header-tint": "color-mix(in srgb, var(--bg) 91%, transparent)",
      "cv-header-tint": "color-mix(in srgb, var(--bg) 94%, transparent)",
      "header-line": "color-mix(in srgb, var(--line-strong) 55%, transparent)",
      "section-line": "color-mix(in srgb, var(--line-strong) 42%, transparent)",
      "row-line": "color-mix(in srgb, var(--line) 62%, transparent)",
      "timeline-grid": "color-mix(in srgb, var(--line-strong) 34%, transparent)",
      "page-grid": "color-mix(in srgb, var(--ink) 1.8%, transparent)",
      "empty-bg": "color-mix(in srgb, var(--surface) 42%, transparent)",
      "shadow-rgb": "73 53 20",
      "ink-shadow-rgb": "73 53 20",
      "cv-shadow-rgb": "73 53 20",
      "media-shadow-rgb": "73 53 20",
      "lightbox-bg": "rgb(73 53 20 / 96%)",
      "lightbox-backdrop": "rgb(73 53 20 / 82%)",
      "lightbox-ink": "#ffffff",
      "lightbox-ink-soft": "rgb(255 255 255 / 82%)",
      "lightbox-ink-faint": "rgb(255 255 255 / 68%)",
      "lightbox-line": "rgb(255 255 255 / 32%)",
      "lightbox-control-bg": "rgb(255 255 255 / 8%)",
      "lightbox-control-hover": "rgb(255 255 255 / 17%)",
      "media-control-bg": "rgb(73 53 20 / 72%)",
      "media-control-hover": "rgb(73 53 20 / 90%)",
      "video-bg": "#714609",
      "shadow": "0 18px 50px rgb(var(--shadow-rgb) / 8%)"
    },
    "graphite": {
      "surface-soft": "var(--surface-muted)",
      "header-tint": "color-mix(in srgb, var(--bg) 91%, transparent)",
      "cv-header-tint": "color-mix(in srgb, var(--bg) 94%, transparent)",
      "header-line": "color-mix(in srgb, var(--line-strong) 55%, transparent)",
      "section-line": "color-mix(in srgb, var(--line-strong) 42%, transparent)",
      "row-line": "color-mix(in srgb, var(--line) 62%, transparent)",
      "timeline-grid": "color-mix(in srgb, var(--line-strong) 34%, transparent)",
      "page-grid": "color-mix(in srgb, var(--ink) 1.8%, transparent)",
      "empty-bg": "color-mix(in srgb, var(--surface) 42%, transparent)",
      "shadow-rgb": "0 0 0",
      "ink-shadow-rgb": "0 0 0",
      "cv-shadow-rgb": "0 0 0",
      "media-shadow-rgb": "0 0 0",
      "lightbox-bg": "rgb(23 23 23 / 96%)",
      "lightbox-backdrop": "rgb(23 23 23 / 82%)",
      "lightbox-ink": "#ffffff",
      "lightbox-ink-soft": "rgb(255 255 255 / 82%)",
      "lightbox-ink-faint": "rgb(255 255 255 / 68%)",
      "lightbox-line": "rgb(255 255 255 / 32%)",
      "lightbox-control-bg": "rgb(255 255 255 / 8%)",
      "lightbox-control-hover": "rgb(255 255 255 / 17%)",
      "media-control-bg": "rgb(45 45 45 / 72%)",
      "media-control-hover": "rgb(45 45 45 / 90%)",
      "video-bg": "#171717",
      "shadow": "0 18px 50px rgb(var(--shadow-rgb) / 32%)"
    }
  };

  const editorPresetDerived = {
    "sage": {
      "tooltip": "#f8fafc",
      "tooltip-bg": "#1e293b",
      "tooltip-line": "#334155",
      "shadow": "#16231d20",
      "tooltip-shadow": "rgb(15 23 42 / 20%)",
      "dialog-shadow": "rgb(0 0 0 / 35%)",
      "backdrop": "rgb(12 20 16 / 72%)"
    },
    "ocean": {
      "tooltip": "#f8fafc",
      "tooltip-bg": "#1e293b",
      "tooltip-line": "#334155",
      "shadow": "#16231d20",
      "tooltip-shadow": "rgb(15 23 42 / 20%)",
      "dialog-shadow": "rgb(0 0 0 / 35%)",
      "backdrop": "rgb(12 20 16 / 72%)"
    },
    "violet": {
      "tooltip": "#f8fafc",
      "tooltip-bg": "#1e293b",
      "tooltip-line": "#334155",
      "shadow": "#16231d20",
      "tooltip-shadow": "rgb(15 23 42 / 20%)",
      "dialog-shadow": "rgb(0 0 0 / 35%)",
      "backdrop": "rgb(12 20 16 / 72%)"
    },
    "rose": {
      "tooltip": "#f8fafc",
      "tooltip-bg": "#1e293b",
      "tooltip-line": "#334155",
      "shadow": "#16231d20",
      "tooltip-shadow": "rgb(15 23 42 / 20%)",
      "dialog-shadow": "rgb(0 0 0 / 35%)",
      "backdrop": "rgb(12 20 16 / 72%)"
    },
    "sand": {
      "tooltip": "#f8fafc",
      "tooltip-bg": "#1e293b",
      "tooltip-line": "#334155",
      "shadow": "#16231d20",
      "tooltip-shadow": "rgb(15 23 42 / 20%)",
      "dialog-shadow": "rgb(0 0 0 / 35%)",
      "backdrop": "rgb(12 20 16 / 72%)"
    },
    "slate": {
      "tooltip": "#edf2fa",
      "tooltip-bg": "#243144",
      "tooltip-line": "#60738e",
      "shadow": "rgb(0 0 0 / 35%)",
      "tooltip-shadow": "rgb(0 0 0 / 35%)",
      "dialog-shadow": "rgb(0 0 0 / 55%)",
      "backdrop": "rgb(3 8 16 / 78%)"
    },
    "mint": {
      "tooltip": "#f8fafc",
      "tooltip-bg": "#1e293b",
      "tooltip-line": "#334155",
      "shadow": "#16231d20",
      "tooltip-shadow": "rgb(15 23 42 / 20%)",
      "dialog-shadow": "rgb(0 0 0 / 35%)",
      "backdrop": "rgb(12 20 16 / 72%)"
    },
    "teal": {
      "tooltip": "#f8fafc",
      "tooltip-bg": "#1e293b",
      "tooltip-line": "#334155",
      "shadow": "#16231d20",
      "tooltip-shadow": "rgb(15 23 42 / 20%)",
      "dialog-shadow": "rgb(0 0 0 / 35%)",
      "backdrop": "rgb(12 20 16 / 72%)"
    },
    "sky": {
      "tooltip": "#f8fafc",
      "tooltip-bg": "#1e293b",
      "tooltip-line": "#334155",
      "shadow": "#16231d20",
      "tooltip-shadow": "rgb(15 23 42 / 20%)",
      "dialog-shadow": "rgb(0 0 0 / 35%)",
      "backdrop": "rgb(12 20 16 / 72%)"
    },
    "indigo": {
      "tooltip": "#f8fafc",
      "tooltip-bg": "#1e293b",
      "tooltip-line": "#334155",
      "shadow": "#16231d20",
      "tooltip-shadow": "rgb(15 23 42 / 20%)",
      "dialog-shadow": "rgb(0 0 0 / 35%)",
      "backdrop": "rgb(12 20 16 / 72%)"
    },
    "amber": {
      "tooltip": "#f8fafc",
      "tooltip-bg": "#1e293b",
      "tooltip-line": "#334155",
      "shadow": "#16231d20",
      "tooltip-shadow": "rgb(15 23 42 / 20%)",
      "dialog-shadow": "rgb(0 0 0 / 35%)",
      "backdrop": "rgb(12 20 16 / 72%)"
    },
    "graphite": {
      "tooltip": "#f3f3f3",
      "tooltip-bg": "#303030",
      "tooltip-line": "#707070",
      "shadow": "rgb(0 0 0 / 35%)",
      "tooltip-shadow": "rgb(0 0 0 / 35%)",
      "dialog-shadow": "rgb(0 0 0 / 55%)",
      "backdrop": "rgb(0 0 0 / 78%)"
    }
  };

  const fields = Object.freeze([
    { key: "bg", label: "페이지 배경", group: "배경과 표면" },
    { key: "surface", label: "카드·입력 배경", group: "배경과 표면" },
    { key: "surface-muted", label: "보조 배경", group: "배경과 표면" },
    { key: "surface-strong", label: "진한 보조 배경", group: "배경과 표면" },
    { key: "ink", label: "본문 글자", group: "텍스트" },
    { key: "ink-soft", label: "보조 글자", group: "텍스트" },
    { key: "ink-faint", label: "옅은 글자", group: "텍스트" },
    { key: "line", label: "기본 테두리", group: "선과 테두리" },
    { key: "line-strong", label: "강조 테두리", group: "선과 테두리" },
    { key: "accent", label: "강조색", group: "강조와 버튼" },
    { key: "accent-strong", label: "진한 강조색", group: "강조와 버튼" },
    { key: "accent-soft", label: "강조 배경", group: "강조와 버튼" },
    { key: "focus", label: "선택·포커스", group: "강조와 버튼" },
    { key: "danger", label: "오류·삭제", group: "상태와 알림" },
    { key: "on-accent", label: "강조색 위 글자", group: "강조와 버튼" },
    { key: "on-danger", label: "오류색 위 글자", group: "상태와 알림" },
    { key: "on-ink", label: "반전 글자", group: "텍스트" },
    { key: "warning", label: "경고 글자", group: "상태와 알림" },
    { key: "warning-soft", label: "경고 배경", group: "상태와 알림" },
    { key: "cv-bg", label: "CV 배경", group: "배경과 표면" },
    { key: "danger-strong", label: "진한 오류색", group: "상태와 알림" },
    { key: "danger-soft", label: "오류 배경", group: "상태와 알림" },
    { key: "danger-line", label: "오류 테두리", group: "상태와 알림" },
    { key: "disabled", label: "비활성 글자", group: "텍스트" },
    { key: "info", label: "안내 글자", group: "상태와 알림" },
    { key: "info-soft", label: "안내 배경", group: "상태와 알림" }
  ].map(Object.freeze));
  const presets = Object.freeze(presetData.map((preset) => Object.freeze({
    ...preset, colors: Object.freeze(preset.colors)
  })));
  const customIDs = Object.freeze(["custom-1", "custom-2", "custom-3"]);
  const presetByID = new Map(presets.map((preset) => [preset.id, preset]));
  const isRecord = (value) => value !== null && typeof value === "object" && !Array.isArray(value);
  const validHex = (value) => typeof value === "string" && /^#[0-9a-f]{6}$/i.test(value);
  const clone = (value) => Array.isArray(value) ? value.map(clone)
    : isRecord(value) ? Object.fromEntries(Object.entries(value).map(([key, item]) => [key, clone(item)])) : value;

  function normaliseThemeID(value) {
    return presetByID.has(value) || customIDs.includes(value) ? value : "sage";
  }

  // Keep incomplete input for editor validation; only resolve() chooses fallbacks.
  function normaliseCustomThemes(raw) {
    const source = isRecord(raw) ? raw : {};
    return Object.fromEntries(customIDs.map((id, index) => {
      const slot = isRecord(source[id]) ? clone(source[id]) : {};
      return [id, {
        ...slot,
        name: Object.hasOwn(slot, "name") ? slot.name : `사용자 지정 ${index + 1}`,
        base_theme: Object.hasOwn(slot, "base_theme") ? slot.base_theme : "sage",
        colors: isRecord(slot.colors) ? slot.colors : {}
      }];
    }));
  }

  function resolve(settings, id = settings?.color_theme) {
    id = normaliseThemeID(id);
    const custom = customIDs.includes(id);
    const slot = custom ? normaliseCustomThemes(settings?.custom_themes)[id] : null;
    const base = presetByID.get(custom ? slot.base_theme : id) || presetByID.get("sage");
    const colors = { ...base.colors };
    if (slot) {
      fields.forEach(({ key }) => {
        if (validHex(slot.colors[key])) colors[key] = slot.colors[key].toLowerCase();
      });
    }
    return {
      id, label: slot && typeof slot.name === "string" && slot.name.trim() ? slot.name.trim()
        : custom ? `사용자 지정 ${customIDs.indexOf(id) + 1}` : base.label,
      description: custom ? `사용자 지정 · ${base.label}` : base.description,
      custom, baseTheme: base.id, colors, scheme: base.dark ? "dark" : "light"
    };
  }

  function rgb(value) {
    return value.slice(1).match(/../g).map((channel) => parseInt(channel, 16)).join(" ");
  }

  function rgba(value, opacity) {
    return `rgb(${rgb(value)} / ${opacity}%)`;
  }

  function websiteDerived(theme) {
    const result = { ...websitePresetDerived[theme.baseTheme] };
    const base = presetByID.get(theme.baseTheme).colors;
    const colors = theme.colors;
    if (theme.scheme === "light" && colors.ink !== base.ink) {
      ["shadow-rgb", "ink-shadow-rgb", "cv-shadow-rgb", "media-shadow-rgb"].forEach((key) => { result[key] = rgb(colors.ink); });
      result["lightbox-bg"] = rgba(colors.ink, 96);
      result["lightbox-backdrop"] = rgba(colors.ink, 82);
      result["media-control-bg"] = rgba(colors.ink, 72);
      result["media-control-hover"] = rgba(colors.ink, 90);
    }
    if (theme.scheme === "dark" && colors.bg !== base.bg) {
      result["lightbox-bg"] = rgba(colors.bg, 96);
      result["lightbox-backdrop"] = rgba(colors.bg, 82);
    }
    if (theme.scheme === "dark" && colors["surface-muted"] !== base["surface-muted"]) {
      result["media-control-bg"] = rgba(colors["surface-muted"], 72);
      result["media-control-hover"] = rgba(colors["surface-muted"], 90);
    }
    const video = theme.scheme === "dark" ? "bg" : "accent-strong";
    if (colors[video] !== base[video]) result["video-bg"] = colors[video];
    const foreground = theme.scheme === "dark" ? "ink" : "on-ink";
    if (colors[foreground] !== base[foreground]) {
      result["lightbox-ink"] = colors[foreground];
      result["lightbox-ink-soft"] = rgba(colors[foreground], 82);
      result["lightbox-ink-faint"] = rgba(colors[foreground], 68);
      result["lightbox-line"] = rgba(colors[foreground], 32);
      result["lightbox-control-bg"] = rgba(colors[foreground], 8);
      result["lightbox-control-hover"] = rgba(colors[foreground], 17);
    }
    return result;
  }

  function editorDerived(theme) {
    const result = { ...editorPresetDerived[theme.baseTheme] };
    const base = presetByID.get(theme.baseTheme).colors;
    const colors = theme.colors;
    if (theme.scheme === "light" && (colors.ink !== base.ink || colors["on-ink"] !== base["on-ink"])) {
      result.tooltip = colors["on-ink"];
      result["tooltip-bg"] = colors.ink;
      result.shadow = rgba(colors.ink, 13);
      result["tooltip-shadow"] = rgba(colors.ink, 20);
      result.backdrop = rgba(colors.ink, 72);
    }
    if (theme.scheme === "dark") {
      if (colors.ink !== base.ink) result.tooltip = colors.ink;
      if (colors["surface-strong"] !== base["surface-strong"]) result["tooltip-bg"] = colors["surface-strong"];
      if (colors.bg !== base.bg) result.backdrop = rgba(colors.bg, 78);
    }
    if (colors["line-strong"] !== base["line-strong"]) result["tooltip-line"] = colors["line-strong"];
    return result;
  }

  function applyVariables(theme, element, values, prefix) {
    if (!element?.style?.setProperty) return theme;
    if (element.dataset) {
      element.dataset.colorTheme = theme.id;
      element.dataset.colorScheme = theme.scheme;
    } else {
      element.setAttribute?.("data-color-theme", theme.id);
      element.setAttribute?.("data-color-scheme", theme.scheme);
    }
    Object.entries(values).forEach(([key, value]) => element.style.setProperty(`--${prefix}${key}`, value));
    element.style.setProperty("color-scheme", theme.scheme);
    return theme;
  }

  function defaultElement() {
    return typeof document === "undefined" ? root.document?.documentElement : document.documentElement;
  }

  function applyWebsite(settings, element = defaultElement()) {
    const theme = resolve(settings);
    return applyVariables(theme, element, { ...theme.colors, ...websiteDerived(theme) }, "");
  }

  function applyEditor(settings, element = defaultElement()) {
    const theme = resolve(settings);
    const names = { ink: "text", "ink-soft": "text-muted", "ink-faint": "text-faint" };
    const values = Object.fromEntries(Object.entries(theme.colors).map(([key, value]) => [names[key] || key, value]));
    return applyVariables(theme, element, { ...values, ...editorDerived(theme) }, "color-");
  }

  const api = Object.freeze({ presets, fields, customIDs, normaliseThemeID, normaliseCustomThemes, resolve, applyWebsite, applyEditor });
  if (typeof module === "object" && module.exports) module.exports = api;
  else root.SiteThemes = api;
})(typeof window === "undefined" ? globalThis : window);
