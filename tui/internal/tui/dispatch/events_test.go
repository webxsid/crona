package dispatch

import (
	"encoding/json"
	"testing"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	uistate "crona/tui/internal/tui/state"

	tea "github.com/charmbracelet/bubbletea"
)

func TestTimerExtendedDismissesMatchingHardLimitDialog(t *testing.T) {
	state, cmd := HandleEvent(
		EventState{
			Dialog:        "hard_limit_expired",
			Timer:         &api.TimerState{SessionID: stringPtr("session-1")},
			CurrentDash:   "2026-07-14",
			Cursor:        map[uistate.Pane]int{},
			CurrentOpsLim: 50,
		},
		testEventDeps(),
		sessionEvent(sharedtypes.EventTypeTimerExtended, "session-1"),
	)

	if state.Dialog != "" {
		t.Fatalf("expected matching hard-limit dialog to close, got %q", state.Dialog)
	}
	if cmd == nil {
		t.Fatal("expected timer.extended to schedule refresh commands")
	}
}

func TestTimerActionEventsRefreshTimer(t *testing.T) {
	for _, eventType := range []string{"timer.advance", "timer.extend_current_session"} {
		state, cmd := HandleEvent(
			EventState{CurrentDash: "2026-07-14", Cursor: map[uistate.Pane]int{}},
			testEventDeps(),
			api.KernelEvent{Type: eventType},
		)
		if cmd == nil {
			t.Fatalf("%s: expected timer refresh command", eventType)
		}
		if state.Timer != nil {
			t.Fatalf("%s: expected event to defer timer replacement to LoadTimer", eventType)
		}
	}
}

func TestTimerExtendedDismissesMatchingHardLimitEndSessionDialog(t *testing.T) {
	state, _ := HandleEvent(
		EventState{
			Dialog:       "end_session",
			DialogParent: "hard_limit_expired",
			Timer:        &api.TimerState{SessionID: stringPtr("session-1")},
			CurrentDash:  "2026-07-14",
			Cursor:       map[uistate.Pane]int{},
		},
		testEventDeps(),
		sessionEvent(sharedtypes.EventTypeTimerExtended, "session-1"),
	)

	if state.Dialog != "" {
		t.Fatalf("expected matching hard-limit end-session dialog to close, got %q", state.Dialog)
	}
}

func TestSessionEndedDismissesMatchingEndSessionDialog(t *testing.T) {
	state, cmd := HandleEvent(
		EventState{
			Dialog:      "end_session",
			Timer:       &api.TimerState{SessionID: stringPtr("session-1")},
			CurrentDash: "2026-07-14",
			Cursor:      map[uistate.Pane]int{},
		},
		testEventDeps(),
		sessionEvent(sharedtypes.EventTypeSessionEnded, "session-1"),
	)

	if state.Dialog != "" {
		t.Fatalf("expected matching end-session dialog to close, got %q", state.Dialog)
	}
	if cmd == nil {
		t.Fatal("expected session.ended to schedule refresh commands")
	}
}

func TestSessionEventIgnoresDifferentSession(t *testing.T) {
	state, _ := HandleEvent(
		EventState{
			Dialog:      "hard_limit_expired",
			Timer:       &api.TimerState{SessionID: stringPtr("session-1")},
			CurrentDash: "2026-07-14",
			Cursor:      map[uistate.Pane]int{},
		},
		testEventDeps(),
		sessionEvent(sharedtypes.EventTypeSessionEnded, "session-2"),
	)

	if state.Dialog != "hard_limit_expired" {
		t.Fatalf("expected unrelated session event to leave dialog open, got %q", state.Dialog)
	}
}

func TestHardLimitReachedEventCarriesCountdownKind(t *testing.T) {
	payload, err := json.Marshal(sharedtypes.TimerHardLimitReachedPayload{
		SessionID:             "session-1",
		IssueID:               7,
		HardLimitKind:         sharedtypes.TimerHardLimitKindCountdown,
		HardLimitTotalSeconds: 1500,
		HardLimitWorkSeconds:  1500,
	})
	if err != nil {
		t.Fatalf("marshal hard-limit event: %v", err)
	}
	state, _ := HandleEvent(
		EventState{
			Timer:         &api.TimerState{SessionID: stringPtr("session-1")},
			CurrentDash:   "2026-07-14",
			Cursor:        map[uistate.Pane]int{},
			CurrentOpsLim: 50,
		},
		testEventDeps(),
		api.KernelEvent{Type: sharedtypes.EventTypeTimerHardLimitReached, Payload: payload},
	)

	if state.Timer.HardLimitKind != sharedtypes.TimerHardLimitKindCountdown {
		t.Fatalf("expected countdown kind from event, got %+v", state.Timer)
	}
}

func sessionEvent(eventType string, sessionID string) api.KernelEvent {
	payload, _ := json.Marshal(sharedtypes.SessionEventPayload{SessionID: sessionID})
	return api.KernelEvent{Type: eventType, Payload: payload}
}

func testEventDeps() EventDeps {
	noop := func() tea.Cmd { return func() tea.Msg { return nil } }
	return EventDeps{
		LoadRepos:                noop,
		LoadStreams:              func(int64) tea.Cmd { return noop() },
		LoadIssues:               func(int64) tea.Cmd { return noop() },
		LoadIssuesSelecting:      func(int64, int64) tea.Cmd { return noop() },
		LoadHabits:               func(int64) tea.Cmd { return noop() },
		LoadAllHabits:            noop,
		LoadAllIssues:            noop,
		LoadAllIssuesSelecting:   func(int64) tea.Cmd { return noop() },
		LoadDailySummary:         func(string) tea.Cmd { return noop() },
		LoadDueHabits:            func(string) tea.Cmd { return noop() },
		LoadDailyStreaks:         func(string) tea.Cmd { return noop() },
		LoadHabitHistory:         func(*api.ActiveContext, *int64) tea.Cmd { return noop() },
		LoadWellbeing:            func(string, int) tea.Cmd { return noop() },
		LoadRollupSummaries:      func(string, string) tea.Cmd { return noop() },
		LoadSessionHistoryFor200: func(EventState) tea.Cmd { return noop() },
		LoadContext:              noop,
		LoadTimer:                noop,
		LoadAlertStatus:          noop,
		LoadUpdateStatus:         noop,
		LoadOps:                  func(int) tea.Cmd { return noop() },
		TickAfter:                func(int) tea.Cmd { return noop() },
	}
}

func stringPtr(value string) *string {
	return &value
}
