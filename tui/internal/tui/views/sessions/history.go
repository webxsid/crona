package sessions

import types "crona.work/tui/internal/tui/views/types"

func renderHistoryView(theme types.Theme, state types.ContentState) string {
	return renderLog(theme, state)
}
