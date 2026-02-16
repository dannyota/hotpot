package instance

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPBigtableInstanceWorkflowParams contains parameters for the instance workflow.
type GCPBigtableInstanceWorkflowParams struct {
	ProjectID string
}

// GCPBigtableInstanceWorkflowResult contains the result of the instance workflow.
type GCPBigtableInstanceWorkflowResult struct {
	ProjectID      string
	InstanceCount  int
	DurationMillis int64
}

// GCPBigtableInstanceWorkflow ingests Bigtable instances for a single project.
func GCPBigtableInstanceWorkflow(ctx workflow.Context, params GCPBigtableInstanceWorkflowParams) (*GCPBigtableInstanceWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPBigtableInstanceWorkflow", "projectID", params.ProjectID)

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

	var result IngestBigtableInstancesResult
	err := workflow.ExecuteActivity(activityCtx, IngestBigtableInstancesActivity, IngestBigtableInstancesParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest Bigtable instances", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPBigtableInstanceWorkflow",
		"projectID", params.ProjectID,
		"instanceCount", result.InstanceCount,
	)

	return &GCPBigtableInstanceWorkflowResult{
		ProjectID:      result.ProjectID,
		InstanceCount:  result.InstanceCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
