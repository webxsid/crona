package keyregistry

import tea "github.com/charmbracelet/bubbletea"

func RegisterMeta[M tea.Model, V comparable, P comparable](r *Router[M, V, P], view V, repoPane, streamPane, issuePane, habitPane P, one, two, three, four, checkout Handler[M]) {
	r.RegisterView(view, "1", one)
	r.RegisterView(view, "2", two)
	r.RegisterView(view, "3", three)
	r.RegisterView(view, "4", four)
	r.RegisterPane(view, repoPane, "c", checkout)
	r.RegisterPane(view, streamPane, "c", checkout)
	_, _ = issuePane, habitPane
}
