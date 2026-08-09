package away

import (
	"fmt"
	"strings"

	helperpkg "crona/tui/internal/tui/helpers"
	viewtypes "crona/tui/internal/tui/views/types"
	viewui "crona/tui/internal/tui/views/ui"

	"github.com/charmbracelet/lipgloss"
)

func Render(theme viewtypes.Theme, state viewtypes.ContentState) string {
	return viewui.NewLayout(theme, state, newView().Render).RenderView()
}

func RenderHistorical(theme viewtypes.Theme, state viewtypes.ContentState) string {
	date := helperpkg.FormatDisplayDate(state.HistoricalAwayDate, state.Settings)
	lines := []string{theme.StylePaneTitle.Render("Away Mode"), ""}
	body := fmt.Sprintf("You chose to rest and recover on %s.", date)
	content := lipgloss.Place(
		max(1, state.Width-6),
		max(1, state.Height-6),
		lipgloss.Center,
		lipgloss.Center,
		theme.StyleHeader.Render(body),
	)
	lines = append(lines, strings.TrimRight(content, "\n"))
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(theme.ColorDim).
		Padding(1, 2).
		Width(max(1, state.Width-2)).
		Height(max(1, state.Height-2)).
		Render(strings.Join(lines, "\n"))
}
