package kernel

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"crona/shared/config"
)

type TUIRuntimeState struct {
	ExecutablePath string `json:"executablePath"`
	RecordedAt     string `json:"recordedAt"`
}

func WriteTUIRuntimeState(executablePath string) error {
	executablePath = strings.TrimSpace(executablePath)
	if executablePath == "" {
		return nil
	}
	path, err := tuiRuntimeStatePath()
	if err != nil {
		return err
	}
	state := TUIRuntimeState{
		ExecutablePath: executablePath,
		RecordedAt:     time.Now().UTC().Format(time.RFC3339),
	}
	body, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, body, 0o600)
}

func ReadTUIRuntimeState() (*TUIRuntimeState, error) {
	path, err := tuiRuntimeStatePath()
	if err != nil {
		return nil, err
	}
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var state TUIRuntimeState
	if err := json.Unmarshal(body, &state); err != nil {
		return nil, err
	}
	return &state, nil
}

func tuiRuntimeStatePath() (string, error) {
	base, err := config.RuntimeBaseDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "tui.json"), nil
}
