package daily

import (
	"fmt"
	"strings"
	"testing"

	"crona/tui/internal/api"
	uistate "crona/tui/internal/tui/state"
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
		lines := strings.Split(ansi.Strip(rendered), "\n")
		if !strings.HasPrefix(fieldsOf(lines[0]), "Tasks") ||
			!strings.HasPrefix(fieldsOf(lines[1]), "Habits 0/2") {
			t.Fatalf("width %d: expected the day strip on rows 0-1, got %q", width, lines[:2])
		}
		if rowOf(t, rendered, "─ Habits Due") <= rowOf(t, rendered, "─ Tasks") {
			t.Fatalf("width %d: expected habits below tasks", width)
		}
	}
}

func TestDailyWideHasDetailPane(t *testing.T) {
	wide := renderView(testTheme(), dailyState(160, 40))
	if rowOf(t, wide, "─ Day ") != rowOf(t, wide, "─ Tasks") {
		t.Fatal("expected the detail pane beside tasks at 160 columns")
	}
	if medium := ansi.Strip(renderView(testTheme(), dailyState(159, 40))); strings.Contains(
		medium,
		"─ Day ",
	) {
		t.Fatal("expected no detail pane below 160 columns")
	}
}

func TestDailySplitsTasksAndHabitsEvenly(t *testing.T) {
	for _, width := range []int{80, 130, 200} {
		for _, tasks := range []int{1, 3, 30} {
			state := manyIssues(dailyState(width, 40), tasks)
			rendered := renderView(testTheme(), state)
			tasksH := rowOf(t, rendered, "─ Habits Due") - rowOf(t, rendered, "─ Tasks")
			habitsH := 40 - len(renderDayStrip(testTheme(), state, width)) - tasksH
			if tasksH-habitsH < 0 || tasksH-habitsH > 1 {
				t.Fatalf(
					"width %d, %d tasks: expected an even split, got tasks %d habits %d",
					width,
					tasks,
					tasksH,
					habitsH,
				)
			}
		}
	}
}

func cardHabits(state types.ContentState, n int) types.ContentState {
	state.DueHabits = nil
	for i := range n {
		state.DueHabits = append(state.DueHabits, api.HabitDailyItem{
			HabitWithMeta: api.HabitWithMeta{
				Habit: api.Habit{ID: int64(i + 1), Name: fmt.Sprintf("Habit %02d", i)},
			},
		})
	}
	return state
}

func meterHabitState(width int) types.ContentState {
	state := dailyState(width, 40)
	target, logged := 10, 15
	state.DueHabits = []api.HabitDailyItem{
		{
			HabitWithMeta: api.HabitWithMeta{
				Habit: api.Habit{ID: 1, Name: "Walk", TargetMinutes: &target},
			},
		},
		{
			HabitWithMeta: api.HabitWithMeta{
				Habit: api.Habit{ID: 2, Name: "Journal", TargetMinutes: &target},
			},
			Status:          "completed",
			DurationMinutes: &logged,
		},
		{
			HabitWithMeta: api.HabitWithMeta{Habit: api.Habit{ID: 3, Name: "Morning Walk"}},
			Status:        "failed",
		},
	}
	return state
}

func TestDailyHabitRowsShowStatusBarAndMinutes(t *testing.T) {
	lines := strings.Split(
		ansi.Strip(renderHabits(testTheme(), meterHabitState(130), 130, 10)),
		"\n",
	)
	assertContainsLine(t, lines, "[x] Journal", "15m / 10m")
	assertContainsLine(t, lines, "[x] Journal", "█")
	assertContainsLine(t, lines, "[ ] Walk", "- / 10m")
	assertContainsLine(t, lines, "[!] Morning Walk", "failed")
	walk, journal := lines[rowOf(t, strings.Join(lines, "\n"), "Walk")], lines[rowOf(t, strings.Join(lines, "\n"), "Journal")]
	if cellIndex(walk, "due") != cellIndex(journal, "done") || cellIndex(walk, "due") < 0 {
		t.Fatalf("expected the status column aligned, got %q / %q", walk, journal)
	}
}

func TestDailyHabitRowsSameFocusedOrNot(t *testing.T) {
	state := meterHabitState(130)
	unfocused := ansi.Strip(renderHabits(testTheme(), state, 130, 10))
	state.Pane = "habits"
	focused := ansi.Strip(renderHabits(testTheme(), state, 130, 10))
	// Rows below the selection shift for its action line; the selection itself stays put.
	for _, name := range []string{"[ ] Walk"} {
		if rowOf(t, focused, name) != rowOf(t, unfocused, name) {
			t.Fatalf("expected %q on the same row focused or not", name)
		}
	}
	if strings.Contains(unfocused, "▶") || !strings.Contains(focused, "▶") {
		t.Fatalf("expected the cursor only when focused, got %q / %q", unfocused, focused)
	}
}

func TestDailyHabitRowsFitNarrowWidths(t *testing.T) {
	for _, width := range []int{40, 60, 90} {
		rendered := ansi.Strip(renderHabits(testTheme(), meterHabitState(width), width, 10))
		if !strings.Contains(rendered, "Morning Walk") {
			t.Fatalf("width %d: expected habit names kept, got %q", width, rendered)
		}
		for i, line := range strings.Split(rendered, "\n") {
			if w := ansi.StringWidth(line); w != width {
				t.Fatalf("width %d: row %d is %d cells: %q", width, i, w, line)
			}
		}
	}
}

func TestDailyHabitRowsMarkFocusedSelection(t *testing.T) {
	state := cardHabits(dailyState(130, 40), 6)
	state.Pane = "habits"
	state.Cursors["habits"] = 5
	rendered := ansi.Strip(renderHabits(testTheme(), state, 130, 14))
	assertContainsLine(t, strings.Split(rendered, "\n"), "▶", "Habit 05")
}

func TestDailyCompactTasksHaveNoDashDividers(t *testing.T) {
	rendered := ansi.Strip(renderIssues(testTheme(), manyIssues(dailyState(80, 30), 3), 80, 20))
	if strings.Contains(rendered, "----") {
		t.Fatalf("expected no dash dividers between tasks, got %q", rendered)
	}
}

func TestDailyFocusedHabitsCursorVisible(t *testing.T) {
	state := dailyState(120, 30)
	state.Pane = "habits"
	state.DueHabits = nil
	for i := range 30 {
		state.DueHabits = append(state.DueHabits, api.HabitDailyItem{
			HabitWithMeta: api.HabitWithMeta{
				Habit: api.Habit{ID: int64(i + 1), Name: fmt.Sprintf("Habit %02d", i)},
			},
		})
	}
	state.Cursors["habits"] = 29
	if rendered := ansi.Strip(renderView(testTheme(), state)); !strings.Contains(
		rendered,
		"Habit 29",
	) {
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
			HabitWithMeta: api.HabitWithMeta{
				Habit: api.Habit{ID: int64(i + 1), Name: fmt.Sprintf("Habit %02d", i)},
			},
		})
	}
	state.Cursors["habits"] = 29
	rendered := ansi.Strip(renderHabits(testTheme(), state, 60, 20))
	if !strings.Contains(rendered, "Habit 29") || !strings.Contains(rendered, "↑ ") {
		t.Fatalf("expected last habit and ↑ marker, got %q", rendered)
	}
}

func TestDailyWideListCursorVisibleAtEnd(t *testing.T) {
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
		state.Pane = "habits" // unfocused, so the card keeps its meta line
		if rendered := ansi.Strip(renderIssues(testTheme(), state, width, 20)); !strings.Contains(
			rendered,
			"crona",
		) ||
			!strings.Contains(rendered, "core") {
			t.Fatalf("width %d: expected the issue's own repo and stream, got %q", width, rendered)
		}
	}
}

func TestDailyScoreCalendarSitsUnderDetailsOnWideScreens(t *testing.T) {
	state := dailyState(200, 40)
	state.DashboardDate = "2026-05-27"
	state.SummarySnapshot = &api.SummarySnapshot{CalendarScores: map[string]api.FocusScoreRangeDay{
		"2026-05-20": {Date: "2026-05-20", Score: 82, HasData: true},
	}}
	rendered := ansi.Strip(renderView(testTheme(), state))
	calendarRow, detailRow := rowOf(t, rendered, "─ Calendar"), rowOf(t, rendered, "─ Day ")
	lines := strings.Split(rendered, "\n")
	if calendarRow <= detailRow || strings.Index(lines[calendarRow], "─ Calendar") < 100 {
		t.Fatalf("expected the calendar in the right column under details, got\n%s", rendered)
	}
	if !strings.Contains(rendered, "May 2026") || !strings.Contains(rendered, "Wed 27 · no score yet") ||
		!strings.Contains(rendered, "■ 75+") {
		t.Fatalf("expected the month with each day's score, got\n%s", rendered)
	}
}

func TestDailyScoreCalendarGoesUnderHabitsOnNarrowScreens(t *testing.T) {
	state := dailyState(130, 40)
	state.DashboardDate = "2026-05-27"
	rendered := ansi.Strip(renderView(testTheme(), state))
	if rowOf(t, rendered, "─ Calendar") <= rowOf(t, rendered, "─ Habits Due") || !strings.Contains(rendered, "May 2026") {
		t.Fatalf("expected the month calendar under Habits, got\n%s", rendered)
	}
	short := dailyState(130, 26)
	short.DashboardDate = "2026-05-27"
	out := ansi.Strip(renderView(testTheme(), short))
	if strings.Contains(out, " 1   2   3   4") || !strings.Contains(out, "25  26  27  28  29  30  31") {
		t.Fatalf("expected only the shown week on a short screen, got\n%s", out)
	}
	tiny := dailyState(130, 18)
	tiny.DashboardDate = "2026-05-27"
	if out := ansi.Strip(renderView(testTheme(), tiny)); !strings.Contains(out, "Calendar ▸") {
		t.Fatalf("expected the calendar folded when even a week does not fit, got\n%s", out)
	}
}

func TestDailyFoldedPanesKeepOnlyTheirTitleRow(t *testing.T) {
	state := dailyState(200, 40)
	state.Settings = &api.CoreSettings{CollapsedPanes: []string{
		uistate.CollapseDailyHabits, uistate.CollapseDailyCalendar, uistate.CollapseDailyDetail,
	}}
	rendered := ansi.Strip(renderView(testTheme(), state))
	lines := strings.Split(rendered, "\n")
	if len(lines) != 40 {
		t.Fatalf("expected the view to keep its height, got %d", len(lines))
	}
	for i, want := range []string{"Habits ▸", "Calendar ▸", "Details ▸"} {
		if row := lines[37+i]; !strings.Contains(row, want) || !strings.Contains(row, "expand") {
			t.Fatalf("expected %q folded on row %d, got\n%s", want, 37+i, rendered)
		}
	}
	if strings.Contains(rendered, "─ Day ") {
		t.Fatalf("expected the detail pane folded away, got\n%s", rendered)
	}
	state.Settings.CollapsedPanes = []string{uistate.CollapseDailyDetail}
	if out := ansi.Strip(renderView(testTheme(), state)); !strings.Contains(out, "Details ▸") ||
		rowOf(t, out, "─ Calendar") <= rowOf(t, out, "Details ▸") {
		t.Fatalf("expected folded details over the calendar in the right column, got\n%s", out)
	}
}

func TestDailyScoreCalendarIsNotFaded(t *testing.T) {
	state := dailyState(130, 40)
	state.DashboardDate = "2026-05-27"
	for _, line := range strings.Split(renderView(testTheme(), state), "\n") {
		if strings.Contains(ansi.Strip(line), "May 2026") && strings.Contains(line, "\x1b[2m") {
			t.Fatalf("expected the calendar body unfaded, got %q", line)
		}
	}
}
