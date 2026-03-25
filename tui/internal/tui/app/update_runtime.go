package app

import (
	"os"
	"path/filepath"
	"strings"
)

func nonStandardRuntimeReason(path, expectedBinaryName string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return "You seem to be running the app from a non-standard location. Please update manually."
	}
	if filepath.Base(path) != expectedBinaryName {
		return "You seem to be running the app from a non-standard location. Please update manually."
	}
	if filepath.Dir(path) != installDir() {
		return "You seem to be running the app from a non-standard location. Please update manually."
	}
	return ""
}

func installDir() string {
	if dir := strings.TrimSpace(os.Getenv("CRONA_INSTALL_DIR")); dir != "" {
		return dir
	}
	home, err := os.UserHomeDir()
	if err != nil || strings.TrimSpace(home) == "" {
		return "."
	}
	return filepath.Join(home, ".local", "bin")
}
