package views

import (
	"fmt"
	"sort"
	"strings"
	"time"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"

	"github.com/charmbracelet/lipgloss"
)

type apiIssue struct {
	ID              int64
	Title           string
	Status          sharedtypes.IssueStatus
	EstimateMinutes *int
	TodoForDate     *string
}

func habitItems(habits []api.Habit) []string {
	items := make([]string, 0, len(habits))
	for _, habit := range habits {
		schedule := string(habit.ScheduleType)
		if habit.ScheduleType == sharedtypes.HabitScheduleWeekly {
			schedule = formatHabitScheduleText(habit.Weekdays)
		}
		items = append(items, fmt.Sprintf("%s [%s]", habit.Name, schedule))
	}
	return items
}

func habitDailyItems(habits []api.HabitDailyItem) []string {
	items := make([]string, 0, len(habits))
	for _, habit := range habits {
		items = append(items, habit.Name)
	}
	return items
}

func formatHabitScheduleText(weekdays []int) string {
	names := []string{"sun", "mon", "tue", "wed", "thu", "fri", "sat"}
	out := make([]string, 0, len(weekdays))
	for _, day := range weekdays {
		if day >= 0 && day < len(names) {
			out = append(out, names[day])
		}
	}
	return strings.Join(out, ",")
}

func newAPIIssue(id int64, title string, status sharedtypes.IssueStatus, estimateMinutes *int, todoForDate *string) apiIssue {
	return apiIssue{
		ID:              id,
		Title:           title,
		Status:          status,
		EstimateMinutes: estimateMinutes,
		TodoForDate:     todoForDate,
	}
}

func plainIssueStatus(status string) string {
	switch status {
	case "in_progress":
		return "in progress"
	case "in_review":
		return "in review"
	default:
		return status
	}
}

func issueStatusStyle(theme Theme, status string) *lipgloss.Style {
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

func onOff(v bool) string {
	if v {
		return "on"
	}
	return "off"
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func truncate(s string, maxLen int) string {
	if maxLen < 4 {
		maxLen = 4
	}
	r := []rune(s)
	if len(r) <= maxLen {
		return s
	}
	return string(r[:maxLen-3]) + "..."
}

func formatClock(totalSeconds int) string {
	return fmt.Sprintf("%02d:%02d", totalSeconds/60, totalSeconds%60)
}

func formatEstimateProgress(elapsedSeconds, estimateMinutes int) string {
	return fmt.Sprintf("%s / %dm", formatClock(elapsedSeconds), estimateMinutes)
}

func splitVertical(total, topMin, bottomMin, topPreferred int) (int, int) {
	if total < topMin+bottomMin {
		if total <= bottomMin {
			return 0, total
		}
		return total - bottomMin, bottomMin
	}
	top := topPreferred
	if top < topMin {
		top = topMin
	}
	if top > total-bottomMin {
		top = total - bottomMin
	}
	return top, total - top
}

func splitHorizontal(total, leftMin, rightMin, leftPreferred int) (int, int) {
	if total < leftMin+rightMin {
		if total <= rightMin {
			return 0, total
		}
		return total - rightMin, rightMin
	}
	left := leftPreferred
	if left < leftMin {
		left = leftMin
	}
	if left > total-rightMin {
		left = total - rightMin
	}
	return left, total - left
}

func issueDueSuffix(todoForDate *string) string {
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

func currentDashboardDate(state ContentState) string {
	if state.DashboardDate != "" {
		return state.DashboardDate
	}
	return time.Now().Format("2006-01-02")
}

func listWindow(cur, total, inner int) (int, int) {
	half := inner / 2
	start := cur - half
	if start < 0 {
		start = 0
	}
	if start+inner > total {
		start = total - inner
	}
	if start < 0 {
		start = 0
	}
	end := start + inner
	if end > total {
		end = total
	}
	return start, end
}

func reverseIndices(in []int) []int {
	out := make([]int, len(in))
	for i := range in {
		out[i] = in[len(in)-1-i]
	}
	return out
}

func sessionHistorySummary(entry api.SessionHistoryEntry) string {
	if entry.ParsedNotes != nil {
		if m := strings.TrimSpace(entry.ParsedNotes[sharedtypes.SessionNoteSectionCommit]); m != "" {
			return m
		}
		if n := strings.TrimSpace(entry.ParsedNotes[sharedtypes.SessionNoteSectionNotes]); n != "" {
			return n
		}
	}
	if entry.Notes != nil && strings.TrimSpace(*entry.Notes) != "" {
		return strings.TrimSpace(*entry.Notes)
	}
	return fmt.Sprintf("Issue #%d", entry.IssueID)
}

func formatSessionTimestamp(value string) string {
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed.Local().Format("2006-01-02 15:04")
	}
	if len(value) >= 16 {
		return strings.Replace(value[:16], "T", " ", 1)
	}
	return value
}

func formatSessionDuration(durationSeconds *int, start string, end *string) string {
	if durationSeconds != nil {
		return formatClock(*durationSeconds)
	}
	if end != nil && *end != "" {
		st, se := time.Parse(time.RFC3339, start)
		et, ee := time.Parse(time.RFC3339, *end)
		if se == nil && ee == nil {
			return formatClock(int(et.Sub(st).Seconds()))
		}
	}
	return "-"
}

func summarizeCompletedSessions(s []api.Session) (workedSeconds int, completedCount int) {
	for _, session := range s {
		if session.DurationSeconds == nil || session.EndTime == nil {
			continue
		}
		workedSeconds += *session.DurationSeconds
		completedCount++
	}
	return
}

func renderBigClock(clock string) string {
	glyphs := map[rune][]string{'0': {"███", "█ █", "█ █", "█ █", "███"}, '1': {" ██", "██ ", " ██", " ██", "███"}, '2': {"███", "  █", "███", "█  ", "███"}, '3': {"███", "  █", "███", "  █", "███"}, '4': {"█ █", "█ █", "███", "  █", "  █"}, '5': {"███", "█  ", "███", "  █", "███"}, '6': {"███", "█  ", "███", "█ █", "███"}, '7': {"███", "  █", "  █", "  █", "  █"}, '8': {"███", "█ █", "███", "█ █", "███"}, '9': {"███", "█ █", "███", "  █", "███"}, ':': {"   ", " █ ", "   ", " █ ", "   "}}
	lines := make([]string, 5)
	for _, char := range clock {
		glyph, ok := glyphs[char]
		if !ok {
			continue
		}
		for i := range lines {
			if lines[i] != "" {
				lines[i] += "  "
			}
			lines[i] += glyph[i]
		}
	}
	return strings.Join(lines, "\n")
}

func issueMetaByID(all []api.IssueWithMeta, issueID int64) *api.IssueWithMeta {
	for i := range all {
		if all[i].ID == issueID {
			return &all[i]
		}
	}
	return nil
}

func repoItems(in []api.Repo) []string {
	out := make([]string, 0, len(in))
	for _, v := range in {
		out = append(out, v.Name)
	}
	return out
}

func streamItems(in []api.Stream) []string {
	out := make([]string, 0, len(in))
	for _, v := range in {
		out = append(out, v.Name)
	}
	return out
}

func scratchpadItems(in []api.ScratchPad) []string {
	out := make([]string, 0, len(in))
	for _, v := range in {
		out = append(out, v.Name)
	}
	return out
}

func filteredStrings(items []string, filter string) []int {
	filter = strings.TrimSpace(strings.ToLower(filter))
	out := []int{}
	for i, item := range items {
		if filter == "" || strings.Contains(strings.ToLower(item), filter) {
			out = append(out, i)
		}
	}
	return out
}

func filteredIssueMetaIndices(issues []api.IssueWithMeta, filter string) []int {
	filter = strings.TrimSpace(strings.ToLower(filter))
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

func filteredIssueIndices(issues []apiIssue, filter string) []int {
	filter = strings.TrimSpace(strings.ToLower(filter))
	out := []int{}
	for i, issue := range issues {
		text := strings.ToLower(strings.Join([]string{issue.Title, string(issue.Status)}, " "))
		if filter == "" || strings.Contains(text, filter) {
			out = append(out, i)
		}
	}
	return out
}

func filteredSettingIndices(filter string, settings *api.CoreSettings) []int {
	if settings == nil {
		return nil
	}
	labels := []string{"Timer Mode", "Breaks Enabled", "Work Duration", "Short Break", "Long Break", "Long Break Enabled", "Cycles Before Long Break", "Auto Start Breaks", "Auto Start Work", "Repo Sort", "Stream Sort", "Issue Sort"}
	return filteredStrings(labels, filter)
}

func filteredOpIndices(ops []api.Op, filter string) []int {
	filter = strings.TrimSpace(strings.ToLower(filter))
	out := []int{}
	for i, op := range ops {
		text := strings.ToLower(string(op.Entity) + " " + string(op.Action) + " " + op.EntityID)
		if filter == "" || strings.Contains(text, filter) {
			out = append(out, i)
		}
	}
	return out
}

func filteredSessionIndices(entries []api.SessionHistoryEntry, filter string) []int {
	filter = strings.TrimSpace(strings.ToLower(filter))
	out := []int{}
	for i, entry := range entries {
		text := strings.ToLower(sessionHistorySummary(entry))
		if filter == "" || strings.Contains(text, filter) {
			out = append(out, i)
		}
	}
	return out
}

func renderProgressBar(theme Theme, done, abandoned, remaining, width int) string {
	if width < 10 {
		width = 10
	}
	total := done + abandoned + remaining
	if total <= 0 {
		total = 1
	}
	doneW := (done * width) / total
	abandonedW := (abandoned * width) / total
	if doneW+abandonedW > width {
		abandonedW = max(0, width-doneW)
	}
	remainingW := width - doneW - abandonedW
	return lipgloss.NewStyle().Foreground(theme.ColorGreen).Render(strings.Repeat("█", doneW)) +
		lipgloss.NewStyle().Foreground(theme.ColorRed).Render(strings.Repeat("█", abandonedW)) +
		lipgloss.NewStyle().Foreground(theme.ColorSubtle).Render(strings.Repeat("█", remainingW))
}

func currentOpsLimit(state ContentState) int {
	visibleRows := state.Height - 6
	if visibleRows < 10 {
		visibleRows = 10
	}
	return visibleRows
}

func stringsJoin(rows []string) string {
	return strings.Join(rows, "\n")
}
