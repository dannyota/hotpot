package targetinstance

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

// IngestComputeTargetInstancesParams contains parameters for the ingest activity.
type IngestComputeTargetInstancesParams struct {
	SessionID string
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
// Client is created/reused per session - lives for workflow duration.
func (a *Activities) IngestComputeTargetInstances(ctx context.Context, params IngestComputeTargetInstancesParams) (*IngestComputeTargetInstancesResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GCP Compute target instance ingestion",
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
