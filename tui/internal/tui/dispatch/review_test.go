package dispatch

import (
	"testing"

	"crona/tui/internal/api"
	"crona/tui/internal/tui/commands"
)

func reviewMessageState() MessageState {
	return MessageState{Date: "2026-10-14", ReviewWindowEnd: "2026-10-14", ReviewWindowDays: 7}
}

func TestStaleReviewMessagesIgnored(t *testing.T) {
	deps := MessageDeps{CurrentDate: func(s MessageState) string { return s.Date }}
	stale := "2026-09-15"
	state, _, _ := HandleMessage(
		reviewMessageState(),
		commands.DashboardWindowLoadedMsg{
			Summary: &api.DashboardWindowSummary{StartDate: stale, EndDate: "2026-10-14"},
		},
		deps,
	)
	state, _, _ = HandleMessage(
		state,
		commands.MetricsRollupLoadedMsg{
			Rollup: &api.MetricsRollup{StartDate: stale, EndDate: "2026-10-14"},
		},
		deps,
	)
	state, _, _ = HandleMessage(
		state,
		commands.MetricsRangeLoadedMsg{
			Days: []api.DailyMetricsDay{{Date: stale}, {Date: "2026-10-14"}},
		},
		deps,
	)
	state, _, _ = HandleMessage(
		state,
		commands.RollupMetricsRangeLoadedMsg{
			Days: []api.DailyMetricsDay{{Date: stale}, {Date: "2026-10-14"}},
		},
		deps,
	)
	if state.DashboardWindow != nil || state.MetricsRollup != nil || state.MetricsRange != nil ||
		state.RollupMetricsRange != nil {
		t.Fatalf("expected results for an old window ignored")
	}
}

func TestCurrentReviewMessagesApplied(t *testing.T) {
	deps := MessageDeps{CurrentDate: func(s MessageState) string { return s.Date }}
	start := "2026-10-08"
	state, _, _ := HandleMessage(
		reviewMessageState(),
		commands.DashboardWindowLoadedMsg{
			Summary: &api.DashboardWindowSummary{StartDate: start, EndDate: "2026-10-14"},
		},
		deps,
	)
	state, _, _ = HandleMessage(
		state,
		commands.MetricsRollupLoadedMsg{
			Rollup: &api.MetricsRollup{StartDate: start, EndDate: "2026-10-14"},
		},
		deps,
	)
	state, _, _ = HandleMessage(
		state,
		commands.MetricsRangeLoadedMsg{
			Days: []api.DailyMetricsDay{{Date: start}, {Date: "2026-10-14"}},
		},
		deps,
	)
	if state.DashboardWindow == nil || state.MetricsRollup == nil || len(state.MetricsRange) != 2 {
		t.Fatalf("expected results for the current window applied")
	}
}
