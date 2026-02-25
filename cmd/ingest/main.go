package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/dannyota/hotpot/pkg/base/app"
	"github.com/dannyota/hotpot/pkg/base/logger"
	"github.com/dannyota/hotpot/pkg/ingest"

	// Register providers — add/remove lines to control what's compiled into the binary.
	//_ "github.com/dannyota/hotpot/pkg/ingest/aws"
	_ "github.com/dannyota/hotpot/pkg/ingest/gcp"
	_ "github.com/dannyota/hotpot/pkg/ingest/greennode"
	_ "github.com/dannyota/hotpot/pkg/ingest/sentinelone"
	//_ "github.com/dannyota/hotpot/pkg/ingest/vault"
	// _ "github.com/dannyota/hotpot/pkg/ingest/digitalocean"
)

func main() {
	slog.SetDefault(logger.New(slog.LevelInfo))
	ctx := context.Background()

	application, err := app.New(app.Options{})
	if err != nil {
		slog.Error("Failed to create app", "error", err)
		os.Exit(1)
	}

	if err := application.Start(ctx); err != nil {
		slog.Error("Failed to start", "error", err)
		os.Exit(1)
	}
	defer application.Stop()

	application.Run(ingest.Run)

	if err := application.Wait(); err != nil {
		slog.Error("Service failed", "error", err)
		os.Exit(1)
	}
}
