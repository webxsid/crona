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
	if settings.DateDisplayPreset != sharedtypes.DateDisplayPresetISO {
		t.Fatalf("expected default date display preset iso, got %q", settings.DateDisplayPreset)
	}
	if settings.PromptGlyphMode != sharedtypes.PromptGlyphModeEmoji {
		t.Fatalf("expected default prompt glyph mode emoji, got %q", settings.PromptGlyphMode)
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
	if err := repo.SetSetting(ctx, "local", sharedtypes.CoreSettingsKeyInactivityAlerts, false); err != nil {
		t.Fatalf("set inactivity alerts: %v", err)
	}
	if err := repo.SetSetting(ctx, "local", sharedtypes.CoreSettingsKeyInactivityThreshold, 45); err != nil {
		t.Fatalf("set inactivity threshold: %v", err)
	}
	if err := repo.SetSetting(ctx, "local", sharedtypes.CoreSettingsKeyInactivityRepeat, 90); err != nil {
		t.Fatalf("set inactivity repeat: %v", err)
	}
	if err := repo.SetSetting(ctx, "local", sharedtypes.CoreSettingsKeyDateDisplayPreset, string(sharedtypes.DateDisplayPresetCustom)); err != nil {
		t.Fatalf("set date display preset: %v", err)
	}
	if err := repo.SetSetting(ctx, "local", sharedtypes.CoreSettingsKeyDateDisplayFormat, "Do MMM YYYY"); err != nil {
		t.Fatalf("set date display format: %v", err)
	}
	if err := repo.SetSetting(ctx, "local", sharedtypes.CoreSettingsKeyPromptGlyphMode, string(sharedtypes.PromptGlyphModeASCII)); err != nil {
		t.Fatalf("set prompt glyph mode: %v", err)
	}
	defs := []sharedtypes.HabitStreakDefinition{{
		ID:            "health",
		Name:          "Health streak",
		Enabled:       true,
		Period:        sharedtypes.HabitStreakPeriodWeek,
		RequiredCount: 2,
		HabitIDs:      []int64{11, 12},
	}}
	if err := repo.SetSetting(ctx, "local", sharedtypes.CoreSettingsKeyHabitStreakDefs, defs); err != nil {
		t.Fatalf("set habit streak definitions: %v", err)
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
	if settings.InactivityAlerts {
		t.Fatalf("expected inactivity alerts disabled")
	}
	if settings.InactivityThreshold != 45 || settings.InactivityRepeat != 90 {
		t.Fatalf("unexpected inactivity settings: threshold=%d repeat=%d", settings.InactivityThreshold, settings.InactivityRepeat)
	}
	if settings.DateDisplayPreset != sharedtypes.DateDisplayPresetCustom {
		t.Fatalf("expected custom date display preset, got %q", settings.DateDisplayPreset)
	}
	if settings.DateDisplayFormat != "Do MMM YYYY" {
		t.Fatalf("unexpected date display format: %q", settings.DateDisplayFormat)
	}
	if settings.PromptGlyphMode != sharedtypes.PromptGlyphModeASCII {
		t.Fatalf("unexpected prompt glyph mode: %q", settings.PromptGlyphMode)
	}
	if len(settings.HabitStreakDefs) != 1 {
		t.Fatalf("expected one custom habit streak, got %+v", settings.HabitStreakDefs)
	}
	if settings.HabitStreakDefs[0].Name != "Health streak" || settings.HabitStreakDefs[0].Period != sharedtypes.HabitStreakPeriodWeek || settings.HabitStreakDefs[0].RequiredCount != 2 {
		t.Fatalf("unexpected habit streak definition: %+v", settings.HabitStreakDefs[0])
	}
}

func TestHabitStreakDailyPeriodNormalizesCountToOne(t *testing.T) {
	ctx := context.Background()
	store := openTestStore(t)

	repo := repositories.NewCoreSettingsRepository(store.DB())
	if err := repo.InitializeDefaults(ctx, "local", "device-1"); err != nil {
		t.Fatalf("initialize defaults: %v", err)
	}
	defs := []sharedtypes.HabitStreakDefinition{{
		ID:            "daily-health",
		Name:          "Daily health",
		Enabled:       true,
		Period:        sharedtypes.HabitStreakPeriodDay,
		RequiredCount: 4,
		HabitIDs:      []int64{11},
	}}
	if err := repo.SetSetting(ctx, "local", sharedtypes.CoreSettingsKeyHabitStreakDefs, defs); err != nil {
		t.Fatalf("set habit streak definitions: %v", err)
	}

	settings, err := repo.Get(ctx, "local")
	if err != nil {
		t.Fatalf("get settings: %v", err)
	}
	if len(settings.HabitStreakDefs) != 1 {
		t.Fatalf("expected one habit streak, got %+v", settings.HabitStreakDefs)
	}
	if settings.HabitStreakDefs[0].RequiredCount != 1 {
		t.Fatalf("expected daily streak count normalized to 1, got %+v", settings.HabitStreakDefs[0])
	}
}
