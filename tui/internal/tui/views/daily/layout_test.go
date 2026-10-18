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
		// Habit cards are the one boxed element, so leave them out here.
		state := dailyState(width, 40)
		state.DueHabits = nil
		rendered := renderView(testTheme(), state)
		if strings.ContainsAny(rendered, "╭╰") {
			t.Fatalf("width %d: expected no box borders, got %q", width, ansi.Strip(rendered))
		}
	}
}

func TestDailyStripThenTasksThenHabits(t *testing.T) {
	for _, width := range []int{99, 120, 160} {
		rendered := renderView(testTheme(), dailyState(width, 40))
		lines := strings.Split(ansi.Strip(rendered), "\n")
		if !strings.HasPrefix(fieldsOf(lines[0]), "Tasks") || !strings.HasPrefix(fieldsOf(lines[1]), "Habits 0/2") {
			t.Fatalf("width %d: expected the day strip on rows 0-1, got %q", width, lines[:2])
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

func TestDailySplitsTasksAndHabitsEvenly(t *testing.T) {
	for _, width := range []int{80, 130, 200} {
		for _, tasks := range []int{1, 3, 30} {
			state := manyIssues(dailyState(width, 40), tasks)
			rendered := renderView(testTheme(), state)
			strip := len(renderDayStrip(testTheme(), state, width))
			tasksH := rowOf(t, rendered, "─ Habits Due") - rowOf(t, rendered, "─ Tasks")
			habitsH := 40 - strip - tasksH
			if tasksH-habitsH < 0 || tasksH-habitsH > 1 {
				t.Fatalf("width %d, %d tasks: expected an even split, got tasks %d habits %d", width, tasks, tasksH, habitsH)
			}
		}
	}
}

func TestDailyHabitsAreChipsUntilFocused(t *testing.T) {
	state := dailyState(80, 40)
	assertContainsLine(t, strings.Split(ansi.Strip(renderView(testTheme(), state)), "\n"), "[ ] Walk", "[ ] Read")
	state.Pane = "habits"
	for _, line := range strings.Split(ansi.Strip(renderView(testTheme(), state)), "\n") {
		if strings.Contains(line, "Walk") && strings.Contains(line, "Read") {
			t.Fatalf("expected focused habits as a vertical list, got %q", line)
		}
	}
}

func cardHabits(state types.ContentState, n int) types.ContentState {
	state.DueHabits = nil
	for i := range n {
		state.DueHabits = append(state.DueHabits, api.HabitDailyItem{
			HabitWithMeta: api.HabitWithMeta{Habit: api.Habit{ID: int64(i + 1), Name: fmt.Sprintf("Habit %02d", i)}},
		})
	}
	return state
}

func TestDailyHabitsAreCardsWhenRoomy(t *testing.T) {
	state := cardHabits(dailyState(130, 40), 6)
	unfocused := ansi.Strip(renderHabits(testTheme(), state, 130, 14))
	if !strings.Contains(unfocused, "╭") {
		t.Fatalf("expected habit cards at 130 columns, got %q", unfocused)
	}
	state.Pane = "habits"
	focused := ansi.Strip(renderHabits(testTheme(), state, 130, 14))
	if rowOf(t, focused, "Habit 00") != rowOf(t, unfocused, "Habit 00") || rowOf(t, focused, "Habit 01") != rowOf(t, focused, "Habit 00") {
		t.Fatalf("expected the same card grid focused and unfocused, got %q / %q", unfocused, focused)
	}
}

func TestDailyHabitCardsPerRowFollowWidth(t *testing.T) {
	for width, want := range map[int]int{90: 2, 100: 3, 130: 4, 200: 4} {
		state := cardHabits(dailyState(width, 40), 6)
		lines := strings.Split(ansi.Strip(renderHabits(testTheme(), state, width, 14)), "\n")
		if got := strings.Count(lines[1], "╭"); got != want {
			t.Fatalf("width %d: expected %d cards per row, got %d: %q", width, want, got, lines[1])
		}
		for i, line := range lines {
			if w := ansi.StringWidth(line); w != width {
				t.Fatalf("width %d: row %d is %d cells: %q", width, i, w, line)
			}
		}
	}
}

func TestDailyHabitCardsMarkFocusedSelection(t *testing.T) {
	state := cardHabits(dailyState(130, 40), 6)
	state.Pane = "habits"
	state.Cursors["habits"] = 5
	rendered := ansi.Strip(renderHabits(testTheme(), state, 130, 14))
	assertContainsLine(t, strings.Split(rendered, "\n"), "▶", "Habit 05")
	if strings.Count(rendered, "▶") != 1 {
		t.Fatalf("expected one selected card, got %q", rendered)
	}
}

func TestDailyHabitCardsScrollToCursor(t *testing.T) {
	state := cardHabits(dailyState(130, 40), 30)
	state.Pane = "habits"
	state.Cursors["habits"] = 29
	rendered := ansi.Strip(renderHabits(testTheme(), state, 130, 10))
	if !strings.Contains(rendered, "Habit 29") || !strings.Contains(rendered, "↑") {
		t.Fatalf("expected the last card and a ↑ marker, got %q", rendered)
	}
	if lines := strings.Split(rendered, "\n"); len(lines) != 10 {
		t.Fatalf("expected 10 rows, got %d", len(lines))
	}
}

func TestDailyHabitChipsWrap(t *testing.T) {
	state := dailyState(60, 40)
	state.DueHabits = nil
	for i := range 12 {
		state.DueHabits = append(state.DueHabits, api.HabitDailyItem{
			HabitWithMeta: api.HabitWithMeta{Habit: api.Habit{ID: int64(i + 1), Name: fmt.Sprintf("Habit %02d", i)}},
		})
	}
	rendered := ansi.Strip(renderHabits(testTheme(), state, 60, 10))
	for _, want := range []string{"Habit 00", "Habit 11"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("expected wrapped chips to include %q, got %q", want, rendered)
		}
	}
	for i, line := range strings.Split(rendered, "\n") {
		if w := ansi.StringWidth(line); w != 60 {
			t.Fatalf("row %d is %d cells: %q", i, w, line)
		}
	}
}

func TestDailyCompactTasksHaveNoDashDividers(t *testing.T) {
	rendered := ansi.Strip(renderIssues(testTheme(), manyIssues(dailyState(80, 30), 3), 80, 20))
	if strings.Contains(rendered, "----") {
		t.Fatalf("expected no dash dividers between tasks, got %q", rendered)
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

func TestDailyTasksUseIssueContextWithoutMeta(t *testing.T) {
	for _, width := range []int{80, 130} {
		state := manyIssues(dailyState(width, 30), 1)
		state.DailyIssues[0].RepoName, state.DailyIssues[0].StreamName = "crona", "core"
		if rendered := ansi.Strip(renderIssues(testTheme(), state, width, 20)); !strings.Contains(rendered, "crona") || !strings.Contains(rendered, "core") {
			t.Fatalf("width %d: expected the issue's own repo and stream, got %q", width, rendered)
		}
	}
}
