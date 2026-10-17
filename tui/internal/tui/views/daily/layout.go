package daily

import (
	viewaway "crona/tui/internal/tui/views/away"
	viewchrome "crona/tui/internal/tui/views/chrome"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
)

func renderView(theme types.Theme, state types.ContentState) string {
	if state.HistoricalAwayDate != "" {
		return viewaway.RenderHistorical(theme, state)
	}
	width, height := state.Width, state.Height
	switch viewhelpers.ClassFor(width) {
	case viewhelpers.WidthWide:
		cols := viewhelpers.Columns(width, 3, 2, 2)
		return viewchrome.JoinColumns(theme, height,
			renderIssues(theme, state, cols[0], height),
			renderHabits(theme, state, cols[1], height),
			renderSummary(theme, state, cols[2], height),
		)
	case viewhelpers.WidthMedium:
		sections, listH := daySection(theme, state, width, height)
		cols := viewhelpers.Columns(width, 3, 2)
		sections = append(sections, viewchrome.JoinColumns(theme, listH,
			renderIssues(theme, state, cols[0], listH),
			renderHabits(theme, state, cols[1], listH),
		))
		return lipgloss.JoinVertical(lipgloss.Left, sections...)
	default:
		sections, listH := daySection(theme, state, width, height)
		issuesH := listH * 3 / 5
		sections = append(sections,
			renderIssues(theme, state, width, issuesH),
			renderHabits(theme, state, width, listH-issuesH),
		)
		return lipgloss.JoinVertical(lipgloss.Left, sections...)
	}
}

// daySection renders the Day summary above the lists when there is room, returning the rows left.
func daySection(theme types.Theme, state types.ContentState, width, height int) ([]string, int) {
	summaryH, listH := viewhelpers.SplitVertical(height, 10, 8, height*2/5)
	if summaryH < 3 {
		return nil, height
	}
	return []string{renderSummary(theme, state, width, summaryH)}, listH
}
