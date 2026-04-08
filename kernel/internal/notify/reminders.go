package notify

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	shareddto "crona/shared/dto"
	sharedtypes "crona/shared/types"
)

var alertNowFn = time.Now

func (s *Service) ListReminders(ctx context.Context) ([]sharedtypes.AlertReminder, error) {
	return s.core.AlertReminders.List(ctx, s.core.UserID)
}

func (s *Service) CreateReminder(ctx context.Context, input shareddto.AlertReminderCreateRequest) (*sharedtypes.AlertReminder, error) {
	enabled := true
	if input.Enabled != nil {
		enabled = *input.Enabled
	}
	reminder := sharedtypes.AlertReminder{
		Kind:         sharedtypes.NormalizeAlertReminderKind(input.Kind),
		Enabled:      enabled,
		ScheduleType: sharedtypes.NormalizeAlertReminderScheduleType(input.ScheduleType),
		Weekdays:     normalizeReminderWeekdays(input.Weekdays),
		TimeHHMM:     strings.TrimSpace(input.TimeHHMM),
	}
	if err := validateReminder(reminder); err != nil {
		return nil, err
	}
	return s.core.AlertReminders.Create(ctx, s.core.UserID, s.core.DeviceID, reminder, s.core.Now())
}

func (s *Service) UpdateReminder(ctx context.Context, input shareddto.AlertReminderUpdateRequest) (*sharedtypes.AlertReminder, error) {
	current, err := s.core.AlertReminders.GetByID(ctx, s.core.UserID, strings.TrimSpace(input.ID))
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, errors.New("alert reminder not found")
	}
	if input.Enabled != nil {
		current.Enabled = *input.Enabled
	}
	if input.ScheduleType != nil {
		current.ScheduleType = sharedtypes.NormalizeAlertReminderScheduleType(*input.ScheduleType)
	}
	if input.Weekdays != nil {
		current.Weekdays = normalizeReminderWeekdays(input.Weekdays)
	}
	if input.TimeHHMM != nil {
		current.TimeHHMM = strings.TrimSpace(*input.TimeHHMM)
	}
	if err := validateReminder(*current); err != nil {
		return nil, err
	}
	return s.core.AlertReminders.Update(ctx, s.core.UserID, *current, s.core.Now())
}

func (s *Service) DeleteReminder(ctx context.Context, id string) error {
	return s.core.AlertReminders.Delete(ctx, s.core.UserID, strings.TrimSpace(id))
}

func (s *Service) ToggleReminder(ctx context.Context, id string, enabled bool) (*sharedtypes.AlertReminder, error) {
	return s.core.AlertReminders.SetEnabled(ctx, s.core.UserID, strings.TrimSpace(id), enabled, s.core.Now())
}

func (s *Service) runReminderScheduler(ctx context.Context) {
	s.processReminderTick(ctx, alertNowFn())
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			s.processReminderTick(ctx, now)
		}
	}
}

func (s *Service) processReminderTick(ctx context.Context, now time.Time) {
	reminders, err := s.core.AlertReminders.List(ctx, s.core.UserID)
	if err != nil {
		s.logger.Error("list alert reminders", err)
		return
	}
	for _, reminder := range reminders {
		if !shouldFireReminder(reminder, now) {
			continue
		}
		slotKey := reminderSlotKey(reminder, now)
		if s.wasReminderDelivered(reminder.ID, slotKey) {
			continue
		}
		if shouldSuppressReminder(ctx, s, reminder, now) {
			s.markReminderDelivered(reminder.ID, slotKey)
			continue
		}
		if err := s.deliver(ctx, reminderAlert(reminder), true); err != nil {
			s.logger.Error("deliver scheduled reminder", err)
			continue
		}
		s.markReminderDelivered(reminder.ID, slotKey)
	}
}

func shouldSuppressReminder(ctx context.Context, s *Service, reminder sharedtypes.AlertReminder, now time.Time) bool {
	switch reminder.Kind {
	case sharedtypes.AlertReminderKindCheckIn:
		checkIn, err := s.core.DailyCheckIns.GetByDate(ctx, s.core.UserID, now.Format("2006-01-02"))
		if err != nil {
			s.logger.Error("lookup daily check-in for reminder", err)
			return false
		}
		return checkIn != nil
	default:
		return false
	}
}

func reminderAlert(reminder sharedtypes.AlertReminder) sharedtypes.AlertRequest {
	switch reminder.Kind {
	case sharedtypes.AlertReminderKindCheckIn:
		return sharedtypes.AlertRequest{
			Kind:      sharedtypes.AlertEventCheckInReminder,
			Title:     "Daily check-in reminder",
			Subtitle:  reminder.TimeHHMM,
			Body:      "Add today’s check-in to keep your wellbeing dashboard current.",
			Urgency:   sharedtypes.AlertUrgencyNormal,
			PlaySound: false,
		}
	default:
		return sharedtypes.AlertRequest{
			Kind:      sharedtypes.AlertEventCheckInReminder,
			Title:     "Crona reminder",
			Subtitle:  reminder.TimeHHMM,
			Body:      "You have a scheduled Crona reminder.",
			Urgency:   sharedtypes.AlertUrgencyNormal,
			PlaySound: false,
		}
	}
}

func shouldFireReminder(reminder sharedtypes.AlertReminder, now time.Time) bool {
	if !reminder.Enabled {
		return false
	}
	hour, minute, ok := parseReminderTime(reminder.TimeHHMM)
	if !ok || now.Hour() != hour || now.Minute() != minute {
		return false
	}
	switch sharedtypes.NormalizeAlertReminderScheduleType(reminder.ScheduleType) {
	case sharedtypes.AlertReminderScheduleWeekly:
		weekday := int(now.Weekday())
		return slices.Contains(normalizeReminderWeekdays(reminder.Weekdays), weekday)
	default:
		return true
	}
}

func validateReminder(reminder sharedtypes.AlertReminder) error {
	reminder.Kind = sharedtypes.NormalizeAlertReminderKind(reminder.Kind)
	reminder.ScheduleType = sharedtypes.NormalizeAlertReminderScheduleType(reminder.ScheduleType)
	if _, _, ok := parseReminderTime(reminder.TimeHHMM); !ok {
		return errors.New("reminder time must be in HH:MM format")
	}
	if reminder.ScheduleType == sharedtypes.AlertReminderScheduleWeekly && len(normalizeReminderWeekdays(reminder.Weekdays)) == 0 {
		return errors.New("weekly reminders require at least one weekday")
	}
	return nil
}

func parseReminderTime(value string) (int, int, bool) {
	parsed, err := time.Parse("15:04", strings.TrimSpace(value))
	if err != nil {
		return 0, 0, false
	}
	return parsed.Hour(), parsed.Minute(), true
}

func normalizeReminderWeekdays(values []int) []int {
	if len(values) == 0 {
		return nil
	}
	out := make([]int, 0, len(values))
	for _, value := range values {
		if value < 0 || value > 6 || slices.Contains(out, value) {
			continue
		}
		out = append(out, value)
	}
	slices.Sort(out)
	return out
}

func reminderSlotKey(reminder sharedtypes.AlertReminder, now time.Time) string {
	return fmt.Sprintf("%s@%s", now.Format("2006-01-02"), reminder.TimeHHMM)
}

func (s *Service) wasReminderDelivered(id, slot string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastReminderSlots[id] == slot
}

func (s *Service) markReminderDelivered(id, slot string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastReminderSlots[id] = slot
}
