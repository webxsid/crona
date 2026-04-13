package dialogs

import (
	"fmt"
	"strings"

	sharedtypes "crona/shared/types"
	tea "github.com/charmbracelet/bubbletea"

	uistate "crona/tui/internal/tui/state"
)

type menuChoice struct {
	Key    string
	Label  string
	Value  string
	Detail string
}

func OpenViewJump(state State, availableViews []uistate.View) State {
	state = Close(state)
	state.Kind = "view_jump"
	state.ViewTitle = "Jump To View"
	state.ViewName = "Press a mnemonic key or use j/k then enter"
	allChoices := []menuChoice{
		{Key: "a", Label: "Away", Value: "away", Detail: "Protected-mode shell when away or rest mode is active."},
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
		{Key: "n", Label: "Session", Value: "session_active", Detail: "Active session view while a timer is running."},
	}
	allowed := make(map[string]struct{}, len(availableViews))
	for _, view := range availableViews {
		allowed[string(view)] = struct{}{}
	}
	choices := make([]menuChoice, 0, len(allChoices))
	for _, choice := range allChoices {
		if _, ok := allowed[choice.Value]; ok {
			choices = append(choices, choice)
		}
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

func OpenStashConflict(state State, conflict sharedtypes.StashConflict) State {
	if len(conflict.Stashes) <= 1 {
		if len(conflict.Stashes) == 0 {
			return Close(state)
		}
		return openSingleStashConflict(state, conflict, 0)
	}
	state = Close(state)
	state.Kind = "stash_conflict_pick"
	state.ViewTitle = "Existing Stash Found"
	state.ViewName = fmt.Sprintf("Issue #%d already has %d stashes", conflict.IssueID, len(conflict.Stashes))
	state.ViewMeta = "Choose a stash to resume, or inspect it before continuing fresh."
	state.ChoiceItems = make([]string, 0, len(conflict.Stashes))
	state.ChoiceValues = make([]string, 0, len(conflict.Stashes))
	state.ChoiceDetails = make([]string, 0, len(conflict.Stashes))
	for _, stash := range conflict.Stashes {
		state.ChoiceItems = append(state.ChoiceItems, stashConflictLabel(stash))
		state.ChoiceValues = append(state.ChoiceValues, stash.ID)
		state.ChoiceDetails = append(state.ChoiceDetails, stashConflictDetail(stash))
	}
	state.ChoiceCursor = 0
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
	case "c":
		switch state.ViewName {
		case "Reports directory":
			return Close(state), &Action{Kind: "open_export_reports_dir_dialog"}, ""
		case "ICS export directory":
			return Close(state), &Action{Kind: "open_export_ics_dir_dialog"}, ""
		}
	case "r":
		switch state.ViewName {
		case "Reports directory":
			return Close(state), &Action{Kind: "reset_export_reports_dir"}, ""
		case "ICS export directory":
			return Close(state), &Action{Kind: "reset_export_ics_dir"}, ""
		}
	default:
		return clearDialogError(state), nil, ""
	}
	return clearDialogError(state), nil, ""
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

func updateStashConflictPick(state State, msg tea.KeyMsg) (State, *Action, string) {
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
		if state.ChoiceCursor < 0 || state.ChoiceCursor >= len(state.ChoiceValues) {
			return clearDialogError(state), nil, ""
		}
		return openSingleStashConflictByID(state, strings.TrimSpace(state.ChoiceValues[state.ChoiceCursor])), nil, ""
	default:
		return clearDialogError(state), nil, ""
	}
}

func updateStashConflict(state State, msg tea.KeyMsg) (State, *Action, string) {
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
	case "r":
		if strings.TrimSpace(state.DeleteID) == "" {
			return clearDialogError(state), nil, "Stash is unavailable"
		}
		return Close(state), &Action{Kind: "apply_stash", ID: strings.TrimSpace(state.DeleteID)}, ""
	case "c":
		return Close(state), &Action{Kind: "continue_focus_fresh", RepoID: state.RepoID, StreamID: state.StreamID, IssueID: state.IssueID}, ""
	case "enter":
		if state.ChoiceCursor < 0 || state.ChoiceCursor >= len(state.ChoiceValues) {
			if strings.TrimSpace(state.DeleteID) == "" {
				return clearDialogError(state), nil, "Stash is unavailable"
			}
			return Close(state), &Action{Kind: "apply_stash", ID: strings.TrimSpace(state.DeleteID)}, ""
		}
		value := strings.TrimSpace(state.ChoiceValues[state.ChoiceCursor])
		switch value {
		case "resume":
			return Close(state), &Action{Kind: "apply_stash", ID: strings.TrimSpace(state.DeleteID)}, ""
		case "continue":
			return Close(state), &Action{Kind: "continue_focus_fresh", RepoID: state.RepoID, StreamID: state.StreamID, IssueID: state.IssueID}, ""
		default:
			return clearDialogError(state), nil, ""
		}
	default:
		return clearDialogError(state), nil, ""
	}
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

func openSingleStashConflict(state State, conflict sharedtypes.StashConflict, idx int) State {
	if idx < 0 || idx >= len(conflict.Stashes) {
		return Close(state)
	}
	return openSingleStashConflictByID(withStashConflictItems(state, conflict), conflict.Stashes[idx].ID)
}

func withStashConflictItems(state State, conflict sharedtypes.StashConflict) State {
	state = Close(state)
	state.ViewTitle = "Existing Stash Found"
	state.ViewName = fmt.Sprintf("Issue #%d already has a stashed session", conflict.IssueID)
	if len(conflict.Stashes) > 1 {
		state.ViewName = fmt.Sprintf("Issue #%d has %d stashes", conflict.IssueID, len(conflict.Stashes))
	}
	state.ChoiceItems = make([]string, 0, len(conflict.Stashes))
	state.ChoiceValues = make([]string, 0, len(conflict.Stashes))
	state.ChoiceDetails = make([]string, 0, len(conflict.Stashes))
	for _, stash := range conflict.Stashes {
		state.ChoiceItems = append(state.ChoiceItems, stashConflictLabel(stash))
		state.ChoiceValues = append(state.ChoiceValues, stash.ID)
		state.ChoiceDetails = append(state.ChoiceDetails, stashConflictDetail(stash))
	}
	return state
}

func openSingleStashConflictByID(state State, stashID string) State {
	state.Kind = "stash_conflict"
	state.DeleteID = stashID
	selectedIdx := 0
	for i, value := range state.ChoiceValues {
		if strings.TrimSpace(value) == strings.TrimSpace(stashID) {
			selectedIdx = i
			break
		}
	}
	state.ViewMeta = ""
	state.ViewBody = ""
	if selectedIdx >= 0 && selectedIdx < len(state.ChoiceItems) {
		state.ViewMeta = state.ChoiceItems[selectedIdx]
	}
	if selectedIdx >= 0 && selectedIdx < len(state.ChoiceDetails) {
		state.ViewBody = state.ChoiceDetails[selectedIdx]
	}
	state.ChoiceItems = []string{"[r] Resume stash", "[c] Continue fresh"}
	state.ChoiceValues = []string{"resume", "continue"}
	state.ChoiceDetails = []string{
		"Apply this stash and continue the paused session instead of starting a new one.",
		"Start a fresh focus session on this issue and keep the stash for later.",
	}
	state.ChoiceCursor = 0
	return state
}

func stashConflictLabel(stash sharedtypes.Stash) string {
	if stash.Note != nil && strings.TrimSpace(*stash.Note) != "" {
		return strings.TrimSpace(*stash.Note)
	}
	return "Stash from " + strings.TrimSpace(stash.CreatedAt)
}

func stashConflictDetail(stash sharedtypes.Stash) string {
	parts := []string{}
	if strings.TrimSpace(stash.CreatedAt) != "" {
		parts = append(parts, "Created "+strings.TrimSpace(stash.CreatedAt))
	}
	if stash.PausedSegmentType != nil {
		parts = append(parts, "Segment "+string(*stash.PausedSegmentType))
	}
	if stash.ElapsedSeconds != nil && *stash.ElapsedSeconds > 0 {
		parts = append(parts, fmt.Sprintf("Elapsed %dm", *stash.ElapsedSeconds/60))
	}
	if len(parts) == 0 {
		return "Resume this stash or continue with a fresh session."
	}
	return strings.Join(parts, "  ·  ")
}
