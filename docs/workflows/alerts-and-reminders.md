---
hosted: true
title: Alerts and Reminders
description: System notifications, sound profiles, inactivity tracking, and scheduled reminder rules.
order: 5.8
---

Alerts are useful when you are deep in work and do not want to watch the clock. Crona evaluates them locally through the background daemon and uses the notification tools available on your machine.

## What can notify you

The local alert layer fires on these events:
- **Timer Boundaries**: Transitions between work segments, short breaks, and long breaks.
- **Hard-Limit Expiry**: Expired Pomodoro and countdown timers prompt you to commit the session or extend it.
- **Inactivity Warning**: Triggered if a focus session continues running without keypress activity from the TUI.
- **System Events**: Completion of exports, diagnostic support bundles, or update detections.
- **Scheduled Reminders**: Recurring alarms for check-ins or planning the day.
- **Day Boundaries**: Start-of-day and end-of-day schedule events.

## Set a reminder

In the TUI's **Alerts** view, you can create, edit, pause, or delete:
- **Schedules**: Daily or weekly reminder alarms.
- **Kinds**: `checkin_reminder` and `daily_plan_reminder`.
- **Action Group**: Create, edit, toggle, or delete rules directly from the interface.
- **Suppression**: Check-in reminders stop firing once today's check-in exists. Daily-plan reminders stop firing once today's plan contains an item.
- **Prerequisite**: Scheduled reminders only trigger while `crona-daemon` is running.

## Start and end of day

Configure **Start of Day** and **End of Day** in the TUI Settings view. Each schedule has:

- an enable/disable toggle
- a default local wall-clock time in `HH:mm` format
- optional weekday-specific overrides (Monday through Sunday)

Start of Day is enabled at `00:00` by default. End of Day is disabled by default. The daemon evaluates these schedules, records each occurrence so a restart cannot duplicate it, and routes End of Day notifications through the normal local alert system. Schedules are daemon-owned and only run while `crona-daemon` is running.

## Tune the sound

You can adjust how notifications behave per alert type:
- **Toggles**: Enable/disable visual toasts, audio cues, or both.
- **Urgency**: Set notification priority levels for system backends.
- **Audible Alerts**: Choose from bundled royalty-free sound effects:
  - `chime`
  - `soft_bell`
  - `notification_ping`
  - `focus_gong`
  - `minimal_click`

## Platform support

Alert delivery depends on platform-specific helpers:

| OS | Notification Helper | Audio Player |
| --- | --- | --- |
| **macOS** | `terminal-notifier` (fallback: `osascript`) | `afplay` |
| **Linux** | `notify-send` | `paplay`, `aplay`, `play` (fallback: `canberra-gtk-play`) |
| **Windows** | `BurntToast` (fallback: PowerShell toasts) | PowerShell `SoundPlayer` |

You can check active backend capabilities (e.g., subtitle, urgency, icon support) in the **Alerts** panel.

The daemon must be running for scheduled reminders and day-boundary alerts. Next: [Focus Sessions](focus-sessions.md) for timer behavior.
