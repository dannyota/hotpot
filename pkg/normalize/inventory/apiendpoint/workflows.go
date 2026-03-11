package apiendpoint

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// NormalizeApiEndpointsWorkflowResult holds the workflow result.
type NormalizeApiEndpointsWorkflowResult struct {
	Result NormalizeApiEndpointsResult
}

// NormalizeApiEndpointsWorkflow normalizes API endpoints from all providers.
func NormalizeApiEndpointsWorkflow(ctx workflow.Context) (*NormalizeApiEndpointsWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting NormalizeApiEndpointsWorkflow")

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

	var result NormalizeApiEndpointsResult
	if err := workflow.ExecuteActivity(activityCtx, NormalizeApiEndpointsActivity).
		Get(ctx, &result); err != nil {
		logger.Error("Failed to normalize API endpoints", "error", err)
		return nil, err
	}

	logger.Info("Completed NormalizeApiEndpointsWorkflow",
		"created", result.Created,
		"updated", result.Updated,
		"deleted", result.Deleted)

	return &NormalizeApiEndpointsWorkflowResult{Result: result}, nil
}
