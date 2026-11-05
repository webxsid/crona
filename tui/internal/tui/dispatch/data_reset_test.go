package dispatch

import (
	"reflect"
	"testing"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	"crona/tui/internal/tui/commands"

	tea "github.com/charmbracelet/bubbletea"
)

// stubMessageDeps fills every func in MessageDeps with one that returns zero values, so a
// test can call paths such as fullReloadCmd that touch most of them.
func stubMessageDeps() MessageDeps {
	var deps MessageDeps
	v := reflect.ValueOf(&deps).Elem()
	for i := range v.NumField() {
		field := v.Field(i)
		if field.Kind() != reflect.Func || !field.IsNil() {
			continue
		}
		fn := field.Type()
		field.Set(reflect.MakeFunc(fn, func([]reflect.Value) []reflect.Value {
			out := make([]reflect.Value, fn.NumOut())
			for j := range out {
				out[j] = reflect.Zero(fn.Out(j))
			}
			return out
		}))
	}
	return deps
}

func TestDataResetReloadsEverythingAndDropsPendingEvents(t *testing.T) {
	var plan, summary string
	deps := stubMessageDeps()
	deps.CurrentDate = func(MessageState) string { return "2026-11-05" }
	deps.LoadDailyPlan = func(date string) tea.Cmd { plan = date; return nil }
	deps.LoadDailySummary = func(date string) tea.Cmd { summary = date; return nil }

	state := MessageState{PendingEvents: []api.KernelEvent{{Type: "issue.updated"}}}
	state, _, handled := HandleMessage(
		state,
		commands.KernelEventMsg{Event: api.KernelEvent{Type: sharedtypes.EventTypeDataReset}},
		deps,
	)
	if !handled {
		t.Fatal("expected data.reset handled")
	}
	if plan != "2026-11-05" || summary != "2026-11-05" {
		t.Fatalf("expected the daily plan and summary reloaded, got plan %q summary %q", plan, summary)
	}
	if len(state.PendingEvents) != 0 {
		t.Fatalf("expected pending events dropped, got %v", state.PendingEvents)
	}
}
