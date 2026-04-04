package input

import (
	"time"

	navigationutil "crona/tui/internal/tui/navigationutil"

	tea "github.com/charmbracelet/bubbletea"
)

func handleShiftRollupStartDate(s State, deps Deps, delta int) (tea.Model, tea.Cmd, bool) {
	start := shiftInputISODate(deps.CurrentRollupStartDate(s), delta)
	end := deps.CurrentRollupEndDate(s)
	if start > end {
		end = start
	}
	s.RollupStartDate = start
	s.RollupEndDate = end
	return s, deps.LoadRollupSummaries(start, end), true
}

func handleShiftRollupEndDate(s State, deps Deps, delta int) (tea.Model, tea.Cmd, bool) {
	end := shiftInputISODate(deps.CurrentRollupEndDate(s), delta)
	start := deps.CurrentRollupStartDate(s)
	if start > end {
		start = end
	}
	s.RollupStartDate = start
	s.RollupEndDate = end
	return s, deps.LoadRollupSummaries(start, end), true
}

func handleResetRollupRange(s State, deps Deps) (tea.Model, tea.Cmd, bool) {
	end := time.Now().Format("2006-01-02")
	start := shiftInputISODate(end, -6)
	s.RollupStartDate = start
	s.RollupEndDate = end
	return s, deps.LoadRollupSummaries(start, end), true
}

func shiftInputISODate(date string, days int) string {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return date
	}
	return t.AddDate(0, 0, days).Format("2006-01-02")
}

func handleShiftDailyDate(s State, deps Deps, dir int) (tea.Model, tea.Cmd, bool) {
	s.DashboardDate = navigationutil.ShiftISODate(deps.CurrentDashboardDate(s), dir)
	return s, tea.Batch(deps.LoadDailySummary(s.DashboardDate), deps.LoadDueHabits(s.DashboardDate)), true
}

func handleResetDailyDate(s State, deps Deps) (tea.Model, tea.Cmd, bool) {
	s.DashboardDate = ""
	return s, tea.Batch(deps.LoadDailySummary(""), deps.LoadDueHabits(deps.CurrentDashboardDate(s))), true
}

func handleShiftWellbeingDate(s State, deps Deps, dir int) (tea.Model, tea.Cmd, bool) {
	s.WellbeingDate = navigationutil.ShiftISODate(deps.CurrentWellbeingDate(s), dir)
	return s, deps.LoadWellbeing(s.WellbeingDate), true
}

func handleResetWellbeingDate(s State, deps Deps) (tea.Model, tea.Cmd, bool) {
	s.WellbeingDate = ""
	return s, deps.LoadWellbeing(deps.CurrentWellbeingDate(s)), true
}
