package runtime

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"crona/shared/localipc"
	"crona/shared/protocol"
	sharedtypes "crona/shared/types"
)

type LaunchCandidate struct {
	Name string
	Cmd  string
	Args []string
	Dir  string
}

type Deps struct {
	RuntimeBaseDir func() (string, error)
	ReadFile       func(string) ([]byte, error)
	KernelBinary   func() string
	TUIBinary      func() string
	OSExecutable   func() (string, error)
	ExecLookPath   func(string) (string, error)
	OSStat         func(string) (os.FileInfo, error)
	OSGetwd        func() (string, error)
	StartKernel    func(LaunchCandidate) error
	TimeSleep      func(time.Duration)
}

func (d Deps) withDefaults() Deps {
	if d.ReadFile == nil {
		d.ReadFile = os.ReadFile
	}
	if d.OSExecutable == nil {
		d.OSExecutable = os.Executable
	}
	if d.ExecLookPath == nil {
		d.ExecLookPath = exec.LookPath
	}
	if d.OSStat == nil {
		d.OSStat = os.Stat
	}
	if d.OSGetwd == nil {
		d.OSGetwd = os.Getwd
	}
	if d.StartKernel == nil {
		d.StartKernel = StartKernelProcess
	}
	if d.TimeSleep == nil {
		d.TimeSleep = time.Sleep
	}
	return d
}

func CallKernel(deps Deps, method string, params, out any) error {
	info, err := ReadKernelInfo(deps)
	if err != nil {
		return err
	}
	conn, err := localipc.Dial(KernelEndpoint(info), 5*time.Second)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()
	_ = conn.SetDeadline(time.Now().Add(10 * time.Second))

	var rawParams json.RawMessage
	if params != nil {
		body, err := json.Marshal(params)
		if err != nil {
			return err
		}
		rawParams = body
	}
	body, err := json.Marshal(protocol.Request{
		ID:     "crona-cli",
		Method: method,
		Params: rawParams,
	})
	if err != nil {
		return err
	}
	if _, err := conn.Write(append(body, '\n')); err != nil {
		return err
	}
	var resp protocol.Response
	if err := json.NewDecoder(bufio.NewReader(conn)).Decode(&resp); err != nil {
		return err
	}
	if resp.Error != nil {
		return fmt.Errorf("%s: %s", resp.Error.Code, resp.Error.Message)
	}
	if out == nil || len(resp.Result) == 0 {
		return nil
	}
	return json.Unmarshal(resp.Result, out)
}

func EnsureKernel(deps Deps) (*sharedtypes.KernelInfo, error) {
	deps = deps.withDefaults()
	if info, err := ReadKernelInfo(deps); err == nil && IsHealthy(info) {
		return info, nil
	}
	if err := LaunchKernel(deps); err != nil {
		return nil, fmt.Errorf("launch kernel: %w", err)
	}
	for i := 0; i < 20; i++ {
		deps.TimeSleep(250 * time.Millisecond)
		if info, err := ReadKernelInfo(deps); err == nil && IsHealthy(info) {
			return info, nil
		}
	}
	return nil, fmt.Errorf("kernel failed to start within 5s")
}

func ReadKernelInfo(deps Deps) (*sharedtypes.KernelInfo, error) {
	deps = deps.withDefaults()
	base, err := deps.RuntimeBaseDir()
	if err != nil {
		return nil, err
	}
	body, err := deps.ReadFile(filepath.Join(base, "kernel.json"))
	if err != nil {
		return nil, err
	}
	var info sharedtypes.KernelInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, err
	}
	NormalizeKernelInfo(&info)
	return &info, nil
}

func IsHealthy(info *sharedtypes.KernelInfo) bool {
	if info == nil || strings.TrimSpace(KernelEndpoint(info)) == "" {
		return false
	}
	conn, err := localipc.Dial(KernelEndpoint(info), 2*time.Second)
	if err != nil {
		return false
	}
	defer func() { _ = conn.Close() }()
	_ = conn.SetDeadline(time.Now().Add(2 * time.Second))
	body, err := json.Marshal(protocol.Request{ID: "healthcheck", Method: protocol.MethodHealthGet})
	if err != nil {
		return false
	}
	if _, err := conn.Write(append(body, '\n')); err != nil {
		return false
	}
	var resp protocol.Response
	if err := json.NewDecoder(bufio.NewReader(conn)).Decode(&resp); err != nil {
		return false
	}
	return resp.Error == nil
}

func NormalizeKernelInfo(info *sharedtypes.KernelInfo) {
	if info == nil {
		return
	}
	if strings.TrimSpace(info.Transport) == "" {
		info.Transport = localipc.DefaultTransport()
	}
	if strings.TrimSpace(info.Endpoint) == "" {
		info.Endpoint = strings.TrimSpace(info.SocketPath)
	}
	if strings.TrimSpace(info.SocketPath) == "" && info.Transport == localipc.TransportUnixSocket {
		info.SocketPath = info.Endpoint
	}
}

func KernelEndpoint(info *sharedtypes.KernelInfo) string {
	if info == nil {
		return ""
	}
	if strings.TrimSpace(info.Endpoint) != "" {
		return info.Endpoint
	}
	return strings.TrimSpace(info.SocketPath)
}

func KernelTransport(info *sharedtypes.KernelInfo) string {
	if info == nil || strings.TrimSpace(info.Transport) == "" {
		return localipc.DefaultTransport()
	}
	return info.Transport
}

func LaunchKernel(deps Deps) error {
	candidates := KernelLaunchCandidates(deps)
	if len(candidates) == 0 {
		return errors.New("no kernel launcher candidates found")
	}
	failures := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		if err := deps.withDefaults().StartKernel(candidate); err == nil {
			return nil
		} else {
			failures = append(failures, fmt.Sprintf("%s: %v", candidate.Name, err))
		}
	}
	return errors.New(strings.Join(failures, "; "))
}

func KernelLaunchCandidates(deps Deps) []LaunchCandidate {
	deps = deps.withDefaults()
	candidates := make([]LaunchCandidate, 0, 3)
	seen := make(map[string]struct{})
	add := func(candidate LaunchCandidate) {
		key := candidate.Cmd + "\x00" + strings.Join(candidate.Args, "\x00") + "\x00" + candidate.Dir
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		candidates = append(candidates, candidate)
	}
	if exe, err := deps.OSExecutable(); err == nil {
		kernelName := deps.KernelBinary()
		sibling := filepath.Join(filepath.Dir(exe), kernelName)
		if info, err := deps.OSStat(sibling); err == nil && !info.IsDir() {
			add(LaunchCandidate{Name: "sibling " + kernelName, Cmd: sibling})
		}
	}
	if pathCmd, err := deps.ExecLookPath(deps.KernelBinary()); err == nil {
		add(LaunchCandidate{Name: "PATH " + deps.KernelBinary(), Cmd: pathCmd})
	}
	if repoRoot, err := FindRepoRoot(deps); err == nil {
		repoBin := filepath.Join(repoRoot, "bin", deps.KernelBinary())
		if info, err := deps.OSStat(repoBin); err == nil && !info.IsDir() {
			add(LaunchCandidate{Name: "repo bin " + deps.KernelBinary(), Cmd: repoBin})
		}
		if _, err := deps.OSStat(filepath.Join(repoRoot, "kernel", "cmd", "crona-kernel")); err == nil {
			if goCmd, lookErr := deps.ExecLookPath("go"); lookErr == nil {
				add(LaunchCandidate{Name: "repo-local go run", Cmd: goCmd, Args: []string{"run", "./kernel/cmd/crona-kernel"}, Dir: repoRoot})
			}
		}
	}
	return candidates
}

func LaunchTUI(deps Deps) error {
	candidates := TUILaunchCandidates(deps)
	if len(candidates) == 0 {
		return errors.New("no TUI launcher candidates found")
	}
	failures := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		if err := RunForeground(candidate); err == nil {
			return nil
		} else {
			failures = append(failures, fmt.Sprintf("%s: %v", candidate.Name, err))
		}
	}
	return errors.New(strings.Join(failures, "; "))
}

func TUILaunchCandidates(deps Deps) []LaunchCandidate {
	deps = deps.withDefaults()
	candidates := make([]LaunchCandidate, 0, 3)
	seen := make(map[string]struct{})
	add := func(candidate LaunchCandidate) {
		key := candidate.Cmd + "\x00" + strings.Join(candidate.Args, "\x00") + "\x00" + candidate.Dir
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		candidates = append(candidates, candidate)
	}
	if exe, err := deps.OSExecutable(); err == nil {
		tuiName := deps.TUIBinary()
		sibling := filepath.Join(filepath.Dir(exe), tuiName)
		if info, err := deps.OSStat(sibling); err == nil && !info.IsDir() {
			add(LaunchCandidate{Name: "sibling " + tuiName, Cmd: sibling})
		}
	}
	if pathCmd, err := deps.ExecLookPath(deps.TUIBinary()); err == nil {
		add(LaunchCandidate{Name: "PATH " + deps.TUIBinary(), Cmd: pathCmd})
	}
	if repoRoot, err := FindRepoRoot(deps); err == nil {
		repoBin := filepath.Join(repoRoot, "bin", deps.TUIBinary())
		if info, err := deps.OSStat(repoBin); err == nil && !info.IsDir() {
			add(LaunchCandidate{Name: "repo bin " + deps.TUIBinary(), Cmd: repoBin})
		}
		if _, err := deps.OSStat(filepath.Join(repoRoot, "tui")); err == nil {
			if goCmd, lookErr := deps.ExecLookPath("go"); lookErr == nil {
				add(LaunchCandidate{Name: "repo-local go run", Cmd: goCmd, Args: []string{"run", "./tui"}, Dir: repoRoot})
			}
		}
	}
	return candidates
}

func FindRepoRoot(deps Deps) (string, error) {
	deps = deps.withDefaults()
	starts := make([]string, 0, 2)
	if wd, err := deps.OSGetwd(); err == nil {
		starts = append(starts, wd)
	}
	if exe, err := deps.OSExecutable(); err == nil {
		starts = append(starts, filepath.Dir(exe))
	}
	seen := make(map[string]struct{})
	for _, start := range starts {
		dir := start
		for {
			if _, ok := seen[dir]; ok {
				break
			}
			seen[dir] = struct{}{}
			if FileExists(deps, filepath.Join(dir, "go.work")) && FileExists(deps, filepath.Join(dir, "kernel", "cmd", "crona-kernel")) {
				return dir, nil
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	return "", errors.New("repo root not found")
}

func FileExists(deps Deps, path string) bool {
	info, err := deps.withDefaults().OSStat(path)
	return err == nil && !info.IsDir()
}

func StartKernelProcess(candidate LaunchCandidate) error {
	cmd := exec.Command(candidate.Cmd, candidate.Args...)
	cmd.Dir = candidate.Dir
	cmd.Stdin = nil
	cmd.Stdout = nil
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		return err
	}
	waitCh := make(chan error, 1)
	go func() { waitCh <- cmd.Wait() }()
	select {
	case err := <-waitCh:
		detail := strings.TrimSpace(stderr.String())
		if detail == "" && err != nil {
			detail = err.Error()
		}
		if detail == "" {
			detail = "exited immediately"
		}
		return errors.New(detail)
	case <-time.After(300 * time.Millisecond):
		return nil
	}
}

func RunForeground(candidate LaunchCandidate) error {
	cmd := exec.Command(candidate.Cmd, candidate.Args...)
	cmd.Dir = candidate.Dir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
