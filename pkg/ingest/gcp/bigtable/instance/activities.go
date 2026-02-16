package instance

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/activity"
	"google.golang.org/api/option"
	"google.golang.org/grpc"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// Activities holds dependencies for Temporal activities.
type Activities struct {
	configService *config.Service
	entClient     *ent.Client
	limiter       ratelimit.Limiter
}

// NewActivities creates a new Activities instance.
func NewActivities(configService *config.Service, entClient *ent.Client, limiter ratelimit.Limiter) *Activities {
	return &Activities{
		configService: configService,
		entClient:     entClient,
		limiter:       limiter,
	}
}

// createClient creates a rate-limited GCP client with credentials.
func (a *Activities) createClient(ctx context.Context, projectID string) (*Client, error) {
	var opts []option.ClientOption
	if credJSON := a.configService.GCPCredentialsJSON(); len(credJSON) > 0 {
		opts = append(opts, option.WithAuthCredentialsJSON(option.ServiceAccount, credJSON))
	}
	opts = append(opts, option.WithGRPCDialOption(
		grpc.WithUnaryInterceptor(ratelimit.UnaryInterceptor(a.limiter)),
	))
	return NewClient(ctx, projectID, opts...)
}

// IngestBigtableInstancesParams contains parameters for the ingest activity.
type IngestBigtableInstancesParams struct {
	ProjectID string
}

// IngestBigtableInstancesResult contains the result of the ingest activity.
type IngestBigtableInstancesResult struct {
	ProjectID      string
	InstanceCount  int
	DurationMillis int64
}

// IngestBigtableInstancesActivity is the activity function reference for workflow registration.
var IngestBigtableInstancesActivity = (*Activities).IngestBigtableInstances

// IngestBigtableInstances is a Temporal activity that ingests Bigtable instances.
func (a *Activities) IngestBigtableInstances(ctx context.Context, params IngestBigtableInstancesParams) (*IngestBigtableInstancesResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting Bigtable instance ingestion",
		"projectID", params.ProjectID,
	)

	client, err := a.createClient(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}
	defer client.Close()

	service := NewService(client, a.entClient)
	result, err := service.Ingest(ctx, IngestParams{
		ProjectID: params.ProjectID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to ingest Bigtable instances: %w", err)
	}

	if err := service.DeleteStaleInstances(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale Bigtable instances", "error", err)
	}

	logger.Info("Completed Bigtable instance ingestion",
		"projectID", params.ProjectID,
		"instanceCount", result.InstanceCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestBigtableInstancesResult{
		ProjectID:      result.ProjectID,
		InstanceCount:  result.InstanceCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
