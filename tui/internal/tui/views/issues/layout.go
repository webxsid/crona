package issues

import (
	issuecore "crona/tui/internal/tui/views/issuecore"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
)

func renderView(theme types.Theme, state types.ContentState) string {
	openIndices, completedIndices := issuecore.SplitDefaultIssueIndices(state.DefaultIssues, state.Filters["issues"], state.Settings)
	if state.Height < 37 {
		return renderCompactView(theme, state, openIndices, completedIndices)
	}
	summaryH := 9
	if state.Height < 44 {
		summaryH = 7
	}
	remainingHeight := max(8, state.Height-summaryH-1)
	primaryPreferred := remainingHeight / 2
	if state.DefaultIssueSection == "completed" {
		completedH, priorityH := viewhelpers.SplitVertical(remainingHeight, 6, 8, remainingHeight*2/3)
		if completedH > 6 {
			completedH--
			priorityH++
		}
		return lipgloss.JoinVertical(
			lipgloss.Left,
			renderSummary(theme, state, summaryH),
			renderIssuePane(theme, state, "Active Issues [1]", "Due work and open issues", openIndices, 0, true, priorityH, "No open issues match the current filter", false),
			renderIssuePane(theme, state, "Completed Issues [2]", "Done and abandoned, ready to revisit", completedIndices, len(openIndices), false, completedH, "No done or abandoned issues", true),
		)
	}
	priorityPreferred := remainingHeight * 2 / 3
	if priorityPreferred < primaryPreferred {
		priorityPreferred = primaryPreferred
	}
	priorityH, completedH := viewhelpers.SplitVertical(remainingHeight, 8, 6, priorityPreferred)
	if completedH > 6 {
		completedH--
		priorityH++
	}
	return lipgloss.JoinVertical(
		lipgloss.Left,
		renderSummary(theme, state, summaryH),
		renderIssuePane(theme, state, "Active Issues [1]", "Due work and open issues", openIndices, 0, true, priorityH, "No open issues match the current filter", state.DefaultIssueSection != "completed"),
		renderIssuePane(theme, state, "Completed Issues [2]", "Done and abandoned, ready to revisit", completedIndices, len(openIndices), false, completedH, "No done or abandoned issues", state.DefaultIssueSection == "completed"),
	)
}

func renderCompactView(theme types.Theme, state types.ContentState, openIndices, completedIndices []int) string {
	summaryH := 6
	footerH := 4
	mainH := max(8, state.Height-summaryH-footerH-1)
	mainTitle := "Active Issues [1]"
	mainSubtitle := "Due work and open issues"
	mainIndices := openIndices
	mainOffset := 0
	mainEmpty := "No open issues match the current filter"
	mainActive := state.DefaultIssueSection != "completed"
	footerTitle := "Closed"
	footerIndices := completedIndices

	if state.DefaultIssueSection == "completed" {
		mainTitle = "Completed Issues [2]"
		mainSubtitle = "Done and abandoned, ready to revisit"
		mainIndices = completedIndices
		mainOffset = len(openIndices)
		mainEmpty = "No done or abandoned issues"
		mainActive = true
		footerTitle = "Open"
		footerIndices = openIndices
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		renderCompactSummary(theme, state, summaryH),
		renderCompactIssuePane(theme, state, mainTitle, mainSubtitle, mainIndices, mainOffset, mainEmpty, mainActive, mainH),
		renderCompactFooter(theme, footerTitle, footerIndices, state, footerH),
	)
}
