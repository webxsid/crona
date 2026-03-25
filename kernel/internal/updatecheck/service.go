package updatecheck

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
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
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		status: sharedtypes.UpdateStatus{
			CurrentVersion: versionpkg.Current(),
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

	release, err := s.fetchLatestRelease(ctx)

	s.mu.Lock()
	defer s.mu.Unlock()
	s.status.CurrentVersion = versionpkg.Current()
	s.status.Enabled = settings != nil && settings.UpdateChecksEnabled && !strings.EqualFold(s.envMode, config.ModeDev) && !versionpkg.IsDevBuild()
	s.status.PromptEnabled = settings != nil && settings.UpdatePromptEnabled && s.status.Enabled
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
	ChecksumsURL string
	PublishedAt  string
}

func (s *Service) fetchLatestRelease(ctx context.Context) (latestRelease, error) {
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

	var payload struct {
		Name        string `json:"name"`
		TagName     string `json:"tag_name"`
		Body        string `json:"body"`
		HTMLURL     string `json:"html_url"`
		PublishedAt string `json:"published_at"`
		Assets      []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return latestRelease{}, err
	}

	tag := strings.TrimSpace(payload.TagName)
	version := normalizeVersion(payload.TagName)
	if version == "" {
		return latestRelease{}, fmt.Errorf("latest release tag is empty")
	}
	installURL := ""
	checksumsURL := ""
	for _, asset := range payload.Assets {
		switch strings.TrimSpace(asset.Name) {
		case "install-crona-tui.sh":
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
		ChecksumsURL: checksumsURL,
		PublishedAt:  strings.TrimSpace(payload.PublishedAt),
	}, nil
}

func (r latestRelease) installUnavailableReason() string {
	switch {
	case strings.TrimSpace(r.InstallURL) == "" && strings.TrimSpace(r.ChecksumsURL) == "":
		return "Release is missing installer and checksums assets."
	case strings.TrimSpace(r.InstallURL) == "":
		return "Release is missing the install-crona-tui.sh asset."
	case strings.TrimSpace(r.ChecksumsURL) == "":
		return "Release is missing the checksums.txt asset."
	default:
		return ""
	}
}
