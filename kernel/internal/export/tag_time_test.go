package export

import (
	"strings"
	"testing"

	sharedtypes "crona/shared/types"
)

func TestDailyReportShowsTimeByTagOnlyWhenTagged(t *testing.T) {
	data := &sharedtypes.DailyReportData{Date: "2026-04-08"}
	rendered, err := RenderTemplate(fallbackDailyReportTemplate, buildTemplateDataMap(data))
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if strings.Contains(rendered, "Time by Tag") {
		t.Fatalf("expected no tag section without tag time:\n%s", rendered)
	}

	data.TagTime = []sharedtypes.TimeDistributionRow{
		{Key: "tag:docs", Label: "+docs", WorkedSeconds: 5400, Percent: 90},
		{Key: "tag:untagged", Label: "+untagged", WorkedSeconds: 600, Percent: 10},
	}
	rendered, err = RenderTemplate(fallbackDailyReportTemplate, buildTemplateDataMap(data))
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	for _, want := range []string{"## Time by Tag", "| +docs | 01:30:00 | 90% |", "| +untagged |"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("missing %q:\n%s", want, rendered)
		}
	}
}
