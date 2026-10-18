package daily

import (
	"strings"
	"testing"

	"crona/tui/internal/api"
	"crona/tui/internal/tui/chrome"
	types "crona/tui/internal/tui/views/types"
)

func dailySummaryTestState() types.ContentState {
	sleep := 7.5
	avgMood := 3.1
	avgEnergy := 3.8
	avgSleep := 6.25
	return types.ContentState{
		View:          "daily",
		Pane:          "issues",
		Width:         120,
		Height:        32,
		DashboardDate: "2026-05-27",
		DailySummary: &api.DailyIssueSummary{
			Date:            "2026-05-27",
			WorkedSeconds:   7200,
			TotalIssues:     1,
			CompletedIssues: 1,
		},
		DailyIssues: []api.Issue{
			{
				ID:              1,
				Title:           "Ship momentum block",
				Status:          "done",
				EstimateMinutes: new(55),
			},
		},
		DailyPlannedIssues: []api.Issue{
			{
				ID:              1,
				Title:           "Ship momentum block",
				Status:          "done",
				EstimateMinutes: new(55),
			},
		},
		DueHabits: []api.HabitDailyItem{
			{
				HabitWithMeta: api.HabitWithMeta{
					Habit: api.Habit{Name: "Stretch"},
				},
				Status: "completed",
			},
		},
		DailyCheckIn: &api.DailyCheckIn{
			Date:       "2026-05-27",
			Mood:       4,
			Energy:     3,
			SleepHours: &sleep,
			CreatedAt:  "2026-05-27T09:00:00Z",
		},
		MetricsRollup: &api.MetricsRollup{
			Days:              14,
			WorkedSeconds:     7200,
			AverageMood:       &avgMood,
			AverageEnergy:     &avgEnergy,
			AverageSleepHours: &avgSleep,
		},
		Streaks: &api.StreakSummary{
			CurrentCheckInDays: 2,
			LongestCheckInDays: 5,
			CurrentFocusDays:   1,
			LongestFocusDays:   3,
			CurrentHabitDays:   4,
			LongestHabitDays:   9,
		},
		DailyStreaks: &api.StreakSummary{
			CurrentCheckInDays: 2,
			LongestCheckInDays: 5,
			CurrentFocusDays:   1,
			LongestFocusDays:   3,
			CurrentHabitDays:   4,
			LongestHabitDays:   9,
		},
	}
}

func testTheme() types.Theme {
	return types.Theme{
		ColorBlue:            chrome.ColorBlue,
		ColorCyan:            chrome.ColorCyan,
		ColorGreen:           chrome.ColorGreen,
		ColorMagenta:         chrome.ColorMagenta,
		ColorSubtle:          chrome.ColorSubtle,
		ColorYellow:          chrome.ColorYellow,
		ColorRed:             chrome.ColorRed,
		ColorDim:             chrome.ColorDim,
		ColorWhite:           chrome.ColorWhite,
		StyleActive:          chrome.StyleActive,
		StyleInactive:        chrome.StyleInactive,
		StylePaneTitle:       chrome.StylePaneTitle,
		StyleDim:             chrome.StyleDim,
		StyleCursor:          chrome.StyleCursor,
		StyleHeader:          chrome.StyleHeader,
		StyleError:           chrome.StyleError,
		StyleSelected:        chrome.StyleSelected,
		StyleSelectedInverse: chrome.StyleSelectedInverse,
		StyleNormal:          chrome.StyleNormal,
	}
}

func assertContainsLine(t *testing.T, lines []string, wantA, wantB string) {
	t.Helper()
	for _, line := range lines {
		if strings.Contains(line, wantA) && strings.Contains(line, wantB) {
			return
		}
	}
	t.Fatalf("expected a line containing both %q and %q, got %#v", wantA, wantB, lines)
}
