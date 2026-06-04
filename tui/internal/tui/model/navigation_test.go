package model

import (
	"strings"
	"testing"

	"crona/tui/internal/api"
)

func TestOpenSelectedViewDialogIncludesWorkedEstimate(t *testing.T) {
	issue := api.Issue{
		ID:            1,
		Title:         "Investigate timer display",
		Status:        "in_progress",
		WorkedSeconds: 0,
	}
	m := Model{
		view:   ViewDefault,
		pane:   PaneIssues,
		cursor: map[Pane]int{PaneIssues: 0},
		issues: []api.Issue{issue},
		allIssues: []api.IssueWithMeta{{
			Issue: api.Issue{
				ID:            1,
				Title:         "Investigate timer display",
				Status:        "in_progress",
				WorkedSeconds: 4500,
			},
			RepoName:   "Core",
			StreamName: "TUI",
		}},
	}

	next, ok := m.openSelectedViewDialog()
	if !ok {
		t.Fatal("expected issue details dialog to open")
	}
	if next.dialog != "view_entity" {
		t.Fatalf("expected view entity dialog, got %q", next.dialog)
	}
	if !strings.Contains(next.dialogViewMeta, "worked 1h15m / est. -") {
		t.Fatalf("expected dialog meta to include worked/estimate summary, got %q", next.dialogViewMeta)
	}
	if !strings.Contains(next.dialogViewBody, "Worked / est.") ||
		!strings.Contains(next.dialogViewBody, "1h15m / est. -") {
		t.Fatalf("expected dialog body to include worked/estimate summary, got %q", next.dialogViewBody)
	}
}
