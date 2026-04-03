package dialogs

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func OpenViewEntity(state State, title string, name string, meta string, body string) State {
	state = Close(state)
	state.Kind = "view_entity"
	state.ViewTitle = title
	state.ViewName = name
	state.ViewMeta = meta
	state.ViewBody = body
	return state
}

func OpenSupportBundleResult(state State, name, meta, body, path string) State {
	state = Close(state)
	state.Kind = "support_bundle_result"
	state.ViewTitle = "Support Bundle Ready"
	state.ViewName = name
	state.ViewMeta = meta
	state.ViewBody = body
	state.SupportBundlePath = path
	return state
}

func updateViewEntity(state State, msg tea.KeyMsg) (State, *Action, string) {
	switch msg.String() {
	case "esc", "enter", "q":
		return Close(state), nil, ""
	default:
		return clearDialogError(state), nil, ""
	}
}

func updateSupportBundleResult(state State, msg tea.KeyMsg) (State, *Action, string) {
	switch msg.String() {
	case "esc", "enter", "q":
		return Close(state), nil, ""
	case "o":
		if strings.TrimSpace(state.SupportBundlePath) == "" {
			return state, nil, "Bundle path is unavailable"
		}
		return clearDialogError(state), &Action{Kind: "open_support_bundle_folder", Path: state.SupportBundlePath}, ""
	case "c":
		if strings.TrimSpace(state.SupportBundlePath) == "" {
			return state, nil, "Bundle path is unavailable"
		}
		return clearDialogError(state), &Action{Kind: "copy_support_bundle_path", Path: state.SupportBundlePath}, ""
	case "g":
		return clearDialogError(state), &Action{Kind: "open_support_issue"}, ""
	default:
		return clearDialogError(state), nil, ""
	}
}
