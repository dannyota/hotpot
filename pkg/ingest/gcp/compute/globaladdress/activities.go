package globaladdress

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

// IngestComputeGlobalAddressesParams contains parameters for the ingest activity.
type IngestComputeGlobalAddressesParams struct {
	SessionID string
	ProjectID string
}

// IngestComputeGlobalAddressesResult contains the result of the ingest activity.
type IngestComputeGlobalAddressesResult struct {
	ProjectID          string
	GlobalAddressCount int
	DurationMillis     int64
}

// IngestComputeGlobalAddressesActivity is the activity function reference for workflow registration.
var IngestComputeGlobalAddressesActivity = (*Activities).IngestComputeGlobalAddresses

// IngestComputeGlobalAddresses is a Temporal activity that ingests GCP Compute global addresses.
// Client is created/reused per session - lives for workflow duration.
func (a *Activities) IngestComputeGlobalAddresses(ctx context.Context, params IngestComputeGlobalAddressesParams) (*IngestComputeGlobalAddressesResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GCP Compute global address ingestion",
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
		return nil, fmt.Errorf("failed to ingest global addresses: %w", err)
	}

	// Delete stale global addresses
	if err := service.DeleteStaleGlobalAddresses(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale global addresses", "error", err)
	}

	logger.Info("Completed GCP Compute global address ingestion",
		"projectID", params.ProjectID,
		"globalAddressCount", result.GlobalAddressCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestComputeGlobalAddressesResult{
		ProjectID:          result.ProjectID,
		GlobalAddressCount: result.GlobalAddressCount,
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
