package helpers

import (
	"fmt"
	"strings"

	"crona/tui/internal/api"
)

func SessionContextContentLines(issue *api.IssueWithMeta) []string {
	if issue == nil {
		return []string{"No active issue selected."}
	}
	lines := []string{
		fmt.Sprintf("Repo: %s", issue.RepoName),
		fmt.Sprintf("Stream: %s", issue.StreamName),
		fmt.Sprintf("Issue: #%d %s", issue.ID, issue.Title),
	}
	if issue.EstimateMinutes != nil && *issue.EstimateMinutes > 0 {
		lines = append(lines, fmt.Sprintf("Estimate: %dm", *issue.EstimateMinutes))
	}
	lines = append(lines, "", "Description:")
	if issue.Description != nil && strings.TrimSpace(*issue.Description) != "" {
		lines = append(lines, strings.TrimSpace(*issue.Description))
	} else {
		lines = append(lines, "-")
	}
	lines = append(lines, "", "Notes:")
	if issue.Notes != nil && strings.TrimSpace(*issue.Notes) != "" {
		lines = append(lines, strings.TrimSpace(*issue.Notes))
	} else {
		lines = append(lines, "-")
	}
	return lines
}

func SessionContextMaxOffset(width, height int, lines []string) int {
	boxWidth := minInt(maxInt(50, width-10), 92)
	innerWidth := boxWidth - 4
	visibleHeight := maxInt(6, height-10)
	wrapped := make([]string, 0, len(lines))
	for _, line := range lines {
		if line == "" {
			wrapped = append(wrapped, "")
			continue
		}
		wrapped = append(wrapped, wrapText(line, innerWidth)...)
	}
	return maxInt(0, len(wrapped)-visibleHeight)
}
