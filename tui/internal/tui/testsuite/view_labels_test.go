package testsuite

import (
	"testing"

	sharedtypes "crona/shared/types"
	settingsmeta "crona/tui/internal/tui/views/settingsmeta"
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
		TimerMode:             sharedtypes.TimerModeStopwatch,
		UpdateChannel:         sharedtypes.UpdateChannelStable,
		RepoSort:              sharedtypes.RepoSortAlphabeticalAsc,
		StreamSort:            sharedtypes.StreamSortAlphabeticalAsc,
		IssueSort:             sharedtypes.IssueSortPriority,
		HabitSort:             sharedtypes.HabitSortSchedule,
		DateDisplayPreset:     sharedtypes.DateDisplayPresetCustom,
		DateDisplayFormat:     "Do MMM YYYY",
		PromptGlyphMode:       sharedtypes.PromptGlyphModeUnicode,
		WeekStart:             sharedtypes.WeekStartSunday,
	})

	rowMap := map[string]string{}
	for _, row := range rows {
		rowMap[row.Label] = row.Value
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
	if rowMap["Week Start"] != "Sunday" {
		t.Fatalf("expected Week Start to be Sunday, got %q", rowMap["Week Start"])
	}
	if rowMap["Wipe Runtime Data"] != "Destructive" {
		t.Fatalf("expected renamed wipe row, got %q", rowMap["Wipe Runtime Data"])
	}
	for _, unwanted := range []string{
		"Timer Mode",
		"Work Duration",
		"Breaks",
		"Short Break",
		"Long Break",
		"Long Breaks",
		"Cycles Before Long Break",
		"Auto-Start Breaks",
		"Auto-Start Work",
	} {
		if _, ok := rowMap[unwanted]; ok {
			t.Fatalf("expected timer settings row %q to be hidden", unwanted)
		}
	}
}
