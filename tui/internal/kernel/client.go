package kernel

import (
	"bufio"
	"bytes"
	"crona/shared/config"
	"crona/shared/localipc"
	"crona/shared/protocol"
	sharedtypes "crona/shared/types"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"crona/tui/internal/logger"
)

type Info = sharedtypes.KernelInfo

func Ensure() (*Info, error) {
	base, err := config.RuntimeBaseDir()
	if err != nil {
		return nil, err
	}
	infoPath := filepath.Join(base, "kernel.json")

	if info, err := readInfo(infoPath); err == nil {
		if isHealthy(info) {
			logger.Infof("Kernel already running at %s (pid %d)", EndpointLabel(info), info.PID)
			return info, nil
		}
	}

	logger.Info("Spawning kernel...")
	if err := launch(); err != nil {
		return nil, fmt.Errorf("launch kernel: %w", err)
	}

	for i := 0; i < 20; i++ {
		time.Sleep(250 * time.Millisecond)
		if info, err := readInfo(infoPath); err == nil {
			if isHealthy(info) {
				logger.Infof("Kernel ready at %s", EndpointLabel(info))
				return info, nil
			}
		}
	}

	return nil, fmt.Errorf("kernel failed to start within 5s")
}

func readInfo(path string) (*Info, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var info sharedtypes.KernelInfo
	if err := json.Unmarshal(b, &info); err != nil {
		return nil, err
	}
	normalizeKernelInfo(&info)
	return &info, nil
}

func isHealthy(info *Info) bool {
	if info.PID > 0 {
		if proc, err := os.FindProcess(info.PID); err != nil || proc == nil {
			return false
		}
	}
	conn, err := localipc.Dial(Endpoint(info), 2*time.Second)
	if err != nil {
		return false
	}
	defer func() { _ = conn.Close() }()
	_ = conn.SetDeadline(time.Now().Add(2 * time.Second))

	reqBody, err := json.Marshal(protocol.Request{
		ID:     "healthcheck",
		Method: protocol.MethodHealthGet,
	})
	if err != nil {
		return false
	}
	if _, err := conn.Write(append(reqBody, '\n')); err != nil {
		return false
	}

	var resp protocol.Response
	if err := json.NewDecoder(bufio.NewReader(conn)).Decode(&resp); err != nil {
		return false
	}
	if resp.Error != nil {
		return false
	}
	var health sharedtypes.Health
	if err := json.Unmarshal(resp.Result, &health); err != nil {
		return false
	}
	return health.DB && health.Status == "ok"
}

func normalizeKernelInfo(info *sharedtypes.KernelInfo) {
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

func Endpoint(info *Info) string {
	if info == nil {
		return ""
	}
	if strings.TrimSpace(info.Endpoint) != "" {
		return info.Endpoint
	}
	return strings.TrimSpace(info.SocketPath)
}

func EndpointLabel(info *Info) string {
	return localipc.Label(info.Transport, Endpoint(info))
}

type launchCandidate struct {
	name string
	cmd  string
	args []string
	dir  string
}

func launch() error {
	candidates := kernelLaunchCandidates()
	if len(candidates) == 0 {
		return errors.New("no kernel launcher candidates found")
	}

	failures := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		logger.Infof("Trying kernel launcher: %s", candidate.name)
		if err := startKernel(candidate); err == nil {
			return nil
		} else {
			failures = append(failures, fmt.Sprintf("%s: %v", candidate.name, err))
		}
	}

	return errors.New(strings.Join(failures, "; "))
}

func kernelLaunchCandidates() []launchCandidate {
	candidates := make([]launchCandidate, 0, 3)
	seen := make(map[string]struct{})

	add := func(candidate launchCandidate) {
		key := candidate.cmd + "\x00" + strings.Join(candidate.args, "\x00") + "\x00" + candidate.dir
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		candidates = append(candidates, candidate)
	}

	if exe, err := os.Executable(); err == nil {
		kernelName := config.KernelBinaryName()
		sibling := filepath.Join(filepath.Dir(exe), kernelName)
		if info, err := os.Stat(sibling); err == nil && !info.IsDir() {
			add(launchCandidate{
				name: "sibling " + kernelName,
				cmd:  sibling,
			})
		}
	}

	if pathCmd, err := exec.LookPath(config.KernelBinaryName()); err == nil {
		add(launchCandidate{
			name: "PATH " + config.KernelBinaryName(),
			cmd:  pathCmd,
		})
	}

	if repoRoot, err := findRepoRoot(); err == nil {
		repoBin := filepath.Join(repoRoot, "bin", config.KernelBinaryName())
		if info, err := os.Stat(repoBin); err == nil && !info.IsDir() {
			add(launchCandidate{
				name: "repo bin " + config.KernelBinaryName(),
				cmd:  repoBin,
			})
		}
		if _, err := os.Stat(filepath.Join(repoRoot, "kernel", "cmd", "crona-kernel")); err == nil {
			if goCmd, lookErr := exec.LookPath("go"); lookErr == nil {
				add(launchCandidate{
					name: "repo-local go run",
					cmd:  goCmd,
					args: []string{"run", "./kernel/cmd/crona-kernel"},
					dir:  repoRoot,
				})
			}
		}
	}

	return candidates
}

func findRepoRoot() (string, error) {
	starts := make([]string, 0, 2)
	if wd, err := os.Getwd(); err == nil {
		starts = append(starts, wd)
	}
	if exe, err := os.Executable(); err == nil {
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

			if fileExists(filepath.Join(dir, "go.work")) && fileExists(filepath.Join(dir, "kernel", "cmd", "crona-kernel")) {
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

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func startKernel(candidate launchCandidate) error {
	cmd := exec.Command(candidate.cmd, candidate.args...)
	cmd.Dir = candidate.dir
	cmd.Stdin = nil
	cmd.Stdout = nil

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return err
	}

	waitCh := make(chan error, 1)
	go func() {
		waitCh <- cmd.Wait()
	}()

	select {
	case err := <-waitCh:
		detail := strings.TrimSpace(stderr.String())
		if err == nil {
			if detail != "" {
				return fmt.Errorf("exited immediately: %s", detail)
			}
			return errors.New("exited immediately")
		}
		if detail != "" {
			return fmt.Errorf("%w: %s", err, detail)
		}
		return err
	case <-time.After(300 * time.Millisecond):
		return nil
	}
}
