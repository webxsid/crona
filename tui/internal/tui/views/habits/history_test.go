package habits_test

import (
	"strings"
	"testing"

	"crona/shared/types"
	"crona/tui/internal/api"
	layoutpkg "crona/tui/internal/tui/layout"
	habitviews "crona/tui/internal/tui/views/habits"
	viewtypes "crona/tui/internal/tui/views/types"
)

func TestRenderHabitHistoryShowsSelectedHabitAndCompletionDetails(t *testing.T) {
	notes := "kept the streak going"
	snapshotName := "Morning Focus"
	snapshotDesc := "Focused work block"
	target := 30
	repoName := "Core"
	streamName := "Backend"
	state := viewtypes.ContentState{
		View:   "habit_history",
		Pane:   "habit_history",
		Width:  96,
		Height: 24,
		Cursors: map[string]int{
			"habit_history": 0,
		},
		Filters: map[string]string{
			"habit_history": "",
		},
		Context: &api.ActiveContext{
			RepoName:   &repoName,
			StreamName: &streamName,
		},
		Settings: &api.CoreSettings{
			DateDisplayPreset: types.DateDisplayPresetISO,
		},
		HabitHistory: []api.HabitCompletion{
			{
				ID:              1,
				HabitID:         99,
				HabitName:       "Inbox Zero",
				RepoName:        repoName,
				StreamName:      streamName,
				Kind:            types.HabitHistoryKindCompletion,
				Date:            "2026-04-04",
				Status:          types.HabitCompletionStatusCompleted,
				DurationMinutes: &target,
				Notes:           &notes,
				SnapshotName:    &snapshotName,
				SnapshotDesc:    &snapshotDesc,
				CreatedAt:       "2026-04-04T09:30:00Z",
				UpdatedAt:       "2026-04-04T09:30:00Z",
			},
			{
				ID:         2,
				HabitID:    99,
				HabitName:  "Inbox Zero",
				RepoName:   repoName,
				StreamName: streamName,
				Date:       "2026-04-03",
				Status:     types.HabitCompletionStatusFailed,
				CreatedAt:  "2026-04-03T09:20:00Z",
				UpdatedAt:  "2026-04-03T09:20:00Z",
			},
		},
	}

	rendered := habitviews.Render(layoutpkg.ViewTheme(), state)
	for _, want := range []string{
		"Habit History",
		"Recent habit activity in Core > Backend",
		"2026-04-04",
		"Inbox Zero",
		"Core > Backend",
		"completed",
		"failed",
		"30m",
		"kept the stre...",
		"[enter] details",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("expected habit history render to contain %q, got %q", want, rendered)
		}
	}
}
