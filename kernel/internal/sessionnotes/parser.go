package sessionnotes

import (
	"fmt"
	"math"
	"strings"
	"time"

	sharedtypes "crona/shared/types"
)

func Parse(raw *string) sharedtypes.ParsedSessionNotes {
	if raw == nil || *raw == "" {
		return sharedtypes.ParsedSessionNotes{}
	}

	lines := strings.Split(*raw, "\n")
	result := sharedtypes.ParsedSessionNotes{}

	var current sharedtypes.SessionNoteSection
	var hasSection bool
	var buffer []string

	flush := func() {
		if hasSection {
			result[current] = strings.TrimSpace(strings.Join(buffer, "\n"))
		}
		buffer = nil
	}

	for _, line := range lines {
		if strings.HasPrefix(line, "::") {
			flush()
			current = sharedtypes.SessionNoteSection(
				strings.TrimSpace(strings.TrimPrefix(line, "::")),
			)
			hasSection = true
			continue
		}
		buffer = append(buffer, line)
	}

	flush()
	return result
}

func Serialize(sections sharedtypes.ParsedSessionNotes) string {
	ordered := []sharedtypes.SessionNoteSection{
		sharedtypes.SessionNoteSectionCommit,
		sharedtypes.SessionNoteSectionContext,
		sharedtypes.SessionNoteSectionWork,
		sharedtypes.SessionNoteSectionNotes,
	}

	blocks := make([]string, 0, len(ordered))
	for _, key := range ordered {
		value := sections[key]
		if value == "" {
			continue
		}
		blocks = append(blocks, fmt.Sprintf("::%s\n%s", key, strings.TrimSpace(value)))
	}
	return strings.Join(blocks, "\n\n")
}

func AssertCommitMessage(notes *string) error {
	parsed := Parse(notes)
	if parsed[sharedtypes.SessionNoteSectionCommit] == "" {
		return fmt.Errorf("commit message is required in session notes")
	}
	return nil
}

func GenerateDefaultSessionNotes(input struct {
	Commit      *string
	RepoID      *int64
	StreamID    *int64
	IssueID     *int64
	WorkSummary []string
	Notes       *string
}) string {
	commit := "Work Session"
	if input.Commit != nil && strings.TrimSpace(*input.Commit) != "" {
		commit = strings.TrimSpace(*input.Commit)
	}

	contextLines := []string{}
	if input.RepoID != nil {
		contextLines = append(contextLines, fmt.Sprintf("Repo ID: %d", *input.RepoID))
	}
	if input.StreamID != nil {
		contextLines = append(contextLines, fmt.Sprintf("Stream ID: %d", *input.StreamID))
	}
	if input.IssueID != nil {
		contextLines = append(contextLines, fmt.Sprintf("Issue ID: %d", *input.IssueID))
	}

	sections := sharedtypes.ParsedSessionNotes{
		sharedtypes.SessionNoteSectionCommit:  commit,
		sharedtypes.SessionNoteSectionContext: strings.Join(contextLines, "\n"),
		sharedtypes.SessionNoteSectionWork:    strings.Join(input.WorkSummary, "\n"),
	}
	if input.Notes != nil && strings.TrimSpace(*input.Notes) != "" {
		sections[sharedtypes.SessionNoteSectionNotes] = strings.TrimSpace(*input.Notes)
	}
	return Serialize(sections)
}

func AmendCommitMessage(notes *string, additionalMessage string) string {
	parsed := Parse(notes)
	parsed[sharedtypes.SessionNoteSectionCommit] = strings.TrimSpace(additionalMessage)
	return Serialize(parsed)
}

func ComputeWorkSummary(segments []sharedtypes.SessionSegment) sharedtypes.SessionWorkSummary {
	var summary sharedtypes.SessionWorkSummary
	for _, segment := range segments {
		if segment.EndTime == nil {
			continue
		}
		duration := elapsedSeconds(segment.StartTime, *segment.EndTime)
		if segment.SegmentType == sharedtypes.SessionSegmentWork {
			summary.WorkSeconds += duration
			summary.WorkSegments++
		} else {
			summary.RestSeconds += duration
			summary.RestSegments++
		}
	}
	summary.TotalSeconds = summary.WorkSeconds + summary.RestSeconds
	return summary
}

func FormatWorkSummary(summary sharedtypes.SessionWorkSummary) []string {
	return []string{
		fmt.Sprintf(
			"Work: %s (%d segments)",
			formatDuration(summary.WorkSeconds),
			summary.WorkSegments,
		),
		fmt.Sprintf(
			"Rest: %s (%d segments)",
			formatDuration(summary.RestSeconds),
			summary.RestSegments,
		),
		fmt.Sprintf("Total: %s", formatDuration(summary.TotalSeconds)),
	}
}

func elapsedSeconds(startTime string, endTime string) int {
	start, err := time.Parse(time.RFC3339, startTime)
	if err != nil {
		return 0
	}
	end, err := time.Parse(time.RFC3339, endTime)
	if err != nil {
		return 0
	}
	seconds := int(math.Floor(end.Sub(start).Seconds()))
	if seconds < 0 {
		return 0
	}
	return seconds
}

func formatDuration(seconds int) string {
	value, unit := convertTimeInUnits(seconds)
	return fmt.Sprintf("%.2f%s", value, unit)
}

func convertTimeInUnits(seconds int) (float64, string) {
	switch {
	case seconds < 60:
		return float64(seconds), "s"
	case seconds < 3600:
		return float64(seconds) / 60, "m"
	case seconds < 86400:
		return float64(seconds) / 3600, "h"
	default:
		return float64(seconds) / 86400, "d"
	}
}
