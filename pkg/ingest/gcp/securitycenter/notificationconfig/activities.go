package notificationconfig

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
	return NewClient(ctx, a.entClient, opts...)
}

// IngestNotificationConfigsParams contains parameters for the ingest activity.
type IngestNotificationConfigsParams struct {
}

// IngestNotificationConfigsResult contains the result of the ingest activity.
type IngestNotificationConfigsResult struct {
	NotificationConfigCount int
	DurationMillis          int64
}

// IngestNotificationConfigsActivity is the activity function reference for workflow registration.
var IngestNotificationConfigsActivity = (*Activities).IngestNotificationConfigs

// IngestNotificationConfigs is a Temporal activity that ingests SCC notification configs.
func (a *Activities) IngestNotificationConfigs(ctx context.Context, params IngestNotificationConfigsParams) (*IngestNotificationConfigsResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting SCC notification config ingestion")

	client, err := a.createClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}
	defer client.Close()

	service := NewService(client, a.entClient)
	result, err := service.Ingest(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to ingest SCC notification configs: %w", err)
	}

	if err := service.DeleteStaleNotificationConfigs(ctx, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale SCC notification configs", "error", err)
	}

	logger.Info("Completed SCC notification config ingestion",
		"notificationConfigCount", result.NotificationConfigCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestNotificationConfigsResult{
		NotificationConfigCount: result.NotificationConfigCount,
		DurationMillis:          result.DurationMillis,
	}, nil
}
