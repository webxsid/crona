package viewruntime

import (
	"time"

	"crona.work/tui/internal/api"
	viewtypes "crona.work/tui/internal/tui/views/types"
)

func CurrentDashboardDate(state viewtypes.ContentState) string {
	if state.DashboardDate != "" {
		return state.DashboardDate
	}
	return time.Now().Format("2006-01-02")
}

func HabitDailyItems(habits []api.HabitDailyItem) []string {
	items := make([]string, 0, len(habits))
	for _, habit := range habits {
		items = append(items, habit.Name)
	}
	return items
}
