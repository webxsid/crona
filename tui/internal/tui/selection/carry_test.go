package selection

import (
	"testing"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	uistate "crona/tui/internal/tui/state"
)

func TestDailyPlannedKeepsIssuesCarriedAwayFromTheDay(t *testing.T) {
	day, next := "2026-10-24", "2026-10-25"
	s := Snapshot{
		View:             uistate.ViewDaily,
		DailyTaskSection: uistate.DailyTaskSectionPlanned,
		DashboardDate:    day,
		AllIssues: []api.IssueWithMeta{
			{Issue: api.Issue{ID: 1, Status: "ready", TodoForDate: &day}},
			{Issue: api.Issue{ID: 2, Status: "in_progress", TodoForDate: &next}},
			{Issue: api.Issue{ID: 3, Status: "ready", TodoForDate: &next}},
		},
		DailyPlan: &api.DailyPlan{Date: day, Entries: []api.DailyPlanEntry{
			{IssueID: 1, Status: sharedtypes.DailyPlanEntryStatusPlanned},
			{IssueID: 2, Status: sharedtypes.DailyPlanEntryStatusCarried},
		}},
	}
	got := DailySectionIssues(s, uistate.DailyTaskSectionPlanned)
	if len(got) != 2 || got[0].ID != 1 || got[1].ID != 2 {
		t.Fatalf("expected the day's issue and the one carried from it, got %+v", got)
	}
	if overdue := DailySectionIssues(s, uistate.DailyTaskSectionOverdue); len(overdue) != 0 {
		t.Fatalf("expected no overdue issues, got %+v", overdue)
	}
}
