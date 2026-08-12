(function () {
  "use strict";

  const CACHE_KEY = "portfolio-publication-citations-v1";
  const DAY = 24 * 60 * 60 * 1000;
  const FAILURE_COOLDOWN = 15 * 60 * 1000;

  function normalizeDOI(value) {
    const doi = String(value || "").trim()
      .replace(/^doi:\s*/i, "")
      .replace(/^(?:https?:\/\/)?(?:dx\.)?doi\.org\//i, "")
      .toLowerCase();
    return /^10\.\d{4,9}\/\S+$/.test(doi) ? doi : "";
  }

  function validRecord(record) {
    if (!record || !Number.isSafeInteger(record.count) || record.count < 0
      || !Number.isFinite(Date.parse(record.checked_at))) return null;
    if (record.provider === "openalex" && /^W\d+$/.test(record.work_id)) {
      return { provider: "openalex", work_id: record.work_id, count: record.count, checked_at: record.checked_at };
    }
    if (record.provider === "kci" && /^ART\d+$/.test(record.article_id)) {
      return { provider: "kci", article_id: record.article_id, count: record.count, checked_at: record.checked_at,
        list_pending: record.list_pending === true };
    }
    return null;
  }

  function citingURL(record) {
    if (record.provider === "kci") {
      return `https://www.kci.go.kr/kciportal/ci/sereArticleSearch/ciSereArtiView.kci?sereArticleSearchBean.artiId=${record.article_id}#:~:text=${encodeURIComponent("인용현황")}`;
    }
    return `https://openalex.org/works?filter=cites:${record.work_id}`;
  }

  async function fetchJSON(url, timeout) {
    const controller = new AbortController();
    const timer = window.setTimeout(() => controller.abort(), timeout);
    try {
      const response = await fetch(url, { signal: controller.signal, credentials: "omit" });
      if (!response.ok) {
        const error = new Error(`Citation lookup: HTTP ${response.status}`);
        error.status = response.status;
        throw error;
      }
      return await response.json();
    } finally {
      window.clearTimeout(timer);
    }
  }

  async function fetchSnapshot() {
    try {
      return await fetchJSON("exports/publication-citations.json", 4000);
    } catch (error) {
      if (error.status !== 404) throw error;
      return fetchJSON("data/publication-citations.json", 4000);
    }
  }

  function createClient() {
    const records = new Map();
    const attempts = new Map();
    let started = false;

    function merge(doi, value) {
      const key = normalizeDOI(doi);
      const record = validRecord(value);
      if (!key || !record) return false;
      const current = records.get(key);
      // KCI counts describe the domestic index; do not silently replace their source.
      if (current?.provider === "kci" && record.provider !== "kci") return false;
      if (current?.provider === record.provider && Date.parse(current.checked_at) > Date.parse(record.checked_at)) return false;
      if (JSON.stringify(current) === JSON.stringify(record)) return false;
      records.set(key, record);
      return true;
    }

    function save() {
      try {
        window.localStorage.setItem(CACHE_KEY, JSON.stringify({
          version: 1, records: Object.fromEntries(records), attempts: Object.fromEntries(attempts)
        }));
      } catch (_) { /* Storage can be disabled or full; live counts still work. */ }
    }

    async function start(publications, onChange) {
      if (started) return;
      started = true;
      const eligible = new Set(publications
        .filter((item) => !window.SiteData.truthy(item.under_review) && !window.SiteData.truthy(item.in_press))
        .map((item) => normalizeDOI(item.doi)).filter(Boolean));
      try {
        const cache = JSON.parse(window.localStorage.getItem(CACHE_KEY));
        if (cache?.version === 1) {
          for (const [doi, record] of Object.entries(cache.records || {})) {
            // KCI's checked-in snapshot remains authoritative, including later corrections.
            if (eligible.has(normalizeDOI(doi)) && record?.provider !== "kci") merge(doi, record);
          }
          for (const [doi, retryAt] of Object.entries(cache.attempts || {})) {
            if (eligible.has(normalizeDOI(doi)) && Number.isFinite(retryAt)
              && retryAt > Date.now() && retryAt <= Date.now() + DAY) attempts.set(normalizeDOI(doi), retryAt);
          }
        }
      } catch (_) { /* A corrupt cache must not prevent rendering or querying. */ }
      if (records.size) onChange();

      try {
        const snapshot = window.SiteDataVersion.decode(await fetchSnapshot(), "publication-citations.json").data;
        let changed = false;
        for (const [doi, record] of Object.entries(snapshot.records || {})) {
          if (eligible.has(normalizeDOI(doi))) changed = merge(doi, record) || changed;
        }
        if (changed) onChange();
      } catch (_) { /* The snapshot is optional; querying can proceed without it. */ }

      const pending = [...eligible].filter((doi) => records.get(doi)?.provider !== "kci"
        && !(attempts.get(doi) > Date.now()));
      for (let offset = 0; offset < pending.length; offset += 100) {
        const batch = pending.slice(offset, offset + 100);
        const params = new URLSearchParams({
          filter: `doi:${batch.join("|")}`, select: "id,doi,cited_by_count", per_page: "100"
        });
        try {
          const response = await fetchJSON(`https://api.openalex.org/works?${params}`, 8000);
          if (!Array.isArray(response.results)) throw new Error("Invalid citation response");
          const checkedAt = new Date().toISOString();
          let changed = false;
          for (const work of response.results) {
            const doi = normalizeDOI(work.doi);
            if (!batch.includes(doi)) continue;
            changed = merge(doi, {
              provider: "openalex", work_id: String(work.id || "").replace(/^https:\/\/openalex\.org\//, ""),
              count: work.cited_by_count, checked_at: checkedAt
            }) || changed;
          }
          // A missing result means unknown, never a fabricated zero citation count.
          batch.forEach((doi) => attempts.set(doi, Date.now() + DAY));
          if (changed) onChange();
        } catch (_) {
          // Keep genuine older counts and their original retrieval dates after outages/429s.
          batch.forEach((doi) => attempts.set(doi, Date.now() + FAILURE_COOLDOWN));
        }
        save();
      }
    }

    function get(item) {
      if (window.SiteData.truthy(item.under_review) || window.SiteData.truthy(item.in_press)) return null;
      return records.get(normalizeDOI(item.doi)) || null;
    }

    function compare(a, b, recency) {
      const first = get(a);
      const second = get(b);
      return (second?.count ?? -1) - (first?.count ?? -1) || recency(a, b);
    }

    return { get, compare, start };
  }

  window.PublicationCitations = Object.freeze({ normalizeDOI, citingURL, createClient });
}());
