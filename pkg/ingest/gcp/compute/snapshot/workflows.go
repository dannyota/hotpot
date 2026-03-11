package snapshot

import (
	"time"

	"danny.vn/hotpot/pkg/base/temporalerr"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const (
	pageSize           = 500
	maxConcurrentPages = 10
)

// GCPComputeSnapshotWorkflowParams contains parameters for the snapshot workflow.
type GCPComputeSnapshotWorkflowParams struct {
	ProjectID      string
	QuotaProjectID string
}

// GCPComputeSnapshotWorkflowResult contains the result of the snapshot workflow.
type GCPComputeSnapshotWorkflowResult struct {
	ProjectID     string
	SnapshotCount int
}

// GCPComputeSnapshotWorkflow ingests GCP Compute snapshots for a single project.
// It pipelines page fetches: as each activity finishes its GCP fetch, it signals
// the workflow with the next page token, and the workflow immediately starts the
// next activity. All activities run concurrently (rate-limited by the GCP limiter).
func GCPComputeSnapshotWorkflow(ctx workflow.Context, params GCPComputeSnapshotWorkflowParams) (*GCPComputeSnapshotWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPComputeSnapshotWorkflow", "projectID", params.ProjectID)

	collectedAt := workflow.Now(ctx)
	wfInfo := workflow.GetInfo(ctx).WorkflowExecution

	fetchOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	fetchCtx := workflow.WithActivityOptions(ctx, fetchOpts)

	signalCh := workflow.GetSignalChannel(ctx, snapshotNextPageSignal)

	totalCount := 0
	var pendingFutures []workflow.Future

	// drainOldest waits for the oldest pending future and accumulates its count.
	drainOldest := func() error {
		var result FetchAndSaveSnapshotsPageResult
		if err := pendingFutures[0].Get(ctx, &result); err != nil {
			return temporalerr.PropagateNonRetryable(err)
		}
		totalCount += result.Count
		pendingFutures = pendingFutures[1:]
		return nil
	}

	// startPage launches a new page activity, draining the oldest if at capacity.
	startPage := func(pageToken string) (workflow.Future, error) {
		if len(pendingFutures) >= maxConcurrentPages {
			if err := drainOldest(); err != nil {
				return nil, err
			}
		}
		f := workflow.ExecuteActivity(fetchCtx, FetchAndSaveSnapshotsPageActivity, FetchAndSaveSnapshotsPageParams{
			ProjectID:      params.ProjectID,
			QuotaProjectID: params.QuotaProjectID,
			PageToken:      pageToken,
			PageSize:       pageSize,
			CollectedAt:    collectedAt,
			WorkflowID:     wfInfo.ID,
			RunID:          wfInfo.RunID,
		})
		pendingFutures = append(pendingFutures, f)
		return f, nil
	}

	// Start first page activity.
	latestFuture, _ := startPage("")

	// Dispatch loop: fire new activities as signals arrive.
	for {
		var signal NextPageTokenSignal
		var gotSignal bool

		selector := workflow.NewSelector(ctx)
		selector.AddReceive(signalCh, func(ch workflow.ReceiveChannel, _ bool) {
			ch.Receive(ctx, &signal)
			gotSignal = true
		})
		selector.AddFuture(latestFuture, func(f workflow.Future) {
			// Latest activity completed before signaling.
		})
		selector.Select(ctx)

		if !gotSignal {
			// Latest activity finished without signaling — failed during fetch
			// or signal delivery failed. Check the result.
			var result FetchAndSaveSnapshotsPageResult
			if err := latestFuture.Get(ctx, &result); err != nil {
				return nil, temporalerr.PropagateNonRetryable(err)
			}
			// Signal delivery failed — fall back to using token from result.
			if result.NextPageToken == "" {
				break
			}
			var err error
			latestFuture, err = startPage(result.NextPageToken)
			if err != nil {
				return nil, err
			}
			continue
		}

		if signal.NextPageToken == "" {
			break // Last page — no more activities to start.
		}

		var err error
		latestFuture, err = startPage(signal.NextPageToken)
		if err != nil {
			return nil, err
		}
	}

	// Drain remaining pending futures.
	for len(pendingFutures) > 0 {
		if err := drainOldest(); err != nil {
			return nil, err
		}
	}

	// Delete stale snapshots.
	deleteOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	deleteCtx := workflow.WithActivityOptions(ctx, deleteOpts)

	err := workflow.ExecuteActivity(deleteCtx, DeleteStaleSnapshotsActivity, DeleteStaleSnapshotsParams{
		ProjectID:   params.ProjectID,
		CollectedAt: collectedAt,
	}).Get(ctx, nil)
	if err != nil {
		return nil, temporalerr.PropagateNonRetryable(err)
	}

	logger.Info("Completed GCPComputeSnapshotWorkflow",
		"projectID", params.ProjectID,
		"snapshotCount", totalCount,
	)

	return &GCPComputeSnapshotWorkflowResult{
		ProjectID:     params.ProjectID,
		SnapshotCount: totalCount,
	}, nil
}
