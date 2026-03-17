package types

import "encoding/json"

// Shared event payloads used across kernel, TUI, and future CLI clients.

const (
	EventTypeRepoCreated          = "repo.created"
	EventTypeRepoUpdated          = "repo.updated"
	EventTypeRepoDeleted          = "repo.deleted"
	EventTypeStreamCreated        = "stream.created"
	EventTypeStreamUpdated        = "stream.updated"
	EventTypeStreamDeleted        = "stream.deleted"
	EventTypeIssueCreated         = "issue.created"
	EventTypeIssueUpdated         = "issue.updated"
	EventTypeIssueDeleted         = "issue.deleted"
	EventTypeHabitCreated         = "habit.created"
	EventTypeHabitUpdated         = "habit.updated"
	EventTypeHabitDeleted         = "habit.deleted"
	EventTypeHabitCompleted       = "habit.completed"
	EventTypeHabitUncompleted     = "habit.uncompleted"
	EventTypeCheckInUpdated       = "checkin.updated"
	EventTypeCheckInDeleted       = "checkin.deleted"
	EventTypeSessionStarted       = "session.started"
	EventTypeSessionStopped       = "session.stopped"
	EventTypeTimerState           = "timer.state"
	EventTypeContextRepoChanged   = "context.repo.changed"
	EventTypeContextStreamChanged = "context.stream.changed"
	EventTypeContextIssueChanged  = "context.issue.changed"
	EventTypeContextCleared       = "context.cleared"
	EventTypeStashCreated         = "stash.created"
	EventTypeStashApplied         = "stash.applied"
	EventTypeStashDropped         = "stash.dropped"
	EventTypeTimerBoundary        = "timer.boundary"
	EventTypeTimerTick            = "timer.tick"
	EventTypeScratchpadCreated    = "scratchpad.created"
	EventTypeScratchpadUpdated    = "scratchpad.updated"
	EventTypeScratchpadDeleted    = "scratchpad.deleted"
)

type KernelEvent struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
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

type StashEventPayload struct {
	ID       string `json:"id"`
	DeviceID string `json:"deviceId"`
	RepoID   *int64 `json:"repoId,omitempty"`
	StreamID *int64 `json:"streamId,omitempty"`
	IssueID  *int64 `json:"issueId,omitempty"`
}

type TimerBoundaryPayload struct {
	From SessionSegmentType `json:"from"`
	To   SessionSegmentType `json:"to"`
}

type TimerTickPayload struct {
	RemainingSeconds int `json:"remainingSeconds"`
}
