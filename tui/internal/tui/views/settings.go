package views

import (
	"fmt"
	"strings"

	sharedtypes "crona/shared/types"
)

type settingsRow struct {
	Section string
	Label   string
	Value   string
}

type settingsVisibleRow struct {
	Header       bool
	Text         string
	SelectableAt int
}

func renderSettingsView(theme Theme, state ContentState) string {
	active := state.Pane == "settings"
	cur := state.Cursors["settings"]
	rows := SettingsRows(state.Settings)
	indices := filteredSettingIndices(state.Filters["settings"], state.Settings)
	total := len(indices)
	lines := []string{theme.StylePaneTitle.Render("Settings"), renderActionLine(theme, state.Width-6, ContextualActions(theme, ActionsState{View: state.View, Pane: state.Pane})), ""}
	if state.Settings == nil {
		lines = append(lines, theme.StyleDim.Render("Loading settings..."))
		return renderPaneBox(theme, active, state.Width, state.Height, stringsJoin(lines))
	}
	if total == 0 {
		lines = append(lines, theme.StyleDim.Render("No settings match the current filter"))
		return renderPaneBox(theme, active, state.Width, state.Height, stringsJoin(lines))
	}
	visibleRows, selectedVisibleIdx := groupedVisibleRows(indices, cur, func(idx int) string { return rows[idx].Section }, func(idx int) string {
		return fmt.Sprintf("%-24s %s", rows[idx].Label, rows[idx].Value)
	})
	inner := state.Height - 5
	if inner < 1 {
		inner = 1
	}
	start, end := listWindow(selectedVisibleIdx, len(visibleRows), inner)
	if start > 0 {
		lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("↑ %d more", start)))
	}
	for i := start; i < end; i++ {
		row := visibleRows[i]
		if row.Header {
			lines = append(lines, theme.StyleHeader.Render(strings.ToUpper(row.Text)))
			continue
		}
		switch {
		case row.SelectableAt == cur && active:
			lines = append(lines, theme.StyleCursor.Render("▶ "+row.Text))
		case row.SelectableAt == cur:
			lines = append(lines, theme.StyleSelected.Render("  "+row.Text))
		default:
			lines = append(lines, theme.StyleNormal.Render("  "+row.Text))
		}
	}
	if remaining := len(visibleRows) - end; remaining > 0 {
		lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("↓ %d more", remaining)))
	}
	return renderPaneBox(theme, active, state.Width, state.Height, stringsJoin(lines))
}

func SettingsRows(settings *sharedtypes.CoreSettings) []settingsRow {
	if settings == nil {
		return nil
	}
	return []settingsRow{
		{Section: "Focus Timer", Label: "Timer Mode", Value: string(settings.TimerMode)},
		{Section: "Focus Timer", Label: "Work Duration", Value: fmt.Sprintf("%d min", settings.WorkDurationMinutes)},
		{Section: "Breaks", Label: "Breaks Enabled", Value: onOff(settings.BreaksEnabled)},
		{Section: "Breaks", Label: "Short Break", Value: fmt.Sprintf("%d min", settings.ShortBreakMinutes)},
		{Section: "Breaks", Label: "Long Break", Value: fmt.Sprintf("%d min", settings.LongBreakMinutes)},
		{Section: "Breaks", Label: "Long Break Enabled", Value: onOff(settings.LongBreakEnabled)},
		{Section: "Breaks", Label: "Cycles Before Long Break", Value: fmt.Sprintf("%d", settings.CyclesBeforeLongBreak)},
		{Section: "Breaks", Label: "Auto Start Breaks", Value: onOff(settings.AutoStartBreaks)},
		{Section: "Breaks", Label: "Auto Start Work", Value: onOff(settings.AutoStartWork)},
		{Section: "Notifications", Label: "Boundary Notifications", Value: onOff(settings.BoundaryNotifications)},
		{Section: "Notifications", Label: "Boundary Sound", Value: onOff(settings.BoundarySound)},
		{Section: "Updates", Label: "Update Checks", Value: onOff(settings.UpdateChecksEnabled)},
		{Section: "Updates", Label: "Update Prompt", Value: onOff(settings.UpdatePromptEnabled)},
		{Section: "Sorting", Label: "Repo Sort", Value: repoSortLabel(settings.RepoSort)},
		{Section: "Sorting", Label: "Stream Sort", Value: streamSortLabel(settings.StreamSort)},
		{Section: "Sorting", Label: "Issue Sort", Value: issueSortLabel(settings.IssueSort)},
		{Section: "Sorting", Label: "Habit Sort", Value: habitSortLabel(settings.HabitSort)},
		{Section: "Recovery", Label: "Away Mode", Value: onOff(settings.AwayModeEnabled)},
		{Section: "Recovery", Label: "Rollback Window", Value: fmt.Sprintf("%d min", effectiveRollbackMinutes(settings.DailyPlanRollbackMins))},
		{Section: "Recovery", Label: "Rest & Streak Protection", Value: restProtectionLabel(settings)},
	}
}

func effectiveRollbackMinutes(value int) int {
	if value <= 0 {
		return 5
	}
	return value
}

func groupedVisibleRows(indices []int, selected int, sectionOf func(int) string, textOf func(int) string) ([]settingsVisibleRow, int) {
	rows := make([]settingsVisibleRow, 0, len(indices)+4)
	lastSection := ""
	selectedVisible := 0
	for i, idx := range indices {
		section := sectionOf(idx)
		if section != "" && section != lastSection {
			rows = append(rows, settingsVisibleRow{Header: true, Text: section, SelectableAt: -1})
			lastSection = section
		}
		rows = append(rows, settingsVisibleRow{Text: textOf(idx), SelectableAt: i})
		if i == selected {
			selectedVisible = len(rows) - 1
		}
	}
	return rows, selectedVisible
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

func streakKindsLabel(values []sharedtypes.StreakKind) string {
	if len(values) == 0 {
		values = sharedtypes.AvailableStreakKinds()
	}
	labels := make([]string, 0, len(values))
	for _, value := range values {
		switch value {
		case sharedtypes.StreakKindCheckInDays:
			labels = append(labels, "check-ins")
		default:
			labels = append(labels, "focus")
		}
	}
	return strings.Join(labels, ",")
}

func weekdaysLabel(values []int) string {
	if len(values) == 0 {
		return "-"
	}
	names := []string{"sun", "mon", "tue", "wed", "thu", "fri", "sat"}
	out := make([]string, 0, len(values))
	for _, value := range values {
		if value >= 0 && value < len(names) {
			out = append(out, names[value])
		}
	}
	if len(out) == 0 {
		return "-"
	}
	return strings.Join(out, ",")
}

func dateListLabel(values []string) string {
	if len(values) == 0 {
		return "-"
	}
	if len(values) <= 2 {
		return strings.Join(values, ",")
	}
	return fmt.Sprintf("%s +%d", values[0], len(values)-1)
}

func restProtectionLabel(settings *sharedtypes.CoreSettings) string {
	if settings == nil {
		return "-"
	}
	parts := []string{}
	if settings.AwayModeEnabled {
		parts = append(parts, "Away mode active")
	}
	if len(settings.FrozenStreakKinds) == 0 || len(settings.FrozenStreakKinds) == len(sharedtypes.AvailableStreakKinds()) {
		parts = append(parts, "All streaks")
	} else {
		kinds := streakKindsLabel(settings.FrozenStreakKinds)
		if kinds != "" && kinds != "-" {
			parts = append(parts, kinds)
		}
	}
	weekdays := weekdaysLabel(settings.RestWeekdays)
	if weekdays != "-" {
		parts = append(parts, weekdays)
	}
	dates := dateListLabel(settings.RestSpecificDates)
	if dates != "-" {
		label := dates
		if len(settings.RestSpecificDates) > 1 {
			label = fmt.Sprintf("%d dates", len(settings.RestSpecificDates))
		}
		parts = append(parts, label)
	}
	if len(parts) == 1 && parts[0] == "All streaks" {
		return "All streaks • No rest days"
	}
	return strings.Join(parts, " • ")
}
