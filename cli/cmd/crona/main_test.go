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

func TestRunNoArgsLaunchesTUI(t *testing.T) {
	restore := stubCLIEnv()
	defer restore()

	called := false
	launchTUIFn = func() error {
		called = true
		return nil
	}

	if err := run(nil); err != nil {
		t.Fatalf("run no args: %v", err)
	}
	if !called {
		t.Fatalf("expected no-arg run to launch TUI")
	}
}

func TestRunHelpFlagStillPrintsUsage(t *testing.T) {
	restore := stubCLIEnv()
	defer restore()

	called := false
	launchTUIFn = func() error {
		called = true
		return nil
	}

	output := captureStdout(t, func() {
		if err := run([]string{"--help"}); err != nil {
			t.Fatalf("run help: %v", err)
		}
	})

	if called {
		t.Fatalf("expected help flag to avoid launching TUI")
	}
	if !strings.Contains(output, "Run without a command to open the TUI.") {
		t.Fatalf("expected updated usage output, got:\n%s", output)
	}
}

func TestRunTimerStartFromContextUsesCheckedOutIssue(t *testing.T) {
	restore := stubCLIEnv()
	defer restore()

	var gotMethods []string
	callKernelFn = func(method string, params, out any) error {
		gotMethods = append(gotMethods, method)
		switch method {
		case protocol.MethodContextGet:
			ctx := out.(*sharedtypes.ActiveContext)
			issueID := int64(77)
			ctx.IssueID = &issueID
			return nil
		case protocol.MethodTimerStart:
			req, ok := params.(struct {
				IssueID *int64 `json:"issueId,omitempty"`
			})
			if !ok || req.IssueID == nil || *req.IssueID != 77 {
				t.Fatalf("unexpected timer start params: %#v", params)
			}
			timer := out.(*sharedtypes.TimerState)
			timer.State = "running"
			timer.ElapsedSeconds = 12
			return nil
		default:
			t.Fatalf("unexpected method: %s", method)
			return nil
		}
	}

	if err := runTimer([]string{"start", "--from-context", "--json"}); err != nil {
		t.Fatalf("runTimer start --from-context: %v", err)
	}
	if len(gotMethods) != 2 || gotMethods[0] != protocol.MethodContextGet || gotMethods[1] != protocol.MethodTimerStart {
		t.Fatalf("unexpected method order: %+v", gotMethods)
	}
}

func TestRunIssueStartFromContextUsesCheckedOutIssue(t *testing.T) {
	restore := stubCLIEnv()
	defer restore()

	callKernelFn = func(method string, params, out any) error {
		switch method {
		case protocol.MethodContextGet:
			ctx := out.(*sharedtypes.ActiveContext)
			issueID := int64(91)
			ctx.IssueID = &issueID
			return nil
		case protocol.MethodTimerStart:
			req, ok := params.(struct {
				IssueID *int64 `json:"issueId,omitempty"`
			})
			if !ok || req.IssueID == nil || *req.IssueID != 91 {
				t.Fatalf("unexpected timer start params: %#v", params)
			}
			timer := out.(*sharedtypes.TimerState)
			timer.State = "running"
			return nil
		default:
			t.Fatalf("unexpected method: %s", method)
			return nil
		}
	}

	if err := runIssue([]string{"start", "--from-context", "--json"}); err != nil {
		t.Fatalf("runIssue start --from-context: %v", err)
	}
}

func TestRunContextSwitchRepoJSONUsesSwitchMethod(t *testing.T) {
	restore := stubCLIEnv()
	defer restore()

	var gotMethod string
	callKernelFn = func(method string, params, out any) error {
		gotMethod = method
		if method != protocol.MethodContextSwitchRepo {
			t.Fatalf("unexpected method: %s", method)
		}
		body, err := json.Marshal(params)
		if err != nil {
			t.Fatalf("marshal params: %v", err)
		}
		var decoded map[string]int64
		if err := json.Unmarshal(body, &decoded); err != nil {
			t.Fatalf("unmarshal params: %v", err)
		}
		if decoded["repoId"] != 12 {
			t.Fatalf("unexpected repoId payload: %#v", decoded)
		}
		ctx := out.(*sharedtypes.ActiveContext)
		repoName := "Work"
		ctx.RepoName = &repoName
		return nil
	}

	output := captureStdout(t, func() {
		if err := runContext([]string{"switch-repo", "--id", "12", "--json"}); err != nil {
			t.Fatalf("runContext switch-repo: %v", err)
		}
	})

	if gotMethod != protocol.MethodContextSwitchRepo {
		t.Fatalf("expected %s, got %s", protocol.MethodContextSwitchRepo, gotMethod)
	}
	if !strings.Contains(output, "\"repoName\": \"Work\"") {
		t.Fatalf("unexpected output: %s", output)
	}
}

func TestRunKernelWipeDataRequiresForce(t *testing.T) {
	restore := stubCLIEnv()
	defer restore()

	if err := runKernel([]string{"wipe-data"}); err == nil || !strings.Contains(err.Error(), "--force") {
		t.Fatalf("expected --force error, got %v", err)
	}
}

func TestRunExportRepoUsesContextRepoWhenMissing(t *testing.T) {
	restore := stubCLIEnv()
	defer restore()

	var gotMethods []string
	callKernelFn = func(method string, params, out any) error {
		gotMethods = append(gotMethods, method)
		switch method {
		case protocol.MethodContextGet:
			ctx := out.(*sharedtypes.ActiveContext)
			repoID := int64(44)
			ctx.RepoID = &repoID
			return nil
		case protocol.MethodExportRepo:
			body, err := json.Marshal(params)
			if err != nil {
				t.Fatalf("marshal export params: %v", err)
			}
			var decoded map[string]any
			if err := json.Unmarshal(body, &decoded); err != nil {
				t.Fatalf("decode export params: %v", err)
			}
			if int(decoded["repoId"].(float64)) != 44 {
				t.Fatalf("unexpected export payload: %#v", decoded)
			}
			result := out.(*sharedtypes.ExportReportResult)
			result.Kind = sharedtypes.ExportReportKindRepo
			result.Format = sharedtypes.ExportFormatMarkdown
			result.OutputMode = sharedtypes.ExportOutputModeFile
			path := "/tmp/repo-report.md"
			result.FilePath = &path
			return nil
		default:
			t.Fatalf("unexpected method: %s", method)
			return nil
		}
	}

	output := captureStdout(t, func() {
		if err := runExport([]string{"repo", "--json"}); err != nil {
			t.Fatalf("runExport repo: %v", err)
		}
	})

	if len(gotMethods) != 2 || gotMethods[0] != protocol.MethodContextGet || gotMethods[1] != protocol.MethodExportRepo {
		t.Fatalf("unexpected method order: %+v", gotMethods)
	}
	if !strings.Contains(output, "\"filePath\": \"/tmp/repo-report.md\"") {
		t.Fatalf("unexpected output: %s", output)
	}
}

func stubCLIEnv() func() {
	origCallKernel := callKernelFn
	origEnsureKernel := ensureKernelFn
	origRuntimeBaseDir := runtimeBaseDir
	origReadFile := readFileFn
	origKernelBinary := kernelBinaryFn
	origTUIBinary := tuiBinaryFn
	origExecutable := osExecutableFn
	origLookPath := execLookPathFn
	origStat := osStatFn
	origGetwd := osGetwdFn
	origStartKernel := startKernelFn
	origLaunchTUI := launchTUIFn
	origSleep := timeSleepFn

	return func() {
		callKernelFn = origCallKernel
		ensureKernelFn = origEnsureKernel
		runtimeBaseDir = origRuntimeBaseDir
		readFileFn = origReadFile
		kernelBinaryFn = origKernelBinary
		tuiBinaryFn = origTUIBinary
		osExecutableFn = origExecutable
		execLookPathFn = origLookPath
		osStatFn = origStat
		osGetwdFn = origGetwd
		startKernelFn = origStartKernel
		launchTUIFn = origLaunchTUI
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
