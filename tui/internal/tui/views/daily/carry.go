package daily

import (
	"fmt"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	helperpkg "crona/tui/internal/tui/helpers"
	types "crona/tui/internal/tui/views/types"
)

// CarryMarkerText is the compact carry-over marker for issue rows, unstyled so it takes the
// row's style: where a carried issue went and how long it was worked, or where a carried-in
// issue came from. It is "" for issues outside a carry chain.
func CarryMarkerText(state types.ContentState, issueID int64) string {
	entry := issuePlanEntry(state, issueID)
	if entry == nil {
		return ""
	}
	if entry.Status == sharedtypes.DailyPlanEntryStatusCarried {
		text := "↪"
		if entry.CarriedToDate != nil {
			text += " " + helperpkg.FormatDisplayDate(*entry.CarriedToDate, state.Settings)
		}
		if entry.WorkedSeconds > 0 {
			text += " · " + helperpkg.FormatCompactDurationSeconds(entry.WorkedSeconds)
		}
		return text
	}
	if entry.CarriedFromDate == nil {
		return ""
	}
	text := "↩ from " + helperpkg.FormatDisplayDate(*entry.CarriedFromDate, state.Settings)
	if entry.CarryCount > 1 {
		text = fmt.Sprintf("↩ ×%d", entry.CarryCount)
	}
	if entry.Stalled {
		text += " stalled"
	}
	return text
}

// RescheduleCarries reports whether rescheduling the issue to a later day will carry it over:
// the kernel carries an issue that was worked on its planned day when that day is today.
func RescheduleCarries(plan *api.DailyPlan, issueID int64, today string, settings *api.CoreSettings) bool {
	if plan == nil || plan.Date != today || settings == nil || !settings.CarryOverEnabled {
		return false
	}
	for _, entry := range plan.Entries {
		if entry.IssueID == issueID {
			return entry.Status == sharedtypes.DailyPlanEntryStatusPlanned && entry.WorkedSeconds > 0
		}
	}
	return false
}

// carryLines explains where the issue sits in its carry chain on the shown day.
func carryLines(theme types.Theme, state types.ContentState, issueID int64, width int) []string {
	entry := issuePlanEntry(state, issueID)
	if entry == nil {
		return nil
	}
	label := func(text string) string { return theme.StyleDim.Render(padCells(text, 14)) }
	lines := []string{}
	if entry.Status == sharedtypes.DailyPlanEntryStatusCarried {
		to := "the next day"
		if entry.CarriedToDate != nil {
			to = helperpkg.FormatDisplayDate(*entry.CarriedToDate, state.Settings)
		}
		lines = append(lines, label("Carried to")+theme.StyleNormal.Render(fmt.Sprintf(
			"%s after %s of work", to, helperpkg.FormatCompactDurationSeconds(entry.WorkedSeconds),
		)))
	}
	if entry.CarriedFromDate != nil {
		text := helperpkg.FormatDisplayDate(*entry.CarriedFromDate, state.Settings)
		if entry.CarryCount > 1 {
			text += fmt.Sprintf(" · %d carries in a row", entry.CarryCount)
		}
		lines = append(lines, label("Carried from")+theme.StyleNormal.Render(text))
	}
	if entry.Stalled && entry.Status != sharedtypes.DailyPlanEntryStatusCarried {
		lines = append(lines, label("")+theme.StyleError.Render("stalled · consider splitting, re-estimating or abandoning it"))
	}
	if len(lines) == 0 {
		return nil
	}
	return truncateLines(append([]string{"", theme.StyleDim.Bold(true).Render("Carry-over")}, lines...), width)
}
