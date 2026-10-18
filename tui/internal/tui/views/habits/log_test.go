package habits

import (
	"strings"
	"testing"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/x/ansi"
)

func completion(date, name string, status sharedtypes.HabitCompletionStatus, minutes int, note string) api.HabitCompletion {
	entry := api.HabitCompletion{Date: date, HabitName: name, Status: status, RepoName: "Personal", StreamName: "wellbeing"}
	if minutes > 0 {
		entry.DurationMinutes = &minutes
	}
	if note != "" {
		entry.Notes = &note
	}
	return entry
}

func logState(width, height int) types.ContentState {
	done, failed := sharedtypes.HabitCompletionStatusCompleted, sharedtypes.HabitCompletionStatusFailed
	return types.ContentState{
		View: "habit_history", Pane: "habit_history", Width: width, Height: height,
		Cursors: map[string]int{"habit_history": 0}, Filters: map[string]string{},
		HabitHistory: []api.HabitCompletion{
			completion("2026-10-17", "Journal", done, 9, "Short reflection after a long day of reviews and planning."),
			completion("2026-10-17", "Morning Walk", failed, 0, "Weather broke the routine."),
			completion("2026-10-17", "Strength", done, 58, ""),
			completion("2026-10-16", "Journal", done, 10, ""),
		},
	}
}

func logText(width, height int) string {
	return ansi.Strip(renderHistoryView(types.Theme{}, logState(width, height)))
}

func TestHabitLogStrip(t *testing.T) {
	text := squash(logText(200, 40))
	for _, want := range []string{"Done 3 failed 1", "Most logged Journal 2"} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected %q in %s", want, text)
		}
	}
}

func TestHabitLogRows(t *testing.T) {
	text := logText(200, 40)
	for _, want := range []string{"✓ Journal", "9m", "Personal › wellbeing", "✗ Morning Walk"} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected %q in\n%s", want, text)
		}
	}
}

func TestHabitLogDividerSummary(t *testing.T) {
	text := logText(200, 40)
	if !strings.Contains(text, "Sat 17 Oct · 2 done · 1 failed") || !strings.Contains(text, "Fri 16 Oct · 1 done ") {
		t.Fatalf("expected day summaries, got\n%s", text)
	}
}

func TestHabitLogPanesByWidth(t *testing.T) {
	if text := logText(60, 20); strings.Contains(text, "Status") {
		t.Fatalf("expected the list alone when narrow, got\n%s", text)
	}
	for _, width := range []int{120, 200} {
		text := logText(width, 40)
		for _, want := range []string{"Date", "Status", "Context", "long day of reviews"} {
			if !strings.Contains(text, want) {
				t.Fatalf("%d: expected %q in\n%s", width, want, text)
			}
		}
		for _, line := range strings.Split(text, "\n") {
			if ansi.StringWidth(line) > width {
				t.Fatalf("%d: expected lines within the width, got %q", width, line)
			}
		}
	}
}

func TestHabitLogNoBox(t *testing.T) {
	if text := logText(120, 36); strings.Contains(text, "╭") || strings.Contains(text, "Habit History") {
		t.Fatalf("expected no box or old header, got\n%s", text)
	}
}

func squash(s string) string { return strings.Join(strings.Fields(s), " ") }
