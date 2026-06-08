package input

import (
	"testing"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	uistate "crona/tui/internal/tui/state"

	tea "github.com/charmbracelet/bubbletea"
)

func TestHandleResumeSessionAllowsReadyHardLimitAdvance(t *testing.T) {
	work := sharedtypes.SessionSegmentWork
	called := false
	state := State{
		ActiveView: uistate.ViewSessionActive,
		Timer: &api.TimerState{
			State:            "ready",
			HardLimitActive:  true,
			ReadySegmentType: &work,
		},
	}

	_, _, handled := handleResumeSession(state, Deps{
		ResumeSession: func(State) tea.Cmd {
			called = true
			return nil
		},
	})
	if !handled {
		t.Fatal("expected ready hard-limit session to handle resume as advance")
	}
	if !called {
		t.Fatal("expected ready hard-limit session to invoke resume/advance command")
	}
}

func TestHandleResumeSessionRejectsRunningHardLimit(t *testing.T) {
	work := sharedtypes.SessionSegmentWork
	called := false
	state := State{
		ActiveView: uistate.ViewSessionActive,
		Timer: &api.TimerState{
			State:           "running",
			HardLimitActive: true,
			SegmentType:     &work,
		},
	}

	_, _, handled := handleResumeSession(state, Deps{
		ResumeSession: func(State) tea.Cmd {
			called = true
			return nil
		},
	})
	if handled {
		t.Fatal("expected running hard-limit session to reject resume")
	}
	if called {
		t.Fatal("expected running hard-limit session to not invoke resume command")
	}
}

func TestHandleStructuredManualPauseBlocksManualLogDuringActiveSession(t *testing.T) {
	work := sharedtypes.SessionSegmentWork
	state := State{
		ActiveView: uistate.ViewSessionActive,
		Timer: &api.TimerState{
			State:           "running",
			SegmentType:     &work,
			IssueID:         int64Ptr(1),
			HardLimitActive: false,
		},
	}

	_, _, handled := handleStructuredManualPause(state, Deps{
		OpenManualSessionDialog: func(*State) (tea.Cmd, bool) {
			t.Fatal("did not expect manual session dialog to open")
			return nil, false
		},
	})
	if !handled {
		t.Fatal("expected active session m to be consumed")
	}
}

func TestHandleStructuredManualPauseAllowsManualLogWhenIdle(t *testing.T) {
	state := State{
		ActiveView: uistate.ViewSessionActive,
		Timer:      &api.TimerState{State: "idle"},
	}

	_, _, handled := handleStructuredManualPause(state, Deps{
		OpenManualSessionDialog: func(*State) (tea.Cmd, bool) {
			return nil, true
		},
	})
	if handled {
		t.Fatal("expected idle session view to fall through to manual log")
	}
}

func TestHandleStashShortcutIsNoOpInSessionActiveView(t *testing.T) {
	state := State{
		ActiveView: uistate.ViewSessionActive,
		Timer: &api.TimerState{
			State: "running",
		},
	}

	next, cmd := Handle(state, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'z'}}, Deps{
		OpenStashSessionDialog: func(*State) bool {
			t.Fatal("did not expect stash dialog to open")
			return true
		},
	})
	if cmd != nil {
		t.Fatalf("expected no command for stash shortcut, got %+v", cmd)
	}
	if next.Dialog != "" {
		t.Fatalf("expected no dialog change for stash shortcut, got %+v", next)
	}
}

func int64Ptr(v int64) *int64 { return &v }
