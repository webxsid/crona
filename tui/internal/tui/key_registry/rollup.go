package keyregistry

import tea "github.com/charmbracelet/bubbletea"

// RegisterRollup binds Rollup's shared-date keys and its window size keys.
func RegisterRollup[M tea.Model, V comparable, P comparable](
	r *Router[M, V, P],
	view V,
	prevDay, nextDay, today, growWindow, shrinkWindow, resetWindow Handler[M],
) {
	r.RegisterView(view, ",", prevDay)
	r.RegisterView(view, ".", nextDay)
	r.RegisterView(view, "g", today)
	r.RegisterView(view, "+", growWindow)
	r.RegisterView(view, "-", shrinkWindow)
	r.RegisterView(view, "0", resetWindow)
}
