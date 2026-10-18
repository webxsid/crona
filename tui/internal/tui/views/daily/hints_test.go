package daily

import (
	"strings"
	"testing"

	"crona/tui/internal/api"

	"github.com/charmbracelet/x/ansi"
)

func TestTasksPaneOffersNewTaskLine(t *testing.T) {
	rendered := ansi.Strip(renderIssues(testTheme(), detailState(), 80, 20))
	if rowOf(t, rendered, "[a] new task") != rowOf(t, rendered, "Planned")+1 {
		t.Fatalf("expected [a] new task on its own line under the selector, got %q", rendered)
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
	for _, hidden := range []string{"[f] focus", "[e] edit", "[m] log"} {
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
	if rendered := ansi.Strip(renderDetail(testTheme(), state, 70, 30)); !strings.Contains(strings.Split(rendered, "\n")[0], "[S] day") {
		t.Fatalf("expected [S] day in the detail title row, got %q", rendered)
	}
	state.DailyDetailDay = true
	if rendered := ansi.Strip(renderDetail(testTheme(), state, 70, 30)); !strings.Contains(strings.Split(rendered, "\n")[0], "[S] selection") {
		t.Fatalf("expected [S] selection while the day shows, got %q", rendered)
	}
}
