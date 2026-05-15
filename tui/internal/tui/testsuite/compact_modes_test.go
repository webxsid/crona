package testsuite

import (
	"testing"

	"crona/tui/internal/api"
	"crona/tui/internal/tui/testsuite/support"
	"crona/tui/internal/tui/views"
)

func TestTinyAndCompactModes(t *testing.T) {
	estimate := 60
	target := 15
	tests := []struct {
		name   string
		height int
		render func(int) string
		assert func(*testing.T, string, int)
	}{
		{
			name:   "daily 36",
			height: 36,
			render: func(height int) string {
				state := views.ContentState{
					View:   "daily",
					Pane:   "issues",
					Width:  70,
					Height: height,
					Cursors: map[string]int{
						"issues": 0,
						"habits": 0,
					},
					Filters: map[string]string{
						"issues": "",
						"habits": "",
					},
					DailySummary: &api.DailyIssueSummary{
						Date: "2026-03-19",
						Issues: []api.Issue{
							{ID: 1, Title: "Add keyboard-first command palette", Status: "planned", EstimateMinutes: &estimate},
						},
					},
					DailyIssues: []api.Issue{
						{ID: 1, Title: "Add keyboard-first command palette", Status: "planned", EstimateMinutes: &estimate},
					},
					DueHabits: []api.HabitDailyItem{
						{HabitWithMeta: api.HabitWithMeta{Habit: api.Habit{Name: "Inbox Zero Sweep", TargetMinutes: &target}}, Status: "pending"},
					},
					Context: &api.ActiveContext{
						RepoName:   strPtr("Work"),
						StreamName: strPtr("app"),
					},
				}
				return support.RenderDaily(state)
			},
			assert: func(t *testing.T, rendered string, _ int) {
				assertTinySummary(t, rendered)
			},
		},
		{
			name:   "daily 30",
			height: 30,
			render: func(height int) string {
				state := views.ContentState{
					View:   "daily",
					Pane:   "issues",
					Width:  70,
					Height: height,
					Cursors: map[string]int{
						"issues": 0,
						"habits": 0,
					},
					Filters: map[string]string{
						"issues": "",
						"habits": "",
					},
					DailySummary: &api.DailyIssueSummary{
						Date: "2026-03-19",
						Issues: []api.Issue{
							{ID: 1, Title: "Add keyboard-first command palette", Status: "planned", EstimateMinutes: &estimate},
						},
					},
					DailyIssues: []api.Issue{
						{ID: 1, Title: "Add keyboard-first command palette", Status: "planned", EstimateMinutes: &estimate},
					},
					DueHabits: []api.HabitDailyItem{
						{HabitWithMeta: api.HabitWithMeta{Habit: api.Habit{Name: "Inbox Zero Sweep", TargetMinutes: &target}}, Status: "pending"},
					},
					Context: &api.ActiveContext{
						RepoName:   strPtr("Work"),
						StreamName: strPtr("app"),
					},
				}
				return support.RenderDaily(state)
			},
			assert: func(t *testing.T, rendered string, _ int) {
				assertTinySummary(t, rendered)
			},
		},
		{
			name:   "default 36",
			height: 36,
			render: func(height int) string {
				return support.RenderDefault(compactDefaultState(height))
			},
			assert: func(t *testing.T, rendered string, height int) {
				assertCompactDefault(t, rendered, height)
			},
		},
		{
			name:   "default 30",
			height: 30,
			render: func(height int) string {
				return support.RenderDefault(compactDefaultState(height))
			},
			assert: func(t *testing.T, rendered string, height int) {
				assertCompactDefault(t, rendered, height)
			},
		},
		{
			name:   "wellbeing 36",
			height: 36,
			render: func(height int) string {
				return support.RenderWellbeing(compactWellbeingState(height))
			},
			assert: func(t *testing.T, rendered string, height int) {
				assertCompactWellbeing(t, rendered, height)
			},
		},
		{
			name:   "wellbeing 30",
			height: 30,
			render: func(height int) string {
				return support.RenderWellbeing(compactWellbeingState(height))
			},
			assert: func(t *testing.T, rendered string, height int) {
				assertCompactWellbeing(t, rendered, height)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rendered := tt.render(tt.height)
			tt.assert(t, rendered, tt.height)
		})
	}
}
