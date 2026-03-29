package filtering

import (
	"crona/tui/internal/tui/chrome"
	helperpkg "crona/tui/internal/tui/helpers"
	uistate "crona/tui/internal/tui/state"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type State struct {
	Filters       map[uistate.Pane]string
	Cursor        map[uistate.Pane]int
	FilterEditing bool
	FilterPane    uistate.Pane
	FilterInput   textinput.Model
}

type Deps struct {
	ItemCount func(State, uistate.Pane) int
	Clamp     func(map[uistate.Pane]int, uistate.Pane, int)
}

func Start(state State, pane uistate.Pane) State {
	input := textinput.New()
	input.Placeholder = "filter..."
	input.SetValue(state.Filters[pane])
	input.CursorEnd()
	input.Focus()
	input.CharLimit = 120
	input.Width = 24

	state.FilterEditing = true
	state.FilterPane = pane
	state.FilterInput = input
	return state
}

func Stop(state State) State {
	state.FilterEditing = false
	state.FilterPane = ""
	state.FilterInput.Blur()
	return state
}

func Update(state State, msg tea.KeyMsg, deps Deps) (State, tea.Cmd) {
	var cmd tea.Cmd
	state.FilterInput, cmd = state.FilterInput.Update(msg)
	state.Filters[state.FilterPane] = state.FilterInput.Value()
	state.Cursor[state.FilterPane] = 0
	if deps.Clamp != nil && deps.ItemCount != nil {
		deps.Clamp(state.Cursor, state.FilterPane, deps.ItemCount(state, state.FilterPane))
	}
	return state, cmd
}

func RenderLine(state State, pane uistate.Pane, width int) string {
	if state.FilterEditing && state.FilterPane == pane {
		value := state.FilterInput.View()
		return chrome.StyleDim.Render("filter: ") + helperpkg.Truncate(value, width-8)
	}

	query := state.Filters[pane]
	if query == "" {
		return chrome.StyleDim.Render("filter: /")
	}
	return chrome.StyleDim.Render("filter: ") + helperpkg.Truncate(query, width-8)
}
