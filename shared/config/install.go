package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const EnvVarInstallDir = "CRONA_INSTALL_DIR"

func InstallDir() (string, error) {
	if dir := InstallDirOverride(); dir != "" {
		return dir, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return DefaultInstallDirForOS(runtime.GOOS, home, os.Getenv("LocalAppData")), nil
}

func InstallDirOverride() string {
	return strings.TrimSpace(os.Getenv(EnvVarInstallDir))
}

func DefaultInstallDirForOS(goos, home, localAppData string) string {
	if goos == "windows" {
		localAppData = strings.TrimSpace(localAppData)
		if localAppData == "" {
			home = strings.TrimSpace(home)
			if home == "" {
				return "."
			}
			localAppData = filepath.Join(home, "AppData", "Local")
		}
		return filepath.Join(localAppData, "Programs", "Crona", "bin")
	}

	home = strings.TrimSpace(home)
	if home == "" {
		return "."
	}
	return filepath.Join(home, ".local", "bin")
}

func InstallerAssetName() string {
	return InstallerAssetNameForGOOS(runtime.GOOS)
}

func InstallerAssetNameForGOOS(goos string) string {
	if goos == "windows" {
		return "install-crona-tui.ps1"
	}
	return "install-crona-tui.sh"
}
