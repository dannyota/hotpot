package revision

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

// IngestRunRevisionsParams contains parameters for the ingest activity.
type IngestRunRevisionsParams struct {
	ProjectID string
}

// IngestRunRevisionsResult contains the result of the ingest activity.
type IngestRunRevisionsResult struct {
	ProjectID      string
	RevisionCount  int
	DurationMillis int64
}

// IngestRunRevisionsActivity is the activity function reference for workflow registration.
var IngestRunRevisionsActivity = (*Activities).IngestRunRevisions

// IngestRunRevisions is a Temporal activity that ingests Cloud Run revisions.
func (a *Activities) IngestRunRevisions(ctx context.Context, params IngestRunRevisionsParams) (*IngestRunRevisionsResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting Cloud Run revision ingestion",
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
		return nil, fmt.Errorf("failed to ingest Cloud Run revisions: %w", err)
	}

	if err := svc.DeleteStaleRevisions(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale Cloud Run revisions", "error", err)
	}

	logger.Info("Completed Cloud Run revision ingestion",
		"projectID", params.ProjectID,
		"revisionCount", result.RevisionCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestRunRevisionsResult{
		ProjectID:      result.ProjectID,
		RevisionCount:  result.RevisionCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
