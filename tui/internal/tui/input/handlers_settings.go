package input

import (
	"strings"

	sharedtypes "crona/shared/types"
	uistate "crona/tui/internal/tui/state"
	alertsmeta "crona/tui/internal/tui/views/alertsmeta"

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
		return s, deps.PatchSetting(sharedtypes.CoreSettingsKeyUpdateChecksEnabled, !s.Settings.UpdateChecksEnabled, repoID, streamID, s.DashboardDate), true
	case 10:
		return s, deps.PatchSetting(sharedtypes.CoreSettingsKeyUpdatePromptEnabled, !s.Settings.UpdatePromptEnabled, repoID, streamID, s.DashboardDate), true
	case 11:
		return s, deps.PatchSetting(sharedtypes.CoreSettingsKeyUpdateChannel, nextUpdateChannel(s.Settings.UpdateChannel, dir), repoID, streamID, s.DashboardDate), true
	case 12:
		return s, deps.PatchSetting(sharedtypes.CoreSettingsKeyRepoSort, nextRepoSort(s.Settings.RepoSort, dir), repoID, streamID, s.DashboardDate), true
	case 13:
		return s, deps.PatchSetting(sharedtypes.CoreSettingsKeyStreamSort, nextStreamSort(s.Settings.StreamSort, dir), repoID, streamID, s.DashboardDate), true
	case 14:
		return s, deps.PatchSetting(sharedtypes.CoreSettingsKeyIssueSort, nextIssueSort(s.Settings.IssueSort, dir), repoID, streamID, s.DashboardDate), true
	case 15:
		return s, deps.PatchSetting(sharedtypes.CoreSettingsKeyHabitSort, nextHabitSort(s.Settings.HabitSort, dir), repoID, streamID, s.DashboardDate), true
	case 16:
		return s, deps.PatchSetting(sharedtypes.CoreSettingsKeyDateDisplayPreset, nextDateDisplayPreset(s.Settings.DateDisplayPreset, dir), repoID, streamID, s.DashboardDate), true
	case 17:
		return s, nil, true
	case 18:
		return s, nil, true
	case 19:
		return s, deps.PatchSetting(sharedtypes.CoreSettingsKeyPromptGlyphMode, nextPromptGlyphMode(s.Settings.PromptGlyphMode, dir), repoID, streamID, s.DashboardDate), true
	case 20:
		return s, nil, true
	case 21:
		return s, deps.PatchSetting(sharedtypes.CoreSettingsKeyAwayModeEnabled, !s.Settings.AwayModeEnabled, repoID, streamID, s.DashboardDate), true
	case 22:
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
	case 17:
		if deps.OpenEditDateDisplayFormatDialog != nil {
			deps.OpenEditDateDisplayFormatDialog(&s)
		}
		return s, nil, true
	case 20:
		if deps.OpenEditHabitStreaksDialog != nil {
			deps.OpenEditHabitStreaksDialog(&s)
		}
		return s, nil, true
	case 23:
		deps.OpenEditRestProtectionDialog(&s)
		return s, nil, true
	case 24:
		deps.OpenConfirmWipeDataDialog(&s)
		return s, nil, true
	case 25:
		deps.OpenConfirmUninstallDialog(&s)
		return s, nil, true
	default:
		return handleAdjustSelectedSetting(s, deps, 1)
	}
}

func handleAdjustSelectedAlert(s State, deps Deps, dir int) (tea.Model, tea.Cmd, bool) {
	if s.ActiveView != uistate.ViewAlerts || s.Settings == nil {
		return s, nil, true
	}
	repoID, streamID := int64(0), int64(0)
	if s.Context != nil && s.Context.RepoID != nil {
		repoID = *s.Context.RepoID
	}
	if s.Context != nil && s.Context.StreamID != nil {
		streamID = *s.Context.StreamID
	}
	row, ok := alertsmeta.SelectedRow(s.FilterState(uistate.PaneAlerts), s.Settings, s.AlertStatus, s.AlertReminders, s.Cursor[uistate.PaneAlerts])
	if !ok {
		deps.ClampFiltered(&s, uistate.PaneAlerts)
		return s, nil, true
	}
	switch row.Key {
	case alertsmeta.RowNotifications:
		return s, deps.PatchSetting(sharedtypes.CoreSettingsKeyBoundaryNotifications, !s.Settings.BoundaryNotifications, repoID, streamID, s.DashboardDate), true
	case alertsmeta.RowSound:
		return s, deps.PatchSetting(sharedtypes.CoreSettingsKeyBoundarySound, !s.Settings.BoundarySound, repoID, streamID, s.DashboardDate), true
	case alertsmeta.RowSoundPreset:
		return s, deps.PatchSetting(sharedtypes.CoreSettingsKeyAlertSoundPreset, nextAlertSoundPreset(s.Settings.AlertSoundPreset, dir), repoID, streamID, s.DashboardDate), true
	case alertsmeta.RowUrgency:
		return s, deps.PatchSetting(sharedtypes.CoreSettingsKeyAlertUrgency, nextAlertUrgency(s.Settings.AlertUrgency, dir), repoID, streamID, s.DashboardDate), true
	case alertsmeta.RowLogoIcon:
		return s, deps.PatchSetting(sharedtypes.CoreSettingsKeyAlertIconEnabled, !s.Settings.AlertIconEnabled, repoID, streamID, s.DashboardDate), true
	case alertsmeta.RowInactivityAlerts:
		return s, deps.PatchSetting(sharedtypes.CoreSettingsKeyInactivityAlerts, !s.Settings.InactivityAlerts, repoID, streamID, s.DashboardDate), true
	case alertsmeta.RowInactivityAfter:
		return s, deps.PatchSetting(sharedtypes.CoreSettingsKeyInactivityThreshold, clampAlertMinutes(s.Settings.InactivityThreshold+dir*15), repoID, streamID, s.DashboardDate), true
	case alertsmeta.RowInactivityRepeat:
		return s, deps.PatchSetting(sharedtypes.CoreSettingsKeyInactivityRepeat, clampAlertMinutes(s.Settings.InactivityRepeat+dir*15), repoID, streamID, s.DashboardDate), true
	case alertsmeta.RowTestNotification:
		return s, deps.TestAlertNotification(), true
	case alertsmeta.RowTestSound:
		return s, deps.TestAlertSound(), true
	case alertsmeta.RowAddReminder:
		if deps.OpenCreateAlertReminderDialog != nil {
			deps.OpenCreateAlertReminderDialog(&s)
		}
		return s, nil, true
	default:
		if row.ReminderID != "" {
			if dir == 0 && deps.OpenEditAlertReminderDialog != nil {
				deps.OpenEditAlertReminderDialog(&s, row.ReminderID)
				return s, nil, true
			}
			return s, nil, true
		}
		return s, nil, true
	}
}

func handleActivateSelectedAlert(s State, deps Deps) (tea.Model, tea.Cmd, bool) {
	if s.ActiveView != uistate.ViewAlerts || s.Settings == nil {
		return s, nil, true
	}
	row, ok := alertsmeta.SelectedRow(s.FilterState(uistate.PaneAlerts), s.Settings, s.AlertStatus, s.AlertReminders, s.Cursor[uistate.PaneAlerts])
	if !ok {
		deps.ClampFiltered(&s, uistate.PaneAlerts)
		return s, nil, true
	}
	if row.ReminderID != "" {
		if deps.OpenEditAlertReminderDialog != nil {
			deps.OpenEditAlertReminderDialog(&s, row.ReminderID)
		}
		return s, nil, true
	}
	return handleAdjustSelectedAlert(s, deps, 1)
}

func handleToggleSelectedAlertReminder(s State, deps Deps) (tea.Model, tea.Cmd, bool) {
	if s.ActiveView != uistate.ViewAlerts || s.Settings == nil {
		return s, nil, false
	}
	row, ok := alertsmeta.SelectedRow(s.FilterState(uistate.PaneAlerts), s.Settings, s.AlertStatus, s.AlertReminders, s.Cursor[uistate.PaneAlerts])
	if !ok || row.ReminderID == "" || deps.ToggleAlertReminder == nil {
		return s, nil, false
	}
	for _, reminder := range s.AlertReminders {
		if reminder.ID == row.ReminderID {
			return s, deps.ToggleAlertReminder(reminder.ID, !reminder.Enabled), true
		}
	}
	return s, nil, false
}

func handleDeleteSelectedAlertReminder(s State, deps Deps) (tea.Model, tea.Cmd, bool) {
	if s.ActiveView != uistate.ViewAlerts || deps.DeleteAlertReminder == nil {
		return s, nil, false
	}
	row, ok := alertsmeta.SelectedRow(s.FilterState(uistate.PaneAlerts), s.Settings, s.AlertStatus, s.AlertReminders, s.Cursor[uistate.PaneAlerts])
	if !ok || row.ReminderID == "" {
		return s, nil, false
	}
	return s, deps.DeleteAlertReminder(row.ReminderID), true
}

func (s State) FilterState(pane uistate.Pane) string {
	if s.Filters == nil {
		return ""
	}
	return s.Filters[pane]
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

func clampAlertMinutes(value int) int {
	if value < 15 {
		return 15
	}
	if value > 720 {
		return 720
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

func nextDateDisplayPreset(current sharedtypes.DateDisplayPreset, dir int) sharedtypes.DateDisplayPreset {
	options := []sharedtypes.DateDisplayPreset{
		sharedtypes.DateDisplayPresetISO,
		sharedtypes.DateDisplayPresetUS,
		sharedtypes.DateDisplayPresetEurope,
		sharedtypes.DateDisplayPresetLong,
		sharedtypes.DateDisplayPresetCustom,
	}
	index := 0
	normalized := sharedtypes.NormalizeDateDisplayPreset(current)
	for i, option := range options {
		if option == normalized {
			index = i
			break
		}
	}
	index = (index + dir + len(options)) % len(options)
	return options[index]
}

func nextPromptGlyphMode(current sharedtypes.PromptGlyphMode, dir int) sharedtypes.PromptGlyphMode {
	options := []sharedtypes.PromptGlyphMode{
		sharedtypes.PromptGlyphModeEmoji,
		sharedtypes.PromptGlyphModeUnicode,
		sharedtypes.PromptGlyphModeASCII,
	}
	return options[nextIndex(sharedtypes.NormalizePromptGlyphMode(current), options, dir)]
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

func nextAlertSoundPreset(current sharedtypes.AlertSoundPreset, dir int) sharedtypes.AlertSoundPreset {
	options := []sharedtypes.AlertSoundPreset{
		sharedtypes.AlertSoundPresetChime,
		sharedtypes.AlertSoundPresetSoftBell,
		sharedtypes.AlertSoundPresetFocusGong,
		sharedtypes.AlertSoundPresetMinimalClick,
	}
	return options[nextIndex(sharedtypes.NormalizeAlertSoundPreset(current), options, dir)]
}

func nextAlertUrgency(current sharedtypes.AlertUrgency, dir int) sharedtypes.AlertUrgency {
	options := []sharedtypes.AlertUrgency{
		sharedtypes.AlertUrgencyLow,
		sharedtypes.AlertUrgencyNormal,
		sharedtypes.AlertUrgencyHigh,
	}
	return options[nextIndex(sharedtypes.NormalizeAlertUrgency(current), options, dir)]
}
