package dialogs

import (
	"strings"
	"testing"

	sharedtypes "crona/shared/types"
	controllerpkg "crona/tui/internal/tui/dialogs/controller"
)

func TestIssueStatusDialogShowsSelectedIssueContext(t *testing.T) {
	estimate := 120
	state := controllerpkg.State{
		Kind:               "issue_status",
		Width:              100,
		ViewName:           "Ship status dialog context",
		RepoName:           "Crona",
		StreamName:         "TUI",
		IssueStatus:        string(sharedtypes.IssueStatusInProgress),
		IssueEstimateMins:  &estimate,
		IssueWorkedSeconds: 3_600,
		StatusItems:        []sharedtypes.IssueStatus{sharedtypes.IssueStatusDone},
	}

	rendered := renderIssueDialog(testTheme(), state)
	for _, want := range []string{
		"Set Issue Status",
		"Ship status dialog context",
		"Crona",
		"TUI",
		"in progress",
		"worked 1h / est. 2h",
		"done",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("expected dialog to contain %q, got %q", want, rendered)
		}
	}
}
