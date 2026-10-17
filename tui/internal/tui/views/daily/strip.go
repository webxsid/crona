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

const dayStripBarWidth = 10

// renderDayStrip is the htop-style header above the Daily lists: day totals with
// inline bars and the date, plus a row of check-in signals and streaks when present.
func renderDayStrip(theme types.Theme, state types.ContentState, width int) []string {
	if width < 1 {
		return nil
	}
	rows := []string{dayTotalsRow(theme, state, width)}
	if row := daySignalsRow(theme, state); row != "" {
		rows = append(rows, padCells(" "+row, width))
	}
	return rows
}

func dayTotalsRow(theme types.Theme, state types.ContentState, width int) string {
	resolved, total, estimate := 0, len(state.DailyIssues), 0
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

	label := func(name string, done, of int) string {
		return theme.StyleHeader.Render(name) + " " + theme.StyleNormal.Render(fmt.Sprintf("%d/%d", done, of))
	}
	tasks := label("Tasks", resolved, total)
	progress := theme.StyleDim.Render(issueProgressText(worked, estimate, true))
	habits := label("Habits", habitsDone, len(state.DueHabits))
	withBars := tasks + " " + renderIssueStatusBar(theme, statusCounts, dayStripBarWidth) + " " + progress +
		"   " + habits + " " + renderHabitBar(theme, habitsDone, habitsFailed, len(state.DueHabits), dayStripBarWidth)
	plain := tasks + " " + progress + "   " + habits

	rawDate := dailyDate(state)
	date := helperpkg.FormatDisplayDate(rawDate, state.Settings)
	dated := date
	if parsed, ok := viewcalendar.ParseDate(rawDate); ok && !displayPatternIncludesWeek(state.Settings) {
		dated += fmt.Sprintf(" · W%02d", viewcalendar.WeekNumber(parsed, state.WeekStart))
	}
	navHint := "  " + theme.StyleDim.Render(",/. g")

	// Richest combination that fits wins: bars go before the date's week and key hint.
	for _, candidate := range [][2]string{
		{withBars, dated + navHint},
		{withBars, dated},
		{plain, dated + navHint},
		{plain, dated},
		{plain, date},
	} {
		left, right := " "+candidate[0], theme.StyleHeader.Render(candidate[1])
		if gap := width - ansi.StringWidth(left) - ansi.StringWidth(right); gap >= 2 {
			return left + strings.Repeat(" ", gap) + right
		}
	}
	return padCells(" "+plain, width)
}

// dailyDate is the day Daily shows: the loaded summary's date, else the dashboard date.
func dailyDate(state types.ContentState) string {
	if state.DailySummary != nil && state.DailySummary.Date != "" {
		return state.DailySummary.Date
	}
	return viewruntime.CurrentDashboardDate(state)
}

func daySignalsRow(theme types.Theme, state types.ContentState) string {
	field := func(name, value string) string {
		if value == "" {
			return ""
		}
		return theme.StyleHeader.Render(name) + " " + theme.StyleNormal.Render(strings.TrimPrefix(value, "today "))
	}
	groups := []string{}
	signals := joinNonEmpty("  ",
		field("Mood", signalMoodValue(state)),
		field("Energy", signalEnergyValue(state)),
		field("Sleep", signalSleepValue(state)),
	)
	if signals != "" {
		groups = append(groups, signals)
	}
	if streaks := state.DailyStreaks; streaks != nil {
		streak := func(name string, current, longest int) string {
			return theme.StyleHeader.Render(name) + " " + theme.StyleNormal.Render(fmt.Sprintf("%dd", current)) +
				theme.StyleDim.Render(fmt.Sprintf(" (best %dd)", longest))
		}
		groups = append(groups, streak("Focus", streaks.CurrentFocusDays, streaks.LongestFocusDays)+"  "+
			streak("Check-ins", streaks.CurrentCheckInDays, streaks.LongestCheckInDays))
	}
	return strings.Join(groups, theme.StyleDim.Render("   │   "))
}

func joinNonEmpty(sep string, parts ...string) string {
	kept := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			kept = append(kept, part)
		}
	}
	return strings.Join(kept, sep)
}

// padCells truncates or pads s to exactly width cells.
func padCells(s string, width int) string {
	s = viewhelpers.TruncateANSI(s, width)
	if w := ansi.StringWidth(s); w < width {
		s += strings.Repeat(" ", width-w)
	}
	return ansi.Truncate(s, width, "")
}
