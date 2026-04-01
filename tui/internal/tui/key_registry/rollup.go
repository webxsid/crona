package keyregistry

import tea "github.com/charmbracelet/bubbletea"

func RegisterRollup[M tea.Model, V comparable, P comparable](r *Router[M, V, P], view V, prevStart, nextStart, prevEnd, nextEnd, reset Handler[M]) {
	r.RegisterView(view, "h", prevStart)
	r.RegisterView(view, "l", nextStart)
	r.RegisterView(view, ",", prevEnd)
	r.RegisterView(view, ".", nextEnd)
	r.RegisterView(view, "g", reset)
}
