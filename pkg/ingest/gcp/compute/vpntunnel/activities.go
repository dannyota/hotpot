package vpntunnel

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/activity"
	"google.golang.org/api/option"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/gcpauth"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/base/temporalerr"
	entvpn "github.com/dannyota/hotpot/pkg/storage/ent/gcp/vpn"
)

// Activities holds dependencies for Temporal activities.
type Activities struct {
	configService *config.Service
	entClient     *entvpn.Client
	limiter       ratelimit.Limiter
}

// NewActivities creates a new Activities instance.
func NewActivities(configService *config.Service, entClient *entvpn.Client, limiter ratelimit.Limiter) *Activities {
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

// IngestComputeVpnTunnelsParams contains parameters for the ingest activity.
type IngestComputeVpnTunnelsParams struct {
	ProjectID string
}

// IngestComputeVpnTunnelsResult contains the result of the ingest activity.
type IngestComputeVpnTunnelsResult struct {
	ProjectID      string
	VpnTunnelCount int
	DurationMillis int64
}

// IngestComputeVpnTunnelsActivity is the activity function reference for workflow registration.
var IngestComputeVpnTunnelsActivity = (*Activities).IngestComputeVpnTunnels

// IngestComputeVpnTunnels is a Temporal activity that ingests GCP Compute VPN tunnels.
func (a *Activities) IngestComputeVpnTunnels(ctx context.Context, params IngestComputeVpnTunnelsParams) (*IngestComputeVpnTunnelsResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GCP Compute VPN tunnel ingestion",
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
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("failed to ingest vpn tunnels: %w", err))
	}

	// Delete stale VPN tunnels
	if err := service.DeleteStaleVpnTunnels(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale vpn tunnels", "error", err)
	}

	logger.Info("Completed GCP Compute VPN tunnel ingestion",
		"projectID", params.ProjectID,
		"vpnTunnelCount", result.VpnTunnelCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestComputeVpnTunnelsResult{
		ProjectID:      result.ProjectID,
		VpnTunnelCount: result.VpnTunnelCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
