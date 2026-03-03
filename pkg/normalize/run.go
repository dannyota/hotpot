package normalize

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"entgo.io/ent/dialect"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.temporal.io/sdk/client"
	sdklog "go.temporal.io/sdk/log"
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/logger"
	"github.com/dannyota/hotpot/pkg/normalize/machine"
)

// Run starts the normalize worker.
func Run(ctx context.Context, configService *config.Service, driver dialect.Driver) error {
	level := configService.LogLevel()
	slog.SetDefault(logger.New(level))

	temporalLevel := max(level, slog.LevelInfo)
	temporalLogger := sdklog.NewStructuredLogger(logger.New(temporalLevel))

	temporalClient, err := client.Dial(client.Options{
		HostPort:  configService.TemporalHostPort(),
		Namespace: configService.TemporalNamespace(),
		Logger:    temporalLogger,
	})
	if err != nil {
		return fmt.Errorf("failed to create Temporal client: %w", err)
	}
	defer temporalClient.Close()

	// Open raw SQL connection for provider Load functions reading bronze.
	db, err := sql.Open("pgx", configService.DatabaseDSN())
	if err != nil {
		return fmt.Errorf("open database for bronze reads: %w", err)
	}
	defer db.Close()

	w := worker.New(temporalClient, "normalize", worker.Options{})

	Register(w, configService, driver, db)

	ensureSchedules(ctx, temporalClient)

	slog.Info("Normalize worker started", "taskQueue", "normalize")

	interruptCh := make(chan any)
	go func() {
		<-ctx.Done()
		close(interruptCh)
	}()

	return w.Run(interruptCh)
}

// ensureSchedules creates paused schedules for normalize workflows.
// Existing schedules are left unchanged.
func ensureSchedules(ctx context.Context, temporalClient client.Client) {
	sc := temporalClient.ScheduleClient()

	scheduleID := "hotpot-normalize-machines-daily"

	// Skip if schedule already exists.
	handle := sc.GetHandle(ctx, scheduleID)
	if _, err := handle.Describe(ctx); err == nil {
		slog.Info("Schedule already exists", "schedule", scheduleID)
		return
	}

	_, err := sc.Create(ctx, client.ScheduleOptions{
		ID: scheduleID,
		Spec: client.ScheduleSpec{
			Intervals: []client.ScheduleIntervalSpec{
				{Every: 24 * time.Hour},
			},
		},
		Action: &client.ScheduleWorkflowAction{
			ID:        "hotpot-normalize-machines",
			Workflow:  machine.NormalizeMachinesWorkflow,
			TaskQueue: "normalize",
		},
		Paused: true,
	})
	if err != nil {
		slog.Error("Failed to create schedule", "schedule", scheduleID, "error", err)
		return
	}

	slog.Info("Created paused schedule", "schedule", scheduleID)
}
