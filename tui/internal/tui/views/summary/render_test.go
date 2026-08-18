package summary

import (
	"strings"
	"testing"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/x/ansi"
)

func TestRenderOwnsResponsiveSummaryLayout(t *testing.T) {
	state := summaryTestState()
	for _, width := range []int{64, 120} {
		state.Width = width
		rendered := ansi.Strip(Render(types.Theme{}, state))
		for _, want := range []string{"Summary", "At a glance", "Agenda", "Issues", "Habits", "Signals", "Accountability", "Wellbeing", "Momentum"} {
			if !strings.Contains(rendered, want) {
				t.Fatalf("width %d: expected %q in summary:\n%s", width, want, rendered)
			}
		}
		if strings.Contains(rendered, "Daily Dashboard") {
			t.Fatalf("width %d: summary must not reuse dashboard presentation", width)
		}
	}
}

func TestRenderHistoricalSummaryOmitsLiveTimer(t *testing.T) {
	state := summaryTestState()
	state.SummaryDate = "2020-01-02"
	state.SummarySnapshot.Date = state.SummaryDate
	rendered := ansi.Strip(Render(types.Theme{}, state))
	if !strings.Contains(rendered, "Historical") {
		t.Fatalf("expected historical label, got:\n%s", rendered)
	}
	if strings.Contains(rendered, "RUNNING") {
		t.Fatalf("historical summary must not show current timer, got:\n%s", rendered)
	}
}

func TestRenderSummaryIncludesStatsMetrics(t *testing.T) {
	state := summaryTestState()
	state.SummarySnapshot.Focus = &api.FocusScoreSummary{Score: 82, Level: sharedtypes.FocusScoreLevelStrong, Reason: sharedtypes.FocusScoreReasonBalanced, TargetWorkedSeconds: 3600}
	state.SummarySnapshot.Metrics = &api.DailyMetricsDay{WorkedSeconds: 2400, RestSeconds: 600, SessionCount: 3, TotalIssues: 4, CompletedIssues: 2, AbandonedIssues: 1, HabitDueCount: 3, HabitCompletedCount: 2, HabitFailedCount: 1}
	rendered := ansi.Strip(Render(types.Theme{}, state))
	for _, want := range []string{"Stats", "Focus 82/100", "worked 40m", "rest 10m", "issues  completed 2", "habits  done 2"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("expected summary stats to contain %q, got %q", want, rendered)
		}
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
