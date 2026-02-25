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

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/logger"
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// Run starts the ingest workers.
// The context is used to signal shutdown - when cancelled, workers will stop.
func Run(ctx context.Context, configService *config.Service, entClient *ent.Client, driver dialect.Driver) error {
	// Set colored logger as default for app-level logging (INFO+).
	slog.SetDefault(logger.New(slog.LevelInfo))

	// Temporal SDK is noisy at INFO — only show WARN+.
	temporalLogger := sdklog.NewStructuredLogger(logger.New(slog.LevelWarn))

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

	// Create paused daily schedules for enabled providers.
	ensureSchedules(ctx, temporalClient, allProviders, configService)

	// Convert context cancellation to interrupt channel for Temporal worker
	interruptCh := make(chan any)
	go func() {
		<-ctx.Done()
		close(interruptCh)
	}()

	// Run workers concurrently
	g, _ := errgroup.WithContext(ctx)

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
		if p.RegisterWithDriver != nil {
			closer = p.RegisterWithDriver(w, configService, driver)
		} else {
			closer = p.Register(w, configService, entClient)
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
