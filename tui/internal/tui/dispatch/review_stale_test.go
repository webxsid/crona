package dispatch

import (
	"testing"

	"crona/tui/internal/api"
	uistate "crona/tui/internal/tui/state"

	tea "github.com/charmbracelet/bubbletea"
)

func TestDataEventsOffReviewMarkItStaleInsteadOfLoading(t *testing.T) {
	for _, eventType := range []string{"habit.completed", "issue.updated", "checkin.updated", "session.stopped"} {
		loaded := false
		deps := testEventDeps()
		deps.LoadReviewWindow = func(start, end string) tea.Cmd { loaded = true; return nil }
		state, _ := HandleEvent(EventState{View: uistate.ViewDaily, CurrentDate: "2026-07-14", Cursor: map[uistate.Pane]int{}}, deps, api.KernelEvent{Type: eventType})
		if loaded || !state.ReviewStale {
			t.Fatalf("%s off Review: loaded=%v stale=%v, want no load and stale", eventType, loaded, state.ReviewStale)
		}
	}
}

func TestDataEventsOnReviewReloadIt(t *testing.T) {
	loaded := false
	deps := testEventDeps()
	deps.LoadReviewWindow = func(start, end string) tea.Cmd { loaded = true; return nil }
	state, _ := HandleEvent(EventState{View: uistate.ViewReview, CurrentDate: "2026-07-14", Cursor: map[uistate.Pane]int{}}, deps, api.KernelEvent{Type: "habit.completed"})
	if !loaded || state.ReviewStale {
		t.Fatalf("on Review: loaded=%v stale=%v, want a load and not stale", loaded, state.ReviewStale)
	}
}
