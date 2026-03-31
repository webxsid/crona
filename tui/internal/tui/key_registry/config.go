package keyregistry

import tea "github.com/charmbracelet/bubbletea"

func RegisterConfig[M tea.Model, V comparable, P comparable](r *Router[M, V, P], view V, pane P, one, rescan, selectDir, openEditor, reset Handler[M]) {
	r.RegisterView(view, "1", one)
	r.RegisterView(view, "R", rescan)
	r.RegisterPane(view, pane, " ", selectDir)
	r.RegisterPane(view, pane, "e", openEditor)
	r.RegisterView(view, "r", reset)
}
