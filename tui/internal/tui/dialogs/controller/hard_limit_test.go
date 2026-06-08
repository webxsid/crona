package controller

import (
	"testing"

	shareddto "crona/shared/dto"

	tea "github.com/charmbracelet/bubbletea"
)

func TestPomodoroStartDialogBuildsTimerStartRequest(t *testing.T) {
	state := OpenPomodoroStart(State{}, 11, 22, 33, "Issue title")

	next, action, status := Update(
		state,
		UpdateContext{},
		"2026-05-26",
		tea.KeyMsg{Type: tea.KeyCtrlS},
	)
	if status != "" {
		t.Fatalf("unexpected status %q", status)
	}
	if next.Kind != "" {
		t.Fatalf("expected pomodoro dialog to close, got %q", next.Kind)
	}
	if action == nil || action.Kind != "start_focus_session" || action.TimerStart == nil {
		t.Fatalf("unexpected action %+v", action)
	}

	want := &shareddto.TimerStartRequest{
		RepoID:                         int64Ptr(11),
		StreamID:                       int64Ptr(22),
		IssueID:                        int64Ptr(33),
		HardLimitTotalSeconds:          intPtr(7800),
		HardLimitWorkSeconds:           intPtr(1500),
		HardLimitBreakSeconds:          intPtr(300),
		HardLimitLongBreakSeconds:      intPtr(900),
		HardLimitCyclesBeforeLongBreak: intPtr(4),
	}
	if *action.TimerStart.RepoID != *want.RepoID ||
		*action.TimerStart.StreamID != *want.StreamID ||
		*action.TimerStart.IssueID != *want.IssueID ||
		*action.TimerStart.HardLimitTotalSeconds != *want.HardLimitTotalSeconds ||
		*action.TimerStart.HardLimitWorkSeconds != *want.HardLimitWorkSeconds ||
		*action.TimerStart.HardLimitBreakSeconds != *want.HardLimitBreakSeconds ||
		*action.TimerStart.HardLimitLongBreakSeconds != *want.HardLimitLongBreakSeconds ||
		*action.TimerStart.HardLimitCyclesBeforeLongBreak != *want.HardLimitCyclesBeforeLongBreak {
		t.Fatalf("unexpected timer start payload %+v", action.TimerStart)
	}
}

func TestTimerStartTypeDialogRoutesToStopwatchOrPomodoro(t *testing.T) {
	state := OpenTimerStartType(State{}, 11, 22, 33, "Issue title")

	next, action, status := Update(
		state,
		UpdateContext{},
		"2026-05-26",
		tea.KeyMsg{Type: tea.KeyEnter},
	)
	if status != "" {
		t.Fatalf("unexpected status %q", status)
	}
	if next.Kind != "" {
		t.Fatalf("expected stopwatch selection to close dialog, got %q", next.Kind)
	}
	if action == nil || action.Kind != "start_focus_session" || action.TimerStart == nil {
		t.Fatalf("unexpected action %+v", action)
	}
	if action.TimerStart.HardLimitTotalSeconds != nil ||
		action.TimerStart.HardLimitWorkSeconds != nil ||
		action.TimerStart.HardLimitBreakSeconds != nil ||
		action.TimerStart.HardLimitLongBreakSeconds != nil ||
		action.TimerStart.HardLimitCyclesBeforeLongBreak != nil {
		t.Fatalf("expected stopwatch timer start to omit pomodoro fields, got %+v", action.TimerStart)
	}

	state = OpenTimerStartType(State{}, 11, 22, 33, "Issue title")
	next, action, status = Update(
		state,
		UpdateContext{},
		"2026-05-26",
		tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}},
	)
	if status != "" {
		t.Fatalf("unexpected status %q", status)
	}
	if action == nil || action.Kind != "start_focus_session" || action.TimerStart == nil {
		t.Fatalf("unexpected action %+v", action)
	}
	if next.Kind != "" {
		t.Fatalf("expected stopwatch shortcut to close dialog, got %q", next.Kind)
	}

	state = OpenTimerStartType(State{}, 11, 22, 33, "Issue title")
	next, action, status = Update(
		state,
		UpdateContext{},
		"2026-05-26",
		tea.KeyMsg{Type: tea.KeyDown},
	)
	if status != "" || action != nil {
		t.Fatalf("unexpected down result status=%q action=%+v", status, action)
	}
	next, _, status = Update(
		next,
		UpdateContext{},
		"2026-05-26",
		tea.KeyMsg{Type: tea.KeyEnter},
	)
	if status != "" {
		t.Fatalf("unexpected status %q", status)
	}
	if next.Kind != "pomodoro_start" {
		t.Fatalf("expected pomodoro choice to open unified pomodoro dialog, got %q", next.Kind)
	}

	state = OpenTimerStartType(State{}, 11, 22, 33, "Issue title")
	next, action, status = Update(
		state,
		UpdateContext{},
		"2026-05-26",
		tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}},
	)
	if status != "" {
		t.Fatalf("unexpected status %q", status)
	}
	if action != nil {
		t.Fatalf("expected pomodoro shortcut to open dialog, got action %+v", action)
	}
	if next.Kind != "pomodoro_start" {
		t.Fatalf("expected pomodoro shortcut to open unified pomodoro dialog, got %q", next.Kind)
	}
}

func TestPomodoroInlineCustomFocusSubmitsInSameDialog(t *testing.T) {
	state := OpenPomodoroStart(State{}, 11, 22, 33, "Issue title")
	var next State
	var action *Action
	var status string

	for i := 0; i < 3; i++ {
		next, action, status = Update(
			state,
			UpdateContext{},
			"2026-05-26",
			tea.KeyMsg{Type: tea.KeyRight},
		)
		if status != "" || action != nil {
			t.Fatalf("unexpected right result status=%q action=%+v", status, action)
		}
		state = next
	}
	if state.PomodoroFocusChoice != pomodoroFocusCustomChoice {
		t.Fatalf("expected custom focus choice, got %d", state.PomodoroFocusChoice)
	}
	if state.FocusIdx != pomodoroFocusCustomIdx {
		t.Fatalf("expected right onto custom focus to enter inline input, got focus %d", state.FocusIdx)
	}
	if !state.Inputs[0].Focused() {
		t.Fatalf("expected custom focus input to be focused immediately")
	}

	next, action, status = Update(
		state,
		UpdateContext{},
		"2026-05-26",
		tea.KeyMsg{Type: tea.KeyEnter},
	)
	if status != "" || action != nil {
		t.Fatalf("unexpected enter result status=%q action=%+v", status, action)
	}
	if next.Kind != "pomodoro_start" || next.FocusIdx != pomodoroBreakRowIdx {
		t.Fatalf("expected enter from active custom focus input to advance to short break row, got %+v", next)
	}

	next.FocusIdx = pomodoroFocusCustomIdx
	next.Inputs[0].SetValue("45m")
	next.Inputs[3].SetValue("4")
	next.Inputs[4].SetValue("4")

	next, action, status = Update(
		next,
		UpdateContext{},
		"2026-05-26",
		tea.KeyMsg{Type: tea.KeyCtrlS},
	)
	if status != "" {
		t.Fatalf("unexpected status %q", status)
	}
	if next.Kind != "" {
		t.Fatalf("expected dialog to close, got %q", next.Kind)
	}
	if action == nil || action.Kind != "start_focus_session" || action.TimerStart == nil {
		t.Fatalf("unexpected action %+v", action)
	}
	if action.TimerStart.HardLimitWorkSeconds == nil || *action.TimerStart.HardLimitWorkSeconds != 45*60 {
		t.Fatalf("expected custom focus seconds, got %+v", action.TimerStart)
	}
	if action.TimerStart.HardLimitTotalSeconds == nil || *action.TimerStart.HardLimitTotalSeconds != 12600 {
		t.Fatalf("unexpected total seconds %+v", action.TimerStart)
	}
}

func TestPomodoroRowNavigationBlursCustomInputFocus(t *testing.T) {
	state := OpenPomodoroStart(State{}, 11, 22, 33, "Issue title")
	var next State
	var action *Action
	var status string

	for i := 0; i < 3; i++ {
		next, action, status = Update(
			state,
			UpdateContext{},
			"2026-05-26",
			tea.KeyMsg{Type: tea.KeyRight},
		)
		if status != "" || action != nil {
			t.Fatalf("unexpected right result status=%q action=%+v", status, action)
		}
		state = next
	}
	if !state.Inputs[0].Focused() {
		t.Fatalf("expected custom focus input to be focused")
	}

	next, action, status = Update(
		state,
		UpdateContext{},
		"2026-05-26",
		tea.KeyMsg{Type: tea.KeyEnter},
	)
	if status != "" || action != nil {
		t.Fatalf("unexpected enter result status=%q action=%+v", status, action)
	}
	if next.FocusIdx != pomodoroBreakRowIdx {
		t.Fatalf("expected focus to advance to short break row, got %d", next.FocusIdx)
	}
	if next.Inputs[0].Focused() {
		t.Fatalf("expected custom focus input to be blurred after leaving the row")
	}

	next, action, status = Update(
		next,
		UpdateContext{},
		"2026-05-26",
		tea.KeyMsg{Type: tea.KeyRight},
	)
	if status != "" || action != nil {
		t.Fatalf("unexpected right result status=%q action=%+v", status, action)
	}
	if next.PomodoroBreakChoice != 1 {
		t.Fatalf("expected right to change short break choice, got %d", next.PomodoroBreakChoice)
	}
}

func TestPomodoroLeftAtCustomInputStartReturnsToPresetRow(t *testing.T) {
	state := OpenPomodoroStart(State{}, 11, 22, 33, "Issue title")
	var next State
	var action *Action
	var status string

	for i := 0; i < 3; i++ {
		next, action, status = Update(
			state,
			UpdateContext{},
			"2026-05-26",
			tea.KeyMsg{Type: tea.KeyRight},
		)
		if status != "" || action != nil {
			t.Fatalf("unexpected right result status=%q action=%+v", status, action)
		}
		state = next
	}
	if state.FocusIdx != pomodoroFocusCustomIdx {
		t.Fatalf("expected focus custom input, got %d", state.FocusIdx)
	}

	state.Inputs[0].CursorStart()
	next, action, status = Update(
		state,
		UpdateContext{},
		"2026-05-26",
		tea.KeyMsg{Type: tea.KeyLeft},
	)
	if status != "" || action != nil {
		t.Fatalf("unexpected left result status=%q action=%+v", status, action)
	}
	if next.FocusIdx != pomodoroFocusRowIdx {
		t.Fatalf("expected left at cursor start to return to focus preset row, got %d", next.FocusIdx)
	}
	if next.Inputs[0].Focused() {
		t.Fatal("expected custom input to blur when returning to preset row")
	}
}

func TestPomodoroNoBreakDisablesLongBreakCyclesField(t *testing.T) {
	state := OpenPomodoroStart(State{}, 11, 22, 33, "Issue title")
	var next State
	var action *Action
	var status string

	next, action, status = Update(
		state,
		UpdateContext{},
		"2026-05-26",
		tea.KeyMsg{Type: tea.KeyEnter},
	)
	if status != "" || action != nil {
		t.Fatalf("unexpected enter result status=%q action=%+v", status, action)
	}
	state = next

	next, action, status = Update(
		state,
		UpdateContext{},
		"2026-05-26",
		tea.KeyMsg{Type: tea.KeyEnter},
	)
	if status != "" || action != nil {
		t.Fatalf("unexpected enter result status=%q action=%+v", status, action)
	}
	state = next

	for i := 0; i < 3; i++ {
		next, action, status = Update(
			state,
			UpdateContext{},
			"2026-05-26",
			tea.KeyMsg{Type: tea.KeyRight},
		)
		if status != "" || action != nil {
			t.Fatalf("unexpected right result status=%q action=%+v", status, action)
		}
		state = next
	}
	if state.PomodoroLongBreakChoice != pomodoroLongBreakNoBreakChoice {
		t.Fatalf("expected no-break long break choice, got %d", state.PomodoroLongBreakChoice)
	}

	next, action, status = Update(
		state,
		UpdateContext{},
		"2026-05-26",
		tea.KeyMsg{Type: tea.KeyTab},
	)
	if status != "" || action != nil {
		t.Fatalf("unexpected tab result status=%q action=%+v", status, action)
	}
	if next.FocusIdx != pomodoroCyclesRowIdx {
		t.Fatalf("expected disabled long break row to be skipped, got focus idx %d", next.FocusIdx)
	}

	next.Inputs[3].SetValue("2")

	next, action, status = Update(
		next,
		UpdateContext{},
		"2026-05-26",
		tea.KeyMsg{Type: tea.KeyCtrlS},
	)
	if status != "" {
		t.Fatalf("unexpected status %q", status)
	}
	if next.Kind != "" {
		t.Fatalf("expected dialog to close, got %q", next.Kind)
	}
	if action == nil || action.Kind != "start_focus_session" || action.TimerStart == nil {
		t.Fatalf("unexpected action %+v", action)
	}
	if action.TimerStart.HardLimitBreakSeconds == nil || *action.TimerStart.HardLimitBreakSeconds != 5*60 {
		t.Fatalf("expected default short break, got %+v", action.TimerStart)
	}
	if action.TimerStart.HardLimitLongBreakSeconds != nil {
		t.Fatalf("expected disabled long break to be omitted, got %+v", action.TimerStart)
	}
	if action.TimerStart.HardLimitCyclesBeforeLongBreak != nil {
		t.Fatalf("expected disabled cycles-before-long-break to be omitted, got %+v", action.TimerStart)
	}
	if action.TimerStart.HardLimitTotalSeconds == nil || *action.TimerStart.HardLimitTotalSeconds != 2*(25*60+5*60) {
		t.Fatalf("unexpected total seconds %+v", action.TimerStart)
	}
}

func TestPomodoroShortBreakNoBreakDisablesCyclesAndStartsContinuousFocus(t *testing.T) {
	state := OpenPomodoroStart(State{}, 11, 22, 33, "Issue title")
	var next State
	var action *Action
	var status string

	next, action, status = Update(
		state,
		UpdateContext{},
		"2026-05-26",
		tea.KeyMsg{Type: tea.KeyEnter},
	)
	if status != "" || action != nil {
		t.Fatalf("unexpected enter result status=%q action=%+v", status, action)
	}
	state = next

	for i := 0; i < 3; i++ {
		next, action, status = Update(
			state,
			UpdateContext{},
			"2026-05-26",
			tea.KeyMsg{Type: tea.KeyRight},
		)
		if status != "" || action != nil {
			t.Fatalf("unexpected right result status=%q action=%+v", status, action)
		}
		state = next
	}
	if state.PomodoroBreakChoice != pomodoroBreakNoBreakChoice {
		t.Fatalf("expected no-break short break choice, got %d", state.PomodoroBreakChoice)
	}

	next, action, status = Update(
		state,
		UpdateContext{},
		"2026-05-26",
		tea.KeyMsg{Type: tea.KeyTab},
	)
	if status != "" || action != nil {
		t.Fatalf("unexpected tab result status=%q action=%+v", status, action)
	}
	if next.FocusIdx != pomodoroFocusRowIdx {
		t.Fatalf("expected continuous mode to skip disabled rows back to focus, got %d", next.FocusIdx)
	}

	next, action, status = Update(
		next,
		UpdateContext{},
		"2026-05-26",
		tea.KeyMsg{Type: tea.KeyCtrlS},
	)
	if status != "" {
		t.Fatalf("unexpected status %q", status)
	}
	if next.Kind != "" {
		t.Fatalf("expected dialog to close, got %q", next.Kind)
	}
	if action == nil || action.Kind != "start_focus_session" || action.TimerStart == nil {
		t.Fatalf("unexpected action %+v", action)
	}
	if action.TimerStart.HardLimitTotalSeconds == nil || *action.TimerStart.HardLimitTotalSeconds != 25*60 {
		t.Fatalf("expected continuous focus total to equal focus duration, got %+v", action.TimerStart)
	}
	if action.TimerStart.HardLimitBreakSeconds != nil {
		t.Fatalf("expected short break to be omitted, got %+v", action.TimerStart)
	}
	if action.TimerStart.HardLimitLongBreakSeconds != nil {
		t.Fatalf("expected long break to be omitted in continuous focus mode, got %+v", action.TimerStart)
	}
	if action.TimerStart.HardLimitCyclesBeforeLongBreak != nil {
		t.Fatalf("expected cycles-before-long-break to be omitted, got %+v", action.TimerStart)
	}
}

func TestPomodoroExpiredDialogRoutesToCommitAndExtend(t *testing.T) {
	state := OpenHardLimitExpired(State{}, "Issue title")
	var next State
	var action *Action
	var status string

	next, action, status = Update(
		state,
		UpdateContext{},
		"2026-05-26",
		tea.KeyMsg{Type: tea.KeyEnter},
	)
	if status != "" {
		t.Fatalf("unexpected status %q", status)
	}
	if action != nil {
		t.Fatalf("expected no immediate action when opening commit dialog, got %+v", action)
	}
	if next.Kind != "end_session" || next.Parent != "hard_limit_expired" {
		t.Fatalf("expected commit path to open end_session with parent, got %+v", next)
	}

	state = OpenHardLimitExpired(State{}, "Issue title")
	next, action, status = Update(
		state,
		UpdateContext{},
		"2026-05-26",
		tea.KeyMsg{Type: tea.KeyDown},
	)
	if status != "" || action != nil {
		t.Fatalf("unexpected down result status=%q action=%+v", status, action)
	}
	next, _, status = Update(
		next,
		UpdateContext{},
		"2026-05-26",
		tea.KeyMsg{Type: tea.KeyEnter},
	)
	if status != "" {
		t.Fatalf("unexpected status %q", status)
	}
	if next.Kind != "hard_limit_extend" || next.Parent != "hard_limit_expired" {
		t.Fatalf("expected extend path to open hard_limit_extend, got %+v", next)
	}

	state = OpenHardLimitExpired(State{}, "Issue title")
	next, action, status = Update(
		state,
		UpdateContext{},
		"2026-05-26",
		tea.KeyMsg{Type: tea.KeyDown},
	)
	if status != "" || action != nil {
		t.Fatalf("unexpected down result status=%q action=%+v", status, action)
	}
	next, action, status = Update(
		next,
		UpdateContext{},
		"2026-05-26",
		tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'z'}},
	)
	if status != "" || action != nil {
		t.Fatalf("unexpected z result status=%q action=%+v", status, action)
	}
	if next.Kind != "hard_limit_expired" {
		t.Fatalf("expected z key to keep the expired dialog open, got %+v", next)
	}
	next, _, status = Update(
		next,
		UpdateContext{},
		"2026-05-26",
		tea.KeyMsg{Type: tea.KeyEnter},
	)
	if status != "" {
		t.Fatalf("unexpected status %q", status)
	}
	if next.Kind != "hard_limit_extend" || next.Parent != "hard_limit_expired" {
		t.Fatalf("expected extend path to open hard_limit_extend, got %+v", next)
	}
}

func TestPomodoroExpiredDialogIgnoresZKey(t *testing.T) {
	state := OpenHardLimitExpired(State{}, "Issue title")
	next, action, status := Update(
		state,
		UpdateContext{},
		"2026-05-26",
		tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'z'}},
	)
	if status != "" {
		t.Fatalf("unexpected status %q", status)
	}
	if action != nil {
		t.Fatalf("expected no action for z key, got %+v", action)
	}
	if next.Kind != "hard_limit_expired" {
		t.Fatalf("expected z key to be a no-op, got %+v", next)
	}
}

func TestPomodoroExtendDialogReturnsExtensionAction(t *testing.T) {
	state := OpenHardLimitExtend(State{})
	state.Inputs[3].SetValue("2")
	state.FocusIdx = pomodoroCyclesRowIdx
	state = SyncDialogFocus(state)

	next, action, status := Update(
		state,
		UpdateContext{},
		"2026-05-26",
		tea.KeyMsg{Type: tea.KeyEnter},
	)
	if status != "" {
		t.Fatalf("unexpected status %q", status)
	}
	if next.Kind != "" {
		t.Fatalf("expected extend dialog to close, got %q", next.Kind)
	}
	if action == nil || action.Kind != "extend_hard_limit" {
		t.Fatalf("unexpected action %+v", action)
	}
	if action.AdditionalSessions != 2 {
		t.Fatalf("expected two-session extension, got %d", action.AdditionalSessions)
	}
}

func TestPomodoroExtendDialogMatchesPomodoroStartInputChrome(t *testing.T) {
	start := OpenPomodoroStart(State{}, 11, 22, 33, "Issue title")
	extend := OpenHardLimitExtend(State{
		ViewName:                       "Issue title",
		HardLimitTotalSeconds:          7800,
		HardLimitFocusSeconds:          25 * 60,
		HardLimitBreakSeconds:          5 * 60,
		HardLimitLongBreakSeconds:      15 * 60,
		HardLimitCyclesBeforeLongBreak: 4,
	})

	if len(start.Inputs) != len(extend.Inputs) {
		t.Fatalf("expected same number of inputs, got %d and %d", len(start.Inputs), len(extend.Inputs))
	}
	for i := range start.Inputs {
		if got, want := extend.Inputs[i].View(), start.Inputs[i].View(); got != want {
			t.Fatalf("expected input %d to match pomodoro start chrome, got %q want %q", i, got, want)
		}
	}
}

func TestPomodoroExtendDialogPreservesCustomValuesAcrossPresetRoundTrip(t *testing.T) {
	base := State{
		ViewName:                       "Issue title",
		HardLimitTotalSeconds:          7800,
		HardLimitFocusSeconds:          25 * 60,
		HardLimitBreakSeconds:          5 * 60,
		HardLimitLongBreakSeconds:      15 * 60,
		HardLimitCyclesBeforeLongBreak: 4,
	}

	press := func(state State, key tea.KeyMsg) State {
		next, action, status := Update(state, UpdateContext{}, "2026-05-26", key)
		if status != "" {
			t.Fatalf("unexpected status %q", status)
		}
		if action != nil {
			t.Fatalf("unexpected action %+v", action)
		}
		return next
	}

	type roundTripCase struct {
		name         string
		inputIdx     int
		customValue  string
		setup        func(State) State
		preset       func(State) State
		returnToCust func(State) State
	}

	cases := []roundTripCase{
		{
			name:        "focus",
			inputIdx:    0,
			customValue: "47m",
			setup: func(state State) State {
				for i := 0; i < 3; i++ {
					state = press(state, tea.KeyMsg{Type: tea.KeyRight})
				}
				return state
			},
			preset: func(state State) State {
				return press(state, tea.KeyMsg{Type: tea.KeyLeft})
			},
			returnToCust: func(state State) State {
				return press(state, tea.KeyMsg{Type: tea.KeyRight})
			},
		},
		{
			name:        "break",
			inputIdx:    1,
			customValue: "11m",
			setup: func(state State) State {
				for i := 0; i < 3; i++ {
					state = press(state, tea.KeyMsg{Type: tea.KeyRight})
				}
				state = press(state, tea.KeyMsg{Type: tea.KeyEnter})
				for i := 0; i < 4; i++ {
					state = press(state, tea.KeyMsg{Type: tea.KeyRight})
				}
				return state
			},
			preset: func(state State) State {
				for i := 0; i < 2; i++ {
					state = press(state, tea.KeyMsg{Type: tea.KeyLeft})
				}
				return state
			},
			returnToCust: func(state State) State {
				for i := 0; i < 2; i++ {
					state = press(state, tea.KeyMsg{Type: tea.KeyRight})
				}
				return state
			},
		},
		{
			name:        "long break",
			inputIdx:    2,
			customValue: "22m",
			setup: func(state State) State {
				for i := 0; i < 3; i++ {
					state = press(state, tea.KeyMsg{Type: tea.KeyRight})
				}
				state = press(state, tea.KeyMsg{Type: tea.KeyEnter})
				state = press(state, tea.KeyMsg{Type: tea.KeyEnter})
				for i := 0; i < 4; i++ {
					state = press(state, tea.KeyMsg{Type: tea.KeyRight})
				}
				return state
			},
			preset: func(state State) State {
				for i := 0; i < 2; i++ {
					state = press(state, tea.KeyMsg{Type: tea.KeyLeft})
				}
				return state
			},
			returnToCust: func(state State) State {
				for i := 0; i < 2; i++ {
					state = press(state, tea.KeyMsg{Type: tea.KeyRight})
				}
				return state
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			state := OpenHardLimitExtend(base)
			state = tc.setup(state)
			state.Inputs[tc.inputIdx].SetValue(tc.customValue)
			state.Inputs[tc.inputIdx].CursorStart()

			state = press(state, tea.KeyMsg{Type: tea.KeyLeft})
			state = tc.preset(state)
			state = tc.returnToCust(state)

			if got := state.Inputs[tc.inputIdx].Value(); got != tc.customValue {
				t.Fatalf("expected custom value to survive round-trip, got %q want %q", got, tc.customValue)
			}
		})
	}
}

func TestPomodoroExtendDialogDisablesCyclesWhenBreakDisabled(t *testing.T) {
	state := OpenHardLimitExtend(State{
		ViewName:                       "Issue title",
		HardLimitTotalSeconds:          25 * 60,
		HardLimitFocusSeconds:          25 * 60,
		HardLimitBreakSeconds:          0,
		HardLimitLongBreakSeconds:      0,
		HardLimitCyclesBeforeLongBreak: 0,
	})
	vm := BuildPomodoroDialogViewModel(state)
	if !vm.CyclesDisabled {
		t.Fatalf("expected cycles to be disabled when short break is disabled")
	}
	if pomodoroFocusIdxEnabled(state, pomodoroCyclesRowIdx) {
		t.Fatalf("expected cycles row to be skipped when short break is disabled")
	}
	if got := vm.CyclesSummary; got != "Uninterrupted focus" {
		t.Fatalf("expected uninterrupted-focus summary, got %q", got)
	}
}
