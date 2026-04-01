package commands

import (
	"time"

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

func WaitForUpdateInstall(ch <-chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return nil
		}
		return msg
	}
}

func LoadWellbeing(c *api.Client, date string) tea.Cmd {
	start := shiftISODate(date, -6)
	return tea.Batch(
		LoadDailyPlan(c, date),
		LoadDailyCheckIn(c, date),
		LoadMetricsRange(c, start, date),
		LoadMetricsRollup(c, start, date),
		LoadMetricsStreaks(c, start, date),
	)
}

func shiftISODate(date string, days int) string {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return date
	}
	return t.AddDate(0, 0, days).Format("2006-01-02")
}
