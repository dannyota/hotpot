package instance

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPComputeInstanceWorkflowParams contains parameters for the instance workflow.
type GCPComputeInstanceWorkflowParams struct {
	ProjectID string
}

// GCPComputeInstanceWorkflowResult contains the result of the instance workflow.
type GCPComputeInstanceWorkflowResult struct {
	ProjectID      string
	InstanceCount  int
	DurationMillis int64
}

// GCPComputeInstanceWorkflow ingests GCP Compute instances for a single project.
func GCPComputeInstanceWorkflow(ctx workflow.Context, params GCPComputeInstanceWorkflowParams) (*GCPComputeInstanceWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPComputeInstanceWorkflow", "projectID", params.ProjectID)

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
	var result IngestComputeInstancesResult
	err := workflow.ExecuteActivity(activityCtx, IngestComputeInstancesActivity, IngestComputeInstancesParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest instances", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPComputeInstanceWorkflow",
		"projectID", params.ProjectID,
		"instanceCount", result.InstanceCount,
	)

	return &GCPComputeInstanceWorkflowResult{
		ProjectID:      result.ProjectID,
		InstanceCount:  result.InstanceCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
