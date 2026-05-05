package habits

import (
	"fmt"
	"strings"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	helperpkg "crona/tui/internal/tui/helpers"
	viewchrome "crona/tui/internal/tui/views/chrome"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
)

func renderHistoryView(theme types.Theme, state types.ContentState) string {
	active := state.Pane == "habit_history"
	cur := state.Cursors["habit_history"]
	items := habitHistoryItems(state)
	indices := viewhelpers.FilteredStrings(items, state.Filters["habit_history"])
	total := len(indices)
	title := state.HabitHistoryTitle
	if strings.TrimSpace(title) == "" {
		title = "Habit History"
	}
	subtitle := state.HabitHistoryMeta
	if strings.TrimSpace(subtitle) == "" {
		subtitle = habitHistoryScopeSubtitle(state.Context)
	}
	lines := []string{
		theme.StylePaneTitle.Render(title),
		theme.StyleDim.Render(subtitle),
		viewchrome.RenderPaneActionLine(theme, state.Filters["habit_history"], state.Width-6, viewchrome.PaneActionsForState(theme, state, active)),
	}
	if total == 0 {
		lines = append(lines, theme.StyleDim.Render("No habit history recorded"))
		return viewchrome.RenderPaneBox(theme, active, state.Width, state.Height, viewhelpers.StringsJoin(lines))
	}
	dateW := 12
	habitW := 18
	contextW := 18
	statusW := 10
	durW := 8
	notesW := state.Width - dateW - habitW - contextW - statusW - durW - 18
	if notesW < 16 {
		notesW = 16
	}
	header := fmt.Sprintf("%-2s %-*s %-*s %-*s %-*s %-*s %s", "", dateW, "Date", habitW, "Habit", contextW, "Context", statusW, "Status", durW, "Duration", "Notes")
	lines = append(lines, theme.StyleDim.Render(viewhelpers.Truncate(header, state.Width-6)))
	inner := viewchrome.RemainingPaneHeight(state.Height, lines)
	start, end := viewchrome.ListWindow(cur, total, inner)
	if start > 0 {
		lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("↑ %d more", start)))
	}
	for i := start; i < end; i++ {
		entry := state.HabitHistory[indices[i]]
		date := helperpkg.FormatDisplayDate(entry.Date, state.Settings)
		habit := habitHistoryHabitLabel(entry)
		contextLabel := habitHistoryContextLabel(entry)
		status := habitHistoryStatusLabel(entry)
		duration := "-"
		if entry.DurationMinutes != nil {
			duration = helperpkg.FormatCompactDurationMinutes(*entry.DurationMinutes)
		}
		notes := habitHistoryDetail(entry, notesW)
		row := fmt.Sprintf("%-2s %-*s %-*s %-*s %-*s %-*s %s", "", dateW, date, habitW, habit, contextW, contextLabel, statusW, status, durW, duration, notes)
		lines = append(lines, viewchrome.RenderPaneRowStyled(theme, i, cur, active, row, habitHistoryStatusStyle(theme, entry), state.Width))
	}
	if remaining := total - end; remaining > 0 {
		lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("↓ %d more", remaining)))
	}
	return viewchrome.RenderPaneBox(theme, active, state.Width, state.Height, viewhelpers.StringsJoin(lines))
}

func habitHistoryItems(state types.ContentState) []string {
	items := make([]string, 0, len(state.HabitHistory))
	for _, entry := range state.HabitHistory {
		items = append(items, habitHistoryItemSummary(entry))
	}
	return items
}

func habitHistoryItemSummary(entry api.HabitCompletion) string {
	parts := []string{entry.Date, habitHistoryHabitLabel(entry), habitHistoryContextLabel(entry), habitHistoryStatusLabel(entry)}
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

func habitHistoryScopeSubtitle(ctx *api.ActiveContext) string {
	if ctx == nil {
		return "Recent habit activity across the workspace"
	}
	repo := ""
	if ctx.RepoName != nil {
		repo = strings.TrimSpace(*ctx.RepoName)
	}
	stream := ""
	if ctx.StreamName != nil {
		stream = strings.TrimSpace(*ctx.StreamName)
	}
	switch {
	case repo != "" && stream != "":
		return "Recent habit activity in " + repo + " > " + stream
	case repo != "":
		return "Recent habit activity in " + repo
	case stream != "":
		return "Recent habit activity in " + stream
	default:
		return "Recent habit activity across the workspace"
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

func habitHistoryStatusStyle(theme types.Theme, entry api.HabitCompletion) *lipgloss.Style {
	switch entry.Status {
	case sharedtypes.HabitCompletionStatusCompleted:
		s := lipgloss.NewStyle().Foreground(theme.ColorGreen)
		return &s
	case sharedtypes.HabitCompletionStatusFailed:
		s := lipgloss.NewStyle().Foreground(theme.ColorRed)
		return &s
	default:
		s := lipgloss.NewStyle().Foreground(theme.ColorDim)
		return &s
	}
}

func habitHistoryDetail(entry api.HabitCompletion, width int) string {
	bits := []string{}
	if entry.Notes != nil && strings.TrimSpace(*entry.Notes) != "" {
		bits = append(bits, strings.TrimSpace(*entry.Notes))
	}
	if len(bits) == 0 && entry.SnapshotName != nil && strings.TrimSpace(*entry.SnapshotName) != "" {
		bits = append(bits, strings.TrimSpace(*entry.SnapshotName))
	}
	if len(bits) == 0 && entry.SnapshotDesc != nil && strings.TrimSpace(*entry.SnapshotDesc) != "" {
		bits = append(bits, strings.TrimSpace(*entry.SnapshotDesc))
	}
	if len(bits) == 0 {
		bits = append(bits, "-")
	}
	return viewhelpers.Truncate(strings.Join(bits, "  "), width)
}
