package dialogs

import (
	"strings"
	"testing"

	sharedtypes "crona/shared/types"
	controllerpkg "crona/tui/internal/tui/dialogs/controller"
	tea "github.com/charmbracelet/bubbletea"
)

func TestRenderDayBoundaryOverviewGroupsDaysAndDisablesAddWhenCovered(t *testing.T) {
	schedule := sharedtypes.DayBoundarySchedule{
		Enabled:     true,
		DefaultTime: "08:30",
		WeekdayOverrides: map[int]string{
			1: "09:00",
			3: "09:00",
		},
	}
	state := controllerpkg.OpenEditDayBoundary(
		controllerpkg.State{Width: 100},
		sharedtypes.CoreSettingsKeyStartOfDay,
		schedule,
		"Asia/Kolkata",
	)
	rendered := renderDayBoundaryDialog(testTheme(), state)
	if !strings.Contains(rendered, "Mon, Wed") || !strings.Contains(rendered, "09:00") {
		t.Fatalf("expected grouped override in render: %q", rendered)
	}

	for day := 1; day <= 7; day++ {
		schedule.WeekdayOverrides[day] = "09:00"
	}
	state = controllerpkg.OpenEditDayBoundary(
		controllerpkg.State{Width: 100},
		sharedtypes.CoreSettingsKeyStartOfDay,
		schedule,
		"Asia/Kolkata",
	)
	rendered = renderDayBoundaryDialog(testTheme(), state)
	if !strings.Contains(rendered, "all days have overrides") || !strings.Contains(rendered, "unavailable") {
		t.Fatalf("expected covered-state affordances in render: %q", rendered)
	}
}

func TestRenderDayBoundaryDaySelectorMarksConfiguredDays(t *testing.T) {
	state := controllerpkg.OpenEditDayBoundary(
		controllerpkg.State{Width: 100},
		sharedtypes.CoreSettingsKeyEndOfDay,
		sharedtypes.DayBoundarySchedule{
			Enabled:     true,
			DefaultTime: "18:00",
			WeekdayOverrides: map[int]string{
				1: "17:00",
			},
		},
		"UTC",
	)
	state, _, _ = controllerpkg.Update(
		state,
		controllerpkg.UpdateContext{},
		"2026-07-29",
		tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}},
	)
	rendered := renderDayBoundaryDialog(testTheme(), state)
	if !strings.Contains(rendered, "Monday (configured)") {
		t.Fatalf("expected configured day marker in render: %q", rendered)
	}
}
