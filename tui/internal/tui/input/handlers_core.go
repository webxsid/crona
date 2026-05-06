package input

import (
	"strings"

	"crona/tui/internal/api"
	uistate "crona/tui/internal/tui/state"

	tea "github.com/charmbracelet/bubbletea"
)

func handleDevCmd(s State, isDev func(State) bool, cmd func() tea.Cmd) (tea.Model, tea.Cmd, bool) {
	if !isDev(s) {
		return s, nil, false
	}
	return s, cmd(), true
}

func handleCycleView(s State, deps Deps, dir int) (tea.Model, tea.Cmd, bool) {
	if s.Timer != nil && s.Timer.State != "idle" {
		s.ActiveView = deps.NextActiveSessionView(s, dir)
	} else {
		s.ActiveView = deps.NextWorkspaceView(s, dir)
	}
	s.ActivePane = deps.DefaultPane(s.ActiveView)
	if s.ActiveView == uistate.ViewHabitHistory && deps.EnterHabitHistoryView != nil {
		cmd, handled := deps.EnterHabitHistoryView(&s)
		if handled {
			return s, cmd, true
		}
	}
	return s, nil, true
}

func handleCyclePane(s State, deps Deps, dir int) (tea.Model, tea.Cmd, bool) {
	if s.ActiveView == uistate.ViewDefault && s.ActivePane == uistate.PaneIssues {
		if dir >= 0 {
			if s.DefaultIssueSection == uistate.DefaultIssueSectionOpen {
				deps.SetDefaultIssueSection(&s, uistate.DefaultIssueSectionCompleted)
			} else {
				deps.SetDefaultIssueSection(&s, uistate.DefaultIssueSectionOpen)
			}
		} else {
			if s.DefaultIssueSection == uistate.DefaultIssueSectionCompleted {
				deps.SetDefaultIssueSection(&s, uistate.DefaultIssueSectionOpen)
			} else {
				deps.SetDefaultIssueSection(&s, uistate.DefaultIssueSectionCompleted)
			}
		}
		return s, nil, true
	}
	if s.ActiveView == uistate.ViewDaily && s.ActivePane == uistate.PaneIssues {
		sections := []uistate.DailyTaskSection{
			uistate.DailyTaskSectionPlanned,
			uistate.DailyTaskSectionPinned,
			uistate.DailyTaskSectionOverdue,
		}
		next := nextIndex(s.DailyTaskSection, sections, dir)
		deps.SetDailyTaskSection(&s, sections[next])
		return s, nil, true
	}
	s.ActivePane = deps.NextPane(s.ActiveView, s.ActivePane, dir)
	return s, nil, true
}

func handleOpenUpdates(s State) (tea.Model, tea.Cmd, bool) {
	s.ActiveView = uistate.ViewUpdates
	s.ActivePane = uistate.DefaultPane(s.ActiveView)
	return s, nil, true
}

func handleOpenViewJump(s State, deps Deps) (tea.Model, tea.Cmd, bool) {
	if deps.OpenViewJumpDialog(&s) {
		return s, nil, true
	}
	return s, nil, false
}

func handleRescanExportAssets(s State, deps Deps) (tea.Model, tea.Cmd, bool) {
	if s.ActiveView != uistate.ViewConfig {
		return s, nil, true
	}
	return s, tea.Batch(deps.SetStatus(&s, "Rescanning export tools...", false), deps.LoadExportAssets()), true
}

func handleCursor(s State, deps Deps, delta int) (tea.Model, tea.Cmd, bool) {
	total := deps.ListLen(s, s.ActivePane)
	if total == 0 {
		return s, nil, true
	}
	next := s.Cursor[s.ActivePane] + delta
	if next < 0 {
		next = 0
	}
	if next >= total {
		next = total - 1
	}
	s.Cursor[s.ActivePane] = next
	return s, nil, true
}

func handleIssueStatus(s State, deps Deps) (tea.Model, tea.Cmd, bool) {
	if s.Timer != nil && s.Timer.State != "idle" && (s.ActiveView == uistate.ViewSessionActive || s.ActiveView == uistate.ViewScratch) {
		return s, deps.SetStatus(&s, "End or stash the active session before changing issue status", true), true
	}
	deps.OpenIssueStatusFromSelection(&s)
	return s, nil, true
}

func handleContextCheckout(s State, deps Deps) (tea.Model, tea.Cmd, bool) {
	if shouldOpenContextDialog(s.ActiveView) {
		deps.OpenCheckoutContextDialog(&s)
		return s, nil, true
	}
	return s, deps.Checkout(&s), true
}

func handleOpenStashList(s State, deps Deps) (tea.Model, tea.Cmd, bool) {
	if !deps.CanOpenStashList(s) {
		return s, nil, false
	}
	deps.OpenStashListDialog(&s)
	return s, deps.LoadStashes(), true
}

func handleAdjustOpsLimit(s State, deps Deps, delta int) (tea.Model, tea.Cmd, bool) {
	if s.ActivePane != uistate.PaneOps {
		return s, nil, false
	}
	s.OpsLimitPinned = true
	s.OpsLimit += delta
	if s.OpsLimit < 10 {
		s.OpsLimit = 10
	}
	if delta < 0 {
		deps.ClampFiltered(&s, uistate.PaneOps)
	}
	return s, deps.LoadOps(deps.CurrentOpsLimit(s)), true
}

func handleStartFilter(s State, deps Deps) (tea.Model, tea.Cmd, bool) {
	switch s.ActivePane {
	case uistate.PaneOps, uistate.PaneIssues, uistate.PaneHabits, uistate.PaneRepos, uistate.PaneStreams, uistate.PaneScratchpads, uistate.PaneSessions, uistate.PaneConfig, uistate.PaneExportReports, uistate.PaneAlerts:
	default:
		return s, nil, false
	}
	if s.ActivePane == uistate.PaneScratchpads && s.ScratchpadOpen {
		return s, nil, true
	}
	deps.StartFilterEdit(&s, s.ActivePane)
	return s, nil, true
}

func handleSpace(s State, deps Deps) (tea.Model, tea.Cmd, bool) {
	if s.ActiveView == uistate.ViewSettings {
		return handleActivateSelectedSetting(s, deps)
	}
	if s.ActiveView == uistate.ViewAlerts {
		if model, cmd, handled := handleToggleSelectedAlertReminder(s, deps); handled {
			return model, cmd, true
		}
		return handleActivateSelectedAlert(s, deps)
	}
	cmd, handled := deps.ToggleHabitCompletedAction(&s)
	return s, cmd, handled
}

func handleOpenExportDaily(s State, deps Deps) (tea.Model, tea.Cmd, bool) {
	if !supportsGlobalExport(s.ActiveView) {
		return s, nil, false
	}
	if s.Dialog != "" {
		return s, nil, false
	}
	deps.OpenExportDailyDialog(&s)
	return s, nil, true
}

func shouldOpenContextDialog(view uistate.View) bool {
	switch view {
	case uistate.ViewDefault,
		uistate.ViewDaily,
		uistate.ViewMeta,
		uistate.ViewRollup,
		uistate.ViewWellbeing,
		uistate.ViewReports,
		uistate.ViewOps,
		uistate.ViewScratch,
		uistate.ViewSessionHistory,
		uistate.ViewHabitHistory:
		return true
	default:
		return false
	}
}

func supportsGlobalExport(view uistate.View) bool {
	switch view {
	case uistate.ViewDefault,
		uistate.ViewDaily,
		uistate.ViewMeta,
		uistate.ViewWellbeing,
		uistate.ViewReports:
		return true
	default:
		return false
	}
}

func viewsShouldShowUpdate(status *api.UpdateStatus) bool {
	if status == nil {
		return false
	}
	if !status.Enabled || !status.PromptEnabled || !status.UpdateAvailable {
		return false
	}
	return strings.TrimSpace(status.LatestVersion) != "" && strings.TrimSpace(status.LatestVersion) != strings.TrimSpace(status.DismissedVersion)
}

func nextIndex[T comparable](current T, options []T, dir int) int {
	if len(options) == 0 {
		return 0
	}
	index := 0
	for i, option := range options {
		if option == current {
			index = i
			break
		}
	}
	index += dir
	if index < 0 {
		index = len(options) - 1
	}
	if index >= len(options) {
		index = 0
	}
	return index
}
