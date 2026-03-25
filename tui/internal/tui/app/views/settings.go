package views

import (
	"fmt"

	sharedtypes "crona/shared/types"
)

func renderSettingsView(theme Theme, state ContentState) string {
	active := state.Pane == "settings"
	cur := state.Cursors["settings"]
	indices := filteredSettingIndices(state.Filters["settings"], state.Settings)
	total := len(indices)
	lines := []string{theme.StylePaneTitle.Render("Settings"), renderActionLine(theme, state.Width-6, ContextualActions(theme, ActionsState{View: state.View, Pane: state.Pane})), ""}
	if state.Settings == nil {
		lines = append(lines, theme.StyleDim.Render("Loading settings..."))
		return renderPaneBox(theme, active, state.Width, state.Height, stringsJoin(lines))
	}
	rows := []struct{ label, value string }{
		{"Timer Mode", string(state.Settings.TimerMode)},
		{"Breaks Enabled", onOff(state.Settings.BreaksEnabled)},
		{"Work Duration", fmt.Sprintf("%d min", state.Settings.WorkDurationMinutes)},
		{"Short Break", fmt.Sprintf("%d min", state.Settings.ShortBreakMinutes)},
		{"Long Break", fmt.Sprintf("%d min", state.Settings.LongBreakMinutes)},
		{"Long Break Enabled", onOff(state.Settings.LongBreakEnabled)},
		{"Cycles Before Long Break", fmt.Sprintf("%d", state.Settings.CyclesBeforeLongBreak)},
		{"Auto Start Breaks", onOff(state.Settings.AutoStartBreaks)},
		{"Auto Start Work", onOff(state.Settings.AutoStartWork)},
		{"Boundary Notifications", onOff(state.Settings.BoundaryNotifications)},
		{"Boundary Sound", onOff(state.Settings.BoundarySound)},
		{"Update Checks", onOff(state.Settings.UpdateChecksEnabled)},
		{"Update Prompt", onOff(state.Settings.UpdatePromptEnabled)},
		{"Repo Sort", repoSortLabel(state.Settings.RepoSort)},
		{"Stream Sort", streamSortLabel(state.Settings.StreamSort)},
		{"Issue Sort", issueSortLabel(state.Settings.IssueSort)},
		{"Habit Sort", habitSortLabel(state.Settings.HabitSort)},
	}
	if total == 0 {
		lines = append(lines, theme.StyleDim.Render("No settings match the current filter"))
		return renderPaneBox(theme, active, state.Width, state.Height, stringsJoin(lines))
	}
	inner := state.Height - 5
	if inner < 1 {
		inner = 1
	}
	start, end := listWindow(cur, total, inner)
	if start > 0 {
		lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("↑ %d more", start)))
	}
	for i := start; i < end; i++ {
		idx := indices[i]
		row := fmt.Sprintf("%-24s %s", rows[idx].label, rows[idx].value)
		if i == cur && active {
			lines = append(lines, theme.StyleCursor.Render("▶ "+row))
		} else if i == cur {
			lines = append(lines, theme.StyleSelected.Render("  "+row))
		} else {
			lines = append(lines, theme.StyleNormal.Render("  "+row))
		}
	}
	if remaining := total - end; remaining > 0 {
		lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("↓ %d more", remaining)))
	}
	return renderPaneBox(theme, active, state.Width, state.Height, stringsJoin(lines))
}

func repoSortLabel(value sharedtypes.RepoSort) string {
	switch value {
	case sharedtypes.RepoSortAlphabeticalAsc:
		return "A -> Z"
	case sharedtypes.RepoSortAlphabeticalDesc:
		return "Z -> A"
	case sharedtypes.RepoSortChronologicalDesc:
		return "Newest first"
	default:
		return "Oldest first"
	}
}

func streamSortLabel(value sharedtypes.StreamSort) string {
	switch value {
	case sharedtypes.StreamSortAlphabeticalAsc:
		return "A -> Z"
	case sharedtypes.StreamSortAlphabeticalDesc:
		return "Z -> A"
	case sharedtypes.StreamSortChronologicalDesc:
		return "Newest first"
	default:
		return "Oldest first"
	}
}

func issueSortLabel(value sharedtypes.IssueSort) string {
	switch value {
	case sharedtypes.IssueSortDueDateAsc:
		return "Due date earliest"
	case sharedtypes.IssueSortDueDateDesc:
		return "Due date latest"
	case sharedtypes.IssueSortAlphabeticalAsc:
		return "A -> Z"
	case sharedtypes.IssueSortAlphabeticalDesc:
		return "Z -> A"
	case sharedtypes.IssueSortChronologicalAsc:
		return "Oldest first"
	case sharedtypes.IssueSortChronologicalDesc:
		return "Newest first"
	default:
		return "Priority"
	}
}

func habitSortLabel(value sharedtypes.HabitSort) string {
	switch value {
	case sharedtypes.HabitSortTargetMinutesAsc:
		return "Target shortest"
	case sharedtypes.HabitSortTargetMinutesDesc:
		return "Target longest"
	case sharedtypes.HabitSortAlphabeticalAsc:
		return "A -> Z"
	case sharedtypes.HabitSortAlphabeticalDesc:
		return "Z -> A"
	case sharedtypes.HabitSortChronologicalAsc:
		return "Oldest first"
	case sharedtypes.HabitSortChronologicalDesc:
		return "Newest first"
	default:
		return "Schedule"
	}
}
