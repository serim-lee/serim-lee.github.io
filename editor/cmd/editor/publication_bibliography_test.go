package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPublicationBibliographySaveReopenEditAndClear(t *testing.T) {
	for _, publicationType := range []string{"international-journal", "domestic-journal", "international-conference", "domestic-conference"} {
		t.Run(publicationType, func(t *testing.T) {
			root := newTempRepoFixture(t, fixtureSettingsJSON())
			app := &App{repoRoot: root}
			loaded, err := app.LoadEditorData()
			if err != nil {
				t.Fatal(err)
			}
			input := publicationInputFromSummary(loaded.Publications[0])
			input.PublicationType = publicationType
			input.Pages = " S101–S110 "
			if strings.HasSuffix(publicationType, "-journal") {
				input.Volume = " 12A "
				input.Issue = " S1 "
				input.ELocator = " e106075 "
				input.OnlineDate = " 2024-02-29 "
				input.IssueDate = " 2026-10 "
			}
			expectedDate := effectivePublicationDate(publicationDocumentItem{PublicationType: publicationType, IssueDate: strings.TrimSpace(input.IssueDate), OnlineDate: strings.TrimSpace(input.OnlineDate)}, loaded.Settings)
			if _, err := app.SaveEditorData(publicationCatalogLinkRequest(loaded, input)); err != nil {
				t.Fatal(err)
			}
			loaded, err = app.LoadEditorData()
			if err != nil {
				t.Fatal(err)
			}
			var selected PublicationItemSummary
			for _, item := range loaded.Publications {
				if item.TitleEN == input.TitleEN {
					selected = item
				}
			}
			if selected.Pages != "S101–S110" || selected.Volume != strings.TrimSpace(input.Volume) || selected.Issue != strings.TrimSpace(input.Issue) || selected.ELocator != strings.TrimSpace(input.ELocator) || selected.OnlineDate != strings.TrimSpace(input.OnlineDate) || selected.IssueDate != strings.TrimSpace(input.IssueDate) {
				t.Fatalf("bibliographic values did not survive save/reopen: %#v", selected)
			}
			if selected.Date != expectedDate {
				t.Fatal("derived date does not follow bibliography priority")
			}

			for _, clear := range []bool{false, true} {
				request := saveRequestRevisions(loaded)
				request.SavePublications = true
				for _, item := range loaded.Publications {
					row := PublicationSaveItem{EditorKey: item.EditorKey}
					if item.TitleEN == input.TitleEN {
						row.Publication = publicationInputFromSummary(item)
						row.Publication.Note = "Unrelated note edit"
						if clear {
							row.Publication.Volume, row.Publication.Issue, row.Publication.Pages = "", "", ""
							row.Publication.ELocator, row.Publication.OnlineDate, row.Publication.IssueDate = "", "", ""
						}
					}
					request.Publications = append(request.Publications, row)
				}
				if _, err := app.SaveEditorData(request); err != nil {
					t.Fatal(err)
				}
				loaded, err = app.LoadEditorData()
				if err != nil {
					t.Fatal(err)
				}
				raw, err := os.ReadFile(filepath.Join(root, "data", "publications.json"))
				if err != nil {
					t.Fatal(err)
				}
				fields := publicationRawItemByTitle(t, raw, input.TitleEN)
				wantPages := "S101–S110"
				if clear {
					wantPages = ""
				}
				assertPublicationRawField(t, fields, "pages", wantPages)
				if strings.HasSuffix(publicationType, "-journal") {
					want := "2024-02-29"
					if clear {
						want = ""
					}
					assertPublicationRawField(t, fields, "online_date", want)
				}
			}
		})
	}
}

func TestPublicationBibliographyOptionalFieldsAndValidation(t *testing.T) {
	for _, field := range []string{"volume", "issue", "pages", "elocator", "online_date", "issue_date"} {
		for _, invalid := range []any{float64(3), nil, "bad\x00text", strings.Repeat("a", 501)} {
			row := publicationFixtureItem("Paper", "2026", false, false)
			row[field] = invalid
			raw, _ := json.Marshal([]any{row})
			if _, err := readPublicationsBytes(raw, publicationPersonIDs(), publicationAwardIDs(), publicationTopicIDs()); err == nil {
				t.Fatalf("invalid %s value accepted: %#v", field, invalid)
			}
		}
	}
	for _, field := range []string{"online_date", "issue_date"} {
		for _, value := range []string{"2024-02-29", "2026-10", "2026", ""} {
			row := publicationFixtureItem("Paper", "2026", false, false)
			row[field] = value
			raw, _ := json.Marshal([]any{row})
			if _, err := readPublicationsBytes(raw, publicationPersonIDs(), publicationAwardIDs(), publicationTopicIDs()); err != nil {
				t.Fatalf("valid %s rejected: %v", field, err)
			}
		}
		for _, value := range []string{"2025-02-29", "2026-13", "2026/10/18"} {
			row := publicationFixtureItem("Paper", "2026", false, false)
			row[field] = value
			raw, _ := json.Marshal([]any{row})
			if _, err := readPublicationsBytes(raw, publicationPersonIDs(), publicationAwardIDs(), publicationTopicIDs()); err == nil {
				t.Fatalf("invalid date accepted for %s: %s", field, value)
			}
		}
	}
}
