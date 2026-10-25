package model

import (
	"strings"
	"time"

	versionpkg "crona/shared/version"
	helperpkg "crona/tui/internal/tui/helpers"
	layoutpkg "crona/tui/internal/tui/layout"
	selectionpkg "crona/tui/internal/tui/selection"
	uistate "crona/tui/internal/tui/state"
	viewchrome "crona/tui/internal/tui/views/chrome"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	viewruntime "crona/tui/internal/tui/views/runtime"
)

func (m Model) View() string {
	return layoutpkg.Render(m.layoutState())
}

// effectiveView is the view the shell shows: Away replaces views that protected rest hides.
func (m Model) effectiveView() View {
	if m.view == ViewReports || m.view == ViewSessionHistory ||
		m.view == ViewHabitHistory || m.view == ViewSettings {
		return m.view
	}
	if protected, _, _ := viewruntime.ProtectedRestMode(m.settings, m.currentLogicalDate()); protected {
		return ViewAway
	}
	return m.view
}

func (m Model) layoutState() layoutpkg.State {
	view := m.effectiveView()
	snapshot := m.selectionSnapshot()
	activeIssue := selectionpkg.ActiveIssue(snapshot)
	selectedIssue, hasSelectedIssue := m.selectedIssueRecord()
	selectedIssueAdhoc := hasSelectedIssue && selectedIssue.Adhoc
	chromeState := m.layoutChromeState()
	paneActions := viewchrome.PaneActions(layoutpkg.ViewTheme(), viewchrome.ActionsState{
		View:                   string(view),
		Pane:                   string(m.pane),
		SelectedIssueAdhoc:     selectedIssueAdhoc,
		TimerState:             chromeState.TimerState,
		TimerSegment:           chromeState.TimerSegment,
		TimerNextSegment:       chromeState.TimerNextSegment,
		RestModeActive:         stateRestModeFromChrome(chromeState),
		AwayModeActive:         stateAwayModeFromContent(m),
		IsDevMode:              m.isDevMode(),
		IsBetaBuild:            m.isBetaBuild(),
		UpdateVisible:          viewruntime.ShouldShowUpdatesView(m.updateStatus),
		UpdateInstallAvailable: m.selfUpdateInstallAvailable(),
		UpdateCommand: updateCommand(
			m.updateStatus,
			m.currentExecutablePath,
			kernelExecutablePath(m.kernelInfo),
		),
	})
	switch view {
	case ViewDaily,
		ViewDefault,
		ViewHabitList,
		ViewReview,
		ViewMomentum,
		ViewSessionHistory,
		ViewHabitHistory,
		ViewOps,
		ViewSettings,
		ViewTags,
		ViewAlerts,
		ViewReports,
		ViewConfig,
		ViewUpdates,
		ViewSupport:
		// These panes show their own keys.
		paneActions = nil
	}
	state := layoutpkg.State{
		Width:               m.width,
		Height:              m.height,
		View:                view,
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
		GlobalActions:       viewchrome.DedupeActionKeys(chromeState.GlobalActions, paneActions),
		PaneActions:         paneActions,
		Tabs:                m.visibleTabs(),
		SubTabs:             m.visibleTabViews(uistate.TabForView(view)),
	}
	state.ContentState = m.viewContentState(
		m.mainContentWidth(),
		layoutpkg.ContentHeight(state),
		snapshot,
		activeIssue,
	)
	// Without a detail pane, Daily's day summary key joins the bottom row.
	if view == ViewDaily && state.ContentState.HistoricalAwayDate == "" &&
		viewhelpers.ClassFor(state.ContentState.Width) != viewhelpers.WidthWide {
		state.GlobalActions = append(
			[]string{viewchrome.KeyHint(layoutpkg.ViewTheme(), "S", "day summary")},
			state.GlobalActions...)
	}
	if view == ViewAway && m.view != ViewAway {
		state.ContentState.View = "away"
		state.ContentState.Pane = ""
	}
	if m.dialog != "" {
		state.DialogState = m.dialogRenderState()
	}
	return state
}

func stateRestModeFromChrome(state layoutChromeState) bool {
	return state.ProtectedMode
}

func stateAwayModeFromContent(m Model) bool {
	protectedMode, awayMode, _ := viewruntime.ProtectedRestMode(
		m.settings,
		time.Now().Format("2006-01-02"),
	)
	if !protectedMode {
		return false
	}
	return awayMode
}

type layoutChromeState struct {
	ProtectedMode    bool
	RepoName         string
	StreamName       string
	TimerState       string
	TimerSegment     string
	TimerNextSegment string
	HeaderState      viewchrome.HeaderState
	GlobalActions    []string
}

func (m Model) layoutChromeState() layoutChromeState {
	view := m.effectiveView()
	repo := ""
	stream := ""
	selectedIssue, hasSelectedIssue := m.selectedIssueRecord()
	selectedIssueAdhoc := hasSelectedIssue && selectedIssue.Adhoc
	if m.context != nil {
		if m.context.RepoName != nil && strings.TrimSpace(*m.context.RepoName) != "" {
			repo = strings.TrimSpace(*m.context.RepoName)
		}
		if m.context.StreamName != nil && strings.TrimSpace(*m.context.StreamName) != "" {
			stream = strings.TrimSpace(*m.context.StreamName)
		}
	}
	timerState := ""
	timerSegment := ""
	timerNextSegment := ""
	hardLimitActive := false
	if m.timer != nil {
		timerState = m.timer.State
		hardLimitActive = m.timer.HardLimitActive
		if m.timer.SegmentType != nil {
			timerSegment = string(*m.timer.SegmentType)
		}
		if m.timer.ReadySegmentType != nil {
			timerNextSegment = string(*m.timer.ReadySegmentType)
		} else if m.timer.NextSegmentType != nil {
			timerNextSegment = string(*m.timer.NextSegmentType)
		}
	}
	protectedMode, _, _ := viewruntime.ProtectedRestMode(
		m.settings,
		time.Now().Format("2006-01-02"),
	)
	headerState := viewchrome.HeaderState{
		Width:         m.width,
		View:          string(view),
		Elapsed:       m.elapsed,
		Timer:         m.timer,
		IssueSessions: m.issueSessions,
		AllIssues:     m.allIssues,
		Health:        m.health,
		UpdateStatus:  m.updateStatus,
	}
	return layoutChromeState{
		ProtectedMode:    protectedMode,
		RepoName:         repo,
		StreamName:       stream,
		TimerState:       timerState,
		TimerSegment:     timerSegment,
		TimerNextSegment: timerNextSegment,
		HeaderState:      headerState,
		GlobalActions: viewchrome.GlobalActions(layoutpkg.ViewTheme(), viewchrome.ActionsState{
			View:                   string(view),
			Pane:                   string(m.pane),
			SelectedIssueAdhoc:     selectedIssueAdhoc,
			TimerState:             timerState,
			TimerSegment:           timerSegment,
			TimerNextSegment:       timerNextSegment,
			HardLimitActive:        hardLimitActive,
			IsDevMode:              m.isDevMode(),
			IsBetaBuild:            m.isBetaBuild(),
			UpdateVisible:          viewruntime.ShouldShowUpdatesView(m.updateStatus),
			UpdateInstallAvailable: m.selfUpdateInstallAvailable(),
			UpdateCommand: updateCommand(
				m.updateStatus,
				m.currentExecutablePath,
				kernelExecutablePath(m.kernelInfo),
			),
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

func (m Model) mainContentWidth() int {
	return max(0, m.width)
}

func (m Model) isDevMode() bool {
	return m.kernelInfo != nil && m.kernelInfo.Env == "Dev"
}

func (m Model) contentHeight() int {
	chromeState := m.layoutChromeState()
	return layoutpkg.ContentHeight(layoutpkg.State{
		View:          m.effectiveView(),
		Width:         m.width,
		Height:        m.height,
		ProtectedMode: chromeState.ProtectedMode,
		IsDevMode:     m.isDevMode(),
		RepoName:      chromeState.RepoName,
		StreamName:    chromeState.StreamName,
		GlobalActions: chromeState.GlobalActions,
		HeaderState:   chromeState.HeaderState,
		SubTabs:       m.visibleTabViews(uistate.TabForView(m.effectiveView())),
	})
}
