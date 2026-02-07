package serviceaccountkey

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/activity"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"gorm.io/gorm"

	"hotpot/pkg/base/config"
	"hotpot/pkg/base/ratelimit"
)

type Activities struct {
	configService *config.Service
	db            *gorm.DB
	limiter       ratelimit.Limiter
}

func NewActivities(configService *config.Service, db *gorm.DB, limiter ratelimit.Limiter) *Activities {
	return &Activities{
		configService: configService,
		db:            db,
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

type IngestIAMServiceAccountKeysParams struct {
	ProjectID string
}

type IngestIAMServiceAccountKeysResult struct {
	ProjectID              string
	ServiceAccountKeyCount int
	DurationMillis         int64
}

var IngestIAMServiceAccountKeysActivity = (*Activities).IngestIAMServiceAccountKeys

func (a *Activities) IngestIAMServiceAccountKeys(ctx context.Context, params IngestIAMServiceAccountKeysParams) (*IngestIAMServiceAccountKeysResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting IAM service account key ingestion", "projectID", params.ProjectID)

	client, err := a.createClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}
	defer client.Close()

	service := NewService(client, a.db)
	result, err := service.Ingest(ctx, IngestParams{ProjectID: params.ProjectID})
	if err != nil {
		return nil, fmt.Errorf("failed to ingest service account keys: %w", err)
	}

	if err := service.DeleteStaleKeys(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale service account keys", "error", err)
	}

	logger.Info("Completed IAM service account key ingestion",
		"projectID", params.ProjectID,
		"serviceAccountKeyCount", result.ServiceAccountKeyCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestIAMServiceAccountKeysResult{
		ProjectID:              result.ProjectID,
		ServiceAccountKeyCount: result.ServiceAccountKeyCount,
		DurationMillis:         result.DurationMillis,
	}, nil
}
