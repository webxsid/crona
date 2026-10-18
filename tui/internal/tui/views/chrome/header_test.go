package viewchrome

import (
	"strings"
	"testing"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"

	"github.com/charmbracelet/x/ansi"
)

func headerState(view string) HeaderState {
	issueID, est := int64(7), 35
	return HeaderState{
		Width: 160, View: view, Elapsed: 0,
		Timer: &api.TimerState{
			State: "running", IssueID: &issueID, SegmentType: new(sharedtypes.SessionSegmentWork),
			HardLimitActive: true, HardLimitKind: sharedtypes.TimerHardLimitKindPomodoro,
			HardLimitWorkSeconds: 1500, HardLimitBreakSeconds: 300, HardLimitTotalSeconds: 7200, HardLimitRemainingSeconds: 6000,
			HardLimitCycles: 4, CompletedWorkCycles: 1, IssueWorkedSeconds: 1200, IssueCompletedSessions: 2,
		},
		AllIssues: []api.IssueWithMeta{{Issue: api.Issue{ID: issueID, Title: "Improve install docs", EstimateMinutes: &est}}},
		Health:    &api.Health{OK: 1, DB: true},
	}
}

func TestStatusChipSummarisesTheSessionOffTheSessionView(t *testing.T) {
	line := ansi.Strip(HeaderSessionLine(Theme{}, headerState("daily")))
	for _, want := range []string{"FOCUS 2/4", "Improve install docs", "20m / est. 35m"} {
		if !strings.Contains(line, want) {
			t.Fatalf("missing %q in %q", want, line)
		}
	}
	for _, gone := range []string{"LIMIT", "sessions:", "engine", "status:"} {
		if strings.Contains(line, gone) {
			t.Fatalf("unexpected %q in %q", gone, line)
		}
	}
}

func TestStatusChipIsEmptyOnTheSessionView(t *testing.T) {
	if line := HeaderSessionLine(Theme{}, headerState("session_active")); line != "" {
		t.Fatalf("expected no chip on the session view, got %q", ansi.Strip(line))
	}
}

func TestStatusChipWarnsWhenTheEngineIsDegraded(t *testing.T) {
	state := headerState("daily")
	state.Health = &api.Health{OK: 0}
	if line := ansi.Strip(HeaderSessionLine(Theme{}, state)); !strings.Contains(line, "engine degraded") {
		t.Fatalf("expected an engine warning, got %q", line)
	}
}
