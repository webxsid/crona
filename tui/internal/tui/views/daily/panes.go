package daily

import (
	"fmt"
	"strings"

	"crona/tui/internal/api"
	helperpkg "crona/tui/internal/tui/helpers"
	viewchrome "crona/tui/internal/tui/views/chrome"
	contextmeta "crona/tui/internal/tui/views/contextmeta"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	issuecore "crona/tui/internal/tui/views/issuecore"
	viewruntime "crona/tui/internal/tui/views/runtime"
	sessionmeta "crona/tui/internal/tui/views/sessionmeta"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
)

func renderIssues(theme types.Theme, state types.ContentState, width, height int) string {
	active := state.Pane == "issues"
	cur := state.Cursors["issues"]
	issues := make([]issuecore.APIIssue, 0, len(state.DailyIssues))
	for _, issue := range filteredDailyIssues(state) {
		issues = append(issues, issuecore.NewAPIIssue(issue.ID, issue.Title, issue.Status, issue.EstimateMinutes, issue.PinnedDaily, issue.TodoForDate, issue.CompletedAt, issue.AbandonedAt))
	}
	indices := issuecore.FilteredIssueIndices(issues, state.Filters["issues"])
	total := len(indices)
	actions := viewchrome.PaneActionsForState(theme, state, active)
	actionLine := viewchrome.RenderPaneActionLine(theme, state.Filters["issues"], width-6, actions)
	titleLine := dailyTaskTitle(theme, state.DailyTaskSection)

	// append [h/l] to the title line
	titleLine += viewhelpers.StringsJoin([]string{
		theme.StyleDim.Render("  [h/l]"),
	})

	lines := []string{titleLine, theme.StyleHeader.Render(contextmeta.DefaultScopeLabel(state.Context)), actionLine}
	if len(issues) == 0 || total == 0 {
		lines = append(lines, theme.StyleDim.Render(dailyTaskEmptyMessage(state.DailyTaskSection)))
		return viewchrome.RenderPaneBox(theme, active, width, height, viewhelpers.StringsJoin(lines))
	}
	statusW := 11
	estimateW := 8
	repoW := max(10, width/7)
	streamW := max(10, width/7)
	titleW := width - statusW - estimateW - repoW - streamW - 16
	titleW = max(14, titleW)
	header := fmt.Sprintf("%-2s %-*s %-*s %-*s %-*s %-*s", "", titleW, "Issue", statusW, "Status", estimateW, "Estimate", repoW, "Repo", streamW, "Stream")
	lines = append(lines, theme.StyleDim.Render(viewhelpers.Truncate(header, width-6)))
	inner := viewchrome.RemainingPaneHeight(height, lines)
	issueSlots := max(1, inner)
	start, end := viewchrome.ListWindow(cur, total, issueSlots)
	if start > 0 {
		lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("↑ %d more", start)))
	}
	for _, rawIdx := range indices[start:end] {
		issue := issues[rawIdx]
		meta := sessionmeta.IssueMetaByID(state.AllIssues, issue.ID)
		repoName, streamName := "-", "-"
		if meta != nil {
			repoName = meta.RepoName
			streamName = meta.StreamName
		}
		estimate := "-"
		if issue.EstimateMinutes != nil {
			estimate = helperpkg.FormatCompactDurationMinutes(*issue.EstimateMinutes)
		}
		title := issue.Title + issuecore.IssueDueSuffix(issue.Status, issue.TodoForDate, issue.CompletedAt, issue.AbandonedAt, state.Settings)
		row := fmt.Sprintf("%-2s %-*s %-*s %-*s %-*s %-*s", "", titleW, viewhelpers.Truncate(title, titleW), statusW, viewhelpers.Truncate(issuecore.PlainIssueStatus(string(issue.Status)), statusW), estimateW, estimate, repoW, viewhelpers.Truncate(repoName, repoW), streamW, viewhelpers.Truncate(streamName, streamW))
		lines = append(lines, viewchrome.RenderPaneRowStyled(theme, rawIdx, cur, active, row, issuecore.IssueStatusStyle(theme, string(issue.Status)), width))
	}
	if remaining := total - end; remaining > 0 {
		lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("↓ %d more", remaining)))
	}
	return viewchrome.RenderPaneBox(theme, active, width, height, viewhelpers.StringsJoin(lines))
}

func dailyTaskTitle(theme types.Theme, section string) string {
	options := []struct {
		key   string
		label string
	}{
		{key: "planned", label: "Planned"},
		{key: "pinned", label: "Pinned"},
		{key: "overdue", label: "Overdue"},
	}
	parts := []string{theme.StylePaneTitle.Render("Tasks [1]")}
	for _, opt := range options {
		label := opt.label
		if section == opt.key {
			label = theme.StyleSelectedInverse.Render(label)
		} else {
			label = theme.StyleDim.Render(label)
		}
		parts = append(parts, label)
	}
	return strings.Join(parts, "  ")
}

func dailyTaskEmptyMessage(section string) string {
	switch section {
	case "pinned":
		return "No pinned tasks in this scope"
	case "overdue":
		return "No overdue tasks in this scope"
	default:
		return "No planned tasks for this date"
	}
}

func filteredDailyIssues(state types.ContentState) []api.Issue {
	if len(state.DailyIssues) == 0 {
		return nil
	}
	anchorDate := strings.TrimSpace(state.DashboardDate)
	if anchorDate == "" && state.DailySummary != nil {
		anchorDate = strings.TrimSpace(state.DailySummary.Date)
	}
	if anchorDate == "" {
		return state.DailyIssues
	}
	out := make([]api.Issue, 0, len(state.DailyIssues))
	for _, issue := range state.DailyIssues {
		if dailyTaskMatchesSection(issue, anchorDate, state.DailyTaskSection) {
			out = append(out, issue)
		}
	}
	return out
}

func dailyTaskMatchesSection(issue api.Issue, anchorDate, section string) bool {
	if issue.TodoForDate != nil {
		due := strings.TrimSpace(*issue.TodoForDate)
		switch section {
		case "overdue":
			return due != "" && due < anchorDate && issue.Status != "done" && issue.Status != "abandoned"
		case "planned":
			return due == anchorDate
		}
	}
	if section == "pinned" {
		if issue.PinnedDaily {
			if issue.TodoForDate == nil {
				return true
			}
			due := strings.TrimSpace(*issue.TodoForDate)
			return due >= anchorDate
		}
	}
	return false
}

func renderHabits(theme types.Theme, state types.ContentState, width, height int) string {
	active := state.Pane == "habits"
	cur := state.Cursors["habits"]
	indices := viewhelpers.FilteredStrings(viewruntime.HabitDailyItems(state.DueHabits), state.Filters["habits"])
	total := len(indices)
	actions := viewchrome.PaneActionsForState(theme, state, active)
	actionLine := viewchrome.RenderPaneActionLine(theme, state.Filters["habits"], width-6, actions)
	lines := []string{theme.StylePaneTitle.Render("Habits Due [2]"), theme.StyleHeader.Render(contextmeta.DefaultScopeLabel(state.Context)), actionLine}
	if total == 0 {
		lines = append(lines, theme.StyleDim.Render("No due habits for this date"))
		return viewchrome.RenderPaneBox(theme, active, width, height, viewhelpers.StringsJoin(lines))
	}
	inner := viewchrome.RemainingPaneHeight(height, lines)
	start, end := viewchrome.ListWindow(cur, total, inner)
	if start > 0 {
		lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("↑ %d more", start)))
	}
	for i := start; i < end; i++ {
		habit := state.DueHabits[indices[i]]
		status := "[ ]"
		style := &theme.StyleNormal
		switch habit.Status {
		case "completed":
			status = "[x]"
			s := lipgloss.NewStyle().Foreground(theme.ColorGreen)
			style = &s
		case "failed":
			status = "[!]"
			s := lipgloss.NewStyle().Foreground(theme.ColorRed)
			style = &s
		}
		duration := ""
		if habit.DurationMinutes != nil {
			duration = "  " + helperpkg.FormatCompactDurationMinutes(*habit.DurationMinutes)
		} else if habit.TargetMinutes != nil {
			duration = "  target " + helperpkg.FormatCompactDurationMinutes(*habit.TargetMinutes)
		}
		lines = append(lines, viewchrome.RenderPaneRowStyled(theme, i, cur, active, fmt.Sprintf("%s %s%s", status, habit.Name, duration), style, width))
	}
	if remaining := total - end; remaining > 0 {
		lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("↓ %d more", remaining)))
	}
	return viewchrome.RenderPaneBox(theme, active, width, height, viewhelpers.StringsJoin(lines))
}

func renderIssueStatusBar(theme types.Theme, counts map[string]int, width int) string {
	if width < 8 {
		width = 8
	}
	order := []struct {
		status string
		color  lipgloss.Color
	}{
		{status: "done", color: theme.ColorGreen},
		{status: "abandoned", color: theme.ColorRed},
		{status: "blocked", color: theme.ColorRed},
		{status: "in_progress", color: theme.ColorYellow},
		{status: "in_review", color: theme.ColorMagenta},
		{status: "ready", color: theme.ColorCyan},
		{status: "planned", color: theme.ColorBlue},
		{status: "backlog", color: theme.ColorSubtle},
	}
	total := 0
	for _, count := range counts {
		total += count
	}
	if total == 0 {
		return theme.StyleDim.Render(strings.Repeat("·", width))
	}
	segments := make([]string, 0, len(order))
	used := 0
	remainingStatuses := 0
	for _, item := range order {
		if counts[item.status] > 0 {
			remainingStatuses++
		}
	}
	for _, item := range order {
		count := counts[item.status]
		if count <= 0 {
			continue
		}
		segmentWidth := (count * width) / total
		if segmentWidth == 0 {
			segmentWidth = 1
		}
		if used+segmentWidth > width {
			segmentWidth = width - used
		}
		if remainingStatuses == 1 {
			segmentWidth = width - used
		}
		if segmentWidth <= 0 {
			continue
		}
		segments = append(segments, lipgloss.NewStyle().Foreground(item.color).Render(strings.Repeat("█", segmentWidth)))
		used += segmentWidth
		remainingStatuses--
	}
	if used < width {
		segments = append(segments, theme.StyleDim.Render(strings.Repeat("█", width-used)))
	}
	return strings.Join(segments, "")
}

func renderIssueLegend(counts map[string]int) string {
	labels := []struct {
		status string
		label  string
	}{
		{status: "done", label: "done"},
		{status: "abandoned", label: "abandoned"},
		{status: "blocked", label: "blocked"},
		{status: "in_progress", label: "active"},
		{status: "in_review", label: "review"},
		{status: "ready", label: "ready"},
		{status: "planned", label: "planned"},
		{status: "backlog", label: "backlog"},
	}
	parts := make([]string, 0, len(labels))
	for _, item := range labels {
		if counts[item.status] <= 0 {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s %d", item.label, counts[item.status]))
	}
	if len(parts) == 0 {
		return "no issues scheduled"
	}
	return strings.Join(parts, "   ")
}

func renderHabitBar(theme types.Theme, completed, failed, total, width int) string {
	if width < 8 {
		width = 8
	}
	if total <= 0 {
		return theme.StyleDim.Render(strings.Repeat("·", width))
	}
	completedWidth := (completed * width) / total
	failedWidth := (failed * width) / total
	if completed > 0 && completedWidth == 0 {
		completedWidth = 1
	}
	if failed > 0 && failedWidth == 0 {
		failedWidth = 1
	}
	if completedWidth+failedWidth > width {
		failedWidth = max(0, width-completedWidth)
	}
	remainingWidth := width - completedWidth - failedWidth
	return lipgloss.NewStyle().Foreground(theme.ColorGreen).Render(strings.Repeat("█", completedWidth)) +
		lipgloss.NewStyle().Foreground(theme.ColorRed).Render(strings.Repeat("█", failedWidth)) +
		theme.StyleDim.Render(strings.Repeat("█", remainingWidth))
}
