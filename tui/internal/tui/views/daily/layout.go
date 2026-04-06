package daily

import (
	viewhelpers "crona/tui/internal/tui/views/helpers"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
)

func renderView(theme types.Theme, state types.ContentState) string {
	summaryH, listH := viewhelpers.SplitVertical(state.Height, 10, 8, state.Height/3)
	sections := []string{}
	if summaryH >= 3 {
		sections = append(sections, renderSummary(theme, state, state.Width, summaryH))
	}
	if state.Width < 56 {
		issuesH, habitsH := viewhelpers.SplitVertical(listH, 8, 8, listH/2)
		sections = append(sections,
			renderIssues(theme, state, state.Width, issuesH),
			renderHabits(theme, state, state.Width, habitsH),
		)
		return lipgloss.JoinVertical(lipgloss.Left, sections...)
	}
	leftW, rightW := viewhelpers.SplitHorizontal(state.Width, 24, 24, state.Width*3/5)
	lists := lipgloss.JoinHorizontal(lipgloss.Top,
		renderIssues(theme, state, leftW, listH),
		renderHabits(theme, state, rightW, listH),
	)
	sections = append(sections, lists)
	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}
