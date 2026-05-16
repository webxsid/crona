package sessions

import (
	"strings"
	"testing"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	types "crona/tui/internal/tui/views/types"
)

func TestRenderActiveViewUsesResponsiveClockVariants(t *testing.T) {
	issueID := int64(7)
	state := types.ContentState{
		View:   "session_active",
		Pane:   "session_active",
		Width:  96,
		Height: 36,
		Timer: &api.TimerState{
			State:          "running",
			IssueID:        &issueID,
			SegmentType:    segmentPtr(sharedtypes.SessionSegmentWork),
			ElapsedSeconds: 754,
		},
		AllIssues: []api.IssueWithMeta{{
			Issue: api.Issue{ID: issueID, Title: "Timer sizing"},
		}},
	}

	rendered := renderActiveView(types.Theme{}, state)
	titleAt := strings.Index(rendered, "Focus Session")
	clockAt := strings.Index(rendered, "█")
	metricAt := strings.Index(rendered, "WORK")
	issueAt := strings.Index(rendered, "Active Issue")
	if titleAt < 0 || clockAt < 0 || metricAt < 0 || issueAt < 0 {
		t.Fatalf("expected title, clock, metadata, and issue pane to render, got %q", rendered)
	}
	if !(titleAt < clockAt && clockAt < metricAt && metricAt < issueAt) {
		t.Fatalf("expected title above clock above metadata above issue pane, got %q", rendered)
	}

	narrow := state
	narrow.Width = 20
	narrow.Height = 14
	rendered = renderActiveView(types.Theme{}, narrow)
	if !strings.Contains(rendered, "00:12:34") {
		t.Fatalf("expected narrow active view to fall back to plain clock text, got %q", rendered)
	}
	if strings.Contains(rendered, "█") {
		t.Fatalf("expected narrow active view to avoid the large glyph clock, got %q", rendered)
	}
	if !strings.Contains(rendered, "Focus") || !strings.Contains(rendered, "Session") || !strings.Contains(rendered, "WORK") || !strings.Contains(rendered, "Active Issue") {
		t.Fatalf("expected narrow active view to keep the title, metadata, and issue pane visible, got %q", rendered)
	}
}

func segmentPtr(value sharedtypes.SessionSegmentType) *sharedtypes.SessionSegmentType {
	return &value
}
