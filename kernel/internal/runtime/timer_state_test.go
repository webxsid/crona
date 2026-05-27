package runtime

import (
	"os"
	"path/filepath"
	"testing"

	"crona/shared/config"
	sharedtypes "crona/shared/types"
)

func TestReadTimerRuntimeStateAcceptsLegacyPreparedSegmentJSON(t *testing.T) {
	base := t.TempDir()
	t.Setenv(config.EnvVarRuntimeDir, base)

	legacy := []byte(`{
		"sessionId": "session-123",
		"issueId": 42,
		"segmentType": "short_break",
		"recordedAt": "2026-05-24T10:00:00Z"
	}`)
	if err := os.WriteFile(filepath.Join(base, "timer.json"), legacy, 0o600); err != nil {
		t.Fatalf("write legacy timer state: %v", err)
	}

	state, err := ReadTimerRuntimeState()
	if err != nil {
		t.Fatalf("read legacy timer state: %v", err)
	}
	if state == nil {
		t.Fatal("expected legacy timer state to load")
	}
	if state.SessionID != "session-123" {
		t.Fatalf("expected session id session-123, got %q", state.SessionID)
	}
	if state.IssueID != 42 {
		t.Fatalf("expected issue id 42, got %d", state.IssueID)
	}
	if state.PreparedSegmentType == nil {
		t.Fatal("expected prepared segment type to be restored")
	}
	if *state.PreparedSegmentType != sharedtypes.SessionSegmentShortBreak {
		t.Fatalf("expected short break segment, got %q", *state.PreparedSegmentType)
	}
}
