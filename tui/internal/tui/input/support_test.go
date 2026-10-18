package input

import (
	"testing"

	uistate "crona/tui/internal/tui/state"

	tea "github.com/charmbracelet/bubbletea"
)

type fakeMsg string

func supportDeps(called *[]string) Deps {
	record := func(name string) tea.Cmd {
		*called = append(*called, name)
		return func() tea.Msg { return fakeMsg(name) }
	}
	return Deps{
		OpenSupportFeedbackRoadmapURL: func() tea.Cmd { return record("feedback") },
		OpenSupportReleasesURL:        func() tea.Cmd { return record("releases") },
		OpenSupportDocumentationURL:   func() tea.Cmd { return record("docs") },
		CopySupportDiagnostics:        func(State) tea.Cmd { return record("copy") },
		GenerateSupportBundle:         func(State) tea.Cmd { return record("bundle") },
	}
}

func TestSupportEnterRunsSelectedAction(t *testing.T) {
	for i, key := range []rune{'o', 'r', 'g', 'c', 'b'} {
		var byLetter, byEnter []string
		s := State{ActiveView: uistate.ViewSupport, SupportCursor: i}
		Handle(s, runeKey(key), supportDeps(&byLetter))
		if _, cmd := Handle(s, tea.KeyMsg{Type: tea.KeyEnter}, supportDeps(&byEnter)); cmd == nil {
			t.Fatalf("%d: expected enter to run an action", i)
		}
		if len(byLetter) != 1 || len(byEnter) != 1 || byLetter[0] != byEnter[0] {
			t.Fatalf("%d: expected enter to match %q, got %v vs %v", i, key, byEnter, byLetter)
		}
	}
}

func TestSupportCursorClamps(t *testing.T) {
	var called []string
	s, _ := Handle(State{ActiveView: uistate.ViewSupport}, tea.KeyMsg{Type: tea.KeyUp}, supportDeps(&called))
	if s.SupportCursor != 0 {
		t.Fatalf("expected up at 0 to stay 0, got %d", s.SupportCursor)
	}
	s.SupportCursor = 4
	s, _ = Handle(s, tea.KeyMsg{Type: tea.KeyDown}, supportDeps(&called))
	if s.SupportCursor != 4 {
		t.Fatalf("expected down at 4 to stay 4, got %d", s.SupportCursor)
	}
	s.SupportCursor = 1
	s, _ = Handle(s, tea.KeyMsg{Type: tea.KeyDown}, supportDeps(&called))
	if s.SupportCursor != 2 {
		t.Fatalf("expected down to move, got %d", s.SupportCursor)
	}
}
