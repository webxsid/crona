package sessions

import (
	"strings"
	"testing"
	"time"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/x/ansi"
)

func layoutState(hardLimit bool) types.ContentState {
	issueID, est := int64(7), 35
	desc, notes := "Add shell completion.", "[2026-10-14T09:00:00Z] planned -> ready"
	start := time.Now().UTC().Add(-42 * time.Second).Format(time.RFC3339)
	state := types.ContentState{
		View: "session_active", Pane: "session_active", Width: 160, Height: 44,
		Timer: &api.TimerState{
			State: "running", IssueID: &issueID, SegmentStartTime: &start, SessionStartTime: &start,
			SegmentType: new(
				sharedtypes.SessionSegmentWork,
			), NextSegmentType: new(sharedtypes.SessionSegmentShortBreak),
		},
		AllIssues: []api.IssueWithMeta{{
			Issue: api.Issue{
				ID:              issueID,
				Title:           "Improve install docs",
				EstimateMinutes: &est,
				Description:     &desc,
				Notes:           &notes,
			},
			RepoName: "OSS", StreamName: "cli",
		}},
		Settings: &api.CoreSettings{TimerMode: sharedtypes.TimerModeStopwatch},
	}
	if hardLimit {
		state.Timer.HardLimitActive = true
		state.Timer.HardLimitWorkSeconds = 60
		state.Timer.HardLimitBreakSeconds = 300
		state.Timer.HardLimitTotalSeconds = 3600
		state.Timer.HardLimitRemainingSeconds = 3000
	}
	return state
}

func TestActiveSessionIsOnePanelWithTheIssueUnderTheClock(t *testing.T) {
	rendered := ansi.Strip(renderActiveView(types.Theme{}, layoutState(true)))
	if strings.Contains(rendered, "Active Issue") || strings.Contains(rendered, "Notes") ||
		strings.Contains(rendered, "planned -> ready") {
		t.Fatalf("expected no separate issue pane or notes:\n%s", rendered)
	}
	clock, issue := strings.Index(rendered, "█"), strings.Index(rendered, "Improve install docs")
	if clock < 0 || issue < clock {
		t.Fatalf("expected the issue under the clock:\n%s", rendered)
	}
	if !strings.Contains(rendered, "OSS › cli") {
		t.Fatalf("expected repo › stream in the header:\n%s", rendered)
	}
}

func TestActiveSessionTitleNamesKindAndPhase(t *testing.T) {
	rendered := ansi.Strip(renderActiveView(types.Theme{}, layoutState(true)))
	for _, want := range []string{"Pomodoro · Focus", "next short break", "left"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("missing %q:\n%s", want, rendered)
		}
	}
	for _, gone := range []string{"WORK", "Completed sessions"} {
		if strings.Contains(rendered, gone) {
			t.Fatalf("unexpected %q:\n%s", gone, rendered)
		}
	}
}

func TestActiveSessionShowsWorkedAgainstEstimate(t *testing.T) {
	for _, hardLimit := range []bool{true, false} {
		rendered := ansi.Strip(renderActiveView(types.Theme{}, layoutState(hardLimit)))
		if !strings.Contains(rendered, "Worked") || !strings.Contains(rendered, "est. 35m") {
			t.Fatalf("hardLimit=%v: missing worked / estimate:\n%s", hardLimit, rendered)
		}
	}
}

func TestStopwatchSessionHasNoProgressBar(t *testing.T) {
	rendered := ansi.Strip(renderActiveView(types.Theme{}, layoutState(false)))
	if strings.Contains(rendered, "░") {
		t.Fatalf("expected no progress bar on a stopwatch session:\n%s", rendered)
	}
	if !strings.Contains(ansi.Strip(renderActiveView(types.Theme{}, layoutState(true))), "░") {
		t.Fatal("expected a progress bar on a timed session")
	}
}

func TestActiveSessionCentresItsContentVertically(t *testing.T) {
	lines := strings.Split(ansi.Strip(renderActiveView(types.Theme{}, layoutState(true))), "\n")
	first, last := -1, -1
	for i, line := range lines[1:] {
		if strings.TrimSpace(line) != "" {
			if first < 0 {
				first = i
			}
			last = i
		}
	}
	above, below := first, len(lines)-2-last
	if above < 3 || above-below > 2 || below-above > 2 {
		t.Fatalf(
			"expected centred content, %d blank above and %d below:\n%s",
			above,
			below,
			strings.Join(lines, "\n"),
		)
	}
}

func TestPomodoroShowsCycleProgress(t *testing.T) {
	state := layoutState(true)
	state.Timer.HardLimitKind = sharedtypes.TimerHardLimitKindPomodoro
	state.Timer.HardLimitCycles, state.Timer.CompletedWorkCycles = 4, 1
	state.Timer.IssueWorkedSeconds = 1200
	rendered := ansi.Strip(renderActiveView(types.Theme{}, state))
	for _, want := range []string{"Pomodoro · Focus 2 of 4", "● ◉ ○ ○", "Worked 20m / est. 35m"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("missing %q:\n%s", want, rendered)
		}
	}
}
