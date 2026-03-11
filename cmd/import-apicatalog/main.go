package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"danny.vn/hotpot/pkg/base/app"
	"danny.vn/hotpot/pkg/base/logger"
	entapicatalog "danny.vn/hotpot/pkg/storage/ent/apicatalog"

	// CSV import logic is inlined here to avoid Temporal dependency.
	"danny.vn/hotpot/pkg/ingest/apicatalog"
)

func main() {
	slog.SetDefault(logger.New(slog.LevelInfo))

	filePath := flag.String("file", "", "path to CSV file")
	logSourceID := flag.String("log-source-id", "", "optional log source ID")
	flag.Parse()

	if *filePath == "" {
		fmt.Fprintln(os.Stderr, "Usage: import-api-endpoints -file <csv-path> [-log-source-id <id>]")
		os.Exit(1)
	}

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

	driver := application.Driver()
	entClient := entapicatalog.NewClient(
		entapicatalog.Driver(driver),
		entapicatalog.AlternateSchema(entapicatalog.DefaultSchemaConfig()),
	)

	activities := apicatalog.NewActivities(application.ConfigService(), entClient)

	start := time.Now()
	result, err := activities.ImportCSV(ctx, apicatalog.ImportCSVParams{
		FilePath:    *filePath,
		LogSourceID: *logSourceID,
	})
	if err != nil {
		slog.Error("Import failed", "error", err)
		os.Exit(1)
	}

	slog.Info("Import complete",
		"created", result.Created,
		"updated", result.Updated,
		"duration", time.Since(start).Round(time.Millisecond))
}
