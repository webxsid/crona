package wellbeing

import (
	"fmt"

	helperpkg "crona/tui/internal/tui/helpers"
	uistate "crona/tui/internal/tui/state"
	viewchrome "crona/tui/internal/tui/views/chrome"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
)

func renderSmallScreen(theme types.Theme, state types.ContentState) string {
	activeSummary := state.Pane != string(uistate.PaneWellbeingTrends)
	displayDate := helperpkg.FormatDisplayDate(state.WellbeingDate, state.Settings)
	header := []string{
		fmt.Sprintf("%s  %s", theme.StylePaneTitle.Render(smallScreenTitle(activeSummary)), theme.StyleHeader.Render(displayDate)),
		viewchrome.RenderActionLine(theme, state.Width-6, viewchrome.ContextualActions(theme, viewchrome.ActionsState{
			View:           state.View,
			Pane:           state.Pane,
			RestModeActive: state.RestModeActive,
			AwayModeActive: state.AwayModeActive,
		})),
	}
	if activeSummary {
		return renderScrollablePane(theme, true, state.Width, state.Height, header, summaryBodyLines(theme, state, state.Width, true), state.Cursors[string(uistate.PaneWellbeingSummary)])
	}
	return renderScrollablePane(theme, true, state.Width, state.Height, header, trendsBodyLines(theme, state, state.Width, true), state.Cursors[string(uistate.PaneWellbeingTrends)])
}

func renderCompact(theme types.Theme, state types.ContentState) string {
	topH := max(10, state.Height*11/20)
	if topH > state.Height-6 {
		topH = state.Height - 6
	}
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
	return renderScrollablePane(theme, active, width, height, header, summaryBodyLines(theme, state, width, false), state.Cursors[string(uistate.PaneWellbeingSummary)])
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

func smallScreenTitle(summary bool) string {
	if summary {
		return "Wellbeing"
	}
	return "Metrics Window"
}
