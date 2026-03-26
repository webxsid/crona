package e2e

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"crona/kernel/internal/app"
	"crona/kernel/internal/runtime"
	shareddto "crona/shared/dto"
	"crona/shared/localipc"
	"crona/shared/protocol"
	sharedtypes "crona/shared/types"
)

type testKernel struct {
	home      string
	info      *sharedtypes.KernelInfo
	cancel    context.CancelFunc
	errCh     chan error
	requestID atomic.Uint64
}

func startTestKernel(t *testing.T) *testKernel {
	t.Helper()

	home, err := os.MkdirTemp("/tmp", "crona-e2e-*")
	if err != nil {
		t.Fatalf("create temp home: %v", err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(home)
	})
	t.Setenv("HOME", home)
	t.Setenv("CRONA_HOME", filepath.Join(home, "crona-runtime"))

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		errCh <- app.Run(ctx)
	}()

	paths, err := runtime.ResolvePaths()
	if err != nil {
		cancel()
		t.Fatalf("resolve paths: %v", err)
	}
	if err := waitForFile(paths.InfoFile, 5*time.Second); err != nil {
		select {
		case runErr := <-errCh:
			cancel()
			t.Fatalf("kernel exited before info file: %v", runErr)
		default:
		}
		cancel()
		t.Fatalf("wait for kernel info: %v", err)
	}

	info, err := runtime.ReadKernelInfo(paths)
	if err != nil {
		cancel()
		t.Fatalf("read kernel info: %v", err)
	}
	normalizeKernelInfo(info)
	if err := waitForHealthyKernel(info, 5*time.Second); err != nil {
		select {
		case runErr := <-errCh:
			cancel()
			t.Fatalf("kernel exited before endpoint became healthy: %v", runErr)
		default:
		}
		cancel()
		t.Fatalf("wait for kernel endpoint: %v", err)
	}

	return &testKernel{
		home:   home,
		info:   info,
		cancel: cancel,
		errCh:  errCh,
	}
}

func (k *testKernel) close(t *testing.T) {
	t.Helper()
	k.cancel()
	select {
	case err := <-k.errCh:
		if err != nil {
			t.Fatalf("kernel shutdown: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("kernel did not shut down within timeout")
	}
}

func (k *testKernel) call(t *testing.T, method string, params any, out any) {
	t.Helper()

	conn, err := localipc.Dial(kernelEndpoint(k.info), 3*time.Second)
	if err != nil {
		t.Fatalf("dial kernel endpoint: %v", err)
	}
	defer func() {
		_ = conn.Close()
	}()
	_ = conn.SetDeadline(time.Now().Add(5 * time.Second))

	var raw json.RawMessage
	if params != nil {
		body, err := json.Marshal(params)
		if err != nil {
			t.Fatalf("marshal params for %s: %v", method, err)
		}
		raw = body
	}

	reqID := k.requestID.Add(1)
	reqBody, err := json.Marshal(protocol.Request{
		ID:     "test-" + strconv.FormatUint(reqID, 10),
		Method: method,
		Params: raw,
	})
	if err != nil {
		t.Fatalf("marshal request for %s: %v", method, err)
	}
	if _, err := conn.Write(append(reqBody, '\n')); err != nil {
		t.Fatalf("write request for %s: %v", method, err)
	}

	var resp protocol.Response
	if err := json.NewDecoder(bufio.NewReader(conn)).Decode(&resp); err != nil {
		t.Fatalf("decode response for %s: %v", method, err)
	}
	if resp.Error != nil {
		t.Fatalf("%s failed: %s: %s", method, resp.Error.Code, resp.Error.Message)
	}
	if out == nil || len(resp.Result) == 0 {
		return
	}
	if err := json.Unmarshal(resp.Result, out); err != nil {
		t.Fatalf("decode result for %s: %v", method, err)
	}
}

func (k *testKernel) callError(t *testing.T, method string, params any) string {
	t.Helper()

	conn, err := localipc.Dial(kernelEndpoint(k.info), 3*time.Second)
	if err != nil {
		t.Fatalf("dial kernel endpoint: %v", err)
	}
	defer func() {
		_ = conn.Close()
	}()
	_ = conn.SetDeadline(time.Now().Add(5 * time.Second))

	var raw json.RawMessage
	if params != nil {
		body, err := json.Marshal(params)
		if err != nil {
			t.Fatalf("marshal params for %s: %v", method, err)
		}
		raw = body
	}

	reqID := k.requestID.Add(1)
	reqBody, err := json.Marshal(protocol.Request{
		ID:     "test-" + strconv.FormatUint(reqID, 10),
		Method: method,
		Params: raw,
	})
	if err != nil {
		t.Fatalf("marshal request for %s: %v", method, err)
	}
	if _, err := conn.Write(append(reqBody, '\n')); err != nil {
		t.Fatalf("write request for %s: %v", method, err)
	}

	var resp protocol.Response
	if err := json.NewDecoder(bufio.NewReader(conn)).Decode(&resp); err != nil {
		t.Fatalf("decode response for %s: %v", method, err)
	}
	if resp.Error == nil {
		t.Fatalf("%s unexpectedly succeeded", method)
	}
	return resp.Error.Message
}

func waitForFile(path string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(path); err == nil {
			return nil
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		time.Sleep(25 * time.Millisecond)
	}
	return errors.New("timeout waiting for " + filepath.Base(path))
}

func waitForHealthyKernel(info *sharedtypes.KernelInfo, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		conn, err := localipc.Dial(kernelEndpoint(info), 250*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return nil
		}
		time.Sleep(25 * time.Millisecond)
	}
	return errors.New("timeout waiting for kernel endpoint")
}

func normalizeKernelInfo(info *sharedtypes.KernelInfo) {
	if info == nil {
		return
	}
	if info.Transport == "" {
		info.Transport = localipc.DefaultTransport()
	}
	if info.Endpoint == "" {
		info.Endpoint = info.SocketPath
	}
	if info.SocketPath == "" && info.Transport == localipc.TransportUnixSocket {
		info.SocketPath = info.Endpoint
	}
}

func kernelEndpoint(info *sharedtypes.KernelInfo) string {
	if info == nil {
		return ""
	}
	if info.Endpoint != "" {
		return info.Endpoint
	}
	return info.SocketPath
}

func createRepo(t *testing.T, k *testKernel, name string) sharedtypes.Repo {
	t.Helper()
	var repo sharedtypes.Repo
	k.call(t, protocol.MethodRepoCreate, shareddto.CreateRepoRequest{Name: name}, &repo)
	return repo
}

func createStream(t *testing.T, k *testKernel, repoID int64, name string) sharedtypes.Stream {
	t.Helper()
	var stream sharedtypes.Stream
	k.call(t, protocol.MethodStreamCreate, shareddto.CreateStreamRequest{
		RepoID: repoID,
		Name:   name,
	}, &stream)
	return stream
}

func createIssue(t *testing.T, k *testKernel, streamID int64, title string, estimate *int) sharedtypes.Issue {
	t.Helper()
	var issue sharedtypes.Issue
	k.call(t, protocol.MethodIssueCreate, shareddto.CreateIssueRequest{
		StreamID:        streamID,
		Title:           title,
		EstimateMinutes: estimate,
	}, &issue)
	return issue
}

func createHabit(t *testing.T, k *testKernel, streamID int64, name string) sharedtypes.Habit {
	t.Helper()
	var habit sharedtypes.Habit
	k.call(t, protocol.MethodHabitCreate, shareddto.CreateHabitRequest{
		StreamID:     streamID,
		Name:         name,
		ScheduleType: "daily",
	}, &habit)
	return habit
}

func changeIssueStatus(t *testing.T, k *testKernel, issueID int64, status sharedtypes.IssueStatus) sharedtypes.Issue {
	t.Helper()
	var issue sharedtypes.Issue
	k.call(t, protocol.MethodIssueChangeStatus, shareddto.ChangeIssueStatusRequest{
		ID:     issueID,
		Status: status,
	}, &issue)
	return issue
}

func listIssues(t *testing.T, k *testKernel, streamID int64) []sharedtypes.Issue {
	t.Helper()
	var issues []sharedtypes.Issue
	k.call(t, protocol.MethodIssueList, shareddto.ListIssuesQuery{StreamID: streamID}, &issues)
	return issues
}
