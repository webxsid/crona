package dialogs

import (
	"strings"
	"testing"
	"time"

	sharedtypes "crona/shared/types"
	controllerpkg "crona/tui/internal/tui/dialogs/controller"
)

func TestMomentumDialogUsesMomentumTitle(t *testing.T) {
	state := controllerpkg.OpenCreateMomentumDirect(controllerpkg.State{}, nil, nil, nil, nil, nil)
	rendered := renderHabitStreakDialog(Theme{}, state)
	if !strings.Contains(rendered, "Momentum") {
		t.Fatalf("expected momentum dialog title, got %q", rendered)
	}
	if !strings.Contains(rendered, "Choose momentum kind") {
		t.Fatalf("expected create momentum dialog to start with kind selection, got %q", rendered)
	}
	if strings.Contains(rendered, "Habit Streaks") {
		t.Fatalf("expected momentum dialog to avoid the settings title, got %q", rendered)
	}
}

func TestEditMomentumDialogStartsWithTargetsSelection(t *testing.T) {
	state := controllerpkg.OpenEditMomentumDirect(
		controllerpkg.State{},
		nil,
		nil,
		nil,
		nil,
		nil,
		sharedtypes.HabitStreakDefinition{
			ID:            "momentum-1",
			Name:          "Recovery Mix",
			Enabled:       true,
			Period:        sharedtypes.HabitStreakPeriodWeek,
			RequiredCount: 2,
		},
	)
	rendered := renderHabitStreakDialog(Theme{}, state)
	if !strings.Contains(rendered, "Select contributing habits") {
		t.Fatalf("expected edit momentum dialog to start with target selection, got %q", rendered)
	}
	if strings.Contains(rendered, "Choose momentum kind") {
		t.Fatalf("expected edit momentum dialog to skip kind selection, got %q", rendered)
	}
}

func TestEditMomentumDialogShowsDescriptionOnDetailsStep(t *testing.T) {
	desc := "Keep the mix steady."
	state := controllerpkg.OpenEditMomentumDirect(
		controllerpkg.State{},
		nil,
		nil,
		nil,
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
	state.HabitStreakStep = 2
	rendered := renderHabitStreakDialog(Theme{}, state)
	if !strings.Contains(rendered, desc) {
		t.Fatalf("expected momentum edit dialog to show description, got %q", rendered)
	}
}

func TestPomodoroStartHighlightsActiveRow(t *testing.T) {
	estimate := 60
	state := controllerpkg.OpenPomodoroStart(
		controllerpkg.State{},
		11,
		22,
		33,
		"Issue title",
		&estimate,
		900,
	)
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
	for _, want := range []string{"Worked / Est", "worked 15m / est. 1h", "Total", "2h10m", "Ends At"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("expected pomodoro dialog to contain %q, got %q", want, rendered)
		}
	}
}

func TestPomodoroStartHighlightsCustomEditingRow(t *testing.T) {
	state := controllerpkg.OpenPomodoroStart(
		controllerpkg.State{},
		11,
		22,
		33,
		"Issue title",
		nil,
		0,
	)
	state.PomodoroFocusChoice = 3
	state.FocusIdx = 1

	rendered := renderSessionDialog(Theme{}, state)
	if !strings.Contains(rendered, "> Focus") {
		t.Fatalf("expected focus row marker while editing custom input, got %q", rendered)
	}
}

func TestPomodoroStartShowsLongBreakForcedOffWhenShortBreakDisabled(t *testing.T) {
	state := controllerpkg.OpenPomodoroStart(
		controllerpkg.State{},
		11,
		22,
		33,
		"Issue title",
		nil,
		0,
	)
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

func TestTimerStartTypeShowsEstimateSummaryAndTimerChoice(t *testing.T) {
	estimate := 45
	state := controllerpkg.OpenTimerStartType(
		controllerpkg.State{},
		11,
		22,
		33,
		"Issue title",
		&estimate,
		1200,
	)

	rendered := renderSessionDialog(Theme{}, state)
	for _, want := range []string{"Worked / Est", "worked 20m / est. 45m", "[t] Timer"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("expected timer start type to contain %q, got %q", want, rendered)
		}
	}
	if strings.Contains(rendered, "Ends At") {
		t.Fatalf("expected timer start type to omit ends-at preview, got %q", rendered)
	}
}

func TestTimerCountdownDialogShowsSingleCountdownCopy(t *testing.T) {
	estimate := 45
	state := controllerpkg.OpenSingleTimerStart(
		controllerpkg.State{},
		11,
		22,
		33,
		"Issue title",
		&estimate,
		0,
	)

	rendered := renderSessionDialog(Theme{}, state)
	for _, want := range []string{"Timer Session", "Focus Time", "Single countdown. No breaks.", "Worked / Est", "worked - / est. 45m", "Ends At"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("expected countdown dialog to contain %q, got %q", want, rendered)
		}
	}
}

func TestCountdownCompletionAndExtensionUseTimerCopy(t *testing.T) {
	expired := controllerpkg.OpenHardLimitExpired(controllerpkg.State{
		HardLimitKind: sharedtypes.TimerHardLimitKindCountdown,
	}, "Issue title")
	rendered := renderSessionDialog(Theme{}, expired)
	for _, want := range []string{"Timer Session Complete", "finish this timer session"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("expected countdown completion to contain %q, got %q", want, rendered)
		}
	}
	if strings.Contains(rendered, "pomodoro") || strings.Contains(rendered, "Pomodoro") {
		t.Fatalf("expected countdown completion to avoid Pomodoro copy, got %q", rendered)
	}

	extend := controllerpkg.OpenHardLimitExtend(controllerpkg.State{
		HardLimitKind:         sharedtypes.TimerHardLimitKindCountdown,
		HardLimitFocusSeconds: 45 * 60,
		ViewName:              "Issue title",
	})
	rendered = renderSessionDialog(Theme{}, extend)
	for _, want := range []string{"Extend Timer Session", "Additional Time", "No breaks or cycles"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("expected countdown extension to contain %q, got %q", want, rendered)
		}
	}
}

func TestTimerEndsAtLabelUsesLocalHHMM(t *testing.T) {
	base := time.Date(2026, 7, 7, 18, 5, 0, 0, time.Local)
	if got := timerEndsAtLabel(base, 90*60); got != "19:35" {
		t.Fatalf("expected local HH:MM ends-at label, got %q", got)
	}
}

func TestTimerEndsAtLabelSkipsNonPositiveDurations(t *testing.T) {
	if got := timerEndsAtLabel(time.Date(2026, 7, 7, 18, 5, 0, 0, time.Local), 0); got != "" {
		t.Fatalf("expected empty ends-at label for zero duration, got %q", got)
	}
}
