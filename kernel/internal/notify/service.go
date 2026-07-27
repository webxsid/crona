package notify

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"crona/kernel/internal/core"
	"crona/kernel/internal/events"
	runtimepkg "crona/kernel/internal/runtime"
	sharedtypes "crona/shared/types"
)

type Service struct {
	core      *core.Context
	bus       *events.Bus
	logger    *runtimepkg.Logger
	paths     runtimepkg.Paths
	queue     chan queuedAlert
	companion *companionBroker

	mu                 sync.Mutex
	lastUpdateNotified string
	lastReminderSlots  map[string]string
	activitySessionID  string
	lastActivityAt     time.Time
	inactivitySession  string
	lastInactivityAt   time.Time
}

var (
	alertStatusFn           = detectAlertStatus
	alertSoundPathFn        = alertSoundPath
	sendAlertNotificationFn = sendAlertNotification
	playAlertSoundFn        = playAlertSound
	companionAckTimeout     = 1500 * time.Millisecond
)

type queuedAlert struct {
	req             sharedtypes.AlertRequest
	respectSettings bool
	actions         []sharedtypes.AlertDeliveryAction
}

func Start(
	ctx context.Context,
	coreCtx *core.Context,
	bus *events.Bus,
	logger *runtimepkg.Logger,
	paths runtimepkg.Paths,
) *Service {
	service := &Service{
		core:              coreCtx,
		bus:               bus,
		logger:            logger,
		paths:             paths,
		queue:             make(chan queuedAlert, 32),
		companion:         newCompanionBroker(),
		lastReminderSlots: make(map[string]string),
	}
	unsubscribe := bus.Subscribe(func(event sharedtypes.KernelEvent) {
		switch event.Type {
		case sharedtypes.EventTypeTimerBoundary:
			var payload sharedtypes.TimerBoundaryPayload
			if err := json.Unmarshal(event.Payload, &payload); err != nil {
				logger.Error("decode timer boundary payload", err)
				return
			}
			if err := service.enqueueWithActions(
				timerBoundaryAlert(payload),
				true,
				timerBoundaryActions(payload),
			); err != nil {
				logger.Error("enqueue timer boundary alert", err)
			}
		case sharedtypes.EventTypeTimerHardLimitReached:
			var payload sharedtypes.TimerHardLimitReachedPayload
			if err := json.Unmarshal(event.Payload, &payload); err != nil {
				logger.Error("decode timer hard limit payload", err)
				return
			}
			if err := service.enqueue(hardLimitReachedAlert(payload), true); err != nil {
				logger.Error("enqueue timer hard limit alert", err)
			}
		case sharedtypes.EventTypeUpdateStatus:
			var status sharedtypes.UpdateStatus
			if err := json.Unmarshal(event.Payload, &status); err != nil {
				logger.Error("decode update status payload", err)
				return
			}
			if req, ok := service.updateAvailableAlert(status); ok {
				if err := service.enqueue(req, true); err != nil {
					logger.Error("enqueue update status alert", err)
				}
			}
		}
	})
	go func() {
		defer unsubscribe()
		for {
			select {
			case <-ctx.Done():
				return
			case req, ok := <-service.queue:
				if !ok {
					return
				}
				if err := service.deliverWithActions(
					ctx,
					req.req,
					req.respectSettings,
					req.actions,
				); err != nil {
					service.logger.Error("deliver alert", err)
				}
			}
		}
	}()
	go service.runReminderScheduler(ctx)
	go service.runInactivityScheduler(ctx)
	return service
}

func (s *Service) Status() sharedtypes.AlertStatus {
	status := alertStatusFn(s.paths)
	status.CompanionDeliverySupported = true
	status.CompanionDeliveryActive = s.companionBroker().isActive()
	if status.CompanionDeliveryActive {
		status.NotificationsAvailable = true
		status.SoundAvailable = true
		status.NotificationBackend = "companion"
		status.SoundBackend = "companion"
	}
	return status
}

func (s *Service) SubscribeCompanion(
	ctx context.Context,
	capabilities sharedtypes.AlertDeliveryCapability,
) (<-chan sharedtypes.AlertDelivery, func()) {
	return s.companionBroker().subscribe(ctx, capabilities)
}

func (s *Service) AcknowledgeCompanion(ack sharedtypes.AlertDeliveryAck) bool {
	return s.companionBroker().acknowledge(ack)
}

func (s *Service) companionBroker() *companionBroker {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.companion == nil {
		s.companion = newCompanionBroker()
	}
	return s.companion
}

func (s *Service) TestNotification(ctx context.Context) error {
	settings, err := s.core.CoreSettings.Get(ctx, s.core.UserID)
	if err != nil {
		return err
	}
	req := sharedtypes.AlertRequest{
		Kind:        sharedtypes.AlertEventTestNotification,
		Title:       "Crona test notification",
		Subtitle:    "Alerts layer check",
		Body:        "Notifications are configured and routed through the richer alerts layer.",
		Urgency:     sharedtypes.AlertUrgencyNormal,
		IconEnabled: settings != nil && settings.AlertIconEnabled,
		PlaySound:   false,
	}
	return s.enqueue(req, false)
}

func (s *Service) TestSound(ctx context.Context) error {
	settings, err := s.core.CoreSettings.Get(ctx, s.core.UserID)
	if err != nil {
		return err
	}
	req := sharedtypes.AlertRequest{
		Kind:        sharedtypes.AlertEventTestSound,
		Title:       "Crona test sound",
		Subtitle:    "Alerts layer check",
		Body:        "Playing the selected bundled alert preset.",
		Urgency:     sharedtypes.AlertUrgencyNormal,
		IconEnabled: settings != nil && settings.AlertIconEnabled,
		PlaySound:   true,
	}
	return s.enqueue(req, false)
}

func (s *Service) Notify(ctx context.Context, req sharedtypes.AlertRequest) error {
	return s.enqueue(req, true)
}

func (s *Service) enqueue(req sharedtypes.AlertRequest, respectSettings bool) error {
	return s.enqueueWithActions(req, respectSettings, nil)
}

func (s *Service) enqueueWithActions(
	req sharedtypes.AlertRequest,
	respectSettings bool,
	actions []sharedtypes.AlertDeliveryAction,
) error {
	select {
	case s.queue <- queuedAlert{
		req:             req,
		respectSettings: respectSettings,
		actions:         actions,
	}:
		return nil
	default:
		return fmt.Errorf("alert queue full")
	}
}

func timerBoundaryActions(
	payload sharedtypes.TimerBoundaryPayload,
) []sharedtypes.AlertDeliveryAction {
	if payload.Started {
		return nil
	}
	switch payload.To {
	case sharedtypes.SessionSegmentShortBreak, sharedtypes.SessionSegmentLongBreak:
		expected := payload.To
		return []sharedtypes.AlertDeliveryAction{{
			ID:                       "timer.advance",
			Title:                    "Start Break",
			ExpectedReadySegmentType: &expected,
		}}
	case sharedtypes.SessionSegmentWork:
		expected := payload.To
		return []sharedtypes.AlertDeliveryAction{{
			ID:                       "timer.advance",
			Title:                    "Resume Focus",
			ExpectedReadySegmentType: &expected,
		}}
	default:
		return nil
	}
}

func (s *Service) updateAvailableAlert(
	status sharedtypes.UpdateStatus,
) (sharedtypes.AlertRequest, bool) {
	if !status.UpdateAvailable || strings.TrimSpace(status.LatestVersion) == "" {
		return sharedtypes.AlertRequest{}, false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if status.LatestVersion == s.lastUpdateNotified {
		return sharedtypes.AlertRequest{}, false
	}
	s.lastUpdateNotified = status.LatestVersion
	subtitle := "Stable release"
	if status.LatestIsBeta {
		subtitle = "Beta release"
	}
	return sharedtypes.AlertRequest{
		Kind:      sharedtypes.AlertEventUpdateAvailable,
		Title:     "Crona update available",
		Subtitle:  subtitle,
		Body:      "Version " + strings.TrimSpace(status.LatestVersion) + " is ready to install.",
		Urgency:   sharedtypes.AlertUrgencyLow,
		PlaySound: false,
	}, true
}

func timerBoundaryAlert(payload sharedtypes.TimerBoundaryPayload) sharedtypes.AlertRequest {
	req := sharedtypes.AlertRequest{
		Title:       strings.TrimSpace(payload.Title),
		Body:        strings.TrimSpace(payload.Message),
		PlaySound:   true,
		Urgency:     sharedtypes.AlertUrgencyNormal,
		IconEnabled: true,
	}
	if req.Title == "" {
		req.Title = "Timer boundary reached"
	}
	if req.Body == "" {
		req.Body = "Structured timer boundary reached"
	}
	subtitle := timerBoundarySubtitle(payload)
	if subtitle != "" {
		req.Subtitle = subtitle
	}
	switch payload.To {
	case sharedtypes.SessionSegmentShortBreak, sharedtypes.SessionSegmentLongBreak:
		req.Kind = sharedtypes.AlertEventTimerWorkComplete
		req.SoundPreset = sharedtypes.AlertSoundPresetFocusGong
		req.Urgency = sharedtypes.AlertUrgencyHigh
	case sharedtypes.SessionSegmentWork:
		req.Kind = sharedtypes.AlertEventTimerBreakComplete
		req.SoundPreset = sharedtypes.AlertSoundPresetSoftBell
	default:
		req.Kind = sharedtypes.AlertEventTimerWorkComplete
	}
	return req
}

func hardLimitReachedAlert(
	payload sharedtypes.TimerHardLimitReachedPayload,
) sharedtypes.AlertRequest {
	if sharedtypes.NormalizeTimerHardLimitKind(payload.HardLimitKind) ==
		sharedtypes.TimerHardLimitKindCountdown {
		return sharedtypes.AlertRequest{
			Kind:        sharedtypes.AlertEventTimerWorkComplete,
			Title:       "Timer complete",
			Subtitle:    "Commit or extend to continue",
			Body:        "The countdown has finished. Commit to finish the session or extend to keep working.",
			Urgency:     sharedtypes.AlertUrgencyHigh,
			SoundPreset: sharedtypes.AlertSoundPresetFocusGong,
			PlaySound:   true,
			IconEnabled: true,
		}
	}
	req := sharedtypes.AlertRequest{
		Kind:        sharedtypes.AlertEventTimerWorkComplete,
		Title:       "Pomodoro session complete",
		Subtitle:    "Commit or extend to continue",
		Body:        "The pomodoro session has reached its hard limit. Commit to finish it or extend to keep working.",
		Urgency:     sharedtypes.AlertUrgencyHigh,
		SoundPreset: sharedtypes.AlertSoundPresetFocusGong,
		PlaySound:   true,
		IconEnabled: true,
	}
	return req
}

func timerBoundarySubtitle(payload sharedtypes.TimerBoundaryPayload) string {
	parts := make([]string, 0, 3)
	if payload.RepoName != nil && strings.TrimSpace(*payload.RepoName) != "" {
		parts = append(parts, strings.TrimSpace(*payload.RepoName))
	}
	if payload.StreamName != nil && strings.TrimSpace(*payload.StreamName) != "" {
		parts = append(parts, strings.TrimSpace(*payload.StreamName))
	}
	if payload.IssueTitle != nil && strings.TrimSpace(*payload.IssueTitle) != "" {
		parts = append(parts, strings.TrimSpace(*payload.IssueTitle))
	}
	return strings.Join(parts, " / ")
}

func (s *Service) deliver(
	ctx context.Context,
	req sharedtypes.AlertRequest,
	respectSettings bool,
) error {
	return s.deliverWithActions(ctx, req, respectSettings, nil)
}

func (s *Service) deliverWithActions(
	ctx context.Context,
	req sharedtypes.AlertRequest,
	respectSettings bool,
	actions []sharedtypes.AlertDeliveryAction,
) error {
	settings, err := s.core.CoreSettings.Get(ctx, s.core.UserID)
	if err != nil {
		return err
	}
	status := alertStatusFn(s.paths)
	if settings == nil {
		settings = &sharedtypes.CoreSettings{
			BoundaryNotifications: true,
			BoundarySound:         true,
			AlertSoundPreset:      sharedtypes.AlertSoundPresetChime,
			AlertUrgency:          sharedtypes.AlertUrgencyNormal,
			AlertIconEnabled:      true,
		}
	}

	req = normalizeRequest(req, settings)
	if respectSettings && !settings.BoundaryNotifications &&
		(!settings.BoundarySound || !req.PlaySound) {
		return nil
	}

	deliverNotification := !respectSettings || settings.BoundaryNotifications
	if req.Kind == sharedtypes.AlertEventTestSound {
		deliverNotification = false
	}
	deliverSound := req.PlaySound && (!respectSettings || settings.BoundarySound)
	notificationAccepted := false
	soundAccepted := false

	deliveryID := fmt.Sprintf("alert-%d", time.Now().UnixNano())
	result, offered := s.companionBroker().offer(sharedtypes.AlertDelivery{
		ID:                  deliveryID,
		Alert:               req,
		DeliverNotification: deliverNotification,
		PlaySound:           deliverSound,
		Actions:             actions,
	})
	if offered {
		timer := time.NewTimer(companionAckTimeout)
		select {
		case accepted := <-result:
			notificationAccepted = accepted.notificationAccepted
			soundAccepted = accepted.soundAccepted
			if !timer.Stop() {
				<-timer.C
			}
		case <-timer.C:
			s.companionBroker().abandon(deliveryID)
		case <-ctx.Done():
			s.companionBroker().abandon(deliveryID)
			return ctx.Err()
		}
	}

	var firstErr error
	if deliverNotification && !notificationAccepted && status.NotificationsAvailable {
		if err := sendAlertNotificationFn(status, req); err != nil {
			s.logger.Error("send alert notification", err)
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	if deliverSound && !soundAccepted && status.SoundAvailable {
		soundPath, err := alertSoundPathFn(s.paths, req.SoundPreset)
		if err != nil {
			return err
		}
		if err := playAlertSoundFn(status, soundPath); err != nil {
			s.logger.Error("play alert sound", err)
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

func normalizeRequest(
	req sharedtypes.AlertRequest,
	settings *sharedtypes.CoreSettings,
) sharedtypes.AlertRequest {
	req.Title = strings.TrimSpace(req.Title)
	req.Subtitle = strings.TrimSpace(req.Subtitle)
	req.Body = strings.TrimSpace(req.Body)
	if req.Title == "" {
		req.Title = "Crona alert"
	}
	if req.Body == "" {
		req.Body = req.Title
	}
	if strings.TrimSpace(string(req.Urgency)) == "" {
		req.Urgency = sharedtypes.NormalizeAlertUrgency(settings.AlertUrgency)
	} else {
		req.Urgency = sharedtypes.NormalizeAlertUrgency(req.Urgency)
	}
	if strings.TrimSpace(string(req.SoundPreset)) == "" {
		req.SoundPreset = sharedtypes.NormalizeAlertSoundPreset(settings.AlertSoundPreset)
	} else {
		req.SoundPreset = sharedtypes.NormalizeAlertSoundPreset(req.SoundPreset)
	}
	req.IconEnabled = settings.AlertIconEnabled
	return req
}
