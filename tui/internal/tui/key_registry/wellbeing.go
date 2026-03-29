package keyregistry

import tea "github.com/charmbracelet/bubbletea"

func RegisterWellbeing[M tea.Model, V comparable, P comparable](r *Router[M, V, P], view V, prevDay, nextDay, today Handler[M]) {
	r.RegisterView(view, ",", prevDay)
	r.RegisterView(view, ".", nextDay)
	r.RegisterView(view, "g", today)
}
