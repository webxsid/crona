package viewrenderer

import (
	alerts "crona/tui/internal/tui/views/alerts"
	away "crona/tui/internal/tui/views/away"
	config "crona/tui/internal/tui/views/config"
	daily "crona/tui/internal/tui/views/daily"
	issues "crona/tui/internal/tui/views/issues"
	habits "crona/tui/internal/tui/views/habits"
	meta "crona/tui/internal/tui/views/meta"
	ops "crona/tui/internal/tui/views/ops"
	reports "crona/tui/internal/tui/views/reports"
	rollup "crona/tui/internal/tui/views/rollup"
	scratchpads "crona/tui/internal/tui/views/scratchpads"
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
	case "meta":
		return meta.Render(theme, state)
	case "session_history", "session_active":
		return sessions.Render(theme, state)
	case "habit_history":
		return habits.Render(theme, state)
	case "scratchpads":
		return scratchpads.Render(theme, state)
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
