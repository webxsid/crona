package api

import (
	"bufio"
	shareddto "crona/shared/dto"
	"crona/shared/localipc"
	"crona/shared/protocol"
	sharedtypes "crona/shared/types"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"
)

const defaultShutdownWait = 5 * time.Second

type Client struct {
	transport  string
	endpoint   string
	scratchDir string
	nextID     atomic.Uint64
}

func NewClient(transport, endpoint, scratchDir string) *Client {
	return &Client{
		transport:  transport,
		endpoint:   endpoint,
		scratchDir: scratchDir,
	}
}

func (c *Client) call(method string, params, out any) error {
	conn, err := localipc.Dial(c.endpoint, 10*time.Second)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()
	_ = conn.SetDeadline(time.Now().Add(10 * time.Second))

	var rawParams json.RawMessage
	if params != nil {
		b, err := json.Marshal(params)
		if err != nil {
			return err
		}
		rawParams = b
	}

	id := fmt.Sprintf("req-%d", c.nextID.Add(1))
	reqBody, err := json.Marshal(protocol.Request{
		ID:     id,
		Method: method,
		Params: rawParams,
	})
	if err != nil {
		return err
	}
	if _, err := conn.Write(append(reqBody, '\n')); err != nil {
		return err
	}

	var resp protocol.Response
	if err := json.NewDecoder(bufio.NewReader(conn)).Decode(&resp); err != nil {
		return err
	}
	if resp.Error != nil {
		return &protocol.RPCError{Code: resp.Error.Code, Message: resp.Error.Message, Data: resp.Error.Data}
	}
	if out == nil || len(resp.Result) == 0 {
		return nil
	}
	return json.Unmarshal(resp.Result, out)
}

func (c *Client) mustOK(method string, params any) error {
	var out shareddto.OKResponse
	if err := c.call(method, params, &out); err != nil {
		return err
	}
	if !out.OK {
		return fmt.Errorf("%s failed", method)
	}
	return nil
}

func (c *Client) ListRepos() ([]Repo, error) {
	var out []Repo
	return out, c.call(protocol.MethodRepoList, nil, &out)
}

func (c *Client) CreateRepo(name string, description *string) (*Repo, error) {
	var out Repo
	return &out, c.call(protocol.MethodRepoCreate, shareddto.CreateRepoRequest{Name: name, Description: description}, &out)
}

func (c *Client) UpdateRepo(id int64, name string, description *string) error {
	return c.call(protocol.MethodRepoUpdate, map[string]any{
		"id":          id,
		"name":        name,
		"description": description,
	}, nil)
}

func (c *Client) DeleteRepo(id int64) error {
	return c.mustOK(protocol.MethodRepoDelete, shareddto.NumericIDRequest{ID: id})
}

func (c *Client) ListStreams(repoID int64) ([]Stream, error) {
	var out []Stream
	return out, c.call(protocol.MethodStreamList, shareddto.ListStreamsQuery{RepoID: repoID}, &out)
}

func (c *Client) CreateStream(repoID int64, name string, description *string) (*Stream, error) {
	var out Stream
	return &out, c.call(protocol.MethodStreamCreate, shareddto.CreateStreamRequest{RepoID: repoID, Name: name, Description: description}, &out)
}

func (c *Client) UpdateStream(id int64, name string, description *string) error {
	return c.call(protocol.MethodStreamUpdate, map[string]any{
		"id":          id,
		"name":        name,
		"description": description,
	}, nil)
}

func (c *Client) DeleteStream(id int64) error {
	return c.mustOK(protocol.MethodStreamDelete, shareddto.NumericIDRequest{ID: id})
}

func (c *Client) ListIssues(streamID int64) ([]Issue, error) {
	var out []Issue
	return out, c.call(protocol.MethodIssueList, shareddto.ListIssuesQuery{StreamID: streamID}, &out)
}

func (c *Client) ListHabits(streamID int64) ([]Habit, error) {
	var out []Habit
	return out, c.call(protocol.MethodHabitList, shareddto.ListHabitsQuery{StreamID: streamID}, &out)
}

func (c *Client) ListDueHabits(date string) ([]HabitDailyItem, error) {
	var out []HabitDailyItem
	return out, c.call(protocol.MethodHabitListDue, shareddto.ListHabitsDueQuery{Date: strings.TrimSpace(date)}, &out)
}

func (c *Client) CreateHabit(streamID int64, name string, description *string, scheduleType string, weekdays []int, targetMinutes *int) (*Habit, error) {
	var out Habit
	return &out, c.call(protocol.MethodHabitCreate, shareddto.CreateHabitRequest{
		StreamID:      streamID,
		Name:          name,
		Description:   description,
		ScheduleType:  scheduleType,
		Weekdays:      weekdays,
		TargetMinutes: targetMinutes,
	}, &out)
}

func (c *Client) UpdateHabit(id int64, name string, description *string, scheduleType string, weekdays []int, targetMinutes *int, active bool) error {
	return c.call(protocol.MethodHabitUpdate, shareddto.UpdateHabitRequest{
		ID:            id,
		Name:          &name,
		Description:   description,
		ScheduleType:  &scheduleType,
		Weekdays:      weekdays,
		TargetMinutes: targetMinutes,
		Active:        &active,
	}, nil)
}

func (c *Client) DeleteHabit(id int64) error {
	return c.mustOK(protocol.MethodHabitDelete, shareddto.NumericIDRequest{ID: id})
}

func (c *Client) CompleteHabit(habitID int64, date string, status sharedtypes.HabitCompletionStatus, durationMinutes *int, notes *string) (*HabitCompletion, error) {
	var out HabitCompletion
	return &out, c.call(protocol.MethodHabitComplete, shareddto.HabitCompletionUpsertRequest{
		HabitID:         habitID,
		Date:            strings.TrimSpace(date),
		Status:          &status,
		DurationMinutes: durationMinutes,
		Notes:           notes,
	}, &out)
}

func (c *Client) UncompleteHabit(habitID int64, date string) error {
	return c.mustOK(protocol.MethodHabitUncomplete, shareddto.HabitCompletionUpsertRequest{
		HabitID: habitID,
		Date:    strings.TrimSpace(date),
	})
}

func (c *Client) ListHabitHistory(repoID, streamID *int64) ([]HabitCompletion, error) {
	var out []HabitCompletion
	return out, c.call(protocol.MethodHabitHistory, shareddto.HabitHistoryQuery{RepoID: repoID, StreamID: streamID}, &out)
}

func (c *Client) ListAllIssues() ([]IssueWithMeta, error) {
	var out []IssueWithMeta
	return out, c.call(protocol.MethodIssueListAll, nil, &out)
}

func (c *Client) CreateIssue(streamID int64, title string, description *string, estimateMinutes *int, todoForDate *string) (*Issue, error) {
	body := shareddto.CreateIssueRequest{
		StreamID:    streamID,
		Title:       title,
		Description: description,
	}
	if estimateMinutes != nil {
		body.EstimateMinutes = estimateMinutes
	}
	if todoForDate != nil && strings.TrimSpace(*todoForDate) != "" {
		trimmed := strings.TrimSpace(*todoForDate)
		body.TodoForDate = &trimmed
	}
	var out Issue
	return &out, c.call(protocol.MethodIssueCreate, body, &out)
}

func (c *Client) UpdateIssue(id int64, title string, description *string, estimateMinutes *int) error {
	body := shareddto.UpdateIssueRequest{
		ID:          id,
		Title:       &title,
		Description: description,
	}
	body.EstimateMinutes = estimateMinutes
	return c.call(protocol.MethodIssueUpdate, body, nil)
}

func (c *Client) DeleteIssue(id int64) error {
	return c.mustOK(protocol.MethodIssueDelete, shareddto.NumericIDRequest{ID: id})
}

func (c *Client) ListSessionsByIssue(issueID int64) ([]Session, error) {
	var out []Session
	return out, c.call(protocol.MethodSessionListByIssue, shareddto.ListSessionsQuery{IssueID: &issueID}, &out)
}

func (c *Client) ListSessionHistory(issueID *int64, limit int) ([]SessionHistoryEntry, error) {
	var out []SessionHistoryEntry
	query := shareddto.SessionHistoryQuery{}
	if issueID != nil && *issueID > 0 {
		query.IssueID = issueID
	}
	if limit > 0 {
		query.Limit = &limit
	}
	return out, c.call(protocol.MethodSessionHistory, query, &out)
}

func (c *Client) GetSessionDetail(id string) (*SessionDetail, error) {
	var out SessionDetail
	return &out, c.call(protocol.MethodSessionDetail, shareddto.SessionIDRequest{ID: id}, &out)
}

func (c *Client) AmendSessionNote(id string, note string) error {
	req := shareddto.AmendSessionNoteRequest{Note: note}
	if strings.TrimSpace(id) != "" {
		req.ID = &id
	}
	return c.call(protocol.MethodSessionAmendNote, req, nil)
}

func (c *Client) LogManualSession(input shareddto.ManualSessionLogRequest) (*Session, error) {
	var out Session
	if err := c.call(protocol.MethodSessionLogManual, input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetDailySummary(date string) (*DailyIssueSummary, error) {
	var out DailyIssueSummary
	query := shareddto.DailyIssueSummaryQuery{}
	if strings.TrimSpace(date) != "" {
		trimmed := strings.TrimSpace(date)
		query.Date = &trimmed
	}
	return &out, c.call(protocol.MethodIssueDailySummary, query, &out)
}

func (c *Client) GetDailyPlan(date string) (*DailyPlan, error) {
	var out DailyPlan
	if err := c.call(protocol.MethodDailyPlanGet, shareddto.DailyPlanQuery{Date: strings.TrimSpace(date)}, &out); err != nil {
		return nil, err
	}
	if strings.TrimSpace(out.Date) == "" {
		return nil, nil
	}
	return &out, nil
}

func (c *Client) GetDailyCheckIn(date string) (*DailyCheckIn, error) {
	var out DailyCheckIn
	if err := c.call(protocol.MethodCheckInGet, shareddto.DailyCheckInQuery{Date: strings.TrimSpace(date)}, &out); err != nil {
		return nil, err
	}
	if strings.TrimSpace(out.Date) == "" {
		return nil, nil
	}
	return &out, nil
}

func (c *Client) ListDailyCheckIns(start, end string) ([]DailyCheckIn, error) {
	var out []DailyCheckIn
	return out, c.call(protocol.MethodCheckInRange, shareddto.DateRangeQuery{
		Start: strings.TrimSpace(start),
		End:   strings.TrimSpace(end),
	}, &out)
}

func (c *Client) UpsertDailyCheckIn(input shareddto.DailyCheckInUpsertRequest) (*DailyCheckIn, error) {
	var out DailyCheckIn
	if err := c.call(protocol.MethodCheckInUpsert, input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteDailyCheckIn(date string) error {
	return c.mustOK(protocol.MethodCheckInDelete, shareddto.DeleteByDateRequest{Date: strings.TrimSpace(date)})
}

func (c *Client) GetMetricsRange(start, end string) ([]DailyMetricsDay, error) {
	var out []DailyMetricsDay
	return out, c.call(protocol.MethodMetricsRange, shareddto.DateRangeQuery{
		Start: strings.TrimSpace(start),
		End:   strings.TrimSpace(end),
	}, &out)
}

func (c *Client) GetMetricsRollup(start, end string) (*MetricsRollup, error) {
	var out MetricsRollup
	if err := c.call(protocol.MethodMetricsRollup, shareddto.DateRangeQuery{
		Start: strings.TrimSpace(start),
		End:   strings.TrimSpace(end),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetMetricsStreaks(start, end string) (*StreakSummary, error) {
	var out StreakSummary
	if err := c.call(protocol.MethodMetricsStreaks, shareddto.DateRangeQuery{
		Start: strings.TrimSpace(start),
		End:   strings.TrimSpace(end),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetDashboardWindowSummary(start, end string, repoID, streamID, issueID *int64) (*DashboardWindowSummary, error) {
	var out DashboardWindowSummary
	if err := c.call(protocol.MethodDashboardWindow, shareddto.DashboardWindowQuery{
		Start:    strings.TrimSpace(start),
		End:      strings.TrimSpace(end),
		RepoID:   repoID,
		StreamID: streamID,
		IssueID:  issueID,
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetFocusScoreSummary(start, end string) (*FocusScoreSummary, error) {
	var out FocusScoreSummary
	if err := c.call(protocol.MethodDashboardFocusScore, shareddto.DashboardSummaryQuery{
		Start: strings.TrimSpace(start),
		End:   strings.TrimSpace(end),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetTimeDistributionSummary(start, end, groupBy string, repoID, streamID, issueID *int64) (*TimeDistributionSummary, error) {
	var out TimeDistributionSummary
	if err := c.call(protocol.MethodDashboardDistribution, shareddto.DashboardSummaryQuery{
		Start:    strings.TrimSpace(start),
		End:      strings.TrimSpace(end),
		GroupBy:  strings.TrimSpace(groupBy),
		RepoID:   repoID,
		StreamID: streamID,
		IssueID:  issueID,
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetGoalProgressSummary(start, end, groupBy string, repoID, streamID, issueID *int64) (*GoalProgressSummary, error) {
	var out GoalProgressSummary
	if err := c.call(protocol.MethodDashboardGoalProgress, shareddto.DashboardSummaryQuery{
		Start:    strings.TrimSpace(start),
		End:      strings.TrimSpace(end),
		GroupBy:  strings.TrimSpace(groupBy),
		RepoID:   repoID,
		StreamID: streamID,
		IssueID:  issueID,
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetExportAssets() (*ExportAssetStatus, error) {
	var out ExportAssetStatus
	return &out, c.call(protocol.MethodExportAssetsGet, nil, &out)
}

func (c *Client) SetExportReportsDir(path string) (*ExportAssetStatus, error) {
	var out ExportAssetStatus
	if err := c.call(protocol.MethodExportReportsDirSet, shareddto.ExportReportsDirUpdateRequest{
		ReportsDir: strings.TrimSpace(path),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) SetExportICSDir(path string) (*ExportAssetStatus, error) {
	var out ExportAssetStatus
	if err := c.call(protocol.MethodExportICSDirSet, shareddto.ExportICSDirUpdateRequest{
		ICSDir: strings.TrimSpace(path),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) ListExportReports() ([]ExportReportFile, error) {
	var out []ExportReportFile
	return out, c.call(protocol.MethodExportReportsList, nil, &out)
}

func (c *Client) DeleteExportReport(path string) error {
	return c.call(protocol.MethodExportReportsDelete, shareddto.ExportReportDeleteRequest{
		Path: strings.TrimSpace(path),
	}, nil)
}

func (c *Client) ResetExportTemplate(reportKind sharedtypes.ExportReportKind, assetKind sharedtypes.ExportAssetKind) (*ExportAssetStatus, error) {
	var out ExportAssetStatus
	if err := c.call(protocol.MethodExportTemplateReset, shareddto.ExportTemplateResetRequest{ReportKind: reportKind, AssetKind: assetKind}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) ApplyExportTemplatePreset(reportKind sharedtypes.ExportReportKind, assetKind sharedtypes.ExportAssetKind, presetID string) (*ExportAssetStatus, error) {
	var out ExportAssetStatus
	if err := c.call(protocol.MethodExportTemplateApply, shareddto.ExportTemplatePresetApplyRequest{
		ReportKind: reportKind,
		AssetKind:  assetKind,
		PresetID:   strings.TrimSpace(presetID),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GenerateReport(input shareddto.ExportReportRequest) (*ExportReportResult, error) {
	var out ExportReportResult
	method := protocol.MethodExportDaily
	switch input.Kind {
	case sharedtypes.ExportReportKindWeekly:
		method = protocol.MethodExportWeekly
	case sharedtypes.ExportReportKindRepo:
		method = protocol.MethodExportRepo
	case sharedtypes.ExportReportKindStream:
		method = protocol.MethodExportStream
	case sharedtypes.ExportReportKindIssueRollup:
		method = protocol.MethodExportIssueRollup
	case sharedtypes.ExportReportKindCSV:
		method = protocol.MethodExportCSV
	case sharedtypes.ExportReportKindCalendar:
		return nil, fmt.Errorf("calendar export uses GenerateCalendarExport")
	default:
		input.Kind = sharedtypes.ExportReportKindDaily
	}
	input.Date = strings.TrimSpace(input.Date)
	input.Start = strings.TrimSpace(input.Start)
	input.End = strings.TrimSpace(input.End)
	if err := c.call(method, input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GenerateCalendarExport(input shareddto.ExportCalendarRequest) (*CalendarExportResult, error) {
	var out CalendarExportResult
	if err := c.call(protocol.MethodExportCalendar, input, &out); err != nil {
		return nil, err
	}
	if strings.TrimSpace(out.IssuesFilePath) == "" || strings.TrimSpace(out.SessionsFilePath) == "" {
		return nil, fmt.Errorf("calendar export response is incomplete; restart the local engine so the updated export handler is loaded")
	}
	return &out, nil
}

func (c *Client) GenerateDailyReport(date string, format sharedtypes.ExportFormat, mode sharedtypes.ExportOutputMode) (*DailyReportResult, error) {
	return c.GenerateReport(shareddto.ExportReportRequest{
		Kind:       sharedtypes.ExportReportKindDaily,
		Date:       date,
		Format:     format,
		OutputMode: mode,
	})
}

func (c *Client) ChangeIssueStatus(issueID int64, status string, note *string) error {
	return c.call(protocol.MethodIssueChangeStatus, shareddto.ChangeIssueStatusRequest{
		ID:     issueID,
		Status: sharedtypes.IssueStatus(status),
		Note:   note,
	}, nil)
}

func (c *Client) MarkIssueTodoForToday(issueID int64) error {
	return c.SetIssueTodoDate(issueID, "")
}

func (c *Client) SetIssueTodoDate(issueID int64, date string) error {
	body := shareddto.SetIssueTodoRequest{ID: issueID}
	if strings.TrimSpace(date) != "" {
		trimmed := strings.TrimSpace(date)
		body.Date = &trimmed
	}
	return c.call(protocol.MethodIssueSetTodo, body, nil)
}

func (c *Client) ClearIssueTodo(issueID int64) error {
	return c.call(protocol.MethodIssueClearTodo, shareddto.NumericIDRequest{ID: issueID}, nil)
}

func (c *Client) GetContext() (*ActiveContext, error) {
	var out ActiveContext
	return &out, c.call(protocol.MethodContextGet, nil, &out)
}

func (c *Client) SwitchRepo(repoID int64) error {
	return c.call(protocol.MethodContextSwitchRepo, shareddto.SwitchRepoRequest{RepoID: repoID}, nil)
}

func (c *Client) SwitchStream(streamID int64) error {
	return c.call(protocol.MethodContextSwitchStream, shareddto.SwitchStreamRequest{StreamID: streamID}, nil)
}

func (c *Client) SwitchIssue(issueID int64) error {
	return c.call(protocol.MethodContextSwitchIssue, shareddto.SwitchIssueRequest{IssueID: issueID}, nil)
}

func (c *Client) SetFullContext(repoID, streamID, issueID int64) error {
	req := map[string]any{}
	if repoID != 0 {
		req["repoId"] = repoID
	}
	if streamID != 0 {
		req["streamId"] = streamID
	}
	if issueID != 0 {
		req["issueId"] = issueID
	}
	return c.call(protocol.MethodContextSet, req, nil)
}

func (c *Client) ClearContext() error {
	return c.mustOK(protocol.MethodContextClear, nil)
}

func (c *Client) GetTimerState() (*TimerState, error) {
	var out TimerState
	return &out, c.call(protocol.MethodTimerGetState, nil, &out)
}

func (c *Client) GetHealth() (*Health, error) {
	var out Health
	return &out, c.call(protocol.MethodHealthGet, nil, &out)
}

func (c *Client) GetUpdateStatus() (*UpdateStatus, error) {
	var out UpdateStatus
	return &out, c.call(protocol.MethodUpdateStatusGet, nil, &out)
}

func (c *Client) GetAlertStatus() (*AlertStatus, error) {
	var out AlertStatus
	return &out, c.call(protocol.MethodAlertsStatusGet, nil, &out)
}

func (c *Client) ListAlertReminders() ([]AlertReminder, error) {
	var out []AlertReminder
	return out, c.call(protocol.MethodAlertsRemindersList, nil, &out)
}

func (c *Client) CheckUpdateNow() (*UpdateStatus, error) {
	var out UpdateStatus
	return &out, c.call(protocol.MethodUpdateCheck, nil, &out)
}

func (c *Client) DismissUpdate() (*UpdateStatus, error) {
	var out UpdateStatus
	return &out, c.call(protocol.MethodUpdateDismiss, nil, &out)
}

func (c *Client) GetSettings() (*CoreSettings, error) {
	var raw json.RawMessage
	if err := c.call(protocol.MethodSettingsGetAll, nil, &raw); err != nil {
		return nil, err
	}
	return decodeSettings(raw)
}

func decodeSettings(raw json.RawMessage) (*CoreSettings, error) {
	var out map[string]CoreSettings
	if len(raw) == 0 {
		return nil, nil
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	if settings, ok := out["local"]; ok {
		return &settings, nil
	}
	for _, settings := range out {
		return &settings, nil
	}
	return nil, nil
}

func (c *Client) PatchSetting(key sharedtypes.CoreSettingsKey, value any) error {
	return c.call(protocol.MethodSettingsPatch, shareddto.PatchCoreSettingRequest{
		Key:   key,
		Value: value,
	}, nil)
}

func (c *Client) TestAlertNotification() error {
	return c.mustOK(protocol.MethodAlertsTestNotification, nil)
}

func (c *Client) TestAlertSound() error {
	return c.mustOK(protocol.MethodAlertsTestSound, nil)
}

func (c *Client) NotifyAlert(input sharedtypes.AlertRequest) error {
	return c.call(protocol.MethodAlertsNotify, input, nil)
}

func (c *Client) CreateAlertReminder(input shareddto.AlertReminderCreateRequest) (*AlertReminder, error) {
	var out AlertReminder
	return &out, c.call(protocol.MethodAlertsRemindersCreate, input, &out)
}

func (c *Client) UpdateAlertReminder(input shareddto.AlertReminderUpdateRequest) (*AlertReminder, error) {
	var out AlertReminder
	return &out, c.call(protocol.MethodAlertsRemindersUpdate, input, &out)
}

func (c *Client) DeleteAlertReminder(id string) error {
	return c.mustOK(protocol.MethodAlertsRemindersDelete, shareddto.AlertReminderIDRequest{ID: id})
}

func (c *Client) ToggleAlertReminder(id string, enabled bool) (*AlertReminder, error) {
	var out AlertReminder
	return &out, c.call(protocol.MethodAlertsRemindersToggle, shareddto.AlertReminderToggleRequest{ID: id, Enabled: enabled}, &out)
}

func (c *Client) ShutdownKernel() error {
	return c.mustOK(protocol.MethodKernelShutdown, nil)
}

func (c *Client) ShutdownKernelAndWait(timeout time.Duration) error {
	if err := c.ShutdownKernel(); err != nil {
		return err
	}
	if timeout <= 0 {
		timeout = defaultShutdownWait
	}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if _, err := c.GetHealth(); err != nil {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	if _, err := c.GetHealth(); err != nil {
		return nil
	}
	return fmt.Errorf("kernel shutdown timed out after %s", timeout)
}

func (c *Client) GetKernelInfo() (*KernelInfo, error) {
	var out KernelInfo
	return &out, c.call(protocol.MethodKernelInfoGet, nil, &out)
}

func (c *Client) SeedDevData() error {
	return c.mustOK(protocol.MethodKernelSeedDev, nil)
}

func (c *Client) ClearDevData() error {
	return c.mustOK(protocol.MethodKernelClearDev, nil)
}

func (c *Client) PrepareLocalUpdate() (*LocalUpdatePrepared, error) {
	var out LocalUpdatePrepared
	return &out, c.call(protocol.MethodKernelPrepareLocalUpdate, nil, &out)
}

func (c *Client) WipeRuntimeData() error {
	return c.mustOK(protocol.MethodKernelWipeData, shareddto.ConfirmDangerousActionRequest{Confirm: true})
}

func (c *Client) StartTimer(repoID, streamID, issueID int64, ignoreExistingStashes bool) error {
	req := shareddto.TimerStartRequest{}
	if repoID != 0 {
		req.RepoID = &repoID
	}
	if streamID != 0 {
		req.StreamID = &streamID
	}
	if issueID != 0 {
		req.IssueID = &issueID
	}
	req.IgnoreExistingStashes = ignoreExistingStashes
	return c.call(protocol.MethodTimerStart, req, nil)
}

func (c *Client) TouchTimerActivity() error {
	return c.mustOK(protocol.MethodTimerActivity, nil)
}

func (c *Client) PauseTimer() error {
	return c.call(protocol.MethodTimerPause, nil, nil)
}

func (c *Client) ResumeTimer() error {
	return c.call(protocol.MethodTimerResume, nil, nil)
}

func (c *Client) EndTimer(input shareddto.EndSessionRequest) error {
	return c.call(protocol.MethodTimerEnd, input, nil)
}

func (c *Client) StashPush(note string) error {
	body := shareddto.CreateStashRequest{}
	if note != "" {
		body.StashNote = &note
	}
	return c.call(protocol.MethodStashPush, body, nil)
}

func (c *Client) ListStashes() ([]Stash, error) {
	var out []Stash
	return out, c.call(protocol.MethodStashList, nil, &out)
}

func (c *Client) ApplyStash(id string) error {
	return c.mustOK(protocol.MethodStashApply, shareddto.StashIDRequest{ID: id})
}

func (c *Client) DropStash(id string) error {
	return c.mustOK(protocol.MethodStashDrop, shareddto.StashIDRequest{ID: id})
}

func (c *Client) ListScratchpads() ([]ScratchPad, error) {
	var out []ScratchPad
	return out, c.call(protocol.MethodScratchpadList, shareddto.ListScratchpadsQuery{}, &out)
}

func (c *Client) RegisterScratchpad(id, name, path string) error {
	pinned := false
	lastOpenedAt := time.Now().UTC().Format(time.RFC3339)
	body := shareddto.RegisterScratchpadRequest{
		ID:           &id,
		Name:         name,
		Path:         path,
		Pinned:       &pinned,
		LastOpenedAt: &lastOpenedAt,
	}
	return c.call(protocol.MethodScratchpadRegister, body, nil)
}

func (c *Client) ReadScratchpad(id string) (string, string, error) {
	var out sharedtypes.ScratchPadRead
	if err := c.call(protocol.MethodScratchpadRead, shareddto.ScratchpadIDRequest{ID: id}, &out); err != nil {
		return "", "", err
	}
	path := ""
	if out.Meta != nil {
		relativePath := out.Meta.Path
		if !strings.HasSuffix(relativePath, ".md") {
			relativePath += ".md"
		}
		if c.scratchDir != "" {
			path = filepath.Join(c.scratchDir, relativePath)
		} else {
			path = relativePath
		}
	}
	content := ""
	if out.Content != nil {
		content = *out.Content
	}
	return path, content, nil
}

func (c *Client) DeleteScratchpad(id string) error {
	return c.mustOK(protocol.MethodScratchpadDelete, shareddto.ScratchpadIDRequest{ID: id})
}

func (c *Client) ListOps(limit int) ([]Op, error) {
	var out []Op
	if limit <= 0 {
		limit = 50
	}
	return out, c.call(protocol.MethodOpsLatest, shareddto.ListLatestOpsQuery{Limit: &limit}, &out)
}

func (c *Client) ListOpsSince(since string) ([]Op, error) {
	var out []Op
	return out, c.call(protocol.MethodOpsSince, shareddto.ListOpsSinceQuery{Since: since}, &out)
}
