package input

import (
	"testing"

	uistate "crona/tui/internal/tui/state"

	tea "github.com/charmbracelet/bubbletea"
)

func reviewState(date string) State {
	return State{ActiveView: uistate.ViewReview, ActivePane: uistate.PaneReviewDays, Date: date, ReviewWindowEnd: "2026-10-14", ReviewWindowDays: 7}
}

func TestReviewWindowPresets(t *testing.T) {
	reloads := 0
	var loaded [2]string
	deps := dateDeps(&reloads)
	deps.LoadReviewWindow = func(start, end string) tea.Cmd {
		loaded = [2]string{start, end}
		return nil
	}
	s := reviewState("2026-10-14")
	for _, want := range []int{14, 30, 90, 90} {
		s, _ = Handle(s, runeKey('+'), deps)
		if s.ReviewWindowDays != want {
			t.Fatalf("expected + to step the window to %d, got %d", want, s.ReviewWindowDays)
		}
	}
	if loaded != [2]string{"2026-07-17", "2026-10-14"} {
		t.Fatalf("expected a 90-day window on the anchor, got %v", loaded)
	}
	s, _ = Handle(s, runeKey('-'), deps)
	if s.ReviewWindowDays != 30 {
		t.Fatalf("expected - to step back to 30, got %d", s.ReviewWindowDays)
	}
	s, _ = Handle(s, runeKey('0'), deps)
	if s.ReviewWindowDays != 7 || loaded != [2]string{"2026-10-08", "2026-10-14"} {
		t.Fatalf("expected 0 to reset to 7 days, got %d %v", s.ReviewWindowDays, loaded)
	}
}

func TestReviewStepDayStaysInWindow(t *testing.T) {
	reloads := 0
	s, _ := Handle(reviewState("2026-10-08"), tea.KeyMsg{Type: tea.KeyUp}, dateDeps(&reloads))
	if s.Date != "2026-10-08" {
		t.Fatalf("expected up on the first day to stay, got %s", s.Date)
	}
	s, _ = Handle(s, tea.KeyMsg{Type: tea.KeyDown}, dateDeps(&reloads))
	if s.Date != "2026-10-09" || s.ReviewWindowEnd != "2026-10-14" || reloads != 1 {
		t.Fatalf("expected down to move one day inside the window, got %s end %s (%d reloads)", s.Date, s.ReviewWindowEnd, reloads)
	}
	s, _ = Handle(reviewState("2026-10-14"), tea.KeyMsg{Type: tea.KeyDown}, dateDeps(&reloads))
	if s.Date != "2026-10-14" {
		t.Fatalf("expected down on the last day to stay, got %s", s.Date)
	}
}

func TestReviewToggleTime(t *testing.T) {
	reloads := 0
	s, _ := Handle(reviewState("2026-10-14"), runeKey('S'), dateDeps(&reloads))
	if !s.ReviewShowTime {
		t.Fatalf("expected S to show Time")
	}
	s, _ = Handle(s, runeKey('S'), dateDeps(&reloads))
	if s.ReviewShowTime {
		t.Fatalf("expected S again to show Trends")
	}
}

func TestReviewEOpensDateDialog(t *testing.T) {
	reloads, opened := 0, false
	deps := dateDeps(&reloads)
	deps.OpenDateDialog = func(*State) bool { opened = true; return true }
	deps.OpenExportDailyDialog = func(*State) bool { t.Fatalf("E must not export on Review"); return true }
	Handle(reviewState("2026-10-14"), runeKey('E'), deps)
	if !opened {
		t.Fatalf("expected E to open the calendar")
	}
}
