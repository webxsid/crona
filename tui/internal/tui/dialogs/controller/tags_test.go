package controller

import (
	"slices"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func typeText(state State, text string) State {
	for _, r := range text {
		state, _, _ = updateTagPicker(state, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	return state
}

func TestTagPickerTogglesAndCreates(t *testing.T) {
	state := OpenTagPicker(State{}, TagTargetIssue, 7, "Fix login", []string{"urgent"}, []string{"quick", "urgent"})
	rows := TagPickerRows(state)
	if len(rows) != 2 || rows[0].Name != "quick" || rows[0].Checked || !rows[1].Checked {
		t.Fatalf("unexpected rows %+v", rows)
	}
	state, _, _ = updateTagPicker(state, tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}})
	state = typeText(state, "client")
	rows = TagPickerRows(state)
	if len(rows) != 1 || !rows[0].Create || rows[0].Name != "client" {
		t.Fatalf("expected a create row, got %+v", rows)
	}
	_, action, _ := updateTagPicker(state, tea.KeyMsg{Type: tea.KeyEnter})
	if action == nil || action.Kind != "set_tags" || action.IssueID != 7 ||
		!slices.Equal(action.StringList, []string{"client", "quick", "urgent"}) {
		t.Fatalf("unexpected action %+v", action)
	}
}

func TestTagPickerFiltersAndUnchecks(t *testing.T) {
	state := OpenTagPicker(State{}, TagTargetHabit, 3, "Walk", []string{"health"}, []string{"health", "quick"})
	state = typeText(state, "hea")
	rows := TagPickerRows(state)
	if len(rows) != 2 || !rows[0].Create || rows[1].Name != "health" {
		t.Fatalf("expected create row and the match, got %+v", rows)
	}
	state, _, _ = updateTagPicker(state, tea.KeyMsg{Type: tea.KeyDown})
	state, _, _ = updateTagPicker(state, tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}})
	state.Inputs[0].SetValue("")
	_, action, _ := updateTagPicker(state, tea.KeyMsg{Type: tea.KeyEnter})
	if action == nil || action.HabitID != 3 || len(action.StringList) != 0 {
		t.Fatalf("expected habit tags cleared, got %+v", action)
	}
}
