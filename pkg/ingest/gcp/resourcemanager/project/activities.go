package project

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

// IngestProjectsParams contains parameters for the ingest activity.
type IngestProjectsParams struct {
	SessionID string
}

// IngestProjectsResult contains the result of the ingest activity.
type IngestProjectsResult struct {
	ProjectCount   int
	ProjectIDs     []string
	DurationMillis int64
}

// IngestProjectsActivity is the activity function reference for workflow registration.
var IngestProjectsActivity = (*Activities).IngestProjects

// IngestProjects is a Temporal activity that discovers and ingests all accessible GCP projects.
// Client is created/reused per session - lives for workflow duration.
func (a *Activities) IngestProjects(ctx context.Context, params IngestProjectsParams) (*IngestProjectsResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GCP project discovery", "sessionID", params.SessionID)

	// Get or create client for this session
	client, err := GetOrCreateSessionClient(ctx, params.SessionID, a.configService, a.limiter)
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	// Create service with session client
	service := NewService(client, a.db)
	result, err := service.Ingest(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to ingest projects: %w", err)
	}

	// Delete stale projects
	if err := service.DeleteStaleProjects(ctx, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale projects", "error", err)
	}

	logger.Info("Completed GCP project discovery",
		"projectCount", result.ProjectCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestProjectsResult{
		ProjectCount:   result.ProjectCount,
		ProjectIDs:     result.ProjectIDs,
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
