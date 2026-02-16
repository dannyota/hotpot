package note

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

// IngestNotesParams contains parameters for the ingest activity.
type IngestNotesParams struct {
	ProjectID string
}

// IngestNotesResult contains the result of the ingest activity.
type IngestNotesResult struct {
	ProjectID      string
	NoteCount      int
	DurationMillis int64
}

// IngestNotesActivity is the activity function reference for workflow registration.
var IngestNotesActivity = (*Activities).IngestNotes

// IngestNotes is a Temporal activity that ingests Grafeas notes.
func (a *Activities) IngestNotes(ctx context.Context, params IngestNotesParams) (*IngestNotesResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting Container Analysis note ingestion",
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
		return nil, fmt.Errorf("failed to ingest notes: %w", err)
	}

	if err := service.DeleteStaleNotes(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale notes", "error", err)
	}

	logger.Info("Completed Container Analysis note ingestion",
		"projectID", params.ProjectID,
		"noteCount", result.NoteCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestNotesResult{
		ProjectID:      result.ProjectID,
		NoteCount:      result.NoteCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
