package commands

import (
	"bufio"
	"encoding/json"
	"net"
	"runtime"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"crona/shared/localipc"
	"crona/shared/protocol"
	"crona/tui/internal/api"
)

func TestDeleteIssueReloadsTheDailyPlan(t *testing.T) {
	endpoint := testCommandEndpoint()
	ln, err := localipc.Listen(endpoint)
	if err != nil {
		if runtime.GOOS != "windows" && strings.Contains(err.Error(), "operation not permitted") {
			t.Skipf("local ipc listen unavailable in this environment: %v", err)
		}
		t.Fatalf("listen: %v", err)
	}
	defer func() { _ = ln.Close() }()
	methods := make(chan string, 64)
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func(conn net.Conn) {
				defer func() { _ = conn.Close() }()
				var req protocol.Request
				if err := json.NewDecoder(bufio.NewReader(conn)).Decode(&req); err != nil {
					return
				}
				methods <- req.Method
				_ = json.NewEncoder(conn).Encode(protocol.Response{ID: req.ID, Result: replyFor(req.Method)})
			}(conn)
		}
	}()

	client := api.NewClient(localipc.DefaultTransport(), endpoint)
	batch, ok := DeleteIssue(client, 20, 0, "2026-10-18")().(tea.BatchMsg)
	if !ok {
		t.Fatalf("expected DeleteIssue to return a batch")
	}
	runBatchCommands(batch)
	deadline := time.After(3 * time.Second)
	for {
		select {
		case method := <-methods:
			if method == protocol.MethodDailyPlanGet {
				return
			}
		case <-deadline:
			t.Fatalf("DeleteIssue never reloaded the daily plan")
		}
	}
}

func replyFor(method string) json.RawMessage {
	if method == protocol.MethodIssueDelete {
		return json.RawMessage(`{"ok":true}`)
	}
	return json.RawMessage("null")
}
