package disk

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPComputeDiskWorkflowParams contains parameters for the disk workflow.
type GCPComputeDiskWorkflowParams struct {
	ProjectID string
}

// GCPComputeDiskWorkflowResult contains the result of the disk workflow.
type GCPComputeDiskWorkflowResult struct {
	ProjectID      string
	DiskCount      int
	DurationMillis int64
}

// GCPComputeDiskWorkflow ingests GCP Compute disks for a single project.
func GCPComputeDiskWorkflow(ctx workflow.Context, params GCPComputeDiskWorkflowParams) (*GCPComputeDiskWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPComputeDiskWorkflow", "projectID", params.ProjectID)

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
	var result IngestComputeDisksResult
	err := workflow.ExecuteActivity(activityCtx, IngestComputeDisksActivity, IngestComputeDisksParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest disks", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPComputeDiskWorkflow",
		"projectID", params.ProjectID,
		"diskCount", result.DiskCount,
	)

	return &GCPComputeDiskWorkflowResult{
		ProjectID:      result.ProjectID,
		DiskCount:      result.DiskCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
