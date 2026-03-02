package firewall

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/activity"
	"google.golang.org/api/option"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/gcpauth"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/base/temporalerr"
	entcompute "github.com/dannyota/hotpot/pkg/storage/ent/gcp/compute"
)

// Activities holds dependencies for Temporal activities.
type Activities struct {
	configService *config.Service
	entClient     *entcompute.Client
	limiter       ratelimit.Limiter
}

// NewActivities creates a new Activities instance.
func NewActivities(configService *config.Service, entClient *entcompute.Client, limiter ratelimit.Limiter) *Activities {
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
	return NewClient(ctx, option.WithHTTPClient(httpClient))
}

// IngestComputeFirewallsParams contains parameters for the ingest activity.
type IngestComputeFirewallsParams struct {
	ProjectID string
}

// IngestComputeFirewallsResult contains the result of the ingest activity.
type IngestComputeFirewallsResult struct {
	ProjectID      string
	FirewallCount  int
	DurationMillis int64
}

// IngestComputeFirewallsActivity is the activity function reference for workflow registration.
var IngestComputeFirewallsActivity = (*Activities).IngestComputeFirewalls

// IngestComputeFirewalls is a Temporal activity that ingests GCP Compute firewalls.
func (a *Activities) IngestComputeFirewalls(ctx context.Context, params IngestComputeFirewallsParams) (*IngestComputeFirewallsResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GCP Compute firewall ingestion",
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
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("failed to ingest firewalls: %w", err))
	}

	// Delete stale firewalls
	if err := service.DeleteStaleFirewalls(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale firewalls", "error", err)
	}

	logger.Info("Completed GCP Compute firewall ingestion",
		"projectID", params.ProjectID,
		"firewallCount", result.FirewallCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestComputeFirewallsResult{
		ProjectID:      result.ProjectID,
		FirewallCount:  result.FirewallCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
