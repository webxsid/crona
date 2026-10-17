package daily

import (
	"strings"

	"crona/tui/internal/api"
	helperpkg "crona/tui/internal/tui/helpers"
	viewcalendar "crona/tui/internal/tui/views/calendar"
	issuecore "crona/tui/internal/tui/views/issuecore"
	sessionmeta "crona/tui/internal/tui/views/sessionmeta"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
)

// detailCalendarMinRows is the space left under the selection before the calendar is shown.
const detailCalendarMinRows = 10

// renderDetail shows the focused pane's selection, with the month calendar below when it fits.
func renderDetail(theme types.Theme, state types.ContentState, width, height int) string {
	bodyWidth := max(1, width-2)
	var lines []string
	if state.Pane == "habits" {
		if habit, ok := selectedDailyHabit(state); ok {
			lines = habitDetailLines(theme, state, habit, bodyWidth)
		}
	} else if issue, ok := selectedDailyIssue(state); ok {
		lines = issueDetailLines(theme, state, issue, bodyWidth)
	}
	if len(lines) == 0 {
		lines = []string{theme.StyleDim.Render("Nothing selected")}
	}
	if spare := height - 1 - len(lines) - 1; spare >= detailCalendarMinRows {
		if calendar := detailCalendar(theme, state, bodyWidth, spare); len(calendar) > 0 {
			lines = append(append(lines, ""), calendar...)
		}
	}
	return dailySection(theme, "Selected", "", false, width, height, lines)
}

func selectedDailyIssue(state types.ContentState) (api.Issue, bool) {
	issues := filteredDailyIssues(state)
	indices := dailyIssueIndices(state, dailyAPIIssues(issues))
	cur := state.Cursors["issues"]
	if cur < 0 || cur >= len(indices) || indices[cur] >= len(issues) {
		return api.Issue{}, false
	}
	return issues[indices[cur]], true
}

func selectedDailyHabit(state types.ContentState) (api.HabitDailyItem, bool) {
	indices := dailyHabitIndices(state)
	cur := state.Cursors["habits"]
	if cur < 0 || cur >= len(indices) || indices[cur] >= len(state.DueHabits) {
		return api.HabitDailyItem{}, false
	}
	return state.DueHabits[indices[cur]], true
}

func issueDetailLines(theme types.Theme, state types.ContentState, issue api.Issue, width int) []string {
	repo, stream, worked := issue.RepoName, issue.StreamName, issue.WorkedSeconds
	if meta := sessionmeta.IssueMetaByID(state.AllIssues, issue.ID); meta != nil {
		repo, stream = meta.RepoName, meta.StreamName
		if worked <= 0 {
			worked = meta.WorkedSeconds
		}
	}
	statusStyle := theme.StyleNormal
	if style := issuecore.IssueStatusStyle(theme, string(issue.Status)); style != nil {
		statusStyle = *style
	}
	estimate := 0
	if issue.EstimateMinutes != nil {
		estimate = *issue.EstimateMinutes
	}
	lines := wrapDetail(theme.StyleHeader.Bold(true), issue.Title, width)
	lines = append(lines,
		detailField(theme, "status", statusStyle.Render(issuecore.PlainIssueStatus(string(issue.Status)))),
		detailField(theme, "context", theme.StyleNormal.Render(detailContext(repo, stream))),
		detailField(theme, "effort", theme.StyleNormal.Render(issueProgressText(worked, estimate, true))),
	)
	if date := issuecore.IssueDateLabel(issue.Status, issue.TodoForDate, issue.CompletedAt, issue.AbandonedAt, state.Settings); date != "" {
		lines = append(lines, detailField(theme, "date", theme.StyleNormal.Render(date)))
	}
	if issue.PinnedDaily {
		lines = append(lines, detailField(theme, "pinned", theme.StyleNormal.Render("yes")))
	}
	lines = append(lines, detailText(theme, "Description", issue.Description, width)...)
	return append(lines, detailText(theme, "Notes", issue.Notes, width)...)
}

func habitDetailLines(theme types.Theme, state types.ContentState, habit api.HabitDailyItem, width int) []string {
	status := "due"
	switch habit.Status {
	case "completed":
		status = "completed"
	case "failed":
		status = "failed"
	}
	lines := wrapDetail(theme.StyleHeader.Bold(true), habit.Name, width)
	lines = append(lines, detailField(theme, "status", theme.StyleNormal.Render(status)))
	if habit.DurationMinutes != nil {
		lines = append(lines, detailField(theme, "logged", theme.StyleNormal.Render(helperpkg.FormatCompactDurationMinutes(*habit.DurationMinutes))))
	}
	if habit.TargetMinutes != nil {
		lines = append(lines, detailField(theme, "target", theme.StyleNormal.Render(helperpkg.FormatCompactDurationMinutes(*habit.TargetMinutes))))
	}
	if meta := strings.TrimSpace(detailContext(habit.RepoName, habit.StreamName)); meta != "-" {
		lines = append(lines, detailField(theme, "context", theme.StyleNormal.Render(meta)))
	}
	lines = append(lines, detailText(theme, "Description", habit.Description, width)...)
	return append(lines, detailText(theme, "Notes", habit.Notes, width)...)
}

func detailField(theme types.Theme, label, value string) string {
	return theme.StyleDim.Render(padCells(label, 9)) + value
}

func detailContext(repo, stream string) string {
	repo, stream = strings.TrimSpace(repo), strings.TrimSpace(stream)
	switch {
	case repo != "" && stream != "":
		return repo + " › " + stream
	case repo != "":
		return repo
	case stream != "":
		return stream
	default:
		return "-"
	}
}

func detailText(theme types.Theme, title string, text *string, width int) []string {
	if text == nil || strings.TrimSpace(*text) == "" {
		return nil
	}
	lines := []string{"", theme.StyleDim.Bold(true).Render(title)}
	return append(lines, wrapDetail(theme.StyleNormal, strings.TrimSpace(*text), width)...)
}

func wrapDetail(style lipgloss.Style, text string, width int) []string {
	wrapped := lipgloss.NewStyle().Width(max(8, width)).Render(text)
	lines := strings.Split(wrapped, "\n")
	for i, line := range lines {
		lines[i] = style.Render(strings.TrimRight(line, " "))
	}
	return lines
}

func detailCalendar(theme types.Theme, state types.ContentState, width, maxLines int) []string {
	rawDate := dailyDate(state)
	if _, ok := viewcalendar.ParseDate(rawDate); !ok {
		return nil
	}
	for _, mode := range []viewcalendar.Mode{viewcalendar.ModeMonth, viewcalendar.ModeWeek} {
		lines := renderCalendarForMode(theme, rawDate, state.WeekStart, maxLines, mode, state.AwayDates())
		if len(lines) > 0 && len(lines) <= maxLines && viewcalendar.MaxLineWidth(lines) <= width {
			return lines
		}
	}
	return nil
}
