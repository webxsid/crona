package summary

import (
	"strings"
	"testing"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/x/ansi"
)

func dayText(state types.ContentState, width int) string {
	return ansi.Strip(strings.Join(DayLines(types.Theme{}, state, width), "\n"))
}

func squash(s string) string { return strings.Join(strings.Fields(s), " ") }

func TestDayLinesIncludeStatsMetrics(t *testing.T) {
	state := summaryTestState()
	state.SummarySnapshot.Focus = &api.FocusScoreSummary{Score: 82, Level: sharedtypes.FocusScoreLevelStrong, Reason: sharedtypes.FocusScoreReasonBalanced, TargetWorkedSeconds: 3600}
	state.SummarySnapshot.Metrics = &api.DailyMetricsDay{WorkedSeconds: 2400, RestSeconds: 600, SessionCount: 3, TotalIssues: 4, CompletedIssues: 2, AbandonedIssues: 1, HabitDueCount: 3, HabitCompletedCount: 2, HabitFailedCount: 1}
	rendered := dayText(state, 42)
	for _, want := range []string{"Stats", "Focus 82/100 Strong", "Worked 40m of 1h", "Rest 10m", "Sessions 3", "Issues 2 of 4 done · 1 abandoned", "Habits 2 of 3 done · 1 failed"} {
		if !strings.Contains(squash(rendered), want) {
			t.Fatalf("expected summary stats to contain %q, got %q", want, rendered)
		}
	}
}

func TestDayLinesOneStatPerRowWithinWidth(t *testing.T) {
	state := summaryTestState()
	state.SummarySnapshot.Focus = &api.FocusScoreSummary{Score: 52, Level: sharedtypes.FocusScoreLevelOverextended, Reason: "overextended", TargetWorkedSeconds: 21600}
	state.SummarySnapshot.Metrics = &api.DailyMetricsDay{WorkedSeconds: 6660, SessionCount: 2, TotalIssues: 4, CompletedIssues: 1, HabitDueCount: 5, HabitCompletedCount: 4}
	for _, line := range DayLines(types.Theme{}, state, 42) {
		if w := ansi.StringWidth(line); w > 42 || strings.Contains(line, "…") {
			t.Fatalf("expected every line whole within 42 cells, got %d: %q", w, ansi.Strip(line))
		}
	}
	if n := strings.Count(strings.ToLower(dayText(state, 42)), "overextended"); n != 1 {
		t.Fatalf("expected the focus level once, got %d", n)
	}
}

func TestDayLinesCoverAccountabilityWellbeingAndStreaks(t *testing.T) {
	rendered := squash(dayText(summaryTestState(), 42))
	for _, want := range []string{"Accountability", "Score 100%", "Done 1 of 1", "Wellbeing", "Mood 4/5", "Energy 3/5", "Momentum", "Focus 3 days", "Check-ins 4 days", "Habits 2 days"} {
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

func TestDayLinesWrapLongNotes(t *testing.T) {
	state := summaryTestState()
	note := "Productive, but sleep debt is starting to show in the afternoon slump"
	state.SummarySnapshot.CheckIn.Notes = &note
	lines := DayLines(types.Theme{}, state, 42)
	for _, line := range lines {
		if ansi.StringWidth(line) > 42 {
			t.Fatalf("expected notes within 42 cells, got %q", ansi.Strip(line))
		}
	}
	if !strings.Contains(squash(dayText(state, 42)), note) {
		t.Fatalf("expected the whole note, got %q", dayText(state, 42))
	}
}

func TestDayLinesListTopPenalties(t *testing.T) {
	state := summaryTestState()
	entry := func(id int64, score float64, highRisk, historical bool) sharedtypes.DailyPlanEntry {
		return sharedtypes.DailyPlanEntry{IssueID: id, FailScore: score, Penalty: &sharedtypes.DailyPlanEntryPenalty{
			HighRisk: highRisk, Historical: historical,
			Contribution: &sharedtypes.DailyPlanPenaltyShare{AccountabilityPct: score / 10},
		}}
	}
	state.SummarySnapshot.Plan = &sharedtypes.DailyPlan{Date: state.SummarySnapshot.Date, Entries: []sharedtypes.DailyPlanEntry{
		entry(1, 1, false, false), entry(2, 3.38, true, false), entry(3, 5, true, true),
	}}
	state.SummarySnapshot.Issues = &sharedtypes.DailyIssueSummary{Issues: []sharedtypes.Issue{
		{ID: 1, Title: "Small slip"}, {ID: 2, Title: "Fix sync conflict"},
	}}
	rendered := squash(dayText(state, 60))
	if !strings.Contains(rendered, "Top ▲3.38 34% Fix sync conflict 1.00 10% Small slip") || strings.Contains(rendered, "5.00") {
		t.Fatalf("expected top penalties largest first without resolved entries, got %q", rendered)
	}
}
