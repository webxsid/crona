package dialogs

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"strings"
	"testing"
	"time"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	controllerpkg "crona/tui/internal/tui/dialogs/controller"
	tea "github.com/charmbracelet/bubbletea"
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

func TestAdhocAttributionChoicesSeparateLabelsAndDescriptions(t *testing.T) {
	state := controllerpkg.OpenAdhocResolution(controllerpkg.State{Width: 100}, nil)
	rendered := renderSessionDialog(Theme{}, state)
	for _, text := range []string{
		"Finish Ad hoc Session",
		"Choose where this session should be recorded.",
		"Attach to an existing issue",
		"Choose an issue, then add a session summary.",
		"Create a new issue",
		"Create an issue and finish this session in one step.",
		"Keep as ad hoc",
		"Create a completed issue in Ad hoc / Sessions.",
		"Convert it",
	} {
		if !strings.Contains(rendered, text) {
			t.Fatalf("attribution dialog is missing %q: %q", text, rendered)
		}
	}
	if strings.Contains(rendered, "Attach to an existing issue  ·") {
		t.Fatalf("option description should be shown below its label, got %q", rendered)
	}
}

func TestAdhocIssuePickerShowsTabsAndExplainsAttachment(t *testing.T) {
	state := controllerpkg.State{
		Kind:                "adhoc_issue_picker",
		Width:               120,
		ViewTitle:           "Attach to Issue",
		ViewName:            "Select an issue to attach this completed session to it.",
		ChoiceItems:         []string{"Fix login flow"},
		ChoiceValues:        []string{"41"},
		ChoiceDetails:       []string{"Atlas / Web · in progress"},
		ChoiceCursor:        0,
		AdhocIssueIDs:       []int64{41},
		AdhocIssueTab:       0,
		AdhocIssueTabCounts: [4]int{1, 2, 3, 5},
	}
	rendered := renderSessionDialog(Theme{}, state)
	for _, want := range []string{
		"Select an issue to attach this completed session to it.",
		"All 5",
		"Due today 1",
		"Overdue 2",
		"Pinned 3",
		"Fix login flow",
		"Atlas / Web · in progress",
		"←/→ tab",
		"enter continue",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("issue picker is missing %q: %q", want, rendered)
		}
	}
	if strings.Contains(rendered, "Fix login flow  ·") {
		t.Fatalf("issue metadata should be separated from the title, got %q", rendered)
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

func TestAdhocPickerKeepsSelectionAndFooterVisible(t *testing.T) {
	for _, size := range [][2]int{{80, 24}, {120, 40}, {180, 60}} {
		state := controllerpkg.State{Kind: "adhoc_issue_picker", Width: size[0], Height: size[1], ViewTitle: "Attach to Issue", ViewName: "Choose an issue. You’ll add a session summary before finishing.", ChoiceCursor: 39}
		for i := 0; i < 60; i++ {
			state.ChoiceItems = append(state.ChoiceItems, fmt.Sprintf("Issue %02d %s", i, strings.Repeat("long title ", 20)))
			state.ChoiceDetails = append(state.ChoiceDetails, "Work / App · planned · 2026-10-07")
			state.AdhocIssueIDs = append(state.AdhocIssueIDs, int64(i+1))
		}
		rendered := renderSessionDialog(Theme{}, state)
		for _, want := range []string{"Issue 39", "40 of 60", "enter continue", "↑ more", "↓ more"} {
			if !strings.Contains(rendered, want) {
				t.Fatalf("size %v missing %q: %s", size, want, rendered)
			}
		}
		if h := lipgloss.Height(rendered); h > size[1] {
			t.Fatalf("size %v: dialog height %d", size, h)
		}
		if w := lipgloss.Width(rendered); w > size[0] {
			t.Fatalf("size %v: dialog width %d", size, w)
		}
	}
}
func TestAdhocConfirmationExplainsDestination(t *testing.T) {
	state := controllerpkg.OpenAdhocEndSession(controllerpkg.State{Width: 100, AdhocElapsedSeconds: 91}, sharedtypes.AdhocSessionResolutionAttach, 41, 2, 3, "")
	state.ViewName = "Fix login"
	state.RepoName = "Work"
	state.StreamName = "App"
	rendered := renderSessionDialog(Theme{}, state)
	for _, want := range []string{"Fix login", "Work / App", "Elapsed 1m 31s", "Session summary (required)", "attach & finish"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("missing %q", want)
		}
	}
}

func TestShortConversionFormKeepsActiveFieldAndErrorVisible(t *testing.T) {
	state := controllerpkg.OpenPromoteAdhocIssue(controllerpkg.State{Width: 80, Height: 24}, api.IssueWithMeta{Issue: sharedtypes.Issue{ID: 1, Title: "Session title", Adhoc: true}}, nil)
	for i := 0; i < 6; i++ {
		state.ErrorMessage = "Choose a repo and stream"
		rendered := renderIssueDialog(Theme{}, state)
		for _, want := range []string{"Convert Ad hoc Issue", "ctrl+s convert issue", "Choose a repo and stream"} {
			if !strings.Contains(rendered, want) {
				t.Fatalf("focus %d missing %q: %s", state.FocusIdx, want, rendered)
			}
		}
		if h := lipgloss.Height(rendered); h > 24 {
			t.Fatalf("focus %d height %d", state.FocusIdx, h)
		}
		for _, input := range state.Inputs {
			text := input.Value()
			if text == "" {
				continue // placeholders are restyled; TestNewIssueShortTerminal covers empty fields
			}
			needle := []rune(strings.TrimSpace(text))
			if len(needle) > 12 {
				needle = needle[:12]
			}
			if input.Focused() && !strings.Contains(ansi.Strip(rendered), string(needle)) {
				t.Fatalf("focused input missing: %s", rendered)
			}
		}
		state, _, _ = controllerpkg.Update(state, controllerpkg.UpdateContext{}, "2026-10-08", tea.KeyMsg{Type: tea.KeyTab})
	}
}
