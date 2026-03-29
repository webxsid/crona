package keyregistry

import tea "github.com/charmbracelet/bubbletea"

func RegisterSettings[M tea.Model, V comparable, P comparable](r *Router[M, V, P], view V, one, left, right, enter, space Handler[M]) {
	r.RegisterView(view, "1", one)
	r.RegisterView(view, "left", left)
	r.RegisterView(view, "h", left)
	r.RegisterView(view, "right", right)
	r.RegisterView(view, "l", right)
	r.RegisterView(view, "enter", enter)
	r.RegisterView(view, " ", space)
}
