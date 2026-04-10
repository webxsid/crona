package meta

import (
	viewchrome "crona/tui/internal/tui/views/chrome"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
)

func renderView(theme types.Theme, state types.ContentState) string {
	topH, botH := viewhelpers.SplitVertical(state.Height, 7, 8, state.Height*30/100)
	streamsEmpty := "No streams — [a] create new"
	if state.Context == nil || state.Context.RepoID == nil {
		streamsEmpty = "No repo checked out — [1] then [c]"
	}
	issuesEmpty := "No issues — [a] create new"
	if state.Context == nil || state.Context.StreamID == nil {
		issuesEmpty = "No stream checked out — [2] then [c]"
	}
	leftW, rightW := viewhelpers.SplitHorizontal(state.Width, 18, 18, state.Width/2)
	topRow := lipgloss.JoinHorizontal(lipgloss.Top,
		renderRepos(theme, state, leftW, topH),
		renderStreams(theme, state, rightW, topH, streamsEmpty),
	)
	leftBottom, rightBottom := viewhelpers.SplitHorizontal(state.Width, 24, 24, state.Width*3/5)
	bottomRow := lipgloss.JoinHorizontal(lipgloss.Top,
		renderIssues(theme, state, leftBottom, botH, issuesEmpty),
		renderHabits(theme, state, rightBottom, botH),
	)
	return lipgloss.JoinVertical(lipgloss.Left, topRow, bottomRow)
}

func paneActions(theme types.Theme, state types.ContentState, pane string) []string {
	return viewchrome.ContextualActions(theme, viewchrome.ActionsState{
		View:           state.View,
		Pane:           pane,
		ScratchpadOpen: state.ScratchpadOpen,
		TimerState:     timerStateFromContentState(state),
		RestModeActive: state.RestModeActive,
		AwayModeActive: state.AwayModeActive,
	})
}

func timerStateFromContentState(state types.ContentState) string {
	if state.Timer == nil {
		return ""
	}
	return state.Timer.State
}
