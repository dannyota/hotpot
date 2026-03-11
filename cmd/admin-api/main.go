package main

import (
	"context"
	"database/sql"
	"log/slog"
	"os"

	"entgo.io/ent/dialect"

	"danny.vn/hotpot/pkg/admin"
	"danny.vn/hotpot/pkg/admin/bronze"
	"danny.vn/hotpot/pkg/admin/gold"
	"danny.vn/hotpot/pkg/admin/silver"
	"danny.vn/hotpot/pkg/admin/stats"
	"danny.vn/hotpot/pkg/base/app"
	"danny.vn/hotpot/pkg/base/logger"
)

func main() {
	slog.SetDefault(logger.New(slog.LevelInfo))
	ctx := context.Background()

	admin.RegisterAll = func(driver dialect.Driver, db *sql.DB) {
		bronze.Register(driver, db)
		silver.Register(driver, db)
		gold.Register(driver, db)
		stats.Register(db)
	}

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

	application.Run(admin.RunAPI)

	if err := application.Wait(); err != nil {
		slog.Error("Admin API server failed", "error", err)
		os.Exit(1)
	}
}
