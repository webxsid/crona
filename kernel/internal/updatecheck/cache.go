package updatecheck

import (
	"encoding/json"
	"os"
	"strings"
	"time"

	runtimepkg "crona/kernel/internal/runtime"
	sharedtypes "crona/shared/types"
	"crona/shared/version"
)

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
	cached.CurrentVersion = version.Current()
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

func (s *Service) persistAndEmitIfChangedLocked(prev sharedtypes.UpdateStatus) error {
	if s.status == prev {
		return nil
	}
	if err := s.persistLocked(); err != nil {
		return err
	}
	s.emitLocked()
	return nil
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
