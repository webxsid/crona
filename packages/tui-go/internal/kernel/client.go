package kernel

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"crona/tui/internal/logger"
)

type Info struct {
	PID        int    `json:"pid"`
	Port       int    `json:"port"`
	Token      string `json:"token"`
	ScratchDir string `json:"scratchDir"`
	BaseURL    string
}

func Ensure() (*Info, error) {
	home, _ := os.UserHomeDir()
	infoPath := filepath.Join(home, ".crona", "kernel.json")

	if info, err := readInfo(infoPath); err == nil {
		if isHealthy(info) {
			logger.Infof("Kernel already running at %s (pid %d)", info.BaseURL, info.PID)
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
				logger.Infof("Kernel ready at %s", info.BaseURL)
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
	var raw struct {
		PID        int    `json:"pid"`
		Port       int    `json:"port"`
		Token      string `json:"token"`
		ScratchDir string `json:"scratchDir"`
	}
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, err
	}
	return &Info{
		PID:        raw.PID,
		Port:       raw.Port,
		Token:      raw.Token,
		ScratchDir: raw.ScratchDir,
		BaseURL:    fmt.Sprintf("http://127.0.0.1:%d", raw.Port),
	}, nil
}

func isHealthy(info *Info) bool {
	if info.PID > 0 {
		if proc, err := os.FindProcess(info.PID); err != nil || proc == nil {
			return false
		}
	}
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(fmt.Sprintf("%s/health", info.BaseURL))
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == 200
}

func launch() error {
	cmd := exec.Command("crona-kernel")
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil
	return cmd.Start()
}
