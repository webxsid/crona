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
	if !strings.Contains(rendered, "> Number of cycles (1 cycle = 1 focus + break)") {
		t.Fatalf("expected active row marker for cycles row, got %q", rendered)
	}
	if strings.Contains(rendered, "> Focus Time") {
		t.Fatalf("expected only the active row to be highlighted, got %q", rendered)
	}
}

func TestPomodoroStartHighlightsCustomEditingRow(t *testing.T) {
	state := controllerpkg.OpenPomodoroStart(controllerpkg.State{}, 11, 22, 33, "Issue title")
	state.PomodoroFocusChoice = 3
	state.FocusIdx = 1

	rendered := renderSessionDialog(Theme{}, state)
	if !strings.Contains(rendered, "> Focus Time") {
		t.Fatalf("expected focus row marker while editing custom input, got %q", rendered)
	}
}
