package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	EnvVarRuntimeDir   = "CRONA_HOME"
	baseDirProdDarwin  = "Crona"
	baseDirDevDarwin   = "Crona Dev"
	baseDirProdLinux   = "crona"
	baseDirDevLinux    = "crona-dev"
	baseDirProdWindows = "Crona"
	baseDirDevWindows  = "Crona Dev"
	legacyBaseDirProd  = ".crona"
	legacyBaseDirDev   = ".crona-dev"
)

func IsDevMode(mode string) bool {
	return strings.EqualFold(strings.TrimSpace(mode), ModeDev)
}

func RuntimeBaseDirNameForMode(mode string) string {
	if override := RuntimeBaseDirOverride(); override != "" {
		return filepath.Base(override)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return runtimeBaseDirName(runtime.GOOS, mode)
	}

	path, err := DefaultRuntimeBaseDirForModeOnOS(mode, runtime.GOOS, home, os.Getenv("XDG_DATA_HOME"), os.Getenv("LocalAppData"))
	if err != nil {
		return runtimeBaseDirName(runtime.GOOS, mode)
	}
	return filepath.Base(path)
}

func RuntimeBaseDirName() string {
	return RuntimeBaseDirNameForMode(Load().Mode)
}

func RuntimeBaseDir() (string, error) {
	return RuntimeBaseDirForMode(Load().Mode)
}

func RuntimeBaseDirForMode(mode string) (string, error) {
	if override := RuntimeBaseDirOverride(); override != "" {
		return override, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return DefaultRuntimeBaseDirForModeOnOS(mode, runtime.GOOS, home, os.Getenv("XDG_DATA_HOME"), os.Getenv("LocalAppData"))
}

func LegacyRuntimeBaseDirForMode(mode string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return LegacyRuntimeBaseDirForModeOnOS(mode, runtime.GOOS, home), nil
}

func RuntimeBaseDirOverride() string {
	return strings.TrimSpace(os.Getenv(EnvVarRuntimeDir))
}

func DefaultRuntimeBaseDirForModeOnOS(mode, goos, home, xdgDataHome, localAppData string) (string, error) {
	home = strings.TrimSpace(home)
	if home == "" {
		return "", os.ErrNotExist
	}

	if goos == "windows" {
		localAppData = strings.TrimSpace(localAppData)
		if localAppData == "" {
			localAppData = filepath.Join(home, "AppData", "Local")
		}
		return filepath.Join(localAppData, runtimeBaseDirName(goos, mode)), nil
	}

	switch goos {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", runtimeBaseDirName(goos, mode)), nil
	default:
		xdgDataHome = strings.TrimSpace(xdgDataHome)
		if xdgDataHome == "" {
			xdgDataHome = filepath.Join(home, ".local", "share")
		}
		return filepath.Join(xdgDataHome, runtimeBaseDirName(goos, mode)), nil
	}
}

func LegacyRuntimeBaseDirForModeOnOS(mode, goos, home string) string {
	home = strings.TrimSpace(home)
	if home == "" {
		return ""
	}

	if goos == "windows" {
		return filepath.Join(home, runtimeBaseDirName(goos, mode))
	}
	if IsDevMode(mode) {
		return filepath.Join(home, legacyBaseDirDev)
	}
	return filepath.Join(home, legacyBaseDirProd)
}

func runtimeBaseDirName(goos, mode string) string {
	switch goos {
	case "windows":
		if IsDevMode(mode) {
			return baseDirDevWindows
		}
		return baseDirProdWindows
	case "darwin":
		if IsDevMode(mode) {
			return baseDirDevDarwin
		}
		return baseDirProdDarwin
	default:
		if IsDevMode(mode) {
			return baseDirDevLinux
		}
		return baseDirProdLinux
	}
}

func BinarySuffixForMode(mode string) string {
	if IsDevMode(mode) {
		return "-dev"
	}
	return ""
}

func CLIBinaryNameForMode(mode string) string {
	return binaryName("crona" + BinarySuffixForMode(mode))
}

func KernelBinaryNameForMode(mode string) string {
	return binaryName("crona-kernel" + BinarySuffixForMode(mode))
}

func TUIBinaryNameForMode(mode string) string {
	return binaryName("crona-tui" + BinarySuffixForMode(mode))
}

func CLIBinaryName() string {
	return CLIBinaryNameForMode(Load().Mode)
}

func KernelBinaryName() string {
	return KernelBinaryNameForMode(Load().Mode)
}

func TUIBinaryName() string {
	return TUIBinaryNameForMode(Load().Mode)
}

func binaryName(base string) string {
	if runtime.GOOS == "windows" {
		return base + ".exe"
	}
	return base
}
