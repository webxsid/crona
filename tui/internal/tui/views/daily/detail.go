package daily

import (
	"fmt"
	"slices"
	"strings"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	helperpkg "crona/tui/internal/tui/helpers"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	issuecore "crona/tui/internal/tui/views/issuecore"
	momentumhelpers "crona/tui/internal/tui/views/momentum"
	sessionmeta "crona/tui/internal/tui/views/sessionmeta"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
)

// renderDetail shows the focused pane's selection.
func renderDetail(theme types.Theme, state types.ContentState, width, height int) string {
	bodyWidth := max(1, width-2)
	var lines []string
	if state.Pane == "habits" {
		if habit, ok := selectedDailyHabit(state); ok {
			lines = habitDetailLines(theme, state, habit, bodyWidth)
		}
	} else if issue, ok := selectedDailyIssue(state); ok {
		lines = issueDetailLines(theme, state, issue, bodyWidth)
	}
	if len(lines) == 0 {
		lines = []string{theme.StyleDim.Render("Nothing selected")}
	}
	return dailySection(theme, "Selected", "", false, width, height, lines)
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

func issueDetailLines(theme types.Theme, state types.ContentState, issue api.Issue, width int) []string {
	repo, stream, worked := issue.RepoName, issue.StreamName, issue.WorkedSeconds
	if meta := sessionmeta.IssueMetaByID(state.AllIssues, issue.ID); meta != nil {
		repo, stream = meta.RepoName, meta.StreamName
		if worked <= 0 {
			worked = meta.WorkedSeconds
		}
	}
	statusStyle := theme.StyleNormal
	if style := issuecore.IssueStatusStyle(theme, string(issue.Status)); style != nil {
		statusStyle = *style
	}
	estimate := 0
	if issue.EstimateMinutes != nil {
		estimate = *issue.EstimateMinutes
	}
	lines := wrapDetail(theme.StyleHeader.Bold(true), issue.Title, width)
	lines = append(lines,
		detailField(theme, "status", statusStyle.Render(issuecore.PlainIssueStatus(string(issue.Status)))),
		detailField(theme, "context", theme.StyleNormal.Render(detailContext(repo, stream))),
		detailField(theme, "effort", theme.StyleNormal.Render(issueProgressText(worked, estimate, true))),
	)
	if date := issuecore.IssueDateLabel(issue.Status, issue.TodoForDate, issue.CompletedAt, issue.AbandonedAt, state.Settings); date != "" {
		lines = append(lines, detailField(theme, "date", theme.StyleNormal.Render(date)))
	}
	if issue.PinnedDaily {
		lines = append(lines, detailField(theme, "pinned", theme.StyleNormal.Render("yes")))
	}
	lines = append(lines, detailText(theme, "Description", issue.Description, width)...)
	return append(lines, detailText(theme, "Notes", issue.Notes, width)...)
}

func habitDetailLines(theme types.Theme, state types.ContentState, habit api.HabitDailyItem, width int) []string {
	status := "due"
	switch habit.Status {
	case "completed":
		status = "completed"
	case "failed":
		status = "failed"
	}
	lines := wrapDetail(theme.StyleHeader.Bold(true), habit.Name, width)
	lines = append(lines, detailField(theme, "status", theme.StyleNormal.Render(status)))
	if habit.DurationMinutes != nil {
		lines = append(lines, detailField(theme, "logged", theme.StyleNormal.Render(helperpkg.FormatCompactDurationMinutes(*habit.DurationMinutes))))
	}
	if habit.TargetMinutes != nil {
		lines = append(lines, detailField(theme, "target", theme.StyleNormal.Render(helperpkg.FormatCompactDurationMinutes(*habit.TargetMinutes))))
	}
	if meta := strings.TrimSpace(detailContext(habit.RepoName, habit.StreamName)); meta != "-" {
		lines = append(lines, detailField(theme, "context", theme.StyleNormal.Render(meta)))
	}
	lines = append(lines, habitMomentumLines(theme, state, habit.ID, width)...)
	lines = append(lines, detailText(theme, "Description", habit.Description, width)...)
	return append(lines, detailText(theme, "Notes", habit.Notes, width)...)
}

// habitMomentumLines lists the enabled habit momentums this habit feeds, one per line:
// cadence, the current period's count against its target, the ladder and the streak.
func habitMomentumLines(theme types.Theme, state types.ContentState, habitID int64, width int) []string {
	cards := []sharedtypes.MomentumCard{}
	nameWidth := 0
	for _, card := range state.MomentumCards {
		def := sharedtypes.NormalizeHabitStreakDefinition(card.Definition)
		if def.Enabled && def.TargetKind == sharedtypes.MomentumTargetKindHabit && slices.Contains(def.HabitIDs, habitID) {
			card.Definition = def
			cards = append(cards, card)
			nameWidth = max(nameWidth, lipgloss.Width(def.Name))
		}
	}
	lines := []string{}
	for _, card := range cards {
		def := card.Definition
		unit := momentumhelpers.Unit(def.Period)
		parts := []string{
			theme.StyleNormal.Render(padCells(def.Name, min(nameWidth, max(8, width/3)))),
			theme.StyleDim.Render(padCells(momentumhelpers.CadenceLabel(def.Period), 7)),
		}
		if len(card.Series) > 0 {
			point := card.Series[len(card.Series)-1]
			parts = append(parts, theme.StyleNormal.Render(padCells(fmt.Sprintf("%d/%d", point.Count, point.Target), 5)))
		}
		parts = append(parts,
			momentumhelpers.Ladder(theme, def.Period, card.Current, card.Longest),
			theme.StyleNormal.Render(momentumhelpers.FormatLength(card.Current, unit))+
				theme.StyleDim.Render(" best "+momentumhelpers.FormatLength(card.Longest, unit)),
		)
		lines = append(lines, viewhelpers.TruncateANSI(strings.Join(parts, "  "), width))
	}
	if len(lines) == 0 {
		return nil
	}
	return append([]string{"", theme.StyleDim.Bold(true).Render("Momentum")}, lines...)
}

func detailField(theme types.Theme, label, value string) string {
	return theme.StyleDim.Render(padCells(label, 9)) + value
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
