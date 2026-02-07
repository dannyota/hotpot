package instance

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/activity"
	"golang.org/x/time/rate"
	"gorm.io/gorm"

	"hotpot/pkg/base/config"
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

// IngestComputeInstancesParams contains parameters for the ingest activity.
type IngestComputeInstancesParams struct {
	SessionID string
	ProjectID string
}

// IngestComputeInstancesResult contains the result of the ingest activity.
type IngestComputeInstancesResult struct {
	ProjectID      string
	InstanceCount  int
	DurationMillis int64
}

// IngestComputeInstancesActivity is the activity function reference for workflow registration.
var IngestComputeInstancesActivity = (*Activities).IngestComputeInstances

// IngestComputeInstances is a Temporal activity that ingests GCP Compute instances.
// Client is created/reused per session - lives for workflow duration.
func (a *Activities) IngestComputeInstances(ctx context.Context, params IngestComputeInstancesParams) (*IngestComputeInstancesResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GCP Compute instance ingestion",
		"sessionID", params.SessionID,
		"projectID", params.ProjectID,
	)

	// Get or create client for this session
	client, err := GetOrCreateSessionClient(ctx, params.SessionID, a.configService, a.limiter)
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	// Create service with session client
	service := NewService(client, a.db)
	result, err := service.Ingest(ctx, IngestParams{
		ProjectID: params.ProjectID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to ingest instances: %w", err)
	}

	// Delete stale instances
	if err := service.DeleteStaleInstances(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale instances", "error", err)
	}

	logger.Info("Completed GCP Compute instance ingestion",
		"projectID", params.ProjectID,
		"instanceCount", result.InstanceCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestComputeInstancesResult{
		ProjectID:      result.ProjectID,
		InstanceCount:  result.InstanceCount,
		DurationMillis: result.DurationMillis,
	}, nil
}

// CloseSessionClientParams contains parameters for cleanup activity.
type CloseSessionClientParams struct {
	SessionID string
}

// CloseSessionClientActivity is the activity function reference for workflow registration.
var CloseSessionClientActivity = (*Activities).CloseSessionClient

// CloseSessionClient closes the client for a session.
func (a *Activities) CloseSessionClient(ctx context.Context, params CloseSessionClientParams) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Closing session client", "sessionID", params.SessionID)

	CloseSessionClient(params.SessionID)
	return nil
}
