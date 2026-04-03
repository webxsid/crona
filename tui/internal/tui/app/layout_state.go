package app

import (
	"fmt"
	"strings"
	"time"

	"crona/tui/internal/api"
	"crona/tui/internal/tui/dialogs"
	helperpkg "crona/tui/internal/tui/helpers"
	layoutpkg "crona/tui/internal/tui/layout"
	selectionpkg "crona/tui/internal/tui/selection"
	"crona/tui/internal/tui/views"
)

func (m Model) viewContentState(width, height int, snapshot selectionpkg.Snapshot, activeIssue *api.IssueWithMeta) views.ContentState {
	sessionIssueID := helperpkg.SessionHistoryScopeIssueID(m.timer)
	sessionIssue := activeIssue
	dailyIssues := selectionpkg.DailyScopedIssues(snapshot)
	defaultIssues := selectionpkg.DefaultScopedIssues(snapshot)
	dueHabits := selectionpkg.FilteredDueHabits(snapshot)
	state := views.ContentState{
		View: string(m.view), Pane: string(m.pane), Width: width, Height: height, Elapsed: m.elapsed, DashboardDate: m.dashboardDate, RollupStartDate: m.currentRollupStartDate(), RollupEndDate: m.currentRollupEndDate(), WellbeingDate: m.currentWellbeingDate(), DefaultIssueSection: string(m.defaultIssueSection), SessionHistoryTitle: helperpkg.SessionHistoryTitle(sessionIssueID, sessionIssue), SessionHistoryMeta: helperpkg.SessionHistorySubtitle(sessionIssueID, sessionIssue),
		Cursors:            map[string]int{"repos": m.cursor[PaneRepos], "streams": m.cursor[PaneStreams], "issues": m.cursor[PaneIssues], "habits": m.cursor[PaneHabits], "rollup_days": m.cursor[PaneRollupDays], "sessions": m.cursor[PaneSessions], "scratchpads": m.cursor[PaneScratchpads], "ops": m.cursor[PaneOps], "export_reports": m.cursor[PaneExportReports], "config": m.cursor[PaneConfig], "settings": m.cursor[PaneSettings]},
		Filters:            map[string]string{"repos": m.filters[PaneRepos], "streams": m.filters[PaneStreams], "issues": m.filters[PaneIssues], "habits": m.filters[PaneHabits], "rollup_days": m.filters[PaneRollupDays], "sessions": m.filters[PaneSessions], "scratchpads": m.filters[PaneScratchpads], "ops": m.filters[PaneOps], "export_reports": m.filters[PaneExportReports], "config": m.filters[PaneConfig], "settings": m.filters[PaneSettings]},
		ScratchpadOpen:     m.scratchpadOpen,
		ScratchpadRendered: m.scratchpadViewport.View(),
		Repos:              m.repos, Streams: m.streams, Issues: m.issues, DailyIssues: dailyIssues, Habits: m.habits, AllIssues: m.allIssues, DefaultIssues: defaultIssues, DueHabits: dueHabits, DailySummary: m.dailySummary, DailyPlan: m.dailyPlan, DailyCheckIn: m.dailyCheckIn, MetricsRange: m.metricsRange, MetricsRollup: m.metricsRollup, Streaks: m.streaks, DashboardWindow: m.dashboardWindow, DailyFocusScore: m.dailyFocusScore, WeeklyFocusScore: m.weeklyFocusScore, RepoDistribution: m.repoDistribution, StreamDistribution: m.streamDistribution, IssueDistribution: m.issueDistribution, SegmentDistribution: m.segmentDistribution, GoalProgress: m.goalProgress, ExportAssets: m.exportAssets, ExportReports: m.exportReports, IssueSessions: m.issueSessions, SessionHistory: m.sessionHistory, Scratchpads: m.scratchpads, Ops: m.ops, Context: m.context, Timer: m.timer, Health: m.health, UpdateStatus: m.updateStatus, UpdateChecking: m.updateChecking, UpdateInstalling: m.updateInstalling, UpdateInstallError: m.updateInstallError, UpdateInstallAvailable: m.selfUpdateInstallAvailable(), UpdateManualReason: m.selfUpdateUnsupportedReason(), TUIExecutablePath: m.currentExecutablePath, KernelExecutablePath: kernelExecutablePath(m.kernelInfo), KernelInfo: m.kernelInfo, Settings: m.settings,
	}
	restDate := time.Now().Format("2006-01-02")
	if active, away, detail := views.ProtectedRestMode(state.Settings, restDate); active {
		state.RestModeActive = true
		state.AwayModeActive = away
		state.RestModeDetail = detail
		state.RestModeMessage = views.RestModeMessage(restDate)
	}
	if m.scratchpadMeta != nil {
		state.ScratchpadName = m.scratchpadMeta.Name
		state.ScratchpadPath = m.scratchpadMeta.Path
	}
	return state
}

func (m Model) dialogRenderState() dialogs.State {
	state := m.dialogState()
	state.Width = m.width
	if m.dialog == "create_issue_default" || m.dialog == "create_habit" {
		state.RepoSelectorLabel, state.StreamSelectorLabel = dialogs.DefaultIssueDialogLabels(m.dialogInputs, m.dialogRepoIndex, m.dialogStreamIndex, m.repos, m.allIssues, m.streams, m.context)
	}
	if m.dialog == "checkout_context" {
		state.RepoSelectorLabel, state.StreamSelectorLabel = dialogs.CheckoutDialogLabels(m.dialogInputs, m.dialogRepoIndex, m.dialogStreamIndex, m.repos, m.allIssues, m.streams, m.context)
	}
	if m.dialog == "pick_date" {
		state = dialogs.PopulateDatePresentation(layoutpkg.DialogTheme(), state, m.currentDashboardDate())
	}
	for _, stash := range m.stashes {
		label := stash.CreatedAt
		if stash.Note != nil && strings.TrimSpace(*stash.Note) != "" {
			label = *stash.Note
		}
		contextBits := []string{}
		if stash.RepoID != nil {
			contextBits = append(contextBits, fmt.Sprintf("repo:%d", *stash.RepoID))
		}
		if stash.StreamID != nil {
			contextBits = append(contextBits, fmt.Sprintf("stream:%d", *stash.StreamID))
		}
		if stash.IssueID != nil {
			contextBits = append(contextBits, fmt.Sprintf("issue:%d", *stash.IssueID))
		}
		meta := stash.CreatedAt
		if len(contextBits) > 0 {
			meta += "  " + strings.Join(contextBits, "  ")
		}
		state.Stashes = append(state.Stashes, dialogs.StashItem{Label: helperpkg.Truncate(label, 42), Meta: helperpkg.Truncate(meta, 48)})
	}
	return state
}
