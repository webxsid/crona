package app

import (
	"context"
	"encoding/json"
	"slices"
	"testing"

	"crona/kernel/internal/events"
	"crona/kernel/internal/runtime"
	"crona/shared/protocol"
	sharedtypes "crona/shared/types"
)

func TestKernelInfoIncludesProtocolVersion(t *testing.T) {
	handler := NewHandler(
		"2026-04-08T00:00:00Z",
		sharedtypes.KernelInfo{
			PID:             42,
			ProtocolVersion: protocol.Version,
		},
		nil,
		nil,
		nil,
		nil,
		"",
		runtime.Paths{},
		nil,
		nil,
		nil,
	)

	resp := handler.Handle(context.Background(), protocol.Request{
		ID:     "req-1",
		Method: protocol.MethodKernelInfoGet,
	})
	if resp.Error != nil {
		t.Fatalf("unexpected error: %s", resp.Error.Message)
	}

	var info sharedtypes.KernelInfo
	if err := json.Unmarshal(resp.Result, &info); err != nil {
		t.Fatalf("unmarshal kernel info: %v", err)
	}
	if info.ProtocolVersion != protocol.Version {
		t.Fatalf("expected protocol version %q, got %q", protocol.Version, info.ProtocolVersion)
	}
}

func TestEmitSettingsChangedSortsAndDeduplicatesKeys(t *testing.T) {
	bus := events.NewBus()
	handler := &Handler{bus: bus}

	var emitted []sharedtypes.KernelEvent
	unsubscribe := bus.Subscribe(func(event sharedtypes.KernelEvent) {
		emitted = append(emitted, event)
	})
	defer unsubscribe()

	handler.emitSettingsChanged(
		sharedtypes.CoreSettingsKeyWorkDurationMinutes,
		sharedtypes.CoreSettingsKeyAwayDates,
		sharedtypes.CoreSettingsKeyAwayDates,
	)

	if len(emitted) != 1 {
		t.Fatalf("expected one settings.changed event, got %d", len(emitted))
	}
	if emitted[0].Type != sharedtypes.EventTypeSettingsChanged {
		t.Fatalf("expected settings.changed event, got %q", emitted[0].Type)
	}
	var payload sharedtypes.SettingsChangedPayload
	if err := json.Unmarshal(emitted[0].Payload, &payload); err != nil {
		t.Fatalf("decode settings.changed payload: %v", err)
	}
	want := []sharedtypes.CoreSettingsKey{
		sharedtypes.CoreSettingsKeyAwayDates,
		sharedtypes.CoreSettingsKeyWorkDurationMinutes,
	}
	if !slices.Equal(payload.Keys, want) {
		t.Fatalf("expected keys %v, got %v", want, payload.Keys)
	}
}
