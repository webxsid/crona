package model

import (
	"crona/tui/internal/tui/commands"
	helperpkg "crona/tui/internal/tui/helpers"
	inputpkg "crona/tui/internal/tui/input"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) supportDiagnosticsInput() helperpkg.SupportDiagnosticsInput {
	return helperpkg.SupportDiagnosticsInput{
		View:                string(m.view),
		Pane:                string(m.pane),
		Width:               m.width,
		Height:              m.height,
		DashboardDate:       m.currentDashboardDate(),
		RollupStartDate:     m.currentRollupStartDate(),
		RollupEndDate:       m.currentRollupEndDate(),
		WellbeingDate:       m.currentWellbeingDate(),
		ReposCount:          len(m.repos),
		StreamsCount:        len(m.streams),
		IssuesCount:         len(m.issues),
		AllIssuesCount:      len(m.allIssues),
		HabitsCount:         len(m.habits),
		DueHabitsCount:      len(m.dueHabits),
		ReportsCount:        len(m.exportReports),
		SessionHistoryCount: len(m.sessionHistory),
		ScratchpadsCount:    len(m.scratchpads),
		OpsCount:            len(m.ops),
		Context:             m.context,
		Timer:               m.timer,
		Settings:            m.settings,
		KernelInfo:          m.kernelInfo,
		ExportAssets:        m.exportAssets,
		UpdateStatus:        m.updateStatus,
		Health:              m.health,
		TUIPath:             m.currentExecutablePath,
		KernelPath:          kernelExecutablePath(m.kernelInfo),
	}
}

func (m Model) openSupportIssueURL() tea.Cmd {
	return commands.OpenExternalURL(helperpkg.SupportBugReportURL(m.supportDiagnosticsInput(), m.dialogSupportBundlePath))
}

func (m Model) openSupportDiscussionsURL() tea.Cmd {
	return commands.OpenExternalURL(helperpkg.SupportDiscussionsURL())
}

func (m Model) openSupportReleasesURL() tea.Cmd {
	return commands.OpenExternalURL(helperpkg.SupportReleasesURL())
}

func (m Model) openSupportRoadmapURL() tea.Cmd {
	return commands.OpenExternalURL(helperpkg.SupportRoadmapURL())
}

func (m Model) copySupportDiagnosticsCmd(state inputpkg.State) tea.Cmd {
	next := m.applyInputState(state)
	report := helperpkg.SupportDiagnosticsReport(next.supportDiagnosticsInput())
	return commands.CopyTextToClipboard(report, "Detailed diagnostics copied")
}

func (m Model) generateSupportBundleCmd(state inputpkg.State) tea.Cmd {
	next := m.applyInputState(state)
	return commands.GenerateSupportBundle(m.client, next.supportDiagnosticsInput(), helperpkg.SupportRecentDiagnosticsWindow)
}
