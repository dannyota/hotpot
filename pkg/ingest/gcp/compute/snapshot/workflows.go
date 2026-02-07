package snapshot

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPComputeSnapshotWorkflowParams contains parameters for the snapshot workflow.
type GCPComputeSnapshotWorkflowParams struct {
	ProjectID string
}

// GCPComputeSnapshotWorkflowResult contains the result of the snapshot workflow.
type GCPComputeSnapshotWorkflowResult struct {
	ProjectID      string
	SnapshotCount  int
	DurationMillis int64
}

// GCPComputeSnapshotWorkflow ingests GCP Compute snapshots for a single project.
func GCPComputeSnapshotWorkflow(ctx workflow.Context, params GCPComputeSnapshotWorkflowParams) (*GCPComputeSnapshotWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPComputeSnapshotWorkflow", "projectID", params.ProjectID)

	// Activity options
	activityOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	activityCtx := workflow.WithActivityOptions(ctx, activityOpts)

	// Execute ingest activity
	var result IngestComputeSnapshotsResult
	err := workflow.ExecuteActivity(activityCtx, IngestComputeSnapshotsActivity, IngestComputeSnapshotsParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest snapshots", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPComputeSnapshotWorkflow",
		"projectID", params.ProjectID,
		"snapshotCount", result.SnapshotCount,
	)

	return &GCPComputeSnapshotWorkflowResult{
		ProjectID:      result.ProjectID,
		SnapshotCount:  result.SnapshotCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
