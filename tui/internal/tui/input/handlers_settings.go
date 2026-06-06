package input

import (
	"strings"
	"time"

	sharedtypes "crona/shared/types"
	uistate "crona/tui/internal/tui/state"
	alertsmeta "crona/tui/internal/tui/views/alertsmeta"
	viewruntime "crona/tui/internal/tui/views/runtime"
	settingsmeta "crona/tui/internal/tui/views/settingsmeta"

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
	row, ok := selectedSettingsRow(s)
	if !ok {
		return s, nil, true
	}
	switch row.Label {
	case "Work Duration":
		return s, deps.PatchSetting(
			sharedtypes.CoreSettingsKeyWorkDurationMinutes,
			max(s.Settings.WorkDurationMinutes+dir*5, 5),
			repoID,
			streamID,
			s.DashboardDate,
		), true
	case "Update Checks":
		return s, deps.PatchSetting(
			sharedtypes.CoreSettingsKeyUpdateChecksEnabled,
			!s.Settings.UpdateChecksEnabled,
			repoID,
			streamID,
			s.DashboardDate,
		), true
	case "Update Prompt":
		return s, deps.PatchSetting(
			sharedtypes.CoreSettingsKeyUpdatePromptEnabled,
			!s.Settings.UpdatePromptEnabled,
			repoID,
			streamID,
			s.DashboardDate,
		), true
	case "Update Channel":
		return s, deps.PatchSetting(
			sharedtypes.CoreSettingsKeyUpdateChannel,
			nextUpdateChannel(s.Settings.UpdateChannel, dir),
			repoID,
			streamID,
			s.DashboardDate,
		), true
	case "Repo Sort":
		return s, deps.PatchSetting(
			sharedtypes.CoreSettingsKeyRepoSort,
			nextRepoSort(s.Settings.RepoSort, dir),
			repoID,
			streamID,
			s.DashboardDate,
		), true
	case "Stream Sort":
		return s, deps.PatchSetting(
			sharedtypes.CoreSettingsKeyStreamSort,
			nextStreamSort(s.Settings.StreamSort, dir),
			repoID,
			streamID,
			s.DashboardDate,
		), true
	case "Issue Sort":
		return s, deps.PatchSetting(
			sharedtypes.CoreSettingsKeyIssueSort,
			nextIssueSort(s.Settings.IssueSort, dir),
			repoID,
			streamID,
			s.DashboardDate,
		), true
	case "Habit Sort":
		return s, deps.PatchSetting(
			sharedtypes.CoreSettingsKeyHabitSort,
			nextHabitSort(s.Settings.HabitSort, dir),
			repoID,
			streamID,
			s.DashboardDate,
		), true
	case "Date Format":
		return s, deps.PatchSetting(
			sharedtypes.CoreSettingsKeyDateDisplayPreset,
			nextDateDisplayPreset(s.Settings.DateDisplayPreset, dir),
			repoID,
			streamID,
			s.DashboardDate,
		), true
	case "Prompt Glyphs":
		return s, deps.PatchSetting(
			sharedtypes.CoreSettingsKeyPromptGlyphMode,
			nextPromptGlyphMode(s.Settings.PromptGlyphMode, dir),
			repoID,
			streamID,
			s.DashboardDate,
		), true
	case "Week Start":
		return s, deps.PatchSetting(
			sharedtypes.CoreSettingsKeyWeekStart,
			nextWeekStart(s.Settings.WeekStart, dir),
			repoID,
			streamID,
			s.DashboardDate,
		), true
	case "Away Mode":
		return s, deps.PatchSetting(
			sharedtypes.CoreSettingsKeyAwayModeEnabled,
			!s.Settings.AwayModeEnabled,
			repoID,
			streamID,
			s.DashboardDate,
		), true
	case "Rollback Window":
		return s, deps.PatchSetting(
			sharedtypes.CoreSettingsKeyDailyPlanRollbackMins,
			max(currentRollbackMinutes(s.Settings.DailyPlanRollbackMins)+dir, 1),
			repoID,
			streamID,
			s.DashboardDate,
		), true
	default:
		return s, nil, true
	}
}

func handleActivateSelectedSetting(s State, deps Deps) (tea.Model, tea.Cmd, bool) {
	if s.ActiveView != uistate.ViewSettings || s.Settings == nil {
		return s, nil, true
	}
	row, ok := selectedSettingsRow(s)
	if !ok {
		return s, nil, true
	}
	switch row.Label {
	case "Custom Date Format":
		if deps.OpenEditDateDisplayFormatDialog != nil {
			deps.OpenEditDateDisplayFormatDialog(&s)
		}
		return s, nil, true
	case "Habit Streaks":
		s.ActiveView = uistate.ViewMomentum
		s.ActivePane = uistate.DefaultPane(uistate.ViewMomentum)
		return s, nil, true
	case "Rest & Streak Protection":
		deps.OpenEditRestProtectionDialog(&s)
		return s, nil, true
	case "Privacy & Diagnostics":
		if deps.OpenEditTelemetrySettingsDialog != nil {
			deps.OpenEditTelemetrySettingsDialog(&s)
		}
		return s, nil, true
	case "Wipe Runtime Data":
		deps.OpenConfirmWipeDataDialog(&s)
		return s, nil, true
	case "Uninstall Crona":
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
	row, ok := alertsmeta.SelectedRow(
		s.FilterState(uistate.PaneAlerts),
		s.Settings,
		s.AlertStatus,
		s.AlertReminders,
		s.Cursor[uistate.PaneAlerts],
	)
	if !ok {
		deps.ClampFiltered(&s, uistate.PaneAlerts)
		return s, nil, true
	}
	switch row.Key {
	case alertsmeta.RowNotifications:
		return s, deps.PatchSetting(
			sharedtypes.CoreSettingsKeyBoundaryNotifications,
			!s.Settings.BoundaryNotifications,
			repoID,
			streamID,
			s.DashboardDate,
		), true
	case alertsmeta.RowSound:
		return s, deps.PatchSetting(
			sharedtypes.CoreSettingsKeyBoundarySound,
			!s.Settings.BoundarySound,
			repoID,
			streamID,
			s.DashboardDate,
		), true
	case alertsmeta.RowSoundPreset:
		return s, deps.PatchSetting(
			sharedtypes.CoreSettingsKeyAlertSoundPreset,
			nextAlertSoundPreset(s.Settings.AlertSoundPreset, dir),
			repoID,
			streamID,
			s.DashboardDate,
		), true
	case alertsmeta.RowUrgency:
		return s, deps.PatchSetting(
			sharedtypes.CoreSettingsKeyAlertUrgency,
			nextAlertUrgency(s.Settings.AlertUrgency, dir),
			repoID,
			streamID,
			s.DashboardDate,
		), true
	case alertsmeta.RowLogoIcon:
		return s, deps.PatchSetting(
			sharedtypes.CoreSettingsKeyAlertIconEnabled,
			!s.Settings.AlertIconEnabled,
			repoID,
			streamID,
			s.DashboardDate,
		), true
	case alertsmeta.RowInactivityAlerts:
		return s, deps.PatchSetting(
			sharedtypes.CoreSettingsKeyInactivityAlerts,
			!s.Settings.InactivityAlerts,
			repoID,
			streamID,
			s.DashboardDate,
		), true
	case alertsmeta.RowInactivityAfter:
		return s, deps.PatchSetting(
			sharedtypes.CoreSettingsKeyInactivityThreshold,
			clampAlertMinutes(s.Settings.InactivityThreshold+dir*15),
			repoID,
			streamID,
			s.DashboardDate,
		), true
	case alertsmeta.RowInactivityRepeat:
		return s, deps.PatchSetting(
			sharedtypes.CoreSettingsKeyInactivityRepeat,
			clampAlertMinutes(s.Settings.InactivityRepeat+dir*15),
			repoID,
			streamID,
			s.DashboardDate,
		), true
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
	row, ok := alertsmeta.SelectedRow(
		s.FilterState(uistate.PaneAlerts),
		s.Settings,
		s.AlertStatus,
		s.AlertReminders,
		s.Cursor[uistate.PaneAlerts],
	)
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
	row, ok := alertsmeta.SelectedRow(
		s.FilterState(uistate.PaneAlerts),
		s.Settings,
		s.AlertStatus,
		s.AlertReminders,
		s.Cursor[uistate.PaneAlerts],
	)
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
	row, ok := alertsmeta.SelectedRow(
		s.FilterState(uistate.PaneAlerts),
		s.Settings,
		s.AlertStatus,
		s.AlertReminders,
		s.Cursor[uistate.PaneAlerts],
	)
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
	if (s.ActiveView != uistate.ViewDaily && s.ActiveView != uistate.ViewWellbeing && s.ActiveView != uistate.ViewAway) ||
		s.Settings == nil {
		return s, nil, false
	}
	protected, awayMode, _ := viewruntime.ProtectedRestMode(s.Settings, time.Now().Format("2006-01-02"))
	if protected && !awayMode {
		return s, nil, true
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

func selectedSettingsRow(s State) (settingsmeta.Row, bool) {
	if s.Settings == nil {
		return settingsmeta.Row{}, false
	}
	rows := settingsmeta.Rows(s.Settings)
	indices := settingsmeta.FilteredIndices(s.FilterState(uistate.PaneSettings), s.Settings)
	cursor := s.Cursor[uistate.PaneSettings]
	if cursor < 0 || cursor >= len(indices) {
		return settingsmeta.Row{}, false
	}
	rawIdx := indices[cursor]
	if rawIdx < 0 || rawIdx >= len(rows) {
		return settingsmeta.Row{}, false
	}
	return rows[rawIdx], true
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

func nextDateDisplayPreset(
	current sharedtypes.DateDisplayPreset,
	dir int,
) sharedtypes.DateDisplayPreset {
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

func nextWeekStart(current sharedtypes.WeekStart, dir int) sharedtypes.WeekStart {
	options := []sharedtypes.WeekStart{
		sharedtypes.WeekStartMonday,
		sharedtypes.WeekStartSunday,
	}
	return options[nextIndex(sharedtypes.NormalizeWeekStart(current), options, dir)]
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

func nextAlertSoundPreset(
	current sharedtypes.AlertSoundPreset,
	dir int,
) sharedtypes.AlertSoundPreset {
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
