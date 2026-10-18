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

	detailWidth := 0
	switch {
	case width >= 160:
		detailWidth = max(44, (width-6)*2/5)
	case width >= 100:
		detailWidth = max(40, (width-3)*2/5)
	}
	listWidth := width
	if detailWidth > 0 {
		listWidth = width - detailWidth - 3
	}
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

func habitLogList(theme types.Theme, entries []api.HabitCompletion, cursor, width, height int) []string {
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
		line := habitLogMark(theme, entry) + " " + name + strings.Repeat(" ", nameWidth-ansi.StringWidth(name)) +
			fmt.Sprintf(" %6s  ", duration) + theme.StyleDim.Render(habitLogContext(entry))
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
	return viewchrome.DayList(theme, rows, cursor, width, height, summary)
}

// HabitLogDetailLines describes one completion from the list data alone.
func HabitLogDetailLines(theme types.Theme, state types.ContentState, entry *api.HabitCompletion, width int) []string {
	if entry == nil {
		return []string{theme.StyleDim.Render("Nothing selected")}
	}
	duration := "–"
	if entry.DurationMinutes != nil {
		duration = helperpkg.FormatCompactDurationMinutes(*entry.DurationMinutes)
	}
	fields := [][2]string{
		{"Date", helperpkg.FormatDisplayDate(entry.Date, state.Settings)},
		{"Status", habitHistoryStatusLabel(*entry)},
		{"Duration", duration},
		{"Context", habitLogContext(*entry)},
		{"Kind", orDash(string(entry.Kind))},
	}
	if entry.SnapshotTarget != nil {
		fields = append(fields, [2]string{"Target", helperpkg.FormatCompactDurationMinutes(*entry.SnapshotTarget)})
	}
	lines := make([]string, 0, len(fields)+4)
	for _, field := range fields {
		lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("%-9s", field[0]))+" "+ansi.Truncate(field[1], max(1, width-10), "…"))
	}
	if entry.Notes != nil && strings.TrimSpace(*entry.Notes) != "" {
		lines = append(lines, "", theme.StyleDim.Render("Note"))
		lines = append(lines, strings.Split(ansi.Wordwrap(strings.TrimSpace(*entry.Notes), max(1, width-1), ""), "\n")...)
	}
	return lines
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

func orDash(value string) string {
	if strings.TrimSpace(value) == "" {
		return "–"
	}
	return value
}
