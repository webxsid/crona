package input

import (
	"strings"

	sharedtypes "crona/shared/types"
	uistate "crona/tui/internal/tui/state"

	tea "github.com/charmbracelet/bubbletea"
)

func handleAdjustSelectedSetting(s State, deps Deps, dir int) (tea.Model, tea.Cmd, bool) {
	if s.ActiveView != uistate.ViewSettings || s.Settings == nil {
		return s, nil, true
	}
	repoID, streamID := int64(0), int64(0)
	if s.Context != nil && s.Context.RepoID != nil {
		repoID = *s.Context.RepoID
	}
	if s.Context != nil && s.Context.StreamID != nil {
		streamID = *s.Context.StreamID
	}
	rawIdx := s.Cursor[uistate.PaneSettings]
	switch rawIdx {
	case 0:
		next := sharedtypes.TimerModeStructured
		if s.Settings.TimerMode == sharedtypes.TimerModeStructured {
			next = sharedtypes.TimerModeStopwatch
		}
		return s, deps.PatchSetting(sharedtypes.CoreSettingsKeyTimerMode, next, repoID, streamID, s.DashboardDate), true
	case 1:
		return s, deps.PatchSetting(sharedtypes.CoreSettingsKeyWorkDurationMinutes, clampMin(s.Settings.WorkDurationMinutes+dir*5, 5), repoID, streamID, s.DashboardDate), true
	case 2:
		return s, deps.PatchSetting(sharedtypes.CoreSettingsKeyBreaksEnabled, !s.Settings.BreaksEnabled, repoID, streamID, s.DashboardDate), true
	case 3:
		return s, deps.PatchSetting(sharedtypes.CoreSettingsKeyShortBreakMinutes, clampMin(s.Settings.ShortBreakMinutes+dir, 1), repoID, streamID, s.DashboardDate), true
	case 4:
		return s, deps.PatchSetting(sharedtypes.CoreSettingsKeyLongBreakMinutes, clampMin(s.Settings.LongBreakMinutes+dir*5, 5), repoID, streamID, s.DashboardDate), true
	case 5:
		return s, deps.PatchSetting(sharedtypes.CoreSettingsKeyLongBreakEnabled, !s.Settings.LongBreakEnabled, repoID, streamID, s.DashboardDate), true
	case 6:
		return s, deps.PatchSetting(sharedtypes.CoreSettingsKeyCyclesBeforeLongBreak, clampMin(s.Settings.CyclesBeforeLongBreak+dir, 1), repoID, streamID, s.DashboardDate), true
	case 7:
		return s, deps.PatchSetting(sharedtypes.CoreSettingsKeyAutoStartBreaks, !s.Settings.AutoStartBreaks, repoID, streamID, s.DashboardDate), true
	case 8:
		return s, deps.PatchSetting(sharedtypes.CoreSettingsKeyAutoStartWork, !s.Settings.AutoStartWork, repoID, streamID, s.DashboardDate), true
	case 9:
		return s, deps.PatchSetting(sharedtypes.CoreSettingsKeyBoundaryNotifications, !s.Settings.BoundaryNotifications, repoID, streamID, s.DashboardDate), true
	case 10:
		return s, deps.PatchSetting(sharedtypes.CoreSettingsKeyBoundarySound, !s.Settings.BoundarySound, repoID, streamID, s.DashboardDate), true
	case 11:
		return s, deps.PatchSetting(sharedtypes.CoreSettingsKeyUpdateChecksEnabled, !s.Settings.UpdateChecksEnabled, repoID, streamID, s.DashboardDate), true
	case 12:
		return s, deps.PatchSetting(sharedtypes.CoreSettingsKeyUpdatePromptEnabled, !s.Settings.UpdatePromptEnabled, repoID, streamID, s.DashboardDate), true
	case 13:
		return s, deps.PatchSetting(sharedtypes.CoreSettingsKeyUpdateChannel, nextUpdateChannel(s.Settings.UpdateChannel, dir), repoID, streamID, s.DashboardDate), true
	case 14:
		return s, deps.PatchSetting(sharedtypes.CoreSettingsKeyRepoSort, nextRepoSort(s.Settings.RepoSort, dir), repoID, streamID, s.DashboardDate), true
	case 15:
		return s, deps.PatchSetting(sharedtypes.CoreSettingsKeyStreamSort, nextStreamSort(s.Settings.StreamSort, dir), repoID, streamID, s.DashboardDate), true
	case 16:
		return s, deps.PatchSetting(sharedtypes.CoreSettingsKeyIssueSort, nextIssueSort(s.Settings.IssueSort, dir), repoID, streamID, s.DashboardDate), true
	case 17:
		return s, deps.PatchSetting(sharedtypes.CoreSettingsKeyHabitSort, nextHabitSort(s.Settings.HabitSort, dir), repoID, streamID, s.DashboardDate), true
	case 18:
		return s, deps.PatchSetting(sharedtypes.CoreSettingsKeyAwayModeEnabled, !s.Settings.AwayModeEnabled, repoID, streamID, s.DashboardDate), true
	case 19:
		return s, deps.PatchSetting(sharedtypes.CoreSettingsKeyDailyPlanRollbackMins, clampMin(currentRollbackMinutes(s.Settings.DailyPlanRollbackMins)+dir, 1), repoID, streamID, s.DashboardDate), true
	default:
		return s, nil, true
	}
}

func handleActivateSelectedSetting(s State, deps Deps) (tea.Model, tea.Cmd, bool) {
	if s.ActiveView != uistate.ViewSettings || s.Settings == nil {
		return s, nil, true
	}
	switch s.Cursor[uistate.PaneSettings] {
	case 20:
		deps.OpenEditRestProtectionDialog(&s)
		return s, nil, true
	case 21:
		deps.OpenConfirmWipeDataDialog(&s)
		return s, nil, true
	case 22:
		deps.OpenConfirmUninstallDialog(&s)
		return s, nil, true
	default:
		return handleAdjustSelectedSetting(s, deps, 1)
	}
}

func handleToggleAwayMode(s State, deps Deps) (tea.Model, tea.Cmd, bool) {
	if (s.ActiveView != uistate.ViewDaily && s.ActiveView != uistate.ViewWellbeing && s.ActiveView != uistate.ViewAway) || s.Settings == nil {
		return s, nil, false
	}
	repoID := int64(0)
	if s.Context != nil && s.Context.RepoID != nil {
		repoID = *s.Context.RepoID
	}
	streamID := int64(0)
	if s.Context != nil && s.Context.StreamID != nil {
		streamID = *s.Context.StreamID
	}
	next := !s.Settings.AwayModeEnabled
	status := "Away mode disabled"
	if next {
		status = "Away mode enabled"
	}
	settingsCopy := *s.Settings
	settingsCopy.AwayModeEnabled = next
	s.Settings = &settingsCopy
	if s.ActiveView == uistate.ViewAway && !next {
		s.ActiveView = uistate.ViewDaily
		s.ActivePane = uistate.DefaultPane(s.ActiveView)
	}
	date := s.DashboardDate
	if s.ActiveView == uistate.ViewWellbeing && strings.TrimSpace(s.WellbeingDate) != "" {
		date = s.WellbeingDate
	}
	return s, tea.Batch(
		deps.PatchSetting(sharedtypes.CoreSettingsKeyAwayModeEnabled, next, repoID, streamID, date),
		deps.SetStatus(&s, status, false),
	), true
}

func clampMin(value, min int) int {
	if value < min {
		return min
	}
	return value
}

func currentRollbackMinutes(value int) int {
	if value <= 0 {
		return 5
	}
	return value
}

func nextRepoSort(current sharedtypes.RepoSort, dir int) sharedtypes.RepoSort {
	options := []sharedtypes.RepoSort{
		sharedtypes.RepoSortAlphabeticalAsc,
		sharedtypes.RepoSortAlphabeticalDesc,
		sharedtypes.RepoSortChronologicalAsc,
		sharedtypes.RepoSortChronologicalDesc,
	}
	return options[nextIndex(current, options, dir)]
}

func nextStreamSort(current sharedtypes.StreamSort, dir int) sharedtypes.StreamSort {
	options := []sharedtypes.StreamSort{
		sharedtypes.StreamSortAlphabeticalAsc,
		sharedtypes.StreamSortAlphabeticalDesc,
		sharedtypes.StreamSortChronologicalAsc,
		sharedtypes.StreamSortChronologicalDesc,
	}
	return options[nextIndex(current, options, dir)]
}

func nextIssueSort(current sharedtypes.IssueSort, dir int) sharedtypes.IssueSort {
	options := []sharedtypes.IssueSort{
		sharedtypes.IssueSortPriority,
		sharedtypes.IssueSortDueDateAsc,
		sharedtypes.IssueSortDueDateDesc,
		sharedtypes.IssueSortAlphabeticalAsc,
		sharedtypes.IssueSortAlphabeticalDesc,
		sharedtypes.IssueSortChronologicalAsc,
		sharedtypes.IssueSortChronologicalDesc,
	}
	return options[nextIndex(current, options, dir)]
}

func nextHabitSort(current sharedtypes.HabitSort, dir int) sharedtypes.HabitSort {
	options := []sharedtypes.HabitSort{
		sharedtypes.HabitSortSchedule,
		sharedtypes.HabitSortTargetMinutesAsc,
		sharedtypes.HabitSortTargetMinutesDesc,
		sharedtypes.HabitSortAlphabeticalAsc,
		sharedtypes.HabitSortAlphabeticalDesc,
		sharedtypes.HabitSortChronologicalAsc,
		sharedtypes.HabitSortChronologicalDesc,
	}
	return options[nextIndex(current, options, dir)]
}

func nextUpdateChannel(current sharedtypes.UpdateChannel, dir int) sharedtypes.UpdateChannel {
	options := []sharedtypes.UpdateChannel{
		sharedtypes.UpdateChannelStable,
		sharedtypes.UpdateChannelBeta,
	}
	return options[nextIndex(sharedtypes.NormalizeUpdateChannel(current), options, dir)]
}
