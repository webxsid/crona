---
hosted: false
title: "Delay Penalty Spec"
description: "Proposed surfacing of each issue's delay penalty, and its effect on today's scores, in the issue details pane."
order: 7.4
---

## Delay Penalty

Status: proposal. Not scheduled. This document describes what surfacing the delay penalty would look like if built; it is not an implementation plan.

### Goal

Delaying a planned issue already carries a penalty, but it is only visible as aggregate numbers in the day summary. Each issue's details pane should show:
- what penalty the issue carries right now,
- why (how long and how often it has been delayed),
- how much it is moving the scores today,
- and what would change it.

The aim is that a user looking at one issue can tell whether it is worth doing, re-planning, or abandoning today.

### Non-goals

- Changing the penalty formula. This spec only surfaces it.
- Feeding the penalty into the focus score or the burnout and recovery indicator. They stay as they are; product wording should say the penalty affects accountability and backlog pressure.
- Penalties for habits. Habit misses already have their own streak and failure signals.
- Notifications or nudges about penalties. A future rules engine or the morning brief may use the same data.

### How the penalty works today

Computed in the kernel's daily plan (`applyDailyPlanScore`, `buildDailyPlanSummary`):

- Each plan entry tracks `BaselineDate` (the first date it was planned for), `CurrentPlannedDate`, `PostponeCount`, `CurrentDelayedDays`, and `MaxDelayedDays`.
- **Delayed days** count calendar days after the baseline up to the current planned date, skipping away-mode dates (`protectedDelayDays`).
- **Per-issue penalty** (`FailScore`):
  `delay = log2(1 + delayedDays)`, `frequency = log2(1 + max(0, postpones − 1))`, `FailScore = delay + 0.5 × frequency`.
- An entry with `FailScore ≥ 2.5` is **high risk**.
- **Day-level effects**, summed over active entries:
  - `AccountabilityScore` = sum of `FailScore`.
  - `BacklogPressure` = sum of `log2(1 + delayedDays)`.
  - `DelayedIssueCount`, `HighRiskIssueCount`, `AvgDelayDays`, `MaxDelayDays`.
- The dashboard window averages `AccountabilityScore` across days.

The focus score (`ComputeFocusScoreSummary`) and the burnout and recovery indicator (`computeBurnout`) do **not** read any of these values, and this spec keeps it that way. The pane shows only the effects that exist and never claims a contribution that isn't computed.

### High-risk threshold setting

The fixed `dailyPlanHighRiskThreshold` (2.5) becomes a core setting, `plan_high_risk_threshold`:
- Default 2.5, so existing behavior is unchanged.
- Allowed range 0.5–10, step 0.1, edited in the existing Settings view next to the other planning settings.
- `HighRiskIssueCount`, the `highRisk` flag, row markers, and `--high-risk` all read the setting.
- Changing it applies to live and historical views alike; stored plan entries keep their raw `FailScore`, so nothing is recomputed or migrated.

### Kernel and IPC

The kernel computes everything shown in the pane, so the TUI does no scoring math.

Add an issue-level penalty view, returned by `issue.get`/detail (or a new `issue.penalty.get` if detail payloads should stay lean):

```
IssuePenalty {
  planned            bool      // has an active plan entry
  baselineDate       string
  currentPlannedDate string
  delayedDays        int
  maxDelayedDays     int
  postponeCount      int
  protectedDaysSkipped int     // away-mode days not counted
  delayComponent     float64   // log2(1 + delayedDays)
  frequencyComponent float64   // 0.5 × log2(1 + max(0, postpones − 1))
  stallComponent     float64   // from the Carry-over spec; 0 unless the carry chain is stalled
  failScore          float64
  highRisk           bool
  highRiskThreshold  float64   // current setting value, so clients don't hardcode it
  historical         bool      // issue is done or abandoned; values frozen at resolution
  todayContribution {
    accountability   float64   // this issue's share of today's AccountabilityScore
    accountabilityPct float64  // share of today's total, 0–1
    backlogPressure  float64
    backlogPressurePct float64
  }
  ifDoneToday {               // effect of completing it today
    accountabilityDelta  float64
    backlogPressureDelta float64
  }
  ifPostponedOneDay {         // effect of pushing it to tomorrow
    failScore            float64
    accountabilityDelta  float64
    becomesHighRisk      bool
  }
}
```

Rules:
- The formula lives in one place. The projections reuse `applyDailyPlanScore` on a copy of the entry; they never reimplement it.
- An issue with no active plan entry returns `planned: false` and no other fields.
- Done or abandoned issues return the final values at resolution time with `historical: true`. `todayContribution`, `ifDoneToday`, and `ifPostponedOneDay` are omitted for them.
- Values are rounded the same way reports round them (`roundReportMetric`).
- The penalty view updates on the same events that already refresh the daily plan, so the pane stays live.

### TUI: issue details pane

A **Delay penalty** section in the issue details pane, shown only when the issue has an active plan entry and some penalty (delayed days > 0 or postpones > 1). Otherwise one dim line: `On schedule · no penalty`.

```
Delay penalty                                    ▲ high risk
  Planned for   Oct 18 → Oct 24  (5 days late, 3 postpones, 1 away day skipped)
  Penalty       3.38  = 2.58 delay + 0.79 frequency          threshold 2.50
  Today         +3.38 accountability  (54% of today's 6.30)
                +2.58 backlog pressure (38% of today's 6.80)
  Done today    accountability −3.38 · backlog −2.58
  Push 1 day    penalty 3.81 (+0.43)
```

- **Planned for**: baseline → current planned date, delayed days, postpone count, and away days skipped if any.
- **Penalty**: total with its two components and the high-risk threshold. The `▲ high risk` marker uses the existing warning style when `highRisk` is true.
- **Today**: this issue's contribution to today's day-level scores and its share of each total, so a user can see which issues dominate.
- **Done today / Push 1 day**: the two projections, so the cost of postponing is visible before doing it.
- Narrow panes keep the Planned for and Penalty lines and fold the rest behind the existing expand behavior of the pane.
- Done or abandoned issues show the section titled **Delay penalty (final)** in the muted style, with only the Planned for and Penalty lines as they stood at resolution. If the issue resolved with no penalty, the section is omitted.
- No new keys. The section is part of the existing details pane.

### TUI: other surfaces (light touch)

- Issue rows on Issues and Daily show a small penalty marker when the issue is high risk (`▲2.6`), using the same style as the pane.
- When re-planning or postponing an issue through the existing flows, the confirm dialog shows the "Push 1 day" line, so the cost is visible at the moment of the decision.
- The day summary keeps its aggregate numbers and adds "top contributors": up to three issues with the largest share of today's accountability score.

### CLI

- `crona issue show <id>` prints the same Delay penalty block.
- `crona issue ls --high-risk` lists issues over the threshold, sorted by penalty.
- JSON output includes the `IssuePenalty` object as is.

### Reports

Daily and weekly reports already include `planFailScore` per issue and day-level `backlogPressure`. They gain the delayed days, postpone count, and today's share of the accountability score per issue, using the same fields as the pane.

