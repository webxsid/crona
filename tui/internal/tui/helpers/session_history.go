package helpers

import (
	"fmt"

	"crona/tui/internal/api"
)

func ActiveTimerIssueID(timer *api.TimerState) *int64 {
	if timer != nil && timer.IssueID != nil && *timer.IssueID > 0 {
		return timer.IssueID
	}
	return nil
}

func SessionHistoryScopeIssueID(timer *api.TimerState) *int64 {
	return ActiveTimerIssueID(timer)
}

func SessionHistoryTitle(issueID *int64, issue *api.IssueWithMeta) string {
	if issueID == nil {
		return "Session History"
	}
	if issue != nil {
		return fmt.Sprintf("Session History For #%d %s", issue.ID, issue.Title)
	}
	return fmt.Sprintf("Session History For Issue #%d", *issueID)
}

func SessionHistorySubtitle(issueID *int64, issue *api.IssueWithMeta) string {
	if issueID == nil {
		return "Recent sessions across the workspace"
	}
	if issue != nil {
		return fmt.Sprintf("Previous sessions for the active issue in [%s/%s]", issue.RepoName, issue.StreamName)
	}
	return "Previous sessions for the active issue"
}
