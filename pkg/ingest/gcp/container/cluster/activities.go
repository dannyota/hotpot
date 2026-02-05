package cluster

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

// IngestContainerClustersParams contains parameters for the ingest activity.
type IngestContainerClustersParams struct {
	SessionID string
	ProjectID string
}

// IngestContainerClustersResult contains the result of the ingest activity.
type IngestContainerClustersResult struct {
	ProjectID      string
	ClusterCount   int
	DurationMillis int64
}

// IngestContainerClustersActivity is the activity function reference for workflow registration.
var IngestContainerClustersActivity = (*Activities).IngestContainerClusters

// IngestContainerClusters is a Temporal activity that ingests GKE clusters.
// Client is created/reused per session - lives for workflow duration.
func (a *Activities) IngestContainerClusters(ctx context.Context, params IngestContainerClustersParams) (*IngestContainerClustersResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GKE cluster ingestion",
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
		return nil, fmt.Errorf("failed to ingest clusters: %w", err)
	}

	// Delete stale clusters
	if err := service.DeleteStaleClusters(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale clusters", "error", err)
	}

	logger.Info("Completed GKE cluster ingestion",
		"projectID", params.ProjectID,
		"clusterCount", result.ClusterCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestContainerClustersResult{
		ProjectID:      result.ProjectID,
		ClusterCount:   result.ClusterCount,
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
