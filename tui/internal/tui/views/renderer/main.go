package viewrenderer

import (
	alerts "crona.work/tui/internal/tui/views/alerts"
	away "crona.work/tui/internal/tui/views/away"
	viewcalendar "crona.work/tui/internal/tui/views/calendar"
	config "crona.work/tui/internal/tui/views/config"
	daily "crona.work/tui/internal/tui/views/daily"
	habitlist "crona.work/tui/internal/tui/views/habitlist"
	habits "crona.work/tui/internal/tui/views/habits"
	issues "crona.work/tui/internal/tui/views/issues"
	momentumview "crona.work/tui/internal/tui/views/momentumview"
	ops "crona.work/tui/internal/tui/views/ops"
	reports "crona.work/tui/internal/tui/views/reports"
	review "crona.work/tui/internal/tui/views/review"
	sessions "crona.work/tui/internal/tui/views/sessions"
	settings "crona.work/tui/internal/tui/views/settings"
	support "crona.work/tui/internal/tui/views/support"
	tagsview "crona.work/tui/internal/tui/views/tags"
	types "crona.work/tui/internal/tui/views/types"
	updates "crona.work/tui/internal/tui/views/updates"
)

func RenderContent(theme types.Theme, state types.ContentState) string {
	switch state.View {
	case "away":
		return away.Render(theme, state)
	case "default":
		return issues.Render(theme, state)
	case "daily":
		return daily.Render(theme, state)
	case "review":
		return review.Render(theme, state)
	case "momentum":
		return momentumview.Render(theme, state)
	case "habit_list":
		return habitlist.Render(theme, state)
	case "session_history", "session_active":
		return sessions.Render(theme, state)
	case "habit_history":
		return habits.Render(theme, state)
	case "ops":
		return ops.Render(theme, state)
	case "config":
		return config.Render(theme, state)
	case "reports":
		return reports.Render(theme, state)
	case "settings":
		return settings.Render(theme, state)
	case "tags":
		return tagsview.Render(theme, state)
	case "alerts":
		return alerts.Render(theme, state)
	case "updates":
		return updates.Render(theme, state)
	case "support":
		return support.Render(theme, state)
	default:
		return ""
	}
}

// HasDateRow reports whether the view shows the shared date's header row.
func HasDateRow(state types.ContentState) bool {
	switch state.View {
	case "daily", "momentum", "review":
		return true
	}
	return false
}

// DateNav is the shared date's navigation for the date row, shortened to fit room.
func DateNav(theme types.Theme, state types.ContentState, room int) string {
	switch state.View {
	case "daily":
		return daily.SubTabMeta(theme, state, room)
	case "momentum":
		return viewcalendar.DateNav(
			theme,
			state.MomentumDate,
			"",
			state.Settings,
			state.WeekStart,
			room,
		)
	case "review":
		return viewcalendar.DateNav(
			theme,
			state.RollupEndDate,
			state.RollupStartDate,
			state.Settings,
			state.WeekStart,
			room,
		)
	}
	return ""
}

// DateCalendarKey opens the view's date picker; it leads the date navigation.
func DateCalendarKey(theme types.Theme, state types.ContentState) string {
	switch state.View {
	case "daily":
		return theme.StyleHeader.Render("C") + theme.StyleDim.Render(" calendar")
	case "review":
		return theme.StyleHeader.Render("E") + theme.StyleDim.Render(" calendar")
	}
	return ""
}
