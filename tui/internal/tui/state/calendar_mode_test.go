package state

import "testing"

func TestCalendarFoldCyclesMonthWeekHidden(t *testing.T) {
	steps := []struct {
		from      CalendarMode
		monthFits bool
		want      CalendarMode
	}{
		{CalendarMonth, true, CalendarWeek},
		{CalendarWeek, true, CalendarHidden},
		{CalendarHidden, true, CalendarMonth},
		// A month without room shows as its week, so the key hides it and reopens to the week.
		{CalendarMonth, false, CalendarHidden},
		{CalendarWeek, false, CalendarHidden},
		{CalendarHidden, false, CalendarWeek},
	}
	for _, step := range steps {
		if got := NextCalendarMode(step.from, step.monthFits); got != step.want {
			t.Fatalf("NextCalendarMode(%d, %v) = %d, want %d", step.from, step.monthFits, got, step.want)
		}
	}
}

func TestWithCalendarModeRoundTrips(t *testing.T) {
	base := []string{CollapseDailyHabits}
	for _, mode := range []CalendarMode{CalendarMonth, CalendarWeek, CalendarHidden} {
		got := WithCalendarMode(WithCalendarMode(base, CalendarHidden), mode)
		if DailyCalendarMode(got) != mode || !PaneCollapsed(got, CollapseDailyHabits) {
			t.Fatalf("mode %d gave %v", mode, got)
		}
	}
}
