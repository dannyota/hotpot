package vpngateway

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/activity"
	"google.golang.org/api/option"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/gcpauth"
	"danny.vn/hotpot/pkg/base/ratelimit"
	"danny.vn/hotpot/pkg/base/temporalerr"
	entvpn "danny.vn/hotpot/pkg/storage/ent/gcp/vpn"
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
func (a *Activities) createClient(ctx context.Context, quotaProjectID string) (*Client, error) {
	httpClient, err := gcpauth.NewHTTPClient(ctx, a.configService.GCPCredentialsJSON(), a.limiter)
	if err != nil {
		return nil, err
	}
	var opts []option.ClientOption
	opts = append(opts, option.WithHTTPClient(httpClient))
	if quotaProjectID != "" {
		opts = append(opts, option.WithQuotaProject(quotaProjectID))
	}
	return NewClient(ctx, opts...)
}

// IngestComputeVpnGatewaysParams contains parameters for the ingest activity.
type IngestComputeVpnGatewaysParams struct {
	ProjectID      string
	QuotaProjectID string
}

// IngestComputeVpnGatewaysResult contains the result of the ingest activity.
type IngestComputeVpnGatewaysResult struct {
	ProjectID       string
	VpnGatewayCount int
	DurationMillis  int64
}

// IngestComputeVpnGatewaysActivity is the activity function reference for workflow registration.
var IngestComputeVpnGatewaysActivity = (*Activities).IngestComputeVpnGateways

// IngestComputeVpnGateways is a Temporal activity that ingests GCP Compute VPN gateways.
func (a *Activities) IngestComputeVpnGateways(ctx context.Context, params IngestComputeVpnGatewaysParams) (*IngestComputeVpnGatewaysResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GCP Compute VPN gateway ingestion",
		"projectID", params.ProjectID,
	)

	// Create client for this activity
	client, err := a.createClient(ctx, params.QuotaProjectID)
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
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("failed to ingest vpn gateways: %w", err))
	}

	// Delete stale VPN gateways
	if err := service.DeleteStaleVpnGateways(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale vpn gateways", "error", err)
	}

	logger.Info("Completed GCP Compute VPN gateway ingestion",
		"projectID", params.ProjectID,
		"vpnGatewayCount", result.VpnGatewayCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestComputeVpnGatewaysResult{
		ProjectID:       result.ProjectID,
		VpnGatewayCount: result.VpnGatewayCount,
		DurationMillis:  result.DurationMillis,
	}, nil
}
