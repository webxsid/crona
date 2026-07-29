package dayboundary

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"

	sharedtypes "crona/shared/types"
)

type ClaimFunc func(context.Context, string, string, string, string, string) (bool, error)
type EmitFunc func(sharedtypes.KernelEvent)
type AlertFunc func(context.Context, sharedtypes.AlertRequest) error
type SettingsFunc func(context.Context) (*sharedtypes.CoreSettings, error)

type Scheduler struct {
	settings SettingsFunc
	claim    ClaimFunc
	emit     EmitFunc
	alert    AlertFunc
	now      func() time.Time
	location func() *time.Location
	refresh  chan struct{}
	mu       sync.RWMutex
	current  string
}

func New(
	settings SettingsFunc,
	claim ClaimFunc,
	emit EmitFunc,
	alert AlertFunc,
) *Scheduler {
	return &Scheduler{
		settings: settings,
		claim:    claim,
		emit:     emit,
		alert:    alert,
		now:      time.Now,
		location: func() *time.Location { return time.Now().Location() },
		refresh:  make(chan struct{}, 1),
	}
}

func (s *Scheduler) SetClock(now func() time.Time, location func() *time.Location) {
	s.now = now
	s.location = location
}

func (s *Scheduler) Refresh() {
	select {
	case s.refresh <- struct{}{}:
	default:
	}
}

func (s *Scheduler) Initialize(ctx context.Context) error {
	settings, err := s.settings(ctx)
	if err != nil {
		return err
	}
	if settings == nil {
		settings = &sharedtypes.CoreSettings{StartOfDay: sharedtypes.DefaultStartOfDaySchedule()}
	}
	location := s.location()
	if location == nil {
		location = time.Local
	}
	s.setCurrent(LogicalDate(s.now().In(location), settings.StartOfDay))
	return nil
}

func (s *Scheduler) CurrentDate() string {
	s.mu.RLock()
	current := s.current
	s.mu.RUnlock()
	if current == "" {
		return s.now().In(s.location()).Format("2006-01-02")
	}
	return current
}

func (s *Scheduler) Run(ctx context.Context) error {
	if s.settings == nil || s.claim == nil || s.emit == nil {
		return fmt.Errorf("day boundary scheduler dependencies are incomplete")
	}
	for {
		settings, err := s.settings(ctx)
		if err != nil {
			return err
		}
		now := s.now()
		location := s.location()
		if location == nil {
			location = time.Local
		}
		now = now.In(location)
		if settings == nil {
			settings = &sharedtypes.CoreSettings{StartOfDay: sharedtypes.DefaultStartOfDaySchedule(), EndOfDay: sharedtypes.DefaultEndOfDaySchedule()}
		}
		s.setCurrent(LogicalDate(now, settings.StartOfDay))
		next, kind, schedule := nextBoundary(now, settings)
		if next.IsZero() {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-s.refresh:
				continue
			case <-time.After(time.Minute):
				continue
			}
		}
		timer := time.NewTimer(time.Until(next))
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-s.refresh:
			timer.Stop()
			continue
		case <-time.After(time.Minute):
			timer.Stop()
			continue
		case <-timer.C:
			if err := s.fire(ctx, kind, next, schedule, location, settings); err != nil {
				return err
			}
		}
	}
}

func (s *Scheduler) fire(
	ctx context.Context,
	kind sharedtypes.DayBoundaryKind,
	at time.Time,
	schedule sharedtypes.DayBoundarySchedule,
	location *time.Location,
	settings *sharedtypes.CoreSettings,
) error {
	local := at.In(location)
	scheduledAtUTC := local.UTC().Format(time.RFC3339Nano)
	occurrenceID := fmt.Sprintf("day-boundary:%s:%s:%s", kind, scheduledAtUTC, location.String())
	claimed, err := s.claim(ctx, occurrenceID, settings.UserID, string(kind), scheduledAtUTC, location.String())
	if err != nil {
		return err
	}
	if !claimed {
		return nil
	}
	dateBefore := LogicalDate(local.Add(-time.Nanosecond), settings.StartOfDay)
	dateAfter := LogicalDate(local.Add(time.Nanosecond), settings.StartOfDay)
	payload := sharedtypes.DayBoundaryEventPayload{
		Kind:               kind,
		DateBefore:         dateBefore,
		DateAfter:          dateAfter,
		EffectiveLocalTime: local.Format(time.RFC3339),
		EffectiveUTCTime:   scheduledAtUTC,
		Timezone:           location.String(),
		OccurrenceID:       occurrenceID,
		LogicalDate:        dateAfter,
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	eventType := sharedtypes.EventTypeDayEnd
	if kind == sharedtypes.DayBoundaryStart {
		eventType = sharedtypes.EventTypeDayStart
		s.setCurrent(dateAfter)
	}
	s.emit(sharedtypes.KernelEvent{Type: eventType, Payload: b})
	if kind == sharedtypes.DayBoundaryEnd && s.alert != nil {
		return s.alert(ctx, sharedtypes.AlertRequest{
			Kind:        sharedtypes.AlertEventDayBoundary,
			Title:       "Time to rest",
			Body:        fmt.Sprintf("The configured Crona day ended at %s.", schedule.TimeForWeekday(local.Weekday())),
			Urgency:     sharedtypes.AlertUrgencyNormal,
			PlaySound:   true,
			IconEnabled: true,
		})
	}
	return nil
}

func (s *Scheduler) setCurrent(date string) {
	s.mu.Lock()
	s.current = date
	s.mu.Unlock()
}

func LogicalDate(now time.Time, schedule sharedtypes.DayBoundarySchedule) string {
	now = now.In(now.Location())
	if !schedule.Enabled {
		return now.Format("2006-01-02")
	}
	boundaryTime := schedule.TimeForWeekday(now.Weekday())
	parsed, err := time.Parse("15:04", boundaryTime)
	if err != nil {
		return now.Format("2006-01-02")
	}
	boundary := time.Date(now.Year(), now.Month(), now.Day(), parsed.Hour(), parsed.Minute(), 0, 0, now.Location())
	if now.Before(boundary) {
		return now.AddDate(0, 0, -1).Format("2006-01-02")
	}
	return now.Format("2006-01-02")
}

func nextBoundary(now time.Time, settings *sharedtypes.CoreSettings) (time.Time, sharedtypes.DayBoundaryKind, sharedtypes.DayBoundarySchedule) {
	type candidate struct {
		at       time.Time
		kind     sharedtypes.DayBoundaryKind
		schedule sharedtypes.DayBoundarySchedule
	}
	var candidates []candidate
	for offset := 0; offset <= 8; offset++ {
		date := now.AddDate(0, 0, offset)
		for _, item := range []struct {
			kind     sharedtypes.DayBoundaryKind
			schedule sharedtypes.DayBoundarySchedule
		}{
			{sharedtypes.DayBoundaryStart, settings.StartOfDay},
			{sharedtypes.DayBoundaryEnd, settings.EndOfDay},
		} {
			if !item.schedule.Enabled {
				continue
			}
			minutes, err := item.schedule.MinutesForWeekday(date.Weekday())
			if err != nil {
				continue
			}
			at := time.Date(date.Year(), date.Month(), date.Day(), minutes/60, minutes%60, 0, 0, now.Location())
			if at.After(now) {
				candidates = append(candidates, candidate{at: at, kind: item.kind, schedule: item.schedule})
			}
		}
	}
	if len(candidates) == 0 {
		return time.Time{}, "", sharedtypes.DayBoundarySchedule{}
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].at.Equal(candidates[j].at) {
			return candidates[i].kind == sharedtypes.DayBoundaryStart
		}
		return candidates[i].at.Before(candidates[j].at)
	})
	return candidates[0].at, candidates[0].kind, candidates[0].schedule
}
