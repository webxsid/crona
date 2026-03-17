package commands

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"strings"
	"time"

	"crona/kernel/internal/core"

	"github.com/google/uuid"

	shareddto "crona/shared/dto"
	sharedtypes "crona/shared/types"
)

func GetDailyCheckIn(ctx context.Context, c *core.Context, date string) (*sharedtypes.DailyCheckIn, error) {
	if err := validateCheckInDate(date, c.Now(), false); err != nil {
		return nil, err
	}
	return c.DailyCheckIns.GetByDate(ctx, c.UserID, date)
}

func UpsertDailyCheckIn(ctx context.Context, c *core.Context, input shareddto.DailyCheckInUpsertRequest) (*sharedtypes.DailyCheckIn, error) {
	if err := validateCheckInDate(input.Date, c.Now(), true); err != nil {
		return nil, err
	}
	if input.Mood < 1 || input.Mood > 5 {
		return nil, errors.New("mood must be between 1 and 5")
	}
	if input.Energy < 1 || input.Energy > 5 {
		return nil, errors.New("energy must be between 1 and 5")
	}
	if input.SleepHours != nil && *input.SleepHours < 0 {
		return nil, errors.New("sleepHours must be >= 0")
	}
	if input.SleepScore != nil && (*input.SleepScore < 0 || *input.SleepScore > 100) {
		return nil, errors.New("sleepScore must be between 0 and 100")
	}
	if input.ScreenTimeMinutes != nil && *input.ScreenTimeMinutes < 0 {
		return nil, errors.New("screenTimeMinutes must be >= 0")
	}

	now := c.Now()
	updated, err := c.DailyCheckIns.Upsert(ctx, sharedtypes.DailyCheckIn{
		Date:              input.Date,
		Mood:              input.Mood,
		Energy:            input.Energy,
		SleepHours:        input.SleepHours,
		SleepScore:        input.SleepScore,
		ScreenTimeMinutes: input.ScreenTimeMinutes,
		Notes:             normalizeOptionalString(input.Notes),
	}, c.UserID, c.DeviceID, now)
	if err != nil {
		return nil, err
	}
	if err := c.Ops.Append(ctx, sharedtypes.Op{
		ID:        uuid.NewString(),
		Entity:    sharedtypes.OpEntityCheckIn,
		EntityID:  input.Date,
		Action:    sharedtypes.OpActionUpdate,
		Payload:   updated,
		Timestamp: now,
		UserID:    c.UserID,
		DeviceID:  c.DeviceID,
	}); err != nil {
		return nil, err
	}
	emit(c, sharedtypes.EventTypeCheckInUpdated, updated)
	return updated, nil
}

func DeleteDailyCheckIn(ctx context.Context, c *core.Context, date string) error {
	if err := validateCheckInDate(date, c.Now(), false); err != nil {
		return err
	}
	if err := c.DailyCheckIns.DeleteByDate(ctx, c.UserID, date); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("check-in not found")
		}
		return err
	}
	now := c.Now()
	if err := c.Ops.Append(ctx, sharedtypes.Op{
		ID:        uuid.NewString(),
		Entity:    sharedtypes.OpEntityCheckIn,
		EntityID:  date,
		Action:    sharedtypes.OpActionDelete,
		Payload:   map[string]string{"date": date},
		Timestamp: now,
		UserID:    c.UserID,
		DeviceID:  c.DeviceID,
	}); err != nil {
		return err
	}
	emit(c, sharedtypes.EventTypeCheckInDeleted, map[string]string{"date": date})
	return nil
}

func ListDailyCheckInsInRange(ctx context.Context, c *core.Context, start string, end string) ([]sharedtypes.DailyCheckIn, error) {
	if err := validateRange(start, end); err != nil {
		return nil, err
	}
	return c.DailyCheckIns.ListRange(ctx, c.UserID, start, end)
}

func ComputeMetricsRange(ctx context.Context, c *core.Context, start string, end string) ([]sharedtypes.DailyMetricsDay, error) {
	if err := validateRange(start, end); err != nil {
		return nil, err
	}
	checkIns, err := c.DailyCheckIns.ListRange(ctx, c.UserID, start, end)
	if err != nil {
		return nil, err
	}
	checkInByDate := map[string]sharedtypes.DailyCheckIn{}
	for _, checkIn := range checkIns {
		checkInByDate[checkIn.Date] = checkIn
	}
	segments, err := c.SessionSegments.ListEndedInRange(ctx, c.UserID, start+"T00:00:00.000Z", end+"T23:59:59.999Z")
	if err != nil {
		return nil, err
	}
	type segmentTotals struct {
		work int
		rest int
	}
	segmentByDate := map[string]segmentTotals{}
	for _, segment := range segments {
		day := extractISODate(segment.StartTime)
		if day == "" {
			continue
		}
		duration := segmentDurationSeconds(segment)
		totals := segmentByDate[day]
		switch segment.SegmentType {
		case sharedtypes.SessionSegmentWork:
			totals.work += duration
		default:
			totals.rest += duration
		}
		segmentByDate[day] = totals
	}

	out := make([]sharedtypes.DailyMetricsDay, 0)
	for day := range eachDate(start, end) {
		summary, err := ComputeDailyIssueSummaryForDate(ctx, c, day)
		if err != nil {
			return nil, err
		}
		dayStart := day + "T00:00:00.000Z"
		dayEnd := day + "T23:59:59.999Z"
		sessions, err := c.Sessions.ListEnded(ctx, struct {
			UserID   string
			RepoID   *int64
			StreamID *int64
			IssueID  *int64
			Since    *string
			Until    *string
			Limit    *int
			Offset   *int
		}{
			UserID: c.UserID,
			Since:  &dayStart,
			Until:  &dayEnd,
		})
		if err != nil {
			return nil, err
		}
		item := sharedtypes.DailyMetricsDay{
			Date:                  day,
			WorkedSeconds:         summary.WorkedSeconds,
			RestSeconds:           segmentByDate[day].rest,
			SessionCount:          len(sessions),
			TotalIssues:           summary.TotalIssues,
			CompletedIssues:       summary.CompletedIssues,
			AbandonedIssues:       summary.AbandonedIssues,
			TotalEstimatedMinutes: summary.TotalEstimatedMinutes,
		}
		if checkIn, ok := checkInByDate[day]; ok {
			copy := checkIn
			item.CheckIn = &copy
		}
		out = append(out, item)
	}
	for i := range out {
		burnout := computeBurnout(out, i)
		out[i].Burnout = &burnout
	}
	return out, nil
}

func ComputeMetricsRollup(ctx context.Context, c *core.Context, start string, end string) (*sharedtypes.MetricsRollup, error) {
	days, err := ComputeMetricsRange(ctx, c, start, end)
	if err != nil {
		return nil, err
	}
	rollup := &sharedtypes.MetricsRollup{
		StartDate: start,
		EndDate:   end,
		Days:      len(days),
	}
	var moodSum, energySum, sleepHoursSum, sleepScoreSum, screenTimeSum float64
	var moodCount, energyCount, sleepHoursCount, sleepScoreCount, screenTimeCount int
	for _, day := range days {
		rollup.WorkedSeconds += day.WorkedSeconds
		rollup.RestSeconds += day.RestSeconds
		rollup.SessionCount += day.SessionCount
		rollup.CompletedIssues += day.CompletedIssues
		rollup.AbandonedIssues += day.AbandonedIssues
		rollup.TotalEstimatedMinutes += day.TotalEstimatedMinutes
		if day.WorkedSeconds > 0 {
			rollup.FocusDays++
		}
		if day.CheckIn != nil {
			rollup.CheckInDays++
			moodSum += float64(day.CheckIn.Mood)
			energySum += float64(day.CheckIn.Energy)
			moodCount++
			energyCount++
			if day.CheckIn.SleepHours != nil {
				sleepHoursSum += *day.CheckIn.SleepHours
				sleepHoursCount++
			}
			if day.CheckIn.SleepScore != nil {
				sleepScoreSum += float64(*day.CheckIn.SleepScore)
				sleepScoreCount++
			}
			if day.CheckIn.ScreenTimeMinutes != nil {
				screenTimeSum += float64(*day.CheckIn.ScreenTimeMinutes)
				screenTimeCount++
			}
		}
		if day.Burnout != nil {
			copy := *day.Burnout
			rollup.LatestBurnout = &copy
		}
	}
	rollup.AverageMood = averageOrNil(moodSum, moodCount)
	rollup.AverageEnergy = averageOrNil(energySum, energyCount)
	rollup.AverageSleepHours = averageOrNil(sleepHoursSum, sleepHoursCount)
	rollup.AverageSleepScore = averageOrNil(sleepScoreSum, sleepScoreCount)
	rollup.AverageScreenTimeMins = averageOrNil(screenTimeSum, screenTimeCount)
	return rollup, nil
}

func ComputeMetricsStreaks(ctx context.Context, c *core.Context, start string, end string) (*sharedtypes.StreakSummary, error) {
	days, err := ComputeMetricsRange(ctx, c, start, end)
	if err != nil {
		return nil, err
	}
	var streaks sharedtypes.StreakSummary
	currentFocus := 0
	currentCheckIn := 0
	for _, day := range days {
		if day.WorkedSeconds > 0 {
			currentFocus++
			if currentFocus > streaks.LongestFocusDays {
				streaks.LongestFocusDays = currentFocus
			}
		} else {
			currentFocus = 0
		}
		if countsForCheckInStreak(day.CheckIn) {
			currentCheckIn++
			if currentCheckIn > streaks.LongestCheckInDays {
				streaks.LongestCheckInDays = currentCheckIn
			}
		} else {
			currentCheckIn = 0
		}
	}
	streaks.CurrentFocusDays = trailingStreak(days, func(day sharedtypes.DailyMetricsDay) bool { return day.WorkedSeconds > 0 })
	streaks.CurrentCheckInDays = trailingStreak(days, func(day sharedtypes.DailyMetricsDay) bool { return countsForCheckInStreak(day.CheckIn) })
	return &streaks, nil
}

func validateCheckInDate(date string, now string, rejectFuture bool) error {
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return errors.New("invalid date")
	}
	if rejectFuture {
		today := extractISODate(now)
		if today != "" && date > today {
			return errors.New("future dates are not allowed")
		}
	}
	return nil
}

func validateRange(start string, end string) error {
	if _, err := time.Parse("2006-01-02", start); err != nil {
		return errors.New("invalid start date")
	}
	if _, err := time.Parse("2006-01-02", end); err != nil {
		return errors.New("invalid end date")
	}
	if start > end {
		return errors.New("start must be <= end")
	}
	return nil
}

func normalizeOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func extractISODate(value string) string {
	if len(value) >= 10 {
		return value[:10]
	}
	return ""
}

func eachDate(start string, end string) func(func(string) bool) {
	return func(yield func(string) bool) {
		current, _ := time.Parse("2006-01-02", start)
		finish, _ := time.Parse("2006-01-02", end)
		for !current.After(finish) {
			if !yield(current.Format("2006-01-02")) {
				return
			}
			current = current.AddDate(0, 0, 1)
		}
	}
}

func segmentDurationSeconds(segment sharedtypes.SessionSegment) int {
	if segment.EndTime == nil {
		return 0
	}
	start, err := time.Parse(time.RFC3339, segment.StartTime)
	if err != nil {
		return 0
	}
	end, err := time.Parse(time.RFC3339, *segment.EndTime)
	if err != nil {
		return 0
	}
	seconds := int(end.Sub(start).Seconds())
	if segment.ElapsedOffsetSeconds != nil {
		seconds += *segment.ElapsedOffsetSeconds
	}
	if seconds < 0 {
		return 0
	}
	return seconds
}

func averageOrNil(sum float64, count int) *float64 {
	if count == 0 {
		return nil
	}
	value := sum / float64(count)
	return &value
}

func trailingStreak(days []sharedtypes.DailyMetricsDay, predicate func(sharedtypes.DailyMetricsDay) bool) int {
	total := 0
	for i := len(days) - 1; i >= 0; i-- {
		if !predicate(days[i]) {
			break
		}
		total++
	}
	return total
}

func countsForCheckInStreak(checkIn *sharedtypes.DailyCheckIn) bool {
	if checkIn == nil {
		return false
	}
	return extractISODate(checkIn.CreatedAt) == checkIn.Date
}

func computeBurnout(days []sharedtypes.DailyMetricsDay, idx int) sharedtypes.BurnoutIndicator {
	start := idx - 6
	if start < 0 {
		start = 0
	}
	window := days[start : idx+1]
	factors := map[string]float64{}
	weights := map[string]float64{
		"sessionDensity":  0.30,
		"breakCompliance": 0.20,
		"moodTrend":       0.20,
		"energyTrend":     0.20,
		"sleepRisk":       0.10,
	}

	var worked int
	for _, day := range window {
		worked += day.WorkedSeconds
	}
	avgWorked := float64(worked) / float64(maxInt(1, len(window)))
	factors["sessionDensity"] = clamp01(avgWorked / 21600.0)

	var rest int
	for _, day := range window {
		rest += day.RestSeconds
	}
	if worked > 0 {
		expectedRestRatio := 0.20
		actualRatio := float64(rest) / float64(worked)
		factors["breakCompliance"] = clamp01((expectedRestRatio - actualRatio) / expectedRestRatio)
	} else {
		factors["breakCompliance"] = 0.0
	}

	moodAvg, moodTrend, moodCount := averageAndTrend(window, func(day sharedtypes.DailyMetricsDay) *int {
		if day.CheckIn == nil {
			return nil
		}
		return &day.CheckIn.Mood
	})
	if moodCount > 0 {
		factors["moodTrend"] = clamp01(((5.0 - moodAvg) / 4.0 * 0.7) + moodTrend*0.3)
	} else {
		delete(weights, "moodTrend")
	}

	energyAvg, energyTrend, energyCount := averageAndTrend(window, func(day sharedtypes.DailyMetricsDay) *int {
		if day.CheckIn == nil {
			return nil
		}
		return &day.CheckIn.Energy
	})
	if energyCount > 0 {
		factors["energyTrend"] = clamp01(((5.0 - energyAvg) / 4.0 * 0.7) + energyTrend*0.3)
	} else {
		delete(weights, "energyTrend")
	}

	sleepRisk, ok := sleepRisk(window)
	if ok {
		factors["sleepRisk"] = sleepRisk
	} else {
		delete(weights, "sleepRisk")
	}

	totalWeight := 0.0
	for _, weight := range weights {
		totalWeight += weight
	}
	score := 0.0
	for key, weight := range weights {
		score += factors[key] * (weight / totalWeight)
	}
	finalScore := int(math.Round(score * 100))
	level := sharedtypes.BurnoutLevelLow
	if finalScore >= 65 {
		level = sharedtypes.BurnoutLevelHigh
	} else if finalScore >= 35 {
		level = sharedtypes.BurnoutLevelGuarded
	}
	return sharedtypes.BurnoutIndicator{
		Score:   finalScore,
		Level:   level,
		Factors: factors,
	}
}

func averageAndTrend(days []sharedtypes.DailyMetricsDay, pick func(sharedtypes.DailyMetricsDay) *int) (avg float64, trend float64, count int) {
	values := make([]float64, 0, len(days))
	for _, day := range days {
		value := pick(day)
		if value == nil {
			continue
		}
		values = append(values, float64(*value))
		avg += float64(*value)
		count++
	}
	if count == 0 {
		return 0, 0, 0
	}
	avg /= float64(count)
	if len(values) >= 2 {
		change := values[0] - values[len(values)-1]
		trend = clamp01(change / 4.0)
	}
	return avg, trend, count
}

func sleepRisk(days []sharedtypes.DailyMetricsDay) (float64, bool) {
	var total float64
	count := 0
	for _, day := range days {
		if day.CheckIn == nil {
			continue
		}
		if day.CheckIn.SleepHours != nil {
			total += clamp01((8.0 - *day.CheckIn.SleepHours) / 4.0)
			count++
			continue
		}
		if day.CheckIn.SleepScore != nil {
			total += clamp01(float64(100-*day.CheckIn.SleepScore) / 100.0)
			count++
		}
	}
	if count == 0 {
		return 0, false
	}
	return total / float64(count), true
}

func clamp01(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
