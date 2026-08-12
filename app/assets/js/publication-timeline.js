(function () {
  "use strict";

  const TYPES = ["international-journal", "domestic-journal", "international-conference", "domestic-conference"];
  const MARK_RADIUS = 6;
  const SPACING = 16; // 12px circles with a 4px gap, horizontally and vertically.
  const PADDING = 14;
  const MIN_HEIGHT = 44;
  const UNDATED_WIDTH = 80;

  function calendarDate(year, month = 1, day = 1) {
    const value = new Date(0);
    value.setUTCFullYear(year, month - 1, day);
    value.setUTCHours(0, 0, 0, 0);
    return value;
  }

  // Partial dates retain their original precision and start at that calendar unit.
  // Invalid and missing dates are never substituted with an estimated date.
  function parseDate(value) {
    const text = typeof value === "string" ? value.trim() : "";
    const match = /^(\d{4})(?:-(\d{2})(?:-(\d{2}))?)?$/.exec(text);
    if (!match) return null;
    const year = Number(match[1]);
    const month = Number(match[2] || 1);
    const day = Number(match[3] || 1);
    if (year < 1 || month < 1 || month > 12 || day < 1 || day > 31) return null;
    const date = calendarDate(year, month, day);
    if (date.getUTCFullYear() !== year || date.getUTCMonth() + 1 !== month || date.getUTCDate() !== day) return null;
    return { text, year, time: date.getTime() };
  }

  function packedRows(records, center) {
    const rows = [];
    const columnCount = Math.min(3, records.length);
    for (let offset = 0; offset < records.length; offset += 3) {
      const items = records.slice(offset, offset + 3);
      rows.push(items.map((record, index) => ({
        record,
        x: center + (index - (columnCount - 1) / 2) * SPACING
      })));
    }
    return rows;
  }

  function firstAuthorsFirst(records) {
    // Stable partition preserves date/ID order within each author role.
    return records.filter((record) => record.firstAuthor)
      .concat(records.filter((record) => !record.firstAuthor));
  }

  function intervalsOverlap(first, second) {
    // Touching intervals have exactly the required 16px center separation.
    return first.left < second.right - 0.000001 && second.left < first.right - 0.000001;
  }

  function rowInterval(marks) {
    return { left: marks[0].x - SPACING / 2, right: marks[marks.length - 1].x + SPACING / 2 };
  }

  function placeCluster(clusterRows, occupied) {
    let base = 0;
    const intervals = clusterRows.map(rowInterval);
    while (intervals.some((interval, index) => (occupied[base + index] || [])
      .some((previous) => intervalsOverlap(interval, previous)))) {
      base += 1;
    }
    intervals.forEach((interval, index) => {
      if (!occupied[base + index]) occupied[base + index] = [];
      occupied[base + index].push(interval);
    });
    return base;
  }

  function mark(record, x, lane, dated) {
    return {
      id: record.id,
      x,
      y: PADDING + MARK_RADIUS + lane * SPACING,
      date: typeof record.date === "string" ? record.date.trim() : "",
      dated
    };
  }

  function layout(records, options = {}) {
    const source = Array.isArray(records) ? records : [];
    const types = Array.isArray(options.types) ? [...options.types] : [...TYPES];
    const proposedYearWidth = Number(options.yearWidth);
    const yearWidth = Number.isFinite(proposedYearWidth) && proposedYearWidth > 0
      ? Math.max(SPACING * 2, proposedYearWidth) : 140;
    const axisDates = Array.isArray(options.axisDates) ? options.axisDates : source.map((record) => record.date);
    const axis = axisDates.map(parseDate);
    const prepared = source.filter((record) => record && types.includes(record.type))
      .map((record) => ({ record, parsed: parseDate(record.date) }));
    const dated = [...axis.filter(Boolean), ...prepared.map((entry) => entry.parsed).filter(Boolean)];
    const firstYear = dated.length ? Math.min(...dated.map((date) => date.year)) : null;
    const lastYear = dated.length ? Math.max(...dated.map((date) => date.year)) : null;
    const width = dated.length ? (lastYear - firstYear + 1) * yearWidth + PADDING * 2 : 0;
    const undatedWidth = axis.some((date) => !date) || prepared.some((entry) => !entry.parsed) ? UNDATED_WIDTH : 0;
    const yearTicks = dated.length ? Array.from({ length: lastYear - firstYear + 2 }, (_, index) => ({
      year: firstYear + index,
      x: PADDING + index * yearWidth
    })) : [];

    function dateX(date) {
      const start = calendarDate(date.year).getTime();
      const end = calendarDate(date.year + 1).getTime();
      return PADDING + (date.year - firstYear + (date.time - start) / (end - start)) * yearWidth;
    }

    const rows = types.map((type) => {
      const entries = prepared.filter((entry) => entry.record.type === type);
      const clusters = new Map();
      const undated = [];
      entries.forEach((entry) => {
        if (!entry.parsed) {
          undated.push(entry.record);
          return;
        }
        const key = entry.parsed.text;
        if (!clusters.has(key)) clusters.set(key, { date: entry.parsed, records: [] });
        clusters.get(key).records.push(entry.record);
      });
      const marks = [];
      const occupied = [];
      const groups = [];
      [...clusters.values()].sort((first, second) => first.date.time - second.date.time
        || first.date.text.localeCompare(second.date.text)).forEach((cluster) => {
        cluster.records.sort((first, second) => first.id - second.id);
        const x = dateX(cluster.date);
        const previous = groups[groups.length - 1];
        // Bound each shared grid by its earliest date, avoiding a transitive chain.
        if (previous && x - previous.firstX <= SPACING * 2 + 0.000001) {
          previous.lastX = x;
          previous.records.push(...cluster.records);
        } else {
          groups.push({ firstX: x, lastX: x, records: [...cluster.records] });
        }
      });
      groups.forEach((group) => {
        const halfSpan = (Math.min(3, group.records.length) - 1) * SPACING / 2;
        const anchor = (group.firstX + group.lastX) / 2;
        const center = Math.max(PADDING + halfSpan, Math.min(width - PADDING - halfSpan, anchor));
        const clusterRows = packedRows(firstAuthorsFirst(group.records), center);
        const base = placeCluster(clusterRows, occupied);
        clusterRows.forEach((clusterRow, index) => clusterRow.forEach((entry) => {
          marks.push(mark(entry.record, entry.x, base + index, true));
        }));
      });
      undated.sort((first, second) => first.id - second.id);
      const undatedRows = packedRows(firstAuthorsFirst(undated), width + undatedWidth / 2);
      undatedRows.forEach((clusterRow, index) => clusterRow.forEach((entry) => {
        marks.push(mark(entry.record, entry.x, index, false));
      }));
      const lanes = Math.max(occupied.length, undatedRows.length);
      return {
        type,
        height: Math.max(MIN_HEIGHT, PADDING * 2 + MARK_RADIUS * 2 + Math.max(0, lanes - 1) * SPACING),
        marks
      };
    });

    return { firstYear, lastYear, yearWidth, width, undatedWidth, yearTicks, rows };
  }

  window.PublicationTimeline = Object.freeze({ layout });
})();
