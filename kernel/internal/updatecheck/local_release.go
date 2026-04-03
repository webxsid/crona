package updatecheck

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"crona/shared/config"
)

func (s *Service) prepareLocalReleaseSource(ctx context.Context) (latestRelease, string, string, error) {
	releaseDir, err := resolveLocalReleaseDir(s.targetGOOS())
	if err != nil {
		return latestRelease{}, "", "", err
	}
	release, err := localReleaseFromDir(releaseDir, s.targetGOOS())
	if err != nil {
		return latestRelease{}, "", "", err
	}
	baseURL, stop, err := startLocalReleaseServer(ctx, releaseDir)
	if err != nil {
		return latestRelease{}, "", "", err
	}
	release.InstallURL = baseURL + "/" + config.InstallerAssetNameForGOOS(s.targetGOOS())
	release.ChecksumsURL = baseURL + "/checksums.txt"
	release.URL = baseURL + "/"

	s.mu.Lock()
	prevStop := s.stopLocalRelease
	s.stopLocalRelease = stop
	s.localRelease = &release
	s.localReleaseBase = baseURL
	s.mu.Unlock()
	if prevStop != nil {
		_ = prevStop()
	}

	return release, releaseDir, baseURL, nil
}

func resolveLocalReleaseDir(goos string) (string, error) {
	if override := strings.TrimSpace(os.Getenv(config.EnvVarDevUpdateReleaseDir)); override != "" {
		if _, err := localReleaseFromDir(override, goos); err != nil {
			return "", fmt.Errorf("local update release dir %s: %w", override, err)
		}
		return override, nil
	}

	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		candidate := filepath.Join(wd, "release")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			dir, err := latestReleaseDir(candidate, goos)
			if err != nil {
				return "", err
			}
			return dir, nil
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			break
		}
		wd = parent
	}
	return "", fmt.Errorf("could not find a local release directory; set %s or run from the repo", config.EnvVarDevUpdateReleaseDir)
}

func latestReleaseDir(root, goos string) (string, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return "", err
	}
	type candidate struct {
		version semver
		path    string
	}
	candidates := make([]candidate, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		path := filepath.Join(root, entry.Name())
		if _, err := localReleaseFromDir(path, goos); err != nil {
			continue
		}
		version, ok := parseSemver(entry.Name())
		if !ok {
			continue
		}
		candidates = append(candidates, candidate{
			version: version,
			path:    path,
		})
	}
	if len(candidates) == 0 {
		return "", fmt.Errorf("no versioned release directories found under %s", root)
	}
	sort.Slice(candidates, func(i, j int) bool {
		return compareSemver(candidates[i].version, candidates[j].version) > 0
	})
	return candidates[0].path, nil
}

func localReleaseFromDir(releaseDir, goos string) (latestRelease, error) {
	tag := strings.TrimSpace(filepath.Base(releaseDir))
	version := normalizeVersion(tag)
	if version == "" || strings.HasSuffix(tag, "-") || strings.HasSuffix(version, "-") {
		return latestRelease{}, fmt.Errorf("release dir %s does not look like a versioned release", releaseDir)
	}
	installerName := config.InstallerAssetNameForGOOS(goos)
	if _, err := os.Stat(filepath.Join(releaseDir, installerName)); err != nil {
		return latestRelease{}, fmt.Errorf("local release %s is missing %s", releaseDir, installerName)
	}
	if _, err := os.Stat(filepath.Join(releaseDir, "checksums.txt")); err != nil {
		return latestRelease{}, fmt.Errorf("local release %s is missing checksums.txt", releaseDir)
	}
	bundleName := fmt.Sprintf("crona-bundle-%s-%s-%s.zip", tag, goos, runtime.GOARCH)
	if _, err := os.Stat(filepath.Join(releaseDir, bundleName)); err != nil {
		return latestRelease{}, fmt.Errorf("local release %s is missing %s", releaseDir, bundleName)
	}
	assetsName := fmt.Sprintf("crona-assets-%s.tar.gz", tag)
	if _, err := os.Stat(filepath.Join(releaseDir, assetsName)); err != nil {
		return latestRelease{}, fmt.Errorf("local release %s is missing %s", releaseDir, assetsName)
	}
	return latestRelease{
		Version:      version,
		Tag:          tag,
		Name:         "Local release " + tag,
		Notes:        "Local dev update simulation from " + filepath.Base(releaseDir),
		InstallAsset: installerName,
		PublishedAt:  "",
		IsPrerelease: strings.Contains(version, "-"),
	}, nil
}

func startLocalReleaseServer(ctx context.Context, releaseDir string) (string, func() error, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", nil, err
	}
	server := &http.Server{Handler: http.FileServer(http.Dir(releaseDir))}
	go func() {
		<-ctx.Done()
		_ = server.Close()
	}()
	go func() {
		_ = server.Serve(listener)
	}()
	baseURL := "http://" + listener.Addr().String()
	return baseURL, server.Close, nil
}
