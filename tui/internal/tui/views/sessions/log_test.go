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

func stamp(day, hour, minute int) string {
	return time.Date(2026, 10, day, hour, minute, 0, 0, time.Local).UTC().Format(time.RFC3339)
}

func session(id string, day, hour, minute, minutes int, note string) api.SessionHistoryEntry {
	end := stamp(day, hour, minute+minutes)
	seconds := minutes * 60
	entry := api.SessionHistoryEntry{
		Session: api.Session{
			ID:              id,
			IssueID:         7,
			StartTime:       stamp(day, hour, minute),
			EndTime:         &end,
			DurationSeconds: &seconds,
		},
	}
	if note != "" {
		entry.Notes = &note
	}
	return entry
}

func sessionState(width int) types.ContentState {
	return types.ContentState{
		View: "session_history", Pane: "sessions", Width: width, Height: 30,
		Cursors: map[string]int{"sessions": 0}, Filters: map[string]string{},
		SessionHistory: []api.SessionHistoryEntry{
			session("a", 17, 16, 30, 19, "Wired the preview\nsecond line"),
			session("b", 17, 10, 0, 42, ""),
			session("c", 16, 9, 0, 39, "Planning"),
		},
	}
}

func sessionText(state types.ContentState) string {
	return ansi.Strip(renderHistoryView(types.Theme{}, state))
}

func TestSessionLogStrip(t *testing.T) {
	text := squash(sessionText(sessionState(200)))
	for _, want := range []string{"Sessions 3 total 1h40m avg 33m", "This week", "longest 42m"} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected %q in %s", want, text)
		}
	}
}

func TestSessionLogRows(t *testing.T) {
	text := sessionText(sessionState(120))
	for _, want := range []string{"16:30–16:49", "19m", "Wired the preview", "no note"} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected %q in\n%s", want, text)
		}
	}
	if text := sessionText(sessionState(60)); strings.Contains(text, "second line") ||
		!strings.Contains(text, "Wired the preview") {
		t.Fatalf("expected only the note's first line in the list, got\n%s", text)
	}
	state := sessionState(120)
	state.SessionHistory[0].EndTime = nil
	if text := sessionText(state); !strings.Contains(text, "16:30–now") {
		t.Fatalf("expected an open session to end at now, got\n%s", text)
	}
}

func TestSessionLogDivider(t *testing.T) {
	if text := sessionText(sessionState(120)); !strings.Contains(text, "Sat 17 Oct · 2 · 1h01m") {
		t.Fatalf("expected day summary, got\n%s", text)
	}
}

func TestSessionDetailWaitsThenFills(t *testing.T) {
	state := sessionState(200)
	text := sessionText(state)
	for _, want := range []string{"16:30–16:49", "19m", "Wired the preview", "Loading…"} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected %q before the preview, got\n%s", want, text)
		}
	}
	segEnd := stamp(17, 16, 49)
	preview := &api.SessionDetail{
		SessionHistoryEntry: state.SessionHistory[0],
		RepoName:            "Crona", StreamName: "tui", IssueTitle: "Log redesign",
		WorkSummary: sharedtypes.SessionWorkSummary{
			WorkSeconds:  900,
			RestSeconds:  240,
			TotalSeconds: 1140,
		},
		Segments: []sharedtypes.SessionSegment{
			{
				SegmentType: sharedtypes.SessionSegmentWork,
				StartTime:   stamp(17, 16, 30),
				EndTime:     &segEnd,
			},
		},
	}
	state.SessionPreview = preview
	text = sessionText(state)
	for _, want := range []string{"Issue", "Log redesign", "Repo", "Crona", "Stream", "Work", "work  16:30–16:49"} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected %q with the preview, got\n%s", want, text)
		}
	}
	other := *preview
	other.ID = "zzz"
	state.SessionPreview = &other
	if text := sessionText(state); !strings.Contains(text, "Loading…") ||
		strings.Contains(text, "Log redesign") {
		t.Fatalf("expected a stale preview to be ignored, got\n%s", text)
	}
	state.SessionPreview, state.SessionPreviewFailed = nil, true
	if text := sessionText(state); !strings.Contains(text, "Details unavailable") {
		t.Fatalf("expected the failed state, got\n%s", text)
	}
}

func TestSessionLogNilDurations(t *testing.T) {
	state := sessionState(200)
	for i := range state.SessionHistory {
		state.SessionHistory[i].DurationSeconds = nil
	}
	if text := squash(sessionText(state)); !strings.Contains(text, "total 0m") {
		t.Fatalf("expected nil durations to count as 0, got %s", text)
	}
}

func squash(s string) string { return strings.Join(strings.Fields(s), " ") }
