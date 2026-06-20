package commands

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"crona/kernel/internal/core"
	sharedtypes "crona/shared/types"
)

func GetMomentumDetail(
	ctx context.Context,
	c *core.Context,
	id string,
	endDate string,
	windowDays int,
) (*sharedtypes.MomentumDetail, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.New("momentum id is required")
	}
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
	var def sharedtypes.HabitStreakDefinition
	found := false
	for _, item := range defs {
		item = sharedtypes.NormalizeHabitStreakDefinition(item)
		if item.ID == id {
			def = item
			found = true
			break
		}
	}
	if !found {
		return nil, errors.New("momentum not found")
	}

	countsByDate, err := loadCustomHabitMomentumCountsByDate(ctx, c, defs, endDate)
	if err != nil {
		return nil, err
	}
	currentBucket, err := momentumCurrentBucket(def, endDate, countsByDate)
	if err != nil {
		return nil, err
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
	streamNamesByID := map[int64]string{}
	for _, repo := range repos {
		streams, err := c.Streams.ListByRepo(ctx, repo.ID, c.UserID)
		if err != nil {
			return nil, err
		}
		for _, stream := range streams {
			streamNamesByID[stream.ID] = stream.Name
		}
	}

	detail := &sharedtypes.MomentumDetail{
		Definition:  def,
		HabitNames:  momentumHabitNames(def.HabitIDs, habitNamesByID),
		TargetNames: momentumTargetNames(def, habitNamesByID, repoNamesByID, streamNamesByID),
		CurrentBucket: sharedtypes.MomentumBucketDetail{
			Label:     currentBucket.Label,
			StartDate: currentBucket.StartDate,
			EndDate:   currentBucket.EndDate,
			Count:     currentBucket.Count,
			Target:    currentBucket.Target,
			MetTarget: currentBucket.MetTarget,
		},
	}
	switch sharedtypes.NormalizeMomentumTargetKind(def.TargetKind) {
	case sharedtypes.MomentumTargetKindContext:
		detail.Contributors, err = momentumContextContributors(
			ctx,
			c,
			def,
			currentBucket.StartDate,
			currentBucket.EndDate,
		)
	default:
		detail.Contributors, err = momentumHabitContributors(
			ctx,
			c,
			def,
			currentBucket.StartDate,
			currentBucket.EndDate,
		)
	}
	if err != nil {
		return nil, err
	}
	_ = countsByDate
	return detail, nil
}

func momentumCurrentBucket(
	def sharedtypes.HabitStreakDefinition,
	endDate string,
	countsByDate map[string]map[string]int,
) (sharedtypes.MomentumSeriesPoint, error) {
	key := customHabitBucketKey(endDate, def.Period)
	startDate, endBucketDate := momentumBucketBounds(key, def.Period)
	if startDate == "" || endBucketDate == "" {
		return sharedtypes.MomentumSeriesPoint{}, errors.New("could not resolve momentum bucket")
	}
	bucketEnd := minISODate(endDate, endBucketDate)
	series := buildMomentumSeries(def, startDate, bucketEnd, countsByDate)
	if len(series) == 0 {
		return sharedtypes.MomentumSeriesPoint{}, errors.New("could not resolve momentum bucket")
	}
	point := series[len(series)-1]
	return sharedtypes.MomentumSeriesPoint{
		BucketKey: key,
		Label:     point.Label,
		StartDate: startDate,
		EndDate:   bucketEnd,
		Count:     point.Count,
		Target:    point.Target,
		MetTarget: point.MetTarget,
	}, nil
}

func momentumHabitContributors(
	ctx context.Context,
	c *core.Context,
	def sharedtypes.HabitStreakDefinition,
	startDate string,
	endDate string,
) ([]sharedtypes.MomentumContributor, error) {
	history, err := c.HabitCompletions.ListHistory(ctx, c.UserID, nil, nil)
	if err != nil {
		return nil, err
	}
	out := make([]sharedtypes.MomentumContributor, 0)
	for _, entry := range history {
		if entry.Status != sharedtypes.HabitCompletionStatusCompleted {
			continue
		}
		if entry.Date < startDate || entry.Date > endDate {
			continue
		}
		if !slices.Contains(def.HabitIDs, entry.HabitID) {
			continue
		}
		contextLabel := strings.TrimSpace(entry.RepoName)
		if stream := strings.TrimSpace(entry.StreamName); stream != "" {
			if contextLabel != "" {
				contextLabel += " / " + stream
			} else {
				contextLabel = stream
			}
		}
		amount := "Completed"
		if entry.DurationMinutes != nil {
			amount = formatCompactDurationMinutes(*entry.DurationMinutes)
		}
		habitID := entry.HabitID
		out = append(out, sharedtypes.MomentumContributor{
			Kind:        sharedtypes.MomentumContributorKindHabitCompletion,
			Title:       fallbackString(strings.TrimSpace(entry.HabitName), fmt.Sprintf("Habit %d", entry.HabitID)),
			Context:     fallbackString(contextLabel, "-"),
			Date:        entry.Date,
			AmountLabel: amount,
			Meta:        strings.ReplaceAll(string(entry.Status), "_", " "),
			HabitID:     &habitID,
		})
	}
	slices.SortFunc(out, func(left, right sharedtypes.MomentumContributor) int {
		switch {
		case left.Date > right.Date:
			return -1
		case left.Date < right.Date:
			return 1
		default:
			return strings.Compare(left.Title, right.Title)
		}
	})
	return out, nil
}

func momentumContextContributors(
	ctx context.Context,
	c *core.Context,
	def sharedtypes.HabitStreakDefinition,
	startDate string,
	endDate string,
) ([]sharedtypes.MomentumContributor, error) {
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
		Since:  stringPtr(startDate + "T00:00:00Z"),
		Until:  stringPtr(endDate + "T23:59:59.999Z"),
	})
	if err != nil {
		return nil, err
	}
	issues, err := c.Issues.ListAll(ctx, c.UserID)
	if err != nil {
		return nil, err
	}
	issueMetaByID := make(map[int64]sharedtypes.IssueWithMeta, len(issues))
	for _, issue := range issues {
		issueMetaByID[issue.ID] = issue
	}
	out := make([]sharedtypes.MomentumContributor, 0)
	for _, session := range sessions {
		day := extractISODate(session.StartTime)
		if day < startDate || day > endDate {
			continue
		}
		meta, ok := issueMetaByID[session.IssueID]
		if !ok {
			continue
		}
		if !momentumContextsContain(def.Contexts, meta.RepoID, meta.StreamID) {
			continue
		}
		duration := 0
		if session.DurationSeconds != nil {
			duration = *session.DurationSeconds
		}
		if duration <= 0 {
			continue
		}
		sessionID := session.ID
		out = append(out, sharedtypes.MomentumContributor{
			Kind:        sharedtypes.MomentumContributorKindSession,
			Title:       fallbackString(strings.TrimSpace(meta.Title), fmt.Sprintf("Issue %d", session.IssueID)),
			Context:     strings.TrimSpace(meta.RepoName + " / " + meta.StreamName),
			Date:        day,
			AmountLabel: formatCompactDurationSeconds(duration),
			Meta:        "Session",
			SessionID:   &sessionID,
		})
	}
	slices.SortFunc(out, func(left, right sharedtypes.MomentumContributor) int {
		switch {
		case left.Date > right.Date:
			return -1
		case left.Date < right.Date:
			return 1
		default:
			return strings.Compare(left.Title, right.Title)
		}
	})
	return out, nil
}

func momentumContextsContain(contexts []sharedtypes.MomentumContext, repoID, streamID int64) bool {
	for _, contextItem := range contexts {
		if contextItem.RepoID != repoID {
			continue
		}
		if contextItem.StreamID == nil || *contextItem.StreamID == streamID {
			return true
		}
	}
	return false
}

func fallbackString(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return strings.TrimSpace(value)
}

func minISODate(left, right string) string {
	if right == "" {
		return left
	}
	if left == "" || left > right {
		return right
	}
	return left
}

func formatCompactDurationMinutes(totalMinutes int) string {
	if totalMinutes <= 0 {
		return "0m"
	}
	hours := totalMinutes / 60
	minutes := totalMinutes % 60
	switch {
	case hours > 0 && minutes > 0:
		return fmt.Sprintf("%dh%dm", hours, minutes)
	case hours > 0:
		return fmt.Sprintf("%dh", hours)
	default:
		return fmt.Sprintf("%dm", minutes)
	}
}

func formatCompactDurationSeconds(totalSeconds int) string {
	if totalSeconds <= 0 {
		return "0s"
	}
	duration := time.Duration(totalSeconds) * time.Second
	hours := int(duration / time.Hour)
	minutes := int(duration%time.Hour) / int(time.Minute)
	seconds := int(duration%time.Minute) / int(time.Second)
	switch {
	case hours > 0 && minutes > 0:
		return fmt.Sprintf("%dh%dm", hours, minutes)
	case hours > 0 && seconds > 0:
		return fmt.Sprintf("%dh%ds", hours, seconds)
	case hours > 0:
		return fmt.Sprintf("%dh", hours)
	case minutes > 0 && seconds > 0:
		return fmt.Sprintf("%dm%ds", minutes, seconds)
	case minutes > 0:
		return fmt.Sprintf("%dm", minutes)
	default:
		return fmt.Sprintf("%ds", seconds)
	}
}
