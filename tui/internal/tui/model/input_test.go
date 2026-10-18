package model

import (
	"strings"
	"testing"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	dialogstate "crona/tui/internal/tui/dialogs/controller"
	inputpkg "crona/tui/internal/tui/input"
	uistate "crona/tui/internal/tui/state"
	tea "github.com/charmbracelet/bubbletea"
)

func TestInputDepsOpenCheckInDialogUsesCreateForMissingCheckIn(t *testing.T) {
	m := Model{
		date: "2026-05-27",
	}
	state := inputpkg.State{
		ActiveView: uistate.ViewDaily,
		Date:       "2026-05-27",
	}

	if !m.inputDeps().OpenCheckInDialog(&state) {
		t.Fatal("expected daily check-in dialog to open")
	}
	if state.DialogState.Kind != "create_checkin" {
		t.Fatalf("expected create_checkin dialog, got %q", state.DialogState.Kind)
	}
	if state.DialogState.CheckInDate != "2026-05-27" {
		t.Fatalf(
			"expected check-in dialog date to use dashboard date, got %q",
			state.DialogState.CheckInDate,
		)
	}
}

func TestInputDepsOpenCheckInDialogUsesSharedDateOnReviewView(t *testing.T) {
	m := Model{
		date: "2026-05-28",
		dailyCheckIn: &api.DailyCheckIn{
			Date: "2026-05-28",
		},
	}
	state := inputpkg.State{
		ActiveView: uistate.ViewReview,
		Date:       "2026-05-28",
		DailyCheckIn: &api.DailyCheckIn{
			Date: "2026-05-28",
		},
	}

	if !m.inputDeps().OpenCheckInDialog(&state) {
		t.Fatal("expected Review check-in dialog to open")
	}
	if state.DialogState.Kind != "edit_checkin" {
		t.Fatalf("expected edit_checkin dialog, got %q", state.DialogState.Kind)
	}
	if state.DialogState.CheckInDate != "2026-05-28" {
		t.Fatalf(
			"expected Review check-in dialog date to use wellbeing date, got %q",
			state.DialogState.CheckInDate,
		)
	}
}

func TestDialogStatePreservesSummaryExportRange(t *testing.T) {
	model := Model{
		dialog:                   "export_preset",
		dialogCheckInDate:        "2026-03-19",
		dialogExportCategory:     "summary",
		dialogExportStart:        "2026-03-13",
		dialogExportEnd:          "2026-03-19",
		dialogExportPresetKind:   sharedtypes.ExportReportKindSummaryRange,
		dialogExportPresetFormat: sharedtypes.ExportFormatMarkdown,
		dialogExportPresetOutput: sharedtypes.ExportOutputModeFile,
	}

	state := model.dialogState()
	if state.ExportCategory != "summary" || state.ExportStart != "2026-03-13" || state.ExportEnd != "2026-03-19" {
		t.Fatalf("dialogState dropped summary range: %#v", state)
	}

	next := model.withDialogState(dialogstate.State{
		Kind:               "export_preset",
		CheckInDate:        "2026-03-19",
		ExportCategory:     "summary",
		ExportStart:        "2026-03-13",
		ExportEnd:          "2026-03-19",
		ExportPresetKind:   sharedtypes.ExportReportKindSummaryRange,
		ExportPresetFormat: sharedtypes.ExportFormatMarkdown,
		ExportPresetOutput: sharedtypes.ExportOutputModeFile,
	})
	if next.dialogExportCategory != "summary" || next.dialogExportStart != "2026-03-13" || next.dialogExportEnd != "2026-03-19" {
		t.Fatalf("withDialogState dropped summary range: %#v", next)
	}
}

func TestSessionViewAdhocStartOpensTimerTypeDialog(t *testing.T) {
	model := Model{view: ViewSessionActive, timer: &api.TimerState{State: "idle"}}
	next := model.openAdhocTimerDialog()
	got := next.dialogState()
	if got.Kind != "timer_start_type" || !got.AdhocTimer {
		t.Fatalf("session view start dialog = %+v, want ad hoc timer choices", got)
	}
}

func TestAdhocShortcutPreservesMarkerIntoTimerStartRequest(t *testing.T) {
	model := Model{view: ViewDaily, pane: PaneIssues, timer: &api.TimerState{State: "idle"}}
	state, _ := inputpkg.Handle(model.inputState(), tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'A'}}, model.inputDeps())
	next := model.applyInputState(state)
	if next.dialog != "timer_start_type" || !next.dialogState().AdhocTimer {
		t.Fatalf("dialog after A = %q adhoc=%t, want ad hoc timer type", next.dialog, next.dialogState().AdhocTimer)
	}
	_, action, status := dialogstate.Update(next.dialogState(), dialogstate.UpdateContext{}, "2026-10-08", tea.KeyMsg{Type: tea.KeyEnter})
	if status != "" || action == nil || action.TimerStart == nil || !action.TimerStart.Adhoc {
		t.Fatalf("timer start request = %+v, status %q; want adhoc marker", action, status)
	}
}

func TestAdhocTimerEndOpensAttributionChoices(t *testing.T) {
	model := Model{timer: &api.TimerState{State: "running", Adhoc: true}}
	state := inputpkg.State{
		ActiveView: uistate.ViewSessionActive,
		Timer:      model.timer,
	}
	if !model.inputDeps().OpenEndSessionDialog(&state) {
		t.Fatal("OpenEndSessionDialog() = false, want ad hoc attribution dialog")
	}
	if state.DialogState.Kind != "adhoc_resolution" {
		t.Fatalf("end dialog kind = %q, want adhoc_resolution", state.DialogState.Kind)
	}
}

func TestAdhocStopwatchEndRendersAttributionDialog(t *testing.T) {
	model := Model{
		view:   ViewSessionActive,
		width:  160,
		height: 40,
		timer:  &api.TimerState{State: "running", Adhoc: true},
	}
	nextAny, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	next := nextAny.(Model)
	if next.dialog != "adhoc_resolution" {
		t.Fatalf("dialog after x = %q, want adhoc_resolution", next.dialog)
	}
	rendered := next.View()
	if !strings.Contains(rendered, "Finish Ad hoc Session") || !strings.Contains(rendered, "Attach to an existing issue") {
		t.Fatalf("ad hoc end dialog rendered blank or incomplete: %q", rendered)
	}
}

func TestDialogStatePreservesDayBoundarySettings(t *testing.T) {
	model := Model{}
	state := dialogstate.State{
		Kind:                "edit_day_boundary",
		DayBoundaryEnabled:  true,
		DayBoundaryTimezone: "Asia/Kolkata",
		SettingKey:          sharedtypes.CoreSettingsKeyStartOfDay,
	}

	next := model.withDialogState(state)
	got := next.dialogState()
	if !got.DayBoundaryEnabled {
		t.Fatal("expected day-boundary enabled state to survive model round-trip")
	}
	if got.DayBoundaryTimezone != state.DayBoundaryTimezone {
		t.Fatalf("timezone = %q, want %q", got.DayBoundaryTimezone, state.DayBoundaryTimezone)
	}
	if got.SettingKey != state.SettingKey {
		t.Fatalf("setting key = %q, want %q", got.SettingKey, state.SettingKey)
	}
}

func TestInputDepsOpenCheckInDialogUsesEditForExistingCheckIn(t *testing.T) {
	m := Model{
		date: "2026-05-27",
		dailyCheckIn: &api.DailyCheckIn{
			Date: "2026-05-27",
		},
	}
	state := inputpkg.State{
		ActiveView: uistate.ViewDaily,
		Date:       "2026-05-27",
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
		t.Fatalf(
			"expected check-in dialog date to use dashboard date, got %q",
			state.DialogState.CheckInDate,
		)
	}
}

func TestInputDepsOpenCreateActionUsesMomentumCreateDialog(t *testing.T) {
	model := Model{
		view:        ViewMomentum,
		pane:        PaneMomentumCards,
		momentumTab: MomentumTabCustom,
	}

	next := model.handleInputCreateAction()
	if next.dialog != "create_momentum" {
		t.Fatalf("expected create momentum dialog, got %q", next.dialog)
	}
	if got := next.dialogInputs[0].Placeholder; got != "Momentum name" {
		t.Fatalf("expected momentum dialog placeholder, got %q", got)
	}
}

func TestDailyIssueCreateOpensDefaultIssueDialog(t *testing.T) {
	model := Model{view: ViewDaily, pane: PaneIssues}
	next := model.handleInputCreateAction()

	if next.dialog != "create_issue_default" {
		t.Fatalf("daily issue create dialog = %q, want create_issue_default", next.dialog)
	}
	if next.dialogFocusIdx != 2 {
		t.Fatalf("daily issue create focus = %d, want title index 2", next.dialogFocusIdx)
	}
}

func TestInputDepsOpenEditorUsesMomentumEditDialog(t *testing.T) {
	model := Model{
		view:        ViewMomentum,
		pane:        PaneMomentumCards,
		momentumTab: MomentumTabCustom,
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

func TestMomentumEditKeyUsesEditDialog(t *testing.T) {
	model := Model{
		view:        ViewMomentum,
		pane:        PaneMomentumCards,
		momentumTab: MomentumTabCustom,
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

	nextModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	next := nextModel.(Model)
	if next.dialog != "edit_momentum" {
		t.Fatalf("expected e to open edit momentum dialog, got %q", next.dialog)
	}
}

func TestMomentumEnterUsesDetailsInsteadOfEdit(t *testing.T) {
	model := Model{
		view:               ViewMomentum,
		pane:               PaneMomentumCards,
		momentumTab:        MomentumTabCustom,
		date:               "2026-06-19",
		momentumWindowDays: 30,
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

	next, cmd, handled := model.handleInputEnter()
	if !handled {
		t.Fatal("expected momentum enter to be handled")
	}
	if cmd == nil {
		t.Fatal("expected momentum enter to load details")
	}
	if next.dialog == "edit_momentum" {
		t.Fatalf("expected enter to avoid edit dialog, got %q", next.dialog)
	}
}

func TestMomentumDeleteSelectionOpensConfirmDialog(t *testing.T) {
	model := Model{
		view:        ViewMomentum,
		pane:        PaneMomentumCards,
		momentumTab: MomentumTabCustom,
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

func TestMomentumEnterOpensDialogAtAnyWidth(t *testing.T) {
	for _, width := range []int{60, 200} {
		model := Model{
			view:               ViewMomentum,
			pane:               PaneMomentumCards,
			width:              width,
			date:               "2026-06-19",
			momentumWindowDays: 30,
			cursor:             map[Pane]int{PaneMomentumCards: 0},
			momentumCards: []api.MomentumCard{{
				Definition: sharedtypes.HabitStreakDefinition{ID: "momentum-1", Name: "Focus", Period: sharedtypes.HabitStreakPeriodDay},
			}},
		}
		if _, cmd, handled := model.handleInputEnter(); !handled || cmd == nil {
			t.Fatalf("width %d: expected enter to load the details dialog", width)
		}
	}
}
