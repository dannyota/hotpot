package occurrence

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

// IngestOccurrencesParams contains parameters for the ingest activity.
type IngestOccurrencesParams struct {
	ProjectID string
}

// IngestOccurrencesResult contains the result of the ingest activity.
type IngestOccurrencesResult struct {
	ProjectID       string
	OccurrenceCount int
	DurationMillis  int64
}

// IngestOccurrencesActivity is the activity function reference for workflow registration.
var IngestOccurrencesActivity = (*Activities).IngestOccurrences

// IngestOccurrences is a Temporal activity that ingests Grafeas occurrences.
func (a *Activities) IngestOccurrences(ctx context.Context, params IngestOccurrencesParams) (*IngestOccurrencesResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting Container Analysis occurrence ingestion",
		"projectID", params.ProjectID,
	)

	client, err := a.createClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}
	defer client.Close()

	service := NewService(client, a.entClient)
	result, err := service.Ingest(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to ingest occurrences: %w", err)
	}

	if err := service.DeleteStaleOccurrences(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale occurrences", "error", err)
	}

	logger.Info("Completed Container Analysis occurrence ingestion",
		"projectID", params.ProjectID,
		"occurrenceCount", result.OccurrenceCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestOccurrencesResult{
		ProjectID:       result.ProjectID,
		OccurrenceCount: result.OccurrenceCount,
		DurationMillis:  result.DurationMillis,
	}, nil
}
