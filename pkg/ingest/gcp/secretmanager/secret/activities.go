package secret

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

// IngestSecretManagerSecretsParams contains parameters for the ingest activity.
type IngestSecretManagerSecretsParams struct {
	ProjectID string
}

// IngestSecretManagerSecretsResult contains the result of the ingest activity.
type IngestSecretManagerSecretsResult struct {
	ProjectID      string
	SecretCount    int
	DurationMillis int64
}

// IngestSecretManagerSecretsActivity is the activity function reference for workflow registration.
var IngestSecretManagerSecretsActivity = (*Activities).IngestSecretManagerSecrets

// IngestSecretManagerSecrets is a Temporal activity that ingests GCP Secret Manager secrets.
func (a *Activities) IngestSecretManagerSecrets(ctx context.Context, params IngestSecretManagerSecretsParams) (*IngestSecretManagerSecretsResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GCP Secret Manager secret ingestion",
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
		return nil, fmt.Errorf("failed to ingest secrets: %w", err)
	}

	if err := service.DeleteStaleSecrets(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale secrets", "error", err)
	}

	logger.Info("Completed GCP Secret Manager secret ingestion",
		"projectID", params.ProjectID,
		"secretCount", result.SecretCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestSecretManagerSecretsResult{
		ProjectID:      result.ProjectID,
		SecretCount:    result.SecretCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
