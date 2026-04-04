package api

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	shareddto "crona/shared/dto"
	"crona/shared/localipc"
	"crona/shared/protocol"
)

func testEndpoint(t *testing.T) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		return `\\.\pipe\crona-test-shutdown-` + time.Now().Format("150405.000000000")
	}
	return fmt.Sprintf("/tmp/crona-test-shutdown-%d.sock", time.Now().UnixNano())
}

func serveShutdownThenClose(t *testing.T, endpoint string, closeDelay time.Duration) net.Listener {
	t.Helper()
	ln, err := localipc.Listen(endpoint)
	if err != nil {
		if runtime.GOOS != "windows" && strings.Contains(err.Error(), "operation not permitted") {
			t.Skipf("local ipc listen unavailable in this environment: %v", err)
		}
		t.Fatalf("listen: %v", err)
	}

	var closeOnce sync.Once
	closeListener := func() { _ = ln.Close() }

	go func() {
		defer closeOnce.Do(closeListener)
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
				var result any
				switch req.Method {
				case protocol.MethodKernelShutdown:
					result = shareddto.OKResponse{OK: true}
					time.AfterFunc(closeDelay, func() { closeOnce.Do(closeListener) })
				case protocol.MethodHealthGet:
					result = Health{Status: "ok", DB: true, OK: 1}
				default:
					_ = json.NewEncoder(conn).Encode(protocol.Response{
						ID: req.ID,
						Error: &protocol.Error{
							Code:    "unsupported",
							Message: req.Method,
						},
					})
					return
				}
				body, err := json.Marshal(result)
				if err != nil {
					return
				}
				_ = json.NewEncoder(conn).Encode(protocol.Response{ID: req.ID, Result: body})
			}(conn)
		}
	}()
	return ln
}

func TestDecodeSettingsReadsBoundarySettingsFromPublicShape(t *testing.T) {
	raw, err := json.Marshal(map[string]any{
		"local": map[string]any{
			"userId":                       "local",
			"deviceId":                     "device-1",
			"timerMode":                    "structured",
			"breaksEnabled":                true,
			"workDurationMinutes":          25,
			"shortBreakMinutes":            5,
			"longBreakMinutes":             15,
			"longBreakEnabled":             true,
			"cyclesBeforeLongBreak":        4,
			"autoStartBreaks":              false,
			"autoStartWork":                false,
			"boundaryNotificationsEnabled": false,
			"boundarySoundEnabled":         true,
			"updateChecksEnabled":          true,
			"updatePromptEnabled":          true,
			"repoSort":                     "chronological_asc",
			"streamSort":                   "chronological_asc",
			"issueSort":                    "priority",
			"habitSort":                    "schedule",
			"createdAt":                    "1",
			"updatedAt":                    "2",
		},
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	settings, err := decodeSettings(raw)
	if err != nil {
		t.Fatalf("decode settings: %v", err)
	}
	if settings == nil {
		t.Fatalf("expected settings, got nil")
	}
	if settings.BoundaryNotifications {
		t.Fatalf("expected boundary notifications false, got true")
	}
	if !settings.BoundarySound {
		t.Fatalf("expected boundary sound true, got false")
	}
}

func TestShutdownKernelAndWaitReturnsAfterKernelStops(t *testing.T) {
	endpoint := testEndpoint(t)
	ln := serveShutdownThenClose(t, endpoint, 150*time.Millisecond)
	defer func() { _ = ln.Close() }()

	client := NewClient(localipc.DefaultTransport(), endpoint, "")
	start := time.Now()
	if err := client.ShutdownKernelAndWait(2 * time.Second); err != nil {
		t.Fatalf("ShutdownKernelAndWait: %v", err)
	}
	if time.Since(start) < 100*time.Millisecond {
		t.Fatalf("expected ShutdownKernelAndWait to wait for shutdown confirmation")
	}
}
