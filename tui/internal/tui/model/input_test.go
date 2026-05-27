package model

import (
	"testing"

	"crona/tui/internal/api"
	inputpkg "crona/tui/internal/tui/input"
	uistate "crona/tui/internal/tui/state"
)

func TestInputDepsOpenCheckInDialogUsesCreateForMissingCheckIn(t *testing.T) {
	m := Model{
		dashboardDate: "2026-05-27",
	}
	state := inputpkg.State{
		ActiveView:    uistate.ViewDaily,
		DashboardDate: "2026-05-27",
	}

	if !m.inputDeps().OpenCheckInDialog(&state) {
		t.Fatal("expected daily check-in dialog to open")
	}
	if state.DialogState.Kind != "create_checkin" {
		t.Fatalf("expected create_checkin dialog, got %q", state.DialogState.Kind)
	}
	if state.DialogState.CheckInDate != "2026-05-27" {
		t.Fatalf("expected check-in dialog date to use dashboard date, got %q", state.DialogState.CheckInDate)
	}
}

func TestInputDepsOpenCheckInDialogUsesWellbeingDateForWellbeingView(t *testing.T) {
	m := Model{
		dashboardDate: "2026-05-27",
		wellbeingDate: "2026-05-28",
		dailyCheckIn: &api.DailyCheckIn{
			Date: "2026-05-28",
		},
	}
	state := inputpkg.State{
		ActiveView:    uistate.ViewWellbeing,
		DashboardDate: "2026-05-27",
		WellbeingDate: "2026-05-28",
		DailyCheckIn: &api.DailyCheckIn{
			Date: "2026-05-28",
		},
	}

	if !m.inputDeps().OpenCheckInDialog(&state) {
		t.Fatal("expected wellbeing check-in dialog to open")
	}
	if state.DialogState.Kind != "edit_checkin" {
		t.Fatalf("expected edit_checkin dialog, got %q", state.DialogState.Kind)
	}
	if state.DialogState.CheckInDate != "2026-05-28" {
		t.Fatalf("expected wellbeing check-in dialog date to use wellbeing date, got %q", state.DialogState.CheckInDate)
	}
}

func TestInputDepsOpenCheckInDialogUsesEditForExistingCheckIn(t *testing.T) {
	m := Model{
		dashboardDate: "2026-05-27",
		dailyCheckIn: &api.DailyCheckIn{
			Date: "2026-05-27",
		},
	}
	state := inputpkg.State{
		ActiveView:    uistate.ViewDaily,
		DashboardDate: "2026-05-27",
		DailyCheckIn: &api.DailyCheckIn{
			Date: "2026-05-27",
		},
	}

	if !m.inputDeps().OpenCheckInDialog(&state) {
		t.Fatal("expected daily check-in dialog to open")
	}
	if state.DialogState.Kind != "edit_checkin" {
		t.Fatalf("expected edit_checkin dialog, got %q", state.DialogState.Kind)
	}
	if state.DialogState.CheckInDate != "2026-05-27" {
		t.Fatalf("expected check-in dialog date to use dashboard date, got %q", state.DialogState.CheckInDate)
	}
}
