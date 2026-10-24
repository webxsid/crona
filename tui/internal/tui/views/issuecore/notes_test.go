package issuecore

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
)

func plainNotes(text string, width int) []string {
	lines := NotesLines(Theme{}, nil, &text, width)
	for i, line := range lines {
		lines[i] = strings.TrimRight(ansi.Strip(line), " ")
	}
	return lines
}

func inUTC(t *testing.T) {
	previous := time.Local
	time.Local = time.UTC
	t.Cleanup(func() { time.Local = previous })
}

func TestIssueNotesWithoutEntriesStayPlain(t *testing.T) {
	got := plainNotes("Just a thought.\nAnother line.", 60)
	want := []string{"", "Notes", "Just a thought.", "Another line."}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestIssueNotesSplitHistoryNewestFirst(t *testing.T) {
	inUTC(t)
	got := strings.Join(plainNotes(
		"Keep the window open.\n"+
			"[2026-10-12T17:40:00Z] ready -> blocked: waiting on API keys\n"+
			"[2026-10-14T09:00:00Z] blocked -> in_review", 80), "\n")
	want := strings.Join([]string{
		"", "Notes", "Keep the window open.",
		"", "History",
		"2026-10-14  09:00  blocked → in review",
		"2026-10-12  17:40  ready → blocked  waiting on API keys",
	}, "\n")
	if got != want {
		t.Fatalf("got\n%s\nwant\n%s", got, want)
	}
}

func TestIssueNotesHideEntryNoteRepeatingFreeText(t *testing.T) {
	inUTC(t)
	got := plainNotes(
		"Future planned work.\n[2026-10-14T09:00:00Z] planned -> ready: Future planned work.",
		80,
	)
	if last := got[len(got)-1]; last != "2026-10-14  09:00  planned → ready" {
		t.Fatalf("expected the repeated note hidden, got %q", got)
	}
}

func TestIssueNotesHistoryOnlyHasNoNotesHeading(t *testing.T) {
	inUTC(t)
	got := plainNotes("[2026-10-14T09:00:00Z] planned -> ready", 80)
	if strings.Join(got, "|") != "|History|2026-10-14  09:00  planned → ready" {
		t.Fatalf("got %q", got)
	}
}

func TestIssueNotesCapHistory(t *testing.T) {
	inUTC(t)
	var raw []string
	for day := 1; day <= 8; day++ {
		raw = append(raw, fmt.Sprintf("[2026-10-%02dT09:00:00Z] ready -> blocked", day))
	}
	got := plainNotes(strings.Join(raw, "\n"), 80)
	if len(got) != 2+5+1 || !strings.HasPrefix(got[2], "2026-10-08") ||
		got[len(got)-1] != "+3 earlier" {
		t.Fatalf("expected the latest 5 entries and +3 earlier, got %q", got)
	}
}

func TestIssueNotesWrapEntryNoteWhenNarrow(t *testing.T) {
	inUTC(t)
	got := plainNotes(
		"[2026-10-12T17:40:00Z] ready -> blocked: waiting on the API keys from the vendor",
		40,
	)
	for _, line := range got {
		if ansi.StringWidth(line) > 40 {
			t.Fatalf("line %q exceeds width 40 in %q", line, got)
		}
	}
	if got[2] != "2026-10-12  17:40  ready → blocked" ||
		!strings.HasPrefix(got[3], "  waiting on") {
		t.Fatalf("expected the note wrapped under the change, got %q", got)
	}
}
