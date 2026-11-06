package summary

import (
	"fmt"
	"strings"

	viewhelpers "crona/tui/internal/tui/views/helpers"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
)

type issueStatus struct {
	status, label string
	color         lipgloss.Color
}

// issueStatuses is the bar's segment order, with each status's legend label and colour.
func issueStatuses(theme types.Theme) []issueStatus {
	return []issueStatus{
		{"done", "done", theme.ColorGreen},
		{"abandoned", "abandoned", theme.ColorRed},
		{"blocked", "blocked", theme.ColorRed},
		{"in_progress", "active", theme.ColorYellow},
		{"in_review", "review", theme.ColorMagenta},
		{"carried", "carried", theme.ColorLavender},
		{"ready", "ready", theme.ColorCyan},
		{"planned", "planned", theme.ColorBlue},
		{"backlog", "backlog", theme.ColorSubtle},
	}
}

// IssueStatusLegend keys the status bar: a coloured dot and count for each status present.
func IssueStatusLegend(theme types.Theme, counts map[string]int) string {
	var parts []string
	for _, item := range issueStatuses(theme) {
		if n := counts[item.status]; n > 0 {
			parts = append(parts, lipgloss.NewStyle().Foreground(item.color).Render("●")+
				theme.StyleNormal.Render(fmt.Sprintf(" %d ", n))+theme.StyleDim.Render(item.label))
		}
	}
	return strings.Join(parts, "  ")
}

// IssueStatusBar splits width cells by the day's task statuses, done first.
func IssueStatusBar(theme types.Theme, counts map[string]int, width int) string {
	width = max(8, width)
	order := issueStatuses(theme)
	total := 0
	for _, count := range counts {
		total += count
	}
	if total == 0 {
		return theme.StyleDim.Render(strings.Repeat("·", width))
	}
	segments := make([]string, 0, len(order))
	used := 0
	remainingStatuses := 0
	for _, item := range order {
		if counts[item.status] > 0 {
			remainingStatuses++
		}
	}
	for _, item := range order {
		count := counts[item.status]
		if count <= 0 {
			continue
		}
		segmentWidth := (count * width) / total
		if segmentWidth == 0 {
			segmentWidth = 1
		}
		if used+segmentWidth > width {
			segmentWidth = width - used
		}
		if remainingStatuses == 1 {
			segmentWidth = width - used
		}
		if segmentWidth <= 0 {
			continue
		}
		segments = append(
			segments,
			lipgloss.NewStyle().Foreground(item.color).Render(strings.Repeat("█", segmentWidth)),
		)
		used += segmentWidth
		remainingStatuses--
	}
	if used < width {
		segments = append(segments, theme.StyleDim.Render(strings.Repeat("█", width-used)))
	}
	return strings.Join(segments, "")
}

// HabitBar fills width cells with done habits, then failed ones, over the due count.
func HabitBar(theme types.Theme, completed, failed, total, width int) string {
	if width < 8 {
		width = 8
	}
	if total <= 0 {
		return theme.StyleDim.Render(strings.Repeat("·", width))
	}
	completedWidth := (completed * width) / total
	failedWidth := (failed * width) / total
	if completed > 0 && completedWidth == 0 {
		completedWidth = 1
	}
	if failed > 0 && failedWidth == 0 {
		failedWidth = 1
	}
	if completedWidth+failedWidth > width {
		failedWidth = max(0, width-completedWidth)
	}
	remainingWidth := width - completedWidth - failedWidth
	return viewhelpers.RenderGradientBar(
		completedWidth,
		completedWidth,
		viewhelpers.GradientBarPalette{
			Start: theme.ColorDullGreen,
			End:   theme.ColorCyan,
			Track: theme.ColorDim,
		},
	) + viewhelpers.RenderGradientBar(
		failedWidth,
		failedWidth,
		viewhelpers.GradientBarPalette{
			Start: theme.ColorDullRed,
			End:   theme.ColorOrange,
			Track: theme.ColorDim,
		},
	) + viewhelpers.RenderGradientBar(
		remainingWidth,
		0,
		viewhelpers.GradientBarPalette{
			Start: theme.ColorDim,
			End:   theme.ColorDim,
			Track: theme.ColorDim,
		},
	)
}
