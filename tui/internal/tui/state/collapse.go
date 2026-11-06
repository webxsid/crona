package state

import (
	"runtime"
	"slices"
	"strings"
)

// Collapsible Today panes, saved in core settings' collapsedPanes as these keys.
const (
	CollapseDailyHabits   = "daily/habits"
	CollapseDailyCalendar = "daily/calendar"
	// CollapseDailyCalendarWeek keeps the calendar to its week; it has no key of its own,
	// as the calendar's fold key cycles month, week and hidden.
	CollapseDailyCalendarWeek = "daily/calendar-week"
)

// DailyCollapseKeys maps Today's alt+number keys to the pane each collapses.
var DailyCollapseKeys = map[string]string{
	"alt+1": CollapseDailyHabits,
	"alt+2": CollapseDailyCalendar,
}

// PaneCollapsed reports whether key is among the collapsed panes.
func PaneCollapsed(collapsed []string, key string) bool {
	return slices.Contains(collapsed, key)
}

// ToggleCollapsed returns collapsed with key flipped, leaving the input untouched.
func ToggleCollapsed(collapsed []string, key string) []string {
	if PaneCollapsed(collapsed, key) {
		return slices.DeleteFunc(slices.Clone(collapsed), func(k string) bool { return k == key })
	}
	return append(slices.Clone(collapsed), key)
}

// CalendarMode is how Today shows the score calendar.
type CalendarMode int

const (
	CalendarMonth CalendarMode = iota
	CalendarWeek
	CalendarHidden
)

// DailyCalendarMode reads the calendar's mode from the collapsed panes.
func DailyCalendarMode(collapsed []string) CalendarMode {
	switch {
	case PaneCollapsed(collapsed, CollapseDailyCalendar):
		return CalendarHidden
	case PaneCollapsed(collapsed, CollapseDailyCalendarWeek):
		return CalendarWeek
	}
	return CalendarMonth
}

// WithCalendarMode returns collapsed set to mode, leaving the input untouched.
func WithCalendarMode(collapsed []string, mode CalendarMode) []string {
	out := slices.DeleteFunc(slices.Clone(collapsed), func(k string) bool {
		return k == CollapseDailyCalendar || k == CollapseDailyCalendarWeek
	})
	switch mode {
	case CalendarHidden:
		out = append(out, CollapseDailyCalendar)
	case CalendarWeek:
		out = append(out, CollapseDailyCalendarWeek)
	}
	return out
}

// NextCalendarMode is the fold key's step: month, week, hidden, then month again. A month
// that doesn't fit shows as its week, so from there the key hides it, and from hidden it
// opens to the week.
func NextCalendarMode(mode CalendarMode, monthFits bool) CalendarMode {
	switch {
	case mode == CalendarMonth && monthFits:
		return CalendarWeek
	case mode == CalendarHidden && monthFits:
		return CalendarMonth
	case mode == CalendarHidden:
		return CalendarWeek
	}
	return CalendarHidden
}

// KeyLabel shows an alt+ key the way the keyboard names it: ⌥ on macOS, where the key is
// Option, and alt+ elsewhere.
func KeyLabel(key string) string {
	return keyLabelFor(runtime.GOOS, key)
}

func keyLabelFor(goos, key string) string {
	if goos == "darwin" {
		if rest, ok := strings.CutPrefix(key, "alt+"); ok {
			return "⌥" + rest
		}
	}
	return key
}

// FoldKey is the key that folds a Today pane, the reverse of DailyCollapseKeys.
func FoldKey(pane string) string {
	for key, p := range DailyCollapseKeys {
		if p == pane {
			return key
		}
	}
	return ""
}
