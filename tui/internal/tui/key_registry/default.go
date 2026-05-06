package keyregistry

import tea "github.com/charmbracelet/bubbletea"

func RegisterDefault[M tea.Model, V comparable, P comparable](r *Router[M, V, P], view V, issuesPane P, openSection Handler[M], completedSection Handler[M], checkout Handler[M]) {
	r.RegisterView(view, "1", openSection)
	r.RegisterView(view, "2", completedSection)
	_ = issuesPane
	_ = checkout
}
