package version

import (
	"testing"

	sharedtypes "crona/shared/types"
)

func TestRunningChannelReflectsCurrentVersion(t *testing.T) {
	original := Version
	t.Cleanup(func() { Version = original })

	Version = "1.0.0-beta.1"
	if !IsBetaRelease() {
		t.Fatalf("expected beta release")
	}
	if got := RunningChannel(); got != sharedtypes.UpdateChannelBeta {
		t.Fatalf("expected beta channel, got %q", got)
	}

	Version = "1.0.0"
	if IsBetaRelease() {
		t.Fatalf("expected non-beta release")
	}
	if got := RunningChannel(); got != sharedtypes.UpdateChannelStable {
		t.Fatalf("expected stable channel, got %q", got)
	}
}
