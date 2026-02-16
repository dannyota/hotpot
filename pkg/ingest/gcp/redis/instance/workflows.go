package instance

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPRedisInstanceWorkflowParams contains parameters for the Redis instance workflow.
type GCPRedisInstanceWorkflowParams struct {
	ProjectID string
}

// GCPRedisInstanceWorkflowResult contains the result of the Redis instance workflow.
type GCPRedisInstanceWorkflowResult struct {
	ProjectID      string
	InstanceCount  int
	DurationMillis int64
}

// GCPRedisInstanceWorkflow ingests GCP Memorystore Redis instances for a single project.
func GCPRedisInstanceWorkflow(ctx workflow.Context, params GCPRedisInstanceWorkflowParams) (*GCPRedisInstanceWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPRedisInstanceWorkflow", "projectID", params.ProjectID)

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

	var result IngestRedisInstancesResult
	err := workflow.ExecuteActivity(activityCtx, IngestRedisInstancesActivity, IngestRedisInstancesParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest Redis instances", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPRedisInstanceWorkflow",
		"projectID", params.ProjectID,
		"instanceCount", result.InstanceCount,
	)

	return &GCPRedisInstanceWorkflowResult{
		ProjectID:      result.ProjectID,
		InstanceCount:  result.InstanceCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
