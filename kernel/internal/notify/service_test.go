package notify

import (
	"context"
	"path/filepath"
	"testing"

	"crona/kernel/internal/core"
	runtimepkg "crona/kernel/internal/runtime"
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

func testLogger(t *testing.T) *runtimepkg.Logger {
	t.Helper()

	logDir := filepath.Join(t.TempDir(), "logs")
	paths := runtimepkg.Paths{CurrentLogDir: logDir}
	if err := runtimepkg.EnsurePaths(paths); err != nil {
		t.Fatalf("ensure paths: %v", err)
	}
	return runtimepkg.NewLogger(paths)
}
