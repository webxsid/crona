package types

import "testing"

func TestIssueLifecycleTransitionsUsedByDevSeed(t *testing.T) {
	cases := []struct {
		name string
		from IssueStatus
		to   IssueStatus
		want bool
	}{
		{name: "backlog to planned", from: IssueStatusBacklog, to: IssueStatusPlanned, want: true},
		{name: "planned to ready", from: IssueStatusPlanned, to: IssueStatusReady, want: true},
		{
			name: "planned to in progress",
			from: IssueStatusPlanned,
			to:   IssueStatusInProgress,
			want: true,
		},
		{name: "planned to blocked", from: IssueStatusPlanned, to: IssueStatusBlocked, want: true},
		{name: "in progress to done", from: IssueStatusInProgress, to: IssueStatusDone, want: true},
		{
			name: "planned to abandoned",
			from: IssueStatusPlanned,
			to:   IssueStatusAbandoned,
			want: true,
		},
		{name: "planned to planned", from: IssueStatusPlanned, to: IssueStatusPlanned, want: false},
	}

	for _, tc := range cases {
		if got := IsValidIssueTransition(tc.from, tc.to); got != tc.want {
			t.Fatalf("%s: expected %t, got %t", tc.name, tc.want, got)
		}
	}
}
