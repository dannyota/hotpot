package attestor

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/activity"
	"google.golang.org/api/option"
	"google.golang.org/grpc"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	"danny.vn/hotpot/pkg/base/temporalerr"
	entbinaryauthorization "danny.vn/hotpot/pkg/storage/ent/gcp/binaryauthorization"
)

// Activities holds dependencies for Temporal activities.
type Activities struct {
	configService *config.Service
	entClient     *entbinaryauthorization.Client
	limiter       ratelimit.Limiter
}

// NewActivities creates a new Activities instance.
func NewActivities(configService *config.Service, entClient *entbinaryauthorization.Client, limiter ratelimit.Limiter) *Activities {
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

// IngestBinaryAuthorizationAttestorsParams contains parameters for the ingest activity.
type IngestBinaryAuthorizationAttestorsParams struct {
	ProjectID string
}

// IngestBinaryAuthorizationAttestorsResult contains the result of the ingest activity.
type IngestBinaryAuthorizationAttestorsResult struct {
	ProjectID      string
	AttestorCount  int
	DurationMillis int64
}

// IngestBinaryAuthorizationAttestorsActivity is the activity function reference for workflow registration.
var IngestBinaryAuthorizationAttestorsActivity = (*Activities).IngestBinaryAuthorizationAttestors

// IngestBinaryAuthorizationAttestors is a Temporal activity that ingests Binary Authorization attestors.
func (a *Activities) IngestBinaryAuthorizationAttestors(ctx context.Context, params IngestBinaryAuthorizationAttestorsParams) (*IngestBinaryAuthorizationAttestorsResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting Binary Authorization attestor ingestion",
		"projectID", params.ProjectID,
	)

	client, err := a.createClient(ctx)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("create client: %w", err))
	}
	defer client.Close()

	service := NewService(client, a.entClient)
	result, err := service.Ingest(ctx, IngestParams{
		ProjectID: params.ProjectID,
	})
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("failed to ingest binary authorization attestors: %w", err))
	}

	if err := service.DeleteStaleAttestors(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale binary authorization attestors", "error", err)
	}

	logger.Info("Completed Binary Authorization attestor ingestion",
		"projectID", params.ProjectID,
		"attestorCount", result.AttestorCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestBinaryAuthorizationAttestorsResult{
		ProjectID:      result.ProjectID,
		AttestorCount:  result.AttestorCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
