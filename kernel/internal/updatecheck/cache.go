package updatecheck

import (
	"encoding/json"
	"strings"
	"time"

	sharedtypes "crona/shared/types"
)

func (s *Service) emitIfChangedLocked(prev sharedtypes.UpdateStatus) error {
	if s.status == prev {
		return nil
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
	s.status.LatestIsBeta = false
	s.status.UpdateAvailable = false
	s.status.InstallAvailable = false
	s.status.InstallUnavailableReason = ""
	s.status.UpdateCommand = ""
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
