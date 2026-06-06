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
	return s, tea.Batch(
		deps.LoadDailySummary(s.DashboardDate),
		deps.LoadDailyStreaks(s.DashboardDate),
		deps.LoadDueHabits(s.DashboardDate),
	), true
}

func handleResetDailyDate(s State, deps Deps) (tea.Model, tea.Cmd, bool) {
	s.DashboardDate = ""
	return s, tea.Batch(
		deps.LoadDailySummary(""),
		deps.LoadDailyStreaks(deps.CurrentDashboardDate(s)),
		deps.LoadDueHabits(deps.CurrentDashboardDate(s)),
	), true
}

func handleShiftMomentumDate(s State, deps Deps, dir int) (tea.Model, tea.Cmd, bool) {
	s.MomentumDate = navigationutil.ShiftISODate(deps.CurrentMomentumDate(s), dir)
	return s, deps.LoadMomentumRange(s.MomentumDate, s.MomentumWindowDays), true
}

func handleResetMomentumDate(s State, deps Deps) (tea.Model, tea.Cmd, bool) {
	s.MomentumDate = ""
	return s, deps.LoadMomentumRange(deps.CurrentMomentumDate(s), s.MomentumWindowDays), true
}

var momentumWindowDaysPresets = []int{7, 14, 30, 90}

func handleShiftMomentumWindow(s State, deps Deps, dir int) (tea.Model, tea.Cmd, bool) {
	s.MomentumWindowDays = shiftMomentumWindowDays(s.MomentumWindowDays, dir)
	return s, deps.LoadMomentumRange(deps.CurrentMomentumDate(s), s.MomentumWindowDays), true
}

func handleResetMomentumWindow(s State, deps Deps) (tea.Model, tea.Cmd, bool) {
	s.MomentumWindowDays = 30
	return s, deps.LoadMomentumRange(deps.CurrentMomentumDate(s), s.MomentumWindowDays), true
}

func shiftMomentumWindowDays(current, dir int) int {
	if current < 1 {
		current = 30
	}
	if current > momentumWindowDaysPresets[len(momentumWindowDaysPresets)-1] {
		current = momentumWindowDaysPresets[len(momentumWindowDaysPresets)-1]
	}
	idx := -1
	for i, days := range momentumWindowDaysPresets {
		if days == current {
			idx = i
			break
		}
	}
	if idx < 0 {
		idx = 0
		for i, days := range momentumWindowDaysPresets {
			if current < days {
				idx = i
				break
			}
			idx = i
		}
	}
	idx += dir
	if idx < 0 {
		idx = 0
	}
	if idx >= len(momentumWindowDaysPresets) {
		idx = len(momentumWindowDaysPresets) - 1
	}
	return momentumWindowDaysPresets[idx]
}

func handleShiftWellbeingDate(s State, deps Deps, dir int) (tea.Model, tea.Cmd, bool) {
	s.WellbeingDate = navigationutil.ShiftISODate(deps.CurrentWellbeingDate(s), dir)
	return s, deps.LoadWellbeing(s.WellbeingDate, s.WellbeingWindowDays), true
}

func handleResetWellbeingDate(s State, deps Deps) (tea.Model, tea.Cmd, bool) {
	s.WellbeingDate = ""
	return s, deps.LoadWellbeing(deps.CurrentWellbeingDate(s), s.WellbeingWindowDays), true
}

var wellbeingWindowDaysPresets = []int{7, 14, 21, 30}

func handleShiftWellbeingWindow(s State, deps Deps, dir int) (tea.Model, tea.Cmd, bool) {
	s.WellbeingWindowDays = shiftWellbeingWindowDays(s.WellbeingWindowDays, dir)
	return s, deps.LoadWellbeing(deps.CurrentWellbeingDate(s), s.WellbeingWindowDays), true
}

func handleResetWellbeingWindow(s State, deps Deps) (tea.Model, tea.Cmd, bool) {
	s.WellbeingWindowDays = 7
	return s, deps.LoadWellbeing(deps.CurrentWellbeingDate(s), s.WellbeingWindowDays), true
}

func shiftWellbeingWindowDays(current, dir int) int {
	if current < 1 {
		current = 7
	}
	if current > wellbeingWindowDaysPresets[len(wellbeingWindowDaysPresets)-1] {
		current = wellbeingWindowDaysPresets[len(wellbeingWindowDaysPresets)-1]
	}
	idx := -1
	for i, days := range wellbeingWindowDaysPresets {
		if days == current {
			idx = i
			break
		}
	}
	if idx < 0 {
		idx = 0
		for i, days := range wellbeingWindowDaysPresets {
			if current < days {
				idx = i
				break
			}
			idx = i
		}
	}
	idx += dir
	if idx < 0 {
		idx = 0
	}
	if idx >= len(wellbeingWindowDaysPresets) {
		idx = len(wellbeingWindowDaysPresets) - 1
	}
	return wellbeingWindowDaysPresets[idx]
}
