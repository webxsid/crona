package notify

import (
	"fmt"
	"os"
	"path/filepath"

	assetbundle "crona.local/assets"
	runtimepkg "crona/kernel/internal/runtime"
	sharedtypes "crona/shared/types"
)

var alertSoundFiles = map[sharedtypes.AlertSoundPreset]string{
	sharedtypes.AlertSoundPresetChime:            "sounds/chime.mp3",
	sharedtypes.AlertSoundPresetSoftBell:         "sounds/soft-bell.mp3",
	sharedtypes.AlertSoundPresetNotificationPing: "sounds/notification-ping.mp3",
	sharedtypes.AlertSoundPresetFocusGong:        "sounds/focus-gong.mp3",
	sharedtypes.AlertSoundPresetMinimalClick:     "sounds/minimal-click.mp3",
}

func alertIconPath(paths runtimepkg.Paths) string {
	for _, candidate := range alertAssetCandidates(paths, "logo.svg") {
		if fileExists(candidate) {
			return candidate
		}
	}
	if err := assetbundle.Ensure(paths.BundledAssetsDir, filepath.Join("alerts", "logo.svg")); err == nil {
		candidate := filepath.Join(paths.BundledAssetsDir, "alerts", "logo.svg")
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
	if err := assetbundle.Ensure(paths.BundledAssetsDir, filepath.Join("alerts", name)); err == nil {
		candidate := filepath.Join(paths.BundledAssetsDir, "alerts", name)
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
