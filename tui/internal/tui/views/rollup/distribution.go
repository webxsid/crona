package rollup

import (
	"fmt"
	"strings"

	"crona/tui/internal/api"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
)

func renderDistributionSection(theme types.Theme, title string, summary *api.TimeDistributionSummary) []string {
	lines := []string{theme.StyleHeader.Render(title)}
	if summary == nil || len(summary.Rows) == 0 {
		return append(lines, theme.StyleDim.Render("  No data"))
	}
	limit := min(3, len(summary.Rows))
	for i := 0; i < limit; i++ {
		row := summary.Rows[i]
		lines = append(lines, fmt.Sprintf("  %s  %d%%  %s", viewhelpers.Truncate(row.Label, 18), int(row.Percent+0.5), viewhelpers.FormatClock(row.WorkedSeconds)))
	}
	return lines
}

func statusStyle(theme types.Theme, status string) lipgloss.Style {
	switch status {
	case "done":
		return lipgloss.NewStyle().Foreground(theme.ColorGreen)
	case "missed":
		return lipgloss.NewStyle().Foreground(theme.ColorRed)
	case "carry_over":
		return lipgloss.NewStyle().Foreground(theme.ColorYellow)
	case "planned":
		return lipgloss.NewStyle().Foreground(theme.ColorBlue)
	case "mixed":
		return lipgloss.NewStyle().Foreground(theme.ColorMagenta)
	default:
		return lipgloss.NewStyle().Foreground(theme.ColorDim)
	}
}

func levelStyle(theme types.Theme, level string) lipgloss.Style {
	switch level {
	case "strong":
		return lipgloss.NewStyle().Foreground(theme.ColorGreen)
	case "steady":
		return lipgloss.NewStyle().Foreground(theme.ColorCyan)
	case "overextended":
		return lipgloss.NewStyle().Foreground(theme.ColorRed)
	default:
		return lipgloss.NewStyle().Foreground(theme.ColorYellow)
	}
}

func biasStyle(theme types.Theme, bias string) lipgloss.Style {
	switch bias {
	case "under":
		return lipgloss.NewStyle().Foreground(theme.ColorYellow)
	case "over":
		return lipgloss.NewStyle().Foreground(theme.ColorCyan)
	default:
		return lipgloss.NewStyle().Foreground(theme.ColorGreen)
	}
}

func progressStyle(theme types.Theme, bias string, status string) lipgloss.Style {
	switch strings.TrimSpace(status) {
	case "on track":
		return lipgloss.NewStyle().Foreground(theme.ColorGreen)
	case "at risk":
		return lipgloss.NewStyle().Foreground(theme.ColorYellow)
	case "over":
		return lipgloss.NewStyle().Foreground(theme.ColorRed)
	case "unestimated":
		return biasStyle(theme, bias)
	default:
		return lipgloss.NewStyle().Foreground(theme.ColorWhite)
	}
}

func prettyWindowStatus(status string) string {
	switch status {
	case "done":
		return "done"
	case "missed":
		return "missed"
	case "carry_over":
		return "carry over"
	case "planned":
		return "planned"
	case "mixed":
		return "mixed"
	default:
		return "no activity"
	}
}

func padStatus(value string, width int) string {
	current := lipgloss.Width(value)
	if current >= width {
		return value
	}
	return value + strings.Repeat(" ", width-current)
}

func itoa(v int) string {
	return fmt.Sprintf("%d", v)
}
