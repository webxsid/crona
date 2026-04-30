package daily

import (
	"fmt"
	"strings"
	"time"

	shareddatefmt "crona/shared/datefmt"
	"crona/tui/internal/api"
	helperpkg "crona/tui/internal/tui/helpers"
	viewchrome "crona/tui/internal/tui/views/chrome"
	contextmeta "crona/tui/internal/tui/views/contextmeta"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	viewruntime "crona/tui/internal/tui/views/runtime"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

func renderSummary(theme types.Theme, state types.ContentState, width, height int) string {
	rawDate := viewruntime.CurrentDashboardDate(state)
	totalIssues, totalEstimate, completedCount, abandonedCount := 0, 0, 0, 0
	totalHabits, completedHabits, failedHabits, habitMinutes := len(state.DueHabits), 0, 0, 0
	habitTargetMinutes := 0
	issueStatusCounts := map[string]int{}
	for _, habit := range state.DueHabits {
		switch habit.Status {
		case "completed":
			completedHabits++
		case "failed":
			failedHabits++
		}
		if habit.DurationMinutes != nil {
			habitMinutes += *habit.DurationMinutes
		}
		if habit.TargetMinutes != nil {
			habitTargetMinutes += *habit.TargetMinutes
		}
	}
	for _, issue := range state.DailyIssues {
		totalIssues++
		issueStatusCounts[string(issue.Status)]++
		if issue.EstimateMinutes != nil {
			totalEstimate += *issue.EstimateMinutes
		}
		switch issue.Status {
		case "done":
			completedCount++
		case "abandoned":
			abandonedCount++
		}
	}
	if state.DailySummary != nil {
		rawDate = state.DailySummary.Date
	}
	selectedDate, hasSelectedDate := parseSummaryDate(rawDate)
	dateText := helperpkg.FormatDisplayDate(rawDate, state.Settings)
	if hasSelectedDate && !displayPatternIncludesWeek(state.Settings) {
		dateText += fmt.Sprintf(" - Week %02d", isoWeek(selectedDate))
	}
	resolvedCount := completedCount + abandonedCount
	scopeText := contextmeta.DefaultScopeLabel(state.Context)
	summaryInnerW := max(24, width-8)
	habitMeta := theme.StyleDim.Render("logged " + helperpkg.FormatCompactDurationMinutes(habitMinutes))
	if habitTargetMinutes > 0 {
		habitMeta = theme.StyleDim.Render(fmt.Sprintf("logged %s / target %s", helperpkg.FormatCompactDurationMinutes(habitMinutes), helperpkg.FormatCompactDurationMinutes(habitTargetMinutes)))
	}
	leftWidth := summaryInnerW
	lines := buildSummaryLines(
		theme,
		state.Height,
		leftWidth,
		dateText,
		scopeText,
		issueStatusCounts,
		resolvedCount,
		totalIssues,
		totalEstimate,
		completedHabits,
		totalHabits,
		failedHabits,
		habitMinutes,
		habitTargetMinutes,
		habitMeta,
	)
	lines = viewchrome.ClipSummaryLines(theme, lines, height)
	calendarLines := []string(nil)
	if shouldRenderSummaryCalendar(summaryInnerW) && hasSelectedDate {
		calendarLines = summaryCalendarWindow(theme, selectedDate, len(lines))
		leftWidth, _ = summaryColumnWidths(summaryInnerW, maxLineWidth(calendarLines), 3)
		if leftWidth != summaryInnerW {
			lines = buildSummaryLines(
				theme,
				state.Height,
				leftWidth,
				dateText,
				scopeText,
				issueStatusCounts,
				resolvedCount,
				totalIssues,
				totalEstimate,
				completedHabits,
				totalHabits,
				failedHabits,
				habitMinutes,
				habitTargetMinutes,
				habitMeta,
			)
			lines = viewchrome.ClipSummaryLines(theme, lines, height)
			calendarLines = summaryCalendarWindow(theme, selectedDate, len(lines))
		}
	}
	if len(calendarLines) > 0 {
		lines = mergeSummaryCalendar(lines, calendarLines, summaryInnerW, 3)
	}
	return lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(theme.ColorDim).Padding(1, 2).Width(width - 2).Height(max(1, height-2)).Render(viewhelpers.StringsJoin(lines))
}

func renderCompactSummaryRow(width int, segments []string, renderBar func(int) string, sizeBar func(int, int) int) string {
	parts := make([]string, 0, len(segments))
	for _, segment := range segments {
		if strings.TrimSpace(segment) != "" {
			parts = append(parts, segment)
		}
	}
	text := strings.Join(parts, "  ")
	if sizeBar == nil {
		sizeBar = compactSummaryBarWidth
	}
	barWidth := sizeBar(width, lipgloss.Width(text))
	if barWidth < 1 || renderBar == nil {
		return viewhelpers.Truncate(text, width)
	}
	bar := renderBar(barWidth)
	if strings.TrimSpace(bar) == "" {
		return viewhelpers.Truncate(text, width)
	}
	bar = ansi.Truncate(bar, barWidth, "")
	return viewhelpers.Truncate(text+"  "+bar, width)
}

func renderCompactMetadataRow(width int, left, right string) string {
	row := left
	if strings.TrimSpace(right) == "" {
		return viewhelpers.Truncate(row, width)
	}
	remaining := width - lipgloss.Width(left) - lipgloss.Width(right) - 2
	if remaining < 1 {
		return viewhelpers.Truncate(left+"  "+right, width)
	}
	return left + strings.Repeat(" ", remaining+2) + right
}

func compactSummaryBarWidth(totalWidth, textWidth int) int {
	remaining := totalWidth - textWidth - 2
	if remaining < 8 {
		return 0
	}
	if remaining > 18 {
		return 18
	}
	return remaining
}

func tinySummaryBarWidth(totalWidth, textWidth int) int {
	remaining := totalWidth - textWidth - 2
	if remaining < 6 {
		return 0
	}
	if remaining > 12 {
		return 12
	}
	return remaining
}

func wideSummaryBarWidth(totalWidth int) int {
	width := totalWidth * 3 / 4
	if width < 20 {
		width = 20
	}
	if width > totalWidth-4 {
		width = max(12, totalWidth-4)
	}
	return width
}

func shouldRenderSummaryCalendar(summaryInnerW int) bool {
	return summaryInnerW >= 84
}

func summaryColumnWidths(summaryInnerW, calendarWidth, gutterWidth int) (int, int) {
	if calendarWidth < 1 {
		return summaryInnerW, 0
	}
	if gutterWidth < 0 {
		gutterWidth = 0
	}
	rightWidth := calendarWidth
	maxRightWidth := max(24, summaryInnerW/4)
	if rightWidth > maxRightWidth {
		rightWidth = maxRightWidth
	}
	leftWidth := summaryInnerW - gutterWidth - rightWidth
	if leftWidth < 32 {
		leftWidth = 32
		rightWidth = max(0, summaryInnerW-gutterWidth-leftWidth)
	}
	return leftWidth, rightWidth
}

func renderSummaryCalendar(theme types.Theme, selected time.Time) []string {
	monthStart := time.Date(selected.Year(), selected.Month(), 1, 0, 0, 0, 0, selected.Location())
	lines := []string{
		theme.StyleHeader.Render(monthStart.Format("January 2006")),
		theme.StyleDim.Render(fmt.Sprintf("Week %02d", isoWeek(selected))),
		theme.StyleDim.Render("Wk  Mo Tu We Th Fr Sa Su"),
	}
	offset := (int(monthStart.Weekday()) + 6) % 7
	gridStart := monthStart.AddDate(0, 0, -offset)
	selectedWeek := isoWeek(selected)
	for week := 0; week < 6; week++ {
		rowStart := gridStart.AddDate(0, 0, week*7)
		rowWeek := isoWeek(rowStart)
		weekLabel := fmt.Sprintf("%2d", rowWeek)
		if rowWeek == selectedWeek {
			weekLabel = theme.StyleHeader.Render(weekLabel)
		} else {
			weekLabel = theme.StyleDim.Render(weekLabel)
		}
		cells := make([]string, 0, 7)
		for day := 0; day < 7; day++ {
			current := rowStart.AddDate(0, 0, day)
			cell := fmt.Sprintf("%2d", current.Day())
			switch {
			case sameSummaryDay(current, selected):
				cell = theme.StyleCursor.Render(cell)
			case current.Month() != monthStart.Month():
				cell = theme.StyleDim.Render(cell)
			case rowWeek == selectedWeek:
				cell = theme.StyleHeader.Render(cell)
			default:
				cell = theme.StyleNormal.Render(cell)
			}
			cells = append(cells, cell)
		}
		lines = append(lines, weekLabel+"  "+strings.Join(cells, " "))
	}
	return lines
}

func summaryCalendarWindow(theme types.Theme, selected time.Time, maxLines int) []string {
	if maxLines <= 0 {
		return nil
	}
	full := renderSummaryCalendar(theme, selected)
	if len(full) <= maxLines {
		return full
	}
	if maxLines <= 3 {
		return full[:maxLines]
	}
	headers := full[:3]
	weeks := full[3:]
	visibleWeeks := maxLines - len(headers)
	if visibleWeeks >= len(weeks) {
		return full
	}
	monthStart := time.Date(selected.Year(), selected.Month(), 1, 0, 0, 0, 0, selected.Location())
	selectedWeek := weekIndexForDate(monthStart, selected)
	start := selectedWeek - (visibleWeeks / 2)
	if start < 0 {
		start = 0
	}
	if start+visibleWeeks > len(weeks) {
		start = len(weeks) - visibleWeeks
	}
	window := append([]string{}, headers...)
	window = append(window, weeks[start:start+visibleWeeks]...)
	return window
}

func buildSummaryLines(theme types.Theme, paneHeight, width int, dateText, scopeText string, issueStatusCounts map[string]int, resolvedCount, totalIssues, totalEstimate, completedHabits, totalHabits, failedHabits, habitMinutes, habitTargetMinutes int, habitMeta string) []string {
	issueSummary := fmt.Sprintf(
		"%s  %s  %s",
		theme.StyleHeader.Render("Issues"),
		theme.StyleNormal.Render(fmt.Sprintf("%d/%d resolved", resolvedCount, totalIssues)),
		theme.StyleDim.Render("estimate "+helperpkg.FormatCompactDurationMinutes(totalEstimate)),
	)
	habitSummary := fmt.Sprintf(
		"%s  %s",
		theme.StyleHeader.Render("Habits"),
		theme.StyleNormal.Render(fmt.Sprintf("%d/%d completed", completedHabits, totalHabits)),
	)
	issueBarWidth := wideSummaryBarWidth(width)
	habitVisual := []string{
		habitSummary,
		renderHabitBar(theme, completedHabits, failedHabits, totalHabits, issueBarWidth),
		habitMeta,
		theme.StyleDim.Render(fmt.Sprintf("failed %d   remaining %d", failedHabits, max(0, totalHabits-completedHabits-failedHabits))),
	}
	lines := []string{
		theme.StylePaneTitle.Render("Daily Dashboard"),
		theme.StyleHeader.Render(fmt.Sprintf("For %s", dateText)),
		theme.StyleDim.Render(scopeText),
		theme.StyleDim.Render("[,] prev   [.] next   [g] today"),
		"",
	}
	switch {
	case paneHeight < 37:
		return []string{
			renderCompactMetadataRow(width,
				theme.StylePaneTitle.Render("Daily Dashboard"),
				theme.StyleHeader.Render(dateText),
			),
			renderCompactMetadataRow(width,
				theme.StyleDim.Render(scopeText),
				theme.StyleDim.Render("[,] [.] [g]"),
			),
			renderCompactSummaryRow(width,
				[]string{
					theme.StyleHeader.Render("Issues"),
					theme.StyleNormal.Render(fmt.Sprintf("%d/%d", resolvedCount, totalIssues)),
					theme.StyleDim.Render(helperpkg.FormatCompactDurationMinutes(totalEstimate)),
					theme.StyleDim.Render(compactIssueLegend(issueStatusCounts)),
				},
				func(barWidth int) string { return renderIssueStatusBar(theme, issueStatusCounts, barWidth) },
				tinySummaryBarWidth,
			),
			renderCompactSummaryRow(width,
				[]string{
					theme.StyleHeader.Render("Habits"),
					theme.StyleNormal.Render(fmt.Sprintf("%d/%d", completedHabits, totalHabits)),
					theme.StyleDim.Render(compactHabitProgress(habitMinutes, habitTargetMinutes)),
					theme.StyleDim.Render(fmt.Sprintf("f%d r%d", failedHabits, max(0, totalHabits-completedHabits-failedHabits))),
				},
				func(barWidth int) string {
					return renderHabitBar(theme, completedHabits, failedHabits, totalHabits, barWidth)
				},
				tinySummaryBarWidth,
			),
		}
	case paneHeight < 48:
		lines = append(lines,
			renderCompactSummaryRow(width,
				[]string{
					theme.StyleHeader.Render("Issues"),
					theme.StyleNormal.Render(fmt.Sprintf("%d/%d resolved", resolvedCount, totalIssues)),
					theme.StyleDim.Render("estimate " + helperpkg.FormatCompactDurationMinutes(totalEstimate)),
				},
				func(barWidth int) string { return renderIssueStatusBar(theme, issueStatusCounts, barWidth) },
				compactSummaryBarWidth,
			),
			renderCompactSummaryRow(width,
				[]string{
					theme.StyleHeader.Render("Habits"),
					theme.StyleNormal.Render(fmt.Sprintf("%d/%d completed", completedHabits, totalHabits)),
					habitMeta,
				},
				func(barWidth int) string {
					return renderHabitBar(theme, completedHabits, failedHabits, totalHabits, barWidth)
				},
				compactSummaryBarWidth,
			),
		)
	case paneHeight < 55:
		lines = append(lines,
			renderCompactSummaryRow(width,
				[]string{
					theme.StyleHeader.Render("Issues"),
					theme.StyleNormal.Render(fmt.Sprintf("%d/%d resolved", resolvedCount, totalIssues)),
					theme.StyleDim.Render("estimate " + helperpkg.FormatCompactDurationMinutes(totalEstimate)),
				},
				func(barWidth int) string { return renderIssueStatusBar(theme, issueStatusCounts, barWidth) },
				compactSummaryBarWidth,
			),
			theme.StyleDim.Render(renderIssueLegend(issueStatusCounts)),
			"",
			renderCompactSummaryRow(width,
				[]string{
					theme.StyleHeader.Render("Habits"),
					theme.StyleNormal.Render(fmt.Sprintf("%d/%d completed", completedHabits, totalHabits)),
					habitMeta,
				},
				func(barWidth int) string {
					return renderHabitBar(theme, completedHabits, failedHabits, totalHabits, barWidth)
				},
				compactSummaryBarWidth,
			),
			theme.StyleDim.Render(fmt.Sprintf("failed %d   remaining %d", failedHabits, max(0, totalHabits-completedHabits-failedHabits))),
		)
	default:
		lines = append(lines,
			issueSummary,
			renderIssueStatusBar(theme, issueStatusCounts, issueBarWidth),
			theme.StyleDim.Render(renderIssueLegend(issueStatusCounts)),
			"",
		)
		lines = append(lines, habitVisual...)
	}
	return lines
}

func mergeSummaryCalendar(leftLines, calendarLines []string, summaryInnerW, gutterWidth int) []string {
	leftWidth, rightWidth := summaryColumnWidths(summaryInnerW, maxLineWidth(calendarLines), gutterWidth)
	if rightWidth < 1 {
		return leftLines
	}
	totalLines := max(len(leftLines), len(calendarLines))
	start := 0
	if totalLines > len(calendarLines) {
		start = (totalLines - len(calendarLines)) / 2
	}
	merged := make([]string, 0, totalLines)
	gutter := strings.Repeat(" ", gutterWidth)
	for i := 0; i < totalLines; i++ {
		left := ""
		if i < len(leftLines) {
			left = ansi.Truncate(leftLines[i], leftWidth, "")
		}
		left = lipgloss.NewStyle().Width(leftWidth).MaxWidth(leftWidth).Render(left)
		if i >= start && i < start+len(calendarLines) {
			right := calendarLines[i-start]
			right = lipgloss.NewStyle().Width(rightWidth).MaxWidth(rightWidth).Render(right)
			merged = append(merged, left+gutter+right)
			continue
		}
		merged = append(merged, left)
	}
	return merged
}

func maxLineWidth(lines []string) int {
	width := 0
	for _, line := range lines {
		if w := lipgloss.Width(line); w > width {
			width = w
		}
	}
	return width
}

func displayPatternIncludesWeek(settings *api.CoreSettings) bool {
	pattern := shareddatefmt.DisplayPattern(settings)
	pattern = strings.ReplaceAll(pattern, "[Week]", "")
	pattern = strings.ReplaceAll(pattern, "[week]", "")
	return strings.Contains(pattern, "W")
}

func parseSummaryDate(raw string) (time.Time, bool) {
	parsed, err := time.Parse("2006-01-02", strings.TrimSpace(raw))
	if err != nil {
		return time.Now(), false
	}
	return parsed, true
}

func isoWeek(value time.Time) int {
	_, week := value.ISOWeek()
	return week
}

func sameSummaryDay(a, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}

func weekIndexForDate(monthStart, selected time.Time) int {
	offset := (int(monthStart.Weekday()) + 6) % 7
	gridStart := monthStart.AddDate(0, 0, -offset)
	days := int(selected.Sub(gridStart).Hours() / 24)
	if days < 0 {
		return 0
	}
	return days / 7
}

func compactIssueLegend(counts map[string]int) string {
	order := []string{"done", "abandoned", "blocked", "in_progress", "in_review", "ready", "planned", "backlog"}
	labels := map[string]string{
		"done":        "d",
		"abandoned":   "a",
		"blocked":     "b",
		"in_progress": "ip",
		"in_review":   "ir",
		"ready":       "r",
		"planned":     "p",
		"backlog":     "bk",
	}
	parts := make([]string, 0, len(order))
	for _, status := range order {
		if count := counts[status]; count > 0 {
			parts = append(parts, fmt.Sprintf("%s%d", labels[status], count))
		}
	}
	if len(parts) == 0 {
		return "none"
	}
	return strings.Join(parts, " ")
}

func compactHabitProgress(loggedMinutes, targetMinutes int) string {
	if targetMinutes > 0 {
		return fmt.Sprintf("%s/%s", helperpkg.FormatCompactDurationMinutes(loggedMinutes), helperpkg.FormatCompactDurationMinutes(targetMinutes))
	}
	return helperpkg.FormatCompactDurationMinutes(loggedMinutes)
}
