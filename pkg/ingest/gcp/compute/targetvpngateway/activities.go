package targetvpngateway

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
		return nil, fmt.Errorf("create client: %w", err)
	}
	defer client.Close()

	// Create service
	service := NewService(client, a.db)
	result, err := service.Ingest(ctx, IngestParams{
		ProjectID: params.ProjectID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to ingest target vpn gateways: %w", err)
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
