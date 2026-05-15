package wellbeing

import (
	"strings"

	uistate "crona/tui/internal/tui/state"
	viewchrome "crona/tui/internal/tui/views/chrome"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	types "crona/tui/internal/tui/views/types"
)

func PaneLineCount(state types.ContentState, pane string) int {
	switch pane {
	case string(uistate.PaneWellbeingSummary):
		if state.Height < 30 {
			return len(flattenLines(summaryBodyLines(types.Theme{}, state, state.Width, true)))
		}
		return len(flattenLines(summaryBodyLines(types.Theme{}, state, state.Width, state.Height < 37)))
	case string(uistate.PaneWellbeingTrends):
		if state.Height < 30 {
			return len(flattenLines(trendsBodyLines(types.Theme{}, state, state.Width, true)))
		}
		if state.Width >= 96 {
			return len(flattenLines(metricsBodyLines(types.Theme{}, state, state.Width, state.Height < 37)))
		}
		return len(flattenLines(trendsBodyLines(types.Theme{}, state, state.Width, state.Height < 37)))
	case string(uistate.PaneWellbeingStreaks):
		return len(flattenLines(streaksBodyLines(types.Theme{}, state, state.Width, state.Height < 37)))
	default:
		return 0
	}
}

func renderScrollablePane(theme types.Theme, active bool, width, height int, header []string, body []string, cursor int) string {
	lines := append([]string{}, header...)
	bodyLines := flattenLines(body)
	if len(bodyLines) == 0 {
		return viewchrome.RenderPaneBox(theme, active, width, height, viewhelpers.StringsJoin(lines))
	}
	inner := viewchrome.RemainingPaneHeight(height, lines)
	start, end := visibleBodyWindow(cursor, len(bodyLines), inner)
	if start > 0 {
		lines = append(lines, theme.StyleDim.Render("↑ more"))
	}
	lines = append(lines, bodyLines[start:end]...)
	if end < len(bodyLines) {
		lines = append(lines, theme.StyleDim.Render("↓ more"))
	}
	return viewchrome.RenderPaneBox(theme, active, width, height, viewhelpers.StringsJoin(lines))
}

func flattenLines(lines []string) []string {
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		out = append(out, strings.Split(line, "\n")...)
	}
	return out
}

func visibleBodyWindow(cursor, total, inner int) (int, int) {
	if total <= 0 {
		return 0, 0
	}
	if inner < 1 {
		inner = 1
	}
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
