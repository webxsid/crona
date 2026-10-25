package flags

import (
	"slices"
	"testing"
)

func TestTagFilterFlags(t *testing.T) {
	fs := New("test")
	filter := TagFilterFlags(fs)
	if err := fs.Parse([]string{"--tag", "Urgent,+quick", "--tag", "untagged", "--not-tag", "waiting"}); err != nil {
		t.Fatal(err)
	}
	got, err := filter()
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(got.Any, []string{"urgent", "quick", "untagged"}) || !slices.Equal(got.None, []string{"waiting"}) {
		t.Fatalf("unexpected filter %+v", got)
	}

	fs = New("empty")
	filter = TagFilterFlags(fs)
	_ = fs.Parse(nil)
	if got, _ := filter(); got != nil {
		t.Fatalf("expected nil filter, got %+v", got)
	}
}
