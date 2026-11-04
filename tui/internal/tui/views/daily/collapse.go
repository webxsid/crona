package daily

import (
	"strings"

	uistate "crona/tui/internal/tui/state"
	viewcalendar "crona/tui/internal/tui/views/calendar"
	viewchrome "crona/tui/internal/tui/views/chrome"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/x/ansi"
)

// The calendar shrinks to a week, then folds, before Tasks or Habits drop below these.
const (
	dailyMinTaskRows  = 8
	dailyMinHabitRows = 5
)

func dailyCollapsed(state types.ContentState, key string) bool {
	if state.Settings == nil {
		return false
	}
	return uistate.PaneCollapsed(state.Settings.CollapsedPanes, key)
}

// collapsedPane is a folded pane: its title row alone, with the key that opens it.
func collapsedPane(theme types.Theme, title, key string, width int) string {
	return viewchrome.RenderSection(theme, viewchrome.Section{
		Title:  title + " ▸",
		Meta:   viewchrome.KeyHint(theme, key, "expand"),
		Dimmed: true,
		Width:  width,
		Height: 1,
	})
}

// calendarLines is the shown date's month, or its week when there is not room for the
// month, centred in width.
func calendarLines(theme types.Theme, state types.ContentState, width int, mode viewcalendar.Mode) []string {
	selection := viewcalendar.Selection{
		AnchorDate:   dailyDate(state),
		SelectedDate: dailyDate(state),
		Mode:         mode,
	}
	if state.Settings != nil {
		selection.WeekStart = state.Settings.WeekStart
		selection.AwayDates = state.Settings.AwayDates
	}
	lines := viewcalendar.Render(theme, selection)
	widest := viewcalendar.MaxLineWidth(lines)
	if widest > width {
		return nil
	}
	pad := (width - widest) / 2
	for i, line := range lines {
		lines[i] = ansi.Truncate(strings.Repeat(" ", max(0, pad))+line, width, "")
	}
	return lines
}

// calendarPane picks the largest calendar fits allows: the month, then the week, else it
// folds to its title row.
func calendarPane(
	theme types.Theme,
	state types.ContentState,
	width int,
	fits func(calendarH int) bool,
) (string, int) {
	for _, mode := range []viewcalendar.Mode{viewcalendar.ModeMonth, viewcalendar.ModeWeek} {
		lines := calendarLines(theme, state, width-2, mode)
		height := len(lines) + 1
		if len(lines) > 0 && fits(height) {
			return dailySection(theme, "Calendar", viewchrome.KeyHint(theme, "alt+2", "fold"),
				false, width, height, lines), height
		}
	}
	return collapsedPane(theme, "Calendar", "alt+2", width), 1
}
