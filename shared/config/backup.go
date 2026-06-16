package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func BackupBaseDir() (string, error) {
	if override := BackupBaseDirOverride(); override != "" {
		return override, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return DefaultBackupBaseDirForModeOnOS(
		Load().Mode,
		runtime.GOOS,
		home,
		os.Getenv("XDG_DATA_HOME"),
		os.Getenv("LocalAppData"),
	)
}

func BackupBaseDirForMode(mode string) (string, error) {
	if override := BackupBaseDirOverride(); override != "" {
		return override, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return DefaultBackupBaseDirForModeOnOS(
		mode,
		runtime.GOOS,
		home,
		os.Getenv("XDG_DATA_HOME"),
		os.Getenv("LocalAppData"),
	)
}

func BackupBaseDirOverride() string {
	return strings.TrimSpace(os.Getenv("CRONA_BACKUP_DIR"))
}

func DefaultBackupBaseDirForModeOnOS(
	mode, goos, home, xdgDataHome, localAppData string,
) (string, error) {
	home = strings.TrimSpace(home)
	if home == "" {
		return "", os.ErrNotExist
	}

	name := backupBaseDirName(goos, mode)
	if goos == "windows" {
		localAppData = strings.TrimSpace(localAppData)
		if localAppData == "" {
			localAppData = filepath.Join(home, "AppData", "Local")
		}
		return filepath.Join(localAppData, name), nil
	}
	if goos == "darwin" {
		return filepath.Join(home, "Library", "Application Support", name), nil
	}

	xdgDataHome = strings.TrimSpace(xdgDataHome)
	if xdgDataHome == "" {
		xdgDataHome = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(xdgDataHome, name), nil
}

func backupBaseDirName(goos, mode string) string {
	switch goos {
	case "windows":
		if IsDevMode(mode) {
			return "Crona Dev Backups"
		}
		return "Crona Backups"
	case "darwin":
		if IsDevMode(mode) {
			return "Crona Dev Backups"
		}
		return "Crona Backups"
	default:
		if IsDevMode(mode) {
			return "crona-dev-backups"
		}
		return "crona-backups"
	}
}
