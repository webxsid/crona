package input

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	uistate "crona/tui/internal/tui/state"
)

func handleInstallUpdate(s State, deps Deps) (tea.Model, tea.Cmd, bool) {
	if s.ActiveView != uistate.ViewUpdates || !viewsShouldShowUpdate(s.UpdateStatus) || s.UpdateInstalling {
		return s, nil, false
	}
	if !deps.SelfUpdateInstallAvailable(s) {
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

func handlePauseSession(s State, deps Deps) (tea.Model, tea.Cmd, bool) {
	if s.ActiveView != uistate.ViewSessionActive || s.Timer == nil || s.Timer.State != "running" {
		return s, nil, false
	}
	return s, deps.PauseSession(), true
}

func handleResumeSession(s State, deps Deps) (tea.Model, tea.Cmd, bool) {
	if s.ActiveView != uistate.ViewSessionActive || s.Timer == nil || s.Timer.State != "paused" {
		return s, nil, false
	}
	return s, deps.ResumeSession(), true
}

func handleEndSession(s State, deps Deps) (tea.Model, tea.Cmd, bool) {
	if s.ActiveView != uistate.ViewSessionActive || s.Timer == nil || s.Timer.State == "idle" {
		return s, nil, false
	}
	deps.OpenEndSessionDialog(&s)
	return s, nil, true
}

func handleStashSession(s State, deps Deps) (tea.Model, tea.Cmd, bool) {
	if s.ActiveView != uistate.ViewSessionActive || s.Timer == nil || s.Timer.State == "idle" {
		return s, nil, false
	}
	deps.OpenStashSessionDialog(&s)
	return s, nil, true
}
