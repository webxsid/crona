//go:build e2e

package e2e

import (
	"testing"
	"time"

	shareddto "crona/shared/dto"
	"crona/shared/protocol"
	sharedtypes "crona/shared/types"
)

func TestDailyCheckInAndMetricsOverIPC(t *testing.T) {
	kernel := startTestKernel(t)
	defer kernel.close(t)

	repo := createRepo(t, kernel, "Work")
	stream := createStream(t, kernel, repo.ID, "app")
	issue := createIssue(t, kernel, stream.ID, "Ship phase 2", nil)
	issue = changeIssueStatus(t, kernel, issue.ID, sharedtypes.IssueStatusPlanned)

	kernel.call(t, protocol.MethodTimerStart, shareddto.TimerStartRequest{IssueID: &issue.ID}, nil)
	time.Sleep(1200 * time.Millisecond)
	kernel.call(t, protocol.MethodTimerEnd, shareddto.EndSessionRequest{}, nil)

	today := time.Now().Format("2006-01-02")
	sleepHours := 6.5
	sleepScore := 74
	screenTime := 180
	notes := "Long day, but steady."

	var checkIn sharedtypes.DailyCheckIn
	kernel.call(t, protocol.MethodCheckInUpsert, shareddto.DailyCheckInUpsertRequest{
		Date:              today,
		Mood:              3,
		Energy:            2,
		SleepHours:        &sleepHours,
		SleepScore:        &sleepScore,
		ScreenTimeMinutes: &screenTime,
		Notes:             &notes,
	}, &checkIn)

	if checkIn.Date != today || checkIn.Mood != 3 || checkIn.Energy != 2 {
		t.Fatalf("unexpected check-in: %#v", checkIn)
	}

	kernel.call(t, protocol.MethodCheckInGet, shareddto.DailyCheckInQuery{Date: today}, &checkIn)
	if checkIn.Date != today {
		t.Fatalf("expected check-in for %s", today)
	}

	var rangeOut []sharedtypes.DailyCheckIn
	kernel.call(t, protocol.MethodCheckInRange, shareddto.DateRangeQuery{Start: today, End: today}, &rangeOut)
	if len(rangeOut) != 1 {
		t.Fatalf("expected one check-in in range, got %d", len(rangeOut))
	}

	var metrics []sharedtypes.DailyMetricsDay
	kernel.call(t, protocol.MethodMetricsRange, shareddto.DateRangeQuery{Start: today, End: today}, &metrics)
	if len(metrics) != 1 {
		t.Fatalf("expected one metrics day, got %d", len(metrics))
	}
	if metrics[0].CheckIn == nil || metrics[0].Burnout == nil {
		t.Fatalf("expected metrics day to include check-in and burnout: %#v", metrics[0])
	}

	var rollup sharedtypes.MetricsRollup
	kernel.call(t, protocol.MethodMetricsRollup, shareddto.DateRangeQuery{Start: today, End: today}, &rollup)
	if rollup.CheckInDays != 1 {
		t.Fatalf("expected rollup check-in days to be 1, got %d", rollup.CheckInDays)
	}

	var streaks sharedtypes.StreakSummary
	kernel.call(t, protocol.MethodMetricsStreaks, shareddto.DateRangeQuery{Start: today, End: today}, &streaks)
	if streaks.CurrentCheckInDays != 1 {
		t.Fatalf("expected current check-in streak of 1, got %d", streaks.CurrentCheckInDays)
	}

	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	kernel.call(t, protocol.MethodCheckInUpsert, shareddto.DailyCheckInUpsertRequest{
		Date:   yesterday,
		Mood:   4,
		Energy: 4,
	}, nil)

	kernel.call(t, protocol.MethodMetricsStreaks, shareddto.DateRangeQuery{Start: yesterday, End: today}, &streaks)
	if streaks.LongestCheckInDays != 1 {
		t.Fatalf("expected backfilled check-in to be excluded from streaks, got %d", streaks.LongestCheckInDays)
	}

	kernel.call(t, protocol.MethodCheckInDelete, shareddto.DeleteByDateRequest{Date: today}, nil)

	var deletedCheckIn sharedtypes.DailyCheckIn
	kernel.call(t, protocol.MethodCheckInGet, shareddto.DailyCheckInQuery{Date: today}, &deletedCheckIn)
	if deletedCheckIn.Date != "" {
		t.Fatalf("expected deleted check-in to return empty result, got %#v", deletedCheckIn)
	}
}
