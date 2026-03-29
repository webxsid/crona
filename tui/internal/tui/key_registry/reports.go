package keyregistry

import tea "github.com/charmbracelet/bubbletea"

func RegisterReports[M tea.Model, V comparable, P comparable](r *Router[M, V, P], view V, pane P, one, openEditor, open, enter, delete Handler[M]) {
	r.RegisterView(view, "1", one)
	r.RegisterPane(view, pane, "e", openEditor)
	r.RegisterPane(view, pane, "o", open)
	r.RegisterPane(view, pane, "enter", enter)
	r.RegisterPane(view, pane, "d", delete)
}
