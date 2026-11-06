package daily

import (
	"strings"
	"testing"

	"crona.work/tui/internal/api"

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
		penalty.IfDoneToday = &api.DailyPlanPenaltyShare{
			Accountability:  -3.38,
			BacklogPressure: -2.58,
		}
		penalty.IfPostponedOneDay = &api.DailyPlanPenaltyProjection{
			FailScore:           3.81,
			AccountabilityDelta: 0.43,
		}
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
	rendered := strings.Join(
		strings.Fields(ansi.Strip(renderDetail(testTheme(), focusDetail(state), 100, 40))),
		" ",
	)
	for _, want := range []string{
		"▲ high risk · 54% of today",
		"3.38",
		"Penalty/2.5",
		"Delay 5 days late · pushed 3× · 1 away day skipped",
		"If today ✓ clears 3.38",
		"If pushed ▲ rises to 3.81 (+0.43)",
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
	rendered := strings.Join(
		strings.Fields(ansi.Strip(renderDetail(testTheme(), focusDetail(state), 100, 40))),
		" ",
	)
	if !strings.Contains(rendered, "Delay penalty (final)") ||
		strings.Contains(rendered, "high risk") ||
		strings.Contains(rendered, "Push a day") ||
		strings.Contains(rendered, "Finish today") {
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
	rendered := ansi.Strip(renderDetail(testTheme(), focusDetail(state), 100, 40))
	if !strings.Contains(rendered, "on schedule · no penalty") ||
		strings.Contains(rendered, "Delay penalty") {
		t.Fatalf("expected on-schedule line, got %q", rendered)
	}
}

func TestRescheduleHintShowsPushCost(t *testing.T) {
	if got := RescheduleHint(penaltyPlan(false), 2); got != "Push a day: penalty rises to 3.81 (+0.43)" {
		t.Fatalf("unexpected hint %q", got)
	}
	if got := RescheduleHint(penaltyPlan(true), 2); got != "" {
		t.Fatalf("expected no hint for a resolved entry, got %q", got)
	}
	if got := RescheduleHint(nil, 2); got != "" {
		t.Fatalf("expected no hint without a plan, got %q", got)
	}
}

func TestIssueInsightLinesMatchTheDetailPane(t *testing.T) {
	state := detailState()
	state.DailyPlan = penaltyPlan(false)
	state.MomentumCards = []api.MomentumCard{{
		Definition: api.HabitStreakDefinition{
			ID: "fit", Name: "Fitness", Enabled: true, Period: "week", RequiredCount: 3, HabitIDs: []int64{2},
		},
		Current: 4, Longest: 9,
	}}
	issue := strings.Join(strings.Fields(ansi.Strip(strings.Join(IssueInsightLines(testTheme(), state, 2, 0, 0, 80), "\n"))), " ")
	if !strings.Contains(issue, "Delay penalty ▲ high risk") {
		t.Fatalf("expected the delay penalty for the dialog, got %q", issue)
	}
	habit := ansi.Strip(strings.Join(HabitMomentumLines(testTheme(), state, 2, 80), "\n"))
	if !strings.Contains(habit, "Momentum") || !strings.Contains(habit, "Fitness") || !strings.ContainsAny(habit, "▰▱") {
		t.Fatalf("expected the habit's momentum ladder for the dialog, got %q", habit)
	}
}
