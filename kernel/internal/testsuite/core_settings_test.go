package testsuite

import (
	"context"
	"testing"

	"crona/kernel/internal/store/repositories"
	sharedtypes "crona/shared/types"
)

func TestGetAllSettingsReturnsPublicCoreSettingsShape(t *testing.T) {
	ctx := context.Background()
	store := openTestStore(t)

	repo := repositories.NewCoreSettingsRepository(store.DB())
	if err := repo.InitializeDefaults(ctx, "local", "device-1"); err != nil {
		t.Fatalf("initialize defaults: %v", err)
	}
	if err := repo.SetSetting(ctx, "local", sharedtypes.CoreSettingsKeyBoundaryNotifications, false); err != nil {
		t.Fatalf("set boundary notifications: %v", err)
	}
	if err := repo.SetSetting(ctx, "local", sharedtypes.CoreSettingsKeyBoundarySound, false); err != nil {
		t.Fatalf("set boundary sound: %v", err)
	}

	all, err := repo.GetAllSettings(ctx)
	if err != nil {
		t.Fatalf("get all settings: %v", err)
	}

	raw, ok := all["local"]
	if !ok {
		t.Fatalf("expected local settings entry, got %+v", all)
	}
	settings, ok := raw.(sharedtypes.CoreSettings)
	if !ok {
		t.Fatalf("expected sharedtypes.CoreSettings, got %T", raw)
	}
	if settings.BoundaryNotifications {
		t.Fatalf("expected boundary notifications false, got true")
	}
	if settings.BoundarySound {
		t.Fatalf("expected boundary sound false, got true")
	}
}

func TestCoreSettingsRoundTripAwayModeFields(t *testing.T) {
	ctx := context.Background()
	store := openTestStore(t)

	repo := repositories.NewCoreSettingsRepository(store.DB())
	if err := repo.InitializeDefaults(ctx, "local", "device-1"); err != nil {
		t.Fatalf("initialize defaults: %v", err)
	}
	if err := repo.SetSetting(ctx, "local", sharedtypes.CoreSettingsKeyAwayModeEnabled, true); err != nil {
		t.Fatalf("set away mode: %v", err)
	}
	if err := repo.SetSetting(ctx, "local", sharedtypes.CoreSettingsKeyFrozenStreakKinds, []string{string(sharedtypes.StreakKindCheckInDays)}); err != nil {
		t.Fatalf("set frozen streak kinds: %v", err)
	}
	if err := repo.SetSetting(ctx, "local", sharedtypes.CoreSettingsKeyRestWeekdays, []int{0, 6}); err != nil {
		t.Fatalf("set rest weekdays: %v", err)
	}
	if err := repo.SetSetting(ctx, "local", sharedtypes.CoreSettingsKeyRestSpecificDates, []string{"2026-03-29"}); err != nil {
		t.Fatalf("set rest specific dates: %v", err)
	}
	if err := repo.SetSetting(ctx, "local", sharedtypes.CoreSettingsKeyDailyPlanRollbackMins, 12); err != nil {
		t.Fatalf("set rollback minutes: %v", err)
	}

	settings, err := repo.Get(ctx, "local")
	if err != nil {
		t.Fatalf("get settings: %v", err)
	}
	if settings == nil || !settings.AwayModeEnabled {
		t.Fatalf("expected away mode enabled, got %+v", settings)
	}
	if len(settings.FrozenStreakKinds) != 1 || settings.FrozenStreakKinds[0] != sharedtypes.StreakKindCheckInDays {
		t.Fatalf("unexpected frozen streak kinds: %+v", settings.FrozenStreakKinds)
	}
	if len(settings.RestWeekdays) != 2 || settings.RestWeekdays[0] != 0 || settings.RestWeekdays[1] != 6 {
		t.Fatalf("unexpected rest weekdays: %+v", settings.RestWeekdays)
	}
	if len(settings.RestSpecificDates) != 1 || settings.RestSpecificDates[0] != "2026-03-29" {
		t.Fatalf("unexpected rest specific dates: %+v", settings.RestSpecificDates)
	}
	if settings.DailyPlanRollbackMins != 12 {
		t.Fatalf("unexpected rollback minutes: %d", settings.DailyPlanRollbackMins)
	}
}
