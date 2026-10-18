package sessions

import types "crona/tui/internal/tui/views/types"

func renderHistoryView(theme types.Theme, state types.ContentState) string {
	return renderLog(theme, state)
}
