package export

import (
	"strings"
	"testing"

	sharedtypes "crona/shared/types"
)

func TestBuildTemplateDataMapIncludesPlanAccountabilitySummary(t *testing.T) {
	data := &sharedtypes.DailyReportData{
		Date:        "2026-04-08",
		GeneratedAt: "2026-04-08T10:00:00Z",
		Summary: sharedtypes.DailyIssueSummary{
			Date:                  "2026-04-08",
			TotalIssues:           3,
			TotalEstimatedMinutes: 180,
			WorkedSeconds:         7200,
		},
		Plan: &sharedtypes.DailyPlan{
			ID:   "plan-1",
			Date: "2026-04-08",
			Summary: sharedtypes.DailyPlanAccountabilitySummary{
				PlannedCount:         3,
				CompletedCount:       1,
				FailedCount:          1,
				AbandonedCount:       0,
				PendingRollbackCount: 1,
				AccountabilityScore:  74.5,
				BacklogPressure:      2.1,
				DelayedIssueCount:    2,
				HighRiskIssueCount:   1,
				AvgDelayDays:         1.5,
				MaxDelayDays:         4,
			},
		},
		Issues: []sharedtypes.DailyReportIssue{
			{
				IssueWithMeta: sharedtypes.IssueWithMeta{
					Issue: sharedtypes.Issue{
						ID:              7,
						Title:           "Tighten exported plan narrative",
						Status:          sharedtypes.IssueStatusBlocked,
						EstimateMinutes: intPtr(90),
					},
					RepoName:   "Work",
					StreamName: "core",
				},
				WorkedSeconds:          1800,
				PlanStatus:             sharedtypes.DailyPlanEntryStatusFailed,
				PlanCurrentDelayedDays: 4,
				PlanFailScore:          0.8,
				PlanFailureReason:      planFailurePtr(sharedtypes.DailyPlanFailureReasonMissed),
			},
		},
	}

	items := buildTemplateDataMap(data)
	summary := items["summary"].(map[string]any)
	if got := summary["planFailedCount"]; got != 1 {
		t.Fatalf("expected planFailedCount 1, got %#v", got)
	}
	if got := summary["accountabilityScore"]; got != 74.5 {
		t.Fatalf("expected accountabilityScore 74.5, got %#v", got)
	}
	if got := summary["delayedIssueCount"]; got != 2 {
		t.Fatalf("expected delayedIssueCount 2, got %#v", got)
	}

	issues := items["issues"].([]map[string]any)
	if got := issues[0]["planStatus"]; got != sharedtypes.DailyPlanEntryStatusFailed {
		t.Fatalf("expected failed plan status, got %#v", got)
	}
	if got := issues[0]["planCurrentDelayedDays"]; got != 4 {
		t.Fatalf("expected delayed days 4, got %#v", got)
	}
}

func TestFallbackDailyReportTemplateRendersAccountabilityAndFailureDetails(t *testing.T) {
	data := &sharedtypes.DailyReportData{
		Date:        "2026-04-08",
		GeneratedAt: "2026-04-08T10:00:00Z",
		Summary: sharedtypes.DailyIssueSummary{
			Date:                  "2026-04-08",
			TotalIssues:           2,
			TotalEstimatedMinutes: 120,
			WorkedSeconds:         5400,
		},
		Plan: &sharedtypes.DailyPlan{
			ID:   "plan-1",
			Date: "2026-04-08",
			Summary: sharedtypes.DailyPlanAccountabilitySummary{
				PlannedCount:         2,
				CompletedCount:       1,
				FailedCount:          1,
				PendingRollbackCount: 1,
				AccountabilityScore:  68.5,
				BacklogPressure:      1.8,
				DelayedIssueCount:    1,
				HighRiskIssueCount:   1,
			},
		},
		Issues: []sharedtypes.DailyReportIssue{
			{
				IssueWithMeta: sharedtypes.IssueWithMeta{
					Issue: sharedtypes.Issue{
						ID:              9,
						Title:           "Recover failed plan item",
						Status:          sharedtypes.IssueStatusBlocked,
						EstimateMinutes: intPtr(60),
					},
					RepoName:   "Work",
					StreamName: "core",
				},
				WorkedSeconds:          1200,
				PlanStatus:             sharedtypes.DailyPlanEntryStatusFailed,
				PlanCurrentDelayedDays: 3,
				PlanFailureReason:      planFailurePtr(sharedtypes.DailyPlanFailureReasonMissed),
			},
		},
	}

	rendered, err := RenderTemplate(fallbackDailyReportTemplate, buildTemplateDataMap(data))
	if err != nil {
		t.Fatalf("render daily report template: %v", err)
	}

	for _, snippet := range []string{
		"## Plan Accountability",
		"- Failed: 1",
		"- Accountability score: 68.5",
		"- Delayed issues: 1",
		"- Plan status: failed",
		"- Failure reason: missed",
		"- Delayed by: 3 day(s)",
	} {
		if !strings.Contains(rendered, snippet) {
			t.Fatalf("expected rendered report to contain %q\n%s", snippet, rendered)
		}
	}
}

func planFailurePtr(reason sharedtypes.DailyPlanFailureReason) *sharedtypes.DailyPlanFailureReason {
	return &reason
}
