package ingest

import (
	"context"
	"fmt"
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"golang.org/x/sync/errgroup"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// Run starts the ingest workers.
// The context is used to signal shutdown - when cancelled, workers will stop.
func Run(ctx context.Context, configService *config.Service, entClient *ent.Client) error {
	allProviders := Providers()
	if len(allProviders) == 0 {
		return fmt.Errorf("no providers registered; import at least one provider package in cmd/ingest/main.go")
	}

	// Create Temporal client
	temporalClient, err := client.Dial(client.Options{
		HostPort:  configService.TemporalHostPort(),
		Namespace: configService.TemporalNamespace(),
	})
	if err != nil {
		return fmt.Errorf("failed to create Temporal client: %w", err)
	}
	defer temporalClient.Close()

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

		closer := p.Register(w, configService, entClient)
		if closer != nil {
			defer closer.Close()
		}

		log.Printf("%s worker enabled (rate limit: %d rpm, Temporal: %.1f activities/sec)",
			p.Name, reqPerMin, activitiesPerSec)

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
	log.Printf("Started %d provider worker(s)", started)

	return g.Wait()
}
