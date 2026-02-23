package servergroup

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/activity"

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

// IngestComputeServerGroupsParams contains parameters for the ingest activity.
type IngestComputeServerGroupsParams struct {
	ProjectID string
}

// IngestComputeServerGroupsResult contains the result of the ingest activity.
type IngestComputeServerGroupsResult struct {
	GroupCount     int
	DurationMillis int64
}

// IngestComputeServerGroupsActivity is the activity function reference for workflow registration.
var IngestComputeServerGroupsActivity = (*Activities).IngestComputeServerGroups

// IngestComputeServerGroups is a Temporal activity that ingests GreenNode server groups.
func (a *Activities) IngestComputeServerGroups(ctx context.Context, params IngestComputeServerGroupsParams) (*IngestComputeServerGroupsResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GreenNode server group ingestion", "projectID", params.ProjectID)

	client, err := NewClient(ctx, a.configService, a.limiter)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}

	service := NewService(client, a.entClient)
	result, err := service.Ingest(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("ingest server groups: %w", err)
	}

	if err := service.DeleteStaleServerGroups(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale server groups", "error", err)
	}

	logger.Info("Completed GreenNode server group ingestion",
		"groupCount", result.GroupCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestComputeServerGroupsResult{
		GroupCount:     result.GroupCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
