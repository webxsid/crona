package commands

import (
	"context"
	"fmt"
	"math"
	"time"

	"crona/kernel/internal/core"
	sharedtypes "crona/shared/types"
)

const defaultDailyPlanRollbackMinutes = 5

func GetDailyPlan(
	ctx context.Context,
	c *core.Context,
	date string,
) (*sharedtypes.DailyPlan, error) {
	if err := finalizeExpiredDailyPlanFailures(ctx, c, c.Now()); err != nil {
		return nil, err
	}
	plan, err := c.DailyPlans.GetByDate(ctx, c.UserID, date)
	if err != nil {
		return nil, err
	}
	if plan == nil {
		plan = &sharedtypes.DailyPlan{Date: date}
	}
	return scoreDailyPlan(ctx, c, plan)
}

func finalizeExpiredDailyPlanFailures(ctx context.Context, c *core.Context, now string) error {
	settings, err := c.CoreSettings.Get(ctx, c.UserID)
	if err != nil {
		return err
	}
	settings = settingsWithLiveAwayDate(settings, c.CurrentDate())
	cutoff, err := time.Parse(time.RFC3339, now)
	if err != nil {
		return nil
	}
	cutoff = cutoff.Add(-dailyPlanRollbackWindow(settings))
	entries, err := c.DailyPlans.ListPendingFailuresBefore(
		ctx,
		c.UserID,
		cutoff.Format(time.RFC3339),
	)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		reason := entry.PendingFailureReason
		if reason == nil {
			continue
		}
		if err := c.DailyPlans.ResolveEntry(ctx, entry.ID, now, sharedtypes.DailyPlanEntryStatusFailed, reason); err != nil {
			return err
		}
		if err := c.DailyPlans.AppendEvent(ctx, entry.ID, c.UserID, c.DeviceID, now, sharedtypes.DailyPlanEventFailed, reason, map[string]any{
			"issueId": entry.IssueID,
			"date":    entry.Date,
			"reason":  *reason,
		}); err != nil {
			return err
		}
	}
	return nil
}

func dailyPlanRollbackWindow(settings *sharedtypes.CoreSettings) time.Duration {
	minutes := defaultDailyPlanRollbackMinutes
	if settings != nil && settings.DailyPlanRollbackMins > 0 {
		minutes = settings.DailyPlanRollbackMins
	}
	return time.Duration(minutes) * time.Minute
}

func commitIssueToDailyPlan(
	ctx context.Context,
	c *core.Context,
	issueID int64,
	date, now string,
) error {
	return commitIssueToDailyPlanWithChain(ctx, c, issueID, date, now, date, date, 0, 0)
}

func commitIssueToDailyPlanWithChain(
	ctx context.Context,
	c *core.Context,
	issueID int64,
	date, now, baselineDate, currentPlannedDate string,
	postponeCount, maxDelayedDays int,
) error {
	entry, action, err := c.DailyPlans.UpsertCommittedEntryWithChain(
		ctx,
		c.UserID,
		date,
		"todo_for_date",
		now,
		issueID,
		baselineDate,
		currentPlannedDate,
		postponeCount,
		maxDelayedDays,
	)
	if err != nil || entry == nil {
		return err
	}
	if action == "restored" {
		if err := c.DailyPlans.AppendEvent(ctx, entry.ID, c.UserID, c.DeviceID, now, sharedtypes.DailyPlanEventFailureReverted, entry.PendingFailureReason, map[string]any{
			"issueId": issueID,
			"date":    date,
		}); err != nil {
			return err
		}
		return nil
	}
	if action == "unchanged" {
		return nil
	}
	return c.DailyPlans.AppendEvent(
		ctx,
		entry.ID,
		c.UserID,
		c.DeviceID,
		now,
		sharedtypes.DailyPlanEventCommitted,
		nil,
		map[string]any{
			"issueId": issueID,
			"date":    date,
		},
	)
}

func markDailyPlanPendingFailure(
	ctx context.Context,
	c *core.Context,
	issueID int64,
	date, now string,
	reason sharedtypes.DailyPlanFailureReason,
	eventType sharedtypes.DailyPlanEventType,
) error {
	entry, err := c.DailyPlans.MarkPendingFailure(ctx, c.UserID, date, now, issueID, reason)
	if err != nil || entry == nil {
		return err
	}
	return c.DailyPlans.AppendEvent(
		ctx,
		entry.ID,
		c.UserID,
		c.DeviceID,
		now,
		eventType,
		&reason,
		map[string]any{
			"issueId": issueID,
			"date":    date,
			"reason":  reason,
		},
	)
}

func resolveDailyPlanEntry(
	ctx context.Context,
	c *core.Context,
	issueID int64,
	date, now string,
	status sharedtypes.DailyPlanEntryStatus,
	eventType sharedtypes.DailyPlanEventType,
	reason *sharedtypes.DailyPlanFailureReason,
) error {
	entry, err := c.DailyPlans.GetEntry(ctx, c.UserID, date, issueID)
	if err != nil || entry == nil {
		return err
	}
	if err := c.DailyPlans.ResolveEntry(ctx, entry.ID, now, status, reason); err != nil {
		return err
	}
	return c.DailyPlans.AppendEvent(
		ctx,
		entry.ID,
		c.UserID,
		c.DeviceID,
		now,
		eventType,
		reason,
		map[string]any{
			"issueId": issueID,
			"date":    date,
			"status":  status,
		},
	)
}

func issueCommittedDate(issue *sharedtypes.Issue) string {
	if issue == nil || issue.TodoForDate == nil {
		return ""
	}
	return *issue.TodoForDate
}

func mustDayPrefix(timestamp string) string {
	if len(timestamp) >= 10 {
		return timestamp[:10]
	}
	return ""
}

func entryResolveDate(issue *sharedtypes.Issue, now string) string {
	if issue != nil {
		if issue.CompletedAt != nil && *issue.CompletedAt != "" {
			return mustDayPrefix(*issue.CompletedAt)
		}
		if issue.AbandonedAt != nil && *issue.AbandonedAt != "" {
			return mustDayPrefix(*issue.AbandonedAt)
		}
	}
	return mustDayPrefix(now)
}

func ensureDailyPlanDate(date string) error {
	if len(date) != len("2006-01-02") {
		return fmt.Errorf("invalid daily plan date")
	}
	_, err := time.Parse("2006-01-02", date)
	return err
}

func scoreDailyPlan(
	ctx context.Context,
	c *core.Context,
	plan *sharedtypes.DailyPlan,
) (*sharedtypes.DailyPlan, error) {
	settings, err := c.CoreSettings.Get(ctx, c.UserID)
	if err != nil {
		return nil, err
	}
	settings = settingsWithLiveAwayDate(settings, plan.Date)
	for i := range plan.Entries {
		applyDailyPlanScore(&plan.Entries[i], settings)
	}
	active, err := c.DailyPlans.ListActiveEntries(ctx, c.UserID)
	if err != nil {
		return nil, err
	}
	for i := range active {
		applyDailyPlanScore(&active[i], settings)
	}
	plan.Summary = buildDailyPlanSummary(plan.Entries, active, settings)
	activeIDs := make(map[string]bool, len(active))
	for _, entry := range active {
		activeIDs[entry.ID] = true
	}
	for i := range plan.Entries {
		entry := &plan.Entries[i]
		entry.Penalty = buildDailyPlanEntryPenalty(
			*entry,
			activeIDs[entry.ID],
			plan.Summary,
			settings,
		)
	}
	return plan, nil
}

func buildDailyPlanSummary(
	entries []sharedtypes.DailyPlanEntry,
	active []sharedtypes.DailyPlanEntry,
	settings *sharedtypes.CoreSettings,
) sharedtypes.DailyPlanAccountabilitySummary {
	summary := sharedtypes.DailyPlanAccountabilitySummary{
		HighRiskThreshold: planHighRiskThreshold(settings),
	}
	for _, entry := range entries {
		summary.PlannedCount++
		if entry.PendingFailureAt != nil {
			summary.PendingRollbackCount++
		}
		switch entry.Status {
		case sharedtypes.DailyPlanEntryStatusCompleted:
			summary.CompletedCount++
		case sharedtypes.DailyPlanEntryStatusFailed:
			summary.FailedCount++
		case sharedtypes.DailyPlanEntryStatusAbandoned:
			summary.AbandonedCount++
		}
	}
	totalDelayDays := 0
	for _, entry := range active {
		summary.AccountabilityScore += entry.FailScore
		summary.BacklogPressure += math.Log2(1 + float64(entry.CurrentDelayedDays))
		if entry.CurrentDelayedDays > 0 {
			summary.DelayedIssueCount++
			totalDelayDays += entry.CurrentDelayedDays
			if entry.CurrentDelayedDays > summary.MaxDelayDays {
				summary.MaxDelayDays = entry.CurrentDelayedDays
			}
		}
		if entry.FailScore >= summary.HighRiskThreshold {
			summary.HighRiskIssueCount++
		}
	}
	if summary.DelayedIssueCount > 0 {
		summary.AvgDelayDays = float64(totalDelayDays) / float64(summary.DelayedIssueCount)
	}
	return summary
}

func applyDailyPlanScore(entry *sharedtypes.DailyPlanEntry, settings *sharedtypes.CoreSettings) {
	if entry == nil {
		return
	}
	baselineDate := entry.BaselineDate
	if baselineDate == "" {
		baselineDate = entry.Date
		entry.BaselineDate = baselineDate
	}
	currentPlannedDate := entry.CurrentPlannedDate
	if currentPlannedDate == "" {
		currentPlannedDate = entry.Date
		entry.CurrentPlannedDate = currentPlannedDate
	}
	entry.CurrentDelayedDays = protectedDelayDays(baselineDate, currentPlannedDate, settings)
	if entry.MaxDelayedDays < entry.CurrentDelayedDays {
		entry.MaxDelayedDays = entry.CurrentDelayedDays
	}
	entry.FailScore = dailyPlanDelayComponent(*entry) + dailyPlanFrequencyComponent(*entry)
}

func dailyPlanDelayComponent(entry sharedtypes.DailyPlanEntry) float64 {
	return math.Log2(1 + float64(entry.CurrentDelayedDays))
}

func dailyPlanFrequencyComponent(entry sharedtypes.DailyPlanEntry) float64 {
	return 0.5 * math.Log2(1+float64(max(0, entry.PostponeCount-1)))
}

func planHighRiskThreshold(settings *sharedtypes.CoreSettings) float64 {
	if settings == nil {
		return float64(sharedtypes.DefaultPlanHighRiskTenths) / 10
	}
	return settings.PlanHighRiskThreshold()
}

// buildDailyPlanEntryPenalty explains a scored entry's penalty. Active entries
// also get their share of the summary totals and the done/postpone projections;
// resolved entries are historical and keep only their final values.
func buildDailyPlanEntryPenalty(
	entry sharedtypes.DailyPlanEntry,
	active bool,
	summary sharedtypes.DailyPlanAccountabilitySummary,
	settings *sharedtypes.CoreSettings,
) *sharedtypes.DailyPlanEntryPenalty {
	threshold := summary.HighRiskThreshold
	penalty := &sharedtypes.DailyPlanEntryPenalty{
		ProtectedDaysSkipped: protectedDaysSkipped(
			entry.BaselineDate,
			entry.CurrentPlannedDate,
			settings,
		),
		DelayComponent:     dailyPlanDelayComponent(entry),
		FrequencyComponent: dailyPlanFrequencyComponent(entry),
		HighRisk:           entry.FailScore >= threshold,
		HighRiskThreshold:  threshold,
		Historical:         !active,
	}
	if !active {
		return penalty
	}
	share := sharedtypes.DailyPlanPenaltyShare{
		Accountability:     entry.FailScore,
		AccountabilityPct:  safeRatio(entry.FailScore, summary.AccountabilityScore),
		BacklogPressure:    penalty.DelayComponent,
		BacklogPressurePct: safeRatio(penalty.DelayComponent, summary.BacklogPressure),
	}
	penalty.Contribution = &share
	penalty.IfDoneToday = &sharedtypes.DailyPlanPenaltyShare{
		Accountability:     -share.Accountability,
		AccountabilityPct:  share.AccountabilityPct,
		BacklogPressure:    -share.BacklogPressure,
		BacklogPressurePct: share.BacklogPressurePct,
	}
	if next, err := time.Parse("2006-01-02", entry.CurrentPlannedDate); err == nil {
		nextDate := next.AddDate(0, 0, 1).Format("2006-01-02")
		projected := entry
		projected.BaselineDate, projected.CurrentPlannedDate, projected.PostponeCount, projected.MaxDelayedDays =
			nextDailyPlanChainState(
				&entry,
				entry.CurrentPlannedDate,
				nextDate,
				settings,
			)
		applyDailyPlanScore(&projected, settings)
		penalty.IfPostponedOneDay = &sharedtypes.DailyPlanPenaltyProjection{
			FailScore:           projected.FailScore,
			AccountabilityDelta: projected.FailScore - entry.FailScore,
			BecomesHighRisk:     entry.FailScore < threshold && projected.FailScore >= threshold,
		}
	}
	return penalty
}

func safeRatio(part, total float64) float64 {
	if total <= 0 {
		return 0
	}
	return part / total
}

// protectedDaysSkipped counts away days between the baseline and the current
// planned date that protectedDelayDays leaves out.
func protectedDaysSkipped(
	baselineDate, currentPlannedDate string,
	settings *sharedtypes.CoreSettings,
) int {
	if baselineDate == "" || currentPlannedDate <= baselineDate {
		return 0
	}
	start, err := time.Parse("2006-01-02", baselineDate)
	if err != nil {
		return 0
	}
	end, err := time.Parse("2006-01-02", currentPlannedDate)
	if err != nil {
		return 0
	}
	calendarDays := int(end.Sub(start).Hours() / 24)
	return max(0, calendarDays-protectedDelayDays(baselineDate, currentPlannedDate, settings))
}

func nextDailyPlanChainState(
	previous *sharedtypes.DailyPlanEntry,
	previousDate, nextDate string,
	settings *sharedtypes.CoreSettings,
) (string, string, int, int) {
	baselineDate := previousDate
	postponeCount := 0
	maxDelayedDays := 0
	if previous != nil {
		if previous.BaselineDate != "" {
			baselineDate = previous.BaselineDate
		}
		postponeCount = previous.PostponeCount
		maxDelayedDays = previous.MaxDelayedDays
	}
	if nextDate > previousDate {
		postponeCount++
	}
	currentDelayedDays := protectedDelayDays(baselineDate, nextDate, settings)
	if currentDelayedDays > maxDelayedDays {
		maxDelayedDays = currentDelayedDays
	}
	return baselineDate, nextDate, postponeCount, maxDelayedDays
}

func protectedDelayDays(
	baselineDate, currentPlannedDate string,
	settings *sharedtypes.CoreSettings,
) int {
	if baselineDate == "" || currentPlannedDate == "" || currentPlannedDate <= baselineDate {
		return 0
	}
	start, err := time.Parse("2006-01-02", baselineDate)
	if err != nil {
		return 0
	}
	end, err := time.Parse("2006-01-02", currentPlannedDate)
	if err != nil {
		return 0
	}
	days := 0
	for current := start.AddDate(0, 0, 1); !current.After(end); current = current.AddDate(0, 0, 1) {
		if isProtectedAccountabilityDay(current.Format("2006-01-02"), settings) {
			continue
		}
		days++
	}
	return days
}

func isProtectedAccountabilityDay(date string, settings *sharedtypes.CoreSettings) bool {
	if settings == nil {
		return false
	}
	return settings.IsAwayDate(date)
}
