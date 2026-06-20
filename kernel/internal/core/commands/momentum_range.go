package commands

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"crona/kernel/internal/core"
	sharedtypes "crona/shared/types"
)

func ListMomentumCards(
	ctx context.Context,
	c *core.Context,
	endDate string,
	windowDays int,
) ([]sharedtypes.MomentumCard, error) {
	if endDate == "" {
		endDate = extractISODate(c.Now())
	}
	if !isISODate(endDate) {
		return nil, errors.New("end date must be YYYY-MM-DD")
	}
	if windowDays < 1 {
		return nil, errors.New("window days must be positive")
	}

	defs, err := c.HabitStreakDefinitions.List(ctx, c.UserID)
	if err != nil {
		return nil, err
	}
	if len(defs) == 0 {
		return nil, nil
	}
	defs = sharedtypes.NormalizeHabitStreakDefinitions(defs)

	countsByDate, err := loadCustomHabitMomentumCountsByDate(ctx, c, defs, endDate)
	if err != nil {
		return nil, err
	}

	snapshot, err := ensureCustomHabitMomentumSnapshot(ctx, c, endDate, defs, countsByDate)
	if err != nil {
		return nil, err
	}
	summaries := snapshot.summaries
	summaryByID := make(map[string]sharedtypes.CustomHabitStreakSummary, len(summaries))
	for _, summary := range summaries {
		summaryByID[summary.ID] = summary
	}

	habits, err := c.Habits.ListAllWithMeta(ctx, c.UserID)
	if err != nil {
		return nil, err
	}
	habitNamesByID := make(map[int64]string, len(habits))
	for _, habit := range habits {
		habitNamesByID[habit.ID] = habit.Name
	}
	repos, err := c.Repos.List(ctx, c.UserID)
	if err != nil {
		return nil, err
	}
	repoNamesByID := make(map[int64]string, len(repos))
	for _, repo := range repos {
		repoNamesByID[repo.ID] = repo.Name
	}
	streamNamesByID := make(map[int64]string, len(habits))
	for _, repo := range repos {
		streams, err := c.Streams.ListByRepo(ctx, repo.ID, c.UserID)
		if err != nil {
			return nil, err
		}
		for _, stream := range streams {
			streamNamesByID[stream.ID] = stream.Name
		}
	}
	startDate := shiftMomentumISODate(endDate, -(windowDays - 1))

	out := make([]sharedtypes.MomentumCard, 0, len(defs))
	for _, rawDef := range defs {
		def := sharedtypes.NormalizeHabitStreakDefinition(rawDef)
		summary := summaryByID[def.ID]
		card := sharedtypes.MomentumCard{
			Definition:  def,
			Current:     summary.Current,
			Longest:     summary.Longest,
			HabitNames:  momentumHabitNames(def.HabitIDs, habitNamesByID),
			TargetNames: momentumTargetNames(def, habitNamesByID, repoNamesByID, streamNamesByID),
			Series:      buildMomentumSeries(def, startDate, endDate, countsByDate),
		}
		out = append(out, card)
	}
	return out, nil
}

func buildMomentumSeries(
	def sharedtypes.HabitStreakDefinition,
	startDate string,
	endDate string,
	countsByDate map[string]map[string]int,
) []sharedtypes.MomentumSeriesPoint {
	buckets := customHabitRangeBuckets(startDate, endDate, def.Period)
	if len(buckets) == 0 {
		return nil
	}
	countsByBucket := make(map[string]int, len(buckets))
	for day := startDate; day <= endDate; day = nextISODate(day) {
		dayCounts := countsByDate[day]
		if dayCounts == nil {
			continue
		}
		key := customHabitBucketKey(day, def.Period)
		countsByBucket[key] += dayCounts[def.ID]
	}
	series := make([]sharedtypes.MomentumSeriesPoint, 0, len(buckets))
	for _, key := range buckets {
		bucketStart, bucketEnd := momentumBucketBounds(key, def.Period)
		if bucketStart < startDate {
			bucketStart = startDate
		}
		if bucketEnd > endDate {
			bucketEnd = endDate
		}
		count := countsByBucket[key]
		series = append(series, sharedtypes.MomentumSeriesPoint{
			BucketKey: key,
			Label:     momentumBucketLabel(key, def.Period, bucketStart, bucketEnd),
			StartDate: bucketStart,
			EndDate:   bucketEnd,
			Count:     count,
			Target:    momentumRequiredUnits(def),
			MetTarget: count >= momentumRequiredUnits(def),
		})
	}
	return series
}

func momentumHabitNames(ids []int64, namesByID map[int64]string) []string {
	if len(ids) == 0 {
		return nil
	}
	names := make([]string, 0, len(ids))
	for _, id := range ids {
		name := namesByID[id]
		if name == "" {
			continue
		}
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

func momentumTargetNames(
	def sharedtypes.HabitStreakDefinition,
	habitNamesByID map[int64]string,
	repoNamesByID map[int64]string,
	streamNamesByID map[int64]string,
) []string {
	switch sharedtypes.NormalizeMomentumTargetKind(def.TargetKind) {
	case sharedtypes.MomentumTargetKindContext:
		if len(def.Contexts) == 0 {
			return nil
		}
		names := make([]string, 0, len(def.Contexts))
		for _, contextItem := range def.Contexts {
			repoName := repoNamesByID[contextItem.RepoID]
			streamName := ""
			if contextItem.StreamID != nil {
				streamName = streamNamesByID[*contextItem.StreamID]
			}
			if repoName == "" && streamName == "" {
				continue
			}
			switch {
			case repoName != "" && streamName != "":
				names = append(names, repoName+"/"+streamName)
			case repoName != "":
				names = append(names, repoName)
			default:
				names = append(names, streamName)
			}
		}
		if len(names) > 0 {
			slices.Sort(names)
			return names
		}
	default:
		return momentumHabitNames(def.HabitIDs, habitNamesByID)
	}
	return nil
}

func momentumBucketBounds(key string, period sharedtypes.HabitStreakPeriod) (string, string) {
	switch sharedtypes.NormalizeHabitStreakPeriod(period) {
	case sharedtypes.HabitStreakPeriodWeek:
		var year, week int
		if _, err := fmt.Sscanf(key, "%4d-W%2d", &year, &week); err != nil {
			return "", ""
		}
		start := isoWeekStart(year, week)
		end := start.AddDate(0, 0, 6)
		return start.Format("2006-01-02"), end.Format("2006-01-02")
	case sharedtypes.HabitStreakPeriodMonth:
		start, err := time.Parse("2006-01", key)
		if err != nil {
			return "", ""
		}
		end := start.AddDate(0, 1, -1)
		return start.Format("2006-01-02"), end.Format("2006-01-02")
	default:
		return key, key
	}
}

func momentumBucketLabel(
	key string,
	period sharedtypes.HabitStreakPeriod,
	startDate string,
	endDate string,
) string {
	switch sharedtypes.NormalizeHabitStreakPeriod(period) {
	case sharedtypes.HabitStreakPeriodWeek:
		start, err := time.Parse("2006-01-02", startDate)
		if err != nil {
			return key
		}
		end, err := time.Parse("2006-01-02", endDate)
		if err != nil {
			return key
		}
		weekLabel := key
		if _, week, err := parseISOWeekKey(key); err == nil {
			weekLabel = fmt.Sprintf("[%d]", week)
		}
		if start.Month() == end.Month() {
			return fmt.Sprintf("%s %s-%d", weekLabel, start.Format("Jan 2"), end.Day())
		}
		return fmt.Sprintf("%s %s-%s", weekLabel, start.Format("Jan 2"), end.Format("Jan 2"))
	case sharedtypes.HabitStreakPeriodMonth:
		start, err := time.Parse("2006-01-02", startDate)
		if err != nil {
			return key
		}
		return start.Format("Jan 2006")
	default:
		start, err := time.Parse("2006-01-02", startDate)
		if err != nil {
			return key
		}
		return start.Format("Jan 2")
	}
}

func parseISOWeekKey(key string) (int, int, error) {
	var year, week int
	if _, err := fmt.Sscanf(key, "%4d-W%2d", &year, &week); err != nil {
		return 0, 0, err
	}
	return year, week, nil
}

func isoWeekStart(year, week int) time.Time {
	jan4 := time.Date(year, time.January, 4, 0, 0, 0, 0, time.UTC)
	start := startOfISOWeek(jan4)
	return start.AddDate(0, 0, (week-1)*7)
}

func shiftMomentumISODate(date string, days int) string {
	parsed, err := time.Parse("2006-01-02", date)
	if err != nil {
		return date
	}
	return parsed.AddDate(0, 0, days).Format("2006-01-02")
}
