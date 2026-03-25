package app

import (
	appcmd "crona/tui/internal/tui/app/commands"

	tea "github.com/charmbracelet/bubbletea"
)

type reposLoadedMsg = appcmd.ReposLoadedMsg
type streamsLoadedMsg = appcmd.StreamsLoadedMsg
type issuesLoadedMsg = appcmd.IssuesLoadedMsg
type habitsLoadedMsg = appcmd.HabitsLoadedMsg
type allIssuesLoadedMsg = appcmd.AllIssuesLoadedMsg
type dueHabitsLoadedMsg = appcmd.DueHabitsLoadedMsg
type dailySummaryLoadedMsg = appcmd.DailySummaryLoadedMsg
type dailyCheckInLoadedMsg = appcmd.DailyCheckInLoadedMsg
type metricsRangeLoadedMsg = appcmd.MetricsRangeLoadedMsg
type metricsRollupLoadedMsg = appcmd.MetricsRollupLoadedMsg
type streaksLoadedMsg = appcmd.StreaksLoadedMsg
type exportAssetsLoadedMsg = appcmd.ExportAssetsLoadedMsg
type exportReportsLoadedMsg = appcmd.ExportReportsLoadedMsg
type exportReportDeletedMsg = appcmd.ExportReportDeletedMsg
type dailyReportGeneratedMsg = appcmd.DailyReportGeneratedMsg
type calendarExportGeneratedMsg = appcmd.CalendarExportGeneratedMsg
type clipboardCopiedMsg = appcmd.ClipboardCopiedMsg
type issueSessionsLoadedMsg = appcmd.IssueSessionsLoadedMsg
type sessionHistoryLoadedMsg = appcmd.SessionHistoryLoadedMsg
type sessionDetailLoadedMsg = appcmd.SessionDetailLoadedMsg
type sessionDetailFailedMsg = appcmd.SessionDetailFailedMsg
type sessionAmendedMsg = appcmd.SessionAmendedMsg
type scratchpadsLoadedMsg = appcmd.ScratchpadsLoadedMsg
type stashesLoadedMsg = appcmd.StashesLoadedMsg
type opsLoadedMsg = appcmd.OpsLoadedMsg
type contextLoadedMsg = appcmd.ContextLoadedMsg
type timerLoadedMsg = appcmd.TimerLoadedMsg
type healthLoadedMsg = appcmd.HealthLoadedMsg
type updateStatusLoadedMsg = appcmd.UpdateStatusLoadedMsg
type updateDismissedMsg = appcmd.UpdateDismissedMsg
type updateInstallFinishedMsg = appcmd.UpdateInstallFinishedMsg
type settingsLoadedMsg = appcmd.SettingsLoadedMsg
type kernelInfoLoadedMsg = appcmd.KernelInfoLoadedMsg
type kernelEventMsg = appcmd.KernelEventMsg
type kernelShutdownMsg = appcmd.KernelShutdownMsg
type devSeededMsg = appcmd.DevSeededMsg
type devClearedMsg = appcmd.DevClearedMsg
type timerTickMsg = appcmd.TimerTickMsg
type healthTickMsg = appcmd.HealthTickMsg
type errMsg = appcmd.ErrMsg
type clearStatusMsg = appcmd.ClearStatusMsg
type openScratchpadMsg = appcmd.OpenScratchpadMsg
type focusSessionChangedMsg = appcmd.FocusSessionChangedMsg
type scratchpadReloadedMsg = appcmd.ScratchpadReloadedMsg

var (
	loadRepos                         = appcmd.LoadRepos
	loadStreams                       = appcmd.LoadStreams
	loadIssues                        = appcmd.LoadIssues
	loadHabits                        = appcmd.LoadHabits
	loadAllIssues                     = appcmd.LoadAllIssues
	loadDueHabits                     = appcmd.LoadDueHabits
	loadDailySummary                  = appcmd.LoadDailySummary
	loadDailyCheckIn                  = appcmd.LoadDailyCheckIn
	loadMetricsRange                  = appcmd.LoadMetricsRange
	loadMetricsRollup                 = appcmd.LoadMetricsRollup
	loadMetricsStreaks                = appcmd.LoadMetricsStreaks
	loadIssueSessions                 = appcmd.LoadIssueSessions
	loadSessionHistory                = appcmd.LoadSessionHistory
	loadSessionDetail                 = appcmd.LoadSessionDetail
	loadScratchpads                   = appcmd.LoadScratchpads
	loadStashes                       = appcmd.LoadStashes
	loadOps                           = appcmd.LoadOps
	loadContext                       = appcmd.LoadContext
	loadTimer                         = appcmd.LoadTimer
	loadHealth                        = appcmd.LoadHealth
	loadUpdateStatus                  = appcmd.LoadUpdateStatus
	loadSettings                      = appcmd.LoadSettings
	loadKernelInfo                    = appcmd.LoadKernelInfo
	loadExportAssets                  = appcmd.LoadExportAssets
	loadExportReports                 = appcmd.LoadExportReports
	tickAfter                         = appcmd.TickAfter
	healthTickAfter                   = appcmd.HealthTickAfter
	clearStatusAfter                  = appcmd.ClearStatusAfter
	waitForEvent                      = appcmd.WaitForEvent
	loadWellbeing                     = appcmd.LoadWellbeing
	cmdPatchSetting                   = appcmd.PatchSetting
	cmdUpsertDailyCheckIn             = appcmd.UpsertDailyCheckIn
	cmdDeleteDailyCheckIn             = appcmd.DeleteDailyCheckIn
	cmdShutdownKernel                 = appcmd.ShutdownKernel
	cmdDismissUpdate                  = appcmd.DismissUpdate
	cmdCheckUpdateNow                 = appcmd.CheckUpdateNow
	cmdOpenExternalURL                = appcmd.OpenExternalURL
	cmdInstallUpdate                  = appcmd.InstallUpdate
	cmdSeedDevData                    = appcmd.SeedDevData
	cmdClearDevData                   = appcmd.ClearDevData
	cmdCreateScratchpad               = appcmd.CreateScratchpad
	cmdCreateRepoOnly                 = appcmd.CreateRepoOnly
	cmdUpdateRepo                     = appcmd.UpdateRepo
	cmdDeleteRepo                     = appcmd.DeleteRepo
	cmdCreateStreamOnly               = appcmd.CreateStreamOnly
	cmdUpdateStream                   = appcmd.UpdateStream
	cmdDeleteStream                   = appcmd.DeleteStream
	cmdCreateIssueOnly                = appcmd.CreateIssueOnly
	cmdCreateHabitOnly                = appcmd.CreateHabitOnly
	cmdUpdateHabit                    = appcmd.UpdateHabit
	cmdDeleteHabit                    = appcmd.DeleteHabit
	cmdSetHabitStatus                 = appcmd.SetHabitStatus
	cmdUncompleteHabit                = appcmd.UncompleteHabit
	cmdUpdateIssue                    = appcmd.UpdateIssue
	cmdDeleteIssue                    = appcmd.DeleteIssue
	cmdCreateIssueWithPath            = appcmd.CreateIssueWithPath
	cmdCreateHabitWithPath            = appcmd.CreateHabitWithPath
	cmdDeleteScratchpad               = appcmd.DeleteScratchpad
	cmdOpenScratchpad                 = appcmd.OpenScratchpad
	cmdCheckoutRepo                   = appcmd.CheckoutRepo
	cmdCheckoutStream                 = appcmd.CheckoutStream
	cmdCheckoutContext                = appcmd.CheckoutContext
	cmdChangeIssueStatus              = appcmd.ChangeIssueStatus
	cmdToggleIssueToday               = appcmd.ToggleIssueToday
	cmdSetIssueTodoDate               = appcmd.SetIssueTodoDate
	cmdChangeIssueStatusAndEndSession = appcmd.ChangeIssueStatusAndEndSession
	cmdAmendSessionNote               = appcmd.AmendSessionNote
	cmdStartFocusSession              = appcmd.StartFocusSession
	cmdPauseFocusSession              = appcmd.PauseFocusSession
	cmdResumeFocusSession             = appcmd.ResumeFocusSession
	cmdEndFocusSession                = appcmd.EndFocusSession
	cmdStashFocusSession              = appcmd.StashFocusSession
	cmdApplyStash                     = appcmd.ApplyStash
	cmdDropStash                      = appcmd.DropStash
	cmdGenerateReport                 = appcmd.GenerateReport
	cmdGenerateCalendarExport         = appcmd.GenerateCalendarExport
	cmdGenerateDailyReport            = appcmd.GenerateDailyReport
	cmdCopyDailyReport                = appcmd.CopyDailyReport
	cmdResetExportTemplate            = appcmd.ResetExportTemplate
	cmdSetExportReportsDir            = appcmd.SetExportReportsDir
	cmdSetExportICSDir                = appcmd.SetExportICSDir
	cmdDeleteExportReport             = appcmd.DeleteExportReport
)

func loadSessionHistoryForModel(m Model, limit int) tea.Cmd {
	return appcmd.LoadSessionHistory(m.client, m.sessionHistoryScopeIssueID(), limit)
}
