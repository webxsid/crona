package daily

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"crona/tui/internal/api"
	helperpkg "crona/tui/internal/tui/helpers"
	issuecore "crona/tui/internal/tui/views/issuecore"
	viewsummary "crona/tui/internal/tui/views/summary"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// habitTrailDays is how many days the habit's trail of dots covers.
const habitTrailDays = 14

// habitHitWindow is the days the hit-rate arc looks back over.
const habitHitWindow = 30

// habitDay is one day of a habit's history: its status and the minutes logged.
type habitDay struct {
	status  string
	minutes int
}

// habitDays folds the habit's history into one entry per date.
func habitDays(state types.ContentState, habitID int64) map[string]habitDay {
	days := map[string]habitDay{}
	history := state.HabitHistory
	if state.DetailHistoryHabitID == habitID {
		history = state.DetailHabitHistory
	}
	for _, entry := range history {
		if entry.HabitID != habitID {
			continue
		}
		day := days[entry.Date]
		if entry.DurationMinutes != nil {
			day.minutes += *entry.DurationMinutes
		}
		if status := string(entry.Status); status != "" && day.status != "completed" {
			day.status = status
		}
		days[entry.Date] = day
	}
	return days
}

// habitDetailView lays a habit out like an issue: a Today arc of minutes logged against the
// target (or done or not), a hit-rate arc over the last 30 days, and the facts beside them;
// then a two-week trail, recent entries, notes and momentum.
func habitDetailView(
	theme types.Theme,
	state types.ContentState,
	habit api.HabitDailyItem,
	width int,
	status string,
) []string {
	days := habitDays(state, habit.ID)
	today := dailyDate(state)
	arcs := []viewsummary.Arc{habitTodayArc(theme, habit, status)}
	if hit, ok := habitHitArc(theme, days, today); ok {
		arcs = append(arcs, hit)
	}
	facts := []string{habitStatusLine(theme, habit, status)}
	if context := strings.TrimSpace(detailContext(habit.RepoName, habit.StreamName)); context != "-" {
		facts = append(facts, theme.StyleNormal.Render(context))
	}
	if len(habit.Tags) > 0 {
		facts = append(facts, issuecore.TagChips(theme, habit.Tags, state.Tags, max(8, width-viewsummary.ArcsWidth(2)-3)))
	}
	arcWidth := viewsummary.ArcsWidth(len(arcs))
	lines := detailDescription(theme, habit.Description, width)
	lines = append(lines, viewsummary.BesideArcs(viewsummary.Arcs(theme, arcs, arcWidth), facts, arcWidth, width)...)
	lines = append(lines, habitTrailLines(theme, days, today, width)...)
	lines = append(lines, habitRecentLines(theme, days, habit, today, width)...)
	lines = append(lines, detailText(theme, "Notes", habit.Notes, width)...)
	return append(lines, momentumLines(theme, state, habitMomentums(state, habit.ID), width)...)
}

// habitTodayArc fills with minutes logged against the target, or full once done without one:
// green done, red failed, yellow on the way.
func habitTodayArc(theme types.Theme, habit api.HabitDailyItem, status string) viewsummary.Arc {
	arc := viewsummary.Arc{Label: "Today", Color: theme.ColorYellow}
	logged := 0
	if habit.DurationMinutes != nil {
		logged = *habit.DurationMinutes
	}
	switch habit.Status {
	case "completed":
		arc.Color, arc.Fraction, arc.Value = theme.ColorGreen, 1, "✓ done"
	case "failed":
		arc.Color, arc.Value = theme.ColorRed, "✗"
	default:
		arc.Value = status
		if status == "not due today" || status == "paused" {
			arc.Value = "off"
		}
	}
	if habit.TargetMinutes != nil && *habit.TargetMinutes > 0 {
		arc.Label += "/" + helperpkg.FormatCompactDurationMinutes(*habit.TargetMinutes)
		arc.Fraction = float64(logged) / float64(*habit.TargetMinutes)
		if logged > 0 || habit.Status == "" {
			arc.Value = helperpkg.FormatCompactDurationMinutes(logged)
		}
		if habit.Status == "failed" {
			arc.Value = "✗ " + helperpkg.FormatCompactDurationMinutes(logged)
		}
	}
	return arc
}

// habitHitArc is the share of the last 30 days with an outcome that the habit was done.
func habitHitArc(theme types.Theme, days map[string]habitDay, today string) (viewsummary.Arc, bool) {
	end, err := time.Parse("2006-01-02", today)
	if err != nil {
		return viewsummary.Arc{}, false
	}
	done, decided := 0, 0
	for i := range habitHitWindow {
		day, ok := days[end.AddDate(0, 0, -i).Format("2006-01-02")]
		if !ok || day.status == "" {
			continue
		}
		decided++
		if day.status == "completed" {
			done++
		}
	}
	if decided == 0 {
		return viewsummary.Arc{}, false
	}
	fraction := float64(done) / float64(decided)
	color := theme.ColorRed
	switch {
	case fraction >= 0.8:
		color = theme.ColorGreen
	case fraction >= 0.5:
		color = theme.ColorYellow
	}
	return viewsummary.Arc{
		Label: fmt.Sprintf("Hit/%dd", habitHitWindow), Value: fmt.Sprintf("%.0f%%", fraction*100),
		Fraction: fraction, Color: color,
	}, true
}

// habitStatusLine is the status as a coloured dot, then the schedule.
func habitStatusLine(theme types.Theme, habit api.HabitDailyItem, status string) string {
	color := theme.ColorYellow
	switch habit.Status {
	case "completed":
		color = theme.ColorGreen
	case "failed":
		color = theme.ColorRed
	}
	if status == "paused" || status == "not due today" {
		color = theme.ColorSubtle
	}
	line := lipgloss.NewStyle().Foreground(color).Render("● " + status)
	if schedule := habitSchedule(habit.Habit); schedule != "" {
		line += theme.StyleDim.Render(" · " + schedule)
	}
	return line
}

// habitSchedule reads the schedule in words: daily, weekdays, or the weekly days.
func habitSchedule(habit api.Habit) string {
	if habit.ScheduleType != "weekly" || len(habit.Weekdays) == 0 {
		return strings.ReplaceAll(string(habit.ScheduleType), "_", " ")
	}
	names := []string{"Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"}
	days := []string{}
	for _, day := range habit.Weekdays {
		if day >= 0 && day < len(names) {
			days = append(days, names[day])
		}
	}
	return strings.Join(days, " ")
}

// habitTrailLines are the last two weeks as dots, oldest first: ● done, ✗ failed, · nothing.
func habitTrailLines(theme types.Theme, days map[string]habitDay, today string, width int) []string {
	end, err := time.Parse("2006-01-02", today)
	if err != nil {
		return nil
	}
	var trail strings.Builder
	for i := habitTrailDays - 1; i >= 0; i-- {
		switch days[end.AddDate(0, 0, -i).Format("2006-01-02")].status {
		case "completed":
			trail.WriteString(lipgloss.NewStyle().Foreground(theme.ColorGreen).Render("●"))
		case "failed":
			trail.WriteString(lipgloss.NewStyle().Foreground(theme.ColorRed).Render("✗"))
		default:
			trail.WriteString(theme.StyleDim.Render("·"))
		}
		if i > 0 {
			trail.WriteString(" ")
		}
	}
	row := theme.StyleDim.Render(padCells("Last 14d", 11)) + trail.String() + theme.StyleDim.Render("  today")
	return []string{"", ansi.Truncate(row, width, "…")}
}

// habitRecentLines list the latest days with an outcome or time, newest first: date, a bar of
// minutes against the target (or the longest listed), the minutes and the status.
func habitRecentLines(theme types.Theme, days map[string]habitDay, habit api.HabitDailyItem, today string, width int) []string {
	dates := []string{}
	for date, day := range days {
		if date <= today && (day.status != "" || day.minutes > 0) {
			dates = append(dates, date)
		}
	}
	title := theme.StyleDim.Bold(true).Render("Recent")
	if len(dates) == 0 {
		return []string{"", title, "  " + theme.StyleDim.Render("nothing logged yet")}
	}
	slices.Sort(dates)
	slices.Reverse(dates)
	shown := dates[:min(len(dates), detailSessionRows)]
	scale := 1
	if habit.TargetMinutes != nil && *habit.TargetMinutes > 0 {
		scale = *habit.TargetMinutes
	}
	for _, date := range shown {
		scale = max(scale, days[date].minutes)
	}
	lines := []string{"", title}
	for _, date := range shown {
		day := days[date]
		when := date
		if parsed, err := time.Parse("2006-01-02", date); err == nil {
			when = parsed.Format("Mon 2 Jan")
		}
		color, mark := theme.ColorYellow, "·"
		switch day.status {
		case "completed":
			color, mark = theme.ColorGreen, "✓"
		case "failed":
			color, mark = theme.ColorRed, "✗"
		}
		filled := 0
		if day.minutes > 0 {
			filled = max(1, day.minutes*sessionBarCells/scale)
		}
		minutes := "-"
		if day.minutes > 0 {
			minutes = helperpkg.FormatCompactDurationMinutes(day.minutes)
		}
		row := "  " + theme.StyleDim.Render(padCells(when, 11)) + "  " +
			lipgloss.NewStyle().Foreground(color).Render(strings.Repeat("█", filled)) +
			theme.StyleDim.Render(strings.Repeat("░", sessionBarCells-filled)) + "  " +
			theme.StyleNormal.Render(padCells(minutes, 6)) + " " +
			lipgloss.NewStyle().Foreground(color).Render(mark)
		lines = append(lines, ansi.Truncate(row, width, "…"))
	}
	if more := len(dates) - len(shown); more > 0 {
		lines = append(lines, "  "+theme.StyleDim.Render(fmt.Sprintf("+%d earlier", more)))
	}
	return lines
}
