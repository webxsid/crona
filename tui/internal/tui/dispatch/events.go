package dispatch

import (
	"encoding/json"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	"crona/tui/internal/logger"
	uistate "crona/tui/internal/tui/state"

	tea "github.com/charmbracelet/bubbletea"
)

type EventState struct {
	View                   uistate.View
	Pane                   uistate.Pane
	Cursor                 map[uistate.Pane]int
	Streams                []api.Stream
	Issues                 []api.Issue
	SelectedIssueID        *int64
	Habits                 []api.Habit
	SelectedHabitHistoryID *int64
	Context                *api.ActiveContext
	Timer                  *api.TimerState
	Elapsed                int
	TimerTickSeq           int
	CurrentDash            string
	CurrentRollupStart     string
	CurrentRollupEnd       string
	CurrentWell            string
	WellbeingWindowDays    int
	CurrentOpsLim          int
}

type EventDeps struct {
	LoadRepos                func() tea.Cmd
	LoadStreams              func(repoID int64) tea.Cmd
	LoadIssues               func(streamID int64) tea.Cmd
	LoadIssuesSelecting      func(streamID, selectedIssueID int64) tea.Cmd
	LoadHabits               func(streamID int64) tea.Cmd
	LoadAllHabits            func() tea.Cmd
	LoadAllIssues            func() tea.Cmd
	LoadAllIssuesSelecting   func(selectedIssueID int64) tea.Cmd
	LoadDailySummary         func(date string) tea.Cmd
	LoadDueHabits            func(date string) tea.Cmd
	LoadHabitHistory         func(*api.ActiveContext, *int64) tea.Cmd
	LoadWellbeing            func(date string, windowDays int) tea.Cmd
	LoadRollupSummaries      func(start, end string) tea.Cmd
	LoadScratchpads          func() tea.Cmd
	LoadSessionHistoryFor200 func(EventState) tea.Cmd
	LoadStashes              func() tea.Cmd
	LoadContext              func() tea.Cmd
	LoadTimer                func() tea.Cmd
	LoadAlertStatus          func() tea.Cmd
	LoadUpdateStatus         func() tea.Cmd
	LoadOps                  func(limit int) tea.Cmd
	TickAfter                func(seq int) tea.Cmd
}

func HandleEvent(state EventState, deps EventDeps, event api.KernelEvent) (EventState, tea.Cmd) {
	logger.Infof("kernel event: %s", event.Type)
	switch event.Type {
	case "repo.created", "repo.updated", "repo.deleted":
		return state, deps.LoadRepos()
	case "stream.created", "stream.updated", "stream.deleted":
		if state.Context != nil && state.Context.RepoID != nil {
			return state, deps.LoadStreams(*state.Context.RepoID)
		}
		return state, nil
	case "issue.created", "issue.updated", "issue.deleted":
		cmds := []tea.Cmd{deps.LoadDailySummary(state.CurrentDash), deps.LoadRollupSummaries(state.CurrentRollupStart, state.CurrentRollupEnd)}
		if state.Context != nil && state.Context.StreamID != nil {
			if state.SelectedIssueID != nil {
				cmds = append(cmds, deps.LoadIssuesSelecting(*state.Context.StreamID, *state.SelectedIssueID))
			} else {
				cmds = append(cmds, deps.LoadIssues(*state.Context.StreamID))
			}
		}
		if state.SelectedIssueID != nil {
			cmds = append(cmds, deps.LoadAllIssuesSelecting(*state.SelectedIssueID))
		} else {
			cmds = append(cmds, deps.LoadAllIssues())
		}
		return state, tea.Batch(cmds...)
	case "habit.created", "habit.updated", "habit.deleted", "habit.completed", "habit.uncompleted":
		cmds := []tea.Cmd{deps.LoadDueHabits(state.CurrentDash), deps.LoadDailySummary(state.CurrentDash), deps.LoadWellbeing(state.CurrentWell, state.WellbeingWindowDays), deps.LoadRollupSummaries(state.CurrentRollupStart, state.CurrentRollupEnd), deps.LoadAllHabits()}
		if state.Context != nil && state.Context.StreamID != nil {
			cmds = append(cmds, deps.LoadHabits(*state.Context.StreamID))
		}
		if state.View == uistate.ViewHabitHistory {
			cmds = append(cmds, deps.LoadHabitHistory(state.Context, state.SelectedHabitHistoryID))
		}
		return state, tea.Batch(cmds...)
	case "checkin.updated", "checkin.deleted":
		return state, tea.Batch(deps.LoadWellbeing(state.CurrentWell, state.WellbeingWindowDays), deps.LoadRollupSummaries(state.CurrentRollupStart, state.CurrentRollupEnd))
	case "scratchpad.created", "scratchpad.updated", "scratchpad.deleted":
		return state, deps.LoadScratchpads()
	case "session.started", "session.stopped":
		return state, tea.Batch(deps.LoadTimer(), deps.LoadContext(), deps.LoadSessionHistoryFor200(state), deps.LoadRollupSummaries(state.CurrentRollupStart, state.CurrentRollupEnd))
	case "stash.created", "stash.applied", "stash.dropped":
		return state, tea.Batch(deps.LoadStashes(), deps.LoadContext(), deps.LoadTimer())
	case "context.repo.changed", "context.stream.changed", "context.issue.changed", "context.cleared":
		var payload sharedtypes.ContextChangedPayload
		_ = json.Unmarshal(event.Payload, &payload)
		cmds := []tea.Cmd{deps.LoadContext(), deps.LoadRollupSummaries(state.CurrentRollupStart, state.CurrentRollupEnd)}
		if payload.RepoID != nil {
			cmds = append(cmds, deps.LoadStreams(*payload.RepoID))
		} else {
			state.Streams = nil
			state.Issues = nil
			state.Cursor[uistate.PaneStreams] = 0
			state.Cursor[uistate.PaneIssues] = 0
		}
		if payload.StreamID != nil {
			cmds = append(cmds, deps.LoadIssues(*payload.StreamID), deps.LoadHabits(*payload.StreamID))
		} else if payload.RepoID != nil {
			state.Issues = nil
			state.Habits = nil
			state.Cursor[uistate.PaneIssues] = 0
			state.Cursor[uistate.PaneHabits] = 0
		}
		return state, tea.Batch(cmds...)
	case "timer.state":
		var timer api.TimerState
		if err := json.Unmarshal(event.Payload, &timer); err == nil {
			state.Timer = &timer
			state.Elapsed = 0
			state.TimerTickSeq++
			if timer.State != "idle" {
				if state.View != uistate.ViewScratch && state.View != uistate.ViewSessionHistory {
					state.View = uistate.ViewSessionActive
				}
				state.Pane = uistate.DefaultPane(state.View)
				return state, tea.Batch(deps.TickAfter(state.TimerTickSeq), deps.LoadSessionHistoryFor200(state))
			}
			if state.View == uistate.ViewSessionActive {
				state.View = uistate.ViewDaily
				state.Pane = uistate.DefaultPane(state.View)
			}
			return state, deps.LoadSessionHistoryFor200(state)
		}
		return state, nil
	case "timer.boundary":
		state.Elapsed = 0
		return state, tea.Batch(deps.LoadTimer(), deps.LoadRollupSummaries(state.CurrentRollupStart, state.CurrentRollupEnd))
	case "update.status":
		return state, deps.LoadUpdateStatus()
	case "ops.created":
		return state, deps.LoadOps(state.CurrentOpsLim)
	default:
		return state, nil
	}
}
