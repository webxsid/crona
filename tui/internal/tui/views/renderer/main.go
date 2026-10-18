package viewrenderer

import (
	alerts "crona/tui/internal/tui/views/alerts"
	away "crona/tui/internal/tui/views/away"
	viewcalendar "crona/tui/internal/tui/views/calendar"
	viewchrome "crona/tui/internal/tui/views/chrome"
	config "crona/tui/internal/tui/views/config"
	daily "crona/tui/internal/tui/views/daily"
	habits "crona/tui/internal/tui/views/habits"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	issues "crona/tui/internal/tui/views/issues"
	meta "crona/tui/internal/tui/views/meta"
	momentumview "crona/tui/internal/tui/views/momentumview"
	ops "crona/tui/internal/tui/views/ops"
	reports "crona/tui/internal/tui/views/reports"
	rollup "crona/tui/internal/tui/views/rollup"
	sessions "crona/tui/internal/tui/views/sessions"
	settings "crona/tui/internal/tui/views/settings"
	support "crona/tui/internal/tui/views/support"
	types "crona/tui/internal/tui/views/types"
	updates "crona/tui/internal/tui/views/updates"
	wellbeing "crona/tui/internal/tui/views/wellbeing"
)

func RenderContent(theme types.Theme, state types.ContentState) string {
	switch state.View {
	case "away":
		return away.Render(theme, state)
	case "default":
		return issues.Render(theme, state)
	case "daily":
		return daily.Render(theme, state)
	case "rollup":
		return rollup.Render(theme, state)
	case "momentum":
		return momentumview.Render(theme, state)
	case "meta":
		return meta.Render(theme, state)
	case "session_history", "session_active":
		return sessions.Render(theme, state)
	case "habit_history":
		return habits.Render(theme, state)
	case "ops":
		return ops.Render(theme, state)
	case "wellbeing":
		return wellbeing.Render(theme, state)
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
	case "daily", "momentum", "wellbeing", "rollup":
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
	case "wellbeing":
		return viewcalendar.DateNav(theme, state.WellbeingDate, "", state.Settings, state.WeekStart, room)
	case "rollup":
		return viewcalendar.DateNav(theme, state.RollupEndDate, state.RollupStartDate, state.Settings, state.WeekStart, room)
	}
	return ""
}

// DateRowKeys are the view's own date keys, shown at the right end of the date row.
func DateRowKeys(theme types.Theme, state types.ContentState) []string {
	switch state.View {
	case "daily":
		keys := []string{viewchrome.KeyHint(theme, "C", "calendar")}
		// Without a detail pane, the day summary's key sits here instead.
		if state.HistoricalAwayDate == "" && viewhelpers.ClassFor(state.Width) != viewhelpers.WidthWide {
			keys = append([]string{viewchrome.KeyHint(theme, "S", "day")}, keys...)
		}
		return keys
	case "rollup":
		return []string{viewchrome.KeyHint(theme, "E", "calendar"), viewchrome.KeyHint(theme, "+/-/0", "window")}
	case "momentum", "wellbeing":
		return []string{viewchrome.KeyHint(theme, "+/-/0", "window")}
	}
	return nil
}
