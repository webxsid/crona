package repositories_test

import (
	"context"
	"path/filepath"
	"testing"

	storedb "crona/kernel/internal/store/db"
	migrations "crona/kernel/internal/store/migrations"
	storemodels "crona/kernel/internal/store/models"
	repositories "crona/kernel/internal/store/repositories"
	sharedtypes "crona/shared/types"
)

func TestEnsureAdhocContextIsIdempotentAndOwned(t *testing.T) {
	ctx := context.Background()
	db := openRepositoryTestStore(t)
	repository := repositories.NewReservedContextRepository(db.DB())

	first, err := repository.EnsureAdhocContext(ctx, "user-a", "2026-09-24T10:00:00Z")
	if err != nil {
		t.Fatalf("ensure first context: %v", err)
	}
	second, err := repository.EnsureAdhocContext(ctx, "user-a", "2026-09-24T11:00:00Z")
	if err != nil {
		t.Fatalf("ensure second context: %v", err)
	}
	if first != second {
		t.Fatalf("expected stable context ids, got %+v then %+v", first, second)
	}

	other, err := repository.EnsureAdhocContext(ctx, "user-b", "2026-09-24T12:00:00Z")
	if err != nil {
		t.Fatalf("ensure other user context: %v", err)
	}
	if other == first {
		t.Fatalf("expected different users to receive different context ids")
	}

	var repos []storemodels.RepoModel
	if err := db.DB().NewSelect().Model(&repos).Where("is_reserved = 1").Scan(ctx); err != nil {
		t.Fatalf("load reserved repos: %v", err)
	}
	if len(repos) != 2 {
		t.Fatalf("expected two reserved repos, got %d", len(repos))
	}
	for _, repo := range repos {
		if !repo.IsReserved || repo.UserID == "" {
			t.Fatalf("invalid reserved repo: %+v", repo)
		}
	}
}

func TestEnsureAdhocContextRestoresSoftDeletedRecords(t *testing.T) {
	ctx := context.Background()
	db := openRepositoryTestStore(t)
	repository := repositories.NewReservedContextRepository(db.DB())

	want, err := repository.EnsureAdhocContext(ctx, "user-a", "2026-09-24T10:00:00Z")
	if err != nil {
		t.Fatalf("ensure context: %v", err)
	}
	if _, err := db.DB().NewUpdate().Model((*storemodels.RepoModel)(nil)).
		Where("public_id = ?", want.RepoID).
		Set("deleted_at = ?", "2026-09-24T11:00:00Z").
		Exec(ctx); err != nil {
		t.Fatalf("soft-delete repo: %v", err)
	}
	if _, err := db.DB().NewUpdate().Model((*storemodels.StreamModel)(nil)).
		Where("public_id = ?", want.StreamID).
		Set("deleted_at = ?", "2026-09-24T11:00:00Z").
		Exec(ctx); err != nil {
		t.Fatalf("soft-delete stream: %v", err)
	}

	got, err := repository.EnsureAdhocContext(ctx, "user-a", "2026-09-24T12:00:00Z")
	if err != nil {
		t.Fatalf("restore context: %v", err)
	}
	if got != want {
		t.Fatalf("expected restored ids %+v, got %+v", want, got)
	}

	var repo storemodels.RepoModel
	if err := db.DB().NewSelect().Model(&repo).Where("public_id = ?", want.RepoID).Scan(ctx); err != nil {
		t.Fatalf("load restored repo: %v", err)
	}
	if repo.DeletedAt != nil {
		t.Fatalf("expected repo to be restored, got deleted at %q", *repo.DeletedAt)
	}
	var stream storemodels.StreamModel
	if err := db.DB().NewSelect().Model(&stream).Where("public_id = ?", want.StreamID).Scan(ctx); err != nil {
		t.Fatalf("load restored stream: %v", err)
	}
	if stream.DeletedAt != nil {
		t.Fatalf("expected stream to be restored, got deleted at %q", *stream.DeletedAt)
	}
}

func TestNormalListsHideReservedEntities(t *testing.T) {
	ctx := context.Background()
	db := openRepositoryTestStore(t)
	reserved := repositories.NewReservedContextRepository(db.DB())
	contextIDs, err := reserved.EnsureAdhocContext(ctx, "user-a", "2026-09-24T10:00:00Z")
	if err != nil {
		t.Fatalf("ensure context: %v", err)
	}

	repos := repositories.NewRepoRepository(db.DB())
	streams := repositories.NewStreamRepository(db.DB())
	normalRepo, err := repos.Create(ctx, sharedtypes.Repo{ID: 100, Name: "normal"}, "user-a", "2026-09-24T10:00:00Z")
	if err != nil {
		t.Fatalf("create normal repo: %v", err)
	}
	if _, err := streams.Create(ctx, sharedtypes.Stream{
		ID:         100,
		RepoID:     normalRepo.ID,
		Name:       "normal-stream",
		Visibility: sharedtypes.StreamVisibilityPersonal,
	}, "user-a", "2026-09-24T10:00:00Z"); err != nil {
		t.Fatalf("create normal stream: %v", err)
	}

	listedRepos, err := repos.List(ctx, "user-a")
	if err != nil {
		t.Fatalf("list repos: %v", err)
	}
	if len(listedRepos) != 1 || listedRepos[0].ID != normalRepo.ID {
		t.Fatalf("expected only normal repo, got %+v", listedRepos)
	}
	listedStreams, err := streams.ListByRepo(ctx, normalRepo.ID, "user-a")
	if err != nil {
		t.Fatalf("list streams: %v", err)
	}
	if len(listedStreams) != 1 || listedStreams[0].ID != 100 {
		t.Fatalf("expected only normal stream, got %+v", listedStreams)
	}
	if got, err := repos.GetByID(ctx, contextIDs.RepoID, "user-a"); err != nil {
		t.Fatalf("get reserved repo: %v", err)
	} else if got != nil {
		t.Fatalf("expected reserved repo to be hidden, got %+v", got)
	}
	if got, err := streams.GetByID(ctx, contextIDs.StreamID, "user-a"); err != nil {
		t.Fatalf("get reserved stream: %v", err)
	} else if got != nil {
		t.Fatalf("expected reserved stream to be hidden, got %+v", got)
	}
}

func TestEnsureAdhocContextRequiresUserID(t *testing.T) {
	db := openRepositoryTestStore(t)
	_, err := repositories.NewReservedContextRepository(db.DB()).EnsureAdhocContext(
		context.Background(),
		" ",
		"2026-09-24T10:00:00Z",
	)
	if err == nil {
		t.Fatal("expected empty user id to fail")
	}
}

func openRepositoryTestStore(t *testing.T) *storedb.Store {
	t.Helper()
	db, err := storedb.Open(filepath.Join(t.TempDir(), "crona.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := migrations.InitSchema(context.Background(), db.DB()); err != nil {
		t.Fatalf("init schema: %v", err)
	}
	return db
}
