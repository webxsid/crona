package sessionmeta

import (
	"fmt"
	"strings"
	"time"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	helperpkg "crona/tui/internal/tui/helpers"
	viewhelpers "crona/tui/internal/tui/views/helpers"
)

func FormatEstimateProgress(elapsedSeconds, estimateMinutes int) string {
	return fmt.Sprintf("%s / %s", viewhelpers.FormatClock(elapsedSeconds), helperpkg.FormatCompactDurationMinutes(estimateMinutes))
}

func SessionHistorySummary(entry api.SessionHistoryEntry) string {
	if entry.ParsedNotes != nil {
		if m := strings.TrimSpace(entry.ParsedNotes[sharedtypes.SessionNoteSectionCommit]); m != "" {
			return m
		}
		if n := strings.TrimSpace(entry.ParsedNotes[sharedtypes.SessionNoteSectionNotes]); n != "" {
			return n
		}
	}
	if entry.Notes != nil && strings.TrimSpace(*entry.Notes) != "" {
		return strings.TrimSpace(*entry.Notes)
	}
	return fmt.Sprintf("Issue #%d", entry.IssueID)
}

func FormatSessionTimestamp(value string) string {
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed.Local().Format("2006-01-02 15:04")
	}
	if len(value) >= 16 {
		return strings.Replace(value[:16], "T", " ", 1)
	}
	return value
}

func FormatSessionDuration(durationSeconds *int, start string, end *string) string {
	if durationSeconds != nil {
		return viewhelpers.FormatClock(*durationSeconds)
	}
	if end != nil && *end != "" {
		st, se := time.Parse(time.RFC3339, start)
		et, ee := time.Parse(time.RFC3339, *end)
		if se == nil && ee == nil {
			return viewhelpers.FormatClock(int(et.Sub(st).Seconds()))
		}
	}
	return "-"
}

func SummarizeCompletedSessions(s []api.Session) (workedSeconds int, completedCount int) {
	for _, session := range s {
		if session.DurationSeconds == nil || session.EndTime == nil {
			continue
		}
		workedSeconds += *session.DurationSeconds
		completedCount++
	}
	return
}

func RenderBigClock(clock string) string {
	glyphs := map[rune][]string{'0': {"███", "█ █", "█ █", "█ █", "███"}, '1': {" ██", "██ ", " ██", " ██", "███"}, '2': {"███", "  █", "███", "█  ", "███"}, '3': {"███", "  █", "███", "  █", "███"}, '4': {"█ █", "█ █", "███", "  █", "  █"}, '5': {"███", "█  ", "███", "  █", "███"}, '6': {"███", "█  ", "███", "█ █", "███"}, '7': {"███", "  █", "  █", "  █", "  █"}, '8': {"███", "█ █", "███", "█ █", "███"}, '9': {"███", "█ █", "███", "  █", "███"}, ':': {"   ", " █ ", "   ", " █ ", "   "}}
	lines := make([]string, 5)
	for _, char := range clock {
		glyph, ok := glyphs[char]
		if !ok {
			continue
		}
		for i := range lines {
			if lines[i] != "" {
				lines[i] += "  "
			}
			lines[i] += glyph[i]
		}
	}
	return strings.Join(lines, "\n")
}

func IssueMetaByID(all []api.IssueWithMeta, issueID int64) *api.IssueWithMeta {
	for i := range all {
		if all[i].ID == issueID {
			return &all[i]
		}
	}
	return nil
}

func FilteredSessionIndices(entries []api.SessionHistoryEntry, filter string) []int {
	filter = normalizeFilter(filter)
	out := []int{}
	for i, entry := range entries {
		text := strings.ToLower(SessionHistorySummary(entry))
		if filter == "" || strings.Contains(text, filter) {
			out = append(out, i)
		}
	}
	return out
}

func normalizeFilter(value string) string {
	return strings.TrimSpace(strings.ToLower(value))
}
