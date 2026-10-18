package model

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	dialogstate "crona/tui/internal/tui/dialogs/controller"
)

func TestCadenceArrowsSurviveTheModel(t *testing.T) {
	m := Model{width: 120, height: 36, cursor: map[Pane]int{}, filters: map[Pane]string{}}
	m = m.withDialogState(dialogstate.OpenCreateAlertReminder(m.dialogState(), ""))
	next, _ := m.updateDialog(tea.KeyMsg{Type: tea.KeyRight})
	state := next.(Model).dialogState()
	if state.CadenceIdx != 1 || state.Inputs[0].Value() != "weekdays" {
		t.Fatalf("right on the cadence picker: idx %d schedule %q, want 1/weekdays", state.CadenceIdx, state.Inputs[0].Value())
	}
}
