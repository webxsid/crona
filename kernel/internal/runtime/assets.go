package runtime

import assetbundle "crona.local/assets"

func EnsureBundledAssets(paths Paths) error {
	return assetbundle.EnsureAll(paths.BundledAssetsDir)
}
