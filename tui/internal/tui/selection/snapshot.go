package selection

import (
	"fmt"
	"strings"

	"crona/tui/internal/api"
	configitems "crona/tui/internal/tui/configitems"
	helperpkg "crona/tui/internal/tui/helpers"
	uistate "crona/tui/internal/tui/state"
	issuecore "crona/tui/internal/tui/views/issuecore"
	settingsmeta "crona/tui/internal/tui/views/settingsmeta"
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

	issueMetaIndex         map[int64]int
	repoNameIndex          map[int64]string
	defaultScopedIssues    []api.IssueWithMeta
	defaultScopedReady     bool
	dailyScopedIssues      []api.Issue
	dailyScopedReady       bool
	filteredDueHabits      []api.HabitDailyItem
	filteredDueHabitsReady bool
	paneItemsCache         map[uistate.Pane][]string
	filteredIndicesCache   map[uistate.Pane][]int
}

type IssueSelection struct {
	ID          int64
	StreamID    int64
	Status      string
	TodoForDate *string
}

func PrepareSnapshot(s Snapshot) Snapshot {
	s.issueMetaIndex = make(map[int64]int, len(s.AllIssues))
	for i := range s.AllIssues {
		s.issueMetaIndex[s.AllIssues[i].ID] = i
	}
	s.repoNameIndex = make(map[int64]string, len(s.Repos))
	for _, repo := range s.Repos {
		s.repoNameIndex[repo.ID] = repo.Name
	}
	s.paneItemsCache = make(map[uistate.Pane][]string)
	s.filteredIndicesCache = make(map[uistate.Pane][]int)
	s.defaultScopedIssues = buildDefaultScopedIssues(s)
	s.defaultScopedReady = true
	s.dailyScopedIssues = buildDailyScopedIssues(s)
	s.dailyScopedReady = true
	s.filteredDueHabits = buildFilteredDueHabits(s)
	s.filteredDueHabitsReady = true
	if s.ActiveIssue == nil {
		var issueID int64
		if s.Timer != nil && s.Timer.IssueID != nil {
			issueID = *s.Timer.IssueID
		} else if s.Context != nil && s.Context.IssueID != nil {
			issueID = *s.Context.IssueID
		}
		if idx, ok := s.issueMetaIndex[issueID]; ok {
			s.ActiveIssue = &s.AllIssues[idx]
		}
	}
	return s
}

func IssueMetaByID(s Snapshot, issueID int64) *api.IssueWithMeta {
	if idx, ok := s.issueMetaIndex[issueID]; ok {
		return &s.AllIssues[idx]
	}
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
	if s.defaultScopedReady {
		return s.defaultScopedIssues
	}
	return buildDefaultScopedIssues(s)
}

func FilteredDueHabits(s Snapshot) []api.HabitDailyItem {
	if s.filteredDueHabitsReady {
		return s.filteredDueHabits
	}
	return buildFilteredDueHabits(s)
}

func DailyScopedIssues(s Snapshot) []api.Issue {
	if s.dailyScopedReady {
		return s.dailyScopedIssues
	}
	return buildDailyScopedIssues(s)
}

func PaneItems(s Snapshot, pane uistate.Pane) []string {
	if items, ok := s.paneItemsCache[pane]; ok {
		return items
	}
	var items []string
	switch pane {
	case uistate.PaneRepos:
		items = make([]string, 0, len(s.Repos))
		for _, repo := range s.Repos {
			items = append(items, repo.Name)
		}
	case uistate.PaneStreams:
		items = make([]string, 0, len(s.Streams))
		for _, stream := range s.Streams {
			items = append(items, stream.Name)
		}
	case uistate.PaneIssues:
		if s.View == uistate.ViewDefault {
			scoped := DefaultScopedIssues(s)
			ordered := issuecore.PrioritizedDefaultIssueIndices(scoped, s.Filters[pane], s.Settings)
			items = make([]string, 0, len(ordered))
			for _, idx := range ordered {
				issue := scoped[idx]
				estimate := ""
				if issue.EstimateMinutes != nil {
					estimate = " " + helperpkg.FormatCompactDurationMinutes(*issue.EstimateMinutes)
				}
				due := helperpkg.IssueScheduleLabel(issue.Issue)
				if due != "" {
					due = " " + due
				}
				items = append(items, fmt.Sprintf("[%s/%s] %s %s%s%s", issue.RepoName, issue.StreamName, issue.Status, issue.Title, estimate, due))
			}
			break
		}
		if s.View == uistate.ViewDaily {
			scoped := DailyScopedIssues(s)
			items = make([]string, 0, len(scoped))
			for _, issue := range scoped {
				meta := IssueMetaByID(s, issue.ID)
				repoName, streamName := "-", "-"
				if meta != nil {
					repoName = meta.RepoName
					streamName = meta.StreamName
				}
				estimate := ""
				if issue.EstimateMinutes != nil {
					estimate = " " + helperpkg.FormatCompactDurationMinutes(*issue.EstimateMinutes)
				}
				due := helperpkg.IssueScheduleLabel(issue)
				if due != "" {
					due = " " + due
				}
				items = append(items, fmt.Sprintf("[%s/%s] %s %s%s%s", repoName, streamName, issue.Status, issue.Title, estimate, due))
			}
			break
		}
		items = make([]string, 0, len(s.Issues))
		for _, issue := range s.Issues {
			due := helperpkg.IssueScheduleLabel(issue)
			if due != "" {
				due = " " + due
			}
			items = append(items, fmt.Sprintf("%s %s%s", issue.Status, issue.Title, due))
		}
	case uistate.PaneHabits:
		if s.View == uistate.ViewDaily {
			habits := FilteredDueHabits(s)
			items = make([]string, 0, len(habits))
			for _, habit := range habits {
				items = append(items, fmt.Sprintf("[%s/%s] %s", habit.RepoName, habit.StreamName, habit.Name))
			}
			break
		}
		items = make([]string, 0, len(s.Habits))
		for _, habit := range s.Habits {
			items = append(items, habit.Name)
		}
	case uistate.PaneScratchpads:
		items = make([]string, 0, len(s.Scratchpads))
		for _, scratchpad := range s.Scratchpads {
			items = append(items, scratchpad.Name)
		}
	case uistate.PaneConfig:
		items = make([]string, 0, len(s.ConfigItems))
		for _, item := range s.ConfigItems {
			items = append(items, item.Label+" "+item.Value)
		}
	case uistate.PaneExportReports:
		items = make([]string, 0, len(s.ExportReports))
		for _, report := range s.ExportReports {
			items = append(items, fmt.Sprintf("%s  [%s] %s", report.Date, report.Format, report.Name))
		}
	case uistate.PaneSessions:
		items = make([]string, 0, len(s.SessionHistory))
		for _, session := range s.SessionHistory {
			items = append(items, helperpkg.SessionHistorySummary(session))
		}
	case uistate.PaneOps:
		items = make([]string, 0, len(s.Ops))
		for _, op := range s.Ops {
			ts := op.Timestamp
			if len(ts) >= 19 {
				ts = strings.Replace(ts[:19], "T", " ", 1)
			}
			items = append(items, fmt.Sprintf("%s %s.%s %s", ts, op.Entity, op.Action, op.EntityID))
		}
	case uistate.PaneSettings:
		items = settingsmeta.ItemLabels(s.Settings)
	default:
		items = nil
	}
	s.paneItemsCache[pane] = items
	return items
}

func FilteredIndices(s Snapshot, pane uistate.Pane) []int {
	if indices, ok := s.filteredIndicesCache[pane]; ok {
		return indices
	}
	var indices []int
	if pane == uistate.PaneIssues && s.View == uistate.ViewDefault {
		indices = issuecore.PrioritizedDefaultIssueIndices(DefaultScopedIssues(s), s.Filters[pane], s.Settings)
	} else {
		items := PaneItems(s, pane)
		query := strings.TrimSpace(strings.ToLower(s.Filters[pane]))
		if query == "" {
			indices = make([]int, len(items))
			for i := range items {
				indices[i] = i
			}
		} else {
			indices = make([]int, 0, len(items))
			for i, item := range items {
				if strings.Contains(strings.ToLower(item), query) {
					indices = append(indices, i)
				}
			}
		}
	}
	s.filteredIndicesCache[pane] = indices
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
	if name, ok := s.repoNameIndex[repoID]; ok {
		return name
	}
	for _, repo := range s.Repos {
		if repo.ID == repoID {
			return repo.Name
		}
	}
	return ""
}

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

func buildDailyScopedIssues(s Snapshot) []api.Issue {
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
