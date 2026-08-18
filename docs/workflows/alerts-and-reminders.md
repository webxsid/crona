---
hosted: true
title: Alerts and Reminders
description: System notifications, sound profiles, inactivity tracking, and scheduled reminder rules.
order: 5.8
---

Crona processes alerts locally through the background daemon. The TUI lets you configure and test alerts; the daemon evaluates timers and reminders, then delivers them through the available local OS helper.

## Trigger Scopes

The local alert layer fires on these events:
- **Timer Boundaries**: Transitions between work segments, short breaks, and long breaks.
- **Hard-Limit Expiry**: Expired Pomodoro and countdown timers prompt you to commit the session or extend it.
- **Inactivity Warning**: Triggered if a focus session continues running without keypress activity from the TUI.
- **System Events**: Completion of exports, diagnostic support bundles, or update detections.
- **Scheduled Reminders**: Recurring alarms for check-ins or planning the day.
- **Day Boundaries**: Start-of-day and end-of-day schedule events.

## Scheduled Reminders

Reminder rules can be configured via the **Alerts** view in the TUI:
- **Schedules**: Daily or weekly reminder alarms.
- **Kinds**: `checkin_reminder` and `daily_plan_reminder`.
- **Action Group**: Create, edit, toggle, or delete rules directly from the interface.
- **Suppression**: Check-in reminders stop firing once today's check-in exists. Daily-plan reminders stop firing once today's plan contains an item.
- **Prerequisite**: Scheduled reminders only trigger while `crona-daemon` is running.

## Day-Boundary Schedules

Configure **Start of Day** and **End of Day** in the TUI Settings view. Each schedule has:

- an enable/disable toggle
- a default local wall-clock time in `HH:mm` format
- optional weekday-specific overrides (Monday through Sunday)

Start of Day is enabled at `00:00` by default. End of Day is disabled by default. The daemon evaluates these schedules, records each occurrence so a restart cannot duplicate it, and routes End of Day notifications through the normal local alert system. Schedules are daemon-owned and only run while `crona-daemon` is running.

## Alert Customization & Presets

You can adjust how notifications behave per alert type:
- **Toggles**: Enable/disable visual toasts, audio cues, or both.
- **Urgency**: Set notification priority levels for system backends.
- **Audible Alerts**: Choose from bundled royalty-free sound effects:
  - `chime`
  - `soft_bell`
  - `notification_ping`
  - `focus_gong`
  - `minimal_click`

## Platform Notification Backends

Alert delivery depends on platform-specific helpers:

| OS | Notification Helper | Audio Player |
| --- | --- | --- |
| **macOS** | `terminal-notifier` (fallback: `osascript`) | `afplay` |
| **Linux** | `notify-send` | `paplay`, `aplay`, `play` (fallback: `canberra-gtk-play`) |
| **Windows** | `BurntToast` (fallback: PowerShell toasts) | PowerShell `SoundPlayer` |

You can check active backend capabilities (e.g., subtitle, urgency, icon support) in the **Alerts** panel.
