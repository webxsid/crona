package types

// Shared domain and wire types used across the Go workspace.

type IssueStatus string

const (
	IssueStatusBacklog    IssueStatus = "backlog"
	IssueStatusPlanned    IssueStatus = "planned"
	IssueStatusReady      IssueStatus = "ready"
	IssueStatusInProgress IssueStatus = "in_progress"
	IssueStatusBlocked    IssueStatus = "blocked"
	IssueStatusInReview   IssueStatus = "in_review"
	IssueStatusDone       IssueStatus = "done"
	IssueStatusAbandoned  IssueStatus = "abandoned"
)

type StreamVisibility string

const (
	StreamVisibilityPersonal StreamVisibility = "personal"
	StreamVisibilityShared   StreamVisibility = "shared"
)

type SessionSegmentType string

const (
	SessionSegmentWork       SessionSegmentType = "work"
	SessionSegmentShortBreak SessionSegmentType = "short_break"
	SessionSegmentLongBreak  SessionSegmentType = "long_break"
	SessionSegmentRest       SessionSegmentType = "rest"
)

type TimerMode string

const (
	TimerModeStopwatch  TimerMode = "stopwatch"
	TimerModeStructured TimerMode = "structured"
)

type UpdateChannel string

const (
	UpdateChannelStable UpdateChannel = "stable"
	UpdateChannelBeta   UpdateChannel = "beta"
)

func NormalizeUpdateChannel(value UpdateChannel) UpdateChannel {
	switch value {
	case UpdateChannelBeta:
		return value
	default:
		return UpdateChannelStable
	}
}

type AlertUrgency string

const (
	AlertUrgencyLow    AlertUrgency = "low"
	AlertUrgencyNormal AlertUrgency = "normal"
	AlertUrgencyHigh   AlertUrgency = "high"
)

func NormalizeAlertUrgency(value AlertUrgency) AlertUrgency {
	switch value {
	case AlertUrgencyLow, AlertUrgencyHigh:
		return value
	default:
		return AlertUrgencyNormal
	}
}

type AlertSoundPreset string

const (
	AlertSoundPresetChime        AlertSoundPreset = "chime"
	AlertSoundPresetSoftBell     AlertSoundPreset = "soft_bell"
	AlertSoundPresetFocusGong    AlertSoundPreset = "focus_gong"
	AlertSoundPresetMinimalClick AlertSoundPreset = "minimal_click"
)

func NormalizeAlertSoundPreset(value AlertSoundPreset) AlertSoundPreset {
	switch value {
	case AlertSoundPresetSoftBell, AlertSoundPresetFocusGong, AlertSoundPresetMinimalClick:
		return value
	default:
		return AlertSoundPresetChime
	}
}

func AvailableAlertSoundPresets() []AlertSoundPreset {
	return []AlertSoundPreset{
		AlertSoundPresetChime,
		AlertSoundPresetSoftBell,
		AlertSoundPresetFocusGong,
		AlertSoundPresetMinimalClick,
	}
}

type AlertEventKind string

const (
	AlertEventTimerWorkComplete  AlertEventKind = "timer.work_complete"
	AlertEventTimerBreakComplete AlertEventKind = "timer.break_complete"
	AlertEventUpdateAvailable    AlertEventKind = "update.available"
	AlertEventUpdateInstalled    AlertEventKind = "update.installed"
	AlertEventSupportBundleReady AlertEventKind = "support.bundle_ready"
	AlertEventExportCompleted    AlertEventKind = "export.completed"
	AlertEventRuntimeDegraded    AlertEventKind = "runtime.degraded"
	AlertEventCheckInReminder    AlertEventKind = "checkin.reminder"
	AlertEventTestNotification   AlertEventKind = "test.notification"
	AlertEventTestSound          AlertEventKind = "test.sound"
)

type AlertReminderKind string

const (
	AlertReminderKindCheckIn AlertReminderKind = "checkin_reminder"
)

func NormalizeAlertReminderKind(value AlertReminderKind) AlertReminderKind {
	switch value {
	case AlertReminderKindCheckIn:
		return value
	default:
		return AlertReminderKindCheckIn
	}
}

type AlertReminderScheduleType string

const (
	AlertReminderScheduleDaily  AlertReminderScheduleType = "daily"
	AlertReminderScheduleWeekly AlertReminderScheduleType = "weekly"
)

func NormalizeAlertReminderScheduleType(value AlertReminderScheduleType) AlertReminderScheduleType {
	switch value {
	case AlertReminderScheduleWeekly:
		return value
	default:
		return AlertReminderScheduleDaily
	}
}

type RepoSort string

const (
	RepoSortChronologicalAsc  RepoSort = "chronological_asc"
	RepoSortChronologicalDesc RepoSort = "chronological_desc"
	RepoSortAlphabeticalAsc   RepoSort = "alphabetical_asc"
	RepoSortAlphabeticalDesc  RepoSort = "alphabetical_desc"
)

func NormalizeRepoSort(value RepoSort) RepoSort {
	switch value {
	case RepoSortChronologicalDesc, RepoSortAlphabeticalAsc, RepoSortAlphabeticalDesc:
		return value
	default:
		return RepoSortChronologicalAsc
	}
}

type StreamSort string

const (
	StreamSortChronologicalAsc  StreamSort = "chronological_asc"
	StreamSortChronologicalDesc StreamSort = "chronological_desc"
	StreamSortAlphabeticalAsc   StreamSort = "alphabetical_asc"
	StreamSortAlphabeticalDesc  StreamSort = "alphabetical_desc"
)

func NormalizeStreamSort(value StreamSort) StreamSort {
	switch value {
	case StreamSortChronologicalDesc, StreamSortAlphabeticalAsc, StreamSortAlphabeticalDesc:
		return value
	default:
		return StreamSortChronologicalAsc
	}
}

type IssueSort string

const (
	IssueSortPriority          IssueSort = "priority"
	IssueSortDueDateAsc        IssueSort = "due_date_asc"
	IssueSortDueDateDesc       IssueSort = "due_date_desc"
	IssueSortChronologicalAsc  IssueSort = "chronological_asc"
	IssueSortChronologicalDesc IssueSort = "chronological_desc"
	IssueSortAlphabeticalAsc   IssueSort = "alphabetical_asc"
	IssueSortAlphabeticalDesc  IssueSort = "alphabetical_desc"
)

func NormalizeIssueSort(value IssueSort) IssueSort {
	switch value {
	case IssueSortDueDateAsc, IssueSortDueDateDesc, IssueSortChronologicalAsc, IssueSortChronologicalDesc, IssueSortAlphabeticalAsc, IssueSortAlphabeticalDesc:
		return value
	default:
		return IssueSortPriority
	}
}

type HabitSort string

const (
	HabitSortSchedule          HabitSort = "schedule"
	HabitSortTargetMinutesAsc  HabitSort = "target_minutes_asc"
	HabitSortTargetMinutesDesc HabitSort = "target_minutes_desc"
	HabitSortChronologicalAsc  HabitSort = "chronological_asc"
	HabitSortChronologicalDesc HabitSort = "chronological_desc"
	HabitSortAlphabeticalAsc   HabitSort = "alphabetical_asc"
	HabitSortAlphabeticalDesc  HabitSort = "alphabetical_desc"
)

func NormalizeHabitSort(value HabitSort) HabitSort {
	switch value {
	case HabitSortTargetMinutesAsc, HabitSortTargetMinutesDesc, HabitSortChronologicalAsc, HabitSortChronologicalDesc, HabitSortAlphabeticalAsc, HabitSortAlphabeticalDesc:
		return value
	default:
		return HabitSortSchedule
	}
}

type OpEntity string

const (
	OpEntityRepo            OpEntity = "repo"
	OpEntityStream          OpEntity = "stream"
	OpEntityIssue           OpEntity = "issue"
	OpEntityHabit           OpEntity = "habit"
	OpEntityHabitCompletion OpEntity = "habit_completion"
	OpEntityCheckIn         OpEntity = "checkin"
	OpEntitySession         OpEntity = "session"
	OpEntitySessionSegment  OpEntity = "session_segment"
	OpEntityActiveContext   OpEntity = "active_context"
	OpEntityStash           OpEntity = "stash"
)

type OpAction string

const (
	OpActionCreate  OpAction = "create"
	OpActionUpdate  OpAction = "update"
	OpActionDelete  OpAction = "delete"
	OpActionRestore OpAction = "restore"
)

type CoreSettingsKey string

const (
	CoreSettingsKeyTimerMode             CoreSettingsKey = "timerMode"
	CoreSettingsKeyBreaksEnabled         CoreSettingsKey = "breaksEnabled"
	CoreSettingsKeyWorkDurationMinutes   CoreSettingsKey = "workDurationMinutes"
	CoreSettingsKeyShortBreakMinutes     CoreSettingsKey = "shortBreakMinutes"
	CoreSettingsKeyLongBreakMinutes      CoreSettingsKey = "longBreakMinutes"
	CoreSettingsKeyLongBreakEnabled      CoreSettingsKey = "longBreakEnabled"
	CoreSettingsKeyCyclesBeforeLongBreak CoreSettingsKey = "cyclesBeforeLongBreak"
	CoreSettingsKeyAutoStartBreaks       CoreSettingsKey = "autoStartBreaks"
	CoreSettingsKeyAutoStartWork         CoreSettingsKey = "autoStartWork"
	CoreSettingsKeyBoundaryNotifications CoreSettingsKey = "boundaryNotificationsEnabled"
	CoreSettingsKeyBoundarySound         CoreSettingsKey = "boundarySoundEnabled"
	CoreSettingsKeyAlertSoundPreset      CoreSettingsKey = "alertSoundPreset"
	CoreSettingsKeyAlertUrgency          CoreSettingsKey = "alertUrgency"
	CoreSettingsKeyAlertIconEnabled      CoreSettingsKey = "alertIconEnabled"
	CoreSettingsKeyUpdateChecksEnabled   CoreSettingsKey = "updateChecksEnabled"
	CoreSettingsKeyUpdatePromptEnabled   CoreSettingsKey = "updatePromptEnabled"
	CoreSettingsKeyUpdateChannel         CoreSettingsKey = "updateChannel"
	CoreSettingsKeyRepoSort              CoreSettingsKey = "repoSort"
	CoreSettingsKeyStreamSort            CoreSettingsKey = "streamSort"
	CoreSettingsKeyIssueSort             CoreSettingsKey = "issueSort"
	CoreSettingsKeyHabitSort             CoreSettingsKey = "habitSort"
	CoreSettingsKeyAwayModeEnabled       CoreSettingsKey = "awayModeEnabled"
	CoreSettingsKeyFrozenStreakKinds     CoreSettingsKey = "frozenStreakKinds"
	CoreSettingsKeyRestWeekdays          CoreSettingsKey = "restWeekdays"
	CoreSettingsKeyRestSpecificDates     CoreSettingsKey = "restSpecificDates"
	CoreSettingsKeyDailyPlanRollbackMins CoreSettingsKey = "dailyPlanRollbackMinutes"
)

type StreakKind string

const (
	StreakKindFocusDays   StreakKind = "focus_days"
	StreakKindCheckInDays StreakKind = "checkin_days"
)

func NormalizeStreakKind(value StreakKind) StreakKind {
	switch value {
	case StreakKindCheckInDays:
		return value
	default:
		return StreakKindFocusDays
	}
}

func AvailableStreakKinds() []StreakKind {
	return []StreakKind{
		StreakKindFocusDays,
		StreakKindCheckInDays,
	}
}

type SessionNoteSection string

const (
	SessionNoteSectionCommit  SessionNoteSection = "commit"
	SessionNoteSectionContext SessionNoteSection = "context"
	SessionNoteSectionWork    SessionNoteSection = "work"
	SessionNoteSectionNotes   SessionNoteSection = "notes"
)

type Repo struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Color       *string `json:"color,omitempty"`
}

type Stream struct {
	ID          int64            `json:"id"`
	RepoID      int64            `json:"repoId"`
	Name        string           `json:"name"`
	Description *string          `json:"description,omitempty"`
	Visibility  StreamVisibility `json:"visibility"`
}

type HabitScheduleType string

const (
	HabitScheduleDaily    HabitScheduleType = "daily"
	HabitScheduleWeekdays HabitScheduleType = "weekdays"
	HabitScheduleWeekly   HabitScheduleType = "weekly"
)

func NormalizeHabitScheduleType(value HabitScheduleType) HabitScheduleType {
	switch value {
	case HabitScheduleWeekdays, HabitScheduleWeekly:
		return value
	default:
		return HabitScheduleDaily
	}
}

type Habit struct {
	ID            int64             `json:"id"`
	StreamID      int64             `json:"streamId"`
	Name          string            `json:"name"`
	Description   *string           `json:"description,omitempty"`
	ScheduleType  HabitScheduleType `json:"scheduleType"`
	Weekdays      []int             `json:"weekdays,omitempty"`
	TargetMinutes *int              `json:"targetMinutes,omitempty"`
	Active        bool              `json:"active"`
}

type HabitCompletionStatus string

const (
	HabitCompletionStatusCompleted HabitCompletionStatus = "completed"
	HabitCompletionStatusFailed    HabitCompletionStatus = "failed"
)

func NormalizeHabitCompletionStatus(value HabitCompletionStatus) HabitCompletionStatus {
	switch value {
	case HabitCompletionStatusFailed:
		return value
	default:
		return HabitCompletionStatusCompleted
	}
}

type HabitCompletion struct {
	ID              int64                 `json:"id"`
	HabitID         int64                 `json:"habitId"`
	Date            string                `json:"date"`
	Status          HabitCompletionStatus `json:"status"`
	DurationMinutes *int                  `json:"durationMinutes,omitempty"`
	Notes           *string               `json:"notes,omitempty"`
	SnapshotName    *string               `json:"snapshotName,omitempty"`
	SnapshotDesc    *string               `json:"snapshotDescription,omitempty"`
	SnapshotType    *HabitScheduleType    `json:"snapshotScheduleType,omitempty"`
	SnapshotDays    []int                 `json:"snapshotWeekdays,omitempty"`
	SnapshotTarget  *int                  `json:"snapshotTargetMinutes,omitempty"`
	CreatedAt       string                `json:"createdAt"`
	UpdatedAt       string                `json:"updatedAt"`
}

type HabitWithMeta struct {
	Habit
	RepoID     int64  `json:"repoId"`
	RepoName   string `json:"repoName"`
	StreamName string `json:"streamName"`
}

type HabitDailyItem struct {
	HabitWithMeta
	Status          HabitCompletionStatus `json:"status"`
	Completed       bool                  `json:"completed"`
	CompletionID    *int64                `json:"completionId,omitempty"`
	CompletionDate  *string               `json:"completionDate,omitempty"`
	DurationMinutes *int                  `json:"durationMinutes,omitempty"`
	Notes           *string               `json:"notes,omitempty"`
	SnapshotName    *string               `json:"snapshotName,omitempty"`
	SnapshotDesc    *string               `json:"snapshotDescription,omitempty"`
	SnapshotType    *HabitScheduleType    `json:"snapshotScheduleType,omitempty"`
	SnapshotDays    []int                 `json:"snapshotWeekdays,omitempty"`
	SnapshotTarget  *int                  `json:"snapshotTargetMinutes,omitempty"`
}

type Issue struct {
	ID              int64       `json:"id"`
	StreamID        int64       `json:"streamId"`
	Title           string      `json:"title"`
	Description     *string     `json:"description,omitempty"`
	Status          IssueStatus `json:"status"`
	EstimateMinutes *int        `json:"estimateMinutes,omitempty"`
	Notes           *string     `json:"notes,omitempty"`
	TodoForDate     *string     `json:"todoForDate,omitempty"`
	CompletedAt     *string     `json:"completedAt,omitempty"`
	AbandonedAt     *string     `json:"abandonedAt,omitempty"`
}

type IssueWithMeta struct {
	Issue
	RepoID     int64  `json:"repoId"`
	RepoName   string `json:"repoName"`
	StreamName string `json:"streamName"`
}

type DailyIssueSummary struct {
	Date                  string  `json:"date"`
	TotalIssues           int     `json:"totalIssues"`
	Issues                []Issue `json:"issues"`
	TotalEstimatedMinutes int     `json:"totalEstimatedMinutes"`
	CompletedIssues       int     `json:"completedIssues"`
	AbandonedIssues       int     `json:"abandonedIssues"`
	WorkedSeconds         int     `json:"workedSeconds"`
}

type DailyPlanEntryStatus string

const (
	DailyPlanEntryStatusPlanned   DailyPlanEntryStatus = "planned"
	DailyPlanEntryStatusCompleted DailyPlanEntryStatus = "completed"
	DailyPlanEntryStatusFailed    DailyPlanEntryStatus = "failed"
	DailyPlanEntryStatusAbandoned DailyPlanEntryStatus = "abandoned"
	DailyPlanEntryStatusCanceled  DailyPlanEntryStatus = "canceled"
)

type DailyPlanFailureReason string

const (
	DailyPlanFailureReasonMoved     DailyPlanFailureReason = "moved"
	DailyPlanFailureReasonCleared   DailyPlanFailureReason = "cleared"
	DailyPlanFailureReasonAbandoned DailyPlanFailureReason = "abandoned"
	DailyPlanFailureReasonMissed    DailyPlanFailureReason = "missed"
)

type DailyPlanEventType string

const (
	DailyPlanEventCommitted       DailyPlanEventType = "committed"
	DailyPlanEventRescheduled     DailyPlanEventType = "rescheduled"
	DailyPlanEventCleared         DailyPlanEventType = "cleared"
	DailyPlanEventRestored        DailyPlanEventType = "restored"
	DailyPlanEventCompleted       DailyPlanEventType = "completed"
	DailyPlanEventAbandoned       DailyPlanEventType = "abandoned"
	DailyPlanEventFailed          DailyPlanEventType = "failed"
	DailyPlanEventFailureReverted DailyPlanEventType = "failure_reverted"
)

type DailyPlanEvent struct {
	ID            string                  `json:"id"`
	EntryID       string                  `json:"entryId"`
	Type          DailyPlanEventType      `json:"type"`
	FailureReason *DailyPlanFailureReason `json:"failureReason,omitempty"`
	Timestamp     string                  `json:"timestamp"`
}

type DailyPlanAccountabilitySummary struct {
	PlannedCount         int     `json:"plannedCount"`
	CompletedCount       int     `json:"completedCount"`
	FailedCount          int     `json:"failedCount"`
	AbandonedCount       int     `json:"abandonedCount"`
	PendingRollbackCount int     `json:"pendingRollbackCount"`
	AccountabilityScore  float64 `json:"accountabilityScore"`
	BacklogPressure      float64 `json:"backlogPressure"`
	DelayedIssueCount    int     `json:"delayedIssueCount"`
	HighRiskIssueCount   int     `json:"highRiskIssueCount"`
	AvgDelayDays         float64 `json:"avgDelayDays"`
	MaxDelayDays         int     `json:"maxDelayDays"`
}

type DailyPlanEntry struct {
	ID                   string                  `json:"id"`
	Date                 string                  `json:"date"`
	IssueID              int64                   `json:"issueId"`
	Source               string                  `json:"source"`
	Status               DailyPlanEntryStatus    `json:"status"`
	FailureReason        *DailyPlanFailureReason `json:"failureReason,omitempty"`
	PendingFailureReason *DailyPlanFailureReason `json:"pendingFailureReason,omitempty"`
	CommittedAt          string                  `json:"committedAt"`
	PendingFailureAt     *string                 `json:"pendingFailureAt,omitempty"`
	ResolvedAt           *string                 `json:"resolvedAt,omitempty"`
	BaselineDate         string                  `json:"baselineDate,omitempty"`
	CurrentPlannedDate   string                  `json:"currentPlannedDate,omitempty"`
	PostponeCount        int                     `json:"postponeCount"`
	CurrentDelayedDays   int                     `json:"currentDelayedDays"`
	MaxDelayedDays       int                     `json:"maxDelayedDays"`
	FailScore            float64                 `json:"failScore"`
	Events               []DailyPlanEvent        `json:"events,omitempty"`
}

type DailyPlan struct {
	ID        string                         `json:"id"`
	Date      string                         `json:"date"`
	CreatedAt string                         `json:"createdAt"`
	UpdatedAt string                         `json:"updatedAt"`
	Summary   DailyPlanAccountabilitySummary `json:"summary"`
	Entries   []DailyPlanEntry               `json:"entries,omitempty"`
}

type BurnoutLevel string

const (
	BurnoutLevelLow     BurnoutLevel = "low"
	BurnoutLevelGuarded BurnoutLevel = "guarded"
	BurnoutLevelHigh    BurnoutLevel = "high"
)

type DailyCheckIn struct {
	Date              string   `json:"date"`
	Mood              int      `json:"mood"`
	Energy            int      `json:"energy"`
	SleepHours        *float64 `json:"sleepHours,omitempty"`
	SleepScore        *int     `json:"sleepScore,omitempty"`
	ScreenTimeMinutes *int     `json:"screenTimeMinutes,omitempty"`
	Notes             *string  `json:"notes,omitempty"`
	CreatedAt         string   `json:"createdAt"`
	UpdatedAt         string   `json:"updatedAt"`
}

type BurnoutIndicator struct {
	Score   int                `json:"score"`
	Level   BurnoutLevel       `json:"level"`
	Factors map[string]float64 `json:"factors,omitempty"`
}

type DailyMetricsDay struct {
	Date                  string            `json:"date"`
	WorkedSeconds         int               `json:"workedSeconds"`
	RestSeconds           int               `json:"restSeconds"`
	SessionCount          int               `json:"sessionCount"`
	TotalIssues           int               `json:"totalIssues"`
	CompletedIssues       int               `json:"completedIssues"`
	AbandonedIssues       int               `json:"abandonedIssues"`
	TotalEstimatedMinutes int               `json:"totalEstimatedMinutes"`
	CheckIn               *DailyCheckIn     `json:"checkIn,omitempty"`
	Burnout               *BurnoutIndicator `json:"burnout,omitempty"`
}

type MetricsRollup struct {
	StartDate             string            `json:"startDate"`
	EndDate               string            `json:"endDate"`
	Days                  int               `json:"days"`
	CheckInDays           int               `json:"checkInDays"`
	FocusDays             int               `json:"focusDays"`
	WorkedSeconds         int               `json:"workedSeconds"`
	RestSeconds           int               `json:"restSeconds"`
	SessionCount          int               `json:"sessionCount"`
	CompletedIssues       int               `json:"completedIssues"`
	AbandonedIssues       int               `json:"abandonedIssues"`
	TotalEstimatedMinutes int               `json:"totalEstimatedMinutes"`
	AverageMood           *float64          `json:"averageMood,omitempty"`
	AverageEnergy         *float64          `json:"averageEnergy,omitempty"`
	AverageSleepHours     *float64          `json:"averageSleepHours,omitempty"`
	AverageSleepScore     *float64          `json:"averageSleepScore,omitempty"`
	AverageScreenTimeMins *float64          `json:"averageScreenTimeMinutes,omitempty"`
	LatestBurnout         *BurnoutIndicator `json:"latestBurnout,omitempty"`
}

type StreakSummary struct {
	CurrentFocusDays   int `json:"currentFocusDays"`
	LongestFocusDays   int `json:"longestFocusDays"`
	CurrentCheckInDays int `json:"currentCheckInDays"`
	LongestCheckInDays int `json:"longestCheckInDays"`
}

type DashboardWindowDayStatus string

const (
	DashboardWindowDayEmpty     DashboardWindowDayStatus = "empty"
	DashboardWindowDayPlanned   DashboardWindowDayStatus = "planned"
	DashboardWindowDayDone      DashboardWindowDayStatus = "done"
	DashboardWindowDayMissed    DashboardWindowDayStatus = "missed"
	DashboardWindowDayCarryOver DashboardWindowDayStatus = "carry_over"
	DashboardWindowDayMixed     DashboardWindowDayStatus = "mixed"
)

type DashboardWindowDay struct {
	Date                string                   `json:"date"`
	Status              DashboardWindowDayStatus `json:"status"`
	PlannedCount        int                      `json:"plannedCount"`
	CompletedCount      int                      `json:"completedCount"`
	FailedCount         int                      `json:"failedCount"`
	CarryOverCount      int                      `json:"carryOverCount"`
	AccountabilityScore float64                  `json:"accountabilityScore"`
}

type DashboardWindowSummary struct {
	StartDate           string               `json:"startDate"`
	EndDate             string               `json:"endDate"`
	PlannedCount        int                  `json:"plannedCount"`
	CompletedCount      int                  `json:"completedCount"`
	FailedCount         int                  `json:"failedCount"`
	AbandonedCount      int                  `json:"abandonedCount"`
	MissedCount         int                  `json:"missedCount"`
	CarryOverCount      int                  `json:"carryOverCount"`
	AccountabilityScore float64              `json:"accountabilityScore"`
	Days                []DashboardWindowDay `json:"days,omitempty"`
}

type FocusScoreLevel string

const (
	FocusScoreLevelLow          FocusScoreLevel = "low"
	FocusScoreLevelSteady       FocusScoreLevel = "steady"
	FocusScoreLevelStrong       FocusScoreLevel = "strong"
	FocusScoreLevelOverextended FocusScoreLevel = "overextended"
)

type FocusScoreSummary struct {
	StartDate           string          `json:"startDate"`
	EndDate             string          `json:"endDate"`
	Score               int             `json:"score"`
	Level               FocusScoreLevel `json:"level"`
	WorkedSeconds       int             `json:"workedSeconds"`
	RestSeconds         int             `json:"restSeconds"`
	SessionCount        int             `json:"sessionCount"`
	FocusDays           int             `json:"focusDays"`
	Days                int             `json:"days"`
	TargetWorkedSeconds int             `json:"targetWorkedSeconds"`
}

type DistributionGroup string

const (
	DistributionGroupRepo        DistributionGroup = "repo"
	DistributionGroupStream      DistributionGroup = "stream"
	DistributionGroupIssue       DistributionGroup = "issue"
	DistributionGroupSegmentType DistributionGroup = "segment_type"
)

type TimeDistributionRow struct {
	Key           string  `json:"key"`
	Label         string  `json:"label"`
	WorkedSeconds int     `json:"workedSeconds"`
	Percent       float64 `json:"percent"`
}

type TimeDistributionSummary struct {
	StartDate    string                `json:"startDate"`
	EndDate      string                `json:"endDate"`
	GroupBy      DistributionGroup     `json:"groupBy"`
	TotalSeconds int                   `json:"totalSeconds"`
	Rows         []TimeDistributionRow `json:"rows,omitempty"`
}

type GoalProgressGroup string

const (
	GoalProgressGroupRepo   GoalProgressGroup = "repo"
	GoalProgressGroupStream GoalProgressGroup = "stream"
	GoalProgressGroupIssue  GoalProgressGroup = "issue"
)

type GoalProgressStatus string

const (
	GoalProgressStatusUnestimated GoalProgressStatus = "unestimated"
	GoalProgressStatusOnTrack     GoalProgressStatus = "on_track"
	GoalProgressStatusAtRisk      GoalProgressStatus = "at_risk"
	GoalProgressStatusOver        GoalProgressStatus = "over"
)

type GoalProgressRow struct {
	Key             string             `json:"key"`
	Label           string             `json:"label"`
	EstimateMinutes int                `json:"estimateMinutes"`
	ActualSeconds   int                `json:"actualSeconds"`
	ProgressPercent float64            `json:"progressPercent"`
	Status          GoalProgressStatus `json:"status"`
}

type GoalProgressSummary struct {
	StartDate            string            `json:"startDate"`
	EndDate              string            `json:"endDate"`
	GroupBy              GoalProgressGroup `json:"groupBy"`
	TotalEstimateMinutes int               `json:"totalEstimateMinutes"`
	TotalActualSeconds   int               `json:"totalActualSeconds"`
	EstimatedItems       int               `json:"estimatedItems"`
	AverageDeltaMinutes  float64           `json:"averageDeltaMinutes"`
	AverageDeltaPercent  float64           `json:"averageDeltaPercent"`
	EstimateBias         string            `json:"estimateBias"`
	Rows                 []GoalProgressRow `json:"rows,omitempty"`
}

type SessionSource string

const (
	SessionSourceTracked SessionSource = "tracked"
	SessionSourceManual  SessionSource = "manual"
)

type Session struct {
	ID              string        `json:"id"`
	IssueID         int64         `json:"issueId"`
	Source          SessionSource `json:"source"`
	StartTime       string        `json:"startTime"`
	EndTime         *string       `json:"endTime,omitempty"`
	DurationSeconds *int          `json:"durationSeconds,omitempty"`
	Notes           *string       `json:"notes,omitempty"`
}

type ParsedSessionNotes map[SessionNoteSection]string

type SessionHistoryEntry struct {
	Session
	ParsedNotes ParsedSessionNotes `json:"parsedNotes,omitempty"`
}

type SessionDetail struct {
	SessionHistoryEntry
	RepoID      int64              `json:"repoId"`
	RepoName    string             `json:"repoName"`
	StreamID    int64              `json:"streamId"`
	StreamName  string             `json:"streamName"`
	IssueTitle  string             `json:"issueTitle"`
	WorkSummary SessionWorkSummary `json:"workSummary"`
}

type SessionWorkSummary struct {
	WorkSeconds  int `json:"workSeconds"`
	RestSeconds  int `json:"restSeconds"`
	WorkSegments int `json:"workSegments"`
	RestSegments int `json:"restSegments"`
	TotalSeconds int `json:"totalSeconds"`
}

type SessionSegment struct {
	ID                   string             `json:"id"`
	UserID               string             `json:"userId"`
	DeviceID             string             `json:"deviceId"`
	SessionID            string             `json:"sessionId"`
	SegmentType          SessionSegmentType `json:"segmentType"`
	StartTime            string             `json:"startTime"`
	EndTime              *string            `json:"endTime,omitempty"`
	ElapsedOffsetSeconds *int               `json:"elapsedOffsetSeconds,omitempty"`
	CreatedAt            string             `json:"createdAt"`
}

type ActiveContext struct {
	UserID     string  `json:"userId"`
	DeviceID   string  `json:"deviceId"`
	RepoID     *int64  `json:"repoId,omitempty"`
	RepoName   *string `json:"repoName,omitempty"`
	StreamID   *int64  `json:"streamId,omitempty"`
	StreamName *string `json:"streamName,omitempty"`
	IssueID    *int64  `json:"issueId,omitempty"`
	IssueTitle *string `json:"issueTitle,omitempty"`
	UpdatedAt  *string `json:"updatedAt,omitempty"`
}

type CoreSettings struct {
	UserID                string           `json:"userId"`
	DeviceID              string           `json:"deviceId"`
	TimerMode             TimerMode        `json:"timerMode"`
	BreaksEnabled         bool             `json:"breaksEnabled"`
	WorkDurationMinutes   int              `json:"workDurationMinutes"`
	ShortBreakMinutes     int              `json:"shortBreakMinutes"`
	LongBreakMinutes      int              `json:"longBreakMinutes"`
	LongBreakEnabled      bool             `json:"longBreakEnabled"`
	CyclesBeforeLongBreak int              `json:"cyclesBeforeLongBreak"`
	AutoStartBreaks       bool             `json:"autoStartBreaks"`
	AutoStartWork         bool             `json:"autoStartWork"`
	BoundaryNotifications bool             `json:"boundaryNotificationsEnabled"`
	BoundarySound         bool             `json:"boundarySoundEnabled"`
	AlertSoundPreset      AlertSoundPreset `json:"alertSoundPreset"`
	AlertUrgency          AlertUrgency     `json:"alertUrgency"`
	AlertIconEnabled      bool             `json:"alertIconEnabled"`
	UpdateChecksEnabled   bool             `json:"updateChecksEnabled"`
	UpdatePromptEnabled   bool             `json:"updatePromptEnabled"`
	UpdateChannel         UpdateChannel    `json:"updateChannel"`
	RepoSort              RepoSort         `json:"repoSort"`
	StreamSort            StreamSort       `json:"streamSort"`
	IssueSort             IssueSort        `json:"issueSort"`
	HabitSort             HabitSort        `json:"habitSort"`
	AwayModeEnabled       bool             `json:"awayModeEnabled"`
	FrozenStreakKinds     []StreakKind     `json:"frozenStreakKinds,omitempty"`
	RestWeekdays          []int            `json:"restWeekdays,omitempty"`
	RestSpecificDates     []string         `json:"restSpecificDates,omitempty"`
	DailyPlanRollbackMins int              `json:"dailyPlanRollbackMinutes"`
	CreatedAt             string           `json:"createdAt"`
	UpdatedAt             string           `json:"updatedAt"`
}

type TimerState struct {
	State          string              `json:"state"`
	SessionID      *string             `json:"sessionId,omitempty"`
	IssueID        *int64              `json:"issueId,omitempty"`
	SegmentType    *SessionSegmentType `json:"segmentType,omitempty"`
	ElapsedSeconds int                 `json:"elapsedSeconds,omitempty"`
}

type Stash struct {
	ID                string              `json:"id"`
	UserID            string              `json:"userId"`
	DeviceID          string              `json:"deviceId"`
	RepoID            *int64              `json:"repoId,omitempty"`
	StreamID          *int64              `json:"streamId,omitempty"`
	IssueID           *int64              `json:"issueId,omitempty"`
	SessionID         *string             `json:"sessionId,omitempty"`
	PausedSegmentType *SessionSegmentType `json:"pausedSegmentType,omitempty"`
	ElapsedSeconds    *int                `json:"elapsedSeconds,omitempty"`
	Note              *string             `json:"note,omitempty"`
	CreatedAt         string              `json:"createdAt"`
	UpdatedAt         string              `json:"updatedAt"`
}

type StashConflict struct {
	IssueID int64   `json:"issueId"`
	Stashes []Stash `json:"stashes"`
}

type ScratchPadMeta struct {
	ID           string `json:"id"`
	Path         string `json:"path"`
	Name         string `json:"name"`
	LastOpenedAt string `json:"lastOpenedAt"`
	Pinned       bool   `json:"pinned"`
}

type ScratchPadRead struct {
	OK      bool            `json:"ok"`
	Error   *string         `json:"error,omitempty"`
	Meta    *ScratchPadMeta `json:"meta,omitempty"`
	Content *string         `json:"content,omitempty"`
}

type Op struct {
	ID        string   `json:"id"`
	Entity    OpEntity `json:"entity"`
	EntityID  string   `json:"entityId"`
	Action    OpAction `json:"action"`
	Payload   any      `json:"payload,omitempty"`
	Timestamp string   `json:"timestamp"`
	UserID    string   `json:"userId"`
	DeviceID  string   `json:"deviceId"`
}

type Health struct {
	Status string  `json:"status"`
	DB     bool    `json:"db"`
	OK     int     `json:"ok"`
	Uptime float64 `json:"uptime"`
}

type KernelInfo struct {
	PID             int           `json:"pid"`
	Port            int           `json:"port,omitempty"`
	Transport       string        `json:"transport,omitempty"`
	Endpoint        string        `json:"endpoint,omitempty"`
	SocketPath      string        `json:"socketPath,omitempty"`
	ProtocolVersion string        `json:"protocolVersion"`
	Token           string        `json:"token"`
	StartedAt       string        `json:"startedAt"`
	ScratchDir      string        `json:"scratchDir"`
	Env             string        `json:"env"`
	ExecutablePath  string        `json:"executablePath,omitempty"`
	RunningChannel  UpdateChannel `json:"runningChannel"`
	RunningIsBeta   bool          `json:"runningIsBeta"`
}

type AlertRequest struct {
	Kind        AlertEventKind   `json:"kind"`
	Title       string           `json:"title"`
	Subtitle    string           `json:"subtitle,omitempty"`
	Body        string           `json:"body"`
	Urgency     AlertUrgency     `json:"urgency"`
	IconEnabled bool             `json:"iconEnabled"`
	SoundPreset AlertSoundPreset `json:"soundPreset,omitempty"`
	PlaySound   bool             `json:"playSound"`
}

type AlertReminder struct {
	ID           string                    `json:"id"`
	Kind         AlertReminderKind         `json:"kind"`
	Enabled      bool                      `json:"enabled"`
	ScheduleType AlertReminderScheduleType `json:"scheduleType"`
	Weekdays     []int                     `json:"weekdays,omitempty"`
	TimeHHMM     string                    `json:"timeHHMM"`
	CreatedAt    string                    `json:"createdAt"`
	UpdatedAt    string                    `json:"updatedAt"`
}

type AlertStatus struct {
	NotificationsAvailable bool               `json:"notificationsAvailable"`
	SoundAvailable         bool               `json:"soundAvailable"`
	NotificationBackend    string             `json:"notificationBackend,omitempty"`
	SoundBackend           string             `json:"soundBackend,omitempty"`
	NotificationOptions    []string           `json:"notificationOptions,omitempty"`
	SoundOptions           []string           `json:"soundOptions,omitempty"`
	SubtitleSupported      bool               `json:"subtitleSupported"`
	UrgencySupported       bool               `json:"urgencySupported"`
	IconSupported          bool               `json:"iconSupported"`
	BundledSoundSupported  bool               `json:"bundledSoundSupported"`
	IconPath               string             `json:"iconPath,omitempty"`
	AvailableSoundPresets  []AlertSoundPreset `json:"availableSoundPresets,omitempty"`
}

type UpdateStatus struct {
	CurrentVersion           string        `json:"currentVersion"`
	LatestVersion            string        `json:"latestVersion,omitempty"`
	ReleaseTag               string        `json:"releaseTag,omitempty"`
	ReleaseName              string        `json:"releaseName,omitempty"`
	ReleaseNotes             string        `json:"releaseNotes,omitempty"`
	ReleaseURL               string        `json:"releaseUrl,omitempty"`
	InstallScriptURL         string        `json:"installScriptUrl,omitempty"`
	ChecksumsURL             string        `json:"checksumsUrl,omitempty"`
	PublishedAt              string        `json:"publishedAt,omitempty"`
	CheckedAt                string        `json:"checkedAt,omitempty"`
	UpdateAvailable          bool          `json:"updateAvailable"`
	InstallAvailable         bool          `json:"installAvailable"`
	InstallUnavailableReason string        `json:"installUnavailableReason,omitempty"`
	Enabled                  bool          `json:"enabled"`
	PromptEnabled            bool          `json:"promptEnabled"`
	Channel                  UpdateChannel `json:"channel"`
	ReleaseIsPrerelease      bool          `json:"releaseIsPrerelease"`
	RunningChannel           UpdateChannel `json:"runningChannel"`
	RunningIsBeta            bool          `json:"runningIsBeta"`
	LatestIsBeta             bool          `json:"latestIsBeta"`
	DismissedVersion         string        `json:"dismissedVersion,omitempty"`
	Error                    string        `json:"error,omitempty"`
}

type ExportOutputMode string

const (
	ExportOutputModeFile      ExportOutputMode = "file"
	ExportOutputModeClipboard ExportOutputMode = "clipboard"
)

type ExportFormat string

const (
	ExportFormatMarkdown ExportFormat = "markdown"
	ExportFormatPDF      ExportFormat = "pdf"
	ExportFormatCSV      ExportFormat = "csv"
	ExportFormatICS      ExportFormat = "ics"
)

type ExportReportKind string

const (
	ExportReportKindDaily       ExportReportKind = "daily"
	ExportReportKindWeekly      ExportReportKind = "weekly"
	ExportReportKindRepo        ExportReportKind = "repo"
	ExportReportKindStream      ExportReportKind = "stream"
	ExportReportKindIssueRollup ExportReportKind = "issue_rollup"
	ExportReportKindCSV         ExportReportKind = "csv"
	ExportReportKindCalendar    ExportReportKind = "calendar"
)

type ExportAssetKind string

const (
	ExportAssetKindTemplateMarkdown ExportAssetKind = "template_markdown"
	ExportAssetKindTemplatePDF      ExportAssetKind = "template_pdf"
	ExportAssetKindTemplatePDFHTML  ExportAssetKind = "template_pdf_html"
	ExportAssetKindTemplatePDFCSS   ExportAssetKind = "template_pdf_css"
	ExportAssetKindVariableDocs     ExportAssetKind = "variable_docs"
	ExportAssetKindCSVSpec          ExportAssetKind = "csv_spec"
	ExportAssetKindCSVDocs          ExportAssetKind = "csv_docs"
)

type ExportReportScope struct {
	RepoID     *int64  `json:"repoId,omitempty"`
	RepoName   *string `json:"repoName,omitempty"`
	StreamID   *int64  `json:"streamId,omitempty"`
	StreamName *string `json:"streamName,omitempty"`
}

type ExportTemplateAsset struct {
	ReportKind      ExportReportKind               `json:"reportKind"`
	AssetKind       ExportAssetKind                `json:"assetKind"`
	Label           string                         `json:"label"`
	Name            string                         `json:"name"`
	Engine          string                         `json:"engine"`
	UserPath        string                         `json:"userPath"`
	BundledPath     string                         `json:"bundledPath"`
	Resettable      bool                           `json:"resettable"`
	Exists          bool                           `json:"exists"`
	Customized      bool                           `json:"customized"`
	UpdateAvailable bool                           `json:"updateAvailable"`
	BaseHash        string                         `json:"baseHash"`
	DefaultHash     string                         `json:"defaultHash"`
	ActiveSource    string                         `json:"activeSource"`
	LastSyncedAt    *string                        `json:"lastSyncedAt,omitempty"`
	SelectedPreset  *ExportTemplatePresetSelection `json:"selectedPreset,omitempty"`
	Presets         []ExportTemplatePreset         `json:"presets,omitempty"`
}

type ExportTemplatePreset struct {
	ID           string `json:"id"`
	Label        string `json:"label"`
	Description  string `json:"description"`
	PreviewTitle string `json:"previewTitle"`
	PreviewBody  string `json:"previewBody"`
}

type ExportTemplatePresetSelection struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

type DailyReportIssue struct {
	IssueWithMeta
	WorkedSeconds            int                     `json:"workedSeconds"`
	PlanStatus               DailyPlanEntryStatus    `json:"planStatus"`
	PlanCommittedAt          string                  `json:"planCommittedAt"`
	PlanResolvedAt           *string                 `json:"planResolvedAt,omitempty"`
	PlanFailureReason        *DailyPlanFailureReason `json:"planFailureReason,omitempty"`
	PlanPendingFailureAt     *string                 `json:"planPendingFailureAt,omitempty"`
	PlanPendingFailureReason *DailyPlanFailureReason `json:"planPendingFailureReason,omitempty"`
	PlanBaselineDate         string                  `json:"planBaselineDate,omitempty"`
	PlanCurrentPlannedDate   string                  `json:"planCurrentPlannedDate,omitempty"`
	PlanPostponeCount        int                     `json:"planPostponeCount"`
	PlanCurrentDelayedDays   int                     `json:"planCurrentDelayedDays"`
	PlanMaxDelayedDays       int                     `json:"planMaxDelayedDays"`
	PlanFailScore            float64                 `json:"planFailScore"`
}

type DailyReportSession struct {
	SessionHistoryEntry
	RepoID     int64  `json:"repoId"`
	RepoName   string `json:"repoName"`
	StreamID   int64  `json:"streamId"`
	StreamName string `json:"streamName"`
	IssueTitle string `json:"issueTitle"`
}

type DailyReportData struct {
	Date          string               `json:"date"`
	GeneratedAt   string               `json:"generatedAt"`
	Summary       DailyIssueSummary    `json:"summary"`
	Plan          *DailyPlan           `json:"plan,omitempty"`
	Issues        []DailyReportIssue   `json:"issues"`
	Sessions      []DailyReportSession `json:"sessions"`
	Habits        []HabitDailyItem     `json:"habits"`
	CheckIn       *DailyCheckIn        `json:"checkIn,omitempty"`
	Metrics       *DailyMetricsDay     `json:"metrics,omitempty"`
	MetricsRollup *MetricsRollup       `json:"metricsRollup,omitempty"`
	Streaks       *StreakSummary       `json:"streaks,omitempty"`
}

type ExportAssetStatus struct {
	TemplatePath           string                `json:"templatePath"`
	TemplateDocsPath       string                `json:"templateDocsPath"`
	BundledTemplatePath    string                `json:"bundledTemplatePath"`
	PDFTemplatePath        string                `json:"pdfTemplatePath"`
	PDFBundledTemplatePath string                `json:"pdfBundledTemplatePath"`
	ReportsDir             string                `json:"reportsDir"`
	DefaultReportsDir      string                `json:"defaultReportsDir"`
	ReportsDirCustomized   bool                  `json:"reportsDirCustomized"`
	ICSDir                 string                `json:"icsDir"`
	DefaultICSDir          string                `json:"defaultIcsDir"`
	ICSDirCustomized       bool                  `json:"icsDirCustomized"`
	UserTemplateExists     bool                  `json:"userTemplateExists"`
	UserTemplateCustomized bool                  `json:"userTemplateCustomized"`
	DefaultUpdateAvailable bool                  `json:"defaultUpdateAvailable"`
	PDFUserTemplateExists  bool                  `json:"pdfUserTemplateExists"`
	PDFTemplateCustomized  bool                  `json:"pdfTemplateCustomized"`
	PDFUpdateAvailable     bool                  `json:"pdfUpdateAvailable"`
	TemplateBaseHash       string                `json:"templateBaseHash"`
	CurrentDefaultHash     string                `json:"currentDefaultHash"`
	PDFTemplateBaseHash    string                `json:"pdfTemplateBaseHash"`
	PDFCurrentDefaultHash  string                `json:"pdfCurrentDefaultHash"`
	TemplateName           string                `json:"templateName"`
	TemplateEngine         string                `json:"templateEngine"`
	ActiveTemplateSource   string                `json:"activeTemplateSource"`
	PDFTemplateName        string                `json:"pdfTemplateName"`
	PDFTemplateEngine      string                `json:"pdfTemplateEngine"`
	PDFTemplateSource      string                `json:"pdfTemplateSource"`
	PDFRendererAvailable   bool                  `json:"pdfRendererAvailable"`
	PDFRendererName        string                `json:"pdfRendererName"`
	PDFRendererPath        string                `json:"pdfRendererPath"`
	LastSyncedAt           *string               `json:"lastSyncedAt,omitempty"`
	PDFLastSyncedAt        *string               `json:"pdfLastSyncedAt,omitempty"`
	TemplateAssets         []ExportTemplateAsset `json:"templateAssets,omitempty"`
}

type ExportReportFile struct {
	Name       string           `json:"name"`
	Path       string           `json:"path"`
	Kind       ExportReportKind `json:"kind"`
	ScopeLabel string           `json:"scopeLabel,omitempty"`
	Date       string           `json:"date,omitempty"`
	StartDate  string           `json:"startDate,omitempty"`
	EndDate    string           `json:"endDate,omitempty"`
	DateLabel  string           `json:"dateLabel,omitempty"`
	Format     string           `json:"format"`
	SizeBytes  int64            `json:"sizeBytes"`
	ModifiedAt string           `json:"modifiedAt"`
}

type ExportReportResult struct {
	Kind       ExportReportKind   `json:"kind"`
	Label      string             `json:"label"`
	Scope      *ExportReportScope `json:"scope,omitempty"`
	Date       string             `json:"date,omitempty"`
	StartDate  string             `json:"startDate,omitempty"`
	EndDate    string             `json:"endDate,omitempty"`
	Format     ExportFormat       `json:"format"`
	OutputMode ExportOutputMode   `json:"outputMode"`
	Content    string             `json:"content,omitempty"`
	FilePath   *string            `json:"filePath,omitempty"`
	Renderer   *string            `json:"renderer,omitempty"`
	Assets     ExportAssetStatus  `json:"assets"`
}

type DailyReportResult = ExportReportResult

type CalendarExportResult struct {
	Kind             ExportReportKind   `json:"kind"`
	Label            string             `json:"label"`
	Scope            *ExportReportScope `json:"scope,omitempty"`
	OutputMode       ExportOutputMode   `json:"outputMode"`
	RepoID           int64              `json:"repoId"`
	RepoName         string             `json:"repoName"`
	IssuesFilePath   string             `json:"issuesFilePath"`
	SessionsFilePath string             `json:"sessionsFilePath"`
	Assets           ExportAssetStatus  `json:"assets"`
}
