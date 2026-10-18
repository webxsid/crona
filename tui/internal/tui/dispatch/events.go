package dispatch

import (
	"encoding/json"
	"slices"

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
	Dialog                 string
	DialogParent           string
	Streams                []api.Stream
	Issues                 []api.Issue
	SelectedIssueID        *int64
	Habits                 []api.Habit
	SelectedHabitHistoryID *int64
	Context                *api.ActiveContext
	Timer                  *api.TimerState
	Elapsed                int
	TimerTickSeq           int
	// Date is the stored shared date ("" follows today), CurrentDate the date it resolves
	// to, and Today the logical today before this event.
	Date               string
	CurrentDate        string
	Today              string
	ReviewWindowDays   int
	ReviewWindowEnd    string
	MomentumWindowDays int
	// ReviewStale marks Review's window as changed while it was off screen; opening Review
	// reloads it.
	ReviewStale   bool
	CurrentOpsLim int
}

type EventDeps struct {
	LoadSummarySnapshot      func(string) tea.Cmd
	LoadIssueSessions        func(int64) tea.Cmd
	LoadRepos                func() tea.Cmd
	LoadStreams              func(repoID int64) tea.Cmd
	LoadIssues               func(streamID int64) tea.Cmd
	LoadIssuesSelecting      func(streamID, selectedIssueID int64) tea.Cmd
	LoadHabits               func(streamID int64) tea.Cmd
	LoadAllHabits            func() tea.Cmd
	LoadAllStreams           func() tea.Cmd
	LoadAllIssues            func() tea.Cmd
	LoadAllIssuesSelecting   func(selectedIssueID int64) tea.Cmd
	LoadDailySummary         func(date string) tea.Cmd
	LoadDailyPlan            func(date string) tea.Cmd
	LoadDueHabits            func(date string) tea.Cmd
	LoadTodayHabits          func(date string) tea.Cmd
	LoadDailyStreaks         func(date string) tea.Cmd
	LoadHabitHistory         func(*api.ActiveContext, *int64) tea.Cmd
	LoadWellbeing            func(date string) tea.Cmd
	LoadMomentumRange        func(date string, windowDays int) tea.Cmd
	LoadReviewWindow         func(start, end string) tea.Cmd
	LoadSessionHistoryFor200 func(EventState) tea.Cmd
	LoadContext              func() tea.Cmd
	LoadTimer                func() tea.Cmd
	LoadAlertStatus          func() tea.Cmd
	LoadSettings             func() tea.Cmd
	LoadUpdateStatus         func() tea.Cmd
	LoadOps                  func(limit int) tea.Cmd
	TickAfter                func(seq int) tea.Cmd
}

// HandleEvent reloads what an event changed. Review's window is the costliest reload, so off
// Review it is only marked stale, and loads when Review opens.
func HandleEvent(state EventState, deps EventDeps, event api.KernelEvent) (EventState, tea.Cmd) {
	if state.View == uistate.ViewReview || deps.LoadReviewWindow == nil {
		return handleEvent(state, deps, event)
	}
	stale := false
	deps.LoadReviewWindow = func(string, string) tea.Cmd { stale = true; return nil }
	state, cmd := handleEvent(state, deps, event)
	state.ReviewStale = state.ReviewStale || stale
	return state, cmd
}

func handleEvent(state EventState, deps EventDeps, event api.KernelEvent) (EventState, tea.Cmd) {
	logger.Infof("kernel event: %s", event.Type)
	switch event.Type {
	case "repo.created", "repo.updated", "repo.deleted":
		return state, deps.LoadRepos()
	case "stream.created", "stream.updated", "stream.deleted":
		var all tea.Cmd
		if deps.LoadAllStreams != nil {
			all = deps.LoadAllStreams()
		}
		if state.Context != nil && state.Context.RepoID != nil {
			return state, tea.Batch(deps.LoadStreams(*state.Context.RepoID), all)
		}
		return state, all
	case "issue.created", "issue.updated", "issue.deleted":
		cmds := []tea.Cmd{
			deps.LoadDailySummary(state.CurrentDate),
			deps.LoadReviewWindow(uistate.ReviewRange(state.CurrentDate, state.ReviewWindowEnd, state.ReviewWindowDays)),
		}
		if state.Context != nil && state.Context.StreamID != nil {
			if state.SelectedIssueID != nil {
				cmds = append(
					cmds,
					deps.LoadIssuesSelecting(*state.Context.StreamID, *state.SelectedIssueID),
				)
			} else {
				cmds = append(cmds, deps.LoadIssues(*state.Context.StreamID))
			}
		}
		if state.SelectedIssueID != nil {
			cmds = append(cmds, deps.LoadAllIssuesSelecting(*state.SelectedIssueID))
		} else {
			cmds = append(cmds, deps.LoadAllIssues())
		}
		return state, tea.Batch(append(cmds, summarySnapshotCmd(state, deps)...)...)
	case "habit.created", "habit.updated", "habit.deleted", "habit.completed", "habit.uncompleted":
		cmds := []tea.Cmd{
			deps.LoadDueHabits(state.CurrentDate),
			deps.LoadDailySummary(state.CurrentDate),
			deps.LoadDailyStreaks(state.CurrentDate),
			deps.LoadWellbeing(state.CurrentDate),
			deps.LoadReviewWindow(uistate.ReviewRange(state.CurrentDate, state.ReviewWindowEnd, state.ReviewWindowDays)),
			deps.LoadAllHabits(),
		}
		if deps.LoadTodayHabits != nil {
			cmds = append(cmds, deps.LoadTodayHabits(state.Today))
		}
		if state.Context != nil && state.Context.StreamID != nil {
			cmds = append(cmds, deps.LoadHabits(*state.Context.StreamID))
		}
		if state.View == uistate.ViewHabitHistory {
			cmds = append(cmds, deps.LoadHabitHistory(state.Context, state.SelectedHabitHistoryID))
		}
		return state, tea.Batch(append(cmds, summarySnapshotCmd(state, deps)...)...)
	case "checkin.updated", "checkin.deleted":
		return state, tea.Batch(append([]tea.Cmd{
			deps.LoadDailyStreaks(state.CurrentDate),
			deps.LoadWellbeing(state.CurrentDate),
			deps.LoadReviewWindow(uistate.ReviewRange(state.CurrentDate, state.ReviewWindowEnd, state.ReviewWindowDays)),
		}, summarySnapshotCmd(state, deps)...)...)
	case sharedtypes.EventTypeSessionHistoryChanged:
		cmds := []tea.Cmd{
			deps.LoadAllIssues(), deps.LoadDailySummary(state.CurrentDate),
			deps.LoadSessionHistoryFor200(state), deps.LoadDailyStreaks(state.CurrentDate),
			deps.LoadReviewWindow(uistate.ReviewRange(state.CurrentDate, state.ReviewWindowEnd, state.ReviewWindowDays)),
			deps.LoadWellbeing(state.CurrentDate),
			deps.LoadMomentumRange(state.CurrentDate, state.MomentumWindowDays),
		}
		if deps.LoadSummarySnapshot != nil {
			cmds = append(cmds, deps.LoadSummarySnapshot(state.CurrentDate))
		}
		if state.Context != nil {
			if state.Context.StreamID != nil {
				cmds = append(cmds, deps.LoadIssues(*state.Context.StreamID))
			}
			if state.Context.IssueID != nil && deps.LoadIssueSessions != nil {
				cmds = append(cmds, deps.LoadIssueSessions(*state.Context.IssueID))
			}
		}
		return state, tea.Batch(cmds...)
	case "session.started", "session.stopped":
		return state, tea.Batch(
			deps.LoadTimer(),
			deps.LoadContext(),
			deps.LoadSessionHistoryFor200(state),
			deps.LoadDailyStreaks(state.CurrentDate),
			deps.LoadReviewWindow(uistate.ReviewRange(state.CurrentDate, state.ReviewWindowEnd, state.ReviewWindowDays)),
		)
	case "context.repo.changed",
		"context.stream.changed",
		"context.issue.changed",
		"context.cleared":
		var payload sharedtypes.ContextChangedPayload
		_ = json.Unmarshal(event.Payload, &payload)
		cmds := []tea.Cmd{
			deps.LoadContext(),
			deps.LoadReviewWindow(uistate.ReviewRange(state.CurrentDate, state.ReviewWindowEnd, state.ReviewWindowDays)),
		}
		if payload.RepoID != nil {
			cmds = append(cmds, deps.LoadStreams(*payload.RepoID))
		} else {
			state.Streams = nil
			state.Issues = nil
			state.Cursor[uistate.PaneStreams] = 0
			state.Cursor[uistate.PaneIssues] = 0
		}
		if payload.StreamID != nil {
			cmds = append(
				cmds,
				deps.LoadIssues(*payload.StreamID),
				deps.LoadHabits(*payload.StreamID),
			)
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
				if state.View != uistate.ViewSessionHistory {
					state.View = uistate.ViewSessionActive
				}
				state.Pane = uistate.DefaultPane(state.View)
				if timer.State == "ready" {
					if timer.HardLimitActive && !timer.HardLimitExpired {
						return state, tea.Batch(
							deps.TickAfter(state.TimerTickSeq),
							deps.LoadSessionHistoryFor200(state),
						)
					}
					return state, deps.LoadSessionHistoryFor200(state)
				}
				if timer.State == "expired" {
					return state, deps.LoadSessionHistoryFor200(state)
				}
				return state, tea.Batch(
					deps.TickAfter(state.TimerTickSeq),
					deps.LoadSessionHistoryFor200(state),
				)
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
		return state, tea.Batch(
			deps.LoadTimer(),
			deps.LoadReviewWindow(uistate.ReviewRange(state.CurrentDate, state.ReviewWindowEnd, state.ReviewWindowDays)),
		)
	case "timer.hard_limit_reached":
		state.Elapsed = 0
		var payload sharedtypes.TimerHardLimitReachedPayload
		if err := json.Unmarshal(event.Payload, &payload); err == nil {
			if state.Timer == nil {
				state.Timer = &api.TimerState{}
			}
			state.Timer.State = "expired"
			state.Timer.HardLimitActive = true
			state.Timer.HardLimitKind = sharedtypes.NormalizeTimerHardLimitKind(
				payload.HardLimitKind,
			)
			state.Timer.HardLimitExpired = true
			state.Timer.HardLimitTotalSeconds = payload.HardLimitTotalSeconds
			state.Timer.HardLimitRemainingSeconds = 0
			state.Timer.HardLimitWorkSeconds = payload.HardLimitWorkSeconds
			state.Timer.HardLimitBreakSeconds = payload.HardLimitBreakSeconds
			state.Timer.HardLimitLongBreakSeconds = payload.HardLimitLongBreakSeconds
			state.Timer.HardLimitCyclesBeforeLongBreak = payload.HardLimitCyclesBeforeLongBreak
			state.Timer.SegmentType = nil
			state.Timer.SegmentStartTime = nil
			state.Timer.SegmentElapsedOffsetSeconds = nil
			state.Timer.ReadySegmentType = payload.SegmentType
			state.Timer.NextSegmentType = payload.SegmentType
			state.Timer.ElapsedSeconds = 0
		}
		return state, tea.Batch(
			deps.LoadTimer(),
			deps.LoadReviewWindow(uistate.ReviewRange(state.CurrentDate, state.ReviewWindowEnd, state.ReviewWindowDays)),
		)
	case "timer.advance", "timer.defer_break", "timer.extended", "timer.break_deferred":
		if shouldDismissDialogForSessionEvent(state, event, false) {
			state.Dialog = ""
			state.DialogParent = ""
		}
		return state, tea.Batch(
			deps.LoadTimer(),
			deps.LoadSessionHistoryFor200(state),
			deps.LoadReviewWindow(uistate.ReviewRange(state.CurrentDate, state.ReviewWindowEnd, state.ReviewWindowDays)),
		)
	case "session.ended":
		if shouldDismissDialogForSessionEvent(state, event, true) {
			state.Dialog = ""
			state.DialogParent = ""
		}
		return state, tea.Batch(
			deps.LoadTimer(),
			deps.LoadContext(),
			deps.LoadSessionHistoryFor200(state),
			deps.LoadDailyStreaks(state.CurrentDate),
			deps.LoadReviewWindow(uistate.ReviewRange(state.CurrentDate, state.ReviewWindowEnd, state.ReviewWindowDays)),
		)
	case "update.status":
		return state, deps.LoadUpdateStatus()
	case sharedtypes.EventTypeSettingsChanged:
		cmds := []tea.Cmd{deps.LoadSettings()}
		var payload sharedtypes.SettingsChangedPayload
		if err := json.Unmarshal(event.Payload, &payload); err != nil || len(payload.Keys) == 0 {
			return state, tea.Batch(cmds...)
		}
		if settingsKeysAffectStreaks(payload.Keys) {
			cmds = append(
				cmds,
				deps.LoadDailyStreaks(state.CurrentDate),
				deps.LoadWellbeing(state.CurrentDate),
				deps.LoadReviewWindow(uistate.ReviewRange(state.CurrentDate, state.ReviewWindowEnd, state.ReviewWindowDays)),
				deps.LoadMomentumRange(state.CurrentDate, state.MomentumWindowDays),
			)
		}
		return state, tea.Batch(cmds...)
	case "ops.created":
		return state, deps.LoadOps(state.CurrentOpsLim)
	case sharedtypes.EventTypeDayStart:
		var payload sharedtypes.DayBoundaryEventPayload
		if err := json.Unmarshal(event.Payload, &payload); err != nil || payload.LogicalDate == "" {
			return state, nil
		}
		state.Today = payload.LogicalDate
		// The shared date follows the new day only when it was on the old one.
		if state.Date != "" && state.Date == payload.DateBefore {
			state.Date = payload.LogicalDate
		}
		date := state.Date
		if date == "" {
			date = payload.LogicalDate
		}
		state.CurrentDate = date
		cmds := []tea.Cmd{
			deps.LoadAllIssues(),
			deps.LoadAllHabits(),
			deps.LoadDailySummary(date),
			deps.LoadDailyPlan(date),
			deps.LoadDueHabits(date),
			deps.LoadDailyStreaks(date),
			deps.LoadWellbeing(date),
			deps.LoadMomentumRange(date, state.MomentumWindowDays),
			deps.LoadReviewWindow(uistate.ReviewRange(date, state.ReviewWindowEnd, state.ReviewWindowDays)),
		}
		if deps.LoadTodayHabits != nil {
			cmds = append(cmds, deps.LoadTodayHabits(payload.LogicalDate))
		}
		if deps.LoadSummarySnapshot != nil {
			cmds = append(cmds, deps.LoadSummarySnapshot(date))
		}
		if state.Context != nil && state.Context.StreamID != nil {
			cmds = append(cmds,
				deps.LoadIssues(*state.Context.StreamID),
				deps.LoadHabits(*state.Context.StreamID),
			)
		}
		return state, tea.Batch(cmds...)
	default:
		return state, nil
	}
}

func settingsKeysAffectStreaks(keys []sharedtypes.CoreSettingsKey) bool {
	for _, key := range []sharedtypes.CoreSettingsKey{
		sharedtypes.CoreSettingsKeyAwayModeEnabled,
		sharedtypes.CoreSettingsKeyAwayDates,
		sharedtypes.CoreSettingsKeyFrozenStreakKinds,
		sharedtypes.CoreSettingsKeyRestWeekdays,
		sharedtypes.CoreSettingsKeyRestSpecificDates,
	} {
		if slices.Contains(keys, key) {
			return true
		}
	}
	return false
}

func shouldDismissDialogForSessionEvent(
	state EventState,
	event api.KernelEvent,
	includeEndSession bool,
) bool {
	if state.Timer == nil || state.Timer.SessionID == nil || *state.Timer.SessionID == "" {
		return false
	}
	var payload sharedtypes.SessionEventPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil || payload.SessionID == "" {
		return false
	}
	if payload.SessionID != *state.Timer.SessionID {
		return false
	}
	switch state.Dialog {
	case "hard_limit_expired", "hard_limit_extend", "timer_countdown_extend":
		return true
	case "end_session":
		return includeEndSession || state.DialogParent == "hard_limit_expired"
	default:
		return false
	}
}

// summarySnapshotCmd reloads the shown date's summary snapshot, which Daily's Score/Plan
// and Day summary read, after an edit that changes it.
func summarySnapshotCmd(state EventState, deps EventDeps) []tea.Cmd {
	if deps.LoadSummarySnapshot == nil {
		return nil
	}
	return []tea.Cmd{deps.LoadSummarySnapshot(state.CurrentDate)}
}
