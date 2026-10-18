# TUI Layout Redesign

Status: draft for review · Branch: `brainstorm/tui-layout` · 2026-10-17

## Goal

Replace the web-style shell (22-column sidebar, two-line bordered header, bordered panes and cards, help bar) with a terminal-native layout that spends the screen on content and works at two real widths:

- **Companion**: a vertical split next to the editor, 25–33% of a large monitor, roughly 70–100 columns and tall.
- **Full**: its own window or tmux window on a 13" laptop, roughly 180–200+ columns.

Every existing capability stays reachable. Views may be merged or regrouped; functions may not be dropped.

### Success criteria

1. The app is fully usable at 80×24. The minimum drops from 135×30 to 60×20.
2. The daily loop (plan, start, note/capture, stop, review today) happens on one tab with no view switching.
3. Every current view is reachable from any tab by its tab number plus sub-tab cycling, or directly from the `:` palette; every current action keeps a key binding or a palette entry.
4. Charts degrade instead of disappearing at narrow widths, and any chart can be zoomed to the full content area.
5. The keymap layer can later take an opt-in vim profile without restructuring (the vim profile itself is out of scope).

## Usage that drives the design

| Frequency | Views |
| --- | --- |
| Daily | Daily, Summary, Issues, Active Session |
| Weekly | Rollup, Momentum, Wellbeing |
| Rarely | Ops, Session History (mostly editing), Support |
| Task-specific | Reports, Config, Settings, Alerts, Updates, Habit History |
| Automatic | Away (rest day or away mode) |

## Information architecture

Five tabs replace eighteen sidebar entries.

| Tab | Key | Contains | Replaces |
| --- | --- | --- | --- |
| Today | `1` | Session strip (when a timer runs), today's issues (Planned with pinned first / Overdue / Upcoming), due habits, the day's score and plan in the strip, `S` Day summary (detail pane, or a dialog when narrow), `C` scored calendar | Daily, Session Active, Away |
| Issues | `2` | Issues and Habits sub-tabs with shared repo/stream browse scope | Issues (Default), Meta |
| Insights | `3` | Sub-tabs: Rollup, Momentum, Wellbeing | Rollup, Momentum, Wellbeing |
| Log | `4` | Sub-tabs: Sessions, Habits, Ops | Session History, Habit History, Ops |
| System | `5` | Sub-tabs: Settings, Alerts, Export (reports + config), Updates, Support | Settings, Alerts, Reports, Config, Updates, Support |

Away keeps today's behaviour: it replaces the content area while protected rest is on. Its reachable set (Reports, History, Habit History, Settings) maps to `System › Export`, `Log › Sessions`, `Log › Habits`, `System › Settings`.

Sub-tabs within a tab are switched with `[` / `]`. This reuses today's view-cycling keys, which become sub-tab cycling.

### Today tab

- With **no timer**: an htop-style two-row **day strip** of aligned columns — Tasks over Habits meters (count, bar, worked/estimate), then columns pairing a check-in signal over a streak (Mood over Focus, Energy over Check-ins, then Sleep) when present. The shown week (weekday names over day numbers, selected day and today highlighted) is pinned to the strip's right end, directly under the sub-tab line's date navigation: `[g] today   [,] prev  <date> · W<nn>  next [.]`, which shortens to `[g]  [,] <date> · W<nn> [.]`, then the date and week, then the date alone when the line is narrow. When narrow, the rightmost columns drop first, then the week, then the remaining columns; rows drop from the bottom when short, and a one-row strip has no week.
- Tasks has three tabs (`←/→`): **Planned** — open pinned tasks first under a `Pinned N` heading (any date, dated ones tagged), then the day's tasks under `Planned N`; headings only appear when something is pinned — **Overdue** (open, dated before the shown day) and **Upcoming** (open, dated within the next 7 days, by date). At most 5 open tasks can be pinned (`MaxPinnedDailyIssues`, enforced by the daemon), so pinned keeps meaning priority.
- Tasks always renders as the two-line list (title and date, then context | worked/estimate | status) at every width; there is no table layout. A long title is truncated before its date.
- Below the strip, **Tasks** and **Habits** each hold half the height, so neither moves as the other's content changes. Habits renders one htop-style meter row per habit at every width, the same whether focused or not: checkbox and name, status, a bar of logged minutes against the target, and `logged / target`. The bar, then the status, drop when narrow. Habit rows carry no momentum data — habits and momentums are many-to-many with mixed cadences.
- On wide screens a **detail pane** sits right of the lists (list │ detail, weights 3:2). It is titled with the selected task's (or habit's, when Habits is focused) name and shows, in order: the description, the fields flowed into up to three columns as width allows, notes, then the momentums it feeds — habit momentums for a habit, and for a task the context momentums whose repo (and stream, if set) match the task's, as the daemon counts sessions. Each momentum line shows cadence, current period progress (time for context momentums), ladder, and current/best streak. The detail pane holds no calendar; the week lives in the strip. On medium and narrow screens the detail pane is hidden; `enter` opens details as today.
- Pattern taken from htop/k9s (header strip over lists) and lazygit/Posting (list │ detail), not peer columns.
- With an **active timer**: a session strip pins to the top of Today. It shows the issue, segment, elapsed/target, and progress bar, plus the session keys `p` pause, `r` resume, `x` stop and `z` advance. Below it, the issue and habit lists stay visible and navigable.
- **Allowed during a timer** (a change from today's lock-out): `n` adds a note to the active session (existing `amend_session` dialog); `a` quick-captures a new issue (existing create-issue dialog, without starting it). Starting another issue with `f` keeps the existing `issue_session_transition` flow.
- **Still locked during a timer**: tabs other than Today and Log, as today. Insights, Issues and System open after the session ends.

### Issues tab

- Sub-tabs are **Issues · Habits**; Meta is removed. Both share browse scope independent of checkout, reset to checkout on entering Issues.
- A two-row summary strip counts the browse scope. At ≥160 columns, Scope, list and detail sit side by side; medium drops detail; below 100 columns Scope becomes one focusable row.
- Scope lists All, repos and streams, with counts and checkout dots. Moving browses; `c` checks out. Scope offers create, open, edit and guarded delete keys inside its focused pane.
- `tab` cycles Scope/list focus. In Issues, `←/→` switches Open/Done. Habits lists all definitions in scope. Create defaults follow browse scope.
- Selected-card actions and create hints live inside focused panes; the bottom row contains globals only, with no duplicate create/status/checkout keys or day-summary key. The `?` sheet lists scope and list keys under Current View.
- Detailed requirements: [Issues tab design](2026-10-18-tui-issues-tab-design.md).

## Shell (chrome)

```
 1 Today  2 Issues  3 Insights  4 Log  5 System               repo › stream
 ── <content: section headers + rules, no outer borders> ─────────────────
 ...
 FOCUS 18:42 ▸ crona/tui   ? help  : palette                  ✓ saved
```

- **Top line (1 row)**: tabs left; active repo › stream right. Dev/beta markers and version move into the `?` help overlay and the System tab.
- **Sub-tab line (1 row, only when the tab has sub-tabs)**: `Rollup · Momentum · Wellbeing`, with the active one highlighted. On every dated view (Daily, Rollup, Momentum, Wellbeing) its right end carries the shared date's navigation, `[g] today   [,] prev  <date> · W<nn>  next [.]`; Rollup puts its window start before the date (`<start> → <date>`). `,`/`.`/`g` move one shared date on all of these views; Rollup, Momentum and Wellbeing size their window with `+`/`-`/`0` (see `2026-10-18-tui-unified-date-design.md`).
- **Content**: no outer borders. Sections are titled with a rule line `─ TITLE ───── right-aligned meta`. The focused section's title is highlighted in place of today's cyan border.
- **Bottom status line (1 row)**: timer state (always visible when a timer runs, on every tab), global keys, `[?] all keys`, and the status toast rendered inline on the right, replacing the floating toast overlay. On Daily it holds globals only (`[tab]`, `[c]`, `[E]`, `[w]`, `[W]`); other views keep their section keys minus the date keys. Tab and sub-tab keys live on the `?` sheet (see `2026-10-18-tui-key-hints-design.md`).
- **Dialogs**: unchanged centred modals, but width follows `min(content, maxWidth)` so they fit 60 columns.
- **Minimum size**: 60×20. Below that, the existing "terminal too small" message.

Chrome costs 2–3 rows and 0 columns, against today's 4 rows and 22 columns.

## Responsive layout

Width classes, computed once per frame:

| Class | Width | Behaviour |
| --- | --- | --- |
| Narrow | < 100 | Sections stack vertically, one column |
| Medium | 100–159 | One primary column; strips and short sections stack around it |
| Wide | ≥ 160 | Primary list │ detail pane for the selection (never peer lists side by side) |

Rules: sections get stable heights — they do not resize as their content changes (Today splits Tasks/Habits evenly). Columns are only used for list │ detail. Key hints for the focused section live in the status line, except on Daily, whose Tasks and Habits panes show their own keys (a create line, and an action line on the focused selection) and whose detail pane title offers `[S]`.

Each tab declares its sections and a per-class arrangement. A new `layout/grid` helper turns an arrangement plus the content box into one `Box{W,H}` per section. Sections never compute their own position.

## Visuals

Each visual implements `Render(box Box) []string` and picks the richest tier that fits:

| Tier | Minimum box | Example |
| --- | --- | --- |
| Full | per chart, e.g. focus graph ≥ 60×10 | today's axes/ticks chart |
| Compact | ≥ 30×3 | `focus ▁▃▅▇▆▃▂  avg 3h12m  max 5h` |
| Glyph | 1 row | `focus 3h12m ▲12%` |

- **Inline visuals** (streak ladders `▰▱`, progress bars, habit bars, sparklines) stay row-sized and appear inside lists and the status line.
- **Block visuals** (rollup focus graph, momentum chart, wellbeing heatmap, distribution bars) live in Insights and implement all three tiers. This replaces today's ad hoc `ViewSizeCompact` and the "No focus data" cut-off below 24 columns.
- **Zoom**: `Z` on a focused chart section expands it to the full content box, and `esc` restores it. `z` stays bound to session advance.

## Navigation and keymap

| Key | Action | Change from today |
| --- | --- | --- |
| `1`–`5` | Switch tab | New global; previously per-view section/pane jumps |
| `[` / `]` | Previous/next sub-tab | Was view cycling |
| `tab` / `shift+tab` | Next/previous section in tab | Same idea, renamed from pane |
| `:` | Command palette: every tab, sub-tab and global action, fuzzy-matched | Replaces `v` view-jump dialog (`v` kept as an alias) |
| `Z` | Zoom focused chart | New |
| `n` | Note on active session | New during a timer |
| `a` | Create / quick-capture issue | Now allowed during a timer on Today |

Per-view digit bindings move off digits:

- Daily `1`/`2` (issues/habits panes) → `tab`.
- Issues `1`/`2` (open/completed) → `←/→` on the list; `tab` cycles Scope/list focus.
- Meta `1`–`4` → Scope/list focus with `tab`; habits move to the Habits sub-tab.
- Config, Settings, Alerts, Reports, Scratch `1` (focus the only pane) → dropped as redundant.

Keymap profiles: `key_registry` gains a `Profile` (default only, for now) that maps logical actions (`nav.down`, `tab.next`, `section.next`, …) to keys. Handlers register against actions, not raw keys, so a later vim profile only adds a mapping table.

## Architecture changes

- `state`: replace `View`/`Pane` enums with `Tab`, `SubTab` and `Section` ids. Keep the old values as aliases during migration so `dispatch`, `selection` and the dialogs keep compiling.
- `layout/render.go`: new shell (tab line, sub-tab line, status line, inline toast). The sidebar, header and help bar are removed. `MinWidth`/`MinHeight` → 60/20. `SidebarWidth` is removed.
- `layout/grid` (new): width classes and section box allocation.
- `views/*`: each existing view becomes one or more sections taking a `Box`. Bordered `PaneBase` chrome is replaced by a rule-line section header. The unused `Pane.HandleKey`/`Focus`/`Resize` surface in `views/ui` is removed.
- `views/charts` (new home): tiered visuals extracted from `rollup`, `momentumview`, `wellbeing` and `daily`.
- `input` + `key_registry`: tab/sub-tab/section routing, profile indirection, relaxed timer lock for `n` and `a` on Today.
- `model`: navigation state becomes `tab`, `subTab[tab]`, `section`, with cursors/filters keyed by section. Timer-lock logic moves from view lists to a tab allow-list.

Untouched: the kernel, the API client, `commands`, the `dispatch` message handling semantics, and the dialog controller and its kinds.

## Migration

The shell and views can change in independent steps, each releasable:

1. Shell swap: tab line + status line + inline toast, sidebar removed, existing views mapped onto tabs/sub-tabs as-is; lower the minimum size.
2. Responsive grid and borderless section headers across views.
3. Today merge (Daily + Summary + Session strip) and the timer-lock relaxation.
4. Issues scope tree and Habits sub-tab (Meta removed).
5. Tiered charts and zoom.
6. Command palette and keymap profile indirection.

## Testing

- Golden-frame tests: render each tab at 80×24, 120×36 and 200×50, with and without an active timer, in protected mode, and with a dialog open; compare to checked-in snapshots.
- Keymap tests: every action listed in today's `docs/reference/tui-keymap-reference.md` resolves to a handler in the new layout (the guard for "no functionality lost").
- Chart tier tests: each visual at boxes straddling its tier thresholds.
- Existing `model`, `input` and `dispatch` tests keep passing through the alias period.

## Docs

Update `docs/reference/tui-keymap-reference.md` and the TUI screenshots in `docs/screenshots` in the same release as step 1.

## Out of scope

- The vim keymap profile (only the indirection it needs).
- Theme/colour changes.
- Any kernel or API changes.
