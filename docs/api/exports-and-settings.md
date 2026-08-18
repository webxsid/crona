---
title: Exports and Settings
description: Export reports, calendar data, application settings, and away mode.
hosted: true
order: 7.5
---

# Exports and Settings

## Exports

| Area | Methods |
| --- | --- |
| Reports | `export.daily`, `export.glance`, `export.weekly`, `export.repo`, `export.stream`, `export.issue_rollup`, `export.csv`, `export.calendar` |
| Assets and directories | `export.assets.get`, `export.reports_dir.set`, `export.ics_dir.set` |
| Report files | `export.reports.list`, `export.reports.delete` |
| Templates | `export.template.reset`, `export.template.apply` |

Export methods return generated content or operation metadata depending on the method. File and directory methods affect local filesystem configuration and should be surfaced with clear success/error state.

## Settings

| Method | Purpose |
| --- | --- |
| `settings.get_all` | Read the complete settings snapshot. |
| `settings.get` | Read one setting or setting group. |
| `settings.patch` | Update selected fields. |
| `settings.put` | Replace a settings payload. |
| `settings.away_mode` | Read or change away-mode behavior. |

Settings changes may alter the logical day, notifications, or dashboard semantics. Refresh dependent views after a successful change and observe settings events.
