package testsuite

import (
	"strings"
	"testing"

	"crona/tui/internal/api"
	uistate "crona/tui/internal/tui/state"
	"crona/tui/internal/tui/views"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

func strPtr(v string) *string { return &v }

func assertTinySummary(t *testing.T, rendered string) {
	t.Helper()
	if !strings.Contains(rendered, "Daily Dashboard") {
		t.Fatalf("expected dashboard title in tiny summary")
	}
	if !strings.Contains(rendered, "2026-03-19") {
		t.Fatalf("expected date in tiny summary")
	}
	if !strings.Contains(rendered, "Scope: Work > app") {
		t.Fatalf("expected scope in tiny summary")
	}
	if !strings.Contains(rendered, "[,] [.] [g]") {
		t.Fatalf("expected compact date hints in tiny summary")
	}
	if !strings.Contains(rendered, "Issues  0/1") || !strings.Contains(rendered, "Habits  0/1") {
		t.Fatalf("expected both issue and habit summary rows in tiny summary")
	}
	if !strings.Contains(rendered, "p1") {
		t.Fatalf("expected abbreviated issue legend in tiny summary")
	}
	if !strings.Contains(rendered, "f0 r1") {
		t.Fatalf("expected abbreviated habit tail in tiny summary")
	}
	if !strings.Contains(rendered, "█") {
		t.Fatalf("expected micro-bars in tiny summary")
	}
}

func compactDefaultState(height int) views.ContentState {
	estimate1, estimate2, estimate3 := 60, 35, 25
	today := "2026-03-19"
	return views.ContentState{
		View:   "default",
		Pane:   "issues",
		Width:  92,
		Height: height,
		Cursors: map[string]int{
			"issues": 0,
		},
		Filters: map[string]string{
			"issues": "",
		},
		DefaultIssueSection: "open",
		DefaultIssues: []api.IssueWithMeta{
			{Issue: api.Issue{ID: 1, Title: "Add keyboard-first command palette", Status: "planned", EstimateMinutes: &estimate1, TodoForDate: &today}, RepoName: "Work", StreamName: "app"},
			{Issue: api.Issue{ID: 2, Title: "Improve install docs for Linux", Status: "planned", EstimateMinutes: &estimate2, TodoForDate: &today}, RepoName: "OSS", StreamName: "cli"},
			{Issue: api.Issue{ID: 3, Title: "Research standing desk options", Status: "abandoned", EstimateMinutes: &estimate3}, RepoName: "Personal", StreamName: "home"},
		},
		Context: &api.ActiveContext{},
	}
}

func compactWellbeingState(height int) views.ContentState {
	avgMood, avgEnergy := 4.0, 3.7
	return views.ContentState{
		View:          "wellbeing",
		Pane:          string(uistate.PaneWellbeingSummary),
		Width:         92,
		Height:        height,
		WellbeingDate: "2026-03-19",
		DailyPlan: &api.DailyPlan{
			Date: "2026-03-19",
			Summary: api.DailyPlanAccountabilitySummary{
				PlannedCount:         3,
				CompletedCount:       1,
				FailedCount:          1,
				PendingRollbackCount: 1,
				AccountabilityScore:  2.6,
				DelayedIssueCount:    1,
			},
			Entries: []api.DailyPlanEntry{
				{ID: "1", Date: "2026-03-19", IssueID: 1, Status: "completed", CommittedAt: "2026-03-19T08:00:00Z"},
				{ID: "2", Date: "2026-03-19", IssueID: 2, Status: "failed", CommittedAt: "2026-03-19T08:00:00Z", FailureReason: dailyPlanFailureReasonPtr("moved")},
				{ID: "3", Date: "2026-03-19", IssueID: 3, Status: "planned", CommittedAt: "2026-03-19T08:00:00Z", PendingFailureAt: strPtr("2026-03-19T10:00:00Z")},
			},
		},
		MetricsRollup: &api.MetricsRollup{
			Days:          7,
			CheckInDays:   6,
			FocusDays:     1,
			WorkedSeconds: 4956,
			RestSeconds:   2,
			AverageMood:   &avgMood,
			AverageEnergy: &avgEnergy,
			LatestBurnout: &api.BurnoutIndicator{
				Level:   "low",
				Score:   31,
				Factors: map[string]float64{"workloadPressure": 0.22, "recoveryBreaks": -0.14},
			},
		},
		Streaks: &api.StreakSummary{
			CurrentCheckInDays: 0,
			LongestCheckInDays: 6,
			CurrentFocusDays:   0,
			LongestFocusDays:   1,
		},
	}
}

func assertCompactDefault(t *testing.T, rendered string, height int) {
	t.Helper()
	plain := ansi.Strip(rendered)
	if !strings.Contains(plain, "Default Dashboard") {
		t.Fatalf("expected compact stats header in default view")
	}
	if !strings.Contains(plain, "Active Issues [1]") {
		t.Fatalf("expected primary issue list in compact default view")
	}
	if !strings.Contains(plain, "Closed") {
		t.Fatalf("expected compact completed footer in default view")
	}
	if !strings.Contains(plain, "Add keyboard-first command palette") {
		t.Fatalf("expected open issue rows in compact default view")
	}
	if got := lipgloss.Height(rendered); got > height {
		t.Fatalf("default compact view height %d exceeds allocated height %d", got, height)
	}
}

func assertCompactWellbeing(t *testing.T, rendered string, height int) {
	t.Helper()
	plain := ansi.Strip(rendered)
	if !strings.Contains(plain, "Wellbeing") || !strings.Contains(plain, "2026-03-19") {
		t.Fatalf("expected compact wellbeing header")
	}
	if !strings.Contains(plain, "[,/.]") && !strings.Contains(plain, "[a/e]") {
		t.Fatalf("expected action hints in compact wellbeing view")
	}
	if !strings.Contains(plain, "No check-in recorded for this date") {
		t.Fatalf("expected current day summary in compact wellbeing view")
	}
	if !strings.Contains(plain, "Burnout") || !strings.Contains(plain, "31 LOW") {
		t.Fatalf("expected burnout summary in compact wellbeing view")
	}
	if !strings.Contains(plain, "Metrics Window") || !strings.Contains(plain, "Days  7") {
		t.Fatalf("expected compact metrics block in wellbeing view")
	}
	if !strings.Contains(plain, "Accountability") || !strings.Contains(plain, "Planned 3  Completed 1  Failed 1  Pending 1") {
		t.Fatalf("expected compact accountability summary in wellbeing view")
	}
	if got := lipgloss.Height(rendered); got > height {
		t.Fatalf("wellbeing compact view height %d exceeds allocated height %d", got, height)
	}
}

func dailyPlanFailureReasonPtr(value string) *api.DailyPlanFailureReason {
	reason := api.DailyPlanFailureReason(value)
	return &reason
}
