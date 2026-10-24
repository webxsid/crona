---
hosted: false
title: "Carry-over Spec"
description: "Proposed first-class carry-over for issues worked on but not finished, so each day keeps its own record of the issue."
order: 7.5
---

## Carry-over

Status: proposal. Not scheduled. This document describes what formal carry-over would be if built; it is not an implementation plan.

### Problem

An issue has one `TodoForDate`. When it is worked on but not finished, nothing changes at the day boundary: the next day it shows under **Overdue** (`dailyIssueOverdue`: `TodoForDate < today`). That is wrong in two ways:

- **The earlier day loses the issue.** Its daily view, summary, and plan no longer show that the issue was planned and worked on that day; the work only exists as sessions.
- **The next day treats it as a failure.** Overdue reads as "you didn't do this", when in fact you did what you could and the work simply spans days. Issues roll over all the time; it is normal, not exceptional.

### Goal

Make carry-over a recognized outcome. When an issue planned for a day is worked on but not finished:
- That day keeps its own copy of the issue, marked **carried over**, with the time worked that day.
- The next working day gets its own copy, marked **carried from** the earlier day, in the normal planned list, not Overdue.
- Scores, summaries, and reports treat it as progress, not as a miss.

An issue planned for a day and not touched at all stays what it is today: overdue and, in the plan, missed.

### Non-goals

- Splitting an issue into separate issues per day. It stays one issue with one status, estimate, and history.
- Changing what Overdue means for untouched issues.
- Carry-over for habits. Habits are per-day by nature.
- Automatic rescheduling to anything other than the next working day.

### Concepts

- **Day entry**: the plan's per-day record of an issue (`daily_plan_entries`, one row per plan date and issue). This spec uses it as the "copy" each day deserves; it already exists and already chains across dates (`BaselineDate`, `PostponeCount`).
- **Carry-over**: resolving a day entry as partially done and creating the next day's entry in the same chain.
- **Carry chain**: the sequence of day entries for one issue across consecutive carries.
- **Progress threshold**: the minimum time worked on a day for an unfinished issue to count as carried rather than missed.

### Data model

Plan entry additions:
- New entry status `carried`, alongside `planned`, `completed`, `failed`, `abandoned`, `canceled`.
- New event type `carried_over`.
- New fields on `DailyPlanEntry`:
  - `WorkedSeconds int`: time from sessions on that issue that started on the entry's date (computed, not stored separately; sessions remain the source of truth).
  - `CarriedFromEntryID *string` and `CarriedToEntryID *string`: links within the chain.
  - `CarryCount int`: consecutive carries in this chain so far.

Issue changes:
- `TodoForDate` moves to the next day's date on carry, exactly as a reschedule does today, so every existing view keyed on it keeps working.
- The issue keeps its own status through every carry (one issue, one status). `carried` is a display status only: clients show it on the issue's row for a past day that carried it away, while the day it was carried into shows the issue's real status.

### Which statuses carry

| Issue status | Carried at the boundary or on reschedule? |
|---|---|
| `planned`, `ready`, `in_progress`, `in_review` | Yes. The status is unchanged; the day carried from shows the row as `carried`. |
| `blocked` | No. It stays on its day and shows as Overdue; a manual carry is refused. |
| `backlog`, `done`, `abandoned` | No. They have no open plan entry to carry. |

No new tables. Day entries are the copies.

### When carry-over happens

At the day boundary (the existing `dayboundary` scheduler), for each `planned` entry on the closing date whose issue is not done or abandoned:

| Worked that day | Result |
|---|---|
| ≥ progress threshold | Entry → `carried`. A new `planned` entry is created on the next working day in the same chain, and `TodoForDate` moves to that date. |
| < threshold | Unchanged from today: the entry is missed, and the issue shows as Overdue. |

- **Next working day** skips away-mode dates, matching `protectedDelayDays`.
- **Progress threshold** is a core setting, `carry_over_min_minutes`, default 15, range 0–240. `0` means any session on that day counts, however short.
- If the daemon was not running at the boundary, carry-over runs for each missed date on startup, in date order, before the daily plan is served.
- Carry-over can be turned off with `carry_over_enabled` (default on). Off restores today's behavior exactly.

#### Manual carry-over

A user can also carry an issue explicitly during the day ("I'm done with this for today"):
- The entry becomes `carried` immediately with the time worked so far, and the next working day's entry is created.
- No threshold applies. Choosing to carry is the signal.
- This replaces "reschedule to tomorrow" for issues already worked on today. Reschedule stays for issues not started.

#### Undo

Carry-over goes through the existing pending-failure rollback window (`dailyPlanRollbackWindow`): within it, undoing restores the entry to `planned`, deletes the next day's entry, and moves `TodoForDate` back. After the window it is final.

### Scoring

A carry is progress, not a postponement:
- A carry does **not** increment `PostponeCount`.
- `BaselineDate` resets to the new day's date on carry, so delayed days start from zero for the new entry. The carried day was a planned day that got work, not a slip.
- A carried entry counts toward neither `FailedCount` nor `DelayedIssueCount`.
- The daily plan summary gains `CarriedCount` alongside `CompletedCount`, `FailedCount`, and `AbandonedCount`.
- A long chain can still signal a problem. Once `CarryCount` reaches a limit (`carry_over_stall_after`, default 5), the issue is flagged **stalled** in the UI, prompting the user to split, re-estimate, or abandon it.
- A stalled entry also carries a **stall component** in its `FailScore`, so it feeds the accountability score:
  `stall = 0.5 × log2(1 + max(0, CarryCount − carry_over_stall_after))`.
  It is zero when the issue first becomes stalled and grows slowly with each further carry. It does not feed backlog pressure, since nothing is delayed.
- `FailScore = delay + frequency + stall`. The Delay Penalty spec's pane and projections show the stall component as a third term when it is non-zero.

A day that ends with the issue untouched is missed exactly as today and is penalized through the existing delay rules (see the Delay Penalty spec).

### TUI

#### Daily view

- **Earlier day** (browsing a past date): carried issues appear in the planned list with a `↪` marker, the time worked that day, and the date they went to, for example `↪ Oct 25 · 1h20m`. They do not disappear.
- **Today**: issues carried in appear in the normal planned list, not Overdue, with a dim `↩ from Oct 24` marker. If the chain has been carried more than once, the marker shows the count, `↩ ×3`.
- Stalled issues show `stalled` in the warning style next to the marker.
- Overdue keeps showing only issues that were planned and not touched.

#### Issue details pane

A **Days** section lists the carry chain, one line per day entry, newest first:

```
Days
  Oct 25  planned    0m
  Oct 24  carried    1h20m
  Oct 23  carried    45m   (from Oct 22)
```

#### Actions

- One action on a selected issue in Daily or Issues: **Carry to next day**. It reuses the key currently bound to reschedule. On an issue already worked on today, that key carries over; on an untouched issue, it reschedules as today. The dialog title says which one is happening, so no new key is needed.
- Undo uses the existing rollback key within the window.

#### Day summary

The summary shows "Carried over: N" next to completed and failed, with the time worked on those issues included in the day's totals as it already is via sessions.

### Kernel and IPC

- `daily_plan.get` returns `carried` entries for their date and the new fields.
- A new operation, carry issue (issue id, date, optional target date), used by both manual carry and the boundary job.
- `issue.daily_summary` and `issue.today_summary` include carried counts and per-issue worked time for the day.
- `issue.detail` includes the carry chain.
- The boundary job is idempotent: running it twice for the same date does nothing the second time.

### CLI

- `crona issue carry <id>` carries an issue now.
- `crona plan show [date]` shows carried entries with their markers.
- JSON output includes the new entry fields.

### Reports and exports

- Daily reports list carried issues in their own group with time worked that day and the target date.
- Weekly reports show each issue once with its chain (`Oct 22 → Oct 25, 4 days, 3h10m`) and count carries separately from misses.
- CSV plan rows gain `carried_from`, `carried_to`, and `worked_seconds`.

### Migration

- No data migration is required. Existing entries keep their statuses.
- Past days that would have qualified as carries are not rewritten. Carry-over applies from the first day boundary after the feature ships.

