# TUI Insights · Momentum redesign

Date: 2026-10-18
Branch: `brainstorm/tui-layout`

## Goal

Bring the Momentum sub-tab in line with Review: a borderless strip, a focused list on the
left and a detail pane on the right. Drop the bordered pane, the nested card boxes and the
duplicated header. Inactive momentums stay visible but dimmed, and each row carries a small
sparkline so the recent trend reads at a glance.

## Layout

Strip (2 rows), then `List │ Detail`.

| Width class | Panes |
|---|---|
| wide (≥160) | List │ Detail |
| medium (≥100) | List │ Detail |
| narrow | List only; `[enter]` opens the details dialog |

The list pane is a fixed width (`listWidth`, 40 cells). The detail pane takes the rest.
Both are `viewchrome.RenderSection` slots joined with `viewchrome.JoinColumns`.

### Strip

Two rows at exact width. Column groups drop from the right when they don't fit, as on
Review's `StripRows`.

- Row 1: `Active 8  inactive 2` · `Latest  met 5  near 2  missed 1`
- Row 2: `Best streak  <name> 27d` (longest current streak) · `Longest ever  <name> 40d`

"Latest" counts only active momentums, using the status of each one's latest bucket.
Skipped buckets are not counted. Any missing value renders as `–`.

### List pane

- Title `Momentums`, meta is the count, and the pane is focused.
- First line: `[a] new momentum`.
- One row per momentum: `▶ Name   ▁▃▅▇▅▂▁▃  27d ✓`
  - The name is truncated to fit.
  - The sparkline covers the last 8 buckets. Each cell is `▁`–`█`, scaled by
    `min(count/target, 1)`, and coloured by that bucket's status. A skipped bucket is
    shown as a dim `·`. With fewer than 8 buckets, the sparkline is left-padded with spaces.
  - Then the current streak in days (`27d`).
  - Then the status mark for the latest bucket: `✓` met, `◐` near, `✗` missed,
    `·` skipped.
- Active momentums come first, in their existing order. Inactive momentums follow, with
  the whole row in `StyleDim` and the mark `paused`.
- `↑/↓` move the selection, and the list scrolls with `viewchrome.ListWindow`.
- Empty state: `No momentums yet  [a] new momentum`.

Statuses come from the existing `momentumStatusForPoint`, and their colours from
`momentumStatusStyle`.

### Detail pane

- Title: the momentum's name. The title meta is `[e] edit  [x] disable`, or
  `[e] edit  [x] enable` when the momentum is inactive. It is fitted with `FitHints`.
- Rows use the summary's label/value style:
  - Cadence
  - Target
  - Contexts (or Habits)
  - Streak: `<current>d now · best <best>d`
  - Status
- A description, when present, wraps with `ansi.Wordwrap`.
- Below the rows, the existing chart for the momentum's kind (bucket timeline, daily
  squares or distribution) fills the remaining height at the pane's inner width.
- The footnote ("Latest bucket …") sits on the last line.

## Keys

| Key | Action |
|---|---|
| `↑/↓` | select a momentum |
| `enter` | open the existing details dialog (all widths) |
| `a` | new momentum |
| `e` | edit |
| `x` | enable/disable |
| `D` | delete (not hinted; listed in the help sheet) |
| `,` `.` `g` `+` `-` `0` | date and window keys, unchanged, shown on the date row |

The bottom row shows global keys only.

## Removed

- The bordered pane box and the card boxes.
- The in-pane action line.
- The "Momentum date · window" header.
- The compact and ultra-compact card modes.
- The `MomentumTab` focus and wellbeing states in `model/core.go`. Momentum always
  behaves as `custom` does today.

## Out of scope

- Daemon and API changes.
- The details dialog.
- Momentum data loading: the commands and messages stay as they are.

## Testing

- The strip's two rows at exact width, with groups dropping right-first.
- The row layout: name, sparkline, streak and mark.
- Sparkline scaling, skipped cells and left padding.
- Inactive momentums are dimmed, listed last and marked `paused`.
- Panes at each width class: narrow shows the list only.
- The title meta switches between `disable` and `enable`.
- `enter` opens the details dialog.
- The empty state.
- The bottom row shows global keys only.
- Visual checks via tmux at 60x20, 120x36 and 200x50.
