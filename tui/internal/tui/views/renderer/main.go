package viewrenderer

import (
	alerts "crona/tui/internal/tui/views/alerts"
	away "crona/tui/internal/tui/views/away"
	viewcalendar "crona/tui/internal/tui/views/calendar"
	viewchrome "crona/tui/internal/tui/views/chrome"
	config "crona/tui/internal/tui/views/config"
	daily "crona/tui/internal/tui/views/daily"
	habitlist "crona/tui/internal/tui/views/habitlist"
	habits "crona/tui/internal/tui/views/habits"
	issues "crona/tui/internal/tui/views/issues"
	momentumview "crona/tui/internal/tui/views/momentumview"
	ops "crona/tui/internal/tui/views/ops"
	reports "crona/tui/internal/tui/views/reports"
	review "crona/tui/internal/tui/views/review"
	sessions "crona/tui/internal/tui/views/sessions"
	settings "crona/tui/internal/tui/views/settings"
	support "crona/tui/internal/tui/views/support"
	types "crona/tui/internal/tui/views/types"
	updates "crona/tui/internal/tui/views/updates"
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
		return viewcalendar.DateNav(theme, state.MomentumDate, "", state.Settings, state.WeekStart, room)
	case "review":
		return viewcalendar.DateNav(theme, state.RollupEndDate, state.RollupStartDate, state.Settings, state.WeekStart, room)
	}
	return ""
}

// DateCalendarKey opens the view's date picker; it leads the date navigation.
func DateCalendarKey(theme types.Theme, state types.ContentState) string {
	switch state.View {
	case "daily":
		return viewchrome.KeyHint(theme, "C", "calendar")
	case "review":
		return viewchrome.KeyHint(theme, "E", "calendar")
	}
	return ""
}

// DateRowKeys are the view's other date-row keys, shown at the left end of the date row.
func DateRowKeys(theme types.Theme, state types.ContentState) []string {
	switch state.View {
	case "review", "momentum":
		return []string{viewchrome.KeyHint(theme, "+/-/0", "window")}
	}
	return nil
}
