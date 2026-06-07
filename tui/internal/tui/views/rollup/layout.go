package rollup

import (
	"strings"

	helperpkg "crona/tui/internal/tui/helpers"
	viewchrome "crona/tui/internal/tui/views/chrome"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
)

func renderView(theme types.Theme, state types.ContentState) string {
	headerH, detailH := viewhelpers.SplitVertical(state.Height, 12, 8, state.Height/2)
	return lipgloss.JoinVertical(
		lipgloss.Left,
		renderSummary(theme, state, state.Width, headerH),
		renderDetails(theme, state, state.Width, detailH),
	)
}

func PaneLineCount(state types.ContentState, pane string) int {
	switch pane {
	case "rollup_breakdown":
		return len(rollupFlattenLines(rollupBreakdownBodyLines(types.Theme{}, state)))
	default:
		return 0
	}
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
		return viewchrome.RenderPaneBox(
			theme,
			state.View == "rollup",
			width,
			height,
			viewhelpers.StringsJoin(lines),
		)
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
		status := statusStyle(
			theme,
			string(day.Status),
		).Render(prettyWindowStatus(string(day.Status)))
		row := helperpkg.FormatDisplayDate(
			day.Date,
			state.Settings,
		) + "  " + padStatus(
			status,
			11,
		) + "  " + "p" + itoa(
			day.PlannedCount,
		) + " d" + itoa(
			day.CompletedCount,
		) + " f" + itoa(
			day.FailedCount,
		) + " c" + itoa(
			day.CarryOverCount,
		)
		switch {
		case active && idx == cur:
			lines = append(lines, theme.StyleCursor.Render(viewchrome.SelectionCursor+" "+row))
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
	header := []string{
		theme.StylePaneTitle.Render("Breakdown"),
		theme.StyleDim.Render("summary and top time allocations"),
		"",
	}
	body := rollupBreakdownBodyLines(theme, state)
	active := state.View == "rollup" && state.Pane == "rollup_breakdown"
	return renderRollupScrollablePane(
		theme,
		active,
		width,
		height,
		header,
		body,
		state.Cursors["rollup_breakdown"],
	)
}

func rollupBreakdownBodyLines(theme types.Theme, state types.ContentState) []string {
	lines := []string{
		renderWindowLine(theme, state),
		renderProgressLine(theme, state),
		renderEstimateBiasLine(theme, state),
		"",
	}
	lines = append(lines, renderDistributionSection(theme, "Repos", state.RepoDistribution)...)
	lines = append(lines, renderDistributionSection(theme, "Streams", state.StreamDistribution)...)
	lines = append(lines, renderDistributionSection(theme, "Issues", state.IssueDistribution)...)
	lines = append(lines, renderDistributionSection(theme, "Segments", state.SegmentDistribution)...)
	return lines
}

func renderRollupScrollablePane(
	theme types.Theme,
	active bool,
	width, height int,
	header []string,
	body []string,
	cursor int,
) string {
	lines := append([]string{}, header...)
	bodyLines := rollupFlattenLines(body)
	if len(bodyLines) == 0 {
		return viewchrome.RenderPaneBox(theme, active, width, height, viewhelpers.StringsJoin(lines))
	}
	inner := viewchrome.RemainingPaneHeight(height, lines)
	start, end := rollupVisibleBodyWindow(cursor, len(bodyLines), inner)
	if start > 0 {
		lines = append(lines, theme.StyleDim.Render("↑ more"))
	}
	lines = append(lines, bodyLines[start:end]...)
	if end < len(bodyLines) {
		lines = append(lines, theme.StyleDim.Render("↓ more"))
	}
	return viewchrome.RenderPaneBox(theme, active, width, height, viewhelpers.StringsJoin(lines))
}

func rollupFlattenLines(lines []string) []string {
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		out = append(out, strings.Split(line, "\n")...)
	}
	return out
}

func rollupVisibleBodyWindow(cursor, total, inner int) (int, int) {
	if total <= 0 {
		return 0, 0
	}
	inner = max(0, inner)
	start, end := viewchrome.ListWindow(cursor, total, inner)
	for {
		used := end - start
		if start > 0 {
			used++
		}
		if end < total {
			used++
		}
		if used <= inner || end-start <= 1 {
			return start, end
		}
		if end < total {
			end--
			continue
		}
		if start > 0 {
			start++
			continue
		}
		return start, end
	}
}
