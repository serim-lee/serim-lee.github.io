"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const vm = require("node:vm");
const versions = require("../assets/js/data-version.js");

const root = path.resolve(__dirname, "../..");
const context = { window: {} };
vm.runInNewContext(fs.readFileSync(path.join(root, "app/assets/js/publication-timeline.js"), "utf8"), context);
const layout = (records, options) => JSON.parse(JSON.stringify(context.window.PublicationTimeline.layout(records, options)));
const types = ["international-journal", "domestic-journal", "international-conference", "domestic-conference"];
const record = (id, date, type = "domestic-conference") => ({ id, date, type });
const axisDates = ["2024-01-01", "2024-12-31"];
const marksOf = (chart, type = "domestic-conference") => chart.rows.find((row) => row.type === type).marks;
let assertions = 0;

function check(condition, message) {
  assert.ok(condition, message);
  assertions += 1;
}

function checkSpacing(chart) {
  for (const row of chart.rows) {
    for (let i = 0; i < row.marks.length; i += 1) {
      const mark = row.marks[i];
      check(mark.y - 6 >= 14 && mark.y + 6 <= row.height - 14, "Marks retain vertical padding.");
      if (mark.dated) {
        check(mark.x - 6 >= 0 && mark.x + 6 <= chart.width, "Dated marks stay inside the time chart.");
      } else {
        check(mark.x - 6 >= chart.width && mark.x + 6 <= chart.width + chart.undatedWidth, "Undated marks stay inside their column.");
      }
      for (let j = i + 1; j < row.marks.length; j += 1) {
        const other = row.marks[j];
        check(Math.hypot(mark.x - other.x, mark.y - other.y) >= 16 - 0.000001, `Marks ${mark.id} and ${other.id} are at least 16px apart.`);
      }
    }
  }
}

for (let count = 1; count <= 10; count += 1) {
  const records = Array.from({ length: count }, (_, index) => record(index, "2024-06-15"));
  const chart = layout(records, { types, yearWidth: 140, axisDates });
  const marks = marksOf(chart);
  assert.equal(marks.length, count);
  const occupiedRows = [...new Set(marks.map((mark) => mark.y))];
  assert.equal(occupiedRows.length, Math.ceil(count / 3));
  const firstRow = marks.filter((mark) => mark.y === occupiedRows[0]);
  const firstColumnX = firstRow[0].x;
  occupiedRows.forEach((y) => {
    const rowMarks = marks.filter((mark) => mark.y === y);
    check(rowMarks.length <= 3, "Same-date marks have at most three columns.");
    check(Math.abs(rowMarks[0].x - firstColumnX) < 0.000001,
      "Overflow rows, including a short final row, align with the first column.");
    rowMarks.forEach((mark, index) => {
      check(Math.abs(mark.x - (firstColumnX + index * 16)) < 0.000001,
        "Same-date columns retain 16px center spacing.");
    });
  });
  checkSpacing(chart);
}

const crowded = [];
let id = 0;
for (let day = 1; day <= 28; day += 1) {
  for (let count = 0; count < (day % 7) + 1; count += 1) {
    crowded.push(record(id++, `2024-06-${String(day).padStart(2, "0")}`));
  }
}
const crowdedChart = layout(crowded, { types, yearWidth: 140, axisDates });
checkSpacing(crowdedChart);
assert.deepEqual(crowdedChart, layout([...crowded].reverse(), { types, yearWidth: 140, axisDates }), "Input sorting cannot alter marks or their positions.");
const crowdedMarks = marksOf(crowdedChart);
const sharedAnchor = 14 + ((152 + 152 + 27) / 2 / 366) * 140;
crowdedMarks.forEach((mark, index) => {
  assert.equal(mark.id, index, "Shared-grid cells follow dates chronologically, then publication IDs.");
  assert.equal(mark.date, crowded[index].date, "Shared-grid marks retain their exact original date.");
  check(Math.abs(mark.x - (sharedAnchor + (index % 3 - 1) * 16)) < 0.000001,
    "Nearby dates share the same three columns and fill cells from left to right.");
  assert.equal(mark.y, 20 + Math.floor(index / 3) * 16, "Shared-grid cells leave no gaps in earlier rows.");
});

const separateDates = layout([record(0, "2024-06-01"), record(1, "2024-09-15")], { types, yearWidth: 140, axisDates });
const separateMarks = marksOf(separateDates);
check(Math.abs(separateMarks[0].x - (14 + 152 / 366 * 140)) < 0.000001,
  "Widely separated dates retain their independent calendar positions.");
check(Math.abs(separateMarks[1].x - (14 + 258 / 366 * 140)) < 0.000001,
  "Dates beyond the 32px grid footprint are not merged.");
checkSpacing(separateDates);

const boundedDates = layout([record(20, "2024-06-01"), record(10, "2024-08-01"), record(1, "2024-10-01")],
  { types, yearWidth: 140, axisDates });
const boundedMarks = marksOf(boundedDates);
const boundedAnchor = 14 + (152 + 213) / 2 / 366 * 140;
assert.deepEqual(boundedMarks.map((mark) => mark.id), [20, 10, 1]);
check(Math.abs(boundedMarks[0].x - (boundedAnchor - 8)) < 0.000001,
  "The first two nearby dates share one centered grid.");
check(Math.abs(boundedMarks[1].x - (boundedAnchor + 8)) < 0.000001,
  "Two-publication grids use two centered columns.");
check(Math.abs(boundedMarks[2].x - (14 + 274 / 366 * 140)) < 0.000001,
  "A chain of nearby dates cannot expand a shared grid beyond its earliest-date bound.");
checkSpacing(boundedDates);

const sameDateRoles = [
  { ...record(4, "2024-06-01"), firstAuthor: false },
  { ...record(3, "2024-06-01"), firstAuthor: true },
  { ...record(0, "2024-06-01"), firstAuthor: false },
  { ...record(1, "2024-06-01"), firstAuthor: true },
  { ...record(2, "2024-06-01"), firstAuthor: false },
  { ...record(5, "2024-06-01"), firstAuthor: true }
];
const sameDateRoleChart = layout(sameDateRoles, { types, yearWidth: 140, axisDates });
assert.deepEqual(marksOf(sameDateRoleChart).map((mark) => mark.id), [1, 3, 5, 0, 2, 4],
  "Same-date first-author marks fill the grid before coauthor marks, with stable ID ties.");
assert.deepEqual(sameDateRoleChart, layout([...sameDateRoles].reverse(), { types, yearWidth: 140, axisDates }));
checkSpacing(sameDateRoleChart);

const nearbyRoles = [
  { ...record(9, "2024-06-03"), firstAuthor: true },
  { ...record(3, "2024-06-02"), firstAuthor: true },
  { ...record(4, "2024-06-01"), firstAuthor: false },
  { ...record(2, "2024-06-02"), firstAuthor: true },
  { ...record(1, "2024-06-01"), firstAuthor: false },
  { ...record(8, "2024-06-03"), firstAuthor: false },
  { ...record(5, "2024-06-02"), firstAuthor: true }
];
const nearbyRoleChart = layout(nearbyRoles, { types, yearWidth: 140, axisDates });
const nearbyRoleMarks = marksOf(nearbyRoleChart);
assert.deepEqual(nearbyRoleMarks.map((mark) => mark.id), [2, 3, 5, 9, 1, 4, 8],
  "Newer first-author dates precede older coauthor dates inside a shared grid, including across row wraps.");
assert.equal(nearbyRoleMarks[3].y, nearbyRoleMarks[4].y,
  "The last first-author mark and first coauthor mark share the next row in priority order.");
nearbyRoleMarks.forEach((mark) => assert.equal(mark.date, nearbyRoles.find((item) => item.id === mark.id).date));
assert.deepEqual(nearbyRoleChart, layout([...nearbyRoles].reverse(), { types, yearWidth: 140, axisDates }));
checkSpacing(nearbyRoleChart);

const separateRoleChart = layout([
  { ...record(0, "2024-06-01"), firstAuthor: false },
  { ...record(1, "2024-09-15"), firstAuthor: true }
], { types, yearWidth: 140, axisDates });
assert.deepEqual(separateRoleChart, separateDates,
  "First-author priority cannot move publications between separate date grids.");

const undatedRoleChart = layout([
  { ...record(9, ""), firstAuthor: true },
  { ...record(3, ""), firstAuthor: true },
  { ...record(0, ""), firstAuthor: false },
  { ...record(2, ""), firstAuthor: false }
], { types });
assert.deepEqual(marksOf(undatedRoleChart).map((mark) => mark.id), [3, 9, 0, 2],
  "The undated grid also places first authors first, with stable ID order within each role.");
checkSpacing(undatedRoleChart);

const edges = layout([
  ...Array.from({ length: 10 }, (_, index) => record(index, "2023-01-01")),
  ...Array.from({ length: 10 }, (_, index) => record(index + 10, "2024-12-31"))
], { types, yearWidth: 140 });
assert.equal(edges.firstYear, 2023);
assert.equal(edges.lastYear, 2024);
assert.equal(edges.width, 308);
assert.deepEqual(edges.yearTicks, [{ year: 2023, x: 14 }, { year: 2024, x: 154 }, { year: 2025, x: 294 }]);
checkSpacing(edges);

const dateCases = [
  "2024", "2024-02", "2024-02-29", "2000-02-29", "1900-02-28", "0001-01-01",
  "2023-02-29", "1900-02-29", "2024-04-31", "2024-00", "2024-13", "2024-02-00", "0000", "2024-2", "", "not a date"
];
const dateChart = layout(dateCases.map((date, index) => record(index, date)), { types });
const dateMarks = marksOf(dateChart);
dateCases.forEach((date, index) => {
  const mark = dateMarks.find((mark) => mark.id === index);
  assert.equal(mark.date, date);
  assert.equal(mark.dated, index < 6, `Strict validation for ${date}.`);
});
checkSpacing(dateChart);
const leapAxis = layout([record(0, "2024-07-01"), record(1, "2023-07-01")], { types, yearWidth: 140 });
check(Math.abs(marksOf(leapAxis).find((mark) => mark.id === 0).x - (154 + 182 / 366 * 140)) < 0.000001,
  "Leap-year days interpolate between their UTC year boundaries.");
check(Math.abs(marksOf(leapAxis).find((mark) => mark.id === 1).x - (14 + 181 / 365 * 140)) < 0.000001,
  "Common-year days interpolate between their UTC year boundaries.");

const missing = layout(Array.from({ length: 10 }, (_, index) => record(index, "")), { types });
assert.equal(missing.width, 0);
assert.equal(missing.firstYear, null);
assert.equal(missing.lastYear, null);
assert.equal(missing.undatedWidth, 80);
assert.deepEqual(missing.yearTicks, []);
checkSpacing(missing);

const publications = versions.decode(JSON.parse(fs.readFileSync(path.join(root, "data/publications.json"), "utf8")), "publications.json", { collection: true }).data;
const people = versions.decode(JSON.parse(fs.readFileSync(path.join(root, "data/people.json"), "utf8")), "people.json", { collection: true }).data;
const selfIds = new Set(people.filter((person) => person.is_self).map((person) => person.id));
const actual = publications.map((publication, index) => ({
  ...record(index, publication.date, publication.publication_type),
  firstAuthor: selfIds.has(publication.author_ids[0])
}));
const options = { types, yearWidth: 140, axisDates: publications.map((publication) => publication.date) };
const actualChart = layout(actual, options);
assert.deepEqual(actualChart.rows.map((row) => row.type), types);
assert.equal(actualChart.rows.reduce((total, row) => total + row.marks.length, 0), publications.length);
actualChart.rows.forEach((row) => assert.equal(row.marks.length, publications.filter((publication) => publication.publication_type === row.type).length));
checkSpacing(actualChart);
assert.deepEqual(actualChart, layout([...actual].reverse(), options));
const filtered = layout(actual.filter((item) => item.type === "domestic-journal"), options);
assert.equal(filtered.firstYear, actualChart.firstYear);
assert.equal(filtered.lastYear, actualChart.lastYear);
assert.equal(filtered.width, actualChart.width);
assert.equal(filtered.undatedWidth, actualChart.undatedWidth);
assert.deepEqual(filtered.rows.map((row) => row.type), types);
checkSpacing(filtered);
const empty = layout([], options);
assert.equal(empty.width, actualChart.width);
assert.equal(empty.undatedWidth, actualChart.undatedWidth);
assert.deepEqual(empty.rows.map((row) => row.type), types);
empty.rows.forEach((row) => {
  assert.equal(row.height, 44);
  assert.deepEqual(row.marks, []);
});
const noData = layout([], { types });
assert.equal(noData.width, 0);
assert.equal(noData.undatedWidth, 0);
assert.deepEqual(noData.rows.map((row) => row.type), types);
const everyType = layout(types.map((type, index) => record(index, "2025-01-01", type)), { types });
assert.deepEqual(everyType.rows.map((row) => [row.type, row.marks.length]), types.map((type) => [type, 1]));
checkSpacing(everyType);

console.log(`Publication timeline layout passed (${assertions} geometric checks; ${publications.length} real publications).`);
