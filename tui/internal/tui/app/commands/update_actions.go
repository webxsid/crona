package commands

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"crona/shared/config"
	"crona/tui/internal/api"
	"crona/tui/internal/kernel"
	"crona/tui/internal/logger"

	tea "github.com/charmbracelet/bubbletea"
)

func CheckUpdateNow(c *api.Client) tea.Cmd {
	return func() tea.Msg {
		status, err := c.CheckUpdateNow()
		if err != nil {
			logger.Errorf("CheckUpdateNow: %v", err)
			return ErrMsg{Err: err}
		}
		return UpdateStatusLoadedMsg{Status: status}
	}
}

func OpenExternalURL(rawURL string) tea.Cmd {
	return func() tea.Msg {
		target := strings.TrimSpace(rawURL)
		if target == "" {
			return ErrMsg{Err: fmt.Errorf("release URL is unavailable")}
		}
		cmd, err := externalOpenCommand(target)
		if err != nil {
			return ErrMsg{Err: err}
		}
		return tea.ExecProcess(cmd, func(err error) tea.Msg {
			if err != nil {
				return ErrMsg{Err: err}
			}
			return nil
		})()
	}
}

func InstallUpdate(status *api.UpdateStatus, currentExecutablePath string, supported bool, unsupportedReason string) tea.Cmd {
	return func() tea.Msg {
		if status == nil {
			return UpdateInstallFinishedMsg{Err: fmt.Errorf("update status is unavailable")}
		}
		if !supported {
			reason := strings.TrimSpace(unsupportedReason)
			if reason == "" {
				reason = "You seem to be running the app from a non-standard location. Please update manually."
			}
			return UpdateInstallFinishedMsg{Err: fmt.Errorf("%s", reason)}
		}
		if !status.InstallAvailable {
			reason := strings.TrimSpace(status.InstallUnavailableReason)
			if reason == "" {
				reason = "release is missing required installer assets"
			}
			return UpdateInstallFinishedMsg{Err: fmt.Errorf("install unavailable: %s", reason)}
		}

		installURL := strings.TrimSpace(status.InstallScriptURL)
		checksumsURL := strings.TrimSpace(status.ChecksumsURL)
		if installURL == "" || checksumsURL == "" {
			return UpdateInstallFinishedMsg{Err: fmt.Errorf("install metadata is incomplete")}
		}

		tmpDir, err := os.MkdirTemp("", "crona-update-*")
		if err != nil {
			return UpdateInstallFinishedMsg{Err: err}
		}
		defer os.RemoveAll(tmpDir)

		checksumsPath := filepath.Join(tmpDir, "checksums.txt")
		output, err := downloadFile(checksumsURL, checksumsPath)
		if err != nil {
			return UpdateInstallFinishedMsg{Output: output, Err: err}
		}
		checksumsBody, err := os.ReadFile(checksumsPath)
		if err != nil {
			return UpdateInstallFinishedMsg{Output: output, Err: err}
		}

		installerAssetName := config.InstallerAssetName()
		scriptPath := filepath.Join(tmpDir, installerAssetName)
		scriptOutput, err := downloadAndVerifyAsset(installURL, scriptPath, installerAssetName, checksumsBody)
		output += scriptOutput
		if err != nil {
			return UpdateInstallFinishedMsg{Output: output, Err: err}
		}
		if runtime.GOOS != "windows" {
			if err := os.Chmod(scriptPath, 0o755); err != nil {
				return UpdateInstallFinishedMsg{Output: output, Err: err}
			}
		}

		installCmd, err := updateInstallCommand(scriptPath)
		if err != nil {
			return UpdateInstallFinishedMsg{Output: output, Err: err}
		}
		installCmd.Env = append(os.Environ(), "CRONA_INSTALL_FORCE=1")
		installOutput, err := installCmd.CombinedOutput()
		output += string(installOutput)
		if err != nil {
			return UpdateInstallFinishedMsg{Output: output, Err: fmt.Errorf("install failed: %w", err)}
		}

		relaunchPath, err := resolveRelaunchPath(currentExecutablePath)
		if err != nil {
			return UpdateInstallFinishedMsg{Output: output, Err: fmt.Errorf("install succeeded but %w", err)}
		}
		relaunchCmd := exec.Command(relaunchPath)
		relaunchCmd.Stdin = nil
		relaunchCmd.Stdout = nil
		relaunchCmd.Stderr = nil
		if err := relaunchCmd.Start(); err != nil {
			return UpdateInstallFinishedMsg{
				Output: output,
				Err:    fmt.Errorf("install succeeded but relaunch failed: %w. Run %s manually", err, relaunchPath),
			}
		}
		return UpdateInstallFinishedMsg{Output: output + fmt.Sprintf("Relaunched %s\n", relaunchPath)}
	}
}

func downloadAndVerifyAsset(rawURL, path, assetName string, checksums []byte) (string, error) {
	output, err := downloadFile(rawURL, path)
	if err != nil {
		return output, err
	}
	if err := verifyAssetChecksum(path, assetName, checksums); err != nil {
		return output, err
	}
	return output + fmt.Sprintf("Verified checksum for %s\n", assetName), nil
}

func downloadFile(rawURL, path string) (string, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(rawURL)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if len(body) == 0 {
		return "", fmt.Errorf("downloaded file from %s is empty", rawURL)
	}
	if err := os.WriteFile(path, body, 0o700); err != nil {
		return "", err
	}
	return fmt.Sprintf("Downloaded %s\n", rawURL), nil
}

func verifyAssetChecksum(path, assetName string, checksums []byte) error {
	expected := expectedChecksum(assetName, checksums)
	if expected == "" {
		return fmt.Errorf("checksums.txt is missing %s", assetName)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	sum := sha256.Sum256(body)
	actual := hex.EncodeToString(sum[:])
	if !strings.EqualFold(actual, expected) {
		return fmt.Errorf("checksum mismatch for %s", assetName)
	}
	return nil
}

func expectedChecksum(assetName string, checksums []byte) string {
	for _, line := range strings.Split(string(checksums), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		if parts[len(parts)-1] == assetName {
			return parts[0]
		}
	}
	return ""
}

func resolveRelaunchPath(currentExecutablePath string) (string, error) {
	binaryName := config.TUIBinaryName()
	candidates := []string{}
	if candidate := normalizedInstalledExecutable(currentExecutablePath, binaryName); candidate != "" {
		candidates = append(candidates, candidate)
	}
	if state, err := kernel.ReadTUIRuntimeState(); err == nil && state != nil {
		if candidate := normalizedInstalledExecutable(state.ExecutablePath, binaryName); candidate != "" {
			candidates = append(candidates, candidate)
		}
	}
	if candidate := filepath.Join(installDir(), binaryName); strings.TrimSpace(candidate) != "" {
		candidates = append(candidates, candidate)
	}

	seen := make(map[string]struct{}, len(candidates))
	for _, candidate := range candidates {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" {
			continue
		}
		if _, ok := seen[candidate]; ok {
			continue
		}
		seen[candidate] = struct{}{}
		if isExecutableFile(candidate) {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("could not find a runnable %s binary to relaunch; run %s manually after restart", binaryName, binaryName)
}

func normalizedInstalledExecutable(path, binaryName string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	if filepath.Base(path) != binaryName {
		return ""
	}
	if !sameInstallDir(filepath.Dir(path), installDir()) {
		return ""
	}
	return path
}

func isExecutableFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return false
	}
	if runtime.GOOS == "windows" {
		return strings.EqualFold(filepath.Ext(path), ".exe")
	}
	return info.Mode().Perm()&0o111 != 0
}

func updateInstallCommand(installerPath string) (*exec.Cmd, error) {
	switch runtime.GOOS {
	case "windows":
		powershellPath, err := exec.LookPath("powershell.exe")
		if err != nil {
			return nil, err
		}
		return exec.Command(powershellPath, "-NoProfile", "-ExecutionPolicy", "Bypass", "-File", installerPath), nil
	default:
		shellPath, err := exec.LookPath("sh")
		if err != nil {
			return nil, err
		}
		return exec.Command(shellPath, installerPath), nil
	}
}

func externalOpenCommand(target string) (*exec.Cmd, error) {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", target), nil
	case "linux":
		return exec.Command("xdg-open", target), nil
	case "windows":
		return exec.Command("cmd", "/c", "start", "", target), nil
	default:
		return nil, os.ErrInvalid
	}
}

func installDir() string {
	dir, err := config.InstallDir()
	if err != nil {
		return "."
	}
	return dir
}

func sameInstallDir(left, right string) bool {
	left = filepath.Clean(strings.TrimSpace(left))
	right = filepath.Clean(strings.TrimSpace(right))
	if runtime.GOOS == "windows" {
		return strings.EqualFold(left, right)
	}
	return left == right
}
