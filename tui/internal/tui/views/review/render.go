// Package review renders the Insights tab's Review sub-tab: one window of days, with
// work and wellbeing side by side.
package review

import (
	"fmt"
	"strings"
	"time"

	helperpkg "crona/tui/internal/tui/helpers"
	viewchrome "crona/tui/internal/tui/views/chrome"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	viewsummary "crona/tui/internal/tui/views/summary"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

const (
	daysWidth = 30
	// minDetailWidth is the wide detail's floor; it takes 2/5 of the room beside Days.
	minDetailWidth = 44
	// wideMinHeight is the shortest screen that keeps the middle slot.
	wideMinHeight = 24
)

// Render shows the strip, the Days list, and the slots its width and height allow.
func Render(theme types.Theme, state types.ContentState) string {
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
		if state.HistoricalAwayDate != "" {
			rows = append(
				rows,
				viewchrome.JoinColumns(
					theme,
					height,
					renderDays(theme, state, daysWidth, height, ""),
					renderAway(theme, state, secondWidth, height),
				),
			)
			break
		}
		second := renderDetail(
			theme,
			state,
			secondWidth,
			height,
			viewchrome.KeyHint(theme, "S", "time"),
		)
		if state.ReviewShowTime {
			second = renderTime(
				theme,
				state,
				secondWidth,
				height,
				viewchrome.KeyHint(theme, "S", "day"),
			)
		}
		rows = append(
			rows,
			viewchrome.JoinColumns(
				theme,
				height,
				renderDays(theme, state, daysWidth, height, ""),
				second,
			),
		)
	default:
		detailWidth := max(minDetailWidth, (state.Width-daysWidth-6)*2/5)
		middleWidth := state.Width - daysWidth - detailWidth - 6
		if state.HistoricalAwayDate != "" {
			rows = append(
				rows,
				viewchrome.JoinColumns(
					theme,
					height,
					renderDays(theme, state, daysWidth, height, ""),
					renderAway(theme, state, middleWidth+detailWidth+3, height),
				),
			)
			break
		}
		middle := renderTrends(theme, state, middleWidth, height)
		if state.ReviewShowTime {
			middle = renderTime(
				theme,
				state,
				middleWidth,
				height,
				viewchrome.KeyHint(theme, "S", "trends"),
			)
		}
		rows = append(rows, viewchrome.JoinColumns(theme, height,
			renderDays(theme, state, daysWidth, height, ""),
			middle,
			renderDetail(theme, state, detailWidth, height, ""),
		))
	}
	return strings.Join(rows, "\n")
}

func renderTime(
	theme types.Theme,
	state types.ContentState,
	width, height int,
	meta string,
) string {
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
	lines = append(lines, renderDistributionSection(theme, "Segments", state.SegmentDistribution)...)
	return append(lines, renderTagDistribution(theme, state.TagDistribution)...)
}

// renderDetail shows the selected day; meta is the medium layout's swap key.
func renderDetail(
	theme types.Theme,
	state types.ContentState,
	width, height int,
	meta string,
) string {
	titleStyle := theme.StyleHeader.Bold(true)
	return viewchrome.RenderSection(theme, viewchrome.Section{
		Title: dayTitle(state.DashboardDate), Meta: meta, TitleStyle: &titleStyle,
		Width: width, Height: height, Lines: DetailLines(theme, state, width-2),
	})
}

// renderAway takes the place of the slots beside Days on a past away date, so the strip
// and the day list stay put.
func renderAway(theme types.Theme, state types.ContentState, width, height int) string {
	lines := make([]string, max(0, (height-2)/2))
	lines = append(
		lines,
		lipgloss.PlaceHorizontal(
			max(1, width-2),
			lipgloss.Center,
			theme.StyleHeader.Render(awayMessage(state)),
		),
	)
	return viewchrome.RenderSection(
		theme,
		viewchrome.Section{Title: "Away", Width: width, Height: height, Lines: lines},
	)
}

func awayMessage(state types.ContentState) string {
	return fmt.Sprintf(
		"You chose to rest and recover on %s.",
		helperpkg.FormatDisplayDate(state.HistoricalAwayDate, state.Settings),
	)
}

// DetailLines is the selected day's summary, then the window's risk and its drivers; a past
// away date shows only the away message.
func DetailLines(theme types.Theme, state types.ContentState, width int) []string {
	if state.HistoricalAwayDate != "" {
		return []string{theme.StyleHeader.Render(awayMessage(state))}
	}
	lines := viewsummary.DayLines(theme, state, width)
	burnout := latestBurnout(state)
	if burnout == nil {
		return lines
	}
	lines = append(lines, "", theme.StyleHeader.Render("Risk")+"  "+burnoutBadge(theme, burnout))
	for _, line := range strings.Split(ansi.Wordwrap(burnoutSummary(burnout), max(1, width), ""), "\n") {
		lines = append(lines, theme.StyleDim.Render(line))
	}
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
