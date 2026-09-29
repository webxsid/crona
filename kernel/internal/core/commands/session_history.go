package commands

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"crona/kernel/internal/core"
	"crona/kernel/internal/sessionnotes"
	shareddto "crona/shared/dto"
	sharedtypes "crona/shared/types"

	"github.com/google/uuid"
)

func UpdateSessionHistory(
	ctx context.Context,
	c *core.Context,
	input shareddto.UpdateSessionHistoryRequest,
) (shareddto.OKResponse, error) {
	if strings.TrimSpace(input.ID) == "" {
		return shareddto.OKResponse{}, errors.New("session id is required")
	}
	session, err := requireEditableSession(ctx, c, input.ID)
	if err != nil {
		return shareddto.OKResponse{}, err
	}
	start, err := time.Parse(time.RFC3339Nano, input.StartTime)
	if err != nil {
		return shareddto.OKResponse{}, errors.New("start time must be RFC3339")
	}
	end, err := time.Parse(time.RFC3339Nano, input.EndTime)
	if err != nil || !end.After(start) {
		return shareddto.OKResponse{}, errors.New("end time must be after start time")
	}
	segments := slices.Clone(input.Segments)
	slices.SortFunc(segments, func(a, b sharedtypes.SessionSegment) int {
		startA, errA := time.Parse(time.RFC3339Nano, a.StartTime)
		startB, errB := time.Parse(time.RFC3339Nano, b.StartTime)
		if errA != nil || errB != nil {
			return strings.Compare(a.StartTime, b.StartTime)
		}
		if startA.Before(startB) {
			return -1
		}
		if startA.After(startB) {
			return 1
		}
		return 0
	})
	for i := range segments {
		segmentStart, startErr := time.Parse(time.RFC3339Nano, segments[i].StartTime)
		segmentEnd, endErr := parseSegmentEnd(segments[i].EndTime)
		if startErr != nil || endErr != nil || !segmentEnd.After(segmentStart) ||
			segmentStart.Before(start) || segmentEnd.After(end) {
			return shareddto.OKResponse{}, errors.New("segments must have valid times within the session")
		}
		switch segments[i].SegmentType {
		case sharedtypes.SessionSegmentWork, sharedtypes.SessionSegmentShortBreak,
			sharedtypes.SessionSegmentLongBreak, sharedtypes.SessionSegmentRest:
		default:
			return shareddto.OKResponse{}, errors.New("invalid session segment type")
		}
		if i > 0 {
			previousEnd, _ := parseSegmentEnd(segments[i-1].EndTime)
			if segmentStart.Before(previousEnd) {
				return shareddto.OKResponse{}, errors.New("session segments cannot overlap")
			}
		}
		segments[i].ID = uuid.NewString()
		segments[i].SessionID = input.ID
		segments[i].UserID = c.UserID
		segments[i].DeviceID = c.DeviceID
		segments[i].ElapsedOffsetSeconds = nil
	}
	parsedNotes := sessionnotes.Parse(session.Notes)
	for _, section := range []sharedtypes.SessionNoteSection{
		sharedtypes.SessionNoteSectionCommit,
		sharedtypes.SessionNoteSectionWork,
		sharedtypes.SessionNoteSectionNotes,
	} {
		parsedNotes[section] = strings.TrimSpace(input.EditableNotes[section])
	}
	serializedNotes := sessionnotes.Serialize(parsedNotes)
	var notes *string
	if serializedNotes != "" {
		notes = &serializedNotes
	}
	duration := int(end.Sub(start).Seconds())
	if err := c.Sessions.UpdateHistory(
		ctx, input.ID, session.StartTime, c.UserID, c.DeviceID, start.Format(time.RFC3339Nano),
		end.Format(time.RFC3339Nano), duration, notes, segments, c.Now(),
	); err != nil {
		return shareddto.OKResponse{}, err
	}
	now := c.Now()
	if err := c.Ops.Append(ctx, sharedtypes.Op{
		ID: uuid.NewString(), Entity: sharedtypes.OpEntitySession, EntityID: input.ID,
		Action: sharedtypes.OpActionUpdate, Payload: input, Timestamp: now,
		UserID: c.UserID, DeviceID: c.DeviceID,
	}); err != nil {
		return shareddto.OKResponse{}, fmt.Errorf("recording session history update: %w", err)
	}
	return shareddto.OKResponse{OK: true}, nil
}

func DeleteSessionHistory(
	ctx context.Context,
	c *core.Context,
	sessionID string,
) (shareddto.OKResponse, error) {
	if strings.TrimSpace(sessionID) == "" {
		return shareddto.OKResponse{}, errors.New("session id is required")
	}
	session, err := requireEditableSession(ctx, c, sessionID)
	if err != nil {
		return shareddto.OKResponse{}, err
	}
	now := c.Now()
	if err := c.Sessions.DeleteHistory(ctx, sessionID, session.StartTime, c.UserID, now); err != nil {
		return shareddto.OKResponse{}, err
	}
	if err := c.Ops.Append(ctx, sharedtypes.Op{
		ID: uuid.NewString(), Entity: sharedtypes.OpEntitySession, EntityID: sessionID,
		Action: sharedtypes.OpActionDelete, Payload: map[string]string{"deletedAt": now},
		Timestamp: now, UserID: c.UserID, DeviceID: c.DeviceID,
	}); err != nil {
		return shareddto.OKResponse{}, fmt.Errorf("recording session history deletion: %w", err)
	}
	return shareddto.OKResponse{OK: true}, nil
}

func requireEditableSession(ctx context.Context, c *core.Context, sessionID string) (*sharedtypes.Session, error) {
	settings, err := c.CoreSettings.Get(ctx, c.UserID)
	if err != nil {
		return nil, err
	}
	if settings == nil || !settings.ManageHistory {
		return nil, errors.New("manage history is disabled")
	}
	session, err := c.Sessions.GetByID(ctx, sessionID, c.UserID)
	if err != nil {
		return nil, err
	}
	if session == nil || session.EndTime == nil || strings.TrimSpace(*session.EndTime) == "" {
		return nil, errors.New("completed session not found")
	}
	started, err := time.Parse(time.RFC3339Nano, session.StartTime)
	if err != nil {
		return nil, errors.New("session start time is invalid")
	}
	now, err := time.Parse(time.RFC3339Nano, c.Now())
	if err != nil {
		return nil, errors.New("daemon time is invalid")
	}
	localNow := now.In(time.Local)
	cutoff := time.Date(localNow.Year(), localNow.Month(), localNow.Day(), 0, 0, 0, 0, time.Local).
		AddDate(0, 0, -min(max(settings.ManageHistoryDays, 0), 5))
	if started.In(time.Local).Before(cutoff) {
		return nil, fmt.Errorf(
			"your History Edit Window starts on %s. Increase it in Settings to edit or delete this session",
			cutoff.Format("January 2, 2006"),
		)
	}
	return session, nil
}

func parseSegmentEnd(value *string) (time.Time, error) {
	if value == nil {
		return time.Time{}, errors.New("segment end time is required")
	}
	return time.Parse(time.RFC3339Nano, *value)
}
