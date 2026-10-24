package settingsmeta

import (
	"testing"

	sharedtypes "crona/shared/types"
)

func TestRowsWithTimezoneIncludesDayBoundaries(t *testing.T) {
	rows := RowsWithTimezone(&sharedtypes.CoreSettings{
		StartOfDay: sharedtypes.DayBoundarySchedule{
			Enabled:     true,
			DefaultTime: "08:30",
			WeekdayOverrides: map[int]string{
				1: "09:00",
			},
		},
		EndOfDay: sharedtypes.DayBoundarySchedule{Enabled: false, DefaultTime: "18:00"},
	}, "Asia/Kolkata")
	values := map[string]string{}
	for _, row := range rows {
		values[row.Label] = row.Value
	}
	if values["Start of Day"] != "Enabled · 08:30 · 1 overrides · Asia/Kolkata" {
		t.Fatalf("unexpected SOD label: %q", values["Start of Day"])
	}
	if values["End of Day"] != "Disabled · Asia/Kolkata" {
		t.Fatalf("unexpected EOD label: %q", values["End of Day"])
	}
}

func TestTagRowsListVocabularyAndCountUnused(t *testing.T) {
	red := "red"
	rows := TagRows([]sharedtypes.Tag{
		{Name: "urgent", Color: &red, System: true},
		{Name: "release", IssueCount: 2},
		{Name: "old"},
	})
	if len(rows) != 4 || rows[0].Tag != "urgent" || rows[0].Label != "+urgent" {
		t.Fatalf("unexpected rows %+v", rows)
	}
	if rows[0].Value != "0 issues, 0 habits · red · system" {
		t.Fatalf("unexpected system row value %q", rows[0].Value)
	}
	if last := rows[3]; last.Label != "Remove Unused Tags" || last.Value != "1 unused" {
		t.Fatalf("unexpected remove-unused row %+v", last)
	}
}
