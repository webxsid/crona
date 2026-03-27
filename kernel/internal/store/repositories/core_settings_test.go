package repositories

import (
	"context"
	"path/filepath"
	"testing"

	dbpkg "crona/kernel/internal/store/db"
	"crona/kernel/internal/store/migrations"
	sharedtypes "crona/shared/types"
)

func TestGetAllSettingsReturnsPublicCoreSettingsShape(t *testing.T) {
	ctx := context.Background()
	store, err := dbpkg.Open(filepath.Join(t.TempDir(), "crona.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer func() { _ = store.Close() }()

	if err := migrations.InitSchema(ctx, store.DB()); err != nil {
		t.Fatalf("init schema: %v", err)
	}

	repo := NewCoreSettingsRepository(store.DB())
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
