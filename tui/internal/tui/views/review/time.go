package review

import (
	"fmt"
	"math"
	"strings"

	helperpkg "crona/tui/internal/tui/helpers"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
)

func renderWindowLine(theme types.Theme, state types.ContentState) string {
	if state.DashboardWindow == nil {
		return "Plan  planned -  done -  missed -  carry -"
	}
	return fmt.Sprintf(
		"%s  %s %d  %s %d  %s %d  %s %d",
		theme.StyleHeader.Render("Plan"),
		lipgloss.NewStyle().
			Foreground(theme.ColorBlue).
			Render("planned"),
		state.DashboardWindow.PlannedCount,
		lipgloss.NewStyle().
			Foreground(theme.ColorGreen).
			Render("done"),
		state.DashboardWindow.CompletedCount,
		lipgloss.NewStyle().
			Foreground(theme.ColorRed).
			Render("missed"),
		state.DashboardWindow.MissedCount,
		lipgloss.NewStyle().
			Foreground(theme.ColorYellow).
			Render("carry"),
		state.DashboardWindow.CarryOverCount,
	)
}

func renderProgressLine(theme types.Theme, state types.ContentState) string {
	if state.GoalProgress == nil {
		return "Progress  estimated -  worked -  status -"
	}
	status := "-"
	if len(state.GoalProgress.Rows) > 0 {
		status = strings.ReplaceAll(string(state.GoalProgress.Rows[0].Status), "_", " ")
	}
	return fmt.Sprintf(
		"%s  estimated %s  worked %s  status %s",
		theme.StyleHeader.Render("Progress"),
		theme.StyleNormal.Render(
			helperpkg.FormatCompactDurationMinutes(state.GoalProgress.TotalEstimateMinutes),
		),
		theme.StyleNormal.Render(
			viewhelpers.FormatClockText(state.GoalProgress.TotalActualSeconds),
		),
		progressStyle(theme, state.GoalProgress.EstimateBias, status).Render(status),
	)
}

func renderEstimateBiasLine(theme types.Theme, state types.ContentState) string {
	if state.GoalProgress == nil || state.GoalProgress.EstimatedItems == 0 {
		return "Estimate Bias  no estimated work in this range"
	}
	delta := fmt.Sprintf(
		"%+s",
		helperpkg.FormatCompactDurationMinutes(
			int(math.Round(state.GoalProgress.AverageDeltaMinutes)),
		),
	)
	percent := fmt.Sprintf("%+.0f%%", state.GoalProgress.AverageDeltaPercent)
	bias := strings.TrimSpace(state.GoalProgress.EstimateBias)
	if bias == "" {
		bias = "balanced"
	}
	return fmt.Sprintf("%s  avg %s  drift %s  bias %s  items %d",
		theme.StyleHeader.Render("Estimate Bias"),
		biasStyle(theme, bias).Render(delta),
		biasStyle(theme, bias).Render(percent),
		biasStyle(theme, bias).Render(bias),
		state.GoalProgress.EstimatedItems,
	)
}
