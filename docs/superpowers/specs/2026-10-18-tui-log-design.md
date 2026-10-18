# TUI Log tab redesign: Sessions · Habits · Ops

Date: 2026-10-18
Branch: `brainstorm/tui-layout`

## Goal

Bring the Log tab's three sub-tabs in line with Review and Momentum. Each sub-tab gets a
borderless strip, a day-grouped list and a detail pane for the selected row. This replaces
the bordered full-width tables, which leave half the screen empty and cut off long values.

## Shared layout

The order is the strip (2 rows), then `List │ Detail`.

| Width class | Panes |
|---|---|
| wide (≥160) | List │ Detail; the detail is `max(44, (W-6)*2/5)` wide |
| medium (≥100) | List │ Detail; the detail is `max(40, (W-3)*2/5)` wide |
| narrow | List only; `[enter]` opens the existing details dialog |

- Both panes are `viewchrome.RenderSection` slots joined with `viewchrome.JoinColumns`. The
  list is focused, and its title meta is the count.
- Strip rows are exactly the width. Column groups drop right-first, as in `review.StripRows`.
- **Day groups:** the list rows are grouped by local date, newest first. A dim divider sits
  before each day, built as `── <Mon 2 Jan> · <summary> ` and padded to the width with `─`.
  Dividers are not selectable. The cursor counts rows only, so the existing cursor, filter and
  enter wiring keep their meaning.
- **Scrolling:** the list keeps the cursor's row in view. A day's divider is shown whenever
  its first visible row is shown.
- **Long text:** detail values wrap with `ansi.Wordwrap`, and list cells truncate with `…`.
- **Removed:** the bordered pane box, the title and subtitle lines, the in-pane action line,
  and the pane keys on the bottom row. The bottom row shows global keys only, as Review does.
  The help sheet keeps listing every key.
- `[enter]` opens the existing details dialog at every width.

## Sessions

- **Strip:**
  - Row 1: `Sessions N` · `total Xh` · `avg Ym`
  - Row 2: `This week Xh` (sessions whose end falls in the current week, by the week-start
    setting) · `longest Xh`
  - Sums use `DurationSeconds`; a missing duration counts as 0.
- **Divider summary:** the day's session count and total, e.g. `2 · 1h01m`.
- **Row:** `16:30–16:49   19m  <note>`, built from:
  - the local start and end times; an open session shows `16:30–now`
  - the compact duration, right-aligned in 6 cells
  - the first line of the note, or a dim `no note`
- **Detail:** the title is the session's date and time range. The body shows:
  - `Issue`, `Repo`, `Stream`
  - `Work` / `Rest` with a two-colour bar
  - the segments, one row each, as kind, time range and duration
  - the parsed note sections, each with a heading and wrapped text
- **Loading the detail:**
  - The list data has no issue or segments. The detail is fetched with the existing
    `GetSessionDetail` through a new `LoadSessionPreview(c, id)` command. That command returns
    `SessionPreviewLoadedMsg{Detail}`, or `SessionPreviewFailedMsg{ID, Err}` on failure, so
    the dialog's own `SessionDetailLoadedMsg` path is untouched.
  - The fetch is debounced: a cursor change schedules a 150 ms tick that carries the session
    ID. The fetch only runs if that ID is still selected when the tick fires.
  - A loaded preview is stored in the model as `sessionPreview` and passed to the view as
    `ContentState.SessionPreview`. A preview for any other session than the selected one is
    ignored.
  - Until the preview arrives, the detail shows the list's own fields (time, duration, note)
    plus a dim `Loading…`. On failure it shows a dim `Details unavailable`.
  - Previews are only fetched on medium and wide screens, where the detail pane is shown.

## Habits

- **Strip:**
  - Row 1: `Done N  failed N` · `total Xh`
  - Row 2: `Most logged  <habit> N`
  - These are computed over the loaded history.
- **Divider summary:** `<done> done · <failed> failed`, leaving out a part that is zero.
- **Row:** `✓ Morning Walk        23m  Personal › wellbeing`, built from:
  - the status mark: `✓` completed in green, `✗` failed in red, `·` any other status, dim
  - the habit name (`HabitName`, falling back to `SnapshotName`), padded to the longest name
    up to 24 cells
  - the duration, or `–`
  - the context `Repo › Stream`, dim
- **Detail:** the title is the habit name. The rows are `Date`, `Status`, `Duration`,
  `Context`, `Kind` and `Target` (only when there is a snapshot target), then the note,
  wrapped. Everything comes from the list data; nothing is loaded.

## Ops

- **Strip:**
  - Row 1: `Ops N` · `create N  update N  delete N`
  - Row 2: `Since <oldest timestamp, local, YYYY-MM-DD HH:MM>`
- **Divider summary:** the day's op count.
- **Row:** `04:34:07  issue             update  10`, built from:
  - the local time
  - the entity as words (`habit_completion` becomes `habit completion`), padded to the
    longest entity
  - the action coloured: create in DullGreen, update in Yellow, delete in DullRed, anything
    else Normal
  - the target ID, whole
- **List title meta:** `N · [+/-] limit`. The `+`/`-` keys and their limit rules are unchanged.
- **Detail:** the title is `<entity> <action>`. The rows are `Target`, `Time` (full local
  timestamp), `Device` and `User`, then `Payload`. The payload is a JSON object rendered as
  indented `key  value` lines, nested objects indented 2 more, and long values wrapped. Any
  other payload is shown as compact JSON, wrapped. A missing payload is a dim `–`.

## Out of scope

- Daemon and API changes.
- The details dialogs.
- Pagination beyond the Ops limit.

## Testing

- For each sub-tab:
  - the strip has 2 rows at exact width, with groups dropping right-first
  - day dividers appear in the right places, and the cursor skips them
  - the row format, with truncation at narrow widths
  - the panes shown at each width class
  - the detail content
  - the bottom row shows global keys only
- Sessions:
  - a preview is requested only after the debounce, and only for the still-selected session
  - a stale preview is ignored
  - the loading and failed states
  - no fetch on narrow screens
- Ops: the payload rendering for an object, a nested object, a non-object and a missing
  payload.
- Visual checks in tmux at 60x20, 120x36 and 200x50.
