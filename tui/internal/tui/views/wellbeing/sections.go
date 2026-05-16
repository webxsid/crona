package wellbeing

import (
	"fmt"

	helperpkg "crona/tui/internal/tui/helpers"
	uistate "crona/tui/internal/tui/state"
	viewcalendar "crona/tui/internal/tui/views/calendar"
	viewchrome "crona/tui/internal/tui/views/chrome"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
)

func renderSmallScreen(theme types.Theme, state types.ContentState) string {
	activePane := state.Pane
	if activePane == "" {
		activePane = string(uistate.PaneWellbeingSummary)
	}
	displayDate := helperpkg.FormatDisplayDate(state.WellbeingDate, state.Settings)
	header := []string{
		fmt.Sprintf("%s  %s", theme.StylePaneTitle.Render(smallScreenTitle(activePane)), theme.StyleHeader.Render(displayDate)),
		viewchrome.RenderActionLine(theme, state.Width-6, viewchrome.ContextualActions(theme, viewchrome.ActionsState{
			View:           state.View,
			Pane:           state.Pane,
			RestModeActive: state.RestModeActive,
			AwayModeActive: state.AwayModeActive,
		})),
	}
	switch activePane {
	case string(uistate.PaneWellbeingTrends):
		return renderScrollablePane(theme, true, state.Width, state.Height, header, trendsBodyLines(theme, state, state.Width, true), state.Cursors[string(uistate.PaneWellbeingTrends)])
	case string(uistate.PaneWellbeingStreaks):
		return renderScrollablePane(theme, true, state.Width, state.Height, header, streaksBodyLines(theme, state, state.Width, true), state.Cursors[string(uistate.PaneWellbeingStreaks)])
	default:
		return renderScrollablePane(theme, true, state.Width, state.Height, header, summaryBodyLines(theme, state, state.Width, true), state.Cursors[string(uistate.PaneWellbeingSummary)])
	}
}

func renderCompact(theme types.Theme, state types.ContentState) string {
	if state.Pane == string(uistate.PaneWellbeingStreaks) {
		return renderStreaks(theme, state, state.Width, state.Height, true)
	}
	topH := max(10, state.Height*11/20)
	topH = min(topH, state.Height-6)
	bottomH := max(6, state.Height-topH)
	return lipglossJoinCompact(theme, state, topH, bottomH)
}

func lipglossJoinCompact(theme types.Theme, state types.ContentState, topH, bottomH int) string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		renderCompactSummary(theme, state, state.Width, topH),
		renderCompactTrends(theme, state, state.Width, bottomH),
	)
}

func renderSplit(theme types.Theme, state types.ContentState) string {
	topH, bottomH := splitWellbeingHeights(state.Height)
	summary := renderSummary(theme, state, state.Width, topH)
	if state.Height < 37 {
		summary = renderCompactSummary(theme, state, state.Width, topH)
	}
	return lipgloss.JoinVertical(
		lipgloss.Left,
		summary,
		renderMetricsAndStreaks(theme, state, state.Width, bottomH, state.Height < 37),
	)
}

func splitWellbeingHeights(height int) (int, int) {
	if height < 37 {
		topH := max(10, height*11/20)
		if topH > height-6 {
			topH = height - 6
		}
		return topH, max(6, height-topH)
	}
	return viewhelpers.SplitVertical(height, 11, 8, height/2)
}

func renderSummary(theme types.Theme, state types.ContentState, width, height int) string {
	active := state.Pane == string(uistate.PaneWellbeingSummary)
	displayDate := helperpkg.FormatDisplayDate(state.WellbeingDate, state.Settings)
	actionLine := ""
	if active {
		actionLine = viewchrome.RenderActionLine(theme, width-6, viewchrome.ContextualActions(theme, viewchrome.ActionsState{
			View:           state.View,
			Pane:           state.Pane,
			RestModeActive: state.RestModeActive,
			AwayModeActive: state.AwayModeActive,
		}))
	}
	header := []string{
		theme.StylePaneTitle.Render("Wellbeing"),
		theme.StylePaneTitle.Render(fmt.Sprintf("date: %s", displayDate)),
		actionLine,
		"",
	}
	body := summaryBodyLines(theme, state, width, false)
	innerWidth := max(24, width-8)
	if viewcalendar.ShouldRender(innerWidth) {
		start := viewcalendar.ShiftDate(state.WellbeingDate, -6)
		calendarLines := viewcalendar.Render(theme, viewcalendar.Selection{
			AnchorDate: state.WellbeingDate,
			RangeStart: start,
			RangeEnd:   state.WellbeingDate,
			MaxLines:   max(4, min(len(body), height-len(header)-2)),
		})
		if len(calendarLines) > 0 {
			body = viewcalendar.MergeBeside(body, calendarLines, innerWidth, 3)
		}
	}
	return renderScrollablePane(theme, active, width, height, header, body, state.Cursors[string(uistate.PaneWellbeingSummary)])
}

func renderCompactSummary(theme types.Theme, state types.ContentState, width, height int) string {
	active := state.Pane == string(uistate.PaneWellbeingSummary)
	displayDate := helperpkg.FormatDisplayDate(state.WellbeingDate, state.Settings)
	actionLine := ""
	if active {
		actionLine = viewchrome.RenderActionLine(theme, width-6, viewchrome.ContextualActions(theme, viewchrome.ActionsState{
			View:           state.View,
			Pane:           state.Pane,
			RestModeActive: state.RestModeActive,
			AwayModeActive: state.AwayModeActive,
		}))
	}
	header := []string{
		fmt.Sprintf("%s  %s", theme.StylePaneTitle.Render("Wellbeing"), theme.StyleHeader.Render(displayDate)),
		actionLine,
	}
	return renderScrollablePane(theme, active, width, height, header, summaryBodyLines(theme, state, width, true), state.Cursors[string(uistate.PaneWellbeingSummary)])
}

func renderTrends(theme types.Theme, state types.ContentState, width, height int) string {
	active := state.Pane == string(uistate.PaneWellbeingTrends)
	header := []string{theme.StylePaneTitle.Render("Metrics Window")}
	if active {
		header = append(header, viewchrome.RenderActionLine(theme, width-6, viewchrome.ContextualActions(theme, viewchrome.ActionsState{
			View:           state.View,
			Pane:           state.Pane,
			RestModeActive: state.RestModeActive,
			AwayModeActive: state.AwayModeActive,
		})))
	}
	return renderScrollablePane(theme, active, width, height, header, trendsBodyLines(theme, state, width, false), state.Cursors[string(uistate.PaneWellbeingTrends)])
}

func renderMetrics(theme types.Theme, state types.ContentState, width, height int, compact bool) string {
	active := state.Pane == string(uistate.PaneWellbeingTrends)
	header := []string{theme.StylePaneTitle.Render("Metrics Window")}
	if active {
		header = append(header, viewchrome.RenderActionLine(theme, width-6, viewchrome.ContextualActions(theme, viewchrome.ActionsState{
			View:           state.View,
			Pane:           state.Pane,
			RestModeActive: state.RestModeActive,
			AwayModeActive: state.AwayModeActive,
		})))
	}
	return renderScrollablePane(theme, active, width, height, header, metricsBodyLines(theme, state, width, compact), state.Cursors[string(uistate.PaneWellbeingTrends)])
}

func renderStreaks(theme types.Theme, state types.ContentState, width, height int, compact bool) string {
	active := state.Pane == string(uistate.PaneWellbeingStreaks)
	header := []string{theme.StylePaneTitle.Render("Momentum")}
	if active {
		header = append(header, viewchrome.RenderActionLine(theme, width-6, viewchrome.ContextualActions(theme, viewchrome.ActionsState{
			View:           state.View,
			Pane:           state.Pane,
			RestModeActive: state.RestModeActive,
			AwayModeActive: state.AwayModeActive,
		})))
	}
	header = append(header, "")
	return renderScrollablePane(theme, active, width, height, header, streaksBodyLines(theme, state, width, compact), state.Cursors[string(uistate.PaneWellbeingStreaks)])
}

func renderMetricsAndStreaks(theme types.Theme, state types.ContentState, width, height int, compact bool) string {
	leftW, rightW := viewhelpers.SplitHorizontal(width, 42, 34, width*3/5)
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		renderMetrics(theme, state, leftW, height, compact),
		renderStreaks(theme, state, rightW, height, compact),
	)
}

func renderCompactTrends(theme types.Theme, state types.ContentState, width, height int) string {
	active := state.Pane == string(uistate.PaneWellbeingTrends)
	header := []string{theme.StylePaneTitle.Render("Metrics Window")}
	if active {
		header = append(header, viewchrome.RenderActionLine(theme, width-6, viewchrome.ContextualActions(theme, viewchrome.ActionsState{
			View:           state.View,
			Pane:           state.Pane,
			RestModeActive: state.RestModeActive,
			AwayModeActive: state.AwayModeActive,
		})))
	}
	return renderScrollablePane(theme, active, width, height, header, trendsBodyLines(theme, state, width, true), state.Cursors[string(uistate.PaneWellbeingTrends)])
}

func smallScreenTitle(pane string) string {
	switch pane {
	case string(uistate.PaneWellbeingTrends):
		return "Metrics Window"
	case string(uistate.PaneWellbeingStreaks):
		return "Momentum"
	default:
		return "Wellbeing"
	}
}
