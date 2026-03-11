package connector

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/activity"
	"google.golang.org/api/option"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/gcpauth"
	"danny.vn/hotpot/pkg/base/ratelimit"
	"danny.vn/hotpot/pkg/base/temporalerr"
	entcompute "danny.vn/hotpot/pkg/storage/ent/gcp/compute"
	"danny.vn/hotpot/pkg/storage/ent/gcp/compute/bronzegcpcomputesubnetwork"
	entvpcaccess "danny.vn/hotpot/pkg/storage/ent/gcp/vpcaccess"
)

// Activities holds dependencies for Temporal activities.
type Activities struct {
	configService *config.Service
	entClient     *entvpcaccess.Client
	computeClient *entcompute.Client
	limiter       ratelimit.Limiter
}

// NewActivities creates a new Activities instance.
func NewActivities(configService *config.Service, entClient *entvpcaccess.Client, computeClient *entcompute.Client, limiter ratelimit.Limiter) *Activities {
	return &Activities{
		configService: configService,
		entClient:     entClient,
		computeClient: computeClient,
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

// IngestVpcAccessConnectorsParams contains parameters for the ingest activity.
type IngestVpcAccessConnectorsParams struct {
	ProjectID      string
	QuotaProjectID string
}

// IngestVpcAccessConnectorsResult contains the result of the ingest activity.
type IngestVpcAccessConnectorsResult struct {
	ProjectID      string
	ConnectorCount int
	DurationMillis int64
}

// IngestVpcAccessConnectorsActivity is the activity function reference for workflow registration.
var IngestVpcAccessConnectorsActivity = (*Activities).IngestVpcAccessConnectors

// IngestVpcAccessConnectors is a Temporal activity that ingests GCP VPC Access connectors.
func (a *Activities) IngestVpcAccessConnectors(ctx context.Context, params IngestVpcAccessConnectorsParams) (*IngestVpcAccessConnectorsResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GCP VPC Access connector ingestion",
		"projectID", params.ProjectID,
	)

	// Query distinct regions from already-ingested subnetworks
	subnetworks, err := a.computeClient.BronzeGCPComputeSubnetwork.Query().
		Where(bronzegcpcomputesubnetwork.ProjectID(params.ProjectID)).
		Select(bronzegcpcomputesubnetwork.FieldRegion).
		All(ctx)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("failed to query regions from subnetworks: %w", err))
	}

	// Extract unique regions
	regionSet := make(map[string]struct{})
	for _, sub := range subnetworks {
		if sub.Region != "" {
			regionSet[sub.Region] = struct{}{}
		}
	}

	var regions []string
	for region := range regionSet {
		regions = append(regions, region)
	}

	if len(regions) == 0 {
		logger.Info("No regions found from subnetworks, skipping VPC Access connector ingestion",
			"projectID", params.ProjectID,
		)
		return &IngestVpcAccessConnectorsResult{
			ProjectID:      params.ProjectID,
			ConnectorCount: 0,
		}, nil
	}

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
		Regions:   regions,
	})
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("failed to ingest connectors: %w", err))
	}

	// Delete stale connectors
	if err := service.DeleteStaleConnectors(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale connectors", "error", err)
	}

	logger.Info("Completed GCP VPC Access connector ingestion",
		"projectID", params.ProjectID,
		"connectorCount", result.ConnectorCount,
		"regionCount", len(regions),
		"durationMillis", result.DurationMillis,
	)

	return &IngestVpcAccessConnectorsResult{
		ProjectID:      result.ProjectID,
		ConnectorCount: result.ConnectorCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
