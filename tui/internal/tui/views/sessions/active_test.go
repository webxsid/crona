package sessions

import (
	"strings"
	"testing"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
)

func TestRenderActiveViewUsesResponsiveClockVariants(t *testing.T) {
	issueID := int64(7)
	state := types.ContentState{
		View:   "session_active",
		Pane:   "session_active",
		Width:  96,
		Height: 36,
		Timer: &api.TimerState{
			State:           "running",
			IssueID:         &issueID,
			SegmentType:     segmentPtr(sharedtypes.SessionSegmentWork),
			NextSegmentType: segmentPtr(sharedtypes.SessionSegmentShortBreak),
			ElapsedSeconds:  754,
		},
		AllIssues: []api.IssueWithMeta{{
			Issue: api.Issue{ID: issueID, Title: "Timer sizing"},
		}},
		Settings: &api.CoreSettings{
			TimerMode:           sharedtypes.TimerModeStructured,
			BreaksEnabled:       true,
			WorkDurationMinutes: 25,
			ShortBreakMinutes:   5,
			LongBreakMinutes:    15,
		},
	}

	rendered := renderActiveView(types.Theme{}, state)
	titleAt := strings.Index(rendered, "Focus Session")
	clockAt := strings.Index(rendered, "█")
	metricAt := strings.Index(rendered, "WORK")
	issueAt := strings.Index(rendered, "Active Issue")
	if titleAt < 0 || clockAt < 0 || metricAt < 0 || issueAt < 0 {
		t.Fatalf("expected title, clock, metadata, and issue pane to render, got %q", rendered)
	}
	if titleAt >= clockAt || clockAt >= metricAt || metricAt >= issueAt {
		t.Fatalf("expected title above clock above metadata above issue pane, got %q", rendered)
	}
	if !strings.Contains(rendered, strings.Repeat("█", 12)) {
		t.Fatalf("expected structured active view to render a wide progress bar, got %q", rendered)
	}
	if !strings.Contains(rendered, "mins until break") {
		t.Fatalf(
			"expected structured active view to show the next break indicator, got %q",
			rendered,
		)
	}

	narrow := state
	narrow.Width = 20
	narrow.Height = 14
	rendered = renderActiveView(types.Theme{}, narrow)
	if !strings.Contains(rendered, "00:12:34") {
		t.Fatalf("expected narrow active view to fall back to plain clock text, got %q", rendered)
	}
	if strings.Contains(rendered, "██ ██") {
		t.Fatalf("expected narrow active view to avoid the large glyph clock, got %q", rendered)
	}
	if !strings.Contains(rendered, "Focus") || !strings.Contains(rendered, "Session") ||
		!strings.Contains(rendered, "WORK") ||
		!strings.Contains(rendered, "Active Issue") {
		t.Fatalf(
			"expected narrow active view to keep the title, metadata, and issue pane visible, got %q",
			rendered,
		)
	}

	breakState := state
	breakState.Timer.SegmentType = segmentPtr(sharedtypes.SessionSegmentShortBreak)
	breakState.Timer.NextSegmentType = segmentPtr(sharedtypes.SessionSegmentWork)
	rendered = renderActiveView(types.Theme{}, breakState)
	if !strings.Contains(rendered, "min left") {
		t.Fatalf("expected active break to show the remaining time indicator, got %q", rendered)
	}

	readyState := state
	readyState.Timer.State = "ready"
	readyState.Timer.SegmentType = nil
	readyState.Timer.ReadySegmentType = segmentPtr(sharedtypes.SessionSegmentShortBreak)
	readyState.Timer.NextSegmentType = segmentPtr(sharedtypes.SessionSegmentShortBreak)
	rendered = renderActiveView(types.Theme{}, readyState)
	if !strings.Contains(rendered, "break ready") {
		t.Fatalf("expected ready state to show the prepared segment indicator, got %q", rendered)
	}
}

func TestActiveTimerColorUsesSegmentType(t *testing.T) {
	theme := types.Theme{
		ColorGreen:   lipgloss.Color("2"),
		ColorCyan:    lipgloss.Color("6"),
		ColorMagenta: lipgloss.Color("5"),
		ColorYellow:  lipgloss.Color("3"),
	}
	if got := activeTimerColor(theme, &api.TimerState{State: "running", SegmentType: segmentPtr(sharedtypes.SessionSegmentWork)}); got != theme.ColorGreen {
		t.Fatalf("expected work to use green, got %v", got)
	}
	if got := activeTimerColor(theme, &api.TimerState{State: "paused", SegmentType: segmentPtr(sharedtypes.SessionSegmentShortBreak)}); got != theme.ColorCyan {
		t.Fatalf("expected short break to use cyan, got %v", got)
	}
	if got := activeTimerColor(theme, &api.TimerState{State: "paused", SegmentType: segmentPtr(sharedtypes.SessionSegmentLongBreak)}); got != theme.ColorMagenta {
		t.Fatalf("expected long break to use magenta, got %v", got)
	}
	if got := activeTimerColor(theme, &api.TimerState{State: "ready"}); got != theme.ColorYellow {
		t.Fatalf("expected ready state to use yellow, got %v", got)
	}
}

func segmentPtr(value sharedtypes.SessionSegmentType) *sharedtypes.SessionSegmentType {
	return &value
}
