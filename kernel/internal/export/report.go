package export

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"crona/kernel/internal/core"
	corecommands "crona/kernel/internal/core/commands"
	"crona/kernel/internal/runtime"
	"crona/kernel/internal/sessionnotes"
	sharedtypes "crona/shared/types"
)

func GenerateDailyReport(ctx context.Context, c *core.Context, paths runtime.Paths, date string, mode sharedtypes.ExportOutputMode) (*sharedtypes.DailyReportResult, error) {
	return GenerateDailyReportWithFormat(ctx, c, paths, date, sharedtypes.ExportFormatMarkdown, mode)
}

func GenerateDailyReportWithFormat(ctx context.Context, c *core.Context, paths runtime.Paths, date string, format sharedtypes.ExportFormat, mode sharedtypes.ExportOutputMode) (*sharedtypes.DailyReportResult, error) {
	return generateDailyReportWithKind(ctx, c, paths, normalizeReportDate(date), normalizeNarrativeFormat(format), mode, "")
}

func generateDailyReportWithKind(ctx context.Context, c *core.Context, paths runtime.Paths, date string, format sharedtypes.ExportFormat, mode sharedtypes.ExportOutputMode, presetID string) (*sharedtypes.DailyReportResult, error) {
	if strings.TrimSpace(date) == "" {
		date = time.Now().Format("2006-01-02")
	}
	format = normalizeExportFormat(format)
	data, err := BuildDailyReportData(ctx, c, date)
	if err != nil {
		return nil, err
	}
	status, err := EnsureAssets(paths)
	if err != nil {
		return nil, err
	}
	result := &sharedtypes.DailyReportResult{
		Kind:       sharedtypes.ExportReportKindDaily,
		Label:      "Daily Report",
		Date:       date,
		Format:     format,
		OutputMode: mode,
		Assets:     status,
	}
	switch format {
	case sharedtypes.ExportFormatPDF:
		htmlTemplate, cssTemplate, _, err := LoadNarrativePDFAssets(paths, sharedtypes.ExportReportKindDaily, presetID)
		if err != nil {
			return nil, err
		}
		renderedHTML, err := RenderTemplate(string(htmlTemplate), buildTemplateDataMap(data))
		if err != nil {
			return nil, err
		}
		result.Content = renderedHTML
		if mode != sharedtypes.ExportOutputModeFile {
			return nil, errors.New("pdf export only supports file output")
		}
		filePath, renderer, err := RenderNarrativePDFReport(paths, reportWriteSpec{
			Kind:     sharedtypes.ExportReportKindDaily,
			Label:    "Daily Report",
			Date:     date,
			Format:   format,
			BaseName: "daily-" + date,
		}, renderedHTML, string(cssTemplate))
		if err != nil {
			return nil, err
		}
		result.FilePath = &filePath
		result.Renderer = &renderer
	default:
		spec := reportWriteSpec{
			Kind:     sharedtypes.ExportReportKindDaily,
			Label:    "Daily Report",
			Date:     date,
			Format:   format,
			BaseName: "daily-" + date,
		}
		templateBody, _, err := LoadReportTemplate(paths, sharedtypes.ExportReportKindDaily, format, presetID)
		if err != nil {
			return nil, err
		}
		templateData := attachFrontmatter(buildTemplateDataMap(data), spec)
		rendered, err := RenderTemplate(string(templateBody), templateData)
		if err != nil {
			return nil, err
		}
		result.Content = ensureMarkdownFrontmatter(rendered, templateData, spec)
		if mode == sharedtypes.ExportOutputModeFile {
			filePath, err := WriteReport(paths, spec, []byte(result.Content))
			if err != nil {
				return nil, err
			}
			result.FilePath = &filePath
		}
	}
	return result, nil
}

func BuildDailyReportData(ctx context.Context, c *core.Context, date string) (*sharedtypes.DailyReportData, error) {
	summary, err := corecommands.ComputeDailyIssueSummaryForDate(ctx, c, date)
	if err != nil {
		return nil, err
	}
	plan, err := corecommands.GetDailyPlan(ctx, c, date)
	if err != nil {
		return nil, err
	}
	allIssues, err := corecommands.ListAllIssues(ctx, c)
	if err != nil {
		return nil, err
	}
	issueMeta := make(map[int64]sharedtypes.IssueWithMeta, len(allIssues))
	for _, issue := range allIssues {
		issueMeta[issue.ID] = issue
	}

	start := date + "T00:00:00.000Z"
	end := date + "T23:59:59.999Z"
	history, err := c.Sessions.ListEnded(ctx, struct {
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
		Since:  &start,
		Until:  &end,
	})
	if err != nil {
		return nil, err
	}
	sessions := make([]sharedtypes.DailyReportSession, 0, len(history))
	workedByIssue := map[int64]int{}
	for _, entry := range history {
		meta := issueMeta[entry.IssueID]
		parsed := sessionnotes.Parse(entry.Notes)
		entry.ParsedNotes = parsed
		if entry.DurationSeconds != nil {
			workedByIssue[entry.IssueID] += *entry.DurationSeconds
		}
		sessions = append(sessions, sharedtypes.DailyReportSession{
			SessionHistoryEntry: entry,
			RepoID:              meta.RepoID,
			RepoName:            meta.RepoName,
			StreamID:            meta.StreamID,
			StreamName:          meta.StreamName,
			IssueTitle:          meta.Title,
		})
	}

	issues := make([]sharedtypes.DailyReportIssue, 0)
	if plan != nil && len(plan.Entries) > 0 {
		for _, entry := range plan.Entries {
			meta, ok := issueMeta[entry.IssueID]
			if !ok {
				continue
			}
			issues = append(issues, sharedtypes.DailyReportIssue{
				IssueWithMeta: sharedtypes.IssueWithMeta{
					Issue:      meta.Issue,
					RepoID:     meta.RepoID,
					RepoName:   meta.RepoName,
					StreamName: meta.StreamName,
				},
				WorkedSeconds:            workedByIssue[entry.IssueID],
				PlanStatus:               entry.Status,
				PlanCommittedAt:          entry.CommittedAt,
				PlanResolvedAt:           entry.ResolvedAt,
				PlanFailureReason:        entry.FailureReason,
				PlanPendingFailureAt:     entry.PendingFailureAt,
				PlanPendingFailureReason: entry.PendingFailureReason,
				PlanBaselineDate:         entry.BaselineDate,
				PlanCurrentPlannedDate:   entry.CurrentPlannedDate,
				PlanPostponeCount:        entry.PostponeCount,
				PlanCurrentDelayedDays:   entry.CurrentDelayedDays,
				PlanMaxDelayedDays:       entry.MaxDelayedDays,
				PlanFailScore:            entry.FailScore,
			})
		}
	} else {
		for _, issue := range summary.Issues {
			meta := issueMeta[issue.ID]
			issues = append(issues, sharedtypes.DailyReportIssue{
				IssueWithMeta: sharedtypes.IssueWithMeta{
					Issue:      issue,
					RepoID:     meta.RepoID,
					RepoName:   meta.RepoName,
					StreamName: meta.StreamName,
				},
				WorkedSeconds:          workedByIssue[issue.ID],
				PlanStatus:             sharedtypes.DailyPlanEntryStatusPlanned,
				PlanCommittedAt:        date + "T00:00:00Z",
				PlanBaselineDate:       date,
				PlanCurrentPlannedDate: date,
			})
		}
	}

	habits, err := corecommands.ListHabitsDueForDate(ctx, c, date)
	if err != nil {
		return nil, err
	}
	checkIn, err := corecommands.GetDailyCheckIn(ctx, c, date)
	if err != nil {
		return nil, err
	}
	days, err := corecommands.ComputeMetricsRange(ctx, c, date, date)
	if err != nil {
		return nil, err
	}
	rollup := corecommands.ComputeMetricsRollupFromDays(date, date, days)
	settings, err := c.CoreSettings.Get(ctx, c.UserID)
	if err != nil {
		return nil, err
	}
	streaks := corecommands.ComputeMetricsStreaksFromDays(days, settings)
	var metrics *sharedtypes.DailyMetricsDay
	if len(days) > 0 {
		copy := days[0]
		metrics = &copy
	}

	return &sharedtypes.DailyReportData{
		Date:          date,
		GeneratedAt:   time.Now().UTC().Format(time.RFC3339),
		Summary:       summary,
		Plan:          plan,
		Issues:        issues,
		Sessions:      sessions,
		Habits:        habits,
		CheckIn:       checkIn,
		Metrics:       metrics,
		MetricsRollup: rollup,
		Streaks:       streaks,
	}, nil
}

func buildTemplateDataMap(data *sharedtypes.DailyReportData) map[string]any {
	summary := buildSummaryMap(data)
	highlights := buildHighlights(data)
	risks := buildRisks(data)
	dayHealth := deriveDayHealth(data, summary, risks)
	items := map[string]any{
		"date":          data.Date,
		"generatedAt":   data.GeneratedAt,
		"summary":       summary,
		"plan":          nil,
		"issues":        make([]map[string]any, 0, len(data.Issues)),
		"repos":         groupIssuesByRepo(data.Issues),
		"habitRepos":    groupHabitsByRepo(data.Habits),
		"sessions":      make([]map[string]any, 0, len(data.Sessions)),
		"habits":        make([]map[string]any, 0, len(data.Habits)),
		"highlights":    highlights,
		"risks":         risks,
		"dayHealth":     dayHealth,
		"checkIn":       nil,
		"metrics":       nil,
		"metricsRollup": nil,
		"streaks":       nil,
	}
	if data.CheckIn != nil {
		items["checkIn"] = mapCheckIn(*data.CheckIn)
	}
	if data.Metrics != nil {
		items["metrics"] = mapMetrics(*data.Metrics)
	}
	if data.MetricsRollup != nil {
		items["metricsRollup"] = mapRollup(*data.MetricsRollup)
	}
	if data.Streaks != nil {
		items["streaks"] = map[string]any{
			"currentFocusDays":   data.Streaks.CurrentFocusDays,
			"longestFocusDays":   data.Streaks.LongestFocusDays,
			"currentCheckInDays": data.Streaks.CurrentCheckInDays,
			"longestCheckInDays": data.Streaks.LongestCheckInDays,
		}
	}
	if data.Plan != nil {
		entries := make([]map[string]any, 0, len(data.Plan.Entries))
		for _, entry := range data.Plan.Entries {
			item := map[string]any{
				"id":                   entry.ID,
				"date":                 entry.Date,
				"issueId":              entry.IssueID,
				"status":               entry.Status,
				"committedAt":          entry.CommittedAt,
				"pendingFailureAt":     entry.PendingFailureAt,
				"pendingFailureReason": entry.PendingFailureReason,
				"resolvedAt":           entry.ResolvedAt,
				"failureReason":        entry.FailureReason,
				"baselineDate":         entry.BaselineDate,
				"currentPlannedDate":   entry.CurrentPlannedDate,
				"postponeCount":        entry.PostponeCount,
				"currentDelayedDays":   entry.CurrentDelayedDays,
				"maxDelayedDays":       entry.MaxDelayedDays,
				"failScore":            entry.FailScore,
			}
			entries = append(entries, item)
		}
		items["plan"] = map[string]any{
			"id":   data.Plan.ID,
			"date": data.Plan.Date,
			"summary": map[string]any{
				"plannedCount":         data.Plan.Summary.PlannedCount,
				"completedCount":       data.Plan.Summary.CompletedCount,
				"failedCount":          data.Plan.Summary.FailedCount,
				"abandonedCount":       data.Plan.Summary.AbandonedCount,
				"pendingRollbackCount": data.Plan.Summary.PendingRollbackCount,
				"accountabilityScore":  data.Plan.Summary.AccountabilityScore,
				"backlogPressure":      data.Plan.Summary.BacklogPressure,
				"delayedIssueCount":    data.Plan.Summary.DelayedIssueCount,
				"highRiskIssueCount":   data.Plan.Summary.HighRiskIssueCount,
				"avgDelayDays":         data.Plan.Summary.AvgDelayDays,
				"maxDelayDays":         data.Plan.Summary.MaxDelayDays,
			},
			"entries": entries,
		}
	}

	for _, issue := range data.Issues {
		items["issues"] = append(items["issues"].([]map[string]any), mapIssueTemplateItem(issue))
	}
	for _, session := range data.Sessions {
		endTime := ""
		if session.EndTime != nil {
			endTime = *session.EndTime
		}
		items["sessions"] = append(items["sessions"].([]map[string]any), map[string]any{
			"id":              session.ID,
			"issueId":         session.IssueID,
			"issueTitle":      session.IssueTitle,
			"repoId":          session.RepoID,
			"repoName":        session.RepoName,
			"streamId":        session.StreamID,
			"streamName":      session.StreamName,
			"startTime":       session.StartTime,
			"endTime":         endTime,
			"durationSeconds": session.DurationSeconds,
			"summary":         sessionSummary(session),
		})
	}
	for _, habit := range data.Habits {
		items["habits"] = append(items["habits"].([]map[string]any), mapHabitTemplateItem(habit))
	}
	return items
}

func mapSummary(summary sharedtypes.DailyIssueSummary) map[string]any {
	return map[string]any{
		"date":                  summary.Date,
		"totalIssues":           summary.TotalIssues,
		"totalEstimatedMinutes": summary.TotalEstimatedMinutes,
		"completedIssues":       summary.CompletedIssues,
		"abandonedIssues":       summary.AbandonedIssues,
		"workedSeconds":         summary.WorkedSeconds,
		"workedTime":            formatDurationHMS(summary.WorkedSeconds),
	}
}

func buildSummaryMap(data *sharedtypes.DailyReportData) map[string]any {
	summary := mapSummary(data.Summary)
	issueDoneCount := 0
	issueActiveCount := 0
	issueBlockedCount := 0
	issueAbandonedCount := 0
	for _, issue := range data.Issues {
		switch categorizeIssue(issue.Status) {
		case "completed":
			issueDoneCount++
		case "blocked":
			issueBlockedCount++
		case "abandoned":
			issueAbandonedCount++
		default:
			issueActiveCount++
		}
	}
	habitsCompletedCount := 0
	habitsPendingCount := 0
	for _, habit := range data.Habits {
		if isHabitCompleted(habit.Status) {
			habitsCompletedCount++
			continue
		}
		habitsPendingCount++
	}
	summary["issueDoneCount"] = issueDoneCount
	summary["issueActiveCount"] = issueActiveCount
	summary["issueBlockedCount"] = issueBlockedCount
	summary["issueAbandonedCount"] = issueAbandonedCount
	summary["issueCompletion"] = formatRatio(issueDoneCount, data.Summary.TotalIssues)
	summary["estimatedTime"] = formatDurationHMS(data.Summary.TotalEstimatedMinutes * 60)
	summary["workedEstimate"] = formatWorkedEstimate(data.Summary.WorkedSeconds, intPtr(data.Summary.TotalEstimatedMinutes))
	summary["varianceTime"] = formatVarianceTime(data.Summary.WorkedSeconds, data.Summary.TotalEstimatedMinutes)
	summary["habitsDueCount"] = len(data.Habits)
	summary["habitsCompletedCount"] = habitsCompletedCount
	summary["habitsPendingCount"] = habitsPendingCount
	summary["habitCompletion"] = formatRatio(habitsCompletedCount, len(data.Habits))
	if data.Plan != nil {
		summary["plannedCount"] = data.Plan.Summary.PlannedCount
		summary["planCompletedCount"] = data.Plan.Summary.CompletedCount
		summary["planFailedCount"] = data.Plan.Summary.FailedCount
		summary["planAbandonedCount"] = data.Plan.Summary.AbandonedCount
		summary["planPendingRollbackCount"] = data.Plan.Summary.PendingRollbackCount
		summary["accountabilityScore"] = data.Plan.Summary.AccountabilityScore
		summary["backlogPressure"] = data.Plan.Summary.BacklogPressure
		summary["delayedIssueCount"] = data.Plan.Summary.DelayedIssueCount
		summary["highRiskIssueCount"] = data.Plan.Summary.HighRiskIssueCount
		summary["avgDelayDays"] = data.Plan.Summary.AvgDelayDays
		summary["maxDelayDays"] = data.Plan.Summary.MaxDelayDays
	}
	return summary
}

func mapCheckIn(checkIn sharedtypes.DailyCheckIn) map[string]any {
	return map[string]any{
		"date":              checkIn.Date,
		"mood":              checkIn.Mood,
		"energy":            checkIn.Energy,
		"sleepHours":        checkIn.SleepHours,
		"sleepScore":        checkIn.SleepScore,
		"screenTimeMinutes": checkIn.ScreenTimeMinutes,
		"screenTime":        formatDurationMinutesHMS(checkIn.ScreenTimeMinutes),
		"notes":             checkIn.Notes,
	}
}

func mapMetrics(metrics sharedtypes.DailyMetricsDay) map[string]any {
	out := map[string]any{
		"date":                  metrics.Date,
		"workedSeconds":         metrics.WorkedSeconds,
		"workedTime":            formatDurationHMS(metrics.WorkedSeconds),
		"restSeconds":           metrics.RestSeconds,
		"restTime":              formatDurationHMS(metrics.RestSeconds),
		"sessionCount":          metrics.SessionCount,
		"totalIssues":           metrics.TotalIssues,
		"completedIssues":       metrics.CompletedIssues,
		"abandonedIssues":       metrics.AbandonedIssues,
		"totalEstimatedMinutes": metrics.TotalEstimatedMinutes,
		"burnout":               nil,
	}
	if metrics.Burnout != nil {
		out["burnout"] = map[string]any{
			"score": metrics.Burnout.Score,
			"level": string(metrics.Burnout.Level),
		}
	}
	return out
}

func mapRollup(rollup sharedtypes.MetricsRollup) map[string]any {
	return map[string]any{
		"startDate":     rollup.StartDate,
		"endDate":       rollup.EndDate,
		"days":          rollup.Days,
		"checkInDays":   rollup.CheckInDays,
		"focusDays":     rollup.FocusDays,
		"workedSeconds": rollup.WorkedSeconds,
		"workedTime":    formatDurationHMS(rollup.WorkedSeconds),
		"restSeconds":   rollup.RestSeconds,
		"restTime":      formatDurationHMS(rollup.RestSeconds),
		"sessionCount":  rollup.SessionCount,
	}
}

func groupIssuesByRepo(issues []sharedtypes.DailyReportIssue) []map[string]any {
	type streamGroup struct {
		repoID     int64
		streamID   int64
		streamName string
		completed  []map[string]any
		active     []map[string]any
		attention  []map[string]any
	}

	repoGroups := make(map[int64]map[string]any)
	streamGroups := make(map[string]*streamGroup)
	repoOrder := make([]int64, 0)

	for _, issue := range issues {
		repo, ok := repoGroups[issue.RepoID]
		if !ok {
			repo = map[string]any{
				"id":      issue.RepoID,
				"name":    issue.RepoName,
				"streams": make([]map[string]any, 0),
			}
			repoGroups[issue.RepoID] = repo
			repoOrder = append(repoOrder, issue.RepoID)
		}

		streamKey := fmt.Sprintf("%d:%d", issue.RepoID, issue.StreamID)
		group, ok := streamGroups[streamKey]
		if !ok {
			group = &streamGroup{
				repoID:     issue.RepoID,
				streamID:   issue.StreamID,
				streamName: issue.StreamName,
				completed:  make([]map[string]any, 0),
				active:     make([]map[string]any, 0),
				attention:  make([]map[string]any, 0),
			}
			streamGroups[streamKey] = group
		}
		item := mapIssueTemplateItem(issue)
		switch categorizeIssue(issue.Status) {
		case "completed":
			group.completed = append(group.completed, item)
		case "blocked", "abandoned":
			group.attention = append(group.attention, item)
		default:
			group.active = append(group.active, item)
		}
		_ = repo
	}

	sort.Slice(repoOrder, func(i, j int) bool {
		return repoGroups[repoOrder[i]]["name"].(string) < repoGroups[repoOrder[j]]["name"].(string)
	})

	for _, repoID := range repoOrder {
		repo := repoGroups[repoID]
		streams := make([]map[string]any, 0)
		for _, group := range streamGroups {
			if group.repoID != repoID {
				continue
			}
			streams = append(streams, map[string]any{
				"id":              group.streamID,
				"name":            group.streamName,
				"completedIssues": group.completed,
				"activeIssues":    group.active,
				"attentionIssues": group.attention,
			})
		}
		sort.Slice(streams, func(i, j int) bool {
			return streams[i]["name"].(string) < streams[j]["name"].(string)
		})
		repo["streams"] = streams
	}

	grouped := make([]map[string]any, 0, len(repoOrder))
	for _, repoID := range repoOrder {
		grouped = append(grouped, repoGroups[repoID])
	}
	return grouped
}

func groupHabitsByRepo(habits []sharedtypes.HabitDailyItem) []map[string]any {
	type streamGroup struct {
		repoID     int64
		streamID   int64
		streamName string
		completed  []map[string]any
		pending    []map[string]any
	}

	repoGroups := make(map[int64]map[string]any)
	streamGroups := make(map[string]*streamGroup)
	repoOrder := make([]int64, 0)

	for _, habit := range habits {
		repo, ok := repoGroups[habit.RepoID]
		if !ok {
			repo = map[string]any{
				"id":      habit.RepoID,
				"name":    habit.RepoName,
				"streams": make([]map[string]any, 0),
			}
			repoGroups[habit.RepoID] = repo
			repoOrder = append(repoOrder, habit.RepoID)
		}

		streamKey := fmt.Sprintf("%d:%d", habit.RepoID, habit.StreamID)
		group, ok := streamGroups[streamKey]
		if !ok {
			group = &streamGroup{
				repoID:     habit.RepoID,
				streamID:   habit.StreamID,
				streamName: habit.StreamName,
				completed:  make([]map[string]any, 0),
				pending:    make([]map[string]any, 0),
			}
			streamGroups[streamKey] = group
		}
		item := mapHabitTemplateItem(habit)
		if isHabitCompleted(habit.Status) {
			group.completed = append(group.completed, item)
		} else {
			group.pending = append(group.pending, item)
		}
		_ = repo
	}

	sort.Slice(repoOrder, func(i, j int) bool {
		return repoGroups[repoOrder[i]]["name"].(string) < repoGroups[repoOrder[j]]["name"].(string)
	})

	for _, repoID := range repoOrder {
		repo := repoGroups[repoID]
		streams := make([]map[string]any, 0)
		for _, group := range streamGroups {
			if group.repoID != repoID {
				continue
			}
			streams = append(streams, map[string]any{
				"id":              group.streamID,
				"name":            group.streamName,
				"completedHabits": group.completed,
				"pendingHabits":   group.pending,
			})
		}
		sort.Slice(streams, func(i, j int) bool {
			return streams[i]["name"].(string) < streams[j]["name"].(string)
		})
		repo["streams"] = streams
	}

	grouped := make([]map[string]any, 0, len(repoOrder))
	for _, repoID := range repoOrder {
		grouped = append(grouped, repoGroups[repoID])
	}
	return grouped
}

func formatDurationHMS(totalSeconds int) string {
	if totalSeconds < 0 {
		totalSeconds = 0
	}
	duration := time.Duration(totalSeconds) * time.Second
	hours := int(duration / time.Hour)
	minutes := int(duration%time.Hour) / int(time.Minute)
	seconds := int(duration%time.Minute) / int(time.Second)
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}

func formatDurationMinutesHMS(minutes *int) string {
	if minutes == nil {
		return ""
	}
	return formatDurationHMS(*minutes * 60)
}

func formatWorkedEstimate(workedSeconds int, estimateMinutes *int) string {
	if estimateMinutes == nil {
		return formatDurationHMS(workedSeconds)
	}
	return fmt.Sprintf("%s / %s", formatDurationHMS(workedSeconds), formatDurationMinutesHMS(estimateMinutes))
}

func formatVarianceTime(workedSeconds int, estimateMinutes int) string {
	variance := workedSeconds - (estimateMinutes * 60)
	if variance == 0 {
		return "00:00:00"
	}
	sign := "+"
	if variance < 0 {
		sign = "-"
		variance = -variance
	}
	return sign + formatDurationHMS(variance)
}

func formatRatio(done int, total int) string {
	if total <= 0 {
		return "0 / 0"
	}
	return fmt.Sprintf("%d / %d", done, total)
}

func formatStatusLabel(value string, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	parts := strings.Split(value, "_")
	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, " ")
}

func mapIssueTemplateItem(issue sharedtypes.DailyReportIssue) map[string]any {
	return map[string]any{
		"id":                       issue.ID,
		"title":                    issue.Title,
		"repoId":                   issue.RepoID,
		"repoName":                 issue.RepoName,
		"streamId":                 issue.StreamID,
		"streamName":               issue.StreamName,
		"status":                   formatStatusLabel(string(issue.Status), ""),
		"estimateMinutes":          issue.EstimateMinutes,
		"estimateTime":             formatDurationMinutesHMS(issue.EstimateMinutes),
		"workedSeconds":            issue.WorkedSeconds,
		"workedTime":               formatDurationHMS(issue.WorkedSeconds),
		"workedEstimate":           formatWorkedEstimate(issue.WorkedSeconds, issue.EstimateMinutes),
		"planStatus":               issue.PlanStatus,
		"planCommittedAt":          issue.PlanCommittedAt,
		"planResolvedAt":           issue.PlanResolvedAt,
		"planFailureReason":        issue.PlanFailureReason,
		"planPendingFailureAt":     issue.PlanPendingFailureAt,
		"planPendingFailureReason": issue.PlanPendingFailureReason,
		"planBaselineDate":         issue.PlanBaselineDate,
		"planCurrentPlannedDate":   issue.PlanCurrentPlannedDate,
		"planPostponeCount":        issue.PlanPostponeCount,
		"planCurrentDelayedDays":   issue.PlanCurrentDelayedDays,
		"planMaxDelayedDays":       issue.PlanMaxDelayedDays,
		"planFailScore":            issue.PlanFailScore,
	}
}

func mapHabitTemplateItem(habit sharedtypes.HabitDailyItem) map[string]any {
	return map[string]any{
		"id":              habit.ID,
		"name":            habit.Name,
		"repoId":          habit.RepoID,
		"repoName":        habit.RepoName,
		"streamId":        habit.StreamID,
		"streamName":      habit.StreamName,
		"status":          formatStatusLabel(string(habit.Status), "Pending"),
		"durationMinutes": habit.DurationMinutes,
		"durationTime":    formatDurationMinutesHMS(habit.DurationMinutes),
		"notes":           habit.Notes,
	}
}

func categorizeIssue(status sharedtypes.IssueStatus) string {
	switch status {
	case sharedtypes.IssueStatusDone:
		return "completed"
	case sharedtypes.IssueStatusBlocked:
		return "blocked"
	case sharedtypes.IssueStatusAbandoned:
		return "abandoned"
	default:
		return "active"
	}
}

func isHabitCompleted(status sharedtypes.HabitCompletionStatus) bool {
	return status == sharedtypes.HabitCompletionStatusCompleted
}

func buildHighlights(data *sharedtypes.DailyReportData) []string {
	highlights := make([]string, 0, 6)
	completedIssues := collectIssueTitlesByStatus(data.Issues, sharedtypes.IssueStatusDone, 3)
	if len(completedIssues) > 0 {
		highlights = append(highlights, fmt.Sprintf("Completed issues: %s", strings.Join(completedIssues, ", ")))
	}
	completedHabits := collectCompletedHabitNames(data.Habits, 3)
	if len(completedHabits) > 0 {
		highlights = append(highlights, fmt.Sprintf("Completed habits: %s", strings.Join(completedHabits, ", ")))
	}
	if data.CheckIn != nil && data.CheckIn.Mood >= 4 && data.CheckIn.Energy >= 4 {
		highlights = append(highlights, fmt.Sprintf("Strong check-in: mood %d/5, energy %d/5", data.CheckIn.Mood, data.CheckIn.Energy))
	}
	if data.Metrics != nil && data.Metrics.Burnout != nil && data.Metrics.Burnout.Level == sharedtypes.BurnoutLevelLow {
		highlights = append(highlights, fmt.Sprintf("Burnout stayed low (%d)", data.Metrics.Burnout.Score))
	}
	return highlights
}

func buildRisks(data *sharedtypes.DailyReportData) []string {
	risks := make([]string, 0, 8)
	blocked := collectIssueTitlesByStatus(data.Issues, sharedtypes.IssueStatusBlocked, 3)
	if len(blocked) > 0 {
		risks = append(risks, fmt.Sprintf("Blocked issues: %s", strings.Join(blocked, ", ")))
	}
	abandoned := collectIssueTitlesByStatus(data.Issues, sharedtypes.IssueStatusAbandoned, 3)
	if len(abandoned) > 0 {
		risks = append(risks, fmt.Sprintf("Abandoned issues: %s", strings.Join(abandoned, ", ")))
	}
	pendingHabits := collectPendingHabitNames(data.Habits, 3)
	if len(pendingHabits) > 0 {
		risks = append(risks, fmt.Sprintf("Pending habits: %s", strings.Join(pendingHabits, ", ")))
	}
	if data.CheckIn != nil {
		if data.CheckIn.Energy <= 2 {
			risks = append(risks, fmt.Sprintf("Low energy: %d/5", data.CheckIn.Energy))
		}
		if data.CheckIn.Mood <= 2 {
			risks = append(risks, fmt.Sprintf("Low mood: %d/5", data.CheckIn.Mood))
		}
		if data.CheckIn.SleepHours != nil && *data.CheckIn.SleepHours < 7 {
			risks = append(risks, fmt.Sprintf("Sleep was low: %.1fh", *data.CheckIn.SleepHours))
		}
		if data.CheckIn.ScreenTimeMinutes != nil && *data.CheckIn.ScreenTimeMinutes >= 360 {
			risks = append(risks, fmt.Sprintf("High screen time: %s", formatDurationMinutesHMS(data.CheckIn.ScreenTimeMinutes)))
		}
	}
	if data.Metrics != nil && data.Metrics.Burnout != nil {
		switch data.Metrics.Burnout.Level {
		case sharedtypes.BurnoutLevelHigh:
			risks = append(risks, fmt.Sprintf("Burnout is high (%d)", data.Metrics.Burnout.Score))
		case sharedtypes.BurnoutLevelGuarded:
			risks = append(risks, fmt.Sprintf("Burnout is guarded (%d)", data.Metrics.Burnout.Score))
		}
	}
	return risks
}

func deriveDayHealth(data *sharedtypes.DailyReportData, summary map[string]any, risks []string) string {
	if data.Metrics != nil && data.Metrics.Burnout != nil && data.Metrics.Burnout.Level == sharedtypes.BurnoutLevelHigh {
		return "Overloaded"
	}
	if len(risks) >= 3 {
		return "Overloaded"
	}
	if len(risks) > 0 {
		return "Mixed"
	}
	if completed, ok := summary["issueDoneCount"].(int); ok && completed > 0 {
		return "Strong"
	}
	if habits, ok := summary["habitsCompletedCount"].(int); ok && habits > 0 {
		return "Strong"
	}
	return "Steady"
}

func collectIssueTitlesByStatus(issues []sharedtypes.DailyReportIssue, status sharedtypes.IssueStatus, limit int) []string {
	items := make([]string, 0, limit)
	for _, issue := range issues {
		if issue.Status != status {
			continue
		}
		items = append(items, "#"+fmt.Sprint(issue.ID)+" "+issue.Title)
		if len(items) == limit {
			break
		}
	}
	return items
}

func collectCompletedHabitNames(habits []sharedtypes.HabitDailyItem, limit int) []string {
	items := make([]string, 0, limit)
	for _, habit := range habits {
		if !isHabitCompleted(habit.Status) {
			continue
		}
		items = append(items, habit.Name)
		if len(items) == limit {
			break
		}
	}
	return items
}

func collectPendingHabitNames(habits []sharedtypes.HabitDailyItem, limit int) []string {
	items := make([]string, 0, limit)
	for _, habit := range habits {
		if isHabitCompleted(habit.Status) {
			continue
		}
		items = append(items, habit.Name)
		if len(items) == limit {
			break
		}
	}
	return items
}

func intPtr(value int) *int {
	return &value
}

func sessionSummary(session sharedtypes.DailyReportSession) string {
	if session.ParsedNotes != nil {
		if text := strings.TrimSpace(session.ParsedNotes[sharedtypes.SessionNoteSectionCommit]); text != "" {
			return text
		}
		if text := strings.TrimSpace(session.ParsedNotes[sharedtypes.SessionNoteSectionNotes]); text != "" {
			return text
		}
	}
	if session.Notes != nil && strings.TrimSpace(*session.Notes) != "" {
		return strings.TrimSpace(*session.Notes)
	}
	return fmt.Sprintf("Issue #%d", session.IssueID)
}
