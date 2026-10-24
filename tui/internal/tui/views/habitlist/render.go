package habitlist

import (
	"fmt"
	"strings"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	helperpkg "crona/tui/internal/tui/helpers"
	viewchrome "crona/tui/internal/tui/views/chrome"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	"crona/tui/internal/tui/views/issuecore"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/x/ansi"
)

// StripRows renders the scoped habit summary in two width-exact rows.
func StripRows(theme types.Theme, state types.ContentState, width int) []string {
	active, paused, due, done, failed := 0, 0, 0, 0, 0
	ids := map[int64]bool{}
	for _, habit := range state.ScopedHabits {
		ids[habit.ID] = true
		if habit.Active {
			active++
		} else {
			paused++
		}
	}
	for _, habit := range state.DueHabits {
		if !ids[habit.ID] {
			continue
		}
		due++
		if habit.Completed || habit.Status == "completed" {
			done++
		}
		if habit.Status == "failed" {
			failed++
		}
	}
	best := -1
	for _, card := range state.MomentumCards {
		def := sharedtypes.NormalizeHabitStreakDefinition(card.Definition)
		if !def.Enabled || def.TargetKind != sharedtypes.MomentumTargetKindHabit ||
			def.Period != sharedtypes.HabitStreakPeriodDay ||
			len(def.HabitIDs) == 0 {
			continue
		}
		contained := true
		for _, id := range def.HabitIDs {
			if !ids[id] {
				contained = false
				break
			}
		}
		if contained {
			best = max(best, card.Longest)
		}
	}
	streak := "—"
	if best >= 0 {
		streak = fmt.Sprintf("%dd", best)
	}

	groups := [][2]string{
		{fmt.Sprintf("Active %d", active), fmt.Sprintf("Done today %d", done)},
		{fmt.Sprintf("paused %d", paused), fmt.Sprintf("failed %d", failed)},
		{fmt.Sprintf("due today %d", due), "best streak " + streak},
	}
	rows := []string{"", ""}
	for _, group := range groups {
		gap := ""
		if rows[0] != "" {
			gap = "  "
		}
		if max(
			ansi.StringWidth(rows[0]+gap+group[0]),
			ansi.StringWidth(rows[1]+gap+group[1]),
		) > width {
			break
		}
		for i := range rows {
			rows[i] += gap + group[i]
		}
	}
	for i := range rows {
		rows[i] = ansi.Truncate(rows[i], max(0, width), "")
		rows[i] += strings.Repeat(" ", max(0, width-ansi.StringWidth(rows[i])))
	}
	return rows
}

func habitIndices(state types.ContentState) []int {
	filter := helperpkg.ParseItemFilter(state.Filters["habits"])
	out := []int{}
	for i, habit := range state.ScopedHabits {
		if filter.Match(habit.Name, habit.Tags) {
			out = append(out, i)
		}
	}
	return out
}

func renderHabits(theme types.Theme, state types.ContentState, width, height int) string {
	active := state.Pane == "habits"
	cur := state.Cursors["habits"]
	indices := habitIndices(state)
	lines := []string{}
	if active {
		lines = append(
			lines,
			viewchrome.FitHints([]string{viewchrome.KeyHint(theme, "a", "new habit")}, width-2),
		)
	}
	if len(indices) == 0 {
		lines = append(lines, theme.StyleDim.Render("No habits in this scope"))
	}
	budget := height - 1 - len(lines)
	slots := max(1, (budget+1)/3)
	start, end := viewchrome.ListWindow(cur, len(indices), slots)
	if start > 0 {
		budget--
	}
	if end < len(indices) {
		budget--
	}
	start, end = viewchrome.ListWindow(cur, len(indices), max(1, (budget+1)/3))
	if start > 0 {
		lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("↑ %d more", start)))
	}
	for pos := start; pos < end; pos++ {
		if pos > start {
			lines = append(lines, "")
		}
		habit := state.ScopedHabits[indices[pos]]
		marker := "  "
		style := theme.StyleNormal
		if active && pos == cur {
			marker = viewchrome.SelectionCursor + " "
			style = theme.StyleCursor
		}
		lines = append(lines, style.Render(marker+ansi.Truncate(habit.Name, max(1, width-4), "…")))
		if active && pos == cur {
			hints := []string{
				viewchrome.KeyHint(theme, "enter", "open"),
				viewchrome.KeyHint(theme, "e", "edit"),
			}
			if state.Timer == nil || state.Timer.State == "idle" {
				hints = append(
					hints,
					viewchrome.KeyHint(theme, "D", "delete"),
					viewchrome.KeyHint(theme, "m", "log"),
				)
			}
			hints = append(hints, viewchrome.KeyHint(theme, "+", "tags"))
			lines = append(lines, "  "+viewchrome.FitHints(hints, width-4))
			continue
		}
		metaLine := theme.StyleDim.Render("  " + ansi.Truncate(habitMeta(habit), max(1, width-4), "…"))
		if chips := issuecore.TagChips(theme, habit.Tags, state.Tags, max(0, (width-4)/2)); chips != "" {
			metaLine = theme.StyleDim.Render("  "+ansi.Truncate(habitMeta(habit), max(1, width-6-ansi.StringWidth(chips)), "…")) + "  " + chips
		}
		lines = append(lines, metaLine)
	}
	if end < len(indices) {
		lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("↓ %d more", len(indices)-end)))
	}
	return viewchrome.RenderSection(
		theme,
		viewchrome.Section{
			Title:   "Habits",
			Meta:    viewhelpers.FilterMeta(fmt.Sprintf("%d", len(indices)), state.Filters["habits"]),
			Focused: active,
			Width:   width,
			Height:  height,
			Lines:   lines,
		},
	)
}
func habitMeta(habit api.HabitWithMeta) string {
	schedule := string(habit.ScheduleType)
	if schedule == "weekly" {
		names := []string{"sun", "mon", "tue", "wed", "thu", "fri", "sat"}
		days := []string{}
		for _, day := range habit.Weekdays {
			if day >= 0 && day < len(names) {
				days = append(days, names[day])
			}
		}
		schedule = strings.Join(days, ",")
	}
	target := "-"
	if habit.TargetMinutes != nil {
		target = helperpkg.FormatCompactDurationMinutes(*habit.TargetMinutes)
	}
	return habit.RepoName + " › " + habit.StreamName + " | " + schedule + " | " + target
}
