package runtime

import (
	"os"
	"path/filepath"

	"crona/shared/config"
	"crona/shared/localipc"
)

const (
	dirPerm  = 0o700
	filePerm = 0o600
)

type Paths struct {
	BaseDir          string
	DBPath           string
	ScratchDir       string
	AssetsDir        string
	BundledAssetsDir string
	UserAssetsDir    string
	ReportsDir       string
	ICSDir           string
	LogsDir          string
	InfoFile         string
	UpdateFile       string
	Transport        string
	Endpoint         string
	SocketPath       string
	CurrentLogDir    string
}

func ResolvePaths() (Paths, error) {
	base, err := config.RuntimeBaseDir()
	if err != nil {
		return Paths{}, err
	}
	logs := filepath.Join(base, "logs")
	assets := filepath.Join(base, "assets")
	transport := localipc.DefaultTransport()
	endpoint := localipc.DefaultEndpoint(base, config.Load().Mode)
	socketPath := ""
	if transport == localipc.TransportUnixSocket {
		socketPath = endpoint
	}

	return Paths{
		BaseDir:          base,
		DBPath:           filepath.Join(base, "crona.db"),
		ScratchDir:       filepath.Join(base, "scratch"),
		AssetsDir:        assets,
		BundledAssetsDir: filepath.Join(assets, "bundled"),
		UserAssetsDir:    filepath.Join(assets, "user"),
		ReportsDir:       filepath.Join(base, "reports"),
		ICSDir:           filepath.Join(base, "calendar"),
		LogsDir:          logs,
		InfoFile:         filepath.Join(base, "kernel.json"),
		UpdateFile:       filepath.Join(base, "update.json"),
		Transport:        transport,
		Endpoint:         endpoint,
		SocketPath:       socketPath,
		CurrentLogDir:    filepath.Join(logs, dateStamp()),
	}, nil
}

func EnsurePaths(paths Paths) error {
	for _, dir := range []string{
		paths.BaseDir,
		paths.ScratchDir,
		paths.AssetsDir,
		paths.BundledAssetsDir,
		paths.UserAssetsDir,
		paths.ReportsDir,
		paths.ICSDir,
		paths.LogsDir,
		paths.CurrentLogDir,
	} {
		if dir == "" {
			continue
		}
		if err := os.MkdirAll(dir, dirPerm); err != nil {
			return err
		}
	}
	return nil
}

func FilePerm() os.FileMode {
	return filePerm
}
