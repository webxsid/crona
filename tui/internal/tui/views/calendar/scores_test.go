package calendar

import (
	"strings"
	"testing"

	sharedtypes "crona/shared/types"

	"github.com/charmbracelet/x/ansi"
)

func TestScoreMonthIsAHeatMapWithTheShownDaysScore(t *testing.T) {
	selection := ScoreSelection{
		AnchorDate: "2026-05-20",
		Today:      "2026-05-27",
		Scores: map[string]sharedtypes.FocusScoreRangeDay{
			"2026-05-20": {Date: "2026-05-20", Score: 82, Level: "strong", HasData: true},
		},
		AwayDates: []string{"2026-05-21"},
	}
	text := ansi.Strip(strings.Join(ScoreMonth(testTheme(), selection), "\n"))
	for _, want := range []string{"May 2026", "Wed 20 · focus 82 strong", "Mo  Tu  We", "18  19  20  21", "■ 75+", "away"} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected %q in the score calendar, got\n%s", want, text)
		}
	}
	if strings.Contains(text, "82 ") && strings.Contains(text, "20  82") {
		t.Fatalf("expected scores as colour, not numbers beside the days, got\n%s", text)
	}
	selection.AnchorDate = "2026-05-21"
	if text := ansi.Strip(ScoreMonth(testTheme(), selection)[1]); text != "Thu 21 · away" {
		t.Fatalf("expected the away day named, got %q", text)
	}
}
