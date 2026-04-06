package meta

import (
	"fmt"
	"strings"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	viewchrome "crona/tui/internal/tui/views/chrome"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	types "crona/tui/internal/tui/views/types"
)

func renderHabits(theme types.Theme, state types.ContentState, width, height int) string {
	active := state.Pane == "habits"
	cur := state.Cursors["habits"]
	items := habitItems(state.Habits)
	indices := viewhelpers.FilteredStrings(items, state.Filters["habits"])
	total := len(indices)
	lines := []string{
		theme.StylePaneTitle.Render("Habits [4]"),
		viewchrome.RenderPaneActionLine(theme, state.Filters["habits"], width-6, paneActions(theme, state, "habits")),
	}
	if total == 0 {
		lines = append(lines, theme.StyleDim.Render("No habits — [a] create new"))
		return viewchrome.RenderPaneBox(theme, active, width, height, viewhelpers.StringsJoin(lines))
	}
	inner := viewchrome.RemainingPaneHeight(height, lines)
	start, end := viewchrome.ListWindow(cur, total, inner)
	if start > 0 {
		lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("↑ %d more", start)))
	}
	for i := start; i < end; i++ {
		lines = append(lines, viewchrome.RenderPaneRowStyled(theme, i, cur, active, items[indices[i]], nil, width))
	}
	if remaining := total - end; remaining > 0 {
		lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("↓ %d more", remaining)))
	}
	return viewchrome.RenderPaneBox(theme, active, width, height, viewhelpers.StringsJoin(lines))
}

func habitItems(habits []api.Habit) []string {
	items := make([]string, 0, len(habits))
	for _, habit := range habits {
		schedule := string(habit.ScheduleType)
		if habit.ScheduleType == sharedtypes.HabitScheduleWeekly {
			schedule = formatHabitScheduleText(habit.Weekdays)
		}
		items = append(items, fmt.Sprintf("%s [%s]", habit.Name, schedule))
	}
	return items
}

func formatHabitScheduleText(weekdays []int) string {
	names := []string{"sun", "mon", "tue", "wed", "thu", "fri", "sat"}
	out := make([]string, 0, len(weekdays))
	for _, day := range weekdays {
		if day >= 0 && day < len(names) {
			out = append(out, names[day])
		}
	}
	return strings.Join(out, ",")
}
