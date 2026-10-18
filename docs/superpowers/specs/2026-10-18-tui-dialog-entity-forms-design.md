# TUI dialogs, sub-project 1: entity forms

Date: 2026-10-18
Branch: `brainstorm/tui-layout`
Builds on: `2026-10-18-tui-dialog-chrome-design.md` (sub-project 0)

## Goal

Move the 12 field-form dialogs onto the shared dialog chrome, through one shared form
builder. New Issue moves onto the builder too, so all 13 forms share one path.

It changes rendering only. Keys, focus order, validation and controller state stay as they
are. Field focus moves with `tab` / `shift+tab` only.

## Scope

| Kind | Title |
|---|---|
| `create_repo`, `edit_repo` | New Repo / Edit Repo |
| `create_stream`, `edit_stream` | New Stream / Edit Stream |
| `create_habit`, `edit_habit` | New Habit / Edit Habit |
| `edit_issue` | Edit Issue |
| `issue_status_note` | the current status title (Block Issue, Send To Review, …) |
| `create_checkin`, `edit_checkin` | New Check-In / Edit Check-In |
| `complete_habit` | `ViewTitle`, else Habit Log |
| `create_alert_reminder`, `edit_alert_reminder` | today's titles (Add/Edit Check-In or Plan-the-Day Reminder) |

The momentum create/edit wizard is out of scope. It is a set of step lists, so it moves in
sub-project 2 (pickers).

## Layouts

| Dialog | Border context | Body, top to bottom |
|---|---|---|
| Repo | none | Name; Description textarea |
| Stream | `RepoName` | Name; Description textarea |
| New Habit | the active repo › stream | Name; Description textarea; Schedule and Target as a pair; Repo and Stream pickers |
| Edit Habit | repo › stream, when known | Name; Description textarea; Schedule and Target as a pair |
| Edit Issue | repo › stream, when known | Title; Description textarea; Estimate and Due as a pair |
| Status note | the issue title (`ViewName`) | one `Note` field, labelled from `StatusLabel` |
| Check-In | the date | Mood and Energy as a pair; Sleep and Sleep score as a pair; Screen time; Notes |
| Habit Log | the date | Duration; Notes textarea |
| Reminder | none | Schedule and Time as a pair; then the two format help lines as dim static text |

- **Labels:** short and without `(Optional)`. Optional fields show the dim `optional`
  placeholder instead, as the chrome spec says.
- **Border colour:** unchanged from today: create forms are cyan, edit forms and the status
  note are yellow, and Habit Log is green.
- **Fixed height:** each dialog keeps one height while you tab through its fields, and its
  footer is padded to its tallest variant.

## Footers

- The primary hint comes first: the submit key plus `create` or `save` (`set` for the status
  note), using the configured submit chord.
- Then `tab next`.
- Then `enter newline`, only while a textarea is focused.
- Then a dialog's existing extra keys while their field is focused, such as `g today` on
  Due, or `type filter` and `←/→ choose` on the pickers.
- `esc cancel` comes last.

## Shared code

- New `dialogs/form.go`:
  - `type FormRow` describes one row: a field, a field pair, a textarea, a picker or static
    text, plus whether it holds the focused input.
  - `formBody(theme, rows []FormRow, inner int) (body []string, focus int)` lays the rows out
    and returns the body line of the focused row.
  - `formFooter(theme, primary string, variants [][]string, current []string, inner int) []string`
    renders the current hints and pads them to the tallest variant.
- `renderIssueCreate` is rewritten onto `formBody` and `formFooter`. Its tests stay green.
- The old helpers stay until sub-project 6.

## Testing

- **Each dialog:** a chrome test that checks:
  - the title is in the top border
  - the border context is right
  - the body has no `(Optional)` and no `[key]` brackets
  - the footer starts with the primary hint and ends with `esc cancel`
- **Height:** tab through every field of each dialog; the rendered height never changes.
- **Short terminal:** at height 20, Check-In and New Habit keep the focused field and the
  footer visible.
- **Existing tests:** controller and dialog tests stay green. Tests that pin old layout
  strings are updated, and each one is listed in its commit body.
- **Visual check** in tmux at 60x20, 120x36 and 200x50.
