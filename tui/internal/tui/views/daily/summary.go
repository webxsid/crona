package daily

import (
	"fmt"
	"strings"

	helperpkg "crona/tui/internal/tui/helpers"
	viewchrome "crona/tui/internal/tui/views/chrome"
	contextmeta "crona/tui/internal/tui/views/contextmeta"
	viewruntime "crona/tui/internal/tui/views/runtime"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

func renderSummary(theme types.Theme, state types.ContentState, width, height int) string {
	dateText := viewruntime.CurrentDashboardDate(state)
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
		dateText = state.DailySummary.Date
	}
	resolvedCount := completedCount + abandonedCount
	scopeText := contextmeta.DefaultScopeLabel(state.Context)
	summaryInnerW := max(24, width-8)
	issueBarWidth := max(16, summaryInnerW-4)
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
	habitMeta := theme.StyleDim.Render("logged " + helperpkg.FormatCompactDurationMinutes(habitMinutes))
	if habitTargetMinutes > 0 {
		habitMeta = theme.StyleDim.Render(fmt.Sprintf("logged %s / target %s", helperpkg.FormatCompactDurationMinutes(habitMinutes), helperpkg.FormatCompactDurationMinutes(habitTargetMinutes)))
	}
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
	case state.Height < 37:
		lines = []string{
			renderCompactMetadataRow(summaryInnerW,
				theme.StylePaneTitle.Render("Daily Dashboard"),
				theme.StyleHeader.Render(dateText),
			),
			renderCompactMetadataRow(summaryInnerW,
				theme.StyleDim.Render(scopeText),
				theme.StyleDim.Render("[,] [.] [g]"),
			),
			renderCompactSummaryRow(summaryInnerW,
				[]string{
					theme.StyleHeader.Render("Issues"),
					theme.StyleNormal.Render(fmt.Sprintf("%d/%d", resolvedCount, totalIssues)),
					theme.StyleDim.Render(helperpkg.FormatCompactDurationMinutes(totalEstimate)),
					theme.StyleDim.Render(compactIssueLegend(issueStatusCounts)),
				},
				func(barWidth int) string { return renderIssueStatusBar(theme, issueStatusCounts, barWidth) },
				tinySummaryBarWidth,
			),
			renderCompactSummaryRow(summaryInnerW,
				[]string{
					theme.StyleHeader.Render("Habits"),
					theme.StyleNormal.Render(fmt.Sprintf("%d/%d", completedHabits, totalHabits)),
					theme.StyleDim.Render(compactHabitProgress(habitMinutes, habitTargetMinutes)),
					theme.StyleDim.Render(fmt.Sprintf("f%d r%d", failedHabits, max(0, totalHabits-completedHabits-failedHabits))),
				},
				func(barWidth int) string { return renderHabitBar(theme, completedHabits, failedHabits, totalHabits, barWidth) },
				tinySummaryBarWidth,
			),
		}
	case state.Height < 48:
		lines = append(lines,
			renderCompactSummaryRow(summaryInnerW,
				[]string{
					theme.StyleHeader.Render("Issues"),
					theme.StyleNormal.Render(fmt.Sprintf("%d/%d resolved", resolvedCount, totalIssues)),
					theme.StyleDim.Render("estimate " + helperpkg.FormatCompactDurationMinutes(totalEstimate)),
				},
				func(barWidth int) string { return renderIssueStatusBar(theme, issueStatusCounts, barWidth) },
				compactSummaryBarWidth,
			),
			renderCompactSummaryRow(summaryInnerW,
				[]string{
					theme.StyleHeader.Render("Habits"),
					theme.StyleNormal.Render(fmt.Sprintf("%d/%d completed", completedHabits, totalHabits)),
					habitMeta,
				},
				func(barWidth int) string { return renderHabitBar(theme, completedHabits, failedHabits, totalHabits, barWidth) },
				compactSummaryBarWidth,
			),
		)
	case state.Height < 55:
		lines = append(lines,
			renderCompactSummaryRow(summaryInnerW,
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
			renderCompactSummaryRow(summaryInnerW,
				[]string{
					theme.StyleHeader.Render("Habits"),
					theme.StyleNormal.Render(fmt.Sprintf("%d/%d completed", completedHabits, totalHabits)),
					habitMeta,
				},
				func(barWidth int) string { return renderHabitBar(theme, completedHabits, failedHabits, totalHabits, barWidth) },
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
	lines = viewchrome.ClipSummaryLines(theme, lines, height)
	return lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(theme.ColorDim).Padding(1, 2).Width(width-2).Height(max(1, height-2)).Render(viewhelpers.StringsJoin(lines))
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
