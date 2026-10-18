package daily

import (
	"strings"
	"testing"

	"crona/tui/internal/api"
	viewchrome "crona/tui/internal/tui/views/chrome"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/x/ansi"
)

func TestRenderIssuesUsesListAtEveryWidth(t *testing.T) {
	state := types.ContentState{
		Pane:   "issues",
		Width:  240,
		Height: 20,
		DailyIssues: []api.Issue{
			{
				ID:              1,
				Title:           "Investigate timer display",
				Status:          "in_progress",
				WorkedSeconds:   4500,
				EstimateMinutes: new(25),
			},
		},
		AllIssues: []api.IssueWithMeta{{
			Issue: api.Issue{
				ID:     1,
				Title:  "Investigate timer display",
				Status: "in_progress",
			},
			RepoName:   "Core",
			StreamName: "TUI",
		}},
		Filters: map[string]string{"issues": ""},
	}

	plain := ansi.Strip(renderIssues(types.Theme{}, state, 240, 20))
	if !strings.Contains(plain, "Core > TUI | 1h15m / 25m | in progress") {
		t.Fatalf("expected the two-line list at 240 columns, got %q", plain)
	}
	for _, unwanted := range []string{"Worked", "Repo", "Stream", "Est."} {
		if strings.Contains(plain, unwanted) {
			t.Fatalf("expected no table header %q, got %q", unwanted, plain)
		}
	}
}

func TestRenderIssuesKeepsDateVisibleWhenTitleTruncates(t *testing.T) {
	todo := "2099-03-20"
	state := types.ContentState{
		Pane:   "issues",
		Width:  80,
		Height: 20,
		Filters: map[string]string{
			"issues": "",
		},
		DailyTaskSection: "pinned",
		DailyIssues: []api.Issue{{
			ID:          1,
			Title:       "A very long issue title that should be truncated before its date is rendered because it exceeds the width",
			Status:      "planned",
			TodoForDate: &todo,
		}},
	}

	plain := ansi.Strip(renderIssues(types.Theme{}, state, 80, 20))
	if !strings.Contains(plain, "due 2099-03-20]") || !strings.Contains(plain, "…") {
		t.Fatalf("expected a truncated title with its date kept, got %q", plain)
	}
}

func TestRenderIssuesUsesCompactListLayoutBelowBreakpoint(t *testing.T) {
	todo := "2099-03-20"
	todo2 := "2099-03-21"
	state := types.ContentState{
		Pane:   "issues",
		Width:  70,
		Height: 20,
		DailyIssues: []api.Issue{
			{
				ID:              1,
				Title:           "Investigate timer display",
				Status:          "in_progress",
				WorkedSeconds:   4500,
				EstimateMinutes: new(25),
				TodoForDate:     &todo,
			},
			{
				ID:              2,
				Title:           "Ship compact issue list",
				Status:          "planned",
				WorkedSeconds:   0,
				EstimateMinutes: new(15),
				TodoForDate:     &todo2,
			},
		},
		AllIssues: []api.IssueWithMeta{{
			Issue: api.Issue{
				ID:              1,
				Title:           "Investigate timer display",
				Status:          "in_progress",
				WorkedSeconds:   4500,
				EstimateMinutes: new(25),
				TodoForDate:     &todo,
			},
			RepoName:   "Core",
			StreamName: "TUI",
		}, {
			Issue: api.Issue{
				ID:              2,
				Title:           "Ship compact issue list",
				Status:          "planned",
				WorkedSeconds:   0,
				EstimateMinutes: new(15),
				TodoForDate:     &todo2,
			},
			RepoName:   "Docs",
			StreamName: "web",
		}},
		Filters: map[string]string{"issues": ""},
	}

	rendered := renderIssues(types.Theme{}, state, 70, 20)
	plain := ansi.Strip(rendered)
	for _, want := range []string{
		"Investigate timer display",
		"[due 2099-03-20]",
		"Core > TUI | 1h15m / 25m | in progress",
		"Ship compact issue list",
		"[due 2099-03-21]",
		"Docs > web | - / 15m | planned",
	} {
		if !strings.Contains(plain, want) {
			t.Fatalf("expected compact list layout to contain %q, got %q", want, rendered)
		}
	}
	if strings.Contains(plain, "*") {
		t.Fatalf("expected compact list layout to avoid asterisk separators, got %q", rendered)
	}
	for _, unwanted := range []string{"Context", "Effort", "Worked", "Repo", "Stream"} {
		if strings.Contains(plain, unwanted) {
			t.Fatalf("expected list layout to omit table header %q, got %q", unwanted, rendered)
		}
	}
}

func normalizeTableLine(line string) string {
	line = strings.TrimLeft(line, " ")
	line = strings.TrimPrefix(line, viewchrome.SelectionCursor+" ")
	line = strings.TrimPrefix(line, "  ")
	line = strings.TrimLeft(line, " ")
	return line
}
