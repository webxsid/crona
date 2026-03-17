package core

import (
	"context"

	"crona/kernel/internal/events"
	"crona/kernel/internal/health"
	"crona/kernel/internal/store"
)

type Context struct {
	Store            *store.Store
	Repos            *store.RepoRepository
	Streams          *store.StreamRepository
	Issues           *store.IssueRepository
	Habits           *store.HabitRepository
	HabitCompletions *store.HabitCompletionRepository
	Sessions         *store.SessionRepository
	Stash            *store.StashRepository
	Ops              *store.OpRepository
	Health           *health.Service
	CoreSettings     *store.CoreSettingsRepository
	SessionSegments  *store.SessionSegmentRepository
	ActiveContext    *store.ActiveContextRepository
	ScratchPads      *store.ScratchPadRepository
	DailyCheckIns    *store.DailyCheckInRepository

	UserID     string
	DeviceID   string
	ScratchDir string
	Now        func() string
	Events     *events.Bus
}

func NewContext(db *store.Store, userID string, deviceID string, scratchDir string, now func() string, bus *events.Bus) *Context {
	return &Context{
		Store:            db,
		Repos:            store.NewRepoRepository(db.DB()),
		Streams:          store.NewStreamRepository(db.DB()),
		Issues:           store.NewIssueRepository(db.DB()),
		Habits:           store.NewHabitRepository(db.DB()),
		HabitCompletions: store.NewHabitCompletionRepository(db.DB()),
		Sessions:         store.NewSessionRepository(db.DB()),
		Stash:            store.NewStashRepository(db.DB()),
		Ops:              store.NewOpRepository(db.DB()),
		Health:           health.NewService(db.Ping),
		CoreSettings:     store.NewCoreSettingsRepository(db.DB()),
		SessionSegments:  store.NewSessionSegmentRepository(db.DB()),
		ActiveContext:    store.NewActiveContextRepository(db.DB()),
		ScratchPads:      store.NewScratchPadRepository(db.DB()),
		DailyCheckIns:    store.NewDailyCheckInRepository(db.DB()),
		UserID:           userID,
		DeviceID:         deviceID,
		ScratchDir:       scratchDir,
		Now:              now,
		Events:           bus,
	}
}

func (c *Context) InitDefaults(ctx context.Context) error {
	if err := c.CoreSettings.InitializeDefaults(ctx, c.UserID, c.DeviceID); err != nil {
		return err
	}
	return c.ActiveContext.InitializeDefaults(ctx, c.UserID, c.DeviceID)
}
