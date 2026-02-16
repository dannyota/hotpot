package service

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

// IngestRunServicesParams contains parameters for the ingest activity.
type IngestRunServicesParams struct {
	ProjectID string
}

// IngestRunServicesResult contains the result of the ingest activity.
type IngestRunServicesResult struct {
	ProjectID      string
	ServiceCount   int
	DurationMillis int64
}

// IngestRunServicesActivity is the activity function reference for workflow registration.
var IngestRunServicesActivity = (*Activities).IngestRunServices

// IngestRunServices is a Temporal activity that ingests Cloud Run services.
func (a *Activities) IngestRunServices(ctx context.Context, params IngestRunServicesParams) (*IngestRunServicesResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting Cloud Run service ingestion",
		"projectID", params.ProjectID,
	)

	client, err := a.createClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}
	defer client.Close()

	svc := NewService(client, a.entClient)
	result, err := svc.Ingest(ctx, IngestParams{
		ProjectID: params.ProjectID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to ingest Cloud Run services: %w", err)
	}

	if err := svc.DeleteStaleServices(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale Cloud Run services", "error", err)
	}

	logger.Info("Completed Cloud Run service ingestion",
		"projectID", params.ProjectID,
		"serviceCount", result.ServiceCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestRunServicesResult{
		ProjectID:      result.ProjectID,
		ServiceCount:   result.ServiceCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
