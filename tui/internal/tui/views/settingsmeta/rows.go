package settingsmeta

import (
	"fmt"
	"strings"

	sharedtypes "crona/shared/types"
	viewhelpers "crona/tui/internal/tui/views/helpers"
)

type Row struct {
	Section string
	Label   string
	Value   string
}

type VisibleRow struct {
	Header       bool
	Text         string
	SelectableAt int
	Danger       bool
}

func Rows(settings *sharedtypes.CoreSettings) []Row {
	if settings == nil {
		return nil
	}
	return []Row{
		{Section: "Focus Timer", Label: "Timer Mode", Value: string(settings.TimerMode)},
		{Section: "Focus Timer", Label: "Work Duration", Value: fmt.Sprintf("%d min", settings.WorkDurationMinutes)},
		{Section: "Breaks", Label: "Breaks", Value: enabledDisabled(settings.BreaksEnabled)},
		{Section: "Breaks", Label: "Short Break", Value: fmt.Sprintf("%d min", settings.ShortBreakMinutes)},
		{Section: "Breaks", Label: "Long Break", Value: fmt.Sprintf("%d min", settings.LongBreakMinutes)},
		{Section: "Breaks", Label: "Long Breaks", Value: enabledDisabled(settings.LongBreakEnabled)},
		{Section: "Breaks", Label: "Cycles Before Long Break", Value: fmt.Sprintf("%d", settings.CyclesBeforeLongBreak)},
		{Section: "Breaks", Label: "Auto-Start Breaks", Value: enabledDisabled(settings.AutoStartBreaks)},
		{Section: "Breaks", Label: "Auto-Start Work", Value: enabledDisabled(settings.AutoStartWork)},
		{Section: "Notifications", Label: "Boundary Notifications", Value: enabledDisabled(settings.BoundaryNotifications)},
		{Section: "Notifications", Label: "Boundary Sound", Value: enabledDisabled(settings.BoundarySound)},
		{Section: "Updates", Label: "Update Checks", Value: enabledDisabled(settings.UpdateChecksEnabled)},
		{Section: "Updates", Label: "Update Prompt", Value: enabledDisabled(settings.UpdatePromptEnabled)},
		{Section: "Updates", Label: "Update Channel", Value: UpdateChannelLabel(settings.UpdateChannel)},
		{Section: "Sorting", Label: "Repo Sort", Value: repoSortLabel(settings.RepoSort)},
		{Section: "Sorting", Label: "Stream Sort", Value: streamSortLabel(settings.StreamSort)},
		{Section: "Sorting", Label: "Issue Sort", Value: issueSortLabel(settings.IssueSort)},
		{Section: "Sorting", Label: "Habit Sort", Value: habitSortLabel(settings.HabitSort)},
		{Section: "Recovery", Label: "Away Mode", Value: enabledDisabled(settings.AwayModeEnabled)},
		{Section: "Recovery", Label: "Rollback Window", Value: fmt.Sprintf("%d min", effectiveRollbackMinutes(settings.DailyPlanRollbackMins))},
		{Section: "Recovery", Label: "Rest & Streak Protection", Value: restProtectionLabel(settings)},
		{Section: "Danger", Label: "Wipe Runtime Data", Value: "Destructive"},
		{Section: "Danger", Label: "Uninstall Crona", Value: "App + binaries"},
	}
}

func GroupedVisibleRows(indices []int, selected int, sectionOf func(int) string, textOf func(int) string) ([]VisibleRow, int) {
	rows := make([]VisibleRow, 0, len(indices)+4)
	lastSection := ""
	selectedVisible := 0
	for i, idx := range indices {
		section := sectionOf(idx)
		if section != "" && section != lastSection {
			rows = append(rows, VisibleRow{Header: true, Text: section, SelectableAt: -1, Danger: section == "Danger"})
			lastSection = section
		}
		rows = append(rows, VisibleRow{Text: textOf(idx), SelectableAt: i, Danger: section == "Danger"})
		if i == selected {
			selectedVisible = len(rows) - 1
		}
	}
	return rows, selectedVisible
}

func ItemLabels(settings *sharedtypes.CoreSettings) []string {
	rows := Rows(settings)
	labels := make([]string, 0, len(rows))
	for _, row := range rows {
		labels = append(labels, row.Label)
	}
	return labels
}

func FilteredIndices(filter string, settings *sharedtypes.CoreSettings) []int {
	if settings == nil {
		return nil
	}
	return viewhelpers.FilteredStrings(ItemLabels(settings), filter)
}

func UpdateChannelLabel(value sharedtypes.UpdateChannel) string {
	switch sharedtypes.NormalizeUpdateChannel(value) {
	case sharedtypes.UpdateChannelBeta:
		return "Beta"
	default:
		return "Stable"
	}
}

func enabledDisabled(v bool) string {
	if v {
		return "Enabled"
	}
	return "Disabled"
}

func effectiveRollbackMinutes(value int) int {
	if value <= 0 {
		return 5
	}
	return value
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
