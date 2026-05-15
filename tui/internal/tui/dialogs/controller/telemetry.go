package controller

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
	case " ", "x", "enter":
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
	case " ", "x", "enter":
		state.TelemetryErrors = !state.TelemetryErrors
		return clearDialogError(state), nil, ""
	case "shift+tab", "left", "h":
		state.TelemetryStep = 0
		return clearDialogError(state), nil, ""
	case "tab", "right", "l":
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

func updateOnboarding(state State, msg tea.KeyMsg) (State, *Action, string) {
	switch state.TelemetryStep {
	case 0:
		switch msg.String() {
		case "tab", "right", "l":
			state.TelemetryStep = 1
			return clearDialogError(state), nil, ""
		case "shift+tab", "left", "h":
			state.TelemetryStep = 3
			return clearDialogError(state), nil, ""
		}
	case 1:
		switch msg.String() {
		case " ", "x":
			state.TelemetryUsage = !state.TelemetryUsage
			return clearDialogError(state), nil, ""
		case "shift+tab", "left", "h":
			state.TelemetryStep = 0
			return clearDialogError(state), nil, ""
		case "tab", "right", "l":
			state.TelemetryStep = 2
			return clearDialogError(state), nil, ""
		}
	case 2:
		switch msg.String() {
		case " ", "x":
			if state.TelemetryPrivacyCursor == 0 {
				state.TelemetryUsage = !state.TelemetryUsage
			} else {
				state.TelemetryErrors = !state.TelemetryErrors
			}
			return clearDialogError(state), nil, ""
		case "j", "down":
			state.TelemetryPrivacyCursor = (state.TelemetryPrivacyCursor + 1) % 2
			return clearDialogError(state), nil, ""
		case "shift+tab", "left", "h":
			state.TelemetryStep = 1
			return clearDialogError(state), nil, ""
		case "k", "up":
			if state.TelemetryPrivacyCursor == 0 {
				state.TelemetryPrivacyCursor = 1
			} else {
				state.TelemetryPrivacyCursor = 0
			}
			return clearDialogError(state), nil, ""
		case "tab", "right", "l":
			state.TelemetryStep = 3
			return clearDialogError(state), nil, ""
		}
	default:
		switch msg.String() {
		case "shift+tab", "left", "h":
			state.TelemetryStep = 2
			return clearDialogError(state), nil, ""
		case "j", "down":
			if state.TelemetryReviewCursor < 2 {
				state.TelemetryReviewCursor++
			}
			return clearDialogError(state), nil, ""
		case "k", "up":
			if state.TelemetryReviewCursor > 0 {
				state.TelemetryReviewCursor--
			}
			return clearDialogError(state), nil, ""
		case "tab", "right", "l":
			state.TelemetryStep = 0
			return clearDialogError(state), nil, ""
		case "enter":
			switch state.TelemetryReviewCursor {
			case 0:
				return Close(state), &Action{
					Kind:           "complete_onboarding",
					UsageTelemetry: state.TelemetryUsage,
					ErrorReporting: state.TelemetryErrors,
					OnboardingDone: true,
				}, ""
			case 1:
				return Close(state), &Action{
					Kind:             "complete_onboarding",
					UsageTelemetry:   state.TelemetryUsage,
					ErrorReporting:   state.TelemetryErrors,
					RestartAfterSave: true,
					OnboardingDone:   true,
				}, ""
			default:
				return Close(state), nil, ""
			}
		}
	}
	return state, nil, ""
}
