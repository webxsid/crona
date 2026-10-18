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
		LoadAllIssues:       noop,
		LoadAllHabits:       noop,
		LoadDailySummary:    func(date string) tea.Cmd { *loaded = date; return noop() },
		LoadDailyPlan:       func(string) tea.Cmd { return noop() },
		LoadDueHabits:       func(string) tea.Cmd { return noop() },
		LoadDailyStreaks:    func(string) tea.Cmd { return noop() },
		LoadWellbeing:       func(string, int) tea.Cmd { return noop() },
		LoadMomentumRange:   func(string, int) tea.Cmd { return noop() },
		LoadRollupSummaries: func(string, string) tea.Cmd { return noop() },
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
