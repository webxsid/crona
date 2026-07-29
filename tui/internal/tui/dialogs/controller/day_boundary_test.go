package controller

import (
	"testing"

	sharedtypes "crona/shared/types"

	tea "github.com/charmbracelet/bubbletea"
)

func TestDayBoundaryEditorLoadsAndSavesSchedule(t *testing.T) {
	state := OpenEditDayBoundary(
		State{},
		sharedtypes.CoreSettingsKeyStartOfDay,
		sharedtypes.DayBoundarySchedule{
			Enabled:     true,
			DefaultTime: "08:30",
			WeekdayOverrides: map[int]string{
				1: "09:00",
			},
		},
		"Asia/Kolkata",
	)
	if state.Inputs[0].Value() != "08:30" || len(state.DayBoundaryOverrides) != 1 {
		t.Fatalf("schedule values were not loaded: default=%q overrides=%+v", state.Inputs[0].Value(), state.DayBoundaryOverrides)
	}
	if got := state.DayBoundaryOverrides[0]; got.Time != "09:00" || len(got.Days) != 1 || got.Days[0] != 1 {
		t.Fatalf("unexpected grouped override: %+v", got)
	}
	if state.DayBoundaryTimezone != "Asia/Kolkata" {
		t.Fatalf("timezone = %q", state.DayBoundaryTimezone)
	}

	next, action, status := Update(state, UpdateContext{}, "2026-07-29", tea.KeyMsg{Type: tea.KeyCtrlS})
	if status != "" {
		t.Fatalf("unexpected save status: %q", status)
	}
	if next.Kind != "" || action == nil {
		t.Fatalf("expected editor to close with save action: state=%q action=%+v", next.Kind, action)
	}
	if action.SettingKey != sharedtypes.CoreSettingsKeyStartOfDay || action.DayBoundarySchedule.DefaultTime != "08:30" {
		t.Fatalf("unexpected save action: %+v", action)
	}
}

func TestDayBoundaryEditorAddsMultiDayOverride(t *testing.T) {
	state := OpenEditDayBoundary(
		State{},
		sharedtypes.CoreSettingsKeyStartOfDay,
		sharedtypes.DayBoundarySchedule{Enabled: true, DefaultTime: "08:30"},
		"UTC",
	)
	state, _, _ = Update(state, UpdateContext{}, "2026-07-29", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if state.DayBoundaryStep != DayBoundaryStepDays {
		t.Fatalf("step = %d, want day selector", state.DayBoundaryStep)
	}
	state.DayBoundarySelectedDays = []int{1, 3, 5}
	state, _, status := Update(state, UpdateContext{}, "2026-07-29", tea.KeyMsg{Type: tea.KeyEnter})
	if status != "" || state.DayBoundaryStep != DayBoundaryStepTime {
		t.Fatalf("expected time step, state=%d status=%q", state.DayBoundaryStep, status)
	}
	state.Inputs[1].SetValue("09:15")
	state, _, status = Update(state, UpdateContext{}, "2026-07-29", tea.KeyMsg{Type: tea.KeyCtrlS})
	if status != "" || state.DayBoundaryStep != DayBoundaryStepOverview {
		t.Fatalf("expected overview after apply, step=%d status=%q", state.DayBoundaryStep, status)
	}
	if got := state.DayBoundarySchedule.WeekdayOverrides; got[1] != "09:15" || got[3] != "09:15" || got[5] != "09:15" {
		t.Fatalf("unexpected overrides: %+v", got)
	}
}

func TestDayBoundaryEditorDisablesConfiguredDays(t *testing.T) {
	state := OpenEditDayBoundary(
		State{},
		sharedtypes.CoreSettingsKeyEndOfDay,
		sharedtypes.DayBoundarySchedule{
			Enabled:     true,
			DefaultTime: "18:00",
			WeekdayOverrides: map[int]string{
				1: "17:00",
				3: "17:00",
			},
		},
		"UTC",
	)
	state, _, _ = Update(state, UpdateContext{}, "2026-07-29", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if DayBoundaryDayAvailable(state, 1) || DayBoundaryDayAvailable(state, 3) {
		t.Fatal("expected configured weekdays to be unavailable")
	}
	if !DayBoundaryDayAvailable(state, 2) {
		t.Fatal("expected unconfigured weekday to be available")
	}
}

func TestDayBoundaryEditorEditsAndRemovesGroupedOverride(t *testing.T) {
	state := OpenEditDayBoundary(
		State{},
		sharedtypes.CoreSettingsKeyStartOfDay,
		sharedtypes.DayBoundarySchedule{
			Enabled:     true,
			DefaultTime: "08:30",
			WeekdayOverrides: map[int]string{
				1: "09:00",
				3: "09:00",
			},
		},
		"UTC",
	)
	state, _, _ = Update(state, UpdateContext{}, "2026-07-29", tea.KeyMsg{Type: tea.KeyDown})
	state, _, _ = Update(state, UpdateContext{}, "2026-07-29", tea.KeyMsg{Type: tea.KeyEnter})
	if state.DayBoundaryStep != DayBoundaryStepDays {
		t.Fatalf("expected edit to open day selector, step=%d", state.DayBoundaryStep)
	}
	if !DayBoundaryDayAvailable(state, 1) || !DayBoundaryDayAvailable(state, 3) {
		t.Fatal("expected the edited override's days to remain selectable")
	}
	state, _, _ = Update(state, UpdateContext{}, "2026-07-29", tea.KeyMsg{Type: tea.KeyEsc})
	state, _, _ = Update(state, UpdateContext{}, "2026-07-29", tea.KeyMsg{Type: tea.KeyDelete})
	if len(state.DayBoundarySchedule.WeekdayOverrides) != 0 {
		t.Fatalf("expected override removal, got %+v", state.DayBoundarySchedule.WeekdayOverrides)
	}
}

func TestDayBoundaryEditorDisablesDefaultAndAddWhenAllDaysCovered(t *testing.T) {
	overrides := map[int]string{}
	for day := 1; day <= 7; day++ {
		overrides[day] = "09:00"
	}
	state := OpenEditDayBoundary(
		State{},
		sharedtypes.CoreSettingsKeyStartOfDay,
		sharedtypes.DayBoundarySchedule{Enabled: true, DefaultTime: "08:30", WeekdayOverrides: overrides},
		"UTC",
	)
	if !AllDayBoundaryDaysOverridden(state.DayBoundarySchedule) {
		t.Fatal("expected all weekdays to be covered")
	}
	if state.DayBoundaryOverrideCursor == 0 {
		t.Fatal("expected default row to be skipped when all weekdays are covered")
	}
	state, _, _ = Update(state, UpdateContext{}, "2026-07-29", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if state.DayBoundaryStep != DayBoundaryStepOverview {
		t.Fatal("add override should be disabled when all weekdays are covered")
	}
}

func TestDayBoundaryEditorRejectsInvalidTime(t *testing.T) {
	state := OpenEditDayBoundary(
		State{},
		sharedtypes.CoreSettingsKeyEndOfDay,
		sharedtypes.DayBoundarySchedule{Enabled: true, DefaultTime: "18:00"},
		"UTC",
	)
	state.Inputs[0].SetValue("8:00")
	next, action, status := Update(state, UpdateContext{}, "2026-07-29", tea.KeyMsg{Type: tea.KeyCtrlS})
	if action != nil || next.Kind != "edit_day_boundary" || status == "" {
		t.Fatalf("expected validation error, state=%q action=%+v status=%q", next.Kind, action, status)
	}
}
