package issues

import (
	"strings"
	"testing"

	"crona/tui/internal/api"
	types "crona/tui/internal/tui/views/types"
	"github.com/charmbracelet/x/ansi"
)

func TestRenderIssuePaneShowsSpentSuffix(t *testing.T) {
	state := types.ContentState{
		Pane:   "issues",
		Width:  120,
		Height: 20,
		DefaultIssues: []api.IssueWithMeta{{
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

	rendered := renderIssuePane(types.Theme{}, state, "Active Issues [1]", "Due work and open issues", []int{0}, 0, true, 20, "No open issues match the current filter", true)
	plain := ansi.Strip(rendered)
	if !strings.Contains(plain, "Spent") {
		t.Fatalf("expected spent column header to render, got %q", rendered)
	}
	if !strings.Contains(plain, "1h15m") {
		t.Fatalf("expected worked seconds to render in the spent column, got %q", rendered)
	}
	lines := strings.Split(plain, "\n")
	var headerLine, rowLine string
	for _, line := range lines {
		if strings.Contains(line, "Status") && strings.Contains(line, "Estimate") && strings.Contains(line, "Spent") {
			headerLine = line
		}
		if strings.Contains(line, "in progress") {
			rowLine = line
		}
	}
	if headerLine == "" || rowLine == "" {
		t.Fatalf("expected header and row lines in %q", plain)
	}
	headerLine = normalizeTableLine(headerLine)
	rowLine = normalizeTableLine(rowLine)
	if strings.Index(headerLine, "Status") != strings.Index(rowLine, "in progress") {
		t.Fatalf("expected status header and value to align, got header %q row %q", headerLine, rowLine)
	}
	if strings.Index(headerLine, "Spent") != strings.Index(rowLine, "1h15m") {
		t.Fatalf("expected spent header and value to align, got header %q row %q", headerLine, rowLine)
	}
}

func normalizeTableLine(line string) string {
	line = strings.TrimLeft(line, " ")
	line = strings.TrimPrefix(line, "▶ ")
	line = strings.TrimPrefix(line, "  ")
	line = strings.TrimLeft(line, " ")
	return line
}
