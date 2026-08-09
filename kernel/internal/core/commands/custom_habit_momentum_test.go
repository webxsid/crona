package commands

import (
	"fmt"
	"testing"
	"time"

	sharedtypes "crona/shared/types"
)

func TestMomentumTypesContinueAfterProtectedBucket(t *testing.T) {
	tests := []struct {
		name           string
		period         sharedtypes.HabitStreakPeriod
		firstStart     string
		protectedStart string
		protectedEnd   string
		lastEnd        string
		completedDates []string
	}{
		{
			name:           "daily",
			period:         sharedtypes.HabitStreakPeriodDay,
			firstStart:     "2026-01-05",
			protectedStart: "2026-01-06",
			protectedEnd:   "2026-01-06",
			lastEnd:        "2026-01-07",
			completedDates: []string{"2026-01-05", "2026-01-07"},
		},
		{
			name:           "weekly",
			period:         sharedtypes.HabitStreakPeriodWeek,
			firstStart:     "2026-01-05",
			protectedStart: "2026-01-12",
			protectedEnd:   "2026-01-18",
			lastEnd:        "2026-01-25",
			completedDates: []string{"2026-01-05", "2026-01-19"},
		},
		{
			name:           "monthly",
			period:         sharedtypes.HabitStreakPeriodMonth,
			firstStart:     "2026-01-01",
			protectedStart: "2026-02-01",
			protectedEnd:   "2026-02-28",
			lastEnd:        "2026-03-31",
			completedDates: []string{"2026-01-01", "2026-03-01"},
		},
	}
	targetKinds := []sharedtypes.MomentumTargetKind{
		sharedtypes.MomentumTargetKindHabit,
		sharedtypes.MomentumTargetKindContext,
	}
	matchModes := []sharedtypes.MomentumMatchMode{
		sharedtypes.MomentumMatchModeAny,
		sharedtypes.MomentumMatchModeAll,
	}

	for _, tc := range tests {
		for _, targetKind := range targetKinds {
			for _, matchMode := range matchModes {
				name := fmt.Sprintf("%s/%s/%s", tc.name, targetKind, matchMode)
				t.Run(name, func(t *testing.T) {
					def := sharedtypes.HabitStreakDefinition{
						ID:            name,
						Name:          name,
						Enabled:       true,
						Period:        tc.period,
						RequiredCount: 1,
						TargetKind:    targetKind,
						MatchMode:     matchMode,
					}
					settings := &sharedtypes.CoreSettings{
						AwayDates: momentumTestDates(tc.protectedStart, tc.protectedEnd),
					}
					counts := make(map[string]map[string]int, len(tc.completedDates))
					for _, date := range tc.completedDates {
						counts[date] = map[string]int{def.ID: 1}
					}

					state := customHabitMomentumSnapshotState{}
					for _, date := range momentumTestDates(tc.firstStart, tc.lastEnd) {
						state = advanceCustomHabitMomentumSnapshotState(
							state,
							date,
							counts,
							[]sharedtypes.HabitStreakDefinition{def},
							settings,
						)
					}

					got := state.Definitions[0]
					if got.Current != 2 || got.Longest != 2 {
						t.Fatalf("expected protected bucket to preserve and continue streak, got %+v", got)
					}
					rangeCurrent, rangeLongest := computeMomentumSummaryFromCounts(
						def,
						counts,
						tc.firstStart,
						tc.lastEnd,
						settings,
					)
					if got.Current != rangeCurrent || got.Longest != rangeLongest {
						t.Fatalf(
							"expected snapshot and range calculations to agree, snapshot=%+v range=%d/%d",
							got,
							rangeCurrent,
							rangeLongest,
						)
					}
				})
			}
		}
	}
}

func TestDailyMomentumContinuesAfterProtectedDay(t *testing.T) {
	def := sharedtypes.HabitStreakDefinition{
		ID:            "journal",
		Name:          "Daily Journal",
		Enabled:       true,
		Period:        sharedtypes.HabitStreakPeriodDay,
		RequiredCount: 1,
		TargetKind:    sharedtypes.MomentumTargetKindHabit,
	}
	settings := &sharedtypes.CoreSettings{AwayDates: []string{"2026-08-08"}}
	counts := map[string]map[string]int{
		"2026-08-07": {def.ID: 1},
		"2026-08-09": {def.ID: 1},
	}

	state := customHabitMomentumSnapshotState{}
	for _, date := range []string{"2026-08-07", "2026-08-08", "2026-08-09"} {
		state = advanceCustomHabitMomentumSnapshotState(state, date, counts, []sharedtypes.HabitStreakDefinition{def}, settings)
	}

	got := state.Definitions[0]
	if got.Current != 2 || got.Longest != 2 {
		t.Fatalf("expected protected day to preserve and continue streak, got %+v", got)
	}
	rangeCurrent, rangeLongest := computeMomentumSummaryFromCounts(
		def,
		counts,
		"2026-08-07",
		"2026-08-09",
		settings,
	)
	if got.Current != rangeCurrent || got.Longest != rangeLongest {
		t.Fatalf(
			"expected snapshot and range calculations to agree, snapshot=%+v range=%d/%d",
			got,
			rangeCurrent,
			rangeLongest,
		)
	}
}

func TestDailyMomentumResetsOnMissAfterProtectedDays(t *testing.T) {
	def := sharedtypes.HabitStreakDefinition{
		ID:            "journal",
		Name:          "Daily Journal",
		Enabled:       true,
		Period:        sharedtypes.HabitStreakPeriodDay,
		RequiredCount: 1,
		TargetKind:    sharedtypes.MomentumTargetKindHabit,
	}
	settings := &sharedtypes.CoreSettings{
		AwayDates: []string{"2026-08-08", "2026-08-09"},
	}
	counts := map[string]map[string]int{
		"2026-08-07": {def.ID: 1},
	}

	state := customHabitMomentumSnapshotState{}
	for _, date := range []string{"2026-08-07", "2026-08-08", "2026-08-09", "2026-08-10"} {
		state = advanceCustomHabitMomentumSnapshotState(state, date, counts, []sharedtypes.HabitStreakDefinition{def}, settings)
	}

	got := state.Definitions[0]
	if got.Current != 0 || got.Longest != 1 {
		t.Fatalf("expected unprotected miss after reset days to reset current streak, got %+v", got)
	}
}

func momentumTestDates(start, end string) []string {
	startTime, err := time.Parse(time.DateOnly, start)
	if err != nil {
		return nil
	}
	endTime, err := time.Parse(time.DateOnly, end)
	if err != nil {
		return nil
	}
	dates := make([]string, 0, int(endTime.Sub(startTime).Hours()/24)+1)
	for date := startTime; !date.After(endTime); date = date.AddDate(0, 0, 1) {
		dates = append(dates, date.Format(time.DateOnly))
	}
	return dates
}
