package commands

import (
	"time"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	"crona/tui/internal/logger"

	tea "github.com/charmbracelet/bubbletea"
)

func LoadRepos(c *api.Client) tea.Cmd {
	return func() tea.Msg {
		repos, err := c.ListRepos()
		if err != nil {
			logger.Errorf("loadRepos: %v", err)
			return ErrMsg{Err: err}
		}
		return ReposLoadedMsg{Repos: repos}
	}
}

func LoadStreams(c *api.Client, repoID int64) tea.Cmd {
	return func() tea.Msg {
		streams, err := c.ListStreams(repoID)
		if err != nil {
			logger.Errorf("loadStreams(%d): %v", repoID, err)
			return ErrMsg{Err: err}
		}
		return StreamsLoadedMsg{Streams: streams}
	}
}

func LoadAllHabits(c *api.Client) tea.Cmd {
	return func() tea.Msg {
		habits, err := c.ListAllHabits()
		if err != nil {
			logger.Errorf("loadAllHabits: %v", err)
			return ErrMsg{Err: err}
		}
		return AllHabitsLoadedMsg{Habits: habits}
	}
}

func LoadIssues(c *api.Client, streamID int64) tea.Cmd {
	return func() tea.Msg {
		issues, err := c.ListIssues(streamID)
		if err != nil {
			logger.Errorf("loadIssues(%d): %v", streamID, err)
			return ErrMsg{Err: err}
		}
		return IssuesLoadedMsg{StreamID: streamID, Issues: issues}
	}
}

func LoadIssuesSelecting(c *api.Client, streamID, selectedIssueID int64) tea.Cmd {
	return func() tea.Msg {
		issues, err := c.ListIssues(streamID)
		if err != nil {
			logger.Errorf("loadIssues(%d): %v", streamID, err)
			return ErrMsg{Err: err}
		}
		return IssuesLoadedMsg{StreamID: streamID, Issues: issues, SelectedIssueID: int64Ptr(selectedIssueID)}
	}
}

func LoadHabits(c *api.Client, streamID int64) tea.Cmd {
	return func() tea.Msg {
		habits, err := c.ListHabits(streamID)
		if err != nil {
			logger.Errorf("loadHabits(%d): %v", streamID, err)
			return ErrMsg{Err: err}
		}
		return HabitsLoadedMsg{StreamID: streamID, Habits: habits}
	}
}

func LoadHabitHistory(c *api.Client, context *api.ActiveContext, selectedHabitHistoryID *int64) tea.Cmd {
	return func() tea.Msg {
		var repoID, streamID *int64
		if context != nil {
			repoID = context.RepoID
			streamID = context.StreamID
		}
		completions, err := c.ListHabitHistory(repoID, streamID)
		if err != nil {
			logger.Errorf("loadHabitHistory: %v", err)
			return ErrMsg{Err: err}
		}
		return HabitHistoryLoadedMsg{Completions: completions, SelectedHabitHistoryID: selectedHabitHistoryID, Scope: context}
	}
}

func LoadAllIssues(c *api.Client) tea.Cmd {
	return func() tea.Msg {
		issues, err := c.ListAllIssues()
		if err != nil {
			logger.Errorf("loadAllIssues: %v", err)
			return ErrMsg{Err: err}
		}
		return AllIssuesLoadedMsg{Issues: issues}
	}
}

func LoadAllIssuesSelecting(c *api.Client, selectedIssueID int64) tea.Cmd {
	return func() tea.Msg {
		issues, err := c.ListAllIssues()
		if err != nil {
			logger.Errorf("loadAllIssues: %v", err)
			return ErrMsg{Err: err}
		}
		return AllIssuesLoadedMsg{Issues: issues, SelectedIssueID: int64Ptr(selectedIssueID)}
	}
}

func LoadDueHabits(c *api.Client, date string) tea.Cmd {
	return func() tea.Msg {
		habits, err := c.ListDueHabits(date)
		if err != nil {
			logger.Errorf("loadDueHabits: %v", err)
			return ErrMsg{Err: err}
		}
		return DueHabitsLoadedMsg{Habits: habits}
	}
}

func LoadDailySummary(c *api.Client, date string) tea.Cmd {
	return func() tea.Msg {
		summary, err := c.GetDailySummary(date)
		if err != nil {
			logger.Errorf("loadDailySummary: %v", err)
			return ErrMsg{Err: err}
		}
		return DailySummaryLoadedMsg{Summary: summary}
	}
}

func int64Ptr(v int64) *int64 { return &v }

func LoadDailyPlan(c *api.Client, date string) tea.Cmd {
	return func() tea.Msg {
		plan, err := c.GetDailyPlan(date)
		if err != nil {
			logger.Errorf("loadDailyPlan: %v", err)
			return ErrMsg{Err: err}
		}
		return DailyPlanLoadedMsg{Plan: plan}
	}
}

func LoadDailyCheckIn(c *api.Client, date string) tea.Cmd {
	return func() tea.Msg {
		checkIn, err := c.GetDailyCheckIn(date)
		if err != nil {
			logger.Errorf("loadDailyCheckIn: %v", err)
			return ErrMsg{Err: err}
		}
		return DailyCheckInLoadedMsg{CheckIn: checkIn}
	}
}

func LoadMetricsRange(c *api.Client, start, end string) tea.Cmd {
	return func() tea.Msg {
		days, err := c.GetMetricsRange(start, end)
		if err != nil {
			logger.Errorf("loadMetricsRange: %v", err)
			return ErrMsg{Err: err}
		}
		return MetricsRangeLoadedMsg{Days: days}
	}
}

func LoadMetricsRollup(c *api.Client, start, end string) tea.Cmd {
	return func() tea.Msg {
		rollup, err := c.GetMetricsRollup(start, end)
		if err != nil {
			logger.Errorf("loadMetricsRollup: %v", err)
			return ErrMsg{Err: err}
		}
		return MetricsRollupLoadedMsg{Rollup: rollup}
	}
}

func LoadMetricsStreaks(c *api.Client, start, end string) tea.Cmd {
	return func() tea.Msg {
		streaks, err := c.GetMetricsStreaks(start, end)
		if err != nil {
			logger.Errorf("loadMetricsStreaks: %v", err)
			return ErrMsg{Err: err}
		}
		return StreaksLoadedMsg{Streaks: streaks}
	}
}

func LoadMetricsLifetimeStreaks(c *api.Client, date string) tea.Cmd {
	return func() tea.Msg {
		streaks, err := c.GetMetricsLifetimeStreaks(date)
		if err != nil {
			logger.Errorf("loadMetricsLifetimeStreaks: %v", err)
			return ErrMsg{Err: err}
		}
		return StreaksLoadedMsg{Streaks: streaks}
	}
}

func LoadIssueSessions(c *api.Client, issueID int64) tea.Cmd {
	return func() tea.Msg {
		sessions, err := c.ListSessionsByIssue(issueID)
		if err != nil {
			logger.Errorf("loadIssueSessions(%d): %v", issueID, err)
			return ErrMsg{Err: err}
		}
		return IssueSessionsLoadedMsg{IssueID: issueID, Sessions: sessions}
	}
}

func LoadSessionHistory(c *api.Client, issueID *int64, limit int) tea.Cmd {
	return func() tea.Msg {
		sessions, err := c.ListSessionHistory(issueID, limit)
		if err != nil {
			logger.Errorf("loadSessionHistory: %v", err)
			return ErrMsg{Err: err}
		}
		return SessionHistoryLoadedMsg{Sessions: sessions}
	}
}

func LoadSessionDetail(c *api.Client, id string) tea.Cmd {
	return func() tea.Msg {
		detail, err := c.GetSessionDetail(id)
		if err != nil {
			logger.Errorf("loadSessionDetail(%s): %v", id, err)
			return SessionDetailFailedMsg{Err: err}
		}
		return SessionDetailLoadedMsg{Detail: detail}
	}
}

func LoadScratchpads(c *api.Client) tea.Cmd {
	return func() tea.Msg {
		pads, err := c.ListScratchpads()
		if err != nil {
			logger.Errorf("loadScratchpads: %v", err)
			return ErrMsg{Err: err}
		}
		return ScratchpadsLoadedMsg{Pads: pads}
	}
}

func LoadStashes(c *api.Client) tea.Cmd {
	return func() tea.Msg {
		stashes, err := c.ListStashes()
		if err != nil {
			logger.Errorf("loadStashes: %v", err)
			return ErrMsg{Err: err}
		}
		return StashesLoadedMsg{Stashes: stashes}
	}
}

func LoadOps(c *api.Client, limit int) tea.Cmd {
	return func() tea.Msg {
		ops, err := c.ListOps(limit)
		if err != nil {
			logger.Errorf("loadOps: %v", err)
			return ErrMsg{Err: err}
		}
		return OpsLoadedMsg{Ops: ops}
	}
}

func LoadContext(c *api.Client) tea.Cmd {
	return func() tea.Msg {
		ctx, err := c.GetContext()
		if err != nil {
			logger.Errorf("loadContext: %v", err)
			return ErrMsg{Err: err}
		}
		return ContextLoadedMsg{Ctx: ctx}
	}
}

func LoadTimer(c *api.Client) tea.Cmd {
	return func() tea.Msg {
		t, err := c.GetTimerState()
		if err != nil {
			logger.Errorf("loadTimer: %v", err)
			return ErrMsg{Err: err}
		}
		return TimerLoadedMsg{Timer: t}
	}
}

func LoadHealth(c *api.Client) tea.Cmd {
	return func() tea.Msg {
		h, err := c.GetHealth()
		if err != nil {
			logger.Errorf("loadHealth: %v", err)
			return ErrMsg{Err: err}
		}
		return HealthLoadedMsg{Health: h}
	}
}

func LoadUpdateStatus(c *api.Client) tea.Cmd {
	return func() tea.Msg {
		status, err := c.GetUpdateStatus()
		if err != nil {
			logger.Errorf("loadUpdateStatus: %v", err)
			return ErrMsg{Err: err}
		}
		return UpdateStatusLoadedMsg{Status: status}
	}
}

func LoadSettings(c *api.Client) tea.Cmd {
	return func() tea.Msg {
		settings, err := c.GetSettings()
		if err != nil {
			logger.Errorf("loadSettings: %v", err)
			return ErrMsg{Err: err}
		}
		return SettingsLoadedMsg{Settings: settings}
	}
}

func LoadKernelInfo(c *api.Client) tea.Cmd {
	return func() tea.Msg {
		info, err := c.GetKernelInfo()
		if err != nil {
			logger.Errorf("loadKernelInfo: %v", err)
			return ErrMsg{Err: err}
		}
		return KernelInfoLoadedMsg{Info: info}
	}
}

func LoadExportAssets(c *api.Client) tea.Cmd {
	return func() tea.Msg {
		assets, err := c.GetExportAssets()
		if err != nil {
			logger.Errorf("loadExportAssets: %v", err)
			return ErrMsg{Err: err}
		}
		return ExportAssetsLoadedMsg{Assets: assets}
	}
}

func LoadExportReports(c *api.Client) tea.Cmd {
	return func() tea.Msg {
		reports, err := c.ListExportReports()
		if err != nil {
			logger.Errorf("loadExportReports: %v", err)
			return ErrMsg{Err: err}
		}
		return ExportReportsLoadedMsg{Reports: reports}
	}
}

func TickAfter(seq int) tea.Cmd {
	return tea.Tick(time.Second, func(_ time.Time) tea.Msg {
		return TimerTickMsg{Seq: seq}
	})
}

func HealthTickAfter() tea.Cmd {
	return tea.Tick(5*time.Second, func(_ time.Time) tea.Msg {
		return HealthTickMsg{}
	})
}

func ClearStatusAfter(seq int, d time.Duration) tea.Cmd {
	return tea.Tick(d, func(_ time.Time) tea.Msg {
		return ClearStatusMsg{Seq: seq}
	})
}

func WaitForEvent(ch <-chan api.KernelEvent) tea.Cmd {
	return func() tea.Msg {
		event, ok := <-ch
		if !ok {
			return nil
		}
		return KernelEventMsg{Event: event}
	}
}

func LoadWellbeing(c *api.Client, date string) tea.Cmd {
	start := shiftISODate(date, -6)
	return tea.Batch(
		LoadDailyPlan(c, date),
		LoadDailyCheckIn(c, date),
		LoadMetricsRange(c, start, date),
		LoadMetricsRollup(c, start, date),
		LoadMetricsLifetimeStreaks(c, date),
		LoadDashboardSummaries(c, date),
	)
}

func LoadDashboardSummaries(c *api.Client, date string) tea.Cmd {
	return LoadRollupSummaries(c, shiftISODate(date, -6), date)
}

func LoadRollupSummaries(c *api.Client, start, end string) tea.Cmd {
	return tea.Batch(
		LoadDashboardWindow(c, start, end),
		LoadFocusScore(c, start, end, 7),
		LoadDistribution(c, start, end, string(sharedtypes.DistributionGroupRepo)),
		LoadDistribution(c, start, end, string(sharedtypes.DistributionGroupStream)),
		LoadDistribution(c, start, end, string(sharedtypes.DistributionGroupIssue)),
		LoadDistribution(c, start, end, string(sharedtypes.DistributionGroupSegmentType)),
		LoadGoalProgress(c, start, end),
	)
}

func LoadDashboardWindow(c *api.Client, start, end string) tea.Cmd {
	return func() tea.Msg {
		repoID, streamID, issueID, err := currentContextScope(c)
		if err != nil {
			logger.Errorf("loadDashboardWindow: %v", err)
			return ErrMsg{Err: err}
		}
		summary, err := c.GetDashboardWindowSummary(start, end, repoID, streamID, issueID)
		if err != nil {
			logger.Errorf("loadDashboardWindow: %v", err)
			return ErrMsg{Err: err}
		}
		return DashboardWindowLoadedMsg{Summary: summary}
	}
}

func LoadFocusScore(c *api.Client, start, end string, windowDays int) tea.Cmd {
	return func() tea.Msg {
		summary, err := c.GetFocusScoreSummary(start, end)
		if err != nil {
			logger.Errorf("loadFocusScore: %v", err)
			return ErrMsg{Err: err}
		}
		return FocusScoreLoadedMsg{WindowDays: windowDays, Summary: summary}
	}
}

func LoadDistribution(c *api.Client, start, end, groupBy string) tea.Cmd {
	return func() tea.Msg {
		repoID, streamID, issueID, err := currentContextScope(c)
		if err != nil {
			logger.Errorf("loadDistribution(%s): %v", groupBy, err)
			return ErrMsg{Err: err}
		}
		summary, err := c.GetTimeDistributionSummary(start, end, groupBy, repoID, streamID, issueID)
		if err != nil {
			logger.Errorf("loadDistribution(%s): %v", groupBy, err)
			return ErrMsg{Err: err}
		}
		return DistributionLoadedMsg{GroupBy: groupBy, Summary: summary}
	}
}

func LoadGoalProgress(c *api.Client, start, end string) tea.Cmd {
	return func() tea.Msg {
		repoID, streamID, issueID, err := currentContextScope(c)
		if err != nil {
			logger.Errorf("loadGoalProgress: %v", err)
			return ErrMsg{Err: err}
		}
		groupBy := goalProgressGroupForScope(repoID, streamID, issueID)
		summary, err := c.GetGoalProgressSummary(start, end, groupBy, repoID, streamID, issueID)
		if err != nil {
			logger.Errorf("loadGoalProgress: %v", err)
			return ErrMsg{Err: err}
		}
		return GoalProgressLoadedMsg{Summary: summary}
	}
}

func SetRollupRange(start, end string) tea.Cmd {
	return func() tea.Msg {
		return RollupRangeChangedMsg{Start: start, End: end}
	}
}

func shiftISODate(date string, days int) string {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return date
	}
	return t.AddDate(0, 0, days).Format("2006-01-02")
}

func currentContextScope(c *api.Client) (*int64, *int64, *int64, error) {
	ctx, err := c.GetContext()
	if err != nil {
		return nil, nil, nil, err
	}
	if ctx == nil {
		return nil, nil, nil, nil
	}
	return ctx.RepoID, ctx.StreamID, ctx.IssueID, nil
}

func goalProgressGroupForScope(repoID, streamID, issueID *int64) string {
	switch {
	case issueID != nil:
		return string(sharedtypes.GoalProgressGroupIssue)
	case streamID != nil:
		return string(sharedtypes.GoalProgressGroupStream)
	default:
		return string(sharedtypes.GoalProgressGroupRepo)
	}
}
