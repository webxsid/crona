package api

import "encoding/json"

// Repo is a top-level container (maps to a project/codebase).
type Repo struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

// Stream is a work stream within a repo (e.g. "backlog", "sprint-1").
type Stream struct {
	ID     string `json:"id"`
	RepoID string `json:"repoId"`
	Name   string `json:"name"`
}

// Issue is a unit of work within a stream.
type Issue struct {
	ID              string  `json:"id"`
	StreamID        string  `json:"streamId"`
	Title           string  `json:"title"`
	Status          string  `json:"status"`
	EstimateMinutes *int    `json:"estimateMinutes"`
	Notes           *string `json:"notes"`
	TodoForDate     *string `json:"todoForDate"`
	CompletedAt     *string `json:"completedAt"`
	AbandonedAt     *string `json:"abandonedAt"`
}

// IssueWithMeta extends Issue with denormalised repo/stream info.
type IssueWithMeta struct {
	Issue
	RepoID     string `json:"repoId"`
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

// ActiveContext is the current repo/stream/issue selection for this device.
type ActiveContext struct {
	UserID     string  `json:"userId"`
	DeviceID   string  `json:"deviceId"`
	RepoID     *string `json:"repoId"`
	RepoName   *string `json:"repoName"`
	StreamID   *string `json:"streamId"`
	StreamName *string `json:"streamName"`
	IssueID    *string `json:"issueId"`
	IssueTitle *string `json:"issueTitle"`
}

// TimerState represents the kernel's authoritative timer payload.
type TimerState struct {
	State          string  `json:"state"` // "idle" | "running" | "paused"
	SessionID      *string `json:"sessionId"`
	IssueID        *string `json:"issueId"`
	SegmentType    *string `json:"segmentType"`
	ElapsedSeconds int     `json:"elapsedSeconds"`
}

type Health struct {
	Status string  `json:"status"`
	DB     bool    `json:"db"`
	OK     int     `json:"ok"`
	Uptime float64 `json:"uptime"`
}

type Session struct {
	ID              string  `json:"id"`
	IssueID         string  `json:"issueId"`
	StartTime       string  `json:"startTime"`
	EndTime         *string `json:"endTime"`
	DurationSeconds *int    `json:"durationSeconds"`
	Notes           *string `json:"notes"`
}

// ScratchPad is a filesystem-backed note.
type ScratchPad struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Path         string `json:"path"`
	Pinned       bool   `json:"pinned"`
	LastOpenedAt string `json:"lastOpenedAt"`
}

// Op is an immutable audit log entry.
type Op struct {
	ID        string `json:"id"`
	Entity    string `json:"entity"`
	EntityID  string `json:"entityId"`
	Action    string `json:"action"`
	Timestamp string `json:"timestamp"`
	UserID    string `json:"userId"`
	DeviceID  string `json:"deviceId"`
}

// KernelEvent is a raw SSE event from the kernel.
type KernelEvent struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}
