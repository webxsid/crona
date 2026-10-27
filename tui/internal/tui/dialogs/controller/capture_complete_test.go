package controller

import (
	"testing"

	"crona/tui/internal/api"
)

func captureCompleteCtx() UpdateContext {
	return UpdateContext{
		CurrentDate: "2026-10-27",
		Repos:       []api.Repo{{ID: 1, Name: "crona"}, {ID: 2, Name: "home"}},
		Streams:     []api.Stream{{ID: 10, RepoID: 1, Name: "api"}, {ID: 11, RepoID: 1, Name: "web"}, {ID: 20, RepoID: 2, Name: "chores"}},
	}
}

func suggestionInserts(state State) []string {
	out := []string{}
	for _, s := range state.CaptureSuggestions {
		out = append(out, s.Insert)
	}
	return out
}

func TestCaptureSuggestsTargetsAndInsertsOnTab(t *testing.T) {
	state := OpenCaptureIssues(State{}, 0, "", []string{"urgent", "home"})
	ctx := captureCompleteCtx()
	for _, r := range "- fix login #w" {
		state, _, _ = Update(state, ctx, ctx.CurrentDate, captureKey(string(r)))
	}
	if got := suggestionInserts(state); len(got) != 1 || got[0] != "#crona/web" {
		t.Fatalf("a stream prefix should suggest its full path, got %v", got)
	}
	state, action, _ := Update(state, ctx, ctx.CurrentDate, captureKey("tab"))
	if action != nil || state.Description.Value() != "- fix login #crona/web " || len(state.CaptureSuggestions) != 0 {
		t.Fatalf("tab should insert the suggestion: %q %v", state.Description.Value(), state.CaptureSuggestions)
	}
}

func TestCaptureSuggestsTagsEstimatesAndDates(t *testing.T) {
	ctx := captureCompleteCtx()
	cases := map[string][]string{
		"- a +u":  {"+urgent"},
		"- a ~1":  {"~15m", "~1h", "~1h30m"},
		"- a @to": {"@today", "@tomorrow"},
		"- a @fr": {"@fri"},
		"a #cr":   {},
		"- a #zz": {},
	}
	for text, want := range cases {
		state := OpenCaptureIssues(State{}, 0, "", []string{"urgent", "home"})
		for _, r := range text {
			state, _, _ = Update(state, ctx, ctx.CurrentDate, captureKey(string(r)))
		}
		got := suggestionInserts(state)
		if len(got) != len(want) {
			t.Fatalf("%q: want %v, got %v", text, want, got)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("%q: want %v, got %v", text, want, got)
			}
		}
	}
}

func TestCaptureSuggestEscDismissesWithoutClosing(t *testing.T) {
	ctx := captureCompleteCtx()
	state := OpenCaptureIssues(State{}, 0, "", nil)
	for _, r := range "- a ~" {
		state, _, _ = Update(state, ctx, ctx.CurrentDate, captureKey(string(r)))
	}
	state, _, _ = Update(state, ctx, ctx.CurrentDate, captureKey("down"))
	if state.CaptureSuggestCursor != 1 {
		t.Fatalf("down should move the popup cursor, got %d", state.CaptureSuggestCursor)
	}
	state, _, _ = Update(state, ctx, ctx.CurrentDate, captureKey("esc"))
	if state.Kind != "capture_issues" || len(state.CaptureSuggestions) != 0 {
		t.Fatalf("esc should only dismiss the popup: %q %v", state.Kind, state.CaptureSuggestions)
	}
}
