package app

import (
	"path/filepath"
	"runtime"
	"strings"

	"crona/shared/config"
)

func nonStandardRuntimeReason(path, expectedBinaryName string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return "You seem to be running the app from a non-standard location. Please update manually."
	}
	if filepath.Base(path) != expectedBinaryName {
		return "You seem to be running the app from a non-standard location. Please update manually."
	}
	if !sameInstallDir(filepath.Dir(path), installDir()) {
		return "You seem to be running the app from a non-standard location. Please update manually."
	}
	return ""
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
