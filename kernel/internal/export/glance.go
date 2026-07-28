package export

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"crona/kernel/internal/core"
	corecommands "crona/kernel/internal/core/commands"
	"crona/kernel/internal/runtime"
	shareddatefmt "crona/shared/datefmt"
	shareddto "crona/shared/dto"
	sharedtypes "crona/shared/types"
)

func generateGlanceExport(
	ctx context.Context,
	c *core.Context,
	paths runtime.Paths,
	input shareddto.ExportReportRequest,
) (*sharedtypes.ExportReportResult, error) {
	settings, err := c.CoreSettings.Get(ctx, c.UserID)
	if err != nil {
		return nil, err
	}
	start, end, day := glancePeriod(input)
	reportKind := sharedtypes.ExportReportKindGlance
	if !day {
		reportKind = sharedtypes.ExportReportKindSummaryRange
	}
	if input.Kind == sharedtypes.ExportReportKindSummaryRange && day {
		return nil, errors.New("range summary requires distinct start and end dates")
	}
	startDate, startErr := time.Parse(time.DateOnly, start)
	endDate, endErr := time.Parse(time.DateOnly, end)
	if startErr != nil || endErr != nil {
		return nil, errors.New("summary dates must use YYYY-MM-DD")
	}
	if endDate.Before(startDate) {
		return nil, errors.New("summary end date must be on or after start date")
	}
	data, err := buildGlanceData(ctx, c, start, end, day, settings)
	if err != nil {
		return nil, err
	}
	baseName := "summary-" + end
	if start != end {
		baseName = fmt.Sprintf("summary-%s-to-%s", start, end)
	}
	return renderNarrativeReport(
		paths,
		reportKind,
		data,
		reportWriteSpec{
			Kind:      reportKind,
			Label:     "Summary",
			Date:      end,
			StartDate: start,
			EndDate:   end,
			Format:    normalizeNarrativeFormat(input.Format),
			BaseName:  baseName,
		},
		input.OutputMode,
		input.PresetID,
	)
}

func glancePeriod(input shareddto.ExportReportRequest) (start, end string, day bool) {
	start = strings.TrimSpace(input.Start)
	end = strings.TrimSpace(input.End)
	if start != "" && end != "" {
		return start, end, start == end
	}
	date := normalizeReportDate(input.Date)
	return date, date, true
}

func buildGlanceData(
	ctx context.Context,
	c *core.Context,
	start, end string,
	day bool,
	settings *sharedtypes.CoreSettings,
) (map[string]any, error) {
	if day {
		report, err := BuildDailyReportData(ctx, c, end)
		if err != nil {
			return nil, err
		}
		data := buildTemplateDataMap(report)
		planScore, planDone, planTotal := 0.0, 0, 0
		if report.Plan != nil {
			planScore = report.Plan.Summary.AccountabilityScore
			planDone = report.Plan.Summary.CompletedCount
			planTotal = report.Plan.Summary.PlannedCount
		}
		habitDone, habitFailed := 0, 0
		for _, habit := range report.Habits {
			if habit.Status == sharedtypes.HabitCompletionStatusFailed {
				habitFailed++
			} else if habit.Completed {
				habitDone++
			}
		}
		resolved := report.Summary.CompletedIssues + report.Summary.AbandonedIssues
		data["isRange"] = false
		data["periodLabel"] = shareddatefmt.FormatISODate(end, settings)
		data["focus"] = glanceMetric(formatDurationHMS(report.Summary.WorkedSeconds), report.Summary.WorkedSeconds, report.Summary.TotalEstimatedMinutes*60)
		data["issueMetric"] = glanceMetric(fmt.Sprintf("%d / %d", resolved, report.Summary.TotalIssues), resolved, report.Summary.TotalIssues)
		data["habitMetric"] = glanceMetric(fmt.Sprintf("%d / %d", habitDone, len(report.Habits)), habitDone, len(report.Habits))
		data["planMetric"] = glanceMetric(fmt.Sprintf("%.0f%%", planScore), int(planScore+.5), 100)
		data["outcomes"] = []map[string]any{
			glanceOutcome("Issue resolution", resolved, report.Summary.TotalIssues),
			glanceOutcome("Habit completion", habitDone, len(report.Habits)),
			glanceOutcome("Plan completion", planDone, planTotal),
		}
		return data, nil
	}

	days, err := corecommands.ComputeMetricsRange(ctx, c, start, end)
	if err != nil {
		return nil, err
	}
	rollup := corecommands.ComputeMetricsRollupFromDays(start, end, days)
	streaks := corecommands.ComputeMetricsStreaksFromDays(days, settings)
	window, err := corecommands.ComputeDashboardWindowSummary(
		ctx,
		c,
		shareddto.DashboardWindowQuery{Start: start, End: end},
	)
	if err != nil {
		return nil, err
	}
	totalIssues, maxWorked := 0, 1
	for _, item := range days {
		totalIssues += item.TotalIssues
		maxWorked = max(maxWorked, item.WorkedSeconds)
	}
	dayItems := make([]map[string]any, 0, len(days))
	for _, item := range days {
		mood, energy := "", ""
		if item.CheckIn != nil {
			mood = fmt.Sprintf("%d/5", item.CheckIn.Mood)
			energy = fmt.Sprintf("%d/5", item.CheckIn.Energy)
		}
		dayItems = append(dayItems, map[string]any{
			"displayDate":  shareddatefmt.FormatISODate(item.Date, settings),
			"workedTime":   formatDurationHMS(item.WorkedSeconds),
			"restTime":     formatDurationHMS(item.RestSeconds),
			"sessionCount": item.SessionCount,
			"issues":       fmt.Sprintf("%d/%d", item.CompletedIssues+item.AbandonedIssues, item.TotalIssues),
			"habits":       fmt.Sprintf("%d/%d", item.HabitCompletedCount, item.HabitDueCount),
			"mood":         mood,
			"energy":       energy,
			"bar":          glanceBar(item.WorkedSeconds, maxWorked),
			"percent":      glancePercent(item.WorkedSeconds, maxWorked),
		})
	}
	return map[string]any{
		"isRange":          true,
		"startDate":        start,
		"endDate":          end,
		"displayStartDate": shareddatefmt.FormatISODate(start, settings),
		"displayEndDate":   shareddatefmt.FormatISODate(end, settings),
		"periodLabel":      shareddatefmt.FormatISODate(start, settings) + " to " + shareddatefmt.FormatISODate(end, settings),
		"generatedAt":      time.Now().UTC().Format(time.RFC3339),
		"focus":            glanceMetric(formatDurationHMS(rollup.WorkedSeconds), rollup.FocusDays, max(1, rollup.Days)),
		"issueMetric":      glanceMetric(fmt.Sprintf("%d resolved", rollup.CompletedIssues+rollup.AbandonedIssues), rollup.CompletedIssues+rollup.AbandonedIssues, totalIssues),
		"habitMetric":      glanceMetric(fmt.Sprintf("%d / %d", rollup.HabitCompletedCount, rollup.HabitDueCount), rollup.HabitCompletedCount, rollup.HabitDueCount),
		"planMetric":       glanceMetric(fmt.Sprintf("%.0f%%", window.AccountabilityScore), int(window.AccountabilityScore+.5), 100),
		"outcomes": []map[string]any{
			glanceOutcome("Issue resolution", rollup.CompletedIssues+rollup.AbandonedIssues, totalIssues),
			glanceOutcome("Habit completion", rollup.HabitCompletedCount, rollup.HabitDueCount),
			glanceOutcome("Check-in coverage", rollup.CheckInDays, rollup.Days),
			glanceOutcome("Plan completion", window.CompletedCount, window.PlannedCount),
		},
		"streaks": map[string]any{
			"currentFocusDays":   streaks.CurrentFocusDays,
			"longestFocusDays":   streaks.LongestFocusDays,
			"currentCheckInDays": streaks.CurrentCheckInDays,
			"longestCheckInDays": streaks.LongestCheckInDays,
			"currentHabitDays":   streaks.CurrentHabitDays,
			"longestHabitDays":   streaks.LongestHabitDays,
		},
		"days": dayItems,
	}, nil
}

func glanceMetric(value string, current, total int) map[string]any {
	return map[string]any{
		"value":   value,
		"bar":     glanceBar(current, total),
		"percent": glancePercent(current, total),
	}
}

func glanceOutcome(label string, current, total int) map[string]any {
	return map[string]any{
		"label":   label,
		"value":   fmt.Sprintf("%d%%", glancePercent(current, total)),
		"bar":     glanceBar(current, total),
		"percent": glancePercent(current, total),
	}
}

func glancePercent(current, total int) int {
	if total <= 0 {
		return 0
	}
	return min(100, max(0, current*100/total))
}

func glanceBar(current, total int) string {
	const width = 20
	filled := glancePercent(current, total) * width / 100
	return strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
}
