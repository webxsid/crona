package wellbeing

import (
	uistate "crona/tui/internal/tui/state"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
)

func Render(theme types.Theme, state types.ContentState) string {
	if state.Height < 30 {
		return renderSmallScreen(theme, state)
	}
	if state.Width >= 96 {
		return renderSplit(theme, state)
	}
	if state.Pane == string(uistate.PaneWellbeingStreaks) {
		return renderStreaks(theme, state, state.Width, state.Height, state.Height < 37)
	}
	if state.Height < 37 {
		return renderCompact(theme, state)
	}
	topH, bottomH := viewhelpers.SplitVertical(state.Height, 11, 8, state.Height/2)
	return lipgloss.JoinVertical(
		lipgloss.Left,
		renderSummary(theme, state, state.Width, topH),
		renderTrends(theme, state, state.Width, bottomH),
	)
}
