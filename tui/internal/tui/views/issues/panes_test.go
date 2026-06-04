package issues

import (
	"strings"
	"testing"

	"crona/tui/internal/api"
	viewchrome "crona/tui/internal/tui/views/chrome"
	types "crona/tui/internal/tui/views/types"
	"github.com/charmbracelet/x/ansi"
)

func TestRenderIssuePaneShowsWorkedSuffix(t *testing.T) {
	state := types.ContentState{
		Pane:   "issues",
		Width:  120,
		Height: 20,
		DefaultIssues: []api.IssueWithMeta{{
			Issue: api.Issue{
				ID:              1,
				Title:           "Investigate timer display",
				Status:          "in_progress",
				WorkedSeconds:   4500,
				EstimateMinutes: ptrInt(25),
			},
			RepoName:   "Core",
			StreamName: "TUI",
		}},
	}

	rendered := renderIssuePane(
		types.Theme{},
		state,
		"Active Issues [1]",
		"Due work and open issues",
		[]int{0},
		0,
		true,
		20,
		"No open issues match the current filter",
		true,
	)
	plain := ansi.Strip(rendered)
	if !strings.Contains(plain, "Worked") {
		t.Fatalf("expected worked column header to render, got %q", rendered)
	}
	if !strings.Contains(plain, "1h15m") {
		t.Fatalf("expected worked seconds to render in the spent column, got %q", rendered)
	}
	lines := strings.Split(plain, "\n")
	var headerLine, rowLine string
	for _, line := range lines {
		if strings.Contains(line, "Status") && strings.Contains(line, "Est.") &&
			strings.Contains(line, "Worked") {
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
		t.Fatalf(
			"expected status header and value to align, got header %q row %q",
			headerLine,
			rowLine,
		)
	}
	if strings.Index(headerLine, "Worked") != strings.Index(rowLine, "1h15m") {
		t.Fatalf(
			"expected worked header and value to align, got header %q row %q",
			headerLine,
			rowLine,
		)
	}
}

func TestRenderIssuePaneUsesCompactContextAndEffortColumns(t *testing.T) {
	state := types.ContentState{
		Pane:   "issues",
		Width:  84,
		Height: 20,
		DefaultIssues: []api.IssueWithMeta{{
			Issue: api.Issue{
				ID:              1,
				Title:           "Investigate timer display",
				Status:          "in_progress",
				WorkedSeconds:   4500,
				EstimateMinutes: ptrInt(25),
			},
			RepoName:   "Core",
			StreamName: "TUI",
		}},
	}

	rendered := renderIssuePane(
		types.Theme{},
		state,
		"Active Issues [1]",
		"Due work and open issues",
		[]int{0},
		0,
		true,
		20,
		"No open issues match the current filter",
		true,
	)
	plain := ansi.Strip(rendered)
	for _, want := range []string{"Context", "Effort", "Core > TUI", "1h15m / 25m"} {
		if !strings.Contains(plain, want) {
			t.Fatalf("expected compact issue table to contain %q, got %q", want, rendered)
		}
	}
}

func TestRenderIssuePaneCollapsesEmptyWorkedEffort(t *testing.T) {
	state := types.ContentState{
		Pane:   "issues",
		Width:  84,
		Height: 20,
		DefaultIssues: []api.IssueWithMeta{{
			Issue: api.Issue{
				ID:              1,
				Title:           "Investigate timer display",
				Status:          "in_progress",
				WorkedSeconds:   0,
				EstimateMinutes: ptrInt(25),
			},
			RepoName:   "Core",
			StreamName: "TUI",
		}},
		Filters: map[string]string{"issues": ""},
	}

	rendered := renderIssuePane(
		types.Theme{},
		state,
		"Active Issues [1]",
		"Due work and open issues",
		[]int{0},
		0,
		true,
		20,
		"No open issues match the current filter",
		true,
	)
	plain := ansi.Strip(rendered)
	if !strings.Contains(plain, "Effort") || !strings.Contains(plain, "-") {
		t.Fatalf("expected compact issue table to collapse empty effort to a dash, got %q", rendered)
	}
}

func normalizeTableLine(line string) string {
	line = strings.TrimLeft(line, " ")
	line = strings.TrimPrefix(line, viewchrome.SelectionCursor+" ")
	line = strings.TrimPrefix(line, "  ")
	line = strings.TrimLeft(line, " ")
	return line
}

func ptrInt(value int) *int {
	return &value
}
