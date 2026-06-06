package momentumview

import (
	"strings"
	"testing"
	"time"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/tui/chrome"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
)

func TestRenderUsesBucketTimelineAndWideCard(t *testing.T) {
	state := types.ContentState{
		View:   "momentum",
		Pane:   "momentum_cards",
		Width:  140,
		Height: 50,
		MomentumCards: []sharedtypes.MomentumCard{
			{
				Definition: sharedtypes.HabitStreakDefinition{
					Name:          "Signal Quality",
					Description:   ptrString("Keep the signal steady and visible."),
					Enabled:       true,
					Period:        sharedtypes.HabitStreakPeriodWeek,
					RequiredCount: 3,
				},
				Current:    4,
				Longest:    7,
				HabitNames: []string{"Alpha", "Beta", "Gamma", "Delta"},
				Series: []sharedtypes.MomentumSeriesPoint{
					{BucketKey: "2026-W18", Label: "May 4-10", Count: 1, Target: 3, MetTarget: false},
					{BucketKey: "2026-W19", Label: "May 11-17", Count: 2, Target: 3, MetTarget: false},
					{BucketKey: "2026-W20", Label: "May 18-24", Count: 3, Target: 3, MetTarget: true},
					{BucketKey: "2026-W21", Label: "May 25-31", Count: 4, Target: 3, MetTarget: true},
					{BucketKey: "2026-W22", Label: "Jun 1-7", Count: 2, Target: 3, MetTarget: false},
					{BucketKey: "2026-W23", Label: "Jun 8-14", Count: 5, Target: 3, MetTarget: true},
					{BucketKey: "2026-W24", Label: "Jun 15-21", Count: 6, Target: 3, MetTarget: true},
					{BucketKey: "2026-W25", Label: "Jun 22-28", Count: 4, Target: 3, MetTarget: true},
					{BucketKey: "2026-W26", Label: "Jun 29-Jul 5", Count: 13, Target: 10, MetTarget: true},
				},
			},
		},
		Cursors: map[string]int{"momentum_cards": 0},
	}

	rendered := Render(testTheme(), state)
	if !strings.Contains(rendered, "Signal Quality") {
		t.Fatalf("expected card title in render, got %q", rendered)
	}
	if !strings.Contains(rendered, "Keep the signal steady and visible.") {
		t.Fatalf("expected card description in render, got %q", rendered)
	}
	if !strings.Contains(rendered, "Habits:") {
		t.Fatalf("expected habit summary in render, got %q", rendered)
	}
	if !strings.Contains(rendered, "weekly · target 3/week · current streak 4w · best 7w") {
		t.Fatalf("expected streak header in render, got %q", rendered)
	}
	if !strings.Contains(rendered, "[18] May 4-10") || !strings.Contains(rendered, "[25] Jun 22-28") {
		t.Fatalf("expected bucket labels in render, got %q", rendered)
	}
	if !strings.Contains(rendered, "1/3") || !strings.Contains(rendered, "3/3") || !strings.Contains(rendered, "4/3") {
		t.Fatalf("expected explicit ratios in render, got %q", rendered)
	}
	if !strings.Contains(rendered, "13/10") {
		t.Fatalf("expected overflow ratio in render, got %q", rendered)
	}
	if !strings.Contains(rendered, "missed") || !strings.Contains(rendered, "met") {
		t.Fatalf("expected bucket statuses in render, got %q", rendered)
	}
	if !strings.Contains(rendered, "┆") {
		t.Fatalf("expected inline target marker in render, got %q", rendered)
	}
	if !strings.Contains(lineContaining(rendered, "[18] May 4-10"), "█") {
		t.Fatalf("expected missed bucket row to show a partial fill, got %q", rendered)
	}
	overflowLine := lineContaining(rendered, "13/10")
	if overflowLine == "" {
		t.Fatalf("expected overflow row in render, got %q", rendered)
	}
	if strings.LastIndex(overflowLine, "┆") < 0 || strings.LastIndex(overflowLine, "█") < 0 {
		t.Fatalf("expected overflow row to show both fill and target marker, got %q", overflowLine)
	}
	if strings.LastIndex(overflowLine, "┆") >= strings.LastIndex(overflowLine, "█") {
		t.Fatalf("expected overflow row fill to extend beyond the target marker, got %q", overflowLine)
	}
	if !hasWidePaneLine(rendered, 130) {
		t.Fatalf("expected a wide card layout in render, got %q", rendered)
	}
	if !strings.Contains(rendered, "Latest bucket:") {
		t.Fatalf("expected latest bucket summary in render, got %q", rendered)
	}
	if strings.ContainsFunc(rendered, func(r rune) bool { return r >= 0x2800 && r <= 0x28FF }) {
		t.Fatalf("did not expect braille output in render, got %q", rendered)
	}
}

func TestRenderShowsDisabledMomentumAsPausedHistory(t *testing.T) {
	state := types.ContentState{
		View:   "momentum",
		Pane:   "momentum_cards",
		Width:  120,
		Height: 34,
		MomentumCards: []sharedtypes.MomentumCard{
			{
				Definition: sharedtypes.HabitStreakDefinition{
					Name:          "Quiet Mix",
					Enabled:       false,
					Period:        sharedtypes.HabitStreakPeriodWeek,
					RequiredCount: 3,
				},
				Current:    4,
				Longest:    7,
				HabitNames: []string{"Alpha", "Beta"},
				Series: []sharedtypes.MomentumSeriesPoint{
					{BucketKey: "2026-W18", Label: "May 4-10", Count: 1, Target: 3, MetTarget: false},
					{BucketKey: "2026-W19", Label: "May 11-17", Count: 3, Target: 3, MetTarget: true},
				},
			},
		},
		Cursors: map[string]int{"momentum_cards": 0},
	}

	rendered := Render(testTheme(), state)
	if !strings.Contains(rendered, "Disabled") {
		t.Fatalf("expected disabled badge in render, got %q", rendered)
	}
	if !strings.Contains(rendered, "paused") {
		t.Fatalf("expected paused status in disabled render, got %q", rendered)
	}
	if strings.Contains(rendered, "missed") || strings.Contains(rendered, "met") {
		t.Fatalf("expected disabled momentum to avoid live status labels, got %q", rendered)
	}
	if !strings.Contains(rendered, "1/3") || !strings.Contains(rendered, "3/3") {
		t.Fatalf("expected disabled momentum to keep historical counts, got %q", rendered)
	}
	if !strings.Contains(rendered, "Momentum disabled") {
		t.Fatalf("expected disabled footer in render, got %q", rendered)
	}
}

func TestRenderUsesHeatmapForDailyCadence(t *testing.T) {
	state := types.ContentState{
		View:   "momentum",
		Pane:   "momentum_cards",
		Width:  120,
		Height: 34,
		MomentumCards: []sharedtypes.MomentumCard{
			{
				Definition: sharedtypes.HabitStreakDefinition{
					Name:          "Daily Focus",
					Enabled:       true,
					Period:        sharedtypes.HabitStreakPeriodDay,
					RequiredCount: 1,
				},
				Current:    11,
				Longest:    11,
				HabitNames: []string{"Journal"},
				Series:     dailyMomentumSeries(time.Date(2026, 5, 4, 0, 0, 0, 0, time.UTC), []int{0, 1, 1, 1, 0, 1, 1, 1, 1, 1, 2, 1, 1, 1}, 1),
			},
		},
		Cursors: map[string]int{"momentum_cards": 0},
	}

	rendered := Render(testTheme(), state)
	if !strings.Contains(rendered, "Daily Focus") {
		t.Fatalf("expected daily card title in render, got %q", rendered)
	}
	if !strings.Contains(rendered, "daily · target 1/day · current streak 11d · best 11d") {
		t.Fatalf("expected daily header in render, got %q", rendered)
	}
	if strings.Contains(rendered, "Mon Tue Wed Thu Fri Sat Sun") || strings.Contains(rendered, "Scale  low") {
		t.Fatalf("expected daily square grid to avoid calendar headers and scale, got %q", rendered)
	}
	if strings.Contains(rendered, "┆") || strings.Contains(rendered, "W1") || strings.Contains(rendered, "May 4-10") {
		t.Fatalf("expected daily square grid to avoid bucket bars, got %q", rendered)
	}
	if !strings.Contains(rendered, "■") || !strings.Contains(rendered, "□") {
		t.Fatalf("expected filled and empty squares in render, got %q", rendered)
	}
}

func TestMomentumDailySquaresWrapByWidth(t *testing.T) {
	series := dailyMomentumSeries(time.Date(2026, 5, 4, 0, 0, 0, 0, time.UTC), []int{1, 0, 1, 1, 0, 1, 1, 1, 0, 1}, 1)
	rows := momentumDailySquareRows(testTheme(), series, 5, true)
	if len(rows) < 2 {
		t.Fatalf("expected wrapped rows for narrow width, got %v", rows)
	}
	for _, row := range rows {
		if lipgloss.Width(row) > 5 {
			t.Fatalf("expected wrapped row width to stay within limit, got %q", row)
		}
	}
	joined := strings.Join(rows, "\n")
	if !strings.Contains(joined, "■") || !strings.Contains(joined, "□") {
		t.Fatalf("expected wrapped square grid to include filled and empty squares, got %q", joined)
	}
}

func TestMomentumBucketBarUsesTargetWidthUntilOverflow(t *testing.T) {
	short := momentumBucketBar(testTheme(), sharedtypes.MomentumSeriesPoint{Count: 5, Target: 10, MetTarget: false}, 40, true)
	if got := lipgloss.Width(short); got != 10 {
		t.Fatalf("expected under-target bar width to match target, got %d for %q", got, short)
	}
	if !strings.Contains(short, "┆") {
		t.Fatalf("expected target marker in under-target bar, got %q", short)
	}

	exact := momentumBucketBar(testTheme(), sharedtypes.MomentumSeriesPoint{Count: 10, Target: 10, MetTarget: true}, 40, true)
	if got := lipgloss.Width(exact); got != 10 {
		t.Fatalf("expected exact-target bar width to match target, got %d for %q", got, exact)
	}
	if !strings.Contains(exact, "┆") {
		t.Fatalf("expected target marker in exact-target bar, got %q", exact)
	}

	overflow := momentumBucketBar(testTheme(), sharedtypes.MomentumSeriesPoint{Count: 13, Target: 10, MetTarget: true}, 40, true)
	if got := lipgloss.Width(overflow); got != 13 {
		t.Fatalf("expected overflow bar width to match achieved count, got %d for %q", got, overflow)
	}
	if !strings.Contains(overflow, "┆") {
		t.Fatalf("expected target marker in overflow bar, got %q", overflow)
	}
}

func TestMomentumDailyHeatmapUsesCompactHeaderWhenNarrow(t *testing.T) {
	series := dailyMomentumSeries(time.Date(2026, 5, 4, 0, 0, 0, 0, time.UTC), []int{0, 1, 1, 1, 0, 1, 1}, 1)
	rendered := renderMomentumDailyHeatmap(testTheme(), series, 20, true)
	if strings.Contains(rendered, "Scale") || strings.Contains(rendered, "Mon Tue") {
		t.Fatalf("did not expect calendar headers or scale in daily square render, got %q", rendered)
	}
	if !strings.Contains(rendered, "■") || !strings.Contains(rendered, "□") {
		t.Fatalf("expected daily square render to include filled and empty squares, got %q", rendered)
	}
}

func TestRenderKeepsFocusedCardFullyVisibleOnShortViewport(t *testing.T) {
	state := types.ContentState{
		View:   "momentum",
		Pane:   "momentum_cards",
		Width:  92,
		Height: 13,
		MomentumCards: []sharedtypes.MomentumCard{
			{
				Definition: sharedtypes.HabitStreakDefinition{
					Name:          "Tiny Signal",
					Enabled:       true,
					Period:        sharedtypes.HabitStreakPeriodWeek,
					RequiredCount: 2,
				},
				Current:    2,
				Longest:    5,
				HabitNames: []string{"One", "Two"},
				Series: []sharedtypes.MomentumSeriesPoint{
					{Label: "W1", Count: 1, Target: 2, MetTarget: false},
					{Label: "W2", Count: 2, Target: 2, MetTarget: true},
					{Label: "W3", Count: 1, Target: 2, MetTarget: false},
				},
			},
		},
		Cursors: map[string]int{"momentum_cards": 0},
	}

	rendered := Render(testTheme(), state)
	if strings.Contains(rendered, "...") {
		t.Fatalf("expected short viewport render to avoid clipping, got %q", rendered)
	}
	if !strings.Contains(rendered, "Tiny Signal") {
		t.Fatalf("expected focused card title in render, got %q", rendered)
	}
	if !strings.Contains(rendered, "╰") {
		t.Fatalf("expected the focused card to render its bottom border, got %q", rendered)
	}
	if !strings.Contains(rendered, "1/2") || !strings.Contains(rendered, "2/2") {
		t.Fatalf("expected ratios in short viewport render, got %q", rendered)
	}
	if !strings.Contains(rendered, "missed") || !strings.Contains(rendered, "met") {
		t.Fatalf("expected statuses in short viewport render, got %q", rendered)
	}
	if !strings.Contains(rendered, "┆") {
		t.Fatalf("expected inline target marker in short viewport render, got %q", rendered)
	}
}

func TestRenderShowsBottomOverflowHintWhenCardsContinueBelow(t *testing.T) {
	state := types.ContentState{
		View:   "momentum",
		Pane:   "momentum_cards",
		Width:  96,
		Height: 18,
		MomentumCards: []sharedtypes.MomentumCard{
			momentumTestCard("Alpha Momentum", "A1", "A2", "A3"),
			momentumTestCard("Beta Momentum", "B1", "B2", "B3"),
			momentumTestCard("Gamma Momentum", "C1", "C2", "C3"),
		},
		Cursors: map[string]int{"momentum_cards": 0},
	}

	rendered := Render(testTheme(), state)
	if !strings.Contains(rendered, "↓ 2 more") {
		t.Fatalf("expected bottom overflow hint in render, got %q", rendered)
	}
	if strings.Contains(rendered, "↑ 1 more") {
		t.Fatalf("did not expect top overflow hint at the start of the list, got %q", rendered)
	}
	if strings.Contains(rendered, "...") {
		t.Fatalf("expected overflow hint render to avoid clipping, got %q", rendered)
	}
	if !strings.Contains(rendered, "1/3") || !strings.Contains(rendered, "2/3") || !strings.Contains(rendered, "3/3") {
		t.Fatalf("expected ratios in overflow render, got %q", rendered)
	}
	if !strings.Contains(rendered, "┆") {
		t.Fatalf("expected inline target marker in overflow render, got %q", rendered)
	}
}

func TestRenderShowsTopOverflowHintWhenCardsContinueAbove(t *testing.T) {
	state := types.ContentState{
		View:   "momentum",
		Pane:   "momentum_cards",
		Width:  96,
		Height: 18,
		MomentumCards: []sharedtypes.MomentumCard{
			momentumTestCard("Alpha Momentum", "A1", "A2", "A3"),
			momentumTestCard("Beta Momentum", "B1", "B2", "B3"),
			momentumTestCard("Gamma Momentum", "C1", "C2", "C3"),
		},
		Cursors: map[string]int{"momentum_cards": 2},
	}

	rendered := Render(testTheme(), state)
	if !strings.Contains(rendered, "↑ 2 more") {
		t.Fatalf("expected top overflow hint in render, got %q", rendered)
	}
	if strings.Contains(rendered, "↓ 1 more") {
		t.Fatalf("did not expect bottom overflow hint at the end of the list, got %q", rendered)
	}
	if strings.Contains(rendered, "...") {
		t.Fatalf("expected overflow hint render to avoid clipping, got %q", rendered)
	}
	if !strings.Contains(rendered, "1/3") || !strings.Contains(rendered, "2/3") || !strings.Contains(rendered, "3/3") {
		t.Fatalf("expected ratios in overflow render, got %q", rendered)
	}
	if !strings.Contains(rendered, "┆") {
		t.Fatalf("expected inline target marker in overflow render, got %q", rendered)
	}
}

func testTheme() types.Theme {
	return types.Theme{
		ColorBlue:            chrome.ColorBlue,
		ColorCyan:            chrome.ColorCyan,
		ColorGreen:           chrome.ColorGreen,
		ColorMagenta:         chrome.ColorMagenta,
		ColorSubtle:          chrome.ColorSubtle,
		ColorYellow:          chrome.ColorYellow,
		ColorRed:             chrome.ColorRed,
		ColorDim:             chrome.ColorDim,
		ColorWhite:           chrome.ColorWhite,
		StyleActive:          chrome.StyleActive,
		StyleInactive:        chrome.StyleInactive,
		StylePaneTitle:       chrome.StylePaneTitle,
		StyleDim:             chrome.StyleDim,
		StyleCursor:          chrome.StyleCursor,
		StyleHeader:          chrome.StyleHeader,
		StyleError:           chrome.StyleError,
		StyleSelected:        chrome.StyleSelected,
		StyleSelectedInverse: chrome.StyleSelectedInverse,
		StyleNormal:          chrome.StyleNormal,
	}
}

func ptrString(value string) *string {
	return &value
}

func hasWidePaneLine(rendered string, minWidth int) bool {
	for _, line := range strings.Split(rendered, "\n") {
		if lipgloss.Width(line) >= minWidth {
			return true
		}
	}
	return false
}

func lineContaining(rendered, pattern string) string {
	for _, line := range strings.Split(rendered, "\n") {
		if strings.Contains(line, pattern) {
			return line
		}
	}
	return ""
}

func momentumTestCard(name string, labels ...string) sharedtypes.MomentumCard {
	series := make([]sharedtypes.MomentumSeriesPoint, 0, len(labels))
	for idx, label := range labels {
		series = append(series, sharedtypes.MomentumSeriesPoint{
			Label:     label,
			Count:     idx + 1,
			Target:    3,
			MetTarget: idx%2 == 1,
		})
	}
	return sharedtypes.MomentumCard{
		Definition: sharedtypes.HabitStreakDefinition{
			Name:          name,
			Enabled:       true,
			Period:        sharedtypes.HabitStreakPeriodWeek,
			RequiredCount: 3,
		},
		Current:    len(labels),
		Longest:    len(labels) + 1,
		HabitNames: []string{"Alpha", "Beta"},
		Series:     series,
	}
}

func dailyMomentumSeries(start time.Time, counts []int, target int) []sharedtypes.MomentumSeriesPoint {
	series := make([]sharedtypes.MomentumSeriesPoint, 0, len(counts))
	for idx, count := range counts {
		day := start.AddDate(0, 0, idx)
		label := day.Format("Jan 2")
		series = append(series, sharedtypes.MomentumSeriesPoint{
			BucketKey: day.Format("2006-01-02"),
			Label:     label,
			StartDate: day.Format("2006-01-02"),
			EndDate:   day.Format("2006-01-02"),
			Count:     count,
			Target:    target,
			MetTarget: count >= target,
		})
	}
	return series
}
