package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"crona/kernel/internal/app"
	versionpkg "crona/shared/version"
)

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Println(versionpkg.Current())
		return
	}
	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(signals)
	go func() {
		sig := <-signals
		cancel(fmt.Errorf("received %s", sig))
	}()

	if err := app.Run(ctx); err != nil {
		log.Printf("crona-kernel: %v", err)
		os.Exit(1)
	}
}
