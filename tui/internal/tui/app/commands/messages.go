package commands

import "crona/tui/internal/api"

type ReposLoadedMsg struct{ Repos []api.Repo }
type StreamsLoadedMsg struct{ Streams []api.Stream }

type IssuesLoadedMsg struct {
	StreamID int64
	Issues   []api.Issue
}

type HabitsLoadedMsg struct {
	StreamID int64
	Habits   []api.Habit
}

type AllIssuesLoadedMsg struct{ Issues []api.IssueWithMeta }
type DueHabitsLoadedMsg struct{ Habits []api.HabitDailyItem }
type DailySummaryLoadedMsg struct{ Summary *api.DailyIssueSummary }
type DailyCheckInLoadedMsg struct{ CheckIn *api.DailyCheckIn }
type MetricsRangeLoadedMsg struct{ Days []api.DailyMetricsDay }
type MetricsRollupLoadedMsg struct{ Rollup *api.MetricsRollup }
type StreaksLoadedMsg struct{ Streaks *api.StreakSummary }
type ExportAssetsLoadedMsg struct{ Assets *api.ExportAssetStatus }
type ExportReportsLoadedMsg struct{ Reports []api.ExportReportFile }
type ExportReportDeletedMsg struct{ Name string }
type DailyReportGeneratedMsg struct{ Result *api.DailyReportResult }
type CalendarExportGeneratedMsg struct{ Result *api.CalendarExportResult }
type ClipboardCopiedMsg struct{ Message string }

type IssueSessionsLoadedMsg struct {
	IssueID  int64
	Sessions []api.Session
}

type SessionHistoryLoadedMsg struct{ Sessions []api.SessionHistoryEntry }
type SessionDetailLoadedMsg struct{ Detail *api.SessionDetail }
type SessionDetailFailedMsg struct{ Err error }
type SessionAmendedMsg struct{ ID string }
type ScratchpadsLoadedMsg struct{ Pads []api.ScratchPad }
type StashesLoadedMsg struct{ Stashes []api.Stash }
type OpsLoadedMsg struct{ Ops []api.Op }
type ContextLoadedMsg struct{ Ctx *api.ActiveContext }
type TimerLoadedMsg struct{ Timer *api.TimerState }
type HealthLoadedMsg struct{ Health *api.Health }
type UpdateStatusLoadedMsg struct{ Status *api.UpdateStatus }
type UpdateDismissedMsg struct{ Status *api.UpdateStatus }
type UpdateInstallFinishedMsg struct {
	Output string
	Err    error
}
type SettingsLoadedMsg struct{ Settings *api.CoreSettings }
type KernelInfoLoadedMsg struct{ Info *api.KernelInfo }
type KernelEventMsg struct{ Event api.KernelEvent }
type KernelShutdownMsg struct{}
type DevSeededMsg struct{}
type DevClearedMsg struct{}
type TimerTickMsg struct{ Seq int }
type HealthTickMsg struct{}
type ErrMsg struct{ Err error }
type ClearStatusMsg struct{ Seq int }

type OpenScratchpadMsg struct {
	Meta     api.ScratchPad
	FilePath string
	Content  string
}

type FocusSessionChangedMsg struct {
	ReloadContext bool
	ReloadTimer   bool
}

type ScratchpadReloadedMsg struct {
	Rendered string
	FilePath string
}
