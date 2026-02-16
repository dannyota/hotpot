package application

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/activity"
	"google.golang.org/api/option"
	"google.golang.org/grpc"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// Activities holds dependencies for Temporal activities.
type Activities struct {
	configService *config.Service
	entClient     *ent.Client
	limiter       ratelimit.Limiter
}

// NewActivities creates a new Activities instance.
func NewActivities(configService *config.Service, entClient *ent.Client, limiter ratelimit.Limiter) *Activities {
	return &Activities{
		configService: configService,
		entClient:     entClient,
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

// IngestAppEngineApplicationsParams contains parameters for the ingest activity.
type IngestAppEngineApplicationsParams struct {
	ProjectID string
}

// IngestAppEngineApplicationsResult contains the result of the ingest activity.
type IngestAppEngineApplicationsResult struct {
	ProjectID        string
	ApplicationCount int
	DurationMillis   int64
}

// IngestAppEngineApplicationsActivity is the activity function reference for workflow registration.
var IngestAppEngineApplicationsActivity = (*Activities).IngestAppEngineApplications

// IngestAppEngineApplications is a Temporal activity that ingests App Engine applications.
func (a *Activities) IngestAppEngineApplications(ctx context.Context, params IngestAppEngineApplicationsParams) (*IngestAppEngineApplicationsResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting App Engine application ingestion",
		"projectID", params.ProjectID,
	)

	client, err := a.createClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}
	defer client.Close()

	service := NewService(client, a.entClient)
	result, err := service.Ingest(ctx, IngestParams{
		ProjectID: params.ProjectID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to ingest App Engine applications: %w", err)
	}

	if result.ApplicationCount > 0 {
		if err := service.DeleteStaleApplications(ctx, params.ProjectID, result.CollectedAt); err != nil {
			logger.Warn("Failed to delete stale App Engine applications", "error", err)
		}
	}

	logger.Info("Completed App Engine application ingestion",
		"projectID", params.ProjectID,
		"applicationCount", result.ApplicationCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestAppEngineApplicationsResult{
		ProjectID:        result.ProjectID,
		ApplicationCount: result.ApplicationCount,
		DurationMillis:   result.DurationMillis,
	}, nil
}
