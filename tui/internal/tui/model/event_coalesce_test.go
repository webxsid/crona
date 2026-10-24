package model

import (
	"testing"

	"crona/tui/internal/api"
	"crona/tui/internal/tui/commands"
)

func TestModelBuffersBurstsOfDataEvents(t *testing.T) {
	m := Model{width: 120, height: 36, cursor: map[Pane]int{}, filters: map[Pane]string{}}
	for range 50 {
		next, _ := m.Update(
			commands.KernelEventMsg{Event: api.KernelEvent{Type: "habit.completed"}},
		)
		m = next.(Model)
	}
	if len(m.pendingEvents) != 1 || !m.eventFlushScheduled {
		t.Fatalf(
			"expected one buffered event and a scheduled flush, got %d pending, scheduled=%v",
			len(m.pendingEvents),
			m.eventFlushScheduled,
		)
	}
}

func TestOpeningStaleReviewReloadsIt(t *testing.T) {
	m := Model{
		width:       120,
		height:      36,
		cursor:      map[Pane]int{},
		filters:     map[Pane]string{},
		view:        ViewReview,
		reviewStale: true,
	}
	next, cmd := m.Update(commands.HealthTickMsg{})
	if next.(Model).reviewStale || cmd == nil {
		t.Fatalf(
			"expected the stale Review to reload, stale=%v cmd=%v",
			next.(Model).reviewStale,
			cmd != nil,
		)
	}
}
