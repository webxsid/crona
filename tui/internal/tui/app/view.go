package app

import (
	"strings"
	"time"

	helperpkg "crona/tui/internal/tui/helpers"
	layoutpkg "crona/tui/internal/tui/layout"
	"crona/tui/internal/tui/views"
)

func (m Model) View() string {
	return layoutpkg.Render(m.layoutState())
}

func (m Model) layoutState() layoutpkg.State {
	repo := "-"
	stream := "-"
	if m.context != nil {
		if m.context.RepoName != nil && strings.TrimSpace(*m.context.RepoName) != "" {
			repo = strings.TrimSpace(*m.context.RepoName)
		}
		if m.context.StreamName != nil && strings.TrimSpace(*m.context.StreamName) != "" {
			stream = strings.TrimSpace(*m.context.StreamName)
		}
	}
	timerState := ""
	if m.timer != nil {
		timerState = m.timer.State
	}
	protectedMode, _, _ := views.ProtectedRestMode(m.settings, time.Now().Format("2006-01-02"))
	state := layoutpkg.State{
		Width:         m.width,
		Height:        m.height,
		View:          m.view,
		Pane:          m.pane,
		ProtectedMode: protectedMode,
		IsDevMode:     m.isDevMode(),
		RepoName:      repo,
		StreamName:    stream,
		TimerActive:   m.timer != nil && m.timer.State != "idle",
		HeaderState: views.HeaderState{
			Width:         m.width,
			View:          string(m.view),
			Elapsed:       m.elapsed,
			Timer:         m.timer,
			IssueSessions: m.issueSessions,
			AllIssues:     m.allIssues,
			Health:        m.health,
			UpdateStatus:  m.updateStatus,
		},
		DialogOpen:         m.dialog != "",
		UpdateInstallOpen:  m.updateInstallPhase != "",
		HelpOpen:           m.helpOpen,
		SessionDetailOpen:  m.sessionDetailOpen,
		SessionDetailY:     m.sessionDetailY,
		SessionDetailLines: helperpkg.SessionDetailContentLines(m.sessionDetail),
		StatusMsg:          m.statusMsg,
		StatusErr:          m.statusErr,
		GlobalActions:      globalActions(m, timerState),
	}
	contentWidth := max(0, m.width-sidebarWidth(m.width))
	state.ContentState = m.viewContentState(contentWidth, layoutpkg.ContentHeight(state))
	if state.ContentState.RestModeActive {
		if m.view != ViewReports && m.view != ViewSessionHistory {
			state.View = ViewAway
			state.ContentState.View = "away"
			state.ContentState.Pane = ""
		}
	}
	state.PaneActions = views.PaneActions(layoutpkg.ViewTheme(), views.ActionsState{
		View:                   string(m.view),
		Pane:                   string(m.pane),
		ScratchpadOpen:         m.scratchpadOpen,
		TimerState:             timerState,
		RestModeActive:         state.ContentState.RestModeActive,
		AwayModeActive:         state.ContentState.AwayModeActive,
		IsDevMode:              m.isDevMode(),
		UpdateVisible:          viewsShouldShowUpdate(m.updateStatus),
		UpdateInstallAvailable: m.selfUpdateInstallAvailable(),
	})
	if m.dialog != "" {
		state.DialogState = m.dialogRenderState()
	}
	return state
}

func globalActions(m Model, timerState string) []string {
	return views.GlobalActions(layoutpkg.ViewTheme(), views.ActionsState{
		View:                   string(m.view),
		Pane:                   string(m.pane),
		ScratchpadOpen:         m.scratchpadOpen,
		TimerState:             timerState,
		IsDevMode:              m.isDevMode(),
		UpdateVisible:          viewsShouldShowUpdate(m.updateStatus),
		UpdateInstallAvailable: m.selfUpdateInstallAvailable(),
	})
}

func sidebarWidth(width int) int {
	if width < 64 {
		return max(14, width/4)
	}
	if width < 90 {
		return 18
	}
	return 22
}

func (m Model) mainContentWidth() int {
	return max(0, m.width-sidebarWidth(m.width))
}

func (m Model) isDevMode() bool {
	return m.kernelInfo != nil && m.kernelInfo.Env == "Dev"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (m Model) contentHeight() int {
	return layoutpkg.ContentHeight(m.layoutState())
}
