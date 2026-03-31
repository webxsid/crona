package dto

import "crona/shared/types"

// Shared request DTOs used across kernel, TUI, and future CLI clients.

type Empty struct{}

type OKResponse struct {
	OK bool `json:"ok"`
}

type ErrorResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error"`
}

type CreateRepoRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Color       *string `json:"color,omitempty"`
}

type UpdateRepoRequest struct {
	ID          string  `json:"id"`
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Color       *string `json:"color,omitempty"`
}

type ListStreamsQuery struct {
	RepoID int64 `json:"repoId"`
}

type CreateStreamRequest struct {
	RepoID      int64                   `json:"repoId"`
	Name        string                  `json:"name"`
	Description *string                 `json:"description,omitempty"`
	Visibility  *types.StreamVisibility `json:"visibility,omitempty"`
}

type UpdateStreamRequest struct {
	ID          int64                   `json:"id"`
	Name        *string                 `json:"name,omitempty"`
	Description *string                 `json:"description,omitempty"`
	Visibility  *types.StreamVisibility `json:"visibility,omitempty"`
}

type ListIssuesQuery struct {
	StreamID int64 `json:"streamId"`
}

type ListHabitsQuery struct {
	StreamID int64 `json:"streamId"`
}

type ListHabitsDueQuery struct {
	Date string `json:"date"`
}

type CreateHabitRequest struct {
	StreamID      int64   `json:"streamId"`
	Name          string  `json:"name"`
	Description   *string `json:"description,omitempty"`
	ScheduleType  string  `json:"scheduleType"`
	Weekdays      []int   `json:"weekdays,omitempty"`
	TargetMinutes *int    `json:"targetMinutes,omitempty"`
}

type UpdateHabitRequest struct {
	ID            int64   `json:"id"`
	Name          *string `json:"name,omitempty"`
	Description   *string `json:"description,omitempty"`
	ScheduleType  *string `json:"scheduleType,omitempty"`
	Weekdays      []int   `json:"weekdays,omitempty"`
	TargetMinutes *int    `json:"targetMinutes,omitempty"`
	Active        *bool   `json:"active,omitempty"`
}

type HabitCompletionUpsertRequest struct {
	HabitID         int64                        `json:"habitId"`
	Date            string                       `json:"date"`
	Status          *types.HabitCompletionStatus `json:"status,omitempty"`
	DurationMinutes *int                         `json:"durationMinutes,omitempty"`
	Notes           *string                      `json:"notes,omitempty"`
}

type HabitHistoryQuery struct {
	HabitID int64 `json:"habitId"`
}

type DailyIssueSummaryQuery struct {
	Date *string `json:"date,omitempty"`
}

type DailyPlanQuery struct {
	Date string `json:"date"`
}

type DailyCheckInQuery struct {
	Date string `json:"date"`
}

type DailyCheckInUpsertRequest struct {
	Date              string   `json:"date"`
	Mood              int      `json:"mood"`
	Energy            int      `json:"energy"`
	SleepHours        *float64 `json:"sleepHours,omitempty"`
	SleepScore        *int     `json:"sleepScore,omitempty"`
	ScreenTimeMinutes *int     `json:"screenTimeMinutes,omitempty"`
	Notes             *string  `json:"notes,omitempty"`
}

type DeleteByDateRequest struct {
	Date string `json:"date"`
}

type DateRangeQuery struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type ExportReportRequest struct {
	Kind       types.ExportReportKind `json:"kind,omitempty"`
	Date       string                 `json:"date,omitempty"`
	Start      string                 `json:"start,omitempty"`
	End        string                 `json:"end,omitempty"`
	RepoID     *int64                 `json:"repoId,omitempty"`
	StreamID   *int64                 `json:"streamId,omitempty"`
	Format     types.ExportFormat     `json:"format,omitempty"`
	OutputMode types.ExportOutputMode `json:"outputMode"`
	PresetID   string                 `json:"presetId,omitempty"`
}

type DailyReportRequest = ExportReportRequest

type ExportCalendarRequest struct {
	RepoID int64 `json:"repoId"`
}

type ExportReportsDirUpdateRequest struct {
	ReportsDir string `json:"reportsDir"`
}

type ExportICSDirUpdateRequest struct {
	ICSDir string `json:"icsDir"`
}

type ExportReportDeleteRequest struct {
	Path string `json:"path"`
}

type ExportTemplateResetRequest struct {
	ReportKind types.ExportReportKind `json:"reportKind,omitempty"`
	AssetKind  types.ExportAssetKind  `json:"assetKind,omitempty"`
}

type ExportTemplatePresetApplyRequest struct {
	ReportKind types.ExportReportKind `json:"reportKind,omitempty"`
	AssetKind  types.ExportAssetKind  `json:"assetKind,omitempty"`
	PresetID   string                 `json:"presetId,omitempty"`
}

type CreateIssueRequest struct {
	StreamID        int64   `json:"streamId"`
	Title           string  `json:"title"`
	Description     *string `json:"description,omitempty"`
	EstimateMinutes *int    `json:"estimateMinutes,omitempty"`
	Notes           *string `json:"notes,omitempty"`
	TodoForDate     *string `json:"todoForDate,omitempty"`
}

type UpdateIssueRequest struct {
	ID              int64   `json:"id"`
	Title           *string `json:"title,omitempty"`
	Description     *string `json:"description,omitempty"`
	EstimateMinutes *int    `json:"estimateMinutes,omitempty"`
	Notes           *string `json:"notes,omitempty"`
}

type ChangeIssueStatusRequest struct {
	ID     int64             `json:"id"`
	Status types.IssueStatus `json:"status"`
	Note   *string           `json:"note,omitempty"`
}

type SetIssueTodoRequest struct {
	ID   int64   `json:"id"`
	Date *string `json:"date,omitempty"`
}

type ListSessionsQuery struct {
	IssueID *int64 `json:"issueId,omitempty"`
}

type SessionIDRequest struct {
	ID string `json:"id"`
}

type StartSessionRequest struct {
	IssueID int64 `json:"issueId"`
}

type EndSessionRequest struct {
	CommitMessage *string `json:"commitMessage,omitempty"`
	WorkedOn      *string `json:"workedOn,omitempty"`
	Outcome       *string `json:"outcome,omitempty"`
	NextStep      *string `json:"nextStep,omitempty"`
	Blockers      *string `json:"blockers,omitempty"`
	Links         *string `json:"links,omitempty"`
}

type ManualSessionLogRequest struct {
	IssueID              int64   `json:"issueId"`
	Date                 string  `json:"date"`
	WorkDurationSeconds  int     `json:"workDurationSeconds"`
	BreakDurationSeconds int     `json:"breakDurationSeconds"`
	StartTime            *string `json:"startTime,omitempty"`
	EndTime              *string `json:"endTime,omitempty"`
	CommitMessage        *string `json:"commitMessage,omitempty"`
	Notes                *string `json:"notes,omitempty"`
}

type AmendSessionNoteRequest struct {
	ID   *string `json:"id,omitempty"`
	Note string  `json:"note"`
}

type SessionHistoryQuery struct {
	RepoID   *int64  `json:"repoId,omitempty"`
	StreamID *int64  `json:"streamId,omitempty"`
	IssueID  *int64  `json:"issueId,omitempty"`
	Since    *string `json:"since,omitempty"`
	Until    *string `json:"until,omitempty"`
	Limit    *int    `json:"limit,omitempty"`
	Offset   *int    `json:"offset,omitempty"`
	Context  *bool   `json:"context,omitempty"`
}

type ListOpsQuery struct {
	Entity   *types.OpEntity `json:"entity,omitempty"`
	EntityID *string         `json:"entityId,omitempty"`
	Limit    *int            `json:"limit,omitempty"`
}

type ListOpsSinceQuery struct {
	Since string `json:"since"`
}

type ListLatestOpsQuery struct {
	Limit *int `json:"limit,omitempty"`
}

type PatchCoreSettingRequest struct {
	Key   types.CoreSettingsKey `json:"key"`
	Value any                   `json:"value"`
}

type GetCoreSettingRequest struct {
	Key types.CoreSettingsKey `json:"key"`
}

type PutCoreSettingsRequest map[types.CoreSettingsKey]any

type UpdateContextRequest struct {
	RepoID   *int64 `json:"repoId,omitempty"`
	StreamID *int64 `json:"streamId,omitempty"`
	IssueID  *int64 `json:"issueId,omitempty"`
}

type SwitchRepoRequest struct {
	RepoID int64 `json:"repoId"`
}

type SwitchStreamRequest struct {
	StreamID int64 `json:"streamId"`
}

type SwitchIssueRequest struct {
	IssueID int64 `json:"issueId"`
}

type ListScratchpadsQuery struct {
	PinnedOnly *bool `json:"pinnedOnly,omitempty"`
}

type RegisterScratchpadRequest struct {
	ID           *string `json:"id,omitempty"`
	Path         string  `json:"path"`
	Name         string  `json:"name"`
	LastOpenedAt *string `json:"lastOpenedAt,omitempty"`
	Pinned       *bool   `json:"pinned,omitempty"`
}

type PinScratchpadRequest struct {
	ID     string `json:"id"`
	Pinned bool   `json:"pinned"`
}

type ScratchpadIDRequest struct {
	ID string `json:"id"`
}

type CreateStashRequest struct {
	StashNote *string `json:"stashNote,omitempty"`
}

type NumericIDRequest struct {
	ID int64 `json:"id"`
}

type StashIDRequest struct {
	ID string `json:"id"`
}

type TimerStartRequest struct {
	IssueID *int64 `json:"issueId,omitempty"`
}
