package meta

import (
	"crona/tui/internal/api"
	viewchrome "crona/tui/internal/tui/views/chrome"
	types "crona/tui/internal/tui/views/types"
)

func renderStreams(theme types.Theme, state types.ContentState, width, height int, emptyText string) string {
	return viewchrome.RenderSimplePaneWithActions(
		theme,
		"Streams [2]",
		state.Filters["streams"],
		state.Cursors["streams"],
		streamItems(state.Streams),
		state.Pane == "streams",
		width,
		height,
		emptyText,
		paneActions(theme, state, "streams"),
	)
}

func streamItems(streams []api.Stream) []string {
	items := make([]string, 0, len(streams))
	for _, stream := range streams {
		items = append(items, stream.Name)
	}
	return items
}
