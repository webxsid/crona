package commands

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"crona/kernel/internal/core"
	sharedtypes "crona/shared/types"
)

type listSortSettings struct {
	repoSort   sharedtypes.RepoSort
	streamSort sharedtypes.StreamSort
	issueSort  sharedtypes.IssueSort
	habitSort  sharedtypes.HabitSort
}

func loadListSortSettings(ctx context.Context, c *core.Context) listSortSettings {
	settings := listSortSettings{
		repoSort:   sharedtypes.RepoSortChronologicalAsc,
		streamSort: sharedtypes.StreamSortChronologicalAsc,
		issueSort:  sharedtypes.IssueSortPriority,
		habitSort:  sharedtypes.HabitSortSchedule,
	}
	current, err := c.CoreSettings.Get(ctx, c.UserID)
	if err != nil || current == nil {
		return settings
	}
	settings.repoSort = sharedtypes.NormalizeRepoSort(current.RepoSort)
	settings.streamSort = sharedtypes.NormalizeStreamSort(current.StreamSort)
	settings.issueSort = sharedtypes.NormalizeIssueSort(current.IssueSort)
	settings.habitSort = sharedtypes.NormalizeHabitSort(current.HabitSort)
	return settings
}

func sortRepos(items []sharedtypes.Repo, order sharedtypes.RepoSort) {
	switch sharedtypes.NormalizeRepoSort(order) {
	case sharedtypes.RepoSortChronologicalDesc:
		reverseRepos(items)
	case sharedtypes.RepoSortAlphabeticalAsc:
		sort.SliceStable(items, func(i, j int) bool {
			return compareStringsAsc(items[i].Name, items[j].Name, items[i].ID, items[j].ID)
		})
	case sharedtypes.RepoSortAlphabeticalDesc:
		sort.SliceStable(items, func(i, j int) bool {
			return compareStringsDesc(items[i].Name, items[j].Name, items[i].ID, items[j].ID)
		})
	}
}

func sortStreams(items []sharedtypes.Stream, order sharedtypes.StreamSort) {
	switch sharedtypes.NormalizeStreamSort(order) {
	case sharedtypes.StreamSortChronologicalDesc:
		reverseStreams(items)
	case sharedtypes.StreamSortAlphabeticalAsc:
		sort.SliceStable(items, func(i, j int) bool {
			return compareStringsAsc(items[i].Name, items[j].Name, items[i].ID, items[j].ID)
		})
	case sharedtypes.StreamSortAlphabeticalDesc:
		sort.SliceStable(items, func(i, j int) bool {
			return compareStringsDesc(items[i].Name, items[j].Name, items[i].ID, items[j].ID)
		})
	}
}

func sortIssues(items []sharedtypes.Issue, order sharedtypes.IssueSort) {
	switch sharedtypes.NormalizeIssueSort(order) {
	case sharedtypes.IssueSortChronologicalAsc:
		return
	case sharedtypes.IssueSortChronologicalDesc:
		reverseIssues(items)
	case sharedtypes.IssueSortAlphabeticalAsc:
		sort.SliceStable(items, func(i, j int) bool {
			return compareStringsAsc(items[i].Title, items[j].Title, items[i].ID, items[j].ID)
		})
	case sharedtypes.IssueSortAlphabeticalDesc:
		sort.SliceStable(items, func(i, j int) bool {
			return compareStringsDesc(items[i].Title, items[j].Title, items[i].ID, items[j].ID)
		})
	case sharedtypes.IssueSortDueDateAsc:
		sort.SliceStable(items, func(i, j int) bool {
			return compareDueDates(items[i], items[j], true)
		})
	case sharedtypes.IssueSortDueDateDesc:
		sort.SliceStable(items, func(i, j int) bool {
			return compareDueDates(items[i], items[j], false)
		})
	default:
		sort.SliceStable(items, func(i, j int) bool {
			return compareIssuePriority(items[i], items[j])
		})
	}
}

func sortIssuesWithMeta(items []sharedtypes.IssueWithMeta, order sharedtypes.IssueSort) {
	switch sharedtypes.NormalizeIssueSort(order) {
	case sharedtypes.IssueSortChronologicalAsc:
		return
	case sharedtypes.IssueSortChronologicalDesc:
		for left, right := 0, len(items)-1; left < right; left, right = left+1, right-1 {
			items[left], items[right] = items[right], items[left]
		}
	case sharedtypes.IssueSortAlphabeticalAsc:
		sort.SliceStable(items, func(i, j int) bool {
			if compareStringsAsc(items[i].Title, items[j].Title, items[i].ID, items[j].ID) {
				return true
			}
			if strings.EqualFold(strings.TrimSpace(items[i].Title), strings.TrimSpace(items[j].Title)) && items[i].ID == items[j].ID {
				return compareStringsAsc(items[i].RepoName, items[j].RepoName, items[i].RepoID, items[j].RepoID)
			}
			return false
		})
	case sharedtypes.IssueSortAlphabeticalDesc:
		sort.SliceStable(items, func(i, j int) bool {
			if compareStringsDesc(items[i].Title, items[j].Title, items[i].ID, items[j].ID) {
				return true
			}
			if strings.EqualFold(strings.TrimSpace(items[i].Title), strings.TrimSpace(items[j].Title)) && items[i].ID == items[j].ID {
				return compareStringsDesc(items[i].RepoName, items[j].RepoName, items[i].RepoID, items[j].RepoID)
			}
			return false
		})
	case sharedtypes.IssueSortDueDateAsc:
		sort.SliceStable(items, func(i, j int) bool {
			return compareDueDates(items[i].Issue, items[j].Issue, true)
		})
	case sharedtypes.IssueSortDueDateDesc:
		sort.SliceStable(items, func(i, j int) bool {
			return compareDueDates(items[i].Issue, items[j].Issue, false)
		})
	default:
		sort.SliceStable(items, func(i, j int) bool {
			return compareIssuePriorityWithMeta(items[i], items[j])
		})
	}
}

func sortHabits(items []sharedtypes.Habit, order sharedtypes.HabitSort) {
	switch sharedtypes.NormalizeHabitSort(order) {
	case sharedtypes.HabitSortChronologicalAsc:
		return
	case sharedtypes.HabitSortChronologicalDesc:
		for left, right := 0, len(items)-1; left < right; left, right = left+1, right-1 {
			items[left], items[right] = items[right], items[left]
		}
	case sharedtypes.HabitSortAlphabeticalAsc:
		sort.SliceStable(items, func(i, j int) bool {
			return compareStringsAsc(items[i].Name, items[j].Name, items[i].ID, items[j].ID)
		})
	case sharedtypes.HabitSortAlphabeticalDesc:
		sort.SliceStable(items, func(i, j int) bool {
			return compareStringsDesc(items[i].Name, items[j].Name, items[i].ID, items[j].ID)
		})
	case sharedtypes.HabitSortTargetMinutesAsc:
		sort.SliceStable(items, func(i, j int) bool {
			return compareHabitTargets(items[i], items[j], true)
		})
	case sharedtypes.HabitSortTargetMinutesDesc:
		sort.SliceStable(items, func(i, j int) bool {
			return compareHabitTargets(items[i], items[j], false)
		})
	default:
		sort.SliceStable(items, func(i, j int) bool {
			return compareHabitsBySchedule(items[i], items[j])
		})
	}
}

func sortHabitDailyItems(items []sharedtypes.HabitDailyItem, order sharedtypes.HabitSort) {
	switch sharedtypes.NormalizeHabitSort(order) {
	case sharedtypes.HabitSortChronologicalAsc:
		return
	case sharedtypes.HabitSortChronologicalDesc:
		for left, right := 0, len(items)-1; left < right; left, right = left+1, right-1 {
			items[left], items[right] = items[right], items[left]
		}
	case sharedtypes.HabitSortAlphabeticalAsc:
		sort.SliceStable(items, func(i, j int) bool {
			return compareStringsAsc(items[i].Name, items[j].Name, items[i].ID, items[j].ID)
		})
	case sharedtypes.HabitSortAlphabeticalDesc:
		sort.SliceStable(items, func(i, j int) bool {
			return compareStringsDesc(items[i].Name, items[j].Name, items[i].ID, items[j].ID)
		})
	case sharedtypes.HabitSortTargetMinutesAsc:
		sort.SliceStable(items, func(i, j int) bool {
			return compareHabitTargets(items[i].Habit, items[j].Habit, true)
		})
	case sharedtypes.HabitSortTargetMinutesDesc:
		sort.SliceStable(items, func(i, j int) bool {
			return compareHabitTargets(items[i].Habit, items[j].Habit, false)
		})
	default:
		sort.SliceStable(items, func(i, j int) bool {
			return compareHabitsBySchedule(items[i].Habit, items[j].Habit)
		})
	}
}

func sortHabitMetas(items []sharedtypes.HabitWithMeta, order sharedtypes.HabitSort) {
	switch sharedtypes.NormalizeHabitSort(order) {
	case sharedtypes.HabitSortChronologicalAsc:
		return
	case sharedtypes.HabitSortChronologicalDesc:
		for left, right := 0, len(items)-1; left < right; left, right = left+1, right-1 {
			items[left], items[right] = items[right], items[left]
		}
	case sharedtypes.HabitSortAlphabeticalAsc:
		sort.SliceStable(items, func(i, j int) bool {
			return compareStringsAsc(items[i].Name, items[j].Name, items[i].ID, items[j].ID)
		})
	case sharedtypes.HabitSortAlphabeticalDesc:
		sort.SliceStable(items, func(i, j int) bool {
			return compareStringsDesc(items[i].Name, items[j].Name, items[i].ID, items[j].ID)
		})
	case sharedtypes.HabitSortTargetMinutesAsc:
		sort.SliceStable(items, func(i, j int) bool {
			return compareHabitTargets(items[i].Habit, items[j].Habit, true)
		})
	case sharedtypes.HabitSortTargetMinutesDesc:
		sort.SliceStable(items, func(i, j int) bool {
			return compareHabitTargets(items[i].Habit, items[j].Habit, false)
		})
	default:
		sort.SliceStable(items, func(i, j int) bool {
			return compareHabitsBySchedule(items[i].Habit, items[j].Habit)
		})
	}
}

func compareStringsAsc(left string, right string, leftID int64, rightID int64) bool {
	lowerLeft := strings.ToLower(strings.TrimSpace(left))
	lowerRight := strings.ToLower(strings.TrimSpace(right))
	if lowerLeft != lowerRight {
		return lowerLeft < lowerRight
	}
	return leftID < rightID
}

func compareStringsDesc(left string, right string, leftID int64, rightID int64) bool {
	lowerLeft := strings.ToLower(strings.TrimSpace(left))
	lowerRight := strings.ToLower(strings.TrimSpace(right))
	if lowerLeft != lowerRight {
		return lowerLeft > lowerRight
	}
	return leftID > rightID
}

func compareDueDates(left sharedtypes.Issue, right sharedtypes.Issue, asc bool) bool {
	leftDue := normalizedDueDateValue(left.TodoForDate)
	rightDue := normalizedDueDateValue(right.TodoForDate)
	if leftDue != rightDue {
		if leftDue == "" {
			return false
		}
		if rightDue == "" {
			return true
		}
		if asc {
			return leftDue < rightDue
		}
		return leftDue > rightDue
	}
	return compareStringsAsc(left.Title, right.Title, left.ID, right.ID)
}

func compareHabitTargets(left sharedtypes.Habit, right sharedtypes.Habit, asc bool) bool {
	leftTarget, leftHasTarget := normalizedHabitTarget(left.TargetMinutes)
	rightTarget, rightHasTarget := normalizedHabitTarget(right.TargetMinutes)
	if leftHasTarget != rightHasTarget {
		return leftHasTarget
	}
	if leftTarget != rightTarget {
		if asc {
			return leftTarget < rightTarget
		}
		return leftTarget > rightTarget
	}
	return compareStringsAsc(left.Name, right.Name, left.ID, right.ID)
}

func compareHabitsBySchedule(left sharedtypes.Habit, right sharedtypes.Habit) bool {
	leftKey := habitScheduleSortKey(left)
	rightKey := habitScheduleSortKey(right)
	if leftKey != rightKey {
		return leftKey < rightKey
	}
	return compareStringsAsc(left.Name, right.Name, left.ID, right.ID)
}

func compareIssuePriority(left sharedtypes.Issue, right sharedtypes.Issue) bool {
	return comparePriorityValues(left, right, "", "")
}

func compareIssuePriorityWithMeta(left sharedtypes.IssueWithMeta, right sharedtypes.IssueWithMeta) bool {
	return comparePriorityValues(left.Issue, right.Issue, left.RepoName+" "+left.StreamName, right.RepoName+" "+right.StreamName)
}

func comparePriorityValues(left sharedtypes.Issue, right sharedtypes.Issue, leftScope string, rightScope string) bool {
	today := time.Now().Format("2006-01-02")
	leftBucket, leftRank, leftDue := issuePriority(left, today)
	rightBucket, rightRank, rightDue := issuePriority(right, today)
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
	if leftScope != rightScope {
		return strings.ToLower(leftScope) < strings.ToLower(rightScope)
	}
	return compareStringsAsc(left.Title, right.Title, left.ID, right.ID)
}

func issuePriority(issue sharedtypes.Issue, today string) (bucket int, statusRank int, due string) {
	if issue.Status == sharedtypes.IssueStatusDone || issue.Status == sharedtypes.IssueStatusAbandoned {
		return 3, closedIssueRank(issue.Status), closedIssueSortDate(issue)
	}
	due = normalizedDueDateValue(issue.TodoForDate)
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

func closedIssueSortDate(issue sharedtypes.Issue) string {
	if issue.CompletedAt != nil && strings.TrimSpace(*issue.CompletedAt) != "" {
		return "0:" + strings.TrimSpace(*issue.CompletedAt)
	}
	if issue.AbandonedAt != nil && strings.TrimSpace(*issue.AbandonedAt) != "" {
		return "1:" + strings.TrimSpace(*issue.AbandonedAt)
	}
	return "2:" + strings.ToLower(strings.TrimSpace(issue.Title))
}

func normalizedDueDateValue(todoForDate *string) string {
	if todoForDate == nil {
		return ""
	}
	return strings.TrimSpace(*todoForDate)
}

func normalizedHabitTarget(target *int) (int, bool) {
	if target == nil {
		return 0, false
	}
	return *target, true
}

func habitScheduleSortKey(habit sharedtypes.Habit) string {
	typeRank := 0
	switch sharedtypes.NormalizeHabitScheduleType(habit.ScheduleType) {
	case sharedtypes.HabitScheduleDaily:
		typeRank = 0
	case sharedtypes.HabitScheduleWeekdays:
		typeRank = 1
	case sharedtypes.HabitScheduleWeekly:
		typeRank = 2
	}
	dayKey := "99"
	if len(habit.Weekdays) > 0 {
		minDay := 99
		for _, day := range habit.Weekdays {
			if day < minDay {
				minDay = day
			}
		}
		dayKey = fmt.Sprintf("%02d", minDay)
	}
	return fmt.Sprintf("%d:%s:%s", typeRank, dayKey, strings.ToLower(strings.TrimSpace(habit.Name)))
}

func reverseRepos(items []sharedtypes.Repo) {
	for left, right := 0, len(items)-1; left < right; left, right = left+1, right-1 {
		items[left], items[right] = items[right], items[left]
	}
}

func reverseStreams(items []sharedtypes.Stream) {
	for left, right := 0, len(items)-1; left < right; left, right = left+1, right-1 {
		items[left], items[right] = items[right], items[left]
	}
}

func reverseIssues(items []sharedtypes.Issue) {
	for left, right := 0, len(items)-1; left < right; left, right = left+1, right-1 {
		items[left], items[right] = items[right], items[left]
	}
}
