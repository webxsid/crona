package review

import (
	"fmt"
	"slices"
	"time"

	"crona/tui/internal/api"
	viewchrome "crona/tui/internal/tui/views/chrome"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	types "crona/tui/internal/tui/views/types"

	sharedtypes "crona/shared/types"
)

const isoDate = "2006-01-02"

// WindowDates lists the window's ISO dates, oldest first.
func WindowDates(state types.ContentState) []string {
	end := state.ReviewWindowEnd
	if end == "" {
		end = state.DashboardDate
	}
	last, err := time.Parse(isoDate, end)
	if err != nil {
		return nil
	}
	days := state.ReviewWindowDays
	if days < 1 {
		days = 7
	}
	dates := make([]string, days)
	for i := range days {
		dates[i] = last.AddDate(0, 0, i-days+1).Format(isoDate)
	}
	return dates
}

func renderDays(theme types.Theme, state types.ContentState, width, height int, meta string) string {
	dates := WindowDates(state)
	metrics := map[string]api.DailyMetricsDay{}
	for _, day := range state.MetricsRange {
		metrics[day.Date] = day
	}
	plans := map[string]sharedtypes.DashboardWindowDay{}
	if state.DashboardWindow != nil {
		for _, day := range state.DashboardWindow.Days {
			plans[day.Date] = day
		}
	}
	meanMood, moods := 0.0, 0
	for _, date := range dates {
		if day, ok := metrics[date]; ok && day.CheckIn != nil {
			meanMood += float64(day.CheckIn.Mood)
			moods++
		}
	}
	if moods > 0 {
		meanMood /= float64(moods)
	}
	cur := slices.Index(dates, state.DashboardDate)
	start, end := viewchrome.ListWindow(max(0, cur), len(dates), max(1, height-1))
	lines := make([]string, 0, end-start)
	for i := start; i < end; i++ {
		line := dayRow(theme, dates[i], plans[dates[i]], metrics[dates[i]], meanMood)
		if i == cur {
			lines = append(lines, theme.StyleCursor.Render(viewchrome.SelectionCursor+" "+line))
		} else {
			lines = append(lines, "  "+line)
		}
	}
	return viewchrome.RenderSection(theme, viewchrome.Section{
		Title: "Days", Meta: meta, Focused: true, Width: width, Height: height, Lines: lines,
	})
}

func dayRow(theme types.Theme, date string, plan sharedtypes.DashboardWindowDay, metrics api.DailyMetricsDay, meanMood float64) string {
	label := date
	if t, err := time.Parse(isoDate, date); err == nil {
		label = t.Format("Mon 2")
	}
	mark := " "
	if metrics.CheckIn != nil {
		mood := float64(metrics.CheckIn.Mood)
		switch {
		case mood > meanMood+0.25:
			mark = "▲"
		case mood < meanMood-0.25:
			mark = "▼"
		default:
			mark = "▬"
		}
	}
	return fmt.Sprintf("%-6s  %s %d/%d  %4s %s",
		label,
		statusStyle(theme, string(plan.Status)).Render(statusMark(plan.Status)),
		plan.CompletedCount, plan.PlannedCount,
		viewhelpers.FormatCompactDurationSeconds(metrics.WorkedSeconds),
		mark,
	)
}

func statusMark(status sharedtypes.DashboardWindowDayStatus) string {
	switch status {
	case sharedtypes.DashboardWindowDayDone:
		return "✓"
	case sharedtypes.DashboardWindowDayMissed:
		return "✗"
	case sharedtypes.DashboardWindowDayCarryOver:
		return "↷"
	case sharedtypes.DashboardWindowDayPlanned:
		return "·"
	case sharedtypes.DashboardWindowDayMixed:
		return "~"
	default:
		return " "
	}
}
