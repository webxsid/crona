package dialogs

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type menuChoice struct {
	Key    string
	Label  string
	Value  string
	Detail string
}

func OpenViewJump(state State, protectedModeActive bool, hasActiveTimer bool) State {
	state = Close(state)
	state.Kind = "view_jump"
	state.ViewTitle = "Jump To View"
	state.ViewName = "Press a mnemonic key or use j/k then enter"
	choices := []menuChoice{
		{Key: "d", Label: "Daily", Value: "daily", Detail: "Daily dashboard with planned issues and habits."},
		{Key: "r", Label: "Rollup", Value: "rollup", Detail: "Range summaries and drill-down day details."},
		{Key: "w", Label: "Wellbeing", Value: "wellbeing", Detail: "Check-ins, burnout signals, and trends."},
		{Key: "p", Label: "Reports", Value: "reports", Detail: "Generated report and export browser."},
		{Key: "c", Label: "Config", Value: "config", Detail: "Export assets, templates, and renderer tools."},
		{Key: "i", Label: "Issues", Value: "default", Detail: "Primary workspace issue list."},
		{Key: "m", Label: "Meta", Value: "meta", Detail: "Repos, streams, issues, and habits by hierarchy."},
		{Key: "x", Label: "Scratchpads", Value: "scratchpads", Detail: "Filesystem-backed notes and scratchpads."},
		{Key: "o", Label: "Ops", Value: "ops", Detail: "Recent operation log."},
		{Key: "s", Label: "Settings", Value: "settings", Detail: "Core settings and danger actions."},
		{Key: "l", Label: "Alerts", Value: "alerts", Detail: "Notification, sound, and alert backend settings."},
		{Key: "u", Label: "Updates", Value: "updates", Detail: "Release notes, update checks, and install status."},
		{Key: "h", Label: "Support", Value: "support", Detail: "Bug reporting, bundles, and GitHub links."},
		{Key: "y", Label: "History", Value: "session_history", Detail: "Session history and past focus work."},
	}
	if protectedModeActive {
		choices = append([]menuChoice{{Key: "a", Label: "Away", Value: "away", Detail: "Protected-mode shell when away or rest mode is active."}}, choices...)
	}
	if hasActiveTimer {
		choices = append(choices, menuChoice{Key: "n", Label: "Session", Value: "session_active", Detail: "Active session view while a timer is running."})
	}
	state.ChoiceItems, state.ChoiceValues, state.ChoiceDetails = menuChoiceLists(choices)
	return state
}

func OpenBetaSupport(state State) State {
	state = Close(state)
	state.Kind = "beta_support"
	state.ViewTitle = "Beta Support"
	state.ViewName = "Quick reporting and diagnostics for beta testers"
	choices := []menuChoice{
		{Key: "o", Label: "Report Bug", Value: "open_support_issue", Detail: "Open the prefilled GitHub issue flow."},
		{Key: "d", Label: "Discussions", Value: "open_support_discussions", Detail: "Open GitHub Discussions for questions and ideas."},
		{Key: "r", Label: "Releases", Value: "open_support_releases", Detail: "Open the GitHub releases feed."},
		{Key: "g", Label: "Roadmap", Value: "open_support_roadmap", Detail: "Open the public roadmap document."},
		{Key: "c", Label: "Copy Diagnostics", Value: "copy_support_diagnostics", Detail: "Copy the lightweight diagnostics summary."},
		{Key: "b", Label: "Generate Bundle", Value: "generate_support_bundle", Detail: "Create a full redacted support bundle."},
	}
	state.ChoiceItems, state.ChoiceValues, state.ChoiceDetails = menuChoiceLists(choices)
	return state
}

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

func updateViewJump(state State, msg tea.KeyMsg) (State, *Action, string) {
	return updateChoiceMenu(state, msg, true)
}

func updateBetaSupport(state State, msg tea.KeyMsg) (State, *Action, string) {
	return updateChoiceMenu(state, msg, false)
}

func updateChoiceMenu(state State, msg tea.KeyMsg, jumpMenu bool) (State, *Action, string) {
	switch msg.String() {
	case "esc", "q":
		return Close(state), nil, ""
	case "j", "down":
		if state.ChoiceCursor < len(state.ChoiceItems)-1 {
			state.ChoiceCursor++
		}
		return clearDialogError(state), nil, ""
	case "k", "up":
		if state.ChoiceCursor > 0 {
			state.ChoiceCursor--
		}
		return clearDialogError(state), nil, ""
	case "enter":
		return resolveChoiceSelection(state, jumpMenu)
	default:
		for idx, item := range state.ChoiceItems {
			if msg.String() == menuChoiceKey(item) {
				state.ChoiceCursor = idx
				return resolveChoiceSelection(state, jumpMenu)
			}
		}
		return clearDialogError(state), nil, ""
	}
}

func resolveChoiceSelection(state State, jumpMenu bool) (State, *Action, string) {
	if state.ChoiceCursor < 0 || state.ChoiceCursor >= len(state.ChoiceValues) {
		return clearDialogError(state), nil, ""
	}
	value := strings.TrimSpace(state.ChoiceValues[state.ChoiceCursor])
	if value == "" {
		return clearDialogError(state), nil, ""
	}
	if jumpMenu {
		return Close(state), &Action{Kind: "jump_view", TargetView: value}, ""
	}
	return Close(state), &Action{Kind: value}, ""
}

func menuChoiceLists(choices []menuChoice) ([]string, []string, []string) {
	items := make([]string, 0, len(choices))
	values := make([]string, 0, len(choices))
	details := make([]string, 0, len(choices))
	for _, choice := range choices {
		items = append(items, "["+choice.Key+"] "+choice.Label)
		values = append(values, choice.Value)
		details = append(details, choice.Detail)
	}
	return items, values, details
}

func menuChoiceKey(label string) string {
	label = strings.TrimSpace(label)
	if !strings.HasPrefix(label, "[") {
		return ""
	}
	end := strings.Index(label, "]")
	if end <= 1 {
		return ""
	}
	return strings.TrimSpace(label[1:end])
}
