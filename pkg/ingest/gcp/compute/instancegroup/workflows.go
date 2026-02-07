package instancegroup

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPComputeInstanceGroupWorkflowParams contains parameters for the instance group workflow.
type GCPComputeInstanceGroupWorkflowParams struct {
	ProjectID string
}

// GCPComputeInstanceGroupWorkflowResult contains the result of the instance group workflow.
type GCPComputeInstanceGroupWorkflowResult struct {
	ProjectID          string
	InstanceGroupCount int
	DurationMillis     int64
}

// GCPComputeInstanceGroupWorkflow ingests GCP Compute instance groups for a single project.
func GCPComputeInstanceGroupWorkflow(ctx workflow.Context, params GCPComputeInstanceGroupWorkflowParams) (*GCPComputeInstanceGroupWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPComputeInstanceGroupWorkflow", "projectID", params.ProjectID)

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
	var result IngestComputeInstanceGroupsResult
	err := workflow.ExecuteActivity(activityCtx, IngestComputeInstanceGroupsActivity, IngestComputeInstanceGroupsParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest instance groups", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPComputeInstanceGroupWorkflow",
		"projectID", params.ProjectID,
		"instanceGroupCount", result.InstanceGroupCount,
	)

	return &GCPComputeInstanceGroupWorkflowResult{
		ProjectID:          result.ProjectID,
		InstanceGroupCount: result.InstanceGroupCount,
		DurationMillis:     result.DurationMillis,
	}, nil
}
