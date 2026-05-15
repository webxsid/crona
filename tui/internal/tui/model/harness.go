package model

import (
	"crona/tui/internal/api"

	appcmd "crona/tui/internal/tui/commands"
	layoutpkg "crona/tui/internal/tui/layout"
	selectionpkg "crona/tui/internal/tui/selection"

	tea "github.com/charmbracelet/bubbletea"
)

func NewDailyRenderModel(width, height int) Model {
	repoName := "Work"
	streamName := "app"
	estimate := 60
	target := 15

	model := Model{
		view:   ViewDaily,
		pane:   PaneIssues,
		width:  width,
		height: height,
		cursor: map[Pane]int{PaneIssues: 0, PaneHabits: 0},
		filters: map[Pane]string{
			PaneIssues: "",
			PaneHabits: "",
		},
		context: &api.ActiveContext{
			RepoName:   &repoName,
			StreamName: &streamName,
		},
		kernelInfo: &api.KernelInfo{Env: "Dev"},
		dailySummary: &api.DailyIssueSummary{
			Date: "2026-03-19",
			Issues: []api.Issue{
				{ID: 1, Title: "Add keyboard-first command palette", Status: "planned", EstimateMinutes: &estimate},
			},
		},
		allIssues: []api.IssueWithMeta{
			{
				Issue:      api.Issue{ID: 1, Title: "Add keyboard-first command palette", Status: "planned", EstimateMinutes: &estimate},
				RepoName:   "Work",
				StreamName: "app",
			},
		},
		dueHabits: []api.HabitDailyItem{
			{
				HabitWithMeta: api.HabitWithMeta{
					Habit: api.Habit{Name: "Inbox Zero Sweep", TargetMinutes: &target},
				},
				Status: "pending",
			},
		},
	}
	return model
}

func (m Model) RenderString() string { return m.View() }
func (m Model) BodyHeight() int      { return m.ContentHeight() }
func (m Model) ContentHeight() int   { return m.contentHeight() }

func MinimumSize() (int, int) {
	return layoutpkg.MinWidth, layoutpkg.MinHeight
}

func NewDailyHabitDeleteModel(habits []api.HabitDailyItem) Model {
	model := Model{
		view:      ViewDaily,
		pane:      PaneHabits,
		cursor:    map[Pane]int{PaneHabits: 0},
		filters:   map[Pane]string{PaneHabits: ""},
		dueHabits: habits,
		timer:     &api.TimerState{State: "idle"},
	}
	return model
}

func OpenSelectedDeleteDialog(m Model) (Model, bool) {
	return m.openSelectedDeleteDialog()
}

func (m Model) DialogDeleteKind() string { return m.dialogDeleteKind }
func (m Model) DialogDeleteID() string   { return m.dialogDeleteID }
func (m Model) DialogStreamID() int64    { return m.dialogStreamID }
func (m Model) DialogKind() string       { return m.dialog }
func (m Model) DialogSessionID() string  { return m.dialogSessionID }

func NewDefaultScopeModel(allIssues []api.IssueWithMeta, context *api.ActiveContext) Model {
	model := Model{
		allIssues: allIssues,
		context:   context,
	}
	return model
}

func DefaultScopedIssuesForTest(m Model) []api.IssueWithMeta {
	snapshot := m.selectionSnapshot()
	return selectionpkg.DefaultScopedIssues(snapshot)
}

func (m Model) AllIssuesForTest() []api.IssueWithMeta {
	return append([]api.IssueWithMeta(nil), m.allIssues...)
}

func OpenCreateIssueDefaultDialogForTest(m Model) Model {
	return m.openCreateIssueDefaultDialog()
}

func OpenCheckoutContextDialogForTest(m Model) Model {
	return m.openCheckoutContextDialog()
}

func (m Model) DialogInputValue(index int) string {
	if index < 0 || index >= len(m.dialogInputs) {
		return ""
	}
	return m.dialogInputs[index].Value()
}

func (m Model) DialogFocusIndex() int { return m.dialogFocusIdx }

func NewUpdateViewModel(view View, pane Pane, statusMsg string, updateStatus *api.UpdateStatus, executablePath string, kernelInfo *api.KernelInfo) Model {
	model := Model{
		view:                  view,
		pane:                  pane,
		statusMsg:             statusMsg,
		updateStatus:          updateStatus,
		currentExecutablePath: executablePath,
		kernelInfo:            kernelInfo,
	}
	return model
}

func AvailableViewsForTest(m Model) []View { return m.availableViews() }

func ApplyUpdateStatusLoadedForTest(m Model, status *api.UpdateStatus) (Model, tea.Cmd) {
	next, cmd := m.Update(appcmd.UpdateStatusLoadedMsg{Status: status})
	return next.(Model), cmd
}

func ApplyUpdateDismissedForTest(m Model, status *api.UpdateStatus) (Model, tea.Cmd) {
	next, cmd := m.Update(appcmd.UpdateDismissedMsg{Status: status})
	return next.(Model), cmd
}

func (m Model) StatusMessage() string { return m.statusMsg }
func (m Model) CurrentView() View     { return m.view }
func (m Model) SelfUpdateUnsupportedReasonForTest() string {
	return m.selfUpdateUnsupportedReason()
}
func (m Model) SelfUpdateInstallAvailableForTest() bool {
	return m.selfUpdateInstallAvailable()
}

func NewSessionDetailModel(detail *api.SessionDetail) Model {
	return Model{
		view:              ViewSessionHistory,
		pane:              PaneSessions,
		sessionDetailOpen: true,
		sessionDetail:     detail,
		cursor:            map[Pane]int{PaneSessions: 0},
		filters:           map[Pane]string{PaneSessions: ""},
	}
}
