package app

import (
	"encoding/json"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	"crona/tui/internal/logger"

	tea "github.com/charmbracelet/bubbletea"
)

func handleKernelEvent(m Model, event api.KernelEvent) (Model, tea.Cmd) {
	logger.Infof("kernel event: %s", event.Type)

	switch event.Type {
	case "repo.created", "repo.updated", "repo.deleted":
		return m, loadRepos(m.client)
	case "stream.created", "stream.updated", "stream.deleted":
		if m.context != nil && m.context.RepoID != nil {
			return m, loadStreams(m.client, *m.context.RepoID)
		}
	case "issue.created", "issue.updated", "issue.deleted":
		cmds := []tea.Cmd{loadAllIssues(m.client), loadDailySummary(m.client, m.dashboardDate)}
		if m.context != nil && m.context.StreamID != nil {
			cmds = append(cmds, loadIssues(m.client, *m.context.StreamID))
		}
		return m, tea.Batch(cmds...)
	case "habit.created", "habit.updated", "habit.deleted", "habit.completed", "habit.uncompleted":
		cmds := []tea.Cmd{loadDueHabits(m.client, m.currentDashboardDate())}
		if m.context != nil && m.context.StreamID != nil {
			cmds = append(cmds, loadHabits(m.client, *m.context.StreamID))
		}
		return m, tea.Batch(cmds...)
	case "checkin.updated", "checkin.deleted":
		return m, loadWellbeing(m.client, m.currentWellbeingDate())
	case "scratchpad.created", "scratchpad.updated", "scratchpad.deleted":
		return m, loadScratchpads(m.client)
	case "session.started", "session.stopped":
		return m, tea.Batch(loadTimer(m.client), loadContext(m.client), loadSessionHistory(m.client, 200))
	case "stash.created", "stash.applied", "stash.dropped":
		return m, tea.Batch(loadStashes(m.client), loadContext(m.client), loadTimer(m.client))
	case "context.repo.changed", "context.stream.changed", "context.issue.changed", "context.cleared":
		var payload sharedtypes.ContextChangedPayload
		_ = json.Unmarshal(event.Payload, &payload)

		cmds := []tea.Cmd{loadContext(m.client)}
		if payload.RepoID != nil {
			cmds = append(cmds, loadStreams(m.client, *payload.RepoID))
		} else {
			m.streams = nil
			m.issues = nil
			m.cursor[PaneStreams] = 0
			m.cursor[PaneIssues] = 0
		}
		if payload.StreamID != nil {
			cmds = append(cmds, loadIssues(m.client, *payload.StreamID), loadHabits(m.client, *payload.StreamID))
		} else if payload.RepoID != nil {
			m.issues = nil
			m.habits = nil
			m.cursor[PaneIssues] = 0
			m.cursor[PaneHabits] = 0
		}
		return m, tea.Batch(cmds...)
	case "timer.state":
		var timer api.TimerState
		if err := json.Unmarshal(event.Payload, &timer); err == nil {
			m.timer = &timer
			m.elapsed = 0
			m.timerTickSeq++
			if timer.State != "idle" {
				if m.view != ViewScratch {
					m.view = ViewSessionActive
				}
				m.pane = viewDefaultPane[m.view]
				return m, tickAfter(m.timerTickSeq)
			} else if m.view == ViewSessionActive {
				m.view = ViewDaily
				m.pane = viewDefaultPane[m.view]
			}
		}
	case "timer.boundary":
		m.elapsed = 0
		return m, loadTimer(m.client)
	case "ops.created":
		return m, loadOps(m.client, m.currentOpsLimit())
	}

	return m, nil
}
