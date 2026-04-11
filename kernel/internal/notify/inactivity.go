package notify

import (
	"context"
	"strings"
	"time"

	sharedtypes "crona/shared/types"
)

func (s *Service) TouchTimerActivity(ctx context.Context) error {
	active, err := s.core.Sessions.GetActiveSession(ctx, s.core.UserID)
	if err != nil || active == nil {
		return err
	}
	s.markTimerActivity(active.ID, alertNowFn())
	return nil
}

func (s *Service) runInactivityScheduler(ctx context.Context) {
	s.processInactivityTick(ctx, alertNowFn())
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			s.processInactivityTick(ctx, now)
		}
	}
}

func (s *Service) processInactivityTick(ctx context.Context, now time.Time) {
	settings, err := s.core.CoreSettings.Get(ctx, s.core.UserID)
	if err != nil {
		s.logger.Error("load inactivity settings", err)
		return
	}
	active, err := s.core.Sessions.GetActiveSession(ctx, s.core.UserID)
	if err != nil {
		s.logger.Error("load active session for inactivity", err)
		return
	}
	if active == nil || settings == nil || !settings.InactivityAlerts {
		s.clearTimerActivity()
		return
	}
	if !s.isActiveWorkSegment(ctx, active.ID) {
		s.markTimerActivity(active.ID, now)
		return
	}

	startedAt := parseSessionTime(active.StartTime, now)
	lastActivity := s.inactivityBaseline(active.ID, startedAt)
	threshold := time.Duration(clampAlertMinutes(settings.InactivityThreshold, 60)) * time.Minute
	if now.Sub(lastActivity) < threshold {
		return
	}
	repeat := time.Duration(clampAlertMinutes(settings.InactivityRepeat, 60)) * time.Minute
	if !s.shouldDeliverInactivity(active.ID, now, repeat) {
		return
	}
	if err := s.deliver(ctx, s.inactivityAlert(ctx, active.IssueID), true); err != nil {
		s.logger.Error("deliver inactivity alert", err)
		return
	}
	s.markInactivityDelivered(active.ID, now)
}

func (s *Service) isActiveWorkSegment(ctx context.Context, sessionID string) bool {
	segment, err := s.core.SessionSegments.GetActive(ctx, s.core.UserID, s.core.DeviceID, sessionID)
	if err != nil {
		s.logger.Error("load active segment for inactivity", err)
		return false
	}
	return segment == nil || segment.SegmentType == sharedtypes.SessionSegmentWork
}

func (s *Service) inactivityAlert(ctx context.Context, issueID int64) sharedtypes.AlertRequest {
	req := sharedtypes.AlertRequest{
		Kind:        sharedtypes.AlertEventFocusInactivity,
		Title:       "Focus session still running",
		Body:        "Review the session and pause, stash, or end it if you stepped away.",
		Urgency:     sharedtypes.AlertUrgencyNormal,
		IconEnabled: true,
		PlaySound:   false,
	}
	if subtitle := s.inactivitySubtitle(ctx, issueID); subtitle != "" {
		req.Subtitle = subtitle
	}
	return req
}

func (s *Service) inactivitySubtitle(ctx context.Context, issueID int64) string {
	parts := make([]string, 0, 3)
	activeContext, err := s.core.ActiveContext.Get(ctx, s.core.UserID, s.core.DeviceID)
	if err == nil && activeContext != nil {
		if activeContext.RepoName != nil && strings.TrimSpace(*activeContext.RepoName) != "" {
			parts = append(parts, strings.TrimSpace(*activeContext.RepoName))
		}
		if activeContext.StreamName != nil && strings.TrimSpace(*activeContext.StreamName) != "" {
			parts = append(parts, strings.TrimSpace(*activeContext.StreamName))
		}
		if activeContext.IssueTitle != nil && strings.TrimSpace(*activeContext.IssueTitle) != "" {
			parts = append(parts, strings.TrimSpace(*activeContext.IssueTitle))
		}
	}
	if len(parts) == 0 {
		if issue, err := s.core.Issues.GetByID(ctx, issueID, s.core.UserID); err == nil && issue != nil {
			if title := strings.TrimSpace(issue.Title); title != "" {
				parts = append(parts, title)
			}
		}
	}
	return strings.Join(parts, " / ")
}

func (s *Service) markTimerActivity(sessionID string, at time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.activitySessionID = sessionID
	s.lastActivityAt = at
}

func (s *Service) clearTimerActivity() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.activitySessionID = ""
	s.lastActivityAt = time.Time{}
	s.inactivitySession = ""
	s.lastInactivityAt = time.Time{}
}

func (s *Service) inactivityBaseline(sessionID string, startedAt time.Time) time.Time {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.activitySessionID != sessionID || s.lastActivityAt.IsZero() {
		s.activitySessionID = sessionID
		s.lastActivityAt = startedAt
	}
	return s.lastActivityAt
}

func (s *Service) shouldDeliverInactivity(sessionID string, now time.Time, repeat time.Duration) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.inactivitySession != sessionID || s.lastInactivityAt.IsZero() || now.Sub(s.lastInactivityAt) >= repeat
}

func (s *Service) markInactivityDelivered(sessionID string, at time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.inactivitySession = sessionID
	s.lastInactivityAt = at
}

func parseSessionTime(value string, fallback time.Time) time.Time {
	parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(value))
	if err != nil {
		return fallback
	}
	return parsed
}

func clampAlertMinutes(value int, fallback int) int {
	if value == 0 {
		value = fallback
	}
	if value < 15 {
		return 15
	}
	if value > 720 {
		return 720
	}
	return value
}
