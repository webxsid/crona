package dispatch

import (
	"testing"

	"crona/tui/internal/api"
	"crona/tui/internal/tui/commands"

	tea "github.com/charmbracelet/bubbletea"
)

func TestBurstsOfDataEventsReloadOnce(t *testing.T) {
	handled := map[string]int{}
	deps := MessageDeps{
		WaitForEvent: func() tea.Cmd { return nil },
		HandleKernelEvent: func(state MessageState, event api.KernelEvent) (MessageState, tea.Cmd) {
			handled[event.Type]++
			return state, nil
		},
	}
	state := MessageState{}
	for range 500 {
		state, _, _ = HandleMessage(state, commands.KernelEventMsg{Event: api.KernelEvent{Type: "habit.completed"}}, deps)
		state, _, _ = HandleMessage(state, commands.KernelEventMsg{Event: api.KernelEvent{Type: "issue.updated"}}, deps)
	}
	if len(handled) != 0 {
		t.Fatalf("expected data events to wait for the flush, handled %v", handled)
	}
	state, _, _ = HandleMessage(state, FlushEventsMsg{}, deps)
	if handled["habit.completed"] != 1 || handled["issue.updated"] != 1 {
		t.Fatalf("expected one reload per event type, got %v", handled)
	}
	if len(state.PendingEvents) != 0 || state.EventFlushScheduled {
		t.Fatal("expected the buffer to be cleared after the flush")
	}
}

func TestTimerEventsAreHandledAtOnce(t *testing.T) {
	handled := 0
	deps := MessageDeps{
		WaitForEvent: func() tea.Cmd { return nil },
		HandleKernelEvent: func(state MessageState, event api.KernelEvent) (MessageState, tea.Cmd) {
			handled++
			return state, nil
		},
	}
	HandleMessage(MessageState{}, commands.KernelEventMsg{Event: api.KernelEvent{Type: "timer.state"}}, deps)
	if handled != 1 {
		t.Fatalf("expected timer.state to be handled at once, got %d", handled)
	}
}
