package dayboundary

import (
	"context"
	"testing"
	"time"

	sharedtypes "crona/shared/types"
)

func TestLogicalDateUsesSODBoundary(t *testing.T) {
	location := time.FixedZone("Test", 5*60*60+30*60)
	schedule := sharedtypes.DayBoundarySchedule{Enabled: true, DefaultTime: "08:30"}
	before := time.Date(2026, 7, 29, 8, 29, 59, 999999999, location)
	after := time.Date(2026, 7, 29, 8, 30, 0, 0, location)
	if got := LogicalDate(before, schedule); got != "2026-07-28" {
		t.Fatalf("date before SOD = %s", got)
	}
	if got := LogicalDate(after, schedule); got != "2026-07-29" {
		t.Fatalf("date after SOD = %s", got)
	}
}

func TestDateRecorderRecordsCompletedDateOnlyOnRollover(t *testing.T) {
	var dates []string
	scheduler := &Scheduler{record: func(_ context.Context, date string, _ *sharedtypes.CoreSettings) error {
		dates = append(dates, date)
		return nil
	}, recordKey: "2026-08-08"}
	settings := &sharedtypes.CoreSettings{RestWeekdays: []int{6}}
	ctx := context.Background()
	if err := scheduler.recordDateIfChanged(ctx, "2026-08-08", settings); err != nil {
		t.Fatalf("record date: %v", err)
	}
	settings.RestSpecificDates = []string{"2026-08-08"}
	if err := scheduler.recordDateIfChanged(ctx, "2026-08-08", settings); err != nil {
		t.Fatalf("record changed rules: %v", err)
	}
	if err := scheduler.recordDateIfChanged(ctx, "2026-08-09", settings); err != nil {
		t.Fatalf("record rollover: %v", err)
	}
	if len(dates) != 1 || dates[0] != "2026-08-08" {
		t.Fatalf("expected only completed date at rollover, got %v", dates)
	}
}

func TestNextBoundaryOrdersSODBeforeEOD(t *testing.T) {
	location := time.FixedZone("Test", 0)
	now := time.Date(2026, 7, 29, 7, 0, 0, 0, location)
	settings := &sharedtypes.CoreSettings{
		StartOfDay: sharedtypes.DayBoundarySchedule{Enabled: true, DefaultTime: "08:00"},
		EndOfDay:   sharedtypes.DayBoundarySchedule{Enabled: true, DefaultTime: "08:00"},
	}
	at, kind, _ := nextBoundary(now, settings)
	if !at.Equal(time.Date(2026, 7, 29, 8, 0, 0, 0, location)) || kind != sharedtypes.DayBoundaryStart {
		t.Fatalf("next boundary = %s %s", at, kind)
	}
}

func TestNextBoundarySkipsDisabledSchedule(t *testing.T) {
	location := time.FixedZone("Test", 0)
	now := time.Date(2026, 7, 29, 7, 0, 0, 0, location)
	settings := &sharedtypes.CoreSettings{
		StartOfDay: sharedtypes.DayBoundarySchedule{Enabled: true, DefaultTime: "08:00"},
		EndOfDay:   sharedtypes.DayBoundarySchedule{Enabled: false, DefaultTime: "17:00"},
	}
	_, kind, schedule := nextBoundary(now, settings)
	if kind != sharedtypes.DayBoundaryStart || !schedule.Enabled {
		t.Fatalf("unexpected disabled schedule candidate: %s %+v", kind, schedule)
	}
}
