//go:build e2e

package e2e

import (
	"testing"

	shareddto "crona/shared/dto"
	"crona/shared/protocol"
	sharedtypes "crona/shared/types"
)

func TestDevSeedAndClearOverIPC(t *testing.T) {
	t.Setenv("CRONA_ENV", "Dev")

	kernel := startTestKernel(t)
	defer kernel.close(t)

	var info sharedtypes.KernelInfo
	kernel.call(t, protocol.MethodKernelInfoGet, nil, &info)
	if info.Env != "Dev" {
		t.Fatalf("expected kernel env Dev, got %q", info.Env)
	}

	var seeded shareddto.OKResponse
	kernel.call(t, protocol.MethodKernelSeedDev, nil, &seeded)
	if !seeded.OK {
		t.Fatalf("expected seed response ok")
	}

	var repos []sharedtypes.Repo
	kernel.call(t, protocol.MethodRepoList, nil, &repos)
	if len(repos) == 0 {
		t.Fatalf("expected seeded repos")
	}

	var issues []sharedtypes.IssueWithMeta
	kernel.call(t, protocol.MethodIssueListAll, nil, &issues)
	if len(issues) < 8 {
		t.Fatalf("expected seeded issues across lifecycle states, got %d", len(issues))
	}

	statusCounts := map[sharedtypes.IssueStatus]int{}
	for _, issue := range issues {
		statusCounts[issue.Status]++
	}
	for _, status := range []sharedtypes.IssueStatus{
		sharedtypes.IssueStatusPlanned,
		sharedtypes.IssueStatusReady,
		sharedtypes.IssueStatusInProgress,
		sharedtypes.IssueStatusBlocked,
		sharedtypes.IssueStatusInReview,
		sharedtypes.IssueStatusDone,
		sharedtypes.IssueStatusAbandoned,
	} {
		if statusCounts[status] == 0 {
			t.Fatalf("expected seeded issue with status %q, got counts: %#v", status, statusCounts)
		}
	}

	var cleared shareddto.OKResponse
	kernel.call(t, protocol.MethodKernelClearDev, nil, &cleared)
	if !cleared.OK {
		t.Fatalf("expected clear response ok")
	}

	kernel.call(t, protocol.MethodRepoList, nil, &repos)
	if len(repos) != 0 {
		t.Fatalf("expected repos to be cleared, got %d", len(repos))
	}
}
