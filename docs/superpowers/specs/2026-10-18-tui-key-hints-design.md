# Key Hints Where They Apply — Design

**Status:** approved in chat, 2026-10-18. Branch: `brainstorm/tui-layout`.

## Problem

Every key hint sits in one bottom line: `[key] long label` runs with the same spacing, so it isn't scannable. Long labels push keys off the end with a mid-word `…`. Globals (`[1-5] tabs`, `[[/]] sub-tabs`, `[f9]`, `[tab]`, `[c]`) repeat on every screen and drown out context changes.

## Design

Hints move to where they act.

1. **Date row.** Dated views (Daily, Rollup, Momentum, Wellbeing) get a header row under the sub-tab line carrying the shared date navigation: `[g] today   [,] prev  <date> · W<nn>  next [.]` (Rollup with its range start). The view's own date keys sit at its right end: Daily `[C] calendar`; Rollup `[E] calendar  [+/-/0] window`; Momentum and Wellbeing `[+/-/0] window`. On Daily without a detail pane, `[S] day` joins them. The sub-tab line no longer carries the date.
2. **Numbered tabs.** The tab line reads `[1] Today  [2] Issues …`. The sub-tab line is framed by its keys: `[[] Daily · Session · Away []]`.
3. **Tasks pane (Daily).**
   - A separate `[a] new task` line sits under the Planned / Overdue / Upcoming selector.
   - The focused, selected task card replaces its meta line with its actions: `[enter] open  [s] status  [d] due  [P] pin`, then, when the timer is idle, `[f] focus  [A] ad hoc  [m] log  [e] edit  [D] delete`. Only whole hints are shown, as many as fit.
4. **Habits pane (Daily).** A separate `[a] new habit` line leads the list. Habit rows are single-line, so the focused, selected habit gets one action line under it: `[enter] open  [x] done  [F] fail  [m] log  [e] edit  [D] delete`, plus `[f] focus` when the timer is idle.
5. **Detail pane.** Its title line carries `[S] day`, or `[S] selection` while the Day summary shows.
6. **Bottom row.** It keeps a few globals: `[tab] next pane`, `[c] context`, `[w] check-in`, `[W] away`, `[E] export` where they apply, then `[?] all keys  ctrl+c quit` at the right. Status messages still take the right end. Tab digits, `[`/`]`, `[f9]` and similar move to the `?` sheet.
   - On Daily, the bottom row has globals only.
   - Other views keep their view keys there this pass, minus the date and window keys, which now sit on the date row.

## Out of scope

- Moving other views' keys into their panes (a later pass).
- Rewording labels outside Daily.
