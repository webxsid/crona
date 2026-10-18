# TUI Insights tab: merge Rollup and Wellbeing into Review

Status: approved in chat 2026-10-18. Builds on `2026-10-17-tui-layout-design.md`, `2026-10-18-tui-key-hints-design.md` and `2026-10-18-tui-issues-tab-design.md`.

## Goal

Rollup (output: plan, work, where time went) and Wellbeing (state: mood, energy, sleep, risk) both look back over a window ending on the shared date, but keep separate windows, separate calendars, and separate per-day detail. Merge them into one sub-tab, **Review**, with one window, one day list, one per-day detail, and the focus and mood charts on the same time axis. Momentum stays its own sub-tab, unchanged.

Sub-tabs become `Review · Momentum`.

## Layout

Wide (≥160):

```
 Worked 17h   Focus 72  ▕bar▏  Plan 81%  done 22  failed 3
 Mood 3.8  Energy 3.4  Sleep 7.1h  Check-ins 12/14  Risk low
─ Days ─────────── 14 │ ─ Trends ──────── [S] time │ ─ Tue 14 Oct ──── [w] check-in
 Mon 13  ✓ 4/5  6h ▲  │ focus  ⣀⣤⣶⣿⣶⣤⣀            │ plan / work …
▶Tue 14  ✓ 3/3  5h ▬  │ mood   ⣀⣤⣤⣶⣤⣀⣀            │ check-in · risk · drivers
 Wed 15  ✗ 1/4  2h ▼  │ energy ⣀⣀⣤⣶⣶⣤⣀            │
```

- **Strip (2 rows, whole window)**: row 1 work — `Worked`, `Focus <score> <bar>`, `Plan %`, `done n  failed n`; row 2 wellbeing — `Mood`, `Energy`, `Sleep`, `Check-ins n/days`, `Risk`. Column groups drop right-first when narrow. Each row is exactly the content width.
- **Days pane (left, the only focusable pane)**: one two-cell-gap row per day in the window, newest last: date, status mark, done/planned, worked, mood mark (`▲ ▬ ▼`, blank without a check-in). The cursor is the shared date: moving it sets the date, and changing the date (date row, calendar) moves it. Title meta: day count.
- **Middle slot (wide)**: one of two modes, toggled by `[S]`:
  - **Trends** (default): focus-score, mood, energy and recovery charts, every chart spanning the window's days in the same order as the list.
  - **Time**: window summary, estimate bias, then top 3 by worked time for Repos, Streams, Issues, Segments (the current Rollup breakdown, borderless).
  - Title meta: `[S] time` in Trends mode, `[S] trends` in Time mode.
- **Detail (right, wide)**: the selected day. Daily's day-summary lines (plan and work), then the check-in, risk snapshot, and top risk/recovery drivers. Title = the date; title meta `[w] check-in  [W] away`.
- **Medium (100–159)**: Days │ secondary pane. The secondary pane is Detail or Time; `[S]` swaps them (title meta `[S] time` / `[S] day`). No Trends.
- **Narrow (<100)**: strip and Days only. `[enter]` opens the day detail as a dialog; `[S]` opens Time as a dialog.
- **Short**: charts fold first (Trends shows as many charts as fit, minimum one); below about 24 rows the middle slot drops and the layout follows medium.
- **Away dates**: a past date marked away shows the away screen, as Wellbeing does today.
- **Bottom row**: globals only, as on Daily.

## State and data

- One window: `reviewWindowDays` replaces `rollupWindowDays` and `wellbeingWindowDays`; default = the current rollup default. Momentum's window is untouched.
- `reviewShowTime bool`: the `[S]` mode, kept for the session (not reset on entering the tab).
- One Days cursor; Wellbeing's summary/trends/details pane cursors are removed.
- One reload path, `reloadReview`, on date or window change, loading every source over the same `[start, end]`: `LoadRollupSummaries`, `LoadRollupFocusWindow`, `LoadDashboardWindow`, `LoadWellbeingWindow`, `LoadMetricsRange`, `LoadMetricsRollup`, `LoadMetricsStreaks`. Existing `ContentState` and snapshot data fields stay; only their loading changes.
- No daemon changes.

## Keys (ViewReview)

- `[↑/↓]` move the Days cursor (sets the date).
- `[S]` toggle Trends/Time (wide), Detail/Time (medium); open Time as a dialog (narrow).
- `[enter]` narrow only: open the day-detail dialog. Not shown or honoured elsewhere.
- `[w]` open the day's check-in; `[W]` toggle away mode (from Wellbeing).
- Date row: `[E]` calendar, `[+/-/0]` window size, `,` `.` `g` date nav — via `keyregistry.RegisterRollup`/`RegisterWellbeing` repointed to `ViewReview`, with `handleShiftReviewWindow`/`handleResetReviewWindow`.
- `[E]` is the calendar on Review. Wellbeing's `[E]` daily export is dropped; export stays on Reports and the Issues tab.
- Every hint shown must be a key the input layer honours in that state; hints are whole (`FitHints`).

## Removed

- `ViewRollup`, `ViewWellbeing` and every reference (tabs, view-jump, `?` sheet, dispatch, model, selection, input).
- `views/rollup` and `views/wellbeing`. Breakdown, chart, check-in, risk and driver renderers move into `views/review` as borderless sections; the bordered panes, in-pane action lines, Rollup's day dialog, Wellbeing's cards and heatmap are deleted.
- `handleShiftRollupWindow`, `handleResetRollupWindow`, `handleShiftWellbeingWindow`, `handleResetWellbeingWindow`, and the two per-view reload paths.

## Out of scope

Momentum; daemon changes; Reports; new metrics.

## Testing

Strip is 2 rows, width-exact at 60/100/160; panes per width (wide/medium/narrow/short); `[S]` swaps the right slot per width and opens a dialog when narrow; `[S]` mode survives leaving and re-entering Review; title-meta hint matches the mode; Days cursor sets the date and date changes move the cursor; `[+/-/0]` changes the one window and reloads every source over the same range; away date shows the away screen; `[w]`/`[W]` work on Review; Rollup and Wellbeing absent from tabs, jump and help; bottom row globals-only; visual check at 60x20, 120x36, 200x50.
