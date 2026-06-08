package sessionnotes

import (
	"testing"

	sharedtypes "crona/shared/types"
)

func TestSegmentDurationsIncludeElapsedOffsets(t *testing.T) {
	offset := 12
	workEnd := "2026-06-08T10:01:00Z"
	breakEnd := "2026-06-08T10:01:30Z"
	segments := []sharedtypes.SessionSegment{
		{
			SegmentType:          sharedtypes.SessionSegmentWork,
			StartTime:            "2026-06-08T10:00:00Z",
			EndTime:              &workEnd,
			ElapsedOffsetSeconds: &offset,
		},
		{
			SegmentType: sharedtypes.SessionSegmentShortBreak,
			StartTime:   "2026-06-08T10:01:00Z",
			EndTime:     &breakEnd,
		},
	}

	if got := SegmentDurationSeconds(segments[0]); got != 72 {
		t.Fatalf("expected offset-bearing work segment to be 72s, got %d", got)
	}
	if got := TotalSegmentDurationSeconds(segments); got != 102 {
		t.Fatalf("expected total segment duration 102s, got %d", got)
	}

	summary := ComputeWorkSummary(segments)
	if summary.WorkSeconds != 72 {
		t.Fatalf("expected work summary to include offsets, got %d", summary.WorkSeconds)
	}
	if summary.RestSeconds != 30 {
		t.Fatalf("expected rest summary 30s, got %d", summary.RestSeconds)
	}
	if summary.TotalSeconds != 102 {
		t.Fatalf("expected summary total 102s, got %d", summary.TotalSeconds)
	}
}
