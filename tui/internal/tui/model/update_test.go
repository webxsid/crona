package model

import (
	"testing"
	"time"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	commands "crona/tui/internal/tui/commands"
	dialogstate "crona/tui/internal/tui/dialogs/controller"
	dispatchpkg "crona/tui/internal/tui/dispatch"
	uistate "crona/tui/internal/tui/state"
	wellbeingview "crona/tui/internal/tui/views/wellbeing"
	tea "github.com/charmbracelet/bubbletea"
)

func TestDispatchMessageStatePreservesStashConflictDialogPayload(t *testing.T) {
	model := Model{width: 100}
	model = model.openStashConflictDialog(sharedtypes.StashConflict{
		IssueID: 42,
		Stashes: []sharedtypes.Stash{{
			ID:        "stash-1",
			CreatedAt: "2026-04-10T09:00:00Z",
		}},
	}, 7, 8, 42, "")

	roundTripped := model.applyDispatchMessageState(model.dispatchMessageState())
	if roundTripped.dialog != "stash_conflict" {
		t.Fatalf("expected stash conflict dialog, got %q", roundTripped.dialog)
	}
	if roundTripped.dialogDeleteID != "stash-1" {
		t.Fatalf(
			"expected stash id to survive dispatch bridge, got %q",
			roundTripped.dialogDeleteID,
		)
	}
	if roundTripped.dialogRepoID != 7 || roundTripped.dialogStreamID != 8 ||
		roundTripped.dialogIssueID != 42 {
		t.Fatalf(
			"expected issue path to survive dispatch bridge, got repo=%d stream=%d issue=%d",
			roundTripped.dialogRepoID,
			roundTripped.dialogStreamID,
			roundTripped.dialogIssueID,
		)
	}
	if len(roundTripped.dialogChoiceValues) != 3 ||
		roundTripped.dialogChoiceValues[0] != "resume" ||
		roundTripped.dialogChoiceValues[1] != "continue" ||
		roundTripped.dialogChoiceValues[2] != "commit" {
		t.Fatalf(
			"expected choice values to survive dispatch bridge, got %#v",
			roundTripped.dialogChoiceValues,
		)
	}
}

func TestIssuePanePreflightMessagesOpenStashConflictForFocusAndManual(t *testing.T) {
	tests := []struct {
		name         string
		view         View
		allIssues    []api.IssueWithMeta
		dailySummary *api.DailyIssueSummary
		issues       []api.Issue
		context      *api.ActiveContext
	}{
		{
			name: "daily",
			view: ViewDaily,
			allIssues: []api.IssueWithMeta{{
				Issue:      sharedtypes.Issue{ID: 42, StreamID: 8, TodoForDate: stringPtr("2026-04-10")},
				RepoID:     7,
				RepoName:   "Work",
				StreamName: "app",
			}},
			dailySummary: &api.DailyIssueSummary{
				Issues: []sharedtypes.Issue{
					{ID: 42, StreamID: 8, Title: "Daily issue", TodoForDate: stringPtr("2026-04-10")},
				},
			},
		},
		{
			name: "default",
			view: ViewDefault,
			allIssues: []api.IssueWithMeta{{
				Issue:      sharedtypes.Issue{ID: 42, StreamID: 8, Title: "Default issue"},
				RepoID:     7,
				RepoName:   "Work",
				StreamName: "app",
			}},
		},
		{
			name: "meta",
			view: ViewMeta,
			allIssues: []api.IssueWithMeta{{
				Issue:      sharedtypes.Issue{ID: 42, StreamID: 8, Title: "Meta issue"},
				RepoID:     7,
				RepoName:   "Work",
				StreamName: "app",
			}},
			issues: []api.Issue{{ID: 42, StreamID: 8, Title: "Meta issue"}},
			context: &api.ActiveContext{
				RepoID:     int64Ptr(7),
				RepoName:   stringPtr("Work"),
				StreamID:   int64Ptr(8),
				StreamName: stringPtr("app"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := Model{
				view: tt.view,
				pane: PaneIssues,
				cursor: map[Pane]int{
					PaneIssues: 0,
				},
				allIssues:     tt.allIssues,
				dailySummary:  tt.dailySummary,
				issues:        tt.issues,
				context:       tt.context,
				dashboardDate: "2026-04-10",
			}

			nextModel, cmd := model.handleInputStartFocusFromSelection()
			next := nextModel.(Model)
			if cmd == nil {
				t.Fatal("expected focus preflight command")
			}
			if next.dialog != "" {
				t.Fatalf("expected focus preflight to defer dialog open, got %q", next.dialog)
			}
			nextAny, _ := next.update(commands.IssueActionPreflightConflictMsg{
				Mode: commands.IssueActionModeFocus,
				Target: commands.IssueActionTarget{
					RepoID:   7,
					StreamID: 8,
					IssueID:  42,
				},
				Conflict: sharedtypes.StashConflict{
					IssueID: 42,
					Stashes: []sharedtypes.Stash{{
						ID:        "stash-1",
						CreatedAt: "2026-04-10T09:00:00Z",
					}},
				},
			})
			next = nextAny.(Model)
			if next.dialog != "stash_conflict" {
				t.Fatalf("expected stash conflict dialog for focus, got %q", next.dialog)
			}
			if next.dialogParent != "" {
				t.Fatalf("expected focus stash conflict parent to be empty, got %q", next.dialogParent)
			}
			if len(next.dialogChoiceValues) != 3 || next.dialogChoiceValues[2] != "commit" {
				t.Fatalf("expected focus stash conflict actions, got %+v", next.dialogChoiceValues)
			}

			inputState := model.inputState()
			cmd, handled := model.inputDeps().OpenManualSessionDialog(&inputState)
			if !handled {
				t.Fatal("expected manual preflight to be handled")
			}
			if cmd == nil {
				t.Fatal("expected manual preflight command")
			}
			next = model.applyInputState(inputState)
			if next.dialog != "" {
				t.Fatalf("expected manual preflight to defer dialog open, got %q", next.dialog)
			}
			nextAny, _ = next.update(commands.IssueActionPreflightConflictMsg{
				Mode: commands.IssueActionModeManual,
				Target: commands.IssueActionTarget{
					RepoID:          7,
					StreamID:        8,
					IssueID:         42,
					EstimateMinutes: nil,
				},
				Conflict: sharedtypes.StashConflict{
					IssueID: 42,
					Stashes: []sharedtypes.Stash{{
						ID:        "stash-1",
						CreatedAt: "2026-04-10T09:00:00Z",
					}},
				},
			})
			next = nextAny.(Model)
			if next.dialog != "stash_conflict" {
				t.Fatalf("expected stash conflict dialog for manual, got %q", next.dialog)
			}
			if next.dialogParent != "manual_session" {
				t.Fatalf("expected manual stash conflict parent, got %q", next.dialogParent)
			}
			if len(next.dialogChoiceValues) != 3 ||
				next.dialogChoiceValues[1] != "manual" ||
				next.dialogChoiceValues[2] != "commit" {
				t.Fatalf("expected manual stash conflict actions, got %+v", next.dialogChoiceValues)
			}
		})
	}
}

func TestIssuePanePreflightClearMessagesOpenRequestedDialog(t *testing.T) {
	model := Model{
		view: ViewDefault,
		pane: PaneIssues,
		cursor: map[Pane]int{
			PaneIssues: 0,
		},
		allIssues: []api.IssueWithMeta{{
			Issue:      sharedtypes.Issue{ID: 42, StreamID: 8, Title: "Default issue"},
			RepoID:     7,
			RepoName:   "Work",
			StreamName: "app",
		}},
		dashboardDate: "2026-04-10",
	}

	nextAny, _ := model.update(commands.IssueActionPreflightClearMsg{
		Mode: commands.IssueActionModeFocus,
		Target: commands.IssueActionTarget{
			RepoID:   7,
			StreamID: 8,
			IssueID:  42,
			Title:    "Default issue",
		},
	})
	next := nextAny.(Model)
	if next.dialog != "timer_start_type" {
		t.Fatalf("expected timer selector for clear focus path, got %q", next.dialog)
	}

	nextAny, _ = model.update(commands.IssueActionPreflightClearMsg{
		Mode: commands.IssueActionModeManual,
		Target: commands.IssueActionTarget{
			RepoID:   7,
			StreamID: 8,
			IssueID:  42,
			Title:    "Default issue",
		},
	})
	next = nextAny.(Model)
	if next.dialog != "manual_session" {
		t.Fatalf("expected manual dialog for clear manual path, got %q", next.dialog)
	}
}

func int64Ptr(v int64) *int64 {
	return &v
}

func stringPtr(v string) *string {
	return &v
}

func TestTimerActivityTouchCmdOnlyForActiveTimerAndThrottles(t *testing.T) {
	now := time.Date(2026, 4, 11, 10, 0, 0, 0, time.UTC)
	model := Model{
		client: api.NewClient("unix", "/tmp/missing.sock"),
		timer:  &api.TimerState{State: "idle"},
	}
	if cmd := model.timerActivityTouchCmd(now); cmd != nil {
		t.Fatalf("expected no touch command while idle")
	}

	model.timer = &api.TimerState{State: "running"}
	if cmd := model.timerActivityTouchCmd(now); cmd == nil {
		t.Fatalf("expected touch command for active timer")
	}
	if cmd := model.timerActivityTouchCmd(now.Add(30 * time.Second)); cmd != nil {
		t.Fatalf("expected touch command to be throttled")
	}
	if cmd := model.timerActivityTouchCmd(now.Add(61 * time.Second)); cmd == nil {
		t.Fatalf("expected touch command after throttle window")
	}
}

func TestAnchorWellbeingScrollUsesCurrentPaneHeight(t *testing.T) {
	model := Model{width: 100, height: 40}
	state := dispatchpkg.MessageState{
		Width:               100,
		Height:              40,
		View:                uistate.ViewWellbeing,
		Pane:                uistate.PaneWellbeingTrends,
		WellbeingWindowDays: 14,
		Cursor: map[uistate.Pane]int{
			uistate.PaneWellbeingSummary: 0,
			uistate.PaneWellbeingTrends:  0,
			"wellbeing_streaks":          0,
		},
		MetricsRange: makeWellbeingRangeForAnchorTest(14),
		MetricsRollup: &api.MetricsRollup{
			Days:          14,
			CheckInDays:   10,
			FocusDays:     8,
			WorkedSeconds: 14400,
			RestSeconds:   2400,
		},
		Streaks: &api.StreakSummary{
			CurrentCheckInDays: 3,
			LongestCheckInDays: 7,
			CurrentFocusDays:   2,
			LongestFocusDays:   5,
		},
	}
	deps := model.dispatchMessageDeps()
	if deps.AnchorWellbeingScroll == nil {
		t.Fatal("expected anchor wellbeing scroll hook to be wired")
	}
	next := model.applyDispatchMessageState(state)
	expected := wellbeingview.PaneLineCount(
		next.viewContentState(
			next.mainContentWidth(),
			next.contentHeight(),
			next.selectionSnapshot(),
			nil,
		),
		string(uistate.PaneWellbeingTrends),
	) - 1
	deps.AnchorWellbeingScroll(&state, uistate.PaneWellbeingTrends)
	if state.Cursor[uistate.PaneWellbeingTrends] <= 0 {
		t.Fatalf(
			"expected trends cursor to anchor near bottom, got %d",
			state.Cursor[uistate.PaneWellbeingTrends],
		)
	}
	if state.Cursor[uistate.PaneWellbeingTrends] != expected {
		t.Fatalf(
			"expected cursor to anchor to bottom of rendered pane, got %d",
			state.Cursor[uistate.PaneWellbeingTrends],
		)
	}
}

func TestDialogModeTreatsQAsCancelAndCtrlCAsQuit(t *testing.T) {
	model := Model{}.withDialogState(dialogstate.State{Kind: "onboarding"})

	next, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd != nil {
		t.Fatalf("expected q to cancel the dialog, not quit")
	}
	if next.(Model).dialog != "" {
		t.Fatalf("expected q to close the dialog, got %q", next.(Model).dialog)
	}

	next, cmd = model.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Fatalf("expected quit command for ctrl+c")
	}
	if next.(Model).dialog != "onboarding" {
		t.Fatalf("expected dialog state to remain unchanged, got %q", next.(Model).dialog)
	}
}

func TestHardLimitExpiredDialogIgnoresQ(t *testing.T) {
	model := Model{}.withDialogState(dialogstate.State{Kind: "hard_limit_expired"})

	next, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd != nil {
		t.Fatalf("expected q to be ignored for hard-limit expired dialog")
	}
	if next.(Model).dialog != "hard_limit_expired" {
		t.Fatalf("expected hard-limit expired dialog to stay open, got %q", next.(Model).dialog)
	}
}

func TestDialogStatePreservesPomodoroFields(t *testing.T) {
	model := Model{}.withDialogState(dialogstate.State{
		Kind:                          "pomodoro_start",
		PomodoroFocusSeconds:          60,
		PomodoroFocusChoice:           2,
		PomodoroBreakSeconds:          30,
		PomodoroBreakChoice:           4,
		PomodoroLongBreakSeconds:      120,
		PomodoroLongBreakChoice:       3,
		PomodoroCyclesBeforeLongBreak: 2,
		PomodoroCycles:                5,
	})

	state := model.dialogState()
	if state.PomodoroFocusSeconds != 60 {
		t.Fatalf("expected focus seconds to survive dialog round trip, got %d", state.PomodoroFocusSeconds)
	}
	if state.PomodoroFocusChoice != 2 {
		t.Fatalf("expected focus choice to survive dialog round trip, got %d", state.PomodoroFocusChoice)
	}
	if state.PomodoroBreakSeconds != 30 {
		t.Fatalf("expected break seconds to survive dialog round trip, got %d", state.PomodoroBreakSeconds)
	}
	if state.PomodoroBreakChoice != 4 {
		t.Fatalf("expected break choice to survive dialog round trip, got %d", state.PomodoroBreakChoice)
	}
	if state.PomodoroLongBreakSeconds != 120 {
		t.Fatalf(
			"expected long break seconds to survive dialog round trip, got %d",
			state.PomodoroLongBreakSeconds,
		)
	}
	if state.PomodoroLongBreakChoice != 3 {
		t.Fatalf(
			"expected long break choice to survive dialog round trip, got %d",
			state.PomodoroLongBreakChoice,
		)
	}
	if state.PomodoroCyclesBeforeLongBreak != 2 {
		t.Fatalf(
			"expected cycle count to survive dialog round trip, got %d",
			state.PomodoroCyclesBeforeLongBreak,
		)
	}
	if state.PomodoroCycles != 5 {
		t.Fatalf("expected cycles to survive dialog round trip, got %d", state.PomodoroCycles)
	}
}

func TestPomodoroDialogRightKeyPersistsAcrossModelUpdates(t *testing.T) {
	model := Model{}.withDialogState(dialogstate.OpenPomodoroStart(dialogstate.State{}, 11, 22, 33, "Issue title"))

	for i := 0; i < 3; i++ {
		next, _ := model.Update(tea.KeyMsg{Type: tea.KeyRight})
		model = next.(Model)
	}

	state := model.dialogState()
	if state.Kind != "pomodoro_start" {
		t.Fatalf("expected pomodoro dialog to remain open, got %q", state.Kind)
	}
	if state.PomodoroFocusChoice != 3 {
		t.Fatalf("expected focus choice to reach custom after three right keys, got %d", state.PomodoroFocusChoice)
	}
	if state.FocusIdx != 1 {
		t.Fatalf("expected focus to move into custom input, got %d", state.FocusIdx)
	}
	if !state.Inputs[0].Focused() {
		t.Fatal("expected custom focus input to be focused")
	}

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = next.(Model)
	state = model.dialogState()
	if state.FocusIdx != 2 {
		t.Fatalf("expected enter from custom focus input to advance to short break row, got %d", state.FocusIdx)
	}
}

func TestDialogRuntimeDepsWireTelemetryHooks(t *testing.T) {
	model := Model{client: api.NewClient("unix", "/tmp/missing.sock")}
	deps := model.dialogRuntimeDeps()
	if deps.PatchTelemetrySettings == nil {
		t.Fatal("expected patch telemetry settings hook to be wired")
	}
	if deps.CompleteOnboarding == nil {
		t.Fatal("expected complete onboarding hook to be wired")
	}
}

func TestHardLimitExpiredDialogSeedsLivePomodoroConfig(t *testing.T) {
	model := Model{
		timer: &api.TimerState{
			State:                          "expired",
			HardLimitActive:                true,
			HardLimitExpired:               true,
			HardLimitWorkSeconds:           60,
			HardLimitBreakSeconds:          0,
			HardLimitLongBreakSeconds:      120,
			HardLimitCyclesBeforeLongBreak: 3,
		},
	}

	next := model.openHardLimitExpiredDialog("Issue title")
	if next.dialog != "hard_limit_expired" {
		t.Fatalf("expected hard-limit expired dialog, got %q", next.dialog)
	}
	if next.dialogHardLimitFocusSeconds != 60 {
		t.Fatalf("expected focus seconds to be seeded, got %d", next.dialogHardLimitFocusSeconds)
	}
	if next.dialogHardLimitBreakSeconds != 0 {
		t.Fatalf("expected short break seconds to remain disabled, got %d", next.dialogHardLimitBreakSeconds)
	}
	if next.dialogHardLimitLongBreakSeconds != 120 {
		t.Fatalf("expected long break seconds to be seeded, got %d", next.dialogHardLimitLongBreakSeconds)
	}
	if next.dialogHardLimitCyclesBeforeLongBreak != 3 {
		t.Fatalf(
			"expected cycles before long break to be seeded, got %d",
			next.dialogHardLimitCyclesBeforeLongBreak,
		)
	}
}

func TestHardLimitDialogRefreshesFromTimerSnapshot(t *testing.T) {
	model := Model{
		dialog:                               "hard_limit_extend",
		dialogHardLimitFocusSeconds:          25 * 60,
		dialogHardLimitBreakSeconds:          5 * 60,
		dialogHardLimitLongBreakSeconds:      15 * 60,
		dialogHardLimitCyclesBeforeLongBreak: 4,
		timer: &api.TimerState{
			State:                          "running",
			HardLimitActive:                true,
			HardLimitTotalSeconds:          120,
			HardLimitRemainingSeconds:      60,
			HardLimitWorkSeconds:           60,
			HardLimitBreakSeconds:          0,
			HardLimitLongBreakSeconds:      0,
			HardLimitCyclesBeforeLongBreak: 0,
		},
	}

	next := model.applyDispatchMessageState(model.dispatchMessageState())
	if next.dialogHardLimitFocusSeconds != 60 {
		t.Fatalf("expected focus seconds to refresh from timer, got %d", next.dialogHardLimitFocusSeconds)
	}
	if next.dialogHardLimitBreakSeconds != 0 {
		t.Fatalf("expected short break seconds to refresh from timer, got %d", next.dialogHardLimitBreakSeconds)
	}
	if next.dialogHardLimitLongBreakSeconds != 0 {
		t.Fatalf("expected long break seconds to refresh from timer, got %d", next.dialogHardLimitLongBreakSeconds)
	}
	if next.dialogHardLimitCyclesBeforeLongBreak != 0 {
		t.Fatalf(
			"expected cycles-before-long-break to refresh from timer, got %d",
			next.dialogHardLimitCyclesBeforeLongBreak,
		)
	}
}

func makeWellbeingRangeForAnchorTest(days int) []api.DailyMetricsDay {
	out := make([]api.DailyMetricsDay, 0, days)
	for i := 0; i < days; i++ {
		sleep := 7.0
		out = append(out, api.DailyMetricsDay{
			Date:          "2026-04-04",
			WorkedSeconds: 3600 + i*120,
			RestSeconds:   300 + i*30,
			CheckIn: &api.DailyCheckIn{
				Date:       "2026-04-04",
				Mood:       3 + (i % 3),
				Energy:     2 + (i % 4),
				SleepHours: &sleep,
			},
		})
	}
	return out
}
