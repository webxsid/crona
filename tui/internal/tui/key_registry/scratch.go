package keyregistry

import tea "github.com/charmbracelet/bubbletea"

func RegisterScratch[M tea.Model, V comparable, P comparable](r *Router[M, V, P], view V, one Handler[M]) {
	r.RegisterView(view, "1", one)
}
