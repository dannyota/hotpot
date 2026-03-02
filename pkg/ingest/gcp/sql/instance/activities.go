package instance

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/activity"
	"google.golang.org/api/option"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/gcpauth"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/base/temporalerr"
	entgcpsql "github.com/dannyota/hotpot/pkg/storage/ent/gcp/sql"
)

// Activities holds dependencies for Temporal activities.
type Activities struct {
	configService *config.Service
	entClient     *entgcpsql.Client
	limiter       ratelimit.Limiter
}

// NewActivities creates a new Activities instance.
func NewActivities(configService *config.Service, entClient *entgcpsql.Client, limiter ratelimit.Limiter) *Activities {
	return &Activities{
		configService: configService,
		entClient:     entClient,
		limiter:       limiter,
	}
}

// createClient creates a rate-limited GCP client with credentials.
func (a *Activities) createClient(ctx context.Context) (*Client, error) {
	httpClient, err := gcpauth.NewHTTPClient(ctx, a.configService.GCPCredentialsJSON(), a.limiter)
	if err != nil {
		return nil, err
	}
	return NewClient(ctx, option.WithHTTPClient(httpClient))
}

// IngestSQLInstancesParams contains parameters for the ingest activity.
type IngestSQLInstancesParams struct {
	ProjectID string
}

// IngestSQLInstancesResult contains the result of the ingest activity.
type IngestSQLInstancesResult struct {
	ProjectID      string
	InstanceCount  int
	DurationMillis int64
}

// IngestSQLInstancesActivity is the activity function reference for workflow registration.
var IngestSQLInstancesActivity = (*Activities).IngestSQLInstances

// IngestSQLInstances is a Temporal activity that ingests GCP Cloud SQL instances.
func (a *Activities) IngestSQLInstances(ctx context.Context, params IngestSQLInstancesParams) (*IngestSQLInstancesResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GCP Cloud SQL instance ingestion",
		"projectID", params.ProjectID,
	)

	// Create client for this activity
	client, err := a.createClient(ctx)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("create client: %w", err))
	}
	defer client.Close()

	// Create service
	service := NewService(client, a.entClient)
	result, err := service.Ingest(ctx, IngestParams{
		ProjectID: params.ProjectID,
	})
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("failed to ingest SQL instances: %w", err))
	}

	// Delete stale instances
	if err := service.DeleteStaleInstances(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale SQL instances", "error", err)
	}

	logger.Info("Completed GCP Cloud SQL instance ingestion",
		"projectID", params.ProjectID,
		"instanceCount", result.InstanceCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestSQLInstancesResult{
		ProjectID:      result.ProjectID,
		InstanceCount:  result.InstanceCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
