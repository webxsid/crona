package store

import (
	"context"

	storedb "crona/kernel/internal/store/db"
	devtools "crona/kernel/internal/store/devtools"
	migrations "crona/kernel/internal/store/migrations"
	"crona/kernel/internal/store/repositories"
	sharedtypes "crona/shared/types"

	"github.com/uptrace/bun"
)

type Store = storedb.Store

type (
	RepoRepository            = repositories.RepoRepository
	StreamRepository          = repositories.StreamRepository
	IssueRepository           = repositories.IssueRepository
	HabitRepository           = repositories.HabitRepository
	HabitCompletionRepository = repositories.HabitCompletionRepository
	SessionRepository         = repositories.SessionRepository
	StashRepository           = repositories.StashRepository
	OpRepository              = repositories.OpRepository
	CoreSettingsRepository    = repositories.CoreSettingsRepository
	SessionSegmentRepository  = repositories.SessionSegmentRepository
	ActiveContextRepository   = repositories.ActiveContextRepository
	ScratchPadRepository      = repositories.ScratchPadRepository
	DailyCheckInRepository    = repositories.DailyCheckInRepository
)

type Patch[T any] = sharedtypes.Patch[T]

func Open(dbPath string) (*Store, error) {
	return storedb.Open(dbPath)
}

func InitSchema(ctx context.Context, db *bun.DB) error {
	return migrations.InitSchema(ctx, db)
}

func ClearAllData(ctx context.Context, db *bun.DB) error {
	return devtools.ClearAllData(ctx, db)
}

func NewRepoRepository(db *bun.DB) *RepoRepository {
	return repositories.NewRepoRepository(db)
}

func NewStreamRepository(db *bun.DB) *StreamRepository {
	return repositories.NewStreamRepository(db)
}

func NewIssueRepository(db *bun.DB) *IssueRepository {
	return repositories.NewIssueRepository(db)
}

func NewHabitRepository(db *bun.DB) *HabitRepository {
	return repositories.NewHabitRepository(db)
}

func NewHabitCompletionRepository(db *bun.DB) *HabitCompletionRepository {
	return repositories.NewHabitCompletionRepository(db)
}

func NewSessionRepository(db *bun.DB) *SessionRepository {
	return repositories.NewSessionRepository(db)
}

func NewStashRepository(db *bun.DB) *StashRepository {
	return repositories.NewStashRepository(db)
}

func NewOpRepository(db *bun.DB) *OpRepository {
	return repositories.NewOpRepository(db)
}

func NewCoreSettingsRepository(db *bun.DB) *CoreSettingsRepository {
	return repositories.NewCoreSettingsRepository(db)
}

func NewSessionSegmentRepository(db *bun.DB) *SessionSegmentRepository {
	return repositories.NewSessionSegmentRepository(db)
}

func NewActiveContextRepository(db *bun.DB) *ActiveContextRepository {
	return repositories.NewActiveContextRepository(db)
}

func NewScratchPadRepository(db *bun.DB) *ScratchPadRepository {
	return repositories.NewScratchPadRepository(db)
}

func NewDailyCheckInRepository(db *bun.DB) *DailyCheckInRepository {
	return repositories.NewDailyCheckInRepository(db)
}
