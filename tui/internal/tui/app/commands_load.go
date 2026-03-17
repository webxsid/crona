package app

import (
	"time"

	"crona/tui/internal/api"
	"crona/tui/internal/logger"

	tea "github.com/charmbracelet/bubbletea"
)

func loadRepos(c *api.Client) tea.Cmd {
	return func() tea.Msg {
		repos, err := c.ListRepos()
		if err != nil {
			logger.Errorf("loadRepos: %v", err)
			return errMsg{err}
		}
		return reposLoadedMsg{repos}
	}
}

func loadStreams(c *api.Client, repoID int64) tea.Cmd {
	return func() tea.Msg {
		streams, err := c.ListStreams(repoID)
		if err != nil {
			logger.Errorf("loadStreams(%d): %v", repoID, err)
			return errMsg{err}
		}
		return streamsLoadedMsg{streams}
	}
}

func loadIssues(c *api.Client, streamID int64) tea.Cmd {
	return func() tea.Msg {
		issues, err := c.ListIssues(streamID)
		if err != nil {
			logger.Errorf("loadIssues(%d): %v", streamID, err)
			return errMsg{err}
		}
		return issuesLoadedMsg{streamID: streamID, issues: issues}
	}
}

func loadHabits(c *api.Client, streamID int64) tea.Cmd {
	return func() tea.Msg {
		habits, err := c.ListHabits(streamID)
		if err != nil {
			logger.Errorf("loadHabits(%d): %v", streamID, err)
			return errMsg{err}
		}
		return habitsLoadedMsg{streamID: streamID, habits: habits}
	}
}

func loadAllIssues(c *api.Client) tea.Cmd {
	return func() tea.Msg {
		issues, err := c.ListAllIssues()
		if err != nil {
			logger.Errorf("loadAllIssues: %v", err)
			return errMsg{err}
		}
		return allIssuesLoadedMsg{issues}
	}
}

func loadDueHabits(c *api.Client, date string) tea.Cmd {
	return func() tea.Msg {
		habits, err := c.ListDueHabits(date)
		if err != nil {
			logger.Errorf("loadDueHabits: %v", err)
			return errMsg{err}
		}
		return dueHabitsLoadedMsg{habits: habits}
	}
}

func loadDailySummary(c *api.Client, date string) tea.Cmd {
	return func() tea.Msg {
		summary, err := c.GetDailySummary(date)
		if err != nil {
			logger.Errorf("loadDailySummary: %v", err)
			return errMsg{err}
		}
		return dailySummaryLoadedMsg{summary}
	}
}

func loadDailyCheckIn(c *api.Client, date string) tea.Cmd {
	return func() tea.Msg {
		checkIn, err := c.GetDailyCheckIn(date)
		if err != nil {
			logger.Errorf("loadDailyCheckIn: %v", err)
			return errMsg{err}
		}
		return dailyCheckInLoadedMsg{checkIn}
	}
}

func loadMetricsRange(c *api.Client, start, end string) tea.Cmd {
	return func() tea.Msg {
		days, err := c.GetMetricsRange(start, end)
		if err != nil {
			logger.Errorf("loadMetricsRange: %v", err)
			return errMsg{err}
		}
		return metricsRangeLoadedMsg{days}
	}
}

func loadMetricsRollup(c *api.Client, start, end string) tea.Cmd {
	return func() tea.Msg {
		rollup, err := c.GetMetricsRollup(start, end)
		if err != nil {
			logger.Errorf("loadMetricsRollup: %v", err)
			return errMsg{err}
		}
		return metricsRollupLoadedMsg{rollup}
	}
}

func loadMetricsStreaks(c *api.Client, start, end string) tea.Cmd {
	return func() tea.Msg {
		streaks, err := c.GetMetricsStreaks(start, end)
		if err != nil {
			logger.Errorf("loadMetricsStreaks: %v", err)
			return errMsg{err}
		}
		return streaksLoadedMsg{streaks}
	}
}

func loadIssueSessions(c *api.Client, issueID int64) tea.Cmd {
	return func() tea.Msg {
		sessions, err := c.ListSessionsByIssue(issueID)
		if err != nil {
			logger.Errorf("loadIssueSessions(%d): %v", issueID, err)
			return errMsg{err}
		}
		return issueSessionsLoadedMsg{issueID: issueID, sessions: sessions}
	}
}

func loadSessionHistory(c *api.Client, limit int) tea.Cmd {
	return func() tea.Msg {
		sessions, err := c.ListSessionHistory(limit)
		if err != nil {
			logger.Errorf("loadSessionHistory: %v", err)
			return errMsg{err}
		}
		return sessionHistoryLoadedMsg{sessions}
	}
}

func loadSessionDetail(c *api.Client, id string) tea.Cmd {
	return func() tea.Msg {
		detail, err := c.GetSessionDetail(id)
		if err != nil {
			logger.Errorf("loadSessionDetail(%s): %v", id, err)
			return sessionDetailFailedMsg{err: err}
		}
		return sessionDetailLoadedMsg{detail: detail}
	}
}

func loadScratchpads(c *api.Client) tea.Cmd {
	return func() tea.Msg {
		pads, err := c.ListScratchpads()
		if err != nil {
			logger.Errorf("loadScratchpads: %v", err)
			return errMsg{err}
		}
		return scratchpadsLoadedMsg{pads}
	}
}

func loadStashes(c *api.Client) tea.Cmd {
	return func() tea.Msg {
		stashes, err := c.ListStashes()
		if err != nil {
			logger.Errorf("loadStashes: %v", err)
			return errMsg{err}
		}
		return stashesLoadedMsg{stashes}
	}
}

func loadOps(c *api.Client, limit int) tea.Cmd {
	return func() tea.Msg {
		ops, err := c.ListOps(limit)
		if err != nil {
			logger.Errorf("loadOps: %v", err)
			return errMsg{err}
		}
		return opsLoadedMsg{ops}
	}
}

func loadContext(c *api.Client) tea.Cmd {
	return func() tea.Msg {
		ctx, err := c.GetContext()
		if err != nil {
			logger.Errorf("loadContext: %v", err)
			return errMsg{err}
		}
		return contextLoadedMsg{ctx}
	}
}

func loadTimer(c *api.Client) tea.Cmd {
	return func() tea.Msg {
		t, err := c.GetTimerState()
		if err != nil {
			logger.Errorf("loadTimer: %v", err)
			return errMsg{err}
		}
		return timerLoadedMsg{t}
	}
}

func loadHealth(c *api.Client) tea.Cmd {
	return func() tea.Msg {
		h, err := c.GetHealth()
		if err != nil {
			logger.Errorf("loadHealth: %v", err)
			return errMsg{err}
		}
		return healthLoadedMsg{h}
	}
}

func loadSettings(c *api.Client) tea.Cmd {
	return func() tea.Msg {
		settings, err := c.GetSettings()
		if err != nil {
			logger.Errorf("loadSettings: %v", err)
			return errMsg{err}
		}
		return settingsLoadedMsg{settings}
	}
}

func loadKernelInfo(c *api.Client) tea.Cmd {
	return func() tea.Msg {
		info, err := c.GetKernelInfo()
		if err != nil {
			logger.Errorf("loadKernelInfo: %v", err)
			return errMsg{err}
		}
		return kernelInfoLoadedMsg{info}
	}
}

func tickAfter(seq int) tea.Cmd {
	return tea.Tick(time.Second, func(_ time.Time) tea.Msg {
		return timerTickMsg{seq: seq}
	})
}

func healthTickAfter() tea.Cmd {
	return tea.Tick(5*time.Second, func(_ time.Time) tea.Msg {
		return healthTickMsg{}
	})
}

func clearStatusAfter(seq int, d time.Duration) tea.Cmd {
	return tea.Tick(d, func(_ time.Time) tea.Msg {
		return clearStatusMsg{seq: seq}
	})
}

func waitForEvent(ch <-chan api.KernelEvent) tea.Cmd {
	return func() tea.Msg {
		event, ok := <-ch
		if !ok {
			return nil
		}
		return kernelEventMsg{event}
	}
}

func loadWellbeing(c *api.Client, date string) tea.Cmd {
	start := shiftISODate(date, -6)
	return tea.Batch(
		loadDailyCheckIn(c, date),
		loadMetricsRange(c, start, date),
		loadMetricsRollup(c, start, date),
		loadMetricsStreaks(c, start, date),
	)
}
