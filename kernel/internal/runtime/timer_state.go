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

type PreparedTimerState struct {
	SessionID   string                         `json:"sessionId"`
	IssueID     int64                          `json:"issueId"`
	SegmentType sharedtypes.SessionSegmentType `json:"segmentType"`
	RecordedAt  string                         `json:"recordedAt"`
}

func WritePreparedTimerState(state PreparedTimerState) error {
	path, err := preparedTimerStatePath()
	if err != nil {
		return err
	}
	body, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, body, FilePerm())
}

func ReadPreparedTimerState() (*PreparedTimerState, error) {
	path, err := preparedTimerStatePath()
	if err != nil {
		return nil, err
	}
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var state PreparedTimerState
	if err := json.Unmarshal(body, &state); err != nil {
		return nil, err
	}
	if strings.TrimSpace(state.SessionID) == "" || state.IssueID == 0 || strings.TrimSpace(string(state.SegmentType)) == "" {
		return nil, errors.New("invalid prepared timer state")
	}
	return &state, nil
}

func ClearPreparedTimerState() error {
	path, err := preparedTimerStatePath()
	if err != nil {
		return err
	}
	err = os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func NewPreparedTimerState(sessionID string, issueID int64, segmentType sharedtypes.SessionSegmentType) PreparedTimerState {
	return PreparedTimerState{
		SessionID:   strings.TrimSpace(sessionID),
		IssueID:     issueID,
		SegmentType: segmentType,
		RecordedAt:  time.Now().UTC().Format(time.RFC3339),
	}
}

func preparedTimerStatePath() (string, error) {
	base, err := config.RuntimeBaseDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "timer.json"), nil
}
