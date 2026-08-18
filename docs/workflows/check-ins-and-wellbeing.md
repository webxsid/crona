---
hosted: true
title: Check-Ins and Wellbeing
description: Log wellbeing metrics, inspect rolling trends, track accountability, and monitor burnout signals.
order: 5.6
---

Crona includes a local wellbeing reflection loop that tracks daily self-reported health signals alongside task execution statistics to help you monitor workload sustainability.

## Daily Check-Ins

A check-in is a daily self-report entry containing:
- **Mood**: Rating from `1` to `5`.
- **Energy**: Rating from `1` to `5`.
- **Sleep**: Total sleep hours and an optional sleep quality score.
- **Screen Time**: Optional tracking of daily computer screen time.
- **Notes**: Narrative logging for context on the day's performance.

To trigger the check-in dialog in either the **Daily** or **Wellbeing** views, press `w`.

## The Wellbeing Dashboard

The Wellbeing dashboard splits data display into two primary panes:
1. **Metrics Window**: A 7-day rolling window showing trends in mood, energy, sleep hours, screen time, focus duration, and habit completions.
2. **Momentum Pane**: A pane displaying custom habit and context Momentum, current versus best streak scores, protected days, adjusted targets, skipped buckets, and completion milestones. On wide terminals, this pane becomes independently scrollable.

## Burnout & Recovery Signals

Crona calculates burnout indicators using a localized heuristics engine. It processes:
- **Focus vs. Rest Ratio**: Compares active focus session durations against scheduled breaks and rest days.
- **Velocity Trends**: Evaluates issue completion volume over a rolling 7-day period.
- **Self-Reported Health**: Correlates mood and energy scores against work volumes.
- **Planning Accountability**: Tracks the ratio of completed plans to overdue, rolled-over, or abandoned issues.

Custom Momentum is rest-aware. Protected rest and away days can skip daily buckets, reduce weekly or monthly targets when real availability shrinks, and preserve continuity when a protected bucket should not count against the story. Away Today is reversible for the current logical date; after the day boundary, qualifying protected dates become stable historical records.

These signals are computed locally and visualized on the Wellbeing dashboard as trend indicators, helping you notice when workload patterns become unsustainable.

## Local & Private Storage

All check-in notes, ratings, and metrics are written directly to your local SQLite database. None of this data is sent to external services, maintaining complete privacy for your health and work logs.
