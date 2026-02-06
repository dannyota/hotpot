package globalforwardingrule

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

// IngestComputeGlobalForwardingRulesParams contains parameters for the ingest activity.
type IngestComputeGlobalForwardingRulesParams struct {
	SessionID string
	ProjectID string
}

// IngestComputeGlobalForwardingRulesResult contains the result of the ingest activity.
type IngestComputeGlobalForwardingRulesResult struct {
	ProjectID                string
	GlobalForwardingRuleCount int
	DurationMillis           int64
}

// IngestComputeGlobalForwardingRulesActivity is the activity function reference for workflow registration.
var IngestComputeGlobalForwardingRulesActivity = (*Activities).IngestComputeGlobalForwardingRules

// IngestComputeGlobalForwardingRules is a Temporal activity that ingests GCP Compute global forwarding rules.
// Client is created/reused per session - lives for workflow duration.
func (a *Activities) IngestComputeGlobalForwardingRules(ctx context.Context, params IngestComputeGlobalForwardingRulesParams) (*IngestComputeGlobalForwardingRulesResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GCP Compute global forwarding rule ingestion",
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
		return nil, fmt.Errorf("failed to ingest global forwarding rules: %w", err)
	}

	// Delete stale global forwarding rules
	if err := service.DeleteStaleGlobalForwardingRules(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale global forwarding rules", "error", err)
	}

	logger.Info("Completed GCP Compute global forwarding rule ingestion",
		"projectID", params.ProjectID,
		"globalForwardingRuleCount", result.GlobalForwardingRuleCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestComputeGlobalForwardingRulesResult{
		ProjectID:                result.ProjectID,
		GlobalForwardingRuleCount: result.GlobalForwardingRuleCount,
		DurationMillis:           result.DurationMillis,
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
