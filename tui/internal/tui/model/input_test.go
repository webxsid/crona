package model

import (
	"testing"

	sharedtypes "crona/shared/types"
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

func TestInputDepsOpenCreateActionUsesMomentumCreateDialog(t *testing.T) {
	model := Model{
		view: ViewMomentum,
		pane: PaneMomentumCards,
	}

	next := model.handleInputCreateAction()
	if next.dialog != "create_momentum" {
		t.Fatalf("expected create momentum dialog, got %q", next.dialog)
	}
	if got := next.dialogInputs[0].Placeholder; got != "Momentum name" {
		t.Fatalf("expected momentum dialog placeholder, got %q", got)
	}
}

func TestInputDepsOpenEditorUsesMomentumEditDialog(t *testing.T) {
	model := Model{
		view: ViewMomentum,
		pane: PaneMomentumCards,
		cursor: map[Pane]int{
			PaneMomentumCards: 0,
		},
		momentumCards: []api.MomentumCard{
			{
				Definition: sharedtypes.HabitStreakDefinition{
					ID:     "momentum-1",
					Name:   "Focus",
					Period: sharedtypes.HabitStreakPeriodDay,
				},
			},
		},
	}

	nextModel, _, handled := model.handleInputOpenEditor()
	if !handled {
		t.Fatal("expected momentum edit to be handled")
	}
	next := nextModel
	if next.dialog != "edit_momentum" {
		t.Fatalf("expected edit momentum dialog, got %q", next.dialog)
	}
}

func TestMomentumDeleteSelectionOpensConfirmDialog(t *testing.T) {
	model := Model{
		view: ViewMomentum,
		pane: PaneMomentumCards,
		cursor: map[Pane]int{
			PaneMomentumCards: 0,
		},
		momentumCards: []api.MomentumCard{
			{
				Definition: sharedtypes.HabitStreakDefinition{
					ID:   "momentum-1",
					Name: "Focus",
				},
			},
		},
	}

	next, cmd, handled := model.handleInputDeleteSelection()
	if !handled {
		t.Fatal("expected momentum delete to be handled")
	}
	if cmd != nil {
		t.Fatal("expected delete selection to defer to confirmation dialog")
	}
	if next.dialog != "confirm_delete" {
		t.Fatalf("expected confirm delete dialog, got %q", next.dialog)
	}
	if next.dialogDeleteKind != "momentum" || next.dialogDeleteID != "momentum-1" {
		t.Fatalf(
			"expected momentum delete payload, got kind=%q id=%q",
			next.dialogDeleteKind,
			next.dialogDeleteID,
		)
	}
}
