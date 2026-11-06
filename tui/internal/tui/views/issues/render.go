// Package issues renders the scoped issue browser.
package issues

import types "crona.work/tui/internal/tui/views/types"

// Render shows the summary strip, scope selector, issue cards and wide-screen detail.
func Render(theme types.Theme, state types.ContentState) string {
	return renderView(theme, state)
}
