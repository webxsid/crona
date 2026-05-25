package meta

import (
	"strings"
	"testing"

	"crona/tui/internal/api"
	types "crona/tui/internal/tui/views/types"
)

func TestRenderIssuesShowsSpentSuffix(t *testing.T) {
	state := types.ContentState{
		Pane:   "issues",
		Width:  80,
		Height: 12,
		Issues: []api.Issue{
			{
				ID:            1,
				Title:         "Investigate timer display",
				Status:        "in_progress",
				WorkedSeconds: 4500,
			},
		},
	}

	rendered := renderIssues(types.Theme{}, state, 80, 12, "No issues")
	if !strings.Contains(rendered, "spent 1h15m") {
		t.Fatalf("expected spent suffix to render, got %q", rendered)
	}
}
