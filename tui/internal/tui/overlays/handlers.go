package overlays

import (
	"crona/tui/internal/api"
	dialogstate "crona/tui/internal/tui/dialogs/controller"

	tea "github.com/charmbracelet/bubbletea"
)

type State struct {
	HelpOpen          bool
	FilterEditing     bool
	DialogState       dialogstate.State
	SessionDetailOpen bool
	SessionDetailY    int
	SessionDetail     *api.SessionDetail

	Cursor map[string]int
	Timer  *api.TimerState
}

type Deps struct {
	StopFilterEdit         func(*State)
	SessionDetailMaxOffset func(State) int
	OpenAmendSessionDialog func(*State, string, string)
	SessionCommit          func(*api.SessionDetail) string
	SetStatus              func(*State, string, bool) tea.Cmd
	AbandonSelectedIssue   func(*State) tea.Cmd
	FilteredIndexAtCursor  func(State, string) int
	ListLen                func(State, string) int
}

func HandleHelp(state State, key tea.KeyMsg) (State, tea.Cmd) {
	switch key.String() {
	case "?", "esc", "q":
		state.HelpOpen = false
		return state, nil
	default:
		return state, nil
	}
}

func HandleFilter(state State, key tea.KeyMsg, deps Deps) (State, tea.Cmd) {
	switch key.String() {
	case "esc", "enter":
		deps.StopFilterEdit(&state)
		return state, nil
	default:
		return state, nil
	}
}

func HandleSessionDetail(state State, key tea.KeyMsg, deps Deps) (State, tea.Cmd) {
	switch key.String() {
	case "esc", "q", "o", "enter":
		state.SessionDetailOpen = false
		state.SessionDetail = nil
		state.SessionDetailY = 0
		return state, nil
	case "j", "down":
		if state.SessionDetailY < deps.SessionDetailMaxOffset(state) {
			state.SessionDetailY++
		}
		return state, nil
	case "k", "up":
		if state.SessionDetailY > 0 {
			state.SessionDetailY--
		}
		return state, nil
	case "e":
		if state.SessionDetail == nil {
			return state, nil
		}
		deps.OpenAmendSessionDialog(
			&state,
			state.SessionDetail.ID,
			deps.SessionCommit(state.SessionDetail),
		)
		return state, nil
	default:
		return state, nil
	}
}
