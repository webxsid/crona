package input

import (
	sharedtypes "crona/shared/types"
	tea "github.com/charmbracelet/bubbletea"

	uistate "crona/tui/internal/tui/state"
)

func timerIsActive(s State) bool {
	return s.Timer != nil && s.Timer.State != "idle"
}

func handleStartAdhocTimer(s State, deps Deps) (tea.Model, tea.Cmd, bool) {
	if s.ProtectedModeActive || timerIsActive(s) ||
		(s.ActiveView != uistate.ViewSessionActive && !issueEditorContext(s)) ||
		deps.OpenAdhocTimerDialog == nil {
		return s, nil, false
	}
	handled := deps.OpenAdhocTimerDialog(&s)
	return s, nil, handled
}

func handleAdvanceSession(s State, deps Deps) (tea.Model, tea.Cmd, bool) {
	if s.ActiveView != uistate.ViewSessionActive || s.Timer == nil ||
		s.Timer.State == "idle" || !s.Timer.HardLimitActive ||
		sharedtypes.NormalizeTimerHardLimitKind(s.Timer.HardLimitKind) !=
			sharedtypes.TimerHardLimitKindPomodoro {
		return s, nil, false
	}
	return s, deps.AdvanceSession(), true
}

func issueEditorContext(s State) bool {
	switch s.ActiveView {
	case uistate.ViewDefault, uistate.ViewDaily:
		return s.ActivePane == uistate.PaneIssues
	case uistate.ViewMomentum:
		return s.MomentumTab == "custom"
	default:
		return false
	}
}

func focusStartContext(s State) bool {
	if s.ProtectedModeActive || timerIsActive(s) {
		return false
	}
	return issueEditorContext(s)
}

func handlePauseSession(s State, deps Deps) (tea.Model, tea.Cmd, bool) {
	if s.ActiveView != uistate.ViewSessionActive || s.Timer == nil || s.Timer.State != "running" ||
		s.Timer.HardLimitActive {
		return s, nil, false
	}
	return s, deps.PauseSession(), true
}

func handleResumeSession(s State, deps Deps) (tea.Model, tea.Cmd, bool) {
	if s.ActiveView != uistate.ViewSessionActive || s.Timer == nil {
		return s, nil, false
	}
	if s.Timer.State == "ready" {
		return s, deps.ResumeSession(s), true
	}
	if s.Timer.HardLimitActive {
		return s, nil, false
	}
	if s.Timer.NextSegmentType != nil {
		return s, deps.ResumeSession(s), true
	}
	if s.Timer.State != "paused" {
		return s, nil, false
	}
	return s, deps.ResumeSession(s), true
}

func handleStructuredManualPause(s State, deps Deps) (tea.Model, tea.Cmd, bool) {
	if s.ActiveView != uistate.ViewSessionActive || s.Timer == nil {
		return s, nil, false
	}
	if s.Timer.HardLimitActive {
		return s, nil, true
	}
	if s.Timer.State != "idle" {
		return s, nil, true
	}
	return s, nil, false
}

func handleEndSession(s State, deps Deps) (tea.Model, tea.Cmd, bool) {
	if s.ActiveView != uistate.ViewSessionActive || s.Timer == nil || s.Timer.State == "idle" {
		return s, nil, false
	}
	deps.OpenEndSessionDialog(&s)
	return s, nil, true
}
