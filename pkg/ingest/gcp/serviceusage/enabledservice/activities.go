package enabledservice

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

// IngestEnabledServicesParams contains parameters for the ingest activity.
type IngestEnabledServicesParams struct {
	ProjectID string
}

// IngestEnabledServicesResult contains the result of the ingest activity.
type IngestEnabledServicesResult struct {
	ProjectID      string
	ServiceCount   int
	DurationMillis int64
}

// IngestEnabledServicesActivity is the activity function reference for workflow registration.
var IngestEnabledServicesActivity = (*Activities).IngestEnabledServices

// IngestEnabledServices is a Temporal activity that ingests GCP enabled services.
func (a *Activities) IngestEnabledServices(ctx context.Context, params IngestEnabledServicesParams) (*IngestEnabledServicesResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GCP Service Usage enabled service ingestion",
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
		return nil, fmt.Errorf("failed to ingest enabled services: %w", err)
	}

	if err := service.DeleteStaleServices(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale enabled services", "error", err)
	}

	logger.Info("Completed GCP Service Usage enabled service ingestion",
		"projectID", params.ProjectID,
		"serviceCount", result.ServiceCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestEnabledServicesResult{
		ProjectID:      result.ProjectID,
		ServiceCount:   result.ServiceCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
