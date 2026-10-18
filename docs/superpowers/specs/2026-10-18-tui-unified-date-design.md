# TUI Unified Date — Design

**Status:** approved in chat, 2026-10-18. Branch: `brainstorm/tui-layout`.

## Problem

The TUI tracks five independent dates — `dashboardDate` (Daily), `summaryDate`, `momentumDate`, `wellbeingDate`, and a free `rollupStartDate`/`rollupEndDate` pair — and three notions of "today":
- `time.Now()` fallbacks in every `current*Date()` accessor and in startup loads.
- The daemon's logical date (`health.CurrentDate`), used for rest mode, dialogs and due-date defaults.
- An empty string, which means "today" for some fields but not others.

Moving Daily to another day leaves Summary, Momentum (including the momentum lines in Daily's detail pane) and Wellbeing on their own dates. Midnight rollover patches each date separately.

## Design

1. **One date.** The model holds a single `date` string. `currentDate()` returns it, or the daemon's logical date when unset. "Today" means the daemon's logical date everywhere. `time.Now()` is only the fallback when no health has loaded yet.
2. **Windows stay per view; the anchor is shared.**
   - Daily and Summary show `date` directly.
   - Momentum, Wellbeing and Rollup end their range on `date`.
   - Each of those three keeps its own window length:
     - Momentum: 7/14/30/90, default 30.
     - Wellbeing: 7/14/21/30, default 7.
     - Rollup: new, 7/14/30/90, default 7.
   - Rollup's range is `[date − (window − 1), date]`, so arbitrary start dates go away.
3. **Same keys everywhere.**
   - On Daily, Summary, Momentum, Wellbeing and Rollup, `,`/`.` shift the shared date and `g` resets it to today.
   - Rollup's window uses `+`/`-`/`0`, like Momentum and Wellbeing.
   - Rollup's old range keys go: ←/→ (start), `S` (start-date dialog) and the separate end keys.
   - Rollup's `E` opens the date picker for the shared date.
   - Summary's `c` calendar sets the shared date.
4. **Loads.** Changing the date reloads every date-bound surface at once: Daily summary, plan, due habits, streaks, summary snapshot, momentum range, wellbeing window and rollup. They are local IPC calls; eager reload avoids per-surface staleness tracking.
5. **Midnight rollover.** When the logical date changes, the shared date follows only if it was on the previous today. A user looking at another day stays there. All surfaces reload.
6. **Visible everywhere.** The sub-tab line's date navigation (`[g] today   [,] prev  <date> · W<nn>  next [.]`) shows on every dated view. Rollup also shows its range.

## State plumbing

- `input.State`, `dispatch.MessageState` and `dialog_runtime` state replace the five date fields with `Date` (+ `RollupWindowDays`).
- `views/types.ContentState` keeps its per-view fields (`DashboardDate`, `SummaryDate`, `MomentumDate`, `WellbeingDate`, `RollupStartDate`, `RollupEndDate`) to limit view churn. The model fills them all from the one date (Rollup's start is derived).

## Out of scope

- Per-view memory of different dates.
- Changing the windows' preset values for Momentum and Wellbeing.
