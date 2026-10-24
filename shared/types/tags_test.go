package types

import "testing"

func TestNormalizeTagName(t *testing.T) {
	for in, want := range map[string]string{
		"Urgent":       "urgent",
		"+quick":       "quick",
		" client acme": "client-acme",
		"client/acme":  "client/acme",
		"v2_beta":      "v2_beta",
	} {
		got, err := NormalizeTagName(in)
		if err != nil || got != want {
			t.Fatalf("NormalizeTagName(%q) = %q, %v; want %q", in, got, err, want)
		}
	}
	for _, bad := range []string{"", "-lead", "a.b", "untagged", "abcdefghijklmnopqrstuvwxyz1234567"} {
		if _, err := NormalizeTagName(bad); err == nil {
			t.Fatalf("expected %q to be rejected", bad)
		}
	}
}

func TestTagFilterMatches(t *testing.T) {
	tagged := []string{"quick", "urgent"}
	cases := []struct {
		filter TagFilter
		tags   []string
		want   bool
	}{
		{TagFilter{}, nil, true},
		{TagFilter{Any: []string{"urgent", "review"}}, tagged, true},
		{TagFilter{Any: []string{"review"}}, tagged, false},
		{TagFilter{None: []string{"quick"}}, tagged, false},
		{TagFilter{Any: []string{TagUntagged}}, nil, true},
		{TagFilter{Any: []string{TagUntagged}}, tagged, false},
		{TagFilter{None: []string{TagUntagged}}, nil, false},
	}
	for _, tc := range cases {
		if got := tc.filter.Matches(tc.tags); got != tc.want {
			t.Fatalf("%+v.Matches(%v) = %v, want %v", tc.filter, tc.tags, got, tc.want)
		}
	}
}
