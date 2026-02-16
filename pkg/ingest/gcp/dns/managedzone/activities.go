package managedzone

import (
	"context"
	"fmt"
	"net/http"

	"go.temporal.io/sdk/activity"
	"google.golang.org/api/option"

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
	httpClient := &http.Client{
		Transport: ratelimit.NewRateLimitedTransport(a.limiter, nil),
	}
	return NewClient(ctx, httpClient, opts...)
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
		return nil, fmt.Errorf("create client: %w", err)
	}
	defer client.Close()

	// Create service
	service := NewService(client, a.entClient)
	result, err := service.Ingest(ctx, IngestParams{
		ProjectID: params.ProjectID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to ingest managed zones: %w", err)
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
