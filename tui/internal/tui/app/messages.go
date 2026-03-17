package app

import "crona/tui/internal/api"

type reposLoadedMsg struct{ repos []api.Repo }
type streamsLoadedMsg struct{ streams []api.Stream }

type issuesLoadedMsg struct {
	streamID int64
	issues   []api.Issue
}
type habitsLoadedMsg struct {
	streamID int64
	habits   []api.Habit
}

type allIssuesLoadedMsg struct{ issues []api.IssueWithMeta }
type dueHabitsLoadedMsg struct{ habits []api.HabitDailyItem }
type dailySummaryLoadedMsg struct{ summary *api.DailyIssueSummary }
type dailyCheckInLoadedMsg struct{ checkIn *api.DailyCheckIn }
type metricsRangeLoadedMsg struct{ days []api.DailyMetricsDay }
type metricsRollupLoadedMsg struct{ rollup *api.MetricsRollup }
type streaksLoadedMsg struct{ streaks *api.StreakSummary }

type issueSessionsLoadedMsg struct {
	issueID  int64
	sessions []api.Session
}

type sessionHistoryLoadedMsg struct{ sessions []api.SessionHistoryEntry }
type sessionDetailLoadedMsg struct{ detail *api.SessionDetail }
type sessionDetailFailedMsg struct{ err error }
type sessionAmendedMsg struct{ id string }
type scratchpadsLoadedMsg struct{ pads []api.ScratchPad }
type stashesLoadedMsg struct{ stashes []api.Stash }
type opsLoadedMsg struct{ ops []api.Op }
type contextLoadedMsg struct{ ctx *api.ActiveContext }
type timerLoadedMsg struct{ timer *api.TimerState }
type healthLoadedMsg struct{ health *api.Health }
type settingsLoadedMsg struct{ settings *api.CoreSettings }
type kernelInfoLoadedMsg struct{ info *api.KernelInfo }
type kernelEventMsg struct{ event api.KernelEvent }
type kernelShutdownMsg struct{}
type devSeededMsg struct{}
type devClearedMsg struct{}
type timerTickMsg struct{ seq int }
type healthTickMsg struct{}
type errMsg struct{ err error }
type clearStatusMsg struct{ seq int }
