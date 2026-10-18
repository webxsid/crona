package summary

import (
	"strings"
	"testing"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/x/ansi"
)

func TestDayLinesIncludeStatsMetrics(t *testing.T) {
	state := summaryTestState()
	state.SummarySnapshot.Focus = &api.FocusScoreSummary{Score: 82, Level: sharedtypes.FocusScoreLevelStrong, Reason: sharedtypes.FocusScoreReasonBalanced, TargetWorkedSeconds: 3600}
	state.SummarySnapshot.Metrics = &api.DailyMetricsDay{WorkedSeconds: 2400, RestSeconds: 600, SessionCount: 3, TotalIssues: 4, CompletedIssues: 2, AbandonedIssues: 1, HabitDueCount: 3, HabitCompletedCount: 2, HabitFailedCount: 1}
	rendered := ansi.Strip(strings.Join(DayLines(types.Theme{}, state, 100), "\n"))
	for _, want := range []string{"Stats", "Focus 82/100", "worked 40m", "rest 10m", "issues  completed 2", "habits  done 2"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("expected summary stats to contain %q, got %q", want, rendered)
		}
	}
}

func TestDayLinesCoverAccountabilityWellbeingAndStreaks(t *testing.T) {
	rendered := ansi.Strip(strings.Join(DayLines(types.Theme{}, summaryTestState(), 100), "\n"))
	for _, want := range []string{"Accountability", "100% score", "Wellbeing", "mood 4/5", "Momentum", "focus 3"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("expected the Day summary to contain %q, got %q", want, rendered)
		}
	}
}

func TestDayLinesWaitForTheShownDate(t *testing.T) {
	state := summaryTestState()
	state.SummaryDate = "2026-07-29"
	if got := ansi.Strip(strings.Join(DayLines(types.Theme{}, state, 100), "\n")); got != "Loading summary…" {
		t.Fatalf("expected a loading line for another day's snapshot, got %q", got)
	}
}

func summaryTestState() types.ContentState {
	estimate := 30
	target := 20
	return types.ContentState{
		View:        "summary",
		Pane:        "summary",
		Width:       120,
		Height:      42,
		SummaryDate: "2026-07-28",
		Timer:       &api.TimerState{State: "running"},
		Elapsed:     900,
		SummarySnapshot: &api.SummarySnapshot{
			Date: "2026-07-28",
			Issues: &api.DailyIssueSummary{
				Date:                  "2026-07-28",
				TotalIssues:           1,
				CompletedIssues:       1,
				TotalEstimatedMinutes: 30,
				WorkedSeconds:         1200,
				Issues: []api.Issue{{
					Title:           "Ship summary",
					Status:          sharedtypes.IssueStatusDone,
					EstimateMinutes: &estimate,
					WorkedSeconds:   1200,
				}},
			},
			Habits: []api.HabitDailyItem{{
				HabitWithMeta: api.HabitWithMeta{Habit: api.Habit{Name: "Walk", TargetMinutes: &target}},
				Completed:     true,
				Status:        sharedtypes.HabitCompletionStatusCompleted,
			}},
			Plan: &api.DailyPlan{
				Date: "2026-07-28",
				Summary: api.DailyPlanAccountabilitySummary{
					PlannedCount:        1,
					CompletedCount:      1,
					AccountabilityScore: 100,
				},
			},
			CheckIn: &api.DailyCheckIn{Date: "2026-07-28", Mood: 4, Energy: 3},
			Streaks: &api.StreakSummary{CurrentFocusDays: 3, CurrentCheckInDays: 4, CurrentHabitDays: 2},
		},
	}
}
