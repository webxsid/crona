package types

import "testing"

func TestNormalizeHabitStreakDefinitionPreservesRepoWideContexts(t *testing.T) {
	def := NormalizeHabitStreakDefinition(HabitStreakDefinition{
		TargetKind: MomentumTargetKindContext,
		Contexts: []MomentumContext{
			{RepoID: 7},
			{RepoID: 7},
		},
	})

	if len(def.Contexts) != 1 {
		t.Fatalf("expected one normalized repo-wide context, got %+v", def.Contexts)
	}
	if def.Contexts[0].RepoID != 7 || def.Contexts[0].StreamID != nil {
		t.Fatalf("expected repo-wide context to preserve nil stream, got %+v", def.Contexts[0])
	}
}

func TestMomentumContextRedundanciesDetectRepoWideCoverage(t *testing.T) {
	redundancies := MomentumContextRedundancies([]MomentumContext{
		{RepoID: 7},
		{RepoID: 7, StreamID: new(int64(9))},
		{RepoID: 7, StreamID: new(int64(10))},
		{RepoID: 8, StreamID: new(int64(11))},
	})

	if len(redundancies) != 1 {
		t.Fatalf("expected one redundancy group, got %+v", redundancies)
	}
	if redundancies[0].RepoWideContext.RepoID != 7 {
		t.Fatalf("expected repo 7 redundancy, got %+v", redundancies[0])
	}
	if len(redundancies[0].RedundantContexts) != 2 {
		t.Fatalf("expected two redundant stream contexts, got %+v", redundancies[0].RedundantContexts)
	}
}
