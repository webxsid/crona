package runtime

import assetbundle "crona.work/assets"

func EnsureBundledAssets(paths Paths) error {
	return assetbundle.EnsureAll(paths.BundledAssetsDir)
}
