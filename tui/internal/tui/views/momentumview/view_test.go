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
		View:        "momentum",
		Pane:        "momentum_cards",
		MomentumTab: "custom",
		Width:       140,
		Height:      50,
		MomentumCards: []sharedtypes.MomentumCard{
			{
				Definition: sharedtypes.HabitStreakDefinition{
					Name:          "Signal Quality",
					Description:   new("Keep the signal steady and visible."),
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
	for _, unwanted := range []string{"1 Focus", "2 Wellbeing", "3 Custom", "Focus Momentum"} {
		if strings.Contains(rendered, unwanted) {
			t.Fatalf("did not expect momentum tab chrome in render, got %q", rendered)
		}
	}
	if !strings.Contains(rendered, "Keep the signal steady and visible.") {
		t.Fatalf("expected card description in render, got %q", rendered)
	}
	if !strings.Contains(rendered, "Habits:") {
		t.Fatalf("expected habit summary in render, got %q", rendered)
	}
	if !strings.Contains(rendered, "weekly · target Any · 3/week · current streak 4w · best 7w") {
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

func TestRenderUsesContextLabelAndDurationValues(t *testing.T) {
	state := types.ContentState{
		View:        "momentum",
		Pane:        "momentum_cards",
		MomentumTab: "custom",
		Width:       180,
		Height:      42,
		MomentumCards: []sharedtypes.MomentumCard{
			{
				Definition: sharedtypes.HabitStreakDefinition{
					Name:       "Delivery Flow",
					Enabled:    true,
					TargetKind: sharedtypes.MomentumTargetKindContext,
					MatchMode:  sharedtypes.MomentumMatchModeAll,
					Contexts: []sharedtypes.MomentumContext{
						{RepoID: 1},
						{RepoID: 2},
					},
					Period:        sharedtypes.HabitStreakPeriodWeek,
					RequiredCount: 7200,
				},
				Current: 4,
				Longest: 7,
				Series: []sharedtypes.MomentumSeriesPoint{
					{BucketKey: "2026-W18", Label: "May 4-10", Count: 7200, Target: 7200, MetTarget: true},
					{BucketKey: "2026-W19", Label: "May 11-17", Count: 3600, Target: 7200, MetTarget: false},
					{BucketKey: "2026-W20", Label: "May 18-24", Count: 10800, Target: 7200, MetTarget: true},
				},
				TargetNames: []string{"Work/app", "OSS/cli"},
			},
		},
		Cursors: map[string]int{"momentum_cards": 0},
	}

	rendered := Render(testTheme(), state)
	if !strings.Contains(rendered, "Contexts:") {
		t.Fatalf("expected context summary label in render, got %q", rendered)
	}
	if !strings.Contains(rendered, "Work/app, OSS/cli") {
		t.Fatalf("expected context target names in render, got %q", rendered)
	}
	if !strings.Contains(rendered, "weekly · target All · 2 contexts, 2h each") {
		t.Fatalf("expected context meta to use duration-based target text, got %q", rendered)
	}
	if !strings.Contains(rendered, "1h/2h") {
		t.Fatalf("expected context ratio to render as compact duration, got %q", rendered)
	}
	if !strings.Contains(rendered, "Latest bucket: 3h/2h met") {
		t.Fatalf("expected context footnote to render duration-based totals, got %q", rendered)
	}
	if strings.Contains(rendered, "3600/7200") {
		t.Fatalf("expected context render to avoid raw seconds counts, got %q", rendered)
	}
	if !strings.Contains(rendered, "█") || !strings.Contains(rendered, "│") {
		t.Fatalf("expected context render to use a vertical distribution chart, got %q", rendered)
	}
	if strings.Contains(rendered, "■") || strings.Contains(rendered, "□") {
		t.Fatalf("expected context render to avoid square cells, got %q", rendered)
	}
	shortRow := lineContaining(rendered, "May 11-17")
	longRow := lineContaining(rendered, "May 18-24")
	if shortRow == "" || longRow == "" {
		t.Fatalf("expected context rows in render, got %q", rendered)
	}
	if strings.Count(shortRow, "█") >= strings.Count(longRow, "█") {
		t.Fatalf("expected longer duration to render a larger fill, got short=%q long=%q", shortRow, longRow)
	}
}

func TestRenderUsesFixedContextYAxisTicks(t *testing.T) {
	shortState := types.ContentState{
		View:        "momentum",
		Pane:        "momentum_cards",
		MomentumTab: "custom",
		Width:       160,
		Height:      40,
		MomentumCards: []sharedtypes.MomentumCard{
			{
				Definition: sharedtypes.HabitStreakDefinition{
					Name:       "Short Context",
					Enabled:    true,
					TargetKind: sharedtypes.MomentumTargetKindContext,
					MatchMode:  sharedtypes.MomentumMatchModeAny,
					Contexts: []sharedtypes.MomentumContext{
						{RepoID: 1},
					},
					Period:        sharedtypes.HabitStreakPeriodDay,
					RequiredCount: 2700,
				},
				Current: 1,
				Series: []sharedtypes.MomentumSeriesPoint{
					{BucketKey: "2026-06-18", Label: "Jun 18", Count: 900, Target: 2700, MetTarget: false},
					{BucketKey: "2026-06-19", Label: "Jun 19", Count: 1800, Target: 2700, MetTarget: false},
				},
				TargetNames: []string{"Work"},
			},
		},
		Cursors: map[string]int{"momentum_cards": 0},
	}
	shortRendered := Render(testTheme(), shortState)
	for _, label := range []string{"00m", "15m", "30m", "45m"} {
		if !strings.Contains(shortRendered, label) {
			t.Fatalf("expected short context axis to use fixed minute ticks, missing %q in %q", label, shortRendered)
		}
	}
	if strings.Contains(shortRendered, "1h23m") {
		t.Fatalf("expected short context axis to avoid dynamic duration labels, got %q", shortRendered)
	}

	longState := types.ContentState{
		View:        "momentum",
		Pane:        "momentum_cards",
		MomentumTab: "custom",
		Width:       160,
		Height:      40,
		MomentumCards: []sharedtypes.MomentumCard{
			{
				Definition: sharedtypes.HabitStreakDefinition{
					Name:       "Long Context",
					Enabled:    true,
					TargetKind: sharedtypes.MomentumTargetKindContext,
					MatchMode:  sharedtypes.MomentumMatchModeAny,
					Contexts: []sharedtypes.MomentumContext{
						{RepoID: 1},
					},
					Period:        sharedtypes.HabitStreakPeriodDay,
					RequiredCount: 5400,
				},
				Current: 1,
				Series: []sharedtypes.MomentumSeriesPoint{
					{BucketKey: "2026-06-18", Label: "Jun 18", Count: 2700, Target: 5400, MetTarget: false},
					{BucketKey: "2026-06-19", Label: "Jun 19", Count: 5400, Target: 5400, MetTarget: true},
				},
				TargetNames: []string{"Work"},
			},
		},
		Cursors: map[string]int{"momentum_cards": 0},
	}
	longRendered := Render(testTheme(), longState)
	for _, label := range []string{"00h00m", "00h30m", "01h00m", "01h30m"} {
		if !strings.Contains(longRendered, label) {
			t.Fatalf("expected long context axis to use fixed hour ticks, missing %q in %q", label, longRendered)
		}
	}
	if strings.Contains(longRendered, "1h23m") {
		t.Fatalf("expected long context axis to avoid dynamic duration labels, got %q", longRendered)
	}
}

func TestRenderShowsDisabledMomentumAsPausedHistory(t *testing.T) {
	state := types.ContentState{
		View:        "momentum",
		Pane:        "momentum_cards",
		MomentumTab: "custom",
		Width:       120,
		Height:      34,
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
	if !strings.Contains(rendered, "Inactive") {
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

func TestRenderUsesSquareGridOnlyForBinaryDailyHabit(t *testing.T) {
	state := types.ContentState{
		View:        "momentum",
		Pane:        "momentum_cards",
		MomentumTab: "custom",
		Width:       120,
		Height:      34,
		MomentumCards: []sharedtypes.MomentumCard{
			{
				Definition: sharedtypes.HabitStreakDefinition{
					Name:          "Daily Focus",
					Enabled:       true,
					Period:        sharedtypes.HabitStreakPeriodDay,
					RequiredCount: 1,
				},
				Current:    5,
				Longest:    6,
				HabitNames: []string{"Journal"},
				Series:     dailyMomentumSeries(time.Date(2026, 5, 4, 0, 0, 0, 0, time.UTC), []int{0, 1, 1, 1, 0, 1, 1}, 1),
			},
		},
		Cursors: map[string]int{"momentum_cards": 0},
	}

	rendered := Render(testTheme(), state)
	if !strings.Contains(rendered, "Daily Focus") {
		t.Fatalf("expected daily card title in render, got %q", rendered)
	}
	if !strings.Contains(rendered, "daily · target Any · 1/day · current streak 5d · best 6d") {
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
	if strings.Contains(rendered, "█") || strings.Contains(rendered, "┆") || strings.Contains(rendered, "May 4-10") {
		t.Fatalf("expected binary daily habit to stay on the square grid, got %q", rendered)
	}
}

func TestRenderUsesCountDistributionForMultiHabitAnyDaily(t *testing.T) {
	state := types.ContentState{
		View:        "momentum",
		Pane:        "momentum_cards",
		MomentumTab: "custom",
		Width:       130,
		Height:      38,
		MomentumCards: []sharedtypes.MomentumCard{
			{
				Definition: sharedtypes.HabitStreakDefinition{
					Name:          "Multi Habit Daily",
					Enabled:       true,
					TargetKind:    sharedtypes.MomentumTargetKindHabit,
					MatchMode:     sharedtypes.MomentumMatchModeAny,
					Period:        sharedtypes.HabitStreakPeriodDay,
					RequiredCount: 1,
					HabitIDs:      []int64{1, 2, 3},
				},
				Current:    7,
				Longest:    9,
				HabitNames: []string{"Journal", "Walk", "Strength"},
				Series:     dailyMomentumSeries(time.Date(2026, 5, 4, 0, 0, 0, 0, time.UTC), []int{0, 1, 2, 3, 1, 0, 2}, 1),
			},
		},
		Cursors: map[string]int{"momentum_cards": 0},
	}

	rendered := Render(testTheme(), state)
	if !strings.Contains(rendered, "Multi Habit Daily") {
		t.Fatalf("expected daily multi-habit card title in render, got %q", rendered)
	}
	if !strings.Contains(rendered, "daily · target Any · 1/day · current streak 7d · best 9d") {
		t.Fatalf("expected daily multi-habit header in render, got %q", rendered)
	}
	if !strings.Contains(rendered, "█") || !strings.Contains(rendered, "│") {
		t.Fatalf("expected multi-habit daily momentum to use a vertical distribution chart, got %q", rendered)
	}
	if strings.Contains(rendered, "■") || strings.Contains(rendered, "□") {
		t.Fatalf("expected multi-habit daily momentum to avoid square cells, got %q", rendered)
	}
	if !strings.Contains(rendered, "Latest bucket: 2/1 met") {
		t.Fatalf("expected count-based footnote in multi-habit daily render, got %q", rendered)
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

func TestMomentumBucketBarUsesSharedScale(t *testing.T) {
	short := momentumBucketBar(testTheme(), sharedtypes.MomentumSeriesPoint{Count: 5, Target: 10, MetTarget: false}, 40, true, 13)
	if got := lipgloss.Width(short); got != 40 {
		t.Fatalf("expected scaled bar width to match the available width, got %d for %q", got, short)
	}
	if !strings.Contains(short, "┆") {
		t.Fatalf("expected target marker in under-target bar, got %q", short)
	}

	exact := momentumBucketBar(testTheme(), sharedtypes.MomentumSeriesPoint{Count: 10, Target: 10, MetTarget: true}, 40, true, 13)
	if got := lipgloss.Width(exact); got != 40 {
		t.Fatalf("expected scaled bar width to match the available width, got %d for %q", got, exact)
	}
	if !strings.Contains(exact, "┆") {
		t.Fatalf("expected target marker in exact-target bar, got %q", exact)
	}

	overflow := momentumBucketBar(testTheme(), sharedtypes.MomentumSeriesPoint{Count: 13, Target: 10, MetTarget: true}, 40, true, 13)
	if got := lipgloss.Width(overflow); got != 40 {
		t.Fatalf("expected overflow bar to stay within the available width, got %d for %q", got, overflow)
	}
	if !strings.Contains(overflow, "┆") {
		t.Fatalf("expected target marker in overflow bar, got %q", overflow)
	}
	if strings.Count(short, "█") >= strings.Count(exact, "█") {
		t.Fatalf("expected exact-target bar to render more fill than the under-target bar, got short=%q exact=%q", short, exact)
	}
	if strings.Count(exact, "█") >= strings.Count(overflow, "█") {
		t.Fatalf("expected overflow bar to render more fill than the exact-target bar, got exact=%q overflow=%q", exact, overflow)
	}
}

func TestMomentumDailySquaresStayCompactWhenNarrow(t *testing.T) {
	series := dailyMomentumSeries(time.Date(2026, 5, 4, 0, 0, 0, 0, time.UTC), []int{0, 1, 1, 1, 0, 1, 1}, 1)
	rendered := renderMomentumDailySquares(testTheme(), series, 20, true)
	if strings.Contains(rendered, "Scale") || strings.Contains(rendered, "Mon Tue") {
		t.Fatalf("did not expect calendar headers or scale in daily square render, got %q", rendered)
	}
	if !strings.Contains(rendered, "■") || !strings.Contains(rendered, "□") {
		t.Fatalf("expected daily square render to include filled and empty squares, got %q", rendered)
	}
}

func TestRenderKeepsFocusedCardFullyVisibleOnShortViewport(t *testing.T) {
	state := types.ContentState{
		View:        "momentum",
		Pane:        "momentum_cards",
		MomentumTab: "custom",
		Width:       92,
		Height:      13,
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
		View:        "momentum",
		Pane:        "momentum_cards",
		MomentumTab: "custom",
		Width:       96,
		Height:      18,
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
		View:        "momentum",
		Pane:        "momentum_cards",
		MomentumTab: "custom",
		Width:       96,
		Height:      18,
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
		ColorDullGreen:       chrome.ColorGreen,
		ColorMagenta:         chrome.ColorMagenta,
		ColorSubtle:          chrome.ColorSubtle,
		ColorYellow:          chrome.ColorYellow,
		ColorRed:             chrome.ColorRed,
		ColorDullRed:         chrome.ColorRed,
		ColorOrange:          chrome.ColorOrange,
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

func hasWidePaneLine(rendered string, minWidth int) bool {
	for line := range strings.SplitSeq(rendered, "\n") {
		if lipgloss.Width(line) >= minWidth {
			return true
		}
	}
	return false
}

func lineContaining(rendered, pattern string) string {
	for line := range strings.SplitSeq(rendered, "\n") {
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
