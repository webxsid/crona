---
hosted: true
title: Check-Ins and Wellbeing
description: Log wellbeing metrics, inspect rolling trends, track accountability, and monitor burnout signals.
order: 5.6
---

Some days are productive; some days are simply heavy. Check-ins give that context a place in Crona, next to the work you actually did.

## Daily Check-Ins

A check-in is a daily self-report entry containing:
- **Mood**: Rating from `1` to `5`.
- **Energy**: Rating from `1` to `5`.
- **Sleep**: Total sleep hours and an optional sleep quality score.
- **Screen Time**: Optional tracking of daily computer screen time.
- **Notes**: Narrative logging for context on the day's performance.

Open the check-in dialog from **Daily** or **Wellbeing** with `w`. Add only what is useful today; an empty day is valid.

## The Wellbeing Dashboard

The Wellbeing dashboard splits data display into two primary panes:
1. **Metrics Window**: A 7-day rolling window showing trends in mood, energy, sleep hours, screen time, focus duration, and habit completions.
2. **Momentum Pane**: A pane displaying custom habit and context Momentum, current versus best streak scores, protected days, adjusted targets, skipped buckets, and completion milestones. On wide terminals, this pane becomes independently scrollable.

## Read the pattern, not a verdict

Crona calculates local workload signals from:
- **Focus vs. Rest Ratio**: Compares active focus session durations against scheduled breaks and rest days.
- **Velocity Trends**: Evaluates issue completion volume over a rolling 7-day period.
- **Self-Reported Health**: Correlates mood and energy scores against work volumes.
- **Planning Accountability**: Tracks the ratio of completed plans to overdue, rolled-over, or abandoned issues.

Custom Momentum is rest-aware. Protected rest and away days can skip daily buckets, reduce weekly or monthly targets when real availability shrinks, and preserve continuity when a protected bucket should not count against the story. Away Today is reversible for the current logical date; after the day boundary, qualifying protected dates become stable historical records.

These are prompts for noticing a pattern, not a diagnosis. Use the trend indicators to decide whether tomorrow needs less work, more rest, or a different plan.

## Local & Private Storage

Check-in notes and ratings stay in the local SQLite database. They are not sent to an external service.

Next: [Habits](habits.md) for recurring routines, or [Focus Sessions](focus-sessions.md) to connect reflection with focused work.
