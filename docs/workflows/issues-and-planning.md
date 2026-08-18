---
hosted: true
title: Issues and Planning
description: Structure your backlog, define issue transitions, and manage daily planning contexts.
order: 5.1
---

Crona structures daily planning around issues instead of a flat list, separating task capture from focused execution.

## The Issue Lifecycle

Issues transition through explicit states. Status transitions are validated by the daemon to keep execution histories consistent:

- **Backlog**: Tasks captured but not yet scheduled.
- **Planned**: Scheduled for a specific date or period.
- **Ready**: Unblocked and ready for immediate work.
- **In Progress**: Actively timed. Starting a focus session automatically promotes an issue to this state.
- **Blocked**: Temporarily suspended by external dependencies.
- **In Review**: Technical implementation complete; awaiting review.
- **Done**: Fully completed.
- **Abandoned**: Intentionally dropped.

## Managing Planning Dates

Issues can carry an optional planning date (`todo_date`). Assigning a planning date:
- Automatically promotes an issue from **Backlog** to **Planned**.
- Registers it in the **Daily View** for the target date.
- Promotes overdue issues to the dedicated overdue section if left incomplete.

## Choosing Active Context

The checked-out context `{ repo -> stream -> issue }` acts as a working pointer. When you check out a context:
- CLI commands reuse it, reducing the need to specify target IDs.
- The TUI updates header navigation and focuses the corresponding active issue.
- Focus sessions and reports default to target the active context.

## Daily View Structure

The Daily dashboard displays:
- **Planned**: Issues assigned to the active date.
- **Pinned**: Important items pinned to the dashboard for ongoing visibility.
- **Overdue**: Open issues from previous dates that require resolution.
- **Habits**: Recurring routines due on the active date.
