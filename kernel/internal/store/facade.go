package store

import (
	"context"

	storedb "crona/kernel/internal/store/db"
	devtools "crona/kernel/internal/store/devtools"
	migrations "crona/kernel/internal/store/migrations"
	"crona/kernel/internal/store/repositories"

	"github.com/uptrace/bun"
)

type Store = storedb.Store

type Registry struct {
	Repos            *repositories.RepoRepository
	Streams          *repositories.StreamRepository
	Issues           *repositories.IssueRepository
	Habits           *repositories.HabitRepository
	HabitCompletions *repositories.HabitCompletionRepository
	Sessions         *repositories.SessionRepository
	Stash            *repositories.StashRepository
	Ops              *repositories.OpRepository
	CoreSettings     *repositories.CoreSettingsRepository
	SessionSegments  *repositories.SessionSegmentRepository
	ActiveContext    *repositories.ActiveContextRepository
	ScratchPads      *repositories.ScratchPadRepository
	DailyCheckIns    *repositories.DailyCheckInRepository
}

func Open(dbPath string) (*Store, error) {
	return storedb.Open(dbPath)
}

func InitSchema(ctx context.Context, db *bun.DB) error {
	return migrations.InitSchema(ctx, db)
}

func ClearAllData(ctx context.Context, db *bun.DB) error {
	return devtools.ClearAllData(ctx, db)
}

func NewRegistry(db *bun.DB) *Registry {
	return &Registry{
		Repos:            repositories.NewRepoRepository(db),
		Streams:          repositories.NewStreamRepository(db),
		Issues:           repositories.NewIssueRepository(db),
		Habits:           repositories.NewHabitRepository(db),
		HabitCompletions: repositories.NewHabitCompletionRepository(db),
		Sessions:         repositories.NewSessionRepository(db),
		Stash:            repositories.NewStashRepository(db),
		Ops:              repositories.NewOpRepository(db),
		CoreSettings:     repositories.NewCoreSettingsRepository(db),
		SessionSegments:  repositories.NewSessionSegmentRepository(db),
		ActiveContext:    repositories.NewActiveContextRepository(db),
		ScratchPads:      repositories.NewScratchPadRepository(db),
		DailyCheckIns:    repositories.NewDailyCheckInRepository(db),
	}
}
