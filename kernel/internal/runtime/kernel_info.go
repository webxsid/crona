package runtime

import (
	"encoding/json"
	"os"

	sharedtypes "crona.work/shared/types"
)

func WriteKernelInfo(paths Paths, info sharedtypes.KernelInfo) error {
	tmpPath := paths.InfoFile + ".tmp"

	body, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(tmpPath, body, FilePerm()); err != nil {
		return err
	}

	return os.Rename(tmpPath, paths.InfoFile)
}

func ClearKernelInfo(paths Paths) error {
	err := os.Remove(paths.InfoFile)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func ReadKernelInfo(paths Paths) (*sharedtypes.KernelInfo, error) {
	body, err := os.ReadFile(paths.InfoFile)
	if err != nil {
		return nil, err
	}

	var info sharedtypes.KernelInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, err
	}

	return &info, nil
}
