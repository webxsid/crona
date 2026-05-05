package wellbeing_test

import (
	"strings"
	"testing"

	"crona/tui/internal/api"
	layoutpkg "crona/tui/internal/tui/layout"
	types "crona/tui/internal/tui/views/types"
	wellbeingviews "crona/tui/internal/tui/views/wellbeing"
)

func TestWellbeingRenderShowsHabitStreakLabel(t *testing.T) {
	state := types.ContentState{
		View:   "wellbeing",
		Pane:   "wellbeing",
		Width:  120,
		Height: 40,
		Streaks: &api.StreakSummary{
			CurrentCheckInDays: 2,
			LongestCheckInDays: 5,
			CurrentFocusDays:   1,
			LongestFocusDays:   3,
			CurrentHabitDays:   4,
			LongestHabitDays:   9,
		},
	}

	rendered := wellbeingviews.Render(layoutpkg.ViewTheme(), state)
	for _, want := range []string{"Streaks", "H4/9"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("expected wellbeing render to contain %q, got %q", want, rendered)
		}
	}
}

func TestWellbeingRenderShowsHabitRollupCounts(t *testing.T) {
	state := types.ContentState{
		View:   "wellbeing",
		Pane:   "wellbeing",
		Width:  120,
		Height: 40,
		MetricsRollup: &api.MetricsRollup{
			Days:                7,
			CheckInDays:         5,
			FocusDays:           4,
			WorkedSeconds:       7200,
			HabitDueCount:       3,
			HabitCompletedCount: 2,
			HabitFailedCount:    1,
		},
	}

	rendered := wellbeingviews.Render(layoutpkg.ViewTheme(), state)
	for _, want := range []string{"Habits due", "Done", "Failed"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("expected wellbeing render to contain %q, got %q", want, rendered)
		}
	}
}
