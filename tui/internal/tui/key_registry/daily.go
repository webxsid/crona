package keyregistry

import tea "github.com/charmbracelet/bubbletea"

func RegisterDaily[M tea.Model, V comparable, P comparable](r *Router[M, V, P], view V, issuesPane, habitsPane P, one, two, prevDay, nextDay, today, exportDaily, checkout, editHabit, logHabit, failHabit, toggleHabit Handler[M]) {
	r.RegisterView(view, "1", one)
	r.RegisterView(view, "2", two)
	r.RegisterView(view, ",", prevDay)
	r.RegisterView(view, ".", nextDay)
	r.RegisterView(view, "g", today)
	r.RegisterView(view, "E", exportDaily)
	r.RegisterView(view, "c", checkout)
	r.RegisterPane(view, habitsPane, "e", editHabit)
	r.RegisterPane(view, habitsPane, "m", logHabit)
	r.RegisterPane(view, habitsPane, "F", failHabit)
	r.RegisterPane(view, habitsPane, "x", toggleHabit)
	_ = issuesPane
}
