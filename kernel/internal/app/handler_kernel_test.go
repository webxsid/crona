package app

import (
	"context"
	"encoding/json"
	"testing"

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
