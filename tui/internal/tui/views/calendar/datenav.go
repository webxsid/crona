package calendar

import (
	"fmt"
	"strings"

	shareddatefmt "crona/shared/datefmt"
	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	helperpkg "crona/tui/internal/tui/helpers"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/x/ansi"
)

// DateNav is the date navigation shown at the right of a dated view's sub-tab line: the
// shared date with its week and the keys that move it, shortened to fit room. prefix
// (Rollup's range start) leads the full form only.
func DateNav(
	theme types.Theme,
	rawDate, prefix string,
	settings *api.CoreSettings,
	weekStart sharedtypes.WeekStart,
	room int,
) string {
	date := theme.StyleHeader.Render(helperpkg.FormatDisplayDate(rawDate, settings))
	dated := date
	if parsed, ok := ParseDate(rawDate); ok && !DisplayPatternIncludesWeek(settings) {
		dated = theme.StyleHeader.Render(helperpkg.FormatDisplayDate(rawDate, settings) +
			fmt.Sprintf(" · W%02d", WeekNumber(parsed, weekStart)))
	}
	key := func(k string) string { return theme.StyleHeader.Render("[" + k + "]") }
	dim := theme.StyleDim.Render
	full := key(
		"g",
	) + dim(
		" today   ",
	) + key(
		",",
	) + dim(
		" prev  ",
	) + dated + dim(
		"  next ",
	) + key(
		".",
	)
	candidates := []string{
		full,
		key("g") + "  " + key(",") + " " + dated + " " + key("."),
		dated,
		date,
	}
	if prefix != "" {
		start := theme.StyleHeader.Render(helperpkg.FormatDisplayDate(prefix, settings))
		ranged := key(
			"g",
		) + dim(
			" today   ",
		) + key(
			",",
		) + dim(
			" prev  ",
		) + start + dim(
			" → ",
		) + dated + dim(
			"  next ",
		) + key(
			".",
		)
		candidates = append([]string{ranged}, candidates...)
	}
	for _, candidate := range candidates {
		if ansi.StringWidth(candidate) <= room {
			return candidate
		}
	}
	return ""
}

// DisplayPatternIncludesWeek reports whether the configured date format already shows the week.
func DisplayPatternIncludesWeek(settings *api.CoreSettings) bool {
	pattern := shareddatefmt.DisplayPattern(settings)
	pattern = strings.ReplaceAll(pattern, "[Week]", "")
	pattern = strings.ReplaceAll(pattern, "[week]", "")
	return strings.Contains(pattern, "W")
}
