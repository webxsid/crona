package commands

import (
	"testing"

	sharedtypes "crona/shared/types"
)

func TestFocusScoreReason(t *testing.T) {
	tests := []struct {
		name           string
		worked, target int
		restRatio      float64
		level          sharedtypes.FocusScoreLevel
		want           sharedtypes.FocusScoreReason
	}{
		{"no activity", 0, 3600, 0, sharedtypes.FocusScoreLevelLow, sharedtypes.FocusScoreReasonNoActivity},
		{"under target", 1800, 3600, 0.25, sharedtypes.FocusScoreLevelLow, sharedtypes.FocusScoreReasonUnderTarget},
		{"needs breaks", 3600, 3600, 0.05, sharedtypes.FocusScoreLevelSteady, sharedtypes.FocusScoreReasonNeedsBreaks},
		{"balanced", 3600, 3600, 0.25, sharedtypes.FocusScoreLevelStrong, sharedtypes.FocusScoreReasonBalanced},
		{"overextended takes precedence", 5000, 3600, 0.05, sharedtypes.FocusScoreLevelOverextended, sharedtypes.FocusScoreReasonOverextended},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := focusScoreReason(tt.worked, tt.target, tt.restRatio, tt.level); got != tt.want {
				t.Fatalf("focusScoreReason() = %q, want %q", got, tt.want)
			}
		})
	}
}
