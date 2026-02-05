package ingest

import (
	"context"
	"fmt"
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"gorm.io/gorm"

	"hotpot/pkg/base/config"
	"hotpot/pkg/ingest/gcp"
)

// TaskQueues for different providers.
const (
	TaskQueueGCP      = "hotpot-ingest-gcp"
	TaskQueueVNGCloud = "hotpot-ingest-vng"
	TaskQueueS1       = "hotpot-ingest-s1"
	TaskQueueFortinet = "hotpot-ingest-fortinet"
)

// Run starts the ingest workers.
// The context is used to signal shutdown - when cancelled, workers will stop.
func Run(ctx context.Context, configService *config.Service, db *gorm.DB) error {
	cfg := configService.Config()

	// Create Temporal client
	hostPort := cfg.Temporal.HostPort
	if hostPort == "" {
		hostPort = "localhost:7233"
	}
	namespace := cfg.Temporal.Namespace
	if namespace == "" {
		namespace = "default"
	}

	temporalClient, err := client.Dial(client.Options{
		HostPort:  hostPort,
		Namespace: namespace,
	})
	if err != nil {
		return fmt.Errorf("failed to create Temporal client: %w", err)
	}
	defer temporalClient.Close()

	// Create GCP worker with session support
	gcpWorker := worker.New(temporalClient, TaskQueueGCP, worker.Options{
		EnableSessionWorker: true, // Required for workflow sessions
	})

	// Register GCP workflows and activities
	// No cleanup needed - clients managed per workflow session
	gcp.Register(gcpWorker, configService, db)

	// Create and register VNGCloud worker (future)
	// vngWorker := worker.New(temporalClient, TaskQueueVNGCloud, worker.Options{EnableSessionWorker: true})
	// vngcloud.Register(vngWorker, cfg.VNGCloud, db)

	// Create and register SentinelOne worker (future)
	// s1Worker := worker.New(temporalClient, TaskQueueS1, worker.Options{EnableSessionWorker: true})
	// sentinelone.Register(s1Worker, cfg.S1, db)

	// Create and register Fortinet worker (future)
	// fortinetWorker := worker.New(temporalClient, TaskQueueFortinet, worker.Options{EnableSessionWorker: true})
	// fortinet.Register(fortinetWorker, cfg.Fortinet, db)

	log.Println("Starting ingest workers...")

	// Convert context cancellation to interrupt channel for Temporal worker
	interruptCh := make(chan interface{})
	go func() {
		<-ctx.Done()
		close(interruptCh)
	}()

	// Run GCP worker (blocking until interrupted)
	err = gcpWorker.Run(interruptCh)
	if err != nil {
		return fmt.Errorf("GCP worker failed: %w", err)
	}

	return nil
}
