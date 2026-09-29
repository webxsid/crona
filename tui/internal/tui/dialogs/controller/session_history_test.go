package controller

import (
	"testing"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"

	tea "github.com/charmbracelet/bubbletea"
)

func TestEditSessionHistoryOpensStructuredFieldsAndEditsSegments(t *testing.T) {
	end := "2026-05-24T09:30:00Z"
	segmentEnd := "2026-05-24T09:30:00Z"
	detail := &api.SessionDetail{
		SessionHistoryEntry: sharedtypes.SessionHistoryEntry{
			Session: sharedtypes.Session{ID: "session-1", StartTime: "2026-05-24T09:00:00Z", EndTime: &end},
			ParsedNotes: sharedtypes.ParsedSessionNotes{
				sharedtypes.SessionNoteSectionCommit: "commit",
				sharedtypes.SessionNoteSectionWork:   "work summary",
				sharedtypes.SessionNoteSectionNotes:  "notes",
			},
		},
		RepoName: "Repo", StreamName: "main", IssueTitle: "Issue",
		Segments: []sharedtypes.SessionSegment{{SegmentType: sharedtypes.SessionSegmentWork,
			StartTime: "2026-05-24T09:00:00Z", EndTime: &segmentEnd}},
	}
	state := OpenEditSessionHistory(State{Width: 100, Height: 40}, detail)
	if len(state.Inputs) != 2 || len(state.SessionHistoryNoteInputs) != 3 || len(state.SessionHistorySegments) != 1 {
		t.Fatalf("expected structured editor fields, got %+v", state)
	}
	if state.SessionHistoryContext == "" || state.SessionHistoryNoteInputs[0].Value() != "commit" {
		t.Fatalf("expected read-only issue context and populated commit field: %+v", state)
	}
	state, _, err := updateEditSessionHistory(state, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if err != "" || len(state.SessionHistorySegments) != 2 {
		t.Fatalf("expected add key to create a segment, err=%q segments=%d", err, len(state.SessionHistorySegments))
	}
	if got, want := state.SessionHistorySegments[1].StartTime.Value(), state.Inputs[0].Value()[:10]+" "; got != want {
		t.Fatalf("expected new segment start date %q, got %q", want, got)
	}
	if got, want := state.SessionHistorySegments[1].EndTime.Value(), state.Inputs[1].Value()[:10]+" "; got != want {
		t.Fatalf("expected new segment end date %q, got %q", want, got)
	}
	state.FocusIdx = 5
	state, _, err = updateEditSessionHistory(state, tea.KeyMsg{Type: tea.KeyRight})
	if err != "" || state.SessionHistorySegments[0].Type != sharedtypes.SessionSegmentShortBreak {
		t.Fatalf("expected right key to cycle segment type, err=%q type=%q", err, state.SessionHistorySegments[0].Type)
	}
	state, _, err = updateEditSessionHistory(state, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'D'}})
	if err != "" || len(state.SessionHistorySegments) != 1 {
		t.Fatalf("expected delete key to remove focused segment, err=%q segments=%d", err, len(state.SessionHistorySegments))
	}
}

func TestSessionHistoryUpdateRequestValidatesRanges(t *testing.T) {
	state := OpenEditSessionHistory(State{}, &api.SessionDetail{
		SessionHistoryEntry: sharedtypes.SessionHistoryEntry{Session: sharedtypes.Session{
			ID: "session-1", StartTime: "2026-05-24T09:00:00Z", EndTime: stringPointer("2026-05-24T09:30:00Z"),
		}},
		Segments: []sharedtypes.SessionSegment{{SegmentType: sharedtypes.SessionSegmentWork,
			StartTime: "2026-05-24T09:00:00Z", EndTime: stringPointer("2026-05-24T09:30:00Z")}},
	})
	request, err := sessionHistoryUpdateRequest(state)
	if err != nil || len(request.Segments) != 1 || request.ID != "session-1" {
		t.Fatalf("expected valid structured request, got request=%+v err=%v", request, err)
	}
	state.SessionHistorySegments[0].EndTime.SetValue("2026-05-24 10:00")
	if _, err := sessionHistoryUpdateRequest(state); err == nil {
		t.Fatal("expected segment outside session range to be rejected")
	}
}

func stringPointer(value string) *string { return &value }
