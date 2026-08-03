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

func TestHandleAdvanceSessionAllowsPomodoroStates(t *testing.T) {
	for _, timerState := range []string{"running", "ready", "paused"} {
		called := false
		state := State{
			ActiveView: uistate.ViewSessionActive,
			Timer: &api.TimerState{
				State:           timerState,
				HardLimitActive: true,
				HardLimitKind:   sharedtypes.TimerHardLimitKindPomodoro,
			},
		}
		_, _, handled := handleAdvanceSession(state, Deps{
			AdvanceSession: func() tea.Cmd {
				called = true
				return nil
			},
		})
		if !handled || !called {
			t.Fatalf("state %q: expected z to advance Pomodoro session", timerState)
		}
	}
}

func TestHandleAdvanceSessionRejectsNonPomodoroSessions(t *testing.T) {
	for _, timer := range []*api.TimerState{
		{State: "idle"},
		{State: "running"},
		{State: "running", HardLimitActive: true, HardLimitKind: sharedtypes.TimerHardLimitKindCountdown},
	} {
		_, _, handled := handleAdvanceSession(State{
			ActiveView: uistate.ViewSessionActive,
			Timer:      timer,
		}, Deps{AdvanceSession: func() tea.Cmd { t.Fatal("unexpected advance"); return nil }})
		if handled {
			t.Fatalf("timer %+v: expected z to be ignored", timer)
		}
	}
}

func TestHandleStructuredManualPauseBlocksManualLogDuringActiveSession(t *testing.T) {
	work := sharedtypes.SessionSegmentWork
	state := State{
		ActiveView: uistate.ViewSessionActive,
		Timer: &api.TimerState{
			State:           "running",
			SegmentType:     &work,
			IssueID:         new(int64(1)),
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
