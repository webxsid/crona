package wellbeing

import (
	"fmt"
	"strings"
	"testing"

	"crona/tui/internal/api"
	"crona/tui/internal/tui/chrome"
	uistate "crona/tui/internal/tui/state"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/x/ansi"
)

func TestWideWellbeingRendersThreePaneLayout(t *testing.T) {
	state := splitTestState()
	state.Height = 60
	state.Pane = string(uistate.PaneWellbeingDetails)

	rendered := ansi.Strip(Render(splitTestTheme(), state))
	for _, want := range []string{"Wellbeing", "Metrics Window", "Details", "Recent Activity", "Check-in", "Accountability", "Risk Snapshot"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("expected wide wellbeing layout to contain %q, got %q", want, rendered)
		}
	}
	for _, unwanted := range []string{"Custom Momentum", "Training"} {
		if strings.Contains(rendered, unwanted) {
			t.Fatalf("expected wide wellbeing layout to omit %q, got %q", unwanted, rendered)
		}
	}
}

func TestSummaryBodyStaysSnapshotFocused(t *testing.T) {
	state := splitTestState()
	body := strings.Join(flattenLines(summaryBodyLines(splitTestTheme(), state, 90, false)), "\n")
	for _, want := range []string{"Recent Activity", "Today", "Check-in"} {
		if !strings.Contains(body, want) {
			t.Fatalf("expected summary body to include %q, got %q", want, body)
		}
	}
	for _, unwanted := range []string{"Accountability", "Risk Snapshot", "Top Risk Drivers", "Notes"} {
		if strings.Contains(body, unwanted) {
			t.Fatalf("expected summary body to omit %q, got %q", unwanted, body)
		}
	}
}

func TestDetailsBodyCarriesMovedSections(t *testing.T) {
	state := splitTestState()
	body := strings.Join(flattenLines(detailsBodyLines(splitTestTheme(), state, 48, false)), "\n")
	for _, want := range []string{"Check-in", "Mood", "Energy", "Notes", "Accountability", "Risk Snapshot", "Top Risk Drivers"} {
		if !strings.Contains(body, want) {
			t.Fatalf("expected details body to include %q, got %q", want, body)
		}
	}
}

func TestDetailsPaneLineCountSupportsScrolling(t *testing.T) {
	state := splitTestState()
	count := PaneLineCount(state, string(uistate.PaneWellbeingDetails))
	if count <= 0 {
		t.Fatalf("expected details pane line count to be positive, got %d", count)
	}
}

func splitTestState() types.ContentState {
	avgMood := 4.1
	avgEnergy := 3.7
	sleepHours := 7.5
	sleepScore := 84
	screenTime := 95
	notes := "Slept better, but backlog pressure still feels high."
	return types.ContentState{
		View:                "wellbeing",
		Pane:                string(uistate.PaneWellbeingTrends),
		Width:               120,
		Height:              40,
		WellbeingDate:       "2026-04-04",
		WellbeingWindowDays: 14,
		Cursors: map[string]int{
			string(uistate.PaneWellbeingSummary): 0,
			string(uistate.PaneWellbeingTrends):  0,
			string(uistate.PaneWellbeingDetails): 0,
		},
		MetricsRollup: &api.MetricsRollup{
			Days:                14,
			CheckInDays:         5,
			FocusDays:           4,
			WorkedSeconds:       7200,
			RestSeconds:         900,
			SessionCount:        6,
			AverageMood:         &avgMood,
			AverageEnergy:       &avgEnergy,
			HabitDueCount:      9,
			HabitCompletedCount: 6,
			HabitFailedCount:    2,
			LatestBurnout: &api.BurnoutIndicator{
				Score: 61,
				Level: "guarded",
				Factors: map[string]float64{
					"workloadPressure":    0.42,
					"breakDebt":           0.28,
					"sleepDebt":           0.19,
					"recoveryConsistency": -0.15,
				},
			},
		},
		MetricsRange: makeWellbeingMetricsRange(14),
		DailyCheckIn: &api.DailyCheckIn{
			Date:              "2026-04-04",
			Mood:              4,
			Energy:            3,
			SleepHours:        &sleepHours,
			SleepScore:        &sleepScore,
			ScreenTimeMinutes: &screenTime,
			Notes:             &notes,
			CreatedAt:         "2026-04-04T20:00:00Z",
		},
		DailyPlan: &api.DailyPlan{
			Summary: api.DailyPlanAccountabilitySummary{
				PlannedCount:         5,
				CompletedCount:       3,
				FailedCount:          1,
				AbandonedCount:       1,
				PendingRollbackCount: 1,
				AccountabilityScore:  62.5,
				BacklogPressure:      3.1,
				DelayedIssueCount:    2,
				HighRiskIssueCount:   1,
			},
		},
	}
}

func makeWellbeingMetricsRange(days int) []api.DailyMetricsDay {
	out := make([]api.DailyMetricsDay, 0, days)
	for i := 0; i < days; i++ {
		mood := 1 + (i % 5)
		energy := 1 + ((i + 2) % 5)
		sleep := 6.0 + float64(i%3)*0.5
		work := 1200 + i*240
		rest := 180 + i*30
		out = append(out, api.DailyMetricsDay{
			Date: fmt.Sprintf("2026-04-%02d", i+1),
			CheckIn: &api.DailyCheckIn{
				Date:       fmt.Sprintf("2026-04-%02d", i+1),
				Mood:       mood,
				Energy:     energy,
				SleepHours: &sleep,
				CreatedAt:  fmt.Sprintf("2026-04-%02dT09:00:00Z", i+1),
			},
			WorkedSeconds: work,
			RestSeconds:   rest,
			SessionCount:  1 + (i % 4),
		})
	}
	return out
}

func splitTestTheme() types.Theme {
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
