package controller

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestFormFocusIgnoresUpDown(t *testing.T) {
	openers := map[string]func() State{
		"create_issue_default": func() State { return OpenCreateIssueDefault(State{}) },
		"checkout_context":     func() State { return OpenCheckoutContext(State{}) },
		"create_repo":          func() State { return OpenCreateRepo(State{}) },
		"create_habit":         func() State { return OpenCreateHabit(State{}) },
		"edit_issue":           func() State { return OpenEditIssue(State{}, 1, 2, "Fix", nil, nil, nil) },
	}
	for name, open := range openers {
		for _, key := range []tea.KeyType{tea.KeyUp, tea.KeyDown} {
			state := open()
			focus, repo := state.FocusIdx, state.RepoIndex
			next, _, _ := Update(state, UpdateContext{}, "2026-10-18", tea.KeyMsg{Type: key})
			if next.FocusIdx != focus || next.RepoIndex != repo {
				t.Errorf("%s %v: focus %d→%d repo %d→%d", name, key, focus, next.FocusIdx, repo, next.RepoIndex)
			}
		}
		state := open()
		next, _, _ := Update(state, UpdateContext{}, "2026-10-18", tea.KeyMsg{Type: tea.KeyTab})
		if next.FocusIdx == state.FocusIdx {
			t.Errorf("%s: tab did not move focus", name)
		}
		back, _, _ := Update(next, UpdateContext{}, "2026-10-18", tea.KeyMsg{Type: tea.KeyShiftTab})
		if back.FocusIdx != state.FocusIdx {
			t.Errorf("%s: shift+tab did not return focus", name)
		}
	}
}
