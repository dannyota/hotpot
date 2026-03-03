package snapshot

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"

	"github.com/dannyota/hotpot/pkg/base/temporalerr"
	entcompute "github.com/dannyota/hotpot/pkg/storage/ent/gcp/compute"
)

const snapshotNextPageSignal = "snapshot-next-page"

// NextPageTokenSignal is the signal payload sent from activity to workflow.
type NextPageTokenSignal struct {
	NextPageToken string
}

// Activities holds dependencies for Temporal activities.
type Activities struct {
	client         *Client
	entClient      *entcompute.Client
	temporalClient client.Client
}

// NewActivities creates a new Activities instance.
func NewActivities(client *Client, entClient *entcompute.Client, temporalClient client.Client) *Activities {
	return &Activities{
		client:         client,
		entClient:      entClient,
		temporalClient: temporalClient,
	}
}

// FetchAndSaveSnapshotsPageParams contains parameters for the page activity.
type FetchAndSaveSnapshotsPageParams struct {
	ProjectID   string
	PageToken   string
	PageSize    int
	CollectedAt time.Time
	WorkflowID  string
	RunID       string
}

// FetchAndSaveSnapshotsPageResult contains the result of the page activity.
type FetchAndSaveSnapshotsPageResult struct {
	Count         int
	NextPageToken string
}

// FetchAndSaveSnapshotsPageActivity is the activity function reference for workflow registration.
var FetchAndSaveSnapshotsPageActivity = (*Activities).FetchAndSaveSnapshotsPage

// FetchAndSaveSnapshotsPage fetches one page of snapshots from GCP and saves to DB.
// After the GCP fetch, it signals the workflow with the next page token so the
// workflow can start the next fetch while this activity saves to DB.
func (a *Activities) FetchAndSaveSnapshotsPage(ctx context.Context, params FetchAndSaveSnapshotsPageParams) (*FetchAndSaveSnapshotsPageResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Fetching snapshot page",
		"projectID", params.ProjectID,
		"pageSize", params.PageSize,
		"hasPageToken", params.PageToken != "",
	)

	snapshots, nextToken, err := a.client.ListSnapshotsPage(ctx, params.ProjectID, params.PageSize, params.PageToken)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("list snapshots page: %w", err))
	}

	// Signal workflow with next page token BEFORE saving to DB.
	// This lets the workflow start the next GCP fetch while we save.
	if err := a.temporalClient.SignalWorkflow(ctx, params.WorkflowID, params.RunID,
		snapshotNextPageSignal, NextPageTokenSignal{NextPageToken: nextToken}); err != nil {
		logger.Warn("Failed to signal next page token", "error", err)
		// Non-fatal: workflow will fall back to reading token from result.
	}

	service := NewService(nil, a.entClient)
	count, err := service.IngestPage(ctx, snapshots, params.ProjectID, params.CollectedAt)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("ingest snapshots page: %w", err))
	}

	logger.Info("Saved snapshot page",
		"projectID", params.ProjectID,
		"count", count,
		"hasNextPage", nextToken != "",
	)

	return &FetchAndSaveSnapshotsPageResult{
		Count:         count,
		NextPageToken: nextToken,
	}, nil
}

// DeleteStaleSnapshotsParams contains parameters for the delete stale activity.
type DeleteStaleSnapshotsParams struct {
	ProjectID   string
	CollectedAt time.Time
}

// DeleteStaleSnapshotsActivity is the activity function reference for workflow registration.
var DeleteStaleSnapshotsActivity = (*Activities).DeleteStaleSnapshots

// DeleteStaleSnapshots removes snapshots not seen in this collection run.
func (a *Activities) DeleteStaleSnapshots(ctx context.Context, params DeleteStaleSnapshotsParams) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Deleting stale snapshots", "projectID", params.ProjectID)

	service := NewService(nil, a.entClient)
	if err := service.DeleteStaleSnapshots(ctx, params.ProjectID, params.CollectedAt); err != nil {
		return temporalerr.MaybeNonRetryable(fmt.Errorf("delete stale snapshots: %w", err))
	}

	return nil
}
