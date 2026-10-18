package input

import (
	"time"

	navigationutil "crona/tui/internal/tui/navigationutil"
	uistate "crona/tui/internal/tui/state"

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
	momentumWindowDaysPresets = []int{7, 14, 30, 90}
	reviewWindowDaysPresets   = []int{7, 14, 30, 90}
)

const defaultMomentumWindowDays = 30

func handleShiftMomentumWindow(s State, deps Deps, dir int) (tea.Model, tea.Cmd, bool) {
	s.MomentumWindowDays = shiftWindowDays(momentumWindowDaysPresets, s.MomentumWindowDays, defaultMomentumWindowDays, dir)
	return s, deps.LoadMomentumRange(deps.CurrentDate(s), s.MomentumWindowDays), true
}

func handleResetMomentumWindow(s State, deps Deps) (tea.Model, tea.Cmd, bool) {
	s.MomentumWindowDays = defaultMomentumWindowDays
	return s, deps.LoadMomentumRange(deps.CurrentDate(s), s.MomentumWindowDays), true
}

func handleShiftReviewWindow(s State, deps Deps, dir int) (tea.Model, tea.Cmd, bool) {
	s.ReviewWindowDays = shiftWindowDays(reviewWindowDaysPresets, s.ReviewWindowDays, uistate.DefaultReviewWindowDays, dir)
	return s, loadReviewWindow(&s, deps), true
}

func handleResetReviewWindow(s State, deps Deps) (tea.Model, tea.Cmd, bool) {
	s.ReviewWindowDays = uistate.DefaultReviewWindowDays
	return s, loadReviewWindow(&s, deps), true
}

// handleReviewStepDay moves the shared date one day inside Review's window; the window
// itself stays put.
func handleReviewStepDay(s State, deps Deps, dir int) (tea.Model, tea.Cmd, bool) {
	date := deps.CurrentDate(s)
	start, end := uistate.ReviewRange(date, s.ReviewWindowEnd, s.ReviewWindowDays)
	s.ReviewWindowEnd = end
	next := shiftInputISODate(date, dir)
	if next < start || next > end {
		return s, nil, true
	}
	s.Date = next
	return s, deps.ReloadDateSurfaces(s), true
}

func handleToggleReviewTime(s State, _ Deps) (tea.Model, tea.Cmd, bool) {
	s.ReviewShowTime = !s.ReviewShowTime
	return s, nil, true
}

// loadReviewWindow re-anchors Review's window on the shared date and loads it.
func loadReviewWindow(s *State, deps Deps) tea.Cmd {
	start, end := uistate.ReviewRange(deps.CurrentDate(*s), s.ReviewWindowEnd, s.ReviewWindowDays)
	s.ReviewWindowEnd = end
	return deps.LoadReviewWindow(start, end)
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
