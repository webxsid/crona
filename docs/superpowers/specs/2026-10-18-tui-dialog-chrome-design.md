# TUI dialogs, sub-project 0: dialog chrome

Date: 2026-10-18
Branch: `brainstorm/tui-layout`

## Goal

Give every dialog one shared look that matches the borderless panes. This sub-project
builds that chrome and proves it on two dialogs: New Issue and the context switcher.
Sub-projects 1–6 move the remaining dialogs onto it:

1. entity forms
2. pickers
3. the timer lifecycle
4. export
5. details
6. settings dialogs

It changes rendering only. Keys, focus order, validation and the controller state of every
dialog stay as they are.

## Frame

- **Border:** dialogs keep a rounded border, which is the only box on screen, so the dialog
  reads as on top of the view.
- **Title:** sits in the top border: `╭─ <Title> ─…─╮`.
- **Context:** an optional dim context label (for example `OSS › cli`) sits right-aligned in
  the top border: `─ OSS › cli ─╮`. It is dropped when the title and context do not both fit.
- **Border colour:** keeps today's meaning. It is the theme's normal dialog colour, and red
  for destructive dialogs (delete, wipe).
- **Padding:** 1 line top and bottom, 2 cells left and right.
- **Width:** fits the content, clamped to `[44, 76]`, and never wider than `screenWidth - 8`.
- **Height:** the content's height. There are no trailing blank rows. When the content is
  taller than `screenHeight - 4`, the body truncates with a dim `↓ more` above the footer.
- **Placement:** centred on screen, as today.

## Fields

- **Field row:**
  - a dim label, padded to the dialog's label column (the longest label, at most 12
    cells), then two spaces, then the value
  - the focused field's label is bold, in the accent colour
  - the focused value shows the text cursor
- **Empty values:** show the placeholder in dim. For optional fields with no placeholder,
  this is `optional`. A field's label is never repeated in its placeholder.
- **Pairs:** two short fields can share one row, as `<field>  <field>`, each taking half the
  inner width. They fall back to two rows when the inner width is below 56.
- **Fixed height (rule):** a dialog never changes height while you move between its fields.
  Every field keeps its row count, and the footer is padded to its tallest variant.
- **Textarea:**
  - the label is on its first row
  - the body has a fixed height (3 rows) behind a dim `┊` gutter
- **Picker** (repo, stream and other searchable choices):
  - a single row: `<label>  ‹ <current> ›  i/n`
  - while focused, the arrows are accented and `filter: <input>` follows it
  - with no options, it shows the placeholder in dim
  - the existing keys keep working: typing filters, `←/→` cycle
  - this replaces the separate `⌕ Search …` input and the `[ > X ]` selector row
  - pickers sit below the other fields, as before the redesign
- **Choices** (a fixed list, such as the timer type):
  - each row is `▶ [k] Label` or `  [k] Label`
  - the selected choice's help line is dim, below the list after a blank line
- **Errors:** a field's validation error is red, on the line under that field.
- **Static text** (a description or a confirm prompt): wrapped to the inner width, dim
  unless it is the dialog's main message.

## Key footer

- **Placement:** the footer is the last body block, after one blank line.
- **Hints:** written as `key action`, with no brackets.
- **Order:** the primary action comes first, with its key in the accent colour, and
  `esc cancel` comes last.
- **Packing:** `viewchrome.PackHints` packs the hints to the inner width, so a hint never
  splits across lines.
- **Help sheet:** unchanged.

## Shared code

- **New `dialogs/frame.go`:**
  - `Frame(theme, FrameSpec{Title, Context string; Danger bool; Body []string; Footer []string; ScreenWidth, ScreenHeight int}) string`
  - `Field(theme, label string, column int, value string, focused bool, placeholder string) string`
  - `FieldPair(theme, left, right string, inner int) []string`
  - `Textarea(theme, label string, column int, lines []string, focused bool, inner int) []string`
  - `Picker(theme, PickerSpec{Label string; Column int; Focused bool; Filter, Current string; Options []string; Selected int; Placeholder string}) []string`
  - `Choices(theme, []Choice{Key, Label, Help string}, selected int, inner int) []string`
  - `Footer(theme, primary string, hints []string, inner int) []string`
- **The old helpers** (`modal`, `formModal`, `renderSelector`, `renderInputColumns`,
  `appendDialogFooter`) stay until every group has moved, then are deleted in sub-project 6.

## Proof dialogs

- **New Issue** (`create_issue_default` and `create_issue_meta`):
  - fields: Title; a Description textarea; Estimate and Due as a pair; then Repo and Stream
    pickers at the bottom
  - footer: `ctrl+s save`, `ctrl+a create more`, `ctrl+f create & focus`, `tab next`,
    `esc cancel`
  - context: the active repo › stream
- **Context switcher** (`checkout_context`):
  - Repo and Stream pickers
  - footer: `enter set context`, `c clear`, `↑/↓/tab move`, `esc cancel`

## Out of scope

- All other dialogs, which belong to sub-projects 1–6.
- Behaviour, keys, validation and controller state.
- The help sheet.

## Testing

- **Frame:** the title is in the top border, and the context is right-aligned and dropped
  when it doesn't fit. The width is clamped to `[44, 76]` and to `screenWidth - 8`. There are
  no trailing blank rows. Tall content gets `↓ more`. A destructive dialog's border is red.
- **Field:** the label column pads, and a focused label is styled. The placeholder shows when
  the value is empty and is never the label.
- **FieldPair:** falls back to two rows below an inner width of 56.
- **Picker:**
  - the list shows only while focused
  - at most 5 options, plus `+N more`
  - `▶` on the selected option
  - `▾` when unfocused
- **Footer:** no hint splits across lines, the primary comes first, and `esc cancel` comes
  last.
- **New Issue and the context switcher:** render with the new chrome at 60x20, 120x36 and
  200x50 (no `[ > ` and no `⌕ Search`). Their existing controller and dialog tests stay
  green.
- **Visual check** in tmux at those three sizes.

## Revision (after review)

The user rejected the expanding picker list and the pickers sitting on top:

- the list was too jumpy, because the row count kept changing
- the pickers belong at the bottom

The Fields section above now describes the replacement:

- one-row `‹ choice ›` pickers at the bottom
- a fixed-height textarea
- a footer padded to its tallest variant
