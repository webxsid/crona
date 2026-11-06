package calendar

import (
	"strings"
	"testing"

	sharedtypes "crona/shared/types"

	"github.com/charmbracelet/x/ansi"
)

func TestScoreGridBoxesEachDayWithItsScore(t *testing.T) {
	selection := ScoreSelection{
		AnchorDate: "2026-05-20",
		Today:      "2026-05-27",
		Scores: map[string]sharedtypes.FocusScoreRangeDay{
			"2026-05-20": {Date: "2026-05-20", Score: 82, Level: "strong", HasData: true},
		},
		AwayDates: []string{"2026-05-21"},
	}
	lines := ScoreGrid(testTheme(), selection, 80, false)
	text := ansi.Strip(strings.Join(lines, "\n"))
	for _, want := range []string{"May 2026  Wed 20 · focus 82 strong", "Mo", "╭", " 82", "Away", "│ 20"} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected %q in the score grid, got\n%s", want, text)
		}
	}
	if got := len(lines); got != 2+6*4 {
		t.Fatalf("expected a header, weekday row and six boxed weeks, got %d lines", got)
	}
	for _, line := range lines {
		if w := ansi.StringWidth(line); w > 80 {
			t.Fatalf("expected the grid within 80 cells, got %d", w)
		}
	}
	if week := ScoreGrid(testTheme(), selection, 80, true); len(week) != 2+4 {
		t.Fatalf("expected one boxed week, got %d lines", len(week))
	}
	if narrow := ScoreGrid(testTheme(), selection, 50, false); narrow == nil {
		t.Fatal("expected narrower boxes at 50 cells")
	}
	if none := ScoreGrid(testTheme(), selection, 40, false); none != nil {
		t.Fatal("expected no grid when even the narrowest boxes do not fit")
	}
}
