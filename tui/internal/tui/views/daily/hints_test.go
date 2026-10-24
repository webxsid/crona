package daily

import (
	"strings"
	"testing"

	"crona/tui/internal/api"

	"github.com/charmbracelet/x/ansi"
)

func TestTasksPaneOffersNewTaskLine(t *testing.T) {
	rendered := ansi.Strip(renderIssues(testTheme(), detailState(), 80, 20))
	if rowOf(t, rendered, "[a] new task  [A] ad hoc") != rowOf(t, rendered, "Planned")+1 {
		t.Fatalf(
			"expected [a] new task and [A] ad hoc on their own line under the selector, got %q",
			rendered,
		)
	}
	state := detailState()
	state.Pane = "issues"
	if focused := ansi.Strip(renderIssues(testTheme(), state, 200, 20)); strings.Count(
		focused,
		"[A] ad hoc",
	) != 1 {
		t.Fatalf("expected [A] ad hoc only on the new-task line, not the card, got %q", focused)
	}
}

func TestSelectedTaskCardShowsActions(t *testing.T) {
	state := detailState()
	state.Pane = "issues"
	focused := ansi.Strip(renderIssues(testTheme(), state, 120, 20))
	row := strings.Split(focused, "\n")[rowOf(t, focused, "Fix sync conflict")+1]
	for _, want := range []string{"[s] status", "[d] due", "[e] edit", "[D] delete"} {
		if !strings.Contains(row, want) {
			t.Fatalf("expected the selected card's second line to offer %q, got %q", want, row)
		}
	}
	if strings.Contains(row, "in progress") {
		t.Fatalf("expected the actions to replace the meta line, got %q", row)
	}
	if !strings.Contains(focused, "ready") {
		t.Fatalf("expected other cards to keep their meta, got %q", focused)
	}
	state.Pane = "habits"
	unfocused := ansi.Strip(renderIssues(testTheme(), state, 120, 20))
	if strings.Contains(unfocused, "[s] status") || !strings.Contains(unfocused, "in progress") {
		t.Fatalf("expected the meta line back when the pane loses focus, got %q", unfocused)
	}
}

func TestTaskActionsHideTimerKeysWhileRunning(t *testing.T) {
	state := detailState()
	state.Pane = "issues"
	state.Timer = &api.TimerState{State: "running"}
	rendered := ansi.Strip(renderIssues(testTheme(), state, 160, 20))
	if !strings.Contains(rendered, "[s] status") {
		t.Fatalf("expected status kept while the timer runs, got %q", rendered)
	}
	for _, hidden := range []string{"[f] focus", "[e] edit", "[m] log", "[A] ad hoc"} {
		if strings.Contains(rendered, hidden) {
			t.Fatalf("expected %q hidden while the timer runs, got %q", hidden, rendered)
		}
	}
}

func TestAdhocTaskActionsOfferConvertOnly(t *testing.T) {
	state := detailState()
	state.Pane = "issues"
	state.SelectedIssueAdhoc = true
	rendered := ansi.Strip(renderIssues(testTheme(), state, 160, 20))
	if !strings.Contains(rendered, "[e] convert") {
		t.Fatalf("expected an ad hoc task to offer convert, got %q", rendered)
	}
	for _, hidden := range []string{"[s] status", "[d] due", "[D] delete", "[f] focus"} {
		if strings.Contains(rendered, hidden) {
			t.Fatalf("expected %q hidden on an ad hoc task, got %q", hidden, rendered)
		}
	}
}

func TestHabitsPaneOffersNewHabitAndActionLine(t *testing.T) {
	state := meterHabitState(130)
	unfocused := ansi.Strip(renderHabits(testTheme(), state, 130, 10))
	if !strings.Contains(unfocused, "[a] new habit") || strings.Contains(unfocused, "[x] done") {
		t.Fatalf("expected the new-habit line and no action line when unfocused, got %q", unfocused)
	}
	state.Pane = "habits"
	focused := ansi.Strip(renderHabits(testTheme(), state, 130, 10))
	row := strings.Split(focused, "\n")[rowOf(t, focused, "▶")+1]
	for _, want := range []string{"[x] done", "[F] fail", "[m] log"} {
		if !strings.Contains(row, want) {
			t.Fatalf("expected the line under the selected habit to offer %q, got %q", want, row)
		}
	}
}

func TestDetailTitleOffersDayKey(t *testing.T) {
	state := detailState()
	if rendered := ansi.Strip(renderDetail(testTheme(), state, 70, 30)); !strings.Contains(
		strings.Split(rendered, "\n")[0],
		"[S] day summary",
	) {
		t.Fatalf("expected [S] day summary in the detail title row, got %q", rendered)
	}
	state.DailyDetailDay = true
	if rendered := ansi.Strip(renderDetail(testTheme(), state, 70, 30)); !strings.Contains(
		strings.Split(rendered, "\n")[0],
		"[S] show selection",
	) {
		t.Fatalf("expected [S] show selection while the day shows, got %q", rendered)
	}
}

func TestTaskActionsFitShortPaneWithPins(t *testing.T) {
	state := groupedPlannedState(60)
	for i := range 8 {
		state.DailyIssues = append(
			state.DailyIssues,
			api.Issue{ID: int64(100 + i), Title: "Later", Status: "planned"},
		)
	}
	state.Pane = "issues"
	state.Cursors["issues"] = 4
	rendered := ansi.Strip(renderIssues(testTheme(), state, 60, 7))
	if lines := strings.Split(rendered, "\n"); len(lines) != 7 ||
		!strings.Contains(rendered, "[s] status") {
		t.Fatalf("expected the selected card's actions within 7 rows, got %q", rendered)
	}
}

func TestRunningTimerHidesRefusedKeys(t *testing.T) {
	state := detailState()
	state.Pane = "issues"
	state.SelectedIssueAdhoc = true
	state.Timer = &api.TimerState{State: "running"}
	if rendered := ansi.Strip(renderIssues(testTheme(), state, 160, 20)); strings.Contains(
		rendered,
		"[e] convert",
	) {
		t.Fatalf("expected convert hidden while a timer runs, got %q", rendered)
	}
	habits := meterHabitState(130)
	habits.Pane = "habits"
	habits.Timer = &api.TimerState{State: "running"}
	if rendered := ansi.Strip(renderHabits(testTheme(), habits, 130, 10)); strings.Contains(
		rendered,
		"[D] delete",
	) {
		t.Fatalf("expected habit delete hidden while a timer runs, got %q", rendered)
	}
}

func TestTaskActionLineKeepsWholeHints(t *testing.T) {
	labels := []string{"open", "status", "due", "pin", "tags", "focus", "edit", "delete", "log"}
	for width := 40; width <= 110; width++ {
		state := detailState()
		state.Pane = "issues"
		rendered := ansi.Strip(renderIssues(testTheme(), state, width, 20))
		row := strings.TrimSpace(
			strings.Split(rendered, "\n")[rowOf(t, rendered, "Fix sync conflict")+1],
		)
		whole := false
		for _, label := range labels {
			whole = whole || strings.HasSuffix(row, " "+label)
		}
		if !whole {
			t.Fatalf(
				"width %d: expected the action line to end on a whole hint, got %q",
				width,
				row,
			)
		}
	}
}

func TestDetailTitleTruncatesToKeepDayKey(t *testing.T) {
	state := detailState()
	state.DailyIssues[1].Title = strings.Repeat("Very long task title ", 6)
	for _, width := range []int{40, 70} {
		header := strings.Split(ansi.Strip(renderDetail(testTheme(), state, width, 30)), "\n")[0]
		if !strings.Contains(header, "Very long") || !strings.Contains(header, "…") ||
			!strings.HasSuffix(header, "[S] day summary") {
			t.Fatalf("width %d: expected a truncated title and the day key, got %q", width, header)
		}
		if got := ansi.StringWidth(header); got != width {
			t.Fatalf("width %d: header is %d cells: %q", width, got, header)
		}
	}
}
