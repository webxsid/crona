package controller

import (
	"testing"

	shareddto "crona/shared/dto"

	tea "github.com/charmbracelet/bubbletea"
)

func TestPomodoroStartDialogBuildsTimerStartRequest(t *testing.T) {
	state := OpenPomodoroStart(State{}, 11, 22, 33, "Issue title", 90, 4, 15)
	state.PomodoroFocusSeconds = 25 * 60
	state.PomodoroBreakSeconds = 5 * 60

	next, _, status := Update(
		state,
		UpdateContext{},
		"2026-05-26",
		tea.KeyMsg{Type: tea.KeyCtrlS},
	)
	if status != "" {
		t.Fatalf("unexpected status %q", status)
	}
	if next.Kind != "" {
		t.Fatalf("expected pomodoro setup dialog to close, got %q", next.Kind)
	}

	_, action, status := Update(
		state,
		UpdateContext{},
		"2026-05-26",
		tea.KeyMsg{Type: tea.KeyCtrlS},
	)
	if status != "" {
		t.Fatalf("unexpected status %q", status)
	}
	if action == nil || action.Kind != "start_focus_session" || action.TimerStart == nil {
		t.Fatalf("unexpected action %+v", action)
	}

	want := &shareddto.TimerStartRequest{
		RepoID:                         int64Ptr(11),
		StreamID:                       int64Ptr(22),
		IssueID:                        int64Ptr(33),
		HardLimitTotalSeconds:          intPtr(5400),
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
		tea.KeyMsg{Type: tea.KeyDown},
	)
	if status != "" || action != nil {
		t.Fatalf("unexpected down result status=%q action=%+v", status, action)
	}
	next, action, status = Update(
		next,
		UpdateContext{},
		"2026-05-26",
		tea.KeyMsg{Type: tea.KeyEnter},
	)
	if status != "" {
		t.Fatalf("unexpected status %q", status)
	}
	if next.Kind != "pomodoro_focus_presets" {
		t.Fatalf("expected pomodoro choice to open focus presets, got %q", next.Kind)
	}
}

func TestPomodoroPresetDialogsRouteThroughFocusBreakAndCustom(t *testing.T) {
	focus := OpenPomodoroFocusPreset(State{}, 11, 22, 33, "Issue title")
	next, action, status := Update(
		focus,
		UpdateContext{},
		"2026-05-26",
		tea.KeyMsg{Type: tea.KeyEnter},
	)
	if status != "" || action != nil {
		t.Fatalf("unexpected focus preset result status=%q action=%+v", status, action)
	}
	if next.Kind != "pomodoro_break_presets" {
		t.Fatalf("expected focus preset to open break presets, got %q", next.Kind)
	}

	focus = OpenPomodoroFocusPreset(State{}, 11, 22, 33, "Issue title")
	for i := 0; i < 3; i++ {
		next, action, status = Update(
			focus,
			UpdateContext{},
			"2026-05-26",
			tea.KeyMsg{Type: tea.KeyDown},
		)
		if status != "" || action != nil {
			t.Fatalf("unexpected down result status=%q action=%+v", status, action)
		}
		focus = next
	}
	next, action, status = Update(
		focus,
		UpdateContext{},
		"2026-05-26",
		tea.KeyMsg{Type: tea.KeyEnter},
	)
	if status != "" || action != nil {
		t.Fatalf("unexpected custom focus result status=%q action=%+v", status, action)
	}
	if next.Kind != "pomodoro_focus_custom" {
		t.Fatalf("expected custom focus path, got %q", next.Kind)
	}

	breaks := OpenPomodoroBreakPreset(State{}, 11, 22, 33, "Issue title")
	next, action, status = Update(
		breaks,
		UpdateContext{},
		"2026-05-26",
		tea.KeyMsg{Type: tea.KeyEnter},
	)
	if status != "" || action != nil {
		t.Fatalf("unexpected break preset result status=%q action=%+v", status, action)
	}
	if next.Kind != "pomodoro_start" {
		t.Fatalf("expected break preset to open setup dialog, got %q", next.Kind)
	}

	breaks = OpenPomodoroBreakPreset(State{}, 11, 22, 33, "Issue title")
	for i := 0; i < 3; i++ {
		next, action, status = Update(
			breaks,
			UpdateContext{},
			"2026-05-26",
			tea.KeyMsg{Type: tea.KeyDown},
		)
		if status != "" || action != nil {
			t.Fatalf("unexpected down result status=%q action=%+v", status, action)
		}
		breaks = next
	}
	next, action, status = Update(
		breaks,
		UpdateContext{},
		"2026-05-26",
		tea.KeyMsg{Type: tea.KeyEnter},
	)
	if status != "" || action != nil {
		t.Fatalf("unexpected break custom result status=%q action=%+v", status, action)
	}
	if next.Kind != "pomodoro_break_custom" {
		t.Fatalf("expected custom break path, got %q", next.Kind)
	}
}

func TestPomodoroSetupCustomFieldsRouteToTimerStartRequest(t *testing.T) {
	state := OpenPomodoroStart(State{}, 11, 22, 33, "Issue title", 90, 4, 15)
	state.PomodoroFocusSeconds = 25 * 60
	state.PomodoroBreakSeconds = 5 * 60
	state.Inputs[0].SetValue("2h")
	state.Inputs[1].SetValue("6")
	state.Inputs[2].SetValue("20m")

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
		t.Fatalf("expected setup dialog to close, got %q", next.Kind)
	}
	if action == nil || action.Kind != "start_focus_session" || action.TimerStart == nil {
		t.Fatalf("unexpected action %+v", action)
	}
	if action.TimerStart.HardLimitTotalSeconds == nil ||
		*action.TimerStart.HardLimitTotalSeconds != 2*60*60 ||
		action.TimerStart.HardLimitWorkSeconds == nil ||
		*action.TimerStart.HardLimitWorkSeconds != 25*60 ||
		action.TimerStart.HardLimitBreakSeconds == nil ||
		*action.TimerStart.HardLimitBreakSeconds != 5*60 ||
		action.TimerStart.HardLimitLongBreakSeconds == nil ||
		*action.TimerStart.HardLimitLongBreakSeconds != 20*60 ||
		action.TimerStart.HardLimitCyclesBeforeLongBreak == nil ||
		*action.TimerStart.HardLimitCyclesBeforeLongBreak != 6 {
		t.Fatalf("unexpected setup timer start payload %+v", action.TimerStart)
	}
}

func TestPomodoroExpiredDialogRoutesToCommitStashAndExtend(t *testing.T) {
	state := OpenHardLimitExpired(State{}, "Issue title")

	next, action, status := Update(
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
	next, action, status = Update(
		next,
		UpdateContext{},
		"2026-05-26",
		tea.KeyMsg{Type: tea.KeyEnter},
	)
	if status != "" {
		t.Fatalf("unexpected status %q", status)
	}
	if next.Kind != "stash_session" || next.Parent != "hard_limit_expired" {
		t.Fatalf("expected stash path to open stash_session with parent, got %+v", next)
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
		tea.KeyMsg{Type: tea.KeyDown},
	)
	if status != "" || action != nil {
		t.Fatalf("unexpected down result status=%q action=%+v", status, action)
	}
	next, action, status = Update(
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

func TestPomodoroExtendDialogReturnsExtensionAction(t *testing.T) {
	state := OpenHardLimitExtend(State{})

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
	if action.AdditionalSeconds != 600 {
		t.Fatalf("expected 10-minute extension, got %d", action.AdditionalSeconds)
	}
}
