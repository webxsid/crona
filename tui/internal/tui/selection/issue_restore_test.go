package selection

import (
	sharedtypes "crona/shared/types"
	"testing"

	"crona/tui/internal/api"

	uistate "crona/tui/internal/tui/state"
)

func TestRegressionRestoreReorderedIssue(t *testing.T) {
	for _, section := range []uistate.DefaultIssueSection{uistate.DefaultIssueSectionOpen, uistate.DefaultIssueSectionCompleted} {
		for _, filter := range []string{"", "Alpha"} {
			t.Run(string(section)+"/"+filter, func(t *testing.T) {
				status := "ready"
				if section == uistate.DefaultIssueSectionCompleted {
					status = "done"
				}
				s := PrepareSnapshot(
					Snapshot{
						View:                uistate.ViewDefault,
						Pane:                uistate.PaneIssues,
						DefaultIssueSection: section,
						Cursors:             map[uistate.Pane]int{},
						Filters:             map[uistate.Pane]string{uistate.PaneIssues: filter},
						AllIssues: []api.IssueWithMeta{
							{
								Issue: api.Issue{
									ID:     1,
									Title:  "Zulu",
									Status: sharedtypes.IssueStatus(status),
								},
							}, {Issue: api.Issue{ID: 2, Title: "Alpha", Status: sharedtypes.IssueStatus(status)}}, {Issue: api.Issue{ID: 3, Title: "Other section", Status: "abandoned"}},
						},
					},
				)
				s.Cursors[uistate.PaneIssues] = FilteredCursorForIssueID(s, uistate.PaneIssues, 2)
				got, ok := SelectedIssue(s)
				if !ok || got.ID != 2 {
					t.Fatalf(
						"restore ID 2: cursor %d selected %+v; indices %v",
						s.Cursors[uistate.PaneIssues],
						got,
						FilteredIndices(s, uistate.PaneIssues),
					)
				}
				if filter != "" && FilteredCursorForIssueID(s, uistate.PaneIssues, 1) != -1 {
					t.Fatal("filtered issue restored")
				}
			})
		}
	}
}
