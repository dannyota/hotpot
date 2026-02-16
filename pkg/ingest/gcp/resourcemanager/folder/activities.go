package folder

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

// IngestFoldersParams contains parameters for the ingest activity.
type IngestFoldersParams struct {
}

// IngestFoldersResult contains the result of the ingest activity.
type IngestFoldersResult struct {
	FolderCount    int
	FolderIDs      []string
	DurationMillis int64
}

// IngestFoldersActivity is the activity function reference for workflow registration.
var IngestFoldersActivity = (*Activities).IngestFolders

// IngestFolders is a Temporal activity that discovers and ingests all accessible GCP folders.
func (a *Activities) IngestFolders(ctx context.Context, params IngestFoldersParams) (*IngestFoldersResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GCP folder discovery")

	// Create client for this activity
	client, err := a.createClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}
	defer client.Close()

	// Create service
	service := NewService(client, a.entClient)
	result, err := service.Ingest(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to ingest folders: %w", err)
	}

	// Delete stale folders
	if err := service.DeleteStaleFolders(ctx, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale folders", "error", err)
	}

	logger.Info("Completed GCP folder discovery",
		"folderCount", result.FolderCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestFoldersResult{
		FolderCount:    result.FolderCount,
		FolderIDs:      result.FolderIDs,
		DurationMillis: result.DurationMillis,
	}, nil
}
