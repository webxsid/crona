package issues

import (
	"fmt"
	"strings"
	"time"

	helperpkg "crona/tui/internal/tui/helpers"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/x/ansi"
)

// StripRows renders the scope summary in two width-exact rows, dropping rightmost groups first.
func StripRows(theme types.Theme, state types.ContentState, width int) []string {
	if width <= 0 {
		return []string{"", ""}
	}
	counts := map[string]int{}
	open, closed := 0, 0
	for _, issue := range state.DefaultIssues {
		counts[string(issue.Status)]++
		if issue.Status == "done" || issue.Status == "abandoned" {
			closed++
		} else {
			open++
		}
	}
	bar := func(n int) string {
		filled := 0
		if open+closed > 0 {
			filled = n * 10 / (open + closed)
		}
		return theme.StyleHeader.Render(strings.Repeat("█", filled)) + theme.StyleDim.Render(strings.Repeat("░", 10-filled))
	}
	groups := [][2]string{
		{fmt.Sprintf("Open %d  %s", open, bar(open)), fmt.Sprintf("Closed %d  %s", closed, bar(closed))},
		{fmt.Sprintf("in progress %d  blocked %d", counts["in_progress"], counts["blocked"]), fmt.Sprintf("done %d  abandoned %d", counts["done"], counts["abandoned"])},
		{"Due  " + dueSummary(state), estimateSummary(state)},
	}
	rows := []string{"", ""}
	for _, group := range groups {
		columnWidth := max(ansi.StringWidth(group[0]), ansi.StringWidth(group[1]))
		gap := ""
		if rows[0] != "" {
			gap = "   "
		}
		if ansi.StringWidth(rows[0])+len(gap)+columnWidth > width {
			break
		}
		for i := range rows {
			rows[i] += gap + group[i] + strings.Repeat(" ", columnWidth-ansi.StringWidth(group[i]))
		}
	}
	for i := range rows {
		rows[i] = ansi.Truncate(rows[i], width, "")
		rows[i] += strings.Repeat(" ", width-ansi.StringWidth(rows[i]))
	}
	return rows
}

func dueSummary(state types.ContentState) string {
	today := 0
	overdue := 0
	now := time.Now().Format("2006-01-02")
	for _, issue := range state.DefaultIssues {
		if issue.TodoForDate == nil || issue.Status == "done" || issue.Status == "abandoned" {
			continue
		}
		if *issue.TodoForDate == now {
			today++
		}
		if *issue.TodoForDate < now {
			overdue++
		}
	}
	return fmt.Sprintf("today %d  overdue %d", today, overdue)
}

func estimateSummary(state types.ContentState) string {
	estimated, scoped := 0, 0
	for _, issue := range state.DefaultIssues {
		if issue.Status == "done" || issue.Status == "abandoned" || issue.EstimateMinutes == nil {
			continue
		}
		estimated += *issue.EstimateMinutes
		scoped++
	}
	return fmt.Sprintf(
		"Est. %s over %d issues",
		helperpkg.FormatCompactDurationMinutes(estimated),
		scoped,
	)
}
