---
hosted: true
title: Habits
description: Track recurring routines, schedules, and custom streak metrics alongside issues.
order: 5.5
---

Use habits for the things you want to remember without turning them into project work: stretch, review your inbox, take a walk, or write a daily note.

Create them in the TUI's Habits view, then let the Daily view show what is due alongside your issues.

## Choose a cadence

Crona supports three cadences:
- **Daily**: Due every day.
- **Weekdays**: Due Monday through Friday.
- **Weekly**: Due once per week, tracked within the configured week boundary.

## Mark a habit complete

Habit completions are recorded against calendar days:
- Log completions directly in the Daily dashboard.
- Toggle completions backward or forward in time to keep logs accurate.
- Habit history is stored in the local SQLite database and can be queried or exported.

## Keep the useful streak

Crona can turn completions into Momentum rules configured in Settings. A rule can target habits or a context (a repository and stream):
- **Cadence**: Daily, weekly, or monthly completion checks.
- **Matching Mode**:
  - `any`: Streak continues if any target meets its completion threshold.
  - `all`: All targets must meet the threshold to maintain the streak.
- **Milestones**: Streaks are visualized via progress ladders on the Wellbeing dashboard.
- **Grace Periods**: Weekly and monthly streaks do not break immediately when a new period starts; they remain valid until the period expires and the threshold fails to be met.

## Take a day off

Protected rest rules and **Away Today** keep time away from breaking a Momentum story. Away Today is reversible until the day boundary; after that, a qualifying protected day becomes history.

Next: [Check-ins and Wellbeing](check-ins-and-wellbeing.md) if you want to pair routines with a daily reflection.
