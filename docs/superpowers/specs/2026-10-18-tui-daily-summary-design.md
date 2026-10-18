# Fold Summary into Daily — Design

**Status:** approved in chat, 2026-10-18. Branch: `brainstorm/tui-layout`.

## Problem

Today has two date views of the same day: Daily (lists + strip) and Summary (a read-only page). About two-thirds of Summary repeats Daily:
- its agenda is Daily's two lists;
- its Issues/Habits/worked figures are the strip's meters;
- its mood/energy/sleep and focus/check-in streaks are in the strip.

What only Summary shows is the day's scores and the detail behind them.

## Design

1. **Headline numbers in the strip.** A last strip column shows `Score <n>/100` over `Plan <n>%` (focus score over plan accountability), read from the summary snapshot for the shown date. Being rightmost, it drops first when narrow. It is absent until the snapshot for the date loads.
2. **Day summary in the detail pane.** `S` on Daily switches the wide-screen detail pane between the selection and a Day summary:
   - Focus score with level, reason and a bar to target; worked, rest, sessions, target;
   - Accountability: score, backlog pressure, done/pending/failed/abandoned;
   - Wellbeing: the check-in (mood, energy, sleep, screen time, notes) or the averages;
   - Streaks: focus, check-ins, habits.

   The pane title is `Day`. Moving the cursor or switching panes returns the pane to the selection.
3. **Dialog when there is no detail pane.** Below the wide breakpoint `S` opens the same Day summary as a dialog (title `Day`, the date as its name).
4. **Calendar on `C`.** The scored month calendar opens from Daily on `C` and sets the shared date. `c` stays context checkout.
5. **No Summary sub-tab.** Today becomes `Daily · Session · Away`. The Summary view, its keys, jump entry and full-page renderer go; its section renderers stay in `views/summary` as the Day summary.

## Out of scope

- Daemon changes: the summary snapshot already reloads with every date change.
- The Summary export report (unrelated, keeps its "summary" category).
