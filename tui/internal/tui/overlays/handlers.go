package overlays

import (
	"crona/tui/internal/api"
	dialogstate "crona/tui/internal/tui/dialogs/controller"

	tea "github.com/charmbracelet/bubbletea"
)

type State struct {
	HelpOpen          bool
	DialogState       dialogstate.State
	SessionDetailOpen bool
	SessionDetailY    int
	SessionDetail     *api.SessionDetail
	ManageHistory     bool

	Cursor map[string]int
	Timer  *api.TimerState
}

type Deps struct {
	SessionDetailMaxOffset   func(State) int
	OpenAmendSessionDialog   func(*State, string, string)
	OpenEditSessionHistory   func(*State, *api.SessionDetail)
	OpenDeleteSessionHistory func(*State, string)
	SessionCommit            func(*api.SessionDetail) string
	SetStatus                func(*State, string, bool) tea.Cmd
	AbandonSelectedIssue     func(*State) tea.Cmd
	FilteredIndexAtCursor    func(State, string) int
	ListLen                  func(State, string) int
}

func HandleHelp(state State, key tea.KeyMsg) (State, tea.Cmd) {
	switch key.String() {
	case "?", "esc":
		state.HelpOpen = false
		return state, nil
	default:
		return state, nil
	}
}

func HandleSessionDetail(state State, key tea.KeyMsg, deps Deps) (State, tea.Cmd) {
	switch key.String() {
	case "esc", "o", "enter":
		state.SessionDetailOpen = false
		state.SessionDetail = nil
		state.SessionDetailY = 0
		return state, nil
	case "down":
		if state.SessionDetailY < deps.SessionDetailMaxOffset(state) {
			state.SessionDetailY++
		}
		return state, nil
	case "up":
		if state.SessionDetailY > 0 {
			state.SessionDetailY--
		}
		return state, nil
	case "e":
		if !state.ManageHistory || state.SessionDetail == nil || deps.OpenEditSessionHistory == nil {
			return state, nil
		}
		deps.OpenEditSessionHistory(&state, state.SessionDetail)
		return state, nil
	case "D":
		if !state.ManageHistory || state.SessionDetail == nil || deps.OpenDeleteSessionHistory == nil {
			return state, nil
		}
		deps.OpenDeleteSessionHistory(&state, state.SessionDetail.ID)
		return state, nil
	default:
		return state, nil
	}
}
