package issues

import (
	"fmt"
	"time"

	helperpkg "crona/tui/internal/tui/helpers"
	viewchrome "crona/tui/internal/tui/views/chrome"
	contextmeta "crona/tui/internal/tui/views/contextmeta"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
)

func renderSummary(theme types.Theme, state types.ContentState, height int) string {
	topH, bottomH := viewhelpers.SplitVertical(height, 3, 3, height/2)
	row1Left, row1Right := viewhelpers.SplitHorizontal(state.Width, 28, 28, state.Width/2)
	row2Left, row2Right := viewhelpers.SplitHorizontal(state.Width, 28, 28, state.Width/2)
	row1 := lipgloss.JoinHorizontal(lipgloss.Top,
		renderStatCard(theme, "Open", openSummary(state), "active workload", theme.ColorYellow, row1Left, topH),
		renderStatCard(theme, "Closed", closedSummary(state), "done + abandoned", theme.ColorCyan, row1Right, topH),
	)
	row2 := lipgloss.JoinHorizontal(lipgloss.Top,
		renderStatCard(theme, "Due", dueSummary(state), "today vs overdue", theme.ColorGreen, row2Left, bottomH),
		renderStatCard(theme, "Estimate", estimateSummary(state), "current issue load", theme.ColorMagenta, row2Right, bottomH),
	)
	return lipgloss.JoinVertical(lipgloss.Left, row1, row2)
}

func renderCompactSummary(theme types.Theme, state types.ContentState, height int) string {
	lines := []string{
		fmt.Sprintf("%s  %s", theme.StylePaneTitle.Render("Default Dashboard"), theme.StyleHeader.Render(contextmeta.DefaultScopeLabel(state.Context))),
		viewhelpers.Truncate(openSummary(state), state.Width-6),
		viewhelpers.Truncate(closedSummary(state), state.Width-6),
		viewhelpers.Truncate(dueSummary(state), state.Width-6),
		viewhelpers.Truncate(estimateSummary(state), state.Width-6),
	}
	return viewchrome.RenderPaneBox(theme, false, state.Width, height, viewhelpers.StringsJoin(lines))
}

func renderStatCard(theme types.Theme, label, value, hint string, border lipgloss.Color, width, height int) string {
	body := []string{
		theme.StyleDim.Render(label),
		lipStyle(theme, border).Render(value),
		theme.StyleDim.Render(hint),
	}
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(border).
		Padding(0, 1).
		Width(width - 2).
		Height(height - 2).
		Render(viewhelpers.StringsJoin(body))
}

func openSummary(state types.ContentState) string {
	open, inProgress, blocked := 0, 0, 0
	for _, issue := range state.DefaultIssues {
		switch issue.Status {
		case "done", "abandoned":
			continue
		case "in_progress":
			inProgress++
		case "blocked":
			blocked++
		}
		open++
	}
	return fmt.Sprintf("open %d  in progress %d  blocked %d", open, inProgress, blocked)
}

func closedSummary(state types.ContentState) string {
	done, abandoned := 0, 0
	for _, issue := range state.DefaultIssues {
		switch issue.Status {
		case "done":
			done++
		case "abandoned":
			abandoned++
		}
	}
	return fmt.Sprintf("done %d  abandoned %d", done, abandoned)
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
	return fmt.Sprintf("estimated %s  scoped %d", helperpkg.FormatCompactDurationMinutes(estimated), scoped)
}

func lipStyle(theme types.Theme, color interface{}) styleLike {
	return styleLike{theme: theme, color: color}
}

type styleLike struct {
	theme types.Theme
	color interface{}
}

func (s styleLike) Render(text string) string {
	return newStyle(s.color).Render(text)
}

func newStyle(color interface{}) lipgloss.Style {
	style := lipgloss.NewStyle().Bold(true)
	if c, ok := color.(lipgloss.Color); ok {
		style = style.Foreground(c)
	}
	return style
}
