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

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/logger"
	hotpottemporal "danny.vn/hotpot/pkg/base/temporal"
	normhttptraffic "danny.vn/hotpot/pkg/normalize/httptraffic"
	"danny.vn/hotpot/pkg/normalize/inventory/apiendpoint"
	"danny.vn/hotpot/pkg/normalize/inventory/k8snode"
	"danny.vn/hotpot/pkg/normalize/inventory/machine"
	"danny.vn/hotpot/pkg/normalize/inventory/software"
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

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("database unreachable: %w", err)
	}

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

func ensureSchedules(ctx context.Context, temporalClient client.Client) {
	sc := temporalClient.ScheduleClient()

	hotpottemporal.EnsureSchedule(ctx, sc, client.ScheduleOptions{
		ID: "hotpot-normalize-machines-daily",
		Spec: client.ScheduleSpec{
			Intervals: []client.ScheduleIntervalSpec{
				{Every: 24 * time.Hour},
			},
		},
		Action: &client.ScheduleWorkflowAction{
			ID:        "hotpot-normalize-machines",
			Workflow:  machine.NormalizeMachinesWorkflow,
			Args:      []interface{}{machine.NormalizeMachinesWorkflowParams{ProviderKeys: []string{"s1", "meec", "greennode", "gcp"}}},
			TaskQueue: "normalize",
		},
		Paused: true,
	})

	hotpottemporal.EnsureSchedule(ctx, sc, client.ScheduleOptions{
		ID: "hotpot-normalize-k8s-nodes-daily",
		Spec: client.ScheduleSpec{
			Intervals: []client.ScheduleIntervalSpec{
				{Every: 24 * time.Hour},
			},
		},
		Action: &client.ScheduleWorkflowAction{
			ID:        "hotpot-normalize-k8s-nodes",
			Workflow:  k8snode.NormalizeK8sNodesWorkflow,
			Args:      []interface{}{k8snode.NormalizeK8sNodesWorkflowParams{ProviderKeys: []string{"gcp"}}},
			TaskQueue: "normalize",
		},
		Paused: true,
	})

	hotpottemporal.EnsureSchedule(ctx, sc, client.ScheduleOptions{
		ID: "hotpot-normalize-software-daily",
		Spec: client.ScheduleSpec{
			Intervals: []client.ScheduleIntervalSpec{
				{Every: 24 * time.Hour},
			},
		},
		Action: &client.ScheduleWorkflowAction{
			ID:        "hotpot-normalize-software",
			Workflow:  software.NormalizeSoftwareWorkflow,
			Args:      []interface{}{software.NormalizeSoftwareWorkflowParams{ProviderKeys: []string{"s1", "meec"}}},
			TaskQueue: "normalize",
		},
		Paused: true,
	})

	hotpottemporal.EnsureSchedule(ctx, sc, client.ScheduleOptions{
		ID: "hotpot-normalize-api-endpoints-daily",
		Spec: client.ScheduleSpec{
			Intervals: []client.ScheduleIntervalSpec{
				{Every: 24 * time.Hour},
			},
		},
		Action: &client.ScheduleWorkflowAction{
			ID:        "hotpot-normalize-api-endpoints",
			Workflow:  apiendpoint.NormalizeApiEndpointsWorkflow,
			TaskQueue: "normalize",
		},
		Paused: true,
	})

	hotpottemporal.EnsureSchedule(ctx, sc, client.ScheduleOptions{
		ID: "hotpot-normalize-httptraffic-5min",
		Spec: client.ScheduleSpec{
			Intervals: []client.ScheduleIntervalSpec{
				{Every: 5 * time.Minute},
			},
		},
		Action: &client.ScheduleWorkflowAction{
			ID:        "hotpot-normalize-httptraffic",
			Workflow:  normhttptraffic.NormalizeHttptrafficWorkflow,
			TaskQueue: "normalize",
		},
		Paused: true,
	})

}

