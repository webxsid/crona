package rollup

import (
	"strings"
	"testing"

	"crona/tui/internal/api"
	"crona/tui/internal/tui/chrome"
	types "crona/tui/internal/tui/views/types"
)

func TestRenderShowsCalendarAndFocusVisual(t *testing.T) {
	state := types.ContentState{
		View:            "rollup",
		Pane:            "rollup_days",
		Width:           140,
		Height:          42,
		RollupStartDate: "2026-06-01",
		RollupEndDate:   "2026-06-07",
		WeekStart:       "monday",
		RollupMetricsRange: []api.DailyMetricsDay{
			{Date: "2026-06-01", TotalEstimatedMinutes: 120, WorkedSeconds: 3600},
			{Date: "2026-06-02", TotalEstimatedMinutes: 180, WorkedSeconds: 5400},
			{Date: "2026-06-03", TotalEstimatedMinutes: 60, WorkedSeconds: 1800},
			{Date: "2026-06-04", TotalEstimatedMinutes: 0, WorkedSeconds: 0},
			{Date: "2026-06-05", TotalEstimatedMinutes: 240, WorkedSeconds: 7200},
		},
		GoalProgress: &api.GoalProgressSummary{
			TotalEstimateMinutes: 600,
			TotalActualSeconds:   19800,
			EstimateBias:         "balanced",
		},
		WeeklyFocusScore: &api.FocusScoreSummary{Score: 77},
	}

	rendered := Render(testTheme(), state)
	for _, want := range []string{
		"Rollup Dashboard",
		"Focus",
		"logged bars with estimate line",
		"focus 77/100",
		"│",
		"June 2026",
		"Wk  Mo Tu We Th Fr Sa Su",
		"estimated 10.0h",
		"logged 5.5h",
		"delta -4.5h   under estimate",
		"Breakdown",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("expected rollup render to include %q, got %q", want, rendered)
		}
	}
	if strings.Contains(rendered, "\n│  ...") {
		t.Fatalf("expected focus summary to fit without clipped ellipsis, got %q", rendered)
	}
	for _, unwanted := range []string{"◆"} {
		if strings.Contains(rendered, unwanted) {
			t.Fatalf("did not expect rollup render to include %q, got %q", unwanted, rendered)
		}
	}
	for _, want := range []string{"█", "╌"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("expected rollup render to include %q, got %q", want, rendered)
		}
	}
	for _, label := range []string{"4h │", "3h │", "2h │", "1h │"} {
		if strings.Count(rendered, label) > 1 {
			t.Fatalf("expected y-axis label %q to appear at most once, got %q", label, rendered)
		}
	}
}

func TestRenderBreakdownPaneSupportsScrolling(t *testing.T) {
	state := types.ContentState{
		View:   "rollup",
		Pane:   "rollup_breakdown",
		Width:  96,
		Height: 18,
		Cursors: map[string]int{
			"rollup_days":      0,
			"rollup_breakdown": 4,
		},
		RepoDistribution: &api.TimeDistributionSummary{
			Rows: []api.TimeDistributionRow{
				{Label: "Alpha", Percent: 50, WorkedSeconds: 3600},
				{Label: "Beta", Percent: 40, WorkedSeconds: 2400},
				{Label: "Gamma", Percent: 30, WorkedSeconds: 1800},
			},
		},
		StreamDistribution: &api.TimeDistributionSummary{
			Rows: []api.TimeDistributionRow{
				{Label: "Main", Percent: 70, WorkedSeconds: 4200},
			},
		},
		IssueDistribution: &api.TimeDistributionSummary{
			Rows: []api.TimeDistributionRow{
				{Label: "Issue A", Percent: 60, WorkedSeconds: 3600},
			},
		},
		SegmentDistribution: &api.TimeDistributionSummary{
			Rows: []api.TimeDistributionRow{
				{Label: "Focus", Percent: 80, WorkedSeconds: 4800},
			},
		},
	}

	rendered := Render(testTheme(), state)
	for _, want := range []string{"Breakdown", "↑ more", "↓ more"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("expected rollup breakdown scroll render to include %q, got %q", want, rendered)
		}
	}
}

func testTheme() types.Theme {
	return types.Theme{
		ColorBlue:            chrome.ColorBlue,
		ColorCyan:            chrome.ColorCyan,
		ColorGreen:           chrome.ColorGreen,
		ColorMagenta:         chrome.ColorMagenta,
		ColorSubtle:          chrome.ColorSubtle,
		ColorYellow:          chrome.ColorYellow,
		ColorRed:             chrome.ColorRed,
		ColorDim:             chrome.ColorDim,
		ColorWhite:           chrome.ColorWhite,
		StyleActive:          chrome.StyleActive,
		StyleInactive:        chrome.StyleInactive,
		StylePaneTitle:       chrome.StylePaneTitle,
		StyleDim:             chrome.StyleDim,
		StyleCursor:          chrome.StyleCursor,
		StyleHeader:          chrome.StyleHeader,
		StyleError:           chrome.StyleError,
		StyleSelected:        chrome.StyleSelected,
		StyleSelectedInverse: chrome.StyleSelectedInverse,
		StyleNormal:          chrome.StyleNormal,
	}
}
