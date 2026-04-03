package commands

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"crona/shared/config"
	"crona/tui/internal/api"
	"crona/tui/internal/logger"

	tea "github.com/charmbracelet/bubbletea"
)

var (
	downloadFileFn         = downloadFile
	downloadAndVerifyFn    = downloadAndVerifyAsset
	updateInstallCommandFn = updateInstallCommand
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

func InstallUpdate(status *api.UpdateStatus, supported bool, unsupportedReason string) tea.Cmd {
	return func() tea.Msg {
		cmd, err := prepareInstallCommand(status, supported, unsupportedReason)
		if err != nil {
			return ErrMsg{Err: err}
		}
		return UpdateInstallPreparedMsg{Cmd: cmd}
	}
}

func prepareInstallCommand(status *api.UpdateStatus, supported bool, unsupportedReason string) (*exec.Cmd, error) {
	if status == nil {
		return nil, fmt.Errorf("update status is unavailable")
	}
	if !supported {
		reason := strings.TrimSpace(unsupportedReason)
		if reason == "" {
			reason = "You seem to be running the app from a non-standard location. Please update manually."
		}
		return nil, fmt.Errorf("%s", reason)
	}
	if !status.InstallAvailable {
		reason := strings.TrimSpace(status.InstallUnavailableReason)
		if reason == "" {
			reason = "release is missing required installer assets"
		}
		return nil, fmt.Errorf("install unavailable: %s", reason)
	}

	installURL := strings.TrimSpace(status.InstallScriptURL)
	checksumsURL := strings.TrimSpace(status.ChecksumsURL)
	if installURL == "" || checksumsURL == "" {
		return nil, fmt.Errorf("install metadata is incomplete")
	}

	tmpDir, err := os.MkdirTemp("", "crona-update-*")
	if err != nil {
		return nil, err
	}

	checksumsPath := filepath.Join(tmpDir, "checksums.txt")
	if _, err := downloadFileFn(checksumsURL, checksumsPath); err != nil {
		return nil, err
	}
	checksumsBody, err := os.ReadFile(checksumsPath)
	if err != nil {
		return nil, err
	}

	installerAssetName := config.InstallerAssetName()
	scriptPath := filepath.Join(tmpDir, installerAssetName)
	if _, err := downloadAndVerifyFn(installURL, scriptPath, installerAssetName, checksumsBody); err != nil {
		return nil, err
	}
	if runtime.GOOS != "windows" {
		if err := os.Chmod(scriptPath, 0o755); err != nil {
			return nil, err
		}
	}

	installCmd, err := updateInstallCommandFn(scriptPath)
	if err != nil {
		return nil, err
	}
	installCmd.Stdin = os.Stdin
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr
	installEnv := append(os.Environ(), "CRONA_INSTALL_FORCE=1")
	if releaseBaseURL := releaseBaseURLForInstaller(installURL); releaseBaseURL != "" {
		installEnv = append(installEnv, config.EnvVarReleaseBaseURL+"="+releaseBaseURL)
	}
	localInstallDir, localRuntimeDir := localUpdateInstallRoots(status)
	if localInstallDir != "" && localRuntimeDir != "" {
		installEnv = append(installEnv,
			config.EnvVarInstallDir+"="+localInstallDir,
			config.EnvVarRuntimeDir+"="+localRuntimeDir,
		)
	}
	installCmd.Env = installEnv
	return installCmd, nil
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

func releaseBaseURLForInstaller(installURL string) string {
	parsed, err := url.Parse(strings.TrimSpace(installURL))
	if err != nil || strings.TrimSpace(parsed.Scheme) == "" || strings.TrimSpace(parsed.Host) == "" {
		return ""
	}
	parsed.RawQuery = ""
	parsed.Fragment = ""
	parsed.Path = strings.TrimSuffix(parsed.Path, "/"+config.InstallerAssetNameForGOOS(runtime.GOOS))
	return strings.TrimSuffix(parsed.String(), "/")
}

func localUpdateInstallRoots(status *api.UpdateStatus) (string, string) {
	if status == nil || !IsLocalLoopbackUpdateURL(status.InstallScriptURL) {
		return "", ""
	}
	version := normalizeInstallVersion(status)
	base := filepath.Join(os.TempDir(), "crona-local-update", version)
	return filepath.Join(base, "bin"), filepath.Join(base, "home")
}

func IsLocalLoopbackUpdateURL(rawURL string) bool {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return false
	}
	host := strings.TrimSpace(parsed.Hostname())
	switch host {
	case "127.0.0.1", "localhost":
		return true
	default:
		return false
	}
}

func normalizeInstallVersion(status *api.UpdateStatus) string {
	if status == nil {
		return "dev"
	}
	value := strings.TrimSpace(status.ReleaseTag)
	if value == "" {
		value = strings.TrimSpace(status.LatestVersion)
	}
	value = strings.TrimPrefix(value, "v")
	value = strings.ReplaceAll(value, "/", "-")
	value = strings.ReplaceAll(value, string(filepath.Separator), "-")
	if value == "" {
		return "dev"
	}
	return value
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
