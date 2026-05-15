package commands

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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
	rangeStart := start + "T00:00:00.000Z"
	rangeEnd := end + "T23:59:59.999Z"
	segments, err := c.SessionSegments.ListEndedInRange(ctx, c.UserID, rangeStart, rangeEnd)
	if err != nil {
		return nil, err
	}
	rangeSessions, err := c.Sessions.ListEnded(ctx, struct {
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
		Since:  &rangeStart,
		Until:  &rangeEnd,
	})
	if err != nil {
		return nil, err
	}
	summariesByDate, err := loadDailyIssueSummariesByDate(ctx, c, start, end, rangeSessions)
	if err != nil {
		return nil, err
	}
	type segmentTotals struct {
		work int
		rest int
	}
	segmentByDate := map[string]segmentTotals{}
	sessionDurationByDate := map[string]int{}
	sessionCountByDate := map[string]int{}
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
	for _, session := range rangeSessions {
		day := extractISODate(session.StartTime)
		if day == "" {
			continue
		}
		sessionDurationByDate[day] += derefIssueEstimate(session.DurationSeconds)
		sessionCountByDate[day]++
	}

	out := make([]sharedtypes.DailyMetricsDay, 0)
	for day := range eachDate(start, end) {
		summary := summariesByDate[day]
		workedSeconds := segmentByDate[day].work
		if workedSeconds == 0 {
			workedSeconds = sessionDurationByDate[day]
		}
		habitDueCount, habitCompletedCount, habitFailedCount, err := loadHabitCountsForDate(ctx, c, day)
		if err != nil {
			return nil, err
		}
		item := sharedtypes.DailyMetricsDay{
			Date:                  day,
			WorkedSeconds:         workedSeconds,
			RestSeconds:           segmentByDate[day].rest,
			SessionCount:          sessionCountByDate[day],
			TotalIssues:           summary.TotalIssues,
			CompletedIssues:       summary.CompletedIssues,
			AbandonedIssues:       summary.AbandonedIssues,
			TotalEstimatedMinutes: summary.TotalEstimatedMinutes,
			HabitDueCount:         habitDueCount,
			HabitCompletedCount:   habitCompletedCount,
			HabitFailedCount:      habitFailedCount,
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
	return ComputeMetricsRollupFromDays(start, end, days), nil
}

func ComputeMetricsRollupFromDays(start string, end string, days []sharedtypes.DailyMetricsDay) *sharedtypes.MetricsRollup {
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
		rollup.HabitDueCount += day.HabitDueCount
		rollup.HabitCompletedCount += day.HabitCompletedCount
		rollup.HabitFailedCount += day.HabitFailedCount
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
	return rollup
}

func ComputeMetricsStreaks(ctx context.Context, c *core.Context, start string, end string) (*sharedtypes.StreakSummary, error) {
	days, err := ComputeMetricsRange(ctx, c, start, end)
	if err != nil {
		return nil, err
	}
	settings, err := c.CoreSettings.Get(ctx, c.UserID)
	if err != nil {
		return nil, err
	}
	streaks := ComputeMetricsStreaksFromDays(days, settings)
	history, err := c.HabitCompletions.ListHistory(ctx, c.UserID, nil, nil)
	if err != nil {
		return nil, err
	}
	streaks.CustomHabitStreaks = ComputeCustomHabitStreaks(history, start, end, settings)
	return streaks, nil
}

func ComputeMetricsLifetimeStreaks(ctx context.Context, c *core.Context, throughDate string) (*sharedtypes.StreakSummary, error) {
	if err := validateCheckInDate(throughDate, "", false); err != nil {
		return nil, err
	}
	start, err := earliestMetricsHistoryDate(ctx, c, throughDate)
	if err != nil {
		return nil, err
	}
	return ComputeMetricsStreaks(ctx, c, start, throughDate)
}

func earliestMetricsHistoryDate(ctx context.Context, c *core.Context, throughDate string) (string, error) {
	start := throughDate
	consider := func(candidate *string) {
		if candidate != nil && *candidate != "" && *candidate < start {
			start = *candidate
		}
	}
	focusDate, err := c.Sessions.EarliestEndedDate(ctx, c.UserID, throughDate)
	if err != nil {
		return "", err
	}
	consider(focusDate)
	checkInDate, err := c.DailyCheckIns.EarliestDate(ctx, c.UserID, throughDate)
	if err != nil {
		return "", err
	}
	consider(checkInDate)
	habitDate, err := c.HabitCompletions.EarliestDate(ctx, c.UserID, throughDate)
	if err != nil {
		return "", err
	}
	consider(habitDate)
	return start, nil
}

func ComputeMetricsStreaksFromDays(days []sharedtypes.DailyMetricsDay, settings *sharedtypes.CoreSettings) *sharedtypes.StreakSummary {
	var streaks sharedtypes.StreakSummary
	currentFocus := 0
	currentCheckIn := 0
	for _, day := range days {
		if day.WorkedSeconds > 0 {
			currentFocus++
			if currentFocus > streaks.LongestFocusDays {
				streaks.LongestFocusDays = currentFocus
			}
		} else if isProtectedStreakDay(day.Date, settings, sharedtypes.StreakKindFocusDays) {
		} else {
			currentFocus = 0
		}
		if countsForCheckInStreak(day.CheckIn) {
			currentCheckIn++
			if currentCheckIn > streaks.LongestCheckInDays {
				streaks.LongestCheckInDays = currentCheckIn
			}
		} else if isProtectedStreakDay(day.Date, settings, sharedtypes.StreakKindCheckInDays) {
		} else {
			currentCheckIn = 0
		}
	}
	streaks.CurrentFocusDays = trailingStreak(days, func(day sharedtypes.DailyMetricsDay) bool { return day.WorkedSeconds > 0 }, func(day sharedtypes.DailyMetricsDay) bool {
		return isProtectedStreakDay(day.Date, settings, sharedtypes.StreakKindFocusDays)
	})
	streaks.CurrentCheckInDays = trailingStreak(days, func(day sharedtypes.DailyMetricsDay) bool { return countsForCheckInStreak(day.CheckIn) }, func(day sharedtypes.DailyMetricsDay) bool {
		return isProtectedStreakDay(day.Date, settings, sharedtypes.StreakKindCheckInDays)
	})
	streaks.CurrentHabitDays, streaks.LongestHabitDays = habitStreakFromDays(days, settings)
	return &streaks
}

func loadHabitCountsForDate(ctx context.Context, c *core.Context, date string) (due, completed, failed int, err error) {
	habits, err := ListHabitsDueForDate(ctx, c, date)
	if err != nil {
		return 0, 0, 0, err
	}
	for _, habit := range habits {
		due++
		switch habit.Status {
		case sharedtypes.HabitCompletionStatusCompleted:
			completed++
		case sharedtypes.HabitCompletionStatusFailed:
			failed++
		}
	}
	return due, completed, failed, nil
}

func habitStreakFromDays(days []sharedtypes.DailyMetricsDay, settings *sharedtypes.CoreSettings) (current int, longest int) {
	for _, day := range days {
		if day.HabitDueCount == 0 {
			continue
		}
		if day.HabitCompletedCount == day.HabitDueCount && day.HabitFailedCount == 0 {
			current++
			if current > longest {
				longest = current
			}
			continue
		}
		if isProtectedStreakDay(day.Date, settings, sharedtypes.StreakKindHabitDays) {
			continue
		}
		current = 0
	}
	return trailingHabitStreak(days, settings), longest
}

func trailingHabitStreak(days []sharedtypes.DailyMetricsDay, settings *sharedtypes.CoreSettings) int {
	streak := 0
	for i := len(days) - 1; i >= 0; i-- {
		day := days[i]
		if day.HabitDueCount == 0 {
			continue
		}
		if day.HabitCompletedCount == day.HabitDueCount && day.HabitFailedCount == 0 {
			streak++
			continue
		}
		if isProtectedStreakDay(day.Date, settings, sharedtypes.StreakKindHabitDays) {
			continue
		}
		break
	}
	return streak
}

func ComputeCustomHabitStreaks(history []sharedtypes.HabitCompletion, start string, end string, settings *sharedtypes.CoreSettings) []sharedtypes.CustomHabitStreakSummary {
	if settings == nil || len(settings.HabitStreakDefs) == 0 {
		return nil
	}
	bucketsByDef := make([]map[string]int, len(settings.HabitStreakDefs))
	for i, def := range settings.HabitStreakDefs {
		bucketsByDef[i] = map[string]int{}
		if !def.Enabled || len(def.HabitIDs) == 0 {
			continue
		}
		habitSet := make(map[int64]struct{}, len(def.HabitIDs))
		for _, id := range def.HabitIDs {
			habitSet[id] = struct{}{}
		}
		for _, entry := range history {
			if entry.Date < start || entry.Date > end {
				continue
			}
			if entry.Status != sharedtypes.HabitCompletionStatusCompleted {
				continue
			}
			if _, ok := habitSet[entry.HabitID]; !ok {
				continue
			}
			bucket := customHabitBucketKey(entry.Date, def.Period)
			if bucket == "" {
				continue
			}
			bucketsByDef[i][bucket]++
		}
	}
	results := make([]sharedtypes.CustomHabitStreakSummary, 0, len(settings.HabitStreakDefs))
	for i, rawDef := range settings.HabitStreakDefs {
		def := sharedtypes.NormalizeHabitStreakDefinition(rawDef)
		summary := sharedtypes.CustomHabitStreakSummary{
			ID:            def.ID,
			Name:          def.Name,
			Enabled:       def.Enabled,
			Period:        def.Period,
			RequiredCount: def.RequiredCount,
		}
		if def.Enabled && len(def.HabitIDs) > 0 {
			buckets := customHabitRangeBuckets(start, end, def.Period)
			current := 0
			for bucketIdx, bucket := range buckets {
				if bucketsByDef[i][bucket] >= def.RequiredCount {
					current++
					if current > summary.Longest {
						summary.Longest = current
					}
				} else {
					if bucketIdx == len(buckets)-1 && customHabitTrailingIncompleteBucketIsOpen(end, def.Period) {
						continue
					}
					current = 0
				}
			}
			summary.Current = current
		}
		results = append(results, summary)
	}
	return results
}

func customHabitTrailingIncompleteBucketIsOpen(end string, period sharedtypes.HabitStreakPeriod) bool {
	parsed, err := time.Parse("2006-01-02", end)
	if err != nil {
		return false
	}
	switch sharedtypes.NormalizeHabitStreakPeriod(period) {
	case sharedtypes.HabitStreakPeriodWeek:
		return parsed.Weekday() != time.Sunday
	case sharedtypes.HabitStreakPeriodMonth:
		return parsed.AddDate(0, 0, 1).Month() == parsed.Month()
	default:
		return false
	}
}

func customHabitBucketKey(date string, period sharedtypes.HabitStreakPeriod) string {
	parsed, err := time.Parse("2006-01-02", date)
	if err != nil {
		return ""
	}
	switch sharedtypes.NormalizeHabitStreakPeriod(period) {
	case sharedtypes.HabitStreakPeriodWeek:
		year, week := parsed.ISOWeek()
		return fmt.Sprintf("%04d-W%02d", year, week)
	case sharedtypes.HabitStreakPeriodMonth:
		return parsed.Format("2006-01")
	default:
		return parsed.Format("2006-01-02")
	}
}

func customHabitRangeBuckets(start string, end string, period sharedtypes.HabitStreakPeriod) []string {
	startTime, err := time.Parse("2006-01-02", start)
	if err != nil {
		return nil
	}
	endTime, err := time.Parse("2006-01-02", end)
	if err != nil || endTime.Before(startTime) {
		return nil
	}
	switch sharedtypes.NormalizeHabitStreakPeriod(period) {
	case sharedtypes.HabitStreakPeriodWeek:
		cursor := startOfISOWeek(startTime)
		last := startOfISOWeek(endTime)
		out := []string{}
		for !cursor.After(last) {
			year, week := cursor.ISOWeek()
			out = append(out, fmt.Sprintf("%04d-W%02d", year, week))
			cursor = cursor.AddDate(0, 0, 7)
		}
		return out
	case sharedtypes.HabitStreakPeriodMonth:
		cursor := time.Date(startTime.Year(), startTime.Month(), 1, 0, 0, 0, 0, startTime.Location())
		last := time.Date(endTime.Year(), endTime.Month(), 1, 0, 0, 0, 0, endTime.Location())
		out := []string{}
		for !cursor.After(last) {
			out = append(out, cursor.Format("2006-01"))
			cursor = cursor.AddDate(0, 1, 0)
		}
		return out
	default:
		out := []string{}
		for cursor := startTime; !cursor.After(endTime); cursor = cursor.AddDate(0, 0, 1) {
			out = append(out, cursor.Format("2006-01-02"))
		}
		return out
	}
}

func startOfISOWeek(value time.Time) time.Time {
	weekday := int(value.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, value.Location()).AddDate(0, 0, -(weekday - 1))
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

func loadDailyIssueSummariesByDate(ctx context.Context, c *core.Context, start string, end string, rangeSessions []sharedtypes.SessionHistoryEntry) (map[string]sharedtypes.DailyIssueSummary, error) {
	issues, err := c.Issues.ListByTodoDateRange(ctx, start, end, c.UserID)
	if err != nil {
		return nil, err
	}
	order := loadListSortSettings(ctx, c).issueSort
	issuesByDate := make(map[string][]sharedtypes.Issue)
	issueIDsByDate := make(map[string]map[int64]struct{})
	totalEstimateByDate := make(map[string]int)
	completedByDate := make(map[string]int)
	abandonedByDate := make(map[string]int)
	for _, issue := range issues {
		if issue.TodoForDate == nil {
			continue
		}
		day := strings.TrimSpace(*issue.TodoForDate)
		if day == "" {
			continue
		}
		issuesByDate[day] = append(issuesByDate[day], issue)
		if issueIDsByDate[day] == nil {
			issueIDsByDate[day] = map[int64]struct{}{}
		}
		issueIDsByDate[day][issue.ID] = struct{}{}
		totalEstimateByDate[day] += derefIssueEstimate(issue.EstimateMinutes)
		if issue.CompletedAt != nil && strings.HasPrefix(*issue.CompletedAt, day) {
			completedByDate[day]++
		}
		if issue.AbandonedAt != nil && strings.HasPrefix(*issue.AbandonedAt, day) {
			abandonedByDate[day]++
		}
	}
	for day := range issuesByDate {
		sortIssues(issuesByDate[day], order)
	}
	workedByDate := make(map[string]int)
	for _, session := range rangeSessions {
		day := extractISODate(session.StartTime)
		if day == "" {
			continue
		}
		issueIDs := issueIDsByDate[day]
		if len(issueIDs) == 0 {
			continue
		}
		if _, ok := issueIDs[session.IssueID]; !ok {
			continue
		}
		workedByDate[day] += derefIssueEstimate(session.DurationSeconds)
	}
	out := make(map[string]sharedtypes.DailyIssueSummary, len(issuesByDate))
	for day, dayIssues := range issuesByDate {
		out[day] = sharedtypes.DailyIssueSummary{
			Date:                  day,
			TotalIssues:           len(dayIssues),
			Issues:                dayIssues,
			TotalEstimatedMinutes: totalEstimateByDate[day],
			CompletedIssues:       completedByDate[day],
			AbandonedIssues:       abandonedByDate[day],
			WorkedSeconds:         workedByDate[day],
		}
	}
	return out, nil
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

func trailingStreak(days []sharedtypes.DailyMetricsDay, predicate func(sharedtypes.DailyMetricsDay) bool, skip func(sharedtypes.DailyMetricsDay) bool) int {
	total := 0
	for i := len(days) - 1; i >= 0; i-- {
		if predicate(days[i]) {
			total++
			continue
		}
		if skip != nil && skip(days[i]) {
			continue
		}
		break
	}
	return total
}

func isProtectedStreakDay(date string, settings *sharedtypes.CoreSettings, kind sharedtypes.StreakKind) bool {
	if settings == nil || !freezesStreakKind(settings, kind) {
		return false
	}
	if settings.AwayModeEnabled {
		return true
	}
	if isRestWeekday(date, settings.RestWeekdays) {
		return true
	}
	if containsString(settings.RestSpecificDates, date) {
		return true
	}
	return false
}

func freezesStreakKind(settings *sharedtypes.CoreSettings, kind sharedtypes.StreakKind) bool {
	selected := settings.FrozenStreakKinds
	if len(selected) == 0 {
		selected = sharedtypes.AvailableStreakKinds()
	}
	for _, current := range selected {
		if current == kind {
			return true
		}
	}
	return false
}

func isRestWeekday(date string, weekdays []int) bool {
	if len(weekdays) == 0 {
		return false
	}
	parsed, err := time.Parse("2006-01-02", date)
	if err != nil {
		return false
	}
	current := int(parsed.Weekday())
	for _, weekday := range weekdays {
		if weekday == current {
			return true
		}
	}
	return false
}

func containsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
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
	stressWeights := map[string]float64{
		"workloadPressure": 0.35,
		"breakDebt":        0.20,
		"moodEnergyDrag":   0.25,
		"sleepDebt":        0.20,
	}
	recoveryWeights := map[string]float64{
		"recoveryConsistency": 0.40,
		"recoveryBreaks":      0.30,
		"loadStability":       0.30,
	}
	const (
		recoveryWeight = 0.75
		baselineOffset = 0.10
		emaAlpha       = 0.45
	)

	avgWorked := averageWorkedSeconds(window)
	workloadPressure := clamp01(avgWorked / 25200.0)
	breakDebt, breakRecovery := breakDebtAndRecovery(window)
	moodEnergyDrag, recoveryConsistency, moodCount, energyCount := moodEnergySignals(window)
	sleepDebt, sleepRecovery, sleepCount := sleepDebtAndRecovery(window)
	loadStability := workloadStability(window)
	if moodCount == 0 && energyCount == 0 {
		delete(stressWeights, "moodEnergyDrag")
		delete(recoveryWeights, "recoveryConsistency")
	}
	if sleepCount == 0 {
		delete(stressWeights, "sleepDebt")
		delete(recoveryWeights, "recoveryConsistency")
	}

	stressSignals := map[string]float64{
		"workloadPressure": workloadPressure,
		"breakDebt":        breakDebt,
		"moodEnergyDrag":   moodEnergyDrag,
		"sleepDebt":        sleepDebt,
	}
	recoverySignals := map[string]float64{
		"recoveryConsistency": clamp01((recoveryConsistency + sleepRecovery) / 2.0),
		"recoveryBreaks":      breakRecovery,
		"loadStability":       loadStability,
	}

	stressScore := weightedAverage(stressSignals, stressWeights)
	recoveryScore := weightedAverage(recoverySignals, recoveryWeights)
	raw := clamp01(stressScore - (recoveryScore * recoveryWeight) + baselineOffset)
	if idx > 0 && days[idx-1].Burnout != nil {
		prev := clamp01(float64(days[idx-1].Burnout.Score) / 100.0)
		raw = (emaAlpha * raw) + ((1.0 - emaAlpha) * prev)
	}

	factors := map[string]float64{
		"workloadPressure":    stressWeights["workloadPressure"] * workloadPressure,
		"breakDebt":           stressWeights["breakDebt"] * breakDebt,
		"moodEnergyDrag":      stressWeights["moodEnergyDrag"] * moodEnergyDrag,
		"sleepDebt":           stressWeights["sleepDebt"] * sleepDebt,
		"recoveryConsistency": -recoveryWeight * recoveryWeights["recoveryConsistency"] * recoverySignals["recoveryConsistency"],
		"recoveryBreaks":      -recoveryWeight * recoveryWeights["recoveryBreaks"] * breakRecovery,
		"loadStability":       -recoveryWeight * recoveryWeights["loadStability"] * loadStability,
	}
	for key, value := range factors {
		if value == 0 {
			delete(factors, key)
		}
	}

	finalScore := int(math.Round(raw * 100))
	level := sharedtypes.BurnoutLevelLow
	if finalScore >= 70 {
		level = sharedtypes.BurnoutLevelHigh
	} else if finalScore >= 40 {
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

func sleepDebtAndRecovery(days []sharedtypes.DailyMetricsDay) (debt float64, recovery float64, count int) {
	var total float64
	var recoveryTotal float64
	for _, day := range days {
		if day.CheckIn == nil {
			continue
		}
		if day.CheckIn.SleepHours != nil {
			hours := *day.CheckIn.SleepHours
			total += clamp01((7.5 - hours) / 3.5)
			recoveryTotal += clamp01((hours - 6.0) / 2.5)
			count++
			continue
		}
		if day.CheckIn.SleepScore != nil {
			score := float64(*day.CheckIn.SleepScore)
			total += clamp01((75.0 - score) / 75.0)
			recoveryTotal += clamp01((score - 55.0) / 45.0)
			count++
		}
	}
	if count == 0 {
		return 0, 0, 0
	}
	return total / float64(count), recoveryTotal / float64(count), count
}

func averageWorkedSeconds(days []sharedtypes.DailyMetricsDay) float64 {
	if len(days) == 0 {
		return 0
	}
	total := 0
	for _, day := range days {
		total += day.WorkedSeconds
	}
	return float64(total) / float64(len(days))
}

func breakDebtAndRecovery(days []sharedtypes.DailyMetricsDay) (debt float64, recovery float64) {
	var worked, rest int
	for _, day := range days {
		worked += day.WorkedSeconds
		rest += day.RestSeconds
	}
	if worked <= 0 {
		return 0, 0
	}
	expectedRestRatio := 0.20
	actualRatio := float64(rest) / float64(worked)
	debt = clamp01((expectedRestRatio - actualRatio) / expectedRestRatio)
	recovery = clamp01(actualRatio / expectedRestRatio)
	return debt, recovery
}

func moodEnergySignals(days []sharedtypes.DailyMetricsDay) (drag float64, recovery float64, moodCount int, energyCount int) {
	moodAvg, moodTrend, moodValues := averageAndTrend(days, func(day sharedtypes.DailyMetricsDay) *int {
		if day.CheckIn == nil {
			return nil
		}
		return &day.CheckIn.Mood
	})
	energyAvg, energyTrend, energyValues := averageAndTrend(days, func(day sharedtypes.DailyMetricsDay) *int {
		if day.CheckIn == nil {
			return nil
		}
		return &day.CheckIn.Energy
	})
	moodCount = moodValues
	energyCount = energyValues
	if moodCount == 0 && energyCount == 0 {
		return 0, 0, 0, 0
	}
	baseDrag := 0.0
	trendDrag := 0.0
	baseRecovery := 0.0
	if moodCount > 0 {
		baseDrag += clamp01((5.0 - moodAvg) / 4.0)
		trendDrag += moodTrend
		baseRecovery += clamp01((moodAvg - 2.5) / 2.5)
	}
	if energyCount > 0 {
		baseDrag += clamp01((5.0 - energyAvg) / 4.0)
		trendDrag += energyTrend
		baseRecovery += clamp01((energyAvg - 2.5) / 2.5)
	}
	scale := float64(maxInt(1, moodCountOnly(moodCount, energyCount)))
	baseDrag /= scale
	trendDrag /= scale
	baseRecovery /= scale
	drag = clamp01(baseDrag*0.75 + trendDrag*0.25)
	recovery = clamp01(baseRecovery)
	return drag, recovery, moodCount, energyCount
}

func moodCountOnly(moodCount, energyCount int) int {
	count := 0
	if moodCount > 0 {
		count++
	}
	if energyCount > 0 {
		count++
	}
	return count
}

func workloadStability(days []sharedtypes.DailyMetricsDay) float64 {
	if len(days) < 2 {
		return 0.5
	}
	values := make([]float64, 0, len(days))
	total := 0.0
	for _, day := range days {
		value := float64(day.WorkedSeconds)
		values = append(values, value)
		total += value
	}
	mean := total / float64(len(values))
	if mean == 0 {
		return 0.6
	}
	variance := 0.0
	for _, value := range values {
		diff := value - mean
		variance += diff * diff
	}
	stddev := math.Sqrt(variance / float64(len(values)))
	normalized := clamp01(stddev / 18000.0)
	return 1.0 - normalized
}

func weightedAverage(values map[string]float64, weights map[string]float64) float64 {
	totalWeight := 0.0
	total := 0.0
	for key, weight := range weights {
		value, ok := values[key]
		if !ok {
			continue
		}
		totalWeight += weight
		total += value * weight
	}
	if totalWeight == 0 {
		return 0
	}
	return clamp01(total / totalWeight)
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
