package types

import (
	"testing"
	"time"
)

func TestDayBoundaryScheduleValidation(t *testing.T) {
	valid := DayBoundarySchedule{
		Enabled:     true,
		DefaultTime: "08:30",
		WeekdayOverrides: map[int]string{
			1: "09:00",
		},
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid schedule rejected: %v", err)
	}
	for name, schedule := range map[string]DayBoundarySchedule{
		"bad default":       {DefaultTime: "8:30"},
		"bad weekday":       {DefaultTime: "08:30", WeekdayOverrides: map[int]string{0: "09:00"}},
		"bad override time": {DefaultTime: "08:30", WeekdayOverrides: map[int]string{1: "25:00"}},
	} {
		t.Run(name, func(t *testing.T) {
			if err := schedule.Validate(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestDayBoundaryScheduleWeekdayResolution(t *testing.T) {
	schedule := DayBoundarySchedule{DefaultTime: "08:00", WeekdayOverrides: map[int]string{1: "09:30", 7: "10:00"}}
	if got := schedule.TimeForWeekday(time.Monday); got != "09:30" {
		t.Fatalf("Monday override = %q", got)
	}
	if got := schedule.TimeForWeekday(time.Sunday); got != "10:00" {
		t.Fatalf("Sunday override = %q", got)
	}
	if got := schedule.TimeForWeekday(time.Wednesday); got != "08:00" {
		t.Fatalf("default time = %q", got)
	}
}
