package summary

import (
	"fmt"
	"slices"
	"strings"

	sharedtypes "crona/shared/types"
	helperpkg "crona/tui/internal/tui/helpers"
	momentumhelpers "crona/tui/internal/tui/views/momentum"
	"crona/tui/internal/tui/views/sessionmeta"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/x/ansi"
)

// meterCells caps a period meter's width; targets above it scale down.
const meterCells = 6

// dayMomentum is one momentum the shown day can move, with its current period.
type dayMomentum struct {
	card  sharedtypes.MomentumCard
	point sharedtypes.MomentumSeriesPoint
}

// dayMomentums are the momentums the shown day feeds: habit momentums counting a habit due
// that day, and context momentums matching a planned task's repo or stream. Unmet periods
// come first, the soonest to end on top; met ones follow.
func dayMomentums(state types.ContentState) []dayMomentum {
	seen := map[string]bool{}
	var out []dayMomentum
	add := func(cards []sharedtypes.MomentumCard) {
		for _, card := range cards {
			if seen[card.Definition.ID] {
				continue
			}
			seen[card.Definition.ID] = true
			point, _ := momentumhelpers.PeriodPoint(card, state.SummaryDate)
			out = append(out, dayMomentum{card: card, point: point})
		}
	}
	for _, habit := range state.DueHabits {
		add(momentumhelpers.ForHabit(state.MomentumCards, habit.ID))
	}
	for _, issue := range state.DailyPlannedIssues {
		if meta := sessionmeta.IssueMetaByID(state.AllIssues, issue.ID); meta != nil {
			add(momentumhelpers.ForContext(state.MomentumCards, meta.RepoID, issue.StreamID))
		}
	}
	slices.SortStableFunc(out, func(a, b dayMomentum) int {
		if a.point.MetTarget != b.point.MetTarget {
			if a.point.MetTarget {
				return 1
			}
			return -1
		}
		return strings.Compare(a.point.EndDate, b.point.EndDate)
	})
	return out
}

// renderMomentum lists the day's momentums, one row each: name, cadence, this period's meter
// and progress (✓ once met), and the streak with its best; then the focus and check-in
// streaks on one footer row.
func renderMomentum(theme types.Theme, state types.ContentState, streaks *sharedtypes.StreakSummary, width int) []string {
	lines := []string{theme.StylePaneTitle.Render("Momentum")}
	items := dayMomentums(state)
	if len(items) == 0 {
		lines = append(lines, theme.StyleDim.Render(ansi.Truncate("Nothing due feeds a momentum.", width, "…")))
	}
	nameWidth, progressWidth := 0, 0
	progress := make([]string, len(items))
	for i, item := range items {
		nameWidth = max(nameWidth, ansi.StringWidth(item.card.Definition.Name))
		progress[i] = periodProgress(item)
		progressWidth = max(progressWidth, ansi.StringWidth(progress[i]))
	}
	nameWidth = min(nameWidth, max(8, width/3))
	for i, item := range items {
		def := item.card.Definition
		unit := momentumhelpers.Unit(def.Period)
		status := "  "
		if item.point.MetTarget {
			status = theme.StyleCursor.Foreground(theme.ColorGreen).Render("✓") + " "
		}
		row := strings.Join([]string{
			theme.StyleNormal.Render(padTo(ansi.Truncate(def.Name, nameWidth, "…"), nameWidth)),
			theme.StyleDim.Render(padTo(momentumhelpers.CadenceLabel(def.Period), 7)),
			periodMeter(theme, item),
			theme.StyleNormal.Render(padTo(progress[i], progressWidth)) + " " + status +
				streakText(theme, item.card.Current, item.card.Longest, unit),
		}, "  ")
		lines = append(lines, ansi.Truncate(row, width, "…"))
	}
	if streaks != nil {
		focus := theme.StyleDim.Render("Focus ") + streakText(theme, streaks.CurrentFocusDays, streaks.LongestFocusDays, "d")
		checkIns := theme.StyleDim.Render("Check-ins ") + streakText(theme, streaks.CurrentCheckInDays, streaks.LongestCheckInDays, "d")
		lines = append(lines, "")
		if one := focus + theme.StyleDim.Render("  ·  ") + checkIns; ansi.StringWidth(one) <= width {
			lines = append(lines, one)
		} else {
			lines = append(lines, ansi.Truncate(focus, width, "…"), ansi.Truncate(checkIns, width, "…"))
		}
	}
	return lines
}

// periodProgress is this period's count against its target, as time for context momentums.
func periodProgress(item dayMomentum) string {
	if item.card.Definition.TargetKind == sharedtypes.MomentumTargetKindContext {
		return helperpkg.FormatCompactDurationSeconds(item.point.Count) + "/" +
			helperpkg.FormatCompactDurationSeconds(item.point.Target)
	}
	return fmt.Sprintf("%d/%d", item.point.Count, item.point.Target)
}

// periodMeter fills one cell per count up to the target, scaled down past meterCells.
func periodMeter(theme types.Theme, item dayMomentum) string {
	target := max(1, item.point.Target)
	cells := min(target, meterCells)
	if item.card.Definition.TargetKind == sharedtypes.MomentumTargetKindContext {
		cells = meterCells
	}
	filled := min(cells, item.point.Count*cells/target)
	color := theme.ColorYellow
	if item.point.MetTarget {
		color = theme.ColorGreen
	}
	return theme.StyleNormal.Foreground(color).Render(strings.Repeat("▰", filled)) +
		theme.StyleDim.Render(strings.Repeat("▱", cells-filled)) +
		strings.Repeat(" ", meterCells-cells)
}

// streakText is the current run, ↑ when live, with the best beside it, dim.
func streakText(theme types.Theme, current, best int, unit string) string {
	now := theme.StyleDim.Render(momentumhelpers.FormatLength(current, unit))
	if current > 0 {
		now = theme.StyleNormal.Render("↑" + momentumhelpers.FormatLength(current, unit))
	}
	return now + theme.StyleDim.Render(" best "+momentumhelpers.FormatLength(best, unit))
}

func padTo(s string, width int) string {
	return s + strings.Repeat(" ", max(0, width-ansi.StringWidth(s)))
}
