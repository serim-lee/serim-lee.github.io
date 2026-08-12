(function (root) {
  "use strict";

  const text = (value) => String(value || "").trim();
  const localized = (record, key, language) => text(record?.[`${key}_${language}`])
    || text(record?.[`${key}_${language === "ko" ? "en" : "ko"}`]);
  const months = ["January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"];

  function catalogFor(item, settings) {
    const conference = text(item.publication_type).endsWith("-conference");
    return (settings?.[conference ? "conferences" : "journals"] || [])
      .find((record) => record.id === item[conference ? "conference_id" : "journal_id"]);
  }

  function publicationDate(item, settings) {
    if (text(item.publication_type).endsWith("-conference")) {
      const conference = catalogFor(item, settings);
      return text(conference?.proceedings_date) || text(conference?.end_date);
    }
    return text(item.issue_date) || text(item.online_date);
  }

  function conferencePeriod(record, language) {
    const start = text(record.start_date), end = text(record.end_date);
    const format = (value) => {
      const parts = value.split("-").map(Number);
      if (parts.length !== 3) return value;
      return language === "ko" ? `${parts[0]}년 ${parts[1]}월 ${parts[2]}일`
        : `${months[parts[1] - 1]} ${parts[2]}, ${parts[0]}`;
    };
    if (!start || !end || start === end) return format(start || end);
    const [sy, sm, sd] = start.split("-").map(Number);
    const [ey, em, ed] = end.split("-").map(Number);
    if (sy !== ey) return `${format(start)} – ${format(end)}`;
    if (language === "ko") return sm === em
      ? `${sy}년 ${sm}월 ${sd}-${ed}일` : `${sy}년 ${sm}월 ${sd}일-${em}월 ${ed}일`;
    return sm === em ? `${months[sm - 1]} ${sd}-${ed}, ${sy}`
      : `${months[sm - 1]} ${sd}-${months[em - 1]} ${ed}, ${sy}`;
  }

  function numberLabel(value, kind, language) {
    return language === "ko" ? `제${value}${kind === "volume" ? "권" : "호"}`
      : `${kind === "volume" ? "Vol." : "No."} ${value}`;
  }

  function containsNumber(value, number, kind) {
    if (!number) return false;
    value = value.replace(/[\p{White_Space}\uFEFF]/gu, " ");
    number = number.replace(/[\p{White_Space}\uFEFF]/gu, " ");
    const escaped = number.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
    const english = kind === "volume" ? "vol(?:ume)?\\.?" : "(?:no\\.?|issue|part)";
    return new RegExp(`(?:^|[^a-z0-9])${english}\\s*${escaped}(?=$|[\\s,;:.()\\[\\]])`, "i").test(value)
      || new RegExp(`(?:^|[^a-z0-9])(?:제\\s*)?${escaped}\\s*${kind === "volume" ? "권" : "호"}(?=$|[\\s,;:.()\\[\\]제])`, "i").test(value);
  }

  function seriesLabel(series, volume, language) {
    if (!volume || containsNumber(series, volume, "volume")) return series;
    const label = numberLabel(volume, "volume", language);
    return series ? `${series}, ${label}` : language === "ko" ? `시리즈 ${label}` : `Series ${label}`;
  }

  function publicationReferenceParts(item, settings, language = "en") {
    const record = catalogFor(item, settings);
    if (!record) return { citation: "", conference: "" };
    // International publications keep their English reference in either UI language.
    const lang = language === "ko" && text(item.publication_type).startsWith("domestic-") ? "ko" : "en";
    const name = localized(record, "name", lang);
    const pages = text(item.pages);
    if (text(item.publication_type).endsWith("-journal")) {
      const parts = [name];
      if (lang === "ko") {
        const numbering = [text(item.volume) ? `제${text(item.volume)}권` : "", text(item.issue) ? `제${text(item.issue)}호` : ""].filter(Boolean).join(" ");
        if (numbering) parts.push(numbering);
      } else {
        if (text(item.volume)) parts.push(`Vol. ${text(item.volume)}`);
        if (text(item.issue)) parts.push(`No. ${text(item.issue)}`);
      }
      if (pages) parts.push(`pp. ${pages}`);
      else if (text(item.elocator)) parts.push(text(item.elocator));
      return { citation: parts.filter(Boolean).join(", ") + (parts.length > 1 ? "." : ""), conference: "" };
    }
    const title = localized(record, "proceedings_title", lang);
    const series = localized(record, "proceedings_series", lang);
    const volume = text(record.proceedings_volume), issue = text(record.proceedings_issue);
    const seriesVolume = text(record.proceedings_series_volume);
    const parts = [title];
    const numbering = [volume && !containsNumber(title, volume, "volume") ? numberLabel(volume, "volume", lang) : "",
      issue && !containsNumber(title, issue, "issue") ? numberLabel(issue, "issue", lang) : ""].filter(Boolean);
    if (numbering.length) parts.push(numbering.join(lang === "ko" ? " " : ", "));
    const seriesIndex = series ? title.toLocaleLowerCase().indexOf(series.toLocaleLowerCase()) : -1;
    if (seriesIndex < 0) {
      const label = seriesLabel(series, seriesVolume, lang);
      if (label) parts.push(label);
    } else if (seriesVolume && !containsNumber(title.slice(seriesIndex), seriesVolume, "volume")) {
      parts.push(seriesLabel("", seriesVolume, lang));
    }
    if (pages) parts.push(`pp. ${pages}`);
    const details = [name, conferencePeriod(record, lang), localized(record, "location", lang)].filter(Boolean).join(", ");
    const citation = parts.filter(Boolean).join(", ").replace(/\.$/, "");
    return {
      citation: citation ? citation + "." : "",
      conference: details ? details.replace(/\.$/, "") + "." : ""
    };
  }

  function publicationVenue(item, settings, language = "en") {
    const parts = publicationReferenceParts(item, settings, language);
    return [parts.citation, parts.conference].filter(Boolean).join(" ");
  }

  function publicationIndexing(item, settings) {
    const values = catalogFor(item, settings)?.indexing;
    return Array.isArray(values) ? values.find((value) => ["SCIE", "SCOPUS", "KCI"].includes(value)) || "" : "";
  }

  function resolvePublication(item, settings) {
    const english = publicationReferenceParts(item, settings, "en");
    const korean = publicationReferenceParts(item, settings, "ko");
    const resolved = {
      ...item,
      date: publicationDate(item, settings),
      venue_en: [english.citation, english.conference].filter(Boolean).join(" "),
      venue_kr: [korean.citation, korean.conference].filter(Boolean).join(" "),
      reference_en: english.citation,
      reference_ko: korean.citation,
      conference_info_en: english.conference,
      conference_info_ko: korean.conference,
      indexing: publicationIndexing(item, settings)
    };
    const conference = text(item.publication_type).endsWith("-conference") ? catalogFor(item, settings) : null;
    if (conference && ["proceedings_volume", "proceedings_issue", "proceedings_series_volume"].some((key) => text(conference[key]))) {
      resolved.volume = text(conference.proceedings_volume);
      resolved.issue = text(conference.proceedings_issue);
      resolved.proceedings_series_volume = text(conference.proceedings_series_volume);
    }
    return resolved;
  }

  const api = { catalogFor, publicationDate, publicationVenue, publicationReferenceParts, publicationIndexing, conferencePeriod, resolvePublication };
  if (typeof module === "object" && module.exports) module.exports = api;
  else root.PublicationBibliography = api;
})(typeof window === "undefined" ? globalThis : window);
