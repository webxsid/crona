package testsuite

import (
	"strings"
	"testing"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	uistate "crona/tui/internal/tui/state"
	"crona/tui/internal/tui/views"
	settingsmeta "crona/tui/internal/tui/views/settingsmeta"

	"crona/tui/internal/tui/testsuite/support"
)

func TestSettingsRowsUseClearerStatusLabels(t *testing.T) {
	rows := settingsmeta.Rows(&sharedtypes.CoreSettings{
		BreaksEnabled:         true,
		LongBreakEnabled:      false,
		AutoStartBreaks:       true,
		AutoStartWork:         false,
		BoundaryNotifications: true,
		BoundarySound:         false,
		UpdateChecksEnabled:   true,
		UpdatePromptEnabled:   false,
		AwayModeEnabled:       true,
		WorkDurationMinutes:   25,
		ShortBreakMinutes:     5,
		LongBreakMinutes:      15,
		CyclesBeforeLongBreak: 4,
		DailyPlanRollbackMins: 5,
		TimerMode:             sharedtypes.TimerModeStructured,
		UpdateChannel:         sharedtypes.UpdateChannelStable,
		RepoSort:              sharedtypes.RepoSortAlphabeticalAsc,
		StreamSort:            sharedtypes.StreamSortAlphabeticalAsc,
		IssueSort:             sharedtypes.IssueSortPriority,
		HabitSort:             sharedtypes.HabitSortSchedule,
		DateDisplayPreset:     sharedtypes.DateDisplayPresetCustom,
		DateDisplayFormat:     "Do MMM YYYY",
		PromptGlyphMode:       sharedtypes.PromptGlyphModeUnicode,
	})

	rowMap := map[string]string{}
	for _, row := range rows {
		rowMap[row.Label] = row.Value
	}

	if rowMap["Breaks"] != "Enabled" {
		t.Fatalf("expected Breaks to be Enabled, got %q", rowMap["Breaks"])
	}
	if rowMap["Long Breaks"] != "Disabled" {
		t.Fatalf("expected Long Breaks to be Disabled, got %q", rowMap["Long Breaks"])
	}
	if rowMap["Date Format"] != "Custom" {
		t.Fatalf("expected Date Format to be Custom, got %q", rowMap["Date Format"])
	}
	if rowMap["Custom Date Format"] != "Do MMM YYYY" {
		t.Fatalf("expected Custom Date Format row, got %q", rowMap["Custom Date Format"])
	}
	if rowMap["Prompt Glyphs"] != "Unicode" {
		t.Fatalf("expected Prompt Glyphs row to be Unicode, got %q", rowMap["Prompt Glyphs"])
	}
	if rowMap["Wipe Runtime Data"] != "Destructive" {
		t.Fatalf("expected renamed wipe row, got %q", rowMap["Wipe Runtime Data"])
	}
}

func TestWellbeingViewLabels(t *testing.T) {
	t.Run("today and average duration", func(t *testing.T) {
		sleep := 7.5
		state := views.ContentState{
			View:          "wellbeing",
			Pane:          "wellbeing",
			Width:         120,
			Height:        40,
			WellbeingDate: "2026-04-04",
			DailyCheckIn: &api.DailyCheckIn{
				Date:       "2026-04-04",
				Mood:       4,
				Energy:     3,
				SleepHours: &sleep,
			},
		}

		rendered := support.RenderWellbeing(state)
		if !strings.Contains(rendered, "today 7h30m") {
			t.Fatalf("expected today sleep label in wellbeing view, got %q", rendered)
		}

		state.DailyCheckIn = nil
		state.MetricsRollup = &api.MetricsRollup{AverageSleepHours: &sleep}
		rendered = support.RenderWellbeing(state)
		if !strings.Contains(rendered, "7d avg 7h30m") {
			t.Fatalf("expected average sleep label in wellbeing view, got %q", rendered)
		}
	})

	t.Run("backfilled indicator", func(t *testing.T) {
		sleep := 7.5
		screen := 80
		createdAt := "2026-04-05T11:30:00Z"
		state := views.ContentState{
			View:          "wellbeing",
			Pane:          "wellbeing",
			Width:         120,
			Height:        36,
			WellbeingDate: "2026-04-04",
			DailyCheckIn: &api.DailyCheckIn{
				Date:              "2026-04-04",
				Mood:              4,
				Energy:            3,
				SleepHours:        &sleep,
				ScreenTimeMinutes: &screen,
				CreatedAt:         createdAt,
			},
		}

		rendered := support.RenderWellbeing(state)
		if !strings.Contains(rendered, "Backfilled check-in") {
			t.Fatalf("expected backfilled indicator in wellbeing view, got %q", rendered)
		}
	})
}

func TestWellbeingSummaryPaneSupportsScrolling(t *testing.T) {
	sleep := 7.5
	screen := 140
	sleepScore := 82
	backfilledAt := "2026-04-05T11:30:00Z"
	state := views.ContentState{
		View:          "wellbeing",
		Pane:          string(uistate.PaneWellbeingSummary),
		Width:         100,
		Height:        40,
		WellbeingDate: "2026-04-04",
		Cursors: map[string]int{
			string(uistate.PaneWellbeingSummary): 0,
			string(uistate.PaneWellbeingTrends):  0,
		},
		DailyCheckIn: &api.DailyCheckIn{
			Date:              "2026-04-04",
			Mood:              4,
			Energy:            3,
			SleepHours:        &sleep,
			SleepScore:        &sleepScore,
			ScreenTimeMinutes: &screen,
			CreatedAt:         backfilledAt,
		},
		DailyPlan: &api.DailyPlan{
			Date: "2026-04-04",
			Summary: api.DailyPlanAccountabilitySummary{
				PlannedCount:         5,
				CompletedCount:       2,
				FailedCount:          2,
				PendingRollbackCount: 1,
				AccountabilityScore:  2.2,
				BacklogPressure:      1.8,
				DelayedIssueCount:    3,
				HighRiskIssueCount:   2,
			},
			Entries: []api.DailyPlanEntry{
				{ID: "1", IssueID: 1, Status: "failed", FailureReason: dailyPlanFailureReasonPtr("missed")},
				{ID: "2", IssueID: 2, Status: "failed", FailureReason: dailyPlanFailureReasonPtr("moved")},
			},
		},
		MetricsRange: makeMetricsRange(21),
		MetricsRollup: &api.MetricsRollup{
			Days:              21,
			CheckInDays:       15,
			FocusDays:         11,
			WorkedSeconds:     7200,
			AverageSleepHours: &sleep,
		},
	}

	rendered := support.RenderWellbeing(state)
	if !strings.Contains(rendered, "↓ more") {
		t.Fatalf("expected summary pane to show downward scroll indicator, got %q", rendered)
	}

	state.Cursors[string(uistate.PaneWellbeingSummary)] = 16
	rendered = support.RenderWellbeing(state)
	if !strings.Contains(rendered, "↑ more") {
		t.Fatalf("expected summary pane to show upward scroll indicator after scrolling, got %q", rendered)
	}
}

func TestWellbeingTrendsPaneSupportsScrolling(t *testing.T) {
	avgMood := 4.2
	avgEnergy := 3.8
	state := views.ContentState{
		View:          "wellbeing",
		Pane:          string(uistate.PaneWellbeingTrends),
		Width:         100,
		Height:        40,
		WellbeingDate: "2026-04-04",
		Cursors: map[string]int{
			string(uistate.PaneWellbeingSummary): 0,
			string(uistate.PaneWellbeingTrends):  0,
		},
		MetricsRange: makeMetricsRange(14),
		MetricsRollup: &api.MetricsRollup{
			Days:          14,
			CheckInDays:   10,
			FocusDays:     8,
			WorkedSeconds: 14400,
			RestSeconds:   2400,
			SessionCount:  9,
			AverageMood:   &avgMood,
			AverageEnergy: &avgEnergy,
			LatestBurnout: &api.BurnoutIndicator{
				Level: "guarded",
				Score: 58,
				Factors: map[string]float64{
					"workloadPressure":    0.36,
					"breakDebt":           0.22,
					"moodEnergyDrag":      0.19,
					"sleepDebt":           0.17,
					"recoveryConsistency": -0.12,
					"recoveryBreaks":      -0.09,
				},
			},
		},
		Streaks: &api.StreakSummary{
			CurrentCheckInDays: 3,
			LongestCheckInDays: 7,
			CurrentFocusDays:   2,
			LongestFocusDays:   5,
		},
	}

	rendered := support.RenderWellbeing(state)
	if !strings.Contains(rendered, "↓ more") {
		t.Fatalf("expected trends pane to show downward scroll indicator, got %q", rendered)
	}

	state.Cursors[string(uistate.PaneWellbeingTrends)] = 12
	rendered = support.RenderWellbeing(state)
	if !strings.Contains(rendered, "↑ more") {
		t.Fatalf("expected trends pane to show upward scroll indicator after scrolling, got %q", rendered)
	}
}

func makeMetricsRange(days int) []api.DailyMetricsDay {
	out := make([]api.DailyMetricsDay, 0, days)
	for i := 0; i < days; i++ {
		mood := 3 + (i % 3)
		energy := 2 + (i % 4)
		sleep := 6.5 + float64(i%3)*0.5
		out = append(out, api.DailyMetricsDay{
			Date:          "2026-04-04",
			SessionCount:  1 + (i % 4),
			WorkedSeconds: 1800 + i*300,
			RestSeconds:   300 + i*60,
			CheckIn: &api.DailyCheckIn{
				Date:       "2026-04-04",
				Mood:       mood,
				Energy:     energy,
				SleepHours: &sleep,
			},
		})
	}
	return out
}
