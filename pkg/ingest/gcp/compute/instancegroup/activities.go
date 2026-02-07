package instancegroup

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

// IngestComputeInstanceGroupsParams contains parameters for the ingest activity.
type IngestComputeInstanceGroupsParams struct {
	ProjectID string
}

// IngestComputeInstanceGroupsResult contains the result of the ingest activity.
type IngestComputeInstanceGroupsResult struct {
	ProjectID          string
	InstanceGroupCount int
	DurationMillis     int64
}

// IngestComputeInstanceGroupsActivity is the activity function reference for workflow registration.
var IngestComputeInstanceGroupsActivity = (*Activities).IngestComputeInstanceGroups

// IngestComputeInstanceGroups is a Temporal activity that ingests GCP Compute instance groups.
func (a *Activities) IngestComputeInstanceGroups(ctx context.Context, params IngestComputeInstanceGroupsParams) (*IngestComputeInstanceGroupsResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GCP Compute instance group ingestion",
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
		return nil, fmt.Errorf("failed to ingest instance groups: %w", err)
	}

	// Delete stale instance groups
	if err := service.DeleteStaleInstanceGroups(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale instance groups", "error", err)
	}

	logger.Info("Completed GCP Compute instance group ingestion",
		"projectID", params.ProjectID,
		"instanceGroupCount", result.InstanceGroupCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestComputeInstanceGroupsResult{
		ProjectID:          result.ProjectID,
		InstanceGroupCount: result.InstanceGroupCount,
		DurationMillis:     result.DurationMillis,
	}, nil
}

