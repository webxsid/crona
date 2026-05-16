package wellbeing

import (
	"fmt"

	viewhelpers "crona/tui/internal/tui/views/helpers"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
)

func cards(theme types.Theme, state types.ContentState, width int) []string {
	all := []string{
		card(theme, "Mood", label(LabelTypeMood, state)),
		card(theme, "Energy", label(LabelTypeEnergy, state)),
		card(theme, "Sleep", label(LabelTypeSleep, state)),
		card(theme, "Worked", label(LabelTypeWorked, state)),
		card(theme, "Streaks", label(LabelTypeStreak, state)),
	}
	return wrapJoinedCards(all, max(24, width-6))
}

func compactCards(theme types.Theme, state types.ContentState) []string {
	return []string{
		fmt.Sprintf("%s  %s   %s  %s", theme.StyleHeader.Render("Mood"), label(LabelTypeMood, state), theme.StyleHeader.Render("Energy"), label(LabelTypeEnergy, state)),
		fmt.Sprintf("%s  %s   %s  %s", theme.StyleHeader.Render("Sleep"), label(LabelTypeSleep, state), theme.StyleHeader.Render("Worked"), label(LabelTypeWorked, state)),
	}
}

func trendCards(theme types.Theme, state types.ContentState) []string {
	return []string{
		fmt.Sprintf("%s  %d   %s  %d", theme.StyleHeader.Render("Days"), state.MetricsRollup.Days, theme.StyleHeader.Render("Check-ins"), state.MetricsRollup.CheckInDays),
		fmt.Sprintf("%s  %d   %s  %s", theme.StyleHeader.Render("Focus Days"), state.MetricsRollup.FocusDays, theme.StyleHeader.Render("Worked"), viewhelpers.FormatClockText(state.MetricsRollup.WorkedSeconds)),
		fmt.Sprintf("%s  %s   %s  %d", theme.StyleHeader.Render("Rest"), viewhelpers.FormatClockText(state.MetricsRollup.RestSeconds), theme.StyleHeader.Render("Sessions"), state.MetricsRollup.SessionCount),
	}
}

func card(theme types.Theme, title, value string) string {
	return theme.StyleHeader.Render(title) + " " + theme.StyleNormal.Render(value)
}

func wrapJoinedCards(cards []string, maxWidth int) []string {
	if len(cards) == 0 {
		return nil
	}
	lines := []string{}
	current := ""
	for _, item := range cards {
		if current == "" {
			current = item
			continue
		}
		candidate := current + "   " + item
		if lipgloss.Width(candidate) > maxWidth {
			lines = append(lines, current)
			current = item
			continue
		}
		current = candidate
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}
