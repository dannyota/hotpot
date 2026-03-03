package snapshot

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"google.golang.org/api/option"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/gcpauth"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	entcompute "github.com/dannyota/hotpot/pkg/storage/ent/gcp/compute"
)

// Register registers snapshot activities and workflows with a Temporal worker.
func Register(w worker.Worker, configService *config.Service, entClient *entcompute.Client, limiter ratelimit.Limiter) {
	httpClient, err := gcpauth.NewHTTPClient(context.Background(), configService.GCPCredentialsJSON(), limiter)
	if err != nil {
		panic(fmt.Sprintf("snapshot: create GCP HTTP client: %v", err))
	}

	gcpClient, err := NewClient(context.Background(), option.WithHTTPClient(httpClient))
	if err != nil {
		panic(fmt.Sprintf("snapshot: create snapshot client: %v", err))
	}

	temporalClient := configService.TemporalClient().(client.Client)

	activities := NewActivities(gcpClient, entClient, temporalClient)
	w.RegisterActivity(activities.FetchAndSaveSnapshotsPage)
	w.RegisterActivity(activities.DeleteStaleSnapshots)
	w.RegisterWorkflow(GCPComputeSnapshotWorkflow)
}
