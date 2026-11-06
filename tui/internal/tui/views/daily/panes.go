package daily

import (
	uistate "crona/tui/internal/tui/state"
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
	"github.com/charmbracelet/x/ansi"
)

func renderIssues(theme types.Theme, state types.ContentState, width, height int) string {
	active := state.Pane == "issues" && !state.DailyDetailFocus
	cur := state.Cursors["issues"]
	sources := filteredDailyIssues(state)
	issues := dailyAPIIssues(sources)
	indices := dailyIssueIndices(state, issues)
	total := len(indices)
	base := viewui.PaneBase{Focused: active, Width: width, Height: height, Cursor: cur}
	meta := fmt.Sprintf("%d  [←/→]", total)
	lines := dailyTaskHeader(theme, state, width, height)
	if len(issues) == 0 || total == 0 {
		lines = append(lines, theme.StyleDim.Render(dailyTaskEmptyMessage(state.DailyTaskSection)))
		return dailySection(theme, "Tasks", viewhelpers.FilterMeta(meta, state.Filters["issues"]), active, width, height, lines)
	}
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
	return dailySection(theme, "Tasks", viewhelpers.FilterMeta(meta, state.Filters["issues"]), active, width, height, lines)
}

// dailyTaskHeader is the section selector over the new-task line; in a short pane the
// new-task keys join the selector's row so a card and its actions still fit.
func dailyTaskHeader(theme types.Theme, state types.ContentState, width, height int) []string {
	selector := dailyTaskTitle(theme, state.DailyTaskSection)
	if sectionBodyRows(height, []string{selector, ""}) >= 5 {
		return []string{selector, newTaskLine(theme, state, width-2)}
	}
	room := width - 2 - ansi.StringWidth(selector) - 3
	if keys := newTaskLine(theme, state, room); keys != "" {
		selector += "   " + keys
	}
	return []string{selector}
}

func dailyAPIIssues(sources []api.Issue) []issuecore.APIIssue {
	issues := make([]issuecore.APIIssue, 0, len(sources))
	for _, issue := range sources {
		converted := issuecore.NewAPIIssue(
			issue.ID,
			issue.Title,
			issue.Status,
			issue.EstimateMinutes,
			issue.WorkedSeconds,
			issue.PinnedDaily,
			issue.TodoForDate,
			issue.CompletedAt,
			issue.AbandonedAt,
		)
		converted.Tags = issue.Tags
		issues = append(issues, converted)
	}
	return issues
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
	// On Planned, open pinned tasks lead under their own heading; the heading of the first
	// visible group takes a row, later headings take the blank row between items.
	pinnedCount := dailyPinnedCount(state, sources)
	if pinnedCount > 0 {
		availableRows--
	}
	groupOf := func(rawIdx int) int {
		if rawIdx < pinnedCount {
			return 0
		}
		return 1
	}
	groupSize := [2]int{}
	for _, rawIdx := range indices {
		groupSize[groupOf(rawIdx)]++
	}
	heading := func(group int) string {
		name := [2]string{"Pinned", "Planned"}[group]
		return theme.StyleDim.Bold(true).Render(fmt.Sprintf("%s %d", name, groupSize[group]))
	}
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
		switch {
		case pinnedCount > 0 && idx == 0:
			lines = append(lines, heading(groupOf(rawIdx)))
		case pinnedCount > 0 && groupOf(rawIdx) != groupOf(indices[start+idx-1]):
			lines = append(lines, heading(groupOf(rawIdx)))
		case idx > 0:
			lines = append(lines, "")
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
		if marker := PenaltyMarkerText(state, issue.ID); marker != "" {
			title += "  " + marker
		}
		if marker := CarryMarkerText(state, issue.ID); marker != "" {
			title += "  " + marker
		}
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
				DisplayIssueStatus(state, issue.ID, issue.Status),
			),
		)
		if start+idx == cur && active {
			lines = append(lines, taskActionLine(theme, state, width))
			continue
		}
		lines = append(lines, renderCompactDailyIssueMeta(
			theme,
			width,
			start+idx,
			cur,
			active,
			repoName,
			streamName,
			workedSeconds,
			issue.EstimateMinutes,
			DisplayIssueStatus(state, issue.ID, issue.Status),
			issuecore.TagChips(theme, issue.Tags, state.Tags, max(0, width/2)),
		))
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
	// A long title gives way before its date does.
	room := max(18, width-4)
	if date != "" {
		field := "  " + dateField(date)
		title = ansi.Truncate(title, max(8, room-ansi.StringWidth(field)), "…") + field
	}
	line := viewhelpers.TruncateANSI(title, room)
	line = rowStyle.Render(line)
	if i == cur && active {
		return viewchrome.SelectedRow(theme, line)
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
	chips string,
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
	if chips == "" {
		return viewhelpers.TruncateANSI(meta, width)
	}
	room := max(0, width-ansi.StringWidth(chips)-2)
	return viewhelpers.TruncateANSI(meta, room) + "  " + chips
}

func compactDailyIssueDate(state types.ContentState, issue issuecore.APIIssue) string {
	// On Planned only pinned tasks can sit on another day, so only they carry a date.
	if state.DailyTaskSection == "planned" &&
		(issue.TodoForDate == nil || strings.TrimSpace(*issue.TodoForDate) == dailyDate(state)) {
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

func dailyTaskTitle(theme types.Theme, section string) string {
	options := []struct {
		key   string
		label string
	}{
		{key: "planned", label: "Planned"},
		{key: "overdue", label: "Overdue"},
		{key: "upcoming", label: "Upcoming"},
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
	case "upcoming":
		return "Nothing due in the next 7 days"
	case "overdue":
		return "No overdue tasks in this scope"
	default:
		return "No planned tasks for this date"
	}
}

// dailyPinnedCount is how many leading tasks on Planned are open pinned ones; the model
// orders them first.
func dailyPinnedCount(state types.ContentState, sources []api.Issue) int {
	if state.DailyTaskSection != "planned" {
		return 0
	}
	count := 0
	for _, issue := range sources {
		if !issue.PinnedDaily || issue.Status == "done" || issue.Status == "abandoned" {
			break
		}
		count++
	}
	return count
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
	active := state.Pane == "habits" && !state.DailyDetailFocus
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
	lines := []string{newItemLine(theme, "new habit")}
	if total == 0 {
		lines = append(lines, theme.StyleDim.Render("No due habits for this date"))
		return dailySection(theme, "Habits Due", viewhelpers.FilterMeta(meta, state.Filters["habits"])+"  "+foldHint(theme, uistate.CollapseDailyHabits, "fold"), active, width, height, lines)
	}
	// Like tasks, each habit is a title row and a meta row, with a blank row between habits;
	// the selected habit's action line takes its meta row.
	inner := sectionBodyRows(height, lines)
	slots := func(rows int) int { return max(1, (rows+1)/3) }
	start, end := viewchrome.ListWindow(cur, total, slots(inner))
	rows := inner
	if start > 0 {
		rows--
	}
	if end < total {
		rows--
	}
	start, end = viewchrome.ListWindow(cur, total, slots(rows))
	if start > 0 {
		lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("↑ %d more", start)))
	}
	nameWidth := 0
	for _, idx := range indices {
		nameWidth = max(nameWidth, ansi.StringWidth(state.DueHabits[idx].Name))
	}
	for i := start; i < end; i++ {
		if i > start {
			lines = append(lines, "")
		}
		habit := state.DueHabits[indices[i]]
		lines = append(lines, habitRow(theme, habit, nameWidth, width-1, i == cur, active))
		if i == cur && active {
			lines = append(lines, habitActionLine(theme, state, width-1))
			continue
		}
		lines = append(lines, habitMetaRow(theme, state, habit, width-1, i == cur))
	}
	if remaining := total - end; remaining > 0 {
		lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("↓ %d more", remaining)))
	}
	return dailySection(theme, "Habits Due", viewhelpers.FilterMeta(meta, state.Filters["habits"])+"  "+foldHint(theme, uistate.CollapseDailyHabits, "fold"), active, width, height, lines)
}

const habitRowBarWidth = 10

// habitRow is a habit's title row: checkbox and name, a bar of logged minutes against the
// target, and the minutes. The bar gives way when narrow.
func habitRow(
	theme types.Theme,
	habit api.HabitDailyItem,
	nameWidth, width int,
	selected, active bool,
) string {
	mark, style := habitMark(theme, habit)
	cursor := "  "
	if selected && active {
		cursor = theme.StyleCursor.Render(viewchrome.SelectionCursor + " ")
		style = style.Bold(true)
	}
	name := style.Render(mark + " " + padCells(habit.Name, min(nameWidth, max(8, width/2))))
	minutes := theme.StyleDim.Render(habitMinutesText(habit))
	// The status sits on the meta row, so the title row is name, meter and minutes.
	for _, parts := range [][]string{
		{name, habitMeter(theme, habit), minutes},
		{name, minutes},
	} {
		if line := cursor + strings.Join(parts, "  "); ansi.StringWidth(line) <= width {
			return line
		}
	}
	return viewhelpers.TruncateANSI(cursor+name, width)
}

// habitMetaRow is a habit's second row, like a task's: its repo and stream, schedule and
// status, with its tags at the end.
func habitMetaRow(theme types.Theme, state types.ContentState, habit api.HabitDailyItem, width int, selected bool) string {
	status := "due"
	switch habit.Status {
	case "completed":
		status = "done"
	case "failed":
		status = "failed"
	}
	parts := []string{issuecore.IssueContextLabel(habit.RepoName, habit.StreamName)}
	if schedule := strings.ReplaceAll(string(habit.ScheduleType), "_", " "); schedule != "" {
		parts = append(parts, schedule)
	}
	meta := "    " + strings.Join(append(parts, status), " | ")
	if selected {
		meta = theme.StyleSelected.Render(meta)
	} else {
		meta = theme.StyleDim.Render(meta)
	}
	chips := issuecore.TagChips(theme, habit.Tags, state.Tags, max(0, width/2))
	if chips == "" {
		return viewhelpers.TruncateANSI(meta, width)
	}
	return viewhelpers.TruncateANSI(meta, max(0, width-ansi.StringWidth(chips)-2)) + "  " + chips
}

func habitMinutesText(habit api.HabitDailyItem) string {
	logged := "-"
	if habit.DurationMinutes != nil {
		logged = helperpkg.FormatCompactDurationMinutes(*habit.DurationMinutes)
	}
	if habit.TargetMinutes == nil {
		if logged == "-" {
			return ""
		}
		return logged
	}
	return logged + " / " + helperpkg.FormatCompactDurationMinutes(*habit.TargetMinutes)
}

// habitMeter fills with logged minutes against the target; a completed habit without a
// target reads as full.
func habitMeter(theme types.Theme, habit api.HabitDailyItem) string {
	filled := 0
	switch {
	case habit.TargetMinutes != nil && *habit.TargetMinutes > 0 && habit.DurationMinutes != nil:
		filled = min(
			habitRowBarWidth,
			*habit.DurationMinutes*habitRowBarWidth / *habit.TargetMinutes,
		)
	case habit.Status == "completed":
		filled = habitRowBarWidth
	}
	color := theme.ColorYellow
	switch habit.Status {
	case "completed":
		color = theme.ColorGreen
	case "failed":
		color = theme.ColorRed
	}
	return lipgloss.NewStyle().Foreground(color).Render(strings.Repeat("█", filled)) +
		theme.StyleDim.Render(strings.Repeat("░", habitRowBarWidth-filled))
}

// habitMark is a habit's checkbox and the colour its row is drawn in.
func habitMark(theme types.Theme, habit api.HabitDailyItem) (string, lipgloss.Style) {
	switch habit.Status {
	case "completed":
		return "[x]", lipgloss.NewStyle().Foreground(theme.ColorGreen)
	case "failed":
		return "[!]", lipgloss.NewStyle().Foreground(theme.ColorRed)
	default:
		return "[ ]", theme.StyleNormal
	}
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
		Dimmed:  !focused,
		Width:   width,
		Height:  height,
		Lines:   body,
	})
}
