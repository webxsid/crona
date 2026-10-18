package habits

import (
	"fmt"
	"strings"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	helperpkg "crona/tui/internal/tui/helpers"
	types "crona/tui/internal/tui/views/types"
)

func renderHistoryView(theme types.Theme, state types.ContentState) string {
	return renderLog(theme, state)
}

func habitHistoryItems(state types.ContentState) []string {
	items := make([]string, 0, len(state.HabitHistory))
	for _, entry := range state.HabitHistory {
		items = append(items, habitHistoryItemSummary(entry))
	}
	return items
}

func habitHistoryItemSummary(entry api.HabitCompletion) string {
	parts := []string{
		entry.Date,
		habitHistoryHabitLabel(entry),
		habitHistoryContextLabel(entry),
		habitHistoryStatusLabel(entry),
	}
	if entry.DurationMinutes != nil {
		parts = append(parts, helperpkg.FormatCompactDurationMinutes(*entry.DurationMinutes))
	}
	if entry.Notes != nil && strings.TrimSpace(*entry.Notes) != "" {
		parts = append(parts, strings.TrimSpace(*entry.Notes))
	} else if entry.SnapshotName != nil && strings.TrimSpace(*entry.SnapshotName) != "" {
		parts = append(parts, strings.TrimSpace(*entry.SnapshotName))
	}
	if entry.SnapshotDesc != nil && strings.TrimSpace(*entry.SnapshotDesc) != "" {
		parts = append(parts, strings.TrimSpace(*entry.SnapshotDesc))
	}
	return strings.Join(parts, "  ")
}

func habitHistoryHabitLabel(entry api.HabitCompletion) string {
	name := strings.TrimSpace(entry.HabitName)
	if name != "" {
		return name
	}
	return fmt.Sprintf("Habit %d", entry.HabitID)
}

func habitHistoryContextLabel(entry api.HabitCompletion) string {
	repo := strings.TrimSpace(entry.RepoName)
	stream := strings.TrimSpace(entry.StreamName)
	switch {
	case repo != "" && stream != "":
		return repo + " > " + stream
	case repo != "":
		return repo
	case stream != "":
		return stream
	default:
		return "-"
	}
}

func habitHistoryStatusLabel(entry api.HabitCompletion) string {
	switch entry.Status {
	case sharedtypes.HabitCompletionStatusCompleted:
		return "completed"
	case sharedtypes.HabitCompletionStatusFailed:
		return "failed"
	default:
		return string(entry.Status)
	}
}
