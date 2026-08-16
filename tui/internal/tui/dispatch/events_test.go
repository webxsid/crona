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
	for _, eventType := range []string{"timer.advance", "timer.defer_break"} {
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

func TestSettingsChangedReloadsSettings(t *testing.T) {
	called := false
	deps := testEventDeps()
	deps.LoadSettings = func() tea.Cmd {
		called = true
		return func() tea.Msg { return nil }
	}
	_, cmd := HandleEvent(EventState{}, deps, api.KernelEvent{Type: sharedtypes.EventTypeSettingsChanged})
	if cmd == nil || !called {
		t.Fatal("expected settings.changed to reload settings")
	}
}

func TestAwayDatesChangedReloadsStreakSurfaces(t *testing.T) {
	called := map[string]bool{}
	deps := testEventDeps()
	deps.LoadSettings = trackedEventCommand(called, "settings")
	deps.LoadDailyStreaks = func(string) tea.Cmd { return trackedEventCommand(called, "daily")() }
	deps.LoadWellbeing = func(string, int) tea.Cmd { return trackedEventCommand(called, "wellbeing")() }
	deps.LoadRollupSummaries = func(string, string) tea.Cmd { return trackedEventCommand(called, "rollup")() }
	deps.LoadMomentumRange = func(string, int) tea.Cmd { return trackedEventCommand(called, "momentum")() }
	payload, err := json.Marshal(sharedtypes.SettingsChangedPayload{
		Keys: []sharedtypes.CoreSettingsKey{sharedtypes.CoreSettingsKeyAwayDates},
	})
	if err != nil {
		t.Fatalf("marshal settings event: %v", err)
	}
	state := EventState{
		CurrentDash:         "2026-08-09",
		CurrentWell:         "2026-08-09",
		CurrentMomentum:     "2026-08-09",
		MomentumWindowDays:  30,
		WellbeingWindowDays: 7,
	}
	_, cmd := HandleEvent(state, deps, api.KernelEvent{
		Type:    sharedtypes.EventTypeSettingsChanged,
		Payload: payload,
	})
	if cmd == nil {
		t.Fatal("expected awayDates change to schedule refresh commands")
	}
	for _, name := range []string{"settings", "daily", "wellbeing", "rollup", "momentum"} {
		if !called[name] {
			t.Fatalf("expected %s reload after awayDates change", name)
		}
	}
}

func TestDayStartReloadsAllDateScopedSurfaces(t *testing.T) {
	called := map[string]bool{}
	deps := testEventDeps()
	deps.LoadAllIssues = func() tea.Cmd { return trackedEventCommand(called, "issues")() }
	deps.LoadAllHabits = func() tea.Cmd { return trackedEventCommand(called, "habits")() }
	deps.LoadDailySummary = func(string) tea.Cmd { return trackedEventCommand(called, "summary")() }
	deps.LoadDailyPlan = func(string) tea.Cmd { return trackedEventCommand(called, "plan")() }
	deps.LoadDueHabits = func(string) tea.Cmd { return trackedEventCommand(called, "due")() }
	deps.LoadDailyStreaks = func(string) tea.Cmd { return trackedEventCommand(called, "streaks")() }
	deps.LoadWellbeing = func(string, int) tea.Cmd { return trackedEventCommand(called, "wellbeing")() }
	deps.LoadMomentumRange = func(string, int) tea.Cmd { return trackedEventCommand(called, "momentum")() }
	deps.LoadRollupSummaries = func(string, string) tea.Cmd { return trackedEventCommand(called, "rollup")() }
	payload, _ := json.Marshal(sharedtypes.DayBoundaryEventPayload{LogicalDate: "2026-08-10"})
	_, cmd := HandleEvent(EventState{WellbeingWindowDays: 7, MomentumWindowDays: 30}, deps, api.KernelEvent{
		Type: sharedtypes.EventTypeDayStart, Payload: payload,
	})
	if cmd == nil {
		t.Fatal("expected day.start to schedule refresh commands")
	}
	for _, name := range []string{"issues", "habits", "summary", "plan", "due", "streaks", "wellbeing", "momentum", "rollup"} {
		if !called[name] {
			t.Fatalf("expected %s reload after day.start", name)
		}
	}
}

func trackedEventCommand(called map[string]bool, name string) func() tea.Cmd {
	return func() tea.Cmd {
		called[name] = true
		return func() tea.Msg { return nil }
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
		LoadDailyPlan:            func(string) tea.Cmd { return noop() },
		LoadDueHabits:            func(string) tea.Cmd { return noop() },
		LoadDailyStreaks:         func(string) tea.Cmd { return noop() },
		LoadHabitHistory:         func(*api.ActiveContext, *int64) tea.Cmd { return noop() },
		LoadWellbeing:            func(string, int) tea.Cmd { return noop() },
		LoadMomentumRange:        func(string, int) tea.Cmd { return noop() },
		LoadRollupSummaries:      func(string, string) tea.Cmd { return noop() },
		LoadSessionHistoryFor200: func(EventState) tea.Cmd { return noop() },
		LoadContext:              noop,
		LoadTimer:                noop,
		LoadAlertStatus:          noop,
		LoadSettings:             noop,
		LoadUpdateStatus:         noop,
		LoadOps:                  func(int) tea.Cmd { return noop() },
		TickAfter:                func(int) tea.Cmd { return noop() },
	}
}

func stringPtr(value string) *string {
	return &value
}
