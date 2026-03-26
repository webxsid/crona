package app

import (
	"context"
	"fmt"
	"os"
	"time"

	"crona/kernel/internal/core"
	corecommands "crona/kernel/internal/core/commands"
	"crona/kernel/internal/events"
	"crona/kernel/internal/export"
	"crona/kernel/internal/ipc"
	"crona/kernel/internal/notify"
	"crona/kernel/internal/runtime"
	"crona/kernel/internal/store"
	"crona/kernel/internal/updatecheck"
	"crona/shared/config"
	"crona/shared/localipc"
	sharedtypes "crona/shared/types"
)

func Run(ctx context.Context) error {
	appEnv := config.Load()

	paths, err := runtime.ResolvePaths()
	if err != nil {
		return fmt.Errorf("resolve runtime paths: %w", err)
	}
	if err := runtime.EnsurePaths(paths); err != nil {
		return fmt.Errorf("ensure runtime paths: %w", err)
	}

	logger := runtime.NewLogger(paths)
	startedAt := time.Now().UTC().Format(time.RFC3339)

	dbStore, err := store.Open(paths.DBPath)
	if err != nil {
		return fmt.Errorf("open sqlite store: %w", err)
	}
	defer func() {
		if err := dbStore.Close(); err != nil {
			logger.Error("close sqlite store", err)
		}
	}()

	if err := dbStore.Ping(ctx); err != nil {
		return fmt.Errorf("ping sqlite store: %w", err)
	}
	if err := store.InitSchema(ctx, dbStore.DB()); err != nil {
		return fmt.Errorf("init sqlite schema: %w", err)
	}

	bus := events.NewBus()
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	registry := store.NewRegistry(dbStore.DB())

	commandCtx := core.NewContext(
		dbStore,
		registry,
		"local",
		hostnameOr("device-1"),
		paths.ScratchDir,
		func() string { return time.Now().UTC().Format(time.RFC3339) },
		bus,
	)
	if err := commandCtx.InitDefaults(runCtx); err != nil {
		return fmt.Errorf("init command defaults: %w", err)
	}
	_ = notify.Start(runCtx, commandCtx, bus, logger)
	updater := updatecheck.Start(runCtx, commandCtx, bus, logger, paths, appEnv.Mode)
	if _, err := export.EnsureAssets(paths); err != nil {
		return fmt.Errorf("ensure export assets: %w", err)
	}

	info := sharedtypes.KernelInfo{
		PID:            os.Getpid(),
		Transport:      paths.Transport,
		Endpoint:       paths.Endpoint,
		SocketPath:     paths.SocketPath,
		Token:          "",
		StartedAt:      startedAt,
		ScratchDir:     paths.ScratchDir,
		Env:            appEnv.Mode,
		ExecutablePath: executablePath(),
	}

	server := ipc.NewServer(paths.Transport, paths.Endpoint, NewHandler(startedAt, info, dbStore.Ping, commandCtx, bus, cancel, appEnv.Mode, paths, updater), logger)
	timer := corecommands.GetTimerService(commandCtx)
	if err := timer.RecoverBoundary(runCtx); err != nil {
		return fmt.Errorf("recover timer boundary: %w", err)
	}
	if err := server.Start(); err != nil {
		return fmt.Errorf("start ipc server: %w", err)
	}
	defer func() {
		if err := server.Close(); err != nil {
			logger.Error("close ipc server", err)
		}
	}()

	if err := runtime.WriteKernelInfo(paths, info); err != nil {
		return fmt.Errorf("write kernel info: %w", err)
	}
	defer func() {
		if err := runtime.ClearKernelInfo(paths); err != nil {
			logger.Error("clear kernel info", err)
		}
	}()

	logger.Info("kernel listening on " + localipc.Label(paths.Transport, paths.Endpoint))

	<-runCtx.Done()
	logger.Info("kernel shutting down")
	return nil
}

func hostnameOr(fallback string) string {
	name, err := os.Hostname()
	if err != nil || name == "" {
		return fallback
	}
	return name
}

func executablePath() string {
	path, err := os.Executable()
	if err != nil {
		return ""
	}
	return path
}
