package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"crona/shared/localipc"
	"crona/shared/protocol"
	sharedtypes "crona/shared/types"
)

func TestRunKernelAttachJSONUsesEnsureKernel(t *testing.T) {
	restore := stubCLIEnv()
	defer restore()

	want := &sharedtypes.KernelInfo{
		PID:        42,
		Transport:  localipc.TransportWindowsNamedPipe,
		Endpoint:   `\\.\pipe\crona-kernel`,
		Env:        "Prod",
		ScratchDir: `C:\Users\alice\AppData\Local\Crona\scratch`,
	}
	ensureKernelFn = func() (*sharedtypes.KernelInfo, error) { return want, nil }

	output := captureStdout(t, func() {
		if err := runKernel([]string{"attach", "--json"}); err != nil {
			t.Fatalf("runKernel attach: %v", err)
		}
	})

	var got sharedtypes.KernelInfo
	if err := json.Unmarshal([]byte(output), &got); err != nil {
		t.Fatalf("unmarshal attach output: %v\noutput=%s", err, output)
	}
	if got.PID != want.PID || got.Transport != want.Transport || got.Endpoint != want.Endpoint {
		t.Fatalf("unexpected attach output: %+v", got)
	}
}

func TestRunKernelInfoAndStatusUseKernelInfoMethod(t *testing.T) {
	for _, command := range []string{"info", "status"} {
		t.Run(command, func(t *testing.T) {
			restore := stubCLIEnv()
			defer restore()

			var gotMethod string
			callKernelFn = func(method string, params, out any) error {
				gotMethod = method
				info := out.(*sharedtypes.KernelInfo)
				*info = sharedtypes.KernelInfo{
					PID:        99,
					Transport:  localipc.TransportWindowsNamedPipe,
					Endpoint:   `\\.\pipe\crona-kernel`,
					Env:        "Prod",
					StartedAt:  "2026-03-27T00:00:00Z",
					ScratchDir: `C:\Users\alice\AppData\Local\Crona\scratch`,
				}
				return nil
			}

			output := captureStdout(t, func() {
				if err := runKernel([]string{command, "--json"}); err != nil {
					t.Fatalf("runKernel %s: %v", command, err)
				}
			})

			if gotMethod != protocol.MethodKernelInfoGet {
				t.Fatalf("expected %s, got %s", protocol.MethodKernelInfoGet, gotMethod)
			}

			var got sharedtypes.KernelInfo
			if err := json.Unmarshal([]byte(output), &got); err != nil {
				t.Fatalf("unmarshal %s output: %v\noutput=%s", command, err, output)
			}
			if got.Transport != localipc.TransportWindowsNamedPipe || got.Endpoint != `\\.\pipe\crona-kernel` {
				t.Fatalf("unexpected %s output: %+v", command, got)
			}
		})
	}
}

func TestRunKernelStatusTextIncludesWindowsTransportAndEndpoint(t *testing.T) {
	restore := stubCLIEnv()
	defer restore()

	callKernelFn = func(method string, params, out any) error {
		info := out.(*sharedtypes.KernelInfo)
		*info = sharedtypes.KernelInfo{
			PID:        7,
			Transport:  localipc.TransportWindowsNamedPipe,
			Endpoint:   `\\.\pipe\crona-kernel-dev`,
			Env:        "Dev",
			StartedAt:  "2026-03-27T00:00:00Z",
			ScratchDir: `C:\Users\alice\AppData\Local\Crona Dev\scratch`,
		}
		return nil
	}

	output := captureStdout(t, func() {
		if err := runKernel([]string{"status"}); err != nil {
			t.Fatalf("runKernel status: %v", err)
		}
	})

	for _, want := range []string{
		"transport: windows_named_pipe",
		`endpoint: \\.\pipe\crona-kernel-dev`,
		"env: Dev",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, output)
		}
	}
}

func TestNormalizeKernelInfoPreservesWindowsNamedPipeFields(t *testing.T) {
	info := &sharedtypes.KernelInfo{
		Transport:  localipc.TransportWindowsNamedPipe,
		Endpoint:   `\\.\pipe\crona-kernel`,
		SocketPath: "",
	}

	normalizeKernelInfo(info)

	if info.Transport != localipc.TransportWindowsNamedPipe {
		t.Fatalf("expected windows transport, got %q", info.Transport)
	}
	if info.Endpoint != `\\.\pipe\crona-kernel` {
		t.Fatalf("expected endpoint to be preserved, got %q", info.Endpoint)
	}
	if info.SocketPath != "" {
		t.Fatalf("expected socket path to stay empty for windows transport, got %q", info.SocketPath)
	}
}

func TestKernelLaunchCandidatesIncludeWindowsExeSources(t *testing.T) {
	restore := stubCLIEnv()
	defer restore()

	root := t.TempDir()
	repoRoot := filepath.Join(root, "repo")
	exeDir := filepath.Join(root, "app")
	pathDir := filepath.Join(root, "path")
	if err := os.MkdirAll(filepath.Join(repoRoot, "kernel", "cmd"), 0o755); err != nil {
		t.Fatalf("mkdir repo kernel cmd: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(repoRoot, "bin"), 0o755); err != nil {
		t.Fatalf("mkdir repo bin: %v", err)
	}
	if err := os.MkdirAll(exeDir, 0o755); err != nil {
		t.Fatalf("mkdir exe dir: %v", err)
	}
	if err := os.MkdirAll(pathDir, 0o755); err != nil {
		t.Fatalf("mkdir path dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repoRoot, "go.work"), []byte("go 1.26.1\n"), 0o644); err != nil {
		t.Fatalf("write go.work: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repoRoot, "kernel", "cmd", "crona-kernel"), []byte("stub"), 0o644); err != nil {
		t.Fatalf("write repo kernel marker: %v", err)
	}

	sibling := filepath.Join(exeDir, "crona-kernel.exe")
	pathCandidate := filepath.Join(pathDir, "crona-kernel.exe")
	repoBin := filepath.Join(repoRoot, "bin", "crona-kernel.exe")
	for _, file := range []string{sibling, pathCandidate, repoBin} {
		if err := os.WriteFile(file, []byte("stub"), 0o755); err != nil {
			t.Fatalf("write candidate %s: %v", file, err)
		}
	}

	kernelBinaryFn = func() string { return "crona-kernel.exe" }
	osExecutableFn = func() (string, error) { return filepath.Join(exeDir, "crona.exe"), nil }
	osGetwdFn = func() (string, error) { return repoRoot, nil }
	execLookPathFn = func(file string) (string, error) {
		switch file {
		case "crona-kernel.exe":
			return pathCandidate, nil
		case "go":
			return filepath.Join(pathDir, "go"), nil
		default:
			return "", os.ErrNotExist
		}
	}

	got := kernelLaunchCandidates()
	if len(got) < 3 {
		t.Fatalf("expected multiple windows candidates, got %+v", got)
	}

	names := []string{}
	cmds := []string{}
	for _, candidate := range got {
		names = append(names, candidate.name)
		cmds = append(cmds, candidate.cmd)
	}

	for _, want := range []string{
		"sibling crona-kernel.exe",
		"PATH crona-kernel.exe",
		"repo bin crona-kernel.exe",
	} {
		if !contains(names, want) {
			t.Fatalf("expected candidate %q in %+v", want, names)
		}
	}

	for _, want := range []string{sibling, pathCandidate, repoBin} {
		if !contains(cmds, want) {
			t.Fatalf("expected command %q in %+v", want, cmds)
		}
	}
}

func stubCLIEnv() func() {
	origCallKernel := callKernelFn
	origEnsureKernel := ensureKernelFn
	origRuntimeBaseDir := runtimeBaseDir
	origReadFile := readFileFn
	origKernelBinary := kernelBinaryFn
	origExecutable := osExecutableFn
	origLookPath := execLookPathFn
	origStat := osStatFn
	origGetwd := osGetwdFn
	origStartKernel := startKernelFn
	origSleep := timeSleepFn

	return func() {
		callKernelFn = origCallKernel
		ensureKernelFn = origEnsureKernel
		runtimeBaseDir = origRuntimeBaseDir
		readFileFn = origReadFile
		kernelBinaryFn = origKernelBinary
		osExecutableFn = origExecutable
		execLookPathFn = origLookPath
		osStatFn = origStat
		osGetwdFn = origGetwd
		startKernelFn = origStartKernel
		timeSleepFn = origSleep
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdout = w
	defer func() {
		os.Stdout = oldStdout
	}()

	fn()

	_ = w.Close()
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	return strings.TrimSpace(buf.String())
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
