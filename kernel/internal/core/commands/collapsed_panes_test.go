package commands

import (
	"context"
	"slices"
	"testing"

	sharedtypes "crona.work/shared/types"
)

func TestCollapsedPanesSettingRoundTrips(t *testing.T) {
	ctx := context.Background()
	coreCtx, _, _ := newTimerTestContext(t, func() string { return "2026-05-24T10:00:00Z" })
	settings, err := coreCtx.CoreSettings.Get(ctx, coreCtx.UserID)
	if err != nil || len(settings.CollapsedPanes) != 0 {
		t.Fatalf("expected no collapsed panes by default, got %v (%v)", settings.CollapsedPanes, err)
	}
	want := []string{"daily/habits", "daily/detail"}
	if err := coreCtx.CoreSettings.SetSetting(ctx, coreCtx.UserID, sharedtypes.CoreSettingsKeyCollapsedPanes, want); err != nil {
		t.Fatalf("SetSetting: %v", err)
	}
	settings, err = coreCtx.CoreSettings.Get(ctx, coreCtx.UserID)
	if err != nil || !slices.Equal(settings.CollapsedPanes, want) {
		t.Fatalf("expected %v back, got %v (%v)", want, settings.CollapsedPanes, err)
	}
}
