package model

import (
	"strings"
	"time"

	sharedtypes "crona/shared/types"
	versionpkg "crona/shared/version"
	"crona/tui/internal/api"
	"crona/tui/internal/tui/dialogs"
	dialogstate "crona/tui/internal/tui/dialogs/controller"
	helperpkg "crona/tui/internal/tui/helpers"
	layoutpkg "crona/tui/internal/tui/layout"
	selectionpkg "crona/tui/internal/tui/selection"
	viewruntime "crona/tui/internal/tui/views/runtime"
	viewtypes "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/bubbles/textinput"
)

func (m Model) viewContentState(
	width, height int,
	snapshot selectionpkg.Snapshot,
	activeIssue *api.IssueWithMeta,
) viewtypes.ContentState {
	sessionIssueID := helperpkg.SessionHistoryScopeIssueID(m.timer)
	sessionIssue := activeIssue
	dailyIssues := selectionpkg.DailyScopedIssues(snapshot)
	defaultIssues := selectionpkg.DefaultScopedIssues(snapshot)
	dueHabits := selectionpkg.FilteredDueHabits(snapshot)
	state := viewtypes.ContentState{
		View:                string(m.view),
		Pane:                string(m.pane),
		Width:               width,
		Height:              height,
		Elapsed:             m.elapsed,
		DashboardDate:       m.dashboardDate,
		RollupStartDate:     m.currentRollupStartDate(),
		RollupEndDate:       m.currentRollupEndDate(),
		MomentumDate:        m.currentMomentumDate(),
		MomentumWindowDays:  m.currentMomentumWindowDays(),
		MomentumTab:         string(m.currentMomentumTab()),
		MomentumHistoryY:    m.momentumHistoryCursor,
		WellbeingDate:       m.currentWellbeingDate(),
		WellbeingWindowDays: m.currentWellbeingWindowDays(),
		WeekStart:           m.currentWeekStart(),
		DefaultIssueSection: string(m.defaultIssueSection),
		DailyTaskSection:    string(m.dailyTaskSection),
		SessionHistoryTitle: helperpkg.SessionHistoryTitle(sessionIssueID, sessionIssue),
		SessionHistoryMeta:  helperpkg.SessionHistorySubtitle(sessionIssueID, sessionIssue),
		HabitHistoryTitle:   m.habitHistoryTitle,
		HabitHistoryMeta:    m.habitHistoryMeta,
		Cursors: map[string]int{
			"repos":             m.cursor[PaneRepos],
			"streams":           m.cursor[PaneStreams],
			"issues":            m.cursor[PaneIssues],
			"habits":            m.cursor[PaneHabits],
			"habit_history":     m.cursor[PaneHabitHistory],
			"momentum_cards":    m.momentumCursorForView(),
			"rollup_days":       m.cursor[PaneRollupDays],
			"rollup_breakdown":  m.cursor[PaneRollupBreakdown],
			"sessions":          m.cursor[PaneSessions],
			"ops":               m.cursor[PaneOps],
			"export_reports":    m.cursor[PaneExportReports],
			"config":            m.cursor[PaneConfig],
			"settings":          m.cursor[PaneSettings],
			"alerts":            m.cursor[PaneAlerts],
			"wellbeing_summary": m.cursor[PaneWellbeingSummary],
			"wellbeing_trends":  m.cursor[PaneWellbeingTrends],
			"wellbeing_details": m.cursor[PaneWellbeingDetails],
		},
		Filters: map[string]string{
			"repos":             m.filters[PaneRepos],
			"streams":           m.filters[PaneStreams],
			"issues":            m.filters[PaneIssues],
			"habits":            m.filters[PaneHabits],
			"habit_history":     m.filters[PaneHabitHistory],
			"momentum_cards":    m.filters[PaneMomentumCards],
			"rollup_days":       m.filters[PaneRollupDays],
			"rollup_breakdown":  m.filters[PaneRollupBreakdown],
			"sessions":          m.filters[PaneSessions],
			"ops":               m.filters[PaneOps],
			"export_reports":    m.filters[PaneExportReports],
			"config":            m.filters[PaneConfig],
			"settings":          m.filters[PaneSettings],
			"alerts":            m.filters[PaneAlerts],
			"wellbeing_summary": m.filters[PaneWellbeingSummary],
			"wellbeing_trends":  m.filters[PaneWellbeingTrends],
			"wellbeing_details": m.filters[PaneWellbeingDetails],
		},
		Repos:                     m.repos,
		Streams:                   m.streams,
		Issues:                    m.issues,
		DailyIssues:               dailyIssues,
		Habits:                    m.habits,
		AllIssues:                 m.allIssues,
		DefaultIssues:             defaultIssues,
		DueHabits:                 dueHabits,
		MomentumCards:             m.momentumCards,
		DailySummary:              m.dailySummary,
		DailyPlan:                 m.dailyPlan,
		DailyCheckIn:              m.dailyCheckIn,
		MetricsRange:              m.metricsRange,
		MetricsRollup:             m.metricsRollup,
		RollupMetricsRange:        m.rollupMetricsRange,
		RollupMetricsRollup:       m.rollupMetricsRollup,
		MomentumMetricsRange:      m.momentumMetricsRange,
		MomentumMetricsRollup:     m.momentumMetricsRollup,
		Streaks:                   m.streaks,
		DailyStreaks:              m.dailyStreaks,
		DashboardWindow:           m.dashboardWindow,
		DailyFocusScore:           m.dailyFocusScore,
		WeeklyFocusScore:          m.weeklyFocusScore,
		RepoDistribution:          m.repoDistribution,
		StreamDistribution:        m.streamDistribution,
		IssueDistribution:         m.issueDistribution,
		SegmentDistribution:       m.segmentDistribution,
		GoalProgress:              m.goalProgress,
		ExportAssets:              m.exportAssets,
		ExportReports:             m.exportReports,
		IssueSessions:             m.issueSessions,
		SessionHistory:            m.sessionHistory,
		HabitHistory:              m.habitHistory,
		Ops:                       m.ops,
		Context:                   m.context,
		Timer:                     m.timer,
		Health:                    m.health,
		AlertStatus:               m.alertStatus,
		AlertReminders:            m.alertReminders,
		UpdateStatus:              m.updateStatus,
		UpdateChecking:            m.updateChecking,
		UpdateInstallAvailable:    m.selfUpdateInstallAvailable(),
		UpdateDiagnosticsExpanded: m.updateDiagnosticsExpanded,
		UpdateCommand:             updateCommand(m.updateStatus, m.currentExecutablePath, kernelExecutablePath(m.kernelInfo)),
		UpdateManualReason:        m.selfUpdateUnsupportedReason(),
		TUIExecutablePath:         m.currentExecutablePath,
		KernelExecutablePath:      kernelExecutablePath(m.kernelInfo),
		KernelInfo:                m.kernelInfo,
		Settings:                  m.settings,
	}
	restDate := time.Now().Format("2006-01-02")
	if active, away, detail := viewruntime.ProtectedRestMode(state.Settings, restDate); active {
		state.RestModeActive = true
		state.AwayModeActive = away
		state.RestModeDetail = detail
		state.RestModeMessage = viewruntime.RestModeMessage(restDate)
	}
	return state
}

func (m Model) dialogRenderState() dialogstate.State {
	state := m.dialogState()
	state.Width = m.width
	if m.dialog == "create_issue_default" || m.dialog == "create_habit" {
		state.RepoSelectorLabel, state.StreamSelectorLabel = dialogstate.DefaultIssueDialogLabels(
			m.dialogInputs,
			m.dialogRepoIndex,
			m.dialogStreamIndex,
			m.repos,
			m.allIssues,
			m.streams,
			m.context,
		)
	}
	if m.dialog == "create_momentum" || m.dialog == "edit_momentum" {
		state.RepoSelectorLabel, state.StreamSelectorLabel = dialogstate.MomentumDialogLabels(
			[]textinput.Model{m.dialogMomentumRepoInput, m.dialogMomentumStreamInput},
			m.dialogRepoIndex,
			m.dialogStreamIndex,
			m.repos,
			m.allIssues,
			m.streams,
			m.context,
		)
	}
	if m.dialog == "checkout_context" {
		state.RepoSelectorLabel, state.StreamSelectorLabel = dialogstate.CheckoutDialogLabels(
			m.dialogInputs,
			m.dialogRepoIndex,
			m.dialogStreamIndex,
			m.repos,
			m.allIssues,
			m.streams,
			m.context,
		)
	}
	if m.dialog == "pick_date" {
		state = dialogstate.PopulateDatePresentation(
			dialogControllerTheme(layoutpkg.DialogTheme()),
			state,
			m.currentDashboardDate(),
		)
	}
	return state
}

func dialogControllerTheme(theme dialogs.Theme) dialogstate.Theme {
	return dialogstate.Theme{
		ColorCyan:      theme.ColorCyan,
		ColorYellow:    theme.ColorYellow,
		ColorRed:       theme.ColorRed,
		ColorGreen:     theme.ColorGreen,
		StylePaneTitle: theme.StylePaneTitle,
		StyleDim:       theme.StyleDim,
		StyleCursor:    theme.StyleCursor,
		StyleHeader:    theme.StyleHeader,
		StyleError:     theme.StyleError,
		StyleSelected:  theme.StyleSelected,
		StyleNormal:    theme.StyleNormal,
	}
}

func updateCommand(status *api.UpdateStatus, currentExecutablePath, kernelExecutablePath string) string {
	if status == nil {
		return ""
	}
	if command := strings.TrimSpace(status.UpdateCommand); command != "" {
		return command
	}
	switch installSourceFromPath(currentExecutablePath) {
	case sharedtypes.InstallSourceBrew:
		return "brew upgrade " + currentBrewFormula()
	case sharedtypes.InstallSourceWinget:
		return "winget upgrade --id Webxsid.Crona -e"
	case sharedtypes.InstallSourceGo:
		return "go install github.com/webxsid/crona/...@latest"
	}
	switch installSourceFromPath(kernelExecutablePath) {
	case sharedtypes.InstallSourceBrew:
		return "brew upgrade " + currentBrewFormula()
	case sharedtypes.InstallSourceWinget:
		return "winget upgrade --id Webxsid.Crona -e"
	case sharedtypes.InstallSourceGo:
		return "go install github.com/webxsid/crona/...@latest"
	}
	return ""
}

func currentBrewFormula() string {
	if versionpkg.IsBetaRelease() {
		return "crona-beta"
	}
	return "crona"
}
