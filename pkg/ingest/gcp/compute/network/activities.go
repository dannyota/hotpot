package network

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

// IngestComputeNetworksParams contains parameters for the ingest activity.
type IngestComputeNetworksParams struct {
	SessionID string
	ProjectID string
}

// IngestComputeNetworksResult contains the result of the ingest activity.
type IngestComputeNetworksResult struct {
	ProjectID      string
	NetworkCount   int
	DurationMillis int64
}

// IngestComputeNetworksActivity is the activity function reference for workflow registration.
var IngestComputeNetworksActivity = (*Activities).IngestComputeNetworks

// IngestComputeNetworks is a Temporal activity that ingests GCP Compute networks.
// Client is created/reused per session - lives for workflow duration.
func (a *Activities) IngestComputeNetworks(ctx context.Context, params IngestComputeNetworksParams) (*IngestComputeNetworksResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GCP Compute network ingestion",
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
		return nil, fmt.Errorf("failed to ingest networks: %w", err)
	}

	// Delete stale networks
	if err := service.DeleteStaleNetworks(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale networks", "error", err)
	}

	logger.Info("Completed GCP Compute network ingestion",
		"projectID", params.ProjectID,
		"networkCount", result.NetworkCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestComputeNetworksResult{
		ProjectID:      result.ProjectID,
		NetworkCount:   result.NetworkCount,
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
