package daily

import (
	"strings"

	sharedtypes "crona.work/shared/types"
	"crona.work/tui/internal/api"
	helperpkg "crona.work/tui/internal/tui/helpers"
	viewchrome "crona.work/tui/internal/tui/views/chrome"
	viewhelpers "crona.work/tui/internal/tui/views/helpers"
	issuecore "crona.work/tui/internal/tui/views/issuecore"
	momentumhelpers "crona.work/tui/internal/tui/views/momentum"
	sessionmeta "crona.work/tui/internal/tui/views/sessionmeta"
	viewsummary "crona.work/tui/internal/tui/views/summary"
	types "crona.work/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// renderDetail shows the focused pane's selection, titled with its name.
func renderDetail(theme types.Theme, state types.ContentState, width, height int) string {
	bodyWidth := max(1, width-2)
	title, lines := detailBody(theme, state, bodyWidth)
	meta := viewchrome.KeyHint(theme, "enter", "details")
	scroll := 0
	if state.DailyDetailFocus {
		meta = viewchrome.KeyHint(theme, "esc", "back")
		scroll = state.DailyDetailScroll
	}
	// Scrolled, the body starts lower; arrows in the title say there is more either way,
	// focused or not.
	visible := max(1, height-1)
	scroll = min(scroll, max(0, len(lines)-visible))
	if scroll > 0 {
		lines = lines[scroll:]
	}
	if more := len(lines) > visible; more || scroll > 0 {
		meta = theme.StyleDim.Render(scrollArrows(scroll > 0, more)) + "  " + meta
	}
	// A long title gives way so the key always fits: lead, padded title, a fill cell, meta.
	title = ansi.Truncate(title, max(1, width-5-ansi.StringWidth(meta)), "…")
	titleStyle := theme.StyleHeader.Bold(true)
	return viewchrome.RenderSection(theme, viewchrome.Section{
		Title:      title,
		Meta:       meta,
		TitleStyle: &titleStyle,
		Width:      width,
		Height:     height,
		Lines:      lines,
		Focused:    state.DailyDetailFocus,
		// The pane takes focus (tab or enter), so it fades like the lists without it.
		Dimmed: !state.DailyDetailFocus,
	})
}

// detailBody is the Day summary, or with the pane focused (enter on a task or habit) the
// selection's details.
func detailBody(theme types.Theme, state types.ContentState, width int) (string, []string) {
	if !state.DailyDetailFocus || state.DailyDetailSummary {
		return "Day", padded(viewsummary.DayLines(theme, state, max(1, width-2*panePad)))
	}
	if state.Pane == "habits" {
		if habit, ok := selectedDailyHabit(state); ok {
			return habit.Name, padded(habitDetailLines(theme, state, habit, max(1, width-2*panePad)))
		}
	} else if issue, ok := selectedDailyIssue(state); ok {
		return issue.Title, padded(IssueDetailLines(theme, state, issue, max(1, width-2*panePad)))
	}
	return "Selected", []string{theme.StyleDim.Render("Nothing selected")}
}

// DetailScrollLimit is the furthest the focused detail pane scrolls: its last line at the top.
func DetailScrollLimit(theme types.Theme, state types.ContentState) int {
	width := viewhelpers.Columns(state.Width, 3, 2)[1]
	_, lines := detailBody(theme, state, max(1, width-2))
	return max(0, len(lines)-1)
}

func scrollArrows(up, down bool) string {
	switch {
	case up && down:
		return "↑↓"
	case up:
		return "↑"
	default:
		return "↓"
	}
}

func selectedDailyIssue(state types.ContentState) (api.Issue, bool) {
	issues := filteredDailyIssues(state)
	indices := dailyIssueIndices(state, dailyAPIIssues(issues))
	cur := state.Cursors["issues"]
	if cur < 0 || cur >= len(indices) || indices[cur] >= len(issues) {
		return api.Issue{}, false
	}
	return issues[indices[cur]], true
}

func selectedDailyHabit(state types.ContentState) (api.HabitDailyItem, bool) {
	indices := dailyHabitIndices(state)
	cur := state.Cursors["habits"]
	if cur < 0 || cur >= len(indices) || indices[cur] >= len(state.DueHabits) {
		return api.HabitDailyItem{}, false
	}
	return state.DueHabits[indices[cur]], true
}

// IssueDetailLines renders the shared issue detail body.
func IssueDetailLines(
	theme types.Theme,
	state types.ContentState,
	issue api.Issue,
	width int,
) []string {
	repo, stream, worked := issue.RepoName, issue.StreamName, issue.WorkedSeconds
	var repoID int64
	if meta := sessionmeta.IssueMetaByID(state.AllIssues, issue.ID); meta != nil {
		repo, stream, repoID = meta.RepoName, meta.StreamName, meta.RepoID
		if worked <= 0 {
			worked = meta.WorkedSeconds
		}
	}
	statusStyle := theme.StyleNormal
	if style := issuecore.IssueStatusStyle(theme, DisplayIssueStatus(state, issue.ID, issue.Status)); style != nil {
		statusStyle = *style
	}
	estimate := 0
	if issue.EstimateMinutes != nil {
		estimate = *issue.EstimateMinutes
	}
	status := DisplayIssueStatus(state, issue.ID, issue.Status)
	// The fact column: status with its date, context, tags, then the penalty's verdict.
	head := statusStyle.Render("● " + issuecore.PlainIssueStatus(status))
	if date := issuecore.IssueDateLabel(issue.Status, issue.TodoForDate, issue.CompletedAt, issue.AbandonedAt, state.Settings); date != "" {
		head += theme.StyleDim.Render(" · ") + theme.StyleNormal.Render(date)
	}
	if issue.PinnedDaily {
		head += theme.StyleDim.Render(" · pinned")
	}
	facts := []string{head, theme.StyleNormal.Render(detailContext(repo, stream))}
	if len(issue.Tags) > 0 {
		facts = append(facts, issuecore.TagChips(theme, issue.Tags, state.Tags, max(8, width-viewsummary.ArcsWidth(2)-3)))
	}
	penalty := issuePenaltyView(theme, state, issue.ID, width)
	if penalty.verdict != "" {
		facts = append(facts, penalty.verdict)
	}
	arcs := []viewsummary.Arc{effortArc(theme, worked, estimate)}
	if penalty.arc != nil {
		arcs = append(arcs, *penalty.arc)
	}
	arcWidth := viewsummary.ArcsWidth(len(arcs))
	lines := detailDescription(theme, issue.Description, width)
	lines = append(lines, viewsummary.BesideArcs(viewsummary.Arcs(theme, arcs, arcWidth), facts, arcWidth, width)...)
	lines = append(lines, penalty.lines...)
	lines = append(lines, issueSessionLines(theme, state, issue.ID, width)...)
	lines = append(lines, carryLines(theme, state, issue.ID, width)...)
	lines = append(lines, issuecore.NotesLines(theme, state.Settings, issue.Notes, width)...)
	return append(
		lines,
		momentumLines(theme, state, issueMomentums(state, repoID, issue.StreamID), width)...)
}

// effortArc fills with worked time against the estimate: cyan on the way, green at the
// estimate and orange past it. Without an estimate it shows the time worked on a dim track.
func effortArc(theme types.Theme, workedSeconds, estimateMinutes int) viewsummary.Arc {
	arc := viewsummary.Arc{Label: "Effort", Value: helperpkg.FormatCompactDurationSeconds(workedSeconds), Color: theme.ColorCyan}
	if estimateMinutes <= 0 {
		return arc
	}
	arc.Label += "/" + helperpkg.FormatCompactDurationMinutes(estimateMinutes)
	arc.Fraction = float64(workedSeconds) / float64(estimateMinutes*60)
	switch {
	case arc.Fraction > 1.05:
		arc.Color = theme.ColorOrange
	case arc.Fraction >= 1:
		arc.Color = theme.ColorGreen
	}
	return arc
}

func habitDetailLines(
	theme types.Theme,
	state types.ContentState,
	habit api.HabitDailyItem,
	width int,
) []string {
	status := "due"
	switch habit.Status {
	case "completed":
		status = "completed"
	case "failed":
		status = "failed"
	}
	return habitDetailWithStatus(theme, state, habit, width, status)
}

func habitDetailWithStatus(
	theme types.Theme,
	state types.ContentState,
	habit api.HabitDailyItem,
	width int,
	status string,
) []string {
	return habitDetailView(theme, state, habit, width, status)
}

// detailDescription leads the pane with the description, followed by a blank row.
func detailDescription(theme types.Theme, text *string, width int) []string {
	if text == nil || strings.TrimSpace(*text) == "" {
		return []string{}
	}
	return append(wrapDetail(theme.StyleNormal, strings.TrimSpace(*text), width), "")
}

// habitMomentums are the enabled habit momentums that count this habit.
func habitMomentums(state types.ContentState, habitID int64) []sharedtypes.MomentumCard {
	return momentumhelpers.ForHabit(state.MomentumCards, habitID)
}

// issueMomentums are the enabled context momentums that sessions on this issue count toward.
func issueMomentums(state types.ContentState, repoID, streamID int64) []sharedtypes.MomentumCard {
	return momentumhelpers.ForContext(state.MomentumCards, repoID, streamID)
}

// momentumLines lists momentums one per line: cadence, the current period's progress (time
// for context momentums), the ladder and the streak.
func momentumLines(theme types.Theme, state types.ContentState, cards []sharedtypes.MomentumCard, width int) []string {
	if len(cards) == 0 {
		return nil
	}
	lines := []string{"", theme.StyleDim.Bold(true).Render("Momentum")}
	for _, row := range viewsummary.MomentumCardRows(theme, cards, dailyDate(state), max(1, width-2)) {
		lines = append(lines, "  "+row)
	}
	return lines
}

func detailContext(repo, stream string) string {
	repo, stream = strings.TrimSpace(repo), strings.TrimSpace(stream)
	switch {
	case repo != "" && stream != "":
		return repo + " › " + stream
	case repo != "":
		return repo
	case stream != "":
		return stream
	default:
		return "-"
	}
}

func detailText(theme types.Theme, title string, text *string, width int) []string {
	if text == nil || strings.TrimSpace(*text) == "" {
		return nil
	}
	lines := []string{"", theme.StyleDim.Bold(true).Render(title)}
	return append(lines, wrapDetail(theme.StyleNormal, strings.TrimSpace(*text), width)...)
}

func wrapDetail(style lipgloss.Style, text string, width int) []string {
	wrapped := lipgloss.NewStyle().Width(max(8, width)).Render(text)
	lines := strings.Split(wrapped, "\n")
	for i, line := range lines {
		lines[i] = style.Render(strings.TrimRight(line, " "))
	}
	return lines
}

// HabitDetailLines reuses the daily detail fields for a habit definition.
func HabitDetailLines(
	theme types.Theme,
	state types.ContentState,
	habit api.HabitWithMeta,
	width int,
) []string {
	for _, due := range state.DueHabits {
		if due.ID == habit.ID {
			return habitDetailLines(theme, state, due, width)
		}
	}
	status := "not due today"
	if !habit.Active {
		status = "paused"
	}
	return habitDetailWithStatus(
		theme,
		state,
		api.HabitDailyItem{HabitWithMeta: habit},
		width,
		status,
	)
}

// IssueInsightLines are the delay penalty, carry-over and momentum sections of the issue
// detail pane, for the issue detail dialog to show the same.
func IssueInsightLines(
	theme types.Theme,
	state types.ContentState,
	issueID, repoID, streamID int64,
	width int,
) []string {
	lines := penaltyLines(theme, state, issueID, width)
	lines = append(lines, carryLines(theme, state, issueID, width)...)
	return append(lines, momentumLines(theme, state, issueMomentums(state, repoID, streamID), width)...)
}

// HabitMomentumLines are the habit detail pane's momentum section, ladders included.
func HabitMomentumLines(theme types.Theme, state types.ContentState, habitID int64, width int) []string {
	return momentumLines(theme, state, habitMomentums(state, habitID), width)
}
