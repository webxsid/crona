package keyregistry

import tea "github.com/charmbracelet/bubbletea"

func RegisterUpdates[M tea.Model, V comparable, P comparable](r *Router[M, V, P], view V, refresh, install, dismiss, open Handler[M]) {
	r.RegisterView(view, "r", refresh)
	r.RegisterView(view, "i", install)
	r.RegisterView(view, "U", dismiss)
	r.RegisterView(view, "o", open)
}
