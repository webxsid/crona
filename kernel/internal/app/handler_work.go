package app

import (
	"context"
	"encoding/json"

	corecommands "crona/kernel/internal/core/commands"
	"crona/kernel/internal/export"
	shareddto "crona/shared/dto"
	"crona/shared/protocol"
	sharedtypes "crona/shared/types"
)

func (h *Handler) handleWorkMethods(ctx context.Context, req protocol.Request) (protocol.Response, bool) {
	switch req.Method {
	case protocol.MethodRepoList:
		return h.handleNoParams(req, func() (any, error) {
			return corecommands.ListRepos(ctx, h.core)
		}), true
	case protocol.MethodRepoCreate:
		return handle(req, func(input shareddto.CreateRepoRequest) (any, error) {
			return corecommands.CreateRepo(ctx, h.core, struct {
				Name        string
				Description *string
				Color       *string
			}{Name: input.Name, Description: input.Description, Color: input.Color})
		}), true
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
		}), true
	case protocol.MethodRepoDelete:
		return handle(req, func(input shareddto.NumericIDRequest) (any, error) {
			return shareddto.OKResponse{OK: true}, corecommands.DeleteRepo(ctx, h.core, input.ID)
		}), true
	case protocol.MethodStreamList:
		return handle(req, func(input shareddto.ListStreamsQuery) (any, error) {
			return corecommands.ListStreamsByRepo(ctx, h.core, input.RepoID)
		}), true
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
		}), true
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
		}), true
	case protocol.MethodStreamDelete:
		return handle(req, func(input shareddto.NumericIDRequest) (any, error) {
			return shareddto.OKResponse{OK: true}, corecommands.DeleteStream(ctx, h.core, input.ID)
		}), true
	case protocol.MethodIssueList:
		return handle(req, func(input shareddto.ListIssuesQuery) (any, error) {
			return corecommands.ListIssuesByStream(ctx, h.core, input.StreamID)
		}), true
	case protocol.MethodIssueListAll:
		return h.handleNoParams(req, func() (any, error) {
			return corecommands.ListAllIssues(ctx, h.core)
		}), true
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
		}), true
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
		}), true
	case protocol.MethodIssueDelete:
		return handle(req, func(input shareddto.NumericIDRequest) (any, error) {
			return shareddto.OKResponse{OK: true}, corecommands.DeleteIssue(ctx, h.core, input.ID)
		}), true
	case protocol.MethodIssueChangeStatus:
		return handle(req, func(input shareddto.ChangeIssueStatusRequest) (any, error) {
			return corecommands.ChangeIssueStatus(ctx, h.core, input.ID, input.Status, input.Note)
		}), true
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
		}), true
	case protocol.MethodIssueClearTodo:
		return handle(req, func(input shareddto.NumericIDRequest) (any, error) {
			return corecommands.ClearIssueTodoForDate(ctx, h.core, input.ID)
		}), true
	case protocol.MethodIssueDailySummary:
		return handle(req, func(input shareddto.DailyIssueSummaryQuery) (any, error) {
			if input.Date != nil && *input.Date != "" {
				return corecommands.ComputeDailyIssueSummaryForDate(ctx, h.core, *input.Date)
			}
			return corecommands.ComputeDailyIssueSummaryForToday(ctx, h.core)
		}), true
	case protocol.MethodIssueTodaySummary:
		return h.handleNoParams(req, func() (any, error) {
			return corecommands.ComputeDailyIssueSummaryForToday(ctx, h.core)
		}), true
	case protocol.MethodDailyPlanGet:
		return handle(req, func(input shareddto.DailyPlanQuery) (any, error) {
			return corecommands.GetDailyPlan(ctx, h.core, input.Date)
		}), true
	case protocol.MethodHabitList:
		return handle(req, func(input shareddto.ListHabitsQuery) (any, error) {
			return corecommands.ListHabitsByStream(ctx, h.core, input.StreamID)
		}), true
	case protocol.MethodHabitListAll:
		return h.handleNoParams(req, func() (any, error) {
			return corecommands.ListAllHabits(ctx, h.core)
		}), true
	case protocol.MethodHabitListDue:
		return handle(req, func(input shareddto.ListHabitsDueQuery) (any, error) {
			return corecommands.ListHabitsDueForDate(ctx, h.core, input.Date)
		}), true
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
		}), true
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
		}), true
	case protocol.MethodHabitDelete:
		return handle(req, func(input shareddto.NumericIDRequest) (any, error) {
			return shareddto.OKResponse{OK: true}, corecommands.DeleteHabit(ctx, h.core, input.ID)
		}), true
	case protocol.MethodHabitComplete:
		return handle(req, func(input shareddto.HabitCompletionUpsertRequest) (any, error) {
			status := sharedtypes.HabitCompletionStatusCompleted
			if input.Status != nil {
				status = *input.Status
			}
			return corecommands.CompleteHabit(ctx, h.core, input.HabitID, input.Date, status, input.DurationMinutes, input.Notes)
		}), true
	case protocol.MethodHabitUncomplete:
		return handle(req, func(input shareddto.HabitCompletionUpsertRequest) (any, error) {
			return shareddto.OKResponse{OK: true}, corecommands.UncompleteHabit(ctx, h.core, input.HabitID, input.Date)
		}), true
	case protocol.MethodHabitHistory:
		return handle(req, func(input shareddto.HabitHistoryQuery) (any, error) {
			return corecommands.ListHabitHistory(ctx, h.core, input.RepoID, input.StreamID)
		}), true
	case protocol.MethodCheckInGet:
		return handle(req, func(input shareddto.DailyCheckInQuery) (any, error) {
			return corecommands.GetDailyCheckIn(ctx, h.core, input.Date)
		}), true
	case protocol.MethodCheckInUpsert:
		return handle(req, func(input shareddto.DailyCheckInUpsertRequest) (any, error) {
			return corecommands.UpsertDailyCheckIn(ctx, h.core, input)
		}), true
	case protocol.MethodCheckInDelete:
		return handle(req, func(input shareddto.DeleteByDateRequest) (any, error) {
			return shareddto.OKResponse{OK: true}, corecommands.DeleteDailyCheckIn(ctx, h.core, input.Date)
		}), true
	case protocol.MethodCheckInRange:
		return handle(req, func(input shareddto.DateRangeQuery) (any, error) {
			return corecommands.ListDailyCheckInsInRange(ctx, h.core, input.Start, input.End)
		}), true
	case protocol.MethodMetricsRange:
		return handle(req, func(input shareddto.DateRangeQuery) (any, error) {
			return corecommands.ComputeMetricsRange(ctx, h.core, input.Start, input.End)
		}), true
	case protocol.MethodMetricsRollup:
		return handle(req, func(input shareddto.DateRangeQuery) (any, error) {
			return corecommands.ComputeMetricsRollup(ctx, h.core, input.Start, input.End)
		}), true
	case protocol.MethodMetricsStreaks:
		return handle(req, func(input shareddto.DateRangeQuery) (any, error) {
			return corecommands.ComputeMetricsStreaks(ctx, h.core, input.Start, input.End)
		}), true
	case protocol.MethodDashboardWindow:
		return handle(req, func(input shareddto.DashboardWindowQuery) (any, error) {
			return corecommands.ComputeDashboardWindowSummary(ctx, h.core, input)
		}), true
	case protocol.MethodDashboardFocusScore:
		return handle(req, func(input shareddto.DashboardSummaryQuery) (any, error) {
			return corecommands.ComputeFocusScoreSummary(ctx, h.core, input)
		}), true
	case protocol.MethodDashboardDistribution:
		return handle(req, func(input shareddto.DashboardSummaryQuery) (any, error) {
			return corecommands.ComputeTimeDistributionSummary(ctx, h.core, input)
		}), true
	case protocol.MethodDashboardGoalProgress:
		return handle(req, func(input shareddto.DashboardSummaryQuery) (any, error) {
			return corecommands.ComputeGoalProgressSummary(ctx, h.core, input)
		}), true
	case protocol.MethodExportAssetsGet:
		return h.handleNoParams(req, func() (any, error) {
			return export.EnsureAssets(h.paths)
		}), true
	case protocol.MethodExportReportsDirSet:
		return handle(req, func(input shareddto.ExportReportsDirUpdateRequest) (any, error) {
			return export.SetReportsDir(h.paths, input.ReportsDir)
		}), true
	case protocol.MethodExportICSDirSet:
		return handle(req, func(input shareddto.ExportICSDirUpdateRequest) (any, error) {
			return export.SetICSDir(h.paths, input.ICSDir)
		}), true
	case protocol.MethodExportReportsList:
		return h.handleNoParams(req, func() (any, error) {
			return export.ListReports(h.paths)
		}), true
	case protocol.MethodExportReportsDelete:
		return handle(req, func(input shareddto.ExportReportDeleteRequest) (any, error) {
			if err := export.DeleteReport(h.paths, input.Path); err != nil {
				return nil, err
			}
			return shareddto.OKResponse{OK: true}, nil
		}), true
	case protocol.MethodExportTemplateReset:
		return handle(req, func(input shareddto.ExportTemplateResetRequest) (any, error) {
			return export.ResetTemplate(h.paths, input.ReportKind, input.AssetKind)
		}), true
	case protocol.MethodExportTemplateApply:
		return handle(req, func(input shareddto.ExportTemplatePresetApplyRequest) (any, error) {
			return export.ApplyTemplatePreset(h.paths, input.ReportKind, input.AssetKind, input.PresetID)
		}), true
	case protocol.MethodExportDaily:
		return handle(req, func(input shareddto.DailyReportRequest) (any, error) {
			input.Kind = sharedtypes.ExportReportKindDaily
			return export.GenerateReport(ctx, h.core, h.paths, input)
		}), true
	case protocol.MethodExportWeekly:
		return handle(req, func(input shareddto.ExportReportRequest) (any, error) {
			input.Kind = sharedtypes.ExportReportKindWeekly
			return export.GenerateReport(ctx, h.core, h.paths, input)
		}), true
	case protocol.MethodExportRepo:
		return handle(req, func(input shareddto.ExportReportRequest) (any, error) {
			input.Kind = sharedtypes.ExportReportKindRepo
			return export.GenerateReport(ctx, h.core, h.paths, input)
		}), true
	case protocol.MethodExportStream:
		return handle(req, func(input shareddto.ExportReportRequest) (any, error) {
			input.Kind = sharedtypes.ExportReportKindStream
			return export.GenerateReport(ctx, h.core, h.paths, input)
		}), true
	case protocol.MethodExportIssueRollup:
		return handle(req, func(input shareddto.ExportReportRequest) (any, error) {
			input.Kind = sharedtypes.ExportReportKindIssueRollup
			return export.GenerateReport(ctx, h.core, h.paths, input)
		}), true
	case protocol.MethodExportCSV:
		return handle(req, func(input shareddto.ExportReportRequest) (any, error) {
			input.Kind = sharedtypes.ExportReportKindCSV
			return export.GenerateReport(ctx, h.core, h.paths, input)
		}), true
	case protocol.MethodExportCalendar:
		return handle(req, func(input shareddto.ExportCalendarRequest) (any, error) {
			return export.GenerateCalendarExport(ctx, h.core, h.paths, input)
		}), true
	default:
		return protocol.Response{}, false
	}
}
