package settings

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

// IngestIAPSettingsParams contains parameters for the ingest activity.
type IngestIAPSettingsParams struct {
	ProjectID string
}

// IngestIAPSettingsResult contains the result of the ingest activity.
type IngestIAPSettingsResult struct {
	ProjectID      string
	SettingsCount  int
	DurationMillis int64
}

// IngestIAPSettingsActivity is the activity function reference for workflow registration.
var IngestIAPSettingsActivity = (*Activities).IngestIAPSettings

// IngestIAPSettings is a Temporal activity that ingests IAP settings.
func (a *Activities) IngestIAPSettings(ctx context.Context, params IngestIAPSettingsParams) (*IngestIAPSettingsResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting IAP settings ingestion",
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
		return nil, fmt.Errorf("failed to ingest IAP settings: %w", err)
	}

	if err := service.DeleteStaleSettings(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale IAP settings", "error", err)
	}

	logger.Info("Completed IAP settings ingestion",
		"projectID", params.ProjectID,
		"settingsCount", result.SettingsCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestIAPSettingsResult{
		ProjectID:      result.ProjectID,
		SettingsCount:  result.SettingsCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
