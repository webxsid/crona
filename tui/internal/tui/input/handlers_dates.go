package input

import (
	"time"

	navigationutil "crona/tui/internal/tui/navigationutil"

	tea "github.com/charmbracelet/bubbletea"
)

// handleShiftDate moves the one shared date every dated view shows and reloads them all.
func handleShiftDate(s State, deps Deps, dir int) (tea.Model, tea.Cmd, bool) {
	s.Date = navigationutil.ShiftISODate(deps.CurrentDate(s), dir)
	return s, deps.ReloadDateSurfaces(s), true
}

// handleResetDate returns the shared date to the daemon's logical today.
func handleResetDate(s State, deps Deps) (tea.Model, tea.Cmd, bool) {
	s.Date = deps.CurrentLogicalDate(s)
	return s, deps.ReloadDateSurfaces(s), true
}

var (
	momentumWindowDaysPresets  = []int{7, 14, 30, 90}
	wellbeingWindowDaysPresets = []int{7, 14, 21, 30}
	rollupWindowDaysPresets    = []int{7, 14, 30, 90}
)

const (
	defaultMomentumWindowDays  = 30
	defaultWellbeingWindowDays = 7
	defaultRollupWindowDays    = 7
)

func handleShiftMomentumWindow(s State, deps Deps, dir int) (tea.Model, tea.Cmd, bool) {
	s.MomentumWindowDays = shiftWindowDays(momentumWindowDaysPresets, s.MomentumWindowDays, defaultMomentumWindowDays, dir)
	return s, deps.LoadMomentumRange(deps.CurrentDate(s), s.MomentumWindowDays), true
}

func handleResetMomentumWindow(s State, deps Deps) (tea.Model, tea.Cmd, bool) {
	s.MomentumWindowDays = defaultMomentumWindowDays
	return s, deps.LoadMomentumRange(deps.CurrentDate(s), s.MomentumWindowDays), true
}

func handleShiftWellbeingWindow(s State, deps Deps, dir int) (tea.Model, tea.Cmd, bool) {
	s.WellbeingWindowDays = shiftWindowDays(wellbeingWindowDaysPresets, s.WellbeingWindowDays, defaultWellbeingWindowDays, dir)
	return s, deps.LoadWellbeing(deps.CurrentDate(s), s.WellbeingWindowDays), true
}

func handleResetWellbeingWindow(s State, deps Deps) (tea.Model, tea.Cmd, bool) {
	s.WellbeingWindowDays = defaultWellbeingWindowDays
	return s, deps.LoadWellbeing(deps.CurrentDate(s), s.WellbeingWindowDays), true
}

func handleShiftRollupWindow(s State, deps Deps, dir int) (tea.Model, tea.Cmd, bool) {
	s.RollupWindowDays = shiftWindowDays(rollupWindowDaysPresets, s.RollupWindowDays, defaultRollupWindowDays, dir)
	return s, loadRollupWindow(s, deps), true
}

func handleResetRollupWindow(s State, deps Deps) (tea.Model, tea.Cmd, bool) {
	s.RollupWindowDays = defaultRollupWindowDays
	return s, loadRollupWindow(s, deps), true
}

// loadRollupWindow loads the rollup range that ends on the shared date.
func loadRollupWindow(s State, deps Deps) tea.Cmd {
	end := deps.CurrentDate(s)
	days := s.RollupWindowDays
	if days < 1 {
		days = defaultRollupWindowDays
	}
	return deps.LoadRollupSummaries(shiftInputISODate(end, -(days-1)), end)
}

func shiftInputISODate(date string, days int) string {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return date
	}
	return t.AddDate(0, 0, days).Format("2006-01-02")
}

// shiftWindowDays steps current to the neighbouring preset in dir, snapping values that
// are not presets to the nearest one first.
func shiftWindowDays(presets []int, current, fallback, dir int) int {
	if current < 1 {
		current = fallback
	}
	current = min(current, presets[len(presets)-1])
	idx := len(presets) - 1
	for i, days := range presets {
		if current <= days {
			idx = i
			break
		}
	}
	if presets[idx] == current {
		idx += dir
	} else if dir < 0 {
		idx--
	}
	return presets[max(0, min(len(presets)-1, idx))]
}
