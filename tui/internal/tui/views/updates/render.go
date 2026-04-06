package updates

import types "crona/tui/internal/tui/views/types"

func Render(theme types.Theme, state types.ContentState) string {
	return renderView(theme, state)
}
