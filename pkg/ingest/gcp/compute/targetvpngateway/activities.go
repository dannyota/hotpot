package targetvpngateway

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

// IngestComputeTargetVpnGatewaysParams contains parameters for the ingest activity.
type IngestComputeTargetVpnGatewaysParams struct {
	ProjectID string
}

// IngestComputeTargetVpnGatewaysResult contains the result of the ingest activity.
type IngestComputeTargetVpnGatewaysResult struct {
	ProjectID             string
	TargetVpnGatewayCount int
	DurationMillis        int64
}

// IngestComputeTargetVpnGatewaysActivity is the activity function reference for workflow registration.
var IngestComputeTargetVpnGatewaysActivity = (*Activities).IngestComputeTargetVpnGateways

// IngestComputeTargetVpnGateways is a Temporal activity that ingests GCP Compute Classic VPN gateways.
func (a *Activities) IngestComputeTargetVpnGateways(ctx context.Context, params IngestComputeTargetVpnGatewaysParams) (*IngestComputeTargetVpnGatewaysResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GCP Compute target VPN gateway ingestion",
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
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("failed to ingest target vpn gateways: %w", err))
	}

	// Delete stale target VPN gateways
	if err := service.DeleteStaleTargetVpnGateways(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale target vpn gateways", "error", err)
	}

	logger.Info("Completed GCP Compute target VPN gateway ingestion",
		"projectID", params.ProjectID,
		"targetVpnGatewayCount", result.TargetVpnGatewayCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestComputeTargetVpnGatewaysResult{
		ProjectID:             result.ProjectID,
		TargetVpnGatewayCount: result.TargetVpnGatewayCount,
		DurationMillis:        result.DurationMillis,
	}, nil
}
