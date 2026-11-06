package calendar

import (
	"fmt"
	"strings"

	shareddatefmt "crona.work/shared/datefmt"
	sharedtypes "crona.work/shared/types"
	"crona.work/tui/internal/api"
	helperpkg "crona.work/tui/internal/tui/helpers"
	types "crona.work/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// DateNav is the date navigation at the right of the row under the tabs: `‹ , date . ›`
// with g for today, shortened to fit room. prefix (Review's range start) leads the date.
func DateNav(
	theme types.Theme,
	rawDate, prefix string,
	settings *api.CoreSettings,
	weekStart sharedtypes.WeekStart,
	room int,
) string {
	// The date reads in bold; keys are bare cyan letters so they do not compete with it.
	bold := lipgloss.NewStyle().Bold(true)
	date := bold.Render(helperpkg.FormatDisplayDate(rawDate, settings))
	dated := date
	if parsed, ok := ParseDate(rawDate); ok && !DisplayPatternIncludesWeek(settings) {
		dated = date + theme.StyleDim.Render(fmt.Sprintf(" · W%02d", WeekNumber(parsed, weekStart)))
	}
	if prefix != "" {
		dated = bold.Render(helperpkg.FormatDisplayDate(prefix, settings)) + theme.StyleDim.Render(" → ") + dated
	}
	key := theme.StyleHeader.Render
	dim := theme.StyleDim.Render
	stepper := dim("‹ ") + key(",") + "  " + dated + "  " + key(".") + dim(" ›")
	candidates := []string{
		stepper + "   " + key("g") + dim(" today"),
		stepper,
		dated,
		date,
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
