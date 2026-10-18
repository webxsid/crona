# TUI System tab redesign: Settings · Alerts · Reports · Config · Updates · Support

Date: 2026-10-18
Branch: `brainstorm/tui-layout`

## Goal

Bring the System tab's six sub-tabs in line with Review, Momentum and Log. Each sub-tab gets
borderless sections, uses the full width and puts its pane keys in its own section header.
This replaces the bordered full-width boxes, which leave over half the screen empty, cut long
paths short and run their lists off the bottom.

## Shared layout

- **Removed:** the bordered pane box, the title and subtitle lines, the in-pane action line,
  and the pane keys on the bottom row. The bottom row shows global keys only. The help sheet
  keeps listing every key.
- **Panes:** `viewchrome.RenderSection` slots joined with `viewchrome.JoinColumns`.
- **Pane widths:** `viewchrome.LogPaneWidths`:
  - wide and medium: List │ Detail
  - narrow: the list alone
- **Keys:** a pane's keys sit in its section title meta, fitted with `viewchrome.FitHints`.
- **Long text:** detail values wrap with `ansi.Wordwrap` (paths wrap whole), and list cells
  truncate with `…`.
- **Keys and cursors:** existing keys, cursors and dialogs keep their meaning.

## Settings, Alerts, Config

- **List:**
  - Group headers become dim dividers, `── <Group> ─…`, padded to the width. Dividers are
    not selectable; the cursor counts selectable rows only, as today.
  - Each row is the label, padded to the longest label (at most 24 cells), then the value.
  - Danger rows keep their red style.
  - The list keeps the cursor in view, along with the divider of its group.
- **Detail:**
  - The title is the row's label. The body has:
    - the row's description: 1–2 sentences, wrapped
    - a blank line
    - the options:
      - for a row with a fixed option set, one line per option, with `●` on the current one
        and `○` on the others
      - for free text and paths, the whole current value, wrapped
    - a blank line
    - the row's keys
  - Descriptions live in one table per sub-tab, keyed by the row's stable key, falling back
    to its label. A test fails if any row has no description.
- **Alerts:** the read-only Backend block (notification backend, sound backend and feature
  support) leaves the list. It becomes a 2-row strip above the panes, packed with
  `viewchrome.StripRows`. The Alerts list holds only editable rows.
- **Config:** directory and docs rows show the full path in the detail, wrapped.

## Reports

- **Strip** (2 rows):
  - Row 1: `Reports N` · `newest <date>`
  - Row 2: `Dir <reports directory>`, truncated with `…` from the left so the tail stays visible
- **List:**
  - Exports are grouped by their created day with `viewchrome.DayList`, newest first.
  - The divider summary is the day's count.
  - Each row is the kind, the period it covers and the format.
- **Detail:**
  - The title is the file name.
  - The rows are `Kind`, `Period`, `Format`, `Size`, `Created`, and `Path` (wrapped whole).
  - The title meta holds `[o] open  [e] edit  [D] delete`.
- **Empty:** the list shows a dim `No exported reports yet`, then a dim line naming where
  reports come from (the export keys in Insights).

## Updates

- **Left:** a status section:
  - the state line (`✓ You're up to date`, or the available version in yellow)
  - `Current`, `Latest`, `Channel` and `Checked`
  - title meta `[r] check again  [d] diagnostics`
  - when diagnostics are shown, a Diagnostics section below it fills the rest of the column
- **Right:** the Release Notes section, full height, wrapped. With no notes, it shows a dim
  `No release notes were published for this release.`
- **Narrow:** the sections stack: status, release notes, diagnostics.

## Support

- **Left:** an Actions list, with one selectable row per action:
  - `Feedback & roadmap`, `Releases`, `Documentation`, `Copy diagnostics`, `Generate bundle`
  - each row is the key hint and the label, then the URL (dim) for link actions
  - `[enter]` runs the selected action; the existing letter keys still work
  - a single dim line above the list replaces the four intro paragraphs
- **Right:** the Diagnostics section, as `label  value` rows, values wrapped.
- **Narrow:** Actions, then Diagnostics stacked.

## Out of scope

- Setting semantics, option sets and validation.
- Kernel and API changes.
- The dialogs.

## Testing

- For each sub-tab:
  - no box, title line or in-pane key line
  - the panes shown at each width class, with lines within the width
  - the detail content for a selected row
  - the bottom row shows global keys only
- Settings, Alerts, Config:
  - dividers are not selectable
  - the cursor keeps its group's divider in view
  - every row has a description
  - options mark the current value
- Alerts: the Backend block is in the strip, not the list.
- Reports: the day grouping, the detail fields and the empty state.
- Support: `[enter]` on each action runs the same command as its letter key.
- Visual checks in tmux at 60x20, 120x36 and 200x50.
