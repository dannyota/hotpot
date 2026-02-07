package project

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/activity"
	"golang.org/x/time/rate"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
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
	opts = append(opts, option.WithGRPCDialOption(
		grpc.WithUnaryInterceptor(ratelimit.UnaryInterceptor(a.limiter)),
	))
	return NewClient(ctx, opts...)
}

// IngestProjectsParams contains parameters for the ingest activity.
type IngestProjectsParams struct {
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
func (a *Activities) IngestProjects(ctx context.Context, params IngestProjectsParams) (*IngestProjectsResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GCP project discovery")

	// Create client for this activity
	client, err := a.createClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}
	defer client.Close()

	// Create service
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

