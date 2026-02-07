package instancegroup

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

// IngestComputeInstanceGroupsParams contains parameters for the ingest activity.
type IngestComputeInstanceGroupsParams struct {
	SessionID string
	ProjectID string
}

// IngestComputeInstanceGroupsResult contains the result of the ingest activity.
type IngestComputeInstanceGroupsResult struct {
	ProjectID          string
	InstanceGroupCount int
	DurationMillis     int64
}

// IngestComputeInstanceGroupsActivity is the activity function reference for workflow registration.
var IngestComputeInstanceGroupsActivity = (*Activities).IngestComputeInstanceGroups

// IngestComputeInstanceGroups is a Temporal activity that ingests GCP Compute instance groups.
// Client is created/reused per session - lives for workflow duration.
func (a *Activities) IngestComputeInstanceGroups(ctx context.Context, params IngestComputeInstanceGroupsParams) (*IngestComputeInstanceGroupsResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GCP Compute instance group ingestion",
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
		return nil, fmt.Errorf("failed to ingest instance groups: %w", err)
	}

	// Delete stale instance groups
	if err := service.DeleteStaleInstanceGroups(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale instance groups", "error", err)
	}

	logger.Info("Completed GCP Compute instance group ingestion",
		"projectID", params.ProjectID,
		"instanceGroupCount", result.InstanceGroupCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestComputeInstanceGroupsResult{
		ProjectID:          result.ProjectID,
		InstanceGroupCount: result.InstanceGroupCount,
		DurationMillis:     result.DurationMillis,
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
