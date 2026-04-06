package wellbeing

import (
	viewhelpers "crona/tui/internal/tui/views/helpers"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
)

func Render(theme types.Theme, state types.ContentState) string {
	if state.Height < 30 {
		return renderSmallScreen(theme, state)
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
