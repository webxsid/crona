package helpers

import (
	"fmt"
	"strings"
	"time"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
)

func IssueScheduleLabel(issue api.Issue) string {
	if date := resolvedOnDate(issue.Status, issue.CompletedAt, issue.AbandonedAt); date != "" {
		return "on " + date
	}
	if issue.TodoForDate == nil {
		return ""
	}
	date := strings.TrimSpace(*issue.TodoForDate)
	if date == "" {
		return ""
	}
	if date == time.Now().Format("2006-01-02") {
		return "today"
	}
	return "due " + date
}

func resolvedOnDate(status sharedtypes.IssueStatus, completedAt, abandonedAt *string) string {
	var raw string
	switch status {
	case sharedtypes.IssueStatusDone:
		if completedAt != nil {
			raw = strings.TrimSpace(*completedAt)
		}
	case sharedtypes.IssueStatusAbandoned:
		if abandonedAt != nil {
			raw = strings.TrimSpace(*abandonedAt)
		}
	}
	if raw == "" {
		return ""
	}
	if parsed, err := time.Parse(time.RFC3339, raw); err == nil {
		return parsed.Format("2006-01-02")
	}
	if len(raw) >= len("2006-01-02") {
		return raw[:10]
	}
	return raw
}

func Deref(s *string) string {
	if s == nil {
		return "-"
	}
	return *s
}

func FirstNonEmpty(a, b *string) string {
	if a != nil && *a != "" {
		return *a
	}
	return Deref(b)
}

func Truncate(s string, max int) string {
	if max < 4 {
		max = 4
	}
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-3]) + "..."
}

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func SessionHistorySummary(entry api.SessionHistoryEntry) string {
	if entry.ParsedNotes != nil {
		if message := strings.TrimSpace(entry.ParsedNotes[sharedtypes.SessionNoteSectionCommit]); message != "" {
			return message
		}
		if note := strings.TrimSpace(entry.ParsedNotes[sharedtypes.SessionNoteSectionNotes]); note != "" {
			return note
		}
	}
	if entry.Notes != nil && strings.TrimSpace(*entry.Notes) != "" {
		return strings.TrimSpace(*entry.Notes)
	}
	return fmt.Sprintf("Issue #%d", entry.IssueID)
}

func NormalizeOptionalValue(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func NormalizeLookupName(value string) string {
	return strings.ToLower(strings.Join(strings.Fields(value), " "))
}

func SameLookupName(a, b string) bool {
	normalizedA := NormalizeLookupName(a)
	return normalizedA != "" && normalizedA == NormalizeLookupName(b)
}

func SessionCommit(detail *api.SessionDetail) string {
	if detail == nil || detail.ParsedNotes == nil {
		return ""
	}
	return strings.TrimSpace(detail.ParsedNotes[sharedtypes.SessionNoteSectionCommit])
}

func FormatClockText(totalSeconds int) string {
	if totalSeconds < 0 {
		totalSeconds = 0
	}
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}

func FormatSessionDurationText(durationSeconds *int, start string, end *string) string {
	if durationSeconds != nil {
		return FormatClockText(*durationSeconds)
	}
	if end != nil && *end != "" {
		st, se := time.Parse(time.RFC3339, start)
		et, ee := time.Parse(time.RFC3339, *end)
		if se == nil && ee == nil {
			return FormatClockText(int(et.Sub(st).Seconds()))
		}
	}
	return "-"
}
