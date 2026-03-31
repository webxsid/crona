package export

import (
	"errors"
	"os"
	"path/filepath"

	"crona/kernel/internal/runtime"
)

func userTemplatePath(paths runtime.Paths, descriptor assetDescriptor) string {
	return filepath.Join(paths.UserAssetsDir, "export", descriptor.userRelativePath)
}

func legacyUserTemplatePath(paths runtime.Paths, descriptor assetDescriptor) string {
	if descriptor.legacyUserPath == "" {
		return ""
	}
	return filepath.Join(paths.UserAssetsDir, "export", descriptor.legacyUserPath)
}

func templateMetaPath(paths runtime.Paths, descriptor assetDescriptor) string {
	return userTemplatePath(paths, descriptor) + ".meta.json"
}

func legacyTemplateMetaPath(paths runtime.Paths, descriptor assetDescriptor) string {
	legacyPath := legacyUserTemplatePath(paths, descriptor)
	if legacyPath == "" {
		return ""
	}
	return legacyPath + ".meta.json"
}

func defaultAssetSource(paths runtime.Paths, descriptor assetDescriptor) ([]byte, string, error) {
	if presetID := selectedPresetIDForAsset(paths, descriptor.reportKind, descriptor.assetKind); presetID != "" {
		if body, ok := presetTemplateBody(descriptor.reportKind, descriptor.assetKind, presetID); ok {
			return []byte(body), "preset:" + presetID, nil
		}
	}
	candidates := []string{
		filepath.Join(paths.BundledAssetsDir, "export", descriptor.bundledPath),
		filepath.Join("assets", "export", descriptor.bundledPath),
		filepath.Join("..", "assets", "export", descriptor.bundledPath),
		filepath.Join("..", "..", "assets", "export", descriptor.bundledPath),
	}
	if descriptor.legacyBundledPath != "" {
		candidates = append(candidates,
			filepath.Join(paths.BundledAssetsDir, "export", descriptor.legacyBundledPath),
			filepath.Join("assets", "export", descriptor.legacyBundledPath),
			filepath.Join("..", "assets", "export", descriptor.legacyBundledPath),
			filepath.Join("..", "..", "assets", "export", descriptor.legacyBundledPath),
		)
	}
	for _, candidate := range candidates {
		body, err := os.ReadFile(candidate)
		if err == nil {
			return body, candidate, nil
		}
	}
	return []byte(descriptor.fallback), filepath.Join(paths.BundledAssetsDir, "export", descriptor.bundledPath), nil
}

func readLegacyUserAsset(paths runtime.Paths, descriptor assetDescriptor) ([]byte, templateMeta, bool, error) {
	legacyPath := legacyUserTemplatePath(paths, descriptor)
	if legacyPath == "" {
		return nil, templateMeta{}, false, nil
	}
	body, err := os.ReadFile(legacyPath)
	if errors.Is(err, os.ErrNotExist) {
		return nil, templateMeta{}, false, nil
	}
	if err != nil {
		return nil, templateMeta{}, false, err
	}
	meta, _ := readTemplateMeta(legacyTemplateMetaPath(paths, descriptor))
	return body, meta, true, nil
}
