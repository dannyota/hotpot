package function

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

// IngestCloudFunctionsFunctionsParams contains parameters for the ingest activity.
type IngestCloudFunctionsFunctionsParams struct {
	ProjectID string
}

// IngestCloudFunctionsFunctionsResult contains the result of the ingest activity.
type IngestCloudFunctionsFunctionsResult struct {
	ProjectID      string
	FunctionCount  int
	DurationMillis int64
}

// IngestCloudFunctionsFunctionsActivity is the activity function reference for workflow registration.
var IngestCloudFunctionsFunctionsActivity = (*Activities).IngestCloudFunctionsFunctions

// IngestCloudFunctionsFunctions is a Temporal activity that ingests GCP Cloud Functions.
func (a *Activities) IngestCloudFunctionsFunctions(ctx context.Context, params IngestCloudFunctionsFunctionsParams) (*IngestCloudFunctionsFunctionsResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GCP Cloud Functions function ingestion",
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
		return nil, fmt.Errorf("failed to ingest Cloud Functions: %w", err)
	}

	if err := service.DeleteStaleFunctions(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale Cloud Functions", "error", err)
	}

	logger.Info("Completed GCP Cloud Functions function ingestion",
		"projectID", params.ProjectID,
		"functionCount", result.FunctionCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestCloudFunctionsFunctionsResult{
		ProjectID:      result.ProjectID,
		FunctionCount:  result.FunctionCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
