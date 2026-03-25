package models

import "github.com/uptrace/bun"

type RepoModel struct {
	bun.BaseModel `bun:"table:repos"`

	InternalID  string  `bun:"id,pk,type:text"`
	PublicID    int64   `bun:"public_id,notnull,type:integer"`
	Name        string  `bun:"name,notnull,unique,type:text"`
	Description *string `bun:"description,type:text,nullzero"`
	Color       *string `bun:"color,type:text,nullzero"`
	UserID      string  `bun:"user_id,notnull,type:text"`
	CreatedAt   string  `bun:"created_at,notnull,type:text"`
	UpdatedAt   string  `bun:"updated_at,notnull,type:text"`
	DeletedAt   *string `bun:"deleted_at,type:text,nullzero"`
}

type StreamModel struct {
	bun.BaseModel `bun:"table:streams"`

	InternalID  string  `bun:"id,pk,type:text"`
	PublicID    int64   `bun:"public_id,notnull,type:integer"`
	RepoID      string  `bun:"repo_id,notnull,type:text"`
	Name        string  `bun:"name,notnull,type:text"`
	Description *string `bun:"description,type:text,nullzero"`
	Visibility  string  `bun:"visibility,notnull,type:text"`
	UserID      string  `bun:"user_id,notnull,type:text"`
	CreatedAt   string  `bun:"created_at,notnull,type:text"`
	UpdatedAt   string  `bun:"updated_at,notnull,type:text"`
	DeletedAt   *string `bun:"deleted_at,type:text,nullzero"`
}

type IssueModel struct {
	bun.BaseModel `bun:"table:issues"`

	InternalID      string  `bun:"id,pk,type:text"`
	PublicID        int64   `bun:"public_id,notnull,type:integer"`
	StreamID        string  `bun:"stream_id,notnull,type:text"`
	Title           string  `bun:"title,notnull,type:text"`
	Description     *string `bun:"description,type:text,nullzero"`
	Status          string  `bun:"status,notnull,type:text"`
	EstimateMinutes *int    `bun:"estimate_minutes,type:integer,nullzero"`
	Notes           *string `bun:"notes,type:text,nullzero"`
	TodoForDate     *string `bun:"todo_for_date,type:text,nullzero"`
	CompletedAt     *string `bun:"completed_at,type:text,nullzero"`
	AbandonedAt     *string `bun:"abandoned_at,type:text,nullzero"`
	UserID          string  `bun:"user_id,notnull,type:text"`
	CreatedAt       string  `bun:"created_at,notnull,type:text"`
	UpdatedAt       string  `bun:"updated_at,notnull,type:text"`
	DeletedAt       *string `bun:"deleted_at,type:text,nullzero"`
}

type HabitModel struct {
	bun.BaseModel `bun:"table:habits"`

	InternalID    string  `bun:"id,pk,type:text"`
	PublicID      int64   `bun:"public_id,notnull,type:integer"`
	StreamID      string  `bun:"stream_id,notnull,type:text"`
	Name          string  `bun:"name,notnull,type:text"`
	Description   *string `bun:"description,type:text,nullzero"`
	ScheduleType  string  `bun:"schedule_type,notnull,type:text"`
	Weekdays      *string `bun:"weekdays,type:text,nullzero"`
	TargetMinutes *int    `bun:"target_minutes,type:integer,nullzero"`
	Active        bool    `bun:"active,notnull,type:integer"`
	UserID        string  `bun:"user_id,notnull,type:text"`
	CreatedAt     string  `bun:"created_at,notnull,type:text"`
	UpdatedAt     string  `bun:"updated_at,notnull,type:text"`
	DeletedAt     *string `bun:"deleted_at,type:text,nullzero"`
}

type HabitCompletionModel struct {
	bun.BaseModel `bun:"table:habit_completions"`

	InternalID      string  `bun:"id,pk,type:text"`
	PublicID        int64   `bun:"public_id,notnull,type:integer"`
	HabitID         string  `bun:"habit_id,notnull,type:text"`
	Date            string  `bun:"date,notnull,type:text"`
	Status          string  `bun:"status,notnull,type:text"`
	DurationMinutes *int    `bun:"duration_minutes,type:integer,nullzero"`
	Notes           *string `bun:"notes,type:text,nullzero"`
	UserID          string  `bun:"user_id,notnull,type:text"`
	CreatedAt       string  `bun:"created_at,notnull,type:text"`
	UpdatedAt       string  `bun:"updated_at,notnull,type:text"`
	DeletedAt       *string `bun:"deleted_at,type:text,nullzero"`
}

type SessionModel struct {
	bun.BaseModel `bun:"table:sessions"`

	ID              string  `bun:",pk,type:text"`
	IssueID         string  `bun:"issue_id,notnull,type:text"`
	StartTime       string  `bun:"start_time,notnull,type:text"`
	EndTime         *string `bun:"end_time,type:text,nullzero"`
	DurationSeconds *int    `bun:"duration_seconds,type:integer,nullzero"`
	Notes           *string `bun:"notes,type:text,nullzero"`
	UserID          string  `bun:"user_id,notnull,type:text"`
	DeviceID        string  `bun:"device_id,notnull,type:text"`
	CreatedAt       string  `bun:"created_at,notnull,type:text"`
	UpdatedAt       string  `bun:"updated_at,notnull,type:text"`
	DeletedAt       *string `bun:"deleted_at,type:text,nullzero"`
}

type StashModel struct {
	bun.BaseModel `bun:"table:stash"`

	ID               string  `bun:",pk,type:text"`
	RepoID           *string `bun:"repo_id,type:text,nullzero"`
	StreamID         *string `bun:"stream_id,type:text,nullzero"`
	IssueID          *string `bun:"issue_id,type:text,nullzero"`
	SessionID        *string `bun:"session_id,type:text,nullzero"`
	SegmentType      *string `bun:"segment_type,type:text,nullzero"`
	SegmentStartedAt *string `bun:"segment_started_at,type:text,nullzero"`
	ElapsedSeconds   *int    `bun:"elapsed_seconds,type:integer,nullzero"`
	Note             *string `bun:"note,type:text,nullzero"`
	UserID           string  `bun:"user_id,notnull,type:text"`
	DeviceID         string  `bun:"device_id,notnull,type:text"`
	CreatedAt        string  `bun:"created_at,notnull,type:text"`
	UpdatedAt        string  `bun:"updated_at,notnull,type:text"`
	DeletedAt        *string `bun:"deleted_at,type:text,nullzero"`
}

type OpModel struct {
	bun.BaseModel `bun:"table:ops"`

	ID        string `bun:",pk,type:text"`
	UserID    string `bun:"user_id,notnull,type:text"`
	DeviceID  string `bun:"device_id,notnull,type:text"`
	Entity    string `bun:"entity,notnull,type:text"`
	EntityID  string `bun:"entity_id,notnull,type:text"`
	Action    string `bun:"action,notnull,type:text"`
	Payload   string `bun:"payload,notnull,type:text"`
	Timestamp string `bun:"timestamp,notnull,type:text"`
}

type CoreSettingsModel struct {
	bun.BaseModel `bun:"table:core_settings"`

	UserID                string `bun:"user_id,pk,type:text"`
	DeviceID              string `bun:"device_id,notnull,type:text"`
	TimerMode             string `bun:"timer_mode,notnull,type:text"`
	BreaksEnabled         bool   `bun:"breaks_enabled,notnull,type:integer"`
	WorkDurationMinutes   int    `bun:"work_duration_minutes,notnull,type:integer"`
	ShortBreakMinutes     int    `bun:"short_break_minutes,notnull,type:integer"`
	LongBreakMinutes      int    `bun:"long_break_minutes,notnull,type:integer"`
	LongBreakEnabled      bool   `bun:"long_break_enabled,notnull,type:integer"`
	CyclesBeforeLongBreak int    `bun:"cycles_before_long_break,notnull,type:integer"`
	AutoStartBreaks       bool   `bun:"auto_start_breaks,notnull,type:integer"`
	AutoStartWork         bool   `bun:"auto_start_work,notnull,type:integer"`
	BoundaryNotifications bool   `bun:"boundary_notifications_enabled,notnull,type:integer"`
	BoundarySound         bool   `bun:"boundary_sound_enabled,notnull,type:integer"`
	UpdateChecksEnabled   bool   `bun:"update_checks_enabled,notnull,type:integer"`
	UpdatePromptEnabled   bool   `bun:"update_prompt_enabled,notnull,type:integer"`
	RepoSort              string `bun:"repo_sort,notnull,type:text"`
	StreamSort            string `bun:"stream_sort,notnull,type:text"`
	IssueSort             string `bun:"issue_sort,notnull,type:text"`
	HabitSort             string `bun:"habit_sort,notnull,type:text"`
	CreatedAt             string `bun:"created_at,notnull,type:text"`
	UpdatedAt             string `bun:"updated_at,notnull,type:text"`
}

type SessionSegmentModel struct {
	bun.BaseModel `bun:"table:session_segments"`

	ID                   string  `bun:",pk,type:text"`
	UserID               string  `bun:"user_id,notnull,type:text"`
	DeviceID             string  `bun:"device_id,notnull,type:text"`
	SessionID            string  `bun:"session_id,notnull,type:text"`
	SegmentType          string  `bun:"segment_type,notnull,type:text"`
	ElapsedOffsetSeconds *int    `bun:"elapsed_offset_seconds,type:integer,nullzero"`
	StartTime            string  `bun:"start_time,notnull,type:text"`
	EndTime              *string `bun:"end_time,type:text,nullzero"`
	CreatedAt            string  `bun:"created_at,notnull,type:text"`
}

type ActiveContextModel struct {
	bun.BaseModel `bun:"table:active_context"`

	UserID    string  `bun:"user_id,pk,type:text"`
	DeviceID  string  `bun:"device_id,notnull,type:text"`
	RepoID    *string `bun:"repo_id,type:text,nullzero"`
	StreamID  *string `bun:"stream_id,type:text,nullzero"`
	IssueID   *string `bun:"issue_id,type:text,nullzero"`
	UpdatedAt string  `bun:"updated_at,notnull,type:text"`
}

type ScratchPadMetaModel struct {
	bun.BaseModel `bun:"table:scratch_pad_meta"`

	ID           string `bun:",pk,type:text"`
	UserID       string `bun:"user_id,notnull,type:text"`
	DeviceID     string `bun:"device_id,notnull,type:text"`
	Name         string `bun:"name,notnull,type:text"`
	Path         string `bun:"path,notnull,unique,type:text"`
	LastOpenedAt string `bun:"last_opened_at,notnull,type:text"`
	Pinned       bool   `bun:"pinned,notnull,type:integer"`
}

type DailyCheckInModel struct {
	bun.BaseModel `bun:"table:daily_checkins"`

	UserID            string   `bun:"user_id,pk,type:text"`
	DeviceID          string   `bun:"device_id,notnull,type:text"`
	Date              string   `bun:"date,pk,type:text"`
	Mood              int      `bun:"mood,notnull,type:integer"`
	Energy            int      `bun:"energy,notnull,type:integer"`
	SleepHours        *float64 `bun:"sleep_hours,type:real,nullzero"`
	SleepScore        *int     `bun:"sleep_score,type:integer,nullzero"`
	ScreenTimeMinutes *int     `bun:"screen_time_minutes,type:integer,nullzero"`
	Notes             *string  `bun:"notes,type:text,nullzero"`
	CreatedAt         string   `bun:"created_at,notnull,type:text"`
	UpdatedAt         string   `bun:"updated_at,notnull,type:text"`
}
