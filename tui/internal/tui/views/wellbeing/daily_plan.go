package wellbeing

import (
	"fmt"

	"crona/tui/internal/api"
)

func dailyPlanCounts(plan *api.DailyPlan) (planned, completed, failed, abandoned, pending int) {
	if plan == nil {
		return 0, 0, 0, 0, 0
	}
	if plan.Summary.PlannedCount > 0 || plan.Summary.CompletedCount > 0 || plan.Summary.FailedCount > 0 || plan.Summary.AbandonedCount > 0 || plan.Summary.PendingRollbackCount > 0 {
		return plan.Summary.PlannedCount, plan.Summary.CompletedCount, plan.Summary.FailedCount, plan.Summary.AbandonedCount, plan.Summary.PendingRollbackCount
	}
	for _, entry := range plan.Entries {
		planned++
		if entry.PendingFailureAt != nil {
			pending++
		}
		switch entry.Status {
		case "completed":
			completed++
		case "failed":
			failed++
		case "abandoned":
			abandoned++
		}
	}
	return planned, completed, failed, abandoned, pending
}

func dailyPlanSignals(plan *api.DailyPlan) (score, pressure float64, delayed, highRisk int) {
	if plan == nil {
		return 0, 0, 0, 0
	}
	return plan.Summary.AccountabilityScore, plan.Summary.BacklogPressure, plan.Summary.DelayedIssueCount, plan.Summary.HighRiskIssueCount
}

func recentPlanFailureLines(plan *api.DailyPlan, limit int) []string {
	if plan == nil || limit <= 0 {
		return nil
	}
	lines := make([]string, 0, limit)
	for _, entry := range plan.Entries {
		if entry.Status == "failed" && entry.FailureReason != nil {
			lines = append(lines, fmt.Sprintf("- issue #%d marked failed (%s)", entry.IssueID, *entry.FailureReason))
			if len(lines) >= limit {
				break
			}
		}
	}
	return lines
}
