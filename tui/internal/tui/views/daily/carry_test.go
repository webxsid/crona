package daily

import (
	"strings"
	"testing"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"

	"github.com/charmbracelet/x/ansi"
)

func carryPlan(date string, entry api.DailyPlanEntry) *api.DailyPlan {
	return &api.DailyPlan{Date: date, Entries: []api.DailyPlanEntry{entry}}
}

func TestCarriedIssueShowsWhereItWent(t *testing.T) {
	state := detailState()
	state.DailyPlan = carryPlan(state.DashboardDate, api.DailyPlanEntry{
		IssueID:       2,
		Status:        sharedtypes.DailyPlanEntryStatusCarried,
		CarriedToDate: new("2026-10-25"),
		WorkedSeconds: 4800,
	})
	list := ansi.Strip(renderIssues(testTheme(), state, 130, 20))
	assertContainsLine(t, strings.Split(list, "\n"), "Fix sync conflict", "↪ 2026-10-25 · 1h20m")
	detail := strings.Join(strings.Fields(ansi.Strip(renderDetail(testTheme(), state, 100, 40))), " ")
	if !strings.Contains(detail, "Carried to 2026-10-25 after 1h20m of work") {
		t.Fatalf("expected carried-to detail, got %q", detail)
	}
}

func TestCarriedInIssueShowsOriginAndStall(t *testing.T) {
	state := detailState()
	state.DailyPlan = carryPlan(state.DashboardDate, api.DailyPlanEntry{
		IssueID:         2,
		Status:          sharedtypes.DailyPlanEntryStatusPlanned,
		CarriedFromDate: new("2026-10-23"),
		CarryCount:      3,
		Stalled:         true,
	})
	list := ansi.Strip(renderIssues(testTheme(), state, 130, 20))
	assertContainsLine(t, strings.Split(list, "\n"), "Fix sync conflict", "↩ ×3 stalled")
	detail := strings.Join(strings.Fields(ansi.Strip(renderDetail(testTheme(), state, 100, 40))), " ")
	for _, want := range []string{"Carried from 2026-10-23 · 3 carries in a row", "stalled · consider splitting"} {
		if !strings.Contains(detail, want) {
			t.Fatalf("expected %q in detail, got %q", want, detail)
		}
	}
	state.DailyPlan.Entries[0].CarryCount = 1
	state.DailyPlan.Entries[0].Stalled = false
	list = ansi.Strip(renderIssues(testTheme(), state, 130, 20))
	assertContainsLine(t, strings.Split(list, "\n"), "Fix sync conflict", "↩ from 2026-10-23")
}

func TestRescheduleCarriesOnlyWorkedIssuesToday(t *testing.T) {
	settings := &api.CoreSettings{CarryOverEnabled: true}
	plan := carryPlan("2026-10-24", api.DailyPlanEntry{IssueID: 2, Status: sharedtypes.DailyPlanEntryStatusPlanned, WorkedSeconds: 60})
	if !RescheduleCarries(plan, 2, "2026-10-24", settings) {
		t.Fatal("expected a worked issue today to carry")
	}
	if RescheduleCarries(plan, 2, "2026-10-25", settings) {
		t.Fatal("expected no carry when the plan is not today's")
	}
	plan.Entries[0].WorkedSeconds = 0
	if RescheduleCarries(plan, 2, "2026-10-24", settings) {
		t.Fatal("expected an untouched issue to reschedule normally")
	}
	plan.Entries[0].WorkedSeconds = 60
	if RescheduleCarries(plan, 2, "2026-10-24", &api.CoreSettings{}) {
		t.Fatal("expected no carry when carry-over is off")
	}
}
