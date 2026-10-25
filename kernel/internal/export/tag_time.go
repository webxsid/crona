package export

import (
	"context"
	"fmt"

	"crona/kernel/internal/core"
	corecommands "crona/kernel/internal/core/commands"
	shareddto "crona/shared/dto"
	sharedtypes "crona/shared/types"
)

// computeTagTime returns focused time per tag for the range, or nil when no worked issue was
// tagged. A session counts fully toward each of its issue's tags, so rows can add up to more
// than the time worked.
func computeTagTime(
	ctx context.Context,
	c *core.Context,
	start, end string,
	repoID, streamID *int64,
) ([]sharedtypes.TimeDistributionRow, error) {
	summary, err := corecommands.ComputeTimeDistributionSummary(ctx, c, shareddto.DashboardSummaryQuery{
		Start:    start,
		End:      end,
		GroupBy:  string(sharedtypes.DistributionGroupTag),
		RepoID:   repoID,
		StreamID: streamID,
	})
	if err != nil {
		return nil, err
	}
	for _, row := range summary.Rows {
		if row.Key != "tag:"+sharedtypes.TagUntagged {
			return summary.Rows, nil
		}
	}
	return nil, nil
}

func mapTagTime(rows []sharedtypes.TimeDistributionRow) []map[string]any {
	if len(rows) == 0 {
		return nil
	}
	out := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		out = append(out, map[string]any{
			"name":          row.Label,
			"workedTime":    formatDurationHMS(row.WorkedSeconds),
			"workedSeconds": row.WorkedSeconds,
			"percent":       fmt.Sprintf("%.0f%%", row.Percent),
		})
	}
	return out
}
