package sshkey

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

// IngestComputeSSHKeysParams contains parameters for the ingest activity.
type IngestComputeSSHKeysParams struct {
	ProjectID string
	Region    string
}

// IngestComputeSSHKeysResult contains the result of the ingest activity.
type IngestComputeSSHKeysResult struct {
	KeyCount       int
	DurationMillis int64
}

// IngestComputeSSHKeysActivity is the activity function reference for workflow registration.
var IngestComputeSSHKeysActivity = (*Activities).IngestComputeSSHKeys

// IngestComputeSSHKeys is a Temporal activity that ingests GreenNode SSH keys.
func (a *Activities) IngestComputeSSHKeys(ctx context.Context, params IngestComputeSSHKeysParams) (*IngestComputeSSHKeysResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GreenNode SSH key ingestion", "projectID", params.ProjectID, "region", params.Region)

	client, err := NewClient(ctx, a.configService, a.limiter, params.Region, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}

	service := NewService(client, a.entClient)
	result, err := service.Ingest(ctx, params.ProjectID, params.Region)
	if err != nil {
		return nil, fmt.Errorf("ingest ssh keys: %w", err)
	}

	if err := service.DeleteStaleSSHKeys(ctx, params.ProjectID, params.Region, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale ssh keys", "error", err)
	}

	logger.Info("Completed GreenNode SSH key ingestion",
		"keyCount", result.KeyCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestComputeSSHKeysResult{
		KeyCount:       result.KeyCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
