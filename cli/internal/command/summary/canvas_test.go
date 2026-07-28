package summary

import (
	"strings"
	"testing"

	sharedtypes "crona/shared/types"
)

func TestDayCanvasUsesRequestedWidthWithoutANSIInPlainMode(t *testing.T) {
	data := daySummaryData{
		Date: "2026-07-28",
		Summary: &sharedtypes.DailyIssueSummary{
			TotalIssues:     1,
			CompletedIssues: 1,
			WorkedSeconds:   1800,
			Issues: []sharedtypes.Issue{{
				Title:  strings.Repeat("long issue ", 20),
				Status: sharedtypes.IssueStatusDone,
			}},
		},
	}
	var out strings.Builder
	if err := renderDayCanvas(&out, data, renderOptions{Width: 72}); err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(strings.TrimSuffix(out.String(), "\n"), "\n") {
		if got := visibleWidth(line); got > 72 {
			t.Fatalf("line width %d exceeds 72: %q", got, line)
		}
	}
	if strings.Contains(out.String(), "\x1b[") {
		t.Fatalf("plain output contains ANSI: %q", out.String())
	}
	for _, want := range []string{"AT A GLANCE", "AGENDA", "SIGNALS", "####"} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("expected %q in output:\n%s", want, out.String())
		}
	}
}

func TestProgressBarUsesColorAndBlocksForTerminalMode(t *testing.T) {
	bar := progressBar(renderOptions{Color: true, Unicode: true}, 3, 1, 5, 10, ansiGreen)
	if !strings.Contains(bar, "\x1b[32m") || !strings.Contains(bar, "█") || !strings.Contains(bar, "░") {
		t.Fatalf("expected colored Unicode progress bar, got %q", bar)
	}
	if visibleWidth(bar) != 10 {
		t.Fatalf("expected visible width 10, got %d", visibleWidth(bar))
	}
}

func TestRangeCanvasAdaptsDailyColumns(t *testing.T) {
	data := rangeSummaryData{
		Start: "2026-07-01",
		End:   "2026-07-07",
		Metrics: []sharedtypes.DailyMetricsDay{{
			Date:                "2026-07-01",
			WorkedSeconds:       3600,
			CompletedIssues:     1,
			TotalIssues:         2,
			HabitCompletedCount: 1,
			HabitDueCount:       2,
		}},
	}
	var narrow, wide strings.Builder
	_ = renderRangeCanvas(&narrow, data, renderOptions{Width: 72})
	_ = renderRangeCanvas(&wide, data, renderOptions{Width: 120})
	if strings.Contains(narrow.String(), "SESSIONS ISSUES") {
		t.Fatalf("narrow output retained wide-only columns:\n%s", narrow.String())
	}
	if !strings.Contains(wide.String(), "SESSIONS") || !strings.Contains(wide.String(), "MOOD") {
		t.Fatalf("wide output omitted detailed columns:\n%s", wide.String())
	}
}
