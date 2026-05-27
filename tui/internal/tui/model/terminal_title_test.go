package model

import (
	"strings"
	"testing"
	"time"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
)

func TestTerminalTitleActiveSessionUsesIssueTimerAndState(t *testing.T) {
	issueID := int64(42)
	segment := sharedtypes.SessionSegmentWork
	segmentStart := time.Now().UTC().Add(-12 * time.Minute).Format(time.RFC3339)
	model := Model{
		timer: &api.TimerState{
			State:             "running",
			IssueID:           &issueID,
			SegmentType:       &segment,
			ElapsedSeconds:    35 * 60,
			SegmentStartTime:   &segmentStart,
		},
		elapsed: 7 * 60,
		allIssues: []api.IssueWithMeta{{
			Issue: sharedtypes.Issue{ID: issueID, Title: "Fix checkout title handling"},
		}},
	}

	if got, want := model.terminalTitle(), "Crona · Fix checkout title handling · 00:12:00 WORK"; got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestTerminalTitleActiveSessionFallsBackToIssueIDAndPausedState(t *testing.T) {
	issueID := int64(42)
	now := time.Now().UTC().Add(time.Hour).Format(time.RFC3339)
	segment := sharedtypes.SessionSegmentRest
	model := Model{
		timer: &api.TimerState{
			State:                       "paused",
			IssueID:                     &issueID,
			ElapsedSeconds:              65 * 60,
			SegmentStartTime:            &now,
			SegmentType:                 &segment,
			SegmentElapsedOffsetSeconds:  intPtr(0),
		},
	}

	if got, want := model.terminalTitle(), "Crona · Issue #42 · 00:00:00 PAUSED"; got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func intPtr(v int) *int { return &v }

func TestTerminalTitleIdleUsesContextAndView(t *testing.T) {
	repo := "Work"
	stream := "Kernel"
	model := Model{
		view: ViewDaily,
		context: &api.ActiveContext{
			RepoName:   &repo,
			StreamName: &stream,
		},
	}

	if got, want := model.terminalTitle(), "Crona · Work / Kernel · Daily"; got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestTerminalTitleIdleOmitsMissingContext(t *testing.T) {
	model := Model{view: ViewAlerts}

	if got, want := model.terminalTitle(), "Crona · Alerts"; got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestWithTerminalTitleEmitsOnlyWhenChanged(t *testing.T) {
	model := Model{view: ViewDaily, terminalTitleEnabled: true}
	model.lastTerminalTitle = model.terminalTitle()
	if _, cmd := model.withTerminalTitle(nil); cmd != nil {
		t.Fatalf("expected unchanged title to skip command")
	}

	model.view = ViewAlerts
	next, cmd := model.withTerminalTitle(nil)
	if cmd == nil {
		t.Fatalf("expected changed title to emit command")
	}
	if !strings.Contains(next.lastTerminalTitle, "Alerts") {
		t.Fatalf("expected stored title to update, got %q", next.lastTerminalTitle)
	}
}
