package controller

import (
	"testing"
	"time"

	sharedtypes "crona.work/shared/types"
	"crona.work/tui/internal/api"

	tea "github.com/charmbracelet/bubbletea"
)

func TestEditSessionHistoryOpensStructuredFieldsAndEditsSegments(t *testing.T) {
	end := "2026-05-24T09:30:00Z"
	segmentEnd := "2026-05-24T09:30:00Z"
	detail := &api.SessionDetail{
		SessionHistoryEntry: sharedtypes.SessionHistoryEntry{
			Session: sharedtypes.Session{
				ID:        "session-1",
				StartTime: "2026-05-24T09:00:00Z",
				EndTime:   &end,
			},
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
	state := OpenEditSessionHistory(State{Width: 100, Height: 40}, detail, "")
	if len(state.Inputs) != 2 || len(state.SessionHistoryNoteInputs) != 3 ||
		len(state.SessionHistorySegments) != 1 {
		t.Fatalf("expected structured editor fields, got %+v", state)
	}
	if state.SessionHistoryContext == "" || state.SessionHistoryNoteInputs[0].Value() != "commit" {
		t.Fatalf("expected read-only issue context and populated commit field: %+v", state)
	}
	state, _, err := updateEditSessionHistory(state, tea.KeyMsg{Type: tea.KeyCtrlA})
	if err != "" || len(state.SessionHistorySegments) != 2 {
		t.Fatalf(
			"expected add key to create a segment, err=%q segments=%d",
			err,
			len(state.SessionHistorySegments),
		)
	}
	if got, want := state.SessionHistorySegments[1].StartTime.Value(), state.SessionHistorySegments[0].EndTime.Value(); got != want {
		t.Fatalf("expected the new segment to start where the last one ends, %q, got %q", want, got)
	}
	if got := state.SessionHistorySegments[1].EndTime.Value(); got != "" {
		t.Fatalf("expected the new segment's end to be left for typing, got %q", got)
	}
	state.FocusIdx = 5
	state, _, err = updateEditSessionHistory(state, tea.KeyMsg{Type: tea.KeyRight})
	if err != "" || state.SessionHistorySegments[0].Type != sharedtypes.SessionSegmentShortBreak {
		t.Fatalf(
			"expected right key to cycle segment type, err=%q type=%q",
			err,
			state.SessionHistorySegments[0].Type,
		)
	}
	state, _, err = updateEditSessionHistory(state, tea.KeyMsg{Type: tea.KeyCtrlD})
	if err != "" || len(state.SessionHistorySegments) != 1 {
		t.Fatalf(
			"expected delete key to remove focused segment, err=%q segments=%d",
			err,
			len(state.SessionHistorySegments),
		)
	}
}

func TestSessionHistoryUpdateRequestValidatesRanges(t *testing.T) {
	state := OpenEditSessionHistory(State{}, &api.SessionDetail{
		SessionHistoryEntry: sharedtypes.SessionHistoryEntry{Session: sharedtypes.Session{
			ID: "session-1", StartTime: "2026-05-24T09:00:00Z", EndTime: stringPointer("2026-05-24T09:30:00Z"),
		}},
		Segments: []sharedtypes.SessionSegment{{SegmentType: sharedtypes.SessionSegmentWork,
			StartTime: "2026-05-24T09:00:00Z", EndTime: stringPointer("2026-05-24T09:30:00Z")}},
	}, "")
	request, err := sessionHistoryUpdateRequest(state)
	if err != nil || len(request.Segments) != 1 || request.ID != "session-1" {
		t.Fatalf("expected valid structured request, got request=%+v err=%v", request, err)
	}
	state.SessionHistorySegments[0].EndTime.SetValue(sessionHistoryStamp(time.Date(2026, 5, 24, 23, 0, 0, 0, time.Local), ""))
	if _, err := sessionHistoryUpdateRequest(state); err == nil {
		t.Fatal("expected segment outside session range to be rejected")
	}
}

func stringPointer(value string) *string { return &value }

func TestHistoryTypingDoesNotChangeSegments(t *testing.T) {
	detail := &api.SessionDetail{
		SessionHistoryEntry: sharedtypes.SessionHistoryEntry{
			Session: sharedtypes.Session{
				ID:        "s",
				StartTime: "2026-05-24T09:00:00Z",
				EndTime:   stringPointer("2026-05-24T09:01:00Z"),
			},
		},
	}
	for _, focus := range []int{2, 3, 4} {
		state := OpenEditSessionHistory(State{}, detail, "")
		state.FocusIdx = focus
		state = focusSessionHistoryControl(state)
		state, _, _ = updateEditSessionHistory(
			state,
			tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}},
		)
		state, _, _ = updateEditSessionHistory(
			state,
			tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'D'}},
		)
		if len(state.SessionHistorySegments) != 0 || state.FocusIdx != focus ||
			state.SessionHistoryNoteInputs[focus-2].Value() != "aD" {
			t.Fatalf("typing mutated segments or lost text: %+v", state)
		}
	}
}
func TestHistoryEditorPreservesSubMinutePrecision(t *testing.T) {
	start := "2026-05-24T09:00:15.123456789Z"
	end := "2026-05-24T09:00:45.123456789Z"
	state := OpenEditSessionHistory(
		State{},
		&api.SessionDetail{
			SessionHistoryEntry: sharedtypes.SessionHistoryEntry{
				Session: sharedtypes.Session{ID: "s", StartTime: start, EndTime: &end},
			},
			Segments: []sharedtypes.SessionSegment{
				{SegmentType: sharedtypes.SessionSegmentWork, StartTime: start, EndTime: &end},
			},
		},
		"",
	)
	request, err := sessionHistoryUpdateRequest(state)
	if err != nil {
		t.Fatal(err)
	}
	for _, pair := range [][2]string{{request.StartTime, start}, {request.EndTime, end}, {request.Segments[0].StartTime, start}, {*request.Segments[0].EndTime, end}} {
		got, _ := time.Parse(time.RFC3339Nano, pair[0])
		want, _ := time.Parse(time.RFC3339Nano, pair[1])
		if !got.Equal(want) {
			t.Fatalf("timestamp changed: %s want %s", pair[0], pair[1])
		}
	}
}

func TestHistorySegmentsSortChronologicallyWithinSecond(t *testing.T) {
	start := "2026-05-24T09:00:00Z"
	middle := "2026-05-24T09:00:00.5Z"
	end := "2026-05-24T09:00:01Z"
	for _, reverse := range []bool{false, true} {
		segments := []sharedtypes.SessionSegment{
			{SegmentType: sharedtypes.SessionSegmentWork, StartTime: start, EndTime: &middle},
			{SegmentType: sharedtypes.SessionSegmentShortBreak, StartTime: middle, EndTime: &end},
		}
		if reverse {
			segments[0], segments[1] = segments[1], segments[0]
		}
		state := OpenEditSessionHistory(
			State{},
			&api.SessionDetail{
				SessionHistoryEntry: sharedtypes.SessionHistoryEntry{
					Session: sharedtypes.Session{ID: "s", StartTime: start, EndTime: &end},
				},
				Segments: segments,
			},
			"",
		)
		request, err := sessionHistoryUpdateRequest(state)
		if err != nil {
			t.Fatalf("unchanged adjacent fractional segments rejected: %v", err)
		}
		if request.Segments[0].StartTime != start || request.Segments[1].StartTime != middle {
			t.Fatalf("incorrect chronological order: %+v", request.Segments)
		}
	}
}

func TestEditSessionHistoryKeepsBoundsOnFirstAndLastSegment(t *testing.T) {
	end, firstEnd, secondEnd := "2026-05-24T10:00:00Z", "2026-05-24T09:30:00Z", "2026-05-24T10:00:00Z"
	detail := &api.SessionDetail{
		SessionHistoryEntry: sharedtypes.SessionHistoryEntry{Session: sharedtypes.Session{
			ID: "session-1", StartTime: "2026-05-24T09:00:00Z", EndTime: &end,
		}},
		Segments: []sharedtypes.SessionSegment{
			{SegmentType: sharedtypes.SessionSegmentWork, StartTime: "2026-05-24T09:00:00Z", EndTime: &firstEnd},
			{SegmentType: sharedtypes.SessionSegmentShortBreak, StartTime: "2026-05-24T09:30:00Z", EndTime: &secondEnd},
		},
	}
	state := OpenEditSessionHistory(State{Width: 100, Height: 40}, detail, "")
	ts := func(utc string) string { return sessionHistoryClock(sessionHistoryParseUTC(utc)) }
	edit := func(state State, focus int, value string) State {
		state.FocusIdx = focus
		state = focusSessionHistoryControl(state)
		for range len([]rune(state.SessionHistorySegments[(focus-5)/3].EndTime.Value())) + 40 {
			state, _, _ = updateEditSessionHistory(state, tea.KeyMsg{Type: tea.KeyBackspace})
		}
		for _, r := range value {
			state, _, _ = updateEditSessionHistory(state, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		}
		return state
	}

	state = edit(state, sessionHistorySegmentFocus(0, 1), ts("2026-05-24T08:45:00Z"))
	if got := state.Inputs[0].Value(); got != sessionHistoryStamp(sessionHistoryParseUTC("2026-05-24T08:45:00Z"), "") {
		t.Fatalf("session start should follow the first segment's start, got %q", got)
	}
	state = edit(state, sessionHistorySegmentFocus(1, 2), ts("2026-05-24T10:15:00Z"))
	if got := state.Inputs[1].Value(); got != ts("2026-05-24T10:15:00Z") {
		t.Fatalf("session end should follow the last segment's end, got %q", got)
	}

	state.FocusIdx = sessionHistorySegmentFocus(1, 0)
	state, _, _ = updateEditSessionHistory(state, tea.KeyMsg{Type: tea.KeyCtrlD})
	if got := state.Inputs[1].Value(); got != ts(firstEnd) {
		t.Fatalf("removing the last segment should move the session end to %q, got %q", ts(firstEnd), got)
	}
}

func TestSessionHistoryShowsShortLocalTimesAndRollsPastMidnight(t *testing.T) {
	day := time.Date(2026, 11, 4, 23, 30, 0, 0, time.Local)
	start := day.UTC().Format(time.RFC3339Nano)
	end := day.Add(time.Hour).UTC().Format(time.RFC3339Nano)
	state := OpenEditSessionHistory(State{}, &api.SessionDetail{
		SessionHistoryEntry: sharedtypes.SessionHistoryEntry{Session: sharedtypes.Session{
			ID: "s", StartTime: start, EndTime: &end,
		}},
		Segments: []sharedtypes.SessionSegment{{SegmentType: sharedtypes.SessionSegmentWork, StartTime: start, EndTime: &end}},
	}, "")
	if got := state.Inputs[0].Value(); got != "2026-11-04 23:30" {
		t.Fatalf("start should show its local date and minute, got %q", got)
	}
	if got := state.Inputs[1].Value(); got != "2026-11-05 00:30" {
		t.Fatalf("an end on the next day should carry its date, got %q", got)
	}
	if got := state.SessionHistorySegments[0].StartTime.Value(); got != "23:30" {
		t.Fatalf("a segment on the start's day should show just the clock, got %q", got)
	}

	state.SessionHistorySegments[0].EndTime.SetValue("00:15")
	state.Inputs[1].SetValue("00:45")
	request, err := sessionHistoryUpdateRequest(state)
	if err != nil {
		t.Fatal(err)
	}
	want := day.Add(75 * time.Minute)
	if got, _ := time.Parse(time.RFC3339Nano, request.EndTime); !got.Equal(want) {
		t.Fatalf("a clock before the start should fall on the next day: got %s want %s", got, want)
	}
	if got, _ := time.Parse(time.RFC3339Nano, *request.Segments[0].EndTime); !got.Equal(day.Add(45 * time.Minute)) {
		t.Fatalf("segment end 00:15 should be the next day, got %s", got)
	}
}

func TestSessionHistoryUsesTheDateDisplayPattern(t *testing.T) {
	day := time.Date(2026, 11, 4, 11, 30, 0, 0, time.Local)
	start := day.UTC().Format(time.RFC3339Nano)
	end := day.AddDate(0, 0, 1).UTC().Format(time.RFC3339Nano)
	state := OpenEditSessionHistory(State{}, &api.SessionDetail{
		SessionHistoryEntry: sharedtypes.SessionHistoryEntry{Session: sharedtypes.Session{
			ID: "s", StartTime: start, EndTime: &end,
		}},
	}, "DD/MM/YYYY")
	if got := state.Inputs[0].Value(); got != "04/11/2026 11:30" {
		t.Fatalf("start should use the display pattern, got %q", got)
	}
	if got := state.Inputs[1].Value(); got != "05/11/2026 11:30" {
		t.Fatalf("an end on another day should use the display pattern, got %q", got)
	}
	for _, typed := range []string{"03/11/2026 09:15", "2026-11-03 09:15"} {
		state.Inputs[0].SetValue(typed)
		request, err := sessionHistoryUpdateRequest(state)
		if err != nil {
			t.Fatalf("%q: %v", typed, err)
		}
		if got, _ := time.Parse(time.RFC3339Nano, request.StartTime); !got.Equal(time.Date(2026, 11, 3, 9, 15, 0, 0, time.Local)) {
			t.Fatalf("%q parsed to %s", typed, got)
		}
	}
}
