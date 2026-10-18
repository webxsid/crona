package daily

import (
	"fmt"
	"slices"
	"strings"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	helperpkg "crona/tui/internal/tui/helpers"
	viewchrome "crona/tui/internal/tui/views/chrome"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	issuecore "crona/tui/internal/tui/views/issuecore"
	momentumhelpers "crona/tui/internal/tui/views/momentum"
	sessionmeta "crona/tui/internal/tui/views/sessionmeta"
	viewsummary "crona/tui/internal/tui/views/summary"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// renderDetail shows the focused pane's selection, titled with its name.
func renderDetail(theme types.Theme, state types.ContentState, width, height int) string {
	bodyWidth := max(1, width-2)
	title := "Selected"
	var lines []string
	if state.DailyDetailDay {
		title, lines = "Day", viewsummary.DayLines(theme, state, bodyWidth)
	} else if state.Pane == "habits" {
		if habit, ok := selectedDailyHabit(state); ok {
			title, lines = habit.Name, habitDetailLines(theme, state, habit, bodyWidth)
		}
	} else if issue, ok := selectedDailyIssue(state); ok {
		title, lines = issue.Title, IssueDetailLines(theme, state, issue, bodyWidth)
	}
	if lines == nil {
		lines = []string{theme.StyleDim.Render("Nothing selected")}
	}
	dayKey := viewchrome.KeyHint(theme, "S", "day summary")
	if state.DailyDetailDay {
		dayKey = viewchrome.KeyHint(theme, "S", "show selection")
	}
	// A long title gives way so the key always fits: lead, padded title, a fill cell, meta.
	title = ansi.Truncate(title, max(1, width-5-ansi.StringWidth(dayKey)), "…")
	titleStyle := theme.StyleHeader.Bold(true)
	return viewchrome.RenderSection(theme, viewchrome.Section{
		Title:      title,
		Meta:       dayKey,
		TitleStyle: &titleStyle,
		Width:      width,
		Height:     height,
		Lines:      lines,
	})
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

// detailEntry is one labelled field in the detail grid.
type detailEntry struct{ label, value string }

// IssueDetailLines renders the shared issue detail body.
func IssueDetailLines(theme types.Theme, state types.ContentState, issue api.Issue, width int) []string {
	repo, stream, worked := issue.RepoName, issue.StreamName, issue.WorkedSeconds
	var repoID int64
	if meta := sessionmeta.IssueMetaByID(state.AllIssues, issue.ID); meta != nil {
		repo, stream, repoID = meta.RepoName, meta.StreamName, meta.RepoID
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
	fields := []detailEntry{
		{"status", statusStyle.Render(issuecore.PlainIssueStatus(string(issue.Status)))},
		{"context", theme.StyleNormal.Render(detailContext(repo, stream))},
		{"effort", theme.StyleNormal.Render(issueProgressText(worked, estimate, true))},
	}
	if date := issuecore.IssueDateLabel(issue.Status, issue.TodoForDate, issue.CompletedAt, issue.AbandonedAt, state.Settings); date != "" {
		fields = append(fields, detailEntry{"date", theme.StyleNormal.Render(date)})
	}
	if issue.PinnedDaily {
		fields = append(fields, detailEntry{"pinned", theme.StyleNormal.Render("yes")})
	}
	lines := detailDescription(theme, issue.Description, width)
	lines = append(lines, detailGrid(theme, fields, width)...)
	lines = append(lines, issuecore.NotesLines(theme, state.Settings, issue.Notes, width)...)
	return append(lines, momentumLines(theme, issueMomentums(state, repoID, issue.StreamID), width)...)
}

func habitDetailLines(theme types.Theme, state types.ContentState, habit api.HabitDailyItem, width int) []string {
	status := "due"
	switch habit.Status {
	case "completed":
		status = "completed"
	case "failed":
		status = "failed"
	}
	fields := []detailEntry{{"status", theme.StyleNormal.Render(status)}}
	if habit.DurationMinutes != nil {
		fields = append(fields, detailEntry{"logged", theme.StyleNormal.Render(helperpkg.FormatCompactDurationMinutes(*habit.DurationMinutes))})
	}
	if habit.TargetMinutes != nil {
		fields = append(fields, detailEntry{"target", theme.StyleNormal.Render(helperpkg.FormatCompactDurationMinutes(*habit.TargetMinutes))})
	}
	if meta := strings.TrimSpace(detailContext(habit.RepoName, habit.StreamName)); meta != "-" {
		fields = append(fields, detailEntry{"context", theme.StyleNormal.Render(meta)})
	}
	lines := detailDescription(theme, habit.Description, width)
	lines = append(lines, detailGrid(theme, fields, width)...)
	lines = append(lines, detailText(theme, "Notes", habit.Notes, width)...)
	return append(lines, momentumLines(theme, habitMomentums(state, habit.ID), width)...)
}

// detailDescription leads the pane with the description, followed by a blank row.
func detailDescription(theme types.Theme, text *string, width int) []string {
	if text == nil || strings.TrimSpace(*text) == "" {
		return []string{}
	}
	return append(wrapDetail(theme.StyleNormal, strings.TrimSpace(*text), width), "")
}

// detailGrid flows fields row by row into as many equal columns (up to 3) as fit width.
func detailGrid(theme types.Theme, fields []detailEntry, width int) []string {
	const gap, labelWidth = 3, 9
	cell := 0
	for _, field := range fields {
		cell = max(cell, labelWidth+ansi.StringWidth(field.value))
	}
	cols := min(3, len(fields), max(1, (width+gap)/(cell+gap)))
	lines := []string{}
	for start := 0; start < len(fields); start += cols {
		cells := []string{}
		for _, field := range fields[start:min(len(fields), start+cols)] {
			cells = append(cells, padCells(theme.StyleDim.Render(padCells(field.label, labelWidth))+field.value, cell))
		}
		lines = append(lines, strings.TrimRight(strings.Join(cells, strings.Repeat(" ", gap)), " "))
	}
	return lines
}

// habitMomentums are the enabled habit momentums that count this habit.
func habitMomentums(state types.ContentState, habitID int64) []sharedtypes.MomentumCard {
	return enabledMomentums(state, func(def sharedtypes.HabitStreakDefinition) bool {
		return def.TargetKind == sharedtypes.MomentumTargetKindHabit && slices.Contains(def.HabitIDs, habitID)
	})
}

// issueMomentums are the enabled context momentums that sessions on this issue count toward:
// a context matches the issue's repo and either its stream or the whole repo, as the daemon
// counts them.
func issueMomentums(state types.ContentState, repoID, streamID int64) []sharedtypes.MomentumCard {
	if repoID == 0 {
		return nil
	}
	return enabledMomentums(state, func(def sharedtypes.HabitStreakDefinition) bool {
		if def.TargetKind != sharedtypes.MomentumTargetKindContext {
			return false
		}
		return slices.ContainsFunc(def.Contexts, func(context sharedtypes.MomentumContext) bool {
			return context.RepoID == repoID && (context.StreamID == nil || *context.StreamID == streamID)
		})
	})
}

func enabledMomentums(state types.ContentState, matches func(sharedtypes.HabitStreakDefinition) bool) []sharedtypes.MomentumCard {
	out := []sharedtypes.MomentumCard{}
	for _, card := range state.MomentumCards {
		def := sharedtypes.NormalizeHabitStreakDefinition(card.Definition)
		if def.Enabled && matches(def) {
			card.Definition = def
			out = append(out, card)
		}
	}
	return out
}

// momentumLines lists momentums one per line: cadence, the current period's progress (time
// for context momentums), the ladder and the streak.
func momentumLines(theme types.Theme, cards []sharedtypes.MomentumCard, width int) []string {
	if len(cards) == 0 {
		return nil
	}
	nameWidth := 0
	progress := make([]string, len(cards))
	for i, card := range cards {
		nameWidth = max(nameWidth, lipgloss.Width(card.Definition.Name))
		if len(card.Series) == 0 {
			continue
		}
		point := card.Series[len(card.Series)-1]
		progress[i] = fmt.Sprintf("%d/%d", point.Count, point.Target)
		if card.Definition.TargetKind == sharedtypes.MomentumTargetKindContext {
			progress[i] = helperpkg.FormatCompactDurationSeconds(point.Count) + "/" +
				helperpkg.FormatCompactDurationSeconds(point.Target)
		}
	}
	progressWidth := 0
	for _, text := range progress {
		progressWidth = max(progressWidth, ansi.StringWidth(text))
	}
	lines := []string{"", theme.StyleDim.Bold(true).Render("Momentum")}
	for i, card := range cards {
		def := card.Definition
		unit := momentumhelpers.Unit(def.Period)
		parts := []string{
			theme.StyleNormal.Render(padCells(def.Name, min(nameWidth, max(8, width/3)))),
			theme.StyleDim.Render(padCells(momentumhelpers.CadenceLabel(def.Period), 7)),
		}
		if progressWidth > 0 {
			parts = append(parts, theme.StyleNormal.Render(padCells(progress[i], progressWidth)))
		}
		parts = append(parts,
			momentumhelpers.Ladder(theme, def.Period, card.Current, card.Longest),
			theme.StyleNormal.Render(momentumhelpers.FormatLength(card.Current, unit))+
				theme.StyleDim.Render(" best "+momentumhelpers.FormatLength(card.Longest, unit)),
		)
		lines = append(lines, viewhelpers.TruncateANSI(strings.Join(parts, "  "), width))
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
func HabitDetailLines(theme types.Theme, state types.ContentState, habit api.HabitWithMeta, width int) []string {
	item := api.HabitDailyItem{HabitWithMeta: habit}
	for _, due := range state.DueHabits {
		if due.ID == habit.ID {
			item = due
			break
		}
	}
	return habitDetailLines(theme, state, item, width)
}
