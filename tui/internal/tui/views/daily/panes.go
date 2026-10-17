package daily

import (
	"fmt"
	"strings"

	"crona/tui/internal/api"
	helperpkg "crona/tui/internal/tui/helpers"
	viewchrome "crona/tui/internal/tui/views/chrome"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	issuecore "crona/tui/internal/tui/views/issuecore"
	viewruntime "crona/tui/internal/tui/views/runtime"
	sessionmeta "crona/tui/internal/tui/views/sessionmeta"
	types "crona/tui/internal/tui/views/types"
	viewui "crona/tui/internal/tui/views/ui"

	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
)

// dailyIssueTableMinWidth is the section width below which tasks render as two-line items.
const dailyIssueTableMinWidth = 120

func renderIssues(theme types.Theme, state types.ContentState, width, height int) string {
	active := state.Pane == "issues"
	cur := state.Cursors["issues"]
	sources := filteredDailyIssues(state)
	issues := dailyAPIIssues(sources)
	indices := dailyIssueIndices(state, issues)
	total := len(indices)
	base := viewui.PaneBase{Focused: active, Width: width, Height: height, Cursor: cur}
	meta := fmt.Sprintf("%d  [←/→]", total)
	lines := []string{dailyTaskTitle(theme, state.DailyTaskSection)}
	if len(issues) == 0 || total == 0 {
		lines = append(lines, theme.StyleDim.Render(dailyTaskEmptyMessage(state.DailyTaskSection)))
		return dailySection(theme, "Tasks", meta, active, width, height, lines)
	}
	if width < dailyIssueTableMinWidth {
		lines = renderCompactDailyIssueList(
			theme,
			state,
			width,
			height,
			base,
			lines,
			issues,
			sources,
			indices,
			total,
			cur,
			active,
		)
		return dailySection(theme, "Tasks", meta, active, width, height, lines)
	}
	layout := issuecore.IssueTableLayoutForWidth(width)
	if state.DailyTaskSection != "planned" {
		layout = issuecore.IssueTableLayoutForWidthWithDate(width)
	}
	inner := sectionBodyRows(height, lines)
	// The table adds a column-header row above the issue rows.
	start, end := listWindow(cur, total, inner-1)
	if start > 0 {
		lines = append(lines, base.MoreAbove(theme, start))
	}
	tableRows := make([]table.Row, 0, end-start)
	for offset, rawIdx := range indices[start:end] {
		pos := start + offset
		issue := issues[rawIdx]
		meta := sessionmeta.IssueMetaByID(state.AllIssues, issue.ID)
		repoName, streamName := dailyIssueContext(sources[rawIdx])
		workedSeconds := issue.WorkedSeconds
		if meta != nil {
			repoName = meta.RepoName
			streamName = meta.StreamName
			if workedSeconds <= 0 && meta.WorkedSeconds > 0 {
				workedSeconds = meta.WorkedSeconds
			}
		}
		title := issue.Title
		rowStyle := lipgloss.NewStyle()
		if statusStyle := issuecore.IssueStatusStyle(theme, string(issue.Status)); statusStyle != nil {
			rowStyle = *statusStyle
		}
		if pos == cur && active {
			rowStyle = rowStyle.Bold(true)
		}
		cursor := " "
		if pos == cur && active {
			cursor = viewchrome.SelectionCursor
		}
		tableRows = append(
			tableRows,
			issuecore.IssueTableRow(cursor, issuecore.IssueTableData{
				Issue:    title,
				Status:   issuecore.PlainIssueStatus(string(issue.Status)),
				Estimate: issuecore.IssueEstimateLabel(issue.EstimateMinutes),
				Worked:   issuecore.IssueWorkedLabel(workedSeconds),
				Repo:     repoName,
				Stream:   streamName,
				Date: issuecore.IssueDateLabel(
					issue.Status,
					issue.TodoForDate,
					issue.CompletedAt,
					issue.AbandonedAt,
					state.Settings,
				),
				Context: issuecore.IssueContextLabel(repoName, streamName),
				Effort: issuecore.IssueWorkedEstimateCompactLabel(
					workedSeconds,
					issue.EstimateMinutes,
				),
			}, rowStyle),
		)
	}
	tablePane := viewui.TablePane{
		Columns:     issuecore.IssueTableColumns(layout),
		Rows:        tableRows,
		HeaderStyle: theme.StyleDim,
	}
	lines = append(lines, tablePane.View())
	if remaining := total - end; remaining > 0 {
		lines = append(lines, base.MoreBelow(theme, remaining))
	}
	return dailySection(theme, "Tasks", meta, active, width, height, lines)
}

func dailyAPIIssues(sources []api.Issue) []issuecore.APIIssue {
	issues := make([]issuecore.APIIssue, 0, len(sources))
	for _, issue := range sources {
		issues = append(
			issues,
			issuecore.NewAPIIssue(
				issue.ID,
				issue.Title,
				issue.Status,
				issue.EstimateMinutes,
				issue.WorkedSeconds,
				issue.PinnedDaily,
				issue.TodoForDate,
				issue.CompletedAt,
				issue.AbandonedAt,
			),
		)
	}
	return issues
}

// dailyTaskRows is the height the Tasks section needs to show every task at width.
func dailyTaskRows(state types.ContentState, width int) int {
	total := len(dailyIssueIndices(state, dailyAPIIssues(filteredDailyIssues(state))))
	// Header and the Planned/Pinned/Overdue row come first.
	switch {
	case total == 0:
		return 3
	case width >= dailyIssueTableMinWidth:
		return 3 + total
	default:
		return 2 + total*3 - 1
	}
}

func renderCompactDailyIssueList(
	theme types.Theme,
	state types.ContentState,
	width, height int,
	base viewui.PaneBase,
	lines []string,
	issues []issuecore.APIIssue,
	sources []api.Issue,
	indices []int,
	total, cur int,
	active bool,
) []string {
	inner := sectionBodyRows(height, lines)
	availableRows := inner
	// Each item is a title row and a meta row, with a divider between items.
	itemSlots := max(1, (availableRows+1)/3)
	start, end := viewchrome.ListWindow(cur, total, itemSlots)
	if start > 0 {
		availableRows--
	}
	if remaining := total - end; remaining > 0 {
		availableRows--
	}
	itemSlots = max(1, (availableRows+1)/3)
	start, end = viewchrome.ListWindow(cur, total, itemSlots)
	if start > 0 {
		lines = append(lines, base.MoreAbove(theme, start))
	}
	for idx, rawIdx := range indices[start:end] {
		if idx > 0 {
			lines = append(lines, theme.StyleDim.Render(renderCompactDailyIssueDivider(width)))
		}
		issue := issues[rawIdx]
		meta := sessionmeta.IssueMetaByID(state.AllIssues, issue.ID)
		repoName, streamName := dailyIssueContext(sources[rawIdx])
		workedSeconds := issue.WorkedSeconds
		if meta != nil {
			repoName = meta.RepoName
			streamName = meta.StreamName
			if workedSeconds <= 0 && meta.WorkedSeconds > 0 {
				workedSeconds = meta.WorkedSeconds
			}
		}
		title := issue.Title
		lines = append(
			lines,
			renderCompactDailyIssueTitle(
				theme,
				width,
				start+idx,
				cur,
				active,
				title,
				compactDailyIssueDate(state, issue),
				string(issue.Status),
			),
			renderCompactDailyIssueMeta(
				theme,
				width,
				start+idx,
				cur,
				active,
				repoName,
				streamName,
				workedSeconds,
				issue.EstimateMinutes,
				string(issue.Status),
			),
		)
	}
	if remaining := total - end; remaining > 0 {
		lines = append(lines, base.MoreBelow(theme, remaining))
	}
	return lines
}

// dailyIssueContext is the issue's own repo and stream, used until session metadata has it.
func dailyIssueContext(issue api.Issue) (string, string) {
	repo, stream := strings.TrimSpace(issue.RepoName), strings.TrimSpace(issue.StreamName)
	if repo == "" {
		repo = "-"
	}
	if stream == "" {
		stream = "-"
	}
	return repo, stream
}

func renderCompactDailyIssueTitle(
	theme types.Theme,
	width int,
	i, cur int,
	active bool,
	title, date, status string,
) string {
	rowStyle := lipgloss.NewStyle()
	if statusStyle := issuecore.IssueStatusStyle(theme, status); statusStyle != nil {
		rowStyle = *statusStyle
	}
	if i == cur && active {
		rowStyle = rowStyle.Bold(true)
	}
	cursor := " "
	if i == cur && active {
		cursor = viewchrome.SelectionCursor
	}
	if date != "" {
		title += "  " + dateField(date)
	}
	line := viewhelpers.TruncateANSI(title, max(18, width-4))
	line = rowStyle.Render(line)
	if i == cur && active {
		return theme.StyleCursor.Render(cursor + " " + line)
	}
	if i == cur {
		return theme.StyleSelected.Render("  " + line)
	}
	return theme.StyleNormal.Render("  " + line)
}

func renderCompactDailyIssueMeta(
	theme types.Theme,
	width int,
	i, cur int,
	active bool,
	repoName, streamName string,
	workedSeconds int,
	estimateMinutes *int,
	status string,
) string {
	metaParts := []string{
		issuecore.IssueContextLabel(repoName, streamName),
		issuecore.IssueWorkedEstimateCompactLabel(workedSeconds, estimateMinutes),
		issuecore.PlainIssueStatus(status),
	}
	meta := strings.Join(metaParts, " | ")
	if i == cur && active {
		meta = theme.StyleSelected.Render("  " + meta)
	} else {
		meta = theme.StyleDim.Render("  " + meta)
	}
	return viewhelpers.TruncateANSI(meta, width)
}

func compactDailyIssueDate(state types.ContentState, issue issuecore.APIIssue) string {
	if state.DailyTaskSection == "planned" {
		return ""
	}
	return issuecore.IssueDateLabel(
		issue.Status,
		issue.TodoForDate,
		issue.CompletedAt,
		issue.AbandonedAt,
		state.Settings,
	)
}

func dateField(date string) string {
	if date == "" {
		return ""
	}
	return "[" + date + "]"
}

func renderCompactDailyIssueDivider(width int) string {
	return strings.Repeat("-", max(8, width-4))
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
	parts := []string{}
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

// filteredDailyIssues is the model's section-scoped task list; the view never re-scopes it.
func filteredDailyIssues(state types.ContentState) []api.Issue {
	return state.DailyIssues
}

// dailyIssueIndices is the model's filtered order for Tasks; the cursor is a position in it.
func dailyIssueIndices(state types.ContentState, issues []issuecore.APIIssue) []int {
	if state.DailyIssueIndices != nil {
		return state.DailyIssueIndices
	}
	return issuecore.FilteredIssueIndices(issues, state.Filters["issues"])
}

// dailyHabitIndices is the model's filtered order for Habits; the cursor is a position in it.
func dailyHabitIndices(state types.ContentState) []int {
	if state.DailyHabitIndices != nil {
		return state.DailyHabitIndices
	}
	return viewhelpers.FilteredStrings(
		viewruntime.HabitDailyItems(state.DueHabits),
		state.Filters["habits"],
	)
}

func renderHabits(theme types.Theme, state types.ContentState, width, height int) string {
	active := state.Pane == "habits"
	cur := state.Cursors["habits"]
	indices := dailyHabitIndices(state)
	total := len(indices)
	done := 0
	for _, habit := range state.DueHabits {
		if habit.Status == "completed" {
			done++
		}
	}
	meta := fmt.Sprintf("%d/%d", done, len(state.DueHabits))
	lines := []string{}
	if total == 0 {
		lines = append(lines, theme.StyleDim.Render("No due habits for this date"))
		return dailySection(theme, "Habits Due", meta, active, width, height, lines)
	}
	inner := sectionBodyRows(height, lines)
	start, end := listWindow(cur, total, inner)
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
		lines = append(
			lines,
			viewchrome.RenderPaneRowStyled(
				theme,
				i,
				cur,
				active,
				fmt.Sprintf("%s %s%s", status, habit.Name, duration),
				style,
				width,
			),
		)
	}
	if remaining := total - end; remaining > 0 {
		lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("↓ %d more", remaining)))
	}
	return dailySection(theme, "Habits Due", meta, active, width, height, lines)
}

// listWindow picks the visible items for budget rows, reserving a row for each ↑/↓ marker.
func listWindow(cur, total, budget int) (int, int) {
	start, end := viewchrome.ListWindow(cur, total, max(1, budget))
	for range 2 {
		slots := budget
		if start > 0 {
			slots--
		}
		if end < total {
			slots--
		}
		start, end = viewchrome.ListWindow(cur, total, max(1, slots))
	}
	return start, end
}

// sectionBodyRows is the list budget left in a section of height rows after its header and lines.
func sectionBodyRows(height int, lines []string) int {
	return max(1, height-1-viewchrome.RenderedLineCount(lines))
}

func dailySection(
	theme types.Theme,
	title, meta string,
	focused bool,
	width, height int,
	lines []string,
) string {
	body := make([]string, 0, len(lines))
	for _, line := range lines {
		body = append(body, strings.Split(line, "\n")...)
	}
	return viewchrome.RenderSection(theme, viewchrome.Section{
		Title:   title,
		Meta:    meta,
		Focused: focused,
		Width:   width,
		Height:  height,
		Lines:   body,
	})
}

func renderIssueStatusBar(theme types.Theme, counts map[string]int, width int) string {
	width = max(8, width)
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
		segments = append(
			segments,
			lipgloss.NewStyle().Foreground(item.color).Render(strings.Repeat("█", segmentWidth)),
		)
		used += segmentWidth
		remainingStatuses--
	}
	if used < width {
		segments = append(segments, theme.StyleDim.Render(strings.Repeat("█", width-used)))
	}
	return strings.Join(segments, "")
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
	return viewhelpers.RenderGradientBar(
		completedWidth,
		completedWidth,
		viewhelpers.GradientBarPalette{
			Start: theme.ColorDullGreen,
			End:   theme.ColorCyan,
			Track: theme.ColorDim,
		},
	) + viewhelpers.RenderGradientBar(
		failedWidth,
		failedWidth,
		viewhelpers.GradientBarPalette{
			Start: theme.ColorDullRed,
			End:   theme.ColorOrange,
			Track: theme.ColorDim,
		},
	) + viewhelpers.RenderGradientBar(
		remainingWidth,
		0,
		viewhelpers.GradientBarPalette{
			Start: theme.ColorDim,
			End:   theme.ColorDim,
			Track: theme.ColorDim,
		},
	)
}
