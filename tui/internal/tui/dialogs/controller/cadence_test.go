package controller

import (
	"strings"
	"testing"

	sharedtypes "crona/shared/types"

	tea "github.com/charmbracelet/bubbletea"
)

func press(state State, keys ...tea.KeyMsg) (State, *Action, string) {
	var action *Action
	status := ""
	for _, key := range keys {
		state, action, status = Update(state, UpdateContext{}, "2026-10-18", key)
	}
	return state, action, status
}

var (
	keyRight = tea.KeyMsg{Type: tea.KeyRight}
	keyLeft  = tea.KeyMsg{Type: tea.KeyLeft}
	keyTab   = tea.KeyMsg{Type: tea.KeyTab}
	keySpace = tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}}
	keySave  = tea.KeyMsg{Type: tea.KeyCtrlS}
)

func editHabit(schedule string) State {
	return OpenEditHabit(State{}, 3, 2, "Read", nil, schedule, nil, true)
}

func TestHabitCadenceCyclesWithArrows(t *testing.T) {
	state := OpenCreateHabit(State{})
	state.FocusIdx = 4
	state = SyncDialogFocus(state)
	state, _, _ = press(state, keyRight)
	if got := state.Inputs[3].Value(); got != "weekdays" || state.CadenceIdx != 1 {
		t.Fatalf("right: schedule %q idx %d, want weekdays/1", got, state.CadenceIdx)
	}
	state, _, _ = press(state, keyLeft, keyLeft)
	if state.CadenceIdx != 2 {
		t.Fatalf("left wraps to Custom days, got %d", state.CadenceIdx)
	}
}

func TestHabitCadenceIgnoresTyping(t *testing.T) {
	state := editHabit("daily")
	state.FocusIdx = 2
	state = SyncDialogFocus(state)
	state, _, _ = press(state, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	if got := state.Inputs[1].Value(); got != "daily" {
		t.Fatalf("typing changed the cadence to %q", got)
	}
}

func TestHabitEditOpensOnItsCustomDays(t *testing.T) {
	state := editHabit("mon,wed")
	if state.CadenceIdx != 2 || len(state.CadenceDays) != 2 {
		t.Fatalf("cadence %d days %v, want custom mon,wed", state.CadenceIdx, state.CadenceDays)
	}
	state.FocusIdx = 4
	state.CadenceCursor = 0
	state = SyncDialogFocus(state)
	state, _, _ = press(state, keyRight, keySpace)
	if got := state.Inputs[1].Value(); got != "mon,tue,wed" {
		t.Fatalf("toggling Tue gave %q", got)
	}
}

func tabs(t *testing.T, state State, key tea.KeyMsg, want ...int) {
	t.Helper()
	for _, focus := range want {
		state, _, _ = press(state, key)
		if state.FocusIdx != focus {
			t.Fatalf(
				"%s %s: focus %d, want %d (order %v)",
				state.Kind,
				key,
				state.FocusIdx,
				focus,
				want,
			)
		}
	}
}

func TestHabitTabGoesCadenceDaysTarget(t *testing.T) {
	// edit_habit focus: Name 0, Description 1, Cadence 2, Target 3, Days 4.
	state := editHabit("daily")
	state.FocusIdx = 2
	tabs(t, state, keyTab, 3, 0)
	state = editHabit("mon")
	state.FocusIdx = 2
	tabs(t, state, keyTab, 4, 3, 0)
	state.FocusIdx = 3
	tabs(t, state, tea.KeyMsg{Type: tea.KeyShiftTab}, 4, 2)
}

func TestCustomCadenceNeedsADay(t *testing.T) {
	state := editHabit("mon")
	state.FocusIdx = 4
	state.CadenceCursor = 0
	state, _, _ = press(state, keySpace)
	_, action, status := press(state, keySave)
	if action != nil || status != "Pick at least one day" {
		t.Fatalf("got action %+v status %q", action, status)
	}
}

func TestHabitSavesItsCadence(t *testing.T) {
	_, action, _ := press(editHabit("tue,thu"), keySave)
	if action == nil || action.Status != "weekly" || len(action.Weekdays) != 2 {
		t.Fatalf("got %+v", action)
	}
}

func TestReminderCadence(t *testing.T) {
	state := OpenEditAlertReminder(State{}, sharedtypes.AlertReminder{
		ScheduleType: sharedtypes.AlertReminderScheduleWeekly, Weekdays: []int{1, 2, 3, 4, 5}, TimeHHMM: "20:00",
	})
	if state.CadenceIdx != 1 {
		t.Fatalf("weekdays reminder opened on cadence %d", state.CadenceIdx)
	}
	state, _, _ = press(state, keyRight)
	tabs(t, state, keyTab, 2, 1, 0)
}

func TestDateGridHeaderLinesUpWithDays(t *testing.T) {
	due := "2026-10-20"
	state := PopulateDatePresentation(
		Theme{},
		OpenDatePicker(State{}, "", 0, 0, &due, "2026-10-18"),
		"2026-10-18",
	)
	lines := strings.Split(state.DateGrid, "\n")
	if lines[0] != " Mo   Tu   We   Th   Fr   Sa   Su " {
		t.Fatalf("header %q does not match the 5-cell day stride of %q", lines[0], lines[1])
	}
}
