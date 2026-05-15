package selection

import (
	"fmt"
	"strings"
	"time"

	"crona/tui/internal/api"
	helperpkg "crona/tui/internal/tui/helpers"
	uistate "crona/tui/internal/tui/state"
	issuecore "crona/tui/internal/tui/views/issuecore"
)

func buildDefaultScopedIssues(s Snapshot) []api.IssueWithMeta {
	if s.Context == nil {
		return s.AllIssues
	}
	if s.Context.StreamID != nil {
		out := make([]api.IssueWithMeta, 0, len(s.AllIssues))
		for _, issue := range s.AllIssues {
			if issue.StreamID == *s.Context.StreamID {
				out = append(out, issue)
			}
		}
		return out
	}
	if s.Context.RepoID != nil {
		out := make([]api.IssueWithMeta, 0, len(s.AllIssues))
		for _, issue := range s.AllIssues {
			if issue.RepoID == *s.Context.RepoID {
				out = append(out, issue)
			}
		}
		return out
	}
	return s.AllIssues
}

func buildFilteredDueHabits(s Snapshot) []api.HabitDailyItem {
	if s.Context == nil {
		return s.DueHabits
	}
	if s.Context.StreamID != nil {
		out := make([]api.HabitDailyItem, 0, len(s.DueHabits))
		for _, habit := range s.DueHabits {
			if habit.StreamID == *s.Context.StreamID {
				out = append(out, habit)
			}
		}
		return out
	}
	if s.Context.RepoID != nil {
		out := make([]api.HabitDailyItem, 0, len(s.DueHabits))
		for _, habit := range s.DueHabits {
			if habit.RepoID == *s.Context.RepoID {
				out = append(out, habit)
			}
		}
		return out
	}
	return s.DueHabits
}

func habitHistorySummary(entry api.HabitCompletion, settings *api.CoreSettings) string {
	parts := []string{helperpkg.FormatDisplayDate(entry.Date, settings), strings.TrimSpace(entry.HabitName)}
	if parts[1] == "" {
		parts[1] = fmt.Sprintf("habit %d", entry.HabitID)
	}
	scope := strings.TrimSpace(entry.RepoName)
	if strings.TrimSpace(entry.StreamName) != "" {
		if scope != "" {
			scope += " > "
		}
		scope += strings.TrimSpace(entry.StreamName)
	}
	if scope != "" {
		parts = append(parts, scope)
	}
	parts = append(parts, string(entry.Status))
	if entry.DurationMinutes != nil {
		parts = append(parts, helperpkg.FormatCompactDurationMinutes(*entry.DurationMinutes))
	}
	if entry.Notes != nil && strings.TrimSpace(*entry.Notes) != "" {
		parts = append(parts, strings.TrimSpace(*entry.Notes))
	} else if entry.SnapshotName != nil && strings.TrimSpace(*entry.SnapshotName) != "" {
		parts = append(parts, strings.TrimSpace(*entry.SnapshotName))
	}
	return strings.Join(parts, "  ")
}

func buildDailyScopedIssues(s Snapshot) []api.Issue {
	anchorDate := dailyAnchorDate(s)
	out := make([]api.Issue, 0)
	for _, issue := range s.AllIssues {
		if !issueMatchesDailyContext(issue, s.Context) {
			continue
		}
		switch s.DailyTaskSection {
		case uistate.DailyTaskSectionPinned:
			if dailyIssuePinned(issue.Issue, anchorDate) {
				out = append(out, issue.Issue)
			}
		case uistate.DailyTaskSectionOverdue:
			if dailyIssueOverdue(issue.Issue, anchorDate) {
				out = append(out, issue.Issue)
			}
		default:
			if dailyIssueToday(issue.Issue, anchorDate) {
				out = append(out, issue.Issue)
			}
		}
	}
	return out
}

func dailyAnchorDate(s Snapshot) string {
	if trimmed := strings.TrimSpace(s.DashboardDate); trimmed != "" {
		return trimmed
	}
	return time.Now().Format("2006-01-02")
}

func issueMatchesDailyContext(issue api.IssueWithMeta, ctx *api.ActiveContext) bool {
	if ctx == nil {
		return true
	}
	if ctx.StreamID != nil {
		return issue.StreamID == *ctx.StreamID
	}
	if ctx.RepoID != nil {
		return issue.RepoID == *ctx.RepoID
	}
	return true
}

func isClosedDailyIssue(issue api.Issue) bool {
	return issue.Status == "done" || issue.Status == "abandoned"
}

func dailyIssueOverdue(issue api.Issue, anchorDate string) bool {
	if isClosedDailyIssue(issue) {
		return false
	}
	if issue.PinnedDaily && issue.TodoForDate == nil {
		return false
	}
	if issue.TodoForDate != nil {
		due := strings.TrimSpace(*issue.TodoForDate)
		return due != "" && due < anchorDate
	}
	return false
}

func dailyIssueToday(issue api.Issue, anchorDate string) bool {
	if issue.TodoForDate == nil {
		return false
	}
	due := strings.TrimSpace(*issue.TodoForDate)
	return due == anchorDate
}

func dailyIssuePinned(issue api.Issue, anchorDate string) bool {
	if !issue.PinnedDaily {
		return false
	}
	return !dailyIssueOverdue(issue, anchorDate)
}

func rawIndexForIssueID(s Snapshot, pane uistate.Pane, issueID int64) int {
	switch pane {
	case uistate.PaneIssues:
		switch s.View {
		case uistate.ViewDefault:
			scoped := DefaultScopedIssues(s)
			ordered := issuecore.PrioritizedDefaultIssueIndices(scoped, s.Filters[pane], s.Settings)
			for rawIdx, scopedIdx := range ordered {
				if scoped[scopedIdx].ID == issueID {
					return rawIdx
				}
			}
		case uistate.ViewDaily:
			issues := DailyScopedIssues(s)
			for i, issue := range issues {
				if issue.ID == issueID {
					return i
				}
			}
		case uistate.ViewMeta:
			for i, issue := range s.Issues {
				if issue.ID == issueID {
					return i
				}
			}
		}
	}
	return -1
}
