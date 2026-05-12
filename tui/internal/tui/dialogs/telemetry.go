package dialogs

import tea "github.com/charmbracelet/bubbletea"

func updateTelemetrySettings(state State, msg tea.KeyMsg) (State, *Action, string) {
	switch state.TelemetryStep {
	case 0:
		return updateTelemetryUsage(state, msg)
	case 1:
		return updateTelemetryErrors(state, msg)
	case 2:
		return updateTelemetryReview(state, msg)
	default:
		return state, nil, ""
	}
}

func updateTelemetryUsage(state State, msg tea.KeyMsg) (State, *Action, string) {
	switch msg.String() {
	case "esc":
		return Close(state), nil, ""
	case "enter":
		state.TelemetryStep = 1
		return clearDialogError(state), nil, ""
	case " ", "x":
		state.TelemetryUsage = !state.TelemetryUsage
		return clearDialogError(state), nil, ""
	case "tab", "right", "l":
		state.TelemetryStep = 1
		return clearDialogError(state), nil, ""
	}
	return state, nil, ""
}

func updateTelemetryErrors(state State, msg tea.KeyMsg) (State, *Action, string) {
	switch msg.String() {
	case "esc":
		return Close(state), nil, ""
	case " ", "x":
		state.TelemetryErrors = !state.TelemetryErrors
		return clearDialogError(state), nil, ""
	case "shift+tab", "left", "h":
		state.TelemetryStep = 0
		return clearDialogError(state), nil, ""
	case "tab", "right", "l", "enter":
		state.TelemetryStep = 2
		state.ChoiceCursor = 0
		return clearDialogError(state), nil, ""
	}
	return state, nil, ""
}

func updateTelemetryReview(state State, msg tea.KeyMsg) (State, *Action, string) {
	switch msg.String() {
	case "esc":
		return Close(state), nil, ""
	case "shift+tab", "left", "h":
		state.TelemetryStep = 1
		return clearDialogError(state), nil, ""
	case "j", "down":
		if state.ChoiceCursor < 2 {
			state.ChoiceCursor++
		}
		return clearDialogError(state), nil, ""
	case "k", "up":
		if state.ChoiceCursor > 0 {
			state.ChoiceCursor--
		}
		return clearDialogError(state), nil, ""
	case "enter":
		switch state.ChoiceCursor {
		case 0:
			return Close(state), &Action{
				Kind:           "patch_telemetry_settings",
				UsageTelemetry: state.TelemetryUsage,
				ErrorReporting: state.TelemetryErrors,
			}, ""
		case 1:
			return Close(state), &Action{
				Kind:             "patch_telemetry_settings",
				UsageTelemetry:   state.TelemetryUsage,
				ErrorReporting:   state.TelemetryErrors,
				RestartAfterSave: true,
			}, ""
		default:
			return Close(state), nil, ""
		}
	}
	return state, nil, ""
}
