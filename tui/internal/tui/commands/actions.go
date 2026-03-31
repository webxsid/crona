package commands

import (
	"bytes"
	"errors"
	"os/exec"
	"strings"
	"time"

	shareddto "crona/shared/dto"
	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	"crona/tui/internal/logger"
	helperpkg "crona/tui/internal/tui/helpers"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
)

func PatchSetting(c *api.Client, key sharedtypes.CoreSettingsKey, value any, repoID, streamID int64, dashboardDate string) tea.Cmd {
	return func() tea.Msg {
		if err := c.PatchSetting(key, value); err != nil {
			logger.Errorf("PatchSetting(%s): %v", key, err)
			return ErrMsg{Err: err}
		}
		cmds := []tea.Cmd{LoadSettings(c)}
		switch key {
		case sharedtypes.CoreSettingsKeyUpdateChecksEnabled, sharedtypes.CoreSettingsKeyUpdatePromptEnabled:
			cmds = append(cmds, LoadUpdateStatus(c))
		case sharedtypes.CoreSettingsKeyRepoSort:
			cmds = append(cmds, LoadRepos(c))
		case sharedtypes.CoreSettingsKeyStreamSort:
			if repoID != 0 {
				cmds = append(cmds, LoadStreams(c, repoID))
			}
		case sharedtypes.CoreSettingsKeyIssueSort:
			cmds = append(cmds, LoadAllIssues(c), LoadDailySummary(c, dashboardDate))
			if streamID != 0 {
				cmds = append(cmds, LoadIssues(c, streamID))
			}
		case sharedtypes.CoreSettingsKeyAwayModeEnabled,
			sharedtypes.CoreSettingsKeyFrozenStreakKinds,
			sharedtypes.CoreSettingsKeyRestWeekdays,
			sharedtypes.CoreSettingsKeyRestSpecificDates:
			cmds = append(cmds, LoadWellbeing(c, dashboardDate))
		}
		return tea.Batch(cmds...)()
	}
}

func UpsertDailyCheckIn(c *api.Client, input shareddto.DailyCheckInUpsertRequest, date string) tea.Cmd {
	return func() tea.Msg {
		if _, err := c.UpsertDailyCheckIn(input); err != nil {
			logger.Errorf("UpsertDailyCheckIn: %v", err)
			return ErrMsg{Err: err}
		}
		return tea.Batch(LoadWellbeing(c, date))()
	}
}

func DeleteDailyCheckIn(c *api.Client, date string) tea.Cmd {
	return func() tea.Msg {
		if err := c.DeleteDailyCheckIn(date); err != nil {
			logger.Errorf("DeleteDailyCheckIn: %v", err)
			return ErrMsg{Err: err}
		}
		return tea.Batch(LoadWellbeing(c, date))()
	}
}

func ShutdownKernel(c *api.Client) tea.Cmd {
	return func() tea.Msg {
		if err := c.ShutdownKernel(); err != nil {
			logger.Errorf("ShutdownKernel: %v", err)
			return ErrMsg{Err: err}
		}
		return KernelShutdownMsg{}
	}
}

func DismissUpdate(c *api.Client) tea.Cmd {
	return func() tea.Msg {
		status, err := c.DismissUpdate()
		if err != nil {
			logger.Errorf("DismissUpdate: %v", err)
			return ErrMsg{Err: err}
		}
		return UpdateDismissedMsg{Status: status}
	}
}

func SeedDevData(c *api.Client) tea.Cmd {
	return func() tea.Msg {
		if err := c.SeedDevData(); err != nil {
			logger.Errorf("SeedDevData: %v", err)
			return ErrMsg{Err: err}
		}
		return DevSeededMsg{}
	}
}

func ClearDevData(c *api.Client) tea.Cmd {
	return func() tea.Msg {
		if err := c.ClearDevData(); err != nil {
			logger.Errorf("ClearDevData: %v", err)
			return ErrMsg{Err: err}
		}
		return DevClearedMsg{}
	}
}

func CreateScratchpad(c *api.Client, name, path string) tea.Cmd {
	return func() tea.Msg {
		id := uuid.New().String()
		if err := c.RegisterScratchpad(id, name, path); err != nil {
			logger.Errorf("RegisterScratchpad: %v", err)
			return ErrMsg{Err: err}
		}
		return LoadScratchpads(c)()
	}
}

func CreateRepoOnly(c *api.Client, name string, description *string) tea.Cmd {
	return func() tea.Msg {
		if _, err := c.CreateRepo(name, description); err != nil {
			logger.Errorf("CreateRepo: %v", err)
			return ErrMsg{Err: err}
		}
		return LoadRepos(c)()
	}
}

func UpdateRepo(c *api.Client, repoID int64, name string, description *string) tea.Cmd {
	return func() tea.Msg {
		if err := c.UpdateRepo(repoID, name, description); err != nil {
			logger.Errorf("UpdateRepo: %v", err)
			return ErrMsg{Err: err}
		}
		return tea.Batch(LoadRepos(c), LoadContext(c), LoadAllIssues(c))()
	}
}

func DeleteRepo(c *api.Client, repoID int64) tea.Cmd {
	return func() tea.Msg {
		if err := c.DeleteRepo(repoID); err != nil {
			logger.Errorf("DeleteRepo: %v", err)
			return ErrMsg{Err: err}
		}
		return tea.Batch(LoadRepos(c), LoadAllIssues(c), LoadDailySummary(c, ""), LoadContext(c))()
	}
}

func CreateStreamOnly(c *api.Client, repoID int64, name string, description *string) tea.Cmd {
	return func() tea.Msg {
		if _, err := c.CreateStream(repoID, name, description); err != nil {
			logger.Errorf("CreateStream: %v", err)
			return ErrMsg{Err: err}
		}
		return tea.Batch(LoadStreams(c, repoID), LoadAllIssues(c), LoadDailySummary(c, ""))()
	}
}

func UpdateStream(c *api.Client, repoID, streamID int64, name string, description *string) tea.Cmd {
	return func() tea.Msg {
		if err := c.UpdateStream(streamID, name, description); err != nil {
			logger.Errorf("UpdateStream: %v", err)
			return ErrMsg{Err: err}
		}
		return tea.Batch(LoadStreams(c, repoID), LoadAllIssues(c), LoadDailySummary(c, ""), LoadContext(c))()
	}
}

func DeleteStream(c *api.Client, repoID, streamID int64) tea.Cmd {
	return func() tea.Msg {
		if err := c.DeleteStream(streamID); err != nil {
			logger.Errorf("DeleteStream: %v", err)
			return ErrMsg{Err: err}
		}
		cmds := []tea.Cmd{LoadAllIssues(c), LoadDailySummary(c, ""), LoadContext(c)}
		if repoID != 0 {
			cmds = append(cmds, LoadStreams(c, repoID))
		}
		return tea.Batch(cmds...)()
	}
}

func CreateIssueOnly(c *api.Client, streamID int64, title string, description *string, estimateMinutes *int, todoForDate *string) tea.Cmd {
	return func() tea.Msg {
		if _, err := c.CreateIssue(streamID, title, description, estimateMinutes, todoForDate); err != nil {
			logger.Errorf("CreateIssue: %v", err)
			return ErrMsg{Err: err}
		}
		return tea.Batch(LoadIssues(c, streamID), LoadAllIssues(c), LoadDailySummary(c, ""))()
	}
}

func CreateHabitOnly(c *api.Client, streamID int64, name string, description *string, scheduleType string, weekdays []int, targetMinutes *int) tea.Cmd {
	return func() tea.Msg {
		if _, err := c.CreateHabit(streamID, name, description, scheduleType, weekdays, targetMinutes); err != nil {
			logger.Errorf("CreateHabit: %v", err)
			return ErrMsg{Err: err}
		}
		return tea.Batch(LoadHabits(c, streamID), LoadDueHabits(c, time.Now().Format("2006-01-02")))()
	}
}

func UpdateHabit(c *api.Client, habitID, streamID int64, name string, description *string, scheduleType string, weekdays []int, targetMinutes *int, active bool, dashboardDate string) tea.Cmd {
	return func() tea.Msg {
		if err := c.UpdateHabit(habitID, name, description, scheduleType, weekdays, targetMinutes, active); err != nil {
			logger.Errorf("UpdateHabit: %v", err)
			return ErrMsg{Err: err}
		}
		return tea.Batch(LoadHabits(c, streamID), LoadDueHabits(c, dashboardDate))()
	}
}

func DeleteHabit(c *api.Client, habitID, streamID int64, dashboardDate string) tea.Cmd {
	return func() tea.Msg {
		if err := c.DeleteHabit(habitID); err != nil {
			logger.Errorf("DeleteHabit: %v", err)
			return ErrMsg{Err: err}
		}
		cmds := []tea.Cmd{LoadDueHabits(c, dashboardDate)}
		if streamID != 0 {
			cmds = append(cmds, LoadHabits(c, streamID))
		}
		return tea.Batch(cmds...)()
	}
}

func SetHabitStatus(c *api.Client, habitID int64, date string, status sharedtypes.HabitCompletionStatus, durationMinutes *int, notes *string) tea.Cmd {
	return func() tea.Msg {
		if _, err := c.CompleteHabit(habitID, date, status, durationMinutes, notes); err != nil {
			logger.Errorf("CompleteHabit: %v", err)
			return ErrMsg{Err: err}
		}
		return LoadDueHabits(c, date)()
	}
}

func UncompleteHabit(c *api.Client, habitID int64, date string) tea.Cmd {
	return func() tea.Msg {
		if err := c.UncompleteHabit(habitID, date); err != nil {
			logger.Errorf("UncompleteHabit: %v", err)
			return ErrMsg{Err: err}
		}
		return LoadDueHabits(c, date)()
	}
}

func UpdateIssue(c *api.Client, issueID, streamID int64, title string, description *string, estimateMinutes *int, todoForDate *string, dashboardDate string) tea.Cmd {
	return func() tea.Msg {
		if err := c.UpdateIssue(issueID, title, description, estimateMinutes); err != nil {
			logger.Errorf("UpdateIssue: %v", err)
			return ErrMsg{Err: err}
		}
		currentDue := ""
		if todoForDate != nil {
			currentDue = *todoForDate
		}
		if strings.TrimSpace(currentDue) == "" {
			if err := c.ClearIssueTodo(issueID); err != nil {
				logger.Errorf("ClearIssueTodo after UpdateIssue: %v", err)
				return ErrMsg{Err: err}
			}
		} else if err := c.SetIssueTodoDate(issueID, currentDue); err != nil {
			logger.Errorf("SetIssueTodoDate after UpdateIssue: %v", err)
			return ErrMsg{Err: err}
		}
		cmds := []tea.Cmd{LoadAllIssues(c), LoadDailySummary(c, dashboardDate), LoadContext(c)}
		if streamID != 0 {
			cmds = append(cmds, LoadIssues(c, streamID))
		}
		return tea.Batch(cmds...)()
	}
}

func DeleteIssue(c *api.Client, issueID, streamID int64, dashboardDate string) tea.Cmd {
	return func() tea.Msg {
		if err := c.DeleteIssue(issueID); err != nil {
			logger.Errorf("DeleteIssue: %v", err)
			return ErrMsg{Err: err}
		}
		cmds := []tea.Cmd{LoadAllIssues(c), LoadDailySummary(c, dashboardDate), LoadContext(c), LoadTimer(c)}
		if streamID != 0 {
			cmds = append(cmds, LoadIssues(c, streamID))
		}
		return tea.Batch(cmds...)()
	}
}

func CreateIssueWithPath(c *api.Client, repoName, repoDescription, streamName, streamDescription, title string, issueDescription *string, estimateMinutes *int, todoForDate *string) tea.Cmd {
	return func() tea.Msg {
		repos, err := c.ListRepos()
		if err != nil {
			logger.Errorf("ListRepos before CreateIssueWithPath: %v", err)
			return ErrMsg{Err: err}
		}
		var repoID int64
		for _, repo := range repos {
			if helperpkg.SameLookupName(repo.Name, repoName) {
				repoID = repo.ID
				break
			}
		}
		if repoID == 0 {
			repo, err := c.CreateRepo(repoName, helperpkg.NormalizeOptionalValue(repoDescription))
			if err != nil {
				logger.Errorf("CreateRepo before CreateIssueWithPath: %v", err)
				return ErrMsg{Err: err}
			}
			repoID = repo.ID
		}
		streams, err := c.ListStreams(repoID)
		if err != nil {
			logger.Errorf("ListStreams before CreateIssueWithPath: %v", err)
			return ErrMsg{Err: err}
		}
		var streamID int64
		for _, stream := range streams {
			if helperpkg.SameLookupName(stream.Name, streamName) {
				streamID = stream.ID
				break
			}
		}
		if streamID == 0 {
			stream, err := c.CreateStream(repoID, streamName, helperpkg.NormalizeOptionalValue(streamDescription))
			if err != nil {
				logger.Errorf("CreateStream before CreateIssueWithPath: %v", err)
				return ErrMsg{Err: err}
			}
			streamID = stream.ID
		}
		if _, err := c.CreateIssue(streamID, title, issueDescription, estimateMinutes, todoForDate); err != nil {
			logger.Errorf("CreateIssue in CreateIssueWithPath: %v", err)
			return ErrMsg{Err: err}
		}
		return tea.Batch(LoadRepos(c), LoadStreams(c, repoID), LoadIssues(c, streamID), LoadAllIssues(c), LoadDailySummary(c, ""))()
	}
}

func CreateHabitWithPath(c *api.Client, repoName, repoDescription, streamName, streamDescription, name string, habitDescription *string, scheduleType string, weekdays []int, targetMinutes *int) tea.Cmd {
	return func() tea.Msg {
		repos, err := c.ListRepos()
		if err != nil {
			logger.Errorf("ListRepos before CreateHabitWithPath: %v", err)
			return ErrMsg{Err: err}
		}
		var repoID int64
		for _, repo := range repos {
			if helperpkg.SameLookupName(repo.Name, repoName) {
				repoID = repo.ID
				break
			}
		}
		if repoID == 0 {
			repo, err := c.CreateRepo(repoName, helperpkg.NormalizeOptionalValue(repoDescription))
			if err != nil {
				logger.Errorf("CreateRepo before CreateHabitWithPath: %v", err)
				return ErrMsg{Err: err}
			}
			repoID = repo.ID
		}
		streams, err := c.ListStreams(repoID)
		if err != nil {
			logger.Errorf("ListStreams before CreateHabitWithPath: %v", err)
			return ErrMsg{Err: err}
		}
		var streamID int64
		for _, stream := range streams {
			if helperpkg.SameLookupName(stream.Name, streamName) {
				streamID = stream.ID
				break
			}
		}
		if streamID == 0 {
			stream, err := c.CreateStream(repoID, streamName, helperpkg.NormalizeOptionalValue(streamDescription))
			if err != nil {
				logger.Errorf("CreateStream before CreateHabitWithPath: %v", err)
				return ErrMsg{Err: err}
			}
			streamID = stream.ID
		}
		if _, err := c.CreateHabit(streamID, name, habitDescription, scheduleType, weekdays, targetMinutes); err != nil {
			logger.Errorf("CreateHabit in CreateHabitWithPath: %v", err)
			return ErrMsg{Err: err}
		}
		return tea.Batch(LoadRepos(c), LoadStreams(c, repoID), LoadHabits(c, streamID), LoadDueHabits(c, time.Now().Format("2006-01-02")))()
	}
}

func DeleteScratchpad(c *api.Client, id string) tea.Cmd {
	return func() tea.Msg {
		if err := c.DeleteScratchpad(id); err != nil {
			logger.Errorf("DeleteScratchpad(%s): %v", id, err)
			return ErrMsg{Err: err}
		}
		return LoadScratchpads(c)()
	}
}

func OpenScratchpad(c *api.Client, scratchpads []api.ScratchPad, idx int) tea.Cmd {
	return func() tea.Msg {
		if idx >= len(scratchpads) {
			return nil
		}
		pad := scratchpads[idx]
		filePath, content, err := c.ReadScratchpad(pad.ID)
		if err != nil {
			logger.Errorf("ReadScratchpad(%s): %v", pad.ID, err)
			return ErrMsg{Err: err}
		}
		return OpenScratchpadMsg{Meta: pad, FilePath: filePath, Content: content}
	}
}

func CheckoutRepo(c *api.Client, repoID int64) tea.Cmd {
	return func() tea.Msg {
		if err := c.SwitchRepo(repoID); err != nil {
			logger.Errorf("SwitchRepo: %v", err)
			return ErrMsg{Err: err}
		}
		return LoadContext(c)()
	}
}

func CheckoutStream(c *api.Client, streamID int64) tea.Cmd {
	return func() tea.Msg {
		if err := c.SwitchStream(streamID); err != nil {
			logger.Errorf("SwitchStream: %v", err)
			return ErrMsg{Err: err}
		}
		return LoadContext(c)()
	}
}

func CheckoutContext(c *api.Client, repoID int64, repoName string, streamID int64, streamName string) tea.Cmd {
	return func() tea.Msg {
		if repoID == 0 && strings.TrimSpace(repoName) == "" && streamID == 0 && strings.TrimSpace(streamName) == "" {
			if err := c.ClearContext(); err != nil {
				logger.Errorf("ClearContext during checkout: %v", err)
				return ErrMsg{Err: err}
			}
			return tea.Batch(LoadContext(c), LoadAllIssues(c))()
		}
		resolvedRepoID := repoID
		if resolvedRepoID == 0 {
			if strings.TrimSpace(repoName) == "" {
				return ErrMsg{Err: errors.New("repo is required")}
			}
			repo, err := c.CreateRepo(repoName, nil)
			if err != nil {
				logger.Errorf("CreateRepo during checkout: %v", err)
				return ErrMsg{Err: err}
			}
			resolvedRepoID = repo.ID
		}
		if streamID != 0 {
			if err := c.SetFullContext(resolvedRepoID, streamID, 0); err != nil {
				logger.Errorf("SetFullContext during checkout: %v", err)
				return ErrMsg{Err: err}
			}
			return tea.Batch(LoadContext(c), LoadRepos(c), LoadStreams(c, resolvedRepoID), LoadIssues(c, streamID), LoadAllIssues(c))()
		}
		if strings.TrimSpace(streamName) != "" {
			stream, err := c.CreateStream(resolvedRepoID, streamName, nil)
			if err != nil {
				logger.Errorf("CreateStream during checkout: %v", err)
				return ErrMsg{Err: err}
			}
			if err := c.SetFullContext(resolvedRepoID, stream.ID, 0); err != nil {
				logger.Errorf("SetFullContext after create during checkout: %v", err)
				return ErrMsg{Err: err}
			}
			return tea.Batch(LoadContext(c), LoadRepos(c), LoadStreams(c, resolvedRepoID), LoadIssues(c, stream.ID), LoadAllIssues(c))()
		}
		if err := c.SwitchRepo(resolvedRepoID); err != nil {
			logger.Errorf("SwitchRepo during checkout: %v", err)
			return ErrMsg{Err: err}
		}
		return tea.Batch(LoadContext(c), LoadRepos(c), LoadStreams(c, resolvedRepoID), LoadAllIssues(c))()
	}
}

func ChangeIssueStatus(c *api.Client, issueID int64, status string, note *string, streamID int64, dashboardDate string) tea.Cmd {
	return func() tea.Msg {
		if err := c.ChangeIssueStatus(issueID, status, note); err != nil {
			logger.Errorf("ChangeIssueStatus: %v", err)
			return ErrMsg{Err: err}
		}
		cmds := []tea.Cmd{LoadAllIssues(c), LoadDailySummary(c, dashboardDate)}
		if streamID != 0 {
			cmds = append(cmds, LoadIssues(c, streamID))
		}
		return tea.Batch(cmds...)()
	}
}

func ToggleIssueToday(c *api.Client, issueID int64, markedForToday bool, streamID int64, dashboardDate string) tea.Cmd {
	return func() tea.Msg {
		var err error
		if markedForToday {
			err = c.ClearIssueTodo(issueID)
		} else {
			err = c.MarkIssueTodoForToday(issueID)
		}
		if err != nil {
			logger.Errorf("ToggleIssueToday: %v", err)
			return ErrMsg{Err: err}
		}
		cmds := []tea.Cmd{LoadAllIssues(c), LoadDailySummary(c, dashboardDate)}
		if streamID != 0 {
			cmds = append(cmds, LoadIssues(c, streamID))
		}
		return tea.Batch(cmds...)()
	}
}

func SetIssueTodoDate(c *api.Client, issueID int64, date string, streamID int64, dashboardDate string) tea.Cmd {
	return func() tea.Msg {
		var err error
		if strings.TrimSpace(date) == "" {
			err = c.ClearIssueTodo(issueID)
		} else {
			err = c.SetIssueTodoDate(issueID, date)
		}
		if err != nil {
			logger.Errorf("SetIssueTodoDate: %v", err)
			return ErrMsg{Err: err}
		}
		cmds := []tea.Cmd{LoadAllIssues(c), LoadDailySummary(c, dashboardDate)}
		if streamID != 0 {
			cmds = append(cmds, LoadIssues(c, streamID))
		}
		return tea.Batch(cmds...)()
	}
}

func ChangeIssueStatusAndEndSession(c *api.Client, issueID int64, status string, note *string, streamID int64, dashboardDate string, endInput shareddto.EndSessionRequest) tea.Cmd {
	return func() tea.Msg {
		if err := c.EndTimer(endInput); err != nil {
			logger.Errorf("EndTimer before ChangeIssueStatus: %v", err)
			return ErrMsg{Err: err}
		}
		if err := c.ChangeIssueStatus(issueID, status, note); err != nil {
			logger.Errorf("ChangeIssueStatus after EndTimer: %v", err)
			return ErrMsg{Err: err}
		}
		cmds := []tea.Cmd{LoadAllIssues(c), LoadDailySummary(c, dashboardDate), LoadContext(c), LoadTimer(c)}
		if streamID != 0 {
			cmds = append(cmds, LoadIssues(c, streamID))
		}
		return tea.Batch(cmds...)()
	}
}

func AmendSessionNote(c *api.Client, id string, note string) tea.Cmd {
	return func() tea.Msg {
		if err := c.AmendSessionNote(id, note); err != nil {
			logger.Errorf("AmendSessionNote(%s): %v", id, err)
			return ErrMsg{Err: err}
		}
		return SessionAmendedMsg{ID: id}
	}
}

func LogManualSession(c *api.Client, input shareddto.ManualSessionLogRequest) tea.Cmd {
	return func() tea.Msg {
		session, err := c.LogManualSession(input)
		if err != nil {
			logger.Errorf("LogManualSession(%d): %v", input.IssueID, err)
			return ErrMsg{Err: err}
		}
		return ManualSessionLoggedMsg{ID: session.ID, IssueID: session.IssueID, Date: input.Date}
	}
}

func StartFocusSession(c *api.Client, repoID, streamID, issueID int64) tea.Cmd {
	return func() tea.Msg {
		if repoID != 0 && streamID != 0 && issueID != 0 {
			if err := c.SetFullContext(repoID, streamID, issueID); err != nil {
				logger.Errorf("SetFullContext before StartTimer: %v", err)
				return ErrMsg{Err: err}
			}
		} else if issueID != 0 {
			if err := c.SwitchIssue(issueID); err != nil {
				logger.Errorf("SwitchIssue before StartTimer: %v", err)
				return ErrMsg{Err: err}
			}
		}
		if err := c.StartTimer(0); err != nil {
			logger.Errorf("StartTimer: %v", err)
			return ErrMsg{Err: err}
		}
		return FocusSessionChangedMsg{ReloadContext: true, ReloadTimer: true}
	}
}

func PauseFocusSession(c *api.Client) tea.Cmd {
	return func() tea.Msg {
		if err := c.PauseTimer(); err != nil {
			logger.Errorf("PauseTimer: %v", err)
			return ErrMsg{Err: err}
		}
		return FocusSessionChangedMsg{ReloadTimer: true}
	}
}

func ResumeFocusSession(c *api.Client) tea.Cmd {
	return func() tea.Msg {
		if err := c.ResumeTimer(); err != nil {
			logger.Errorf("ResumeTimer: %v", err)
			return ErrMsg{Err: err}
		}
		return FocusSessionChangedMsg{ReloadTimer: true}
	}
}

func EndFocusSession(c *api.Client, streamID int64, dashboardDate string, endInput shareddto.EndSessionRequest) tea.Cmd {
	return func() tea.Msg {
		if err := c.EndTimer(endInput); err != nil {
			logger.Errorf("EndTimer: %v", err)
			return ErrMsg{Err: err}
		}
		cmds := []tea.Cmd{LoadAllIssues(c), LoadDailySummary(c, dashboardDate), LoadContext(c), LoadTimer(c), LoadSessionHistory(c, nil, 200)}
		if streamID != 0 {
			cmds = append(cmds, LoadIssues(c, streamID))
		}
		return tea.Batch(cmds...)()
	}
}

func StashFocusSession(c *api.Client, note string) tea.Cmd {
	return func() tea.Msg {
		if err := c.StashPush(note); err != nil {
			logger.Errorf("StashPush: %v", err)
			return ErrMsg{Err: err}
		}
		return FocusSessionChangedMsg{ReloadContext: true, ReloadTimer: true}
	}
}

func ApplyStash(c *api.Client, id string) tea.Cmd {
	return func() tea.Msg {
		if err := c.ApplyStash(id); err != nil {
			logger.Errorf("ApplyStash: %v", err)
			return ErrMsg{Err: err}
		}
		return tea.Batch(LoadStashes(c), LoadContext(c), LoadTimer(c), LoadSessionHistory(c, nil, 200))()
	}
}

func DropStash(c *api.Client, id string) tea.Cmd {
	return func() tea.Msg {
		if err := c.DropStash(id); err != nil {
			logger.Errorf("DropStash: %v", err)
			return ErrMsg{Err: err}
		}
		return LoadStashes(c)()
	}
}

func GenerateReport(c *api.Client, input shareddto.ExportReportRequest) tea.Cmd {
	return func() tea.Msg {
		result, err := c.GenerateReport(input)
		if err != nil {
			logger.Errorf("GenerateReport: %v", err)
			return ErrMsg{Err: err}
		}
		return DailyReportGeneratedMsg{Result: result}
	}
}

func GenerateCalendarExport(c *api.Client, input shareddto.ExportCalendarRequest) tea.Cmd {
	return func() tea.Msg {
		result, err := c.GenerateCalendarExport(input)
		if err != nil {
			logger.Errorf("GenerateCalendarExport: %v", err)
			return ErrMsg{Err: err}
		}
		return CalendarExportGeneratedMsg{Result: result}
	}
}

func GenerateDailyReport(c *api.Client, date string, format sharedtypes.ExportFormat, mode sharedtypes.ExportOutputMode) tea.Cmd {
	return GenerateReport(c, shareddto.ExportReportRequest{
		Kind:       sharedtypes.ExportReportKindDaily,
		Date:       date,
		Format:     format,
		OutputMode: mode,
	})
}

func CopyDailyReport(c *api.Client, date string) tea.Cmd {
	return func() tea.Msg {
		result, err := c.GenerateDailyReport(date, sharedtypes.ExportFormatMarkdown, sharedtypes.ExportOutputModeClipboard)
		if err != nil {
			logger.Errorf("GenerateDailyReport clipboard: %v", err)
			return ErrMsg{Err: err}
		}
		if err := copyToClipboard(result.Content); err != nil {
			return ErrMsg{Err: err}
		}
		return ClipboardCopiedMsg{Message: "Daily report copied to clipboard"}
	}
}

func ResetExportTemplate(c *api.Client, reportKind sharedtypes.ExportReportKind, assetKind sharedtypes.ExportAssetKind) tea.Cmd {
	return func() tea.Msg {
		assets, err := c.ResetExportTemplate(reportKind, assetKind)
		if err != nil {
			logger.Errorf("ResetExportTemplate: %v", err)
			return ErrMsg{Err: err}
		}
		return ExportAssetsLoadedMsg{Assets: assets}
	}
}

func ApplyExportTemplatePreset(c *api.Client, reportKind sharedtypes.ExportReportKind, assetKind sharedtypes.ExportAssetKind, presetID string) tea.Cmd {
	return func() tea.Msg {
		assets, err := c.ApplyExportTemplatePreset(reportKind, assetKind, presetID)
		if err != nil {
			logger.Errorf("ApplyExportTemplatePreset: %v", err)
			return ErrMsg{Err: err}
		}
		return ExportAssetsLoadedMsg{Assets: assets}
	}
}

func SetExportReportsDir(c *api.Client, path string) tea.Cmd {
	return func() tea.Msg {
		assets, err := c.SetExportReportsDir(path)
		if err != nil {
			logger.Errorf("SetExportReportsDir: %v", err)
			return ErrMsg{Err: err}
		}
		return ExportAssetsLoadedMsg{Assets: assets}
	}
}

func SetExportICSDir(c *api.Client, path string) tea.Cmd {
	return func() tea.Msg {
		assets, err := c.SetExportICSDir(path)
		if err != nil {
			logger.Errorf("SetExportICSDir: %v", err)
			return ErrMsg{Err: err}
		}
		return ExportAssetsLoadedMsg{Assets: assets}
	}
}

func DeleteExportReport(c *api.Client, report api.ExportReportFile) tea.Cmd {
	return func() tea.Msg {
		if err := c.DeleteExportReport(report.Path); err != nil {
			logger.Errorf("DeleteExportReport: %v", err)
			return ErrMsg{Err: err}
		}
		return ExportReportDeletedMsg{Name: report.Name}
	}
}

func copyToClipboard(text string) error {
	commands := [][]string{
		{"pbcopy"},
		{"wl-copy"},
		{"xclip", "-selection", "clipboard"},
		{"xsel", "--clipboard", "--input"},
		{"clip"},
	}
	for _, args := range commands {
		path, err := exec.LookPath(args[0])
		if err != nil {
			continue
		}
		cmd := exec.Command(path, args[1:]...)
		cmd.Stdin = bytes.NewBufferString(text)
		if err := cmd.Run(); err == nil {
			return nil
		}
	}
	return errors.New("no supported clipboard command found")
}
