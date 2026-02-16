package appservice

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

// IngestAppEngineServicesParams contains parameters for the ingest activity.
type IngestAppEngineServicesParams struct {
	ProjectID string
}

// IngestAppEngineServicesResult contains the result of the ingest activity.
type IngestAppEngineServicesResult struct {
	ProjectID      string
	ServiceCount   int
	DurationMillis int64
}

// IngestAppEngineServicesActivity is the activity function reference for workflow registration.
var IngestAppEngineServicesActivity = (*Activities).IngestAppEngineServices

// IngestAppEngineServices is a Temporal activity that ingests App Engine services.
func (a *Activities) IngestAppEngineServices(ctx context.Context, params IngestAppEngineServicesParams) (*IngestAppEngineServicesResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting App Engine service ingestion",
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
		return nil, fmt.Errorf("failed to ingest App Engine services: %w", err)
	}

	if err := service.DeleteStaleServices(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale App Engine services", "error", err)
	}

	logger.Info("Completed App Engine service ingestion",
		"projectID", params.ProjectID,
		"serviceCount", result.ServiceCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestAppEngineServicesResult{
		ProjectID:      result.ProjectID,
		ServiceCount:   result.ServiceCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
