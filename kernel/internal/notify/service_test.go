package notify

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"crona/kernel/internal/core"
	corecommands "crona/kernel/internal/core/commands"
	"crona/kernel/internal/events"
	runtimepkg "crona/kernel/internal/runtime"
	"crona/kernel/internal/store"
	dbpkg "crona/kernel/internal/store/db"
	"crona/kernel/internal/store/migrations"
	"crona/kernel/internal/store/repositories"
	sharedtypes "crona/shared/types"
)

func TestDispatchHonorsNotificationAndSoundSettingsIndependently(t *testing.T) {
	cases := []struct {
		name              string
		notifications     bool
		sound             bool
		wantNotifications int
		wantSounds        int
	}{
		{name: "both disabled", notifications: false, sound: false, wantNotifications: 0, wantSounds: 0},
		{name: "notification only", notifications: true, sound: false, wantNotifications: 1, wantSounds: 0},
		{name: "sound only", notifications: false, sound: true, wantNotifications: 0, wantSounds: 1},
		{name: "both enabled", notifications: true, sound: true, wantNotifications: 1, wantSounds: 1},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			coreCtx := testCoreContext(t, ctx)
			if err := coreCtx.CoreSettings.SetSetting(ctx, coreCtx.UserID, sharedtypes.CoreSettingsKeyBoundaryNotifications, tc.notifications); err != nil {
				t.Fatalf("set boundary notifications: %v", err)
			}
			if err := coreCtx.CoreSettings.SetSetting(ctx, coreCtx.UserID, sharedtypes.CoreSettingsKeyBoundarySound, tc.sound); err != nil {
				t.Fatalf("set boundary sound: %v", err)
			}

			notificationCalls := 0
			soundCalls := 0
			alertStatusFn = func(paths runtimepkg.Paths) sharedtypes.AlertStatus {
				return sharedtypes.AlertStatus{
					NotificationsAvailable: true,
					SoundAvailable:         true,
				}
			}
			alertSoundPathFn = func(paths runtimepkg.Paths, preset sharedtypes.AlertSoundPreset) (string, error) {
				return filepath.Join(t.TempDir(), "sound.wav"), nil
			}
			sendAlertNotificationFn = func(status sharedtypes.AlertStatus, req sharedtypes.AlertRequest) error {
				notificationCalls++
				return nil
			}
			playAlertSoundFn = func(status sharedtypes.AlertStatus, soundPath string) error {
				soundCalls++
				return nil
			}
			defer func() {
				alertStatusFn = detectAlertStatus
				alertSoundPathFn = alertSoundPath
				sendAlertNotificationFn = sendAlertNotification
				playAlertSoundFn = playAlertSound
			}()

			service := &Service{
				core:   coreCtx,
				logger: testLogger(t),
			}
			if err := service.deliver(ctx, sharedtypes.AlertRequest{
				Kind:      sharedtypes.AlertEventTimerBreakComplete,
				Title:     "Break done",
				Body:      "Back to work",
				PlaySound: true,
			}, true); err != nil {
				t.Fatalf("deliver alert: %v", err)
			}

			if notificationCalls != tc.wantNotifications {
				t.Fatalf("expected %d notification calls, got %d", tc.wantNotifications, notificationCalls)
			}
			if soundCalls != tc.wantSounds {
				t.Fatalf("expected %d sound calls, got %d", tc.wantSounds, soundCalls)
			}
		})
	}
}

func TestInactivityAlertFiresAfterThresholdAndRepeats(t *testing.T) {
	ctx := context.Background()
	currentNow := "2026-04-11T10:00:00Z"
	coreCtx := testFullCoreContext(t, func() string { return currentNow })
	mustStartInactivitySession(t, ctx, coreCtx)

	var delivered []sharedtypes.AlertRequest
	restore := stubAlertDelivery(t, &delivered)
	defer restore()

	service := &Service{core: coreCtx, logger: testLogger(t)}
	service.processInactivityTick(ctx, mustTime(t, "2026-04-11T10:59:00Z"))
	if len(delivered) != 0 {
		t.Fatalf("expected no alert before threshold, got %d", len(delivered))
	}
	service.processInactivityTick(ctx, mustTime(t, "2026-04-11T11:00:00Z"))
	if len(delivered) != 1 || delivered[0].Kind != sharedtypes.AlertEventFocusInactivity {
		t.Fatalf("expected first inactivity alert, got %+v", delivered)
	}
	service.processInactivityTick(ctx, mustTime(t, "2026-04-11T11:30:00Z"))
	if len(delivered) != 1 {
		t.Fatalf("expected repeat suppression before interval, got %d", len(delivered))
	}
	service.processInactivityTick(ctx, mustTime(t, "2026-04-11T12:00:00Z"))
	if len(delivered) != 2 {
		t.Fatalf("expected hourly repeat, got %d", len(delivered))
	}
}

func TestTimerActivityTouchPostponesInactivityAlert(t *testing.T) {
	ctx := context.Background()
	currentNow := "2026-04-11T10:00:00Z"
	coreCtx := testFullCoreContext(t, func() string { return currentNow })
	mustStartInactivitySession(t, ctx, coreCtx)

	var delivered []sharedtypes.AlertRequest
	restore := stubAlertDelivery(t, &delivered)
	defer restore()

	service := &Service{core: coreCtx, logger: testLogger(t)}
	alertNowFn = func() time.Time { return mustTime(t, "2026-04-11T10:50:00Z") }
	if err := service.TouchTimerActivity(ctx); err != nil {
		t.Fatalf("touch timer activity: %v", err)
	}
	service.processInactivityTick(ctx, mustTime(t, "2026-04-11T11:00:00Z"))
	if len(delivered) != 0 {
		t.Fatalf("expected touched session to suppress original threshold, got %d", len(delivered))
	}
	service.processInactivityTick(ctx, mustTime(t, "2026-04-11T11:50:00Z"))
	if len(delivered) != 1 {
		t.Fatalf("expected alert after touched threshold, got %d", len(delivered))
	}
}

func TestInactivityAlertRespectsDisabledAndPausedStates(t *testing.T) {
	ctx := context.Background()
	currentNow := "2026-04-11T10:00:00Z"
	coreCtx := testFullCoreContext(t, func() string { return currentNow })
	mustStartInactivitySession(t, ctx, coreCtx)
	if err := coreCtx.CoreSettings.SetSetting(ctx, coreCtx.UserID, sharedtypes.CoreSettingsKeyInactivityAlerts, false); err != nil {
		t.Fatalf("disable inactivity alerts: %v", err)
	}

	var delivered []sharedtypes.AlertRequest
	restore := stubAlertDelivery(t, &delivered)
	defer restore()

	service := &Service{core: coreCtx, logger: testLogger(t)}
	service.processInactivityTick(ctx, mustTime(t, "2026-04-11T11:30:00Z"))
	if len(delivered) != 0 {
		t.Fatalf("expected disabled setting to suppress alert, got %d", len(delivered))
	}
	if err := coreCtx.CoreSettings.SetSetting(ctx, coreCtx.UserID, sharedtypes.CoreSettingsKeyInactivityAlerts, true); err != nil {
		t.Fatalf("enable inactivity alerts: %v", err)
	}
	if err := corecommands.PauseSession(ctx, coreCtx, sharedtypes.SessionSegmentRest); err != nil {
		t.Fatalf("pause session: %v", err)
	}
	service.processInactivityTick(ctx, mustTime(t, "2026-04-11T12:30:00Z"))
	if len(delivered) != 0 {
		t.Fatalf("expected paused session to suppress alert, got %d", len(delivered))
	}
}

func testCoreContext(t *testing.T, ctx context.Context) *core.Context {
	t.Helper()

	store, err := dbpkg.Open(filepath.Join(t.TempDir(), "crona.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	if err := migrations.InitSchema(ctx, store.DB()); err != nil {
		t.Fatalf("init schema: %v", err)
	}

	repo := repositories.NewCoreSettingsRepository(store.DB())
	if err := repo.InitializeDefaults(ctx, "local", "device-1"); err != nil {
		t.Fatalf("initialize defaults: %v", err)
	}

	return &core.Context{
		CoreSettings: repo,
		UserID:       "local",
		DeviceID:     "device-1",
	}
}

func testFullCoreContext(t *testing.T, now func() string) *core.Context {
	t.Helper()

	db, err := store.Open(filepath.Join(t.TempDir(), "crona.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := store.InitSchema(context.Background(), db.DB()); err != nil {
		t.Fatalf("init schema: %v", err)
	}
	coreCtx := core.NewContext(db, store.NewRegistry(db.DB()), "local", "device-1", t.TempDir(), now, events.NewBus())
	if err := coreCtx.InitDefaults(context.Background()); err != nil {
		t.Fatalf("init defaults: %v", err)
	}
	return coreCtx
}

func mustStartInactivitySession(t *testing.T, ctx context.Context, coreCtx *core.Context) {
	t.Helper()

	repo, err := coreCtx.Repos.Create(ctx, sharedtypes.Repo{ID: 1, Name: "Work"}, coreCtx.UserID, coreCtx.Now())
	if err != nil {
		t.Fatalf("create repo: %v", err)
	}
	stream, err := coreCtx.Streams.Create(ctx, sharedtypes.Stream{ID: 1, RepoID: repo.ID, Name: "App", Visibility: sharedtypes.StreamVisibilityPersonal}, coreCtx.UserID, coreCtx.Now())
	if err != nil {
		t.Fatalf("create stream: %v", err)
	}
	issue, err := corecommands.CreateIssue(ctx, coreCtx, struct {
		StreamID        int64
		Title           string
		Description     *string
		EstimateMinutes *int
		Notes           *string
		TodoForDate     *string
	}{StreamID: stream.ID, Title: "Long focus block"})
	if err != nil {
		t.Fatalf("create issue: %v", err)
	}
	if _, err := corecommands.StartSession(ctx, coreCtx, issue.ID); err != nil {
		t.Fatalf("start session: %v", err)
	}
}

func stubAlertDelivery(t *testing.T, delivered *[]sharedtypes.AlertRequest) func() {
	t.Helper()

	alertStatusFn = func(paths runtimepkg.Paths) sharedtypes.AlertStatus {
		return sharedtypes.AlertStatus{NotificationsAvailable: true}
	}
	sendAlertNotificationFn = func(status sharedtypes.AlertStatus, req sharedtypes.AlertRequest) error {
		*delivered = append(*delivered, req)
		return nil
	}
	return func() {
		alertStatusFn = detectAlertStatus
		alertNowFn = time.Now
		sendAlertNotificationFn = sendAlertNotification
	}
}

func mustTime(t *testing.T, value string) time.Time {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		t.Fatalf("parse time: %v", err)
	}
	return parsed
}

func testLogger(t *testing.T) *runtimepkg.Logger {
	t.Helper()

	logDir := filepath.Join(t.TempDir(), "logs")
	paths := runtimepkg.Paths{CurrentLogDir: logDir}
	if err := runtimepkg.EnsurePaths(paths); err != nil {
		t.Fatalf("ensure paths: %v", err)
	}
	return runtimepkg.NewLogger(paths)
}
