---
hosted: true
title: Habits
description: Track recurring routines, schedules, and custom streak metrics alongside issues.
order: 5.5
---

Habits in Crona are recurring routines tracked independently of issues. They represent repeatable processes, wellness metrics, or daily routines.

## Habit Schedules

Crona supports three cadences:
- **Daily**: Due every day.
- **Weekdays**: Due Monday through Friday.
- **Weekly**: Due once per week, tracked within the configured week boundary.

## Logging & Completions

Habit completions are recorded against calendar days:
- Log completions directly in the Daily dashboard.
- Toggle completions backward or forward in time to keep logs accurate.
- Habit history is stored in the local SQLite database and can be queried or exported.

## Custom Streaks and Momentum

Crona tracks streaks using customizable rules configured in Settings. These definitions can target habits or contexts (specific repository and stream combinations):
- **Cadence**: Daily, weekly, or monthly completion checks.
- **Matching Mode**:
  - `any`: Streak continues if any target meets its completion threshold.
  - `all`: All targets must meet the threshold to maintain the streak.
- **Milestones**: Streaks are visualized via progress ladders on the Wellbeing dashboard.
- **Grace Periods**: Weekly and monthly streaks do not break immediately when a new period starts; they remain valid until the period expires and the threshold fails to be met.

## Protected Days and Away Today

Protected rest rules and away days keep a completed Momentum streak continuous rather than treating time away as a miss. Turning on **Away Today** protects the current logical date immediately. Turn it off again on the same date and that provisional protection disappears; it does not create a historical away record. Once a protected day has passed the configured day boundary, Crona records it as immutable history.
