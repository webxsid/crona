package dispatch

import (
	"encoding/json"
	"testing"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	"crona/tui/internal/tui/commands"

	tea "github.com/charmbracelet/bubbletea"
)

func rolloverMessageDeps(loaded *string) MessageDeps {
	noop := func() tea.Cmd { return func() tea.Msg { return nil } }
	return MessageDeps{
		LoadAllIssues:     noop,
		LoadAllHabits:     noop,
		LoadDailySummary:  func(date string) tea.Cmd { *loaded = date; return noop() },
		LoadDailyPlan:     func(string) tea.Cmd { return noop() },
		LoadDueHabits:     func(string) tea.Cmd { return noop() },
		LoadDailyStreaks:  func(string) tea.Cmd { return noop() },
		LoadWellbeing:     func(string) tea.Cmd { return noop() },
		LoadMomentumRange: func(string, int) tea.Cmd { return noop() },
		LoadReviewWindow:  func(string, string) tea.Cmd { return noop() },
	}
}

func TestLogicalDateRollover(t *testing.T) {
	cases := []struct {
		name, date, want string
	}{
		{"on the old today it follows", "2026-08-09", "2026-08-10"},
		{"on another day it stays", "2026-08-01", "2026-08-01"},
		{"unset it takes the new day", "", "2026-08-10"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			loaded := ""
			state, cmd, _ := HandleMessage(MessageState{
				Date:   tc.date,
				Health: &api.Health{CurrentDate: "2026-08-09"},
			}, commands.HealthLoadedMsg{Health: &api.Health{CurrentDate: "2026-08-10"}}, rolloverMessageDeps(&loaded))
			if state.Date != tc.want {
				t.Fatalf("Date = %q, want %q", state.Date, tc.want)
			}
			if cmd == nil || loaded != tc.want {
				t.Fatalf("expected surfaces reloaded on %q, got %q", tc.want, loaded)
			}
		})
	}
}

func TestDayStartRollover(t *testing.T) {
	cases := []struct {
		name, date, wantDate, wantLoaded string
	}{
		{"on the old today it follows", "2026-08-09", "2026-08-10", "2026-08-10"},
		{"on another day it stays", "2026-08-01", "2026-08-01", "2026-08-01"},
		{"unset it keeps following today", "", "", "2026-08-10"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			loaded := ""
			deps := testEventDeps()
			deps.LoadDailySummary = func(date string) tea.Cmd { loaded = date; return func() tea.Msg { return nil } }
			payload, _ := json.Marshal(sharedtypes.DayBoundaryEventPayload{LogicalDate: "2026-08-10", DateBefore: "2026-08-09"})
			state, _ := HandleEvent(EventState{Date: tc.date}, deps, api.KernelEvent{Type: sharedtypes.EventTypeDayStart, Payload: payload})
			if state.Date != tc.wantDate || loaded != tc.wantLoaded {
				t.Fatalf("Date = %q loaded %q, want %q loaded %q", state.Date, loaded, tc.wantDate, tc.wantLoaded)
			}
		})
	}
}

func TestReviewRefreshUsesSharedDateAndWindow(t *testing.T) {
	var start, end string
	deps := MessageDeps{
		CurrentDate:      func(s MessageState) string { return s.Date },
		LoadReviewWindow: func(s, e string) tea.Cmd { start, end = s, e; return func() tea.Msg { return nil } },
	}
	_, cmd, handled := HandleMessage(MessageState{Date: "2026-08-10", ReviewWindowDays: 30}, commands.RequestRollupRefresh()(), deps)
	if !handled || cmd == nil || start != "2026-07-12" || end != "2026-08-10" {
		t.Fatalf("expected Review reloaded on 2026-07-12..2026-08-10, got %q..%q", start, end)
	}
}

func TestRolloverReloadsSummarySnapshot(t *testing.T) {
	snapshot := ""
	loaded := ""
	deps := rolloverMessageDeps(&loaded)
	deps.LoadSummarySnapshot = func(date string) tea.Cmd { snapshot = date; return func() tea.Msg { return nil } }
	HandleMessage(MessageState{Date: "2026-08-09", Health: &api.Health{CurrentDate: "2026-08-09"}},
		commands.HealthLoadedMsg{Health: &api.Health{CurrentDate: "2026-08-10"}}, deps)
	if snapshot != "2026-08-10" {
		t.Fatalf("expected the summary snapshot reloaded on rollover, got %q", snapshot)
	}
	snapshot = ""
	eventDeps := testEventDeps()
	eventDeps.LoadSummarySnapshot = func(date string) tea.Cmd { snapshot = date; return func() tea.Msg { return nil } }
	payload, _ := json.Marshal(sharedtypes.DayBoundaryEventPayload{LogicalDate: "2026-08-10", DateBefore: "2026-08-09"})
	HandleEvent(EventState{Date: "2026-08-09"}, eventDeps, api.KernelEvent{Type: sharedtypes.EventTypeDayStart, Payload: payload})
	if snapshot != "2026-08-10" {
		t.Fatalf("expected the summary snapshot reloaded on day_start, got %q", snapshot)
	}
}

func TestDailyEditsReloadSummarySnapshot(t *testing.T) {
	for _, eventType := range []string{"issue.updated", "habit.completed", "checkin.updated"} {
		snapshot := ""
		deps := testEventDeps()
		deps.LoadSummarySnapshot = func(date string) tea.Cmd { snapshot = date; return func() tea.Msg { return nil } }
		HandleEvent(EventState{CurrentDate: "2026-08-10"}, deps, api.KernelEvent{Type: eventType})
		if snapshot != "2026-08-10" {
			t.Fatalf("%s: expected the summary snapshot reloaded for the shown date, got %q", eventType, snapshot)
		}
	}
}

func TestRolloverRefreshesTodayWhileDailyStaysHistorical(t *testing.T) {
	historical, newToday := "2026-08-01", "2026-08-10"
	t.Run("health", func(t *testing.T) {
		shown, today := "", ""
		deps := rolloverMessageDeps(&shown)
		deps.LoadTodayHabits = func(date string) tea.Cmd { today = date; return nil }
		state, _, _ := HandleMessage(MessageState{Date: historical, Health: &api.Health{CurrentDate: "2026-08-09"}}, commands.HealthLoadedMsg{Health: &api.Health{CurrentDate: newToday}}, deps)
		if state.Date != historical || shown != historical || today != newToday {
			t.Fatalf("health rollover: stored %q, displayed %q, today load %q; want %q, %q, %q", state.Date, shown, today, historical, historical, newToday)
		}
	})
	t.Run("day start", func(t *testing.T) {
		shown, today := "", ""
		deps := testEventDeps()
		deps.LoadDueHabits = func(date string) tea.Cmd { shown = date; return nil }
		deps.LoadTodayHabits = func(date string) tea.Cmd { today = date; return nil }
		payload, _ := json.Marshal(sharedtypes.DayBoundaryEventPayload{DateBefore: "2026-08-09", LogicalDate: newToday})
		state, _ := HandleEvent(EventState{Date: historical, Today: "2026-08-09"}, deps, api.KernelEvent{Type: sharedtypes.EventTypeDayStart, Payload: payload})
		if state.Date != historical || shown != historical || today != newToday {
			t.Fatalf("day start: stored %q, displayed %q, today load %q; want %q, %q, %q", state.Date, shown, today, historical, historical, newToday)
		}
	})
}
