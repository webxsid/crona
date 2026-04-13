package export

import (
	"strings"
	"testing"

	sharedtypes "crona/shared/types"
)

func TestRenderTemplateTrimsStandaloneControlLines(t *testing.T) {
	rendered, err := RenderTemplate(`A
{{#if enabled}}
B
{{/if}}
C`, map[string]any{"enabled": true})
	if err != nil {
		t.Fatalf("render template: %v", err)
	}
	if rendered != "A\nB\nC" {
		t.Fatalf("expected standalone control lines to be removed, got %q", rendered)
	}
}

func TestRenderTemplateTrimsSkippedBlockWithoutLeavingBlankLines(t *testing.T) {
	rendered, err := RenderTemplate(`A
{{#if enabled}}
B
{{/if}}
C`, map[string]any{"enabled": false})
	if err != nil {
		t.Fatalf("render template: %v", err)
	}
	if rendered != "A\nC" {
		t.Fatalf("expected skipped block not to leave blank lines, got %q", rendered)
	}
}

func TestRenderTemplatePreservesIntentionalBlankLinesAndInlineVariables(t *testing.T) {
	rendered, err := RenderTemplate("A\n\n- {{value}}\n", map[string]any{"value": "B"})
	if err != nil {
		t.Fatalf("render template: %v", err)
	}
	if rendered != "A\n\n- B\n" {
		t.Fatalf("expected explicit spacing and inline variables to remain, got %q", rendered)
	}
}

func TestRenderedDailyTemplateAvoidsControlLineWhitespaceNoise(t *testing.T) {
	data := buildTemplateDataMap(&sharedtypes.DailyReportData{
		Date:        "2026-04-13",
		GeneratedAt: "2026-04-13T10:00:00Z",
		Summary: sharedtypes.DailyIssueSummary{
			Date:          "2026-04-13",
			TotalIssues:   1,
			WorkedSeconds: 1800,
		},
	})
	rendered, err := RenderTemplate(fallbackDailyReportTemplate, attachFrontmatter(data, reportWriteSpec{
		Kind:     "daily",
		Label:    "Daily Report",
		Date:     "2026-04-13",
		Format:   "markdown",
		BaseName: "daily-2026-04-13",
	}))
	if err != nil {
		t.Fatalf("render daily report: %v", err)
	}
	if strings.Contains(rendered, "\n\n\n\n") {
		t.Fatalf("expected rendered report not to contain four adjacent newlines\n%s", rendered)
	}
}
