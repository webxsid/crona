package testsuite

import (
	"strings"
	"testing"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	settingsmeta "crona/tui/internal/tui/views/settingsmeta"
	"crona/tui/internal/tui/views"

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
	if rowMap["Wipe Runtime Data"] != "Destructive" {
		t.Fatalf("expected renamed wipe row, got %q", rowMap["Wipe Runtime Data"])
	}
}

func TestWellbeingViewShowsTodayAndAverageDurationLabels(t *testing.T) {
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
}

func TestWellbeingViewShowsBackfilledIndicator(t *testing.T) {
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
}
