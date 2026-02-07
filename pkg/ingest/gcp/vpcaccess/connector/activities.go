package connector

import (
	"context"
	"fmt"
	"net/http"

	"go.temporal.io/sdk/activity"
	"google.golang.org/api/option"
	"gorm.io/gorm"

	"hotpot/pkg/base/config"
	"hotpot/pkg/base/models/bronze"
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

// IngestVpcAccessConnectorsParams contains parameters for the ingest activity.
type IngestVpcAccessConnectorsParams struct {
	ProjectID string
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
	var regions []string
	if err := a.db.Model(&bronze.GCPComputeSubnetwork{}).
		Where("project_id = ?", params.ProjectID).
		Distinct("region").
		Pluck("region", &regions).Error; err != nil {
		return nil, fmt.Errorf("failed to query regions from subnetworks: %w", err)
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
	client, err := a.createClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}
	defer client.Close()

	// Create service
	service := NewService(client, a.db)
	result, err := service.Ingest(ctx, IngestParams{
		ProjectID: params.ProjectID,
		Regions:   regions,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to ingest connectors: %w", err)
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
