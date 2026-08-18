---
hosted: true
title: Features Overview
description: A high-level map of Crona's core workflows, views, and surfaces.
order: 4
---

Crona is a local-first work hub built around a daemon, terminal clients, and deterministic local state. The core surfaces are described below.

## Planning & Issue Lifecycle

Crona models work using repositories, streams, and issues:
- **Repositories**: Top-level logical boundaries (e.g., `office`, `OSS`).
- **Streams**: Long-lived project streams or branches (e.g., `feature-sync`, `v2-release`).
- **Issues**: Individual tasks with statuses (`backlog`, `planned`, `ready`, `in_progress`, `blocked`, `in_review`, `done`, `abandoned`).

The active context `{ repo -> stream -> issue }` remains synchronized across TUI and CLI clients.

## Summary and Daily Work

- **Summary**: A read-only TUI and CLI snapshot for a selected day or date range.
- **Daily**: The interactive planning and execution view for changing issues, habits, and check-ins.

## Focus Sessions & Stashes

Focused work intervals are executed from the issue itself:
- **Profiles**: Timers are configured per-issue as Pomodoro or no-break `Timer` sessions.
- **Countdown Timers**: No-break timers count down to a hard limit. Extending a countdown adds time without adding Pomodoro breaks or cycles.
- **Estimate Context**: Session start flows show worked time, estimate context, total duration, and an `Ends At` preview before focus begins.
- **Stashes**: If interrupted, you can stash active timer metadata and resume it later, preventing context loss.
- **Commit Summaries**: Focus sessions end with a commit message summarizing the work done.

## Habits & Momentum

Routines are tracked in parallel to issues:
- **Schedules**: Set daily, weekday, or weekly cadences.
- **History**: Completion logs are visualised on a calendar timeline.
- **Custom Momentum**: Configure naming, cadences, and target lists to track custom Momentum with protected rest, skipped buckets, and adjusted targets.

## Wellbeing & Heatmaps

Self-reported tracking helps prevent burnout:
- **Metrics**: Log daily energy, mood, sleep hours, sleep score, and screen time.
- **Heatmaps**: A terminal-based activity heatmap visualizes focus session density and habit compliance over a rolling metrics window.

## Outputs & Local Automation

Crona enforces local ownership of your metrics:
- **Narrative Reports**: Generate daily or weekly summaries using Handlebars templates.
- **Summary Dashboards**: Generate day or range summary exports as Markdown, PDF, or clipboard-ready Markdown.
- **PDF/CSV exports**: Compile structured sheets or PDFs (requires local renderers like `pandoc` or `weasyprint`).
- **Calendar (.ics) exports**: Generate deterministic files for native calendars.
- **Daemon Alerts**: Background alerts deliver scheduled notifications, focus inactivity warnings, update indicators, and export/support completion signals through local OS helpers.
