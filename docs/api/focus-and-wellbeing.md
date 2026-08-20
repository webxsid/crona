---
title: Focus and Wellbeing
description: Sessions, timers, context, check-ins, momentum, dashboards, and metrics.
hosted: true
badge: Beta
order: 7.4
---

# Focus and Wellbeing

:::note[Protocol compatibility]
The focus-score dashboard methods and logical-date synchronization require protocol `1.5`. Read `kernel.info.get` before using `dashboard.focus_score`, `dashboard.focus_score_range`, or relying on configured Start of Day semantics across clients.
:::

## Sessions and timers

Sessions: `session.list_by_issue`, `session.get`, `session.detail`, `session.get_active`, `session.start`, `session.pause`, `session.resume`, `session.end`, `session.log_manual`, `session.amend_note`, and `session.history`.

Timers: `timer.get_state`, `timer.start`, `timer.activity.touch`, `timer.pause`, `timer.resume`, `timer.advance`, `timer.extend`, `timer.defer_break`, and `timer.end`.

Timer mutations may emit session and timer events. Treat the daemon response as the immediate source of truth, then apply later events in order.

## Context

| Method | Purpose |
| --- | --- |
| `context.get` | Read active repository, stream, and issue context. |
| `context.set` | Set a complete context. |
| `context.switch_repo` / `context.switch_stream` / `context.switch_issue` | Change one context dimension. |
| `context.clear_issue` / `context.clear` | Clear issue context or all context. |

## Check-ins and momentum

Check-ins use `checkin.get`, `checkin.upsert`, `checkin.delete`, and `checkin.range`. Momentum uses `momentum.list`, `momentum.create`, `momentum.update`, `momentum.delete`, `momentum.range`, and `momentum.detail`.

## Metrics and dashboards

Use `metrics.range`, `metrics.rollup`, `metrics.streaks`, and `metrics.streaks_lifetime` for metric data. Dashboard views are provided by `dashboard.window`, `dashboard.focus_score`, `dashboard.focus_score_range`, `dashboard.distribution`, and `dashboard.goal_progress`.

Date ranges should use the logical date defined by Crona’s configured start-of-day boundary. Do not independently roll a view over at local calendar midnight.
