package notify

import (
	"fmt"
	"os"
	"path/filepath"

	runtimepkg "crona/kernel/internal/runtime"
	sharedtypes "crona/shared/types"
)

var alertSoundFiles = map[sharedtypes.AlertSoundPreset]string{
	sharedtypes.AlertSoundPresetChime:        "sounds/chime.wav",
	sharedtypes.AlertSoundPresetSoftBell:     "sounds/soft-bell.wav",
	sharedtypes.AlertSoundPresetFocusGong:    "sounds/focus-gong.wav",
	sharedtypes.AlertSoundPresetMinimalClick: "sounds/minimal-click.wav",
}

func alertIconPath(paths runtimepkg.Paths) string {
	for _, candidate := range alertAssetCandidates(paths, "logo.svg") {
		if fileExists(candidate) {
			return candidate
		}
	}
	return ""
}

func alertSoundPath(paths runtimepkg.Paths, preset sharedtypes.AlertSoundPreset) (string, error) {
	name, ok := alertSoundFiles[sharedtypes.NormalizeAlertSoundPreset(preset)]
	if !ok {
		name = alertSoundFiles[sharedtypes.AlertSoundPresetChime]
	}
	for _, candidate := range alertAssetCandidates(paths, name) {
		if fileExists(candidate) {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("bundled alert sound asset is unavailable: %s", name)
}

func alertAssetCandidates(paths runtimepkg.Paths, relative string) []string {
	return []string{
		filepath.Join(paths.BundledAssetsDir, "alerts", relative),
		filepath.Join("assets", "alerts", relative),
		filepath.Join("..", "assets", "alerts", relative),
		filepath.Join("..", "..", "assets", "alerts", relative),
	}
}

func fileExists(path string) bool {
	if path == "" {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
