package input

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	commands "crona/tui/internal/tui/commands"
	uistate "crona/tui/internal/tui/state"
)

func handleInstallUpdate(s State, deps Deps) (tea.Model, tea.Cmd, bool) {
	if s.ActiveView != uistate.ViewUpdates || !viewsShouldShowUpdate(s.UpdateStatus) ||
		s.UpdateInstalling {
		return s, nil, false
	}
	if !deps.SelfUpdateInstallAvailable(s) {
		if s.UpdateStatus != nil && strings.TrimSpace(s.UpdateStatus.UpdateCommand) != "" {
			return s, commands.CopyTextToClipboard(
				strings.TrimSpace(s.UpdateStatus.UpdateCommand),
				"Update command copied",
			), true
		}
		reason := strings.TrimSpace(deps.SelfUpdateUnsupportedReason(s))
		if reason == "" && s.UpdateStatus != nil {
			reason = strings.TrimSpace(s.UpdateStatus.InstallUnavailableReason)
		}
		if reason == "" {
			reason = "Please update manually."
		}
		return s, deps.SetStatus(&s, reason, true), true
	}
	s.UpdateInstalling = true
	s.UpdateInstallError = ""
	return s, deps.InstallUpdate(s), true
}

func handleDismissUpdate(s State, deps Deps) (tea.Model, tea.Cmd, bool) {
	if !viewsShouldShowUpdate(s.UpdateStatus) {
		return s, nil, false
	}
	return s, deps.DismissUpdate(), true
}

func timerIsActive(s State) bool {
	return s.Timer != nil && s.Timer.State != "idle"
}

func issueEditorContext(s State) bool {
	switch s.ActiveView {
	case uistate.ViewDefault, uistate.ViewDaily, uistate.ViewMeta:
		return s.ActivePane == uistate.PaneIssues
	default:
		return false
	}
}

func focusStartContext(s State) bool {
	if s.ProtectedModeActive || timerIsActive(s) {
		return false
	}
	if issueEditorContext(s) {
		return true
	}
	return s.ActiveView == uistate.ViewSessionActive &&
		(s.Timer == nil || s.Timer.State == "idle")
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
