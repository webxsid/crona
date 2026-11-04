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
	CollapseDailyDetail   = "daily/detail"
)

// DailyCollapseKeys maps Today's alt+number keys to the pane each collapses.
var DailyCollapseKeys = map[string]string{
	"alt+1": CollapseDailyHabits,
	"alt+2": CollapseDailyCalendar,
	"alt+3": CollapseDailyDetail,
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
