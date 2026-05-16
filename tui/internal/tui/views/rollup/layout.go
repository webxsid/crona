package rollup

import (
	helperpkg "crona/tui/internal/tui/helpers"
	viewchrome "crona/tui/internal/tui/views/chrome"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
)

func renderView(theme types.Theme, state types.ContentState) string {
	headerH, detailH := viewhelpers.SplitVertical(state.Height, 10, 8, state.Height/3)
	return lipgloss.JoinVertical(
		lipgloss.Left,
		renderSummary(theme, state, state.Width, headerH),
		renderDetails(theme, state, state.Width, detailH),
	)
}

func renderDetails(theme types.Theme, state types.ContentState, width, height int) string {
	leftW, rightW := viewhelpers.SplitHorizontal(width, 34, 34, width/2)
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		renderWindowDays(theme, state, leftW, height),
		renderDistributionPane(theme, state, rightW, height),
	)
}

func renderWindowDays(theme types.Theme, state types.ContentState, width, height int) string {
	lines := []string{
		theme.StylePaneTitle.Render("Daily Status"),
		theme.StyleDim.Render("[enter] details"),
	}
	if state.DashboardWindow == nil || len(state.DashboardWindow.Days) == 0 {
		lines = append(lines, theme.StyleDim.Render("No daily rollup data for this range"))
		return viewchrome.RenderPaneBox(theme, state.View == "rollup", width, height, viewhelpers.StringsJoin(lines))
	}
	active := state.View == "rollup" && state.Pane == "rollup_days"
	cur := state.Cursors["rollup_days"]
	inner := viewchrome.RemainingPaneHeight(height, lines)
	start, end := viewchrome.ListWindow(cur, len(state.DashboardWindow.Days), inner)
	if start > 0 {
		lines = append(lines, theme.StyleDim.Render("  ↑ "+itoa(start)+" more"))
	}
	for idx := start; idx < end; idx++ {
		day := state.DashboardWindow.Days[idx]
		status := statusStyle(theme, string(day.Status)).Render(prettyWindowStatus(string(day.Status)))
		row := helperpkg.FormatDisplayDate(day.Date, state.Settings) + "  " + padStatus(status, 11) + "  " + "p" + itoa(day.PlannedCount) + " d" + itoa(day.CompletedCount) + " f" + itoa(day.FailedCount) + " c" + itoa(day.CarryOverCount)
		switch {
		case active && idx == cur:
			lines = append(lines, theme.StyleCursor.Render("▶ "+row))
		case idx == cur:
			lines = append(lines, theme.StyleSelected.Render("  "+row))
		default:
			lines = append(lines, "  "+row)
		}
	}
	if remaining := len(state.DashboardWindow.Days) - end; remaining > 0 {
		lines = append(lines, theme.StyleDim.Render("  ↓ "+itoa(remaining)+" more"))
	}
	return viewchrome.RenderPaneBox(theme, active, width, height, viewhelpers.StringsJoin(lines))
}

func renderDistributionPane(theme types.Theme, state types.ContentState, width, height int) string {
	lines := []string{
		theme.StylePaneTitle.Render("Breakdown"),
		theme.StyleDim.Render("Top time allocations"),
	}
	lines = append(lines, renderDistributionSection(theme, "Repos", state.RepoDistribution)...)
	lines = append(lines, renderDistributionSection(theme, "Streams", state.StreamDistribution)...)
	lines = append(lines, renderDistributionSection(theme, "Issues", state.IssueDistribution)...)
	lines = append(lines, renderDistributionSection(theme, "Segments", state.SegmentDistribution)...)
	return viewchrome.RenderPaneBox(theme, false, width, height, viewhelpers.StringsJoin(lines))
}
