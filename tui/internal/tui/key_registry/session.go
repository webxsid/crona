package keyregistry

import tea "github.com/charmbracelet/bubbletea"

func RegisterSession[M tea.Model, V comparable, P comparable](r *Router[M, V, P], activeView, historyView V, pause, resume, end, stash, context, editHistory, enterHistory Handler[M]) {
	r.RegisterView(activeView, "p", pause)
	r.RegisterView(activeView, "r", resume)
	r.RegisterView(activeView, "x", end)
	r.RegisterView(activeView, "z", stash)
	r.RegisterView(activeView, "i", context)
	r.RegisterView(historyView, "e", editHistory)
	r.RegisterView(historyView, "enter", enterHistory)
}
