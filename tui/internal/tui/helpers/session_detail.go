package helpers

import (
	"fmt"
	"strings"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
)

func SessionDetailContentLines(detail *api.SessionDetail) []string {
	if detail == nil {
		return []string{"Loading session detail...", "", "[esc] close"}
	}

	ended := "-"
	if detail.EndTime != nil && strings.TrimSpace(*detail.EndTime) != "" {
		ended = *detail.EndTime
	}
	duration := FormatSessionDurationText(detail.DurationSeconds, detail.StartTime, detail.EndTime)
	lines := []string{
		fmt.Sprintf("Repo: %s", detail.RepoName),
		fmt.Sprintf("Stream: %s", detail.StreamName),
		fmt.Sprintf("Issue: #%d %s", detail.IssueID, detail.IssueTitle),
		fmt.Sprintf("Started: %s", detail.StartTime),
		fmt.Sprintf("Ended: %s", ended),
		fmt.Sprintf("Duration: %s", duration),
		"",
		fmt.Sprintf("Work: %s", FormatClockText(detail.WorkSummary.WorkSeconds)),
		fmt.Sprintf("Rest: %s", FormatClockText(detail.WorkSummary.RestSeconds)),
		fmt.Sprintf("Segments: %d work / %d rest", detail.WorkSummary.WorkSegments, detail.WorkSummary.RestSegments),
	}

	sectionOrder := []sharedtypes.SessionNoteSection{
		sharedtypes.SessionNoteSectionCommit,
		sharedtypes.SessionNoteSectionContext,
		sharedtypes.SessionNoteSectionWork,
		sharedtypes.SessionNoteSectionNotes,
	}
	labels := map[sharedtypes.SessionNoteSection]string{
		sharedtypes.SessionNoteSectionCommit:  "Commit",
		sharedtypes.SessionNoteSectionContext: "Context",
		sharedtypes.SessionNoteSectionWork:    "Work Summary",
		sharedtypes.SessionNoteSectionNotes:   "Notes",
	}
	for _, section := range sectionOrder {
		value := ""
		if detail.ParsedNotes != nil {
			value = strings.TrimSpace(detail.ParsedNotes[section])
		}
		if value == "" {
			continue
		}
		lines = append(lines, "", labels[section]+":")
		lines = append(lines, strings.Split(value, "\n")...)
	}
	return lines
}

func SessionDetailViewportHeight(height int) int {
	if height < 16 {
		return maxInt(6, height-8)
	}
	return minInt(18, height-8)
}

func SessionDetailMaxOffset(width, height int, lines []string) int {
	boxWidth := minInt(maxInt(52, width-10), 96)
	innerWidth := boxWidth - 4
	wrapped := make([]string, 0, len(lines))
	for _, line := range lines {
		if line == "" {
			wrapped = append(wrapped, "")
			continue
		}
		wrapped = append(wrapped, wrapText(line, innerWidth)...)
	}
	return maxInt(0, len(wrapped)-SessionDetailViewportHeight(height))
}

func wrapText(text string, width int) []string {
	if width < 4 {
		return []string{text}
	}
	if strings.TrimSpace(text) == "" {
		return []string{""}
	}
	words := strings.Fields(text)
	lines := make([]string, 0, len(words))
	current := ""
	for _, word := range words {
		if current == "" {
			current = word
			continue
		}
		if len([]rune(current))+1+len([]rune(word)) <= width {
			current += " " + word
			continue
		}
		lines = append(lines, current)
		current = word
	}
	if current != "" {
		lines = append(lines, current)
	}
	if len(lines) == 0 {
		return []string{text}
	}
	return lines
}
