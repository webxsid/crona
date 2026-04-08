package commands

import (
	shareddto "crona/shared/dto"
	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	"crona/tui/internal/logger"

	tea "github.com/charmbracelet/bubbletea"
)

func LoadAlertStatus(c *api.Client) tea.Cmd {
	return func() tea.Msg {
		status, err := c.GetAlertStatus()
		if err != nil {
			logger.Errorf("loadAlertStatus: %v", err)
			return ErrMsg{Err: err}
		}
		return AlertStatusLoadedMsg{Status: status}
	}
}

func TestAlertNotification(c *api.Client) tea.Cmd {
	return func() tea.Msg {
		if err := c.TestAlertNotification(); err != nil {
			logger.Errorf("TestAlertNotification: %v", err)
			return ErrMsg{Err: err}
		}
		return AlertTestedMsg{Label: "Test notification sent"}
	}
}

func TestAlertSound(c *api.Client) tea.Cmd {
	return func() tea.Msg {
		if err := c.TestAlertSound(); err != nil {
			logger.Errorf("TestAlertSound: %v", err)
			return ErrMsg{Err: err}
		}
		return AlertTestedMsg{Label: "Test sound played"}
	}
}

func LoadAlertReminders(c *api.Client) tea.Cmd {
	return func() tea.Msg {
		reminders, err := c.ListAlertReminders()
		if err != nil {
			logger.Errorf("LoadAlertReminders: %v", err)
			return ErrMsg{Err: err}
		}
		return AlertRemindersLoadedMsg{Reminders: reminders}
	}
}

func CreateAlertReminder(c *api.Client, input shareddto.AlertReminderCreateRequest) tea.Cmd {
	return func() tea.Msg {
		_, err := c.CreateAlertReminder(input)
		if err != nil {
			logger.Errorf("CreateAlertReminder: %v", err)
			return ErrMsg{Err: err}
		}
		return AlertReminderChangedMsg{Label: "Reminder created"}
	}
}

func UpdateAlertReminder(c *api.Client, input shareddto.AlertReminderUpdateRequest) tea.Cmd {
	return func() tea.Msg {
		_, err := c.UpdateAlertReminder(input)
		if err != nil {
			logger.Errorf("UpdateAlertReminder: %v", err)
			return ErrMsg{Err: err}
		}
		return AlertReminderChangedMsg{Label: "Reminder updated"}
	}
}

func ToggleAlertReminder(c *api.Client, id string, enabled bool) tea.Cmd {
	return func() tea.Msg {
		_, err := c.ToggleAlertReminder(id, enabled)
		if err != nil {
			logger.Errorf("ToggleAlertReminder(%s): %v", id, err)
			return ErrMsg{Err: err}
		}
		label := "Reminder disabled"
		if enabled {
			label = "Reminder enabled"
		}
		return AlertReminderChangedMsg{Label: label}
	}
}

func DeleteAlertReminder(c *api.Client, id string) tea.Cmd {
	return func() tea.Msg {
		if err := c.DeleteAlertReminder(id); err != nil {
			logger.Errorf("DeleteAlertReminder(%s): %v", id, err)
			return ErrMsg{Err: err}
		}
		return AlertReminderChangedMsg{Label: "Reminder deleted"}
	}
}

func NotifyAlert(c *api.Client, input sharedtypes.AlertRequest) tea.Cmd {
	return func() tea.Msg {
		if err := c.NotifyAlert(input); err != nil {
			logger.Errorf("NotifyAlert(%s): %v", input.Kind, err)
			return ErrMsg{Err: err}
		}
		return nil
	}
}
