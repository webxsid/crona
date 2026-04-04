package ipc

import (
	"context"
	"encoding/json"
	"net"
	"testing"
	"time"

	runtimepkg "crona/kernel/internal/runtime"
	"crona/shared/protocol"
)

type streamOnlyHandler struct {
	started chan struct{}
	stopped chan struct{}
}

func (h *streamOnlyHandler) Handle(ctx context.Context, req protocol.Request) protocol.Response {
	return protocol.Response{ID: req.ID}
}

func (h *streamOnlyHandler) Stream(ctx context.Context, req protocol.Request, writer *json.Encoder) error {
	close(h.started)
	<-ctx.Done()
	close(h.stopped)
	return ctx.Err()
}

func TestServerCloseCancelsActiveEventStream(t *testing.T) {
	handler := &streamOnlyHandler{
		started: make(chan struct{}),
		stopped: make(chan struct{}),
	}
	logger := runtimepkg.NewLogger(runtimepkg.Paths{CurrentLogDir: t.TempDir()})
	server := NewServer("", "", handler, logger)

	serverConn, clientConn := net.Pipe()
	server.wg.Add(1)
	go server.handleConn(serverConn)
	defer func() { _ = clientConn.Close() }()

	reqBody, err := json.Marshal(protocol.Request{
		ID:     "stream-1",
		Method: protocol.MethodEventsSubscribe,
	})
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	if _, err := clientConn.Write(append(reqBody, '\n')); err != nil {
		t.Fatalf("write request: %v", err)
	}

	select {
	case <-handler.started:
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for event stream to start")
	}

	closeDone := make(chan error, 1)
	go func() {
		closeDone <- server.Close()
	}()

	select {
	case err := <-closeDone:
		if err != nil {
			t.Fatalf("server close: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("server.Close did not return")
	}

	select {
	case <-handler.stopped:
	case <-time.After(1 * time.Second):
		t.Fatal("event stream did not stop on close")
	}
}
