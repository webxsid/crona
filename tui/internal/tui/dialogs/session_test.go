package dialogs

import (
	"strings"
	"testing"

	sharedtypes "crona/shared/types"
	controllerpkg "crona/tui/internal/tui/dialogs/controller"
)

func TestMomentumDialogUsesMomentumTitle(t *testing.T) {
	state := controllerpkg.OpenCreateMomentumDirect(controllerpkg.State{}, nil, nil)
	rendered := renderHabitStreakDialog(Theme{}, state)
	if !strings.Contains(rendered, "Momentum") {
		t.Fatalf("expected momentum dialog title, got %q", rendered)
	}
	if !strings.Contains(rendered, "Description") {
		t.Fatalf("expected momentum dialog to include description field, got %q", rendered)
	}
	if strings.Contains(rendered, "Habit Streaks") {
		t.Fatalf("expected momentum dialog to avoid the settings title, got %q", rendered)
	}
}

func TestEditMomentumDialogShowsDescription(t *testing.T) {
	desc := "Keep the mix steady."
	state := controllerpkg.OpenEditMomentumDirect(
		controllerpkg.State{},
		nil,
		nil,
		sharedtypes.HabitStreakDefinition{
			ID:            "momentum-1",
			Name:          "Recovery Mix",
			Description:   &desc,
			Enabled:       true,
			Period:        sharedtypes.HabitStreakPeriodWeek,
			RequiredCount: 2,
		},
	)
	rendered := renderHabitStreakDialog(Theme{}, state)
	if !strings.Contains(rendered, desc) {
		t.Fatalf("expected momentum edit dialog to show description, got %q", rendered)
	}
}

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
