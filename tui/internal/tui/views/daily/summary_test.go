package daily

import (
	"strings"
	"testing"

	"crona/tui/internal/api"
	"crona/tui/internal/tui/chrome"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/x/ansi"
)

func TestRenderSummaryShowsMomentumBlock(t *testing.T) {
	state := dailySummaryTestState()
	rendered := ansi.Strip(renderSummary(testTheme(), state, 120, 60))
	lines := strings.Split(rendered, "\n")

	for _, want := range []string{
		"Signals",
		"Momentums",
		"Work",
		"Energy",
		"Mood",
		"Sleep",
		"Check-ins",
		"Focus",
		"today 4/5",
		"today 3/5",
		"today 7h30m",
		"2h",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("expected daily summary to contain %q, got %q", want, rendered)
		}
	}
	assertContainsLine(t, lines, "Work", "Energy")
	assertContainsLine(t, lines, "Mood", "Sleep")
	assertContainsLine(t, lines, "Check-ins", "Focus")
	for _, unwanted := range []string{"Custom Momentum"} {
		if strings.Contains(rendered, unwanted) {
			t.Fatalf("expected daily momentum block to omit %q, got %q", unwanted, rendered)
		}
	}
}

func TestRenderSummaryUsesCompactMomentumBlockAtShortHeight(t *testing.T) {
	state := dailySummaryTestState()
	lines := renderMomentumBlock(testTheme(), state, 120)
	rendered := ansi.Strip(strings.Join(lines, "\n"))
	for _, want := range []string{"Signals", "Momentums", "Check-ins", "Focus", "|"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("expected compact momentum block to contain %q, got %#v", want, rendered)
		}
	}
	if !strings.Contains(rendered, "today 4/5") || !strings.Contains(rendered, "today 3/5") {
		t.Fatalf("expected compact momentum block to keep check-in values, got %#v", rendered)
	}
	assertContainsLine(t, strings.Split(rendered, "\n"), "Work", "Energy")
	assertContainsLine(t, strings.Split(rendered, "\n"), "Mood", "Sleep")
	assertContainsLine(t, strings.Split(rendered, "\n"), "Check-ins", "Focus")
}

func TestRenderSummaryUsesAverageSignalsWhenCheckInMissing(t *testing.T) {
	state := dailySummaryTestState()
	state.DailyCheckIn = nil

	rendered := ansi.Strip(renderSummary(testTheme(), state, 120, 60))

	for _, want := range []string{"Signals", "Momentums", "avg 3.1/5", "avg 3.8/5", "avg 6h15m", "2h"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("expected daily summary to fall back to averages for %q, got %q", want, rendered)
		}
	}
	if strings.Contains(rendered, "today 4/5") || strings.Contains(rendered, "today 3/5") {
		t.Fatalf("expected daily summary to use averages instead of today's check-in values, got %q", rendered)
	}
}

func TestRenderSummaryOmitsMomentumWhenNoSignalData(t *testing.T) {
	state := dailySummaryTestState()
	state.DailyCheckIn = nil
	state.MetricsRollup = nil
	state.Streaks = nil

	rendered := ansi.Strip(renderSummary(testTheme(), state, 120, 60))

	for _, unwanted := range []string{"Signals", "Momentums"} {
		if strings.Contains(rendered, unwanted) {
			t.Fatalf("expected daily summary to omit %q when no signal data is present, got %q", unwanted, rendered)
		}
	}
	if !strings.Contains(rendered, "Issues") || !strings.Contains(rendered, "Habits") {
		t.Fatalf("expected existing daily summary content to remain, got %q", rendered)
	}
}

func dailySummaryTestState() types.ContentState {
	sleep := 7.5
	avgMood := 3.1
	avgEnergy := 3.8
	avgSleep := 6.25
	return types.ContentState{
		View:          "daily",
		Pane:          "issues",
		Width:         120,
		Height:        32,
		DashboardDate: "2026-05-27",
		DailySummary: &api.DailyIssueSummary{
			Date: "2026-05-27",
		},
		DailyIssues: []api.Issue{
			{
				ID:     1,
				Title:  "Ship momentum block",
				Status: "done",
			},
		},
		DueHabits: []api.HabitDailyItem{
			{
				HabitWithMeta: api.HabitWithMeta{
					Habit: api.Habit{Name: "Stretch"},
				},
				Status: "completed",
			},
		},
		DailyCheckIn: &api.DailyCheckIn{
			Date:       "2026-05-27",
			Mood:       4,
			Energy:     3,
			SleepHours: &sleep,
			CreatedAt:  "2026-05-27T09:00:00Z",
		},
		MetricsRollup: &api.MetricsRollup{
			Days:              14,
			WorkedSeconds:     7200,
			AverageMood:       &avgMood,
			AverageEnergy:     &avgEnergy,
			AverageSleepHours: &avgSleep,
		},
		Streaks: &api.StreakSummary{
			CurrentCheckInDays: 2,
			LongestCheckInDays: 5,
			CurrentFocusDays:   1,
			LongestFocusDays:   3,
			CurrentHabitDays:   4,
			LongestHabitDays:   9,
		},
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

func assertContainsLine(t *testing.T, lines []string, wantA, wantB string) {
	t.Helper()
	for _, line := range lines {
		if strings.Contains(line, wantA) && strings.Contains(line, wantB) {
			return
		}
	}
	t.Fatalf("expected a line containing both %q and %q, got %#v", wantA, wantB, lines)
}
