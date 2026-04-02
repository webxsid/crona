package updatecheck

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"crona/kernel/internal/core"
	"crona/kernel/internal/events"
	runtimepkg "crona/kernel/internal/runtime"
	"crona/shared/config"
	sharedtypes "crona/shared/types"
	versionpkg "crona/shared/version"
)

const checkInterval = 24 * time.Hour

type Service struct {
	core      *core.Context
	bus       *events.Bus
	logger    *runtimepkg.Logger
	cachePath string
	envMode   string
	goos      string
	client    *http.Client

	mu     sync.RWMutex
	status sharedtypes.UpdateStatus
}

func Start(ctx context.Context, coreCtx *core.Context, bus *events.Bus, logger *runtimepkg.Logger, paths runtimepkg.Paths, envMode string) *Service {
	service := &Service{
		core:      coreCtx,
		bus:       bus,
		logger:    logger,
		cachePath: paths.UpdateFile,
		envMode:   envMode,
		goos:      runtime.GOOS,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		status: sharedtypes.UpdateStatus{
			CurrentVersion: versionpkg.Current(),
			Channel:        sharedtypes.UpdateChannelStable,
		},
	}
	service.loadCache()
	go service.run(ctx)
	return service
}

func (s *Service) Status() sharedtypes.UpdateStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.status
}

func (s *Service) CheckNow(ctx context.Context) (sharedtypes.UpdateStatus, error) {
	return s.refresh(ctx, true)
}

func (s *Service) DismissLatest() (sharedtypes.UpdateStatus, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if strings.TrimSpace(s.status.LatestVersion) != "" {
		s.status.DismissedVersion = s.status.LatestVersion
	}
	if err := s.persistLocked(); err != nil {
		return s.status, err
	}
	s.emitLocked()
	return s.status, nil
}

func (s *Service) run(ctx context.Context) {
	initialCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	_, err := s.refresh(initialCtx, false)
	cancel()
	if err != nil {
		s.logger.Error("initial update check", err)
	}

	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			checkCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
			_, err := s.refresh(checkCtx, false)
			cancel()
			if err != nil {
				s.logger.Error("scheduled update check", err)
			}
		}
	}
}

func (s *Service) refresh(ctx context.Context, force bool) (sharedtypes.UpdateStatus, error) {
	settings, err := s.core.CoreSettings.Get(ctx, s.core.UserID)
	if err != nil {
		return s.Status(), err
	}

	s.mu.Lock()
	s.status.CurrentVersion = versionpkg.Current()
	s.status.Enabled = settings != nil && settings.UpdateChecksEnabled && !strings.EqualFold(s.envMode, config.ModeDev) && !versionpkg.IsDevBuild()
	s.status.PromptEnabled = settings != nil && settings.UpdatePromptEnabled && s.status.Enabled
	s.status.Channel = effectiveUpdateChannel(settings)

	if !s.status.Enabled {
		s.clearReleaseLocked()
		s.status.Error = ""
		err := s.persistLocked()
		s.emitLocked()
		status := s.status
		s.mu.Unlock()
		return status, err
	}

	if !force && isFresh(s.status.CheckedAt, checkInterval) {
		err := s.persistLocked()
		s.emitLocked()
		status := s.status
		s.mu.Unlock()
		return status, err
	}
	s.mu.Unlock()

	channel := effectiveUpdateChannel(settings)
	release, err := s.fetchLatestRelease(ctx, channel)

	s.mu.Lock()
	defer s.mu.Unlock()
	s.status.CurrentVersion = versionpkg.Current()
	s.status.Enabled = settings != nil && settings.UpdateChecksEnabled && !strings.EqualFold(s.envMode, config.ModeDev) && !versionpkg.IsDevBuild()
	s.status.PromptEnabled = settings != nil && settings.UpdatePromptEnabled && s.status.Enabled
	s.status.Channel = channel
	s.status.CheckedAt = time.Now().UTC().Format(time.RFC3339)

	if err != nil {
		s.clearReleaseLocked()
		s.status.Error = err.Error()
		if persistErr := s.persistLocked(); persistErr != nil {
			return s.status, persistErr
		}
		s.emitLocked()
		return s.status, err
	}

	s.status.Error = ""
	s.status.LatestVersion = release.Version
	s.status.ReleaseTag = release.Tag
	s.status.ReleaseName = release.Name
	s.status.ReleaseNotes = release.Notes
	s.status.ReleaseURL = release.URL
	s.status.InstallScriptURL = release.InstallURL
	s.status.ChecksumsURL = release.ChecksumsURL
	s.status.PublishedAt = release.PublishedAt
	s.status.ReleaseIsPrerelease = release.IsPrerelease
	s.status.UpdateAvailable = isNewerVersion(s.status.CurrentVersion, release.Version)
	s.status.InstallAvailable = release.InstallURL != "" && release.ChecksumsURL != ""
	s.status.InstallUnavailableReason = release.installUnavailableReason()
	if s.status.DismissedVersion != "" && s.status.DismissedVersion != s.status.LatestVersion {
		s.status.DismissedVersion = ""
	}

	if err := s.persistLocked(); err != nil {
		return s.status, err
	}
	s.emitLocked()
	return s.status, nil
}

func (s *Service) loadCache() {
	body, err := os.ReadFile(s.cachePath)
	if err != nil {
		return
	}
	var cached sharedtypes.UpdateStatus
	if err := json.Unmarshal(body, &cached); err != nil {
		s.logger.Error("decode update cache", err)
		return
	}
	cached.CurrentVersion = versionpkg.Current()
	cached.Channel = sharedtypes.NormalizeUpdateChannel(cached.Channel)
	s.status = cached
}

func (s *Service) persistLocked() error {
	body, err := json.MarshalIndent(s.status, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.cachePath, body, runtimepkg.FilePerm())
}

func (s *Service) clearReleaseLocked() {
	s.status.LatestVersion = ""
	s.status.ReleaseTag = ""
	s.status.ReleaseName = ""
	s.status.ReleaseNotes = ""
	s.status.ReleaseURL = ""
	s.status.InstallScriptURL = ""
	s.status.ChecksumsURL = ""
	s.status.PublishedAt = ""
	s.status.ReleaseIsPrerelease = false
	s.status.UpdateAvailable = false
	s.status.InstallAvailable = false
	s.status.InstallUnavailableReason = ""
}

func (s *Service) emitLocked() {
	body, err := json.Marshal(s.status)
	if err != nil {
		s.logger.Error("encode update status event", err)
		return
	}
	s.bus.Emit(sharedtypes.KernelEvent{
		Type:    sharedtypes.EventTypeUpdateStatus,
		Payload: body,
	})
}

func isFresh(checkedAt string, maxAge time.Duration) bool {
	if strings.TrimSpace(checkedAt) == "" {
		return false
	}
	ts, err := time.Parse(time.RFC3339, checkedAt)
	if err != nil {
		return false
	}
	return time.Since(ts) < maxAge
}

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
