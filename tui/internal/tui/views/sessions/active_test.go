package sessions

import (
	"strings"
	"testing"
	"time"

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
			TimerMode:           sharedtypes.TimerModeStopwatch,
			BreaksEnabled:       true,
			WorkDurationMinutes: 25,
			ShortBreakMinutes:   5,
			LongBreakMinutes:    15,
		},
	}
	sessionStart := time.Now().UTC().Add(-2 * time.Hour).Format(time.RFC3339)
	segmentStart := time.Now().UTC().Add(-12*time.Minute - 34*time.Second).Format(time.RFC3339)
	state.Timer.SessionStartTime = &sessionStart
	state.Timer.SegmentStartTime = &segmentStart

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
	if strings.Contains(rendered, "mins until break") || strings.Contains(rendered, "mins left") {
		t.Fatalf(
			"expected stopwatch active view to avoid session timing labels, got %q",
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
	if strings.Contains(rendered, "min left") || strings.Contains(rendered, "until break") {
		t.Fatalf("expected stopwatch break state to avoid timing labels, got %q", rendered)
	}

	pausedState := state
	pausedState.Timer.State = "paused"
	pausedState.Timer.SegmentType = segmentPtr(sharedtypes.SessionSegmentRest)
	pausedState.Timer.ElapsedSeconds = 754
	pausedNow := time.Now().UTC().Format(time.RFC3339)
	pausedState.Timer.SegmentStartTime = &pausedNow
	rendered = renderActiveView(types.Theme{}, pausedState)
	if !strings.Contains(rendered, "Paused For") {
		t.Fatalf("expected paused stopwatch title, got %q", rendered)
	}
	if strings.Contains(rendered, "00:12:34") || strings.Contains(rendered, "12:34") ||
		strings.Contains(rendered, "00:00:12") {
		t.Fatalf("expected paused stopwatch to ignore prior elapsed time, got %q", rendered)
	}

	readyState := state
	readyState.Timer.State = "ready"
	readyState.Timer.SegmentType = nil
	readyState.Timer.ReadySegmentType = segmentPtr(sharedtypes.SessionSegmentShortBreak)
	readyState.Timer.NextSegmentType = segmentPtr(sharedtypes.SessionSegmentShortBreak)
	rendered = renderActiveView(types.Theme{}, readyState)
	if strings.Contains(rendered, "mins until break") || strings.Contains(rendered, "mins left") {
		t.Fatalf("expected stopwatch ready state to avoid session timing labels, got %q", rendered)
	}

	hardLimitState := state
	hardLimitState.Timer.State = "running"
	hardLimitState.Timer.SegmentType = segmentPtr(sharedtypes.SessionSegmentWork)
	hardLimitState.Timer.ReadySegmentType = nil
	hardLimitState.Timer.NextSegmentType = segmentPtr(sharedtypes.SessionSegmentShortBreak)
	hardLimitState.Timer.HardLimitActive = true
	hardLimitState.Timer.HardLimitTotalSeconds = 5400
	hardLimitState.Timer.HardLimitWorkSeconds = 1200
	hardLimitState.Timer.HardLimitBreakSeconds = 420
	staleRemaining := 1800
	sessionStart = time.Now().UTC().Add(-85 * time.Minute).Format(time.RFC3339)
	segmentStart = time.Now().UTC().Add(-12 * time.Minute).Format(time.RFC3339)
	hardLimitState.Timer.HardLimitRemainingSeconds = staleRemaining
	hardLimitState.Timer.SessionStartTime = &sessionStart
	hardLimitState.Timer.SegmentStartTime = &segmentStart
	rendered = renderActiveView(types.Theme{}, hardLimitState)
	if !strings.Contains(rendered, "Pomodoro Session") {
		t.Fatalf("expected hard-limit title, got %q", rendered)
	}
	if strings.Contains(rendered, "until long break") {
		t.Fatalf("expected hard-limit work label to avoid long-break wording, got %q", rendered)
	}
	if strings.Contains(rendered, "left in cap") || strings.Contains(rendered, "Cap:") {
		t.Fatalf("expected hard-limit view to avoid cap countdown text, got %q", rendered)
	}
	if !strings.Contains(rendered, "[x] commit issue  [z] stash session  [i] change context") {
		t.Fatalf("expected hard-limit action hints, got %q", rendered)
	}
	if strings.Contains(rendered, "Ready For") {
		t.Fatalf("expected hard-limit view to avoid ready-for semantics, got %q", rendered)
	}

	pomodoroState := hardLimitState
	pomodoroState.Timer.HardLimitWorkSeconds = 120
	pomodoroState.Timer.HardLimitBreakSeconds = 60
	pomodoroState.Timer.HardLimitLongBreakSeconds = 120
	pomodoroState.Timer.SegmentType = segmentPtr(sharedtypes.SessionSegmentWork)
	pomodoroState.Timer.ReadySegmentType = nil
	pomodoroState.Timer.NextSegmentType = segmentPtr(sharedtypes.SessionSegmentShortBreak)
	workStart := time.Now().UTC().Format(time.RFC3339)
	pomodoroState.Timer.SegmentStartTime = &workStart
	if got := sessionTimingLabel(pomodoroState, time.Now().UTC()); got != "2 mins until break" {
		t.Fatalf("expected hard-limit work label to use the selected focus duration, got %q", got)
	}

	hardLimitReadyState := hardLimitState
	hardLimitReadyState.Timer.State = "ready"
	hardLimitReadyState.Timer.SegmentType = nil
	hardLimitReadyState.Timer.ReadySegmentType = segmentPtr(sharedtypes.SessionSegmentWork)
	rendered = renderActiveView(types.Theme{}, hardLimitReadyState)
	if strings.Contains(rendered, "Ready For") {
		t.Fatalf("expected hard-limit ready state to avoid ready-for title, got %q", rendered)
	}
	if !strings.Contains(rendered, "Pomodoro Session") {
		t.Fatalf("expected hard-limit ready state to keep the hard-limit title, got %q", rendered)
	}

	pomodoroReadyState := pomodoroState
	pomodoroReadyState.Timer.State = "ready"
	pomodoroReadyState.Timer.SegmentType = nil
	pomodoroReadyState.Timer.ReadySegmentType = segmentPtr(sharedtypes.SessionSegmentWork)
	if got := sessionTimingLabel(pomodoroReadyState, time.Now().UTC()); got != "2 mins until break" {
		t.Fatalf("expected hard-limit ready label to use the selected focus duration, got %q", got)
	}

	hardLimitState.Timer.HardLimitExpired = true
	rendered = renderActiveView(types.Theme{}, hardLimitState)
	if !strings.Contains(rendered, "commit, stash, or extend") {
		t.Fatalf("expected expired hard-limit prompt, got %q", rendered)
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
