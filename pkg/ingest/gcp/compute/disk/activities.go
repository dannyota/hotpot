package disk

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/activity"
	"gorm.io/gorm"

	"hotpot/pkg/base/config"
)

// Activities holds dependencies for Temporal activities.
type Activities struct {
	configService *config.Service
	db            *gorm.DB
}

// NewActivities creates a new Activities instance.
func NewActivities(configService *config.Service, db *gorm.DB) *Activities {
	return &Activities{
		configService: configService,
		db:            db,
	}
}

// IngestComputeDisksParams contains parameters for the ingest activity.
type IngestComputeDisksParams struct {
	SessionID string
	ProjectID string
}

// IngestComputeDisksResult contains the result of the ingest activity.
type IngestComputeDisksResult struct {
	ProjectID      string
	DiskCount      int
	DurationMillis int64
}

// IngestComputeDisksActivity is the activity function reference for workflow registration.
var IngestComputeDisksActivity = (*Activities).IngestComputeDisks

// IngestComputeDisks is a Temporal activity that ingests GCP Compute disks.
// Client is created/reused per session - lives for workflow duration.
func (a *Activities) IngestComputeDisks(ctx context.Context, params IngestComputeDisksParams) (*IngestComputeDisksResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GCP Compute disk ingestion",
		"sessionID", params.SessionID,
		"projectID", params.ProjectID,
	)

	// Get or create client for this session
	client, err := GetOrCreateSessionClient(ctx, params.SessionID, a.configService)
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	// Create service with session client
	service := NewService(client, a.db)
	result, err := service.Ingest(ctx, IngestParams{
		ProjectID: params.ProjectID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to ingest disks: %w", err)
	}

	// Delete stale disks
	if err := service.DeleteStaleDisks(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale disks", "error", err)
	}

	logger.Info("Completed GCP Compute disk ingestion",
		"projectID", params.ProjectID,
		"diskCount", result.DiskCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestComputeDisksResult{
		ProjectID:      result.ProjectID,
		DiskCount:      result.DiskCount,
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
