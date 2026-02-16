package subscription

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

// IngestPubSubSubscriptionsParams contains parameters for the ingest activity.
type IngestPubSubSubscriptionsParams struct {
	ProjectID string
}

// IngestPubSubSubscriptionsResult contains the result of the ingest activity.
type IngestPubSubSubscriptionsResult struct {
	ProjectID         string
	SubscriptionCount int
	DurationMillis    int64
}

// IngestPubSubSubscriptionsActivity is the activity function reference for workflow registration.
var IngestPubSubSubscriptionsActivity = (*Activities).IngestPubSubSubscriptions

// IngestPubSubSubscriptions is a Temporal activity that ingests Pub/Sub subscriptions.
func (a *Activities) IngestPubSubSubscriptions(ctx context.Context, params IngestPubSubSubscriptionsParams) (*IngestPubSubSubscriptionsResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting Pub/Sub subscription ingestion",
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
		return nil, fmt.Errorf("failed to ingest Pub/Sub subscriptions: %w", err)
	}

	if err := service.DeleteStaleSubscriptions(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale Pub/Sub subscriptions", "error", err)
	}

	logger.Info("Completed Pub/Sub subscription ingestion",
		"projectID", params.ProjectID,
		"subscriptionCount", result.SubscriptionCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestPubSubSubscriptionsResult{
		ProjectID:         result.ProjectID,
		SubscriptionCount: result.SubscriptionCount,
		DurationMillis:    result.DurationMillis,
	}, nil
}
