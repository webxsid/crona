package export

import (
	"context"
	"fmt"
	"slices"

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
	filter *sharedtypes.TagFilter,
) ([]sharedtypes.TimeDistributionRow, error) {
	summary, err := corecommands.ComputeTimeDistributionSummary(ctx, c, shareddto.DashboardSummaryQuery{
		Start:    start,
		End:      end,
		GroupBy:  string(sharedtypes.DistributionGroupTag),
		RepoID:   repoID,
		StreamID: streamID,
		Tags:     filter,
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

// keepTagged reports whether an item with tags passes filter; a nil or empty filter keeps all.
func keepTagged(filter *sharedtypes.TagFilter, tags []string) bool {
	return filter == nil || filter.Empty() || filter.Matches(tags)
}

// filterDailyReportData narrows a daily report's issues, sessions and tag table to filter.
func filterDailyReportData(
	ctx context.Context,
	c *core.Context,
	data *sharedtypes.DailyReportData,
	filter *sharedtypes.TagFilter,
) error {
	if filter == nil || filter.Empty() {
		return nil
	}
	allIssues, err := corecommands.ListAllIssues(ctx, c)
	if err != nil {
		return err
	}
	tagsByIssue := make(map[int64][]string, len(allIssues))
	for _, issue := range allIssues {
		tagsByIssue[issue.ID] = issue.Tags
	}
	data.Issues = slices.DeleteFunc(data.Issues, func(issue sharedtypes.DailyReportIssue) bool {
		return !filter.Matches(issue.Tags)
	})
	data.Sessions = slices.DeleteFunc(data.Sessions, func(session sharedtypes.DailyReportSession) bool {
		return !filter.Matches(tagsByIssue[session.IssueID])
	})
	data.TagTime, err = computeTagTime(ctx, c, data.Date, data.Date, nil, nil, filter)
	return err
}
