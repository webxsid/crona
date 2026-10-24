package daily

import (
	"strings"
	"testing"

	"crona/tui/internal/api"

	"github.com/charmbracelet/x/ansi"
)

func penaltyPlan(historical bool) *api.DailyPlan {
	penalty := &api.DailyPlanEntryPenalty{
		ProtectedDaysSkipped: 1,
		DelayComponent:       2.58,
		FrequencyComponent:   0.79,
		HighRisk:             true,
		HighRiskThreshold:    2.5,
		Historical:           historical,
	}
	if !historical {
		penalty.Contribution = &api.DailyPlanPenaltyShare{
			Accountability: 3.38, AccountabilityPct: 0.54, BacklogPressure: 2.58, BacklogPressurePct: 0.38,
		}
		penalty.IfDoneToday = &api.DailyPlanPenaltyShare{Accountability: -3.38, BacklogPressure: -2.58}
		penalty.IfPostponedOneDay = &api.DailyPlanPenaltyProjection{FailScore: 3.81, AccountabilityDelta: 0.43}
	}
	plan := &api.DailyPlan{Entries: []api.DailyPlanEntry{{
		IssueID:            2,
		BaselineDate:       "2026-10-18",
		CurrentPlannedDate: "2026-10-24",
		PostponeCount:      3,
		CurrentDelayedDays: 5,
		FailScore:          3.38,
		Penalty:            penalty,
	}}}
	plan.Summary.AccountabilityScore = 6.3
	plan.Summary.BacklogPressure = 6.8
	return plan
}

func TestDetailShowsLiveDelayPenalty(t *testing.T) {
	state := detailState()
	state.DailyPlan = penaltyPlan(false)
	rendered := strings.Join(strings.Fields(ansi.Strip(renderDetail(testTheme(), state, 100, 40))), " ")
	for _, want := range []string{
		"Delay penalty ▲ high risk",
		"5 days late, 3 postpones, 1 away day skipped",
		"3.38 = 2.58 delay + 0.79 frequency threshold 2.50",
		"+3.38 accountability (54% of 6.30)",
		"+2.58 backlog pressure (38% of 6.80)",
		"accountability -3.38 · backlog -2.58",
		"penalty 3.81 (+0.43)",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("expected detail to contain %q, got %q", want, rendered)
		}
	}
	list := ansi.Strip(renderIssues(testTheme(), state, 130, 20))
	assertContainsLine(t, strings.Split(list, "\n"), "Fix sync conflict", "▲3.4")
}

func TestDetailShowsFinalPenaltyForResolvedIssue(t *testing.T) {
	state := detailState()
	state.DailyPlan = penaltyPlan(true)
	rendered := strings.Join(strings.Fields(ansi.Strip(renderDetail(testTheme(), state, 100, 40))), " ")
	if !strings.Contains(rendered, "Delay penalty (final)") || strings.Contains(rendered, "high risk") ||
		strings.Contains(rendered, "Push 1 day") || strings.Contains(rendered, "Today") {
		t.Fatalf("expected only final penalty lines, got %q", rendered)
	}
	list := ansi.Strip(renderIssues(testTheme(), state, 130, 20))
	if strings.Contains(list, "▲") {
		t.Fatalf("expected no row marker for a resolved entry, got %q", list)
	}
}

func TestDetailShowsOnScheduleWithoutPenalty(t *testing.T) {
	state := detailState()
	state.DailyPlan = &api.DailyPlan{Entries: []api.DailyPlanEntry{{
		IssueID: 2, PostponeCount: 1, Penalty: &api.DailyPlanEntryPenalty{HighRiskThreshold: 2.5},
	}}}
	rendered := ansi.Strip(renderDetail(testTheme(), state, 100, 40))
	if !strings.Contains(rendered, "On schedule · no penalty") || strings.Contains(rendered, "Delay penalty") {
		t.Fatalf("expected on-schedule line, got %q", rendered)
	}
}
