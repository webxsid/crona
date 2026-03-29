package runtime

import (
	"path/filepath"
	goruntime "runtime"
	"strings"

	"crona/shared/config"
)

const manualUpdateMessage = "You seem to be running the app from a non-standard location. Please update manually."

func NonStandardRuntimeReason(path, expectedBinaryName string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return manualUpdateMessage
	}
	if filepath.Base(path) != expectedBinaryName {
		return manualUpdateMessage
	}
	if !SameInstallDir(filepath.Dir(path), InstallDir()) {
		return manualUpdateMessage
	}
	return ""
}

func InstallDir() string {
	dir, err := config.InstallDir()
	if err != nil {
		return "."
	}
	return dir
}

func SameInstallDir(left, right string) bool {
	left = filepath.Clean(strings.TrimSpace(left))
	right = filepath.Clean(strings.TrimSpace(right))
	if goruntime.GOOS == "windows" {
		return strings.EqualFold(left, right)
	}
	return left == right
}
