package vpngateway

import (
	"context"
	"fmt"
	"net/http"

	"go.temporal.io/sdk/activity"
	"google.golang.org/api/option"
	"gorm.io/gorm"

	"hotpot/pkg/base/config"
	"hotpot/pkg/base/ratelimit"
)

// Activities holds dependencies for Temporal activities.
type Activities struct {
	configService *config.Service
	db            *gorm.DB
	limiter       ratelimit.Limiter
}

// NewActivities creates a new Activities instance.
func NewActivities(configService *config.Service, db *gorm.DB, limiter ratelimit.Limiter) *Activities {
	return &Activities{
		configService: configService,
		db:            db,
		limiter:       limiter,
	}
}

// createClient creates a rate-limited GCP client with credentials.
func (a *Activities) createClient(ctx context.Context) (*Client, error) {
	var opts []option.ClientOption
	if credJSON := a.configService.GCPCredentialsJSON(); len(credJSON) > 0 {
		opts = append(opts, option.WithAuthCredentialsJSON(option.ServiceAccount, credJSON))
	}
	opts = append(opts, option.WithHTTPClient(&http.Client{
		Transport: ratelimit.NewRateLimitedTransport(a.limiter, nil),
	}))
	return NewClient(ctx, opts...)
}

// IngestComputeVpnGatewaysParams contains parameters for the ingest activity.
type IngestComputeVpnGatewaysParams struct {
	ProjectID string
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
	client, err := a.createClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}
	defer client.Close()

	// Create service
	service := NewService(client, a.db)
	result, err := service.Ingest(ctx, IngestParams{
		ProjectID: params.ProjectID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to ingest vpn gateways: %w", err)
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
