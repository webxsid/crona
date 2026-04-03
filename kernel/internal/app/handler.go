package app

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"crona/kernel/internal/core"
	corecommands "crona/kernel/internal/core/commands"
	"crona/kernel/internal/events"
	"crona/kernel/internal/export"
	"crona/kernel/internal/runtime"
	"crona/kernel/internal/scratchfile"
	"crona/kernel/internal/updatecheck"
	"crona/shared/config"
	shareddto "crona/shared/dto"
	"crona/shared/protocol"
	sharedtypes "crona/shared/types"
)

type Handler struct {
	startedAt string
	info      sharedtypes.KernelInfo
	pingDB    func(context.Context) error
	core      *core.Context
	bus       *events.Bus
	timer     *corecommands.TimerService
	shutdown  func()
	envMode   string
	paths     runtime.Paths
	updater   *updatecheck.Service
}

func NewHandler(startedAt string, info sharedtypes.KernelInfo, pingDB func(context.Context) error, coreCtx *core.Context, bus *events.Bus, shutdown func(), envMode string, paths runtime.Paths, updater *updatecheck.Service) *Handler {
	return &Handler{
		startedAt: startedAt,
		info:      info,
		pingDB:    pingDB,
		core:      coreCtx,
		bus:       bus,
		timer:     corecommands.GetTimerService(coreCtx),
		shutdown:  shutdown,
		envMode:   envMode,
		paths:     paths,
		updater:   updater,
	}
}

func (h *Handler) Stream(ctx context.Context, req protocol.Request, writer *json.Encoder) error {
	if req.Method != protocol.MethodEventsSubscribe {
		return errors.New("unsupported stream method")
	}

	eventsCh := make(chan sharedtypes.KernelEvent, 64)
	unsubscribe := h.bus.Subscribe(func(event sharedtypes.KernelEvent) {
		select {
		case eventsCh <- event:
		default:
		}
	})
	defer unsubscribe()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event := <-eventsCh:
			if err := writer.Encode(protocol.Event{
				Type:    event.Type,
				Payload: event.Payload,
			}); err != nil {
				return err
			}
		}
	}
}

func (h *Handler) Handle(ctx context.Context, req protocol.Request) protocol.Response {
	switch req.Method {
	case protocol.MethodHealthGet:
		ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		dbOK := h.pingDB == nil || h.pingDB(ctx) == nil
		return mustResult(req.ID, sharedtypes.Health{
			Status: map[bool]string{true: "ok", false: "degraded"}[dbOK],
			DB:     dbOK,
			OK:     map[bool]int{true: 1, false: 0}[dbOK],
			Uptime: time.Since(parseStartedAt(h.startedAt)).Seconds(),
		})
	case protocol.MethodKernelInfoGet:
		return mustResult(req.ID, h.info)
	case protocol.MethodKernelShutdown:
		if h.shutdown != nil {
			h.shutdown()
		}
		return mustResult(req.ID, shareddto.OKResponse{OK: true})
	case protocol.MethodKernelRestart:
		if h.shutdown != nil {
			h.shutdown()
		}
		return mustResult(req.ID, shareddto.OKResponse{OK: true})
	case protocol.MethodKernelSeedDev:
		return h.handleNoParams(req, func() (any, error) {
			if !strings.EqualFold(h.envMode, config.ModeDev) {
				return nil, errors.New("kernel.dev.seed is only available in Dev mode")
			}
			return shareddto.OKResponse{OK: true}, h.seedDevData(ctx)
		})
	case protocol.MethodKernelClearDev:
		return h.handleNoParams(req, func() (any, error) {
			if !strings.EqualFold(h.envMode, config.ModeDev) {
				return nil, errors.New("kernel.dev.clear is only available in Dev mode")
			}
			return shareddto.OKResponse{OK: true}, h.clearDevData(ctx)
		})
	case protocol.MethodKernelPrepareLocalUpdate:
		return h.handleNoParams(req, func() (any, error) {
			if !strings.EqualFold(h.envMode, config.ModeDev) {
				return nil, errors.New("kernel.dev.prepare_local_update is only available in Dev mode")
			}
			if h.updater == nil {
				return nil, errors.New("update service is unavailable")
			}
			return h.updater.PrepareLocalRelease(ctx)
		})
	case protocol.MethodKernelWipeData:
		return handle(req, func(input shareddto.ConfirmDangerousActionRequest) (any, error) {
			if !input.Confirm {
				return nil, errors.New("kernel.data.wipe requires explicit confirmation")
			}
			return shareddto.OKResponse{OK: true}, h.wipeRuntimeData(ctx)
		})
	case protocol.MethodUpdateStatusGet:
		return h.handleNoParams(req, func() (any, error) {
			if h.updater == nil {
				return sharedtypes.UpdateStatus{CurrentVersion: "unknown", Enabled: false}, nil
			}
			return h.updater.Status(), nil
		})
	case protocol.MethodUpdateCheck:
		return h.handleNoParams(req, func() (any, error) {
			if h.updater == nil {
				return sharedtypes.UpdateStatus{CurrentVersion: "unknown", Enabled: false}, nil
			}
			return h.updater.CheckNow(ctx)
		})
	case protocol.MethodUpdateDismiss:
		return h.handleNoParams(req, func() (any, error) {
			if h.updater == nil {
				return sharedtypes.UpdateStatus{CurrentVersion: "unknown", Enabled: false}, nil
			}
			return h.updater.DismissLatest()
		})

	case protocol.MethodRepoList:
		return h.handleNoParams(req, func() (any, error) {
			return corecommands.ListRepos(ctx, h.core)
		})
	case protocol.MethodRepoCreate:
		return handle(req, func(input shareddto.CreateRepoRequest) (any, error) {
			return corecommands.CreateRepo(ctx, h.core, struct {
				Name        string
				Description *string
				Color       *string
			}{Name: input.Name, Description: input.Description, Color: input.Color})
		})
	case protocol.MethodRepoUpdate:
		return h.handleNoParams(req, func() (any, error) {
			raw, err := decodeObject(req.Params)
			if err != nil {
				return nil, err
			}
			id, err := decodeRequiredInt64(raw, "id")
			if err != nil {
				return nil, err
			}
			name, nameSet, err := decodeOptionalStringFromMap(raw, "name")
			if err != nil {
				return nil, err
			}
			color, colorSet, err := decodeOptionalStringFromMap(raw, "color")
			if err != nil {
				return nil, err
			}
			description, descriptionSet, err := decodeOptionalStringFromMap(raw, "description")
			if err != nil {
				return nil, err
			}
			return corecommands.UpdateRepo(ctx, h.core, id, struct {
				Name        sharedtypes.Patch[string]
				Description sharedtypes.Patch[string]
				Color       sharedtypes.Patch[string]
			}{
				Name:        sharedtypes.Patch[string]{Set: nameSet, Value: name},
				Description: sharedtypes.Patch[string]{Set: descriptionSet, Value: description},
				Color:       sharedtypes.Patch[string]{Set: colorSet, Value: color},
			})
		})
	case protocol.MethodRepoDelete:
		return handle(req, func(input shareddto.NumericIDRequest) (any, error) {
			return shareddto.OKResponse{OK: true}, corecommands.DeleteRepo(ctx, h.core, input.ID)
		})

	case protocol.MethodStreamList:
		return handle(req, func(input shareddto.ListStreamsQuery) (any, error) {
			return corecommands.ListStreamsByRepo(ctx, h.core, input.RepoID)
		})
	case protocol.MethodStreamCreate:
		return handle(req, func(input shareddto.CreateStreamRequest) (any, error) {
			return corecommands.CreateStream(ctx, h.core, struct {
				RepoID      int64
				Name        string
				Description *string
				Visibility  *sharedtypes.StreamVisibility
			}{
				RepoID:      input.RepoID,
				Name:        input.Name,
				Description: input.Description,
				Visibility:  input.Visibility,
			})
		})
	case protocol.MethodStreamUpdate:
		return h.handleNoParams(req, func() (any, error) {
			raw, err := decodeObject(req.Params)
			if err != nil {
				return nil, err
			}
			id, err := decodeRequiredInt64(raw, "id")
			if err != nil {
				return nil, err
			}
			name, _, err := decodeOptionalStringFromMap(raw, "name")
			if err != nil {
				return nil, err
			}
			description, descriptionSet, err := decodeOptionalStringFromMap(raw, "description")
			if err != nil {
				return nil, err
			}
			var visibility *sharedtypes.StreamVisibility
			if rawValue, ok := raw["visibility"]; ok && string(rawValue) != "null" {
				var out sharedtypes.StreamVisibility
				if err := json.Unmarshal(rawValue, &out); err != nil {
					return nil, err
				}
				visibility = &out
			}
			return corecommands.UpdateStream(ctx, h.core, id, struct {
				Name        *string
				Description sharedtypes.Patch[string]
				Visibility  *sharedtypes.StreamVisibility
			}{Name: name, Description: sharedtypes.Patch[string]{Set: descriptionSet, Value: description}, Visibility: visibility})
		})
	case protocol.MethodStreamDelete:
		return handle(req, func(input shareddto.NumericIDRequest) (any, error) {
			return shareddto.OKResponse{OK: true}, corecommands.DeleteStream(ctx, h.core, input.ID)
		})

	case protocol.MethodIssueList:
		return handle(req, func(input shareddto.ListIssuesQuery) (any, error) {
			return corecommands.ListIssuesByStream(ctx, h.core, input.StreamID)
		})
	case protocol.MethodIssueListAll:
		return h.handleNoParams(req, func() (any, error) {
			return corecommands.ListAllIssues(ctx, h.core)
		})
	case protocol.MethodIssueCreate:
		return handle(req, func(input shareddto.CreateIssueRequest) (any, error) {
			return corecommands.CreateIssue(ctx, h.core, struct {
				StreamID        int64
				Title           string
				Description     *string
				EstimateMinutes *int
				Notes           *string
				TodoForDate     *string
			}{
				StreamID:        input.StreamID,
				Title:           input.Title,
				Description:     input.Description,
				EstimateMinutes: input.EstimateMinutes,
				Notes:           input.Notes,
				TodoForDate:     input.TodoForDate,
			})
		})
	case protocol.MethodIssueUpdate:
		return h.handleNoParams(req, func() (any, error) {
			raw, err := decodeObject(req.Params)
			if err != nil {
				return nil, err
			}
			id, err := decodeRequiredInt64(raw, "id")
			if err != nil {
				return nil, err
			}
			title, titleSet, err := decodeOptionalStringFromMap(raw, "title")
			if err != nil {
				return nil, err
			}
			estimate, estimateSet, err := decodeOptionalIntFromMap(raw, "estimateMinutes")
			if err != nil {
				return nil, err
			}
			notes, notesSet, err := decodeOptionalStringFromMap(raw, "notes")
			if err != nil {
				return nil, err
			}
			description, descriptionSet, err := decodeOptionalStringFromMap(raw, "description")
			if err != nil {
				return nil, err
			}
			return corecommands.UpdateIssue(ctx, h.core, id, struct {
				Title           sharedtypes.Patch[string]
				Description     sharedtypes.Patch[string]
				EstimateMinutes sharedtypes.Patch[int]
				Notes           sharedtypes.Patch[string]
			}{
				Title:           sharedtypes.Patch[string]{Set: titleSet, Value: title},
				Description:     sharedtypes.Patch[string]{Set: descriptionSet, Value: description},
				EstimateMinutes: sharedtypes.Patch[int]{Set: estimateSet, Value: estimate},
				Notes:           sharedtypes.Patch[string]{Set: notesSet, Value: notes},
			})
		})
	case protocol.MethodIssueDelete:
		return handle(req, func(input shareddto.NumericIDRequest) (any, error) {
			return shareddto.OKResponse{OK: true}, corecommands.DeleteIssue(ctx, h.core, input.ID)
		})
	case protocol.MethodIssueChangeStatus:
		return handle(req, func(input shareddto.ChangeIssueStatusRequest) (any, error) {
			return corecommands.ChangeIssueStatus(ctx, h.core, input.ID, input.Status, input.Note)
		})
	case protocol.MethodIssueSetTodo:
		return h.handleNoParams(req, func() (any, error) {
			raw, err := decodeObject(req.Params)
			if err != nil {
				return nil, err
			}
			id, err := decodeRequiredInt64(raw, "id")
			if err != nil {
				return nil, err
			}
			date, dateSet, err := decodeOptionalStringFromMap(raw, "date")
			if err != nil {
				return nil, err
			}
			if dateSet && date == nil {
				return corecommands.ClearIssueTodoForDate(ctx, h.core, id)
			}
			if dateSet && date != nil && *date != "" {
				return corecommands.MarkIssueTodoForDate(ctx, h.core, id, *date)
			}
			return corecommands.MarkIssueTodoForToday(ctx, h.core, id)
		})
	case protocol.MethodIssueClearTodo:
		return handle(req, func(input shareddto.NumericIDRequest) (any, error) {
			return corecommands.ClearIssueTodoForDate(ctx, h.core, input.ID)
		})
	case protocol.MethodIssueDailySummary:
		return handle(req, func(input shareddto.DailyIssueSummaryQuery) (any, error) {
			if input.Date != nil && *input.Date != "" {
				return corecommands.ComputeDailyIssueSummaryForDate(ctx, h.core, *input.Date)
			}
			return corecommands.ComputeDailyIssueSummaryForToday(ctx, h.core)
		})
	case protocol.MethodIssueTodaySummary:
		return h.handleNoParams(req, func() (any, error) {
			return corecommands.ComputeDailyIssueSummaryForToday(ctx, h.core)
		})
	case protocol.MethodDailyPlanGet:
		return handle(req, func(input shareddto.DailyPlanQuery) (any, error) {
			return corecommands.GetDailyPlan(ctx, h.core, input.Date)
		})
	case protocol.MethodHabitList:
		return handle(req, func(input shareddto.ListHabitsQuery) (any, error) {
			return corecommands.ListHabitsByStream(ctx, h.core, input.StreamID)
		})
	case protocol.MethodHabitListDue:
		return handle(req, func(input shareddto.ListHabitsDueQuery) (any, error) {
			return corecommands.ListHabitsDueForDate(ctx, h.core, input.Date)
		})
	case protocol.MethodHabitCreate:
		return handle(req, func(input shareddto.CreateHabitRequest) (any, error) {
			return corecommands.CreateHabit(ctx, h.core, struct {
				StreamID      int64
				Name          string
				Description   *string
				ScheduleType  string
				Weekdays      []int
				TargetMinutes *int
			}{
				StreamID:      input.StreamID,
				Name:          input.Name,
				Description:   input.Description,
				ScheduleType:  input.ScheduleType,
				Weekdays:      input.Weekdays,
				TargetMinutes: input.TargetMinutes,
			})
		})
	case protocol.MethodHabitUpdate:
		return handle(req, func(input shareddto.UpdateHabitRequest) (any, error) {
			var active *bool
			if input.Active != nil {
				active = input.Active
			}
			return corecommands.UpdateHabit(ctx, h.core, input.ID, struct {
				Name          sharedtypes.Patch[string]
				Description   sharedtypes.Patch[string]
				ScheduleType  *string
				Weekdays      []int
				WeekdaysSet   bool
				TargetMinutes sharedtypes.Patch[int]
				Active        *bool
			}{
				Name:          sharedtypes.Patch[string]{Set: input.Name != nil, Value: input.Name},
				Description:   sharedtypes.Patch[string]{Set: true, Value: input.Description},
				ScheduleType:  input.ScheduleType,
				Weekdays:      input.Weekdays,
				WeekdaysSet:   input.Weekdays != nil,
				TargetMinutes: sharedtypes.Patch[int]{Set: true, Value: input.TargetMinutes},
				Active:        active,
			})
		})
	case protocol.MethodHabitDelete:
		return handle(req, func(input shareddto.NumericIDRequest) (any, error) {
			return shareddto.OKResponse{OK: true}, corecommands.DeleteHabit(ctx, h.core, input.ID)
		})
	case protocol.MethodHabitComplete:
		return handle(req, func(input shareddto.HabitCompletionUpsertRequest) (any, error) {
			status := sharedtypes.HabitCompletionStatusCompleted
			if input.Status != nil {
				status = *input.Status
			}
			return corecommands.CompleteHabit(ctx, h.core, input.HabitID, input.Date, status, input.DurationMinutes, input.Notes)
		})
	case protocol.MethodHabitUncomplete:
		return handle(req, func(input shareddto.HabitCompletionUpsertRequest) (any, error) {
			return shareddto.OKResponse{OK: true}, corecommands.UncompleteHabit(ctx, h.core, input.HabitID, input.Date)
		})
	case protocol.MethodHabitHistory:
		return handle(req, func(input shareddto.HabitHistoryQuery) (any, error) {
			return corecommands.ListHabitHistory(ctx, h.core, input.HabitID)
		})
	case protocol.MethodCheckInGet:
		return handle(req, func(input shareddto.DailyCheckInQuery) (any, error) {
			return corecommands.GetDailyCheckIn(ctx, h.core, input.Date)
		})
	case protocol.MethodCheckInUpsert:
		return handle(req, func(input shareddto.DailyCheckInUpsertRequest) (any, error) {
			return corecommands.UpsertDailyCheckIn(ctx, h.core, input)
		})
	case protocol.MethodCheckInDelete:
		return handle(req, func(input shareddto.DeleteByDateRequest) (any, error) {
			return shareddto.OKResponse{OK: true}, corecommands.DeleteDailyCheckIn(ctx, h.core, input.Date)
		})
	case protocol.MethodCheckInRange:
		return handle(req, func(input shareddto.DateRangeQuery) (any, error) {
			return corecommands.ListDailyCheckInsInRange(ctx, h.core, input.Start, input.End)
		})
	case protocol.MethodMetricsRange:
		return handle(req, func(input shareddto.DateRangeQuery) (any, error) {
			return corecommands.ComputeMetricsRange(ctx, h.core, input.Start, input.End)
		})
	case protocol.MethodMetricsRollup:
		return handle(req, func(input shareddto.DateRangeQuery) (any, error) {
			return corecommands.ComputeMetricsRollup(ctx, h.core, input.Start, input.End)
		})
	case protocol.MethodMetricsStreaks:
		return handle(req, func(input shareddto.DateRangeQuery) (any, error) {
			return corecommands.ComputeMetricsStreaks(ctx, h.core, input.Start, input.End)
		})
	case protocol.MethodDashboardWindow:
		return handle(req, func(input shareddto.DashboardWindowQuery) (any, error) {
			return corecommands.ComputeDashboardWindowSummary(ctx, h.core, input)
		})
	case protocol.MethodDashboardFocusScore:
		return handle(req, func(input shareddto.DashboardSummaryQuery) (any, error) {
			return corecommands.ComputeFocusScoreSummary(ctx, h.core, input)
		})
	case protocol.MethodDashboardDistribution:
		return handle(req, func(input shareddto.DashboardSummaryQuery) (any, error) {
			return corecommands.ComputeTimeDistributionSummary(ctx, h.core, input)
		})
	case protocol.MethodDashboardGoalProgress:
		return handle(req, func(input shareddto.DashboardSummaryQuery) (any, error) {
			return corecommands.ComputeGoalProgressSummary(ctx, h.core, input)
		})
	case protocol.MethodExportAssetsGet:
		return h.handleNoParams(req, func() (any, error) {
			return export.EnsureAssets(h.paths)
		})
	case protocol.MethodExportReportsDirSet:
		return handle(req, func(input shareddto.ExportReportsDirUpdateRequest) (any, error) {
			return export.SetReportsDir(h.paths, input.ReportsDir)
		})
	case protocol.MethodExportICSDirSet:
		return handle(req, func(input shareddto.ExportICSDirUpdateRequest) (any, error) {
			return export.SetICSDir(h.paths, input.ICSDir)
		})
	case protocol.MethodExportReportsList:
		return h.handleNoParams(req, func() (any, error) {
			return export.ListReports(h.paths)
		})
	case protocol.MethodExportReportsDelete:
		return handle(req, func(input shareddto.ExportReportDeleteRequest) (any, error) {
			if err := export.DeleteReport(h.paths, input.Path); err != nil {
				return nil, err
			}
			return shareddto.OKResponse{OK: true}, nil
		})
	case protocol.MethodExportTemplateReset:
		return handle(req, func(input shareddto.ExportTemplateResetRequest) (any, error) {
			return export.ResetTemplate(h.paths, input.ReportKind, input.AssetKind)
		})
	case protocol.MethodExportTemplateApply:
		return handle(req, func(input shareddto.ExportTemplatePresetApplyRequest) (any, error) {
			return export.ApplyTemplatePreset(h.paths, input.ReportKind, input.AssetKind, input.PresetID)
		})
	case protocol.MethodExportDaily:
		return handle(req, func(input shareddto.DailyReportRequest) (any, error) {
			input.Kind = sharedtypes.ExportReportKindDaily
			return export.GenerateReport(ctx, h.core, h.paths, input)
		})
	case protocol.MethodExportWeekly:
		return handle(req, func(input shareddto.ExportReportRequest) (any, error) {
			input.Kind = sharedtypes.ExportReportKindWeekly
			return export.GenerateReport(ctx, h.core, h.paths, input)
		})
	case protocol.MethodExportRepo:
		return handle(req, func(input shareddto.ExportReportRequest) (any, error) {
			input.Kind = sharedtypes.ExportReportKindRepo
			return export.GenerateReport(ctx, h.core, h.paths, input)
		})
	case protocol.MethodExportStream:
		return handle(req, func(input shareddto.ExportReportRequest) (any, error) {
			input.Kind = sharedtypes.ExportReportKindStream
			return export.GenerateReport(ctx, h.core, h.paths, input)
		})
	case protocol.MethodExportIssueRollup:
		return handle(req, func(input shareddto.ExportReportRequest) (any, error) {
			input.Kind = sharedtypes.ExportReportKindIssueRollup
			return export.GenerateReport(ctx, h.core, h.paths, input)
		})
	case protocol.MethodExportCSV:
		return handle(req, func(input shareddto.ExportReportRequest) (any, error) {
			input.Kind = sharedtypes.ExportReportKindCSV
			return export.GenerateReport(ctx, h.core, h.paths, input)
		})
	case protocol.MethodExportCalendar:
		return handle(req, func(input shareddto.ExportCalendarRequest) (any, error) {
			return export.GenerateCalendarExport(ctx, h.core, h.paths, input)
		})

	case protocol.MethodContextGet:
		return h.handleNoParams(req, func() (any, error) {
			return corecommands.GetActiveContext(ctx, h.core)
		})
	case protocol.MethodContextSet:
		return h.handleNoParams(req, func() (any, error) {
			raw, err := decodeObject(req.Params)
			if err != nil {
				return nil, err
			}
			repoID, repoSet, err := decodeOptionalInt64FromMap(raw, "repoId")
			if err != nil {
				return nil, err
			}
			streamID, streamSet, err := decodeOptionalInt64FromMap(raw, "streamId")
			if err != nil {
				return nil, err
			}
			issueID, issueSet, err := decodeOptionalInt64FromMap(raw, "issueId")
			if err != nil {
				return nil, err
			}
			return corecommands.SetContext(ctx, h.core, corecommands.ContextPatch{
				RepoSet:   repoSet,
				RepoID:    repoID,
				StreamSet: streamSet,
				StreamID:  streamID,
				IssueSet:  issueSet,
				IssueID:   issueID,
			})
		})
	case protocol.MethodContextSwitchRepo:
		return handle(req, func(input shareddto.SwitchRepoRequest) (any, error) {
			return corecommands.SwitchRepo(ctx, h.core, input.RepoID)
		})
	case protocol.MethodContextSwitchStream:
		return handle(req, func(input shareddto.SwitchStreamRequest) (any, error) {
			return corecommands.SwitchStream(ctx, h.core, input.StreamID)
		})
	case protocol.MethodContextSwitchIssue:
		return handle(req, func(input shareddto.SwitchIssueRequest) (any, error) {
			return corecommands.SwitchIssue(ctx, h.core, input.IssueID)
		})
	case protocol.MethodContextClearIssue:
		return h.handleNoParams(req, func() (any, error) {
			return corecommands.ClearIssue(ctx, h.core)
		})
	case protocol.MethodContextClear:
		return h.handleNoParams(req, func() (any, error) {
			return shareddto.OKResponse{OK: true}, corecommands.ClearContext(ctx, h.core)
		})

	case protocol.MethodSessionListByIssue:
		return handle(req, func(input shareddto.ListSessionsQuery) (any, error) {
			if input.IssueID == nil || *input.IssueID == 0 {
				return nil, errors.New("issueId is required")
			}
			return h.core.Sessions.ListByIssue(ctx, *input.IssueID, h.core.UserID)
		})
	case protocol.MethodSessionGet:
		return handle(req, func(input shareddto.SessionIDRequest) (any, error) {
			return h.core.Sessions.GetByID(ctx, input.ID, h.core.UserID)
		})
	case protocol.MethodSessionDetail:
		return handle(req, func(input shareddto.SessionIDRequest) (any, error) {
			return corecommands.GetSessionDetail(ctx, h.core, input.ID)
		})
	case protocol.MethodSessionGetActive:
		return h.handleNoParams(req, func() (any, error) {
			return h.core.Sessions.GetActiveSession(ctx, h.core.UserID)
		})
	case protocol.MethodSessionStart:
		return handle(req, func(input shareddto.StartSessionRequest) (any, error) {
			return corecommands.StartSession(ctx, h.core, input.IssueID)
		})
	case protocol.MethodSessionPause:
		return h.handleNoParams(req, func() (any, error) {
			return shareddto.OKResponse{OK: true}, corecommands.PauseSession(ctx, h.core, sharedtypes.SessionSegmentRest)
		})
	case protocol.MethodSessionResume:
		return h.handleNoParams(req, func() (any, error) {
			return shareddto.OKResponse{OK: true}, corecommands.ResumeSession(ctx, h.core)
		})
	case protocol.MethodSessionEnd:
		return handle(req, func(input shareddto.EndSessionRequest) (any, error) {
			return corecommands.StopSession(ctx, h.core, corecommands.SessionEndInput{
				CommitMessage: input.CommitMessage,
				WorkedOn:      input.WorkedOn,
				Outcome:       input.Outcome,
				NextStep:      input.NextStep,
				Blockers:      input.Blockers,
				Links:         input.Links,
			})
		})
	case protocol.MethodSessionLogManual:
		return handle(req, func(input shareddto.ManualSessionLogRequest) (any, error) {
			return corecommands.LogManualSession(ctx, h.core, corecommands.ManualSessionInput{
				IssueID:              input.IssueID,
				Date:                 input.Date,
				WorkDurationSeconds:  input.WorkDurationSeconds,
				BreakDurationSeconds: input.BreakDurationSeconds,
				StartTime:            input.StartTime,
				EndTime:              input.EndTime,
				CommitMessage:        input.CommitMessage,
				Notes:                input.Notes,
			})
		})
	case protocol.MethodSessionAmendNote:
		return handle(req, func(input shareddto.AmendSessionNoteRequest) (any, error) {
			return corecommands.AmendSessionNotes(ctx, h.core, input.Note, input.ID)
		})
	case protocol.MethodSessionHistory:
		return handle(req, func(input shareddto.SessionHistoryQuery) (any, error) {
			useContext := input.Context != nil && *input.Context
			return corecommands.ListSessionHistory(ctx, h.core, struct {
				RepoID   *int64
				StreamID *int64
				IssueID  *int64
				Since    *string
				Until    *string
				Limit    *int
				Offset   *int
			}{
				RepoID:   input.RepoID,
				StreamID: input.StreamID,
				IssueID:  input.IssueID,
				Since:    input.Since,
				Until:    input.Until,
				Limit:    input.Limit,
				Offset:   input.Offset,
			}, useContext)
		})

	case protocol.MethodTimerGetState:
		return h.handleNoParams(req, func() (any, error) {
			return h.timer.GetState(ctx)
		})
	case protocol.MethodTimerStart:
		return handle(req, func(input shareddto.TimerStartRequest) (any, error) {
			return h.timer.Start(ctx, input.IssueID)
		})
	case protocol.MethodTimerPause:
		return h.handleNoParams(req, func() (any, error) {
			return h.timer.Pause(ctx)
		})
	case protocol.MethodTimerResume:
		return h.handleNoParams(req, func() (any, error) {
			return h.timer.Resume(ctx)
		})
	case protocol.MethodTimerEnd:
		return handle(req, func(input shareddto.EndSessionRequest) (any, error) {
			return h.timer.End(ctx, corecommands.SessionEndInput{
				CommitMessage: input.CommitMessage,
				WorkedOn:      input.WorkedOn,
				Outcome:       input.Outcome,
				NextStep:      input.NextStep,
				Blockers:      input.Blockers,
				Links:         input.Links,
			})
		})

	case protocol.MethodStashList:
		return h.handleNoParams(req, func() (any, error) {
			return corecommands.ListStashes(ctx, h.core)
		})
	case protocol.MethodStashGet:
		return handle(req, func(input shareddto.StashIDRequest) (any, error) {
			return corecommands.GetStash(ctx, h.core, input.ID)
		})
	case protocol.MethodStashPush:
		return handle(req, func(input shareddto.CreateStashRequest) (any, error) {
			return corecommands.StashPush(ctx, h.core, input.StashNote)
		})
	case protocol.MethodStashApply:
		return handle(req, func(input shareddto.StashIDRequest) (any, error) {
			return shareddto.OKResponse{OK: true}, corecommands.StashPop(ctx, h.core, h.timer, input.ID)
		})
	case protocol.MethodStashDrop:
		return handle(req, func(input shareddto.StashIDRequest) (any, error) {
			return shareddto.OKResponse{OK: true}, corecommands.StashDrop(ctx, h.core, input.ID)
		})

	case protocol.MethodScratchpadList:
		return handle(req, func(input shareddto.ListScratchpadsQuery) (any, error) {
			return corecommands.ListScratchpads(ctx, h.core, input.PinnedOnly != nil && *input.PinnedOnly)
		})
	case protocol.MethodScratchpadRegister:
		return handle(req, func(input shareddto.RegisterScratchpadRequest) (any, error) {
			pinned := false
			if input.Pinned != nil {
				pinned = *input.Pinned
			}
			lastOpenedAt := ""
			if input.LastOpenedAt != nil {
				lastOpenedAt = *input.LastOpenedAt
			}
			id := ""
			if input.ID != nil {
				id = *input.ID
			}
			filePath, err := corecommands.RegisterScratchpad(ctx, h.core, sharedtypes.ScratchPadMeta{
				ID:           id,
				Name:         input.Name,
				Path:         input.Path,
				Pinned:       pinned,
				LastOpenedAt: lastOpenedAt,
			})
			if err != nil {
				return nil, err
			}
			if _, err := scratchfile.Create(h.core.ScratchDir, filePath, input.Name); err != nil {
				_ = corecommands.RemoveScratchpad(ctx, h.core, id)
				return nil, err
			}
			return map[string]any{"ok": true, "filePath": filePath}, nil
		})
	case protocol.MethodScratchpadGetMeta:
		return handle(req, func(input shareddto.ScratchpadIDRequest) (any, error) {
			meta, err := corecommands.GetScratchpad(ctx, h.core, input.ID)
			if err != nil {
				return nil, err
			}
			if meta == nil {
				return shareddto.ErrorResponse{OK: false, Error: "Scratchpad not found"}, nil
			}
			return map[string]any{"ok": true, "meta": meta}, nil
		})
	case protocol.MethodScratchpadRead:
		return handle(req, func(input shareddto.ScratchpadIDRequest) (any, error) {
			meta, err := corecommands.GetScratchpad(ctx, h.core, input.ID)
			if err != nil {
				return nil, err
			}
			if meta == nil {
				return sharedtypes.ScratchPadRead{OK: false, Error: stringPtr("Scratchpad not found")}, nil
			}
			content, err := scratchfile.Read(h.core.ScratchDir, meta.Path)
			if err != nil {
				return nil, err
			}
			return sharedtypes.ScratchPadRead{OK: true, Meta: meta, Content: &content}, nil
		})
	case protocol.MethodScratchpadPin:
		return handle(req, func(input shareddto.PinScratchpadRequest) (any, error) {
			return shareddto.OKResponse{OK: true}, corecommands.PinScratchpad(ctx, h.core, input.ID, input.Pinned)
		})
	case protocol.MethodScratchpadDelete:
		return handle(req, func(input shareddto.ScratchpadIDRequest) (any, error) {
			meta, err := corecommands.GetScratchpad(ctx, h.core, input.ID)
			if err != nil {
				return nil, err
			}
			if err := corecommands.RemoveScratchpad(ctx, h.core, input.ID); err != nil {
				return nil, err
			}
			if meta != nil {
				if err := scratchfile.Delete(h.core.ScratchDir, meta.Path); err != nil {
					return nil, err
				}
			}
			return shareddto.OKResponse{OK: true}, nil
		})

	case protocol.MethodSettingsGetAll:
		return h.handleNoParams(req, func() (any, error) {
			return h.core.CoreSettings.GetAllSettings(ctx)
		})
	case protocol.MethodSettingsGet:
		return handle(req, func(input shareddto.GetCoreSettingRequest) (any, error) {
			return h.core.CoreSettings.GetSetting(ctx, h.core.UserID, input.Key)
		})
	case protocol.MethodSettingsPatch:
		return handle(req, func(input shareddto.PatchCoreSettingRequest) (any, error) {
			if err := h.core.CoreSettings.SetSetting(ctx, h.core.UserID, input.Key, input.Value); err != nil {
				return nil, err
			}
			return shareddto.OKResponse{OK: true}, nil
		})
	case protocol.MethodSettingsPut:
		return handle(req, func(input shareddto.PutCoreSettingsRequest) (any, error) {
			updated := map[string]any{}
			for key, value := range input {
				if err := h.core.CoreSettings.SetSetting(ctx, h.core.UserID, key, value); err != nil {
					return nil, err
				}
				updated[string(key)] = value
			}
			return updated, nil
		})

	case protocol.MethodOpsLatest:
		return handle(req, func(input shareddto.ListLatestOpsQuery) (any, error) {
			limit := 50
			if input.Limit != nil {
				limit = *input.Limit
			}
			return corecommands.ListLatestOps(ctx, h.core, limit)
		})
	case protocol.MethodOpsSince:
		return handle(req, func(input shareddto.ListOpsSinceQuery) (any, error) {
			return corecommands.ListOpsSince(ctx, h.core, input.Since)
		})
	case protocol.MethodOpsList:
		return handle(req, func(input shareddto.ListOpsQuery) (any, error) {
			if input.Entity == nil || input.EntityID == nil {
				return nil, errors.New("entity and entityId are required")
			}
			limit := 100
			if input.Limit != nil {
				limit = *input.Limit
			}
			return corecommands.ListOpsByEntity(ctx, h.core, *input.Entity, *input.EntityID, limit)
		})

	default:
		return protocol.Response{
			ID: req.ID,
			Error: &protocol.Error{
				Code:    "not_implemented",
				Message: "kernel method not implemented yet",
			},
		}
	}
}

func (h *Handler) handleNoParams(req protocol.Request, fn func() (any, error)) protocol.Response {
	value, err := fn()
	if err != nil {
		return errorResponse(req.ID, err)
	}
	return mustResult(req.ID, value)
}

func handle[T any](req protocol.Request, fn func(T) (any, error)) protocol.Response {
	var input T
	if len(req.Params) > 0 {
		if err := json.Unmarshal(req.Params, &input); err != nil {
			return errorResponse(req.ID, err)
		}
	}
	value, err := fn(input)
	if err != nil {
		return errorResponse(req.ID, err)
	}
	return mustResult(req.ID, value)
}

func parseStartedAt(startedAt string) time.Time {
	t, err := time.Parse(time.RFC3339, startedAt)
	if err != nil {
		return time.Now().UTC()
	}
	return t
}

func mustResult(id string, value any) protocol.Response {
	body, err := json.Marshal(value)
	if err != nil {
		return protocol.Response{
			ID: id,
			Error: &protocol.Error{
				Code:    "internal_error",
				Message: err.Error(),
			},
		}
	}

	return protocol.Response{
		ID:     id,
		Result: body,
	}
}

func errorResponse(id string, err error) protocol.Response {
	return protocol.Response{
		ID: id,
		Error: &protocol.Error{
			Code:    "request_failed",
			Message: err.Error(),
		},
	}
}

func decodeObject(raw json.RawMessage) (map[string]json.RawMessage, error) {
	if len(raw) == 0 {
		return map[string]json.RawMessage{}, nil
	}
	var out map[string]json.RawMessage
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func decodeRequiredInt64(raw map[string]json.RawMessage, key string) (int64, error) {
	value, ok := raw[key]
	if !ok {
		return 0, errors.New(key + " is required")
	}
	var out int64
	if err := json.Unmarshal(value, &out); err != nil {
		return 0, err
	}
	return out, nil
}

func decodeOptionalStringFromMap(raw map[string]json.RawMessage, key string) (*string, bool, error) {
	value, ok := raw[key]
	if !ok {
		return nil, false, nil
	}
	if string(value) == "null" {
		return nil, true, nil
	}
	var out string
	if err := json.Unmarshal(value, &out); err != nil {
		return nil, false, err
	}
	return &out, true, nil
}

func decodeOptionalIntFromMap(raw map[string]json.RawMessage, key string) (*int, bool, error) {
	value, ok := raw[key]
	if !ok {
		return nil, false, nil
	}
	if string(value) == "null" {
		return nil, true, nil
	}
	var out int
	if err := json.Unmarshal(value, &out); err != nil {
		return nil, false, err
	}
	return &out, true, nil
}

func decodeOptionalInt64FromMap(raw map[string]json.RawMessage, key string) (*int64, bool, error) {
	value, ok := raw[key]
	if !ok {
		return nil, false, nil
	}
	if string(value) == "null" {
		return nil, true, nil
	}
	var out int64
	if err := json.Unmarshal(value, &out); err != nil {
		return nil, false, err
	}
	return &out, true, nil
}

func stringPtr(value string) *string {
	return &value
}
