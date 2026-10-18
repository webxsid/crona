// Package review renders the Insights tab's Review sub-tab: one window of days, with
// work and wellbeing side by side.
package review

import (
	"strings"
	"time"

	viewaway "crona/tui/internal/tui/views/away"
	viewchrome "crona/tui/internal/tui/views/chrome"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	viewsummary "crona/tui/internal/tui/views/summary"
	types "crona/tui/internal/tui/views/types"
)

const (
	daysWidth   = 30
	detailWidth = 44
	// wideMinHeight is the shortest screen that keeps the middle slot.
	wideMinHeight = 24
)

// Render shows the strip, the Days list, and the slots its width and height allow.
func Render(theme types.Theme, state types.ContentState) string {
	if state.HistoricalAwayDate != "" {
		return viewaway.RenderHistorical(theme, state)
	}
	if state.Width < 1 || state.Height < 1 {
		return ""
	}
	rows := StripRows(theme, state, state.Width)
	rows = rows[:min(len(rows), state.Height)]
	height := state.Height - len(rows)
	if height == 0 {
		return strings.Join(rows, "\n")
	}
	class := viewhelpers.ClassFor(state.Width)
	if class == viewhelpers.WidthWide && state.Height < wideMinHeight {
		class = viewhelpers.WidthMedium
	}
	switch class {
	case viewhelpers.WidthNarrow:
		meta := viewchrome.FitHints([]string{
			viewchrome.KeyHint(theme, "enter", "day"),
			viewchrome.KeyHint(theme, "S", "time"),
		}, max(0, state.Width-12))
		rows = append(rows, renderDays(theme, state, state.Width, height, meta))
	case viewhelpers.WidthMedium:
		secondWidth := state.Width - daysWidth - 3
		second := renderDetail(theme, state, secondWidth, height, viewchrome.KeyHint(theme, "S", "time"))
		if state.ReviewShowTime {
			second = renderTime(theme, state, secondWidth, height, viewchrome.KeyHint(theme, "S", "day"))
		}
		rows = append(rows, viewchrome.JoinColumns(theme, height, renderDays(theme, state, daysWidth, height, ""), second))
	default:
		middleWidth := state.Width - daysWidth - detailWidth - 6
		middle := renderTrends(theme, state, middleWidth, height)
		if state.ReviewShowTime {
			middle = renderTime(theme, state, middleWidth, height, viewchrome.KeyHint(theme, "S", "trends"))
		}
		rows = append(rows, viewchrome.JoinColumns(theme, height,
			renderDays(theme, state, daysWidth, height, ""),
			middle,
			renderDetail(theme, state, detailWidth, height, ""),
		))
	}
	return strings.Join(rows, "\n")
}

func renderTime(theme types.Theme, state types.ContentState, width, height int, meta string) string {
	return viewchrome.RenderSection(theme, viewchrome.Section{
		Title: "Time", Meta: meta, Width: width, Height: height, Lines: TimeLines(theme, state, width-2),
	})
}

// TimeLines is the window's plan, progress, estimate bias and top time allocations.
func TimeLines(theme types.Theme, state types.ContentState, width int) []string {
	lines := []string{
		renderWindowLine(theme, state),
		renderProgressLine(theme, state),
		renderEstimateBiasLine(theme, state),
		"",
	}
	lines = append(lines, renderDistributionSection(theme, "Repos", state.RepoDistribution)...)
	lines = append(lines, renderDistributionSection(theme, "Streams", state.StreamDistribution)...)
	lines = append(lines, renderDistributionSection(theme, "Issues", state.IssueDistribution)...)
	return append(lines, renderDistributionSection(theme, "Segments", state.SegmentDistribution)...)
}

// renderDetail shows the selected day; meta is the medium layout's swap key.
func renderDetail(theme types.Theme, state types.ContentState, width, height int, meta string) string {
	titleStyle := theme.StyleHeader.Bold(true)
	return viewchrome.RenderSection(theme, viewchrome.Section{
		Title: dayTitle(state.DashboardDate), Meta: meta, TitleStyle: &titleStyle,
		Width: width, Height: height, Lines: DetailLines(theme, state, width-2),
	})
}

// DetailLines is the selected day's summary, then the window's risk and its drivers.
func DetailLines(theme types.Theme, state types.ContentState, width int) []string {
	lines := viewsummary.DayLines(theme, state, width)
	burnout := latestBurnout(state)
	if burnout == nil {
		return lines
	}
	lines = append(lines, "", theme.StyleHeader.Render("Risk")+"  "+burnoutBadge(theme, burnout),
		theme.StyleDim.Render(viewhelpers.Truncate(burnoutSummary(burnout), max(1, width))))
	risks, recoveries := burnoutContributorLines(burnout)
	if len(risks) > 0 {
		lines = append(lines, "", theme.StyleHeader.Render("Risk drivers"))
		lines = append(lines, risks...)
	}
	if len(recoveries) > 0 {
		lines = append(lines, "", theme.StyleHeader.Render("Recovery drivers"))
		lines = append(lines, recoveries...)
	}
	return lines
}

func dayTitle(date string) string {
	if t, err := time.Parse(isoDate, date); err == nil {
		return t.Format("Mon 2 Jan")
	}
	return date
}
