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
	// Unfocused, so the selected card keeps its meta line.
	state := types.ContentState{
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
		DailyTaskSection: "overdue",
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
	// Unfocused, so the selected card keeps its meta line.
	state := types.ContentState{
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

func groupedPlannedState(width int) types.ContentState {
	state := dailyState(width, 30)
	state.DailyTaskSection = "planned"
	state.DailyIssues = []api.Issue{
		{ID: 1, Title: "Ship release notes", Status: "ready", PinnedDaily: true, TodoForDate: new("2026-05-25")},
		{ID: 2, Title: "Fix sync conflict", Status: "ready", PinnedDaily: true},
		{ID: 3, Title: "Plan errands", Status: "planned", TodoForDate: new(state.DashboardDate)},
	}
	return state
}

func TestPlannedTabGroupsPinnedAbovePlanned(t *testing.T) {
	for _, width := range []int{80, 160} {
		plain := ansi.Strip(renderIssues(testTheme(), groupedPlannedState(width), width, 20))
		pinned, planned := strings.Index(plain, "Pinned 2"), strings.Index(plain, "Planned 1")
		if pinned < 0 || planned < 0 || pinned > strings.Index(plain, "Ship release notes") ||
			planned < strings.Index(plain, "Fix sync conflict") || planned > strings.Index(plain, "Plan errands") {
			t.Fatalf("width %d: expected Pinned 2 then its tasks, then Planned 1, got %q", width, plain)
		}
	}
}

func TestPlannedTabTagsPinnedTaskDates(t *testing.T) {
	plain := ansi.Strip(renderIssues(testTheme(), groupedPlannedState(80), 80, 20))
	assertContainsLine(t, strings.Split(plain, "\n"), "Ship release notes", "[")
	assertContainsLine(t, strings.Split(plain, "\n"), "Plan errands", "Plan errands")
	for _, line := range strings.Split(plain, "\n") {
		if strings.Contains(line, "Plan errands") && strings.Contains(line, "[") {
			t.Fatalf("expected no date on a task planned for the shown day, got %q", line)
		}
	}
}

func TestPlannedTabWithoutPinsHasNoHeadings(t *testing.T) {
	state := groupedPlannedState(80)
	state.DailyIssues = state.DailyIssues[2:]
	if plain := ansi.Strip(renderIssues(testTheme(), state, 80, 20)); strings.Contains(plain, "Pinned 0") || strings.Contains(plain, "Planned 1") {
		t.Fatalf("expected no group headings without pinned tasks, got %q", plain)
	}
}

func TestTaskTabsAreTodayOverdueUpcoming(t *testing.T) {
	plain := ansi.Strip(dailyTaskTitle(testTheme(), "planned"))
	if fieldsOf(plain) != "Planned Overdue Upcoming" {
		t.Fatalf("expected Planned / Overdue / Upcoming tabs, got %q", plain)
	}
}
