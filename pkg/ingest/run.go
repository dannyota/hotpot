package ingest

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"entgo.io/ent/dialect"
	"go.temporal.io/sdk/client"
	sdklog "go.temporal.io/sdk/log"
	"go.temporal.io/sdk/worker"
	"golang.org/x/sync/errgroup"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/logger"
)

// Run starts the ingest workers.
// The context is used to signal shutdown - when cancelled, workers will stop.
func Run(ctx context.Context, configService *config.Service, driver dialect.Driver) error {
	// Set colored logger as default for app-level logging.
	level := configService.LogLevel()
	slog.SetDefault(logger.New(level))

	// Temporal logger controls workflow/activity log output.
	// Cap at INFO minimum — Temporal SDK internals are noisy at DEBUG.
	temporalLevel := max(level, slog.LevelInfo)
	temporalLogger := sdklog.NewStructuredLogger(logger.New(temporalLevel))

	allProviders := Providers()
	if len(allProviders) == 0 {
		return fmt.Errorf("no providers registered; import at least one provider package in cmd/ingest/main.go")
	}

	// Create Temporal client
	temporalClient, err := client.Dial(client.Options{
		HostPort:  configService.TemporalHostPort(),
		Namespace: configService.TemporalNamespace(),
		Logger:    temporalLogger,
	})
	if err != nil {
		return fmt.Errorf("failed to create Temporal client: %w", err)
	}
	defer temporalClient.Close()

	// Store client for activities that need to signal workflows.
	configService.SetTemporalClient(temporalClient)

	// Create paused daily schedules for enabled providers.
	ensureSchedules(ctx, temporalClient, allProviders, configService)

	// Convert context cancellation to interrupt channel for Temporal worker
	interruptCh := make(chan any)
	go func() {
		<-ctx.Done()
		close(interruptCh)
	}()

	// Utility worker for GeoIP downloads — always runs regardless of providers.
	utilWorker := worker.New(temporalClient, "hotpot-ingest-geoip", worker.Options{})
	geoipAct := &GeoIPActivities{configService: configService}
	utilWorker.RegisterActivity(geoipAct.UpdateGeoIPFiles)
	utilWorker.RegisterWorkflow(UpdateGeoIPWorkflow)

	// Run workers concurrently
	var g errgroup.Group

	g.Go(func() error {
		if err := utilWorker.Run(interruptCh); err != nil {
			return fmt.Errorf("ingest utility worker failed: %w", err)
		}
		return nil
	})

	var started int
	for _, p := range allProviders {
		if !p.Enabled(configService) {
			continue
		}
		started++

		reqPerMin := p.RateLimitPerMinute(configService)
		activitiesPerSec := float64(reqPerMin) / 60.0

		w := worker.New(temporalClient, p.TaskQueue, worker.Options{
			TaskQueueActivitiesPerSecond: activitiesPerSec,
		})

		var closer io.Closer
		if p.Register != nil {
			closer = p.Register(w, configService, driver)
		}
		if closer != nil {
			defer closer.Close()
		}

		slog.Info(fmt.Sprintf("%s worker enabled", p.Name),
			"rate_limit", fmt.Sprintf("%d rpm", reqPerMin),
			"activities_per_sec", fmt.Sprintf("%.1f", activitiesPerSec))

		g.Go(func() error {
			if err := w.Run(interruptCh); err != nil {
				return fmt.Errorf("%s worker failed: %w", p.Name, err)
			}
			return nil
		})
	}

	if started == 0 {
		names := make([]string, len(allProviders))
		for i, p := range allProviders {
			names[i] = p.Name
		}
		return fmt.Errorf("no providers enabled in config; registered providers: %v", names)
	}
	slog.Info(fmt.Sprintf("Started %d provider worker(s)", started))

	return g.Wait()
}
