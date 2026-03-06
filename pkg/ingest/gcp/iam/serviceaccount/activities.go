package serviceaccount

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/activity"
	"google.golang.org/api/option"
	"google.golang.org/grpc"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	"danny.vn/hotpot/pkg/base/temporalerr"
	entiam "danny.vn/hotpot/pkg/storage/ent/gcp/iam"
)

type Activities struct {
	configService *config.Service
	entClient     *entiam.Client
	limiter       ratelimit.Limiter
}

func NewActivities(configService *config.Service, entClient *entiam.Client, limiter ratelimit.Limiter) *Activities {
	return &Activities{
		configService: configService,
		entClient:     entClient,
		limiter:       limiter,
	}
}

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

type IngestIAMServiceAccountsParams struct {
	ProjectID string
}

type IngestIAMServiceAccountsResult struct {
	ProjectID           string
	ServiceAccountCount int
	DurationMillis      int64
}

var IngestIAMServiceAccountsActivity = (*Activities).IngestIAMServiceAccounts

func (a *Activities) IngestIAMServiceAccounts(ctx context.Context, params IngestIAMServiceAccountsParams) (*IngestIAMServiceAccountsResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting IAM service account ingestion", "projectID", params.ProjectID)

	client, err := a.createClient(ctx)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("create client: %w", err))
	}
	defer client.Close()

	service := NewService(client, a.entClient)
	result, err := service.Ingest(ctx, IngestParams{ProjectID: params.ProjectID})
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("failed to ingest service accounts: %w", err))
	}

	if err := service.DeleteStaleServiceAccounts(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale service accounts", "error", err)
	}

	logger.Info("Completed IAM service account ingestion",
		"projectID", params.ProjectID,
		"serviceAccountCount", result.ServiceAccountCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestIAMServiceAccountsResult{
		ProjectID:           result.ProjectID,
		ServiceAccountCount: result.ServiceAccountCount,
		DurationMillis:      result.DurationMillis,
	}, nil
}
