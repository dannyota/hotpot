package topic

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

// IngestPubSubTopicsParams contains parameters for the ingest activity.
type IngestPubSubTopicsParams struct {
	ProjectID string
}

// IngestPubSubTopicsResult contains the result of the ingest activity.
type IngestPubSubTopicsResult struct {
	ProjectID      string
	TopicCount     int
	DurationMillis int64
}

// IngestPubSubTopicsActivity is the activity function reference for workflow registration.
var IngestPubSubTopicsActivity = (*Activities).IngestPubSubTopics

// IngestPubSubTopics is a Temporal activity that ingests Pub/Sub topics.
func (a *Activities) IngestPubSubTopics(ctx context.Context, params IngestPubSubTopicsParams) (*IngestPubSubTopicsResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting Pub/Sub topic ingestion",
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
		return nil, fmt.Errorf("failed to ingest Pub/Sub topics: %w", err)
	}

	if err := service.DeleteStaleTopics(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale Pub/Sub topics", "error", err)
	}

	logger.Info("Completed Pub/Sub topic ingestion",
		"projectID", params.ProjectID,
		"topicCount", result.TopicCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestPubSubTopicsResult{
		ProjectID:      result.ProjectID,
		TopicCount:     result.TopicCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
