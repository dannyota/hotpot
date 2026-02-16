package uptimecheck

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

// IngestUptimeChecksParams contains parameters for the ingest activity.
type IngestUptimeChecksParams struct {
	ProjectID string
}

// IngestUptimeChecksResult contains the result of the ingest activity.
type IngestUptimeChecksResult struct {
	ProjectID        string
	UptimeCheckCount int
	DurationMillis   int64
}

// IngestUptimeChecksActivity is the activity function reference for workflow registration.
var IngestUptimeChecksActivity = (*Activities).IngestUptimeChecks

// IngestUptimeChecks is a Temporal activity that ingests Monitoring uptime check configs.
func (a *Activities) IngestUptimeChecks(ctx context.Context, params IngestUptimeChecksParams) (*IngestUptimeChecksResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting Monitoring uptime check config ingestion",
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
		return nil, fmt.Errorf("failed to ingest uptime check configs: %w", err)
	}

	if err := service.DeleteStaleUptimeChecks(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale uptime check configs", "error", err)
	}

	logger.Info("Completed Monitoring uptime check config ingestion",
		"projectID", params.ProjectID,
		"uptimeCheckCount", result.UptimeCheckCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestUptimeChecksResult{
		ProjectID:        result.ProjectID,
		UptimeCheckCount: result.UptimeCheckCount,
		DurationMillis:   result.DurationMillis,
	}, nil
}
