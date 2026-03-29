package keyregistry

import tea "github.com/charmbracelet/bubbletea"

func RegisterOps[M tea.Model, V comparable, P comparable](r *Router[M, V, P], view V, pane P, increase, decrease Handler[M]) {
	r.RegisterPane(view, pane, "+", increase)
	r.RegisterPane(view, pane, "=", increase)
	r.RegisterPane(view, pane, "-", decrease)
}
