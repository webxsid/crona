package keyregistry

import tea "github.com/charmbracelet/bubbletea"

func RegisterSettings[M tea.Model, V comparable, P comparable](
	r *Router[M, V, P],
	view V,
	left, right, enter, space Handler[M],
) {
	r.RegisterView(view, "left", left)
	r.RegisterView(view, "right", right)
	r.RegisterView(view, "enter", enter)
	r.RegisterView(view, " ", space)
}
