package keyregistry

import tea "github.com/charmbracelet/bubbletea"

func RegisterGlobal[M tea.Model, V comparable, P comparable](r *Router[M, V, P], bindings map[string]Handler[M]) {
	for key, handler := range bindings {
		r.RegisterGlobal(key, handler)
	}
}
