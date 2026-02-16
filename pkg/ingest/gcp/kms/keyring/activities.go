package keyring

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

// KMSLocations are common GCP locations where KMS key rings can exist.
var KMSLocations = []string{
	"global",
	"us", "us-central1", "us-east1", "us-east4", "us-west1", "us-west2", "us-west3", "us-west4",
	"europe", "europe-west1", "europe-west2", "europe-west3", "europe-west4", "europe-west6",
	"asia", "asia-east1", "asia-east2", "asia-northeast1", "asia-northeast2", "asia-northeast3",
	"asia-south1", "asia-south2", "asia-southeast1", "asia-southeast2",
	"australia-southeast1", "australia-southeast2",
	"northamerica-northeast1", "northamerica-northeast2",
	"southamerica-east1",
}

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

// IngestKMSKeyRingsParams contains parameters for the ingest activity.
type IngestKMSKeyRingsParams struct {
	ProjectID string
}

// IngestKMSKeyRingsResult contains the result of the ingest activity.
type IngestKMSKeyRingsResult struct {
	ProjectID      string
	KeyRingCount   int
	DurationMillis int64
}

// IngestKMSKeyRingsActivity is the activity function reference for workflow registration.
var IngestKMSKeyRingsActivity = (*Activities).IngestKMSKeyRings

// IngestKMSKeyRings is a Temporal activity that ingests GCP KMS key rings.
func (a *Activities) IngestKMSKeyRings(ctx context.Context, params IngestKMSKeyRingsParams) (*IngestKMSKeyRingsResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GCP KMS key ring ingestion",
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
		Locations: KMSLocations,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to ingest key rings: %w", err)
	}

	if err := service.DeleteStaleKeyRings(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale key rings", "error", err)
	}

	logger.Info("Completed GCP KMS key ring ingestion",
		"projectID", params.ProjectID,
		"keyRingCount", result.KeyRingCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestKMSKeyRingsResult{
		ProjectID:      result.ProjectID,
		KeyRingCount:   result.KeyRingCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
