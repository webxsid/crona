package runtime

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"crona/shared/config"
	sharedtypes "crona/shared/types"
)

type TimerRuntimeState struct {
	SessionID                      string                          `json:"sessionId"`
	IssueID                        int64                           `json:"issueId"`
	PreparedSegmentType            *sharedtypes.SessionSegmentType `json:"preparedSegmentType,omitempty"`
	HardLimitTotalSeconds          int                             `json:"hardLimitTotalSeconds,omitempty"`
	HardLimitWorkSeconds           int                             `json:"hardLimitWorkSeconds,omitempty"`
	HardLimitBreakSeconds          int                             `json:"hardLimitBreakSeconds,omitempty"`
	HardLimitLongBreakSeconds      int                             `json:"hardLimitLongBreakSeconds,omitempty"`
	HardLimitCyclesBeforeLongBreak int                             `json:"hardLimitCyclesBeforeLongBreak,omitempty"`
	HardLimitExpired               bool                            `json:"hardLimitExpired,omitempty"`
	HardLimitExpiredAt             string                          `json:"hardLimitExpiredAt,omitempty"`
	RecordedAt                     string                          `json:"recordedAt"`
}

func (s *TimerRuntimeState) UnmarshalJSON(data []byte) error {
	type timerRuntimeStateJSON struct {
		SessionID                      string                          `json:"sessionId"`
		IssueID                        int64                           `json:"issueId"`
		PreparedSegmentType            *sharedtypes.SessionSegmentType `json:"preparedSegmentType,omitempty"`
		LegacyPreparedSegmentType      *sharedtypes.SessionSegmentType `json:"segmentType,omitempty"`
		HardLimitTotalSeconds          int                             `json:"hardLimitTotalSeconds,omitempty"`
		HardLimitWorkSeconds           int                             `json:"hardLimitWorkSeconds,omitempty"`
		HardLimitBreakSeconds          int                             `json:"hardLimitBreakSeconds,omitempty"`
		HardLimitLongBreakSeconds      int                             `json:"hardLimitLongBreakSeconds,omitempty"`
		HardLimitCyclesBeforeLongBreak int                             `json:"hardLimitCyclesBeforeLongBreak,omitempty"`
		HardLimitExpired               bool                            `json:"hardLimitExpired,omitempty"`
		HardLimitExpiredAt             string                          `json:"hardLimitExpiredAt,omitempty"`
		RecordedAt                     string                          `json:"recordedAt"`
	}
	var decoded timerRuntimeStateJSON
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	s.SessionID = decoded.SessionID
	s.IssueID = decoded.IssueID
	s.PreparedSegmentType = decoded.PreparedSegmentType
	if s.PreparedSegmentType == nil {
		s.PreparedSegmentType = decoded.LegacyPreparedSegmentType
	}
	s.HardLimitTotalSeconds = decoded.HardLimitTotalSeconds
	s.HardLimitWorkSeconds = decoded.HardLimitWorkSeconds
	s.HardLimitBreakSeconds = decoded.HardLimitBreakSeconds
	s.HardLimitLongBreakSeconds = decoded.HardLimitLongBreakSeconds
	s.HardLimitCyclesBeforeLongBreak = decoded.HardLimitCyclesBeforeLongBreak
	s.HardLimitExpired = decoded.HardLimitExpired
	s.HardLimitExpiredAt = decoded.HardLimitExpiredAt
	s.RecordedAt = decoded.RecordedAt
	return nil
}

func (s TimerRuntimeState) HasPreparedSegment() bool {
	return s.PreparedSegmentType != nil && strings.TrimSpace(string(*s.PreparedSegmentType)) != ""
}

func (s TimerRuntimeState) HasHardLimit() bool {
	return s.HardLimitTotalSeconds > 0
}

func WriteTimerRuntimeState(state TimerRuntimeState) error {
	path, err := timerRuntimeStatePath()
	if err != nil {
		return err
	}
	body, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, body, FilePerm())
}

func ReadTimerRuntimeState() (*TimerRuntimeState, error) {
	path, err := timerRuntimeStatePath()
	if err != nil {
		return nil, err
	}
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var state TimerRuntimeState
	if err := json.Unmarshal(body, &state); err != nil {
		return nil, err
	}
	if strings.TrimSpace(state.SessionID) == "" || state.IssueID == 0 {
		return nil, errors.New("invalid timer runtime state")
	}
	if state.HasPreparedSegment() {
		return &state, nil
	}
	if state.HasHardLimit() {
		return &state, nil
	}
	return nil, errors.New("invalid timer runtime state")
}

func ClearTimerRuntimeState() error {
	path, err := timerRuntimeStatePath()
	if err != nil {
		return err
	}
	err = os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func NewPreparedTimerRuntimeState(
	sessionID string,
	issueID int64,
	segmentType sharedtypes.SessionSegmentType,
) TimerRuntimeState {
	return TimerRuntimeState{
		SessionID:           strings.TrimSpace(sessionID),
		IssueID:             issueID,
		PreparedSegmentType: valuePtr(segmentType),
		RecordedAt:          time.Now().UTC().Format(time.RFC3339),
	}
}

func NewHardLimitTimerRuntimeState(
	sessionID string,
	issueID int64,
	totalSeconds, workSeconds, breakSeconds int,
	longBreakSeconds, cyclesBeforeLongBreak int,
) TimerRuntimeState {
	return TimerRuntimeState{
		SessionID:                      strings.TrimSpace(sessionID),
		IssueID:                        issueID,
		HardLimitTotalSeconds:          totalSeconds,
		HardLimitWorkSeconds:           workSeconds,
		HardLimitBreakSeconds:          breakSeconds,
		HardLimitLongBreakSeconds:      longBreakSeconds,
		HardLimitCyclesBeforeLongBreak: cyclesBeforeLongBreak,
		RecordedAt:                     time.Now().UTC().Format(time.RFC3339),
	}
}

func timerRuntimeStatePath() (string, error) {
	base, err := config.RuntimeBaseDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "timer.json"), nil
}

func valuePtr(value sharedtypes.SessionSegmentType) *sharedtypes.SessionSegmentType {
	return &value
}
