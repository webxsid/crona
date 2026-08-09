package notify

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	runtimepkg "crona/kernel/internal/runtime"
	sharedtypes "crona/shared/types"
)

func TestCompanionBrokerReplacesActiveSubscriber(t *testing.T) {
	broker := newCompanionBroker()
	firstContext, cancelFirst := context.WithCancel(context.Background())
	defer cancelFirst()
	first, unsubscribeFirst := broker.subscribe(firstContext, sharedtypes.AlertDeliveryCapability{
		ClientID:      "first",
		Notifications: true,
	})
	defer unsubscribeFirst()

	secondContext, cancelSecond := context.WithCancel(context.Background())
	defer cancelSecond()
	second, unsubscribeSecond := broker.subscribe(
		secondContext,
		sharedtypes.AlertDeliveryCapability{
			ClientID:      "second",
			Notifications: true,
		},
	)
	defer unsubscribeSecond()

	if _, ok := <-first; ok {
		t.Fatal("expected replaced subscriber stream to close")
	}
	result, offered := broker.offer(sharedtypes.AlertDelivery{
		ID:                  "delivery-1",
		DeliverNotification: true,
	})
	if !offered {
		t.Fatal("expected delivery to be offered to active subscriber")
	}
	select {
	case delivery := <-second:
		if delivery.ID != "delivery-1" {
			t.Fatalf("unexpected delivery: %+v", delivery)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for active subscriber delivery")
	}
	if !broker.acknowledge(sharedtypes.AlertDeliveryAck{
		DeliveryID:           "delivery-1",
		NotificationAccepted: true,
	}) {
		t.Fatal("expected acknowledgement to resolve pending delivery")
	}
	if accepted := <-result; !accepted.notificationAccepted {
		t.Fatal("expected notification acknowledgement")
	}
}

func TestCompanionBrokerOnlyForwardsOfferedActions(t *testing.T) {
	broker := newCompanionBroker()
	deliveries, unsubscribe := broker.subscribe(context.Background(), sharedtypes.AlertDeliveryCapability{
		ClientID:      "mac",
		Notifications: true,
	})
	defer unsubscribe()
	result, offered := broker.offer(sharedtypes.AlertDelivery{
		ID:                  "delivery-action",
		DeliverNotification: true,
		Actions: []sharedtypes.AlertDeliveryAction{{
			ID:        "timer.defer_break",
			SessionID: "session-1",
		}},
	})
	if !offered {
		t.Fatal("expected action delivery to be offered")
	}
	<-deliveries
	if !broker.acknowledge(sharedtypes.AlertDeliveryAck{
		DeliveryID:    "delivery-action",
		ActionID:      "timer.fabricated",
		ActionSeconds: 120,
	}) {
		t.Fatal("expected delivery acknowledgement")
	}
	accepted := <-result
	if accepted.actionID != "" || accepted.sessionID != "" || accepted.actionSeconds != 0 {
		t.Fatalf("expected fabricated action to be discarded, got %+v", accepted)
	}
}

func TestCompanionAcknowledgementSuppressesLegacyDelivery(t *testing.T) {
	ctx := context.Background()
	coreCtx := testCoreContext(t, ctx)
	service := &Service{
		core:      coreCtx,
		logger:    testLogger(t),
		companion: newCompanionBroker(),
	}
	streamContext, cancelStream := context.WithCancel(ctx)
	defer cancelStream()
	deliveries, unsubscribe := service.SubscribeCompanion(
		streamContext,
		sharedtypes.AlertDeliveryCapability{
			ClientID:      "mac",
			Notifications: true,
			Sounds:        true,
		},
	)
	defer unsubscribe()

	legacyNotifications := 0
	legacySounds := 0
	restore := stubCompanionFallbacks(t, &legacyNotifications, &legacySounds)
	defer restore()

	go func() {
		delivery := <-deliveries
		service.AcknowledgeCompanion(sharedtypes.AlertDeliveryAck{
			DeliveryID:           delivery.ID,
			NotificationAccepted: true,
			SoundAccepted:        true,
		})
	}()

	err := service.deliver(ctx, sharedtypes.AlertRequest{
		Kind:      sharedtypes.AlertEventTimerWorkComplete,
		Title:     "Break",
		Body:      "Time to step away.",
		PlaySound: true,
	}, false)
	if err != nil {
		t.Fatalf("deliver through companion: %v", err)
	}
	if legacyNotifications != 0 || legacySounds != 0 {
		t.Fatalf(
			"expected no legacy delivery, got notifications=%d sounds=%d",
			legacyNotifications,
			legacySounds,
		)
	}
}

func TestCompanionPartialAcknowledgementFallsBackPerChannel(t *testing.T) {
	ctx := context.Background()
	coreCtx := testCoreContext(t, ctx)
	service := &Service{
		core:      coreCtx,
		logger:    testLogger(t),
		companion: newCompanionBroker(),
	}
	deliveries, unsubscribe := service.SubscribeCompanion(
		ctx,
		sharedtypes.AlertDeliveryCapability{
			ClientID:      "mac",
			Notifications: true,
			Sounds:        true,
		},
	)
	defer unsubscribe()

	legacyNotifications := 0
	legacySounds := 0
	restore := stubCompanionFallbacks(t, &legacyNotifications, &legacySounds)
	defer restore()

	go func() {
		delivery := <-deliveries
		service.AcknowledgeCompanion(sharedtypes.AlertDeliveryAck{
			DeliveryID:           delivery.ID,
			NotificationAccepted: true,
			SoundAccepted:        false,
		})
	}()

	err := service.deliver(ctx, sharedtypes.AlertRequest{
		Title:     "Break",
		Body:      "Time to step away.",
		PlaySound: true,
	}, false)
	if err != nil {
		t.Fatalf("deliver through companion: %v", err)
	}
	if legacyNotifications != 0 || legacySounds != 1 {
		t.Fatalf(
			"expected sound-only fallback, got notifications=%d sounds=%d",
			legacyNotifications,
			legacySounds,
		)
	}
}

func TestCompanionCannotAcknowledgeUnsupportedChannel(t *testing.T) {
	broker := newCompanionBroker()
	deliveries, unsubscribe := broker.subscribe(
		context.Background(),
		sharedtypes.AlertDeliveryCapability{
			ClientID:      "notifications-only",
			Notifications: true,
			Sounds:        false,
		},
	)
	defer unsubscribe()
	result, offered := broker.offer(sharedtypes.AlertDelivery{
		ID:                  "delivery-unsupported-sound",
		DeliverNotification: true,
		PlaySound:           true,
	})
	if !offered {
		t.Fatal("expected notification channel to be offered")
	}
	delivery := <-deliveries
	if delivery.PlaySound {
		t.Fatal("expected unsupported sound channel to be removed from delivery")
	}
	broker.acknowledge(sharedtypes.AlertDeliveryAck{
		DeliveryID:           delivery.ID,
		NotificationAccepted: true,
		SoundAccepted:        true,
	})
	accepted := <-result
	if !accepted.notificationAccepted || accepted.soundAccepted {
		t.Fatalf("unexpected accepted channels: %+v", accepted)
	}
}

func TestTimerBoundaryActionsOnlyForPreparedSegments(t *testing.T) {
	prepared := timerBoundaryActions(sharedtypes.TimerBoundaryPayload{
		To:      sharedtypes.SessionSegmentShortBreak,
		Started: false,
	})
	if len(prepared) != 1 || prepared[0].Title != "Start Break" {
		t.Fatalf("unexpected prepared break actions: %+v", prepared)
	}
	if prepared[0].ExpectedReadySegmentType == nil ||
		*prepared[0].ExpectedReadySegmentType != sharedtypes.SessionSegmentShortBreak {
		t.Fatalf("unexpected expected segment: %+v", prepared[0])
	}
	if actions := timerBoundaryActions(sharedtypes.TimerBoundaryPayload{
		To:      sharedtypes.SessionSegmentShortBreak,
		Started: true,
	}); len(actions) != 0 {
		t.Fatalf("expected no action for auto-started boundary: %+v", actions)
	}
}

func TestSoundTestUsesOnlySoundChannel(t *testing.T) {
	ctx := context.Background()
	coreCtx := testCoreContext(t, ctx)
	service := &Service{core: coreCtx, logger: testLogger(t)}
	legacyNotifications := 0
	legacySounds := 0
	restore := stubCompanionFallbacks(t, &legacyNotifications, &legacySounds)
	defer restore()

	err := service.deliver(ctx, sharedtypes.AlertRequest{
		Kind:        sharedtypes.AlertEventTestSound,
		Title:       "Test sound",
		Body:        "Playing a sound.",
		SoundPreset: sharedtypes.AlertSoundPresetChime,
		PlaySound:   true,
	}, false)
	if err != nil {
		t.Fatalf("deliver sound test: %v", err)
	}
	if legacyNotifications != 0 || legacySounds != 1 {
		t.Fatalf(
			"expected sound-only delivery, got notifications=%d sounds=%d",
			legacyNotifications,
			legacySounds,
		)
	}
}

func stubCompanionFallbacks(
	t *testing.T,
	notifications *int,
	sounds *int,
) func() {
	t.Helper()
	alertStatusFn = func(paths runtimepkg.Paths) sharedtypes.AlertStatus {
		return sharedtypes.AlertStatus{
			NotificationsAvailable: true,
			SoundAvailable:         true,
		}
	}
	alertSoundPathFn = func(paths runtimepkg.Paths, preset sharedtypes.AlertSoundPreset) (string, error) {
		return filepath.Join(t.TempDir(), "sound.caf"), nil
	}
	sendAlertNotificationFn = func(status sharedtypes.AlertStatus, req sharedtypes.AlertRequest) error {
		*notifications++
		return nil
	}
	playAlertSoundFn = func(status sharedtypes.AlertStatus, soundPath string) error {
		*sounds++
		return nil
	}
	return func() {
		alertStatusFn = detectAlertStatus
		alertSoundPathFn = alertSoundPath
		sendAlertNotificationFn = sendAlertNotification
		playAlertSoundFn = playAlertSound
	}
}
