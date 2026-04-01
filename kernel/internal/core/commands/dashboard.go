package commands

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"

	"crona/kernel/internal/core"
	shareddto "crona/shared/dto"
	sharedtypes "crona/shared/types"
)

func ComputeDashboardWindowSummary(ctx context.Context, c *core.Context, input shareddto.DashboardWindowQuery) (*sharedtypes.DashboardWindowSummary, error) {
	if err := validateRange(input.Start, input.End); err != nil {
		return nil, err
	}
	issues, err := scopedIssues(ctx, c, input.RepoID, input.StreamID, input.IssueID)
	if err != nil {
		return nil, err
	}
	allowed := make(map[int64]sharedtypes.IssueWithMeta, len(issues))
	for _, issue := range issues {
		allowed[issue.ID] = issue
	}
	summary := &sharedtypes.DashboardWindowSummary{
		StartDate: input.Start,
		EndDate:   input.End,
	}
	var accountabilityTotal float64
	var accountabilityDays int
	for day := range eachDate(input.Start, input.End) {
		plan, err := GetDailyPlan(ctx, c, day)
		if err != nil {
			return nil, err
		}
		item := sharedtypes.DashboardWindowDay{
			Date:   day,
			Status: sharedtypes.DashboardWindowDayEmpty,
		}
		if plan == nil {
			summary.Days = append(summary.Days, item)
			continue
		}
		for _, entry := range plan.Entries {
			if len(allowed) > 0 {
				if _, ok := allowed[entry.IssueID]; !ok {
					continue
				}
			}
			item.PlannedCount++
			summary.PlannedCount++
			if isCarryOverEntry(entry, day) {
				item.CarryOverCount++
				summary.CarryOverCount++
			}
			switch entry.Status {
			case sharedtypes.DailyPlanEntryStatusCompleted:
				item.CompletedCount++
				summary.CompletedCount++
			case sharedtypes.DailyPlanEntryStatusFailed:
				item.FailedCount++
				summary.FailedCount++
				if entry.FailureReason != nil && *entry.FailureReason == sharedtypes.DailyPlanFailureReasonMissed {
					summary.MissedCount++
				}
			case sharedtypes.DailyPlanEntryStatusAbandoned:
				summary.AbandonedCount++
			}
		}
		if item.PlannedCount > 0 {
			accountabilityTotal += plan.Summary.AccountabilityScore
			accountabilityDays++
			item.AccountabilityScore = plan.Summary.AccountabilityScore
		}
		item.Status = dashboardWindowDayStatus(item)
		summary.Days = append(summary.Days, item)
	}
	if accountabilityDays > 0 {
		summary.AccountabilityScore = accountabilityTotal / float64(accountabilityDays)
	}
	return summary, nil
}

func ComputeFocusScoreSummary(ctx context.Context, c *core.Context, input shareddto.DashboardSummaryQuery) (*sharedtypes.FocusScoreSummary, error) {
	if err := validateRange(input.Start, input.End); err != nil {
		return nil, err
	}
	days, err := ComputeMetricsRange(ctx, c, input.Start, input.End)
	if err != nil {
		return nil, err
	}
	rollup, err := ComputeMetricsRollup(ctx, c, input.Start, input.End)
	if err != nil {
		return nil, err
	}
	settings, err := c.CoreSettings.Get(ctx, c.UserID)
	if err != nil {
		return nil, err
	}
	targetPerDay := 100 * 60
	if settings != nil && settings.WorkDurationMinutes > 0 {
		targetPerDay = settings.WorkDurationMinutes * 4 * 60
	}
	targetTotal := targetPerDay * max(1, len(days))
	workedRatio := 0.0
	if targetTotal > 0 {
		workedRatio = float64(rollup.WorkedSeconds) / float64(targetTotal)
	}
	restRatio := 0.0
	if rollup.WorkedSeconds > 0 {
		restRatio = float64(rollup.RestSeconds) / float64(rollup.WorkedSeconds)
	}
	consistency := 0.0
	if len(days) > 0 {
		consistency = float64(rollup.FocusDays) / float64(len(days))
	}
	breakScore := 0.0
	switch {
	case restRatio >= 0.15 && restRatio <= 0.40:
		breakScore = 1.0
	case restRatio > 0:
		breakScore = clamp01(1.0 - math.Abs(restRatio-0.22)/0.22)
	}
	base := math.Min(workedRatio, 1.0) * 55.0
	consistencyScore := consistency * 20.0
	breakContribution := breakScore * 25.0
	overworkPenalty := 0.0
	if workedRatio > 1.20 {
		overworkPenalty = math.Min((workedRatio-1.20)/0.80*20.0, 20.0)
	}
	noBreakPenalty := 0.0
	if rollup.WorkedSeconds > targetPerDay && restRatio < 0.10 {
		noBreakPenalty = 10.0
	}
	spikePenalty := focusSpikePenalty(days)
	score := int(math.Round(math.Max(0, math.Min(100, base+consistencyScore+breakContribution-overworkPenalty-noBreakPenalty-spikePenalty))))
	level := sharedtypes.FocusScoreLevelLow
	switch {
	case workedRatio > 1.30 && restRatio < 0.12:
		level = sharedtypes.FocusScoreLevelOverextended
	case score >= 75:
		level = sharedtypes.FocusScoreLevelStrong
	case score >= 45:
		level = sharedtypes.FocusScoreLevelSteady
	}
	return &sharedtypes.FocusScoreSummary{
		StartDate:           input.Start,
		EndDate:             input.End,
		Score:               score,
		Level:               level,
		WorkedSeconds:       rollup.WorkedSeconds,
		RestSeconds:         rollup.RestSeconds,
		SessionCount:        rollup.SessionCount,
		FocusDays:           rollup.FocusDays,
		Days:                rollup.Days,
		TargetWorkedSeconds: targetTotal,
	}, nil
}

func ComputeTimeDistributionSummary(ctx context.Context, c *core.Context, input shareddto.DashboardSummaryQuery) (*sharedtypes.TimeDistributionSummary, error) {
	if err := validateRange(input.Start, input.End); err != nil {
		return nil, err
	}
	groupBy := parseDistributionGroup(input.GroupBy)
	issues, err := scopedIssues(ctx, c, input.RepoID, input.StreamID, input.IssueID)
	if err != nil {
		return nil, err
	}
	issueByID := make(map[int64]sharedtypes.IssueWithMeta, len(issues))
	for _, issue := range issues {
		issueByID[issue.ID] = issue
	}
	sessions, err := listScopedSessions(ctx, c, input.Start, input.End, input.RepoID, input.StreamID, input.IssueID)
	if err != nil {
		return nil, err
	}
	workBySession, segmentTotals, err := sessionWorkTotals(ctx, c, input.Start, input.End, sessions)
	if err != nil {
		return nil, err
	}
	accum := map[string]int{}
	labels := map[string]string{}
	switch groupBy {
	case sharedtypes.DistributionGroupSegmentType:
		for key, total := range segmentTotals {
			accum[key] = total
			labels[key] = prettifySegmentGroup(key)
		}
	default:
		for _, session := range sessions {
			seconds := workBySession[session.ID]
			if seconds <= 0 {
				continue
			}
			meta, ok := issueByID[session.IssueID]
			if !ok {
				continue
			}
			key, label := distributionKeyAndLabel(groupBy, meta)
			accum[key] += seconds
			labels[key] = label
		}
	}
	rows, total := distributionRows(accum, labels)
	return &sharedtypes.TimeDistributionSummary{
		StartDate:    input.Start,
		EndDate:      input.End,
		GroupBy:      groupBy,
		TotalSeconds: total,
		Rows:         rows,
	}, nil
}

func ComputeGoalProgressSummary(ctx context.Context, c *core.Context, input shareddto.DashboardSummaryQuery) (*sharedtypes.GoalProgressSummary, error) {
	if err := validateRange(input.Start, input.End); err != nil {
		return nil, err
	}
	groupBy := parseGoalProgressGroup(input.GroupBy)
	issues, err := scopedIssues(ctx, c, input.RepoID, input.StreamID, input.IssueID)
	if err != nil {
		return nil, err
	}
	sessions, err := listScopedSessions(ctx, c, input.Start, input.End, input.RepoID, input.StreamID, input.IssueID)
	if err != nil {
		return nil, err
	}
	workBySession, _, err := sessionWorkTotals(ctx, c, input.Start, input.End, sessions)
	if err != nil {
		return nil, err
	}
	workByIssue := map[int64]int{}
	for _, session := range sessions {
		workByIssue[session.IssueID] += workBySession[session.ID]
	}
	type aggregate struct {
		label    string
		estimate int
		actual   int
	}
	grouped := map[string]*aggregate{}
	for _, issue := range issues {
		key, label := goalProgressKeyAndLabel(groupBy, issue)
		current := grouped[key]
		if current == nil {
			current = &aggregate{label: label}
			grouped[key] = current
		}
		if issue.EstimateMinutes != nil && *issue.EstimateMinutes > 0 {
			current.estimate += *issue.EstimateMinutes
		}
		current.actual += workByIssue[issue.ID]
	}
	rows := make([]sharedtypes.GoalProgressRow, 0, len(grouped))
	totalEstimate := 0
	totalActual := 0
	estimatedItems := 0
	deltaMinutesTotal := 0.0
	deltaPercentTotal := 0.0
	for key, item := range grouped {
		totalEstimate += item.estimate
		totalActual += item.actual
		if item.estimate > 0 && item.actual > 0 {
			deltaMinutes := float64(item.actual)/60.0 - float64(item.estimate)
			deltaPercent := (deltaMinutes / float64(item.estimate)) * 100.0
			deltaMinutesTotal += deltaMinutes
			deltaPercentTotal += deltaPercent
			estimatedItems++
		}
		row := sharedtypes.GoalProgressRow{
			Key:             key,
			Label:           item.label,
			EstimateMinutes: item.estimate,
			ActualSeconds:   item.actual,
			Status:          sharedtypes.GoalProgressStatusUnestimated,
		}
		if item.estimate > 0 {
			progress := (float64(item.actual) / float64(item.estimate*60)) * 100.0
			row.ProgressPercent = progress
			row.Status = goalProgressStatus(progress)
		}
		rows = append(rows, row)
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].ActualSeconds == rows[j].ActualSeconds {
			return rows[i].Label < rows[j].Label
		}
		return rows[i].ActualSeconds > rows[j].ActualSeconds
	})
	averageDeltaMinutes := 0.0
	averageDeltaPercent := 0.0
	bias := "balanced"
	if estimatedItems > 0 {
		averageDeltaMinutes = deltaMinutesTotal / float64(estimatedItems)
		averageDeltaPercent = deltaPercentTotal / float64(estimatedItems)
		switch {
		case averageDeltaMinutes > 5:
			bias = "under"
		case averageDeltaMinutes < -5:
			bias = "over"
		}
	}
	return &sharedtypes.GoalProgressSummary{
		StartDate:            input.Start,
		EndDate:              input.End,
		GroupBy:              groupBy,
		TotalEstimateMinutes: totalEstimate,
		TotalActualSeconds:   totalActual,
		EstimatedItems:       estimatedItems,
		AverageDeltaMinutes:  averageDeltaMinutes,
		AverageDeltaPercent:  averageDeltaPercent,
		EstimateBias:         bias,
		Rows:                 rows,
	}, nil
}

func scopedIssues(ctx context.Context, c *core.Context, repoID, streamID, issueID *int64) ([]sharedtypes.IssueWithMeta, error) {
	allIssues, err := c.Issues.ListAll(ctx, c.UserID)
	if err != nil {
		return nil, err
	}
	out := make([]sharedtypes.IssueWithMeta, 0, len(allIssues))
	for _, issue := range allIssues {
		if repoID != nil && issue.RepoID != *repoID {
			continue
		}
		if streamID != nil && issue.StreamID != *streamID {
			continue
		}
		if issueID != nil && issue.ID != *issueID {
			continue
		}
		out = append(out, issue)
	}
	return out, nil
}

func listScopedSessions(ctx context.Context, c *core.Context, start, end string, repoID, streamID, issueID *int64) ([]sharedtypes.SessionHistoryEntry, error) {
	startTime := start + "T00:00:00.000Z"
	endTime := end + "T23:59:59.999Z"
	return c.Sessions.ListEnded(ctx, struct {
		UserID   string
		RepoID   *int64
		StreamID *int64
		IssueID  *int64
		Since    *string
		Until    *string
		Limit    *int
		Offset   *int
	}{
		UserID:   c.UserID,
		RepoID:   repoID,
		StreamID: streamID,
		IssueID:  issueID,
		Since:    &startTime,
		Until:    &endTime,
	})
}

func sessionWorkTotals(ctx context.Context, c *core.Context, start, end string, sessions []sharedtypes.SessionHistoryEntry) (map[string]int, map[string]int, error) {
	workBySession := map[string]int{}
	segmentTotals := map[string]int{
		string(sharedtypes.SessionSegmentWork):       0,
		string(sharedtypes.SessionSegmentShortBreak): 0,
		string(sharedtypes.SessionSegmentLongBreak):  0,
		string(sharedtypes.SessionSegmentRest):       0,
	}
	if len(sessions) == 0 {
		return workBySession, segmentTotals, nil
	}
	allowed := make(map[string]struct{}, len(sessions))
	for _, session := range sessions {
		allowed[session.ID] = struct{}{}
	}
	startTime := start + "T00:00:00.000Z"
	endTime := end + "T23:59:59.999Z"
	segments, err := c.SessionSegments.ListEndedInRange(ctx, c.UserID, startTime, endTime)
	if err != nil {
		return nil, nil, err
	}
	for _, segment := range segments {
		if _, ok := allowed[segment.SessionID]; !ok {
			continue
		}
		seconds := segmentDurationSeconds(segment)
		if seconds <= 0 {
			continue
		}
		switch segment.SegmentType {
		case sharedtypes.SessionSegmentWork:
			workBySession[segment.SessionID] += seconds
			segmentTotals[string(sharedtypes.SessionSegmentWork)] += seconds
		default:
			segmentTotals[string(segment.SegmentType)] += seconds
		}
	}
	for _, session := range sessions {
		if workBySession[session.ID] == 0 && session.DurationSeconds != nil && *session.DurationSeconds > 0 {
			workBySession[session.ID] = *session.DurationSeconds
		}
	}
	return workBySession, segmentTotals, nil
}

func distributionRows(accum map[string]int, labels map[string]string) ([]sharedtypes.TimeDistributionRow, int) {
	total := 0
	for _, seconds := range accum {
		total += seconds
	}
	rows := make([]sharedtypes.TimeDistributionRow, 0, len(accum))
	for key, seconds := range accum {
		row := sharedtypes.TimeDistributionRow{
			Key:           key,
			Label:         firstNonEmptyString(labels[key], key),
			WorkedSeconds: seconds,
		}
		if total > 0 {
			row.Percent = (float64(seconds) / float64(total)) * 100.0
		}
		rows = append(rows, row)
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].WorkedSeconds == rows[j].WorkedSeconds {
			return rows[i].Label < rows[j].Label
		}
		return rows[i].WorkedSeconds > rows[j].WorkedSeconds
	})
	return rows, total
}

func distributionKeyAndLabel(groupBy sharedtypes.DistributionGroup, issue sharedtypes.IssueWithMeta) (string, string) {
	switch groupBy {
	case sharedtypes.DistributionGroupStream:
		return fmt.Sprintf("stream:%d", issue.StreamID), issue.StreamName
	case sharedtypes.DistributionGroupIssue:
		return fmt.Sprintf("issue:%d", issue.ID), issue.Title
	default:
		return fmt.Sprintf("repo:%d", issue.RepoID), issue.RepoName
	}
}

func goalProgressKeyAndLabel(groupBy sharedtypes.GoalProgressGroup, issue sharedtypes.IssueWithMeta) (string, string) {
	switch groupBy {
	case sharedtypes.GoalProgressGroupIssue:
		return fmt.Sprintf("issue:%d", issue.ID), issue.Title
	case sharedtypes.GoalProgressGroupStream:
		return fmt.Sprintf("stream:%d", issue.StreamID), issue.StreamName
	default:
		return fmt.Sprintf("repo:%d", issue.RepoID), issue.RepoName
	}
}

func dashboardWindowDayStatus(day sharedtypes.DashboardWindowDay) sharedtypes.DashboardWindowDayStatus {
	switch {
	case day.CarryOverCount > 0 && (day.CompletedCount > 0 || day.FailedCount > 0):
		return sharedtypes.DashboardWindowDayMixed
	case day.CarryOverCount > 0:
		return sharedtypes.DashboardWindowDayCarryOver
	case day.FailedCount > 0:
		return sharedtypes.DashboardWindowDayMissed
	case day.CompletedCount > 0:
		return sharedtypes.DashboardWindowDayDone
	case day.PlannedCount > 0:
		return sharedtypes.DashboardWindowDayPlanned
	default:
		return sharedtypes.DashboardWindowDayEmpty
	}
}

func isCarryOverEntry(entry sharedtypes.DailyPlanEntry, day string) bool {
	if entry.Status != sharedtypes.DailyPlanEntryStatusPlanned {
		return false
	}
	baseline := strings.TrimSpace(entry.BaselineDate)
	current := strings.TrimSpace(entry.CurrentPlannedDate)
	if baseline == "" {
		baseline = day
	}
	if current == "" {
		current = day
	}
	return baseline < current && current == day
}

func parseDistributionGroup(value string) sharedtypes.DistributionGroup {
	switch sharedtypes.DistributionGroup(strings.TrimSpace(value)) {
	case sharedtypes.DistributionGroupStream:
		return sharedtypes.DistributionGroupStream
	case sharedtypes.DistributionGroupIssue:
		return sharedtypes.DistributionGroupIssue
	case sharedtypes.DistributionGroupSegmentType:
		return sharedtypes.DistributionGroupSegmentType
	default:
		return sharedtypes.DistributionGroupRepo
	}
}

func parseGoalProgressGroup(value string) sharedtypes.GoalProgressGroup {
	switch sharedtypes.GoalProgressGroup(strings.TrimSpace(value)) {
	case sharedtypes.GoalProgressGroupIssue:
		return sharedtypes.GoalProgressGroupIssue
	case sharedtypes.GoalProgressGroupStream:
		return sharedtypes.GoalProgressGroupStream
	default:
		return sharedtypes.GoalProgressGroupRepo
	}
}

func goalProgressStatus(progress float64) sharedtypes.GoalProgressStatus {
	switch {
	case progress <= 100:
		return sharedtypes.GoalProgressStatusOnTrack
	case progress <= 120:
		return sharedtypes.GoalProgressStatusAtRisk
	default:
		return sharedtypes.GoalProgressStatusOver
	}
}

func focusSpikePenalty(days []sharedtypes.DailyMetricsDay) float64 {
	if len(days) < 2 {
		return 0
	}
	total := 0.0
	maxWorked := 0.0
	for _, day := range days {
		worked := float64(day.WorkedSeconds)
		total += worked
		if worked > maxWorked {
			maxWorked = worked
		}
	}
	avg := total / float64(len(days))
	if avg <= 0 {
		return 0
	}
	return math.Min(clamp01((maxWorked-(avg*1.5))/(avg*1.5))*10.0, 10.0)
}

func prettifySegmentGroup(value string) string {
	switch value {
	case string(sharedtypes.SessionSegmentWork):
		return "Work"
	case string(sharedtypes.SessionSegmentShortBreak):
		return "Short Break"
	case string(sharedtypes.SessionSegmentLongBreak):
		return "Long Break"
	case string(sharedtypes.SessionSegmentRest):
		return "Rest"
	default:
		return value
	}
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
