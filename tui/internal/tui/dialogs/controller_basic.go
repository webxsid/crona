package dialogs

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func updateSingleInput(state State, msg tea.KeyMsg, requiredMsg string, submit func(string) *Action) (State, *Action, string) {
	if len(state.Inputs) == 0 {
		return Close(state), nil, "dialog input unavailable"
	}
	switch msg.String() {
	case "esc":
		return Close(state), nil, ""
	default:
		if isDialogSubmitKey(state, msg.String()) {
			name := strings.TrimSpace(state.Inputs[0].Value())
			if name == "" {
				return state, nil, requiredMsg
			}
			return Close(state), submit(name), ""
		}
	}
	var cmd tea.Cmd
	state.Inputs[0], cmd = state.Inputs[0].Update(msg)
	_ = cmd
	return clearDialogError(state), nil, ""
}

func updateNameDescription(state State, msg tea.KeyMsg, requiredMsg string, submit func(string, *string) *Action) (State, *Action, string) {
	switch msg.String() {
	case "esc":
		return Close(state), nil, ""
	case "tab", "shift+tab", "down", "up":
		dir := 1
		if msg.String() == "shift+tab" || msg.String() == "up" {
			dir = -1
		}
		state.FocusIdx = (state.FocusIdx + dir + 2) % 2
		state = SyncDialogFocus(state)
		return clearDialogError(state), nil, ""
	default:
		if isDialogSubmitKey(state, msg.String()) {
			name := strings.TrimSpace(state.Inputs[0].Value())
			if name == "" {
				return state, nil, requiredMsg
			}
			return Close(state), submit(name, ValueToPointer(strings.TrimSpace(state.Description.Value()))), ""
		}
	}
	if state.DescriptionEnabled && state.FocusIdx == state.DescriptionIndex {
		var cmd tea.Cmd
		state.Description, cmd = state.Description.Update(msg)
		_ = cmd
		return clearDialogError(state), nil, ""
	}
	var cmd tea.Cmd
	state.Inputs[state.FocusIdx], cmd = state.Inputs[state.FocusIdx].Update(msg)
	_ = cmd
	return clearDialogError(state), nil, ""
}

func updateCreateScratchpad(state State, msg tea.KeyMsg) (State, *Action, string) {
	switch msg.String() {
	case "esc":
		return Close(state), nil, ""
	case "tab", "shift+tab", "down", "up":
		dir := 1
		if msg.String() == "shift+tab" || msg.String() == "up" {
			dir = -1
		}
		state.FocusIdx = (state.FocusIdx + dir + len(state.Inputs)) % len(state.Inputs)
		state = SyncDialogFocus(state)
		return clearDialogError(state), nil, ""
	default:
		if isDialogSubmitKey(state, msg.String()) {
			name := strings.TrimSpace(state.Inputs[0].Value())
			path := strings.TrimSpace(state.Inputs[1].Value())
			if name == "" || path == "" {
				return state, nil, "Name and path are required"
			}
			return Close(state), &Action{Kind: "create_scratchpad", Name: name, Path: path}, ""
		}
	}
	var cmd tea.Cmd
	state.Inputs[state.FocusIdx], cmd = state.Inputs[state.FocusIdx].Update(msg)
	_ = cmd
	return clearDialogError(state), nil, ""
}

func updateConfirmDelete(state State, msg tea.KeyMsg) (State, *Action, string) {
	switch msg.String() {
	case "esc":
		return Close(state), nil, ""
	case "enter":
		action := &Action{Kind: "delete", ID: state.DeleteID, RepoID: state.RepoID, StreamID: state.StreamID}
		action.Name = state.DeleteKind
		action.Title = state.DeleteLabel
		return Close(state), action, ""
	}
	return state, nil, ""
}

func updateConfirmWipe(state State, msg tea.KeyMsg) (State, *Action, string) {
	switch msg.String() {
	case "esc":
		return Close(state), nil, ""
	case "enter":
		return Close(state), &Action{Kind: "wipe_runtime_data"}, ""
	}
	return state, nil, ""
}

func updateConfirmUninstall(state State, msg tea.KeyMsg) (State, *Action, string) {
	switch msg.String() {
	case "esc":
		return Close(state), nil, ""
	case "enter":
		return Close(state), &Action{Kind: "uninstall_crona"}, ""
	}
	return state, nil, ""
}
