package runtime

import "os"

// ResetManagedData removes runtime-managed user data while leaving the live
// kernel transport files and installed binaries alone.
func ResetManagedData(paths Paths) error {
	for _, path := range []string{
		paths.ScratchDir,
		paths.ReportsDir,
		paths.ICSDir,
		paths.UserAssetsDir,
		paths.LogsDir,
	} {
		if path == "" {
			continue
		}
		if err := os.RemoveAll(path); err != nil {
			return err
		}
	}
	for _, path := range []string{
		paths.UpdateFile,
	} {
		if path == "" {
			continue
		}
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}
