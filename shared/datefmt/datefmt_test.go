package datefmt

import (
	"testing"
	"time"

	sharedtypes "crona/shared/types"
)

func TestFormatDateUsesPreset(t *testing.T) {
	settings := &sharedtypes.CoreSettings{DateDisplayPreset: sharedtypes.DateDisplayPresetUS}
	got := FormatISODate("2026-04-30", settings)
	if got != "04/30/2026" {
		t.Fatalf("expected US preset, got %q", got)
	}
}

func TestFormatDateSupportsMomentTokens(t *testing.T) {
	settings := &sharedtypes.CoreSettings{
		DateDisplayPreset: sharedtypes.DateDisplayPresetCustom,
		DateDisplayFormat: "Do MMM YYYY [Week] W",
	}
	got := FormatISODate("2026-04-30", settings)
	if got != "30th Apr 2026 Week 18" {
		t.Fatalf("expected custom pattern, got %q", got)
	}
}

func TestFormatDateFallsBackOnInvalidCustom(t *testing.T) {
	settings := &sharedtypes.CoreSettings{
		DateDisplayPreset: sharedtypes.DateDisplayPresetCustom,
		DateDisplayFormat: "[broken",
	}
	got := Preview(settings, time.Date(2026, time.April, 30, 0, 0, 0, 0, time.UTC))
	if got != "2026-04-30" {
		t.Fatalf("expected fallback ISO preview, got %q", got)
	}
}
