package targetpool

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPComputeTargetPoolWorkflowParams contains parameters for the target pool workflow.
type GCPComputeTargetPoolWorkflowParams struct {
	ProjectID string
}

// GCPComputeTargetPoolWorkflowResult contains the result of the target pool workflow.
type GCPComputeTargetPoolWorkflowResult struct {
	ProjectID       string
	TargetPoolCount int
	DurationMillis  int64
}

// GCPComputeTargetPoolWorkflow ingests GCP Compute target pools for a single project.
func GCPComputeTargetPoolWorkflow(ctx workflow.Context, params GCPComputeTargetPoolWorkflowParams) (*GCPComputeTargetPoolWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPComputeTargetPoolWorkflow", "projectID", params.ProjectID)

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
	var result IngestComputeTargetPoolsResult
	err := workflow.ExecuteActivity(activityCtx, IngestComputeTargetPoolsActivity, IngestComputeTargetPoolsParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest target pools", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPComputeTargetPoolWorkflow",
		"projectID", params.ProjectID,
		"targetPoolCount", result.TargetPoolCount,
	)

	return &GCPComputeTargetPoolWorkflowResult{
		ProjectID:       result.ProjectID,
		TargetPoolCount: result.TargetPoolCount,
		DurationMillis:  result.DurationMillis,
	}, nil
}
