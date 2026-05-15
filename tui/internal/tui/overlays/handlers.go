package overlays

import (
	"crona/tui/internal/api"
	dialogstate "crona/tui/internal/tui/dialogs/controller"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type State struct {
	HelpOpen          bool
	FilterEditing     bool
	DialogState       dialogstate.State
	SessionDetailOpen bool
	SessionDetailY    int
	SessionDetail     *api.SessionDetail

	ScratchpadOpen     bool
	ScratchpadMeta     *api.ScratchPad
	ScratchpadFilePath string
	ScratchpadRendered string
	ScratchpadViewport viewport.Model
	Scratchpads        []api.ScratchPad

	Cursor map[string]int
	Timer  *api.TimerState
}

type Deps struct {
	StopFilterEdit             func(*State)
	SessionDetailMaxOffset     func(State) int
	OpenAmendSessionDialog     func(*State, string, string)
	SessionCommit              func(*api.SessionDetail) string
	OpenEditor                 func(string) tea.Cmd
	OpenDefaultViewer          func(string) tea.Cmd
	SetStatus                  func(*State, string, bool) tea.Cmd
	AbandonSelectedIssue       func(*State) tea.Cmd
	FilteredIndexAtCursor      func(State, string) int
	SetActiveScratchpadByIndex func(*State, int)
	ListLen                    func(State, string) int
	OpenScratchpad             func(idx int) tea.Cmd
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
		deps.OpenAmendSessionDialog(&state, state.SessionDetail.ID, deps.SessionCommit(state.SessionDetail))
		return state, nil
	default:
		return state, nil
	}
}

func HandleScratchpad(state State, key tea.KeyMsg, deps Deps) (State, tea.Cmd) {
	switch key.String() {
	case "esc":
		state.ScratchpadOpen = false
		state.ScratchpadMeta = nil
		state.ScratchpadFilePath = ""
		state.ScratchpadRendered = ""
		return state, nil
	case "e":
		if state.ScratchpadFilePath == "" {
			return state, nil
		}
		return state, deps.OpenEditor(state.ScratchpadFilePath)
	case "o":
		if state.ScratchpadFilePath == "" {
			return state, nil
		}
		return state, deps.OpenDefaultViewer(state.ScratchpadFilePath)
	case "s":
		if state.Timer != nil && state.Timer.State != "idle" {
			return state, deps.SetStatus(&state, "End or stash the active session before changing issue status", true)
		}
		return state, nil
	case "A":
		if state.Timer == nil || state.Timer.State == "idle" {
			return state, nil
		}
		return state, deps.AbandonSelectedIssue(&state)
	case "left", "h":
		if state.Cursor["scratchpads"] <= 0 {
			return state, nil
		}
		state.Cursor["scratchpads"]--
		rawIdx := deps.FilteredIndexAtCursor(state, "scratchpads")
		if rawIdx >= 0 {
			deps.SetActiveScratchpadByIndex(&state, rawIdx)
			return state, deps.OpenScratchpad(rawIdx)
		}
		return state, nil
	case "right", "l":
		if state.Cursor["scratchpads"] >= deps.ListLen(state, "scratchpads")-1 {
			return state, nil
		}
		state.Cursor["scratchpads"]++
		rawIdx := deps.FilteredIndexAtCursor(state, "scratchpads")
		if rawIdx >= 0 {
			deps.SetActiveScratchpadByIndex(&state, rawIdx)
			return state, deps.OpenScratchpad(rawIdx)
		}
		return state, nil
	case "j", "down":
		state.ScratchpadViewport.ScrollDown(1)
		return state, nil
	case "k", "up":
		state.ScratchpadViewport.ScrollUp(1)
		return state, nil
	case "d", "ctrl+d":
		state.ScratchpadViewport.HalfPageDown()
		return state, nil
	case "u", "ctrl+u":
		state.ScratchpadViewport.HalfPageUp()
		return state, nil
	case "g":
		state.ScratchpadViewport.GotoTop()
		return state, nil
	case "G":
		state.ScratchpadViewport.GotoBottom()
		return state, nil
	default:
		return state, nil
	}
}
