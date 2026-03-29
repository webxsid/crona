package selection

import (
	"fmt"
	"strings"

	"crona/tui/internal/api"
	configitems "crona/tui/internal/tui/configitems"
	helperpkg "crona/tui/internal/tui/helpers"
	uistate "crona/tui/internal/tui/state"
	"crona/tui/internal/tui/views"
)

type Snapshot struct {
	View                uistate.View
	Pane                uistate.Pane
	DefaultIssueSection uistate.DefaultIssueSection
	PreferActiveIssue   bool
	Cursors             map[uistate.Pane]int
	Filters             map[uistate.Pane]string
	Context             *api.ActiveContext
	Timer               *api.TimerState
	ActiveIssue         *api.IssueWithMeta
	Repos               []api.Repo
	Streams             []api.Stream
	Issues              []api.Issue
	Habits              []api.Habit
	AllIssues           []api.IssueWithMeta
	DueHabits           []api.HabitDailyItem
	ExportReports       []api.ExportReportFile
	Scratchpads         []api.ScratchPad
	SessionHistory      []api.SessionHistoryEntry
	Ops                 []api.Op
	Settings            *api.CoreSettings
	ConfigItems         []configitems.Item
}

type IssueSelection struct {
	ID          int64
	StreamID    int64
	Status      string
	TodoForDate *string
}

func IssueMetaByID(s Snapshot, issueID int64) *api.IssueWithMeta {
	for i := range s.AllIssues {
		if s.AllIssues[i].ID == issueID {
			return &s.AllIssues[i]
		}
	}
	return nil
}

func ActiveIssue(s Snapshot) *api.IssueWithMeta {
	if s.ActiveIssue != nil {
		return s.ActiveIssue
	}
	var issueID int64
	if s.Timer != nil && s.Timer.IssueID != nil {
		issueID = *s.Timer.IssueID
	} else if s.Context != nil && s.Context.IssueID != nil {
		issueID = *s.Context.IssueID
	}
	if issueID == 0 {
		return nil
	}
	return IssueMetaByID(s, issueID)
}

func DefaultScopedIssues(s Snapshot) []api.IssueWithMeta {
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

func FilteredDueHabits(s Snapshot) []api.HabitDailyItem {
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

func DailyScopedIssues(s Snapshot) []api.Issue {
	issues := s.Issues
	if s.Context == nil {
		return issues
	}
	if s.Context.StreamID != nil {
		out := make([]api.Issue, 0, len(issues))
		for _, issue := range issues {
			if issue.StreamID == *s.Context.StreamID {
				out = append(out, issue)
			}
		}
		return out
	}
	if s.Context.RepoID != nil {
		out := make([]api.Issue, 0, len(issues))
		for _, issue := range issues {
			meta := IssueMetaByID(s, issue.ID)
			if meta != nil && meta.RepoID == *s.Context.RepoID {
				out = append(out, issue)
			}
		}
		return out
	}
	return issues
}

func PaneItems(s Snapshot, pane uistate.Pane) []string {
	switch pane {
	case uistate.PaneRepos:
		items := make([]string, 0, len(s.Repos))
		for _, repo := range s.Repos {
			items = append(items, repo.Name)
		}
		return items
	case uistate.PaneStreams:
		items := make([]string, 0, len(s.Streams))
		for _, stream := range s.Streams {
			items = append(items, stream.Name)
		}
		return items
	case uistate.PaneIssues:
		if s.View == uistate.ViewDefault {
			scoped := DefaultScopedIssues(s)
			ordered := views.PrioritizedDefaultIssueIndices(scoped, s.Filters[pane], s.Settings)
			items := make([]string, 0, len(ordered))
			for _, idx := range ordered {
				issue := scoped[idx]
				estimate := ""
				if issue.EstimateMinutes != nil {
					estimate = fmt.Sprintf(" %dm", *issue.EstimateMinutes)
				}
				due := helperpkg.IssueScheduleLabel(issue.Issue)
				if due != "" {
					due = " " + due
				}
				items = append(items, fmt.Sprintf("[%s/%s] %s %s%s%s", issue.RepoName, issue.StreamName, issue.Status, issue.Title, estimate, due))
			}
			return items
		}
		if s.View == uistate.ViewDaily {
			items := make([]string, 0, len(DailyScopedIssues(s)))
			for _, issue := range DailyScopedIssues(s) {
				meta := IssueMetaByID(s, issue.ID)
				repoName, streamName := "-", "-"
				if meta != nil {
					repoName = meta.RepoName
					streamName = meta.StreamName
				}
				estimate := ""
				if issue.EstimateMinutes != nil {
					estimate = fmt.Sprintf(" %dm", *issue.EstimateMinutes)
				}
				due := helperpkg.IssueScheduleLabel(issue)
				if due != "" {
					due = " " + due
				}
				items = append(items, fmt.Sprintf("[%s/%s] %s %s%s%s", repoName, streamName, issue.Status, issue.Title, estimate, due))
			}
			return items
		}
		items := make([]string, 0, len(s.Issues))
		for _, issue := range s.Issues {
			due := helperpkg.IssueScheduleLabel(issue)
			if due != "" {
				due = " " + due
			}
			items = append(items, fmt.Sprintf("%s %s%s", issue.Status, issue.Title, due))
		}
		return items
	case uistate.PaneHabits:
		if s.View == uistate.ViewDaily {
			items := make([]string, 0, len(FilteredDueHabits(s)))
			for _, habit := range FilteredDueHabits(s) {
				items = append(items, fmt.Sprintf("[%s/%s] %s", habit.RepoName, habit.StreamName, habit.Name))
			}
			return items
		}
		items := make([]string, 0, len(s.Habits))
		for _, habit := range s.Habits {
			items = append(items, habit.Name)
		}
		return items
	case uistate.PaneScratchpads:
		items := make([]string, 0, len(s.Scratchpads))
		for _, scratchpad := range s.Scratchpads {
			items = append(items, scratchpad.Name)
		}
		return items
	case uistate.PaneConfig:
		items := make([]string, 0, len(s.ConfigItems))
		for _, item := range s.ConfigItems {
			items = append(items, item.Label+" "+item.Value)
		}
		return items
	case uistate.PaneExportReports:
		items := make([]string, 0, len(s.ExportReports))
		for _, report := range s.ExportReports {
			items = append(items, fmt.Sprintf("%s  [%s] %s", report.Date, report.Format, report.Name))
		}
		return items
	case uistate.PaneSessions:
		items := make([]string, 0, len(s.SessionHistory))
		for _, session := range s.SessionHistory {
			items = append(items, helperpkg.SessionHistorySummary(session))
		}
		return items
	case uistate.PaneOps:
		items := make([]string, 0, len(s.Ops))
		for _, op := range s.Ops {
			ts := op.Timestamp
			if len(ts) >= 19 {
				ts = strings.Replace(ts[:19], "T", " ", 1)
			}
			items = append(items, fmt.Sprintf("%s %s.%s %s", ts, op.Entity, op.Action, op.EntityID))
		}
		return items
	case uistate.PaneSettings:
		return views.SettingsItemLabels(s.Settings)
	default:
		return nil
	}
}

func FilteredIndices(s Snapshot, pane uistate.Pane) []int {
	if pane == uistate.PaneIssues && s.View == uistate.ViewDefault {
		return views.PrioritizedDefaultIssueIndices(DefaultScopedIssues(s), s.Filters[pane], s.Settings)
	}
	items := PaneItems(s, pane)
	query := strings.TrimSpace(strings.ToLower(s.Filters[pane]))
	if query == "" {
		indices := make([]int, len(items))
		for i := range items {
			indices[i] = i
		}
		return indices
	}
	indices := make([]int, 0, len(items))
	for i, item := range items {
		if strings.Contains(strings.ToLower(item), query) {
			indices = append(indices, i)
		}
	}
	return indices
}

func FilteredIndexAtCursor(s Snapshot, pane uistate.Pane) int {
	indices := FilteredIndices(s, pane)
	cur := s.Cursors[pane]
	if cur < 0 || cur >= len(indices) {
		return -1
	}
	return indices[cur]
}

func FilteredCursorForRawIndex(s Snapshot, pane uistate.Pane, rawIdx int) int {
	for i, idx := range FilteredIndices(s, pane) {
		if idx == rawIdx {
			return i
		}
	}
	return -1
}

func SelectedMetaRepo(s Snapshot) (int64, string, bool) {
	if s.View != uistate.ViewMeta {
		return 0, "", false
	}
	rawIdx := FilteredIndexAtCursor(s, uistate.PaneRepos)
	if rawIdx >= 0 && rawIdx < len(s.Repos) {
		return s.Repos[rawIdx].ID, s.Repos[rawIdx].Name, true
	}
	if s.Context != nil && s.Context.RepoID != nil {
		return *s.Context.RepoID, helperpkg.FirstNonEmpty(s.Context.RepoName, nil), true
	}
	return 0, "", false
}

func SelectedMetaStream(s Snapshot) (int64, string, string, bool) {
	if s.View != uistate.ViewMeta {
		return 0, "", "", false
	}
	rawIdx := FilteredIndexAtCursor(s, uistate.PaneStreams)
	if rawIdx >= 0 && rawIdx < len(s.Streams) {
		stream := s.Streams[rawIdx]
		return stream.ID, stream.Name, RepoNameByID(s, stream.RepoID), true
	}
	if s.Context != nil && s.Context.StreamID != nil {
		repoName := "-"
		if s.Context.RepoName != nil {
			repoName = *s.Context.RepoName
		}
		return *s.Context.StreamID, helperpkg.FirstNonEmpty(s.Context.StreamName, nil), repoName, true
	}
	return 0, "", "", false
}

func SelectedIssue(s Snapshot) (*api.Issue, bool) {
	if s.PreferActiveIssue && s.ActiveIssue != nil {
		copy := s.ActiveIssue.Issue
		return &copy, true
	}
	if s.Pane != uistate.PaneIssues {
		return nil, false
	}
	switch s.View {
	case uistate.ViewDefault:
		scoped := DefaultScopedIssues(s)
		rawIdx := FilteredIndexAtCursor(s, uistate.PaneIssues)
		if rawIdx < 0 || rawIdx >= len(scoped) {
			return nil, false
		}
		copy := scoped[rawIdx].Issue
		return &copy, true
	case uistate.ViewDaily:
		issues := DailyScopedIssues(s)
		rawIdx := FilteredIndexAtCursor(s, uistate.PaneIssues)
		if rawIdx < 0 || rawIdx >= len(issues) {
			return nil, false
		}
		copy := issues[rawIdx]
		return &copy, true
	case uistate.ViewMeta:
		rawIdx := FilteredIndexAtCursor(s, uistate.PaneIssues)
		if rawIdx < 0 || rawIdx >= len(s.Issues) {
			return nil, false
		}
		copy := s.Issues[rawIdx]
		return &copy, true
	default:
		return nil, false
	}
}

func SelectedIssueDetail(s Snapshot) (IssueSelection, bool) {
	issue, ok := SelectedIssue(s)
	if !ok {
		return IssueSelection{}, false
	}
	return IssueSelection{
		ID:          issue.ID,
		StreamID:    issue.StreamID,
		Status:      string(issue.Status),
		TodoForDate: issue.TodoForDate,
	}, true
}

func SelectedHabit(s Snapshot) (*api.Habit, bool) {
	if s.Pane != uistate.PaneHabits || s.View != uistate.ViewMeta {
		return nil, false
	}
	rawIdx := FilteredIndexAtCursor(s, uistate.PaneHabits)
	if rawIdx < 0 || rawIdx >= len(s.Habits) {
		return nil, false
	}
	copy := s.Habits[rawIdx]
	return &copy, true
}

func SelectedDailyHabit(s Snapshot) (*api.HabitDailyItem, bool) {
	if s.View != uistate.ViewDaily || s.Pane != uistate.PaneHabits {
		return nil, false
	}
	habits := FilteredDueHabits(s)
	rawIdx := FilteredIndexAtCursor(s, uistate.PaneHabits)
	if rawIdx < 0 || rawIdx >= len(habits) {
		return nil, false
	}
	copy := habits[rawIdx]
	return &copy, true
}

func SelectedSessionHistoryEntry(s Snapshot) (*api.SessionHistoryEntry, bool) {
	if s.View != uistate.ViewSessionHistory || s.Pane != uistate.PaneSessions {
		return nil, false
	}
	rawIdx := FilteredIndexAtCursor(s, uistate.PaneSessions)
	if rawIdx < 0 || rawIdx >= len(s.SessionHistory) {
		return nil, false
	}
	copy := s.SessionHistory[rawIdx]
	return &copy, true
}

func SelectedConfigItem(s Snapshot) (configitems.Item, bool) {
	if s.View != uistate.ViewConfig || s.Pane != uistate.PaneConfig {
		return configitems.Item{}, false
	}
	rawIdx := FilteredIndexAtCursor(s, uistate.PaneConfig)
	if rawIdx < 0 || rawIdx >= len(s.ConfigItems) {
		return configitems.Item{}, false
	}
	return s.ConfigItems[rawIdx], true
}

func SelectedExportReport(s Snapshot) (api.ExportReportFile, bool) {
	if s.View != uistate.ViewReports || s.Pane != uistate.PaneExportReports {
		return api.ExportReportFile{}, false
	}
	rawIdx := FilteredIndexAtCursor(s, uistate.PaneExportReports)
	if rawIdx < 0 || rawIdx >= len(s.ExportReports) {
		return api.ExportReportFile{}, false
	}
	return s.ExportReports[rawIdx], true
}

func RepoNameByID(s Snapshot, repoID int64) string {
	for _, repo := range s.Repos {
		if repo.ID == repoID {
			return repo.Name
		}
	}
	return ""
}
