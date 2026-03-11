package detect

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

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/logger"
	hotpottemporal "danny.vn/hotpot/pkg/base/temporal"
	detecthttpmon "danny.vn/hotpot/pkg/detect/httpmonitor"
	"danny.vn/hotpot/pkg/detect/lifecycle"
)

// Run starts the detect worker.
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

	db, err := sql.Open("pgx", configService.DatabaseDSN())
	if err != nil {
		return fmt.Errorf("open database for reads: %w", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("database unreachable: %w", err)
	}

	w := worker.New(temporalClient, "detect", worker.Options{})

	Register(w, configService, driver, db)

	ensureSchedules(ctx, temporalClient)

	slog.Info("Detect worker started", "taskQueue", "detect")

	interruptCh := make(chan any)
	go func() {
		<-ctx.Done()
		close(interruptCh)
	}()

	return w.Run(interruptCh)
}

func ensureSchedules(ctx context.Context, temporalClient client.Client) {
	sc := temporalClient.ScheduleClient()

	hotpottemporal.EnsureSchedule(ctx, sc, client.ScheduleOptions{
		ID: "hotpot-detect-lifecycle-daily",
		Spec: client.ScheduleSpec{
			Intervals: []client.ScheduleIntervalSpec{
				{Every: 24 * time.Hour},
			},
		},
		Action: &client.ScheduleWorkflowAction{
			ID:        "hotpot-detect-lifecycle",
			Workflow:  lifecycle.SoftwareLifecycleWorkflow,
			TaskQueue: "detect",
		},
		Paused: true,
	})

	hotpottemporal.EnsureSchedule(ctx, sc, client.ScheduleOptions{
		ID: "hotpot-detect-lifecycle-os-daily",
		Spec: client.ScheduleSpec{
			Intervals: []client.ScheduleIntervalSpec{
				{Every: 24 * time.Hour},
			},
		},
		Action: &client.ScheduleWorkflowAction{
			ID:        "hotpot-detect-lifecycle-os",
			Workflow:  lifecycle.OSLifecycleWorkflow,
			TaskQueue: "detect",
		},
		Paused: true,
	})

	hotpottemporal.EnsureSchedule(ctx, sc, client.ScheduleOptions{
		ID: "hotpot-detect-httpmonitor-5min",
		Spec: client.ScheduleSpec{
			Intervals: []client.ScheduleIntervalSpec{
				{Every: 5 * time.Minute},
			},
		},
		Action: &client.ScheduleWorkflowAction{
			ID:        "hotpot-detect-httpmonitor",
			Workflow:  detecthttpmon.HttpMonitorAnomalyWorkflow,
			TaskQueue: "detect",
		},
		Paused: true,
	})
}

