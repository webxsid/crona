package app

import (
	"strings"
	"time"

	versionpkg "crona/shared/version"
	helperpkg "crona/tui/internal/tui/helpers"
	layoutpkg "crona/tui/internal/tui/layout"
	selectionpkg "crona/tui/internal/tui/selection"
	viewchrome "crona/tui/internal/tui/views/chrome"
	viewruntime "crona/tui/internal/tui/views/runtime"
)

func (m Model) View() string {
	return layoutpkg.Render(m.layoutState())
}

func (m Model) layoutState() layoutpkg.State {
	snapshot := m.selectionSnapshot()
	activeIssue := selectionpkg.ActiveIssue(snapshot)
	chromeState := m.layoutChromeState()
	state := layoutpkg.State{
		Width:               m.width,
		Height:              m.height,
		View:                m.view,
		Pane:                m.pane,
		ProtectedMode:       chromeState.ProtectedMode,
		IsDevMode:           m.isDevMode(),
		RepoName:            chromeState.RepoName,
		StreamName:          chromeState.StreamName,
		TimerActive:         m.timer != nil && m.timer.State != "idle",
		HeaderState:         chromeState.HeaderState,
		DialogOpen:          m.dialog != "",
		HelpOpen:            m.helpOpen,
		SessionDetailOpen:   m.sessionDetailOpen,
		SessionDetailY:      m.sessionDetailY,
		SessionDetailLines:  helperpkg.SessionDetailContentLines(m.sessionDetail),
		SessionContextOpen:  m.sessionContextOpen,
		SessionContextY:     m.sessionContextY,
		SessionContextLines: helperpkg.SessionContextContentLines(activeIssue),
		StatusMsg:           m.statusMsg,
		StatusErr:           m.statusErr,
		GlobalActions:       chromeState.GlobalActions,
	}
	contentWidth := max(0, m.width-sidebarWidth(m.width))
	state.ContentState = m.viewContentState(contentWidth, layoutpkg.ContentHeight(state), snapshot, activeIssue)
	if state.ContentState.RestModeActive {
		if m.view != ViewReports && m.view != ViewSessionHistory {
			state.View = ViewAway
			state.ContentState.View = "away"
			state.ContentState.Pane = ""
		}
	}
	state.PaneActions = viewchrome.PaneActions(layoutpkg.ViewTheme(), viewchrome.ActionsState{
		View:                   string(m.view),
		Pane:                   string(m.pane),
		ScratchpadOpen:         m.scratchpadOpen,
		TimerState:             chromeState.TimerState,
		RestModeActive:         state.ContentState.RestModeActive,
		AwayModeActive:         state.ContentState.AwayModeActive,
		IsDevMode:              m.isDevMode(),
		IsBetaBuild:            m.isBetaBuild(),
		UpdateVisible:          viewsShouldShowUpdate(m.updateStatus),
		UpdateInstallAvailable: m.selfUpdateInstallAvailable(),
	})
	if m.dialog != "" {
		state.DialogState = m.dialogRenderState()
	}
	return state
}

type layoutChromeState struct {
	ProtectedMode bool
	RepoName      string
	StreamName    string
	TimerState    string
	HeaderState   viewchrome.HeaderState
	GlobalActions []string
}

func (m Model) layoutChromeState() layoutChromeState {
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
	protectedMode, _, _ := viewruntime.ProtectedRestMode(m.settings, time.Now().Format("2006-01-02"))
	headerState := viewchrome.HeaderState{
		Width:         m.width,
		View:          string(m.view),
		Elapsed:       m.elapsed,
		Timer:         m.timer,
		IssueSessions: m.issueSessions,
		AllIssues:     m.allIssues,
		Health:        m.health,
		UpdateStatus:  m.updateStatus,
	}
	return layoutChromeState{
		ProtectedMode: protectedMode,
		RepoName:      repo,
		StreamName:    stream,
		TimerState:    timerState,
		HeaderState:   headerState,
		GlobalActions: viewchrome.GlobalActions(layoutpkg.ViewTheme(), viewchrome.ActionsState{
			View:                   string(m.view),
			Pane:                   string(m.pane),
			ScratchpadOpen:         m.scratchpadOpen,
			TimerState:             timerState,
			IsDevMode:              m.isDevMode(),
			IsBetaBuild:            m.isBetaBuild(),
			UpdateVisible:          viewsShouldShowUpdate(m.updateStatus),
			UpdateInstallAvailable: m.selfUpdateInstallAvailable(),
		}),
	}
}

func (m Model) isBetaBuild() bool {
	if m.kernelInfo != nil && m.kernelInfo.RunningIsBeta {
		return true
	}
	if m.updateStatus != nil && m.updateStatus.RunningIsBeta {
		return true
	}
	return versionpkg.IsBetaRelease()
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

func (m Model) contentHeight() int {
	chromeState := m.layoutChromeState()
	return layoutpkg.ContentHeight(layoutpkg.State{
		Width:         m.width,
		Height:        m.height,
		ProtectedMode: chromeState.ProtectedMode,
		IsDevMode:     m.isDevMode(),
		RepoName:      chromeState.RepoName,
		StreamName:    chromeState.StreamName,
		GlobalActions: chromeState.GlobalActions,
		HeaderState:   chromeState.HeaderState,
	})
}
