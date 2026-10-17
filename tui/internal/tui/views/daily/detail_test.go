package daily

import (
	"strings"
	"testing"

	"crona/tui/internal/api"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/x/ansi"
)

func detailState() (state types.ContentState) {
	state = dailyState(200, 40)
	state.DailyTaskSection = "planned"
	description := "Make the sync resolve conflicts by keeping both edits"
	state.DailyIssues = []api.Issue{
		{ID: 1, Title: "First task", Status: "ready", TodoForDate: new(state.DashboardDate)},
		{
			ID:              2,
			Title:           "Fix sync conflict",
			Status:          "in_progress",
			TodoForDate:     new(state.DashboardDate),
			EstimateMinutes: new(60),
			WorkedSeconds:   1800,
			Description:     &description,
			RepoName:        "crona",
			StreamName:      "core",
		},
	}
	state.Cursors["issues"] = 1
	return state
}

func TestDetailShowsSelectedTask(t *testing.T) {
	rendered := ansi.Strip(renderDetail(testTheme(), detailState(), 70, 30))
	for _, want := range []string{"─ Selected", "Fix sync conflict", "in progress", "crona › core", "30m/1h", "keeping both edits"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("expected detail to contain %q, got %q", want, rendered)
		}
	}
	if strings.Contains(rendered, "First task") {
		t.Fatalf("expected only the selected task, got %q", rendered)
	}
}

func TestDetailFollowsHabitsFocus(t *testing.T) {
	state := detailState()
	state.Pane = "habits"
	state.Cursors["habits"] = 1
	target := 20
	state.DueHabits[1].TargetMinutes = &target
	rendered := ansi.Strip(renderDetail(testTheme(), state, 70, 30))
	if !strings.Contains(rendered, "Read") {
		t.Fatalf("expected selected habit details, got %q", rendered)
	}
	assertContainsLine(t, strings.Split(rendered, "\n"), "target", "20m")
	if strings.Contains(rendered, "Fix sync conflict") {
		t.Fatalf("expected habit focus to replace task detail, got %q", rendered)
	}
}

func TestDetailEmptySelection(t *testing.T) {
	state := detailState()
	state.DailyIssues = nil
	rendered := ansi.Strip(renderDetail(testTheme(), state, 70, 30))
	if !strings.Contains(rendered, "Nothing selected") {
		t.Fatalf("expected empty-selection hint, got %q", rendered)
	}
}

func TestDetailShowsCalendarWhenTall(t *testing.T) {
	state := detailState()
	tall := ansi.Strip(renderDetail(testTheme(), state, 70, 40))
	if !strings.Contains(tall, "Mo Tu We Th Fr Sa Su") {
		t.Fatalf("expected calendar in a tall detail pane, got %q", tall)
	}
	short := ansi.Strip(renderDetail(testTheme(), state, 70, 12))
	if strings.Contains(short, "Mo Tu We Th Fr Sa Su") {
		t.Fatalf("expected no calendar in a short detail pane, got %q", short)
	}
	if lines := strings.Split(short, "\n"); len(lines) != 12 {
		t.Fatalf("expected 12 rows, got %d", len(lines))
	}
}
