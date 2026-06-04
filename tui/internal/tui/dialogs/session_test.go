package dialogs

import (
	"strings"
	"testing"

	controllerpkg "crona/tui/internal/tui/dialogs/controller"
)

func TestPomodoroStartHighlightsActiveRow(t *testing.T) {
	state := controllerpkg.OpenPomodoroStart(controllerpkg.State{}, 11, 22, 33, "Issue title")
	state.FocusIdx = 6

	rendered := renderSessionDialog(Theme{}, state)
	if !strings.Contains(rendered, "> Cycles") {
		t.Fatalf("expected active row marker for cycles row, got %q", rendered)
	}
	if strings.Contains(rendered, "> Focus") {
		t.Fatalf("expected only the active row to be highlighted, got %q", rendered)
	}
	if !strings.Contains(rendered, "25m Focus  ·  5m Short Break  ·  15m Long Break") {
		t.Fatalf("expected unified summary line to remain visible, got %q", rendered)
	}
}

func TestPomodoroStartHighlightsCustomEditingRow(t *testing.T) {
	state := controllerpkg.OpenPomodoroStart(controllerpkg.State{}, 11, 22, 33, "Issue title")
	state.PomodoroFocusChoice = 3
	state.FocusIdx = 1

	rendered := renderSessionDialog(Theme{}, state)
	if !strings.Contains(rendered, "> Focus") {
		t.Fatalf("expected focus row marker while editing custom input, got %q", rendered)
	}
}

func TestPomodoroStartShowsLongBreakForcedOffWhenShortBreakDisabled(t *testing.T) {
	state := controllerpkg.OpenPomodoroStart(controllerpkg.State{}, 11, 22, 33, "Issue title")
	state.PomodoroBreakChoice = 3
	state.PomodoroBreakSeconds = 0
	state.PomodoroLongBreakChoice = 0

	rendered := renderSessionDialog(Theme{}, state)
	if !strings.Contains(rendered, "Long Break: disabled") {
		t.Fatalf("expected compact long-break disabled text, got %q", rendered)
	}
	if !strings.Contains(rendered, "Cycles: disabled") {
		t.Fatalf("expected compact cycles disabled text, got %q", rendered)
	}
	if !strings.Contains(rendered, "Long Break: disabled") {
		t.Fatalf("expected compact long-break cycle disabled text, got %q", rendered)
	}
}
