package runtime

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	sharedtypes "crona/shared/types"
)

type InstallSourceFile struct {
	InstallSource sharedtypes.InstallSource `json:"installSource"`
	BrewFormula   string                    `json:"brewFormula,omitempty"`
}

func LoadInstallSource(path string) (sharedtypes.InstallSource, error) {
	file, err := LoadInstallSourceFile(path)
	if err != nil {
		return sharedtypes.InstallSourceUnknown, err
	}
	return sharedtypes.NormalizeInstallSource(file.InstallSource), nil
}

func LoadInstallSourceFile(path string) (InstallSourceFile, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return InstallSourceFile{}, err
	}
	var file InstallSourceFile
	if err := json.Unmarshal(body, &file); err != nil {
		return InstallSourceFile{}, err
	}
	return file, nil
}

func WriteInstallSource(path string, source sharedtypes.InstallSource) error {
	return WriteInstallSourceDetails(path, source, "")
}

func WriteInstallSourceDetails(path string, source sharedtypes.InstallSource, brewFormula string) error {
	if strings.TrimSpace(path) == "" {
		return nil
	}
	source = sharedtypes.NormalizeInstallSource(source)
	if source == sharedtypes.InstallSourceUnknown {
		return nil
	}
	body, err := json.MarshalIndent(InstallSourceFile{
		InstallSource: sharedtypes.NormalizeInstallSource(source),
		BrewFormula:   strings.TrimSpace(brewFormula),
	}, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), dirPerm); err != nil {
		return err
	}
	return os.WriteFile(path, body, filePerm)
}
