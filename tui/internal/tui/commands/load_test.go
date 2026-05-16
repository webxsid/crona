package commands

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

	tea "github.com/charmbracelet/bubbletea"

	shareddto "crona/shared/dto"
	"crona/shared/localipc"
	"crona/shared/protocol"
	"crona/tui/internal/api"
)

func TestLoadWellbeingUsesLifetimeStreaksAndSevenDayMetrics(t *testing.T) {
	endpoint := testCommandEndpoint()
	ln, err := localipc.Listen(endpoint)
	if err != nil {
		if runtime.GOOS != "windows" && strings.Contains(err.Error(), "operation not permitted") {
			t.Skipf("local ipc listen unavailable in this environment: %v", err)
		}
		t.Fatalf("listen: %v", err)
	}
	defer func() { _ = ln.Close() }()

	requests := make(chan protocol.Request, 16)
	var wg sync.WaitGroup
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			wg.Add(1)
			go func(conn net.Conn) {
				defer wg.Done()
				defer func() { _ = conn.Close() }()
				var req protocol.Request
				if err := json.NewDecoder(bufio.NewReader(conn)).Decode(&req); err != nil {
					return
				}
				requests <- req
				body, _ := json.Marshal(resultForWellbeingLoadMethod(req.Method))
				_ = json.NewEncoder(conn).Encode(protocol.Response{ID: req.ID, Result: body})
			}(conn)
		}
	}()

	client := api.NewClient(localipc.DefaultTransport(), endpoint, "")
	msg := LoadWellbeing(client, "2026-04-10")()
	batch, ok := msg.(tea.BatchMsg)
	if !ok {
		t.Fatalf("expected LoadWellbeing to return tea.BatchMsg, got %T", msg)
	}
	done := make(chan struct{})
	go func() {
		runBatchCommands(batch)
		close(done)
	}()

	got := map[string][]protocol.Request{}
	deadline := time.After(3 * time.Second)
	collecting := true
	for collecting {
		select {
		case req := <-requests:
			got[req.Method] = append(got[req.Method], req)
		case <-done:
			collecting = false
		case <-deadline:
			t.Fatalf("timed out waiting for wellbeing load requests, got methods %+v", methodKeys(got))
		}
	}
	for {
		select {
		case req := <-requests:
			got[req.Method] = append(got[req.Method], req)
		default:
			goto drained
		}
	}
drained:
	_ = ln.Close()
	wg.Wait()

	assertDateRangeQuery(t, firstRequest(got, protocol.MethodMetricsRange), "2026-04-04", "2026-04-10")
	assertDateRangeQuery(t, firstRequest(got, protocol.MethodMetricsRollup), "2026-04-04", "2026-04-10")
	assertDateRangeQuery(t, firstRequest(got, protocol.MethodDashboardWindow), "2026-04-04", "2026-04-10")
	assertDateRangeQuery(t, firstRequest(got, protocol.MethodDashboardFocusScore), "2026-04-04", "2026-04-10")
	for _, req := range got[protocol.MethodDashboardDistribution] {
		assertDateRangeQuery(t, req, "2026-04-04", "2026-04-10")
	}
	if len(got[protocol.MethodDashboardDistribution]) == 0 {
		t.Fatalf("expected at least one distribution request")
	}
	assertDateRangeQuery(t, firstRequest(got, protocol.MethodDashboardGoalProgress), "2026-04-04", "2026-04-10")
	if len(got[protocol.MethodMetricsStreaks]) > 0 {
		t.Fatalf("wellbeing load should not call range-based %s", protocol.MethodMetricsStreaks)
	}
	var streakQuery shareddto.DailyCheckInQuery
	if err := json.Unmarshal(firstRequest(got, protocol.MethodMetricsStreaksLifetime).Params, &streakQuery); err != nil {
		t.Fatalf("unmarshal lifetime streak query: %v", err)
	}
	if streakQuery.Date != "2026-04-10" {
		t.Fatalf("expected lifetime streak date 2026-04-10, got %+v", streakQuery)
	}
}

func TestLoadWellbeingWindowUsesRequestedRange(t *testing.T) {
	endpoint := testCommandEndpoint()
	ln, err := localipc.Listen(endpoint)
	if err != nil {
		if runtime.GOOS != "windows" && strings.Contains(err.Error(), "operation not permitted") {
			t.Skipf("local ipc listen unavailable in this environment: %v", err)
		}
		t.Fatalf("listen: %v", err)
	}
	defer func() { _ = ln.Close() }()

	requests := make(chan protocol.Request, 16)
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
				requests <- req
				body, _ := json.Marshal(resultForWellbeingLoadMethod(req.Method))
				_ = json.NewEncoder(conn).Encode(protocol.Response{ID: req.ID, Result: body})
			}(conn)
		}
	}()

	client := api.NewClient(localipc.DefaultTransport(), endpoint, "")
	msg := LoadWellbeingWindow(client, "2026-04-10", 14)()
	batch, ok := msg.(tea.BatchMsg)
	if !ok {
		t.Fatalf("expected LoadWellbeingWindow to return tea.BatchMsg, got %T", msg)
	}
	done := make(chan struct{})
	go func() {
		runBatchCommands(batch)
		close(done)
	}()

	got := map[string][]protocol.Request{}
	deadline := time.After(3 * time.Second)
	collecting := true
	for collecting {
		select {
		case req := <-requests:
			got[req.Method] = append(got[req.Method], req)
		case <-done:
			collecting = false
		case <-deadline:
			t.Fatalf("timed out waiting for wellbeing load requests, got methods %+v", methodKeys(got))
		}
	}
	for {
		select {
		case req := <-requests:
			got[req.Method] = append(got[req.Method], req)
		default:
			goto drained
		}
	}
drained:
	assertDateRangeQuery(t, firstRequest(got, protocol.MethodMetricsRange), "2026-03-28", "2026-04-10")
	assertDateRangeQuery(t, firstRequest(got, protocol.MethodMetricsRollup), "2026-03-28", "2026-04-10")
	assertDateRangeQuery(t, firstRequest(got, protocol.MethodDashboardWindow), "2026-03-28", "2026-04-10")
}

func TestLoadWellbeingWindowCapsAtThirtyDays(t *testing.T) {
	endpoint := testCommandEndpoint()
	ln, err := localipc.Listen(endpoint)
	if err != nil {
		if runtime.GOOS != "windows" && strings.Contains(err.Error(), "operation not permitted") {
			t.Skipf("local ipc listen unavailable in this environment: %v", err)
		}
		t.Fatalf("listen: %v", err)
	}
	defer func() { _ = ln.Close() }()

	requests := make(chan protocol.Request, 16)
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
				requests <- req
				body, _ := json.Marshal(resultForWellbeingLoadMethod(req.Method))
				_ = json.NewEncoder(conn).Encode(protocol.Response{ID: req.ID, Result: body})
			}(conn)
		}
	}()

	client := api.NewClient(localipc.DefaultTransport(), endpoint, "")
	msg := LoadWellbeingWindow(client, "2026-04-10", 60)()
	batch, ok := msg.(tea.BatchMsg)
	if !ok {
		t.Fatalf("expected LoadWellbeingWindow to return tea.BatchMsg, got %T", msg)
	}
	done := make(chan struct{})
	go func() {
		runBatchCommands(batch)
		close(done)
	}()

	got := map[string][]protocol.Request{}
	deadline := time.After(3 * time.Second)
	collecting := true
	for collecting {
		select {
		case req := <-requests:
			got[req.Method] = append(got[req.Method], req)
		case <-done:
			collecting = false
		case <-deadline:
			t.Fatalf("timed out waiting for wellbeing load requests, got methods %+v", methodKeys(got))
		}
	}
	for {
		select {
		case req := <-requests:
			got[req.Method] = append(got[req.Method], req)
		default:
			goto drained
		}
	}
drained:
	assertDateRangeQuery(t, firstRequest(got, protocol.MethodMetricsRange), "2026-03-12", "2026-04-10")
	assertDateRangeQuery(t, firstRequest(got, protocol.MethodMetricsRollup), "2026-03-12", "2026-04-10")
	assertDateRangeQuery(t, firstRequest(got, protocol.MethodDashboardWindow), "2026-03-12", "2026-04-10")
}

func runBatchCommands(batch tea.BatchMsg) {
	var cmdWG sync.WaitGroup
	for _, cmd := range batch {
		if cmd == nil {
			continue
		}
		cmdWG.Add(1)
		go func(cmd tea.Cmd) {
			defer cmdWG.Done()
			if nested, ok := cmd().(tea.BatchMsg); ok {
				runBatchCommands(nested)
			}
		}(cmd)
	}
	cmdWG.Wait()
}

func testCommandEndpoint() string {
	if runtime.GOOS == "windows" {
		return `\\.\pipe\crona-test-load-wellbeing-` + time.Now().Format("150405.000000000")
	}
	return fmt.Sprintf("/tmp/crona-test-load-wellbeing-%d.sock", time.Now().UnixNano())
}

func resultForWellbeingLoadMethod(method string) any {
	switch method {
	case protocol.MethodDailyPlanGet:
		return map[string]any{}
	case protocol.MethodCheckInGet:
		return map[string]any{}
	case protocol.MethodMetricsRange:
		return []any{}
	case protocol.MethodMetricsRollup:
		return map[string]any{}
	case protocol.MethodMetricsStreaksLifetime:
		return map[string]any{}
	case protocol.MethodDashboardWindow:
		return map[string]any{}
	case protocol.MethodDashboardFocusScore:
		return map[string]any{}
	case protocol.MethodDashboardDistribution:
		return map[string]any{}
	case protocol.MethodDashboardGoalProgress:
		return map[string]any{}
	default:
		return map[string]any{}
	}
}

func assertDateRangeQuery(t *testing.T, req protocol.Request, start string, end string) {
	t.Helper()
	if req.Method == "" {
		t.Fatalf("expected date range request for %s to %s", start, end)
	}
	var query shareddto.DateRangeQuery
	if err := json.Unmarshal(req.Params, &query); err != nil {
		t.Fatalf("unmarshal %s query: %v", req.Method, err)
	}
	if query.Start != start || query.End != end {
		t.Fatalf("expected %s query %s to %s, got %+v", req.Method, start, end, query)
	}
}

func firstRequest(requests map[string][]protocol.Request, method string) protocol.Request {
	if len(requests[method]) == 0 {
		return protocol.Request{}
	}
	return requests[method][0]
}

func methodKeys(requests map[string][]protocol.Request) []string {
	keys := make([]string, 0, len(requests))
	for method := range requests {
		keys = append(keys, method)
	}
	return keys
}
