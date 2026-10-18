package daily

import (
	"fmt"
	"strings"

	helperpkg "crona/tui/internal/tui/helpers"
	viewcalendar "crona/tui/internal/tui/views/calendar"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	viewruntime "crona/tui/internal/tui/views/runtime"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/x/ansi"
)

const (
	dayStripBarWidth = 10
	// dayStripColumnGap separates the meter, signal and streak columns.
	dayStripColumnGap = "     "
)

// renderDayStrip is the htop-style header above the Daily lists: stacked Tasks and Habits
// meters, then check-in signals and streaks as aligned columns. Columns that do not fit
// the width are dropped from the right, streaks first.
func renderDayStrip(theme types.Theme, state types.ContentState, width int) []string {
	if width < 1 {
		return nil
	}
	columns := [][]string{dayMeterColumn(theme, state, true)}
	if columnWidth(columns[0])+1 > width {
		columns[0] = dayMeterColumn(theme, state, false)
	}
	for _, extra := range [][]string{daySignalColumn(theme, state), dayStreakColumn(theme, state)} {
		if len(extra) == 0 {
			continue
		}
		candidate := append(append([][]string{}, columns...), extra)
		if stripWidth(candidate)+1 > width {
			break
		}
		columns = candidate
	}
	height := 0
	for _, column := range columns {
		height = max(height, len(column))
	}
	rows := make([]string, height)
	for i := range rows {
		cells := make([]string, 0, len(columns))
		for _, column := range columns {
			cell := ""
			if i < len(column) {
				cell = column[i]
			}
			cells = append(cells, padCells(cell, columnWidth(column)))
		}
		rows[i] = padCells(" "+strings.TrimRight(strings.Join(cells, dayStripColumnGap), " "), width)
	}
	return rows
}

// SubTabMeta is the right side of Daily's sub-tab line: the shown date, its week and the
// date-navigation keys.
func SubTabMeta(theme types.Theme, state types.ContentState) string {
	rawDate := dailyDate(state)
	dated := helperpkg.FormatDisplayDate(rawDate, state.Settings)
	if parsed, ok := viewcalendar.ParseDate(rawDate); ok && !displayPatternIncludesWeek(state.Settings) {
		dated += fmt.Sprintf(" · W%02d", viewcalendar.WeekNumber(parsed, state.WeekStart))
	}
	return theme.StyleHeader.Render(dated) + "  " + theme.StyleDim.Render(",/. g")
}

func dayMeterColumn(theme types.Theme, state types.ContentState, withBars bool) []string {
	resolved, estimate := 0, 0
	statusCounts := map[string]int{}
	for _, issue := range state.DailyIssues {
		statusCounts[string(issue.Status)]++
		if issue.EstimateMinutes != nil {
			estimate += *issue.EstimateMinutes
		}
		if issue.Status == "done" || issue.Status == "abandoned" {
			resolved++
		}
	}
	habitsDone, habitsFailed := 0, 0
	for _, habit := range state.DueHabits {
		switch habit.Status {
		case "completed":
			habitsDone++
		case "failed":
			habitsFailed++
		}
	}
	worked := 0
	if state.DailySummary != nil {
		worked = state.DailySummary.WorkedSeconds
	}
	values := []string{
		fmt.Sprintf("%d/%d", resolved, len(state.DailyIssues)),
		fmt.Sprintf("%d/%d", habitsDone, len(state.DueHabits)),
	}
	valueWidth := max(len(values[0]), len(values[1]))
	row := func(label, value, bar, tail string) string {
		line := theme.StyleHeader.Render(padCells(label, 6)) + "  " + theme.StyleNormal.Render(padCells(value, valueWidth))
		if bar != "" {
			line += "  " + bar
		}
		if tail != "" {
			line += "  " + tail
		}
		return line
	}
	tasksBar, habitsBar := "", ""
	if withBars {
		tasksBar = renderIssueStatusBar(theme, statusCounts, dayStripBarWidth)
		habitsBar = renderHabitBar(theme, habitsDone, habitsFailed, len(state.DueHabits), dayStripBarWidth)
	}
	return []string{
		row("Tasks", values[0], tasksBar, theme.StyleDim.Render(issueProgressText(worked, estimate, true))),
		row("Habits", values[1], habitsBar, ""),
	}
}

func daySignalColumn(theme types.Theme, state types.ContentState) []string {
	column := []string{}
	for _, signal := range []struct{ name, value string }{
		{"Mood", signalMoodValue(state)},
		{"Energy", signalEnergyValue(state)},
		{"Sleep", signalSleepValue(state)},
	} {
		if signal.value == "" {
			continue
		}
		column = append(column, theme.StyleHeader.Render(padCells(signal.name, 6))+"  "+
			theme.StyleNormal.Render(strings.TrimPrefix(signal.value, "today ")))
	}
	return column
}

func dayStreakColumn(theme types.Theme, state types.ContentState) []string {
	streaks := state.DailyStreaks
	if streaks == nil {
		return nil
	}
	current := []string{fmt.Sprintf("%dd", streaks.CurrentFocusDays), fmt.Sprintf("%dd", streaks.CurrentCheckInDays)}
	currentWidth := max(len(current[0]), len(current[1]))
	row := func(label, now string, best int) string {
		return theme.StyleHeader.Render(padCells(label, 9)) + "  " + theme.StyleNormal.Render(padCells(now, currentWidth)) +
			"  " + theme.StyleDim.Render(fmt.Sprintf("best %dd", best))
	}
	return []string{
		row("Focus", current[0], streaks.LongestFocusDays),
		row("Check-ins", current[1], streaks.LongestCheckInDays),
	}
}

func columnWidth(column []string) int {
	width := 0
	for _, cell := range column {
		width = max(width, ansi.StringWidth(cell))
	}
	return width
}

func stripWidth(columns [][]string) int {
	width := 0
	for i, column := range columns {
		if i > 0 {
			width += len(dayStripColumnGap)
		}
		width += columnWidth(column)
	}
	return width
}

// dailyDate is the day Daily shows: the loaded summary's date, else the dashboard date.
func dailyDate(state types.ContentState) string {
	if state.DailySummary != nil && state.DailySummary.Date != "" {
		return state.DailySummary.Date
	}
	return viewruntime.CurrentDashboardDate(state)
}

// padCells truncates or pads s to exactly width cells.
func padCells(s string, width int) string {
	s = viewhelpers.TruncateANSI(s, width)
	if w := ansi.StringWidth(s); w < width {
		s += strings.Repeat(" ", width-w)
	}
	return ansi.Truncate(s, width, "")
}
