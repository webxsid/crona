package keyregistry

import tea "github.com/charmbracelet/bubbletea"

func RegisterMeta[M tea.Model, V comparable, P comparable](
	r *Router[M, V, P],
	view V,
	repoPane, streamPane, issuePane, habitPane P,
	checkout Handler[M],
) {
	r.RegisterPane(view, repoPane, "c", checkout)
	r.RegisterPane(view, streamPane, "c", checkout)
	_, _ = issuePane, habitPane
}
