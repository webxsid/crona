package wellbeing

import (
	"fmt"
	"strings"
	"testing"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	"crona/tui/internal/tui/chrome"
	uistate "crona/tui/internal/tui/state"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/x/ansi"
)

func TestWideWellbeingSplitsMetricsAndStreaks(t *testing.T) {
	state := splitTestState()
	state.Height = 60
	state.Pane = string(uistate.PaneWellbeingStreaks)

	rendered := ansi.Strip(Render(splitTestTheme(), state))
	for _, want := range []string{"Metrics Window", "Momentum", "Signals (14d)", "Mood", "Energy", "Work", "Recovery", "Training", "▰"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("expected wide wellbeing split to contain %q, got %q", want, rendered)
		}
	}
	if !strings.ContainsFunc(rendered, func(r rune) bool { return r >= 0x2800 && r <= 0x28ff }) {
		t.Fatalf("expected wide wellbeing split to contain braille graph output, got %q", rendered)
	}
}

func TestMetricsBodyExcludesStreaksInSplitMode(t *testing.T) {
	state := splitTestState()
	body := strings.Join(flattenLines(metricsBodyLines(splitTestTheme(), state, 70, false)), "\n")
	for _, unwanted := range []string{"Training", "Custom Momentum", "current ·"} {
		if strings.Contains(body, unwanted) {
			t.Fatalf("expected metrics-only body to omit %q, got %q", unwanted, body)
		}
	}
}

func TestCombinedMetricsBodyKeepsStreaksForNarrowFallback(t *testing.T) {
	state := splitTestState()
	body := strings.Join(flattenLines(trendsBodyLines(splitTestTheme(), state, 92, false)), "\n")
	for _, want := range []string{"Check-ins", "Focus", "Custom Momentum", "Training", "weekly", "current ·", "3d", "2w"} {
		if !strings.Contains(body, want) {
			t.Fatalf("expected combined metrics fallback to include %q, got %q", want, body)
		}
	}
	if strings.Contains(body, "\nHabit Streak") {
		t.Fatalf("expected generic habit streak row to be removed, got %q", body)
	}
}

func TestMomentumRowsRenderCadenceAndMeters(t *testing.T) {
	state := splitTestState()
	state.Streaks.CustomHabitStreaks = []sharedtypes.CustomHabitStreakSummary{
		{Name: "Daily Reflection", Period: sharedtypes.HabitStreakPeriodDay, Current: 3, Longest: 5},
		{Name: "Training Week", Period: sharedtypes.HabitStreakPeriodWeek, Current: 2, Longest: 6},
		{Name: "Wellbeing Month", Period: sharedtypes.HabitStreakPeriodMonth, Current: 1, Longest: 4},
	}
	body := ansi.Strip(strings.Join(flattenLines(streaksBodyLines(splitTestTheme(), state, 76, false)), "\n"))
	for _, want := range []string{"Custom Momentum", "daily", "weekly", "monthly", "▰", "▱", "current ·", "3d current", "2w current", "1mo current", "5d best", "6w best", "4mo best"} {
		if !strings.Contains(body, want) {
			t.Fatalf("expected momentum rows to include %q, got %q", want, body)
		}
	}
}

func TestMomentumLadderShowsAbsoluteLengthWhenBothAreBest(t *testing.T) {
	short := ansi.Strip(momentumLadder(splitTestTheme(), sharedtypes.HabitStreakPeriodDay, 2, 2))
	long := ansi.Strip(momentumLadder(splitTestTheme(), sharedtypes.HabitStreakPeriodDay, 40, 40))

	if strings.Count(short, "▰") >= strings.Count(long, "▰") {
		t.Fatalf("expected longer streak to fill more tiers, short %q long %q", short, long)
	}
}

func TestMomentumLadderHandlesZeroBest(t *testing.T) {
	state := splitTestState()
	state.Streaks.CurrentCheckInDays = 0
	state.Streaks.LongestCheckInDays = 0

	body := ansi.Strip(strings.Join(flattenLines(streaksBodyLines(splitTestTheme(), state, 76, false)), "\n"))
	for _, want := range []string{"▱▱▱▱", "0d current", "0d best"} {
		if !strings.Contains(body, want) {
			t.Fatalf("expected empty momentum meter to include %q, got %q", want, body)
		}
	}
}

func TestMomentumTierCountsUseCadenceThresholds(t *testing.T) {
	tests := []struct {
		name    string
		period  sharedtypes.HabitStreakPeriod
		current int
		want    int
	}{
		{name: "daily", period: sharedtypes.HabitStreakPeriodDay, current: 14, want: 4},
		{name: "weekly", period: sharedtypes.HabitStreakPeriodWeek, current: 4, want: 3},
		{name: "monthly", period: sharedtypes.HabitStreakPeriodMonth, current: 3, want: 3},
	}

	for _, tt := range tests {
		if got := momentumTierCount(tt.period, tt.current); got != tt.want {
			t.Fatalf("%s tier count = %d, want %d", tt.name, got, tt.want)
		}
	}
}

func TestStreaksPaneLineCountSupportsIndependentScrolling(t *testing.T) {
	state := splitTestState()
	state.Streaks.CustomHabitStreaks = append(state.Streaks.CustomHabitStreaks,
		sharedtypes.CustomHabitStreakSummary{Name: "Reading", Period: sharedtypes.HabitStreakPeriodDay, Current: 4, Longest: 8},
		sharedtypes.CustomHabitStreakSummary{Name: "Mobility", Period: sharedtypes.HabitStreakPeriodWeek, Current: 3, Longest: 9},
		sharedtypes.CustomHabitStreakSummary{Name: "Meditation", Period: sharedtypes.HabitStreakPeriodMonth, Current: 6, Longest: 12},
	)
	count := PaneLineCount(state, string(uistate.PaneWellbeingStreaks))
	if count < 12 {
		t.Fatalf("expected streaks pane line count to include custom streak rows, got %d", count)
	}
}

func splitTestState() types.ContentState {
	avgMood := 4.1
	avgEnergy := 3.7
	return types.ContentState{
		View:          "wellbeing",
		Pane:          string(uistate.PaneWellbeingTrends),
		Width:         120,
		Height:        40,
		WellbeingDate: "2026-04-04",
		WellbeingWindowDays: 14,
		Cursors: map[string]int{
			string(uistate.PaneWellbeingSummary): 0,
			string(uistate.PaneWellbeingTrends):  0,
			string(uistate.PaneWellbeingStreaks): 0,
		},
		MetricsRollup: &api.MetricsRollup{
			Days:          14,
			CheckInDays:   5,
			FocusDays:     4,
			WorkedSeconds: 7200,
			RestSeconds:   900,
			SessionCount:  6,
			AverageMood:   &avgMood,
			AverageEnergy: &avgEnergy,
		},
		MetricsRange: makeWellbeingMetricsRange(14),
		Streaks: &api.StreakSummary{
			CurrentCheckInDays: 3,
			LongestCheckInDays: 7,
			CurrentFocusDays:   2,
			LongestFocusDays:   5,
			CurrentHabitDays:   4,
			LongestHabitDays:   9,
			CustomHabitStreaks: []sharedtypes.CustomHabitStreakSummary{
				{Name: "Training", Period: sharedtypes.HabitStreakPeriodWeek, Current: 2, Longest: 6},
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
				Date:        fmt.Sprintf("2026-04-%02d", i+1),
				Mood:        mood,
				Energy:      energy,
				SleepHours:  &sleep,
				CreatedAt:   fmt.Sprintf("2026-04-%02dT09:00:00Z", i+1),
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
