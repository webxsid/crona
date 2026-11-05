package events

import (
	"testing"

	sharedtypes "crona/shared/types"
)

func TestMuteDropsEventsUntilTheOutermostRelease(t *testing.T) {
	bus := NewBus()
	var got []string
	bus.Subscribe(func(event sharedtypes.KernelEvent) { got = append(got, event.Type) })

	release := bus.Mute()
	inner := bus.Mute()
	bus.Emit(sharedtypes.KernelEvent{Type: "issue.created"})
	inner()
	inner()
	bus.Emit(sharedtypes.KernelEvent{Type: "repo.created"})
	if len(got) != 0 {
		t.Fatalf("expected muted events dropped, got %v", got)
	}
	release()
	bus.Emit(sharedtypes.KernelEvent{Type: sharedtypes.EventTypeDataReset})
	if len(got) != 1 || got[0] != sharedtypes.EventTypeDataReset {
		t.Fatalf("expected events to flow after release, got %v", got)
	}
}
