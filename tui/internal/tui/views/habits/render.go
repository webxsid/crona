package habits

import types "crona/tui/internal/tui/views/types"

func Render(theme types.Theme, state types.ContentState) string {
	return renderHistoryView(theme, state)
}
