package summary

import (
	"bytes"
	"strings"
	"testing"

	sharedtypes "crona/shared/types"
)

func penaltyDayData() daySummaryData {
	live := func(id int64, score float64, highRisk bool) sharedtypes.DailyPlanEntry {
		return sharedtypes.DailyPlanEntry{
			IssueID:            id,
			CurrentDelayedDays: 5,
			PostponeCount:      3,
			FailScore:          score,
			Penalty: &sharedtypes.DailyPlanEntryPenalty{
				HighRisk:     highRisk,
				Contribution: &sharedtypes.DailyPlanPenaltyShare{AccountabilityPct: 0.54},
				IfPostponedOneDay: &sharedtypes.DailyPlanPenaltyProjection{
					AccountabilityDelta: 0.43,
				},
			},
		}
	}
	return daySummaryData{
		Summary: &sharedtypes.DailyIssueSummary{Issues: []sharedtypes.Issue{
			{ID: 1, Title: "Small slip"},
			{ID: 2, Title: "Fix sync conflict"},
		}},
		Plan: &sharedtypes.DailyPlan{Date: "2026-10-24", Entries: []sharedtypes.DailyPlanEntry{
			live(1, 1.0, false),
			live(2, 3.38, true),
			{
				IssueID:   3,
				FailScore: 2,
				Penalty:   &sharedtypes.DailyPlanEntryPenalty{Historical: true},
			},
			{IssueID: 4, Penalty: &sharedtypes.DailyPlanEntryPenalty{}},
		}},
	}
}

func TestPenaltyRowsSkipHistoricalAndUnpenalised(t *testing.T) {
	rows := penaltyRows(penaltyDayData())
	if len(rows) != 2 || rows[0].Title != "Fix sync conflict" || rows[1].Title != "Small slip" {
		t.Fatalf("expected live penalised rows, largest first, got %+v", rows)
	}
}

func TestPrintPenaltySectionShowsTable(t *testing.T) {
	var out bytes.Buffer
	if err := printPenaltySection(&out, penaltyRows(penaltyDayData())); err != nil {
		t.Fatal(err)
	}
	text := out.String()
	for _, want := range []string{"PENALTIES", "Fix sync conflict", "▲ 3.38", "54%", "+0.43", "5d"} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected %q in %q", want, text)
		}
	}
}

func TestRenderTopPenaltiesMarksHighRisk(t *testing.T) {
	lines := strings.Join(
		renderTopPenalties(renderOptions{Width: 80}, 60, penaltyRows(penaltyDayData())),
		"\n",
	)
	if !strings.Contains(lines, "top penalties") ||
		!strings.Contains(lines, "▲ 3.38   54%  Fix sync conflict") {
		t.Fatalf("unexpected top penalties: %q", lines)
	}
}
