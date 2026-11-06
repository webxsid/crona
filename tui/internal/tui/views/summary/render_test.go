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
	state.SummarySnapshot.Focus = &api.FocusScoreSummary{
		Score:               82,
		Level:               sharedtypes.FocusScoreLevelStrong,
		Reason:              sharedtypes.FocusScoreReasonBalanced,
		TargetWorkedSeconds: 3600,
	}
	state.SummarySnapshot.Metrics = &api.DailyMetricsDay{
		WorkedSeconds:       2400,
		RestSeconds:         600,
		SessionCount:        3,
		TotalIssues:         4,
		CompletedIssues:     2,
		AbandonedIssues:     1,
		HabitDueCount:       3,
		HabitCompletedCount: 2,
		HabitFailedCount:    1,
	}
	rendered := dayText(state, 42)
	for _, want := range []string{"Stats", "Focus 82/100 Strong", "Worked 40m of 1h", "Rest 10m", "Sessions 3", "Issues 2 of 4 done · 1 abandoned", "Habits 2 of 3 done · 1 failed"} {
		if !strings.Contains(squash(rendered), want) {
			t.Fatalf("expected summary stats to contain %q, got %q", want, rendered)
		}
	}
}

func TestDayLinesOneStatPerRowWithinWidth(t *testing.T) {
	state := summaryTestState()
	state.SummarySnapshot.Focus = &api.FocusScoreSummary{
		Score:               52,
		Level:               sharedtypes.FocusScoreLevelOverextended,
		Reason:              "overextended",
		TargetWorkedSeconds: 21600,
	}
	state.SummarySnapshot.Metrics = &api.DailyMetricsDay{
		WorkedSeconds:       6660,
		SessionCount:        2,
		TotalIssues:         4,
		CompletedIssues:     1,
		HabitDueCount:       5,
		HabitCompletedCount: 4,
	}
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
	for _, want := range []string{"Accountability", "Score 100%", "Done 1 of 1", "Wellbeing", "Wellbeing", "4/5", "3/5", "Momentum", "Focus ↑3d", "Check-ins ↑4d"} {
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
				HabitWithMeta: api.HabitWithMeta{
					Habit: api.Habit{Name: "Walk", TargetMinutes: &target},
				},
				Completed: true,
				Status:    sharedtypes.HabitCompletionStatusCompleted,
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
			Streaks: &api.StreakSummary{
				CurrentFocusDays:   3,
				CurrentCheckInDays: 4,
				CurrentHabitDays:   2,
			},
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
		return sharedtypes.DailyPlanEntry{
			IssueID:   id,
			FailScore: score,
			Penalty: &sharedtypes.DailyPlanEntryPenalty{
				HighRisk: highRisk, Historical: historical,
				Contribution: &sharedtypes.DailyPlanPenaltyShare{AccountabilityPct: score / 10},
			},
		}
	}
	state.SummarySnapshot.Plan = &sharedtypes.DailyPlan{
		Date: state.SummarySnapshot.Date,
		Entries: []sharedtypes.DailyPlanEntry{
			entry(1, 1, false, false), entry(2, 3.38, true, false), entry(3, 5, true, true),
		},
	}
	state.SummarySnapshot.Issues = &sharedtypes.DailyIssueSummary{Issues: []sharedtypes.Issue{
		{ID: 1, Title: "Small slip"}, {ID: 2, Title: "Fix sync conflict"},
	}}
	rendered := squash(dayText(state, 60))
	if !strings.Contains(rendered, "Top ▲3.38 █░░░░░ 34% Fix sync conflict 1.00 ░░░░░░ 10% Small slip") ||
		strings.Contains(rendered, "5.00") {
		t.Fatalf("expected top penalties largest first without resolved entries, got %q", rendered)
	}
}

func TestDayLineFitsTheDayInOneLine(t *testing.T) {
	state := summaryTestState()
	state.SummarySnapshot.Focus = &api.FocusScoreSummary{
		Score: 82, Level: sharedtypes.FocusScoreLevelStrong, TargetWorkedSeconds: 3600,
	}
	state.SummarySnapshot.Metrics = &api.DailyMetricsDay{WorkedSeconds: 2400, RestSeconds: 600, SessionCount: 3}
	wide := ansi.Strip(DayLine(types.Theme{}, state, 120))
	for _, want := range []string{"Focus 82 strong", "Worked 40m/1h", "Rest 10m", "Sessions 3"} {
		if !strings.Contains(wide, want) {
			t.Fatalf("expected the day line to contain %q, got %q", want, wide)
		}
	}
	narrow := ansi.Strip(DayLine(types.Theme{}, state, 40))
	if ansi.StringWidth(narrow) > 40 || !strings.Contains(narrow, "Focus") || strings.Contains(narrow, "Sessions") {
		t.Fatalf("expected parts to drop from the right at 40, got %q", narrow)
	}
	state.SummarySnapshot.Focus = nil
	if line := DayLine(types.Theme{}, state, 120); line != "" {
		t.Fatalf("expected no day line before the summary loads, got %q", line)
	}
}

func TestDayLinesCarryTheStripsBarsAndBestStreaks(t *testing.T) {
	state := summaryTestState()
	state.SummarySnapshot.Focus = &api.FocusScoreSummary{Score: 60, TargetWorkedSeconds: 3600}
	state.SummarySnapshot.Metrics = &api.DailyMetricsDay{
		TotalIssues: 2, CompletedIssues: 1, HabitDueCount: 2, HabitCompletedCount: 1,
	}
	state.DailyPlannedIssues = []api.Issue{{Status: "done"}, {Status: "planned"}}
	lines := strings.Split(dayText(state, 42), "\n")
	for _, label := range []string{"Issues", "Habits"} {
		for i, line := range lines {
			if strings.HasPrefix(line, label) {
				if i+1 >= len(lines) || !strings.Contains(lines[i+1], "█") {
					t.Fatalf("expected a bar under %s, got %q", label, lines)
				}
				break
			}
		}
	}
	if rendered := squash(strings.Join(lines, "\n")); !strings.Contains(rendered, "best") {
		t.Fatalf("expected best streaks under Momentum, got %q", rendered)
	}
}

func TestWellbeingShowsArcGaugesAndTheWeekTrend(t *testing.T) {
	state := summaryTestState()
	sleep, screen := 7.5, 300
	state.SummarySnapshot.CheckIn.SleepHours = &sleep
	state.SummarySnapshot.CheckIn.ScreenTimeMinutes = &screen
	avgMood, avgSleep, avgScreen := 3.0, 7.4, 120.0
	state.SummarySnapshot.Rollup = &sharedtypes.MetricsRollup{
		AverageMood: &avgMood, AverageSleepHours: &avgSleep, AverageScreenTimeMins: &avgScreen,
	}
	state.Settings = &sharedtypes.CoreSettings{SleepTargetMinutes: 480, ScreenTimeLimitMins: 360}
	rendered := squash(dayText(state, 78))
	// 5h of a 6h limit leaves 1h; the week averaged 2h, so 4h left, and today is worse.
	for _, want := range []string{"4/5", "3/5", "7h30m", "1h", "Mood Energy Sleep Screen left", "▲ 7d 3.0", "= 7d 7h24m", "▼ 7d 4h", "⣿"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("expected %q in the wellbeing gauges, got %q", want, rendered)
		}
	}
	screen = 400
	if rendered := squash(dayText(state, 78)); !strings.Contains(rendered, "+40m") {
		t.Fatalf("expected screen time over the limit as +40m, got %q", rendered)
	}
	for _, line := range DayLines(types.Theme{}, state, 30) {
		if w := ansi.StringWidth(line); w > 30 {
			t.Fatalf("expected gauges to wrap within 30 cells, got %d: %q", w, ansi.Strip(line))
		}
	}
}

func TestMomentumListsTheDaysMomentumsUnmetFirst(t *testing.T) {
	state := summaryTestState()
	state.SummaryDate = "2026-07-28"
	state.DueHabits = []api.HabitDailyItem{{HabitWithMeta: api.HabitWithMeta{Habit: api.Habit{ID: 7}}}}
	state.AllIssues = []api.IssueWithMeta{{Issue: api.Issue{ID: 3, StreamID: 9}, RepoID: 4}}
	state.DailyPlannedIssues = []api.Issue{{ID: 3, StreamID: 9}}
	point := func(count, target int, met bool) []sharedtypes.MomentumSeriesPoint {
		return []sharedtypes.MomentumSeriesPoint{{StartDate: "2026-07-27", EndDate: "2026-08-02", Count: count, Target: target, MetTarget: met}}
	}
	state.MomentumCards = []api.MomentumCard{
		{Definition: api.HabitStreakDefinition{ID: "fit", Name: "Fitness", Enabled: true, Period: "week", RequiredCount: 3, HabitIDs: []int64{7}},
			Current: 4, Longest: 9, Series: point(3, 3, true)},
		{Definition: api.HabitStreakDefinition{ID: "core", Name: "Core work", Enabled: true, TargetKind: sharedtypes.MomentumTargetKindContext,
			Period: "week", RequiredCount: 1, Contexts: []sharedtypes.MomentumContext{{RepoID: 4}}},
			Series: point(3600, 7200, false)},
		{Definition: api.HabitStreakDefinition{ID: "other", Name: "Unrelated", Enabled: true, Period: "week", HabitIDs: []int64{99}}},
	}
	text := dayText(state, 80)
	core, fit := strings.Index(text, "Core work"), strings.Index(text, "Fitness")
	if core < 0 || fit < 0 || core > fit {
		t.Fatalf("expected the unmet Core work before the met Fitness, got %q", text)
	}
	for _, want := range []string{"3/3 ✓ ↑4w best 9w", "1h/2h", "▰▰▰▱▱▱"} {
		if !strings.Contains(squash(text), want) {
			t.Fatalf("expected %q in the momentum rows, got %q", want, text)
		}
	}
	if strings.Contains(text, "Unrelated") {
		t.Fatalf("expected momentums the day does not feed left out, got %q", text)
	}
}
