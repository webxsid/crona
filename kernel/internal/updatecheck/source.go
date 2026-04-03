package updatecheck

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"

	"crona/shared/config"
	sharedtypes "crona/shared/types"
	versionpkg "crona/shared/version"
)

type latestRelease struct {
	Version      string
	Tag          string
	Name         string
	Notes        string
	URL          string
	InstallURL   string
	InstallAsset string
	ChecksumsURL string
	PublishedAt  string
	IsPrerelease bool
}

func (s *Service) fetchLatestRelease(ctx context.Context, channel sharedtypes.UpdateChannel) (latestRelease, error) {
	if release, ok := s.currentLocalRelease(); ok {
		return release, nil
	}
	if sharedtypes.NormalizeUpdateChannel(channel) == sharedtypes.UpdateChannelBeta {
		return s.fetchLatestReleaseFromList(ctx)
	}
	return s.fetchLatestStableRelease(ctx)
}

func (s *Service) fetchLatestStableRelease(ctx context.Context) (latestRelease, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", versionpkg.RepoOwner, versionpkg.RepoName), nil)
	if err != nil {
		return latestRelease{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "crona/"+versionpkg.Current())

	resp, err := s.client.Do(req)
	if err != nil {
		return latestRelease{}, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return latestRelease{}, fmt.Errorf("github releases returned %s", resp.Status)
	}

	var payload githubReleasePayload
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return latestRelease{}, err
	}
	return s.releaseFromPayload(payload)
}

func (s *Service) fetchLatestReleaseFromList(ctx context.Context) (latestRelease, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", versionpkg.RepoOwner, versionpkg.RepoName), nil)
	if err != nil {
		return latestRelease{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "crona/"+versionpkg.Current())

	resp, err := s.client.Do(req)
	if err != nil {
		return latestRelease{}, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return latestRelease{}, fmt.Errorf("github releases returned %s", resp.Status)
	}

	var payloads []githubReleasePayload
	if err := json.NewDecoder(resp.Body).Decode(&payloads); err != nil {
		return latestRelease{}, err
	}

	best := latestRelease{}
	bestVersion := semver{}
	found := false
	for _, payload := range payloads {
		if payload.Draft {
			continue
		}
		release, err := s.releaseFromPayload(payload)
		if err != nil {
			continue
		}
		version, ok := parseSemver(release.Version)
		if !ok {
			continue
		}
		if !found || compareSemver(version, bestVersion) > 0 {
			best = release
			bestVersion = version
			found = true
		}
	}
	if !found {
		return latestRelease{}, fmt.Errorf("no eligible releases found")
	}
	return best, nil
}

type githubReleasePayload struct {
	Name        string `json:"name"`
	TagName     string `json:"tag_name"`
	Body        string `json:"body"`
	HTMLURL     string `json:"html_url"`
	PublishedAt string `json:"published_at"`
	Prerelease  bool   `json:"prerelease"`
	Draft       bool   `json:"draft"`
	Assets      []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func (s *Service) releaseFromPayload(payload githubReleasePayload) (latestRelease, error) {
	tag := strings.TrimSpace(payload.TagName)
	version := normalizeVersion(payload.TagName)
	if version == "" {
		return latestRelease{}, fmt.Errorf("latest release tag is empty")
	}
	installURL := ""
	installAsset := config.InstallerAssetNameForGOOS(s.targetGOOS())
	checksumsURL := ""
	for _, asset := range payload.Assets {
		switch strings.TrimSpace(asset.Name) {
		case installAsset:
			installURL = strings.TrimSpace(asset.BrowserDownloadURL)
		case "checksums.txt":
			checksumsURL = strings.TrimSpace(asset.BrowserDownloadURL)
		}
	}
	return latestRelease{
		Version:      version,
		Tag:          tag,
		Name:         strings.TrimSpace(payload.Name),
		Notes:        strings.TrimSpace(payload.Body),
		URL:          strings.TrimSpace(payload.HTMLURL),
		InstallURL:   installURL,
		InstallAsset: installAsset,
		ChecksumsURL: checksumsURL,
		PublishedAt:  strings.TrimSpace(payload.PublishedAt),
		IsPrerelease: payload.Prerelease,
	}, nil
}

func (r latestRelease) installUnavailableReason() string {
	installerAsset := strings.TrimSpace(r.InstallAsset)
	if installerAsset == "" {
		installerAsset = config.InstallerAssetName()
	}
	switch {
	case strings.TrimSpace(r.InstallURL) == "" && strings.TrimSpace(r.ChecksumsURL) == "":
		return "Release is missing installer and checksums assets."
	case strings.TrimSpace(r.InstallURL) == "":
		return fmt.Sprintf("Release is missing the %s asset.", installerAsset)
	case strings.TrimSpace(r.ChecksumsURL) == "":
		return "Release is missing the checksums.txt asset."
	default:
		return ""
	}
}

func effectiveUpdateChannel(settings *sharedtypes.CoreSettings) sharedtypes.UpdateChannel {
	if settings == nil {
		return sharedtypes.UpdateChannelStable
	}
	return sharedtypes.NormalizeUpdateChannel(settings.UpdateChannel)
}

func (s *Service) targetGOOS() string {
	if strings.TrimSpace(s.goos) != "" {
		return s.goos
	}
	return runtime.GOOS
}

func (s *Service) hasLocalRelease() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.localRelease != nil
}

func (s *Service) currentLocalRelease() (latestRelease, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.localRelease == nil {
		return latestRelease{}, false
	}
	return *s.localRelease, true
}

func (s *Service) stopLocalReleaseServer() {
	s.mu.Lock()
	stop := s.stopLocalRelease
	s.stopLocalRelease = nil
	s.localRelease = nil
	s.localReleaseBase = ""
	s.mu.Unlock()
	if stop != nil {
		_ = stop()
	}
}
