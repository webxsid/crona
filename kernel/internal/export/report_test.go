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
				AccountabilityScore:  74.54,
				BacklogPressure:      2.14,
				DelayedIssueCount:    2,
				HighRiskIssueCount:   1,
				AvgDelayDays:         1.46,
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
				PlanFailScore:          0.84,
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
		t.Fatalf("expected rounded accountabilityScore 74.5, got %#v", got)
	}
	if got := summary["backlogPressure"]; got != 2.1 {
		t.Fatalf("expected rounded backlogPressure 2.1, got %#v", got)
	}
	if got := summary["avgDelayDays"]; got != 1.5 {
		t.Fatalf("expected rounded avgDelayDays 1.5, got %#v", got)
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
	if got := issues[0]["planFailScore"]; got != 0.8 {
		t.Fatalf("expected rounded fail score 0.8, got %#v", got)
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

func TestBuildTemplateDataMapSuppressesStalePlanRiskForResolvedIssues(t *testing.T) {
	pendingAt := "2026-04-08T08:00:00Z"
	data := &sharedtypes.DailyReportData{
		Date:        "2026-04-08",
		GeneratedAt: "2026-04-08T10:00:00Z",
		Summary: sharedtypes.DailyIssueSummary{
			Date:        "2026-04-08",
			TotalIssues: 1,
		},
		Issues: []sharedtypes.DailyReportIssue{
			{
				IssueWithMeta: sharedtypes.IssueWithMeta{
					Issue: sharedtypes.Issue{
						ID:     11,
						Title:  "Already shipped",
						Status: sharedtypes.IssueStatusDone,
					},
					RepoName:   "Work",
					StreamName: "core",
				},
				PlanStatus:               sharedtypes.DailyPlanEntryStatusCompleted,
				PlanCurrentDelayedDays:   5,
				PlanMaxDelayedDays:       7,
				PlanFailScore:            2.84,
				PlanPendingFailureAt:     &pendingAt,
				PlanPendingFailureReason: planFailurePtr(sharedtypes.DailyPlanFailureReasonMoved),
			},
		},
	}

	items := buildTemplateDataMap(data)
	issues := items["issues"].([]map[string]any)
	if got := issues[0]["planCurrentDelayedDays"]; got != 0 {
		t.Fatalf("expected resolved issue delayed days cleared, got %#v", got)
	}
	if got := issues[0]["planMaxDelayedDays"]; got != 0 {
		t.Fatalf("expected resolved issue max delayed days cleared, got %#v", got)
	}
	if got := issues[0]["planFailScore"]; got != 0.0 {
		t.Fatalf("expected resolved issue fail score cleared, got %#v", got)
	}
	if got := issues[0]["planPendingFailureAt"]; got != nil {
		pending, ok := got.(*string)
		if !ok || pending != nil {
			t.Fatalf("expected resolved issue pending failure timestamp cleared, got %#v", got)
		}
	}
	if got := issues[0]["planPendingFailureReason"]; got != nil {
		reason, ok := got.(*sharedtypes.DailyPlanFailureReason)
		if !ok || reason != nil {
			t.Fatalf("expected resolved issue pending failure reason cleared, got %#v", got)
		}
	}
}

func TestMarkdownReportTemplatesRenderYamlFrontmatter(t *testing.T) {
	data := &sharedtypes.DailyReportData{
		Date:        "2026-04-13",
		GeneratedAt: "2026-04-13T10:00:00Z",
		Summary: sharedtypes.DailyIssueSummary{
			Date:          "2026-04-13",
			TotalIssues:   1,
			WorkedSeconds: 1800,
		},
	}
	spec := reportWriteSpec{
		Kind:     sharedtypes.ExportReportKindDaily,
		Label:    "Daily Report",
		Date:     "2026-04-13",
		Format:   sharedtypes.ExportFormatMarkdown,
		BaseName: "daily-2026-04-13",
	}
	rendered, err := RenderTemplate(fallbackDailyReportTemplate, attachFrontmatter(buildTemplateDataMap(data), spec))
	if err != nil {
		t.Fatalf("render daily report template: %v", err)
	}
	for _, snippet := range []string{
		"---\ntitle: \"Daily Report - 2026-04-13\"",
		"  - \"crona\"",
		"  - \"report\"",
		"  - \"crona/daily\"",
		"report_kind: \"daily\"",
		"date: \"2026-04-13\"",
		"generated_at: \"2026-04-13T10:00:00Z\"",
		"crona_version: \"",
		"---\n\n# Daily Report - 2026-04-13",
	} {
		if !strings.Contains(rendered, snippet) {
			t.Fatalf("expected rendered report to contain %q\n%s", snippet, rendered)
		}
	}
	for _, unwanted := range []string{"start_date:", "end_date:", "repo:", "stream:", "scope:"} {
		if strings.Contains(rendered, unwanted) {
			t.Fatalf("expected daily frontmatter not to contain %q\n%s", unwanted, rendered)
		}
	}
}

func TestScopedMarkdownFrontmatterIncludesObsidianScopeTags(t *testing.T) {
	data := map[string]any{
		"generatedAt": "2026-04-13T10:00:00Z",
		"startDate":   "2026-04-07",
		"endDate":     "2026-04-13",
		"repo":        map[string]any{"name": "Webxsid Core"},
		"stream":      map[string]any{"name": "Report Templates"},
	}
	spec := reportWriteSpec{
		Kind:       sharedtypes.ExportReportKindStream,
		Label:      "Stream Report",
		ScopeLabel: "Webxsid Core / Report Templates",
		Date:       "2026-04-13",
		StartDate:  "2026-04-07",
		EndDate:    "2026-04-13",
		Format:     sharedtypes.ExportFormatMarkdown,
		BaseName:   "stream-report",
	}
	rendered, err := RenderTemplate(fallbackStreamReportTemplate, attachFrontmatter(data, spec))
	if err != nil {
		t.Fatalf("render stream report template: %v", err)
	}
	for _, snippet := range []string{
		"title: \"Stream Report - Report Templates\"",
		"scope: \"Webxsid Core / Report Templates\"",
		"repo: \"Webxsid Core\"",
		"stream: \"Report Templates\"",
		"  - \"repo/webxsid-core\"",
		"  - \"stream/report-templates\"",
		"start_date: \"2026-04-07\"",
		"end_date: \"2026-04-13\"",
	} {
		if !strings.Contains(rendered, snippet) {
			t.Fatalf("expected rendered report to contain %q\n%s", snippet, rendered)
		}
	}
}

func TestRuntimeFrontmatterWrapsMarkdownTemplatesWithoutFrontmatter(t *testing.T) {
	data := map[string]any{
		"generatedAt": "2026-04-13T10:00:00Z",
		"startDate":   "2026-04-07",
		"endDate":     "2026-04-13",
		"repo":        map[string]any{"name": "Webxsid Core"},
	}
	spec := reportWriteSpec{
		Kind:       sharedtypes.ExportReportKindRepo,
		Label:      "Repo Report",
		ScopeLabel: "Webxsid Core",
		Date:       "2026-04-13",
		StartDate:  "2026-04-07",
		EndDate:    "2026-04-13",
		Format:     sharedtypes.ExportFormatMarkdown,
		BaseName:   "repo-report",
	}
	rendered, err := RenderTemplate("# Custom Repo Report\n\nBody", attachFrontmatter(data, spec))
	if err != nil {
		t.Fatalf("render custom repo template: %v", err)
	}
	wrapped := ensureMarkdownFrontmatter(rendered, data, spec)
	for _, snippet := range []string{
		"---\ntitle: \"Repo Report - Webxsid Core\"",
		"report_kind: \"repo\"",
		"repo: \"Webxsid Core\"",
		"---\n\n# Custom Repo Report",
	} {
		if !strings.Contains(wrapped, snippet) {
			t.Fatalf("expected wrapped report to contain %q\n%s", snippet, wrapped)
		}
	}
	if strings.Contains(wrapped, "stream:") {
		t.Fatalf("expected repo frontmatter not to contain stream\n%s", wrapped)
	}
}

func TestMarkdownPresetsRenderFrontmatter(t *testing.T) {
	body, ok := presetTemplateBody(sharedtypes.ExportReportKindWeekly, sharedtypes.ExportAssetKindTemplateMarkdown, "brief")
	if !ok {
		t.Fatal("expected weekly brief markdown preset")
	}
	data := map[string]any{
		"generatedAt": "2026-04-13T10:00:00Z",
		"startDate":   "2026-04-07",
		"endDate":     "2026-04-13",
		"summary":     map[string]any{},
		"streaks":     map[string]any{},
		"days":        []map[string]any{},
	}
	rendered, err := RenderTemplate(body, attachFrontmatter(data, reportWriteSpec{
		Kind:      sharedtypes.ExportReportKindWeekly,
		Label:     "Weekly Summary",
		Date:      "2026-04-13",
		StartDate: "2026-04-07",
		EndDate:   "2026-04-13",
		Format:    sharedtypes.ExportFormatMarkdown,
		BaseName:  "weekly-2026-04-07-to-2026-04-13",
	}))
	if err != nil {
		t.Fatalf("render weekly preset: %v", err)
	}
	for _, snippet := range []string{"---\ntitle: \"Weekly Summary - 2026-04-07 to 2026-04-13\"", "report_kind: \"weekly\"", "# 📅 Weekly Snapshot"} {
		if !strings.Contains(rendered, snippet) {
			t.Fatalf("expected rendered preset to contain %q\n%s", snippet, rendered)
		}
	}
}

func planFailurePtr(reason sharedtypes.DailyPlanFailureReason) *sharedtypes.DailyPlanFailureReason {
	return &reason
}
