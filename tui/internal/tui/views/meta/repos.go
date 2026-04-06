package meta

import (
	"crona/tui/internal/api"
	viewchrome "crona/tui/internal/tui/views/chrome"
	types "crona/tui/internal/tui/views/types"
)

func renderRepos(theme types.Theme, state types.ContentState, width, height int) string {
	return viewchrome.RenderSimplePaneWithActions(
		theme,
		"Repos [1]",
		state.Filters["repos"],
		state.Cursors["repos"],
		repoItems(state.Repos),
		state.Pane == "repos",
		width,
		height,
		"No repos — [a] create new",
		paneActions(theme, state, "repos"),
	)
}

func repoItems(repos []api.Repo) []string {
	items := make([]string, 0, len(repos))
	for _, repo := range repos {
		items = append(items, repo.Name)
	}
	return items
}
