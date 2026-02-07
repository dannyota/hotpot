package targetinstance

import (
	"context"
	"fmt"
	"net/http"

	"go.temporal.io/sdk/activity"
	"golang.org/x/time/rate"
	"google.golang.org/api/option"
	"gorm.io/gorm"

	"hotpot/pkg/base/config"
	"hotpot/pkg/base/ratelimit"
)

// Activities holds dependencies for Temporal activities.
type Activities struct {
	configService *config.Service
	db            *gorm.DB
	limiter       *rate.Limiter
}

// NewActivities creates a new Activities instance.
func NewActivities(configService *config.Service, db *gorm.DB, limiter *rate.Limiter) *Activities {
	return &Activities{
		configService: configService,
		db:            db,
		limiter:       limiter,
	}
}

// createClient creates a rate-limited GCP client with credentials.
func (a *Activities) createClient(ctx context.Context) (*Client, error) {
	var opts []option.ClientOption
	if credJSON := a.configService.GCPCredentialsJSON(); len(credJSON) > 0 {
		opts = append(opts, option.WithAuthCredentialsJSON(option.ServiceAccount, credJSON))
	}
	opts = append(opts, option.WithHTTPClient(&http.Client{
		Transport: ratelimit.NewRateLimitedTransport(a.limiter, nil),
	}))
	return NewClient(ctx, opts...)
}

// IngestComputeTargetInstancesParams contains parameters for the ingest activity.
type IngestComputeTargetInstancesParams struct {
	ProjectID string
}

// IngestComputeTargetInstancesResult contains the result of the ingest activity.
type IngestComputeTargetInstancesResult struct {
	ProjectID           string
	TargetInstanceCount int
	DurationMillis      int64
}

// IngestComputeTargetInstancesActivity is the activity function reference for workflow registration.
var IngestComputeTargetInstancesActivity = (*Activities).IngestComputeTargetInstances

// IngestComputeTargetInstances is a Temporal activity that ingests GCP Compute target instances.
func (a *Activities) IngestComputeTargetInstances(ctx context.Context, params IngestComputeTargetInstancesParams) (*IngestComputeTargetInstancesResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GCP Compute target instance ingestion",
		"projectID", params.ProjectID,
	)

	// Create client for this activity
	client, err := a.createClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}
	defer client.Close()

	// Create service
	service := NewService(client, a.db)
	result, err := service.Ingest(ctx, IngestParams{
		ProjectID: params.ProjectID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to ingest target instances: %w", err)
	}

	// Delete stale target instances
	if err := service.DeleteStaleTargetInstances(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale target instances", "error", err)
	}

	logger.Info("Completed GCP Compute target instance ingestion",
		"projectID", params.ProjectID,
		"targetInstanceCount", result.TargetInstanceCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestComputeTargetInstancesResult{
		ProjectID:           result.ProjectID,
		TargetInstanceCount: result.TargetInstanceCount,
		DurationMillis:      result.DurationMillis,
	}, nil
}
