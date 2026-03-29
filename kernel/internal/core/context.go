package core

import (
	"context"

	"crona/kernel/internal/events"
	"crona/kernel/internal/health"
	"crona/kernel/internal/store"
	"crona/kernel/internal/store/repositories"
)

type Context struct {
	Store            *store.Store
	Repos            *repositories.RepoRepository
	Streams          *repositories.StreamRepository
	Issues           *repositories.IssueRepository
	Habits           *repositories.HabitRepository
	HabitCompletions *repositories.HabitCompletionRepository
	Sessions         *repositories.SessionRepository
	Stash            *repositories.StashRepository
	Ops              *repositories.OpRepository
	Health           *health.Service
	CoreSettings     *repositories.CoreSettingsRepository
	SessionSegments  *repositories.SessionSegmentRepository
	ActiveContext    *repositories.ActiveContextRepository
	ScratchPads      *repositories.ScratchPadRepository
	DailyCheckIns    *repositories.DailyCheckInRepository
	DailyPlans       *repositories.DailyPlanRepository

	UserID     string
	DeviceID   string
	ScratchDir string
	Now        func() string
	Events     *events.Bus
}

func NewContext(db *store.Store, registry *store.Registry, userID string, deviceID string, scratchDir string, now func() string, bus *events.Bus) *Context {
	return &Context{
		Store:            db,
		Repos:            registry.Repos,
		Streams:          registry.Streams,
		Issues:           registry.Issues,
		Habits:           registry.Habits,
		HabitCompletions: registry.HabitCompletions,
		Sessions:         registry.Sessions,
		Stash:            registry.Stash,
		Ops:              registry.Ops,
		Health:           health.NewService(db.Ping),
		CoreSettings:     registry.CoreSettings,
		SessionSegments:  registry.SessionSegments,
		ActiveContext:    registry.ActiveContext,
		ScratchPads:      registry.ScratchPads,
		DailyCheckIns:    registry.DailyCheckIns,
		DailyPlans:       registry.DailyPlans,
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
