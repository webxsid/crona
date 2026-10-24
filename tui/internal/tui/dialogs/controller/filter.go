package controller

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// OpenPaneFilter opens the filter for one list pane, prefilled with its current filter.
func OpenPaneFilter(state State, pane, current string) State {
	input := textinput.New()
	input.Placeholder = "text  +tag  -tag"
	input.SetValue(current)
	input.CursorEnd()
	input.CharLimit = 120
	input.Width = 52
	input.Focus()
	state = Close(state)
	state.Kind = "edit_filter"
	state.Inputs = []textinput.Model{input}
	state.ViewName = pane
	return state
}

func updatePaneFilter(state State, msg tea.KeyMsg) (State, *Action, string) {
	switch msg.String() {
	case "esc":
		return Close(state), nil, ""
	case "enter":
		value := strings.Join(strings.Fields(state.Inputs[0].Value()), " ")
		return Close(state), &Action{Kind: "set_filter", Name: state.ViewName, Path: value}, ""
	}
	var cmd tea.Cmd
	state.Inputs[0], cmd = state.Inputs[0].Update(msg)
	_ = cmd
	return clearDialogError(state), nil, ""
}
