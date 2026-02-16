package sink

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

// IngestLoggingSinksParams contains parameters for the ingest activity.
type IngestLoggingSinksParams struct {
	ProjectID string
}

// IngestLoggingSinksResult contains the result of the ingest activity.
type IngestLoggingSinksResult struct {
	ProjectID      string
	SinkCount      int
	DurationMillis int64
}

// IngestLoggingSinksActivity is the activity function reference for workflow registration.
var IngestLoggingSinksActivity = (*Activities).IngestLoggingSinks

// IngestLoggingSinks is a Temporal activity that ingests GCP Cloud Logging sinks.
func (a *Activities) IngestLoggingSinks(ctx context.Context, params IngestLoggingSinksParams) (*IngestLoggingSinksResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting Logging sink ingestion", "projectID", params.ProjectID)

	client, err := a.createClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}
	defer client.Close()

	service := NewService(client, a.entClient)
	result, err := service.Ingest(ctx, IngestParams{ProjectID: params.ProjectID})
	if err != nil {
		return nil, fmt.Errorf("failed to ingest sinks: %w", err)
	}

	if err := service.DeleteStaleSinks(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale sinks", "error", err)
	}

	logger.Info("Completed Logging sink ingestion",
		"projectID", params.ProjectID,
		"sinkCount", result.SinkCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestLoggingSinksResult{
		ProjectID:      result.ProjectID,
		SinkCount:      result.SinkCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
