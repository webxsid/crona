package capture

import (
	"slices"
	"testing"

	shareddto "crona/shared/dto"
	sharedtypes "crona/shared/types"
)

// 2026-10-27 is a Tuesday.
const today = "2026-10-27"

var catalog = Catalog{
	Repos: []sharedtypes.Repo{{ID: 1, Name: "crona"}, {ID: 2, Name: "home"}, {ID: 3, Name: "docs-site"}, {ID: 4, Name: "Side Projects"}},
	Streams: []sharedtypes.Stream{
		{ID: 10, RepoID: 1, Name: "api"},
		{ID: 11, RepoID: 1, Name: "docs"},
		{ID: 20, RepoID: 2, Name: "chores"},
		{ID: 30, RepoID: 3, Name: "main"},
		{ID: 40, RepoID: 4, Name: "Blog Posts"},
	},
}

func parseOne(t *testing.T, line string) shareddto.IssueCaptureProposal {
	t.Helper()
	out := Parse("- "+line, catalog, nil, today)
	if len(out) != 1 {
		t.Fatalf("parsed %d proposals from %q", len(out), line)
	}
	return out[0]
}

func deref[T any](v *T) any {
	if v == nil {
		return nil
	}
	return *v
}

func TestAnnotationsAndPhrases(t *testing.T) {
	cases := []struct {
		line, title, date string
		estimate          int
		stream            int64
		tags              []string
	}{
		{"fix login redirect #api +urgent for 45 min tomorrow", "fix login redirect", "2026-10-28", 45, 10, []string{"urgent"}},
		{"fix login #crona/docs ~1h30m @fri", "fix login", "2026-10-30", 90, 11, nil},
		{"read 2 min intro", "read 2 min intro", "", 0, 0, nil},
		{"call mom friday", "call mom", "2026-10-30", 0, 0, nil},
		{"call mom next friday", "call mom", "2026-11-06", 0, 0, nil},
		{"plan trip next week", "plan trip", "2026-11-02", 0, 0, nil},
		{"renew passport in 3 days takes 20m", "renew passport", "2026-10-30", 20, 0, nil},
		{"tax filing on 5 jan", "tax filing", "2027-01-05", 0, 0, nil},
		{"ship it @tuesday", "ship it", "2026-10-27", 0, 0, nil},
		{"sweep #chores for 1.5 hours", "sweep", "", 90, 20, nil},
		// An annotation wins and the phrase stays in the title.
		{"review notes @today tomorrow", "review notes tomorrow", "2026-10-27", 0, 0, nil},
		// A phrase alone never empties the title.
		{"tomorrow", "tomorrow", "", 0, 0, nil},
	}
	for _, tc := range cases {
		p := parseOne(t, tc.line)
		if p.Title != tc.title {
			t.Errorf("%q: title %q, want %q", tc.line, p.Title, tc.title)
		}
		if got := deref(p.TodoForDate); (tc.date == "" && got != nil) || (tc.date != "" && got != tc.date) {
			t.Errorf("%q: date %v, want %q", tc.line, got, tc.date)
		}
		if got := deref(p.EstimateMinutes); (tc.estimate == 0 && got != nil) || (tc.estimate != 0 && got != tc.estimate) {
			t.Errorf("%q: estimate %v, want %d", tc.line, got, tc.estimate)
		}
		if tc.stream != 0 && p.StreamID != tc.stream {
			t.Errorf("%q: stream %d, want %d", tc.line, p.StreamID, tc.stream)
		}
		if !slices.Equal(p.Tags, tc.tags) {
			t.Errorf("%q: tags %v, want %v", tc.line, p.Tags, tc.tags)
		}
	}
}

func TestTargetsAndFlags(t *testing.T) {
	cases := []struct {
		line   string
		repo   int64
		stream int64
		flag   string
	}{
		{"a #home", 2, 20, ""},             // repo with one stream
		{"a #crona", 1, 0, "stream?"},      // repo with several streams
		{"a #doc", 0, 0, "ambiguous #doc"}, // stream docs and repo docs-site
		{"a #nope", 0, 0, "unknown #nope"},
		{"a #side-projects/blog-posts", 4, 40, ""}, // names with spaces as slugs
		{"a #Side-Projects", 4, 40, ""},
		{"a #side-pro/blog", 4, 40, ""},
		{"a #blog-posts", 4, 40, ""},
		{"a #crona/nope", 1, 0, "unknown #crona/nope"},
		{"a @someday", 0, 0, "bad date @someday"},
		{"a +Bad_Tag!", 0, 0, "bad tag +Bad_Tag!"},
	}
	for _, tc := range cases {
		p := parseOne(t, tc.line)
		if p.RepoID != tc.repo || p.StreamID != tc.stream {
			t.Errorf("%q: target %d/%d, want %d/%d", tc.line, p.RepoID, p.StreamID, tc.repo, tc.stream)
		}
		if tc.flag == "" && len(p.Flags) > 0 || tc.flag != "" && !slices.Contains(p.Flags, tc.flag) {
			t.Errorf("%q: flags %v, want %q", tc.line, p.Flags, tc.flag)
		}
	}
}

func TestDefaultTargetAndInbox(t *testing.T) {
	if p := parseOne(t, "call bank"); !p.Inbox || p.StreamName != InboxName || p.StreamID != 0 {
		t.Fatalf("untargeted line should go to the Inbox: %+v", p)
	}
	out := Parse("- call bank", catalog, &Target{RepoID: 2, RepoName: "home", StreamID: 20, StreamName: "chores"}, today)
	if out[0].Inbox || out[0].StreamID != 20 {
		t.Fatalf("untargeted line should use the default target: %+v", out[0])
	}
}

func TestDetectionAndDescriptions(t *testing.T) {
	text := "Intro prose\n" +
		"- first #api\n" +
		"  happens after the SSO change\n" +
		"\n" +
		"  - second detail line\n" +
		"* star lists are not tasks\n" +
		"1. nor are numbered ones\n" +
		"```\n- inside a fence\n```\n" +
		"- second\n" +
		"prose ends the description scope\n" +
		"  orphan indented line\n"
	out := Parse(text, catalog, nil, today)
	if len(out) != 2 {
		t.Fatalf("want 2 proposals, got %d: %+v", len(out), out)
	}
	if out[0].Line != 2 || deref(out[0].Description) != "happens after the SSO change\n- second detail line" {
		t.Fatalf("unexpected first proposal %+v (description %v)", out[0], deref(out[0].Description))
	}
	if out[1].Title != "second" || out[1].Description != nil {
		t.Fatalf("unexpected second proposal %+v", out[1])
	}
}
