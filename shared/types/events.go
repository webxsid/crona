package types

import "encoding/json"

// Shared event payloads used across kernel, TUI, and future CLI clients.

const (
	EventTypeRepoCreated               = "repo.created"
	EventTypeRepoUpdated               = "repo.updated"
	EventTypeRepoDeleted               = "repo.deleted"
	EventTypeStreamCreated             = "stream.created"
	EventTypeStreamUpdated             = "stream.updated"
	EventTypeStreamDeleted             = "stream.deleted"
	EventTypeIssueCreated              = "issue.created"
	EventTypeIssueUpdated              = "issue.updated"
	EventTypeIssueDeleted              = "issue.deleted"
	EventTypeHabitCreated              = "habit.created"
	EventTypeHabitUpdated              = "habit.updated"
	EventTypeHabitDeleted              = "habit.deleted"
	EventTypeHabitCompleted            = "habit.completed"
	EventTypeHabitUncompleted          = "habit.uncompleted"
	EventTypeCheckInUpdated            = "checkin.updated"
	EventTypeCheckInDeleted            = "checkin.deleted"
	EventTypeSessionStarted            = "session.started"
	EventTypeSessionStopped            = "session.stopped"
	EventTypeSessionEnded              = "session.ended"
	EventTypeTimerState                = "timer.state"
	EventTypeContextRepoChanged        = "context.repo.changed"
	EventTypeContextStreamChanged      = "context.stream.changed"
	EventTypeContextIssueChanged       = "context.issue.changed"
	EventTypeContextCleared            = "context.cleared"
	EventTypeTimerBoundary             = "timer.boundary"
	EventTypeTimerHardLimitReached     = "timer.hard_limit_reached"
	EventTypeTimerExtended             = "timer.extended"
	EventTypeTimerBreakDeferred        = "timer.break_deferred"
	EventTypeTimerBreakDeferralWarning = "timer.break_deferral_warning"
	EventTypeTimerTick                 = "timer.tick"
	EventTypeUpdateStatus              = "update.status"
	EventTypeAlertDelivery             = "alert.delivery"
	EventTypeSettingsChanged           = "settings.changed"
	EventTypeDayStart                  = "day.start"
	EventTypeDayEnd                    = "day.end"
)

type KernelEvent struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type DayBoundaryKind string

const (
	DayBoundaryStart DayBoundaryKind = "start"
	DayBoundaryEnd   DayBoundaryKind = "end"
)

type DayBoundaryEventPayload struct {
	Kind               DayBoundaryKind `json:"kind"`
	DateBefore         string          `json:"dateBefore"`
	DateAfter          string          `json:"dateAfter"`
	EffectiveLocalTime string          `json:"effectiveLocalTime"`
	EffectiveUTCTime   string          `json:"effectiveUtcTime"`
	Timezone           string          `json:"timezone"`
	OccurrenceID       string          `json:"occurrenceId"`
	LogicalDate        string          `json:"logicalDate"`
}

type IDEventPayload struct {
	ID int64 `json:"id"`
}

type ContextChangedPayload struct {
	DeviceID string `json:"deviceId"`
	RepoID   *int64 `json:"repoId,omitempty"`
	StreamID *int64 `json:"streamId,omitempty"`
	IssueID  *int64 `json:"issueId,omitempty"`
}

type ContextClearedPayload struct {
	DeviceID string `json:"deviceId"`
}

type TimerBoundaryPayload struct {
	From       SessionSegmentType `json:"from"`
	To         SessionSegmentType `json:"to"`
	Started    bool               `json:"started"`
	Title      string             `json:"title"`
	Message    string             `json:"message"`
	RepoName   *string            `json:"repoName,omitempty"`
	StreamName *string            `json:"streamName,omitempty"`
	IssueID    *int64             `json:"issueId,omitempty"`
	IssueTitle *string            `json:"issueTitle,omitempty"`
}

type TimerTickPayload struct {
	RemainingSeconds int `json:"remainingSeconds"`
}

type SessionEventPayload struct {
	SessionID string `json:"sessionId"`
}

type SettingsChangedPayload struct {
	Keys []CoreSettingsKey `json:"keys"`
}

type TimerBreakDeferralWarningPayload struct {
	SessionID        string `json:"sessionId"`
	SecondsRemaining int    `json:"secondsRemaining"`
	SuggestedSeconds int    `json:"suggestedSeconds"`
}

type TimerHardLimitReachedPayload struct {
	SessionID                      string              `json:"sessionId"`
	IssueID                        int64               `json:"issueId"`
	SegmentType                    *SessionSegmentType `json:"segmentType,omitempty"`
	HardLimitKind                  TimerHardLimitKind  `json:"hardLimitKind,omitempty"`
	HardLimitTotalSeconds          int                 `json:"hardLimitTotalSeconds"`
	HardLimitWorkSeconds           int                 `json:"hardLimitWorkSeconds"`
	HardLimitBreakSeconds          int                 `json:"hardLimitBreakSeconds"`
	HardLimitLongBreakSeconds      int                 `json:"hardLimitLongBreakSeconds"`
	HardLimitCyclesBeforeLongBreak int                 `json:"hardLimitCyclesBeforeLongBreak"`
	ElapsedSeconds                 int                 `json:"elapsedSeconds"`
}
