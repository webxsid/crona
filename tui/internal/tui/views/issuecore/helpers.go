package issuecore

import (
	"fmt"
	"sort"
	"strings"
	"time"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	viewtypes "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
)

type Theme = viewtypes.Theme

type APIIssue struct {
	ID              int64
	Title           string
	Status          sharedtypes.IssueStatus
	EstimateMinutes *int
	TodoForDate     *string
	CompletedAt     *string
	AbandonedAt     *string
}

func NewAPIIssue(id int64, title string, status sharedtypes.IssueStatus, estimateMinutes *int, todoForDate, completedAt, abandonedAt *string) APIIssue {
	return APIIssue{ID: id, Title: title, Status: status, EstimateMinutes: estimateMinutes, TodoForDate: todoForDate, CompletedAt: completedAt, AbandonedAt: abandonedAt}
}

func PlainIssueStatus(status string) string {
	switch status {
	case "in_progress":
		return "in progress"
	case "in_review":
		return "in review"
	default:
		return status
	}
}

func IssueStatusStyle(theme Theme, status string) *lipgloss.Style {
	switch status {
	case "backlog":
		s := lipgloss.NewStyle().Foreground(theme.ColorSubtle)
		return &s
	case "planned":
		s := lipgloss.NewStyle().Foreground(theme.ColorBlue)
		return &s
	case "ready":
		s := lipgloss.NewStyle().Foreground(theme.ColorCyan)
		return &s
	case "in_progress":
		s := lipgloss.NewStyle().Foreground(theme.ColorYellow)
		return &s
	case "blocked":
		s := lipgloss.NewStyle().Foreground(theme.ColorRed)
		return &s
	case "in_review":
		s := lipgloss.NewStyle().Foreground(theme.ColorMagenta)
		return &s
	case "done":
		s := lipgloss.NewStyle().Foreground(theme.ColorGreen)
		return &s
	case "abandoned":
		s := lipgloss.NewStyle().Foreground(theme.ColorRed)
		return &s
	default:
		s := lipgloss.NewStyle().Foreground(theme.ColorWhite)
		return &s
	}
}

func IssueDueSuffix(status sharedtypes.IssueStatus, todoForDate, completedAt, abandonedAt *string) string {
	if resolvedOn := resolvedOnDate(status, completedAt, abandonedAt); resolvedOn != "" {
		return "  [on " + resolvedOn + "]"
	}
	if todoForDate == nil || strings.TrimSpace(*todoForDate) == "" {
		return ""
	}
	date := strings.TrimSpace(*todoForDate)
	today := time.Now().Format("2006-01-02")
	if date == today {
		return "  [today]"
	}
	dueTime, err := time.Parse("2006-01-02", date)
	if err == nil {
		todayTime, todayErr := time.Parse("2006-01-02", today)
		if todayErr == nil && dueTime.Before(todayTime) {
			overdueDays := int(todayTime.Sub(dueTime).Hours() / 24)
			if overdueDays < 1 {
				overdueDays = 1
			}
			return fmt.Sprintf("  [overdue %dd]", overdueDays)
		}
	}
	return "  [due " + date + "]"
}

func resolvedOnDate(status sharedtypes.IssueStatus, completedAt, abandonedAt *string) string {
	var raw string
	switch status {
	case sharedtypes.IssueStatusDone:
		if completedAt != nil {
			raw = strings.TrimSpace(*completedAt)
		}
	case sharedtypes.IssueStatusAbandoned:
		if abandonedAt != nil {
			raw = strings.TrimSpace(*abandonedAt)
		}
	}
	if raw == "" {
		return ""
	}
	if len(raw) >= len("2006-01-02") {
		if parsed, err := time.Parse(time.RFC3339, raw); err == nil {
			return parsed.Format("2006-01-02")
		}
		return raw[:10]
	}
	return raw
}

func FilteredIssueIndices(issues []APIIssue, filter string) []int {
	filter = normalizeFilter(filter)
	out := []int{}
	for i, issue := range issues {
		text := strings.ToLower(strings.Join([]string{issue.Title, string(issue.Status)}, " "))
		if filter == "" || strings.Contains(text, filter) {
			out = append(out, i)
		}
	}
	return out
}

func filteredIssueMetaIndices(issues []api.IssueWithMeta, filter string) []int {
	filter = normalizeFilter(filter)
	out := []int{}
	for i, issue := range issues {
		text := strings.ToLower(strings.Join([]string{issue.Title, issue.RepoName, issue.StreamName, string(issue.Status)}, " "))
		if filter == "" || strings.Contains(text, filter) {
			out = append(out, i)
		}
	}
	return out
}

func PrioritizedDefaultIssueIndices(issues []api.IssueWithMeta, filter string, settings *api.CoreSettings) []int {
	indices := filteredIssueMetaIndices(issues, filter)
	if settings != nil && settings.IssueSort != "" && settings.IssueSort != sharedtypes.IssueSortPriority {
		open := make([]int, 0, len(indices))
		completed := make([]int, 0, len(indices))
		for _, idx := range indices {
			if isClosedIssueStatus(issues[idx].Status) {
				completed = append(completed, idx)
			} else {
				open = append(open, idx)
			}
		}
		return append(open, completed...)
	}
	today := time.Now().Format("2006-01-02")
	sort.SliceStable(indices, func(i, j int) bool {
		left := issues[indices[i]]
		right := issues[indices[j]]
		leftBucket, leftRank, leftDue := defaultIssuePriority(left, today)
		rightBucket, rightRank, rightDue := defaultIssuePriority(right, today)
		if leftBucket != rightBucket {
			return leftBucket < rightBucket
		}
		if leftDue != rightDue {
			if leftDue == "" {
				return false
			}
			if rightDue == "" {
				return true
			}
			return leftDue < rightDue
		}
		if leftRank != rightRank {
			return leftRank < rightRank
		}
		if left.RepoName != right.RepoName {
			return left.RepoName < right.RepoName
		}
		if left.StreamName != right.StreamName {
			return left.StreamName < right.StreamName
		}
		return left.Title < right.Title
	})
	return indices
}

func SplitDefaultIssueIndices(issues []api.IssueWithMeta, filter string, settings *api.CoreSettings) ([]int, []int) {
	ordered := PrioritizedDefaultIssueIndices(issues, filter, settings)
	open := make([]int, 0, len(ordered))
	completed := make([]int, 0, len(ordered))
	for _, idx := range ordered {
		if isClosedIssueStatus(issues[idx].Status) {
			completed = append(completed, idx)
			continue
		}
		open = append(open, idx)
	}
	return open, completed
}

func defaultIssuePriority(issue api.IssueWithMeta, today string) (bucket int, statusRank int, due string) {
	if isClosedIssueStatus(issue.Status) {
		return 3, closedIssueRank(issue.Status), closedIssueSortDate(issue)
	}
	due = normalizedDueDate(issue.TodoForDate)
	switch {
	case due != "" && due <= today:
		bucket = 0
	case due != "":
		bucket = 1
	default:
		bucket = 2
	}
	return bucket, openIssueStatusRank(issue.Status), due
}

func isClosedIssueStatus(status sharedtypes.IssueStatus) bool {
	return status == sharedtypes.IssueStatusDone || status == sharedtypes.IssueStatusAbandoned
}

func openIssueStatusRank(status sharedtypes.IssueStatus) int {
	switch status {
	case sharedtypes.IssueStatusInProgress:
		return 0
	case sharedtypes.IssueStatusBlocked:
		return 1
	case sharedtypes.IssueStatusReady:
		return 2
	case sharedtypes.IssueStatusInReview:
		return 3
	case sharedtypes.IssueStatusPlanned:
		return 4
	case sharedtypes.IssueStatusBacklog:
		return 5
	default:
		return 6
	}
}

func closedIssueRank(status sharedtypes.IssueStatus) int {
	if status == sharedtypes.IssueStatusDone {
		return 0
	}
	return 1
}

func closedIssueSortDate(issue api.IssueWithMeta) string {
	if issue.CompletedAt != nil && strings.TrimSpace(*issue.CompletedAt) != "" {
		return "0:" + strings.TrimSpace(*issue.CompletedAt)
	}
	if issue.AbandonedAt != nil && strings.TrimSpace(*issue.AbandonedAt) != "" {
		return "1:" + strings.TrimSpace(*issue.AbandonedAt)
	}
	return "2:" + issue.Title
}

func normalizedDueDate(todoForDate *string) string {
	if todoForDate == nil {
		return ""
	}
	return strings.TrimSpace(*todoForDate)
}

func normalizeFilter(value string) string {
	return strings.TrimSpace(strings.ToLower(value))
}
