package main

import (
	"context"
	"log/slog"
	"os"

	"danny.vn/hotpot/pkg/base/app"
	"danny.vn/hotpot/pkg/base/logger"
	"danny.vn/hotpot/pkg/normalize"
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

	application.Run(normalize.Run)

	if err := application.Wait(); err != nil {
		slog.Error("Service failed", "error", err)
		os.Exit(1)
	}
}
