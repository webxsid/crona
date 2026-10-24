package helpers

import "testing"

func TestItemFilter(t *testing.T) {
	f := ParseItemFilter("+Urgent login -waiting")
	if f.Text != "login" || len(f.Tags.Any) != 1 || f.Tags.Any[0] != "urgent" || f.Tags.None[0] != "waiting" {
		t.Fatalf("unexpected parse: %+v", f)
	}
	if !f.Match("Fix login redirect", []string{"urgent"}) {
		t.Fatal("expected match")
	}
	if f.Match("Fix login redirect", []string{"urgent", "waiting"}) {
		t.Fatal("expected -waiting to exclude")
	}
	if f.Match("Write docs", []string{"urgent"}) {
		t.Fatal("expected text to still apply")
	}
	if !ParseItemFilter("re-do").Match("re-do list", nil) {
		t.Fatal("expected a word with an inner dash to stay text")
	}
	if got := TagFilterQuery("+urgent login -waiting"); got != "+urgent -waiting" {
		t.Fatalf("unexpected tag query %q", got)
	}
}
