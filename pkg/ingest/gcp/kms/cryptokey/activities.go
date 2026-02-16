package cryptokey

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

// IngestKMSCryptoKeysParams contains parameters for the ingest activity.
type IngestKMSCryptoKeysParams struct {
	ProjectID string
}

// IngestKMSCryptoKeysResult contains the result of the ingest activity.
type IngestKMSCryptoKeysResult struct {
	ProjectID      string
	CryptoKeyCount int
	DurationMillis int64
}

// IngestKMSCryptoKeysActivity is the activity function reference for workflow registration.
var IngestKMSCryptoKeysActivity = (*Activities).IngestKMSCryptoKeys

// IngestKMSCryptoKeys is a Temporal activity that ingests GCP KMS crypto keys.
func (a *Activities) IngestKMSCryptoKeys(ctx context.Context, params IngestKMSCryptoKeysParams) (*IngestKMSCryptoKeysResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GCP KMS crypto key ingestion",
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
		return nil, fmt.Errorf("failed to ingest crypto keys: %w", err)
	}

	if err := service.DeleteStaleCryptoKeys(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale crypto keys", "error", err)
	}

	logger.Info("Completed GCP KMS crypto key ingestion",
		"projectID", params.ProjectID,
		"cryptoKeyCount", result.CryptoKeyCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestKMSCryptoKeysResult{
		ProjectID:      result.ProjectID,
		CryptoKeyCount: result.CryptoKeyCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
