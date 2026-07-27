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
	if state.HardLimitElapsedOffsetSeconds != 0 {
		t.Fatalf(
			"expected zero hard-limit offset by default, got %d",
			state.HardLimitElapsedOffsetSeconds,
		)
	}
}

func TestReadTimerRuntimeStateDefaultsLegacyHardLimitToPomodoro(t *testing.T) {
	base := t.TempDir()
	t.Setenv(config.EnvVarRuntimeDir, base)

	legacy := []byte(`{
		"sessionId": "session-123",
		"issueId": 42,
		"hardLimitTotalSeconds": 1500,
		"hardLimitWorkSeconds": 1500,
		"recordedAt": "2026-05-24T10:00:00Z"
	}`)
	if err := os.WriteFile(filepath.Join(base, "timer.json"), legacy, 0o600); err != nil {
		t.Fatalf("write legacy timer state: %v", err)
	}

	state, err := ReadTimerRuntimeState()
	if err != nil {
		t.Fatalf("read legacy hard-limit state: %v", err)
	}
	if state.HardLimitKind != sharedtypes.TimerHardLimitKindPomodoro {
		t.Fatalf("expected legacy hard limit to remain pomodoro, got %q", state.HardLimitKind)
	}
}

func TestTimerRuntimeStateRoundTripsCountdownKind(t *testing.T) {
	base := t.TempDir()
	t.Setenv(config.EnvVarRuntimeDir, base)

	state := NewHardLimitTimerRuntimeState(
		"session-123",
		42,
		sharedtypes.TimerHardLimitKindCountdown,
		1500,
		1500,
		0,
		0,
		0,
	)
	if err := WriteTimerRuntimeState(state); err != nil {
		t.Fatalf("write countdown timer state: %v", err)
	}
	loaded, err := ReadTimerRuntimeState()
	if err != nil {
		t.Fatalf("read countdown timer state: %v", err)
	}
	if loaded.HardLimitKind != sharedtypes.TimerHardLimitKindCountdown {
		t.Fatalf("expected countdown kind, got %q", loaded.HardLimitKind)
	}
}
