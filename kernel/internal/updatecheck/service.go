package updatecheck

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"time"

	"crona/kernel/internal/core"
	"crona/kernel/internal/events"
	runtimepkg "crona/kernel/internal/runtime"
	"crona/shared/config"
	shareddto "crona/shared/dto"
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

	mu               sync.RWMutex
	status           sharedtypes.UpdateStatus
	localRelease     *latestRelease
	localReleaseBase string
	stopLocalRelease func() error
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
	go func() {
		<-ctx.Done()
		service.stopLocalReleaseServer()
	}()
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

func (s *Service) PrepareLocalRelease(ctx context.Context) (shareddto.LocalUpdatePreparedResponse, error) {
	if !strings.EqualFold(strings.TrimSpace(s.envMode), config.ModeDev) {
		return shareddto.LocalUpdatePreparedResponse{}, fmt.Errorf("local update simulation is only available in Dev mode")
	}
	release, releaseDir, baseURL, err := s.prepareLocalReleaseSource(ctx)
	if err != nil {
		return shareddto.LocalUpdatePreparedResponse{}, err
	}
	status, err := s.refresh(ctx, true)
	if err != nil {
		return shareddto.LocalUpdatePreparedResponse{}, err
	}
	if !status.InstallAvailable {
		return shareddto.LocalUpdatePreparedResponse{}, fmt.Errorf("local release %s is missing required installer assets", release.Version)
	}
	return shareddto.LocalUpdatePreparedResponse{
		Version:    release.Version,
		Tag:        release.Tag,
		ReleaseDir: releaseDir,
		BaseURL:    baseURL,
	}, nil
}

func (s *Service) DismissLatest() (sharedtypes.UpdateStatus, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	prev := s.status
	if strings.TrimSpace(s.status.LatestVersion) != "" {
		s.status.DismissedVersion = s.status.LatestVersion
	}
	if err := s.persistAndEmitIfChangedLocked(prev); err != nil {
		return s.status, err
	}
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
	localOverrideActive := s.hasLocalRelease()
	enabled := settings != nil && settings.UpdateChecksEnabled
	promptEnabled := settings != nil && settings.UpdatePromptEnabled
	if localOverrideActive {
		enabled = true
		promptEnabled = true
	}

	s.mu.Lock()
	prev := s.status
	s.status.CurrentVersion = versionpkg.Current()
	s.status.Enabled = enabled && (!strings.EqualFold(s.envMode, config.ModeDev) || localOverrideActive) && (!versionpkg.IsDevBuild() || localOverrideActive)
	s.status.PromptEnabled = promptEnabled && s.status.Enabled
	s.status.Channel = effectiveUpdateChannel(settings)

	if !s.status.Enabled {
		s.clearReleaseLocked()
		s.status.Error = ""
		err := s.persistAndEmitIfChangedLocked(prev)
		status := s.status
		s.mu.Unlock()
		return status, err
	}

	if !force && isFresh(s.status.CheckedAt, checkInterval) {
		err := s.persistAndEmitIfChangedLocked(prev)
		status := s.status
		s.mu.Unlock()
		return status, err
	}
	s.mu.Unlock()

	channel := effectiveUpdateChannel(settings)
	release, err := s.fetchLatestRelease(ctx, channel)

	s.mu.Lock()
	defer s.mu.Unlock()
	prev = s.status
	s.status.CurrentVersion = versionpkg.Current()
	s.status.Enabled = enabled && (!strings.EqualFold(s.envMode, config.ModeDev) || localOverrideActive) && (!versionpkg.IsDevBuild() || localOverrideActive)
	s.status.PromptEnabled = promptEnabled && s.status.Enabled
	s.status.Channel = channel
	s.status.CheckedAt = time.Now().UTC().Format(time.RFC3339)

	if err != nil {
		s.clearReleaseLocked()
		s.status.Error = err.Error()
		if persistErr := s.persistAndEmitIfChangedLocked(prev); persistErr != nil {
			return s.status, persistErr
		}
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

	if err := s.persistAndEmitIfChangedLocked(prev); err != nil {
		return s.status, err
	}
	return s.status, nil
}
