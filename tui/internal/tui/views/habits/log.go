package habits

import (
	viewsummary "crona.work/tui/internal/tui/views/summary"
	"fmt"
	"strings"

	sharedtypes "crona.work/shared/types"
	"crona.work/tui/internal/api"
	helperpkg "crona.work/tui/internal/tui/helpers"
	viewchrome "crona.work/tui/internal/tui/views/chrome"
	viewhelpers "crona.work/tui/internal/tui/views/helpers"
	types "crona.work/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

const maxHabitNameWidth = 24

// renderLog lays the habit history out as a strip over a day-grouped list and a detail pane.
func renderLog(theme types.Theme, state types.ContentState) string {
	width, height := state.Width, state.Height
	indices := viewhelpers.FilteredStrings(habitHistoryItems(state), state.Filters["habit_history"])
	entries := make([]api.HabitCompletion, 0, len(indices))
	for _, index := range indices {
		entries = append(entries, state.HabitHistory[index])
	}
	cursor := min(max(0, state.Cursors["habit_history"]), max(0, len(entries)-1))
	strip := viewchrome.StripRows(habitLogStrip(state.HabitHistory), width)
	paneHeight := max(1, height-len(strip)-1)

	listWidth, detailWidth := viewchrome.LogPaneWidths(width)
	list := viewchrome.RenderSection(theme, viewchrome.Section{
		Title: "Habits", Meta: fmt.Sprint(len(entries)), Focused: true,
		Width: listWidth, Height: paneHeight,
		Lines: habitLogList(theme, entries, cursor, listWidth-1, paneHeight-1),
	})
	panes := list
	if detailWidth > 0 {
		var selected *api.HabitCompletion
		if len(entries) > 0 {
			selected = &entries[cursor]
		}
		title := "Detail"
		if selected != nil {
			title = habitHistoryHabitLabel(*selected)
		}
		detail := viewchrome.RenderSection(theme, viewchrome.Section{
			Title: title, Width: detailWidth, Height: paneHeight,
			Lines: HabitLogDetailLines(theme, state, selected, detailWidth-1),
		})
		panes = viewchrome.JoinColumns(theme, paneHeight, list, detail)
	}
	return strings.Join(append(strip, "", panes), "\n")
}

func habitLogStrip(history []api.HabitCompletion) [][2]string {
	done, failed, minutes := 0, 0, 0
	counts := map[string]int{}
	order := []string{}
	for _, entry := range history {
		switch entry.Status {
		case sharedtypes.HabitCompletionStatusCompleted:
			done++
		case sharedtypes.HabitCompletionStatusFailed:
			failed++
		}
		if entry.DurationMinutes != nil {
			minutes += *entry.DurationMinutes
		}
		name := habitLogName(entry)
		if counts[name] == 0 {
			order = append(order, name)
		}
		counts[name]++
	}
	most := "–"
	best := 0
	for _, name := range order {
		if counts[name] > best {
			most, best = fmt.Sprintf("%s %d", name, counts[name]), counts[name]
		}
	}
	return [][2]string{
		{fmt.Sprintf("Done %d  failed %d", done, failed), "Most logged  " + most},
		{"total " + helperpkg.FormatCompactDurationMinutes(minutes), ""},
	}
}

func habitLogList(
	theme types.Theme,
	entries []api.HabitCompletion,
	cursor, width, height int,
) []string {
	if len(entries) == 0 {
		return []string{theme.StyleDim.Render("No habit history recorded")}
	}
	nameWidth := 0
	for _, entry := range entries {
		nameWidth = max(nameWidth, min(maxHabitNameWidth, ansi.StringWidth(habitLogName(entry))))
	}
	rows := make([]viewchrome.DayRow, 0, len(entries))
	for _, entry := range entries {
		name := ansi.Truncate(habitLogName(entry), nameWidth, "…")
		duration := "–"
		if entry.DurationMinutes != nil {
			duration = helperpkg.FormatCompactDurationMinutes(*entry.DurationMinutes)
		}
		line := habitLogMark(
			theme,
			entry,
		) + " " + name + strings.Repeat(
			" ",
			nameWidth-ansi.StringWidth(name),
		) +
			" " + lipgloss.NewStyle().Foreground(theme.ColorBlue).Render(fmt.Sprintf("%6s", duration)) +
			"  " + theme.StyleDim.Render(
			habitLogContext(entry),
		)
		rows = append(rows, viewchrome.DayRow{Date: entry.Date, Line: line})
	}
	summary := func(date string) string {
		done, failed := 0, 0
		for _, entry := range entries {
			if entry.Date != date {
				continue
			}
			switch entry.Status {
			case sharedtypes.HabitCompletionStatusCompleted:
				done++
			case sharedtypes.HabitCompletionStatusFailed:
				failed++
			}
		}
		parts := []string{}
		if done > 0 {
			parts = append(parts, fmt.Sprintf("%d done", done))
		}
		if failed > 0 {
			parts = append(parts, fmt.Sprintf("%d failed", failed))
		}
		return strings.Join(parts, " · ")
	}
	return viewchrome.DateColumnList(theme, rows, cursor, width, height, summary)
}

// HabitLogDetailLines describes one completion from the list data alone.
func HabitLogDetailLines(
	theme types.Theme,
	state types.ContentState,
	entry *api.HabitCompletion,
	width int,
) []string {
	if entry == nil {
		return []string{theme.StyleDim.Render("Nothing selected")}
	}
	duration := "–"
	if entry.DurationMinutes != nil {
		duration = helperpkg.FormatCompactDurationMinutes(*entry.DurationMinutes)
	}
	lines := habitLogArcLines(theme, state, *entry, duration, width)
	if entry.Notes != nil && strings.TrimSpace(*entry.Notes) != "" {
		lines = append(lines, "", theme.StyleDim.Render("Note"))
		lines = append(
			lines,
			strings.Split(
				ansi.Wordwrap(strings.TrimSpace(*entry.Notes), max(1, width-1), ""),
				"\n",
			)...)
	}
	return lines
}

// habitLogArcLines leads the details with the entry's time against its target as an arc,
// beside its status, date, context and kind.
func habitLogArcLines(
	theme types.Theme,
	state types.ContentState,
	entry api.HabitCompletion,
	duration string,
	width int,
) []string {
	color := theme.ColorCyan
	statusColor := theme.ColorDim
	switch entry.Status {
	case sharedtypes.HabitCompletionStatusCompleted:
		color, statusColor = theme.ColorGreen, theme.ColorGreen
	case sharedtypes.HabitCompletionStatusFailed:
		color, statusColor = theme.ColorRed, theme.ColorRed
	}
	arc := viewsummary.Arc{Label: "Time", Value: duration, Color: color}
	if entry.SnapshotTarget != nil && *entry.SnapshotTarget > 0 {
		arc.Label = "of " + helperpkg.FormatCompactDurationMinutes(*entry.SnapshotTarget)
		if entry.DurationMinutes != nil {
			arc.Fraction = float64(*entry.DurationMinutes) / float64(*entry.SnapshotTarget)
		}
	} else if entry.Status == sharedtypes.HabitCompletionStatusCompleted {
		arc.Fraction = 1
	}
	facts := []string{
		lipgloss.NewStyle().Foreground(statusColor).Render("● "+habitHistoryStatusLabel(entry)) +
			theme.StyleDim.Render(" · ") + helperpkg.FormatDisplayDate(entry.Date, state.Settings),
		theme.StyleNormal.Render(habitLogContext(entry)),
	}
	if kind := strings.TrimSpace(string(entry.Kind)); kind != "" {
		facts = append(facts, theme.StyleDim.Render(kind))
	}
	arcWidth := viewsummary.ArcsWidth(1)
	return viewsummary.BesideArcs(viewsummary.Arcs(theme, []viewsummary.Arc{arc}, arcWidth), facts, arcWidth, width)
}

func habitLogMark(theme types.Theme, entry api.HabitCompletion) string {
	switch entry.Status {
	case sharedtypes.HabitCompletionStatusCompleted:
		return lipgloss.NewStyle().Foreground(theme.ColorGreen).Render("✓")
	case sharedtypes.HabitCompletionStatusFailed:
		return lipgloss.NewStyle().Foreground(theme.ColorRed).Render("✗")
	default:
		return theme.StyleDim.Render("·")
	}
}

func habitLogName(entry api.HabitCompletion) string {
	if name := strings.TrimSpace(entry.HabitName); name != "" {
		return name
	}
	if entry.SnapshotName != nil && strings.TrimSpace(*entry.SnapshotName) != "" {
		return strings.TrimSpace(*entry.SnapshotName)
	}
	return habitHistoryHabitLabel(entry)
}

func habitLogContext(entry api.HabitCompletion) string {
	return strings.Replace(habitHistoryContextLabel(entry), " > ", " › ", 1)
}
