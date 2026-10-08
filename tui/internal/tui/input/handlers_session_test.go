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

func TestHandleStartAdhocTimerOnlyHandlesIdleSessionView(t *testing.T) {
	for _, test := range []struct {
		name  string
		state State
		want  bool
	}{
		{name: "idle session", state: State{ActiveView: uistate.ViewSessionActive}, want: true},
		{name: "idle issue pane", state: State{ActiveView: uistate.ViewDaily, ActivePane: uistate.PaneIssues}, want: true},
		{name: "running session", state: State{ActiveView: uistate.ViewSessionActive, Timer: &api.TimerState{State: "running"}}},
		{name: "issue view", state: State{ActiveView: uistate.ViewDaily}},
	} {
		t.Run(test.name, func(t *testing.T) {
			called := false
			_, _, handled := handleStartAdhocTimer(test.state, Deps{
				OpenAdhocTimerDialog: func(*State) bool {
					called = true
					return true
				},
			})
			if handled != test.want || called != test.want {
				t.Fatalf("handled=%t callback_called=%t, want %t", handled, called, test.want)
			}
		})
	}
}

func TestAdhocShortcutReturnsDialogStateFromCallback(t *testing.T) {
	state := State{ActiveView: uistate.ViewDaily, ActivePane: uistate.PaneIssues}
	next, _ := Handle(state, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'A'}}, Deps{
		OpenAdhocTimerDialog: func(state *State) bool {
			state.Dialog = "timer_start_type"
			return true
		},
	})
	if next.Dialog != "timer_start_type" {
		t.Fatalf("dialog after A = %q, want timer_start_type", next.Dialog)
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
