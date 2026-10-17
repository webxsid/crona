package daily

import (
	"fmt"
	"strings"
	"testing"

	"crona/tui/internal/api"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/x/ansi"
)

func dailyState(width, height int) types.ContentState {
	state := dailySummaryTestState()
	state.Width = width
	state.Height = height
	state.Cursors = map[string]int{"issues": 0, "habits": 0}
	state.Filters = map[string]string{"issues": "", "habits": ""}
	state.DueHabits = []api.HabitDailyItem{
		{HabitWithMeta: api.HabitWithMeta{Habit: api.Habit{ID: 1, Name: "Walk"}}},
		{HabitWithMeta: api.HabitWithMeta{Habit: api.Habit{ID: 2, Name: "Read"}}},
	}
	return state
}

func rowOf(t *testing.T, rendered, needle string) int {
	t.Helper()
	for i, line := range strings.Split(ansi.Strip(rendered), "\n") {
		if strings.Contains(line, needle) {
			return i
		}
	}
	t.Fatalf("%q not found in %q", needle, ansi.Strip(rendered))
	return -1
}

func TestDailyHasNoBorders(t *testing.T) {
	for _, width := range []int{80, 120, 200} {
		rendered := renderView(testTheme(), dailyState(width, 40))
		if strings.ContainsAny(rendered, "╭╰") {
			t.Fatalf("width %d: expected no box borders, got %q", width, ansi.Strip(rendered))
		}
	}
}

func TestDailyStripThenTasksThenHabits(t *testing.T) {
	for _, width := range []int{99, 120, 160} {
		rendered := renderView(testTheme(), dailyState(width, 40))
		first := strings.Split(ansi.Strip(rendered), "\n")[0]
		if !strings.Contains(first, "Habits 0/2") {
			t.Fatalf("width %d: expected the day strip on row 0, got %q", width, first)
		}
		if rowOf(t, rendered, "─ Habits Due") <= rowOf(t, rendered, "─ Tasks") {
			t.Fatalf("width %d: expected habits below tasks", width)
		}
	}
}

func TestDailyWideHasDetailPane(t *testing.T) {
	wide := renderView(testTheme(), dailyState(160, 40))
	if rowOf(t, wide, "─ Selected") != rowOf(t, wide, "─ Tasks") {
		t.Fatal("expected the detail pane beside tasks at 160 columns")
	}
	if medium := ansi.Strip(renderView(testTheme(), dailyState(159, 40))); strings.Contains(medium, "─ Selected") {
		t.Fatal("expected no detail pane below 160 columns")
	}
}

func TestDailyHabitsSizedToContent(t *testing.T) {
	rendered := renderView(testTheme(), dailyState(120, 60))
	lines := strings.Split(ansi.Strip(rendered), "\n")
	header := rowOf(t, rendered, "─ Habits Due")
	if !strings.Contains(lines[header+2], "Read") || strings.TrimSpace(lines[header+3]) != "" {
		t.Fatalf("expected Habits to hold just its two rows, got %q", lines[header:header+4])
	}
}

func TestDailyNoActionRowsInSections(t *testing.T) {
	for _, pane := range []string{"issues", "habits"} {
		state := dailyState(120, 40)
		state.Pane = pane
		if rendered := ansi.Strip(renderView(testTheme(), state)); strings.Contains(rendered, "[enter]") {
			t.Fatalf("%s focused: expected key hints to live in the status line, got %q", pane, rendered)
		}
	}
}

func TestDailyFocusedHabitsCursorVisible(t *testing.T) {
	state := dailyState(120, 30)
	state.Pane = "habits"
	state.DueHabits = nil
	for i := range 30 {
		state.DueHabits = append(state.DueHabits, api.HabitDailyItem{
			HabitWithMeta: api.HabitWithMeta{Habit: api.Habit{ID: int64(i + 1), Name: fmt.Sprintf("Habit %02d", i)}},
		})
	}
	state.Cursors["habits"] = 29
	if rendered := ansi.Strip(renderView(testTheme(), state)); !strings.Contains(rendered, "Habit 29") {
		t.Fatalf("expected the focused habit cursor to stay visible, got %q", rendered)
	}
}

func TestDailyIssuesCursorVisibleInShortSection(t *testing.T) {
	state := dailyState(80, 18)
	state.DailyIssues = nil
	for i := range 30 {
		state.DailyIssues = append(state.DailyIssues, api.Issue{
			ID:          int64(i + 1),
			Title:       fmt.Sprintf("Issue number %02d", i),
			Status:      "ready",
			TodoForDate: new(state.DashboardDate),
		})
	}
	state.DailyTaskSection = "planned"
	state.Cursors["issues"] = 29
	rendered := ansi.Strip(renderIssues(testTheme(), state, 80, 18))
	if !strings.Contains(rendered, "Issue number 29") || !strings.Contains(rendered, "↑") {
		t.Fatalf("expected cursor row and ↑ marker, got %q", rendered)
	}
}

func TestDailyFitsFrame(t *testing.T) {
	for _, size := range [][2]int{{60, 18}, {99, 30}, {159, 40}, {200, 47}, {285, 73}} {
		rendered := renderView(testTheme(), dailyState(size[0], size[1]))
		lines := strings.Split(rendered, "\n")
		if len(lines) != size[1] {
			t.Fatalf("%dx%d: got %d lines", size[0], size[1], len(lines))
		}
		for i, line := range lines {
			if got := ansi.StringWidth(line); got > size[0] {
				t.Fatalf("%dx%d: line %d width %d", size[0], size[1], i, got)
			}
		}
	}
}

func manyIssues(state types.ContentState, n int) types.ContentState {
	state.DailyTaskSection = "planned"
	state.DailyIssues = nil
	for i := range n {
		state.DailyIssues = append(state.DailyIssues, api.Issue{
			ID:          int64(i + 1),
			Title:       fmt.Sprintf("Issue number %02d", i),
			Status:      "ready",
			TodoForDate: new(state.DashboardDate),
		})
	}
	return state
}

func TestDailyHabitsCursorVisibleAtEnd(t *testing.T) {
	state := dailyState(120, 24)
	state.Pane = "habits"
	state.DueHabits = nil
	for i := range 30 {
		state.DueHabits = append(state.DueHabits, api.HabitDailyItem{
			HabitWithMeta: api.HabitWithMeta{Habit: api.Habit{ID: int64(i + 1), Name: fmt.Sprintf("Habit %02d", i)}},
		})
	}
	state.Cursors["habits"] = 29
	rendered := ansi.Strip(renderHabits(testTheme(), state, 60, 20))
	if !strings.Contains(rendered, "Habit 29") || !strings.Contains(rendered, "↑ ") {
		t.Fatalf("expected last habit and ↑ marker, got %q", rendered)
	}
}

func TestDailyTableCursorVisibleAtEnd(t *testing.T) {
	state := manyIssues(dailyState(240, 24), 30)
	state.Cursors["issues"] = 29
	rendered := ansi.Strip(renderIssues(testTheme(), state, 240, 24))
	if !strings.Contains(rendered, "Issue number 29") {
		t.Fatalf("expected selected table row to be visible, got %q", rendered)
	}
}

func TestDailyNarrowShowsSelectedTask(t *testing.T) {
	state := manyIssues(dailyState(60, 18), 30)
	state.Cursors["issues"] = 29
	rendered := ansi.Strip(renderView(testTheme(), state))
	if !strings.Contains(rendered, "Issue number 29") {
		t.Fatalf("expected selected task visible at 60x18, got %q", rendered)
	}
}

func TestDailyShortTaskListKeepsHabitsClose(t *testing.T) {
	for _, width := range []int{80, 130, 200} {
		rendered := renderView(testTheme(), manyIssues(dailyState(width, 36), 3))
		if gap := rowOf(t, rendered, "─ Habits Due") - rowOf(t, rendered, "─ Tasks"); gap > 12 {
			t.Fatalf("width %d: expected habits right below 3 tasks, gap %d rows", width, gap)
		}
	}
}

func TestDailyTasksUseIssueContextWithoutMeta(t *testing.T) {
	for _, width := range []int{80, 130} {
		state := manyIssues(dailyState(width, 30), 1)
		state.DailyIssues[0].RepoName, state.DailyIssues[0].StreamName = "crona", "core"
		if rendered := ansi.Strip(renderIssues(testTheme(), state, width, 20)); !strings.Contains(rendered, "crona") || !strings.Contains(rendered, "core") {
			t.Fatalf("width %d: expected the issue's own repo and stream, got %q", width, rendered)
		}
	}
}
