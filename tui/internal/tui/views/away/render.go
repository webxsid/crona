package away

import (
	"strings"

	viewchrome "crona/tui/internal/tui/views/chrome"
	viewtypes "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
)

func Render(theme viewtypes.Theme, state viewtypes.ContentState) string {
	actions := viewchrome.ContextualActions(theme, viewchrome.ActionsState{
		View:           state.View,
		Pane:           state.Pane,
		RestModeActive: state.RestModeActive,
		AwayModeActive: state.AwayModeActive,
	})

	lines := []string{theme.StylePaneTitle.Render("Away")}
	if len(actions) > 0 {
		lines = append(lines, viewchrome.RenderActionLine(theme, state.Width-6, actions))
	}
	lines = append(lines, "")

	body := []string{theme.StyleHeader.Render(state.RestModeMessage)}
	if strings.TrimSpace(state.RestModeDetail) != "" {
		body = append(body, theme.StyleDim.Render(state.RestModeDetail))
	}

	content := lipgloss.Place(
		max(1, state.Width-6),
		max(1, state.Height-6-viewchrome.RenderedLineCount(lines)),
		lipgloss.Center,
		lipgloss.Center,
		strings.Join(body, "\n\n"),
	)
	lines = append(lines, content)

	return viewchrome.RenderPaneBox(theme, false, state.Width, state.Height, strings.Join(lines, "\n"))
}
