package managedzone

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/activity"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/gcpauth"
	"danny.vn/hotpot/pkg/base/ratelimit"
	"danny.vn/hotpot/pkg/base/temporalerr"
	entdns "danny.vn/hotpot/pkg/storage/ent/gcp/dns"
)

// Activities holds dependencies for Temporal activities.
type Activities struct {
	configService *config.Service
	entClient     *entdns.Client
	limiter       ratelimit.Limiter
}

// NewActivities creates a new Activities instance.
func NewActivities(configService *config.Service, entClient *entdns.Client, limiter ratelimit.Limiter) *Activities {
	return &Activities{
		configService: configService,
		entClient:     entClient,
		limiter:       limiter,
	}
}

// createClient creates a rate-limited GCP client with credentials.
func (a *Activities) createClient(ctx context.Context) (*Client, error) {
	httpClient, err := gcpauth.NewHTTPClient(ctx, a.configService.GCPCredentialsJSON(), a.limiter)
	if err != nil {
		return nil, err
	}
	return NewClient(ctx, httpClient)
}

// IngestDNSManagedZonesParams contains parameters for the ingest activity.
type IngestDNSManagedZonesParams struct {
	ProjectID string
}

// IngestDNSManagedZonesResult contains the result of the ingest activity.
type IngestDNSManagedZonesResult struct {
	ProjectID        string
	ManagedZoneCount int
	DurationMillis   int64
}

// IngestDNSManagedZonesActivity is the activity function reference for workflow registration.
var IngestDNSManagedZonesActivity = (*Activities).IngestDNSManagedZones

// IngestDNSManagedZones is a Temporal activity that ingests GCP DNS managed zones.
func (a *Activities) IngestDNSManagedZones(ctx context.Context, params IngestDNSManagedZonesParams) (*IngestDNSManagedZonesResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GCP DNS managed zone ingestion",
		"projectID", params.ProjectID,
	)

	// Create client for this activity
	client, err := a.createClient(ctx)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("create client: %w", err))
	}
	defer client.Close()

	// Create service
	service := NewService(client, a.entClient)
	result, err := service.Ingest(ctx, IngestParams{
		ProjectID: params.ProjectID,
	})
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("failed to ingest managed zones: %w", err))
	}

	// Delete stale managed zones
	if err := service.DeleteStaleManagedZones(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale managed zones", "error", err)
	}

	logger.Info("Completed GCP DNS managed zone ingestion",
		"projectID", params.ProjectID,
		"managedZoneCount", result.ManagedZoneCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestDNSManagedZonesResult{
		ProjectID:        result.ProjectID,
		ManagedZoneCount: result.ManagedZoneCount,
		DurationMillis:   result.DurationMillis,
	}, nil
}
